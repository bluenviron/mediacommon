// Package h265 contains utilities to work with the H265 codec.
package h265

const (
	// MaxNALUSize is the maximum size of a NALU.
	// With a 50 Mbps 2160p60 H265 video, the maximum NALU size does not seem to exceed 4 MiB.
	MaxNALUSize = 4 * 1024 * 1024

	// MaxNALUsPerAccessUnit is the maximum number of NALUs per access unit.
	MaxNALUsPerAccessUnit = 20
)
