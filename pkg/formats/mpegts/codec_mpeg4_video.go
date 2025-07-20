package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecMPEG4Video is a MPEG-4 Video codec.
// Specification: ISO 13818-1
type CodecMPEG4Video struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (CodecMPEG4Video) IsVideo() bool {
	return true
}

func (*CodecMPEG4Video) isCodec() {}

func (c CodecMPEG4Video) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypeMPEG4Video,
	}, nil
}
