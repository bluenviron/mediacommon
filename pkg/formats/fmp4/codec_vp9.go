package fmp4

// CodecVP9 is a VP9 codec.
type CodecVP9 struct {
	Width             int
	Height            int
	Profile           uint8
	BitDepth          uint8
	ChromaSubsampling uint8
	ColorRange        bool
}

func (CodecVP9) isVideo() bool {
	return true
}
