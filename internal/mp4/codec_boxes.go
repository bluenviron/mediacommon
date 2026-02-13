package mp4

import (
	"fmt"

	amp4 "github.com/abema/go-mp4"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4/codecs"
)

// ErrReadEnded is returned when reading codec boxes has ended.
var ErrReadEnded = fmt.Errorf("OK")

func boolToUint8(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}

func av1FindSequenceHeader(buf []byte) ([]byte, error) {
	var tu av1.Bitstream
	err := tu.Unmarshal(buf)
	if err != nil {
		return nil, err
	}

	for _, obu := range tu {
		typ := av1.OBUType((obu[0] >> 3) & 0b1111)

		if typ == av1.OBUTypeSequenceHeader {
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

type readState int

const (
	initial readState = iota
	waitingAv1C
	waitingVpcC
	waitingHvcC
	waitingAvcC
	waitingVideoEsds
	waitingAudioEsds
	waitingDOps
	waitingDac3
	waitingDec3
	waitingPcmC
	waitingAdditional
)

// CodecBoxesReader reads codec-related boxes.
type CodecBoxesReader struct {
	Codec codecs.Codec

	state        readState
	width        int
	height       int
	sampleRate   int
	channelCount int
}

// ReadCodecBoxes reads codec-related boxes.
func (r *CodecBoxesReader) Read(h *amp4.ReadHandle) (any, error) {
	if len(h.Path) < 7 {
		if r.state != waitingAdditional {
			return nil, fmt.Errorf("codec information not found")
		}

		return nil, ErrReadEnded
	}

	switch h.BoxInfo.Type.String() {
	// codecs not supported yet
	case "c608":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		r.state = waitingAdditional

	case "avc1":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		r.state = waitingAvcC
		return h.Expand()

	case "avcC":
		if r.state != waitingAvcC {
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

		r.Codec = &codecs.H264{
			SPS: sps,
			PPS: pps,
		}
		r.state = waitingAdditional

	case "vp09":
		if r.state != initial {
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

		r.width = int(vp09.Width)
		r.height = int(vp09.Height)
		r.state = waitingVpcC
		return h.Expand()

	case "vpcC":
		if r.state != waitingVpcC {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		vpcc := box.(*amp4.VpcC)

		r.Codec = &codecs.VP9{
			Width:             r.width,
			Height:            r.height,
			Profile:           vpcc.Profile,
			BitDepth:          vpcc.BitDepth,
			ChromaSubsampling: vpcc.ChromaSubsampling,
			ColorRange:        vpcc.VideoFullRangeFlag != 0,
		}
		r.state = waitingAdditional

	case "hvc1", "hev1":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		r.state = waitingHvcC
		return h.Expand()

	case "hvcC":
		if r.state != waitingHvcC {
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

		r.Codec = &codecs.H265{
			VPS: vps,
			SPS: sps,
			PPS: pps,
		}
		r.state = waitingAdditional

	case "mp4a":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		mp4a := box.(*amp4.AudioSampleEntry)

		r.sampleRate = int(mp4a.SampleRate / 65536)
		r.channelCount = int(mp4a.ChannelCount)
		r.state = waitingAudioEsds
		return h.Expand()

	case "av01":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		r.state = waitingAv1C
		return h.Expand()

	case "av1C":
		if r.state != waitingAv1C {
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

		r.Codec = &codecs.AV1{
			SequenceHeader: sequenceHeader,
		}
		r.state = waitingAdditional

	case "Opus":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		r.state = waitingDOps
		return h.Expand()

	case "dOps":
		if r.state != waitingDOps {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		dops := box.(*amp4.DOps)

		r.Codec = &codecs.Opus{
			ChannelCount: int(dops.OutputChannelCount),
		}
		r.state = waitingAdditional

	case "mp4v":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		mp4v := box.(*amp4.VisualSampleEntry)

		r.width = int(mp4v.Width)
		r.height = int(mp4v.Height)
		r.state = waitingVideoEsds
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

		switch r.state {
		case waitingVideoEsds:
			switch conf.ObjectTypeIndication {
			case ObjectTypeIndicationVisualISO14496part2:
				spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
				if len(spec) == 0 {
					return nil, fmt.Errorf("unable to find decoder specific info")
				}

				r.Codec = &codecs.MPEG4Video{
					Config: spec,
				}
				r.state = waitingAdditional

			case ObjectTypeIndicationVisualISO1318part2Main:
				spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
				if len(spec) == 0 {
					return nil, fmt.Errorf("unable to find decoder specific info")
				}

				r.Codec = &codecs.MPEG1Video{
					Config: spec,
				}
				r.state = waitingAdditional

			case ObjectTypeIndicationVisualISO10918part1:
				if r.width == 0 || r.height == 0 {
					return nil, fmt.Errorf("M-JPEG parameters not provided")
				}

				r.Codec = &codecs.MJPEG{
					Width:  r.width,
					Height: r.height,
				}
				r.state = waitingAdditional

			default:
				return nil, fmt.Errorf("unsupported object type indication: 0x%.2x", conf.ObjectTypeIndication)
			}

		case waitingAudioEsds:
			switch conf.ObjectTypeIndication {
			case ObjectTypeIndicationAudioISO14496part3:
				spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
				if len(spec) == 0 {
					return nil, fmt.Errorf("unable to find decoder specific info")
				}

				var c mpeg4audio.AudioSpecificConfig
				err = c.Unmarshal(spec)
				if err != nil {
					return nil, fmt.Errorf("invalid MPEG-4 Audio configuration: %w", err)
				}

				r.Codec = &codecs.MPEG4Audio{
					Config: c,
				}
				r.state = waitingAdditional

			case ObjectTypeIndicationAudioISO11172part3:
				r.Codec = &codecs.MPEG1Audio{
					SampleRate:   r.sampleRate,
					ChannelCount: r.channelCount,
				}
				r.state = waitingAdditional

			default:
				return nil, fmt.Errorf("unsupported object type indication: 0x%.2x", conf.ObjectTypeIndication)
			}

		default:
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

	case "ac-3":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		ac3 := box.(*amp4.AudioSampleEntry)

		r.sampleRate = int(ac3.SampleRate / 65536)
		r.channelCount = int(ac3.ChannelCount)
		r.state = waitingDac3
		return h.Expand()

	case "dac3":
		if r.state != waitingDac3 {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		dac3 := box.(*amp4.Dac3)

		r.Codec = &codecs.AC3{
			SampleRate:   r.sampleRate,
			ChannelCount: r.channelCount,
			Fscod:        dac3.Fscod,
			Bsid:         dac3.Bsid,
			Bsmod:        dac3.Bsmod,
			Acmod:        dac3.Acmod,
			LfeOn:        dac3.LfeOn != 0,
			BitRateCode:  dac3.BitRateCode,
		}
		r.state = waitingAdditional

	case "ec-3":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		eac3 := box.(*amp4.AudioSampleEntry)

		r.sampleRate = int(eac3.SampleRate / 65536)
		r.channelCount = int(eac3.ChannelCount)
		r.state = waitingDec3
		return h.Expand()

	case "dec3":
		if r.state != waitingDec3 {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		dec3 := box.(*Dec3)

		r.Codec = &codecs.EAC3{
			SampleRate:   r.sampleRate,
			ChannelCount: r.channelCount,
			DataRate:     dec3.DataRate,
			Asvc:         dec3.Asvc != 0,
			Bsmod:        dec3.Bsmod,
			Acmod:        dec3.Acmod,
			LfeOn:        dec3.LfeOn != 0,
			NumDepSub:    dec3.NumDepSub,
			ChanLoc:      dec3.ChanLoc,
		}
		r.state = waitingAdditional

	case "ipcm":
		if r.state != initial {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		ac3 := box.(*amp4.AudioSampleEntry)

		r.sampleRate = int(ac3.SampleRate / 65536)
		r.channelCount = int(ac3.ChannelCount)
		r.state = waitingPcmC
		return h.Expand()

	case "pcmC":
		if r.state != waitingPcmC {
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		pcmc := box.(*amp4.PcmC)

		r.Codec = &codecs.LPCM{
			LittleEndian: (pcmc.FormatFlags & 0x01) != 0,
			BitDepth:     int(pcmc.PCMSampleSize),
			SampleRate:   r.sampleRate,
			ChannelCount: r.channelCount,
		}
		r.state = waitingAdditional
	}

	return nil, nil
}

// WriteCodecBoxes writes codec-related boxes.
func WriteCodecBoxes(w *Writer, codec codecs.Codec, trackID int, info *CodecInfo, avgBitrate, maxBitrate uint32) error {
	/*
		|av01| (AV1)
		|    |av1C|
		|vp09| (VP9)
		|    |vpcC|
		|hvc1| (H265)
		|    |hvcC|
		|avc1| (H264)
		|    |avcC|
		|mp4v| (MPEG-4/2/1 video, MJPEG)
		|    |esds|
		|Opus| (Opus)
		|    |dOps|
		|mp4a| (MPEG-4/1 audio)
		|    |esds|
		|ac-3| (AC-3)
		|    |dac3|
		|ec-3| (E-AC-3 / Dolby Digital Plus)
		|    |dec3|
		|ipcm| (LPCM)
		|    |pcmC|
	*/

	switch codec := codec.(type) {
	case *codecs.AV1:
		_, err := w.WriteBoxStart(&amp4.VisualSampleEntry{ // <av01>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeAv01(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(info.Width),
			Height:          uint16(info.Height),
			Horizresolution: 4718592,
			Vertresolution:  4718592,
			FrameCount:      1,
			Depth:           24,
			PreDefined3:     -1,
		})
		if err != nil {
			return err
		}

		var enc []byte
		enc, err = av1.Bitstream([][]byte{codec.SequenceHeader}).Marshal()
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.Av1C{ // <av1C/>
			Marker:               1,
			Version:              1,
			SeqProfile:           info.AV1SequenceHeader.SeqProfile,
			SeqLevelIdx0:         info.AV1SequenceHeader.SeqLevelIdx[0],
			SeqTier0:             boolToUint8(info.AV1SequenceHeader.SeqTier[0]),
			HighBitdepth:         boolToUint8(info.AV1SequenceHeader.ColorConfig.HighBitDepth),
			TwelveBit:            boolToUint8(info.AV1SequenceHeader.ColorConfig.TwelveBit),
			Monochrome:           boolToUint8(info.AV1SequenceHeader.ColorConfig.MonoChrome),
			ChromaSubsamplingX:   boolToUint8(info.AV1SequenceHeader.ColorConfig.SubsamplingX),
			ChromaSubsamplingY:   boolToUint8(info.AV1SequenceHeader.ColorConfig.SubsamplingY),
			ChromaSamplePosition: uint8(info.AV1SequenceHeader.ColorConfig.ChromaSamplePosition),
			ConfigOBUs:           enc,
		})
		if err != nil {
			return err
		}

	case *codecs.VP9:
		_, err := w.WriteBoxStart(&amp4.VisualSampleEntry{ // <vp09>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeVp09(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(info.Width),
			Height:          uint16(info.Height),
			Horizresolution: 4718592,
			Vertresolution:  4718592,
			FrameCount:      1,
			Depth:           24,
			PreDefined3:     -1,
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.VpcC{ // <vpcC/>
			FullBox: amp4.FullBox{
				Version: 1,
			},
			Profile:            codec.Profile,
			Level:              10, // level 1
			BitDepth:           codec.BitDepth,
			ChromaSubsampling:  codec.ChromaSubsampling,
			VideoFullRangeFlag: boolToUint8(codec.ColorRange),
		})
		if err != nil {
			return err
		}

	case *codecs.H265:
		_, err := w.WriteBoxStart(&amp4.VisualSampleEntry{ // <hvc1>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeHvc1(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(info.Width),
			Height:          uint16(info.Height),
			Horizresolution: 4718592,
			Vertresolution:  4718592,
			FrameCount:      1,
			Depth:           24,
			PreDefined3:     -1,
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.HvcC{ // <hvcC/>
			ConfigurationVersion:        1,
			GeneralProfileIdc:           info.H265SPS.ProfileTierLevel.GeneralProfileIdc,
			GeneralProfileCompatibility: info.H265SPS.ProfileTierLevel.GeneralProfileCompatibilityFlag,
			GeneralConstraintIndicator: [6]uint8{
				codec.SPS[7], codec.SPS[8], codec.SPS[9],
				codec.SPS[10], codec.SPS[11], codec.SPS[12],
			},
			GeneralLevelIdc: info.H265SPS.ProfileTierLevel.GeneralLevelIdc,
			// MinSpatialSegmentationIdc
			// ParallelismType
			ChromaFormatIdc:      uint8(info.H265SPS.ChromaFormatIdc),
			BitDepthLumaMinus8:   uint8(info.H265SPS.BitDepthLumaMinus8),
			BitDepthChromaMinus8: uint8(info.H265SPS.BitDepthChromaMinus8),
			// AvgFrameRate
			// ConstantFrameRate
			NumTemporalLayers: 1,
			// TemporalIdNested
			LengthSizeMinusOne: 3,
			NumOfNaluArrays:    3,
			NaluArrays: []amp4.HEVCNaluArray{
				{
					NaluType: byte(h265.NALUType_VPS_NUT),
					NumNalus: 1,
					Nalus: []amp4.HEVCNalu{{
						Length:  uint16(len(codec.VPS)),
						NALUnit: codec.VPS,
					}},
				},
				{
					NaluType: byte(h265.NALUType_SPS_NUT),
					NumNalus: 1,
					Nalus: []amp4.HEVCNalu{{
						Length:  uint16(len(codec.SPS)),
						NALUnit: codec.SPS,
					}},
				},
				{
					NaluType: byte(h265.NALUType_PPS_NUT),
					NumNalus: 1,
					Nalus: []amp4.HEVCNalu{{
						Length:  uint16(len(codec.PPS)),
						NALUnit: codec.PPS,
					}},
				},
			},
		})
		if err != nil {
			return err
		}

	case *codecs.H264:
		_, err := w.WriteBoxStart(&amp4.VisualSampleEntry{ // <avc1>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeAvc1(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(info.Width),
			Height:          uint16(info.Height),
			Horizresolution: 4718592,
			Vertresolution:  4718592,
			FrameCount:      1,
			Depth:           24,
			PreDefined3:     -1,
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.AVCDecoderConfiguration{ // <avcc/>
			AnyTypeBox: amp4.AnyTypeBox{
				Type: amp4.BoxTypeAvcC(),
			},
			ConfigurationVersion:       1,
			Profile:                    info.H264SPS.ProfileIdc,
			ProfileCompatibility:       codec.SPS[2],
			Level:                      info.H264SPS.LevelIdc,
			LengthSizeMinusOne:         3,
			NumOfSequenceParameterSets: 1,
			SequenceParameterSets: []amp4.AVCParameterSet{
				{
					Length:  uint16(len(codec.SPS)),
					NALUnit: codec.SPS,
				},
			},
			NumOfPictureParameterSets: 1,
			PictureParameterSets: []amp4.AVCParameterSet{
				{
					Length:  uint16(len(codec.PPS)),
					NALUnit: codec.PPS,
				},
			},
		})
		if err != nil {
			return err
		}

	case *codecs.MPEG4Video: //nolint:dupl
		_, err := w.WriteBoxStart(&amp4.VisualSampleEntry{ // <mp4v>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeMp4v(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(info.Width),
			Height:          uint16(info.Height),
			Horizresolution: 4718592,
			Vertresolution:  4718592,
			FrameCount:      1,
			Depth:           24,
			PreDefined3:     -1,
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.Esds{ // <esds/>
			Descriptors: []amp4.Descriptor{
				{
					Tag:  amp4.ESDescrTag,
					Size: 32 + uint32(len(codec.Config)),
					ESDescriptor: &amp4.ESDescriptor{
						ESID: uint16(trackID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 18 + uint32(len(codec.Config)),
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: ObjectTypeIndicationVisualISO14496part2,
						StreamType:           StreamTypeVisualStream,
						Reserved:             true,
						MaxBitrate:           maxBitrate,
						AvgBitrate:           avgBitrate,
					},
				},
				{
					Tag:  amp4.DecSpecificInfoTag,
					Size: uint32(len(codec.Config)),
					Data: codec.Config,
				},
				{
					Tag:  amp4.SLConfigDescrTag,
					Size: 1,
					Data: []byte{0x02},
				},
			},
		})
		if err != nil {
			return err
		}

	case *codecs.MPEG1Video: //nolint:dupl
		_, err := w.WriteBoxStart(&amp4.VisualSampleEntry{ // <mp4v>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeMp4v(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(info.Width),
			Height:          uint16(info.Height),
			Horizresolution: 4718592,
			Vertresolution:  4718592,
			FrameCount:      1,
			Depth:           24,
			PreDefined3:     -1,
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.Esds{ // <esds/>
			Descriptors: []amp4.Descriptor{
				{
					Tag:  amp4.ESDescrTag,
					Size: 32 + uint32(len(codec.Config)),
					ESDescriptor: &amp4.ESDescriptor{
						ESID: uint16(trackID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 18 + uint32(len(codec.Config)),
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: ObjectTypeIndicationVisualISO1318part2Main,
						StreamType:           StreamTypeVisualStream,
						Reserved:             true,
						MaxBitrate:           maxBitrate,
						AvgBitrate:           avgBitrate,
					},
				},
				{
					Tag:  amp4.DecSpecificInfoTag,
					Size: uint32(len(codec.Config)),
					Data: codec.Config,
				},
				{
					Tag:  amp4.SLConfigDescrTag,
					Size: 1,
					Data: []byte{0x02},
				},
			},
		})
		if err != nil {
			return err
		}

	case *codecs.MJPEG: //nolint:dupl
		_, err := w.WriteBoxStart(&amp4.VisualSampleEntry{ // <mp4v>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeMp4v(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(info.Width),
			Height:          uint16(info.Height),
			Horizresolution: 4718592,
			Vertresolution:  4718592,
			FrameCount:      1,
			Depth:           24,
			PreDefined3:     -1,
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.Esds{ // <esds/>
			Descriptors: []amp4.Descriptor{
				{
					Tag:  amp4.ESDescrTag,
					Size: 27,
					ESDescriptor: &amp4.ESDescriptor{
						ESID: uint16(trackID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 13,
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: ObjectTypeIndicationVisualISO10918part1,
						StreamType:           StreamTypeVisualStream,
						Reserved:             true,
						MaxBitrate:           maxBitrate,
						AvgBitrate:           avgBitrate,
					},
				},
				{
					Tag:  amp4.SLConfigDescrTag,
					Size: 1,
					Data: []byte{0x02},
				},
			},
		})
		if err != nil {
			return err
		}

	case *codecs.Opus:
		_, err := w.WriteBoxStart(&amp4.AudioSampleEntry{ // <Opus>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeOpus(),
				},
				DataReferenceIndex: 1,
			},
			ChannelCount: uint16(codec.ChannelCount),
			SampleSize:   16,
			SampleRate:   48000 * 65536,
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.DOps{ // <dOps/>
			OutputChannelCount: uint8(codec.ChannelCount),
			PreSkip:            312,
			InputSampleRate:    48000,
		})
		if err != nil {
			return err
		}

	case *codecs.MPEG4Audio:
		var channelCount uint16

		if codec.Config.ChannelCount != 0 { //nolint:staticcheck
			channelCount = uint16(codec.Config.ChannelCount) //nolint:staticcheck
		} else {
			switch {
			case codec.Config.ChannelConfig >= 1 && codec.Config.ChannelConfig <= 6:
				channelCount = uint16(codec.Config.ChannelConfig)

			case codec.Config.ChannelConfig == 7:
				channelCount = 8

			default:
				return fmt.Errorf("MPEG-4 audio channelConfig = 0 is not supported (yet)")
			}
		}

		_, err := w.WriteBoxStart(&amp4.AudioSampleEntry{ // <mp4a>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeMp4a(),
				},
				DataReferenceIndex: 1,
			},
			ChannelCount: channelCount,
			SampleSize:   16,
			SampleRate:   uint32(codec.Config.SampleRate * 65536),
		})
		if err != nil {
			return err
		}

		enc, _ := codec.Config.Marshal()

		_, err = w.WriteBox(&amp4.Esds{ // <esds/>
			Descriptors: []amp4.Descriptor{
				{
					Tag:  amp4.ESDescrTag,
					Size: 32 + uint32(len(enc)),
					ESDescriptor: &amp4.ESDescriptor{
						ESID: uint16(trackID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 18 + uint32(len(enc)),
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: ObjectTypeIndicationAudioISO14496part3,
						StreamType:           StreamTypeAudioStream,
						Reserved:             true,
						MaxBitrate:           maxBitrate,
						AvgBitrate:           avgBitrate,
					},
				},
				{
					Tag:  amp4.DecSpecificInfoTag,
					Size: uint32(len(enc)),
					Data: enc,
				},
				{
					Tag:  amp4.SLConfigDescrTag,
					Size: 1,
					Data: []byte{0x02},
				},
			},
		})
		if err != nil {
			return err
		}

	case *codecs.MPEG1Audio:
		_, err := w.WriteBoxStart(&amp4.AudioSampleEntry{ // <mp4a>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeMp4a(),
				},
				DataReferenceIndex: 1,
			},
			ChannelCount: uint16(codec.ChannelCount),
			SampleSize:   16,
			SampleRate:   uint32(codec.SampleRate * 65536),
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.Esds{ // <esds/>
			Descriptors: []amp4.Descriptor{
				{
					Tag:  amp4.ESDescrTag,
					Size: 27,
					ESDescriptor: &amp4.ESDescriptor{
						ESID: uint16(trackID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 13,
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: ObjectTypeIndicationAudioISO11172part3,
						StreamType:           StreamTypeAudioStream,
						Reserved:             true,
						MaxBitrate:           maxBitrate,
						AvgBitrate:           avgBitrate,
					},
				},
				{
					Tag:  amp4.SLConfigDescrTag,
					Size: 1,
					Data: []byte{0x02},
				},
			},
		})
		if err != nil {
			return err
		}

	case *codecs.AC3:
		_, err := w.WriteBoxStart(&amp4.AudioSampleEntry{ // <ac-3>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeAC3(),
				},
				DataReferenceIndex: 1,
			},
			ChannelCount: uint16(codec.ChannelCount),
			SampleSize:   16,
			SampleRate:   uint32(codec.SampleRate * 65536),
		})
		if err != nil {
			return err
		}

		var fscod uint8
		switch codec.SampleRate {
		case 48000:
			fscod = 0
		case 44100:
			fscod = 1
		case 32000:
			fscod = 2
		default:
			return fmt.Errorf("unsupported sample rate: %v", codec.SampleRate)
		}

		_, err = w.WriteBox(&amp4.Dac3{ // <dac3/>
			Fscod: fscod,
			Bsid:  codec.Bsid,
			Bsmod: codec.Bsmod,
			Acmod: codec.Acmod,
			LfeOn: func() uint8 {
				if codec.LfeOn {
					return 1
				}
				return 0
			}(),
			BitRateCode: codec.BitRateCode,
		})
		if err != nil {
			return err
		}

	case *codecs.EAC3:
		_, err := w.WriteBoxStart(&amp4.AudioSampleEntry{ // <ec-3>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.StrToBoxType("ec-3"),
				},
				DataReferenceIndex: 1,
			},
			ChannelCount: uint16(codec.ChannelCount),
			SampleSize:   16,
			SampleRate:   uint32(codec.SampleRate * 65536),
		})
		if err != nil {
			return err
		}

		var fscod uint8
		switch codec.SampleRate {
		case 48000:
			fscod = 0
		case 44100:
			fscod = 1
		case 32000:
			fscod = 2
		default:
			return fmt.Errorf("unsupported sample rate: %v", codec.SampleRate)
		}

		_, err = w.WriteBox(&Dec3{
			DataRate:  codec.DataRate,
			NumIndSub: 0,
			Fscod:     fscod,
			Bsid:      16,
			Asvc:      boolToUint8(codec.Asvc),
			Bsmod:     codec.Bsmod,
			Acmod:     codec.Acmod,
			LfeOn:     boolToUint8(codec.LfeOn),
			NumDepSub: codec.NumDepSub,
			ChanLoc:   codec.ChanLoc,
		})
		if err != nil {
			return err
		}

	case *codecs.LPCM:
		_, err := w.WriteBoxStart(&amp4.AudioSampleEntry{ // <ipcm>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeIpcm(),
				},
				DataReferenceIndex: 1,
			},
			ChannelCount: uint16(codec.ChannelCount),
			SampleSize:   uint16(codec.BitDepth), // FFmpeg leaves this to 16 instead of using real bit depth
			SampleRate:   uint32(codec.SampleRate * 65536),
		})
		if err != nil {
			return err
		}

		_, err = w.WriteBox(&amp4.PcmC{ // <pcmC/>
			FormatFlags: func() uint8 {
				if codec.LittleEndian {
					return 1
				}
				return 0
			}(),
			PCMSampleSize: uint8(codec.BitDepth),
		})
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported codec: %T", codec)
	}

	return nil
}
