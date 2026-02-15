package mpegts

import (
	"fmt"

	"github.com/asticode/go-astits"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
)

const (
	opusIdentifier = 'O'<<24 | 'p'<<16 | 'u'<<8 | 's'
	klvaIdentifier = 'K'<<24 | 'L'<<16 | 'V'<<8 | 'A'
)

// MISB ST 1402, Table 4
const (
	metadataApplicationFormatGeneral            = 0x0100
	metadataApplicationFormatGeographicMetadata = 0x0101
	metadataApplicationFormatAnnotationMetadata = 0x0102
	metadataApplicationFormatStillImageOnDemand = 0x0103
)

func boolToUint8(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}

func findAC3Descriptor(descriptors []*astits.Descriptor) *astits.DescriptorAC3 {
	for _, sd := range descriptors {
		if sd.Tag == astits.DescriptorTagAC3 && sd.AC3 != nil {
			return sd.AC3
		}
	}
	return nil
}

func findEAC3Descriptor(descriptors []*astits.Descriptor) *astits.DescriptorEnhancedAC3 {
	for _, sd := range descriptors {
		if sd.Tag == astits.DescriptorTagEnhancedAC3 && sd.EnhancedAC3 != nil {
			return sd.EnhancedAC3
		}
	}
	return nil
}

func findRegistrationIdentifier(descriptors []*astits.Descriptor) (uint32, bool) {
	ret := uint32(0)

	for _, sd := range descriptors {
		if sd.Registration != nil {
			// in case of multiple registrations, do not return anything
			if ret != 0 {
				return 0, false
			}
			ret = sd.Registration.FormatIdentifier
		}
	}

	if ret == 0 {
		return 0, false
	}

	return ret, true
}

func findKLVMetadataDescriptor(descriptors []*astits.Descriptor) *metadataDescriptor {
	var ret *metadataDescriptor
	for _, sd := range descriptors {
		if sd.Unknown != nil {
			if sd.Unknown.Tag == descriptorTagMetadata {
				var dm metadataDescriptor
				err := dm.unmarshal(sd.Unknown.Content)
				if err != nil {
					continue
				}

				if dm.MetadataFormatIdentifier == klvaIdentifier {
					// in case of multiple metadata, do not return anything
					if ret != nil {
						return nil
					}
					ret = &dm
				}
			}
		}
	}
	return ret
}

func findDVBSubtitlingDescriptor(descriptors []*astits.Descriptor) []*astits.DescriptorSubtitlingItem {
	for _, sd := range descriptors {
		if sd.Tag == astits.DescriptorTagSubtitling && sd.Subtitling != nil {
			return sd.Subtitling.Items
		}
	}
	return nil
}

func findOpusChannelCount(descriptors []*astits.Descriptor) int {
	for _, sd := range descriptors {
		if sd.Extension != nil && sd.Extension.Tag == 0x80 &&
			sd.Extension.Unknown != nil && len(*sd.Extension.Unknown) >= 1 {
			return int((*sd.Extension.Unknown)[0])
		}
	}
	return 0
}

func findCodec(es *astits.PMTElementaryStream) (codecs.Codec, error) {
	switch es.StreamType {
	// video

	case astits.StreamTypeH265Video:
		return &codecs.H265{}, nil

	case astits.StreamTypeH264Video:
		return &codecs.H264{}, nil

	case astits.StreamTypeMPEG4Video:
		return &codecs.MPEG4Video{}, nil

	case astits.StreamTypeMPEG2Video, astits.StreamTypeMPEG1Video:
		return &codecs.MPEG1Video{}, nil

		// audio

	case astits.StreamTypeAACAudio:
		return &codecs.MPEG4Audio{}, nil

	case astits.StreamTypeAACLATMAudio:
		return &codecs.MPEG4AudioLATM{}, nil

	case astits.StreamTypeMPEG1Audio:
		return &codecs.MPEG1Audio{}, nil

	case astits.StreamTypeAC3Audio:
		var fullService bool
		var channelCoding uint8

		desc := findAC3Descriptor(es.ElementaryStreamDescriptors)
		if desc != nil {
			fullService = (desc.ComponentType & 0b1) != 0
			channelCoding = (desc.ComponentType >> 1) & 0b111
		}

		return &codecs.AC3{
			FullService:    fullService,
			ChannelsCoding: channelCoding,
		}, nil

	case astits.StreamTypeEAC3Audio:
		var fullService bool
		var channelCoding uint8

		desc := findEAC3Descriptor(es.ElementaryStreamDescriptors)
		if desc != nil {
			fullService = (desc.ComponentType & 0b1) != 0
			channelCoding = (desc.ComponentType >> 1) & 0b111
		}

		return &codecs.EAC3{
			FullService:    fullService,
			ChannelsCoding: channelCoding,
		}, nil

		// other

	case astits.StreamTypePrivateData:
		if id, ok := findRegistrationIdentifier(es.ElementaryStreamDescriptors); ok {
			switch id {
			case opusIdentifier:
				channelCount := findOpusChannelCount(es.ElementaryStreamDescriptors)
				if channelCount <= 0 {
					return nil, fmt.Errorf("invalid Opus channel count")
				}

				return &codecs.Opus{
					ChannelCount: channelCount,
				}, nil

			case klvaIdentifier:
				return &codecs.KLV{
					Synchronous: false,
				}, nil
			}
		} else if items := findDVBSubtitlingDescriptor(es.ElementaryStreamDescriptors); items != nil {
			return &codecs.DVBSubtitle{
				Items: items,
			}, nil
		}

	case astits.StreamTypeMetadata:
		desc := findKLVMetadataDescriptor(es.ElementaryStreamDescriptors)
		if desc != nil {
			return &codecs.KLV{
				Synchronous: true,
			}, nil
		}
	}

	return &codecs.Unsupported{}, nil
}

