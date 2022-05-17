package txnidcache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/contention/contentionutils"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
)

const blockSize = 168

type block [blockSize]contentionpb.ResolvedTxnID

var blockPool = &sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(459338)
		return &block{}
	},
}

type concurrentWriteBuffer struct {
	guard struct {
		*contentionutils.ConcurrentBufferGuard

		block *block
	}

	sink blockSink
}

var _ Writer = &concurrentWriteBuffer{}

func newConcurrentWriteBuffer(sink blockSink) *concurrentWriteBuffer {
	__antithesis_instrumentation__.Notify(459339)
	writeBuffer := &concurrentWriteBuffer{
		sink: sink,
	}

	writeBuffer.guard.block = blockPool.Get().(*block)
	writeBuffer.guard.ConcurrentBufferGuard = contentionutils.NewConcurrentBufferGuard(
		func() int64 {
			__antithesis_instrumentation__.Notify(459341)
			return blockSize
		},
		func(_ int64) {
			__antithesis_instrumentation__.Notify(459342)
			writeBuffer.sink.push(writeBuffer.guard.block)

			writeBuffer.guard.block = blockPool.Get().(*block)
		})
	__antithesis_instrumentation__.Notify(459340)

	return writeBuffer
}

func (c *concurrentWriteBuffer) Record(resolvedTxnID contentionpb.ResolvedTxnID) {
	__antithesis_instrumentation__.Notify(459343)
	c.guard.AtomicWrite(func(writerIdx int64) {
		__antithesis_instrumentation__.Notify(459344)
		c.guard.block[writerIdx] = resolvedTxnID
	})
}

func (c *concurrentWriteBuffer) DrainWriteBuffer() {
	__antithesis_instrumentation__.Notify(459345)
	c.guard.ForceSync()
}
