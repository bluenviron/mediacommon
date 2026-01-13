package codecs

// EAC3 is the E-AC-3 (Enhanced AC-3 / Dolby Digital Plus) codec.
// Fields are based on the dec3 (EC3SpecificBox) structure per ETSI TS 102 366.
type EAC3 struct {
	SampleRate   int
	ChannelCount int

	// DataRate is the data rate in kbps (from dec3 data_rate field, 13 bits).
	DataRate uint16

	// Fields below are for the first (primary) independent substream.
	// For multi-substream content, additional parsing would be needed.

	// Asvc indicates if this is an associated audio service (1 bit).
	Asvc bool

	// Bsmod is the bit stream mode (3 bits), indicating the type of service.
	Bsmod uint8

	// Acmod is the audio coding mode (3 bits), indicating channel configuration.
	Acmod uint8

	// LfeOn indicates if the LFE (Low Frequency Effects) channel is present (1 bit).
	LfeOn bool

	// NumDepSub is the number of dependent substreams associated with this
	// independent substream (4 bits).
	NumDepSub uint8

	// ChanLoc is the channel location bitmap for dependent substreams (9 bits).
	// Only valid when NumDepSub > 0.
	ChanLoc uint16
}

// IsVideo implements Codec.
func (EAC3) IsVideo() bool {
	return false
}

func (*EAC3) isCodec() {}
