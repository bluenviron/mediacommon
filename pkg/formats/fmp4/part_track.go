package fmp4

import (
	"github.com/abema/go-mp4"

	"github.com/bluenviron/mediacommon/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
)

// PartSample is a sample of a PartTrack.
type PartSample struct {
	Duration        uint32
	PTSOffset       int32
	IsNonSyncSample bool
	Payload         []byte
}

// NewPartSampleAV1 creates a sample with AV1 data.
func NewPartSampleAV1(sequenceHeaderPresent bool, tu [][]byte) (*PartSample, error) {
	bs, err := av1.BitstreamMarshal(tu)
	if err != nil {
		return nil, err
	}

	return &PartSample{
		IsNonSyncSample: !sequenceHeaderPresent,
		Payload:         bs,
	}, nil
}

// NewPartSampleH26x creates a sample with H26x data.
func NewPartSampleH26x(ptsOffset int32, randomAccessPresent bool, au [][]byte) (*PartSample, error) {
	avcc, err := h264.AVCCMarshal(au)
	if err != nil {
		return nil, err
	}

	return &PartSample{
		PTSOffset:       ptsOffset,
		IsNonSyncSample: !randomAccessPresent,
		Payload:         avcc,
	}, nil
}

// GetAV1 gets AV1 data from the sample.
func (ps PartSample) GetAV1() ([][]byte, error) {
	tu, err := av1.BitstreamUnmarshal(ps.Payload, true)
	if err != nil {
		return nil, err
	}

	return tu, nil
}

// GetH26x gets H26x data from the sample.
func (ps PartSample) GetH26x() ([][]byte, error) {
	au, err := h264.AVCCUnmarshal(ps.Payload)
	if err != nil {
		return nil, err
	}

	return au, nil
}

// PartTrack is a track of Part.
type PartTrack struct {
	ID       int
	BaseTime uint64
	Samples  []*PartSample
}

func (pt *PartTrack) marshal(w *mp4Writer) (*mp4.Trun, int, error) {
	/*
		traf
		- tfhd
		- tfdt
		- trun
	*/

	_, err := w.writeBoxStart(&mp4.Traf{}) // <traf>
	if err != nil {
		return nil, 0, err
	}

	flags := 0

	_, err = w.writeBox(&mp4.Tfhd{ // <tfhd/>
		FullBox: mp4.FullBox{
			Flags: [3]byte{2, byte(flags >> 8), byte(flags)},
		},
		TrackID: uint32(pt.ID),
	})
	if err != nil {
		return nil, 0, err
	}

	_, err = w.writeBox(&mp4.Tfdt{ // <tfdt/>
		FullBox: mp4.FullBox{
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

	trun := &mp4.Trun{ // <trun/>
		FullBox: mp4.FullBox{
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

		trun.Entries = append(trun.Entries, mp4.TrunEntry{
			SampleDuration:                sample.Duration,
			SampleSize:                    uint32(len(sample.Payload)),
			SampleFlags:                   flags,
			SampleCompositionTimeOffsetV1: sample.PTSOffset,
		})
	}

	trunOffset, err := w.writeBox(trun)
	if err != nil {
		return nil, 0, err
	}

	err = w.writeBoxEnd() // </traf>
	if err != nil {
		return nil, 0, err
	}

	return trun, trunOffset, nil
}
