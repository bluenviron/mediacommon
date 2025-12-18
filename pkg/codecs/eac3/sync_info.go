package eac3

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

// E-AC-3 sample rate table
// Index corresponds to fscod value
var sampleRates = []int{48000, 44100, 32000}

// E-AC-3 sample rate table for fscod2 (when fscod == 0b11)
var sampleRates2 = []int{24000, 22050, 16000}

// E-AC-3 channel count table based on acmod
// acmod: audio coding mode
var acmodChannels = []int{
	2, // 0b000: 1+1 (dual mono)
	1, // 0b001: 1/0 (mono)
	2, // 0b010: 2/0 (stereo)
	3, // 0b011: 3/0
	3, // 0b100: 2/1
	4, // 0b101: 3/1
	4, // 0b110: 2/2
	5, // 0b111: 3/2
}

// SyncInfo is E-AC-3 synchronization information.
// Specification: ETSI TS 102 366 V1.4.1, Annex E
type SyncInfo struct {
	// Strmtyp: stream type (0=independent, 1=dependent, 2=independent + conversion, 3=reserved)
	Strmtyp uint8
	// Substreamid: substream identification
	Substreamid uint8
	// Frmsiz: frame size in 16-bit words minus 1
	Frmsiz uint16
	// Fscod: sample rate code
	Fscod uint8
	// Fscod2: sample rate code 2 (used when fscod == 0b11)
	Fscod2 uint8
	// Numblkscod: number of audio blocks
	Numblkscod uint8
	// Acmod: audio coding mode
	Acmod uint8
	// Lfeon: LFE channel on
	Lfeon bool
	// Bsid: bitstream identification (16 for E-AC-3)
	Bsid uint8
}

// Unmarshal decodes an E-AC-3 SyncInfo from a frame.
func (s *SyncInfo) Unmarshal(frame []byte) error {
	if len(frame) < 8 {
		return fmt.Errorf("not enough bytes")
	}

	// Check sync word
	if frame[0] != 0x0B || frame[1] != 0x77 {
		return fmt.Errorf("invalid sync word")
	}

	// Parse E-AC-3 BSI (Bit Stream Information)
	// ETSI TS 102 366, Annex E.1.2

	pos := 16 // Start after sync word (16 bits)

	// strmtyp (2 bits)
	s.Strmtyp = uint8(bits.ReadBitsUnsafe(frame, &pos, 2))

	// substreamid (3 bits)
	s.Substreamid = uint8(bits.ReadBitsUnsafe(frame, &pos, 3))

	// frmsiz (11 bits) - frame size in 16-bit words minus 1
	s.Frmsiz = uint16(bits.ReadBitsUnsafe(frame, &pos, 11))

	// fscod (2 bits)
	s.Fscod = uint8(bits.ReadBitsUnsafe(frame, &pos, 2))

	if s.Fscod == 0b11 {
		// fscod2 (2 bits) - only present when fscod == 0b11
		s.Fscod2 = uint8(bits.ReadBitsUnsafe(frame, &pos, 2))
		if s.Fscod2 == 0b11 {
			return fmt.Errorf("invalid fscod2")
		}
		// numblkscod is implicitly 0b11 (6 blocks) when fscod == 0b11
		s.Numblkscod = 0b11
	} else {
		// numblkscod (2 bits)
		s.Numblkscod = uint8(bits.ReadBitsUnsafe(frame, &pos, 2))
	}

	// acmod (3 bits)
	s.Acmod = uint8(bits.ReadBitsUnsafe(frame, &pos, 3))

	// lfeon (1 bit)
	s.Lfeon = bits.ReadFlagUnsafe(frame, &pos)

	// bsid (5 bits) - should be 16 for E-AC-3
	s.Bsid = uint8(bits.ReadBitsUnsafe(frame, &pos, 5))
	if s.Bsid != 16 {
		return fmt.Errorf("invalid bsid for E-AC-3: %d (expected 16)", s.Bsid)
	}

	return nil
}

// FrameSize returns the frame size in bytes.
func (s SyncInfo) FrameSize() int {
	// frmsiz is frame size in 16-bit words minus 1
	return (int(s.Frmsiz) + 1) * 2
}

// SampleRate returns the sample rate in Hz.
func (s SyncInfo) SampleRate() int {
	if s.Fscod == 0b11 {
		if int(s.Fscod2) < len(sampleRates2) {
			return sampleRates2[s.Fscod2]
		}
		return 0
	}
	if int(s.Fscod) < len(sampleRates) {
		return sampleRates[s.Fscod]
	}
	return 0
}

// ChannelCount returns the number of audio channels.
func (s SyncInfo) ChannelCount() int {
	channels := 0
	if int(s.Acmod) < len(acmodChannels) {
		channels = acmodChannels[s.Acmod]
	}
	if s.Lfeon {
		channels++
	}
	return channels
}

// NumBlocks returns the number of audio blocks per frame.
// Each block represents 256 samples.
func (s SyncInfo) NumBlocks() int {
	switch s.Numblkscod {
	case 0:
		return 1
	case 1:
		return 2
	case 2:
		return 3
	default:
		return 6
	}
}
