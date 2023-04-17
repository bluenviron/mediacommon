//go:build go1.18
// +build go1.18

package mpeg4audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var streamMuxConfigCases = []struct {
	name string
	enc  []byte
	dec  StreamMuxConfig
}{
	{
		"lc",
		[]byte{0x40, 0x00, 0x26, 0x20, 0x3f, 0xc0},
		StreamMuxConfig{
			Programs: []*StreamMuxConfigProgram{{
				Layers: []*StreamMuxConfigLayer{{
					AudioSpecificConfig: &AudioSpecificConfig{
						Type:         2,
						SampleRate:   24000,
						ChannelCount: 2,
					},
					LatmBufferFullness: 255,
				}},
			}},
		},
	},
	{
		"he-aac",
		[]byte{0x40, 0x00, 0x26, 0x10, 0x3f, 0xc0},
		StreamMuxConfig{
			Programs: []*StreamMuxConfigProgram{{
				Layers: []*StreamMuxConfigLayer{{
					AudioSpecificConfig: &AudioSpecificConfig{
						Type:         2,
						SampleRate:   24000,
						ChannelCount: 1,
					},
					LatmBufferFullness: 255,
				}},
			}},
		},
	},
	{
		"sbr",
		[]byte{0x40, 0x00, 0x56, 0x23, 0x10, 0x1f, 0xe0},
		StreamMuxConfig{
			Programs: []*StreamMuxConfigProgram{{
				Layers: []*StreamMuxConfigLayer{{
					AudioSpecificConfig: &AudioSpecificConfig{
						Type:                2,
						ExtensionType:       5,
						ExtensionSampleRate: 48000,
						SampleRate:          24000,
						ChannelCount:        2,
					},
					LatmBufferFullness: 255,
				}},
			}},
		},
	},
	{
		"ps",
		[]byte{0x40, 0x01, 0xd6, 0x13, 0x10, 0x1f, 0xe0},
		StreamMuxConfig{
			Programs: []*StreamMuxConfigProgram{{
				Layers: []*StreamMuxConfigLayer{{
					AudioSpecificConfig: &AudioSpecificConfig{
						Type:                2,
						ExtensionType:       29,
						ExtensionSampleRate: 48000,
						SampleRate:          24000,
						ChannelCount:        2,
					},
					LatmBufferFullness: 255,
				}},
			}},
		},
	},
}

func TestStreamMuxConfigUnmarshal(t *testing.T) {
	for _, ca := range streamMuxConfigCases {
		t.Run(ca.name, func(t *testing.T) {
			var dec StreamMuxConfig
			err := dec.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, dec)
		})
	}
}

func TestStreamMuxConfigUnmarshalTruncated(t *testing.T) {
	var dec StreamMuxConfig
	err := dec.Unmarshal([]byte{0x40, 0x00, 0x23, 0x10})
	require.NoError(t, err)
	require.Equal(t, StreamMuxConfig{
		Programs: []*StreamMuxConfigProgram{{
			Layers: []*StreamMuxConfigLayer{{
				AudioSpecificConfig: &AudioSpecificConfig{
					Type:         2,
					SampleRate:   48000,
					ChannelCount: 1,
				},
				LatmBufferFullness: 255,
			}},
		}},
	}, dec)
}

func TestStreamMuxConfigMarshal(t *testing.T) {
	for _, ca := range streamMuxConfigCases {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func FuzzStreamMuxConfigUnmarshal(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		var conf StreamMuxConfig
		conf.Unmarshal(b)
	})
}
