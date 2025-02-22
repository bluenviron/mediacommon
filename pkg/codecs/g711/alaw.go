// Package g711 contains utilities to work with the G711 codec.
package g711

import "fmt"

var alawDecodeTable = func() [256]uint16 {
	var ret [256]uint16
	for i := 0; i < 256; i++ {
		v := i ^ 0x55

		mantissa := uint16(v & 0x0F)
		exponent := uint16((v & 0x70) >> 4)
		sign := v & 0x80

		if exponent != 0 {
			mantissa = (mantissa*2 + 1 + 32) << (exponent + 2)
		} else {
			mantissa = (mantissa*2 + 1) << 3
		}

		if sign != 0 {
			ret[i] = mantissa
		} else {
			ret[i] = -mantissa
		}
	}
	return ret
}()

var alawEncodeTable = invertTable(alawDecodeTable, 0xD5)

// DecodeAlaw decodes 8-bit G711 samples (A-law) into 16-bit LPCM samples.
//
// Deprecated: replaced by Alaw.Unmarshal.
func DecodeAlaw(in []byte) []byte {
	var lpcm Alaw
	lpcm.Unmarshal(in)
	return lpcm
}

// Alaw is a list of 16-bit LPCM samples that can be encoded/decoded from/to the A-law variant of the G711 codec.
type Alaw []byte

// Unmarshal decodes A-law samples.
func (c *Alaw) Unmarshal(enc []byte) {
	*c = make([]byte, len(enc)*2)
	for i, sample := range enc {
		(*c)[i*2] = uint8(alawDecodeTable[sample] >> 8)
		(*c)[(i*2)+1] = uint8(alawDecodeTable[sample])
	}
}

// Marshal encodes 16-bit LPCM samples with the A-law G711 codec.
func (c Alaw) Marshal() ([]byte, error) {
	if (len(c) % 2) != 0 {
		return nil, fmt.Errorf("wrong sample size")
	}

	le := len(c) / 2
	enc := make([]byte, le)

	for i := 0; i < le; i++ {
		sample := uint16(c[i*2])<<8 | uint16(c[(i*2)+1])
		enc[i] = alawEncodeTable[(sample+32768)>>2]
	}

	return enc, nil
}
