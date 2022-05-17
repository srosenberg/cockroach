package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/constraint"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/tracker"
)

const (
	leaseRebalanceThreshold = 0.05

	baseLoadBasedLeaseRebalanceThreshold = 2 * leaseRebalanceThreshold

	minReplicaWeight = 0.001
)

var MinLeaseTransferStatsDuration = 30 * time.Second

var enableLoadBasedLeaseRebalancing = settings.RegisterBoolSetting(
	settings.SystemOnly,
	"kv.allocator.load_based_lease_rebalancing.enabled",
	"set to enable rebalancing of range leases based on load and latency",
	true,
).WithPublic()

var leaseRebalancingAggressiveness = settings.RegisterFloatSetting(
	settings.SystemOnly,
	"kv.allocator.lease_rebalancing_aggressiveness",
	"set greater than 1.0 to rebalance leases toward load more aggressively, "+
		"or between 0 and 1.0 to be more conservative about rebalancing leases",
	1.0,
	settings.NonNegativeFloat,
)

type AllocatorAction int

const (
	_ AllocatorAction = iota
	AllocatorNoop
	AllocatorRemoveVoter
	AllocatorRemoveNonVoter
	AllocatorAddVoter
	AllocatorAddNonVoter
	AllocatorReplaceDeadVoter
	AllocatorReplaceDeadNonVoter
	AllocatorRemoveDeadVoter
	AllocatorRemoveDeadNonVoter
	AllocatorReplaceDecommissioningVoter
	AllocatorReplaceDecommissioningNonVoter
	AllocatorRemoveDecommissioningVoter
	AllocatorRemoveDecommissioningNonVoter
	AllocatorRemoveLearner
	AllocatorConsiderRebalance
	AllocatorRangeUnavailable
	AllocatorFinalizeAtomicReplicationChange
)

var allocatorActionNames = map[AllocatorAction]string{
	AllocatorNoop:                            "noop",
	AllocatorRemoveVoter:                     "remove voter",
	AllocatorRemoveNonVoter:                  "remove non-voter",
	AllocatorAddVoter:                        "add voter",
	AllocatorAddNonVoter:                     "add non-voter",
	AllocatorReplaceDeadVoter:                "replace dead voter",
	AllocatorReplaceDeadNonVoter:             "replace dead non-voter",
	AllocatorRemoveDeadVoter:                 "remove dead voter",
	AllocatorRemoveDeadNonVoter:              "remove dead non-voter",
	AllocatorReplaceDecommissioningVoter:     "replace decommissioning voter",
	AllocatorReplaceDecommissioningNonVoter:  "replace decommissioning non-voter",
	AllocatorRemoveDecommissioningVoter:      "remove decommissioning voter",
	AllocatorRemoveDecommissioningNonVoter:   "remove decommissioning non-voter",
	AllocatorRemoveLearner:                   "remove learner",
	AllocatorConsiderRebalance:               "consider rebalance",
	AllocatorRangeUnavailable:                "range unavailable",
	AllocatorFinalizeAtomicReplicationChange: "finalize conf change",
}

func (a AllocatorAction) String() string {
	__antithesis_instrumentation__.Notify(94115)
	return allocatorActionNames[a]
}

func (a AllocatorAction) Priority() float64 {
	__antithesis_instrumentation__.Notify(94116)
	switch a {
	case AllocatorFinalizeAtomicReplicationChange:
		__antithesis_instrumentation__.Notify(94117)
		return 12002
	case AllocatorRemoveLearner:
		__antithesis_instrumentation__.Notify(94118)
		return 12001
	case AllocatorReplaceDeadVoter:
		__antithesis_instrumentation__.Notify(94119)
		return 12000
	case AllocatorAddVoter:
		__antithesis_instrumentation__.Notify(94120)
		return 10000
	case AllocatorReplaceDecommissioningVoter:
		__antithesis_instrumentation__.Notify(94121)
		return 5000
	case AllocatorRemoveDeadVoter:
		__antithesis_instrumentation__.Notify(94122)
		return 1000
	case AllocatorRemoveDecommissioningVoter:
		__antithesis_instrumentation__.Notify(94123)
		return 900
	case AllocatorRemoveVoter:
		__antithesis_instrumentation__.Notify(94124)
		return 800
	case AllocatorReplaceDeadNonVoter:
		__antithesis_instrumentation__.Notify(94125)
		return 700
	case AllocatorAddNonVoter:
		__antithesis_instrumentation__.Notify(94126)
		return 600
	case AllocatorReplaceDecommissioningNonVoter:
		__antithesis_instrumentation__.Notify(94127)
		return 500
	case AllocatorRemoveDeadNonVoter:
		__antithesis_instrumentation__.Notify(94128)
		return 400
	case AllocatorRemoveDecommissioningNonVoter:
		__antithesis_instrumentation__.Notify(94129)
		return 300
	case AllocatorRemoveNonVoter:
		__antithesis_instrumentation__.Notify(94130)
		return 200
	case AllocatorConsiderRebalance, AllocatorRangeUnavailable, AllocatorNoop:
		__antithesis_instrumentation__.Notify(94131)
		return 0
	default:
		__antithesis_instrumentation__.Notify(94132)
		panic(fmt.Sprintf("unknown AllocatorAction: %s", a))
	}
}

type targetReplicaType int

const (
	_ targetReplicaType = iota
	voterTarget
	nonVoterTarget
)

type replicaStatus int

const (
	_ replicaStatus = iota
	alive
	dead
	decommissioning
)

func (t targetReplicaType) AddChangeType() roachpb.ReplicaChangeType {
	__antithesis_instrumentation__.Notify(94133)
	switch t {
	case voterTarget:
		__antithesis_instrumentation__.Notify(94134)
		return roachpb.ADD_VOTER
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(94135)
		return roachpb.ADD_NON_VOTER
	default:
		__antithesis_instrumentation__.Notify(94136)
		panic(fmt.Sprintf("unknown targetReplicaType %d", t))
	}
}

func (t targetReplicaType) RemoveChangeType() roachpb.ReplicaChangeType {
	__antithesis_instrumentation__.Notify(94137)
	switch t {
	case voterTarget:
		__antithesis_instrumentation__.Notify(94138)
		return roachpb.REMOVE_VOTER
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(94139)
		return roachpb.REMOVE_NON_VOTER
	default:
		__antithesis_instrumentation__.Notify(94140)
		panic(fmt.Sprintf("unknown targetReplicaType %d", t))
	}
}

func (t targetReplicaType) String() string {
	__antithesis_instrumentation__.Notify(94141)
	switch t {
	case voterTarget:
		__antithesis_instrumentation__.Notify(94142)
		return "voter"
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(94143)
		return "non-voter"
	default:
		__antithesis_instrumentation__.Notify(94144)
		panic(fmt.Sprintf("unknown targetReplicaType %d", t))
	}
}

type transferDecision int

const (
	_ transferDecision = iota
	shouldTransfer
	shouldNotTransfer
	decideWithoutStats
)

type allocatorError struct {
	constraints           []roachpb.ConstraintsConjunction
	voterConstraints      []roachpb.ConstraintsConjunction
	existingVoterCount    int
	existingNonVoterCount int
	aliveStores           int
	throttledStores       int
}

func (ae *allocatorError) Error() string {
	__antithesis_instrumentation__.Notify(94145)
	var existingVoterStr string
	if ae.existingVoterCount == 1 {
		__antithesis_instrumentation__.Notify(94152)
		existingVoterStr = "1 already has a voter"
	} else {
		__antithesis_instrumentation__.Notify(94153)
		existingVoterStr = fmt.Sprintf("%d already have a voter", ae.existingVoterCount)
	}
	__antithesis_instrumentation__.Notify(94146)

	var existingNonVoterStr string
	if ae.existingNonVoterCount == 1 {
		__antithesis_instrumentation__.Notify(94154)
		existingNonVoterStr = "1 already has a non-voter"
	} else {
		__antithesis_instrumentation__.Notify(94155)
		existingNonVoterStr = fmt.Sprintf("%d already have a non-voter", ae.existingNonVoterCount)
	}
	__antithesis_instrumentation__.Notify(94147)

	var baseMsg string
	if ae.throttledStores != 0 {
		__antithesis_instrumentation__.Notify(94156)
		baseMsg = fmt.Sprintf(
			"0 of %d live stores are able to take a new replica for the range (%d throttled, %s, %s)",
			ae.aliveStores, ae.throttledStores, existingVoterStr, existingNonVoterStr)
	} else {
		__antithesis_instrumentation__.Notify(94157)
		baseMsg = fmt.Sprintf(
			"0 of %d live stores are able to take a new replica for the range (%s, %s)",
			ae.aliveStores, existingVoterStr, existingNonVoterStr)
	}
	__antithesis_instrumentation__.Notify(94148)

	if len(ae.constraints) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(94158)
		return len(ae.voterConstraints) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94159)
		if ae.throttledStores > 0 {
			__antithesis_instrumentation__.Notify(94161)
			return baseMsg
		} else {
			__antithesis_instrumentation__.Notify(94162)
		}
		__antithesis_instrumentation__.Notify(94160)
		return baseMsg + "; likely not enough nodes in cluster"
	} else {
		__antithesis_instrumentation__.Notify(94163)
	}
	__antithesis_instrumentation__.Notify(94149)

	var b strings.Builder
	b.WriteString(baseMsg)
	b.WriteString("; replicas must match constraints [")
	for i := range ae.constraints {
		__antithesis_instrumentation__.Notify(94164)
		if i > 0 {
			__antithesis_instrumentation__.Notify(94166)
			b.WriteByte(' ')
		} else {
			__antithesis_instrumentation__.Notify(94167)
		}
		__antithesis_instrumentation__.Notify(94165)
		b.WriteByte('{')
		b.WriteString(ae.constraints[i].String())
		b.WriteByte('}')
	}
	__antithesis_instrumentation__.Notify(94150)
	b.WriteString("]")

	b.WriteString("; voting replicas must match voter_constraints [")
	for i := range ae.voterConstraints {
		__antithesis_instrumentation__.Notify(94168)
		if i > 0 {
			__antithesis_instrumentation__.Notify(94170)
			b.WriteByte(' ')
		} else {
			__antithesis_instrumentation__.Notify(94171)
		}
		__antithesis_instrumentation__.Notify(94169)
		b.WriteByte('{')
		b.WriteString(ae.voterConstraints[i].String())
		b.WriteByte('}')
	}
	__antithesis_instrumentation__.Notify(94151)
	b.WriteString("]")

	return b.String()
}

func (*allocatorError) purgatoryErrorMarker() { __antithesis_instrumentation__.Notify(94172) }

var _ purgatoryError = &allocatorError{}

type allocatorRand struct {
	*syncutil.Mutex
	*rand.Rand
}

func makeAllocatorRand(source rand.Source) allocatorRand {
	__antithesis_instrumentation__.Notify(94173)
	return allocatorRand{
		Mutex: &syncutil.Mutex{},
		Rand:  rand.New(source),
	}
}

type RangeUsageInfo struct {
	LogicalBytes     int64
	QueriesPerSecond float64
	WritesPerSecond  float64
}

