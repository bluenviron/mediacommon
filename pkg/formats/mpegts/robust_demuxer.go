package mpegts

import (
	"context"
	"errors"
	"io"

	"github.com/asticode/go-astits"
)

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

// this is a wrapper around astits.Demuxer that returns fatal errors only,
// while decode errors are redirected to a callback.
type robustDemuxer struct {
	R             io.Reader
	OnDecodeError func(err error)

	dem *astits.Demuxer
}

func (d *robustDemuxer) initialize() {
	if d.OnDecodeError == nil {
		d.OnDecodeError = func(_ error) {}
	}

	d.dem = astits.NewDemuxer(
		context.Background(),
		&errorWrapper{R: d.R},
		astits.DemuxerOptPacketSize(188))
}

func (d *robustDemuxer) nextData() (*astits.DemuxerData, error) {
	for {
		data, err := d.dem.NextData()
		if err != nil {
			// error from underlying reader or astits.ErrNoMorePackets: this is fatal.
			if errors.Is(err, astits.ErrNoMorePackets) {
				// convert astits.ErrNoMorePackets back into io.EOF
				return nil, io.EOF
			}
			var w *wrappedError
			if errors.As(err, &w) {
				return nil, w.w
			}

			// error from MPEG-TS demuxer: this is not fatal.
			d.OnDecodeError(err)
			continue
		}

		return data, nil
	}
}
