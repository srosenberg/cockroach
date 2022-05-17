package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

func (tsdb *DB) ContainsTimeSeries(start, end roachpb.RKey) bool {
	__antithesis_instrumentation__.Notify(648053)
	return !lastTSRKey.Less(start) && func() bool {
		__antithesis_instrumentation__.Notify(648054)
		return !end.Less(firstTSRKey) == true
	}() == true
}

func (tsdb *DB) MaintainTimeSeries(
	ctx context.Context,
	snapshot storage.Reader,
	start, end roachpb.RKey,
	db *kv.DB,
	mem *mon.BytesMonitor,
	budgetBytes int64,
	now hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(648055)
	series, err := tsdb.findTimeSeries(snapshot, start, end, now)
	if err != nil {
		__antithesis_instrumentation__.Notify(648058)
		return err
	} else {
		__antithesis_instrumentation__.Notify(648059)
	}
	__antithesis_instrumentation__.Notify(648056)
	if tsdb.WriteRollups() {
		__antithesis_instrumentation__.Notify(648060)
		qmc := MakeQueryMemoryContext(mem, mem, QueryMemoryOptions{
			BudgetBytes: budgetBytes,
		})
		if err := tsdb.rollupTimeSeries(ctx, series, now, qmc); err != nil {
			__antithesis_instrumentation__.Notify(648061)
			return err
		} else {
			__antithesis_instrumentation__.Notify(648062)
		}
	} else {
		__antithesis_instrumentation__.Notify(648063)
	}
	__antithesis_instrumentation__.Notify(648057)
	return tsdb.pruneTimeSeries(ctx, db, series, now)
}

var _ kvserver.TimeSeriesDataStore = (*DB)(nil)
