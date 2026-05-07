package effects

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/zToasty/lumactl/internal/protocol"
)

type Static struct{}

// StaticParams описывает, что мы ждем в JSON
type StaticParams struct {
	Color string `json:"color"` // Принимает HEX: "#RRGGBB"
}

func (s *Static) Run(ctx context.Context, p protocol.DeviceProvider, numLEDs int, params json.RawMessage) error {
	// Устанавливаем дефолтный цвет (белый), если что-то пойдет не так
	targetColor := protocol.RGB{R: 255, G: 255, B: 255}

	// Парсим параметры, если они есть
	if len(params) > 0 {
		var prms StaticParams
		if err := json.Unmarshal(params, &prms); err != nil {
			slog.Warn("failed to parse static params, using default", "error", err)
		} else if prms.Color != "" {
			// Парсим HEX строку в нашу структуру RGB
			parsed, err := parseHexColor(prms.Color)
			if err != nil {
				slog.Warn("invalid hex color format", "color", prms.Color)
			} else {
				targetColor = parsed
			}
		}
	}

	// Подготавливаем кадр
	frame := make([]protocol.RGB, numLEDs)
	for i := range frame {
		frame[i] = targetColor
	}

	if err := p.SetColors(frame); err != nil {
		return err
	}

	slog.Debug("static effect started", "color", targetColor)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Debug("static effect stopped")
			return nil

		case <-ticker.C:
			if err := p.SetColors(frame); err != nil {
				return err
			}
		}
	}
}

// Вспомогательная функция для парсинга HEX (#RRGGBB)
func parseHexColor(s string) (protocol.RGB, error) {
	var c protocol.RGB
	var err error

	if len(s) > 0 && s[0] == '#' {
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	} else {
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	}

	if err != nil {
		return c, fmt.Errorf("invalid color: %w", err)
	}
	return c, nil
}
