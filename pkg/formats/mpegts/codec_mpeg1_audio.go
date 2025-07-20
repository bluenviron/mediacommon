package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecMPEG1Audio is a MPEG-1 Audio codec.
// Specification: ISO 13818-1
type CodecMPEG1Audio struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (CodecMPEG1Audio) IsVideo() bool {
	return true
}

func (*CodecMPEG1Audio) isCodec() {}

func (c CodecMPEG1Audio) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypeMPEG1Audio,
	}, nil
}
