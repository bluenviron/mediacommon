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
			name: "truncated recovery point",
			nalu: []byte{byte(NALUTypeSEI), 0x06, 0x01, 0xC4, 0x80},
			out:  true,
		},
		{
			name: "recovery point, recovery_frame_cnt is 0",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x06, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x80, // rbsp trailing bits
			},
			out: true,
		},
		{
			// same as above, but recovery_frame_cnt = 1
			name: "recovery point, recovery_frame_cnt is not 0",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x06, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
				0x80, // rbsp trailing bits
			},
			out: false,
		},
		{
			// recovery_frame_cnt_msb = 1, so the full value is 32
			name: "recovery point, recovery_frame_cnt is not 0 (msb)",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x06, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x20,
				0x80, // rbsp trailing bits
			},
			out: false,
		},
		{
			// cpb_cnt_flag = 1
			name: "recovery point, cpb_cnt_flag set",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x06, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE0, 0x00,
				0x80, // rbsp trailing bits
			},
			out: true,
		},
		{
			// clock_timestamp_flag = 0
			name: "recovery point, time_code without clock timestamp",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x06, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00,
				0x80, // rbsp trailing bits
			},
			out: true,
		},
		{
			// full clock timestamp
			name: "recovery point, time_code with full clock timestamp",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x06, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x60, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04,
			},
			out: true,
		},
		{
			// same as above, but recovery_frame_cnt = 1
			name: "recovery point, time_code with full clock timestamp, recovery_frame_cnt is not 0",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x06, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x60, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14,
			},
			out: false,
		},
		{
			// the parser must keep scanning after a non-matching recovery point
			name: "truncated recovery point after a non-matching one",
			nalu: []byte{
				byte(NALUTypeSEI),
				0x06, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, // recovery point, recovery_frame_cnt = 1
				0x06, 0x01, 0xC4, // truncated recovery point
				0x80, // rbsp trailing bits
			},
			out: true,
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
