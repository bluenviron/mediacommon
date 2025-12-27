package codecs

// H265 is a H265 codec.
// Specification: ISO 13818-1
type H265 struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (*H265) IsVideo() bool {
	return true
}

func (*H265) isCodec() {}
