package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericByteParser(t *testing.T) {
	// Define the header and tail for the protocol
	edger := PacketEdger{
		Head: [2]byte{0xAF, 0x00},
		Tail: [2]byte{0xFA, 0x00},
	}

	// Create an instance of GenericByteParser
	parser := NewGenericByteParser(edger, 1, 1024)

	// Test case: Valid packet
	validPayload := []byte{0x01, 0x02, 0x03}
	validChecksum := calculateCRC16(validPayload)
	validPacket := append([]byte{
		0xAF, 0x00, // Header
		0x00, 0x03, // Length (3 bytes payload)
	}, validPayload...)
	validPacket = append(validPacket, byte(validChecksum>>8), byte(validChecksum&0xFF)) // CRC16 checksum
	validPacket = append(validPacket, 0xFA, 0x00)                                       // Tail

	t.Run("Valid Packet", func(t *testing.T) {
		data, err := parser.ParseBytes(validPacket)
		assert.NoError(t, err)
		assert.Equal(t, validPayload, data)
	})

	// Test case: Invalid checksum
	invalidChecksumPacket := append([]byte{
		0xAF, 0x00, // Header
		0x00, 0x03, // Length (3 bytes payload)
	}, validPayload...)
	invalidChecksumPacket = append(invalidChecksumPacket, 0x00, 0x00) // Invalid checksum
	invalidChecksumPacket = append(invalidChecksumPacket, 0xFA, 0x00) // Tail

	t.Run("Invalid Checksum", func(t *testing.T) {
		data, err := parser.ParseBytes(invalidChecksumPacket)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Equal(t, "checksum validation failed", err.Error())
	})

	// Test case: Invalid tail
	invalidTailPacket := append(validPacket[:len(validPacket)-2], 0x00, 0x00) // Invalid tail

	t.Run("Invalid Tail", func(t *testing.T) {
		data, err := parser.ParseBytes(invalidTailPacket)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Equal(t, "invalid tail", err.Error())
	})

	// Test case: Length mismatch
	lengthMismatchPacket := append([]byte{
		0xAF, 0x00, // Header
		0x00, 0x05, // Length (5 bytes, but only 3 bytes provided)
	}, validPayload...)
	lengthMismatchPacket = append(lengthMismatchPacket, byte(validChecksum>>8), byte(validChecksum&0xFF)) // CRC16 checksum
	lengthMismatchPacket = append(lengthMismatchPacket, 0xFA, 0x00)                                       // Tail

	t.Run("Length Mismatch", func(t *testing.T) {
		data, err := parser.ParseBytes(lengthMismatchPacket)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Equal(t, "data length out of range", err.Error())
	})

	// Test case: Empty data
	t.Run("Empty Data", func(t *testing.T) {
		data, err := parser.ParseBytes([]byte{})
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Equal(t, "no valid header found", err.Error())
	})

	// Test case: Header and tail mismatch
	mismatchedPacket := []byte{
		0xAF, 0x00, // Header
		0x00, 0x03, // Length
		0x01, 0x02, 0x03, // Payload
		0x00, 0x00, // Invalid checksum
		0xFB, 0x00, // Invalid tail
	}

	t.Run("Header Tail Mismatch", func(t *testing.T) {
		data, err := parser.ParseBytes(mismatchedPacket)
		assert.Error(t, err)
		assert.Nil(t, data)
		assert.Equal(t, "invalid tail", err.Error())
	})
}
