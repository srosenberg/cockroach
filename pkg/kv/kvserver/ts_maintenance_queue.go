package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

const (
	TimeSeriesMaintenanceInterval = 24 * time.Hour

	TimeSeriesMaintenanceMemoryBudget = int64(8 * 1024 * 1024)
)

type TimeSeriesDataStore interface {
	ContainsTimeSeries(roachpb.RKey, roachpb.RKey) bool
	MaintainTimeSeries(
		context.Context,
		storage.Reader,
		roachpb.RKey,
		roachpb.RKey,
		*kv.DB,
		*mon.BytesMonitor,
		int64,
		hlc.Timestamp,
	) error
}

type timeSeriesMaintenanceQueue struct {
	*baseQueue
	tsData         TimeSeriesDataStore
	replicaCountFn func() int
	db             *kv.DB
	mem            *mon.BytesMonitor
}

func newTimeSeriesMaintenanceQueue(
	store *Store, db *kv.DB, tsData TimeSeriesDataStore,
) *timeSeriesMaintenanceQueue {
	__antithesis_instrumentation__.Notify(126768)
	q := &timeSeriesMaintenanceQueue{
		tsData:         tsData,
		replicaCountFn: store.ReplicaCount,
		db:             db,
		mem: mon.NewUnlimitedMonitor(
			context.Background(),
			"timeseries-maintenance-queue",
			mon.MemoryResource,
			nil,
			nil,

			TimeSeriesMaintenanceMemoryBudget*3,
			store.cfg.Settings,
		),
	}
	q.baseQueue = newBaseQueue(
		"timeSeriesMaintenance", q, store,
		queueConfig{
			maxSize:              defaultQueueMaxSize,
			needsLease:           true,
			needsSystemConfig:    false,
			acceptsUnsplitRanges: true,
			successes:            store.metrics.TimeSeriesMaintenanceQueueSuccesses,
			failures:             store.metrics.TimeSeriesMaintenanceQueueFailures,
			pending:              store.metrics.TimeSeriesMaintenanceQueuePending,
			processingNanos:      store.metrics.TimeSeriesMaintenanceQueueProcessingNanos,
		},
	)

	return q
}

func (q *timeSeriesMaintenanceQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, repl *Replica, _ spanconfig.StoreReader,
) (shouldQ bool, priority float64) {
	__antithesis_instrumentation__.Notify(126769)
	if !repl.store.cfg.TestingKnobs.DisableLastProcessedCheck {
		__antithesis_instrumentation__.Notify(126772)
		lpTS, err := repl.getQueueLastProcessed(ctx, q.name)
		if err != nil {
			__antithesis_instrumentation__.Notify(126774)
			return false, 0
		} else {
			__antithesis_instrumentation__.Notify(126775)
		}
		__antithesis_instrumentation__.Notify(126773)
		shouldQ, priority = shouldQueueAgain(now.ToTimestamp(), lpTS, TimeSeriesMaintenanceInterval)
		if !shouldQ {
			__antithesis_instrumentation__.Notify(126776)
			return
		} else {
			__antithesis_instrumentation__.Notify(126777)
		}
	} else {
		__antithesis_instrumentation__.Notify(126778)
	}
	__antithesis_instrumentation__.Notify(126770)
	desc := repl.Desc()
	if q.tsData.ContainsTimeSeries(desc.StartKey, desc.EndKey) {
		__antithesis_instrumentation__.Notify(126779)
		return
	} else {
		__antithesis_instrumentation__.Notify(126780)
	}
	__antithesis_instrumentation__.Notify(126771)
	return false, 0
}

func (q *timeSeriesMaintenanceQueue) process(
	ctx context.Context, repl *Replica, _ spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(126781)
	desc := repl.Desc()
	snap := repl.store.Engine().NewSnapshot()
	now := repl.store.Clock().Now()
	defer snap.Close()
	if err := q.tsData.MaintainTimeSeries(
		ctx, snap, desc.StartKey, desc.EndKey, q.db, q.mem, TimeSeriesMaintenanceMemoryBudget, now,
	); err != nil {
		__antithesis_instrumentation__.Notify(126784)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(126785)
	}
	__antithesis_instrumentation__.Notify(126782)

	if err := repl.setQueueLastProcessed(ctx, q.name, now); err != nil {
		__antithesis_instrumentation__.Notify(126786)
		log.VErrEventf(ctx, 2, "failed to update last processed time: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(126787)
	}
	__antithesis_instrumentation__.Notify(126783)
	return true, nil
}

func (q *timeSeriesMaintenanceQueue) timer(duration time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(126788)

	replicaCount := q.replicaCountFn()
	if replicaCount == 0 {
		__antithesis_instrumentation__.Notify(126791)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(126792)
	}
	__antithesis_instrumentation__.Notify(126789)
	replInterval := TimeSeriesMaintenanceInterval / time.Duration(replicaCount)
	if replInterval < duration {
		__antithesis_instrumentation__.Notify(126793)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(126794)
	}
	__antithesis_instrumentation__.Notify(126790)
	return replInterval - duration
}

func (*timeSeriesMaintenanceQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(126795)
	return nil
}
