package effects

import (
	"context"
	"time"

	"github.com/zToasty/lumactl/internal/protocol"
)

type Static struct {
	Color protocol.RGB
}

func (s *Static) Run(ctx context.Context, p protocol.DeviceProvider, numLEDs int) error {
	frame := make([]protocol.RGB, numLEDs)
	for i := range frame {
		frame[i] = s.Color
	}

	if err := p.SetColors(frame); err != nil {
		return err
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := p.SetColors(frame); err != nil {
				return err
			}
		}
	}
}
