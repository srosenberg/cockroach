package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3"
)

const (
	defaultLoadBasedRebalancingInterval = time.Minute

	minQPSThresholdDifference = 100
)

var (
	metaStoreRebalancerLeaseTransferCount = metric.Metadata{
		Name:        "rebalancing.lease.transfers",
		Help:        "Number of lease transfers motivated by store-level load imbalances",
		Measurement: "Lease Transfers",
		Unit:        metric.Unit_COUNT,
	}
	metaStoreRebalancerRangeRebalanceCount = metric.Metadata{
		Name:        "rebalancing.range.rebalances",
		Help:        "Number of range rebalance operations motivated by store-level load imbalances",
		Measurement: "Range Rebalances",
		Unit:        metric.Unit_COUNT,
	}
)

type StoreRebalancerMetrics struct {
	LeaseTransferCount  *metric.Counter
	RangeRebalanceCount *metric.Counter
}

func makeStoreRebalancerMetrics() StoreRebalancerMetrics {
	__antithesis_instrumentation__.Notify(125492)
	return StoreRebalancerMetrics{
		LeaseTransferCount:  metric.NewCounter(metaStoreRebalancerLeaseTransferCount),
		RangeRebalanceCount: metric.NewCounter(metaStoreRebalancerRangeRebalanceCount),
	}
}

var LoadBasedRebalancingMode = settings.RegisterEnumSetting(
	settings.SystemOnly,
	"kv.allocator.load_based_rebalancing",
	"whether to rebalance based on the distribution of QPS across stores",
	"leases and replicas",
	map[int64]string{
		int64(LBRebalancingOff):               "off",
		int64(LBRebalancingLeasesOnly):        "leases",
		int64(LBRebalancingLeasesAndReplicas): "leases and replicas",
	},
).WithPublic()

var qpsRebalanceThreshold = func() *settings.FloatSetting {
	__antithesis_instrumentation__.Notify(125493)
	s := settings.RegisterFloatSetting(
		settings.SystemOnly,
		"kv.allocator.qps_rebalance_threshold",
		"minimum fraction away from the mean a store's QPS (such as queries per second) can be before it is considered overfull or underfull",
		0.25,
		settings.NonNegativeFloat,
	)
	s.SetVisibility(settings.Public)
	return s
}()

var loadBasedRebalanceInterval = settings.RegisterPublicDurationSettingWithExplicitUnit(
	settings.SystemOnly,
	"kv.allocator.load_based_rebalancing_interval",
	"the rough interval at which each store will check for load-based lease / replica rebalancing opportunities",
	defaultLoadBasedRebalancingInterval,
	func(d time.Duration) error {
		__antithesis_instrumentation__.Notify(125494)

		const min = 10 * time.Second
		if d < min {
			__antithesis_instrumentation__.Notify(125496)
			return errors.Errorf("must specify a minimum of %s", min)
		} else {
			__antithesis_instrumentation__.Notify(125497)
		}
		__antithesis_instrumentation__.Notify(125495)
		return nil
	},
)

var minQPSDifferenceForTransfers = func() *settings.FloatSetting {
	__antithesis_instrumentation__.Notify(125498)
	s := settings.RegisterFloatSetting(
		settings.SystemOnly,
		"kv.allocator.min_qps_difference_for_transfers",
		"the minimum qps difference that must exist between any two stores"+
			" for the allocator to allow a lease or replica transfer between them",
		2*minQPSThresholdDifference,
		settings.NonNegativeFloat,
	)
	s.SetVisibility(settings.Reserved)
	return s
}()

type LBRebalancingMode int64

const (
	LBRebalancingOff LBRebalancingMode = iota

	LBRebalancingLeasesOnly

	LBRebalancingLeasesAndReplicas
)

type StoreRebalancer struct {
	log.AmbientContext
	metrics         StoreRebalancerMetrics
	st              *cluster.Settings
	rq              *replicateQueue
	replRankings    *replicaRankings
	getRaftStatusFn func(replica *Replica) *raft.Status
}

