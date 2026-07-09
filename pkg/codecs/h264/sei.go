package h264

// IsSEIRecoveryPoint checks whether a NALU is a recovery point SEI message.
func IsSEIRecoveryPoint(nalu []byte) bool {
	return isSEIRecoveryPoint(nalu)
}

// isSEIRecoveryPoint checks if a SEI NALU contains a recovery point message (payload type 6).
func isSEIRecoveryPoint(nalu []byte) bool {
	typ := NALUType(nalu[0] & 0x1F)
	if typ != NALUTypeSEI {
		return false
	}

	pos := 1 // skip NALU header byte

	for pos < len(nalu) {
		payloadType, p := readSEIByteValue(nalu, pos)
		pos = p

		if payloadType == 6 {
			return true
		}

		payloadSize, p := readSEIByteValue(nalu, pos)
		pos = p

		pos += payloadSize
	}

	return false
}

// readSEIByteValue reads a SEI ff-style encoded value (payloadType/payloadSize),
// where each byte equal to 0xFF contributes 255 and the first byte different from
// 0xFF terminates the value (adding its own value). It returns the value and the
// updated position.
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
