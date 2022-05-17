package concurrency

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/list"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

const defaultLockTableSize = 10000

type waitKind int

const (
	_ waitKind = iota

	waitFor

	waitForDistinguished

	waitElsewhere

	waitSelf

	waitQueueMaxLengthExceeded

	doneWaiting
)

type waitingState struct {
	kind waitKind

	txn           *enginepb.TxnMeta
	key           roachpb.Key
	held          bool
	queuedWriters int
	queuedReaders int

	guardAccess spanset.SpanAccess

	lockWaitStart time.Time
}

func (s waitingState) String() string {
	__antithesis_instrumentation__.Notify(99402)
	switch s.kind {
	case waitFor, waitForDistinguished:
		__antithesis_instrumentation__.Notify(99403)
		distinguished := ""
		if s.kind == waitForDistinguished {
			__antithesis_instrumentation__.Notify(99412)
			distinguished = " (distinguished)"
		} else {
			__antithesis_instrumentation__.Notify(99413)
		}
		__antithesis_instrumentation__.Notify(99404)
		target := "holding lock"
		if !s.held {
			__antithesis_instrumentation__.Notify(99414)
			target = "running request"
		} else {
			__antithesis_instrumentation__.Notify(99415)
		}
		__antithesis_instrumentation__.Notify(99405)
		return fmt.Sprintf("wait for%s txn %s %s @ key %s (queuedWriters: %d, queuedReaders: %d)",
			distinguished, s.txn.ID.Short(), target, s.key, s.queuedWriters, s.queuedReaders)
	case waitSelf:
		__antithesis_instrumentation__.Notify(99406)
		return fmt.Sprintf("wait self @ key %s", s.key)
	case waitElsewhere:
		__antithesis_instrumentation__.Notify(99407)
		if !s.held {
			__antithesis_instrumentation__.Notify(99416)
			return "wait elsewhere by proceeding to evaluation"
		} else {
			__antithesis_instrumentation__.Notify(99417)
		}
		__antithesis_instrumentation__.Notify(99408)
		return fmt.Sprintf("wait elsewhere for txn %s @ key %s", s.txn.ID.Short(), s.key)
	case waitQueueMaxLengthExceeded:
		__antithesis_instrumentation__.Notify(99409)
		return fmt.Sprintf("wait-queue maximum length exceeded @ key %s with length %d",
			s.key, s.queuedWriters)
	case doneWaiting:
		__antithesis_instrumentation__.Notify(99410)
		return "done waiting"
	default:
		__antithesis_instrumentation__.Notify(99411)
		panic("unhandled waitingState.kind")
	}
}

type treeMu struct {
	mu syncutil.RWMutex

	lockIDSeqNum uint64

	btree

	numLocks int64

	lockAddMaxLocksCheckInterval uint64
}

type lockTableImpl struct {
	rID roachpb.RangeID

	enabled    bool
	enabledMu  syncutil.RWMutex
	enabledSeq roachpb.LeaseSequence

	seqNum uint64

	locks [spanset.NumSpanScope]treeMu

	maxLocks int64

	minLocks int64

	finalizedTxnCache txnCache

	clock *hlc.Clock
}

var _ lockTable = &lockTableImpl{}

func newLockTable(maxLocks int64, rangeID roachpb.RangeID, clock *hlc.Clock) *lockTableImpl {
	__antithesis_instrumentation__.Notify(99418)
	lt := &lockTableImpl{
		rID:   rangeID,
		clock: clock,
	}
	lt.setMaxLocks(maxLocks)
	return lt
}

func (t *lockTableImpl) setMaxLocks(maxLocks int64) {
	__antithesis_instrumentation__.Notify(99419)

	lockAddMaxLocksCheckInterval := maxLocks / (int64(spanset.NumSpanScope) * 20)
	if lockAddMaxLocksCheckInterval == 0 {
		__antithesis_instrumentation__.Notify(99421)
		lockAddMaxLocksCheckInterval = 1
	} else {
		__antithesis_instrumentation__.Notify(99422)
	}
	__antithesis_instrumentation__.Notify(99420)
	t.maxLocks = maxLocks
	t.minLocks = maxLocks / 2
	for i := 0; i < int(spanset.NumSpanScope); i++ {
		__antithesis_instrumentation__.Notify(99423)
		t.locks[i].lockAddMaxLocksCheckInterval = uint64(lockAddMaxLocksCheckInterval)
	}
}

type lockTableGuardImpl struct {
	seqNum uint64
	lt     *lockTableImpl

	txn                *enginepb.TxnMeta
	ts                 hlc.Timestamp
	spans              *spanset.SpanSet
	maxWaitQueueLength int

	tableSnapshot [spanset.NumSpanScope]btree

	notRemovableLock *lockState

	key roachpb.Key

	ss    spanset.SpanScope
	sa    spanset.SpanAccess
	index int

	mu struct {
		syncutil.Mutex
		startWait bool

		curLockWaitStart time.Time

		state  waitingState
		signal chan struct{}

		locks map[*lockState]struct{}

		mustFindNextLockAfter bool
	}

	toResolve []roachpb.LockUpdate
}

var _ lockTableGuard = &lockTableGuardImpl{}

var lockTableGuardImplPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(99424)
		g := new(lockTableGuardImpl)
		g.mu.signal = make(chan struct{}, 1)
		g.mu.locks = make(map[*lockState]struct{})
		return g
	},
}

func newLockTableGuardImpl() *lockTableGuardImpl {
	__antithesis_instrumentation__.Notify(99425)
	return lockTableGuardImplPool.Get().(*lockTableGuardImpl)
}

func releaseLockTableGuardImpl(g *lockTableGuardImpl) {
	__antithesis_instrumentation__.Notify(99426)

	signal, locks := g.mu.signal, g.mu.locks
	select {
	case <-signal:
		__antithesis_instrumentation__.Notify(99429)
	default:
		__antithesis_instrumentation__.Notify(99430)
	}
	__antithesis_instrumentation__.Notify(99427)
	if len(locks) != 0 {
		__antithesis_instrumentation__.Notify(99431)
		panic("lockTableGuardImpl.mu.locks not empty after Dequeue")
	} else {
		__antithesis_instrumentation__.Notify(99432)
	}
	__antithesis_instrumentation__.Notify(99428)

	*g = lockTableGuardImpl{}
	g.mu.signal = signal
	g.mu.locks = locks
	lockTableGuardImplPool.Put(g)
}

func (g *lockTableGuardImpl) ShouldWait() bool {
	__antithesis_instrumentation__.Notify(99433)
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.mu.startWait
}

func (g *lockTableGuardImpl) ResolveBeforeScanning() []roachpb.LockUpdate {
	__antithesis_instrumentation__.Notify(99434)
	return g.toResolve
}

func (g *lockTableGuardImpl) NewStateChan() chan struct{} {
	__antithesis_instrumentation__.Notify(99435)
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.mu.signal
}

func (g *lockTableGuardImpl) CurState() waitingState {
	__antithesis_instrumentation__.Notify(99436)
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.mu.mustFindNextLockAfter {
		__antithesis_instrumentation__.Notify(99438)
		return g.mu.state
	} else {
		__antithesis_instrumentation__.Notify(99439)
	}
	__antithesis_instrumentation__.Notify(99437)

	g.mu.mustFindNextLockAfter = false
	g.mu.Unlock()
	g.findNextLockAfter(false)
	g.mu.Lock()
	return g.mu.state
}

func (g *lockTableGuardImpl) updateStateLocked(newState waitingState) {
	__antithesis_instrumentation__.Notify(99440)
	g.mu.state = newState
	switch newState.kind {
	case waitFor, waitForDistinguished, waitSelf, waitElsewhere:
		__antithesis_instrumentation__.Notify(99441)
		g.mu.state.lockWaitStart = g.mu.curLockWaitStart
	default:
		__antithesis_instrumentation__.Notify(99442)
		g.mu.state.lockWaitStart = time.Time{}
	}
}

func (g *lockTableGuardImpl) CheckOptimisticNoConflicts(spanSet *spanset.SpanSet) (ok bool) {
	__antithesis_instrumentation__.Notify(99443)

	originalSpanSet := g.spans
	g.spans = spanSet
	g.sa = spanset.NumSpanAccess - 1
	g.ss = spanset.SpanScope(0)
	g.index = -1
	defer func() {
		__antithesis_instrumentation__.Notify(99446)
		g.spans = originalSpanSet
	}()
	__antithesis_instrumentation__.Notify(99444)
	span := stepToNextSpan(g)
	for span != nil {
		__antithesis_instrumentation__.Notify(99447)
		startKey := span.Key
		tree := g.tableSnapshot[g.ss]
		iter := tree.MakeIter()
		ltRange := &lockState{key: startKey, endKey: span.EndKey}
		for iter.FirstOverlap(ltRange); iter.Valid(); iter.NextOverlap(ltRange) {
			__antithesis_instrumentation__.Notify(99449)
			l := iter.Cur()
			if !l.isNonConflictingLock(g, g.sa) {
				__antithesis_instrumentation__.Notify(99450)
				return false
			} else {
				__antithesis_instrumentation__.Notify(99451)
			}
		}
		__antithesis_instrumentation__.Notify(99448)
		span = stepToNextSpan(g)
	}
	__antithesis_instrumentation__.Notify(99445)
	return true
}

func (g *lockTableGuardImpl) notify() {
	__antithesis_instrumentation__.Notify(99452)
	select {
	case g.mu.signal <- struct{}{}:
		__antithesis_instrumentation__.Notify(99453)
	default:
		__antithesis_instrumentation__.Notify(99454)
	}
}

func (g *lockTableGuardImpl) doneWaitingAtLock(hasReservation bool, l *lockState) {
	__antithesis_instrumentation__.Notify(99455)
	g.mu.Lock()
	if !hasReservation {
		__antithesis_instrumentation__.Notify(99457)
		delete(g.mu.locks, l)
	} else {
		__antithesis_instrumentation__.Notify(99458)
	}
	__antithesis_instrumentation__.Notify(99456)
	g.mu.mustFindNextLockAfter = true
	g.notify()
	g.mu.Unlock()
}

func (g *lockTableGuardImpl) isSameTxn(txn *enginepb.TxnMeta) bool {
	__antithesis_instrumentation__.Notify(99459)
	return g.txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(99460)
		return g.txn.ID == txn.ID == true
	}() == true
}

func (g *lockTableGuardImpl) isSameTxnAsReservation(ws waitingState) bool {
	__antithesis_instrumentation__.Notify(99461)
	return !ws.held && func() bool {
		__antithesis_instrumentation__.Notify(99462)
		return g.isSameTxn(ws.txn) == true
	}() == true
}

