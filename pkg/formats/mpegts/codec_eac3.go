package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecEAC3 is an Enhanced AC-3 (Dolby Digital Plus) codec.
// Specification: ETSI TS 102 366 V1.4.1, Annex E
type CodecEAC3 struct {
	SampleRate   int
	ChannelCount int
}

// IsVideo implements Codec.
func (CodecEAC3) IsVideo() bool {
	return false
}

func (*CodecEAC3) isCodec() {}

// componentTypeFromConfig builds the DVB component_type byte.
// Per ETSI EN 300 468, table D.1:
// Bits 7-4: service_type_flag (0=complete main, 1=music/effects, etc.)
// Bits 3-1: number_of_channels mapping
// Bit 0: full_service_flag
//
// For E-AC-3, the component_type encodes channel configuration:
//
//	0x00-0x3F: Full service, complete main
//	Bits 2-0 encode channel config: 0=mono/stereo, 1=mono, 2=stereo, 3=2ch, etc.
func componentTypeFromConfig(channels int, fullService bool) uint8 {
	// Start with full service, complete main audio (bits 7-4 = 0000)
	var ct uint8

	// Set full_service_flag (bit 0)
	if fullService {
		ct |= 0x01
	}

	// Encode channel configuration in bits 3-1 (number_of_channels)
	// Per EN 300 468: 0=1-2ch, 1=mono, 2=2ch stereo, 3=2ch surround,
	//                 4=multichannel mono, 5=multichannel stereo, 6=multichannel surround
	switch {
	case channels <= 2:
		ct |= (0x02 << 1) // 2ch stereo
	case channels <= 4:
		ct |= (0x05 << 1) // multichannel stereo
	default:
		ct |= (0x06 << 1) // multichannel surround (5.1, 7.1, etc.)
	}

	return ct
}

func (c CodecEAC3) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	// Build Enhanced AC-3 descriptor per ETSI EN 300 468, section 6.2.16
	// This descriptor provides codec parameters that decoders need
	// before parsing actual E-AC3 frames.
	componentType := componentTypeFromConfig(c.ChannelCount, true)

	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypeEAC3Audio,
		ElementaryStreamDescriptors: []*astits.Descriptor{
			{
				// Length must be different than zero for astits writer
				// 1 byte flags + 1 byte component_type + 1 byte BSID = 3 bytes
				Length: 3,
				Tag:    astits.DescriptorTagEnhancedAC3,
				EnhancedAC3: &astits.DescriptorEnhancedAC3{
					HasComponentType: true,
					ComponentType:    componentType,
					// BSID=16 indicates E-AC-3
					HasBSID: true,
					BSID:    16,
				},
			},
		},
	}, nil
}
