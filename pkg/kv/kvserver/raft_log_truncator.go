package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type pendingLogTruncations struct {
	mu struct {
		syncutil.Mutex

		truncs [2]pendingTruncation
	}
}

func (p *pendingLogTruncations) computePostTruncLogSize(raftLogSize int64) int64 {
	__antithesis_instrumentation__.Notify(112840)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.iterateLocked(func(_ int, trunc pendingTruncation) {
		__antithesis_instrumentation__.Notify(112843)
		raftLogSize += trunc.logDeltaBytes
	})
	__antithesis_instrumentation__.Notify(112841)
	if raftLogSize < 0 {
		__antithesis_instrumentation__.Notify(112844)
		raftLogSize = 0
	} else {
		__antithesis_instrumentation__.Notify(112845)
	}
	__antithesis_instrumentation__.Notify(112842)
	return raftLogSize
}

func (p *pendingLogTruncations) computePostTruncFirstIndex(firstIndex uint64) uint64 {
	__antithesis_instrumentation__.Notify(112846)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.iterateLocked(func(_ int, trunc pendingTruncation) {
		__antithesis_instrumentation__.Notify(112848)
		firstIndexAfterTrunc := trunc.firstIndexAfterTrunc()
		if firstIndex < firstIndexAfterTrunc {
			__antithesis_instrumentation__.Notify(112849)
			firstIndex = firstIndexAfterTrunc
		} else {
			__antithesis_instrumentation__.Notify(112850)
		}
	})
	__antithesis_instrumentation__.Notify(112847)
	return firstIndex
}

func (p *pendingLogTruncations) isEmptyLocked() bool {
	__antithesis_instrumentation__.Notify(112851)
	return p.mu.truncs[0] == (pendingTruncation{})
}

func (p *pendingLogTruncations) frontLocked() pendingTruncation {
	__antithesis_instrumentation__.Notify(112852)
	return p.mu.truncs[0]
}

func (p *pendingLogTruncations) popLocked() {
	__antithesis_instrumentation__.Notify(112853)
	p.mu.truncs[0] = p.mu.truncs[1]
	p.mu.truncs[1] = pendingTruncation{}
}

func (p *pendingLogTruncations) iterateLocked(f func(index int, trunc pendingTruncation)) {
	__antithesis_instrumentation__.Notify(112854)
	for i, trunc := range p.mu.truncs {
		__antithesis_instrumentation__.Notify(112855)
		if !(trunc == (pendingTruncation{})) {
			__antithesis_instrumentation__.Notify(112856)
			f(i, trunc)
		} else {
			__antithesis_instrumentation__.Notify(112857)
		}
	}
}

func (p *pendingLogTruncations) reset() {
	__antithesis_instrumentation__.Notify(112858)
	p.mu.Lock()
	defer p.mu.Unlock()
	for !p.isEmptyLocked() {
		__antithesis_instrumentation__.Notify(112859)
		p.popLocked()
	}
}

func (p *pendingLogTruncations) capacity() int {
	__antithesis_instrumentation__.Notify(112860)

	return len(p.mu.truncs)
}

type pendingTruncation struct {
	roachpb.RaftTruncatedState

	expectedFirstIndex uint64

	logDeltaBytes  int64
	isDeltaTrusted bool
}

func (pt *pendingTruncation) firstIndexAfterTrunc() uint64 {
	__antithesis_instrumentation__.Notify(112861)

	return pt.Index + 1
}

type raftLogTruncator struct {
	ambientCtx context.Context
	store      storeForTruncator
	stopper    *stop.Stopper
	mu         struct {
		syncutil.Mutex

		addRanges, drainRanges map[roachpb.RangeID]struct{}

		runningTruncation  bool
		queuedDurabilityCB bool
	}
}

