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

	// Test case 4: PCE with mono_mixdown_present=1
	t.Run("PCE with mono mixdown", func(t *testing.T) {
		// Same structure as stereo PCE but with mono_mixdown_present=1
		// Bit layout:
		// Bits 0-33: same as original (id=PCE, tag=0, obj=1, sfi=4, nfront=1, nside=0, nback=0, nlfe=0, nassoc=0, ncc=0)
		// Bit 34: mono_mixdown_present = 1
		// Bits 35-38: mono_mixdown_element_number = 0000
		// Bit 39: stereo_mixdown_present = 0
		// Bit 40: matrix_mixdown_idx_present = 0
		// Bit 41: front[0].is_cpe = 1
		// Bits 42-45: front[0].tag = 0000
		//
		// Calculated bytes:
		// Byte 0-3: same as original (0xA0, 0xA0, 0x80, 0x00)
		// Byte 4: bits 32-39 = 00 1 0000 0 = 0x20 (num_cc=0, mono=1, mono_elem=0000, stereo=0)
		// Byte 5: bits 40-47 = 0 1 0000 00 = 0x40 (matrix=0, is_cpe=1, tag=0000)
		buf := []byte{0xA0, 0xA0, 0x80, 0x00, 0x20, 0x40, 0x00}

		pce, err := ParsePCEFromRawDataBlock(buf)
		require.NoError(t, err)
		require.Equal(t, 2, pce.ChannelCount)
	})

	// Test case 5: PCE with side and back channels (5.1 surround)
	t.Run("PCE 5.1 surround", func(t *testing.T) {
		// 5.1 = 3 front + 2 back + 1 LFE = 6 channels
		// PCE configuration: 2 front elements (SCE + CPE) + 1 back element (CPE) + 1 LFE
		//
		// Bit layout:
		// Bits 0-2: id_syn_ele = 101 (PCE)
		// Bits 3-6: element_instance_tag = 0000
		// Bits 7-8: object_type = 01
		// Bits 9-12: sampling_frequency_index = 0100 (44100)
		// Bits 13-16: num_front = 0010 (2 elements)
		// Bits 17-20: num_side = 0000
		// Bits 21-24: num_back = 0001 (1 element)
		// Bits 25-26: num_lfe = 01 (1 LFE)
		// Bits 27-29: num_assoc = 000
		// Bits 30-33: num_cc = 0000
		// Bit 34: mono_mixdown = 0
		// Bit 35: stereo_mixdown = 0
		// Bit 36: matrix_mixdown = 0
		// Bits 37-41: front[0] = 0 0000 (SCE, 1 channel)
		// Bits 42-46: front[1] = 1 0000 (CPE, 2 channels)
		// Bits 47-51: back[0] = 1 0000 (CPE, 2 channels)
		// Bits 52-55: lfe[0] = 0000 (1 channel)
		//
		// Calculated bytes:
		// Byte 0: 101 0000 0 = 0xA0
		// Byte 1: 1 0100 001 = 0xA1 (obj=1, sfi=4, nfront high bits)
		// Byte 2: 0 0000 000 = 0x00 (nfront low, nside, nback high)
		// Byte 3: 1 01 000 00 = 0xA0 (nback low, nlfe=1, nassoc=0, ncc high)
		// Byte 4: 00 0 0 0 0 00 = 0x00 (ncc low, mono=0, stereo=0, matrix=0, front[0] start)
		// Byte 5: 00 1 0000 1 = 0x21 (front[0] tag, front[1] is_cpe=1, front[1] tag, back start)
		// Byte 6: 0000 0000 = 0x00 (back[0] tag, lfe[0] tag)
		buf := []byte{0xA0, 0xA1, 0x00, 0xA0, 0x00, 0x21, 0x00, 0x00}

		pce, err := ParsePCEFromRawDataBlock(buf)
		require.NoError(t, err)
		require.Equal(t, 6, pce.ChannelCount) // 5.1 surround
		require.Equal(t, uint8(2), pce.NumFrontChannelElements)
		require.Equal(t, uint8(1), pce.NumBackChannelElements)
		require.Equal(t, uint8(1), pce.NumLFEChannelElements)
	})
}
