package opus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var idHeaderCases = []struct {
	name string
	dec  IDHeader
	enc  []byte
}{
	{
		"stereo 48kHz",
		IDHeader{
			Version:              1,
			ChannelCount:         2,
			PreSkip:              312,
			InputSampleRate:      48000,
			OutputGain:           0,
			ChannelMappingFamily: 0,
			ChannelMappingTable:  []uint8{},
		},
		[]byte{
			0x4F, 0x70, 0x75, 0x73, 0x48, 0x65, 0x61, 0x64, // OpusHead
			0x01,       // version
			0x02,       // channel count
			0x01, 0x38, // pre-skip = 312
			0x00, 0x00, 0xBB, 0x80, // input sample rate = 48000
			0x00, 0x00, // output gain = 0
			0x00, // channel mapping family
		},
	},
	{
		"5.1 with mapping table",
		IDHeader{
			Version:              1,
			ChannelCount:         6,
			PreSkip:              0,
			InputSampleRate:      48000,
			OutputGain:           256,
			ChannelMappingFamily: 1,
			ChannelMappingTable:  []uint8{0, 4, 1, 2, 3, 5},
		},
		[]byte{
			0x4F, 0x70, 0x75, 0x73, 0x48, 0x65, 0x61, 0x64, // OpusHead
			0x01,       // version
			0x06,       // channel count
			0x00, 0x00, // pre-skip = 0
			0x00, 0x00, 0xBB, 0x80, // input sample rate = 48000
			0x01, 0x00, // output gain = 256
			0x01,                               // channel mapping family
			0x00, 0x04, 0x01, 0x02, 0x03, 0x05, // channel mapping table
		},
	},
	{
		"mono 44.1kHz with gain",
		IDHeader{
			Version:              1,
			ChannelCount:         1,
			PreSkip:              0xABCD,
			InputSampleRate:      44100,
			OutputGain:           0xFF00,
			ChannelMappingFamily: 255,
			ChannelMappingTable:  []uint8{},
		},
		[]byte{
			0x4F, 0x70, 0x75, 0x73, 0x48, 0x65, 0x61, 0x64, // OpusHead
			0x01,       // version
			0x01,       // channel count
			0xAB, 0xCD, // pre-skip = 0xABCD
			0x00, 0x00, 0xAC, 0x44, // input sample rate = 44100
			0xFF, 0x00, // output gain = 0xFF00
			0xFF, // channel mapping family
		},
	},
}

func TestIDHeaderMarshal(t *testing.T) {
	for _, ca := range idHeaderCases {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func TestIDHeaderUnmarshal(t *testing.T) {
	for _, ca := range idHeaderCases {
		t.Run(ca.name, func(t *testing.T) {
			var h IDHeader
			err := h.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, h)
		})
	}
}

func FuzzIDHeaderUnmarshal(f *testing.F) {
	for _, ca := range idHeaderCases {
		f.Add(ca.enc)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		var h IDHeader
		err := h.Unmarshal(b)
		if err != nil {
			return
		}

		_, err = h.Marshal()
		require.NoError(t, err)
	})
}
