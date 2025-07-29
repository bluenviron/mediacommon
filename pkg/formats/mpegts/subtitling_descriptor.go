package mpegts

import "github.com/asticode/go-astits"

type subtitlingDescriptor struct {
	tag    uint8
	length uint8
	items  []*astits.DescriptorSubtitlingItem
}
