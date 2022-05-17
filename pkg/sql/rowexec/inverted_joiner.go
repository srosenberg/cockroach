package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/invertedexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/invertedidx"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/scrub"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

var invertedJoinerBatchSize = util.ConstantWithMetamorphicTestValue(
	"inverted-joiner-batch-size",
	100,
	1,
)

type invertedJoinerState int

const (
	ijStateUnknown invertedJoinerState = iota

	ijReadingInput

	ijPerformingIndexScan

	ijEmittingRows
)

type invertedJoiner struct {
	execinfra.ProcessorBase

	runningState invertedJoinerState
	diskMonitor  *mon.BytesMonitor
	desc         catalog.TableDescriptor

	colIdxMap catalog.TableColMap
	index     catalog.Index

	invertedColID descpb.ColumnID

	prefixEqualityCols []uint32

	onExprHelper execinfrapb.ExprHelper
	combinedRow  rowenc.EncDatumRow

	joinType descpb.JoinType

	fetcher rowFetcher
	row     rowenc.EncDatumRow

	rowsRead int64
	alloc    tree.DatumAlloc
	rowAlloc rowenc.EncDatumRowAlloc

	tableRow rowenc.EncDatumRow

	indexRow rowenc.EncDatumRow

	indexRowTypes []*types.T

	indexRowToTableRowMap []int

	input                execinfra.RowSource
	inputTypes           []*types.T
	datumsToInvertedExpr invertedexpr.DatumsToInvertedExpr
	canPreFilter         bool

	batchSize int

	inputRows       rowenc.EncDatumRows
	batchedExprEval batchedInvertedExprEvaluator

	joinedRowIdx [][]KeyIndex

	indexRows *rowcontainer.DiskBackedNumberedRowContainer

	indexSpans roachpb.Spans

	emitCursor struct {
		inputRowIdx int

		outputRowIdx int

		seenMatch bool
	}

	spanBuilder           span.Builder
	outputContinuationCol bool

	scanStats execinfra.ScanStats
}

var _ execinfra.Processor = &invertedJoiner{}
var _ execinfra.RowSource = &invertedJoiner{}
var _ execinfra.OpNode = &invertedJoiner{}

const invertedJoinerProcName = "inverted joiner"

