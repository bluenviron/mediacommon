package mpegts

import "fmt"

// ISO 13818-1, table 2-45
const (
	descriptorTagMetadata = 0x26
)

func flagToByte(v bool) byte {
	if v {
		return 1
	}
	return 0
}

// metadataDescriptor is a metadata_descriptor.
// Specification: ISO 13818-1, table 2-86
type metadataDescriptor struct {
	MetadataApplicationFormat uint16

	// metadata_application_format == 0xFFFF
	MetadataApplicationFormatIdentifier uint32

	MetadataFormat uint8

	// metadata_format == 0xFF
	MetadataFormatIdentifier uint32

	MetadataServiceID  uint8
	DecoderConfigFlags uint8
	DSMCCFlag          bool

	// DSM-CC_flag == 1
	ServiceIdentification []byte

	// decoder_config_flags == '001'
	DecoderConfig []byte

	// decoder_config_flags == '011'
	DecConfigIdentification []byte

	// decoder_config_flags == '100'
	DecoderConfigMetadataServiceID uint8

	PrivateData []uint8
}

func (d *metadataDescriptor) unmarshal(buf []byte) error {
	n := 0

	if len(buf[n:]) < 2 {
		return fmt.Errorf("buffer too short")
	}

	d.MetadataApplicationFormat = uint16(buf[n])<<8 | uint16(buf[n+1])
	n += 2

	if d.MetadataApplicationFormat == 0xFFFF {
		if len(buf[n:]) < 4 {
			return fmt.Errorf("buffer too short")
		}

		d.MetadataApplicationFormatIdentifier = uint32(buf[n])<<24 | uint32(buf[n+1])<<16 |
			uint32(buf[n+2])<<8 | uint32(buf[n+3])
		n += 4
	}

	if len(buf[n:]) < 1 {
		return fmt.Errorf("buffer too short")
	}

	d.MetadataFormat = buf[n]
	n++

	if d.MetadataFormat == 0xFF {
		if len(buf[n:]) < 4 {
			return fmt.Errorf("buffer too short")
		}

		d.MetadataFormatIdentifier = uint32(buf[n])<<24 | uint32(buf[n+1])<<16 |
			uint32(buf[n+2])<<8 | uint32(buf[n+3])
		n += 4
	}

	if len(buf[n:]) < 2 {
		return fmt.Errorf("buffer too short")
	}

	d.MetadataServiceID = buf[n]
	n++
	d.DecoderConfigFlags = buf[n] >> 5
	d.DSMCCFlag = ((buf[n] >> 4) & 0b1) != 0
	n++

	if d.DSMCCFlag {
		if len(buf[n:]) < 1 {
			return fmt.Errorf("buffer too short")
		}

		le := int(buf[n])
		n++

		if len(buf[n:]) < le {
			return fmt.Errorf("buffer too short")
		}

		d.ServiceIdentification = buf[n : n+le]
		n += le
	}

	switch d.DecoderConfigFlags {
	case 0b001:
		if len(buf[n:]) < 1 {
			return fmt.Errorf("buffer too short")
		}

		le := int(buf[n])
		n++

		if len(buf[n:]) < le {
			return fmt.Errorf("buffer too short")
		}

		d.DecoderConfig = buf[n : n+le]
		n += le

	case 0b011:
		if len(buf[n:]) < 1 {
			return fmt.Errorf("buffer too short")
		}

		le := int(buf[n])
		n++

		if len(buf[n:]) < le {
			return fmt.Errorf("buffer too short")
		}

		d.DecConfigIdentification = buf[n : n+le]
		n += le

	case 0b100:
		if len(buf[n:]) < 1 {
			return fmt.Errorf("buffer too short")
		}

		d.DecoderConfigMetadataServiceID = buf[n]
		n++

	case 0b101, 0b110:
		return fmt.Errorf("DecoderConfigFlags %v is unsupported", d.DecoderConfigFlags)
	}

	d.PrivateData = buf[n:]

	return nil
}

func (d metadataDescriptor) marshalSize() int {
	v := 5

	if d.MetadataApplicationFormat == 0xFFFF {
		v += 4
	}

	if d.MetadataFormat == 0xFF {
		v += 4
	}

	if d.DSMCCFlag {
		v += 1 + len(d.ServiceIdentification)
	}

	switch d.DecoderConfigFlags {
	case 0b001:
		v += 1 + len(d.DecoderConfig)

	case 0b011:
		v += 1 + len(d.DecConfigIdentification)

	case 0b100:
		v++

	case 0b101, 0b110:
		// unsupported
	}

	v += len(d.PrivateData)

	return v
}

func (d metadataDescriptor) marshal() ([]byte, error) {
	buf := make([]byte, d.marshalSize())
	n := 0

	buf[n] = byte(d.MetadataApplicationFormat >> 8)
	buf[n+1] = byte(d.MetadataApplicationFormat)
	n += 2

	if d.MetadataApplicationFormat == 0xFFFF {
		buf[n] = byte(d.MetadataApplicationFormatIdentifier >> 24)
		buf[n+1] = byte(d.MetadataApplicationFormatIdentifier >> 16)
		buf[n+2] = byte(d.MetadataApplicationFormatIdentifier >> 8)
		buf[n+3] = byte(d.MetadataApplicationFormatIdentifier)
		n += 4
	}

	buf[n] = d.MetadataFormat
	n++

	if d.MetadataFormat == 0xFF {
		buf[n] = byte(d.MetadataFormatIdentifier >> 24)
		buf[n+1] = byte(d.MetadataFormatIdentifier >> 16)
		buf[n+2] = byte(d.MetadataFormatIdentifier >> 8)
		buf[n+3] = byte(d.MetadataFormatIdentifier)
		n += 4
	}

	buf[n] = d.MetadataServiceID
	n++
	buf[n] = d.DecoderConfigFlags<<5 | flagToByte(d.DSMCCFlag)<<4 | 0b1111
	n++

	if d.DSMCCFlag {
		buf[n] = uint8(len(d.ServiceIdentification))
		n++
		n += copy(buf[n:], d.ServiceIdentification)
	}

	switch d.DecoderConfigFlags {
	case 0b001:
		buf[n] = uint8(len(d.DecoderConfig))
		n++
		n += copy(buf[n:], d.DecoderConfig)

	case 0b011:
		buf[n] = uint8(len(d.DecConfigIdentification))
		n++
		n += copy(buf[n:], d.DecConfigIdentification)

	case 0b100:
		buf[n] = d.DecoderConfigMetadataServiceID
		n++

	case 0b101, 0b110:
		// unsupported
	}

	copy(buf[n:], d.PrivateData)

	return buf, nil
}
