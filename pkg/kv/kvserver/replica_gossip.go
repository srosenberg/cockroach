package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const configGossipTTL = 0

func (r *Replica) gossipFirstRange(ctx context.Context) {
	__antithesis_instrumentation__.Notify(117522)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gossipFirstRangeLocked(ctx)
}

func (r *Replica) gossipFirstRangeLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(117523)

	if r.store.Gossip() == nil {
		__antithesis_instrumentation__.Notify(117528)
		return
	} else {
		__antithesis_instrumentation__.Notify(117529)
	}
	__antithesis_instrumentation__.Notify(117524)
	log.Event(ctx, "gossiping sentinel and first range")
	if log.V(1) {
		__antithesis_instrumentation__.Notify(117530)
		log.Infof(ctx, "gossiping sentinel from store %d, r%d", r.store.StoreID(), r.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(117531)
	}
	__antithesis_instrumentation__.Notify(117525)
	if err := r.store.Gossip().AddInfo(
		gossip.KeySentinel, r.store.ClusterID().GetBytes(),
		r.store.cfg.SentinelGossipTTL()); err != nil {
		__antithesis_instrumentation__.Notify(117532)
		log.Errorf(ctx, "failed to gossip sentinel: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(117533)
	}
	__antithesis_instrumentation__.Notify(117526)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(117534)
		log.Infof(ctx, "gossiping first range from store %d, r%d: %s",
			r.store.StoreID(), r.RangeID, r.mu.state.Desc.Replicas())
	} else {
		__antithesis_instrumentation__.Notify(117535)
	}
	__antithesis_instrumentation__.Notify(117527)
	if err := r.store.Gossip().AddInfoProto(
		gossip.KeyFirstRangeDescriptor, r.mu.state.Desc, configGossipTTL); err != nil {
		__antithesis_instrumentation__.Notify(117536)
		log.Errorf(ctx, "failed to gossip first range metadata: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(117537)
	}
}

func (r *Replica) shouldGossip(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(117538)
	return r.OwnsValidLease(ctx, r.store.Clock().NowAsClockTimestamp())
}

func (r *Replica) MaybeGossipSystemConfigRaftMuLocked(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(117539)
	if r.ClusterSettings().Version.IsActive(
		ctx, clusterversion.DisableSystemConfigGossipTrigger,
	) {
		__antithesis_instrumentation__.Notify(117548)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117549)
	}
	__antithesis_instrumentation__.Notify(117540)
	r.raftMu.AssertHeld()
	if r.store.Gossip() == nil {
		__antithesis_instrumentation__.Notify(117550)
		log.VEventf(ctx, 2, "not gossiping system config because gossip isn't initialized")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117551)
	}
	__antithesis_instrumentation__.Notify(117541)
	if !r.IsInitialized() {
		__antithesis_instrumentation__.Notify(117552)
		log.VEventf(ctx, 2, "not gossiping system config because the replica isn't initialized")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117553)
	}
	__antithesis_instrumentation__.Notify(117542)
	if !r.ContainsKey(keys.SystemConfigSpan.Key) {
		__antithesis_instrumentation__.Notify(117554)
		log.VEventf(ctx, 3,
			"not gossiping system config because the replica doesn't contain the system config's start key")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117555)
	}
	__antithesis_instrumentation__.Notify(117543)
	if !r.shouldGossip(ctx) {
		__antithesis_instrumentation__.Notify(117556)
		log.VEventf(ctx, 2, "not gossiping system config because the replica doesn't hold the lease")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117557)
	}
	__antithesis_instrumentation__.Notify(117544)

	loadedCfg, err := r.loadSystemConfig(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(117558)
		if errors.Is(err, errSystemConfigIntent) {
			__antithesis_instrumentation__.Notify(117560)
			log.VEventf(ctx, 2, "not gossiping system config because intents were found on SystemConfigSpan")
			r.markSystemConfigGossipFailed()
			return nil
		} else {
			__antithesis_instrumentation__.Notify(117561)
		}
		__antithesis_instrumentation__.Notify(117559)
		return errors.Wrap(err, "could not load SystemConfig span")
	} else {
		__antithesis_instrumentation__.Notify(117562)
	}
	__antithesis_instrumentation__.Notify(117545)

	if gossipedCfg := r.store.Gossip().DeprecatedGetSystemConfig(); gossipedCfg != nil && func() bool {
		__antithesis_instrumentation__.Notify(117563)
		return gossipedCfg.Equal(loadedCfg) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(117564)
		return r.store.Gossip().InfoOriginatedHere(gossip.KeyDeprecatedSystemConfig) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117565)
		log.VEventf(ctx, 2, "not gossiping unchanged system config")

		r.markSystemConfigGossipSuccess()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117566)
	}
	__antithesis_instrumentation__.Notify(117546)

	log.VEventf(ctx, 2, "gossiping system config")
	if err := r.store.Gossip().AddInfoProto(gossip.KeyDeprecatedSystemConfig, loadedCfg, 0); err != nil {
		__antithesis_instrumentation__.Notify(117567)
		return errors.Wrap(err, "failed to gossip system config")
	} else {
		__antithesis_instrumentation__.Notify(117568)
	}
	__antithesis_instrumentation__.Notify(117547)
	r.markSystemConfigGossipSuccess()
	return nil
}

