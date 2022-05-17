package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"container/heap"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type resolvedTimestamp struct {
	init       bool
	closedTS   hlc.Timestamp
	resolvedTS hlc.Timestamp
	intentQ    unresolvedIntentQueue
}

func makeResolvedTimestamp() resolvedTimestamp {
	__antithesis_instrumentation__.Notify(113981)
	return resolvedTimestamp{
		intentQ: makeUnresolvedIntentQueue(),
	}
}

func (rts *resolvedTimestamp) Get() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(113982)
	return rts.resolvedTS
}

func (rts *resolvedTimestamp) Init() bool {
	__antithesis_instrumentation__.Notify(113983)
	rts.init = true

	rts.intentQ.AllowNegRefCount(false)
	return rts.recompute()
}

func (rts *resolvedTimestamp) IsInit() bool {
	__antithesis_instrumentation__.Notify(113984)
	return rts.init
}

func (rts *resolvedTimestamp) ForwardClosedTS(newClosedTS hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(113985)
	if rts.closedTS.Forward(newClosedTS) {
		__antithesis_instrumentation__.Notify(113987)
		return rts.recompute()
	} else {
		__antithesis_instrumentation__.Notify(113988)
	}
	__antithesis_instrumentation__.Notify(113986)
	rts.assertNoChange()
	return false
}

func (rts *resolvedTimestamp) ConsumeLogicalOp(op enginepb.MVCCLogicalOp) bool {
	__antithesis_instrumentation__.Notify(113989)
	if rts.consumeLogicalOp(op) {
		__antithesis_instrumentation__.Notify(113991)
		return rts.recompute()
	} else {
		__antithesis_instrumentation__.Notify(113992)
	}
	__antithesis_instrumentation__.Notify(113990)
	rts.assertNoChange()
	return false
}

func (rts *resolvedTimestamp) consumeLogicalOp(op enginepb.MVCCLogicalOp) bool {
	__antithesis_instrumentation__.Notify(113993)
	switch t := op.GetValue().(type) {
	case *enginepb.MVCCWriteValueOp:
		__antithesis_instrumentation__.Notify(113994)
		rts.assertOpAboveRTS(op, t.Timestamp)
		return false

	case *enginepb.MVCCWriteIntentOp:
		__antithesis_instrumentation__.Notify(113995)
		rts.assertOpAboveRTS(op, t.Timestamp)
		return rts.intentQ.IncRef(t.TxnID, t.TxnKey, t.TxnMinTimestamp, t.Timestamp)

	case *enginepb.MVCCUpdateIntentOp:
		__antithesis_instrumentation__.Notify(113996)
		return rts.intentQ.UpdateTS(t.TxnID, t.Timestamp)

	case *enginepb.MVCCCommitIntentOp:
		__antithesis_instrumentation__.Notify(113997)
		return rts.intentQ.DecrRef(t.TxnID, t.Timestamp)

	case *enginepb.MVCCAbortIntentOp:
		__antithesis_instrumentation__.Notify(113998)

		return rts.intentQ.DecrRef(t.TxnID, hlc.Timestamp{})

	case *enginepb.MVCCAbortTxnOp:
		__antithesis_instrumentation__.Notify(113999)

		if !rts.IsInit() {
			__antithesis_instrumentation__.Notify(114002)

			return false
		} else {
			__antithesis_instrumentation__.Notify(114003)
		}
		__antithesis_instrumentation__.Notify(114000)
		return rts.intentQ.Del(t.TxnID)

	default:
		__antithesis_instrumentation__.Notify(114001)
		panic(errors.AssertionFailedf("unknown logical op %T", t))
	}
}

func (rts *resolvedTimestamp) recompute() bool {
	__antithesis_instrumentation__.Notify(114004)
	if !rts.IsInit() {
		__antithesis_instrumentation__.Notify(114009)
		return false
	} else {
		__antithesis_instrumentation__.Notify(114010)
	}
	__antithesis_instrumentation__.Notify(114005)
	if rts.closedTS.Less(rts.resolvedTS) {
		__antithesis_instrumentation__.Notify(114011)
		panic(fmt.Sprintf("closed timestamp below resolved timestamp: %s < %s",
			rts.closedTS, rts.resolvedTS))
	} else {
		__antithesis_instrumentation__.Notify(114012)
	}
	__antithesis_instrumentation__.Notify(114006)
	newTS := rts.closedTS

	if txn := rts.intentQ.Oldest(); txn != nil {
		__antithesis_instrumentation__.Notify(114013)
		if txn.timestamp.LessEq(rts.resolvedTS) {
			__antithesis_instrumentation__.Notify(114015)
			panic(fmt.Sprintf("unresolved txn equal to or below resolved timestamp: %s <= %s",
				txn.timestamp, rts.resolvedTS))
		} else {
			__antithesis_instrumentation__.Notify(114016)
		}
		__antithesis_instrumentation__.Notify(114014)

		txnTS := txn.timestamp.Prev()
		newTS.Backward(txnTS)
	} else {
		__antithesis_instrumentation__.Notify(114017)
	}
	__antithesis_instrumentation__.Notify(114007)

	newTS.Logical = 0

	if newTS.Less(rts.resolvedTS) {
		__antithesis_instrumentation__.Notify(114018)
		panic(fmt.Sprintf("resolved timestamp regression, was %s, recomputed as %s",
			rts.resolvedTS, newTS))
	} else {
		__antithesis_instrumentation__.Notify(114019)
	}
	__antithesis_instrumentation__.Notify(114008)
	return rts.resolvedTS.Forward(newTS)
}

