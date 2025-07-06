package mpegts

import "fmt"

// ISO 13818-1, table 2-45
const (
	descriptorTagMetadataSTD = 0x27
)

// metadataSTDDescriptor is a metadata_std_descriptor.
// Specification: ISO 13818-1, table 2-88
type metadataSTDDescriptor struct {
	MetadataInputLeakRate  uint32
	MetadataBufferSize     uint32
	MetadataOutputLeakRate uint32
}

func (d *metadataSTDDescriptor) unmarshal(buf []byte) error {
	if len(buf) < 9 {
		return fmt.Errorf("buffer is too small")
	}

	n := 0

	d.MetadataInputLeakRate = uint32(buf[n]&0b00111111)<<16 | uint32(buf[n+1])<<8 | uint32(buf[n+2])
	n += 3

	d.MetadataBufferSize = uint32(buf[n]&0b00111111)<<16 | uint32(buf[n+1])<<8 | uint32(buf[n+2])
	n += 3

	d.MetadataOutputLeakRate = uint32(buf[n]&0b00111111)<<16 | uint32(buf[n+1])<<8 | uint32(buf[n+2])
	n += 3

	if len(buf[n:]) != 0 {
		return fmt.Errorf("unread bytes detected")
	}

	return nil
}

func (d metadataSTDDescriptor) marshalSize() int {
	return 9
}

func (d metadataSTDDescriptor) marshal() ([]byte, error) {
	buf := make([]byte, d.marshalSize())
	n := 0

	buf[n] = 0b11000000 | byte(d.MetadataInputLeakRate>>16)
	buf[n+1] = byte(d.MetadataInputLeakRate >> 8)
	buf[n+2] = byte(d.MetadataInputLeakRate)
	n += 3

	buf[n] = 0b11000000 | byte(d.MetadataBufferSize>>16)
	buf[n+1] = byte(d.MetadataBufferSize >> 8)
	buf[n+2] = byte(d.MetadataBufferSize)
	n += 3

	buf[n] = 0b11000000 | byte(d.MetadataOutputLeakRate>>16)
	buf[n+1] = byte(d.MetadataOutputLeakRate >> 8)
	buf[n+2] = byte(d.MetadataOutputLeakRate)

	return buf, nil
}