func (g *lockTableGuardImpl) findNextLockAfter(notify bool) {
	__antithesis_instrumentation__.Notify(99463)
	spans := g.spans.GetSpans(g.sa, g.ss)
	var span *spanset.Span
	resumingInSameSpan := false
	if g.index == -1 || func() bool {
		__antithesis_instrumentation__.Notify(99468)
		return len(spans[g.index].EndKey) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(99469)
		span = stepToNextSpan(g)
	} else {
		__antithesis_instrumentation__.Notify(99470)
		span = &spans[g.index]
		resumingInSameSpan = true
	}
	__antithesis_instrumentation__.Notify(99464)

	var locksToGC [spanset.NumSpanScope][]*lockState
	defer func() {
		__antithesis_instrumentation__.Notify(99471)
		for i := 0; i < len(locksToGC); i++ {
			__antithesis_instrumentation__.Notify(99472)
			if len(locksToGC[i]) > 0 {
				__antithesis_instrumentation__.Notify(99473)
				g.lt.tryGCLocks(&g.lt.locks[i], locksToGC[i])
			} else {
				__antithesis_instrumentation__.Notify(99474)
			}
		}
	}()
	__antithesis_instrumentation__.Notify(99465)

	for span != nil {
		__antithesis_instrumentation__.Notify(99475)
		startKey := span.Key
		if resumingInSameSpan {
			__antithesis_instrumentation__.Notify(99478)
			startKey = g.key
		} else {
			__antithesis_instrumentation__.Notify(99479)
		}
		__antithesis_instrumentation__.Notify(99476)
		tree := g.tableSnapshot[g.ss]
		iter := tree.MakeIter()

		ltRange := &lockState{key: startKey, endKey: span.EndKey}
		for iter.FirstOverlap(ltRange); iter.Valid(); iter.NextOverlap(ltRange) {
			__antithesis_instrumentation__.Notify(99480)
			l := iter.Cur()
			if resumingInSameSpan {
				__antithesis_instrumentation__.Notify(99483)
				resumingInSameSpan = false
				if l.key.Equal(startKey) {
					__antithesis_instrumentation__.Notify(99484)

					continue
				} else {
					__antithesis_instrumentation__.Notify(99485)
				}

			} else {
				__antithesis_instrumentation__.Notify(99486)
			}
			__antithesis_instrumentation__.Notify(99481)
			wait, transitionedToFree := l.tryActiveWait(g, g.sa, notify, g.lt.clock)
			if transitionedToFree {
				__antithesis_instrumentation__.Notify(99487)
				locksToGC[g.ss] = append(locksToGC[g.ss], l)
			} else {
				__antithesis_instrumentation__.Notify(99488)
			}
			__antithesis_instrumentation__.Notify(99482)
			if wait {
				__antithesis_instrumentation__.Notify(99489)
				return
			} else {
				__antithesis_instrumentation__.Notify(99490)
			}
		}
		__antithesis_instrumentation__.Notify(99477)
		resumingInSameSpan = false
		span = stepToNextSpan(g)
	}
	__antithesis_instrumentation__.Notify(99466)
	if len(g.toResolve) > 0 {
		__antithesis_instrumentation__.Notify(99491)
		j := 0

		for i := range g.toResolve {
			__antithesis_instrumentation__.Notify(99493)
			if heldByTxn := g.lt.updateLockInternal(&g.toResolve[i]); heldByTxn {
				__antithesis_instrumentation__.Notify(99494)
				g.toResolve[j] = g.toResolve[i]
				j++
			} else {
				__antithesis_instrumentation__.Notify(99495)
			}
		}
		__antithesis_instrumentation__.Notify(99492)
		g.toResolve = g.toResolve[:j]
	} else {
		__antithesis_instrumentation__.Notify(99496)
	}
	__antithesis_instrumentation__.Notify(99467)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.updateStateLocked(waitingState{kind: doneWaiting})

	if notify {
		__antithesis_instrumentation__.Notify(99497)
		if len(g.toResolve) > 0 {
			__antithesis_instrumentation__.Notify(99499)

			g.mu.startWait = true
		} else {
			__antithesis_instrumentation__.Notify(99500)
		}
		__antithesis_instrumentation__.Notify(99498)
		g.notify()
	} else {
		__antithesis_instrumentation__.Notify(99501)
	}
}

type queuedGuard struct {
	guard  *lockTableGuardImpl
	active bool
}

type lockHolderInfo struct {
	txn *enginepb.TxnMeta

	seqs []enginepb.TxnSeq

	ts hlc.Timestamp
}

func (lh *lockHolderInfo) isEmpty() bool {
	__antithesis_instrumentation__.Notify(99502)
	return lh.txn == nil && func() bool {
		__antithesis_instrumentation__.Notify(99503)
		return lh.seqs == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(99504)
		return lh.ts.IsEmpty() == true
	}() == true
}

type lockState struct {
	id     uint64
	endKey []byte

	key roachpb.Key
	ss  spanset.SpanScope

	mu syncutil.Mutex

	holder struct {
		locked bool

		holder [lock.MaxDurability + 1]lockHolderInfo

		startTime time.Time
	}

	lockWaitQueue

	notRemovable int
}

type lockWaitQueue struct {
	reservation *lockTableGuardImpl

	queuedWriters list.List

	waitingReaders list.List

	distinguishedWaiter *lockTableGuardImpl
}

func (l *lockState) ID() uint64     { __antithesis_instrumentation__.Notify(99505); return l.id }
func (l *lockState) Key() []byte    { __antithesis_instrumentation__.Notify(99506); return l.key }
func (l *lockState) EndKey() []byte { __antithesis_instrumentation__.Notify(99507); return l.endKey }
func (l *lockState) New() *lockState {
	__antithesis_instrumentation__.Notify(99508)
	return new(lockState)
}
func (l *lockState) SetID(v uint64)     { __antithesis_instrumentation__.Notify(99509); l.id = v }
func (l *lockState) SetKey(v []byte)    { __antithesis_instrumentation__.Notify(99510); l.key = v }
func (l *lockState) SetEndKey(v []byte) { __antithesis_instrumentation__.Notify(99511); l.endKey = v }

func (l *lockState) String() string {
	__antithesis_instrumentation__.Notify(99512)
	var sb redact.StringBuilder
	l.safeFormat(&sb, nil)
	return sb.String()
}

func (l *lockState) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(99513)
	var sb redact.StringBuilder
	l.safeFormat(&sb, nil)
	w.Print(sb)
}

