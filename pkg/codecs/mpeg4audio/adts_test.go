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
	{
		"aac-ssr",
		[]byte{0xff, 0xf1, 0x8c, 0x80, 0x1, 0x3f, 0xfc, 0xaa, 0xbb},
		ADTSPackets{
			{
				Type:         3,
				SampleRate:   48000,
				ChannelCount: 2,
				AU:           []byte{0xaa, 0xbb},
			},
		},
	},
	{
		// channel_config=0 with CPE (stereo pair) in AU
		// ADTS header with channel_config=0, frame_length=11
		// AU contains: CPE id_syn_ele (3 bits) = 001, element_instance_tag (4 bits) = 0000
		"channel_config_0_cpe",
		[]byte{
			0xff, 0xf1, 0x4c, 0x00, 0x01, 0x7f, 0xfc,
			0x20, 0x00, 0x00, 0x00, // AU: CPE element (stereo)
		},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC,
				SampleRate:   48000,
				ChannelCount: 0, // Preserved as 0, use CountChannelsFromRawDataBlock to get 2
				AU:           []byte{0x20, 0x00, 0x00, 0x00},
			},
		},
	},
	{
		// channel_config=0 with SCE (mono) in AU
		// AU contains: SCE id_syn_ele (3 bits) = 000, element_instance_tag (4 bits) = 0000
		"channel_config_0_sce",
		[]byte{
			0xff, 0xf1, 0x4c, 0x00, 0x01, 0x7f, 0xfc,
			0x00, 0x00, 0x00, 0x00, // AU: SCE element (mono)
		},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC,
				SampleRate:   48000,
				ChannelCount: 0, // Preserved as 0, use CountChannelsFromRawDataBlock to get 1
				AU:           []byte{0x00, 0x00, 0x00, 0x00},
			},
		},
	},
	{
		// channel_config=0 with LFE in AU
		// AU contains: LFE id_syn_ele (3 bits) = 011, element_instance_tag (4 bits) = 0000
		"channel_config_0_lfe",
		[]byte{
			0xff, 0xf1, 0x4c, 0x00, 0x01, 0x7f, 0xfc,
			0x60, 0x00, 0x00, 0x00, // AU: LFE element
		},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC,
				SampleRate:   48000,
				ChannelCount: 0, // Preserved as 0, use CountChannelsFromRawDataBlock to get 1
				AU:           []byte{0x60, 0x00, 0x00, 0x00},
			},
		},
	},
	{
		// channel_config=0 with PCE (stereo) in AU
		// AU contains: PCE id_syn_ele (3 bits) = 101, then PCE data
		// PCE with 1 front CPE = 2 channels
		"channel_config_0_pce",
		[]byte{
			0xff, 0xf1, 0x4c, 0x00, 0x01, 0xbf, 0xfc,
			0xA0, 0xA0, 0x80, 0x00, 0x04, 0x00, // AU: PCE element (stereo)
		},
		ADTSPackets{
			{
				Type:         ObjectTypeAACLC,
				SampleRate:   48000,
				ChannelCount: 0, // Preserved as 0, use ParsePCEFromRawDataBlock to get 2
				AU:           []byte{0xA0, 0xA0, 0x80, 0x00, 0x04, 0x00},
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
