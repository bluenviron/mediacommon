package mpeg4audio

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

type payloadLengthInfoLayer struct {
	MuxSlotLengthBytes uint64
	MuxSlotLengthCoded uint8
}

type payloadLengthInfoProgram struct {
	Layers []payloadLengthInfoLayer
}

// payloadLengthInfo is a PayloadLengthInfo.
// Specification: ISO 14496-3, Table 1.44
type payloadLengthInfo struct {
	// configuration needed to perform unmarshaling and marshaling
	StreamMuxConfig *StreamMuxConfig

	Programs []payloadLengthInfoProgram
}

func (i *payloadLengthInfo) unmarshalBits(buf []byte, pos *int) error {
	// we are assuming allStreamsSameTimeFraming = 1

	i.Programs = make([]payloadLengthInfoProgram, len(i.StreamMuxConfig.Programs))

	for prog, p := range i.StreamMuxConfig.Programs {
		i.Programs[prog].Layers = make([]payloadLengthInfoLayer, len(p.Layers))

		for lay, l := range p.Layers {
			switch l.FrameLengthType {
			case 0:
				i.Programs[prog].Layers[lay].MuxSlotLengthBytes = 0

				for {
					tmp, err := bits.ReadBits(buf, pos, 8)
					if err != nil {
						return err
					}

					i.Programs[prog].Layers[lay].MuxSlotLengthBytes += tmp
					if tmp != 255 {
						break
					}
				}

			case 5, 7, 3:
				tmp, err := bits.ReadBits(buf, pos, 2)
				if err != nil {
					return err
				}
				i.Programs[prog].Layers[lay].MuxSlotLengthCoded = uint8(tmp)
			}
		}
	}

	return nil
}

func (i payloadLengthInfo) marshalSizeBits() int {
	n := 0

	for prog, p := range i.StreamMuxConfig.Programs {
		for lay := range p.Layers {
			frameLengthType := i.StreamMuxConfig.Programs[prog].Layers[lay].FrameLengthType

			switch frameLengthType {
			case 0:
				n += int(i.Programs[prog].Layers[lay].MuxSlotLengthBytes/255+1) * 8

			case 5, 7, 3:
				n += 2
			}
		}
	}

	return n
}

func (i *payloadLengthInfo) marshalToBits(buf []byte, pos *int) error {
	for prog, p := range i.StreamMuxConfig.Programs {
		for lay := range p.Layers {
			frameLengthType := i.StreamMuxConfig.Programs[prog].Layers[lay].FrameLengthType

			switch frameLengthType {
			case 0:
				v := i.Programs[prog].Layers[lay].MuxSlotLengthBytes

				for {
					if v >= 255 {
						bits.WriteBitsUnsafe(buf, pos, 255, 8)
						v -= 255
					} else {
						bits.WriteBitsUnsafe(buf, pos, v, 8)
						break
					}
				}

			case 5, 7, 3:
				bits.WriteBitsUnsafe(buf, pos, uint64(i.Programs[prog].Layers[lay].MuxSlotLengthCoded), 2)
			}
		}
	}

	return nil
}

func (i payloadLengthInfo) payloadLength(prog int, lay int) (uint64, error) {
	frameLengthType := i.StreamMuxConfig.Programs[prog].Layers[lay].FrameLengthType

	switch frameLengthType {
	case 0:
		return i.Programs[prog].Layers[lay].MuxSlotLengthBytes, nil

	default:
		return 0, fmt.Errorf("unsupported frameLengthType %d", frameLengthType)
	}
}

func (i *payloadLengthInfo) setPayloadLength(prog int, lay int, l uint64) error {
	frameLengthType := i.StreamMuxConfig.Programs[prog].Layers[lay].FrameLengthType

	switch frameLengthType {
	case 0:
		i.Programs[prog].Layers[lay].MuxSlotLengthBytes = l

	default:
		return fmt.Errorf("unsupported frameLengthType %d", frameLengthType)
	}

	return nil
}
