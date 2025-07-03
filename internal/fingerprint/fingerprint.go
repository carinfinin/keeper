// fingerprint помогает идентифицировать клиентв без необходимоти их хранения на клинте.
package fingerprint

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// DeviceInfo собирает основные характеристики
type DeviceInfo struct {
	OS       string
	Arch     string
	CPU      string
	Hostname string
}

// Get собирает данные устройства
func Get() DeviceInfo {
	info := DeviceInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPU:      getCPUInfo(),
		Hostname: getHostname(),
	}

	return info
}

// GenerateHash создает уникальный fingerprint
func (d DeviceInfo) GenerateHash() string {
	data := fmt.Sprintf(
		"%s|%s|%s|%s",
		d.OS,
		d.Arch,
		d.CPU,
		d.Hostname,
	)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:10]
}
func getHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}

var cpuModel string
var cpuOnce sync.Once

func getCPUInfo() string {
	cpuOnce.Do(func() {
		switch runtime.GOOS {
		case "linux":
			data, _ := os.ReadFile("/proc/cpuinfo")
			cpuModel = parseCPUInfo(string(data))
		case "darwin":
			out, _ := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
			cpuModel = strings.TrimSpace(string(out))
		case "windows":
			out, _ := exec.Command("wmic", "cpu", "get", "name").Output()
			parts := strings.Split(string(out), "\n")
			if len(parts) > 1 {
				cpuModel = strings.TrimSpace(parts[1])
			}
		default:
			cpuModel = "unknown"
		}

		if cpuModel == "" {
			cpuModel = "unknown"
		}
	})
	return cpuModel
}

func parseCPUInfo(data string) string {
	for _, line := range strings.Split(data, "\n") {
		if strings.Contains(line, "model name") || strings.Contains(line, "Model") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}
