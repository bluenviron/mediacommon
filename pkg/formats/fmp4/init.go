package fmp4

import (
	"fmt"
	"io"

	amp4 "github.com/abema/go-mp4"

	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

func av1FindSequenceHeader(buf []byte) ([]byte, error) {
	var tu av1.Bitstream
	err := tu.Unmarshal(buf)
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
			var parsed av1.SequenceHeader
			err = parsed.Unmarshal(obu)
			if err != nil {
				return nil, err
			}

			return obu, nil
		}
	}

	return nil, fmt.Errorf("AV1 sequence header not found")
}

func h265FindParams(params []amp4.HEVCNaluArray) ([]byte, []byte, []byte, error) {
	var vps []byte
	var sps []byte
	var pps []byte

	for _, arr := range params {
		switch h265.NALUType(arr.NaluType) {
		case h265.NALUType_VPS_NUT, h265.NALUType_SPS_NUT, h265.NALUType_PPS_NUT:
			if arr.NumNalus != 1 {
				return nil, nil, nil, fmt.Errorf("multiple H265 VPS/SPS/PPS are not supported")
			}

			switch h265.NALUType(arr.NaluType) {
			case h265.NALUType_VPS_NUT:
				vps = arr.Nalus[0].NALUnit

			case h265.NALUType_SPS_NUT:
				sps = arr.Nalus[0].NALUnit

				var spsp h265.SPS
				err := spsp.Unmarshal(sps)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("unable to parse H265 SPS: %w", err)
				}

			case h265.NALUType_PPS_NUT:
				pps = arr.Nalus[0].NALUnit
			}
		}
	}

	if len(vps) == 0 {
		return nil, nil, nil, fmt.Errorf("H265 VPS not provided")
	}

	if len(sps) == 0 {
		return nil, nil, nil, fmt.Errorf("H265 SPS not provided")
	}

	if len(pps) == 0 {
		return nil, nil, nil, fmt.Errorf("H265 PPS not provided")
	}

	return vps, sps, pps, nil
}

func h264FindParams(avcc *amp4.AVCDecoderConfiguration) ([]byte, []byte, error) {
	if len(avcc.SequenceParameterSets) == 0 {
		return nil, nil, fmt.Errorf("H264 SPS not provided")
	}
	if len(avcc.SequenceParameterSets) > 1 {
		return nil, nil, fmt.Errorf("multiple H264 SPS are not supported")
	}
	if len(avcc.PictureParameterSets) == 0 {
		return nil, nil, fmt.Errorf("H264 PPS not provided")
	}
	if len(avcc.PictureParameterSets) > 1 {
		return nil, nil, fmt.Errorf("multiple H264 PPS are not supported")
	}

	sps := avcc.SequenceParameterSets[0].NALUnit
	var spsp h264.SPS
	err := spsp.Unmarshal(sps)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse H264 SPS: %w", err)
	}

	pps := avcc.PictureParameterSets[0].NALUnit
	if len(pps) == 0 {
		return nil, nil, fmt.Errorf("invalid H264 PPS")
	}

	return sps, pps, nil
}

func esdsFindDecoderConf(descriptors []amp4.Descriptor) *amp4.DecoderConfigDescriptor {
	for _, desc := range descriptors {
		if desc.Tag == amp4.DecoderConfigDescrTag {
			return desc.DecoderConfigDescriptor
		}
	}
	return nil
}

func esdsFindDecoderSpecificInfo(descriptors []amp4.Descriptor) []byte {
	for _, desc := range descriptors {
		if desc.Tag == amp4.DecSpecificInfoTag {
			return desc.Data
		}
	}
	return nil
}

// Init is a fMP4 initialization block.
type Init struct {
	Tracks []*InitTrack
}

