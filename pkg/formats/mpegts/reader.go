package mpegts

import (
	"context"
	"fmt"
	"io"

	"github.com/asticode/go-astits"

	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

// ReaderOnDecodeErrorFunc is the prototype of the callback passed to OnDecodeError.
type ReaderOnDecodeErrorFunc func(err error)

// ReaderOnDataH26xFunc is the prototype of the callback passed to OnDataH26x.
type ReaderOnDataH26xFunc func(pts int64, dts int64, au [][]byte) error

// ReaderOnDataMPEG4AudioFunc is the prototype of the callback passed to OnDataMPEG4Audio.
type ReaderOnDataMPEG4AudioFunc func(pts int64, dts int64, aus [][]byte) error

// ReaderOnDataOpusFunc is the prototype of the callback passed to OnDataOpus.
type ReaderOnDataOpusFunc func(pts int64, dts int64, packets [][]byte) error

func findPMT(dem *astits.Demuxer) (*astits.PMTData, error) {
	for {
		data, err := dem.NextData()
		if err != nil {
			return nil, err
		}

		if data.PMT != nil {
			return data.PMT, nil
		}
	}
}

// Reader is a MPEG-TS reader.
type Reader struct {
	tracks        []*Track
	dem           *astits.Demuxer
	onDecodeError ReaderOnDecodeErrorFunc
	onData        map[uint16]func(int64, int64, []byte) error
}

// NewReader allocates a Reader.
func NewReader(br io.Reader) (*Reader, error) {
	rr := &recordReader{r: br}

	dem := astits.NewDemuxer(
		context.Background(),
		rr,
		astits.DemuxerOptPacketSize(188))

	pmt, err := findPMT(dem)
	if err != nil {
		return nil, err
	}

	var tracks []*Track //nolint:prealloc

	for _, es := range pmt.ElementaryStreams {
		var track Track
		err := track.unmarshal(dem, es)
		if err != nil {
			if err == errUnsupportedTrack {
				continue
			}
			return nil, err
		}

		tracks = append(tracks, &track)
	}

	if tracks == nil {
		return nil, fmt.Errorf("no tracks found")
	}

	// rewind demuxer
	dem = astits.NewDemuxer(
		context.Background(),
		&playbackReader{r: br, buf: rr.buf},
		astits.DemuxerOptPacketSize(188))

	return &Reader{
		tracks:        tracks,
		dem:           dem,
		onDecodeError: func(error) {},
		onData:        make(map[uint16]func(int64, int64, []byte) error),
	}, nil
}

// Tracks returns detected tracks.
func (r *Reader) Tracks() []*Track {
	return r.tracks
}

// OnDecodeError sets a callback that is called when a non-fatal decode error occurs.
func (r *Reader) OnDecodeError(cb ReaderOnDecodeErrorFunc) {
	r.onDecodeError = cb
}

// OnDataH26x sets a callback that is called when data from an H26x track is received.
func (r *Reader) OnDataH26x(track *Track, cb ReaderOnDataH26xFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		au, err := h264.AnnexBUnmarshal(data)
		if err != nil {
			r.onDecodeError(err)
			return nil
		}

		return cb(pts, dts, au)
	}
}

// OnDataMPEG4Audio sets a callback that is called when data from an MPEG-4 Audio track is received.
func (r *Reader) OnDataMPEG4Audio(track *Track, cb ReaderOnDataMPEG4AudioFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		var pkts mpeg4audio.ADTSPackets
		err := pkts.Unmarshal(data)
		if err != nil {
			r.onDecodeError(err)
			return nil
		}

		aus := make([][]byte, len(pkts))
		for i, pkt := range pkts {
			aus[i] = pkt.AU
		}

		return cb(pts, dts, aus)
	}
}

// OnDataOpus sets a callback that is called when data from an Opus track is received.
func (r *Reader) OnDataOpus(track *Track, cb ReaderOnDataOpusFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		pos := 0
		var packets [][]byte

		for {
			var au opusAccessUnit
			n, err := au.unmarshal(data[pos:])
			if err != nil {
				r.onDecodeError(err)
				return nil
			}
			pos += n

			packets = append(packets, au.Packet)

			if len(data[pos:]) == 0 {
				break
			}
		}

		return cb(pts, dts, packets)
	}
}

// Read reads data.
func (r *Reader) Read() error {
	data, err := r.dem.NextData()
	if err != nil {
		return err
	}

	if data.PES == nil {
		return nil
	}

	if data.PES.Header.OptionalHeader == nil ||
		data.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorNoPTSOrDTS ||
		data.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorIsForbidden {
		return fmt.Errorf("PTS is missing")
	}

	pts := data.PES.Header.OptionalHeader.PTS.Base

	var dts int64
	if data.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorBothPresent {
		dts = data.PES.Header.OptionalHeader.DTS.Base
	} else {
		dts = pts
	}

	onData, ok := r.onData[data.PID]
	if !ok {
		return nil
	}

	return onData(pts, dts, data.PES.Data)
}
