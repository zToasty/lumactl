package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/zToasty/lumactl/internal/daemon"
)

var colorCmd = &cobra.Command{
	Use:   "color [color_name_or_hex]",
	Short: "Set static color (red, green, blue, white, off, #FF00FF)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		colorArg := args[0]

		var payload daemon.Command

		// Обрабатываем команду выключения красиво
		if colorArg == "off" || colorArg == "black" {
			payload = daemon.Command{
				Action: "stop",
			}
		} else {
			// Все остальные цвета (имена или HEX) отправляем в эффект static
			payload = daemon.Command{
				Action: "switch",
				Effect: "static",
				Params: json.RawMessage(fmt.Sprintf(`{"color": "%s"}`, colorArg)),
			}
		}

		// 1. Пакуем в JSON
		data, err := json.Marshal(payload)
		if err != nil {
			slog.Error("failed to encode command", "error", err)
			os.Exit(1)
		}

		// 2. Подключаемся к демону (как к локальному серверу)
		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			slog.Error("Failed to connect to daemon. Is it running?", "error", err)
			fmt.Println("Подсказка: запустите сервер командой ./lumactl (или ./lumactl daemon)")
			os.Exit(1)
		}
		defer conn.Close()

		// 3. Отправляем приказ и сразу завершаем работу
		if _, err := conn.Write(data); err != nil {
			slog.Error("failed to send command to daemon", "error", err)
			os.Exit(1)
		}

		slog.Info("Command sent to daemon", "color", colorArg)
	},
}

func init() {
	rootCmd.AddCommand(colorCmd)
}
