package h265

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesSPS = []struct {
	name   string
	byts   []byte
	sps    SPS
	width  int
	height int
	fps    float64
}{
	{
		"1920x1080",
		[]byte{
			0x42, 0x01, 0x01, 0x01, 0x60, 0x00, 0x00, 0x03,
			0x00, 0x90, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03,
			0x00, 0x78, 0xa0, 0x03, 0xc0, 0x80, 0x10, 0xe5,
			0x96, 0x66, 0x69, 0x24, 0xca, 0xe0, 0x10, 0x00,
			0x00, 0x03, 0x00, 0x10, 0x00, 0x00, 0x03, 0x01,
			0xe0, 0x80,
		},
		SPS{
			TemporalIDNestingFlag: true,
			ProfileTierLevel: SPS_ProfileTierLevel{
				GeneralProfileIdc: 1,
				GeneralProfileCompatibilityFlag: [32]bool{
					false, true, true, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
				},
				GeneralProgressiveSourceFlag:   true,
				GeneralFrameOnlyConstraintFlag: true,
				GeneralLevelIdc:                120,
			},
			ChromaFormatIdc:                      1,
			PicWidthInLumaSamples:                1920,
			PicHeightInLumaSamples:               1080,
			Log2MaxPicOrderCntLsbMinus4:          4,
			SubLayerOrderingInfoPresentFlag:      true,
			MaxDecPicBufferingMinus1:             []uint32{5},
			MaxNumReorderPics:                    []uint32{2},
			MaxLatencyIncreasePlus1:              []uint32{5},
			Log2DiffMaxMinLumaCodingBlockSize:    3,
			Log2DiffMaxMinLumaTransformBlockSize: 3,
			SampleAdaptiveOffsetEnabledFlag:      true,
			TemporalMvpEnabledFlag:               true,
			StrongIntraSmoothingEnabledFlag:      true,
			VUI: &SPS_VUI{
				TimingInfo: &SPS_TimingInfo{
					NumUnitsInTick: 1,
					TimeScale:      30,
				},
			},
		},
		1920,
		1080,
		30,
	},
	{
		"1920x800",
		[]byte{
			0x42, 0x01, 0x01, 0x01, 0x60, 0x00, 0x00, 0x03,
			0x00, 0x90, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03,
			0x00, 0x78, 0xa0, 0x03, 0xc0, 0x80, 0x32, 0x16,
			0x59, 0x59, 0xa4, 0x93, 0x2b, 0xc0, 0x5a, 0x80,
			0x80, 0x80, 0x82, 0x00, 0x00, 0x07, 0xd2, 0x00,
			0x00, 0xbb, 0x80, 0x10,
		},
		SPS{
			TemporalIDNestingFlag: true,
			ProfileTierLevel: SPS_ProfileTierLevel{
				GeneralProfileIdc: 1,
				GeneralProfileCompatibilityFlag: [32]bool{
					false, true, true, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
				},
				GeneralProgressiveSourceFlag:   true,
				GeneralFrameOnlyConstraintFlag: true,
				GeneralLevelIdc:                120,
			},
			ChromaFormatIdc:                      1,
			PicWidthInLumaSamples:                1920,
			PicHeightInLumaSamples:               800,
			Log2MaxPicOrderCntLsbMinus4:          4,
			SubLayerOrderingInfoPresentFlag:      true,
			MaxDecPicBufferingMinus1:             []uint32{4},
			MaxNumReorderPics:                    []uint32{2},
			MaxLatencyIncreasePlus1:              []uint32{5},
			Log2DiffMaxMinLumaCodingBlockSize:    3,
			Log2DiffMaxMinLumaTransformBlockSize: 3,
			SampleAdaptiveOffsetEnabledFlag:      true,
			TemporalMvpEnabledFlag:               true,
			StrongIntraSmoothingEnabledFlag:      true,
			VUI: &SPS_VUI{
				AspectRatioInfoPresentFlag:   true,
				AspectRatioIdc:               1,
				VideoSignalTypePresentFlag:   true,
				VideoFormat:                  5,
				ColourDescriptionPresentFlag: true,
				ColourPrimaries:              1,
				TransferCharacteristics:      1,
				MatrixCoefficients:           1,
				TimingInfo: &SPS_TimingInfo{
					NumUnitsInTick: 1001,
					TimeScale:      24000,
				},
			},
		},
		1920,
		800,
		23.976023976023978,
	},
	{
		"1280x720",
		[]byte{
			0x42, 0x01, 0x01, 0x04, 0x08, 0x00, 0x00, 0x03,
			0x00, 0x98, 0x08, 0x00, 0x00, 0x03, 0x00, 0x00,
			0x5d, 0x90, 0x00, 0x50, 0x10, 0x05, 0xa2, 0x29,
			0x4b, 0x74, 0x94, 0x98, 0x5f, 0xfe, 0x00, 0x02,
			0x00, 0x02, 0xd4, 0x04, 0x04, 0x04, 0x10, 0x00,
			0x00, 0x03, 0x00, 0x10, 0x00, 0x00, 0x03, 0x01,
			0xe0, 0x80,
		},
		SPS{
			TemporalIDNestingFlag: true,
			ProfileTierLevel: SPS_ProfileTierLevel{
				GeneralProfileIdc: 4,
				GeneralProfileCompatibilityFlag: [32]bool{
					false, false, false, false, true, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
				},
				GeneralProgressiveSourceFlag:      true,
				GeneralFrameOnlyConstraintFlag:    true,
				GeneralMax12bitConstraintFlag:     true,
				GeneralLowerBitRateConstraintFlag: true,
				GeneralLevelIdc:                   93,
			},
			ChromaFormatIdc:                      3,
			PicWidthInLumaSamples:                1280,
			PicHeightInLumaSamples:               720,
			BitDepthLumaMinus8:                   4,
			BitDepthChromaMinus8:                 4,
			Log2MaxPicOrderCntLsbMinus4:          4,
			SubLayerOrderingInfoPresentFlag:      true,
			MaxDecPicBufferingMinus1:             []uint32{2},
			MaxNumReorderPics:                    []uint32{0},
			MaxLatencyIncreasePlus1:              []uint32{1},
			Log2MinLumaCodingBlockSizeMinus3:     1,
			Log2DiffMaxMinLumaCodingBlockSize:    1,
			Log2DiffMaxMinLumaTransformBlockSize: 3,
			TemporalMvpEnabledFlag:               true,
			StrongIntraSmoothingEnabledFlag:      true,
			VUI: &SPS_VUI{
				AspectRatioInfoPresentFlag:   true,
				AspectRatioIdc:               255,
				SarWidth:                     1,
				SarHeight:                    1,
				VideoSignalTypePresentFlag:   true,
				VideoFormat:                  5,
				ColourDescriptionPresentFlag: true,
				ColourPrimaries:              1,
				TransferCharacteristics:      1,
				MatrixCoefficients:           1,
				TimingInfo: &SPS_TimingInfo{
					NumUnitsInTick: 1,
					TimeScale:      30,
				},
			},
		},
		1280,
		720,
		30,
	},
	{
		"10 bit",
		[]byte{
			0x42, 0x01, 0x01, 0x22, 0x20, 0x00, 0x00, 0x03,
			0x00, 0x90, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03,
			0x00, 0x78, 0xa0, 0x03, 0xc0, 0x80, 0x10, 0xe4,
			0xd9, 0x66, 0x66, 0x92, 0x4c, 0xaf, 0x01, 0x01,
			0x00, 0x00, 0x03, 0x00, 0x64, 0x00, 0x00, 0x0b,
			0xb5, 0x08,
		},
		SPS{
			TemporalIDNestingFlag: true,
			ProfileTierLevel: SPS_ProfileTierLevel{
				GeneralTierFlag:   1,
				GeneralProfileIdc: 2,
				GeneralProfileCompatibilityFlag: [32]bool{
					false, false, true, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
				},
				GeneralProgressiveSourceFlag:   true,
				GeneralFrameOnlyConstraintFlag: true,
				GeneralLevelIdc:                120,
			},
			ChromaFormatIdc:                      1,
			PicWidthInLumaSamples:                1920,
			PicHeightInLumaSamples:               1080,
			BitDepthLumaMinus8:                   2,
			BitDepthChromaMinus8:                 2,
			Log2MaxPicOrderCntLsbMinus4:          4,
			SubLayerOrderingInfoPresentFlag:      true,
			MaxDecPicBufferingMinus1:             []uint32{5},
			MaxNumReorderPics:                    []uint32{2},
			MaxLatencyIncreasePlus1:              []uint32{5},
			Log2DiffMaxMinLumaCodingBlockSize:    3,
			Log2DiffMaxMinLumaTransformBlockSize: 3,
			SampleAdaptiveOffsetEnabledFlag:      true,
			TemporalMvpEnabledFlag:               true,
			StrongIntraSmoothingEnabledFlag:      true,
			VUI: &SPS_VUI{
				AspectRatioInfoPresentFlag: true,
				AspectRatioIdc:             1,
				TimingInfo: &SPS_TimingInfo{
					NumUnitsInTick: 100,
					TimeScale:      2997,
				},
			},
		},
		1920,
		1080,
		29.97,
	},
	{
		"nvenc",
		[]byte{
			0x42, 0x01, 0x01, 0x01, 0x40, 0x00, 0x00, 0x03,
			0x00, 0x00, 0x03, 0x00, 0x00, 0x03, 0x00, 0x00,
			0x03, 0x00, 0x7b, 0xa0, 0x03, 0xc0, 0x80, 0x11,
			0x07, 0xcb, 0x96, 0xb4, 0xa4, 0x25, 0x92, 0xe3,
			0x01, 0x6a, 0x02, 0x02, 0x02, 0x08, 0x00, 0x00,
			0x03, 0x00, 0x08, 0x00, 0x00, 0x03, 0x01, 0xe3,
			0x00, 0x2e, 0xf2, 0x88, 0x00, 0x07, 0x27, 0x0c,
			0x00, 0x00, 0x98, 0x96, 0x82,
		},
		SPS{
			TemporalIDNestingFlag: true,
			ProfileTierLevel: SPS_ProfileTierLevel{
				GeneralProfileIdc: 1,
				GeneralProfileCompatibilityFlag: [32]bool{
					false, true, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
				},
				GeneralLevelIdc: 123,
			},
			ChromaFormatIdc:        1,
			PicWidthInLumaSamples:  1920,
			PicHeightInLumaSamples: 1088,
			ConformanceWindow: &SPS_Window{
				BottomOffset: 4,
			},
			Log2MaxPicOrderCntLsbMinus4:          4,
			SubLayerOrderingInfoPresentFlag:      true,
			MaxDecPicBufferingMinus1:             []uint32{1},
			MaxNumReorderPics:                    []uint32{0},
			MaxLatencyIncreasePlus1:              []uint32{0},
			Log2MinLumaCodingBlockSizeMinus3:     1,
			Log2DiffMaxMinLumaCodingBlockSize:    1,
			Log2DiffMaxMinLumaTransformBlockSize: 3,
			MaxTransformHierarchyDepthInter:      3,
			AmpEnabledFlag:                       true,
			SampleAdaptiveOffsetEnabledFlag:      true,
			ShortTermRefPicSets: []*SPS_ShortTermRefPicSet{{
				NumNegativePics:     1,
				DeltaPocS0:          []int32{-1},
				UsedByCurrPicS0Flag: []bool{true},
			}},
			VUI: &SPS_VUI{
				AspectRatioInfoPresentFlag:   true,
				AspectRatioIdc:               1,
				VideoSignalTypePresentFlag:   true,
				VideoFormat:                  5,
				ColourDescriptionPresentFlag: true,
				ColourPrimaries:              1,
				TransferCharacteristics:      1,
				MatrixCoefficients:           1,
				TimingInfo: &SPS_TimingInfo{
					NumUnitsInTick: 1,
					TimeScale:      60,
				},
			},
		},
		1920,
		1080,
		60,
	},
	{
		"avigilon",
		[]byte{
			0x42, 0x01, 0x01, 0x01, 0x60, 0x00, 0x00, 0x03,
			0x00, 0x80, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03,
			0x00, 0x96, 0xa0, 0x01, 0x80, 0x20, 0x06, 0xc1,
			0xfe, 0x36, 0xbb, 0xb5, 0x37, 0x77, 0x25, 0xd6,
			0x02, 0xdc, 0x04, 0x04, 0x04, 0x10, 0x00, 0x00,
			0x3e, 0x80, 0x00, 0x04, 0x26, 0x87, 0x21, 0xde,
			0xe5, 0x10, 0x01, 0x6e, 0x20, 0x00, 0x66, 0xff,
			0x00, 0x0b, 0x71, 0x00, 0x03, 0x37, 0xf8, 0x80,
		},
		SPS{
			TemporalIDNestingFlag: true,
			ProfileTierLevel: SPS_ProfileTierLevel{
				GeneralProfileIdc: 1,
				GeneralProfileCompatibilityFlag: [32]bool{
					false, true, true, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
				},
				GeneralProgressiveSourceFlag: true,
				GeneralLevelIdc:              150,
			},
			ChromaFormatIdc:                      1,
			PicWidthInLumaSamples:                3072,
			PicHeightInLumaSamples:               1728,
			ConformanceWindow:                    &SPS_Window{},
			Log2MaxPicOrderCntLsbMinus4:          12,
			SubLayerOrderingInfoPresentFlag:      true,
			MaxDecPicBufferingMinus1:             []uint32{1},
			MaxNumReorderPics:                    []uint32{0},
			MaxLatencyIncreasePlus1:              []uint32{0},
			Log2DiffMaxMinLumaCodingBlockSize:    2,
			SampleAdaptiveOffsetEnabledFlag:      true,
			PcmEnabledFlag:                       true,
			PcmSampleBitDepthLumaMinus1:          7,
			PcmSampleBitDepthChromaMinus1:        7,
			Log2DiffMaxMinLumaTransformBlockSize: 2,
			MaxTransformHierarchyDepthInter:      1,
			Log2MinPcmLumaCodingBlockSizeMinus3:  2,
			ShortTermRefPicSets: []*SPS_ShortTermRefPicSet{
				{
					NumNegativePics:     1,
					DeltaPocS0:          []int32{-1},
					UsedByCurrPicS0Flag: []bool{true},
				},
			},
			TemporalMvpEnabledFlag: true,
			VUI: &SPS_VUI{
				AspectRatioInfoPresentFlag:   true,
				AspectRatioIdc:               1,
				VideoSignalTypePresentFlag:   true,
				VideoFormat:                  5,
				VideoFullRangeFlag:           true,
				ColourDescriptionPresentFlag: true,
				ColourPrimaries:              1,
				TransferCharacteristics:      1,
				MatrixCoefficients:           1,
				TimingInfo: &SPS_TimingInfo{
					NumUnitsInTick: 1000,
					TimeScale:      17000,
				},
			},
		},
		3072,
		1728,
		17,
	},
	{
		"long_term_ref_pics_present_flag",
		[]byte{
			0x42, 0x01, 0x01, 0x01, 0x60, 0x00, 0x00, 0x03,
			0x00, 0xb0, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03,
			0x00, 0x5d, 0xa0, 0x02, 0x80, 0x80, 0x2d, 0x16,
			0x36, 0xb9, 0x24, 0xcb, 0xf0, 0x08, 0x00, 0x00,
			0x03, 0x00, 0x08, 0x00, 0x00, 0x03, 0x01, 0x95,
			0x08,
		},
		SPS{
			TemporalIDNestingFlag: true,
			ProfileTierLevel: SPS_ProfileTierLevel{
				GeneralProfileIdc: 1,
				GeneralProfileCompatibilityFlag: [32]bool{
					false, true, true, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
				},
				GeneralProgressiveSourceFlag:   true,
				GeneralNonPackedConstraintFlag: true,
				GeneralFrameOnlyConstraintFlag: true,
				GeneralLevelIdc:                93,
			},
			ChromaFormatIdc:                      1,
			PicWidthInLumaSamples:                1280,
			PicHeightInLumaSamples:               720,
			Log2MaxPicOrderCntLsbMinus4:          12,
			SubLayerOrderingInfoPresentFlag:      true,
			MaxDecPicBufferingMinus1:             []uint32{1},
			MaxNumReorderPics:                    []uint32{0},
			MaxLatencyIncreasePlus1:              []uint32{0},
			Log2DiffMaxMinLumaCodingBlockSize:    3,
			Log2DiffMaxMinLumaTransformBlockSize: 3,
			SampleAdaptiveOffsetEnabledFlag:      true,
			LongTermRefPicsPresentFlag:           true,
			TemporalMvpEnabledFlag:               true,
			StrongIntraSmoothingEnabledFlag:      true,
			VUI: &SPS_VUI{
				TimingInfo: &SPS_TimingInfo{
					NumUnitsInTick:              1,
					TimeScale:                   50,
					POCProportionalToTimingFlag: true,
					NumTicksPOCDiffOneMinus1:    1,
				},
			},
		},
		1280,
		720,
		50,
	},
	{
		"scaling list data",
		[]byte{
			0x42, 0x01, 0x01, 0x01, 0x60, 0x00, 0x00, 0x03,
			0x00, 0x90, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03,
			0x00, 0x96, 0xa0, 0x01, 0x00, 0x20, 0x06, 0x01,
			0x63, 0x4b, 0xb9, 0x08, 0x4e, 0x51, 0x11, 0x18,
			0x8e, 0x09, 0x30, 0x24, 0x40, 0xdd, 0x28, 0x52,
			0x1c, 0xc1, 0x27, 0x06, 0x18, 0x1b, 0xb2, 0xa1,
			0x91, 0x58, 0xae, 0x16, 0xc0, 0xf1, 0x07, 0xd0,
			0x80, 0x20, 0x82, 0x8c, 0x16, 0x70, 0x35, 0x7c,
			0xa5, 0x24, 0x99, 0x3a, 0xaf, 0x4b, 0xa4, 0xbb,
			0x49, 0x2f, 0x20, 0x81, 0x11, 0x32, 0x0c, 0x18,
			0x30, 0x68, 0xd1, 0x80, 0xb0, 0x08, 0x10, 0x20,
			0xc0, 0x80, 0x0f, 0x81, 0xfc, 0x1f, 0x7c, 0xa3,
			0x22, 0x30, 0x87, 0x19, 0xe3, 0x3e, 0x3b, 0xf0,
			0x97, 0xf0, 0xc7, 0xe1, 0x0f, 0x83, 0x0f, 0x07,
			0xdf, 0xf2, 0xa1, 0x12, 0x34, 0x4e, 0x4f, 0x25,
			0x5c, 0x95, 0xb9, 0x29, 0x5b, 0x9a, 0x23, 0x13,
			0x10, 0x08, 0x01, 0x04, 0x10, 0x82, 0x10, 0x20,
			0x01, 0x03, 0x02, 0x08, 0x1f, 0xbf, 0xf0, 0x80,
			0x42, 0x10, 0xc2, 0x1c, 0x31, 0xe1, 0x0f, 0x84,
			0x3f, 0x08, 0x7f, 0x0a, 0x7e, 0x14, 0xf8, 0x3e,
			0xff, 0xfc, 0xa5, 0x26, 0x4c, 0x9d, 0x57, 0xa5,
			0xd2, 0x5d, 0xa4, 0x97, 0x90, 0x40, 0x88, 0x99,
			0x06, 0x0c, 0x18, 0x34, 0x68, 0xc0, 0x58, 0x04,
			0x08, 0x10, 0x60, 0x40, 0x07, 0xc0, 0xfe, 0x0f,
			0xbe, 0x51, 0x04, 0x88, 0xc2, 0x1c, 0x67, 0x8c,
			0xf8, 0xef, 0xc2, 0x5f, 0xc3, 0x1f, 0x84, 0x3e,
			0x0c, 0x3c, 0x1f, 0x7f, 0xca, 0x88, 0x49, 0x1a,
			0x27, 0x27, 0x92, 0xae, 0x4a, 0xdc, 0x94, 0xad,
			0xcd, 0x11, 0x89, 0x88, 0x04, 0x00, 0x82, 0x08,
			0x41, 0x08, 0x10, 0x00, 0x81, 0x81, 0x04, 0x0f,
			0xdf, 0xf8, 0x42, 0x10, 0x84, 0x30, 0x87, 0x0c,
			0x78, 0x43, 0xe1, 0x0f, 0xc2, 0x1f, 0xc2, 0x9f,
			0x85, 0x3e, 0x0f, 0xbf, 0xff, 0x29, 0x49, 0x93,
			0x27, 0x55, 0xe9, 0x74, 0x97, 0x69, 0x25, 0xe4,
			0x10, 0x22, 0x26, 0x41, 0x83, 0x06, 0x0d, 0x1a,
			0x30, 0x16, 0x01, 0x02, 0x04, 0x18, 0x10, 0x01,
			0xf0, 0x3f, 0x83, 0xef, 0xa2, 0x12, 0x46, 0x89,
			0xc9, 0xe4, 0xab, 0x92, 0xb7, 0x25, 0x2b, 0x73,
			0x44, 0x62, 0x62, 0x01, 0x00, 0x20, 0x82, 0x10,
			0x42, 0x04, 0x00, 0x20, 0x60, 0x41, 0x03, 0xf7,
			0xfd, 0x3c, 0xb8, 0x9a, 0x81, 0x01, 0x01, 0x02,
			0x00, 0x00, 0x03, 0x00, 0xc8, 0x00, 0x00, 0x17,
			0x70, 0xe0, 0x0b, 0xbc, 0xae, 0x00, 0x03, 0xe8,
			0x00, 0x00, 0x03, 0x01, 0xf4, 0x00, 0x00, 0x03,
			0x00, 0x7d, 0x00, 0x00, 0x03, 0x00, 0x3e, 0x80,
			0x05, 0x70, 0x80, 0x41,
		},
		SPS{
			TemporalIDNestingFlag: true,
			ProfileTierLevel: SPS_ProfileTierLevel{
				GeneralProfileIdc: 1,
				GeneralProfileCompatibilityFlag: [32]bool{
					false, true, true, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
					false, false, false, false, false, false, false, false,
				},
				GeneralProgressiveSourceFlag:   true,
				GeneralFrameOnlyConstraintFlag: true,
				GeneralLevelIdc:                150,
			},
			ChromaFormatIdc:                      1,
			PicWidthInLumaSamples:                2048,
			PicHeightInLumaSamples:               1536,
			Log2MaxPicOrderCntLsbMinus4:          12,
			MaxDecPicBufferingMinus1:             []uint32{1},
			MaxNumReorderPics:                    []uint32{0},
			MaxLatencyIncreasePlus1:              []uint32{0},
			Log2DiffMaxMinLumaCodingBlockSize:    2,
			Log2DiffMaxMinLumaTransformBlockSize: 3,
			MaxTransformHierarchyDepthInter:      3,
			MaxTransformHierarchyDepthIntra:      3,
			ScalingListEnabledFlag:               true,
			ScalingListData: &SPS_ScalingListData{
				ScalingListPredModeFlag: [4][6]bool{
					{true, true, false, true, true, false},
					{true, true, false, true, true, false},
					{true, true, false, true, true, false},
					{true, false, false, true, false, false},
				},
				ScalingListPredmatrixIDDelta: [4][6]uint32{
					{0x0, 0x0, 0x1, 0x0, 0x0, 0x1},
					{0x0, 0x0, 0x1, 0x0, 0x0, 0x1},
					{0x0, 0x0, 0x1, 0x0, 0x0, 0x1},
					{0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
				},
				ScalingListDcCoefMinus8: [4][6]int32{
					{-2, -2, 0, 1, 8, 0},
					{-2, 0, 0, 1, 0, 0},
					{0, 0, 0, 0, 0, 0},
					{0, 0, 0, 0, 0, 0},
				},
			},
			SampleAdaptiveOffsetEnabledFlag: true,
			ShortTermRefPicSets: []*SPS_ShortTermRefPicSet{
				{},
				{
					NumNegativePics:     1,
					DeltaPocS0:          []int32{-1},
					UsedByCurrPicS0Flag: []bool{true},
				},
			},
			VUI: &SPS_VUI{
				VideoSignalTypePresentFlag:   true,
				VideoFormat:                  5,
				ColourDescriptionPresentFlag: true,
				ColourPrimaries:              2,
				TransferCharacteristics:      2,
				MatrixCoefficients:           2,
				TimingInfo: &SPS_TimingInfo{
					NumUnitsInTick: 100,
					TimeScale:      3000,
				},
			},
		},
		2048,
		1536,
		30,
	},
}

func TestSPSUnmarshal(t *testing.T) {
	for _, ca := range casesSPS {
		t.Run(ca.name, func(t *testing.T) {
			var sps SPS
			err := sps.Unmarshal(ca.byts)
			require.NoError(t, err)
			require.Equal(t, ca.sps, sps)
			require.Equal(t, ca.width, sps.Width())
			require.Equal(t, ca.height, sps.Height())
			require.Equal(t, ca.fps, sps.FPS())
		})
	}
}

func FuzzSPSUnmarshal(f *testing.F) {
	for _, ca := range casesSPS {
		f.Add(ca.byts)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var sps SPS
		err := sps.Unmarshal(b)
		if err != nil {
			return
		}

		sps.Width()
		sps.Height()
		sps.FPS()
	})
}
