// Package codecs contains MP4 codecs.
package codecs

// Codec is a MP4 codec.
type Codec interface {
	IsVideo() bool

	isCodec()
}
