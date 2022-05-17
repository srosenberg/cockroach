package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

func RangeLookup(
	ctx context.Context,
	sender Sender,
	key roachpb.Key,
	rc roachpb.ReadConsistencyType,
	prefetchNum int64,
	prefetchReverse bool,
) (rs, preRs []roachpb.RangeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(127599)

	opts := retry.Options{
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     500 * time.Millisecond,
		Multiplier:     2,
	}

	for r := retry.StartWithCtx(ctx, opts); r.Next(); {
		__antithesis_instrumentation__.Notify(127602)

		rkey, err := addrForDir(prefetchReverse)(key)
		if err != nil {
			__antithesis_instrumentation__.Notify(127609)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(127610)
		}
		__antithesis_instrumentation__.Notify(127603)

		descs, intentDescs, err := lookupRangeFwdScan(ctx, sender, rkey, rc, prefetchNum, prefetchReverse)
		if err != nil {
			__antithesis_instrumentation__.Notify(127611)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(127612)
		}
		__antithesis_instrumentation__.Notify(127604)
		if prefetchReverse {
			__antithesis_instrumentation__.Notify(127613)
			descs, intentDescs, err = lookupRangeRevScan(ctx, sender, rkey, rc, prefetchNum,
				prefetchReverse, descs, intentDescs)
			if err != nil {
				__antithesis_instrumentation__.Notify(127614)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(127615)
			}
		} else {
			__antithesis_instrumentation__.Notify(127616)
		}
		__antithesis_instrumentation__.Notify(127605)

		desiredDesc := containsForDir(prefetchReverse, rkey)
		var matchingRanges []roachpb.RangeDescriptor
		var prefetchedRanges []roachpb.RangeDescriptor
		for index := range descs {
			__antithesis_instrumentation__.Notify(127617)
			desc := &descs[index]
			if desiredDesc(desc) {
				__antithesis_instrumentation__.Notify(127618)
				if len(matchingRanges) == 0 {
					__antithesis_instrumentation__.Notify(127619)
					matchingRanges = append(matchingRanges, *desc)
				} else {
					__antithesis_instrumentation__.Notify(127620)

					if desc.Generation > matchingRanges[0].Generation {
						__antithesis_instrumentation__.Notify(127621)

						matchingRanges[0] = *desc
					} else {
						__antithesis_instrumentation__.Notify(127622)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(127623)

				prefetchedRanges = append(prefetchedRanges, *desc)
			}
		}
		__antithesis_instrumentation__.Notify(127606)
		for i := range intentDescs {
			__antithesis_instrumentation__.Notify(127624)
			desc := &intentDescs[i]
			if desiredDesc(desc) {
				__antithesis_instrumentation__.Notify(127625)
				matchingRanges = append(matchingRanges, *desc)

				break
			} else {
				__antithesis_instrumentation__.Notify(127626)
			}
		}
		__antithesis_instrumentation__.Notify(127607)
		if len(matchingRanges) > 0 {
			__antithesis_instrumentation__.Notify(127627)
			return matchingRanges, prefetchedRanges, nil
		} else {
			__antithesis_instrumentation__.Notify(127628)
		}
		__antithesis_instrumentation__.Notify(127608)

		log.Warningf(ctx, "range lookup of key %s found only non-matching ranges %v; retrying",
			key, prefetchedRanges)
	}
	__antithesis_instrumentation__.Notify(127600)

	ctxErr := ctx.Err()
	if ctxErr == nil {
		__antithesis_instrumentation__.Notify(127629)
		log.Fatalf(ctx, "retry loop broke before context expired")
	} else {
		__antithesis_instrumentation__.Notify(127630)
	}
	__antithesis_instrumentation__.Notify(127601)
	return nil, nil, ctxErr
}

func lookupRangeFwdScan(
	ctx context.Context,
	sender Sender,
	key roachpb.RKey,
	rc roachpb.ReadConsistencyType,
	prefetchNum int64,
	prefetchReverse bool,
) (rs, preRs []roachpb.RangeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(127631)
	if skipFwd := prefetchReverse && func() bool {
		__antithesis_instrumentation__.Notify(127640)
		return key.Equal(roachpb.KeyMax) == true
	}() == true; skipFwd {
		__antithesis_instrumentation__.Notify(127641)

		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(127642)
	}
	__antithesis_instrumentation__.Notify(127632)

	metaKey := keys.RangeMetaKey(key)
	bounds, err := keys.MetaScanBounds(metaKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(127643)
		return nil, nil, errors.Wrap(err, "could not create scan bounds for range lookup")
	} else {
		__antithesis_instrumentation__.Notify(127644)
	}
	__antithesis_instrumentation__.Notify(127633)

	ba := roachpb.BatchRequest{}
	ba.ReadConsistency = rc
	if prefetchReverse {
		__antithesis_instrumentation__.Notify(127645)

		ba.MaxSpanRequestKeys = 1
	} else {
		__antithesis_instrumentation__.Notify(127646)
		ba.MaxSpanRequestKeys = prefetchNum + 1
	}
	__antithesis_instrumentation__.Notify(127634)
	ba.Add(&roachpb.ScanRequest{
		RequestHeader: roachpb.RequestHeaderFromSpan(bounds.AsRawSpanWithNoLocals()),
	})
	if !TestingIsRangeLookup(ba) {
		__antithesis_instrumentation__.Notify(127647)
		log.Fatalf(ctx, "BatchRequest %v not detectable as RangeLookup", ba)
	} else {
		__antithesis_instrumentation__.Notify(127648)
	}
	__antithesis_instrumentation__.Notify(127635)

	br, pErr := sender.Send(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(127649)
		return nil, nil, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(127650)
	}
	__antithesis_instrumentation__.Notify(127636)
	scanRes := br.Responses[0].GetInner().(*roachpb.ScanResponse)

	descs, err := kvsToRangeDescriptors(scanRes.Rows)
	if err != nil {
		__antithesis_instrumentation__.Notify(127651)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(127652)
	}
	__antithesis_instrumentation__.Notify(127637)
	intentDescs, err := kvsToRangeDescriptors(scanRes.IntentRows)
	if err != nil {
		__antithesis_instrumentation__.Notify(127653)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(127654)
	}
	__antithesis_instrumentation__.Notify(127638)

	if prefetchReverse {
		__antithesis_instrumentation__.Notify(127655)
		desiredDesc := containsForDir(prefetchReverse, key)
		if len(descs) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(127657)
			return !desiredDesc(&descs[0]) == true
		}() == true {
			__antithesis_instrumentation__.Notify(127658)
			descs = nil
		} else {
			__antithesis_instrumentation__.Notify(127659)
		}
		__antithesis_instrumentation__.Notify(127656)
		if len(intentDescs) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(127660)
			return !desiredDesc(&intentDescs[0]) == true
		}() == true {
			__antithesis_instrumentation__.Notify(127661)
			intentDescs = nil
		} else {
			__antithesis_instrumentation__.Notify(127662)
		}
	} else {
		__antithesis_instrumentation__.Notify(127663)
	}
	__antithesis_instrumentation__.Notify(127639)
	return descs, intentDescs, nil
}

