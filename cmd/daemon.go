package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zToasty/lumactl/internal/daemon"
	"github.com/zToasty/lumactl/internal/discovery"
	"github.com/zToasty/lumactl/internal/effects"
	"github.com/zToasty/lumactl/internal/protocol"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Запустить сервер демона (владелец USB-порта и сокета)",
	Long: `Daemon захватывает USB-устройство, открывает Unix-сокет
и обслуживает команды от клиентов (color, stop, status).

Это долгоживущий процесс. Завершается по SIGINT или SIGTERM.`,
	Run: runDaemon, // ← вынесено в отдельную функцию, чтобы не раздувать Run-литерал
}

// runDaemon — точка входа демона.
// Вся работа с железом происходит здесь и только здесь.
func runDaemon(cmd *cobra.Command, args []string) {
	// 1. Graceful shutdown через контекст
	//    signal.NotifyContext ловит SIGINT/SIGTERM и отменяет ctx.
	//    TODO: создай ctx, который отменится по сигналу ОС.
	//    Подсказка: signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Определяем порт
	portName := cfg.Device.Port
	if portName == "auto" {
		foundPort, err := discovery.FindDevicePort(cfg.Device.VID, cfg.Device.PID)
		if err != nil {
			slog.Error("device discovery failed", "error", err)
			os.Exit(1)
		}
		portName = foundPort
	}

	// 3. Инициализируем провайдер протокола)

	provider := protocol.NewAdalightProvider(portName, cfg.Device.LEDCount)
	if err := provider.Init(); err != nil {
		slog.Error("failed to initialize provider protocol", "error", err)
		os.Exit(1)
	}
	defer provider.Close()

	// 5. Создаём менеджер эффектов
	mgr := effects.NewManager(provider, cfg.Device.LEDCount)
	srv := daemon.NewServer(mgr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	slog.Info("daemon started", "port", portName, "leds", cfg.Device.LEDCount)

	select {
	case err := <-errCh:
		if err != nil {
			slog.Error("server error", "error", err)
		}
	case <-ctx.Done():
		slog.Info("shutting down daemon...")
	}

	slog.Info("daemon stopped")
}

// TODO: после того как daemonCmd заработает, не забудь почистить main.go —
//       там должно остаться только: slog-дефолты + cmd.Execute().
//       Вся эта логика отсюда переезжает сюда.

func init() {
	rootCmd.AddCommand(daemonCmd)
}
