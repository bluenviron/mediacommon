package mpegts

import (
	"fmt"
)

// maximum size of a private section (section_length is limited to 4093)
const maxPrivateSectionSize = 4096

func sectionSize(b []byte) int {
	return 3 + (int(b[1]&0x0F)<<8 | int(b[2]))
}

// sectionAssembler reassembles sections from MPEG-TS packets of a single PID.
// astits is not used since it doesn't support private sections
// and it returns a payload only when the next one starts.
type sectionAssembler struct {
	buf    []byte
	lastCC int
}

func (a *sectionAssembler) initialize() {
	a.lastCC = -1
}

// push processes a raw 188-byte packet and returns completed sections.
func (a *sectionAssembler) push(pkt []byte) ([][]byte, error) {
	transportErrorIndicator := (pkt[1] & 0x80) != 0
	payloadUnitStartIndicator := (pkt[1] & 0x40) != 0
	hasAdaptationField := (pkt[3] & 0x20) != 0
	hasPayload := (pkt[3] & 0x10) != 0
	cc := int(pkt[3] & 0x0F)

	if transportErrorIndicator {
		a.buf = nil
		return nil, fmt.Errorf("transport error indicator is set")
	}

	if !hasPayload {
		return nil, nil
	}

	var err error

	if a.lastCC >= 0 {
		// duplicate packet
		if cc == a.lastCC {
			return nil, nil
		}

		if cc != ((a.lastCC + 1) & 0x0F) {
			if a.buf != nil {
				err = fmt.Errorf("section discarded because of packet loss")
			}
			a.buf = nil
		}
	}
	a.lastCC = cc

	pos := 4
	if hasAdaptationField {
		pos += 1 + int(pkt[4])
	}
	if pos >= len(pkt) {
		return nil, err
	}
	payload := pkt[pos:]

	var out [][]byte

	if !payloadUnitStartIndicator {
		if a.buf == nil {
			return nil, err
		}

		a.buf = append(a.buf, payload...)
		return a.flushBuffer(out)
	}

	pointerField := int(payload[0])
	payload = payload[1:]

	if pointerField > len(payload) {
		a.buf = nil
		return nil, fmt.Errorf("invalid pointer_field")
	}

	// bytes before pointer_field complete the previous section
	if a.buf != nil {
		a.buf = append(a.buf, payload[:pointerField]...)
		out, err = a.flushBuffer(out)
		if a.buf != nil {
			a.buf = nil
			err = fmt.Errorf("section discarded because it is truncated")
		}
	}

	payload = payload[pointerField:]

	// multiple sections can start in the same packet. 0xFF is stuffing.
	for len(payload) != 0 && payload[0] != 0xFF {
		if len(payload) < 3 {
			a.buf = append([]byte(nil), payload...)
			break
		}

		size := sectionSize(payload)

		if len(payload) < size {
			a.buf = append([]byte(nil), payload...)
			break
		}

		out = append(out, append([]byte(nil), payload[:size]...))
		payload = payload[size:]
	}

	return out, err
}

func (a *sectionAssembler) flushBuffer(out [][]byte) ([][]byte, error) {
	if len(a.buf) < 3 {
		return out, nil
	}

	size := sectionSize(a.buf)

	if size > maxPrivateSectionSize {
		a.buf = nil
		return out, fmt.Errorf("section is too big")
	}

	if len(a.buf) < size {
		return out, nil
	}

	out = append(out, a.buf[:size])
	a.buf = nil
	return out, nil
}
