package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

var backpressureLogLimiter = log.Every(500 * time.Millisecond)

var backpressureRangeSizeMultiplier = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"kv.range.backpressure_range_size_multiplier",
	"multiple of range_max_bytes that a range is allowed to grow to without "+
		"splitting before writes to that range are blocked, or 0 to disable",
	2.0,
	func(v float64) error {
		__antithesis_instrumentation__.Notify(115575)
		if v != 0 && func() bool {
			__antithesis_instrumentation__.Notify(115577)
			return v < 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(115578)
			return errors.Errorf("backpressure multiplier cannot be smaller than 1: %f", v)
		} else {
			__antithesis_instrumentation__.Notify(115579)
		}
		__antithesis_instrumentation__.Notify(115576)
		return nil
	},
)

var backpressureByteTolerance = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"kv.range.backpressure_byte_tolerance",
	"defines the number of bytes above the product of "+
		"backpressure_range_size_multiplier and the range_max_size at which "+
		"backpressure will not apply",
	32<<20)

var backpressurableSpans = []roachpb.Span{
	{Key: keys.TimeseriesPrefix, EndKey: keys.TimeseriesKeyMax},

	{Key: keys.SystemConfigTableDataMax, EndKey: keys.TableDataMax},
}

func canBackpressureBatch(ba *roachpb.BatchRequest) bool {
	__antithesis_instrumentation__.Notify(115580)

	if ba.Txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(115583)
		return ba.Txn.Name == splitTxnName == true
	}() == true {
		__antithesis_instrumentation__.Notify(115584)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115585)
	}
	__antithesis_instrumentation__.Notify(115581)

	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(115586)
		req := ru.GetInner()
		if !roachpb.CanBackpressure(req) {
			__antithesis_instrumentation__.Notify(115588)
			continue
		} else {
			__antithesis_instrumentation__.Notify(115589)
		}
		__antithesis_instrumentation__.Notify(115587)

		for _, s := range backpressurableSpans {
			__antithesis_instrumentation__.Notify(115590)
			if s.Contains(req.Header().Span()) {
				__antithesis_instrumentation__.Notify(115591)
				return true
			} else {
				__antithesis_instrumentation__.Notify(115592)
			}
		}
	}
	__antithesis_instrumentation__.Notify(115582)
	return false
}

func (r *Replica) signallerForBatch(ba *roachpb.BatchRequest) signaller {
	__antithesis_instrumentation__.Notify(115593)
	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(115595)
		req := ru.GetInner()
		if roachpb.BypassesReplicaCircuitBreaker(req) {
			__antithesis_instrumentation__.Notify(115596)
			return neverTripSignaller{}
		} else {
			__antithesis_instrumentation__.Notify(115597)
		}
	}
	__antithesis_instrumentation__.Notify(115594)
	return r.breaker.Signal()
}

func (r *Replica) shouldBackpressureWrites() bool {
	__antithesis_instrumentation__.Notify(115598)
	mult := backpressureRangeSizeMultiplier.Get(&r.store.cfg.Settings.SV)
	if mult == 0 {
		__antithesis_instrumentation__.Notify(115602)

		return false
	} else {
		__antithesis_instrumentation__.Notify(115603)
	}
	__antithesis_instrumentation__.Notify(115599)

	r.mu.RLock()
	defer r.mu.RUnlock()
	exceeded, bytesOver := r.exceedsMultipleOfSplitSizeRLocked(mult)
	if !exceeded {
		__antithesis_instrumentation__.Notify(115604)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115605)
	}
	__antithesis_instrumentation__.Notify(115600)
	if bytesOver > backpressureByteTolerance.Get(&r.store.cfg.Settings.SV) {
		__antithesis_instrumentation__.Notify(115606)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115607)
	}
	__antithesis_instrumentation__.Notify(115601)
	return true
}

func (r *Replica) maybeBackpressureBatch(ctx context.Context, ba *roachpb.BatchRequest) error {
	__antithesis_instrumentation__.Notify(115608)
	if !canBackpressureBatch(ba) {
		__antithesis_instrumentation__.Notify(115611)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(115612)
	}
	__antithesis_instrumentation__.Notify(115609)

	for first := true; r.shouldBackpressureWrites(); first = false {
		__antithesis_instrumentation__.Notify(115613)
		if first {
			__antithesis_instrumentation__.Notify(115616)
			r.store.metrics.BackpressuredOnSplitRequests.Inc(1)
			defer r.store.metrics.BackpressuredOnSplitRequests.Dec(1)

			if backpressureLogLimiter.ShouldLog() {
				__antithesis_instrumentation__.Notify(115617)
				log.Warningf(ctx, "applying backpressure to limit range growth on batch %s", ba)
			} else {
				__antithesis_instrumentation__.Notify(115618)
			}
		} else {
			__antithesis_instrumentation__.Notify(115619)
		}
		__antithesis_instrumentation__.Notify(115614)

		splitC := make(chan error, 1)
		if !r.store.splitQueue.MaybeAddCallback(r.RangeID, func(err error) {
			__antithesis_instrumentation__.Notify(115620)
			splitC <- err
		}) {
			__antithesis_instrumentation__.Notify(115621)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(115622)
		}
		__antithesis_instrumentation__.Notify(115615)

		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(115623)
			return errors.Wrapf(
				ctx.Err(), "aborted while applying backpressure to %s on range %s", ba, r.Desc(),
			)
		case err := <-splitC:
			__antithesis_instrumentation__.Notify(115624)
			if err != nil {
				__antithesis_instrumentation__.Notify(115625)
				return errors.Wrapf(
					err, "split failed while applying backpressure to %s on range %s", ba, r.Desc(),
				)
			} else {
				__antithesis_instrumentation__.Notify(115626)
			}
		}
	}
	__antithesis_instrumentation__.Notify(115610)
	return nil
}
