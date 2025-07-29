package mpegts

import "github.com/asticode/go-astits"

type SubtitlingDescriptor struct {
	Tag    uint8
	Length uint8
	Items  []*astits.DescriptorSubtitlingItem
}
