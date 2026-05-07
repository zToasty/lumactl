package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/zToasty/lumactl/internal/permission"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Настроить права доступа к USB-устройству (требует sudo)",
	Long: `Команда setup генерирует udev-правило для Linux,
чтобы текущий пользователь мог взаимодействовать с контроллером подсветки
без необходимости запускать программу от имени root.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Здесь мы предполагаем, что конфиг уже загружен в rootCmd
		// Если конфига по какой-то причине нет, используем дефолтные значения
		vid := "1a86"
		pid := "7523"
		if cfg != nil {
			vid = cfg.Device.VID
			pid = cfg.Device.PID
		}

		fmt.Println("🔧 Начинаем настройку прав доступа...")

		if err := permission.InstallUdevRule(vid, pid); err != nil {
			slog.Error("Сбой настройки", "error", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	// Регистрируем команду setup внутри главной команды (root)
	rootCmd.AddCommand(setupCmd)
}
