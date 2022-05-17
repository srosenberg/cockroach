// Package replicaoracle provides functionality for physicalplan to choose a
// replica for a range.
package replicaoracle

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

type Policy byte

var (
	RandomChoice = RegisterPolicy(newRandomOracle)

	BinPackingChoice = RegisterPolicy(newBinPackingOracle)

	ClosestChoice = RegisterPolicy(newClosestOracle)
)

type Config struct {
	NodeDescs  kvcoord.NodeDescStore
	NodeDesc   roachpb.NodeDescriptor
	Settings   *cluster.Settings
	RPCContext *rpc.Context
}

type Oracle interface {
	ChoosePreferredReplica(
		ctx context.Context,
		txn *kv.Txn,
		rng *roachpb.RangeDescriptor,
		leaseholder *roachpb.ReplicaDescriptor,
		ctPolicy roachpb.RangeClosedTimestampPolicy,
		qState QueryState,
	) (roachpb.ReplicaDescriptor, error)
}

type OracleFactory func(Config) Oracle

func NewOracle(policy Policy, cfg Config) Oracle {
	__antithesis_instrumentation__.Notify(562445)
	ff, ok := oracleFactories[policy]
	if !ok {
		__antithesis_instrumentation__.Notify(562447)
		panic(errors.Errorf("unknown Policy %v", policy))
	} else {
		__antithesis_instrumentation__.Notify(562448)
	}
	__antithesis_instrumentation__.Notify(562446)
	return ff(cfg)
}

func RegisterPolicy(f OracleFactory) Policy {
	__antithesis_instrumentation__.Notify(562449)
	if len(oracleFactories) == 255 {
		__antithesis_instrumentation__.Notify(562451)
		panic("Can only register 255 Policy instances")
	} else {
		__antithesis_instrumentation__.Notify(562452)
	}
	__antithesis_instrumentation__.Notify(562450)
	r := Policy(len(oracleFactories))
	oracleFactories[r] = f
	return r
}

var oracleFactories = map[Policy]OracleFactory{}

type QueryState struct {
	RangesPerNode  util.FastIntMap
	AssignedRanges map[roachpb.RangeID]roachpb.ReplicaDescriptor
}

func MakeQueryState() QueryState {
	__antithesis_instrumentation__.Notify(562453)
	return QueryState{
		AssignedRanges: make(map[roachpb.RangeID]roachpb.ReplicaDescriptor),
	}
}

type randomOracle struct {
	nodeDescs kvcoord.NodeDescStore
}

func newRandomOracle(cfg Config) Oracle {
	__antithesis_instrumentation__.Notify(562454)
	return &randomOracle{nodeDescs: cfg.NodeDescs}
}

func (o *randomOracle) ChoosePreferredReplica(
	ctx context.Context,
	_ *kv.Txn,
	desc *roachpb.RangeDescriptor,
	_ *roachpb.ReplicaDescriptor,
	_ roachpb.RangeClosedTimestampPolicy,
	_ QueryState,
) (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(562455)
	replicas, err := replicaSliceOrErr(ctx, o.nodeDescs, desc, kvcoord.OnlyPotentialLeaseholders)
	if err != nil {
		__antithesis_instrumentation__.Notify(562457)
		return roachpb.ReplicaDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(562458)
	}
	__antithesis_instrumentation__.Notify(562456)
	return replicas[rand.Intn(len(replicas))].ReplicaDescriptor, nil
}

type closestOracle struct {
	nodeDescs kvcoord.NodeDescStore

	nodeDesc    roachpb.NodeDescriptor
	latencyFunc kvcoord.LatencyFunc
}

func newClosestOracle(cfg Config) Oracle {
	__antithesis_instrumentation__.Notify(562459)
	return &closestOracle{
		nodeDescs:   cfg.NodeDescs,
		nodeDesc:    cfg.NodeDesc,
		latencyFunc: latencyFunc(cfg.RPCContext),
	}
}

func (o *closestOracle) ChoosePreferredReplica(
	ctx context.Context,
	_ *kv.Txn,
	desc *roachpb.RangeDescriptor,
	_ *roachpb.ReplicaDescriptor,
	_ roachpb.RangeClosedTimestampPolicy,
	_ QueryState,
) (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(562460)

	replicas, err := replicaSliceOrErr(ctx, o.nodeDescs, desc, kvcoord.AllExtantReplicas)
	if err != nil {
		__antithesis_instrumentation__.Notify(562462)
		return roachpb.ReplicaDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(562463)
	}
	__antithesis_instrumentation__.Notify(562461)
	replicas.OptimizeReplicaOrder(&o.nodeDesc, o.latencyFunc)
	return replicas[0].ReplicaDescriptor, nil
}

