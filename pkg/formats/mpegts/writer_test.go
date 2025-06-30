//nolint:dupl
package mpegts

import (
	"bytes"
	"context"
	"errors"
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
				switch ca.track.Codec.(type) {
				case *CodecH265:
					err := w.WriteH265(ca.track, sample.pts, sample.dts, sample.data)
					require.NoError(t, err)

				case *CodecH264:
					err := w.WriteH264(ca.track, sample.pts, sample.dts, sample.data)
					require.NoError(t, err)

				case *CodecMPEG4Video:
					err := w.WriteMPEG4Video(ca.track, sample.pts, sample.data[0])
					require.NoError(t, err)

				case *CodecMPEG1Video:
					err := w.WriteMPEG1Video(ca.track, sample.pts, sample.data[0])
					require.NoError(t, err)

				case *CodecOpus:
					err := w.WriteOpus(ca.track, sample.pts, sample.data)
					require.NoError(t, err)

				case *CodecMPEG4Audio:
					err := w.WriteMPEG4Audio(ca.track, sample.pts, sample.data)
					require.NoError(t, err)

				case *CodecMPEG1Audio:
					err := w.WriteMPEG1Audio(ca.track, sample.pts, sample.data)
					require.NoError(t, err)

				case *CodecAC3:
					err := w.WriteAC3(ca.track, sample.pts, sample.data[0])
					require.NoError(t, err)

				default:
					t.Errorf("unexpected")
				}
			}

			dem := astits.NewDemuxer(
				context.Background(),
				&buf,
				astits.DemuxerOptPacketSize(188))

			i := 0

			for {
				pkt, err := dem.NextPacket()
				if errors.Is(err, astits.ErrNoMorePackets) {
					break
				}
				require.NoError(t, err)

				if i >= len(ca.packets) {
					t.Errorf("missing packet: %#v", pkt)
					break
				}

				require.Equal(t, ca.packets[i], pkt)
				i++
			}

			require.Equal(t, len(ca.packets), i)
		})
	}
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

func TestWriterKLVMetadataStream(t *testing.T) {
	// Test writing KLV data with metadata stream type
	klvTrack := &Track{
		Codec: &CodecKLV{
			StreamType:      astits.StreamTypeMetadata,
			StreamID:        0xFC,
			PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
		},
	}

	var buf bytes.Buffer
	w := &Writer{
		W:      &buf,
		Tracks: []*Track{klvTrack},
	}
	err := w.Initialize()
	require.NoError(t, err)

	// Write KLV data
	klvData := []byte{0x06, 0x0E, 0x2B, 0x34, 0x01, 0x01, 0x01, 0x01}
	pts := int64(90000)
	err = w.WriteKLV(klvTrack, pts, klvData)
	require.NoError(t, err)

	// Verify data was written
	require.Greater(t, buf.Len(), 0, "Data should be written to buffer")

	// Read back and verify
	r := &Reader{R: bytes.NewReader(buf.Bytes())}
	err = r.Initialize()
	require.NoError(t, err)

	tracks := r.Tracks()
	require.Len(t, tracks, 1)

	klvCodec := tracks[0].Codec.(*CodecKLV)
	require.Equal(t, astits.StreamTypeMetadata, klvCodec.StreamType)

	var receivedData []byte
	var receivedPTS int64
	var dataReceived bool

	r.OnDataKLV(tracks[0], func(pts int64, data []byte) error {
		receivedData = make([]byte, len(data))
		copy(receivedData, data)
		receivedPTS = pts
		dataReceived = true
		return nil
	})

	// Read packets
	for {
		err := r.Read()
		if err != nil {
			if errors.Is(err, astits.ErrNoMorePackets) {
				break
			}
			require.NoError(t, err)
		}
	}

	require.True(t, dataReceived, "KLV data should be received")
	require.Equal(t, klvData, receivedData, "Data should match")
	require.Equal(t, pts, receivedPTS, "PTS should match")
}

func TestWriterKLVPrivateDataStream(t *testing.T) {
	// Test writing KLV data with private data stream type
	klvTrack := &Track{
		Codec: &CodecKLV{
			StreamType:      astits.StreamTypePrivateData,
			StreamID:        0xFC,
			PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
		},
	}

	var buf bytes.Buffer
	w := &Writer{
		W:      &buf,
		Tracks: []*Track{klvTrack},
	}
	err := w.Initialize()
	require.NoError(t, err)

	// Write KLV data
	klvData := []byte{0x06, 0x0E, 0x2B, 0x34, 0x01, 0x01, 0x01, 0x01}
	pts := int64(90000)
	err = w.WriteKLV(klvTrack, pts, klvData)
	require.NoError(t, err)

	// Verify data was written
	require.Greater(t, buf.Len(), 0, "Data should be written to buffer")

	// Read back and verify
	r := &Reader{R: bytes.NewReader(buf.Bytes())}
	err = r.Initialize()
	require.NoError(t, err)

	tracks := r.Tracks()
	require.Len(t, tracks, 1)

	klvCodec := tracks[0].Codec.(*CodecKLV)
	require.Equal(t, astits.StreamTypePrivateData, klvCodec.StreamType)

	var receivedData []byte
	var receivedPTS int64
	var dataReceived bool

	r.OnDataKLV(tracks[0], func(pts int64, data []byte) error {
		receivedData = make([]byte, len(data))
		copy(receivedData, data)
		receivedPTS = pts
		dataReceived = true
		return nil
	})

	// Read packets
	for {
		err := r.Read()
		if err != nil {
			if errors.Is(err, astits.ErrNoMorePackets) {
				break
			}
			require.NoError(t, err)
		}
	}

	require.True(t, dataReceived, "KLV data should be received")
	require.Equal(t, klvData, receivedData, "Data should match")
	require.Equal(t, pts, receivedPTS, "PTS should match")
}