func makeRaftLogTruncator(
	ambientCtx log.AmbientContext, store storeForTruncator, stopper *stop.Stopper,
) *raftLogTruncator {
	__antithesis_instrumentation__.Notify(112862)
	t := &raftLogTruncator{
		ambientCtx: ambientCtx.AnnotateCtx(context.Background()),
		store:      store,
		stopper:    stopper,
	}
	t.mu.addRanges = make(map[roachpb.RangeID]struct{})
	t.mu.drainRanges = make(map[roachpb.RangeID]struct{})
	return t
}

type storeForTruncator interface {
	acquireReplicaForTruncator(rangeID roachpb.RangeID) replicaForTruncator

	releaseReplicaForTruncator(r replicaForTruncator)

	getEngine() storage.Engine
}

type replicaForTruncator interface {
	getRangeID() roachpb.RangeID

	getTruncatedState() roachpb.RaftTruncatedState

	setTruncatedStateAndSideEffects(
		_ context.Context, _ *roachpb.RaftTruncatedState, expectedFirstIndexPreTruncation uint64,
	) (expectedFirstIndexWasAccurate bool)

	setTruncationDeltaAndTrusted(deltaBytes int64, isDeltaTrusted bool)

	getPendingTruncs() *pendingLogTruncations

	sideloadedBytesIfTruncatedFromTo(
		_ context.Context, from, to uint64) (freed int64, _ error)
	getStateLoader() stateloader.StateLoader
}

func (t *raftLogTruncator) addPendingTruncation(
	ctx context.Context,
	r replicaForTruncator,
	trunc roachpb.RaftTruncatedState,
	raftExpectedFirstIndex uint64,
	raftLogDelta int64,
) {
	__antithesis_instrumentation__.Notify(112863)
	pendingTrunc := pendingTruncation{
		RaftTruncatedState: trunc,
		expectedFirstIndex: raftExpectedFirstIndex,
		logDeltaBytes:      raftLogDelta,
		isDeltaTrusted:     true,
	}
	pendingTruncs := r.getPendingTruncs()

	i := -1

	alreadyTruncIndex := r.getTruncatedState().Index

	pendingTruncs.iterateLocked(func(index int, trunc pendingTruncation) {
		__antithesis_instrumentation__.Notify(112869)
		i = index
		if trunc.Index > alreadyTruncIndex {
			__antithesis_instrumentation__.Notify(112870)
			alreadyTruncIndex = trunc.Index
		} else {
			__antithesis_instrumentation__.Notify(112871)
		}
	})
	__antithesis_instrumentation__.Notify(112864)
	if alreadyTruncIndex >= pendingTrunc.Index {
		__antithesis_instrumentation__.Notify(112872)

		return
	} else {
		__antithesis_instrumentation__.Notify(112873)
	}
	__antithesis_instrumentation__.Notify(112865)

	pos := i + 1
	mergeWithPending := false
	if pos == pendingTruncs.capacity() {
		__antithesis_instrumentation__.Notify(112874)

		pos--
		mergeWithPending = true
	} else {
		__antithesis_instrumentation__.Notify(112875)
	}
	__antithesis_instrumentation__.Notify(112866)

	sideloadedFreed, err := r.sideloadedBytesIfTruncatedFromTo(
		ctx, alreadyTruncIndex+1, pendingTrunc.firstIndexAfterTrunc())
	if err != nil {
		__antithesis_instrumentation__.Notify(112876)

		log.Errorf(ctx, "while computing size of sideloaded files to truncate: %+v", err)
		pendingTrunc.isDeltaTrusted = false
	} else {
		__antithesis_instrumentation__.Notify(112877)
	}
	__antithesis_instrumentation__.Notify(112867)
	pendingTrunc.logDeltaBytes -= sideloadedFreed
	if mergeWithPending {
		__antithesis_instrumentation__.Notify(112878)

		pendingTrunc.isDeltaTrusted = pendingTrunc.isDeltaTrusted && func() bool {
			__antithesis_instrumentation__.Notify(112880)
			return pendingTruncs.mu.truncs[pos].isDeltaTrusted == true
		}() == true
		if pendingTruncs.mu.truncs[pos].firstIndexAfterTrunc() != pendingTrunc.expectedFirstIndex {
			__antithesis_instrumentation__.Notify(112881)
			pendingTrunc.isDeltaTrusted = false
		} else {
			__antithesis_instrumentation__.Notify(112882)
		}
		__antithesis_instrumentation__.Notify(112879)
		pendingTrunc.logDeltaBytes += pendingTruncs.mu.truncs[pos].logDeltaBytes
		pendingTrunc.expectedFirstIndex = pendingTruncs.mu.truncs[pos].expectedFirstIndex
	} else {
		__antithesis_instrumentation__.Notify(112883)
	}
	__antithesis_instrumentation__.Notify(112868)
	pendingTruncs.mu.Lock()

	pendingTruncs.mu.truncs[pos] = pendingTrunc
	pendingTruncs.mu.Unlock()

	if pos == 0 {
		__antithesis_instrumentation__.Notify(112884)
		if mergeWithPending {
			__antithesis_instrumentation__.Notify(112886)
			panic("should never be merging pending truncations at pos 0")
		} else {
			__antithesis_instrumentation__.Notify(112887)
		}
		__antithesis_instrumentation__.Notify(112885)

		t.enqueueRange(r.getRangeID())
	} else {
		__antithesis_instrumentation__.Notify(112888)
	}
}