func rangeUsageInfoForRepl(repl *Replica) RangeUsageInfo {
	__antithesis_instrumentation__.Notify(94174)
	info := RangeUsageInfo{
		LogicalBytes: repl.GetMVCCStats().Total(),
	}
	if queriesPerSecond, dur := repl.leaseholderStats.avgQPS(); dur >= MinStatsDuration {
		__antithesis_instrumentation__.Notify(94177)
		info.QueriesPerSecond = queriesPerSecond
	} else {
		__antithesis_instrumentation__.Notify(94178)
	}
	__antithesis_instrumentation__.Notify(94175)
	if writesPerSecond, dur := repl.writeStats.avgQPS(); dur >= MinStatsDuration {
		__antithesis_instrumentation__.Notify(94179)
		info.WritesPerSecond = writesPerSecond
	} else {
		__antithesis_instrumentation__.Notify(94180)
	}
	__antithesis_instrumentation__.Notify(94176)
	return info
}

var (
	metaLBLeaseTransferCannotFindBetterCandidate = metric.Metadata{
		Name: "kv.allocator.load_based_lease_transfers.cannot_find_better_candidate",
		Help: "The number times the allocator determined that the lease was on the best" +
			" possible replica",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBLeaseTransferExistingNotOverfull = metric.Metadata{
		Name: "kv.allocator.load_based_lease_transfers.existing_not_overfull",
		Help: "The number times the allocator determined that the lease was not on an" +
			" overfull store",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBLeaseTransferDeltaNotSignificant = metric.Metadata{
		Name: "kv.allocator.load_based_lease_transfers.delta_not_significant",
		Help: "The number times the allocator determined that the delta between the existing" +
			" store and the best candidate was not significant",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBLeaseTransferMissingStatsForExistingStore = metric.Metadata{
		Name:        "kv.allocator.load_based_lease_transfers.missing_stats_for_existing_stores",
		Help:        "The number times the allocator was missing qps stats for the leaseholder",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBLeaseTransferSignificantlySwitchesRelativeDisposition = metric.Metadata{
		Name: "kv.allocator.load_based_lease_transfers.significantly_switches_relative_disposition",
		Help: "The number times the allocator decided to not transfer the lease because" +
			" it would invert the dispositions of the sending and the receiving stores",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBLeaseTransferShouldTransfer = metric.Metadata{
		Name: "kv.allocator.load_based_lease_transfers.should_transfer",
		Help: "The number times the allocator determined that the lease should be" +
			" transferred to another replica for better load distribution",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}

	metaLBReplicaRebalancingCannotFindBetterCandidate = metric.Metadata{
		Name: "kv.allocator.load_based_replica_rebalancing.cannot_find_better_candidate",
		Help: "The number times the allocator determined that the range was on the best" +
			" possible stores",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBReplicaRebalancingExistingNotOverfull = metric.Metadata{
		Name: "kv.allocator.load_based_replica_rebalancing.existing_not_overfull",
		Help: "The number times the allocator determined that none of the range's replicas" +
			" were on overfull stores",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBReplicaRebalancingDeltaNotSignificant = metric.Metadata{
		Name: "kv.allocator.load_based_replica_rebalancing.delta_not_significant",
		Help: "The number times the allocator determined that the delta between an" +
			" existing store and the best replacement candidate was not high enough",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBReplicaRebalancingMissingStatsForExistingStore = metric.Metadata{
		Name:        "kv.allocator.load_based_replica_rebalancing.missing_stats_for_existing_store",
		Help:        "The number times the allocator was missing the qps stats for the existing store",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBReplicaRebalancingSignificantlySwitchesRelativeDisposition = metric.Metadata{
		Name: "kv.allocator.load_based_replica_rebalancing.significantly_switches_relative_disposition",
		Help: "The number times the allocator decided to not rebalance the replica" +
			" because it would invert the dispositions of the sending and the receiving stores",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
	metaLBReplicaRebalancingShouldTransfer = metric.Metadata{
		Name: "kv.allocator.load_based_replica_rebalancing.should_transfer",
		Help: "The number times the allocator determined that the replica should be" +
			" rebalanced to another store for better load distribution",
		Measurement: "Attempts",
		Unit:        metric.Unit_COUNT,
	}
)

type loadBasedLeaseTransferMetrics struct {
	CannotFindBetterCandidate                *metric.Counter
	ExistingNotOverfull                      *metric.Counter
	DeltaNotSignificant                      *metric.Counter
	MissingStatsForExistingStore             *metric.Counter
	SignificantlySwitchesRelativeDisposition *metric.Counter
	ShouldTransfer                           *metric.Counter
}

type loadBasedReplicaRebalanceMetrics struct {
	CannotFindBetterCandidate                *metric.Counter
	ExistingNotOverfull                      *metric.Counter
	DeltaNotSignificant                      *metric.Counter
	MissingStatsForExistingStore             *metric.Counter
	SignificantlySwitchesRelativeDisposition *metric.Counter
	ShouldRebalance                          *metric.Counter
}

type AllocatorMetrics struct {
	loadBasedLeaseTransferMetrics
	loadBasedReplicaRebalanceMetrics
}

type Allocator struct {
	storePool     *StorePool
	nodeLatencyFn func(addr string) (time.Duration, bool)

	randGen allocatorRand
	metrics AllocatorMetrics

	knobs *AllocatorTestingKnobs
}

func makeAllocatorMetrics() AllocatorMetrics {
	__antithesis_instrumentation__.Notify(94181)
	return AllocatorMetrics{
		loadBasedLeaseTransferMetrics: loadBasedLeaseTransferMetrics{
			CannotFindBetterCandidate:                metric.NewCounter(metaLBLeaseTransferCannotFindBetterCandidate),
			ExistingNotOverfull:                      metric.NewCounter(metaLBLeaseTransferExistingNotOverfull),
			DeltaNotSignificant:                      metric.NewCounter(metaLBLeaseTransferDeltaNotSignificant),
			MissingStatsForExistingStore:             metric.NewCounter(metaLBLeaseTransferMissingStatsForExistingStore),
			SignificantlySwitchesRelativeDisposition: metric.NewCounter(metaLBLeaseTransferSignificantlySwitchesRelativeDisposition),
			ShouldTransfer:                           metric.NewCounter(metaLBLeaseTransferShouldTransfer),
		},
		loadBasedReplicaRebalanceMetrics: loadBasedReplicaRebalanceMetrics{
			CannotFindBetterCandidate:                metric.NewCounter(metaLBReplicaRebalancingCannotFindBetterCandidate),
			ExistingNotOverfull:                      metric.NewCounter(metaLBReplicaRebalancingExistingNotOverfull),
			DeltaNotSignificant:                      metric.NewCounter(metaLBReplicaRebalancingDeltaNotSignificant),
			MissingStatsForExistingStore:             metric.NewCounter(metaLBReplicaRebalancingMissingStatsForExistingStore),
			SignificantlySwitchesRelativeDisposition: metric.NewCounter(metaLBReplicaRebalancingSignificantlySwitchesRelativeDisposition),
			ShouldRebalance:                          metric.NewCounter(metaLBReplicaRebalancingShouldTransfer),
		},
	}
}

func MakeAllocator(
	storePool *StorePool,
	nodeLatencyFn func(addr string) (time.Duration, bool),
	knobs *AllocatorTestingKnobs,
	storeMetrics *StoreMetrics,
) Allocator {
	__antithesis_instrumentation__.Notify(94182)
	var randSource rand.Source

	if storePool != nil && func() bool {
		__antithesis_instrumentation__.Notify(94185)
		return storePool.deterministic == true
	}() == true {
		__antithesis_instrumentation__.Notify(94186)
		randSource = rand.NewSource(777)
	} else {
		__antithesis_instrumentation__.Notify(94187)
		randSource = rand.NewSource(rand.Int63())
	}
	__antithesis_instrumentation__.Notify(94183)
	allocator := Allocator{
		storePool:     storePool,
		nodeLatencyFn: nodeLatencyFn,
		randGen:       makeAllocatorRand(randSource),
		metrics:       makeAllocatorMetrics(),
		knobs:         knobs,
	}
	if storeMetrics != nil {
		__antithesis_instrumentation__.Notify(94188)
		storeMetrics.registry.AddMetricStruct(allocator.metrics.loadBasedLeaseTransferMetrics)
		storeMetrics.registry.AddMetricStruct(allocator.metrics.loadBasedReplicaRebalanceMetrics)
	} else {
		__antithesis_instrumentation__.Notify(94189)
	}
	__antithesis_instrumentation__.Notify(94184)
	return allocator
}

func GetNeededVoters(zoneConfigVoterCount int32, clusterNodes int) int {
	__antithesis_instrumentation__.Notify(94190)
	numZoneReplicas := int(zoneConfigVoterCount)
	need := numZoneReplicas

	if clusterNodes < need {
		__antithesis_instrumentation__.Notify(94196)
		need = clusterNodes
	} else {
		__antithesis_instrumentation__.Notify(94197)
	}
	__antithesis_instrumentation__.Notify(94191)

	if need == numZoneReplicas {
		__antithesis_instrumentation__.Notify(94198)
		return need
	} else {
		__antithesis_instrumentation__.Notify(94199)
	}
	__antithesis_instrumentation__.Notify(94192)
	if need%2 == 0 {
		__antithesis_instrumentation__.Notify(94200)
		need = need - 1
	} else {
		__antithesis_instrumentation__.Notify(94201)
	}
	__antithesis_instrumentation__.Notify(94193)
	if need < 3 {
		__antithesis_instrumentation__.Notify(94202)
		need = 3
	} else {
		__antithesis_instrumentation__.Notify(94203)
	}
	__antithesis_instrumentation__.Notify(94194)
	if need > numZoneReplicas {
		__antithesis_instrumentation__.Notify(94204)
		need = numZoneReplicas
	} else {
		__antithesis_instrumentation__.Notify(94205)
	}
	__antithesis_instrumentation__.Notify(94195)

	return need
}

func GetNeededNonVoters(numVoters, zoneConfigNonVoterCount, clusterNodes int) int {
	__antithesis_instrumentation__.Notify(94206)
	need := zoneConfigNonVoterCount
	if clusterNodes-numVoters < need {
		__antithesis_instrumentation__.Notify(94209)

		need = clusterNodes - numVoters
	} else {
		__antithesis_instrumentation__.Notify(94210)
	}
	__antithesis_instrumentation__.Notify(94207)
	if need < 0 {
		__antithesis_instrumentation__.Notify(94211)
		need = 0
	} else {
		__antithesis_instrumentation__.Notify(94212)
	}
	__antithesis_instrumentation__.Notify(94208)
	return need
}

func (a *Allocator) ComputeAction(
	ctx context.Context, conf roachpb.SpanConfig, desc *roachpb.RangeDescriptor,
) (action AllocatorAction, priority float64) {
	__antithesis_instrumentation__.Notify(94213)
	if a.storePool == nil {
		__antithesis_instrumentation__.Notify(94217)

		action = AllocatorNoop
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94218)
	}
	__antithesis_instrumentation__.Notify(94214)

	if desc.Replicas().InAtomicReplicationChange() {
		__antithesis_instrumentation__.Notify(94219)

		action = AllocatorFinalizeAtomicReplicationChange
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94220)
	}
	__antithesis_instrumentation__.Notify(94215)

	if learners := desc.Replicas().LearnerDescriptors(); len(learners) > 0 {
		__antithesis_instrumentation__.Notify(94221)

		action = AllocatorRemoveLearner
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94222)
	}
	__antithesis_instrumentation__.Notify(94216)

	return a.computeAction(ctx, conf, desc.Replicas().VoterDescriptors(),
		desc.Replicas().NonVoterDescriptors())

}

func (a *Allocator) computeAction(
	ctx context.Context,
	conf roachpb.SpanConfig,
	voterReplicas []roachpb.ReplicaDescriptor,
	nonVoterReplicas []roachpb.ReplicaDescriptor,
) (action AllocatorAction, adjustedPriority float64) {
	__antithesis_instrumentation__.Notify(94223)

	haveVoters := len(voterReplicas)
	decommissioningVoters := a.storePool.decommissioningReplicas(voterReplicas)

	clusterNodes := a.storePool.ClusterNodeCount()
	neededVoters := GetNeededVoters(conf.GetNumVoters(), clusterNodes)
	desiredQuorum := computeQuorum(neededVoters)
	quorum := computeQuorum(haveVoters)

	if haveVoters < neededVoters {
		__antithesis_instrumentation__.Notify(94237)

		action = AllocatorAddVoter
		adjustedPriority = action.Priority() + float64(desiredQuorum-haveVoters)
		log.VEventf(ctx, 3, "%s - missing voter need=%d, have=%d, priority=%.2f",
			action, neededVoters, haveVoters, adjustedPriority)
		return action, adjustedPriority
	} else {
		__antithesis_instrumentation__.Notify(94238)
	}
	__antithesis_instrumentation__.Notify(94224)

	const includeSuspectAndDrainingStores = true
	liveVoters, deadVoters := a.storePool.liveAndDeadReplicas(voterReplicas, includeSuspectAndDrainingStores)

	if len(liveVoters) < quorum {
		__antithesis_instrumentation__.Notify(94239)

		action = AllocatorRangeUnavailable
		log.VEventf(ctx, 1, "unable to take action - live voters %v don't meet quorum of %d",
			liveVoters, quorum)
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94240)
	}
	__antithesis_instrumentation__.Notify(94225)

	if haveVoters == neededVoters && func() bool {
		__antithesis_instrumentation__.Notify(94241)
		return len(deadVoters) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94242)

		action = AllocatorReplaceDeadVoter
		log.VEventf(ctx, 3, "%s - replacement for %d dead voters priority=%.2f",
			action, len(deadVoters), action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94243)
	}
	__antithesis_instrumentation__.Notify(94226)

	if haveVoters == neededVoters && func() bool {
		__antithesis_instrumentation__.Notify(94244)
		return len(decommissioningVoters) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94245)

		action = AllocatorReplaceDecommissioningVoter
		log.VEventf(ctx, 3, "%s - replacement for %d decommissioning voters priority=%.2f",
			action, len(decommissioningVoters), action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94246)
	}
	__antithesis_instrumentation__.Notify(94227)

	if len(deadVoters) > 0 {
		__antithesis_instrumentation__.Notify(94247)

		action = AllocatorRemoveDeadVoter
		adjustedPriority = action.Priority() + float64(quorum-len(liveVoters))
		log.VEventf(ctx, 3, "%s - dead=%d, live=%d, quorum=%d, priority=%.2f",
			action, len(deadVoters), len(liveVoters), quorum, adjustedPriority)
		return action, adjustedPriority
	} else {
		__antithesis_instrumentation__.Notify(94248)
	}
	__antithesis_instrumentation__.Notify(94228)

	if len(decommissioningVoters) > 0 {
		__antithesis_instrumentation__.Notify(94249)

		action = AllocatorRemoveDecommissioningVoter
		log.VEventf(ctx, 3,
			"%s - need=%d, have=%d, num_decommissioning=%d, priority=%.2f",
			action, neededVoters, haveVoters, len(decommissioningVoters), action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94250)
	}
	__antithesis_instrumentation__.Notify(94229)

	if haveVoters > neededVoters {
		__antithesis_instrumentation__.Notify(94251)

		action = AllocatorRemoveVoter
		adjustedPriority = action.Priority() - float64(haveVoters%2)
		log.VEventf(ctx, 3, "%s - need=%d, have=%d, priority=%.2f", action, neededVoters,
			haveVoters, adjustedPriority)
		return action, adjustedPriority
	} else {
		__antithesis_instrumentation__.Notify(94252)
	}
	__antithesis_instrumentation__.Notify(94230)

	haveNonVoters := len(nonVoterReplicas)
	neededNonVoters := GetNeededNonVoters(haveVoters, int(conf.GetNumNonVoters()), clusterNodes)
	if haveNonVoters < neededNonVoters {
		__antithesis_instrumentation__.Notify(94253)
		action = AllocatorAddNonVoter
		log.VEventf(ctx, 3, "%s - missing non-voter need=%d, have=%d, priority=%.2f",
			action, neededNonVoters, haveNonVoters, action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94254)
	}
	__antithesis_instrumentation__.Notify(94231)

	liveNonVoters, deadNonVoters := a.storePool.liveAndDeadReplicas(
		nonVoterReplicas, includeSuspectAndDrainingStores,
	)
	if haveNonVoters == neededNonVoters && func() bool {
		__antithesis_instrumentation__.Notify(94255)
		return len(deadNonVoters) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94256)

		action = AllocatorReplaceDeadNonVoter
		log.VEventf(ctx, 3, "%s - replacement for %d dead non-voters priority=%.2f",
			action, len(deadNonVoters), action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94257)
	}
	__antithesis_instrumentation__.Notify(94232)

	decommissioningNonVoters := a.storePool.decommissioningReplicas(nonVoterReplicas)
	if haveNonVoters == neededNonVoters && func() bool {
		__antithesis_instrumentation__.Notify(94258)
		return len(decommissioningNonVoters) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94259)

		action = AllocatorReplaceDecommissioningNonVoter
		log.VEventf(ctx, 3, "%s - replacement for %d decommissioning non-voters priority=%.2f",
			action, len(decommissioningNonVoters), action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94260)
	}
	__antithesis_instrumentation__.Notify(94233)

	if len(deadNonVoters) > 0 {
		__antithesis_instrumentation__.Notify(94261)

		action = AllocatorRemoveDeadNonVoter
		log.VEventf(ctx, 3, "%s - dead=%d, live=%d, priority=%.2f",
			action, len(deadNonVoters), len(liveNonVoters), action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94262)
	}
	__antithesis_instrumentation__.Notify(94234)

	if len(decommissioningNonVoters) > 0 {
		__antithesis_instrumentation__.Notify(94263)

		action = AllocatorRemoveDecommissioningNonVoter
		log.VEventf(ctx, 3,
			"%s - need=%d, have=%d, num_decommissioning=%d, priority=%.2f",
			action, neededNonVoters, haveNonVoters, len(decommissioningNonVoters), action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94264)
	}
	__antithesis_instrumentation__.Notify(94235)

	if haveNonVoters > neededNonVoters {
		__antithesis_instrumentation__.Notify(94265)

		action = AllocatorRemoveNonVoter
		log.VEventf(ctx, 3, "%s - need=%d, have=%d, priority=%.2f", action,
			neededNonVoters, haveNonVoters, action.Priority())
		return action, action.Priority()
	} else {
		__antithesis_instrumentation__.Notify(94266)
	}
	__antithesis_instrumentation__.Notify(94236)

	action = AllocatorConsiderRebalance
	return action, action.Priority()
}

func getReplicasForDiversityCalc(
	targetType targetReplicaType, existingVoters, allExistingReplicas []roachpb.ReplicaDescriptor,
) []roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(94267)
	switch t := targetType; t {
	case voterTarget:
		__antithesis_instrumentation__.Notify(94268)

		return existingVoters
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(94269)
		return allExistingReplicas
	default:
		__antithesis_instrumentation__.Notify(94270)
		panic(fmt.Sprintf("unsupported targetReplicaType: %v", t))
	}
}

type decisionDetails struct {
	Target   string
	Existing string `json:",omitempty"`
}

func (a *Allocator) allocateTarget(
	ctx context.Context,
	conf roachpb.SpanConfig,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
	targetType targetReplicaType,
) (roachpb.ReplicationTarget, string, error) {
	__antithesis_instrumentation__.Notify(94271)
	candidateStoreList, aliveStoreCount, throttled := a.storePool.getStoreList(storeFilterThrottled)

	target, details := a.allocateTargetFromList(
		ctx,
		candidateStoreList,
		conf,
		existingVoters,
		existingNonVoters,
		a.scorerOptions(),

		false,
		targetType,
	)

	if !roachpb.Empty(target) {
		__antithesis_instrumentation__.Notify(94274)
		return target, details, nil
	} else {
		__antithesis_instrumentation__.Notify(94275)
	}
	__antithesis_instrumentation__.Notify(94272)

	if len(throttled) > 0 {
		__antithesis_instrumentation__.Notify(94276)
		return roachpb.ReplicationTarget{}, "", errors.Errorf(
			"%d matching stores are currently throttled: %v", len(throttled), throttled,
		)
	} else {
		__antithesis_instrumentation__.Notify(94277)
	}
	__antithesis_instrumentation__.Notify(94273)
	return roachpb.ReplicationTarget{}, "", &allocatorError{
		voterConstraints:      conf.VoterConstraints,
		constraints:           conf.Constraints,
		existingVoterCount:    len(existingVoters),
		existingNonVoterCount: len(existingNonVoters),
		aliveStores:           aliveStoreCount,
		throttledStores:       len(throttled),
	}
}

func (a *Allocator) AllocateVoter(
	ctx context.Context,
	conf roachpb.SpanConfig,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
) (roachpb.ReplicationTarget, string, error) {
	__antithesis_instrumentation__.Notify(94278)
	return a.allocateTarget(ctx, conf, existingVoters, existingNonVoters, voterTarget)
}

func (a *Allocator) AllocateNonVoter(
	ctx context.Context,
	conf roachpb.SpanConfig,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
) (roachpb.ReplicationTarget, string, error) {
	__antithesis_instrumentation__.Notify(94279)
	return a.allocateTarget(ctx, conf, existingVoters, existingNonVoters, nonVoterTarget)
}

func (a *Allocator) allocateTargetFromList(
	ctx context.Context,
	candidateStores StoreList,
	conf roachpb.SpanConfig,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
	options *rangeCountScorerOptions,
	allowMultipleReplsPerNode bool,
	targetType targetReplicaType,
) (roachpb.ReplicationTarget, string) {
	__antithesis_instrumentation__.Notify(94280)
	existingReplicas := append(existingVoters, existingNonVoters...)
	analyzedOverallConstraints := constraint.AnalyzeConstraints(ctx, a.storePool.getStoreDescriptor,
		existingReplicas, conf.NumReplicas, conf.Constraints)
	analyzedVoterConstraints := constraint.AnalyzeConstraints(ctx, a.storePool.getStoreDescriptor,
		existingVoters, conf.GetNumVoters(), conf.VoterConstraints)

	var constraintsChecker constraintsCheckFn
	switch t := targetType; t {
	case voterTarget:
		__antithesis_instrumentation__.Notify(94283)
		constraintsChecker = voterConstraintsCheckerForAllocation(
			analyzedOverallConstraints,
			analyzedVoterConstraints,
		)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(94284)
		constraintsChecker = nonVoterConstraintsCheckerForAllocation(analyzedOverallConstraints)
	default:
		__antithesis_instrumentation__.Notify(94285)
		log.Fatalf(ctx, "unsupported targetReplicaType: %v", t)
	}
	__antithesis_instrumentation__.Notify(94281)

	existingReplicaSet := getReplicasForDiversityCalc(targetType, existingVoters, existingReplicas)
	candidates := rankedCandidateListForAllocation(
		ctx,
		candidateStores,
		constraintsChecker,
		existingReplicaSet,
		a.storePool.getLocalitiesByStore(existingReplicaSet),
		a.storePool.isStoreReadyForRoutineReplicaTransfer,
		allowMultipleReplsPerNode,
		options,
	)

	log.VEventf(ctx, 3, "allocate %s: %s", targetType, candidates)
	if target := candidates.selectGood(a.randGen); target != nil {
		__antithesis_instrumentation__.Notify(94286)
		log.VEventf(ctx, 3, "add target: %s", target)
		details := decisionDetails{Target: target.compactString()}
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			__antithesis_instrumentation__.Notify(94288)
			log.Warningf(ctx, "failed to marshal details for choosing allocate target: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(94289)
		}
		__antithesis_instrumentation__.Notify(94287)
		return roachpb.ReplicationTarget{
			NodeID: target.store.Node.NodeID, StoreID: target.store.StoreID,
		}, string(detailsBytes)
	} else {
		__antithesis_instrumentation__.Notify(94290)
	}
	__antithesis_instrumentation__.Notify(94282)

	return roachpb.ReplicationTarget{}, ""
}

func (a Allocator) simulateRemoveTarget(
	ctx context.Context,
	targetStore roachpb.StoreID,
	conf roachpb.SpanConfig,
	candidates []roachpb.ReplicaDescriptor,
	existingVoters []roachpb.ReplicaDescriptor,
	existingNonVoters []roachpb.ReplicaDescriptor,
	sl StoreList,
	rangeUsageInfo RangeUsageInfo,
	targetType targetReplicaType,
	options scorerOptions,
) (roachpb.ReplicationTarget, string, error) {
	__antithesis_instrumentation__.Notify(94291)
	candidateStores := make([]roachpb.StoreDescriptor, 0, len(candidates))
	for _, cand := range candidates {
		__antithesis_instrumentation__.Notify(94293)
		for _, store := range sl.stores {
			__antithesis_instrumentation__.Notify(94294)
			if cand.StoreID == store.StoreID {
				__antithesis_instrumentation__.Notify(94295)
				candidateStores = append(candidateStores, store)
			} else {
				__antithesis_instrumentation__.Notify(94296)
			}
		}
	}
	__antithesis_instrumentation__.Notify(94292)

	switch t := targetType; t {
	case voterTarget:
		__antithesis_instrumentation__.Notify(94297)
		a.storePool.updateLocalStoreAfterRebalance(targetStore, rangeUsageInfo, roachpb.ADD_VOTER)
		defer a.storePool.updateLocalStoreAfterRebalance(
			targetStore,
			rangeUsageInfo,
			roachpb.REMOVE_VOTER,
		)
		log.VEventf(ctx, 3, "simulating which voter would be removed after adding s%d",
			targetStore)

		return a.removeTarget(
			ctx, conf, makeStoreList(candidateStores),
			existingVoters, existingNonVoters, voterTarget, options,
		)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(94298)
		a.storePool.updateLocalStoreAfterRebalance(targetStore, rangeUsageInfo, roachpb.ADD_NON_VOTER)
		defer a.storePool.updateLocalStoreAfterRebalance(
			targetStore,
			rangeUsageInfo,
			roachpb.REMOVE_NON_VOTER,
		)
		log.VEventf(ctx, 3, "simulating which non-voter would be removed after adding s%d",
			targetStore)
		return a.removeTarget(
			ctx, conf, makeStoreList(candidateStores),
			existingVoters, existingNonVoters, nonVoterTarget, options,
		)
	default:
		__antithesis_instrumentation__.Notify(94299)
		panic(fmt.Sprintf("unknown targetReplicaType: %s", t))
	}
}

func (a Allocator) storeListForTargets(candidates []roachpb.ReplicationTarget) StoreList {
	__antithesis_instrumentation__.Notify(94300)
	result := make([]roachpb.StoreDescriptor, 0, len(candidates))
	sl, _, _ := a.storePool.getStoreList(storeFilterNone)
	for _, cand := range candidates {
		__antithesis_instrumentation__.Notify(94302)
		for _, store := range sl.stores {
			__antithesis_instrumentation__.Notify(94303)
			if cand.StoreID == store.StoreID {
				__antithesis_instrumentation__.Notify(94304)
				result = append(result, store)
			} else {
				__antithesis_instrumentation__.Notify(94305)
			}
		}
	}
	__antithesis_instrumentation__.Notify(94301)
	return makeStoreList(result)
}

func (a Allocator) removeTarget(
	ctx context.Context,
	conf roachpb.SpanConfig,
	candidateStoreList StoreList,
	existingVoters []roachpb.ReplicaDescriptor,
	existingNonVoters []roachpb.ReplicaDescriptor,
	targetType targetReplicaType,
	options scorerOptions,
) (roachpb.ReplicationTarget, string, error) {
	__antithesis_instrumentation__.Notify(94306)
	if len(candidateStoreList.stores) == 0 {
		__antithesis_instrumentation__.Notify(94310)
		return roachpb.ReplicationTarget{}, "", errors.Errorf(
			"must supply at least one" +
				" candidate replica to allocator.removeTarget()",
		)
	} else {
		__antithesis_instrumentation__.Notify(94311)
	}
	__antithesis_instrumentation__.Notify(94307)

	existingReplicas := append(existingVoters, existingNonVoters...)
	analyzedOverallConstraints := constraint.AnalyzeConstraints(ctx, a.storePool.getStoreDescriptor,
		existingReplicas, conf.NumReplicas, conf.Constraints)
	analyzedVoterConstraints := constraint.AnalyzeConstraints(ctx, a.storePool.getStoreDescriptor,
		existingVoters, conf.GetNumVoters(), conf.VoterConstraints)

	var constraintsChecker constraintsCheckFn
	switch t := targetType; t {
	case voterTarget:
		__antithesis_instrumentation__.Notify(94312)

		constraintsChecker = voterConstraintsCheckerForRemoval(
			analyzedOverallConstraints,
			analyzedVoterConstraints,
		)
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(94313)
		constraintsChecker = nonVoterConstraintsCheckerForRemoval(analyzedOverallConstraints)
	default:
		__antithesis_instrumentation__.Notify(94314)
		log.Fatalf(ctx, "unsupported targetReplicaType: %v", t)
	}
	__antithesis_instrumentation__.Notify(94308)

	replicaSetForDiversityCalc := getReplicasForDiversityCalc(targetType, existingVoters, existingReplicas)
	rankedCandidates := candidateListForRemoval(
		candidateStoreList,
		constraintsChecker,
		a.storePool.getLocalitiesByStore(replicaSetForDiversityCalc),
		options,
	)

	log.VEventf(ctx, 3, "remove %s: %s", targetType, rankedCandidates)
	if bad := rankedCandidates.selectBad(a.randGen); bad != nil {
		__antithesis_instrumentation__.Notify(94315)
		for _, exist := range existingReplicas {
			__antithesis_instrumentation__.Notify(94316)
			if exist.StoreID == bad.store.StoreID {
				__antithesis_instrumentation__.Notify(94317)
				log.VEventf(ctx, 3, "remove target: %s", bad)
				details := decisionDetails{Target: bad.compactString()}
				detailsBytes, err := json.Marshal(details)
				if err != nil {
					__antithesis_instrumentation__.Notify(94319)
					log.Warningf(ctx, "failed to marshal details for choosing remove target: %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(94320)
				}
				__antithesis_instrumentation__.Notify(94318)
				return roachpb.ReplicationTarget{
					StoreID: exist.StoreID, NodeID: exist.NodeID,
				}, string(detailsBytes), nil
			} else {
				__antithesis_instrumentation__.Notify(94321)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(94322)
	}
	__antithesis_instrumentation__.Notify(94309)

	return roachpb.ReplicationTarget{}, "", errors.New("could not select an appropriate replica to be removed")
}

func (a Allocator) RemoveVoter(
	ctx context.Context,
	conf roachpb.SpanConfig,
	voterCandidates []roachpb.ReplicaDescriptor,
	existingVoters []roachpb.ReplicaDescriptor,
	existingNonVoters []roachpb.ReplicaDescriptor,
	options scorerOptions,
) (roachpb.ReplicationTarget, string, error) {
	__antithesis_instrumentation__.Notify(94323)

	candidateStoreIDs := make(roachpb.StoreIDSlice, len(voterCandidates))
	for i, exist := range voterCandidates {
		__antithesis_instrumentation__.Notify(94325)
		candidateStoreIDs[i] = exist.StoreID
	}
	__antithesis_instrumentation__.Notify(94324)
	candidateStoreList, _, _ := a.storePool.getStoreListFromIDs(candidateStoreIDs, storeFilterNone)

	return a.removeTarget(
		ctx,
		conf,
		candidateStoreList,
		existingVoters,
		existingNonVoters,
		voterTarget,
		options,
	)
}

func (a Allocator) RemoveNonVoter(
	ctx context.Context,
	conf roachpb.SpanConfig,
	nonVoterCandidates []roachpb.ReplicaDescriptor,
	existingVoters []roachpb.ReplicaDescriptor,
	existingNonVoters []roachpb.ReplicaDescriptor,
	options scorerOptions,
) (roachpb.ReplicationTarget, string, error) {
	__antithesis_instrumentation__.Notify(94326)

	candidateStoreIDs := make(roachpb.StoreIDSlice, len(nonVoterCandidates))
	for i, exist := range nonVoterCandidates {
		__antithesis_instrumentation__.Notify(94328)
		candidateStoreIDs[i] = exist.StoreID
	}
	__antithesis_instrumentation__.Notify(94327)
	candidateStoreList, _, _ := a.storePool.getStoreListFromIDs(candidateStoreIDs, storeFilterNone)

	return a.removeTarget(
		ctx,
		conf,
		candidateStoreList,
		existingVoters,
		existingNonVoters,
		nonVoterTarget,
		options,
	)
}

func (a Allocator) rebalanceTarget(
	ctx context.Context,
	conf roachpb.SpanConfig,
	raftStatus *raft.Status,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
	rangeUsageInfo RangeUsageInfo,
	filter storeFilter,
	targetType targetReplicaType,
	options scorerOptions,
) (add, remove roachpb.ReplicationTarget, details string, ok bool) {
	__antithesis_instrumentation__.Notify(94329)
	sl, _, _ := a.storePool.getStoreList(filter)

	sl = options.maybeJitterStoreStats(sl, a.randGen)

	existingReplicas := append(existingVoters, existingNonVoters...)

	zero := roachpb.ReplicationTarget{}
	analyzedOverallConstraints := constraint.AnalyzeConstraints(
		ctx, a.storePool.getStoreDescriptor, existingReplicas, conf.NumReplicas, conf.Constraints)
	analyzedVoterConstraints := constraint.AnalyzeConstraints(
		ctx, a.storePool.getStoreDescriptor, existingVoters, conf.GetNumVoters(), conf.VoterConstraints)
	var removalConstraintsChecker constraintsCheckFn
	var rebalanceConstraintsChecker rebalanceConstraintsCheckFn
	var replicaSetToRebalance, replicasWithExcludedStores []roachpb.ReplicaDescriptor
	var otherReplicaSet []roachpb.ReplicaDescriptor

	switch t := targetType; t {
	case voterTarget:
		__antithesis_instrumentation__.Notify(94334)
		removalConstraintsChecker = voterConstraintsCheckerForRemoval(
			analyzedOverallConstraints,
			analyzedVoterConstraints,
		)
		rebalanceConstraintsChecker = voterConstraintsCheckerForRebalance(
			analyzedOverallConstraints,
			analyzedVoterConstraints,
		)
		replicaSetToRebalance = existingVoters
		otherReplicaSet = existingNonVoters
	case nonVoterTarget:
		__antithesis_instrumentation__.Notify(94335)
		removalConstraintsChecker = nonVoterConstraintsCheckerForRemoval(analyzedOverallConstraints)
		rebalanceConstraintsChecker = nonVoterConstraintsCheckerForRebalance(analyzedOverallConstraints)
		replicaSetToRebalance = existingNonVoters

		replicasWithExcludedStores = existingVoters
		otherReplicaSet = existingVoters
	default:
		__antithesis_instrumentation__.Notify(94336)
		log.Fatalf(ctx, "unsupported targetReplicaType: %v", t)
	}
	__antithesis_instrumentation__.Notify(94330)

	replicaSetForDiversityCalc := getReplicasForDiversityCalc(targetType, existingVoters, existingReplicas)
	results := rankedCandidateListForRebalancing(
		ctx,
		sl,
		removalConstraintsChecker,
		rebalanceConstraintsChecker,
		replicaSetToRebalance,
		replicasWithExcludedStores,
		a.storePool.getLocalitiesByStore(replicaSetForDiversityCalc),
		a.storePool.isStoreReadyForRoutineReplicaTransfer,
		options,
		a.metrics,
	)

	if len(results) == 0 {
		__antithesis_instrumentation__.Notify(94337)
		return zero, zero, "", false
	} else {
		__antithesis_instrumentation__.Notify(94338)
	}
	__antithesis_instrumentation__.Notify(94331)

	var target, existingCandidate *candidate
	var removeReplica roachpb.ReplicationTarget
	for {
		__antithesis_instrumentation__.Notify(94339)
		target, existingCandidate = bestRebalanceTarget(a.randGen, results)
		if target == nil {
			__antithesis_instrumentation__.Notify(94345)
			return zero, zero, "", false
		} else {
			__antithesis_instrumentation__.Notify(94346)
		}
		__antithesis_instrumentation__.Notify(94340)

		newReplica := roachpb.ReplicaDescriptor{
			NodeID:    target.store.Node.NodeID,
			StoreID:   target.store.StoreID,
			ReplicaID: maxReplicaID(existingReplicas) + 1,
		}

		existingPlusOneNew := append([]roachpb.ReplicaDescriptor(nil), replicaSetToRebalance...)
		existingPlusOneNew = append(existingPlusOneNew, newReplica)
		replicaCandidates := existingPlusOneNew

		if targetType == voterTarget && func() bool {
			__antithesis_instrumentation__.Notify(94347)
			return raftStatus != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(94348)
			return raftStatus.Progress != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(94349)
			replicaCandidates = simulateFilterUnremovableReplicas(
				ctx, raftStatus, replicaCandidates, newReplica.ReplicaID)
		} else {
			__antithesis_instrumentation__.Notify(94350)
		}
		__antithesis_instrumentation__.Notify(94341)
		if len(replicaCandidates) == 0 {
			__antithesis_instrumentation__.Notify(94351)

			log.VEventf(ctx, 2, "not rebalancing %s to s%d because there are no existing "+
				"replicas that can be removed", targetType, target.store.StoreID)
			return zero, zero, "", false
		} else {
			__antithesis_instrumentation__.Notify(94352)
		}
		__antithesis_instrumentation__.Notify(94342)

		var removeDetails string
		var err error
		removeReplica, removeDetails, err = a.simulateRemoveTarget(
			ctx,
			target.store.StoreID,
			conf,
			replicaCandidates,
			existingPlusOneNew,
			otherReplicaSet,
			sl,
			rangeUsageInfo,
			targetType,
			options,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(94353)
			log.Warningf(ctx, "simulating removal of %s failed: %+v", targetType, err)
			return zero, zero, "", false
		} else {
			__antithesis_instrumentation__.Notify(94354)
		}
		__antithesis_instrumentation__.Notify(94343)
		if target.store.StoreID != removeReplica.StoreID {
			__antithesis_instrumentation__.Notify(94355)

			_, _ = target, removeReplica
			break
		} else {
			__antithesis_instrumentation__.Notify(94356)
		}
		__antithesis_instrumentation__.Notify(94344)

		log.VEventf(ctx, 2, "not rebalancing to s%d because we'd immediately remove it: %s",
			target.store.StoreID, removeDetails)
	}
	__antithesis_instrumentation__.Notify(94332)

	dDetails := decisionDetails{
		Target:   target.compactString(),
		Existing: existingCandidate.compactString(),
	}
	detailsBytes, err := json.Marshal(dDetails)
	if err != nil {
		__antithesis_instrumentation__.Notify(94357)
		log.Warningf(ctx, "failed to marshal details for choosing rebalance target: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(94358)
	}
	__antithesis_instrumentation__.Notify(94333)

	addTarget := roachpb.ReplicationTarget{
		NodeID:  target.store.Node.NodeID,
		StoreID: target.store.StoreID,
	}
	removeTarget := roachpb.ReplicationTarget{
		NodeID:  removeReplica.NodeID,
		StoreID: removeReplica.StoreID,
	}
	return addTarget, removeTarget, string(detailsBytes), true
}

func (a Allocator) RebalanceVoter(
	ctx context.Context,
	conf roachpb.SpanConfig,
	raftStatus *raft.Status,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
	rangeUsageInfo RangeUsageInfo,
	filter storeFilter,
	options scorerOptions,
) (add, remove roachpb.ReplicationTarget, details string, ok bool) {
	__antithesis_instrumentation__.Notify(94359)
	return a.rebalanceTarget(
		ctx,
		conf,
		raftStatus,
		existingVoters,
		existingNonVoters,
		rangeUsageInfo,
		filter,
		voterTarget,
		options,
	)
}

func (a Allocator) RebalanceNonVoter(
	ctx context.Context,
	conf roachpb.SpanConfig,
	raftStatus *raft.Status,
	existingVoters, existingNonVoters []roachpb.ReplicaDescriptor,
	rangeUsageInfo RangeUsageInfo,
	filter storeFilter,
	options scorerOptions,
) (add, remove roachpb.ReplicationTarget, details string, ok bool) {
	__antithesis_instrumentation__.Notify(94360)
	return a.rebalanceTarget(
		ctx,
		conf,
		raftStatus,
		existingVoters,
		existingNonVoters,
		rangeUsageInfo,
		filter,
		nonVoterTarget,
		options,
	)
}

func (a *Allocator) scorerOptions() *rangeCountScorerOptions {
	__antithesis_instrumentation__.Notify(94361)
	return &rangeCountScorerOptions{
		deterministic:           a.storePool.deterministic,
		rangeRebalanceThreshold: rangeRebalanceThreshold.Get(&a.storePool.st.SV),
	}
}

func (a *Allocator) scorerOptionsForScatter() *scatterScorerOptions {
	__antithesis_instrumentation__.Notify(94362)
	return &scatterScorerOptions{
		rangeCountScorerOptions: rangeCountScorerOptions{
			deterministic:           a.storePool.deterministic,
			rangeRebalanceThreshold: 0,
		},

		jitter: rangeRebalanceThreshold.Get(&a.storePool.st.SV),
	}
}

func (a *Allocator) TransferLeaseTarget(
	ctx context.Context,
	conf roachpb.SpanConfig,
	existing []roachpb.ReplicaDescriptor,
	leaseRepl interface {
		RaftStatus() *raft.Status
		StoreID() roachpb.StoreID
		GetRangeID() roachpb.RangeID
	},
	stats *replicaStats,
	forceDecisionWithoutStats bool,
	opts transferLeaseOptions,
) roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(94363)
	allStoresList, _, _ := a.storePool.getStoreList(storeFilterNone)
	storeDescMap := storeListToMap(allStoresList)

	sl, _, _ := a.storePool.getStoreList(storeFilterSuspect)
	sl = sl.excludeInvalid(conf.Constraints)
	sl = sl.excludeInvalid(conf.VoterConstraints)

	candidateLeasesMean := sl.candidateLeases.mean

	source, ok := a.storePool.getStoreDescriptor(leaseRepl.StoreID())
	if !ok {
		__antithesis_instrumentation__.Notify(94370)
		return roachpb.ReplicaDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(94371)
	}
	__antithesis_instrumentation__.Notify(94364)

	var preferred []roachpb.ReplicaDescriptor
	checkTransferLeaseSource := opts.checkTransferLeaseSource
	if checkTransferLeaseSource {
		__antithesis_instrumentation__.Notify(94372)
		preferred = a.preferredLeaseholders(conf, existing)
	} else {
		__antithesis_instrumentation__.Notify(94373)

		var candidates []roachpb.ReplicaDescriptor
		for _, repl := range existing {
			__antithesis_instrumentation__.Notify(94375)
			if repl.StoreID != leaseRepl.StoreID() {
				__antithesis_instrumentation__.Notify(94376)
				candidates = append(candidates, repl)
			} else {
				__antithesis_instrumentation__.Notify(94377)
			}
		}
		__antithesis_instrumentation__.Notify(94374)
		preferred = a.preferredLeaseholders(conf, candidates)
	}
	__antithesis_instrumentation__.Notify(94365)
	if len(preferred) == 1 {
		__antithesis_instrumentation__.Notify(94378)
		if preferred[0].StoreID == leaseRepl.StoreID() {
			__antithesis_instrumentation__.Notify(94381)
			return roachpb.ReplicaDescriptor{}
		} else {
			__antithesis_instrumentation__.Notify(94382)
		}
		__antithesis_instrumentation__.Notify(94379)

		preferred, _ = a.storePool.liveAndDeadReplicas(preferred, false)
		if len(preferred) == 1 {
			__antithesis_instrumentation__.Notify(94383)
			return preferred[0]
		} else {
			__antithesis_instrumentation__.Notify(94384)
		}
		__antithesis_instrumentation__.Notify(94380)
		return roachpb.ReplicaDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(94385)
		if len(preferred) > 1 {
			__antithesis_instrumentation__.Notify(94386)

			existing = preferred
			if !storeHasReplica(leaseRepl.StoreID(), roachpb.MakeReplicaSet(preferred).ReplicationTargets()) {
				__antithesis_instrumentation__.Notify(94387)
				checkTransferLeaseSource = false
			} else {
				__antithesis_instrumentation__.Notify(94388)
			}
		} else {
			__antithesis_instrumentation__.Notify(94389)
		}
	}
	__antithesis_instrumentation__.Notify(94366)

	existing, _ = a.storePool.liveAndDeadReplicas(existing, false)

	if a.knobs == nil || func() bool {
		__antithesis_instrumentation__.Notify(94390)
		return !a.knobs.AllowLeaseTransfersToReplicasNeedingSnapshots == true
	}() == true {
		__antithesis_instrumentation__.Notify(94391)

		existing = excludeReplicasInNeedOfSnapshots(ctx, leaseRepl.RaftStatus(), existing)
	} else {
		__antithesis_instrumentation__.Notify(94392)
	}
	__antithesis_instrumentation__.Notify(94367)

	if len(existing) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(94393)
		return (len(existing) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(94394)
			return existing[0].StoreID == leaseRepl.StoreID() == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(94395)
		log.VEventf(ctx, 2, "no lease transfer target found for r%d", leaseRepl.GetRangeID())
		return roachpb.ReplicaDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(94396)
	}
	__antithesis_instrumentation__.Notify(94368)

	switch g := opts.goal; g {
	case followTheWorkload:
		__antithesis_instrumentation__.Notify(94397)

		transferDec, repl := a.shouldTransferLeaseForAccessLocality(
			ctx, source, existing, stats, nil, candidateLeasesMean,
		)
		if checkTransferLeaseSource {
			__antithesis_instrumentation__.Notify(94408)
			switch transferDec {
			case shouldNotTransfer:
				__antithesis_instrumentation__.Notify(94409)
				if !forceDecisionWithoutStats {
					__antithesis_instrumentation__.Notify(94414)
					return roachpb.ReplicaDescriptor{}
				} else {
					__antithesis_instrumentation__.Notify(94415)
				}
				__antithesis_instrumentation__.Notify(94410)
				fallthrough
			case decideWithoutStats:
				__antithesis_instrumentation__.Notify(94411)
				if !a.shouldTransferLeaseForLeaseCountConvergence(ctx, sl, source, existing) {
					__antithesis_instrumentation__.Notify(94416)
					return roachpb.ReplicaDescriptor{}
				} else {
					__antithesis_instrumentation__.Notify(94417)
				}
			case shouldTransfer:
				__antithesis_instrumentation__.Notify(94412)
			default:
				__antithesis_instrumentation__.Notify(94413)
				log.Fatalf(ctx, "unexpected transfer decision %d with replica %+v", transferDec, repl)
			}
		} else {
			__antithesis_instrumentation__.Notify(94418)
		}
		__antithesis_instrumentation__.Notify(94398)
		if repl != (roachpb.ReplicaDescriptor{}) {
			__antithesis_instrumentation__.Notify(94419)
			return repl
		} else {
			__antithesis_instrumentation__.Notify(94420)
		}
		__antithesis_instrumentation__.Notify(94399)
		fallthrough

	case leaseCountConvergence:
		__antithesis_instrumentation__.Notify(94400)

		candidates := make([]roachpb.ReplicaDescriptor, 0, len(existing))
		var bestOption roachpb.ReplicaDescriptor
		bestOptionLeaseCount := int32(math.MaxInt32)
		for _, repl := range existing {
			__antithesis_instrumentation__.Notify(94421)
			if leaseRepl.StoreID() == repl.StoreID {
				__antithesis_instrumentation__.Notify(94424)
				continue
			} else {
				__antithesis_instrumentation__.Notify(94425)
			}
			__antithesis_instrumentation__.Notify(94422)
			storeDesc, ok := a.storePool.getStoreDescriptor(repl.StoreID)
			if !ok {
				__antithesis_instrumentation__.Notify(94426)
				continue
			} else {
				__antithesis_instrumentation__.Notify(94427)
			}
			__antithesis_instrumentation__.Notify(94423)
			if !opts.checkCandidateFullness || func() bool {
				__antithesis_instrumentation__.Notify(94428)
				return float64(storeDesc.Capacity.LeaseCount) < candidateLeasesMean-0.5 == true
			}() == true {
				__antithesis_instrumentation__.Notify(94429)
				candidates = append(candidates, repl)
			} else {
				__antithesis_instrumentation__.Notify(94430)
				if storeDesc.Capacity.LeaseCount < bestOptionLeaseCount {
					__antithesis_instrumentation__.Notify(94431)
					bestOption = repl
					bestOptionLeaseCount = storeDesc.Capacity.LeaseCount
				} else {
					__antithesis_instrumentation__.Notify(94432)
				}
			}
		}
		__antithesis_instrumentation__.Notify(94401)
		if len(candidates) == 0 {
			__antithesis_instrumentation__.Notify(94433)

			if !checkTransferLeaseSource {
				__antithesis_instrumentation__.Notify(94435)
				return bestOption
			} else {
				__antithesis_instrumentation__.Notify(94436)
			}
			__antithesis_instrumentation__.Notify(94434)
			return roachpb.ReplicaDescriptor{}
		} else {
			__antithesis_instrumentation__.Notify(94437)
		}
		__antithesis_instrumentation__.Notify(94402)
		a.randGen.Lock()
		defer a.randGen.Unlock()
		return candidates[a.randGen.Intn(len(candidates))]

	case qpsConvergence:
		__antithesis_instrumentation__.Notify(94403)
		leaseReplQPS, _ := stats.avgQPS()
		candidates := make([]roachpb.StoreID, 0, len(existing)-1)
		for _, repl := range existing {
			__antithesis_instrumentation__.Notify(94438)
			if repl.StoreID != leaseRepl.StoreID() {
				__antithesis_instrumentation__.Notify(94439)
				candidates = append(candidates, repl.StoreID)
			} else {
				__antithesis_instrumentation__.Notify(94440)
			}
		}
		__antithesis_instrumentation__.Notify(94404)

		bestStore, noRebalanceReason := bestStoreToMinimizeQPSDelta(
			leaseReplQPS,
			qpsRebalanceThreshold.Get(&a.storePool.st.SV),
			minQPSDifferenceForTransfers.Get(&a.storePool.st.SV),
			leaseRepl.StoreID(),
			candidates,
			storeDescMap,
		)

		switch noRebalanceReason {
		case noBetterCandidate:
			__antithesis_instrumentation__.Notify(94441)
			a.metrics.loadBasedLeaseTransferMetrics.CannotFindBetterCandidate.Inc(1)
			log.VEventf(ctx, 5, "r%d: could not find a better target for lease", leaseRepl.GetRangeID())
			return roachpb.ReplicaDescriptor{}
		case existingNotOverfull:
			__antithesis_instrumentation__.Notify(94442)
			a.metrics.loadBasedLeaseTransferMetrics.ExistingNotOverfull.Inc(1)
			log.VEventf(
				ctx, 5, "r%d: existing leaseholder s%d is not overfull",
				leaseRepl.GetRangeID(), leaseRepl.StoreID(),
			)
			return roachpb.ReplicaDescriptor{}
		case deltaNotSignificant:
			__antithesis_instrumentation__.Notify(94443)
			a.metrics.loadBasedLeaseTransferMetrics.DeltaNotSignificant.Inc(1)
			log.VEventf(
				ctx, 5,
				"r%d: delta between s%d and the coldest follower (ignoring r%d's lease) is not large enough",
				leaseRepl.GetRangeID(), leaseRepl.StoreID(), leaseRepl.GetRangeID(),
			)
			return roachpb.ReplicaDescriptor{}
		case significantlySwitchesRelativeDisposition:
			__antithesis_instrumentation__.Notify(94444)
			a.metrics.loadBasedLeaseTransferMetrics.SignificantlySwitchesRelativeDisposition.Inc(1)
			log.VEventf(ctx, 5,
				"r%d: lease transfer away from s%d would make it hotter than the coldest follower",
				leaseRepl.GetRangeID(), leaseRepl.StoreID())
			return roachpb.ReplicaDescriptor{}
		case missingStatsForExistingStore:
			__antithesis_instrumentation__.Notify(94445)
			a.metrics.loadBasedLeaseTransferMetrics.MissingStatsForExistingStore.Inc(1)
			log.VEventf(
				ctx, 5, "r%d: missing stats for leaseholder s%d",
				leaseRepl.GetRangeID(), leaseRepl.StoreID(),
			)
			return roachpb.ReplicaDescriptor{}
		case shouldRebalance:
			__antithesis_instrumentation__.Notify(94446)
			a.metrics.loadBasedLeaseTransferMetrics.ShouldTransfer.Inc(1)
			log.VEventf(
				ctx,
				5,
				"r%d: should transfer lease (qps=%0.2f) from s%d (qps=%0.2f) to s%d (qps=%0.2f)",
				leaseRepl.GetRangeID(),
				leaseReplQPS,
				leaseRepl.StoreID(),
				storeDescMap[leaseRepl.StoreID()].Capacity.QueriesPerSecond,
				bestStore,
				storeDescMap[bestStore].Capacity.QueriesPerSecond,
			)
		default:
			__antithesis_instrumentation__.Notify(94447)
			log.Fatalf(ctx, "unknown declineReason: %v", noRebalanceReason)
		}
		__antithesis_instrumentation__.Notify(94405)

		for _, repl := range existing {
			__antithesis_instrumentation__.Notify(94448)
			if repl.StoreID == bestStore {
				__antithesis_instrumentation__.Notify(94449)
				return repl
			} else {
				__antithesis_instrumentation__.Notify(94450)
			}
		}
		__antithesis_instrumentation__.Notify(94406)
		panic("unreachable")
	default:
		__antithesis_instrumentation__.Notify(94407)
		log.Fatalf(ctx, "unexpected lease transfer goal %d", g)
	}
	__antithesis_instrumentation__.Notify(94369)
	panic("unreachable")
}

func getCandidateWithMinQPS(
	storeQPSMap map[roachpb.StoreID]float64, candidates []roachpb.StoreID,
) (bestCandidate roachpb.StoreID) {
	__antithesis_instrumentation__.Notify(94451)
	minCandidateQPS := math.MaxFloat64
	for _, store := range candidates {
		__antithesis_instrumentation__.Notify(94453)
		candidateQPS, ok := storeQPSMap[store]
		if !ok {
			__antithesis_instrumentation__.Notify(94455)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94456)
		}
		__antithesis_instrumentation__.Notify(94454)
		if minCandidateQPS > candidateQPS {
			__antithesis_instrumentation__.Notify(94457)
			minCandidateQPS = candidateQPS
			bestCandidate = store
		} else {
			__antithesis_instrumentation__.Notify(94458)
		}
	}
	__antithesis_instrumentation__.Notify(94452)
	return bestCandidate
}

func getQPSDelta(storeQPSMap map[roachpb.StoreID]float64, domain []roachpb.StoreID) float64 {
	__antithesis_instrumentation__.Notify(94459)
	maxCandidateQPS := float64(0)
	minCandidateQPS := math.MaxFloat64
	for _, cand := range domain {
		__antithesis_instrumentation__.Notify(94461)
		candidateQPS, ok := storeQPSMap[cand]
		if !ok {
			__antithesis_instrumentation__.Notify(94464)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94465)
		}
		__antithesis_instrumentation__.Notify(94462)
		if maxCandidateQPS < candidateQPS {
			__antithesis_instrumentation__.Notify(94466)
			maxCandidateQPS = candidateQPS
		} else {
			__antithesis_instrumentation__.Notify(94467)
		}
		__antithesis_instrumentation__.Notify(94463)
		if minCandidateQPS > candidateQPS {
			__antithesis_instrumentation__.Notify(94468)
			minCandidateQPS = candidateQPS
		} else {
			__antithesis_instrumentation__.Notify(94469)
		}
	}
	__antithesis_instrumentation__.Notify(94460)
	return maxCandidateQPS - minCandidateQPS
}

func (a *Allocator) ShouldTransferLease(
	ctx context.Context,
	conf roachpb.SpanConfig,
	existing []roachpb.ReplicaDescriptor,
	leaseStoreID roachpb.StoreID,
	stats *replicaStats,
) bool {
	__antithesis_instrumentation__.Notify(94470)
	source, ok := a.storePool.getStoreDescriptor(leaseStoreID)
	if !ok {
		__antithesis_instrumentation__.Notify(94475)
		return false
	} else {
		__antithesis_instrumentation__.Notify(94476)
	}
	__antithesis_instrumentation__.Notify(94471)

	preferred := a.preferredLeaseholders(conf, existing)
	if len(preferred) == 1 {
		__antithesis_instrumentation__.Notify(94477)
		return preferred[0].StoreID != leaseStoreID
	} else {
		__antithesis_instrumentation__.Notify(94478)
		if len(preferred) > 1 {
			__antithesis_instrumentation__.Notify(94479)
			existing = preferred

			if !storeHasReplica(leaseStoreID, roachpb.MakeReplicaSet(existing).ReplicationTargets()) {
				__antithesis_instrumentation__.Notify(94480)
				return true
			} else {
				__antithesis_instrumentation__.Notify(94481)
			}
		} else {
			__antithesis_instrumentation__.Notify(94482)
		}
	}
	__antithesis_instrumentation__.Notify(94472)

	sl, _, _ := a.storePool.getStoreList(storeFilterSuspect)
	sl = sl.excludeInvalid(conf.Constraints)
	sl = sl.excludeInvalid(conf.VoterConstraints)
	log.VEventf(ctx, 3, "ShouldTransferLease (lease-holder=%d):\n%s", leaseStoreID, sl)

	existing, _ = a.storePool.liveAndDeadReplicas(existing, false)

	if len(existing) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(94483)
		return (len(existing) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(94484)
			return existing[0].StoreID == source.StoreID == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(94485)
		return false
	} else {
		__antithesis_instrumentation__.Notify(94486)
	}
	__antithesis_instrumentation__.Notify(94473)

	transferDec, _ := a.shouldTransferLeaseForAccessLocality(
		ctx,
		source,
		existing,
		stats,
		nil,
		sl.candidateLeases.mean,
	)
	var result bool
	switch transferDec {
	case shouldNotTransfer:
		__antithesis_instrumentation__.Notify(94487)
		result = false
	case shouldTransfer:
		__antithesis_instrumentation__.Notify(94488)
		result = true
	case decideWithoutStats:
		__antithesis_instrumentation__.Notify(94489)
		result = a.shouldTransferLeaseForLeaseCountConvergence(ctx, sl, source, existing)
	default:
		__antithesis_instrumentation__.Notify(94490)
		log.Fatalf(ctx, "unexpected transfer decision %d", transferDec)
	}
	__antithesis_instrumentation__.Notify(94474)

	log.VEventf(ctx, 3, "ShouldTransferLease decision (lease-holder=%d): %t", leaseStoreID, result)
	return result
}

func (a Allocator) followTheWorkloadPrefersLocal(
	ctx context.Context,
	sl StoreList,
	source roachpb.StoreDescriptor,
	candidate roachpb.StoreID,
	existing []roachpb.ReplicaDescriptor,
	stats *replicaStats,
) bool {
	__antithesis_instrumentation__.Notify(94491)
	adjustments := make(map[roachpb.StoreID]float64)
	decision, _ := a.shouldTransferLeaseForAccessLocality(ctx, source, existing, stats, adjustments, sl.candidateLeases.mean)
	if decision == decideWithoutStats {
		__antithesis_instrumentation__.Notify(94494)
		return false
	} else {
		__antithesis_instrumentation__.Notify(94495)
	}
	__antithesis_instrumentation__.Notify(94492)
	adjustment := adjustments[candidate]
	if adjustment > baseLoadBasedLeaseRebalanceThreshold {
		__antithesis_instrumentation__.Notify(94496)
		log.VEventf(ctx, 3,
			"s%d is a better fit than s%d due to follow-the-workload (score: %.2f; threshold: %.2f)",
			source.StoreID, candidate, adjustment, baseLoadBasedLeaseRebalanceThreshold)
		return true
	} else {
		__antithesis_instrumentation__.Notify(94497)
	}
	__antithesis_instrumentation__.Notify(94493)
	return false
}

func (a Allocator) shouldTransferLeaseForAccessLocality(
	ctx context.Context,
	source roachpb.StoreDescriptor,
	existing []roachpb.ReplicaDescriptor,
	stats *replicaStats,
	rebalanceAdjustments map[roachpb.StoreID]float64,
	candidateLeasesMean float64,
) (transferDecision, roachpb.ReplicaDescriptor) {
	__antithesis_instrumentation__.Notify(94498)

	if stats == nil || func() bool {
		__antithesis_instrumentation__.Notify(94506)
		return !enableLoadBasedLeaseRebalancing.Get(&a.storePool.st.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(94507)
		return decideWithoutStats, roachpb.ReplicaDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(94508)
	}
	__antithesis_instrumentation__.Notify(94499)
	replicaLocalities := a.storePool.getLocalitiesByNode(existing)
	for _, locality := range replicaLocalities {
		__antithesis_instrumentation__.Notify(94509)
		if len(locality.Tiers) == 0 {
			__antithesis_instrumentation__.Notify(94510)
			return decideWithoutStats, roachpb.ReplicaDescriptor{}
		} else {
			__antithesis_instrumentation__.Notify(94511)
		}
	}
	__antithesis_instrumentation__.Notify(94500)

	qpsStats, qpsStatsDur := stats.perLocalityDecayingQPS()

	if qpsStatsDur < MinLeaseTransferStatsDuration {
		__antithesis_instrumentation__.Notify(94512)
		return shouldNotTransfer, roachpb.ReplicaDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(94513)
	}
	__antithesis_instrumentation__.Notify(94501)

	delete(qpsStats, "")
	if len(qpsStats) == 0 {
		__antithesis_instrumentation__.Notify(94514)
		return decideWithoutStats, roachpb.ReplicaDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(94515)
	}
	__antithesis_instrumentation__.Notify(94502)

	replicaWeights := make(map[roachpb.NodeID]float64)
	for requestLocalityStr, qps := range qpsStats {
		__antithesis_instrumentation__.Notify(94516)
		var requestLocality roachpb.Locality
		if err := requestLocality.Set(requestLocalityStr); err != nil {
			__antithesis_instrumentation__.Notify(94518)
			log.Errorf(ctx, "unable to parse locality string %q: %+v", requestLocalityStr, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94519)
		}
		__antithesis_instrumentation__.Notify(94517)
		for nodeID, replicaLocality := range replicaLocalities {
			__antithesis_instrumentation__.Notify(94520)

			replicaWeights[nodeID] += (1 - replicaLocality.DiversityScore(requestLocality)) * qps
		}
	}
	__antithesis_instrumentation__.Notify(94503)

	log.VEventf(ctx, 1,
		"shouldTransferLease qpsStats: %+v, replicaLocalities: %+v, replicaWeights: %+v",
		qpsStats, replicaLocalities, replicaWeights)
	sourceWeight := math.Max(minReplicaWeight, replicaWeights[source.Node.NodeID])

	var bestRepl roachpb.ReplicaDescriptor
	bestReplScore := int32(math.MinInt32)
	for _, repl := range existing {
		__antithesis_instrumentation__.Notify(94521)
		if repl.NodeID == source.Node.NodeID {
			__antithesis_instrumentation__.Notify(94527)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94528)
		}
		__antithesis_instrumentation__.Notify(94522)
		storeDesc, ok := a.storePool.getStoreDescriptor(repl.StoreID)
		if !ok {
			__antithesis_instrumentation__.Notify(94529)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94530)
		}
		__antithesis_instrumentation__.Notify(94523)
		addr, err := a.storePool.gossip.GetNodeIDAddress(repl.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(94531)
			log.Errorf(ctx, "missing address for n%d: %+v", repl.NodeID, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94532)
		}
		__antithesis_instrumentation__.Notify(94524)
		remoteLatency, ok := a.nodeLatencyFn(addr.String())
		if !ok {
			__antithesis_instrumentation__.Notify(94533)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94534)
		}
		__antithesis_instrumentation__.Notify(94525)

		remoteWeight := math.Max(minReplicaWeight, replicaWeights[repl.NodeID])
		replScore, rebalanceAdjustment := loadBasedLeaseRebalanceScore(
			ctx, a.storePool.st, remoteWeight, remoteLatency, storeDesc, sourceWeight, source, candidateLeasesMean)
		if replScore > bestReplScore {
			__antithesis_instrumentation__.Notify(94535)
			bestReplScore = replScore
			bestRepl = repl
		} else {
			__antithesis_instrumentation__.Notify(94536)
		}
		__antithesis_instrumentation__.Notify(94526)
		if rebalanceAdjustments != nil {
			__antithesis_instrumentation__.Notify(94537)
			rebalanceAdjustments[repl.StoreID] = rebalanceAdjustment
		} else {
			__antithesis_instrumentation__.Notify(94538)
		}
	}
	__antithesis_instrumentation__.Notify(94504)

	if bestReplScore > 0 {
		__antithesis_instrumentation__.Notify(94539)
		return shouldTransfer, bestRepl
	} else {
		__antithesis_instrumentation__.Notify(94540)
	}
	__antithesis_instrumentation__.Notify(94505)

	return shouldNotTransfer, bestRepl
}

