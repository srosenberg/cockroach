package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

type windowerState int

const (
	windowerStateUnknown windowerState = iota

	windowerAccumulating

	windowerEmittingRows
)

const memRequiredByWindower = 100 * 1024

type windower struct {
	execinfra.ProcessorBase

	runningState windowerState
	input        execinfra.RowSource
	inputDone    bool
	inputTypes   []*types.T
	outputTypes  []*types.T
	datumAlloc   tree.DatumAlloc
	acc          mon.BoundAccount
	diskMonitor  *mon.BytesMonitor

	scratch       []byte
	cancelChecker cancelchecker.CancelChecker

	partitionBy                []uint32
	allRowsPartitioned         *rowcontainer.HashDiskBackedRowContainer
	partition                  *rowcontainer.DiskBackedIndexedRowContainer
	orderOfWindowFnsProcessing []int
	windowFns                  []*windowFunc
	builtins                   []tree.WindowFunc

	populated           bool
	partitionIdx        int
	rowsInBucketEmitted int
	partitionSizes      []int
	windowValues        [][][]tree.Datum
	allRowsIterator     rowcontainer.RowIterator
	outputRow           rowenc.EncDatumRow
}

var _ execinfra.Processor = &windower{}
var _ execinfra.RowSource = &windower{}
var _ execinfra.OpNode = &windower{}

const windowerProcName = "windower"

func newWindower(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.WindowerSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*windower, error) {
	__antithesis_instrumentation__.Notify(575244)
	w := &windower{
		input: input,
	}
	evalCtx := flowCtx.NewEvalCtx()
	w.inputTypes = input.OutputTypes()
	ctx := evalCtx.Ctx()

	w.partitionBy = spec.PartitionBy
	windowFns := spec.WindowFns
	w.windowFns = make([]*windowFunc, 0, len(windowFns))
	w.builtins = make([]tree.WindowFunc, 0, len(windowFns))

	w.outputTypes = make([]*types.T, len(w.inputTypes)+len(windowFns))
	copy(w.outputTypes, w.inputTypes)
	for _, windowFn := range windowFns {
		__antithesis_instrumentation__.Notify(575250)

		argTypes := make([]*types.T, len(windowFn.ArgsIdxs))
		for i, argIdx := range windowFn.ArgsIdxs {
			__antithesis_instrumentation__.Notify(575253)
			argTypes[i] = w.inputTypes[argIdx]
		}
		__antithesis_instrumentation__.Notify(575251)
		windowConstructor, outputType, err := execinfra.GetWindowFunctionInfo(windowFn.Func, argTypes...)
		if err != nil {
			__antithesis_instrumentation__.Notify(575254)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575255)
		}
		__antithesis_instrumentation__.Notify(575252)
		w.outputTypes[windowFn.OutputColIdx] = outputType

		w.builtins = append(w.builtins, windowConstructor(evalCtx))
		wf := &windowFunc{
			ordering:     windowFn.Ordering,
			argsIdxs:     windowFn.ArgsIdxs,
			frame:        windowFn.Frame,
			filterColIdx: int(windowFn.FilterColIdx),
			outputColIdx: int(windowFn.OutputColIdx),
		}

		w.windowFns = append(w.windowFns, wf)
	}
	__antithesis_instrumentation__.Notify(575245)
	w.outputRow = make(rowenc.EncDatumRow, len(w.outputTypes))

	limit := execinfra.GetWorkMemLimit(flowCtx)
	if limit < memRequiredByWindower {
		__antithesis_instrumentation__.Notify(575256)
		if !flowCtx.Cfg.TestingKnobs.ForceDiskSpill && func() bool {
			__antithesis_instrumentation__.Notify(575258)
			return flowCtx.Cfg.TestingKnobs.MemoryLimitBytes == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(575259)
			return nil, errors.Errorf(
				"window functions require %d bytes of RAM but only %d are in the budget. "+
					"Consider increasing sql.distsql.temp_storage.workmem cluster setting or distsql_workmem session variable",
				memRequiredByWindower, limit)
		} else {
			__antithesis_instrumentation__.Notify(575260)
		}
		__antithesis_instrumentation__.Notify(575257)

		limit = memRequiredByWindower
	} else {
		__antithesis_instrumentation__.Notify(575261)
	}
	__antithesis_instrumentation__.Notify(575246)
	limitedMon := mon.NewMonitorInheritWithLimit("windower-limited", limit, evalCtx.Mon)
	limitedMon.Start(ctx, evalCtx.Mon, mon.BoundAccount{})

	if err := w.InitWithEvalCtx(
		w,
		post,
		w.outputTypes,
		flowCtx,
		evalCtx,
		processorID,
		output,
		limitedMon,
		execinfra.ProcStateOpts{InputsToDrain: []execinfra.RowSource{w.input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(575262)
				w.close()
				return nil
			}},
	); err != nil {
		__antithesis_instrumentation__.Notify(575263)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(575264)
	}
	__antithesis_instrumentation__.Notify(575247)

	w.diskMonitor = execinfra.NewMonitor(ctx, flowCtx.DiskMonitor, "windower-disk")
	w.allRowsPartitioned = rowcontainer.NewHashDiskBackedRowContainer(
		evalCtx, w.MemMonitor, w.diskMonitor, flowCtx.Cfg.TempStorage,
	)
	if err := w.allRowsPartitioned.Init(
		ctx,
		false,
		w.inputTypes,
		w.partitionBy,
		true,
	); err != nil {
		__antithesis_instrumentation__.Notify(575265)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(575266)
	}
	__antithesis_instrumentation__.Notify(575248)

	w.acc = w.MemMonitor.MakeBoundAccount()

	evalCtx.SingleDatumAggMemAccount = &w.acc

	if execinfra.ShouldCollectStats(ctx, flowCtx) {
		__antithesis_instrumentation__.Notify(575267)
		w.input = newInputStatCollector(w.input)
		w.ExecStatsForTrace = w.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(575268)
	}
	__antithesis_instrumentation__.Notify(575249)

	return w, nil
}

