package opus_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/opus"
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
			require.Equal(t, ca.duration, opus.PacketDuration(ca.byts))
		})
	}
}

func FuzzPacketDuration(f *testing.F) {
	for _, ca := range casesPacketDuration {
		f.Add(ca.byts)
	}

	f.Fuzz(func(_ *testing.T, b []byte) {
		opus.PacketDuration(b)
	})
}
