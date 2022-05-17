package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

type sorterBase struct {
	execinfra.ProcessorBase

	input    execinfra.RowSource
	ordering colinfo.ColumnOrdering
	matchLen uint32

	rows rowcontainer.SortableRowContainer
	i    rowcontainer.RowIterator

	diskMonitor *mon.BytesMonitor
}

func (s *sorterBase) init(
	self execinfra.RowSource,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	processorName string,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
	ordering colinfo.ColumnOrdering,
	matchLen uint32,
	opts execinfra.ProcStateOpts,
) error {
	ctx := flowCtx.EvalCtx.Ctx()
	if execinfra.ShouldCollectStats(ctx, flowCtx) {
		input = newInputStatCollector(input)
		s.ExecStatsForTrace = s.execStatsForTrace
	}

	memMonitor := execinfra.NewLimitedMonitor(ctx, flowCtx.EvalCtx.Mon, flowCtx, fmt.Sprintf("%s-limited", processorName))
	if err := s.ProcessorBase.Init(
		self, post, input.OutputTypes(), flowCtx, processorID, output, memMonitor, opts,
	); err != nil {
		memMonitor.Stop(ctx)
		return err
	}

	s.diskMonitor = execinfra.NewMonitor(ctx, flowCtx.DiskMonitor, fmt.Sprintf("%s-disk", processorName))
	rc := rowcontainer.DiskBackedRowContainer{}
	rc.Init(
		ordering,
		input.OutputTypes(),
		s.EvalCtx,
		flowCtx.Cfg.TempStorage,
		memMonitor,
		s.diskMonitor,
	)
	s.rows = &rc

	s.input = input
	s.ordering = ordering
	s.matchLen = matchLen
	return nil
}

func (s *sorterBase) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(574829)
	for s.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(574831)
		if ok, err := s.i.Valid(); err != nil || func() bool {
			__antithesis_instrumentation__.Notify(574834)
			return !ok == true
		}() == true {
			__antithesis_instrumentation__.Notify(574835)
			s.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(574836)
		}
		__antithesis_instrumentation__.Notify(574832)

		row, err := s.i.Row()
		if err != nil {
			__antithesis_instrumentation__.Notify(574837)
			s.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(574838)
		}
		__antithesis_instrumentation__.Notify(574833)
		s.i.Next()

		if outRow := s.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(574839)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(574840)
		}
	}
	__antithesis_instrumentation__.Notify(574830)
	return nil, s.DrainHelper()
}

func (s *sorterBase) close() {
	__antithesis_instrumentation__.Notify(574841)

	if s.InternalClose() {
		__antithesis_instrumentation__.Notify(574842)
		if s.i != nil {
			__antithesis_instrumentation__.Notify(574844)
			s.i.Close()
		} else {
			__antithesis_instrumentation__.Notify(574845)
		}
		__antithesis_instrumentation__.Notify(574843)
		s.rows.Close(s.Ctx)
		s.MemMonitor.Stop(s.Ctx)
		if s.diskMonitor != nil {
			__antithesis_instrumentation__.Notify(574846)
			s.diskMonitor.Stop(s.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(574847)
		}
	} else {
		__antithesis_instrumentation__.Notify(574848)
	}
}

func (s *sorterBase) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(574849)
	is, ok := getInputStats(s.input)
	if !ok {
		__antithesis_instrumentation__.Notify(574851)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(574852)
	}
	__antithesis_instrumentation__.Notify(574850)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Exec: execinfrapb.ExecStats{
			MaxAllocatedMem:  optional.MakeUint(uint64(s.MemMonitor.MaximumBytes())),
			MaxAllocatedDisk: optional.MakeUint(uint64(s.diskMonitor.MaximumBytes())),
		},
		Output: s.OutputHelper.Stats(),
	}
}

