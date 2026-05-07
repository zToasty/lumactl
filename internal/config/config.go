package config

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Вшиваем дефолтный конфиг в бинарник
//
//go:embed default_config.yaml
var defaultConfigYAML []byte

type Config struct {
	Device  DeviceConfig  `mapstructure:"device"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type DeviceConfig struct {
	Port     string `mapstructure:"port"`
	VID      string `mapstructure:"vid"`
	PID      string `mapstructure:"pid"`
	LEDCount int    `mapstructure:"led_count"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// LoadConfig инициализирует настройки, проверяя наличие файла в ~/.config/lumactl
func LoadConfig() (*Config, error) {
	// 1. Определяем пути
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine user config directory: %w", err)
	}

	appConfigDir := filepath.Join(configRoot, "lumactl")
	configPath := filepath.Join(appConfigDir, "config.yaml")

	// 2. Гарантируем наличие файла
	if err := ensureConfigExists(appConfigDir, configPath); err != nil {
		return nil, err
	}

	// 3. Настройка Viper
	viper.SetConfigFile(configPath)
	viper.SetEnvPrefix("LUMACTL")
	viper.AutomaticEnv()

	// Дефолты в памяти на случай "дырявого" YAML
	setDefaults()

	// 4. Чтение и парсинг
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	slog.Debug("configuration loaded", "path", configPath)
	return &cfg, nil
}

// ensureConfigExists создает папку и файл, если их нет
func ensureConfigExists(dir, path string) error {
	// Создаем папку (MkdirAll не вернет ошибку, если она уже есть)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	// Проверяем файл
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		slog.Info("config file not found, generating default", "path", path)
		if err := os.WriteFile(path, defaultConfigYAML, 0644); err != nil {
			return fmt.Errorf("failed to write default config to %s: %w", path, err)
		}
		// Печатаем через fmt, так как логгер может быть еще не настроен на нужный уровень
		fmt.Printf("💡 Default config created at: %s\n", path)
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("device.port", "auto")
	viper.SetDefault("device.vid", "1a86")
	viper.SetDefault("device.pid", "7523")
	viper.SetDefault("device.led_count", 255)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")
}
