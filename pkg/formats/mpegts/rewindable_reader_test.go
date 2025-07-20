package mpegts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type dummyReader2 struct {
	i int
}

func (r *dummyReader2) Read(buf []byte) (int, error) {
	r.i++
	switch r.i {
	case 1:
		copy(buf, []byte{1, 2, 3, 4})
		return 4, nil

	case 2:
		copy(buf, []byte{5, 6})
		return 2, nil

	case 3:
		copy(buf, []byte{7, 8})
		return 2, nil
	}
	panic("should not happen")
}

func TestRewindableReader(t *testing.T) {
	r := &rewindableReader{R: &dummyReader2{}}

	for i := range 2 {
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		require.NoError(t, err)
		require.Equal(t, []byte{1, 2, 3, 4}, buf[:n])

		n, err = r.Read(buf)
		require.NoError(t, err)
		require.Equal(t, []byte{5, 6}, buf[:n])

		if i == 0 {
			r.Rewind()
		} else {
			n, err = r.Read(buf)
			require.NoError(t, err)
			require.Equal(t, []byte{7, 8}, buf[:n])
		}
	}
}

func TestRewindableReaderDifferentBufSize(t *testing.T) {
	r := &rewindableReader{R: &dummyReader2{}}

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3, 4}, buf[:n])

	r.Rewind()

	buf = make([]byte, 2)

	n, err = r.Read(buf)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2}, buf[:n])

	n, err = r.Read(buf)
	require.NoError(t, err)
	require.Equal(t, []byte{3, 4}, buf[:n])
}
