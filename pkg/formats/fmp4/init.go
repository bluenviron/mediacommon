package fmp4

import (
	"errors"
	"fmt"
	"io"

	amp4 "github.com/abema/go-mp4"

	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
)

// Init is a fMP4 initialization block.
type Init struct {
	Tracks   []*InitTrack
	UserData []amp4.IBox
}

// Unmarshal decodes a fMP4 initialization block.
func (i *Init) Unmarshal(r io.ReadSeeker) error {
	type readState int

	const (
		waitingFtyp readState = iota
		waitingMoov
		waitingMvhd
		waitingTrak
		readingUserData
		waitingTkhd
		waitingMdhd
		waitingStsd
		readingCodec
	)

	var state readState
	var curTrack *InitTrack
	var codecBoxesReader *imp4.CodecBoxesReader

	_, err := amp4.ReadBoxStructure(r, func(h *amp4.ReadHandle) (any, error) {
		switch state {
		case readingUserData:
			if len(h.Path) < 3 {
				state = waitingTrak
			} else {
				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}

				i.UserData = append(i.UserData, box)
				return nil, nil
			}

		case readingCodec:
			ret, err := codecBoxesReader.Read(h)
			if err != nil {
				if errors.Is(err, imp4.ErrReadEnded) {
					if codecBoxesReader.Codec != nil {
						curTrack.Codec = codecBoxesReader.Codec
					} else {
						i.Tracks = i.Tracks[:len(i.Tracks)-1]
					}
					state = waitingTrak
				} else {
					return nil, err
				}
			} else {
				return ret, nil
			}
		}

		switch h.BoxInfo.Type.String() {
		case "ftyp":
			if state != waitingFtyp {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingMoov

		case "moov":
			if state != waitingMoov {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingMvhd
			return h.Expand()

		case "udta":
			if state != waitingTrak {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = readingUserData
			return h.Expand()

		case "mvhd":
			if state != waitingMvhd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingTrak
			return h.Expand()

		case "trak":
			if state != waitingTrak {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			curTrack = &InitTrack{}
			i.Tracks = append(i.Tracks, curTrack)
			state = waitingTkhd
			return h.Expand()

		case "tkhd":
			if state != waitingTkhd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			tkhd := box.(*amp4.Tkhd)

			curTrack.ID = int(tkhd.TrackID)
			state = waitingMdhd

		case "mdia":
			if state != waitingMdhd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			return h.Expand()

		case "mdhd":
			if state != waitingMdhd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			mdhd := box.(*amp4.Mdhd)

			if mdhd.Timescale == 0 {
				return nil, fmt.Errorf("invalid timescale")
			}

			curTrack.TimeScale = mdhd.Timescale
			state = waitingStsd

		case "minf", "stbl":
			if state != waitingStsd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			return h.Expand()

		case "stsd":
			if state != waitingStsd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			codecBoxesReader = &imp4.CodecBoxesReader{}
			state = readingCodec
			return h.Expand()
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	if state != waitingTrak && state != readingUserData {
		return fmt.Errorf("parse error")
	}

	if len(i.Tracks) == 0 {
		return fmt.Errorf("no tracks found")
	}

	return nil
}

// Marshal encodes a fMP4 initialization file.
func (i Init) Marshal(w io.WriteSeeker) error {
	/*
		|ftyp|
		|moov|
		|    |mvhd|
		|    |trak|
		|    |trak|
		|    |....|
		|    |mvex|
		|    |    |trex|
		|    |    |trex|
		|    |    |....|
		|    |udta|
		|    |    |....|
	*/

	mw := &imp4.Writer{W: w}
	mw.Initialize()

	_, err := mw.WriteBox(&amp4.Ftyp{ // <ftyp/>
		MajorBrand:   [4]byte{'m', 'p', '4', '2'},
		MinorVersion: 1,
		CompatibleBrands: []amp4.CompatibleBrandElem{
			{CompatibleBrand: [4]byte{'m', 'p', '4', '1'}},
			{CompatibleBrand: [4]byte{'m', 'p', '4', '2'}},
			{CompatibleBrand: [4]byte{'i', 's', 'o', 'm'}},
			{CompatibleBrand: [4]byte{'h', 'l', 's', 'f'}},
		},
	})
	if err != nil {
		return err
	}

	_, err = mw.WriteBoxStart(&amp4.Moov{}) // <moov>
	if err != nil {
		return err
	}

	_, err = mw.WriteBox(&amp4.Mvhd{ // <mvhd/>
		Timescale:   1000,
		Rate:        65536,
		Volume:      256,
		Matrix:      [9]int32{0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000},
		NextTrackID: 4294967295,
	})
	if err != nil {
		return err
	}

	for _, track := range i.Tracks {
		err = track.marshal(mw)
		if err != nil {
			return err
		}
	}

	_, err = mw.WriteBoxStart(&amp4.Mvex{}) // <mvex>
	if err != nil {
		return err
	}

	for _, track := range i.Tracks {
		_, err = mw.WriteBox(&amp4.Trex{ // <trex/>
			TrackID:                       uint32(track.ID),
			DefaultSampleDescriptionIndex: 1,
		})
		if err != nil {
			return err
		}
	}

	err = mw.WriteBoxEnd() // </mvex>
	if err != nil {
		return err
	}

	if len(i.UserData) != 0 {
		_, err = mw.WriteBoxStart(&amp4.Udta{}) // <udta>
		if err != nil {
			return err
		}

		for _, box := range i.UserData {
			_, err = mw.WriteBox(box)
			if err != nil {
				return err
			}
		}

		err = mw.WriteBoxEnd() // </udta>
		if err != nil {
			return err
		}
	}

	err = mw.WriteBoxEnd() // </moov>
	if err != nil {
		return err
	}

	return nil
}
