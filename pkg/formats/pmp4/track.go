package pmp4

import (
	amp4 "github.com/abema/go-mp4"

	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

const (
	videoBitrate = 1000000
	audioBitrate = 128825
)

func allSamplesAreSync(samples []*Sample) bool {
	for _, sa := range samples {
		if sa.IsNonSyncSample {
			return false
		}
	}
	return true
}

type headerTrackMarshalResult struct {
	stco                 *amp4.Stco
	stcoOffset           int
	presentationDuration uint32
}

// Track is a track of a Presentation.
type Track struct {
	ID         int
	TimeScale  uint32
	TimeOffset int32
	Codec      mp4.Codec
	Samples    []*Sample
}

func (t *Track) marshal(w *imp4.Writer) (*headerTrackMarshalResult, error) {
	/*
		|trak|
		|    |tkhd|
		|    |edts|
		|    |    |elst|
		|    |mdia|
		|    |    |mdhd|
		|    |    |hdlr|
		|    |    |minf|
		|    |    |    |vmhd| (video)
		|    |    |    |smhd| (audio)
		|    |    |    |dinf|
		|    |    |    |    |dref|
		|    |    |    |    |    |url|
		|    |    |    |stbl|
		|    |    |    |    |stsd|
		|    |    |    |    |    |XXXX|
		|    |    |    |    |    |    |YYYY|
		|    |    |    |    |stts|
		|    |    |    |    |stss|
		|    |    |    |    |ctts|
		|    |    |    |    |stsc|
		|    |    |    |    |stsz|
		|    |    |    |    |stco|
	*/

	_, err := w.WriteBoxStart(&amp4.Trak{}) // <trak>
	if err != nil {
		return nil, err
	}

	info, err := imp4.ExtractCodecInfo(t.Codec)
	if err != nil {
		return nil, err
	}

	sampleDuration := uint32(0)
	for _, sa := range t.Samples {
		sampleDuration += sa.Duration
	}

	presentationDuration := uint32(((int64(sampleDuration) + int64(t.TimeOffset)) * globalTimescale) / int64(t.TimeScale))

	if t.Codec.IsVideo() {
		_, err = w.WriteBox(&amp4.Tkhd{ // <tkhd/>
			FullBox: amp4.FullBox{
				Flags: [3]byte{0, 0, 3},
			},
			TrackID:    uint32(t.ID),
			DurationV0: presentationDuration,
			Width:      uint32(info.Width * 65536),
			Height:     uint32(info.Height * 65536),
			Matrix:     [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		})
		if err != nil {
			return nil, err
		}
	} else {
		_, err = w.WriteBox(&amp4.Tkhd{ // <tkhd/>
			FullBox: amp4.FullBox{
				Flags: [3]byte{0, 0, 3},
			},
			TrackID:        uint32(t.ID),
			DurationV0:     presentationDuration,
			AlternateGroup: 1,
			Volume:         256,
			Matrix:         [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		})
		if err != nil {
			return nil, err
		}
	}

	_, err = w.WriteBoxStart(&amp4.Edts{}) // <edts>
	if err != nil {
		return nil, err
	}

	err = t.marshalELST(w, sampleDuration) // <elst/>
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </edts>
	if err != nil {
		return nil, err
	}

	_, err = w.WriteBoxStart(&amp4.Mdia{}) // <mdia>
	if err != nil {
		return nil, err
	}

	_, err = w.WriteBox(&amp4.Mdhd{ // <mdhd/>
		Timescale:  t.TimeScale,
		DurationV0: uint32(int64(sampleDuration) + int64(t.TimeOffset)),
		Language:   [3]byte{'u', 'n', 'd'},
	})
	if err != nil {
		return nil, err
	}

	if t.Codec.IsVideo() {
		_, err = w.WriteBox(&amp4.Hdlr{ // <hdlr/>
			HandlerType: [4]byte{'v', 'i', 'd', 'e'},
			Name:        "VideoHandler",
		})
		if err != nil {
			return nil, err
		}
	} else {
		_, err = w.WriteBox(&amp4.Hdlr{ // <hdlr/>
			HandlerType: [4]byte{'s', 'o', 'u', 'n'},
			Name:        "SoundHandler",
		})
		if err != nil {
			return nil, err
		}
	}

	_, err = w.WriteBoxStart(&amp4.Minf{}) // <minf>
	if err != nil {
		return nil, err
	}

	if t.Codec.IsVideo() {
		_, err = w.WriteBox(&amp4.Vmhd{ // <vmhd/>
			FullBox: amp4.FullBox{
				Flags: [3]byte{0, 0, 1},
			},
		})
		if err != nil {
			return nil, err
		}
	} else {
		_, err = w.WriteBox(&amp4.Smhd{}) // <smhd/>
		if err != nil {
			return nil, err
		}
	}

	_, err = w.WriteBoxStart(&amp4.Dinf{}) // <dinf>
	if err != nil {
		return nil, err
	}

	_, err = w.WriteBoxStart(&amp4.Dref{ // <dref>
		EntryCount: 1,
	})
	if err != nil {
		return nil, err
	}

	_, err = w.WriteBox(&amp4.Url{ // <url/>
		FullBox: amp4.FullBox{
			Flags: [3]byte{0, 0, 1},
		},
	})
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </dref>
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </dinf>
	if err != nil {
		return nil, err
	}

	_, err = w.WriteBoxStart(&amp4.Stbl{}) // <stbl>
	if err != nil {
		return nil, err
	}

	_, err = w.WriteBoxStart(&amp4.Stsd{ // <stsd>
		EntryCount: 1,
	})
	if err != nil {
		return nil, err
	}

	var avgBitrate int
	if t.Codec.IsVideo() {
		avgBitrate = videoBitrate
	} else {
		avgBitrate = audioBitrate
	}

	var maxBitrate int
	if t.Codec.IsVideo() {
		maxBitrate = videoBitrate
	} else {
		maxBitrate = audioBitrate
	}

	err = imp4.WriteCodecBoxes(w, t.Codec, t.ID, info, uint32(avgBitrate), uint32(maxBitrate))
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </*>
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </stsd>
	if err != nil {
		return nil, err
	}

	err = t.marshalSTTS(w) // <stts/>
	if err != nil {
		return nil, err
	}

	err = t.marshalSTSS(w) // <stss/>
	if err != nil {
		return nil, err
	}

	err = t.marshalCTTS(w) // <ctts/>
	if err != nil {
		return nil, err
	}

	err = t.marshalSTSC(w) // <stsc/>
	if err != nil {
		return nil, err
	}

	err = t.marshalSTSZ(w) // <stsz/>
	if err != nil {
		return nil, err
	}

	stco, stcoOffset, err := t.marshalSTCO(w) // <stco/>
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </stbl>
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </minf>
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </mdia>
	if err != nil {
		return nil, err
	}

	err = w.WriteBoxEnd() // </trak>
	if err != nil {
		return nil, err
	}

	return &headerTrackMarshalResult{
		stco:                 stco,
		stcoOffset:           stcoOffset,
		presentationDuration: presentationDuration,
	}, nil
}

func (t *Track) marshalELST(w *imp4.Writer, sampleDuration uint32) error {
	if t.TimeOffset > 0 {
		_, err := w.WriteBox(&amp4.Elst{
			EntryCount: 2,
			Entries: []amp4.ElstEntry{
				{ // pause
					SegmentDurationV0: uint32((uint64(t.TimeOffset) * globalTimescale) / uint64(t.TimeScale)),
					MediaTimeV0:       -1,
					MediaRateInteger:  1,
					MediaRateFraction: 0,
				},
				{ // presentation
					SegmentDurationV0: uint32((uint64(sampleDuration) * globalTimescale) / uint64(t.TimeScale)),
					MediaTimeV0:       0,
					MediaRateInteger:  1,
					MediaRateFraction: 0,
				},
			},
		})
		return err
	}

	_, err := w.WriteBox(&amp4.Elst{
		EntryCount: 1,
		Entries: []amp4.ElstEntry{{
			SegmentDurationV0: uint32(((uint64(sampleDuration) +
				uint64(-t.TimeOffset)) * globalTimescale) / uint64(t.TimeScale)),
			MediaTimeV0:       -t.TimeOffset,
			MediaRateInteger:  1,
			MediaRateFraction: 0,
		}},
	})
	return err
}

func (t *Track) marshalSTTS(w *imp4.Writer) error {
	entries := []amp4.SttsEntry{{
		SampleCount: 1,
		SampleDelta: t.Samples[0].Duration,
	}}

	for _, sa := range t.Samples[1:] {
		if sa.Duration == entries[len(entries)-1].SampleDelta {
			entries[len(entries)-1].SampleCount++
		} else {
			entries = append(entries, amp4.SttsEntry{
				SampleCount: 1,
				SampleDelta: sa.Duration,
			})
		}
	}

	_, err := w.WriteBox(&amp4.Stts{
		EntryCount: uint32(len(entries)),
		Entries:    entries,
	})
	return err
}

func (t *Track) marshalSTSS(w *imp4.Writer) error {
	// ISO 14496-12 2015:
	// "If the sync sample box is not present, every sample is a sync sample."
	if allSamplesAreSync(t.Samples) {
		return nil
	}

	var sampleNumbers []uint32

	for i, sa := range t.Samples {
		if !sa.IsNonSyncSample {
			sampleNumbers = append(sampleNumbers, uint32(i+1))
		}
	}

	_, err := w.WriteBox(&amp4.Stss{
		EntryCount:   uint32(len(sampleNumbers)),
		SampleNumber: sampleNumbers,
	})
	return err
}

func (t *Track) marshalCTTS(w *imp4.Writer) error {
	entries := []amp4.CttsEntry{{
		SampleCount:    1,
		SampleOffsetV1: t.Samples[0].PTSOffset,
	}}

	for _, sa := range t.Samples[1:] {
		if sa.PTSOffset == entries[len(entries)-1].SampleOffsetV1 {
			entries[len(entries)-1].SampleCount++
		} else {
			entries = append(entries, amp4.CttsEntry{
				SampleCount:    1,
				SampleOffsetV1: sa.PTSOffset,
			})
		}
	}

	_, err := w.WriteBox(&amp4.Ctts{
		FullBox: amp4.FullBox{
			Version: 1,
		},
		EntryCount: uint32(len(entries)),
		Entries:    entries,
	})
	return err
}

func (t *Track) marshalSTSC(w *imp4.Writer) error {
	entries := []amp4.StscEntry{{
		FirstChunk:             1,
		SamplesPerChunk:        1,
		SampleDescriptionIndex: 1,
	}}

	firstSample := t.Samples[0]
	off := firstSample.offset + firstSample.PayloadSize

	for _, sa := range t.Samples[1:] {
		if sa.offset == off {
			entries[len(entries)-1].SamplesPerChunk++
		} else {
			entries = append(entries, amp4.StscEntry{
				FirstChunk:             uint32(len(entries) + 1),
				SamplesPerChunk:        1,
				SampleDescriptionIndex: 1,
			})
		}

		off = sa.offset + sa.PayloadSize
	}

	// further compression
	for i := len(entries) - 1; i >= 1; i-- {
		if entries[i].SamplesPerChunk == entries[i-1].SamplesPerChunk {
			for j := i; j < len(entries)-1; j++ {
				entries[j] = entries[j+1]
			}
			entries = entries[:len(entries)-1]
		}
	}

	_, err := w.WriteBox(&amp4.Stsc{
		EntryCount: uint32(len(entries)),
		Entries:    entries,
	})
	return err
}

func (t *Track) marshalSTSZ(w *imp4.Writer) error {
	sampleSizes := make([]uint32, len(t.Samples))

	for i, sa := range t.Samples {
		sampleSizes[i] = sa.PayloadSize
	}

	_, err := w.WriteBox(&amp4.Stsz{
		SampleSize:  0,
		SampleCount: uint32(len(sampleSizes)),
		EntrySize:   sampleSizes,
	})
	return err
}

func (t *Track) marshalSTCO(w *imp4.Writer) (*amp4.Stco, int, error) {
	firstSample := t.Samples[0]
	off := firstSample.offset + firstSample.PayloadSize

	entries := []uint32{firstSample.offset}

	for _, sa := range t.Samples[1:] {
		if sa.offset != off {
			entries = append(entries, sa.offset)
		}
		off = sa.offset + sa.PayloadSize
	}

	stco := &amp4.Stco{
		EntryCount:  uint32(len(entries)),
		ChunkOffset: entries,
	}

	offset, err := w.WriteBox(stco)
	if err != nil {
		return nil, 0, err
	}

	return stco, offset, err
}
