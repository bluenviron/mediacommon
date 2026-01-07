package mpeg4audio

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

// ProgramConfigElement represents an AAC Program Config Element.
// Specification: ISO/IEC 14496-3, table 4.2
type ProgramConfigElement struct {
	ElementInstanceTag      uint8
	ObjectType              uint8
	SamplingFrequencyIndex  uint8
	NumFrontChannelElements uint8
	NumSideChannelElements  uint8
	NumBackChannelElements  uint8
	NumLFEChannelElements   uint8
	NumAssocDataElements    uint8
	NumValidCCElements      uint8
	ChannelCount            int
}

// unmarshal parses a Program Config Element from an AAC raw_data_block.
func (pce *ProgramConfigElement) unmarshal(buf []byte, pos *int) error {
	err := bits.HasSpace(buf, *pos, 32)
	if err != nil {
		return err
	}

	// element_instance_tag (4 bits)
	pce.ElementInstanceTag = uint8(bits.ReadBitsUnsafe(buf, pos, 4))

	// object_type (2 bits)
	pce.ObjectType = uint8(bits.ReadBitsUnsafe(buf, pos, 2))

	// sampling_frequency_index (4 bits)
	pce.SamplingFrequencyIndex = uint8(bits.ReadBitsUnsafe(buf, pos, 4))

	// num_front_channel_elements (4 bits)
	pce.NumFrontChannelElements = uint8(bits.ReadBitsUnsafe(buf, pos, 4))

	// num_side_channel_elements (4 bits)
	pce.NumSideChannelElements = uint8(bits.ReadBitsUnsafe(buf, pos, 4))

	// num_back_channel_elements (4 bits)
	pce.NumBackChannelElements = uint8(bits.ReadBitsUnsafe(buf, pos, 4))

	// num_lfe_channel_elements (2 bits)
	pce.NumLFEChannelElements = uint8(bits.ReadBitsUnsafe(buf, pos, 2))

	// num_assoc_data_elements (3 bits)
	pce.NumAssocDataElements = uint8(bits.ReadBitsUnsafe(buf, pos, 3))

	// num_valid_cc_elements (4 bits)
	pce.NumValidCCElements = uint8(bits.ReadBitsUnsafe(buf, pos, 4))

	// mono_mixdown_present (1 bit)
	monoMixdownPresent := bits.ReadFlagUnsafe(buf, pos)

	if monoMixdownPresent {
		// Skip mono_mixdown_element_number (4 bits)
		_, err = bits.ReadBits(buf, pos, 4)
		if err != nil {
			return fmt.Errorf("reading mono_mixdown_element_number: %w", err)
		}
	}

	// stereo_mixdown_present (1 bit)
	stereoMixdownPresent, err := bits.ReadFlag(buf, pos)
	if err != nil {
		return fmt.Errorf("reading stereo_mixdown_present: %w", err)
	}

	if stereoMixdownPresent {
		// Skip stereo_mixdown_element_number (4 bits)
		_, err = bits.ReadBits(buf, pos, 4)
		if err != nil {
			return fmt.Errorf("reading stereo_mixdown_element_number: %w", err)
		}
	}

	// matrix_mixdown_idx_present (1 bit)
	matrixMixdownIdxPresent, err := bits.ReadFlag(buf, pos)
	if err != nil {
		return fmt.Errorf("reading matrix_mixdown_idx_present: %w", err)
	}

	if matrixMixdownIdxPresent {
		// Skip matrix_mixdown_idx (2 bits) + pseudo_surround_enable (1 bit)
		_, err = bits.ReadBits(buf, pos, 3)
		if err != nil {
			return fmt.Errorf("reading matrix_mixdown_idx: %w", err)
		}
	}

	// Count channels from front elements
	// Each element: is_cpe (1 bit) + element_tag_select (4 bits)
	// is_cpe=1 means stereo pair (2 channels), is_cpe=0 means mono (1 channel)
	channels := 0

	for i := uint8(0); i < pce.NumFrontChannelElements; i++ {
		var isCPE bool
		isCPE, err = bits.ReadFlag(buf, pos)
		if err != nil {
			return fmt.Errorf("reading front element is_cpe: %w", err)
		}
		// Skip element_tag_select (4 bits)
		_, err = bits.ReadBits(buf, pos, 4)
		if err != nil {
			return fmt.Errorf("reading front element tag: %w", err)
		}
		if isCPE {
			channels += 2
		} else {
			channels++
		}
	}

	// Count channels from side elements
	for i := uint8(0); i < pce.NumSideChannelElements; i++ {
		var isCPE bool
		isCPE, err = bits.ReadFlag(buf, pos)
		if err != nil {
			return fmt.Errorf("reading side element is_cpe: %w", err)
		}
		// Skip element_tag_select (4 bits)
		_, err = bits.ReadBits(buf, pos, 4)
		if err != nil {
			return fmt.Errorf("reading side element tag: %w", err)
		}
		if isCPE {
			channels += 2
		} else {
			channels++
		}
	}

	// Count channels from back elements
	for i := uint8(0); i < pce.NumBackChannelElements; i++ {
		var isCPE bool
		isCPE, err = bits.ReadFlag(buf, pos)
		if err != nil {
			return fmt.Errorf("reading back element is_cpe: %w", err)
		}
		// Skip element_tag_select (4 bits)
		_, err = bits.ReadBits(buf, pos, 4)
		if err != nil {
			return fmt.Errorf("reading back element tag: %w", err)
		}
		if isCPE {
			channels += 2
		} else {
			channels++
		}
	}

	// LFE elements are always mono (no is_cpe flag)
	// Just element_tag_select (4 bits) per LFE element
	for i := uint8(0); i < pce.NumLFEChannelElements; i++ {
		_, err = bits.ReadBits(buf, pos, 4)
		if err != nil {
			return fmt.Errorf("reading lfe element tag: %w", err)
		}
		channels++
	}

	if channels == 0 {
		return fmt.Errorf("PCE has zero channels")
	}

	pce.ChannelCount = channels

	// We've extracted the channel count - skip remaining fields
	// (assoc_data, valid_cc, byte_alignment, comment)
	// These aren't needed for channel count determination

	return nil
}

// ParsePCEFromRawDataBlock attempts to parse a PCE from the start of an AAC raw_data_block.
// This is used when channel_configuration is 0 in the ADTS header.
// The raw_data_block starts with an id_syn_ele (3 bits) which identifies the element type.
// PCE has id_syn_ele = 5.
func ParsePCEFromRawDataBlock(au []byte) (*ProgramConfigElement, error) {
	if len(au) < 4 {
		return nil, fmt.Errorf("raw_data_block too short")
	}

	pos := 0

	// Read id_syn_ele (3 bits)
	idSynEle, err := bits.ReadBits(au, &pos, 3)
	if err != nil {
		return nil, fmt.Errorf("reading id_syn_ele: %w", err)
	}

	// ID_PCE = 5 (Program Config Element)
	const idPCE = 5
	if idSynEle != idPCE {
		return nil, fmt.Errorf("expected PCE (id=5), got id=%d", idSynEle)
	}

	var pce ProgramConfigElement
	err = pce.unmarshal(au, &pos)
	if err != nil {
		return nil, err
	}

	return &pce, nil
}
