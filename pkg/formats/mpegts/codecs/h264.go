package codecs

// H264 is a H264 codec.
// Specification: ISO 13818-1
type H264 struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (*H264) IsVideo() bool {
	return true
}

func (*H264) isCodec() {}
