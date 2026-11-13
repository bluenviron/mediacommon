package limitedreader_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/limitedreader"
)

func TestReader(t *testing.T) {
	r := &limitedreader.Reader{R: bytes.NewReader([]byte("0123456789")), N: 11}
	buf, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, []byte("0123456789"), buf)
}

func TestReaderExceeded(t *testing.T) {
	r := &limitedreader.Reader{R: bytes.NewReader([]byte("0123456789")), N: 5}
	buf, err := io.ReadAll(r)
	require.Equal(t, limitedreader.ErrSizeExceeded, err)
	require.Equal(t, []byte("01234"), buf)
}

func TestReaderChangeLimit(t *testing.T) {
	r := &limitedreader.Reader{R: bytes.NewReader([]byte("0123456789")), N: 5}
	buf := make([]byte, 5)
	_, err := io.ReadFull(r, buf)
	require.NoError(t, err)

	_, err = r.Read(buf)
	require.Equal(t, limitedreader.ErrSizeExceeded, err)

	r.N = 5
	_, err = io.ReadFull(r, buf)
	require.NoError(t, err)
	require.Equal(t, []byte("56789"), buf)
}
