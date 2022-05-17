package colfetcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecspan"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/kvstreamer"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type ColIndexJoin struct {
	colexecop.InitHelper
	colexecop.OneInputNode

	state indexJoinState

	spanAssembler colexecspan.ColSpanAssembler

	batch coldata.Batch

	startIdx int

	limitHintHelper execinfra.LimitHintHelper

	mem struct {
		inputBatchSize int64

		inputBatchSizeLimit int64

		currentBatchSize int64

		constRowSize int64

		hasVarSizeCols bool
		varSizeVecIdxs util.FastIntSet
		byteLikeCols   []*coldata.Bytes
		decimalCols    []coldata.Decimals
		datumCols      []coldata.DatumVec
	}

	flowCtx *execinfra.FlowCtx
	cf      *cFetcher

	tracingSpan *tracing.Span
	mu          struct {
		syncutil.Mutex

		rowsRead int64
	}

	ResultTypes []*types.T

	maintainOrdering bool

	usesStreamer bool
	streamerInfo struct {
		*kvstreamer.Streamer
		budgetAcc   *mon.BoundAccount
		budgetLimit int64
		diskMonitor *mon.BytesMonitor
	}
}

var _ colexecop.KVReader = &ColIndexJoin{}
var _ execinfra.Releasable = &ColIndexJoin{}
var _ colexecop.ClosableOperator = &ColIndexJoin{}

func (s *ColIndexJoin) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(455634)
	if !s.InitHelper.Init(ctx) {
		__antithesis_instrumentation__.Notify(455636)
		return
	} else {
		__antithesis_instrumentation__.Notify(455637)
	}
	__antithesis_instrumentation__.Notify(455635)

	s.Ctx, s.tracingSpan = execinfra.ProcessorSpan(s.Ctx, "colindexjoin")
	s.Input.Init(s.Ctx)
	if s.usesStreamer {
		__antithesis_instrumentation__.Notify(455638)
		s.streamerInfo.Streamer = kvstreamer.NewStreamer(
			s.flowCtx.Cfg.DistSender,
			s.flowCtx.Stopper(),
			s.flowCtx.Txn,
			s.flowCtx.EvalCtx.Settings,
			row.GetWaitPolicy(s.cf.lockWaitPolicy),
			s.streamerInfo.budgetLimit,
			s.streamerInfo.budgetAcc,
		)
		mode := kvstreamer.OutOfOrder
		if s.maintainOrdering {
			__antithesis_instrumentation__.Notify(455640)
			mode = kvstreamer.InOrder
		} else {
			__antithesis_instrumentation__.Notify(455641)
		}
		__antithesis_instrumentation__.Notify(455639)
		s.streamerInfo.Streamer.Init(
			mode,
			kvstreamer.Hints{UniqueRequests: true},
			int(s.cf.table.spec.MaxKeysPerRow),
			s.flowCtx.Cfg.TempStorage,
			s.streamerInfo.diskMonitor,
		)
	} else {
		__antithesis_instrumentation__.Notify(455642)
	}
}

type indexJoinState uint8

const (
	indexJoinConstructingSpans indexJoinState = iota
	indexJoinScanning
	indexJoinDone
)

