package g711

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMulawUnmarshal(t *testing.T) {
	var dec Mulaw
	dec.Unmarshal([]byte{1, 2, 3, 255, 254, 253})

	require.Equal(t,
		dec,
		Mulaw{
			0x86, 0x84, 0x8a, 0x84, 0x8e, 0x84, 0x00, 0x00,
			0x00, 0x08, 0x00, 0x10,
		},
	)
}

func TestMulawMarshal(t *testing.T) {
	in := []byte{1, 2, 3, 4, 5, 6}
	enc, err := Mulaw(in).Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0xe7, 0xd3, 0xc9}, enc)
}
