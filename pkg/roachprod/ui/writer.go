package ui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Writer struct {
	buf       bytes.Buffer
	lineCount int
}

func (w *Writer) Flush(out io.Writer) error {
	__antithesis_instrumentation__.Notify(182542)
	if len(w.buf.Bytes()) == 0 {
		__antithesis_instrumentation__.Notify(182545)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(182546)
	}
	__antithesis_instrumentation__.Notify(182543)
	w.clearLines(out)

	for _, b := range w.buf.Bytes() {
		__antithesis_instrumentation__.Notify(182547)
		if b == '\n' {
			__antithesis_instrumentation__.Notify(182548)
			w.lineCount++
		} else {
			__antithesis_instrumentation__.Notify(182549)
		}
	}
	__antithesis_instrumentation__.Notify(182544)
	_, err := out.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}

func (w *Writer) Write(b []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(182550)
	return w.buf.Write(b)
}

func (w *Writer) clearLines(out io.Writer) {
	__antithesis_instrumentation__.Notify(182551)
	fmt.Fprint(out, strings.Repeat("\033[1A\033[2K\r", w.lineCount))
	w.lineCount = 0
}
