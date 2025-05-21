package fmp4

import (
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
)

// PartSample is a fMP4 sample.
//
// Deprecated: replaced by Sample.
type PartSample = Sample

// Sample is a fMP4 sample.
type Sample struct {
	Duration        uint32
	PTSOffset       int32
	IsNonSyncSample bool
	Payload         []byte
}

// NewSampleAV12 creates a sample with AV1 data.
//
// Deprecated: replaced by FillAV1.
func NewSampleAV12(tu [][]byte) (*Sample, error) {
	ps := &Sample{}
	err := ps.FillAV1(tu)
	return ps, err
}

// FillAV1 fills a Sample with AV1 data.
func (ps *Sample) FillAV1(tu [][]byte) error {
	bs, err := av1.Bitstream(tu).Marshal()
	if err != nil {
		return err
	}

	ps.IsNonSyncSample = !av1.IsRandomAccess2(tu)
	ps.Payload = bs

	return nil
}

// NewSampleH265 creates a sample with H265 data.
//
// Deprecated: replaced by FillH265.
func NewSampleH265(ptsOffset int32, au [][]byte) (*Sample, error) {
	ps := &Sample{}
	err := ps.FillH265(ptsOffset, au)
	return ps, err
}

// FillH265 fills a Sample with H265 data.
func (ps *Sample) FillH265(ptsOffset int32, au [][]byte) error {
	avcc, err := h264.AVCC(au).Marshal()
	if err != nil {
		return err
	}

	ps.PTSOffset = ptsOffset
	ps.IsNonSyncSample = !h265.IsRandomAccess(au)
	ps.Payload = avcc

	return nil
}

// NewSampleH264 creates a sample with H264 data.
//
// Deprecated: replaced by FillH264.
func NewSampleH264(ptsOffset int32, au [][]byte) (*Sample, error) {
	ps := &Sample{}
	err := ps.FillH264(ptsOffset, au)
	return ps, err
}

// FillH264 fills a Sample with H264 data.
func (ps *Sample) FillH264(ptsOffset int32, au [][]byte) error {
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
func (ps Sample) GetAV1() ([][]byte, error) {
	var tu av1.Bitstream
	err := tu.Unmarshal(ps.Payload)
	if err != nil {
		return nil, err
	}

	return tu, nil
}

// GetH265 gets H265 data from the sample.
func (ps Sample) GetH265() ([][]byte, error) {
	return ps.GetH264()
}

// GetH264 gets H264 data from the sample.
func (ps Sample) GetH264() ([][]byte, error) {
	var au h264.AVCC
	err := au.Unmarshal(ps.Payload)
	if err != nil {
		return nil, err
	}

	return au, nil
}
