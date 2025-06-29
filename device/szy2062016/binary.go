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

package szy2062016

import (
	"fmt"
	"strconv"
)

func Hex2Byte(str string) []byte {
	slen := len(str)
	bHex := make([]byte, len(str)/2)
	ii := 0
	for i := 0; i < len(str); i = i + 2 {
		if slen != 1 {
			ss := string(str[i]) + string(str[i+1])
			bt, _ := strconv.ParseInt(ss, 16, 32)
			bHex[ii] = byte(bt)
			ii = ii + 1
			slen = slen - 2
		}
	}
	return bHex
}

func Byte2Hex(bs []byte) string {
	s := ""
	for _, b := range bs {
		s += fmt.Sprintf("%02x", b)
	}
	return s
}

func ByteSub(bs []byte, sub byte) []byte {
	r := []byte{}
	for _, b := range bs {
		r = append(r, b-sub)
	}
	return r
}

func ByteAdd(bs []byte, add byte) []byte {
	r := []byte{}
	for _, b := range bs {
		r = append(r, b+add)
	}
	return r
}

func ByteReverse(bs []byte) []byte {
	r := make([]byte, len(bs))
	for i, b := range bs {
		r[len(bs)-i-1] = b
	}
	return r
}
