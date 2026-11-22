package mpeg4audio_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
)

var audioSpecificConfigCases = []struct {
	name string
	enc  []byte
	dec  mpeg4audio.AudioSpecificConfig
}{
	{
		"aac-lc 16khz mono",
		[]byte{0x14, 0x08},
		mpeg4audio.AudioSpecificConfig{
			Type:          mpeg4audio.ObjectTypeAACLC,
			SampleRate:    16000,
			ChannelConfig: 1,
			ChannelCount:  1,
		},
	},
	{
		"aac-lc 44.1khz mono",
		[]byte{0x12, 0x08},
		mpeg4audio.AudioSpecificConfig{
			Type:          mpeg4audio.ObjectTypeAACLC,
			SampleRate:    44100,
			ChannelConfig: 1,
			ChannelCount:  1,
		},
	},
	{
		"aac-lc 44.1khz 5.1",
		[]byte{0x12, 0x30},
		mpeg4audio.AudioSpecificConfig{
			Type:          mpeg4audio.ObjectTypeAACLC,
			SampleRate:    44100,
			ChannelConfig: 6,
			ChannelCount:  6,
		},
	},
	{
		"aac-lc 48khz stereo",
		[]byte{17, 144},
		mpeg4audio.AudioSpecificConfig{
			Type:          mpeg4audio.ObjectTypeAACLC,
			SampleRate:    48000,
			ChannelConfig: 2,
			ChannelCount:  2,
		},
	},
	{
		"aac-lc 53khz stereo",
		[]byte{0x17, 0x80, 0x67, 0x84, 0x10},
		mpeg4audio.AudioSpecificConfig{
			Type:          mpeg4audio.ObjectTypeAACLC,
			SampleRate:    53000,
			ChannelConfig: 2,
			ChannelCount:  2,
		},
	},
	{
		"aac-lc 96khz stereo delay",
		[]byte{0x10, 0x12, 0x0c, 0x08},
		mpeg4audio.AudioSpecificConfig{
			Type:               mpeg4audio.ObjectTypeAACLC,
			SampleRate:         96000,
			ChannelConfig:      2,
			ChannelCount:       2,
			DependsOnCoreCoder: true,
			CoreCoderDelay:     385,
		},
	},
	{
		"aac-lc 44.1khz 8 chans",
		[]byte{0x12, 0x38},
		mpeg4audio.AudioSpecificConfig{
			Type:          mpeg4audio.ObjectTypeAACLC,
			SampleRate:    44100,
			ChannelConfig: 7,
			ChannelCount:  8,
		},
	},
	{
		"aac-lc 48khz channel_config=0",
		// Object type (5 bits): 2 (AAC-LC), Sample rate idx (4 bits): 3 (48kHz),
		// Channel config (4 bits): 0, GASpecificConfig: 000
		// 00010 0011 0000 000 = 0x11 0x80
		[]byte{0x11, 0x80},
		mpeg4audio.AudioSpecificConfig{
			Type:          mpeg4audio.ObjectTypeAACLC,
			SampleRate:    48000,
			ChannelConfig: 0,
			ChannelCount:  0, // PCE defines channel layout
		},
	},
	{
		"sbr (he-aac v1) 44.1khz mono",
		[]byte{0x2b, 0x8a, 0x08, 0x00},
		mpeg4audio.AudioSpecificConfig{
			Type:                mpeg4audio.ObjectTypeAACLC,
			SampleRate:          22050,
			ChannelConfig:       1,
			ChannelCount:        1,
			ExtensionSampleRate: 44100,
			ExtensionType:       mpeg4audio.ObjectTypeSBR,
		},
	},
	{
		"sbr (he-aac v1) 44.1khz stereo",
		[]byte{0x2b, 0x92, 0x08, 0x00}, // the data from fdk_aac
		mpeg4audio.AudioSpecificConfig{
			Type:                mpeg4audio.ObjectTypeAACLC,
			SampleRate:          22050,
			ChannelConfig:       2,
			ChannelCount:        2,
			ExtensionSampleRate: 44100,
			ExtensionType:       mpeg4audio.ObjectTypeSBR,
		},
	},
	{
		"sbr (he-aac v1) 96khz stereo",
		[]byte{0x29, 0x90, 0x08, 0x00},
		mpeg4audio.AudioSpecificConfig{
			Type:                mpeg4audio.ObjectTypeAACLC,
			SampleRate:          48000,
			ChannelConfig:       2,
			ChannelCount:        2,
			ExtensionSampleRate: 96000,
			ExtensionType:       mpeg4audio.ObjectTypeSBR,
		},
	},
	{
		"ps (he-aac v2) 48khz stereo",
		[]byte{0xeb, 0x09, 0x88, 0x00}, // the data from fdk_aac
		mpeg4audio.AudioSpecificConfig{
			Type:                mpeg4audio.ObjectTypeAACLC,
			SampleRate:          24000,
			ChannelConfig:       1,
			ChannelCount:        1,
			ExtensionSampleRate: 48000,
			ExtensionType:       mpeg4audio.ObjectTypePS,
		},
	},
}

