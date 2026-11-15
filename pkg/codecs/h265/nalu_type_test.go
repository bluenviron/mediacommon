package h265_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
)

func TestNALUType(t *testing.T) {
	require.NotEqual(t, true, strings.HasPrefix(h265.NALUType(10).String(), "unknown"))
	require.Equal(t, true, strings.HasPrefix(h265.NALUType(60).String(), "unknown"))
}
