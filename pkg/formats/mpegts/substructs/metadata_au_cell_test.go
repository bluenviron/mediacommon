package substructs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/substructs"
)

var casesMetadataAUCell = []struct {
	name string
	dec  substructs.MetadataAUCell
	enc  []byte
}{
	{
		"a",
		substructs.MetadataAUCell{
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
			var h substructs.MetadataAUCell
			n, err := h.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, n, len(ca.enc))
			require.Equal(t, ca.dec, h)
		})
	}
}

func TestMetadataAUCellMarshal(t *testing.T) {
	for _, ca := range casesMetadataAUCell {
		t.Run(ca.name, func(t *testing.T) {
			buf := make([]byte, len(ca.enc))
			n, err := ca.dec.MarshalTo(buf)
			require.NoError(t, err)
			require.Equal(t, len(buf), n)
			require.Equal(t, ca.enc, buf)
		})
	}
}

func FuzzMetadataAUCell(f *testing.F) {
	f.Fuzz(func(t *testing.T, buf []byte) {
		var c substructs.MetadataAUCell
		_, err := c.Unmarshal(buf)
		if err != nil {
			return
		}

		require.NotEmpty(t, c.AUCellData)

		out := make([]byte, len(c.AUCellData)+5)
		_, err = c.MarshalTo(out)
		require.NoError(t, err)
	})
}
