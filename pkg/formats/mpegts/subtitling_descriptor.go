package mpegts

import "github.com/asticode/go-astits"

// SubtitlingDescriptor is a subtitling descriptor.
type SubtitlingDescriptor struct {
	Items []*astits.DescriptorSubtitlingItem
}
