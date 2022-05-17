package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const (
	splitQueueTimerDuration = 0

	splitQueuePurgatoryCheckInterval = 1 * time.Minute

	splitQueueConcurrency = 4
)

type splitQueue struct {
	*baseQueue
	db       *kv.DB
	purgChan <-chan time.Time

	loadBasedCount telemetry.Counter
}

func newSplitQueue(store *Store, db *kv.DB) *splitQueue {
	__antithesis_instrumentation__.Notify(123501)
	var purgChan <-chan time.Time
	if c := store.TestingKnobs().SplitQueuePurgatoryChan; c != nil {
		__antithesis_instrumentation__.Notify(123503)
		purgChan = c
	} else {
		__antithesis_instrumentation__.Notify(123504)
		purgTicker := time.NewTicker(splitQueuePurgatoryCheckInterval)
		purgChan = purgTicker.C
	}
	__antithesis_instrumentation__.Notify(123502)

	sq := &splitQueue{
		db:             db,
		purgChan:       purgChan,
		loadBasedCount: telemetry.GetCounter("kv.split.load"),
	}
	sq.baseQueue = newBaseQueue(
		"split", sq, store,
		queueConfig{
			maxSize:              defaultQueueMaxSize,
			maxConcurrency:       splitQueueConcurrency,
			needsLease:           true,
			needsSystemConfig:    true,
			acceptsUnsplitRanges: true,
			successes:            store.metrics.SplitQueueSuccesses,
			failures:             store.metrics.SplitQueueFailures,
			pending:              store.metrics.SplitQueuePending,
			processingNanos:      store.metrics.SplitQueueProcessingNanos,
			purgatory:            store.metrics.SplitQueuePurgatory,
		},
	)
	return sq
}

func shouldSplitRange(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	ms enginepb.MVCCStats,
	maxBytes int64,
	shouldBackpressureWrites bool,
	confReader spanconfig.StoreReader,
) (shouldQ bool, priority float64) {
	__antithesis_instrumentation__.Notify(123505)
	if confReader.NeedsSplit(ctx, desc.StartKey, desc.EndKey) {
		__antithesis_instrumentation__.Notify(123509)

		priority = 1
		shouldQ = true
	} else {
		__antithesis_instrumentation__.Notify(123510)
	}
	__antithesis_instrumentation__.Notify(123506)

	if ratio := float64(ms.Total()) / float64(maxBytes); ratio > 1 {
		__antithesis_instrumentation__.Notify(123511)
		priority += ratio
		shouldQ = true
	} else {
		__antithesis_instrumentation__.Notify(123512)
	}
	__antithesis_instrumentation__.Notify(123507)

	const additionalPriorityDueToBackpressure = 50
	if shouldQ && func() bool {
		__antithesis_instrumentation__.Notify(123513)
		return shouldBackpressureWrites == true
	}() == true {
		__antithesis_instrumentation__.Notify(123514)
		priority += additionalPriorityDueToBackpressure
	} else {
		__antithesis_instrumentation__.Notify(123515)
	}
	__antithesis_instrumentation__.Notify(123508)

	return shouldQ, priority
}

func (sq *splitQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, repl *Replica, confReader spanconfig.StoreReader,
) (shouldQ bool, priority float64) {
	__antithesis_instrumentation__.Notify(123516)
	shouldQ, priority = shouldSplitRange(ctx, repl.Desc(), repl.GetMVCCStats(),
		repl.GetMaxBytes(), repl.shouldBackpressureWrites(), confReader)

	if !shouldQ && func() bool {
		__antithesis_instrumentation__.Notify(123518)
		return repl.SplitByLoadEnabled() == true
	}() == true {
		__antithesis_instrumentation__.Notify(123519)
		if splitKey := repl.loadBasedSplitter.MaybeSplitKey(timeutil.Now()); splitKey != nil {
			__antithesis_instrumentation__.Notify(123520)
			shouldQ, priority = true, 1.0
		} else {
			__antithesis_instrumentation__.Notify(123521)
		}
	} else {
		__antithesis_instrumentation__.Notify(123522)
	}
	__antithesis_instrumentation__.Notify(123517)

	return shouldQ, priority
}

type unsplittableRangeError struct{}

func (unsplittableRangeError) Error() string {
	__antithesis_instrumentation__.Notify(123523)
	return "could not find valid split key"
}
func (unsplittableRangeError) purgatoryErrorMarker() { __antithesis_instrumentation__.Notify(123524) }

var _ purgatoryError = unsplittableRangeError{}

