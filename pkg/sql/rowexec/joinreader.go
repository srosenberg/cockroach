package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"sort"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/kvstreamer"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/scrub"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

type joinReaderState int

const (
	jrStateUnknown joinReaderState = iota

	jrReadingInput

	jrPerformingLookup

	jrEmittingRows

	jrReadyToDrain
)

type joinReaderType int

const (
	lookupJoinReaderType joinReaderType = iota

	indexJoinReaderType
)

type joinReader struct {
	joinerBase
	strategy joinReaderStrategy

	runningState joinReaderState

	memAcc mon.BoundAccount

	accountedFor struct {
		scratchInputRows int64
		groupingState    int64
	}

	limitedMemMonitor *mon.BytesMonitor
	diskMonitor       *mon.BytesMonitor

	fetchSpec      descpb.IndexFetchSpec
	splitFamilyIDs []descpb.FamilyID

	maintainOrdering bool

	fetcher            rowFetcher
	alloc              tree.DatumAlloc
	rowAlloc           rowenc.EncDatumRowAlloc
	shouldLimitBatches bool
	readerType         joinReaderType

	keyLocking     descpb.ScanLockingStrength
	lockWaitPolicy lock.WaitPolicy

	usesStreamer bool
	streamerInfo struct {
		*kvstreamer.Streamer
		unlimitedMemMonitor *mon.BytesMonitor
		budgetAcc           mon.BoundAccount
		budgetLimit         int64
		maxKeysPerRow       int
		diskMonitor         *mon.BytesMonitor
	}

	input execinfra.RowSource

	lookupCols       []uint32
	lookupExpr       execinfrapb.ExprHelper
	remoteLookupExpr execinfrapb.ExprHelper

	batchSizeBytes    int64
	curBatchSizeBytes int64

	pendingRow rowenc.EncDatumRow

	rowsRead int64

	curBatchRowsRead int64

	curBatchInputRowCount int64

	scratchInputRows rowenc.EncDatumRows

	resetScratchWhenReadingInput bool

	groupingState *inputBatchGroupingState

	lastBatchState struct {
		lastInputRow       rowenc.EncDatumRow
		lastGroupMatched   bool
		lastGroupContinued bool
	}

	outputGroupContinuationForLeftRow bool

	lookupBatchBytesLimit rowinfra.BytesLimit

	limitHintHelper execinfra.LimitHintHelper

	scanStats execinfra.ScanStats
}

var _ execinfra.Processor = &joinReader{}
var _ execinfra.RowSource = &joinReader{}
var _ execinfra.OpNode = &joinReader{}

const joinReaderProcName = "join reader"

var ParallelizeMultiKeyLookupJoinsEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.distsql.parallelize_multi_key_lookup_joins.enabled",
	"determines whether KV batches are executed in parallel for lookup joins in all cases. "+
		"Enabling this will increase the speed of lookup joins when each input row might get "+
		"multiple looked up rows at the cost of increased memory usage.",
	false,
)

