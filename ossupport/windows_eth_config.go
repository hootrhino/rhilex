package ossupport

import (
	"os/exec"
	"strings"
)

// GetWindowsMACAddress 获取网卡MAC
func GetWindowsMACAddress() (string, error) {
	cmd := exec.Command("powershell.exe", "-Command",
		`wmic nicconfig where "Index=1" get MACAddress`)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	macAddress := strings.TrimSpace(string(output))
	if strings.Contains(macAddress, "MACAddress") {
		macAddress = strings.Split(macAddress, "\n")[1]
	}
	return strings.ToUpper(macAddress), nil
}
