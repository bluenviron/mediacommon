package fmp4

// CodecAV1 is a AV1 codec.
type CodecAV1 struct {
	SequenceHeader []byte
}

func (CodecAV1) isVideo() bool {
	return true
}
