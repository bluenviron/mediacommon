package mpegts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTimeDecoder2NegativeDiff(t *testing.T) {
	d := NewTimeDecoder2()

	ts := d.Decode(64523434)
	require.Equal(t, int64(0), ts)

	ts = d.Decode(64523434 - 90000)
	require.Equal(t, int64(-90000), ts)

	ts = d.Decode(64523434)
	require.Equal(t, int64(0), ts)

	ts = d.Decode(64523434 + 90000*2)
	require.Equal(t, int64(2*90000), ts)

	ts = d.Decode(64523434 + 90000)
	require.Equal(t, int64(1*90000), ts)
}

func TestTimeDecoder2Overflow(t *testing.T) {
	d := NewTimeDecoder2()

	ts := d.Decode(0x1FFFFFFFF - 20)
	require.Equal(t, int64(0), ts)

	i := int64(0x1FFFFFFFF - 20)
	secs := int64(0)
	const stride = 150
	lim := int64(uint64(0x1FFFFFFFF - (stride * 90000)))

	for n := 0; n < 100; n++ {
		// overflow
		i += 90000 * stride
		secs += stride
		ts := d.Decode(i)
		require.Equal(t, secs*90000, ts)

		// reach 2^32 slowly
		secs += stride
		i += 90000 * stride
		for ; i < lim; i += 90000 * stride {
			ts = d.Decode(i)
			require.Equal(t, secs*90000, ts)
			secs += stride
		}
	}
}

func TestTimeDecoder2OverflowAndBack(t *testing.T) {
	d := NewTimeDecoder2()

	ts := d.Decode(0x1FFFFFFFF - 90000 + 1)
	require.Equal(t, int64(0), ts)

	ts = d.Decode(0x1FFFFFFFF - 90000 + 1)
	require.Equal(t, int64(0), ts)

	ts = d.Decode(90000)
	require.Equal(t, int64(2*90000), ts)

	ts = d.Decode(0x1FFFFFFFF - 90000 + 1)
	require.Equal(t, int64(0), ts)

	ts = d.Decode(0x1FFFFFFFF - 90000*2 + 1)
	require.Equal(t, int64(-1*90000), ts)

	ts = d.Decode(0x1FFFFFFFF - 90000 + 1)
	require.Equal(t, int64(0), ts)

	ts = d.Decode(90000)
	require.Equal(t, int64(2*90000), ts)
}