func (r *Replica) MaybeGossipSystemConfigIfHaveFailureRaftMuLocked(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(117569)
	r.mu.RLock()
	failed := r.mu.failureToGossipSystemConfig
	r.mu.RUnlock()
	if !failed {
		__antithesis_instrumentation__.Notify(117571)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117572)
	}
	__antithesis_instrumentation__.Notify(117570)
	return r.MaybeGossipSystemConfigRaftMuLocked(ctx)
}

func (r *Replica) MaybeGossipNodeLivenessRaftMuLocked(
	ctx context.Context, span roachpb.Span,
) error {
	__antithesis_instrumentation__.Notify(117573)
	r.raftMu.AssertHeld()
	if r.store.Gossip() == nil || func() bool {
		__antithesis_instrumentation__.Notify(117579)
		return !r.IsInitialized() == true
	}() == true {
		__antithesis_instrumentation__.Notify(117580)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117581)
	}
	__antithesis_instrumentation__.Notify(117574)
	if !r.ContainsKeyRange(span.Key, span.EndKey) || func() bool {
		__antithesis_instrumentation__.Notify(117582)
		return !r.shouldGossip(ctx) == true
	}() == true {
		__antithesis_instrumentation__.Notify(117583)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117584)
	}
	__antithesis_instrumentation__.Notify(117575)

	ba := roachpb.BatchRequest{}
	ba.Timestamp = r.store.Clock().Now()
	ba.Add(&roachpb.ScanRequest{RequestHeader: roachpb.RequestHeaderFromSpan(span)})

	rec := NewReplicaEvalContext(r, todoSpanSet)
	rw := r.Engine().NewReadOnly(storage.StandardDurability)
	defer rw.Close()

	br, result, pErr :=
		evaluateBatch(ctx, kvserverbase.CmdIDKey(""), rw, rec, nil, &ba, uncertainty.Interval{}, true)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(117585)
		return errors.Wrapf(pErr.GoError(), "couldn't scan node liveness records in span %s", span)
	} else {
		__antithesis_instrumentation__.Notify(117586)
	}
	__antithesis_instrumentation__.Notify(117576)
	if len(result.Local.EncounteredIntents) > 0 {
		__antithesis_instrumentation__.Notify(117587)
		return errors.Errorf("unexpected intents on node liveness span %s: %+v", span, result.Local.EncounteredIntents)
	} else {
		__antithesis_instrumentation__.Notify(117588)
	}
	__antithesis_instrumentation__.Notify(117577)
	kvs := br.Responses[0].GetInner().(*roachpb.ScanResponse).Rows
	log.VEventf(ctx, 2, "gossiping %d node liveness record(s) from span %s", len(kvs), span)
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(117589)
		var kvLiveness, gossipLiveness livenesspb.Liveness
		if err := kv.Value.GetProto(&kvLiveness); err != nil {
			__antithesis_instrumentation__.Notify(117592)
			return errors.Wrapf(err, "failed to unmarshal liveness value %s", kv.Key)
		} else {
			__antithesis_instrumentation__.Notify(117593)
		}
		__antithesis_instrumentation__.Notify(117590)
		key := gossip.MakeNodeLivenessKey(kvLiveness.NodeID)

		if err := r.store.Gossip().GetInfoProto(key, &gossipLiveness); err == nil {
			__antithesis_instrumentation__.Notify(117594)
			if gossipLiveness == kvLiveness && func() bool {
				__antithesis_instrumentation__.Notify(117595)
				return r.store.Gossip().InfoOriginatedHere(key) == true
			}() == true {
				__antithesis_instrumentation__.Notify(117596)
				continue
			} else {
				__antithesis_instrumentation__.Notify(117597)
			}
		} else {
			__antithesis_instrumentation__.Notify(117598)
		}
		__antithesis_instrumentation__.Notify(117591)
		if err := r.store.Gossip().AddInfoProto(key, &kvLiveness, 0); err != nil {
			__antithesis_instrumentation__.Notify(117599)
			return errors.Wrapf(err, "failed to gossip node liveness (%+v)", kvLiveness)
		} else {
			__antithesis_instrumentation__.Notify(117600)
		}
	}
	__antithesis_instrumentation__.Notify(117578)
	return nil
}

var errSystemConfigIntent = errors.New("must retry later due to intent on SystemConfigSpan")

