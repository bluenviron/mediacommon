package fmp4

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/fmp4/seekablebuffer"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4/codecs"
)

func TestEAC3Roundtrip(t *testing.T) {
	// Create an EAC3 codec with realistic values
	codec := &codecs.EAC3{
		SampleRate:   48000,
		ChannelCount: 6, // 5.1 surround
		DataRate:     640,
		NumIndSub:    0, // 1 independent substream
		Fscod:        0, // 48 kHz
		Bsid:         16,
		Asvc:         false,
		Bsmod:        0,
		Acmod:        7, // 3/2 (L, C, R, Ls, Rs)
		LfeOn:        true,
		NumDepSub:    0,
		ChanLoc:      0,
	}

	init := Init{
		Tracks: []*InitTrack{
			{
				ID:        1,
				TimeScale: 48000,
				Codec:     codec,
			},
		},
	}

	// Marshal to bytes
	var buf seekablebuffer.Buffer
	err := init.Marshal(&buf)
	require.NoError(t, err)

	// Unmarshal back
	var init2 Init
	err = init2.Unmarshal(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)

	// Verify the roundtrip
	require.Len(t, init2.Tracks, 1)
	require.Equal(t, int(1), init2.Tracks[0].ID)
	require.Equal(t, uint32(48000), init2.Tracks[0].TimeScale)

	codec2, ok := init2.Tracks[0].Codec.(*codecs.EAC3)
	require.True(t, ok, "expected CodecEAC3")

	require.Equal(t, 48000, codec2.SampleRate)
	require.Equal(t, 6, codec2.ChannelCount)
	require.Equal(t, uint16(640), codec2.DataRate)
	require.Equal(t, uint8(0), codec2.NumIndSub)
	require.Equal(t, uint8(0), codec2.Fscod)
	require.Equal(t, uint8(16), codec2.Bsid)
	require.Equal(t, false, codec2.Asvc)
	require.Equal(t, uint8(0), codec2.Bsmod)
	require.Equal(t, uint8(7), codec2.Acmod)
	require.Equal(t, true, codec2.LfeOn)
	require.Equal(t, uint8(0), codec2.NumDepSub)
	require.Equal(t, uint16(0), codec2.ChanLoc)
}

func TestEAC3StereoRoundtrip(t *testing.T) {
	// Test with stereo configuration
	codec := &codecs.EAC3{
		SampleRate:   48000,
		ChannelCount: 2, // Stereo
		DataRate:     128,
		NumIndSub:    0,
		Fscod:        0, // 48 kHz
		Bsid:         16,
		Asvc:         false,
		Bsmod:        0,
		Acmod:        2, // 2/0 (L, R)
		LfeOn:        false,
		NumDepSub:    0,
		ChanLoc:      0,
	}

	init := Init{
		Tracks: []*InitTrack{
			{
				ID:        1,
				TimeScale: 48000,
				Codec:     codec,
			},
		},
	}

	var buf seekablebuffer.Buffer
	err := init.Marshal(&buf)
	require.NoError(t, err)

	var init2 Init
	err = init2.Unmarshal(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)

	codec2, ok := init2.Tracks[0].Codec.(*codecs.EAC3)
	require.True(t, ok, "expected CodecEAC3")

	require.Equal(t, 2, codec2.ChannelCount)
	require.Equal(t, uint8(2), codec2.Acmod)
	require.Equal(t, false, codec2.LfeOn)
}

func TestEAC3_44100HzRoundtrip(t *testing.T) {
	// Test with 44.1 kHz sample rate
	codec := &codecs.EAC3{
		SampleRate:   44100,
		ChannelCount: 2,
		DataRate:     128,
		NumIndSub:    0,
		Fscod:        1, // 44.1 kHz
		Bsid:         16,
		Asvc:         false,
		Bsmod:        0,
		Acmod:        2,
		LfeOn:        false,
		NumDepSub:    0,
		ChanLoc:      0,
	}

	init := Init{
		Tracks: []*InitTrack{
			{
				ID:        1,
				TimeScale: 44100,
				Codec:     codec,
			},
		},
	}

	var buf seekablebuffer.Buffer
	err := init.Marshal(&buf)
	require.NoError(t, err)

	var init2 Init
	err = init2.Unmarshal(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)

	codec2, ok := init2.Tracks[0].Codec.(*codecs.EAC3)
	require.True(t, ok, "expected CodecEAC3")

	require.Equal(t, 44100, codec2.SampleRate)
	require.Equal(t, uint8(1), codec2.Fscod)
}

func TestEAC3WithDependentSubstream(t *testing.T) {
	// Test with a dependent substream (7.1 surround via extension)
	codec := &codecs.EAC3{
		SampleRate:   48000,
		ChannelCount: 8, // 7.1 surround
		DataRate:     768,
		NumIndSub:    0,
		Fscod:        0,
		Bsid:         16,
		Asvc:         false,
		Bsmod:        0,
		Acmod:        7,
		LfeOn:        true,
		NumDepSub:    1,           // 1 dependent substream
		ChanLoc:      0b000011000, // Lrs, Rrs positions
	}

	init := Init{
		Tracks: []*InitTrack{
			{
				ID:        1,
				TimeScale: 48000,
				Codec:     codec,
			},
		},
	}

	var buf seekablebuffer.Buffer
	err := init.Marshal(&buf)
	require.NoError(t, err)

	var init2 Init
	err = init2.Unmarshal(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)

	codec2, ok := init2.Tracks[0].Codec.(*codecs.EAC3)
	require.True(t, ok, "expected CodecEAC3")

	require.Equal(t, 8, codec2.ChannelCount)
	require.Equal(t, uint8(1), codec2.NumDepSub)
	require.Equal(t, uint16(0b000011000), codec2.ChanLoc)
}
