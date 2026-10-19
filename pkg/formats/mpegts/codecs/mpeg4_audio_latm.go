package codecs

import "github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"

// MPEG4AudioLATM is a MPEG-4 Audio LATM codec.
// Specification: ISO 13818-1
type MPEG4AudioLATM struct {
	// StreamMuxConfig of the first incoming audio packet with a configuration.
	mpeg4audio.StreamMuxConfig
}

// IsVideo implements Codec.
func (*MPEG4AudioLATM) IsVideo() bool {
	return false
}

func (*MPEG4AudioLATM) isCodec() {}
