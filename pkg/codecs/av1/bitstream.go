package av1

import (
	"fmt"
)

func obuRemoveSize(h *OBUHeader, sizeN int, ob []byte) []byte {
	newOBU := make([]byte, len(ob)-sizeN)
	newOBU[0] = (byte(h.Type) << 3)
	copy(newOBU[1:], ob[1+sizeN:])
	return newOBU
}

// BitstreamUnmarshal extracts OBUs from a bitstream.
// Optionally, it also removes the size field from OBUs.
// Specification: https://aomediacodec.github.io/av1-spec/#low-overhead-bitstream-format
func BitstreamUnmarshal(bs []byte, removeSizeField bool) ([][]byte, error) {
	var ret [][]byte

	for {
		var h OBUHeader
		err := h.Unmarshal(bs)
		if err != nil {
			return nil, err
		}

		if !h.HasSize {
			return nil, fmt.Errorf("OBU size not present")
		}

		size, sizeN, err := LEB128Unmarshal(bs[1:])
		if err != nil {
			return nil, err
		}

		obuLen := 1 + sizeN + int(size)
		if len(bs) < obuLen {
			return nil, fmt.Errorf("not enough bytes")
		}

		obu := bs[:obuLen]

		if removeSizeField {
			obu = obuRemoveSize(&h, sizeN, obu)
		}

		ret = append(ret, obu)
		bs = bs[obuLen:]

		if len(bs) == 0 {
			break
		}
	}

	return ret, nil
}
