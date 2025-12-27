package codecs

// MPEG1Audio is a MPEG-1 Audio codec.
type MPEG1Audio struct {
	SampleRate   int
	ChannelCount int
}

// IsVideo implements Codec.
func (*MPEG1Audio) IsVideo() bool {
	return false
}

func (*MPEG1Audio) isCodec() {}
