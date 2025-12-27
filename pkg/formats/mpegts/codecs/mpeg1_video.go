package codecs

// MPEG1Video is a MPEG-1/2 Video codec.
// Specification: ISO 13818-1
type MPEG1Video struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (*MPEG1Video) IsVideo() bool {
	return true
}

func (*MPEG1Video) isCodec() {}