func (w *windower) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575269)
	ctx = w.StartInternal(ctx, windowerProcName)
	w.input.Start(ctx)
	w.cancelChecker.Reset(ctx)
	w.runningState = windowerAccumulating
}

func (w *windower) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575270)
	for w.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(575272)
		var row rowenc.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch w.runningState {
		case windowerAccumulating:
			__antithesis_instrumentation__.Notify(575275)
			w.runningState, row, meta = w.accumulateRows()
		case windowerEmittingRows:
			__antithesis_instrumentation__.Notify(575276)
			w.runningState, row, meta = w.emitRow()
		default:
			__antithesis_instrumentation__.Notify(575277)
			log.Fatalf(w.Ctx, "unsupported state: %d", w.runningState)
		}
		__antithesis_instrumentation__.Notify(575273)

		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(575278)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(575279)
			continue
		} else {
			__antithesis_instrumentation__.Notify(575280)
		}
		__antithesis_instrumentation__.Notify(575274)
		return row, meta
	}
	__antithesis_instrumentation__.Notify(575271)
	return nil, w.DrainHelper()
}

func (w *windower) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(575281)

	w.close()
}

func (w *windower) close() {
	__antithesis_instrumentation__.Notify(575282)
	if w.InternalClose() {
		__antithesis_instrumentation__.Notify(575283)
		if w.allRowsIterator != nil {
			__antithesis_instrumentation__.Notify(575287)
			w.allRowsIterator.Close()
		} else {
			__antithesis_instrumentation__.Notify(575288)
		}
		__antithesis_instrumentation__.Notify(575284)
		w.allRowsPartitioned.Close(w.Ctx)
		if w.partition != nil {
			__antithesis_instrumentation__.Notify(575289)
			w.partition.Close(w.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(575290)
		}
		__antithesis_instrumentation__.Notify(575285)
		for _, builtin := range w.builtins {
			__antithesis_instrumentation__.Notify(575291)
			builtin.Close(w.Ctx, w.EvalCtx)
		}
		__antithesis_instrumentation__.Notify(575286)
		w.acc.Close(w.Ctx)
		w.MemMonitor.Stop(w.Ctx)
		w.diskMonitor.Stop(w.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(575292)
	}
}

func (w *windower) accumulateRows() (
	windowerState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(575293)
	for {
		__antithesis_instrumentation__.Notify(575295)
		row, meta := w.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(575298)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(575300)

				w.MoveToDraining(nil)
				return windowerStateUnknown, nil, meta
			} else {
				__antithesis_instrumentation__.Notify(575301)
			}
			__antithesis_instrumentation__.Notify(575299)
			return windowerAccumulating, nil, meta
		} else {
			__antithesis_instrumentation__.Notify(575302)
		}
		__antithesis_instrumentation__.Notify(575296)
		if row == nil {
			__antithesis_instrumentation__.Notify(575303)
			log.VEvent(w.Ctx, 1, "accumulation complete")
			w.inputDone = true

			w.allRowsPartitioned.Sort(w.Ctx)
			break
		} else {
			__antithesis_instrumentation__.Notify(575304)
		}
		__antithesis_instrumentation__.Notify(575297)

		if err := w.allRowsPartitioned.AddRow(w.Ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(575305)
			w.MoveToDraining(err)
			return windowerStateUnknown, nil, w.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(575306)
		}
	}
	__antithesis_instrumentation__.Notify(575294)

	return windowerEmittingRows, nil, nil
}