func (rts *resolvedTimestamp) assertNoChange() {
	__antithesis_instrumentation__.Notify(114020)
	before := rts.resolvedTS
	changed := rts.recompute()
	if changed || func() bool {
		__antithesis_instrumentation__.Notify(114021)
		return !before.EqOrdering(rts.resolvedTS) == true
	}() == true {
		__antithesis_instrumentation__.Notify(114022)
		panic(fmt.Sprintf("unexpected resolved timestamp change on recomputation, "+
			"was %s, recomputed as %s", before, rts.resolvedTS))
	} else {
		__antithesis_instrumentation__.Notify(114023)
	}
}

func (rts *resolvedTimestamp) assertOpAboveRTS(op enginepb.MVCCLogicalOp, opTS hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(114024)
	if opTS.LessEq(rts.resolvedTS) {
		__antithesis_instrumentation__.Notify(114025)
		panic(fmt.Sprintf("resolved timestamp %s equal to or above timestamp of operation %v",
			rts.resolvedTS, op))
	} else {
		__antithesis_instrumentation__.Notify(114026)
	}
}

type unresolvedTxn struct {
	txnID           uuid.UUID
	txnKey          roachpb.Key
	txnMinTimestamp hlc.Timestamp
	timestamp       hlc.Timestamp
	refCount        int

	index int
}

func (t *unresolvedTxn) asTxnMeta() enginepb.TxnMeta {
	__antithesis_instrumentation__.Notify(114027)
	return enginepb.TxnMeta{
		ID:             t.txnID,
		Key:            t.txnKey,
		MinTimestamp:   t.txnMinTimestamp,
		WriteTimestamp: t.timestamp,
	}
}

type unresolvedTxnHeap []*unresolvedTxn

func (h unresolvedTxnHeap) Len() int { __antithesis_instrumentation__.Notify(114028); return len(h) }

func (h unresolvedTxnHeap) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(114029)

	if h[i].timestamp.EqOrdering(h[j].timestamp) {
		__antithesis_instrumentation__.Notify(114031)
		return bytes.Compare(h[i].txnID.GetBytes(), h[j].txnID.GetBytes()) < 0
	} else {
		__antithesis_instrumentation__.Notify(114032)
	}
	__antithesis_instrumentation__.Notify(114030)
	return h[i].timestamp.Less(h[j].timestamp)
}

func (h unresolvedTxnHeap) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(114033)
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}

func (h *unresolvedTxnHeap) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(114034)
	n := len(*h)
	txn := x.(*unresolvedTxn)
	txn.index = n
	*h = append(*h, txn)
}

func (h *unresolvedTxnHeap) Pop() interface{} {
	__antithesis_instrumentation__.Notify(114035)
	old := *h
	n := len(old)
	txn := old[n-1]
	txn.index = -1
	old[n-1] = nil
	*h = old[0 : n-1]
	return txn
}

type unresolvedIntentQueue struct {
	txns             map[uuid.UUID]*unresolvedTxn
	minHeap          unresolvedTxnHeap
	allowNegRefCount bool
}

func makeUnresolvedIntentQueue() unresolvedIntentQueue {
	__antithesis_instrumentation__.Notify(114036)
	return unresolvedIntentQueue{
		txns:             make(map[uuid.UUID]*unresolvedTxn),
		allowNegRefCount: true,
	}
}

func (uiq *unresolvedIntentQueue) Len() int {
	__antithesis_instrumentation__.Notify(114037)
	return uiq.minHeap.Len()
}

func (uiq *unresolvedIntentQueue) Oldest() *unresolvedTxn {
	__antithesis_instrumentation__.Notify(114038)
	if uiq.Len() == 0 {
		__antithesis_instrumentation__.Notify(114040)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(114041)
	}
	__antithesis_instrumentation__.Notify(114039)
	return uiq.minHeap[0]
}

