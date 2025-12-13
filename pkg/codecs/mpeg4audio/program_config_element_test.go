package mpeg4audio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePCEFromRawDataBlock(t *testing.T) {
	// Test case 1: Valid PCE for stereo (2 channels)
	t.Run("valid stereo PCE", func(t *testing.T) {
		// PCE with 1 front CPE (channel pair element) = 2 channels
		// Layout per ISO 14496-3:
		// - id_syn_ele (3): 101 = 5 (PCE)
		// - element_instance_tag (4): 0000
		// - object_type (2): 01 (AAC LC)
		// - sampling_frequency_index (4): 0100 (44100 Hz)
		// - num_front_channel_elements (4): 0001 (1 element)
		// - num_side_channel_elements (4): 0000
		// - num_back_channel_elements (4): 0000
		// - num_lfe_channel_elements (2): 00
		// - num_assoc_data_elements (3): 000
		// - num_valid_cc_elements (4): 0000
		// - mono_mixdown_present (1): 0
		// - stereo_mixdown_present (1): 0
		// - matrix_mixdown_idx_present (1): 0
		// - front[0]: is_cpe(1)=1, tag(4)=0000 (CPE = stereo pair)
		buf := []byte{0xA0, 0xA0, 0x80, 0x00, 0x04, 0x00}

		pce, err := ParsePCEFromRawDataBlock(buf)
		require.NoError(t, err)
		require.Equal(t, 2, pce.ChannelCount)
		require.Equal(t, uint8(1), pce.NumFrontChannelElements)
		require.Equal(t, uint8(0), pce.NumSideChannelElements)
		require.Equal(t, uint8(0), pce.NumBackChannelElements)
		require.Equal(t, uint8(0), pce.NumLFEChannelElements)
	})

	// Test case 2: Invalid - not a PCE
	t.Run("invalid not PCE", func(t *testing.T) {
		// id_syn_ele = 0 (SCE, not PCE)
		buf := []byte{0x00, 0x00, 0x00, 0x00}

		_, err := ParsePCEFromRawDataBlock(buf)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expected PCE")
	})

	// Test case 3: Too short
	t.Run("too short", func(t *testing.T) {
		buf := []byte{0xA0, 0x00}

		_, err := ParsePCEFromRawDataBlock(buf)
		require.Error(t, err)
	})
}

func TestCountChannelsFromRawDataBlock(t *testing.T) {
	// Test case 1: CPE (stereo pair) - id=1
	t.Run("CPE stereo", func(t *testing.T) {
		// id_syn_ele (3 bits) = 001 (CPE)
		// element_instance_tag (4 bits) = 0000
		// Byte: 001 0000 0 = 0x20
		buf := []byte{0x20, 0x00, 0x00, 0x00}

		channels, err := CountChannelsFromRawDataBlock(buf)
		require.NoError(t, err)
		require.Equal(t, 2, channels)
	})

	// Test case 2: SCE (mono) - id=0
	t.Run("SCE mono", func(t *testing.T) {
		// id_syn_ele (3 bits) = 000 (SCE)
		// element_instance_tag (4 bits) = 0000
		// Byte: 000 0000 0 = 0x00
		buf := []byte{0x00, 0x00, 0x00, 0x00}

		channels, err := CountChannelsFromRawDataBlock(buf)
		require.NoError(t, err)
		require.Equal(t, 1, channels)
	})

	// Test case 3: LFE - id=3
	t.Run("LFE", func(t *testing.T) {
		// id_syn_ele (3 bits) = 011 (LFE)
		// element_instance_tag (4 bits) = 0000
		// Byte: 011 0000 0 = 0x60
		buf := []byte{0x60, 0x00, 0x00, 0x00}

		channels, err := CountChannelsFromRawDataBlock(buf)
		require.NoError(t, err)
		require.Equal(t, 1, channels)
	})

	// Test case 4: Too short
	t.Run("too short", func(t *testing.T) {
		buf := []byte{}

		_, err := CountChannelsFromRawDataBlock(buf)
		require.Error(t, err)
	})

	// Test case 5: FIL element followed by CPE
	t.Run("FIL then CPE", func(t *testing.T) {
		// FIL element: id_syn_ele (3 bits) = 110 (FIL=6)
		// count (4 bits) = 0000 (no fill bytes)
		// Then CPE: id_syn_ele (3 bits) = 001, tag (4 bits) = 0000
		// Byte 0: 110 0000 0 = 0xC0 (FIL with count=0)
		// Byte 1: 01 0000 00 = 0x40 (CPE continuation - but we need proper alignment)
		// Actually: FIL takes 7 bits (3+4), CPE starts at bit 7
		// Bit 0-2: 110 (FIL)
		// Bit 3-6: 0000 (count=0)
		// Bit 7-9: 001 (CPE)
		// Bit 10-13: 0000 (tag)
		// Byte 0: 1100_0000 = 0xC0
		// Byte 1: 0100_0000 = 0x40
		buf := []byte{0xC0, 0x40, 0x00, 0x00}

		channels, err := CountChannelsFromRawDataBlock(buf)
		require.NoError(t, err)
		require.Equal(t, 2, channels)
	})

	// Test case 6: DSE element followed by CPE
	t.Run("DSE then CPE", func(t *testing.T) {
		// DSE element: id_syn_ele (3 bits) = 100 (DSE=4)
		// element_instance_tag (4 bits) = 0000
		// data_byte_align_flag (1 bit) = 0
		// count (8 bits) = 00000000 (no data bytes)
		// Then CPE: id_syn_ele (3 bits) = 001, tag (4 bits) = 0000
		// DSE takes 16 bits (3+4+1+8), CPE starts at bit 16
		// Byte 0: 100 0000 0 = 0x80 (DSE, tag=0, align=0 start)
		// Byte 1: 0000 0000 = 0x00 (count=0)
		// Byte 2: 001 0000 0 = 0x20 (CPE)
		buf := []byte{0x80, 0x00, 0x20, 0x00}

		channels, err := CountChannelsFromRawDataBlock(buf)
		require.NoError(t, err)
		require.Equal(t, 2, channels)
	})
}
