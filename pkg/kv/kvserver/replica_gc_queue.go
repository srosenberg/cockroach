package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3"
)

const (
	replicaGCQueueTimerDuration = 100 * time.Millisecond

	ReplicaGCQueueCheckInterval = 12 * time.Hour

	ReplicaGCQueueSuspectCheckInterval = 3 * time.Second
)

const (
	replicaGCPriorityDefault = 0.0

	replicaGCPrioritySuspect = 1.0

	replicaGCPriorityRemoved = 2.0
)

var (
	metaReplicaGCQueueRemoveReplicaCount = metric.Metadata{
		Name:        "queue.replicagc.removereplica",
		Help:        "Number of replica removals attempted by the replica GC queue",
		Measurement: "Replica Removals",
		Unit:        metric.Unit_COUNT,
	}
)

type ReplicaGCQueueMetrics struct {
	RemoveReplicaCount *metric.Counter
}

func makeReplicaGCQueueMetrics() ReplicaGCQueueMetrics {
	__antithesis_instrumentation__.Notify(117444)
	return ReplicaGCQueueMetrics{
		RemoveReplicaCount: metric.NewCounter(metaReplicaGCQueueRemoveReplicaCount),
	}
}

type replicaGCQueue struct {
	*baseQueue
	metrics ReplicaGCQueueMetrics
	db      *kv.DB
}

func newReplicaGCQueue(store *Store, db *kv.DB) *replicaGCQueue {
	__antithesis_instrumentation__.Notify(117445)
	rgcq := &replicaGCQueue{
		metrics: makeReplicaGCQueueMetrics(),
		db:      db,
	}
	store.metrics.registry.AddMetricStruct(&rgcq.metrics)
	rgcq.baseQueue = newBaseQueue(
		"replicaGC", rgcq, store,
		queueConfig{
			maxSize:                  defaultQueueMaxSize,
			needsLease:               false,
			needsRaftInitialized:     true,
			needsSystemConfig:        false,
			acceptsUnsplitRanges:     true,
			processDestroyedReplicas: true,
			successes:                store.metrics.ReplicaGCQueueSuccesses,
			failures:                 store.metrics.ReplicaGCQueueFailures,
			pending:                  store.metrics.ReplicaGCQueuePending,
			processingNanos:          store.metrics.ReplicaGCQueueProcessingNanos,
		},
	)
	return rgcq
}

func (rgcq *replicaGCQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, repl *Replica, _ spanconfig.StoreReader,
) (shouldQueue bool, priority float64) {
	__antithesis_instrumentation__.Notify(117446)
	if _, currentMember := repl.Desc().GetReplicaDescriptor(repl.store.StoreID()); !currentMember {
		__antithesis_instrumentation__.Notify(117449)
		return true, replicaGCPriorityRemoved
	} else {
		__antithesis_instrumentation__.Notify(117450)
	}
	__antithesis_instrumentation__.Notify(117447)
	lastCheck, err := repl.GetLastReplicaGCTimestamp(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(117451)
		log.Errorf(ctx, "could not read last replica GC timestamp: %+v", err)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(117452)
	}
	__antithesis_instrumentation__.Notify(117448)
	isSuspect := replicaIsSuspect(repl)

	return replicaGCShouldQueueImpl(now.ToTimestamp(), lastCheck, isSuspect)
}

func replicaIsSuspect(repl *Replica) bool {
	__antithesis_instrumentation__.Notify(117453)

	replDesc, ok := repl.Desc().GetReplicaDescriptor(repl.store.StoreID())
	if !ok {
		__antithesis_instrumentation__.Notify(117459)
		return true
	} else {
		__antithesis_instrumentation__.Notify(117460)
	}
	__antithesis_instrumentation__.Notify(117454)
	if t := replDesc.GetType(); t != roachpb.VOTER_FULL && func() bool {
		__antithesis_instrumentation__.Notify(117461)
		return t != roachpb.NON_VOTER == true
	}() == true {
		__antithesis_instrumentation__.Notify(117462)
		return true
	} else {
		__antithesis_instrumentation__.Notify(117463)
	}
	__antithesis_instrumentation__.Notify(117455)

	if repl.store.cfg.NodeLiveness == nil {
		__antithesis_instrumentation__.Notify(117464)
		return false
	} else {
		__antithesis_instrumentation__.Notify(117465)
	}
	__antithesis_instrumentation__.Notify(117456)

	raftStatus := repl.RaftStatus()
	if raftStatus == nil {
		__antithesis_instrumentation__.Notify(117466)
		liveness, ok := repl.store.cfg.NodeLiveness.Self()
		return ok && func() bool {
			__antithesis_instrumentation__.Notify(117467)
			return !liveness.Membership.Active() == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(117468)
	}
	__antithesis_instrumentation__.Notify(117457)

	livenessMap := repl.store.cfg.NodeLiveness.GetIsLiveMap()
	switch raftStatus.SoftState.RaftState {

	case raft.StateCandidate, raft.StatePreCandidate:
		__antithesis_instrumentation__.Notify(117469)
		return true

	case raft.StateFollower:
		__antithesis_instrumentation__.Notify(117470)
		leadDesc, ok := repl.Desc().GetReplicaDescriptorByID(roachpb.ReplicaID(raftStatus.Lead))
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(117473)
			return !livenessMap[leadDesc.NodeID].IsLive == true
		}() == true {
			__antithesis_instrumentation__.Notify(117474)
			return true
		} else {
			__antithesis_instrumentation__.Notify(117475)
		}

	case raft.StateLeader:
		__antithesis_instrumentation__.Notify(117471)
		if !repl.Desc().Replicas().CanMakeProgress(func(d roachpb.ReplicaDescriptor) bool {
			__antithesis_instrumentation__.Notify(117476)
			return livenessMap[d.NodeID].IsLive
		}) {
			__antithesis_instrumentation__.Notify(117477)
			return true
		} else {
			__antithesis_instrumentation__.Notify(117478)
		}
	default:
		__antithesis_instrumentation__.Notify(117472)
	}
	__antithesis_instrumentation__.Notify(117458)

	return false
}