func loadBasedLeaseRebalanceScore(
	ctx context.Context,
	st *cluster.Settings,
	remoteWeight float64,
	remoteLatency time.Duration,
	remoteStore roachpb.StoreDescriptor,
	sourceWeight float64,
	source roachpb.StoreDescriptor,
	meanLeases float64,
) (int32, float64) {
	__antithesis_instrumentation__.Notify(94541)
	remoteLatencyMillis := float64(remoteLatency) / float64(time.Millisecond)
	rebalanceAdjustment :=
		leaseRebalancingAggressiveness.Get(&st.SV) * 0.1 * math.Log10(remoteWeight/sourceWeight) * math.Log1p(remoteLatencyMillis)

	rebalanceThreshold := baseLoadBasedLeaseRebalanceThreshold - rebalanceAdjustment

	overfullLeaseThreshold := int32(math.Ceil(meanLeases * (1 + rebalanceThreshold)))
	overfullScore := source.Capacity.LeaseCount - overfullLeaseThreshold
	underfullLeaseThreshold := int32(math.Floor(meanLeases * (1 - rebalanceThreshold)))
	underfullScore := underfullLeaseThreshold - remoteStore.Capacity.LeaseCount
	totalScore := overfullScore + underfullScore

	log.VEventf(ctx, 1,
		"node: %d, sourceWeight: %.2f, remoteWeight: %.2f, remoteLatency: %v, "+
			"rebalanceThreshold: %.2f, meanLeases: %.2f, sourceLeaseCount: %d, overfullThreshold: %d, "+
			"remoteLeaseCount: %d, underfullThreshold: %d, totalScore: %d",
		remoteStore.Node.NodeID, sourceWeight, remoteWeight, remoteLatency,
		rebalanceThreshold, meanLeases, source.Capacity.LeaseCount, overfullLeaseThreshold,
		remoteStore.Capacity.LeaseCount, underfullLeaseThreshold, totalScore,
	)
	return totalScore, rebalanceAdjustment
}

