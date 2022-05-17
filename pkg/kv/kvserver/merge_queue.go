package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const (
	mergeQueuePurgatoryCheckInterval = 1 * time.Minute

	mergeQueueConcurrency = 1
)

var MergeQueueInterval = settings.RegisterDurationSetting(
	settings.SystemOnly,
	"kv.range_merge.queue_interval",
	"how long the merge queue waits between processing replicas",
	5*time.Second,
	settings.NonNegativeDuration,
)

type mergeQueue struct {
	*baseQueue
	db       *kv.DB
	purgChan <-chan time.Time
}

func newMergeQueue(store *Store, db *kv.DB) *mergeQueue {
	__antithesis_instrumentation__.Notify(110122)
	mq := &mergeQueue{
		db:       db,
		purgChan: time.NewTicker(mergeQueuePurgatoryCheckInterval).C,
	}
	mq.baseQueue = newBaseQueue(
		"merge", mq, store,
		queueConfig{
			maxSize:        defaultQueueMaxSize,
			maxConcurrency: mergeQueueConcurrency,

			processTimeoutFunc:   makeRateLimitedTimeoutFunc(rebalanceSnapshotRate),
			needsLease:           true,
			needsSystemConfig:    true,
			acceptsUnsplitRanges: false,
			successes:            store.metrics.MergeQueueSuccesses,
			failures:             store.metrics.MergeQueueFailures,
			pending:              store.metrics.MergeQueuePending,
			processingNanos:      store.metrics.MergeQueueProcessingNanos,
			purgatory:            store.metrics.MergeQueuePurgatory,
		},
	)
	return mq
}

func (mq *mergeQueue) enabled() bool {
	__antithesis_instrumentation__.Notify(110123)
	if !mq.store.cfg.SpanConfigsDisabled {
		__antithesis_instrumentation__.Notify(110125)
		if mq.store.cfg.SpanConfigSubscriber.LastUpdated().IsEmpty() {
			__antithesis_instrumentation__.Notify(110126)

			return false
		} else {
			__antithesis_instrumentation__.Notify(110127)
		}
	} else {
		__antithesis_instrumentation__.Notify(110128)
	}
	__antithesis_instrumentation__.Notify(110124)

	st := mq.store.ClusterSettings()
	return kvserverbase.MergeQueueEnabled.Get(&st.SV)
}

func (mq *mergeQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, repl *Replica, confReader spanconfig.StoreReader,
) (shouldQueue bool, priority float64) {
	__antithesis_instrumentation__.Notify(110129)
	if !mq.enabled() {
		__antithesis_instrumentation__.Notify(110134)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(110135)
	}
	__antithesis_instrumentation__.Notify(110130)

	desc := repl.Desc()

	if desc.EndKey.Equal(roachpb.RKeyMax) {
		__antithesis_instrumentation__.Notify(110136)

		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(110137)
	}
	__antithesis_instrumentation__.Notify(110131)

	if confReader.NeedsSplit(ctx, desc.StartKey, desc.EndKey.Next()) {
		__antithesis_instrumentation__.Notify(110138)

		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(110139)
	}
	__antithesis_instrumentation__.Notify(110132)

	sizeRatio := float64(repl.GetMVCCStats().Total()) / float64(repl.GetMinBytes())
	if math.IsNaN(sizeRatio) || func() bool {
		__antithesis_instrumentation__.Notify(110140)
		return sizeRatio >= 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(110141)

		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(110142)
	}
	__antithesis_instrumentation__.Notify(110133)

	priority = 1 - sizeRatio
	return true, priority
}

type rangeMergePurgatoryError struct{ error }

func (rangeMergePurgatoryError) purgatoryErrorMarker() { __antithesis_instrumentation__.Notify(110143) }

var _ purgatoryError = rangeMergePurgatoryError{}

