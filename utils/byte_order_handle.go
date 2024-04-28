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
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"encoding/hex"
	"fmt"
	"math"
)

func ParseBinaryValue(DataBlockType string, DataBlockOrder string,
	Weight float32, byteSlice [256]byte) string {

	if DataBlockType == "UTF8" {
		acc := 0
		for _, v := range byteSlice {
			if v != 0 {
				acc++
			} else {
				continue
			}
		}
		if acc == 0 {
			return ""
		}
		if DataBlockOrder == "BIG_ENDIAN" {
			return string(byteSlice[:acc])
		}
		if DataBlockOrder == "LITTLE_ENDIAN" {
			return stringReverse(string(byteSlice[:acc]))
		}
	}

	if DataBlockType == "RAW" {
		acc := 0
		for _, v := range byteSlice {
			if v != 0 {
				acc++
			} else {
				continue
			}
		}
		if acc == 0 {
			return ""
		}
		return hex.EncodeToString(byteSlice[:acc])
	}

	if DataBlockType == "BYTE" {

		return fmt.Sprintf("%d", byteSlice[0])
	}

	return ""
}

/*
*
*解析西门子的值 无符号
*
 */
func ParseSignedValue(DataBlockType string, DataBlockOrder string,
	Weight float32, byteSlice [256]byte) string {
	if DataBlockType == "SHORT" || DataBlockType == "INT16" {
		// AB: 1234
		// BA: 3412
		if DataBlockOrder == "AB" {
			uint16Value := int16(byteSlice[0])<<8 | int16(byteSlice[1])
			return fmt.Sprintf("%d", uint16Value*int16(Weight))

		}
		if DataBlockOrder == "BA" {
			uint16Value := int16(byteSlice[0]) | int16(byteSlice[1])<<8
			return fmt.Sprintf("%d", uint16Value*int16(Weight))
		}
	}
	if DataBlockType == "INT" || DataBlockType == "INT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := int32(byteSlice[0])<<24 | int32(byteSlice[1])<<16 |
				int32(byteSlice[2])<<8 | int32(byteSlice[3])
			return fmt.Sprintf("%d", intValue*int32(Weight))

		}
		if DataBlockOrder == "CDAB" {
			intValue := int32(byteSlice[0])<<8 | int32(byteSlice[1]) |
				int32(byteSlice[2])<<24 | int32(byteSlice[3])<<16
			return fmt.Sprintf("%d", intValue*int32(Weight))
		}
		if DataBlockOrder == "DCBA" {
			intValue := int32(byteSlice[0]) | int32(byteSlice[1])<<8 |
				int32(byteSlice[2])<<16 | int32(byteSlice[3])<<24
			return fmt.Sprintf("%d", intValue*int32(Weight))
		}
	}

	return ""
}
func ParseUSignedValue(DataBlockType string, DataBlockOrder string,
	Weight float32, byteSlice [256]byte) string {
	if DataBlockType == "USHORT" || DataBlockType == "UINT16" {
		// AB: 1234
		// BA: 3412
		if DataBlockOrder == "AB" {
			uint16Value := uint16(byteSlice[0])<<8 | uint16(byteSlice[1])
			return fmt.Sprintf("%d", uint16Value*uint16(Weight))

		}
		if DataBlockOrder == "BA" {
			uint16Value := uint16(byteSlice[0]) | uint16(byteSlice[1])<<8
			return fmt.Sprintf("%d", uint16Value*uint16(Weight))
		}
	}
	if DataBlockType == "UINT" || DataBlockType == "UINT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := uint32(byteSlice[0])<<24 | uint32(byteSlice[1])<<16 |
				uint32(byteSlice[2])<<8 | uint32(byteSlice[3])
			return fmt.Sprintf("%d", intValue*uint32(Weight))

		}
		if DataBlockOrder == "CDAB" {
			intValue := uint32(byteSlice[0])<<8 | uint32(byteSlice[1]) |
				uint32(byteSlice[2])<<24 | uint32(byteSlice[3])<<16
			return fmt.Sprintf("%d", intValue*uint32(Weight))
		}
		if DataBlockOrder == "DCBA" {
			intValue := uint32(byteSlice[0]) | uint32(byteSlice[1])<<8 |
				uint32(byteSlice[2])<<16 | uint32(byteSlice[3])<<24
			return fmt.Sprintf("%d", intValue*uint32(Weight))
		}
	}

	return ""
}

func stringReverse(str string) string {
	bytes := []byte(str)
	for i := 0; i < len(str)/2; i++ {
		tmp := bytes[len(str)-i-1]
		bytes[len(str)-i-1] = bytes[i]
		bytes[i] = tmp
	}
	return string(bytes)
}

/*
*
* 默认字节序
*
 */
func GetDefaultDataOrder(Type, Order string) string {
	if Order == "" {
		switch Type {
		case "INT", "UINT", "INT32", "UINT32", "FLOAT", "FLOAT32":
			return "DCBA"
		case "BYTE", "I", "Q":
			return "A"
		case "INT16", "UINT16", "SHORT", "USHORT":
			return "BA"
		case "LONG", "ULONG":
			return "HGFEDCBA"
		}
	}
	return Order
}

/*
*
* 处理空指针初始值
*
 */
func HandleZeroValue[V int16 | int32 | int64 | float32 | float64](v *V) *V {
	if v == nil {
		return new(V)
	}
	return v
}

