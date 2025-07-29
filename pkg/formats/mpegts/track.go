package mpegts

import (
	"fmt"

	"github.com/asticode/go-astits"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/ac3"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
)

func findMPEG4AudioConfig(dem *robustDemuxer, pid uint16) (*mpeg4audio.AudioSpecificConfig, error) {
	for {
		data, err := dem.nextData()
		if err != nil {
			return nil, err
		}

		if data.PES == nil || data.PID != pid {
			continue
		}

		var adtsPkts mpeg4audio.ADTSPackets
		err = adtsPkts.Unmarshal(data.PES.Data)
		if err != nil {
			return nil, fmt.Errorf("unable to decode ADTS: %w", err)
		}

		pkt := adtsPkts[0]
		return &mpeg4audio.AudioSpecificConfig{
			Type:         pkt.Type,
			SampleRate:   pkt.SampleRate,
			ChannelCount: pkt.ChannelCount,
		}, nil
	}
}

func findAC3Parameters(dem *robustDemuxer, pid uint16) (int, int, error) {
	for {
		data, err := dem.nextData()
		if err != nil {
			return 0, 0, err
		}

		if data.PES == nil || data.PID != pid {
			continue
		}

		var syncInfo ac3.SyncInfo
		err = syncInfo.Unmarshal(data.PES.Data)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid AC-3 frame: %w", err)
		}

		var bsi ac3.BSI
		err = bsi.Unmarshal(data.PES.Data[5:])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid AC-3 frame: %w", err)
		}

		return syncInfo.SampleRate(), bsi.ChannelCount(), nil
	}
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

func findDVBSubtitlingDescriptor(descriptors []*astits.Descriptor) *subtitlingDescriptor {
	for _, sd := range descriptors {
		if sd.Tag == astits.DescriptorTagSubtitling && sd.Subtitling != nil {
			return &subtitlingDescriptor{
				tag:    sd.Tag,
				length: sd.Length,
				items:  sd.Subtitling.Items,
			}
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

func findCodec(dem *robustDemuxer, es *astits.PMTElementaryStream) (Codec, error) {
	switch es.StreamType {
	case astits.StreamTypeH265Video:
		return &CodecH265{}, nil

	case astits.StreamTypeH264Video:
		return &CodecH264{}, nil

	case astits.StreamTypeMPEG4Video:
		return &CodecMPEG4Video{}, nil

	case astits.StreamTypeMPEG2Video, astits.StreamTypeMPEG1Video:
		return &CodecMPEG1Video{}, nil

	case astits.StreamTypeAACAudio:
		conf, err := findMPEG4AudioConfig(dem, es.ElementaryPID)
		if err != nil {
			return nil, err
		}

		return &CodecMPEG4Audio{
			Config: *conf,
		}, nil

	case astits.StreamTypeAACLATMAudio:
		return &CodecMPEG4AudioLATM{}, nil

	case astits.StreamTypeMPEG1Audio:
		return &CodecMPEG1Audio{}, nil

	case astits.StreamTypeAC3Audio:
		sampleRate, channelCount, err := findAC3Parameters(dem, es.ElementaryPID)
		if err != nil {
			return nil, err
		}

		return &CodecAC3{
			SampleRate:   sampleRate,
			ChannelCount: channelCount,
		}, nil

	case astits.StreamTypePrivateData:
		id, ok := findRegistrationIdentifier(es.ElementaryStreamDescriptors)
		if ok {
			switch id {
			case opusIdentifier:
				channelCount := findOpusChannelCount(es.ElementaryStreamDescriptors)
				if channelCount <= 0 {
					return nil, fmt.Errorf("invalid Opus channel count")
				}

				return &CodecOpus{
					ChannelCount: channelCount,
				}, nil

			case klvaIdentifier:
				return &CodecKLV{
					Synchronous: false,
				}, nil
			}
		} else {
			subtitlingDescriptor := findDVBSubtitlingDescriptor(es.ElementaryStreamDescriptors)
			if subtitlingDescriptor != nil {
				return &CodecDVB{
					descriptor: subtitlingDescriptor,
				}, nil
			}
		}

	case astits.StreamTypeMetadata:
		desc := findKLVMetadataDescriptor(es.ElementaryStreamDescriptors)
		if desc != nil {
			return &CodecKLV{
				Synchronous: true,
			}, nil
		}
	}

	return &CodecUnsupported{}, nil
}

// Track is a MPEG-TS track.
type Track struct {
	PID   uint16
	Codec Codec

	isLeading  bool // Writer-only
	mp3Checked bool // Writer-only
}

func (t *Track) marshal() (*astits.PMTElementaryStream, error) {
	return t.Codec.marshal(t.PID)
}

func (t *Track) unmarshal(dem *robustDemuxer, es *astits.PMTElementaryStream) error {
	t.PID = es.ElementaryPID

	codec, err := findCodec(dem, es)
	if err != nil {
		return err
	}
	t.Codec = codec

	return nil
}
