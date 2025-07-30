//nolint:dupl
package mpegts

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/asticode/go-astits"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	for _, ca := range casesReadWriter {
		t.Run(ca.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewWriter(&buf, []*Track{ca.track})

			for _, sample := range ca.samples {
				var err error

				switch ca.track.Codec.(type) {
				case *CodecH265:
					err = w.WriteH265(ca.track, sample.pts, sample.dts, sample.data)

				case *CodecH264:
					err = w.WriteH264(ca.track, sample.pts, sample.dts, sample.data)

				case *CodecMPEG4Video:
					err = w.WriteMPEG4Video(ca.track, sample.pts, sample.data[0])

				case *CodecMPEG1Video:
					err = w.WriteMPEG1Video(ca.track, sample.pts, sample.data[0])

				case *CodecOpus:
					err = w.WriteOpus(ca.track, sample.pts, sample.data)

				case *CodecMPEG4Audio:
					err = w.WriteMPEG4Audio(ca.track, sample.pts, sample.data)

				case *CodecMPEG4AudioLATM:
					err = w.WriteMPEG4AudioLATM(ca.track, sample.pts, sample.data)

				case *CodecMPEG1Audio:
					err = w.WriteMPEG1Audio(ca.track, sample.pts, sample.data)

				case *CodecAC3:
					err = w.WriteAC3(ca.track, sample.pts, sample.data[0])

				case *CodecKLV:
					err = w.WriteKLV(ca.track, sample.pts, sample.data[0])

				case *CodecDVBSubtitle:
					err = w.WriteDVBSubtitle(ca.track, sample.pts, sample.data[0])

				default:
					panic("unexpected")
				}

				require.NoError(t, err)
			}

			dem := astits.NewDemuxer(
				context.Background(),
				&buf,
				astits.DemuxerOptPacketSize(188))

			var pkts []*astits.Packet
			for {
				pkt, err := dem.NextPacket()
				if errors.Is(err, astits.ErrNoMorePackets) {
					break
				}
				require.NoError(t, err)
				pkts = append(pkts, pkt)
			}

			require.Equal(t, ca.packets, pkts)
		})
	}
}

func TestWriterKLVAsync(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{
		W: &buf,
		Tracks: []*Track{
			{
				Codec: &CodecH264{},
			},
			{
				Codec: &CodecKLV{
					Synchronous: false,
				},
			},
		},
	}
	err := w.Initialize()
	require.NoError(t, err)

	err = w.WriteH264(w.Tracks[0], 90000, 90000, [][]byte{{1, 2, 3, 4}})
	require.NoError(t, err)

	err = w.WriteKLV(w.Tracks[1], 90000, []byte{5, 6, 7, 8})
	require.NoError(t, err)

	dem := astits.NewDemuxer(
		context.Background(),
		&buf,
		astits.DemuxerOptPacketSize(188))

	var pkts []*astits.Packet
	for {
		var pkt *astits.Packet
		pkt, err = dem.NextPacket()
		if errors.Is(err, astits.ErrNoMorePackets) {
			break
		}
		require.NoError(t, err)
		pkts = append(pkts, pkt)
	}

	expected := []*astits.Packet{
		{
			Header: astits.PacketHeader{
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
			},
			Payload: append(
				[]byte{
					0x00, 0x00, 0xb0, 0x0d, 0x00, 0x00, 0xc1, 0x00,
					0x00, 0x00, 0x01, 0xf0, 0x00, 0x71, 0x10, 0xd8,
					0x78,
				},
				bytes.Repeat([]byte{0xff}, 167)...,
			),
		},
		{
			Header: astits.PacketHeader{
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       4096,
			},
			Payload: append(
				[]byte{
					0x00, 0x02, 0xb0, 0x1d, 0x00, 0x01, 0xc1, 0x00,
					0x00, 0xe1, 0x00, 0xf0, 0x00, 0x1b, 0xe1, 0x00,
					0xf0, 0x00, 0x06, 0xe1, 0x01, 0xf0, 0x06, 0x05,
					0x04, 0x4b, 0x4c, 0x56, 0x41, 0x06, 0x71, 0x49,
					0xd4,
				},
				bytes.Repeat([]byte{0xff}, 151)...,
			),
		},
		{
			Header: astits.PacketHeader{
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       256,
				HasAdaptationField:        true,
			},
			AdaptationField: &astits.PacketAdaptationField{
				PCR: &astits.ClockReference{
					Base: 81000,
				},
				Length:         155,
				StuffingLength: 148,
				HasPCR:         true,
			},
			Payload: []byte{
				0x00, 0x00, 0x01, 0xe0, 0x00, 0x00, 0x80, 0x80,
				0x05, 0x21, 0x00, 0x05, 0xbf, 0x21, 0x00, 0x00,
				0x00, 0x01, 0x09, 0xf0, 0x00, 0x00, 0x00, 0x01,
				0x01, 0x02, 0x03, 0x04,
			},
		},
		{
			Header: astits.PacketHeader{
				HasAdaptationField:        true,
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       257,
			},
			AdaptationField: &astits.PacketAdaptationField{
				Length:                170,
				StuffingLength:        169,
				RandomAccessIndicator: true,
			},
			Payload: []byte{
				0x00, 0x00, 0x01, 0xbd, 0x00, 0x07, 0x80, 0x00,
				0x00, 0x05, 0x06, 0x07, 0x08,
			},
		},
	}

	require.Equal(t, expected, pkts)
}

func TestWriterReaderLongKLVSync(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{
		W: &buf,
		Tracks: []*Track{
			{
				Codec: &CodecKLV{
					Synchronous: true,
				},
			},
		},
	}
	err := w.Initialize()
	require.NoError(t, err)

	err = w.WriteKLV(w.Tracks[0], 90000, bytes.Repeat([]byte{1, 2, 3, 4}, 200000/4))
	require.NoError(t, err)

	r := &Reader{
		R: bytes.NewReader(buf.Bytes()),
	}
	err = r.Initialize()
	require.NoError(t, err)

	require.Equal(t, []*Track{{
		PID: 256,
		Codec: &CodecKLV{
			Synchronous: true,
		},
	}}, r.Tracks())

	ok := false

	r.OnDataKLV(r.Tracks()[0], func(_ int64, data []byte) error {
		require.Equal(t, bytes.Repeat([]byte{1, 2, 3, 4}, 200000/4), data)
		ok = true
		return nil
	})

	for {
		err = r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
	}

	require.True(t, ok)
}

func TestWriterAutomaticPID(t *testing.T) {
	track := &Track{
		Codec: &CodecH265{},
	}

	var buf bytes.Buffer
	NewWriter(&buf, []*Track{track})
	require.NotEqual(t, 0, track.PID)
}

func TestWriterError(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{
		W: &buf,
		Tracks: []*Track{
			{
				PID:   11,
				Codec: &CodecH265{},
			},
			{
				PID:   11,
				Codec: &CodecH265{},
			},
		},
	}
	err := w.Initialize()
	require.Error(t, err)
}
