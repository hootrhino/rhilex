package netmanager

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

type EthernetManager struct{}

// Set network configuration
func (e *EthernetManager) SetConfig(params NetworkInterfaceParam) error {
	if params.Type != "ETH" {
		return fmt.Errorf("invalid interface type for EthernetManager")
	}

	tmpl := `auto {{.Interface}}
{{if .DHCPEnabled}}iface {{.Interface}} inet dhcp{{else}}iface {{.Interface}} inet static
address {{.Address}}
netmask {{.Netmask}}
gateway {{.Gateway}}{{end}}`

	t := template.Must(template.New("ethernetConfig").Parse(tmpl))

	filename := fmt.Sprintf("/etc/network/interfaces.d/%s.conf", params.Interface)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create ethernet config file: %v", err)
	}
	defer file.Close()

	err = t.Execute(file, params)
	if err != nil {
		return fmt.Errorf("failed to write ethernet config: %v", err)
	}

	// Apply changes (restart networking) - This may vary depending on distribution
	err = restartNetwork()
	if err != nil {
		return fmt.Errorf("failed to restart networking: %v", err)
	}

	return nil
}

func restartNetwork() error {
	// Example for Debian/Ubuntu
	cmd := exec.Command("systemctl", "restart", "networking")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart networking: %v, output: %s", err, string(output))
	}
	return nil
}
