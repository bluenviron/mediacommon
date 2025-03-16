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
//
// Deprecated: replaced by FillAV1.
func NewPartSampleAV12(tu [][]byte) (*PartSample, error) {
	ps := &PartSample{}
	err := ps.FillAV1(tu)
	return ps, err
}

// FillAV1 fills a PartSample with AV1 data.
func (ps *PartSample) FillAV1(tu [][]byte) error {
	bs, err := av1.Bitstream(tu).Marshal()
	if err != nil {
		return err
	}

	ps.IsNonSyncSample = !av1.IsRandomAccess2(tu)
	ps.Payload = bs

	return nil
}

// NewPartSampleH265 creates a sample with H265 data.
//
// Deprecated: replaced by FillH265.
func NewPartSampleH265(ptsOffset int32, au [][]byte) (*PartSample, error) {
	ps := &PartSample{}
	err := ps.FillH265(ptsOffset, au)
	return ps, err
}

// FillH265 fills a PartSample with H265 data.
func (ps *PartSample) FillH265(ptsOffset int32, au [][]byte) error {
	avcc, err := h264.AVCC(au).Marshal()
	if err != nil {
		return err
	}

	ps.PTSOffset = ptsOffset
	ps.IsNonSyncSample = !h265.IsRandomAccess(au)
	ps.Payload = avcc

	return nil
}

// NewPartSampleH264 creates a sample with H264 data.
//
// Deprecated: replaced by FillH264.
func NewPartSampleH264(ptsOffset int32, au [][]byte) (*PartSample, error) {
	ps := &PartSample{}
	err := ps.FillH264(ptsOffset, au)
	return ps, err
}

// FillH264 fills a PartSample with H264 data.
func (ps *PartSample) FillH264(ptsOffset int32, au [][]byte) error {
	avcc, err := h264.AVCC(au).Marshal()
	if err != nil {
		return err
	}

	ps.PTSOffset = ptsOffset
	ps.IsNonSyncSample = !h264.IsRandomAccess(au)
	ps.Payload = avcc

	return nil
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

// GetH265 gets H265 data from the sample.
func (ps PartSample) GetH265() ([][]byte, error) {
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
