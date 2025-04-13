package protocol

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"testing"
)

func createValidFrame(data []byte) []byte {
	header := [2]byte{Header1, Header2}
	lenBytes := [2]byte{}
	binary.BigEndian.PutUint16(lenBytes[:], uint16(len(data)))
	crc := crc32.ChecksumIEEE(data)
	crcBytes := [2]byte{}
	binary.BigEndian.PutUint16(crcBytes[:], uint16(crc))
	tail := [2]byte{Tail1, Tail2}

	frame := append(header[:], lenBytes[:]...)
	frame = append(frame, data...)
	frame = append(frame, crcBytes[:]...)
	frame = append(frame, tail[:]...)

	return frame
}

// Function to create a frame with a too-long length field
func createFrameWithTooLongData() []byte {
	header := [2]byte{Header1, Header2}
	// Set the length field to a value larger than the actual data length
	lenBytes := [2]byte{}
	actualData := []byte{0x01, 0x02, 0x03}
	binary.BigEndian.PutUint16(lenBytes[:], uint16(len(actualData)+10)) // Make the length field too long
	crc := crc32.ChecksumIEEE(actualData)
	crcBytes := [2]byte{}
	binary.BigEndian.PutUint16(crcBytes[:], uint16(crc))
	tail := [2]byte{Tail1, Tail2}

	frame := append(header[:], lenBytes[:]...)
	frame = append(frame, actualData...)
	frame = append(frame, crcBytes[:]...)
	frame = append(frame, tail[:]...)

	return frame
}

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected BaseFrame
		err      error
	}{
		{
			name:  "Valid frame",
			input: createValidFrame([]byte{0x01, 0x02, 0x03}),
			expected: BaseFrame{
				Header: [2]byte{Header1, Header2},
				Len:    [2]byte{0x00, 0x03},
				Data:   []byte{0x01, 0x02, 0x03},
				CRC:    [2]byte{0, 0}, // Will be calculated in createValidFrame
				Tail:   [2]byte{Tail1, Tail2},
			},
			err: nil,
		},
		{
			name:  "Invalid header",
			input: []byte{0x00, 0x00, 0x00, 0x00},
			err:   ErrInvalidHeader,
		},
		{
			name:  "Invalid tail",
			input: createValidFrame([]byte{0x01, 0x02, 0x03})[:len(createValidFrame([]byte{0x01, 0x02, 0x03}))-2],
			err:   ErrInvalidTail,
		},
		{
			name: "Invalid CRC",
			input: func() []byte {
				frame := createValidFrame([]byte{0x01, 0x02, 0x03})
				frame[len(frame)-4] ^= 0xFF // Modify CRC byte
				return frame
			}(),
			err: ErrInvalidCRC,
		},

		{
			name:  "Partial data",
			input: createValidFrame([]byte{0x01, 0x02, 0x03})[:3],
			err:   fmt.Errorf("partial data: incomplete frame"),
		},

		{
			name:  "Context canceled",
			input: createValidFrame([]byte{0x01, 0x02, 0x03}),
			err:   ErrContextCanceled,
		},

		{
			name:  "Too long Data",
			input: createFrameWithTooLongData(),
			err:   ErrInvalidDataLen,
		},
	}
	parser := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if tt.err == ErrContextCanceled {
				cancel()
			}
			defer cancel()
			frame, err := parser.ParseBytes(ctx, tt.input)
			if err != nil && tt.err == nil {
				t.Errorf("ParseBytes() error = %v, wantErr %v", err, tt.err)
			}
			if err == nil && tt.err != nil {
				t.Errorf("ParseBytes() error = %v, wantErr %v", err, tt.err)
			}
			if err == nil {
				if frame.Header != tt.expected.Header {
					t.Errorf("ParseBytes() Header = %v, want %v", frame.Header, tt.expected.Header)
				}
				if frame.Len != tt.expected.Len {
					t.Errorf("ParseBytes() Len = %v, want %v", frame.Len, tt.expected.Len)
				}
				if string(frame.Data) != string(tt.expected.Data) {
					t.Errorf("ParseBytes() Data = %v, want %v", frame.Data, tt.expected.Data)
				}
				if frame.Tail != tt.expected.Tail {
					t.Errorf("ParseBytes() Tail = %v, want %v", frame.Tail, tt.expected.Tail)
				}
			}
		})
	}
}

// BenchmarkParseBytesValidFrame benchmarks the ParseBytes function with a valid frame
func BenchmarkParseBytesValidFrame(b *testing.B) {
	input := createValidFrame([]byte{0x01, 0x02, 0x03})
	ctx := context.Background()
	b.ResetTimer()
	parser := NewParser()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseBytes(ctx, input)
	}
}

// BenchmarkParseBytesInvalidHeader benchmarks the ParseBytes function with an invalid header
func BenchmarkParseBytesInvalidHeader(b *testing.B) {
	input := []byte{0x00, 0x00, 0x00, 0x00}
	ctx := context.Background()
	b.ResetTimer()
	parser := NewParser()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseBytes(ctx, input)
	}
}

func TestParseMultipleFrames(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    [][]byte
		expectedErr bool
	}{
		{
			name:        "Valid two frames",
			input:       append(createValidFrame([]byte{0x01, 0x02, 0x03}), createValidFrame([]byte{0x04, 0x05, 0x06})...),
			expected:    [][]byte{{0x01, 0x02, 0x03}, {0x04, 0x05, 0x06}},
			expectedErr: false,
		},
		{
			name:        "Empty input",
			input:       []byte{},
			expected:    [][]byte{},
			expectedErr: true,
		},
		{
			name:        "Invalid frame",
			input:       []byte{0x01, 0x02, 0x03}, // Assuming this is an invalid frame
			expected:    nil,
			expectedErr: true,
		},
		{
			name:        "Single valid frame",
			input:       createValidFrame([]byte{0x01, 0x02, 0x03}),
			expected:    [][]byte{{0x01, 0x02, 0x03}},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			ctx := context.Background()

			frames, err := parser.ParseMultipleFrames(ctx, tt.input)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(tt.expected) != len(frames) {
					t.Errorf("Expected %d frames, got %d", len(tt.expected), len(frames))
				}
				for i, frame := range frames {
					if string(tt.expected[i]) != string(frame.Data) {
						t.Errorf("Frame %d data mismatch: expected %v, got %v", i, tt.expected[i], frame.Data)
					}
				}
			}
		})
	}
}