func newJoinReader(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.JoinReaderSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
	readerType joinReaderType,
) (execinfra.RowSourcedProcessor, error) {
	__antithesis_instrumentation__.Notify(573157)
	if spec.OutputGroupContinuationForLeftRow && func() bool {
		__antithesis_instrumentation__.Notify(573174)
		return !spec.MaintainOrdering == true
	}() == true {
		__antithesis_instrumentation__.Notify(573175)
		return nil, errors.AssertionFailedf(
			"lookup join must maintain ordering since it is first join in paired-joins")
	} else {
		__antithesis_instrumentation__.Notify(573176)
	}
	__antithesis_instrumentation__.Notify(573158)
	switch readerType {
	case lookupJoinReaderType:
		__antithesis_instrumentation__.Notify(573177)
		switch spec.Type {
		case descpb.InnerJoin, descpb.LeftOuterJoin, descpb.LeftSemiJoin, descpb.LeftAntiJoin:
			__antithesis_instrumentation__.Notify(573183)
		default:
			__antithesis_instrumentation__.Notify(573184)
			return nil, errors.AssertionFailedf("only inner and left {outer, semi, anti} lookup joins are supported, %s requested", spec.Type)
		}
	case indexJoinReaderType:
		__antithesis_instrumentation__.Notify(573178)
		if spec.Type != descpb.InnerJoin {
			__antithesis_instrumentation__.Notify(573185)
			return nil, errors.AssertionFailedf("only inner index joins are supported, %s requested", spec.Type)
		} else {
			__antithesis_instrumentation__.Notify(573186)
		}
		__antithesis_instrumentation__.Notify(573179)
		if !spec.LookupExpr.Empty() {
			__antithesis_instrumentation__.Notify(573187)
			return nil, errors.AssertionFailedf("non-empty lookup expressions are not supported for index joins")
		} else {
			__antithesis_instrumentation__.Notify(573188)
		}
		__antithesis_instrumentation__.Notify(573180)
		if !spec.RemoteLookupExpr.Empty() {
			__antithesis_instrumentation__.Notify(573189)
			return nil, errors.AssertionFailedf("non-empty remote lookup expressions are not supported for index joins")
		} else {
			__antithesis_instrumentation__.Notify(573190)
		}
		__antithesis_instrumentation__.Notify(573181)
		if !spec.OnExpr.Empty() {
			__antithesis_instrumentation__.Notify(573191)
			return nil, errors.AssertionFailedf("non-empty ON expressions are not supported for index joins")
		} else {
			__antithesis_instrumentation__.Notify(573192)
		}
	default:
		__antithesis_instrumentation__.Notify(573182)
	}
	__antithesis_instrumentation__.Notify(573159)

	var lookupCols []uint32
	switch readerType {
	case indexJoinReaderType:
		__antithesis_instrumentation__.Notify(573193)
		lookupCols = make([]uint32, len(spec.FetchSpec.KeyColumns()))
		for i := range lookupCols {
			__antithesis_instrumentation__.Notify(573196)
			lookupCols[i] = uint32(i)
		}
	case lookupJoinReaderType:
		__antithesis_instrumentation__.Notify(573194)
		lookupCols = spec.LookupColumns
	default:
		__antithesis_instrumentation__.Notify(573195)
		return nil, errors.Errorf("unsupported joinReaderType")
	}
	__antithesis_instrumentation__.Notify(573160)

	shouldLimitBatches := !spec.LookupColumnsAreKey && func() bool {
		__antithesis_instrumentation__.Notify(573197)
		return readerType == lookupJoinReaderType == true
	}() == true
	if flowCtx.EvalCtx.SessionData().ParallelizeMultiKeyLookupJoinsEnabled {
		__antithesis_instrumentation__.Notify(573198)
		shouldLimitBatches = false
	} else {
		__antithesis_instrumentation__.Notify(573199)
	}
	__antithesis_instrumentation__.Notify(573161)
	useStreamer := flowCtx.Txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(573200)
		return flowCtx.Txn.Type() == kv.LeafTxn == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(573201)
		return readerType == indexJoinReaderType == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(573202)
		return row.CanUseStreamer(flowCtx.EvalCtx.Ctx(), flowCtx.EvalCtx.Settings) == true
	}() == true

	jr := &joinReader{
		fetchSpec:                         spec.FetchSpec,
		splitFamilyIDs:                    spec.SplitFamilyIDs,
		maintainOrdering:                  spec.MaintainOrdering,
		input:                             input,
		lookupCols:                        lookupCols,
		outputGroupContinuationForLeftRow: spec.OutputGroupContinuationForLeftRow,
		shouldLimitBatches:                shouldLimitBatches,
		readerType:                        readerType,
		keyLocking:                        spec.LockingStrength,
		lockWaitPolicy:                    row.GetWaitPolicy(spec.LockingWaitPolicy),
		usesStreamer:                      useStreamer,
		lookupBatchBytesLimit:             rowinfra.BytesLimit(spec.LookupBatchBytesLimit),
		limitHintHelper:                   execinfra.MakeLimitHintHelper(spec.LimitHint, post),
	}
	if readerType != indexJoinReaderType {
		__antithesis_instrumentation__.Notify(573203)
		jr.groupingState = &inputBatchGroupingState{doGrouping: spec.LeftJoinWithPairedJoiner}
	} else {
		__antithesis_instrumentation__.Notify(573204)
	}
	__antithesis_instrumentation__.Notify(573162)

	resolver := flowCtx.NewTypeResolver(flowCtx.Txn)
	for i := range spec.FetchSpec.KeyAndSuffixColumns {
		__antithesis_instrumentation__.Notify(573205)
		if err := typedesc.EnsureTypeIsHydrated(
			flowCtx.EvalCtx.Ctx(), spec.FetchSpec.KeyAndSuffixColumns[i].Type, &resolver,
		); err != nil {
			__antithesis_instrumentation__.Notify(573206)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(573207)
		}
	}
	__antithesis_instrumentation__.Notify(573163)

	var leftTypes []*types.T
	switch readerType {
	case indexJoinReaderType:
		__antithesis_instrumentation__.Notify(573208)

	case lookupJoinReaderType:
		__antithesis_instrumentation__.Notify(573209)
		leftTypes = input.OutputTypes()
	default:
		__antithesis_instrumentation__.Notify(573210)
		return nil, errors.Errorf("unsupported joinReaderType")
	}
	__antithesis_instrumentation__.Notify(573164)
	rightTypes := make([]*types.T, len(spec.FetchSpec.FetchedColumns))
	for i := range rightTypes {
		__antithesis_instrumentation__.Notify(573211)
		rightTypes[i] = spec.FetchSpec.FetchedColumns[i].Type
	}
	__antithesis_instrumentation__.Notify(573165)

	if err := jr.joinerBase.init(
		jr,
		flowCtx,
		processorID,
		leftTypes,
		rightTypes,
		spec.Type,
		spec.OnExpr,
		spec.OutputGroupContinuationForLeftRow,
		post,
		output,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{jr.input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(573212)

				trailingMeta := jr.generateMeta()
				jr.close()
				return trailingMeta
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(573213)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(573214)
	}
	__antithesis_instrumentation__.Notify(573166)

	var fetcher row.Fetcher
	if err := fetcher.Init(
		flowCtx.EvalCtx.Context,
		false,
		spec.LockingStrength,
		spec.LockingWaitPolicy,
		flowCtx.EvalCtx.SessionData().LockTimeout,
		&jr.alloc,
		flowCtx.EvalCtx.Mon,
		&spec.FetchSpec,
	); err != nil {
		__antithesis_instrumentation__.Notify(573215)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(573216)
	}
	__antithesis_instrumentation__.Notify(573167)

	if execinfra.ShouldCollectStats(flowCtx.EvalCtx.Ctx(), flowCtx) {
		__antithesis_instrumentation__.Notify(573217)
		jr.input = newInputStatCollector(jr.input)
		jr.fetcher = newRowFetcherStatCollector(&fetcher)
		jr.ExecStatsForTrace = jr.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(573218)
		jr.fetcher = &fetcher
	}
	__antithesis_instrumentation__.Notify(573168)

	if !spec.LookupExpr.Empty() {
		__antithesis_instrumentation__.Notify(573219)
		lookupExprTypes := make([]*types.T, 0, len(leftTypes)+len(rightTypes))
		lookupExprTypes = append(lookupExprTypes, leftTypes...)
		lookupExprTypes = append(lookupExprTypes, rightTypes...)

		semaCtx := flowCtx.NewSemaContext(flowCtx.EvalCtx.Txn)
		if err := jr.lookupExpr.Init(spec.LookupExpr, lookupExprTypes, semaCtx, jr.EvalCtx); err != nil {
			__antithesis_instrumentation__.Notify(573221)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(573222)
		}
		__antithesis_instrumentation__.Notify(573220)
		if !spec.RemoteLookupExpr.Empty() {
			__antithesis_instrumentation__.Notify(573223)
			if err := jr.remoteLookupExpr.Init(
				spec.RemoteLookupExpr, lookupExprTypes, semaCtx, jr.EvalCtx,
			); err != nil {
				__antithesis_instrumentation__.Notify(573224)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(573225)
			}
		} else {
			__antithesis_instrumentation__.Notify(573226)
		}
	} else {
		__antithesis_instrumentation__.Notify(573227)
	}
	__antithesis_instrumentation__.Notify(573169)

	minMemoryLimit := int64(8 << 20)

	if jr.usesStreamer {
		__antithesis_instrumentation__.Notify(573228)
		minMemoryLimit = 100 << 10
	} else {
		__antithesis_instrumentation__.Notify(573229)
	}
	__antithesis_instrumentation__.Notify(573170)
	memoryLimit := execinfra.GetWorkMemLimit(flowCtx)
	if memoryLimit < minMemoryLimit {
		__antithesis_instrumentation__.Notify(573230)
		memoryLimit = minMemoryLimit
	} else {
		__antithesis_instrumentation__.Notify(573231)
	}
	__antithesis_instrumentation__.Notify(573171)

	jr.MemMonitor = mon.NewMonitorInheritWithLimit(
		"joinreader-mem", memoryLimit, flowCtx.EvalCtx.Mon,
	)
	jr.MemMonitor.Start(flowCtx.EvalCtx.Ctx(), flowCtx.EvalCtx.Mon, mon.BoundAccount{})
	jr.memAcc = jr.MemMonitor.MakeBoundAccount()

	if err := jr.initJoinReaderStrategy(flowCtx, rightTypes, readerType); err != nil {
		__antithesis_instrumentation__.Notify(573232)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(573233)
	}
	__antithesis_instrumentation__.Notify(573172)
	jr.batchSizeBytes = jr.strategy.getLookupRowsBatchSizeHint(flowCtx.EvalCtx.SessionData())

	if jr.usesStreamer {
		__antithesis_instrumentation__.Notify(573234)

		if jr.batchSizeBytes > memoryLimit/4 {
			__antithesis_instrumentation__.Notify(573236)
			jr.batchSizeBytes = memoryLimit / 4
		} else {
			__antithesis_instrumentation__.Notify(573237)
		}
		__antithesis_instrumentation__.Notify(573235)

		jr.streamerInfo.budgetLimit = memoryLimit - jr.batchSizeBytes

		jr.streamerInfo.unlimitedMemMonitor = mon.NewMonitorInheritWithLimit(
			"joinreader-streamer-unlimited", math.MaxInt64, flowCtx.EvalCtx.Mon,
		)
		jr.streamerInfo.unlimitedMemMonitor.Start(flowCtx.EvalCtx.Ctx(), flowCtx.EvalCtx.Mon, mon.BoundAccount{})
		jr.streamerInfo.budgetAcc = jr.streamerInfo.unlimitedMemMonitor.MakeBoundAccount()
		jr.streamerInfo.maxKeysPerRow = int(jr.fetchSpec.MaxKeysPerRow)
	} else {
		__antithesis_instrumentation__.Notify(573238)

		if jr.batchSizeBytes > memoryLimit/2 {
			__antithesis_instrumentation__.Notify(573239)
			jr.batchSizeBytes = memoryLimit / 2
		} else {
			__antithesis_instrumentation__.Notify(573240)
		}
	}
	__antithesis_instrumentation__.Notify(573173)

	return jr, nil
}

