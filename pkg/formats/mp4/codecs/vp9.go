package codecs

// VP9 is the VP9 codec.
type VP9 struct {
	Width             int
	Height            int
	Profile           uint8
	BitDepth          uint8
	ChromaSubsampling uint8
	ColorRange        bool
}

// IsVideo implements Codec.
func (*VP9) IsVideo() bool {
	return true
}

func (*VP9) isCodec() {}
