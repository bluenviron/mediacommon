package mpeg4audio

import (
	"fmt"
)

// ADTSPacket is an ADTS packet.
// Specification: ISO 14496-3, Table 1.A.5
type ADTSPacket struct {
	Type       ObjectType
	SampleRate int

	// 0: channel layout defined by a PCE at the start of raw_data_block
	// 1-6: channel count is equal to channel configuration
	// 7: channel count is 8
	// 8-15: reserved
	ChannelConfig uint8

	// Deprecated: replaced by ChannelConfig
	ChannelCount int

	AU []byte
}

// ADTSPackets is a group of ADTS packets.
type ADTSPackets []*ADTSPacket

// Unmarshal decodes an ADTS stream into ADTS packets.
func (ps *ADTSPackets) Unmarshal(buf []byte) error {
	bl := len(buf)
	pos := 0

	for {
		if (bl - pos) < 8 {
			return fmt.Errorf("invalid length")
		}

		syncWord := (uint16(buf[pos]) << 4) | (uint16(buf[pos+1]) >> 4)
		if syncWord != 0xfff {
			return fmt.Errorf("invalid syncword")
		}

		protectionAbsent := buf[pos+1] & 0x01
		if protectionAbsent != 1 {
			return fmt.Errorf("CRC is not supported")
		}

		pkt := &ADTSPacket{}

		// ADTS profile (2 bits) maps to ObjectType = profile + 1
		// All profiles 0-3 are valid per ISO 14496-3
		pkt.Type = ObjectType((buf[pos+2] >> 6) + 1)

		sampleRateIndex := (buf[pos+2] >> 2) & 0x0F
		switch {
		case sampleRateIndex <= 12:
			pkt.SampleRate = sampleRates[sampleRateIndex]

		default:
			return fmt.Errorf("invalid sample rate index: %d", sampleRateIndex)
		}

		pkt.ChannelConfig = ((buf[pos+2] & 0x01) << 2) | ((buf[pos+3] >> 6) & 0x03)
		switch {
		case pkt.ChannelConfig >= 1 && pkt.ChannelConfig <= 6:
			pkt.ChannelCount = int(pkt.ChannelConfig)

		case pkt.ChannelConfig == 7:
			pkt.ChannelCount = 8

		case pkt.ChannelConfig == 0:
			// Channel configuration 0 means the channel layout is defined by a
			// Program Config Element (PCE), which may be present either:
			// 1. Within GASpecificConfig in the AudioSpecificConfig, or
			// 2. At the start of raw_data_block in each access unit
			//
			// We preserve the original value (0) to allow re-encoding the packet
			// in its original form. Callers needing the actual channel count
			// should use ParsePCEFromRawDataBlock or CountChannelsFromRawDataBlock
			// on the access unit.
			pkt.ChannelCount = 0

		default:
			// Channel configs 8-15 are reserved.
			return fmt.Errorf("unsupported channel configuration: %d", pkt.ChannelConfig)
		}

		frameLen := int(((uint16(buf[pos+3])&0x03)<<11)|
			(uint16(buf[pos+4])<<3)|
			((uint16(buf[pos+5])>>5)&0x07)) - 7

		if frameLen <= 0 {
			return fmt.Errorf("invalid FrameLen")
		}

		if frameLen > MaxAccessUnitSize {
			return fmt.Errorf("access unit size (%d) is too big, maximum is %d", frameLen, MaxAccessUnitSize)
		}

		frameCount := buf[pos+6] & 0x03
		if frameCount != 0 {
			return fmt.Errorf("frame count greater than 1 is not supported")
		}

		if len(buf[pos+7:]) < frameLen {
			return fmt.Errorf("invalid frame length")
		}

		pkt.AU = buf[pos+7 : pos+7+frameLen]
		pos += 7 + frameLen

		*ps = append(*ps, pkt)

		if (bl - pos) == 0 {
			break
		}
	}

	return nil
}

func (ps ADTSPackets) marshalSize() int {
	n := 0
	for _, pkt := range ps {
		n += 7 + len(pkt.AU)
	}
	return n
}

// Marshal encodes ADTS packets into an ADTS stream.
func (ps ADTSPackets) Marshal() ([]byte, error) {
	buf := make([]byte, ps.marshalSize())
	pos := 0

	for _, pkt := range ps {
		// ADTS only supports ObjectType 1-4 (profile 0-3)
		if pkt.Type < 1 || pkt.Type > 4 {
			return nil, fmt.Errorf("ADTS only supports ObjectType 1-4, got %d", pkt.Type)
		}

		sampleRateIndex, ok := reverseSampleRates[pkt.SampleRate]
		if !ok {
			return nil, fmt.Errorf("invalid sample rate: %d", pkt.SampleRate)
		}

		if pkt.ChannelConfig == 0 {
			switch {
			case pkt.ChannelCount == 0:
				pkt.ChannelConfig = 0

			case pkt.ChannelCount >= 1 && pkt.ChannelCount <= 6:
				pkt.ChannelConfig = uint8(pkt.ChannelCount)

			case pkt.ChannelCount == 8:
				pkt.ChannelConfig = 7

			default:
				return nil, fmt.Errorf("invalid channel count (%d)", pkt.ChannelCount)
			}
		}

		frameLen := len(pkt.AU) + 7

		const fullness = 0x07FF // like ffmpeg does

		buf[pos+0] = 0xFF
		buf[pos+1] = 0xF1
		buf[pos+2] = byte(pkt.Type-1)<<6 | byte(sampleRateIndex)<<2 | (pkt.ChannelConfig >> 2)
		buf[pos+3] = (pkt.ChannelConfig&0x03)<<6 | byte(frameLen>>11)
		buf[pos+4] = byte(frameLen >> 3)
		buf[pos+5] = byte(frameLen&0x07)<<5 | byte(fullness>>6)
		buf[pos+6] = byte(fullness&0x3F) << 2
		pos += 7

		pos += copy(buf[pos:], pkt.AU)
	}

	return buf, nil
}