func (jr *joinReader) initJoinReaderStrategy(
	flowCtx *execinfra.FlowCtx, typs []*types.T, readerType joinReaderType,
) error {
	__antithesis_instrumentation__.Notify(573241)
	strategyMemAcc := jr.MemMonitor.MakeBoundAccount()
	spanGeneratorMemAcc := jr.MemMonitor.MakeBoundAccount()
	var generator joinReaderSpanGenerator
	if jr.lookupExpr.Expr == nil {
		__antithesis_instrumentation__.Notify(573246)
		var keyToInputRowIndices map[string][]int

		if readerType != indexJoinReaderType {
			__antithesis_instrumentation__.Notify(573249)
			keyToInputRowIndices = make(map[string][]int)
		} else {
			__antithesis_instrumentation__.Notify(573250)
		}
		__antithesis_instrumentation__.Notify(573247)
		defGen := &defaultSpanGenerator{}
		if err := defGen.init(
			flowCtx.EvalCtx,
			flowCtx.Codec(),
			&jr.fetchSpec,
			jr.splitFamilyIDs,
			keyToInputRowIndices,
			jr.lookupCols,
			&spanGeneratorMemAcc,
		); err != nil {
			__antithesis_instrumentation__.Notify(573251)
			return err
		} else {
			__antithesis_instrumentation__.Notify(573252)
		}
		__antithesis_instrumentation__.Notify(573248)
		generator = defGen
	} else {
		__antithesis_instrumentation__.Notify(573253)

		var fetchedOrdToIndexKeyOrd util.FastIntMap
		fullColumns := jr.fetchSpec.KeyFullColumns()
		for keyOrdinal := range fullColumns {
			__antithesis_instrumentation__.Notify(573255)
			keyColID := fullColumns[keyOrdinal].ColumnID
			for fetchedOrdinal := range jr.fetchSpec.FetchedColumns {
				__antithesis_instrumentation__.Notify(573256)
				if jr.fetchSpec.FetchedColumns[fetchedOrdinal].ColumnID == keyColID {
					__antithesis_instrumentation__.Notify(573257)
					fetchedOrdToIndexKeyOrd.Set(fetchedOrdinal, keyOrdinal)
					break
				} else {
					__antithesis_instrumentation__.Notify(573258)
				}
			}
		}
		__antithesis_instrumentation__.Notify(573254)

		if jr.remoteLookupExpr.Expr == nil {
			__antithesis_instrumentation__.Notify(573259)
			multiSpanGen := &multiSpanGenerator{}
			if err := multiSpanGen.init(
				flowCtx.EvalCtx,
				flowCtx.Codec(),
				&jr.fetchSpec,
				jr.splitFamilyIDs,
				len(jr.input.OutputTypes()),
				&jr.lookupExpr,
				fetchedOrdToIndexKeyOrd,
				&spanGeneratorMemAcc,
			); err != nil {
				__antithesis_instrumentation__.Notify(573261)
				return err
			} else {
				__antithesis_instrumentation__.Notify(573262)
			}
			__antithesis_instrumentation__.Notify(573260)
			generator = multiSpanGen
		} else {
			__antithesis_instrumentation__.Notify(573263)
			localityOptSpanGen := &localityOptimizedSpanGenerator{}
			remoteSpanGenMemAcc := jr.MemMonitor.MakeBoundAccount()
			if err := localityOptSpanGen.init(
				flowCtx.EvalCtx,
				flowCtx.Codec(),
				&jr.fetchSpec,
				jr.splitFamilyIDs,
				len(jr.input.OutputTypes()),
				&jr.lookupExpr,
				&jr.remoteLookupExpr,
				fetchedOrdToIndexKeyOrd,
				&spanGeneratorMemAcc,
				&remoteSpanGenMemAcc,
			); err != nil {
				__antithesis_instrumentation__.Notify(573265)
				return err
			} else {
				__antithesis_instrumentation__.Notify(573266)
			}
			__antithesis_instrumentation__.Notify(573264)
			generator = localityOptSpanGen
		}
	}
	__antithesis_instrumentation__.Notify(573242)

	if readerType == indexJoinReaderType {
		__antithesis_instrumentation__.Notify(573267)
		jr.strategy = &joinReaderIndexJoinStrategy{
			joinerBase:              &jr.joinerBase,
			joinReaderSpanGenerator: generator,
			memAcc:                  &strategyMemAcc,
		}
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573268)
	}
	__antithesis_instrumentation__.Notify(573243)

	if !jr.maintainOrdering {
		__antithesis_instrumentation__.Notify(573269)
		jr.strategy = &joinReaderNoOrderingStrategy{
			joinerBase:              &jr.joinerBase,
			joinReaderSpanGenerator: generator,
			isPartialJoin: jr.joinType == descpb.LeftSemiJoin || func() bool {
				__antithesis_instrumentation__.Notify(573270)
				return jr.joinType == descpb.LeftAntiJoin == true
			}() == true,
			groupingState: jr.groupingState,
			memAcc:        &strategyMemAcc,
		}
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573271)
	}
	__antithesis_instrumentation__.Notify(573244)

	ctx := flowCtx.EvalCtx.Ctx()

	limit := execinfra.GetWorkMemLimit(flowCtx)

	jr.limitedMemMonitor = execinfra.NewLimitedMonitor(ctx, jr.MemMonitor, flowCtx, "joinreader-limited")
	jr.diskMonitor = execinfra.NewMonitor(ctx, flowCtx.DiskMonitor, "joinreader-disk")
	drc := rowcontainer.NewDiskBackedNumberedRowContainer(
		false,
		typs,
		jr.EvalCtx,
		jr.FlowCtx.Cfg.TempStorage,
		jr.limitedMemMonitor,
		jr.diskMonitor,
	)
	if limit < mon.DefaultPoolAllocationSize {
		__antithesis_instrumentation__.Notify(573272)

		drc.DisableCache = true
	} else {
		__antithesis_instrumentation__.Notify(573273)
	}
	__antithesis_instrumentation__.Notify(573245)
	jr.strategy = &joinReaderOrderingStrategy{
		joinerBase:              &jr.joinerBase,
		joinReaderSpanGenerator: generator,
		isPartialJoin: jr.joinType == descpb.LeftSemiJoin || func() bool {
			__antithesis_instrumentation__.Notify(573274)
			return jr.joinType == descpb.LeftAntiJoin == true
		}() == true,
		lookedUpRows:                      drc,
		groupingState:                     jr.groupingState,
		outputGroupContinuationForLeftRow: jr.outputGroupContinuationForLeftRow,
		memAcc:                            &strategyMemAcc,
	}
	return nil
}

