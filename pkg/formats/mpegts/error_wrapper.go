package mpegts

import (
	"errors"
	"io"
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
