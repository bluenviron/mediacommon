package codecs

// AC3 is an AC-3 codec.
// Specification: ISO 13818-1
// Specification: ETSI EN 300 468
type AC3 struct {
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

	// Deprecated: not filled and not used anymore.
	SampleRate int

	// Deprecated: not filled and not used anymore.
	ChannelCount int
}

// IsVideo implements Codec.
func (*AC3) IsVideo() bool {
	return false
}

func (*AC3) isCodec() {}
