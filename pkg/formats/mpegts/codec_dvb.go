package mpegts

import "github.com/asticode/go-astits"

// CodecDVB is a DVB codec.
type CodecDVB struct {
	// subtitling descriptor
	descriptor *subtitlingDescriptor
}

// IsVideo implements Codec.
func (CodecDVB) IsVideo() bool {
	return false
}

func (*CodecDVB) isCodec() {}

func (c CodecDVB) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypePrivateData,
		ElementaryStreamDescriptors: []*astits.Descriptor{
			{
				Tag:    c.descriptor.tag,
				Length: c.descriptor.length,
				Subtitling: &astits.DescriptorSubtitling{
					Items: c.descriptor.items,
				},
			},
		},
	}, nil
}
