package mp4

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4/codecs"
)

// CodecInfo contains codec-dependent infos.
type CodecInfo struct {
	Width             int
	Height            int
	AV1SequenceHeader *av1.SequenceHeader
	H265SPS           *h265.SPS
	H264SPS           *h264.SPS
}

// Fill fills CodecInfo from a codecs.Codec.
func (ci *CodecInfo) Fill(codec codecs.Codec) error {
	switch codec := codec.(type) {
	case *codecs.AV1:
		av1SequenceHeader := &av1.SequenceHeader{}
		err := av1SequenceHeader.Unmarshal(codec.SequenceHeader)
		if err != nil {
			return fmt.Errorf("unable to parse AV1 sequence header: %w", err)
		}

		ci.Width = av1SequenceHeader.Width()
		ci.Height = av1SequenceHeader.Height()
		ci.AV1SequenceHeader = av1SequenceHeader
		return nil

	case *codecs.VP9:
		if codec.Width == 0 {
			return fmt.Errorf("VP9 parameters not provided")
		}

		ci.Width = codec.Width
		ci.Height = codec.Height
		return nil

	case *codecs.H265:
		if len(codec.VPS) == 0 || len(codec.SPS) == 0 || len(codec.PPS) == 0 {
			return fmt.Errorf("H265 parameters not provided")
		}

		h265SPS := &h265.SPS{}
		err := h265SPS.Unmarshal(codec.SPS)
		if err != nil {
			return fmt.Errorf("unable to parse H265 SPS: %w", err)
		}

		ci.Width = h265SPS.Width()
		ci.Height = h265SPS.Height()
		ci.H265SPS = h265SPS
		return nil

	case *codecs.H264:
		if len(codec.SPS) == 0 || len(codec.PPS) == 0 {
			return fmt.Errorf("H264 parameters not provided")
		}

		h264SPS := &h264.SPS{}
		err := h264SPS.Unmarshal(codec.SPS)
		if err != nil {
			return fmt.Errorf("unable to parse H264 SPS: %w", err)
		}

		ci.Width = h264SPS.Width()
		ci.Height = h264SPS.Height()
		ci.H264SPS = h264SPS
		return nil

	case *codecs.MPEG4Video:
		if len(codec.Config) == 0 {
			return fmt.Errorf("MPEG-4 Video config not provided")
		}

		// TODO: parse config and use real values
		ci.Width = 800
		ci.Height = 600
		return nil

	case *codecs.MPEG1Video:
		if len(codec.Config) == 0 {
			return fmt.Errorf("MPEG-1/2 Video config not provided")
		}

		// TODO: parse config and use real values
		ci.Width = 800
		ci.Height = 600
		return nil

	case *codecs.MJPEG:
		if codec.Width == 0 {
			return fmt.Errorf("M-JPEG parameters not provided")
		}

		ci.Width = codec.Width
		ci.Height = codec.Height
		return nil

	case *codecs.Opus, *codecs.MPEG4Audio, *codecs.MPEG1Audio, *codecs.AC3, *codecs.EAC3, *codecs.LPCM:
		return nil

	default:
		return fmt.Errorf("unsupported codec: %T", codec)
	}
}
