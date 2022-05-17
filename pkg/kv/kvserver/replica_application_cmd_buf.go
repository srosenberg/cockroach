package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/apply"
)

const replicatedCmdBufNodeSize = 8

const replicatedCmdBufSliceFreeListSize = 3

type replicatedCmdBuf struct {
	len        int32
	head, tail *replicatedCmdBufNode
	free       replicatedCmdBufSliceFreeList
}

type replicatedCmdBufNode struct {
	len  int32
	buf  [replicatedCmdBufNodeSize]replicatedCmd
	next *replicatedCmdBufNode
}

var replicatedCmdBufBufNodeSyncPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(115000); return new(replicatedCmdBufNode) },
}

func (buf *replicatedCmdBuf) allocate() *replicatedCmd {
	__antithesis_instrumentation__.Notify(115001)
	if buf.tail == nil {
		__antithesis_instrumentation__.Notify(115004)
		n := replicatedCmdBufBufNodeSyncPool.Get().(*replicatedCmdBufNode)
		buf.head, buf.tail = n, n
	} else {
		__antithesis_instrumentation__.Notify(115005)
	}
	__antithesis_instrumentation__.Notify(115002)
	if buf.tail.len == replicatedCmdBufNodeSize {
		__antithesis_instrumentation__.Notify(115006)
		newTail := replicatedCmdBufBufNodeSyncPool.Get().(*replicatedCmdBufNode)
		buf.tail.next = newTail
		buf.tail = newTail
	} else {
		__antithesis_instrumentation__.Notify(115007)
	}
	__antithesis_instrumentation__.Notify(115003)
	ret := &buf.tail.buf[buf.tail.len]
	buf.tail.len++
	buf.len++
	return ret
}

func (buf *replicatedCmdBuf) clear() {
	__antithesis_instrumentation__.Notify(115008)
	for buf.head != nil {
		__antithesis_instrumentation__.Notify(115010)
		buf.len -= buf.head.len
		oldHead := buf.head
		newHead := oldHead.next
		buf.head = newHead
		*oldHead = replicatedCmdBufNode{}
		replicatedCmdBufBufNodeSyncPool.Put(oldHead)
	}
	__antithesis_instrumentation__.Notify(115009)
	*buf = replicatedCmdBuf{}
}

func (buf *replicatedCmdBuf) newIter() *replicatedCmdBufSlice {
	__antithesis_instrumentation__.Notify(115011)
	return buf.free.get()
}

type replicatedCmdBufPtr struct {
	idx  int32
	buf  *replicatedCmdBuf
	node *replicatedCmdBufNode
}

func (ptr *replicatedCmdBufPtr) valid() bool {
	__antithesis_instrumentation__.Notify(115012)
	return ptr.idx < ptr.buf.len
}

func (ptr *replicatedCmdBufPtr) cur() *replicatedCmd {
	__antithesis_instrumentation__.Notify(115013)
	return &ptr.node.buf[ptr.idx%replicatedCmdBufNodeSize]
}

func (ptr *replicatedCmdBufPtr) next() {
	__antithesis_instrumentation__.Notify(115014)
	ptr.idx++
	if !ptr.valid() {
		__antithesis_instrumentation__.Notify(115016)
		return
	} else {
		__antithesis_instrumentation__.Notify(115017)
	}
	__antithesis_instrumentation__.Notify(115015)
	if ptr.idx%replicatedCmdBufNodeSize == 0 {
		__antithesis_instrumentation__.Notify(115018)
		ptr.node = ptr.node.next
	} else {
		__antithesis_instrumentation__.Notify(115019)
	}
}

type replicatedCmdBufSlice struct {
	head, tail replicatedCmdBufPtr
}

func (it *replicatedCmdBufSlice) init(buf *replicatedCmdBuf) {
	*it = replicatedCmdBufSlice{
		head: replicatedCmdBufPtr{idx: 0, buf: buf, node: buf.head},
		tail: replicatedCmdBufPtr{idx: buf.len, buf: buf, node: buf.tail},
	}
}

func (it *replicatedCmdBufSlice) initEmpty(buf *replicatedCmdBuf) {
	__antithesis_instrumentation__.Notify(115020)
	*it = replicatedCmdBufSlice{
		head: replicatedCmdBufPtr{idx: 0, buf: buf, node: buf.head},
		tail: replicatedCmdBufPtr{idx: 0, buf: buf, node: buf.head},
	}
}

