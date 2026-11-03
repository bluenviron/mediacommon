package mpegts_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/asticode/go-astits"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/scte35"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
)

const testSectionPID = 257

// testSection returns a splice_info_section with the given size and a valid CRC_32.
func testSection(size int, fill byte) []byte {
	sec := bytes.Repeat([]byte{fill}, size)
	sec[0] = scte35.TableID
	sec[1] = 0x30 | byte((size-3)>>8)
	sec[2] = byte(size - 3)
	return withPTSAdjustment(sec, 0)
}

func withPTSAdjustment(sec []byte, v int64) []byte {
	out, err := scte35.WithPTSAdjustment(sec, v)
	if err != nil {
		panic(err)
	}
	return out
}

// testSectionPacket returns a raw packet of the SCTE-35 track.
func testSectionPacket(pusi bool, cc uint8, payload []byte) []byte {
	pkt := bytes.Repeat([]byte{0xFF}, 188)
	pkt[0] = 0x47
	pkt[1] = byte(testSectionPID >> 8)
	if pusi {
		pkt[1] |= 0x40
	}
	pkt[2] = byte(testSectionPID & 0xFF)
	pkt[3] = 0x10 | cc
	copy(pkt[4:], payload)
	return pkt
}

func readAll(t *testing.T, r *mpegts.Reader) {
	t.Helper()

	for {
		err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
	}
}

