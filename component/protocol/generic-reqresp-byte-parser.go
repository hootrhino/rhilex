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
	"errors"
	"fmt"
	"io"
)

// GenericByteParser implements a state-machine-based parser
type GenericByteParser struct {
	edger         PacketEdger
	buffer        bytes.Buffer
	state         parserState
	payloadLength int
	minPayloadLen int
	maxPayloadLen int
}

// parserState defines the states of the state machine
type parserState int

const (
	stateIdle parserState = iota
	stateHeader
	stateLength
	statePayload
	stateChecksum
	stateTail
)

// NewGenericByteParser creates a new parser instance
func NewGenericByteParser(edger PacketEdger, minPayloadLen, maxPayloadLen int) *GenericByteParser {
	return &GenericByteParser{
		edger:         edger,
		state:         stateIdle,
		minPayloadLen: minPayloadLen,
		maxPayloadLen: maxPayloadLen,
	}
}

// PackBytes creates a packet with the given frame
func (parser *GenericByteParser) PackBytes(frame ApplicationFrame) ([]byte, error) {
	data, err := frame.Encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode frame: %w", err)
	}

	bodyLength := len(data)
	if bodyLength < parser.minPayloadLen || bodyLength > parser.maxPayloadLen {
		return nil, errors.New("payload length out of range")
	}

	packet := bytes.Buffer{}
	packet.Write(parser.edger.Head[:]) // Write header

	// Write length
	packet.WriteByte(byte(bodyLength >> 8))
	packet.WriteByte(byte(bodyLength & 0xFF))

	// Write payload
	packet.Write(data)

	// Write checksum
	checksum := calculateCRC16(data)
	packet.WriteByte(byte(checksum >> 8))
	packet.WriteByte(byte(checksum & 0xFF))

	// Write tail
	packet.Write(parser.edger.Tail[:])

	return packet.Bytes(), nil
}

// ParseBytes parses the input bytes and returns the payload if valid
func (parser *GenericByteParser) ParseBytes(b []byte) ([]byte, error) {
	parser.buffer.Write(b)

	for {
		switch parser.state {
		case stateIdle:
			if parser.buffer.Len() < len(parser.edger.Head) {
				return nil, nil // Wait for more data
			}
			header := parser.buffer.Next(len(parser.edger.Head))
			if !bytes.Equal(header, parser.edger.Head[:]) {
				return nil, errors.New("invalid header")
			}
			parser.state = stateLength

		case stateLength:
			if parser.buffer.Len() < 2 {
				return nil, nil // Wait for more data
			}
			lengthBytes := parser.buffer.Next(2)
			length := int(lengthBytes[0])<<8 | int(lengthBytes[1])
			if length < parser.minPayloadLen || length > parser.maxPayloadLen {
				parser.state = stateIdle
				return nil, errors.New("data length out of range")
			}
			parser.payloadLength = length
			parser.state = statePayload

		case statePayload:
			if parser.buffer.Len() < parser.payloadLength {
				return nil, nil // Wait for more data
			}
			payload := parser.buffer.Next(parser.payloadLength)
			parser.state = stateChecksum
			return payload, nil

		case stateChecksum:
			if parser.buffer.Len() < 2 {
				return nil, nil // Wait for more data
			}
			checksumBytes := parser.buffer.Next(2)
			calculatedChecksum := calculateCRC16(parser.buffer.Bytes()[:parser.payloadLength])
			receivedChecksum := int(checksumBytes[0])<<8 | int(checksumBytes[1])
			if calculatedChecksum != receivedChecksum {
				parser.state = stateIdle
				return nil, errors.New("checksum validation failed")
			}
			parser.state = stateTail

		case stateTail:
			if parser.buffer.Len() < len(parser.edger.Tail) {
				return nil, nil // Wait for more data
			}
			tail := parser.buffer.Next(len(parser.edger.Tail))
			if !bytes.Equal(tail, parser.edger.Tail[:]) {
				parser.state = stateIdle
				return nil, errors.New("invalid tail")
			}
			parser.state = stateIdle
			return parser.buffer.Bytes()[:parser.payloadLength], nil
		}
	}
}

// calculateCRC16 calculates the CRC16 checksum for the given data
func calculateCRC16(data []byte) int {
	var crc uint16 = 0xFFFF
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
	return int(crc)
}

// ReadAndParse reads data from an io.Reader and parses it
func (parser *GenericByteParser) ReadAndParse(r io.Reader) ([]byte, error) {
	buffer := make([]byte, 1024)
	n, err := r.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	return parser.ParseBytes(buffer[:n])
}

// WritePacket writes a packet to an io.Writer
func (parser *GenericByteParser) WritePacket(frame ApplicationFrame, w io.Writer) error {
	packet, err := parser.PackBytes(frame)
	if err != nil {
		return err
	}
	_, err = w.Write(packet)
	return err
}
