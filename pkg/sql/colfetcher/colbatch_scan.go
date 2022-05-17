package colfetcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type ColBatchScan struct {
	colexecop.ZeroInputNode
	colexecop.InitHelper
	execinfra.SpansWithCopy

	flowCtx         *execinfra.FlowCtx
	bsHeader        *roachpb.BoundedStalenessHeader
	cf              *cFetcher
	limitHint       rowinfra.RowLimit
	batchBytesLimit rowinfra.BytesLimit
	parallelize     bool

	tracingSpan *tracing.Span
	mu          struct {
		syncutil.Mutex

		rowsRead int64
	}

	ResultTypes []*types.T
}

type ScanOperator interface {
	colexecop.KVReader
	execinfra.Releasable
	colexecop.ClosableOperator
}

var _ ScanOperator = &ColBatchScan{}

func (s *ColBatchScan) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455561)
	if !s.InitHelper.Init(ctx) {
		__antithesis_instrumentation__.Notify(455563)
		return
	} else {
		__antithesis_instrumentation__.Notify(455564)
	}
	__antithesis_instrumentation__.Notify(455562)

	s.Ctx, s.tracingSpan = execinfra.ProcessorSpan(s.Ctx, "colbatchscan")
	limitBatches := !s.parallelize
	if err := s.cf.StartScan(
		s.Ctx,
		s.flowCtx.Txn,
		s.Spans,
		s.bsHeader,
		limitBatches,
		s.batchBytesLimit,
		s.limitHint,
		s.flowCtx.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
	); err != nil {
		__antithesis_instrumentation__.Notify(455565)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(455566)
	}
}

func (s *ColBatchScan) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(455567)
	bat, err := s.cf.NextBatch(s.Ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(455570)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(455571)
	}
	__antithesis_instrumentation__.Notify(455568)
	if bat.Selection() != nil {
		__antithesis_instrumentation__.Notify(455572)
		colexecerror.InternalError(errors.AssertionFailedf("unexpectedly a selection vector is set on the batch coming from CFetcher"))
	} else {
		__antithesis_instrumentation__.Notify(455573)
	}
	__antithesis_instrumentation__.Notify(455569)
	s.mu.Lock()
	s.mu.rowsRead += int64(bat.Length())
	s.mu.Unlock()
	return bat
}

func (s *ColBatchScan) DrainMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(455574)
	var trailingMeta []execinfrapb.ProducerMetadata
	if !s.flowCtx.Local {
		__antithesis_instrumentation__.Notify(455578)
		nodeID, ok := s.flowCtx.NodeID.OptionalNodeID()
		if ok {
			__antithesis_instrumentation__.Notify(455579)
			ranges := execinfra.MisplannedRanges(s.Ctx, s.SpansCopy, nodeID, s.flowCtx.Cfg.RangeCache)
			if ranges != nil {
				__antithesis_instrumentation__.Notify(455580)
				trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{Ranges: ranges})
			} else {
				__antithesis_instrumentation__.Notify(455581)
			}
		} else {
			__antithesis_instrumentation__.Notify(455582)
		}
	} else {
		__antithesis_instrumentation__.Notify(455583)
	}
	__antithesis_instrumentation__.Notify(455575)
	if tfs := execinfra.GetLeafTxnFinalState(s.Ctx, s.flowCtx.Txn); tfs != nil {
		__antithesis_instrumentation__.Notify(455584)
		trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{LeafTxnFinalState: tfs})
	} else {
		__antithesis_instrumentation__.Notify(455585)
	}
	__antithesis_instrumentation__.Notify(455576)
	meta := execinfrapb.GetProducerMeta()
	meta.Metrics = execinfrapb.GetMetricsMeta()
	meta.Metrics.BytesRead = s.GetBytesRead()
	meta.Metrics.RowsRead = s.GetRowsRead()
	trailingMeta = append(trailingMeta, *meta)
	if trace := execinfra.GetTraceData(s.Ctx); trace != nil {
		__antithesis_instrumentation__.Notify(455586)
		trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{TraceData: trace})
	} else {
		__antithesis_instrumentation__.Notify(455587)
	}
	__antithesis_instrumentation__.Notify(455577)
	return trailingMeta
}

func (s *ColBatchScan) GetBytesRead() int64 {
	__antithesis_instrumentation__.Notify(455588)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cf.getBytesRead()
}

func (s *ColBatchScan) GetRowsRead() int64 {
	__antithesis_instrumentation__.Notify(455589)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.rowsRead
}

func (s *ColBatchScan) GetCumulativeContentionTime() time.Duration {
	__antithesis_instrumentation__.Notify(455590)
	return execinfra.GetCumulativeContentionTime(s.Ctx)
}

func (s *ColBatchScan) GetScanStats() execinfra.ScanStats {
	__antithesis_instrumentation__.Notify(455591)
	return execinfra.GetScanStats(s.Ctx)
}

var colBatchScanPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(455592)
		return &ColBatchScan{}
	},
}