func TestReaderSCTE35Packets(t *testing.T) {
	sec20 := testSection(20, 0x01)
	sec30 := testSection(30, 0x02)
	sec181 := testSection(181, 0x03)
	sec200 := testSection(200, 0x04)
	sec400 := testSection(400, 0x05)

	invalidCRC := append([]byte(nil), sec20...)
	invalidCRC[19] ^= 0xFF

	tooBig := bytes.Repeat([]byte{0x06}, 183)
	tooBig[0] = scte35.TableID
	tooBig[1] = 0x3F
	tooBig[2] = 0xFF

	for _, ca := range []struct {
		name     string
		packets  func() [][]byte
		sections [][]byte
		errors   []string
	}{
		{
			"single packet",
			func() [][]byte {
				return [][]byte{testSectionPacket(true, 0, append([]byte{0}, sec20...))}
			},
			[][]byte{sec20},
			nil,
		},
		{
			"multiple packets",
			func() [][]byte {
				payload := append([]byte{0}, sec400...)
				return [][]byte{
					testSectionPacket(true, 0, payload[:184]),
					testSectionPacket(false, 1, payload[184:368]),
					testSectionPacket(false, 2, payload[368:]),
				}
			},
			[][]byte{sec400},
			nil,
		},
		{
			"multiple sections in packet",
			func() [][]byte {
				payload := append([]byte{0}, sec20...)
				payload = append(payload, sec30...)
				return [][]byte{testSectionPacket(true, 0, payload)}
			},
			[][]byte{sec20, sec30},
			nil,
		},
		{
			"pointer field",
			func() [][]byte {
				// the tail of the first section precedes the start of the second one
				payload := make([]byte, 0, 1+len(sec200)-183+len(sec20))
				payload = append(payload, byte(len(sec200)-183))
				payload = append(payload, sec200[183:]...)
				payload = append(payload, sec20...)
				return [][]byte{
					testSectionPacket(true, 0, append([]byte{0}, sec200[:183]...)),
					testSectionPacket(true, 1, payload),
				}
			},
			[][]byte{sec200, sec20},
			nil,
		},
		{
			"adaptation field",
			func() [][]byte {
				// adaptation field without payload
				pkt1 := testSectionPacket(true, 0, nil)
				pkt1[3] = 0x20
				pkt1[4] = 183
				pkt1[5] = 0x00

				// adaptation field that fills the packet
				pkt2 := testSectionPacket(true, 0, nil)
				pkt2[3] = 0x30
				pkt2[4] = 183
				pkt2[5] = 0x00

				// adaptation field followed by payload
				pkt3 := testSectionPacket(true, 1, nil)
				pkt3[3] |= 0x20
				pkt3[4] = 1
				pkt3[5] = 0x00
				copy(pkt3[6:], append([]byte{0}, sec20...))

				return [][]byte{pkt1, pkt2, pkt3}
			},
			[][]byte{sec20},
			nil,
		},
		{
			"duplicate packet",
			func() [][]byte {
				pkt := testSectionPacket(true, 0, append([]byte{0}, sec20...))
				return [][]byte{pkt, pkt}
			},
			[][]byte{sec20},
			nil,
		},
		{
			"packet loss",
			func() [][]byte {
				payload := append([]byte{0}, sec400...)
				return [][]byte{
					testSectionPacket(true, 0, payload[:184]),
					testSectionPacket(false, 2, payload[368:]),
					testSectionPacket(true, 3, append([]byte{0}, sec20...)),
				}
			},
			[][]byte{sec20},
			[]string{"section discarded because of packet loss"},
		},
		{
			"transport error",
			func() [][]byte {
				payload := append([]byte{0}, sec400...)
				pkt := testSectionPacket(false, 1, payload[184:368])
				pkt[1] |= 0x80
				return [][]byte{
					testSectionPacket(true, 0, payload[:184]),
					pkt,
				}
			},
			nil,
			[]string{"transport error indicator is set"},
		},
		{
			"invalid pointer field",
			func() [][]byte {
				return [][]byte{testSectionPacket(true, 0, []byte{184})}
			},
			nil,
			[]string{"invalid pointer_field"},
		},
		{
			"truncated",
			func() [][]byte {
				// the header of the second section is split between packets,
				// then a new section begins.
				payload := append([]byte{0}, sec181...)
				payload = append(payload, sec20[:2]...)
				return [][]byte{
					testSectionPacket(true, 0, payload),
					testSectionPacket(true, 1, append([]byte{0}, sec30...)),
				}
			},
			[][]byte{sec181, sec30},
			[]string{"section discarded because it is truncated"},
		},
		{
			"too big",
			func() [][]byte {
				return [][]byte{
					testSectionPacket(true, 0, append([]byte{0}, tooBig...)),
					testSectionPacket(false, 1, bytes.Repeat([]byte{0x06}, 184)),
				}
			},
			nil,
			[]string{"section is too big"},
		},
		{
			"invalid crc",
			func() [][]byte {
				// sections are not validated
				return [][]byte{testSectionPacket(true, 0, append([]byte{0}, invalidCRC...))}
			},
			[][]byte{invalidCRC},
			nil,
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			var buf bytes.Buffer
			mux := astits.NewMuxer(context.Background(), &buf)

			err := mux.AddElementaryStream(astits.PMTElementaryStream{
				ElementaryPID: 256,
				StreamType:    astits.StreamTypeH264Video,
			})
			require.NoError(t, err)

			err = mux.AddElementaryStream(astits.PMTElementaryStream{
				ElementaryPID: testSectionPID,
				StreamType:    astits.StreamTypeSCTE35,
			})
			require.NoError(t, err)

			mux.SetPCRPID(256)

			_, err = mux.WriteTables()
			require.NoError(t, err)

			writePES := func(pts int64) {
				_, err2 := mux.WriteData(&astits.MuxerData{
					PID: 256,
					PES: &astits.PESData{
						Header: &astits.PESHeader{
							OptionalHeader: &astits.PESOptionalHeader{
								MarkerBits:      2,
								PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
								PTS:             &astits.ClockReference{Base: pts},
							},
							StreamID: 224,
						},
						Data: []byte{0, 0, 0, 1, byte(h264.NALUTypeIDR)},
					},
				})
				require.NoError(t, err2)
			}

			writePES(90000)

			for _, pkt := range ca.packets() {
				buf.Write(pkt)
			}

			writePES(93000)

			r := &mpegts.Reader{R: &buf}
			err = r.Initialize()
			require.NoError(t, err)

			require.Equal(t, &codecs.SCTE35{}, r.Tracks()[1].Codec)

			var decodeErrors []string
			r.OnDecodeError(func(err error) {
				decodeErrors = append(decodeErrors, err.Error())
			})

			r.OnDataH264(r.Tracks()[0], func(_ int64, _ int64, _ [][]byte) error {
				return nil
			})

			var sections [][]byte
			r.OnDataSCTE35(r.Tracks()[1], func(section []byte) error {
				sections = append(sections, section)
				return nil
			})

			readAll(t, r)

			require.Equal(t, ca.sections, sections)
			require.Equal(t, ca.errors, decodeErrors)
		})
	}
}

