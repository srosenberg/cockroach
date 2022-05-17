package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3"
)

const (
	replicateQueueTimerDuration = 0

	newReplicaGracePeriod = 5 * time.Minute
)

var MinLeaseTransferInterval = settings.RegisterDurationSetting(
	settings.SystemOnly,
	"kv.allocator.min_lease_transfer_interval",
	"controls how frequently leases can be transferred for rebalancing. "+
		"It does not prevent transferring leases in order to allow a "+
		"replica to be removed from a range.",
	1*time.Second,
	settings.NonNegativeDuration,
)

var (
	metaReplicateQueueAddReplicaCount = metric.Metadata{
		Name:        "queue.replicate.addreplica",
		Help:        "Number of replica additions attempted by the replicate queue",
		Measurement: "Replica Additions",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueAddVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.addvoterreplica",
		Help:        "Number of voter replica additions attempted by the replicate queue",
		Measurement: "Replica Additions",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueAddNonVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.addnonvoterreplica",
		Help:        "Number of non-voter replica additions attempted by the replicate queue",
		Measurement: "Replica Additions",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removereplica",
		Help:        "Number of replica removals attempted by the replicate queue (typically in response to a rebalancer-initiated addition)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removevoterreplica",
		Help:        "Number of voter replica removals attempted by the replicate queue (typically in response to a rebalancer-initiated addition)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveNonVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removenonvoterreplica",
		Help:        "Number of non-voter replica removals attempted by the replicate queue (typically in response to a rebalancer-initiated addition)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveDeadReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removedeadreplica",
		Help:        "Number of dead replica removals attempted by the replicate queue (typically in response to a node outage)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveDeadVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removedeadvoterreplica",
		Help:        "Number of dead voter replica removals attempted by the replicate queue (typically in response to a node outage)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveDeadNonVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removedeadnonvoterreplica",
		Help:        "Number of dead non-voter replica removals attempted by the replicate queue (typically in response to a node outage)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveDecommissioningReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removedecommissioningreplica",
		Help:        "Number of decommissioning replica removals attempted by the replicate queue (typically in response to a node outage)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveDecommissioningVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removedecommissioningvoterreplica",
		Help:        "Number of decommissioning voter replica removals attempted by the replicate queue (typically in response to a node outage)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveDecommissioningNonVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removedecommissioningnonvoterreplica",
		Help:        "Number of decommissioning non-voter replica removals attempted by the replicate queue (typically in response to a node outage)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRemoveLearnerReplicaCount = metric.Metadata{
		Name:        "queue.replicate.removelearnerreplica",
		Help:        "Number of learner replica removals attempted by the replicate queue (typically due to internal race conditions)",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRebalanceReplicaCount = metric.Metadata{
		Name:        "queue.replicate.rebalancereplica",
		Help:        "Number of replica rebalancer-initiated additions attempted by the replicate queue",
		Measurement: "Replica Additions",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRebalanceVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.rebalancevoterreplica",
		Help:        "Number of voter replica rebalancer-initiated additions attempted by the replicate queue",
		Measurement: "Replica Additions",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueRebalanceNonVoterReplicaCount = metric.Metadata{
		Name:        "queue.replicate.rebalancenonvoterreplica",
		Help:        "Number of non-voter replica rebalancer-initiated additions attempted by the replicate queue",
		Measurement: "Replica Additions",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueTransferLeaseCount = metric.Metadata{
		Name:        "queue.replicate.transferlease",
		Help:        "Number of range lease transfers attempted by the replicate queue",
		Measurement: "Lease Transfers",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueNonVoterPromotionsCount = metric.Metadata{
		Name:        "queue.replicate.nonvoterpromotions",
		Help:        "Number of non-voters promoted to voters by the replicate queue",
		Measurement: "Promotions of Non Voters to Voters",
		Unit:        metric.Unit_COUNT,
	}
	metaReplicateQueueVoterDemotionsCount = metric.Metadata{
		Name:        "queue.replicate.voterdemotions",
		Help:        "Number of voters demoted to non-voters by the replicate queue",
		Measurement: "Demotions of Voters to Non Voters",
		Unit:        metric.Unit_COUNT,
	}
)

type quorumError struct {
	msg string
}

func newQuorumError(f string, args ...interface{}) *quorumError {
	__antithesis_instrumentation__.Notify(121117)
	return &quorumError{
		msg: fmt.Sprintf(f, args...),
	}
}

func (e *quorumError) Error() string {
	__antithesis_instrumentation__.Notify(121118)
	return e.msg
}

func (*quorumError) purgatoryErrorMarker() { __antithesis_instrumentation__.Notify(121119) }

type ReplicateQueueMetrics struct {
	AddReplicaCount                           *metric.Counter
	AddVoterReplicaCount                      *metric.Counter
	AddNonVoterReplicaCount                   *metric.Counter
	RemoveReplicaCount                        *metric.Counter
	RemoveVoterReplicaCount                   *metric.Counter
	RemoveNonVoterReplicaCount                *metric.Counter
	RemoveDeadReplicaCount                    *metric.Counter
	RemoveDeadVoterReplicaCount               *metric.Counter
	RemoveDeadNonVoterReplicaCount            *metric.Counter
	RemoveDecommissioningReplicaCount         *metric.Counter
	RemoveDecommissioningVoterReplicaCount    *metric.Counter
	RemoveDecommissioningNonVoterReplicaCount *metric.Counter
	RemoveLearnerReplicaCount                 *metric.Counter
	RebalanceReplicaCount                     *metric.Counter
	RebalanceVoterReplicaCount                *metric.Counter
	RebalanceNonVoterReplicaCount             *metric.Counter
	TransferLeaseCount                        *metric.Counter
	NonVoterPromotionsCount                   *metric.Counter
	VoterDemotionsCount                       *metric.Counter
}

func makeReplicateQueueMetrics() ReplicateQueueMetrics {
	__antithesis_instrumentation__.Notify(121120)
	return ReplicateQueueMetrics{
		AddReplicaCount:                           metric.NewCounter(metaReplicateQueueAddReplicaCount),
		AddVoterReplicaCount:                      metric.NewCounter(metaReplicateQueueAddVoterReplicaCount),
		AddNonVoterReplicaCount:                   metric.NewCounter(metaReplicateQueueAddNonVoterReplicaCount),
		RemoveReplicaCount:                        metric.NewCounter(metaReplicateQueueRemoveReplicaCount),
		RemoveVoterReplicaCount:                   metric.NewCounter(metaReplicateQueueRemoveVoterReplicaCount),
		RemoveNonVoterReplicaCount:                metric.NewCounter(metaReplicateQueueRemoveNonVoterReplicaCount),
		RemoveDeadReplicaCount:                    metric.NewCounter(metaReplicateQueueRemoveDeadReplicaCount),
		RemoveDeadVoterReplicaCount:               metric.NewCounter(metaReplicateQueueRemoveDeadVoterReplicaCount),
		RemoveDeadNonVoterReplicaCount:            metric.NewCounter(metaReplicateQueueRemoveDeadNonVoterReplicaCount),
		RemoveLearnerReplicaCount:                 metric.NewCounter(metaReplicateQueueRemoveLearnerReplicaCount),
		RemoveDecommissioningReplicaCount:         metric.NewCounter(metaReplicateQueueRemoveDecommissioningReplicaCount),
		RemoveDecommissioningVoterReplicaCount:    metric.NewCounter(metaReplicateQueueRemoveDecommissioningVoterReplicaCount),
		RemoveDecommissioningNonVoterReplicaCount: metric.NewCounter(metaReplicateQueueRemoveDecommissioningNonVoterReplicaCount),
		RebalanceReplicaCount:                     metric.NewCounter(metaReplicateQueueRebalanceReplicaCount),
		RebalanceVoterReplicaCount:                metric.NewCounter(metaReplicateQueueRebalanceVoterReplicaCount),
		RebalanceNonVoterReplicaCount:             metric.NewCounter(metaReplicateQueueRebalanceNonVoterReplicaCount),
		TransferLeaseCount:                        metric.NewCounter(metaReplicateQueueTransferLeaseCount),
		NonVoterPromotionsCount:                   metric.NewCounter(metaReplicateQueueNonVoterPromotionsCount),
		VoterDemotionsCount:                       metric.NewCounter(metaReplicateQueueVoterDemotionsCount),
	}
}

func (metrics *ReplicateQueueMetrics) trackAddReplicaCount(targetType targetReplicaType) {
	__antithesis_instrumentation__.Notify(121121)
	metrics.AddReplicaCount.Inc(1)
	switch targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(121122)
		metrics.AddVoterReplicaCount.Inc(1)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(121123)
		metrics.AddNonVoterReplicaCount.Inc(1)
	default:
		__antithesis_instrumentation__.Notify(121124)
		panic(fmt.Sprintf("unsupported targetReplicaType: %v", targetType))
	}
}

func (metrics *ReplicateQueueMetrics) trackRemoveMetric(
	targetType targetReplicaType, replicaStatus replicaStatus,
) {
	__antithesis_instrumentation__.Notify(121125)
	metrics.trackRemoveReplicaCount(targetType)
	switch replicaStatus {
	case dead:
		__antithesis_instrumentation__.Notify(121126)
		metrics.trackRemoveDeadReplicaCount(targetType)
	case decommissioning:
		__antithesis_instrumentation__.Notify(121127)
		metrics.trackRemoveDecommissioningReplicaCount(targetType)
	case alive:
		__antithesis_instrumentation__.Notify(121128)
		return
	default:
		__antithesis_instrumentation__.Notify(121129)
		panic(fmt.Sprintf("unknown replicaStatus %v", replicaStatus))
	}
}

func (metrics *ReplicateQueueMetrics) trackRemoveReplicaCount(targetType targetReplicaType) {
	__antithesis_instrumentation__.Notify(121130)
	metrics.RemoveReplicaCount.Inc(1)
	switch targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(121131)
		metrics.RemoveVoterReplicaCount.Inc(1)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(121132)
		metrics.RemoveNonVoterReplicaCount.Inc(1)
	default:
		__antithesis_instrumentation__.Notify(121133)
		panic(fmt.Sprintf("unsupported targetReplicaType: %v", targetType))
	}
}

func (metrics *ReplicateQueueMetrics) trackRemoveDeadReplicaCount(targetType targetReplicaType) {
	__antithesis_instrumentation__.Notify(121134)
	metrics.RemoveDeadReplicaCount.Inc(1)
	switch targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(121135)
		metrics.RemoveDeadVoterReplicaCount.Inc(1)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(121136)
		metrics.RemoveDeadNonVoterReplicaCount.Inc(1)
	default:
		__antithesis_instrumentation__.Notify(121137)
		panic(fmt.Sprintf("unsupported targetReplicaType: %v", targetType))
	}
}

func (metrics *ReplicateQueueMetrics) trackRemoveDecommissioningReplicaCount(
	targetType targetReplicaType,
) {
	__antithesis_instrumentation__.Notify(121138)
	metrics.RemoveDecommissioningReplicaCount.Inc(1)
	switch targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(121139)
		metrics.RemoveDecommissioningVoterReplicaCount.Inc(1)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(121140)
		metrics.RemoveDecommissioningNonVoterReplicaCount.Inc(1)
	default:
		__antithesis_instrumentation__.Notify(121141)
		panic(fmt.Sprintf("unsupported targetReplicaType: %v", targetType))
	}
}

