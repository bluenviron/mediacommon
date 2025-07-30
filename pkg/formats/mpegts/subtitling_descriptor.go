package mpegts

import "github.com/asticode/go-astits"

// SubtitlingDescriptor is a subtitling descriptor.
// Specification: ETSI EN 300 468
type SubtitlingDescriptor struct {
	Items []*astits.DescriptorSubtitlingItem
}
