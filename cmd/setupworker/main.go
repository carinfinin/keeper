package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	jobName    = "keeper-worker-cron"
	binPath    = "/usr/local/bin/keeper-worker" // Путь к бинарнику
	cronConfig = "*/1 * * * *"
)

func main() {
	switch runtime.GOOS {
	case "linux", "darwin":
		setupUnixCron()
	case "windows":
		setupWindowsScheduler()
	default:
		fmt.Printf("Unsupported OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}
}

// Настройка cron для Linux/macOS
func setupUnixCron() {
	cronJob := fmt.Sprintf("%s %s >> /var/log/%s.log 2>&1", cronConfig, binPath, jobName)

	// Для macOS нужно добавить переменные окружения
	if runtime.GOOS == "darwin" {
		cronJob = fmt.Sprintf("PATH=/usr/local/bin:/usr/bin:/bin\n%s", cronJob)
	}

	// Добавляем перенос строки в конце
	if !strings.HasSuffix(cronJob, "\n") {
		cronJob += "\n"
	}

	// Временный файл для cron
	tmpFile := fmt.Sprintf("/tmp/%s.cron", jobName)
	if err := os.WriteFile(tmpFile, []byte(cronJob), 0644); err != nil {
		fmt.Printf("Error creating cron file: %v\n", err)
		os.Exit(1)
	}

	// Безопасное добавление в crontab
	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		`(crontab -l 2>/dev/null | grep -vF "%s"; 
        cat "%s") | crontab -`,
		jobName, tmpFile))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error adding to crontab: %v\n", err)
		// Для отладки выводим содержимое файла
		if data, err := os.ReadFile(tmpFile); err == nil {
			fmt.Printf("Cron content that failed to install:\n%s\n", data)
		}
		os.Exit(1)
	}

	os.Remove(tmpFile)
	fmt.Printf("Cron job '%s' successfully installed\n", jobName)
}

// Настройка планировщика задач для Windows
func setupWindowsScheduler() {
	// Полный путь к бинарнику
	exePath, err := filepath.Abs(binPath)
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	// Создаем задание в планировщике
	cmd := exec.Command("schtasks", "/Create",
		"/TN", jobName,
		"/SC", "MINUTE",
		"/MO", "1", // Интервал в минутах (минимальный)
		"/TR", exePath,
		"/F") // Принудительное создание

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error creating scheduled task: %v\n", err)
	}

	fmt.Printf("Scheduled task '%s' successfully installed\n", jobName)
}

func runCommand(name string, arg ...string) error {
	fmt.Printf("Executing: %s %s\n", name, strings.Join(arg, " "))

	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
