package mpegts

import (
	"fmt"
	"io"
)

// this is needed to make sure that astits.Demuxer receives valid, 188 byte-long, MPEG-TS packets,
// since it uses io.ReadFull which can only read full packets and cannot detect or skip garbage.
// (https://github.com/asticode/go-astits/blob/b0b19247aa31633650c32638fb55f597fa6e2468/packet_buffer.go#L133C1-L133C5)
type preDemuxer struct {
	R             io.Reader
	OnDecodeError func(err error)

	buf1    []byte
	buf1Pos int
	buf2    []byte
	buf2Pos int
}

func (r *preDemuxer) initialize() {
	if r.OnDecodeError == nil {
		r.OnDecodeError = func(_ error) {}
	}

	r.buf1 = make([]byte, 0, 1316)
	r.buf1Pos = 0
	r.buf2 = make([]byte, 188)
	r.buf2Pos = len(r.buf2)
}

func (r *preDemuxer) Read(p []byte) (int, error) {
	if len(r.buf2[r.buf2Pos:]) == 0 {
		n := 0

		for {
			// use buf1 to read multiple packets at once.
			// This is mandatory in case of packet-based connections (UDP)
			// and improves performance in case of stream-based connections.
			if len(r.buf1[r.buf1Pos:]) == 0 {
				n3, err := r.R.Read(r.buf1[:cap(r.buf1)])
				if n3 == 0 && err != nil {
					return 0, err
				}
				r.buf1Pos = 0
				r.buf1 = r.buf1[:n3]
			}

			n2 := copy(r.buf2[n:], r.buf1[r.buf1Pos:])
			r.buf1Pos += n2
			n += n2

			if n != 188 {
				continue
			}

			// skip garbage
			skipped := 0
			for r.buf2[skipped] != 0x47 {
				skipped++
				if skipped == 188 {
					break
				}
			}

			if skipped != 0 {
				r.OnDecodeError(fmt.Errorf("skipped %d bytes", skipped))
				n = copy(r.buf2, r.buf2[skipped:])
				continue
			}

			break
		}

		r.buf2Pos = 0
	}

	n := copy(p, r.buf2[r.buf2Pos:])
	r.buf2Pos += n
	return n, nil
}
