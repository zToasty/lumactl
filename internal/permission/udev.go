package permission

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

const udevRulePath = "/etc/udev/rules.d/99-lumactl.rules"

// InstallUdevRule создает файл правил udev для заданных PID и VID
func InstallUdevRule(vid, pid string) error {
	// Формируем строку правила. MODE="0666" дает права на чтение и запись всем пользователям.
	rule := fmt.Sprintf(`SUBSYSTEM=="tty", ATTRS{idVendor}=="%s", ATTRS{idProduct}=="%s", MODE="0666"`, vid, pid)

	slog.Info("Генерация udev-правила", "path", udevRulePath, "vid", vid, "pid", pid)

	// Пытаемся записать файл. Если программа запущена без sudo, здесь мы получим ошибку прав.
	err := os.WriteFile(udevRulePath, []byte(rule+"\n"), 0644)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("нет прав для записи в %s. Пожалуйста, запустите команду через sudo:\n👉 sudo lumactl setup", udevRulePath)
		}
		return fmt.Errorf("ошибка записи udev правила: %w", err)
	}

	slog.Info("Правило успешно записано. Применение настроек...")

	// Выполняем `udevadm control --reload-rules`, чтобы ядро увидело новый файл
	cmdReload := exec.Command("udevadm", "control", "--reload-rules")
	if err := cmdReload.Run(); err != nil {
		return fmt.Errorf("ошибка при перезагрузке правил (reload-rules): %w", err)
	}

	// Выполняем `udevadm trigger`, чтобы правило применилось к уже воткнутому устройству
	cmdTrigger := exec.Command("udevadm", "trigger")
	if err := cmdTrigger.Run(); err != nil {
		return fmt.Errorf("ошибка при применении правил (trigger): %w", err)
	}

	slog.Info("Настройка udev завершена! Теперь lumactl работает без sudo.")
	return nil
}
