// Copyright (C) 2024 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package en6400

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const wifiConfigTemplate = `#
# GENERATED BY RHILEX, DON'T MODIFY THIS CONFIG!!!
#
update_config=1

network={
    priority=1
    ssid="{{.SSID}}"
    psk="{{.PSK}}"
    key_mgmt=WPA-PSK
    proto=RSN
    pairwise=CCMP
    auth_alg=OPEN
}
`

// isWirelessInterface checks if the given interface name corresponds to a wireless interface.
func isWirelessInterface(ifName string) bool {
	// On Linux, wireless interfaces typically have a directory under /sys/class/net/<iface>/wireless
	_, err := os.Stat(fmt.Sprintf("/sys/class/net/%s/wireless", ifName))
	return !os.IsNotExist(err)
}

// getWlanList returns a list of wireless interfaces.
func getWlanList() ([]net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var wlanIfaces []net.Interface
	for _, iface := range ifaces {
		if isWirelessInterface(iface.Name) {
			wlanIfaces = append(wlanIfaces, iface)
		}
	}

	return wlanIfaces, nil
}

func SetWifi(iface, ssid, psk string, timeout time.Duration) error {
	if len(psk) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	configData := struct {
		SSID string
		PSK  string
	}{
		SSID: ssid,
		PSK:  psk,
	}
	tmpl, err := template.New("wifiConfig").Parse(wifiConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}
	configFilePath := fmt.Sprintf("/etc/wpa_supplicant/wpa_supplicant-%s.conf", iface)
	// configFilePath := fmt.Sprintf("./data/wpa_supplicant-%s.conf", iface)
	file, err := os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()
	err = tmpl.Execute(file, configData)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	{

		cmd := exec.CommandContext(ctx, "wpa_supplicant", "-B", "-i", iface, "-c", configFilePath)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start wpa_supplicant: %v", err)
		}
	}
	{
		// dhclient wlx0cc6551c5026
		cmd := exec.CommandContext(ctx, "dhclient", iface)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start dhclient: %v", err)
		}
	}
	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

/*
*
* 升级版，带上了WIFI信号强度
*
 */
func ScanWlanList(WFace string) ([][2]string, error) {
	wifiList := [][2]string{}
	shell := `
iw dev %s scan | awk '
  /SSID/ { ssid=$2 } \
  /signal/ { signal=$2; if (!seen[ssid] || signal > seen[ssid]) { seen[ssid]=signal } } \
  END { for (s in seen) print s "," seen[s]}
' | sort
`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf(shell, WFace))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error executing nmcli: %s", err.Error()+":"+string(output))
	}
	lines := bufio.NewScanner(strings.NewReader(string(output)))
	for lines.Scan() {
		line := lines.Text()
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			ssid := parts[0]
			signal := parts[1]
			number, _ := strconv.ParseFloat(signal, 64)
			if ssid != "" {
				wifiList = append(wifiList, [2]string{ssid, fmt.Sprintf("%v", math.Abs(number))})
			}
		}
	}
	return wifiList, nil
}
