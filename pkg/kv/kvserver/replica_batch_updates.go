package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func maybeStripInFlightWrites(ba *roachpb.BatchRequest) (*roachpb.BatchRequest, error) {
	__antithesis_instrumentation__.Notify(115627)
	args, hasET := ba.GetArg(roachpb.EndTxn)
	if !hasET {
		__antithesis_instrumentation__.Notify(115633)
		return ba, nil
	} else {
		__antithesis_instrumentation__.Notify(115634)
	}
	__antithesis_instrumentation__.Notify(115628)

	et := args.(*roachpb.EndTxnRequest)
	otherReqs := ba.Requests[:len(ba.Requests)-1]
	if !et.IsParallelCommit() || func() bool {
		__antithesis_instrumentation__.Notify(115635)
		return len(otherReqs) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(115636)
		return ba, nil
	} else {
		__antithesis_instrumentation__.Notify(115637)
	}
	__antithesis_instrumentation__.Notify(115629)

	origET := et
	etAlloc := new(struct {
		et    roachpb.EndTxnRequest
		union roachpb.RequestUnion_EndTxn
	})
	etAlloc.et = *origET
	etAlloc.union.EndTxn = &etAlloc.et
	et = &etAlloc.et
	et.InFlightWrites = nil
	et.LockSpans = et.LockSpans[:len(et.LockSpans):len(et.LockSpans)]
	ba.Requests = append([]roachpb.RequestUnion(nil), ba.Requests...)
	ba.Requests[len(ba.Requests)-1].Value = &etAlloc.union

	if len(otherReqs) >= len(origET.InFlightWrites) {
		__antithesis_instrumentation__.Notify(115638)
		writes := 0
		for _, ru := range otherReqs {
			__antithesis_instrumentation__.Notify(115640)
			req := ru.GetInner()
			switch {
			case roachpb.IsIntentWrite(req) && func() bool {
				__antithesis_instrumentation__.Notify(115644)
				return !roachpb.IsRange(req) == true
			}() == true:
				__antithesis_instrumentation__.Notify(115641)

				writes++
			case req.Method() == roachpb.QueryIntent:
				__antithesis_instrumentation__.Notify(115642)

				writes++
			default:
				__antithesis_instrumentation__.Notify(115643)

			}
		}
		__antithesis_instrumentation__.Notify(115639)
		if len(origET.InFlightWrites) < writes {
			__antithesis_instrumentation__.Notify(115645)
			return ba, errors.New("more write in batch with EndTxn than listed in in-flight writes")
		} else {
			__antithesis_instrumentation__.Notify(115646)
			if len(origET.InFlightWrites) == writes {
				__antithesis_instrumentation__.Notify(115647)
				et.LockSpans = make([]roachpb.Span, len(origET.LockSpans)+len(origET.InFlightWrites))
				copy(et.LockSpans, origET.LockSpans)
				for i, w := range origET.InFlightWrites {
					__antithesis_instrumentation__.Notify(115649)
					et.LockSpans[len(origET.LockSpans)+i] = roachpb.Span{Key: w.Key}
				}
				__antithesis_instrumentation__.Notify(115648)

				et.LockSpans, ba.Header.DistinctSpans = roachpb.MergeSpans(&et.LockSpans)
				return ba, nil
			} else {
				__antithesis_instrumentation__.Notify(115650)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(115651)
	}
	__antithesis_instrumentation__.Notify(115630)

	copiedTo := 0
	for _, ru := range otherReqs {
		__antithesis_instrumentation__.Notify(115652)
		req := ru.GetInner()
		seq := req.Header().Sequence
		switch {
		case roachpb.IsIntentWrite(req) && func() bool {
			__antithesis_instrumentation__.Notify(115659)
			return !roachpb.IsRange(req) == true
		}() == true:
			__antithesis_instrumentation__.Notify(115656)

		case req.Method() == roachpb.QueryIntent:
			__antithesis_instrumentation__.Notify(115657)

			continue
		default:
			__antithesis_instrumentation__.Notify(115658)

			continue
		}
		__antithesis_instrumentation__.Notify(115653)

		match := -1
		for i, w := range origET.InFlightWrites[copiedTo:] {
			__antithesis_instrumentation__.Notify(115660)
			if w.Sequence == seq {
				__antithesis_instrumentation__.Notify(115661)
				match = i + copiedTo
				break
			} else {
				__antithesis_instrumentation__.Notify(115662)
			}
		}
		__antithesis_instrumentation__.Notify(115654)
		if match == -1 {
			__antithesis_instrumentation__.Notify(115663)
			return ba, errors.New("write in batch with EndTxn missing from in-flight writes")
		} else {
			__antithesis_instrumentation__.Notify(115664)
		}
		__antithesis_instrumentation__.Notify(115655)
		w := origET.InFlightWrites[match]
		notInBa := origET.InFlightWrites[copiedTo:match]
		et.InFlightWrites = append(et.InFlightWrites, notInBa...)
		copiedTo = match + 1

		et.LockSpans = append(et.LockSpans, roachpb.Span{Key: w.Key})
	}
	__antithesis_instrumentation__.Notify(115631)
	if et != origET {
		__antithesis_instrumentation__.Notify(115665)

		notInBa := origET.InFlightWrites[copiedTo:]
		et.InFlightWrites = append(et.InFlightWrites, notInBa...)

		et.LockSpans, ba.Header.DistinctSpans = roachpb.MergeSpans(&et.LockSpans)
	} else {
		__antithesis_instrumentation__.Notify(115666)
	}
	__antithesis_instrumentation__.Notify(115632)
	return ba, nil
}

func maybeBumpReadTimestampToWriteTimestamp(
	ctx context.Context, ba *roachpb.BatchRequest, g *concurrency.Guard,
) bool {
	__antithesis_instrumentation__.Notify(115667)
	if ba.Txn == nil {
		__antithesis_instrumentation__.Notify(115673)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115674)
	}
	__antithesis_instrumentation__.Notify(115668)
	if !ba.CanForwardReadTimestamp {
		__antithesis_instrumentation__.Notify(115675)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115676)
	}
	__antithesis_instrumentation__.Notify(115669)
	if ba.Txn.ReadTimestamp == ba.Txn.WriteTimestamp {
		__antithesis_instrumentation__.Notify(115677)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115678)
	}
	__antithesis_instrumentation__.Notify(115670)
	arg, ok := ba.GetArg(roachpb.EndTxn)
	if !ok {
		__antithesis_instrumentation__.Notify(115679)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115680)
	}
	__antithesis_instrumentation__.Notify(115671)
	et := arg.(*roachpb.EndTxnRequest)
	if batcheval.IsEndTxnExceedingDeadline(ba.Txn.WriteTimestamp, et.Deadline) {
		__antithesis_instrumentation__.Notify(115681)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115682)
	}
	__antithesis_instrumentation__.Notify(115672)
	return tryBumpBatchTimestamp(ctx, ba, g, ba.Txn.WriteTimestamp)
}

