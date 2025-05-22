package mp4

import (
	"fmt"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

// CodecInfo contains codec-dependent infos.
type CodecInfo struct {
	Width             int
	Height            int
	AV1SequenceHeader *av1.SequenceHeader
	H265SPS           *h265.SPS
	H264SPS           *h264.SPS
}

// ExtractCodecInfo fills CodecInfo.
func ExtractCodecInfo(codec mp4.Codec) (*CodecInfo, error) {
	switch codec := codec.(type) {
	case *mp4.CodecAV1:
		av1SequenceHeader := &av1.SequenceHeader{}
		err := av1SequenceHeader.Unmarshal(codec.SequenceHeader)
		if err != nil {
			return nil, fmt.Errorf("unable to parse AV1 sequence header: %w", err)
		}

		return &CodecInfo{
			Width:             av1SequenceHeader.Width(),
			Height:            av1SequenceHeader.Height(),
			AV1SequenceHeader: av1SequenceHeader,
		}, nil

	case *mp4.CodecVP9:
		if codec.Width == 0 {
			return nil, fmt.Errorf("VP9 parameters not provided")
		}

		return &CodecInfo{
			Width:  codec.Width,
			Height: codec.Height,
		}, nil

	case *mp4.CodecH265:
		if len(codec.VPS) == 0 || len(codec.SPS) == 0 || len(codec.PPS) == 0 {
			return nil, fmt.Errorf("H265 parameters not provided")
		}

		h265SPS := &h265.SPS{}
		err := h265SPS.Unmarshal(codec.SPS)
		if err != nil {
			return nil, fmt.Errorf("unable to parse H265 SPS: %w", err)
		}

		return &CodecInfo{
			Width:   h265SPS.Width(),
			Height:  h265SPS.Height(),
			H265SPS: h265SPS,
		}, nil

	case *mp4.CodecH264:
		if len(codec.SPS) == 0 || len(codec.PPS) == 0 {
			return nil, fmt.Errorf("H264 parameters not provided")
		}

		h264SPS := &h264.SPS{}
		err := h264SPS.Unmarshal(codec.SPS)
		if err != nil {
			return nil, fmt.Errorf("unable to parse H264 SPS: %w", err)
		}

		return &CodecInfo{
			Width:   h264SPS.Width(),
			Height:  h264SPS.Height(),
			H264SPS: h264SPS,
		}, nil

	case *mp4.CodecMPEG4Video:
		if len(codec.Config) == 0 {
			return nil, fmt.Errorf("MPEG-4 Video config not provided")
		}

		// TODO: parse config and use real values
		return &CodecInfo{
			Width:  800,
			Height: 600,
		}, nil

	case *mp4.CodecMPEG1Video:
		if len(codec.Config) == 0 {
			return nil, fmt.Errorf("MPEG-1/2 Video config not provided")
		}

		// TODO: parse config and use real values
		return &CodecInfo{
			Width:  800,
			Height: 600,
		}, nil

	case *mp4.CodecMJPEG:
		if codec.Width == 0 {
			return nil, fmt.Errorf("M-JPEG parameters not provided")
		}

		return &CodecInfo{
			Width:  codec.Width,
			Height: codec.Height,
		}, nil

	case *mp4.CodecOpus, *mp4.CodecMPEG4Audio, *mp4.CodecMPEG1Audio, *mp4.CodecAC3, *mp4.CodecLPCM:
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported codec: %T", codec)
	}
}