func (l *lockState) safeFormat(sb *redact.StringBuilder, finalizedTxnCache *txnCache) {
	__antithesis_instrumentation__.Notify(99514)
	sb.Printf(" lock: %s\n", l.key)
	if l.isEmptyLock() {
		__antithesis_instrumentation__.Notify(99521)
		sb.SafeString("  empty\n")
		return
	} else {
		__antithesis_instrumentation__.Notify(99522)
	}
	__antithesis_instrumentation__.Notify(99515)
	writeResInfo := func(sb *redact.StringBuilder, txn *enginepb.TxnMeta, ts hlc.Timestamp) {
		__antithesis_instrumentation__.Notify(99523)

		sb.Printf("txn: %v, ts: %v, seq: %v\n",
			redact.Safe(txn.ID), redact.Safe(ts), redact.Safe(txn.Sequence))
	}
	__antithesis_instrumentation__.Notify(99516)
	writeHolderInfo := func(sb *redact.StringBuilder, txn *enginepb.TxnMeta, ts hlc.Timestamp) {
		__antithesis_instrumentation__.Notify(99524)
		sb.Printf("  holder: txn: %v, ts: %v, info: ", redact.Safe(txn.ID), redact.Safe(ts))
		first := true
		for i := range l.holder.holder {
			__antithesis_instrumentation__.Notify(99526)
			h := &l.holder.holder[i]
			if h.txn == nil {
				__antithesis_instrumentation__.Notify(99532)
				continue
			} else {
				__antithesis_instrumentation__.Notify(99533)
			}
			__antithesis_instrumentation__.Notify(99527)
			if !first {
				__antithesis_instrumentation__.Notify(99534)
				sb.SafeString(", ")
			} else {
				__antithesis_instrumentation__.Notify(99535)
			}
			__antithesis_instrumentation__.Notify(99528)
			first = false
			if lock.Durability(i) == lock.Replicated {
				__antithesis_instrumentation__.Notify(99536)
				sb.SafeString("repl ")
			} else {
				__antithesis_instrumentation__.Notify(99537)
				sb.SafeString("unrepl ")
			}
			__antithesis_instrumentation__.Notify(99529)
			if finalizedTxnCache != nil {
				__antithesis_instrumentation__.Notify(99538)
				finalizedTxn, ok := finalizedTxnCache.get(h.txn.ID)
				if ok {
					__antithesis_instrumentation__.Notify(99539)
					var statusStr string
					switch finalizedTxn.Status {
					case roachpb.COMMITTED:
						__antithesis_instrumentation__.Notify(99541)
						statusStr = "committed"
					case roachpb.ABORTED:
						__antithesis_instrumentation__.Notify(99542)
						statusStr = "aborted"
					default:
						__antithesis_instrumentation__.Notify(99543)
					}
					__antithesis_instrumentation__.Notify(99540)
					sb.Printf("[holder finalized: %s] ", redact.Safe(statusStr))
				} else {
					__antithesis_instrumentation__.Notify(99544)
				}
			} else {
				__antithesis_instrumentation__.Notify(99545)
			}
			__antithesis_instrumentation__.Notify(99530)
			sb.Printf("epoch: %d, seqs: [%d", redact.Safe(h.txn.Epoch), redact.Safe(h.seqs[0]))
			for j := 1; j < len(h.seqs); j++ {
				__antithesis_instrumentation__.Notify(99546)
				sb.Printf(", %d", redact.Safe(h.seqs[j]))
			}
			__antithesis_instrumentation__.Notify(99531)
			sb.SafeString("]")
		}
		__antithesis_instrumentation__.Notify(99525)
		sb.SafeString("\n")
	}
	__antithesis_instrumentation__.Notify(99517)
	txn, ts := l.getLockHolder()
	if txn == nil {
		__antithesis_instrumentation__.Notify(99547)
		sb.Printf("  res: req: %d, ", l.reservation.seqNum)
		writeResInfo(sb, l.reservation.txn, l.reservation.ts)
	} else {
		__antithesis_instrumentation__.Notify(99548)
		writeHolderInfo(sb, txn, ts)
	}
	__antithesis_instrumentation__.Notify(99518)

	if l.waitingReaders.Len() > 0 {
		__antithesis_instrumentation__.Notify(99549)
		sb.SafeString("   waiting readers:\n")
		for e := l.waitingReaders.Front(); e != nil; e = e.Next() {
			__antithesis_instrumentation__.Notify(99550)
			g := e.Value.(*lockTableGuardImpl)
			sb.Printf("    req: %d, txn: ", redact.Safe(g.seqNum))
			if g.txn == nil {
				__antithesis_instrumentation__.Notify(99551)
				sb.SafeString("none\n")
			} else {
				__antithesis_instrumentation__.Notify(99552)
				sb.Printf("%v\n", redact.Safe(g.txn.ID))
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(99553)
	}
	__antithesis_instrumentation__.Notify(99519)
	if l.queuedWriters.Len() > 0 {
		__antithesis_instrumentation__.Notify(99554)
		sb.SafeString("   queued writers:\n")
		for e := l.queuedWriters.Front(); e != nil; e = e.Next() {
			__antithesis_instrumentation__.Notify(99555)
			qg := e.Value.(*queuedGuard)
			g := qg.guard
			sb.Printf("    active: %t req: %d, txn: ", redact.Safe(qg.active), redact.Safe(qg.guard.seqNum))
			if g.txn == nil {
				__antithesis_instrumentation__.Notify(99556)
				sb.SafeString("none\n")
			} else {
				__antithesis_instrumentation__.Notify(99557)
				sb.Printf("%v\n", redact.Safe(g.txn.ID))
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(99558)
	}
	__antithesis_instrumentation__.Notify(99520)
	if l.distinguishedWaiter != nil {
		__antithesis_instrumentation__.Notify(99559)
		sb.Printf("   distinguished req: %d\n", redact.Safe(l.distinguishedWaiter.seqNum))
	} else {
		__antithesis_instrumentation__.Notify(99560)
	}
}

func (l *lockState) collectLockStateInfo(
	includeUncontended bool, now time.Time,
) (bool, roachpb.LockStateInfo) {
	__antithesis_instrumentation__.Notify(99561)
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isEmptyLock() {
		__antithesis_instrumentation__.Notify(99564)
		return false, roachpb.LockStateInfo{}
	} else {
		__antithesis_instrumentation__.Notify(99565)
	}
	__antithesis_instrumentation__.Notify(99562)

	if !includeUncontended && func() bool {
		__antithesis_instrumentation__.Notify(99566)
		return l.waitingReaders.Len() == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(99567)
		return l.queuedWriters.Len() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(99568)
		return false, roachpb.LockStateInfo{}
	} else {
		__antithesis_instrumentation__.Notify(99569)
	}
	__antithesis_instrumentation__.Notify(99563)

	return true, l.lockStateInfo(now)
}

func (l *lockState) lockStateInfo(now time.Time) roachpb.LockStateInfo {
	__antithesis_instrumentation__.Notify(99570)
	var txnHolder *enginepb.TxnMeta

	durability := lock.Unreplicated
	if l.holder.locked {
		__antithesis_instrumentation__.Notify(99576)
		if l.holder.holder[lock.Replicated].txn != nil {
			__antithesis_instrumentation__.Notify(99577)
			durability = lock.Replicated
			txnHolder = l.holder.holder[lock.Replicated].txn
		} else {
			__antithesis_instrumentation__.Notify(99578)
			if l.holder.holder[lock.Unreplicated].txn != nil {
				__antithesis_instrumentation__.Notify(99579)
				txnHolder = l.holder.holder[lock.Unreplicated].txn
			} else {
				__antithesis_instrumentation__.Notify(99580)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(99581)
	}
	__antithesis_instrumentation__.Notify(99571)

	waiterCount := l.waitingReaders.Len() + l.queuedWriters.Len()
	hasReservation := l.reservation != nil && func() bool {
		__antithesis_instrumentation__.Notify(99582)
		return l.reservation.txn != nil == true
	}() == true
	if hasReservation {
		__antithesis_instrumentation__.Notify(99583)
		waiterCount++
	} else {
		__antithesis_instrumentation__.Notify(99584)
	}
	__antithesis_instrumentation__.Notify(99572)
	lockWaiters := make([]lock.Waiter, 0, waiterCount)

	if hasReservation {
		__antithesis_instrumentation__.Notify(99585)
		l.reservation.mu.Lock()
		lockWaiters = append(lockWaiters, lock.Waiter{
			WaitingTxn:   l.reservation.txn,
			ActiveWaiter: true,
			Strength:     lock.Exclusive,
			WaitDuration: now.Sub(l.reservation.mu.curLockWaitStart),
		})
		l.reservation.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(99586)
	}
	__antithesis_instrumentation__.Notify(99573)

	for e := l.waitingReaders.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(99587)
		readerGuard := e.Value.(*lockTableGuardImpl)
		readerGuard.mu.Lock()
		lockWaiters = append(lockWaiters, lock.Waiter{
			WaitingTxn:   readerGuard.txn,
			ActiveWaiter: false,
			Strength:     lock.None,
			WaitDuration: now.Sub(readerGuard.mu.curLockWaitStart),
		})
		readerGuard.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(99574)

	for e := l.queuedWriters.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(99588)
		qg := e.Value.(*queuedGuard)
		writerGuard := qg.guard
		writerGuard.mu.Lock()
		lockWaiters = append(lockWaiters, lock.Waiter{
			WaitingTxn:   writerGuard.txn,
			ActiveWaiter: qg.active,
			Strength:     lock.Exclusive,
			WaitDuration: now.Sub(writerGuard.mu.curLockWaitStart),
		})
		writerGuard.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(99575)

	return roachpb.LockStateInfo{
		Key:          l.key,
		LockHolder:   txnHolder,
		Durability:   durability,
		HoldDuration: l.lockHeldDuration(now),
		Waiters:      lockWaiters,
	}
}

func (l *lockState) addToMetrics(m *LockTableMetrics, now time.Time) {
	__antithesis_instrumentation__.Notify(99589)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.isEmptyLock() {
		__antithesis_instrumentation__.Notify(99591)
		return
	} else {
		__antithesis_instrumentation__.Notify(99592)
	}
	__antithesis_instrumentation__.Notify(99590)
	totalWaitDuration, maxWaitDuration := l.totalAndMaxWaitDuration(now)
	lm := LockMetrics{
		Key:                  l.key,
		Held:                 l.holder.locked,
		HoldDurationNanos:    l.lockHeldDuration(now).Nanoseconds(),
		WaitingReaders:       int64(l.waitingReaders.Len()),
		WaitingWriters:       int64(l.queuedWriters.Len()),
		WaitDurationNanos:    totalWaitDuration.Nanoseconds(),
		MaxWaitDurationNanos: maxWaitDuration.Nanoseconds(),
	}
	lm.Waiters = lm.WaitingReaders + lm.WaitingWriters
	m.addLockMetrics(lm)
}

func (l *lockState) tryBreakReservation(seqNum uint64) bool {
	__antithesis_instrumentation__.Notify(99593)
	if l.reservation.seqNum > seqNum {
		__antithesis_instrumentation__.Notify(99595)
		qg := &queuedGuard{
			guard:  l.reservation,
			active: false,
		}
		l.queuedWriters.PushFront(qg)
		l.reservation = nil
		return true
	} else {
		__antithesis_instrumentation__.Notify(99596)
	}
	__antithesis_instrumentation__.Notify(99594)
	return false
}

func (l *lockState) informActiveWaiters() {
	__antithesis_instrumentation__.Notify(99597)
	waitForState := waitingState{
		kind:          waitFor,
		key:           l.key,
		queuedWriters: l.queuedWriters.Len(),
		queuedReaders: l.waitingReaders.Len(),
	}
	findDistinguished := l.distinguishedWaiter == nil
	if lockHolderTxn, _ := l.getLockHolder(); lockHolderTxn != nil {
		__antithesis_instrumentation__.Notify(99600)
		waitForState.txn = lockHolderTxn
		waitForState.held = true
	} else {
		__antithesis_instrumentation__.Notify(99601)
		waitForState.txn = l.reservation.txn
		if !findDistinguished && func() bool {
			__antithesis_instrumentation__.Notify(99602)
			return l.distinguishedWaiter.isSameTxnAsReservation(waitForState) == true
		}() == true {
			__antithesis_instrumentation__.Notify(99603)
			findDistinguished = true
			l.distinguishedWaiter = nil
		} else {
			__antithesis_instrumentation__.Notify(99604)
		}
	}
	__antithesis_instrumentation__.Notify(99598)

	for e := l.waitingReaders.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(99605)
		state := waitForState
		state.guardAccess = spanset.SpanReadOnly

		g := e.Value.(*lockTableGuardImpl)
		if findDistinguished {
			__antithesis_instrumentation__.Notify(99608)
			l.distinguishedWaiter = g
			findDistinguished = false
		} else {
			__antithesis_instrumentation__.Notify(99609)
		}
		__antithesis_instrumentation__.Notify(99606)
		g.mu.Lock()
		g.updateStateLocked(state)
		if l.distinguishedWaiter == g {
			__antithesis_instrumentation__.Notify(99610)
			g.mu.state.kind = waitForDistinguished
		} else {
			__antithesis_instrumentation__.Notify(99611)
		}
		__antithesis_instrumentation__.Notify(99607)
		g.notify()
		g.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(99599)
	for e := l.queuedWriters.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(99612)
		qg := e.Value.(*queuedGuard)
		if !qg.active {
			__antithesis_instrumentation__.Notify(99615)
			continue
		} else {
			__antithesis_instrumentation__.Notify(99616)
		}
		__antithesis_instrumentation__.Notify(99613)
		g := qg.guard
		state := waitForState
		if g.isSameTxnAsReservation(state) {
			__antithesis_instrumentation__.Notify(99617)
			state.kind = waitSelf
		} else {
			__antithesis_instrumentation__.Notify(99618)
			state.guardAccess = spanset.SpanReadWrite
			if findDistinguished {
				__antithesis_instrumentation__.Notify(99620)
				l.distinguishedWaiter = g
				findDistinguished = false
			} else {
				__antithesis_instrumentation__.Notify(99621)
			}
			__antithesis_instrumentation__.Notify(99619)
			if l.distinguishedWaiter == g {
				__antithesis_instrumentation__.Notify(99622)
				state.kind = waitForDistinguished
			} else {
				__antithesis_instrumentation__.Notify(99623)
			}
		}
		__antithesis_instrumentation__.Notify(99614)
		g.mu.Lock()
		g.updateStateLocked(state)
		g.notify()
		g.mu.Unlock()
	}
}

func (l *lockState) releaseWritersFromTxn(txn *enginepb.TxnMeta) {
	__antithesis_instrumentation__.Notify(99624)
	for e := l.queuedWriters.Front(); e != nil; {
		__antithesis_instrumentation__.Notify(99625)
		qg := e.Value.(*queuedGuard)
		curr := e
		e = e.Next()
		g := qg.guard
		if g.isSameTxn(txn) {
			__antithesis_instrumentation__.Notify(99626)
			if qg.active {
				__antithesis_instrumentation__.Notify(99628)
				if g == l.distinguishedWaiter {
					__antithesis_instrumentation__.Notify(99630)
					l.distinguishedWaiter = nil
				} else {
					__antithesis_instrumentation__.Notify(99631)
				}
				__antithesis_instrumentation__.Notify(99629)
				g.doneWaitingAtLock(false, l)
			} else {
				__antithesis_instrumentation__.Notify(99632)
				g.mu.Lock()
				delete(g.mu.locks, l)
				g.mu.Unlock()
			}
			__antithesis_instrumentation__.Notify(99627)
			l.queuedWriters.Remove(curr)
		} else {
			__antithesis_instrumentation__.Notify(99633)
		}
	}
}

func (l *lockState) tryMakeNewDistinguished() {
	__antithesis_instrumentation__.Notify(99634)
	var g *lockTableGuardImpl
	if l.waitingReaders.Len() > 0 {
		__antithesis_instrumentation__.Notify(99636)
		g = l.waitingReaders.Front().Value.(*lockTableGuardImpl)
	} else {
		__antithesis_instrumentation__.Notify(99637)
		if l.queuedWriters.Len() > 0 {
			__antithesis_instrumentation__.Notify(99638)
			for e := l.queuedWriters.Front(); e != nil; e = e.Next() {
				__antithesis_instrumentation__.Notify(99639)
				qg := e.Value.(*queuedGuard)
				if qg.active && func() bool {
					__antithesis_instrumentation__.Notify(99640)
					return (l.reservation == nil || func() bool {
						__antithesis_instrumentation__.Notify(99641)
						return !qg.guard.isSameTxn(l.reservation.txn) == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(99642)
					g = qg.guard
					break
				} else {
					__antithesis_instrumentation__.Notify(99643)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(99644)
		}
	}
	__antithesis_instrumentation__.Notify(99635)
	if g != nil {
		__antithesis_instrumentation__.Notify(99645)
		l.distinguishedWaiter = g
		g.mu.Lock()
		g.mu.state.kind = waitForDistinguished

		g.notify()
		g.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(99646)
	}
}

func (l *lockState) isEmptyLock() bool {
	__antithesis_instrumentation__.Notify(99647)
	if !l.holder.locked && func() bool {
		__antithesis_instrumentation__.Notify(99649)
		return l.reservation == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(99650)
		for i := range l.holder.holder {
			__antithesis_instrumentation__.Notify(99653)
			if !l.holder.holder[i].isEmpty() {
				__antithesis_instrumentation__.Notify(99654)
				panic("lockState with !locked but non-zero lockHolderInfo")
			} else {
				__antithesis_instrumentation__.Notify(99655)
			}
		}
		__antithesis_instrumentation__.Notify(99651)
		if l.waitingReaders.Len() > 0 || func() bool {
			__antithesis_instrumentation__.Notify(99656)
			return l.queuedWriters.Len() > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(99657)
			panic("lockState with waiters but no holder or reservation")
		} else {
			__antithesis_instrumentation__.Notify(99658)
		}
		__antithesis_instrumentation__.Notify(99652)
		return true
	} else {
		__antithesis_instrumentation__.Notify(99659)
	}
	__antithesis_instrumentation__.Notify(99648)
	return false
}

func (l *lockState) lockHeldDuration(now time.Time) time.Duration {
	__antithesis_instrumentation__.Notify(99660)
	if !l.holder.locked {
		__antithesis_instrumentation__.Notify(99662)
		return time.Duration(0)
	} else {
		__antithesis_instrumentation__.Notify(99663)
	}
	__antithesis_instrumentation__.Notify(99661)

	return now.Sub(l.holder.startTime)
}

func (l *lockState) totalAndMaxWaitDuration(now time.Time) (time.Duration, time.Duration) {
	__antithesis_instrumentation__.Notify(99664)
	var totalWaitDuration time.Duration
	var maxWaitDuration time.Duration
	for e := l.waitingReaders.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(99667)
		g := e.Value.(*lockTableGuardImpl)
		g.mu.Lock()
		waitDuration := now.Sub(g.mu.curLockWaitStart)
		totalWaitDuration += waitDuration
		if waitDuration > maxWaitDuration {
			__antithesis_instrumentation__.Notify(99669)
			maxWaitDuration = waitDuration
		} else {
			__antithesis_instrumentation__.Notify(99670)
		}
		__antithesis_instrumentation__.Notify(99668)
		g.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(99665)
	for e := l.queuedWriters.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(99671)
		qg := e.Value.(*queuedGuard)
		g := qg.guard
		g.mu.Lock()
		waitDuration := now.Sub(g.mu.curLockWaitStart)
		totalWaitDuration += waitDuration
		if waitDuration > maxWaitDuration {
			__antithesis_instrumentation__.Notify(99673)
			maxWaitDuration = waitDuration
		} else {
			__antithesis_instrumentation__.Notify(99674)
		}
		__antithesis_instrumentation__.Notify(99672)
		g.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(99666)
	return totalWaitDuration, maxWaitDuration
}

func (l *lockState) isLockedBy(id uuid.UUID) bool {
	__antithesis_instrumentation__.Notify(99675)
	if l.holder.locked {
		__antithesis_instrumentation__.Notify(99677)
		var holderID uuid.UUID
		if l.holder.holder[lock.Unreplicated].txn != nil {
			__antithesis_instrumentation__.Notify(99679)
			holderID = l.holder.holder[lock.Unreplicated].txn.ID
		} else {
			__antithesis_instrumentation__.Notify(99680)
			holderID = l.holder.holder[lock.Replicated].txn.ID
		}
		__antithesis_instrumentation__.Notify(99678)
		return id == holderID
	} else {
		__antithesis_instrumentation__.Notify(99681)
	}
	__antithesis_instrumentation__.Notify(99676)
	return false
}

func (l *lockState) getLockHolder() (*enginepb.TxnMeta, hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(99682)
	if !l.holder.locked {
		__antithesis_instrumentation__.Notify(99685)
		return nil, hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(99686)
	}
	__antithesis_instrumentation__.Notify(99683)

	index := lock.Replicated

	if l.holder.holder[index].txn == nil || func() bool {
		__antithesis_instrumentation__.Notify(99687)
		return (l.holder.holder[lock.Unreplicated].txn != nil && func() bool {
			__antithesis_instrumentation__.Notify(99688)
			return l.holder.holder[lock.Unreplicated].ts.Less(l.holder.holder[lock.Replicated].ts) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(99689)
		index = lock.Unreplicated
	} else {
		__antithesis_instrumentation__.Notify(99690)
	}
	__antithesis_instrumentation__.Notify(99684)
	return l.holder.holder[index].txn, l.holder.holder[index].ts
}

func (l *lockState) clearLockHolder() {
	__antithesis_instrumentation__.Notify(99691)
	l.holder.locked = false
	l.holder.startTime = time.Time{}
	for i := range l.holder.holder {
		__antithesis_instrumentation__.Notify(99692)
		l.holder.holder[i] = lockHolderInfo{}
	}
}

func (l *lockState) tryActiveWait(
	g *lockTableGuardImpl, sa spanset.SpanAccess, notify bool, clock *hlc.Clock,
) (wait bool, transitionedToFree bool) {
	__antithesis_instrumentation__.Notify(99693)
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isEmptyLock() {
		__antithesis_instrumentation__.Notify(99704)
		return false, false
	} else {
		__antithesis_instrumentation__.Notify(99705)
	}
	__antithesis_instrumentation__.Notify(99694)

	lockHolderTxn, lockHolderTS := l.getLockHolder()
	if lockHolderTxn != nil && func() bool {
		__antithesis_instrumentation__.Notify(99706)
		return g.isSameTxn(lockHolderTxn) == true
	}() == true {
		__antithesis_instrumentation__.Notify(99707)

		return false, false
	} else {
		__antithesis_instrumentation__.Notify(99708)
	}
	__antithesis_instrumentation__.Notify(99695)

	var replicatedLockFinalizedTxn *roachpb.Transaction
	if lockHolderTxn != nil {
		__antithesis_instrumentation__.Notify(99709)
		finalizedTxn, ok := g.lt.finalizedTxnCache.get(lockHolderTxn.ID)
		if ok {
			__antithesis_instrumentation__.Notify(99710)
			if l.holder.holder[lock.Replicated].txn == nil {
				__antithesis_instrumentation__.Notify(99711)

				l.clearLockHolder()
				if l.lockIsFree() {
					__antithesis_instrumentation__.Notify(99713)

					return false, true
				} else {
					__antithesis_instrumentation__.Notify(99714)
				}
				__antithesis_instrumentation__.Notify(99712)
				lockHolderTxn = nil

			} else {
				__antithesis_instrumentation__.Notify(99715)
				replicatedLockFinalizedTxn = finalizedTxn
			}
		} else {
			__antithesis_instrumentation__.Notify(99716)
		}
	} else {
		__antithesis_instrumentation__.Notify(99717)
	}
	__antithesis_instrumentation__.Notify(99696)

	if sa == spanset.SpanReadOnly {
		__antithesis_instrumentation__.Notify(99718)
		if lockHolderTxn == nil {
			__antithesis_instrumentation__.Notify(99721)

			return false, false
		} else {
			__antithesis_instrumentation__.Notify(99722)
		}
		__antithesis_instrumentation__.Notify(99719)

		if g.ts.Less(lockHolderTS) {
			__antithesis_instrumentation__.Notify(99723)
			return false, false
		} else {
			__antithesis_instrumentation__.Notify(99724)
		}
		__antithesis_instrumentation__.Notify(99720)
		g.mu.Lock()
		_, alsoHasStrongerAccess := g.mu.locks[l]
		g.mu.Unlock()

		if alsoHasStrongerAccess {
			__antithesis_instrumentation__.Notify(99725)
			return false, false
		} else {
			__antithesis_instrumentation__.Notify(99726)
		}
	} else {
		__antithesis_instrumentation__.Notify(99727)
	}
	__antithesis_instrumentation__.Notify(99697)

	waitForState := waitingState{
		kind:          waitFor,
		key:           l.key,
		queuedWriters: l.queuedWriters.Len(),
		queuedReaders: l.waitingReaders.Len(),
		guardAccess:   sa,
	}
	if lockHolderTxn != nil {
		__antithesis_instrumentation__.Notify(99728)
		waitForState.txn = lockHolderTxn
		waitForState.held = true
	} else {
		__antithesis_instrumentation__.Notify(99729)
		if l.reservation == g {
			__antithesis_instrumentation__.Notify(99732)

			return false, false
		} else {
			__antithesis_instrumentation__.Notify(99733)
		}
		__antithesis_instrumentation__.Notify(99730)

		if g.txn == nil && func() bool {
			__antithesis_instrumentation__.Notify(99734)
			return l.reservation.seqNum > g.seqNum == true
		}() == true {
			__antithesis_instrumentation__.Notify(99735)

			return false, false
		} else {
			__antithesis_instrumentation__.Notify(99736)
		}
		__antithesis_instrumentation__.Notify(99731)
		waitForState.txn = l.reservation.txn
	}
	__antithesis_instrumentation__.Notify(99698)

	if l.reservation != nil && func() bool {
		__antithesis_instrumentation__.Notify(99737)
		return sa == spanset.SpanReadWrite == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(99738)
		return l.tryBreakReservation(g.seqNum) == true
	}() == true {
		__antithesis_instrumentation__.Notify(99739)
		l.reservation = g
		g.mu.Lock()
		g.mu.locks[l] = struct{}{}
		g.mu.Unlock()

		l.informActiveWaiters()
		return false, false
	} else {
		__antithesis_instrumentation__.Notify(99740)
	}
	__antithesis_instrumentation__.Notify(99699)

	wait = true
	g.mu.Lock()
	defer g.mu.Unlock()
	if sa == spanset.SpanReadWrite {
		__antithesis_instrumentation__.Notify(99741)
		var qg *queuedGuard
		if _, inQueue := g.mu.locks[l]; inQueue {
			__antithesis_instrumentation__.Notify(99743)

			for e := l.queuedWriters.Front(); e != nil; e = e.Next() {
				__antithesis_instrumentation__.Notify(99746)
				qqg := e.Value.(*queuedGuard)
				if qqg.guard == g {
					__antithesis_instrumentation__.Notify(99747)
					qg = qqg
					break
				} else {
					__antithesis_instrumentation__.Notify(99748)
				}
			}
			__antithesis_instrumentation__.Notify(99744)
			if qg == nil {
				__antithesis_instrumentation__.Notify(99749)
				panic("lockTable bug")
			} else {
				__antithesis_instrumentation__.Notify(99750)
			}
			__antithesis_instrumentation__.Notify(99745)

			qg.active = true
		} else {
			__antithesis_instrumentation__.Notify(99751)

			qg = &queuedGuard{
				guard:  g,
				active: true,
			}
			if curLen := l.queuedWriters.Len(); curLen == 0 {
				__antithesis_instrumentation__.Notify(99753)
				l.queuedWriters.PushFront(qg)
			} else {
				__antithesis_instrumentation__.Notify(99754)
				if g.maxWaitQueueLength > 0 && func() bool {
					__antithesis_instrumentation__.Notify(99755)
					return curLen >= g.maxWaitQueueLength == true
				}() == true {
					__antithesis_instrumentation__.Notify(99756)

					g.mu.startWait = true
					state := waitForState
					state.kind = waitQueueMaxLengthExceeded
					g.updateStateLocked(state)
					if notify {
						__antithesis_instrumentation__.Notify(99758)
						g.notify()
					} else {
						__antithesis_instrumentation__.Notify(99759)
					}
					__antithesis_instrumentation__.Notify(99757)

					return true, false
				} else {
					__antithesis_instrumentation__.Notify(99760)
					var e *list.Element
					for e = l.queuedWriters.Back(); e != nil; e = e.Prev() {
						__antithesis_instrumentation__.Notify(99762)
						qqg := e.Value.(*queuedGuard)
						if qqg.guard.seqNum < qg.guard.seqNum {
							__antithesis_instrumentation__.Notify(99763)
							break
						} else {
							__antithesis_instrumentation__.Notify(99764)
						}
					}
					__antithesis_instrumentation__.Notify(99761)
					if e == nil {
						__antithesis_instrumentation__.Notify(99765)
						l.queuedWriters.PushFront(qg)
					} else {
						__antithesis_instrumentation__.Notify(99766)
						l.queuedWriters.InsertAfter(qg, e)
					}
				}
			}
			__antithesis_instrumentation__.Notify(99752)
			g.mu.locks[l] = struct{}{}
			waitForState.queuedWriters = l.queuedWriters.Len()
		}
		__antithesis_instrumentation__.Notify(99742)
		if replicatedLockFinalizedTxn != nil && func() bool {
			__antithesis_instrumentation__.Notify(99767)
			return l.queuedWriters.Front().Value.(*queuedGuard) == qg == true
		}() == true {
			__antithesis_instrumentation__.Notify(99768)

			qg.active = false
			wait = false
		} else {
			__antithesis_instrumentation__.Notify(99769)
		}
	} else {
		__antithesis_instrumentation__.Notify(99770)
		if replicatedLockFinalizedTxn != nil {
			__antithesis_instrumentation__.Notify(99771)

			wait = false
		} else {
			__antithesis_instrumentation__.Notify(99772)
			l.waitingReaders.PushFront(g)
			g.mu.locks[l] = struct{}{}
			waitForState.queuedReaders = l.waitingReaders.Len()
		}
	}
	__antithesis_instrumentation__.Notify(99700)
	if !wait {
		__antithesis_instrumentation__.Notify(99773)
		g.toResolve = append(
			g.toResolve, roachpb.MakeLockUpdate(replicatedLockFinalizedTxn, roachpb.Span{Key: l.key}))
		return false, false
	} else {
		__antithesis_instrumentation__.Notify(99774)
	}
	__antithesis_instrumentation__.Notify(99701)

	g.key = l.key
	g.mu.startWait = true
	g.mu.curLockWaitStart = clock.PhysicalTime()
	if g.isSameTxnAsReservation(waitForState) {
		__antithesis_instrumentation__.Notify(99775)
		state := waitForState
		state.kind = waitSelf
		g.updateStateLocked(state)
	} else {
		__antithesis_instrumentation__.Notify(99776)
		state := waitForState
		if l.distinguishedWaiter == nil {
			__antithesis_instrumentation__.Notify(99778)
			l.distinguishedWaiter = g
			state.kind = waitForDistinguished
		} else {
			__antithesis_instrumentation__.Notify(99779)
		}
		__antithesis_instrumentation__.Notify(99777)
		g.updateStateLocked(state)
	}
	__antithesis_instrumentation__.Notify(99702)
	if notify {
		__antithesis_instrumentation__.Notify(99780)
		g.notify()
	} else {
		__antithesis_instrumentation__.Notify(99781)
	}
	__antithesis_instrumentation__.Notify(99703)
	return true, false
}

func (l *lockState) isNonConflictingLock(g *lockTableGuardImpl, sa spanset.SpanAccess) bool {
	__antithesis_instrumentation__.Notify(99782)
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.isEmptyLock() {
		__antithesis_instrumentation__.Notify(99787)
		return true
	} else {
		__antithesis_instrumentation__.Notify(99788)
	}
	__antithesis_instrumentation__.Notify(99783)

	lockHolderTxn, lockHolderTS := l.getLockHolder()
	if lockHolderTxn == nil {
		__antithesis_instrumentation__.Notify(99789)

		return true
	} else {
		__antithesis_instrumentation__.Notify(99790)
	}
	__antithesis_instrumentation__.Notify(99784)
	if g.isSameTxn(lockHolderTxn) {
		__antithesis_instrumentation__.Notify(99791)

		return true
	} else {
		__antithesis_instrumentation__.Notify(99792)
	}
	__antithesis_instrumentation__.Notify(99785)

	if sa == spanset.SpanReadOnly && func() bool {
		__antithesis_instrumentation__.Notify(99793)
		return g.ts.Less(lockHolderTS) == true
	}() == true {
		__antithesis_instrumentation__.Notify(99794)
		return true
	} else {
		__antithesis_instrumentation__.Notify(99795)
	}
	__antithesis_instrumentation__.Notify(99786)

	return false
}

func (l *lockState) acquireLock(
	_ lock.Strength,
	durability lock.Durability,
	txn *enginepb.TxnMeta,
	ts hlc.Timestamp,
	clock *hlc.Clock,
) error {
	__antithesis_instrumentation__.Notify(99796)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.holder.locked {
		__antithesis_instrumentation__.Notify(99799)

		beforeTxn, beforeTs := l.getLockHolder()
		if txn.ID != beforeTxn.ID {
			__antithesis_instrumentation__.Notify(99804)
			return errors.AssertionFailedf("existing lock cannot be acquired by different transaction")
		} else {
			__antithesis_instrumentation__.Notify(99805)
		}
		__antithesis_instrumentation__.Notify(99800)
		seqs := l.holder.holder[durability].seqs
		if l.holder.holder[durability].txn != nil && func() bool {
			__antithesis_instrumentation__.Notify(99806)
			return l.holder.holder[durability].txn.Epoch < txn.Epoch == true
		}() == true {
			__antithesis_instrumentation__.Notify(99807)

			seqs = seqs[:0]
		} else {
			__antithesis_instrumentation__.Notify(99808)
		}
		__antithesis_instrumentation__.Notify(99801)
		if len(seqs) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(99809)
			return seqs[len(seqs)-1] >= txn.Sequence == true
		}() == true {
			__antithesis_instrumentation__.Notify(99810)

			if i := sort.Search(len(seqs), func(i int) bool {
				__antithesis_instrumentation__.Notify(99812)
				return seqs[i] >= txn.Sequence
			}); i == len(seqs) {
				__antithesis_instrumentation__.Notify(99813)
				panic("lockTable bug - search value <= last element")
			} else {
				__antithesis_instrumentation__.Notify(99814)
				if seqs[i] != txn.Sequence {
					__antithesis_instrumentation__.Notify(99815)
					seqs = append(seqs, 0)
					copy(seqs[i+1:], seqs[i:])
					seqs[i] = txn.Sequence
					l.holder.holder[durability].seqs = seqs
				} else {
					__antithesis_instrumentation__.Notify(99816)
				}
			}
			__antithesis_instrumentation__.Notify(99811)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(99817)
		}
		__antithesis_instrumentation__.Notify(99802)
		l.holder.holder[durability].txn = txn

		l.holder.holder[durability].ts.Forward(ts)
		l.holder.holder[durability].seqs = append(seqs, txn.Sequence)

		_, afterTs := l.getLockHolder()
		if beforeTs.Less(afterTs) {
			__antithesis_instrumentation__.Notify(99818)
			l.increasedLockTs(afterTs)
		} else {
			__antithesis_instrumentation__.Notify(99819)
		}
		__antithesis_instrumentation__.Notify(99803)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(99820)
	}
	__antithesis_instrumentation__.Notify(99797)

	if l.reservation != nil {
		__antithesis_instrumentation__.Notify(99821)
		if l.reservation.txn.ID != txn.ID {
			__antithesis_instrumentation__.Notify(99823)

			qg := &queuedGuard{
				guard:  l.reservation,
				active: false,
			}
			l.queuedWriters.PushFront(qg)
		} else {
			__antithesis_instrumentation__.Notify(99824)

			l.reservation.mu.Lock()
			delete(l.reservation.mu.locks, l)
			l.reservation.mu.Unlock()
		}
		__antithesis_instrumentation__.Notify(99822)
		if l.waitingReaders.Len() > 0 {
			__antithesis_instrumentation__.Notify(99825)
			panic("lockTable bug")
		} else {
			__antithesis_instrumentation__.Notify(99826)
		}
	} else {
		__antithesis_instrumentation__.Notify(99827)
		if l.queuedWriters.Len() > 0 || func() bool {
			__antithesis_instrumentation__.Notify(99828)
			return l.waitingReaders.Len() > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(99829)
			panic("lockTable bug")
		} else {
			__antithesis_instrumentation__.Notify(99830)
		}
	}
	__antithesis_instrumentation__.Notify(99798)
	l.reservation = nil
	l.holder.locked = true
	l.holder.holder[durability].txn = txn
	l.holder.holder[durability].ts = ts
	l.holder.holder[durability].seqs = append([]enginepb.TxnSeq(nil), txn.Sequence)
	l.holder.startTime = clock.PhysicalTime()

	l.releaseWritersFromTxn(txn)

	l.informActiveWaiters()
	return nil
}

func (l *lockState) discoveredLock(
	txn *enginepb.TxnMeta,
	ts hlc.Timestamp,
	g *lockTableGuardImpl,
	sa spanset.SpanAccess,
	notRemovable bool,
	clock *hlc.Clock,
) error {
	__antithesis_instrumentation__.Notify(99831)
	l.mu.Lock()
	defer l.mu.Unlock()

	if notRemovable {
		__antithesis_instrumentation__.Notify(99837)
		l.notRemovable++
	} else {
		__antithesis_instrumentation__.Notify(99838)
	}
	__antithesis_instrumentation__.Notify(99832)
	if l.holder.locked {
		__antithesis_instrumentation__.Notify(99839)
		if !l.isLockedBy(txn.ID) {
			__antithesis_instrumentation__.Notify(99840)
			return errors.AssertionFailedf(
				"discovered lock by different transaction (%s) than existing lock (see issue #63592): %s",
				txn, l)
		} else {
			__antithesis_instrumentation__.Notify(99841)
		}
	} else {
		__antithesis_instrumentation__.Notify(99842)
		l.holder.locked = true
		l.holder.startTime = clock.PhysicalTime()
	}
	__antithesis_instrumentation__.Notify(99833)
	holder := &l.holder.holder[lock.Replicated]
	if holder.txn == nil {
		__antithesis_instrumentation__.Notify(99843)
		holder.txn = txn
		holder.ts = ts
		holder.seqs = append(holder.seqs, txn.Sequence)
	} else {
		__antithesis_instrumentation__.Notify(99844)
	}
	__antithesis_instrumentation__.Notify(99834)

	if l.reservation != nil {
		__antithesis_instrumentation__.Notify(99845)
		qg := &queuedGuard{
			guard:  l.reservation,
			active: false,
		}
		l.queuedWriters.PushFront(qg)
		l.reservation = nil
	} else {
		__antithesis_instrumentation__.Notify(99846)
	}
	__antithesis_instrumentation__.Notify(99835)

	switch sa {
	case spanset.SpanReadOnly:
		__antithesis_instrumentation__.Notify(99847)

		if g.ts.Less(ts) {
			__antithesis_instrumentation__.Notify(99851)
			return errors.AssertionFailedf("discovered non-conflicting lock")
		} else {
			__antithesis_instrumentation__.Notify(99852)
		}

	case spanset.SpanReadWrite:
		__antithesis_instrumentation__.Notify(99848)

		g.mu.Lock()
		_, presentHere := g.mu.locks[l]
		if !presentHere {
			__antithesis_instrumentation__.Notify(99853)

			g.mu.locks[l] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(99854)
		}
		__antithesis_instrumentation__.Notify(99849)
		g.mu.Unlock()

		if !presentHere {
			__antithesis_instrumentation__.Notify(99855)

			qg := &queuedGuard{
				guard:  g,
				active: false,
			}

			var e *list.Element
			for e = l.queuedWriters.Front(); e != nil; e = e.Next() {
				__antithesis_instrumentation__.Notify(99857)
				qqg := e.Value.(*queuedGuard)
				if qqg.guard.seqNum > g.seqNum {
					__antithesis_instrumentation__.Notify(99858)
					break
				} else {
					__antithesis_instrumentation__.Notify(99859)
				}
			}
			__antithesis_instrumentation__.Notify(99856)
			if e == nil {
				__antithesis_instrumentation__.Notify(99860)
				l.queuedWriters.PushBack(qg)
			} else {
				__antithesis_instrumentation__.Notify(99861)
				l.queuedWriters.InsertBefore(qg, e)
			}
		} else {
			__antithesis_instrumentation__.Notify(99862)
		}
	default:
		__antithesis_instrumentation__.Notify(99850)
	}
	__antithesis_instrumentation__.Notify(99836)

	l.releaseWritersFromTxn(txn)

	l.informActiveWaiters()
	return nil
}

func (l *lockState) decrementNotRemovable() {
	__antithesis_instrumentation__.Notify(99863)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.notRemovable--
	if l.notRemovable < 0 {
		__antithesis_instrumentation__.Notify(99864)
		panic(fmt.Sprintf("lockState.notRemovable is negative: %d", l.notRemovable))
	} else {
		__antithesis_instrumentation__.Notify(99865)
	}
}

func (l *lockState) tryClearLock(force bool) bool {
	__antithesis_instrumentation__.Notify(99866)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.notRemovable > 0 && func() bool {
		__antithesis_instrumentation__.Notify(99872)
		return !force == true
	}() == true {
		__antithesis_instrumentation__.Notify(99873)
		return false
	} else {
		__antithesis_instrumentation__.Notify(99874)
	}
	__antithesis_instrumentation__.Notify(99867)
	replicatedHeld := l.holder.locked && func() bool {
		__antithesis_instrumentation__.Notify(99875)
		return l.holder.holder[lock.Replicated].txn != nil == true
	}() == true

	l.holder.holder[lock.Unreplicated] = lockHolderInfo{}
	var waitState waitingState
	if replicatedHeld && func() bool {
		__antithesis_instrumentation__.Notify(99876)
		return !force == true
	}() == true {
		__antithesis_instrumentation__.Notify(99877)
		lockHolderTxn, _ := l.getLockHolder()

		waitState = waitingState{
			kind:        waitElsewhere,
			txn:         lockHolderTxn,
			key:         l.key,
			held:        true,
			guardAccess: spanset.SpanReadOnly,
		}
	} else {
		__antithesis_instrumentation__.Notify(99878)

		l.clearLockHolder()
		waitState = waitingState{kind: doneWaiting}
	}
	__antithesis_instrumentation__.Notify(99868)

	l.distinguishedWaiter = nil
	if l.reservation != nil {
		__antithesis_instrumentation__.Notify(99879)
		g := l.reservation
		g.mu.Lock()
		delete(g.mu.locks, l)
		g.mu.Unlock()
		l.reservation = nil
	} else {
		__antithesis_instrumentation__.Notify(99880)
	}
	__antithesis_instrumentation__.Notify(99869)
	for e := l.waitingReaders.Front(); e != nil; {
		__antithesis_instrumentation__.Notify(99881)
		g := e.Value.(*lockTableGuardImpl)
		curr := e
		e = e.Next()
		l.waitingReaders.Remove(curr)

		g.mu.Lock()
		g.updateStateLocked(waitState)
		g.notify()
		delete(g.mu.locks, l)
		g.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(99870)

	waitState.guardAccess = spanset.SpanReadWrite
	for e := l.queuedWriters.Front(); e != nil; {
		__antithesis_instrumentation__.Notify(99882)
		qg := e.Value.(*queuedGuard)
		curr := e
		e = e.Next()
		l.queuedWriters.Remove(curr)

		g := qg.guard
		g.mu.Lock()
		if qg.active {
			__antithesis_instrumentation__.Notify(99884)
			g.updateStateLocked(waitState)
			g.notify()
		} else {
			__antithesis_instrumentation__.Notify(99885)
		}
		__antithesis_instrumentation__.Notify(99883)
		delete(g.mu.locks, l)
		g.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(99871)
	return true
}

func removeIgnored(
	heldSeqNums []enginepb.TxnSeq, ignoredSeqNums []enginepb.IgnoredSeqNumRange,
) []enginepb.TxnSeq {
	__antithesis_instrumentation__.Notify(99886)
	if len(ignoredSeqNums) == 0 {
		__antithesis_instrumentation__.Notify(99889)
		return heldSeqNums
	} else {
		__antithesis_instrumentation__.Notify(99890)
	}
	__antithesis_instrumentation__.Notify(99887)
	held := heldSeqNums[:0]
	for _, n := range heldSeqNums {
		__antithesis_instrumentation__.Notify(99891)
		i := sort.Search(len(ignoredSeqNums), func(i int) bool { __antithesis_instrumentation__.Notify(99893); return ignoredSeqNums[i].End >= n })
		__antithesis_instrumentation__.Notify(99892)
		if i == len(ignoredSeqNums) || func() bool {
			__antithesis_instrumentation__.Notify(99894)
			return ignoredSeqNums[i].Start > n == true
		}() == true {
			__antithesis_instrumentation__.Notify(99895)
			held = append(held, n)
		} else {
			__antithesis_instrumentation__.Notify(99896)
		}
	}
	__antithesis_instrumentation__.Notify(99888)
	return held
}

func (l *lockState) tryUpdateLock(up *roachpb.LockUpdate) (heldByTxn, gc bool) {
	__antithesis_instrumentation__.Notify(99897)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.isEmptyLock() {
		__antithesis_instrumentation__.Notify(99904)

		return false, true
	} else {
		__antithesis_instrumentation__.Notify(99905)
	}
	__antithesis_instrumentation__.Notify(99898)
	if !l.isLockedBy(up.Txn.ID) {
		__antithesis_instrumentation__.Notify(99906)
		return false, false
	} else {
		__antithesis_instrumentation__.Notify(99907)
	}
	__antithesis_instrumentation__.Notify(99899)
	if up.Status.IsFinalized() {
		__antithesis_instrumentation__.Notify(99908)
		l.clearLockHolder()
		gc = l.lockIsFree()
		return true, gc
	} else {
		__antithesis_instrumentation__.Notify(99909)
	}
	__antithesis_instrumentation__.Notify(99900)

	txn := &up.Txn
	ts := up.Txn.WriteTimestamp
	_, beforeTs := l.getLockHolder()
	advancedTs := beforeTs.Less(ts)
	isLocked := false
	for i := range l.holder.holder {
		__antithesis_instrumentation__.Notify(99910)
		holder := &l.holder.holder[i]
		if holder.txn == nil {
			__antithesis_instrumentation__.Notify(99915)
			continue
		} else {
			__antithesis_instrumentation__.Notify(99916)
		}
		__antithesis_instrumentation__.Notify(99911)

		if lock.Durability(i) == lock.Replicated || func() bool {
			__antithesis_instrumentation__.Notify(99917)
			return txn.Epoch > holder.txn.Epoch == true
		}() == true {
			__antithesis_instrumentation__.Notify(99918)
			*holder = lockHolderInfo{}
			continue
		} else {
			__antithesis_instrumentation__.Notify(99919)
		}
		__antithesis_instrumentation__.Notify(99912)

		if advancedTs {
			__antithesis_instrumentation__.Notify(99920)

			holder.ts = ts
		} else {
			__antithesis_instrumentation__.Notify(99921)
		}
		__antithesis_instrumentation__.Notify(99913)
		if txn.Epoch == holder.txn.Epoch {
			__antithesis_instrumentation__.Notify(99922)
			holder.seqs = removeIgnored(holder.seqs, up.IgnoredSeqNums)
			if len(holder.seqs) == 0 {
				__antithesis_instrumentation__.Notify(99924)
				*holder = lockHolderInfo{}
				continue
			} else {
				__antithesis_instrumentation__.Notify(99925)
			}
			__antithesis_instrumentation__.Notify(99923)
			if advancedTs {
				__antithesis_instrumentation__.Notify(99926)
				holder.txn = txn
			} else {
				__antithesis_instrumentation__.Notify(99927)
			}
		} else {
			__antithesis_instrumentation__.Notify(99928)
		}
		__antithesis_instrumentation__.Notify(99914)

		isLocked = true
	}
	__antithesis_instrumentation__.Notify(99901)

	if !isLocked {
		__antithesis_instrumentation__.Notify(99929)
		l.clearLockHolder()
		gc = l.lockIsFree()
		return true, gc
	} else {
		__antithesis_instrumentation__.Notify(99930)
	}
	__antithesis_instrumentation__.Notify(99902)

	if advancedTs {
		__antithesis_instrumentation__.Notify(99931)
		l.increasedLockTs(ts)
	} else {
		__antithesis_instrumentation__.Notify(99932)
	}
	__antithesis_instrumentation__.Notify(99903)

	return true, false
}

func (l *lockState) increasedLockTs(newTs hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(99933)
	distinguishedRemoved := false
	for e := l.waitingReaders.Front(); e != nil; {
		__antithesis_instrumentation__.Notify(99935)
		g := e.Value.(*lockTableGuardImpl)
		curr := e
		e = e.Next()
		if g.ts.Less(newTs) {
			__antithesis_instrumentation__.Notify(99936)

			l.waitingReaders.Remove(curr)
			if g == l.distinguishedWaiter {
				__antithesis_instrumentation__.Notify(99938)
				distinguishedRemoved = true
				l.distinguishedWaiter = nil
			} else {
				__antithesis_instrumentation__.Notify(99939)
			}
			__antithesis_instrumentation__.Notify(99937)
			g.doneWaitingAtLock(false, l)
		} else {
			__antithesis_instrumentation__.Notify(99940)
		}

	}
	__antithesis_instrumentation__.Notify(99934)
	if distinguishedRemoved {
		__antithesis_instrumentation__.Notify(99941)
		l.tryMakeNewDistinguished()
	} else {
		__antithesis_instrumentation__.Notify(99942)
	}
}

func (l *lockState) requestDone(g *lockTableGuardImpl) (gc bool) {
	__antithesis_instrumentation__.Notify(99943)
	l.mu.Lock()
	defer l.mu.Unlock()

	g.mu.Lock()
	if _, present := g.mu.locks[l]; !present {
		__antithesis_instrumentation__.Notify(99950)
		g.mu.Unlock()
		return false
	} else {
		__antithesis_instrumentation__.Notify(99951)
	}
	__antithesis_instrumentation__.Notify(99944)
	delete(g.mu.locks, l)
	g.mu.Unlock()

	if l.reservation == g {
		__antithesis_instrumentation__.Notify(99952)
		l.reservation = nil
		return l.lockIsFree()
	} else {
		__antithesis_instrumentation__.Notify(99953)
	}
	__antithesis_instrumentation__.Notify(99945)

	distinguishedRemoved := false
	doneRemoval := false
	for e := l.queuedWriters.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(99954)
		qg := e.Value.(*queuedGuard)
		if qg.guard == g {
			__antithesis_instrumentation__.Notify(99955)
			l.queuedWriters.Remove(e)
			if qg.guard == l.distinguishedWaiter {
				__antithesis_instrumentation__.Notify(99957)
				distinguishedRemoved = true
				l.distinguishedWaiter = nil
			} else {
				__antithesis_instrumentation__.Notify(99958)
			}
			__antithesis_instrumentation__.Notify(99956)
			doneRemoval = true
			break
		} else {
			__antithesis_instrumentation__.Notify(99959)
		}
	}
	__antithesis_instrumentation__.Notify(99946)
	if !doneRemoval {
		__antithesis_instrumentation__.Notify(99960)
		for e := l.waitingReaders.Front(); e != nil; e = e.Next() {
			__antithesis_instrumentation__.Notify(99961)
			gg := e.Value.(*lockTableGuardImpl)
			if gg == g {
				__antithesis_instrumentation__.Notify(99962)
				l.waitingReaders.Remove(e)
				if g == l.distinguishedWaiter {
					__antithesis_instrumentation__.Notify(99964)
					distinguishedRemoved = true
					l.distinguishedWaiter = nil
				} else {
					__antithesis_instrumentation__.Notify(99965)
				}
				__antithesis_instrumentation__.Notify(99963)
				doneRemoval = true
				break
			} else {
				__antithesis_instrumentation__.Notify(99966)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(99967)
	}
	__antithesis_instrumentation__.Notify(99947)
	if !doneRemoval {
		__antithesis_instrumentation__.Notify(99968)
		panic("lockTable bug")
	} else {
		__antithesis_instrumentation__.Notify(99969)
	}
	__antithesis_instrumentation__.Notify(99948)
	if distinguishedRemoved {
		__antithesis_instrumentation__.Notify(99970)
		l.tryMakeNewDistinguished()
	} else {
		__antithesis_instrumentation__.Notify(99971)
	}
	__antithesis_instrumentation__.Notify(99949)
	return false
}

func (l *lockState) tryFreeLockOnReplicatedAcquire() bool {
	__antithesis_instrumentation__.Notify(99972)
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.holder.locked || func() bool {
		__antithesis_instrumentation__.Notify(99976)
		return l.holder.holder[lock.Replicated].txn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(99977)
		return false
	} else {
		__antithesis_instrumentation__.Notify(99978)
	}
	__antithesis_instrumentation__.Notify(99973)

	if l.queuedWriters.Len() != 0 {
		__antithesis_instrumentation__.Notify(99979)
		return false
	} else {
		__antithesis_instrumentation__.Notify(99980)
	}
	__antithesis_instrumentation__.Notify(99974)

	l.clearLockHolder()
	gc := l.lockIsFree()
	if !gc {
		__antithesis_instrumentation__.Notify(99981)
		panic("expected lockIsFree to return true")
	} else {
		__antithesis_instrumentation__.Notify(99982)
	}
	__antithesis_instrumentation__.Notify(99975)
	return true
}

func (l *lockState) lockIsFree() (gc bool) {
	__antithesis_instrumentation__.Notify(99983)
	if l.holder.locked {
		__antithesis_instrumentation__.Notify(99990)
		panic("called lockIsFree on lock with holder")
	} else {
		__antithesis_instrumentation__.Notify(99991)
	}
	__antithesis_instrumentation__.Notify(99984)
	if l.reservation != nil {
		__antithesis_instrumentation__.Notify(99992)
		panic("called lockIsFree on lock with reservation")
	} else {
		__antithesis_instrumentation__.Notify(99993)
	}
	__antithesis_instrumentation__.Notify(99985)

	for e := l.waitingReaders.Front(); e != nil; {
		__antithesis_instrumentation__.Notify(99994)
		g := e.Value.(*lockTableGuardImpl)
		curr := e
		e = e.Next()
		l.waitingReaders.Remove(curr)
		if g == l.distinguishedWaiter {
			__antithesis_instrumentation__.Notify(99996)
			l.distinguishedWaiter = nil
		} else {
			__antithesis_instrumentation__.Notify(99997)
		}
		__antithesis_instrumentation__.Notify(99995)
		g.doneWaitingAtLock(false, l)
	}
	__antithesis_instrumentation__.Notify(99986)

	for e := l.queuedWriters.Front(); e != nil; {
		__antithesis_instrumentation__.Notify(99998)
		qg := e.Value.(*queuedGuard)
		g := qg.guard
		if g.txn == nil {
			__antithesis_instrumentation__.Notify(99999)
			curr := e
			e = e.Next()
			l.queuedWriters.Remove(curr)
			if qg.active {
				__antithesis_instrumentation__.Notify(100000)
				if g == l.distinguishedWaiter {
					__antithesis_instrumentation__.Notify(100002)
					l.distinguishedWaiter = nil
				} else {
					__antithesis_instrumentation__.Notify(100003)
				}
				__antithesis_instrumentation__.Notify(100001)
				g.doneWaitingAtLock(false, l)
			} else {
				__antithesis_instrumentation__.Notify(100004)
				g.mu.Lock()
				delete(g.mu.locks, l)
				g.mu.Unlock()
			}
		} else {
			__antithesis_instrumentation__.Notify(100005)
			break
		}
	}
	__antithesis_instrumentation__.Notify(99987)

	if l.queuedWriters.Len() == 0 {
		__antithesis_instrumentation__.Notify(100006)
		return true
	} else {
		__antithesis_instrumentation__.Notify(100007)
	}
	__antithesis_instrumentation__.Notify(99988)

	e := l.queuedWriters.Front()
	qg := e.Value.(*queuedGuard)
	g := qg.guard
	l.reservation = g
	l.queuedWriters.Remove(e)
	if qg.active {
		__antithesis_instrumentation__.Notify(100008)
		if g == l.distinguishedWaiter {
			__antithesis_instrumentation__.Notify(100010)
			l.distinguishedWaiter = nil
		} else {
			__antithesis_instrumentation__.Notify(100011)
		}
		__antithesis_instrumentation__.Notify(100009)
		g.doneWaitingAtLock(true, l)
	} else {
		__antithesis_instrumentation__.Notify(100012)
	}
	__antithesis_instrumentation__.Notify(99989)

	l.informActiveWaiters()
	return false
}

func (t *treeMu) nextLockSeqNum() (seqNum uint64, checkMaxLocks bool) {
	__antithesis_instrumentation__.Notify(100013)
	t.lockIDSeqNum++
	checkMaxLocks = t.lockIDSeqNum%t.lockAddMaxLocksCheckInterval == 0
	return t.lockIDSeqNum, checkMaxLocks
}

func (t *lockTableImpl) ScanOptimistic(req Request) lockTableGuard {
	__antithesis_instrumentation__.Notify(100014)
	g := t.newGuardForReq(req)
	t.doSnapshotForGuard(g)
	return g
}

func (t *lockTableImpl) ScanAndEnqueue(req Request, guard lockTableGuard) lockTableGuard {
	__antithesis_instrumentation__.Notify(100015)

	var g *lockTableGuardImpl
	if guard == nil {
		__antithesis_instrumentation__.Notify(100018)
		g = t.newGuardForReq(req)
	} else {
		__antithesis_instrumentation__.Notify(100019)
		g = guard.(*lockTableGuardImpl)
		g.key = nil
		g.sa = spanset.NumSpanAccess - 1
		g.ss = spanset.SpanScope(0)
		g.index = -1
		g.mu.Lock()
		g.mu.startWait = false
		g.mu.mustFindNextLockAfter = false
		g.mu.Unlock()
		g.toResolve = g.toResolve[:0]
	}
	__antithesis_instrumentation__.Notify(100016)
	t.doSnapshotForGuard(g)
	g.findNextLockAfter(true)
	if g.notRemovableLock != nil {
		__antithesis_instrumentation__.Notify(100020)

		g.notRemovableLock.decrementNotRemovable()
		g.notRemovableLock = nil
	} else {
		__antithesis_instrumentation__.Notify(100021)
	}
	__antithesis_instrumentation__.Notify(100017)
	return g
}

func (t *lockTableImpl) newGuardForReq(req Request) *lockTableGuardImpl {
	__antithesis_instrumentation__.Notify(100022)
	g := newLockTableGuardImpl()
	g.seqNum = atomic.AddUint64(&t.seqNum, 1)
	g.lt = t
	g.txn = req.txnMeta()
	g.ts = req.Timestamp
	g.spans = req.LockSpans
	g.maxWaitQueueLength = req.MaxLockWaitQueueLength
	g.sa = spanset.NumSpanAccess - 1
	g.index = -1
	return g
}

func (t *lockTableImpl) doSnapshotForGuard(g *lockTableGuardImpl) {
	__antithesis_instrumentation__.Notify(100023)
	for ss := spanset.SpanScope(0); ss < spanset.NumSpanScope; ss++ {
		__antithesis_instrumentation__.Notify(100024)
		for sa := spanset.SpanAccess(0); sa < spanset.NumSpanAccess; sa++ {
			__antithesis_instrumentation__.Notify(100025)
			if len(g.spans.GetSpans(sa, ss)) > 0 {
				__antithesis_instrumentation__.Notify(100026)

				t.locks[ss].mu.RLock()
				g.tableSnapshot[ss].Reset()
				g.tableSnapshot[ss] = t.locks[ss].Clone()
				t.locks[ss].mu.RUnlock()
				break
			} else {
				__antithesis_instrumentation__.Notify(100027)
			}
		}
	}
}

func (t *lockTableImpl) Dequeue(guard lockTableGuard) {
	__antithesis_instrumentation__.Notify(100028)

	g := guard.(*lockTableGuardImpl)
	defer releaseLockTableGuardImpl(g)
	if g.notRemovableLock != nil {
		__antithesis_instrumentation__.Notify(100032)
		g.notRemovableLock.decrementNotRemovable()
		g.notRemovableLock = nil
	} else {
		__antithesis_instrumentation__.Notify(100033)
	}
	__antithesis_instrumentation__.Notify(100029)
	var candidateLocks []*lockState
	g.mu.Lock()
	for l := range g.mu.locks {
		__antithesis_instrumentation__.Notify(100034)
		candidateLocks = append(candidateLocks, l)
	}
	__antithesis_instrumentation__.Notify(100030)
	g.mu.Unlock()
	var locksToGC [spanset.NumSpanScope][]*lockState
	for _, l := range candidateLocks {
		__antithesis_instrumentation__.Notify(100035)
		if gc := l.requestDone(g); gc {
			__antithesis_instrumentation__.Notify(100036)
			locksToGC[l.ss] = append(locksToGC[l.ss], l)
		} else {
			__antithesis_instrumentation__.Notify(100037)
		}
	}
	__antithesis_instrumentation__.Notify(100031)

	for i := 0; i < len(locksToGC); i++ {
		__antithesis_instrumentation__.Notify(100038)
		if len(locksToGC[i]) > 0 {
			__antithesis_instrumentation__.Notify(100039)
			t.tryGCLocks(&t.locks[i], locksToGC[i])
		} else {
			__antithesis_instrumentation__.Notify(100040)
		}
	}
}

func (t *lockTableImpl) AddDiscoveredLock(
	intent *roachpb.Intent,
	seq roachpb.LeaseSequence,
	consultFinalizedTxnCache bool,
	guard lockTableGuard,
) (added bool, _ error) {
	__antithesis_instrumentation__.Notify(100041)
	t.enabledMu.RLock()
	defer t.enabledMu.RUnlock()
	if !t.enabled {
		__antithesis_instrumentation__.Notify(100049)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(100050)
	}
	__antithesis_instrumentation__.Notify(100042)
	if seq < t.enabledSeq {
		__antithesis_instrumentation__.Notify(100051)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(100052)
		if seq > t.enabledSeq {
			__antithesis_instrumentation__.Notify(100053)

			return false, errors.AssertionFailedf("unexpected lease sequence: %d > %d", seq, t.enabledSeq)
		} else {
			__antithesis_instrumentation__.Notify(100054)
		}
	}
	__antithesis_instrumentation__.Notify(100043)
	g := guard.(*lockTableGuardImpl)
	key := intent.Key
	sa, ss, err := findAccessInSpans(key, g.spans)
	if err != nil {
		__antithesis_instrumentation__.Notify(100055)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(100056)
	}
	__antithesis_instrumentation__.Notify(100044)
	if consultFinalizedTxnCache {
		__antithesis_instrumentation__.Notify(100057)
		finalizedTxn, ok := t.finalizedTxnCache.get(intent.Txn.ID)
		if ok {
			__antithesis_instrumentation__.Notify(100058)
			g.toResolve = append(
				g.toResolve, roachpb.MakeLockUpdate(finalizedTxn, roachpb.Span{Key: key}))
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(100059)
		}
	} else {
		__antithesis_instrumentation__.Notify(100060)
	}
	__antithesis_instrumentation__.Notify(100045)
	var l *lockState
	tree := &t.locks[ss]
	tree.mu.Lock()
	iter := tree.MakeIter()
	iter.FirstOverlap(&lockState{key: key})
	checkMaxLocks := false
	if !iter.Valid() {
		__antithesis_instrumentation__.Notify(100061)
		var lockSeqNum uint64
		lockSeqNum, checkMaxLocks = tree.nextLockSeqNum()
		l = &lockState{id: lockSeqNum, key: key, ss: ss}
		l.queuedWriters.Init()
		l.waitingReaders.Init()
		tree.Set(l)
		atomic.AddInt64(&tree.numLocks, 1)
	} else {
		__antithesis_instrumentation__.Notify(100062)
		l = iter.Cur()
	}
	__antithesis_instrumentation__.Notify(100046)
	notRemovableLock := false
	if g.notRemovableLock == nil {
		__antithesis_instrumentation__.Notify(100063)

		g.notRemovableLock = l
		notRemovableLock = true
	} else {
		__antithesis_instrumentation__.Notify(100064)
	}
	__antithesis_instrumentation__.Notify(100047)
	err = l.discoveredLock(&intent.Txn, intent.Txn.WriteTimestamp, g, sa, notRemovableLock, g.lt.clock)

	tree.mu.Unlock()
	if checkMaxLocks {
		__antithesis_instrumentation__.Notify(100065)
		t.checkMaxLocksAndTryClear()
	} else {
		__antithesis_instrumentation__.Notify(100066)
	}
	__antithesis_instrumentation__.Notify(100048)
	return true, err
}

func (t *lockTableImpl) AcquireLock(
	txn *enginepb.TxnMeta, key roachpb.Key, strength lock.Strength, durability lock.Durability,
) error {
	__antithesis_instrumentation__.Notify(100067)
	t.enabledMu.RLock()
	defer t.enabledMu.RUnlock()
	if !t.enabled {
		__antithesis_instrumentation__.Notify(100073)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(100074)
	}
	__antithesis_instrumentation__.Notify(100068)
	if strength != lock.Exclusive {
		__antithesis_instrumentation__.Notify(100075)
		return errors.AssertionFailedf("lock strength not Exclusive")
	} else {
		__antithesis_instrumentation__.Notify(100076)
	}
	__antithesis_instrumentation__.Notify(100069)
	ss := spanset.SpanGlobal
	if keys.IsLocal(key) {
		__antithesis_instrumentation__.Notify(100077)
		ss = spanset.SpanLocal
	} else {
		__antithesis_instrumentation__.Notify(100078)
	}
	__antithesis_instrumentation__.Notify(100070)
	var l *lockState
	tree := &t.locks[ss]
	tree.mu.Lock()

	iter := tree.MakeIter()
	iter.FirstOverlap(&lockState{key: key})
	checkMaxLocks := false
	if !iter.Valid() {
		__antithesis_instrumentation__.Notify(100079)
		if durability == lock.Replicated {
			__antithesis_instrumentation__.Notify(100081)

			tree.mu.Unlock()
			return nil
		} else {
			__antithesis_instrumentation__.Notify(100082)
		}
		__antithesis_instrumentation__.Notify(100080)
		var lockSeqNum uint64
		lockSeqNum, checkMaxLocks = tree.nextLockSeqNum()
		l = &lockState{id: lockSeqNum, key: key, ss: ss}
		l.queuedWriters.Init()
		l.waitingReaders.Init()
		tree.Set(l)
		atomic.AddInt64(&tree.numLocks, 1)
	} else {
		__antithesis_instrumentation__.Notify(100083)
		l = iter.Cur()
		if durability == lock.Replicated && func() bool {
			__antithesis_instrumentation__.Notify(100084)
			return l.tryFreeLockOnReplicatedAcquire() == true
		}() == true {
			__antithesis_instrumentation__.Notify(100085)

			tree.Delete(l)
			tree.mu.Unlock()
			atomic.AddInt64(&tree.numLocks, -1)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(100086)
		}
	}
	__antithesis_instrumentation__.Notify(100071)
	err := l.acquireLock(strength, durability, txn, txn.WriteTimestamp, t.clock)
	tree.mu.Unlock()

	if checkMaxLocks {
		__antithesis_instrumentation__.Notify(100087)
		t.checkMaxLocksAndTryClear()
	} else {
		__antithesis_instrumentation__.Notify(100088)
	}
	__antithesis_instrumentation__.Notify(100072)
	return err
}

func (t *lockTableImpl) checkMaxLocksAndTryClear() {
	__antithesis_instrumentation__.Notify(100089)
	var totalLocks int64
	for i := 0; i < len(t.locks); i++ {
		__antithesis_instrumentation__.Notify(100091)
		totalLocks += atomic.LoadInt64(&t.locks[i].numLocks)
	}
	__antithesis_instrumentation__.Notify(100090)
	if totalLocks > t.maxLocks {
		__antithesis_instrumentation__.Notify(100092)
		numToClear := totalLocks - t.minLocks
		t.tryClearLocks(false, int(numToClear))
	} else {
		__antithesis_instrumentation__.Notify(100093)
	}
}

func (t *lockTableImpl) lockCountForTesting() int64 {
	__antithesis_instrumentation__.Notify(100094)
	var totalLocks int64
	for i := 0; i < len(t.locks); i++ {
		__antithesis_instrumentation__.Notify(100096)
		totalLocks += atomic.LoadInt64(&t.locks[i].numLocks)
	}
	__antithesis_instrumentation__.Notify(100095)
	return totalLocks
}

func (t *lockTableImpl) tryClearLocks(force bool, numToClear int) {
	__antithesis_instrumentation__.Notify(100097)
	done := false
	clearCount := 0
	for i := 0; i < int(spanset.NumSpanScope) && func() bool {
		__antithesis_instrumentation__.Notify(100098)
		return !done == true
	}() == true; i++ {
		__antithesis_instrumentation__.Notify(100099)
		tree := &t.locks[i]
		tree.mu.Lock()
		var locksToClear []*lockState
		iter := tree.MakeIter()
		for iter.First(); iter.Valid(); iter.Next() {
			__antithesis_instrumentation__.Notify(100102)
			l := iter.Cur()
			if l.tryClearLock(force) {
				__antithesis_instrumentation__.Notify(100103)
				locksToClear = append(locksToClear, l)
				clearCount++
				if !force && func() bool {
					__antithesis_instrumentation__.Notify(100104)
					return clearCount >= numToClear == true
				}() == true {
					__antithesis_instrumentation__.Notify(100105)
					done = true
					break
				} else {
					__antithesis_instrumentation__.Notify(100106)
				}
			} else {
				__antithesis_instrumentation__.Notify(100107)
			}
		}
		__antithesis_instrumentation__.Notify(100100)
		atomic.AddInt64(&tree.numLocks, int64(-len(locksToClear)))
		if tree.Len() == len(locksToClear) {
			__antithesis_instrumentation__.Notify(100108)

			tree.Reset()
		} else {
			__antithesis_instrumentation__.Notify(100109)
			for _, l := range locksToClear {
				__antithesis_instrumentation__.Notify(100110)
				tree.Delete(l)
			}
		}
		__antithesis_instrumentation__.Notify(100101)
		tree.mu.Unlock()
	}
}

func findAccessInSpans(
	key roachpb.Key, spans *spanset.SpanSet,
) (spanset.SpanAccess, spanset.SpanScope, error) {
	__antithesis_instrumentation__.Notify(100111)
	ss := spanset.SpanGlobal
	if keys.IsLocal(key) {
		__antithesis_instrumentation__.Notify(100114)
		ss = spanset.SpanLocal
	} else {
		__antithesis_instrumentation__.Notify(100115)
	}
	__antithesis_instrumentation__.Notify(100112)
	for sa := spanset.NumSpanAccess - 1; sa >= 0; sa-- {
		__antithesis_instrumentation__.Notify(100116)
		s := spans.GetSpans(sa, ss)

		i := sort.Search(len(s), func(i int) bool {
			__antithesis_instrumentation__.Notify(100118)
			return key.Compare(s[i].Key) < 0
		})
		__antithesis_instrumentation__.Notify(100117)
		if i > 0 && func() bool {
			__antithesis_instrumentation__.Notify(100119)
			return ((len(s[i-1].EndKey) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(100120)
				return key.Compare(s[i-1].EndKey) < 0 == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(100121)
				return key.Equal(s[i-1].Key) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(100122)
			return sa, ss, nil
		} else {
			__antithesis_instrumentation__.Notify(100123)
		}
	}
	__antithesis_instrumentation__.Notify(100113)
	return 0, 0, errors.AssertionFailedf("could not find access in spans")
}

func (t *lockTableImpl) tryGCLocks(tree *treeMu, locks []*lockState) {
	__antithesis_instrumentation__.Notify(100124)
	tree.mu.Lock()
	defer tree.mu.Unlock()
	for _, l := range locks {
		__antithesis_instrumentation__.Notify(100125)
		iter := tree.MakeIter()
		iter.FirstOverlap(l)

		if !iter.Valid() {
			__antithesis_instrumentation__.Notify(100127)
			continue
		} else {
			__antithesis_instrumentation__.Notify(100128)
		}
		__antithesis_instrumentation__.Notify(100126)
		l = iter.Cur()
		l.mu.Lock()
		empty := l.isEmptyLock()
		l.mu.Unlock()
		if empty {
			__antithesis_instrumentation__.Notify(100129)
			tree.Delete(l)
			atomic.AddInt64(&tree.numLocks, -1)
		} else {
			__antithesis_instrumentation__.Notify(100130)
		}
	}
}

func (t *lockTableImpl) UpdateLocks(up *roachpb.LockUpdate) error {
	__antithesis_instrumentation__.Notify(100131)
	_ = t.updateLockInternal(up)
	return nil
}

func (t *lockTableImpl) updateLockInternal(up *roachpb.LockUpdate) (heldByTxn bool) {
	__antithesis_instrumentation__.Notify(100132)

	span := up.Span
	ss := spanset.SpanGlobal
	if keys.IsLocal(span.Key) {
		__antithesis_instrumentation__.Notify(100137)
		ss = spanset.SpanLocal
	} else {
		__antithesis_instrumentation__.Notify(100138)
	}
	__antithesis_instrumentation__.Notify(100133)
	tree := &t.locks[ss]
	var locksToGC []*lockState
	heldByTxn = false
	changeFunc := func(l *lockState) {
		__antithesis_instrumentation__.Notify(100139)
		held, gc := l.tryUpdateLock(up)
		heldByTxn = heldByTxn || func() bool {
			__antithesis_instrumentation__.Notify(100140)
			return held == true
		}() == true
		if gc {
			__antithesis_instrumentation__.Notify(100141)
			locksToGC = append(locksToGC, l)
		} else {
			__antithesis_instrumentation__.Notify(100142)
		}
	}
	__antithesis_instrumentation__.Notify(100134)
	tree.mu.RLock()
	iter := tree.MakeIter()
	ltRange := &lockState{key: span.Key, endKey: span.EndKey}
	for iter.FirstOverlap(ltRange); iter.Valid(); iter.NextOverlap(ltRange) {
		__antithesis_instrumentation__.Notify(100143)
		changeFunc(iter.Cur())

		if len(span.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(100144)
			break
		} else {
			__antithesis_instrumentation__.Notify(100145)
		}
	}
	__antithesis_instrumentation__.Notify(100135)
	tree.mu.RUnlock()

	if len(locksToGC) > 0 {
		__antithesis_instrumentation__.Notify(100146)
		t.tryGCLocks(tree, locksToGC)
	} else {
		__antithesis_instrumentation__.Notify(100147)
	}
	__antithesis_instrumentation__.Notify(100136)
	return heldByTxn
}

func stepToNextSpan(g *lockTableGuardImpl) *spanset.Span {
	__antithesis_instrumentation__.Notify(100148)
	g.index++
	for ; g.ss < spanset.NumSpanScope; g.ss++ {
		__antithesis_instrumentation__.Notify(100150)
		for ; g.sa >= 0; g.sa-- {
			__antithesis_instrumentation__.Notify(100152)
			spans := g.spans.GetSpans(g.sa, g.ss)
			if g.index < len(spans) {
				__antithesis_instrumentation__.Notify(100154)
				span := &spans[g.index]
				g.key = span.Key
				return span
			} else {
				__antithesis_instrumentation__.Notify(100155)
			}
			__antithesis_instrumentation__.Notify(100153)
			g.index = 0
		}
		__antithesis_instrumentation__.Notify(100151)
		g.sa = spanset.NumSpanAccess - 1
	}
	__antithesis_instrumentation__.Notify(100149)
	return nil
}

func (t *lockTableImpl) TransactionIsFinalized(txn *roachpb.Transaction) {
	__antithesis_instrumentation__.Notify(100156)

	t.finalizedTxnCache.add(txn)
}

func (t *lockTableImpl) Enable(seq roachpb.LeaseSequence) {
	__antithesis_instrumentation__.Notify(100157)

	t.enabledMu.RLock()
	enabled, enabledSeq := t.enabled, t.enabledSeq
	t.enabledMu.RUnlock()
	if enabled && func() bool {
		__antithesis_instrumentation__.Notify(100159)
		return enabledSeq == seq == true
	}() == true {
		__antithesis_instrumentation__.Notify(100160)
		return
	} else {
		__antithesis_instrumentation__.Notify(100161)
	}
	__antithesis_instrumentation__.Notify(100158)
	t.enabledMu.Lock()
	t.enabled = true
	t.enabledSeq = seq
	t.enabledMu.Unlock()
}

func (t *lockTableImpl) Clear(disable bool) {
	__antithesis_instrumentation__.Notify(100162)

	if disable {
		__antithesis_instrumentation__.Notify(100164)
		t.enabledMu.Lock()
		defer t.enabledMu.Unlock()
		t.enabled = false
	} else {
		__antithesis_instrumentation__.Notify(100165)
	}
	__antithesis_instrumentation__.Notify(100163)

	t.tryClearLocks(true, 0)

	t.finalizedTxnCache.clear()
}

func (t *lockTableImpl) QueryLockTableState(
	span roachpb.Span, opts QueryLockTableOptions,
) ([]roachpb.LockStateInfo, QueryLockTableResumeState) {
	__antithesis_instrumentation__.Notify(100166)
	t.enabledMu.RLock()
	defer t.enabledMu.RUnlock()
	if !t.enabled {
		__antithesis_instrumentation__.Notify(100171)

		return []roachpb.LockStateInfo{}, QueryLockTableResumeState{}
	} else {
		__antithesis_instrumentation__.Notify(100172)
	}
	__antithesis_instrumentation__.Notify(100167)

	var snap btree
	{
		__antithesis_instrumentation__.Notify(100173)
		tree := &t.locks[opts.KeyScope]
		tree.mu.RLock()
		snap = tree.Clone()
		tree.mu.RUnlock()
	}
	__antithesis_instrumentation__.Notify(100168)

	now := t.clock.PhysicalTime()

	lockTableState := make([]roachpb.LockStateInfo, 0, snap.Len())
	resumeState := QueryLockTableResumeState{}
	var numLocks int64
	var numBytes int64
	var nextKey roachpb.Key
	var nextByteSize int64

	iter := snap.MakeIter()
	ltRange := &lockState{key: span.Key, endKey: span.EndKey}
	for iter.FirstOverlap(ltRange); iter.Valid(); iter.NextOverlap(ltRange) {
		__antithesis_instrumentation__.Notify(100174)
		l := iter.Cur()

		if ok, lInfo := l.collectLockStateInfo(opts.IncludeUncontended, now); ok {
			__antithesis_instrumentation__.Notify(100175)
			nextKey = l.key
			nextByteSize = int64(lInfo.Size())
			lInfo.RangeID = t.rID

			if len(lockTableState) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(100177)
				return opts.TargetBytes > 0 == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(100178)
				return (numBytes + nextByteSize) > opts.TargetBytes == true
			}() == true {
				__antithesis_instrumentation__.Notify(100179)
				resumeState.ResumeReason = roachpb.RESUME_BYTE_LIMIT
				break
			} else {
				__antithesis_instrumentation__.Notify(100180)
				if len(lockTableState) > 0 && func() bool {
					__antithesis_instrumentation__.Notify(100181)
					return opts.MaxLocks > 0 == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(100182)
					return numLocks >= opts.MaxLocks == true
				}() == true {
					__antithesis_instrumentation__.Notify(100183)
					resumeState.ResumeReason = roachpb.RESUME_KEY_LIMIT
					break
				} else {
					__antithesis_instrumentation__.Notify(100184)
				}
			}
			__antithesis_instrumentation__.Notify(100176)

			lockTableState = append(lockTableState, lInfo)
			numLocks++
			numBytes += nextByteSize
		} else {
			__antithesis_instrumentation__.Notify(100185)
		}
	}
	__antithesis_instrumentation__.Notify(100169)

	if resumeState.ResumeReason != 0 {
		__antithesis_instrumentation__.Notify(100186)
		resumeState.ResumeNextBytes = nextByteSize
		resumeState.ResumeSpan = &roachpb.Span{Key: nextKey, EndKey: span.EndKey}
	} else {
		__antithesis_instrumentation__.Notify(100187)
	}
	__antithesis_instrumentation__.Notify(100170)
	resumeState.TotalBytes = numBytes

	return lockTableState, resumeState
}

func (t *lockTableImpl) Metrics() LockTableMetrics {
	__antithesis_instrumentation__.Notify(100188)
	var m LockTableMetrics
	for i := 0; i < len(t.locks); i++ {
		__antithesis_instrumentation__.Notify(100190)

		var snap btree
		{
			__antithesis_instrumentation__.Notify(100193)
			tree := &t.locks[i]
			tree.mu.RLock()
			snap = tree.Clone()
			tree.mu.RUnlock()
		}
		__antithesis_instrumentation__.Notify(100191)

		now := t.clock.PhysicalTime()
		iter := snap.MakeIter()
		for iter.First(); iter.Valid(); iter.Next() {
			__antithesis_instrumentation__.Notify(100194)
			iter.Cur().addToMetrics(&m, now)
		}
		__antithesis_instrumentation__.Notify(100192)

		snap.Reset()
	}
	__antithesis_instrumentation__.Notify(100189)
	return m
}

func (t *lockTableImpl) String() string {
	__antithesis_instrumentation__.Notify(100195)
	var sb redact.StringBuilder
	for i := 0; i < len(t.locks); i++ {
		__antithesis_instrumentation__.Notify(100197)
		tree := &t.locks[i]
		scope := spanset.SpanScope(i).String()
		tree.mu.RLock()
		sb.Printf("%s: num=%d\n", scope, atomic.LoadInt64(&tree.numLocks))
		iter := tree.MakeIter()
		for iter.First(); iter.Valid(); iter.Next() {
			__antithesis_instrumentation__.Notify(100199)
			l := iter.Cur()
			l.mu.Lock()
			l.safeFormat(&sb, &t.finalizedTxnCache)
			l.mu.Unlock()
		}
		__antithesis_instrumentation__.Notify(100198)
		tree.mu.RUnlock()
	}
	__antithesis_instrumentation__.Notify(100196)
	return sb.String()
}
