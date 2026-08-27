package h264_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
)

func TestNALUType(t *testing.T) {
	require.NotEqual(t, true, strings.HasPrefix(h264.NALUType(10).String(), "unknown"))
	require.Equal(t, true, strings.HasPrefix(h264.NALUType(50).String(), "unknown"))
}
