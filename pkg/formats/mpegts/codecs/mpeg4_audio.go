package codecs

import (
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
)

// MPEG4Audio is a MPEG-4 Audio codec.
// Specification: ISO 13818-1
type MPEG4Audio struct {
	mpeg4audio.Config
}

// IsVideo implements Codec.
func (*MPEG4Audio) IsVideo() bool {
	return false
}

func (*MPEG4Audio) isCodec() {}
