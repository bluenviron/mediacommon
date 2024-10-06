package mpegts

const (
	maximum           = 0x1FFFFFFFF // 33 bits
	negativeThreshold = 0x1FFFFFFFF / 2
)

// TimeDecoder2 is a MPEG-TS timestamp decoder.
type TimeDecoder2 struct {
	overall int64
	prev    int64
}

// NewTimeDecoder2 allocates a TimeDecoder.
func NewTimeDecoder2(start int64) *TimeDecoder2 {
	return &TimeDecoder2{
		prev: start,
	}
}

// Decode decodes a MPEG-TS timestamp.
func (d *TimeDecoder2) Decode(ts int64) int64 {
	diff := (ts - d.prev) & maximum

	// negative difference
	if diff > negativeThreshold {
		diff = (d.prev - ts) & maximum
		d.prev = ts
		d.overall -= diff
	} else {
		d.prev = ts
		d.overall += diff
	}

	return d.overall
}
