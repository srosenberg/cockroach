package rangefeedbuffer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

func RangeFeedValueEventToKV(event Event) roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(89880)
	rfv := event.(*roachpb.RangeFeedValue)
	return roachpb.KeyValue{Key: rfv.Key, Value: rfv.Value}
}

func EventsToKVs(events []Event, f func(ev Event) roachpb.KeyValue) []roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(89881)
	kvs := make([]roachpb.KeyValue, 0, len(events))
	for _, ev := range events {
		__antithesis_instrumentation__.Notify(89883)
		kvs = append(kvs, f(ev))
	}
	__antithesis_instrumentation__.Notify(89882)
	return kvs
}

func MergeKVs(base, updates []roachpb.KeyValue) []roachpb.KeyValue {
	__antithesis_instrumentation__.Notify(89884)
	if len(updates) == 0 {
		__antithesis_instrumentation__.Notify(89888)
		return base
	} else {
		__antithesis_instrumentation__.Notify(89889)
	}
	__antithesis_instrumentation__.Notify(89885)
	combined := make([]roachpb.KeyValue, 0, len(base)+len(updates))
	combined = append(append(combined, base...), updates...)
	sort.Slice(combined, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(89890)
		cmp := combined[i].Key.Compare(combined[j].Key)
		if cmp == 0 {
			__antithesis_instrumentation__.Notify(89892)
			return combined[i].Value.Timestamp.Less(combined[j].Value.Timestamp)
		} else {
			__antithesis_instrumentation__.Notify(89893)
		}
		__antithesis_instrumentation__.Notify(89891)
		return cmp < 0
	})
	__antithesis_instrumentation__.Notify(89886)
	r := combined[:0]
	for _, kv := range combined {
		__antithesis_instrumentation__.Notify(89894)
		prevIsSameKey := len(r) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(89895)
			return r[len(r)-1].Key.Equal(kv.Key) == true
		}() == true
		if kv.Value.IsPresent() {
			__antithesis_instrumentation__.Notify(89896)
			if prevIsSameKey {
				__antithesis_instrumentation__.Notify(89897)
				r[len(r)-1] = kv
			} else {
				__antithesis_instrumentation__.Notify(89898)
				r = append(r, kv)
			}
		} else {
			__antithesis_instrumentation__.Notify(89899)
			if prevIsSameKey {
				__antithesis_instrumentation__.Notify(89900)
				r = r[:len(r)-1]
			} else {
				__antithesis_instrumentation__.Notify(89901)
			}
		}
	}
	__antithesis_instrumentation__.Notify(89887)
	return r
}
