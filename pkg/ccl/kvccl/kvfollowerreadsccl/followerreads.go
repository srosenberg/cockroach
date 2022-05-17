// Package kvfollowerreadsccl implements and injects the functionality needed to
// expose follower reads to clients.
package kvfollowerreadsccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan/replicaoracle"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

var ClosedTimestampPropagationSlack = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.closed_timestamp.propagation_slack",
	"a conservative estimate of the amount of time expect for closed timestamps to "+
		"propagate from a leaseholder to followers. This is taken into account by "+
		"follower_read_timestamp().",
	time.Second,
	settings.NonNegativeDuration,
)

func getFollowerReadLag(st *cluster.Settings) time.Duration {
	__antithesis_instrumentation__.Notify(19610)
	targetDuration := closedts.TargetDuration.Get(&st.SV)
	sideTransportInterval := closedts.SideTransportCloseInterval.Get(&st.SV)
	slack := ClosedTimestampPropagationSlack.Get(&st.SV)

	if targetDuration == 0 {
		__antithesis_instrumentation__.Notify(19612)

		return math.MinInt64
	} else {
		__antithesis_instrumentation__.Notify(19613)
	}
	__antithesis_instrumentation__.Notify(19611)
	return -targetDuration - sideTransportInterval - slack
}

func getGlobalReadsLead(clock *hlc.Clock) time.Duration {
	__antithesis_instrumentation__.Notify(19614)
	return clock.MaxOffset()
}

func checkEnterpriseEnabled(logicalClusterID uuid.UUID, st *cluster.Settings) error {
	__antithesis_instrumentation__.Notify(19615)
	org := sql.ClusterOrganization.Get(&st.SV)
	return utilccl.CheckEnterpriseEnabled(st, logicalClusterID, org, "follower reads")
}

func isEnterpriseEnabled(logicalClusterID uuid.UUID, st *cluster.Settings) bool {
	__antithesis_instrumentation__.Notify(19616)
	org := sql.ClusterOrganization.Get(&st.SV)
	return utilccl.IsEnterpriseEnabled(st, logicalClusterID, org, "follower reads")
}

func checkFollowerReadsEnabled(logicalClusterID uuid.UUID, st *cluster.Settings) bool {
	__antithesis_instrumentation__.Notify(19617)
	if !kvserver.FollowerReadsEnabled.Get(&st.SV) {
		__antithesis_instrumentation__.Notify(19619)
		return false
	} else {
		__antithesis_instrumentation__.Notify(19620)
	}
	__antithesis_instrumentation__.Notify(19618)
	return isEnterpriseEnabled(logicalClusterID, st)
}

func evalFollowerReadOffset(
	logicalClusterID uuid.UUID, st *cluster.Settings,
) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(19621)
	if err := checkEnterpriseEnabled(logicalClusterID, st); err != nil {
		__antithesis_instrumentation__.Notify(19623)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(19624)
	}
	__antithesis_instrumentation__.Notify(19622)

	return getFollowerReadLag(st), nil
}

func closedTimestampLikelySufficient(
	st *cluster.Settings,
	clock *hlc.Clock,
	ctPolicy roachpb.RangeClosedTimestampPolicy,
	requiredFrontierTS hlc.Timestamp,
) bool {
	__antithesis_instrumentation__.Notify(19625)
	var offset time.Duration
	switch ctPolicy {
	case roachpb.LAG_BY_CLUSTER_SETTING:
		__antithesis_instrumentation__.Notify(19627)
		offset = getFollowerReadLag(st)
	case roachpb.LEAD_FOR_GLOBAL_READS:
		__antithesis_instrumentation__.Notify(19628)
		offset = getGlobalReadsLead(clock)
	default:
		__antithesis_instrumentation__.Notify(19629)
		panic("unknown RangeClosedTimestampPolicy")
	}
	__antithesis_instrumentation__.Notify(19626)
	expectedClosedTS := clock.Now().Add(offset.Nanoseconds(), 0)
	return requiredFrontierTS.LessEq(expectedClosedTS)
}

func canSendToFollower(
	logicalClusterID uuid.UUID,
	st *cluster.Settings,
	clock *hlc.Clock,
	ctPolicy roachpb.RangeClosedTimestampPolicy,
	ba roachpb.BatchRequest,
) bool {
	__antithesis_instrumentation__.Notify(19630)
	return kvserver.BatchCanBeEvaluatedOnFollower(ba) && func() bool {
		__antithesis_instrumentation__.Notify(19631)
		return closedTimestampLikelySufficient(st, clock, ctPolicy, ba.RequiredFrontier()) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(19632)
		return checkFollowerReadsEnabled(logicalClusterID, st) == true
	}() == true
}

type followerReadOracle struct {
	logicalClusterID *base.ClusterIDContainer
	st               *cluster.Settings
	clock            *hlc.Clock

	closest    replicaoracle.Oracle
	binPacking replicaoracle.Oracle
}

func newFollowerReadOracle(cfg replicaoracle.Config) replicaoracle.Oracle {
	__antithesis_instrumentation__.Notify(19633)
	return &followerReadOracle{
		logicalClusterID: cfg.RPCContext.LogicalClusterID,
		st:               cfg.Settings,
		clock:            cfg.RPCContext.Clock,
		closest:          replicaoracle.NewOracle(replicaoracle.ClosestChoice, cfg),
		binPacking:       replicaoracle.NewOracle(replicaoracle.BinPackingChoice, cfg),
	}
}

func (o *followerReadOracle) ChoosePreferredReplica(
	ctx context.Context,
	txn *kv.Txn,
	desc *roachpb.RangeDescriptor,
	leaseholder *roachpb.ReplicaDescriptor,
	ctPolicy roachpb.RangeClosedTimestampPolicy,
	queryState replicaoracle.QueryState,
) (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(19634)
	var oracle replicaoracle.Oracle
	if o.useClosestOracle(txn, ctPolicy) {
		__antithesis_instrumentation__.Notify(19636)
		oracle = o.closest
	} else {
		__antithesis_instrumentation__.Notify(19637)
		oracle = o.binPacking
	}
	__antithesis_instrumentation__.Notify(19635)
	return oracle.ChoosePreferredReplica(ctx, txn, desc, leaseholder, ctPolicy, queryState)
}

func (o *followerReadOracle) useClosestOracle(
	txn *kv.Txn, ctPolicy roachpb.RangeClosedTimestampPolicy,
) bool {
	__antithesis_instrumentation__.Notify(19638)

	return txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(19639)
		return closedTimestampLikelySufficient(o.st, o.clock, ctPolicy, txn.RequiredFrontier()) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(19640)
		return checkFollowerReadsEnabled(o.logicalClusterID.Get(), o.st) == true
	}() == true
}

var followerReadOraclePolicy = replicaoracle.RegisterPolicy(newFollowerReadOracle)

func init() {
	sql.ReplicaOraclePolicy = followerReadOraclePolicy
	builtins.EvalFollowerReadOffset = evalFollowerReadOffset
	kvcoord.CanSendToFollower = canSendToFollower
}
