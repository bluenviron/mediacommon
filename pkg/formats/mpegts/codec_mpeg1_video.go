package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecMPEG1Video is a MPEG-1/2 Video codec.
type CodecMPEG1Video struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (CodecMPEG1Video) IsVideo() bool {
	return true
}

func (*CodecMPEG1Video) isCodec() {}

func (c CodecMPEG1Video) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		// we use MPEG-2 to notify readers that video can be either MPEG-1 or MPEG-2
		StreamType: astits.StreamTypeMPEG2Video,
	}, nil
}
