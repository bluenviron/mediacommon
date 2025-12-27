package codecs

// Unsupported is an unsupported codec.
type Unsupported struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (*Unsupported) IsVideo() bool {
	return false
}

func (*Unsupported) isCodec() {}