func (sq *splitQueue) process(
	ctx context.Context, r *Replica, confReader spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(123525)
	processed, err = sq.processAttempt(ctx, r, confReader)
	if errors.HasType(err, (*roachpb.ConditionFailedError)(nil)) {
		__antithesis_instrumentation__.Notify(123527)

		log.Infof(ctx, "split saw concurrent descriptor modification; maybe retrying")
		sq.MaybeAddAsync(ctx, r, sq.store.Clock().NowAsClockTimestamp())
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(123528)
	}
	__antithesis_instrumentation__.Notify(123526)

	return processed, err
}

func (sq *splitQueue) processAttempt(
	ctx context.Context, r *Replica, confReader spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(123529)
	desc := r.Desc()

	if splitKey := confReader.ComputeSplitKey(ctx, desc.StartKey, desc.EndKey); splitKey != nil {
		__antithesis_instrumentation__.Notify(123533)
		if _, err := r.adminSplitWithDescriptor(
			ctx,
			roachpb.AdminSplitRequest{
				RequestHeader: roachpb.RequestHeader{
					Key: splitKey.AsRawKey(),
				},
				SplitKey:       splitKey.AsRawKey(),
				ExpirationTime: hlc.Timestamp{},
			},
			desc,
			false,
			"span config",
		); err != nil {
			__antithesis_instrumentation__.Notify(123535)
			return false, errors.Wrapf(err, "unable to split %s at key %q", r, splitKey)
		} else {
			__antithesis_instrumentation__.Notify(123536)
		}
		__antithesis_instrumentation__.Notify(123534)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(123537)
	}
	__antithesis_instrumentation__.Notify(123530)

	size := r.GetMVCCStats().Total()
	maxBytes := r.GetMaxBytes()
	if maxBytes > 0 && func() bool {
		__antithesis_instrumentation__.Notify(123538)
		return float64(size)/float64(maxBytes) > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(123539)
		_, err := r.adminSplitWithDescriptor(
			ctx,
			roachpb.AdminSplitRequest{},
			desc,
			false,
			fmt.Sprintf("%s above threshold size %s", humanizeutil.IBytes(size), humanizeutil.IBytes(maxBytes)),
		)

		return err == nil, err
	} else {
		__antithesis_instrumentation__.Notify(123540)
	}
	__antithesis_instrumentation__.Notify(123531)

	now := timeutil.Now()
	if splitByLoadKey := r.loadBasedSplitter.MaybeSplitKey(now); splitByLoadKey != nil {
		__antithesis_instrumentation__.Notify(123541)
		batchHandledQPS, _ := r.QueriesPerSecond()
		raftAppliedQPS := r.WritesPerSecond()
		splitQPS := r.loadBasedSplitter.LastQPS(now)
		reason := fmt.Sprintf(
			"load at key %s (%.2f splitQPS, %.2f batches/sec, %.2f raft mutations/sec)",
			splitByLoadKey,
			splitQPS,
			batchHandledQPS,
			raftAppliedQPS,
		)

		var expTime hlc.Timestamp
		if expDelay := kvserverbase.SplitByLoadMergeDelay.Get(&sq.store.cfg.Settings.SV); expDelay > 0 {
			__antithesis_instrumentation__.Notify(123544)
			expTime = sq.store.Clock().Now().Add(expDelay.Nanoseconds(), 0)
		} else {
			__antithesis_instrumentation__.Notify(123545)
		}
		__antithesis_instrumentation__.Notify(123542)
		if _, pErr := r.adminSplitWithDescriptor(
			ctx,
			roachpb.AdminSplitRequest{
				RequestHeader: roachpb.RequestHeader{
					Key: splitByLoadKey,
				},
				SplitKey:       splitByLoadKey,
				ExpirationTime: expTime,
			},
			desc,
			false,
			reason,
		); pErr != nil {
			__antithesis_instrumentation__.Notify(123546)
			return false, errors.Wrapf(pErr, "unable to split %s at key %q", r, splitByLoadKey)
		} else {
			__antithesis_instrumentation__.Notify(123547)
		}
		__antithesis_instrumentation__.Notify(123543)

		telemetry.Inc(sq.loadBasedCount)

		r.loadBasedSplitter.Reset(sq.store.Clock().PhysicalTime())
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(123548)
	}
	__antithesis_instrumentation__.Notify(123532)
	return false, nil
}

func (*splitQueue) timer(_ time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(123549)
	return splitQueueTimerDuration
}

func (sq *splitQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(123550)
	return sq.purgChan
}
