package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

const maxDelaySplitTriggerDur = 20 * time.Second

type replicaMsgAppDropper Replica

func (rd *replicaMsgAppDropper) Args() (initialized bool, age time.Duration) {
	__antithesis_instrumentation__.Notify(123551)
	r := (*Replica)(rd)
	r.mu.RLock()
	initialized = r.isInitializedRLocked()
	creationTime := r.creationTime
	r.mu.RUnlock()
	age = timeutil.Since(creationTime)
	return initialized, age
}

func (rd *replicaMsgAppDropper) ShouldDrop(
	ctx context.Context, startKey roachpb.RKey,
) (fmt.Stringer, bool) {
	__antithesis_instrumentation__.Notify(123552)
	lhsRepl := (*Replica)(rd).store.LookupReplica(startKey)
	if lhsRepl == nil {
		__antithesis_instrumentation__.Notify(123554)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(123555)
	}
	__antithesis_instrumentation__.Notify(123553)
	lhsRepl.store.replicaGCQueue.AddAsync(ctx, lhsRepl, replicaGCPriorityDefault)
	return lhsRepl, true
}

type msgAppDropper interface {
	Args() (initialized bool, age time.Duration)
	ShouldDrop(ctx context.Context, key roachpb.RKey) (fmt.Stringer, bool)
}

func maybeDropMsgApp(
	ctx context.Context, r msgAppDropper, msg *raftpb.Message, startKey roachpb.RKey,
) (drop bool) {
	__antithesis_instrumentation__.Notify(123556)

	if msg.Type != raftpb.MsgApp || func() bool {
		__antithesis_instrumentation__.Notify(123563)
		return startKey == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(123564)
		return false
	} else {
		__antithesis_instrumentation__.Notify(123565)
	}
	__antithesis_instrumentation__.Notify(123557)

	initialized, age := r.Args()

	if initialized {
		__antithesis_instrumentation__.Notify(123566)
		return false
	} else {
		__antithesis_instrumentation__.Notify(123567)
	}
	__antithesis_instrumentation__.Notify(123558)

	verbose := verboseRaftLoggingEnabled()

	lhsRepl, drop := r.ShouldDrop(ctx, startKey)
	if !drop {
		__antithesis_instrumentation__.Notify(123568)
		return false
	} else {
		__antithesis_instrumentation__.Notify(123569)
	}
	__antithesis_instrumentation__.Notify(123559)

	if verbose {
		__antithesis_instrumentation__.Notify(123570)
		log.Infof(ctx, "start key is contained in replica %v", lhsRepl)
	} else {
		__antithesis_instrumentation__.Notify(123571)
	}
	__antithesis_instrumentation__.Notify(123560)
	if age > maxDelaySplitTriggerDur {
		__antithesis_instrumentation__.Notify(123572)

		log.Warningf(
			ctx,
			"would have dropped incoming MsgApp to wait for split trigger, "+
				"but allowing because uninitialized replica was created %s (>%s) ago",
			age, maxDelaySplitTriggerDur)
		return false
	} else {
		__antithesis_instrumentation__.Notify(123573)
	}
	__antithesis_instrumentation__.Notify(123561)
	if verbose {
		__antithesis_instrumentation__.Notify(123574)
		log.Infof(ctx, "dropping MsgApp at index %d to wait for split trigger", msg.Index)
	} else {
		__antithesis_instrumentation__.Notify(123575)
	}
	__antithesis_instrumentation__.Notify(123562)
	return true
}
