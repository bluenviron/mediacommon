package fmp4

// CodecOpus is a Opus codec.
type CodecOpus struct {
	ChannelCount int
}

func (CodecOpus) isCodec() {}
