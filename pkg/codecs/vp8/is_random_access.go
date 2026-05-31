package vp8

// IsRandomAccess checks whether a frame can be randomly accessed.
func IsRandomAccess(frame []byte) bool {
	return len(frame) > 0 && (frame[0]&0x01) == 0
}