func (w *windower) emitRow() (windowerState, rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575307)
	if w.inputDone {
		__antithesis_instrumentation__.Notify(575309)
		for !w.populated {
			__antithesis_instrumentation__.Notify(575312)
			if err := w.cancelChecker.Check(); err != nil {
				__antithesis_instrumentation__.Notify(575315)
				w.MoveToDraining(err)
				return windowerStateUnknown, nil, w.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(575316)
			}
			__antithesis_instrumentation__.Notify(575313)

			if err := w.computeWindowFunctions(w.Ctx, w.EvalCtx); err != nil {
				__antithesis_instrumentation__.Notify(575317)
				w.MoveToDraining(err)
				return windowerStateUnknown, nil, w.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(575318)
			}
			__antithesis_instrumentation__.Notify(575314)
			w.populated = true
		}
		__antithesis_instrumentation__.Notify(575310)

		if rowOutputted, err := w.populateNextOutputRow(); err != nil {
			__antithesis_instrumentation__.Notify(575319)
			w.MoveToDraining(err)
			return windowerStateUnknown, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(575320)
			if rowOutputted {
				__antithesis_instrumentation__.Notify(575321)
				return windowerEmittingRows, w.ProcessRowHelper(w.outputRow), nil
			} else {
				__antithesis_instrumentation__.Notify(575322)
			}
		}
		__antithesis_instrumentation__.Notify(575311)

		w.MoveToDraining(nil)
		return windowerStateUnknown, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(575323)
	}
	__antithesis_instrumentation__.Notify(575308)

	w.MoveToDraining(errors.Errorf("unexpected: emitRow() is called on a windower before all input rows are accumulated"))
	return windowerStateUnknown, nil, w.DrainHelper()
}

