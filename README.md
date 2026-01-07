# mediacommon

[![Test](https://github.com/bluenviron/mediacommon/actions/workflows/test.yml/badge.svg)](https://github.com/bluenviron/mediacommon/actions/workflows/test.yml)
[![Lint](https://github.com/bluenviron/mediacommon/actions/workflows/lint.yml/badge.svg)](https://github.com/bluenviron/mediacommon/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bluenviron/mediacommon)](https://goreportcard.com/report/github.com/bluenviron/mediacommon)
[![CodeCov](https://codecov.io/gh/bluenviron/mediacommon/branch/main/graph/badge.svg)](https://app.codecov.io/gh/bluenviron/mediacommon/tree/main)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/bluenviron/mediacommon/v2)](https://pkg.go.dev/github.com/bluenviron/mediacommon/v2#pkg-index)

Definitions and functions shared between [gortsplib](https://github.com/bluenviron/gortsplib), [gohlslib](https://github.com/bluenviron/gohlslib), [gortmplib](https://github.com/bluenviron/gortmplib) and [MediaMTX](https://github.com/bluenviron/mediamtx), in particular:

* [Codec utilities](https://pkg.go.dev/github.com/bluenviron/mediacommon/v2/pkg/codecs)
* [Format utilities](https://pkg.go.dev/github.com/bluenviron/mediacommon/v2/pkg/formats)
* [Bit reader and writer](https://pkg.go.dev/github.com/bluenviron/mediacommon/v2/pkg/bits)
* [Rewindable reader](https://pkg.go.dev/github.com/bluenviron/mediacommon/v2/pkg/rewindablereader)

Go &ge; 1.24 is required.

## Specifications

|name|area|
|----|----|
|ISO 13818-2, Generic Coding of Moving Pictures and Associated Audio information, Part 2, Video|codecs / MPEG-1/2 Video|
|ISO 14496-2, Coding of audio-visual objects, Part 2, Visual|codecs / MPEG-4 Video|
|[ITU-T Rec. T-871, JPEG File Interchange Format](https://www.itu.int/rec/T-REC-T.871)|codecs / JPEG|
|[ITU-T Rec. H.264 (08/2021)](https://www.itu.int/rec/T-REC-H.264)|codecs / H264|
|[ITU-T Rec. H.265 (08/2021)](https://www.itu.int/rec/T-REC-H.265)|codecs / H265|
|[VP9 Bitstream & Decoding Process Specification v0.6](https://storage.googleapis.com/downloads.webmproject.org/docs/vp9/vp9-bitstream-specification-v0.6-20160331-draft.pdf)|codecs / VP9|
|[AV1 Bitstream & Decoding Process](https://aomediacodec.github.io/av1-spec/av1-spec.pdf)|codecs / AV1|
|[ITU-T Rec. G.711 (11/88)](https://www.itu.int/rec/T-REC-G.711)|codecs / G711|
|ISO 11172-3, Coding of moving pictures and associated audio|codecs / MPEG-1/2 Audio|
|ISO 13818-3, Generic Coding of Moving Pictures and Associated Audio information, Part 3, Audio|codecs / MPEG-1/2 Audio|
|ISO 14496-3, Coding of audio-visual objects, Part 3, Audio|codecs / MPEG-4 Audio|
|[RFC6716, Definition of the Opus Audio Codec](https://datatracker.ietf.org/doc/html/rfc6716)|codecs / Opus|
|[ATSC A/52, Digital Audio Compression (AC-3) (E-AC-3) Standard](https://www.atsc.org/wp-content/uploads/2021/04/A52-2018.pdf)|codecs / AC-3, E-AC-3|
|[ETSI TS 102 366, Digital Audio Compression (AC-3, Enhanced AC-3) Standard](https://www.etsi.org/deliver/etsi_ts/102300_102399/102366/01.04.01_60/ts_102366v010401p.pdf)|codecs / AC-3, E-AC-3|
|ISO 14496-1, Coding of audio-visual objects, Part 1, Systems|formats / fMP4|
|ISO 14496-12, Coding of audio-visual objects, Part 12, ISO base media file format|formats / fMP4|
|ISO 14496-14, Coding of audio-visual objects, Part 14, MP4 file format|formats / fMP4|
|ISO 14496-15, Coding of audio-visual objects, Part 15, Advanced Video Coding (AVC) file format|formats / fMP4 + H264 / H265|
|[VP9 Codec ISO Media File Format Binding](https://www.webmproject.org/vp9/mp4/)|formats / fMP4 + VP9|
|[AV1 Codec ISO Media File Format Binding](https://aomediacodec.github.io/av1-isobmff)|formats / fMP4 + AV1|
|[Opus in MP4/ISOBMFF](https://opus-codec.org/docs/opus_in_isobmff.html)|formats / fMP4 + Opus|
|ISO 23003-5, MPEG audio technologies, Part 5, Uncompressed audio in MPEG-4 file format|formats / fMP4 + LPCM|
|ISO 13818-1, Generic coding of moving pictures and associated audio information: Systems|formats / MPEG-TS|
|[ETSI TS Opus 0.1.3-draft](https://opus-codec.org/docs/ETSI_TS_opus-v0.1.3-draft.pdf)|formats / MPEG-TS + Opus|
|[MISB ST 1402, MPEG-2 Transport Stream for Class 1/Class 2 Motion Imagery, Audio and Metadata](https://nsgreg.nga.mil/doc/view?i=4273)|formats / MPEG-TS + KLV|
|[ETSI EN 300 743, Digital Video Broadcasting (DVB), Subtitling systems](https://www.etsi.org/deliver/etsi_en/300700_300799/300743/01.06.01_20/en_300743v010601a.pdf)|formats / MPEG-TS + DVB subtitles|
|[ETSI EN 300 468, Digital Video Broadcasting (DVB), Specification for Service Information (SI) in DVB systems](https://www.etsi.org/deliver/etsi_en/300400_300499/300468/01.17.01_20/en_300468v011701a.pdf)|formats / MPEG-TS + DVB subtitles|

## Related projects

* [MediaMTX](https://github.com/bluenviron/mediamtx)
* [gortsplib](https://github.com/bluenviron/gortsplib)
* [gohlslib](https://github.com/bluenviron/gohlslib)
* [gortmplib](https://github.com/bluenviron/gortmplib)
