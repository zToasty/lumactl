package discovery

import (
	"fmt"
	"log/slog"
	"strings"

	"go.bug.st/serial/enumerator"
)

// FindDevicePort сканирует систему и возвращает путь к порту устройства (например, /dev/ttyUSB0)
func FindDevicePort(targetVID, targetPID string) (string, error) {
	slog.Debug("Scanning for serial ports", "target_vid", targetVID, "target_pid", targetPID)

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		slog.Error("failed to enumerate ports", "error", err)
		return "", fmt.Errorf("failed to scan ports: %w", err)
	}

	for _, port := range ports {
		if port.IsUSB {
			vid := strings.ToLower(port.VID)
			pid := strings.ToLower(port.PID)

			slog.Debug("found USB serial port", "port", port.Name, "vid", vid, "pid", pid)

			// Переводим targetVID/PID в нижний регистр для надежного сравнения
			if vid == strings.ToLower(targetVID) && pid == strings.ToLower(targetPID) {
				slog.Info("target device found", "port", port.Name, "vid", vid, "pid", pid)
				return port.Name, nil
			}
		}
	}

	slog.Warn("target device not found", "target_vid", targetVID, "target_pid", targetPID)
	return "", fmt.Errorf("device with VID:%s and PID:%s not found", targetVID, targetPID)
}