// Track is a MPEG-TS track.
type Track struct {
	PID   uint16
	Codec codecs.Codec

	isLeading  bool // Writer-only
	mp3Checked bool // Writer-only
}

func (t Track) marshal() (*astits.PMTElementaryStream, error) {
	switch c := t.Codec.(type) {
	// video

	case *codecs.H265:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypeH265Video,
		}, nil

	case *codecs.H264:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypeH264Video,
		}, nil

	case *codecs.MPEG4Video:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypeMPEG4Video,
		}, nil

	case *codecs.MPEG1Video:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			// we use MPEG-2 to signal that video can be either MPEG-1 or MPEG-2
			StreamType: astits.StreamTypeMPEG2Video,
		}, nil

	// audio

	case *codecs.Opus:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypePrivateData,
			ElementaryStreamDescriptors: []*astits.Descriptor{
				{
					// Length must be different than zero.
					// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
					Length: 1,
					Tag:    astits.DescriptorTagRegistration,
					Registration: &astits.DescriptorRegistration{
						FormatIdentifier: opusIdentifier,
					},
				},
				{
					// Length must be different than zero.
					// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
					Length: 1,
					Tag:    astits.DescriptorTagExtension,
					Extension: &astits.DescriptorExtension{
						Tag:     0x80,
						Unknown: &[]uint8{uint8(c.ChannelCount)},
					},
				},
			},
		}, nil

	case *codecs.AC3:
		var componentType uint8
		if c.ChannelCount != 0 { //nolint:staticcheck
			var channelCoding uint8
			switch c.ChannelCount { //nolint:staticcheck
			case 1:
				channelCoding = 1
			case 2:
				channelCoding = 2
			case 3, 4:
				channelCoding = 5
			default:
				channelCoding = 6
			}
			componentType = channelCoding<<1 | boolToUint8(c.FullService)
		} else {
			componentType = c.ChannelsCoding&0b111<<1 | boolToUint8(c.FullService)
		}

		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypeAC3Audio,
			ElementaryStreamDescriptors: []*astits.Descriptor{
				{
					// Length must be different than zero for astits writer
					// 1 byte flags + 1 byte component_type + 1 byte BSID = 3 bytes
					Length: 3,
					Tag:    astits.DescriptorTagAC3,
					AC3: &astits.DescriptorAC3{
						HasComponentType: true,
						ComponentType:    componentType,
						// BSID for standard AC-3 (not E-AC-3)
						HasBSID: true,
						BSID:    8,
					},
				},
			},
		}, nil

	case *codecs.EAC3:
		var componentType uint8
		if c.ChannelCount != 0 { //nolint:staticcheck
			var channelCoding uint8
			switch c.ChannelCount { //nolint:staticcheck
			case 1:
				channelCoding = 1
			case 2:
				channelCoding = 2
			case 3, 4:
				channelCoding = 5
			default:
				channelCoding = 6
			}
			componentType = channelCoding<<1 | boolToUint8(c.FullService)
		} else {
			componentType = c.ChannelsCoding&0b111<<1 | boolToUint8(c.FullService)
		}

		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypeEAC3Audio,
			ElementaryStreamDescriptors: []*astits.Descriptor{
				{
					// Length must be different than zero for astits writer
					// 1 byte flags + 1 byte component_type + 1 byte BSID = 3 bytes
					Length: 3,
					Tag:    astits.DescriptorTagEnhancedAC3,
					EnhancedAC3: &astits.DescriptorEnhancedAC3{
						HasComponentType: true,
						ComponentType:    componentType,
						// BSID=16 indicates E-AC-3
						HasBSID: true,
						BSID:    16,
					},
				},
			},
		}, nil

	case *codecs.MPEG4Audio:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypeAACAudio,
		}, nil

	case *codecs.MPEG4AudioLATM:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypeAACLATMAudio,
		}, nil

	case *codecs.MPEG1Audio:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypeMPEG1Audio,
		}, nil

		// other

	case *codecs.KLV:
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
				ElementaryPID: t.PID,
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
			ElementaryPID: t.PID,
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

	case *codecs.DVBSubtitle:
		return &astits.PMTElementaryStream{
			ElementaryPID: t.PID,
			StreamType:    astits.StreamTypePrivateData,
			ElementaryStreamDescriptors: []*astits.Descriptor{
				{
					// Length must be different than zero.
					// https://github.com/asticode/go-astits/blob/7c2bf6b71173d24632371faa01f28a9122db6382/descriptor.go#L2146-L2148
					Length: 1,
					Tag:    astits.DescriptorTagSubtitling,
					Subtitling: &astits.DescriptorSubtitling{
						Items: c.Items,
					},
				},
			},
		}, nil

	default:
		panic("unsupported codec")
	}
}

func (t *Track) unmarshal(es *astits.PMTElementaryStream) error {
	t.PID = es.ElementaryPID

	codec, err := findCodec(es)
	if err != nil {
		return err
	}
	t.Codec = codec

	return nil
}
