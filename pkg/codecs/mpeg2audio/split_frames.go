package mpeg2audio

import (
	"fmt"
)

// SplitFrames splits a sequence of MPEG-1 or MPEG-2 audio frames.
func SplitFrames(buf []byte) ([][]byte, error) {
	var frames [][]byte

	for {
		var h FrameHeader
		err := h.Unmarshal(buf)
		if err != nil {
			return nil, err
		}

		fl := h.FrameLen()
		if len(buf) < fl {
			return nil, fmt.Errorf("not enough bits")
		}

		frames = append(frames, buf[:fl])
		buf = buf[fl:]

		if len(buf) == 0 {
			break
		}
	}

	return frames, nil
}
