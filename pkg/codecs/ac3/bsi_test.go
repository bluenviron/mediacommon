package ac3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBSIUnmarshal(t *testing.T) {
	for _, ca := range ac3Cases {
		t.Run(ca.name, func(t *testing.T) {
			var bsi BSI
			err := bsi.Unmarshal(ca.enc[5:])
			require.NoError(t, err)
			require.Equal(t, ca.bsi, bsi)
		})
	}
}

func FuzzBSIUnmarshal(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		var bsi BSI
		bsi.Unmarshal(b) //nolint:errcheck
	})
}
