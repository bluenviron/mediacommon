//nolint:gochecknoinits,revive,gocritic
package fmp4

import (
	gomp4 "github.com/abema/go-mp4"
)

func BoxTypeMp4v() gomp4.BoxType { return gomp4.StrToBoxType("mp4v") }

func init() {
	gomp4.AddAnyTypeBoxDef(&gomp4.VisualSampleEntry{}, BoxTypeMp4v())
}
