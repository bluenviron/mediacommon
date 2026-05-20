package mp4

import (
	amp4 "github.com/abema/go-mp4"
)

func init() { //nolint:gochecknoinits
	// Register fLaC as an AudioSampleEntry type.
	amp4.AddAnyTypeBoxDef(&amp4.AudioSampleEntry{}, amp4.StrToBoxType("fLaC"))
	// Register dfLa box type.
	amp4.AddBoxDef(&DfLa{}, 0)
}

// DfLa is the FLAC-specific box (FLACSpecificBox).
// Specification: isoflac.txt
type DfLa struct {
	amp4.FullBox `mp4:"0,extend"`

	// Blocks contains FLAC metadata blocks (at minimum the STREAMINFO block).
	Blocks []byte `mp4:"1,size=8"`
}

// GetType returns the box type for dfLa.
func (*DfLa) GetType() amp4.BoxType {
	return amp4.StrToBoxType("dfLa")
}
