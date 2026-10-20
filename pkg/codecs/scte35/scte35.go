// Package scte35 contains utilities to work with SCTE-35 splice information.
// Specification: ANSI/SCTE 35
package scte35

import (
	"fmt"
)

const (
	// TableID is the table_id of a splice_info_section.
	TableID = 0xFC

	// minimum size of a splice_info_section:
	// table_id (1) + section_length fields (2) + protocol_version (1) +
	// encryption fields and pts_adjustment (5) + cw_index (1) +
	// tier and splice_command_length (3) + splice_command_type (1) +
	// descriptor_loop_length (2) + CRC_32 (4)
	minSectionSize = 20

	ptsMask = 0x1FFFFFFFF // 33 bits
)

// Section is a SCTE-35 splice_info_section, kept in binary form.
// Only the fields needed to relay the section are decoded.
type Section []byte

// Validate checks that the section is a well-formed splice_info_section
// with a valid CRC_32.
func (s Section) Validate() error {
	if len(s) < minSectionSize {
		return fmt.Errorf("section is too short")
	}

	if s[0] != TableID {
		return fmt.Errorf("invalid table_id: 0x%02x", s[0])
	}

	sectionLength := int(s[1]&0x0F)<<8 | int(s[2])
	if (3 + sectionLength) != len(s) {
		return fmt.Errorf("invalid section_length: %d, expected %d", sectionLength, len(s)-3)
	}

	if crc32MPEG2(s) != 0 {
		return fmt.Errorf("CRC_32 mismatch")
	}

	return nil
}

// PTSAdjustment returns the pts_adjustment field.
// The section must have been validated.
func (s Section) PTSAdjustment() int64 {
	return int64(s[4]&0x01)<<32 | int64(s[5])<<24 | int64(s[6])<<16 | int64(s[7])<<8 | int64(s[8])
}

// WithPTSAdjustment returns a copy of the section with the pts_adjustment field
// set to v (modulo 2^33) and the CRC_32 recomputed.
// The section must have been validated.
//
// pts_adjustment is transmitted in clear even when the section is encrypted,
// therefore this is safe to use on encrypted sections too.
func (s Section) WithPTSAdjustment(v int64) Section {
	v &= ptsMask

	out := make(Section, len(s))
	copy(out, s)

	out[4] = (out[4] & 0xFE) | byte(v>>32)
	out[5] = byte(v >> 24)
	out[6] = byte(v >> 16)
	out[7] = byte(v >> 8)
	out[8] = byte(v)

	n := len(out) - 4
	crc := crc32MPEG2(out[:n])
	out[n] = byte(crc >> 24)
	out[n+1] = byte(crc >> 16)
	out[n+2] = byte(crc >> 8)
	out[n+3] = byte(crc)

	return out
}

// CRC-32/MPEG-2: polynomial 0x04C11DB7, initial value 0xFFFFFFFF, no reflection, no final XOR.
// When computed over a section including its CRC_32 field, the result is zero.
func crc32MPEG2(buf []byte) uint32 {
	crc := uint32(0xFFFFFFFF)
	for _, b := range buf {
		crc ^= uint32(b) << 24
		for range 8 {
			if (crc & 0x80000000) != 0 {
				crc = (crc << 1) ^ 0x04C11DB7
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
