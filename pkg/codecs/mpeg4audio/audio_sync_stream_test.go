package mpeg4audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesAudioSyncStream = []struct {
	name string
	dec  AudioSyncStream
	enc  []byte
}{
	{
		"a",
		AudioSyncStream{
			AudioMuxElements: [][]byte{
				{1, 2, 3, 4},
				{5, 6, 7, 8},
			},
		},
		[]byte{
			0x56, 0xe0, 0x04, 0x01, 0x02, 0x03, 0x04, 0x56,
			0xe0, 0x04, 0x05, 0x06, 0x07, 0x08,
		},
	},
}

func TestAudioSyncStreamUnmarshal(t *testing.T) {
	for _, ca := range casesAudioSyncStream {
		t.Run(ca.name, func(t *testing.T) {
			var dec AudioSyncStream
			err := dec.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, dec)
		})
	}
}

func TestAudioSyncStreamMarshal(t *testing.T) {
	for _, ca := range casesAudioSyncStream {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func FuzzAudioSyncStreamUnmarshal(f *testing.F) {
	for _, ca := range casesAudioSyncStream {
		f.Add(ca.enc)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		var dec AudioSyncStream
		err := dec.Unmarshal(b)
		if err != nil {
			return
		}

		_, err = dec.Marshal()
		require.NoError(t, err)
	})
}
