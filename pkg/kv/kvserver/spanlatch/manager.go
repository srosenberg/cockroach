package spanlatch

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/poison"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Manager struct {
	mu      syncutil.Mutex
	idAlloc uint64
	scopes  [spanset.NumSpanScope]scopedManager

	stopper  *stop.Stopper
	slowReqs *metric.Gauge
}

type scopedManager struct {
	readSet latchList
	trees   [spanset.NumSpanAccess]btree
}

func Make(stopper *stop.Stopper, slowReqs *metric.Gauge) Manager {
	__antithesis_instrumentation__.Notify(122730)
	return Manager{
		stopper:  stopper,
		slowReqs: slowReqs,
	}
}

type latch struct {
	*signals
	id         uint64
	span       roachpb.Span
	ts         hlc.Timestamp
	next, prev *latch
}

func (la *latch) inReadSet() bool {
	__antithesis_instrumentation__.Notify(122731)
	return la.next != nil
}

func (la *latch) ID() uint64  { __antithesis_instrumentation__.Notify(122732); return la.id }
func (la *latch) Key() []byte { __antithesis_instrumentation__.Notify(122733); return la.span.Key }
func (la *latch) EndKey() []byte {
	__antithesis_instrumentation__.Notify(122734)
	return la.span.EndKey
}
func (la *latch) String() string {
	__antithesis_instrumentation__.Notify(122735)
	return fmt.Sprintf("%s@%s", la.span, la.ts)
}
func (la *latch) New() *latch     { __antithesis_instrumentation__.Notify(122736); return new(latch) }
func (la *latch) SetID(v uint64)  { __antithesis_instrumentation__.Notify(122737); la.id = v }
func (la *latch) SetKey(v []byte) { __antithesis_instrumentation__.Notify(122738); la.span.Key = v }
func (la *latch) SetEndKey(v []byte) {
	__antithesis_instrumentation__.Notify(122739)
	la.span.EndKey = v
}

type signals struct {
	done   signal
	poison idempotentSignal
}

type Guard struct {
	signals
	pp poison.Policy

	latchesPtrs [spanset.NumSpanScope][spanset.NumSpanAccess]unsafe.Pointer
	latchesLens [spanset.NumSpanScope][spanset.NumSpanAccess]int32

	snap *snapshot
}

func (lg *Guard) latches(s spanset.SpanScope, a spanset.SpanAccess) []latch {
	__antithesis_instrumentation__.Notify(122740)
	len := lg.latchesLens[s][a]
	if len == 0 {
		__antithesis_instrumentation__.Notify(122742)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(122743)
	}
	__antithesis_instrumentation__.Notify(122741)
	const maxArrayLen = 1 << 31
	return (*[maxArrayLen]latch)(lg.latchesPtrs[s][a])[:len:len]
}

func (lg *Guard) setLatches(s spanset.SpanScope, a spanset.SpanAccess, latches []latch) {
	__antithesis_instrumentation__.Notify(122744)
	lg.latchesPtrs[s][a] = unsafe.Pointer(&latches[0])
	lg.latchesLens[s][a] = int32(len(latches))
}

