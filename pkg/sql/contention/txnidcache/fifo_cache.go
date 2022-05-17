package txnidcache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/contention/contentionutils"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

var nodePool = &sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(459346)
		return &blockListNode{}
	},
}

type fifoCache struct {
	capacity contentionutils.CapacityLimiter

	mu struct {
		syncutil.RWMutex

		data map[uuid.UUID]roachpb.TransactionFingerprintID

		eviction blockList
	}
}

type blockListNode struct {
	*block
	next *blockListNode
}

type blockList struct {
	numNodes int
	head     *blockListNode
	tail     *blockListNode
}

func newFIFOCache(capacity contentionutils.CapacityLimiter) *fifoCache {
	__antithesis_instrumentation__.Notify(459347)
	c := &fifoCache{
		capacity: capacity,
	}

	c.mu.data = make(map[uuid.UUID]roachpb.TransactionFingerprintID)
	c.mu.eviction = blockList{}
	return c
}

func (c *fifoCache) add(b *block) {
	__antithesis_instrumentation__.Notify(459348)
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range b {
		__antithesis_instrumentation__.Notify(459350)
		if !b[i].Valid() {
			__antithesis_instrumentation__.Notify(459352)
			break
		} else {
			__antithesis_instrumentation__.Notify(459353)
		}
		__antithesis_instrumentation__.Notify(459351)

		c.mu.data[b[i].TxnID] = b[i].TxnFingerprintID
	}
	__antithesis_instrumentation__.Notify(459349)
	c.mu.eviction.addNode(b)
	c.maybeEvictLocked()
}

func (c *fifoCache) get(txnID uuid.UUID) (roachpb.TransactionFingerprintID, bool) {
	__antithesis_instrumentation__.Notify(459354)
	c.mu.RLock()
	defer c.mu.RUnlock()

	fingerprintID, found := c.mu.data[txnID]
	return fingerprintID, found
}

func (c *fifoCache) size() int64 {
	__antithesis_instrumentation__.Notify(459355)
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sizeLocked()
}

func (c *fifoCache) sizeLocked() int64 {
	__antithesis_instrumentation__.Notify(459356)
	return int64(c.mu.eviction.numNodes)*
		((entrySize*blockSize)+int64(unsafe.Sizeof(blockListNode{}))) +
		int64(len(c.mu.data))*entrySize
}

func (c *fifoCache) maybeEvictLocked() {
	__antithesis_instrumentation__.Notify(459357)
	for c.sizeLocked() > c.capacity() {
		__antithesis_instrumentation__.Notify(459358)
		node := c.mu.eviction.removeFront()
		if node == nil {
			__antithesis_instrumentation__.Notify(459360)
			return
		} else {
			__antithesis_instrumentation__.Notify(459361)
		}
		__antithesis_instrumentation__.Notify(459359)
		c.evictNodeLocked(node)
	}
}

func (c *fifoCache) evictNodeLocked(node *blockListNode) {
	__antithesis_instrumentation__.Notify(459362)
	for i := 0; i < blockSize; i++ {
		__antithesis_instrumentation__.Notify(459364)
		if !node.block[i].Valid() {
			__antithesis_instrumentation__.Notify(459366)
			break
		} else {
			__antithesis_instrumentation__.Notify(459367)
		}
		__antithesis_instrumentation__.Notify(459365)

		delete(c.mu.data, node.block[i].TxnID)
	}
	__antithesis_instrumentation__.Notify(459363)

	*node.block = block{}
	blockPool.Put(node.block)
	*node = blockListNode{}
	nodePool.Put(node)
}

func (e *blockList) addNode(b *block) {
	__antithesis_instrumentation__.Notify(459368)
	newNode := nodePool.Get().(*blockListNode)
	newNode.block = b
	if e.head == nil {
		__antithesis_instrumentation__.Notify(459370)
		e.head = newNode
	} else {
		__antithesis_instrumentation__.Notify(459371)
		e.tail.next = newNode
	}
	__antithesis_instrumentation__.Notify(459369)
	e.tail = newNode
	e.numNodes++
}

func (e *blockList) removeFront() *blockListNode {
	__antithesis_instrumentation__.Notify(459372)
	if e.head == nil {
		__antithesis_instrumentation__.Notify(459375)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(459376)
	}
	__antithesis_instrumentation__.Notify(459373)

	e.numNodes--
	removedBlock := e.head
	e.head = e.head.next
	if e.head == nil {
		__antithesis_instrumentation__.Notify(459377)
		e.tail = nil
	} else {
		__antithesis_instrumentation__.Notify(459378)
	}
	__antithesis_instrumentation__.Notify(459374)
	return removedBlock
}
