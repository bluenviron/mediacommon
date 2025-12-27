package codecs

// MJPEG is the M-JPEG codec.
type MJPEG struct {
	Width  int
	Height int
}

// IsVideo implements Codec.
func (*MJPEG) IsVideo() bool {
	return true
}

func (*MJPEG) isCodec() {}
