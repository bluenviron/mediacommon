package bits_test //nolint:revive

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

func TestWriteBitsUnsafe(t *testing.T) {
	buf := make([]byte, 6)
	pos := 0

	bits.WriteBitsUnsafe(buf, &pos, uint64(0x2a), 6)
	bits.WriteBitsUnsafe(buf, &pos, uint64(0x0c), 6)
	bits.WriteBitsUnsafe(buf, &pos, uint64(0x1f), 6)
	bits.WriteBitsUnsafe(buf, &pos, uint64(0x5a), 8)
	bits.WriteBitsUnsafe(buf, &pos, uint64(0xaaec4), 20)

	require.Equal(t, []byte{0xA8, 0xC7, 0xD6, 0xAA, 0xBB, 0x10}, buf)
}

func TestWriteFlagUnsafe(t *testing.T) {
	buf := make([]byte, 1)
	pos := 0

	bits.WriteFlagUnsafe(buf, &pos, true)
	bits.WriteFlagUnsafe(buf, &pos, false)

	require.Equal(t, 2, pos)
	require.Equal(t, []byte{0x80}, buf)
}
