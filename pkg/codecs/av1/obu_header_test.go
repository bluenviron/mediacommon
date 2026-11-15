package av1_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
)

var casesOBUHeader = []struct {
	name string
	byts []byte
	h    av1.OBUHeader
}{
	{
		"sequence header",
		[]byte{
			0x0a, 0x0e, 0x00, 0x00, 0x00, 0x4a, 0xab, 0xbf,
			0xc3, 0x77, 0x6b, 0xe4, 0x40, 0x40, 0x40, 0x41,
		},
		av1.OBUHeader{
			Type:    av1.OBUTypeSequenceHeader,
			HasSize: true,
		},
	},
}

func TestOBUHeaderUnmarshal(t *testing.T) {
	for _, ca := range casesOBUHeader {
		t.Run(ca.name, func(t *testing.T) {
			var h av1.OBUHeader
			err := h.Unmarshal(ca.byts)
			require.NoError(t, err)
			require.Equal(t, ca.h, h)
		})
	}
}

func FuzzOBUHeaderUnmarshal(f *testing.F) {
	for _, ca := range casesOBUHeader {
		f.Add(ca.byts)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var h av1.OBUHeader
		h.Unmarshal(b) //nolint:errcheck
	})
}
