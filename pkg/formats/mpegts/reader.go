package mpegts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/asticode/go-astits"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/ac3"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg1audio"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
)

// ReaderOnDecodeErrorFunc is the prototype of the callback passed to OnDecodeError.
type ReaderOnDecodeErrorFunc func(err error)

// ReaderOnDataH264Func is the prototype of the callback passed to OnDataH264.
type ReaderOnDataH264Func func(pts int64, dts int64, au [][]byte) error

// ReaderOnDataH265Func is the prototype of the callback passed to OnDataH265.
type ReaderOnDataH265Func func(pts int64, dts int64, au [][]byte) error

// ReaderOnDataMPEGxVideoFunc is the prototype of the callback passed to OnDataMPEGxVideo.
type ReaderOnDataMPEGxVideoFunc func(pts int64, frame []byte) error

// ReaderOnDataOpusFunc is the prototype of the callback passed to OnDataOpus.
type ReaderOnDataOpusFunc func(pts int64, packets [][]byte) error

// ReaderOnDataMPEG4AudioFunc is the prototype of the callback passed to OnDataMPEG4Audio.
type ReaderOnDataMPEG4AudioFunc func(pts int64, aus [][]byte) error

// ReaderOnDataMPEG1AudioFunc is the prototype of the callback passed to OnDataMPEG1Audio.
type ReaderOnDataMPEG1AudioFunc func(pts int64, frames [][]byte) error

// ReaderOnDataAC3Func is the prototype of the callback passed to OnDataAC3.
type ReaderOnDataAC3Func func(pts int64, frame []byte) error

// ReaderOnDataKLVFunc is the prototype of the callback passed to OnDataKLV.
type ReaderOnDataKLVFunc func(pts int64, packets []byte) error

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
	R io.Reader

	tracks        []*Track
	dem           *astits.Demuxer
	onDecodeError ReaderOnDecodeErrorFunc
	onData        map[uint16]func(int64, int64, []byte) error
	lastValidPTS  int64 // stores the last valid PTS for KLV packets without PTS
}

// Initialize initializes a Reader.
func (r *Reader) Initialize() error {
	rr := &recordReader{r: r.R}

	dem := astits.NewDemuxer(
		context.Background(),
		rr,
		astits.DemuxerOptPacketSize(188))

	pmt, err := findPMT(dem)
	if err != nil {
		return err
	}

	tracks := make([]*Track, len(pmt.ElementaryStreams))

	for i, es := range pmt.ElementaryStreams {
		var track Track
		err := track.unmarshal(dem, es)
		if err != nil {
			return err
		}

		tracks[i] = &track
	}

	// rewind demuxer
	dem = astits.NewDemuxer(
		context.Background(),
		&playbackReader{r: r.R, buf: rr.buf},
		astits.DemuxerOptPacketSize(188))

	r.tracks = tracks
	r.dem = dem
	r.onDecodeError = func(error) {}
	r.onData = make(map[uint16]func(int64, int64, []byte) error)

	return nil
}

// NewReader allocates a Reader.
//
// Deprecated: replaced by Reader.Initialize.
func NewReader(br io.Reader) (*Reader, error) {
	r := &Reader{
		R: br,
	}
	err := r.Initialize()
	return r, err
}

// Tracks returns detected tracks.
func (r *Reader) Tracks() []*Track {
	return r.tracks
}

// OnDecodeError sets a callback that is called when a non-fatal decode error occurs.
func (r *Reader) OnDecodeError(cb ReaderOnDecodeErrorFunc) {
	r.onDecodeError = cb
}

// OnDataH265 sets a callback that is called when data from an H265 track is received.
func (r *Reader) OnDataH265(track *Track, cb ReaderOnDataH265Func) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		var au h264.AnnexB
		err := au.Unmarshal(data)
		if err != nil {
			r.onDecodeError(err)
			return nil
		}

		if au[0][0] == byte(h265.NALUType_AUD_NUT<<1) {
			au = au[1:]
		}

		return cb(pts, dts, au)
	}
}

// OnDataH264 sets a callback that is called when data from an H264 track is received.
func (r *Reader) OnDataH264(track *Track, cb ReaderOnDataH264Func) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		var au h264.AnnexB
		err := au.Unmarshal(data)
		if err != nil {
			r.onDecodeError(err)
			return nil
		}

		if au[0][0] == byte(h264.NALUTypeAccessUnitDelimiter) {
			au = au[1:]
		}

		return cb(pts, dts, au)
	}
}

// OnDataMPEGxVideo sets a callback that is called when data from an MPEG-1/2/4 Video track is received.
func (r *Reader) OnDataMPEGxVideo(track *Track, cb ReaderOnDataMPEGxVideoFunc) {
	r.onData[track.PID] = func(pts int64, _ int64, data []byte) error {
		return cb(pts, data)
	}
}

// OnDataOpus sets a callback that is called when data from an Opus track is received.
func (r *Reader) OnDataOpus(track *Track, cb ReaderOnDataOpusFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		if pts != dts {
			r.onDecodeError(fmt.Errorf("PTS is not equal to DTS"))
			return nil
		}

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

		return cb(pts, packets)
	}
}

