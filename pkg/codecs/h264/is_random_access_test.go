package h264_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
)

func TestIsRandomAccess(t *testing.T) {
	require.Equal(t, true, h264.IsRandomAccess([][]byte{
		{0x05},
		{0x07},
	}))
	require.Equal(t, false, h264.IsRandomAccess([][]byte{
		{0x01},
	}))
}
