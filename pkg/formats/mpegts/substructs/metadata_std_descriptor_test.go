package substructs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/substructs"
)

var casesMetadataSTDDescriptor = []struct {
	name string
	dec  substructs.MetadataSTDDescriptor
	enc  []byte
}{
	{
		"a",
		substructs.MetadataSTDDescriptor{
			MetadataInputLeakRate:  463412,
			MetadataBufferSize:     834523,
			MetadataOutputLeakRate: 845324,
		},
		[]byte{0xc7, 0x12, 0x34, 0xcc, 0xbb, 0xdb, 0xcc, 0xe6, 0xc},
	},
}

func TestMetadataSTDDescriptorUnmarshal(t *testing.T) {
	for _, ca := range casesMetadataSTDDescriptor {
		t.Run(ca.name, func(t *testing.T) {
			var h substructs.MetadataSTDDescriptor
			err := h.Unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, h)
		})
	}
}

func TestMetadataSTDDescriptorMarshal(t *testing.T) {
	for _, ca := range casesMetadataSTDDescriptor {
		t.Run(ca.name, func(t *testing.T) {
			buf, err := ca.dec.Marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, buf)
		})
	}
}

func FuzzMetadataSTDDescriptor(f *testing.F) {
	f.Fuzz(func(t *testing.T, buf []byte) {
		var dm substructs.MetadataSTDDescriptor
		err := dm.Unmarshal(buf)
		if err != nil {
			return
		}

		_, err = dm.Marshal()
		require.NoError(t, err)
	})
}
