package fmp4

import (
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
)

// PartSample is a sample of a PartTrack.
type PartSample struct {
	Duration        uint32
	PTSOffset       int32
	IsNonSyncSample bool
	Payload         []byte
}

// NewPartSampleAV12 creates a sample with AV1 data.
func NewPartSampleAV12(tu [][]byte) (*PartSample, error) {
	randomAccess, err := av1.IsRandomAccess(tu)
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

// NewPartSampleH264 creates a sample with H264 data.
func NewPartSampleH264(ptsOffset int32, au [][]byte) (*PartSample, error) {
	avcc, err := h264.AVCC(au).Marshal()
	if err != nil {
		return nil, err
	}

	return &PartSample{
		PTSOffset:       ptsOffset,
		IsNonSyncSample: !h264.IsRandomAccess(au),
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