func (a Allocator) shouldTransferLeaseForLeaseCountConvergence(
	ctx context.Context,
	sl StoreList,
	source roachpb.StoreDescriptor,
	existing []roachpb.ReplicaDescriptor,
) bool {
	__antithesis_instrumentation__.Notify(94542)

	overfullLeaseThreshold := int32(math.Ceil(sl.candidateLeases.mean * (1 + leaseRebalanceThreshold)))
	minOverfullThreshold := int32(math.Ceil(sl.candidateLeases.mean + 5))
	if overfullLeaseThreshold < minOverfullThreshold {
		__antithesis_instrumentation__.Notify(94546)
		overfullLeaseThreshold = minOverfullThreshold
	} else {
		__antithesis_instrumentation__.Notify(94547)
	}
	__antithesis_instrumentation__.Notify(94543)
	if source.Capacity.LeaseCount > overfullLeaseThreshold {
		__antithesis_instrumentation__.Notify(94548)
		return true
	} else {
		__antithesis_instrumentation__.Notify(94549)
	}
	__antithesis_instrumentation__.Notify(94544)

	if float64(source.Capacity.LeaseCount) > sl.candidateLeases.mean {
		__antithesis_instrumentation__.Notify(94550)
		underfullLeaseThreshold := int32(math.Ceil(sl.candidateLeases.mean * (1 - leaseRebalanceThreshold)))
		minUnderfullThreshold := int32(math.Ceil(sl.candidateLeases.mean - 5))
		if underfullLeaseThreshold > minUnderfullThreshold {
			__antithesis_instrumentation__.Notify(94552)
			underfullLeaseThreshold = minUnderfullThreshold
		} else {
			__antithesis_instrumentation__.Notify(94553)
		}
		__antithesis_instrumentation__.Notify(94551)

		for _, repl := range existing {
			__antithesis_instrumentation__.Notify(94554)
			storeDesc, ok := a.storePool.getStoreDescriptor(repl.StoreID)
			if !ok {
				__antithesis_instrumentation__.Notify(94556)
				continue
			} else {
				__antithesis_instrumentation__.Notify(94557)
			}
			__antithesis_instrumentation__.Notify(94555)
			if storeDesc.Capacity.LeaseCount < underfullLeaseThreshold {
				__antithesis_instrumentation__.Notify(94558)
				return true
			} else {
				__antithesis_instrumentation__.Notify(94559)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(94560)
	}
	__antithesis_instrumentation__.Notify(94545)
	return false
}

func (a Allocator) preferredLeaseholders(
	conf roachpb.SpanConfig, existing []roachpb.ReplicaDescriptor,
) []roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(94561)

	for _, preference := range conf.LeasePreferences {
		__antithesis_instrumentation__.Notify(94563)
		var preferred []roachpb.ReplicaDescriptor
		for _, repl := range existing {
			__antithesis_instrumentation__.Notify(94565)

			storeDesc, ok := a.storePool.getStoreDescriptor(repl.StoreID)
			if !ok {
				__antithesis_instrumentation__.Notify(94567)
				continue
			} else {
				__antithesis_instrumentation__.Notify(94568)
			}
			__antithesis_instrumentation__.Notify(94566)
			if constraint.ConjunctionsCheck(storeDesc, preference.Constraints) {
				__antithesis_instrumentation__.Notify(94569)
				preferred = append(preferred, repl)
			} else {
				__antithesis_instrumentation__.Notify(94570)
			}
		}
		__antithesis_instrumentation__.Notify(94564)
		if len(preferred) > 0 {
			__antithesis_instrumentation__.Notify(94571)
			return preferred
		} else {
			__antithesis_instrumentation__.Notify(94572)
		}
	}
	__antithesis_instrumentation__.Notify(94562)
	return nil
}

func computeQuorum(nodes int) int {
	__antithesis_instrumentation__.Notify(94573)
	return (nodes / 2) + 1
}

func filterBehindReplicas(
	ctx context.Context, raftStatus *raft.Status, replicas []roachpb.ReplicaDescriptor,
) []roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(94574)
	if raftStatus == nil || func() bool {
		__antithesis_instrumentation__.Notify(94577)
		return len(raftStatus.Progress) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94578)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(94579)
	}
	__antithesis_instrumentation__.Notify(94575)
	candidates := make([]roachpb.ReplicaDescriptor, 0, len(replicas))
	for _, r := range replicas {
		__antithesis_instrumentation__.Notify(94580)
		if !replicaIsBehind(raftStatus, r.ReplicaID) {
			__antithesis_instrumentation__.Notify(94581)
			candidates = append(candidates, r)
		} else {
			__antithesis_instrumentation__.Notify(94582)
		}
	}
	__antithesis_instrumentation__.Notify(94576)
	return candidates
}

