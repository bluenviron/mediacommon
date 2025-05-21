// Package pmp4 contains a MP4 presentation reader and writer.
package pmp4

import (
	"fmt"
	"io"
	"time"

	amp4 "github.com/abema/go-mp4"

	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/av1"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/mpeg4audio"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/fmp4/seekablebuffer"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4"
)

const (
	globalTimescale = 1000
	maxSamples      = 30 * 60 * 60 * 48 // 30 fps @ 2 days
	maxChunks       = maxSamples
)

func av1FindSequenceHeader(buf []byte) ([]byte, error) {
	var tu av1.Bitstream
	err := tu.Unmarshal(buf)
	if err != nil {
		return nil, err
	}

	for _, obu := range tu {
		var h av1.OBUHeader
		err := h.Unmarshal(obu)
		if err != nil {
			return nil, err
		}

		if h.Type == av1.OBUTypeSequenceHeader {
			var parsed av1.SequenceHeader
			err = parsed.Unmarshal(obu)
			if err != nil {
				return nil, err
			}

			return obu, nil
		}
	}

	return nil, fmt.Errorf("AV1 sequence header not found")
}

func h264FindParams(avcc *amp4.AVCDecoderConfiguration) ([]byte, []byte, error) {
	if len(avcc.SequenceParameterSets) == 0 {
		return nil, nil, fmt.Errorf("H264 SPS not provided")
	}
	if len(avcc.SequenceParameterSets) > 1 {
		return nil, nil, fmt.Errorf("multiple H264 SPS are not supported")
	}
	if len(avcc.PictureParameterSets) == 0 {
		return nil, nil, fmt.Errorf("H264 PPS not provided")
	}
	if len(avcc.PictureParameterSets) > 1 {
		return nil, nil, fmt.Errorf("multiple H264 PPS are not supported")
	}

	sps := avcc.SequenceParameterSets[0].NALUnit
	var spsp h264.SPS
	err := spsp.Unmarshal(sps)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse H264 SPS: %w", err)
	}

	pps := avcc.PictureParameterSets[0].NALUnit
	if len(pps) == 0 {
		return nil, nil, fmt.Errorf("invalid H264 PPS")
	}

	return sps, pps, nil
}

func h265FindParams(params []amp4.HEVCNaluArray) ([]byte, []byte, []byte, error) {
	var vps []byte
	var sps []byte
	var pps []byte

	for _, arr := range params {
		switch h265.NALUType(arr.NaluType) {
		case h265.NALUType_VPS_NUT, h265.NALUType_SPS_NUT, h265.NALUType_PPS_NUT:
			if arr.NumNalus != 1 {
				return nil, nil, nil, fmt.Errorf("multiple H265 VPS/SPS/PPS are not supported")
			}

			switch h265.NALUType(arr.NaluType) {
			case h265.NALUType_VPS_NUT:
				vps = arr.Nalus[0].NALUnit

			case h265.NALUType_SPS_NUT:
				sps = arr.Nalus[0].NALUnit

				var spsp h265.SPS
				err := spsp.Unmarshal(sps)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("unable to parse H265 SPS: %w", err)
				}

			case h265.NALUType_PPS_NUT:
				pps = arr.Nalus[0].NALUnit
			}
		}
	}

	if len(vps) == 0 {
		return nil, nil, nil, fmt.Errorf("H265 VPS not provided")
	}

	if len(sps) == 0 {
		return nil, nil, nil, fmt.Errorf("H265 SPS not provided")
	}

	if len(pps) == 0 {
		return nil, nil, nil, fmt.Errorf("H265 PPS not provided")
	}

	return vps, sps, pps, nil
}

func esdsFindDecoderConf(descriptors []amp4.Descriptor) *amp4.DecoderConfigDescriptor {
	for _, desc := range descriptors {
		if desc.Tag == amp4.DecoderConfigDescrTag {
			return desc.DecoderConfigDescriptor
		}
	}
	return nil
}

func esdsFindDecoderSpecificInfo(descriptors []amp4.Descriptor) []byte {
	for _, desc := range descriptors {
		if desc.Tag == amp4.DecSpecificInfoTag {
			return desc.Data
		}
	}
	return nil
}

func durationMp4ToGo(v int64, timeScale uint32) time.Duration {
	timeScale64 := int64(timeScale)
	secs := v / timeScale64
	dec := v % timeScale64
	return time.Duration(secs)*time.Second + time.Duration(dec)*time.Second/time.Duration(timeScale64)
}

// Presentation is timed sequence of video/audio samples.
type Presentation struct {
	Tracks []*Track
}

