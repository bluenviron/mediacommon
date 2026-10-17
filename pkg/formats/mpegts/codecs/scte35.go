package codecs

// SCTE35 is a SCTE-35 codec.
// Specification: ANSI/SCTE 35
type SCTE35 struct {
	// in Go, empty structs share the same pointer,
	// therefore they cannot be used as map keys
	// or in equality operations. Prevent this.
	unused int //nolint:unused
}

// IsVideo implements Codec.
func (*SCTE35) IsVideo() bool {
	return false
}

func (*SCTE35) isCodec() {}