func newInvertedJoiner(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.InvertedJoinerSpec,
	datumsToInvertedExpr invertedexpr.DatumsToInvertedExpr,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.RowSourcedProcessor, error) {
	__antithesis_instrumentation__.Notify(572856)
	switch spec.Type {
	case descpb.InnerJoin, descpb.LeftOuterJoin, descpb.LeftSemiJoin, descpb.LeftAntiJoin:
		__antithesis_instrumentation__.Notify(572871)
	default:
		__antithesis_instrumentation__.Notify(572872)
		return nil, errors.AssertionFailedf("unexpected inverted join type %s", spec.Type)
	}
	__antithesis_instrumentation__.Notify(572857)
	ij := &invertedJoiner{
		desc:                 flowCtx.TableDescriptor(&spec.Table),
		input:                input,
		inputTypes:           input.OutputTypes(),
		prefixEqualityCols:   spec.PrefixEqualityColumns,
		datumsToInvertedExpr: datumsToInvertedExpr,
		joinType:             spec.Type,
		batchSize:            invertedJoinerBatchSize,
	}
	ij.colIdxMap = catalog.ColumnIDToOrdinalMap(ij.desc.PublicColumns())

	var err error
	indexIdx := int(spec.IndexIdx)
	if indexIdx >= len(ij.desc.ActiveIndexes()) {
		__antithesis_instrumentation__.Notify(572873)
		return nil, errors.Errorf("invalid indexIdx %d", indexIdx)
	} else {
		__antithesis_instrumentation__.Notify(572874)
	}
	__antithesis_instrumentation__.Notify(572858)
	ij.index = ij.desc.ActiveIndexes()[indexIdx]
	ij.invertedColID = ij.index.InvertedColumnID()

	indexColumns := ij.desc.IndexFullColumns(ij.index)

	ij.tableRow = make(rowenc.EncDatumRow, len(ij.desc.PublicColumns()))
	ij.indexRow = make(rowenc.EncDatumRow, len(indexColumns)-1)
	ij.indexRowTypes = make([]*types.T, len(ij.indexRow))
	ij.indexRowToTableRowMap = make([]int, len(ij.indexRow))
	indexRowIdx := 0
	for _, col := range indexColumns {
		__antithesis_instrumentation__.Notify(572875)
		if col == nil {
			__antithesis_instrumentation__.Notify(572878)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572879)
		}
		__antithesis_instrumentation__.Notify(572876)
		colID := col.GetID()

		if colID == ij.invertedColID {
			__antithesis_instrumentation__.Notify(572880)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572881)
		}
		__antithesis_instrumentation__.Notify(572877)
		tableRowIdx := ij.colIdxMap.GetDefault(colID)
		ij.indexRowToTableRowMap[indexRowIdx] = tableRowIdx
		ij.indexRowTypes[indexRowIdx] = ij.desc.PublicColumns()[tableRowIdx].GetType()
		indexRowIdx++
	}
	__antithesis_instrumentation__.Notify(572859)

	outputColCount := len(ij.inputTypes)

	rightColTypes := catalog.ColumnTypes(ij.desc.PublicColumns())
	var includeRightCols bool
	if ij.joinType == descpb.InnerJoin || func() bool {
		__antithesis_instrumentation__.Notify(572882)
		return ij.joinType == descpb.LeftOuterJoin == true
	}() == true {
		__antithesis_instrumentation__.Notify(572883)
		outputColCount += len(rightColTypes)
		includeRightCols = true
		if spec.OutputGroupContinuationForLeftRow {
			__antithesis_instrumentation__.Notify(572884)
			outputColCount++
		} else {
			__antithesis_instrumentation__.Notify(572885)
		}
	} else {
		__antithesis_instrumentation__.Notify(572886)
	}
	__antithesis_instrumentation__.Notify(572860)
	outputColTypes := make([]*types.T, 0, outputColCount)
	outputColTypes = append(outputColTypes, ij.inputTypes...)
	if includeRightCols {
		__antithesis_instrumentation__.Notify(572887)
		outputColTypes = append(outputColTypes, rightColTypes...)
	} else {
		__antithesis_instrumentation__.Notify(572888)
	}
	__antithesis_instrumentation__.Notify(572861)
	if spec.OutputGroupContinuationForLeftRow {
		__antithesis_instrumentation__.Notify(572889)
		outputColTypes = append(outputColTypes, types.Bool)
	} else {
		__antithesis_instrumentation__.Notify(572890)
	}
	__antithesis_instrumentation__.Notify(572862)
	if err := ij.ProcessorBase.Init(
		ij, post, outputColTypes, flowCtx, processorID, output, nil,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{ij.input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(572891)

				trailingMeta := ij.generateMeta()
				ij.close()
				return trailingMeta
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(572892)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572893)
	}
	__antithesis_instrumentation__.Notify(572863)

	semaCtx := flowCtx.NewSemaContext(flowCtx.EvalCtx.Txn)
	onExprColTypes := make([]*types.T, 0, len(ij.inputTypes)+len(rightColTypes))
	onExprColTypes = append(onExprColTypes, ij.inputTypes...)
	onExprColTypes = append(onExprColTypes, rightColTypes...)
	if err := ij.onExprHelper.Init(spec.OnExpr, onExprColTypes, semaCtx, ij.EvalCtx); err != nil {
		__antithesis_instrumentation__.Notify(572894)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572895)
	}
	__antithesis_instrumentation__.Notify(572864)
	combinedRowLen := len(onExprColTypes)
	if spec.OutputGroupContinuationForLeftRow {
		__antithesis_instrumentation__.Notify(572896)
		combinedRowLen++
	} else {
		__antithesis_instrumentation__.Notify(572897)
	}
	__antithesis_instrumentation__.Notify(572865)
	ij.combinedRow = make(rowenc.EncDatumRow, 0, combinedRowLen)

	if ij.datumsToInvertedExpr == nil {
		__antithesis_instrumentation__.Notify(572898)
		var invertedExprHelper execinfrapb.ExprHelper
		if err := invertedExprHelper.Init(spec.InvertedExpr, onExprColTypes, semaCtx, ij.EvalCtx); err != nil {
			__antithesis_instrumentation__.Notify(572900)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(572901)
		}
		__antithesis_instrumentation__.Notify(572899)
		ij.datumsToInvertedExpr, err = invertedidx.NewDatumsToInvertedExpr(
			ij.EvalCtx, onExprColTypes, invertedExprHelper.Expr, ij.index,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(572902)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(572903)
		}
	} else {
		__antithesis_instrumentation__.Notify(572904)
	}
	__antithesis_instrumentation__.Notify(572866)
	ij.canPreFilter = ij.datumsToInvertedExpr.CanPreFilter()
	if ij.canPreFilter {
		__antithesis_instrumentation__.Notify(572905)
		ij.batchedExprEval.filterer = ij.datumsToInvertedExpr
	} else {
		__antithesis_instrumentation__.Notify(572906)
	}
	__antithesis_instrumentation__.Notify(572867)

	allIndexCols := util.MakeFastIntSet()
	for _, col := range indexColumns {
		__antithesis_instrumentation__.Notify(572907)
		if col == nil {
			__antithesis_instrumentation__.Notify(572909)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572910)
		}
		__antithesis_instrumentation__.Notify(572908)
		allIndexCols.Add(ij.colIdxMap.GetDefault(col.GetID()))
	}
	__antithesis_instrumentation__.Notify(572868)
	fetcher, err := makeRowFetcherLegacy(
		flowCtx, ij.desc, int(spec.IndexIdx), false,
		allIndexCols, flowCtx.EvalCtx.Mon, &ij.alloc,
		descpb.ScanLockingStrength_FOR_NONE, descpb.ScanLockingWaitPolicy_BLOCK,
		false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(572911)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572912)
	}
	__antithesis_instrumentation__.Notify(572869)
	ij.row = make(rowenc.EncDatumRow, len(ij.desc.PublicColumns()))

	if execinfra.ShouldCollectStats(flowCtx.EvalCtx.Ctx(), flowCtx) {
		__antithesis_instrumentation__.Notify(572913)
		ij.input = newInputStatCollector(ij.input)
		ij.fetcher = newRowFetcherStatCollector(fetcher)
		ij.ExecStatsForTrace = ij.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(572914)
		ij.fetcher = fetcher
	}
	__antithesis_instrumentation__.Notify(572870)

	ij.spanBuilder.Init(flowCtx.EvalCtx, flowCtx.Codec(), ij.desc, ij.index)

	ctx := flowCtx.EvalCtx.Ctx()
	ij.MemMonitor = execinfra.NewLimitedMonitor(ctx, flowCtx.EvalCtx.Mon, flowCtx, "invertedjoiner-limited")
	ij.diskMonitor = execinfra.NewMonitor(ctx, flowCtx.DiskMonitor, "invertedjoiner-disk")
	ij.indexRows = rowcontainer.NewDiskBackedNumberedRowContainer(
		true,
		ij.indexRowTypes,
		ij.EvalCtx,
		ij.FlowCtx.Cfg.TempStorage,
		ij.MemMonitor,
		ij.diskMonitor,
	)

	ij.outputContinuationCol = spec.OutputGroupContinuationForLeftRow

	return ij, nil
}

