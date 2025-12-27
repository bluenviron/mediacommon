package codecs

// MPEG4Video is a MPEG-4 Video codec.
type MPEG4Video struct {
	Config []byte
}

// IsVideo implements Codec.
func (*MPEG4Video) IsVideo() bool {
	return true
}

func (*MPEG4Video) isCodec() {}
