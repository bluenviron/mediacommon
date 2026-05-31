package vp8

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRandomAccess(t *testing.T) {
	require.Equal(t, true, IsRandomAccess([]byte{0x00, 0x01, 0x02}))

	require.Equal(t, false, IsRandomAccess([]byte{0x01, 0x00, 0x00}))

	require.Equal(t, false, IsRandomAccess([]byte{}))
}