func (jr *joinReader) SetBatchSizeBytes(batchSize int64) {
	__antithesis_instrumentation__.Notify(573275)
	jr.batchSizeBytes = batchSize
}

func (jr *joinReader) Spilled() bool {
	__antithesis_instrumentation__.Notify(573276)
	return jr.strategy.spilled()
}

func (jr *joinReader) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(573277)

	for jr.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(573279)
		var row rowenc.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch jr.runningState {
		case jrReadingInput:
			__antithesis_instrumentation__.Notify(573283)
			jr.runningState, row, meta = jr.readInput()
		case jrPerformingLookup:
			__antithesis_instrumentation__.Notify(573284)
			jr.runningState, meta = jr.performLookup()
		case jrEmittingRows:
			__antithesis_instrumentation__.Notify(573285)
			jr.runningState, row, meta = jr.emitRow()
		case jrReadyToDrain:
			__antithesis_instrumentation__.Notify(573286)
			jr.MoveToDraining(nil)
			meta = jr.DrainHelper()
			jr.runningState = jrStateUnknown
		default:
			__antithesis_instrumentation__.Notify(573287)
			log.Fatalf(jr.Ctx, "unsupported state: %d", jr.runningState)
		}
		__antithesis_instrumentation__.Notify(573280)
		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(573288)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(573289)
			continue
		} else {
			__antithesis_instrumentation__.Notify(573290)
		}
		__antithesis_instrumentation__.Notify(573281)
		if meta != nil {
			__antithesis_instrumentation__.Notify(573291)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(573292)
		}
		__antithesis_instrumentation__.Notify(573282)
		if outRow := jr.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(573293)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(573294)
		}
	}
	__antithesis_instrumentation__.Notify(573278)
	return nil, jr.DrainHelper()
}

func addWorkmemHint(err error) error {
	__antithesis_instrumentation__.Notify(573295)
	if err == nil {
		__antithesis_instrumentation__.Notify(573297)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573298)
	}
	__antithesis_instrumentation__.Notify(573296)
	return errors.WithHint(
		err, "consider increasing sql.distsql.temp_storage.workmem cluster"+
			" setting or distsql_workmem session variable",
	)
}

