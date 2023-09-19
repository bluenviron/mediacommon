# mediacommon

[![Test](https://github.com/bluenviron/mediacommon/workflows/test/badge.svg)](https://github.com/bluenviron/mediacommon/actions?query=workflow:test)
[![Lint](https://github.com/bluenviron/mediacommon/workflows/lint/badge.svg)](https://github.com/bluenviron/mediacommon/actions?query=workflow:lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/bluenviron/mediacommon)](https://goreportcard.com/report/github.com/bluenviron/mediacommon)
[![CodeCov](https://codecov.io/gh/bluenviron/mediacommon/branch/main/graph/badge.svg)](https://app.codecov.io/gh/bluenviron/mediacommon/branch/main)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/bluenviron/mediacommon)](https://pkg.go.dev/github.com/bluenviron/mediacommon#pkg-index)

Definitions and functions shared between [gortsplib](https://github.com/bluenviron/gortsplib), [gohlslib](https://github.com/bluenviron/gohlslib) and [MediaMTX](https://github.com/bluenviron/mediamtx), in particular:

* [Codec utilities](https://pkg.go.dev/github.com/bluenviron/mediacommon/pkg/codecs)
* [Format utilities](https://pkg.go.dev/github.com/bluenviron/mediacommon/pkg/formats)
* [Bit reader and writer](https://pkg.go.dev/github.com/bluenviron/mediacommon/pkg/bits)

## Specifications

|name|area|
|----|----|
|ISO 13818-2, Generic Coding of Moving Pictures and Associated Audio information, Part 2, Video|MPEG-1/2 Video codec|
|ISO 14496-2, Coding of audio-visual objects, Part 2, Visual|MPEG-4 Video codec|
|[ITU-T Rec. T-871, JPEG File Interchange Format](https://www.itu.int/rec/dologin_pub.asp?lang=e&id=T-REC-T.871-201105-I!!PDF-E&type=items)|JPEG codec|
|[ITU-T Rec. H.264 (08/2021)](https://www.itu.int/rec/dologin_pub.asp?lang=e&id=T-REC-H.264-202108-I!!PDF-E&type=items)|H264 codec|
|[ITU-T Rec. H.265 (08/2021)](https://www.itu.int/rec/dologin_pub.asp?lang=e&id=T-REC-H.265-202108-I!!PDF-E&type=items)|H265 codec|
|[VP9 Bitstream & Decoding Process Specification v0.6](https://storage.googleapis.com/downloads.webmproject.org/docs/vp9/vp9-bitstream-specification-v0.6-20160331-draft.pdf)|VP9 codec|
|[VP9 Codec ISO Media File Format Binding](https://www.webmproject.org/vp9/mp4/)|VP9 inside MP4|
|[AV1 Bitstream & Decoding Process](https://aomediacodec.github.io/av1-spec/av1-spec.pdf)|AV1 codec|
|[AV1 Codec ISO Media File Format Binding](https://aomediacodec.github.io/av1-isobmff)|AV1 inside MP4|
|ISO 11172-3, Coding of moving pictures and associated audio|MPEG-1/2 Audio codec|
|ISO 13818-3, Generic Coding of Moving Pictures and Associated Audio information, Part 3, Audio|MPEG-1/2 Audio codec|
|ISO 14496-3, Coding of audio-visual objects, Part 3, Audio|MPEG-4 Audio codec|
|[RFC6716, Definition of the Opus Audio Codec](https://datatracker.ietf.org/doc/html/rfc6716)|Opus codec|
|[Opus in MP4/ISOBMFF](https://opus-codec.org/docs/opus_in_isobmff.html)|Opus inside MP4|
|[ATSC Standard: Digital Audio Compression (AC-3, E-AC-3)](http://www.atsc.org/wp-content/uploads/2015/03/A52-201212-17.pdf)|AC-3 codec|
|[ETSI TS 102 366](https://www.etsi.org/deliver/etsi_ts/102300_102399/102366/01.04.01_60/ts_102366v010401p.pdf)|AC-3 inside MP4|
|ISO 14496-1, Coding of audio-visual objects, Part 1, Systems|MP4 format|
|ISO 14496-12, Coding of audio-visual objects, Part 12, ISO base media file format|MP4 format|
|ISO 14496-14, Coding of audio-visual objects, Part 14, MP4 file format|MP4 format|

## Related projects

* [MediaMTX](https://github.com/bluenviron/mediamtx)
* [gortsplib](https://github.com/bluenviron/gortsplib)
* [gohlslib](https://github.com/bluenviron/gohlslib)
