package h264

// IsRandomAccess checks whether the access unit can be randomly accessed.
// AU is assumed to contain byte slices each with at least 1 byte.
func IsRandomAccess(au [][]byte) bool {
	var recoveryPoint bool
	var pictureBeforeSEIFound bool

	for _, nalu := range au {
		typ := NALUType(nalu[0] & 0x1F)
		switch typ {
		case NALUTypeIDR:
			return true

		case NALUTypeSEI:
			if !pictureBeforeSEIFound && isSEIRecoveryPoint(nalu) {
				recoveryPoint = true
			}

		case NALUTypeNonIDR:
			if !pictureBeforeSEIFound && recoveryPoint {
				return true
			}
			pictureBeforeSEIFound = true
		}
	}

	return false
}
