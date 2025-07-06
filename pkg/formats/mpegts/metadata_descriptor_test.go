package mpegts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesMetadataDescriptor = []struct {
	name string
	dec  metadataDescriptor
	enc  []byte
}{
	{
		"a",
		metadataDescriptor{
			MetadataApplicationFormat:           0xFFFF,
			MetadataApplicationFormatIdentifier: 893234,
			MetadataFormat:                      23,
			MetadataServiceID:                   234,
			DecoderConfigFlags:                  1,
			DSMCCFlag:                           true,
			ServiceIdentification:               []byte{1, 2},
			DecoderConfig:                       []byte{3, 4},
			PrivateData:                         []byte{5, 6},
		},
		[]byte{0xff, 0xff, 0x0, 0xd, 0xa1, 0x32, 0x17, 0xea, 0x3f, 0x2, 0x1, 0x2, 0x2, 0x3, 0x4, 0x5, 0x6},
	},
}

func TestMetadataDescriptorUnmarshal(t *testing.T) {
	for _, ca := range casesMetadataDescriptor {
		t.Run(ca.name, func(t *testing.T) {
			var h metadataDescriptor
			err := h.unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, ca.dec, h)
		})
	}
}

func TestMetadataDescriptorMarshal(t *testing.T) {
	for _, ca := range casesMetadataDescriptor {
		t.Run(ca.name, func(t *testing.T) {
			buf, err := ca.dec.marshal()
			require.NoError(t, err)
			require.Equal(t, ca.enc, buf)
		})
	}
}

func FuzzMetadataDescriptor(f *testing.F) {
	f.Fuzz(func(t *testing.T, buf []byte) {
		var dm metadataDescriptor
		err := dm.unmarshal(buf)
		if err != nil {
			return
		}

		_, err = dm.marshal()
		require.NoError(t, err)
	})
}
