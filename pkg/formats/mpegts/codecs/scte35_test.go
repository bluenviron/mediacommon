package codecs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
)

func TestSCTE35(t *testing.T) {
	require.False(t, (&codecs.SCTE35{}).IsVideo())
}
