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

import "strings"

// IsValidMacAddress1: AA:BB:CC:DD:EE:FF
func IsValidMacAddress1(macAddress string) bool {
	if len(macAddress) != 17 || !strings.Contains(macAddress, ":") {
		return false
	}
	for i := 0; i < 11; i += 3 { // 跳过冒号
		byteStr := macAddress[i : i+3]
		if !isValidHex(byteStr) {
			return false
		}
	}

	return true
}

// IsValidMacAddress2: AABBCCDDEEFF
func IsValidMacAddress2(macAddress string) bool {
	if len(macAddress) != 12 {
		return false
	}
	for i := 0; i < len(macAddress); i += 2 {
		byteStr := macAddress[i : i+2]
		if !isValidHex(byteStr) {
			return false
		}
	}

	return true
}
func isValidHex(hexStr string) bool {
	for _, char := range hexStr {
		if !(char >= '0' && char <= '9' ||
			char >= 'a' && char <= 'f' ||
			char >= 'A' && char <= 'F') {
			return false
		}
	}
	return true
}
