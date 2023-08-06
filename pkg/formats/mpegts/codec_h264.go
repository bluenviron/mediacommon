package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecH264 is a H264 codec.
type CodecH264 struct{}

func (c CodecH264) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID:               pid,
		ElementaryStreamDescriptors: nil,
		StreamType:                  astits.StreamTypeH264Video,
	}, nil
}

// IsVideo implements Codec.
func (CodecH264) IsVideo() bool {
	return true
}
