package physicalplan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

const avgRangesPerNode = 5

type fakeSpanResolver struct {
	nodes []*roachpb.NodeDescriptor
}

var _ SpanResolver = &fakeSpanResolver{}

func NewFakeSpanResolver(nodes []*roachpb.NodeDescriptor) SpanResolver {
	__antithesis_instrumentation__.Notify(562121)
	return &fakeSpanResolver{
		nodes: nodes,
	}
}

type fakeRange struct {
	startKey roachpb.Key
	endKey   roachpb.Key
	replica  *roachpb.NodeDescriptor
}

type fakeSpanResolverIterator struct {
	fsr *fakeSpanResolver

	db  *kv.DB
	err error

	ranges []fakeRange
}

func (fsr *fakeSpanResolver) NewSpanResolverIterator(txn *kv.Txn) SpanResolverIterator {
	__antithesis_instrumentation__.Notify(562122)
	return &fakeSpanResolverIterator{fsr: fsr, db: txn.DB()}
}

func (fit *fakeSpanResolverIterator) Seek(
	ctx context.Context, span roachpb.Span, scanDir kvcoord.ScanDirection,
) {
	__antithesis_instrumentation__.Notify(562123)

	var prevRange fakeRange
	if fit.ranges != nil {
		__antithesis_instrumentation__.Notify(562132)
		prevRange = fit.ranges[len(fit.ranges)-1]
	} else {
		__antithesis_instrumentation__.Notify(562133)
	}
	__antithesis_instrumentation__.Notify(562124)

	var b kv.Batch
	b.Header.ReadConsistency = roachpb.READ_UNCOMMITTED
	b.Scan(span.Key, span.EndKey)
	err := fit.db.Run(ctx, &b)
	if err != nil {
		__antithesis_instrumentation__.Notify(562134)
		log.Errorf(ctx, "error in fake span resolver scan: %s", err)
		fit.err = err
		return
	} else {
		__antithesis_instrumentation__.Notify(562135)
	}
	__antithesis_instrumentation__.Notify(562125)
	kvs := b.Results[0].Rows

	var splitKeys []roachpb.Key
	lastKey := span.Key
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(562136)

		splitKey, err := keys.EnsureSafeSplitKey(kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(562138)
			fit.err = err
			return
		} else {
			__antithesis_instrumentation__.Notify(562139)
		}
		__antithesis_instrumentation__.Notify(562137)
		if !splitKey.Equal(lastKey) && func() bool {
			__antithesis_instrumentation__.Notify(562140)
			return span.ContainsKey(splitKey) == true
		}() == true {
			__antithesis_instrumentation__.Notify(562141)
			splitKeys = append(splitKeys, splitKey)
			lastKey = splitKey
		} else {
			__antithesis_instrumentation__.Notify(562142)
		}
	}
	__antithesis_instrumentation__.Notify(562126)

	maxSplits := 2 * len(fit.fsr.nodes) * avgRangesPerNode
	if maxSplits > len(splitKeys) {
		__antithesis_instrumentation__.Notify(562143)
		maxSplits = len(splitKeys)
	} else {
		__antithesis_instrumentation__.Notify(562144)
	}
	__antithesis_instrumentation__.Notify(562127)
	numSplits := rand.Intn(maxSplits + 1)

	chosen := make(map[int]struct{})
	for j := len(splitKeys) - numSplits; j < len(splitKeys); j++ {
		__antithesis_instrumentation__.Notify(562145)
		t := rand.Intn(j + 1)
		if _, alreadyChosen := chosen[t]; !alreadyChosen {
			__antithesis_instrumentation__.Notify(562146)

			chosen[t] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(562147)

			chosen[j] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(562128)

	splits := make([]roachpb.Key, 0, numSplits+2)
	splits = append(splits, span.Key)
	for i := range splitKeys {
		__antithesis_instrumentation__.Notify(562148)
		if _, ok := chosen[i]; ok {
			__antithesis_instrumentation__.Notify(562149)
			splits = append(splits, splitKeys[i])
		} else {
			__antithesis_instrumentation__.Notify(562150)
		}
	}
	__antithesis_instrumentation__.Notify(562129)
	splits = append(splits, span.EndKey)

	if scanDir == kvcoord.Descending {
		__antithesis_instrumentation__.Notify(562151)

		for i := 0; i < len(splits)/2; i++ {
			__antithesis_instrumentation__.Notify(562152)
			j := len(splits) - i - 1
			splits[i], splits[j] = splits[j], splits[i]
		}
	} else {
		__antithesis_instrumentation__.Notify(562153)
	}
	__antithesis_instrumentation__.Notify(562130)

	fit.ranges = make([]fakeRange, len(splits)-1)
	for i := range fit.ranges {
		__antithesis_instrumentation__.Notify(562154)
		fit.ranges[i] = fakeRange{
			startKey: splits[i],
			endKey:   splits[i+1],
			replica:  fit.fsr.nodes[rand.Intn(len(fit.fsr.nodes))],
		}
	}
	__antithesis_instrumentation__.Notify(562131)

	if prevRange.endKey != nil {
		__antithesis_instrumentation__.Notify(562155)
		prefix, err := keys.EnsureSafeSplitKey(span.Key)

		if err == nil && func() bool {
			__antithesis_instrumentation__.Notify(562156)
			return len(prevRange.endKey) >= len(prefix) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(562157)
			return bytes.Equal(prefix, prevRange.endKey[:len(prefix)]) == true
		}() == true {
			__antithesis_instrumentation__.Notify(562158)
			fit.ranges[0].replica = prevRange.replica
		} else {
			__antithesis_instrumentation__.Notify(562159)
		}
	} else {
		__antithesis_instrumentation__.Notify(562160)
	}
}

func (fit *fakeSpanResolverIterator) Valid() bool {
	__antithesis_instrumentation__.Notify(562161)
	return fit.err == nil
}

func (fit *fakeSpanResolverIterator) Error() error {
	__antithesis_instrumentation__.Notify(562162)
	return fit.err
}

func (fit *fakeSpanResolverIterator) NeedAnother() bool {
	__antithesis_instrumentation__.Notify(562163)
	return len(fit.ranges) > 1
}

func (fit *fakeSpanResolverIterator) Next(_ context.Context) {
	__antithesis_instrumentation__.Notify(562164)
	if len(fit.ranges) <= 1 {
		__antithesis_instrumentation__.Notify(562166)
		panic("Next called with no more ranges")
	} else {
		__antithesis_instrumentation__.Notify(562167)
	}
	__antithesis_instrumentation__.Notify(562165)
	fit.ranges = fit.ranges[1:]
}

func (fit *fakeSpanResolverIterator) Desc() roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(562168)
	return roachpb.RangeDescriptor{
		StartKey: roachpb.RKey(fit.ranges[0].startKey),
		EndKey:   roachpb.RKey(fit.ranges[0].endKey),
	}
}

func (fit *fakeSpanResolverIterator) ReplicaInfo(
	_ context.Context,
) (roachpb.ReplicaDescriptor, error) {
	__antithesis_instrumentation__.Notify(562169)
	n := fit.ranges[0].replica
	return roachpb.ReplicaDescriptor{NodeID: n.NodeID}, nil
}
