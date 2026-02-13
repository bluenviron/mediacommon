package bits //nolint:revive

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
	buf := make([]byte, 1)
	pos := 0

	WriteFlagUnsafe(buf, &pos, true)
	WriteFlagUnsafe(buf, &pos, false)

	require.Equal(t, 2, pos)
	require.Equal(t, []byte{0x80}, buf)
}
