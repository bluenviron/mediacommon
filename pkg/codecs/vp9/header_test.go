package vp9

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesHeader = []struct {
	name   string
	byts   []byte
	sh     Header
	width  int
	height int
}{
	{
		"chrome webrtc",
		[]byte{
			0x82, 0x49, 0x83, 0x42, 0x00, 0x77, 0xf0, 0x32,
			0x34, 0x30, 0x38, 0x24, 0x1c, 0x19, 0x40, 0x18,
			0x03, 0x40, 0x5f, 0xb4,
		},
		Header{
			ShowFrame: true,
			ColorConfig: &Header_ColorConfig{
				BitDepth:     8,
				SubsamplingX: true,
				SubsamplingY: true,
			},
			FrameSize: &Header_FrameSize{
				FrameWidthMinus1:  1919,
				FrameHeightMinus1: 803,
			},
		},
		1920,
		804,
	},
	{
		"vp9 sample",
		[]byte{
			0x82, 0x49, 0x83, 0x42, 0x40, 0xef, 0xf0, 0x86,
			0xf4, 0x04, 0x21, 0xa0, 0xe0, 0x00, 0x30, 0x70,
			0x00, 0x00, 0x00, 0x01,
		},
		Header{
			ShowFrame: true,
			ColorConfig: &Header_ColorConfig{
				BitDepth:     8,
				ColorSpace:   2,
				SubsamplingX: true,
				SubsamplingY: true,
			},
			FrameSize: &Header_FrameSize{
				FrameWidthMinus1:  3839,
				FrameHeightMinus1: 2159,
			},
		},
		3840,
		2160,
	},
}

func TestHeaderUnmarshal(t *testing.T) {
	for _, ca := range casesHeader {
		t.Run(ca.name, func(t *testing.T) {
			var sh Header
			err := sh.Unmarshal(ca.byts)
			require.NoError(t, err)
			require.Equal(t, ca.sh, sh)
			require.Equal(t, ca.width, sh.Width())
			require.Equal(t, ca.height, sh.Height())
		})
	}
}

func FuzzHeaderUnmarshal(f *testing.F) {
	for _, ca := range casesHeader {
		f.Add(ca.byts)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var sh Header
		err := sh.Unmarshal(b) //nolint:errcheck
		if err == nil {
			sh.Width()
			sh.Height()
			sh.ChromaSubsampling()
		}
	})
}
