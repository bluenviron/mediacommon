package g711

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlawUnmarshal(t *testing.T) {
	var dec Alaw
	dec.Unmarshal([]byte{1, 2, 3, 255, 254, 253})

	require.Equal(t,
		dec,
		Alaw{
			0xeb, 0x80, 0xe8, 0x80, 0xe9, 0x80, 0x03, 0x50,
			0x03, 0x70, 0x03, 0x10,
		},
	)
}

func TestAlawMarshal(t *testing.T) {
	in := []byte{1, 2, 3, 4, 5, 6}
	enc, err := Alaw(in).Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0xc5, 0xfd, 0xe1}, enc)
}