func (w *windower) spillAllRowsToDisk() error {
	__antithesis_instrumentation__.Notify(575324)
	if w.allRowsPartitioned != nil {
		__antithesis_instrumentation__.Notify(575326)
		if !w.allRowsPartitioned.UsingDisk() {
			__antithesis_instrumentation__.Notify(575327)
			if err := w.allRowsPartitioned.SpillToDisk(w.Ctx); err != nil {
				__antithesis_instrumentation__.Notify(575328)
				return err
			} else {
				__antithesis_instrumentation__.Notify(575329)
			}
		} else {
			__antithesis_instrumentation__.Notify(575330)

			if w.partition != nil {
				__antithesis_instrumentation__.Notify(575331)
				if !w.partition.UsingDisk() {
					__antithesis_instrumentation__.Notify(575332)
					if err := w.partition.SpillToDisk(w.Ctx); err != nil {
						__antithesis_instrumentation__.Notify(575333)
						return err
					} else {
						__antithesis_instrumentation__.Notify(575334)
					}
				} else {
					__antithesis_instrumentation__.Notify(575335)
				}
			} else {
				__antithesis_instrumentation__.Notify(575336)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(575337)
	}
	__antithesis_instrumentation__.Notify(575325)
	return nil
}

func (w *windower) growMemAccount(acc *mon.BoundAccount, usage int64) error {
	__antithesis_instrumentation__.Notify(575338)
	if err := acc.Grow(w.Ctx, usage); err != nil {
		__antithesis_instrumentation__.Notify(575340)
		if sqlerrors.IsOutOfMemoryError(err) {
			__antithesis_instrumentation__.Notify(575341)
			if err := w.spillAllRowsToDisk(); err != nil {
				__antithesis_instrumentation__.Notify(575343)
				return err
			} else {
				__antithesis_instrumentation__.Notify(575344)
			}
			__antithesis_instrumentation__.Notify(575342)
			if err := acc.Grow(w.Ctx, usage); err != nil {
				__antithesis_instrumentation__.Notify(575345)
				return err
			} else {
				__antithesis_instrumentation__.Notify(575346)
			}
		} else {
			__antithesis_instrumentation__.Notify(575347)
			return err
		}
	} else {
		__antithesis_instrumentation__.Notify(575348)
	}
	__antithesis_instrumentation__.Notify(575339)
	return nil
}

func (w *windower) findOrderOfWindowFnsToProcessIn() {
	__antithesis_instrumentation__.Notify(575349)
	w.orderOfWindowFnsProcessing = make([]int, 0, len(w.windowFns))
	windowFnAdded := make([]bool, len(w.windowFns))
	for i, windowFn := range w.windowFns {
		__antithesis_instrumentation__.Notify(575350)
		if !windowFnAdded[i] {
			__antithesis_instrumentation__.Notify(575352)
			w.orderOfWindowFnsProcessing = append(w.orderOfWindowFnsProcessing, i)
			windowFnAdded[i] = true
		} else {
			__antithesis_instrumentation__.Notify(575353)
		}
		__antithesis_instrumentation__.Notify(575351)
		for j := i + 1; j < len(w.windowFns); j++ {
			__antithesis_instrumentation__.Notify(575354)
			if windowFnAdded[j] {
				__antithesis_instrumentation__.Notify(575356)

				continue
			} else {
				__antithesis_instrumentation__.Notify(575357)
			}
			__antithesis_instrumentation__.Notify(575355)
			if windowFn.ordering.Equal(w.windowFns[j].ordering) {
				__antithesis_instrumentation__.Notify(575358)
				w.orderOfWindowFnsProcessing = append(w.orderOfWindowFnsProcessing, j)
				windowFnAdded[j] = true
			} else {
				__antithesis_instrumentation__.Notify(575359)
			}
		}
	}
}

func (w *windower) processPartition(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	partition *rowcontainer.DiskBackedIndexedRowContainer,
	partitionIdx int,
) error {
	__antithesis_instrumentation__.Notify(575360)
	peerGrouper := &partitionPeerGrouper{
		ctx:     ctx,
		evalCtx: evalCtx,
		rowCopy: make(rowenc.EncDatumRow, len(w.inputTypes)),
	}
	usage := memsize.RowsOverhead + memsize.RowsOverhead + memsize.DatumsOverhead*int64(len(w.windowFns))
	if err := w.growMemAccount(&w.acc, usage); err != nil {
		__antithesis_instrumentation__.Notify(575364)
		return err
	} else {
		__antithesis_instrumentation__.Notify(575365)
	}
	__antithesis_instrumentation__.Notify(575361)
	w.windowValues = append(w.windowValues, make([][]tree.Datum, len(w.windowFns)))

	partition.Sort(ctx)

	var prevWindowFn *windowFunc
	for _, windowFnIdx := range w.orderOfWindowFnsProcessing {
		__antithesis_instrumentation__.Notify(575366)
		windowFn := w.windowFns[windowFnIdx]

		frameRun := &tree.WindowFrameRun{
			ArgsIdxs:     windowFn.argsIdxs,
			FilterColIdx: windowFn.filterColIdx,
		}

		if windowFn.frame != nil {
			__antithesis_instrumentation__.Notify(575373)
			var err error
			if frameRun.Frame, err = windowFn.frame.ConvertToAST(); err != nil {
				__antithesis_instrumentation__.Notify(575377)
				return err
			} else {
				__antithesis_instrumentation__.Notify(575378)
			}
			__antithesis_instrumentation__.Notify(575374)
			startBound, endBound := windowFn.frame.Bounds.Start, windowFn.frame.Bounds.End
			if startBound.BoundType == execinfrapb.WindowerSpec_Frame_OFFSET_PRECEDING || func() bool {
				__antithesis_instrumentation__.Notify(575379)
				return startBound.BoundType == execinfrapb.WindowerSpec_Frame_OFFSET_FOLLOWING == true
			}() == true {
				__antithesis_instrumentation__.Notify(575380)
				switch windowFn.frame.Mode {
				case execinfrapb.WindowerSpec_Frame_ROWS:
					__antithesis_instrumentation__.Notify(575381)
					frameRun.StartBoundOffset = tree.NewDInt(tree.DInt(int(startBound.IntOffset)))
				case execinfrapb.WindowerSpec_Frame_RANGE:
					__antithesis_instrumentation__.Notify(575382)
					datum, err := execinfra.DecodeDatum(&w.datumAlloc, startBound.OffsetType.Type, startBound.TypedOffset)
					if err != nil {
						__antithesis_instrumentation__.Notify(575386)
						return err
					} else {
						__antithesis_instrumentation__.Notify(575387)
					}
					__antithesis_instrumentation__.Notify(575383)
					frameRun.StartBoundOffset = datum
				case execinfrapb.WindowerSpec_Frame_GROUPS:
					__antithesis_instrumentation__.Notify(575384)
					frameRun.StartBoundOffset = tree.NewDInt(tree.DInt(int(startBound.IntOffset)))
				default:
					__antithesis_instrumentation__.Notify(575385)
					return errors.AssertionFailedf(
						"unexpected WindowFrameMode: %d", errors.Safe(windowFn.frame.Mode))
				}
			} else {
				__antithesis_instrumentation__.Notify(575388)
			}
			__antithesis_instrumentation__.Notify(575375)
			if endBound != nil {
				__antithesis_instrumentation__.Notify(575389)
				if endBound.BoundType == execinfrapb.WindowerSpec_Frame_OFFSET_PRECEDING || func() bool {
					__antithesis_instrumentation__.Notify(575390)
					return endBound.BoundType == execinfrapb.WindowerSpec_Frame_OFFSET_FOLLOWING == true
				}() == true {
					__antithesis_instrumentation__.Notify(575391)
					switch windowFn.frame.Mode {
					case execinfrapb.WindowerSpec_Frame_ROWS:
						__antithesis_instrumentation__.Notify(575392)
						frameRun.EndBoundOffset = tree.NewDInt(tree.DInt(int(endBound.IntOffset)))
					case execinfrapb.WindowerSpec_Frame_RANGE:
						__antithesis_instrumentation__.Notify(575393)
						datum, err := execinfra.DecodeDatum(&w.datumAlloc, endBound.OffsetType.Type, endBound.TypedOffset)
						if err != nil {
							__antithesis_instrumentation__.Notify(575397)
							return err
						} else {
							__antithesis_instrumentation__.Notify(575398)
						}
						__antithesis_instrumentation__.Notify(575394)
						frameRun.EndBoundOffset = datum
					case execinfrapb.WindowerSpec_Frame_GROUPS:
						__antithesis_instrumentation__.Notify(575395)
						frameRun.EndBoundOffset = tree.NewDInt(tree.DInt(int(endBound.IntOffset)))
					default:
						__antithesis_instrumentation__.Notify(575396)
						return errors.AssertionFailedf("unexpected WindowFrameMode: %d",
							errors.Safe(windowFn.frame.Mode))
					}
				} else {
					__antithesis_instrumentation__.Notify(575399)
				}
			} else {
				__antithesis_instrumentation__.Notify(575400)
			}
			__antithesis_instrumentation__.Notify(575376)
			if frameRun.RangeModeWithOffsets() {
				__antithesis_instrumentation__.Notify(575401)
				ordCol := windowFn.ordering.Columns[0]
				frameRun.OrdColIdx = int(ordCol.ColIdx)

				frameRun.OrdDirection = encoding.Direction(ordCol.Direction + 1)

				colTyp := w.inputTypes[ordCol.ColIdx]

				offsetTyp := colTyp
				if types.IsDateTimeType(colTyp) {
					__antithesis_instrumentation__.Notify(575404)

					offsetTyp = types.Interval
				} else {
					__antithesis_instrumentation__.Notify(575405)
				}
				__antithesis_instrumentation__.Notify(575402)
				plusOp, minusOp, found := tree.WindowFrameRangeOps{}.LookupImpl(colTyp, offsetTyp)
				if !found {
					__antithesis_instrumentation__.Notify(575406)
					return pgerror.Newf(pgcode.Windowing,
						"given logical offset cannot be combined with ordering column")
				} else {
					__antithesis_instrumentation__.Notify(575407)
				}
				__antithesis_instrumentation__.Notify(575403)
				frameRun.PlusOp, frameRun.MinusOp = plusOp, minusOp
			} else {
				__antithesis_instrumentation__.Notify(575408)
			}
		} else {
			__antithesis_instrumentation__.Notify(575409)
		}
		__antithesis_instrumentation__.Notify(575367)

		builtin := w.builtins[windowFnIdx]
		builtin.Reset(ctx)

		usage = memsize.DatumsOverhead + memsize.DatumOverhead*int64(partition.Len())
		if err := w.growMemAccount(&w.acc, usage); err != nil {
			__antithesis_instrumentation__.Notify(575410)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575411)
		}
		__antithesis_instrumentation__.Notify(575368)
		w.windowValues[partitionIdx][windowFnIdx] = make([]tree.Datum, partition.Len())

		if len(windowFn.ordering.Columns) > 0 {
			__antithesis_instrumentation__.Notify(575412)

			if prevWindowFn != nil && func() bool {
				__antithesis_instrumentation__.Notify(575413)
				return !windowFn.ordering.Equal(prevWindowFn.ordering) == true
			}() == true {
				__antithesis_instrumentation__.Notify(575414)
				if err := partition.Reorder(ctx, execinfrapb.ConvertToColumnOrdering(windowFn.ordering)); err != nil {
					__antithesis_instrumentation__.Notify(575416)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575417)
				}
				__antithesis_instrumentation__.Notify(575415)
				partition.Sort(ctx)
			} else {
				__antithesis_instrumentation__.Notify(575418)
			}
		} else {
			__antithesis_instrumentation__.Notify(575419)
		}
		__antithesis_instrumentation__.Notify(575369)
		peerGrouper.ordering = windowFn.ordering
		peerGrouper.partition = partition

		frameRun.Rows = partition
		frameRun.RowIdx = 0

		if !frameRun.Frame.IsDefaultFrame() {
			__antithesis_instrumentation__.Notify(575420)

			builtins.ShouldReset(builtin)
		} else {
			__antithesis_instrumentation__.Notify(575421)
		}
		__antithesis_instrumentation__.Notify(575370)

		if err := frameRun.PeerHelper.Init(frameRun, peerGrouper); err != nil {
			__antithesis_instrumentation__.Notify(575422)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575423)
		}
		__antithesis_instrumentation__.Notify(575371)
		frameRun.CurRowPeerGroupNum = 0

		var prevRes tree.Datum
		for frameRun.RowIdx < partition.Len() {
			__antithesis_instrumentation__.Notify(575424)

			peerGroupEndIdx := frameRun.PeerHelper.GetFirstPeerIdx(frameRun.CurRowPeerGroupNum) + frameRun.PeerHelper.GetRowCount(frameRun.CurRowPeerGroupNum)
			for ; frameRun.RowIdx < peerGroupEndIdx; frameRun.RowIdx++ {
				__antithesis_instrumentation__.Notify(575427)
				if err := w.cancelChecker.Check(); err != nil {
					__antithesis_instrumentation__.Notify(575432)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575433)
				}
				__antithesis_instrumentation__.Notify(575428)
				res, err := builtin.Compute(ctx, evalCtx, frameRun)
				if err != nil {
					__antithesis_instrumentation__.Notify(575434)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575435)
				}
				__antithesis_instrumentation__.Notify(575429)
				row, err := frameRun.Rows.GetRow(ctx, frameRun.RowIdx)
				if err != nil {
					__antithesis_instrumentation__.Notify(575436)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575437)
				}
				__antithesis_instrumentation__.Notify(575430)
				if prevRes == nil || func() bool {
					__antithesis_instrumentation__.Notify(575438)
					return prevRes != res == true
				}() == true {
					__antithesis_instrumentation__.Notify(575439)

					if err := w.growMemAccount(&w.acc, int64(res.Size())-memsize.DatumOverhead); err != nil {
						__antithesis_instrumentation__.Notify(575440)
						return err
					} else {
						__antithesis_instrumentation__.Notify(575441)
					}
				} else {
					__antithesis_instrumentation__.Notify(575442)
				}
				__antithesis_instrumentation__.Notify(575431)
				w.windowValues[partitionIdx][windowFnIdx][row.GetIdx()] = res
				prevRes = res
			}
			__antithesis_instrumentation__.Notify(575425)
			if err := frameRun.PeerHelper.Update(frameRun); err != nil {
				__antithesis_instrumentation__.Notify(575443)
				return err
			} else {
				__antithesis_instrumentation__.Notify(575444)
			}
			__antithesis_instrumentation__.Notify(575426)
			frameRun.CurRowPeerGroupNum++
		}
		__antithesis_instrumentation__.Notify(575372)

		prevWindowFn = windowFn
	}
	__antithesis_instrumentation__.Notify(575362)

	if err := w.growMemAccount(&w.acc, memsize.Int); err != nil {
		__antithesis_instrumentation__.Notify(575445)
		return err
	} else {
		__antithesis_instrumentation__.Notify(575446)
	}
	__antithesis_instrumentation__.Notify(575363)
	w.partitionSizes = append(w.partitionSizes, w.partition.Len())
	return nil
}

