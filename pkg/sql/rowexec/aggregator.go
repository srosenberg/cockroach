package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/cockroach/pkg/util/stringarena"
	"github.com/cockroachdb/errors"
)

type aggregateFuncs []tree.AggregateFunc

func (af aggregateFuncs) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(571624)
	for _, f := range af {
		__antithesis_instrumentation__.Notify(571625)
		f.Close(ctx)
	}
}

type aggregatorBase struct {
	execinfra.ProcessorBase

	runningState aggregatorState
	input        execinfra.RowSource
	inputDone    bool
	inputTypes   []*types.T
	funcs        []*aggregateFuncHolder
	outputTypes  []*types.T
	datumAlloc   tree.DatumAlloc
	rowAlloc     rowenc.EncDatumRowAlloc

	bucketsAcc  mon.BoundAccount
	aggFuncsAcc mon.BoundAccount

	isScalar         bool
	groupCols        []uint32
	orderedGroupCols []uint32
	aggregations     []execinfrapb.AggregatorSpec_Aggregation

	lastOrdGroupCols rowenc.EncDatumRow
	arena            stringarena.Arena
	row              rowenc.EncDatumRow
	scratch          []byte

	cancelChecker cancelchecker.CancelChecker
}

func (ag *aggregatorBase) init(
	self execinfra.RowSource,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.AggregatorSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
	trailingMetaCallback func() []execinfrapb.ProducerMetadata,
) error {
	ctx := flowCtx.EvalCtx.Ctx()
	memMonitor := execinfra.NewMonitor(ctx, flowCtx.EvalCtx.Mon, "aggregator-mem")
	if execinfra.ShouldCollectStats(ctx, flowCtx) {
		input = newInputStatCollector(input)
		ag.ExecStatsForTrace = ag.execStatsForTrace
	}
	ag.input = input
	ag.isScalar = spec.IsScalar()
	ag.groupCols = spec.GroupCols
	ag.orderedGroupCols = spec.OrderedGroupCols
	ag.aggregations = spec.Aggregations
	ag.funcs = make([]*aggregateFuncHolder, len(spec.Aggregations))
	ag.outputTypes = make([]*types.T, len(spec.Aggregations))
	ag.row = make(rowenc.EncDatumRow, len(spec.Aggregations))
	ag.bucketsAcc = memMonitor.MakeBoundAccount()
	ag.arena = stringarena.Make(&ag.bucketsAcc)
	ag.aggFuncsAcc = memMonitor.MakeBoundAccount()

	ag.inputTypes = input.OutputTypes()
	semaCtx := flowCtx.NewSemaContext(flowCtx.EvalCtx.Txn)
	for i, aggInfo := range spec.Aggregations {
		if aggInfo.FilterColIdx != nil {
			col := *aggInfo.FilterColIdx
			if col >= uint32(len(ag.inputTypes)) {
				return errors.Errorf("FilterColIdx out of range (%d)", col)
			}
			t := ag.inputTypes[col].Family()
			if t != types.BoolFamily && t != types.UnknownFamily {
				return errors.Errorf(
					"filter column %d must be of boolean type, not %s", *aggInfo.FilterColIdx, t,
				)
			}
		}
		constructor, arguments, outputType, err := execinfra.GetAggregateConstructor(
			flowCtx.EvalCtx, semaCtx, &aggInfo, ag.inputTypes,
		)
		if err != nil {
			return err
		}
		ag.funcs[i] = ag.newAggregateFuncHolder(constructor, arguments)
		if aggInfo.Distinct {
			ag.funcs[i].seen = make(map[string]struct{})
		}
		ag.outputTypes[i] = outputType
	}

	return ag.ProcessorBase.Init(
		self, post, ag.outputTypes, flowCtx, processorID, output, memMonitor,
		execinfra.ProcStateOpts{
			InputsToDrain:        []execinfra.RowSource{ag.input},
			TrailingMetaCallback: trailingMetaCallback,
		},
	)
}

func (ag *aggregatorBase) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(571626)
	is, ok := getInputStats(ag.input)
	if !ok {
		__antithesis_instrumentation__.Notify(571628)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(571629)
	}
	__antithesis_instrumentation__.Notify(571627)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Exec: execinfrapb.ExecStats{
			MaxAllocatedMem: optional.MakeUint(uint64(ag.MemMonitor.MaximumBytes())),
		},
		Output: ag.OutputHelper.Stats(),
	}
}

