// Package pmp4 contains a MP4 presentation reader and writer.
package pmp4

import (
	"errors"
	"fmt"
	"io"
	"time"

	amp4 "github.com/abema/go-mp4"

	imp4 "github.com/bluenviron/mediacommon/v2/internal/mp4"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/fmp4/seekablebuffer"
)

const (
	globalTimescale = 1000
	maxSamples      = 30 * 60 * 60 * 48 // 30 fps @ 2 days
	maxChunks       = maxSamples
)

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
		waitingFtyp readState = iota
		waitingMoov
		waitingMvhd
		waitingTrak
		waitingElst
		waitingTkhd
		waitingMdhd
		waitingStsd
		readingCodec
		waitingSamples
		waitingSampleProps
		mdat
	)

	var state readState
	var trackDuration uint32
	var curTrack *Track
	var codecBoxesReader *imp4.CodecBoxesReader

	type chunk struct {
		sampleCount int
		offset      uint32
	}

	var curChunks []*chunk
	var curSampleSizes []uint32

	_, err := amp4.ReadBoxStructure(r, func(h *amp4.ReadHandle) (any, error) {
		if state == readingCodec {
			ret, err := codecBoxesReader.Read(h)
			if err != nil {
				if errors.Is(err, imp4.ErrReadEnded) {
					if codecBoxesReader.Codec != nil {
						curTrack.Codec = codecBoxesReader.Codec
					} else {
						p.Tracks = p.Tracks[:len(p.Tracks)-1]
					}
					state = waitingSamples
				} else {
					return nil, err
				}
			} else {
				return ret, nil
			}
		}

		switch h.BoxInfo.Type.String() {
		case "ftyp":
			if state != waitingFtyp {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingMoov

		case "free":
			if state == waitingFtyp {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

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

		case "udta":
			if state != waitingTrak && state != waitingSampleProps {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

		case "trak":
			if state != waitingTrak && state != waitingSampleProps {
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
			if state != waitingElst {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			return h.Expand()

		case "elst":
			if state != waitingElst {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			state = waitingMdhd

		case "mdia":
			if state != waitingMdhd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

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
			state = waitingStsd

		case "hdlr", "vmhd", "smhd", "nmhd", "dinf":
			if state != waitingStsd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

		case "minf", "stbl":
			if state != waitingStsd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			return h.Expand()

		case "stsd":
			if state != waitingStsd {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			codecBoxesReader = &imp4.CodecBoxesReader{}
			state = readingCodec
			return h.Expand()

		case "stts":
			if state != waitingSamples {
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

			state = waitingSampleProps

		case "stss":
			if state != waitingSampleProps {
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
			if state != waitingSampleProps {
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
			if state != waitingSampleProps {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			stsc := box.(*amp4.Stsc)

			if len(stsc.Entries) != 0 {
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
			}

		case "stsz":
			if state != waitingSampleProps {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			stsz := box.(*amp4.Stsz)

			curSampleSizes = stsz.EntrySize

		case "stco":
			if state != waitingSampleProps {
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
						_, err2 := r.Seek(int64(sampleOffset), io.SeekStart)
						if err != nil {
							return nil, err2
						}

						buf := make([]byte, sampleSize)
						_, err2 = io.ReadFull(r, buf)
						return buf, err2
					}

					off += sampleSize
					i++
				}
			}

		case "sdtp":
			if state != waitingSampleProps {
				return nil, fmt.Errorf("unexpected box '%v'", h.BoxInfo.Type)
			}

		case "mdat":
			if state != waitingTrak && state != waitingSampleProps {
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
func (p Presentation) Marshal(w io.Writer) error {
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
		var pl []byte
		pl, err = sa.GetPayload()
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
