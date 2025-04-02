package netmanager

import (
	"os"
	"strings"
	"testing"
)

func TestConfigureNetworkInterface(t *testing.T) {
	// 1. 创建临时目录，作为测试配置目录
	configDir := "./test_config"
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	defer os.RemoveAll(configDir)
	// 2. 定义各种网络配置
	testCases := []struct {
		name   string
		params NetworkInterfaceParam
		errMsg string //期望的错误信息，如果希望成功则为空
	}{
		{
			name: "WLAN DHCP",
			params: NetworkInterfaceParam{
				Type:        "WLAN",
				Interface:   "wlan0",
				DHCPEnabled: true,
				SSID:        "MyWiFiNetwork",
				Password:    "MyWiFiPassword",
				SecureType:  "WPA2",
				Hidden:      false,
				Channel:     6,
			},
		},
		{
			name: "WLAN Static IP",
			params: NetworkInterfaceParam{
				Type:        "WLAN",
				Interface:   "wlan0",
				DHCPEnabled: false,
				SSID:        "MyWiFiNetwork",
				Password:    "MyWiFiPassword",
				SecureType:  "WPA2",
				Hidden:      false,
				Channel:     6,
				Address:     "192.168.2.100",
				Netmask:     "255.255.255.0",
				Gateway:     "192.168.2.1",
			},
		},
		{
			name: "ETH DHCP",
			params: NetworkInterfaceParam{
				Type:        "ETH",
				Interface:   "eth0",
				DHCPEnabled: true,
			},
		},
		{
			name: "ETH Static IP",
			params: NetworkInterfaceParam{
				Type:            "ETH",
				Interface:       "eth0",
				Address:         "192.168.1.100",
				Netmask:         "255.255.255.0",
				Gateway:         "192.168.1.1",
				DHCPEnabled:     false,
				EthernetSpeed:   "1000Mbps", // 以太网相关的
				EthernetDuplex:  "full",
				EthernetAutoNeg: true,
			},
		},
		{
			name: "Invalid Type",
			params: NetworkInterfaceParam{
				Type:      "INVALID",
				Interface: "invalid0",
			},
			errMsg: "invalid interface type: INVALID",
		},
	}

	// 3. 循环执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ConfigureNetworkInterface(true, tc.params, configDir)
			if tc.errMsg != "" {
				if err == nil {
					t.Errorf("Expected error: %s, but got nil", tc.errMsg)
				} else if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("Expected error: %s, but got: %v", tc.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// 可以添加代码来检查配置文件是否正确生成
				// 例如：读取文件内容，并验证是否包含预期的配置
			}
		})
	}
}
