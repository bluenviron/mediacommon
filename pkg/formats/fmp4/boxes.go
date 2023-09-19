//nolint:gochecknoinits,revive,gocritic
package fmp4

import (
	gomp4 "github.com/abema/go-mp4"
)

func BoxTypeMp4v() gomp4.BoxType { return gomp4.StrToBoxType("mp4v") }

func init() {
	gomp4.AddAnyTypeBoxDef(&gomp4.VisualSampleEntry{}, BoxTypeMp4v())
}

func BoxTypeAC3() gomp4.BoxType { return gomp4.StrToBoxType("ac-3") }
func BoxTypeDA3() gomp4.BoxType { return gomp4.StrToBoxType("dac3") }

type Dac3 struct {
	gomp4.Box
	Fscod       uint8 `mp4:"0,size=2"`
	Bsid        uint8 `mp4:"1,size=5"`
	Bsmod       uint8 `mp4:"2,size=3"`
	Acmod       uint8 `mp4:"3,size=3"`
	LfeOn       uint8 `mp4:"4,size=1"`
	BitRateCode uint8 `mp4:"5,size=5"`
	Reserved    uint8 `mp4:"6,size=5,const=0"`
}

func (Dac3) GetType() gomp4.BoxType {
	return BoxTypeDA3()
}

func init() {
	gomp4.AddAnyTypeBoxDef(&gomp4.AudioSampleEntry{}, BoxTypeAC3())
	gomp4.AddBoxDef(&Dac3{})
}
