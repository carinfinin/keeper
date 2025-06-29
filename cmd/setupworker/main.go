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
	cronConfig = "*/20 * * * * *"               // Каждые 20 секунд
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

	// Временный файл для cron
	tmpFile := fmt.Sprintf("/tmp/%s.cron", jobName)
	if err := os.WriteFile(tmpFile, []byte(cronJob), 0644); err != nil {
		fmt.Printf("Error creating cron file: %v\n", err)
		os.Exit(1)
	}

	// Добавляем в crontab
	cmd := exec.Command("bash", "-c", fmt.Sprintf("crontab -l | grep -v '%s' | cat - %s | crontab -", jobName, tmpFile))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error adding to crontab: %v\n", err)
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

		// Альтернатива через PowerShell
		psScript := fmt.Sprintf(
			`$action = New-ScheduledTaskAction -Execute '%s';
			$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Minutes 1);
			Register-ScheduledTask -TaskName '%s' -Action $action -Trigger $trigger -Force`,
			exePath, jobName)

		psCmd := exec.Command("powershell", "-Command", psScript)
		psCmd.Stdout = os.Stdout
		psCmd.Stderr = os.Stderr
		if psErr := psCmd.Run(); psErr != nil {
			fmt.Printf("PowerShell fallback also failed: %v\n", psErr)
			os.Exit(1)
		}
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