func (ij *invertedJoiner) SetBatchSize(batchSize int) {
	__antithesis_instrumentation__.Notify(572915)
	ij.batchSize = batchSize
}

func (ij *invertedJoiner) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572916)

	for ij.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(572918)
		var row rowenc.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch ij.runningState {
		case ijReadingInput:
			__antithesis_instrumentation__.Notify(572922)
			ij.runningState, meta = ij.readInput()
		case ijPerformingIndexScan:
			__antithesis_instrumentation__.Notify(572923)
			ij.runningState, meta = ij.performScan()
		case ijEmittingRows:
			__antithesis_instrumentation__.Notify(572924)
			ij.runningState, row, meta = ij.emitRow()
		default:
			__antithesis_instrumentation__.Notify(572925)
			log.Fatalf(ij.Ctx, "unsupported state: %d", ij.runningState)
		}
		__antithesis_instrumentation__.Notify(572919)
		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(572926)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(572927)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572928)
		}
		__antithesis_instrumentation__.Notify(572920)
		if meta != nil {
			__antithesis_instrumentation__.Notify(572929)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(572930)
		}
		__antithesis_instrumentation__.Notify(572921)
		if outRow := ij.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(572931)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(572932)
		}
	}
	__antithesis_instrumentation__.Notify(572917)
	return nil, ij.DrainHelper()
}