// Unmarshal decodes a Presentation.
func (p *Presentation) Unmarshal(r io.ReadSeeker) error {
	type readState int

	const (
		waitingMoov readState = iota
		waitingMvhd
		waitingTrak
		waitingElst
		waitingTkhd
		waitingMdhd
		waitingCodec
		waitingAv1C
		waitingVpcC
		waitingHvcC
		waitingAvcC
		waitingVideoEsds
		waitingAudioEsds
		waitingDOps
		waitingDac3
		waitingPcmC
		waitingStts
		mdat
	)

	state := waitingMoov
	var trackDuration uint32
	var curTrack *Track
	var width int
	var height int
	var sampleRate int
	var channelCount int

	type chunk struct {
		sampleCount int
		offset      uint32
	}

	var curChunks []*chunk
	var curSampleSizes []uint32

	_, err := amp4.ReadBoxStructure(r, func(h *amp4.ReadHandle) (interface{}, error) {
		switch h.BoxInfo.Type.String() {
		case "ftyp", "hdlr", "vmhd", "dinf", "smhd":

		case "moov":
			if state != waitingMoov {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingMvhd
			return h.Expand()

		case "mvhd":
			if state != waitingMvhd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingTrak
			return h.Expand()

		case "trak":
			if state != waitingTrak {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			curTrack = &Track{}
			curChunks = nil
			curSampleSizes = nil
			p.Tracks = append(p.Tracks, curTrack)
			state = waitingTkhd
			return h.Expand()

		case "tkhd":
			if state != waitingTkhd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			tkhd := box.(*amp4.Tkhd)

			curTrack.ID = int(tkhd.TrackID)
			state = waitingElst

		case "edts":
			return h.Expand()

		case "elst":
			if state != waitingElst {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingMdhd

		case "mdia":
			return h.Expand()

		case "mdhd":
			if state != waitingMdhd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			mdhd := box.(*amp4.Mdhd)

			if mdhd.Timescale == 0 {
				return nil, fmt.Errorf("invalid timescale")
			}

			curTrack.TimeScale = mdhd.Timescale
			trackDuration = mdhd.DurationV0
			state = waitingCodec

		case "minf", "stbl", "stsd":
			return h.Expand()

		case "avc1":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}
			state = waitingAvcC
			return h.Expand()

		case "avcC":
			if state != waitingAvcC {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			avcc := box.(*amp4.AVCDecoderConfiguration)

			sps, pps, err := h264FindParams(avcc)
			if err != nil {
				return nil, err
			}

			curTrack.Codec = &mp4.CodecH264{
				SPS: sps,
				PPS: pps,
			}
			state = waitingStts

		case "vp09":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			vp09 := box.(*amp4.VisualSampleEntry)

			if vp09.Width == 0 || vp09.Height == 0 {
				return nil, fmt.Errorf("VP9 parameters not provided")
			}

			width = int(vp09.Width)
			height = int(vp09.Height)
			state = waitingVpcC
			return h.Expand()

		case "vpcC":
			if state != waitingVpcC {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			vpcc := box.(*amp4.VpcC)

			curTrack.Codec = &mp4.CodecVP9{
				Width:             width,
				Height:            height,
				Profile:           vpcc.Profile,
				BitDepth:          vpcc.BitDepth,
				ChromaSubsampling: vpcc.ChromaSubsampling,
				ColorRange:        vpcc.VideoFullRangeFlag != 0,
			}
			state = waitingStts

		case "vp08": // VP8, not supported yet
			return nil, nil

		case "hev1", "hvc1":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}
			state = waitingHvcC
			return h.Expand()

		case "hvcC":
			if state != waitingHvcC {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			hvcc := box.(*amp4.HvcC)

			vps, sps, pps, err := h265FindParams(hvcc.NaluArrays)
			if err != nil {
				return nil, err
			}

			curTrack.Codec = &mp4.CodecH265{
				VPS: vps,
				SPS: sps,
				PPS: pps,
			}
			state = waitingStts

		case "mp4a":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			mp4a := box.(*amp4.AudioSampleEntry)

			sampleRate = int(mp4a.SampleRate / 65536)
			channelCount = int(mp4a.ChannelCount)
			state = waitingAudioEsds
			return h.Expand()

		case "av01":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}
			state = waitingAv1C
			return h.Expand()

		case "av1C":
			if state != waitingAv1C {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			av1c := box.(*amp4.Av1C)

			sequenceHeader, err := av1FindSequenceHeader(av1c.ConfigOBUs)
			if err != nil {
				return nil, err
			}

			curTrack.Codec = &mp4.CodecAV1{
				SequenceHeader: sequenceHeader,
			}
			state = waitingStts

		case "Opus":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}
			state = waitingDOps
			return h.Expand()

		case "dOps":
			if state != waitingDOps {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			dops := box.(*amp4.DOps)

			curTrack.Codec = &mp4.CodecOpus{
				ChannelCount: int(dops.OutputChannelCount),
			}
			state = waitingStts

		case "mp4v":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			mp4v := box.(*amp4.VisualSampleEntry)

			width = int(mp4v.Width)
			height = int(mp4v.Height)
			state = waitingVideoEsds
			return h.Expand()

		case "esds":
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			esds := box.(*amp4.Esds)

			conf := esdsFindDecoderConf(esds.Descriptors)
			if conf == nil {
				return nil, fmt.Errorf("unable to find decoder config")
			}

			switch state {
			case waitingVideoEsds:
				switch conf.ObjectTypeIndication {
				case imp4.ObjectTypeIndicationVisualISO14496part2:
					spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
					if len(spec) == 0 {
						return nil, fmt.Errorf("unable to find decoder specific info")
					}

					curTrack.Codec = &mp4.CodecMPEG4Video{
						Config: spec,
					}

				case imp4.ObjectTypeIndicationVisualISO1318part2Main:
					spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
					if len(spec) == 0 {
						return nil, fmt.Errorf("unable to find decoder specific info")
					}

					curTrack.Codec = &mp4.CodecMPEG1Video{
						Config: spec,
					}

				case imp4.ObjectTypeIndicationVisualISO10918part1:
					if width == 0 || height == 0 {
						return nil, fmt.Errorf("M-JPEG parameters not provided")
					}

					curTrack.Codec = &mp4.CodecMJPEG{
						Width:  width,
						Height: height,
					}

				default:
					return nil, fmt.Errorf("unsupported object type indication: 0x%.2x", conf.ObjectTypeIndication)
				}

				state = waitingStts

			case waitingAudioEsds:
				switch conf.ObjectTypeIndication {
				case imp4.ObjectTypeIndicationAudioISO14496part3:
					spec := esdsFindDecoderSpecificInfo(esds.Descriptors)
					if len(spec) == 0 {
						return nil, fmt.Errorf("unable to find decoder specific info")
					}

					var c mpeg4audio.Config
					err := c.Unmarshal(spec)
					if err != nil {
						return nil, fmt.Errorf("invalid MPEG-4 Audio configuration: %w", err)
					}

					curTrack.Codec = &mp4.CodecMPEG4Audio{
						Config: c,
					}

				case imp4.ObjectTypeIndicationAudioISO11172part3:
					curTrack.Codec = &mp4.CodecMPEG1Audio{
						SampleRate:   sampleRate,
						ChannelCount: channelCount,
					}

				default:
					return nil, fmt.Errorf("unsupported object type indication: 0x%.2x", conf.ObjectTypeIndication)
				}

			default:
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingStts

		case "ac-3":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			ac3 := box.(*amp4.AudioSampleEntry)

			sampleRate = int(ac3.SampleRate / 65536)
			channelCount = int(ac3.ChannelCount)
			state = waitingDac3
			return h.Expand()

		case "dac3":
			if state != waitingDac3 {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			dac3 := box.(*amp4.Dac3)

			curTrack.Codec = &mp4.CodecAC3{
				SampleRate:   sampleRate,
				ChannelCount: channelCount,
				Fscod:        dac3.Fscod,
				Bsid:         dac3.Bsid,
				Bsmod:        dac3.Bsmod,
				Acmod:        dac3.Acmod,
				LfeOn:        dac3.LfeOn != 0,
				BitRateCode:  dac3.BitRateCode,
			}
			state = waitingStts

		case "ipcm":
			if state != waitingCodec {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			ac3 := box.(*amp4.AudioSampleEntry)

			sampleRate = int(ac3.SampleRate / 65536)
			channelCount = int(ac3.ChannelCount)
			state = waitingPcmC
			return h.Expand()

		case "pcmC":
			if state != waitingPcmC {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			pcmc := box.(*amp4.PcmC)

			curTrack.Codec = &mp4.CodecLPCM{
				LittleEndian: (pcmc.FormatFlags & 0x01) != 0,
				BitDepth:     int(pcmc.PCMSampleSize),
				SampleRate:   sampleRate,
				ChannelCount: channelCount,
			}
			state = waitingStts

		case "stts":
			if state != waitingStts {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			stts := box.(*amp4.Stts)

			for _, entry := range stts.Entries {
				if (len(curTrack.Samples) + int(entry.SampleCount)) > maxSamples {
					return nil, fmt.Errorf("max samples reached")
				}

				for range entry.SampleCount {
					curTrack.Samples = append(curTrack.Samples, &Sample{
						Duration: entry.SampleDelta,
					})
				}
			}

			sampleDuration := uint32(0)
			for _, sa := range curTrack.Samples {
				sampleDuration += sa.Duration
			}

			curTrack.TimeOffset = int32(trackDuration) - int32(sampleDuration)

			state = waitingTrak

		case "stss":
			if state != waitingTrak || curTrack == nil {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			stss := box.(*amp4.Stss)

			for _, sample := range curTrack.Samples {
				sample.IsNonSyncSample = true
			}

			for _, number := range stss.SampleNumber {
				if int(number-1) >= len(curTrack.Samples) {
					return nil, fmt.Errorf("invalid stss")
				}
				curTrack.Samples[number-1].IsNonSyncSample = false
			}

		case "ctts":
			if state != waitingTrak || curTrack == nil {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			ctts := box.(*amp4.Ctts)

			i := 0

			for _, entry := range ctts.Entries {
				if (i + int(entry.SampleCount)) > len(curTrack.Samples) {
					return nil, fmt.Errorf("invalid ctts")
				}

				for range entry.SampleCount {
					curTrack.Samples[i].PTSOffset = entry.SampleOffsetV1
					i++
				}
			}

		case "stsc":
			if state != waitingTrak || curTrack == nil {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			stsc := box.(*amp4.Stsc)

			if len(stsc.Entries) == 0 {
				return nil, fmt.Errorf("invalid stsc")
			}

			prevFirstChunk := uint32(0)
			i := 0

			for _, entry := range stsc.Entries {
				chunkCount := entry.FirstChunk - prevFirstChunk

				if (len(curChunks) + int(chunkCount)) > maxChunks {
					return nil, fmt.Errorf("invalid stsc")
				}

				if entry.SamplesPerChunk == 0 {
					return nil, fmt.Errorf("invalid stsc")
				}

				for range chunkCount {
					if (i + int(entry.SamplesPerChunk)) > len(curTrack.Samples) {
						return nil, fmt.Errorf("invalid stsc")
					}

					curChunks = append(curChunks, &chunk{
						sampleCount: int(entry.SamplesPerChunk),
					})

					i += int(entry.SamplesPerChunk)
				}
				prevFirstChunk = entry.FirstChunk
			}

			if i != len(curTrack.Samples) {
				remaining := len(curTrack.Samples) - i
				lastEntry := stsc.Entries[len(stsc.Entries)-1]

				if (remaining % int(lastEntry.SamplesPerChunk)) != 0 {
					return nil, fmt.Errorf("invalid stsc")
				}

				count := remaining / int(lastEntry.SamplesPerChunk)
				for range count {
					curChunks = append(curChunks, &chunk{
						sampleCount: int(lastEntry.SamplesPerChunk),
					})
				}
			}

		case "stsz":
			if state != waitingTrak || curTrack == nil {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			stsz := box.(*amp4.Stsz)

			curSampleSizes = stsz.EntrySize

		case "stco":
			if state != waitingTrak || curTrack == nil {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			stco := box.(*amp4.Stco)

			if len(stco.ChunkOffset) != len(curChunks) {
				return nil, fmt.Errorf("invalid stco")
			}

			for i, chunk := range curChunks {
				chunk.offset = stco.ChunkOffset[i]
			}

			if len(curSampleSizes) != len(curTrack.Samples) {
				return nil, fmt.Errorf("invalid stsz")
			}

			i := 0

			for _, chunk := range curChunks {
				off := chunk.offset

				for range chunk.sampleCount {
					sampleSize := curSampleSizes[i]
					sampleOffset := off

					curTrack.Samples[i].PayloadSize = sampleSize

					curTrack.Samples[i].GetPayload = func() ([]byte, error) {
						_, err = r.Seek(int64(sampleOffset), io.SeekStart)
						if err != nil {
							return nil, err
						}

						buf := make([]byte, sampleSize)
						_, err = io.ReadFull(r, buf)
						return buf, err
					}

					off += sampleSize
					i++
				}
			}

		case "mdat":
			if state != waitingTrak {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = mdat

		default:
			return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	if state != mdat {
		return fmt.Errorf("parse error")
	}

	if len(p.Tracks) == 0 {
		return fmt.Errorf("no tracks found")
	}

	return nil
}

// Marshal encodes a Presentation.
func (p *Presentation) Marshal(w io.Writer) error {
	/*
		|ftyp|
		|moov|
		|    |mvhd|
		|    |trak|
		|    |trak|
		|    |....|
		|mdat|
	*/

	dataSize, sortedSamples := p.sortSamples()

	err := p.marshalFtypAndMoov(w)
	if err != nil {
		return err
	}

	return p.marshalMdat(w, dataSize, sortedSamples)
}

func (p *Presentation) sortSamples() (uint32, []*Sample) {
	sampleCount := 0
	for _, track := range p.Tracks {
		sampleCount += len(track.Samples)
	}

	processedSamples := make([]int, len(p.Tracks))
	elapsed := make([]int64, len(p.Tracks))
	offset := uint32(0)
	sortedSamples := make([]*Sample, sampleCount)
	pos := 0

	for i, track := range p.Tracks {
		elapsed[i] = int64(track.TimeOffset)
	}

	for {
		bestTrack := -1
		var bestElapsed time.Duration

		for i, track := range p.Tracks {
			if processedSamples[i] < len(track.Samples) {
				elapsedGo := durationMp4ToGo(elapsed[i], track.TimeScale)

				if bestTrack == -1 || elapsedGo < bestElapsed {
					bestTrack = i
					bestElapsed = elapsedGo
				}
			}
		}

		if bestTrack == -1 {
			break
		}

		sample := p.Tracks[bestTrack].Samples[processedSamples[bestTrack]]
		sample.offset = offset

		processedSamples[bestTrack]++
		elapsed[bestTrack] += int64(sample.Duration)
		offset += sample.PayloadSize
		sortedSamples[pos] = sample
		pos++
	}

	return offset, sortedSamples
}

func (p *Presentation) marshalFtypAndMoov(w io.Writer) error {
	var outBuf seekablebuffer.Buffer
	mw := &imp4.Writer{W: &outBuf}
	mw.Initialize()

	_, err := mw.WriteBox(&amp4.Ftyp{ // <ftyp/>
		MajorBrand:   [4]byte{'i', 's', 'o', 'm'},
		MinorVersion: 1,
		CompatibleBrands: []amp4.CompatibleBrandElem{
			{CompatibleBrand: [4]byte{'i', 's', 'o', 'm'}},
			{CompatibleBrand: [4]byte{'i', 's', 'o', '2'}},
			{CompatibleBrand: [4]byte{'m', 'p', '4', '1'}},
			{CompatibleBrand: [4]byte{'m', 'p', '4', '2'}},
		},
	})
	if err != nil {
		return err
	}

	_, err = mw.WriteBoxStart(&amp4.Moov{}) // <moov>
	if err != nil {
		return err
	}

	mvhd := &amp4.Mvhd{ // <mvhd/>
		Timescale:   globalTimescale,
		Rate:        65536,
		Volume:      256,
		Matrix:      [9]int32{0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000},
		NextTrackID: uint32(len(p.Tracks) + 1),
	}
	mvhdOffset, err := mw.WriteBox(mvhd)
	if err != nil {
		return err
	}

	stcos := make([]*amp4.Stco, len(p.Tracks))
	stcosOffsets := make([]int, len(p.Tracks))

	for i, track := range p.Tracks {
		var res *headerTrackMarshalResult
		res, err = track.marshal(mw)
		if err != nil {
			return err
		}

		stcos[i] = res.stco
		stcosOffsets[i] = res.stcoOffset

		if res.presentationDuration > mvhd.DurationV0 {
			mvhd.DurationV0 = res.presentationDuration
		}
	}

	err = mw.RewriteBox(mvhdOffset, mvhd)
	if err != nil {
		return err
	}

	err = mw.WriteBoxEnd() // </moov>
	if err != nil {
		return err
	}

	moovEndOffset, err := outBuf.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	dataOffset := moovEndOffset + 8

	for i := range p.Tracks {
		for j := range stcos[i].ChunkOffset {
			stcos[i].ChunkOffset[j] += uint32(dataOffset)
		}

		err = mw.RewriteBox(stcosOffsets[i], stcos[i])
		if err != nil {
			return err
		}
	}

	_, err = w.Write(outBuf.Bytes())
	return err
}

func (p *Presentation) marshalMdat(w io.Writer, dataSize uint32, sortedSamples []*Sample) error {
	mdatSize := uint32(8) + dataSize

	_, err := w.Write([]byte{byte(mdatSize >> 24), byte(mdatSize >> 16), byte(mdatSize >> 8), byte(mdatSize)})
	if err != nil {
		return err
	}

	_, err = w.Write([]byte{'m', 'd', 'a', 't'})
	if err != nil {
		return err
	}

	for _, sa := range sortedSamples {
		pl, err := sa.GetPayload()
		if err != nil {
			return err
		}

		_, err = w.Write(pl)
		if err != nil {
			return err
		}
	}

	return nil
}