func (metrics *ReplicateQueueMetrics) trackRebalanceReplicaCount(targetType targetReplicaType) {
	__antithesis_instrumentation__.Notify(121142)
	metrics.RebalanceReplicaCount.Inc(1)
	switch targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(121143)
		metrics.RebalanceVoterReplicaCount.Inc(1)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(121144)
		metrics.RebalanceNonVoterReplicaCount.Inc(1)
	default:
		__antithesis_instrumentation__.Notify(121145)
		panic(fmt.Sprintf("unsupported targetReplicaType: %v", targetType))
	}
}

type replicateQueue struct {
	*baseQueue
	metrics           ReplicateQueueMetrics
	allocator         Allocator
	updateChan        chan time.Time
	lastLeaseTransfer atomic.Value
}

func newReplicateQueue(store *Store, allocator Allocator) *replicateQueue {
	__antithesis_instrumentation__.Notify(121146)
	rq := &replicateQueue{
		metrics:    makeReplicateQueueMetrics(),
		allocator:  allocator,
		updateChan: make(chan time.Time, 1),
	}
	store.metrics.registry.AddMetricStruct(&rq.metrics)
	rq.baseQueue = newBaseQueue(
		"replicate", rq, store,
		queueConfig{
			maxSize:              defaultQueueMaxSize,
			needsLease:           true,
			needsSystemConfig:    true,
			acceptsUnsplitRanges: store.TestingKnobs().ReplicateQueueAcceptsUnsplit,

			processTimeoutFunc: makeRateLimitedTimeoutFunc(rebalanceSnapshotRate),
			successes:          store.metrics.ReplicateQueueSuccesses,
			failures:           store.metrics.ReplicateQueueFailures,
			pending:            store.metrics.ReplicateQueuePending,
			processingNanos:    store.metrics.ReplicateQueueProcessingNanos,
			purgatory:          store.metrics.ReplicateQueuePurgatory,
		},
	)

	updateFn := func() {
		__antithesis_instrumentation__.Notify(121150)
		select {
		case rq.updateChan <- timeutil.Now():
			__antithesis_instrumentation__.Notify(121151)
		default:
			__antithesis_instrumentation__.Notify(121152)
		}
	}
	__antithesis_instrumentation__.Notify(121147)

	if g := store.cfg.Gossip; g != nil {
		__antithesis_instrumentation__.Notify(121153)
		g.RegisterCallback(gossip.MakePrefixPattern(gossip.KeyStorePrefix), func(key string, _ roachpb.Value) {
			__antithesis_instrumentation__.Notify(121154)
			if !rq.store.IsStarted() {
				__antithesis_instrumentation__.Notify(121157)
				return
			} else {
				__antithesis_instrumentation__.Notify(121158)
			}
			__antithesis_instrumentation__.Notify(121155)

			if storeID, err := gossip.StoreIDFromKey(key); err == nil && func() bool {
				__antithesis_instrumentation__.Notify(121159)
				return storeID == rq.store.StoreID() == true
			}() == true {
				__antithesis_instrumentation__.Notify(121160)
				return
			} else {
				__antithesis_instrumentation__.Notify(121161)
			}
			__antithesis_instrumentation__.Notify(121156)
			updateFn()
		})
	} else {
		__antithesis_instrumentation__.Notify(121162)
	}
	__antithesis_instrumentation__.Notify(121148)
	if nl := store.cfg.NodeLiveness; nl != nil {
		__antithesis_instrumentation__.Notify(121163)
		nl.RegisterCallback(func(_ livenesspb.Liveness) {
			__antithesis_instrumentation__.Notify(121164)
			updateFn()
		})
	} else {
		__antithesis_instrumentation__.Notify(121165)
	}
	__antithesis_instrumentation__.Notify(121149)

	return rq
}

func (rq *replicateQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, repl *Replica, _ spanconfig.StoreReader,
) (shouldQueue bool, priority float64) {
	__antithesis_instrumentation__.Notify(121166)
	desc, conf := repl.DescAndSpanConfig()
	action, priority := rq.allocator.ComputeAction(ctx, conf, desc)

	if action == AllocatorNoop {
		__antithesis_instrumentation__.Notify(121170)
		log.VEventf(ctx, 2, "no action to take")
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(121171)
		if action != AllocatorConsiderRebalance {
			__antithesis_instrumentation__.Notify(121172)
			log.VEventf(ctx, 2, "repair needed (%s), enqueuing", action)
			return true, priority
		} else {
			__antithesis_instrumentation__.Notify(121173)
		}
	}
	__antithesis_instrumentation__.Notify(121167)

	voterReplicas := desc.Replicas().VoterDescriptors()
	nonVoterReplicas := desc.Replicas().NonVoterDescriptors()
	if !rq.store.TestingKnobs().DisableReplicaRebalancing {
		__antithesis_instrumentation__.Notify(121174)
		rangeUsageInfo := rangeUsageInfoForRepl(repl)
		_, _, _, ok := rq.allocator.RebalanceVoter(
			ctx,
			conf,
			repl.RaftStatus(),
			voterReplicas,
			nonVoterReplicas,
			rangeUsageInfo,
			storeFilterThrottled,
			rq.allocator.scorerOptions(),
		)
		if ok {
			__antithesis_instrumentation__.Notify(121177)
			log.VEventf(ctx, 2, "rebalance target found for voter, enqueuing")
			return true, 0
		} else {
			__antithesis_instrumentation__.Notify(121178)
		}
		__antithesis_instrumentation__.Notify(121175)
		_, _, _, ok = rq.allocator.RebalanceNonVoter(
			ctx,
			conf,
			repl.RaftStatus(),
			voterReplicas,
			nonVoterReplicas,
			rangeUsageInfo,
			storeFilterThrottled,
			rq.allocator.scorerOptions(),
		)
		if ok {
			__antithesis_instrumentation__.Notify(121179)
			log.VEventf(ctx, 2, "rebalance target found for non-voter, enqueuing")
			return true, 0
		} else {
			__antithesis_instrumentation__.Notify(121180)
		}
		__antithesis_instrumentation__.Notify(121176)
		log.VEventf(ctx, 2, "no rebalance target found, not enqueuing")
	} else {
		__antithesis_instrumentation__.Notify(121181)
	}
	__antithesis_instrumentation__.Notify(121168)

	status := repl.LeaseStatusAt(ctx, now)
	if status.IsValid() && func() bool {
		__antithesis_instrumentation__.Notify(121182)
		return rq.canTransferLeaseFrom(ctx, repl) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(121183)
		return rq.allocator.ShouldTransferLease(ctx, conf, voterReplicas, status.Lease.Replica.StoreID, repl.leaseholderStats) == true
	}() == true {
		__antithesis_instrumentation__.Notify(121184)

		log.VEventf(ctx, 2, "lease transfer needed, enqueuing")
		return true, 0
	} else {
		__antithesis_instrumentation__.Notify(121185)
	}
	__antithesis_instrumentation__.Notify(121169)

	return false, 0
}

