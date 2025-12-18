package eac3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncInfoUnmarshal(t *testing.T) {
	// E-AC-3 frame header example
	// This is a minimal E-AC-3 frame header with:
	// - sync word: 0x0B77
	// - strmtyp: 0 (independent stream)
	// - substreamid: 0
	// - frmsiz: varies
	// - fscod: 0 (48 kHz)
	// - numblkscod: 3 (6 blocks)
	// - acmod: 7 (3/2 - 5 channels)
	// - lfeon: 1 (.1 channel)
	// - bsid: 16 (E-AC-3)

	// Build a test frame:
	// Bits 0-15: sync word (0x0B77)
	// Bits 16-17: strmtyp = 0b00
	// Bits 18-20: substreamid = 0b000
	// Bits 21-31: frmsiz = 0x0FF (255) -> frame size = 512 bytes
	// Bits 32-33: fscod = 0b00 (48 kHz)
	// Bits 34-35: numblkscod = 0b11 (6 blocks)
	// Bits 36-38: acmod = 0b111 (3/2)
	// Bit 39: lfeon = 1
	// Bits 40-44: bsid = 0b10000 (16)

	// Byte 0: 0x0B
	// Byte 1: 0x77
	// Byte 2: 0b00_000_011 = 0x03 (strmtyp=0, substreamid=0, frmsiz high 3 bits=011)
	// Byte 3: 0xFF (frmsiz low 8 bits)
	// Byte 4: 0b00_11_111_1 = 0x3F (fscod=0, numblkscod=3, acmod=7, lfeon=1)
	// Byte 5: 0b10000_xxx = 0x80 (bsid=16, plus 3 bits of dialnorm)

	frame := []byte{
		0x0B, 0x77, // sync word
		0x01, 0xFF, // strmtyp=0, substreamid=0, frmsiz=0x1FF
		0x3F, // fscod=0, numblkscod=3, acmod=7, lfeon=1
		0x80, // bsid=16
		0x00, 0x00, // padding
	}

	var s SyncInfo
	err := s.Unmarshal(frame)
	require.NoError(t, err)

	require.Equal(t, uint8(0), s.Strmtyp)
	require.Equal(t, uint8(0), s.Substreamid)
	require.Equal(t, uint16(0x1FF), s.Frmsiz)
	require.Equal(t, uint8(0), s.Fscod)
	require.Equal(t, uint8(3), s.Numblkscod)
	require.Equal(t, uint8(7), s.Acmod)
	require.Equal(t, true, s.Lfeon)
	require.Equal(t, uint8(16), s.Bsid)

	// Frame size = (0x1FF + 1) * 2 = 1024 bytes
	require.Equal(t, 1024, s.FrameSize())

	// Sample rate = 48000 Hz
	require.Equal(t, 48000, s.SampleRate())

	// Channel count = 5 (3/2) + 1 (LFE) = 6
	require.Equal(t, 6, s.ChannelCount())

	// Num blocks = 6
	require.Equal(t, 6, s.NumBlocks())
}

func TestSyncInfoUnmarshalInvalidSyncWord(t *testing.T) {
	frame := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	var s SyncInfo
	err := s.Unmarshal(frame)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid sync word")
}

func TestSyncInfoUnmarshalTooShort(t *testing.T) {
	frame := []byte{0x0B, 0x77, 0x00}
	var s SyncInfo
	err := s.Unmarshal(frame)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not enough bytes")
}

func TestSyncInfoUnmarshalInvalidBsid(t *testing.T) {
	// Build a frame with bsid = 8 (AC-3, not E-AC-3)
	frame := []byte{
		0x0B, 0x77, // sync word
		0x01, 0xFF, // strmtyp=0, substreamid=0, frmsiz=0x1FF
		0x3F, // fscod=0, numblkscod=3, acmod=7, lfeon=1
		0x40, // bsid=8 (invalid for E-AC-3)
		0x00, 0x00,
	}

	var s SyncInfo
	err := s.Unmarshal(frame)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid bsid for E-AC-3")
}

func TestSyncInfoSampleRateFscod2(t *testing.T) {
	// Test fscod = 0b11 with fscod2
	// When fscod == 0b11, we read fscod2 instead of numblkscod

	// Byte layout for fscod=3, fscod2=0:
	// Bits 32-33: fscod = 0b11
	// Bits 34-35: fscod2 = 0b00 (24 kHz)
	// Bits 36-38: acmod = 0b010 (2/0 stereo)
	// Bit 39: lfeon = 0
	// Bits 40-44: bsid = 0b10000 (16)

	// Byte 4: 0b11_00_010_0 = 0xC4 (fscod=3, fscod2=0, acmod=2, lfeon=0)
	// Byte 5: 0b10000_000 = 0x80 (bsid=16)

	frame := []byte{
		0x0B, 0x77, // sync word
		0x01, 0xFF, // strmtyp=0, substreamid=0, frmsiz=0x1FF
		0xC4, // fscod=3, fscod2=0, acmod=2, lfeon=0
		0x80, // bsid=16
		0x00, 0x00,
	}

	var s SyncInfo
	err := s.Unmarshal(frame)
	require.NoError(t, err)

	require.Equal(t, uint8(3), s.Fscod)
	require.Equal(t, uint8(0), s.Fscod2)
	require.Equal(t, uint8(3), s.Numblkscod) // Implicitly 3 (6 blocks) when fscod==3

	// Sample rate with fscod=3, fscod2=0 should be 24000 Hz
	require.Equal(t, 24000, s.SampleRate())

	// acmod=2 (stereo) = 2 channels, no LFE
	require.Equal(t, 2, s.ChannelCount())
}
