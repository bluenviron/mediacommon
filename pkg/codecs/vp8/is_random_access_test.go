package vp8_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/vp8"
)

func TestIsRandomAccess(t *testing.T) {
	require.Equal(t, true, vp8.IsRandomAccess([]byte{0x00, 0x01, 0x02}))

	require.Equal(t, false, vp8.IsRandomAccess([]byte{0x01, 0x00, 0x00}))

	require.Equal(t, false, vp8.IsRandomAccess([]byte{}))
}
