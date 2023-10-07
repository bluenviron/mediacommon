//nolint:gochecknoinits,revive,gocritic
package fmp4

import (
	gomp4 "github.com/abema/go-mp4"
)

func BoxTypeIpcm() gomp4.BoxType { return gomp4.StrToBoxType("ipcm") }

func init() {
	gomp4.AddAnyTypeBoxDef(&gomp4.AudioSampleEntry{}, BoxTypeIpcm())
}

func BoxTypePcmC() gomp4.BoxType { return gomp4.StrToBoxType("pcmC") }

func init() {
	gomp4.AddBoxDef(&PcmC{}, 0, 1)
}

type PcmC struct {
	gomp4.FullBox `mp4:"0,extend"`
	FormatFlags   uint8 `mp4:"1,size=8"`
	PCMSampleSize uint8 `mp4:"1,size=8"`
}

func (PcmC) GetType() gomp4.BoxType {
	return BoxTypePcmC()
}