// Unmarshal decodes a fMP4 initialization block.
func (i *Init) Unmarshal(r io.ReadSeeker) error {
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
		waitingVideoEsds
		waitingAudioEsds
		waitingDOps
		waitingDac3
		waitingPcmC
	)

	state := waitingTrak
	var curTrack *InitTrack
	var width int
	var height int
	var sampleRate int
	var channelCount int

	_, err := amp4.ReadBoxStructure(r, func(h *amp4.ReadHandle) (interface{}, error) {
		if !h.BoxInfo.IsSupportedType() {
			if state != waitingTrak {
				i.Tracks = i.Tracks[:len(i.Tracks)-1]
				state = waitingTrak
			}
		} else {
			switch h.BoxInfo.Type.String() {
			case "moov":
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
				state = waitingCodec

			case "minf", "stbl", "stsd":
				return h.Expand()

			case "avc1":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}
				state = waitingAvcC
				return h.Expand()

			case "avcC":
				if state != waitingAvcC {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				avcc := box.(*amp4.AVCDecoderConfiguration)

				sps, pps, err := h264FindParams(avcc)
				if err != nil {
					return nil, err
				}

				curTrack.Codec = &mp4.CodecH264{
					SPS: sps,
					PPS: pps,
				}
				state = waitingTrak

			case "vp09":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				vp09 := box.(*amp4.VisualSampleEntry)

				if vp09.Width == 0 || vp09.Height == 0 {
					return nil, fmt.Errorf("VP9 parameters not provided")
				}

				width = int(vp09.Width)
				height = int(vp09.Height)
				state = waitingVpcC
				return h.Expand()

			case "vpcC":
				if state != waitingVpcC {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				vpcc := box.(*amp4.VpcC)

				curTrack.Codec = &mp4.CodecVP9{
					Width:             width,
					Height:            height,
					Profile:           vpcc.Profile,
					BitDepth:          vpcc.BitDepth,
					ChromaSubsampling: vpcc.ChromaSubsampling,
					ColorRange:        vpcc.VideoFullRangeFlag != 0,
				}
				state = waitingTrak

			case "vp08": // VP8, not supported yet
				return nil, nil

			case "hev1", "hvc1":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}
				state = waitingHvcC
				return h.Expand()

			case "hvcC":
				if state != waitingHvcC {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				hvcc := box.(*amp4.HvcC)

				vps, sps, pps, err := h265FindParams(hvcc.NaluArrays)
				if err != nil {
					return nil, err
				}

				curTrack.Codec = &mp4.CodecH265{
					VPS: vps,
					SPS: sps,
					PPS: pps,
				}
				state = waitingTrak

			case "av01":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}
				state = waitingAv1C
				return h.Expand()

			case "av1C":
				if state != waitingAv1C {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				av1c := box.(*amp4.Av1C)

				sequenceHeader, err := av1FindSequenceHeader(av1c.ConfigOBUs)
				if err != nil {
					return nil, err
				}

				curTrack.Codec = &mp4.CodecAV1{
					SequenceHeader: sequenceHeader,
				}
				state = waitingTrak

			case "Opus":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}
				state = waitingDOps
				return h.Expand()

			case "dOps":
				if state != waitingDOps {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				dops := box.(*amp4.DOps)

				curTrack.Codec = &mp4.CodecOpus{
					ChannelCount: int(dops.OutputChannelCount),
				}
				state = waitingTrak

			case "mp4v":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				mp4v := box.(*amp4.VisualSampleEntry)

				width = int(mp4v.Width)
				height = int(mp4v.Height)
				state = waitingVideoEsds
				return h.Expand()

			case "mp4a":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				mp4a := box.(*amp4.AudioSampleEntry)

				sampleRate = int(mp4a.SampleRate / 65536)
				channelCount = int(mp4a.ChannelCount)
				state = waitingAudioEsds
				return h.Expand()

			case "esds":
				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				esds := box.(*amp4.Esds)

				conf := esdsFindDecoderConf(esds.Descriptors)
				if conf == nil {
					return nil, fmt.Errorf("unable to find decoder config")
				}

				switch state {
				case waitingVideoEsds:
					switch conf.ObjectTypeIndication {
					case imp4.ObjectTypeIndicationVisualISO14496part2:
						spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
						if len(spec) == 0 {
							return nil, fmt.Errorf("unable to find decoder specific info")
						}

						curTrack.Codec = &mp4.CodecMPEG4Video{
							Config: spec,
						}

					case imp4.ObjectTypeIndicationVisualISO1318part2Main:
						spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
						if len(spec) == 0 {
							return nil, fmt.Errorf("unable to find decoder specific info")
						}

						curTrack.Codec = &mp4.CodecMPEG1Video{
							Config: spec,
						}

					case imp4.ObjectTypeIndicationVisualISO10918part1:
						if width == 0 || height == 0 {
							return nil, fmt.Errorf("M-JPEG parameters not provided")
						}

						curTrack.Codec = &mp4.CodecMJPEG{
							Width:  width,
							Height: height,
						}

					default:
						return nil, fmt.Errorf("unsupported object type indication: 0x%.2x", conf.ObjectTypeIndication)
					}

					state = waitingTrak

				case waitingAudioEsds:
					switch conf.ObjectTypeIndication {
					case imp4.ObjectTypeIndicationAudioISO14496part3:
						spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
						if len(spec) == 0 {
							return nil, fmt.Errorf("unable to find decoder specific info")
						}

						var c mpeg4audio.Config
						err := c.Unmarshal(spec)
						if err != nil {
							return nil, fmt.Errorf("invalid MPEG-4 Audio configuration: %w", err)
						}

						curTrack.Codec = &mp4.CodecMPEG4Audio{
							Config: c,
						}

					case imp4.ObjectTypeIndicationAudioISO11172part3:
						curTrack.Codec = &mp4.CodecMPEG1Audio{
							SampleRate:   sampleRate,
							ChannelCount: channelCount,
						}

					default:
						return nil, fmt.Errorf("unsupported object type indication: 0x%.2x", conf.ObjectTypeIndication)
					}

				default:
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				state = waitingTrak

			case "ac-3":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				ac3 := box.(*amp4.AudioSampleEntry)

				sampleRate = int(ac3.SampleRate / 65536)
				channelCount = int(ac3.ChannelCount)
				state = waitingDac3
				return h.Expand()

			case "dac3":
				if state != waitingDac3 {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				dac3 := box.(*amp4.Dac3)

				curTrack.Codec = &mp4.CodecAC3{
					SampleRate:   sampleRate,
					ChannelCount: channelCount,
					Fscod:        dac3.Fscod,
					Bsid:         dac3.Bsid,
					Bsmod:        dac3.Bsmod,
					Acmod:        dac3.Acmod,
					LfeOn:        dac3.LfeOn != 0,
					BitRateCode:  dac3.BitRateCode,
				}
				state = waitingTrak

			case "ipcm":
				if state != waitingCodec {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				ac3 := box.(*amp4.AudioSampleEntry)

				sampleRate = int(ac3.SampleRate / 65536)
				channelCount = int(ac3.ChannelCount)
				state = waitingPcmC
				return h.Expand()

			case "pcmC":
				if state != waitingPcmC {
					return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
				}

				box, _, err := h.ReadPayload()
				if err != nil {
					return nil, err
				}
				pcmc := box.(*amp4.PcmC)

				curTrack.Codec = &mp4.CodecLPCM{
					LittleEndian: (pcmc.FormatFlags & 0x01) != 0,
					BitDepth:     int(pcmc.PCMSampleSize),
					SampleRate:   sampleRate,
					ChannelCount: channelCount,
				}
				state = waitingTrak
			}
		}

		return nil, nil
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

	err = mw.WriteBoxEnd() // </moov>
	if err != nil {
		return err
	}

	return nil
}
