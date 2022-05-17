package contentionutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type CapacityLimiter func() int64

type ConcurrentBufferGuard struct {
	flushSyncLock syncutil.RWMutex
	flushDone     sync.Cond

	limiter          CapacityLimiter
	onBufferFullSync onBufferFullHandler

	atomicIdx int64
}

type onBufferFullHandler func(currentWriterIndex int64)

type bufferWriteOp func(writerIdx int64)

func NewConcurrentBufferGuard(
	limiter CapacityLimiter, fullHandler onBufferFullHandler,
) *ConcurrentBufferGuard {
	__antithesis_instrumentation__.Notify(459031)
	writeBuffer := &ConcurrentBufferGuard{
		limiter:          limiter,
		onBufferFullSync: fullHandler,
	}
	writeBuffer.flushDone.L = writeBuffer.flushSyncLock.RLocker()
	return writeBuffer
}

func (c *ConcurrentBufferGuard) AtomicWrite(op bufferWriteOp) {
	__antithesis_instrumentation__.Notify(459032)
	size := c.limiter()
	c.flushSyncLock.RLock()
	defer c.flushSyncLock.RUnlock()
	for {
		__antithesis_instrumentation__.Notify(459033)
		reservedIdx := c.reserveMsgBlockIndex()
		if reservedIdx < size {
			__antithesis_instrumentation__.Notify(459034)
			op(reservedIdx)
			return
		} else {
			__antithesis_instrumentation__.Notify(459035)
			if reservedIdx == size {
				__antithesis_instrumentation__.Notify(459036)
				c.syncRLocked()
			} else {
				__antithesis_instrumentation__.Notify(459037)
				c.flushDone.Wait()
			}
		}
	}
}

func (c *ConcurrentBufferGuard) ForceSync() {
	__antithesis_instrumentation__.Notify(459038)
	c.flushSyncLock.Lock()
	c.syncLocked()
	c.flushSyncLock.Unlock()
}

func (c *ConcurrentBufferGuard) syncRLocked() {
	__antithesis_instrumentation__.Notify(459039)

	c.flushSyncLock.RUnlock()
	defer c.flushSyncLock.RLock()
	c.flushSyncLock.Lock()
	defer c.flushSyncLock.Unlock()
	c.syncLocked()
}

func (c *ConcurrentBufferGuard) syncLocked() {
	__antithesis_instrumentation__.Notify(459040)
	c.onBufferFullSync(c.currentWriterIndex())
	c.flushDone.Broadcast()
	c.rewindBuffer()
}

func (c *ConcurrentBufferGuard) rewindBuffer() {
	__antithesis_instrumentation__.Notify(459041)
	atomic.StoreInt64(&c.atomicIdx, 0)
}

func (c *ConcurrentBufferGuard) reserveMsgBlockIndex() int64 {
	__antithesis_instrumentation__.Notify(459042)
	return atomic.AddInt64(&c.atomicIdx, 1) - 1
}

func (c *ConcurrentBufferGuard) currentWriterIndex() int64 {
	__antithesis_instrumentation__.Notify(459043)
	sizeLimit := c.limiter()
	if curIdx := atomic.LoadInt64(&c.atomicIdx); curIdx < sizeLimit {
		__antithesis_instrumentation__.Notify(459045)
		return curIdx
	} else {
		__antithesis_instrumentation__.Notify(459046)
	}
	__antithesis_instrumentation__.Notify(459044)
	return sizeLimit
}
