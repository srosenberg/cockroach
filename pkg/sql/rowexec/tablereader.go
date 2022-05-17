package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

type tableReader struct {
	execinfra.ProcessorBase
	execinfra.SpansWithCopy

	limitHint       rowinfra.RowLimit
	parallelize     bool
	batchBytesLimit rowinfra.BytesLimit

	scanStarted bool

	maxTimestampAge time.Duration

	ignoreMisplannedRanges bool

	fetcher rowFetcher
	alloc   tree.DatumAlloc

	scanStats execinfra.ScanStats

	rowsRead int64
}

var _ execinfra.Processor = &tableReader{}
var _ execinfra.RowSource = &tableReader{}
var _ execinfra.Releasable = &tableReader{}
var _ execinfra.OpNode = &tableReader{}

const tableReaderProcName = "table reader"

var trPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(575125)
		return &tableReader{}
	},
}

func newTableReader(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.TableReaderSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*tableReader, error) {
	__antithesis_instrumentation__.Notify(575126)

	if nodeID, ok := flowCtx.NodeID.OptionalNodeID(); ok && func() bool {
		__antithesis_instrumentation__.Notify(575136)
		return nodeID == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(575137)
		return nil, errors.Errorf("attempting to create a tableReader with uninitialized NodeID")
	} else {
		__antithesis_instrumentation__.Notify(575138)
	}
	__antithesis_instrumentation__.Notify(575127)

	if spec.LimitHint > 0 || func() bool {
		__antithesis_instrumentation__.Notify(575139)
		return spec.BatchBytesLimit > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(575140)

		spec.Parallelize = false
	} else {
		__antithesis_instrumentation__.Notify(575141)
	}
	__antithesis_instrumentation__.Notify(575128)
	var batchBytesLimit rowinfra.BytesLimit
	if !spec.Parallelize {
		__antithesis_instrumentation__.Notify(575142)
		batchBytesLimit = rowinfra.BytesLimit(spec.BatchBytesLimit)
		if batchBytesLimit == 0 {
			__antithesis_instrumentation__.Notify(575143)
			batchBytesLimit = rowinfra.DefaultBatchBytesLimit
		} else {
			__antithesis_instrumentation__.Notify(575144)
		}
	} else {
		__antithesis_instrumentation__.Notify(575145)
	}
	__antithesis_instrumentation__.Notify(575129)

	tr := trPool.Get().(*tableReader)

	tr.limitHint = rowinfra.RowLimit(execinfra.LimitHint(spec.LimitHint, post))
	tr.parallelize = spec.Parallelize
	tr.batchBytesLimit = batchBytesLimit
	tr.maxTimestampAge = time.Duration(spec.MaxTimestampAgeNanos)

	resolver := flowCtx.NewTypeResolver(flowCtx.Txn)
	for i := range spec.FetchSpec.KeyAndSuffixColumns {
		__antithesis_instrumentation__.Notify(575146)
		if err := typedesc.EnsureTypeIsHydrated(
			flowCtx.EvalCtx.Ctx(), spec.FetchSpec.KeyAndSuffixColumns[i].Type, &resolver,
		); err != nil {
			__antithesis_instrumentation__.Notify(575147)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575148)
		}
	}
	__antithesis_instrumentation__.Notify(575130)

	resultTypes := make([]*types.T, len(spec.FetchSpec.FetchedColumns))
	for i := range resultTypes {
		__antithesis_instrumentation__.Notify(575149)
		resultTypes[i] = spec.FetchSpec.FetchedColumns[i].Type
	}
	__antithesis_instrumentation__.Notify(575131)

	tr.ignoreMisplannedRanges = flowCtx.Local
	if err := tr.Init(
		tr,
		post,
		resultTypes,
		flowCtx,
		processorID,
		output,
		nil,
		execinfra.ProcStateOpts{

			InputsToDrain:        nil,
			TrailingMetaCallback: tr.generateTrailingMeta,
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(575150)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(575151)
	}
	__antithesis_instrumentation__.Notify(575132)

	var fetcher row.Fetcher
	if err := fetcher.Init(
		flowCtx.EvalCtx.Context,
		spec.Reverse,
		spec.LockingStrength,
		spec.LockingWaitPolicy,
		flowCtx.EvalCtx.SessionData().LockTimeout,
		&tr.alloc,
		flowCtx.EvalCtx.Mon,
		&spec.FetchSpec,
	); err != nil {
		__antithesis_instrumentation__.Notify(575152)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(575153)
	}
	__antithesis_instrumentation__.Notify(575133)

	tr.Spans = spec.Spans
	if !tr.ignoreMisplannedRanges {
		__antithesis_instrumentation__.Notify(575154)

		tr.MakeSpansCopy()
	} else {
		__antithesis_instrumentation__.Notify(575155)
	}
	__antithesis_instrumentation__.Notify(575134)

	if execinfra.ShouldCollectStats(flowCtx.EvalCtx.Ctx(), flowCtx) {
		__antithesis_instrumentation__.Notify(575156)
		tr.fetcher = newRowFetcherStatCollector(&fetcher)
		tr.ExecStatsForTrace = tr.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(575157)
		tr.fetcher = &fetcher
	}
	__antithesis_instrumentation__.Notify(575135)

	return tr, nil
}

