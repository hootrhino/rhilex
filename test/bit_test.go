package test

import (
	"testing"

	"github.com/hootrhino/rhilex/rhilexlib"

	"github.com/go-playground/assert/v2"
)

func TestReverseBitOrder(t *testing.T) {
	assert.Equal(t, byte(0b1011_1111), rhilexlib.ReverseBits(byte(0b1111_1101)))
	assert.Equal(t, byte(0b1100_0000), rhilexlib.ReverseBits(byte(0b0000_0011)))
	assert.Equal(t, byte(0b0000_0101), rhilexlib.ReverseBits(byte(0b1010_0000)))
	assert.Equal(t, byte(0b1010_1010), rhilexlib.ReverseBits(byte(0b0101_0101)))
	assert.Equal(t, []byte{3, 2, 1}, rhilexlib.ReverseByteOrder([]byte{1, 2, 3}))
}
