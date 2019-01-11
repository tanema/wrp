package cwriter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	isatty "github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

// ESC is the ASCII code for escape character
const ESC = 27

// ErrNotATTY is an error returned when not running in a ATTY terminal
var ErrNotATTY = errors.New("not a terminal")

var (
	cursorUp           = fmt.Sprintf("%c[%dA", ESC, 1)
	clearLine          = fmt.Sprintf("%c[2K\r", ESC)
	clearCursorAndLine = cursorUp + clearLine
)

// Writer is a buffered the writer that updates the terminal.
// The contents of writer will be flushed when Flush is called.
type Writer struct {
	out        io.Writer
	buf        bytes.Buffer
	isTerminal bool
	fd         int
	lineCount  int
}

// New returns a new Writer with defaults
func New(out io.Writer) *Writer {
	w := &Writer{out: out}
	if f, ok := out.(*os.File); ok {
		fd := f.Fd()
		w.isTerminal = isatty.IsTerminal(fd)
		w.fd = int(fd)
	}
	return w
}

// Flush flushes the underlying buffer
func (w *Writer) Flush(lineCount int) error {
	if err := w.clearLines(); err != nil {
		return err
	}
	w.lineCount = lineCount
	// WriteTo takes care of w.buf.Reset
	if _, err := w.buf.WriteTo(w.out); err == nil {
		return err
	}
	return nil
}

// Write appends the contents of p to the underlying buffer
func (w *Writer) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

// WriteString writes string to the underlying buffer
func (w *Writer) WriteString(s string) (n int, err error) {
	return w.buf.WriteString(s)
}

// ReadFrom reads from the provided io.Reader and writes to the underlying buffer.
func (w *Writer) ReadFrom(r io.Reader) (n int64, err error) {
	return w.buf.ReadFrom(r)
}

// GetWidth gets the width of the terminal in the running context
func (w *Writer) GetWidth() (int, error) {
	if w.isTerminal {
		tw, _, err := terminal.GetSize(w.fd)
		return tw, err
	}
	return -1, ErrNotATTY
}
