package mpeg4audio_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
)

var casesAudioMuxElement = []struct {
	name string
	dec  mpeg4audio.AudioMuxElement
	enc  []byte
}{
	{
		"muxConfigPresent + !useSameStreamMux",
		mpeg4audio.AudioMuxElement{
			MuxConfigPresent: true,
			StreamMuxConfig: &mpeg4audio.StreamMuxConfig{
				Programs: []*mpeg4audio.StreamMuxConfigProgram{
					{
						Layers: []*mpeg4audio.StreamMuxConfigLayer{
							{
								AudioSpecificConfig: &mpeg4audio.AudioSpecificConfig{
									Type:          2,
									SampleRate:    24000,
									ChannelConfig: 1,
									ChannelCount:  1,
								},
								LatmBufferFullness: 255,
							},
							{
								AudioSpecificConfig: &mpeg4audio.AudioSpecificConfig{
									Type:          2,
									SampleRate:    48000,
									ChannelConfig: 1,
									ChannelCount:  1,
								},
								LatmBufferFullness: 255,
							},
						},
					},
					{
						Layers: []*mpeg4audio.StreamMuxConfigLayer{
							{
								AudioSpecificConfig: &mpeg4audio.AudioSpecificConfig{
									Type:          2,
									SampleRate:    44100,
									ChannelConfig: 1,
									ChannelCount:  1,
								},
								LatmBufferFullness: 255,
							},
						},
					},
				},
			},
			UseSameStreamMux: false,
			Payloads: [][][][]byte{
				{
					{
						{1, 2, 3, 4},
						{5, 6, 7, 8},
					},
					{
						{9, 10, 11, 12},
					},
				},
			},
		},
		[]byte{
			0x20, 0x09, 0x13, 0x08, 0x1f, 0xe1, 0x18, 0x81,
			0xfe, 0x02, 0x41, 0x03, 0xfc, 0x04, 0x04, 0x04,
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c,
		},
	},
	{
		"muxConfigPresent + useSameStreamMux",
		mpeg4audio.AudioMuxElement{
			MuxConfigPresent: true,
			StreamMuxConfig: &mpeg4audio.StreamMuxConfig{
				Programs: []*mpeg4audio.StreamMuxConfigProgram{{
					Layers: []*mpeg4audio.StreamMuxConfigLayer{{
						AudioSpecificConfig: &mpeg4audio.AudioSpecificConfig{
							Type:         2,
							SampleRate:   24000,
							ChannelCount: 1,
						},
						LatmBufferFullness: 255,
					}},
				}},
			},
			UseSameStreamMux: true,
			Payloads: [][][][]byte{
				{
					{
						{1, 2, 3, 4, 5},
					},
				},
			},
		},
		[]byte{
			0x82, 0x80, 0x81, 0x1, 0x82, 0x2, 0x80,
		},
	},
	{
		"!muxConfigPresent",
		mpeg4audio.AudioMuxElement{
			MuxConfigPresent: false,
			StreamMuxConfig: &mpeg4audio.StreamMuxConfig{
				Programs: []*mpeg4audio.StreamMuxConfigProgram{{
					Layers: []*mpeg4audio.StreamMuxConfigLayer{{
						AudioSpecificConfig: &mpeg4audio.AudioSpecificConfig{
							Type:         2,
							SampleRate:   24000,
							ChannelCount: 1,
						},
						LatmBufferFullness: 255,
					}},
				}},
			},
			Payloads: [][][][]byte{
				{
					{
						{1, 2, 3, 4, 5},
					},
				},
			},
		},
		[]byte{
			0x5, 0x1, 0x2, 0x3, 0x4, 0x5,
		},
	},
}

func TestAudioMuxElementUnmarshal(t *testing.T) {
	for _, ca := range casesAudioMuxElement {
		t.Run(ca.name, func(t *testing.T) {
			var dec mpeg4audio.AudioMuxElement
			dec.MuxConfigPresent = ca.dec.MuxConfigPresent
			if !dec.MuxConfigPresent || ca.dec.UseSameStreamMux {
				dec.StreamMuxConfig = ca.dec.StreamMuxConfig
			}
			err := dec.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, dec)
		})
	}
}

func TestAudioMuxElementMarshal(t *testing.T) {
	for _, ca := range casesAudioMuxElement {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func FuzzAudioMuxElementUnmarshal(f *testing.F) {
	for _, ca := range casesAudioMuxElement {
		f.Add(ca.dec.MuxConfigPresent, ca.enc)
	}

	f.Fuzz(func(t *testing.T, muxConfigPresent bool, b []byte) {
		var dec mpeg4audio.AudioMuxElement
		dec.MuxConfigPresent = muxConfigPresent
		if !muxConfigPresent {
			dec.StreamMuxConfig = &mpeg4audio.StreamMuxConfig{
				Programs: []*mpeg4audio.StreamMuxConfigProgram{
					{
						Layers: []*mpeg4audio.StreamMuxConfigLayer{
							{
								AudioSpecificConfig: &mpeg4audio.AudioSpecificConfig{
									Type:         2,
									SampleRate:   24000,
									ChannelCount: 1,
								},
								LatmBufferFullness: 255,
							},
						},
					},
				},
			}
		}

		err := dec.Unmarshal(b)
		if err != nil {
			return
		}

		_, err = dec.Marshal()
		require.NoError(t, err)
	})
}
