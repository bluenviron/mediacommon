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
	for _, ca := range ac3Cases {
		f.Add(ca.enc[5:])
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var bsi BSI
		err := bsi.Unmarshal(b)
		if err == nil {
			bsi.ChannelCount() //nolint:staticcheck
		}
	})
}
