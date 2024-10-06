package h265

import (
	"fmt"
	"time"
)

// DTSExtractor computes DTS from PTS.
//
// Deprecated: replaced by DTSExtractor2.
type DTSExtractor struct {
	spsp          *SPS
	ppsp          *PPS
	prevDTSFilled bool
	prevDTS       time.Duration
}

// NewDTSExtractor allocates a DTSExtractor.
//
// Deprecated: replaced by NewDTSExtractor2.
func NewDTSExtractor() *DTSExtractor {
	return &DTSExtractor{}
}

func (d *DTSExtractor) extractInner(au [][]byte, pts time.Duration) (time.Duration, error) {
	var idr []byte
	var nonIDR []byte

	for _, nalu := range au {
		typ := NALUType((nalu[0] >> 1) & 0b111111)
		switch typ {
		case NALUType_SPS_NUT:
			var spsp SPS
			err := spsp.Unmarshal(nalu)
			if err != nil {
				return 0, fmt.Errorf("invalid SPS: %w", err)
			}
			d.spsp = &spsp

		case NALUType_PPS_NUT:
			var ppsp PPS
			err := ppsp.Unmarshal(nalu)
			if err != nil {
				return 0, fmt.Errorf("invalid PPS: %w", err)
			}
			d.ppsp = &ppsp

		case NALUType_IDR_W_RADL, NALUType_IDR_N_LP:
			idr = nalu

		case NALUType_TRAIL_N, NALUType_TRAIL_R, NALUType_CRA_NUT, NALUType_RASL_N, NALUType_RASL_R:
			nonIDR = nalu
		}
	}

	if d.spsp == nil {
		return 0, fmt.Errorf("SPS not received yet")
	}

	if d.ppsp == nil {
		return 0, fmt.Errorf("PPS not received yet")
	}

	if len(d.spsp.MaxNumReorderPics) != 1 || d.spsp.MaxNumReorderPics[0] == 0 {
		return pts, nil
	}

	if d.spsp.VUI == nil || d.spsp.VUI.TimingInfo == nil {
		return pts, nil
	}

	var samplesDiff uint32

	switch {
	case idr != nil:
		samplesDiff = d.spsp.MaxNumReorderPics[0]

	case nonIDR != nil:
		var err error
		samplesDiff, err = getPTSDTSDiff(nonIDR, d.spsp, d.ppsp)
		if err != nil {
			return 0, err
		}

	default:
		return 0, fmt.Errorf("access unit doesn't contain an IDR or non-IDR NALU")
	}

	timeDiff := time.Duration(samplesDiff) * time.Second *
		time.Duration(d.spsp.VUI.TimingInfo.NumUnitsInTick) / time.Duration(d.spsp.VUI.TimingInfo.TimeScale)
	dts := pts - timeDiff

	return dts, nil
}

// Extract extracts the DTS of a access unit.
func (d *DTSExtractor) Extract(au [][]byte, pts time.Duration) (time.Duration, error) {
	dts, err := d.extractInner(au, pts)
	if err != nil {
		return 0, err
	}

	if dts > pts {
		return 0, fmt.Errorf("DTS is greater than PTS")
	}

	if d.prevDTSFilled && dts < d.prevDTS {
		return 0, fmt.Errorf("DTS is not monotonically increasing, was %v, now is %v",
			d.prevDTS, dts)
	}

	d.prevDTSFilled = true
	d.prevDTS = dts

	return dts, err
}
