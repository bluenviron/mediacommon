package vp9_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/vp9"
)

func TestIsRandomAccess(t *testing.T) {
	require.Equal(t, true, vp9.IsRandomAccess([]byte{
		0x82, 0x49, 0x83, 0x42, 0x00, 0x77, 0xf0, 0x32,
		0x34, 0x30, 0x38, 0x24, 0x1c, 0x19, 0x40, 0x18,
		0x03, 0x40, 0x5f, 0xb4,
	}))

	require.Equal(t, false, vp9.IsRandomAccess([]byte{0x86}))

	require.Equal(t, false, vp9.IsRandomAccess([]byte{}))
}
