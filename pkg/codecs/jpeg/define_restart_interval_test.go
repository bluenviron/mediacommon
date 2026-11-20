package jpeg_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/jpeg"
)

var casesDefineRestartInterval = []struct {
	name string
	enc  []byte
	dec  jpeg.DefineRestartInterval
}{
	{
		"base",
		[]byte{
			0xff, jpeg.MarkerDefineRestartInterval, 0x00, 0x04, 0xd0, 0xc7,
		},
		jpeg.DefineRestartInterval{
			Interval: 53447,
		},
	},
}

func TestDefineRestartIntervalUnmarshal(t *testing.T) {
	for _, ca := range casesDefineRestartInterval {
		t.Run(ca.name, func(t *testing.T) {
			var h jpeg.DefineRestartInterval
			err := h.Unmarshal(ca.enc[4:])
			require.NoError(t, err)
			require.Equal(t, ca.dec, h)
		})
	}
}

func FuzzDefineRestartIntervalUnmarshal(f *testing.F) {
	for _, ca := range casesDefineRestartInterval {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var h jpeg.DefineRestartInterval
		h.Unmarshal(b) //nolint:errcheck
	})
}
