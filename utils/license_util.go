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

package utils

import "fmt"

// FetchLoadLicense rhilex active -H 127.0.0.1 -U admin -P 123456
func FetchLoadLicense(host, username, password, macAddr string) error {
	activeParams := fmt.Sprintf(`%s&%s&%s&%s&0&0`,
		"SN000001", username, password, macAddr)
	CLog("\n*>> BEGIN LICENCE ACTIVE\n"+
		"*# Vendor Admin: (%s, %s)\n"+
		"*# Local Mac Address: (%s)\n"+
		"*# Try to request license from server:(%s) ...\n",
		username, password, macAddr, host)
	err := Download(host, activeParams)
	if err != nil {
		return fmt.Errorf("[LICENCE ACTIVE]: Download license failed, error:%s", err)
	}
	fmt.Println("*# License fetch success, save as: license.zip")
	fmt.Println("*<< END LICENCE ACTIVE")
	return nil
}
