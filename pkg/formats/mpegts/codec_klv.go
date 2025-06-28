package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecKLV is a KLV codec.
type CodecKLV struct {
	StreamType      astits.StreamType
	StreamID        uint8
	PTSDTSIndicator uint8
}

// IsVideo implements Codec.
func (c CodecKLV) IsVideo() bool {
	return false
}

// IsVideo implements Codec.
func (*CodecKLV) isCodec() {}

func (c CodecKLV) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		ElementaryStreamDescriptors: []*astits.Descriptor{
			{
				Length: 4,
				Tag:    astits.DescriptorTagRegistration,
				Registration: &astits.DescriptorRegistration{
					FormatIdentifier: klvaIdentifier,
				},
			},
		},
		StreamType: c.StreamType,
		// StreamType: astits.StreamTypeMetadata,
		// StreamType: astits.StreamTypePrivateData,
	}, nil
}
