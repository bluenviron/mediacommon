package mp4

import (
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
)

// CodecMPEG4Audio is a MPEG-4 Audio codec.
type CodecMPEG4Audio struct {
	Config mpeg4audio.AudioSpecificConfig
}

// IsVideo implements Codec.
func (CodecMPEG4Audio) IsVideo() bool {
	return false
}

func (*CodecMPEG4Audio) isCodec() {}
