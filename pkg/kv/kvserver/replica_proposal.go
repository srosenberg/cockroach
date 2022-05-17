package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/redact"
	"github.com/kr/pretty"
	"golang.org/x/time/rate"
)

type ProposalData struct {
	ctx context.Context

	sp *tracing.Span

	idKey kvserverbase.CmdIDKey

	proposedAtTicks int

	createdAtTicks int

	command *kvserverpb.RaftCommand

	encodedCommand []byte

	quotaAlloc *quotapool.IntAlloc

	ec endCmds

	applied bool

	doneCh chan proposalResult

	Local *result.LocalResult

	Request *roachpb.BatchRequest

	leaseStatus kvserverpb.LeaseStatus

	tok TrackedRequestToken
}

func (proposal *ProposalData) finishApplication(ctx context.Context, pr proposalResult) {
	__antithesis_instrumentation__.Notify(117814)
	proposal.ec.done(ctx, proposal.Request, pr.Reply, pr.Err)
	proposal.signalProposalResult(pr)
	if proposal.sp != nil {
		__antithesis_instrumentation__.Notify(117815)
		proposal.sp.Finish()
		proposal.sp = nil
	} else {
		__antithesis_instrumentation__.Notify(117816)
	}
}

func (proposal *ProposalData) signalProposalResult(pr proposalResult) {
	__antithesis_instrumentation__.Notify(117817)
	if proposal.doneCh != nil {
		__antithesis_instrumentation__.Notify(117818)
		proposal.doneCh <- pr
		proposal.doneCh = nil

		proposal.ctx = context.Background()
	} else {
		__antithesis_instrumentation__.Notify(117819)
	}
}

func (proposal *ProposalData) releaseQuota() {
	__antithesis_instrumentation__.Notify(117820)
	if proposal.quotaAlloc != nil {
		__antithesis_instrumentation__.Notify(117821)
		proposal.quotaAlloc.Release()
		proposal.quotaAlloc = nil
	} else {
		__antithesis_instrumentation__.Notify(117822)
	}
}

type leaseJumpOption bool

const (
	assertNoLeaseJump leaseJumpOption = false

	allowLeaseJump = true
)

