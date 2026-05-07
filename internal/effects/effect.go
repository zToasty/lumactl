package effects

import (
	"context"
	"encoding/json"

	"github.com/zToasty/lumactl/internal/protocol"
)

// Effect определяет общий интерфейс для всех режимов подсветки
type Effect interface {

	//Run запукает бесконечный цикл генерации кадров.
	// Он завершится только тогда, когда сработает ctx.Done() (например при Ctrl+C).
	Run(ctx context.Context, p protocol.DeviceProvider, numLEDs int, params json.RawMessage) error
}