func (jr *joinReader) readInput() (
	joinReaderState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(573299)
	if jr.groupingState != nil {
		__antithesis_instrumentation__.Notify(573313)

		if jr.groupingState.initialized {
			__antithesis_instrumentation__.Notify(573314)

			jr.lastBatchState.lastGroupMatched = jr.groupingState.lastGroupMatched()
			jr.groupingState.reset()
			jr.lastBatchState.lastGroupContinued = false
		} else {
			__antithesis_instrumentation__.Notify(573315)
		}

	} else {
		__antithesis_instrumentation__.Notify(573316)
	}
	__antithesis_instrumentation__.Notify(573300)

	if jr.resetScratchWhenReadingInput {
		__antithesis_instrumentation__.Notify(573317)

		for i := range jr.scratchInputRows {
			__antithesis_instrumentation__.Notify(573320)
			jr.scratchInputRows[i] = nil
		}
		__antithesis_instrumentation__.Notify(573318)

		newSz := jr.accountedFor.scratchInputRows + jr.accountedFor.groupingState
		if err := jr.memAcc.ResizeTo(jr.Ctx, newSz); err != nil {
			__antithesis_instrumentation__.Notify(573321)
			jr.MoveToDraining(addWorkmemHint(err))
			return jrStateUnknown, nil, jr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573322)
		}
		__antithesis_instrumentation__.Notify(573319)
		jr.scratchInputRows = jr.scratchInputRows[:0]
		jr.resetScratchWhenReadingInput = false
	} else {
		__antithesis_instrumentation__.Notify(573323)
	}
	__antithesis_instrumentation__.Notify(573301)

	for {
		__antithesis_instrumentation__.Notify(573324)
		var encDatumRow rowenc.EncDatumRow
		var rowSize int64
		if jr.pendingRow == nil {
			__antithesis_instrumentation__.Notify(573328)

			var meta *execinfrapb.ProducerMetadata
			encDatumRow, meta = jr.input.Next()
			if meta != nil {
				__antithesis_instrumentation__.Notify(573331)
				if meta.Err != nil {
					__antithesis_instrumentation__.Notify(573334)
					jr.MoveToDraining(nil)
					return jrStateUnknown, nil, meta
				} else {
					__antithesis_instrumentation__.Notify(573335)
				}
				__antithesis_instrumentation__.Notify(573332)

				if err := jr.performMemoryAccounting(); err != nil {
					__antithesis_instrumentation__.Notify(573336)
					jr.MoveToDraining(err)
					return jrStateUnknown, nil, meta
				} else {
					__antithesis_instrumentation__.Notify(573337)
				}
				__antithesis_instrumentation__.Notify(573333)

				return jrReadingInput, nil, meta
			} else {
				__antithesis_instrumentation__.Notify(573338)
			}
			__antithesis_instrumentation__.Notify(573329)
			if encDatumRow == nil {
				__antithesis_instrumentation__.Notify(573339)
				break
			} else {
				__antithesis_instrumentation__.Notify(573340)
			}
			__antithesis_instrumentation__.Notify(573330)
			rowSize = int64(encDatumRow.Size())
			if jr.curBatchSizeBytes > 0 && func() bool {
				__antithesis_instrumentation__.Notify(573341)
				return jr.curBatchSizeBytes+rowSize > jr.batchSizeBytes == true
			}() == true {
				__antithesis_instrumentation__.Notify(573342)

				jr.pendingRow = encDatumRow
				break
			} else {
				__antithesis_instrumentation__.Notify(573343)
			}
		} else {
			__antithesis_instrumentation__.Notify(573344)
			encDatumRow = jr.pendingRow
			jr.pendingRow = nil
			rowSize = int64(encDatumRow.Size())
		}
		__antithesis_instrumentation__.Notify(573325)
		jr.curBatchSizeBytes += rowSize
		if jr.groupingState != nil {
			__antithesis_instrumentation__.Notify(573345)

			if err := jr.processContinuationValForRow(encDatumRow); err != nil {
				__antithesis_instrumentation__.Notify(573346)
				jr.MoveToDraining(err)
				return jrStateUnknown, nil, jr.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(573347)
			}
		} else {
			__antithesis_instrumentation__.Notify(573348)
		}
		__antithesis_instrumentation__.Notify(573326)

		if err := jr.memAcc.Grow(jr.Ctx, rowSize-int64(rowenc.EncDatumRowOverhead)); err != nil {
			__antithesis_instrumentation__.Notify(573349)
			jr.MoveToDraining(addWorkmemHint(err))
			return jrStateUnknown, nil, jr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573350)
		}
		__antithesis_instrumentation__.Notify(573327)
		jr.scratchInputRows = append(jr.scratchInputRows, jr.rowAlloc.CopyRow(encDatumRow))

		if l := jr.limitHintHelper.LimitHint(); l != 0 && func() bool {
			__antithesis_instrumentation__.Notify(573351)
			return l == int64(len(jr.scratchInputRows)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(573352)
			break
		} else {
			__antithesis_instrumentation__.Notify(573353)
		}
	}
	__antithesis_instrumentation__.Notify(573302)

	if err := jr.performMemoryAccounting(); err != nil {
		__antithesis_instrumentation__.Notify(573354)
		jr.MoveToDraining(err)
		return jrStateUnknown, nil, jr.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(573355)
	}
	__antithesis_instrumentation__.Notify(573303)

	var outRow rowenc.EncDatumRow

	if jr.groupingState != nil {
		__antithesis_instrumentation__.Notify(573356)

		outRow = jr.allContinuationValsProcessed()
	} else {
		__antithesis_instrumentation__.Notify(573357)
	}
	__antithesis_instrumentation__.Notify(573304)

	if len(jr.scratchInputRows) == 0 {
		__antithesis_instrumentation__.Notify(573358)
		log.VEventf(jr.Ctx, 1, "no more input rows")
		if outRow != nil {
			__antithesis_instrumentation__.Notify(573360)
			return jrReadyToDrain, outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(573361)
		}
		__antithesis_instrumentation__.Notify(573359)

		jr.MoveToDraining(nil)
		return jrStateUnknown, nil, jr.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(573362)
	}
	__antithesis_instrumentation__.Notify(573305)
	log.VEventf(jr.Ctx, 1, "read %d input rows", len(jr.scratchInputRows))

	if jr.groupingState != nil && func() bool {
		__antithesis_instrumentation__.Notify(573363)
		return len(jr.scratchInputRows) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(573364)
		jr.updateGroupingStateForNonEmptyBatch()
	} else {
		__antithesis_instrumentation__.Notify(573365)
	}
	__antithesis_instrumentation__.Notify(573306)

	if err := jr.limitHintHelper.ReadSomeRows(int64(len(jr.scratchInputRows))); err != nil {
		__antithesis_instrumentation__.Notify(573366)
		jr.MoveToDraining(err)
		return jrStateUnknown, nil, jr.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(573367)
	}
	__antithesis_instrumentation__.Notify(573307)

	spans, err := jr.strategy.processLookupRows(jr.scratchInputRows)
	if err != nil {
		__antithesis_instrumentation__.Notify(573368)
		jr.MoveToDraining(err)
		return jrStateUnknown, nil, jr.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(573369)
	}
	__antithesis_instrumentation__.Notify(573308)
	jr.curBatchInputRowCount = int64(len(jr.scratchInputRows))
	jr.resetScratchWhenReadingInput = true
	jr.curBatchSizeBytes = 0
	jr.curBatchRowsRead = 0
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(573370)

		return jrEmittingRows, outRow, nil
	} else {
		__antithesis_instrumentation__.Notify(573371)
	}
	__antithesis_instrumentation__.Notify(573309)

	if jr.readerType == indexJoinReaderType && func() bool {
		__antithesis_instrumentation__.Notify(573372)
		return jr.maintainOrdering == true
	}() == true {
		__antithesis_instrumentation__.Notify(573373)

		if jr.shouldLimitBatches {
			__antithesis_instrumentation__.Notify(573374)
			err := errors.AssertionFailedf("index join configured with both maintainOrdering and " +
				"shouldLimitBatched; this shouldn't have happened as the implementation doesn't support it")
			jr.MoveToDraining(err)
			return jrStateUnknown, nil, jr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573375)
		}
	} else {
		__antithesis_instrumentation__.Notify(573376)
		sort.Sort(spans)
	}
	__antithesis_instrumentation__.Notify(573310)

	log.VEventf(jr.Ctx, 1, "scanning %d spans", len(spans))

	if jr.usesStreamer {
		__antithesis_instrumentation__.Notify(573377)
		var kvBatchFetcher *row.TxnKVStreamer
		kvBatchFetcher, err = row.NewTxnKVStreamer(jr.Ctx, jr.streamerInfo.Streamer, spans, jr.keyLocking)
		if err != nil {
			__antithesis_instrumentation__.Notify(573379)
			jr.MoveToDraining(err)
			return jrStateUnknown, nil, jr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573380)
		}
		__antithesis_instrumentation__.Notify(573378)
		err = jr.fetcher.StartScanFrom(jr.Ctx, kvBatchFetcher, jr.FlowCtx.TraceKV)
	} else {
		__antithesis_instrumentation__.Notify(573381)
		var bytesLimit rowinfra.BytesLimit
		if !jr.shouldLimitBatches {
			__antithesis_instrumentation__.Notify(573383)
			bytesLimit = rowinfra.NoBytesLimit
		} else {
			__antithesis_instrumentation__.Notify(573384)
			bytesLimit = jr.lookupBatchBytesLimit
			if jr.lookupBatchBytesLimit == 0 {
				__antithesis_instrumentation__.Notify(573385)
				bytesLimit = rowinfra.DefaultBatchBytesLimit
			} else {
				__antithesis_instrumentation__.Notify(573386)
			}
		}
		__antithesis_instrumentation__.Notify(573382)
		err = jr.fetcher.StartScan(
			jr.Ctx, jr.FlowCtx.Txn, spans, bytesLimit, rowinfra.NoRowLimit,
			jr.FlowCtx.TraceKV, jr.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
		)
	}
	__antithesis_instrumentation__.Notify(573311)
	if err != nil {
		__antithesis_instrumentation__.Notify(573387)
		jr.MoveToDraining(err)
		return jrStateUnknown, nil, jr.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(573388)
	}
	__antithesis_instrumentation__.Notify(573312)

	return jrPerformingLookup, outRow, nil
}

