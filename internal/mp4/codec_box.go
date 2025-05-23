package mp4

import (
	"fmt"

	amp4 "github.com/abema/go-mp4"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

func boolToUint8(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}

// WriteCodecBoxes writes codec-related boxes.
func WriteCodecBoxes(w *Writer, codec mp4.Codec, trackID int, info *CodecInfo, avgBitrate, maxBitrate uint32) error {
	/*
		|av01| (AV1)
		|    |av1C|
		|vp09| (VP9)
		|    |vpcC|
		|hev1| (H265)
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
		|ipcm| (LPCM)
		|    |pcmC|
	*/

	switch codec := codec.(type) {
	case *mp4.CodecAV1:
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

	case *mp4.CodecVP9:
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

	case *mp4.CodecH265:
		_, err := w.WriteBoxStart(&amp4.VisualSampleEntry{ // <hev1>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeHev1(),
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

	case *mp4.CodecH264:
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

	case *mp4.CodecMPEG4Video: //nolint:dupl
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

	case *mp4.CodecMPEG1Video: //nolint:dupl
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

	case *mp4.CodecMJPEG: //nolint:dupl
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

	case *mp4.CodecOpus:
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

	case *mp4.CodecMPEG4Audio:
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

		enc, _ := codec.Marshal()

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

	case *mp4.CodecMPEG1Audio:
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

	case *mp4.CodecAC3:
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

		_, err = w.WriteBox(&amp4.Dac3{ // <dac3/>
			Fscod: codec.Fscod,
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

	case *mp4.CodecLPCM:
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
