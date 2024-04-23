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

package shellymanager

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ScanCIDR scans the devices in the given CIDR range within the specified timeout.
func ScanCIDR(cidr string, timeout time.Duration) ([]string, error) {
	IP, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	for ip := IP.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			if CheckHTTP(ip, timeout) {
				mu.Lock()
				ips = append(ips, ip)
				mu.Unlock()
			}
		}((ip.String()))
	}

	wg.Wait()

	return ips, nil
}

// Increment IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// Check if HTTP endpoint returns HTTP 200 OK
func CheckHTTP(ip string, timeout time.Duration) bool {
	client := &http.Client{
		Timeout: timeout,
	}
	url := fmt.Sprintf("http://%s/rpc/Shelly.GetDeviceInfo", ip)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return strings.Contains(string(body), `model`)
}
func IsPortOpen(ip string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err == nil {
		defer conn.Close()
		return true
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return false
	}
	return false
}
