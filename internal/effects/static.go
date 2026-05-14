package effects

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zToasty/lumactl/internal/protocol"
)

var standardColors = map[string]protocol.RGB{
	"red":    {R: 255, G: 0, B: 0},
	"green":  {R: 0, G: 255, B: 0},
	"blue":   {R: 0, G: 0, B: 255},
	"yellow": {R: 255, G: 255, B: 0},
	"purple": {R: 255, G: 0, B: 255},
	"cyan":   {R: 0, G: 255, B: 255},
	"white":  {R: 255, G: 255, B: 255},
	"orange": {R: 255, G: 165, B: 0},
	"off":    {R: 0, G: 0, B: 0},
}

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
	if c, ok := standardColors[s]; ok {
		return c, nil
	}

	s = strings.TrimPrefix(s, "#")
	s = strings.TrimSpace(s)

	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 3 {
		return protocol.RGB{}, fmt.Errorf("invalid color format: %s", s)
	}

	return protocol.RGB{R: b[0], G: b[1], B: b[2]}, nil
}
