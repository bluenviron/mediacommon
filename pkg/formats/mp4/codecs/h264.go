package codecs

// H264 is the H264 codec.
type H264 struct {
	SPS []byte
	PPS []byte
}

// IsVideo implements Codec.
func (*H264) IsVideo() bool {
	return true
}

func (*H264) isCodec() {}