func (mq *mergeQueue) requestRangeStats(
	ctx context.Context, key roachpb.Key,
) (desc *roachpb.RangeDescriptor, stats enginepb.MVCCStats, qps float64, qpsOK bool, err error) {
	__antithesis_instrumentation__.Notify(110144)

	var ba roachpb.BatchRequest
	ba.Add(&roachpb.RangeStatsRequest{
		RequestHeader: roachpb.RequestHeader{Key: key},
	})

	br, pErr := mq.db.NonTransactionalSender().Send(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(110147)
		return nil, enginepb.MVCCStats{}, 0, false, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(110148)
	}
	__antithesis_instrumentation__.Notify(110145)
	res := br.Responses[0].GetInner().(*roachpb.RangeStatsResponse)

	desc = &res.RangeInfo.Desc
	stats = res.MVCCStats
	if res.MaxQueriesPerSecondSet {
		__antithesis_instrumentation__.Notify(110149)
		qps = res.MaxQueriesPerSecond
		qpsOK = qps >= 0
	} else {
		__antithesis_instrumentation__.Notify(110150)
		qps = res.DeprecatedLastQueriesPerSecond
		qpsOK = true
	}
	__antithesis_instrumentation__.Notify(110146)
	return desc, stats, qps, qpsOK, nil
}

