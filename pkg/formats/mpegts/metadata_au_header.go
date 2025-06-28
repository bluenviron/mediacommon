package mpegts

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/bits"
)

type metadataAuCellHeader struct {
	PayloadSize                 int
	MetadataServiceID           uint8
	CellFragmentationIndication uint8
	DecoderConfigFlag           bool
	RandomAccessIndicator       bool
	SequenceNumber              uint8
}

func (h *metadataAuCellHeader) unmarshal(buf []byte) (int, error) {
	pos := 0

	err := bits.HasSpace(buf, pos, 40)
	if err != nil {
		return 0, err
	}

	h.MetadataServiceID = uint8(bits.ReadBitsUnsafe(buf, &pos, 8))
	if h.MetadataServiceID != 0xFC {
		return 0, fmt.Errorf("invalid prefix: %v", h.MetadataServiceID)
	}
	h.SequenceNumber = uint8(bits.ReadBitsUnsafe(buf, &pos, 8))
	h.CellFragmentationIndication = uint8(bits.ReadBitsUnsafe(buf, &pos, 2))

	h.DecoderConfigFlag = bits.ReadFlagUnsafe(buf, &pos)
	h.RandomAccessIndicator = bits.ReadFlagUnsafe(buf, &pos)

	pos += 4 // reserved

	h.PayloadSize = int(bits.ReadBitsUnsafe(buf, &pos, 16))

	return pos / 8, nil
}

func (h *metadataAuCellHeader) marshalTo(buf []byte) (int, error) {
	buf[0] = h.MetadataServiceID
	buf[1] = h.SequenceNumber
	buf[2] = 0
	buf[2] |= h.CellFragmentationIndication << 6
	if h.DecoderConfigFlag {
		buf[2] |= 1 << 5
	}
	if h.RandomAccessIndicator {
		buf[2] |= 1 << 4
	}
	pos := 24
	bits.WriteBitsUnsafe(buf, &pos, uint64(h.PayloadSize), 16)

	return 5, nil
}
