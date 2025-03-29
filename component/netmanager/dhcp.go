package netmanager

import (
	"fmt"
	"os/exec"
)

type TinyDHCPServer struct {
	Interface string
}

// Set DHCP server configuration
func (d *TinyDHCPServer) SetConfig(ipRangeStart, ipRangeEnd, leaseTime, gateway string, staticLeases map[string]string) error {
	conf := fmt.Sprintf("interface=%s\n", d.Interface)
	conf += fmt.Sprintf("dhcp-range=%s,%s,%s\n", ipRangeStart, ipRangeEnd, leaseTime)
	conf += fmt.Sprintf("dhcp-option=3,%s\n", gateway)

	for mac, ip := range staticLeases {
		conf += fmt.Sprintf("dhcp-host=%s,%s\n", mac, ip)
	}

	err := writeDnsmasqConfig(conf)
	if err != nil {
		return fmt.Errorf("failed to write dnsmasq config: %v", err)
	}

	err = restartDnsmasq()
	if err != nil {
		return fmt.Errorf("failed to restart dnsmasq: %v", err)
	}

	return nil
}

func writeDnsmasqConfig(conf string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' > /etc/dnsmasq.conf", conf))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to write dnsmasq.conf: %v, output: %s", err, string(output))
	}
	return nil
}

func restartDnsmasq() error {
	cmd := exec.Command("systemctl", "restart", "dnsmasq")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart dnsmasq: %v, output: %s", err, string(output))
	}
	return nil
}
