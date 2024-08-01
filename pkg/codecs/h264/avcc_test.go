package h264

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesAVCC = []struct {
	name string
	enc  []byte
	dec  [][]byte
}{
	{
		"single",
		[]byte{
			0x00, 0x00, 0x00, 0x03,
			0xaa, 0xbb, 0xcc,
		},
		[][]byte{
			{0xaa, 0xbb, 0xcc},
		},
	},
	{
		"multiple",
		[]byte{
			0x00, 0x00, 0x00, 0x02,
			0xaa, 0xbb,
			0x00, 0x00, 0x00, 0x02,
			0xcc, 0xdd,
			0x00, 0x00, 0x00, 0x02,
			0xee, 0xff,
		},
		[][]byte{
			{0xaa, 0xbb},
			{0xcc, 0xdd},
			{0xee, 0xff},
		},
	},
}

func TestAVCCUnmarshal(t *testing.T) {
	for _, ca := range casesAVCC {
		t.Run(ca.name, func(t *testing.T) {
			dec, err := AVCCUnmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, dec)
		})
	}
}

// issue mediamtx/2375
func TestAVCCUnmarshalEmpty(t *testing.T) {
	dec, err := AVCCUnmarshal([]byte{
		0x0, 0x0, 0x0, 0x0,
	})

	require.Equal(t, ErrAVCCNoNALUs, err)
	require.Equal(t, [][]byte(nil), dec)

	dec, err = AVCCUnmarshal([]byte{
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x1, 0x2, 0x3,
	})

	require.NoError(t, err)
	require.Equal(t, [][]byte{
		{1, 2, 3},
	}, dec)
}

func TestAVCCMarshal(t *testing.T) {
	for _, ca := range casesAVCC {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := AVCCMarshal(ca.dec)
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func FuzzAVCCUnmarshal(f *testing.F) {
	for _, ca := range casesAVCC {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		au, err := AVCCUnmarshal(b)
		if err == nil {
			AVCCMarshal(au) //nolint:errcheck
		}
	})
}
