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
// - SCTE-35 sections are intercepted and returned.

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
	R        io.Reader
	OnPacket func(p []byte)

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
		r.OnPacket(p)

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
	Section []byte
}

type robustDemuxer struct {
	R             io.Reader
	OnDecodeError func(err error)

	dem             *astits.Demuxer
	ptsInterceptor  *ptsInterceptor
	sectionPIDs     map[uint16]*sectionAssembler
	pendingSections []*robustDemuxerData
	releasable      int
	pendingErr      error
}

func (d *robustDemuxer) initialize() {
	if d.OnDecodeError == nil {
		d.OnDecodeError = func(_ error) {}
	}

	errorWrapper := &errorWrapper{R: d.R}
	d.ptsInterceptor = &ptsInterceptor{
		R:        errorWrapper,
		OnPacket: d.onPacket,
	}
	d.sectionPIDs = make(map[uint16]*sectionAssembler)

	d.dem = astits.NewDemuxer(
		context.Background(),
		d.ptsInterceptor,
		astits.DemuxerOptPacketSize(packetSize),
		// packets of section PIDs are handled by onPacket.
		astits.DemuxerOptPacketSkipper(func(p *astits.Packet) bool {
			_, ok := d.sectionPIDs[p.Header.PID]
			return ok
		}),
	)
}

func (d *robustDemuxer) lastPTS() *int64 {
	if d.ptsInterceptor.hasLastPTS {
		pts := d.ptsInterceptor.lastPTS
		return &pts
	}
	return nil
}

func (d *robustDemuxer) onPacket(p []byte) {
	pid := uint16(p[1]&0x1F)<<8 | uint16(p[2])

	asm, ok := d.sectionPIDs[pid]
	if !ok {
		return
	}

	sections, err := asm.push(p)
	if err != nil {
		d.OnDecodeError(err)
	}

	for _, sec := range sections {
		d.pendingSections = append(d.pendingSections, &robustDemuxerData{
			lastPTS: d.lastPTS(),
			PID:     pid,
			Section: sec,
		})
	}
}

func (d *robustDemuxer) updateSectionPIDs(pmt *astits.PMTData) {
	sectionPIDs := make(map[uint16]*sectionAssembler)

	for _, es := range pmt.ElementaryStreams {
		if es.StreamType == astits.StreamTypeSCTE35 {
			if asm, ok := d.sectionPIDs[es.ElementaryPID]; ok {
				sectionPIDs[es.ElementaryPID] = asm
			} else {
				asm = &sectionAssembler{}
				asm.initialize()
				sectionPIDs[es.ElementaryPID] = asm
			}
		}
	}

	d.sectionPIDs = sectionPIDs
}

func (d *robustDemuxer) nextData() (*robustDemuxerData, error) {
	for {
		if d.releasable != 0 {
			data := d.pendingSections[0]
			d.pendingSections = d.pendingSections[1:]
			d.releasable--
			return data, nil
		}

		if d.pendingErr != nil {
			return nil, d.pendingErr
		}

		data, err := d.dem.NextData()
		if err != nil {
			// astits.ErrNoMorePackets / io.EOF: this is fatal.
			if errors.Is(err, astits.ErrNoMorePackets) {
				err = io.EOF // convert back to io.EOF.
			} else if w, ok := errors.AsType[*wrappedError](err); ok {
				// error from underlying reader: this is fatal.
				err = w.w
			} else {
				// error from MPEG-TS demuxer: this is not fatal.
				d.OnDecodeError(err)
				continue
			}

			// return sections read before the error first.
			if len(d.pendingSections) != 0 {
				d.pendingErr = err
				d.releasable = len(d.pendingSections)
				continue
			}

			return nil, err
		}

		if data.PMT != nil {
			d.updateSectionPIDs(data.PMT)
		}

		// astits returns a PES when the next one begins, therefore sections
		// may complete before the PES that precedes them.
		// Release sections after the PES, in order to preserve stream order.
		if data.PES != nil {
			d.releasable = len(d.pendingSections)
		}

		return &robustDemuxerData{
			lastPTS: d.lastPTS(),
			PID:     data.PID,
			PMT:     data.PMT,
			PES:     data.PES,
		}, nil
	}
}
