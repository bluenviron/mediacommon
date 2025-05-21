package fmp4

import (
	"fmt"

	amp4 "github.com/abema/go-mp4"

	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

func boolToUint8(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}

// InitTrack is a track of Init.
type InitTrack struct {
	// ID, starts from 1.
	ID int

	// time scale.
	TimeScale uint32

	// maximum bitrate.
	// it defaults to 1MB for video tracks, 128k for audio tracks.
	MaxBitrate uint32

	// average bitrate.
	// it defaults to 1MB for video tracks, 128k for audio tracks.
	AvgBitrate uint32

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
		|    |    |    |    |    |av01| (AV1)
		|    |    |    |    |    |    |av1C|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |    |vp09| (VP9)
		|    |    |    |    |    |    |vpcC|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |    |hev1| (H265)
		|    |    |    |    |    |    |hvcC|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |    |avc1| (H264)
		|    |    |    |    |    |    |avcC|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |    |mp4v| (MPEG-4/2/1 video, MJPEG)
		|    |    |    |    |    |    |esds|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |    |Opus| (Opus)
		|    |    |    |    |    |    |dOps|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |    |mp4a| (MPEG-4/1 audio)
		|    |    |    |    |    |    |esds|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |    |ac-3| (AC-3)
		|    |    |    |    |    |    |dac3|
		|    |    |    |    |    |    |btrt|
		|    |    |    |    |    |ipcm| (LPCM)
		|    |    |    |    |    |    |pcmC|
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

	var av1SequenceHeader *av1.SequenceHeader
	var h265SPS *h265.SPS
	var h264SPS *h264.SPS

	var width int
	var height int

	switch codec := it.Codec.(type) {
	case *mp4.CodecAV1:
		av1SequenceHeader = &av1.SequenceHeader{}
		err = av1SequenceHeader.Unmarshal(codec.SequenceHeader)
		if err != nil {
			return fmt.Errorf("unable to parse AV1 sequence header: %w", err)
		}

		width = av1SequenceHeader.Width()
		height = av1SequenceHeader.Height()

	case *mp4.CodecVP9:
		if codec.Width == 0 {
			return fmt.Errorf("VP9 parameters not provided")
		}

		width = codec.Width
		height = codec.Height

	case *mp4.CodecH265:
		if len(codec.VPS) == 0 || len(codec.SPS) == 0 || len(codec.PPS) == 0 {
			return fmt.Errorf("H265 parameters not provided")
		}

		h265SPS = &h265.SPS{}
		err = h265SPS.Unmarshal(codec.SPS)
		if err != nil {
			return fmt.Errorf("unable to parse H265 SPS: %w", err)
		}

		width = h265SPS.Width()
		height = h265SPS.Height()

	case *mp4.CodecH264:
		if len(codec.SPS) == 0 || len(codec.PPS) == 0 {
			return fmt.Errorf("H264 parameters not provided")
		}

		h264SPS = &h264.SPS{}
		err = h264SPS.Unmarshal(codec.SPS)
		if err != nil {
			return fmt.Errorf("unable to parse H264 SPS: %w", err)
		}

		width = h264SPS.Width()
		height = h264SPS.Height()

	case *mp4.CodecMPEG4Video:
		if len(codec.Config) == 0 {
			return fmt.Errorf("MPEG-4 Video config not provided")
		}

		// TODO: parse config and use real values
		width = 800
		height = 600

	case *mp4.CodecMPEG1Video:
		if len(codec.Config) == 0 {
			return fmt.Errorf("MPEG-1/2 Video config not provided")
		}

		// TODO: parse config and use real values
		width = 800
		height = 600

	case *mp4.CodecMJPEG:
		if codec.Width == 0 {
			return fmt.Errorf("M-JPEG parameters not provided")
		}

		width = codec.Width
		height = codec.Height
	}

	if it.Codec.IsVideo() {
		_, err = w.WriteBox(&amp4.Tkhd{ // <tkhd/>
			FullBox: amp4.FullBox{
				Flags: [3]byte{0, 0, 3},
			},
			TrackID: uint32(it.ID),
			Width:   uint32(width * 65536),
			Height:  uint32(height * 65536),
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

	maxBitrate := it.MaxBitrate
	if maxBitrate == 0 {
		if it.Codec.IsVideo() {
			maxBitrate = 1000000
		} else {
			maxBitrate = 128825
		}
	}

	avgBitrate := it.AvgBitrate
	if avgBitrate == 0 {
		if it.Codec.IsVideo() {
			avgBitrate = 1000000
		} else {
			avgBitrate = 128825
		}
	}

	switch codec := it.Codec.(type) {
	case *mp4.CodecAV1:
		_, err = w.WriteBoxStart(&amp4.VisualSampleEntry{ // <av01>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeAv01(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(width),
			Height:          uint16(height),
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
			SeqProfile:           av1SequenceHeader.SeqProfile,
			SeqLevelIdx0:         av1SequenceHeader.SeqLevelIdx[0],
			SeqTier0:             boolToUint8(av1SequenceHeader.SeqTier[0]),
			HighBitdepth:         boolToUint8(av1SequenceHeader.ColorConfig.HighBitDepth),
			TwelveBit:            boolToUint8(av1SequenceHeader.ColorConfig.TwelveBit),
			Monochrome:           boolToUint8(av1SequenceHeader.ColorConfig.MonoChrome),
			ChromaSubsamplingX:   boolToUint8(av1SequenceHeader.ColorConfig.SubsamplingX),
			ChromaSubsamplingY:   boolToUint8(av1SequenceHeader.ColorConfig.SubsamplingY),
			ChromaSamplePosition: uint8(av1SequenceHeader.ColorConfig.ChromaSamplePosition),
			ConfigOBUs:           enc,
		})
		if err != nil {
			return err
		}

	case *mp4.CodecVP9:
		_, err = w.WriteBoxStart(&amp4.VisualSampleEntry{ // <vp09>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeVp09(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(width),
			Height:          uint16(height),
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
		_, err = w.WriteBoxStart(&amp4.VisualSampleEntry{ // <hev1>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeHev1(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(width),
			Height:          uint16(height),
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
			GeneralProfileIdc:           h265SPS.ProfileTierLevel.GeneralProfileIdc,
			GeneralProfileCompatibility: h265SPS.ProfileTierLevel.GeneralProfileCompatibilityFlag,
			GeneralConstraintIndicator: [6]uint8{
				codec.SPS[7], codec.SPS[8], codec.SPS[9],
				codec.SPS[10], codec.SPS[11], codec.SPS[12],
			},
			GeneralLevelIdc: h265SPS.ProfileTierLevel.GeneralLevelIdc,
			// MinSpatialSegmentationIdc
			// ParallelismType
			ChromaFormatIdc:      uint8(h265SPS.ChromaFormatIdc),
			BitDepthLumaMinus8:   uint8(h265SPS.BitDepthLumaMinus8),
			BitDepthChromaMinus8: uint8(h265SPS.BitDepthChromaMinus8),
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
		_, err = w.WriteBoxStart(&amp4.VisualSampleEntry{ // <avc1>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeAvc1(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(width),
			Height:          uint16(height),
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
			Profile:                    h264SPS.ProfileIdc,
			ProfileCompatibility:       codec.SPS[2],
			Level:                      h264SPS.LevelIdc,
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
		_, err = w.WriteBoxStart(&amp4.VisualSampleEntry{ // <mp4v>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeMp4v(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(width),
			Height:          uint16(height),
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
						ESID: uint16(it.ID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 18 + uint32(len(codec.Config)),
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: imp4.ObjectTypeIndicationVisualISO14496part2,
						StreamType:           imp4.StreamTypeVisualStream,
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
		_, err = w.WriteBoxStart(&amp4.VisualSampleEntry{ // <mp4v>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeMp4v(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(width),
			Height:          uint16(height),
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
						ESID: uint16(it.ID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 18 + uint32(len(codec.Config)),
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: imp4.ObjectTypeIndicationVisualISO1318part2Main,
						StreamType:           imp4.StreamTypeVisualStream,
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
		_, err = w.WriteBoxStart(&amp4.VisualSampleEntry{ // <mp4v>
			SampleEntry: amp4.SampleEntry{
				AnyTypeBox: amp4.AnyTypeBox{
					Type: amp4.BoxTypeMp4v(),
				},
				DataReferenceIndex: 1,
			},
			Width:           uint16(width),
			Height:          uint16(height),
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
						ESID: uint16(it.ID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 13,
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: imp4.ObjectTypeIndicationVisualISO10918part1,
						StreamType:           imp4.StreamTypeVisualStream,
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
		_, err = w.WriteBoxStart(&amp4.AudioSampleEntry{ // <Opus>
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
		_, err = w.WriteBoxStart(&amp4.AudioSampleEntry{ // <mp4a>
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

		enc, _ := codec.Config.Marshal()

		_, err = w.WriteBox(&amp4.Esds{ // <esds/>
			Descriptors: []amp4.Descriptor{
				{
					Tag:  amp4.ESDescrTag,
					Size: 32 + uint32(len(enc)),
					ESDescriptor: &amp4.ESDescriptor{
						ESID: uint16(it.ID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 18 + uint32(len(enc)),
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: imp4.ObjectTypeIndicationAudioISO14496part3,
						StreamType:           imp4.StreamTypeAudioStream,
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
		_, err = w.WriteBoxStart(&amp4.AudioSampleEntry{ // <mp4a>
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
						ESID: uint16(it.ID),
					},
				},
				{
					Tag:  amp4.DecoderConfigDescrTag,
					Size: 13,
					DecoderConfigDescriptor: &amp4.DecoderConfigDescriptor{
						ObjectTypeIndication: imp4.ObjectTypeIndicationAudioISO11172part3,
						StreamType:           imp4.StreamTypeAudioStream,
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
		_, err = w.WriteBoxStart(&amp4.AudioSampleEntry{ // <ac-3>
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
		_, err = w.WriteBoxStart(&amp4.AudioSampleEntry{ // <ipcm>
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