func (ij *invertedJoiner) readInput() (invertedJoinerState, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572933)

	for len(ij.inputRows) < ij.batchSize {
		__antithesis_instrumentation__.Notify(572940)
		row, meta := ij.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(572946)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(572948)
				ij.MoveToDraining(nil)
				return ijStateUnknown, meta
			} else {
				__antithesis_instrumentation__.Notify(572949)
			}
			__antithesis_instrumentation__.Notify(572947)
			return ijReadingInput, meta
		} else {
			__antithesis_instrumentation__.Notify(572950)
		}
		__antithesis_instrumentation__.Notify(572941)
		if row == nil {
			__antithesis_instrumentation__.Notify(572951)
			break
		} else {
			__antithesis_instrumentation__.Notify(572952)
		}
		__antithesis_instrumentation__.Notify(572942)

		expr, preFilterState, err := ij.datumsToInvertedExpr.Convert(ij.Ctx, row)
		if err != nil {
			__antithesis_instrumentation__.Notify(572953)
			ij.MoveToDraining(err)
			return ijStateUnknown, ij.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572954)
		}
		__antithesis_instrumentation__.Notify(572943)
		if expr == nil && func() bool {
			__antithesis_instrumentation__.Notify(572955)
			return (ij.joinType != descpb.LeftOuterJoin && func() bool {
				__antithesis_instrumentation__.Notify(572956)
				return ij.joinType != descpb.LeftAntiJoin == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(572957)

			ij.inputRows = append(ij.inputRows, nil)
		} else {
			__antithesis_instrumentation__.Notify(572958)
			ij.inputRows = append(ij.inputRows, ij.rowAlloc.CopyRow(row))
		}
		__antithesis_instrumentation__.Notify(572944)
		if expr == nil {
			__antithesis_instrumentation__.Notify(572959)

			ij.batchedExprEval.exprs = append(ij.batchedExprEval.exprs, nil)
			if ij.canPreFilter {
				__antithesis_instrumentation__.Notify(572960)
				ij.batchedExprEval.preFilterState = append(ij.batchedExprEval.preFilterState, nil)
			} else {
				__antithesis_instrumentation__.Notify(572961)
			}
		} else {
			__antithesis_instrumentation__.Notify(572962)
			ij.batchedExprEval.exprs = append(ij.batchedExprEval.exprs, expr)
			if ij.canPreFilter {
				__antithesis_instrumentation__.Notify(572963)
				ij.batchedExprEval.preFilterState = append(ij.batchedExprEval.preFilterState, preFilterState)
			} else {
				__antithesis_instrumentation__.Notify(572964)
			}
		}
		__antithesis_instrumentation__.Notify(572945)
		if len(ij.prefixEqualityCols) > 0 {
			__antithesis_instrumentation__.Notify(572965)
			if expr == nil {
				__antithesis_instrumentation__.Notify(572966)

				ij.batchedExprEval.nonInvertedPrefixes = append(ij.batchedExprEval.nonInvertedPrefixes, roachpb.Key{})
			} else {
				__antithesis_instrumentation__.Notify(572967)
				for prefixIdx, colIdx := range ij.prefixEqualityCols {
					__antithesis_instrumentation__.Notify(572970)
					ij.indexRow[prefixIdx] = row[colIdx]
				}
				__antithesis_instrumentation__.Notify(572968)

				keyCols := ij.desc.IndexFetchSpecKeyAndSuffixColumns(ij.index)
				prefixKey, _, err := rowenc.MakeKeyFromEncDatums(
					ij.indexRow[:len(ij.prefixEqualityCols)],
					keyCols,
					&ij.alloc,
					nil,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(572971)
					ij.MoveToDraining(err)
					return ijStateUnknown, ij.DrainHelper()
				} else {
					__antithesis_instrumentation__.Notify(572972)
				}
				__antithesis_instrumentation__.Notify(572969)
				ij.batchedExprEval.nonInvertedPrefixes = append(ij.batchedExprEval.nonInvertedPrefixes, prefixKey)
			}
		} else {
			__antithesis_instrumentation__.Notify(572973)
		}
	}
	__antithesis_instrumentation__.Notify(572934)

	if len(ij.inputRows) == 0 {
		__antithesis_instrumentation__.Notify(572974)
		log.VEventf(ij.Ctx, 1, "no more input rows")

		ij.MoveToDraining(nil)
		return ijStateUnknown, ij.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572975)
	}
	__antithesis_instrumentation__.Notify(572935)
	log.VEventf(ij.Ctx, 1, "read %d input rows", len(ij.inputRows))

	spans, err := ij.batchedExprEval.init()
	if err != nil {
		__antithesis_instrumentation__.Notify(572976)
		ij.MoveToDraining(err)
		return ijStateUnknown, ij.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572977)
	}
	__antithesis_instrumentation__.Notify(572936)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(572978)

		ij.joinedRowIdx = ij.joinedRowIdx[:0]
		for range ij.inputRows {
			__antithesis_instrumentation__.Notify(572980)
			ij.joinedRowIdx = append(ij.joinedRowIdx, nil)
		}
		__antithesis_instrumentation__.Notify(572979)
		return ijEmittingRows, nil
	} else {
		__antithesis_instrumentation__.Notify(572981)
	}
	__antithesis_instrumentation__.Notify(572937)

	ij.indexSpans, err = ij.spanBuilder.SpansFromInvertedSpans(spans, nil, ij.indexSpans)
	if err != nil {
		__antithesis_instrumentation__.Notify(572982)
		ij.MoveToDraining(err)
		return ijStateUnknown, ij.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572983)
	}
	__antithesis_instrumentation__.Notify(572938)

	log.VEventf(ij.Ctx, 1, "scanning %d spans", len(ij.indexSpans))
	if err = ij.fetcher.StartScan(
		ij.Ctx, ij.FlowCtx.Txn, ij.indexSpans, rowinfra.NoBytesLimit, rowinfra.NoRowLimit,
		ij.FlowCtx.TraceKV, ij.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
	); err != nil {
		__antithesis_instrumentation__.Notify(572984)
		ij.MoveToDraining(err)
		return ijStateUnknown, ij.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572985)
	}
	__antithesis_instrumentation__.Notify(572939)

	return ijPerformingIndexScan, nil
}