func NewStoreRebalancer(
	ambientCtx log.AmbientContext,
	st *cluster.Settings,
	rq *replicateQueue,
	replRankings *replicaRankings,
) *StoreRebalancer {
	__antithesis_instrumentation__.Notify(125499)
	sr := &StoreRebalancer{
		AmbientContext: ambientCtx,
		metrics:        makeStoreRebalancerMetrics(),
		st:             st,
		rq:             rq,
		replRankings:   replRankings,
		getRaftStatusFn: func(replica *Replica) *raft.Status {
			__antithesis_instrumentation__.Notify(125501)
			return replica.RaftStatus()
		},
	}
	__antithesis_instrumentation__.Notify(125500)
	sr.AddLogTag("store-rebalancer", nil)
	sr.rq.store.metrics.registry.AddMetricStruct(&sr.metrics)
	return sr
}

func (sr *StoreRebalancer) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(125502)
	ctx = sr.AnnotateCtx(ctx)

	_ = stopper.RunAsyncTask(ctx, "store-rebalancer", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(125503)
		timer := timeutil.NewTimer()
		defer timer.Stop()
		timer.Reset(jitteredInterval(loadBasedRebalanceInterval.Get(&sr.st.SV)))
		for {
			__antithesis_instrumentation__.Notify(125504)

			select {
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(125507)
				return
			case <-timer.C:
				__antithesis_instrumentation__.Notify(125508)
				timer.Read = true
				timer.Reset(jitteredInterval(loadBasedRebalanceInterval.Get(&sr.st.SV)))
			}
			__antithesis_instrumentation__.Notify(125505)

			mode := LBRebalancingMode(LoadBasedRebalancingMode.Get(&sr.st.SV))
			if mode == LBRebalancingOff {
				__antithesis_instrumentation__.Notify(125509)
				continue
			} else {
				__antithesis_instrumentation__.Notify(125510)
			}
			__antithesis_instrumentation__.Notify(125506)

			storeList, _, _ := sr.rq.allocator.storePool.getStoreList(storeFilterSuspect)
			sr.rebalanceStore(ctx, mode, storeList)
		}
	})
}

func (sr *StoreRebalancer) scorerOptions() *qpsScorerOptions {
	__antithesis_instrumentation__.Notify(125511)
	return &qpsScorerOptions{
		deterministic:         sr.rq.allocator.storePool.deterministic,
		qpsRebalanceThreshold: qpsRebalanceThreshold.Get(&sr.st.SV),
		minRequiredQPSDiff:    minQPSDifferenceForTransfers.Get(&sr.st.SV),
	}
}

