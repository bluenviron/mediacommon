package mpeg4audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var audioSpecificConfigCases = []struct {
	name string
	enc  []byte
	dec  AudioSpecificConfig
}{
	{
		"aac-lc 16khz mono",
		[]byte{0x14, 0x08},
		AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   16000,
			ChannelCount: 1,
		},
	},
	{
		"aac-lc 44.1khz mono",
		[]byte{0x12, 0x08},
		AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   44100,
			ChannelCount: 1,
		},
	},
	{
		"aac-lc 44.1khz 5.1",
		[]byte{0x12, 0x30},
		AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   44100,
			ChannelCount: 6,
		},
	},
	{
		"aac-lc 48khz stereo",
		[]byte{17, 144},
		AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   48000,
			ChannelCount: 2,
		},
	},
	{
		"aac-lc 53khz stereo",
		[]byte{0x17, 0x80, 0x67, 0x84, 0x10},
		AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   53000,
			ChannelCount: 2,
		},
	},
	{
		"aac-lc 96khz stereo delay",
		[]byte{0x10, 0x12, 0x0c, 0x08},
		AudioSpecificConfig{
			Type:               ObjectTypeAACLC,
			SampleRate:         96000,
			ChannelCount:       2,
			DependsOnCoreCoder: true,
			CoreCoderDelay:     385,
		},
	},
	{
		"aac-lc 44.1khz 8 chans",
		[]byte{0x12, 0x38},
		AudioSpecificConfig{
			Type:         ObjectTypeAACLC,
			SampleRate:   44100,
			ChannelCount: 8,
		},
	},
	{
		"sbr (he-aac v1) 44.1khz mono",
		[]byte{0x2b, 0x8a, 0x08, 0x00},
		AudioSpecificConfig{
			Type:                ObjectTypeAACLC,
			SampleRate:          22050,
			ChannelCount:        1,
			ExtensionSampleRate: 44100,
			ExtensionType:       ObjectTypeSBR,
		},
	},
	{
		"sbr (he-aac v1) 44.1khz stereo",
		[]byte{0x2b, 0x92, 0x08, 0x00}, // the data from fdk_aac
		AudioSpecificConfig{
			Type:                ObjectTypeAACLC,
			SampleRate:          22050,
			ChannelCount:        2,
			ExtensionSampleRate: 44100,
			ExtensionType:       ObjectTypeSBR,
		},
	},
	{
		"ps (he-aac v2) 48khz stereo",
		[]byte{0xeb, 0x09, 0x88, 0x00}, // the data from fdk_aac
		AudioSpecificConfig{
			Type:                ObjectTypeAACLC,
			SampleRate:          24000,
			ChannelCount:        1,
			ExtensionSampleRate: 48000,
			ExtensionType:       ObjectTypePS,
		},
	},
}

func TestAudioSpecificConfigUnmarshal(t *testing.T) {
	for _, ca := range audioSpecificConfigCases {
		t.Run(ca.name, func(t *testing.T) {
			var dec AudioSpecificConfig
			err := dec.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, dec)
		})
	}
}

func TestAudioSpecificConfigMarshal(t *testing.T) {
	for _, ca := range audioSpecificConfigCases {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func TestAudioSpecificConfigMarshalErrors(t *testing.T) {
	_, err := AudioSpecificConfig{
		Type:         ObjectTypeAACLC,
		SampleRate:   44100,
		ChannelCount: 0,
	}.Marshal()
	require.Error(t, err)
}

func FuzzAudioSpecificConfigUnmarshal(f *testing.F) {
	for _, ca := range audioSpecificConfigCases {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var conf AudioSpecificConfig
		err := conf.Unmarshal(b)
		if err == nil {
			conf.Marshal() //nolint:errcheck
		}
	})
}