func (ij *invertedJoiner) performScan() (invertedJoinerState, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572986)
	log.VEventf(ij.Ctx, 1, "joining rows")

	for {
		__antithesis_instrumentation__.Notify(572988)

		ok, err := ij.fetcher.NextRowInto(ij.Ctx, ij.row, ij.colIdxMap)
		if err != nil {
			__antithesis_instrumentation__.Notify(572993)
			ij.MoveToDraining(scrub.UnwrapScrubError(err))
			return ijStateUnknown, ij.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572994)
		}
		__antithesis_instrumentation__.Notify(572989)
		if !ok {
			__antithesis_instrumentation__.Notify(572995)

			break
		} else {
			__antithesis_instrumentation__.Notify(572996)
		}
		__antithesis_instrumentation__.Notify(572990)
		scannedRow := ij.row
		ij.rowsRead++

		ij.transformToIndexRow(scannedRow)
		idx := ij.colIdxMap.GetDefault(ij.invertedColID)
		encInvertedVal := scannedRow[idx].EncodedBytes()
		var encFullVal []byte
		if len(ij.prefixEqualityCols) > 0 {
			__antithesis_instrumentation__.Notify(572997)

			keyCols := ij.desc.IndexFetchSpecKeyAndSuffixColumns(ij.index)
			prefixKey, _, err := rowenc.MakeKeyFromEncDatums(
				ij.indexRow[:len(ij.prefixEqualityCols)],
				keyCols,
				&ij.alloc,
				nil,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(572999)
				ij.MoveToDraining(err)
				return ijStateUnknown, ij.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(573000)
			}
			__antithesis_instrumentation__.Notify(572998)

			encFullVal = append(prefixKey, encInvertedVal...)
		} else {
			__antithesis_instrumentation__.Notify(573001)
		}
		__antithesis_instrumentation__.Notify(572991)
		shouldAdd, err := ij.batchedExprEval.prepareAddIndexRow(encInvertedVal, encFullVal)
		if err != nil {
			__antithesis_instrumentation__.Notify(573002)
			ij.MoveToDraining(err)
			return ijStateUnknown, ij.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573003)
		}
		__antithesis_instrumentation__.Notify(572992)
		if shouldAdd {
			__antithesis_instrumentation__.Notify(573004)
			rowIdx, err := ij.indexRows.AddRow(ij.Ctx, ij.indexRow)
			if err != nil {
				__antithesis_instrumentation__.Notify(573006)
				ij.MoveToDraining(err)
				return ijStateUnknown, ij.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(573007)
			}
			__antithesis_instrumentation__.Notify(573005)
			if err = ij.batchedExprEval.addIndexRow(rowIdx); err != nil {
				__antithesis_instrumentation__.Notify(573008)
				ij.MoveToDraining(err)
				return ijStateUnknown, ij.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(573009)
			}
		} else {
			__antithesis_instrumentation__.Notify(573010)
		}
	}
	__antithesis_instrumentation__.Notify(572987)
	ij.joinedRowIdx = ij.batchedExprEval.evaluate()
	ij.indexRows.SetupForRead(ij.Ctx, ij.joinedRowIdx)
	log.VEventf(ij.Ctx, 1, "done evaluating expressions")

	return ijEmittingRows, nil
}