func (sr *StoreRebalancer) rebalanceStore(
	ctx context.Context, mode LBRebalancingMode, allStoresList StoreList,
) {
	__antithesis_instrumentation__.Notify(125512)
	options := sr.scorerOptions()
	var localDesc *roachpb.StoreDescriptor
	for i := range allStoresList.stores {
		__antithesis_instrumentation__.Notify(125520)
		if allStoresList.stores[i].StoreID == sr.rq.store.StoreID() {
			__antithesis_instrumentation__.Notify(125521)
			localDesc = &allStoresList.stores[i]
			break
		} else {
			__antithesis_instrumentation__.Notify(125522)
		}
	}
	__antithesis_instrumentation__.Notify(125513)
	if localDesc == nil {
		__antithesis_instrumentation__.Notify(125523)
		log.Warningf(ctx, "StorePool missing descriptor for local store")
		return
	} else {
		__antithesis_instrumentation__.Notify(125524)
	}
	__antithesis_instrumentation__.Notify(125514)

	qpsMaxThreshold := overfullQPSThreshold(options, allStoresList.candidateQueriesPerSecond.mean)
	if !(localDesc.Capacity.QueriesPerSecond > qpsMaxThreshold) {
		__antithesis_instrumentation__.Notify(125525)
		log.Infof(ctx, "local QPS %.2f is below max threshold %.2f (mean=%.2f); no rebalancing needed",
			localDesc.Capacity.QueriesPerSecond, qpsMaxThreshold, allStoresList.candidateQueriesPerSecond.mean)
		return
	} else {
		__antithesis_instrumentation__.Notify(125526)
	}
	__antithesis_instrumentation__.Notify(125515)

	var replicasToMaybeRebalance []replicaWithStats
	storeMap := storeListToMap(allStoresList)

	log.Infof(ctx,
		"considering load-based lease transfers for s%d with %.2f qps (mean=%.2f, upperThreshold=%.2f)",
		localDesc.StoreID, localDesc.Capacity.QueriesPerSecond, allStoresList.candidateQueriesPerSecond.mean, qpsMaxThreshold)
	hottestRanges := sr.replRankings.topQPS()
	for localDesc.Capacity.QueriesPerSecond > qpsMaxThreshold {
		__antithesis_instrumentation__.Notify(125527)
		replWithStats, target, considerForRebalance := sr.chooseLeaseToTransfer(
			ctx,
			&hottestRanges,
			localDesc,
			allStoresList,
			storeMap,
		)
		replicasToMaybeRebalance = append(replicasToMaybeRebalance, considerForRebalance...)
		if replWithStats.repl == nil {
			__antithesis_instrumentation__.Notify(125530)
			break
		} else {
			__antithesis_instrumentation__.Notify(125531)
		}
		__antithesis_instrumentation__.Notify(125528)

		timeout := sr.rq.processTimeoutFunc(sr.st, replWithStats.repl)
		if err := contextutil.RunWithTimeout(ctx, "transfer lease", timeout, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(125532)
			return sr.rq.transferLease(ctx, replWithStats.repl, target, replWithStats.qps)
		}); err != nil {
			__antithesis_instrumentation__.Notify(125533)
			log.Errorf(ctx, "unable to transfer lease to s%d: %+v", target.StoreID, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125534)
		}
		__antithesis_instrumentation__.Notify(125529)
		sr.metrics.LeaseTransferCount.Inc(1)

		localDesc.Capacity.LeaseCount--
		localDesc.Capacity.QueriesPerSecond -= replWithStats.qps
		if otherDesc := storeMap[target.StoreID]; otherDesc != nil {
			__antithesis_instrumentation__.Notify(125535)
			otherDesc.Capacity.LeaseCount++
			otherDesc.Capacity.QueriesPerSecond += replWithStats.qps
		} else {
			__antithesis_instrumentation__.Notify(125536)
		}
	}
	__antithesis_instrumentation__.Notify(125516)

	if !(localDesc.Capacity.QueriesPerSecond > qpsMaxThreshold) {
		__antithesis_instrumentation__.Notify(125537)
		log.Infof(ctx,
			"load-based lease transfers successfully brought s%d down to %.2f qps (mean=%.2f, upperThreshold=%.2f)",
			localDesc.StoreID, localDesc.Capacity.QueriesPerSecond, allStoresList.candidateQueriesPerSecond.mean, qpsMaxThreshold)
		return
	} else {
		__antithesis_instrumentation__.Notify(125538)
	}
	__antithesis_instrumentation__.Notify(125517)

	if mode != LBRebalancingLeasesAndReplicas {
		__antithesis_instrumentation__.Notify(125539)
		log.Infof(ctx,
			"ran out of leases worth transferring and qps (%.2f) is still above desired threshold (%.2f)",
			localDesc.Capacity.QueriesPerSecond, qpsMaxThreshold)
		return
	} else {
		__antithesis_instrumentation__.Notify(125540)
	}
	__antithesis_instrumentation__.Notify(125518)
	log.Infof(ctx,
		"ran out of leases worth transferring and qps (%.2f) is still above desired threshold (%.2f); considering load-based replica rebalances",
		localDesc.Capacity.QueriesPerSecond, qpsMaxThreshold)

	replicasToMaybeRebalance = append(replicasToMaybeRebalance, hottestRanges...)

	for localDesc.Capacity.QueriesPerSecond > qpsMaxThreshold {
		__antithesis_instrumentation__.Notify(125541)
		replWithStats, voterTargets, nonVoterTargets := sr.chooseRangeToRebalance(
			ctx,
			&replicasToMaybeRebalance,
			localDesc,
			allStoresList,
			sr.scorerOptions(),
		)
		if replWithStats.repl == nil {
			__antithesis_instrumentation__.Notify(125545)
			log.Infof(ctx,
				"ran out of replicas worth transferring and qps (%.2f) is still above desired threshold (%.2f); will check again soon",
				localDesc.Capacity.QueriesPerSecond, qpsMaxThreshold)
			return
		} else {
			__antithesis_instrumentation__.Notify(125546)
		}
		__antithesis_instrumentation__.Notify(125542)

		descBeforeRebalance := replWithStats.repl.Desc()
		log.VEventf(
			ctx,
			1,
			"rebalancing r%d (%.2f qps) to better balance load: voters from %v to %v; non-voters from %v to %v",
			replWithStats.repl.RangeID,
			replWithStats.qps,
			descBeforeRebalance.Replicas().Voters(),
			voterTargets,
			descBeforeRebalance.Replicas().NonVoters(),
			nonVoterTargets,
		)

		timeout := sr.rq.processTimeoutFunc(sr.st, replWithStats.repl)
		if err := contextutil.RunWithTimeout(ctx, "relocate range", timeout, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(125547)
			return sr.rq.store.DB().AdminRelocateRange(
				ctx,
				descBeforeRebalance.StartKey.AsRawKey(),
				voterTargets,
				nonVoterTargets,
				true,
			)
		}); err != nil {
			__antithesis_instrumentation__.Notify(125548)
			log.Errorf(ctx, "unable to relocate range to %v: %v", voterTargets, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125549)
		}
		__antithesis_instrumentation__.Notify(125543)
		sr.metrics.RangeRebalanceCount.Inc(1)

		replicasBeforeRebalance := descBeforeRebalance.Replicas().Descriptors()
		for i := range replicasBeforeRebalance {
			__antithesis_instrumentation__.Notify(125550)
			if storeDesc := storeMap[replicasBeforeRebalance[i].StoreID]; storeDesc != nil {
				__antithesis_instrumentation__.Notify(125551)
				storeDesc.Capacity.RangeCount--
			} else {
				__antithesis_instrumentation__.Notify(125552)
			}
		}
		__antithesis_instrumentation__.Notify(125544)
		localDesc.Capacity.LeaseCount--
		localDesc.Capacity.QueriesPerSecond -= replWithStats.qps
		for i := range voterTargets {
			__antithesis_instrumentation__.Notify(125553)
			if storeDesc := storeMap[voterTargets[i].StoreID]; storeDesc != nil {
				__antithesis_instrumentation__.Notify(125554)
				storeDesc.Capacity.RangeCount++
				if i == 0 {
					__antithesis_instrumentation__.Notify(125555)
					storeDesc.Capacity.LeaseCount++
					storeDesc.Capacity.QueriesPerSecond += replWithStats.qps
				} else {
					__antithesis_instrumentation__.Notify(125556)
				}
			} else {
				__antithesis_instrumentation__.Notify(125557)
			}
		}
	}
	__antithesis_instrumentation__.Notify(125519)

	log.Infof(ctx,
		"load-based replica transfers successfully brought s%d down to %.2f qps (mean=%.2f, upperThreshold=%.2f)",
		localDesc.StoreID, localDesc.Capacity.QueriesPerSecond, allStoresList.candidateQueriesPerSecond.mean, qpsMaxThreshold)
}

