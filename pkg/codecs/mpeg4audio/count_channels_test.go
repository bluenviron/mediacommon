package mpeg4audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var countChannelsCases = []struct {
	name  string
	buf   []byte
	count int
}{
	{
		name: "CPE stereo",
		// id_syn_ele (3 bits) = 001 (CPE)
		// element_instance_tag (4 bits) = 0000
		// Byte: 001 0000 0 = 0x20
		buf:   []byte{0x20, 0x00, 0x00, 0x00},
		count: 2,
	},
	{
		name: "SCE mono",
		// id_syn_ele (3 bits) = 000 (SCE)
		// element_instance_tag (4 bits) = 0000
		// Byte: 000 0000 0 = 0x00
		buf:   []byte{0x00, 0x00, 0x00, 0x00},
		count: 1,
	},
	{
		name: "LFE",
		// id_syn_ele (3 bits) = 011 (LFE)
		// element_instance_tag (4 bits) = 0000
		// Byte: 011 0000 0 = 0x60
		buf:   []byte{0x60, 0x00, 0x00, 0x00},
		count: 1,
	},
	{
		name: "FIL then CPE",
		// FIL element: id_syn_ele (3 bits) = 110 (FIL=6)
		// count (4 bits) = 0000 (no fill bytes)
		// Then CPE: id_syn_ele (3 bits) = 001, tag (4 bits) = 0000
		// FIL takes 7 bits (3+4), CPE starts at bit 7
		// Bit 0-2: 110 (FIL)
		// Bit 3-6: 0000 (count=0)
		// Bit 7-9: 001 (CPE)
		// Bit 10-13: 0000 (tag)
		// Byte 0: 1100_0000 = 0xC0
		// Byte 1: 0100_0000 = 0x40
		buf:   []byte{0xC0, 0x40, 0x00, 0x00},
		count: 2,
	},
	{
		name: "DSE then CPE",
		// DSE element: id_syn_ele (3 bits) = 100 (DSE=4)
		// element_instance_tag (4 bits) = 0000
		// data_byte_align_flag (1 bit) = 0
		// count (8 bits) = 00000000 (no data bytes)
		// Then CPE: id_syn_ele (3 bits) = 001, tag (4 bits) = 0000
		// DSE takes 16 bits (3+4+1+8), CPE starts at bit 16
		// Byte 0: 100 0000 0 = 0x80 (DSE, tag=0, align=0 start)
		// Byte 1: 0000 0000 = 0x00 (count=0)
		// Byte 2: 001 0000 0 = 0x20 (CPE)
		buf:   []byte{0x80, 0x00, 0x20, 0x00},
		count: 2,
	},
	{
		name: "DSE aligned then CPE",
		// DSE: id (3) = 100, tag (4) = 0000, align (1) = 1
		// Byte 0: 100 0000 1 = 0x81
		// Then count (8 bits) = 0x00 (no data)
		// Byte 1: 0000 0000 = 0x00
		// Byte align: at bit 16, which is already byte-aligned
		// CPE: id (3) = 001, tag (4) = 0000
		// Byte 2: 001 0000 0 = 0x20
		buf:   []byte{0x81, 0x00, 0x20, 0x00},
		count: 2,
	},
	{
		name: "DSE with count=1 then CPE",
		// DSE: id (3) = 100, tag (4) = 0000, align (1) = 0
		// Byte 0: 100 0000 0 = 0x80 (DSE, tag=0, align start)
		// Bits 0-2: 100 = DSE
		// Bits 3-6: 0000 = tag
		// Bit 7: 0 = align flag
		// Bits 8-15: count
		// Byte 1: count = 1
		// After count, we have 1 data byte
		// DSE takes 3+4+1+8 = 16 bits header, then count*8 = 8 bits data
		// Total DSE = 24 bits = 3 bytes
		// CPE starts at byte 3
		// Byte 3: 001 0000 0 = 0x20
		buf:   []byte{0x80, 0x01, 0x00, 0x20, 0x00, 0x00},
		count: 2,
	},
	{
		name: "FIL with count < 15",
		// Same as FIL then CPE test (count=0)
		buf:   []byte{0xC0, 0x40, 0x00, 0x00},
		count: 2,
	},
}

func TestCountChannelsFromRawDataBlock(t *testing.T) {
	for _, tt := range countChannelsCases {
		t.Run(tt.name, func(t *testing.T) {
			channels, err := CountChannelsFromRawDataBlock(tt.buf)
			require.NoError(t, err)
			require.Equal(t, tt.count, channels)
		})
	}
}

func FuzzCountChannelsFromRawDataBlock(f *testing.F) {
	for _, ca := range countChannelsCases {
		f.Add(ca.buf)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		count, err := CountChannelsFromRawDataBlock(b)
		if err == nil {
			require.NotZero(t, count)
		}
	})
}