func (r *Replica) leasePostApplyLocked(
	ctx context.Context,
	prevLease, newLease *roachpb.Lease,
	priorReadSum *rspb.ReadSummary,
	jumpOpt leaseJumpOption,
) {
	__antithesis_instrumentation__.Notify(117823)

	if s1, s2 := prevLease.Sequence, newLease.Sequence; s1 != 0 {
		__antithesis_instrumentation__.Notify(117832)

		switch {
		case s2 < s1:
			__antithesis_instrumentation__.Notify(117833)
			log.Fatalf(ctx, "lease sequence inversion, prevLease=%s, newLease=%s",
				redact.Safe(prevLease), redact.Safe(newLease))
		case s2 == s1:
			__antithesis_instrumentation__.Notify(117834)

			if !prevLease.Equivalent(*newLease) {
				__antithesis_instrumentation__.Notify(117838)
				log.Fatalf(ctx, "sequence identical for different leases, prevLease=%s, newLease=%s",
					redact.Safe(prevLease), redact.Safe(newLease))
			} else {
				__antithesis_instrumentation__.Notify(117839)
			}
		case s2 == s1+1:
			__antithesis_instrumentation__.Notify(117835)

		case s2 > s1+1 && func() bool {
			__antithesis_instrumentation__.Notify(117840)
			return jumpOpt == assertNoLeaseJump == true
		}() == true:
			__antithesis_instrumentation__.Notify(117836)
			log.Fatalf(ctx, "lease sequence jump, prevLease=%s, newLease=%s",
				redact.Safe(prevLease), redact.Safe(newLease))
		default:
			__antithesis_instrumentation__.Notify(117837)
		}
	} else {
		__antithesis_instrumentation__.Notify(117841)
	}
	__antithesis_instrumentation__.Notify(117824)

	iAmTheLeaseHolder := newLease.Replica.ReplicaID == r.replicaID

	leaseChangingHands := prevLease.Replica.StoreID != newLease.Replica.StoreID || func() bool {
		__antithesis_instrumentation__.Notify(117842)
		return prevLease.Sequence != newLease.Sequence == true
	}() == true

	if iAmTheLeaseHolder {
		__antithesis_instrumentation__.Notify(117843)

		if newLease.Type() == roachpb.LeaseEpoch && func() bool {
			__antithesis_instrumentation__.Notify(117844)
			return leaseChangingHands == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(117845)
			return log.V(1) == true
		}() == true {
			__antithesis_instrumentation__.Notify(117846)
			log.VEventf(ctx, 1, "new range lease %s following %s", newLease, prevLease)
		} else {
			__antithesis_instrumentation__.Notify(117847)
		}
	} else {
		__antithesis_instrumentation__.Notify(117848)
	}
	__antithesis_instrumentation__.Notify(117825)

	if leaseChangingHands && func() bool {
		__antithesis_instrumentation__.Notify(117849)
		return iAmTheLeaseHolder == true
	}() == true {
		__antithesis_instrumentation__.Notify(117850)

		if _, err := r.maybeWatchForMergeLocked(ctx); err != nil {
			__antithesis_instrumentation__.Notify(117854)

			log.Fatalf(ctx, "failed checking for in-progress merge while installing new lease %s: %s",
				newLease, err)
		} else {
			__antithesis_instrumentation__.Notify(117855)
		}
		__antithesis_instrumentation__.Notify(117851)

		r.Clock().Update(newLease.Start)

		var sum rspb.ReadSummary
		if priorReadSum != nil {
			__antithesis_instrumentation__.Notify(117856)
			sum = *priorReadSum
		} else {
			__antithesis_instrumentation__.Notify(117857)
			sum = rspb.FromTimestamp(newLease.Start.ToTimestamp())
		}
		__antithesis_instrumentation__.Notify(117852)
		applyReadSummaryToTimestampCache(r.store.tsCache, r.descRLocked(), sum)

		if r.leaseholderStats != nil {
			__antithesis_instrumentation__.Notify(117858)
			r.leaseholderStats.resetRequestCounts()
		} else {
			__antithesis_instrumentation__.Notify(117859)
		}
		__antithesis_instrumentation__.Notify(117853)
		r.loadBasedSplitter.Reset(r.Clock().PhysicalTime())
	} else {
		__antithesis_instrumentation__.Notify(117860)
	}
	__antithesis_instrumentation__.Notify(117826)

	r.concMgr.OnRangeLeaseUpdated(newLease.Sequence, iAmTheLeaseHolder)

	r.mu.proposalBuf.OnLeaseChangeLocked(iAmTheLeaseHolder, r.mu.state.RaftClosedTimestamp, r.mu.state.LeaseAppliedIndex)

	r.mu.state.Lease = newLease
	expirationBasedLease := r.requiresExpiringLeaseRLocked()

	now := r.store.Clock().NowAsClockTimestamp()
	if leaseChangingHands && func() bool {
		__antithesis_instrumentation__.Notify(117861)
		return iAmTheLeaseHolder == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(117862)
		return r.IsFirstRange() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(117863)
		return r.ownsValidLeaseRLocked(ctx, now) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117864)
		r.gossipFirstRangeLocked(ctx)
	} else {
		__antithesis_instrumentation__.Notify(117865)
	}
	__antithesis_instrumentation__.Notify(117827)

	if leaseChangingHands && func() bool {
		__antithesis_instrumentation__.Notify(117866)
		return iAmTheLeaseHolder == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(117867)
		return expirationBasedLease == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(117868)
		return r.ownsValidLeaseRLocked(ctx, now) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117869)
		r.store.renewableLeases.Store(int64(r.RangeID), unsafe.Pointer(r))
		select {
		case r.store.renewableLeasesSignal <- struct{}{}:
			__antithesis_instrumentation__.Notify(117870)
		default:
			__antithesis_instrumentation__.Notify(117871)
		}
	} else {
		__antithesis_instrumentation__.Notify(117872)
	}
	__antithesis_instrumentation__.Notify(117828)

	r.maybeTransferRaftLeadershipToLeaseholderLocked(ctx, now)

	prevOwner := prevLease.OwnedBy(r.store.StoreID())
	currentOwner := newLease.OwnedBy(r.store.StoreID())
	if leaseChangingHands && func() bool {
		__antithesis_instrumentation__.Notify(117873)
		return (prevOwner || func() bool {
			__antithesis_instrumentation__.Notify(117874)
			return currentOwner == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117875)
		if currentOwner {
			__antithesis_instrumentation__.Notify(117877)
			r.store.maybeGossipOnCapacityChange(ctx, leaseAddEvent)
		} else {
			__antithesis_instrumentation__.Notify(117878)
			if prevOwner {
				__antithesis_instrumentation__.Notify(117879)
				r.store.maybeGossipOnCapacityChange(ctx, leaseRemoveEvent)
			} else {
				__antithesis_instrumentation__.Notify(117880)
			}
		}
		__antithesis_instrumentation__.Notify(117876)
		if r.leaseholderStats != nil {
			__antithesis_instrumentation__.Notify(117881)
			r.leaseholderStats.resetRequestCounts()
		} else {
			__antithesis_instrumentation__.Notify(117882)
		}
	} else {
		__antithesis_instrumentation__.Notify(117883)
	}
	__antithesis_instrumentation__.Notify(117829)

	if iAmTheLeaseHolder {
		__antithesis_instrumentation__.Notify(117884)

		_ = r.store.stopper.RunAsyncTask(r.AnnotateCtx(context.Background()), "lease-triggers", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(117886)

			r.raftMu.Lock()
			defer r.raftMu.Unlock()
			if _, err := r.IsDestroyed(); err != nil {
				__antithesis_instrumentation__.Notify(117889)

				return
			} else {
				__antithesis_instrumentation__.Notify(117890)
			}
			__antithesis_instrumentation__.Notify(117887)
			if err := r.MaybeGossipSystemConfigRaftMuLocked(ctx); err != nil {
				__antithesis_instrumentation__.Notify(117891)
				log.Errorf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(117892)
			}
			__antithesis_instrumentation__.Notify(117888)
			if err := r.MaybeGossipNodeLivenessRaftMuLocked(ctx, keys.NodeLivenessSpan); err != nil {
				__antithesis_instrumentation__.Notify(117893)
				log.Errorf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(117894)
			}
		})
		__antithesis_instrumentation__.Notify(117885)
		if leaseChangingHands && func() bool {
			__antithesis_instrumentation__.Notify(117895)
			return log.V(1) == true
		}() == true {
			__antithesis_instrumentation__.Notify(117896)

			log.Info(ctx, "is now leaseholder")
		} else {
			__antithesis_instrumentation__.Notify(117897)
		}
	} else {
		__antithesis_instrumentation__.Notify(117898)
	}
	__antithesis_instrumentation__.Notify(117830)

	if iAmTheLeaseHolder {
		__antithesis_instrumentation__.Notify(117899)
		r.store.registerLeaseholder(ctx, r, newLease.Sequence)
	} else {
		__antithesis_instrumentation__.Notify(117900)
		r.store.unregisterLeaseholder(ctx, r)
	}
	__antithesis_instrumentation__.Notify(117831)

	if r.leaseHistory != nil {
		__antithesis_instrumentation__.Notify(117901)
		r.leaseHistory.add(*newLease)
	} else {
		__antithesis_instrumentation__.Notify(117902)
	}
}

