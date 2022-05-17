package interceptor

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgproto3/v2"
)

var errInvalidRead = errors.New("invalid read in chunkReader")

var _ pgproto3.ChunkReader = &chunkReader{}

type chunkReader struct {
	msg []byte
	pos int
}

func newChunkReader(msg []byte) pgproto3.ChunkReader {
	__antithesis_instrumentation__.Notify(21751)
	return &chunkReader{msg: msg}
}

func (cr *chunkReader) Next(n int) (buf []byte, err error) {
	__antithesis_instrumentation__.Notify(21752)

	if n == 0 {
		__antithesis_instrumentation__.Notify(21756)
		return []byte{}, nil
	} else {
		__antithesis_instrumentation__.Notify(21757)
	}
	__antithesis_instrumentation__.Notify(21753)
	if cr.pos == len(cr.msg) {
		__antithesis_instrumentation__.Notify(21758)
		return nil, io.EOF
	} else {
		__antithesis_instrumentation__.Notify(21759)
	}
	__antithesis_instrumentation__.Notify(21754)
	if cr.pos+n > len(cr.msg) {
		__antithesis_instrumentation__.Notify(21760)
		return nil, errInvalidRead
	} else {
		__antithesis_instrumentation__.Notify(21761)
	}
	__antithesis_instrumentation__.Notify(21755)
	buf = cr.msg[cr.pos : cr.pos+n]
	cr.pos += n
	return buf, nil
}
