package fmp4

import (
	"github.com/bluenviron/mediacommon/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/pkg/codecs/h265"
)

// PartSample is a sample of a PartTrack.
type PartSample struct {
	Duration        uint32
	PTSOffset       int32
	IsNonSyncSample bool
	Payload         []byte
}

// NewPartSampleAV1 creates a sample with AV1 data.
//
// Deprecated: replaced by NewPartSampleAV12
func NewPartSampleAV1(_ bool, tu [][]byte) (*PartSample, error) {
	return NewPartSampleAV12(tu)
}

// NewPartSampleAV12 creates a sample with AV1 data.
func NewPartSampleAV12(tu [][]byte) (*PartSample, error) {
	randomAccess, err := av1.ContainsKeyFrame(tu)
	if err != nil {
		return nil, err
	}

	bs, err := av1.Bitstream(tu).Marshal()
	if err != nil {
		return nil, err
	}

	return &PartSample{
		IsNonSyncSample: !randomAccess,
		Payload:         bs,
	}, nil
}

// NewPartSampleH26x creates a sample with H26x data.
//
// Deprecated: replaced by NewPartSampleH264 and NewPartSampleH265
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

// NewPartSampleH264 creates a sample with H264 data.
func NewPartSampleH264(ptsOffset int32, au [][]byte) (*PartSample, error) {
	avcc, err := h264.AVCC(au).Marshal()
	if err != nil {
		return nil, err
	}

	return &PartSample{
		PTSOffset:       ptsOffset,
		IsNonSyncSample: !h264.IDRPresent(au),
		Payload:         avcc,
	}, nil
}

// NewPartSampleH265 creates a sample with H265 data.
func NewPartSampleH265(ptsOffset int32, au [][]byte) (*PartSample, error) {
	avcc, err := h264.AVCC(au).Marshal()
	if err != nil {
		return nil, err
	}

	return &PartSample{
		PTSOffset:       ptsOffset,
		IsNonSyncSample: !h265.IsRandomAccess(au),
		Payload:         avcc,
	}, nil
}

// GetAV1 gets AV1 data from the sample.
func (ps PartSample) GetAV1() ([][]byte, error) {
	var tu av1.Bitstream
	err := tu.Unmarshal(ps.Payload)
	if err != nil {
		return nil, err
	}

	return tu, nil
}

// GetH26x gets H26x data from the sample.
//
// Deprecated: replaced by GetH264 and GetH265
func (ps PartSample) GetH26x() ([][]byte, error) {
	return ps.GetH264()
}

// GetH264 gets H264 data from the sample.
func (ps PartSample) GetH264() ([][]byte, error) {
	var au h264.AVCC
	err := au.Unmarshal(ps.Payload)
	if err != nil {
		return nil, err
	}

	return au, nil
}

// GetH265 gets H265 data from the sample.
func (ps PartSample) GetH265() ([][]byte, error) {
	return ps.GetH264()
}
