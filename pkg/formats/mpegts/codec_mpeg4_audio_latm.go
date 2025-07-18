package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecMPEG4AudioLATM is a MPEG-4 Audio LATM codec.
// Specification: ISO 13818-1
type CodecMPEG4AudioLATM struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (CodecMPEG4AudioLATM) IsVideo() bool {
	return false
}

func (*CodecMPEG4AudioLATM) isCodec() {}

func (c CodecMPEG4AudioLATM) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypeAACLATMAudio,
	}, nil
}
