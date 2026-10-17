package mpegts

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func sectionOfSize(size int, fill byte) []byte {
	sec := bytes.Repeat([]byte{fill}, size)
	sec[0] = 0xFC
	sec[1] = 0x30 | byte((size-3)>>8)
	sec[2] = byte(size - 3)
	return sec
}

func sectionPacket(pusi bool, cc uint8, payload []byte) []byte {
	pkt := bytes.Repeat([]byte{0xFF}, packetSize)
	pkt[0] = syncByte
	pkt[1] = 0x01 // PID 0x100
	if pusi {
		pkt[1] |= 0x40
	}
	pkt[2] = 0x00
	pkt[3] = 0x10 | cc
	copy(pkt[4:], payload)
	return pkt
}

func TestSectionAssemblerSinglePacket(t *testing.T) {
	sec := sectionOfSize(20, 0x01)

	a := &sectionAssembler{}
	a.initialize()

	out, err := a.push(sectionPacket(true, 0, append([]byte{0}, sec...)))
	require.NoError(t, err)
	require.Equal(t, [][]byte{sec}, out)
}

func TestSectionAssemblerMultiplePackets(t *testing.T) {
	sec := sectionOfSize(400, 0x02)
	payload := append([]byte{0}, sec...)

	a := &sectionAssembler{}
	a.initialize()

	out, err := a.push(sectionPacket(true, 0, payload[:184]))
	require.NoError(t, err)
	require.Empty(t, out)

	out, err = a.push(sectionPacket(false, 1, payload[184:368]))
	require.NoError(t, err)
	require.Empty(t, out)

	// section is returned as soon as it is complete
	out, err = a.push(sectionPacket(false, 2, payload[368:]))
	require.NoError(t, err)
	require.Equal(t, [][]byte{sec}, out)
}

func TestSectionAssemblerMultipleSectionsInPacket(t *testing.T) {
	sec1 := sectionOfSize(20, 0x03)
	sec2 := sectionOfSize(30, 0x04)

	a := &sectionAssembler{}
	a.initialize()

	payload := append([]byte{0}, sec1...)
	payload = append(payload, sec2...)

	out, err := a.push(sectionPacket(true, 0, payload))
	require.NoError(t, err)
	require.Equal(t, [][]byte{sec1, sec2}, out)
}

func TestSectionAssemblerPointerField(t *testing.T) {
	sec1 := sectionOfSize(200, 0x05)
	sec2 := sectionOfSize(20, 0x06)

	a := &sectionAssembler{}
	a.initialize()

	out, err := a.push(sectionPacket(true, 0, append([]byte{0}, sec1[:183]...)))
	require.NoError(t, err)
	require.Empty(t, out)

	// the tail of sec1 precedes the start of sec2
	payload := make([]byte, 0, 1+len(sec1)-183+len(sec2))
	payload = append(payload, byte(len(sec1)-183))
	payload = append(payload, sec1[183:]...)
	payload = append(payload, sec2...)

	out, err = a.push(sectionPacket(true, 1, payload))
	require.NoError(t, err)
	require.Equal(t, [][]byte{sec1, sec2}, out)
}

func TestSectionAssemblerPacketLoss(t *testing.T) {
	sec := sectionOfSize(400, 0x07)
	payload := append([]byte{0}, sec...)

	a := &sectionAssembler{}
	a.initialize()

	_, err := a.push(sectionPacket(true, 0, payload[:184]))
	require.NoError(t, err)

	out, err := a.push(sectionPacket(false, 2, payload[368:]))
	require.EqualError(t, err, "section discarded because of packet loss")
	require.Empty(t, out)

	// next section is received correctly
	sec2 := sectionOfSize(20, 0x08)
	out, err = a.push(sectionPacket(true, 3, append([]byte{0}, sec2...)))
	require.NoError(t, err)
	require.Equal(t, [][]byte{sec2}, out)
}

func TestSectionAssemblerDuplicatePacket(t *testing.T) {
	sec := sectionOfSize(20, 0x09)
	pkt := sectionPacket(true, 0, append([]byte{0}, sec...))

	a := &sectionAssembler{}
	a.initialize()

	out, err := a.push(pkt)
	require.NoError(t, err)
	require.Equal(t, [][]byte{sec}, out)

	out, err = a.push(pkt)
	require.NoError(t, err)
	require.Empty(t, out)
}
