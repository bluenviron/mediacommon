package codecs

// Opus is the Opus codec.
type Opus struct {
	ChannelCount int
}

// IsVideo implements Codec.
func (*Opus) IsVideo() bool {
	return false
}

func (*Opus) isCodec() {}
