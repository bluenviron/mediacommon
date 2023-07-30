package mpegts

import (
	"bytes"
	"context"
	"testing"

	"github.com/asticode/go-astits"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

func TestReader(t *testing.T) {
	t.Run("h264 + aac", func(t *testing.T) {
		var buf bytes.Buffer
		mux := astits.NewMuxer(context.Background(), &buf)

		// PMT
		_, err := mux.WritePacket(&astits.Packet{
			Header: astits.PacketHeader{
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       0,
			},
			Payload: append([]byte{
				0x00, 0x00, 0xb0, 0x0d, 0x00, 0x00, 0xc1, 0x00,
				0x00, 0x00, 0x01, 0xf0, 0x00, 0x71, 0x10, 0xd8,
				0x78,
			}, bytes.Repeat([]byte{0xff}, 167)...),
		})
		require.NoError(t, err)

		// PAT
		_, err = mux.WritePacket(&astits.Packet{
			Header: astits.PacketHeader{
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       4096,
			},
			Payload: append([]byte{
				0x00, 0x02, 0xb0, 0x17, 0x00, 0x01, 0xc1, 0x00,
				0x00, 0xe1, 0x00, 0xf0, 0x00, 0x1b, 0xe1, 0x00,
				0xf0, 0x00, 0x0f, 0xe1, 0x01, 0xf0, 0x00, 0x2f,
				0x44, 0xb9, 0x9b,
			}, bytes.Repeat([]byte{0xff}, 157)...),
		})
		require.NoError(t, err)

		// PES (H264)
		_, err = mux.WritePacket(&astits.Packet{
			AdaptationField: &astits.PacketAdaptationField{
				Length:                130,
				StuffingLength:        129,
				RandomAccessIndicator: true,
			},
			Header: astits.PacketHeader{
				HasAdaptationField:        true,
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       256,
			},
			Payload: []byte{
				0x00, 0x00, 0x01, 0xe0, 0x00, 0x00, 0x80, 0x80,
				0x05, 0x21, 0x00, 0x0b, 0x7e, 0x41, 0x00, 0x00,
				0x00, 0x01,
				0x67, 0x42, 0xc0, 0x28, 0xd9, 0x00, 0x78, 0x02,
				0x27, 0xe5, 0x84, 0x00, 0x00, 0x03, 0x00, 0x04,
				0x00, 0x00, 0x03, 0x00, 0xf0, 0x3c, 0x60, 0xc9,
				0x20,
				0x00, 0x00, 0x00, 0x01, 0x08,
				0x00, 0x00, 0x00, 0x01, 0x05,
			},
		})
		require.NoError(t, err)

		// PES (AAC)
		_, err = mux.WritePacket(&astits.Packet{
			AdaptationField: &astits.PacketAdaptationField{
				Length:                158,
				StuffingLength:        157,
				RandomAccessIndicator: true,
			},
			Header: astits.PacketHeader{
				HasAdaptationField:        true,
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       257,
			},
			Payload: []byte{
				0x00, 0x00, 0x01, 0xc0, 0x00, 0x13, 0x80, 0x80,
				0x05, 0x21, 0x00, 0x11, 0x3d, 0x61, 0xff, 0xf1,
				0x50, 0x80, 0x01, 0x7f, 0xfc, 0x01, 0x02, 0x03,
				0x04,
			},
		})
		require.NoError(t, err)

		r, err := NewReader(&buf)
		require.NoError(t, err)

		require.Equal(t, []*Track{
			{
				PID:   256,
				Codec: &CodecH264{},
			},
			{
				PID: 257,
				Codec: &CodecMPEG4Audio{
					Config: mpeg4audio.AudioSpecificConfig{
						Type:         2,
						SampleRate:   44100,
						ChannelCount: 2,
					},
				},
			},
		}, r.Tracks())

		done := false

		r.OnDataH26x(r.Tracks()[0], func(pts int64, dts int64, au [][]byte) error {
			require.Equal(t, int64(180000), pts)
			require.Equal(t, [][]byte{
				{
					0x67, 0x42, 0xc0, 0x28, 0xd9, 0x00, 0x78, 0x02,
					0x27, 0xe5, 0x84, 0x00, 0x00, 0x03, 0x00, 0x04,
					0x00, 0x00, 0x03, 0x00, 0xf0, 0x3c, 0x60, 0xc9,
					0x20,
				},
				{8},
				{5},
			}, au)
			done = true
			return nil
		})

		for {
			err = r.Read()
			require.NoError(t, err)
			if done {
				break
			}
		}
	})
}

func FuzzReader(f *testing.F) {
	f.Add(true, []byte{
		0x00, 0x00, 0x01, 0xe0, 0x00, 0x00, 0x80, 0x80,
		0x05, 0x21, 0x00, 0x0b, 0x7e, 0x41, 0x00, 0x00,
		0x00, 0x01,
		0x67, 0x42, 0xc0, 0x28, 0xd9, 0x00, 0x78, 0x02,
		0x27, 0xe5, 0x84, 0x00, 0x00, 0x03, 0x00, 0x04,
		0x00, 0x00, 0x03, 0x00, 0xf0, 0x3c, 0x60, 0xc9,
		0x20,
		0x00, 0x00, 0x00, 0x01, 0x08,
		0x00, 0x00, 0x00, 0x01, 0x05,
	})

	f.Fuzz(func(t *testing.T, pid bool, b []byte) {
		var buf bytes.Buffer
		mux := astits.NewMuxer(context.Background(), &buf)

		// PMT
		mux.WritePacket(&astits.Packet{
			Header: astits.PacketHeader{
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       0,
			},
			Payload: append([]byte{
				0x00, 0x00, 0xb0, 0x0d, 0x00, 0x00, 0xc1, 0x00,
				0x00, 0x00, 0x01, 0xf0, 0x00, 0x71, 0x10, 0xd8,
				0x78,
			}, bytes.Repeat([]byte{0xff}, 167)...),
		})

		// PAT
		mux.WritePacket(&astits.Packet{
			Header: astits.PacketHeader{
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       4096,
			},
			Payload: append([]byte{
				0x00, 0x02, 0xb0, 0x17, 0x00, 0x01, 0xc1, 0x00,
				0x00, 0xe1, 0x00, 0xf0, 0x00, 0x1b, 0xe1, 0x00,
				0xf0, 0x00, 0x0f, 0xe1, 0x01, 0xf0, 0x00, 0x2f,
				0x44, 0xb9, 0x9b,
			}, bytes.Repeat([]byte{0xff}, 157)...),
		})

		// AAC config
		mux.WritePacket(&astits.Packet{
			AdaptationField: &astits.PacketAdaptationField{
				Length:                158,
				StuffingLength:        157,
				RandomAccessIndicator: true,
			},
			Header: astits.PacketHeader{
				HasAdaptationField:        true,
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID:                       257,
			},
			Payload: []byte{
				0x00, 0x00, 0x01, 0xc0, 0x00, 0x13, 0x80, 0x80,
				0x05, 0x21, 0x00, 0x11, 0x3d, 0x61, 0xff, 0xf1,
				0x50, 0x80, 0x01, 0x7f, 0xfc, 0x01, 0x02, 0x03,
				0x04,
			},
		})

		r, err := NewReader(&buf)
		if err != nil {
			panic(err)
		}

		// PES
		mux.WritePacket(&astits.Packet{
			AdaptationField: &astits.PacketAdaptationField{
				Length:                130,
				StuffingLength:        129,
				RandomAccessIndicator: true,
			},
			Header: astits.PacketHeader{
				HasAdaptationField:        true,
				HasPayload:                true,
				PayloadUnitStartIndicator: true,
				PID: func() uint16 {
					if pid {
						return 256
					}
					return 257
				}(),
			},
			Payload: b,
		})

		r.Read()
		r.Read()
		r.Read()
	})
}
