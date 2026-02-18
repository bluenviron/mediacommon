package codecs

import "github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/substructs"

// Opus is a Opus codec.
// Specification: ETSI TS Opus 0.1.3-draft
type Opus struct {
	Desc *substructs.OpusAudioDescriptor

	// Deprecated: use Desc instead.
	ChannelCount int
}

// IsVideo implements Codec.
func (*Opus) IsVideo() bool {
	return false
}

func (*Opus) isCodec() {}
