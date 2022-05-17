package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type raftTruncatorReplica Replica

var _ replicaForTruncator = &raftTruncatorReplica{}

func (r *raftTruncatorReplica) getRangeID() roachpb.RangeID {
	__antithesis_instrumentation__.Notify(113214)
	return r.RangeID
}

func (r *raftTruncatorReplica) getTruncatedState() roachpb.RaftTruncatedState {
	__antithesis_instrumentation__.Notify(113215)
	r.mu.Lock()
	defer r.mu.Unlock()

	return *r.mu.state.TruncatedState
}

func (r *raftTruncatorReplica) setTruncatedStateAndSideEffects(
	ctx context.Context, trunc *roachpb.RaftTruncatedState, expectedFirstIndexPreTruncation uint64,
) (expectedFirstIndexWasAccurate bool) {
	__antithesis_instrumentation__.Notify(113216)
	_, expectedFirstIndexAccurate := (*Replica)(r).handleTruncatedStateResult(
		ctx, trunc, expectedFirstIndexPreTruncation)
	return expectedFirstIndexAccurate
}

func (r *raftTruncatorReplica) setTruncationDeltaAndTrusted(deltaBytes int64, isDeltaTrusted bool) {
	__antithesis_instrumentation__.Notify(113217)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mu.raftLogSize += deltaBytes
	r.mu.raftLogLastCheckSize += deltaBytes

	if r.mu.raftLogSize < 0 {
		__antithesis_instrumentation__.Notify(113220)
		r.mu.raftLogSize = 0
	} else {
		__antithesis_instrumentation__.Notify(113221)
	}
	__antithesis_instrumentation__.Notify(113218)
	if r.mu.raftLogLastCheckSize < 0 {
		__antithesis_instrumentation__.Notify(113222)
		r.mu.raftLogLastCheckSize = 0
	} else {
		__antithesis_instrumentation__.Notify(113223)
	}
	__antithesis_instrumentation__.Notify(113219)
	if !isDeltaTrusted {
		__antithesis_instrumentation__.Notify(113224)
		r.mu.raftLogSizeTrusted = false
	} else {
		__antithesis_instrumentation__.Notify(113225)
	}
}

func (r *raftTruncatorReplica) getPendingTruncs() *pendingLogTruncations {
	__antithesis_instrumentation__.Notify(113226)
	return &r.pendingLogTruncations
}

func (r *raftTruncatorReplica) sideloadedBytesIfTruncatedFromTo(
	ctx context.Context, from, to uint64,
) (freed int64, err error) {
	__antithesis_instrumentation__.Notify(113227)
	freed, _, err = r.raftMu.sideloaded.BytesIfTruncatedFromTo(ctx, from, to)
	return freed, err
}

func (r *raftTruncatorReplica) getStateLoader() stateloader.StateLoader {
	__antithesis_instrumentation__.Notify(113228)

	return r.raftMu.stateLoader
}
