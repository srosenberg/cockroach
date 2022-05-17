package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

type writeBuffer struct {
	_ util.NoCopy

	wrapped bytes.Buffer
	err     error

	putbuf [64]byte

	textFormatter   *tree.FmtCtx
	simpleFormatter *tree.FmtCtx

	bytecount *metric.Counter
}

func newWriteBuffer(bytecount *metric.Counter) *writeBuffer {
	__antithesis_instrumentation__.Notify(562032)
	b := new(writeBuffer)
	b.init(bytecount)
	return b
}

func (b *writeBuffer) init(bytecount *metric.Counter) {
	b.bytecount = bytecount
	b.textFormatter = tree.NewFmtCtx(tree.FmtPgwireText)
	b.simpleFormatter = tree.NewFmtCtx(tree.FmtSimple)
}

func (b *writeBuffer) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(562033)
	b.write(p)
	return len(p), b.err
}

func (b *writeBuffer) writeByte(c byte) {
	__antithesis_instrumentation__.Notify(562034)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562035)
		b.err = b.wrapped.WriteByte(c)
	} else {
		__antithesis_instrumentation__.Notify(562036)
	}
}

func (b *writeBuffer) write(p []byte) {
	__antithesis_instrumentation__.Notify(562037)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562038)
		_, b.err = b.wrapped.Write(p)
	} else {
		__antithesis_instrumentation__.Notify(562039)
	}
}

func (b *writeBuffer) writeString(s string) {
	__antithesis_instrumentation__.Notify(562040)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562041)
		_, b.err = b.wrapped.WriteString(s)
	} else {
		__antithesis_instrumentation__.Notify(562042)
	}
}

func (b *writeBuffer) Len() int {
	__antithesis_instrumentation__.Notify(562043)
	return b.wrapped.Len()
}

func (b *writeBuffer) nullTerminate() {
	__antithesis_instrumentation__.Notify(562044)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562045)
		b.err = b.wrapped.WriteByte(0)
	} else {
		__antithesis_instrumentation__.Notify(562046)
	}
}

func (b *writeBuffer) writeFromFmtCtx(fmtCtx *tree.FmtCtx) {
	__antithesis_instrumentation__.Notify(562047)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562048)
		b.putInt32(int32(fmtCtx.Buffer.Len()))

		_, b.err = fmtCtx.Buffer.WriteTo(&b.wrapped)
	} else {
		__antithesis_instrumentation__.Notify(562049)
	}
}

func (b *writeBuffer) writeLengthPrefixedString(s string) {
	__antithesis_instrumentation__.Notify(562050)
	b.putInt32(int32(len(s)))
	b.writeString(s)
}

func (b *writeBuffer) writeLengthPrefixedDatum(d tree.Datum) {
	__antithesis_instrumentation__.Notify(562051)
	b.simpleFormatter.FormatNode(d)
	b.writeFromFmtCtx(b.simpleFormatter)
}

func (b *writeBuffer) writeTerminatedString(s string) {
	__antithesis_instrumentation__.Notify(562052)
	b.writeString(s)
	b.nullTerminate()
}

func (b *writeBuffer) putInt16(v int16) {
	__antithesis_instrumentation__.Notify(562053)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562054)
		binary.BigEndian.PutUint16(b.putbuf[:], uint16(v))
		_, b.err = b.wrapped.Write(b.putbuf[:2])
	} else {
		__antithesis_instrumentation__.Notify(562055)
	}
}

func (b *writeBuffer) putInt32(v int32) {
	__antithesis_instrumentation__.Notify(562056)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562057)
		binary.BigEndian.PutUint32(b.putbuf[:], uint32(v))
		_, b.err = b.wrapped.Write(b.putbuf[:4])
	} else {
		__antithesis_instrumentation__.Notify(562058)
	}
}

func (b *writeBuffer) putInt64(v int64) {
	__antithesis_instrumentation__.Notify(562059)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562060)
		binary.BigEndian.PutUint64(b.putbuf[:], uint64(v))
		_, b.err = b.wrapped.Write(b.putbuf[:8])
	} else {
		__antithesis_instrumentation__.Notify(562061)
	}
}

func (b *writeBuffer) putInt32AtIndex(index int, v int32) {
	__antithesis_instrumentation__.Notify(562062)
	binary.BigEndian.PutUint32(b.wrapped.Bytes()[index:index+4], uint32(v))
}

func (b *writeBuffer) putErrFieldMsg(field pgwirebase.ServerErrFieldType) {
	__antithesis_instrumentation__.Notify(562063)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562064)
		b.err = b.wrapped.WriteByte(byte(field))
	} else {
		__antithesis_instrumentation__.Notify(562065)
	}
}

func (b *writeBuffer) reset() {
	__antithesis_instrumentation__.Notify(562066)
	b.wrapped.Reset()
	b.err = nil
}

func (b *writeBuffer) initMsg(typ pgwirebase.ServerMessageType) {
	__antithesis_instrumentation__.Notify(562067)
	b.reset()
	b.putbuf[0] = byte(typ)
	_, b.err = b.wrapped.Write(b.putbuf[:5])
}

func (b *writeBuffer) finishMsg(w io.Writer) error {
	__antithesis_instrumentation__.Notify(562068)
	defer b.reset()
	if b.err != nil {
		__antithesis_instrumentation__.Notify(562070)
		return b.err
	} else {
		__antithesis_instrumentation__.Notify(562071)
	}
	__antithesis_instrumentation__.Notify(562069)
	bytes := b.wrapped.Bytes()
	binary.BigEndian.PutUint32(bytes[1:5], uint32(b.wrapped.Len()-1))
	n, err := w.Write(bytes)
	b.bytecount.Inc(int64(n))
	return err
}

func (b *writeBuffer) setError(err error) {
	__antithesis_instrumentation__.Notify(562072)
	if b.err == nil {
		__antithesis_instrumentation__.Notify(562073)
		b.err = err
	} else {
		__antithesis_instrumentation__.Notify(562074)
	}
}
