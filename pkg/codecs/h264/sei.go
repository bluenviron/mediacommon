package h264

import "github.com/bluenviron/mediacommon/v2/pkg/bits"

// isSEIRecoveryPoint checks if a SEI NALU contains a recovery point message
// (payload type 6) that can be used as a random access point.
func isSEIRecoveryPoint(nalu []byte) bool {
	typ := NALUType(nalu[0] & 0x1F)
	if typ != NALUTypeSEI {
		return false
	}

	pos := 1 // skip NALU header byte

	for pos < len(nalu) {
		payloadType, p := readSEIByteValue(nalu, pos)
		pos = p

		payloadSize, p := readSEIByteValue(nalu, pos)
		pos = p

		end := pos + payloadSize
		if end > len(nalu) {
			end = len(nalu)
		}

		if payloadType == 6 && isSEIRecoveryPointPayload(nalu[pos:end]) {
			return true
		}

		pos = end
	}

	return false
}

// isSEIRecoveryPointPayload checks if a recovery point message payload
// identifies a random access point, i.e. its recovery_frame_cnt field is 0.
// Truncated or unparseable payloads are assumed to identify one as well.
func isSEIRecoveryPointPayload(payload []byte) bool {
	payload = EmulationPreventionRemove(payload)

	// recovery_point_sei_message() is at least 8 bytes long
	if len(payload) < 8 {
		return true
	}

	pos := 48 // skip bits_per_second and picture_bits_per_second

	cpbCntFlag, err := bits.ReadFlag(payload, &pos)
	if err != nil {
		return true
	}

	if cpbCntFlag {
		// initial_cpb_removal_delay
		_, err = bits.ReadGolombUnsigned(payload, &pos)
		if err != nil {
			return true
		}

		// cpb_removal_delay_increment
		_, err = bits.ReadGolombUnsigned(payload, &pos)
		if err != nil {
			return true
		}
	}

	timeCodeCnt, err := bits.ReadFlag(payload, &pos)
	if err != nil {
		return true
	}

	if timeCodeCnt {
		err = skipSEITimeCode(payload, &pos)
		if err != nil {
			return true
		}
	}

	// frames_until_next_recovery_point
	_, err = bits.ReadBits(payload, &pos, 8)
	if err != nil {
		return true
	}

	recoveryFrameCnt, err := bits.ReadBits(payload, &pos, 5) // recovery_frame_cnt
	if err != nil {
		return true
	}

	recoveryFrameCntFlag, err := bits.ReadFlag(payload, &pos) // recovery_frame_cnt_flag
	if err != nil {
		return true
	}

	if recoveryFrameCntFlag {
		recoveryFrameCntMSB, err := bits.ReadBits(payload, &pos, 3) // recovery_frame_cnt_msb
		if err != nil {
			return true
		}
		recoveryFrameCnt |= recoveryFrameCntMSB << 5
	}

	return recoveryFrameCnt == 0
}

// skipSEITimeCode skips the time_code() structure (H.264 spec, 7.32.8).
func skipSEITimeCode(payload []byte, pos *int) error {
	// clock_timestamp_flag + 8 flag fields, 1 bit each
	v, err := bits.ReadBits(payload, pos, 9)
	if err != nil {
		return err
	}

	clockTimestampFlag := v&0x100 != 0
	localHoursFlag := v&0x08 != 0
	secondsFlag := v&0x04 != 0
	minutesFlag := v&0x02 != 0
	hoursFlag := v&0x01 != 0

	if !clockTimestampFlag {
		// pic_struct + count
		_, err = bits.ReadBits(payload, pos, 4)
		return err
	}

	// second_value
	if _, err = bits.ReadBits(payload, pos, 6); err != nil {
		return err
	}
	if secondsFlag {
		// second_value
		if _, err = bits.ReadBits(payload, pos, 6); err != nil {
			return err
		}
	}
	if minutesFlag {
		// minute_value
		if _, err = bits.ReadBits(payload, pos, 6); err != nil {
			return err
		}
	}
	if hoursFlag {
		// hour_value
		if _, err = bits.ReadBits(payload, pos, 5); err != nil {
			return err
		}
	}
	if localHoursFlag {
		// local_hour_value + local_hour_minute_offset + local_hour_second_offset
		if _, err = bits.ReadBits(payload, pos, 17); err != nil {
			return err
		}
		localHourSecondFlag, err := bits.ReadFlag(payload, pos)
		if err != nil {
			return err
		}
		if localHourSecondFlag {
			// local_hour_second_value
			if _, err = bits.ReadBits(payload, pos, 6); err != nil {
				return err
			}
		}
	}

	// time_offset_value
	_, err = bits.ReadBits(payload, pos, 32)
	return err
}

// readSEIByteValue reads a SEI payloadType/payloadSize value, where each byte
// equal to 0xFF contributes 255 and the first byte different from 0xFF
// terminates the value (adding its own value).
func readSEIByteValue(nalu []byte, pos int) (int, int) {
	var v int
	for pos < len(nalu) {
		b := nalu[pos]
		pos++
		v += int(b)
		if b != 0xFF {
			break
		}
	}
	return v, pos
}