func newSorter(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.SorterSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(574853)

	if spec.OrderingMatchLen == 0 {
		__antithesis_instrumentation__.Notify(574855)
		if spec.Limit == 0 {
			__antithesis_instrumentation__.Notify(574857)

			return newSortAllProcessor(ctx, flowCtx, processorID, spec, input, post, output)
		} else {
			__antithesis_instrumentation__.Notify(574858)
		}
		__antithesis_instrumentation__.Notify(574856)

		return newSortTopKProcessor(flowCtx, processorID, spec, input, post, output, uint64(spec.Limit))
	} else {
		__antithesis_instrumentation__.Notify(574859)
	}
	__antithesis_instrumentation__.Notify(574854)

	return newSortChunksProcessor(flowCtx, processorID, spec, input, post, output)
}

type sortAllProcessor struct {
	sorterBase
}

var _ execinfra.Processor = &sortAllProcessor{}
var _ execinfra.RowSource = &sortAllProcessor{}

const sortAllProcName = "sortAll"

func newSortAllProcessor(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.SorterSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	out execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(574860)
	proc := &sortAllProcessor{}
	if err := proc.sorterBase.init(
		proc, flowCtx, processorID, sortAllProcName, input, post, out,
		execinfrapb.ConvertToColumnOrdering(spec.OutputOrdering),
		spec.OrderingMatchLen,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(574862)
				proc.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(574863)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574864)
	}
	__antithesis_instrumentation__.Notify(574861)
	return proc, nil
}

func (s *sortAllProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574865)
	ctx = s.StartInternal(ctx, sortAllProcName)
	s.input.Start(ctx)

	valid, err := s.fill()
	if !valid || func() bool {
		__antithesis_instrumentation__.Notify(574866)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(574867)
		s.MoveToDraining(err)
	} else {
		__antithesis_instrumentation__.Notify(574868)
	}
}

func (s *sortAllProcessor) fill() (ok bool, _ error) {
	__antithesis_instrumentation__.Notify(574869)
	ctx := s.EvalCtx.Ctx()

	for {
		__antithesis_instrumentation__.Notify(574871)
		row, meta := s.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(574874)
			s.AppendTrailingMeta(*meta)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(574876)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(574877)
			}
			__antithesis_instrumentation__.Notify(574875)
			continue
		} else {
			__antithesis_instrumentation__.Notify(574878)
		}
		__antithesis_instrumentation__.Notify(574872)
		if row == nil {
			__antithesis_instrumentation__.Notify(574879)
			break
		} else {
			__antithesis_instrumentation__.Notify(574880)
		}
		__antithesis_instrumentation__.Notify(574873)

		if err := s.rows.AddRow(ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(574881)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(574882)
		}
	}
	__antithesis_instrumentation__.Notify(574870)
	s.rows.Sort(ctx)

	s.i = s.rows.NewFinalIterator(ctx)
	s.i.Rewind()
	return true, nil
}

func (s *sortAllProcessor) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(574883)

	s.close()
}

type sortTopKProcessor struct {
	sorterBase
	k uint64
}

var _ execinfra.Processor = &sortTopKProcessor{}
var _ execinfra.RowSource = &sortTopKProcessor{}

const sortTopKProcName = "sortTopK"

var errSortTopKZeroK = errors.New("invalid value 0 for k")

func newSortTopKProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.SorterSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	out execinfra.RowReceiver,
	k uint64,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(574884)
	if k == 0 {
		__antithesis_instrumentation__.Notify(574887)
		return nil, errors.NewAssertionErrorWithWrappedErrf(errSortTopKZeroK,
			"error creating top k sorter")
	} else {
		__antithesis_instrumentation__.Notify(574888)
	}
	__antithesis_instrumentation__.Notify(574885)
	ordering := execinfrapb.ConvertToColumnOrdering(spec.OutputOrdering)
	proc := &sortTopKProcessor{k: k}
	if err := proc.sorterBase.init(
		proc, flowCtx, processorID, sortTopKProcName, input, post, out,
		ordering, spec.OrderingMatchLen,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(574889)
				proc.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(574890)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574891)
	}
	__antithesis_instrumentation__.Notify(574886)
	return proc, nil
}

