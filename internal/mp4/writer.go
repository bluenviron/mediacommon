// Package mp4 contains internal MP4 utilities.
package mp4

import (
	"io"

	amp4 "github.com/abema/go-mp4"
)

// Writer is a wrapper around abema/go-mp4.Writer.
type Writer struct {
	W io.WriteSeeker

	mw *amp4.Writer
}

// Initialize initializes a Writer.
func (w *Writer) Initialize() {
	w.mw = amp4.NewWriter(w.W)
}

// WriteBoxStart writes a box start.
func (w *Writer) WriteBoxStart(box amp4.IImmutableBox) (int, error) {
	bi := &amp4.BoxInfo{
		Type: box.GetType(),
	}
	var err error
	bi, err = w.mw.StartBox(bi)
	if err != nil {
		return 0, err
	}

	_, err = amp4.Marshal(w.mw, box, amp4.Context{})
	if err != nil {
		return 0, err
	}

	return int(bi.Offset), nil
}

// WriteBoxEnd writes a box end.
func (w *Writer) WriteBoxEnd() error {
	_, err := w.mw.EndBox()
	return err
}

// WriteBox writes a box.
func (w *Writer) WriteBox(box amp4.IImmutableBox) (int, error) {
	off, err := w.WriteBoxStart(box)
	if err != nil {
		return 0, err
	}

	err = w.WriteBoxEnd()
	if err != nil {
		return 0, err
	}

	return off, nil
}

// RewriteBox rewrites a box.
func (w *Writer) RewriteBox(off int, box amp4.IImmutableBox) error {
	prevOff, err := w.mw.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	_, err = w.mw.Seek(int64(off), io.SeekStart)
	if err != nil {
		return err
	}

	_, err = w.WriteBoxStart(box)
	if err != nil {
		return err
	}

	err = w.WriteBoxEnd()
	if err != nil {
		return err
	}

	_, err = w.mw.Seek(prevOff, io.SeekStart)
	if err != nil {
		return err
	}

	return nil
}
