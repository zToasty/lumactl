package effects

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/zToasty/lumactl/internal/capture"
	"github.com/zToasty/lumactl/internal/protocol"
)

type LiquidSpectrum struct {
	mu      sync.Mutex
	bass    float64
	mid     float64
	high    float64
	maxBass float64
	maxMid  float64
	maxHigh float64
}

// gammaCorrect делает цвета сочнее
func gammaCorrect(v float64) byte {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return byte(math.Pow(v, 2.2) * 255)
}

func (m *LiquidSpectrum) Run(ctx context.Context, p protocol.DeviceProvider, numLEDs int, params json.RawMessage) error {
	engine := capture.NewEngine()

	stream, err := engine.Stream(ctx, 1024)
	if err != nil {
		return err
	}

	m.maxBass = 5000.0
	m.maxMid = 2000.0
	m.maxHigh = 1000.0

	attack := 0.6
	decay := 0.15

	lastLog := time.Now()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case samples, ok := <-stream:
				if !ok {
					return
				}

				b, mid, h := getBands(samples)

				if b > m.maxBass {
					m.maxBass = b
				} else {
					m.maxBass *= 0.999
				}
				if mid > m.maxMid {
					m.maxMid = mid
				} else {
					m.maxMid *= 0.999
				}
				if h > m.maxHigh {
					m.maxHigh = h
				} else {
					m.maxHigh *= 0.999
				}

				if m.maxBass < 3000 {
					m.maxBass = 3000
				}
				if m.maxMid < 1000 {
					m.maxMid = 1000
				}
				if m.maxHigh < 500 {
					m.maxHigh = 500
				}

				nb := b / m.maxBass
				nm := mid / m.maxMid
				nh := h / m.maxHigh

				if time.Since(lastLog) > 2*time.Second {
					slog.Info("Spectrum levels", "bass", b, "nb", nb, "mid", mid, "nm", nm, "high", h, "nh", nh)
					lastLog = time.Now()
				}

				m.mu.Lock()
				if nb > m.bass {
					m.bass = m.bass*(1-attack) + nb*attack
				} else {
					m.bass = m.bass*(1-decay) + nb*decay
				}
				if nm > m.mid {
					m.mid = m.mid*(1-attack) + nm*attack
				} else {
					m.mid = m.mid*(1-decay) + nm*decay
				}
				if nh > m.high {
					m.high = m.high*(1-attack) + nh*attack
				} else {
					m.high = m.high*(1-decay) + nh*decay
				}
				m.mu.Unlock()
			}
		}
	}()

	// Ограничиваем FPS, так как на 115200 бод 255 светодиодов передаются ~66мс!
	// Если слать быстрее, драйвер CH340 захлебнется и выдаст "input/output error".
	ticker := time.NewTicker(70 * time.Millisecond) // ~14 FPS
	defer ticker.Stop()

	// Фаза для движения "волны"
	phase := 0.0

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			m.mu.Lock()
			b, mid, h := m.bass, m.mid, m.high
			m.mu.Unlock()

			// Базовая скорость движения волны
			phase += 0.05 + (b * 0.067) // Бас ускоряет волну!

			frame := make([]protocol.RGB, numLEDs)

			for i := 0; i < numLEDs; i++ {
				// Вычисляем позицию светодиода по кругу (от 0 до 2*Pi)
				pos := (float64(i) / float64(numLEDs)) * 2 * math.Pi

				// Смешиваем три синусоиды с разной частотой и сдвигом фазы
				// Бас (Красный) - широкая плавная волна
				rWave := (math.Sin(pos+phase) + 1.0) / 2.0
				// Середина (Зеленый) - более частая волна
				gWave := (math.Sin(pos*2.0-phase*1.5) + 1.0) / 2.0
				// Высокие (Синий) - частая волна в противофазе
				bWave := (math.Sin(pos*3.0+phase*2.0) + 1.0) / 2.0

				// Умножаем волны на энергию из музыки
				frame[i] = protocol.RGB{
					R: gammaCorrect(rWave * b),
					G: gammaCorrect(gWave * mid),
					B: gammaCorrect(bWave * h),
				}
			}

			p.SetColors(frame)
		}
	}
}