func allocGuardAndLatches(nLatches int) (*Guard, []latch) {
	__antithesis_instrumentation__.Notify(122745)

	if nLatches <= 1 {
		__antithesis_instrumentation__.Notify(122747)
		alloc := new(struct {
			g       Guard
			latches [1]latch
		})
		return &alloc.g, alloc.latches[:nLatches]
	} else {
		__antithesis_instrumentation__.Notify(122748)
		if nLatches <= 2 {
			__antithesis_instrumentation__.Notify(122749)
			alloc := new(struct {
				g       Guard
				latches [2]latch
			})
			return &alloc.g, alloc.latches[:nLatches]
		} else {
			__antithesis_instrumentation__.Notify(122750)
			if nLatches <= 4 {
				__antithesis_instrumentation__.Notify(122751)
				alloc := new(struct {
					g       Guard
					latches [4]latch
				})
				return &alloc.g, alloc.latches[:nLatches]
			} else {
				__antithesis_instrumentation__.Notify(122752)
				if nLatches <= 8 {
					__antithesis_instrumentation__.Notify(122753)
					alloc := new(struct {
						g       Guard
						latches [8]latch
					})
					return &alloc.g, alloc.latches[:nLatches]
				} else {
					__antithesis_instrumentation__.Notify(122754)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(122746)
	return new(Guard), make([]latch, nLatches)
}

func newGuard(spans *spanset.SpanSet, pp poison.Policy) *Guard {
	__antithesis_instrumentation__.Notify(122755)
	nLatches := spans.Len()
	guard, latches := allocGuardAndLatches(nLatches)
	guard.pp = pp
	for s := spanset.SpanScope(0); s < spanset.NumSpanScope; s++ {
		__antithesis_instrumentation__.Notify(122758)
		for a := spanset.SpanAccess(0); a < spanset.NumSpanAccess; a++ {
			__antithesis_instrumentation__.Notify(122759)
			ss := spans.GetSpans(a, s)
			n := len(ss)
			if n == 0 {
				__antithesis_instrumentation__.Notify(122762)
				continue
			} else {
				__antithesis_instrumentation__.Notify(122763)
			}
			__antithesis_instrumentation__.Notify(122760)

			ssLatches := latches[:n]
			for i := range ssLatches {
				__antithesis_instrumentation__.Notify(122764)
				latch := &latches[i]
				latch.span = ss[i].Span
				latch.signals = &guard.signals
				latch.ts = ss[i].Timestamp

			}
			__antithesis_instrumentation__.Notify(122761)
			guard.setLatches(s, a, ssLatches)
			latches = latches[n:]
		}
	}
	__antithesis_instrumentation__.Notify(122756)
	if len(latches) != 0 {
		__antithesis_instrumentation__.Notify(122765)
		panic("alloc too large")
	} else {
		__antithesis_instrumentation__.Notify(122766)
	}
	__antithesis_instrumentation__.Notify(122757)
	return guard
}

func (m *Manager) Acquire(
	ctx context.Context, spans *spanset.SpanSet, pp poison.Policy,
) (*Guard, error) {
	__antithesis_instrumentation__.Notify(122767)
	lg, snap := m.sequence(spans, pp)
	defer snap.close()

	err := m.wait(ctx, lg, snap)
	if err != nil {
		__antithesis_instrumentation__.Notify(122769)
		m.Release(lg)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(122770)
	}
	__antithesis_instrumentation__.Notify(122768)
	return lg, nil
}

func (m *Manager) AcquireOptimistic(spans *spanset.SpanSet, pp poison.Policy) *Guard {
	__antithesis_instrumentation__.Notify(122771)
	lg, snap := m.sequence(spans, pp)
	lg.snap = &snap
	return lg
}

func (m *Manager) WaitFor(ctx context.Context, spans *spanset.SpanSet, pp poison.Policy) error {
	__antithesis_instrumentation__.Notify(122772)

	lg := newGuard(spans, pp)

	m.mu.Lock()
	snap := m.snapshotLocked(spans)
	defer snap.close()
	m.mu.Unlock()

	return m.wait(ctx, lg, snap)
}

func (m *Manager) CheckOptimisticNoConflicts(lg *Guard, spans *spanset.SpanSet) bool {
	__antithesis_instrumentation__.Notify(122773)
	if lg.snap == nil {
		__antithesis_instrumentation__.Notify(122776)
		panic(errors.AssertionFailedf("snap must not be nil"))
	} else {
		__antithesis_instrumentation__.Notify(122777)
	}
	__antithesis_instrumentation__.Notify(122774)
	snap := lg.snap
	var search latch
	for s := spanset.SpanScope(0); s < spanset.NumSpanScope; s++ {
		__antithesis_instrumentation__.Notify(122778)
		tr := &snap.trees[s]
		for a := spanset.SpanAccess(0); a < spanset.NumSpanAccess; a++ {
			__antithesis_instrumentation__.Notify(122779)
			ss := spans.GetSpans(a, s)
			for _, sp := range ss {
				__antithesis_instrumentation__.Notify(122780)
				search.span = sp.Span
				search.ts = sp.Timestamp
				switch a {
				case spanset.SpanReadOnly:
					__antithesis_instrumentation__.Notify(122781)

					it := tr[spanset.SpanReadWrite].MakeIter()
					if overlaps(&it, &search, ignoreLater) {
						__antithesis_instrumentation__.Notify(122785)
						return false
					} else {
						__antithesis_instrumentation__.Notify(122786)
					}
				case spanset.SpanReadWrite:
					__antithesis_instrumentation__.Notify(122782)

					it := tr[spanset.SpanReadWrite].MakeIter()
					if overlaps(&it, &search, ignoreNothing) {
						__antithesis_instrumentation__.Notify(122787)
						return false
					} else {
						__antithesis_instrumentation__.Notify(122788)
					}
					__antithesis_instrumentation__.Notify(122783)

					it = tr[spanset.SpanReadOnly].MakeIter()
					if overlaps(&it, &search, ignoreEarlier) {
						__antithesis_instrumentation__.Notify(122789)
						return false
					} else {
						__antithesis_instrumentation__.Notify(122790)
					}
				default:
					__antithesis_instrumentation__.Notify(122784)
					panic("unknown access")
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(122775)

	return true
}

func overlaps(it *iterator, search *latch, ignore ignoreFn) bool {
	__antithesis_instrumentation__.Notify(122791)
	for it.FirstOverlap(search); it.Valid(); it.NextOverlap(search) {
		__antithesis_instrumentation__.Notify(122793)

		held := it.Cur()
		if !ignore(search.ts, held.ts) {
			__antithesis_instrumentation__.Notify(122794)
			return true
		} else {
			__antithesis_instrumentation__.Notify(122795)
		}
	}
	__antithesis_instrumentation__.Notify(122792)
	return false
}

func (m *Manager) WaitUntilAcquired(ctx context.Context, lg *Guard) (*Guard, error) {
	__antithesis_instrumentation__.Notify(122796)
	if lg.snap == nil {
		__antithesis_instrumentation__.Notify(122800)
		panic(errors.AssertionFailedf("snap must not be nil"))
	} else {
		__antithesis_instrumentation__.Notify(122801)
	}
	__antithesis_instrumentation__.Notify(122797)
	defer func() {
		__antithesis_instrumentation__.Notify(122802)
		lg.snap.close()
		lg.snap = nil
	}()
	__antithesis_instrumentation__.Notify(122798)
	err := m.wait(ctx, lg, *lg.snap)
	if err != nil {
		__antithesis_instrumentation__.Notify(122803)
		m.Release(lg)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(122804)
	}
	__antithesis_instrumentation__.Notify(122799)
	return lg, nil
}

func (m *Manager) sequence(spans *spanset.SpanSet, pp poison.Policy) (*Guard, snapshot) {
	__antithesis_instrumentation__.Notify(122805)
	lg := newGuard(spans, pp)

	m.mu.Lock()
	snap := m.snapshotLocked(spans)
	m.insertLocked(lg)
	m.mu.Unlock()
	return lg, snap
}

type snapshot struct {
	trees [spanset.NumSpanScope][spanset.NumSpanAccess]btree
}

func (sn *snapshot) close() {
	__antithesis_instrumentation__.Notify(122806)
	for s := spanset.SpanScope(0); s < spanset.NumSpanScope; s++ {
		__antithesis_instrumentation__.Notify(122807)
		for a := spanset.SpanAccess(0); a < spanset.NumSpanAccess; a++ {
			__antithesis_instrumentation__.Notify(122808)
			sn.trees[s][a].Reset()
		}
	}
}

func (m *Manager) snapshotLocked(spans *spanset.SpanSet) snapshot {
	__antithesis_instrumentation__.Notify(122809)
	var snap snapshot
	for s := spanset.SpanScope(0); s < spanset.NumSpanScope; s++ {
		__antithesis_instrumentation__.Notify(122811)
		sm := &m.scopes[s]
		reading := len(spans.GetSpans(spanset.SpanReadOnly, s)) > 0
		writing := len(spans.GetSpans(spanset.SpanReadWrite, s)) > 0

		if writing {
			__antithesis_instrumentation__.Notify(122813)
			sm.flushReadSetLocked()
			snap.trees[s][spanset.SpanReadOnly] = sm.trees[spanset.SpanReadOnly].Clone()
		} else {
			__antithesis_instrumentation__.Notify(122814)
		}
		__antithesis_instrumentation__.Notify(122812)
		if writing || func() bool {
			__antithesis_instrumentation__.Notify(122815)
			return reading == true
		}() == true {
			__antithesis_instrumentation__.Notify(122816)
			snap.trees[s][spanset.SpanReadWrite] = sm.trees[spanset.SpanReadWrite].Clone()
		} else {
			__antithesis_instrumentation__.Notify(122817)
		}
	}
	__antithesis_instrumentation__.Notify(122810)
	return snap
}

func (sm *scopedManager) flushReadSetLocked() {
	__antithesis_instrumentation__.Notify(122818)
	for sm.readSet.len > 0 {
		__antithesis_instrumentation__.Notify(122819)
		latch := sm.readSet.front()
		sm.readSet.remove(latch)
		sm.trees[spanset.SpanReadOnly].Set(latch)
	}
}

func (m *Manager) insertLocked(lg *Guard) {
	__antithesis_instrumentation__.Notify(122820)
	for s := spanset.SpanScope(0); s < spanset.NumSpanScope; s++ {
		__antithesis_instrumentation__.Notify(122821)
		sm := &m.scopes[s]
		for a := spanset.SpanAccess(0); a < spanset.NumSpanAccess; a++ {
			__antithesis_instrumentation__.Notify(122822)
			latches := lg.latches(s, a)
			for i := range latches {
				__antithesis_instrumentation__.Notify(122823)
				latch := &latches[i]
				latch.id = m.nextIDLocked()
				switch a {
				case spanset.SpanReadOnly:
					__antithesis_instrumentation__.Notify(122824)

					sm.readSet.pushBack(latch)
				case spanset.SpanReadWrite:
					__antithesis_instrumentation__.Notify(122825)

					sm.trees[spanset.SpanReadWrite].Set(latch)
				default:
					__antithesis_instrumentation__.Notify(122826)
					panic("unknown access")
				}
			}
		}
	}
}

func (m *Manager) nextIDLocked() uint64 {
	__antithesis_instrumentation__.Notify(122827)
	m.idAlloc++
	return m.idAlloc
}

type ignoreFn func(ts, other hlc.Timestamp) bool

func ignoreLater(ts, other hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(122828)
	return !ts.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(122829)
		return ts.Less(other) == true
	}() == true
}
func ignoreEarlier(ts, other hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(122830)
	return !other.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(122831)
		return other.Less(ts) == true
	}() == true
}
func ignoreNothing(ts, other hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(122832)
	return false
}

func (m *Manager) wait(ctx context.Context, lg *Guard, snap snapshot) error {
	__antithesis_instrumentation__.Notify(122833)
	timer := timeutil.NewTimer()
	timer.Reset(base.SlowRequestThreshold)
	defer timer.Stop()

	for s := spanset.SpanScope(0); s < spanset.NumSpanScope; s++ {
		__antithesis_instrumentation__.Notify(122835)
		tr := &snap.trees[s]
		for a := spanset.SpanAccess(0); a < spanset.NumSpanAccess; a++ {
			__antithesis_instrumentation__.Notify(122836)
			latches := lg.latches(s, a)
			for i := range latches {
				__antithesis_instrumentation__.Notify(122837)
				latch := &latches[i]
				switch a {
				case spanset.SpanReadOnly:
					__antithesis_instrumentation__.Notify(122838)

					a2 := spanset.SpanReadWrite
					it := tr[a2].MakeIter()
					if err := m.iterAndWait(ctx, timer, &it, lg.pp, a, a2, latch, ignoreLater); err != nil {
						__antithesis_instrumentation__.Notify(122842)
						return err
					} else {
						__antithesis_instrumentation__.Notify(122843)
					}
				case spanset.SpanReadWrite:
					__antithesis_instrumentation__.Notify(122839)

					a2 := spanset.SpanReadWrite
					it := tr[a2].MakeIter()
					if err := m.iterAndWait(ctx, timer, &it, lg.pp, a, a2, latch, ignoreNothing); err != nil {
						__antithesis_instrumentation__.Notify(122844)
						return err
					} else {
						__antithesis_instrumentation__.Notify(122845)
					}
					__antithesis_instrumentation__.Notify(122840)

					a2 = spanset.SpanReadOnly
					it = tr[a2].MakeIter()
					if err := m.iterAndWait(ctx, timer, &it, lg.pp, a, a2, latch, ignoreEarlier); err != nil {
						__antithesis_instrumentation__.Notify(122846)
						return err
					} else {
						__antithesis_instrumentation__.Notify(122847)
					}
				default:
					__antithesis_instrumentation__.Notify(122841)
					panic("unknown access")
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(122834)
	return nil
}

func (m *Manager) iterAndWait(
	ctx context.Context,
	t *timeutil.Timer,
	it *iterator,
	pp poison.Policy,
	waitType, heldType spanset.SpanAccess,
	wait *latch,
	ignore ignoreFn,
) error {
	__antithesis_instrumentation__.Notify(122848)
	for it.FirstOverlap(wait); it.Valid(); it.NextOverlap(wait) {
		__antithesis_instrumentation__.Notify(122850)
		held := it.Cur()
		if held.done.signaled() {
			__antithesis_instrumentation__.Notify(122853)
			continue
		} else {
			__antithesis_instrumentation__.Notify(122854)
		}
		__antithesis_instrumentation__.Notify(122851)
		if ignore(wait.ts, held.ts) {
			__antithesis_instrumentation__.Notify(122855)
			continue
		} else {
			__antithesis_instrumentation__.Notify(122856)
		}
		__antithesis_instrumentation__.Notify(122852)
		if err := m.waitForSignal(ctx, t, pp, waitType, heldType, wait, held); err != nil {
			__antithesis_instrumentation__.Notify(122857)
			return err
		} else {
			__antithesis_instrumentation__.Notify(122858)
		}
	}
	__antithesis_instrumentation__.Notify(122849)
	return nil
}

func (m *Manager) waitForSignal(
	ctx context.Context,
	t *timeutil.Timer,
	pp poison.Policy,
	waitType, heldType spanset.SpanAccess,
	wait, held *latch,
) error {
	__antithesis_instrumentation__.Notify(122859)
	log.Eventf(ctx, "waiting to acquire %s latch %s, held by %s latch %s", waitType, wait, heldType, held)
	poisonCh := held.poison.signalChan()
	for {
		__antithesis_instrumentation__.Notify(122860)
		select {
		case <-held.done.signalChan():
			__antithesis_instrumentation__.Notify(122861)
			return nil
		case <-poisonCh:
			__antithesis_instrumentation__.Notify(122862)

			switch pp {
			case poison.Policy_Error:
				__antithesis_instrumentation__.Notify(122866)
				return poison.NewPoisonedError(held.span, held.ts)
			case poison.Policy_Wait:
				__antithesis_instrumentation__.Notify(122867)
				log.Eventf(ctx, "encountered poisoned latch; continuing to wait")
				wait.poison.signal()

				poisonCh = nil
			default:
				__antithesis_instrumentation__.Notify(122868)
				return errors.Errorf("unsupported poison.Policy %d", pp)
			}
		case <-t.C:
			__antithesis_instrumentation__.Notify(122863)
			t.Read = true
			defer t.Reset(base.SlowRequestThreshold)

			log.Warningf(ctx, "have been waiting %s to acquire %s latch %s, held by %s latch %s",
				base.SlowRequestThreshold, waitType, wait, heldType, held)
			if m.slowReqs != nil {
				__antithesis_instrumentation__.Notify(122869)
				m.slowReqs.Inc(1)
				defer m.slowReqs.Dec(1)
			} else {
				__antithesis_instrumentation__.Notify(122870)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(122864)
			log.VEventf(ctx, 2, "%s while acquiring %s latch %s, held by %s latch %s",
				ctx.Err(), waitType, wait, heldType, held)
			return ctx.Err()
		case <-m.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(122865)

			return &roachpb.NodeUnavailableError{}
		}
	}
}

func (m *Manager) Poison(lg *Guard) {
	__antithesis_instrumentation__.Notify(122871)
	lg.poison.signal()
}

func (m *Manager) Release(lg *Guard) {
	__antithesis_instrumentation__.Notify(122872)
	lg.done.signal()
	if lg.snap != nil {
		__antithesis_instrumentation__.Notify(122874)
		lg.snap.close()
	} else {
		__antithesis_instrumentation__.Notify(122875)
	}
	__antithesis_instrumentation__.Notify(122873)

	m.mu.Lock()
	m.removeLocked(lg)
	m.mu.Unlock()
}

func (m *Manager) removeLocked(lg *Guard) {
	__antithesis_instrumentation__.Notify(122876)
	for s := spanset.SpanScope(0); s < spanset.NumSpanScope; s++ {
		__antithesis_instrumentation__.Notify(122877)
		sm := &m.scopes[s]
		for a := spanset.SpanAccess(0); a < spanset.NumSpanAccess; a++ {
			__antithesis_instrumentation__.Notify(122878)
			latches := lg.latches(s, a)
			for i := range latches {
				__antithesis_instrumentation__.Notify(122879)
				latch := &latches[i]
				if latch.inReadSet() {
					__antithesis_instrumentation__.Notify(122880)
					sm.readSet.remove(latch)
				} else {
					__antithesis_instrumentation__.Notify(122881)
					sm.trees[a].Delete(latch)
				}
			}
		}
	}
}

type Metrics struct {
	ReadCount  int64
	WriteCount int64
}

func (m *Manager) Metrics() Metrics {
	__antithesis_instrumentation__.Notify(122882)
	m.mu.Lock()
	defer m.mu.Unlock()
	globalReadCount, globalWriteCount := m.scopes[spanset.SpanGlobal].metricsLocked()
	localReadCount, localWriteCount := m.scopes[spanset.SpanLocal].metricsLocked()
	return Metrics{
		ReadCount:  globalReadCount + localReadCount,
		WriteCount: globalWriteCount + localWriteCount,
	}
}

func (sm *scopedManager) metricsLocked() (readCount, writeCount int64) {
	__antithesis_instrumentation__.Notify(122883)
	readCount = int64(sm.trees[spanset.SpanReadOnly].Len() + sm.readSet.len)
	writeCount = int64(sm.trees[spanset.SpanReadWrite].Len())
	return readCount, writeCount
}
