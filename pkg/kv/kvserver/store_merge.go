package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"runtime"
	"runtime/debug"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func (s *Store) maybeAssertNoHole(ctx context.Context, from, to roachpb.RKey) func() {
	__antithesis_instrumentation__.Notify(124972)

	const disabled = true
	if disabled {
		__antithesis_instrumentation__.Notify(124977)
		return func() { __antithesis_instrumentation__.Notify(124978) }
	} else {
		__antithesis_instrumentation__.Notify(124979)
	}
	__antithesis_instrumentation__.Notify(124973)

	goroutineStopped := make(chan struct{})
	caller := string(debug.Stack())
	if from.Equal(roachpb.RKeyMax) {
		__antithesis_instrumentation__.Notify(124980)

		return func() { __antithesis_instrumentation__.Notify(124981) }
	} else {
		__antithesis_instrumentation__.Notify(124982)
	}
	__antithesis_instrumentation__.Notify(124974)
	if from.Equal(to) {
		__antithesis_instrumentation__.Notify(124983)

		return func() { __antithesis_instrumentation__.Notify(124984) }
	} else {
		__antithesis_instrumentation__.Notify(124985)
	}
	__antithesis_instrumentation__.Notify(124975)
	ctx, cancel := context.WithCancel(ctx)
	if s.stopper.RunAsyncTask(ctx, "force-assertion", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(124986)
		defer close(goroutineStopped)
		for ctx.Err() == nil {
			__antithesis_instrumentation__.Notify(124987)
			func() {
				__antithesis_instrumentation__.Notify(124988)
				s.mu.Lock()
				defer s.mu.Unlock()
				var last replicaOrPlaceholder
				err := s.mu.replicasByKey.VisitKeyRange(
					context.Background(), from, to, AscendingKeyOrder,
					func(ctx context.Context, cur replicaOrPlaceholder) error {
						__antithesis_instrumentation__.Notify(124992)

						if last.item != nil {
							__antithesis_instrumentation__.Notify(124994)
							gapStart, gapEnd := last.Desc().EndKey, cur.Desc().StartKey
							if !gapStart.Equal(gapEnd) {
								__antithesis_instrumentation__.Notify(124995)
								return errors.AssertionFailedf(
									"found hole [%s,%s) in keyspace [%s,%s), during:\n%s",
									gapStart, gapEnd, from, to, caller,
								)
							} else {
								__antithesis_instrumentation__.Notify(124996)
							}
						} else {
							__antithesis_instrumentation__.Notify(124997)
						}
						__antithesis_instrumentation__.Notify(124993)
						last = cur
						return nil
					})
				__antithesis_instrumentation__.Notify(124989)
				if err != nil {
					__antithesis_instrumentation__.Notify(124998)
					log.Fatalf(ctx, "%v", err)
				} else {
					__antithesis_instrumentation__.Notify(124999)
				}
				__antithesis_instrumentation__.Notify(124990)
				if last.item == nil {
					__antithesis_instrumentation__.Notify(125000)
					log.Fatalf(ctx, "found hole in keyspace [%s,%s), during:\n%s", from, to, caller)
				} else {
					__antithesis_instrumentation__.Notify(125001)
				}
				__antithesis_instrumentation__.Notify(124991)
				runtime.Gosched()
			}()
		}
	}) != nil {
		__antithesis_instrumentation__.Notify(125002)
		close(goroutineStopped)
	} else {
		__antithesis_instrumentation__.Notify(125003)
	}
	__antithesis_instrumentation__.Notify(124976)
	return func() {
		__antithesis_instrumentation__.Notify(125004)
		cancel()
		<-goroutineStopped
	}
}