func tryBumpBatchTimestamp(
	ctx context.Context, ba *roachpb.BatchRequest, g *concurrency.Guard, ts hlc.Timestamp,
) bool {
	__antithesis_instrumentation__.Notify(115683)
	if g != nil && func() bool {
		__antithesis_instrumentation__.Notify(115688)
		return !g.IsolatedAtLaterTimestamps() == true
	}() == true {
		__antithesis_instrumentation__.Notify(115689)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115690)
	}
	__antithesis_instrumentation__.Notify(115684)
	if ts.Less(ba.Timestamp) {
		__antithesis_instrumentation__.Notify(115691)
		log.Fatalf(ctx, "trying to bump to %s <= ba.Timestamp: %s", ts, ba.Timestamp)
	} else {
		__antithesis_instrumentation__.Notify(115692)
	}
	__antithesis_instrumentation__.Notify(115685)
	if ba.Txn == nil {
		__antithesis_instrumentation__.Notify(115693)
		log.VEventf(ctx, 2, "bumping batch timestamp to %s from %s", ts, ba.Timestamp)
		ba.Timestamp = ts
		return true
	} else {
		__antithesis_instrumentation__.Notify(115694)
	}
	__antithesis_instrumentation__.Notify(115686)
	if ts.Less(ba.Txn.ReadTimestamp) || func() bool {
		__antithesis_instrumentation__.Notify(115695)
		return ts.Less(ba.Txn.WriteTimestamp) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115696)
		log.Fatalf(ctx, "trying to bump to %s inconsistent with ba.Txn.ReadTimestamp: %s, "+
			"ba.Txn.WriteTimestamp: %s", ts, ba.Txn.ReadTimestamp, ba.Txn.WriteTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(115697)
	}
	__antithesis_instrumentation__.Notify(115687)
	log.VEventf(ctx, 2, "bumping batch timestamp to: %s from read: %s, write: %s",
		ts, ba.Txn.ReadTimestamp, ba.Txn.WriteTimestamp)
	ba.Txn = ba.Txn.Clone()
	ba.Txn.Refresh(ts)
	ba.Timestamp = ba.Txn.ReadTimestamp
	return true
}
