package av1

// ContainsKeyFrame checks whether OBUs contain a key frame.
func ContainsKeyFrame(obus [][]byte) (bool, error) {
	if len(obus) == 0 {
		return false, nil
	}

	var h OBUHeader
	err := h.Unmarshal(obus[0])
	if err != nil {
		return false, err
	}

	return (h.Type == OBUTypeSequenceHeader), nil
}
