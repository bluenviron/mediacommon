package scte35

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// splice_null(), base64 "/DARAAAAAAAAAP/wAAAAAHpPv/8="
var testSpliceNull = Section{
	0xfc, 0x30, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xff, 0xf0, 0x00, 0x00, 0x00, 0x00,
	0x7a, 0x4f, 0xbf, 0xff,
}

func TestSectionValidate(t *testing.T) {
	require.NoError(t, testSpliceNull.Validate())
}

func TestSectionValidateErrors(t *testing.T) {
	for _, ca := range []struct {
		name string
		sec  Section
		err  string
	}{
		{
			"too short",
			testSpliceNull[:10],
			"section is too short",
		},
		{
			"table id",
			append(Section{0xfd}, testSpliceNull[1:]...),
			"invalid table_id: 0xfd",
		},
		{
			"section length",
			append(append(Section{}, testSpliceNull...), 0x00),
			"invalid section_length: 17, expected 18",
		},
		{
			"crc",
			append(append(Section{}, testSpliceNull[:19]...), 0x00),
			"CRC_32 mismatch",
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			require.EqualError(t, ca.sec.Validate(), ca.err)
		})
	}
}

func TestSectionPTSAdjustment(t *testing.T) {
	require.Equal(t, int64(0), testSpliceNull.PTSAdjustment())

	// 33-bit value, highest bit set
	sec := testSpliceNull.WithPTSAdjustment(0x1_2345_6789)
	require.NoError(t, sec.Validate())
	require.Equal(t, int64(0x1_2345_6789), sec.PTSAdjustment())

	// original is untouched
	require.Equal(t, int64(0), testSpliceNull.PTSAdjustment())

	// other bits of byte 4 are preserved
	enc := append(Section{}, testSpliceNull...)
	enc[4] = 0xFE // encrypted_packet + encryption_algorithm
	enc = enc.WithPTSAdjustment(0)
	sec = enc.WithPTSAdjustment(1 << 32)
	require.NoError(t, sec.Validate())
	require.Equal(t, byte(0xFF), sec[4])
	require.Equal(t, int64(1<<32), sec.PTSAdjustment())

	// negative values wrap modulo 2^33
	sec = testSpliceNull.WithPTSAdjustment(-1)
	require.NoError(t, sec.Validate())
	require.Equal(t, int64(0x1FFFFFFFF), sec.PTSAdjustment())
}
