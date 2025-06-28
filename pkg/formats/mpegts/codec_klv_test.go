package mpegts

import (
	"testing"

	"github.com/asticode/go-astits"
	"github.com/stretchr/testify/require"
)

func TestCodecKLVIsVideo(t *testing.T) {
	codec := CodecKLV{}
	require.False(t, codec.IsVideo(), "KLV codec should not be a video codec")
}

func TestCodecKLVIsCodec(_ *testing.T) {
	codec := &CodecKLV{}
	// This test ensures the isCodec method exists and can be called
	// The method is used for interface compliance
	codec.isCodec()
}

func TestCodecKLVMarshal(t *testing.T) {
	tests := []struct {
		name     string
		codec    CodecKLV
		pid      uint16
		expected astits.PMTElementaryStream
	}{
		{
			"metadata_stream_type",
			CodecKLV{
				StreamType:      astits.StreamTypeMetadata,
				StreamID:        0xFC,
				PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
			},
			257,
			astits.PMTElementaryStream{
				ElementaryPID: 257,
				StreamType:    astits.StreamTypeMetadata,
				ElementaryStreamDescriptors: []*astits.Descriptor{
					{
						Length: 4,
						Tag:    astits.DescriptorTagRegistration,
						Registration: &astits.DescriptorRegistration{
							FormatIdentifier: klvaIdentifier,
						},
					},
				},
			},
		},
		{
			"private_data_stream_type",
			CodecKLV{
				StreamType:      astits.StreamTypePrivateData,
				StreamID:        0xFC,
				PTSDTSIndicator: astits.PTSDTSIndicatorNoPTSOrDTS,
			},
			258,
			astits.PMTElementaryStream{
				ElementaryPID: 258,
				StreamType:    astits.StreamTypePrivateData,
				ElementaryStreamDescriptors: []*astits.Descriptor{
					{
						Length: 4,
						Tag:    astits.DescriptorTagRegistration,
						Registration: &astits.DescriptorRegistration{
							FormatIdentifier: klvaIdentifier,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.codec.marshal(tt.pid)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tt.expected.ElementaryPID, result.ElementaryPID)
			require.Equal(t, tt.expected.StreamType, result.StreamType)
			require.Len(t, result.ElementaryStreamDescriptors, 1)

			desc := result.ElementaryStreamDescriptors[0]
			require.Equal(t, uint8(astits.DescriptorTagRegistration), desc.Tag)
			require.Equal(t, uint8(4), desc.Length)
			require.NotNil(t, desc.Registration)
			require.Equal(t, uint32(klvaIdentifier), desc.Registration.FormatIdentifier)
		})
	}
}

func TestCodecKLVMarshalDifferentPIDs(t *testing.T) {
	codec := CodecKLV{
		StreamType:      astits.StreamTypeMetadata,
		StreamID:        0xFC,
		PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
	}

	pids := []uint16{256, 257, 258, 1000, 8191} // Various valid PIDs

	for _, pid := range pids {
		t.Run("pid_"+string(rune(pid)), func(t *testing.T) {
			result, err := codec.marshal(pid)
			require.NoError(t, err)
			require.Equal(t, pid, result.ElementaryPID)
		})
	}
}

func TestCodecKLVStreamTypes(t *testing.T) {
	streamTypes := []astits.StreamType{
		astits.StreamTypeMetadata,
		astits.StreamTypePrivateData,
	}

	for _, streamType := range streamTypes {
		t.Run("stream_type_"+string(rune(streamType)), func(t *testing.T) {
			codec := CodecKLV{
				StreamType:      streamType,
				StreamID:        0xFC,
				PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
			}

			result, err := codec.marshal(257)
			require.NoError(t, err)
			require.Equal(t, streamType, result.StreamType)
		})
	}
}

func TestCodecKLVPTSDTSIndicators(t *testing.T) {
	indicators := []uint8{
		astits.PTSDTSIndicatorNoPTSOrDTS,
		astits.PTSDTSIndicatorOnlyPTS,
		astits.PTSDTSIndicatorBothPresent,
	}

	for _, indicator := range indicators {
		t.Run("pts_dts_indicator_"+string(rune(indicator)), func(t *testing.T) {
			codec := CodecKLV{
				StreamType:      astits.StreamTypeMetadata,
				StreamID:        0xFC,
				PTSDTSIndicator: indicator,
			}

			// The marshal method doesn't use PTSDTSIndicator, but we test that it doesn't affect marshaling
			result, err := codec.marshal(257)
			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestCodecKLVStreamIDs(t *testing.T) {
	streamIDs := []uint8{0xFC, 0xFD, 0xFE, 0xFF}

	for _, streamID := range streamIDs {
		t.Run("stream_id_"+string(rune(streamID)), func(t *testing.T) {
			codec := CodecKLV{
				StreamType:      astits.StreamTypeMetadata,
				StreamID:        streamID,
				PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
			}

			// The marshal method doesn't use StreamID, but we test that it doesn't affect marshaling
			result, err := codec.marshal(257)
			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestCodecKLVRegistrationDescriptor(t *testing.T) {
	codec := CodecKLV{
		StreamType:      astits.StreamTypeMetadata,
		StreamID:        0xFC,
		PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
	}

	result, err := codec.marshal(257)
	require.NoError(t, err)
	require.Len(t, result.ElementaryStreamDescriptors, 1)

	desc := result.ElementaryStreamDescriptors[0]
	require.Equal(t, uint8(astits.DescriptorTagRegistration), desc.Tag)
	require.Equal(t, uint8(4), desc.Length)
	require.NotNil(t, desc.Registration)

	// Test that the KLVA identifier is correctly set
	expectedIdentifier := klvaIdentifier
	require.Equal(t, uint32(expectedIdentifier), desc.Registration.FormatIdentifier)
}

func TestCodecKLVInterfaceCompliance(t *testing.T) {
	// Test that CodecKLV implements the Codec interface
	var codec Codec = &CodecKLV{}
	require.False(t, codec.IsVideo())

	// Test that it can be used in a slice of codecs
	codecs := []Codec{
		&CodecKLV{},
		&CodecH264{},
		&CodecH265{},
	}

	for i, c := range codecs {
		if i == 0 {
			require.False(t, c.IsVideo(), "KLV should not be video")
		} else {
			require.True(t, c.IsVideo(), "H264/H265 should be video")
		}
	}
}
