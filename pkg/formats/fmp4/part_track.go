package fmp4

import (
	amp4 "github.com/abema/go-mp4"
	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
)

// PartTrack is a track of Part.
type PartTrack struct {
	ID       int
	BaseTime uint64
	Samples  []*Sample
}

func (pt *PartTrack) marshal(w *imp4.Writer) (*amp4.Trun, int, error) {
	/*
		|traf|
		|    |tfhd|
		|    |tfdt|
		|    |trun|
	*/

	_, err := w.WriteBoxStart(&amp4.Traf{}) // <traf>
	if err != nil {
		return nil, 0, err
	}

	flags := 0

	_, err = w.WriteBox(&amp4.Tfhd{ // <tfhd/>
		FullBox: amp4.FullBox{
			Flags: [3]byte{2, byte(flags >> 8), byte(flags)},
		},
		TrackID: uint32(pt.ID),
	})
	if err != nil {
		return nil, 0, err
	}

	_, err = w.WriteBox(&amp4.Tfdt{ // <tfdt/>
		FullBox: amp4.FullBox{
			Version: 1,
		},
		// sum of decode durations of all earlier samples
		BaseMediaDecodeTimeV1: pt.BaseTime,
	})
	if err != nil {
		return nil, 0, err
	}

	flags = trunFlagDataOffsetPreset |
		trunFlagSampleDurationPresent |
		trunFlagSampleSizePresent

	for _, sample := range pt.Samples {
		if sample.IsNonSyncSample {
			flags |= trunFlagSampleFlagsPresent
		}
		if sample.PTSOffset != 0 {
			flags |= trunFlagSampleCompositionTimeOffsetPresentOrV1
		}
	}

	trun := &amp4.Trun{ // <trun/>
		FullBox: amp4.FullBox{
			Version: 1,
			Flags:   [3]byte{0, byte(flags >> 8), byte(flags)},
		},
		SampleCount: uint32(len(pt.Samples)),
	}

	for _, sample := range pt.Samples {
		var flags uint32
		if sample.IsNonSyncSample {
			flags |= sampleFlagIsNonSyncSample
		}

		trun.Entries = append(trun.Entries, amp4.TrunEntry{
			SampleDuration:                sample.Duration,
			SampleSize:                    uint32(len(sample.Payload)),
			SampleFlags:                   flags,
			SampleCompositionTimeOffsetV1: sample.PTSOffset,
		})
	}

	trunOffset, err := w.WriteBox(trun)
	if err != nil {
		return nil, 0, err
	}

	err = w.WriteBoxEnd() // </traf>
	if err != nil {
		return nil, 0, err
	}

	return trun, trunOffset, nil
}
