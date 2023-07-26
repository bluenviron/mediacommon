package mpegts

import (
	"fmt"
)

type opusAccessUnit struct {
	ControlHeader opusControlHeader
	Packet        []byte
}

func (au *opusAccessUnit) unmarshal(buf []byte) (int, error) {
	n, err := au.ControlHeader.unmarshal(buf)
	if err != nil {
		return 0, fmt.Errorf("could not decode Opus control header: %v", err)
	}
	buf = buf[n:]

	if len(buf) < int(au.ControlHeader.PayloadSize) {
		return 0, fmt.Errorf("buffer is too small")
	}

	au.Packet = buf[:au.ControlHeader.PayloadSize]

	return n + int(au.ControlHeader.PayloadSize), nil
}