func (rq *replicateQueue) process(
	ctx context.Context, repl *Replica, confReader spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(121186)
	retryOpts := retry.Options{
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2,
		MaxRetries:     5,
	}

	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(121188)
		for {
			__antithesis_instrumentation__.Notify(121189)
			requeue, err := rq.processOneChange(
				ctx, repl, rq.canTransferLeaseFrom, false, false,
			)
			if isSnapshotError(err) {
				__antithesis_instrumentation__.Notify(121194)

				log.Infof(ctx, "%v", err)
				break
			} else {
				__antithesis_instrumentation__.Notify(121195)
			}
			__antithesis_instrumentation__.Notify(121190)

			if err != nil {
				__antithesis_instrumentation__.Notify(121196)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(121197)
			}
			__antithesis_instrumentation__.Notify(121191)

			if testingAggressiveConsistencyChecks {
				__antithesis_instrumentation__.Notify(121198)
				if _, err := rq.store.consistencyQueue.process(ctx, repl, confReader); err != nil {
					__antithesis_instrumentation__.Notify(121199)
					log.Warningf(ctx, "%v", err)
				} else {
					__antithesis_instrumentation__.Notify(121200)
				}
			} else {
				__antithesis_instrumentation__.Notify(121201)
			}
			__antithesis_instrumentation__.Notify(121192)

			if !requeue {
				__antithesis_instrumentation__.Notify(121202)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(121203)
			}
			__antithesis_instrumentation__.Notify(121193)

			log.VEventf(ctx, 1, "re-processing")
		}
	}
	__antithesis_instrumentation__.Notify(121187)

	return false, errors.Errorf("failed to replicate after %d retries", retryOpts.MaxRetries)
}

func (rq *replicateQueue) processOneChange(
	ctx context.Context,
	repl *Replica,
	canTransferLeaseFrom func(ctx context.Context, repl *Replica) bool,
	scatter, dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121204)

	if _, err := repl.IsDestroyed(); err != nil {
		__antithesis_instrumentation__.Notify(121208)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121209)
	}
	__antithesis_instrumentation__.Notify(121205)
	if _, pErr := repl.redirectOnOrAcquireLease(ctx); pErr != nil {
		__antithesis_instrumentation__.Notify(121210)
		return false, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(121211)
	}
	__antithesis_instrumentation__.Notify(121206)

	desc, conf := repl.DescAndSpanConfig()

	voterReplicas := desc.Replicas().VoterDescriptors()
	nonVoterReplicas := desc.Replicas().NonVoterDescriptors()
	liveVoterReplicas, deadVoterReplicas := rq.allocator.storePool.liveAndDeadReplicas(
		voterReplicas, true,
	)
	liveNonVoterReplicas, deadNonVoterReplicas := rq.allocator.storePool.liveAndDeadReplicas(
		nonVoterReplicas, true,
	)

	_ = execChangeReplicasTxn

	action, _ := rq.allocator.ComputeAction(ctx, conf, desc)
	log.VEventf(ctx, 1, "next replica action: %s", action)

	if action == AllocatorRemoveLearner {
		__antithesis_instrumentation__.Notify(121212)
		return rq.removeLearner(ctx, repl, dryRun)
	} else {
		__antithesis_instrumentation__.Notify(121213)
	}
	__antithesis_instrumentation__.Notify(121207)

	switch action {
	case AllocatorNoop, AllocatorRangeUnavailable:
		__antithesis_instrumentation__.Notify(121214)

		return false, nil

	case AllocatorAddVoter:
		__antithesis_instrumentation__.Notify(121215)
		return rq.addOrReplaceVoters(
			ctx, repl, liveVoterReplicas, liveNonVoterReplicas, -1, alive, dryRun,
		)
	case AllocatorAddNonVoter:
		__antithesis_instrumentation__.Notify(121216)
		return rq.addOrReplaceNonVoters(
			ctx, repl, liveVoterReplicas, liveNonVoterReplicas, -1, alive, dryRun,
		)

	case AllocatorRemoveVoter:
		__antithesis_instrumentation__.Notify(121217)
		return rq.removeVoter(ctx, repl, voterReplicas, nonVoterReplicas, dryRun)
	case AllocatorRemoveNonVoter:
		__antithesis_instrumentation__.Notify(121218)
		return rq.removeNonVoter(ctx, repl, voterReplicas, nonVoterReplicas, dryRun)

	case AllocatorReplaceDeadVoter:
		__antithesis_instrumentation__.Notify(121219)
		if len(deadVoterReplicas) == 0 {
			__antithesis_instrumentation__.Notify(121239)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(121240)
		}
		__antithesis_instrumentation__.Notify(121220)
		removeIdx := getRemoveIdx(voterReplicas, deadVoterReplicas[0])
		if removeIdx < 0 {
			__antithesis_instrumentation__.Notify(121241)
			return false, errors.AssertionFailedf(
				"dead voter %v unexpectedly not found in %v",
				deadVoterReplicas[0], voterReplicas)
		} else {
			__antithesis_instrumentation__.Notify(121242)
		}
		__antithesis_instrumentation__.Notify(121221)
		return rq.addOrReplaceVoters(
			ctx, repl, liveVoterReplicas, liveNonVoterReplicas, removeIdx, dead, dryRun)
	case AllocatorReplaceDeadNonVoter:
		__antithesis_instrumentation__.Notify(121222)
		if len(deadNonVoterReplicas) == 0 {
			__antithesis_instrumentation__.Notify(121243)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(121244)
		}
		__antithesis_instrumentation__.Notify(121223)
		removeIdx := getRemoveIdx(nonVoterReplicas, deadNonVoterReplicas[0])
		if removeIdx < 0 {
			__antithesis_instrumentation__.Notify(121245)
			return false, errors.AssertionFailedf(
				"dead non-voter %v unexpectedly not found in %v",
				deadNonVoterReplicas[0], nonVoterReplicas)
		} else {
			__antithesis_instrumentation__.Notify(121246)
		}
		__antithesis_instrumentation__.Notify(121224)
		return rq.addOrReplaceNonVoters(
			ctx, repl, liveVoterReplicas, liveNonVoterReplicas, removeIdx, dead, dryRun)

	case AllocatorReplaceDecommissioningVoter:
		__antithesis_instrumentation__.Notify(121225)
		decommissioningVoterReplicas := rq.allocator.storePool.decommissioningReplicas(voterReplicas)
		if len(decommissioningVoterReplicas) == 0 {
			__antithesis_instrumentation__.Notify(121247)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(121248)
		}
		__antithesis_instrumentation__.Notify(121226)
		removeIdx := getRemoveIdx(voterReplicas, decommissioningVoterReplicas[0])
		if removeIdx < 0 {
			__antithesis_instrumentation__.Notify(121249)
			return false, errors.AssertionFailedf(
				"decommissioning voter %v unexpectedly not found in %v",
				decommissioningVoterReplicas[0], voterReplicas)
		} else {
			__antithesis_instrumentation__.Notify(121250)
		}
		__antithesis_instrumentation__.Notify(121227)
		return rq.addOrReplaceVoters(
			ctx, repl, liveVoterReplicas, liveNonVoterReplicas, removeIdx, decommissioning, dryRun)
	case AllocatorReplaceDecommissioningNonVoter:
		__antithesis_instrumentation__.Notify(121228)
		decommissioningNonVoterReplicas := rq.allocator.storePool.decommissioningReplicas(nonVoterReplicas)
		if len(decommissioningNonVoterReplicas) == 0 {
			__antithesis_instrumentation__.Notify(121251)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(121252)
		}
		__antithesis_instrumentation__.Notify(121229)
		removeIdx := getRemoveIdx(nonVoterReplicas, decommissioningNonVoterReplicas[0])
		if removeIdx < 0 {
			__antithesis_instrumentation__.Notify(121253)
			return false, errors.AssertionFailedf(
				"decommissioning non-voter %v unexpectedly not found in %v",
				decommissioningNonVoterReplicas[0], nonVoterReplicas)
		} else {
			__antithesis_instrumentation__.Notify(121254)
		}
		__antithesis_instrumentation__.Notify(121230)
		return rq.addOrReplaceNonVoters(
			ctx, repl, liveVoterReplicas, liveNonVoterReplicas, removeIdx, decommissioning, dryRun)

	case AllocatorRemoveDecommissioningVoter:
		__antithesis_instrumentation__.Notify(121231)
		return rq.removeDecommissioning(ctx, repl, voterTarget, dryRun)
	case AllocatorRemoveDecommissioningNonVoter:
		__antithesis_instrumentation__.Notify(121232)
		return rq.removeDecommissioning(ctx, repl, nonVoterTarget, dryRun)

	case AllocatorRemoveDeadVoter:
		__antithesis_instrumentation__.Notify(121233)
		return rq.removeDead(ctx, repl, deadVoterReplicas, voterTarget, dryRun)
	case AllocatorRemoveDeadNonVoter:
		__antithesis_instrumentation__.Notify(121234)
		return rq.removeDead(ctx, repl, deadNonVoterReplicas, nonVoterTarget, dryRun)

	case AllocatorRemoveLearner:
		__antithesis_instrumentation__.Notify(121235)
		return rq.removeLearner(ctx, repl, dryRun)
	case AllocatorConsiderRebalance:
		__antithesis_instrumentation__.Notify(121236)
		return rq.considerRebalance(
			ctx,
			repl,
			voterReplicas,
			nonVoterReplicas,
			canTransferLeaseFrom,
			scatter,
			dryRun,
		)
	case AllocatorFinalizeAtomicReplicationChange:
		__antithesis_instrumentation__.Notify(121237)
		_, err :=
			repl.maybeLeaveAtomicChangeReplicasAndRemoveLearners(ctx, repl.Desc())

		return true, err
	default:
		__antithesis_instrumentation__.Notify(121238)
		return false, errors.Errorf("unknown allocator action %v", action)
	}
}

