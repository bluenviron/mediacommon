package mpeg4audio

import "fmt"

// AudioSyncStream is an AudioSyncStream.
// Specification: ISO 14496-3, Table 1.36
type AudioSyncStream struct {
	AudioMuxElements [][]byte
}

// Unmarshal decodes an AudioSyncStream.
func (s *AudioSyncStream) Unmarshal(buf []byte) error {
	for {
		if len(buf) < 3 {
			return fmt.Errorf("buffer is too short")
		}

		syncWord := (uint16(buf[0])<<8 | uint16(buf[1])) >> 5
		if syncWord != 0x2B7 {
			return fmt.Errorf("invalid syncword")
		}

		le := (uint16(buf[1])<<8 | uint16(buf[2])) & 0b1111111111111

		if len(buf) < (3 + int(le)) {
			return fmt.Errorf("buffer is too short")
		}

		raw := buf[3 : 3+le]
		buf = buf[3+le:]

		s.AudioMuxElements = append(s.AudioMuxElements, raw)

		if len(buf) == 0 {
			return nil
		}
	}
}

func (s AudioSyncStream) marshalSize() int {
	n := 3 * len(s.AudioMuxElements)
	for _, el := range s.AudioMuxElements {
		n += len(el)
	}
	return n
}

// Marshal encodes an AudioSyncStream.
func (s AudioSyncStream) Marshal() ([]byte, error) {
	buf := make([]byte, s.marshalSize())
	n := 0

	for _, el := range s.AudioMuxElements {
		shiftedSyncWord := 0x2B7 << 5
		le := len(el)
		buf[n] = byte(shiftedSyncWord >> 8)
		buf[n+1] = byte(shiftedSyncWord) | byte(le>>8)
		buf[n+2] = byte(le)
		n += 3

		n += copy(buf[n:], el)
	}

	return buf, nil
}
