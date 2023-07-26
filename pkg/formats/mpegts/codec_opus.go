package mpegts

// CodecOpus is a Opus codec.
type CodecOpus struct {
	Channels int
}

func (*CodecOpus) isCodec() {
}
