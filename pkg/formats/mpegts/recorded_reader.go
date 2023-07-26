package mpegts

import (
	"errors"
	"io"
)

const (
	recordedReaderMaxSize = 1 * 1024 * 1024
)

type recordedReader struct {
	r    io.Reader
	buf  []byte
	size int
}

func (r *recordedReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)

	if (r.size + n) > recordedReaderMaxSize {
		return 0, errors.New("max buffer size exceeded")
	}

	r.buf = append(r.buf, p...)
	r.size += n
	return n, err
}
