package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type rowFetcher interface {
	StartScan(
		_ context.Context, _ *kv.Txn, _ roachpb.Spans, batchBytesLimit rowinfra.BytesLimit,
		rowLimitHint rowinfra.RowLimit, traceKV bool, forceProductionKVBatchSize bool,
	) error
	StartScanFrom(_ context.Context, _ row.KVBatchFetcher, traceKV bool) error
	StartInconsistentScan(
		_ context.Context,
		_ *kv.DB,
		initialTimestamp hlc.Timestamp,
		maxTimestampAge time.Duration,
		spans roachpb.Spans,
		batchBytesLimit rowinfra.BytesLimit,
		rowLimitHint rowinfra.RowLimit,
		traceKV bool,
		forceProductionKVBatchSize bool,
		qualityOfService sessiondatapb.QoSLevel,
	) error

	NextRow(ctx context.Context) (rowenc.EncDatumRow, error)
	NextRowInto(
		ctx context.Context, destination rowenc.EncDatumRow, colIdxMap catalog.TableColMap,
	) (ok bool, err error)

	PartialKey(nCols int) (roachpb.Key, error)
	Reset()
	GetBytesRead() int64

	Close(ctx context.Context)
}

func makeRowFetcherLegacy(
	flowCtx *execinfra.FlowCtx,
	desc catalog.TableDescriptor,
	indexIdx int,
	reverseScan bool,
	valNeededForCol util.FastIntSet,
	mon *mon.BytesMonitor,
	alloc *tree.DatumAlloc,
	lockStrength descpb.ScanLockingStrength,
	lockWaitPolicy descpb.ScanLockingWaitPolicy,
	withSystemColumns bool,
) (*row.Fetcher, error) {
	__antithesis_instrumentation__.Notify(574443)
	colIDs := make([]descpb.ColumnID, 0, len(desc.AllColumns()))
	for i, col := range desc.ReadableColumns() {
		__antithesis_instrumentation__.Notify(574449)
		if valNeededForCol.Contains(i) {
			__antithesis_instrumentation__.Notify(574450)
			colIDs = append(colIDs, col.GetID())
		} else {
			__antithesis_instrumentation__.Notify(574451)
		}
	}
	__antithesis_instrumentation__.Notify(574444)
	if withSystemColumns {
		__antithesis_instrumentation__.Notify(574452)
		start := len(desc.ReadableColumns())
		for i, col := range desc.SystemColumns() {
			__antithesis_instrumentation__.Notify(574453)
			if valNeededForCol.Contains(start + i) {
				__antithesis_instrumentation__.Notify(574454)
				colIDs = append(colIDs, col.GetID())
			} else {
				__antithesis_instrumentation__.Notify(574455)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(574456)
	}
	__antithesis_instrumentation__.Notify(574445)

	if indexIdx >= len(desc.ActiveIndexes()) {
		__antithesis_instrumentation__.Notify(574457)
		return nil, errors.Errorf("invalid indexIdx %d", indexIdx)
	} else {
		__antithesis_instrumentation__.Notify(574458)
	}
	__antithesis_instrumentation__.Notify(574446)
	index := desc.ActiveIndexes()[indexIdx]

	var spec descpb.IndexFetchSpec
	if err := rowenc.InitIndexFetchSpec(&spec, flowCtx.Codec(), desc, index, colIDs); err != nil {
		__antithesis_instrumentation__.Notify(574459)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574460)
	}
	__antithesis_instrumentation__.Notify(574447)

	fetcher := &row.Fetcher{}
	if err := fetcher.Init(
		flowCtx.EvalCtx.Context,
		reverseScan,
		lockStrength,
		lockWaitPolicy,
		flowCtx.EvalCtx.SessionData().LockTimeout,
		alloc,
		mon,
		&spec,
	); err != nil {
		__antithesis_instrumentation__.Notify(574461)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574462)
	}
	__antithesis_instrumentation__.Notify(574448)
	return fetcher, nil
}
