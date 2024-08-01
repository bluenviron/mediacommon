package mpeg1audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesFrameHeader = []struct {
	name        string
	enc         []byte
	dec         FrameHeader
	frameLen    int
	sampleCount int
}{
	{
		"mpeg-1 layer 2 32k",
		[]byte{
			0xff, 0xfd, 0x48, 0x00, 0x00,
		},
		FrameHeader{
			Layer:       2,
			Bitrate:     64000,
			SampleRate:  32000,
			ChannelMode: ChannelModeStereo,
		},
		288,
		1152,
	},
	{
		"mpeg-1 layer 3 32k",
		[]byte{
			0xff, 0xfb, 0x18, 0x64, 0x00,
		},
		FrameHeader{
			Layer:       3,
			Bitrate:     32000,
			SampleRate:  32000,
			ChannelMode: ChannelModeJointStereo,
		},
		144,
		1152,
	},
	{
		"mpeg-1 layer 3 32k mono",
		[]byte{
			0xff, 0xfb, 0x68, 0xc4, 0x00,
		},
		FrameHeader{
			Layer:       3,
			Bitrate:     80000,
			SampleRate:  32000,
			ChannelMode: ChannelModeMono,
		},
		360,
		1152,
	},
	{
		"mpeg-1 layer 3 44.1k",
		[]byte{
			0xff, 0xfa, 0x52, 0x04, 0x00,
		},
		FrameHeader{
			Layer:       3,
			Bitrate:     64000,
			SampleRate:  44100,
			Padding:     true,
			ChannelMode: ChannelModeStereo,
		},
		209,
		1152,
	},
	{
		"mpeg-1 layer 3 48k",
		[]byte{
			0xff, 0xfb, 0x14, 0x64, 0x00,
		},
		FrameHeader{
			Layer:       3,
			Bitrate:     32000,
			SampleRate:  48000,
			ChannelMode: ChannelModeJointStereo,
		},
		96,
		1152,
	},
	{
		"mpeg-2 layer 2 16khz",
		[]byte{
			0xff, 0xf5, 0x88, 0x4, 0x00,
		},
		FrameHeader{
			MPEG2:       true,
			Layer:       2,
			Bitrate:     64000,
			SampleRate:  16000,
			ChannelMode: ChannelModeStereo,
		},
		576,
		1152,
	},
}

func TestFrameHeaderUnmarshal(t *testing.T) {
	for _, ca := range casesFrameHeader {
		t.Run(ca.name, func(t *testing.T) {
			var h FrameHeader
			err := h.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, h)
			require.Equal(t, ca.frameLen, h.FrameLen())
			require.Equal(t, ca.sampleCount, h.SampleCount())
		})
	}
}

func FuzzFrameHeaderUnmarshal(f *testing.F) {
	for _, ca := range casesFrameHeader {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var h FrameHeader
		err := h.Unmarshal(b)
		if err == nil {
			h.FrameLen()    //nolint:staticcheck
			h.SampleCount() //nolint:staticcheck
		}
	})
}