func lookupRangeRevScan(
	ctx context.Context,
	sender Sender,
	key roachpb.RKey,
	rc roachpb.ReadConsistencyType,
	prefetchNum int64,
	prefetchReverse bool,
	fwdDescs, fwdIntentDescs []roachpb.RangeDescriptor,
) (rs, preRs []roachpb.RangeDescriptor, err error) {
	__antithesis_instrumentation__.Notify(127664)

	maxKeys := prefetchNum + 1
	if len(fwdDescs) > 0 {
		__antithesis_instrumentation__.Notify(127671)
		maxKeys--
		if maxKeys == 0 {
			__antithesis_instrumentation__.Notify(127672)
			return fwdDescs, fwdIntentDescs, nil
		} else {
			__antithesis_instrumentation__.Notify(127673)
		}
	} else {
		__antithesis_instrumentation__.Notify(127674)
	}
	__antithesis_instrumentation__.Notify(127665)

	metaKey := keys.RangeMetaKey(key)
	revBounds, err := keys.MetaReverseScanBounds(metaKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(127675)
		return nil, nil, errors.Wrap(err, "could not create scan bounds for reverse range lookup")
	} else {
		__antithesis_instrumentation__.Notify(127676)
	}
	__antithesis_instrumentation__.Notify(127666)

	ba := roachpb.BatchRequest{}
	ba.ReadConsistency = rc
	ba.MaxSpanRequestKeys = maxKeys
	ba.Add(&roachpb.ReverseScanRequest{
		RequestHeader: roachpb.RequestHeaderFromSpan(revBounds.AsRawSpanWithNoLocals()),
	})
	if !TestingIsRangeLookup(ba) {
		__antithesis_instrumentation__.Notify(127677)
		log.Fatalf(ctx, "BatchRequest %v not detectable as RangeLookup", ba)
	} else {
		__antithesis_instrumentation__.Notify(127678)
	}
	__antithesis_instrumentation__.Notify(127667)

	br, pErr := sender.Send(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(127679)
		return nil, nil, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(127680)
	}
	__antithesis_instrumentation__.Notify(127668)
	revScanRes := br.Responses[0].GetInner().(*roachpb.ReverseScanResponse)

	revDescs, err := kvsToRangeDescriptors(revScanRes.Rows)
	if err != nil {
		__antithesis_instrumentation__.Notify(127681)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(127682)
	}
	__antithesis_instrumentation__.Notify(127669)
	revIntentDescs, err := kvsToRangeDescriptors(revScanRes.IntentRows)
	if err != nil {
		__antithesis_instrumentation__.Notify(127683)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(127684)
	}
	__antithesis_instrumentation__.Notify(127670)
	return append(fwdDescs, revDescs...), append(fwdIntentDescs, revIntentDescs...), nil
}