func (jr *joinReader) performLookup() (joinReaderState, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(573389)
	for {
		__antithesis_instrumentation__.Notify(573392)

		var key roachpb.Key

		if jr.readerType != indexJoinReaderType {
			__antithesis_instrumentation__.Notify(573396)
			var err error
			key, err = jr.fetcher.PartialKey(jr.strategy.getMaxLookupKeyCols())
			if err != nil {
				__antithesis_instrumentation__.Notify(573397)
				jr.MoveToDraining(err)
				return jrStateUnknown, jr.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(573398)
			}
		} else {
			__antithesis_instrumentation__.Notify(573399)
		}
		__antithesis_instrumentation__.Notify(573393)

		lookedUpRow, err := jr.fetcher.NextRow(jr.Ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(573400)
			jr.MoveToDraining(scrub.UnwrapScrubError(err))
			return jrStateUnknown, jr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573401)
		}
		__antithesis_instrumentation__.Notify(573394)
		if lookedUpRow == nil {
			__antithesis_instrumentation__.Notify(573402)

			break
		} else {
			__antithesis_instrumentation__.Notify(573403)
		}
		__antithesis_instrumentation__.Notify(573395)
		jr.rowsRead++
		jr.curBatchRowsRead++

		if nextState, err := jr.strategy.processLookedUpRow(jr.Ctx, lookedUpRow, key); err != nil {
			__antithesis_instrumentation__.Notify(573404)
			jr.MoveToDraining(err)
			return jrStateUnknown, jr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573405)
			if nextState != jrPerformingLookup {
				__antithesis_instrumentation__.Notify(573406)
				return nextState, nil
			} else {
				__antithesis_instrumentation__.Notify(573407)
			}
		}
	}
	__antithesis_instrumentation__.Notify(573390)

	if jr.remoteLookupExpr.Expr != nil && func() bool {
		__antithesis_instrumentation__.Notify(573408)
		return !jr.strategy.generatedRemoteSpans() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(573409)
		return jr.curBatchRowsRead != jr.curBatchInputRowCount == true
	}() == true {
		__antithesis_instrumentation__.Notify(573410)
		spans, err := jr.strategy.generateRemoteSpans()
		if err != nil {
			__antithesis_instrumentation__.Notify(573412)
			jr.MoveToDraining(err)
			return jrStateUnknown, jr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573413)
		}
		__antithesis_instrumentation__.Notify(573411)

		if len(spans) != 0 {
			__antithesis_instrumentation__.Notify(573414)

			sort.Sort(spans)

			log.VEventf(jr.Ctx, 1, "scanning %d remote spans", len(spans))
			bytesLimit := rowinfra.DefaultBatchBytesLimit
			if !jr.shouldLimitBatches {
				__antithesis_instrumentation__.Notify(573417)
				bytesLimit = rowinfra.NoBytesLimit
			} else {
				__antithesis_instrumentation__.Notify(573418)
			}
			__antithesis_instrumentation__.Notify(573415)
			if err := jr.fetcher.StartScan(
				jr.Ctx, jr.FlowCtx.Txn, spans, bytesLimit, rowinfra.NoRowLimit,
				jr.FlowCtx.TraceKV, jr.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
			); err != nil {
				__antithesis_instrumentation__.Notify(573419)
				jr.MoveToDraining(err)
				return jrStateUnknown, jr.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(573420)
			}
			__antithesis_instrumentation__.Notify(573416)
			return jrPerformingLookup, nil
		} else {
			__antithesis_instrumentation__.Notify(573421)
		}
	} else {
		__antithesis_instrumentation__.Notify(573422)
	}
	__antithesis_instrumentation__.Notify(573391)

	log.VEvent(jr.Ctx, 1, "done joining rows")
	jr.strategy.prepareToEmit(jr.Ctx)

	return jrEmittingRows, nil
}

func (jr *joinReader) emitRow() (
	joinReaderState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(573423)
	rowToEmit, nextState, err := jr.strategy.nextRowToEmit(jr.Ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(573425)
		jr.MoveToDraining(err)
		return jrStateUnknown, nil, jr.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(573426)
	}
	__antithesis_instrumentation__.Notify(573424)
	return nextState, rowToEmit, nil
}

func (jr *joinReader) performMemoryAccounting() error {
	__antithesis_instrumentation__.Notify(573427)
	oldSz := jr.accountedFor.scratchInputRows + jr.accountedFor.groupingState
	jr.accountedFor.scratchInputRows = int64(cap(jr.scratchInputRows)) * int64(rowenc.EncDatumRowOverhead)
	jr.accountedFor.groupingState = jr.groupingState.memUsage()
	newSz := jr.accountedFor.scratchInputRows + jr.accountedFor.groupingState
	return addWorkmemHint(jr.memAcc.Resize(jr.Ctx, oldSz, newSz))
}

func (jr *joinReader) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573428)
	ctx = jr.StartInternal(ctx, joinReaderProcName)
	jr.input.Start(ctx)
	if jr.usesStreamer {
		__antithesis_instrumentation__.Notify(573430)
		jr.streamerInfo.Streamer = kvstreamer.NewStreamer(
			jr.FlowCtx.Cfg.DistSender,
			jr.FlowCtx.Stopper(),
			jr.FlowCtx.Txn,
			jr.FlowCtx.EvalCtx.Settings,
			jr.lockWaitPolicy,
			jr.streamerInfo.budgetLimit,
			&jr.streamerInfo.budgetAcc,
		)
		mode := kvstreamer.OutOfOrder
		if jr.maintainOrdering {
			__antithesis_instrumentation__.Notify(573432)
			mode = kvstreamer.InOrder
			jr.streamerInfo.diskMonitor = execinfra.NewMonitor(
				ctx, jr.FlowCtx.DiskMonitor, "streamer-disk",
			)
		} else {
			__antithesis_instrumentation__.Notify(573433)
		}
		__antithesis_instrumentation__.Notify(573431)
		jr.streamerInfo.Streamer.Init(
			mode,
			kvstreamer.Hints{UniqueRequests: true},
			jr.streamerInfo.maxKeysPerRow,
			jr.FlowCtx.Cfg.TempStorage,
			jr.streamerInfo.diskMonitor,
		)
	} else {
		__antithesis_instrumentation__.Notify(573434)
	}
	__antithesis_instrumentation__.Notify(573429)
	jr.runningState = jrReadingInput
}

