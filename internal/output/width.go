package output

import (
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

// ttySize returns the controlling terminal size. Prefer /dev/tty so size is
// still available when stdout is a pipe (as under watch(1)).
func ttySize() (width, height int, ok bool) {
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil && (w > 0 || h > 0) {
		return w, h, true
	}
	if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		defer f.Close()
		if w, h, err := term.GetSize(int(f.Fd())); err == nil && (w > 0 || h > 0) {
			return w, h, true
		}
	}
	w, h := 0, 0
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil {
			w = n
		}
	}
	if rows := os.Getenv("LINES"); rows != "" {
		if n, err := strconv.Atoi(rows); err == nil {
			h = n
		}
	}
	return w, h, w > 0 || h > 0
}

// fitTable caps row width to the live terminal so box tables do not wrap.
func fitTable(t table.Writer) {
	if w, _, ok := ttySize(); ok && w > 0 {
		t.SetAllowedRowLength(w)
	}
}

// OutputHeight returns the terminal height, or 0 if unknown.
func OutputHeight() int {
	if _, h, ok := ttySize(); ok {
		return h
	}
	return 0
}
