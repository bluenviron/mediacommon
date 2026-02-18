package substructs

import (
	"fmt"
)

// OpusAccessUnit is the opus_access_unit structure.
// specification: ETSI TS Opus v0.1.3-draft, table 6-6
type OpusAccessUnit struct {
	ControlHeader OpusControlHeader
	Packet        []byte
}

// Unmarshal decodes an OpusAccessUnit.
func (au *OpusAccessUnit) Unmarshal(buf []byte) (int, error) {
	n, err := au.ControlHeader.unmarshal(buf)
	if err != nil {
		return 0, fmt.Errorf("invalid control header: %w", err)
	}
	buf = buf[n:]

	if len(buf) < au.ControlHeader.PayloadSize {
		return 0, fmt.Errorf("buffer is too small")
	}

	au.Packet = buf[:au.ControlHeader.PayloadSize]

	return n + au.ControlHeader.PayloadSize, nil
}

// MarshalSize returns the size of an OpusAccessUnit when marshaled.
func (au *OpusAccessUnit) MarshalSize() int {
	return au.ControlHeader.marshalSize() + au.ControlHeader.PayloadSize
}

// MarshalTo marshals an OpusAccessUnit to a buffer.
func (au *OpusAccessUnit) MarshalTo(buf []byte) (int, error) {
	if au.ControlHeader.PayloadSize != len(au.Packet) {
		return 0, fmt.Errorf("invalid packet")
	}

	n, err := au.ControlHeader.marshalTo(buf)
	if err != nil {
		return 0, err
	}

	n += copy(buf[n:], au.Packet)

	return n, nil
}