type rangesByRangeID []roachpb.RangeID

func (r rangesByRangeID) Len() int {
	__antithesis_instrumentation__.Notify(112889)
	return len(r)
}
func (r rangesByRangeID) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(112890)
	return r[i] < r[j]
}
func (r rangesByRangeID) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(112891)
	r[i], r[j] = r[j], r[i]
}

func (t *raftLogTruncator) durabilityAdvancedCallback() {
	__antithesis_instrumentation__.Notify(112892)
	runTruncation := false
	t.mu.Lock()
	if !t.mu.runningTruncation && func() bool {
		__antithesis_instrumentation__.Notify(112896)
		return len(t.mu.addRanges) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(112897)
		runTruncation = true
		t.mu.runningTruncation = true
	} else {
		__antithesis_instrumentation__.Notify(112898)
	}
	__antithesis_instrumentation__.Notify(112893)
	if !runTruncation && func() bool {
		__antithesis_instrumentation__.Notify(112899)
		return len(t.mu.addRanges) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(112900)
		t.mu.queuedDurabilityCB = true
	} else {
		__antithesis_instrumentation__.Notify(112901)
	}
	__antithesis_instrumentation__.Notify(112894)
	t.mu.Unlock()
	if !runTruncation {
		__antithesis_instrumentation__.Notify(112902)
		return
	} else {
		__antithesis_instrumentation__.Notify(112903)
	}
	__antithesis_instrumentation__.Notify(112895)
	if err := t.stopper.RunAsyncTask(t.ambientCtx, "raft-log-truncation",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(112904)
			for {
				__antithesis_instrumentation__.Notify(112905)
				t.durabilityAdvanced(ctx)
				shouldReturn := false
				t.mu.Lock()
				queued := t.mu.queuedDurabilityCB
				t.mu.queuedDurabilityCB = false
				if !queued {
					__antithesis_instrumentation__.Notify(112907)
					t.mu.runningTruncation = false
					shouldReturn = true
				} else {
					__antithesis_instrumentation__.Notify(112908)
				}
				__antithesis_instrumentation__.Notify(112906)
				t.mu.Unlock()
				if shouldReturn {
					__antithesis_instrumentation__.Notify(112909)
					return
				} else {
					__antithesis_instrumentation__.Notify(112910)
				}
			}
		}); err != nil {
		__antithesis_instrumentation__.Notify(112911)

		func() {
			__antithesis_instrumentation__.Notify(112912)
			t.mu.Lock()
			defer t.mu.Unlock()
			if !t.mu.runningTruncation {
				__antithesis_instrumentation__.Notify(112914)
				panic("expected runningTruncation")
			} else {
				__antithesis_instrumentation__.Notify(112915)
			}
			__antithesis_instrumentation__.Notify(112913)
			t.mu.runningTruncation = false
		}()
	} else {
		__antithesis_instrumentation__.Notify(112916)
	}
}