func (uiq *unresolvedIntentQueue) Before(ts hlc.Timestamp) []*unresolvedTxn {
	__antithesis_instrumentation__.Notify(114042)
	var txns []*unresolvedTxn
	var collect func(int)
	collect = func(i int) {
		__antithesis_instrumentation__.Notify(114044)
		if len(uiq.minHeap) > i && func() bool {
			__antithesis_instrumentation__.Notify(114045)
			return uiq.minHeap[i].timestamp.Less(ts) == true
		}() == true {
			__antithesis_instrumentation__.Notify(114046)
			txns = append(txns, uiq.minHeap[i])
			collect((2 * i) + 1)
			collect((2 * i) + 2)
		} else {
			__antithesis_instrumentation__.Notify(114047)
		}
	}
	__antithesis_instrumentation__.Notify(114043)
	collect(0)
	return txns
}

func (uiq *unresolvedIntentQueue) IncRef(
	txnID uuid.UUID, txnKey roachpb.Key, txnMinTS, ts hlc.Timestamp,
) bool {
	__antithesis_instrumentation__.Notify(114048)
	return uiq.updateTxn(txnID, txnKey, txnMinTS, ts, +1)
}

func (uiq *unresolvedIntentQueue) DecrRef(txnID uuid.UUID, ts hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(114049)
	return uiq.updateTxn(txnID, nil, hlc.Timestamp{}, ts, -1)
}

func (uiq *unresolvedIntentQueue) UpdateTS(txnID uuid.UUID, ts hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(114050)
	return uiq.updateTxn(txnID, nil, hlc.Timestamp{}, ts, 0)
}

func (uiq *unresolvedIntentQueue) updateTxn(
	txnID uuid.UUID, txnKey roachpb.Key, txnMinTS, ts hlc.Timestamp, delta int,
) bool {
	__antithesis_instrumentation__.Notify(114051)
	txn, ok := uiq.txns[txnID]
	if !ok {
		__antithesis_instrumentation__.Notify(114055)
		if delta == 0 || func() bool {
			__antithesis_instrumentation__.Notify(114057)
			return (delta < 0 && func() bool {
				__antithesis_instrumentation__.Notify(114058)
				return !uiq.allowNegRefCount == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(114059)

			return false
		} else {
			__antithesis_instrumentation__.Notify(114060)
		}
		__antithesis_instrumentation__.Notify(114056)

		txn = &unresolvedTxn{
			txnID:           txnID,
			txnKey:          txnKey,
			txnMinTimestamp: txnMinTS,
			timestamp:       ts,
			refCount:        delta,
		}
		uiq.txns[txn.txnID] = txn
		heap.Push(&uiq.minHeap, txn)

		return false
	} else {
		__antithesis_instrumentation__.Notify(114061)
	}
	__antithesis_instrumentation__.Notify(114052)

	wasMin := txn.index == 0

	txn.refCount += delta
	if txn.refCount == 0 || func() bool {
		__antithesis_instrumentation__.Notify(114062)
		return (txn.refCount < 0 && func() bool {
			__antithesis_instrumentation__.Notify(114063)
			return !uiq.allowNegRefCount == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(114064)

		delete(uiq.txns, txn.txnID)
		heap.Remove(&uiq.minHeap, txn.index)
		return wasMin
	} else {
		__antithesis_instrumentation__.Notify(114065)
	}
	__antithesis_instrumentation__.Notify(114053)

	if txn.timestamp.Forward(ts) {
		__antithesis_instrumentation__.Notify(114066)
		heap.Fix(&uiq.minHeap, txn.index)
		return wasMin
	} else {
		__antithesis_instrumentation__.Notify(114067)
	}
	__antithesis_instrumentation__.Notify(114054)
	return false
}

func (uiq *unresolvedIntentQueue) Del(txnID uuid.UUID) bool {
	__antithesis_instrumentation__.Notify(114068)

	txn, ok := uiq.txns[txnID]
	if !ok {
		__antithesis_instrumentation__.Notify(114070)

		return false
	} else {
		__antithesis_instrumentation__.Notify(114071)
	}
	__antithesis_instrumentation__.Notify(114069)

	wasMin := txn.index == 0

	delete(uiq.txns, txn.txnID)
	heap.Remove(&uiq.minHeap, txn.index)
	return wasMin
}

func (uiq *unresolvedIntentQueue) AllowNegRefCount(b bool) {
	__antithesis_instrumentation__.Notify(114072)
	if !b {
		__antithesis_instrumentation__.Notify(114074)

		uiq.assertOnlyPositiveRefCounts()
	} else {
		__antithesis_instrumentation__.Notify(114075)
	}
	__antithesis_instrumentation__.Notify(114073)
	uiq.allowNegRefCount = b
}

func (uiq *unresolvedIntentQueue) assertOnlyPositiveRefCounts() {
	__antithesis_instrumentation__.Notify(114076)
	for _, txn := range uiq.txns {
		__antithesis_instrumentation__.Notify(114077)
		if txn.refCount <= 0 {
			__antithesis_instrumentation__.Notify(114078)
			panic(fmt.Sprintf("negative refcount %d for txn %+v", txn.refCount, txn))
		} else {
			__antithesis_instrumentation__.Notify(114079)
		}
	}
}