func TestReaderSCTE35Order(t *testing.T) {
	// a section that spans over multiple packets
	sec1 := withPTSAdjustment(testSection(400, 0x01), 123456)
	sec2 := withPTSAdjustment(testSection(20, 0x00), 0x1FFFFFFFF)

	var buf bytes.Buffer

	videoTrack := &mpegts.Track{Codec: &codecs.H264{}}
	scte35Track := &mpegts.Track{Codec: &codecs.SCTE35{}}

	w := &mpegts.Writer{W: &buf, Tracks: []*mpegts.Track{videoTrack, scte35Track}}
	err := w.Initialize()
	require.NoError(t, err)

	err = w.WriteH264(videoTrack, 90000, 90000, [][]byte{
		testH264SPS,
		{8}, // PPS
		{byte(h264.NALUTypeIDR)},
	})
	require.NoError(t, err)

	err = w.WriteSCTE35(scte35Track, sec1)
	require.NoError(t, err)

	err = w.WriteH264(videoTrack, 93000, 93000, [][]byte{{byte(h264.NALUTypeNonIDR)}})
	require.NoError(t, err)

	err = w.WriteSCTE35(scte35Track, sec2)
	require.NoError(t, err)

	err = w.WriteH264(videoTrack, 96000, 96000, [][]byte{{byte(h264.NALUTypeNonIDR)}})
	require.NoError(t, err)

	r := &mpegts.Reader{R: &buf}
	err = r.Initialize()
	require.NoError(t, err)

	require.Equal(t, []*mpegts.Track{
		{PID: 256, Codec: &codecs.H264{}},
		{PID: 257, Codec: &codecs.SCTE35{}},
	}, r.Tracks())

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

	r.OnDataSCTE35(r.Tracks()[1], func(section []byte) error {
		received = append(received, event{0, section})
		return nil
	})

	readAll(t, r)

	// sections are returned in stream order
	require.Equal(t, []event{
		{90000, nil},
		{0, sec1},
		{93000, nil},
		{0, sec2},
		{96000, nil},
	}, received)
}

func TestReaderSCTE35Only(t *testing.T) {
	var buf bytes.Buffer

	// a stream with a SCTE-35 track only
	track := &mpegts.Track{Codec: &codecs.SCTE35{}}

	w := &mpegts.Writer{W: &buf, Tracks: []*mpegts.Track{track}}
	err := w.Initialize()
	require.NoError(t, err)

	err = w.WriteSCTE35(track, testSection(20, 0x00))
	require.NoError(t, err)

	r := &mpegts.Reader{R: &buf}
	err = r.Initialize()
	require.NoError(t, err)

	require.Equal(t, []*mpegts.Track{{PID: 256, Codec: &codecs.SCTE35{}}}, r.Tracks())

	r.OnDecodeError(func(err error) {
		t.Errorf("unexpected decode error: %v", err)
	})

	// sections do not depend on timestamps of other tracks
	var sections [][]byte
	r.OnDataSCTE35(r.Tracks()[0], func(section []byte) error {
		sections = append(sections, section)
		return nil
	})

	readAll(t, r)
	require.Equal(t, [][]byte{testSection(20, 0x00)}, sections)
}

