package mpeg4audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// casesADTS contains test cases for roundtrip (unmarshal -> marshal) tests.
// These use AAC-LC profile which marshals back to the same bytes.
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

// casesADTSUnmarshalOnly contains test cases for unmarshal-only tests.
// These test profile normalization where non-LC profiles are normalized to AAC-LC.
var casesADTSUnmarshalOnly = []struct {
	name string
	byts []byte
	pkts ADTSPackets
}{
	{
		// AAC Main profile in ADTS header (profile=0) - normalized to AAC-LC
		// Byte 2: 0x0c = profile(0)<<6 | sampleRateIndex(3)<<2 | channelConfig>>2
		// Many streams incorrectly signal AAC Main but contain AAC-LC data
		"aac_main_normalized",
		[]byte{0xff, 0xf1, 0x0c, 0x80, 0x1, 0x3f, 0xfc, 0xaa, 0xbb},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC, // Normalized to AAC-LC
				SampleRate:   48000,
				ChannelCount: 2,
				AU:           []byte{0xaa, 0xbb},
			},
		},
	},
	{
		// AAC SSR profile in ADTS header (profile=2) - normalized to AAC-LC
		// Byte 2: 0x8c = profile(2)<<6 | sampleRateIndex(3)<<2 | channelConfig>>2
		"aac_ssr_normalized",
		[]byte{0xff, 0xf1, 0x8c, 0x80, 0x1, 0x3f, 0xfc, 0xaa, 0xbb},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC, // Normalized to AAC-LC
				SampleRate:   48000,
				ChannelCount: 2,
				AU:           []byte{0xaa, 0xbb},
			},
		},
	},
	{
		// AAC LTP profile in ADTS header (profile=3) - normalized to AAC-LC
		// Byte 2: 0xcc = profile(3)<<6 | sampleRateIndex(3)<<2 | channelConfig>>2
		"aac_ltp_normalized",
		[]byte{0xff, 0xf1, 0xcc, 0x80, 0x1, 0x3f, 0xfc, 0xaa, 0xbb},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC, // Normalized to AAC-LC
				SampleRate:   48000,
				ChannelCount: 2,
				AU:           []byte{0xaa, 0xbb},
			},
		},
	},
}

func TestADTSUnmarshal(t *testing.T) {
	// Test roundtrip cases
	for _, ca := range casesADTS {
		t.Run(ca.name, func(t *testing.T) {
			var pkts ADTSPackets
			err := pkts.Unmarshal(ca.byts)
			require.NoError(t, err)
			require.Equal(t, ca.pkts, pkts)
		})
	}
	// Test unmarshal-only cases (profile normalization)
	for _, ca := range casesADTSUnmarshalOnly {
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

	f.Fuzz(func(t *testing.T, b []byte) {
		var pkts ADTSPackets
		err := pkts.Unmarshal(b)
		if err != nil {
			return
		}

		_, err = pkts.Marshal()
		require.NoError(t, err)
	})
}
