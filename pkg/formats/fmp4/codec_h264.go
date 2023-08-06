package fmp4

// CodecH264 is a H264 codec.
type CodecH264 struct {
	SPS []byte
	PPS []byte
}

func (CodecH264) isVideo() bool {
	return true
}
