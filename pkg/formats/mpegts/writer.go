package mpegts

import (
	"context"
	"io"
	"time"

	"github.com/asticode/go-astits"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

func trackToElementaryStream(t *Track) *astits.PMTElementaryStream {
	switch t.Codec.(type) {
	case *CodecH264:
		return &astits.PMTElementaryStream{
			ElementaryPID: 256,
			StreamType:    astits.StreamTypeH264Video,
		}

	case *CodecMPEG4Audio:
		return &astits.PMTElementaryStream{
			ElementaryPID: 257,
			StreamType:    astits.StreamTypeAACAudio,
		}
	}

	return nil
}

// Writer is a MPEG-TS writer.
type Writer struct {
	videoTrack *Track
	audioTrack *Track

	tsw *astits.Muxer
}

// NewWriter allocates a Writer.
func NewWriter(
	bw io.Writer,
	videoTrack *Track,
	audioTrack *Track,
) *Writer {
	w := &Writer{
		videoTrack: videoTrack,
		audioTrack: audioTrack,
	}

	w.tsw = astits.NewMuxer(
		context.Background(),
		bw)

	if videoTrack != nil {
		w.tsw.AddElementaryStream(*trackToElementaryStream(videoTrack))
	}

	if audioTrack != nil {
		w.tsw.AddElementaryStream(*trackToElementaryStream(audioTrack))
	}

	if videoTrack != nil {
		w.tsw.SetPCRPID(256)
	} else {
		w.tsw.SetPCRPID(257)
	}

	// WriteTable() is not necessary
	// since it's called automatically when WriteData() is called with
	// * PID == PCRPID
	// * AdaptationField != nil
	// * RandomAccessIndicator = true

	return w
}

// WriteH264 writes a H264 access unit.
func (w *Writer) WriteH264(
	dts time.Duration,
	pts time.Duration,
	idrPresent bool,
	nalus [][]byte,
) error {
	// prepend an AUD. This is required by video.js and iOS
	nalus = append([][]byte{{byte(h264.NALUTypeAccessUnitDelimiter), 240}}, nalus...)

	enc, err := h264.AnnexBMarshal(nalus)
	if err != nil {
		return err
	}

	var af *astits.PacketAdaptationField

	if idrPresent {
		af = &astits.PacketAdaptationField{}
		af.RandomAccessIndicator = true
	}

	oh := &astits.PESOptionalHeader{
		MarkerBits: 2,
	}

	if dts == pts {
		oh.PTSDTSIndicator = astits.PTSDTSIndicatorOnlyPTS
		oh.PTS = &astits.ClockReference{Base: int64(pts.Seconds() * 90000)}
	} else {
		oh.PTSDTSIndicator = astits.PTSDTSIndicatorBothPresent
		oh.DTS = &astits.ClockReference{Base: int64(dts.Seconds() * 90000)}
		oh.PTS = &astits.ClockReference{Base: int64(pts.Seconds() * 90000)}
	}

	_, err = w.tsw.WriteData(&astits.MuxerData{
		PID:             256,
		AdaptationField: af,
		PES: &astits.PESData{
			Header: &astits.PESHeader{
				OptionalHeader: oh,
				StreamID:       224, // video
			},
			Data: enc,
		},
	})
	return err
}

// WriteAAC writes an AAC AU.
func (w *Writer) WriteAAC(
	pts time.Duration,
	au []byte,
) error {
	pkts := mpeg4audio.ADTSPackets{
		{
			Type:         w.audioTrack.Codec.(*CodecMPEG4Audio).Config.Type,
			SampleRate:   w.audioTrack.Codec.(*CodecMPEG4Audio).Config.SampleRate,
			ChannelCount: w.audioTrack.Codec.(*CodecMPEG4Audio).Config.ChannelCount,
			AU:           au,
		},
	}

	enc, err := pkts.Marshal()
	if err != nil {
		return err
	}

	af := &astits.PacketAdaptationField{
		RandomAccessIndicator: true,
	}

	_, err = w.tsw.WriteData(&astits.MuxerData{
		PID:             257,
		AdaptationField: af,
		PES: &astits.PESData{
			Header: &astits.PESHeader{
				OptionalHeader: &astits.PESOptionalHeader{
					MarkerBits:      2,
					PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
					PTS:             &astits.ClockReference{Base: int64(pts.Seconds() * 90000)},
				},
				PacketLength: uint16(len(enc) + 8),
				StreamID:     192, // audio
			},
			Data: enc,
		},
	})
	return err
}
