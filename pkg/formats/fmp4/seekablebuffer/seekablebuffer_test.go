package seekablebuffer

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuffer(t *testing.T) {
	t.Run("seek start", func(t *testing.T) {
		var b Buffer

		_, err := b.Seek(-2, io.SeekStart)
		require.Error(t, err)

		n1, err := b.Seek(0, io.SeekStart)
		require.NoError(t, err)
		require.Equal(t, int64(0), n1)

		n2, err := b.Write([]byte{1, 2, 3, 4})
		require.NoError(t, err)
		require.Equal(t, 4, n2)
	})

	t.Run("seek current", func(t *testing.T) {
		var b Buffer

		n2, err := b.Write([]byte{1, 2, 3, 4})
		require.NoError(t, err)
		require.Equal(t, 4, n2)

		n1, err := b.Seek(-2, io.SeekCurrent)
		require.NoError(t, err)
		require.Equal(t, int64(2), n1)

		n2, err = b.Write([]byte{1, 2, 3, 4})
		require.NoError(t, err)
		require.Equal(t, 4, n2)

		require.Equal(t, []byte{1, 2, 1, 2, 3, 4}, b.Bytes())
	})

	t.Run("seek end", func(t *testing.T) {
		var b Buffer

		n2, err := b.Write([]byte{1, 2, 3, 4})
		require.NoError(t, err)
		require.Equal(t, 4, n2)

		n1, err := b.Seek(2, io.SeekEnd)
		require.NoError(t, err)
		require.Equal(t, int64(6), n1)

		n2, err = b.Write([]byte{1, 2, 3, 4})
		require.NoError(t, err)
		require.Equal(t, 4, n2)

		require.Equal(t, []byte{1, 2, 3, 4, 0, 0, 1, 2, 3, 4}, b.Bytes())
	})
}
