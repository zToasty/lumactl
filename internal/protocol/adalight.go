package protocol

import (
	"fmt"
	"log/slog"

	"go.bug.st/serial"
)

// AdalightProvider implements DeviceProvider for USB controllers
type AdalightProvider struct {
	portName string
	port     serial.Port
	numLEDs  int
}

// NewAdalightProvider creates a new provider instance
func NewAdalightProvider(portName string, numLEDs int) *AdalightProvider {
	return &AdalightProvider{
		portName: portName,
		numLEDs:  numLEDs,
	}
}

func (a *AdalightProvider) Init() error {
	slog.Debug("initializing adalight provider", "port", a.portName, "baud_rate", 115200)

	mode := &serial.Mode{BaudRate: 115200}
	port, err := serial.Open(a.portName, mode)
	if err != nil {
		slog.Error("failed to open port", "port", a.portName, "error", err)
		return fmt.Errorf("failed to open port %s: %w", a.portName, err)
	}
	a.port = port

	slog.Info("adalight provider initialized successfully", "port", a.portName)
	return nil
}

func (a *AdalightProvider) SetColors(colors []RGB) error {
	if a.port == nil {
		return fmt.Errorf("Port is not initialized (call Init before SetColors)")
	}

	countMinusOne := len(colors) - 1
	hi := byte(countMinusOne >> 8)
	lo := byte(countMinusOne & 0xff)
	chk := hi ^ lo ^ 0x55

	packetSize := 6 + len(colors)*3
	packet := make([]byte, packetSize)

	packet[0] = 'A'
	packet[1] = 'd'
	packet[2] = 'a'
	packet[3] = hi
	packet[4] = lo
	packet[5] = chk

	idx := 6
	for _, c := range colors {
		packet[idx] = c.R
		packet[idx+1] = c.G
		packet[idx+2] = c.B
		idx += 3
	}

	_, err := a.port.Write(packet)
	if err != nil {
		return fmt.Errorf("Error writing data in port: %w", err)
	}

	return nil
}

func (a *AdalightProvider) Close() error {
	slog.Debug("closing adalight provider", "port", a.portName)
	if a.port != nil {
		if err := a.port.Close(); err != nil {
			slog.Error("failed to close port", "port", a.portName, "error", err)
			return err
		}
		slog.Info("port closed successfully", "port", a.portName)
	}
	return nil
}
