package mpegts

import (
	"github.com/asticode/go-astits"
)

const (
	opusIdentifier = 'O'<<24 | 'p'<<16 | 'u'<<8 | 's'
)

// CodecOpus is a Opus codec.
// Specification: ETSI TS Opus 0.1.3-draft
type CodecOpus struct {
	ChannelCount int
}

// IsVideo implements Codec.
func (CodecOpus) IsVideo() bool {
	return false
}

func (*CodecOpus) isCodec() {}

func (c CodecOpus) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypePrivateData,
		ElementaryStreamDescriptors: []*astits.Descriptor{
			{
				// Length must be different than zero.
				// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
				Length: 1,
				Tag:    astits.DescriptorTagRegistration,
				Registration: &astits.DescriptorRegistration{
					FormatIdentifier: opusIdentifier,
				},
			},
			{
				// Length must be different than zero.
				// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
				Length: 1,
				Tag:    astits.DescriptorTagExtension,
				Extension: &astits.DescriptorExtension{
					Tag:     0x80,
					Unknown: &[]uint8{uint8(c.ChannelCount)},
				},
			},
		},
	}, nil
}