func getRemoveIdx(
	repls []roachpb.ReplicaDescriptor, deadOrDecommissioningRepl roachpb.ReplicaDescriptor,
) (removeIdx int) {
	__antithesis_instrumentation__.Notify(121255)
	for i, rDesc := range repls {
		__antithesis_instrumentation__.Notify(121257)
		if rDesc.StoreID == deadOrDecommissioningRepl.StoreID {
			__antithesis_instrumentation__.Notify(121258)
			removeIdx = i
			break
		} else {
			__antithesis_instrumentation__.Notify(121259)
		}
	}
	__antithesis_instrumentation__.Notify(121256)
	return removeIdx
}

func (rq *replicateQueue) addOrReplaceVoters(
	ctx context.Context,
	repl *Replica,
	liveVoterReplicas, liveNonVoterReplicas []roachpb.ReplicaDescriptor,
	removeIdx int,
	replicaStatus replicaStatus,
	dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121260)
	desc, conf := repl.DescAndSpanConfig()
	existingVoters := desc.Replicas().VoterDescriptors()
	if len(existingVoters) == 1 {
		__antithesis_instrumentation__.Notify(121269)

		removeIdx = -1
	} else {
		__antithesis_instrumentation__.Notify(121270)
	}
	__antithesis_instrumentation__.Notify(121261)

	remainingLiveVoters := liveVoterReplicas
	remainingLiveNonVoters := liveNonVoterReplicas
	if removeIdx >= 0 {
		__antithesis_instrumentation__.Notify(121271)
		replToRemove := existingVoters[removeIdx]
		for i, r := range liveVoterReplicas {
			__antithesis_instrumentation__.Notify(121273)
			if r.ReplicaID == replToRemove.ReplicaID {
				__antithesis_instrumentation__.Notify(121274)
				remainingLiveVoters = append(liveVoterReplicas[:i:i], liveVoterReplicas[i+1:]...)
				break
			} else {
				__antithesis_instrumentation__.Notify(121275)
			}
		}
		__antithesis_instrumentation__.Notify(121272)

		lhRemovalAllowed := repl.store.cfg.Settings.Version.IsActive(ctx,
			clusterversion.EnableLeaseHolderRemoval)
		if !lhRemovalAllowed {
			__antithesis_instrumentation__.Notify(121276)

			done, err := rq.maybeTransferLeaseAway(
				ctx, repl, existingVoters[removeIdx].StoreID, dryRun, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(121278)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(121279)
			}
			__antithesis_instrumentation__.Notify(121277)
			if done {
				__antithesis_instrumentation__.Notify(121280)

				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(121281)
			}
		} else {
			__antithesis_instrumentation__.Notify(121282)
		}
	} else {
		__antithesis_instrumentation__.Notify(121283)
	}
	__antithesis_instrumentation__.Notify(121262)

	lhBeingRemoved := removeIdx >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(121284)
		return existingVoters[removeIdx].StoreID == repl.store.StoreID() == true
	}() == true

	newVoter, details, err := rq.allocator.AllocateVoter(ctx, conf, remainingLiveVoters, remainingLiveNonVoters)
	if err != nil {
		__antithesis_instrumentation__.Notify(121285)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121286)
	}
	__antithesis_instrumentation__.Notify(121263)
	if removeIdx >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(121287)
		return newVoter.StoreID == existingVoters[removeIdx].StoreID == true
	}() == true {
		__antithesis_instrumentation__.Notify(121288)
		return false, errors.AssertionFailedf("allocator suggested to replace replica on s%d with itself", newVoter.StoreID)
	} else {
		__antithesis_instrumentation__.Notify(121289)
	}
	__antithesis_instrumentation__.Notify(121264)

	clusterNodes := rq.allocator.storePool.ClusterNodeCount()
	neededVoters := GetNeededVoters(conf.GetNumVoters(), clusterNodes)

	if willHave := len(existingVoters) + 1; removeIdx < 0 && func() bool {
		__antithesis_instrumentation__.Notify(121290)
		return willHave < neededVoters == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(121291)
		return willHave%2 == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(121292)

		oldPlusNewReplicas := append([]roachpb.ReplicaDescriptor(nil), existingVoters...)
		oldPlusNewReplicas = append(
			oldPlusNewReplicas,
			roachpb.ReplicaDescriptor{NodeID: newVoter.NodeID, StoreID: newVoter.StoreID},
		)
		_, _, err := rq.allocator.AllocateVoter(ctx, conf, oldPlusNewReplicas, remainingLiveNonVoters)
		if err != nil {
			__antithesis_instrumentation__.Notify(121293)

			return false, errors.Wrap(err, "avoid up-replicating to fragile quorum")
		} else {
			__antithesis_instrumentation__.Notify(121294)
		}
	} else {
		__antithesis_instrumentation__.Notify(121295)
	}
	__antithesis_instrumentation__.Notify(121265)

	var ops []roachpb.ReplicationChange
	replDesc, found := desc.GetReplicaDescriptor(newVoter.StoreID)
	if found {
		__antithesis_instrumentation__.Notify(121296)
		if replDesc.GetType() != roachpb.NON_VOTER {
			__antithesis_instrumentation__.Notify(121299)
			return false, errors.AssertionFailedf("allocation target %s for a voter"+
				" already has an unexpected replica: %s", newVoter, replDesc)
		} else {
			__antithesis_instrumentation__.Notify(121300)
		}
		__antithesis_instrumentation__.Notify(121297)

		if !dryRun {
			__antithesis_instrumentation__.Notify(121301)
			rq.metrics.NonVoterPromotionsCount.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(121302)
		}
		__antithesis_instrumentation__.Notify(121298)
		ops = roachpb.ReplicationChangesForPromotion(newVoter)
	} else {
		__antithesis_instrumentation__.Notify(121303)
		if !dryRun {
			__antithesis_instrumentation__.Notify(121305)
			rq.metrics.trackAddReplicaCount(voterTarget)
		} else {
			__antithesis_instrumentation__.Notify(121306)
		}
		__antithesis_instrumentation__.Notify(121304)
		ops = roachpb.MakeReplicationChanges(roachpb.ADD_VOTER, newVoter)
	}
	__antithesis_instrumentation__.Notify(121266)
	if removeIdx < 0 {
		__antithesis_instrumentation__.Notify(121307)
		log.VEventf(ctx, 1, "adding voter %+v: %s",
			newVoter, rangeRaftProgress(repl.RaftStatus(), existingVoters))
	} else {
		__antithesis_instrumentation__.Notify(121308)
		if !dryRun {
			__antithesis_instrumentation__.Notify(121310)
			rq.metrics.trackRemoveMetric(voterTarget, replicaStatus)
		} else {
			__antithesis_instrumentation__.Notify(121311)
		}
		__antithesis_instrumentation__.Notify(121309)
		removeVoter := existingVoters[removeIdx]
		log.VEventf(ctx, 1, "replacing voter %s with %+v: %s",
			removeVoter, newVoter, rangeRaftProgress(repl.RaftStatus(), existingVoters))

		ops = append(ops,
			roachpb.MakeReplicationChanges(roachpb.REMOVE_VOTER, roachpb.ReplicationTarget{
				StoreID: removeVoter.StoreID,
				NodeID:  removeVoter.NodeID,
			})...)
	}
	__antithesis_instrumentation__.Notify(121267)

	if err := rq.changeReplicas(
		ctx,
		repl,
		ops,
		desc,
		kvserverpb.SnapshotRequest_RECOVERY,
		kvserverpb.ReasonRangeUnderReplicated,
		details,
		dryRun,
	); err != nil {
		__antithesis_instrumentation__.Notify(121312)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121313)
	}
	__antithesis_instrumentation__.Notify(121268)

	return !lhBeingRemoved, nil
}

