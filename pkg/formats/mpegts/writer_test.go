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
					err := w.WriteH2652(ca.track, sample.pts, sample.dts, sample.data)
					require.NoError(t, err)

				case *CodecH264:
					err := w.WriteH2642(ca.track, sample.pts, sample.dts, sample.data)
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
