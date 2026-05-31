package vp9

// IsRandomAccess checks whether a frame can be randomly accessed.
func IsRandomAccess(frame []byte) bool {
	var h Header
	err := h.Unmarshal(frame)
	if err != nil {
		return false
	}
	return !h.NonKeyFrame
}