func (t *raftLogTruncator) durabilityAdvanced(ctx context.Context) {
	__antithesis_instrumentation__.Notify(112917)
	t.mu.Lock()
	t.mu.addRanges, t.mu.drainRanges = t.mu.drainRanges, t.mu.addRanges

	drainRanges := t.mu.drainRanges
	t.mu.Unlock()
	if len(drainRanges) == 0 {
		__antithesis_instrumentation__.Notify(112920)
		return
	} else {
		__antithesis_instrumentation__.Notify(112921)
	}
	__antithesis_instrumentation__.Notify(112918)
	ranges := make([]roachpb.RangeID, 0, len(drainRanges))
	for k := range drainRanges {
		__antithesis_instrumentation__.Notify(112922)
		ranges = append(ranges, k)
		delete(drainRanges, k)
	}
	__antithesis_instrumentation__.Notify(112919)

	sort.Sort(rangesByRangeID(ranges))

	reader := t.store.getEngine().NewReadOnly(storage.GuaranteedDurability)
	defer reader.Close()
	shouldQuiesce := t.stopper.ShouldQuiesce()
	quiesced := false
	for _, rangeID := range ranges {
		__antithesis_instrumentation__.Notify(112923)
		t.tryEnactTruncations(ctx, rangeID, reader)

		select {
		case <-shouldQuiesce:
			__antithesis_instrumentation__.Notify(112925)
			quiesced = true
		default:
			__antithesis_instrumentation__.Notify(112926)
		}
		__antithesis_instrumentation__.Notify(112924)
		if quiesced {
			__antithesis_instrumentation__.Notify(112927)
			break
		} else {
			__antithesis_instrumentation__.Notify(112928)
		}
	}
}