func (ag *aggregatorBase) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(571630)
	if _, ok := ag.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(571632)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(571633)
	}
	__antithesis_instrumentation__.Notify(571631)
	return 0
}

func (ag *aggregatorBase) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(571634)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(571636)
		if n, ok := ag.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(571638)
			return n
		} else {
			__antithesis_instrumentation__.Notify(571639)
		}
		__antithesis_instrumentation__.Notify(571637)
		panic("input to aggregatorBase is not an execinfra.OpNode")
	} else {
		__antithesis_instrumentation__.Notify(571640)
	}
	__antithesis_instrumentation__.Notify(571635)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}

const (
	hashAggregatorBucketsInitialLen = 8
)

type hashAggregator struct {
	aggregatorBase

	buckets     map[string]aggregateFuncs
	bucketsIter []string

	bucketsLenGrowThreshold int

	alreadyAccountedFor int
}

type orderedAggregator struct {
	aggregatorBase

	bucket aggregateFuncs
}

var _ execinfra.Processor = &hashAggregator{}
var _ execinfra.RowSource = &hashAggregator{}
var _ execinfra.OpNode = &hashAggregator{}

const hashAggregatorProcName = "hash aggregator"

var _ execinfra.Processor = &orderedAggregator{}
var _ execinfra.RowSource = &orderedAggregator{}
var _ execinfra.OpNode = &orderedAggregator{}

const orderedAggregatorProcName = "ordered aggregator"

type aggregatorState int

const (
	aggStateUnknown aggregatorState = iota

	aggAccumulating

	aggEmittingRows
)

func newAggregator(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.AggregatorSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(571641)
	if spec.IsRowCount() {
		__antithesis_instrumentation__.Notify(571645)
		return newCountAggregator(flowCtx, processorID, input, post, output)
	} else {
		__antithesis_instrumentation__.Notify(571646)
	}
	__antithesis_instrumentation__.Notify(571642)
	if len(spec.OrderedGroupCols) == len(spec.GroupCols) {
		__antithesis_instrumentation__.Notify(571647)
		return newOrderedAggregator(flowCtx, processorID, spec, input, post, output)
	} else {
		__antithesis_instrumentation__.Notify(571648)
	}
	__antithesis_instrumentation__.Notify(571643)

	ag := &hashAggregator{
		buckets:                 make(map[string]aggregateFuncs),
		bucketsLenGrowThreshold: hashAggregatorBucketsInitialLen,
	}

	if err := ag.init(
		ag,
		flowCtx,
		processorID,
		spec,
		input,
		post,
		output,
		func() []execinfrapb.ProducerMetadata {
			__antithesis_instrumentation__.Notify(571649)
			ag.close()
			return nil
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(571650)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(571651)
	}
	__antithesis_instrumentation__.Notify(571644)

	ag.EvalCtx.SingleDatumAggMemAccount = &ag.aggFuncsAcc
	return ag, nil
}

func newOrderedAggregator(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.AggregatorSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*orderedAggregator, error) {
	__antithesis_instrumentation__.Notify(571652)
	ag := &orderedAggregator{}

	if err := ag.init(
		ag,
		flowCtx,
		processorID,
		spec,
		input,
		post,
		output,
		func() []execinfrapb.ProducerMetadata {
			__antithesis_instrumentation__.Notify(571654)
			ag.close()
			return nil
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(571655)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(571656)
	}
	__antithesis_instrumentation__.Notify(571653)

	ag.EvalCtx.SingleDatumAggMemAccount = &ag.aggFuncsAcc
	return ag, nil
}

func (ag *hashAggregator) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(571657)
	ag.start(ctx, hashAggregatorProcName)
}

func (ag *orderedAggregator) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(571658)
	ag.start(ctx, orderedAggregatorProcName)
}

func (ag *aggregatorBase) start(ctx context.Context, procName string) {
	__antithesis_instrumentation__.Notify(571659)
	ctx = ag.StartInternal(ctx, procName)
	ag.input.Start(ctx)
	ag.cancelChecker.Reset(ctx)
	ag.runningState = aggAccumulating
}

