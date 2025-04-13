package protocol

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

const (
	Header1 byte = 0xA1
	Header2 byte = 0xB1
	Tail1   byte = 0xA2
	Tail2   byte = 0xB2
)

// Define frame structure
type BaseFrame struct {
	Header [2]byte // Frame header
	Len    [2]byte // Data length
	Data   []byte  // Frame data
	CRC    [2]byte // CRC16 checksum
	Tail   [2]byte // Frame tail
}

// State machine states
const (
	StateHeader1 = iota
	StateHeader2
	StateLen1
	StateLen2
	StateData
	StateCRC1
	StateCRC2
	StateTail1
	StateTail2
)

var (
	ErrInvalidHeader   = errors.New("invalid frame header")
	ErrInvalidTail     = errors.New("invalid frame tail")
	ErrInvalidDataLen  = errors.New("invalid data length")
	ErrInvalidCRC      = errors.New("invalid CRC checksum")
	ErrContextCanceled = errors.New("context canceled")
	ErrInvalidInput    = errors.New("invalid input: empty byte slice")
)

// Parser holds the state and un-parsed data
type Parser struct {
	state     int
	dataLen   int
	frame     BaseFrame
	remaining []byte
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{
		state: StateHeader1,
	}
}

// ParseBytes parses the input bytes using a state machine and saves un-parsed data
func (p *Parser) ParseBytes(ctx context.Context, b []byte) (BaseFrame, error) {
	input := append(p.remaining, b...)
	p.remaining = nil
	index := 0

	for index < len(input) {
		select {
		case <-ctx.Done():
			p.remaining = input[index:]
			return BaseFrame{}, ErrContextCanceled
		default:
			switch p.state {
			case StateHeader1:
				if input[index] == Header1 {
					p.frame.Header[0] = input[index]
					p.state = StateHeader2
				}
				index++
			case StateHeader2:
				if index < len(input) && input[index] == Header2 {
					p.frame.Header[1] = input[index]
					p.state = StateLen1
				} else {
					// Reset state if invalid header
					p.state = StateHeader1
				}
				index++
			case StateLen1:
				if index < len(input) {
					p.frame.Len[0] = input[index]
					p.state = StateLen2
				}
				index++
			case StateLen2:
				if index < len(input) {
					p.frame.Len[1] = input[index]
					p.dataLen = int(binary.BigEndian.Uint16(p.frame.Len[:]))
					p.state = StateData
				}
				index++
			case StateData:
				end := index + p.dataLen
				if end <= len(input) {
					p.frame.Data = input[index:end]
					index = end
					p.state = StateCRC1
				} else {
					// Save un-parsed data
					p.remaining = input[index:]
					return BaseFrame{}, fmt.Errorf("partial data: need %d more bytes", end-len(input))
				}
			case StateCRC1:
				if index < len(input) {
					p.frame.CRC[0] = input[index]
					p.state = StateCRC2
				}
				index++
			case StateCRC2:
				if index < len(input) {
					p.frame.CRC[1] = input[index]
					// Calculate CRC
					crc := crc32.ChecksumIEEE(p.frame.Data)
					expectedCRC := binary.BigEndian.Uint16(p.frame.CRC[:])
					if uint16(crc) != expectedCRC {
						p.state = StateHeader1
						return BaseFrame{}, ErrInvalidCRC
					}
					p.state = StateTail1
				}
				index++
			case StateTail1:
				if index < len(input) && input[index] == Tail1 {
					p.frame.Tail[0] = input[index]
					p.state = StateTail2
				} else {
					p.state = StateHeader1
					return BaseFrame{}, ErrInvalidTail
				}
				index++
			case StateTail2:
				if index < len(input) && input[index] == Tail2 {
					p.frame.Tail[1] = input[index]
					p.state = StateHeader1
					return p.frame, nil
				} else {
					p.state = StateHeader1
					return BaseFrame{}, ErrInvalidTail
				}
			}
		}
	}
	p.remaining = input[index:]
	return BaseFrame{}, fmt.Errorf("partial data: incomplete frame")
}

// ParseMultipleFrames parses multiple frames from the input bytes
func (p *Parser) ParseMultipleFrames(ctx context.Context, b []byte) ([]BaseFrame, error) {
	if len(b) == 0 {
		return nil, ErrInvalidInput
	}

	var frames []BaseFrame
	remaining := b
	for len(remaining) > 0 {
		frame, err := p.ParseBytes(ctx, remaining)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
		frameLen := 2 + 2 + len(frame.Data) + 2 + 2
		remaining = remaining[frameLen:]
	}

	return frames, nil
}
