package h264_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
)

func TestIsRandomAccess(t *testing.T) {
	cases := []struct {
		name string
		au   [][]byte
		out  bool
	}{
		{
			name: "IDR",
			au:   [][]byte{{0x05}, {0x07}},
			out:  true,
		},
		{
			name: "non-IDR",
			au:   [][]byte{{0x01}},
		},
		{
			name: "recovery point before non-IDR",
			au:   [][]byte{{0x67}, {0x68}, {0x06, 0x06, 0x01, 0xc4, 0x80}, {0x41}},
			out:  true,
		},
		{
			name: "recovery point alone",
			au:   [][]byte{{0x06, 0x06, 0x01, 0xc4, 0x80}},
		},
		{
			name: "recovery point after non-IDR",
			au:   [][]byte{{0x41}, {0x06, 0x06, 0x01, 0xc4, 0x80}},
		},
		{
			name: "nonzero recovery count",
			au:   [][]byte{{0x06, 0x06, 0x01, 0x51, 0x80}, {0x41}},
		},
		{
			name: "malformed recovery point",
			au:   [][]byte{{0x06, 0x06, 0x01, 0x00, 0x80}, {0x41}},
		},
	}

	for _, ca := range cases {
		t.Run(ca.name, func(t *testing.T) {
			require.Equal(t, ca.out, h264.IsRandomAccess(ca.au))
		})
	}
}
