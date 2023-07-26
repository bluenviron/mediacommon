package mpegts

import (
	"github.com/asticode/go-astits"
)

// Codec is a MPEG-TS codec.
type Codec interface {
	Marshal(pid uint16) (*astits.PMTElementaryStream, error)
}