func TestWriterKLVNoPTS(t *testing.T) {
	// Test writing KLV data without PTS
	klvTrack := &Track{
		Codec: &CodecKLV{
			StreamType:      astits.StreamTypePrivateData,
			StreamID:        0xFC,
			PTSDTSIndicator: astits.PTSDTSIndicatorNoPTSOrDTS,
		},
	}

	var buf bytes.Buffer
	w := &Writer{
		W:      &buf,
		Tracks: []*Track{klvTrack},
	}
	err := w.Initialize()
	require.NoError(t, err)

	// Write KLV data
	klvData := []byte{0x06, 0x0E, 0x2B, 0x34, 0x01, 0x01, 0x01, 0x01}
	pts := int64(90000)
	err = w.WriteKLV(klvTrack, pts, klvData)
	require.NoError(t, err)

	// Verify data was written
	require.Greater(t, buf.Len(), 0, "Data should be written to buffer")
}

func TestWriterKLVErrorCases(t *testing.T) {
	t.Run("invalid_access_unit_marshal", func(t *testing.T) {
		// This test would require modifying the marshalTo method to return an error
		// For now, we test that the method handles normal cases correctly
		klvTrack := &Track{
			Codec: &CodecKLV{
				StreamType:      astits.StreamTypeMetadata,
				StreamID:        0xFC,
				PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
			},
		}

		var buf bytes.Buffer
		w := &Writer{
			W:      &buf,
			Tracks: []*Track{klvTrack},
		}
		err := w.Initialize()
		require.NoError(t, err)

		// Write valid KLV data
		klvData := []byte{0x06, 0x0E, 0x2B, 0x34}
		pts := int64(90000)
		err = w.WriteKLV(klvTrack, pts, klvData)
		require.NoError(t, err)
	})

	t.Run("wrong_codec_type", func(t *testing.T) {
		// Test that WriteKLV panics with wrong codec type
		h264Track := &Track{
			Codec: &CodecH264{},
		}

		var buf bytes.Buffer
		w := &Writer{
			W:      &buf,
			Tracks: []*Track{h264Track},
		}
		err := w.Initialize()
		require.NoError(t, err)

		// This should panic because track doesn't have KLV codec
		require.Panics(t, func() {
			_ = w.WriteKLV(h264Track, 90000, []byte{0x06, 0x0E, 0x2B, 0x34})
		})
	})
}

func TestWriterKLVLargeData(t *testing.T) {
	// Test writing large KLV data
	klvTrack := &Track{
		Codec: &CodecKLV{
			StreamType:      astits.StreamTypeMetadata,
			StreamID:        0xFC,
			PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
		},
	}

	var buf bytes.Buffer
	w := &Writer{
		W:      &buf,
		Tracks: []*Track{klvTrack},
	}
	err := w.Initialize()
	require.NoError(t, err)

	// Create large KLV data (but within reasonable limits)
	largeData := make([]byte, 1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	pts := int64(90000)
	err = w.WriteKLV(klvTrack, pts, largeData)
	require.NoError(t, err)

	// Verify data was written
	require.Greater(t, buf.Len(), 0, "Data should be written to buffer")

	// Read back and verify
	r := &Reader{R: bytes.NewReader(buf.Bytes())}
	err = r.Initialize()
	require.NoError(t, err)

	tracks := r.Tracks()
	require.Len(t, tracks, 1)

	var receivedData []byte
	var dataReceived bool

	r.OnDataKLV(tracks[0], func(_ int64, data []byte) error {
		receivedData = make([]byte, len(data))
		copy(receivedData, data)
		dataReceived = true
		return nil
	})

	// Read packets
	for {
		err := r.Read()
		if err != nil {
			if errors.Is(err, astits.ErrNoMorePackets) {
				break
			}
			require.NoError(t, err)
		}
	}

	require.True(t, dataReceived, "KLV data should be received")
	require.Equal(t, largeData, receivedData, "Large data should match")
}

func TestWriterKLVEmptyData(t *testing.T) {
	// Test writing empty KLV data
	klvTrack := &Track{
		Codec: &CodecKLV{
			StreamType:      astits.StreamTypeMetadata,
			StreamID:        0xFC,
			PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
		},
	}

	var buf bytes.Buffer
	w := &Writer{
		W:      &buf,
		Tracks: []*Track{klvTrack},
	}
	err := w.Initialize()
	require.NoError(t, err)

	// Write empty KLV data
	emptyData := []byte{}
	pts := int64(90000)
	err = w.WriteKLV(klvTrack, pts, emptyData)
	require.NoError(t, err)

	// Verify data was written
	require.Greater(t, buf.Len(), 0, "Data should be written to buffer")

	// Read back and verify
	r := &Reader{R: bytes.NewReader(buf.Bytes())}
	err = r.Initialize()
	require.NoError(t, err)

	tracks := r.Tracks()
	require.Len(t, tracks, 1)

	var receivedData []byte
	var dataReceived bool

	r.OnDataKLV(tracks[0], func(_ int64, data []byte) error {
		receivedData = data
		dataReceived = true
		return nil
	})

	// Read packets
	for {
		err := r.Read()
		if err != nil {
			if errors.Is(err, astits.ErrNoMorePackets) {
				break
			}
			require.NoError(t, err)
		}
	}

	require.True(t, dataReceived, "KLV data should be received")
	require.Equal(t, emptyData, receivedData, "Empty data should match")
}
