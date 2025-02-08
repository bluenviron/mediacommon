package av1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRandomAccess(t *testing.T) {
	ok, err := IsRandomAccess([][]byte{{
		0x0a, 0x0e, 0x00, 0x00, 0x00, 0x4a, 0xab, 0xbf,
		0xc3, 0x77, 0x6b, 0xe4, 0x40, 0x40, 0x40, 0x41,
	}})
	require.NoError(t, err)
	require.Equal(t, true, ok)

	_, err = IsRandomAccess([][]byte{})
	require.Error(t, err)

	_, err = IsRandomAccess([][]byte{{}})
	require.Error(t, err)
}