func (w *windower) computeWindowFunctions(ctx context.Context, evalCtx *tree.EvalContext) error {
	__antithesis_instrumentation__.Notify(575447)
	w.findOrderOfWindowFnsToProcessIn()

	usage := memsize.IntSliceOverhead + memsize.RowsSliceOverhead
	if err := w.growMemAccount(&w.acc, usage); err != nil {
		__antithesis_instrumentation__.Notify(575451)
		return err
	} else {
		__antithesis_instrumentation__.Notify(575452)
	}
	__antithesis_instrumentation__.Notify(575448)
	w.partitionSizes = make([]int, 0, 8)
	w.windowValues = make([][][]tree.Datum, 0, 8)
	bucket := ""

	ordering := execinfrapb.ConvertToColumnOrdering(w.windowFns[w.orderOfWindowFnsProcessing[0]].ordering)
	w.partition = rowcontainer.NewDiskBackedIndexedRowContainer(
		ordering,
		w.inputTypes,
		w.EvalCtx,
		w.FlowCtx.Cfg.TempStorage,
		w.MemMonitor,
		w.diskMonitor,
	)
	i, err := w.allRowsPartitioned.NewAllRowsIterator(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(575453)
		return err
	} else {
		__antithesis_instrumentation__.Notify(575454)
	}
	__antithesis_instrumentation__.Notify(575449)
	defer i.Close()

	for i.Rewind(); ; i.Next() {
		__antithesis_instrumentation__.Notify(575455)
		if ok, err := i.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(575460)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575461)
			if !ok {
				__antithesis_instrumentation__.Notify(575462)
				break
			} else {
				__antithesis_instrumentation__.Notify(575463)
			}
		}
		__antithesis_instrumentation__.Notify(575456)
		row, err := i.Row()
		if err != nil {
			__antithesis_instrumentation__.Notify(575464)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575465)
		}
		__antithesis_instrumentation__.Notify(575457)
		if err := w.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(575466)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575467)
		}
		__antithesis_instrumentation__.Notify(575458)
		if len(w.partitionBy) > 0 {
			__antithesis_instrumentation__.Notify(575468)

			w.scratch = w.scratch[:0]
			for _, col := range w.partitionBy {
				__antithesis_instrumentation__.Notify(575470)
				if int(col) >= len(row) {
					__antithesis_instrumentation__.Notify(575472)
					return errors.AssertionFailedf(
						"hash column %d, row with only %d columns", errors.Safe(col), errors.Safe(len(row)))
				} else {
					__antithesis_instrumentation__.Notify(575473)
				}
				__antithesis_instrumentation__.Notify(575471)
				var err error

				w.scratch, err = row[col].Fingerprint(
					ctx, w.inputTypes[int(col)], &w.datumAlloc, w.scratch, &w.acc,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(575474)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575475)
				}
			}
			__antithesis_instrumentation__.Notify(575469)
			if string(w.scratch) != bucket {
				__antithesis_instrumentation__.Notify(575476)

				if bucket != "" {
					__antithesis_instrumentation__.Notify(575479)
					if err := w.processPartition(ctx, evalCtx, w.partition, len(w.partitionSizes)); err != nil {
						__antithesis_instrumentation__.Notify(575480)
						return err
					} else {
						__antithesis_instrumentation__.Notify(575481)
					}
				} else {
					__antithesis_instrumentation__.Notify(575482)
				}
				__antithesis_instrumentation__.Notify(575477)
				bucket = string(w.scratch)
				if err := w.partition.UnsafeReset(ctx); err != nil {
					__antithesis_instrumentation__.Notify(575483)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575484)
				}
				__antithesis_instrumentation__.Notify(575478)
				if !w.windowFns[w.orderOfWindowFnsProcessing[0]].ordering.Equal(w.windowFns[w.orderOfWindowFnsProcessing[len(w.windowFns)-1]].ordering) {
					__antithesis_instrumentation__.Notify(575485)

					if err = w.partition.Reorder(ctx, ordering); err != nil {
						__antithesis_instrumentation__.Notify(575486)
						return err
					} else {
						__antithesis_instrumentation__.Notify(575487)
					}
				} else {
					__antithesis_instrumentation__.Notify(575488)
				}
			} else {
				__antithesis_instrumentation__.Notify(575489)
			}
		} else {
			__antithesis_instrumentation__.Notify(575490)
		}
		__antithesis_instrumentation__.Notify(575459)
		if err := w.partition.AddRow(w.Ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(575491)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575492)
		}
	}
	__antithesis_instrumentation__.Notify(575450)
	return w.processPartition(ctx, evalCtx, w.partition, len(w.partitionSizes))
}

