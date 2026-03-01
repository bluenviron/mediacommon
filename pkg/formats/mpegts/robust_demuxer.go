package mpegts

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/asticode/go-astits"
)

// this is a wrapper around astits.Demuxer with the following differences:
// - non-fatal errors are not returned, but routed to OnDecodeError.
// - last PTS is intercepted and returned.

const (
	packetSize = 188
	syncByte   = 0x47
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

type wrappedError struct {
	w error
}

func (w wrappedError) Error() string {
	return w.w.Error()
}

type errorWrapper struct {
	R io.Reader
}

func (r *errorWrapper) Read(p []byte) (int, error) {
	n, err := r.R.Read(p)
	if err != nil {
		// keep io.EOF and io.ErrUnexpectedEOF untouched
		// since they are needed by astits to handle the last packet of a stream.
		// they are transformed into astits.ErrNoMorePackets.
		// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/packet_buffer.go#L131
		if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			err = &wrappedError{w: err}
		}
	}
	return n, err
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

type robustDemuxerData struct {
	lastPTS *int64
	PID     uint16
	PMT     *astits.PMTData
	PES     *astits.PESData
}

type robustDemuxer struct {
	R             io.Reader
	OnDecodeError func(err error)

	dem            *astits.Demuxer
	ptsInterceptor *ptsInterceptor
}

func (d *robustDemuxer) initialize() {
	if d.OnDecodeError == nil {
		d.OnDecodeError = func(_ error) {}
	}

	errorWrapper := &errorWrapper{R: d.R}
	d.ptsInterceptor = &ptsInterceptor{R: errorWrapper}

	d.dem = astits.NewDemuxer(
		context.Background(),
		d.ptsInterceptor,
		astits.DemuxerOptPacketSize(packetSize),
	)
}

func (d *robustDemuxer) nextData() (*robustDemuxerData, error) {
	for {
		data, err := d.dem.NextData()
		if err != nil {
			// astits.ErrNoMorePackets / io.EOF: this is fatal.
			if errors.Is(err, astits.ErrNoMorePackets) {
				return nil, io.EOF // convert back to io.EOF.
			}

			// error from underlying reader: this is fatal.
			var w *wrappedError
			if errors.As(err, &w) {
				return nil, w.w
			}

			// error from MPEG-TS demuxer: this is not fatal.
			d.OnDecodeError(err)
			continue
		}

		return &robustDemuxerData{
			lastPTS: func() *int64 {
				if d.ptsInterceptor.hasLastPTS {
					pts := d.ptsInterceptor.lastPTS
					return &pts
				}
				return nil
			}(),
			PID: data.PID,
			PMT: data.PMT,
			PES: data.PES,
		}, nil
	}
}