var addSSTPreApplyWarn = struct {
	threshold time.Duration
	log.EveryN
}{30 * time.Second, log.Every(5 * time.Second)}

func addSSTablePreApply(
	ctx context.Context,
	st *cluster.Settings,
	eng storage.Engine,
	sideloaded SideloadStorage,
	term, index uint64,
	sst kvserverpb.ReplicatedEvalResult_AddSSTable,
	limiter *rate.Limiter,
) bool {
	__antithesis_instrumentation__.Notify(117903)
	checksum := util.CRC32(sst.Data)

	if checksum != sst.CRC32 {
		__antithesis_instrumentation__.Notify(117909)
		log.Fatalf(
			ctx,
			"checksum for AddSSTable at index term %d, index %d does not match; at proposal time %x (%d), now %x (%d)",
			term, index, sst.CRC32, sst.CRC32, checksum, checksum,
		)
	} else {
		__antithesis_instrumentation__.Notify(117910)
	}
	__antithesis_instrumentation__.Notify(117904)

	path, err := sideloaded.Filename(ctx, index, term)
	if err != nil {
		__antithesis_instrumentation__.Notify(117911)
		log.Fatalf(ctx, "sideloaded SSTable at term %d, index %d is missing", term, index)
	} else {
		__antithesis_instrumentation__.Notify(117912)
	}
	__antithesis_instrumentation__.Notify(117905)

	tBegin := timeutil.Now()
	var tEndDelayed time.Time
	defer func() {
		__antithesis_instrumentation__.Notify(117913)
		if dur := timeutil.Since(tBegin); dur > addSSTPreApplyWarn.threshold && func() bool {
			__antithesis_instrumentation__.Notify(117914)
			return addSSTPreApplyWarn.ShouldLog() == true
		}() == true {
			__antithesis_instrumentation__.Notify(117915)
			log.Infof(ctx,
				"ingesting SST of size %s at index %d took %.2fs (%.2fs on which in PreIngestDelay)",
				humanizeutil.IBytes(int64(len(sst.Data))), index, dur.Seconds(), tEndDelayed.Sub(tBegin).Seconds(),
			)
		} else {
			__antithesis_instrumentation__.Notify(117916)
		}
	}()
	__antithesis_instrumentation__.Notify(117906)

	eng.PreIngestDelay(ctx)
	tEndDelayed = timeutil.Now()

	copied := false
	if eng.InMem() {
		__antithesis_instrumentation__.Notify(117917)

		data := make([]byte, len(sst.Data))
		copy(data, sst.Data)
		path = fmt.Sprintf("%x", checksum)
		if err := eng.WriteFile(path, data); err != nil {
			__antithesis_instrumentation__.Notify(117918)
			log.Fatalf(ctx, "unable to write sideloaded SSTable at term %d, index %d: %s", term, index, err)
		} else {
			__antithesis_instrumentation__.Notify(117919)
		}
	} else {
		__antithesis_instrumentation__.Notify(117920)
		ingestPath := path + ".ingested"

		if linkErr := eng.Link(path, ingestPath); linkErr == nil {
			__antithesis_instrumentation__.Notify(117925)
			ingestErr := eng.IngestExternalFiles(ctx, []string{ingestPath})
			if ingestErr != nil {
				__antithesis_instrumentation__.Notify(117927)
				log.Fatalf(ctx, "while ingesting %s: %v", ingestPath, ingestErr)
			} else {
				__antithesis_instrumentation__.Notify(117928)
			}
			__antithesis_instrumentation__.Notify(117926)

			log.Eventf(ctx, "ingested SSTable at index %d, term %d: %s", index, term, ingestPath)
			return false
		} else {
			__antithesis_instrumentation__.Notify(117929)
		}
		__antithesis_instrumentation__.Notify(117921)

		path = ingestPath

		log.Eventf(ctx, "copying SSTable for ingestion at index %d, term %d: %s", index, term, path)

		if err := eng.MkdirAll(filepath.Dir(path)); err != nil {
			__antithesis_instrumentation__.Notify(117930)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(117931)
		}
		__antithesis_instrumentation__.Notify(117922)
		if _, err := eng.Stat(path); err == nil {
			__antithesis_instrumentation__.Notify(117932)

			if err := eng.Remove(path); err != nil {
				__antithesis_instrumentation__.Notify(117933)
				log.Fatalf(ctx, "while removing existing file during ingestion of %s: %+v", path, err)
			} else {
				__antithesis_instrumentation__.Notify(117934)
			}
		} else {
			__antithesis_instrumentation__.Notify(117935)
		}
		__antithesis_instrumentation__.Notify(117923)

		if err := writeFileSyncing(ctx, path, sst.Data, eng, 0600, st, limiter); err != nil {
			__antithesis_instrumentation__.Notify(117936)
			log.Fatalf(ctx, "while ingesting %s: %+v", path, err)
		} else {
			__antithesis_instrumentation__.Notify(117937)
		}
		__antithesis_instrumentation__.Notify(117924)
		copied = true
	}
	__antithesis_instrumentation__.Notify(117907)

	if err := eng.IngestExternalFiles(ctx, []string{path}); err != nil {
		__antithesis_instrumentation__.Notify(117938)
		log.Fatalf(ctx, "while ingesting %s: %+v", path, err)
	} else {
		__antithesis_instrumentation__.Notify(117939)
	}
	__antithesis_instrumentation__.Notify(117908)
	log.Eventf(ctx, "ingested SSTable at index %d, term %d: %s", index, term, path)
	return copied
}