func (s *ColIndexJoin) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(455643)
	for {
		__antithesis_instrumentation__.Notify(455644)
		switch s.state {
		case indexJoinConstructingSpans:
			__antithesis_instrumentation__.Notify(455645)
			var rowCount int64
			var spans roachpb.Spans
			s.mem.inputBatchSize = 0
			for s.next() {
				__antithesis_instrumentation__.Notify(455658)

				endIdx := s.findEndIndex(rowCount > 0)

				if l := s.limitHintHelper.LimitHint(); l != 0 && func() bool {
					__antithesis_instrumentation__.Notify(455661)
					return rowCount+int64(endIdx-s.startIdx) > l == true
				}() == true {
					__antithesis_instrumentation__.Notify(455662)
					endIdx = s.startIdx + int(l-rowCount)
				} else {
					__antithesis_instrumentation__.Notify(455663)
				}
				__antithesis_instrumentation__.Notify(455659)
				rowCount += int64(endIdx - s.startIdx)
				s.spanAssembler.ConsumeBatch(s.batch, s.startIdx, endIdx)
				s.startIdx = endIdx
				if l := s.limitHintHelper.LimitHint(); l != 0 && func() bool {
					__antithesis_instrumentation__.Notify(455664)
					return rowCount == l == true
				}() == true {
					__antithesis_instrumentation__.Notify(455665)

					break
				} else {
					__antithesis_instrumentation__.Notify(455666)
				}
				__antithesis_instrumentation__.Notify(455660)
				if endIdx < s.batch.Length() {
					__antithesis_instrumentation__.Notify(455667)

					break
				} else {
					__antithesis_instrumentation__.Notify(455668)
				}
			}
			__antithesis_instrumentation__.Notify(455646)
			if err := s.limitHintHelper.ReadSomeRows(rowCount); err != nil {
				__antithesis_instrumentation__.Notify(455669)
				colexecerror.InternalError(err)
			} else {
				__antithesis_instrumentation__.Notify(455670)
			}
			__antithesis_instrumentation__.Notify(455647)
			spans = s.spanAssembler.GetSpans()
			if len(spans) == 0 {
				__antithesis_instrumentation__.Notify(455671)

				s.state = indexJoinDone
				continue
			} else {
				__antithesis_instrumentation__.Notify(455672)
			}
			__antithesis_instrumentation__.Notify(455648)

			if !s.maintainOrdering {
				__antithesis_instrumentation__.Notify(455673)

				sort.Sort(spans)
			} else {
				__antithesis_instrumentation__.Notify(455674)
			}
			__antithesis_instrumentation__.Notify(455649)

			s.cf.setEstimatedRowCount(uint64(rowCount))

			var err error
			if s.usesStreamer {
				__antithesis_instrumentation__.Notify(455675)
				err = s.cf.StartScanStreaming(
					s.Ctx,
					s.streamerInfo.Streamer,
					spans,
					rowinfra.NoRowLimit,
				)
			} else {
				__antithesis_instrumentation__.Notify(455676)
				err = s.cf.StartScan(
					s.Ctx,
					s.flowCtx.Txn,
					spans,
					nil,
					false,
					rowinfra.NoBytesLimit,
					rowinfra.NoRowLimit,
					s.flowCtx.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
				)
			}
			__antithesis_instrumentation__.Notify(455650)
			if err != nil {
				__antithesis_instrumentation__.Notify(455677)
				colexecerror.InternalError(err)
			} else {
				__antithesis_instrumentation__.Notify(455678)
			}
			__antithesis_instrumentation__.Notify(455651)
			s.state = indexJoinScanning
		case indexJoinScanning:
			__antithesis_instrumentation__.Notify(455652)
			batch, err := s.cf.NextBatch(s.Ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(455679)
				colexecerror.InternalError(err)
			} else {
				__antithesis_instrumentation__.Notify(455680)
			}
			__antithesis_instrumentation__.Notify(455653)
			if batch.Selection() != nil {
				__antithesis_instrumentation__.Notify(455681)
				colexecerror.InternalError(
					errors.AssertionFailedf("unexpected selection vector on the batch coming from CFetcher"))
			} else {
				__antithesis_instrumentation__.Notify(455682)
			}
			__antithesis_instrumentation__.Notify(455654)
			n := batch.Length()
			if n == 0 {
				__antithesis_instrumentation__.Notify(455683)

				s.spanAssembler.AccountForSpans()
				s.state = indexJoinConstructingSpans
				continue
			} else {
				__antithesis_instrumentation__.Notify(455684)
			}
			__antithesis_instrumentation__.Notify(455655)
			s.mu.Lock()
			s.mu.rowsRead += int64(n)
			s.mu.Unlock()
			return batch
		case indexJoinDone:
			__antithesis_instrumentation__.Notify(455656)

			s.closeInternal()
			return coldata.ZeroBatch
		default:
			__antithesis_instrumentation__.Notify(455657)
		}
	}
}

