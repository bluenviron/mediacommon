//nolint:gochecknoinits,revive,gocritic
package fmp4

import (
	"github.com/abema/go-mp4"
)

func BoxTypeVp08() mp4.BoxType { return mp4.StrToBoxType("vp08") }

func BoxTypeVp09() mp4.BoxType { return mp4.StrToBoxType("vp09") }

func init() {
	mp4.AddAnyTypeBoxDef(&mp4.VisualSampleEntry{}, BoxTypeVp08())
	mp4.AddAnyTypeBoxDef(&mp4.VisualSampleEntry{}, BoxTypeVp09())
}

func BoxTypeVpcC() mp4.BoxType { return mp4.StrToBoxType("vpcC") }

func init() {
	mp4.AddBoxDef(&VpcC{})
}

type VpcC struct {
	mp4.FullBox                 `mp4:"0,extend"`
	Profile                     uint8   `mp4:"1,size=8"`
	Level                       uint8   `mp4:"2,size=8"`
	BitDepth                    uint8   `mp4:"3,size=4"`
	ChromaSubsampling           uint8   `mp4:"4,size=3"`
	VideoFullRangeFlag          uint8   `mp4:"5,size=1"`
	ColourPrimaries             uint8   `mp4:"6,size=8"`
	TransferCharacteristics     uint8   `mp4:"7,size=8"`
	MatrixCoefficients          uint8   `mp4:"8,size=8"`
	CodecInitializationDataSize uint16  `mp4:"9,size=16"`
	CodecInitializationData     []uint8 `mp4:"10,size=8,len=dynamic"`
}

func (VpcC) GetType() mp4.BoxType {
	return BoxTypeVpcC()
}

func (vpcc VpcC) GetFieldLength(name string, ctx mp4.Context) uint {
	switch name {
	case "CodecInitializationData":
		return uint(vpcc.CodecInitializationDataSize)
	}
	return 0
}