func (rq *replicateQueue) addOrReplaceNonVoters(
	ctx context.Context,
	repl *Replica,
	liveVoterReplicas, liveNonVoterReplicas []roachpb.ReplicaDescriptor,
	removeIdx int,
	replicaStatus replicaStatus,
	dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121314)
	desc, conf := repl.DescAndSpanConfig()
	existingNonVoters := desc.Replicas().NonVoterDescriptors()

	newNonVoter, details, err := rq.allocator.AllocateNonVoter(ctx, conf, liveVoterReplicas, liveNonVoterReplicas)
	if err != nil {
		__antithesis_instrumentation__.Notify(121319)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121320)
	}
	__antithesis_instrumentation__.Notify(121315)
	if !dryRun {
		__antithesis_instrumentation__.Notify(121321)
		rq.metrics.trackAddReplicaCount(nonVoterTarget)
	} else {
		__antithesis_instrumentation__.Notify(121322)
	}
	__antithesis_instrumentation__.Notify(121316)

	ops := roachpb.MakeReplicationChanges(roachpb.ADD_NON_VOTER, newNonVoter)
	if removeIdx < 0 {
		__antithesis_instrumentation__.Notify(121323)
		log.VEventf(ctx, 1, "adding non-voter %+v: %s",
			newNonVoter, rangeRaftProgress(repl.RaftStatus(), existingNonVoters))
	} else {
		__antithesis_instrumentation__.Notify(121324)
		if !dryRun {
			__antithesis_instrumentation__.Notify(121326)
			rq.metrics.trackRemoveMetric(nonVoterTarget, replicaStatus)
		} else {
			__antithesis_instrumentation__.Notify(121327)
		}
		__antithesis_instrumentation__.Notify(121325)
		removeNonVoter := existingNonVoters[removeIdx]
		log.VEventf(ctx, 1, "replacing non-voter %s with %+v: %s",
			removeNonVoter, newNonVoter, rangeRaftProgress(repl.RaftStatus(), existingNonVoters))
		ops = append(ops,
			roachpb.MakeReplicationChanges(roachpb.REMOVE_NON_VOTER, roachpb.ReplicationTarget{
				StoreID: removeNonVoter.StoreID,
				NodeID:  removeNonVoter.NodeID,
			})...)
	}
	__antithesis_instrumentation__.Notify(121317)

	if err := rq.changeReplicas(
		ctx,
		repl,
		ops,
		desc,
		kvserverpb.SnapshotRequest_RECOVERY,
		kvserverpb.ReasonRangeUnderReplicated,
		details,
		dryRun,
	); err != nil {
		__antithesis_instrumentation__.Notify(121328)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121329)
	}
	__antithesis_instrumentation__.Notify(121318)

	return true, nil
}

func (rq *replicateQueue) findRemoveVoter(
	ctx context.Context,
	repl interface {
		DescAndSpanConfig() (*roachpb.RangeDescriptor, roachpb.SpanConfig)
		LastReplicaAdded() (roachpb.ReplicaID, time.Time)
		RaftStatus() *raft.Status
	},
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
) (roachpb.ReplicationTarget, string, error) {
	__antithesis_instrumentation__.Notify(121330)
	_, zone := repl.DescAndSpanConfig()

	retryOpts := retry.Options{
		InitialBackoff: time.Millisecond,
		MaxBackoff:     200 * time.Millisecond,
		Multiplier:     2,
	}

	var candidates []roachpb.ReplicaDescriptor
	deadline := timeutil.Now().Add(2 * base.NetworkTimeout)
	for r := retry.StartWithCtx(ctx, retryOpts); r.Next() && func() bool {
		__antithesis_instrumentation__.Notify(121333)
		return timeutil.Now().Before(deadline) == true
	}() == true; {
		__antithesis_instrumentation__.Notify(121334)
		lastReplAdded, lastAddedTime := repl.LastReplicaAdded()
		if timeutil.Since(lastAddedTime) > newReplicaGracePeriod {
			__antithesis_instrumentation__.Notify(121338)
			lastReplAdded = 0
		} else {
			__antithesis_instrumentation__.Notify(121339)
		}
		__antithesis_instrumentation__.Notify(121335)
		raftStatus := repl.RaftStatus()
		if raftStatus == nil || func() bool {
			__antithesis_instrumentation__.Notify(121340)
			return raftStatus.RaftState != raft.StateLeader == true
		}() == true {
			__antithesis_instrumentation__.Notify(121341)

			return roachpb.ReplicationTarget{}, "", &benignError{errors.Errorf("not raft leader while range needs removal")}
		} else {
			__antithesis_instrumentation__.Notify(121342)
		}
		__antithesis_instrumentation__.Notify(121336)
		candidates = filterUnremovableReplicas(ctx, raftStatus, existingVoters, lastReplAdded)
		log.VEventf(ctx, 3, "filtered unremovable replicas from %v to get %v as candidates for removal: %s",
			existingVoters, candidates, rangeRaftProgress(raftStatus, existingVoters))
		if len(candidates) > 0 {
			__antithesis_instrumentation__.Notify(121343)
			break
		} else {
			__antithesis_instrumentation__.Notify(121344)
		}
		__antithesis_instrumentation__.Notify(121337)
		if len(raftStatus.Progress) <= 2 {
			__antithesis_instrumentation__.Notify(121345)

			break
		} else {
			__antithesis_instrumentation__.Notify(121346)
		}

	}
	__antithesis_instrumentation__.Notify(121331)
	if len(candidates) == 0 {
		__antithesis_instrumentation__.Notify(121347)

		return roachpb.ReplicationTarget{}, "", &benignError{
			errors.Errorf(
				"no removable replicas from range that needs a removal: %s",
				rangeRaftProgress(repl.RaftStatus(), existingVoters),
			),
		}
	} else {
		__antithesis_instrumentation__.Notify(121348)
	}
	__antithesis_instrumentation__.Notify(121332)

	return rq.allocator.RemoveVoter(
		ctx,
		zone,
		candidates,
		existingVoters,
		existingNonVoters,
		rq.allocator.scorerOptions(),
	)
}

func (rq *replicateQueue) maybeTransferLeaseAway(
	ctx context.Context,
	repl *Replica,
	removeStoreID roachpb.StoreID,
	dryRun bool,
	canTransferLeaseFrom func(ctx context.Context, repl *Replica) bool,
) (done bool, _ error) {
	__antithesis_instrumentation__.Notify(121349)
	if removeStoreID != repl.store.StoreID() {
		__antithesis_instrumentation__.Notify(121352)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(121353)
	}
	__antithesis_instrumentation__.Notify(121350)
	if canTransferLeaseFrom != nil && func() bool {
		__antithesis_instrumentation__.Notify(121354)
		return !canTransferLeaseFrom(ctx, repl) == true
	}() == true {
		__antithesis_instrumentation__.Notify(121355)
		return false, errors.Errorf("cannot transfer lease")
	} else {
		__antithesis_instrumentation__.Notify(121356)
	}
	__antithesis_instrumentation__.Notify(121351)
	desc, conf := repl.DescAndSpanConfig()

	transferred, err := rq.shedLease(
		ctx,
		repl,
		desc,
		conf,
		transferLeaseOptions{
			dryRun: dryRun,
		},
	)
	return transferred == transferOK, err
}