func (it *replicatedCmdBufSlice) len() int {
	__antithesis_instrumentation__.Notify(115021)
	return int(it.tail.idx - it.head.idx)
}

func (it *replicatedCmdBufSlice) Valid() bool {
	__antithesis_instrumentation__.Notify(115022)
	return it.len() > 0
}

func (it *replicatedCmdBufSlice) Next() {
	__antithesis_instrumentation__.Notify(115023)
	it.head.next()
}

func (it *replicatedCmdBufSlice) cur() *replicatedCmd {
	__antithesis_instrumentation__.Notify(115024)
	return it.head.cur()
}
func (it *replicatedCmdBufSlice) Cur() apply.Command {
	__antithesis_instrumentation__.Notify(115025)
	return it.head.cur()
}
func (it *replicatedCmdBufSlice) CurChecked() apply.CheckedCommand {
	__antithesis_instrumentation__.Notify(115026)
	return it.head.cur()
}
func (it *replicatedCmdBufSlice) CurApplied() apply.AppliedCommand {
	__antithesis_instrumentation__.Notify(115027)
	return it.head.cur()
}

func (it *replicatedCmdBufSlice) append(cmd *replicatedCmd) {
	__antithesis_instrumentation__.Notify(115028)
	cur := it.tail.cur()
	if cur == cmd {
		__antithesis_instrumentation__.Notify(115030)

	} else {
		__antithesis_instrumentation__.Notify(115031)
		*cur = *cmd
	}
	__antithesis_instrumentation__.Notify(115029)
	it.tail.next()
}
func (it *replicatedCmdBufSlice) Append(cmd apply.Command) {
	__antithesis_instrumentation__.Notify(115032)
	it.append(cmd.(*replicatedCmd))
}
func (it *replicatedCmdBufSlice) AppendChecked(cmd apply.CheckedCommand) {
	__antithesis_instrumentation__.Notify(115033)
	it.append(cmd.(*replicatedCmd))
}
func (it *replicatedCmdBufSlice) AppendApplied(cmd apply.AppliedCommand) {
	__antithesis_instrumentation__.Notify(115034)
	it.append(cmd.(*replicatedCmd))
}

func (it *replicatedCmdBufSlice) newList() *replicatedCmdBufSlice {
	__antithesis_instrumentation__.Notify(115035)
	it2 := it.head.buf.newIter()
	it2.initEmpty(it.head.buf)
	return it2
}
func (it *replicatedCmdBufSlice) NewList() apply.CommandList {
	__antithesis_instrumentation__.Notify(115036)
	return it.newList()
}
func (it *replicatedCmdBufSlice) NewCheckedList() apply.CheckedCommandList {
	__antithesis_instrumentation__.Notify(115037)
	return it.newList()
}
func (it *replicatedCmdBufSlice) NewAppliedList() apply.AppliedCommandList {
	__antithesis_instrumentation__.Notify(115038)
	return it.newList()
}

func (it *replicatedCmdBufSlice) Close() {
	__antithesis_instrumentation__.Notify(115039)
	it.head.buf.free.put(it)
}

type replicatedCmdBufSliceFreeList struct {
	iters [replicatedCmdBufSliceFreeListSize]replicatedCmdBufSlice
	inUse [replicatedCmdBufSliceFreeListSize]bool
}

func (f *replicatedCmdBufSliceFreeList) put(it *replicatedCmdBufSlice) {
	__antithesis_instrumentation__.Notify(115040)
	*it = replicatedCmdBufSlice{}
	for i := range f.iters {
		__antithesis_instrumentation__.Notify(115041)
		if &f.iters[i] == it {
			__antithesis_instrumentation__.Notify(115042)
			f.inUse[i] = false
			return
		} else {
			__antithesis_instrumentation__.Notify(115043)
		}
	}
}

func (f *replicatedCmdBufSliceFreeList) get() *replicatedCmdBufSlice {
	__antithesis_instrumentation__.Notify(115044)
	for i, inUse := range f.inUse {
		__antithesis_instrumentation__.Notify(115046)
		if !inUse {
			__antithesis_instrumentation__.Notify(115047)
			f.inUse[i] = true
			return &f.iters[i]
		} else {
			__antithesis_instrumentation__.Notify(115048)
		}
	}
	__antithesis_instrumentation__.Notify(115045)
	panic("replicatedCmdBufSliceFreeList has no free elements. Is " +
		"replicatedCmdBufSliceFreeListSize tuned properly? Are we " +
		"leaking iterators by not calling Close?")
}