func (t *raftLogTruncator) tryEnactTruncations(
	ctx context.Context, rangeID roachpb.RangeID, reader storage.Reader,
) {
	__antithesis_instrumentation__.Notify(112929)
	r := t.store.acquireReplicaForTruncator(rangeID)
	if r == nil {
		__antithesis_instrumentation__.Notify(112940)

		return
	} else {
		__antithesis_instrumentation__.Notify(112941)
	}
	__antithesis_instrumentation__.Notify(112930)
	defer t.store.releaseReplicaForTruncator(r)
	truncState := r.getTruncatedState()
	pendingTruncs := r.getPendingTruncs()

	pendingTruncs.mu.Lock()
	for !pendingTruncs.isEmptyLocked() {
		__antithesis_instrumentation__.Notify(112942)
		pendingTrunc := pendingTruncs.frontLocked()
		if pendingTrunc.Index <= truncState.Index {
			__antithesis_instrumentation__.Notify(112943)

			pendingTruncs.popLocked()
		} else {
			__antithesis_instrumentation__.Notify(112944)
			break
		}
	}
	__antithesis_instrumentation__.Notify(112931)

	pendingTruncs.mu.Unlock()
	if pendingTruncs.isEmptyLocked() {
		__antithesis_instrumentation__.Notify(112945)

		return
	} else {
		__antithesis_instrumentation__.Notify(112946)
	}
	__antithesis_instrumentation__.Notify(112932)

	stateLoader := r.getStateLoader()
	as, err := stateLoader.LoadRangeAppliedState(ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(112947)
		log.Errorf(ctx, "error loading RangeAppliedState, dropping all pending log truncations: %s",
			err)
		pendingTruncs.reset()
		return
	} else {
		__antithesis_instrumentation__.Notify(112948)
	}
	__antithesis_instrumentation__.Notify(112933)

	enactIndex := -1
	pendingTruncs.iterateLocked(func(index int, trunc pendingTruncation) {
		__antithesis_instrumentation__.Notify(112949)
		if trunc.Index > as.RaftAppliedIndex {
			__antithesis_instrumentation__.Notify(112951)
			return
		} else {
			__antithesis_instrumentation__.Notify(112952)
		}
		__antithesis_instrumentation__.Notify(112950)
		enactIndex = index
	})
	__antithesis_instrumentation__.Notify(112934)
	if enactIndex < 0 {
		__antithesis_instrumentation__.Notify(112953)

		t.enqueueRange(rangeID)
		return
	} else {
		__antithesis_instrumentation__.Notify(112954)
	}
	__antithesis_instrumentation__.Notify(112935)

	batch := t.store.getEngine().NewUnindexedBatch(false)
	defer batch.Close()
	apply, err := handleTruncatedStateBelowRaftPreApply(ctx, &truncState,
		&pendingTruncs.mu.truncs[enactIndex].RaftTruncatedState, stateLoader, batch)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(112955)
		return !apply == true
	}() == true {
		__antithesis_instrumentation__.Notify(112956)
		if err != nil {
			__antithesis_instrumentation__.Notify(112958)
			log.Errorf(ctx, "while attempting to truncate raft log: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(112959)
			err := errors.AssertionFailedf(
				"unexpected !apply from handleTruncatedStateBelowRaftPreApply")
			if buildutil.CrdbTestBuild || func() bool {
				__antithesis_instrumentation__.Notify(112960)
				return util.RaceEnabled == true
			}() == true {
				__antithesis_instrumentation__.Notify(112961)
				log.Fatalf(ctx, "%s", err)
			} else {
				__antithesis_instrumentation__.Notify(112962)
				log.Errorf(ctx, "%s", err)
			}
		}
		__antithesis_instrumentation__.Notify(112957)
		pendingTruncs.reset()
		return
	} else {
		__antithesis_instrumentation__.Notify(112963)
	}
	__antithesis_instrumentation__.Notify(112936)

	if err := batch.Commit(false); err != nil {
		__antithesis_instrumentation__.Notify(112964)
		log.Errorf(ctx, "while committing batch to truncate raft log: %+v", err)
		pendingTruncs.reset()
		return
	} else {
		__antithesis_instrumentation__.Notify(112965)
	}
	__antithesis_instrumentation__.Notify(112937)

	pendingTruncs.iterateLocked(func(index int, trunc pendingTruncation) {
		__antithesis_instrumentation__.Notify(112966)
		if index > enactIndex {
			__antithesis_instrumentation__.Notify(112969)
			return
		} else {
			__antithesis_instrumentation__.Notify(112970)
		}
		__antithesis_instrumentation__.Notify(112967)
		isDeltaTrusted := true
		expectedFirstIndexWasAccurate := r.setTruncatedStateAndSideEffects(
			ctx, &trunc.RaftTruncatedState, trunc.expectedFirstIndex)
		if !expectedFirstIndexWasAccurate || func() bool {
			__antithesis_instrumentation__.Notify(112971)
			return !trunc.isDeltaTrusted == true
		}() == true {
			__antithesis_instrumentation__.Notify(112972)
			isDeltaTrusted = false
		} else {
			__antithesis_instrumentation__.Notify(112973)
		}
		__antithesis_instrumentation__.Notify(112968)
		r.setTruncationDeltaAndTrusted(trunc.logDeltaBytes, isDeltaTrusted)
	})
	__antithesis_instrumentation__.Notify(112938)

	pendingTruncs.mu.Lock()
	for i := 0; i <= enactIndex; i++ {
		__antithesis_instrumentation__.Notify(112974)
		pendingTruncs.popLocked()
	}
	__antithesis_instrumentation__.Notify(112939)
	pendingTruncs.mu.Unlock()
	if !pendingTruncs.isEmptyLocked() {
		__antithesis_instrumentation__.Notify(112975)
		t.enqueueRange(rangeID)
	} else {
		__antithesis_instrumentation__.Notify(112976)
	}
}

func (t *raftLogTruncator) enqueueRange(rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(112977)
	t.mu.Lock()
	t.mu.addRanges[rangeID] = struct{}{}
	t.mu.Unlock()
}
