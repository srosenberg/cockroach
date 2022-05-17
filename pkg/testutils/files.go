package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ReadAllFiles(pattern string) {
	__antithesis_instrumentation__.Notify(644225)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		__antithesis_instrumentation__.Notify(644227)
		return
	} else {
		__antithesis_instrumentation__.Notify(644228)
	}
	__antithesis_instrumentation__.Notify(644226)
	for _, m := range matches {
		__antithesis_instrumentation__.Notify(644229)
		f, err := os.Open(m)
		if err != nil {
			__antithesis_instrumentation__.Notify(644231)
			continue
		} else {
			__antithesis_instrumentation__.Notify(644232)
		}
		__antithesis_instrumentation__.Notify(644230)
		_, _ = io.Copy(ioutil.Discard, bufio.NewReader(f))
		f.Close()
	}
}
