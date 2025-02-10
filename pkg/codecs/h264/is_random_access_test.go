package h264

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRandomAccess(t *testing.T) {
	require.Equal(t, true, IsRandomAccess([][]byte{
		{0x05},
		{0x07},
	}))
	require.Equal(t, false, IsRandomAccess([][]byte{
		{0x01},
	}))
}