func (r *Replica) handleReadWriteLocalEvalResult(ctx context.Context, lResult result.LocalResult) {
	__antithesis_instrumentation__.Notify(117940)

	{
		__antithesis_instrumentation__.Notify(117953)
		lResult.Reply = nil
	}
	__antithesis_instrumentation__.Notify(117941)

	if lResult.EncounteredIntents != nil {
		__antithesis_instrumentation__.Notify(117954)
		log.Fatalf(ctx, "LocalEvalResult.EncounteredIntents should be nil: %+v", lResult.EncounteredIntents)
	} else {
		__antithesis_instrumentation__.Notify(117955)
	}
	__antithesis_instrumentation__.Notify(117942)
	if lResult.EndTxns != nil {
		__antithesis_instrumentation__.Notify(117956)
		log.Fatalf(ctx, "LocalEvalResult.EndTxns should be nil: %+v", lResult.EndTxns)
	} else {
		__antithesis_instrumentation__.Notify(117957)
	}
	__antithesis_instrumentation__.Notify(117943)

	if lResult.AcquiredLocks != nil {
		__antithesis_instrumentation__.Notify(117958)
		for i := range lResult.AcquiredLocks {
			__antithesis_instrumentation__.Notify(117960)
			r.concMgr.OnLockAcquired(ctx, &lResult.AcquiredLocks[i])
		}
		__antithesis_instrumentation__.Notify(117959)
		lResult.AcquiredLocks = nil
	} else {
		__antithesis_instrumentation__.Notify(117961)
	}
	__antithesis_instrumentation__.Notify(117944)

	if lResult.ResolvedLocks != nil {
		__antithesis_instrumentation__.Notify(117962)
		for i := range lResult.ResolvedLocks {
			__antithesis_instrumentation__.Notify(117964)
			r.concMgr.OnLockUpdated(ctx, &lResult.ResolvedLocks[i])
		}
		__antithesis_instrumentation__.Notify(117963)
		lResult.ResolvedLocks = nil
	} else {
		__antithesis_instrumentation__.Notify(117965)
	}
	__antithesis_instrumentation__.Notify(117945)

	if lResult.UpdatedTxns != nil {
		__antithesis_instrumentation__.Notify(117966)
		for _, txn := range lResult.UpdatedTxns {
			__antithesis_instrumentation__.Notify(117968)
			r.concMgr.OnTransactionUpdated(ctx, txn)
		}
		__antithesis_instrumentation__.Notify(117967)
		lResult.UpdatedTxns = nil
	} else {
		__antithesis_instrumentation__.Notify(117969)
	}
	__antithesis_instrumentation__.Notify(117946)

	if lResult.GossipFirstRange {
		__antithesis_instrumentation__.Notify(117970)

		if err := r.store.Stopper().RunAsyncTask(
			ctx, "storage.Replica: gossipping first range",
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(117972)
				hasLease, pErr := r.getLeaseForGossip(ctx)

				if pErr != nil {
					__antithesis_instrumentation__.Notify(117974)
					log.Infof(ctx, "unable to gossip first range; hasLease=%t, err=%s", hasLease, pErr)
				} else {
					__antithesis_instrumentation__.Notify(117975)
					if !hasLease {
						__antithesis_instrumentation__.Notify(117976)
						return
					} else {
						__antithesis_instrumentation__.Notify(117977)
					}
				}
				__antithesis_instrumentation__.Notify(117973)
				r.gossipFirstRange(ctx)
			}); err != nil {
			__antithesis_instrumentation__.Notify(117978)
			log.Infof(ctx, "unable to gossip first range: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(117979)
		}
		__antithesis_instrumentation__.Notify(117971)
		lResult.GossipFirstRange = false
	} else {
		__antithesis_instrumentation__.Notify(117980)
	}
	__antithesis_instrumentation__.Notify(117947)

	if lResult.MaybeAddToSplitQueue {
		__antithesis_instrumentation__.Notify(117981)
		r.store.splitQueue.MaybeAddAsync(ctx, r, r.store.Clock().NowAsClockTimestamp())
		lResult.MaybeAddToSplitQueue = false
	} else {
		__antithesis_instrumentation__.Notify(117982)
	}
	__antithesis_instrumentation__.Notify(117948)

	if lResult.MaybeGossipSystemConfig {
		__antithesis_instrumentation__.Notify(117983)
		if err := r.MaybeGossipSystemConfigRaftMuLocked(ctx); err != nil {
			__antithesis_instrumentation__.Notify(117985)
			log.Errorf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(117986)
		}
		__antithesis_instrumentation__.Notify(117984)
		lResult.MaybeGossipSystemConfig = false
	} else {
		__antithesis_instrumentation__.Notify(117987)
	}
	__antithesis_instrumentation__.Notify(117949)

	if lResult.MaybeGossipSystemConfigIfHaveFailure {
		__antithesis_instrumentation__.Notify(117988)
		if err := r.MaybeGossipSystemConfigIfHaveFailureRaftMuLocked(ctx); err != nil {
			__antithesis_instrumentation__.Notify(117990)
			log.Errorf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(117991)
		}
		__antithesis_instrumentation__.Notify(117989)
		lResult.MaybeGossipSystemConfigIfHaveFailure = false
	} else {
		__antithesis_instrumentation__.Notify(117992)
	}
	__antithesis_instrumentation__.Notify(117950)

	if lResult.MaybeGossipNodeLiveness != nil {
		__antithesis_instrumentation__.Notify(117993)
		if err := r.MaybeGossipNodeLivenessRaftMuLocked(ctx, *lResult.MaybeGossipNodeLiveness); err != nil {
			__antithesis_instrumentation__.Notify(117995)
			log.Errorf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(117996)
		}
		__antithesis_instrumentation__.Notify(117994)
		lResult.MaybeGossipNodeLiveness = nil
	} else {
		__antithesis_instrumentation__.Notify(117997)
	}
	__antithesis_instrumentation__.Notify(117951)

	if lResult.Metrics != nil {
		__antithesis_instrumentation__.Notify(117998)
		r.store.metrics.handleMetricsResult(ctx, *lResult.Metrics)
		lResult.Metrics = nil
	} else {
		__antithesis_instrumentation__.Notify(117999)
	}
	__antithesis_instrumentation__.Notify(117952)

	if !lResult.IsZero() {
		__antithesis_instrumentation__.Notify(118000)
		log.Fatalf(ctx, "unhandled field in LocalEvalResult: %s", pretty.Diff(lResult, result.LocalResult{}))
	} else {
		__antithesis_instrumentation__.Notify(118001)
	}
}

