package mpegts

import (
	"context"
	"io"
	"time"

	"github.com/asticode/go-astits"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

const (
	streamIDVideo = 224
	streamIDAudio = 192
)

func leadingTrack(tracks []*Track) *Track {
	for _, track := range tracks {
		if track.Codec.IsVideo() {
			return track
		}
	}
	return tracks[0]
}

// Writer is a MPEG-TS writer.
type Writer struct {
	tsw *astits.Muxer
}

// NewWriter allocates a Writer.
func NewWriter(
	bw io.Writer,
	tracks []*Track,
) *Writer {
	w := &Writer{}

	w.tsw = astits.NewMuxer(
		context.Background(),
		bw)

	for _, track := range tracks {
		es, _ := track.Marshal()
		w.tsw.AddElementaryStream(*es)
	}

	w.tsw.SetPCRPID(leadingTrack(tracks).PID)

	// WriteTables() is not necessary
	// since it's called automatically when WriteData() is called with
	// * PID == PCRPID
	// * AdaptationField != nil
	// * RandomAccessIndicator = true

	return w
}

// WriteH26x writes a H26x access unit.
func (w *Writer) WriteH26x(
	track *Track,
	dts time.Duration,
	pts time.Duration,
	idrPresent bool,
	au [][]byte,
) error {
	enc, err := h264.AnnexBMarshal(au)
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
		PID:             track.PID,
		AdaptationField: af,
		PES: &astits.PESData{
			Header: &astits.PESHeader{
				OptionalHeader: oh,
				StreamID:       streamIDVideo,
			},
			Data: enc,
		},
	})
	return err
}

// WriteAAC writes an AAC access unit.
func (w *Writer) WriteAAC(
	track *Track,
	pts time.Duration,
	au []byte,
) error {
	aacCodec := track.Codec.(*CodecMPEG4Audio)

	pkts := mpeg4audio.ADTSPackets{
		{
			Type:         aacCodec.Config.Type,
			SampleRate:   aacCodec.SampleRate,
			ChannelCount: aacCodec.Config.ChannelCount,
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
		PID:             track.PID,
		AdaptationField: af,
		PES: &astits.PESData{
			Header: &astits.PESHeader{
				OptionalHeader: &astits.PESOptionalHeader{
					MarkerBits:      2,
					PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
					PTS:             &astits.ClockReference{Base: int64(pts.Seconds() * 90000)},
				},
				PacketLength: uint16(len(enc) + 8),
				StreamID:     streamIDAudio,
			},
			Data: enc,
		},
	})
	return err
}