func (tr *tableReader) generateTrailingMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(575158)

	trailingMeta := tr.generateMeta()
	tr.close()
	return trailingMeta
}

func (tr *tableReader) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575159)
	if tr.FlowCtx.Txn == nil {
		__antithesis_instrumentation__.Notify(575161)
		log.Fatalf(ctx, "tableReader outside of txn")
	} else {
		__antithesis_instrumentation__.Notify(575162)
	}
	__antithesis_instrumentation__.Notify(575160)

	ctx = tr.StartInternal(ctx, tableReaderProcName)

	_ = ctx
}

func (tr *tableReader) startScan(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(575163)
	limitBatches := !tr.parallelize
	var bytesLimit rowinfra.BytesLimit
	if !limitBatches {
		__antithesis_instrumentation__.Notify(575166)
		bytesLimit = rowinfra.NoBytesLimit
	} else {
		__antithesis_instrumentation__.Notify(575167)
		bytesLimit = tr.batchBytesLimit
	}
	__antithesis_instrumentation__.Notify(575164)
	log.VEventf(ctx, 1, "starting scan with limitBatches %t", limitBatches)
	var err error
	if tr.maxTimestampAge == 0 {
		__antithesis_instrumentation__.Notify(575168)
		err = tr.fetcher.StartScan(
			ctx, tr.FlowCtx.Txn, tr.Spans, bytesLimit, tr.limitHint,
			tr.FlowCtx.TraceKV,
			tr.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
		)
	} else {
		__antithesis_instrumentation__.Notify(575169)
		initialTS := tr.FlowCtx.Txn.ReadTimestamp()
		err = tr.fetcher.StartInconsistentScan(
			ctx, tr.FlowCtx.Cfg.DB, initialTS, tr.maxTimestampAge, tr.Spans,
			bytesLimit, tr.limitHint, tr.FlowCtx.TraceKV,
			tr.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
			tr.EvalCtx.QualityOfService(),
		)
	}
	__antithesis_instrumentation__.Notify(575165)
	tr.scanStarted = true
	return err
}

func (tr *tableReader) Release() {
	__antithesis_instrumentation__.Notify(575170)
	tr.ProcessorBase.Reset()
	tr.fetcher.Reset()

	tr.SpansWithCopy.Reset()
	*tr = tableReader{
		ProcessorBase: tr.ProcessorBase,
		SpansWithCopy: tr.SpansWithCopy,
		fetcher:       tr.fetcher,
		rowsRead:      0,
	}
	trPool.Put(tr)
}

var tableReaderProgressFrequency int64 = 5000

func TestingSetScannedRowProgressFrequency(val int64) func() {
	__antithesis_instrumentation__.Notify(575171)
	oldVal := tableReaderProgressFrequency
	tableReaderProgressFrequency = val
	return func() { __antithesis_instrumentation__.Notify(575172); tableReaderProgressFrequency = oldVal }
}

