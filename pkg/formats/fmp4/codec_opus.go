package fmp4

// CodecOpus is a Opus codec.
type CodecOpus struct {
	ChannelCount int
}

// IsVideo implements Codec.
func (CodecOpus) IsVideo() bool {
	return false
}

func (CodecOpus) isCodec() {}
