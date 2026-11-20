package jpeg_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/jpeg"
)

func TestStartOfImageMarshal(t *testing.T) {
	buf := jpeg.StartOfImage{}.Marshal(nil)
	require.Equal(t, []byte{0xff, 0xd8}, buf)
}
