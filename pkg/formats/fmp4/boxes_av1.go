//nolint:gochecknoinits,revive,gocritic
package fmp4

import (
	"github.com/abema/go-mp4"
)

func BoxTypeAv01() mp4.BoxType { return mp4.StrToBoxType("av01") }

func init() {
	mp4.AddAnyTypeBoxDef(&mp4.VisualSampleEntry{}, BoxTypeAv01())
}

func BoxTypeAv1C() mp4.BoxType { return mp4.StrToBoxType("av1C") }

func init() {
	mp4.AddBoxDef(&Av1C{})
}

type Av1C struct {
	mp4.Box
	Marker                           uint8   `mp4:"0,size=1,const=1"`
	Version                          uint8   `mp4:"1,size=7,const=1"`
	SeqProfile                       uint8   `mp4:"2,size=3"`
	SeqLevelIdx0                     uint8   `mp4:"3,size=5"`
	SeqTier0                         uint8   `mp4:"4,size=1"`
	HighBitdepth                     uint8   `mp4:"5,size=1"`
	TwelveBit                        uint8   `mp4:"6,size=1"`
	Monochrome                       uint8   `mp4:"7,size=1"`
	ChromaSubsamplingX               uint8   `mp4:"8,size=1"`
	ChromaSubsamplingY               uint8   `mp4:"9,size=1"`
	ChromaSamplePosition             uint8   `mp4:"10,size=2"`
	Reserved                         uint8   `mp4:"11,size=3,const=0"`
	InitialPresentationDelayPresent  uint8   `mp4:"12,size=1"`
	InitialPresentationDelayMinusOne uint8   `mp4:"13,size=4"`
	ConfigOBUs                       []uint8 `mp4:"14,size=8"`
}

func (Av1C) GetType() mp4.BoxType {
	return BoxTypeAv1C()
}