func TestReaderSCTE35WithoutCallback(t *testing.T) {
	var buf bytes.Buffer

	videoTrack := &mpegts.Track{Codec: &codecs.H264{}}
	scte35Track := &mpegts.Track{Codec: &codecs.SCTE35{}}

	w := &mpegts.Writer{W: &buf, Tracks: []*mpegts.Track{videoTrack, scte35Track}}
	err := w.Initialize()
	require.NoError(t, err)

	err = w.WriteH264(videoTrack, 90000, 90000, [][]byte{testH264SPS, {8}, {byte(h264.NALUTypeIDR)}})
	require.NoError(t, err)

	err = w.WriteSCTE35(scte35Track, testSection(20, 0x00))
	require.NoError(t, err)

	r := &mpegts.Reader{R: &buf}
	err = r.Initialize()
	require.NoError(t, err)

	r.OnDecodeError(func(err error) {
		t.Errorf("unexpected decode error: %v", err)
	})

	n := 0
	r.OnDataH264(r.Tracks()[0], func(_ int64, _ int64, _ [][]byte) error {
		n++
		return nil
	})

	readAll(t, r)
	require.Equal(t, 1, n)
}

func TestReaderSCTE35UndeclaredTrack(t *testing.T) {
	var buf bytes.Buffer
	mux := astits.NewMuxer(context.Background(), &buf)

	err := mux.AddElementaryStream(astits.PMTElementaryStream{
		ElementaryPID: 256,
		StreamType:    astits.StreamTypeH264Video,
	})
	require.NoError(t, err)
	mux.SetPCRPID(256)

	writePES := func(pts int64) {
		_, err2 := mux.WriteData(&astits.MuxerData{
			PID:             256,
			AdaptationField: &astits.PacketAdaptationField{RandomAccessIndicator: true},
			PES: &astits.PESData{
				Header: &astits.PESHeader{
					OptionalHeader: &astits.PESOptionalHeader{
						MarkerBits:      2,
						PTSDTSIndicator: astits.PTSDTSIndicatorOnlyPTS,
						PTS:             &astits.ClockReference{Base: pts},
					},
					StreamID: 224,
				},
				Data: []byte{0, 0, 0, 1, byte(h264.NALUTypeIDR)},
			},
		})
		require.NoError(t, err2)
	}

	writePES(90000)

	// the PMT is updated with a SCTE-35 track
	err = mux.AddElementaryStream(astits.PMTElementaryStream{
		ElementaryPID: testSectionPID,
		StreamType:    astits.StreamTypeSCTE35,
	})
	require.NoError(t, err)

	writePES(93000)

	buf.Write(testSectionPacket(true, 0, append([]byte{0}, testSection(20, 0x00)...)))

	writePES(96000)

	r := &mpegts.Reader{R: &buf}
	err = r.Initialize()
	require.NoError(t, err)

	require.Len(t, r.Tracks(), 1)

	var decodeErrors []string
	r.OnDecodeError(func(err error) {
		decodeErrors = append(decodeErrors, err.Error())
	})

	r.OnDataH264(r.Tracks()[0], func(_ int64, _ int64, _ [][]byte) error {
		return nil
	})

	readAll(t, r)
	require.Equal(t, []string{"received data from undeclared track with PID 257"}, decodeErrors)
}

type limitedWriter struct {
	n int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > w.n {
		return 0, errors.New("write limit reached")
	}
	w.n -= len(p)
	return len(p), nil
}

func TestWriterSCTE35Errors(t *testing.T) {
	for _, ca := range []struct {
		name    string
		limit   int
		section []byte
		err     string
	}{
		{
			"empty section",
			1000,
			nil,
			"empty section",
		},
		{
			"tables",
			0,
			testSection(20, 0x00),
			"write limit reached",
		},
		{
			"section",
			2 * 188,
			testSection(20, 0x00),
			"write limit reached",
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			track := &mpegts.Track{Codec: &codecs.SCTE35{}}

			w := &mpegts.Writer{W: &limitedWriter{n: ca.limit}, Tracks: []*mpegts.Track{track}}
			err := w.Initialize()
			require.NoError(t, err)

			err = w.WriteSCTE35(track, ca.section)
			require.ErrorContains(t, err, ca.err)
		})
	}
}
