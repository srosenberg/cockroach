package kvclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

func ScanMetaKVs(ctx context.Context, txn *kv.Txn, span roachpb.Span) ([]kv.KeyValue, error) {
	__antithesis_instrumentation__.Notify(90043)
	metaStart := keys.RangeMetaKey(keys.MustAddr(span.Key).Next())
	metaEnd := keys.RangeMetaKey(keys.MustAddr(span.EndKey))

	kvs, err := txn.Scan(ctx, metaStart, metaEnd, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(90046)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(90047)
	}
	__antithesis_instrumentation__.Notify(90044)
	if len(kvs) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(90048)
		return !kvs[len(kvs)-1].Key.Equal(metaEnd.AsRawKey()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(90049)

		extraKV, err := txn.Scan(ctx, metaEnd, keys.Meta2Prefix.PrefixEnd(), 1)
		if err != nil {
			__antithesis_instrumentation__.Notify(90051)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(90052)
		}
		__antithesis_instrumentation__.Notify(90050)
		kvs = append(kvs, extraKV[0])
	} else {
		__antithesis_instrumentation__.Notify(90053)
	}
	__antithesis_instrumentation__.Notify(90045)
	return kvs, nil
}

func GetRangeWithID(
	ctx context.Context, txn *kv.Txn, id roachpb.RangeID,
) (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(90054)

	var ranges []kv.KeyValue
	var err error
	var rangeDesc roachpb.RangeDescriptor
	const chunkSize = 100
	metaStart := keys.RangeMetaKey(keys.MustAddr(keys.MinKey).Next())
	metaEnd := keys.MustAddr(keys.Meta2Prefix.PrefixEnd())
	for {
		__antithesis_instrumentation__.Notify(90056)

		ranges, err = txn.Scan(ctx, metaStart, metaEnd, chunkSize)
		if err != nil {
			__antithesis_instrumentation__.Notify(90060)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(90061)
		}
		__antithesis_instrumentation__.Notify(90057)

		if len(ranges) == 0 {
			__antithesis_instrumentation__.Notify(90062)
			break
		} else {
			__antithesis_instrumentation__.Notify(90063)
		}
		__antithesis_instrumentation__.Notify(90058)
		for _, r := range ranges {
			__antithesis_instrumentation__.Notify(90064)
			if err := r.ValueProto(&rangeDesc); err != nil {
				__antithesis_instrumentation__.Notify(90066)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(90067)
			}
			__antithesis_instrumentation__.Notify(90065)

			if rangeDesc.RangeID == id {
				__antithesis_instrumentation__.Notify(90068)
				return &rangeDesc, nil
			} else {
				__antithesis_instrumentation__.Notify(90069)
			}
		}
		__antithesis_instrumentation__.Notify(90059)

		metaStart = keys.MustAddr(ranges[len(ranges)-1].Key.Next())
	}
	__antithesis_instrumentation__.Notify(90055)
	return nil, nil
}
