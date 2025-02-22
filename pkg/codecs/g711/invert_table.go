package g711

func invertTable(in [256]uint16, mask uint8) [16384]uint8 {
	var res [16384]uint8

	j := uint16(1)
	res[8192] = mask

	for i := uint8(0); i < 127; i++ {
		v1 := in[i^mask]
		v2 := in[(i+1)^mask]
		v := (v1 + v2 + 4) >> 3

		for j < v {
			res[8192-j] = (i ^ (mask ^ 0x80))
			res[8192+j] = (i ^ mask)
			j++
		}
	}

	for j < 8192 {
		res[8192-j] = 127 ^ (mask ^ 0x80)
		res[8192+j] = 127 ^ mask
		j++
	}

	res[0] = res[1]

	return res
}
