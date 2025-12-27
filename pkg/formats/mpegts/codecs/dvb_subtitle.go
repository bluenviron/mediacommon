package codecs

import "github.com/asticode/go-astits"

// DVBSubtitle is a DVB Subtitle codec.
// Specification: ISO 13818-1
// Specification: ETSI EN 300 743
// Specification: ETSI EN 300 468
type DVBSubtitle struct {
	Items []*astits.DescriptorSubtitlingItem
}

// IsVideo implements Codec.
func (*DVBSubtitle) IsVideo() bool {
	return false
}

func (*DVBSubtitle) isCodec() {}
