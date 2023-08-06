package fmp4

import (
	"github.com/abema/go-mp4"

	"github.com/bluenviron/mediacommon/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/pkg/codecs/h265"
)

func boolToUint8(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}

// InitTrack is a track of Init.
type InitTrack struct {
	ID        int
	TimeScale uint32
	Codec     Codec
}

func (track *InitTrack) marshal(w *mp4Writer) error {
	/*
		   trak
		   - tkhd
		   - mdia
			 - mdhd
			 - hdlr
			 - minf
			   - vmhd (video)
			   - smhd (audio)
			   - dinf
				 - dref
				   - url
			   - stbl
				 - stsd
				   - av01 (AV1)
					 - av1C
					 - btrt
				   - vp09 (VP9)
					 - vpcC
					 - btrt
				   - hev1 (H265)
					 - hvcC
					 - btrt
				   - avc1 (H264)
					 - avcC
					 - btrt
				   - Opus (Opus)
					 - dOps
					 - btrt
				   - mp4a (MPEG-4 audio)
					 - esds
					 - btrt
				   - mp4a (MPEG-1 audio)
					 - esds
					 - btrt
				 - stts
				 - stsc
				 - stsz
				 - stco
	*/

	_, err := w.writeBoxStart(&mp4.Trak{}) // <trak>
	if err != nil {
		return err
	}

	var av1SequenceHeader *av1.SequenceHeader
	var h265SPS *h265.SPS
	var h264SPS *h264.SPS

	var width int
	var height int

	switch codec := track.Codec.(type) {
	case *CodecAV1:
		av1SequenceHeader = &av1.SequenceHeader{}
		err = av1SequenceHeader.Unmarshal(codec.SequenceHeader)
		if err != nil {
			return err
		}

		width = av1SequenceHeader.Width()
		height = av1SequenceHeader.Height()

	case *CodecVP9:
		width = codec.Width
		height = codec.Height

	case *CodecH265:
		h265SPS = &h265.SPS{}
		err = h265SPS.Unmarshal(codec.SPS)
		if err != nil {
			return err
		}

		width = h265SPS.Width()
		height = h265SPS.Height()

	case *CodecH264:
		h264SPS = &h264.SPS{}
		err = h264SPS.Unmarshal(codec.SPS)
		if err != nil {
			return err
		}

		width = h264SPS.Width()
		height = h264SPS.Height()
	}

	if track.Codec.IsVideo() {
		_, err = w.writeBox(&mp4.Tkhd{ // <tkhd/>
			FullBox: mp4.FullBox{
				Flags: [3]byte{0, 0, 3},
			},
			TrackID: uint32(track.ID),
			Width:   uint32(width * 65536),
			Height:  uint32(height * 65536),
			Matrix:  [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		})
		if err != nil {
			return err
		}
	} else {
		_, err = w.writeBox(&mp4.Tkhd{ // <tkhd/>
			FullBox: mp4.FullBox{
				Flags: [3]byte{0, 0, 3},
			},
			TrackID:        uint32(track.ID),
			AlternateGroup: 1,
			Volume:         256,
			Matrix:         [9]int32{0x10000, 0, 0, 0, 0x10000, 0, 0, 0, 0x40000000},
		})
		if err != nil {
			return err
		}
	}

	_, err = w.writeBoxStart(&mp4.Mdia{}) // <mdia>
	if err != nil {
		return err
	}

	_, err = w.writeBox(&mp4.Mdhd{ // <mdhd/>
		Timescale: track.TimeScale,
		Language:  [3]byte{'u', 'n', 'd'},
	})
	if err != nil {
		return err
	}

	if track.Codec.IsVideo() {
		_, err = w.writeBox(&mp4.Hdlr{ // <hdlr/>
			HandlerType: [4]byte{'v', 'i', 'd', 'e'},
			Name:        "VideoHandler",
		})
		if err != nil {
			return err
		}
	} else {
		_, err = w.writeBox(&mp4.Hdlr{ // <hdlr/>
			HandlerType: [4]byte{'s', 'o', 'u', 'n'},
			Name:        "SoundHandler",
		})
		if err != nil {
			return err
		}
	}

	_, err = w.writeBoxStart(&mp4.Minf{}) // <minf>
	if err != nil {
		return err
	}

	if track.Codec.IsVideo() {
		_, err = w.writeBox(&mp4.Vmhd{ // <vmhd/>
			FullBox: mp4.FullBox{
				Flags: [3]byte{0, 0, 1},
			},
		})
		if err != nil {
			return err
		}
	} else {
		_, err = w.writeBox(&mp4.Smhd{}) // <smhd/>
		if err != nil {
			return err
		}
	}

	_, err = w.writeBoxStart(&mp4.Dinf{}) // <dinf>
	if err != nil {
		return err
	}

	_, err = w.writeBoxStart(&mp4.Dref{ // <dref>
		EntryCount: 1,
	})
	if err != nil {
		return err
	}

	_, err = w.writeBox(&mp4.Url{ // <url/>
		FullBox: mp4.FullBox{
			Flags: [3]byte{0, 0, 1},
		},
	})
	if err != nil {
		return err
	}

	err = w.writeBoxEnd() // </dref>
	if err != nil {
		return err
	}

	err = w.writeBoxEnd() // </dinf>
	if err != nil {
		return err
	}

	_, err = w.writeBoxStart(&mp4.Stbl{}) // <stbl>
	if err != nil {
		return err
	}

	_, err = w.writeBoxStart(&mp4.Stsd{ // <stsd>
		EntryCount: 1,
	})
	if err != nil {
		return err
	}

	switch codec := track.Codec.(type) {
	case *CodecAV1:
		_, err = w.writeBoxStart(&mp4.VisualSampleEntry{ // <av01>
			SampleEntry: mp4.SampleEntry{
				AnyTypeBox: mp4.AnyTypeBox{
					Type: BoxTypeAv01(),
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

		bs, err := av1.BitstreamMarshal([][]byte{codec.SequenceHeader})
		if err != nil {
			return err
		}

		_, err = w.writeBox(&Av1C{ // <av1C/>
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
			ConfigOBUs:           bs,
		})
		if err != nil {
			return err
		}

	case *CodecVP9:
		_, err = w.writeBoxStart(&mp4.VisualSampleEntry{ // <vp09>
			SampleEntry: mp4.SampleEntry{
				AnyTypeBox: mp4.AnyTypeBox{
					Type: BoxTypeVp09(),
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

		_, err = w.writeBox(&VpcC{ // <vpcC/>
			FullBox: mp4.FullBox{
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

	case *CodecH265:
		_, err = w.writeBoxStart(&mp4.VisualSampleEntry{ // <hev1>
			SampleEntry: mp4.SampleEntry{
				AnyTypeBox: mp4.AnyTypeBox{
					Type: mp4.BoxTypeHev1(),
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

		_, err = w.writeBox(&mp4.HvcC{ // <hvcC/>
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
			NaluArrays: []mp4.HEVCNaluArray{
				{
					NaluType: byte(h265.NALUType_VPS_NUT),
					NumNalus: 1,
					Nalus: []mp4.HEVCNalu{{
						Length:  uint16(len(codec.VPS)),
						NALUnit: codec.VPS,
					}},
				},
				{
					NaluType: byte(h265.NALUType_SPS_NUT),
					NumNalus: 1,
					Nalus: []mp4.HEVCNalu{{
						Length:  uint16(len(codec.SPS)),
						NALUnit: codec.SPS,
					}},
				},
				{
					NaluType: byte(h265.NALUType_PPS_NUT),
					NumNalus: 1,
					Nalus: []mp4.HEVCNalu{{
						Length:  uint16(len(codec.PPS)),
						NALUnit: codec.PPS,
					}},
				},
			},
		})
		if err != nil {
			return err
		}

	case *CodecH264:
		_, err = w.writeBoxStart(&mp4.VisualSampleEntry{ // <avc1>
			SampleEntry: mp4.SampleEntry{
				AnyTypeBox: mp4.AnyTypeBox{
					Type: mp4.BoxTypeAvc1(),
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

		_, err = w.writeBox(&mp4.AVCDecoderConfiguration{ // <avcc/>
			AnyTypeBox: mp4.AnyTypeBox{
				Type: mp4.BoxTypeAvcC(),
			},
			ConfigurationVersion:       1,
			Profile:                    h264SPS.ProfileIdc,
			ProfileCompatibility:       codec.SPS[2],
			Level:                      h264SPS.LevelIdc,
			LengthSizeMinusOne:         3,
			NumOfSequenceParameterSets: 1,
			SequenceParameterSets: []mp4.AVCParameterSet{
				{
					Length:  uint16(len(codec.SPS)),
					NALUnit: codec.SPS,
				},
			},
			NumOfPictureParameterSets: 1,
			PictureParameterSets: []mp4.AVCParameterSet{
				{
					Length:  uint16(len(codec.PPS)),
					NALUnit: codec.PPS,
				},
			},
		})
		if err != nil {
			return err
		}

	case *CodecOpus:
		_, err = w.writeBoxStart(&mp4.AudioSampleEntry{ // <Opus>
			SampleEntry: mp4.SampleEntry{
				AnyTypeBox: mp4.AnyTypeBox{
					Type: BoxTypeOpus(),
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

		_, err = w.writeBox(&DOps{ // <dOps/>
			OutputChannelCount: uint8(codec.ChannelCount),
			PreSkip:            312,
			InputSampleRate:    48000,
		})
		if err != nil {
			return err
		}

	case *CodecMPEG4Audio:
		_, err = w.writeBoxStart(&mp4.AudioSampleEntry{ // <mp4a>
			SampleEntry: mp4.SampleEntry{
				AnyTypeBox: mp4.AnyTypeBox{
					Type: mp4.BoxTypeMp4a(),
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

		_, err = w.writeBox(&mp4.Esds{ // <esds/>
			Descriptors: []mp4.Descriptor{
				{
					Tag:  mp4.ESDescrTag,
					Size: 32 + uint32(len(enc)),
					ESDescriptor: &mp4.ESDescriptor{
						ESID: uint16(track.ID),
					},
				},
				{
					Tag:  mp4.DecoderConfigDescrTag,
					Size: 18 + uint32(len(enc)),
					DecoderConfigDescriptor: &mp4.DecoderConfigDescriptor{
						ObjectTypeIndication: 0x40,
						StreamType:           0x05,
						UpStream:             false,
						Reserved:             true,
						MaxBitrate:           128825,
						AvgBitrate:           128825,
					},
				},
				{
					Tag:  mp4.DecSpecificInfoTag,
					Size: uint32(len(enc)),
					Data: enc,
				},
				{
					Tag:  mp4.SLConfigDescrTag,
					Size: 1,
					Data: []byte{0x02},
				},
			},
		})
		if err != nil {
			return err
		}
	}

	if track.Codec.IsVideo() {
		_, err = w.writeBox(&mp4.Btrt{ // <btrt/>
			MaxBitrate: 1000000,
			AvgBitrate: 1000000,
		})
		if err != nil {
			return err
		}
	} else {
		_, err = w.writeBox(&mp4.Btrt{ // <btrt/>
			MaxBitrate: 128825,
			AvgBitrate: 128825,
		})
		if err != nil {
			return err
		}
	}

	err = w.writeBoxEnd() // </*>
	if err != nil {
		return err
	}

	err = w.writeBoxEnd() // </stsd>
	if err != nil {
		return err
	}

	_, err = w.writeBox(&mp4.Stts{ // <stts>
	})
	if err != nil {
		return err
	}

	_, err = w.writeBox(&mp4.Stsc{ // <stsc>
	})
	if err != nil {
		return err
	}

	_, err = w.writeBox(&mp4.Stsz{ // <stsz>
	})
	if err != nil {
		return err
	}

	_, err = w.writeBox(&mp4.Stco{ // <stco>
	})
	if err != nil {
		return err
	}

	err = w.writeBoxEnd() // </stbl>
	if err != nil {
		return err
	}

	err = w.writeBoxEnd() // </minf>
	if err != nil {
		return err
	}

	err = w.writeBoxEnd() // </mdia>
	if err != nil {
		return err
	}

	err = w.writeBoxEnd() // </trak>
	if err != nil {
		return err
	}

	return nil
}
