package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zToasty/lumactl/internal/discovery"
	"github.com/zToasty/lumactl/internal/effects"
	"github.com/zToasty/lumactl/internal/protocol"
)

var colorCmd = &cobra.Command{
	Use:   "color [color_name]",
	Short: "Set static color (red, green, blue, white, off)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		colorName := args[0]
		var c protocol.RGB

		// 1. Simple color parser
		switch colorName {
		case "red":
			c = protocol.RGB{R: 255, G: 0, B: 0}
		case "green":
			c = protocol.RGB{R: 0, G: 255, B: 0}
		case "blue":
			c = protocol.RGB{R: 0, G: 0, B: 255}
		case "white":
			c = protocol.RGB{R: 255, G: 255, B: 255}
		case "off", "black":
			c = protocol.RGB{R: 0, G: 0, B: 0}
		default:
			slog.Error("Unknown color. Supported colors: red, green, blue, white, off", "color", colorName)
			os.Exit(1)
		}

		// 2. Find port (using logic from config)
		portName := cfg.Device.Port
		if portName == "auto" {
			foundPort, err := discovery.FindDevicePort(cfg.Device.VID, cfg.Device.PID)
			if err != nil {
				slog.Error("Device not found", "error", err)
				os.Exit(1)
			}
			portName = foundPort
		}

		// 3. Connect to the device
		provider := protocol.NewAdalightProvider(portName, cfg.Device.LEDCount)
		if err := provider.Init(); err != nil {
			slog.Error("Failed to connect to port", "error", err)
			os.Exit(1)
		}
		// Guarantee port closing on exit
		defer provider.Close()

		// 4. Context magic and Ctrl+C interception
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Create channel for Linux system signals (SIGINT = Ctrl+C)
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		// Run background goroutine waiting for Ctrl+C
		go func() {
			<-sigs // Blocked here until Ctrl+C
			fmt.Println("\nTermination signal received. Turning off backlight...")
			cancel() // Sends drop to ctx.Done() pipe!
		}()

		// 5. Run effect
		slog.Info("Starting static color", "color", colorName, "port", portName, "leds", cfg.Device.LEDCount)
		effect := &effects.Static{Color: c}

		// Blocks until cancel() is called
		if err := effect.Run(ctx, provider, cfg.Device.LEDCount); err != nil {
			slog.Error("Effect execution failed", "error", err)
		}

		// 6. Turn off LEDs before exiting
		provider.SetColors(make([]protocol.RGB, cfg.Device.LEDCount))
		slog.Info("Backlight turned off")
	},
}

func init() {
	rootCmd.AddCommand(colorCmd)
}
