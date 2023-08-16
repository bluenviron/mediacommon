package h264

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bluenviron/mediacommon/pkg/bits"
)

func getPictureOrderCount(buf []byte, sps *SPS) (uint32, error) {
	if len(buf) < 6 {
		return 0, fmt.Errorf("not enough bits")
	}

	buf = EmulationPreventionRemove(buf[:6])

	buf = buf[1:]
	pos := 0

	_, err := bits.ReadGolombUnsigned(buf, &pos) // first_mb_in_slice
	if err != nil {
		return 0, err
	}

	_, err = bits.ReadGolombUnsigned(buf, &pos) // slice_type
	if err != nil {
		return 0, err
	}

	_, err = bits.ReadGolombUnsigned(buf, &pos) // pic_parameter_set_id
	if err != nil {
		return 0, err
	}

	_, err = bits.ReadBits(buf, &pos, int(sps.Log2MaxFrameNumMinus4+4)) // frame_num
	if err != nil {
		return 0, err
	}

	if !sps.FrameMbsOnlyFlag {
		return 0, fmt.Errorf("frame_mbs_only_flag = 0 is not supported")
	}

	picOrderCntLsb, err := bits.ReadBits(buf, &pos, int(sps.Log2MaxPicOrderCntLsbMinus4+4))
	if err != nil {
		return 0, err
	}

	return uint32(picOrderCntLsb), nil
}

func getPictureOrderCountDiff(a uint32, b uint32, sps *SPS) int32 {
	max := uint32(1 << (sps.Log2MaxPicOrderCntLsbMinus4 + 4))
	d := (a - b) & (max - 1)
	if d > (max / 2) {
		return int32(d) - int32(max)
	}
	return int32(d)
}

// DTSExtractor allows to extract DTS from PTS.
type DTSExtractor struct {
	sps             []byte
	spsp            *SPS
	prevDTSFilled   bool
	prevDTS         time.Duration
	expectedPOC     uint32
	reorderedFrames int
	pauseDTS        int
	pocIncrement    int
}

// NewDTSExtractor allocates a DTSExtractor.
func NewDTSExtractor() *DTSExtractor {
	return &DTSExtractor{
		pocIncrement: 2,
	}
}

func (d *DTSExtractor) extractInner(au [][]byte, pts time.Duration) (time.Duration, bool, error) {
	var idr []byte
	var nonIDR []byte

	for _, nalu := range au {
		typ := NALUType(nalu[0] & 0x1F)
		switch typ {
		case NALUTypeSPS:
			if !bytes.Equal(d.sps, nalu) {
				var spsp SPS
				err := spsp.Unmarshal(nalu)
				if err != nil {
					return 0, false, fmt.Errorf("invalid SPS: %v", err)
				}
				d.sps = nalu
				d.spsp = &spsp

				// reset state
				d.reorderedFrames = 0
				d.pocIncrement = 2
			}

		case NALUTypeIDR:
			idr = nalu

		case NALUTypeNonIDR:
			nonIDR = nalu
		}
	}

	if d.spsp == nil {
		return 0, false, fmt.Errorf("SPS not received yet")
	}

	if d.spsp.PicOrderCntType == 2 {
		return pts, false, nil
	}

	if d.spsp.PicOrderCntType == 1 {
		return 0, false, fmt.Errorf("pic_order_cnt_type = 1 is not supported yet")
	}

	switch {
	case idr != nil:
		d.expectedPOC = 0
		d.pauseDTS = 0

		if !d.prevDTSFilled || d.reorderedFrames == 0 {
			return pts, false, nil
		}

		return d.prevDTS + (pts-d.prevDTS)/time.Duration(d.reorderedFrames+1), false, nil

	case nonIDR != nil:
		d.expectedPOC += uint32(d.pocIncrement)
		d.expectedPOC &= ((1 << (d.spsp.Log2MaxPicOrderCntLsbMinus4 + 4)) - 1)

		if d.pauseDTS > 0 {
			d.pauseDTS--
			return d.prevDTS + 1*time.Millisecond, true, nil
		}

		poc, err := getPictureOrderCount(nonIDR, d.spsp)
		if err != nil {
			return 0, false, err
		}

		if d.pocIncrement == 2 && (poc%2) != 0 {
			d.pocIncrement = 1
			d.expectedPOC /= 2
		}

		pocDiff := int(getPictureOrderCountDiff(poc, d.expectedPOC, d.spsp)) + d.reorderedFrames*d.pocIncrement

		if pocDiff < 0 {
			if pocDiff < -20 {
				return 0, false, fmt.Errorf("POC difference between frames is too big (%d)", pocDiff)
			}

			// this happens when there are B-frames immediately following an IDR frame
			d.reorderedFrames -= pocDiff
			d.pauseDTS = -pocDiff
			return d.prevDTS + 1*time.Millisecond, true, nil
		}

		if pocDiff == 0 {
			return pts, false, nil
		}

		if pocDiff > 20 {
			return 0, false, fmt.Errorf("POC difference between frames is too big (%d)", pocDiff)
		}

		reorderedFrames := (pocDiff)/d.pocIncrement - d.reorderedFrames
		if reorderedFrames > d.reorderedFrames {
			// reordered frames detected, add them to the count and pause DTS
			d.pauseDTS = (reorderedFrames - d.reorderedFrames - 1)
			d.reorderedFrames = reorderedFrames
			return d.prevDTS + 1*time.Millisecond, false, nil
		}

		return d.prevDTS + (pts-d.prevDTS)*time.Duration(d.pocIncrement)/time.Duration(pocDiff+d.pocIncrement), false, nil

	default:
		return 0, false, fmt.Errorf("access unit doesn't contain an IDR or non-IDR NALU")
	}
}

// Extract extracts the DTS of an access unit.
func (d *DTSExtractor) Extract(au [][]byte, pts time.Duration) (time.Duration, error) {
	dts, skipChecks, err := d.extractInner(au, pts)
	if err != nil {
		return 0, err
	}

	if !skipChecks {
		if dts > pts {
			return 0, fmt.Errorf("DTS is greater than PTS")
		}
	}

	if d.prevDTSFilled && dts <= d.prevDTS {
		return 0, fmt.Errorf("DTS is not monotonically increasing, was %v, now is %v",
			d.prevDTS, dts)
	}

	d.prevDTS = dts
	d.prevDTSFilled = true

	return dts, err
}
