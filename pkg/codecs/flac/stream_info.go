// Package flac contains utilities to work with the FLAC codec.
package flac

import (
	"fmt"
)

// StreamInfo is a STREAMINFO metadata block.
// Specification: RFC9639, section 8.2
type StreamInfo struct {
	MinBlockSize uint16
	MaxBlockSize uint16
	MinFrameSize uint32 // 24-bit
	MaxFrameSize uint32 // 24-bit
	SampleRate   uint32 // 20-bit
	ChannelCount uint8  // actual count (1-8)
	BitDepth     uint8  // actual bit depth (4-32)
	TotalSamples uint64 // 36-bit
	MD5          [16]byte
}

// Unmarshal decodes a StreamInfo.
func (s *StreamInfo) Unmarshal(buf []byte) error {
	if len(buf) < 34 {
		return fmt.Errorf("not enough bytes")
	}

	s.MinBlockSize = uint16(buf[0])<<8 | uint16(buf[1])
	s.MaxBlockSize = uint16(buf[2])<<8 | uint16(buf[3])
	s.MinFrameSize = uint32(buf[4])<<16 | uint32(buf[5])<<8 | uint32(buf[6])
	s.MaxFrameSize = uint32(buf[7])<<16 | uint32(buf[8])<<8 | uint32(buf[9])
	s.SampleRate = uint32(buf[10])<<12 | uint32(buf[11])<<4 | uint32(buf[12])>>4
	s.ChannelCount = (buf[12]>>1)&0x7 + 1
	s.BitDepth = ((buf[12]&0x1)<<4 | buf[13]>>4) + 1
	s.TotalSamples = uint64(buf[13]&0xF)<<32 | uint64(buf[14])<<24 |
		uint64(buf[15])<<16 | uint64(buf[16])<<8 | uint64(buf[17])
	copy(s.MD5[:], buf[18:34])

	return nil
}

func (s StreamInfo) marshalSize() int {
	return 34
}

func (s StreamInfo) marshalTo(buf []byte) (int, error) {
	buf[0] = byte(s.MinBlockSize >> 8)
	buf[1] = byte(s.MinBlockSize)
	buf[2] = byte(s.MaxBlockSize >> 8)
	buf[3] = byte(s.MaxBlockSize)
	buf[4] = byte(s.MinFrameSize >> 16)
	buf[5] = byte(s.MinFrameSize >> 8)
	buf[6] = byte(s.MinFrameSize)
	buf[7] = byte(s.MaxFrameSize >> 16)
	buf[8] = byte(s.MaxFrameSize >> 8)
	buf[9] = byte(s.MaxFrameSize)
	buf[10] = byte(s.SampleRate >> 12)
	buf[11] = byte(s.SampleRate >> 4)
	buf[12] = byte(s.SampleRate&0xF)<<4 | (s.ChannelCount-1)<<1 | (s.BitDepth-1)>>4
	buf[13] = (s.BitDepth-1)&0xF<<4 | byte(s.TotalSamples>>32)&0xF
	buf[14] = byte(s.TotalSamples >> 24)
	buf[15] = byte(s.TotalSamples >> 16)
	buf[16] = byte(s.TotalSamples >> 8)
	buf[17] = byte(s.TotalSamples)
	copy(buf[18:34], s.MD5[:])

	return 34, nil
}

// Marshal encodes a StreamInfo.
func (s StreamInfo) Marshal() ([]byte, error) {
	buf := make([]byte, s.marshalSize())
	_, err := s.marshalTo(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