func NewColBatchScan(
	ctx context.Context,
	allocator *colmem.Allocator,
	kvFetcherMemAcc *mon.BoundAccount,
	flowCtx *execinfra.FlowCtx,
	spec *execinfrapb.TableReaderSpec,
	post *execinfrapb.PostProcessSpec,
	estimatedRowCount uint64,
) (*ColBatchScan, error) {
	__antithesis_instrumentation__.Notify(455593)

	if nodeID, ok := flowCtx.NodeID.OptionalNodeID(); nodeID == 0 && func() bool {
		__antithesis_instrumentation__.Notify(455601)
		return ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(455602)
		return nil, errors.Errorf("attempting to create a ColBatchScan with uninitialized NodeID")
	} else {
		__antithesis_instrumentation__.Notify(455603)
	}
	__antithesis_instrumentation__.Notify(455594)
	limitHint := rowinfra.RowLimit(execinfra.LimitHint(spec.LimitHint, post))
	tableArgs, err := populateTableArgs(ctx, flowCtx, &spec.FetchSpec)
	if err != nil {
		__antithesis_instrumentation__.Notify(455604)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455605)
	}
	__antithesis_instrumentation__.Notify(455595)

	fetcher := cFetcherPool.Get().(*cFetcher)
	fetcher.cFetcherArgs = cFetcherArgs{
		spec.LockingStrength,
		spec.LockingWaitPolicy,
		flowCtx.EvalCtx.SessionData().LockTimeout,
		execinfra.GetWorkMemLimit(flowCtx),
		estimatedRowCount,
		spec.Reverse,
		flowCtx.TraceKV,
	}

	if err = fetcher.Init(allocator, kvFetcherMemAcc, tableArgs); err != nil {
		__antithesis_instrumentation__.Notify(455606)
		fetcher.Release()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455607)
	}
	__antithesis_instrumentation__.Notify(455596)

	var bsHeader *roachpb.BoundedStalenessHeader
	if aost := flowCtx.EvalCtx.AsOfSystemTime; aost != nil && func() bool {
		__antithesis_instrumentation__.Notify(455608)
		return aost.BoundedStaleness == true
	}() == true {
		__antithesis_instrumentation__.Notify(455609)
		ts := aost.Timestamp

		if aost.Timestamp.Less(spec.TableDescriptorModificationTime) {
			__antithesis_instrumentation__.Notify(455611)
			ts = spec.TableDescriptorModificationTime
		} else {
			__antithesis_instrumentation__.Notify(455612)
		}
		__antithesis_instrumentation__.Notify(455610)
		bsHeader = &roachpb.BoundedStalenessHeader{
			MinTimestampBound:       ts,
			MinTimestampBoundStrict: aost.NearestOnly,
			MaxTimestampBound:       flowCtx.EvalCtx.AsOfSystemTime.MaxTimestampBound,
		}
	} else {
		__antithesis_instrumentation__.Notify(455613)
	}
	__antithesis_instrumentation__.Notify(455597)

	s := colBatchScanPool.Get().(*ColBatchScan)
	s.Spans = spec.Spans
	if !flowCtx.Local {
		__antithesis_instrumentation__.Notify(455614)

		allocator.AdjustMemoryUsage(s.Spans.MemUsage())
		s.MakeSpansCopy()
	} else {
		__antithesis_instrumentation__.Notify(455615)
	}
	__antithesis_instrumentation__.Notify(455598)

	if spec.LimitHint > 0 || func() bool {
		__antithesis_instrumentation__.Notify(455616)
		return spec.BatchBytesLimit > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(455617)

		spec.Parallelize = false
	} else {
		__antithesis_instrumentation__.Notify(455618)
	}
	__antithesis_instrumentation__.Notify(455599)
	var batchBytesLimit rowinfra.BytesLimit
	if !spec.Parallelize {
		__antithesis_instrumentation__.Notify(455619)
		batchBytesLimit = rowinfra.BytesLimit(spec.BatchBytesLimit)
		if batchBytesLimit == 0 {
			__antithesis_instrumentation__.Notify(455620)
			batchBytesLimit = rowinfra.DefaultBatchBytesLimit
		} else {
			__antithesis_instrumentation__.Notify(455621)
		}
	} else {
		__antithesis_instrumentation__.Notify(455622)
	}
	__antithesis_instrumentation__.Notify(455600)

	*s = ColBatchScan{
		SpansWithCopy:   s.SpansWithCopy,
		flowCtx:         flowCtx,
		bsHeader:        bsHeader,
		cf:              fetcher,
		limitHint:       limitHint,
		batchBytesLimit: batchBytesLimit,
		parallelize:     spec.Parallelize,
		ResultTypes:     tableArgs.typs,
	}
	return s, nil
}

func (s *ColBatchScan) Release() {
	__antithesis_instrumentation__.Notify(455623)
	s.cf.Release()

	s.SpansWithCopy.Reset()
	*s = ColBatchScan{
		SpansWithCopy: s.SpansWithCopy,
	}
	colBatchScanPool.Put(s)
}

func (s *ColBatchScan) Close(context.Context) error {
	__antithesis_instrumentation__.Notify(455624)

	ctx := s.EnsureCtx()
	s.cf.Close(ctx)
	if s.tracingSpan != nil {
		__antithesis_instrumentation__.Notify(455626)
		s.tracingSpan.Finish()
		s.tracingSpan = nil
	} else {
		__antithesis_instrumentation__.Notify(455627)
	}
	__antithesis_instrumentation__.Notify(455625)
	return nil
}
