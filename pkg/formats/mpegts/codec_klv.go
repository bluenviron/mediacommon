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
					// Length must be different than zero.
					// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
					Length: 1,
					Tag:    descriptorTagMetadata,
					Unknown: &astits.DescriptorUnknown{
						Content: metadataDesc,
					},
				},
				{
					// Length must be different than zero.
					// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
					Length: 1,
					Tag:    descriptorTagMetadataSTD,
					Unknown: &astits.DescriptorUnknown{
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
				// Length must be different than zero.
				// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
				Length: 1,
				Tag:    astits.DescriptorTagRegistration,
				Registration: &astits.DescriptorRegistration{
					FormatIdentifier: klvaIdentifier,
				},
			},
		},
	}, nil
}
