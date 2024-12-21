package av1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesSequenceHeader = []struct {
	name   string
	byts   []byte
	sh     SequenceHeader
	width  int
	height int
}{
	{
		"chrome webrtc",
		[]byte{
			8, 0, 0, 0, 66, 167, 191, 228, 96, 13, 0, 64,
		},
		SequenceHeader{
			OperatingPointIdc:              []uint16{0},
			SeqLevelIdx:                    []uint8{8},
			SeqTier:                        []bool{false},
			DecoderModelPresentForThisOp:   []bool{false},
			InitialDisplayPresentForThisOp: []bool{false},
			InitialDisplayDelayMinus1:      []uint8{0},
			MaxFrameWidthMinus1:            1919,
			MaxFrameHeightMinus1:           803,
			SeqChooseScreenContentTools:    true,
			SeqForceScreenContentTools:     2,
			SeqChooseIntegerMv:             true,
			SeqForceIntegerMv:              2,
			EnableCdef:                     true,
			ColorConfig: SequenceHeader_ColorConfig{
				BitDepth:                8,
				ColorPrimaries:          2,
				TransferCharacteristics: 2,
				MatrixCoefficients:      2,
				SubsamplingX:            true,
				SubsamplingY:            true,
			},
		},
		1920,
		804,
	},
	{
		"av1 sample",
		[]byte{
			10, 11, 0, 0, 0, 66, 167, 191, 230, 46, 223, 200, 66,
		},
		SequenceHeader{
			OperatingPointIdc:              []uint16{0},
			SeqLevelIdx:                    []uint8{8},
			SeqTier:                        []bool{false},
			DecoderModelPresentForThisOp:   []bool{false},
			InitialDisplayPresentForThisOp: []bool{false},
			InitialDisplayDelayMinus1:      []uint8{0},
			MaxFrameWidthMinus1:            1919,
			MaxFrameHeightMinus1:           817,
			Use128x128Superblock:           true,
			EnableFilterIntra:              true,
			EnableIntraEdgeFilter:          true,
			EnableMaskedCompound:           true,
			EnableWarpedMotion:             true,
			EnableOrderHint:                true,
			EnableJntComp:                  true,
			EnableRefFrameMvs:              true,
			SeqChooseScreenContentTools:    true,
			SeqForceScreenContentTools:     2,
			SeqChooseIntegerMv:             true,
			OrderHintBitsMinus1:            6,
			SeqForceIntegerMv:              2,
			EnableCdef:                     true,
			ColorConfig: SequenceHeader_ColorConfig{
				BitDepth:                8,
				ColorPrimaries:          2,
				TransferCharacteristics: 2,
				MatrixCoefficients:      2,
				ColorRange:              true,
				SubsamplingX:            true,
				SubsamplingY:            true,
			},
		},
		1920,
		818,
	},
	{
		"libsvtav1",
		[]byte{
			0x8, 0x0, 0x0, 0x0, 0x42, 0xab, 0xbf, 0xc3, 0x71, 0xab, 0xe6, 0x1,
		},
		SequenceHeader{
			OperatingPointIdc:              []uint16{0},
			SeqLevelIdx:                    []uint8{8},
			SeqTier:                        []bool{false},
			DecoderModelPresentForThisOp:   []bool{false},
			InitialDisplayPresentForThisOp: []bool{false},
			InitialDisplayDelayMinus1:      []uint8{0},
			MaxFrameWidthMinus1:            1919,
			MaxFrameHeightMinus1:           1079,
			EnableIntraEdgeFilter:          true,
			EnableInterintraCompound:       true,
			EnableWarpedMotion:             true,
			EnableOrderHint:                true,
			EnableRefFrameMvs:              true,
			SeqChooseScreenContentTools:    true,
			SeqForceScreenContentTools:     2,
			SeqChooseIntegerMv:             true,
			SeqForceIntegerMv:              2,
			OrderHintBitsMinus1:            6,
			EnableCdef:                     true,
			EnableRestoration:              true,
			ColorConfig: SequenceHeader_ColorConfig{
				BitDepth:                8,
				ColorPrimaries:          2,
				TransferCharacteristics: 2,
				MatrixCoefficients:      2,
				SubsamplingX:            true,
				SubsamplingY:            true,
			},
		},
		1920,
		1080,
	},
	{
		"amd hardware av1",
		[]byte{
			0x08, 0x04, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00,
			0x00, 0xf3, 0x00, 0x00, 0x0e, 0x55, 0x77, 0xf8,
			0x73, 0xd0, 0x02, 0x7d, 0x10, 0x10, 0x10, 0x10,
			0x40,
		},
		SequenceHeader{
			TimingInfo: &SequenceHeader_TimingInfo{
				NumUnitsInDisplayTick: 1,
				TimeScale:             60,
				EqualPictureInterval:  true,
			},
			OperatingPointIdc:              []uint16{0},
			SeqLevelIdx:                    []uint8{0x0e},
			SeqTier:                        []bool{false},
			DecoderModelPresentForThisOp:   []bool{false},
			InitialDisplayPresentForThisOp: []bool{false},
			InitialDisplayDelayMinus1:      []uint8{0},
			MaxFrameWidthMinus1:            1919,
			MaxFrameHeightMinus1:           1081,
			FrameIDNumbersPresentFlag:      true,
			DeltaFrameIDLengthMinus2:       13,
			EnableOrderHint:                true,
			SeqChooseScreenContentTools:    true,
			SeqForceScreenContentTools:     SequenceHeader_SeqForceScreenContentTools(2),
			SeqChooseIntegerMv:             true,
			SeqForceIntegerMv:              2,
			OrderHintBitsMinus1:            7,
			EnableCdef:                     true,
			ColorConfig: SequenceHeader_ColorConfig{
				BitDepth:                    8,
				ColorDescriptionPresentFlag: true,
				ColorPrimaries:              1,
				TransferCharacteristics:     SequenceHeader_TransferCharacteristics(1),
				MatrixCoefficients:          SequenceHeader_MatrixCoefficients(1),
				SubsamplingX:                true,
				SubsamplingY:                true,
			},
		},
		1920,
		1082,
	},
}

func TestSequenceHeaderUnmarshal(t *testing.T) {
	for _, ca := range casesSequenceHeader {
		t.Run(ca.name, func(t *testing.T) {
			var sh SequenceHeader
			err := sh.Unmarshal(ca.byts)
			require.NoError(t, err)
			require.Equal(t, ca.sh, sh)
			require.Equal(t, ca.width, sh.Width())
			require.Equal(t, ca.height, sh.Height())
		})
	}
}

func FuzzSequenceHeaderUnmarshal(f *testing.F) {
	for _, ca := range casesSequenceHeader {
		f.Add(ca.byts)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		var sh SequenceHeader
		err := sh.Unmarshal(b)
		if err == nil {
			sh.Width()
			sh.Height()
		}
	})
}