func (w *windower) populateNextOutputRow() (bool, error) {
	__antithesis_instrumentation__.Notify(575493)
	if w.partitionIdx < len(w.partitionSizes) {
		__antithesis_instrumentation__.Notify(575495)
		if w.allRowsIterator == nil {
			__antithesis_instrumentation__.Notify(575501)
			w.allRowsIterator = w.allRowsPartitioned.NewUnmarkedIterator(w.Ctx)
			w.allRowsIterator.Rewind()
		} else {
			__antithesis_instrumentation__.Notify(575502)
		}
		__antithesis_instrumentation__.Notify(575496)

		rowIdx := w.rowsInBucketEmitted
		if ok, err := w.allRowsIterator.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(575503)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(575504)
			if !ok {
				__antithesis_instrumentation__.Notify(575505)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(575506)
			}
		}
		__antithesis_instrumentation__.Notify(575497)
		inputRow, err := w.allRowsIterator.Row()
		w.allRowsIterator.Next()
		if err != nil {
			__antithesis_instrumentation__.Notify(575507)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(575508)
		}
		__antithesis_instrumentation__.Notify(575498)
		copy(w.outputRow, inputRow[:len(w.inputTypes)])
		for windowFnIdx, windowFn := range w.windowFns {
			__antithesis_instrumentation__.Notify(575509)
			windowFnRes := w.windowValues[w.partitionIdx][windowFnIdx][rowIdx]
			encWindowFnRes := rowenc.DatumToEncDatum(w.outputTypes[windowFn.outputColIdx], windowFnRes)
			w.outputRow[windowFn.outputColIdx] = encWindowFnRes
		}
		__antithesis_instrumentation__.Notify(575499)
		w.rowsInBucketEmitted++
		if w.rowsInBucketEmitted == w.partitionSizes[w.partitionIdx] {
			__antithesis_instrumentation__.Notify(575510)

			w.partitionIdx++
			w.rowsInBucketEmitted = 0
		} else {
			__antithesis_instrumentation__.Notify(575511)
		}
		__antithesis_instrumentation__.Notify(575500)
		return true, nil

	} else {
		__antithesis_instrumentation__.Notify(575512)
	}
	__antithesis_instrumentation__.Notify(575494)
	return false, nil
}