func (ag *hashAggregator) close() {
	__antithesis_instrumentation__.Notify(571660)
	if ag.InternalClose() {
		__antithesis_instrumentation__.Notify(571661)
		log.VEventf(ag.Ctx, 2, "exiting aggregator")

		if ag.bucketsIter == nil {
			__antithesis_instrumentation__.Notify(571663)
			for _, bucket := range ag.buckets {
				__antithesis_instrumentation__.Notify(571664)
				bucket.close(ag.Ctx)
			}
		} else {
			__antithesis_instrumentation__.Notify(571665)
			for _, bucket := range ag.bucketsIter {
				__antithesis_instrumentation__.Notify(571666)
				ag.buckets[bucket].close(ag.Ctx)
			}
		}
		__antithesis_instrumentation__.Notify(571662)

		ag.buckets = nil

		ag.bucketsAcc.Close(ag.Ctx)
		ag.aggFuncsAcc.Close(ag.Ctx)
		ag.MemMonitor.Stop(ag.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(571667)
	}
}

func (ag *orderedAggregator) close() {
	__antithesis_instrumentation__.Notify(571668)
	if ag.InternalClose() {
		__antithesis_instrumentation__.Notify(571669)
		log.VEventf(ag.Ctx, 2, "exiting aggregator")
		if ag.bucket != nil {
			__antithesis_instrumentation__.Notify(571671)
			ag.bucket.close(ag.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(571672)
		}
		__antithesis_instrumentation__.Notify(571670)

		ag.bucketsAcc.Close(ag.Ctx)
		ag.aggFuncsAcc.Close(ag.Ctx)
		ag.MemMonitor.Stop(ag.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(571673)
	}
}

func (ag *aggregatorBase) matchLastOrdGroupCols(row rowenc.EncDatumRow) (bool, error) {
	__antithesis_instrumentation__.Notify(571674)
	for _, colIdx := range ag.orderedGroupCols {
		__antithesis_instrumentation__.Notify(571676)
		res, err := ag.lastOrdGroupCols[colIdx].Compare(
			ag.inputTypes[colIdx], &ag.datumAlloc, ag.EvalCtx, &row[colIdx],
		)
		if res != 0 || func() bool {
			__antithesis_instrumentation__.Notify(571677)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(571678)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(571679)
		}
	}
	__antithesis_instrumentation__.Notify(571675)
	return true, nil
}

func (ag *hashAggregator) accumulateRows() (
	aggregatorState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(571680)
	for {
		__antithesis_instrumentation__.Notify(571685)
		row, meta := ag.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(571689)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(571691)
				ag.MoveToDraining(nil)
				return aggStateUnknown, nil, meta
			} else {
				__antithesis_instrumentation__.Notify(571692)
			}
			__antithesis_instrumentation__.Notify(571690)
			return aggAccumulating, nil, meta
		} else {
			__antithesis_instrumentation__.Notify(571693)
		}
		__antithesis_instrumentation__.Notify(571686)
		if row == nil {
			__antithesis_instrumentation__.Notify(571694)
			log.VEvent(ag.Ctx, 1, "accumulation complete")
			ag.inputDone = true
			break
		} else {
			__antithesis_instrumentation__.Notify(571695)
		}
		__antithesis_instrumentation__.Notify(571687)

		if ag.lastOrdGroupCols == nil {
			__antithesis_instrumentation__.Notify(571696)
			ag.lastOrdGroupCols = ag.rowAlloc.CopyRow(row)
		} else {
			__antithesis_instrumentation__.Notify(571697)
			matched, err := ag.matchLastOrdGroupCols(row)
			if err != nil {
				__antithesis_instrumentation__.Notify(571699)
				ag.MoveToDraining(err)
				return aggStateUnknown, nil, nil
			} else {
				__antithesis_instrumentation__.Notify(571700)
			}
			__antithesis_instrumentation__.Notify(571698)
			if !matched {
				__antithesis_instrumentation__.Notify(571701)
				copy(ag.lastOrdGroupCols, row)
				break
			} else {
				__antithesis_instrumentation__.Notify(571702)
			}
		}
		__antithesis_instrumentation__.Notify(571688)
		if err := ag.accumulateRow(row); err != nil {
			__antithesis_instrumentation__.Notify(571703)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571704)
		}
	}
	__antithesis_instrumentation__.Notify(571681)

	if len(ag.buckets) < 1 && func() bool {
		__antithesis_instrumentation__.Notify(571705)
		return len(ag.groupCols) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(571706)
		bucket, err := ag.createAggregateFuncs()
		if err != nil {
			__antithesis_instrumentation__.Notify(571708)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571709)
		}
		__antithesis_instrumentation__.Notify(571707)
		ag.buckets[""] = bucket
	} else {
		__antithesis_instrumentation__.Notify(571710)
	}
	__antithesis_instrumentation__.Notify(571682)

	if err := ag.bucketsAcc.Grow(ag.Ctx, int64(len(ag.buckets))*memsize.String); err != nil {
		__antithesis_instrumentation__.Notify(571711)
		ag.MoveToDraining(err)
		return aggStateUnknown, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(571712)
	}
	__antithesis_instrumentation__.Notify(571683)
	ag.bucketsIter = make([]string, 0, len(ag.buckets))
	for bucket := range ag.buckets {
		__antithesis_instrumentation__.Notify(571713)
		ag.bucketsIter = append(ag.bucketsIter, bucket)
	}
	__antithesis_instrumentation__.Notify(571684)

	return aggEmittingRows, nil, nil
}

