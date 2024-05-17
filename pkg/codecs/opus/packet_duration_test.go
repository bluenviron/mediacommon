package opus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var casesPacketDuration = []struct {
	name     string
	byts     []byte
	duration time.Duration
}{
	{
		"aa",
		[]byte{1},
		20 * time.Millisecond,
	},
}

func TestPacketDuration(t *testing.T) {
	for _, ca := range casesPacketDuration {
		t.Run(ca.name, func(t *testing.T) {
			require.Equal(t, ca.duration, PacketDuration(ca.byts))
		})
	}
}

func FuzzPacketDuration(f *testing.F) {
	for _, ca := range casesPacketDuration {
		f.Add(ca.byts)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		PacketDuration(b)
	})
}
