package flac

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var streamInfoCases = []struct {
	name string
	dec  StreamInfo
	enc  []byte
}{
	{
		"stereo 44.1kHz 16-bit",
		StreamInfo{
			MinBlockSize: 4096,
			MaxBlockSize: 4096,
			MinFrameSize: 0,
			MaxFrameSize: 0,
			SampleRate:   44100,
			ChannelCount: 2,
			BitDepth:     16,
			TotalSamples: 0,
			MD5:          [16]byte{},
		},
		[]byte{
			0x10, 0x00, // MinBlockSize = 4096
			0x10, 0x00, // MaxBlockSize = 4096
			0x00, 0x00, 0x00, // MinFrameSize = 0
			0x00, 0x00, 0x00, // MaxFrameSize = 0
			0x0A, 0xC4, 0x42, 0xF0, // SampleRate=44100, ChannelCount=2, BitDepth=16
			0x00, 0x00, 0x00, 0x00, // TotalSamples = 0
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // MD5
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
	},
	{
		"5.1 surround 96kHz 24-bit with samples and MD5",
		StreamInfo{
			MinBlockSize: 256,
			MaxBlockSize: 65535,
			MinFrameSize: 100,
			MaxFrameSize: 50000,
			SampleRate:   96000,
			ChannelCount: 6,
			BitDepth:     24,
			TotalSamples: 1000000,
			MD5: [16]byte{
				0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
				0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
			},
		},
		[]byte{
			0x01, 0x00, // MinBlockSize = 256
			0xFF, 0xFF, // MaxBlockSize = 65535
			0x00, 0x00, 0x64, // MinFrameSize = 100
			0x00, 0xC3, 0x50, // MaxFrameSize = 50000
			0x17, 0x70, 0x0B, 0x70, // SampleRate=96000, ChannelCount=6, BitDepth=24
			0x00, 0x0F, 0x42, 0x40, // TotalSamples = 1000000
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, // MD5
			0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
		},
	},
	{
		"mono 48kHz 32-bit large TotalSamples",
		StreamInfo{
			MinBlockSize: 16,
			MaxBlockSize: 4096,
			MinFrameSize: 1024,
			MaxFrameSize: 131072,
			SampleRate:   48000,
			ChannelCount: 1,
			BitDepth:     32,
			TotalSamples: 0x100000000,
			MD5:          [16]byte{},
		},
		[]byte{
			0x00, 0x10, // MinBlockSize = 16
			0x10, 0x00, // MaxBlockSize = 4096
			0x00, 0x04, 0x00, // MinFrameSize = 1024
			0x02, 0x00, 0x00, // MaxFrameSize = 131072
			0x0B, 0xB8, 0x01, 0xF1, // SampleRate=48000, ChannelCount=1, BitDepth=32
			0x00, 0x00, 0x00, 0x00, // TotalSamples = 0x100000000
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // MD5
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
	},
}

func TestStreamInfoMarshal(t *testing.T) {
	for _, ca := range streamInfoCases {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func TestStreamInfoUnmarshal(t *testing.T) {
	for _, ca := range streamInfoCases {
		t.Run(ca.name, func(t *testing.T) {
			var s StreamInfo
			err := s.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, s)
		})
	}
}

func FuzzStreamInfoUnmarshal(f *testing.F) {
	for _, ca := range streamInfoCases {
		f.Add(ca.enc)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		var s StreamInfo
		err := s.Unmarshal(b)
		if err != nil {
			return
		}

		_, err = s.Marshal()
		require.NoError(t, err)
	})
}
