// Package g711 contains utilities to work with the G711 codec.
package g711

var mulawTable = func() [256]uint16 {
	var ret [256]uint16
	for i := 0; i < 256; i++ {
		v := ^i

		tmp := (((uint16(v) & 0x0F) << 3) + 0x84) << ((v & 0x70) >> 4)

		if (v & 0x80) != 0 {
			ret[i] = 0x84 - tmp
		} else {
			ret[i] = tmp - 0x84
		}
	}
	return ret
}()

var alawTable = func() [256]uint16 {
	var ret [256]uint16
	for i := 0; i < 256; i++ {
		v := i ^ 0x55

		t := uint16(v) & 0x0F
		seg := (uint16(v) & 0x70) >> 4

		if seg != 0 {
			t = (t*2 + 1 + 32) << (seg + 2)
		} else {
			t = (t*2 + 1) << 3
		}

		if (v & 0x80) != 0 {
			ret[i] = t
		} else {
			ret[i] = -t
		}
	}
	return ret
}()

// DecodeMulaw decodes 8-bit G711 samples (MU-law) into 16-bit LPCM samples.
func DecodeMulaw(in []byte) []byte {
	out := make([]byte, len(in)*2)
	for i, sample := range in {
		out[i*2] = uint8(mulawTable[sample] >> 8)
		out[(i*2)+1] = uint8(mulawTable[sample])
	}
	return out
}

// DecodeAlaw decodes 8-bit G711 samples (A-law) into 16-bit LPCM samples.
func DecodeAlaw(in []byte) []byte {
	out := make([]byte, len(in)*2)
	for i, sample := range in {
		out[i*2] = uint8(alawTable[sample] >> 8)
		out[(i*2)+1] = uint8(alawTable[sample])
	}
	return out
}
