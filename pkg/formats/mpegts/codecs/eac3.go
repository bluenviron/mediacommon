package codecs

// EAC3 is an Enhanced AC-3 (Dolby Digital Plus) codec.
// Specification: ETSI TS 102 366 V1.4.1, Annex E
type EAC3 struct {
	SampleRate   int
	ChannelCount int
}

// IsVideo implements Codec.
func (*EAC3) IsVideo() bool {
	return false
}

func (*EAC3) isCodec() {}
