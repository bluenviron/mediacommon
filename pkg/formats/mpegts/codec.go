package mpegts

import (
	"github.com/asticode/go-astits"
)

// Codec is a MPEG-TS codec.
type Codec interface {
	IsVideo() bool

	marshal(pid uint16) (*astits.PMTElementaryStream, error)
}
