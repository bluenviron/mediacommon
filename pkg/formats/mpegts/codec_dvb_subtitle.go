package mpegts

import "github.com/asticode/go-astits"

// CodecDVBSubtitle is a DVB Subtitle codec.
// Specification: ISO 13818-1
// Specification: ETSI EN 300 743
// Specification: ETSI EN 300 468
type CodecDVBSubtitle struct {
	// subtitling descriptor
	Descriptor *SubtitlingDescriptor
}

// IsVideo implements Codec.
func (CodecDVBSubtitle) IsVideo() bool {
	return false
}

func (*CodecDVBSubtitle) isCodec() {}

func (c CodecDVBSubtitle) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypePrivateData,
		ElementaryStreamDescriptors: []*astits.Descriptor{
			{
				// Length must be different than zero.
				// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
				Length: 1,
				Tag:    astits.DescriptorTagSubtitling,
				Subtitling: &astits.DescriptorSubtitling{
					Items: c.Descriptor.Items,
				},
			},
		},
	}, nil
}
