package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/abortspan"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/gc"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/split"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/tenantrate"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/kr/pretty"
	"go.etcd.io/etcd/raft/v3"
)

const (
	optimizePutThreshold = 10

	replicaChangeTxnName = "change-replica"
	splitTxnName         = "split"
	mergeTxnName         = "merge"

	defaultReplicaRaftMuWarnThreshold = 500 * time.Millisecond
)

var testingDisableQuiescence = envutil.EnvOrDefaultBool("COCKROACH_DISABLE_QUIESCENCE", false)

var disableSyncRaftLog = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.raft_log.disable_synchronization_unsafe",
	"set to true to disable synchronization on Raft log writes to persistent storage. "+
		"Setting to true risks data loss or data corruption on server crashes. "+
		"The setting is meant for internal testing only and SHOULD NOT be used in production.",
	false,
)

const (
	MaxCommandSizeFloor = 4 << 20

	MaxCommandSizeDefault = 64 << 20
)

var MaxCommandSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"kv.raft.command.max_size",
	"maximum size of a raft command",
	MaxCommandSizeDefault,
	func(size int64) error {
		__antithesis_instrumentation__.Notify(114578)
		if size < MaxCommandSizeFloor {
			__antithesis_instrumentation__.Notify(114580)
			return fmt.Errorf("max_size must be greater than %s", humanizeutil.IBytes(MaxCommandSizeFloor))
		} else {
			__antithesis_instrumentation__.Notify(114581)
		}
		__antithesis_instrumentation__.Notify(114579)
		return nil
	},
)

var StrictGCEnforcement = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.gc_ttl.strict_enforcement.enabled",
	"if true, fail to serve requests at timestamps below the TTL even if the data still exists",
	true,
)

type proposalReevaluationReason int

const (
	proposalNoReevaluation proposalReevaluationReason = iota

	proposalIllegalLeaseIndex
)

type atomicDescString struct {
	strPtr unsafe.Pointer
}

func (d *atomicDescString) store(replicaID roachpb.ReplicaID, desc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(114582)
	str := redact.Sprintfn(func(w redact.SafePrinter) {
		__antithesis_instrumentation__.Notify(114584)
		w.Printf("%d/", desc.RangeID)
		if replicaID == 0 {
			__antithesis_instrumentation__.Notify(114586)
			w.SafeString("?:")
		} else {
			__antithesis_instrumentation__.Notify(114587)
			w.Printf("%d:", replicaID)
		}
		__antithesis_instrumentation__.Notify(114585)

		if !desc.IsInitialized() {
			__antithesis_instrumentation__.Notify(114588)
			w.SafeString("{-}")
		} else {
			__antithesis_instrumentation__.Notify(114589)
			const maxRangeChars = 30
			rngStr := keys.PrettyPrintRange(roachpb.Key(desc.StartKey), roachpb.Key(desc.EndKey), maxRangeChars)
			w.UnsafeString(rngStr)
		}
	})
	__antithesis_instrumentation__.Notify(114583)

	atomic.StorePointer(&d.strPtr, unsafe.Pointer(&str))
}

func (d *atomicDescString) String() string {
	__antithesis_instrumentation__.Notify(114590)
	return d.get().StripMarkers()
}

func (d *atomicDescString) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(114591)
	w.Print(d.get())
}

func (d *atomicDescString) get() redact.RedactableString {
	__antithesis_instrumentation__.Notify(114592)
	return *(*redact.RedactableString)(atomic.LoadPointer(&d.strPtr))
}

type atomicConnectionClass uint32

func (c *atomicConnectionClass) get() rpc.ConnectionClass {
	__antithesis_instrumentation__.Notify(114593)
	return rpc.ConnectionClass(atomic.LoadUint32((*uint32)(c)))
}

func (c *atomicConnectionClass) set(cc rpc.ConnectionClass) {
	__antithesis_instrumentation__.Notify(114594)
	atomic.StoreUint32((*uint32)(c), uint32(cc))
}

type Replica struct {
	log.AmbientContext

	RangeID roachpb.RangeID

	replicaID roachpb.ReplicaID

	startKey roachpb.RKey

	creationTime time.Time

	store     *Store
	abortSpan *abortspan.AbortSpan

	leaseholderStats *replicaStats

	writeStats *replicaStats

	creatingReplica *roachpb.ReplicaDescriptor

	readOnlyCmdMu syncutil.RWMutex

	rangeStr atomicDescString

	connectionClass atomicConnectionClass

	raftCtx context.Context

	breaker *replicaCircuitBreaker

	raftMu struct {
		syncutil.Mutex

		stateLoader stateloader.StateLoader

		sideloaded SideloadStorage

		stateMachine replicaStateMachine

		decoder replicaDecoder

		lastToReplica, lastFromReplica roachpb.ReplicaDescriptor
	}

	leaseHistory *leaseHistory

	concMgr concurrency.Manager

	tenantLimiter tenantrate.Limiter

	tenantMetricsRef *tenantMetricsRef

	sideTransportClosedTimestamp sidetransportAccess

	mu struct {
		syncutil.RWMutex

		destroyStatus

		quiescent bool

		laggingFollowersOnQuiesce laggingReplicaSet

		mergeComplete chan struct{}

		mergeTxnID uuid.UUID

		freezeStart hlc.Timestamp

		state kvserverpb.ReplicaState

		lastIndex, lastTerm uint64

		snapshotLogTruncationConstraints map[uuid.UUID]snapTruncationInfo

		raftLogSize int64

		raftLogSizeTrusted bool

		raftLogLastCheckSize int64

		pendingLeaseRequest pendingLeaseRequest

		minLeaseProposedTS hlc.ClockTimestamp

		conf roachpb.SpanConfig

		spanConfigExplicitlySet bool

		proposalBuf propBuf

		proposals map[kvserverbase.CmdIDKey]*ProposalData

		applyingEntries bool

		internalRaftGroup *raft.RawNode

		tombstoneMinReplicaID roachpb.ReplicaID

		leaderID roachpb.ReplicaID

		lastReplicaAdded     roachpb.ReplicaID
		lastReplicaAddedTime time.Time

		lastUpdateTimes lastUpdateTimesMap

		checksums map[uuid.UUID]replicaChecksum

		proposalQuota *quotapool.IntPool

		proposalQuotaBaseIndex uint64

		quotaReleaseQueue []*quotapool.IntAlloc

		ticks int

		droppedMessages int

		stateLoader stateloader.StateLoader

		cachedProtectedTS cachedProtectedTimestampState

		largestPreviousMaxRangeSizeBytes int64

		failureToGossipSystemConfig bool

		tenantID roachpb.TenantID

		closedTimestampSetter closedTimestampSetterInfo
	}

	pendingLogTruncations pendingLogTruncations

	rangefeedMu struct {
		syncutil.RWMutex

		proc *rangefeed.Processor

		opFilter *rangefeed.Filter
	}

	splitQueueThrottle, mergeQueueThrottle util.EveryN

	loadBasedSplitter split.Decider

	unreachablesMu struct {
		syncutil.Mutex
		remotes map[roachpb.ReplicaID]struct{}
	}

	protectedTimestampMu struct {
		syncutil.Mutex

		minStateReadTimestamp hlc.Timestamp

		pendingGCThreshold hlc.Timestamp
	}
}