func (s *sortTopKProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574892)
	ctx = s.StartInternal(ctx, sortTopKProcName)
	s.input.Start(ctx)

	heapCreated := false
	for {
		__antithesis_instrumentation__.Notify(574894)
		row, meta := s.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(574897)
			s.AppendTrailingMeta(*meta)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(574899)
				s.MoveToDraining(nil)
				break
			} else {
				__antithesis_instrumentation__.Notify(574900)
			}
			__antithesis_instrumentation__.Notify(574898)
			continue
		} else {
			__antithesis_instrumentation__.Notify(574901)
		}
		__antithesis_instrumentation__.Notify(574895)
		if row == nil {
			__antithesis_instrumentation__.Notify(574902)
			break
		} else {
			__antithesis_instrumentation__.Notify(574903)
		}
		__antithesis_instrumentation__.Notify(574896)

		if uint64(s.rows.Len()) < s.k {
			__antithesis_instrumentation__.Notify(574904)

			if err := s.rows.AddRow(ctx, row); err != nil {
				__antithesis_instrumentation__.Notify(574905)
				s.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(574906)
			}
		} else {
			__antithesis_instrumentation__.Notify(574907)
			if !heapCreated {
				__antithesis_instrumentation__.Notify(574909)

				s.rows.InitTopK()
				heapCreated = true
			} else {
				__antithesis_instrumentation__.Notify(574910)
			}
			__antithesis_instrumentation__.Notify(574908)

			if err := s.rows.MaybeReplaceMax(ctx, row); err != nil {
				__antithesis_instrumentation__.Notify(574911)
				s.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(574912)
			}
		}
	}
	__antithesis_instrumentation__.Notify(574893)
	s.rows.Sort(ctx)
	s.i = s.rows.NewFinalIterator(ctx)
	s.i.Rewind()
}

func (s *sortTopKProcessor) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(574913)

	s.close()
}

type sortChunksProcessor struct {
	sorterBase

	alloc tree.DatumAlloc

	nextChunkRow rowenc.EncDatumRow
}

var _ execinfra.Processor = &sortChunksProcessor{}
var _ execinfra.RowSource = &sortChunksProcessor{}

const sortChunksProcName = "sortChunks"

func newSortChunksProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.SorterSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	out execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(574914)
	ordering := execinfrapb.ConvertToColumnOrdering(spec.OutputOrdering)

	proc := &sortChunksProcessor{}
	if err := proc.sorterBase.init(
		proc, flowCtx, processorID, sortChunksProcName, input, post, out, ordering, spec.OrderingMatchLen,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(574916)
				proc.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(574917)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574918)
	}
	__antithesis_instrumentation__.Notify(574915)
	proc.i = proc.rows.NewFinalIterator(proc.Ctx)
	return proc, nil
}

func (s *sortChunksProcessor) chunkCompleted(
	nextChunkRow, prefix rowenc.EncDatumRow,
) (bool, error) {
	__antithesis_instrumentation__.Notify(574919)
	types := s.input.OutputTypes()
	for _, ord := range s.ordering[:s.matchLen] {
		__antithesis_instrumentation__.Notify(574921)
		col := ord.ColIdx
		cmp, err := nextChunkRow[col].Compare(types[col], &s.alloc, s.EvalCtx, &prefix[col])
		if cmp != 0 || func() bool {
			__antithesis_instrumentation__.Notify(574922)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(574923)
			return true, err
		} else {
			__antithesis_instrumentation__.Notify(574924)
		}
	}
	__antithesis_instrumentation__.Notify(574920)
	return false, nil
}

