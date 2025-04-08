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
	"bytes"
	"time"
)

// GenericReadWriteCloser 是一个简单的 ReadWriteCloser 实现
type GenericReadWriteCloser struct {
	buffer *bytes.Buffer
}

func NewGenericReadWriteCloser() *GenericReadWriteCloser {
	return &GenericReadWriteCloser{
		buffer: new(bytes.Buffer),
	}
}
func (s *GenericReadWriteCloser) Read(p []byte) (n int, err error) {
	GenericByteParser := NewGenericByteParser(PacketEdger{
		Head: [2]byte{0xAB, 0xAB},
		Tail: [2]byte{0xBA, 0xBA},
	}, 1, 1024)
	ApplicationFrame := BuildApplicationFrame([]byte{0xAA, 0xBB, 0xCC, 0xDD})
	Response, _ := GenericByteParser.PackBytes(ApplicationFrame)
	copy(p, Response)
	// fmt.Println("[= TEST OUTPUT =] .Read:", p)
	return len(p), nil
}
func (s *GenericReadWriteCloser) Write(p []byte) (n int, err error) {
	// fmt.Println("[= TEST OUTPUT =] .Write:", p)
	return s.buffer.Write(p)
}
func (s *GenericReadWriteCloser) Close() error {
	s.buffer.Reset()
	return nil
}

func (s *GenericReadWriteCloser) SetReadDeadline(t time.Time) error {
	return nil

}
func (s *GenericReadWriteCloser) SetWriteDeadline(t time.Time) error {
	return nil

}