func addrForDir(prefetchReverse bool) func(roachpb.Key) (roachpb.RKey, error) {
	__antithesis_instrumentation__.Notify(127685)
	if prefetchReverse {
		__antithesis_instrumentation__.Notify(127687)
		return keys.AddrUpperBound
	} else {
		__antithesis_instrumentation__.Notify(127688)
	}
	__antithesis_instrumentation__.Notify(127686)
	return keys.Addr
}

func containsForDir(prefetchReverse bool, key roachpb.RKey) func(*roachpb.RangeDescriptor) bool {
	__antithesis_instrumentation__.Notify(127689)
	return func(desc *roachpb.RangeDescriptor) bool {
		__antithesis_instrumentation__.Notify(127690)
		contains := (*roachpb.RangeDescriptor).ContainsKey
		if prefetchReverse {
			__antithesis_instrumentation__.Notify(127692)
			contains = (*roachpb.RangeDescriptor).ContainsKeyInverted
		} else {
			__antithesis_instrumentation__.Notify(127693)
		}
		__antithesis_instrumentation__.Notify(127691)
		return contains(desc, key)
	}
}

func kvsToRangeDescriptors(kvs []roachpb.KeyValue) ([]roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(127694)
	descs := make([]roachpb.RangeDescriptor, len(kvs))
	for i, kv := range kvs {
		__antithesis_instrumentation__.Notify(127696)
		if err := kv.Value.GetProto(&descs[i]); err != nil {
			__antithesis_instrumentation__.Notify(127697)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(127698)
		}
	}
	__antithesis_instrumentation__.Notify(127695)
	return descs, nil
}

func TestingIsRangeLookup(ba roachpb.BatchRequest) bool {
	__antithesis_instrumentation__.Notify(127699)
	if ba.IsSingleRequest() {
		__antithesis_instrumentation__.Notify(127701)
		return TestingIsRangeLookupRequest(ba.Requests[0].GetInner())
	} else {
		__antithesis_instrumentation__.Notify(127702)
	}
	__antithesis_instrumentation__.Notify(127700)
	return false
}

var rangeLookupStartKeyBounds = roachpb.Span{
	Key:    keys.Meta1Prefix,
	EndKey: keys.Meta2KeyMax.Next(),
}
var rangeLookupEndKeyBounds = roachpb.Span{
	Key:    keys.Meta1Prefix.Next(),
	EndKey: keys.SystemPrefix.Next(),
}

func TestingIsRangeLookupRequest(req roachpb.Request) bool {
	__antithesis_instrumentation__.Notify(127703)
	switch req.(type) {
	case *roachpb.ScanRequest:
		__antithesis_instrumentation__.Notify(127705)
	case *roachpb.ReverseScanRequest:
		__antithesis_instrumentation__.Notify(127706)
	default:
		__antithesis_instrumentation__.Notify(127707)
		return false
	}
	__antithesis_instrumentation__.Notify(127704)
	s := req.Header()
	return rangeLookupStartKeyBounds.ContainsKey(s.Key) && func() bool {
		__antithesis_instrumentation__.Notify(127708)
		return rangeLookupEndKeyBounds.ContainsKey(s.EndKey) == true
	}() == true
}
