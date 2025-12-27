package codecs

// Opus is a Opus codec.
// Specification: ETSI TS Opus 0.1.3-draft
type Opus struct {
	ChannelCount int
}

// IsVideo implements Codec.
func (*Opus) IsVideo() bool {
	return false
}

func (*Opus) isCodec() {}
