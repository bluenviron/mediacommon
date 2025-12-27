package codecs

// KLV is a KLV codec.
// Specification: MISB ST 1402
type KLV struct {
	Synchronous bool
}

// IsVideo implements Codec.
func (*KLV) IsVideo() bool {
	return false
}

func (*KLV) isCodec() {}
