package h265

import (
	"bytes"
	"fmt"
	"math"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
)

const (
	maxBytesToGetPOC = 12
)

func getPTSDTSDiff(buf []byte, sps *SPS, pps *PPS) (int, error) {
	typ := NALUType((buf[0] >> 1) & 0b111111)
	buf = buf[1:]
	lb := min(len(buf), maxBytesToGetPOC)

	buf = h264.EmulationPreventionRemove(buf[:lb])
	pos := 8

	firstSliceSegmentInPicFlag, err := bits.ReadFlag(buf, &pos)
	if err != nil {
		return 0, err
	}

	if !firstSliceSegmentInPicFlag {
		return 0, fmt.Errorf("first_slice_segment_in_pic_flag = 0 is not supported")
	}

	if typ >= NALUType_BLA_W_LP && typ <= NALUType_RSV_IRAP_VCL23 {
		pos++ // no_output_of_prior_pics_flag
	}

	_, err = bits.ReadGolombUnsigned(buf, &pos) // slice_pic_parameter_set_id
	if err != nil {
		return 0, err
	}

	if pps.NumExtraSliceHeaderBits > 0 {
		err = bits.HasSpace(buf, pos, int(pps.NumExtraSliceHeaderBits))
		if err != nil {
			return 0, err
		}
		pos += int(pps.NumExtraSliceHeaderBits)
	}

	sliceType, err := bits.ReadGolombUnsigned(buf, &pos) // slice_type
	if err != nil {
		return 0, err
	}

	if pps.OutputFlagPresentFlag {
		_, err = bits.ReadFlag(buf, &pos) // pic_output_flag
		if err != nil {
			return 0, err
		}
	}

	if sps.SeparateColourPlaneFlag {
		_, err = bits.ReadBits(buf, &pos, 2) // colour_plane_id
		if err != nil {
			return 0, err
		}
	}

	_, err = bits.ReadBits(buf, &pos, int(sps.Log2MaxPicOrderCntLsbMinus4+4)) // pic_order_cnt_lsb
	if err != nil {
		return 0, err
	}

	shortTermRefPicSetSpsFlag, err := bits.ReadFlag(buf, &pos)
	if err != nil {
		return 0, err
	}

	var rps *SPS_ShortTermRefPicSet

	if !shortTermRefPicSetSpsFlag {
		rps = &SPS_ShortTermRefPicSet{}
		err = rps.unmarshal(buf, &pos, uint32(len(sps.ShortTermRefPicSets)),
			uint32(len(sps.ShortTermRefPicSets)), sps.ShortTermRefPicSets)
		if err != nil {
			return 0, err
		}
	} else {
		if len(sps.ShortTermRefPicSets) <= 1 {
			return 0, nil
		}

		b := int(math.Ceil(math.Log2(float64(len(sps.ShortTermRefPicSets)))))
		var tmp uint64
		tmp, err = bits.ReadBits(buf, &pos, b)
		if err != nil {
			return 0, err
		}
		shortTermRefPicSetIdx := int(tmp)

		if len(sps.ShortTermRefPicSets) <= shortTermRefPicSetIdx {
			return 0, fmt.Errorf("invalid short_term_ref_pic_set_idx")
		}

		rps = sps.ShortTermRefPicSets[shortTermRefPicSetIdx]
	}

	if sliceType == 0 { // B-frame
		switch typ {
		case NALUType_TRAIL_N, NALUType_RASL_N:
			return -len(rps.DeltaPocS1), nil

		case NALUType_TRAIL_R, NALUType_RASL_R:
			if len(rps.DeltaPocS0) == 0 {
				return 0, fmt.Errorf("invalid DeltaPocS0")
			}
			return int(-rps.DeltaPocS0[0]-1) - len(rps.DeltaPocS1), nil

		default:
			return 0, nil
		}
	} else { // I or P-frame
		if len(rps.DeltaPocS0) == 0 {
			return 0, fmt.Errorf("invalid DeltaPocS0")
		}
		return int(-rps.DeltaPocS0[0] - 1), nil
	}
}

