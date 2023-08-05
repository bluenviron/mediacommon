package av1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSequenceHeaderUnmarshal(t *testing.T) {
	for _, ca := range []struct {
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
				EnableSuperRes:                 true,
				SeqForceIntegerMv:              2,
				EnableCdef:                     true,
				ColorConfig: SequenceHeader_ColorConfig{
					BitDepth:                8,
					MonoChrome:              true,
					ColorPrimaries:          2,
					TransferCharacteristics: 2,
					MatrixCoefficients:      2,
					SubsamplingX:            true,
					SubsamplingY:            true,
				},
			},
			1920,
			818,
		},
	} {
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
	f.Fuzz(func(t *testing.T, b []byte) {
		var sh SequenceHeader
		sh.Unmarshal(b)
	})
}