func (rq *replicateQueue) removeVoter(
	ctx context.Context,
	repl *Replica,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
	dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121357)
	removeVoter, details, err := rq.findRemoveVoter(ctx, repl, existingVoters, existingNonVoters)
	if err != nil {
		__antithesis_instrumentation__.Notify(121363)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121364)
	}
	__antithesis_instrumentation__.Notify(121358)
	done, err := rq.maybeTransferLeaseAway(
		ctx, repl, removeVoter.StoreID, dryRun, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(121365)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121366)
	}
	__antithesis_instrumentation__.Notify(121359)
	if done {
		__antithesis_instrumentation__.Notify(121367)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(121368)
	}
	__antithesis_instrumentation__.Notify(121360)

	if !dryRun {
		__antithesis_instrumentation__.Notify(121369)
		rq.metrics.trackRemoveMetric(voterTarget, alive)
	} else {
		__antithesis_instrumentation__.Notify(121370)
	}
	__antithesis_instrumentation__.Notify(121361)

	log.VEventf(ctx, 1, "removing voting replica %+v due to over-replication: %s",
		removeVoter, rangeRaftProgress(repl.RaftStatus(), existingVoters))
	desc := repl.Desc()

	if err := rq.changeReplicas(
		ctx,
		repl,
		roachpb.MakeReplicationChanges(roachpb.REMOVE_VOTER, removeVoter),
		desc,
		kvserverpb.SnapshotRequest_UNKNOWN,
		kvserverpb.ReasonRangeOverReplicated,
		details,
		dryRun,
	); err != nil {
		__antithesis_instrumentation__.Notify(121371)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121372)
	}
	__antithesis_instrumentation__.Notify(121362)
	return true, nil
}

func (rq *replicateQueue) removeNonVoter(
	ctx context.Context,
	repl *Replica,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
	dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121373)

	desc, conf := repl.DescAndSpanConfig()
	removeNonVoter, details, err := rq.allocator.RemoveNonVoter(
		ctx,
		conf,
		existingNonVoters,
		existingVoters,
		existingNonVoters,
		rq.allocator.scorerOptions(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(121377)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121378)
	}
	__antithesis_instrumentation__.Notify(121374)
	if !dryRun {
		__antithesis_instrumentation__.Notify(121379)
		rq.metrics.trackRemoveMetric(nonVoterTarget, alive)
	} else {
		__antithesis_instrumentation__.Notify(121380)
	}
	__antithesis_instrumentation__.Notify(121375)
	log.VEventf(ctx, 1, "removing non-voting replica %+v due to over-replication: %s",
		removeNonVoter, rangeRaftProgress(repl.RaftStatus(), existingVoters))
	target := roachpb.ReplicationTarget{
		NodeID:  removeNonVoter.NodeID,
		StoreID: removeNonVoter.StoreID,
	}

	if err := rq.changeReplicas(
		ctx,
		repl,
		roachpb.MakeReplicationChanges(roachpb.REMOVE_NON_VOTER, target),
		desc,
		kvserverpb.SnapshotRequest_UNKNOWN,
		kvserverpb.ReasonRangeOverReplicated,
		details,
		dryRun,
	); err != nil {
		__antithesis_instrumentation__.Notify(121381)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121382)
	}
	__antithesis_instrumentation__.Notify(121376)
	return true, nil
}

func (rq *replicateQueue) removeDecommissioning(
	ctx context.Context, repl *Replica, targetType targetReplicaType, dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121383)
	desc := repl.Desc()
	var decommissioningReplicas []roachpb.ReplicaDescriptor
	switch targetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(121390)
		decommissioningReplicas = rq.allocator.storePool.decommissioningReplicas(
			desc.Replicas().VoterDescriptors(),
		)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(121391)
		decommissioningReplicas = rq.allocator.storePool.decommissioningReplicas(
			desc.Replicas().NonVoterDescriptors(),
		)
	default:
		__antithesis_instrumentation__.Notify(121392)
		panic(fmt.Sprintf("unknown targetReplicaType: %s", targetType))
	}
	__antithesis_instrumentation__.Notify(121384)

	if len(decommissioningReplicas) == 0 {
		__antithesis_instrumentation__.Notify(121393)
		log.VEventf(ctx, 1, "range of %[1]ss %[2]s was identified as having decommissioning %[1]ss, "+
			"but no decommissioning %[1]ss were found", targetType, repl)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(121394)
	}
	__antithesis_instrumentation__.Notify(121385)
	decommissioningReplica := decommissioningReplicas[0]

	done, err := rq.maybeTransferLeaseAway(
		ctx, repl, decommissioningReplica.StoreID, dryRun, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(121395)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121396)
	}
	__antithesis_instrumentation__.Notify(121386)
	if done {
		__antithesis_instrumentation__.Notify(121397)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(121398)
	}
	__antithesis_instrumentation__.Notify(121387)

	if !dryRun {
		__antithesis_instrumentation__.Notify(121399)
		rq.metrics.trackRemoveMetric(targetType, decommissioning)
	} else {
		__antithesis_instrumentation__.Notify(121400)
	}
	__antithesis_instrumentation__.Notify(121388)
	log.VEventf(ctx, 1, "removing decommissioning %s %+v from store", targetType, decommissioningReplica)
	target := roachpb.ReplicationTarget{
		NodeID:  decommissioningReplica.NodeID,
		StoreID: decommissioningReplica.StoreID,
	}
	if err := rq.changeReplicas(
		ctx,
		repl,
		roachpb.MakeReplicationChanges(targetType.RemoveChangeType(), target),
		desc,
		kvserverpb.SnapshotRequest_UNKNOWN,
		kvserverpb.ReasonStoreDecommissioning, "", dryRun,
	); err != nil {
		__antithesis_instrumentation__.Notify(121401)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121402)
	}
	__antithesis_instrumentation__.Notify(121389)

	return true, nil
}

func (rq *replicateQueue) removeDead(
	ctx context.Context,
	repl *Replica,
	deadReplicas []roachpb.ReplicaDescriptor,
	targetType targetReplicaType,
	dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121403)
	desc := repl.Desc()
	if len(deadReplicas) == 0 {
		__antithesis_instrumentation__.Notify(121407)
		log.VEventf(
			ctx,
			1,
			"range of %[1]s %[2]s was identified as having dead %[1]ss, but no dead %[1]ss were found",
			targetType,
			repl,
		)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(121408)
	}
	__antithesis_instrumentation__.Notify(121404)
	deadReplica := deadReplicas[0]
	if !dryRun {
		__antithesis_instrumentation__.Notify(121409)
		rq.metrics.trackRemoveMetric(targetType, dead)
	} else {
		__antithesis_instrumentation__.Notify(121410)
	}
	__antithesis_instrumentation__.Notify(121405)
	log.VEventf(ctx, 1, "removing dead %s %+v from store", targetType, deadReplica)
	target := roachpb.ReplicationTarget{
		NodeID:  deadReplica.NodeID,
		StoreID: deadReplica.StoreID,
	}

	if err := rq.changeReplicas(
		ctx,
		repl,
		roachpb.MakeReplicationChanges(targetType.RemoveChangeType(), target),
		desc,
		kvserverpb.SnapshotRequest_UNKNOWN,
		kvserverpb.ReasonStoreDead,
		"",
		dryRun,
	); err != nil {
		__antithesis_instrumentation__.Notify(121411)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121412)
	}
	__antithesis_instrumentation__.Notify(121406)
	return true, nil
}