func (jr *joinReader) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(573435)

	jr.close()
}

func (jr *joinReader) close() {
	__antithesis_instrumentation__.Notify(573436)
	if jr.InternalClose() {
		__antithesis_instrumentation__.Notify(573437)
		if jr.fetcher != nil {
			__antithesis_instrumentation__.Notify(573442)
			jr.fetcher.Close(jr.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(573443)
		}
		__antithesis_instrumentation__.Notify(573438)
		if jr.usesStreamer {
			__antithesis_instrumentation__.Notify(573444)

			if jr.streamerInfo.Streamer != nil {
				__antithesis_instrumentation__.Notify(573446)
				jr.streamerInfo.Streamer.Close(jr.Ctx)
			} else {
				__antithesis_instrumentation__.Notify(573447)
			}
			__antithesis_instrumentation__.Notify(573445)
			jr.streamerInfo.budgetAcc.Close(jr.Ctx)
			jr.streamerInfo.unlimitedMemMonitor.Stop(jr.Ctx)
			if jr.streamerInfo.diskMonitor != nil {
				__antithesis_instrumentation__.Notify(573448)
				jr.streamerInfo.diskMonitor.Stop(jr.Ctx)
			} else {
				__antithesis_instrumentation__.Notify(573449)
			}
		} else {
			__antithesis_instrumentation__.Notify(573450)
		}
		__antithesis_instrumentation__.Notify(573439)
		jr.strategy.close(jr.Ctx)
		jr.memAcc.Close(jr.Ctx)
		if jr.limitedMemMonitor != nil {
			__antithesis_instrumentation__.Notify(573451)
			jr.limitedMemMonitor.Stop(jr.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(573452)
		}
		__antithesis_instrumentation__.Notify(573440)
		if jr.MemMonitor != nil {
			__antithesis_instrumentation__.Notify(573453)
			jr.MemMonitor.Stop(jr.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(573454)
		}
		__antithesis_instrumentation__.Notify(573441)
		if jr.diskMonitor != nil {
			__antithesis_instrumentation__.Notify(573455)
			jr.diskMonitor.Stop(jr.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(573456)
		}
	} else {
		__antithesis_instrumentation__.Notify(573457)
	}
}

func (jr *joinReader) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(573458)
	is, ok := getInputStats(jr.input)
	if !ok {
		__antithesis_instrumentation__.Notify(573463)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573464)
	}
	__antithesis_instrumentation__.Notify(573459)
	fis, ok := getFetcherInputStats(jr.fetcher)
	if !ok {
		__antithesis_instrumentation__.Notify(573465)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573466)
	}
	__antithesis_instrumentation__.Notify(573460)

	jr.scanStats = execinfra.GetScanStats(jr.Ctx)
	ret := &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		KV: execinfrapb.KVStats{
			BytesRead:      optional.MakeUint(uint64(jr.fetcher.GetBytesRead())),
			TuplesRead:     fis.NumTuples,
			KVTime:         fis.WaitTime,
			ContentionTime: optional.MakeTimeValue(execinfra.GetCumulativeContentionTime(jr.Ctx)),
		},
		Output: jr.OutputHelper.Stats(),
	}

	ret.Exec.MaxAllocatedMem.Add(jr.MemMonitor.MaximumBytes())
	if jr.diskMonitor != nil {
		__antithesis_instrumentation__.Notify(573467)
		ret.Exec.MaxAllocatedDisk.Add(jr.diskMonitor.MaximumBytes())
	} else {
		__antithesis_instrumentation__.Notify(573468)
	}
	__antithesis_instrumentation__.Notify(573461)
	if jr.usesStreamer {
		__antithesis_instrumentation__.Notify(573469)
		ret.Exec.MaxAllocatedMem.Add(jr.streamerInfo.unlimitedMemMonitor.MaximumBytes())
		if jr.streamerInfo.diskMonitor != nil {
			__antithesis_instrumentation__.Notify(573470)
			ret.Exec.MaxAllocatedDisk.Add(jr.streamerInfo.diskMonitor.MaximumBytes())
		} else {
			__antithesis_instrumentation__.Notify(573471)
		}
	} else {
		__antithesis_instrumentation__.Notify(573472)
	}
	__antithesis_instrumentation__.Notify(573462)
	execinfra.PopulateKVMVCCStats(&ret.KV, &jr.scanStats)
	return ret
}

func (jr *joinReader) generateMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(573473)
	trailingMeta := make([]execinfrapb.ProducerMetadata, 1, 2)
	meta := &trailingMeta[0]
	meta.Metrics = execinfrapb.GetMetricsMeta()
	meta.Metrics.RowsRead = jr.rowsRead
	meta.Metrics.BytesRead = jr.fetcher.GetBytesRead()
	if tfs := execinfra.GetLeafTxnFinalState(jr.Ctx, jr.FlowCtx.Txn); tfs != nil {
		__antithesis_instrumentation__.Notify(573475)
		trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{LeafTxnFinalState: tfs})
	} else {
		__antithesis_instrumentation__.Notify(573476)
	}
	__antithesis_instrumentation__.Notify(573474)
	return trailingMeta
}

func (jr *joinReader) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(573477)
	if _, ok := jr.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(573479)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(573480)
	}
	__antithesis_instrumentation__.Notify(573478)
	return 0
}

func (jr *joinReader) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(573481)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(573483)
		if n, ok := jr.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(573485)
			return n
		} else {
			__antithesis_instrumentation__.Notify(573486)
		}
		__antithesis_instrumentation__.Notify(573484)
		panic("input to joinReader is not an execinfra.OpNode")
	} else {
		__antithesis_instrumentation__.Notify(573487)
	}
	__antithesis_instrumentation__.Notify(573482)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}

func (jr *joinReader) processContinuationValForRow(row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(573488)
	if !jr.groupingState.doGrouping {
		__antithesis_instrumentation__.Notify(573490)

		jr.groupingState.addContinuationValForRow(false)
	} else {
		__antithesis_instrumentation__.Notify(573491)
		continuationEncDatum := row[len(row)-1]
		if err := continuationEncDatum.EnsureDecoded(types.Bool, &jr.alloc); err != nil {
			__antithesis_instrumentation__.Notify(573493)
			return err
		} else {
			__antithesis_instrumentation__.Notify(573494)
		}
		__antithesis_instrumentation__.Notify(573492)
		continuationVal := bool(*continuationEncDatum.Datum.(*tree.DBool))
		jr.groupingState.addContinuationValForRow(continuationVal)
		if len(jr.scratchInputRows) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(573495)
			return continuationVal == true
		}() == true {
			__antithesis_instrumentation__.Notify(573496)

			jr.lastBatchState.lastGroupContinued = true
		} else {
			__antithesis_instrumentation__.Notify(573497)
		}
	}
	__antithesis_instrumentation__.Notify(573489)
	return nil
}

