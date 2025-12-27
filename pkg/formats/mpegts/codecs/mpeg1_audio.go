package codecs

// MPEG1Audio is a MPEG-1 Audio codec.
// Specification: ISO 13818-1
type MPEG1Audio struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (*MPEG1Audio) IsVideo() bool {
	return true
}

func (*MPEG1Audio) isCodec() {}
