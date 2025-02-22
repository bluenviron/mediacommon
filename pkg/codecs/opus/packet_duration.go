package opus

import (
	"time"
)

var frameSizes = [32]int{
	480, 960, 1920, 2880, // Silk NB
	480, 960, 1920, 2880, // Silk MB
	480, 960, 1920, 2880, // Silk WB
	480, 960, // Hybrid SWB
	480, 960, // Hybrid FB
	120, 240, 480, 960, // CELT NB
	120, 240, 480, 960, // CELT NB
	120, 240, 480, 960, // CELT NB
	120, 240, 480, 960, // CELT NB
}

// PacketDuration returns the duration of an Opus packet.
//
// Deprecated: replaced by PacketDuration2
func PacketDuration(pkt []byte) time.Duration {
	return (time.Duration(PacketDuration2(pkt)) * time.Second) / 48000
}

// PacketDuration2 returns the duration of an Opus packet, in 1/48000 seconds.
// Specification: RFC6716, 3.1
func PacketDuration2(pkt []byte) int64 {
	if len(pkt) == 0 {
		return 0
	}

	frameDuration := int64(frameSizes[pkt[0]>>3])

	frameCount := int64(0)
	switch pkt[0] & 3 {
	case 0:
		frameCount = 1
	case 1:
		frameCount = 2
	case 2:
		frameCount = 2
	case 3:
		if len(pkt) < 2 {
			return 0
		}
		frameCount = int64(pkt[1] & 63)
	}

	return frameDuration * frameCount
}