var trueEncDatum = rowenc.DatumToEncDatum(types.Bool, tree.DBoolTrue)
var falseEncDatum = rowenc.DatumToEncDatum(types.Bool, tree.DBoolFalse)

func (ij *invertedJoiner) emitRow() (
	invertedJoinerState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(573011)

	if ij.emitCursor.inputRowIdx >= len(ij.joinedRowIdx) {
		__antithesis_instrumentation__.Notify(573018)
		log.VEventf(ij.Ctx, 1, "done emitting rows")

		ij.inputRows = ij.inputRows[:0]
		ij.batchedExprEval.reset()
		ij.joinedRowIdx = nil
		ij.emitCursor.outputRowIdx = 0
		ij.emitCursor.inputRowIdx = 0
		ij.emitCursor.seenMatch = false
		if err := ij.indexRows.UnsafeReset(ij.Ctx); err != nil {
			__antithesis_instrumentation__.Notify(573020)
			ij.MoveToDraining(err)
			return ijStateUnknown, nil, ij.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(573021)
		}
		__antithesis_instrumentation__.Notify(573019)
		return ijReadingInput, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(573022)
	}
	__antithesis_instrumentation__.Notify(573012)

	if ij.emitCursor.outputRowIdx >= len(ij.joinedRowIdx[ij.emitCursor.inputRowIdx]) {
		__antithesis_instrumentation__.Notify(573023)
		inputRowIdx := ij.emitCursor.inputRowIdx
		seenMatch := ij.emitCursor.seenMatch
		ij.emitCursor.inputRowIdx++
		ij.emitCursor.outputRowIdx = 0
		ij.emitCursor.seenMatch = false

		if !seenMatch {
			__antithesis_instrumentation__.Notify(573025)
			switch ij.joinType {
			case descpb.LeftOuterJoin:
				__antithesis_instrumentation__.Notify(573026)
				ij.renderUnmatchedRow(ij.inputRows[inputRowIdx])
				return ijEmittingRows, ij.combinedRow, nil
			case descpb.LeftAntiJoin:
				__antithesis_instrumentation__.Notify(573027)
				return ijEmittingRows, ij.inputRows[inputRowIdx], nil
			default:
				__antithesis_instrumentation__.Notify(573028)
			}
		} else {
			__antithesis_instrumentation__.Notify(573029)
		}
		__antithesis_instrumentation__.Notify(573024)
		return ijEmittingRows, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(573030)
	}
	__antithesis_instrumentation__.Notify(573013)

	inputRow := ij.inputRows[ij.emitCursor.inputRowIdx]
	joinedRowIdx := ij.joinedRowIdx[ij.emitCursor.inputRowIdx][ij.emitCursor.outputRowIdx]
	indexedRow, err := ij.indexRows.GetRow(ij.Ctx, joinedRowIdx, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(573031)
		ij.MoveToDraining(err)
		return ijStateUnknown, nil, ij.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(573032)
	}
	__antithesis_instrumentation__.Notify(573014)
	ij.emitCursor.outputRowIdx++
	ij.transformToTableRow(indexedRow)
	renderedRow, err := ij.render(inputRow, ij.tableRow)
	if err != nil {
		__antithesis_instrumentation__.Notify(573033)
		ij.MoveToDraining(err)
		return ijStateUnknown, nil, ij.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(573034)
	}
	__antithesis_instrumentation__.Notify(573015)
	skipRemaining := func() error {
		__antithesis_instrumentation__.Notify(573035)
		for ; ij.emitCursor.outputRowIdx < len(ij.joinedRowIdx[ij.emitCursor.inputRowIdx]); ij.emitCursor.outputRowIdx++ {
			__antithesis_instrumentation__.Notify(573037)
			idx := ij.joinedRowIdx[ij.emitCursor.inputRowIdx][ij.emitCursor.outputRowIdx]
			if _, err := ij.indexRows.GetRow(ij.Ctx, idx, true); err != nil {
				__antithesis_instrumentation__.Notify(573038)
				return err
			} else {
				__antithesis_instrumentation__.Notify(573039)
			}
		}
		__antithesis_instrumentation__.Notify(573036)
		return nil
	}
	__antithesis_instrumentation__.Notify(573016)
	if renderedRow != nil {
		__antithesis_instrumentation__.Notify(573040)
		seenMatch := ij.emitCursor.seenMatch
		ij.emitCursor.seenMatch = true
		switch ij.joinType {
		case descpb.InnerJoin, descpb.LeftOuterJoin:
			__antithesis_instrumentation__.Notify(573041)
			if ij.outputContinuationCol {
				__antithesis_instrumentation__.Notify(573048)
				if seenMatch {
					__antithesis_instrumentation__.Notify(573050)

					ij.combinedRow = append(ij.combinedRow, trueEncDatum)
				} else {
					__antithesis_instrumentation__.Notify(573051)

					ij.combinedRow = append(ij.combinedRow, falseEncDatum)
				}
				__antithesis_instrumentation__.Notify(573049)
				renderedRow = ij.combinedRow
			} else {
				__antithesis_instrumentation__.Notify(573052)
			}
			__antithesis_instrumentation__.Notify(573042)
			return ijEmittingRows, renderedRow, nil
		case descpb.LeftSemiJoin:
			__antithesis_instrumentation__.Notify(573043)

			if err := skipRemaining(); err != nil {
				__antithesis_instrumentation__.Notify(573053)
				ij.MoveToDraining(err)
				return ijStateUnknown, nil, ij.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(573054)
			}
			__antithesis_instrumentation__.Notify(573044)
			return ijEmittingRows, inputRow, nil
		case descpb.LeftAntiJoin:
			__antithesis_instrumentation__.Notify(573045)

			if err := skipRemaining(); err != nil {
				__antithesis_instrumentation__.Notify(573055)
				ij.MoveToDraining(err)
				return ijStateUnknown, nil, ij.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(573056)
			}
			__antithesis_instrumentation__.Notify(573046)
			ij.emitCursor.outputRowIdx = len(ij.joinedRowIdx[ij.emitCursor.inputRowIdx])
		default:
			__antithesis_instrumentation__.Notify(573047)
		}
	} else {
		__antithesis_instrumentation__.Notify(573057)
	}
	__antithesis_instrumentation__.Notify(573017)
	return ijEmittingRows, nil, nil
}

