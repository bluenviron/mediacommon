package eac3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var eac3Cases = []struct {
	name         string
	enc          []byte
	syncInfo     SyncInfo
	frameSizes   int
	sampleRate   int
	channelCount int
	numBlocks    int
}{
	{
		"fscod 0",
		[]byte{
			0x0B, 0x77, // sync word
			0x01, 0xFF, // strmtyp=0, substreamid=0, frmsiz=0x1FF
			0x3F,       // fscod=0, numblkscod=3, acmod=7, lfeon=1
			0x80,       // bsid=16
			0x00, 0x00, // padding
		},
		SyncInfo{
			Strmtyp:     0,
			Substreamid: 0,
			Frmsiz:      0x1FF,
			Fscod:       0,
			Numblkscod:  3,
			Acmod:       7,
			Lfeon:       true,
			Bsid:        16,
		},
		1024,
		48000,
		6,
		6,
	},
	{
		"fscod 2",
		[]byte{
			0x0B, 0x77, // sync word
			0x01, 0xFF, // strmtyp=0, substreamid=0, frmsiz=0x1FF
			0xC4, // fscod=3, fscod2=0, acmod=2, lfeon=0
			0x80, // bsid=16
			0x00, 0x00,
		},
		SyncInfo{
			Strmtyp:     0,
			Substreamid: 0,
			Frmsiz:      0x1FF,
			Fscod:       3,
			Fscod2:      0,
			Numblkscod:  3,
			Acmod:       2,
			Lfeon:       false,
			Bsid:        16,
		},
		1024,
		24000,
		2,
		6,
	},
}

func TestSyncInfoUnmarshal(t *testing.T) {
	for _, ca := range eac3Cases {
		t.Run(ca.name, func(t *testing.T) {
			var s SyncInfo
			err := s.Unmarshal(ca.enc)
			require.NoError(t, err)

			require.Equal(t, ca.syncInfo, s)

			require.Equal(t, ca.frameSizes, s.FrameSize())
			require.Equal(t, ca.sampleRate, s.SampleRate())
			require.Equal(t, ca.channelCount, s.ChannelCount())
			require.Equal(t, ca.numBlocks, s.NumBlocks())
		})
	}
}

func FuzzSyncInfoUnmarshal(f *testing.F) {
	for _, ca := range eac3Cases {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var syncInfo SyncInfo
		err := syncInfo.Unmarshal(b)
		if err != nil {
			return
		}

		syncInfo.FrameSize() //nolint:staticcheck
		syncInfo.SampleRate()
		syncInfo.ChannelCount()
		syncInfo.NumBlocks() //nolint:staticcheck
	})
}