func (sr *StoreRebalancer) chooseLeaseToTransfer(
	ctx context.Context,
	hottestRanges *[]replicaWithStats,
	localDesc *roachpb.StoreDescriptor,
	storeList StoreList,
	storeMap map[roachpb.StoreID]*roachpb.StoreDescriptor,
) (replicaWithStats, roachpb.ReplicaDescriptor, []replicaWithStats) {
	__antithesis_instrumentation__.Notify(125558)
	var considerForRebalance []replicaWithStats
	now := sr.rq.store.Clock().NowAsClockTimestamp()
	for {
		__antithesis_instrumentation__.Notify(125559)
		if len(*hottestRanges) == 0 {
			__antithesis_instrumentation__.Notify(125567)
			return replicaWithStats{}, roachpb.ReplicaDescriptor{}, considerForRebalance
		} else {
			__antithesis_instrumentation__.Notify(125568)
		}
		__antithesis_instrumentation__.Notify(125560)
		replWithStats := (*hottestRanges)[0]
		*hottestRanges = (*hottestRanges)[1:]

		if replWithStats.repl == nil {
			__antithesis_instrumentation__.Notify(125569)
			return replicaWithStats{}, roachpb.ReplicaDescriptor{}, considerForRebalance
		} else {
			__antithesis_instrumentation__.Notify(125570)
		}
		__antithesis_instrumentation__.Notify(125561)

		if !replWithStats.repl.OwnsValidLease(ctx, now) {
			__antithesis_instrumentation__.Notify(125571)
			log.VEventf(ctx, 3, "store doesn't own the lease for r%d", replWithStats.repl.RangeID)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125572)
		}
		__antithesis_instrumentation__.Notify(125562)

		const minQPSFraction = .001
		if replWithStats.qps < localDesc.Capacity.QueriesPerSecond*minQPSFraction {
			__antithesis_instrumentation__.Notify(125573)
			log.VEventf(ctx, 3, "r%d's %.2f qps is too little to matter relative to s%d's %.2f total qps",
				replWithStats.repl.RangeID, replWithStats.qps, localDesc.StoreID, localDesc.Capacity.QueriesPerSecond)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125574)
		}
		__antithesis_instrumentation__.Notify(125563)

		desc, conf := replWithStats.repl.DescAndSpanConfig()
		log.VEventf(ctx, 3, "considering lease transfer for r%d with %.2f qps",
			desc.RangeID, replWithStats.qps)

		candidates := desc.Replicas().DeepCopy().VoterDescriptors()

		candidates = filterBehindReplicas(ctx, sr.getRaftStatusFn(replWithStats.repl), candidates)

		candidate := sr.rq.allocator.TransferLeaseTarget(
			ctx,
			conf,
			candidates,
			replWithStats.repl,
			replWithStats.repl.leaseholderStats,
			true,
			transferLeaseOptions{
				goal:                     qpsConvergence,
				checkTransferLeaseSource: true,
			},
		)

		if candidate == (roachpb.ReplicaDescriptor{}) {
			__antithesis_instrumentation__.Notify(125575)
			log.VEventf(
				ctx,
				3,
				"could not find a better lease transfer target for r%d; considering replica rebalance instead",
				desc.RangeID,
			)
			considerForRebalance = append(considerForRebalance, replWithStats)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125576)
		}
		__antithesis_instrumentation__.Notify(125564)

		filteredStoreList := storeList.excludeInvalid(conf.Constraints)
		filteredStoreList = storeList.excludeInvalid(conf.VoterConstraints)
		if sr.rq.allocator.followTheWorkloadPrefersLocal(
			ctx,
			filteredStoreList,
			*localDesc,
			candidate.StoreID,
			candidates,
			replWithStats.repl.leaseholderStats,
		) {
			__antithesis_instrumentation__.Notify(125577)
			log.VEventf(
				ctx, 3, "r%d is on s%d due to follow-the-workload; considering replica rebalance instead",
				desc.RangeID, localDesc.StoreID,
			)
			considerForRebalance = append(considerForRebalance, replWithStats)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125578)
		}
		__antithesis_instrumentation__.Notify(125565)
		if targetStore, ok := storeMap[candidate.StoreID]; ok {
			__antithesis_instrumentation__.Notify(125579)
			log.VEventf(
				ctx,
				1,
				"transferring lease for r%d (qps=%.2f) to store s%d (qps=%.2f) from local store s%d (qps=%.2f)",
				desc.RangeID,
				replWithStats.qps,
				targetStore.StoreID,
				targetStore.Capacity.QueriesPerSecond,
				localDesc.StoreID,
				localDesc.Capacity.QueriesPerSecond,
			)
		} else {
			__antithesis_instrumentation__.Notify(125580)
		}
		__antithesis_instrumentation__.Notify(125566)
		return replWithStats, candidate, considerForRebalance
	}
}

