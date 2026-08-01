package g711_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/g711"
)

func TestAlawUnmarshal(t *testing.T) {
	var dec g711.Alaw
	dec.Unmarshal([]byte{1, 2, 3, 255, 254, 253})

	require.Equal(t,
		dec,
		g711.Alaw{
			0xeb, 0x80, 0xe8, 0x80, 0xe9, 0x80, 0x03, 0x50,
			0x03, 0x70, 0x03, 0x10,
		},
	)
}

func TestAlawMarshal(t *testing.T) {
	in := []byte{1, 2, 3, 4, 5, 6}
	enc, err := g711.Alaw(in).Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0xc5, 0xfd, 0xe1}, enc)
}
