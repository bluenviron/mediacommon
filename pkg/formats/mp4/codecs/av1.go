package codecs

// AV1 is the AV1 codec.
type AV1 struct {
	SequenceHeader []byte
}

// IsVideo implements Codec.
func (*AV1) IsVideo() bool {
	return true
}

func (*AV1) isCodec() {}
