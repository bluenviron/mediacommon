package mpegts

import (
	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

// CodecMPEG4Audio is a MPEG-4 Audio codec.
type CodecMPEG4Audio struct {
	mpeg4audio.Config
}

func (*CodecMPEG4Audio) isCodec() {
}
