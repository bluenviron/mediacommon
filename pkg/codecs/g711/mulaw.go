package g711

import "fmt"

var mulawDecodeTable = func() [256]uint16 {
	var ret [256]uint16
	for i := 0; i < 256; i++ {
		v := ^i

		tmp := (((uint16(v) & 0x0F) << 3) + 0x84) << ((v & 0x70) >> 4)
		sign := v & 0x80

		if sign != 0 {
			ret[i] = 0x84 - tmp
		} else {
			ret[i] = tmp - 0x84
		}
	}
	return ret
}()

var mulawEncodeTable = invertTable(mulawDecodeTable, 0xFF)

// DecodeMulaw decodes 8-bit G711 samples (MU-law) into 16-bit LPCM samples.
//
// Deprecated: replaced by Mulaw.Unmarshal.
func DecodeMulaw(in []byte) []byte {
	var lpcm Mulaw
	lpcm.Unmarshal(in)
	return lpcm
}

// Mulaw is a list of 16-bit LPCM samples that can be encoded/decoded from/to the MU-law variant of the G711 codec.
type Mulaw []byte

// Unmarshal decodes MU-law samples.
func (c *Mulaw) Unmarshal(enc []byte) {
	*c = make([]byte, len(enc)*2)
	for i, sample := range enc {
		(*c)[i*2] = uint8(mulawDecodeTable[sample] >> 8)
		(*c)[(i*2)+1] = uint8(mulawDecodeTable[sample])
	}
}

// Marshal encodes 16-bit LPCM samples with the MU-law G711 codec.
func (c Mulaw) Marshal() ([]byte, error) {
	if (len(c) % 2) != 0 {
		return nil, fmt.Errorf("wrong sample size")
	}

	le := len(c) / 2
	enc := make([]byte, le)

	for i := 0; i < le; i++ {
		sample := uint16(c[i*2])<<8 | uint16(c[(i*2)+1])
		enc[i] = mulawEncodeTable[(sample+32768)>>2]
	}

	return enc, nil
}
