package main

import (
	"log/slog"
	"os"

	"github.com/zToasty/lumactl/internal/config"
	"github.com/zToasty/lumactl/internal/daemon"
	"github.com/zToasty/lumactl/internal/discovery"
	"github.com/zToasty/lumactl/internal/effects"
	"github.com/zToasty/lumactl/internal/protocol"
)

func main() {
	// Настраиваем логирование (уровень DEBUG для разработки)
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	// 1. Загружаем конфиг
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Поиск порта (если в конфиге стоит "auto")
	portName := cfg.Device.Port
	if portName == "auto" {
		foundPort, err := discovery.FindDevicePort(cfg.Device.VID, cfg.Device.PID)
		if err != nil {
			slog.Error("device discovery failed", "error", err)
			os.Exit(1)
		}
		portName = foundPort
	}

	// 3. Инициализация протокола
	provider := protocol.NewAdalightProvider(portName, cfg.Device.LEDCount)
	if err := provider.Init(); err != nil {
		slog.Error("failed to initialize protocol", "error", err)
		os.Exit(1)
	}
	defer provider.Close()

	// 4. Создаем менеджер эффектов
	mgr := effects.NewManager(provider, cfg.Device.LEDCount)

	// 5. Запуск сервера демона
	// Пока сделаем это основной задачей main.go
	srv := daemon.NewServer(mgr)

	slog.Info("starting lumactl daemon")
	if err := srv.Start(); err != nil {
		slog.Error("daemon server failed", "error", err)
		os.Exit(1)
	}
}
