package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/tenantrate"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/txnwait"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type StoreTestingKnobs struct {
	EvalKnobs               kvserverbase.BatchEvalTestingKnobs
	IntentResolverKnobs     kvserverbase.IntentResolverTestingKnobs
	TxnWaitKnobs            txnwait.TestingKnobs
	ConsistencyTestingKnobs ConsistencyTestingKnobs
	TenantRateKnobs         tenantrate.TestingKnobs
	StorageKnobs            storage.TestingKnobs
	AllocatorKnobs          *AllocatorTestingKnobs

	TestingRequestFilter kvserverbase.ReplicaRequestFilter

	TestingConcurrencyRetryFilter kvserverbase.ReplicaConcurrencyRetryFilter

	TestingProposalFilter kvserverbase.ReplicaProposalFilter

	TestingProposalSubmitFilter func(*ProposalData) (drop bool, err error)

	TestingApplyFilter kvserverbase.ReplicaApplyFilter

	TestingApplyForcedErrFilter kvserverbase.ReplicaApplyFilter

	TestingPostApplyFilter kvserverbase.ReplicaApplyFilter

	TestingResponseErrorEvent func(context.Context, *roachpb.BatchRequest, error)

	TestingResponseFilter kvserverbase.ReplicaResponseFilter

	SlowReplicationThresholdOverride func(ba *roachpb.BatchRequest) time.Duration

	TestingRangefeedFilter kvserverbase.ReplicaRangefeedFilter

	MaxOffset time.Duration

	DisableMaxOffsetCheck bool

	DisableAutomaticLeaseRenewal bool

	LeaseRequestEvent func(ts hlc.Timestamp, storeID roachpb.StoreID, rangeID roachpb.RangeID) *roachpb.Error

	PinnedLeases *PinnedLeasesKnob

	LeaseTransferBlockedOnExtensionEvent func(nextLeader roachpb.ReplicaDescriptor)

	DisableGCQueue bool

	DisableMergeQueue bool

	DisableRaftLogQueue bool

	DisableReplicaGCQueue bool

	DisableReplicateQueue bool

	DisableReplicaRebalancing bool

	DisableLoadBasedSplitting bool

	DisableSplitQueue bool

	DisableTimeSeriesMaintenanceQueue bool

	DisableRaftSnapshotQueue bool

	DisableConsistencyQueue bool

	DisableScanner bool

	DisableLeaderFollowsLeaseholder bool

	DisableRefreshReasonNewLeader bool

	DisableRefreshReasonNewLeaderOrConfigChange bool

	DisableRefreshReasonSnapshotApplied bool

	DisableRefreshReasonTicks bool

	DisableEagerReplicaRemoval bool

	RefreshReasonTicksPeriod int

	DisableProcessRaft bool

	DisableLastProcessedCheck bool

	ReplicateQueueAcceptsUnsplit bool

	SplitQueuePurgatoryChan <-chan time.Time

	SkipMinSizeCheck bool

	DisableLeaseCapacityGossip bool

	SystemLogsGCPeriod time.Duration

	SystemLogsGCGCDone chan<- struct{}

	DontPushOnWriteIntentError bool

	DontRetryPushTxnFailures bool

	DontRecoverIndeterminateCommits bool

	TraceAllRaftEvents bool

	EnableUnconditionalRefreshesInRaftReady bool

	ReceiveSnapshot func(*kvserverpb.SnapshotRequest_Header) error

	ReplicaAddSkipLearnerRollback func() bool

	VoterAddStopAfterLearnerSnapshot func([]roachpb.ReplicationTarget) bool

	NonVoterAfterInitialization func()

	ReplicaSkipInitialSnapshot func() bool

	RaftSnapshotQueueSkipReplica func() bool

	VoterAddStopAfterJointConfig func() bool

	ReplicationAlwaysUseJointConfig func() bool

	BeforeSnapshotSSTIngestion func(IncomingSnapshot, kvserverpb.SnapshotRequest_Type, []string) error

	OnRelocatedOne func(_ []roachpb.ReplicationChange, leaseTarget *roachpb.ReplicationTarget)

	DontIgnoreFailureToTransferLease bool

	MaxApplicationBatchSize int

	RangeFeedPushTxnsInterval time.Duration

	RangeFeedPushTxnsAge time.Duration

	AllowLeaseRequestProposalsWhenNotLeader bool

	DontCloseTimestamps bool

	AllowDangerousReplicationChanges bool

	AllowUnsynchronizedReplicationChanges bool

	PurgeOutdatedReplicasInterceptor func()

	SpanConfigUpdateInterceptor func(spanconfig.Update)

	SetSpanConfigInterceptor func(*roachpb.RangeDescriptor, roachpb.SpanConfig) roachpb.SpanConfig

	InitialReplicaVersionOverride *roachpb.Version

	GossipWhenCapacityDeltaExceedsFraction float64

	TimeSeriesDataStore TimeSeriesDataStore

	OnRaftTimeoutCampaign func(roachpb.RangeID)

	LeaseRenewalSignalChan chan struct{}

	LeaseRenewalOnPostCycle func()

	LeaseRenewalDurationOverride time.Duration

	MakeSystemConfigSpanUnavailableToQueues bool

	UseSystemConfigSpanForQueues bool

	IgnoreStrictGCEnforcement bool
}

