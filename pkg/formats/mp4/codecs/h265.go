package codecs

// H265 is the H265 codec.
type H265 struct {
	VPS []byte
	SPS []byte
	PPS []byte
}

// IsVideo implements Codec.
func (*H265) IsVideo() bool {
	return true
}

func (*H265) isCodec() {}
