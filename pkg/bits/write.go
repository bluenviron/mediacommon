package bits

// WriteBits writes N bits.
//
// Deprecated: replaced by WriteBitsUnsafe.
func WriteBits(buf []byte, pos *int, v uint64, n int) {
	WriteBitsUnsafe(buf, pos, v, n)
}

// WriteBitsUnsafe writes N bits.
func WriteBitsUnsafe(buf []byte, pos *int, v uint64, n int) {
	res := 8 - (*pos & 0x07)
	if n < res {
		buf[*pos>>0x03] |= byte(v << (res - n))
		*pos += n
		return
	}

	buf[*pos>>3] |= byte(v >> (n - res))
	*pos += res
	n -= res

	for n >= 8 {
		buf[*pos>>3] = byte(v >> (n - 8))
		*pos += 8
		n -= 8
	}

	if n > 0 {
		buf[*pos>>3] = byte((v & (1<<n - 1)) << (8 - n))
		*pos += n
	}
}

// WriteFlagUnsafe writes a boolean flag.
func WriteFlagUnsafe(buf []byte, pos *int, v bool) {
	if v {
		WriteBitsUnsafe(buf, pos, 1, 1)
	} else {
		WriteBitsUnsafe(buf, pos, 0, 1)
	}
}

// WriteGolombUnsigned writes an unsigned golomb-encoded value.
func WriteGolombUnsignedUnsafe(buf []byte, pos *int, v uint32) {
	v += 1
	bitCount := 31

	for bitCount > 0 {
		bit := (v >> bitCount) & 0x01
		if bit != 0 {
			break
		}
		bitCount--
	}

	*pos += bitCount
	WriteBitsUnsafe(buf, pos, uint64(v), bitCount+1)
}
