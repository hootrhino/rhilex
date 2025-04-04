package netmanager

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	Interface string
	Timeout   time.Duration
	Script    string
}

type LeaseInfo struct {
	IPAddress string
	Gateway   string
	DNS       []string
}

// UDHCPCController controls udhcpc operations
type UDHCPCController struct {
	Config Config
}

// NewController creates a new UDHCPCController
func NewController(cfg Config) *UDHCPCController {
	return &UDHCPCController{
		Config: cfg,
	}
}
func (c *UDHCPCController) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// Start runs the udhcpc process with given config
func (c *UDHCPCController) Start() error {
	if c.Config.Interface == "" {
		return errors.New("interface name is required")
	}

	args := []string{"-i", c.Config.Interface}
	if c.Config.Script != "" {
		args = append(args, "-s", c.Config.Script)
	}
	if c.Config.Timeout > 0 {
		args = append(args, "-T", fmt.Sprintf("%d", int(c.Config.Timeout.Seconds())))
	}

	out, err := c.Run("udhcpc", args...)
	if err != nil {
		return fmt.Errorf("udhcpc failed: %v\noutput: %s", err, string(out))
	}

	return nil
}

// Release removes DHCP lease from the interface
func (c *UDHCPCController) Release() error {
	if c.Config.Interface == "" {
		return errors.New("interface name is required")
	}
	_, err := c.Run("udhcpc", "-i", c.Config.Interface, "-n", "-q", "-R")
	return err
}

// GetLeaseInfo fetches IP address, gateway, and DNS info
func (c *UDHCPCController) GetLeaseInfo() (*LeaseInfo, error) {
	if c.Config.Interface == "" {
		return nil, errors.New("interface name is required")
	}

	// Get IP
	out, err := c.Run("ip", "-4", "addr", "show", "dev", c.Config.Interface)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP info: %v", err)
	}
	var ip string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "inet ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ip = strings.Split(fields[1], "/")[0]
				break
			}
		}
	}
	if ip == "" {
		return nil, errors.New("no IP address found")
	}

	// Get gateway
	routeOut, err := c.Run("ip", "route", "show", "default", "dev", c.Config.Interface)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway info: %v", err)
	}
	var gw string
	fields := strings.Fields(strings.TrimSpace(string(routeOut)))
	if len(fields) >= 3 && fields[0] == "default" && fields[1] == "via" {
		gw = fields[2]
	}

	// Get DNS
	dnsList := []string{}
	resolvOut, err := c.Run("cat", "/etc/resolv.conf")
	if err == nil {
		for _, line := range strings.Split(string(resolvOut), "\n") {
			if strings.HasPrefix(line, "nameserver") {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					dnsList = append(dnsList, parts[1])
				}
			}
		}
	}

	return &LeaseInfo{
		IPAddress: ip,
		Gateway:   gw,
		DNS:       dnsList,
	}, nil
}