func (*StoreTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(126718) }

type NodeLivenessTestingKnobs struct {
	LivenessDuration time.Duration

	RenewalDuration time.Duration

	StorePoolNodeLivenessFn NodeLivenessFunc
}

var _ base.ModuleTestingKnobs = NodeLivenessTestingKnobs{}

func (NodeLivenessTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(126719) }

type AllocatorTestingKnobs struct {
	AllowLeaseTransfersToReplicasNeedingSnapshots bool
}

type PinnedLeasesKnob struct {
	mu struct {
		syncutil.Mutex
		pinned map[roachpb.RangeID]roachpb.StoreID
	}
}

func NewPinnedLeases() *PinnedLeasesKnob {
	__antithesis_instrumentation__.Notify(126720)
	p := &PinnedLeasesKnob{}
	p.mu.pinned = make(map[roachpb.RangeID]roachpb.StoreID)
	return p
}

func (p *PinnedLeasesKnob) PinLease(rangeID roachpb.RangeID, storeID roachpb.StoreID) {
	__antithesis_instrumentation__.Notify(126721)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mu.pinned[rangeID] = storeID
}

func (p *PinnedLeasesKnob) rejectLeaseIfPinnedElsewhere(r *Replica) *roachpb.Error {
	__antithesis_instrumentation__.Notify(126722)
	if p == nil {
		__antithesis_instrumentation__.Notify(126727)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(126728)
	}
	__antithesis_instrumentation__.Notify(126723)

	p.mu.Lock()
	defer p.mu.Unlock()
	pinnedStore, ok := p.mu.pinned[r.RangeID]
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(126729)
		return pinnedStore == r.StoreID() == true
	}() == true {
		__antithesis_instrumentation__.Notify(126730)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(126731)
	}
	__antithesis_instrumentation__.Notify(126724)

	repDesc, err := r.getReplicaDescriptorRLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(126732)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(126733)
	}
	__antithesis_instrumentation__.Notify(126725)
	var pinned *roachpb.ReplicaDescriptor
	if pinnedRep, ok := r.descRLocked().GetReplicaDescriptor(pinnedStore); ok {
		__antithesis_instrumentation__.Notify(126734)
		pinned = &pinnedRep
	} else {
		__antithesis_instrumentation__.Notify(126735)
	}
	__antithesis_instrumentation__.Notify(126726)
	return roachpb.NewError(&roachpb.NotLeaseHolderError{
		Replica:     repDesc,
		LeaseHolder: pinned,
		RangeID:     r.RangeID,
		CustomMsg:   "injected: lease pinned to another store",
	})
}
