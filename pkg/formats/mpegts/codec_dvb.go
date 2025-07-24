package mpegts

import "github.com/asticode/go-astits"

const (
	subtIdentifier = 's'<<24 | 'u'<<16 | 'b'<<8 | 't'
)

// CodecDVB is a DVB codec.
type CodecDVB struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int // nolint:unused
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
				Length: 4,
				Tag:    astits.DescriptorTagSubtitling,
				Registration: &astits.DescriptorRegistration{
					FormatIdentifier: subtIdentifier,
				},
			},
		},
	}, nil
}
