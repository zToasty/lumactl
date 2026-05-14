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

var musicCmd = &cobra.Command{
	Use:   "music",
	Short: "Включить режим цветомузыки",
	Long:  `Переключает контроллер в режим захвата системного звука и визуализации громкости.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Формируем команду для демона
		payload := daemon.Command{
			Action: "switch",
			Effect: "liquid_spectrum",
			Params: json.RawMessage("{}"), // Пока без доп. параметров
		}

		data, err := json.Marshal(payload)
		if err != nil {
			slog.Error("failed to encode command", "error", err)
			os.Exit(1)
		}

		// 2. Подключаемся к сокету
		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			slog.Error("Failed to connect to daemon", "error", err)
			os.Exit(1)
		}
		defer conn.Close()

		// 3. Отправляем команду
		if _, err := conn.Write(data); err != nil {
			slog.Error("failed to send command to daemon", "error", err)
			os.Exit(1)
		}

		// 4. Читаем подтверждение (так как мы внедрили Request-Response)
		var resp daemon.Response
		if err := json.NewDecoder(conn).Decode(&resp); err != nil {
			slog.Error("failed to read daemon response", "error", err)
			os.Exit(1)
		}

		if resp.Status == "ok" {
			fmt.Println("🎶 Music reactive mode enabled!")
		} else {
			fmt.Printf("❌ Failed to enable music mode: %s\n", resp.Message)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(musicCmd)
}
