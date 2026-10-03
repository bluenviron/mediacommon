package h264

import "github.com/bluenviron/mediacommon/v2/pkg/bits"

const seiPayloadTypeRecoveryPoint = 6

// isSEIRecoveryPoint checks whether a SEI NALU contains a recovery point
// message whose recovery_frame_cnt is 0. It assumes the NALU type is NALUTypeSEI.
func isSEIRecoveryPoint(nalu []byte) bool {
	if len(nalu) < 2 {
		return false
	}

	rbsp := EmulationPreventionRemove(nalu[1:])
	pos := 0
	found := false

	for pos < len(rbsp) {
		if rbsp[pos] == 0x80 && pos == len(rbsp)-1 {
			return found
		}

		payloadType, next, ok := readSEIByteValue(rbsp, pos)
		if !ok {
			return false
		}
		pos = next

		payloadSize, next, ok := readSEIByteValue(rbsp, pos)
		if !ok {
			return false
		}
		pos = next

		if payloadSize > len(rbsp)-pos {
			return false
		}

		if payloadType == seiPayloadTypeRecoveryPoint && isSEIRecoveryPointPayload(rbsp[pos:pos+payloadSize]) {
			found = true
		}
		pos += payloadSize
	}

	return false
}

// isSEIRecoveryPointPayload checks whether a recovery point message identifies
// an immediate random access point.
func isSEIRecoveryPointPayload(payload []byte) bool {
	pos := 0
	recoveryFrameCnt, err := bits.ReadGolombUnsigned(payload, &pos)
	if err != nil {
		return false
	}

	_, err = bits.ReadFlag(payload, &pos) // exact_match_flag
	if err != nil {
		return false
	}

	_, err = bits.ReadFlag(payload, &pos) // broken_link_flag
	if err != nil {
		return false
	}

	_, err = bits.ReadBits(payload, &pos, 2) // changing_slice_group_idc
	if err != nil {
		return false
	}

	if pos%8 != 0 {
		one, err := bits.ReadFlag(payload, &pos) // payload_bit_equal_to_one
		if err != nil || !one {
			return false
		}
		for pos%8 != 0 {
			zero, err := bits.ReadFlag(payload, &pos) // payload_bit_equal_to_zero
			if err != nil || zero {
				return false
			}
		}
	}

	return pos == len(payload)*8 && recoveryFrameCnt == 0
}

// readSEIByteValue reads a SEI payloadType or payloadSize value.
func readSEIByteValue(rbsp []byte, pos int) (int, int, bool) {
	var value int
	for pos < len(rbsp) {
		b := rbsp[pos]
		pos++
		if value > int(^uint(0)>>1)-int(b) {
			return 0, pos, false
		}
		value += int(b)
		if b != 0xFF {
			return value, pos, true
		}
	}
	return 0, pos, false
}
