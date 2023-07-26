package mpegts

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/asticode/go-astits"

	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

type onDataH26xFunc func(pts int64, dts int64, au [][]byte) error

type onDataMPEG4AudioFunc func(pts int64, dts int64, aus [][]byte) error

type onDataOpusFunc func(pts int64, dts int64, packets [][]byte) error

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
	tracks []*Track
	dem    *astits.Demuxer
	onData map[uint16]func(int64, int64, []byte) error
}

// NewReader allocates a Reader.
func NewReader(br io.Reader) (*Reader, error) {
	rr := &recordedReader{r: br}

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
		io.MultiReader(bytes.NewReader(rr.buf), br),
		astits.DemuxerOptPacketSize(188))

	return &Reader{
		tracks: tracks,
		dem:    dem,
		onData: make(map[uint16]func(int64, int64, []byte) error),
	}, nil
}

// Tracks returns detected tracks.
func (r *Reader) Tracks() []*Track {
	return r.tracks
}

// OnDataH26x sets a callback that is called when data from an H26x track is received.
func (r *Reader) OnDataH26x(track *Track, onData onDataH26xFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		au, err := h264.AnnexBUnmarshal(data)
		if err != nil {
			return err
		}

		return onData(pts, dts, au)
	}
}

// OnDataMPEG4Audio sets a callback that is called when data from an AAC track is received.
func (r *Reader) OnDataMPEG4Audio(track *Track, onData onDataMPEG4AudioFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		var pkts mpeg4audio.ADTSPackets
		err := pkts.Unmarshal(data)
		if err != nil {
			return err
		}

		aus := make([][]byte, len(pkts))
		for i, pkt := range pkts {
			aus[i] = pkt.AU
		}

		return onData(pts, dts, aus)
	}
}

// OnDataOpus sets a callback that is called when data from an Opus track is received.
func (r *Reader) OnDataOpus(track *Track, onData onDataOpusFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		pos := 0
		var packets [][]byte

		for {
			var au opusAccessUnit
			n, err := au.unmarshal(data[pos:])
			if err != nil {
				return err
			}
			pos += n

			packets = append(packets, au.Packet)

			if len(data[pos:]) == 0 {
				break
			}
		}

		return onData(pts, dts, packets)
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
