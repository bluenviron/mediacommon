package av1

import (
	"fmt"
)

// LEB128Unmarshal decodes an unsigned integer from the LEB128 format.
//
// Deprecated: replaced by LEB128.Unmarshal
func LEB128Unmarshal(buf []byte) (uint, int, error) {
	var v LEB128
	n, err := v.Unmarshal(buf)
	return uint(v), n, err
}

// LEB128MarshalSize returns the marshal size of an unsigned integer in LEB128 format.
//
// Deprecated: replaced by LEB128.MarshalSize
func LEB128MarshalSize(v uint) int {
	return LEB128(v).MarshalSize()
}

// LEB128MarshalTo encodes an unsigned integer with the LEB128 format.
//
// Deprecated: replaced by LEB128.MarshalTo
func LEB128MarshalTo(v uint, buf []byte) int {
	return LEB128(v).MarshalTo(buf)
}

// LEB128 is a unsigned integer that can be decoded/encoded from/to the LEB128 format.
// Specification: https://aomediacodec.github.io/av1-spec/#leb128
type LEB128 uint32

// Unmarshal decodes an unsigned integer from the LEB128 format.
// It returns the number of consumed bytes.
func (l *LEB128) Unmarshal(buf []byte) (int, error) {
	*l = 0
	n := 0

	for i := 0; i < 8; i++ {
		if len(buf) == 0 {
			return 0, fmt.Errorf("not enough bytes")
		}

		var b byte
		b, buf = buf[0], buf[1:]

		*l |= (LEB128(b&0b01111111) << (i * 7))
		n++

		if (b & 0b10000000) == 0 {
			break
		}
	}

	return n, nil
}

// MarshalSize returns the marshal size in bytes of the unsigned integer in LEB128 format.
func (l LEB128) MarshalSize() int {
	n := 0

	for {
		l >>= 7
		n++

		if l <= 0 {
			break
		}
	}

	return n
}

// MarshalTo encodes the unsigned integer with the LEB128 format.
// It returns the number of consumed bytes.
func (l LEB128) MarshalTo(buf []byte) int {
	n := 0

	for {
		curbyte := byte(l) & 0b01111111
		l >>= 7

		if l <= 0 {
			buf[n] = curbyte
			n++
			break
		}

		curbyte |= 0b10000000
		buf[n] = curbyte
		n++
	}

	return n
}
