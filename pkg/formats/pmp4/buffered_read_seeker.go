package pmp4

import (
	"errors"
	"io"
)

// bufferedReadSeeker is a read-only, buffered io.ReadSeeker.
// go-mp4 decodes boxes with one Read call per byte; when reading
// from a file, each one is a system call.
type bufferedReadSeeker struct {
	r      io.ReadSeeker
	buf    []byte
	bufPos int64 // position of buf[0]
	bufLen int   // valid bytes in buf
	pos    int64
}

func newBufferedReadSeeker(r io.ReadSeeker, size int) *bufferedReadSeeker {
	return &bufferedReadSeeker{
		r:   r,
		buf: make([]byte, size),
	}
}

// Read implements io.Reader.
func (b *bufferedReadSeeker) Read(p []byte) (int, error) {
	if b.pos < b.bufPos || b.pos >= b.bufPos+int64(b.bufLen) {
		_, err := b.r.Seek(b.pos, io.SeekStart)
		if err != nil {
			return 0, err
		}

		n, err := io.ReadFull(b.r, b.buf)
		b.bufPos = b.pos
		b.bufLen = n

		if n == 0 {
			return 0, err
		}
	}

	n := copy(p, b.buf[b.pos-b.bufPos:b.bufLen])
	b.pos += int64(n)
	return n, nil
}

// Seek implements io.Seeker.
func (b *bufferedReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:

	case io.SeekCurrent:
		offset += b.pos

	case io.SeekEnd:
		end, err := b.r.Seek(0, io.SeekEnd)
		if err != nil {
			return 0, err
		}
		offset += end

	default:
		return 0, errors.New("invalid whence")
	}

	if offset < 0 {
		return 0, errors.New("negative position")
	}

	b.pos = offset
	return offset, nil
}