var _ batcheval.EvalContext = &Replica{}

func (r *Replica) String() string {
	__antithesis_instrumentation__.Notify(114595)
	return redact.StringWithoutMarkers(r)
}

func (r *Replica) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(114596)
	w.Printf("[n%d,s%d,r%s]",
		r.store.Ident.NodeID, r.store.Ident.StoreID, r.rangeStr.get())
}

func (r *Replica) ReplicaID() roachpb.ReplicaID {
	__antithesis_instrumentation__.Notify(114597)
	return r.replicaID
}

func (r *Replica) cleanupFailedProposalLocked(p *ProposalData) {
	__antithesis_instrumentation__.Notify(114598)
	r.mu.AssertHeld()
	delete(r.mu.proposals, p.idKey)
	p.releaseQuota()
}

func (r *Replica) GetMinBytes() int64 {
	__antithesis_instrumentation__.Notify(114599)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.conf.RangeMinBytes
}

func (r *Replica) GetMaxBytes() int64 {
	__antithesis_instrumentation__.Notify(114600)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.conf.RangeMaxBytes
}

func (r *Replica) SetSpanConfig(conf roachpb.SpanConfig) {
	__antithesis_instrumentation__.Notify(114601)
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isInitializedRLocked() && func() bool {
		__antithesis_instrumentation__.Notify(114604)
		return !r.mu.conf.IsEmpty() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(114605)
		return !conf.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(114606)
		total := r.mu.state.Stats.Total()

		if total > conf.RangeMaxBytes && func() bool {
			__antithesis_instrumentation__.Notify(114607)
			return conf.RangeMaxBytes < r.mu.conf.RangeMaxBytes == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(114608)
			return r.mu.largestPreviousMaxRangeSizeBytes < r.mu.conf.RangeMaxBytes == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(114609)
			return r.mu.spanConfigExplicitlySet == true
		}() == true {
			__antithesis_instrumentation__.Notify(114610)
			r.mu.largestPreviousMaxRangeSizeBytes = r.mu.conf.RangeMaxBytes
		} else {
			__antithesis_instrumentation__.Notify(114611)
			if r.mu.largestPreviousMaxRangeSizeBytes > 0 && func() bool {
				__antithesis_instrumentation__.Notify(114612)
				return r.mu.largestPreviousMaxRangeSizeBytes < conf.RangeMaxBytes == true
			}() == true {
				__antithesis_instrumentation__.Notify(114613)

				r.mu.largestPreviousMaxRangeSizeBytes = 0
			} else {
				__antithesis_instrumentation__.Notify(114614)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(114615)
	}
	__antithesis_instrumentation__.Notify(114602)

	if knobs := r.store.TestingKnobs(); knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(114616)
		return knobs.SetSpanConfigInterceptor != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(114617)
		conf = knobs.SetSpanConfigInterceptor(r.descRLocked(), conf)
	} else {
		__antithesis_instrumentation__.Notify(114618)
	}
	__antithesis_instrumentation__.Notify(114603)
	r.mu.conf, r.mu.spanConfigExplicitlySet = conf, true
}

func (r *Replica) IsFirstRange() bool {
	__antithesis_instrumentation__.Notify(114619)
	return r.RangeID == 1
}

func (r *Replica) IsDestroyed() (DestroyReason, error) {
	__antithesis_instrumentation__.Notify(114620)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isDestroyedRLocked()
}

func (r *Replica) isDestroyedRLocked() (DestroyReason, error) {
	__antithesis_instrumentation__.Notify(114621)
	return r.mu.destroyStatus.reason, r.mu.destroyStatus.err
}

func (r *Replica) IsQuiescent() bool {
	__antithesis_instrumentation__.Notify(114622)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.quiescent
}

func (r *Replica) DescAndSpanConfig() (*roachpb.RangeDescriptor, roachpb.SpanConfig) {
	__antithesis_instrumentation__.Notify(114623)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.state.Desc, r.mu.conf
}

func (r *Replica) SpanConfig() roachpb.SpanConfig {
	__antithesis_instrumentation__.Notify(114624)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.conf
}

func (r *Replica) Desc() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(114625)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.state.Desc
}

func (r *Replica) descRLocked() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(114626)
	r.mu.AssertRHeld()
	return r.mu.state.Desc
}

func (r *Replica) closedTimestampPolicyRLocked() roachpb.RangeClosedTimestampPolicy {
	__antithesis_instrumentation__.Notify(114627)
	if r.mu.conf.GlobalReads {
		__antithesis_instrumentation__.Notify(114629)
		if !r.mu.state.Desc.ContainsKey(roachpb.RKey(keys.NodeLivenessPrefix)) {
			__antithesis_instrumentation__.Notify(114630)
			return roachpb.LEAD_FOR_GLOBAL_READS
		} else {
			__antithesis_instrumentation__.Notify(114631)
		}

	} else {
		__antithesis_instrumentation__.Notify(114632)
	}
	__antithesis_instrumentation__.Notify(114628)
	return roachpb.LAG_BY_CLUSTER_SETTING
}

func (r *Replica) NodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(114633)
	return r.store.NodeID()
}

func (r *Replica) GetNodeLocality() roachpb.Locality {
	__antithesis_instrumentation__.Notify(114634)
	return r.store.nodeDesc.Locality
}

func (r *Replica) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(114635)
	return r.store.cfg.Settings
}

func (r *Replica) StoreID() roachpb.StoreID {
	__antithesis_instrumentation__.Notify(114636)
	return r.store.StoreID()
}

func (r *Replica) EvalKnobs() kvserverbase.BatchEvalTestingKnobs {
	__antithesis_instrumentation__.Notify(114637)
	return r.store.cfg.TestingKnobs.EvalKnobs
}

func (r *Replica) Clock() *hlc.Clock {
	__antithesis_instrumentation__.Notify(114638)
	return r.store.Clock()
}

func (r *Replica) Engine() storage.Engine {
	__antithesis_instrumentation__.Notify(114639)
	return r.store.Engine()
}

func (r *Replica) AbortSpan() *abortspan.AbortSpan {
	__antithesis_instrumentation__.Notify(114640)

	return r.abortSpan
}

func (r *Replica) GetConcurrencyManager() concurrency.Manager {
	__antithesis_instrumentation__.Notify(114641)
	return r.concMgr
}

func (r *Replica) GetTerm(i uint64) (uint64, error) {
	__antithesis_instrumentation__.Notify(114642)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.raftTermRLocked(i)
}

