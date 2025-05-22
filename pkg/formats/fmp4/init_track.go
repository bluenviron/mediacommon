package fmp4

import (
	amp4 "github.com/abema/go-mp4"

	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

const (
	videoBitrate = 1000000
	audioBitrate = 128825
)

// InitTrack is a track of Init.
type InitTrack struct {
	// ID, starts from 1.
	ID int

	// time scale.
	TimeScale uint32

	// average bitrate.
	// it defaults to 1MB for video tracks, 128k for audio tracks.
	AvgBitrate uint32

	// maximum bitrate.
	// it defaults to 1MB for video tracks, 128k for audio tracks.
	MaxBitrate uint32

	// codec.
	Codec mp4.Codec
}

func (it *InitTrack) marshal(w *imp4.Writer) error {
	/*
		|trak|
		|    |tkhd|
		|    |mdia|
		|    |    |mdhd|
		|    |    |hdlr|
		|    |    |minf|
		|    |    |    |vmhd| (video)
		|    |    |    |smhd| (audio)
		|    |    |    |dinf|
		|    |    |    |    |dref|
		|    |    |    |    |    |url|
		|    |    |    |stbl|
		|    |    |    |    |stsd|
		|    |    |    |    |    |XXXX|
		|    |    |    |    |    |    |YYYY|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |stts|
		|    |    |    |    |stsc|
		|    |    |    |    |stsz|
		|    |    |    |    |stco|
	*/

	_, err := w.WriteBoxStart(&amp4.Trak{}) // <trak>
	if err != nil {
		return err
	}

	info, err := imp4.ExtractCodecInfo(it.Codec)
	if err != nil {
		return err
	}

	if it.Codec.IsVideo() {
		_, err = w.WriteBox(&amp4.Tkhd{ // <tkhd/>
			FullBox: amp4.FullBox{
				Flags: [3]byte{0, 0, 3},
			},
			TrackID: uint32(it.ID),
			Width:   uint32(info.Width * 65536),
			Height:  uint32(info.Height * 65536),
			Matrix:  [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		})
		if err != nil {
			return err
		}
	} else {
		_, err = w.WriteBox(&amp4.Tkhd{ // <tkhd/>
			FullBox: amp4.FullBox{
				Flags: [3]byte{0, 0, 3},
			},
			TrackID:        uint32(it.ID),
			AlternateGroup: 1,
			Volume:         256,
			Matrix:         [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		})
		if err != nil {
			return err
		}
	}

	_, err = w.WriteBoxStart(&amp4.Mdia{}) // <mdia>
	if err != nil {
		return err
	}

	_, err = w.WriteBox(&amp4.Mdhd{ // <mdhd/>
		Timescale: it.TimeScale,
		Language:  [3]byte{'u', 'n', 'd'},
	})
	if err != nil {
		return err
	}

	if it.Codec.IsVideo() {
		_, err = w.WriteBox(&amp4.Hdlr{ // <hdlr/>
			HandlerType: [4]byte{'v', 'i', 'd', 'e'},
			Name:        "VideoHandler",
		})
		if err != nil {
			return err
		}
	} else {
		_, err = w.WriteBox(&amp4.Hdlr{ // <hdlr/>
			HandlerType: [4]byte{'s', 'o', 'u', 'n'},
			Name:        "SoundHandler",
		})
		if err != nil {
			return err
		}
	}

	_, err = w.WriteBoxStart(&amp4.Minf{}) // <minf>
	if err != nil {
		return err
	}

	if it.Codec.IsVideo() {
		_, err = w.WriteBox(&amp4.Vmhd{ // <vmhd/>
			FullBox: amp4.FullBox{
				Flags: [3]byte{0, 0, 1},
			},
		})
		if err != nil {
			return err
		}
	} else {
		_, err = w.WriteBox(&amp4.Smhd{}) // <smhd/>
		if err != nil {
			return err
		}
	}

	_, err = w.WriteBoxStart(&amp4.Dinf{}) // <dinf>
	if err != nil {
		return err
	}

	_, err = w.WriteBoxStart(&amp4.Dref{ // <dref>
		EntryCount: 1,
	})
	if err != nil {
		return err
	}

	_, err = w.WriteBox(&amp4.Url{ // <url/>
		FullBox: amp4.FullBox{
			Flags: [3]byte{0, 0, 1},
		},
	})
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd() // </dref>
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd() // </dinf>
	if err != nil {
		return err
	}

	_, err = w.WriteBoxStart(&amp4.Stbl{}) // <stbl>
	if err != nil {
		return err
	}

	_, err = w.WriteBoxStart(&amp4.Stsd{ // <stsd>
		EntryCount: 1,
	})
	if err != nil {
		return err
	}

	avgBitrate := it.AvgBitrate
	if avgBitrate == 0 {
		if it.Codec.IsVideo() {
			avgBitrate = videoBitrate
		} else {
			avgBitrate = audioBitrate
		}
	}

	maxBitrate := it.MaxBitrate
	if maxBitrate == 0 {
		if it.Codec.IsVideo() {
			maxBitrate = videoBitrate
		} else {
			maxBitrate = audioBitrate
		}
	}

	err = imp4.WriteCodecBoxes(w, it.Codec, it.ID, info, avgBitrate, maxBitrate)
	if err != nil {
		return err
	}

	_, err = w.WriteBox(&amp4.Btrt{ // <btrt/>
		MaxBitrate: maxBitrate,
		AvgBitrate: avgBitrate,
	})
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd() // </*>
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd() // </stsd>
	if err != nil {
		return err
	}

	_, err = w.WriteBox(&amp4.Stts{ // <stts/>
	})
	if err != nil {
		return err
	}

	_, err = w.WriteBox(&amp4.Stsc{ // <stsc/>
	})
	if err != nil {
		return err
	}

	_, err = w.WriteBox(&amp4.Stsz{ // <stsz/>
	})
	if err != nil {
		return err
	}

	_, err = w.WriteBox(&amp4.Stco{ // <stco/>
	})
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd() // </stbl>
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd() // </minf>
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd() // </mdia>
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd() // </trak>
	if err != nil {
		return err
	}

	return nil
}
