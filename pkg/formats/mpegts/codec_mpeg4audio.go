package mpegts

import (
	"github.com/asticode/go-astits"

	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

// CodecMPEG4Audio is a MPEG-4 Audio codec.
type CodecMPEG4Audio struct {
	mpeg4audio.Config
}

// Marshal implements Codec.
func (c *CodecMPEG4Audio) Marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID:               pid,
		ElementaryStreamDescriptors: nil,
		StreamType:                  astits.StreamTypeAACAudio,
	}, nil
}
