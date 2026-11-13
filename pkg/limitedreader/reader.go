// Package limitedreader contains a reader that limits the amount of bytes that can be read.
package limitedreader

import (
	"fmt"
	"io"
)

// ErrSizeExceeded is returned when the limit is exceeded.
var ErrSizeExceeded = fmt.Errorf("size exceeds maximum allowed")

// Reader is like io.LimitedReader, but returns a dedicated error if the limit is exceeded.
type Reader struct {
	// wrapped reader.
	R io.Reader

	// maximum number of remaining bytes. It can be changed at any time.
	N int64
}

// Read implements io.Reader.
func (l *Reader) Read(p []byte) (int, error) {
	if l.N <= 0 {
		return 0, ErrSizeExceeded
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err := l.R.Read(p)
	l.N -= int64(n)
	return n, err
}