func (s *ColIndexJoin) findEndIndex(hasSpans bool) (endIdx int) {
	__antithesis_instrumentation__.Notify(455685)
	n := s.batch.Length()
	if n == 0 || func() bool {
		__antithesis_instrumentation__.Notify(455690)
		return s.startIdx >= n == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(455691)
		return s.mem.inputBatchSize >= s.mem.inputBatchSizeLimit == true
	}() == true {
		__antithesis_instrumentation__.Notify(455692)

		return s.startIdx
	} else {
		__antithesis_instrumentation__.Notify(455693)
	}
	__antithesis_instrumentation__.Notify(455686)
	if s.mem.inputBatchSize+s.mem.currentBatchSize <= s.mem.inputBatchSizeLimit {
		__antithesis_instrumentation__.Notify(455694)

		s.mem.inputBatchSize += s.mem.currentBatchSize
		return n
	} else {
		__antithesis_instrumentation__.Notify(455695)
	}
	__antithesis_instrumentation__.Notify(455687)
	for endIdx = s.startIdx; endIdx < n; endIdx++ {
		__antithesis_instrumentation__.Notify(455696)
		s.mem.inputBatchSize += s.getRowSize(endIdx)
		if s.mem.inputBatchSize > s.mem.inputBatchSizeLimit {
			__antithesis_instrumentation__.Notify(455698)

			break
		} else {
			__antithesis_instrumentation__.Notify(455699)
		}
		__antithesis_instrumentation__.Notify(455697)
		if s.mem.inputBatchSize == s.mem.inputBatchSizeLimit {
			__antithesis_instrumentation__.Notify(455700)

			endIdx++
			break
		} else {
			__antithesis_instrumentation__.Notify(455701)
		}
	}
	__antithesis_instrumentation__.Notify(455688)
	if !hasSpans && func() bool {
		__antithesis_instrumentation__.Notify(455702)
		return endIdx == s.startIdx == true
	}() == true {
		__antithesis_instrumentation__.Notify(455703)

		return s.startIdx + 1
	} else {
		__antithesis_instrumentation__.Notify(455704)
	}
	__antithesis_instrumentation__.Notify(455689)
	return endIdx
}

func (s *ColIndexJoin) getRowSize(idx int) int64 {
	__antithesis_instrumentation__.Notify(455705)
	rowSize := s.mem.constRowSize
	if s.mem.hasVarSizeCols {
		__antithesis_instrumentation__.Notify(455707)
		for i := range s.mem.byteLikeCols {
			__antithesis_instrumentation__.Notify(455710)
			rowSize += adjustMemEstimate(s.mem.byteLikeCols[i].ElemSize(idx))
		}
		__antithesis_instrumentation__.Notify(455708)
		for i := range s.mem.decimalCols {
			__antithesis_instrumentation__.Notify(455711)
			rowSize += adjustMemEstimate(int64(s.mem.decimalCols[i][idx].Size()))
		}
		__antithesis_instrumentation__.Notify(455709)
		for i := range s.mem.datumCols {
			__antithesis_instrumentation__.Notify(455712)
			memEstimate := int64(s.mem.datumCols[i].Get(idx).(tree.Datum).Size()) + memsize.DatumOverhead
			rowSize += adjustMemEstimate(memEstimate)
		}
	} else {
		__antithesis_instrumentation__.Notify(455713)
	}
	__antithesis_instrumentation__.Notify(455706)
	return rowSize
}

func (s *ColIndexJoin) getBatchSize() int64 {
	__antithesis_instrumentation__.Notify(455714)
	n := s.batch.Length()
	batchSize := colmem.GetBatchMemSize(s.batch)
	batchSize += int64(n*s.batch.Width()) * memEstimateAdditive
	batchSize += int64(n) * int64(rowenc.EncDatumRowOverhead)
	return batchSize
}