type proposalResult struct {
	Reply              *roachpb.BatchResponse
	Err                *roachpb.Error
	EncounteredIntents []roachpb.Intent
	EndTxns            []result.EndTxnIntents
}

func (r *Replica) evaluateProposal(
	ctx context.Context,
	idKey kvserverbase.CmdIDKey,
	ba *roachpb.BatchRequest,
	ui uncertainty.Interval,
	g *concurrency.Guard,
) (*result.Result, bool, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(118002)
	if ba.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(118007)
		return nil, false, roachpb.NewErrorf("can't propose Raft command with zero timestamp")
	} else {
		__antithesis_instrumentation__.Notify(118008)
	}
	__antithesis_instrumentation__.Notify(118003)

	batch, ms, br, res, pErr := r.evaluateWriteBatch(ctx, idKey, ba, ui, g)

	if batch != nil {
		__antithesis_instrumentation__.Notify(118009)
		defer batch.Close()
	} else {
		__antithesis_instrumentation__.Notify(118010)
	}
	__antithesis_instrumentation__.Notify(118004)

	if pErr != nil {
		__antithesis_instrumentation__.Notify(118011)
		if _, ok := pErr.GetDetail().(*roachpb.ReplicaCorruptionError); ok {
			__antithesis_instrumentation__.Notify(118014)
			return &res, false, pErr
		} else {
			__antithesis_instrumentation__.Notify(118015)
		}
		__antithesis_instrumentation__.Notify(118012)

		txn := pErr.GetTxn()
		if txn != nil && func() bool {
			__antithesis_instrumentation__.Notify(118016)
			return ba.Txn == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(118017)
			log.Fatalf(ctx, "error had a txn but batch is non-transactional. Err txn: %s", txn)
		} else {
			__antithesis_instrumentation__.Notify(118018)
		}
		__antithesis_instrumentation__.Notify(118013)

		res.Local = result.LocalResult{
			EncounteredIntents: res.Local.DetachEncounteredIntents(),
			EndTxns:            res.Local.DetachEndTxns(true),
			Metrics:            res.Local.Metrics,
		}
		res.Replicated.Reset()
		return &res, false, pErr
	} else {
		__antithesis_instrumentation__.Notify(118019)
	}
	__antithesis_instrumentation__.Notify(118005)

	res.Local.Reply = br

	needConsensus := !batch.Empty() || func() bool {
		__antithesis_instrumentation__.Notify(118020)
		return ms != (enginepb.MVCCStats{}) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(118021)
		return !res.Replicated.IsZero() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(118022)
		return ba.RequiresConsensus() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(118023)
		return res.Local.RequiresRaft() == true
	}() == true

	if needConsensus {
		__antithesis_instrumentation__.Notify(118024)
		log.VEventf(ctx, 2, "need consensus on write batch with op count=%d", batch.Count())

		res.WriteBatch = &kvserverpb.WriteBatch{
			Data: batch.Repr(),
		}

		res.Replicated.IsLeaseRequest = ba.IsLeaseRequest()
		if ba.AppliesTimestampCache() {
			__antithesis_instrumentation__.Notify(118026)
			res.Replicated.WriteTimestamp = ba.WriteTimestamp()
		} else {
			__antithesis_instrumentation__.Notify(118027)
			if !r.ClusterSettings().Version.IsActive(ctx, clusterversion.DontProposeWriteTimestampForLeaseTransfers) {
				__antithesis_instrumentation__.Notify(118028)

				res.Replicated.WriteTimestamp = r.store.Clock().Now()
			} else {
				__antithesis_instrumentation__.Notify(118029)
			}
		}
		__antithesis_instrumentation__.Notify(118025)
		res.Replicated.Delta = ms.ToStatsDelta()

		if res.Replicated.Delta.ContainsEstimates > 0 {
			__antithesis_instrumentation__.Notify(118030)
			res.Replicated.Delta.ContainsEstimates *= 2
		} else {
			__antithesis_instrumentation__.Notify(118031)
		}
	} else {
		__antithesis_instrumentation__.Notify(118032)
	}
	__antithesis_instrumentation__.Notify(118006)

	return &res, needConsensus, nil
}

