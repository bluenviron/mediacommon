package h265_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
)

func TestIsRandomAccess(t *testing.T) {
	u := [][]byte{{byte(h265.NALUType_IDR_W_RADL) << 1}}
	require.Equal(t, true, h265.IsRandomAccess(u))

	u = [][]byte{{byte(h265.NALUType_TRAIL_N) << 1}}
	require.Equal(t, false, h265.IsRandomAccess(u))
}