func (r *Replica) GetRangeID() roachpb.RangeID {
	__antithesis_instrumentation__.Notify(114643)
	return r.RangeID
}

func (r *Replica) GetGCThreshold() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(114644)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return *r.mu.state.GCThreshold
}

func (r *Replica) ExcludeDataFromBackup() bool {
	__antithesis_instrumentation__.Notify(114645)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.conf.ExcludeDataFromBackup
}

func (r *Replica) excludeReplicaFromBackupRLocked() bool {
	__antithesis_instrumentation__.Notify(114646)
	return r.mu.conf.ExcludeDataFromBackup
}

func (r *Replica) Version() roachpb.Version {
	__antithesis_instrumentation__.Notify(114647)
	if r.mu.state.Version == nil {
		__antithesis_instrumentation__.Notify(114649)

		return roachpb.Version{}
	} else {
		__antithesis_instrumentation__.Notify(114650)
	}
	__antithesis_instrumentation__.Notify(114648)

	r.mu.RLock()
	defer r.mu.RUnlock()
	return *r.mu.state.Version
}

func (r *Replica) GetRangeInfo(ctx context.Context) roachpb.RangeInfo {
	__antithesis_instrumentation__.Notify(114651)
	r.mu.RLock()
	defer r.mu.RUnlock()
	desc := r.descRLocked()
	l, _ := r.getLeaseRLocked()
	closedts := r.closedTimestampPolicyRLocked()

	if !l.Empty() {
		__antithesis_instrumentation__.Notify(114653)
		if _, ok := desc.GetReplicaDescriptorByID(l.Replica.ReplicaID); !ok {
			__antithesis_instrumentation__.Notify(114654)

			log.Errorf(ctx, "leaseholder replica not in descriptor; desc: %s, lease: %s", desc, l)

			l = roachpb.Lease{}
		} else {
			__antithesis_instrumentation__.Notify(114655)
		}
	} else {
		__antithesis_instrumentation__.Notify(114656)
	}
	__antithesis_instrumentation__.Notify(114652)

	return roachpb.RangeInfo{
		Desc:                  *desc,
		Lease:                 l,
		ClosedTimestampPolicy: closedts,
	}
}

func (r *Replica) getImpliedGCThresholdRLocked(
	st kvserverpb.LeaseStatus, isAdmin bool,
) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(114657)

	if isAdmin || func() bool {
		__antithesis_instrumentation__.Notify(114661)
		return !StrictGCEnforcement.Get(&r.store.ClusterSettings().SV) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(114662)
		return r.shouldIgnoreStrictGCEnforcementRLocked() == true
	}() == true {
		__antithesis_instrumentation__.Notify(114663)
		return *r.mu.state.GCThreshold
	} else {
		__antithesis_instrumentation__.Notify(114664)
	}
	__antithesis_instrumentation__.Notify(114658)

	c := r.mu.cachedProtectedTS
	if st.State != kvserverpb.LeaseState_VALID || func() bool {
		__antithesis_instrumentation__.Notify(114665)
		return c.readAt.Less(st.Lease.Start.ToTimestamp()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(114666)
		return *r.mu.state.GCThreshold
	} else {
		__antithesis_instrumentation__.Notify(114667)
	}
	__antithesis_instrumentation__.Notify(114659)

	threshold := gc.CalculateThreshold(st.Now.ToTimestamp(), r.mu.conf.TTL())
	threshold.Forward(*r.mu.state.GCThreshold)

	if !c.earliestProtectionTimestamp.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(114668)
		return c.earliestProtectionTimestamp.Less(threshold) == true
	}() == true {
		__antithesis_instrumentation__.Notify(114669)
		return c.earliestProtectionTimestamp.Prev()
	} else {
		__antithesis_instrumentation__.Notify(114670)
	}
	__antithesis_instrumentation__.Notify(114660)
	return threshold
}

func (r *Replica) isRangefeedEnabled() (ret bool) {
	__antithesis_instrumentation__.Notify(114671)
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.mu.spanConfigExplicitlySet {
		__antithesis_instrumentation__.Notify(114673)
		return true
	} else {
		__antithesis_instrumentation__.Notify(114674)
	}
	__antithesis_instrumentation__.Notify(114672)
	return r.mu.conf.RangefeedEnabled
}

func (r *Replica) shouldIgnoreStrictGCEnforcementRLocked() (ret bool) {
	__antithesis_instrumentation__.Notify(114675)
	if !r.mu.spanConfigExplicitlySet {
		__antithesis_instrumentation__.Notify(114678)
		return true
	} else {
		__antithesis_instrumentation__.Notify(114679)
	}
	__antithesis_instrumentation__.Notify(114676)

	if knobs := r.store.TestingKnobs(); knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(114680)
		return knobs.IgnoreStrictGCEnforcement == true
	}() == true {
		__antithesis_instrumentation__.Notify(114681)
		return true
	} else {
		__antithesis_instrumentation__.Notify(114682)
	}
	__antithesis_instrumentation__.Notify(114677)

	return r.mu.conf.GCPolicy.IgnoreStrictEnforcement
}

func maxReplicaIDOfAny(desc *roachpb.RangeDescriptor) roachpb.ReplicaID {
	__antithesis_instrumentation__.Notify(114683)
	if desc == nil || func() bool {
		__antithesis_instrumentation__.Notify(114686)
		return !desc.IsInitialized() == true
	}() == true {
		__antithesis_instrumentation__.Notify(114687)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(114688)
	}
	__antithesis_instrumentation__.Notify(114684)
	var maxID roachpb.ReplicaID
	for _, repl := range desc.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(114689)
		if repl.ReplicaID > maxID {
			__antithesis_instrumentation__.Notify(114690)
			maxID = repl.ReplicaID
		} else {
			__antithesis_instrumentation__.Notify(114691)
		}
	}
	__antithesis_instrumentation__.Notify(114685)
	return maxID
}

func (r *Replica) LastReplicaAdded() (roachpb.ReplicaID, time.Time) {
	__antithesis_instrumentation__.Notify(114692)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.lastReplicaAdded, r.mu.lastReplicaAddedTime
}

func (r *Replica) GetReplicaDescriptor() (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(114693)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getReplicaDescriptorRLocked()
}

func (r *Replica) getReplicaDescriptorRLocked() (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(114694)
	repDesc, ok := r.mu.state.Desc.GetReplicaDescriptor(r.store.StoreID())
	if ok {
		__antithesis_instrumentation__.Notify(114696)
		return repDesc, nil
	} else {
		__antithesis_instrumentation__.Notify(114697)
	}
	__antithesis_instrumentation__.Notify(114695)
	return roachpb.ReplicaDescriptor{}, roachpb.NewRangeNotFoundError(r.RangeID, r.store.StoreID())
}

func (r *Replica) getMergeCompleteCh() chan struct{} {
	__antithesis_instrumentation__.Notify(114698)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getMergeCompleteChRLocked()
}

