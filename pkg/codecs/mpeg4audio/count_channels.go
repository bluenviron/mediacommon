package mpeg4audio

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

// AAC syntactic element IDs per ISO 14496-3 Table 4.85
const (
	idSCE = 0 // Single Channel Element (mono)
	idCPE = 1 // Channel Pair Element (stereo)
	idCCE = 2 // Coupling Channel Element
	idLFE = 3 // LFE Channel Element
	idDSE = 4 // Data Stream Element
	idPCE = 5 // Program Config Element
	idFIL = 6 // Fill Element
	idEND = 7 // End marker
)

// CountChannelsFromRawDataBlock counts channels by parsing the syntactic elements
// in an AAC raw_data_block. This is used when channel_configuration is 0 and
// no PCE is present (FFmpeg often omits PCE and uses implicit default ordering).
//
// Per ISO 14496-3, when channel_configuration=0, the channel layout can be
// determined either from an explicit PCE or inferred from the elements present.
// Common implicit layouts:
//   - 1 CPE = stereo (2 channels)
//   - 1 SCE + 1 CPE = 3.0 (3 channels: C, L, R)
//   - 1 SCE + 1 CPE + 1 CPE + 1 LFE = 5.1 (6 channels: C, L, R, Ls, Rs, LFE)
func CountChannelsFromRawDataBlock(au []byte) (int, error) {
	if len(au) < 1 {
		return 0, fmt.Errorf("raw_data_block too short")
	}

	pos := 0
	channels := 0

	// Parse elements until we hit END or run out of data
	for pos < len(au)*8 {
		// Read id_syn_ele (3 bits)
		idSynEle, err := bits.ReadBits(au, &pos, 3)
		if err != nil {
			// Ran out of bits - use what we've counted
			break
		}

		switch idSynEle {
		case idSCE:
			// Single Channel Element - 1 channel
			// Skip element_instance_tag (4 bits)
			_, err = bits.ReadBits(au, &pos, 4)
			if err != nil {
				break
			}
			channels++
			// We can't easily skip the rest of the SCE without fully parsing it,
			// so just count what we've seen and return
			return channels, nil

		case idCPE:
			// Channel Pair Element - 2 channels
			// Skip element_instance_tag (4 bits)
			_, err = bits.ReadBits(au, &pos, 4)
			if err != nil {
				break
			}
			channels += 2
			return channels, nil

		case idLFE:
			// LFE Channel Element - 1 channel
			// Skip element_instance_tag (4 bits)
			_, err = bits.ReadBits(au, &pos, 4)
			if err != nil {
				break
			}
			channels++
			return channels, nil

		case idPCE:
			// Found a PCE - parse it properly
			var pce ProgramConfigElement
			err = pce.unmarshal(au, &pos)
			if err != nil {
				return 0, fmt.Errorf("parsing PCE: %w", err)
			}
			return pce.ChannelCount, nil

		case idDSE:
			// Data Stream Element - skip it
			// DSE structure: element_instance_tag (4) + data_byte_align_flag (1) +
			// count (8) + [esc_count (8) if count==255] + data bytes
			// Skip element_instance_tag (4 bits)
			_, err = bits.ReadBits(au, &pos, 4)
			if err != nil {
				break
			}
			// data_byte_align_flag (1 bit)
			var alignFlag bool
			alignFlag, err = bits.ReadFlag(au, &pos)
			if err != nil {
				break
			}
			// count (8 bits)
			var count uint64
			count, err = bits.ReadBits(au, &pos, 8)
			if err != nil {
				break
			}
			totalCount := int(count)
			// If count == 255, read esc_count
			if count == 255 {
				var escCount uint64
				escCount, err = bits.ReadBits(au, &pos, 8)
				if err != nil {
					break
				}
				totalCount += int(escCount)
			}
			// Byte align if flag is set
			if alignFlag {
				pos = ((pos + 7) / 8) * 8
			}
			// Skip data_stream_byte[totalCount] (totalCount * 8 bits)
			pos += totalCount * 8
			// Continue to next element
			continue

		case idFIL:
			// Fill Element - skip it
			// FIL structure: count (4) + [esc_count (8) if count==15] + fill bytes
			var count uint64
			count, err = bits.ReadBits(au, &pos, 4)
			if err != nil {
				break
			}
			totalCount := int(count)
			if count == 15 {
				var escCount uint64
				escCount, err = bits.ReadBits(au, &pos, 8)
				if err != nil {
					break
				}
				totalCount += int(escCount) - 1
			}
			// Skip fill_byte[totalCount] or extension_payload
			pos += totalCount * 8
			// Continue to next element
			continue

		case idCCE:
			// Coupling Channel Element - complex, skip
			if channels > 0 {
				return channels, nil
			}
			return 0, fmt.Errorf("cannot determine channels from CCE element")

		case idEND:
			// End of raw_data_block
			if channels > 0 {
				return channels, nil
			}
			return 0, fmt.Errorf("no channel elements found before END")

		default:
			return 0, fmt.Errorf("unknown element id: %d", idSynEle)
		}
	}

	if channels == 0 {
		return 0, fmt.Errorf("no channel elements found")
	}

	return channels, nil
}
