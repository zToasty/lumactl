package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zToasty/lumactl/internal/config"
)

// Объявляем глобальные переменные для всего пакета cmd
var (
	cfg     *config.Config
	rootCmd = &cobra.Command{
		Use:   "lumactl",
		Short: "Управление контроллерами Adalight (Skydimo) в Linux",
		Long:  `lumactl — это легковесный демон для синхронизации подсветки монитора.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Если вызвали lumactl без команд, просто показываем помощь
			cmd.Help()
		},
	}
)

// Execute вызывается из main.go и запускает всё приложение
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Эта функция запускается самой первой в пакете cmd
func init() {
	// Говорим Кобре: "Перед выполнением любой команды запусти initConfig"
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("CRITICAL_ERROR while config initialization: %v\n", err)
		os.Exit(1)
	}
	setupLogger(cfg.Logging.Level, cfg.Logging.Format)
}

func setupLogger(levelStr, format string) {
	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	slog.SetDefault(slog.New(handler))
}
