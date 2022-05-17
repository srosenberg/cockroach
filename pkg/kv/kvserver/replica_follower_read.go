package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

var FollowerReadsEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.closed_timestamp.follower_reads_enabled",
	"allow (all) replicas to serve consistent historical reads based on closed timestamp information",
	true,
).WithPublic()

func BatchCanBeEvaluatedOnFollower(ba roachpb.BatchRequest) bool {
	__antithesis_instrumentation__.Notify(117417)

	tsFromServerClock := ba.Txn == nil && func() bool {
		__antithesis_instrumentation__.Notify(117419)
		return (ba.Timestamp.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(117420)
			return ba.TimestampFromServerClock != nil == true
		}() == true) == true
	}() == true
	if tsFromServerClock {
		__antithesis_instrumentation__.Notify(117421)
		return false
	} else {
		__antithesis_instrumentation__.Notify(117422)
	}
	__antithesis_instrumentation__.Notify(117418)
	return ba.IsAllTransactional() && func() bool {
		__antithesis_instrumentation__.Notify(117423)
		return ba.IsReadOnly() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(117424)
		return !ba.IsLocking() == true
	}() == true
}

func (r *Replica) canServeFollowerReadRLocked(ctx context.Context, ba *roachpb.BatchRequest) bool {
	__antithesis_instrumentation__.Notify(117425)
	eligible := BatchCanBeEvaluatedOnFollower(*ba) && func() bool {
		__antithesis_instrumentation__.Notify(117430)
		return FollowerReadsEnabled.Get(&r.store.cfg.Settings.SV) == true
	}() == true
	if !eligible {
		__antithesis_instrumentation__.Notify(117431)

		return false
	} else {
		__antithesis_instrumentation__.Notify(117432)
	}
	__antithesis_instrumentation__.Notify(117426)

	repDesc, err := r.getReplicaDescriptorRLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(117433)
		return false
	} else {
		__antithesis_instrumentation__.Notify(117434)
	}
	__antithesis_instrumentation__.Notify(117427)

	switch typ := repDesc.GetType(); typ {
	case roachpb.VOTER_FULL, roachpb.VOTER_INCOMING, roachpb.NON_VOTER:
		__antithesis_instrumentation__.Notify(117435)
	default:
		__antithesis_instrumentation__.Notify(117436)
		log.Eventf(ctx, "%s replicas cannot serve follower reads", typ)
		return false
	}
	__antithesis_instrumentation__.Notify(117428)

	requiredFrontier := ba.RequiredFrontier()
	maxClosed := r.getClosedTimestampRLocked(ctx, requiredFrontier)
	canServeFollowerRead := requiredFrontier.LessEq(maxClosed)
	tsDiff := requiredFrontier.GoTime().Sub(maxClosed.GoTime())
	if !canServeFollowerRead {
		__antithesis_instrumentation__.Notify(117437)
		uncertaintyLimitStr := "n/a"
		if ba.Txn != nil {
			__antithesis_instrumentation__.Notify(117439)
			uncertaintyLimitStr = ba.Txn.GlobalUncertaintyLimit.String()
		} else {
			__antithesis_instrumentation__.Notify(117440)
		}
		__antithesis_instrumentation__.Notify(117438)

		log.Eventf(ctx, "can't serve follower read; closed timestamp too low by: %s; maxClosed: %s ts: %s uncertaintyLimit: %s",
			tsDiff, maxClosed, ba.Timestamp, uncertaintyLimitStr)
		return false
	} else {
		__antithesis_instrumentation__.Notify(117441)
	}
	__antithesis_instrumentation__.Notify(117429)

	log.Eventf(ctx, "%s; query timestamp below closed timestamp by %s", kvbase.FollowerReadServingMsg, -tsDiff)
	r.store.metrics.FollowerReadsCount.Inc(1)
	return true
}

func (r *Replica) getClosedTimestampRLocked(
	ctx context.Context, sufficient hlc.Timestamp,
) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(117442)
	appliedLAI := ctpb.LAI(r.mu.state.LeaseAppliedIndex)
	leaseholder := r.mu.state.Lease.Replica.NodeID
	raftClosed := r.mu.state.RaftClosedTimestamp
	sideTransportClosed := r.sideTransportClosedTimestamp.get(ctx, leaseholder, appliedLAI, sufficient)

	var maxClosed hlc.Timestamp
	maxClosed.Forward(raftClosed)
	maxClosed.Forward(sideTransportClosed)
	return maxClosed
}

func (r *Replica) GetClosedTimestamp(ctx context.Context) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(117443)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getClosedTimestampRLocked(ctx, hlc.Timestamp{})
}