func (ij *invertedJoiner) render(lrow, rrow rowenc.EncDatumRow) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(573058)
	ij.combinedRow = append(ij.combinedRow[:0], lrow...)
	ij.combinedRow = append(ij.combinedRow, rrow...)
	if ij.onExprHelper.Expr != nil {
		__antithesis_instrumentation__.Notify(573060)
		res, err := ij.onExprHelper.EvalFilter(ij.combinedRow)
		if !res || func() bool {
			__antithesis_instrumentation__.Notify(573061)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(573062)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(573063)
		}
	} else {
		__antithesis_instrumentation__.Notify(573064)
	}
	__antithesis_instrumentation__.Notify(573059)
	return ij.combinedRow, nil
}

func (ij *invertedJoiner) renderUnmatchedRow(row rowenc.EncDatumRow) {
	__antithesis_instrumentation__.Notify(573065)
	ij.combinedRow = ij.combinedRow[:cap(ij.combinedRow)]

	copy(ij.combinedRow, row)

	for i := len(row); i < len(ij.combinedRow); i++ {
		__antithesis_instrumentation__.Notify(573067)
		ij.combinedRow[i].Datum = tree.DNull
	}
	__antithesis_instrumentation__.Notify(573066)
	if ij.outputContinuationCol {
		__antithesis_instrumentation__.Notify(573068)

		ij.combinedRow[len(ij.combinedRow)-1] = falseEncDatum
	} else {
		__antithesis_instrumentation__.Notify(573069)
	}
}

func (ij *invertedJoiner) transformToIndexRow(row rowenc.EncDatumRow) {
	__antithesis_instrumentation__.Notify(573070)
	for keyIdx, rowIdx := range ij.indexRowToTableRowMap {
		__antithesis_instrumentation__.Notify(573071)
		ij.indexRow[keyIdx] = row[rowIdx]
	}
}