func (ag *orderedAggregator) accumulateRows() (
	aggregatorState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(571714)
	for {
		__antithesis_instrumentation__.Notify(571717)
		row, meta := ag.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(571721)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(571723)
				ag.MoveToDraining(nil)
				return aggStateUnknown, nil, meta
			} else {
				__antithesis_instrumentation__.Notify(571724)
			}
			__antithesis_instrumentation__.Notify(571722)
			return aggAccumulating, nil, meta
		} else {
			__antithesis_instrumentation__.Notify(571725)
		}
		__antithesis_instrumentation__.Notify(571718)
		if row == nil {
			__antithesis_instrumentation__.Notify(571726)
			log.VEvent(ag.Ctx, 1, "accumulation complete")
			ag.inputDone = true
			break
		} else {
			__antithesis_instrumentation__.Notify(571727)
		}
		__antithesis_instrumentation__.Notify(571719)

		if ag.lastOrdGroupCols == nil {
			__antithesis_instrumentation__.Notify(571728)
			ag.lastOrdGroupCols = ag.rowAlloc.CopyRow(row)
		} else {
			__antithesis_instrumentation__.Notify(571729)
			matched, err := ag.matchLastOrdGroupCols(row)
			if err != nil {
				__antithesis_instrumentation__.Notify(571731)
				ag.MoveToDraining(err)
				return aggStateUnknown, nil, nil
			} else {
				__antithesis_instrumentation__.Notify(571732)
			}
			__antithesis_instrumentation__.Notify(571730)
			if !matched {
				__antithesis_instrumentation__.Notify(571733)
				copy(ag.lastOrdGroupCols, row)
				break
			} else {
				__antithesis_instrumentation__.Notify(571734)
			}
		}
		__antithesis_instrumentation__.Notify(571720)
		if err := ag.accumulateRow(row); err != nil {
			__antithesis_instrumentation__.Notify(571735)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571736)
		}
	}
	__antithesis_instrumentation__.Notify(571715)

	if ag.bucket == nil && func() bool {
		__antithesis_instrumentation__.Notify(571737)
		return ag.isScalar == true
	}() == true {
		__antithesis_instrumentation__.Notify(571738)
		var err error
		ag.bucket, err = ag.createAggregateFuncs()
		if err != nil {
			__antithesis_instrumentation__.Notify(571739)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571740)
		}
	} else {
		__antithesis_instrumentation__.Notify(571741)
	}
	__antithesis_instrumentation__.Notify(571716)

	return aggEmittingRows, nil, nil
}

