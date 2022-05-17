package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"compress/gzip"
	"net/http"
	"sync"
)

type gzipResponseWriter struct {
	gz gzip.Writer
	http.ResponseWriter
}

var gzipResponseWriterPool sync.Pool

func newGzipResponseWriter(rw http.ResponseWriter) *gzipResponseWriter {
	__antithesis_instrumentation__.Notify(193488)
	var w *gzipResponseWriter
	if wI := gzipResponseWriterPool.Get(); wI == nil {
		__antithesis_instrumentation__.Notify(193490)
		w = new(gzipResponseWriter)
	} else {
		__antithesis_instrumentation__.Notify(193491)
		w = wI.(*gzipResponseWriter)
	}
	__antithesis_instrumentation__.Notify(193489)
	w.Reset(rw)
	return w
}

func (w *gzipResponseWriter) Reset(rw http.ResponseWriter) {
	__antithesis_instrumentation__.Notify(193492)
	w.gz.Reset(rw)
	w.ResponseWriter = rw
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	__antithesis_instrumentation__.Notify(193493)

	if w.Header().Get("Content-Type") == "" {
		__antithesis_instrumentation__.Notify(193495)
		w.Header().Set("Content-Type", http.DetectContentType(b))
	} else {
		__antithesis_instrumentation__.Notify(193496)
	}
	__antithesis_instrumentation__.Notify(193494)
	return w.gz.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	__antithesis_instrumentation__.Notify(193497)

	if err := w.gz.Flush(); err == nil {
		__antithesis_instrumentation__.Notify(193498)

		if f, ok := w.ResponseWriter.(http.Flusher); ok {
			__antithesis_instrumentation__.Notify(193499)
			f.Flush()
		} else {
			__antithesis_instrumentation__.Notify(193500)
		}
	} else {
		__antithesis_instrumentation__.Notify(193501)
	}
}

func (w *gzipResponseWriter) Close() error {
	__antithesis_instrumentation__.Notify(193502)
	err := w.gz.Close()
	w.Reset(nil)
	gzipResponseWriterPool.Put(w)
	return err
}
