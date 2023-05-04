package av1

import (
	"fmt"
)

// LEB128Unmarshal decodes an unsigned integer from the LEB128 format.
func LEB128Unmarshal(buf []byte) (uint, int, error) {
	v := uint(0)
	n := 0

	for i := 0; i < 8; i++ {
		if len(buf) == 0 {
			return 0, 0, fmt.Errorf("not enough bytes")
		}

		b := buf[0]

		v |= (uint(b&0b01111111) << (i * 7))
		n++

		if (b & 0b10000000) == 0 {
			break
		}

		buf = buf[1:]
	}

	return v, n, nil
}

// LEB128Marshal encodes an unsigned integer with the LEB128 format.
func LEB128Marshal(v uint) []byte {
	var out []byte

	for {
		curbyte := byte(v) & 0b01111111
		v >>= 7

		if v <= 0 {
			out = append(out, curbyte)
			break
		}

		curbyte |= 0b10000000
		out = append(out, curbyte)
	}

	return out
}
