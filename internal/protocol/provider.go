package protocol

// RGB описывает один цвет
type RGB struct {
	R, G, B byte
}

// DeviceProvider — общий интерфейс для любого устройства подсветки
type DeviceProvider interface {
	// Init открывает подключение и подготавливает устройство к работе
	Init() error
	// SetColors отправляет массив цветов на устройство
	SetColors(colors []RGB) error
	// Close корректно завершает работу с устройством
	Close() error
}