/*
*
*解析 Modbus 的值 有符号,
注意：如果想解析值，必须不能超过4字节，目前常见的数一般都是4字节，也许后期会有8字节，但是目前暂时不支持
*
*/
func ParseModbusValue(DataBlockType string, DataBlockOrder string,
	Weight float32, byteSlice [256]byte) string {
	// binary
	if DataBlockType == "UTF8" {
		acc := 0
		for _, v := range byteSlice {
			if v != 0 {
				acc++
			} else {
				continue
			}
		}
		if acc == 0 {
			return ""
		}
		if DataBlockOrder == "BIG_ENDIAN" {
			return string(byteSlice[:acc])
		}
		if DataBlockOrder == "LITTLE_ENDIAN" {
			return stringReverse(string(byteSlice[:acc]))
		}
	}

	if DataBlockType == "RAW" {
		acc := 0
		for _, v := range byteSlice {
			if v != 0 {
				acc++
			} else {
				continue
			}
		}
		if acc == 0 {
			return ""
		}
		return hex.EncodeToString(byteSlice[:acc])
	}

	if DataBlockType == "BYTE" {

		return fmt.Sprintf("%d", byteSlice[0])
	}
	// signed
	if DataBlockType == "SHORT" || DataBlockType == "INT16" {
		// AB: 1234
		// BA: 3412
		if DataBlockOrder == "AB" {
			int16Value := int16(byteSlice[0])<<8 | int16(byteSlice[1])
			if Weight == 1 {
				return fmt.Sprintf("%d", int16Value)
			}
			return fmt.Sprintf("%.4f", float32(int16Value)*(Weight))
		}
		if DataBlockOrder == "BA" {
			int16Value := int16(byteSlice[0]) | int16(byteSlice[1])<<8
			if Weight == 1 {
				return fmt.Sprintf("%d", int16Value)
			}
			return fmt.Sprintf("%.4f", float32(int16Value)*(Weight))
		}
	}
	if DataBlockType == "INT" || DataBlockType == "INT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := int32(byteSlice[0])<<24 | int32(byteSlice[1])<<16 |
				int32(byteSlice[2])<<8 | int32(byteSlice[3])
			if Weight == 1 {
				return fmt.Sprintf("%d", intValue)
			}
			return fmt.Sprintf("%.4f", float32(intValue)*(Weight))
		}
		if DataBlockOrder == "CDAB" {
			intValue := int32(byteSlice[0])<<8 | int32(byteSlice[1]) |
				int32(byteSlice[2])<<24 | int32(byteSlice[3])<<16
			if Weight == 1 {
				return fmt.Sprintf("%d", intValue)
			}
			return fmt.Sprintf("%.4f", float32(intValue)*(Weight))
		}
		if DataBlockOrder == "DCBA" {
			intValue := int32(byteSlice[0]) | int32(byteSlice[1])<<8 |
				int32(byteSlice[2])<<16 | int32(byteSlice[3])<<24
			if Weight == 1 {
				return fmt.Sprintf("%d", intValue)
			}
			return fmt.Sprintf("%.4f", float32(intValue)*(Weight))
		}
	}
	// Unsigned
	if DataBlockType == "USHORT" || DataBlockType == "UINT16" {
		// AB: 1234
		// BA: 3412
		if DataBlockOrder == "AB" {
			uint16Value := uint16(byteSlice[0])<<8 | uint16(byteSlice[1])
			if Weight == 1 {
				return fmt.Sprintf("%d", uint16Value)
			}
			return fmt.Sprintf("%.4f", float32(uint16Value)*(Weight))
		}
		if DataBlockOrder == "BA" {
			uint16Value := uint16(byteSlice[0]) | uint16(byteSlice[1])<<8
			if Weight == 1 {
				return fmt.Sprintf("%d", uint16Value)
			}
			return fmt.Sprintf("%.4f", float32(uint16Value)*(Weight))
		}
	}
	if DataBlockType == "UINT" || DataBlockType == "UINT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := uint32(byteSlice[0])<<24 | uint32(byteSlice[1])<<16 |
				uint32(byteSlice[2])<<8 | uint32(byteSlice[3])
			if Weight == 1 {
				return fmt.Sprintf("%d", intValue)
			}
			return fmt.Sprintf("%.4f", float32(intValue)*(Weight))
		}
		if DataBlockOrder == "CDAB" {
			intValue := uint32(byteSlice[0])<<8 | uint32(byteSlice[1]) |
				uint32(byteSlice[2])<<24 | uint32(byteSlice[3])<<16
			if Weight == 1 {
				return fmt.Sprintf("%d", intValue)
			}
			return fmt.Sprintf("%.4f", float32(intValue)*(Weight))
		}
		if DataBlockOrder == "DCBA" {
			intValue := uint32(byteSlice[0]) | uint32(byteSlice[1])<<8 |
				uint32(byteSlice[2])<<16 | uint32(byteSlice[3])<<24
			if Weight == 1 {
				return fmt.Sprintf("%d", intValue)
			}
			return fmt.Sprintf("%.4f", float32(intValue)*(Weight))
		}
	}
	// 3.14159:DCBA -> 40490FDC
	if DataBlockType == "FLOAT" || DataBlockType == "FLOAT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := int32(byteSlice[0])<<24 | int32(byteSlice[1])<<16 |
				int32(byteSlice[2])<<8 | int32(byteSlice[3])
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return fmt.Sprintf("%.4f", floatValue)
		}
		if DataBlockOrder == "CDAB" {
			intValue := int32(byteSlice[0])<<8 | int32(byteSlice[1]) |
				int32(byteSlice[2])<<24 | int32(byteSlice[3])<<16
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return fmt.Sprintf("%.4f", floatValue)
		}
		if DataBlockOrder == "DCBA" {
			intValue := int32(byteSlice[0]) | int32(byteSlice[1])<<8 |
				int32(byteSlice[2])<<16 | int32(byteSlice[3])<<24
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return fmt.Sprintf("%.4f", floatValue)

		}
	}
	return "0"
}