func (s *ColIndexJoin) next() bool {
	__antithesis_instrumentation__.Notify(455715)
	if s.batch == nil || func() bool {
		__antithesis_instrumentation__.Notify(455719)
		return s.startIdx >= s.batch.Length() == true
	}() == true {
		__antithesis_instrumentation__.Notify(455720)

		s.startIdx = 0
		s.batch = s.Input.Next()
		if s.batch.Length() == 0 {
			__antithesis_instrumentation__.Notify(455722)
			return false
		} else {
			__antithesis_instrumentation__.Notify(455723)
		}
		__antithesis_instrumentation__.Notify(455721)
		s.mem.currentBatchSize = s.getBatchSize()
	} else {
		__antithesis_instrumentation__.Notify(455724)
	}
	__antithesis_instrumentation__.Notify(455716)
	if !s.mem.hasVarSizeCols {
		__antithesis_instrumentation__.Notify(455725)
		return true
	} else {
		__antithesis_instrumentation__.Notify(455726)
	}
	__antithesis_instrumentation__.Notify(455717)
	s.mem.byteLikeCols = s.mem.byteLikeCols[:0]
	s.mem.decimalCols = s.mem.decimalCols[:0]
	s.mem.datumCols = s.mem.datumCols[:0]
	for i, ok := s.mem.varSizeVecIdxs.Next(0); ok; i, ok = s.mem.varSizeVecIdxs.Next(i + 1) {
		__antithesis_instrumentation__.Notify(455727)
		vec := s.batch.ColVec(i)
		switch vec.CanonicalTypeFamily() {
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(455728)
			s.mem.byteLikeCols = append(s.mem.byteLikeCols, vec.Bytes())
		case types.JsonFamily:
			__antithesis_instrumentation__.Notify(455729)
			s.mem.byteLikeCols = append(s.mem.byteLikeCols, &vec.JSON().Bytes)
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(455730)
			s.mem.decimalCols = append(s.mem.decimalCols, vec.Decimal())
		case typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(455731)
			s.mem.datumCols = append(s.mem.datumCols, vec.Datum())
		default:
			__antithesis_instrumentation__.Notify(455732)
		}
	}
	__antithesis_instrumentation__.Notify(455718)
	return true
}

func (s *ColIndexJoin) DrainMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(455733)
	var trailingMeta []execinfrapb.ProducerMetadata
	if tfs := execinfra.GetLeafTxnFinalState(s.Ctx, s.flowCtx.Txn); tfs != nil {
		__antithesis_instrumentation__.Notify(455736)
		trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{LeafTxnFinalState: tfs})
	} else {
		__antithesis_instrumentation__.Notify(455737)
	}
	__antithesis_instrumentation__.Notify(455734)
	meta := execinfrapb.GetProducerMeta()
	meta.Metrics = execinfrapb.GetMetricsMeta()
	meta.Metrics.BytesRead = s.GetBytesRead()
	meta.Metrics.RowsRead = s.GetRowsRead()
	trailingMeta = append(trailingMeta, *meta)
	if trace := execinfra.GetTraceData(s.Ctx); trace != nil {
		__antithesis_instrumentation__.Notify(455738)
		trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{TraceData: trace})
	} else {
		__antithesis_instrumentation__.Notify(455739)
	}
	__antithesis_instrumentation__.Notify(455735)
	return trailingMeta
}

func (s *ColIndexJoin) GetBytesRead() int64 {
	__antithesis_instrumentation__.Notify(455740)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cf.getBytesRead()
}

func (s *ColIndexJoin) GetRowsRead() int64 {
	__antithesis_instrumentation__.Notify(455741)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.rowsRead
}

func (s *ColIndexJoin) GetCumulativeContentionTime() time.Duration {
	__antithesis_instrumentation__.Notify(455742)
	return execinfra.GetCumulativeContentionTime(s.Ctx)
}

var inputBatchSizeLimit = int64(util.ConstantWithMetamorphicTestRange(
	"ColIndexJoin-batch-size",
	productionIndexJoinBatchSize,
	1,
	productionIndexJoinBatchSize,
))

const productionIndexJoinBatchSize = 4 << 20

func getIndexJoinBatchSize(forceProductionValue bool) int64 {
	__antithesis_instrumentation__.Notify(455743)
	if forceProductionValue {
		__antithesis_instrumentation__.Notify(455745)
		return productionIndexJoinBatchSize
	} else {
		__antithesis_instrumentation__.Notify(455746)
	}
	__antithesis_instrumentation__.Notify(455744)
	return inputBatchSizeLimit
}