func (rq *replicateQueue) removeLearner(
	ctx context.Context, repl *Replica, dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121413)
	desc := repl.Desc()
	learnerReplicas := desc.Replicas().LearnerDescriptors()
	if len(learnerReplicas) == 0 {
		__antithesis_instrumentation__.Notify(121417)
		log.VEventf(ctx, 1, "range of replica %s was identified as having learner replicas, "+
			"but no learner replicas were found", repl)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(121418)
	}
	__antithesis_instrumentation__.Notify(121414)
	learnerReplica := learnerReplicas[0]
	if !dryRun {
		__antithesis_instrumentation__.Notify(121419)
		rq.metrics.RemoveLearnerReplicaCount.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(121420)
	}
	__antithesis_instrumentation__.Notify(121415)
	log.VEventf(ctx, 1, "removing learner replica %+v from store", learnerReplica)
	target := roachpb.ReplicationTarget{
		NodeID:  learnerReplica.NodeID,
		StoreID: learnerReplica.StoreID,
	}

	if err := rq.changeReplicas(
		ctx,
		repl,
		roachpb.MakeReplicationChanges(roachpb.REMOVE_VOTER, target),
		desc,
		kvserverpb.SnapshotRequest_UNKNOWN,
		kvserverpb.ReasonAbandonedLearner,
		"",
		dryRun,
	); err != nil {
		__antithesis_instrumentation__.Notify(121421)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(121422)
	}
	__antithesis_instrumentation__.Notify(121416)
	return true, nil
}

func (rq *replicateQueue) considerRebalance(
	ctx context.Context,
	repl *Replica,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
	canTransferLeaseFrom func(ctx context.Context, repl *Replica) bool,
	scatter, dryRun bool,
) (requeue bool, _ error) {
	__antithesis_instrumentation__.Notify(121423)
	desc, conf := repl.DescAndSpanConfig()
	rebalanceTargetType := voterTarget

	scorerOpts := scorerOptions(rq.allocator.scorerOptions())
	if scatter {
		__antithesis_instrumentation__.Notify(121427)
		scorerOpts = rq.allocator.scorerOptionsForScatter()
	} else {
		__antithesis_instrumentation__.Notify(121428)
	}
	__antithesis_instrumentation__.Notify(121424)
	if !rq.store.TestingKnobs().DisableReplicaRebalancing {
		__antithesis_instrumentation__.Notify(121429)
		rangeUsageInfo := rangeUsageInfoForRepl(repl)
		addTarget, removeTarget, details, ok := rq.allocator.RebalanceVoter(
			ctx,
			conf,
			repl.RaftStatus(),
			existingVoters,
			existingNonVoters,
			rangeUsageInfo,
			storeFilterThrottled,
			scorerOpts,
		)
		if !ok {
			__antithesis_instrumentation__.Notify(121432)

			log.VEventf(ctx, 1, "no suitable rebalance target for voters")
			addTarget, removeTarget, details, ok = rq.allocator.RebalanceNonVoter(
				ctx,
				conf,
				repl.RaftStatus(),
				existingVoters,
				existingNonVoters,
				rangeUsageInfo,
				storeFilterThrottled,
				scorerOpts,
			)
			rebalanceTargetType = nonVoterTarget
		} else {
			__antithesis_instrumentation__.Notify(121433)
		}
		__antithesis_instrumentation__.Notify(121430)

		lhRemovalAllowed := addTarget != (roachpb.ReplicationTarget{}) && func() bool {
			__antithesis_instrumentation__.Notify(121434)
			return repl.store.cfg.Settings.Version.IsActive(ctx, clusterversion.EnableLeaseHolderRemoval) == true
		}() == true
		lhBeingRemoved := removeTarget.StoreID == repl.store.StoreID()

		if !ok {
			__antithesis_instrumentation__.Notify(121435)
			log.VEventf(ctx, 1, "no suitable rebalance target for non-voters")
		} else {
			__antithesis_instrumentation__.Notify(121436)
			if !lhRemovalAllowed {
				__antithesis_instrumentation__.Notify(121437)
				if done, err := rq.maybeTransferLeaseAway(
					ctx, repl, removeTarget.StoreID, dryRun, canTransferLeaseFrom,
				); err != nil {
					__antithesis_instrumentation__.Notify(121438)
					log.VEventf(ctx, 1, "want to remove self, but failed to transfer lease away: %s", err)
					ok = false
				} else {
					__antithesis_instrumentation__.Notify(121439)
					if done {
						__antithesis_instrumentation__.Notify(121440)

						return false, nil
					} else {
						__antithesis_instrumentation__.Notify(121441)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(121442)
			}
		}
		__antithesis_instrumentation__.Notify(121431)
		if ok {
			__antithesis_instrumentation__.Notify(121443)

			chgs, performingSwap, err := replicationChangesForRebalance(ctx, desc, len(existingVoters), addTarget,
				removeTarget, rebalanceTargetType)
			if err != nil {
				__antithesis_instrumentation__.Notify(121447)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(121448)
			}
			__antithesis_instrumentation__.Notify(121444)
			if !dryRun {
				__antithesis_instrumentation__.Notify(121449)
				rq.metrics.trackRebalanceReplicaCount(rebalanceTargetType)
				if performingSwap {
					__antithesis_instrumentation__.Notify(121450)
					rq.metrics.VoterDemotionsCount.Inc(1)
					rq.metrics.NonVoterPromotionsCount.Inc(1)
				} else {
					__antithesis_instrumentation__.Notify(121451)
				}
			} else {
				__antithesis_instrumentation__.Notify(121452)
			}
			__antithesis_instrumentation__.Notify(121445)
			log.VEventf(ctx,
				1,
				"rebalancing %s %+v to %+v: %s",
				rebalanceTargetType,
				removeTarget,
				addTarget,
				rangeRaftProgress(repl.RaftStatus(), existingVoters))

			if err := rq.changeReplicas(
				ctx,
				repl,
				chgs,
				desc,
				kvserverpb.SnapshotRequest_REBALANCE,
				kvserverpb.ReasonRebalance,
				details,
				dryRun,
			); err != nil {
				__antithesis_instrumentation__.Notify(121453)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(121454)
			}
			__antithesis_instrumentation__.Notify(121446)

			return !lhBeingRemoved, nil
		} else {
			__antithesis_instrumentation__.Notify(121455)
		}
	} else {
		__antithesis_instrumentation__.Notify(121456)
	}
	__antithesis_instrumentation__.Notify(121425)

	if !canTransferLeaseFrom(ctx, repl) {
		__antithesis_instrumentation__.Notify(121457)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(121458)
	}
	__antithesis_instrumentation__.Notify(121426)

	_, err := rq.shedLease(
		ctx,
		repl,
		desc,
		conf,
		transferLeaseOptions{
			goal:                     followTheWorkload,
			checkTransferLeaseSource: true,
			checkCandidateFullness:   true,
			dryRun:                   dryRun,
		},
	)
	return false, err

}

func replicationChangesForRebalance(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	numExistingVoters int,
	addTarget, removeTarget roachpb.ReplicationTarget,
	rebalanceTargetType targetReplicaType,
) (chgs []roachpb.ReplicationChange, performingSwap bool, err error) {
	__antithesis_instrumentation__.Notify(121459)
	if rebalanceTargetType == voterTarget && func() bool {
		__antithesis_instrumentation__.Notify(121462)
		return numExistingVoters == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(121463)

		chgs = []roachpb.ReplicationChange{
			{ChangeType: roachpb.ADD_VOTER, Target: addTarget},
		}
		log.VEventf(ctx, 1, "can't swap replica due to lease; falling back to add")
		return chgs, false, err
	} else {
		__antithesis_instrumentation__.Notify(121464)
	}
	__antithesis_instrumentation__.Notify(121460)

	rdesc, found := desc.GetReplicaDescriptor(addTarget.StoreID)
	switch rebalanceTargetType {
	case voterTarget:
		__antithesis_instrumentation__.Notify(121465)

		if found && func() bool {
			__antithesis_instrumentation__.Notify(121469)
			return rdesc.GetType() == roachpb.NON_VOTER == true
		}() == true {
			__antithesis_instrumentation__.Notify(121470)

			promo := roachpb.ReplicationChangesForPromotion(addTarget)
			demo := roachpb.ReplicationChangesForDemotion(removeTarget)
			chgs = append(promo, demo...)
			performingSwap = true
		} else {
			__antithesis_instrumentation__.Notify(121471)
			if found {
				__antithesis_instrumentation__.Notify(121472)
				return nil, false, errors.AssertionFailedf(
					"programming error:"+
						" store being rebalanced to(%s) already has a voting replica", addTarget.StoreID,
				)
			} else {
				__antithesis_instrumentation__.Notify(121473)

				chgs = []roachpb.ReplicationChange{
					{ChangeType: roachpb.ADD_VOTER, Target: addTarget},
					{ChangeType: roachpb.REMOVE_VOTER, Target: removeTarget},
				}
			}
		}
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(121466)
		if found {
			__antithesis_instrumentation__.Notify(121474)

			return nil, false, errors.AssertionFailedf(
				"invalid rebalancing decision: trying to"+
					" move non-voter to a store that already has a replica %s for the range", rdesc,
			)
		} else {
			__antithesis_instrumentation__.Notify(121475)
		}
		__antithesis_instrumentation__.Notify(121467)
		chgs = []roachpb.ReplicationChange{
			{ChangeType: roachpb.ADD_NON_VOTER, Target: addTarget},
			{ChangeType: roachpb.REMOVE_NON_VOTER, Target: removeTarget},
		}
	default:
		__antithesis_instrumentation__.Notify(121468)
	}
	__antithesis_instrumentation__.Notify(121461)
	return chgs, performingSwap, nil
}

type transferLeaseGoal int

const (
	followTheWorkload transferLeaseGoal = iota
	leaseCountConvergence
	qpsConvergence
)

type transferLeaseOptions struct {
	goal transferLeaseGoal

	checkTransferLeaseSource bool

	checkCandidateFullness bool
	dryRun                 bool
}

type leaseTransferOutcome int

const (
	transferErr leaseTransferOutcome = iota
	transferOK
	noTransferDryRun
	noSuitableTarget
)

func (o leaseTransferOutcome) String() string {
	__antithesis_instrumentation__.Notify(121476)
	switch o {
	case transferErr:
		__antithesis_instrumentation__.Notify(121477)
		return "err"
	case transferOK:
		__antithesis_instrumentation__.Notify(121478)
		return "ok"
	case noTransferDryRun:
		__antithesis_instrumentation__.Notify(121479)
		return "no transfer; dry run"
	case noSuitableTarget:
		__antithesis_instrumentation__.Notify(121480)
		return "no suitable transfer target found"
	default:
		__antithesis_instrumentation__.Notify(121481)
		return fmt.Sprintf("unexpected status value: %d", o)
	}
}

func (rq *replicateQueue) shedLease(
	ctx context.Context,
	repl *Replica,
	desc *roachpb.RangeDescriptor,
	conf roachpb.SpanConfig,
	opts transferLeaseOptions,
) (leaseTransferOutcome, error) {
	__antithesis_instrumentation__.Notify(121482)

	target := rq.allocator.TransferLeaseTarget(
		ctx,
		conf,
		desc.Replicas().VoterDescriptors(),
		repl,
		repl.leaseholderStats,
		false,
		opts,
	)
	if target == (roachpb.ReplicaDescriptor{}) {
		__antithesis_instrumentation__.Notify(121487)
		return noSuitableTarget, nil
	} else {
		__antithesis_instrumentation__.Notify(121488)
	}
	__antithesis_instrumentation__.Notify(121483)

	if opts.dryRun {
		__antithesis_instrumentation__.Notify(121489)
		log.VEventf(ctx, 1, "transferring lease to s%d", target.StoreID)
		return noTransferDryRun, nil
	} else {
		__antithesis_instrumentation__.Notify(121490)
	}
	__antithesis_instrumentation__.Notify(121484)

	avgQPS, qpsMeasurementDur := repl.leaseholderStats.avgQPS()
	if qpsMeasurementDur < MinStatsDuration {
		__antithesis_instrumentation__.Notify(121491)
		avgQPS = 0
	} else {
		__antithesis_instrumentation__.Notify(121492)
	}
	__antithesis_instrumentation__.Notify(121485)
	if err := rq.transferLease(ctx, repl, target, avgQPS); err != nil {
		__antithesis_instrumentation__.Notify(121493)
		return transferErr, err
	} else {
		__antithesis_instrumentation__.Notify(121494)
	}
	__antithesis_instrumentation__.Notify(121486)
	return transferOK, nil
}

func (rq *replicateQueue) transferLease(
	ctx context.Context, repl *Replica, target roachpb.ReplicaDescriptor, rangeQPS float64,
) error {
	__antithesis_instrumentation__.Notify(121495)
	rq.metrics.TransferLeaseCount.Inc(1)
	log.VEventf(ctx, 1, "transferring lease to s%d", target.StoreID)
	if err := repl.AdminTransferLease(ctx, target.StoreID); err != nil {
		__antithesis_instrumentation__.Notify(121497)
		return errors.Wrapf(err, "%s: unable to transfer lease to s%d", repl, target.StoreID)
	} else {
		__antithesis_instrumentation__.Notify(121498)
	}
	__antithesis_instrumentation__.Notify(121496)
	rq.lastLeaseTransfer.Store(timeutil.Now())
	rq.allocator.storePool.updateLocalStoresAfterLeaseTransfer(
		repl.store.StoreID(), target.StoreID, rangeQPS)
	return nil
}

func (rq *replicateQueue) changeReplicas(
	ctx context.Context,
	repl *Replica,
	chgs roachpb.ReplicationChanges,
	desc *roachpb.RangeDescriptor,
	priority kvserverpb.SnapshotRequest_Priority,
	reason kvserverpb.RangeLogEventReason,
	details string,
	dryRun bool,
) error {
	__antithesis_instrumentation__.Notify(121499)
	if dryRun {
		__antithesis_instrumentation__.Notify(121503)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(121504)
	}
	__antithesis_instrumentation__.Notify(121500)

	if _, err := repl.changeReplicasImpl(ctx, desc, priority, reason, details, chgs); err != nil {
		__antithesis_instrumentation__.Notify(121505)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121506)
	}
	__antithesis_instrumentation__.Notify(121501)
	rangeUsageInfo := rangeUsageInfoForRepl(repl)
	for _, chg := range chgs {
		__antithesis_instrumentation__.Notify(121507)
		rq.allocator.storePool.updateLocalStoreAfterRebalance(
			chg.Target.StoreID, rangeUsageInfo, chg.ChangeType)
	}
	__antithesis_instrumentation__.Notify(121502)
	return nil
}

func (rq *replicateQueue) canTransferLeaseFrom(ctx context.Context, repl *Replica) bool {
	__antithesis_instrumentation__.Notify(121508)

	respectsLeasePreferences, err := repl.checkLeaseRespectsPreferences(ctx)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(121511)
		return !respectsLeasePreferences == true
	}() == true {
		__antithesis_instrumentation__.Notify(121512)
		return true
	} else {
		__antithesis_instrumentation__.Notify(121513)
	}
	__antithesis_instrumentation__.Notify(121509)
	if lastLeaseTransfer := rq.lastLeaseTransfer.Load(); lastLeaseTransfer != nil {
		__antithesis_instrumentation__.Notify(121514)
		minInterval := MinLeaseTransferInterval.Get(&rq.store.cfg.Settings.SV)
		return timeutil.Since(lastLeaseTransfer.(time.Time)) > minInterval
	} else {
		__antithesis_instrumentation__.Notify(121515)
	}
	__antithesis_instrumentation__.Notify(121510)
	return true
}

