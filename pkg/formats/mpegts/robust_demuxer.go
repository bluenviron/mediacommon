package mpegts

import (
	"errors"
	"io"

	"github.com/asticode/go-astits"
)

const (
	packetSize = 188
	syncByte   = 0x47
)

type robustDemuxerData struct {
	lastPTS *int64
	PID     uint16
	PMT     *astits.PMTData
	PES     *astits.PESData
	Section []byte
}

type robustDemuxer struct {
	ptsInterceptor *ptsInterceptor
	astitsDem      *astits.Demuxer
	OnDecodeError  func(err error)

	sectionPIDs     map[uint16]*sectionAssembler
	pendingSections []*robustDemuxerData
	releasable      int
	pendingErr      error
}

func (d *robustDemuxer) initialize() {
	if d.OnDecodeError == nil {
		d.OnDecodeError = func(_ error) {}
	}

	d.sectionPIDs = make(map[uint16]*sectionAssembler)
}

// skipPacket prevents astits from handling packets of section PIDs, which are handled by onPacket.
func (d *robustDemuxer) skipPacket(p *astits.Packet) bool {
	_, ok := d.sectionPIDs[p.Header.PID]
	return ok
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

		data, err := d.astitsDem.NextData()
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
