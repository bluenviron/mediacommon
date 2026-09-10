package h264

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSEIRecoveryPoint(t *testing.T) {
	cases := []struct {
		name string
		nalu []byte
		out  bool
	}{
		{
			name: "not a SEI NALU",
			nalu: []byte{byte(NALUTypeNonIDR), 0x06},
			out:  false,
		},
		{
			name: "single recovery point",
			nalu: []byte{byte(NALUTypeSEI), 0x06, 0x01, 0x00, 0x80},
			out:  true,
		},
		{
			name: "recovery point after another SEI message",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x05, 0x02, 0xAA, 0xBB, // user_data_unregistered, size 2
				0x06, 0x01, 0x00, // recovery point
				0x80, // rbsp trailing bits
			},
			out: true,
		},
		{
			name: "payload type with 0xff continuation",
			nalu: []byte{
				byte(NALUTypeSEI),
				0xFF, 0x05, 0x02, 0xAA, 0xBB, // payload type 260, size 2
				0x06, 0x01, 0x00, // recovery point
				0x80, // rbsp trailing bits
			},
			out: true,
		},
		{
			name: "payload size with 0xff continuation",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x05, 0xFF, 0x01, 0xaa, 0xbb, // payload type 5, size 256
				0x80, // rbsp trailing bits
			},
			out: false,
		},
		{
			name: "no recovery point",
			nalu: []byte{byte(NALUTypeSEI), 0x05, 0x02, 0xAA, 0xBB, 0x80},
			out:  false,
		},
		{
			name: "only NALU header",
			nalu: []byte{byte(NALUTypeSEI)},
			out:  false,
		},
	}

	for _, ca := range cases {
		t.Run(ca.name, func(t *testing.T) {
			require.Equal(t, ca.out, isSEIRecoveryPoint(ca.nalu))
		})
	}
}
