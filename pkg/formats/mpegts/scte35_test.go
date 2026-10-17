package mpegts

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/scte35"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
)

func TestSCTE35ReadWrite(t *testing.T) {
	// a section that spans over multiple packets
	sec1 := []byte(scte35.Section(sectionOfSize(400, 0x01)).WithPTSAdjustment(123456))
	sec2 := []byte(scte35.Section(sectionOfSize(20, 0x00)).WithPTSAdjustment(0x1FFFFFFFF))

	var buf bytes.Buffer

	videoTrack := &Track{Codec: &codecs.H264{}}
	scte35Track := &Track{Codec: &codecs.SCTE35{}}

	w := &Writer{W: &buf, Tracks: []*Track{videoTrack, scte35Track}}
	err := w.Initialize()
	require.NoError(t, err)

	idr := [][]byte{
		testH264SPS,
		{8}, // PPS
		{byte(h264.NALUTypeIDR)},
	}

	err = w.WriteH264(videoTrack, 90000, 90000, idr)
	require.NoError(t, err)

	err = w.WriteSCTE35(scte35Track, sec1)
	require.NoError(t, err)

	err = w.WriteH264(videoTrack, 93000, 93000, [][]byte{{byte(h264.NALUTypeNonIDR)}})
	require.NoError(t, err)

	err = w.WriteSCTE35(scte35Track, sec2)
	require.NoError(t, err)

	err = w.WriteH264(videoTrack, 96000, 96000, [][]byte{{byte(h264.NALUTypeNonIDR)}})
	require.NoError(t, err)

	r := &Reader{R: &buf}
	err = r.Initialize()
	require.NoError(t, err)

	require.Len(t, r.Tracks(), 2)
	require.Equal(t, &codecs.SCTE35{}, r.Tracks()[1].Codec)

	r.OnDecodeError(func(err error) {
		t.Errorf("unexpected decode error: %v", err)
	})

	type event struct {
		pts     int64
		section []byte
	}
	var received []event

	r.OnDataH264(r.Tracks()[0], func(pts int64, _ int64, _ [][]byte) error {
		received = append(received, event{pts, nil})
		return nil
	})

	r.OnDataSCTE35(r.Tracks()[1], func(pts int64, section []byte) error {
		received = append(received, event{pts, section})
		return nil
	})

	for {
		err = r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
	}

	// sections are returned in stream order,
	// even though astits returns each PES when the next one begins.
	require.Equal(t, []event{
		{90000, nil},
		{90000, sec1},
		{93000, nil},
		{93000, sec2},
		{96000, nil},
	}, received)
}

func TestSCTE35ReadInvalidCRC(t *testing.T) {
	var buf bytes.Buffer

	videoTrack := &Track{Codec: &codecs.H264{}}
	scte35Track := &Track{Codec: &codecs.SCTE35{}}

	w := &Writer{W: &buf, Tracks: []*Track{videoTrack, scte35Track}}
	err := w.Initialize()
	require.NoError(t, err)

	err = w.WriteH264(videoTrack, 90000, 90000, [][]byte{
		testH264SPS,
		{8},
		{byte(h264.NALUTypeIDR)},
	})
	require.NoError(t, err)

	// CRC is not set
	err = w.WriteSCTE35(scte35Track, sectionOfSize(20, 0x00))
	require.NoError(t, err)

	r := &Reader{R: &buf}
	err = r.Initialize()
	require.NoError(t, err)

	var decodeErrors []string
	r.OnDecodeError(func(err error) {
		decodeErrors = append(decodeErrors, err.Error())
	})

	r.OnDataSCTE35(r.Tracks()[1], func(_ int64, _ []byte) error {
		t.Errorf("should not happen")
		return nil
	})

	for {
		err = r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
	}

	require.Equal(t, []string{"invalid SCTE-35 section: CRC_32 mismatch"}, decodeErrors)
}
