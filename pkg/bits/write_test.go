package bits

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteBitsUnsafe(t *testing.T) {
	buf := make([]byte, 6)
	pos := 0
	WriteBitsUnsafe(buf, &pos, uint64(0x2a), 6)
	WriteBitsUnsafe(buf, &pos, uint64(0x0c), 6)
	WriteBitsUnsafe(buf, &pos, uint64(0x1f), 6)
	WriteBitsUnsafe(buf, &pos, uint64(0x5a), 8)
	WriteBitsUnsafe(buf, &pos, uint64(0xaaec4), 20)
	require.Equal(t, []byte{0xA8, 0xC7, 0xD6, 0xAA, 0xBB, 0x10}, buf)
}

func TestWriteFlagUnsafe(t *testing.T) {
	buf := []byte{0}
	pos := 2
	WriteFlagUnsafe(buf, &pos, true)
	require.Equal(t, []byte{0b00100000}, buf)
	require.Equal(t, 3, pos)

	buf = []byte{0}
	pos = 2
	WriteFlagUnsafe(buf, &pos, false)
	require.Equal(t, []byte{0}, buf)
	require.Equal(t, 3, pos)
}

func TestWriteGolombUnsignedUnsafe(t *testing.T) {
	for _, ca := range casesGolomb {
		buf := []byte{0}
		pos := 0
		WriteGolombUnsignedUnsafe(buf, &pos, ca.dec)
		require.Equal(t, ca.enc, buf)
		require.Equal(t, ca.bitSize, pos)
	}
}