func (r *Replica) getMergeCompleteChRLocked() chan struct{} {
	__antithesis_instrumentation__.Notify(114699)
	return r.mu.mergeComplete
}

func (r *Replica) mergeInProgressRLocked() bool {
	__antithesis_instrumentation__.Notify(114700)
	return r.mu.mergeComplete != nil
}

func (r *Replica) setLastReplicaDescriptorsRaftMuLocked(req *kvserverpb.RaftMessageRequest) {
	__antithesis_instrumentation__.Notify(114701)
	r.raftMu.AssertHeld()
	r.raftMu.lastFromReplica = req.FromReplica
	r.raftMu.lastToReplica = req.ToReplica
}

func (r *Replica) GetMVCCStats() enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(114702)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return *r.mu.state.Stats
}

func (r *Replica) GetMaxSplitQPS() (float64, bool) {
	__antithesis_instrumentation__.Notify(114703)
	return r.loadBasedSplitter.MaxQPS(r.Clock().PhysicalTime())
}

func (r *Replica) GetLastSplitQPS() float64 {
	__antithesis_instrumentation__.Notify(114704)
	return r.loadBasedSplitter.LastQPS(r.Clock().PhysicalTime())
}

func (r *Replica) ContainsKey(key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(114705)
	return kvserverbase.ContainsKey(r.Desc(), key)
}

func (r *Replica) ContainsKeyRange(start, end roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(114706)
	return kvserverbase.ContainsKeyRange(r.Desc(), start, end)
}

func (r *Replica) GetLastReplicaGCTimestamp(ctx context.Context) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(114707)
	key := keys.RangeLastReplicaGCTimestampKey(r.RangeID)
	var timestamp hlc.Timestamp
	_, err := storage.MVCCGetProto(ctx, r.store.Engine(), key, hlc.Timestamp{}, &timestamp,
		storage.MVCCGetOptions{})
	if err != nil {
		__antithesis_instrumentation__.Notify(114709)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(114710)
	}
	__antithesis_instrumentation__.Notify(114708)
	return timestamp, nil
}

func (r *Replica) setLastReplicaGCTimestamp(ctx context.Context, timestamp hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(114711)
	key := keys.RangeLastReplicaGCTimestampKey(r.RangeID)
	return storage.MVCCPutProto(ctx, r.store.Engine(), nil, key, hlc.Timestamp{}, nil, &timestamp)
}

func (r *Replica) getQueueLastProcessed(ctx context.Context, queue string) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(114712)
	key := keys.QueueLastProcessedKey(r.Desc().StartKey, queue)
	var timestamp hlc.Timestamp
	if r.store != nil {
		__antithesis_instrumentation__.Notify(114714)
		_, err := storage.MVCCGetProto(ctx, r.store.Engine(), key, hlc.Timestamp{}, &timestamp,
			storage.MVCCGetOptions{})
		if err != nil {
			__antithesis_instrumentation__.Notify(114715)
			log.VErrEventf(ctx, 2, "last processed timestamp unavailable: %s", err)
			return hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(114716)
		}
	} else {
		__antithesis_instrumentation__.Notify(114717)
	}
	__antithesis_instrumentation__.Notify(114713)
	log.VEventf(ctx, 2, "last processed timestamp: %s", timestamp)
	return timestamp, nil
}

func (r *Replica) setQueueLastProcessed(
	ctx context.Context, queue string, timestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(114718)
	key := keys.QueueLastProcessedKey(r.Desc().StartKey, queue)
	return r.store.DB().PutInline(ctx, key, &timestamp)
}

func (r *Replica) RaftStatus() *raft.Status {
	__antithesis_instrumentation__.Notify(114719)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.raftStatusRLocked()
}

func (r *Replica) raftStatusRLocked() *raft.Status {
	__antithesis_instrumentation__.Notify(114720)
	if rg := r.mu.internalRaftGroup; rg != nil {
		__antithesis_instrumentation__.Notify(114722)
		s := rg.Status()
		return &s
	} else {
		__antithesis_instrumentation__.Notify(114723)
	}
	__antithesis_instrumentation__.Notify(114721)
	return nil
}

func (r *Replica) raftBasicStatusRLocked() raft.BasicStatus {
	__antithesis_instrumentation__.Notify(114724)
	if rg := r.mu.internalRaftGroup; rg != nil {
		__antithesis_instrumentation__.Notify(114726)
		return rg.BasicStatus()
	} else {
		__antithesis_instrumentation__.Notify(114727)
	}
	__antithesis_instrumentation__.Notify(114725)
	return raft.BasicStatus{}
}

