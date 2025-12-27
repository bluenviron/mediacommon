package codecs

// H265 is the H265 codec.
type H265 struct {
	SPS []byte
	PPS []byte
	VPS []byte
}

// IsVideo implements Codec.
func (*H265) IsVideo() bool {
	return true
}

func (*H265) isCodec() {}
