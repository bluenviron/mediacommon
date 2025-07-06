package mpegts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesMetadataAUCell = []struct {
	name string
	dec  metadataAUCell
	enc  []byte
}{
	{
		"a",
		metadataAUCell{
			MetadataServiceID:      15,
			SequenceNumber:         18,
			CellFragmentIndication: 3,
			DecoderConfigFlag:      true,
			RandomAccessIndicator:  false,
			AUCellData:             []byte{1, 2, 3, 4},
		},
		[]byte{0xf, 0x12, 0xef, 0x0, 0x4, 0x1, 0x2, 0x3, 0x4},
	},
}

func TestMetadataAUCellUnmarshal(t *testing.T) {
	for _, ca := range casesMetadataAUCell {
		t.Run(ca.name, func(t *testing.T) {
			var h metadataAUCell
			err := h.unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, h)
		})
	}
}

func TestMetadataAUCellMarshal(t *testing.T) {
	for _, ca := range casesMetadataAUCell {
		t.Run(ca.name, func(t *testing.T) {
			buf, err := ca.dec.marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, buf)
		})
	}
}

func FuzzMetadataAUCell(f *testing.F) {
	f.Fuzz(func(t *testing.T, buf []byte) {
		var c metadataAUCell
		err := c.unmarshal(buf)
		if err != nil {
			return
		}

		_, err = c.marshal()
		require.NoError(t, err)
	})
}
