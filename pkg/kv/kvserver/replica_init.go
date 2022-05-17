package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/abortspan"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/tracker"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/split"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"go.etcd.io/etcd/raft/v3"
)

const (
	splitQueueThrottleDuration = 5 * time.Second
	mergeQueueThrottleDuration = 5 * time.Second
)

func newReplica(
	ctx context.Context, desc *roachpb.RangeDescriptor, store *Store, replicaID roachpb.ReplicaID,
) (*Replica, error) {
	__antithesis_instrumentation__.Notify(117646)
	repl := newUnloadedReplica(ctx, desc, store, replicaID)
	repl.raftMu.Lock()
	defer repl.raftMu.Unlock()
	repl.mu.Lock()
	defer repl.mu.Unlock()
	if err := repl.loadRaftMuLockedReplicaMuLocked(desc); err != nil {
		__antithesis_instrumentation__.Notify(117648)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(117649)
	}
	__antithesis_instrumentation__.Notify(117647)
	return repl, nil
}

func newUnloadedReplica(
	ctx context.Context, desc *roachpb.RangeDescriptor, store *Store, replicaID roachpb.ReplicaID,
) *Replica {
	__antithesis_instrumentation__.Notify(117650)
	if replicaID == 0 {
		__antithesis_instrumentation__.Notify(117657)
		log.Fatalf(ctx, "cannot construct a replica for range %d with a 0 replica ID", desc.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(117658)
	}
	__antithesis_instrumentation__.Notify(117651)
	r := &Replica{
		AmbientContext: store.cfg.AmbientCtx,
		RangeID:        desc.RangeID,
		replicaID:      replicaID,
		creationTime:   timeutil.Now(),
		store:          store,
		abortSpan:      abortspan.New(desc.RangeID),
		concMgr: concurrency.NewManager(concurrency.Config{
			NodeDesc:          store.nodeDesc,
			RangeDesc:         desc,
			Settings:          store.ClusterSettings(),
			DB:                store.DB(),
			Clock:             store.Clock(),
			Stopper:           store.Stopper(),
			IntentResolver:    store.intentResolver,
			TxnWaitMetrics:    store.txnWaitMetrics,
			SlowLatchGauge:    store.metrics.SlowLatchRequests,
			DisableTxnPushing: store.TestingKnobs().DontPushOnWriteIntentError,
			TxnWaitKnobs:      store.TestingKnobs().TxnWaitKnobs,
		}),
	}
	r.mu.pendingLeaseRequest = makePendingLeaseRequest(r)
	r.mu.stateLoader = stateloader.Make(desc.RangeID)
	r.mu.quiescent = true
	r.mu.conf = store.cfg.DefaultSpanConfig
	split.Init(&r.loadBasedSplitter, rand.Intn, func() float64 {
		__antithesis_instrumentation__.Notify(117659)
		return float64(SplitByLoadQPSThreshold.Get(&store.cfg.Settings.SV))
	}, func() time.Duration {
		__antithesis_instrumentation__.Notify(117660)
		return kvserverbase.SplitByLoadMergeDelay.Get(&store.cfg.Settings.SV)
	})
	__antithesis_instrumentation__.Notify(117652)
	r.mu.proposals = map[kvserverbase.CmdIDKey]*ProposalData{}
	r.mu.checksums = map[uuid.UUID]replicaChecksum{}
	r.mu.proposalBuf.Init((*replicaProposer)(r), tracker.NewLockfreeTracker(), r.Clock(), r.ClusterSettings())
	r.mu.proposalBuf.testing.allowLeaseProposalWhenNotLeader = store.cfg.TestingKnobs.AllowLeaseRequestProposalsWhenNotLeader
	r.mu.proposalBuf.testing.dontCloseTimestamps = store.cfg.TestingKnobs.DontCloseTimestamps
	r.mu.proposalBuf.testing.submitProposalFilter = store.cfg.TestingKnobs.TestingProposalSubmitFilter

	if leaseHistoryMaxEntries > 0 {
		__antithesis_instrumentation__.Notify(117661)
		r.leaseHistory = newLeaseHistory()
	} else {
		__antithesis_instrumentation__.Notify(117662)
	}
	__antithesis_instrumentation__.Notify(117653)
	if store.cfg.StorePool != nil {
		__antithesis_instrumentation__.Notify(117663)
		r.leaseholderStats = newReplicaStats(store.Clock(), store.cfg.StorePool.getNodeLocalityString)
	} else {
		__antithesis_instrumentation__.Notify(117664)
	}
	__antithesis_instrumentation__.Notify(117654)

	r.writeStats = newReplicaStats(store.Clock(), nil)

	r.rangeStr.store(replicaID, &roachpb.RangeDescriptor{RangeID: desc.RangeID})

	r.AmbientContext.AddLogTag("r", &r.rangeStr)
	r.raftCtx = logtags.AddTag(r.AnnotateCtx(context.Background()), "raft", nil)

	r.raftMu.stateLoader = stateloader.Make(desc.RangeID)

	r.splitQueueThrottle = util.Every(splitQueueThrottleDuration)
	r.mergeQueueThrottle = util.Every(mergeQueueThrottleDuration)

	onTrip := func() {
		__antithesis_instrumentation__.Notify(117665)
		telemetry.Inc(telemetryTripAsync)
		r.store.Metrics().ReplicaCircuitBreakerCumTripped.Inc(1)
		store.Metrics().ReplicaCircuitBreakerCurTripped.Inc(1)
	}
	__antithesis_instrumentation__.Notify(117655)
	onReset := func() {
		__antithesis_instrumentation__.Notify(117666)
		store.Metrics().ReplicaCircuitBreakerCurTripped.Dec(1)
	}
	__antithesis_instrumentation__.Notify(117656)
	r.breaker = newReplicaCircuitBreaker(
		store.cfg.Settings, store.stopper, r.AmbientContext, r, onTrip, onReset,
	)
	return r
}

func (r *Replica) setStartKeyLocked(startKey roachpb.RKey) {
	__antithesis_instrumentation__.Notify(117667)
	r.mu.AssertHeld()
	if r.startKey != nil {
		__antithesis_instrumentation__.Notify(117669)
		log.Fatalf(
			r.AnnotateCtx(context.Background()),
			"start key written twice: was %s, now %s", r.startKey, startKey,
		)
	} else {
		__antithesis_instrumentation__.Notify(117670)
	}
	__antithesis_instrumentation__.Notify(117668)
	r.startKey = startKey
}

func (r *Replica) loadRaftMuLockedReplicaMuLocked(desc *roachpb.RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(117671)
	ctx := r.AnnotateCtx(context.TODO())
	if r.mu.state.Desc != nil && func() bool {
		__antithesis_instrumentation__.Notify(117680)
		return r.isInitializedRLocked() == true
	}() == true {
		__antithesis_instrumentation__.Notify(117681)
		log.Fatalf(ctx, "r%d: cannot reinitialize an initialized replica", desc.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(117682)
		if r.replicaID == 0 {
			__antithesis_instrumentation__.Notify(117683)

			log.Fatalf(ctx, "r%d: cannot initialize replica without a replicaID", desc.RangeID)
		} else {
			__antithesis_instrumentation__.Notify(117684)
		}
	}
	__antithesis_instrumentation__.Notify(117672)
	if desc.IsInitialized() {
		__antithesis_instrumentation__.Notify(117685)
		r.setStartKeyLocked(desc.StartKey)
	} else {
		__antithesis_instrumentation__.Notify(117686)
	}
	__antithesis_instrumentation__.Notify(117673)

	r.mu.internalRaftGroup = nil

	var err error
	if r.mu.state, err = r.mu.stateLoader.Load(ctx, r.Engine(), desc); err != nil {
		__antithesis_instrumentation__.Notify(117687)
		return err
	} else {
		__antithesis_instrumentation__.Notify(117688)
	}
	__antithesis_instrumentation__.Notify(117674)
	r.mu.lastIndex, err = r.mu.stateLoader.LoadLastIndex(ctx, r.Engine())
	if err != nil {
		__antithesis_instrumentation__.Notify(117689)
		return err
	} else {
		__antithesis_instrumentation__.Notify(117690)
	}
	__antithesis_instrumentation__.Notify(117675)
	r.mu.lastTerm = invalidLastTerm

	replicaID := r.replicaID
	if replicaDesc, found := r.mu.state.Desc.GetReplicaDescriptor(r.StoreID()); found {
		__antithesis_instrumentation__.Notify(117691)
		replicaID = replicaDesc.ReplicaID
	} else {
		__antithesis_instrumentation__.Notify(117692)
		if desc.IsInitialized() {
			__antithesis_instrumentation__.Notify(117693)
			log.Fatalf(ctx, "r%d: cannot initialize replica which is not in descriptor %v", desc.RangeID, desc)
		} else {
			__antithesis_instrumentation__.Notify(117694)
		}
	}
	__antithesis_instrumentation__.Notify(117676)
	if r.replicaID != replicaID {
		__antithesis_instrumentation__.Notify(117695)
		log.Fatalf(ctx, "attempting to initialize a replica which has ID %d with ID %d",
			r.replicaID, replicaID)
	} else {
		__antithesis_instrumentation__.Notify(117696)
	}
	__antithesis_instrumentation__.Notify(117677)

	r.setDescLockedRaftMuLocked(ctx, desc)

	if r.mu.state.Lease.Sequence > 0 {
		__antithesis_instrumentation__.Notify(117697)
		r.mu.minLeaseProposedTS = r.Clock().NowAsClockTimestamp()
	} else {
		__antithesis_instrumentation__.Notify(117698)
	}
	__antithesis_instrumentation__.Notify(117678)

	ssBase := r.Engine().GetAuxiliaryDir()
	if r.raftMu.sideloaded, err = newDiskSideloadStorage(
		r.store.cfg.Settings,
		desc.RangeID,
		replicaID,
		ssBase,
		r.store.limiters.BulkIOWriteRate,
		r.store.engine,
	); err != nil {
		__antithesis_instrumentation__.Notify(117699)
		return errors.Wrap(err, "while initializing sideloaded storage")
	} else {
		__antithesis_instrumentation__.Notify(117700)
	}
	__antithesis_instrumentation__.Notify(117679)
	r.assertStateRaftMuLockedReplicaMuRLocked(ctx, r.store.Engine())

	r.sideTransportClosedTimestamp.init(r.store.cfg.ClosedTimestampReceiver, desc.RangeID)

	return nil
}

func (r *Replica) IsInitialized() bool {
	__antithesis_instrumentation__.Notify(117701)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isInitializedRLocked()
}

func (r *Replica) TenantID() (roachpb.TenantID, bool) {
	__antithesis_instrumentation__.Notify(117702)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getTenantIDRLocked()
}

func (r *Replica) getTenantIDRLocked() (roachpb.TenantID, bool) {
	__antithesis_instrumentation__.Notify(117703)
	return r.mu.tenantID, r.mu.tenantID != (roachpb.TenantID{})
}

func (r *Replica) isInitializedRLocked() bool {
	__antithesis_instrumentation__.Notify(117704)
	return r.mu.state.Desc.IsInitialized()
}

func (r *Replica) maybeInitializeRaftGroup(ctx context.Context) {
	__antithesis_instrumentation__.Notify(117705)
	r.mu.RLock()

	initialized := r.mu.internalRaftGroup != nil

	removed := !r.mu.destroyStatus.IsAlive()
	r.mu.RUnlock()
	if initialized || func() bool {
		__antithesis_instrumentation__.Notify(117707)
		return removed == true
	}() == true {
		__antithesis_instrumentation__.Notify(117708)
		return
	} else {
		__antithesis_instrumentation__.Notify(117709)
	}
	__antithesis_instrumentation__.Notify(117706)

	r.raftMu.Lock()
	defer r.raftMu.Unlock()
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.withRaftGroupLocked(true, func(raftGroup *raft.RawNode) (bool, error) {
		__antithesis_instrumentation__.Notify(117710)
		return true, nil
	}); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(117711)
		return !errors.Is(err, errRemoved) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117712)
		log.VErrEventf(ctx, 1, "unable to initialize raft group: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(117713)
	}
}

func (r *Replica) setDescRaftMuLocked(ctx context.Context, desc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(117714)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setDescLockedRaftMuLocked(ctx, desc)
}

func (r *Replica) setDescLockedRaftMuLocked(ctx context.Context, desc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(117715)
	if desc.RangeID != r.RangeID {
		__antithesis_instrumentation__.Notify(117722)
		log.Fatalf(ctx, "range descriptor ID (%d) does not match replica's range ID (%d)",
			desc.RangeID, r.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(117723)
	}
	__antithesis_instrumentation__.Notify(117716)
	if r.mu.state.Desc.IsInitialized() && func() bool {
		__antithesis_instrumentation__.Notify(117724)
		return (desc == nil || func() bool {
			__antithesis_instrumentation__.Notify(117725)
			return !desc.IsInitialized() == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117726)
		log.Fatalf(ctx, "cannot replace initialized descriptor with uninitialized one: %+v -> %+v",
			r.mu.state.Desc, desc)
	} else {
		__antithesis_instrumentation__.Notify(117727)
	}
	__antithesis_instrumentation__.Notify(117717)
	if r.mu.state.Desc.IsInitialized() && func() bool {
		__antithesis_instrumentation__.Notify(117728)
		return !r.mu.state.Desc.StartKey.Equal(desc.StartKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117729)
		log.Fatalf(ctx, "attempted to change replica's start key from %s to %s",
			r.mu.state.Desc.StartKey, desc.StartKey)
	} else {
		__antithesis_instrumentation__.Notify(117730)
	}
	__antithesis_instrumentation__.Notify(117718)

	replDesc, found := desc.GetReplicaDescriptor(r.StoreID())
	if found && func() bool {
		__antithesis_instrumentation__.Notify(117731)
		return replDesc.ReplicaID != r.replicaID == true
	}() == true {
		__antithesis_instrumentation__.Notify(117732)
		log.Fatalf(ctx, "attempted to change replica's ID from %d to %d",
			r.replicaID, replDesc.ReplicaID)
	} else {
		__antithesis_instrumentation__.Notify(117733)
	}
	__antithesis_instrumentation__.Notify(117719)

	if desc.IsInitialized() && func() bool {
		__antithesis_instrumentation__.Notify(117734)
		return r.mu.tenantID == (roachpb.TenantID{}) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117735)
		_, tenantID, err := keys.DecodeTenantPrefix(desc.StartKey.AsRawKey())
		if err != nil {
			__antithesis_instrumentation__.Notify(117737)
			log.Fatalf(ctx, "failed to decode tenant prefix from key for "+
				"replica %v: %v", r, err)
		} else {
			__antithesis_instrumentation__.Notify(117738)
		}
		__antithesis_instrumentation__.Notify(117736)
		r.mu.tenantID = tenantID
		r.tenantMetricsRef = r.store.metrics.acquireTenant(tenantID)
		if tenantID != roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(117739)
			r.tenantLimiter = r.store.tenantRateLimiters.GetTenant(ctx, tenantID, r.store.stopper.ShouldQuiesce())
		} else {
			__antithesis_instrumentation__.Notify(117740)
		}
	} else {
		__antithesis_instrumentation__.Notify(117741)
	}
	__antithesis_instrumentation__.Notify(117720)

	oldMaxID := maxReplicaIDOfAny(r.mu.state.Desc)
	newMaxID := maxReplicaIDOfAny(desc)
	if newMaxID > oldMaxID {
		__antithesis_instrumentation__.Notify(117742)
		r.mu.lastReplicaAdded = newMaxID
		r.mu.lastReplicaAddedTime = timeutil.Now()
	} else {
		__antithesis_instrumentation__.Notify(117743)
		if r.mu.lastReplicaAdded > newMaxID {
			__antithesis_instrumentation__.Notify(117744)

			r.mu.lastReplicaAdded = 0
			r.mu.lastReplicaAddedTime = time.Time{}
		} else {
			__antithesis_instrumentation__.Notify(117745)
		}
	}
	__antithesis_instrumentation__.Notify(117721)

	r.rangeStr.store(r.replicaID, desc)
	r.connectionClass.set(rpc.ConnectionClassForKey(desc.StartKey))
	r.concMgr.OnRangeDescUpdated(desc)
	r.mu.state.Desc = desc

	if bytes.HasPrefix(desc.StartKey, keys.NodeLivenessPrefix) {
		__antithesis_instrumentation__.Notify(117746)
		r.store.scheduler.SetPriorityID(desc.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(117747)
	}
}