type rangeRebalanceContext struct {
	replWithStats replicaWithStats
	rangeDesc     *roachpb.RangeDescriptor
	conf          roachpb.SpanConfig
}

func (sr *StoreRebalancer) chooseRangeToRebalance(
	ctx context.Context,
	hottestRanges *[]replicaWithStats,
	localDesc *roachpb.StoreDescriptor,
	allStoresList StoreList,
	options *qpsScorerOptions,
) (replWithStats replicaWithStats, voterTargets, nonVoterTargets []roachpb.ReplicationTarget) {
	__antithesis_instrumentation__.Notify(125581)
	now := sr.rq.store.Clock().NowAsClockTimestamp()
	for {
		__antithesis_instrumentation__.Notify(125582)
		if len(*hottestRanges) == 0 {
			__antithesis_instrumentation__.Notify(125591)
			return replicaWithStats{}, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(125592)
		}
		__antithesis_instrumentation__.Notify(125583)
		replWithStats := (*hottestRanges)[0]
		*hottestRanges = (*hottestRanges)[1:]

		if replWithStats.repl == nil {
			__antithesis_instrumentation__.Notify(125593)
			return replicaWithStats{}, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(125594)
		}
		__antithesis_instrumentation__.Notify(125584)

		const minQPSFraction = .001
		if replWithStats.qps < localDesc.Capacity.QueriesPerSecond*minQPSFraction {
			__antithesis_instrumentation__.Notify(125595)
			log.VEventf(
				ctx,
				5,
				"r%d's %.2f qps is too little to matter relative to s%d's %.2f total qps",
				replWithStats.repl.RangeID,
				replWithStats.qps,
				localDesc.StoreID,
				localDesc.Capacity.QueriesPerSecond,
			)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125596)
		}
		__antithesis_instrumentation__.Notify(125585)

		rangeDesc, conf := replWithStats.repl.DescAndSpanConfig()
		clusterNodes := sr.rq.allocator.storePool.ClusterNodeCount()
		numDesiredVoters := GetNeededVoters(conf.GetNumVoters(), clusterNodes)
		numDesiredNonVoters := GetNeededNonVoters(numDesiredVoters, int(conf.GetNumNonVoters()), clusterNodes)
		if expected, actual := numDesiredVoters, len(rangeDesc.Replicas().VoterDescriptors()); expected != actual {
			__antithesis_instrumentation__.Notify(125597)
			log.VEventf(
				ctx,
				3,
				"r%d is either over or under replicated (expected %d voters, found %d); ignoring",
				rangeDesc.RangeID,
				expected,
				actual,
			)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125598)
		}
		__antithesis_instrumentation__.Notify(125586)
		if expected, actual := numDesiredNonVoters, len(rangeDesc.Replicas().NonVoterDescriptors()); expected != actual {
			__antithesis_instrumentation__.Notify(125599)
			log.VEventf(
				ctx,
				3,
				"r%d is either over or under replicated (expected %d non-voters, found %d); ignoring",
				rangeDesc.RangeID,
				expected,
				actual,
			)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125600)
		}
		__antithesis_instrumentation__.Notify(125587)
		rebalanceCtx := rangeRebalanceContext{
			replWithStats: replWithStats,
			rangeDesc:     rangeDesc,
			conf:          conf,
		}

		options.qpsPerReplica = replWithStats.qps

		if !replWithStats.repl.OwnsValidLease(ctx, now) {
			__antithesis_instrumentation__.Notify(125601)
			log.VEventf(ctx, 3, "store doesn't own the lease for r%d", replWithStats.repl.RangeID)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125602)
		}
		__antithesis_instrumentation__.Notify(125588)

		log.VEventf(
			ctx,
			3,
			"considering replica rebalance for r%d with %.2f qps",
			replWithStats.repl.GetRangeID(),
			replWithStats.qps,
		)

		targetVoterRepls, targetNonVoterRepls, foundRebalance := sr.getRebalanceTargetsBasedOnQPS(
			ctx,
			rebalanceCtx,
			options,
		)

		if !foundRebalance {
			__antithesis_instrumentation__.Notify(125603)

			log.VEventf(ctx, 3, "could not find rebalance opportunities for r%d", replWithStats.repl.RangeID)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125604)
		}
		__antithesis_instrumentation__.Notify(125589)

		storeDescMap := storeListToMap(allStoresList)

		newLeaseIdx := 0
		newLeaseQPS := math.MaxFloat64
		var raftStatus *raft.Status
		for i := 0; i < len(targetVoterRepls); i++ {
			__antithesis_instrumentation__.Notify(125605)

			if replica, ok := rangeDesc.GetReplicaDescriptor(targetVoterRepls[i].StoreID); ok {
				__antithesis_instrumentation__.Notify(125607)
				if raftStatus == nil {
					__antithesis_instrumentation__.Notify(125609)
					raftStatus = sr.getRaftStatusFn(replWithStats.repl)
				} else {
					__antithesis_instrumentation__.Notify(125610)
				}
				__antithesis_instrumentation__.Notify(125608)
				if replicaIsBehind(raftStatus, replica.ReplicaID) {
					__antithesis_instrumentation__.Notify(125611)
					continue
				} else {
					__antithesis_instrumentation__.Notify(125612)
				}
			} else {
				__antithesis_instrumentation__.Notify(125613)
			}
			__antithesis_instrumentation__.Notify(125606)

			storeDesc, ok := storeDescMap[targetVoterRepls[i].StoreID]
			if ok && func() bool {
				__antithesis_instrumentation__.Notify(125614)
				return storeDesc.Capacity.QueriesPerSecond < newLeaseQPS == true
			}() == true {
				__antithesis_instrumentation__.Notify(125615)
				newLeaseIdx = i
				newLeaseQPS = storeDesc.Capacity.QueriesPerSecond
			} else {
				__antithesis_instrumentation__.Notify(125616)
			}
		}
		__antithesis_instrumentation__.Notify(125590)
		targetVoterRepls[0], targetVoterRepls[newLeaseIdx] = targetVoterRepls[newLeaseIdx], targetVoterRepls[0]
		return replWithStats,
			roachpb.MakeReplicaSet(targetVoterRepls).ReplicationTargets(),
			roachpb.MakeReplicaSet(targetNonVoterRepls).ReplicationTargets()
	}
}