func (ag *aggregatorBase) getAggResults(
	bucket aggregateFuncs,
) (aggregatorState, rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(571742)
	defer bucket.close(ag.Ctx)

	for i, b := range bucket {
		__antithesis_instrumentation__.Notify(571745)
		result, err := b.Result()
		if err != nil {
			__antithesis_instrumentation__.Notify(571748)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571749)
		}
		__antithesis_instrumentation__.Notify(571746)
		if result == nil {
			__antithesis_instrumentation__.Notify(571750)

			result = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(571751)
		}
		__antithesis_instrumentation__.Notify(571747)
		ag.row[i] = rowenc.DatumToEncDatum(ag.outputTypes[i], result)
	}
	__antithesis_instrumentation__.Notify(571743)

	if outRow := ag.ProcessRowHelper(ag.row); outRow != nil {
		__antithesis_instrumentation__.Notify(571752)
		return aggEmittingRows, outRow, nil
	} else {
		__antithesis_instrumentation__.Notify(571753)
	}
	__antithesis_instrumentation__.Notify(571744)

	return aggEmittingRows, nil, nil
}

func (ag *hashAggregator) emitRow() (
	aggregatorState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(571754)
	if len(ag.bucketsIter) == 0 {
		__antithesis_instrumentation__.Notify(571756)

		if ag.inputDone {
			__antithesis_instrumentation__.Notify(571761)

			ag.MoveToDraining(nil)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571762)
		}
		__antithesis_instrumentation__.Notify(571757)

		if err := ag.arena.UnsafeReset(ag.Ctx); err != nil {
			__antithesis_instrumentation__.Notify(571763)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571764)
		}
		__antithesis_instrumentation__.Notify(571758)

		ag.bucketsAcc.Shrink(ag.Ctx, int64(ag.alreadyAccountedFor)*memsize.MapEntryOverhead)

		ag.bucketsAcc.Shrink(ag.Ctx, int64(len(ag.buckets))*memsize.String)
		ag.bucketsIter = nil
		ag.buckets = make(map[string]aggregateFuncs)
		ag.bucketsLenGrowThreshold = hashAggregatorBucketsInitialLen
		ag.alreadyAccountedFor = 0
		for _, f := range ag.funcs {
			__antithesis_instrumentation__.Notify(571765)
			if f.seen != nil {
				__antithesis_instrumentation__.Notify(571766)

				for s := range f.seen {
					__antithesis_instrumentation__.Notify(571767)
					delete(f.seen, s)
				}
			} else {
				__antithesis_instrumentation__.Notify(571768)
			}
		}
		__antithesis_instrumentation__.Notify(571759)

		if err := ag.accumulateRow(ag.lastOrdGroupCols); err != nil {
			__antithesis_instrumentation__.Notify(571769)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571770)
		}
		__antithesis_instrumentation__.Notify(571760)

		return aggAccumulating, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(571771)
	}
	__antithesis_instrumentation__.Notify(571755)

	bucket := ag.bucketsIter[0]
	ag.bucketsIter = ag.bucketsIter[1:]

	state, row, meta := ag.getAggResults(ag.buckets[bucket])
	delete(ag.buckets, bucket)
	return state, row, meta
}

func (ag *orderedAggregator) emitRow() (
	aggregatorState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(571772)
	if ag.bucket == nil {
		__antithesis_instrumentation__.Notify(571774)

		if ag.inputDone {
			__antithesis_instrumentation__.Notify(571779)

			ag.MoveToDraining(nil)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571780)
		}
		__antithesis_instrumentation__.Notify(571775)

		if err := ag.arena.UnsafeReset(ag.Ctx); err != nil {
			__antithesis_instrumentation__.Notify(571781)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571782)
		}
		__antithesis_instrumentation__.Notify(571776)
		for _, f := range ag.funcs {
			__antithesis_instrumentation__.Notify(571783)
			if f.seen != nil {
				__antithesis_instrumentation__.Notify(571784)

				for s := range f.seen {
					__antithesis_instrumentation__.Notify(571785)
					delete(f.seen, s)
				}
			} else {
				__antithesis_instrumentation__.Notify(571786)
			}
		}
		__antithesis_instrumentation__.Notify(571777)

		if err := ag.accumulateRow(ag.lastOrdGroupCols); err != nil {
			__antithesis_instrumentation__.Notify(571787)
			ag.MoveToDraining(err)
			return aggStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(571788)
		}
		__antithesis_instrumentation__.Notify(571778)

		return aggAccumulating, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(571789)
	}
	__antithesis_instrumentation__.Notify(571773)

	bucket := ag.bucket
	ag.bucket = nil
	return ag.getAggResults(bucket)
}

