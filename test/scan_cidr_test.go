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

package test

import (
	"testing"
	"time"

	"github.com/hootrhino/rhilex/component/shellymanager"
)

// go test -timeout 30s -run ^Test_scan_cidr github.com/hootrhino/rhilex/test -v -count=1
func Test_scan_cidr(t *testing.T) {
	cidr := "192.168.1.0/24"
	timeout := 5 * time.Second
	devices, err := shellymanager.ScanCIDR(cidr, timeout)
	if err != nil {
		t.Log("Error:", err)
		return
	}
	for _, device := range devices {
		t.Log("Devices found:", device)
	}
}
