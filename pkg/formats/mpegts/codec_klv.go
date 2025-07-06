package mpegts

import (
	"github.com/asticode/go-astits"
)

const (
	klvaIdentifier = 'K'<<24 | 'L'<<16 | 'V'<<8 | 'A'
)

// MISB ST 1402, Table 4
const (
	metadataApplicationFormatGeneral            = 0x0100
	metadataApplicationFormatGeographicMetadata = 0x0101
	metadataApplicationFormatAnnotationMetadata = 0x0102
	metadataApplicationFormatStillImageOnDemand = 0x0103
)

// CodecKLV is a KLV codec.
// Specification: MISB ST 1402
type CodecKLV struct {
	Synchronous bool
}

// IsVideo implements Codec.
func (CodecKLV) IsVideo() bool {
	return false
}

func (*CodecKLV) isCodec() {}

func (c CodecKLV) marshal(pid uint16) (*astits.PMTElementaryStream, error) {
	if c.Synchronous {
		metadataDesc, err := metadataDescriptor{
			MetadataApplicationFormat: metadataApplicationFormatGeneral,
			MetadataFormat:            0xFF,
			MetadataFormatIdentifier:  klvaIdentifier,
			MetadataServiceID:         0x00,
			DecoderConfigFlags:        0,
			DSMCCFlag:                 false,
		}.marshal()
		if err != nil {
			return nil, err
		}

		metadataSTDDesc, err := metadataSTDDescriptor{
			MetadataInputLeakRate:  0,
			MetadataBufferSize:     0,
			MetadataOutputLeakRate: 0,
		}.marshal()
		if err != nil {
			return nil, err
		}

		return &astits.PMTElementaryStream{
			ElementaryPID: pid,
			StreamType:    astits.StreamTypeMetadata,
			ElementaryStreamDescriptors: []*astits.Descriptor{
				{
					Length: uint8(len(metadataDesc)),
					Tag:    descriptorTagMetadata,
					Unknown: &astits.DescriptorUnknown{
						Tag:     descriptorTagMetadata,
						Content: metadataDesc,
					},
				},
				{
					Length: uint8(len(metadataSTDDesc)),
					Tag:    descriptorTagMetadataSTD,
					Unknown: &astits.DescriptorUnknown{
						Tag:     descriptorTagMetadataSTD,
						Content: metadataSTDDesc,
					},
				},
			},
		}, nil
	}

	return &astits.PMTElementaryStream{
		ElementaryPID: pid,
		StreamType:    astits.StreamTypePrivateData,
		ElementaryStreamDescriptors: []*astits.Descriptor{
			{
				Length: 4,
				Tag:    astits.DescriptorTagRegistration,
				Registration: &astits.DescriptorRegistration{
					FormatIdentifier: klvaIdentifier,
				},
			},
		},
	}, nil
}