func (ij *invertedJoiner) transformToTableRow(indexRow rowenc.EncDatumRow) {
	__antithesis_instrumentation__.Notify(573072)
	for keyIdx, rowIdx := range ij.indexRowToTableRowMap {
		__antithesis_instrumentation__.Notify(573073)
		ij.tableRow[rowIdx] = indexRow[keyIdx]
	}
}

func (ij *invertedJoiner) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573074)
	ctx = ij.StartInternal(ctx, invertedJoinerProcName)
	ij.input.Start(ctx)
	ij.runningState = ijReadingInput
}

func (ij *invertedJoiner) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(573075)

	ij.close()
}

func (ij *invertedJoiner) close() {
	__antithesis_instrumentation__.Notify(573076)
	if ij.InternalClose() {
		__antithesis_instrumentation__.Notify(573077)
		if ij.fetcher != nil {
			__antithesis_instrumentation__.Notify(573080)
			ij.fetcher.Close(ij.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(573081)
		}
		__antithesis_instrumentation__.Notify(573078)
		if ij.indexRows != nil {
			__antithesis_instrumentation__.Notify(573082)
			ij.indexRows.Close(ij.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(573083)
		}
		__antithesis_instrumentation__.Notify(573079)
		ij.MemMonitor.Stop(ij.Ctx)
		if ij.diskMonitor != nil {
			__antithesis_instrumentation__.Notify(573084)
			ij.diskMonitor.Stop(ij.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(573085)
		}
	} else {
		__antithesis_instrumentation__.Notify(573086)
	}
}

func (ij *invertedJoiner) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(573087)
	is, ok := getInputStats(ij.input)
	if !ok {
		__antithesis_instrumentation__.Notify(573090)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573091)
	}
	__antithesis_instrumentation__.Notify(573088)
	fis, ok := getFetcherInputStats(ij.fetcher)
	if !ok {
		__antithesis_instrumentation__.Notify(573092)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573093)
	}
	__antithesis_instrumentation__.Notify(573089)
	ij.scanStats = execinfra.GetScanStats(ij.Ctx)
	ret := execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		KV: execinfrapb.KVStats{
			BytesRead:      optional.MakeUint(uint64(ij.fetcher.GetBytesRead())),
			TuplesRead:     fis.NumTuples,
			KVTime:         fis.WaitTime,
			ContentionTime: optional.MakeTimeValue(execinfra.GetCumulativeContentionTime(ij.Ctx)),
		},
		Exec: execinfrapb.ExecStats{
			MaxAllocatedMem:  optional.MakeUint(uint64(ij.MemMonitor.MaximumBytes())),
			MaxAllocatedDisk: optional.MakeUint(uint64(ij.diskMonitor.MaximumBytes())),
		},
		Output: ij.OutputHelper.Stats(),
	}
	execinfra.PopulateKVMVCCStats(&ret.KV, &ij.scanStats)
	return &ret
}

func (ij *invertedJoiner) generateMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(573094)
	trailingMeta := make([]execinfrapb.ProducerMetadata, 1, 2)
	meta := &trailingMeta[0]
	meta.Metrics = execinfrapb.GetMetricsMeta()
	meta.Metrics.BytesRead = ij.fetcher.GetBytesRead()
	meta.Metrics.RowsRead = ij.rowsRead
	if tfs := execinfra.GetLeafTxnFinalState(ij.Ctx, ij.FlowCtx.Txn); tfs != nil {
		__antithesis_instrumentation__.Notify(573096)
		trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{LeafTxnFinalState: tfs})
	} else {
		__antithesis_instrumentation__.Notify(573097)
	}
	__antithesis_instrumentation__.Notify(573095)
	return trailingMeta
}

func (ij *invertedJoiner) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(573098)
	if _, ok := ij.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(573100)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(573101)
	}
	__antithesis_instrumentation__.Notify(573099)
	return 0
}

func (ij *invertedJoiner) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(573102)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(573104)
		if n, ok := ij.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(573106)
			return n
		} else {
			__antithesis_instrumentation__.Notify(573107)
		}
		__antithesis_instrumentation__.Notify(573105)
		panic("input to invertedJoiner is not an execinfra.OpNode")
	} else {
		__antithesis_instrumentation__.Notify(573108)
	}
	__antithesis_instrumentation__.Notify(573103)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
