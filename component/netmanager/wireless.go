package netmanager

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

type WirelessManager struct{}

// Set WIFI configuration
func (w *WirelessManager) SetConfig(params NetworkInterfaceParam) error {
	if params.Type != "WLAN" {
		return fmt.Errorf("invalid interface type for WirelessManager")
	}

	tmpl := `ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=CN

network={
        ssid="{{.SSID}}"
        psk="{{.Password}}"
}`

	t := template.Must(template.New("wifiConfig").Parse(tmpl))

	filename := fmt.Sprintf("/etc/wpa_supplicant/wpa_supplicant-%s.conf", params.Interface)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create wpa_supplicant config file: %v", err)
	}
	defer file.Close()

	err = t.Execute(file, params)
	if err != nil {
		return fmt.Errorf("failed to write wpa_supplicant config: %v", err)
	}

	// Restart wpa_supplicant service
	err = restartWpaSupplicant(params.Interface)
	if err != nil {
		return fmt.Errorf("failed to restart wpa_supplicant: %v", err)
	}

	// Apply DHCP
	err = applyDHCP(params.Interface)
	if err != nil {
		return fmt.Errorf("failed to apply DHCP: %v", err)
	}

	return nil
}

func restartWpaSupplicant(iface string) error {
	cmd := exec.Command("systemctl", "restart", fmt.Sprintf("wpa_supplicant@%s.service", iface))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart wpa_supplicant: %v, output: %s", err, string(output))
	}
	return nil
}

func applyDHCP(iface string) error {
	cmd := exec.Command("dhclient", iface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply DHCP: %v, output: %s", err, string(output))
	}
	return nil
}
