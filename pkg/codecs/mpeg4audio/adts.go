package mpeg4audio

import (
	"fmt"
)

// ADTSPacket is an ADTS packet.
// Specification: ISO 14496-3, Table 1.A.5
type ADTSPacket struct {
	Type         ObjectType
	SampleRate   int
	ChannelCount int
	AU           []byte
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

		channelConfig := ((buf[pos+2] & 0x01) << 2) | ((buf[pos+3] >> 6) & 0x03)
		switch {
		case channelConfig >= 1 && channelConfig <= 6:
			pkt.ChannelCount = int(channelConfig)

		case channelConfig == 7:
			pkt.ChannelCount = 8

		case channelConfig == 0:
			// Channel configuration 0 means the channel layout is defined either by:
			// 1. An explicit PCE at the start of the raw_data_block, or
			// 2. Implicit ordering based on the syntactic elements present
			//
			// FFmpeg typically uses implicit ordering without PCE, so we try both.
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

			// Try to parse PCE first, fall back to counting elements
			pce, err := ParsePCEFromRawDataBlock(pkt.AU)
			if err == nil {
				pkt.ChannelCount = pce.ChannelCount
			} else {
				// No PCE - count channels from syntactic elements
				channelCount, err := CountChannelsFromRawDataBlock(pkt.AU)
				if err != nil {
					return fmt.Errorf("channel_config=0: %w", err)
				}
				pkt.ChannelCount = channelCount
			}

			pos += 7 + frameLen
			*ps = append(*ps, pkt)

			if (bl - pos) == 0 {
				return nil
			}
			continue

		default:
			// Channel configs 8-15 are reserved.
			return fmt.Errorf("unsupported channel configuration: %d", channelConfig)
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

		var channelConfig int
		switch {
		case pkt.ChannelCount >= 1 && pkt.ChannelCount <= 6:
			channelConfig = pkt.ChannelCount

		case pkt.ChannelCount == 8:
			channelConfig = 7

		default:
			return nil, fmt.Errorf("invalid channel count (%d)", pkt.ChannelCount)
		}

		frameLen := len(pkt.AU) + 7

		fullness := 0x07FF // like ffmpeg does

		buf[pos+0] = 0xFF
		buf[pos+1] = 0xF1
		buf[pos+2] = uint8((int(pkt.Type-1) << 6) | (sampleRateIndex << 2) | ((channelConfig >> 2) & 0x01))
		buf[pos+3] = uint8((channelConfig&0x03)<<6 | (frameLen>>11)&0x03)
		buf[pos+4] = uint8((frameLen >> 3) & 0xFF)
		buf[pos+5] = uint8((frameLen&0x07)<<5 | ((fullness >> 6) & 0x1F))
		buf[pos+6] = uint8((fullness & 0x3F) << 2)
		pos += 7

		pos += copy(buf[pos:], pkt.AU)
	}

	return buf, nil
}
