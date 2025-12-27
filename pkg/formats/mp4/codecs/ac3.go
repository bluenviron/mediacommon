package codecs

// AC3 is the AC-3 codec.
type AC3 struct {
	SampleRate   int
	ChannelCount int
	Fscod        uint8
	Bsid         uint8
	Bsmod        uint8
	Acmod        uint8
	LfeOn        bool
	BitRateCode  uint8
}

// IsVideo implements Codec.
func (*AC3) IsVideo() bool {
	return false
}

func (*AC3) isCodec() {}
