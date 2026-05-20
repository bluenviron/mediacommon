package codecs

import "github.com/bluenviron/mediacommon/v2/pkg/codecs/flac"

// FLAC is the FLAC codec.
type FLAC struct {
	StreamInfo *flac.StreamInfo
}

// IsVideo implements Codec.
func (*FLAC) IsVideo() bool {
	return false
}

func (*FLAC) isCodec() {}
