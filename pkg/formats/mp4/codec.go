package mp4

// Codec is a MP4 codec.
type Codec interface {
	IsVideo() bool

	isCodec()
}
