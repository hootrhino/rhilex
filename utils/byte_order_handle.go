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
		case "BOOL":
			return "A"
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
func ParseRegisterValue(dataLen int, DataBlockType string, DataBlockOrder string,
	Weight float32, byteSlice [256]byte) any {
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
	// TODO: RAW 类型需要扩展
	if DataBlockType == "RAW" {
		return hex.EncodeToString(byteSlice[:dataLen])
	}
	if DataBlockType == "BYTE" {
		return byteSlice[0]
	}
	if DataBlockType == "BOOL" {
		return byteSlice[0] == 1
	}
	// signed
	if DataBlockType == "SHORT" || DataBlockType == "INT16" {
		// AB: 1234
		// BA: 3412
		if DataBlockOrder == "AB" {
			int16Value := int16(byteSlice[0])<<8 | int16(byteSlice[1])
			if Weight == 1 {
				return int16Value
			}
			return float32(int16Value) * (Weight)
		}
		if DataBlockOrder == "BA" {
			int16Value := int16(byteSlice[0]) | int16(byteSlice[1])<<8
			if Weight == 1 {
				return int16Value
			}
			return float32(int16Value) * (Weight)
		}
	}
	if DataBlockType == "INT" || DataBlockType == "INT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := int32(byteSlice[0])<<24 | int32(byteSlice[1])<<16 |
				int32(byteSlice[2])<<8 | int32(byteSlice[3])
			if Weight == 1 {
				return intValue
			}
			return float32(intValue) * (Weight)
		}
		if DataBlockOrder == "CDAB" {
			intValue := int32(byteSlice[0])<<8 | int32(byteSlice[1]) |
				int32(byteSlice[2])<<24 | int32(byteSlice[3])<<16
			if Weight == 1 {
				return intValue
			}
			return float32(intValue) * (Weight)
		}
		if DataBlockOrder == "DCBA" {
			intValue := int32(byteSlice[0]) | int32(byteSlice[1])<<8 |
				int32(byteSlice[2])<<16 | int32(byteSlice[3])<<24
			if Weight == 1 {
				return intValue
			}
			return float32(intValue) * (Weight)
		}
	}
	// Unsigned
	if DataBlockType == "USHORT" || DataBlockType == "UINT16" {
		// AB: 1234
		// BA: 3412
		if DataBlockOrder == "AB" {
			uint16Value := uint16(byteSlice[0])<<8 | uint16(byteSlice[1])
			if Weight == 1 {
				return uint16Value
			}
			return float32(uint16Value) * (Weight)
		}
		if DataBlockOrder == "BA" {
			uint16Value := uint16(byteSlice[0]) | uint16(byteSlice[1])<<8
			if Weight == 1 {
				return uint16Value
			}
			return float32(uint16Value) * (Weight)
		}
	}
	if DataBlockType == "UINT" || DataBlockType == "UINT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := uint32(byteSlice[0])<<24 | uint32(byteSlice[1])<<16 |
				uint32(byteSlice[2])<<8 | uint32(byteSlice[3])
			if Weight == 1 {
				return intValue
			}
			return float32(intValue) * (Weight)
		}
		if DataBlockOrder == "CDAB" {
			intValue := uint32(byteSlice[0])<<8 | uint32(byteSlice[1]) |
				uint32(byteSlice[2])<<24 | uint32(byteSlice[3])<<16
			if Weight == 1 {
				return intValue
			}
			return float32(intValue) * (Weight)
		}
		if DataBlockOrder == "DCBA" {
			intValue := uint32(byteSlice[0]) | uint32(byteSlice[1])<<8 |
				uint32(byteSlice[2])<<16 | uint32(byteSlice[3])<<24
			if Weight == 1 {
				return intValue
			}
			return float32(intValue) * (Weight)
		}
	}
	// 3.14159:DCBA -> 40490FDC
	if DataBlockType == "FLOAT" || DataBlockType == "FLOAT32" || DataBlockType == "UFLOAT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := int32(byteSlice[0])<<24 | int32(byteSlice[1])<<16 |
				int32(byteSlice[2])<<8 | int32(byteSlice[3])
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return floatValue
		}
		if DataBlockOrder == "CDAB" {
			intValue := int32(byteSlice[0])<<8 | int32(byteSlice[1]) |
				int32(byteSlice[2])<<24 | int32(byteSlice[3])<<16
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return floatValue
		}
		if DataBlockOrder == "DCBA" {
			intValue := int32(byteSlice[0]) | int32(byteSlice[1])<<8 |
				int32(byteSlice[2])<<16 | int32(byteSlice[3])<<24
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return floatValue

		}
	}
	// -3.14159:DCBA -> +40490FDC
	if DataBlockType == "UFLOAT32" {
		// ABCD
		if DataBlockOrder == "ABCD" {
			intValue := int32(byteSlice[0])<<24 | int32(byteSlice[1])<<16 |
				int32(byteSlice[2])<<8 | int32(byteSlice[3])
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return floatValue
		}
		if DataBlockOrder == "CDAB" {
			intValue := int32(byteSlice[0])<<8 | int32(byteSlice[1]) |
				int32(byteSlice[2])<<24 | int32(byteSlice[3])<<16
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return floatValue
		}
		if DataBlockOrder == "DCBA" {
			intValue := int32(byteSlice[0]) | int32(byteSlice[1])<<8 |
				int32(byteSlice[2])<<16 | int32(byteSlice[3])<<24
			floatValue := float32(math.Float32frombits(uint32(intValue)))
			return floatValue
		}
	}
	return 0
}

/**
 * 将Any转换成具体类型的字符串表示形式
 *
 */
func CovertAnyType(v any) string {
	switch T := v.(type) {
	case bool:
		return fmt.Sprintf("%v", T)
	case byte:
		return fmt.Sprintf("%d", v)
	case int8:
		return fmt.Sprintf("%d", v)
	case int16:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return fmt.Sprintf("%.4f", float32(T))
	case float64:
		return fmt.Sprintf("%.4f", float64(T))
	case *bool:
		return fmt.Sprintf("%v", T)
	case *byte:
		return fmt.Sprintf("%d", *T)
	case *int8:
		return fmt.Sprintf("%d", *T)
	case *int16:
		return fmt.Sprintf("%d", *T)
	case *int32:
		return fmt.Sprintf("%d", *T)
	case *int64:
		return fmt.Sprintf("%d", *T)
	case *uint16:
		return fmt.Sprintf("%d", *T)
	case *uint32:
		return fmt.Sprintf("%d", *T)
	case *uint64:
		return fmt.Sprintf("%d", *T)
	case *float32:
		return fmt.Sprintf("%.4f", float32(*T))
	case *float64:
		return fmt.Sprintf("%.4f", float64(*T))
	case string:
		return T
	}
	return "0.0"
}

/**
 * 解析Modbus值
 *
 */
func ParseModbusValue(dataLen int, DataBlockType string, DataBlockOrder string,
	Weight float32, byteSlice [256]byte) string {
	return CovertAnyType(ParseRegisterValue(dataLen, DataBlockType, DataBlockOrder, Weight, byteSlice))
}
