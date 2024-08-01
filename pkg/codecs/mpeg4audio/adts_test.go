package mpeg4audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesADTS = []struct {
	name string
	byts []byte
	pkts ADTSPackets
}{
	{
		"single",
		[]byte{0xff, 0xf1, 0x4c, 0x80, 0x1, 0x3f, 0xfc, 0xaa, 0xbb},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC,
				SampleRate:   48000,
				ChannelCount: 2,
				AU:           []byte{0xaa, 0xbb},
			},
		},
	},
	{
		"multiple",
		[]byte{
			0xff, 0xf1, 0x50, 0x40, 0x1, 0x3f, 0xfc, 0xaa,
			0xbb, 0xff, 0xf1, 0x4c, 0x80, 0x1, 0x3f, 0xfc,
			0xcc, 0xdd,
		},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC,
				SampleRate:   44100,
				ChannelCount: 1,
				AU:           []byte{0xaa, 0xbb},
			},
			{
				Type:         ObjectTypeAACLC,
				SampleRate:   48000,
				ChannelCount: 2,
				AU:           []byte{0xcc, 0xdd},
			},
		},
	},
}

func TestADTSUnmarshal(t *testing.T) {
	for _, ca := range casesADTS {
		t.Run(ca.name, func(t *testing.T) {
			var pkts ADTSPackets
			err := pkts.Unmarshal(ca.byts)
			require.NoError(t, err)
			require.Equal(t, ca.pkts, pkts)
		})
	}
}

func TestADTSMarshal(t *testing.T) {
	for _, ca := range casesADTS {
		t.Run(ca.name, func(t *testing.T) {
			byts, err := ca.pkts.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.byts, byts)
		})
	}
}

func FuzzADTSUnmarshal(f *testing.F) {
	for _, ca := range casesADTS {
		f.Add(ca.byts)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var pkts ADTSPackets
		err := pkts.Unmarshal(b)
		if err == nil {
			pkts.Marshal() //nolint:errcheck
		}
	})
}
