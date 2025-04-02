package netmanager

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type NetworkController struct {
	Interface string
}

// Get network interface state (up/down)
func (n *NetworkController) GetState() (string, error) {
	cmd := exec.Command("ip", "link", "show", n.Interface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get network state: %v, output: %s", err, string(output))
	}

	if strings.Contains(string(output), "state UP") {
		return "UP", nil
	}
	return "DOWN", nil
}

// Get network interface speed
func (n *NetworkController) GetSpeed() (string, error) {
	cmd := exec.Command("ethtool", n.Interface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get network speed: %v, output: %s", err, string(output))
	}

	re := regexp.MustCompile(`Speed: (\d+Mb/s)`)
	match := re.FindStringSubmatch(string(output))
	if len(match) > 1 {
		return match[1], nil
	}
	return "Unknown", nil
}

// Get network interface IP address
func (n *NetworkController) GetIPAddress() (string, error) {
	cmd := exec.Command("ip", "addr", "show", n.Interface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get IP address: %v, output: %s", err, string(output))
	}

	re := regexp.MustCompile(`inet (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/\d+`)
	match := re.FindStringSubmatch(string(output))
	if len(match) > 1 {
		return match[1], nil
	}
	return "Unknown", nil
}

// Get network interface MAC address
func (n *NetworkController) GetMACAddress() (string, error) {
	cmd := exec.Command("ip", "link", "show", n.Interface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get MAC address: %v, output: %s", err, string(output))
	}

	re := regexp.MustCompile(`link/ether ([0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2})`)
	match := re.FindStringSubmatch(string(output))
	if len(match) > 1 {
		return match[1], nil
	}
	return "Unknown", nil
}

// Get network interface statistics
func (n *NetworkController) GetStatistics() (string, error) {
	cmd := exec.Command("ip", "addr", "show", n.Interface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get statistics: %v, output: %s", err, string(output))
	}

	return string(output), nil
}

// Get network interface receive bytes
func (n *NetworkController) GetReceiveBytes() (uint64, error) {
	cmd := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", n.Interface))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to get receive bytes: %v, output: %s", err, string(output))
	}
	bytesStr := strings.TrimSpace(string(output))
	bytes, err := strconv.ParseUint(bytesStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse receive bytes: %v", err)
	}
	return bytes, nil
}

// Get network interface transmit bytes
func (n *NetworkController) GetTransmitBytes() (uint64, error) {
	cmd := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", n.Interface))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to get transmit bytes: %v, output: %s", err, string(output))
	}
	bytesStr := strings.TrimSpace(string(output))
	bytes, err := strconv.ParseUint(bytesStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse transmit bytes: %v", err)
	}
	return bytes, nil
}

// Restart network interface
func (n *NetworkController) RestartNetwork() error {
	cmd := exec.Command("systemctl", "restart", "networking") // For Debian/Ubuntu
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart network: %v, output: %s", err, string(output))
	}
	return nil
}
