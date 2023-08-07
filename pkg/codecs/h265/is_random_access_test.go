package h265

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRandomAccess(t *testing.T) {
	u := [][]byte{{byte(NALUType_IDR_W_RADL) << 1}}
	require.Equal(t, true, IsRandomAccess(u))

	u = [][]byte{{byte(NALUType_TRAIL_N) << 1}}
	require.Equal(t, false, IsRandomAccess(u))
}
