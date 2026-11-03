package scte35

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// splice_null(), base64 "/DARAAAAAAAAAP/wAAAAAHpPv/8="
var testSpliceNull = []byte{
	0xfc, 0x30, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xff, 0xf0, 0x00, 0x00, 0x00, 0x00,
	0x7a, 0x4f, 0xbf, 0xff,
}

func TestPTSAdjustment(t *testing.T) {
	v, err := PTSAdjustment(testSpliceNull)
	require.NoError(t, err)
	require.Equal(t, int64(0), v)
}

func TestWithPTSAdjustment(t *testing.T) {
	// recomputing pts_adjustment of a valid section produces the same section
	sec, err := WithPTSAdjustment(testSpliceNull, 0)
	require.NoError(t, err)
	require.Equal(t, testSpliceNull, sec)

	// 33-bit value, highest bit set
	sec, err = WithPTSAdjustment(testSpliceNull, 0x1_2345_6789)
	require.NoError(t, err)
	require.Equal(t, uint32(0), crc32MPEG2(sec))
	v, err := PTSAdjustment(sec)
	require.NoError(t, err)
	require.Equal(t, int64(0x1_2345_6789), v)

	// original is untouched
	v, err = PTSAdjustment(testSpliceNull)
	require.NoError(t, err)
	require.Equal(t, int64(0), v)

	// other bits of byte 4 are preserved
	enc := append([]byte{}, testSpliceNull...)
	enc[4] = 0xFE // encrypted_packet + encryption_algorithm
	sec, err = WithPTSAdjustment(enc, 1<<32)
	require.NoError(t, err)
	require.Equal(t, uint32(0), crc32MPEG2(sec))
	require.Equal(t, byte(0xFF), sec[4])
	v, err = PTSAdjustment(sec)
	require.NoError(t, err)
	require.Equal(t, int64(1<<32), v)

	// negative values wrap modulo 2^33
	sec, err = WithPTSAdjustment(testSpliceNull, -1)
	require.NoError(t, err)
	v, err = PTSAdjustment(sec)
	require.NoError(t, err)
	require.Equal(t, int64(0x1FFFFFFFF), v)
}

func TestPTSAdjustmentErrors(t *testing.T) {
	for _, ca := range []struct {
		name string
		sec  []byte
		err  string
	}{
		{
			"too short",
			testSpliceNull[:10],
			"section is too short",
		},
		{
			"table id",
			append([]byte{0xfd}, testSpliceNull[1:]...),
			"invalid table_id: 0xfd",
		},
		{
			"section length",
			append(append([]byte{}, testSpliceNull...), 0x00),
			"invalid section_length: 17, expected 18",
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			_, err := PTSAdjustment(ca.sec)
			require.EqualError(t, err, ca.err)

			_, err = WithPTSAdjustment(ca.sec, 0)
			require.EqualError(t, err, ca.err)
		})
	}
}
