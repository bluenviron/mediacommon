package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecUnsupported is an unsupported codec.
type CodecUnsupported struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (CodecUnsupported) IsVideo() bool {
	return false
}

func (*CodecUnsupported) isCodec() {}

func (c CodecUnsupported) marshal(uint16) (*astits.PMTElementaryStream, error) {
	panic("this should not happen")
}
