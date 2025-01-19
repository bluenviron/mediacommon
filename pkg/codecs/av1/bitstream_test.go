package av1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesBitstream = []struct {
	name string
	enc  []byte
	dec  Bitstream
}{
	{
		"standard",
		[]byte{
			0x0a, 0x0e, 0x00, 0x00, 0x00, 0x4a, 0xab, 0xbf,
			0xc3, 0x77, 0x6b, 0xe4, 0x40, 0x40, 0x40, 0x41,
			0x0a, 0x0e, 0x00, 0x00, 0x00, 0x4a, 0xab, 0xbf,
			0xc3, 0x77, 0x6b, 0xe4, 0x40, 0x40, 0x40, 0x41,
		},
		[][]byte{
			{
				0x0a, 0x0e, 0x00, 0x00, 0x00, 0x4a, 0xab, 0xbf,
				0xc3, 0x77, 0x6b, 0xe4, 0x40, 0x40, 0x40, 0x41,
			},
			{
				0x0a, 0x0e, 0x00, 0x00, 0x00, 0x4a, 0xab, 0xbf,
				0xc3, 0x77, 0x6b, 0xe4, 0x40, 0x40, 0x40, 0x41,
			},
		},
	},
}

func TestBitstreamUnmarshal(t *testing.T) {
	for _, ca := range casesBitstream {
		t.Run(ca.name, func(t *testing.T) {
			var dec Bitstream
			err := dec.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, dec)
		})
	}
}

func TestBitstreamMarshal(t *testing.T) {
	for _, ca := range casesBitstream {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func TestBitstreamMarshalWithoutSize(t *testing.T) {
	for _, ca := range casesBitstream {
		t.Run(ca.name, func(t *testing.T) {
			var tu Bitstream
			for _, obu := range ca.dec {
				var size LEB128
				n, err := size.Unmarshal(obu[1:])
				require.NoError(t, err)

				newObu := make([]byte, len(obu)-n)
				newObu[0] = obu[0] & 0b01111000
				copy(newObu[1:], obu[1+n:])
				tu = append(tu, newObu)
			}

			enc, err := tu.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func FuzzBitstreamUnmarshal(f *testing.F) {
	for _, ca := range casesBitstream {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var tu Bitstream
		err := tu.Unmarshal(b)
		if err == nil {
			tu.Marshal() //nolint:errcheck
		}
	})
}