const maxPreferredRangesPerLeaseHolder = 10

type binPackingOracle struct {
	maxPreferredRangesPerLeaseHolder int
	nodeDescs                        kvcoord.NodeDescStore

	nodeDesc    roachpb.NodeDescriptor
	latencyFunc kvcoord.LatencyFunc
}

func newBinPackingOracle(cfg Config) Oracle {
	__antithesis_instrumentation__.Notify(562464)
	return &binPackingOracle{
		maxPreferredRangesPerLeaseHolder: maxPreferredRangesPerLeaseHolder,
		nodeDescs:                        cfg.NodeDescs,
		nodeDesc:                         cfg.NodeDesc,
		latencyFunc:                      latencyFunc(cfg.RPCContext),
	}
}

func (o *binPackingOracle) ChoosePreferredReplica(
	ctx context.Context,
	_ *kv.Txn,
	desc *roachpb.RangeDescriptor,
	leaseholder *roachpb.ReplicaDescriptor,
	_ roachpb.RangeClosedTimestampPolicy,
	queryState QueryState,
) (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(562465)

	if leaseholder != nil {
		__antithesis_instrumentation__.Notify(562469)
		return *leaseholder, nil
	} else {
		__antithesis_instrumentation__.Notify(562470)
	}
	__antithesis_instrumentation__.Notify(562466)

	replicas, err := replicaSliceOrErr(ctx, o.nodeDescs, desc, kvcoord.OnlyPotentialLeaseholders)
	if err != nil {
		__antithesis_instrumentation__.Notify(562471)
		return roachpb.ReplicaDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(562472)
	}
	__antithesis_instrumentation__.Notify(562467)
	replicas.OptimizeReplicaOrder(&o.nodeDesc, o.latencyFunc)

	minLoad := int(math.MaxInt32)
	var leastLoadedIdx int
	for i, repl := range replicas {
		__antithesis_instrumentation__.Notify(562473)
		assignedRanges := queryState.RangesPerNode.GetDefault(int(repl.NodeID))
		if assignedRanges != 0 && func() bool {
			__antithesis_instrumentation__.Notify(562475)
			return assignedRanges < o.maxPreferredRangesPerLeaseHolder == true
		}() == true {
			__antithesis_instrumentation__.Notify(562476)
			return repl.ReplicaDescriptor, nil
		} else {
			__antithesis_instrumentation__.Notify(562477)
		}
		__antithesis_instrumentation__.Notify(562474)
		if assignedRanges < minLoad {
			__antithesis_instrumentation__.Notify(562478)
			leastLoadedIdx = i
			minLoad = assignedRanges
		} else {
			__antithesis_instrumentation__.Notify(562479)
		}
	}
	__antithesis_instrumentation__.Notify(562468)

	return replicas[leastLoadedIdx].ReplicaDescriptor, nil
}

func replicaSliceOrErr(
	ctx context.Context,
	nodeDescs kvcoord.NodeDescStore,
	desc *roachpb.RangeDescriptor,
	filter kvcoord.ReplicaSliceFilter,
) (kvcoord.ReplicaSlice, error) {
	__antithesis_instrumentation__.Notify(562480)
	replicas, err := kvcoord.NewReplicaSlice(ctx, nodeDescs, desc, nil, filter)
	if err != nil {
		__antithesis_instrumentation__.Notify(562482)
		return kvcoord.ReplicaSlice{}, sqlerrors.NewRangeUnavailableError(desc.RangeID, err)
	} else {
		__antithesis_instrumentation__.Notify(562483)
	}
	__antithesis_instrumentation__.Notify(562481)
	return replicas, nil
}

func latencyFunc(rpcCtx *rpc.Context) kvcoord.LatencyFunc {
	__antithesis_instrumentation__.Notify(562484)
	if rpcCtx != nil {
		__antithesis_instrumentation__.Notify(562486)
		return rpcCtx.RemoteClocks.Latency
	} else {
		__antithesis_instrumentation__.Notify(562487)
	}
	__antithesis_instrumentation__.Notify(562485)
	return nil
}