func (jr *joinReader) allContinuationValsProcessed() rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(573498)
	var outRow rowenc.EncDatumRow
	jr.groupingState.initialized = true
	if jr.lastBatchState.lastInputRow != nil && func() bool {
		__antithesis_instrumentation__.Notify(573500)
		return !jr.lastBatchState.lastGroupContinued == true
	}() == true {
		__antithesis_instrumentation__.Notify(573501)

		if !jr.lastBatchState.lastGroupMatched {
			__antithesis_instrumentation__.Notify(573502)

			switch jr.joinType {
			case descpb.LeftOuterJoin:
				__antithesis_instrumentation__.Notify(573503)
				outRow = jr.renderUnmatchedRow(jr.lastBatchState.lastInputRow, leftSide)
			case descpb.LeftAntiJoin:
				__antithesis_instrumentation__.Notify(573504)
				outRow = jr.lastBatchState.lastInputRow
			default:
				__antithesis_instrumentation__.Notify(573505)
			}
		} else {
			__antithesis_instrumentation__.Notify(573506)
		}

	} else {
		__antithesis_instrumentation__.Notify(573507)
	}
	__antithesis_instrumentation__.Notify(573499)

	jr.lastBatchState.lastInputRow = nil
	return outRow
}

func (jr *joinReader) updateGroupingStateForNonEmptyBatch() {
	__antithesis_instrumentation__.Notify(573508)
	if jr.groupingState.doGrouping {
		__antithesis_instrumentation__.Notify(573509)

		jr.lastBatchState.lastInputRow = jr.scratchInputRows[len(jr.scratchInputRows)-1]

		if jr.lastBatchState.lastGroupMatched && func() bool {
			__antithesis_instrumentation__.Notify(573510)
			return jr.lastBatchState.lastGroupContinued == true
		}() == true {
			__antithesis_instrumentation__.Notify(573511)
			jr.groupingState.setFirstGroupMatched()
		} else {
			__antithesis_instrumentation__.Notify(573512)
		}
	} else {
		__antithesis_instrumentation__.Notify(573513)
	}
}

type inputBatchGroupingState struct {
	doGrouping bool

	initialized bool

	batchRowToGroupIndex []int

	groupState []groupState
}

type groupState struct {
	matched bool

	lastRow int
}

func (ib *inputBatchGroupingState) reset() {
	__antithesis_instrumentation__.Notify(573514)
	ib.batchRowToGroupIndex = ib.batchRowToGroupIndex[:0]
	ib.groupState = ib.groupState[:0]
	ib.initialized = false
}

func (ib *inputBatchGroupingState) addContinuationValForRow(cont bool) {
	__antithesis_instrumentation__.Notify(573515)
	if len(ib.groupState) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(573517)
		return !cont == true
	}() == true {
		__antithesis_instrumentation__.Notify(573518)

		ib.groupState = append(ib.groupState,
			groupState{matched: false, lastRow: len(ib.batchRowToGroupIndex)})
	} else {
		__antithesis_instrumentation__.Notify(573519)
	}
	__antithesis_instrumentation__.Notify(573516)
	if ib.doGrouping {
		__antithesis_instrumentation__.Notify(573520)
		groupIndex := len(ib.groupState) - 1
		ib.groupState[groupIndex].lastRow = len(ib.batchRowToGroupIndex)
		ib.batchRowToGroupIndex = append(ib.batchRowToGroupIndex, groupIndex)
	} else {
		__antithesis_instrumentation__.Notify(573521)
	}
}

func (ib *inputBatchGroupingState) setFirstGroupMatched() {
	__antithesis_instrumentation__.Notify(573522)
	ib.groupState[0].matched = true
}

func (ib *inputBatchGroupingState) setMatched(rowIndex int) bool {
	__antithesis_instrumentation__.Notify(573523)
	groupIndex := rowIndex
	if ib.doGrouping {
		__antithesis_instrumentation__.Notify(573525)
		groupIndex = ib.batchRowToGroupIndex[rowIndex]
	} else {
		__antithesis_instrumentation__.Notify(573526)
	}
	__antithesis_instrumentation__.Notify(573524)
	rv := ib.groupState[groupIndex].matched
	ib.groupState[groupIndex].matched = true
	return rv
}

func (ib *inputBatchGroupingState) getMatched(rowIndex int) bool {
	__antithesis_instrumentation__.Notify(573527)
	groupIndex := rowIndex
	if ib.doGrouping {
		__antithesis_instrumentation__.Notify(573529)
		groupIndex = ib.batchRowToGroupIndex[rowIndex]
	} else {
		__antithesis_instrumentation__.Notify(573530)
	}
	__antithesis_instrumentation__.Notify(573528)
	return ib.groupState[groupIndex].matched
}

func (ib *inputBatchGroupingState) lastGroupMatched() bool {
	__antithesis_instrumentation__.Notify(573531)
	if !ib.doGrouping || func() bool {
		__antithesis_instrumentation__.Notify(573533)
		return len(ib.groupState) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(573534)
		return false
	} else {
		__antithesis_instrumentation__.Notify(573535)
	}
	__antithesis_instrumentation__.Notify(573532)
	return ib.groupState[len(ib.groupState)-1].matched
}

func (ib *inputBatchGroupingState) isUnmatched(rowIndex int) bool {
	__antithesis_instrumentation__.Notify(573536)
	if !ib.doGrouping {
		__antithesis_instrumentation__.Notify(573539)

		return !ib.groupState[rowIndex].matched
	} else {
		__antithesis_instrumentation__.Notify(573540)
	}
	__antithesis_instrumentation__.Notify(573537)
	groupIndex := ib.batchRowToGroupIndex[rowIndex]
	if groupIndex == len(ib.groupState)-1 {
		__antithesis_instrumentation__.Notify(573541)

		return false
	} else {
		__antithesis_instrumentation__.Notify(573542)
	}
	__antithesis_instrumentation__.Notify(573538)

	return !ib.groupState[groupIndex].matched && func() bool {
		__antithesis_instrumentation__.Notify(573543)
		return ib.groupState[groupIndex].lastRow == rowIndex == true
	}() == true
}

func (ib *inputBatchGroupingState) memUsage() int64 {
	__antithesis_instrumentation__.Notify(573544)
	if ib == nil {
		__antithesis_instrumentation__.Notify(573546)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(573547)
	}
	__antithesis_instrumentation__.Notify(573545)
	return int64(cap(ib.groupState))*int64(unsafe.Sizeof(groupState{})) +
		int64(cap(ib.batchRowToGroupIndex))*memsize.Int
}
