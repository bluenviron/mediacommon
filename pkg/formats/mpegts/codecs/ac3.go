package codecs

// AC3 is an AC-3 codec.
// Specification: ISO 13818-1
type AC3 struct {
	SampleRate   int
	ChannelCount int
}

// IsVideo implements Codec.
func (*AC3) IsVideo() bool {
	return false
}

func (*AC3) isCodec() {}
