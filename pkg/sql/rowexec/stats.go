package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type inputStatCollector struct {
	execinfra.RowSource
	stats execinfrapb.InputStats
}

var _ execinfra.RowSource = &inputStatCollector{}
var _ execinfra.OpNode = &inputStatCollector{}

func newInputStatCollector(input execinfra.RowSource) *inputStatCollector {
	__antithesis_instrumentation__.Notify(574984)
	res := &inputStatCollector{RowSource: input}
	res.stats.NumTuples.Set(0)
	return res
}

func (isc *inputStatCollector) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(574985)
	if _, ok := isc.RowSource.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(574987)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(574988)
	}
	__antithesis_instrumentation__.Notify(574986)
	return 0
}

func (isc *inputStatCollector) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(574989)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(574991)
		return isc.RowSource.(execinfra.OpNode)
	} else {
		__antithesis_instrumentation__.Notify(574992)
	}
	__antithesis_instrumentation__.Notify(574990)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}

func (isc *inputStatCollector) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(574993)
	start := timeutil.Now()
	row, meta := isc.RowSource.Next()
	if row != nil {
		__antithesis_instrumentation__.Notify(574995)
		isc.stats.NumTuples.Add(1)
	} else {
		__antithesis_instrumentation__.Notify(574996)
	}
	__antithesis_instrumentation__.Notify(574994)
	isc.stats.WaitTime.Add(timeutil.Since(start))
	return row, meta
}

type rowFetcherStatCollector struct {
	fetcher *row.Fetcher

	stats              execinfrapb.InputStats
	startScanStallTime time.Duration
}

var _ rowFetcher = &rowFetcherStatCollector{}

func newRowFetcherStatCollector(f *row.Fetcher) *rowFetcherStatCollector {
	__antithesis_instrumentation__.Notify(574997)
	res := &rowFetcherStatCollector{fetcher: f}
	res.stats.NumTuples.Set(0)
	return res
}

func (c *rowFetcherStatCollector) StartScan(
	ctx context.Context,
	txn *kv.Txn,
	spans roachpb.Spans,
	batchBytesLimit rowinfra.BytesLimit,
	limitHint rowinfra.RowLimit,
	traceKV bool,
	forceProductionKVBatchSize bool,
) error {
	__antithesis_instrumentation__.Notify(574998)
	start := timeutil.Now()
	err := c.fetcher.StartScan(ctx, txn, spans, batchBytesLimit, limitHint, traceKV, forceProductionKVBatchSize)
	c.startScanStallTime += timeutil.Since(start)
	return err
}

func (c *rowFetcherStatCollector) StartScanFrom(
	ctx context.Context, f row.KVBatchFetcher, traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(574999)
	start := timeutil.Now()
	err := c.fetcher.StartScanFrom(ctx, f, traceKV)
	c.startScanStallTime += timeutil.Since(start)
	return err
}

func (c *rowFetcherStatCollector) StartInconsistentScan(
	ctx context.Context,
	db *kv.DB,
	initialTimestamp hlc.Timestamp,
	maxTimestampAge time.Duration,
	spans roachpb.Spans,
	batchBytesLimit rowinfra.BytesLimit,
	limitHint rowinfra.RowLimit,
	traceKV bool,
	forceProductionKVBatchSize bool,
	qualityOfService sessiondatapb.QoSLevel,
) error {
	__antithesis_instrumentation__.Notify(575000)
	start := timeutil.Now()
	err := c.fetcher.StartInconsistentScan(
		ctx, db, initialTimestamp, maxTimestampAge, spans, batchBytesLimit, limitHint, traceKV,
		forceProductionKVBatchSize, qualityOfService,
	)
	c.startScanStallTime += timeutil.Since(start)
	return err
}

func (c *rowFetcherStatCollector) NextRow(ctx context.Context) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(575001)
	start := timeutil.Now()
	row, err := c.fetcher.NextRow(ctx)
	if row != nil {
		__antithesis_instrumentation__.Notify(575003)
		c.stats.NumTuples.Add(1)
	} else {
		__antithesis_instrumentation__.Notify(575004)
	}
	__antithesis_instrumentation__.Notify(575002)
	c.stats.WaitTime.Add(timeutil.Since(start))
	return row, err
}

func (c *rowFetcherStatCollector) NextRowInto(
	ctx context.Context, destination rowenc.EncDatumRow, colIdxMap catalog.TableColMap,
) (ok bool, err error) {
	__antithesis_instrumentation__.Notify(575005)
	start := timeutil.Now()
	ok, err = c.fetcher.NextRowInto(ctx, destination, colIdxMap)
	if ok {
		__antithesis_instrumentation__.Notify(575007)
		c.stats.NumTuples.Add(1)
	} else {
		__antithesis_instrumentation__.Notify(575008)
	}
	__antithesis_instrumentation__.Notify(575006)
	c.stats.WaitTime.Add(timeutil.Since(start))
	return ok, err
}

func (c *rowFetcherStatCollector) PartialKey(nCols int) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(575009)
	return c.fetcher.PartialKey(nCols)
}

func (c *rowFetcherStatCollector) Reset() {
	__antithesis_instrumentation__.Notify(575010)
	c.fetcher.Reset()
}

func (c *rowFetcherStatCollector) GetBytesRead() int64 {
	__antithesis_instrumentation__.Notify(575011)
	return c.fetcher.GetBytesRead()
}

func (c *rowFetcherStatCollector) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575012)
	c.fetcher.Close(ctx)
}

func getInputStats(input execinfra.RowSource) (execinfrapb.InputStats, bool) {
	__antithesis_instrumentation__.Notify(575013)
	isc, ok := input.(*inputStatCollector)
	if !ok {
		__antithesis_instrumentation__.Notify(575015)
		return execinfrapb.InputStats{}, false
	} else {
		__antithesis_instrumentation__.Notify(575016)
	}
	__antithesis_instrumentation__.Notify(575014)
	return isc.stats, true
}

func getFetcherInputStats(f rowFetcher) (execinfrapb.InputStats, bool) {
	__antithesis_instrumentation__.Notify(575017)
	rfsc, ok := f.(*rowFetcherStatCollector)
	if !ok {
		__antithesis_instrumentation__.Notify(575019)
		return execinfrapb.InputStats{}, false
	} else {
		__antithesis_instrumentation__.Notify(575020)
	}
	__antithesis_instrumentation__.Notify(575018)

	rfsc.stats.WaitTime.Add(rfsc.startScanStallTime)
	return rfsc.stats, true
}
