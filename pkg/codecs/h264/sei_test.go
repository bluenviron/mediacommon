package h264

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var seiCases = []struct {
	name string
	nalu []byte
	out  bool
}{
	{
		name: "sample recovery point with zero count",
		nalu: []byte{0x06, 0x06, 0x01, 0xc4, 0x80},
		out:  true,
	},
	{
		name: "nonzero count",
		nalu: []byte{0x06, 0x06, 0x01, 0x51, 0x80},
	},
	{
		name: "multi-byte nonzero count",
		nalu: []byte{0x06, 0x06, 0x02, 0x24, 0x40, 0x80},
	},
	{
		name: "zero count with different flags",
		nalu: []byte{0x06, 0x06, 0x01, 0xfc, 0x80},
		out:  true,
	},
	{
		name: "missing golomb value",
		nalu: []byte{0x06, 0x06, 0x01, 0x00, 0x80},
	},
	{
		name: "missing flags",
		nalu: []byte{0x06, 0x06, 0x00, 0x80},
	},
	{
		name: "invalid payload alignment",
		nalu: []byte{0x06, 0x06, 0x01, 0xc0, 0x80},
	},
	{
		name: "extra payload data",
		nalu: []byte{0x06, 0x06, 0x02, 0xc4, 0x00, 0x80},
	},
	{
		name: "missing RBSP trailing bits",
		nalu: []byte{0x06, 0x06, 0x01, 0xc4},
	},
	{
		name: "truncated message",
		nalu: []byte{0x06, 0x06, 0x02, 0xc4},
	},
	{
		name: "truncated payload type",
		nalu: []byte{0x06, 0xff},
	},
	{
		name: "truncated payload size",
		nalu: []byte{0x06, 0x06, 0xff},
	},
	{
		name: "extended payload type before recovery point",
		nalu: []byte{0x06, 0xff, 0x05, 0x02, 0xaa, 0xbb, 0x06, 0x01, 0xc4, 0x80},
		out:  true,
	},
	{
		name: "extended payload size before recovery point",
		nalu: append(append([]byte{0x06, 0x05, 0xff, 0x00}, make([]byte, 255)...), 0x06, 0x01, 0xc4, 0x80),
		out:  true,
	},
	{
		name: "nonzero followed by zero",
		nalu: []byte{0x06, 0x06, 0x01, 0x51, 0x06, 0x01, 0xc4, 0x80},
		out:  true,
	},
	{
		name: "non-recovery message before recovery point",
		nalu: []byte{0x06, 0x05, 0x02, 0xaa, 0xbb, 0x06, 0x01, 0xc4, 0x80},
		out:  true,
	},
	{
		name: "emulation prevention before recovery point",
		nalu: []byte{0x06, 0x05, 0x04, 0x00, 0x00, 0x03, 0x01, 0x42, 0x06, 0x01, 0xc4, 0x80},
		out:  true,
	},
	{
		name: "oversized message before recovery point",
		nalu: []byte{0x06, 0x05, 0x05, 0xaa, 0xbb, 0x06, 0x01, 0xc4, 0x80},
	},
	{
		name: "oversized message after recovery point",
		nalu: []byte{0x06, 0x06, 0x01, 0xc4, 0x05, 0x04, 0x80},
	},
	{
		name: "no recovery point",
		nalu: []byte{0x06, 0x05, 0x02, 0xaa, 0xbb, 0x80},
	},
	{
		name: "empty NALU",
	},
	{
		name: "only NALU header",
		nalu: []byte{0x06},
	},
}

func TestIsSEIRecoveryPoint(t *testing.T) {
	for _, ca := range seiCases {
		t.Run(ca.name, func(t *testing.T) {
			require.Equal(t, ca.out, isSEIRecoveryPoint(ca.nalu))
		})
	}
}

func FuzzIsSEIRecoveryPoint(f *testing.F) {
	for _, ca := range seiCases {
		f.Add(ca.nalu)
	}

	f.Fuzz(func(_ *testing.T, payload []byte) {
		isSEIRecoveryPoint(payload)
	})
}
