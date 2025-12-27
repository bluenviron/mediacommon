package codecs

// LPCM is the LPCM codec.
type LPCM struct {
	LittleEndian bool
	BitDepth     int
	SampleRate   int
	ChannelCount int
}

// IsVideo implements Codec.
func (*LPCM) IsVideo() bool {
	return false
}

func (*LPCM) isCodec() {}
