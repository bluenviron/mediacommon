package mpegts

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBufferedReader(t *testing.T) {
	var buf bytes.Buffer

	buf.Write(bytes.Repeat([]byte{1}, 188))
	buf.Write(bytes.Repeat([]byte{2}, 188))
	buf.Write(bytes.Repeat([]byte{3}, 188))

	r := NewBufferedReader(&buf)

	byts := make([]byte, 188)
	n, err := r.Read(byts)
	require.NoError(t, err)
	require.Equal(t, 188, n)
	require.Equal(t, bytes.Repeat([]byte{1}, 188), byts[:n])

	require.Equal(t, 0, len(buf.Bytes()))

	byts = make([]byte, 188)
	n, err = r.Read(byts)
	require.NoError(t, err)
	require.Equal(t, 188, n)
	require.Equal(t, bytes.Repeat([]byte{2}, 188), byts[:n])
}

func TestBufferedReaderError(t *testing.T) {
	var buf bytes.Buffer

	buf.Write(bytes.Repeat([]byte{1}, 1000))

	r := NewBufferedReader(&buf)
	byts := make([]byte, 188)
	_, err := r.Read(byts)
	require.EqualError(t, err, "received packet with size 1000 not multiple of 188")
}
