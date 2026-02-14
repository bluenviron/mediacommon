package h264

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesAVCC = []struct {
	name string
	enc  []byte
	dec  AVCC
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
			var dec AVCC
			err := dec.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, dec)
		})
	}
}

// issue mediamtx/2375
func TestAVCCUnmarshalEmpty(t *testing.T) {
	var dec AVCC
	err := dec.Unmarshal([]byte{
		0x0, 0x0, 0x0, 0x0,
	})

	require.Equal(t, ErrAVCCNoNALUs, err)
	require.Equal(t, AVCC(nil), dec)

	err = dec.Unmarshal([]byte{
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x1, 0x2, 0x3,
	})

	require.NoError(t, err)
	require.Equal(t, AVCC{
		{1, 2, 3},
	}, dec)
}

func TestAVCCUnmarshalExceedsMaxNALUs(t *testing.T) {
	buf := make([]byte, 0, 5*51)
	for range 51 {
		buf = append(buf, 0x00, 0x00, 0x00, 0x01, 0xAA)
	}

	var dec AVCC
	err := dec.Unmarshal(buf)
	require.EqualError(t, err, "NALU count (51) exceeds maximum allowed (50)")
}

func TestAVCCMarshal(t *testing.T) {
	for _, ca := range casesAVCC {
		t.Run(ca.name, func(t *testing.T) {
			enc, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func FuzzAVCCUnmarshal(f *testing.F) {
	for _, ca := range casesAVCC {
		f.Add(ca.enc)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		var au AVCC
		err := au.Unmarshal(b)
		if err != nil {
			return
		}

		require.NotZero(t, len(au))

		for _, nalu := range au {
			require.NotZero(t, len(nalu))
		}

		_, err = au.Marshal()
		require.NoError(t, err)
	})
}
