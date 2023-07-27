package mpegts

import (
	"io"
)

type playbackReader struct {
	r   io.Reader
	buf []byte
}

func (r *playbackReader) Read(p []byte) (int, error) {
	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]

		if len(r.buf) == 0 { // release buf from memory
			r.buf = nil
		}

		return n, nil
	}

	return r.r.Read(p)
}
