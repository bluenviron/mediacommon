package ac3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/ac3"
)

func TestBSIUnmarshal(t *testing.T) {
	for _, ca := range ac3Cases {
		t.Run(ca.name, func(t *testing.T) {
			var bsi ac3.BSI
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
		var bsi ac3.BSI
		err := bsi.Unmarshal(b)
		if err != nil {
			return
		}

		bsi.ChannelCount() //nolint:staticcheck
	})
}
