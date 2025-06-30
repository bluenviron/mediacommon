package mpegts

import (
	"fmt"
)

type klvaAccessUnit struct {
	Packet []byte
	Header metadataAuCellHeader
}

func (au *klvaAccessUnit) unmarshal(buf []byte) (int, error) {
	n, err := au.Header.unmarshal(buf)
	if err != nil {
		return 0, fmt.Errorf("could not decode metadata AU header: %w", err)
	}

	if len(buf) < n+au.Header.PayloadSize {
		return 0, fmt.Errorf("buffer is too small")
	}

	au.Packet = buf[n : n+au.Header.PayloadSize]

	return n + au.Header.PayloadSize, nil
}

func (au *klvaAccessUnit) marshalSize() int {
	return au.Header.PayloadSize
}

func (au *klvaAccessUnit) marshalTo(buf []byte) (int, error) {
	if au.Header.PayloadSize != len(au.Packet) {
		return 0, fmt.Errorf("invalid packet")
	}

	n, err := au.Header.marshalTo(buf)
	if err != nil {
		return 0, err
	}

	n += copy(buf[n:], au.Packet)

	return n, nil
}
