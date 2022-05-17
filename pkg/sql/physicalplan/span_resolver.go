package physicalplan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan/replicaoracle"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type SpanResolver interface {
	NewSpanResolverIterator(txn *kv.Txn) SpanResolverIterator
}

type SpanResolverIterator interface {
	Seek(ctx context.Context, span roachpb.Span, scanDir kvcoord.ScanDirection)

	NeedAnother() bool

	Next(ctx context.Context)

	Valid() bool

	Error() error

	Desc() roachpb.RangeDescriptor

	ReplicaInfo(ctx context.Context) (roachpb.ReplicaDescriptor, error)
}

type spanResolver struct {
	st         *cluster.Settings
	distSender *kvcoord.DistSender
	nodeDesc   roachpb.NodeDescriptor
	oracle     replicaoracle.Oracle
}

var _ SpanResolver = &spanResolver{}

func NewSpanResolver(
	st *cluster.Settings,
	distSender *kvcoord.DistSender,
	nodeDescs kvcoord.NodeDescStore,
	nodeDesc roachpb.NodeDescriptor,
	rpcCtx *rpc.Context,
	policy replicaoracle.Policy,
) SpanResolver {
	__antithesis_instrumentation__.Notify(562488)
	return &spanResolver{
		st:       st,
		nodeDesc: nodeDesc,
		oracle: replicaoracle.NewOracle(policy, replicaoracle.Config{
			NodeDescs:  nodeDescs,
			NodeDesc:   nodeDesc,
			Settings:   st,
			RPCContext: rpcCtx,
		}),
		distSender: distSender,
	}
}

type spanResolverIterator struct {
	txn *kv.Txn

	it kvcoord.RangeIterator

	oracle replicaoracle.Oracle

	curSpan roachpb.RSpan

	dir kvcoord.ScanDirection

	queryState replicaoracle.QueryState

	err error
}

var _ SpanResolverIterator = &spanResolverIterator{}

func (sr *spanResolver) NewSpanResolverIterator(txn *kv.Txn) SpanResolverIterator {
	__antithesis_instrumentation__.Notify(562489)
	return &spanResolverIterator{
		txn:        txn,
		it:         kvcoord.MakeRangeIterator(sr.distSender),
		oracle:     sr.oracle,
		queryState: replicaoracle.MakeQueryState(),
	}
}

func (it *spanResolverIterator) Valid() bool {
	__antithesis_instrumentation__.Notify(562490)
	return it.err == nil && func() bool {
		__antithesis_instrumentation__.Notify(562491)
		return it.it.Valid() == true
	}() == true
}

func (it *spanResolverIterator) Error() error {
	__antithesis_instrumentation__.Notify(562492)
	if it.err != nil {
		__antithesis_instrumentation__.Notify(562494)
		return it.err
	} else {
		__antithesis_instrumentation__.Notify(562495)
	}
	__antithesis_instrumentation__.Notify(562493)
	return it.it.Error()
}

func (it *spanResolverIterator) Seek(
	ctx context.Context, span roachpb.Span, scanDir kvcoord.ScanDirection,
) {
	__antithesis_instrumentation__.Notify(562496)
	rSpan, err := keys.SpanAddr(span)
	if err != nil {
		__antithesis_instrumentation__.Notify(562501)
		it.err = err
		return
	} else {
		__antithesis_instrumentation__.Notify(562502)
	}
	__antithesis_instrumentation__.Notify(562497)

	oldDir := it.dir
	it.curSpan = rSpan
	it.dir = scanDir

	var seekKey roachpb.RKey
	if scanDir == kvcoord.Ascending {
		__antithesis_instrumentation__.Notify(562503)
		seekKey = it.curSpan.Key
	} else {
		__antithesis_instrumentation__.Notify(562504)
		seekKey = it.curSpan.EndKey
	}
	__antithesis_instrumentation__.Notify(562498)

	if it.dir == oldDir && func() bool {
		__antithesis_instrumentation__.Notify(562505)
		return it.it.Valid() == true
	}() == true {
		__antithesis_instrumentation__.Notify(562506)
		reverse := (it.dir == kvcoord.Descending)
		desc := it.it.Desc()
		if (reverse && func() bool {
			__antithesis_instrumentation__.Notify(562507)
			return desc.ContainsKeyInverted(seekKey) == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(562508)
			return (!reverse && func() bool {
				__antithesis_instrumentation__.Notify(562509)
				return desc.ContainsKey(seekKey) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(562510)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(562512)
				log.Infof(ctx, "not seeking (key=%s); existing descriptor %s", seekKey, desc)
			} else {
				__antithesis_instrumentation__.Notify(562513)
			}
			__antithesis_instrumentation__.Notify(562511)
			return
		} else {
			__antithesis_instrumentation__.Notify(562514)
		}
	} else {
		__antithesis_instrumentation__.Notify(562515)
	}
	__antithesis_instrumentation__.Notify(562499)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(562516)
		log.Infof(ctx, "seeking (key=%s)", seekKey)
	} else {
		__antithesis_instrumentation__.Notify(562517)
	}
	__antithesis_instrumentation__.Notify(562500)
	it.it.Seek(ctx, seekKey, scanDir)
}

func (it *spanResolverIterator) Next(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562518)
	if !it.Valid() {
		__antithesis_instrumentation__.Notify(562520)
		panic(it.Error())
	} else {
		__antithesis_instrumentation__.Notify(562521)
	}
	__antithesis_instrumentation__.Notify(562519)
	it.it.Next(ctx)
}

func (it *spanResolverIterator) NeedAnother() bool {
	__antithesis_instrumentation__.Notify(562522)
	return it.it.NeedAnother(it.curSpan)
}

func (it *spanResolverIterator) Desc() roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(562523)
	return *it.it.Desc()
}

func (it *spanResolverIterator) ReplicaInfo(
	ctx context.Context,
) (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(562524)
	if !it.Valid() {
		__antithesis_instrumentation__.Notify(562528)
		panic(it.Error())
	} else {
		__antithesis_instrumentation__.Notify(562529)
	}
	__antithesis_instrumentation__.Notify(562525)

	rngID := it.it.Desc().RangeID
	if repl, ok := it.queryState.AssignedRanges[rngID]; ok {
		__antithesis_instrumentation__.Notify(562530)
		return repl, nil
	} else {
		__antithesis_instrumentation__.Notify(562531)
	}
	__antithesis_instrumentation__.Notify(562526)

	repl, err := it.oracle.ChoosePreferredReplica(
		ctx, it.txn, it.it.Desc(), it.it.Leaseholder(), it.it.ClosedTimestampPolicy(), it.queryState)
	if err != nil {
		__antithesis_instrumentation__.Notify(562532)
		return roachpb.ReplicaDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(562533)
	}
	__antithesis_instrumentation__.Notify(562527)
	prev := it.queryState.RangesPerNode.GetDefault(int(repl.NodeID))
	it.queryState.RangesPerNode.Set(int(repl.NodeID), prev+1)
	it.queryState.AssignedRanges[rngID] = repl
	return repl, nil
}
