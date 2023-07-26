package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecOpus is a Opus codec.
type CodecOpus struct {
	ChannelCount int
}

// Marshal implements Codec.
func (c *CodecOpus) Marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		ElementaryStreamDescriptors: []*astits.Descriptor{
			{
				Length: 4,
				Tag:    astits.DescriptorTagRegistration,
				Registration: &astits.DescriptorRegistration{
					FormatIdentifier: opusIdentifier,
				},
			},
			{
				Length: 2,
				Tag:    astits.DescriptorTagExtension,
				Extension: &astits.DescriptorExtension{
					Tag:     0x80,
					Unknown: &[]uint8{uint8(c.ChannelCount)},
				},
			},
		},
		StreamType: astits.StreamTypePrivateData,
	}, nil
}
