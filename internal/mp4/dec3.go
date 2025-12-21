package mp4

import (
	amp4 "github.com/abema/go-mp4"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

func init() {
	// Register ec-3 as an AudioSampleEntry type (like ac-3)
	amp4.AddAnyTypeBoxDef(&amp4.AudioSampleEntry{}, amp4.StrToBoxType("ec-3"))
	// Register dec3 box type
	amp4.AddBoxDef(&Dec3{})
}

// Dec3 is the E-AC-3 decoder configuration box (EC3SpecificBox).
// This implements the structure defined in ETSI TS 102 366.
//
// Note: This implementation supports the common case of a single independent
// substream with no dependent substreams. For content with multiple substreams,
// the raw Payload field can be used.
type Dec3 struct {
	amp4.Box

	// DataRate is the peak data rate in kbps (13 bits).
	DataRate uint16 `mp4:"0,size=13"`

	// NumIndSub is the number of independent substreams minus 1 (3 bits).
	// A value of 0 means 1 independent substream.
	NumIndSub uint8 `mp4:"1,size=3"`

	// Fields for the first (primary) independent substream:

	// Fscod is the sample rate code (2 bits).
	// 0 = 48 kHz, 1 = 44.1 kHz, 2 = 32 kHz, 3 = reserved.
	Fscod uint8 `mp4:"2,size=2"`

	// Bsid is the bit stream identification (5 bits).
	// For E-AC-3, this should be 16.
	Bsid uint8 `mp4:"3,size=5"`

	// Reserved1 is a reserved bit, always 0 (1 bit).
	Reserved1 uint8 `mp4:"4,size=1,const=0"`

	// Asvc indicates if this is an associated audio service (1 bit).
	Asvc uint8 `mp4:"5,size=1"`

	// Bsmod is the bit stream mode (3 bits).
	Bsmod uint8 `mp4:"6,size=3"`

	// Acmod is the audio coding mode (3 bits).
	Acmod uint8 `mp4:"7,size=3"`

	// LfeOn indicates if the LFE channel is present (1 bit).
	LfeOn uint8 `mp4:"8,size=1"`

	// Reserved2 is reserved (3 bits).
	Reserved2 uint8 `mp4:"9,size=3,const=0"`

	// NumDepSub is the number of dependent substreams (4 bits).
	NumDepSub uint8 `mp4:"10,size=4"`

	// ChanLoc is the channel location bitmap (9 bits).
	// Only present when NumDepSub > 0, otherwise 1 reserved bit.
	// For simplicity, we always include the 9-bit field.
	ChanLoc uint16 `mp4:"11,size=9"`
}

// GetType returns the box type for dec3.
func (*Dec3) GetType() amp4.BoxType {
	return amp4.StrToBoxType("dec3")
}

// ToCodec converts the Dec3 box to a CodecEAC3.
func (d *Dec3) ToCodec(sampleRate, channelCount int) *mp4.CodecEAC3 {
	return &mp4.CodecEAC3{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		DataRate:     d.DataRate,
		NumIndSub:    d.NumIndSub,
		Fscod:        d.Fscod,
		Bsid:         d.Bsid,
		Asvc:         d.Asvc != 0,
		Bsmod:        d.Bsmod,
		Acmod:        d.Acmod,
		LfeOn:        d.LfeOn != 0,
		NumDepSub:    d.NumDepSub,
		ChanLoc:      d.ChanLoc,
	}
}

// FromCodec creates a Dec3 box from a CodecEAC3.
func FromCodec(codec *mp4.CodecEAC3) *Dec3 {
	var asvc, lfeon uint8
	if codec.Asvc {
		asvc = 1
	}
	if codec.LfeOn {
		lfeon = 1
	}
	return &Dec3{
		DataRate:  codec.DataRate,
		NumIndSub: codec.NumIndSub,
		Fscod:     codec.Fscod,
		Bsid:      codec.Bsid,
		Asvc:      asvc,
		Bsmod:     codec.Bsmod,
		Acmod:     codec.Acmod,
		LfeOn:     lfeon,
		NumDepSub: codec.NumDepSub,
		ChanLoc:   codec.ChanLoc,
	}
}