func (mq *mergeQueue) process(
	ctx context.Context, lhsRepl *Replica, confReader spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(110151)
	if !mq.enabled() {
		__antithesis_instrumentation__.Notify(110165)
		log.VEventf(ctx, 2, "skipping merge: queue has been disabled")
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(110166)
	}
	__antithesis_instrumentation__.Notify(110152)

	lhsDesc := lhsRepl.Desc()
	lhsStats := lhsRepl.GetMVCCStats()
	lhsQPS, lhsQPSOK := lhsRepl.GetMaxSplitQPS()
	minBytes := lhsRepl.GetMinBytes()
	if lhsStats.Total() >= minBytes {
		__antithesis_instrumentation__.Notify(110167)
		log.VEventf(ctx, 2, "skipping merge: LHS meets minimum size threshold %d with %d bytes",
			minBytes, lhsStats.Total())
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(110168)
	}
	__antithesis_instrumentation__.Notify(110153)

	rhsDesc, rhsStats, rhsQPS, rhsQPSOK, err := mq.requestRangeStats(ctx, lhsDesc.EndKey.AsRawKey())
	if err != nil {
		__antithesis_instrumentation__.Notify(110169)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(110170)
	}
	__antithesis_instrumentation__.Notify(110154)
	if rhsStats.Total() >= minBytes {
		__antithesis_instrumentation__.Notify(110171)
		log.VEventf(ctx, 2, "skipping merge: RHS meets minimum size threshold %d with %d bytes",
			minBytes, lhsStats.Total())
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(110172)
	}
	__antithesis_instrumentation__.Notify(110155)

	now := mq.store.Clock().NowAsClockTimestamp()
	if now.ToTimestamp().Less(rhsDesc.GetStickyBit()) {
		__antithesis_instrumentation__.Notify(110173)
		log.VEventf(ctx, 2, "skipping merge: ranges were manually split and sticky bit was not expired")

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(110174)
	}
	__antithesis_instrumentation__.Notify(110156)

	mergedDesc := &roachpb.RangeDescriptor{
		StartKey: lhsDesc.StartKey,
		EndKey:   rhsDesc.EndKey,
	}
	mergedStats := lhsStats
	mergedStats.Add(rhsStats)

	var mergedQPS float64
	if lhsRepl.SplitByLoadEnabled() {
		__antithesis_instrumentation__.Notify(110175)

		if !lhsQPSOK {
			__antithesis_instrumentation__.Notify(110178)
			log.VEventf(ctx, 2, "skipping merge: LHS QPS measurement not yet reliable")
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(110179)
		}
		__antithesis_instrumentation__.Notify(110176)
		if !rhsQPSOK {
			__antithesis_instrumentation__.Notify(110180)
			log.VEventf(ctx, 2, "skipping merge: RHS QPS measurement not yet reliable")
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(110181)
		}
		__antithesis_instrumentation__.Notify(110177)
		mergedQPS = lhsQPS + rhsQPS
	} else {
		__antithesis_instrumentation__.Notify(110182)
	}
	__antithesis_instrumentation__.Notify(110157)

	conservativeLoadBasedSplitThreshold := 0.5 * lhsRepl.SplitByLoadQPSThreshold()
	shouldSplit, _ := shouldSplitRange(ctx, mergedDesc, mergedStats,
		lhsRepl.GetMaxBytes(), lhsRepl.shouldBackpressureWrites(), confReader)
	if shouldSplit || func() bool {
		__antithesis_instrumentation__.Notify(110183)
		return mergedQPS >= conservativeLoadBasedSplitThreshold == true
	}() == true {
		__antithesis_instrumentation__.Notify(110184)
		log.VEventf(ctx, 2,
			"skipping merge to avoid thrashing: merged range %s may split "+
				"(estimated size, estimated QPS: %d, %v)",
			mergedDesc, mergedStats.Total(), mergedQPS)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(110185)
	}

	{
		__antithesis_instrumentation__.Notify(110186)

		var err error
		lhsDesc, err =
			lhsRepl.maybeLeaveAtomicChangeReplicasAndRemoveLearners(ctx, lhsDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(110187)
			log.VEventf(ctx, 2, `%v`, err)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(110188)
		}
	}
	__antithesis_instrumentation__.Notify(110158)
	leftRepls, rightRepls := lhsDesc.Replicas().Descriptors(), rhsDesc.Replicas().Descriptors()

	for i := range leftRepls {
		__antithesis_instrumentation__.Notify(110189)
		if typ := leftRepls[i].GetType(); !(typ == roachpb.VOTER_FULL || func() bool {
			__antithesis_instrumentation__.Notify(110190)
			return typ == roachpb.NON_VOTER == true
		}() == true) {
			__antithesis_instrumentation__.Notify(110191)
			return false,
				errors.AssertionFailedf(
					`cannot merge because lhs is either in a joint state or has learner replicas: %v`,
					leftRepls,
				)
		} else {
			__antithesis_instrumentation__.Notify(110192)
		}
	}
	__antithesis_instrumentation__.Notify(110159)

	if !replicasCollocated(leftRepls, rightRepls) || func() bool {
		__antithesis_instrumentation__.Notify(110193)
		return rhsDesc.Replicas().InAtomicReplicationChange() == true
	}() == true {
		__antithesis_instrumentation__.Notify(110194)

		voterTargets := lhsDesc.Replicas().Voters().ReplicationTargets()
		nonVoterTargets := lhsDesc.Replicas().NonVoters().ReplicationTargets()

		lease, _ := lhsRepl.GetLease()
		for i := range voterTargets {
			__antithesis_instrumentation__.Notify(110198)
			if t := voterTargets[i]; t.NodeID == lease.Replica.NodeID && func() bool {
				__antithesis_instrumentation__.Notify(110199)
				return t.StoreID == lease.Replica.StoreID == true
			}() == true {
				__antithesis_instrumentation__.Notify(110200)
				if i > 0 {
					__antithesis_instrumentation__.Notify(110202)
					voterTargets[0], voterTargets[i] = voterTargets[i], voterTargets[0]
				} else {
					__antithesis_instrumentation__.Notify(110203)
				}
				__antithesis_instrumentation__.Notify(110201)
				break
			} else {
				__antithesis_instrumentation__.Notify(110204)
			}
		}
		__antithesis_instrumentation__.Notify(110195)

		if err := mq.store.DB().AdminRelocateRange(
			ctx,
			rhsDesc.StartKey,
			voterTargets,
			nonVoterTargets,
			false,
		); err != nil {
			__antithesis_instrumentation__.Notify(110205)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(110206)
		}
		__antithesis_instrumentation__.Notify(110196)

		rhsDesc, _, _, _, err = mq.requestRangeStats(ctx, lhsDesc.EndKey.AsRawKey())
		if err != nil {
			__antithesis_instrumentation__.Notify(110207)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(110208)
		}
		__antithesis_instrumentation__.Notify(110197)
		rightRepls = rhsDesc.Replicas().Descriptors()
	} else {
		__antithesis_instrumentation__.Notify(110209)
	}
	__antithesis_instrumentation__.Notify(110160)
	for i := range rightRepls {
		__antithesis_instrumentation__.Notify(110210)
		if typ := rightRepls[i].GetType(); !(typ == roachpb.VOTER_FULL || func() bool {
			__antithesis_instrumentation__.Notify(110211)
			return typ == roachpb.NON_VOTER == true
		}() == true) {
			__antithesis_instrumentation__.Notify(110212)
			log.Infof(ctx, "RHS Type: %s", typ)
			return false,
				errors.AssertionFailedf(
					`cannot merge because rhs is either in a joint state or has learner replicas: %v`,
					rightRepls,
				)
		} else {
			__antithesis_instrumentation__.Notify(110213)
		}
	}
	__antithesis_instrumentation__.Notify(110161)

	log.VEventf(ctx, 2, "merging to produce range: %s-%s", mergedDesc.StartKey, mergedDesc.EndKey)
	reason := fmt.Sprintf("lhs+rhs has (size=%s+%s=%s qps=%.2f+%.2f=%.2fqps) below threshold (size=%s, qps=%.2f)",
		humanizeutil.IBytes(lhsStats.Total()),
		humanizeutil.IBytes(rhsStats.Total()),
		humanizeutil.IBytes(mergedStats.Total()),
		lhsQPS,
		rhsQPS,
		mergedQPS,
		humanizeutil.IBytes(minBytes),
		conservativeLoadBasedSplitThreshold,
	)
	_, pErr := lhsRepl.AdminMerge(ctx, roachpb.AdminMergeRequest{
		RequestHeader: roachpb.RequestHeader{Key: lhsRepl.Desc().StartKey.AsRawKey()},
	}, reason)
	if err := pErr.GoError(); errors.HasType(err, (*roachpb.ConditionFailedError)(nil)) {
		__antithesis_instrumentation__.Notify(110214)

		log.Infof(ctx, "merge saw concurrent descriptor modification; maybe retrying")
		mq.MaybeAddAsync(ctx, lhsRepl, now)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(110215)
		if err != nil {
			__antithesis_instrumentation__.Notify(110216)

			log.Warningf(ctx, "%v", err)
			return false, rangeMergePurgatoryError{err}
		} else {
			__antithesis_instrumentation__.Notify(110217)
		}
	}
	__antithesis_instrumentation__.Notify(110162)
	if testingAggressiveConsistencyChecks {
		__antithesis_instrumentation__.Notify(110218)
		if _, err := mq.store.consistencyQueue.process(ctx, lhsRepl, confReader); err != nil {
			__antithesis_instrumentation__.Notify(110219)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(110220)
		}
	} else {
		__antithesis_instrumentation__.Notify(110221)
	}
	__antithesis_instrumentation__.Notify(110163)

	if mergedQPS != 0 {
		__antithesis_instrumentation__.Notify(110222)
		lhsRepl.loadBasedSplitter.RecordMax(mq.store.Clock().PhysicalTime(), mergedQPS)
	} else {
		__antithesis_instrumentation__.Notify(110223)
	}
	__antithesis_instrumentation__.Notify(110164)
	return true, nil
}

func (mq *mergeQueue) timer(time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(110224)
	return MergeQueueInterval.Get(&mq.store.ClusterSettings().SV)
}

func (mq *mergeQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(110225)
	return mq.purgChan
}