func TestAudioSpecificConfigUnmarshal(t *testing.T) {
	for _, ca := range audioSpecificConfigCases {
		t.Run(ca.name, func(t *testing.T) {
			var dec mpeg4audio.AudioSpecificConfig
			err := dec.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, dec)
		})
	}
}

func TestAudioSpecificConfigUnmarshalZeroSampleRate(t *testing.T) {
	enc, err := mpeg4audio.AudioSpecificConfig{
		Type:          mpeg4audio.ObjectTypeAACLC,
		SampleRate:    0,
		ChannelConfig: 2,
	}.Marshal()
	require.NoError(t, err)

	var dec mpeg4audio.AudioSpecificConfig
	require.EqualError(t, dec.Unmarshal(enc), "invalid sample rate: 0")
}

func TestAudioSpecificConfigUnmarshalZeroExtensionSampleRate(t *testing.T) {
	for _, ca := range []struct {
		name string
		enc  []byte
	}{
		{"sbr", []byte{0x2b, 0x0f, 0x80, 0x00, 0x00, 0x08, 0x00}},
		{"ps", []byte{0xeb, 0x0f, 0x80, 0x00, 0x00, 0x08, 0x00}},
	} {
		t.Run(ca.name, func(t *testing.T) {
			var dec mpeg4audio.AudioSpecificConfig
			require.EqualError(t, dec.Unmarshal(ca.enc), "invalid extension sample rate: 0")
		})
	}
}

func TestAudioSpecificConfigUnmarshalAdditional(t *testing.T) {
	for _, ca := range []struct {
		name string
		enc  []byte
		dec  mpeg4audio.AudioSpecificConfig
	}{
		{
			"additional bytes",
			[]byte{0x11, 0x90, 0x08, 0x10},
			mpeg4audio.AudioSpecificConfig{
				Type:          mpeg4audio.ObjectTypeAACLC,
				SampleRate:    48000,
				ChannelConfig: 2,
				ChannelCount:  2,
			},
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			var dec mpeg4audio.AudioSpecificConfig
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
	// Invalid channel count should error
	_, err := mpeg4audio.AudioSpecificConfig{
		Type:         mpeg4audio.ObjectTypeAACLC,
		SampleRate:   44100,
		ChannelCount: 9,
	}.Marshal()
	require.Error(t, err)
}

func FuzzAudioSpecificConfigUnmarshal(f *testing.F) {
	for _, ca := range audioSpecificConfigCases {
		f.Add(ca.enc)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		var conf mpeg4audio.AudioSpecificConfig
		err := conf.Unmarshal(b)
		if err != nil {
			return
		}

		// make sure that sample rates are never zero
		require.NotZero(t, conf.SampleRate)
		if conf.ExtensionType == mpeg4audio.ObjectTypeSBR || conf.ExtensionType == mpeg4audio.ObjectTypePS {
			require.NotZero(t, conf.ExtensionSampleRate)
		}

		_, err = conf.Marshal()
		require.NoError(t, err)
	})
}
