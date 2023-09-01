package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecMPEG4Video is a MPEG-4 Video codec.
type CodecMPEG4Video struct{}

func (c CodecMPEG4Video) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID:               pid,
		ElementaryStreamDescriptors: nil,
		StreamType:                  astits.StreamTypeMPEG4Video,
	}, nil
}

// IsVideo implements Codec.
func (CodecMPEG4Video) IsVideo() bool {
	return true
}
