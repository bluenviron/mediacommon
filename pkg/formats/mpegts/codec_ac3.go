package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecAC3 is an AC-3 codec.
// Specification: ISO 13818-1
type CodecAC3 struct {
	SampleRate   int
	ChannelCount int
}

// IsVideo implements Codec.
func (CodecAC3) IsVideo() bool {
	return false
}

func (*CodecAC3) isCodec() {}

// ac3ComponentType builds the DVB component_type byte for AC-3.
// Per ETSI EN 300 468, the AC3 descriptor uses a similar format to E-AC-3.
func ac3ComponentType(channels int, fullService bool) uint8 {
	var ct uint8 = 0

	// Set full_service_flag (bit 0)
	if fullService {
		ct |= 0x01
	}

	// Encode channel configuration in bits 3-1
	switch {
	case channels <= 2:
		ct |= (0x02 << 1) // 2ch stereo
	case channels <= 4:
		ct |= (0x05 << 1) // multichannel stereo
	default:
		ct |= (0x06 << 1) // multichannel surround (5.1, etc.)
	}

	return ct
}

func (c CodecAC3) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	// Build AC-3 descriptor per ETSI EN 300 468, section 6.2.1
	// This descriptor provides codec parameters that decoders need.
	componentType := ac3ComponentType(c.ChannelCount, true)

	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypeAC3Audio,
		ElementaryStreamDescriptors: []*astits.Descriptor{
			{
				// Length must be different than zero for astits writer
				// 1 byte flags + 1 byte component_type + 1 byte BSID = 3 bytes
				Length: 3,
				Tag:    astits.DescriptorTagAC3,
				AC3: &astits.DescriptorAC3{
					HasComponentType: true,
					ComponentType:    componentType,
					// BSID for standard AC-3 (not E-AC-3)
					HasBSID: true,
					BSID:    8,
				},
			},
		},
	}, nil
}
