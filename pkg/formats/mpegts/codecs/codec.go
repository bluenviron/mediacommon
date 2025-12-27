// Package codecs contains MPEG-TS codecs.
package codecs

// Codec is a MPEG-TS codec.
type Codec interface {
	IsVideo() bool

	isCodec()
}
