package effects

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/zToasty/lumactl/internal/protocol"
)

// registry — внутренняя карта всех доступных эффектов
var registry = map[string]Effect{
	"static": &Static{},

	// "rainbow": &Rainbow{},
	// "ambient": &Ambient{},
}

// GetEffect ищет эффект в реестре по имени
func GetEffect(name string) (Effect, error) {
	effect, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown effect: %s", name)
	}
	return effect, nil
}

// Manager управляет жизненным циклом активного эффекта
type Manager struct {
	mu           sync.Mutex
	activeCancel context.CancelFunc
	provider     protocol.DeviceProvider
	ledCount     int
}

// NewManager создает новый экземпляр менеджера
func NewManager(p protocol.DeviceProvider, leds int) *Manager {
	return &Manager{
		provider: p,
		ledCount: leds,
	}
}

func (m *Manager) SwitchEffect(name string, params json.RawMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	effect, err := GetEffect(name)
	if err != nil {
		slog.Warn("Attempted to switch to non-existent effect", "name", name)
		return err
	}

	if m.activeCancel != nil {
		slog.Debug("Stopping active effect to switch", "new_effect", name)
		m.activeCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.activeCancel = cancel

	slog.Info("Switching effect", "name", name)

	go func() {
		if err := effect.Run(ctx, m.provider, m.ledCount, params); err != nil {
			slog.Error("Effect execution failed", "name", name, "error", err)
		}
	}()

	return nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeCancel != nil {
		slog.Info("stopping active effect and turning off LEDs")
		m.activeCancel()
		m.activeCancel = nil
	}

	black := make([]protocol.RGB, m.ledCount)
	if err := m.provider.SetColors(black); err != nil {
		slog.Error("failed to turn off LEDs during stop", "error", err)
	}
}