func (r *Replica) requestToProposal(
	ctx context.Context,
	idKey kvserverbase.CmdIDKey,
	ba *roachpb.BatchRequest,
	st kvserverpb.LeaseStatus,
	ui uncertainty.Interval,
	g *concurrency.Guard,
) (*ProposalData, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(118033)
	res, needConsensus, pErr := r.evaluateProposal(ctx, idKey, ba, ui, g)

	proposal := &ProposalData{
		ctx:         ctx,
		idKey:       idKey,
		doneCh:      make(chan proposalResult, 1),
		Local:       &res.Local,
		Request:     ba,
		leaseStatus: st,
	}

	if needConsensus {
		__antithesis_instrumentation__.Notify(118035)
		proposal.command = &kvserverpb.RaftCommand{
			ReplicatedEvalResult: res.Replicated,
			WriteBatch:           res.WriteBatch,
			LogicalOpLog:         res.LogicalOpLog,
			TraceData:            r.getTraceData(ctx),
		}
	} else {
		__antithesis_instrumentation__.Notify(118036)
	}
	__antithesis_instrumentation__.Notify(118034)

	return proposal, pErr
}

func (r *Replica) getTraceData(ctx context.Context) map[string]string {
	__antithesis_instrumentation__.Notify(118037)
	sp := tracing.SpanFromContext(ctx)
	if sp == nil {
		__antithesis_instrumentation__.Notify(118040)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(118041)
	}
	__antithesis_instrumentation__.Notify(118038)

	if !sp.IsVerbose() {
		__antithesis_instrumentation__.Notify(118042)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(118043)
	}
	__antithesis_instrumentation__.Notify(118039)

	traceCarrier := tracing.MapCarrier{
		Map: make(map[string]string),
	}
	r.AmbientContext.Tracer.InjectMetaInto(sp.Meta(), traceCarrier)
	return traceCarrier.Map
}