func (ag *hashAggregator) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(571790)
	for ag.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(571792)
		var row rowenc.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch ag.runningState {
		case aggAccumulating:
			__antithesis_instrumentation__.Notify(571795)
			ag.runningState, row, meta = ag.accumulateRows()
		case aggEmittingRows:
			__antithesis_instrumentation__.Notify(571796)
			ag.runningState, row, meta = ag.emitRow()
		default:
			__antithesis_instrumentation__.Notify(571797)
			log.Fatalf(ag.Ctx, "unsupported state: %d", ag.runningState)
		}
		__antithesis_instrumentation__.Notify(571793)

		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(571798)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(571799)
			continue
		} else {
			__antithesis_instrumentation__.Notify(571800)
		}
		__antithesis_instrumentation__.Notify(571794)
		return row, meta
	}
	__antithesis_instrumentation__.Notify(571791)
	return nil, ag.DrainHelper()
}

func (ag *orderedAggregator) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(571801)
	for ag.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(571803)
		var row rowenc.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch ag.runningState {
		case aggAccumulating:
			__antithesis_instrumentation__.Notify(571806)
			ag.runningState, row, meta = ag.accumulateRows()
		case aggEmittingRows:
			__antithesis_instrumentation__.Notify(571807)
			ag.runningState, row, meta = ag.emitRow()
		default:
			__antithesis_instrumentation__.Notify(571808)
			log.Fatalf(ag.Ctx, "unsupported state: %d", ag.runningState)
		}
		__antithesis_instrumentation__.Notify(571804)

		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(571809)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(571810)
			continue
		} else {
			__antithesis_instrumentation__.Notify(571811)
		}
		__antithesis_instrumentation__.Notify(571805)
		return row, meta
	}
	__antithesis_instrumentation__.Notify(571802)
	return nil, ag.DrainHelper()
}

func (ag *hashAggregator) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(571812)

	ag.close()
}

func (ag *orderedAggregator) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(571813)

	ag.close()
}

func (ag *aggregatorBase) accumulateRowIntoBucket(
	row rowenc.EncDatumRow, groupKey []byte, bucket aggregateFuncs,
) error {
	__antithesis_instrumentation__.Notify(571814)
	var err error

	for i, a := range ag.aggregations {
		__antithesis_instrumentation__.Notify(571816)
		if a.FilterColIdx != nil {
			__antithesis_instrumentation__.Notify(571822)
			col := *a.FilterColIdx
			if err := row[col].EnsureDecoded(ag.inputTypes[col], &ag.datumAlloc); err != nil {
				__antithesis_instrumentation__.Notify(571824)
				return err
			} else {
				__antithesis_instrumentation__.Notify(571825)
			}
			__antithesis_instrumentation__.Notify(571823)
			if row[*a.FilterColIdx].Datum != tree.DBoolTrue {
				__antithesis_instrumentation__.Notify(571826)

				continue
			} else {
				__antithesis_instrumentation__.Notify(571827)
			}
		} else {
			__antithesis_instrumentation__.Notify(571828)
		}
		__antithesis_instrumentation__.Notify(571817)

		var firstArg tree.Datum
		var otherArgs tree.Datums
		if len(a.ColIdx) > 1 {
			__antithesis_instrumentation__.Notify(571829)
			otherArgs = make(tree.Datums, len(a.ColIdx)-1)
		} else {
			__antithesis_instrumentation__.Notify(571830)
		}
		__antithesis_instrumentation__.Notify(571818)
		isFirstArg := true
		for j, c := range a.ColIdx {
			__antithesis_instrumentation__.Notify(571831)
			if err := row[c].EnsureDecoded(ag.inputTypes[c], &ag.datumAlloc); err != nil {
				__antithesis_instrumentation__.Notify(571834)
				return err
			} else {
				__antithesis_instrumentation__.Notify(571835)
			}
			__antithesis_instrumentation__.Notify(571832)
			if isFirstArg {
				__antithesis_instrumentation__.Notify(571836)
				firstArg = row[c].Datum
				isFirstArg = false
				continue
			} else {
				__antithesis_instrumentation__.Notify(571837)
			}
			__antithesis_instrumentation__.Notify(571833)
			otherArgs[j-1] = row[c].Datum
		}
		__antithesis_instrumentation__.Notify(571819)

		canAdd := true
		if a.Distinct {
			__antithesis_instrumentation__.Notify(571838)
			canAdd, err = ag.funcs[i].isDistinct(
				ag.Ctx,
				&ag.datumAlloc,
				groupKey,
				firstArg,
				otherArgs,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(571839)
				return err
			} else {
				__antithesis_instrumentation__.Notify(571840)
			}
		} else {
			__antithesis_instrumentation__.Notify(571841)
		}
		__antithesis_instrumentation__.Notify(571820)
		if !canAdd {
			__antithesis_instrumentation__.Notify(571842)
			continue
		} else {
			__antithesis_instrumentation__.Notify(571843)
		}
		__antithesis_instrumentation__.Notify(571821)
		if err := bucket[i].Add(ag.Ctx, firstArg, otherArgs...); err != nil {
			__antithesis_instrumentation__.Notify(571844)
			return err
		} else {
			__antithesis_instrumentation__.Notify(571845)
		}
	}
	__antithesis_instrumentation__.Notify(571815)
	return nil
}

