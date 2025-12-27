package codecs

// MPEG4Video is a MPEG-4 Video codec.
// Specification: ISO 13818-1
type MPEG4Video struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (*MPEG4Video) IsVideo() bool {
	return true
}

func (*MPEG4Video) isCodec() {}