type windowFunc struct {
	ordering     execinfrapb.Ordering
	argsIdxs     []uint32
	frame        *execinfrapb.WindowerSpec_Frame
	filterColIdx int
	outputColIdx int
}

type partitionPeerGrouper struct {
	ctx       context.Context
	evalCtx   *tree.EvalContext
	partition *rowcontainer.DiskBackedIndexedRowContainer
	ordering  execinfrapb.Ordering
	rowCopy   rowenc.EncDatumRow
	err       error
}

func (n *partitionPeerGrouper) InSameGroup(i, j int) (bool, error) {
	__antithesis_instrumentation__.Notify(575513)
	if len(n.ordering.Columns) == 0 {
		__antithesis_instrumentation__.Notify(575519)

		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(575520)
	}
	__antithesis_instrumentation__.Notify(575514)
	if n.err != nil {
		__antithesis_instrumentation__.Notify(575521)
		return false, n.err
	} else {
		__antithesis_instrumentation__.Notify(575522)
	}
	__antithesis_instrumentation__.Notify(575515)
	indexedRow, err := n.partition.GetRow(n.ctx, i)
	if err != nil {
		__antithesis_instrumentation__.Notify(575523)
		n.err = err
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(575524)
	}
	__antithesis_instrumentation__.Notify(575516)
	row := indexedRow.(rowcontainer.IndexedRow)

	copy(n.rowCopy, row.Row)
	rb, err := n.partition.GetRow(n.ctx, j)
	if err != nil {
		__antithesis_instrumentation__.Notify(575525)
		n.err = err
		return false, n.err
	} else {
		__antithesis_instrumentation__.Notify(575526)
	}
	__antithesis_instrumentation__.Notify(575517)
	for _, o := range n.ordering.Columns {
		__antithesis_instrumentation__.Notify(575527)
		da := n.rowCopy[o.ColIdx].Datum
		db, err := rb.GetDatum(int(o.ColIdx))
		if err != nil {
			__antithesis_instrumentation__.Notify(575529)
			n.err = err
			return false, n.err
		} else {
			__antithesis_instrumentation__.Notify(575530)
		}
		__antithesis_instrumentation__.Notify(575528)
		if c := da.Compare(n.evalCtx, db); c != 0 {
			__antithesis_instrumentation__.Notify(575531)
			if o.Direction != execinfrapb.Ordering_Column_ASC {
				__antithesis_instrumentation__.Notify(575533)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(575534)
			}
			__antithesis_instrumentation__.Notify(575532)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(575535)
		}
	}
	__antithesis_instrumentation__.Notify(575518)
	return true, nil
}

