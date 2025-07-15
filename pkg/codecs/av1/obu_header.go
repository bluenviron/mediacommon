package av1

import (
	"fmt"
)

// OBUHeader is a OBU header.
// Specification: AV1 Bitstream & Decoding Process, section 5.3.2
type OBUHeader struct {
	Type    OBUType
	HasSize bool
}

// Unmarshal decodes a OBUHeader.
func (h *OBUHeader) Unmarshal(buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("not enough bytes")
	}

	forbidden := (buf[0] >> 7) != 0
	if forbidden {
		return fmt.Errorf("forbidden bit is set")
	}

	h.Type = OBUType(buf[0] >> 3)

	extensionFlag := ((buf[0] >> 2) & 0b1) != 0
	if extensionFlag {
		return fmt.Errorf("extension flag is not supported yet")
	}

	h.HasSize = ((buf[0] >> 1) & 0b1) != 0

	return nil
}
