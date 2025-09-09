package av1

// OBUType is an OBU type.
type OBUType uint8

// OBU types.
// Specification: AV1 Bitstream & Decoding Process, section 6.2.2
const (
	OBUTypeSequenceHeader    OBUType = 1
	OBUTypeTemporalDelimiter OBUType = 2
)