func (ag *hashAggregator) encode(
	appendTo []byte, row rowenc.EncDatumRow,
) (encoding []byte, err error) {
	__antithesis_instrumentation__.Notify(571846)
	for _, colIdx := range ag.groupCols {
		__antithesis_instrumentation__.Notify(571848)

		appendTo, err = row[colIdx].Fingerprint(
			ag.Ctx, ag.inputTypes[colIdx], &ag.datumAlloc, appendTo, &ag.bucketsAcc,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(571849)
			return appendTo, err
		} else {
			__antithesis_instrumentation__.Notify(571850)
		}
	}
	__antithesis_instrumentation__.Notify(571847)
	return appendTo, nil
}

func (ag *hashAggregator) accumulateRow(row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(571851)
	if err := ag.cancelChecker.Check(); err != nil {
		__antithesis_instrumentation__.Notify(571855)
		return err
	} else {
		__antithesis_instrumentation__.Notify(571856)
	}
	__antithesis_instrumentation__.Notify(571852)

	encoded, err := ag.encode(ag.scratch, row)
	if err != nil {
		__antithesis_instrumentation__.Notify(571857)
		return err
	} else {
		__antithesis_instrumentation__.Notify(571858)
	}
	__antithesis_instrumentation__.Notify(571853)
	ag.scratch = encoded[:0]

	bucket, ok := ag.buckets[string(encoded)]
	if !ok {
		__antithesis_instrumentation__.Notify(571859)
		s, err := ag.arena.AllocBytes(ag.Ctx, encoded)
		if err != nil {
			__antithesis_instrumentation__.Notify(571862)
			return err
		} else {
			__antithesis_instrumentation__.Notify(571863)
		}
		__antithesis_instrumentation__.Notify(571860)
		bucket, err = ag.createAggregateFuncs()
		if err != nil {
			__antithesis_instrumentation__.Notify(571864)
			return err
		} else {
			__antithesis_instrumentation__.Notify(571865)
		}
		__antithesis_instrumentation__.Notify(571861)
		ag.buckets[s] = bucket
		if len(ag.buckets) == ag.bucketsLenGrowThreshold {
			__antithesis_instrumentation__.Notify(571866)
			toAccountFor := ag.bucketsLenGrowThreshold - ag.alreadyAccountedFor
			if err := ag.bucketsAcc.Grow(ag.Ctx, int64(toAccountFor)*memsize.MapEntryOverhead); err != nil {
				__antithesis_instrumentation__.Notify(571868)
				return err
			} else {
				__antithesis_instrumentation__.Notify(571869)
			}
			__antithesis_instrumentation__.Notify(571867)
			ag.alreadyAccountedFor = ag.bucketsLenGrowThreshold
			ag.bucketsLenGrowThreshold *= 2
		} else {
			__antithesis_instrumentation__.Notify(571870)
		}
	} else {
		__antithesis_instrumentation__.Notify(571871)
	}
	__antithesis_instrumentation__.Notify(571854)

	return ag.accumulateRowIntoBucket(row, encoded, bucket)
}

