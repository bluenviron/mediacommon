package codecs

// EAC3 is an Enhanced AC-3 (Dolby Digital Plus) codec.
// Specification: ETSI TS 102 366 V1.4.1, Annex E
// Specification: ETSI EN 300 468
type EAC3 struct {
	// Full service flag.
	FullService bool

	// Channels coding.
	// 0=1-2ch
	// 1=mono
	// 2=2ch stereo
	// 3=2ch surround
	// 4=multichannel mono
	// 5=multichannel stereo
	// 6=multichannel surround
	ChannelsCoding uint8

	// Deprecated: not filled anymore.
	SampleRate int

	// Deprecated: not filled anymore.
	ChannelCount int
}

// IsVideo implements Codec.
func (*EAC3) IsVideo() bool {
	return false
}

func (*EAC3) isCodec() {}
