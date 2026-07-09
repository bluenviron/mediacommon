package h264

// IsRandomAccess checks whether the access unit can be randomly accessed.
func IsRandomAccess(au [][]byte) bool {
	for _, nalu := range au {
		typ := NALUType(nalu[0] & 0x1F)
		switch typ {
		case NALUTypeIDR:
			return true
		case NALUTypeSEI:
			if isSEIRecoveryPoint(nalu) {
				return true
			}
		}
	}
	return false
}