func NewColIndexJoin(
	ctx context.Context,
	allocator *colmem.Allocator,
	fetcherAllocator *colmem.Allocator,
	kvFetcherMemAcc *mon.BoundAccount,
	streamerBudgetAcc *mon.BoundAccount,
	flowCtx *execinfra.FlowCtx,
	input colexecop.Operator,
	spec *execinfrapb.JoinReaderSpec,
	post *execinfrapb.PostProcessSpec,
	inputTypes []*types.T,
	diskMonitor *mon.BytesMonitor,
) (*ColIndexJoin, error) {
	__antithesis_instrumentation__.Notify(455747)

	if nodeID, ok := flowCtx.NodeID.OptionalNodeID(); nodeID == 0 && func() bool {
		__antithesis_instrumentation__.Notify(455756)
		return ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(455757)
		return nil, errors.Errorf("attempting to create a ColIndexJoin with uninitialized NodeID")
	} else {
		__antithesis_instrumentation__.Notify(455758)
	}
	__antithesis_instrumentation__.Notify(455748)
	if !spec.LookupExpr.Empty() {
		__antithesis_instrumentation__.Notify(455759)
		return nil, errors.AssertionFailedf("non-empty lookup expressions are not supported for index joins")
	} else {
		__antithesis_instrumentation__.Notify(455760)
	}
	__antithesis_instrumentation__.Notify(455749)
	if !spec.RemoteLookupExpr.Empty() {
		__antithesis_instrumentation__.Notify(455761)
		return nil, errors.AssertionFailedf("non-empty remote lookup expressions are not supported for index joins")
	} else {
		__antithesis_instrumentation__.Notify(455762)
	}
	__antithesis_instrumentation__.Notify(455750)
	if !spec.OnExpr.Empty() {
		__antithesis_instrumentation__.Notify(455763)
		return nil, errors.AssertionFailedf("non-empty ON expressions are not supported for index joins")
	} else {
		__antithesis_instrumentation__.Notify(455764)
	}
	__antithesis_instrumentation__.Notify(455751)

	tableArgs, err := populateTableArgs(ctx, flowCtx, &spec.FetchSpec)
	if err != nil {
		__antithesis_instrumentation__.Notify(455765)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455766)
	}
	__antithesis_instrumentation__.Notify(455752)

	memoryLimit := execinfra.GetWorkMemLimit(flowCtx)

	useStreamer := flowCtx.Txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(455767)
		return flowCtx.Txn.Type() == kv.LeafTxn == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(455768)
		return row.CanUseStreamer(ctx, flowCtx.EvalCtx.Settings) == true
	}() == true
	if useStreamer {
		__antithesis_instrumentation__.Notify(455769)
		if streamerBudgetAcc == nil {
			__antithesis_instrumentation__.Notify(455771)
			return nil, errors.AssertionFailedf("streamer budget account is nil when the Streamer API is desired")
		} else {
			__antithesis_instrumentation__.Notify(455772)
		}
		__antithesis_instrumentation__.Notify(455770)

		memoryLimit = int64(math.Ceil(float64(memoryLimit) / 4.0))
	} else {
		__antithesis_instrumentation__.Notify(455773)
	}
	__antithesis_instrumentation__.Notify(455753)

	fetcher := cFetcherPool.Get().(*cFetcher)
	fetcher.cFetcherArgs = cFetcherArgs{
		spec.LockingStrength,
		spec.LockingWaitPolicy,
		flowCtx.EvalCtx.SessionData().LockTimeout,
		memoryLimit,

		0,
		false,
		flowCtx.TraceKV,
	}
	if err = fetcher.Init(
		fetcherAllocator, kvFetcherMemAcc, tableArgs,
	); err != nil {
		__antithesis_instrumentation__.Notify(455774)
		fetcher.Release()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(455775)
	}
	__antithesis_instrumentation__.Notify(455754)

	spanAssembler := colexecspan.NewColSpanAssembler(
		flowCtx.Codec(), allocator, &spec.FetchSpec, spec.SplitFamilyIDs, inputTypes,
	)

	op := &ColIndexJoin{
		OneInputNode:     colexecop.NewOneInputNode(input),
		flowCtx:          flowCtx,
		cf:               fetcher,
		spanAssembler:    spanAssembler,
		ResultTypes:      tableArgs.typs,
		maintainOrdering: spec.MaintainOrdering,
		usesStreamer:     useStreamer,
		limitHintHelper:  execinfra.MakeLimitHintHelper(spec.LimitHint, post),
	}
	op.mem.inputBatchSizeLimit = getIndexJoinBatchSize(flowCtx.EvalCtx.TestingKnobs.ForceProductionBatchSizes)
	op.prepareMemLimit(inputTypes)
	if useStreamer {
		__antithesis_instrumentation__.Notify(455776)
		op.streamerInfo.budgetLimit = 3 * memoryLimit
		op.streamerInfo.budgetAcc = streamerBudgetAcc
		if spec.MaintainOrdering && func() bool {
			__antithesis_instrumentation__.Notify(455778)
			return diskMonitor == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(455779)
			return nil, errors.AssertionFailedf("diskMonitor is nil when ordering needs to be maintained")
		} else {
			__antithesis_instrumentation__.Notify(455780)
		}
		__antithesis_instrumentation__.Notify(455777)
		op.streamerInfo.diskMonitor = diskMonitor
		if memoryLimit < op.mem.inputBatchSizeLimit {
			__antithesis_instrumentation__.Notify(455781)

			op.mem.inputBatchSizeLimit = memoryLimit
		} else {
			__antithesis_instrumentation__.Notify(455782)
		}
	} else {
		__antithesis_instrumentation__.Notify(455783)
	}
	__antithesis_instrumentation__.Notify(455755)

	return op, nil
}

