package mpegts

import (
	"fmt"
	"io"

	"github.com/asticode/go-astits"
)

// https://github.com/asticode/go-astits/blob/f593538dc04ea116394690b6cf3c916e16f0195f/data.go#L120
func isPESPayload(i []byte) bool {
	return uint32(i[0])<<16|uint32(i[1])<<8|uint32(i[2]) == 1
}

// https://github.com/asticode/go-astits/blob/f593538dc04ea116394690b6cf3c916e16f0195f/data_pes.go#L150
func hasPESOptionalHeader(streamID uint8) bool {
	return streamID != astits.StreamIDPaddingStream && streamID != astits.StreamIDPrivateStream2
}

func parseClock(b []byte) int64 {
	return int64((((uint64(b[0]) >> 1) & 0x7) << 30) | (uint64(b[1]) << 22) |
		(((uint64(b[2]) >> 1) & 0x7f) << 15) | (uint64(b[3]) << 7) | ((uint64(b[4]) >> 1) & 0x7f))
}

type ptsInterceptor struct {
	R io.Reader

	hasLastPTS bool
	lastPTS    int64
}

func (r *ptsInterceptor) Read(p []byte) (int, error) {
	if len(p) != packetSize {
		return 0, fmt.Errorf("unsupported")
	}

	_, err := io.ReadFull(r.R, p)
	if err != nil {
		return 0, err
	}

	if p[0] == syncByte {
		hasPayload := (p[3]&0x10 > 0)
		payloadStartIndicator := (p[1]&0x40 > 0)

		if payloadStartIndicator && hasPayload {
			payloadPos := 4
			hasAdaptationField := (p[3]&0x20 > 0)

			if hasAdaptationField {
				adaptationFieldLength := int(p[4])
				payloadPos += 1 + adaptationFieldLength
			}

			if payloadPos <= (packetSize - 14) {
				payload := p[payloadPos:]

				if isPESPayload(payload) {
					streamID := payload[3]

					if hasPESOptionalHeader(streamID) {
						ptsDTSIndicator := (payload[7] >> 6) & 0x03

						if ptsDTSIndicator == astits.PTSDTSIndicatorOnlyPTS ||
							ptsDTSIndicator == astits.PTSDTSIndicatorBothPresent {
							r.hasLastPTS = true
							r.lastPTS = parseClock(payload[9:14])
						}
					}
				}
			}
		}
	}

	return packetSize, nil
}
