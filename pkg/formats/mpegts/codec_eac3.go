package mpegts

import (
	"github.com/asticode/go-astits"
)

// CodecEAC3 is an Enhanced AC-3 (Dolby Digital Plus) codec.
// Specification: ETSI TS 102 366 V1.4.1, Annex E
type CodecEAC3 struct {
	SampleRate   int
	ChannelCount int
}

// IsVideo implements Codec.
func (CodecEAC3) IsVideo() bool {
	return false
}

func (*CodecEAC3) isCodec() {}

func (c CodecEAC3) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypeEAC3Audio,
	}, nil
}