func (s *sortChunksProcessor) fill() (bool, error) {
	__antithesis_instrumentation__.Notify(574925)
	ctx := s.Ctx

	var meta *execinfrapb.ProducerMetadata

	nextChunkRow := s.nextChunkRow
	s.nextChunkRow = nil
	for nextChunkRow == nil {
		__antithesis_instrumentation__.Notify(574929)
		nextChunkRow, meta = s.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(574931)
			s.AppendTrailingMeta(*meta)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(574933)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(574934)
			}
			__antithesis_instrumentation__.Notify(574932)
			continue
		} else {
			__antithesis_instrumentation__.Notify(574935)
			if nextChunkRow == nil {
				__antithesis_instrumentation__.Notify(574936)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(574937)
			}
		}
		__antithesis_instrumentation__.Notify(574930)
		break
	}
	__antithesis_instrumentation__.Notify(574926)
	prefix := nextChunkRow

	if err := s.rows.AddRow(ctx, nextChunkRow); err != nil {
		__antithesis_instrumentation__.Notify(574938)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(574939)
	}
	__antithesis_instrumentation__.Notify(574927)

	for {
		__antithesis_instrumentation__.Notify(574940)
		nextChunkRow, meta = s.input.Next()

		if meta != nil {
			__antithesis_instrumentation__.Notify(574945)
			s.AppendTrailingMeta(*meta)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(574947)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(574948)
			}
			__antithesis_instrumentation__.Notify(574946)
			continue
		} else {
			__antithesis_instrumentation__.Notify(574949)
		}
		__antithesis_instrumentation__.Notify(574941)
		if nextChunkRow == nil {
			__antithesis_instrumentation__.Notify(574950)
			break
		} else {
			__antithesis_instrumentation__.Notify(574951)
		}
		__antithesis_instrumentation__.Notify(574942)

		chunkCompleted, err := s.chunkCompleted(nextChunkRow, prefix)

		if err != nil {
			__antithesis_instrumentation__.Notify(574952)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(574953)
		}
		__antithesis_instrumentation__.Notify(574943)
		if chunkCompleted {
			__antithesis_instrumentation__.Notify(574954)
			s.nextChunkRow = nextChunkRow
			break
		} else {
			__antithesis_instrumentation__.Notify(574955)
		}
		__antithesis_instrumentation__.Notify(574944)

		if err := s.rows.AddRow(ctx, nextChunkRow); err != nil {
			__antithesis_instrumentation__.Notify(574956)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(574957)
		}
	}
	__antithesis_instrumentation__.Notify(574928)

	s.rows.Sort(ctx)

	return true, nil
}

func (s *sortChunksProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574958)
	ctx = s.StartInternal(ctx, sortChunksProcName)
	s.input.Start(ctx)
}

func (s *sortChunksProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(574959)
	ctx := s.Ctx
	for s.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(574961)
		ok, err := s.i.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(574965)
			s.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(574966)
		}
		__antithesis_instrumentation__.Notify(574962)

		if !ok {
			__antithesis_instrumentation__.Notify(574967)
			if err := s.rows.UnsafeReset(ctx); err != nil {
				__antithesis_instrumentation__.Notify(574970)
				s.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(574971)
			}
			__antithesis_instrumentation__.Notify(574968)
			valid, err := s.fill()
			if !valid || func() bool {
				__antithesis_instrumentation__.Notify(574972)
				return err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(574973)
				s.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(574974)
			}
			__antithesis_instrumentation__.Notify(574969)
			s.i.Close()
			s.i = s.rows.NewFinalIterator(ctx)
			s.i.Rewind()
			if ok, err := s.i.Valid(); err != nil || func() bool {
				__antithesis_instrumentation__.Notify(574975)
				return !ok == true
			}() == true {
				__antithesis_instrumentation__.Notify(574976)
				s.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(574977)
			}
		} else {
			__antithesis_instrumentation__.Notify(574978)
		}
		__antithesis_instrumentation__.Notify(574963)

		row, err := s.i.Row()
		if err != nil {
			__antithesis_instrumentation__.Notify(574979)
			s.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(574980)
		}
		__antithesis_instrumentation__.Notify(574964)
		s.i.Next()

		if outRow := s.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(574981)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(574982)
		}
	}
	__antithesis_instrumentation__.Notify(574960)
	return nil, s.DrainHelper()
}

func (s *sortChunksProcessor) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(574983)

	s.close()
}