func (s *Store) MergeRange(
	ctx context.Context,
	leftRepl *Replica,
	newLeftDesc, rightDesc roachpb.RangeDescriptor,
	freezeStart hlc.ClockTimestamp,
	rightClosedTS hlc.Timestamp,
	rightReadSum *rspb.ReadSummary,
) error {
	__antithesis_instrumentation__.Notify(125005)
	defer s.maybeAssertNoHole(ctx, leftRepl.Desc().EndKey, newLeftDesc.EndKey)()
	if oldLeftDesc := leftRepl.Desc(); !oldLeftDesc.EndKey.Less(newLeftDesc.EndKey) {
		__antithesis_instrumentation__.Notify(125015)
		return errors.Errorf("the new end key is not greater than the current one: %+v <= %+v",
			newLeftDesc.EndKey, oldLeftDesc.EndKey)
	} else {
		__antithesis_instrumentation__.Notify(125016)
	}
	__antithesis_instrumentation__.Notify(125006)

	rightRepl, err := s.GetReplica(rightDesc.RangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(125017)
		return err
	} else {
		__antithesis_instrumentation__.Notify(125018)
	}
	__antithesis_instrumentation__.Notify(125007)

	leftRepl.raftMu.AssertHeld()
	rightRepl.raftMu.AssertHeld()

	ph, err := s.removeInitializedReplicaRaftMuLocked(ctx, rightRepl, rightDesc.NextReplicaID, RemoveOptions{

		DestroyData:       false,
		InsertPlaceholder: true,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(125019)
		return errors.Wrap(err, "cannot remove range")
	} else {
		__antithesis_instrumentation__.Notify(125020)
	}
	__antithesis_instrumentation__.Notify(125008)

	if err := rightRepl.postDestroyRaftMuLocked(ctx, rightRepl.GetMVCCStats()); err != nil {
		__antithesis_instrumentation__.Notify(125021)
		return err
	} else {
		__antithesis_instrumentation__.Notify(125022)
	}
	__antithesis_instrumentation__.Notify(125009)

	if leftRepl.leaseholderStats != nil {
		__antithesis_instrumentation__.Notify(125023)
		leftRepl.leaseholderStats.resetRequestCounts()
	} else {
		__antithesis_instrumentation__.Notify(125024)
	}
	__antithesis_instrumentation__.Notify(125010)
	if leftRepl.writeStats != nil {
		__antithesis_instrumentation__.Notify(125025)

		leftRepl.writeStats.resetRequestCounts()
	} else {
		__antithesis_instrumentation__.Notify(125026)
	}
	__antithesis_instrumentation__.Notify(125011)

	rightRepl.concMgr.OnRangeMerge()

	leftLease, _ := leftRepl.GetLease()
	rightLease, _ := rightRepl.GetLease()
	if leftLease.OwnedBy(s.Ident.StoreID) {
		__antithesis_instrumentation__.Notify(125027)
		if !rightLease.OwnedBy(s.Ident.StoreID) {
			__antithesis_instrumentation__.Notify(125029)

			s.Clock().Update(freezeStart)

			var sum rspb.ReadSummary
			if rightReadSum != nil {
				__antithesis_instrumentation__.Notify(125031)
				sum = *rightReadSum
			} else {
				__antithesis_instrumentation__.Notify(125032)
				sum = rspb.FromTimestamp(freezeStart.ToTimestamp())
			}
			__antithesis_instrumentation__.Notify(125030)
			applyReadSummaryToTimestampCache(s.tsCache, &rightDesc, sum)
		} else {
			__antithesis_instrumentation__.Notify(125033)
		}
		__antithesis_instrumentation__.Notify(125028)

		sum := rspb.FromTimestamp(rightClosedTS)
		applyReadSummaryToTimestampCache(s.tsCache, &rightDesc, sum)
	} else {
		__antithesis_instrumentation__.Notify(125034)
	}
	__antithesis_instrumentation__.Notify(125012)

	s.mu.Lock()
	defer s.mu.Unlock()
	removed, err := s.removePlaceholderLocked(ctx, ph, removePlaceholderFilled)
	if err != nil {
		__antithesis_instrumentation__.Notify(125035)
		return err
	} else {
		__antithesis_instrumentation__.Notify(125036)
	}
	__antithesis_instrumentation__.Notify(125013)
	if !removed {
		__antithesis_instrumentation__.Notify(125037)
		return errors.AssertionFailedf("did not find placeholder %s", ph)
	} else {
		__antithesis_instrumentation__.Notify(125038)
	}
	__antithesis_instrumentation__.Notify(125014)

	leftRepl.mu.Lock()
	defer leftRepl.mu.Unlock()
	leftRepl.setDescLockedRaftMuLocked(ctx, &newLeftDesc)
	return nil
}