func replicaIsBehind(raftStatus *raft.Status, replicaID roachpb.ReplicaID) bool {
	__antithesis_instrumentation__.Notify(94583)
	if raftStatus == nil || func() bool {
		__antithesis_instrumentation__.Notify(94586)
		return len(raftStatus.Progress) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94587)
		return true
	} else {
		__antithesis_instrumentation__.Notify(94588)
	}
	__antithesis_instrumentation__.Notify(94584)

	if progress, ok := raftStatus.Progress[uint64(replicaID)]; ok {
		__antithesis_instrumentation__.Notify(94589)
		if uint64(replicaID) == raftStatus.Lead || func() bool {
			__antithesis_instrumentation__.Notify(94590)
			return (progress.State == tracker.StateReplicate && func() bool {
				__antithesis_instrumentation__.Notify(94591)
				return progress.Match >= raftStatus.Commit == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(94592)
			return false
		} else {
			__antithesis_instrumentation__.Notify(94593)
		}
	} else {
		__antithesis_instrumentation__.Notify(94594)
	}
	__antithesis_instrumentation__.Notify(94585)
	return true
}

func replicaMayNeedSnapshot(raftStatus *raft.Status, replica roachpb.ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(94595)

	if replica.GetType() == roachpb.VOTER_INCOMING {
		__antithesis_instrumentation__.Notify(94599)
		return false
	} else {
		__antithesis_instrumentation__.Notify(94600)
	}
	__antithesis_instrumentation__.Notify(94596)
	if raftStatus == nil || func() bool {
		__antithesis_instrumentation__.Notify(94601)
		return len(raftStatus.Progress) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(94602)
		return true
	} else {
		__antithesis_instrumentation__.Notify(94603)
	}
	__antithesis_instrumentation__.Notify(94597)
	if progress, ok := raftStatus.Progress[uint64(replica.ReplicaID)]; ok {
		__antithesis_instrumentation__.Notify(94604)

		return progress.State != tracker.StateReplicate
	} else {
		__antithesis_instrumentation__.Notify(94605)
	}
	__antithesis_instrumentation__.Notify(94598)
	return true
}

func excludeReplicasInNeedOfSnapshots(
	ctx context.Context, raftStatus *raft.Status, replicas []roachpb.ReplicaDescriptor,
) []roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(94606)
	filled := 0
	for _, repl := range replicas {
		__antithesis_instrumentation__.Notify(94608)
		if replicaMayNeedSnapshot(raftStatus, repl) {
			__antithesis_instrumentation__.Notify(94610)
			log.VEventf(
				ctx,
				5,
				"not considering [n%d, s%d] as a potential candidate for a lease transfer"+
					" because the replica may be waiting for a snapshot",
				repl.NodeID, repl.StoreID,
			)
			continue
		} else {
			__antithesis_instrumentation__.Notify(94611)
		}
		__antithesis_instrumentation__.Notify(94609)
		replicas[filled] = repl
		filled++
	}
	__antithesis_instrumentation__.Notify(94607)
	return replicas[:filled]
}

