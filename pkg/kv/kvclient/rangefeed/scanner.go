package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/span"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

func (f *RangeFeed) runInitialScan(
	ctx context.Context, n *log.EveryN, r *retry.Retry,
) (canceled bool) {
	__antithesis_instrumentation__.Notify(89972)
	onValue := func(kv roachpb.KeyValue) {
		__antithesis_instrumentation__.Notify(89975)
		v := roachpb.RangeFeedValue{
			Key:   kv.Key,
			Value: kv.Value,
		}

		if !f.useRowTimestampInInitialScan {
			__antithesis_instrumentation__.Notify(89978)
			v.Value.Timestamp = f.initialTimestamp
		} else {
			__antithesis_instrumentation__.Notify(89979)
		}
		__antithesis_instrumentation__.Notify(89976)

		if f.withDiff {
			__antithesis_instrumentation__.Notify(89980)
			v.PrevValue = v.Value
			v.PrevValue.Timestamp = hlc.Timestamp{}
		} else {
			__antithesis_instrumentation__.Notify(89981)
		}
		__antithesis_instrumentation__.Notify(89977)

		f.onValue(ctx, &v)
	}
	__antithesis_instrumentation__.Notify(89973)

	getSpansToScan := f.getSpansToScan(ctx)

	r.Reset()
	for r.Next() {
		__antithesis_instrumentation__.Notify(89982)
		if err := f.client.Scan(ctx, getSpansToScan(), f.initialTimestamp, onValue, f.scanConfig); err != nil {
			__antithesis_instrumentation__.Notify(89983)
			if f.onInitialScanError != nil {
				__antithesis_instrumentation__.Notify(89985)
				if shouldStop := f.onInitialScanError(ctx, err); shouldStop {
					__antithesis_instrumentation__.Notify(89986)
					log.VEventf(ctx, 1, "stopping due to error: %v", err)
					return true
				} else {
					__antithesis_instrumentation__.Notify(89987)
				}
			} else {
				__antithesis_instrumentation__.Notify(89988)
			}
			__antithesis_instrumentation__.Notify(89984)
			if n.ShouldLog() {
				__antithesis_instrumentation__.Notify(89989)
				log.Warningf(ctx, "failed to perform initial scan: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(89990)
			}
		} else {
			__antithesis_instrumentation__.Notify(89991)
			if f.onInitialScanDone != nil {
				__antithesis_instrumentation__.Notify(89993)
				f.onInitialScanDone(ctx)
			} else {
				__antithesis_instrumentation__.Notify(89994)
			}
			__antithesis_instrumentation__.Notify(89992)
			break
		}
	}
	__antithesis_instrumentation__.Notify(89974)

	return false
}

func (f *RangeFeed) getSpansToScan(ctx context.Context) func() []roachpb.Span {
	__antithesis_instrumentation__.Notify(89995)
	retryAll := func() []roachpb.Span {
		__antithesis_instrumentation__.Notify(90000)
		return f.spans
	}
	__antithesis_instrumentation__.Notify(89996)

	if f.retryBehavior == ScanRetryAll {
		__antithesis_instrumentation__.Notify(90001)
		return retryAll
	} else {
		__antithesis_instrumentation__.Notify(90002)
	}
	__antithesis_instrumentation__.Notify(89997)

	frontier, err := span.MakeFrontier(f.spans...)
	if err != nil {
		__antithesis_instrumentation__.Notify(90003)

		log.Errorf(ctx, "failed to build frontier for the initial scan; "+
			"falling back to retry all behavior: err=%v", err)
		return retryAll
	} else {
		__antithesis_instrumentation__.Notify(90004)
	}
	__antithesis_instrumentation__.Notify(89998)

	var fm syncutil.Mutex
	userSpanDoneCallback := f.onSpanDone
	f.onSpanDone = func(ctx context.Context, sp roachpb.Span) error {
		__antithesis_instrumentation__.Notify(90005)
		if userSpanDoneCallback != nil {
			__antithesis_instrumentation__.Notify(90007)
			if err := userSpanDoneCallback(ctx, sp); err != nil {
				__antithesis_instrumentation__.Notify(90008)
				return err
			} else {
				__antithesis_instrumentation__.Notify(90009)
			}
		} else {
			__antithesis_instrumentation__.Notify(90010)
		}
		__antithesis_instrumentation__.Notify(90006)
		fm.Lock()
		defer fm.Unlock()
		_, err := frontier.Forward(sp, f.initialTimestamp)
		return err
	}
	__antithesis_instrumentation__.Notify(89999)

	isRetry := false
	var retrySpans []roachpb.Span
	return func() []roachpb.Span {
		__antithesis_instrumentation__.Notify(90011)
		if !isRetry {
			__antithesis_instrumentation__.Notify(90014)
			isRetry = true
			return f.spans
		} else {
			__antithesis_instrumentation__.Notify(90015)
		}
		__antithesis_instrumentation__.Notify(90012)

		retrySpans = retrySpans[:0]
		frontier.Entries(func(sp roachpb.Span, ts hlc.Timestamp) (done span.OpResult) {
			__antithesis_instrumentation__.Notify(90016)
			if ts.IsEmpty() {
				__antithesis_instrumentation__.Notify(90018)
				retrySpans = append(retrySpans, sp)
			} else {
				__antithesis_instrumentation__.Notify(90019)
			}
			__antithesis_instrumentation__.Notify(90017)
			return span.ContinueMatch
		})
		__antithesis_instrumentation__.Notify(90013)
		return retrySpans
	}

}