func (ag *orderedAggregator) accumulateRow(row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(571872)
	if err := ag.cancelChecker.Check(); err != nil {
		__antithesis_instrumentation__.Notify(571875)
		return err
	} else {
		__antithesis_instrumentation__.Notify(571876)
	}
	__antithesis_instrumentation__.Notify(571873)

	if ag.bucket == nil {
		__antithesis_instrumentation__.Notify(571877)
		var err error
		ag.bucket, err = ag.createAggregateFuncs()
		if err != nil {
			__antithesis_instrumentation__.Notify(571878)
			return err
		} else {
			__antithesis_instrumentation__.Notify(571879)
		}
	} else {
		__antithesis_instrumentation__.Notify(571880)
	}
	__antithesis_instrumentation__.Notify(571874)

	return ag.accumulateRowIntoBucket(row, nil, ag.bucket)
}

type aggregateFuncHolder struct {
	create func(*tree.EvalContext, tree.Datums) tree.AggregateFunc

	arguments tree.Datums

	seen  map[string]struct{}
	arena *stringarena.Arena
}

const (
	sizeOfAggregateFuncs = int64(unsafe.Sizeof(aggregateFuncs{}))
	sizeOfAggregateFunc  = int64(unsafe.Sizeof(tree.AggregateFunc(nil)))
)

func (ag *aggregatorBase) newAggregateFuncHolder(
	create func(*tree.EvalContext, tree.Datums) tree.AggregateFunc, arguments tree.Datums,
) *aggregateFuncHolder {
	__antithesis_instrumentation__.Notify(571881)
	return &aggregateFuncHolder{
		create:    create,
		arena:     &ag.arena,
		arguments: arguments,
	}
}

func (a *aggregateFuncHolder) isDistinct(
	ctx context.Context,
	alloc *tree.DatumAlloc,
	prefix []byte,
	firstArg tree.Datum,
	otherArgs tree.Datums,
) (bool, error) {
	__antithesis_instrumentation__.Notify(571882)

	ed := rowenc.EncDatum{Datum: firstArg}

	encoded, err := ed.Fingerprint(ctx, firstArg.ResolvedType(), alloc, prefix, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(571887)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(571888)
	}
	__antithesis_instrumentation__.Notify(571883)
	for _, arg := range otherArgs {
		__antithesis_instrumentation__.Notify(571889)

		ed.Datum = arg
		encoded, err = ed.Fingerprint(ctx, arg.ResolvedType(), alloc, encoded, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(571890)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(571891)
		}
	}
	__antithesis_instrumentation__.Notify(571884)

	if _, ok := a.seen[string(encoded)]; ok {
		__antithesis_instrumentation__.Notify(571892)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(571893)
	}
	__antithesis_instrumentation__.Notify(571885)
	s, err := a.arena.AllocBytes(ctx, encoded)
	if err != nil {
		__antithesis_instrumentation__.Notify(571894)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(571895)
	}
	__antithesis_instrumentation__.Notify(571886)
	a.seen[s] = struct{}{}
	return true, nil
}

func (ag *aggregatorBase) createAggregateFuncs() (aggregateFuncs, error) {
	__antithesis_instrumentation__.Notify(571896)
	if err := ag.bucketsAcc.Grow(ag.Ctx, sizeOfAggregateFuncs+sizeOfAggregateFunc*int64(len(ag.funcs))); err != nil {
		__antithesis_instrumentation__.Notify(571899)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(571900)
	}
	__antithesis_instrumentation__.Notify(571897)
	bucket := make(aggregateFuncs, len(ag.funcs))
	for i, f := range ag.funcs {
		__antithesis_instrumentation__.Notify(571901)
		agg := f.create(ag.EvalCtx, f.arguments)
		if err := ag.bucketsAcc.Grow(ag.Ctx, agg.Size()); err != nil {
			__antithesis_instrumentation__.Notify(571903)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571904)
		}
		__antithesis_instrumentation__.Notify(571902)
		bucket[i] = agg
	}
	__antithesis_instrumentation__.Notify(571898)
	return bucket, nil
}
