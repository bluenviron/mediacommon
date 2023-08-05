package mpeg1audio

import (
	"fmt"
)

// http://www.mp3-tech.org/programmer/frame_header.html
var bitrates = map[uint8]map[uint8]int{
	2: {
		0b1:    32000,
		0b10:   48000,
		0b11:   56000,
		0b100:  64000,
		0b101:  80000,
		0b110:  96000,
		0b111:  112000,
		0b1000: 128000,
		0b1001: 160000,
		0b1010: 192000,
		0b1011: 224000,
		0b1100: 256000,
		0b1101: 320000,
		0b1110: 384000,
	},
	3: {
		0b1:    32000,
		0b10:   40000,
		0b11:   48000,
		0b100:  56000,
		0b101:  64000,
		0b110:  80000,
		0b111:  96000,
		0b1000: 112000,
		0b1001: 128000,
		0b1010: 160000,
		0b1011: 192000,
		0b1100: 224000,
		0b1101: 256000,
		0b1110: 320000,
	},
}

var sampleRates = map[uint8]int{
	0:    44100,
	0b1:  48000,
	0b10: 32000,
}

// ChannelMode is a channel mode of a MPEG-1/2 audio frame.
type ChannelMode int

// standard channel modes.
const (
	ChannelModeStereo      ChannelMode = 0
	ChannelModeJointStereo ChannelMode = 1
	ChannelModeDualChannel ChannelMode = 2
	ChannelModeMono        ChannelMode = 3
)

// FrameHeader is the header of a MPEG-1/2 audio frame.
// Specification: ISO 11172-3, 2.4.1.3
type FrameHeader struct {
	MPEG2       bool
	Layer       uint8
	Bitrate     int
	SampleRate  int
	Padding     bool
	ChannelMode ChannelMode
}

// Unmarshal decodes a FrameHeader.
func (h *FrameHeader) Unmarshal(buf []byte) error {
	if len(buf) < 5 {
		return fmt.Errorf("not enough bytes")
	}

	syncWord := uint16(buf[0])<<4 | uint16(buf[1])>>4
	if syncWord != 0x0FFF {
		return fmt.Errorf("sync word not found: %x", syncWord)
	}

	h.MPEG2 = ((buf[1] >> 3) & 0x01) == 0
	h.Layer = 4 - ((buf[1] >> 1) & 0b11)

	switch {
	case !h.MPEG2 && h.Layer == 2:
	case !h.MPEG2 && h.Layer == 3:
	default:
		return fmt.Errorf("unsupported MPEG version or layer: %v %v", h.MPEG2, h.Layer)
	}

	bitrateIndex := buf[2] >> 4
	var ok bool
	h.Bitrate, ok = bitrates[h.Layer][bitrateIndex]
	if !ok {
		return fmt.Errorf("invalid bitrate")
	}

	sampleRateIndex := (buf[2] >> 2) & 0b11
	h.SampleRate, ok = sampleRates[sampleRateIndex]
	if !ok {
		return fmt.Errorf("invalid sample rate")
	}

	h.Padding = ((buf[2] >> 1) & 0b1) != 0
	h.ChannelMode = ChannelMode(buf[3] >> 6)

	return nil
}

// FrameLen returns the length of the frame associated to the header.
func (h FrameHeader) FrameLen() int {
	if h.Padding {
		return (144 * h.Bitrate / h.SampleRate) + 1
	}
	return (144 * h.Bitrate / h.SampleRate)
}

// SampleCount returns the number of samples contained into the frame.
func (h FrameHeader) SampleCount() int {
	/*
			 MPEG-1:
			 * layer 1: 384
			 * layer 2: 1152
			 * layer 3: 1152

		     MPEG-2:
			 * layer 1: 384
			 * layer 2: 1152
			 * layer 3: 576
	*/
	return 1152
}
