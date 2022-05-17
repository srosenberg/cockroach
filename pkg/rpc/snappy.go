package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"sync"

	"github.com/golang/snappy"
	"google.golang.org/grpc/encoding"
)

var snappyWriterPool sync.Pool
var snappyReaderPool sync.Pool

type snappyWriter struct {
	*snappy.Writer
}

func (w *snappyWriter) Close() error {
	__antithesis_instrumentation__.Notify(185537)
	defer snappyWriterPool.Put(w)
	return w.Writer.Close()
}

type snappyReader struct {
	*snappy.Reader
}

func (r *snappyReader) Read(p []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(185538)
	n, err = r.Reader.Read(p)
	if err == io.EOF {
		__antithesis_instrumentation__.Notify(185540)
		snappyReaderPool.Put(r)
	} else {
		__antithesis_instrumentation__.Notify(185541)
	}
	__antithesis_instrumentation__.Notify(185539)
	return n, err
}

type snappyCompressor struct {
}

func (snappyCompressor) Name() string {
	__antithesis_instrumentation__.Notify(185542)
	return "snappy"
}

func (snappyCompressor) Compress(w io.Writer) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(185543)
	sw, ok := snappyWriterPool.Get().(*snappyWriter)
	if !ok {
		__antithesis_instrumentation__.Notify(185545)
		sw = &snappyWriter{snappy.NewBufferedWriter(w)}
	} else {
		__antithesis_instrumentation__.Notify(185546)
		sw.Reset(w)
	}
	__antithesis_instrumentation__.Notify(185544)
	return sw, nil
}

func (snappyCompressor) Decompress(r io.Reader) (io.Reader, error) {
	__antithesis_instrumentation__.Notify(185547)
	sr, ok := snappyReaderPool.Get().(*snappyReader)
	if !ok {
		__antithesis_instrumentation__.Notify(185549)
		sr = &snappyReader{snappy.NewReader(r)}
	} else {
		__antithesis_instrumentation__.Notify(185550)
		sr.Reset(r)
	}
	__antithesis_instrumentation__.Notify(185548)
	return sr, nil
}

func init() {
	encoding.RegisterCompressor(snappyCompressor{})
}
