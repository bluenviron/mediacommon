package mpegts

import (
	"errors"
	"fmt"

	"github.com/asticode/go-astits"
	"github.com/bluenviron/mediacommon/pkg/codecs/mpeg4audio"
)

const (
	h264Identifier = 'H'<<24 | 'D'<<16 | 'M'<<8 | 'V'
	h265Identifier = 'H'<<24 | 'E'<<16 | 'V'<<8 | 'C'
	opusIdentifier = 'O'<<24 | 'p'<<16 | 'u'<<8 | 's'
)

var errUnsupportedTrack = errors.New("unsupported track")

func findMPEG4AudioConfig(dem *astits.Demuxer, pid uint16) (*mpeg4audio.Config, error) {
	for {
		data, err := dem.NextData()
		if err != nil {
			return nil, err
		}

		if data.PES == nil || data.PID != pid {
			continue
		}

		var adtsPkts mpeg4audio.ADTSPackets
		err = adtsPkts.Unmarshal(data.PES.Data)
		if err != nil {
			return nil, fmt.Errorf("unable to decode ADTS: %s", err)
		}

		pkt := adtsPkts[0]
		return &mpeg4audio.Config{
			Type:         pkt.Type,
			SampleRate:   pkt.SampleRate,
			ChannelCount: pkt.ChannelCount,
		}, nil
	}
}

func findOpusRegistration(descriptors []*astits.Descriptor) bool {
	for _, sd := range descriptors {
		if sd.Registration != nil {
			if sd.Registration.FormatIdentifier == opusIdentifier {
				return true
			}
		}
	}
	return false
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

func findOpusCodec(descriptors []*astits.Descriptor) *CodecOpus {
	if !findOpusRegistration(descriptors) {
		return nil
	}

	channelCount := findOpusChannelCount(descriptors)
	if channelCount <= 0 {
		return nil
	}

	return &CodecOpus{
		ChannelCount: channelCount,
	}
}

// Track is a MPEG-TS track.
type Track struct {
	PID   uint16
	Codec Codec
}

// Marshal encodes a track into a astits.PMTElementaryStream.
func (t *Track) Marshal() (*astits.PMTElementaryStream, error) {
	return t.Codec.Marshal(t.PID)
}

// Unmarshal decodes a track from a astits.PMTElementaryStream.
func (t *Track) Unmarshal(dem *astits.Demuxer, es *astits.PMTElementaryStream) error {
	t.PID = es.ElementaryPID

	switch es.StreamType {
	case astits.StreamTypeH264Video:
		t.Codec = &CodecH264{}
		return nil

	case astits.StreamTypeH265Video:
		t.Codec = &CodecH265{}
		return nil

	case astits.StreamTypeAACAudio:
		conf, err := findMPEG4AudioConfig(dem, es.ElementaryPID)
		if err != nil {
			return err
		}

		t.Codec = &CodecMPEG4Audio{
			Config: *conf,
		}
		return nil

	case astits.StreamTypePrivateData:
		codec := findOpusCodec(es.ElementaryStreamDescriptors)
		if codec != nil {
			t.Codec = codec
			return nil
		}
	}

	return errUnsupportedTrack
}

// FindTracks finds the tracks in a MPEG-TS stream.
func FindTracks(dem *astits.Demuxer) ([]*Track, error) {
	var tracks []*Track

	for {
		data, err := dem.NextData()
		if err != nil {
			return nil, err
		}

		if data.PMT != nil {
			for _, es := range data.PMT.ElementaryStreams {
				var track Track
				err := track.Unmarshal(dem, es)
				if err != nil {
					if err == errUnsupportedTrack {
						continue
					}
					return nil, err
				}

				tracks = append(tracks, &track)
			}
			break
		}
	}

	if tracks == nil {
		return nil, fmt.Errorf("no tracks found")
	}

	return tracks, nil
}
