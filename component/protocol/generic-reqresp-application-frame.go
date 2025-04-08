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

package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// ApplicationFrame 应用层帧结构体
type ApplicationFrame struct {
	Length  uint16
	Payload []byte
	Crc16   uint16
}

func NewApplicationFrame(payload []byte) *ApplicationFrame {
	return &ApplicationFrame{
		Length:  uint16(len(payload)),
		Payload: payload,
		Crc16:   CRC16(payload),
	}
}
func BuildApplicationFrame(payload []byte) ApplicationFrame {
	return ApplicationFrame{
		Length:  uint16(len(payload)),
		Payload: payload,
		Crc16:   CRC16(payload),
	}
}
func CRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// ToString 转换为字符串
func (f *ApplicationFrame) ToString() string {
	return fmt.Sprintf("ApplicationFrame | Length: %d, Payload: %x, Crc16: %x",
		f.Length, f.Payload, f.Crc16)
}

// DecodeApplicationFrame 解码数据字节数组为 ApplicationFrame 结构
func DecodeApplicationFrame(data []byte) (*ApplicationFrame, error) {
	if len(data) < 4 {
		return nil, errors.New("data too short to decode")
	}
	frame := ApplicationFrame{}
	frame.Length = binary.BigEndian.Uint16(data[:2])

	if len(data) < int(frame.Length)+4 {
		return nil, errors.New("data length mismatch")
	}
	frame.Payload = data[2 : 2+frame.Length]
	frame.Crc16 = binary.BigEndian.Uint16(data[2+frame.Length:])
	if CRC16(frame.Payload) != frame.Crc16 {
		return nil, errors.New("crc16 check failed")
	}
	return &frame, nil
}

// Encode 将 ApplicationFrame 编码为字节数组
func (f *ApplicationFrame) Encode() ([]byte, error) {
	if len(f.Payload) != int(f.Length) {
		return nil, errors.New("payload length does not match the specified Length")
	}
	data := make([]byte, 2+len(f.Payload)+2)
	binary.BigEndian.PutUint16(data[:2], f.Length)
	copy(data[2:2+len(f.Payload)], f.Payload)
	binary.BigEndian.PutUint16(data[len(data)-2:], f.Crc16)
	return data, nil
}

// 消息结构体
type Message struct {
	Type uint16
	Data []byte
}

// Encode 将消息编码为字节数组
func (m *Message) Encode() ([]byte, error) {
	buf := make([]byte, 2+len(m.Data))
	binary.BigEndian.PutUint16(buf[:2], m.Type)
	copy(buf[2:], m.Data)
	return buf, nil
}

// Decode 将字节数组解码为消息
func DecodeMessage(data []byte) (Message, error) {
	if len(data) < 2 {
		return Message{}, errors.New("data too short to decode")
	}
	m := Message{
		Type: binary.BigEndian.Uint16(data[:2]),
		Data: data[2:],
	}
	return m, nil
}