func (*replicateQueue) timer(_ time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(121516)
	return replicateQueueTimerDuration
}

func (rq *replicateQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(121517)
	return rq.updateChan
}

func rangeRaftProgress(raftStatus *raft.Status, replicas []roachpb.ReplicaDescriptor) string {
	__antithesis_instrumentation__.Notify(121518)
	if raftStatus == nil {
		__antithesis_instrumentation__.Notify(121521)
		return "[no raft status]"
	} else {
		__antithesis_instrumentation__.Notify(121522)
		if len(raftStatus.Progress) == 0 {
			__antithesis_instrumentation__.Notify(121523)
			return "[no raft progress]"
		} else {
			__antithesis_instrumentation__.Notify(121524)
		}
	}
	__antithesis_instrumentation__.Notify(121519)
	var buf bytes.Buffer
	buf.WriteString("[")
	for i, r := range replicas {
		__antithesis_instrumentation__.Notify(121525)
		if i > 0 {
			__antithesis_instrumentation__.Notify(121528)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(121529)
		}
		__antithesis_instrumentation__.Notify(121526)
		fmt.Fprintf(&buf, "%d", r.ReplicaID)
		if uint64(r.ReplicaID) == raftStatus.Lead {
			__antithesis_instrumentation__.Notify(121530)
			buf.WriteString("*")
		} else {
			__antithesis_instrumentation__.Notify(121531)
		}
		__antithesis_instrumentation__.Notify(121527)
		if progress, ok := raftStatus.Progress[uint64(r.ReplicaID)]; ok {
			__antithesis_instrumentation__.Notify(121532)
			fmt.Fprintf(&buf, ":%d", progress.Match)
		} else {
			__antithesis_instrumentation__.Notify(121533)
			buf.WriteString(":?")
		}
	}
	__antithesis_instrumentation__.Notify(121520)
	buf.WriteString("]")
	return buf.String()
}
