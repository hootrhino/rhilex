package netmanager

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
)

// NetworkInterfaceParam 结构体 (公开，供外部调用)
type NetworkInterfaceParam struct {
	Type        string `json:"type"` // "ETH" or "WLAN"
	Interface   string `json:"interface"`
	Address     string `json:"address"`
	Netmask     string `json:"netmask"`
	Gateway     string `json:"gateway"`
	DHCPEnabled bool   `json:"dhcp_enabled"`
	SSID        string `json:"ssid,omitempty"`
	Password    string `json:"password,omitempty"`
	SecureType  string `json:"secure_type,omitempty"`
	Hidden      bool   `json:"hidden,omitempty"`
	Channel     int    `json:"channel,omitempty"`
	// 以太网特定参数
	EthernetSpeed   string `json:"ethernet_speed,omitempty"`   // 以太网连接速度，例如 "100Mbps", "1000Mbps"
	EthernetDuplex  string `json:"ethernet_duplex,omitempty"`  // 以太网双工模式，例如 "full", "half"
	EthernetAutoNeg bool   `json:"ethernet_autoneg,omitempty"` //以太网自动协商
}

// to json
func (n *NetworkInterfaceParam) ToJson() string {
	bytes, err := json.Marshal(n)
	if err != nil {
		return ""
	}
	return string(bytes)
}

// ConfigureNetworkInterface 配置网络接口 (公开函数)
func ConfigureNetworkInterface(debug bool, params NetworkInterfaceParam, configDir string) error {
	// 根据接口类型生成配置文件
	var configFile string
	var configContent string
	var err error

	switch params.Type {
	case "WLAN":
		configFile = configDir + "/wpa_supplicant-" + params.Interface + ".conf"
		configContent, err = generateWPAConfig(params)
		if err != nil {
			return fmt.Errorf("failed to generate WLAN config: %w", err)
		}
	case "ETH":
		configFile = configDir + "/interfaces-" + params.Interface // 示例位置，需要根据系统调整
		configContent, err = generateETHConfig(params)
		if err != nil {
			return fmt.Errorf("failed to generate ETH config: %w", err)
		}
	default:
		return fmt.Errorf("invalid interface type: %s", params.Type)
	}
	if err := writeConfigFile(configFile, configContent); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	if debug {
		fmt.Println("configFile:", configFile)
		fmt.Println("configContent:", configContent)
		return nil
	}
	if err := restartNetwork(); err != nil {
		return fmt.Errorf("failed to restart network: %w", err)
	}

	return nil
}

// generateWPAConfig 生成 wpa_supplicant.conf 文件的内容 (内部函数)
func generateWPAConfig(params NetworkInterfaceParam) (string, error) {
	if params.Type != "WLAN" {
		return "", fmt.Errorf("invalid interface type: %s", params.Type)
	}
	config := fmt.Sprintf(`ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=CN

network={
	ssid="%s"
	psk="%s"
	key_mgmt=WPA-PSK
	scan_ssid=%d
}
`, params.SSID, params.Password, boolToInt(params.Hidden))
	return config, nil
}

// generateETHConfig 生成以太网接口配置 (简单的示例，仅适用于静态 IP) (内部函数)
func generateETHConfig(params NetworkInterfaceParam) (string, error) {
	if params.Type != "ETH" {
		return "", fmt.Errorf("invalid interface type: %s", params.Type)
	}

	var config string

	if params.DHCPEnabled {
		// DHCP 配置
		config = fmt.Sprintf(`auto %s
iface %s inet dhcp
`, params.Interface, params.Interface)
	} else {
		// 静态 IP 配置
		config = fmt.Sprintf(`auto %s
iface %s inet static
address %s
netmask %s
gateway %s
`, params.Interface, params.Interface, params.Address, params.Netmask, params.Gateway)
	}

	return config, nil
}

// boolToInt 将布尔值转换为整数 (0 或 1) (内部函数)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// writeConfigFile 将配置文件内容写入到文件中 (内部函数)
func writeConfigFile(filename, content string) error {
	// 这里需要使用更安全的方法写入文件，例如先写入临时文件，然后重命名
	file, err := os.Create(filename) // 这会覆盖现有文件！
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	return nil
}

// restartNetwork 服务，使其跨平台 (内部函数)
func restartNetwork() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		// 尝试使用 systemctl，如果失败，则尝试 ifdown/ifup
		err := trySystemctlRestart("networking")
		if err != nil {
			log.Println("Systemctl restart networking failed:", err)
			err = tryIfdownIfupAll() //尝试全部重启
			if err != nil {
				log.Println("IfdownIfupAll restart networking failed:", err)
				return err
			}
		}
		return nil

	case "windows":
		// Windows 下可以使用 PowerShell 命令
		cmd = exec.Command("powershell", "Restart-Service", "Dnscache") // 示例：重启 DNS 缓存服务
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart networking: %w, output: %s", err, string(output))
	}
	log.Println("Networking restarted successfully")
	return nil
}

// trySystemctlRestart 尝试使用 systemctl 重启服务 (内部函数)
func trySystemctlRestart(serviceName string) error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl restart %s failed: %w, output: %s", serviceName, err, string(output))
	}
	log.Printf("Successfully restarted %s via systemctl\n", serviceName)
	return nil
}

// tryIfdownIfupAll 尝试用ifdown/ifup 命令重启所有接口 (内部函数)
func tryIfdownIfupAll() error {
	cmd := exec.Command("sh", "-c", "ifdown -a && ifup -a")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ifdown -a && ifup -a failed: %w, output: %s", err, string(output))
	}
	log.Printf("Successfully restarted all interfaces via ifdown/ifup\n")
	return nil
}
