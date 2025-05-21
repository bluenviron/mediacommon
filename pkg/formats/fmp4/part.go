package fmp4

import (
	"io"

	amp4 "github.com/abema/go-mp4"
	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
)

const (
	trunFlagDataOffsetPreset                       = 0x01
	trunFlagSampleDurationPresent                  = 0x100
	trunFlagSampleSizePresent                      = 0x200
	trunFlagSampleFlagsPresent                     = 0x400
	trunFlagSampleCompositionTimeOffsetPresentOrV1 = 0x800

	sampleFlagIsNonSyncSample = 1 << 16
)

// Part is a fMP4 part.
type Part struct {
	SequenceNumber uint32
	Tracks         []*PartTrack
}

// Marshal encodes a fMP4 part.
func (p *Part) Marshal(w io.WriteSeeker) error {
	/*
		|moof|
		|    |mfhd|
		|    |traf|
		|    |traf|
		|    |....|
		|mdat|
	*/

	mw := &imp4.Writer{W: w}
	mw.Initialize()

	moofOffset, err := mw.WriteBoxStart(&amp4.Moof{}) // <moof>
	if err != nil {
		return err
	}

	_, err = mw.WriteBox(&amp4.Mfhd{ // <mfhd/>
		SequenceNumber: p.SequenceNumber,
	})
	if err != nil {
		return err
	}

	trackLen := len(p.Tracks)
	truns := make([]*amp4.Trun, trackLen)
	trunOffsets := make([]int, trackLen)
	dataOffsets := make([]int, trackLen)
	dataSize := 0

	for i, track := range p.Tracks {
		var trun *amp4.Trun
		var trunOffset int
		trun, trunOffset, err = track.marshal(mw)
		if err != nil {
			return err
		}

		dataOffsets[i] = dataSize

		for _, sample := range track.Samples {
			dataSize += len(sample.Payload)
		}

		truns[i] = trun
		trunOffsets[i] = trunOffset
	}

	err = mw.WriteBoxEnd() // </moof>
	if err != nil {
		return err
	}

	mdat := &amp4.Mdat{} // <mdat/>
	mdat.Data = make([]byte, dataSize)
	pos := 0

	for _, track := range p.Tracks {
		for _, sample := range track.Samples {
			pos += copy(mdat.Data[pos:], sample.Payload)
		}
	}

	mdatOffset, err := mw.WriteBox(mdat)
	if err != nil {
		return err
	}

	for i := range p.Tracks {
		truns[i].DataOffset = int32(dataOffsets[i] + mdatOffset - moofOffset + 8)
		err = mw.RewriteBox(trunOffsets[i], truns[i])
		if err != nil {
			return err
		}
	}

	return nil
}
