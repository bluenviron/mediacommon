package h265_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
)

var casesPPS = []struct {
	name string
	byts []byte
	pps  h265.PPS
}{
	{
		"default",
		[]byte{
			0x44, 0x01, 0xc1, 0x72, 0xb4, 0x62, 0x40,
		},
		h265.PPS{},
	},
}

func TestPPSUnmarshal(t *testing.T) {
	for _, ca := range casesPPS {
		t.Run(ca.name, func(t *testing.T) {
			var pps h265.PPS
			err := pps.Unmarshal(ca.byts)
			require.NoError(t, err)
			require.Equal(t, ca.pps, pps)
		})
	}
}

func FuzzPPSUnmarshal(f *testing.F) {
	for _, ca := range casesPPS {
		f.Add(ca.byts)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var pps h265.PPS
		pps.Unmarshal(b) //nolint:errcheck
	})
}