func replicaGCShouldQueueImpl(now, lastCheck hlc.Timestamp, isSuspect bool) (bool, float64) {
	__antithesis_instrumentation__.Notify(117479)
	timeout := ReplicaGCQueueCheckInterval
	priority := replicaGCPriorityDefault

	if isSuspect {
		__antithesis_instrumentation__.Notify(117482)
		timeout = ReplicaGCQueueSuspectCheckInterval
		priority = replicaGCPrioritySuspect
	} else {
		__antithesis_instrumentation__.Notify(117483)
	}
	__antithesis_instrumentation__.Notify(117480)

	if !lastCheck.Add(timeout.Nanoseconds(), 0).Less(now) {
		__antithesis_instrumentation__.Notify(117484)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(117485)
	}
	__antithesis_instrumentation__.Notify(117481)
	return true, priority
}

func (rgcq *replicaGCQueue) process(
	ctx context.Context, repl *Replica, _ spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(117486)

	desc := repl.Desc()

	rs, _, err := kv.RangeLookup(ctx, rgcq.db.NonTransactionalSender(), desc.StartKey.AsRawKey(),
		roachpb.CONSISTENT, 0, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(117490)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(117491)
	}
	__antithesis_instrumentation__.Notify(117487)
	if len(rs) != 1 {
		__antithesis_instrumentation__.Notify(117492)

		return false, errors.Errorf("expected 1 range descriptor, got %d", len(rs))
	} else {
		__antithesis_instrumentation__.Notify(117493)
	}
	__antithesis_instrumentation__.Notify(117488)
	replyDesc := rs[0]

	currentDesc, currentMember := replyDesc.GetReplicaDescriptor(repl.store.StoreID())
	sameRange := desc.RangeID == replyDesc.RangeID
	if sameRange && func() bool {
		__antithesis_instrumentation__.Notify(117494)
		return currentMember == true
	}() == true {
		__antithesis_instrumentation__.Notify(117495)

		log.VEventf(ctx, 1, "not gc'able, replica is still in range descriptor: %v", currentDesc)
		if err := repl.setLastReplicaGCTimestamp(ctx, repl.store.Clock().Now()); err != nil {
			__antithesis_instrumentation__.Notify(117497)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(117498)
		}
		__antithesis_instrumentation__.Notify(117496)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(117499)
		if sameRange {
			__antithesis_instrumentation__.Notify(117500)

			if replyDesc.EndKey.Less(desc.EndKey) {
				__antithesis_instrumentation__.Notify(117502)

				log.Infof(ctx, "removing replica with pending split; will incur Raft snapshot for right hand side")
			} else {
				__antithesis_instrumentation__.Notify(117503)
			}
			__antithesis_instrumentation__.Notify(117501)

			rgcq.metrics.RemoveReplicaCount.Inc(1)
			log.VEventf(ctx, 1, "destroying local data")

			nextReplicaID := replyDesc.NextReplicaID

			if err := repl.store.RemoveReplica(ctx, repl, nextReplicaID, RemoveOptions{
				DestroyData: true,
			}); err != nil {
				__antithesis_instrumentation__.Notify(117504)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(117505)
			}
		} else {
			__antithesis_instrumentation__.Notify(117506)

			leftRepl := repl.store.lookupPrecedingReplica(desc.StartKey)
			if leftRepl != nil {
				__antithesis_instrumentation__.Notify(117508)
				leftDesc := leftRepl.Desc()
				rs, _, err := kv.RangeLookup(ctx, rgcq.db.NonTransactionalSender(), leftDesc.StartKey.AsRawKey(),
					roachpb.CONSISTENT, 0, false)
				if err != nil {
					__antithesis_instrumentation__.Notify(117511)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(117512)
				}
				__antithesis_instrumentation__.Notify(117509)
				if len(rs) != 1 {
					__antithesis_instrumentation__.Notify(117513)
					return false, errors.Errorf("expected 1 range descriptor, got %d", len(rs))
				} else {
					__antithesis_instrumentation__.Notify(117514)
				}
				__antithesis_instrumentation__.Notify(117510)
				if leftReplyDesc := &rs[0]; !leftDesc.Equal(leftReplyDesc) {
					__antithesis_instrumentation__.Notify(117515)
					log.VEventf(ctx, 1, "left neighbor %s not up-to-date with meta descriptor %s; cannot safely GC range yet",
						leftDesc, leftReplyDesc)

					rgcq.AddAsync(ctx, leftRepl, replicaGCPriorityDefault)
					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(117516)
				}
			} else {
				__antithesis_instrumentation__.Notify(117517)
			}
			__antithesis_instrumentation__.Notify(117507)

			if err := repl.store.RemoveReplica(ctx, repl, mergedTombstoneReplicaID, RemoveOptions{
				DestroyData: true,
			}); err != nil {
				__antithesis_instrumentation__.Notify(117518)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(117519)
			}
		}
	}
	__antithesis_instrumentation__.Notify(117489)
	return true, nil
}

func (*replicaGCQueue) timer(_ time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(117520)
	return replicaGCQueueTimerDuration
}

func (*replicaGCQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(117521)
	return nil
}
