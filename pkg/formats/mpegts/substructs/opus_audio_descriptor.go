package substructs

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

func ceilLog2(n int) int {
	if n <= 1 {
		return 0
	}
	v := uint(n - 1)
	result := 0
	for v > 0 {
		v >>= 1
		result++
	}
	return result
}

// channelConfigTable maps channel_config_code to channel count.
// Values come directly from Table 4-3 of the specification.
var channelConfigTable = map[uint8]int{
	0x00: 2, // dual mono
	0x01: 1, // mono
	0x02: 2, // stereo
	0x03: 3, // 3ch surround
	0x04: 4, // quad
	0x05: 5, // 5ch
	0x06: 6, // 5.1
	0x07: 7, // 6.1
	0x08: 8, // 7.1
	// 0x09…0x7F: reserved
	0x80: 2, // dual mono (alias)
	// 0x81: explicit configuration (parsed below)
	0x82: 1,
	0x83: 2,
	0x84: 3,
	0x85: 4,
	0x86: 5,
	0x87: 6,
	0x88: 7,
	// 0x89…0xFF: reserved
}

// OpusAudioDescriptor is the opus_audio_descriptor structure.
// specification: ETSI TS Opus v0.1.3-draft, table 4-2.
type OpusAudioDescriptor struct {
	ChannelConfigCode uint8

	// ChannelConfigCode = 0x81
	ExplicitChannelCount int
	MappingFamily        int

	// ChannelConfigCode = 0x81, MappingFamily > 0
	StreamCount        int
	CoupledStreamCount int
	ChannelMapping     []byte
}

// Unmarshal decodes an AudioDescriptor.
func (d *OpusAudioDescriptor) Unmarshal(buf []byte) error {
	if len(buf) < 1 {
		return fmt.Errorf("buffer is too short")
	}

	d.ChannelConfigCode = buf[0]

	switch d.ChannelConfigCode {
	case 0x81:
		if len(buf) < 3 {
			return fmt.Errorf("buffer is too short")
		}

		d.ExplicitChannelCount = int(buf[1])
		if d.ExplicitChannelCount == 0 {
			return fmt.Errorf("channel_count must not be zero")
		}

		d.MappingFamily = int(buf[2])

		buf = buf[3:]

		if d.MappingFamily == 0 {
			switch d.ExplicitChannelCount {
			case 1, 2:
			default:
				return fmt.Errorf("unsupported channel_count for mapping_family 0: %d", d.ExplicitChannelCount)
			}
			return nil
		}

		pos := 0
		if d.ExplicitChannelCount == 1 {
			d.StreamCount = 1
		} else {
			tmp, err := bits.ReadBits(buf, &pos, ceilLog2(d.ExplicitChannelCount))
			if err != nil {
				return err
			}
			d.StreamCount = int(tmp) + 1
		}

		tmp, err := bits.ReadBits(buf, &pos, ceilLog2(d.StreamCount+1))
		if err != nil {
			return err
		}
		d.CoupledStreamCount = int(tmp)

		d.ChannelMapping = make([]byte, d.ExplicitChannelCount)
		for i := range d.ChannelMapping {
			var v uint64
			v, err = bits.ReadBits(buf, &pos, ceilLog2(d.StreamCount+d.CoupledStreamCount+1))
			if err != nil {
				return err
			}
			d.ChannelMapping[i] = byte(v)
		}
		return nil

	default:
		if _, ok := channelConfigTable[d.ChannelConfigCode]; !ok {
			return fmt.Errorf("unsupported channel_config_code: 0x%02X", d.ChannelConfigCode)
		}
	}

	return nil
}

func (d OpusAudioDescriptor) marshalSize() int {
	switch d.ChannelConfigCode {
	case 0x81:
		if d.MappingFamily == 0 {
			return 3
		}

		if d.ExplicitChannelCount == 1 {
			return 4
		}

		bitsForStreamCountMinus1 := ceilLog2(d.ExplicitChannelCount)
		bitsForCoupled := ceilLog2(d.StreamCount + 1)
		bitsPerMapping := ceilLog2(d.StreamCount + d.CoupledStreamCount + 1)
		totalBits := bitsForStreamCountMinus1 + bitsForCoupled + d.ExplicitChannelCount*bitsPerMapping
		packedBytes := (totalBits + 7) / 8
		return 3 + packedBytes

	default:
		return 1
	}
}

// Marshal encodes an AudioDescriptor.
func (d OpusAudioDescriptor) Marshal() ([]byte, error) {
	out := make([]byte, d.marshalSize())

	switch d.ChannelConfigCode {
	case 0x81:
		out[0] = 0x81
		out[1] = byte(d.ExplicitChannelCount)
		out[2] = byte(d.MappingFamily)

		if d.MappingFamily == 0 {
			return out, nil
		}

		if d.ExplicitChannelCount == 1 {
			out[3] = 0
			return out, nil
		}

		bitsForStreamCountMinus1 := ceilLog2(d.ExplicitChannelCount)
		bitsForCoupled := ceilLog2(d.StreamCount + 1)
		bitsPerMapping := ceilLog2(d.StreamCount + d.CoupledStreamCount + 1)
		packedOffset := 3
		pos := 0

		if bitsForStreamCountMinus1 > 0 {
			bits.WriteBitsUnsafe(out[packedOffset:], &pos, uint64(d.StreamCount-1), bitsForStreamCountMinus1)
		}

		if bitsForCoupled > 0 {
			bits.WriteBitsUnsafe(out[packedOffset:], &pos, uint64(d.CoupledStreamCount), bitsForCoupled)
		}

		for i := 0; i < d.ExplicitChannelCount; i++ {
			if bitsPerMapping > 0 {
				bits.WriteBitsUnsafe(out[packedOffset:], &pos, uint64(d.ChannelMapping[i]), bitsPerMapping)
			}
		}

		return out, nil

	default:
		out[0] = d.ChannelConfigCode
		return out, nil
	}
}

// ChannelCount returns channel count.
func (d *OpusAudioDescriptor) ChannelCount() int {
	if d.ChannelConfigCode == 0x81 {
		return d.ExplicitChannelCount
	}

	if count, ok := channelConfigTable[d.ChannelConfigCode]; ok {
		return count
	}

	return 0
}
