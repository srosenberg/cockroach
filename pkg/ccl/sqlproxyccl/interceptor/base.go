package interceptor

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

const pgHeaderSizeBytes = 5

const defaultBufferSize = 1 << 13

var ErrProtocolError = errors.New("protocol error")

type pgInterceptor struct {
	_ util.NoCopy

	src io.Reader

	buf []byte

	readPos, writePos int
}

func newPgInterceptor(src io.Reader, bufSize int) *pgInterceptor {
	__antithesis_instrumentation__.Notify(21691)

	if bufSize < pgHeaderSizeBytes {
		__antithesis_instrumentation__.Notify(21693)
		bufSize = defaultBufferSize
	} else {
		__antithesis_instrumentation__.Notify(21694)
	}
	__antithesis_instrumentation__.Notify(21692)
	return &pgInterceptor{
		src: src,
		buf: make([]byte, bufSize),
	}
}

func (p *pgInterceptor) PeekMsg() (typ byte, size int, err error) {
	__antithesis_instrumentation__.Notify(21695)
	if err := p.ensureNextNBytes(pgHeaderSizeBytes); err != nil {
		__antithesis_instrumentation__.Notify(21698)

		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(21699)
	}
	__antithesis_instrumentation__.Notify(21696)

	typ = p.buf[p.readPos]
	size = int(binary.BigEndian.Uint32(p.buf[p.readPos+1:]))

	if size < 4 || func() bool {
		__antithesis_instrumentation__.Notify(21700)
		return size >= math.MaxInt32 == true
	}() == true {
		__antithesis_instrumentation__.Notify(21701)
		return 0, 0, ErrProtocolError
	} else {
		__antithesis_instrumentation__.Notify(21702)
	}
	__antithesis_instrumentation__.Notify(21697)

	return typ, size + 1, nil
}

func (p *pgInterceptor) ReadMsg() (msg []byte, err error) {
	__antithesis_instrumentation__.Notify(21703)

	_, size, err := p.PeekMsg()
	if err != nil {
		__antithesis_instrumentation__.Notify(21707)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21708)
	}
	__antithesis_instrumentation__.Notify(21704)

	if size <= len(p.buf) {
		__antithesis_instrumentation__.Notify(21709)
		if err := p.ensureNextNBytes(size); err != nil {
			__antithesis_instrumentation__.Notify(21711)

			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(21712)
		}
		__antithesis_instrumentation__.Notify(21710)

		retBuf := p.buf[p.readPos : p.readPos+size]
		p.readPos += size
		return retBuf, nil
	} else {
		__antithesis_instrumentation__.Notify(21713)
	}
	__antithesis_instrumentation__.Notify(21705)

	msg = make([]byte, size)

	n := copy(msg, p.buf[p.readPos:p.writePos])
	p.readPos += n

	if _, err := io.ReadFull(p.src, msg[n:]); err != nil {
		__antithesis_instrumentation__.Notify(21714)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21715)
	}
	__antithesis_instrumentation__.Notify(21706)

	return msg, nil
}

func (p *pgInterceptor) ForwardMsg(dst io.Writer) (n int, err error) {
	__antithesis_instrumentation__.Notify(21716)

	_, size, err := p.PeekMsg()
	if err != nil {
		__antithesis_instrumentation__.Notify(21722)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(21723)
	}
	__antithesis_instrumentation__.Notify(21717)

	startPos := p.readPos
	endPos := startPos + size
	remainingBytes := 0
	if endPos > p.writePos {
		__antithesis_instrumentation__.Notify(21724)
		remainingBytes = endPos - p.writePos
		endPos = p.writePos
	} else {
		__antithesis_instrumentation__.Notify(21725)
	}
	__antithesis_instrumentation__.Notify(21718)
	p.readPos = endPos

	n, err = dst.Write(p.buf[startPos:endPos])
	if err != nil {
		__antithesis_instrumentation__.Notify(21726)
		return n, err
	} else {
		__antithesis_instrumentation__.Notify(21727)
	}
	__antithesis_instrumentation__.Notify(21719)

	if n < endPos-startPos {
		__antithesis_instrumentation__.Notify(21728)
		return n, io.ErrShortWrite
	} else {
		__antithesis_instrumentation__.Notify(21729)
	}
	__antithesis_instrumentation__.Notify(21720)

	if remainingBytes > 0 {
		__antithesis_instrumentation__.Notify(21730)
		m, err := io.CopyN(dst, p.src, int64(remainingBytes))
		n += int(m)
		if err != nil {
			__antithesis_instrumentation__.Notify(21732)
			return n, err
		} else {
			__antithesis_instrumentation__.Notify(21733)
		}
		__antithesis_instrumentation__.Notify(21731)

		if int(m) < remainingBytes {
			__antithesis_instrumentation__.Notify(21734)
			return n, io.ErrShortWrite
		} else {
			__antithesis_instrumentation__.Notify(21735)
		}
	} else {
		__antithesis_instrumentation__.Notify(21736)
	}
	__antithesis_instrumentation__.Notify(21721)
	return n, nil
}

func (p *pgInterceptor) readSize() int {
	__antithesis_instrumentation__.Notify(21737)
	return p.writePos - p.readPos
}

func (p *pgInterceptor) writeSize() int {
	__antithesis_instrumentation__.Notify(21738)
	return len(p.buf) - p.writePos
}

func (p *pgInterceptor) ensureNextNBytes(n int) error {
	__antithesis_instrumentation__.Notify(21739)
	if n < 0 || func() bool {
		__antithesis_instrumentation__.Notify(21743)
		return n > len(p.buf) == true
	}() == true {
		__antithesis_instrumentation__.Notify(21744)
		return errors.AssertionFailedf(
			"invalid number of bytes %d for buffer size %d", n, len(p.buf))
	} else {
		__antithesis_instrumentation__.Notify(21745)
	}
	__antithesis_instrumentation__.Notify(21740)

	if p.readSize() >= n {
		__antithesis_instrumentation__.Notify(21746)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(21747)
	}
	__antithesis_instrumentation__.Notify(21741)

	minReadCount := n - p.readSize()
	if p.writeSize() < minReadCount {
		__antithesis_instrumentation__.Notify(21748)
		p.writePos = copy(p.buf, p.buf[p.readPos:p.writePos])
		p.readPos = 0
	} else {
		__antithesis_instrumentation__.Notify(21749)
	}
	__antithesis_instrumentation__.Notify(21742)

	c, err := io.ReadAtLeast(p.src, p.buf[p.writePos:], minReadCount)
	p.writePos += c
	return err
}

var _ io.Writer = &errWriter{}

type errWriter struct{}

func (w *errWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(21750)
	return 0, errors.AssertionFailedf("unexpected Write call")
}
