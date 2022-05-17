package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/tracker"
)

type splitDelayHelperI interface {
	RaftStatus(context.Context) (roachpb.RangeID, *raft.Status)
	ProposeEmptyCommand(ctx context.Context)
	MaxTicks() int
	TickDuration() time.Duration
	Sleep(context.Context, time.Duration)
}

type splitDelayHelper Replica

func (sdh *splitDelayHelper) RaftStatus(ctx context.Context) (roachpb.RangeID, *raft.Status) {
	__antithesis_instrumentation__.Notify(123453)
	r := (*Replica)(sdh)
	r.mu.RLock()
	raftStatus := r.raftStatusRLocked()
	if raftStatus != nil {
		__antithesis_instrumentation__.Notify(123455)
		updateRaftProgressFromActivity(
			ctx, raftStatus.Progress, r.descRLocked().Replicas().Descriptors(),
			func(replicaID roachpb.ReplicaID) bool {
				__antithesis_instrumentation__.Notify(123456)
				return r.mu.lastUpdateTimes.isFollowerActiveSince(
					ctx, replicaID, timeutil.Now(), r.store.cfg.RangeLeaseActiveDuration())
			},
		)
	} else {
		__antithesis_instrumentation__.Notify(123457)
	}
	__antithesis_instrumentation__.Notify(123454)
	r.mu.RUnlock()
	return r.RangeID, raftStatus
}

func (sdh *splitDelayHelper) Sleep(ctx context.Context, dur time.Duration) {
	__antithesis_instrumentation__.Notify(123458)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(123459)
	case <-time.After(dur):
		__antithesis_instrumentation__.Notify(123460)
	}
}

func (sdh *splitDelayHelper) ProposeEmptyCommand(ctx context.Context) {
	__antithesis_instrumentation__.Notify(123461)
	r := (*Replica)(sdh)
	r.raftMu.Lock()
	_ = r.withRaftGroup(true, func(rawNode *raft.RawNode) (bool, error) {
		__antithesis_instrumentation__.Notify(123463)

		data := kvserverbase.EncodeRaftCommand(kvserverbase.RaftVersionStandard, makeIDKey(), nil)
		_ = rawNode.Propose(data)

		return true, nil
	})
	__antithesis_instrumentation__.Notify(123462)
	r.raftMu.Unlock()
}

func (sdh *splitDelayHelper) MaxTicks() int {
	__antithesis_instrumentation__.Notify(123464)

	_ = maybeDropMsgApp

	return (*Replica)(sdh).store.cfg.RaftDelaySplitToSuppressSnapshotTicks
}

func (sdh *splitDelayHelper) TickDuration() time.Duration {
	__antithesis_instrumentation__.Notify(123465)
	r := (*Replica)(sdh)
	return r.store.cfg.RaftTickInterval
}

func maybeDelaySplitToAvoidSnapshot(ctx context.Context, sdh splitDelayHelperI) string {
	__antithesis_instrumentation__.Notify(123466)
	maxDelaySplitToAvoidSnapshotTicks := sdh.MaxTicks()
	tickDur := sdh.TickDuration()
	budget := tickDur * time.Duration(maxDelaySplitToAvoidSnapshotTicks)

	var slept time.Duration
	var problems []string
	var lastProblems []string
	var i int
	for slept < budget {
		__antithesis_instrumentation__.Notify(123470)
		i++
		problems = problems[:0]
		rangeID, raftStatus := sdh.RaftStatus(ctx)

		if raftStatus == nil || func() bool {
			__antithesis_instrumentation__.Notify(123475)
			return raftStatus.RaftState == raft.StateFollower == true
		}() == true {
			__antithesis_instrumentation__.Notify(123476)

			problems = append(problems, "replica is raft follower")
			break
		} else {
			__antithesis_instrumentation__.Notify(123477)
		}
		__antithesis_instrumentation__.Notify(123471)

		if raftStatus.RaftState != raft.StateLeader {
			__antithesis_instrumentation__.Notify(123478)
			problems = append(problems, fmt.Sprintf("not leader (%s)", raftStatus.RaftState))
		} else {
			__antithesis_instrumentation__.Notify(123479)
		}
		__antithesis_instrumentation__.Notify(123472)

		for replicaID, pr := range raftStatus.Progress {
			__antithesis_instrumentation__.Notify(123480)
			if pr.State != tracker.StateReplicate {
				__antithesis_instrumentation__.Notify(123481)
				if !pr.RecentActive {
					__antithesis_instrumentation__.Notify(123483)
					if slept < tickDur {
						__antithesis_instrumentation__.Notify(123485)

						problems = append(problems, fmt.Sprintf("r%d/%d inactive", rangeID, replicaID))
						if i == 1 {
							__antithesis_instrumentation__.Notify(123486)

							sdh.ProposeEmptyCommand(ctx)
						} else {
							__antithesis_instrumentation__.Notify(123487)
						}
					} else {
						__antithesis_instrumentation__.Notify(123488)
					}
					__antithesis_instrumentation__.Notify(123484)
					continue
				} else {
					__antithesis_instrumentation__.Notify(123489)
				}
				__antithesis_instrumentation__.Notify(123482)
				problems = append(problems, fmt.Sprintf("replica r%d/%d not caught up: %+v", rangeID, replicaID, &pr))
			} else {
				__antithesis_instrumentation__.Notify(123490)
			}
		}
		__antithesis_instrumentation__.Notify(123473)
		if len(problems) == 0 {
			__antithesis_instrumentation__.Notify(123491)
			break
		} else {
			__antithesis_instrumentation__.Notify(123492)
		}
		__antithesis_instrumentation__.Notify(123474)

		lastProblems = problems

		sleepDur := time.Duration(float64(tickDur) * (1.0 - math.Exp(-float64(i-1)/float64(maxDelaySplitToAvoidSnapshotTicks+1))))
		sdh.Sleep(ctx, sleepDur)
		slept += sleepDur

		if err := ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(123493)
			problems = append(problems, err.Error())
			break
		} else {
			__antithesis_instrumentation__.Notify(123494)
		}
	}
	__antithesis_instrumentation__.Notify(123467)

	var msg string

	if len(problems) != 0 {
		__antithesis_instrumentation__.Notify(123495)
		lastProblems = problems
	} else {
		__antithesis_instrumentation__.Notify(123496)
	}
	__antithesis_instrumentation__.Notify(123468)
	if len(lastProblems) != 0 {
		__antithesis_instrumentation__.Notify(123497)
		msg = fmt.Sprintf("; delayed by %.1fs to resolve: %s", slept.Seconds(), strings.Join(lastProblems, "; "))
		if len(problems) != 0 {
			__antithesis_instrumentation__.Notify(123498)
			msg += " (without success)"
		} else {
			__antithesis_instrumentation__.Notify(123499)
		}
	} else {
		__antithesis_instrumentation__.Notify(123500)
	}
	__antithesis_instrumentation__.Notify(123469)

	return msg
}
