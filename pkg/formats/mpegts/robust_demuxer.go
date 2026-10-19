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
}

type robustDemuxer struct {
	ptsInterceptor *ptsInterceptor
	astitsDem      *astits.Demuxer
	OnDecodeError  func(err error)
}

func (d *robustDemuxer) initialize() {
	if d.OnDecodeError == nil {
		d.OnDecodeError = func(_ error) {}
	}
}

func (d *robustDemuxer) nextData() (*robustDemuxerData, error) {
	for {
		data, err := d.astitsDem.NextData()
		if err != nil {
			// astits.ErrNoMorePackets / io.EOF: this is fatal.
			if errors.Is(err, astits.ErrNoMorePackets) {
				return nil, io.EOF // convert back to io.EOF.
			}

			// error from underlying reader: this is fatal.
			if w, ok := errors.AsType[*wrappedError](err); ok {
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
