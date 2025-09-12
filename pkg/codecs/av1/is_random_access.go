package av1

import (
	"fmt"
)

// IsRandomAccess2 checks whether a temporal unit can be randomly accessed.
func IsRandomAccess2(tu [][]byte) bool {
	for _, obu := range tu {
		if len(obu) != 0 {
			typ := OBUType((obu[0] >> 3) & 0b1111)

			if typ == OBUTypeSequenceHeader {
				return true
			}
		}
	}

	return false
}

// IsRandomAccess checks whether a temporal unit can be randomly accessed.
//
// Deprecated: replaced by IsRandomAccess2.
func IsRandomAccess(tu [][]byte) (bool, error) {
	if len(tu) == 0 {
		return false, fmt.Errorf("temporal unit is empty")
	}

	typ := OBUType((tu[0][0] >> 3) & 0b1111)

	return (typ == OBUTypeSequenceHeader), nil
}
