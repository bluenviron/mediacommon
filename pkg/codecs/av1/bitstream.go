package av1

import (
	"fmt"
)

// Bitstream is an AV1 bitstream.
// Specification: https://aomediacodec.github.io/av1-spec/#low-overhead-bitstream-format
type Bitstream [][]byte

// Unmarshal decodes a Bitstream.
func (bs *Bitstream) Unmarshal(buf []byte) error {
	for {
		var h OBUHeader
		err := h.Unmarshal(buf)
		if err != nil {
			return err
		}

		if !h.HasSize {
			return fmt.Errorf("OBU size not present")
		}

		var size LEB128
		n, err := size.Unmarshal(buf[1:])
		if err != nil {
			return err
		}

		obuAndSizeLen := 1 + int(size) + n
		if len(buf) < obuAndSizeLen {
			return fmt.Errorf("not enough bytes")
		}

		obu := make([]byte, 1+int(size))
		obu[0] = buf[0] & 0b11111101
		copy(obu[1:], buf[1+n:])

		buf = buf[obuAndSizeLen:]

		*bs = append(*bs, obu)

		if len(buf) == 0 {
			break
		}
	}

	return nil
}

// Marshal encodes a Bitstream.
func (bs Bitstream) Marshal() ([]byte, error) {
	n := 0

	for _, obu := range bs {
		n += len(obu)

		var h OBUHeader
		err := h.Unmarshal(obu)
		if err != nil {
			return nil, err
		}

		if !h.HasSize {
			size := len(obu) - 1
			n += LEB128(uint32(size)).MarshalSize()
		}
	}

	buf := make([]byte, n)
	n = 0

	for _, obu := range bs {
		var h OBUHeader
		h.Unmarshal(obu) //nolint:errcheck

		if !h.HasSize {
			buf[n] = obu[0] | 0b00000010
			n++
			size := len(obu) - 1
			n += LEB128(uint32(size)).MarshalTo(buf[n:])
			n += copy(buf[n:], obu[1:])
		} else {
			n += copy(buf[n:], obu)
		}
	}

	return buf, nil
}
