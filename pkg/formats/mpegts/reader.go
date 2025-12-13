package mpegts

import (
	"fmt"
	"io"

	"github.com/asticode/go-astits"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/ac3"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/eac3"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg1audio"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
	"github.com/bluenviron/mediacommon/v2/pkg/rewindablereader"
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

// ReaderOnDataMPEG4AudioLATMFunc is the prototype of the callback passed to OnDataMPEG4AudioLATM.
type ReaderOnDataMPEG4AudioLATMFunc func(pts int64, els [][]byte) error

// ReaderOnDataMPEG1AudioFunc is the prototype of the callback passed to OnDataMPEG1Audio.
type ReaderOnDataMPEG1AudioFunc func(pts int64, frames [][]byte) error

// ReaderOnDataAC3Func is the prototype of the callback passed to OnDataAC3.
type ReaderOnDataAC3Func func(pts int64, frame []byte) error

// ReaderOnDataEAC3Func is the prototype of the callback passed to OnDataEAC3.
type ReaderOnDataEAC3Func func(pts int64, frame []byte) error

// ReaderOnDataKLVFunc is the prototype of the callback passed to OnDataKLV.
type ReaderOnDataKLVFunc func(pts int64, data []byte) error

// ReaderOnDataDVBSubtitleFunc is the prototype of the callback passed to OnDataDVBSubtitle.
type ReaderOnDataDVBSubtitleFunc func(pts int64, data []byte) error

func findPMT(dem *robustDemuxer) (*astits.PMTData, error) {
	for {
		data, err := dem.nextData()
		if err != nil {
			return nil, err
		}

		if data.PMT != nil {
			return data.PMT, nil
		}
	}
}

func readMetadataAUWrapper(in []byte) ([]byte, error) {
	expectedSeqNum := 0

	var au metadataAUCell
	n, err := au.unmarshal(in)
	if err != nil {
		return nil, err
	}

	if int(au.SequenceNumber) != expectedSeqNum {
		return nil, fmt.Errorf("unexpected sequence number: %d, expected %d", au.SequenceNumber, expectedSeqNum)
	}
	expectedSeqNum++

	switch au.CellFragmentIndication {
	case 0b11:
		if n != len(in) {
			return nil, fmt.Errorf("unread bytes detected")
		}
		return au.AUCellData, nil

	case 0b10:

	default:
		return nil, fmt.Errorf("unexpected cell_fragment_indication %v", au.CellFragmentIndication)
	}

	out := au.AUCellData

	for {
		var n2 int
		n2, err = au.unmarshal(in[n:])
		if err != nil {
			return nil, err
		}
		n += n2

		if int(au.SequenceNumber) != expectedSeqNum {
			return nil, fmt.Errorf("unexpected sequence number: %d, expected %d", au.SequenceNumber, expectedSeqNum)
		}
		expectedSeqNum++

		out = append(out, au.AUCellData...)

		switch au.CellFragmentIndication {
		case 0b01:
			if n != len(in) {
				return nil, fmt.Errorf("unread bytes detected")
			}
			return out, nil

		case 0b00:

		default:
			return nil, fmt.Errorf("unexpected cell_fragment_indication %v", au.CellFragmentIndication)
		}
	}
}

func writeMetadataAUWrapper(in []byte) ([]byte, error) {
	const maxDataPerCell = 65535
	dataLen := len(in)
	cellCount := dataLen / maxDataPerCell
	if (dataLen % maxDataPerCell) != 0 {
		cellCount++
	}

	bufLen := 5*cellCount + dataLen
	out := make([]byte, bufLen)
	n := 0

	for i := range cellCount {
		cellDataLen := min(maxDataPerCell, len(in))
		cellData := in[:cellDataLen]
		in = in[cellDataLen:]

		var fragmentIndication uint8
		switch {
		case cellCount == 1:
			fragmentIndication = 0b11

		case i == 0:
			fragmentIndication = 0b10

		case i == cellCount-1:
			fragmentIndication = 0b01

		default:
			fragmentIndication = 0b00
		}

		n2, err := metadataAUCell{
			MetadataServiceID:      0,
			SequenceNumber:         uint8(i),
			CellFragmentIndication: fragmentIndication,
			DecoderConfigFlag:      false,
			RandomAccessIndicator:  true,
			AUCellData:             cellData,
		}.marshalTo(out[n:])
		if err != nil {
			return nil, err
		}
		n += n2
	}

	return out, nil
}

// Reader is a MPEG-TS reader.
type Reader struct {
	R io.Reader

	tracks          []*Track
	tracksByPID     map[uint16]*Track
	preDem          *preDemuxer
	dem             *robustDemuxer
	onDecodeError   ReaderOnDecodeErrorFunc
	onData          map[uint16]func(int64, int64, []byte) error
	lastPTSReceived bool
	lastPTS         int64
}

// Initialize initializes a Reader.
func (r *Reader) Initialize() error {
	rr := &rewindablereader.Reader{R: r.R}

	preDem := &preDemuxer{R: rr}
	preDem.initialize()
	dem := &robustDemuxer{R: preDem}
	dem.initialize()

	pmt, err := findPMT(dem)
	if err != nil {
		return err
	}

	tracks := make([]*Track, len(pmt.ElementaryStreams))

	for i, es := range pmt.ElementaryStreams {
		var track Track
		err = track.unmarshal(dem, es)
		if err != nil {
			return err
		}

		tracks[i] = &track
	}

	r.tracks = tracks

	r.tracksByPID = make(map[uint16]*Track)
	for _, track := range tracks {
		r.tracksByPID[track.PID] = track
	}

	// rewind demuxer
	rr.Rewind()
	r.preDem = &preDemuxer{R: rr}
	r.preDem.initialize()
	r.dem = &robustDemuxer{R: r.preDem}
	r.dem.initialize()

	r.onDecodeError = func(_ error) {}
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
	r.preDem.OnDecodeError = cb
	r.dem.OnDecodeError = cb
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

// OnDataMPEG4AudioLATM sets a callback that is called when data from an MPEG-4 Audio LATM track is received.
func (r *Reader) OnDataMPEG4AudioLATM(track *Track, cb ReaderOnDataMPEG4AudioLATMFunc) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		if pts != dts {
			r.onDecodeError(fmt.Errorf("PTS is not equal to DTS"))
			return nil
		}

		var s mpeg4audio.AudioSyncStream
		err := s.Unmarshal(data)
		if err != nil {
			r.onDecodeError(err)
			return nil
		}

		return cb(pts, s.AudioMuxElements)
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

// OnDataEAC3 sets a callback that is called when data from an E-AC-3 track is received.
func (r *Reader) OnDataEAC3(track *Track, cb ReaderOnDataEAC3Func) {
	r.onData[track.PID] = func(pts int64, dts int64, data []byte) error {
		if pts != dts {
			r.onDecodeError(fmt.Errorf("PTS is not equal to DTS"))
			return nil
		}

		var syncInfo eac3.SyncInfo
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

// OnDataKLV sets a callback that is called when data from a KLV track is received.
func (r *Reader) OnDataKLV(track *Track, cb ReaderOnDataKLVFunc) {
	codec := track.Codec.(*codecs.KLV)

	if codec.Synchronous {
		r.onData[track.PID] = func(pts int64, _ int64, data []byte) error {
			out, err := readMetadataAUWrapper(data)
			if err != nil {
				r.onDecodeError(err)
				return nil
			}

			return cb(pts, out)
		}
	} else {
		r.onData[track.PID] = func(pts int64, _ int64, data []byte) error {
			return cb(pts, data)
		}
	}
}

// OnDataDVBSubtitle sets a callback that is called when data from a DVB subtitle track is received.
func (r *Reader) OnDataDVBSubtitle(track *Track, cb ReaderOnDataDVBSubtitleFunc) {
	r.onData[track.PID] = func(pts int64, _ int64, data []byte) error {
		return cb(pts, data)
	}
}

// Read reads data.
func (r *Reader) Read() error {
	data, err := r.dem.nextData()
	if err != nil {
		return err
	}

	if data.PES == nil {
		return nil
	}

	track, ok := r.tracksByPID[data.PID]
	if !ok {
		r.onDecodeError(fmt.Errorf("received data from undeclared track with PID %d", data.PID))
		return nil
	}

	var pts int64
	var dts int64

	if klvCodec, ok2 := track.Codec.(*codecs.KLV); ok2 && !klvCodec.Synchronous {
		if !r.lastPTSReceived {
			return nil
		}

		pts = r.lastPTS
	} else {
		if data.PES.Header.OptionalHeader == nil ||
			data.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorNoPTSOrDTS ||
			data.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorIsForbidden {
			r.onDecodeError(fmt.Errorf("PTS is missing"))
			return nil
		}

		pts = data.PES.Header.OptionalHeader.PTS.Base

		if data.PES.Header.OptionalHeader.PTSDTSIndicator == astits.PTSDTSIndicatorBothPresent {
			dts = data.PES.Header.OptionalHeader.DTS.Base
		} else {
			dts = pts
		}

		r.lastPTS = pts
		r.lastPTSReceived = true
	}

	onData, ok := r.onData[data.PID]
	if !ok {
		return nil
	}

	return onData(pts, dts, data.PES.Data)
}