func (r *Replica) loadSystemConfig(ctx context.Context) (*config.SystemConfigEntries, error) {
	__antithesis_instrumentation__.Notify(117601)
	ba := roachpb.BatchRequest{}
	ba.ReadConsistency = roachpb.INCONSISTENT
	ba.Timestamp = r.store.Clock().Now()
	ba.Add(&roachpb.ScanRequest{RequestHeader: roachpb.RequestHeaderFromSpan(keys.SystemConfigSpan)})

	rec := NewReplicaEvalContext(r, todoSpanSet)
	rw := r.Engine().NewReadOnly(storage.StandardDurability)
	defer rw.Close()

	br, result, pErr := evaluateBatch(
		ctx, kvserverbase.CmdIDKey(""), rw, rec, nil, &ba, uncertainty.Interval{}, true,
	)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(117604)
		return nil, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(117605)
	}
	__antithesis_instrumentation__.Notify(117602)
	if intents := result.Local.DetachEncounteredIntents(); len(intents) > 0 {
		__antithesis_instrumentation__.Notify(117606)

		if err := r.store.intentResolver.CleanupIntentsAsync(ctx, intents, false); err != nil {
			__antithesis_instrumentation__.Notify(117608)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(117609)
		}
		__antithesis_instrumentation__.Notify(117607)
		return nil, errSystemConfigIntent
	} else {
		__antithesis_instrumentation__.Notify(117610)
	}
	__antithesis_instrumentation__.Notify(117603)
	kvs := br.Responses[0].GetInner().(*roachpb.ScanResponse).Rows
	sysCfg := &config.SystemConfigEntries{}
	sysCfg.Values = kvs
	return sysCfg, nil
}

func (r *Replica) getLeaseForGossip(ctx context.Context) (bool, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(117611)

	if r.store.Gossip() == nil || func() bool {
		__antithesis_instrumentation__.Notify(117614)
		return !r.IsInitialized() == true
	}() == true {
		__antithesis_instrumentation__.Notify(117615)
		return false, roachpb.NewErrorf("no gossip or range not initialized")
	} else {
		__antithesis_instrumentation__.Notify(117616)
	}
	__antithesis_instrumentation__.Notify(117612)
	var hasLease bool
	var pErr *roachpb.Error
	if err := r.store.Stopper().RunTask(
		ctx, "storage.Replica: acquiring lease to gossip",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(117617)

			_, pErr = r.redirectOnOrAcquireLease(ctx)
			hasLease = pErr == nil
			if pErr != nil {
				__antithesis_instrumentation__.Notify(117618)
				switch e := pErr.GetDetail().(type) {
				case *roachpb.NotLeaseHolderError:
					__antithesis_instrumentation__.Notify(117619)

					if e.LeaseHolder != nil {
						__antithesis_instrumentation__.Notify(117621)
						pErr = nil
					} else {
						__antithesis_instrumentation__.Notify(117622)
					}
				default:
					__antithesis_instrumentation__.Notify(117620)

					log.Warningf(ctx, "could not acquire lease for range gossip: %s", pErr)
				}
			} else {
				__antithesis_instrumentation__.Notify(117623)
			}
		}); err != nil {
		__antithesis_instrumentation__.Notify(117624)
		pErr = roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(117625)
	}
	__antithesis_instrumentation__.Notify(117613)
	return hasLease, pErr
}

func (r *Replica) maybeGossipFirstRange(ctx context.Context) *roachpb.Error {
	__antithesis_instrumentation__.Notify(117626)
	if !r.IsFirstRange() {
		__antithesis_instrumentation__.Notify(117632)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(117633)
	}
	__antithesis_instrumentation__.Notify(117627)

	if gossipClusterID, err := r.store.Gossip().GetClusterID(); err == nil {
		__antithesis_instrumentation__.Notify(117634)
		if gossipClusterID != r.store.ClusterID() {
			__antithesis_instrumentation__.Notify(117635)
			log.Fatalf(
				ctx, "store %d belongs to cluster %s, but attempted to join cluster %s via gossip",
				r.store.StoreID(), r.store.ClusterID(), gossipClusterID)
		} else {
			__antithesis_instrumentation__.Notify(117636)
		}
	} else {
		__antithesis_instrumentation__.Notify(117637)
	}
	__antithesis_instrumentation__.Notify(117628)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(117638)
		log.Infof(ctx, "gossiping cluster ID %q from store %d, r%d", r.store.ClusterID(),
			r.store.StoreID(), r.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(117639)
	}
	__antithesis_instrumentation__.Notify(117629)
	if err := r.store.Gossip().AddClusterID(r.store.ClusterID()); err != nil {
		__antithesis_instrumentation__.Notify(117640)
		log.Errorf(ctx, "failed to gossip cluster ID: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(117641)
	}
	__antithesis_instrumentation__.Notify(117630)

	hasLease, pErr := r.getLeaseForGossip(ctx)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(117642)
		return pErr
	} else {
		__antithesis_instrumentation__.Notify(117643)
		if !hasLease {
			__antithesis_instrumentation__.Notify(117644)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(117645)
		}
	}
	__antithesis_instrumentation__.Notify(117631)
	r.gossipFirstRange(ctx)
	return nil
}
