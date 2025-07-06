package mpegts

import "fmt"

// metadataAUCell is a metadata_AU_cell.
// Specification: ISO 13818-1, table 2-97
type metadataAUCell struct {
	MetadataServiceID      uint8
	SequenceNumber         uint8
	CellFragmentIndication uint8
	DecoderConfigFlag      bool
	RandomAccessIndicator  bool
	AUCellData             []byte
}

func (c *metadataAUCell) unmarshal(buf []byte) (int, error) {
	if len(buf) < 5 {
		return 0, fmt.Errorf("buffer is too small")
	}

	n := 0

	c.MetadataServiceID = buf[n]
	n++
	c.SequenceNumber = buf[n]
	n++
	c.CellFragmentIndication = buf[n] >> 6
	c.DecoderConfigFlag = ((buf[n] >> 5) & 0b1) != 0
	c.RandomAccessIndicator = ((buf[n] >> 4) & 0b1) != 0
	n++

	auCellDataLength := int(uint16(buf[n])<<8 | uint16(buf[n+1]))
	n += 2

	if len(buf[n:]) < auCellDataLength {
		return 0, fmt.Errorf("buffer is too small")
	}

	c.AUCellData = buf[n : n+auCellDataLength]
	n += auCellDataLength

	return n, nil
}

func (c metadataAUCell) marshalSize() int {
	return 5 + len(c.AUCellData)
}

func (c metadataAUCell) marshal() ([]byte, error) {
	buf := make([]byte, c.marshalSize())
	_, err := c.marshalTo(buf)
	return buf, err
}

func (c metadataAUCell) marshalTo(buf []byte) (int, error) {
	n := 0

	buf[n] = c.MetadataServiceID
	n++
	buf[n] = c.SequenceNumber
	n++
	buf[n] = c.CellFragmentIndication<<6 | flagToByte(c.DecoderConfigFlag)<<5 |
		flagToByte(c.RandomAccessIndicator)<<4 | 0b1111
	n++

	auCellDataLength := len(c.AUCellData)
	buf[n] = byte(auCellDataLength >> 8)
	buf[n+1] = byte(auCellDataLength)
	n += 2

	n += copy(buf[n:], c.AUCellData)

	return n, nil
}
