package mpeg4audio

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

// StreamMuxConfigLayer is a layer of a StreamMuxConfig.
type StreamMuxConfigLayer struct {
	AudioSpecificConfig       *AudioSpecificConfig
	FrameLengthType           uint
	LatmBufferFullness        uint
	FrameLength               uint
	CELPframeLengthTableIndex uint
	HVXCframeLengthTableIndex bool
}

// StreamMuxConfigProgram is a program of a StreamMuxConfig.
type StreamMuxConfigProgram struct {
	Layers []*StreamMuxConfigLayer
}

// StreamMuxConfig is a StreamMuxConfig.
// Specification: ISO 14496-3, Table 1.42
type StreamMuxConfig struct {
	NumSubFrames     uint
	Programs         []*StreamMuxConfigProgram
	OtherDataPresent bool
	OtherDataLenBits uint32
	CRCCheckPresent  bool
	CRCCheckSum      uint8
}

// Unmarshal decodes a StreamMuxConfig.
func (c *StreamMuxConfig) Unmarshal(buf []byte) error {
	pos := 0
	err := c.unmarshalBits(buf, &pos)
	if err != nil {
		return err
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

// unmarshalBits decodes a StreamMuxConfig.
func (c *StreamMuxConfig) unmarshalBits(buf []byte, pos *int) error {
	err := bits.HasSpace(buf, *pos, 12)
	if err != nil {
		return err
	}

	audioMuxVersion := bits.ReadFlagUnsafe(buf, pos)
	if audioMuxVersion {
		return fmt.Errorf("audioMuxVersion = 1 is not supported")
	}

	allStreamsSameTimeFraming := bits.ReadFlagUnsafe(buf, pos)
	if !allStreamsSameTimeFraming {
		return fmt.Errorf("allStreamsSameTimeFraming = 0 is not supported")
	}

	c.NumSubFrames = uint(bits.ReadBitsUnsafe(buf, pos, 6))
	numProgram := uint(bits.ReadBitsUnsafe(buf, pos, 4))

	c.Programs = make([]*StreamMuxConfigProgram, numProgram+1)

	for prog := uint(0); prog <= numProgram; prog++ {
		p := &StreamMuxConfigProgram{}
		c.Programs[prog] = p

		var numLayer uint64
		numLayer, err = bits.ReadBits(buf, pos, 3)
		if err != nil {
			return err
		}

		p.Layers = make([]*StreamMuxConfigLayer, numLayer+1)

		for lay := uint(0); lay <= uint(numLayer); lay++ {
			l := &StreamMuxConfigLayer{}
			p.Layers[lay] = l

			var useSameConfig bool

			if prog == 0 && lay == 0 {
				useSameConfig = false
			} else {
				useSameConfig, err = bits.ReadFlag(buf, pos)
				if err != nil {
					return err
				}
			}

			if !useSameConfig {
				l.AudioSpecificConfig = &AudioSpecificConfig{}
				err = l.AudioSpecificConfig.unmarshalBits(buf, pos)
				if err != nil {
					return err
				}
			}

			var tmp uint64
			tmp, err = bits.ReadBits(buf, pos, 3)
			if err != nil {
				// support truncated configs
				l.LatmBufferFullness = 255
				c.Programs = c.Programs[:prog+1]
				p.Layers = p.Layers[:lay+1]
				return nil //nolint:nilerr
			}
			l.FrameLengthType = uint(tmp)

			switch l.FrameLengthType {
			case 0:
				tmp, err = bits.ReadBits(buf, pos, 8)
				if err != nil {
					return err
				}
				l.LatmBufferFullness = uint(tmp)

			case 1:
				tmp, err = bits.ReadBits(buf, pos, 9)
				if err != nil {
					return err
				}
				l.FrameLength = uint(tmp)

			case 4, 5, 3:
				tmp, err = bits.ReadBits(buf, pos, 6)
				if err != nil {
					return err
				}
				l.CELPframeLengthTableIndex = uint(tmp)

			case 6, 7:
				l.HVXCframeLengthTableIndex, err = bits.ReadFlag(buf, pos)
				if err != nil {
					return err
				}
			}
		}
	}

	c.OtherDataPresent, err = bits.ReadFlag(buf, pos)
	if err != nil {
		return err
	}

	if c.OtherDataPresent {
		for {
			c.OtherDataLenBits *= 256

			err = bits.HasSpace(buf, *pos, 9)
			if err != nil {
				return err
			}

			otherDataLenEsc := bits.ReadFlagUnsafe(buf, pos)
			otherDataLenTmp := uint32(bits.ReadBitsUnsafe(buf, pos, 8))
			c.OtherDataLenBits += otherDataLenTmp

			if !otherDataLenEsc {
				break
			}
		}
	}

	c.CRCCheckPresent, err = bits.ReadFlag(buf, pos)
	if err != nil {
		return err
	}

	if c.CRCCheckPresent {
		var tmp uint64
		tmp, err = bits.ReadBits(buf, pos, 8)
		if err != nil {
			return err
		}
		c.CRCCheckSum = uint8(tmp)
	}

	return nil
}

func (c StreamMuxConfig) marshalSizeBits() int {
	n := 12

	for prog, p := range c.Programs {
		n += 3

		for lay, l := range p.Layers {
			if prog != 0 || lay != 0 {
				n++
			}

			if l.AudioSpecificConfig != nil {
				n += l.AudioSpecificConfig.marshalSizeBits()
			}

			n += 3

			switch l.FrameLengthType {
			case 0:
				n += 8

			case 1:
				n += 9

			case 4, 5, 3:
				n += 6

			case 6, 7:
				n++
			}
		}
	}

	n++ // otherDataPresent

	if c.OtherDataPresent {
		tmp := c.OtherDataLenBits
		for {
			tmp /= 256
			n += 9

			if tmp == 0 {
				break
			}
		}
	}

	n++ // crcCheckPresent

	if c.CRCCheckPresent {
		n += 8
	}

	return n
}

func (c StreamMuxConfig) marshalSize() int {
	n := c.marshalSizeBits()

	ret := n / 8
	if (n % 8) != 0 {
		ret++
	}

	return ret
}

// Marshal encodes a StreamMuxConfig.
func (c StreamMuxConfig) Marshal() ([]byte, error) {
	buf := make([]byte, c.marshalSize())
	pos := 0
	err := c.marshalToBits(buf, &pos)
	return buf, err
}

func (c StreamMuxConfig) marshalToBits(buf []byte, pos *int) error {
	bits.WriteFlagUnsafe(buf, pos, false) // audioMuxVersion
	bits.WriteFlagUnsafe(buf, pos, true)  // allStreamsSameTimeFraming
	bits.WriteBitsUnsafe(buf, pos, uint64(c.NumSubFrames), 6)
	bits.WriteBitsUnsafe(buf, pos, uint64(len(c.Programs)-1), 4)

	for prog, p := range c.Programs {
		bits.WriteBitsUnsafe(buf, pos, uint64(len(p.Layers)-1), 3)

		for lay, l := range p.Layers {
			if prog != 0 || lay != 0 {
				bits.WriteFlagUnsafe(buf, pos, l.AudioSpecificConfig == nil)
			}

			if l.AudioSpecificConfig != nil {
				err := l.AudioSpecificConfig.marshalToBits(buf, pos)
				if err != nil {
					return err
				}
			}

			bits.WriteBitsUnsafe(buf, pos, uint64(l.FrameLengthType), 3)

			switch l.FrameLengthType {
			case 0:
				bits.WriteBitsUnsafe(buf, pos, uint64(l.LatmBufferFullness), 8)

			case 1:
				bits.WriteBitsUnsafe(buf, pos, uint64(l.FrameLength), 9)

			case 4, 5, 3:
				bits.WriteBitsUnsafe(buf, pos, uint64(l.CELPframeLengthTableIndex), 6)

			case 6, 7:
				bits.WriteFlagUnsafe(buf, pos, l.HVXCframeLengthTableIndex)
			}
		}
	}

	if c.OtherDataPresent {
		bits.WriteFlagUnsafe(buf, pos, true)

		var lenBytes []byte
		tmp := c.OtherDataLenBits

		for {
			mod := tmp % 256
			tmp -= mod
			tmp /= 256
			lenBytes = append(lenBytes, uint8(mod))

			if tmp == 0 {
				break
			}
		}

		for i := len(lenBytes) - 1; i > 0; i-- {
			bits.WriteFlagUnsafe(buf, pos, true)
			bits.WriteBitsUnsafe(buf, pos, uint64(lenBytes[i]), 8)
		}

		bits.WriteFlagUnsafe(buf, pos, false)
		bits.WriteBitsUnsafe(buf, pos, uint64(lenBytes[0]), 8)
	} else {
		bits.WriteFlagUnsafe(buf, pos, false)
	}

	bits.WriteFlagUnsafe(buf, pos, c.CRCCheckPresent)

	if c.CRCCheckPresent {
		bits.WriteBitsUnsafe(buf, pos, uint64(c.CRCCheckSum), 8)
	}

	return nil
}
