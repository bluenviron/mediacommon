package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecH265 is a H265 codec.
type CodecH265 struct{}

// Marshal implements Codec.
func (c *CodecH265) Marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID:               pid,
		ElementaryStreamDescriptors: nil,
		StreamType:                  astits.StreamTypeH265Video,
	}, nil
}
