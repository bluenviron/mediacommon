package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecH265 is a H265 codec.
type CodecH265 struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (CodecH265) IsVideo() bool {
	return true
}

func (*CodecH265) isCodec() {}

func (c CodecH265) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypeH265Video,
	}, nil
}
