package mpegts

import (
	"time"
)

const (
	clockRate = 90000
)

// avoid an int64 overflow and preserve resolution by splitting division into two parts:
// first add the integer part, then the decimal part.
func multiplyAndDivide(v, m, d time.Duration) time.Duration {
	secs := v / d
	dec := v % d
	return (secs*m + dec*m/d)
}

// TimeDecoder is a MPEG-TS timestamp decoder.
//
// Deprecated: replaced by TimeDecoder2.
type TimeDecoder struct {
	wrapped *TimeDecoder2
}

// NewTimeDecoder allocates a TimeDecoder.
//
// Deprecated: replaced by NewTimeDecoder2.
func NewTimeDecoder(start int64) *TimeDecoder {
	td := NewTimeDecoder2()
	td.Decode(start)

	return &TimeDecoder{
		wrapped: td,
	}
}

// Decode decodes a MPEG-TS timestamp.
func (d *TimeDecoder) Decode(ts int64) time.Duration {
	return multiplyAndDivide(time.Duration(d.wrapped.Decode(ts)), time.Second, clockRate)
}
