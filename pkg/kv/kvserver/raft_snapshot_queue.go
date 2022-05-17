package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/tracker"
)

const (
	raftSnapshotQueueTimerDuration = 0

	raftSnapshotPriority float64 = 0
)

type raftSnapshotQueue struct {
	*baseQueue
}

func newRaftSnapshotQueue(store *Store) *raftSnapshotQueue {
	__antithesis_instrumentation__.Notify(112978)
	rq := &raftSnapshotQueue{}
	rq.baseQueue = newBaseQueue(
		"raftsnapshot", rq, store,
		queueConfig{
			maxSize: defaultQueueMaxSize,

			needsLease:           false,
			needsSystemConfig:    false,
			acceptsUnsplitRanges: true,
			processTimeoutFunc:   makeRateLimitedTimeoutFunc(recoverySnapshotRate),
			successes:            store.metrics.RaftSnapshotQueueSuccesses,
			failures:             store.metrics.RaftSnapshotQueueFailures,
			pending:              store.metrics.RaftSnapshotQueuePending,
			processingNanos:      store.metrics.RaftSnapshotQueueProcessingNanos,
		},
	)
	return rq
}

func (rq *raftSnapshotQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, repl *Replica, _ spanconfig.StoreReader,
) (shouldQueue bool, priority float64) {
	__antithesis_instrumentation__.Notify(112979)

	if status := repl.RaftStatus(); status != nil {
		__antithesis_instrumentation__.Notify(112981)

		for _, p := range status.Progress {
			__antithesis_instrumentation__.Notify(112982)
			if p.State == tracker.StateSnapshot {
				__antithesis_instrumentation__.Notify(112983)
				if log.V(2) {
					__antithesis_instrumentation__.Notify(112985)
					log.Infof(ctx, "raft snapshot needed, enqueuing")
				} else {
					__antithesis_instrumentation__.Notify(112986)
				}
				__antithesis_instrumentation__.Notify(112984)
				return true, raftSnapshotPriority
			} else {
				__antithesis_instrumentation__.Notify(112987)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(112988)
	}
	__antithesis_instrumentation__.Notify(112980)
	return false, 0
}

func (rq *raftSnapshotQueue) process(
	ctx context.Context, repl *Replica, _ spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(112989)

	if status := repl.RaftStatus(); status != nil {
		__antithesis_instrumentation__.Notify(112991)

		for id, p := range status.Progress {
			__antithesis_instrumentation__.Notify(112992)
			if p.State == tracker.StateSnapshot {
				__antithesis_instrumentation__.Notify(112993)
				if log.V(1) {
					__antithesis_instrumentation__.Notify(112996)
					log.Infof(ctx, "sending raft snapshot")
				} else {
					__antithesis_instrumentation__.Notify(112997)
				}
				__antithesis_instrumentation__.Notify(112994)
				if err := rq.processRaftSnapshot(ctx, repl, roachpb.ReplicaID(id)); err != nil {
					__antithesis_instrumentation__.Notify(112998)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(112999)
				}
				__antithesis_instrumentation__.Notify(112995)
				processed = true
			} else {
				__antithesis_instrumentation__.Notify(113000)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(113001)
	}
	__antithesis_instrumentation__.Notify(112990)
	return processed, nil
}

func (rq *raftSnapshotQueue) processRaftSnapshot(
	ctx context.Context, repl *Replica, id roachpb.ReplicaID,
) error {
	__antithesis_instrumentation__.Notify(113002)
	desc := repl.Desc()
	repDesc, ok := desc.GetReplicaDescriptorByID(id)
	if !ok {
		__antithesis_instrumentation__.Notify(113005)
		return errors.Errorf("%s: replica %d not present in %v", repl, id, desc.Replicas())
	} else {
		__antithesis_instrumentation__.Notify(113006)
	}
	__antithesis_instrumentation__.Notify(113003)
	snapType := kvserverpb.SnapshotRequest_VIA_SNAPSHOT_QUEUE
	skipSnapLogLimiter := log.Every(10 * time.Second)

	if typ := repDesc.GetType(); typ == roachpb.LEARNER || func() bool {
		__antithesis_instrumentation__.Notify(113007)
		return typ == roachpb.NON_VOTER == true
	}() == true {
		__antithesis_instrumentation__.Notify(113008)
		if fn := repl.store.cfg.TestingKnobs.RaftSnapshotQueueSkipReplica; fn != nil && func() bool {
			__antithesis_instrumentation__.Notify(113010)
			return fn() == true
		}() == true {
			__antithesis_instrumentation__.Notify(113011)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(113012)
		}
		__antithesis_instrumentation__.Notify(113009)
		if index := repl.getAndGCSnapshotLogTruncationConstraints(
			timeutil.Now(), repDesc.StoreID,
		); index > 0 {
			__antithesis_instrumentation__.Notify(113013)

			err := errors.Errorf(
				"skipping snapshot; replica is likely a %s in the process of being added: %s",
				typ,
				repDesc,
			)
			if skipSnapLogLimiter.ShouldLog() {
				__antithesis_instrumentation__.Notify(113015)
				log.Infof(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(113016)
				log.VEventf(ctx, 3, "%v", err)
			}
			__antithesis_instrumentation__.Notify(113014)

			repl.reportSnapshotStatus(ctx, repDesc.ReplicaID, err)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(113017)
		}
	} else {
		__antithesis_instrumentation__.Notify(113018)
	}
	__antithesis_instrumentation__.Notify(113004)

	err := repl.sendSnapshot(ctx, repDesc, snapType, kvserverpb.SnapshotRequest_RECOVERY)

	return err
}

func (*raftSnapshotQueue) timer(_ time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(113019)
	return raftSnapshotQueueTimerDuration
}

func (rq *raftSnapshotQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(113020)
	return nil
}