func (tr *tableReader) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575173)
	for tr.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(575175)
		if !tr.scanStarted {
			__antithesis_instrumentation__.Notify(575179)
			err := tr.startScan(tr.Ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(575180)
				tr.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(575181)
			}
		} else {
			__antithesis_instrumentation__.Notify(575182)
		}
		__antithesis_instrumentation__.Notify(575176)

		if tr.rowsRead >= tableReaderProgressFrequency {
			__antithesis_instrumentation__.Notify(575183)
			meta := execinfrapb.GetProducerMeta()
			meta.Metrics = execinfrapb.GetMetricsMeta()
			meta.Metrics.RowsRead = tr.rowsRead
			tr.rowsRead = 0
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(575184)
		}
		__antithesis_instrumentation__.Notify(575177)

		row, err := tr.fetcher.NextRow(tr.Ctx)
		if row == nil || func() bool {
			__antithesis_instrumentation__.Notify(575185)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(575186)
			tr.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(575187)
		}
		__antithesis_instrumentation__.Notify(575178)

		tr.rowsRead++
		if outRow := tr.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(575188)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(575189)
		}
	}
	__antithesis_instrumentation__.Notify(575174)
	return nil, tr.DrainHelper()
}

func (tr *tableReader) close() {
	__antithesis_instrumentation__.Notify(575190)
	if tr.InternalClose() {
		__antithesis_instrumentation__.Notify(575191)
		if tr.fetcher != nil {
			__antithesis_instrumentation__.Notify(575192)
			tr.fetcher.Close(tr.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(575193)
		}
	} else {
		__antithesis_instrumentation__.Notify(575194)
	}
}

func (tr *tableReader) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(575195)
	tr.close()
}

func (tr *tableReader) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(575196)
	is, ok := getFetcherInputStats(tr.fetcher)
	if !ok {
		__antithesis_instrumentation__.Notify(575198)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(575199)
	}
	__antithesis_instrumentation__.Notify(575197)
	tr.scanStats = execinfra.GetScanStats(tr.Ctx)
	ret := &execinfrapb.ComponentStats{
		KV: execinfrapb.KVStats{
			BytesRead:      optional.MakeUint(uint64(tr.fetcher.GetBytesRead())),
			TuplesRead:     is.NumTuples,
			KVTime:         is.WaitTime,
			ContentionTime: optional.MakeTimeValue(execinfra.GetCumulativeContentionTime(tr.Ctx)),
		},
		Output: tr.OutputHelper.Stats(),
	}
	execinfra.PopulateKVMVCCStats(&ret.KV, &tr.scanStats)
	return ret
}

func (tr *tableReader) generateMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(575200)
	var trailingMeta []execinfrapb.ProducerMetadata
	if !tr.ignoreMisplannedRanges {
		__antithesis_instrumentation__.Notify(575203)
		nodeID, ok := tr.FlowCtx.NodeID.OptionalNodeID()
		if ok {
			__antithesis_instrumentation__.Notify(575204)
			ranges := execinfra.MisplannedRanges(tr.Ctx, tr.SpansCopy, nodeID, tr.FlowCtx.Cfg.RangeCache)
			if ranges != nil {
				__antithesis_instrumentation__.Notify(575205)
				trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{Ranges: ranges})
			} else {
				__antithesis_instrumentation__.Notify(575206)
			}
		} else {
			__antithesis_instrumentation__.Notify(575207)
		}
	} else {
		__antithesis_instrumentation__.Notify(575208)
	}
	__antithesis_instrumentation__.Notify(575201)
	if tfs := execinfra.GetLeafTxnFinalState(tr.Ctx, tr.FlowCtx.Txn); tfs != nil {
		__antithesis_instrumentation__.Notify(575209)
		trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{LeafTxnFinalState: tfs})
	} else {
		__antithesis_instrumentation__.Notify(575210)
	}
	__antithesis_instrumentation__.Notify(575202)

	meta := execinfrapb.GetProducerMeta()
	meta.Metrics = execinfrapb.GetMetricsMeta()
	meta.Metrics.BytesRead = tr.fetcher.GetBytesRead()
	meta.Metrics.RowsRead = tr.rowsRead
	return append(trailingMeta, *meta)
}

func (tr *tableReader) ChildCount(bool) int {
	__antithesis_instrumentation__.Notify(575211)
	return 0
}

func (tr *tableReader) Child(nth int, _ bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(575212)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
