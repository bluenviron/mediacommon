package fmp4

// CodecH265 is a H265 codec.
type CodecH265 struct {
	SPS []byte
	PPS []byte
	VPS []byte
}

func (CodecH265) isCodec() {}
