package mpegts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var metadataAuHeaderCases = []struct {
	name string
	dec  metadataAuCellHeader
	enc  []byte
}{
	{
		"basic_header",
		metadataAuCellHeader{
			PayloadSize:                 8,
			MetadataServiceID:           0xFC,
			CellFragmentationIndication: 0,
			DecoderConfigFlag:           false,
			RandomAccessIndicator:       true,
			SequenceNumber:              1,
		},
		[]byte{0xFC, 0x01, 0x10, 0x00, 0x08},
	},
	{
		"header_with_fragmentation",
		metadataAuCellHeader{
			PayloadSize:                 1024,
			MetadataServiceID:           0xFC,
			CellFragmentationIndication: 2,
			DecoderConfigFlag:           true,
			RandomAccessIndicator:       false,
			SequenceNumber:              255,
		},
		[]byte{0xFC, 0xFF, 0xA0, 0x04, 0x00},
	},
	{
		"header_with_all_flags",
		metadataAuCellHeader{
			PayloadSize:                 65535,
			MetadataServiceID:           0xFC,
			CellFragmentationIndication: 3,
			DecoderConfigFlag:           true,
			RandomAccessIndicator:       true,
			SequenceNumber:              0,
		},
		[]byte{0xFC, 0x00, 0xF0, 0xFF, 0xFF},
	},
	{
		"minimal_header",
		metadataAuCellHeader{
			PayloadSize:                 0,
			MetadataServiceID:           0xFC,
			CellFragmentationIndication: 0,
			DecoderConfigFlag:           false,
			RandomAccessIndicator:       false,
			SequenceNumber:              0,
		},
		[]byte{0xFC, 0x00, 0x00, 0x00, 0x00},
	},
}

func TestMetadataAuHeaderUnmarshal(t *testing.T) {
	for _, ca := range metadataAuHeaderCases {
		t.Run(ca.name, func(t *testing.T) {
			var dec metadataAuCellHeader
			n, err := dec.unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, 5, n) // Header is always 5 bytes
			require.Equal(t, ca.dec, dec)
		})
	}
}

func TestMetadataAuHeaderMarshal(t *testing.T) {
	for _, ca := range metadataAuHeaderCases {
		t.Run(ca.name, func(t *testing.T) {
			enc := make([]byte, 5)
			n, err := ca.dec.marshalTo(enc)
			require.NoError(t, err)
			require.Equal(t, 5, n)
			require.Equal(t, ca.enc, enc)
		})
	}
}

func TestMetadataAuHeaderUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedErr string
	}{
		{
			"buffer_too_small",
			[]byte{0xFC, 0x01, 0x10}, // Only 3 bytes, need 5
			"not enough bits",
		},
		{
			"invalid_service_id",
			[]byte{0xFB, 0x01, 0x10, 0x00, 0x08}, // 0xFB instead of 0xFC
			"invalid prefix: 251",
		},
		{
			"empty_buffer",
			[]byte{},
			"not enough bits",
		},
		{
			"single_byte",
			[]byte{0xFC},
			"not enough bits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var header metadataAuCellHeader
			_, err := header.unmarshal(tt.data)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestMetadataAuHeaderFieldValues(t *testing.T) {
	tests := []struct {
		name   string
		header metadataAuCellHeader
	}{
		{
			"max_sequence_number",
			metadataAuCellHeader{
				MetadataServiceID: 0xFC,
				SequenceNumber:    255,
				PayloadSize:       100,
			},
		},
		{
			"max_fragmentation",
			metadataAuCellHeader{
				MetadataServiceID:           0xFC,
				CellFragmentationIndication: 3, // 2 bits max
				PayloadSize:                 100,
			},
		},
		{
			"max_payload_size",
			metadataAuCellHeader{
				MetadataServiceID: 0xFC,
				PayloadSize:       65535, // 16 bits max
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshal
			buf := make([]byte, 5)
			n, err := tt.header.marshalTo(buf)
			require.NoError(t, err)
			require.Equal(t, 5, n)

			// Test unmarshal
			var decoded metadataAuCellHeader
			n, err = decoded.unmarshal(buf)
			require.NoError(t, err)
			require.Equal(t, 5, n)
			require.Equal(t, tt.header, decoded)
		})
	}
}

func TestMetadataAuHeaderBitFields(t *testing.T) {
	t.Run("fragmentation_bits", func(t *testing.T) {
		for i := uint8(0); i < 4; i++ { // 2 bits = 4 values
			header := metadataAuCellHeader{
				MetadataServiceID:           0xFC,
				CellFragmentationIndication: i,
				PayloadSize:                 100,
			}

			buf := make([]byte, 5)
			_, err := header.marshalTo(buf)
			require.NoError(t, err)

			var decoded metadataAuCellHeader
			_, err = decoded.unmarshal(buf)
			require.NoError(t, err)
			require.Equal(t, i, decoded.CellFragmentationIndication)
		}
	})

	t.Run("flag_combinations", func(t *testing.T) {
		combinations := []struct {
			decoderConfig bool
			randomAccess  bool
		}{
			{false, false},
			{false, true},
			{true, false},
			{true, true},
		}

		for _, combo := range combinations {
			header := metadataAuCellHeader{
				MetadataServiceID:     0xFC,
				DecoderConfigFlag:     combo.decoderConfig,
				RandomAccessIndicator: combo.randomAccess,
				PayloadSize:           100,
			}

			buf := make([]byte, 5)
			_, err := header.marshalTo(buf)
			require.NoError(t, err)

			var decoded metadataAuCellHeader
			_, err = decoded.unmarshal(buf)
			require.NoError(t, err)
			require.Equal(t, combo.decoderConfig, decoded.DecoderConfigFlag)
			require.Equal(t, combo.randomAccess, decoded.RandomAccessIndicator)
		}
	})
}

func TestMetadataAuHeaderReservedBits(t *testing.T) {
	// Test that reserved bits are properly handled
	header := metadataAuCellHeader{
		MetadataServiceID: 0xFC,
		SequenceNumber:    1,
		PayloadSize:       100,
	}

	buf := make([]byte, 5)
	_, err := header.marshalTo(buf)
	require.NoError(t, err)

	// Check that reserved bits (bits 4-7 of byte 2) are zero
	reservedBits := buf[2] & 0x0F
	require.Equal(t, uint8(0), reservedBits, "Reserved bits should be zero")

	// Verify unmarshal still works
	var decoded metadataAuCellHeader
	_, err = decoded.unmarshal(buf)
	require.NoError(t, err)
	require.Equal(t, header, decoded)
}

func FuzzMetadataAuHeaderUnmarshal(f *testing.F) {
	// Add seed cases
	for _, ca := range metadataAuHeaderCases {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var header metadataAuCellHeader
		_, err := header.unmarshal(b)
		if err != nil {
			return
		}

		// If unmarshal succeeds, try to marshal back
		buf := make([]byte, 5)
		header.marshalTo(buf) //nolint:errcheck
	})
}