func (sr *StoreRebalancer) getRebalanceTargetsBasedOnQPS(
	ctx context.Context, rbCtx rangeRebalanceContext, options scorerOptions,
) (finalVoterTargets, finalNonVoterTargets []roachpb.ReplicaDescriptor, foundRebalance bool) {
	__antithesis_instrumentation__.Notify(125617)
	finalVoterTargets = rbCtx.rangeDesc.Replicas().VoterDescriptors()
	finalNonVoterTargets = rbCtx.rangeDesc.Replicas().NonVoterDescriptors()

	for i := 0; i < len(finalVoterTargets); i++ {
		__antithesis_instrumentation__.Notify(125620)

		add, remove, _, shouldRebalance := sr.rq.allocator.rebalanceTarget(
			ctx,
			rbCtx.conf,
			rbCtx.replWithStats.repl.RaftStatus(),
			finalVoterTargets,
			finalNonVoterTargets,
			rangeUsageInfoForRepl(rbCtx.replWithStats.repl),
			storeFilterSuspect,
			voterTarget,
			options,
		)
		if !shouldRebalance {
			__antithesis_instrumentation__.Notify(125624)
			log.VEventf(
				ctx,
				3,
				"no more rebalancing opportunities for r%d voters that improve QPS balance",
				rbCtx.rangeDesc.RangeID,
			)
			break
		} else {
			__antithesis_instrumentation__.Notify(125625)

			foundRebalance = true
		}
		__antithesis_instrumentation__.Notify(125621)
		log.VEventf(
			ctx,
			3,
			"rebalancing voter (qps=%.2f) for r%d on %v to %v in order to improve QPS balance",
			rbCtx.replWithStats.qps,
			rbCtx.rangeDesc.RangeID,
			remove,
			add,
		)

		afterVoters := make([]roachpb.ReplicaDescriptor, 0, len(finalVoterTargets))
		afterNonVoters := make([]roachpb.ReplicaDescriptor, 0, len(finalNonVoterTargets))
		for _, voter := range finalVoterTargets {
			__antithesis_instrumentation__.Notify(125626)
			if voter.StoreID == remove.StoreID {
				__antithesis_instrumentation__.Notify(125627)
				afterVoters = append(
					afterVoters, roachpb.ReplicaDescriptor{
						StoreID: add.StoreID,
						NodeID:  add.NodeID,
					})
			} else {
				__antithesis_instrumentation__.Notify(125628)
				afterVoters = append(afterVoters, voter)
			}
		}
		__antithesis_instrumentation__.Notify(125622)

		for _, nonVoter := range finalNonVoterTargets {
			__antithesis_instrumentation__.Notify(125629)
			if nonVoter.StoreID == add.StoreID {
				__antithesis_instrumentation__.Notify(125630)
				afterNonVoters = append(afterNonVoters, roachpb.ReplicaDescriptor{
					StoreID: remove.StoreID,
					NodeID:  remove.NodeID,
				})
			} else {
				__antithesis_instrumentation__.Notify(125631)
				afterNonVoters = append(afterNonVoters, nonVoter)
			}
		}
		__antithesis_instrumentation__.Notify(125623)

		finalVoterTargets = afterVoters
		finalNonVoterTargets = afterNonVoters
	}
	__antithesis_instrumentation__.Notify(125618)

	for i := 0; i < len(finalNonVoterTargets); i++ {
		__antithesis_instrumentation__.Notify(125632)
		add, remove, _, shouldRebalance := sr.rq.allocator.rebalanceTarget(
			ctx,
			rbCtx.conf,
			rbCtx.replWithStats.repl.RaftStatus(),
			finalVoterTargets,
			finalNonVoterTargets,
			rangeUsageInfoForRepl(rbCtx.replWithStats.repl),
			storeFilterSuspect,
			nonVoterTarget,
			options,
		)
		if !shouldRebalance {
			__antithesis_instrumentation__.Notify(125635)
			log.VEventf(
				ctx,
				3,
				"no more rebalancing opportunities for r%d non-voters that improve QPS balance",
				rbCtx.rangeDesc.RangeID,
			)
			break
		} else {
			__antithesis_instrumentation__.Notify(125636)

			foundRebalance = true
		}
		__antithesis_instrumentation__.Notify(125633)
		log.VEventf(
			ctx,
			3,
			"rebalancing non-voter (qps=%.2f) for r%d on %v to %v in order to improve QPS balance",
			rbCtx.replWithStats.qps,
			rbCtx.rangeDesc.RangeID,
			remove,
			add,
		)
		var newNonVoters []roachpb.ReplicaDescriptor
		for _, nonVoter := range finalNonVoterTargets {
			__antithesis_instrumentation__.Notify(125637)
			if nonVoter.StoreID == remove.StoreID {
				__antithesis_instrumentation__.Notify(125638)
				newNonVoters = append(
					newNonVoters, roachpb.ReplicaDescriptor{
						StoreID: add.StoreID,
						NodeID:  add.NodeID,
					})
			} else {
				__antithesis_instrumentation__.Notify(125639)
				newNonVoters = append(newNonVoters, nonVoter)
			}
		}
		__antithesis_instrumentation__.Notify(125634)

		finalNonVoterTargets = newNonVoters
	}
	__antithesis_instrumentation__.Notify(125619)
	return finalVoterTargets, finalNonVoterTargets, foundRebalance
}

func storeListToMap(sl StoreList) map[roachpb.StoreID]*roachpb.StoreDescriptor {
	__antithesis_instrumentation__.Notify(125640)
	storeMap := make(map[roachpb.StoreID]*roachpb.StoreDescriptor)
	for i := range sl.stores {
		__antithesis_instrumentation__.Notify(125642)
		storeMap[sl.stores[i].StoreID] = &sl.stores[i]
	}
	__antithesis_instrumentation__.Notify(125641)
	return storeMap
}

func jitteredInterval(interval time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(125643)
	return time.Duration(float64(interval) * (0.75 + 0.5*rand.Float64()))
}