// OnDataMPEG4Audio sets a callback that is called when data from an MPEG-4 Audio track is received.
func (r *Reader) OnDataMPEG4Audio(track *Track, cb ReaderOnDataMPEG4AudioFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		if pts != dts {
			r.onDecodeError(fmt.Errorf("PTS is not equal to DTS"))
			return nil
		}

		var pkts mpeg4audio.ADTSPackets
		err := pkts.Unmarshal(data)
		if err != nil {
			r.onDecodeError(fmt.Errorf("invalid ADTS: %w", err))
			return nil
		}

		aus := make([][]byte, len(pkts))
		for i, pkt := range pkts {
			aus[i] = pkt.AU
		}

		return cb(pts, aus)
	}
}

// OnDataMPEG1Audio sets a callback that is called when data from an MPEG-1 Audio track is received.
func (r *Reader) OnDataMPEG1Audio(track *Track, cb ReaderOnDataMPEG1AudioFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		if pts != dts {
			r.onDecodeError(fmt.Errorf("PTS is not equal to DTS"))
			return nil
		}

		var frames [][]byte

		for len(data) > 0 {
			var h mpeg1audio.FrameHeader
			err := h.Unmarshal(data)
			if err != nil {
				r.onDecodeError(err)
				return nil
			}

			fl := h.FrameLen()
			if len(data) < fl {
				r.onDecodeError(fmt.Errorf("buffer is too short"))
				return nil
			}

			var frame []byte
			frame, data = data[:fl], data[fl:]

			frames = append(frames, frame)
		}

		return cb(pts, frames)
	}
}

// OnDataAC3 sets a callback that is called when data from an AC-3 track is received.
func (r *Reader) OnDataAC3(track *Track, cb ReaderOnDataAC3Func) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		if pts != dts {
			r.onDecodeError(fmt.Errorf("PTS is not equal to DTS"))
			return nil
		}

		var syncInfo ac3.SyncInfo
		err := syncInfo.Unmarshal(data)
		if err != nil {
			r.onDecodeError(err)
			return nil
		}
		size := syncInfo.FrameSize()

		if size != len(data) {
			r.onDecodeError(fmt.Errorf("unexpected frame size: got %d, expected %d", len(data), size))
			return nil
		}

		return cb(pts, data)
	}
}

// OnDataKLV sets a callback that is called when data from an KLV track is received.
func (r *Reader) OnDataKLV(track *Track, cb ReaderOnDataKLVFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		if pts != dts {
			r.onDecodeError(fmt.Errorf("PTS is not equal to DTS"))
			return nil
		}

		// Check if this is a metadata stream type that needs access unit decoding
		klvCodec, ok := track.Codec.(*CodecKLV)
		if !ok {
			r.onDecodeError(fmt.Errorf("track is not a KLV track"))
			return nil
		}

		if klvCodec.StreamType == astits.StreamTypeMetadata {
			// For metadata stream type, decode the access unit to extract KLV packets
			var au klvaAccessUnit
			_, err := au.unmarshal(data)
			if err != nil {
				r.onDecodeError(fmt.Errorf("unable to decode KLV access unit: %w", err))
				return nil
			}
			return cb(pts, au.Packet)
		}
		// For private data stream type, pass data directly
		return cb(pts, data)
	}
}

// Read reads data.
func (r *Reader) Read() error {
	for {
		data, err := r.dem.NextData()
		if err != nil {
			// https://github.com/asticode/go-astits/blob/b0b19247aa31633650c32638fb55f597fa6e2468/packet_buffer.go#L133C1-L133C5
			if errors.Is(err, astits.ErrNoMorePackets) || strings.Contains(err.Error(), "astits: reading ") {
				return err
			}
			r.onDecodeError(err)
			continue
		}

		if data.PES == nil {
			return nil
		}

		if data.PES.Header.OptionalHeader == nil ||
			data.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorIsForbidden {
			r.onDecodeError(fmt.Errorf("PTS is missing"))
			return nil
		}

		// Handle PTS for different packet types
		var pts int64
		if data.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorNoPTSOrDTS {
			// For Asynchronous KLV, found in StreamType Private Data (0x06), there is no PTS.
			// Use the last valid PTS from any stream to enable proper RTP muxing.
			isKLVTrack := false
			for _, track := range r.tracks {
				if data.PID == track.PID {
					// Check if this is a KLV track
					switch track.Codec.(type) {
					case *CodecKLV:
						isKLVTrack = true
						pts = r.lastValidPTS // Use last valid PTS for KLV packets
					default:
						// Maintain backward compatibility with existing error message
						r.onDecodeError(fmt.Errorf("PTS is missing"))
						return nil
					}
					break
				}
			}

			if !isKLVTrack {
				r.onDecodeError(fmt.Errorf("PTS is missing"))
				return nil
			}
		} else {
			pts = data.PES.Header.OptionalHeader.PTS.Base
			// Update last valid PTS for future KLV packets
			r.lastValidPTS = pts
		}

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
}