func simulateFilterUnremovableReplicas(
	ctx context.Context,
	raftStatus *raft.Status,
	replicas []roachpb.ReplicaDescriptor,
	brandNewReplicaID roachpb.ReplicaID,
) []roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(94612)
	status := *raftStatus
	status.Progress[uint64(brandNewReplicaID)] = tracker.Progress{
		State: tracker.StateReplicate,
		Match: status.Commit,
	}
	return filterUnremovableReplicas(ctx, &status, replicas, brandNewReplicaID)
}

func filterUnremovableReplicas(
	ctx context.Context,
	raftStatus *raft.Status,
	replicas []roachpb.ReplicaDescriptor,
	brandNewReplicaID roachpb.ReplicaID,
) []roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(94613)
	upToDateReplicas := filterBehindReplicas(ctx, raftStatus, replicas)
	oldQuorum := computeQuorum(len(replicas))
	if len(upToDateReplicas) < oldQuorum {
		__antithesis_instrumentation__.Notify(94618)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(94619)
	}
	__antithesis_instrumentation__.Notify(94614)

	newQuorum := computeQuorum(len(replicas) - 1)
	if len(upToDateReplicas) > newQuorum {
		__antithesis_instrumentation__.Notify(94620)

		if brandNewReplicaID != 0 {
			__antithesis_instrumentation__.Notify(94622)
			candidates := make([]roachpb.ReplicaDescriptor, 0, len(replicas)-len(upToDateReplicas))
			for _, r := range replicas {
				__antithesis_instrumentation__.Notify(94624)
				if r.ReplicaID != brandNewReplicaID {
					__antithesis_instrumentation__.Notify(94625)
					candidates = append(candidates, r)
				} else {
					__antithesis_instrumentation__.Notify(94626)
				}
			}
			__antithesis_instrumentation__.Notify(94623)
			return candidates
		} else {
			__antithesis_instrumentation__.Notify(94627)
		}
		__antithesis_instrumentation__.Notify(94621)
		return replicas
	} else {
		__antithesis_instrumentation__.Notify(94628)
	}
	__antithesis_instrumentation__.Notify(94615)

	candidates := make([]roachpb.ReplicaDescriptor, 0, len(replicas)-len(upToDateReplicas))
	necessary := func(r roachpb.ReplicaDescriptor) bool {
		__antithesis_instrumentation__.Notify(94629)
		if r.ReplicaID == brandNewReplicaID {
			__antithesis_instrumentation__.Notify(94632)
			return true
		} else {
			__antithesis_instrumentation__.Notify(94633)
		}
		__antithesis_instrumentation__.Notify(94630)
		for _, t := range upToDateReplicas {
			__antithesis_instrumentation__.Notify(94634)
			if t == r {
				__antithesis_instrumentation__.Notify(94635)
				return true
			} else {
				__antithesis_instrumentation__.Notify(94636)
			}
		}
		__antithesis_instrumentation__.Notify(94631)
		return false
	}
	__antithesis_instrumentation__.Notify(94616)
	for _, r := range replicas {
		__antithesis_instrumentation__.Notify(94637)
		if !necessary(r) {
			__antithesis_instrumentation__.Notify(94638)
			candidates = append(candidates, r)
		} else {
			__antithesis_instrumentation__.Notify(94639)
		}
	}
	__antithesis_instrumentation__.Notify(94617)
	return candidates
}

func maxReplicaID(replicas []roachpb.ReplicaDescriptor) roachpb.ReplicaID {
	__antithesis_instrumentation__.Notify(94640)
	var max roachpb.ReplicaID
	for i := range replicas {
		__antithesis_instrumentation__.Notify(94642)
		if replicaID := replicas[i].ReplicaID; replicaID > max {
			__antithesis_instrumentation__.Notify(94643)
			max = replicaID
		} else {
			__antithesis_instrumentation__.Notify(94644)
		}
	}
	__antithesis_instrumentation__.Notify(94641)
	return max
}
