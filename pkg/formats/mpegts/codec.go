package mpegts

import (
	"github.com/asticode/go-astits"
)

// Codec is a MPEG-TS codec.
type Codec interface {
	// Marshal encodes the codec into a astits.PMTElementaryStream.
	Marshal(pid uint16) (*astits.PMTElementaryStream, error)

	// IsVideo returns whether the codec is a video codec.
	IsVideo() bool
}