func (r *Replica) State(ctx context.Context) kvserverpb.RangeInfo {
	__antithesis_instrumentation__.Notify(114728)
	var ri kvserverpb.RangeInfo

	ri.ActiveClosedTimestamp = r.GetClosedTimestamp(ctx)

	ri.RangefeedRegistrations = int64(r.numRangefeedRegistrations())

	r.mu.RLock()
	defer r.mu.RUnlock()
	ri.ReplicaState = *(protoutil.Clone(&r.mu.state)).(*kvserverpb.ReplicaState)
	ri.LastIndex = r.mu.lastIndex
	ri.NumPending = uint64(r.numPendingProposalsRLocked())
	ri.RaftLogSize = r.mu.raftLogSize
	ri.RaftLogSizeTrusted = r.mu.raftLogSizeTrusted
	ri.NumDropped = uint64(r.mu.droppedMessages)
	if r.mu.proposalQuota != nil {
		__antithesis_instrumentation__.Notify(114732)
		ri.ApproximateProposalQuota = int64(r.mu.proposalQuota.ApproximateQuota())
		ri.ProposalQuotaBaseIndex = int64(r.mu.proposalQuotaBaseIndex)
		ri.ProposalQuotaReleaseQueue = make([]int64, len(r.mu.quotaReleaseQueue))
		for i, a := range r.mu.quotaReleaseQueue {
			__antithesis_instrumentation__.Notify(114733)
			if a != nil {
				__antithesis_instrumentation__.Notify(114734)
				ri.ProposalQuotaReleaseQueue[i] = int64(a.Acquired())
			} else {
				__antithesis_instrumentation__.Notify(114735)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(114736)
	}
	__antithesis_instrumentation__.Notify(114729)
	ri.RangeMaxBytes = r.mu.conf.RangeMaxBytes
	if r.mu.tenantID != (roachpb.TenantID{}) {
		__antithesis_instrumentation__.Notify(114737)
		ri.TenantID = r.mu.tenantID.ToUint64()
	} else {
		__antithesis_instrumentation__.Notify(114738)
	}
	__antithesis_instrumentation__.Notify(114730)
	ri.ClosedTimestampPolicy = r.closedTimestampPolicyRLocked()
	r.sideTransportClosedTimestamp.mu.Lock()
	ri.ClosedTimestampSideTransportInfo.ReplicaClosed = r.sideTransportClosedTimestamp.mu.cur.ts
	ri.ClosedTimestampSideTransportInfo.ReplicaLAI = r.sideTransportClosedTimestamp.mu.cur.lai
	r.sideTransportClosedTimestamp.mu.Unlock()
	centralClosed, centralLAI := r.store.cfg.ClosedTimestampReceiver.GetClosedTimestamp(
		ctx, r.RangeID, r.mu.state.Lease.Replica.NodeID)
	ri.ClosedTimestampSideTransportInfo.CentralClosed = centralClosed
	ri.ClosedTimestampSideTransportInfo.CentralLAI = centralLAI
	if err := r.breaker.Signal().Err(); err != nil {
		__antithesis_instrumentation__.Notify(114739)
		ri.CircuitBreakerError = err.Error()
	} else {
		__antithesis_instrumentation__.Notify(114740)
	}
	__antithesis_instrumentation__.Notify(114731)

	return ri
}

func (r *Replica) assertStateRaftMuLockedReplicaMuRLocked(
	ctx context.Context, reader storage.Reader,
) {
	__antithesis_instrumentation__.Notify(114741)
	diskState, err := r.mu.stateLoader.Load(ctx, reader, r.mu.state.Desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(114745)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(114746)
	}
	__antithesis_instrumentation__.Notify(114742)

	diskState.DeprecatedUsingAppliedStateKey = r.mu.state.DeprecatedUsingAppliedStateKey
	if !diskState.Equal(r.mu.state) {
		__antithesis_instrumentation__.Notify(114747)

		log.Errorf(ctx, "on-disk and in-memory state diverged:\n%s",
			pretty.Diff(diskState, r.mu.state))
		r.mu.state.Desc, diskState.Desc = nil, nil
		log.Fatalf(ctx, "on-disk and in-memory state diverged: %s",
			redact.Safe(pretty.Diff(diskState, r.mu.state)))
	} else {
		__antithesis_instrumentation__.Notify(114748)
	}
	__antithesis_instrumentation__.Notify(114743)
	if r.isInitializedRLocked() {
		__antithesis_instrumentation__.Notify(114749)
		if !r.startKey.Equal(r.mu.state.Desc.StartKey) {
			__antithesis_instrumentation__.Notify(114750)
			log.Fatalf(ctx, "denormalized start key %s diverged from %s", r.startKey, r.mu.state.Desc.StartKey)
		} else {
			__antithesis_instrumentation__.Notify(114751)
		}
	} else {
		__antithesis_instrumentation__.Notify(114752)
	}
	__antithesis_instrumentation__.Notify(114744)

	if !r.store.TestingKnobs().DisableEagerReplicaRemoval && func() bool {
		__antithesis_instrumentation__.Notify(114753)
		return r.mu.state.Desc.IsInitialized() == true
	}() == true {
		__antithesis_instrumentation__.Notify(114754)
		replDesc, ok := r.mu.state.Desc.GetReplicaDescriptor(r.store.StoreID())
		if !ok {
			__antithesis_instrumentation__.Notify(114756)
			log.Fatalf(ctx, "%+v does not contain local store s%d", r.mu.state.Desc, r.store.StoreID())
		} else {
			__antithesis_instrumentation__.Notify(114757)
		}
		__antithesis_instrumentation__.Notify(114755)
		if replDesc.ReplicaID != r.replicaID {
			__antithesis_instrumentation__.Notify(114758)
			log.Fatalf(ctx, "replica's replicaID %d diverges from descriptor %+v", r.replicaID, r.mu.state.Desc)
		} else {
			__antithesis_instrumentation__.Notify(114759)
		}
	} else {
		__antithesis_instrumentation__.Notify(114760)
	}
}

func (r *Replica) checkExecutionCanProceed(
	ctx context.Context, ba *roachpb.BatchRequest, g *concurrency.Guard,
) (kvserverpb.LeaseStatus, error) {
	__antithesis_instrumentation__.Notify(114761)
	rSpan, err := keys.Range(ba.Requests)
	if err != nil {
		__antithesis_instrumentation__.Notify(114771)
		return kvserverpb.LeaseStatus{}, err
	} else {
		__antithesis_instrumentation__.Notify(114772)
	}
	__antithesis_instrumentation__.Notify(114762)

	var shouldExtend bool
	postRUnlock := func() { __antithesis_instrumentation__.Notify(114773) }
	__antithesis_instrumentation__.Notify(114763)
	r.mu.RLock()
	defer func() {
		__antithesis_instrumentation__.Notify(114774)
		r.mu.RUnlock()
		postRUnlock()
	}()
	__antithesis_instrumentation__.Notify(114764)

	if !r.isInitializedRLocked() {
		__antithesis_instrumentation__.Notify(114775)
		return kvserverpb.LeaseStatus{}, errors.Errorf("%s not initialized", r)
	} else {
		__antithesis_instrumentation__.Notify(114776)
	}
	__antithesis_instrumentation__.Notify(114765)

	if _, err := r.isDestroyedRLocked(); err != nil {
		__antithesis_instrumentation__.Notify(114777)
		return kvserverpb.LeaseStatus{}, err
	} else {
		__antithesis_instrumentation__.Notify(114778)
	}
	__antithesis_instrumentation__.Notify(114766)

	if err := r.checkSpanInRangeRLocked(ctx, rSpan); err != nil {
		__antithesis_instrumentation__.Notify(114779)
		return kvserverpb.LeaseStatus{}, err
	} else {
		__antithesis_instrumentation__.Notify(114780)
	}
	__antithesis_instrumentation__.Notify(114767)

	st, shouldExtend, err := r.checkGCThresholdAndLeaseRLocked(ctx, ba)
	if err != nil {
		__antithesis_instrumentation__.Notify(114781)
		return kvserverpb.LeaseStatus{}, err
	} else {
		__antithesis_instrumentation__.Notify(114782)
	}
	__antithesis_instrumentation__.Notify(114768)

	if r.mergeInProgressRLocked() && func() bool {
		__antithesis_instrumentation__.Notify(114783)
		return g.HoldingLatches() == true
	}() == true {
		__antithesis_instrumentation__.Notify(114784)

		if err := r.shouldWaitForPendingMergeRLocked(ctx, ba); err != nil {
			__antithesis_instrumentation__.Notify(114785)

			return kvserverpb.LeaseStatus{}, err
		} else {
			__antithesis_instrumentation__.Notify(114786)
		}
	} else {
		__antithesis_instrumentation__.Notify(114787)
	}
	__antithesis_instrumentation__.Notify(114769)

	if shouldExtend {
		__antithesis_instrumentation__.Notify(114788)

		postRUnlock = func() { __antithesis_instrumentation__.Notify(114789); r.maybeExtendLeaseAsync(ctx, st) }
	} else {
		__antithesis_instrumentation__.Notify(114790)
	}
	__antithesis_instrumentation__.Notify(114770)
	return st, nil
}

func (r *Replica) checkGCThresholdAndLeaseRLocked(
	ctx context.Context, ba *roachpb.BatchRequest,
) (kvserverpb.LeaseStatus, bool, error) {
	__antithesis_instrumentation__.Notify(114791)
	now := r.Clock().NowAsClockTimestamp()

	reqTS := ba.WriteTimestamp()
	st := r.leaseStatusForRequestRLocked(ctx, now, reqTS)

	var shouldExtend bool

	if !ba.IsSingleSkipsLeaseCheckRequest() && func() bool {
		__antithesis_instrumentation__.Notify(114794)
		return ba.ReadConsistency != roachpb.INCONSISTENT == true
	}() == true {
		__antithesis_instrumentation__.Notify(114795)

		var err error
		shouldExtend, err = r.leaseGoodToGoForStatusRLocked(ctx, now, reqTS, st)
		if err != nil {
			__antithesis_instrumentation__.Notify(114796)

			if !r.canServeFollowerReadRLocked(ctx, ba) {
				__antithesis_instrumentation__.Notify(114798)

				return kvserverpb.LeaseStatus{}, false, err
			} else {
				__antithesis_instrumentation__.Notify(114799)
			}
			__antithesis_instrumentation__.Notify(114797)

			st, shouldExtend, err = kvserverpb.LeaseStatus{}, false, nil
		} else {
			__antithesis_instrumentation__.Notify(114800)
		}
	} else {
		__antithesis_instrumentation__.Notify(114801)
	}
	__antithesis_instrumentation__.Notify(114792)

	if err := r.checkTSAboveGCThresholdRLocked(ba.EarliestActiveTimestamp(), st, ba.IsAdmin()); err != nil {
		__antithesis_instrumentation__.Notify(114802)
		return kvserverpb.LeaseStatus{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(114803)
	}
	__antithesis_instrumentation__.Notify(114793)

	return st, shouldExtend, nil
}

func (r *Replica) checkExecutionCanProceedForRangeFeed(
	ctx context.Context, rSpan roachpb.RSpan, ts hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(114804)
	now := r.Clock().NowAsClockTimestamp()
	r.mu.RLock()
	defer r.mu.RUnlock()
	status := r.leaseStatusForRequestRLocked(ctx, now, ts)
	if _, err := r.isDestroyedRLocked(); err != nil {
		__antithesis_instrumentation__.Notify(114806)
		return err
	} else {
		__antithesis_instrumentation__.Notify(114807)
		if err := r.checkSpanInRangeRLocked(ctx, rSpan); err != nil {
			__antithesis_instrumentation__.Notify(114808)
			return err
		} else {
			__antithesis_instrumentation__.Notify(114809)
			if err := r.checkTSAboveGCThresholdRLocked(ts, status, false); err != nil {
				__antithesis_instrumentation__.Notify(114810)
				return err
			} else {
				__antithesis_instrumentation__.Notify(114811)
				if r.requiresExpiringLeaseRLocked() {
					__antithesis_instrumentation__.Notify(114812)

					return errors.New("expiration-based leases are incompatible with rangefeeds")
				} else {
					__antithesis_instrumentation__.Notify(114813)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(114805)
	return nil
}

func (r *Replica) checkSpanInRangeRLocked(ctx context.Context, rspan roachpb.RSpan) error {
	__antithesis_instrumentation__.Notify(114814)
	desc := r.mu.state.Desc
	if desc.ContainsKeyRange(rspan.Key, rspan.EndKey) {
		__antithesis_instrumentation__.Notify(114816)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(114817)
	}
	__antithesis_instrumentation__.Notify(114815)
	return roachpb.NewRangeKeyMismatchErrorWithCTPolicy(
		ctx, rspan.Key.AsRawKey(), rspan.EndKey.AsRawKey(), desc, r.mu.state.Lease, r.closedTimestampPolicyRLocked())
}

func (r *Replica) checkTSAboveGCThresholdRLocked(
	ts hlc.Timestamp, st kvserverpb.LeaseStatus, isAdmin bool,
) error {
	__antithesis_instrumentation__.Notify(114818)
	threshold := r.getImpliedGCThresholdRLocked(st, isAdmin)
	if threshold.Less(ts) {
		__antithesis_instrumentation__.Notify(114820)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(114821)
	}
	__antithesis_instrumentation__.Notify(114819)
	return &roachpb.BatchTimestampBeforeGCError{
		Timestamp:              ts,
		Threshold:              threshold,
		DataExcludedFromBackup: r.excludeReplicaFromBackupRLocked(),
	}
}

func (r *Replica) shouldWaitForPendingMergeRLocked(
	ctx context.Context, ba *roachpb.BatchRequest,
) error {
	__antithesis_instrumentation__.Notify(114822)
	if !r.mergeInProgressRLocked() {
		__antithesis_instrumentation__.Notify(114826)
		log.Fatal(ctx, "programming error: shouldWaitForPendingMergeRLocked should"+
			" only be called when a range merge is in progress")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(114827)
	}
	__antithesis_instrumentation__.Notify(114823)

	if ba.IsSingleSubsumeRequest() {
		__antithesis_instrumentation__.Notify(114828)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(114829)
	}
	__antithesis_instrumentation__.Notify(114824)

	if ba.Txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(114830)
		return ba.Txn.ID == r.mu.mergeTxnID == true
	}() == true {
		__antithesis_instrumentation__.Notify(114831)
		if ba.IsSingleRefreshRequest() {
			__antithesis_instrumentation__.Notify(114833)
			desc := r.descRLocked()
			descKey := keys.RangeDescriptorKey(desc.StartKey)
			if ba.Requests[0].GetRefresh().Key.Equal(descKey) {
				__antithesis_instrumentation__.Notify(114834)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(114835)
			}
		} else {
			__antithesis_instrumentation__.Notify(114836)
		}
		__antithesis_instrumentation__.Notify(114832)
		return errors.Errorf("merge transaction attempting to issue "+
			"batch on right-hand side range after subsumption: %s", ba.Summary())
	} else {
		__antithesis_instrumentation__.Notify(114837)
	}
	__antithesis_instrumentation__.Notify(114825)

	return &roachpb.MergeInProgressError{}
}

func (r *Replica) isNewerThanSplit(split *roachpb.SplitTrigger) bool {
	__antithesis_instrumentation__.Notify(114838)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isNewerThanSplitRLocked(split)
}

func (r *Replica) isNewerThanSplitRLocked(split *roachpb.SplitTrigger) bool {
	__antithesis_instrumentation__.Notify(114839)
	rightDesc, _ := split.RightDesc.GetReplicaDescriptor(r.StoreID())

	return r.mu.tombstoneMinReplicaID != 0 || func() bool {
		__antithesis_instrumentation__.Notify(114840)
		return r.replicaID > rightDesc.ReplicaID == true
	}() == true
}

func (r *Replica) WatchForMerge(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(114841)
	ok, err := r.maybeWatchForMerge(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(114843)
		return err
	} else {
		__antithesis_instrumentation__.Notify(114844)
		if !ok {
			__antithesis_instrumentation__.Notify(114845)
			return errors.AssertionFailedf("range merge unexpectedly not in-progress")
		} else {
			__antithesis_instrumentation__.Notify(114846)
		}
	}
	__antithesis_instrumentation__.Notify(114842)
	return nil
}

func (r *Replica) maybeWatchForMerge(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(114847)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.maybeWatchForMergeLocked(ctx)
}

func (r *Replica) maybeWatchForMergeLocked(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(114848)

	desc := r.descRLocked()
	descKey := keys.RangeDescriptorKey(desc.StartKey)
	_, intent, err := storage.MVCCGet(ctx, r.Engine(), descKey, hlc.MaxTimestamp,
		storage.MVCCGetOptions{Inconsistent: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(114854)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(114855)
		if intent == nil {
			__antithesis_instrumentation__.Notify(114856)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(114857)
		}
	}
	__antithesis_instrumentation__.Notify(114849)
	val, _, err := storage.MVCCGetAsTxn(
		ctx, r.Engine(), descKey, intent.Txn.WriteTimestamp, intent.Txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(114858)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(114859)
		if val != nil {
			__antithesis_instrumentation__.Notify(114860)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(114861)
		}
	}
	__antithesis_instrumentation__.Notify(114850)

	mergeCompleteCh := make(chan struct{})
	if r.mu.mergeComplete != nil {
		__antithesis_instrumentation__.Notify(114862)

		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(114863)
	}
	__antithesis_instrumentation__.Notify(114851)
	r.mu.mergeComplete = mergeCompleteCh
	r.mu.mergeTxnID = intent.Txn.ID

	r.maybeUnquiesceLocked()

	taskCtx := r.AnnotateCtx(context.Background())
	err = r.store.stopper.RunAsyncTask(taskCtx, "wait-for-merge", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(114864)
		var pushTxnRes *roachpb.PushTxnResponse
		for retry := retry.Start(base.DefaultRetryOptions()); retry.Next(); {
			__antithesis_instrumentation__.Notify(114868)

			b := &kv.Batch{}
			b.Header.Timestamp = r.Clock().Now()
			b.AddRawRequest(&roachpb.PushTxnRequest{
				RequestHeader: roachpb.RequestHeader{Key: intent.Txn.Key},
				PusherTxn: roachpb.Transaction{
					TxnMeta: enginepb.TxnMeta{Priority: enginepb.MinTxnPriority},
				},
				PusheeTxn: intent.Txn,
				PushType:  roachpb.PUSH_ABORT,
			})
			if err := r.store.DB().Run(ctx, b); err != nil {
				__antithesis_instrumentation__.Notify(114870)
				select {
				case <-r.store.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(114871)

					return
				default:
					__antithesis_instrumentation__.Notify(114872)
					log.Warningf(ctx, "error while watching for merge to complete: PushTxn: %+v", err)

					continue
				}
			} else {
				__antithesis_instrumentation__.Notify(114873)
			}
			__antithesis_instrumentation__.Notify(114869)
			pushTxnRes = b.RawResponse().Responses[0].GetInner().(*roachpb.PushTxnResponse)
			break
		}
		__antithesis_instrumentation__.Notify(114865)

		var mergeCommitted bool
		switch pushTxnRes.PusheeTxn.Status {
		case roachpb.PENDING, roachpb.STAGING:
			__antithesis_instrumentation__.Notify(114874)
			log.Fatalf(ctx, "PushTxn returned while merge transaction %s was still %s",
				intent.Txn.ID.Short(), pushTxnRes.PusheeTxn.Status)
		case roachpb.COMMITTED:
			__antithesis_instrumentation__.Notify(114875)

			mergeCommitted = true
		case roachpb.ABORTED:
			__antithesis_instrumentation__.Notify(114876)

			var getRes *roachpb.GetResponse
			for retry := retry.Start(base.DefaultRetryOptions()); retry.Next(); {
				__antithesis_instrumentation__.Notify(114879)
				metaKey := keys.RangeMetaKey(desc.EndKey)
				res, pErr := kv.SendWrappedWith(ctx, r.store.DB().NonTransactionalSender(), roachpb.Header{

					ReadConsistency: roachpb.READ_UNCOMMITTED,
				}, &roachpb.GetRequest{
					RequestHeader: roachpb.RequestHeader{Key: metaKey.AsRawKey()},
				})
				if pErr != nil {
					__antithesis_instrumentation__.Notify(114881)
					select {
					case <-r.store.stopper.ShouldQuiesce():
						__antithesis_instrumentation__.Notify(114882)

						return
					default:
						__antithesis_instrumentation__.Notify(114883)
						log.Warningf(ctx, "error while watching for merge to complete: Get %s: %s", metaKey, pErr)

						continue
					}
				} else {
					__antithesis_instrumentation__.Notify(114884)
				}
				__antithesis_instrumentation__.Notify(114880)
				getRes = res.(*roachpb.GetResponse)
				break
			}
			__antithesis_instrumentation__.Notify(114877)
			if getRes.Value == nil {
				__antithesis_instrumentation__.Notify(114885)

				mergeCommitted = true
			} else {
				__antithesis_instrumentation__.Notify(114886)

				var meta2Desc roachpb.RangeDescriptor
				if err := getRes.Value.GetProto(&meta2Desc); err != nil {
					__antithesis_instrumentation__.Notify(114888)
					log.Fatalf(ctx, "error while watching for merge to complete: "+
						"unmarshaling meta2 range descriptor: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(114889)
				}
				__antithesis_instrumentation__.Notify(114887)
				if meta2Desc.RangeID != r.RangeID {
					__antithesis_instrumentation__.Notify(114890)
					mergeCommitted = true
				} else {
					__antithesis_instrumentation__.Notify(114891)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(114878)
		}
		__antithesis_instrumentation__.Notify(114866)
		r.raftMu.Lock()
		r.readOnlyCmdMu.Lock()
		r.mu.Lock()
		if mergeCommitted && func() bool {
			__antithesis_instrumentation__.Notify(114892)
			return r.mu.destroyStatus.IsAlive() == true
		}() == true {
			__antithesis_instrumentation__.Notify(114893)

			r.mu.destroyStatus.Set(roachpb.NewRangeNotFoundError(r.RangeID, r.store.StoreID()), destroyReasonMergePending)
		} else {
			__antithesis_instrumentation__.Notify(114894)
		}
		__antithesis_instrumentation__.Notify(114867)

		r.mu.mergeComplete = nil
		r.mu.mergeTxnID = uuid.UUID{}
		close(mergeCompleteCh)
		r.mu.Unlock()
		r.readOnlyCmdMu.Unlock()
		r.raftMu.Unlock()
	})
	__antithesis_instrumentation__.Notify(114852)
	if errors.Is(err, stop.ErrUnavailable) {
		__antithesis_instrumentation__.Notify(114895)

		err = nil
	} else {
		__antithesis_instrumentation__.Notify(114896)
	}
	__antithesis_instrumentation__.Notify(114853)
	return true, err
}

func (r *Replica) maybeTransferRaftLeadershipToLeaseholderLocked(
	ctx context.Context, now hlc.ClockTimestamp,
) {
	__antithesis_instrumentation__.Notify(114897)
	if r.store.TestingKnobs().DisableLeaderFollowsLeaseholder {
		__antithesis_instrumentation__.Notify(114901)
		return
	} else {
		__antithesis_instrumentation__.Notify(114902)
	}
	__antithesis_instrumentation__.Notify(114898)
	status := r.leaseStatusAtRLocked(ctx, now)
	if !status.IsValid() || func() bool {
		__antithesis_instrumentation__.Notify(114903)
		return status.OwnedBy(r.StoreID()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(114904)
		return
	} else {
		__antithesis_instrumentation__.Notify(114905)
	}
	__antithesis_instrumentation__.Notify(114899)
	raftStatus := r.raftStatusRLocked()
	if raftStatus == nil || func() bool {
		__antithesis_instrumentation__.Notify(114906)
		return raftStatus.RaftState != raft.StateLeader == true
	}() == true {
		__antithesis_instrumentation__.Notify(114907)
		return
	} else {
		__antithesis_instrumentation__.Notify(114908)
	}
	__antithesis_instrumentation__.Notify(114900)
	lhReplicaID := uint64(status.Lease.Replica.ReplicaID)
	lhProgress, ok := raftStatus.Progress[lhReplicaID]
	if (ok && func() bool {
		__antithesis_instrumentation__.Notify(114909)
		return lhProgress.Match >= raftStatus.Commit == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(114910)
		return r.store.IsDraining() == true
	}() == true {
		__antithesis_instrumentation__.Notify(114911)
		log.VEventf(ctx, 1, "transferring raft leadership to replica ID %v", lhReplicaID)
		r.store.metrics.RangeRaftLeaderTransfers.Inc(1)
		r.mu.internalRaftGroup.TransferLeader(lhReplicaID)
	} else {
		__antithesis_instrumentation__.Notify(114912)
	}
}

func (r *Replica) getReplicaDescriptorByIDRLocked(
	replicaID roachpb.ReplicaID, fallback roachpb.ReplicaDescriptor,
) (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(114913)
	if repDesc, ok := r.mu.state.Desc.GetReplicaDescriptorByID(replicaID); ok {
		__antithesis_instrumentation__.Notify(114916)
		return repDesc, nil
	} else {
		__antithesis_instrumentation__.Notify(114917)
	}
	__antithesis_instrumentation__.Notify(114914)
	if fallback.ReplicaID == replicaID {
		__antithesis_instrumentation__.Notify(114918)
		return fallback, nil
	} else {
		__antithesis_instrumentation__.Notify(114919)
	}
	__antithesis_instrumentation__.Notify(114915)
	return roachpb.ReplicaDescriptor{},
		errors.Errorf("replica %d not present in %v, %v",
			replicaID, fallback, r.mu.state.Desc.Replicas())
}

func checkIfTxnAborted(
	ctx context.Context, rec batcheval.EvalContext, reader storage.Reader, txn roachpb.Transaction,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(114920)
	var entry roachpb.AbortSpanEntry
	aborted, err := rec.AbortSpan().Get(ctx, reader, txn.ID, &entry)
	if err != nil {
		__antithesis_instrumentation__.Notify(114923)
		return roachpb.NewError(roachpb.NewReplicaCorruptionError(
			errors.Wrap(err, "could not read from AbortSpan")))
	} else {
		__antithesis_instrumentation__.Notify(114924)
	}
	__antithesis_instrumentation__.Notify(114921)
	if aborted {
		__antithesis_instrumentation__.Notify(114925)

		log.VEventf(ctx, 1, "found AbortSpan entry for %s with priority %d",
			txn.ID.Short(), entry.Priority)
		newTxn := txn.Clone()
		if entry.Priority > newTxn.Priority {
			__antithesis_instrumentation__.Notify(114927)
			newTxn.Priority = entry.Priority
		} else {
			__antithesis_instrumentation__.Notify(114928)
		}
		__antithesis_instrumentation__.Notify(114926)
		newTxn.Status = roachpb.ABORTED
		return roachpb.NewErrorWithTxn(
			roachpb.NewTransactionAbortedError(roachpb.ABORT_REASON_ABORT_SPAN), newTxn)
	} else {
		__antithesis_instrumentation__.Notify(114929)
	}
	__antithesis_instrumentation__.Notify(114922)
	return nil
}

func (r *Replica) GetLeaseHistory() []roachpb.Lease {
	__antithesis_instrumentation__.Notify(114930)
	if r.leaseHistory == nil {
		__antithesis_instrumentation__.Notify(114932)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(114933)
	}
	__antithesis_instrumentation__.Notify(114931)

	return r.leaseHistory.get()
}

func EnableLeaseHistory(maxEntries int) func() {
	__antithesis_instrumentation__.Notify(114934)
	originalValue := leaseHistoryMaxEntries
	leaseHistoryMaxEntries = maxEntries
	return func() {
		__antithesis_instrumentation__.Notify(114935)
		leaseHistoryMaxEntries = originalValue
	}
}

func (r *Replica) GetExternalStorage(
	ctx context.Context, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(114936)
	return r.store.cfg.ExternalStorage(ctx, dest)
}

func (r *Replica) GetExternalStorageFromURI(
	ctx context.Context, uri string, user security.SQLUsername,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(114937)
	return r.store.cfg.ExternalStorageFromURI(ctx, uri, user)
}

func (r *Replica) markSystemConfigGossipSuccess() {
	__antithesis_instrumentation__.Notify(114938)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mu.failureToGossipSystemConfig = false
}

func (r *Replica) markSystemConfigGossipFailed() {
	__antithesis_instrumentation__.Notify(114939)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mu.failureToGossipSystemConfig = true
}

func (r *Replica) GetResponseMemoryAccount() *mon.BoundAccount {
	__antithesis_instrumentation__.Notify(114940)

	return nil
}

func (r *Replica) GetEngineCapacity() (roachpb.StoreCapacity, error) {
	__antithesis_instrumentation__.Notify(114941)
	return r.store.Engine().Capacity()
}

func init() {
	tracing.RegisterTagRemapping("r", "range")
}

func (r *Replica) LockRaftMuForTesting() (unlockFunc func()) {
	__antithesis_instrumentation__.Notify(114942)
	r.raftMu.Lock()
	return func() {
		__antithesis_instrumentation__.Notify(114943)
		r.raftMu.Unlock()
	}
}