func CreateWindowerSpecFunc(funcStr string) (execinfrapb.WindowerSpec_Func, error) {
	__antithesis_instrumentation__.Notify(575536)
	if aggBuiltin, err := execinfrapb.GetAggregateFuncIdx(funcStr); err == nil {
		__antithesis_instrumentation__.Notify(575537)
		aggSpec := execinfrapb.AggregatorSpec_Func(aggBuiltin)
		return execinfrapb.WindowerSpec_Func{AggregateFunc: &aggSpec}, nil
	} else {
		__antithesis_instrumentation__.Notify(575538)
		if winBuiltin, err := execinfrapb.GetWindowFuncIdx(funcStr); err == nil {
			__antithesis_instrumentation__.Notify(575539)
			winSpec := execinfrapb.WindowerSpec_WindowFunc(winBuiltin)
			return execinfrapb.WindowerSpec_Func{WindowFunc: &winSpec}, nil
		} else {
			__antithesis_instrumentation__.Notify(575540)
			return execinfrapb.WindowerSpec_Func{}, errors.Errorf("unknown aggregate/window function %s", funcStr)
		}
	}
}

func (w *windower) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(575541)
	is, ok := getInputStats(w.input)
	if !ok {
		__antithesis_instrumentation__.Notify(575543)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(575544)
	}
	__antithesis_instrumentation__.Notify(575542)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Exec: execinfrapb.ExecStats{
			MaxAllocatedMem:  optional.MakeUint(uint64(w.MemMonitor.MaximumBytes())),
			MaxAllocatedDisk: optional.MakeUint(uint64(w.diskMonitor.MaximumBytes())),
		},
		Output: w.OutputHelper.Stats(),
	}
}

func (w *windower) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(575545)
	if _, ok := w.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(575547)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(575548)
	}
	__antithesis_instrumentation__.Notify(575546)
	return 0
}

func (w *windower) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(575549)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(575551)
		if n, ok := w.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(575553)
			return n
		} else {
			__antithesis_instrumentation__.Notify(575554)
		}
		__antithesis_instrumentation__.Notify(575552)
		panic("input to windower is not an execinfra.OpNode")
	} else {
		__antithesis_instrumentation__.Notify(575555)
	}
	__antithesis_instrumentation__.Notify(575550)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
