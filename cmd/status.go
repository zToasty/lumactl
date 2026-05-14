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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Показать текущее состояние демона",
	Long:  `Запрашивает у запущенного демона информацию о текущем активном эффекте и состоянии системы.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Подключаемся к сокету
		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			slog.Error("Failed to connect to daemon. Is it running?", "error", err)
			os.Exit(1)
		}
		defer conn.Close()

		// 2. Отправляем запрос статуса
		payload := daemon.Command{Action: "status"}
		if err := json.NewEncoder(conn).Encode(payload); err != nil {
			slog.Error("Failed to send status request", "error", err)
			os.Exit(1)
		}

		// 3. Читаем ответ от демона
		var resp daemon.Response
		if err := json.NewDecoder(conn).Decode(&resp); err != nil {
			slog.Error("Failed to decode daemon response", "error", err)
			os.Exit(1)
		}

		// 4. Выводим информацию пользователю
		if resp.Status != "ok" {
			fmt.Printf("❌ Ошибка демона: %s\n", resp.Message)
			os.Exit(1)
		}

		// Парсим данные из Data (там лежит {"active_effect": "..."})
		var statusData map[string]string
		if err := json.Unmarshal(resp.Data, &statusData); err != nil {
			slog.Debug("Failed to parse status data", "error", err)
			fmt.Println("✔ Daemon is alive")
			return
		}

		effect := statusData["active_effect"]
		fmt.Println("📊 Состояние lumactl:")
		fmt.Printf("  • Статус: %s\n", resp.Status)
		fmt.Printf("  • Текущий эффект: %s\n", effect)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
