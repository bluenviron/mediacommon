package mpegts

const (
	maximum           = 0x1FFFFFFFF // 33 bits
	negativeThreshold = 0x1FFFFFFFF / 2
)

// TimeDecoder is a MPEG-TS timestamp decoder.
type TimeDecoder struct {
	initialized bool
	overall     int64
	prev        int64
}

// NewTimeDecoder allocates a TimeDecoder.
func NewTimeDecoder() *TimeDecoder {
	return &TimeDecoder{}
}

// Decode decodes a MPEG-TS timestamp.
func (d *TimeDecoder) Decode(ts int64) int64 {
	if !d.initialized {
		d.initialized = true
		d.prev = ts
	}

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