func (s *ColIndexJoin) prepareMemLimit(inputTypes []*types.T) {
	__antithesis_instrumentation__.Notify(455784)

	s.mem.constRowSize = int64(rowenc.EncDatumRowOverhead)
	for i, t := range inputTypes {
		__antithesis_instrumentation__.Notify(455786)
		switch typeconv.TypeFamilyToCanonicalTypeFamily(t.Family()) {
		case
			types.BoolFamily,
			types.IntFamily,
			types.FloatFamily,
			types.TimestampTZFamily,
			types.IntervalFamily:
			__antithesis_instrumentation__.Notify(455787)
			s.mem.constRowSize += adjustMemEstimate(colmem.GetFixedSizeTypeSize(t))
		case
			types.DecimalFamily,
			types.BytesFamily,
			types.JsonFamily,
			typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(455788)
			s.mem.varSizeVecIdxs.Add(i)
			s.mem.hasVarSizeCols = true
		default:
			__antithesis_instrumentation__.Notify(455789)
			colexecerror.InternalError(errors.AssertionFailedf("unhandled type %s", t))
		}
	}
	__antithesis_instrumentation__.Notify(455785)
	s.mem.hasVarSizeCols = !s.mem.varSizeVecIdxs.Empty()
}

var (
	memEstimateAdditive = int64(rowenc.EncDatumOverhead)

	memEstimateMultiplier = int64(2)
)

func adjustMemEstimate(estimate int64) int64 {
	__antithesis_instrumentation__.Notify(455790)
	return estimate*memEstimateMultiplier + memEstimateAdditive
}

func (s *ColIndexJoin) GetScanStats() execinfra.ScanStats {
	__antithesis_instrumentation__.Notify(455791)
	return execinfra.GetScanStats(s.Ctx)
}

func (s *ColIndexJoin) Release() {
	__antithesis_instrumentation__.Notify(455792)
	s.cf.Release()
	s.spanAssembler.Release()
	*s = ColIndexJoin{}
}

func (s *ColIndexJoin) Close(context.Context) error {
	__antithesis_instrumentation__.Notify(455793)
	s.closeInternal()
	if s.tracingSpan != nil {
		__antithesis_instrumentation__.Notify(455795)
		s.tracingSpan.Finish()
		s.tracingSpan = nil
	} else {
		__antithesis_instrumentation__.Notify(455796)
	}
	__antithesis_instrumentation__.Notify(455794)
	return nil
}

func (s *ColIndexJoin) closeInternal() {
	__antithesis_instrumentation__.Notify(455797)

	ctx := s.EnsureCtx()
	s.cf.Close(ctx)
	if s.spanAssembler != nil {
		__antithesis_instrumentation__.Notify(455800)

		s.spanAssembler.Close()
	} else {
		__antithesis_instrumentation__.Notify(455801)
	}
	__antithesis_instrumentation__.Notify(455798)
	if s.streamerInfo.Streamer != nil {
		__antithesis_instrumentation__.Notify(455802)
		s.streamerInfo.Streamer.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(455803)
	}
	__antithesis_instrumentation__.Notify(455799)
	s.batch = nil
}
