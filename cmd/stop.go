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

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Остановить активный эффект и выключить все светодиоды",
	Long:  `Отправляет демону команду stop. Светодиоды гаснут, активная анимация завершается.`,
	Run: func(cmd *cobra.Command, args []string) {
		payload := daemon.Command{Action: "stop"}
		data, err := json.Marshal(payload)
		if err != nil {
			slog.Error("failed to marshal command", "error", err)
			os.Exit(1)
		}

		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			slog.Error("failed to connect to daemon", "error", err)
			os.Exit(1)
		}
		defer conn.Close()

		if _, err := conn.Write(data); err != nil {
			slog.Error("failed to send command", "error", err)
			os.Exit(1)
		}

		fmt.Println("effect stopped, LEDs off")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