// DTSExtractor computes DTS from PTS.
type DTSExtractor struct {
	sps             []byte
	spsp            *SPS
	pps             []byte
	ppsp            *PPS
	prevDTSFilled   bool
	prevDTS         int64
	pause           int
	reorderedFrames int
}

// Initialize initializes a DTSExtractor.
func (d *DTSExtractor) Initialize() {
}

// NewDTSExtractor allocates a DTSExtractor.
//
// Deprecated: replaced by DTSExtractor.Initialize.
func NewDTSExtractor() *DTSExtractor {
	return &DTSExtractor{}
}

func (d *DTSExtractor) extractInner(au [][]byte, pts int64) (int64, error) {
	var idr []byte
	var nonIDR []byte

	for _, nalu := range au {
		typ := NALUType((nalu[0] >> 1) & 0b111111)

		switch typ {
		case NALUType_SPS_NUT:
			if !bytes.Equal(d.sps, nalu) {
				var spsp SPS
				err := spsp.Unmarshal(nalu)
				if err != nil {
					return 0, fmt.Errorf("invalid SPS: %w", err)
				}

				d.spsp = &spsp
				d.sps = nalu

				// reset state
				if len(d.spsp.MaxNumReorderPics) == 1 {
					d.reorderedFrames = int(d.spsp.MaxNumReorderPics[0])
				} else {
					d.reorderedFrames = 0
				}
				d.pause = d.reorderedFrames
			}

		case NALUType_PPS_NUT:
			if !bytes.Equal(d.pps, nalu) {
				var ppsp PPS
				err := ppsp.Unmarshal(nalu)
				if err != nil {
					return 0, fmt.Errorf("invalid PPS: %w", err)
				}

				d.ppsp = &ppsp
				d.pps = nalu
			}

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

	var ptsDTSDiff int

	switch {
	case idr != nil:
		ptsDTSDiff = 0

	case nonIDR != nil:
		var err error
		ptsDTSDiff, err = getPTSDTSDiff(nonIDR, d.spsp, d.ppsp)
		if err != nil {
			return 0, err
		}

	default:
		return 0, fmt.Errorf("access unit doesn't contain an IDR or non-IDR NALU")
	}

	ptsDTSDiff += d.reorderedFrames

	if ptsDTSDiff < 0 {
		return 0, fmt.Errorf("negative pts-dts difference")
	}

	if d.pause > 0 {
		d.pause--
		if !d.prevDTSFilled {
			var timeDiff int64
			if d.spsp.VUI != nil && d.spsp.VUI.TimingInfo != nil && d.spsp.VUI.TimingInfo.TimeScale != 0 {
				timeDiff = int64(ptsDTSDiff) * 90000 *
					int64(d.spsp.VUI.TimingInfo.NumUnitsInTick) / int64(d.spsp.VUI.TimingInfo.TimeScale)
			} else {
				timeDiff = 9000
			}
			return pts - timeDiff, nil
		}
		return d.prevDTS + 90, nil
	}

	if !d.prevDTSFilled {
		var timeDiff int64
		if d.spsp.VUI != nil && d.spsp.VUI.TimingInfo != nil && d.spsp.VUI.TimingInfo.TimeScale != 0 {
			timeDiff = int64(ptsDTSDiff) * 90000 *
				int64(d.spsp.VUI.TimingInfo.NumUnitsInTick) / int64(d.spsp.VUI.TimingInfo.TimeScale)
		} else {
			timeDiff = 9000
		}
		return pts - timeDiff, nil
	}

	return d.prevDTS + (pts-d.prevDTS)/(int64(ptsDTSDiff)+1), nil
}

// Extract extracts the DTS of a access unit.
func (d *DTSExtractor) Extract(au [][]byte, pts int64) (int64, error) {
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
