package mpegts

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

var klvaAccessUnitCases = []struct {
	name string
	dec  klvaAccessUnit
	enc  []byte
}{
	{
		"basic_klv_packet",
		klvaAccessUnit{
			Header: metadataAuCellHeader{
				PayloadSize:                 8,
				MetadataServiceID:           0xFC,
				CellFragmentationIndication: 0,
				DecoderConfigFlag:           false,
				RandomAccessIndicator:       true,
				SequenceNumber:              1,
			},
			Packet: []byte{0x06, 0x0E, 0x2B, 0x34, 0x01, 0x01, 0x01, 0x01},
		},
		[]byte{
			0xFC, 0x01, 0x10, 0x00, 0x08, // header (5 bytes)
			0x06, 0x0E, 0x2B, 0x34, 0x01, 0x01, 0x01, 0x01, // KLV packet
		},
	},
	{
		"klv_with_fragmentation",
		klvaAccessUnit{
			Header: metadataAuCellHeader{
				PayloadSize:                 12,
				MetadataServiceID:           0xFC,
				CellFragmentationIndication: 2,
				DecoderConfigFlag:           true,
				RandomAccessIndicator:       false,
				SequenceNumber:              255,
			},
			Packet: bytes.Repeat([]byte{0xAB}, 12),
		},
		[]byte{
			0xFC, 0xFF, 0xA0, 0x00, 0x0C, // header with fragmentation and decoder config
			0xAB, 0xAB, 0xAB, 0xAB, 0xAB, 0xAB, 0xAB, 0xAB, 0xAB, 0xAB, 0xAB, 0xAB,
		},
	},
}

func TestKLVAAccessUnitUnmarshal(t *testing.T) {
	for _, ca := range klvaAccessUnitCases {
		t.Run(ca.name, func(t *testing.T) {
			var dec klvaAccessUnit
			n, err := dec.unmarshal(ca.enc)
			require.NoError(t, err)
			require.Equal(t, len(ca.enc), n)
			require.Equal(t, ca.dec.Header, dec.Header)
			require.Equal(t, ca.dec.Packet, dec.Packet)
		})
	}
}

func TestKLVAAccessUnitMarshal(t *testing.T) {
	for _, ca := range klvaAccessUnitCases {
		t.Run(ca.name, func(t *testing.T) {
			size := ca.dec.marshalSize() + 5 // 5 bytes for header
			enc := make([]byte, size)
			n, err := ca.dec.marshalTo(enc)
			require.NoError(t, err)
			require.Equal(t, ca.enc, enc)
			require.Equal(t, len(ca.enc), n)
		})
	}
}

func TestKLVAAccessUnitUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedErr string
	}{
		{
			"header_decode_error",
			[]byte{0x00, 0x01, 0x10}, // invalid header (too short)
			"could not decode metadata AU header",
		},
		{
			"buffer_too_small",
			[]byte{
				0xFC, 0x01, 0x10, 0x00, 0x08, // header says 8 bytes payload
				0x06, 0x0E, 0x2B, // only 3 bytes available
			},
			"buffer is too small",
		},
		{
			"invalid_service_id",
			[]byte{
				0xFB, 0x01, 0x10, 0x00, 0x04, // invalid service ID (0xFB instead of 0xFC)
				0x06, 0x0E, 0x2B, 0x34,
			},
			"invalid prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var au klvaAccessUnit
			_, err := au.unmarshal(tt.data)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestKLVAAccessUnitMarshalErrors(t *testing.T) {
	tests := []struct {
		name        string
		au          klvaAccessUnit
		expectedErr string
	}{
		{
			"invalid_packet_size",
			klvaAccessUnit{
				Header: metadataAuCellHeader{
					PayloadSize:       8,
					MetadataServiceID: 0xFC,
				},
				Packet: []byte{0x01, 0x02, 0x03}, // only 3 bytes, but header says 8
			},
			"invalid packet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 100)
			_, err := tt.au.marshalTo(buf)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestKLVAAccessUnitMarshalSize(t *testing.T) {
	au := klvaAccessUnit{
		Header: metadataAuCellHeader{
			PayloadSize: 10,
		},
		Packet: make([]byte, 10),
	}

	size := au.marshalSize()
	require.Equal(t, 10, size) // Should return PayloadSize
}

func TestKLVAAccessUnitEdgeCases(t *testing.T) {
	t.Run("empty_packet", func(t *testing.T) {
		au := klvaAccessUnit{
			Header: metadataAuCellHeader{
				PayloadSize:       0,
				MetadataServiceID: 0xFC,
			},
			Packet: []byte{},
		}

		// Test marshal
		buf := make([]byte, 5)
		n, err := au.marshalTo(buf)
		require.NoError(t, err)
		require.Equal(t, 5, n) // Only header

		// Test unmarshal
		var decoded klvaAccessUnit
		n, err = decoded.unmarshal(buf)
		require.NoError(t, err)
		require.Equal(t, 5, n)
		require.Equal(t, au.Header.PayloadSize, decoded.Header.PayloadSize)
		require.Equal(t, 0, len(decoded.Packet))
	})

	t.Run("large_packet", func(t *testing.T) {
		largePacket := make([]byte, 65535) // Maximum size for 16-bit payload size
		for i := range largePacket {
			largePacket[i] = byte(i % 256)
		}

		au := klvaAccessUnit{
			Header: metadataAuCellHeader{
				PayloadSize:       65535,
				MetadataServiceID: 0xFC,
			},
			Packet: largePacket,
		}

		// Test marshal size
		size := au.marshalSize()
		require.Equal(t, 65535, size)

		// Test round trip
		buf := make([]byte, 65535+5)
		n, err := au.marshalTo(buf)
		require.NoError(t, err)
		require.Equal(t, 65535+5, n)

		var decoded klvaAccessUnit
		n, err = decoded.unmarshal(buf)
		require.NoError(t, err)
		require.Equal(t, 65535+5, n)
		require.Equal(t, largePacket, decoded.Packet)
	})
}

func FuzzKLVAAccessUnitUnmarshal(f *testing.F) {
	// Add seed cases
	for _, ca := range klvaAccessUnitCases {
		f.Add(ca.enc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var au klvaAccessUnit
		_, err := au.unmarshal(b)
		if err != nil {
			return
		}

		// If unmarshal succeeds, try to marshal back
		size := au.marshalSize() + 5
		if size > 0 && size < 1000000 { // Reasonable size limit
			buf := make([]byte, size)
			au.marshalTo(buf) //nolint:errcheck
		}
	})
}
