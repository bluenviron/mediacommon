package fmp4

import (
	"bytes"
	"fmt"
	"io"

	"github.com/abema/go-mp4"

	"github.com/bluenviron/mediacommon/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

func av1FindSequenceHeader(bs []byte) ([]byte, error) {
	tu, err := av1.BitstreamUnmarshal(bs, true)
	if err != nil {
		return nil, err
	}

	for _, obu := range tu {
		var h av1.OBUHeader
		err := h.Unmarshal(obu)
		if err != nil {
			return nil, err
		}

		if h.Type == av1.OBUTypeSequenceHeader {
			return obu, nil
		}
	}

	return nil, fmt.Errorf("sequence header not found")
}

func h265FindParams(params []mp4.HEVCNaluArray) ([]byte, []byte, []byte, error) {
	var vps []byte
	var sps []byte
	var pps []byte

	for _, arr := range params {
		switch h265.NALUType(arr.NaluType) {
		case h265.NALUType_VPS_NUT, h265.NALUType_SPS_NUT, h265.NALUType_PPS_NUT:
			if arr.NumNalus != 1 {
				return nil, nil, nil, fmt.Errorf("multiple VPS/SPS/PPS are not supported")
			}
		}

		switch h265.NALUType(arr.NaluType) {
		case h265.NALUType_VPS_NUT:
			vps = arr.Nalus[0].NALUnit

		case h265.NALUType_SPS_NUT:
			sps = arr.Nalus[0].NALUnit

		case h265.NALUType_PPS_NUT:
			pps = arr.Nalus[0].NALUnit
		}
	}

	if vps == nil {
		return nil, nil, nil, fmt.Errorf("VPS not provided")
	}

	if sps == nil {
		return nil, nil, nil, fmt.Errorf("SPS not provided")
	}

	if pps == nil {
		return nil, nil, nil, fmt.Errorf("PPS not provided")
	}

	return vps, sps, pps, nil
}

func h264FindParams(avcc *mp4.AVCDecoderConfiguration) ([]byte, []byte, error) {
	if len(avcc.SequenceParameterSets) > 1 {
		return nil, nil, fmt.Errorf("multiple SPS are not supported")
	}

	var sps []byte
	if len(avcc.SequenceParameterSets) == 1 {
		sps = avcc.SequenceParameterSets[0].NALUnit
	}

	if len(avcc.PictureParameterSets) > 1 {
		return nil, nil, fmt.Errorf("multiple PPS are not supported")
	}

	var pps []byte
	if len(avcc.PictureParameterSets) == 1 {
		pps = avcc.PictureParameterSets[0].NALUnit
	}

	return sps, pps, nil
}

func mpeg4AudioFindConfig(descriptors []mp4.Descriptor) (*mpeg4audio.Config, error) {
	encodedConf := func() []byte {
		for _, desc := range descriptors {
			if desc.Tag == mp4.DecSpecificInfoTag {
				return desc.Data
			}
		}
		return nil
	}()
	if encodedConf == nil {
		return nil, fmt.Errorf("unable to find MPEG-4 Audio configuration")
	}

	var c mpeg4audio.Config
	err := c.Unmarshal(encodedConf)
	if err != nil {
		return nil, fmt.Errorf("invalid MPEG-4 Audio configuration: %s", err)
	}

	return &c, nil
}

// Init is a fMP4 initialization block.
type Init struct {
	Tracks []*InitTrack
}

// Unmarshal decodes a fMP4 initialization block.
func (i *Init) Unmarshal(byts []byte) error {
	type readState int

	const (
		waitingTrak readState = iota
		waitingTkhd
		waitingMdhd
		waitingCodec
		waitingAv1C
		waitingVpcC
		waitingHvcC
		waitingAvcC
		waitingEsds
		waitingDOps
	)

	state := waitingTrak
	var curTrack *InitTrack

	_, err := mp4.ReadBoxStructure(bytes.NewReader(byts), func(h *mp4.ReadHandle) (interface{}, error) {
		switch h.BoxInfo.Type.String() {
		case "trak":
			if state != waitingTrak {
				return nil, fmt.Errorf("unexpected box 'trak'")
			}

			curTrack = &InitTrack{}
			i.Tracks = append(i.Tracks, curTrack)
			state = waitingTkhd

		case "tkhd":
			if state != waitingTkhd {
				return nil, fmt.Errorf("unexpected box 'tkhd'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			tkhd := box.(*mp4.Tkhd)

			curTrack.ID = int(tkhd.TrackID)
			state = waitingMdhd

		case "mdhd":
			if state != waitingMdhd {
				return nil, fmt.Errorf("unexpected box 'mdhd'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			mdhd := box.(*mp4.Mdhd)

			curTrack.TimeScale = mdhd.Timescale
			state = waitingCodec

		case "avc1":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box 'avc1'")
			}
			state = waitingAvcC

		case "avcC":
			if state != waitingAvcC {
				return nil, fmt.Errorf("unexpected box 'avcC'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			avcc := box.(*mp4.AVCDecoderConfiguration)

			sps, pps, err := h264FindParams(avcc)
			if err != nil {
				return nil, err
			}

			curTrack.Codec = &CodecH264{
				SPS: sps,
				PPS: pps,
			}
			state = waitingTrak

		case "vp09":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box 'vp09'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			vp09 := box.(*mp4.VisualSampleEntry)

			curTrack.Codec = &CodecVP9{
				Width:  int(vp09.Width),
				Height: int(vp09.Height),
			}
			state = waitingVpcC

		case "vpcC":
			if state != waitingVpcC {
				return nil, fmt.Errorf("unexpected box 'vpcC'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			vpcc := box.(*VpcC)

			curTrack.Codec.(*CodecVP9).Profile = vpcc.Profile
			curTrack.Codec.(*CodecVP9).BitDepth = vpcc.BitDepth
			curTrack.Codec.(*CodecVP9).ChromaSubsampling = vpcc.ChromaSubsampling
			curTrack.Codec.(*CodecVP9).ColorRange = vpcc.VideoFullRangeFlag != 0
			state = waitingTrak

		case "vp08": // VP8, not supported yet
			return nil, nil

		case "hev1", "hvc1":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box 'hev1'")
			}
			state = waitingHvcC

		case "hvcC":
			if state != waitingHvcC {
				return nil, fmt.Errorf("unexpected box 'hvcC'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			hvcc := box.(*mp4.HvcC)

			vps, sps, pps, err := h265FindParams(hvcc.NaluArrays)
			if err != nil {
				return nil, err
			}

			curTrack.Codec = &CodecH265{
				VPS: vps,
				SPS: sps,
				PPS: pps,
			}
			state = waitingTrak

		case "av01":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box 'av01'")
			}
			state = waitingAv1C

		case "av1C":
			if state != waitingAv1C {
				return nil, fmt.Errorf("unexpected box 'av1C'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			av1c := box.(*mp4.Av1C)

			sequenceHeader, err := av1FindSequenceHeader(av1c.ConfigOBUs)
			if err != nil {
				return nil, err
			}

			curTrack.Codec = &CodecAV1{
				SequenceHeader: sequenceHeader,
			}
			state = waitingTrak

		case "Opus":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box 'Opus'")
			}
			state = waitingDOps

		case "dOps":
			if state != waitingDOps {
				return nil, fmt.Errorf("unexpected box 'dOps'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			dops := box.(*mp4.DOps)

			curTrack.Codec = &CodecOpus{
				ChannelCount: int(dops.OutputChannelCount),
			}
			state = waitingTrak

		case "mp4a":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box 'mp4a'")
			}
			state = waitingEsds

		case "esds":
			if state != waitingEsds {
				return nil, fmt.Errorf("unexpected box 'esds'")
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			esds := box.(*mp4.Esds)

			config, err := mpeg4AudioFindConfig(esds.Descriptors)
			if err != nil {
				return nil, err
			}

			curTrack.Codec = &CodecMPEG4Audio{
				Config: *config,
			}
			state = waitingTrak

		case "ac-3": // ac-3, not supported yet
			i.Tracks = i.Tracks[:len(i.Tracks)-1]
			state = waitingTrak
			return nil, nil

		case "ec-3": // ec-3, not supported yet
			i.Tracks = i.Tracks[:len(i.Tracks)-1]
			state = waitingTrak
			return nil, nil

		case "c608", "c708": // closed captions, not supported yet
			i.Tracks = i.Tracks[:len(i.Tracks)-1]
			state = waitingTrak
			return nil, nil

		case "chrm", "nmhd":
			return nil, nil
		}

		return h.Expand()
	})
	if err != nil {
		return err
	}

	if state != waitingTrak {
		return fmt.Errorf("parse error")
	}

	if len(i.Tracks) == 0 {
		return fmt.Errorf("no tracks found")
	}

	return nil
}

// Marshal encodes a fMP4 initialization file.
func (i *Init) Marshal(w io.WriteSeeker) error {
	/*
		- ftyp
		- moov
		  - mvhd
		  - trak
		  - trak
		  - ...
		- mvex
		  - trex
		  - trex
		  - ...
	*/

	mw := newMP4Writer(w)

	_, err := mw.writeBox(&mp4.Ftyp{ // <ftyp/>
		MajorBrand:   [4]byte{'m', 'p', '4', '2'},
		MinorVersion: 1,
		CompatibleBrands: []mp4.CompatibleBrandElem{
			{CompatibleBrand: [4]byte{'m', 'p', '4', '1'}},
			{CompatibleBrand: [4]byte{'m', 'p', '4', '2'}},
			{CompatibleBrand: [4]byte{'i', 's', 'o', 'm'}},
			{CompatibleBrand: [4]byte{'h', 'l', 's', 'f'}},
		},
	})
	if err != nil {
		return err
	}

	_, err = mw.writeBoxStart(&mp4.Moov{}) // <moov>
	if err != nil {
		return err
	}

	_, err = mw.writeBox(&mp4.Mvhd{ // <mvhd/>
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
		err := track.marshal(mw)
		if err != nil {
			return err
		}
	}

	_, err = mw.writeBoxStart(&mp4.Mvex{}) // <mvex>
	if err != nil {
		return err
	}

	for _, track := range i.Tracks {
		_, err = mw.writeBox(&mp4.Trex{ // <trex/>
			TrackID:                       uint32(track.ID),
			DefaultSampleDescriptionIndex: 1,
		})
		if err != nil {
			return err
		}
	}

	err = mw.writeBoxEnd() // </mvex>
	if err != nil {
		return err
	}

	err = mw.writeBoxEnd() // </moov>
	if err != nil {
		return err
	}

	return nil
}
