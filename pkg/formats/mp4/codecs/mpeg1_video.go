package codecs

// MPEG1Video is a MPEG-1 Video codec.
type MPEG1Video struct {
	Config []byte
}

// IsVideo implements Codec.
func (*MPEG1Video) IsVideo() bool {
	return true
}

func (*MPEG1Video) isCodec() {}
