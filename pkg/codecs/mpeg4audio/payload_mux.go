package mpeg4audio

import (
	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

// payloadMux is a PayloadMux.
// Specification: ISO 14496-3, Table 1.45
type payloadMux struct {
	// payloads.
	// first index is program, second index is layer.
	Payloads [][][]byte
}

func (m *payloadMux) unmarshalBits(buf []byte, pos *int, li *payloadLengthInfo) error {
	// we are assuming allStreamsSameTimeFraming = 1

	m.Payloads = make([][][]byte, len(li.StreamMuxConfig.Programs))

	for prog, p := range li.StreamMuxConfig.Programs {
		m.Payloads[prog] = make([][]byte, len(p.Layers))

		for lay := range p.Layers {
			payloadLength, err := li.payloadLength(prog, lay)
			if err != nil {
				return err
			}

			payload := make([]byte, payloadLength)

			for i := range payloadLength {
				var byt uint64
				byt, err = bits.ReadBits(buf, pos, 8)
				if err != nil {
					return err
				}

				payload[i] = byte(byt)
			}

			m.Payloads[prog][lay] = payload
		}
	}

	return nil
}

func (m payloadMux) marshalSizeBits() int {
	n := 0

	for _, p := range m.Payloads {
		for _, l := range p {
			n += len(l) * 8
		}
	}

	return n
}

func (m payloadMux) marshalToBits(buf []byte, pos *int) {
	for _, p := range m.Payloads {
		for _, l := range p {
			for _, byt := range l {
				bits.WriteBitsUnsafe(buf, pos, uint64(byt), 8)
			}
		}
	}
}
