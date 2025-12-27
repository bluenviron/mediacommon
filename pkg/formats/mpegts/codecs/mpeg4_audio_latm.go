package codecs

// MPEG4AudioLATM is a MPEG-4 Audio LATM codec.
// Specification: ISO 13818-1
type MPEG4AudioLATM struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (*MPEG4AudioLATM) IsVideo() bool {
	return false
}

func (*MPEG4AudioLATM) isCodec() {}
