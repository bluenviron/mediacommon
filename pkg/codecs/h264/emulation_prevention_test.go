package h264

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var casesEmulationPreventionRemove = []struct {
	name   string
	unproc []byte
	proc   []byte
}{
	{
		"base",
		[]byte{
			0x00, 0x00, 0x00,
			0x00, 0x00, 0x01,
			0x00, 0x00, 0x02,
			0x00, 0x00, 0x03,
		},
		[]byte{
			0x00, 0x00, 0x03, 0x00,
			0x00, 0x00, 0x03, 0x01,
			0x00, 0x00, 0x03, 0x02,
			0x00, 0x00, 0x03, 0x03,
		},
	},
	{
		"double emulation byte",
		[]byte{
			0x00, 0x00, 0x00,
			0x00, 0x00,
		},
		[]byte{
			0x00, 0x00, 0x03,
			0x00, 0x00, 0x03, 0x00,
		},
	},
	{
		"terminal emulation byte",
		[]byte{
			0x00, 0x00,
		},
		[]byte{
			0x00, 0x00, 0x03,
		},
	},
}

func TestEmulationPreventionRemove(t *testing.T) {
	for _, ca := range casesEmulationPreventionRemove {
		t.Run(ca.name, func(t *testing.T) {
			unproc := EmulationPreventionRemove(ca.proc)
			require.Equal(t, ca.unproc, unproc)
		})
	}
}

func FuzzEmulationPreventionRemove(f *testing.F) {
	for _, ca := range casesEmulationPreventionRemove {
		f.Add(ca.proc)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		EmulationPreventionRemove(b)
	})
}
