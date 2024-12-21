package av1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesLEB128 = []struct {
	name string
	dec  LEB128
	enc  []byte
}{
	{
		"a",
		1234567,
		[]byte{0x87, 0xad, 0x4b},
	},
	{
		"b",
		127,
		[]byte{0x7f},
	},
	{
		"c",
		651321342,
		[]byte{0xfe, 0xbf, 0xc9, 0xb6, 0x2},
	},
	{
		"max",
		(1 << 32) - 1,
		[]byte{0xff, 0xff, 0xff, 0xff, 0xf},
	},
}

func TestLEB128Unmarshal(t *testing.T) {
	for _, ca := range casesLEB128 {
		t.Run(ca.name, func(t *testing.T) {
			var dec LEB128
			n, err := dec.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, len(ca.enc), n)
			require.Equal(t, ca.dec, dec)
		})
	}
}

func TestLEB128Marshal(t *testing.T) {
	for _, ca := range casesLEB128 {
		t.Run(ca.name, func(t *testing.T) {
			enc := make([]byte, ca.dec.MarshalSize())
			n := ca.dec.MarshalTo(enc)
			require.Equal(t, ca.enc, enc)
			require.Equal(t, len(ca.enc), n)
		})
	}
}

func FuzzLEB128Unmarshal(f *testing.F) {
	for _, ca := range casesLEB128 {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var v LEB128
		_, err := v.Unmarshal(b)
		if err == nil {
			enc := make([]byte, v.MarshalSize())
			v.MarshalTo(enc)
		}
	})
}
