package mpeg4audio

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

// AudioMuxElement is an AudioMuxElement.
// Specification: ISO 14496-3, Table 1.41
type AudioMuxElement struct {
	// when unmarshaling, it must be filled
	MuxConfigPresent bool

	// when unmarshaling,
	// it must be filled when MuxConfigPresent=false
	// and it is filled when MuxConfigPresent=true and there's a config in the element
	StreamMuxConfig *StreamMuxConfig

	// used when MuxConfigPresent=true
	UseSameStreamMux bool

	// Payloads. Indexes are: subframe, program, layer.
	Payloads [][][][]byte
}

// Unmarshal decodes an AudioMuxElement.
func (e *AudioMuxElement) Unmarshal(buf []byte) error {
	pos := 0

	if e.MuxConfigPresent {
		var err error
		e.UseSameStreamMux, err = bits.ReadFlag(buf, &pos)
		if err != nil {
			return err
		}

		if !e.UseSameStreamMux {
			e.StreamMuxConfig = &StreamMuxConfig{}
			err = e.StreamMuxConfig.unmarshalBits(buf, &pos)
			if err != nil {
				return err
			}
		}
	}

	// we are assuming AudioMuxVersionA = 0

	if e.StreamMuxConfig == nil {
		return fmt.Errorf("no StreamMuxConfig available")
	}

	e.Payloads = make([][][][]byte, e.StreamMuxConfig.NumSubFrames+1)

	for i := range e.StreamMuxConfig.NumSubFrames + 1 {
		var li payloadLengthInfo
		li.StreamMuxConfig = e.StreamMuxConfig
		err := li.unmarshalBits(buf, &pos)
		if err != nil {
			return err
		}

		var payloadMux payloadMux
		err = payloadMux.unmarshalBits(buf, &pos, &li)
		if err != nil {
			return err
		}
		e.Payloads[i] = payloadMux.Payloads
	}

	if e.StreamMuxConfig.OtherDataPresent {
		return fmt.Errorf("OtherDataPresent is not supported (yet)")
	}

	n := pos / 8
	if pos%8 != 0 {
		n++
	}

	if n != len(buf) {
		return fmt.Errorf("detected unread bytes")
	}

	return nil
}

func (e AudioMuxElement) marshalSize() int {
	n := 0

	if e.MuxConfigPresent {
		n++

		if !e.UseSameStreamMux {
			n += e.StreamMuxConfig.marshalSizeBits()
		}
	}

	for i := range e.StreamMuxConfig.NumSubFrames + 1 {
		var li payloadLengthInfo
		li.StreamMuxConfig = e.StreamMuxConfig
		li.Programs = make([]payloadLengthInfoProgram, len(e.StreamMuxConfig.Programs))

		for prog, p := range e.StreamMuxConfig.Programs {
			li.Programs[prog].Layers = make([]payloadLengthInfoLayer, len(p.Layers))

			for lay := range p.Layers {
				err := li.setPayloadLength(prog, lay, uint64(len(e.Payloads[i][prog][lay])))
				if err != nil {
					return 0
				}
			}
		}

		n += li.marshalSizeBits()
		n += payloadMux{Payloads: e.Payloads[i]}.marshalSizeBits()
	}

	// we are assuming OtherDataPresent=false

	n2 := n / 8
	if n%8 != 0 {
		n2++
	}

	return n2
}

// Marshal encodes an AudioMuxElement.
func (e AudioMuxElement) Marshal() ([]byte, error) {
	buf := make([]byte, e.marshalSize())
	pos := 0

	if e.MuxConfigPresent {
		bits.WriteFlagUnsafe(buf, &pos, e.UseSameStreamMux)

		if !e.UseSameStreamMux {
			err := e.StreamMuxConfig.marshalToBits(buf, &pos)
			if err != nil {
				return nil, err
			}
		}
	}

	for i := range e.StreamMuxConfig.NumSubFrames + 1 {
		var li payloadLengthInfo
		li.StreamMuxConfig = e.StreamMuxConfig

		li.Programs = make([]payloadLengthInfoProgram, len(e.StreamMuxConfig.Programs))

		for prog, p := range e.StreamMuxConfig.Programs {
			li.Programs[prog].Layers = make([]payloadLengthInfoLayer, len(p.Layers))

			for lay := range p.Layers {
				err := li.setPayloadLength(prog, lay, uint64(len(e.Payloads[i][prog][lay])))
				if err != nil {
					return nil, err
				}
			}
		}

		err := li.marshalToBits(buf, &pos)
		if err != nil {
			return nil, err
		}

		payloadMux{Payloads: e.Payloads[i]}.marshalToBits(buf, &pos)
	}

	// we are assuming OtherDataPresent=false

	return buf, nil
}
