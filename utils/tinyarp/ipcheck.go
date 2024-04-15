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

package tinyarp

import (
	"net"
	"strings"
)

func IsValidIP(line string) bool {
	Ip := net.ParseIP(line)
	if Ip == nil {
		return false
	}
	if isClassC(line) && !isLocalhost(line) && !isMulticast(line) && !isBroadcast(line) {
		return true
	}

	return false
}

// 判断是否为C类IP地址
func isClassC(ip string) bool {
	firstOctet := strings.Split(ip, ".")[0]
	return firstOctet == "192" || firstOctet == "193" || firstOctet == "194" || firstOctet == "195" ||
		firstOctet == "196" || firstOctet == "197" || firstOctet == "198" || firstOctet == "199"
}

// 判断是否为本地回环地址
func isLocalhost(ip string) bool {
	return strings.HasPrefix(ip, "127.")
}

// 判断是否为多播地址
func isMulticast(ip string) bool {
	return strings.HasPrefix(ip, "224.")
}

// 判断是否为广播地址
func isBroadcast(ip string) bool {
	return strings.HasPrefix(ip, "255.")
}
