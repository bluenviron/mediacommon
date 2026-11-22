package h264

import (
	"bytes"
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

const (
	maxReorderedFrames = 10
	/*
		(max_size(first_mb_in_slice) + max_size(slice_type) + max_size(pic_parameter_set_id) +
		max_size(frame_num) + max_size(pic_order_cnt_lsb)) * 4 / 3 =
		(3 * max_size(golomb) + (max(Log2MaxFrameNumMinus4) + 4) / 8 + (max(Log2MaxPicOrderCntLsbMinus4) + 4) / 8) * 4 / 3 =
		(3 * 4 + 2 + 2) * 4 / 3 = 22,
		plus colour_plane_id (2 bits) and field_pic_flag (1 bit): 23
	*/
	maxBytesToGetPOC = 23
)

// getPictureOrderCount returns the pic_order_cnt_lsb of a frame picture, or reports a field picture.
func getPictureOrderCount(buf []byte, sps *SPS, idr bool) (uint32, bool, error) {
	buf = buf[1:]
	lb := min(len(buf), maxBytesToGetPOC)

	buf = EmulationPreventionRemove(buf[:lb])
	pos := 0

	_, err := bits.ReadGolombUnsigned(buf, &pos) // first_mb_in_slice
	if err != nil {
		return 0, false, err
	}

	_, err = bits.ReadGolombUnsigned(buf, &pos) // slice_type
	if err != nil {
		return 0, false, err
	}

	_, err = bits.ReadGolombUnsigned(buf, &pos) // pic_parameter_set_id
	if err != nil {
		return 0, false, err
	}

	if sps.SeparateColourPlaneFlag {
		_, err = bits.ReadBits(buf, &pos, 2) // colour_plane_id
		if err != nil {
			return 0, false, err
		}
	}

	_, err = bits.ReadBits(buf, &pos, int(sps.Log2MaxFrameNumMinus4+4)) // frame_num
	if err != nil {
		return 0, false, err
	}

	if !sps.FrameMbsOnlyFlag {
		fieldPicFlag, err := bits.ReadFlag(buf, &pos)
		if err != nil {
			return 0, false, err
		}
		if fieldPicFlag {
			return 0, true, nil
		}
	}

	if idr {
		_, err = bits.ReadGolombUnsigned(buf, &pos) // idr_pic_id
		if err != nil {
			return 0, false, err
		}
	}

	picOrderCntLsb, err := bits.ReadBits(buf, &pos, int(sps.Log2MaxPicOrderCntLsbMinus4+4))
	if err != nil {
		return 0, false, err
	}

	return uint32(picOrderCntLsb), false, nil
}

func pictureOrderCountDiff(a uint32, b uint32, sps *SPS) int32 {
	maxVal := uint32(1 << (sps.Log2MaxPicOrderCntLsbMinus4 + 4))
	d := (a - b) & (maxVal - 1)
	if d > (maxVal / 2) {
		return int32(d) - int32(maxVal)
	}
	return int32(d)
}

// DTSExtractor computes DTS from PTS.
type DTSExtractor struct {
	sps             []byte
	spsp            *SPS
	prevDTSFilled   bool
	prevDTS         int64
	randomReceived  bool
	expectedPOC     uint32
	reorderedFrames int
	pause           int
	auCount         int
	pocIncrement    int
	prevPOC         uint32
	prevPrevPOC     uint32
}

// Initialize initializes a DTSExtractor.
func (d *DTSExtractor) Initialize() {
	d.pocIncrement = 2
}

// NewDTSExtractor allocates a DTSExtractor.
//
// Deprecated: replaced by DTSExtractor.Initialize.
func NewDTSExtractor() *DTSExtractor {
	d := &DTSExtractor{}
	d.Initialize()
	return d
}

func (d *DTSExtractor) extractInner(au [][]byte, pts int64) (int64, bool, error) {
	var idr []byte
	var nonIDR []byte
	var seiRecoveryPoint bool
	var pictureBeforeSEIFound bool

outer:
	for _, nalu := range au {
		typ := NALUType(nalu[0] & 0x1F)

		switch typ {
		case NALUTypeSPS:
			if !bytes.Equal(d.sps, nalu) {
				var spsp SPS
				err := spsp.Unmarshal(nalu)
				if err != nil {
					return 0, false, fmt.Errorf("invalid SPS: %w", err)
				}
				d.sps = nalu
				d.spsp = &spsp

				// reset state
				d.randomReceived = false
				d.expectedPOC = 0
				d.reorderedFrames = 0
				d.pause = 0
				d.auCount = 0
				d.pocIncrement = 2
				d.prevPOC = 0
				d.prevPrevPOC = 0
			}

		case NALUTypeIDR:
			idr = nalu
			break outer

		case NALUTypeNonIDR:
			if nonIDR == nil {
				nonIDR = nalu
			}
			if seiRecoveryPoint && !pictureBeforeSEIFound {
				break outer
			}
			pictureBeforeSEIFound = true

		case NALUTypeSEI:
			if !pictureBeforeSEIFound && isSEIRecoveryPoint(nalu) {
				seiRecoveryPoint = true
			}
		}
	}

	if d.spsp == nil {
		return 0, false, fmt.Errorf("SPS not received yet")
	}

	switch d.spsp.PicOrderCntType {
	case 2:
		return pts, false, nil

	case 1:
		return 0, false, fmt.Errorf("pic_order_cnt_type = 1 is not supported yet")
	}

	// Interlaced streams (frame_mbs_only_flag = 0) can contain frame pictures (MBAFF, or frame-coded pictures of a
	// PAFF stream) and field pictures. Frame pictures carry a frame POC and are handled like progressive ones;
	// field pictures are not supported yet and keep the previous behavior (DTS = PTS).
	if !d.spsp.FrameMbsOnlyFlag {
		slice := idr
		if slice == nil {
			slice = nonIDR
		}
		if slice == nil {
			return pts, false, nil
		}

		_, isField, err := getPictureOrderCount(slice, d.spsp, idr != nil)
		if err != nil {
			return 0, false, err
		}
		if isField {
			return pts, false, nil
		}
	}

	if !d.randomReceived && idr == nil && (pictureBeforeSEIFound || !seiRecoveryPoint || nonIDR == nil) {
		return 0, false, fmt.Errorf("random access frame not received yet")
	}

	var ptsDTSDiff int

	switch {
	case idr != nil:
		var err error
		d.expectedPOC, _, err = getPictureOrderCount(idr, d.spsp, true)
		if err != nil {
			return 0, false, err
		}

		if d.pocIncrement == 2 {
			if (d.expectedPOC % 2) != 0 {
				d.pocIncrement = 1
			}

			d.auCount = 1
			d.prevPrevPOC = 0
			d.prevPOC = d.expectedPOC
		}

		ptsDTSDiff = 0

	case nonIDR != nil && seiRecoveryPoint && !pictureBeforeSEIFound:
		// Like CRA in H265: initialize state from first frame after recovery point
		var err error
		d.expectedPOC, _, err = getPictureOrderCount(nonIDR, d.spsp, false)
		if err != nil {
			return 0, false, err
		}

		if d.pocIncrement == 2 {
			if (d.expectedPOC % 2) != 0 {
				d.pocIncrement = 1
			}

			d.auCount = 1
			d.prevPrevPOC = 0
			d.prevPOC = d.expectedPOC
		}

		ptsDTSDiff = 0

	case nonIDR != nil:
		poc, _, err := getPictureOrderCount(nonIDR, d.spsp, false)
		if err != nil {
			return 0, false, err
		}

		if d.auCount < 5 && d.pocIncrement == 2 {
			if (poc % 2) != 0 {
				d.pocIncrement = 1
				d.expectedPOC /= 2

				if d.reorderedFrames != 0 {
					increase := d.reorderedFrames
					if (d.reorderedFrames + increase) > maxReorderedFrames {
						return 0, false, fmt.Errorf("too many reordered frames (%d)", d.reorderedFrames+increase)
					}

					d.reorderedFrames += increase
					d.pause += increase
				}
			} else if d.auCount >= 2 && (poc%4) == 0 && poc == (d.prevPOC+4) && d.prevPOC == (d.prevPrevPOC+4) {
				d.pocIncrement = 4
				d.expectedPOC *= 2
			}

			d.auCount++
			d.prevPrevPOC = d.prevPOC
			d.prevPOC = poc
		}

		d.expectedPOC += uint32(d.pocIncrement)
		d.expectedPOC &= ((1 << (d.spsp.Log2MaxPicOrderCntLsbMinus4 + 4)) - 1)

		ptsDTSDiff = int(pictureOrderCountDiff(poc, d.expectedPOC, d.spsp)) / d.pocIncrement

	default:
		return 0, false, fmt.Errorf("access unit doesn't contain an IDR or non-IDR NALU")
	}

	d.randomReceived = true
	ptsDTSDiff += d.reorderedFrames

	switch {
	case ptsDTSDiff > (2*d.reorderedFrames + 1):
		increase := ptsDTSDiff - (2*d.reorderedFrames + 1)
		if (d.reorderedFrames + increase) > maxReorderedFrames {
			return 0, false, fmt.Errorf("too many reordered frames (%d)", d.reorderedFrames+increase)
		}

		d.reorderedFrames += increase
		d.pause += increase
		ptsDTSDiff += increase

	case ptsDTSDiff < 0:
		increase := -ptsDTSDiff
		if (d.reorderedFrames + increase) > maxReorderedFrames {
			return 0, false, fmt.Errorf("too many reordered frames (%d)", d.reorderedFrames+increase)
		}

		d.reorderedFrames += increase
		d.pause += increase
		ptsDTSDiff += increase
	}

	if d.pause > 0 {
		d.pause--
		if !d.prevDTSFilled {
			return pts, true, nil
		}
		return d.prevDTS + 90, true, nil
	}

	if !d.prevDTSFilled {
		return pts, true, nil
	}

	return d.prevDTS + (pts-d.prevDTS)/(int64(ptsDTSDiff)+1), false, nil
}

// Extract extracts the DTS of an access unit.
// AU is assumed to contain byte slices each with at least 1 byte.
func (d *DTSExtractor) Extract(au [][]byte, pts int64) (int64, error) {
	dts, skipChecks, err := d.extractInner(au, pts)
	if err != nil {
		return 0, err
	}

	if !skipChecks && dts > pts {
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
