package substructs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var audioDescriptorCases = []struct {
	name         string
	enc          []byte
	desc         *OpusAudioDescriptor
	channelCount int
}{
	// pre-defined channel_config_codes (Table 4-3)
	{
		"dual mono (0x00)",
		[]byte{0x00},
		&OpusAudioDescriptor{
			ChannelConfigCode: 0x00,
		},
		2,
	},
	{
		"mono (0x01)",
		[]byte{0x01},
		&OpusAudioDescriptor{
			ChannelConfigCode: 1,
		},
		1,
	},
	{
		"stereo (0x02)",
		[]byte{0x02},
		&OpusAudioDescriptor{
			ChannelConfigCode: 2,
		},
		2,
	},
	{
		"3ch surround (0x03)",
		[]byte{0x03},
		&OpusAudioDescriptor{
			ChannelConfigCode: 3,
		},
		3,
	},
	{
		"quad (0x04)",
		[]byte{0x04},
		&OpusAudioDescriptor{
			ChannelConfigCode: 4,
		},
		4,
	},
	{
		"5ch (0x05)",
		[]byte{0x05},
		&OpusAudioDescriptor{
			ChannelConfigCode: 5,
		},
		5,
	},
	{
		"5.1 (0x06)",
		[]byte{0x06},
		&OpusAudioDescriptor{
			ChannelConfigCode: 6,
		},
		6,
	},
	{
		"6.1 (0x07)",
		[]byte{0x07},
		&OpusAudioDescriptor{
			ChannelConfigCode: 7,
		},
		7,
	},
	{
		"7.1 (0x08)",
		[]byte{0x08},
		&OpusAudioDescriptor{
			ChannelConfigCode: 8,
		},
		8,
	},
	{
		"dual mono alias (0x80)",
		[]byte{0x80},
		&OpusAudioDescriptor{
			ChannelConfigCode: 0x80,
		},
		2,
	},
	{
		"0x82 - 1ch mapping_family 1",
		[]byte{0x82},
		&OpusAudioDescriptor{
			ChannelConfigCode: 0x82,
		},
		1,
	},
	{
		"0x88 - 7ch mapping_family 1",
		[]byte{0x88},
		&OpusAudioDescriptor{
			ChannelConfigCode: 0x88,
		},
		7,
	},
	//  explicit configuration (channel_config_code == 0x81)
	{
		// mapping_family=0, mono: no bit-packed fields.
		"explicit mono mapping_family 0",
		[]byte{0x81, 0x01, 0x00},
		&OpusAudioDescriptor{
			ChannelConfigCode:    0x81,
			ExplicitChannelCount: 1,
		},
		1,
	},
	{
		// mapping_family=0, stereo: no bit-packed fields.
		"explicit stereo mapping_family 0",
		[]byte{0x81, 0x02, 0x00},
		&OpusAudioDescriptor{
			ChannelConfigCode:    0x81,
			ExplicitChannelCount: 2,
		},
		2,
	},
	{
		// mapping_family=1, 1ch:
		//   bitsForStreamCountMinus1 = ceil(log2(1)) = 0 → stream_count=1
		//   bitsForCoupled           = ceil(log2(2)) = 1 → coupled=0  (bit: 0)
		//   bitsPerMapping           = ceil(log2(2)) = 1 → mapping[0]=0 (bit: 0)
		//   total: 2 bits → padded to 1 byte → 0x00
		"explicit 1ch mapping_family 1",
		[]byte{0x81, 0x01, 0x01, 0x00},
		&OpusAudioDescriptor{
			ChannelConfigCode:    0x81,
			ExplicitChannelCount: 1,
			MappingFamily:        1,
			StreamCount:          1,
			CoupledStreamCount:   0,
			ChannelMapping:       []byte{0},
		},
		1,
	},
	{
		// mapping_family=1, 6ch, 5.1 layout: stream_count=4, coupled=2, mapping={0,4,1,2,3,5}
		//   bitsForStreamCountMinus1 = ceil(log2(6)) = 3 → encode 3 → 011
		//   bitsForCoupled           = ceil(log2(5)) = 3 → encode 2 → 010
		//   bitsPerMapping           = ceil(log2(7)) = 3
		//   mapping bytes: 011 010 | 000 100 0 | 01 010 01 | 1 101 000(pad)
		//   → 0x68, 0x42, 0x9D
		"explicit 5.1 mapping_family 1",
		[]byte{0x81, 0x06, 0x01, 0x68, 0x42, 0x9D},
		&OpusAudioDescriptor{
			ChannelConfigCode:    0x81,
			ExplicitChannelCount: 6,
			MappingFamily:        1,
			StreamCount:          4,
			CoupledStreamCount:   2,
			ChannelMapping:       []byte{0, 4, 1, 2, 3, 5},
		},
		6,
	},
	{
		// mapping_family=255, 3ch application-defined: stream_count=3, coupled=0, mapping={0,1,2}
		//   bitsForStreamCountMinus1 = ceil(log2(3)) = 2 → encode 2 → 10
		//   bitsForCoupled           = ceil(log2(4)) = 2 → encode 0 → 00
		//   bitsPerMapping           = ceil(log2(4)) = 2
		//   mapping: 00 01 10 → packed: 1000 0110 (pad 6 bits) → 0x81 0x80
		"explicit 3ch mapping_family 255",
		[]byte{0x81, 0x03, 0xFF, 0x81, 0x80},
		&OpusAudioDescriptor{
			ChannelConfigCode:    0x81,
			ExplicitChannelCount: 3,
			MappingFamily:        255,
			StreamCount:          3,
			CoupledStreamCount:   0,
			ChannelMapping:       []byte{0, 1, 2},
		},
		3,
	},
}

func TestOpusAudioDescriptorUnmarshal(t *testing.T) {
	for _, ca := range audioDescriptorCases {
		t.Run(ca.name, func(t *testing.T) {
			desc := &OpusAudioDescriptor{}
			err := desc.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.desc, desc)
			require.Equal(t, ca.channelCount, desc.ChannelCount())
		})
	}
}

func TestOpusAudioDescriptorMarshal(t *testing.T) {
	for _, ca := range audioDescriptorCases {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.desc.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func FuzzOpusAudioDescriptorUnmarshal(f *testing.F) {
	for _, ca := range audioDescriptorCases {
		f.Add(ca.enc)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		var h OpusAudioDescriptor
		err := h.Unmarshal(b)
		if err != nil {
			return
		}

		_, err = h.Marshal()
		require.NoError(t, err)
	})
}
