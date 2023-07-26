package mpegts

import (
	"github.com/asticode/go-astits"
)

// Codec is a MPEG-TS codec.
type Codec interface {
	marshal(pid uint16) (*astits.PMTElementaryStream, error)
	isVideo() bool
}
