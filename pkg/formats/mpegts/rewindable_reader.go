package mpegts

import (
	"errors"
	"io"
)

const (
	maxRecordedSize = 1 * 1024 * 1024
)

// rewindableReader is a reader that can be (and must be) rewinded once.
type rewindableReader struct {
	R io.Reader

	entries  [][]byte
	size     int
	rewinded bool
}

// Read implements io.Reader.
func (r *rewindableReader) Read(p []byte) (int, error) {
	if !r.rewinded {
		n, err := r.R.Read(p)

		if (r.size + n) > maxRecordedSize {
			return 0, errors.New("max recorded size exceeded")
		}

		entry := make([]byte, n)
		copy(entry, p[:n])
		r.entries = append(r.entries, entry)
		r.size += n
		return n, err
	}

	if r.entries != nil {
		entry := r.entries[0]
		n := copy(p, entry)

		if n != len(entry) {
			r.entries[0] = entry[n:]
		} else {
			r.entries = r.entries[1:]

			if len(r.entries) == 0 {
				r.entries = nil // release entries from memory
			}
		}

		return n, nil
	}

	return r.R.Read(p)
}

// Rewind rewinds the reader. This can be (and must be) called once.
func (r *rewindableReader) Rewind() {
	r.rewinded = true
}
