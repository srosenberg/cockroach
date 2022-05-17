package colbuilder

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"reflect"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/coldataext"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/colconv"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecagg"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecargs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecbase"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecjoin"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecproj"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecsel"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecutils"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecwindow"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colfetcher"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

func checkNumIn(inputs []colexecargs.OpWithMetaInfo, numIn int) error {
	__antithesis_instrumentation__.Notify(275683)
	if len(inputs) != numIn {
		__antithesis_instrumentation__.Notify(275685)
		return errors.Errorf("expected %d input(s), got %d", numIn, len(inputs))
	} else {
		__antithesis_instrumentation__.Notify(275686)
	}
	__antithesis_instrumentation__.Notify(275684)
	return nil
}

func wrapRowSources(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	inputs []colexecargs.OpWithMetaInfo,
	inputTypes [][]*types.T,
	streamingMemAccount *mon.BoundAccount,
	processorID int32,
	newToWrap func([]execinfra.RowSource) (execinfra.RowSource, error),
	materializerSafeToRelease bool,
	factory coldata.ColumnFactory,
) (*colexec.Columnarizer, []execinfra.Releasable, error) {
	__antithesis_instrumentation__.Notify(275687)
	var toWrapInputs []execinfra.RowSource
	var releasables []execinfra.Releasable
	for i := range inputs {
		__antithesis_instrumentation__.Notify(275692)

		if c, ok := inputs[i].Root.(*colexec.Columnarizer); ok {
			__antithesis_instrumentation__.Notify(275693)

			c.MarkAsRemovedFromFlow()
			toWrapInputs = append(toWrapInputs, c.Input())
		} else {
			__antithesis_instrumentation__.Notify(275694)
			toWrapInput := colexec.NewMaterializer(
				flowCtx,
				processorID,
				inputs[i],
				inputTypes[i],
			)

			inputs[i].StatsCollectors = nil
			inputs[i].MetadataSources = nil
			inputs[i].ToClose = nil
			toWrapInputs = append(toWrapInputs, toWrapInput)
			if materializerSafeToRelease {
				__antithesis_instrumentation__.Notify(275695)
				releasables = append(releasables, toWrapInput)
			} else {
				__antithesis_instrumentation__.Notify(275696)
			}
		}
	}
	__antithesis_instrumentation__.Notify(275688)

	toWrap, err := newToWrap(toWrapInputs)
	if err != nil {
		__antithesis_instrumentation__.Notify(275697)
		return nil, releasables, err
	} else {
		__antithesis_instrumentation__.Notify(275698)
	}
	__antithesis_instrumentation__.Notify(275689)

	proc, isProcessor := toWrap.(execinfra.Processor)
	if !isProcessor {
		__antithesis_instrumentation__.Notify(275699)
		return nil, nil, errors.AssertionFailedf("unexpectedly %T is not an execinfra.Processor", toWrap)
	} else {
		__antithesis_instrumentation__.Notify(275700)
	}
	__antithesis_instrumentation__.Notify(275690)
	var c *colexec.Columnarizer
	if proc.MustBeStreaming() {
		__antithesis_instrumentation__.Notify(275701)
		c = colexec.NewStreamingColumnarizer(
			colmem.NewAllocator(ctx, streamingMemAccount, factory), flowCtx, processorID, toWrap,
		)
	} else {
		__antithesis_instrumentation__.Notify(275702)
		c = colexec.NewBufferingColumnarizer(
			colmem.NewAllocator(ctx, streamingMemAccount, factory), flowCtx, processorID, toWrap,
		)
	}
	__antithesis_instrumentation__.Notify(275691)
	return c, releasables, nil
}

type opResult struct {
	*colexecargs.NewColOperatorResult
}

func needHashAggregator(aggSpec *execinfrapb.AggregatorSpec) (bool, error) {
	__antithesis_instrumentation__.Notify(275703)
	var groupCols, orderedCols util.FastIntSet
	for _, col := range aggSpec.OrderedGroupCols {
		__antithesis_instrumentation__.Notify(275707)
		orderedCols.Add(int(col))
	}
	__antithesis_instrumentation__.Notify(275704)
	for _, col := range aggSpec.GroupCols {
		__antithesis_instrumentation__.Notify(275708)
		if !orderedCols.Contains(int(col)) {
			__antithesis_instrumentation__.Notify(275710)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(275711)
		}
		__antithesis_instrumentation__.Notify(275709)
		groupCols.Add(int(col))
	}
	__antithesis_instrumentation__.Notify(275705)
	if !orderedCols.SubsetOf(groupCols) {
		__antithesis_instrumentation__.Notify(275712)
		return false, errors.AssertionFailedf("ordered cols must be a subset of grouping cols")
	} else {
		__antithesis_instrumentation__.Notify(275713)
	}
	__antithesis_instrumentation__.Notify(275706)
	return false, nil
}

func IsSupported(mode sessiondatapb.VectorizeExecMode, spec *execinfrapb.ProcessorSpec) error {
	__antithesis_instrumentation__.Notify(275714)
	err := supportedNatively(spec)
	if err != nil {
		__antithesis_instrumentation__.Notify(275716)
		if wrapErr := canWrap(mode, spec); wrapErr == nil {
			__antithesis_instrumentation__.Notify(275717)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(275718)
		}
	} else {
		__antithesis_instrumentation__.Notify(275719)
	}
	__antithesis_instrumentation__.Notify(275715)
	return err
}

func supportedNatively(spec *execinfrapb.ProcessorSpec) error {
	__antithesis_instrumentation__.Notify(275720)
	switch {
	case spec.Core.Noop != nil:
		__antithesis_instrumentation__.Notify(275721)
		return nil

	case spec.Core.Values != nil:
		__antithesis_instrumentation__.Notify(275722)
		return nil

	case spec.Core.TableReader != nil:
		__antithesis_instrumentation__.Notify(275723)
		return nil

	case spec.Core.JoinReader != nil:
		__antithesis_instrumentation__.Notify(275724)
		if !spec.Core.JoinReader.IsIndexJoin() {
			__antithesis_instrumentation__.Notify(275740)
			return errLookupJoinUnsupported
		} else {
			__antithesis_instrumentation__.Notify(275741)
		}
		__antithesis_instrumentation__.Notify(275725)
		return nil

	case spec.Core.Filterer != nil:
		__antithesis_instrumentation__.Notify(275726)
		return nil

	case spec.Core.Aggregator != nil:
		__antithesis_instrumentation__.Notify(275727)
		for _, agg := range spec.Core.Aggregator.Aggregations {
			__antithesis_instrumentation__.Notify(275742)
			if agg.FilterColIdx != nil {
				__antithesis_instrumentation__.Notify(275743)
				return errors.Newf("filtering aggregation not supported")
			} else {
				__antithesis_instrumentation__.Notify(275744)
			}
		}
		__antithesis_instrumentation__.Notify(275728)
		return nil

	case spec.Core.Distinct != nil:
		__antithesis_instrumentation__.Notify(275729)
		return nil

	case spec.Core.Ordinality != nil:
		__antithesis_instrumentation__.Notify(275730)
		return nil

	case spec.Core.HashJoiner != nil:
		__antithesis_instrumentation__.Notify(275731)
		if !spec.Core.HashJoiner.OnExpr.Empty() && func() bool {
			__antithesis_instrumentation__.Notify(275745)
			return spec.Core.HashJoiner.Type != descpb.InnerJoin == true
		}() == true {
			__antithesis_instrumentation__.Notify(275746)
			return errors.Newf("can't plan vectorized non-inner hash joins with ON expressions")
		} else {
			__antithesis_instrumentation__.Notify(275747)
		}
		__antithesis_instrumentation__.Notify(275732)
		return nil

	case spec.Core.MergeJoiner != nil:
		__antithesis_instrumentation__.Notify(275733)
		if !spec.Core.MergeJoiner.OnExpr.Empty() && func() bool {
			__antithesis_instrumentation__.Notify(275748)
			return spec.Core.MergeJoiner.Type != descpb.InnerJoin == true
		}() == true {
			__antithesis_instrumentation__.Notify(275749)
			return errors.Errorf("can't plan non-inner merge join with ON expressions")
		} else {
			__antithesis_instrumentation__.Notify(275750)
		}
		__antithesis_instrumentation__.Notify(275734)
		return nil

	case spec.Core.Sorter != nil:
		__antithesis_instrumentation__.Notify(275735)
		return nil

	case spec.Core.Windower != nil:
		__antithesis_instrumentation__.Notify(275736)
		for _, wf := range spec.Core.Windower.WindowFns {
			__antithesis_instrumentation__.Notify(275751)
			if wf.FilterColIdx != tree.NoColumnIdx {
				__antithesis_instrumentation__.Notify(275753)
				return errors.Newf("window functions with FILTER clause are not supported")
			} else {
				__antithesis_instrumentation__.Notify(275754)
			}
			__antithesis_instrumentation__.Notify(275752)
			if wf.Func.AggregateFunc != nil {
				__antithesis_instrumentation__.Notify(275755)
				if !colexecagg.IsAggOptimized(*wf.Func.AggregateFunc) {
					__antithesis_instrumentation__.Notify(275756)
					return errors.Newf("default aggregate window functions not supported")
				} else {
					__antithesis_instrumentation__.Notify(275757)
				}
			} else {
				__antithesis_instrumentation__.Notify(275758)
			}
		}
		__antithesis_instrumentation__.Notify(275737)
		return nil

	case spec.Core.LocalPlanNode != nil:
		__antithesis_instrumentation__.Notify(275738)

		return errLocalPlanNodeWrap

	default:
		__antithesis_instrumentation__.Notify(275739)
		return errCoreUnsupportedNatively
	}
}

var (
	errCoreUnsupportedNatively        = errors.New("unsupported processor core")
	errLocalPlanNodeWrap              = errors.New("LocalPlanNode core needs to be wrapped")
	errMetadataTestSenderWrap         = errors.New("core.MetadataTestSender is not supported")
	errMetadataTestReceiverWrap       = errors.New("core.MetadataTestReceiver is not supported")
	errChangeAggregatorWrap           = errors.New("core.ChangeAggregator is not supported")
	errChangeFrontierWrap             = errors.New("core.ChangeFrontier is not supported")
	errReadImportWrap                 = errors.New("core.ReadImport is not supported")
	errBackupDataWrap                 = errors.New("core.BackupData is not supported")
	errBackfillerWrap                 = errors.New("core.Backfiller is not supported (not an execinfra.RowSource)")
	errExporterWrap                   = errors.New("core.Exporter is not supported (not an execinfra.RowSource)")
	errSamplerWrap                    = errors.New("core.Sampler is not supported (not an execinfra.RowSource)")
	errSampleAggregatorWrap           = errors.New("core.SampleAggregator is not supported (not an execinfra.RowSource)")
	errExperimentalWrappingProhibited = errors.New("wrapping for non-JoinReader and non-LocalPlanNode cores is prohibited in vectorize=experimental_always")
	errWrappedCast                    = errors.New("mismatched types in NewColOperator and unsupported casts")
	errLookupJoinUnsupported          = errors.New("lookup join reader is unsupported in vectorized")
)

func canWrap(mode sessiondatapb.VectorizeExecMode, spec *execinfrapb.ProcessorSpec) error {
	__antithesis_instrumentation__.Notify(275759)
	if mode == sessiondatapb.VectorizeExperimentalAlways && func() bool {
		__antithesis_instrumentation__.Notify(275762)
		return spec.Core.JoinReader == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(275763)
		return spec.Core.LocalPlanNode == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(275764)
		return errExperimentalWrappingProhibited
	} else {
		__antithesis_instrumentation__.Notify(275765)
	}
	__antithesis_instrumentation__.Notify(275760)
	switch {
	case spec.Core.Noop != nil:
		__antithesis_instrumentation__.Notify(275766)
	case spec.Core.TableReader != nil:
		__antithesis_instrumentation__.Notify(275767)
	case spec.Core.JoinReader != nil:
		__antithesis_instrumentation__.Notify(275768)
	case spec.Core.Sorter != nil:
		__antithesis_instrumentation__.Notify(275769)
	case spec.Core.Aggregator != nil:
		__antithesis_instrumentation__.Notify(275770)
	case spec.Core.Distinct != nil:
		__antithesis_instrumentation__.Notify(275771)
	case spec.Core.MergeJoiner != nil:
		__antithesis_instrumentation__.Notify(275772)
	case spec.Core.HashJoiner != nil:
		__antithesis_instrumentation__.Notify(275773)
	case spec.Core.Values != nil:
		__antithesis_instrumentation__.Notify(275774)
	case spec.Core.Backfiller != nil:
		__antithesis_instrumentation__.Notify(275775)
		return errBackfillerWrap
	case spec.Core.ReadImport != nil:
		__antithesis_instrumentation__.Notify(275776)
		return errReadImportWrap
	case spec.Core.Exporter != nil:
		__antithesis_instrumentation__.Notify(275777)
		return errExporterWrap
	case spec.Core.Sampler != nil:
		__antithesis_instrumentation__.Notify(275778)
		return errSamplerWrap
	case spec.Core.SampleAggregator != nil:
		__antithesis_instrumentation__.Notify(275779)
		return errSampleAggregatorWrap
	case spec.Core.MetadataTestSender != nil:
		__antithesis_instrumentation__.Notify(275780)

		return errMetadataTestSenderWrap
	case spec.Core.MetadataTestReceiver != nil:
		__antithesis_instrumentation__.Notify(275781)

		return errMetadataTestReceiverWrap
	case spec.Core.ZigzagJoiner != nil:
		__antithesis_instrumentation__.Notify(275782)
	case spec.Core.ProjectSet != nil:
		__antithesis_instrumentation__.Notify(275783)
	case spec.Core.Windower != nil:
		__antithesis_instrumentation__.Notify(275784)
	case spec.Core.LocalPlanNode != nil:
		__antithesis_instrumentation__.Notify(275785)
	case spec.Core.ChangeAggregator != nil:
		__antithesis_instrumentation__.Notify(275786)

		return errChangeAggregatorWrap
	case spec.Core.ChangeFrontier != nil:
		__antithesis_instrumentation__.Notify(275787)

		return errChangeFrontierWrap
	case spec.Core.Ordinality != nil:
		__antithesis_instrumentation__.Notify(275788)
	case spec.Core.BulkRowWriter != nil:
		__antithesis_instrumentation__.Notify(275789)
	case spec.Core.InvertedFilterer != nil:
		__antithesis_instrumentation__.Notify(275790)
	case spec.Core.InvertedJoiner != nil:
		__antithesis_instrumentation__.Notify(275791)
	case spec.Core.BackupData != nil:
		__antithesis_instrumentation__.Notify(275792)
		return errBackupDataWrap
	case spec.Core.SplitAndScatter != nil:
		__antithesis_instrumentation__.Notify(275793)
	case spec.Core.RestoreData != nil:
		__antithesis_instrumentation__.Notify(275794)
	case spec.Core.Filterer != nil:
		__antithesis_instrumentation__.Notify(275795)
	case spec.Core.StreamIngestionData != nil:
		__antithesis_instrumentation__.Notify(275796)
	case spec.Core.StreamIngestionFrontier != nil:
		__antithesis_instrumentation__.Notify(275797)
	default:
		__antithesis_instrumentation__.Notify(275798)
		return errors.AssertionFailedf("unexpected processor core %q", spec.Core)
	}
	__antithesis_instrumentation__.Notify(275761)
	return nil
}

func (r opResult) createDiskBackedSort(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	args *colexecargs.NewColOperatorArgs,
	input colexecop.Operator,
	inputTypes []*types.T,
	ordering execinfrapb.Ordering,
	limit int64,
	matchLen uint32,
	maxNumberPartitions int,
	processorID int32,
	opNamePrefix string,
	factory coldata.ColumnFactory,
) colexecop.Operator {
	__antithesis_instrumentation__.Notify(275799)
	var (
		sorterMemMonitorName string
		inMemorySorter       colexecop.Operator
	)
	if len(ordering.Columns) == int(matchLen) {
		__antithesis_instrumentation__.Notify(275804)

		return input
	} else {
		__antithesis_instrumentation__.Notify(275805)
	}
	__antithesis_instrumentation__.Notify(275800)
	totalMemLimit := execinfra.GetWorkMemLimit(flowCtx)
	spoolMemLimit := totalMemLimit * 4 / 5
	maxOutputBatchMemSize := totalMemLimit - spoolMemLimit
	if totalMemLimit == 1 {
		__antithesis_instrumentation__.Notify(275806)

		spoolMemLimit = 1
		maxOutputBatchMemSize = 1
	} else {
		__antithesis_instrumentation__.Notify(275807)
	}
	__antithesis_instrumentation__.Notify(275801)
	if limit != 0 {
		__antithesis_instrumentation__.Notify(275808)

		var topKSorterMemAccount *mon.BoundAccount
		topKSorterMemAccount, sorterMemMonitorName = args.MonitorRegistry.CreateMemAccountForSpillStrategyWithLimit(
			ctx, flowCtx, spoolMemLimit, opNamePrefix+"topk-sort", processorID,
		)
		inMemorySorter = colexec.NewTopKSorter(
			colmem.NewAllocator(ctx, topKSorterMemAccount, factory), input,
			inputTypes, ordering.Columns, int(matchLen), uint64(limit), maxOutputBatchMemSize,
		)
	} else {
		__antithesis_instrumentation__.Notify(275809)
		if matchLen > 0 {
			__antithesis_instrumentation__.Notify(275810)

			opName := opNamePrefix + "sort-chunks"
			deselectorUnlimitedAllocator := colmem.NewAllocator(
				ctx, args.MonitorRegistry.CreateUnlimitedMemAccount(
					ctx, flowCtx, opName, processorID,
				), factory,
			)
			var sortChunksMemAccount *mon.BoundAccount
			sortChunksMemAccount, sorterMemMonitorName = args.MonitorRegistry.CreateMemAccountForSpillStrategyWithLimit(
				ctx, flowCtx, spoolMemLimit, opName, processorID,
			)
			inMemorySorter = colexec.NewSortChunks(
				deselectorUnlimitedAllocator, colmem.NewAllocator(ctx, sortChunksMemAccount, factory),
				input, inputTypes, ordering.Columns, int(matchLen), maxOutputBatchMemSize,
			)
		} else {
			__antithesis_instrumentation__.Notify(275811)

			var sorterMemAccount *mon.BoundAccount
			sorterMemAccount, sorterMemMonitorName = args.MonitorRegistry.CreateMemAccountForSpillStrategyWithLimit(
				ctx, flowCtx, spoolMemLimit, opNamePrefix+"sort-all", processorID,
			)
			inMemorySorter = colexec.NewSorter(
				colmem.NewAllocator(ctx, sorterMemAccount, factory), input,
				inputTypes, ordering.Columns, maxOutputBatchMemSize,
			)
		}
	}
	__antithesis_instrumentation__.Notify(275802)
	if args.TestingKnobs.DiskSpillingDisabled {
		__antithesis_instrumentation__.Notify(275812)

		return inMemorySorter
	} else {
		__antithesis_instrumentation__.Notify(275813)
	}
	__antithesis_instrumentation__.Notify(275803)

	return colexec.NewOneInputDiskSpiller(
		input, inMemorySorter.(colexecop.BufferingInMemoryOperator),
		sorterMemMonitorName,
		func(input colexecop.Operator) colexecop.Operator {
			__antithesis_instrumentation__.Notify(275814)
			opName := opNamePrefix + "external-sorter"

			sortUnlimitedAllocator := colmem.NewAllocator(
				ctx, args.MonitorRegistry.CreateUnlimitedMemAccount(
					ctx, flowCtx, opName+"-sort", processorID,
				), factory)
			mergeUnlimitedAllocator := colmem.NewAllocator(
				ctx, args.MonitorRegistry.CreateUnlimitedMemAccount(
					ctx, flowCtx, opName+"-merge", processorID,
				), factory)
			outputUnlimitedAllocator := colmem.NewAllocator(
				ctx, args.MonitorRegistry.CreateUnlimitedMemAccount(
					ctx, flowCtx, opName+"-output", processorID,
				), factory)
			diskAccount := args.MonitorRegistry.CreateDiskAccount(ctx, flowCtx, opName, processorID)
			es := colexec.NewExternalSorter(
				sortUnlimitedAllocator,
				mergeUnlimitedAllocator,
				outputUnlimitedAllocator,
				input, inputTypes, ordering, uint64(limit),
				int(matchLen),
				execinfra.GetWorkMemLimit(flowCtx),
				maxNumberPartitions,
				args.TestingKnobs.NumForcedRepartitions,
				args.TestingKnobs.DelegateFDAcquisitions,
				args.DiskQueueCfg,
				args.FDSemaphore,
				diskAccount,
			)
			r.ToClose = append(r.ToClose, es.(colexecop.Closer))
			return es
		},
		args.TestingKnobs.SpillingCallbackFn,
	)
}

func (r opResult) makeDiskBackedSorterConstructor(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	args *colexecargs.NewColOperatorArgs,
	opNamePrefix string,
	factory coldata.ColumnFactory,
) colexec.DiskBackedSorterConstructor {
	__antithesis_instrumentation__.Notify(275815)
	return func(input colexecop.Operator, inputTypes []*types.T, orderingCols []execinfrapb.Ordering_Column, maxNumberPartitions int) colexecop.Operator {
		__antithesis_instrumentation__.Notify(275816)
		if maxNumberPartitions < colexecop.ExternalSorterMinPartitions {
			__antithesis_instrumentation__.Notify(275819)
			colexecerror.InternalError(errors.AssertionFailedf(
				"external sorter is attempted to be created with %d partitions, minimum %d required",
				maxNumberPartitions, colexecop.ExternalSorterMinPartitions,
			))
		} else {
			__antithesis_instrumentation__.Notify(275820)
		}
		__antithesis_instrumentation__.Notify(275817)
		sortArgs := *args
		if !args.TestingKnobs.DelegateFDAcquisitions {
			__antithesis_instrumentation__.Notify(275821)

			sortArgs.FDSemaphore = nil
		} else {
			__antithesis_instrumentation__.Notify(275822)
		}
		__antithesis_instrumentation__.Notify(275818)
		return r.createDiskBackedSort(
			ctx, flowCtx, &sortArgs, input, inputTypes,
			execinfrapb.Ordering{Columns: orderingCols}, 0,
			0, maxNumberPartitions, args.Spec.ProcessorID,
			opNamePrefix+"-", factory,
		)
	}
}

func takeOverMetaInfo(target *colexecargs.OpWithMetaInfo, inputs []colexecargs.OpWithMetaInfo) {
	__antithesis_instrumentation__.Notify(275823)
	for i := range inputs {
		__antithesis_instrumentation__.Notify(275824)
		target.StatsCollectors = append(target.StatsCollectors, inputs[i].StatsCollectors...)
		target.MetadataSources = append(target.MetadataSources, inputs[i].MetadataSources...)
		target.ToClose = append(target.ToClose, inputs[i].ToClose...)
		inputs[i].MetadataSources = nil
		inputs[i].StatsCollectors = nil
		inputs[i].ToClose = nil
	}
}

func (r opResult) createAndWrapRowSource(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	args *colexecargs.NewColOperatorArgs,
	inputs []colexecargs.OpWithMetaInfo,
	inputTypes [][]*types.T,
	spec *execinfrapb.ProcessorSpec,
	factory coldata.ColumnFactory,
	causeToWrap error,
) error {
	__antithesis_instrumentation__.Notify(275825)
	if args.ProcessorConstructor == nil {
		__antithesis_instrumentation__.Notify(275831)
		return errors.New("processorConstructor is nil")
	} else {
		__antithesis_instrumentation__.Notify(275832)
	}
	__antithesis_instrumentation__.Notify(275826)
	log.VEventf(ctx, 1, "planning a row-execution processor in the vectorized flow: %v", causeToWrap)
	if err := canWrap(flowCtx.EvalCtx.SessionData().VectorizeMode, spec); err != nil {
		__antithesis_instrumentation__.Notify(275833)
		log.VEventf(ctx, 1, "planning a wrapped processor failed: %v", err)

		return causeToWrap
	} else {
		__antithesis_instrumentation__.Notify(275834)
	}
	__antithesis_instrumentation__.Notify(275827)

	materializerSafeToRelease := len(args.LocalProcessors) == 0
	c, releasables, err := wrapRowSources(
		ctx,
		flowCtx,
		inputs,
		inputTypes,
		args.StreamingMemAccount,
		spec.ProcessorID,
		func(inputs []execinfra.RowSource) (execinfra.RowSource, error) {
			__antithesis_instrumentation__.Notify(275835)

			proc, err := args.ProcessorConstructor(
				ctx, flowCtx, spec.ProcessorID, &spec.Core, &spec.Post, inputs,
				[]execinfra.RowReceiver{nil}, args.LocalProcessors,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(275838)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(275839)
			}
			__antithesis_instrumentation__.Notify(275836)
			var (
				rs execinfra.RowSource
				ok bool
			)
			if rs, ok = proc.(execinfra.RowSource); !ok {
				__antithesis_instrumentation__.Notify(275840)
				return nil, errors.AssertionFailedf(
					"processor %s is not an execinfra.RowSource", spec.Core.String(),
				)
			} else {
				__antithesis_instrumentation__.Notify(275841)
			}
			__antithesis_instrumentation__.Notify(275837)
			r.ColumnTypes = rs.OutputTypes()
			return rs, nil
		},
		materializerSafeToRelease,
		factory,
	)
	__antithesis_instrumentation__.Notify(275828)
	if err != nil {
		__antithesis_instrumentation__.Notify(275842)
		return err
	} else {
		__antithesis_instrumentation__.Notify(275843)
	}
	__antithesis_instrumentation__.Notify(275829)
	r.Root = c
	r.Columnarizer = c
	if buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(275844)
		r.Root = colexec.NewInvariantsChecker(r.Root)
	} else {
		__antithesis_instrumentation__.Notify(275845)
	}
	__antithesis_instrumentation__.Notify(275830)
	takeOverMetaInfo(&r.OpWithMetaInfo, inputs)
	r.MetadataSources = append(r.MetadataSources, r.Root.(colexecop.MetadataSource))
	r.ToClose = append(r.ToClose, r.Root.(colexecop.Closer))
	r.Releasables = append(r.Releasables, releasables...)
	return nil
}

func MaybeRemoveRootColumnarizer(r colexecargs.OpWithMetaInfo) execinfra.RowSource {
	__antithesis_instrumentation__.Notify(275846)
	root := r.Root
	if buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(275850)

		root = colexec.MaybeUnwrapInvariantsChecker(root)
	} else {
		__antithesis_instrumentation__.Notify(275851)
	}
	__antithesis_instrumentation__.Notify(275847)
	c, isColumnarizer := root.(*colexec.Columnarizer)
	if !isColumnarizer {
		__antithesis_instrumentation__.Notify(275852)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(275853)
	}
	__antithesis_instrumentation__.Notify(275848)

	if len(r.StatsCollectors) != 0 || func() bool {
		__antithesis_instrumentation__.Notify(275854)
		return len(r.MetadataSources) != 1 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(275855)
		return len(r.ToClose) != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(275856)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(275857)
	}
	__antithesis_instrumentation__.Notify(275849)
	c.MarkAsRemovedFromFlow()
	return c.Input()
}

func getStreamingAllocator(
	ctx context.Context, args *colexecargs.NewColOperatorArgs,
) *colmem.Allocator {
	__antithesis_instrumentation__.Notify(275858)
	return colmem.NewAllocator(ctx, args.StreamingMemAccount, args.Factory)
}

func NewColOperator(
	ctx context.Context, flowCtx *execinfra.FlowCtx, args *colexecargs.NewColOperatorArgs,
) (_ *colexecargs.NewColOperatorResult, err error) {
	__antithesis_instrumentation__.Notify(275859)
	result := opResult{NewColOperatorResult: colexecargs.GetNewColOperatorResult()}
	r := result.NewColOperatorResult
	spec := args.Spec
	inputs := args.Inputs
	if args.Factory == nil {
		__antithesis_instrumentation__.Notify(275871)

		args.Factory = coldataext.NewExtendedColumnFactory(flowCtx.EvalCtx)
	} else {
		__antithesis_instrumentation__.Notify(275872)
	}
	__antithesis_instrumentation__.Notify(275860)
	factory := args.Factory
	if args.ExprHelper == nil {
		__antithesis_instrumentation__.Notify(275873)
		args.ExprHelper = colexecargs.NewExprHelper()
		args.ExprHelper.SemaCtx = flowCtx.NewSemaContext(flowCtx.Txn)
	} else {
		__antithesis_instrumentation__.Notify(275874)
	}
	__antithesis_instrumentation__.Notify(275861)
	if args.MonitorRegistry == nil {
		__antithesis_instrumentation__.Notify(275875)
		args.MonitorRegistry = &colexecargs.MonitorRegistry{}
	} else {
		__antithesis_instrumentation__.Notify(275876)
	}
	__antithesis_instrumentation__.Notify(275862)

	core := &spec.Core
	post := &spec.Post

	if err = supportedNatively(spec); err != nil {
		__antithesis_instrumentation__.Notify(275877)
		inputTypes := make([][]*types.T, len(spec.Input))
		for inputIdx, input := range spec.Input {
			__antithesis_instrumentation__.Notify(275879)
			inputTypes[inputIdx] = make([]*types.T, len(input.ColumnTypes))
			copy(inputTypes[inputIdx], input.ColumnTypes)
		}
		__antithesis_instrumentation__.Notify(275878)

		err = result.createAndWrapRowSource(ctx, flowCtx, args, inputs, inputTypes, spec, factory, err)

		post = &execinfrapb.PostProcessSpec{}
	} else {
		__antithesis_instrumentation__.Notify(275880)
		switch {
		case core.Noop != nil:
			__antithesis_instrumentation__.Notify(275881)
			if err := checkNumIn(inputs, 1); err != nil {
				__antithesis_instrumentation__.Notify(275917)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275918)
			}
			__antithesis_instrumentation__.Notify(275882)
			result.Root = colexecop.NewNoop(inputs[0].Root)
			result.ColumnTypes = make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(result.ColumnTypes, spec.Input[0].ColumnTypes)

		case core.Values != nil:
			__antithesis_instrumentation__.Notify(275883)
			if err := checkNumIn(inputs, 0); err != nil {
				__antithesis_instrumentation__.Notify(275919)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275920)
			}
			__antithesis_instrumentation__.Notify(275884)
			if core.Values.NumRows == 0 || func() bool {
				__antithesis_instrumentation__.Notify(275921)
				return len(core.Values.Columns) == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(275922)

				result.Root = colexecutils.NewFixedNumTuplesNoInputOp(
					getStreamingAllocator(ctx, args), int(core.Values.NumRows), nil,
				)
			} else {
				__antithesis_instrumentation__.Notify(275923)
				result.Root = colexec.NewValuesOp(getStreamingAllocator(ctx, args), core.Values)
			}
			__antithesis_instrumentation__.Notify(275885)
			result.ColumnTypes = make([]*types.T, len(core.Values.Columns))
			for i, col := range core.Values.Columns {
				__antithesis_instrumentation__.Notify(275924)
				result.ColumnTypes[i] = col.Type
			}

		case core.TableReader != nil:
			__antithesis_instrumentation__.Notify(275886)
			if err := checkNumIn(inputs, 0); err != nil {
				__antithesis_instrumentation__.Notify(275925)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275926)
			}
			__antithesis_instrumentation__.Notify(275887)

			cFetcherMemAcc := args.MonitorRegistry.CreateUnlimitedMemAccount(
				ctx, flowCtx, "cfetcher", spec.ProcessorID,
			)
			kvFetcherMemAcc := args.MonitorRegistry.CreateUnlimitedMemAccount(
				ctx, flowCtx, "kvfetcher", spec.ProcessorID,
			)
			estimatedRowCount := spec.EstimatedRowCount
			scanOp, err := colfetcher.NewColBatchScan(
				ctx, colmem.NewAllocator(ctx, cFetcherMemAcc, factory), kvFetcherMemAcc,
				flowCtx, core.TableReader, post, estimatedRowCount,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(275927)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275928)
			}
			__antithesis_instrumentation__.Notify(275888)
			result.finishScanPlanning(scanOp, scanOp.ResultTypes)

		case core.JoinReader != nil:
			__antithesis_instrumentation__.Notify(275889)
			if err := checkNumIn(inputs, 1); err != nil {
				__antithesis_instrumentation__.Notify(275929)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275930)
			}
			__antithesis_instrumentation__.Notify(275890)
			if !core.JoinReader.IsIndexJoin() {
				__antithesis_instrumentation__.Notify(275931)
				return r, errors.AssertionFailedf("lookup join reader is unsupported in vectorized")
			} else {
				__antithesis_instrumentation__.Notify(275932)
			}
			__antithesis_instrumentation__.Notify(275891)

			cFetcherMemAcc := args.MonitorRegistry.CreateUnlimitedMemAccount(
				ctx, flowCtx, "cfetcher", spec.ProcessorID,
			)
			kvFetcherMemAcc := args.MonitorRegistry.CreateUnlimitedMemAccount(
				ctx, flowCtx, "kvfetcher", spec.ProcessorID,
			)

			streamerBudgetAcc := args.MonitorRegistry.CreateUnlimitedMemAccount(
				ctx, flowCtx, "streamer", spec.ProcessorID,
			)
			streamerDiskMonitor := args.MonitorRegistry.CreateDiskMonitor(
				ctx, flowCtx, "streamer", spec.ProcessorID,
			)
			inputTypes := make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(inputTypes, spec.Input[0].ColumnTypes)
			indexJoinOp, err := colfetcher.NewColIndexJoin(
				ctx, getStreamingAllocator(ctx, args),
				colmem.NewAllocator(ctx, cFetcherMemAcc, factory),
				kvFetcherMemAcc, streamerBudgetAcc, flowCtx,
				inputs[0].Root, core.JoinReader, post, inputTypes, streamerDiskMonitor,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(275933)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275934)
			}
			__antithesis_instrumentation__.Notify(275892)
			result.finishScanPlanning(indexJoinOp, indexJoinOp.ResultTypes)

		case core.Filterer != nil:
			__antithesis_instrumentation__.Notify(275893)
			if err := checkNumIn(inputs, 1); err != nil {
				__antithesis_instrumentation__.Notify(275935)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275936)
			}
			__antithesis_instrumentation__.Notify(275894)

			result.ColumnTypes = make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(result.ColumnTypes, spec.Input[0].ColumnTypes)
			result.Root = inputs[0].Root
			if err := result.planAndMaybeWrapFilter(
				ctx, flowCtx, args, spec.ProcessorID, core.Filterer.Filter, factory,
			); err != nil {
				__antithesis_instrumentation__.Notify(275937)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275938)
			}

		case core.Aggregator != nil:
			__antithesis_instrumentation__.Notify(275895)
			if err := checkNumIn(inputs, 1); err != nil {
				__antithesis_instrumentation__.Notify(275939)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275940)
			}
			__antithesis_instrumentation__.Notify(275896)
			aggSpec := core.Aggregator
			if len(aggSpec.Aggregations) == 0 {
				__antithesis_instrumentation__.Notify(275941)

				result.Root, err = colexecutils.NewFixedNumTuplesNoInputOp(
					getStreamingAllocator(ctx, args), 1, inputs[0].Root,
				), nil

				result.ColumnTypes = []*types.T{}
				break
			} else {
				__antithesis_instrumentation__.Notify(275942)
			}
			__antithesis_instrumentation__.Notify(275897)
			if aggSpec.IsRowCount() {
				__antithesis_instrumentation__.Notify(275943)
				result.Root, err = colexec.NewCountOp(getStreamingAllocator(ctx, args), inputs[0].Root), nil
				result.ColumnTypes = []*types.T{types.Int}
				break
			} else {
				__antithesis_instrumentation__.Notify(275944)
			}
			__antithesis_instrumentation__.Notify(275898)

			var needHash bool
			needHash, err = needHashAggregator(aggSpec)
			if err != nil {
				__antithesis_instrumentation__.Notify(275945)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275946)
			}
			__antithesis_instrumentation__.Notify(275899)
			inputTypes := make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(inputTypes, spec.Input[0].ColumnTypes)

			evalCtx := flowCtx.NewEvalCtx()
			newAggArgs := &colexecagg.NewAggregatorArgs{
				Input:      inputs[0].Root,
				InputTypes: inputTypes,
				Spec:       aggSpec,
				EvalCtx:    evalCtx,
			}
			newAggArgs.Constructors, newAggArgs.ConstArguments, newAggArgs.OutputTypes, err = colexecagg.ProcessAggregations(
				evalCtx, args.ExprHelper.SemaCtx, aggSpec.Aggregations, inputTypes,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(275947)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275948)
			}
			__antithesis_instrumentation__.Notify(275900)
			result.ColumnTypes = newAggArgs.OutputTypes

			if needHash {
				__antithesis_instrumentation__.Notify(275949)
				opName := "hash-aggregator"
				outputUnlimitedAllocator := colmem.NewAllocator(
					ctx,
					args.MonitorRegistry.CreateUnlimitedMemAccount(ctx, flowCtx, opName+"-output", spec.ProcessorID),
					factory,
				)

				diskSpillingDisabled := !colexec.HashAggregationDiskSpillingEnabled.Get(&flowCtx.Cfg.Settings.SV)
				if diskSpillingDisabled {
					__antithesis_instrumentation__.Notify(275950)

					hashAggregatorUnlimitedMemAccount := args.MonitorRegistry.CreateUnlimitedMemAccount(
						ctx, flowCtx, opName, spec.ProcessorID,
					)
					newAggArgs.Allocator = colmem.NewAllocator(
						ctx, hashAggregatorUnlimitedMemAccount, factory,
					)
					newAggArgs.MemAccount = hashAggregatorUnlimitedMemAccount
					evalCtx.SingleDatumAggMemAccount = hashAggregatorUnlimitedMemAccount
					maxOutputBatchMemSize := execinfra.GetWorkMemLimit(flowCtx)

					result.Root = colexec.NewHashAggregator(
						newAggArgs, nil,
						outputUnlimitedAllocator, maxOutputBatchMemSize,
					)
				} else {
					__antithesis_instrumentation__.Notify(275951)

					totalMemLimit := execinfra.GetWorkMemLimit(flowCtx)

					maxOutputBatchMemSize := totalMemLimit / 10
					hashAggregationMemLimit := totalMemLimit/2 - maxOutputBatchMemSize
					inputTuplesTrackingMemLimit := totalMemLimit / 2
					if totalMemLimit == 1 {
						__antithesis_instrumentation__.Notify(275953)

						maxOutputBatchMemSize = 1
						hashAggregationMemLimit = 1
						inputTuplesTrackingMemLimit = 1
					} else {
						__antithesis_instrumentation__.Notify(275954)
					}
					__antithesis_instrumentation__.Notify(275952)
					hashAggregatorMemAccount, hashAggregatorMemMonitorName := args.MonitorRegistry.CreateMemAccountForSpillStrategyWithLimit(
						ctx, flowCtx, hashAggregationMemLimit, opName, spec.ProcessorID,
					)
					spillingQueueMemMonitorName := hashAggregatorMemMonitorName + "-spilling-queue"

					spillingQueueMemAccount := args.MonitorRegistry.CreateUnlimitedMemAccount(
						ctx, flowCtx, spillingQueueMemMonitorName, spec.ProcessorID,
					)
					newAggArgs.Allocator = colmem.NewAllocator(ctx, hashAggregatorMemAccount, factory)
					newAggArgs.MemAccount = hashAggregatorMemAccount
					inMemoryHashAggregator := colexec.NewHashAggregator(
						newAggArgs,
						&colexecutils.NewSpillingQueueArgs{
							UnlimitedAllocator: colmem.NewAllocator(ctx, spillingQueueMemAccount, factory),
							Types:              inputTypes,
							MemoryLimit:        inputTuplesTrackingMemLimit,
							DiskQueueCfg:       args.DiskQueueCfg,
							FDSemaphore:        args.FDSemaphore,
							DiskAcc:            args.MonitorRegistry.CreateDiskAccount(ctx, flowCtx, spillingQueueMemMonitorName, spec.ProcessorID),
						},
						outputUnlimitedAllocator,
						maxOutputBatchMemSize,
					)
					ehaOpName := "external-hash-aggregator"
					ehaMemAccount := args.MonitorRegistry.CreateUnlimitedMemAccount(ctx, flowCtx, ehaOpName, spec.ProcessorID)

					evalCtx.SingleDatumAggMemAccount = ehaMemAccount
					result.Root = colexec.NewOneInputDiskSpiller(
						inputs[0].Root, inMemoryHashAggregator.(colexecop.BufferingInMemoryOperator),
						hashAggregatorMemMonitorName,
						func(input colexecop.Operator) colexecop.Operator {
							__antithesis_instrumentation__.Notify(275955)
							newAggArgs := *newAggArgs

							newAggArgs.Allocator = colmem.NewAllocator(ctx, ehaMemAccount, factory)
							newAggArgs.MemAccount = ehaMemAccount
							newAggArgs.Input = input
							return colexec.NewExternalHashAggregator(
								flowCtx,
								args,
								&newAggArgs,
								result.makeDiskBackedSorterConstructor(ctx, flowCtx, args, ehaOpName, factory),
								args.MonitorRegistry.CreateDiskAccount(ctx, flowCtx, ehaOpName, spec.ProcessorID),

								outputUnlimitedAllocator,
								maxOutputBatchMemSize,
							)
						},
						args.TestingKnobs.SpillingCallbackFn,
					)
				}
			} else {
				__antithesis_instrumentation__.Notify(275956)
				evalCtx.SingleDatumAggMemAccount = args.StreamingMemAccount
				newAggArgs.Allocator = getStreamingAllocator(ctx, args)
				newAggArgs.MemAccount = args.StreamingMemAccount
				result.Root = colexec.NewOrderedAggregator(newAggArgs)
			}
			__antithesis_instrumentation__.Notify(275901)
			result.ToClose = append(result.ToClose, result.Root.(colexecop.Closer))

		case core.Distinct != nil:
			__antithesis_instrumentation__.Notify(275902)
			if err := checkNumIn(inputs, 1); err != nil {
				__antithesis_instrumentation__.Notify(275957)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275958)
			}
			__antithesis_instrumentation__.Notify(275903)
			result.ColumnTypes = make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(result.ColumnTypes, spec.Input[0].ColumnTypes)
			if len(core.Distinct.OrderedColumns) == len(core.Distinct.DistinctColumns) {
				__antithesis_instrumentation__.Notify(275959)
				result.Root = colexecbase.NewOrderedDistinct(
					inputs[0].Root, core.Distinct.OrderedColumns, result.ColumnTypes,
					core.Distinct.NullsAreDistinct, core.Distinct.ErrorOnDup,
				)
			} else {
				__antithesis_instrumentation__.Notify(275960)

				distinctMemAccount, distinctMemMonitorName := args.MonitorRegistry.CreateMemAccountForSpillStrategy(
					ctx, flowCtx, "distinct", spec.ProcessorID,
				)

				allocator := colmem.NewAllocator(ctx, distinctMemAccount, factory)
				inMemoryUnorderedDistinct := colexec.NewUnorderedDistinct(
					allocator, inputs[0].Root, core.Distinct.DistinctColumns, result.ColumnTypes,
					core.Distinct.NullsAreDistinct, core.Distinct.ErrorOnDup,
				)
				edOpName := "external-distinct"
				diskAccount := args.MonitorRegistry.CreateDiskAccount(ctx, flowCtx, edOpName, spec.ProcessorID)
				result.Root = colexec.NewOneInputDiskSpiller(
					inputs[0].Root, inMemoryUnorderedDistinct.(colexecop.BufferingInMemoryOperator),
					distinctMemMonitorName,
					func(input colexecop.Operator) colexecop.Operator {
						__antithesis_instrumentation__.Notify(275962)
						unlimitedAllocator := colmem.NewAllocator(
							ctx, args.MonitorRegistry.CreateUnlimitedMemAccount(ctx, flowCtx, edOpName, spec.ProcessorID), factory,
						)
						return colexec.NewExternalDistinct(
							unlimitedAllocator,
							flowCtx,
							args,
							input,
							result.ColumnTypes,
							result.makeDiskBackedSorterConstructor(ctx, flowCtx, args, edOpName, factory),
							inMemoryUnorderedDistinct,
							diskAccount,
						)
					},
					args.TestingKnobs.SpillingCallbackFn,
				)
				__antithesis_instrumentation__.Notify(275961)
				result.ToClose = append(result.ToClose, result.Root.(colexecop.Closer))
			}

		case core.Ordinality != nil:
			__antithesis_instrumentation__.Notify(275904)
			if err := checkNumIn(inputs, 1); err != nil {
				__antithesis_instrumentation__.Notify(275963)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275964)
			}
			__antithesis_instrumentation__.Notify(275905)
			outputIdx := len(spec.Input[0].ColumnTypes)
			result.Root = colexecbase.NewOrdinalityOp(
				getStreamingAllocator(ctx, args), inputs[0].Root, outputIdx,
			)
			result.ColumnTypes = appendOneType(spec.Input[0].ColumnTypes, types.Int)

		case core.HashJoiner != nil:
			__antithesis_instrumentation__.Notify(275906)
			if err := checkNumIn(inputs, 2); err != nil {
				__antithesis_instrumentation__.Notify(275965)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275966)
			}
			__antithesis_instrumentation__.Notify(275907)
			leftTypes := make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(leftTypes, spec.Input[0].ColumnTypes)
			rightTypes := make([]*types.T, len(spec.Input[1].ColumnTypes))
			copy(rightTypes, spec.Input[1].ColumnTypes)

			memoryLimit := execinfra.GetWorkMemLimit(flowCtx)
			if len(core.HashJoiner.LeftEqColumns) == 0 {
				__antithesis_instrumentation__.Notify(275967)

				opName := "cross-joiner"
				crossJoinerMemAccount := args.MonitorRegistry.CreateUnlimitedMemAccount(ctx, flowCtx, opName, spec.ProcessorID)
				crossJoinerDiskAcc := args.MonitorRegistry.CreateDiskAccount(ctx, flowCtx, opName, spec.ProcessorID)
				unlimitedAllocator := colmem.NewAllocator(ctx, crossJoinerMemAccount, factory)
				result.Root = colexecjoin.NewCrossJoiner(
					unlimitedAllocator,
					memoryLimit,
					args.DiskQueueCfg,
					args.FDSemaphore,
					core.HashJoiner.Type,
					inputs[0].Root, inputs[1].Root,
					leftTypes, rightTypes,
					crossJoinerDiskAcc,
				)
				result.ToClose = append(result.ToClose, result.Root.(colexecop.Closer))
			} else {
				__antithesis_instrumentation__.Notify(275968)
				opName := "hash-joiner"
				hashJoinerMemAccount, hashJoinerMemMonitorName := args.MonitorRegistry.CreateMemAccountForSpillStrategy(
					ctx, flowCtx, opName, spec.ProcessorID,
				)
				hashJoinerUnlimitedAllocator := colmem.NewAllocator(
					ctx, args.MonitorRegistry.CreateUnlimitedMemAccount(ctx, flowCtx, opName, spec.ProcessorID), factory,
				)
				hjSpec := colexecjoin.MakeHashJoinerSpec(
					core.HashJoiner.Type,
					core.HashJoiner.LeftEqColumns,
					core.HashJoiner.RightEqColumns,
					leftTypes,
					rightTypes,
					core.HashJoiner.RightEqColumnsAreKey,
				)

				inMemoryHashJoiner := colexecjoin.NewHashJoiner(
					colmem.NewAllocator(ctx, hashJoinerMemAccount, factory),
					hashJoinerUnlimitedAllocator, hjSpec, inputs[0].Root, inputs[1].Root,
					colexecjoin.HashJoinerInitialNumBuckets,
				)
				if args.TestingKnobs.DiskSpillingDisabled {
					__antithesis_instrumentation__.Notify(275969)

					result.Root = inMemoryHashJoiner
				} else {
					__antithesis_instrumentation__.Notify(275970)
					opName := "external-hash-joiner"
					diskAccount := args.MonitorRegistry.CreateDiskAccount(ctx, flowCtx, opName, spec.ProcessorID)
					result.Root = colexec.NewTwoInputDiskSpiller(
						inputs[0].Root, inputs[1].Root, inMemoryHashJoiner.(colexecop.BufferingInMemoryOperator),
						hashJoinerMemMonitorName,
						func(inputOne, inputTwo colexecop.Operator) colexecop.Operator {
							__antithesis_instrumentation__.Notify(275971)
							unlimitedAllocator := colmem.NewAllocator(
								ctx, args.MonitorRegistry.CreateUnlimitedMemAccount(ctx, flowCtx, opName, spec.ProcessorID), factory,
							)
							ehj := colexec.NewExternalHashJoiner(
								unlimitedAllocator,
								flowCtx,
								args,
								hjSpec,
								inputOne, inputTwo,
								result.makeDiskBackedSorterConstructor(ctx, flowCtx, args, opName, factory),
								diskAccount,
							)
							result.ToClose = append(result.ToClose, ehj.(colexecop.Closer))
							return ehj
						},
						args.TestingKnobs.SpillingCallbackFn,
					)
				}
			}
			__antithesis_instrumentation__.Notify(275908)

			result.ColumnTypes = core.HashJoiner.Type.MakeOutputTypes(leftTypes, rightTypes)

			if !core.HashJoiner.OnExpr.Empty() && func() bool {
				__antithesis_instrumentation__.Notify(275972)
				return core.HashJoiner.Type == descpb.InnerJoin == true
			}() == true {
				__antithesis_instrumentation__.Notify(275973)
				if err = result.planAndMaybeWrapFilter(
					ctx, flowCtx, args, spec.ProcessorID, core.HashJoiner.OnExpr, factory,
				); err != nil {
					__antithesis_instrumentation__.Notify(275974)
					return r, err
				} else {
					__antithesis_instrumentation__.Notify(275975)
				}
			} else {
				__antithesis_instrumentation__.Notify(275976)
			}

		case core.MergeJoiner != nil:
			__antithesis_instrumentation__.Notify(275909)
			if err := checkNumIn(inputs, 2); err != nil {
				__antithesis_instrumentation__.Notify(275977)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275978)
			}
			__antithesis_instrumentation__.Notify(275910)

			leftTypes := make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(leftTypes, spec.Input[0].ColumnTypes)
			rightTypes := make([]*types.T, len(spec.Input[1].ColumnTypes))
			copy(rightTypes, spec.Input[1].ColumnTypes)

			joinType := core.MergeJoiner.Type
			var onExpr *execinfrapb.Expression
			if !core.MergeJoiner.OnExpr.Empty() {
				__antithesis_instrumentation__.Notify(275979)
				if joinType != descpb.InnerJoin {
					__antithesis_instrumentation__.Notify(275981)
					return r, errors.AssertionFailedf(
						"ON expression (%s) was unexpectedly planned for merge joiner with join type %s",
						core.MergeJoiner.OnExpr.String(), core.MergeJoiner.Type.String(),
					)
				} else {
					__antithesis_instrumentation__.Notify(275982)
				}
				__antithesis_instrumentation__.Notify(275980)
				onExpr = &core.MergeJoiner.OnExpr
			} else {
				__antithesis_instrumentation__.Notify(275983)
			}
			__antithesis_instrumentation__.Notify(275911)

			opName := "merge-joiner"

			unlimitedAllocator := colmem.NewAllocator(
				ctx, args.MonitorRegistry.CreateUnlimitedMemAccount(
					ctx, flowCtx, opName, spec.ProcessorID,
				), factory)
			diskAccount := args.MonitorRegistry.CreateDiskAccount(ctx, flowCtx, opName, spec.ProcessorID)
			mj := colexecjoin.NewMergeJoinOp(
				unlimitedAllocator, execinfra.GetWorkMemLimit(flowCtx),
				args.DiskQueueCfg, args.FDSemaphore,
				joinType, inputs[0].Root, inputs[1].Root, leftTypes, rightTypes,
				core.MergeJoiner.LeftOrdering.Columns, core.MergeJoiner.RightOrdering.Columns,
				diskAccount, flowCtx.EvalCtx,
			)

			result.Root = mj
			result.ToClose = append(result.ToClose, mj.(colexecop.Closer))
			result.ColumnTypes = core.MergeJoiner.Type.MakeOutputTypes(leftTypes, rightTypes)

			if onExpr != nil {
				__antithesis_instrumentation__.Notify(275984)
				if err = result.planAndMaybeWrapFilter(
					ctx, flowCtx, args, spec.ProcessorID, *onExpr, factory,
				); err != nil {
					__antithesis_instrumentation__.Notify(275985)
					return r, err
				} else {
					__antithesis_instrumentation__.Notify(275986)
				}
			} else {
				__antithesis_instrumentation__.Notify(275987)
			}

		case core.Sorter != nil:
			__antithesis_instrumentation__.Notify(275912)
			if err := checkNumIn(inputs, 1); err != nil {
				__antithesis_instrumentation__.Notify(275988)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275989)
			}
			__antithesis_instrumentation__.Notify(275913)
			input := inputs[0].Root
			result.ColumnTypes = make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(result.ColumnTypes, spec.Input[0].ColumnTypes)
			ordering := core.Sorter.OutputOrdering
			matchLen := core.Sorter.OrderingMatchLen
			limit := core.Sorter.Limit
			result.Root = result.createDiskBackedSort(
				ctx, flowCtx, args, input, result.ColumnTypes, ordering, limit, matchLen, 0,
				spec.ProcessorID, "", factory,
			)

		case core.Windower != nil:
			__antithesis_instrumentation__.Notify(275914)
			if err := checkNumIn(inputs, 1); err != nil {
				__antithesis_instrumentation__.Notify(275990)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(275991)
			}
			__antithesis_instrumentation__.Notify(275915)
			opNamePrefix := "window-"
			input := inputs[0].Root
			result.ColumnTypes = make([]*types.T, len(spec.Input[0].ColumnTypes))
			copy(result.ColumnTypes, spec.Input[0].ColumnTypes)
			for _, wf := range core.Windower.WindowFns {
				__antithesis_instrumentation__.Notify(275992)

				typs := make([]*types.T, len(result.ColumnTypes), len(result.ColumnTypes)+len(wf.ArgsIdxs)+2)
				copy(typs, result.ColumnTypes)

				wf.Frame = colexecwindow.NormalizeWindowFrame(wf.Frame)

				tempColOffset := uint32(0)
				argTypes := make([]*types.T, len(wf.ArgsIdxs))
				argIdxs := make([]int, len(wf.ArgsIdxs))
				for i, idx := range wf.ArgsIdxs {
					__antithesis_instrumentation__.Notify(275999)

					needsCast, expectedType := colexecwindow.WindowFnArgNeedsCast(wf.Func, typs[idx], i)
					if needsCast {
						__antithesis_instrumentation__.Notify(276001)

						castIdx := len(typs)
						input, err = colexecbase.GetCastOperator(
							getStreamingAllocator(ctx, args), input, int(idx),
							castIdx, typs[idx], expectedType, flowCtx.EvalCtx,
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(276003)
							colexecerror.InternalError(errors.AssertionFailedf(
								"failed to cast window function argument to type %v", expectedType))
						} else {
							__antithesis_instrumentation__.Notify(276004)
						}
						__antithesis_instrumentation__.Notify(276002)
						typs = append(typs, expectedType)
						idx = uint32(castIdx)
						tempColOffset++
					} else {
						__antithesis_instrumentation__.Notify(276005)
					}
					__antithesis_instrumentation__.Notify(276000)
					argTypes[i] = expectedType
					argIdxs[i] = int(idx)
				}
				__antithesis_instrumentation__.Notify(275993)
				partitionColIdx := tree.NoColumnIdx
				peersColIdx := tree.NoColumnIdx

				if len(core.Windower.PartitionBy) > 0 {
					__antithesis_instrumentation__.Notify(276006)

					partitionColIdx = int(wf.OutputColIdx + tempColOffset)
					input = colexecwindow.NewWindowSortingPartitioner(
						getStreamingAllocator(ctx, args), input, typs,
						core.Windower.PartitionBy, wf.Ordering.Columns, partitionColIdx,
						func(input colexecop.Operator, inputTypes []*types.T, orderingCols []execinfrapb.Ordering_Column) colexecop.Operator {
							__antithesis_instrumentation__.Notify(276008)
							return result.createDiskBackedSort(
								ctx, flowCtx, args, input, inputTypes,
								execinfrapb.Ordering{Columns: orderingCols}, 0, 0,
								0, spec.ProcessorID,
								opNamePrefix, factory)
						},
					)
					__antithesis_instrumentation__.Notify(276007)

					tempColOffset++
					typs = typs[:len(typs)+1]
					typs[len(typs)-1] = types.Bool
				} else {
					__antithesis_instrumentation__.Notify(276009)
					if len(wf.Ordering.Columns) > 0 {
						__antithesis_instrumentation__.Notify(276010)
						input = result.createDiskBackedSort(
							ctx, flowCtx, args, input, typs,
							wf.Ordering, 0, 0, 0,
							spec.ProcessorID, opNamePrefix, factory,
						)
					} else {
						__antithesis_instrumentation__.Notify(276011)
					}
				}
				__antithesis_instrumentation__.Notify(275994)
				if colexecwindow.WindowFnNeedsPeersInfo(&wf) {
					__antithesis_instrumentation__.Notify(276012)
					peersColIdx = int(wf.OutputColIdx + tempColOffset)
					input = colexecwindow.NewWindowPeerGrouper(
						getStreamingAllocator(ctx, args), input, typs,
						wf.Ordering.Columns, partitionColIdx, peersColIdx,
					)

					tempColOffset++
					typs = typs[:len(typs)+1]
					typs[len(typs)-1] = types.Bool
				} else {
					__antithesis_instrumentation__.Notify(276013)
				}
				__antithesis_instrumentation__.Notify(275995)
				outputIdx := int(wf.OutputColIdx + tempColOffset)

				windowArgs := &colexecwindow.WindowArgs{
					EvalCtx:         flowCtx.EvalCtx,
					MemoryLimit:     execinfra.GetWorkMemLimit(flowCtx),
					QueueCfg:        args.DiskQueueCfg,
					FdSemaphore:     args.FDSemaphore,
					Input:           input,
					InputTypes:      typs,
					OutputColIdx:    outputIdx,
					PartitionColIdx: partitionColIdx,
					PeersColIdx:     peersColIdx,
				}

				returnType := types.Int
				if wf.Func.WindowFunc != nil {
					__antithesis_instrumentation__.Notify(276014)

					windowFn := *wf.Func.WindowFunc

					switch windowFn {
					case execinfrapb.WindowerSpec_ROW_NUMBER:
						__antithesis_instrumentation__.Notify(276015)
						windowArgs.MainAllocator = getStreamingAllocator(ctx, args)
						result.Root = colexecwindow.NewRowNumberOperator(windowArgs)
					case execinfrapb.WindowerSpec_RANK, execinfrapb.WindowerSpec_DENSE_RANK:
						__antithesis_instrumentation__.Notify(276016)
						windowArgs.MainAllocator = getStreamingAllocator(ctx, args)
						result.Root, err = colexecwindow.NewRankOperator(windowArgs, windowFn, wf.Ordering.Columns)
					case execinfrapb.WindowerSpec_PERCENT_RANK, execinfrapb.WindowerSpec_CUME_DIST:
						__antithesis_instrumentation__.Notify(276017)
						opName := opNamePrefix + "relative-rank"
						result.finishBufferedWindowerArgs(
							ctx, flowCtx, args.MonitorRegistry, windowArgs, opName,
							spec.ProcessorID, factory, false,
						)
						result.Root, err = colexecwindow.NewRelativeRankOperator(
							windowArgs, windowFn, wf.Ordering.Columns)
						returnType = types.Float
					case execinfrapb.WindowerSpec_NTILE:
						__antithesis_instrumentation__.Notify(276018)
						opName := opNamePrefix + "ntile"
						result.finishBufferedWindowerArgs(
							ctx, flowCtx, args.MonitorRegistry, windowArgs, opName,
							spec.ProcessorID, factory, false,
						)
						result.Root = colexecwindow.NewNTileOperator(windowArgs, argIdxs[0])
					case execinfrapb.WindowerSpec_LAG:
						__antithesis_instrumentation__.Notify(276019)
						opName := opNamePrefix + "lag"
						result.finishBufferedWindowerArgs(
							ctx, flowCtx, args.MonitorRegistry, windowArgs, opName,
							spec.ProcessorID, factory, true,
						)
						result.Root, err = colexecwindow.NewLagOperator(
							windowArgs, argIdxs[0], argIdxs[1], argIdxs[2])
						returnType = typs[argIdxs[0]]
					case execinfrapb.WindowerSpec_LEAD:
						__antithesis_instrumentation__.Notify(276020)
						opName := opNamePrefix + "lead"
						result.finishBufferedWindowerArgs(
							ctx, flowCtx, args.MonitorRegistry, windowArgs, opName,
							spec.ProcessorID, factory, true,
						)
						result.Root, err = colexecwindow.NewLeadOperator(
							windowArgs, argIdxs[0], argIdxs[1], argIdxs[2])
						returnType = typs[argIdxs[0]]
					case execinfrapb.WindowerSpec_FIRST_VALUE:
						__antithesis_instrumentation__.Notify(276021)
						opName := opNamePrefix + "first_value"
						result.finishBufferedWindowerArgs(
							ctx, flowCtx, args.MonitorRegistry, windowArgs, opName,
							spec.ProcessorID, factory, true,
						)
						result.Root, err = colexecwindow.NewFirstValueOperator(
							windowArgs, wf.Frame, &wf.Ordering, argIdxs)
						returnType = typs[argIdxs[0]]
					case execinfrapb.WindowerSpec_LAST_VALUE:
						__antithesis_instrumentation__.Notify(276022)
						opName := opNamePrefix + "last_value"
						result.finishBufferedWindowerArgs(
							ctx, flowCtx, args.MonitorRegistry, windowArgs, opName,
							spec.ProcessorID, factory, true,
						)
						result.Root, err = colexecwindow.NewLastValueOperator(
							windowArgs, wf.Frame, &wf.Ordering, argIdxs)
						returnType = typs[argIdxs[0]]
					case execinfrapb.WindowerSpec_NTH_VALUE:
						__antithesis_instrumentation__.Notify(276023)
						opName := opNamePrefix + "nth_value"
						result.finishBufferedWindowerArgs(
							ctx, flowCtx, args.MonitorRegistry, windowArgs, opName,
							spec.ProcessorID, factory, true,
						)
						result.Root, err = colexecwindow.NewNthValueOperator(
							windowArgs, wf.Frame, &wf.Ordering, argIdxs)
						returnType = typs[argIdxs[0]]
					default:
						__antithesis_instrumentation__.Notify(276024)
						return r, errors.AssertionFailedf("window function %s is not supported", wf.String())
					}
				} else {
					__antithesis_instrumentation__.Notify(276025)
					if wf.Func.AggregateFunc != nil {
						__antithesis_instrumentation__.Notify(276026)

						opName := opNamePrefix + strings.ToLower(wf.Func.AggregateFunc.String())
						result.finishBufferedWindowerArgs(
							ctx, flowCtx, args.MonitorRegistry, windowArgs, opName,
							spec.ProcessorID, factory, true,
						)
						aggType := *wf.Func.AggregateFunc
						switch *wf.Func.AggregateFunc {
						case execinfrapb.CountRows:
							__antithesis_instrumentation__.Notify(276027)

							result.Root = colexecwindow.NewCountRowsOperator(windowArgs, wf.Frame, &wf.Ordering)
						default:
							__antithesis_instrumentation__.Notify(276028)
							aggArgs := colexecagg.NewAggregatorArgs{
								Allocator:  windowArgs.MainAllocator,
								InputTypes: argTypes,
								EvalCtx:    flowCtx.EvalCtx,
							}

							colIdx := make([]uint32, len(argTypes))
							for i := range argIdxs {
								__antithesis_instrumentation__.Notify(276031)
								colIdx[i] = uint32(i)
							}
							__antithesis_instrumentation__.Notify(276029)
							aggregations := []execinfrapb.AggregatorSpec_Aggregation{{
								Func:   aggType,
								ColIdx: colIdx,
							}}
							aggArgs.Constructors, aggArgs.ConstArguments, aggArgs.OutputTypes, err =
								colexecagg.ProcessAggregations(flowCtx.EvalCtx, args.ExprHelper.SemaCtx, aggregations, argTypes)
							var toClose colexecop.Closers
							var aggFnsAlloc *colexecagg.AggregateFuncsAlloc
							if (aggType != execinfrapb.Min && func() bool {
								__antithesis_instrumentation__.Notify(276032)
								return aggType != execinfrapb.Max == true
							}() == true) || func() bool {
								__antithesis_instrumentation__.Notify(276033)
								return wf.Frame.Exclusion != execinfrapb.WindowerSpec_Frame_NO_EXCLUSION == true
							}() == true || func() bool {
								__antithesis_instrumentation__.Notify(276034)
								return !colexecwindow.WindowFrameCanShrink(wf.Frame, &wf.Ordering) == true
							}() == true {
								__antithesis_instrumentation__.Notify(276035)

								aggFnsAlloc, _, toClose, err = colexecagg.NewAggregateFuncsAlloc(
									&aggArgs, aggregations, 1, colexecagg.WindowAggKind,
								)
								if err != nil {
									__antithesis_instrumentation__.Notify(276036)
									colexecerror.InternalError(err)
								} else {
									__antithesis_instrumentation__.Notify(276037)
								}
							} else {
								__antithesis_instrumentation__.Notify(276038)
							}
							__antithesis_instrumentation__.Notify(276030)
							result.Root = colexecwindow.NewWindowAggregatorOperator(
								windowArgs, aggType, wf.Frame, &wf.Ordering, argIdxs,
								aggArgs.OutputTypes[0], aggFnsAlloc, toClose)
							returnType = aggArgs.OutputTypes[0]
						}
					} else {
						__antithesis_instrumentation__.Notify(276039)
						colexecerror.InternalError(errors.AssertionFailedf("window function spec is nil"))
					}
				}
				__antithesis_instrumentation__.Notify(275996)
				if c, ok := result.Root.(colexecop.Closer); ok {
					__antithesis_instrumentation__.Notify(276040)
					result.ToClose = append(result.ToClose, c)
				} else {
					__antithesis_instrumentation__.Notify(276041)
				}
				__antithesis_instrumentation__.Notify(275997)

				if tempColOffset > 0 {
					__antithesis_instrumentation__.Notify(276042)

					projection := make([]uint32, 0, wf.OutputColIdx+tempColOffset)
					for i := uint32(0); i < wf.OutputColIdx; i++ {
						__antithesis_instrumentation__.Notify(276044)
						projection = append(projection, i)
					}
					__antithesis_instrumentation__.Notify(276043)
					projection = append(projection, wf.OutputColIdx+tempColOffset)
					result.Root = colexecbase.NewSimpleProjectOp(result.Root, int(wf.OutputColIdx+tempColOffset), projection)
				} else {
					__antithesis_instrumentation__.Notify(276045)
				}
				__antithesis_instrumentation__.Notify(275998)

				result.ColumnTypes = appendOneType(result.ColumnTypes, returnType)
				input = result.Root
			}

		default:
			__antithesis_instrumentation__.Notify(275916)
			return r, errors.AssertionFailedf("unsupported processor core %q", core)
		}
	}
	__antithesis_instrumentation__.Notify(275863)

	if err != nil {
		__antithesis_instrumentation__.Notify(276046)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(276047)
	}
	__antithesis_instrumentation__.Notify(275864)

	ppr := postProcessResult{
		Op:          result.Root,
		ColumnTypes: result.ColumnTypes,
	}
	err = ppr.planPostProcessSpec(ctx, flowCtx, args, post, factory, &r.Releasables)
	if err != nil {
		__antithesis_instrumentation__.Notify(276048)
		err = result.wrapPostProcessSpec(ctx, flowCtx, args, post, args.Spec.ResultTypes, factory, err)
	} else {
		__antithesis_instrumentation__.Notify(276049)

		result.updateWithPostProcessResult(ppr)
	}
	__antithesis_instrumentation__.Notify(275865)
	if err != nil {
		__antithesis_instrumentation__.Notify(276050)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(276051)
	}
	__antithesis_instrumentation__.Notify(275866)

	if len(args.Spec.ResultTypes) != len(r.ColumnTypes) {
		__antithesis_instrumentation__.Notify(276052)
		return r, errors.AssertionFailedf("unexpectedly different number of columns are output: expected %v, actual %v", args.Spec.ResultTypes, r.ColumnTypes)
	} else {
		__antithesis_instrumentation__.Notify(276053)
	}
	__antithesis_instrumentation__.Notify(275867)
	numMismatchedTypes, needWrappedCast := 0, false
	for i := range args.Spec.ResultTypes {
		__antithesis_instrumentation__.Notify(276054)
		expected, actual := args.Spec.ResultTypes[i], r.ColumnTypes[i]
		if !actual.Identical(expected) {
			__antithesis_instrumentation__.Notify(276055)
			numMismatchedTypes++
			if !colexecbase.IsCastSupported(actual, expected) {
				__antithesis_instrumentation__.Notify(276056)
				needWrappedCast = true
			} else {
				__antithesis_instrumentation__.Notify(276057)
			}
		} else {
			__antithesis_instrumentation__.Notify(276058)
		}
	}
	__antithesis_instrumentation__.Notify(275868)

	if needWrappedCast {
		__antithesis_instrumentation__.Notify(276059)
		post := &execinfrapb.PostProcessSpec{
			RenderExprs: make([]execinfrapb.Expression, len(args.Spec.ResultTypes)),
		}
		for i := range args.Spec.ResultTypes {
			__antithesis_instrumentation__.Notify(276061)
			expected, actual := args.Spec.ResultTypes[i], r.ColumnTypes[i]
			if !actual.Identical(expected) {
				__antithesis_instrumentation__.Notify(276062)
				post.RenderExprs[i].LocalExpr = tree.NewTypedCastExpr(tree.NewTypedOrdinalReference(i, actual), expected)
			} else {
				__antithesis_instrumentation__.Notify(276063)
				post.RenderExprs[i].LocalExpr = tree.NewTypedOrdinalReference(i, args.Spec.ResultTypes[i])
			}
		}
		__antithesis_instrumentation__.Notify(276060)
		if err = result.wrapPostProcessSpec(ctx, flowCtx, args, post, args.Spec.ResultTypes, factory, errWrappedCast); err != nil {
			__antithesis_instrumentation__.Notify(276064)
			return r, err
		} else {
			__antithesis_instrumentation__.Notify(276065)
		}
	} else {
		__antithesis_instrumentation__.Notify(276066)
		if numMismatchedTypes > 0 {
			__antithesis_instrumentation__.Notify(276067)

			projection := make([]uint32, len(args.Spec.ResultTypes))
			typesWithCasts := make([]*types.T, len(args.Spec.ResultTypes), len(args.Spec.ResultTypes)+numMismatchedTypes)

			copy(typesWithCasts, r.ColumnTypes)
			for i := range args.Spec.ResultTypes {
				__antithesis_instrumentation__.Notify(276069)
				expected, actual := args.Spec.ResultTypes[i], r.ColumnTypes[i]
				if !actual.Identical(expected) {
					__antithesis_instrumentation__.Notify(276070)
					castedIdx := len(typesWithCasts)
					r.Root, err = colexecbase.GetCastOperator(
						getStreamingAllocator(ctx, args), r.Root, i, castedIdx,
						actual, expected, flowCtx.EvalCtx,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(276072)
						return r, errors.NewAssertionErrorWithWrappedErrf(err, "unexpectedly couldn't plan a cast although IsCastSupported returned true")
					} else {
						__antithesis_instrumentation__.Notify(276073)
					}
					__antithesis_instrumentation__.Notify(276071)
					projection[i] = uint32(castedIdx)
					typesWithCasts = append(typesWithCasts, expected)
				} else {
					__antithesis_instrumentation__.Notify(276074)
					projection[i] = uint32(i)
				}
			}
			__antithesis_instrumentation__.Notify(276068)
			r.Root, r.ColumnTypes = addProjection(r.Root, typesWithCasts, projection)
		} else {
			__antithesis_instrumentation__.Notify(276075)
		}
	}
	__antithesis_instrumentation__.Notify(275869)

	takeOverMetaInfo(&result.OpWithMetaInfo, inputs)
	if buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(276076)

		if i := colexec.MaybeUnwrapInvariantsChecker(r.Root); i == r.Root {
			__antithesis_instrumentation__.Notify(276078)
			r.Root = colexec.NewInvariantsChecker(r.Root)
		} else {
			__antithesis_instrumentation__.Notify(276079)
		}
		__antithesis_instrumentation__.Notify(276077)

		args.MonitorRegistry.AssertInvariants()
	} else {
		__antithesis_instrumentation__.Notify(276080)
	}
	__antithesis_instrumentation__.Notify(275870)
	return r, err
}

func (r opResult) planAndMaybeWrapFilter(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	args *colexecargs.NewColOperatorArgs,
	processorID int32,
	filter execinfrapb.Expression,
	factory coldata.ColumnFactory,
) error {
	__antithesis_instrumentation__.Notify(276081)
	op, err := planFilterExpr(
		ctx, flowCtx, r.Root, r.ColumnTypes, filter, args.StreamingMemAccount, factory, args.ExprHelper, &r.Releasables,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(276083)

		filtererSpec := &execinfrapb.ProcessorSpec{
			Core: execinfrapb.ProcessorCoreUnion{
				Filterer: &execinfrapb.FiltererSpec{
					Filter: filter,
				},
			},
			ProcessorID: processorID,
			ResultTypes: args.Spec.ResultTypes,
		}
		inputToMaterializer := colexecargs.OpWithMetaInfo{Root: r.Root}
		takeOverMetaInfo(&inputToMaterializer, args.Inputs)
		return r.createAndWrapRowSource(
			ctx, flowCtx, args, []colexecargs.OpWithMetaInfo{inputToMaterializer},
			[][]*types.T{r.ColumnTypes}, filtererSpec, factory, err,
		)
	} else {
		__antithesis_instrumentation__.Notify(276084)
	}
	__antithesis_instrumentation__.Notify(276082)
	r.Root = op
	return nil
}

func (r opResult) wrapPostProcessSpec(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	args *colexecargs.NewColOperatorArgs,
	post *execinfrapb.PostProcessSpec,
	resultTypes []*types.T,
	factory coldata.ColumnFactory,
	causeToWrap error,
) error {
	__antithesis_instrumentation__.Notify(276085)
	noopSpec := &execinfrapb.ProcessorSpec{
		Core: execinfrapb.ProcessorCoreUnion{
			Noop: &execinfrapb.NoopCoreSpec{},
		},
		Post:        *post,
		ResultTypes: resultTypes,
	}
	inputToMaterializer := colexecargs.OpWithMetaInfo{Root: r.Root}
	takeOverMetaInfo(&inputToMaterializer, args.Inputs)

	return r.createAndWrapRowSource(
		ctx, flowCtx, args, []colexecargs.OpWithMetaInfo{inputToMaterializer},
		[][]*types.T{r.ColumnTypes}, noopSpec, factory, causeToWrap,
	)
}

func (r *postProcessResult) planPostProcessSpec(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	args *colexecargs.NewColOperatorArgs,
	post *execinfrapb.PostProcessSpec,
	factory coldata.ColumnFactory,
	releasables *[]execinfra.Releasable,
) error {
	__antithesis_instrumentation__.Notify(276086)
	if post.Projection {
		__antithesis_instrumentation__.Notify(276090)
		r.Op, r.ColumnTypes = addProjection(r.Op, r.ColumnTypes, post.OutputColumns)
	} else {
		__antithesis_instrumentation__.Notify(276091)
		if post.RenderExprs != nil {
			__antithesis_instrumentation__.Notify(276092)
			var renderedCols []uint32
			for _, renderExpr := range post.RenderExprs {
				__antithesis_instrumentation__.Notify(276095)
				expr, err := args.ExprHelper.ProcessExpr(renderExpr, flowCtx.EvalCtx, r.ColumnTypes)
				if err != nil {
					__antithesis_instrumentation__.Notify(276099)
					return err
				} else {
					__antithesis_instrumentation__.Notify(276100)
				}
				__antithesis_instrumentation__.Notify(276096)
				var outputIdx int
				r.Op, outputIdx, r.ColumnTypes, err = planProjectionOperators(
					ctx, flowCtx.EvalCtx, expr, r.ColumnTypes, r.Op, args.StreamingMemAccount, factory, releasables,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(276101)
					return errors.Wrapf(err, "unable to columnarize render expression %q", expr)
				} else {
					__antithesis_instrumentation__.Notify(276102)
				}
				__antithesis_instrumentation__.Notify(276097)
				if outputIdx < 0 {
					__antithesis_instrumentation__.Notify(276103)
					return errors.AssertionFailedf("missing outputIdx")
				} else {
					__antithesis_instrumentation__.Notify(276104)
				}
				__antithesis_instrumentation__.Notify(276098)
				renderedCols = append(renderedCols, uint32(outputIdx))
			}
			__antithesis_instrumentation__.Notify(276093)
			r.Op = colexecbase.NewSimpleProjectOp(r.Op, len(r.ColumnTypes), renderedCols)
			newTypes := make([]*types.T, len(renderedCols))
			for i, j := range renderedCols {
				__antithesis_instrumentation__.Notify(276105)
				newTypes[i] = r.ColumnTypes[j]
			}
			__antithesis_instrumentation__.Notify(276094)
			r.ColumnTypes = newTypes
		} else {
			__antithesis_instrumentation__.Notify(276106)
		}
	}
	__antithesis_instrumentation__.Notify(276087)
	if post.Offset != 0 {
		__antithesis_instrumentation__.Notify(276107)
		r.Op = colexec.NewOffsetOp(r.Op, post.Offset)
	} else {
		__antithesis_instrumentation__.Notify(276108)
	}
	__antithesis_instrumentation__.Notify(276088)
	if post.Limit != 0 {
		__antithesis_instrumentation__.Notify(276109)
		r.Op = colexec.NewLimitOp(r.Op, post.Limit)
	} else {
		__antithesis_instrumentation__.Notify(276110)
	}
	__antithesis_instrumentation__.Notify(276089)
	return nil
}

type postProcessResult struct {
	Op          colexecop.Operator
	ColumnTypes []*types.T
}

func (r opResult) updateWithPostProcessResult(ppr postProcessResult) {
	__antithesis_instrumentation__.Notify(276111)
	r.Root = ppr.Op
	r.ColumnTypes = make([]*types.T, len(ppr.ColumnTypes))
	copy(r.ColumnTypes, ppr.ColumnTypes)
}

func (r opResult) finishBufferedWindowerArgs(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	monitorRegistry *colexecargs.MonitorRegistry,
	args *colexecwindow.WindowArgs,
	opName string,
	processorID int32,
	factory coldata.ColumnFactory,
	needsBuffer bool,
) {
	__antithesis_instrumentation__.Notify(276112)
	args.DiskAcc = monitorRegistry.CreateDiskAccount(ctx, flowCtx, opName, processorID)
	mainAcc := monitorRegistry.CreateUnlimitedMemAccount(ctx, flowCtx, opName, processorID)
	args.MainAllocator = colmem.NewAllocator(ctx, mainAcc, factory)
	if needsBuffer {
		__antithesis_instrumentation__.Notify(276113)
		bufferAcc := monitorRegistry.CreateUnlimitedMemAccount(ctx, flowCtx, opName, processorID)
		args.BufferAllocator = colmem.NewAllocator(ctx, bufferAcc, factory)
	} else {
		__antithesis_instrumentation__.Notify(276114)
	}
}

func (r opResult) finishScanPlanning(op colfetcher.ScanOperator, resultTypes []*types.T) {
	__antithesis_instrumentation__.Notify(276115)
	r.Root = op
	if buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(276117)
		r.Root = colexec.NewInvariantsChecker(r.Root)
	} else {
		__antithesis_instrumentation__.Notify(276118)
	}
	__antithesis_instrumentation__.Notify(276116)
	r.KVReader = op
	r.MetadataSources = append(r.MetadataSources, r.Root.(colexecop.MetadataSource))
	r.Releasables = append(r.Releasables, op)

	r.Root = colexecutils.NewCancelChecker(r.Root)
	r.ColumnTypes = resultTypes
	r.ToClose = append(r.ToClose, op)
}

func planFilterExpr(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	input colexecop.Operator,
	columnTypes []*types.T,
	filter execinfrapb.Expression,
	acc *mon.BoundAccount,
	factory coldata.ColumnFactory,
	helper *colexecargs.ExprHelper,
	releasables *[]execinfra.Releasable,
) (colexecop.Operator, error) {
	__antithesis_instrumentation__.Notify(276119)
	expr, err := helper.ProcessExpr(filter, flowCtx.EvalCtx, columnTypes)
	if err != nil {
		__antithesis_instrumentation__.Notify(276124)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(276125)
	}
	__antithesis_instrumentation__.Notify(276120)
	if expr == tree.DNull {
		__antithesis_instrumentation__.Notify(276126)

		return colexecutils.NewZeroOp(input), nil
	} else {
		__antithesis_instrumentation__.Notify(276127)
	}
	__antithesis_instrumentation__.Notify(276121)
	op, _, filterColumnTypes, err := planSelectionOperators(
		ctx, flowCtx.EvalCtx, expr, columnTypes, input, acc, factory, releasables,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(276128)
		return nil, errors.Wrapf(err, "unable to columnarize filter expression %q", filter)
	} else {
		__antithesis_instrumentation__.Notify(276129)
	}
	__antithesis_instrumentation__.Notify(276122)
	if len(filterColumnTypes) > len(columnTypes) {
		__antithesis_instrumentation__.Notify(276130)

		var outputColumns []uint32
		for i := range columnTypes {
			__antithesis_instrumentation__.Notify(276132)
			outputColumns = append(outputColumns, uint32(i))
		}
		__antithesis_instrumentation__.Notify(276131)
		op = colexecbase.NewSimpleProjectOp(op, len(filterColumnTypes), outputColumns)
	} else {
		__antithesis_instrumentation__.Notify(276133)
	}
	__antithesis_instrumentation__.Notify(276123)
	return op, nil
}

func addProjection(
	op colexecop.Operator, typs []*types.T, projection []uint32,
) (colexecop.Operator, []*types.T) {
	__antithesis_instrumentation__.Notify(276134)
	newTypes := make([]*types.T, len(projection))
	for i, j := range projection {
		__antithesis_instrumentation__.Notify(276136)
		newTypes[i] = typs[j]
	}
	__antithesis_instrumentation__.Notify(276135)
	return colexecbase.NewSimpleProjectOp(op, len(typs), projection), newTypes
}

func planSelectionOperators(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	expr tree.TypedExpr,
	columnTypes []*types.T,
	input colexecop.Operator,
	acc *mon.BoundAccount,
	factory coldata.ColumnFactory,
	releasables *[]execinfra.Releasable,
) (op colexecop.Operator, resultIdx int, typs []*types.T, err error) {
	__antithesis_instrumentation__.Notify(276137)
	switch t := expr.(type) {
	case *tree.AndExpr:
		__antithesis_instrumentation__.Notify(276138)

		var leftOp, rightOp colexecop.Operator
		leftOp, _, typs, err = planSelectionOperators(
			ctx, evalCtx, t.TypedLeft(), columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276158)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276159)
		}
		__antithesis_instrumentation__.Notify(276139)
		rightOp, resultIdx, typs, err = planSelectionOperators(
			ctx, evalCtx, t.TypedRight(), typs, leftOp, acc, factory, releasables,
		)
		return rightOp, resultIdx, typs, err
	case *tree.CaseExpr:
		__antithesis_instrumentation__.Notify(276140)
		op, resultIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, expr, columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276160)
			return op, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276161)
		}
		__antithesis_instrumentation__.Notify(276141)
		op, err = colexecutils.BoolOrUnknownToSelOp(op, typs, resultIdx)
		return op, resultIdx, typs, err
	case *tree.ComparisonExpr:
		__antithesis_instrumentation__.Notify(276142)
		cmpOp := t.Operator
		leftOp, leftIdx, ct, err := planProjectionOperators(
			ctx, evalCtx, t.TypedLeft(), columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276162)
			return nil, resultIdx, ct, err
		} else {
			__antithesis_instrumentation__.Notify(276163)
		}
		__antithesis_instrumentation__.Notify(276143)
		lTyp := ct[leftIdx]
		if constArg, ok := t.Right.(tree.Datum); ok {
			__antithesis_instrumentation__.Notify(276164)
			switch cmpOp.Symbol {
			case treecmp.Like, treecmp.NotLike:
				__antithesis_instrumentation__.Notify(276167)
				negate := cmpOp.Symbol == treecmp.NotLike
				op, err = colexecsel.GetLikeOperator(
					evalCtx, leftOp, leftIdx, string(tree.MustBeDString(constArg)), negate,
				)
			case treecmp.In, treecmp.NotIn:
				__antithesis_instrumentation__.Notify(276168)
				negate := cmpOp.Symbol == treecmp.NotIn
				datumTuple, ok := tree.AsDTuple(constArg)
				if !ok || func() bool {
					__antithesis_instrumentation__.Notify(276173)
					return useDefaultCmpOpForIn(datumTuple) == true
				}() == true {
					__antithesis_instrumentation__.Notify(276174)
					break
				} else {
					__antithesis_instrumentation__.Notify(276175)
				}
				__antithesis_instrumentation__.Notify(276169)
				op, err = colexec.GetInOperator(evalCtx, lTyp, leftOp, leftIdx, datumTuple, negate)
			case treecmp.IsDistinctFrom, treecmp.IsNotDistinctFrom:
				__antithesis_instrumentation__.Notify(276170)
				if constArg != tree.DNull {
					__antithesis_instrumentation__.Notify(276176)

					break
				} else {
					__antithesis_instrumentation__.Notify(276177)
				}
				__antithesis_instrumentation__.Notify(276171)

				negate := cmpOp.Symbol == treecmp.IsDistinctFrom
				op = colexec.NewIsNullSelOp(leftOp, leftIdx, negate, false)
			default:
				__antithesis_instrumentation__.Notify(276172)
			}
			__antithesis_instrumentation__.Notify(276165)
			if op == nil || func() bool {
				__antithesis_instrumentation__.Notify(276178)
				return err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(276179)

				op, err = colexecsel.GetSelectionConstOperator(
					cmpOp, leftOp, ct, leftIdx, constArg, evalCtx, t,
				)
				if r, ok := op.(execinfra.Releasable); ok {
					__antithesis_instrumentation__.Notify(276180)
					*releasables = append(*releasables, r)
				} else {
					__antithesis_instrumentation__.Notify(276181)
				}
			} else {
				__antithesis_instrumentation__.Notify(276182)
			}
			__antithesis_instrumentation__.Notify(276166)
			return op, resultIdx, ct, err
		} else {
			__antithesis_instrumentation__.Notify(276183)
		}
		__antithesis_instrumentation__.Notify(276144)
		rightOp, rightIdx, ct, err := planProjectionOperators(
			ctx, evalCtx, t.TypedRight(), ct, leftOp, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276184)
			return nil, resultIdx, ct, err
		} else {
			__antithesis_instrumentation__.Notify(276185)
		}
		__antithesis_instrumentation__.Notify(276145)
		op, err = colexecsel.GetSelectionOperator(
			cmpOp, rightOp, ct, leftIdx, rightIdx, evalCtx, t,
		)
		if r, ok := op.(execinfra.Releasable); ok {
			__antithesis_instrumentation__.Notify(276186)
			*releasables = append(*releasables, r)
		} else {
			__antithesis_instrumentation__.Notify(276187)
		}
		__antithesis_instrumentation__.Notify(276146)
		return op, resultIdx, ct, err
	case *tree.IndexedVar:
		__antithesis_instrumentation__.Notify(276147)
		op, err = colexecutils.BoolOrUnknownToSelOp(input, columnTypes, t.Idx)
		return op, -1, columnTypes, err
	case *tree.IsNotNullExpr:
		__antithesis_instrumentation__.Notify(276148)
		op, resultIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, t.TypedInnerExpr(), columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276188)
			return op, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276189)
		}
		__antithesis_instrumentation__.Notify(276149)
		op = colexec.NewIsNullSelOp(
			op, resultIdx, true, typs[resultIdx].Family() == types.TupleFamily,
		)
		return op, resultIdx, typs, err
	case *tree.IsNullExpr:
		__antithesis_instrumentation__.Notify(276150)
		op, resultIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, t.TypedInnerExpr(), columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276190)
			return op, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276191)
		}
		__antithesis_instrumentation__.Notify(276151)
		op = colexec.NewIsNullSelOp(
			op, resultIdx, false, typs[resultIdx].Family() == types.TupleFamily,
		)
		return op, resultIdx, typs, err
	case *tree.NotExpr:
		__antithesis_instrumentation__.Notify(276152)
		op, resultIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, t.TypedInnerExpr(), columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276192)
			return op, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276193)
		}
		__antithesis_instrumentation__.Notify(276153)
		op, err = colexec.NewNotExprSelOp(t.TypedInnerExpr().ResolvedType().Family(), op, resultIdx)
		return op, resultIdx, typs, err
	case *tree.OrExpr:
		__antithesis_instrumentation__.Notify(276154)

		caseExpr, err := tree.NewTypedCaseExpr(
			nil,
			[]*tree.When{
				{Cond: t.Left, Val: tree.DBoolTrue},
				{Cond: t.Right, Val: tree.DBoolTrue},
			},
			tree.DBoolFalse,
			types.Bool)
		if err != nil {
			__antithesis_instrumentation__.Notify(276194)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276195)
		}
		__antithesis_instrumentation__.Notify(276155)
		op, resultIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, caseExpr, columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276196)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276197)
		}
		__antithesis_instrumentation__.Notify(276156)
		op, err = colexecutils.BoolOrUnknownToSelOp(op, typs, resultIdx)
		return op, resultIdx, typs, err
	default:
		__antithesis_instrumentation__.Notify(276157)
		return nil, resultIdx, nil, errors.Errorf("unhandled selection expression type: %s", reflect.TypeOf(t))
	}
}

func planCastOperator(
	ctx context.Context,
	acc *mon.BoundAccount,
	columnTypes []*types.T,
	input colexecop.Operator,
	inputIdx int,
	fromType *types.T,
	toType *types.T,
	factory coldata.ColumnFactory,
	evalCtx *tree.EvalContext,
) (op colexecop.Operator, resultIdx int, typs []*types.T, err error) {
	__antithesis_instrumentation__.Notify(276198)
	outputIdx := len(columnTypes)
	op, err = colexecbase.GetCastOperator(colmem.NewAllocator(ctx, acc, factory), input, inputIdx, outputIdx, fromType, toType, evalCtx)
	typs = appendOneType(columnTypes, toType)
	return op, outputIdx, typs, err
}

func planProjectionOperators(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	expr tree.TypedExpr,
	columnTypes []*types.T,
	input colexecop.Operator,
	acc *mon.BoundAccount,
	factory coldata.ColumnFactory,
	releasables *[]execinfra.Releasable,
) (op colexecop.Operator, resultIdx int, typs []*types.T, err error) {
	__antithesis_instrumentation__.Notify(276199)

	projectDatum := func(datum tree.Datum) (colexecop.Operator, error) {
		__antithesis_instrumentation__.Notify(276201)
		resultIdx = len(columnTypes)
		datumType := datum.ResolvedType()
		typs = appendOneType(columnTypes, datumType)
		if datumType.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(276203)

			return colexecbase.NewConstNullOp(colmem.NewAllocator(ctx, acc, factory), input, resultIdx), nil
		} else {
			__antithesis_instrumentation__.Notify(276204)
		}
		__antithesis_instrumentation__.Notify(276202)
		constVal := colconv.GetDatumToPhysicalFn(datumType)(datum)
		return colexecbase.NewConstOp(colmem.NewAllocator(ctx, acc, factory), input, datumType, constVal, resultIdx)
	}
	__antithesis_instrumentation__.Notify(276200)
	resultIdx = -1
	switch t := expr.(type) {
	case *tree.AndExpr:
		__antithesis_instrumentation__.Notify(276205)
		return planLogicalProjectionOp(ctx, evalCtx, expr, columnTypes, input, acc, factory, releasables)
	case *tree.BinaryExpr:
		__antithesis_instrumentation__.Notify(276206)
		if err = checkSupportedBinaryExpr(t.TypedLeft(), t.TypedRight(), t.ResolvedType()); err != nil {
			__antithesis_instrumentation__.Notify(276238)
			return op, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276239)
		}
		__antithesis_instrumentation__.Notify(276207)
		return planProjectionExpr(
			ctx, evalCtx, t.Operator, t.ResolvedType(), t.TypedLeft(), t.TypedRight(),
			columnTypes, input, acc, factory, t.Fn.Fn, nil, releasables,
		)
	case *tree.CaseExpr:
		__antithesis_instrumentation__.Notify(276208)
		allocator := colmem.NewAllocator(ctx, acc, factory)
		caseOutputType := t.ResolvedType()
		caseOutputIdx := len(columnTypes)

		schemaEnforcer := colexecutils.NewBatchSchemaSubsetEnforcer(
			allocator, input, nil, caseOutputIdx, -1,
		)
		buffer := colexec.NewBufferOp(schemaEnforcer)
		caseOps := make([]colexecop.Operator, len(t.Whens))
		typs = appendOneType(columnTypes, caseOutputType)
		thenIdxs := make([]int, len(t.Whens)+1)
		for i, when := range t.Whens {
			__antithesis_instrumentation__.Notify(276240)

			whenTyped := when.Cond.(tree.TypedExpr)
			caseOps[i], resultIdx, typs, err = planProjectionOperators(
				ctx, evalCtx, whenTyped, typs, buffer, acc, factory, releasables,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(276245)
				return nil, resultIdx, typs, err
			} else {
				__antithesis_instrumentation__.Notify(276246)
			}
			__antithesis_instrumentation__.Notify(276241)
			if t.Expr != nil {
				__antithesis_instrumentation__.Notify(276247)

				left := t.Expr.(tree.TypedExpr)

				right := tree.NewTypedOrdinalReference(resultIdx, whenTyped.ResolvedType())
				cmpExpr := tree.NewTypedComparisonExpr(treecmp.MakeComparisonOperator(treecmp.EQ), left, right)
				caseOps[i], resultIdx, typs, err = planProjectionOperators(
					ctx, evalCtx, cmpExpr, typs, caseOps[i], acc, factory, releasables,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(276248)
					return nil, resultIdx, typs, err
				} else {
					__antithesis_instrumentation__.Notify(276249)
				}
			} else {
				__antithesis_instrumentation__.Notify(276250)
			}
			__antithesis_instrumentation__.Notify(276242)
			caseOps[i], err = colexecutils.BoolOrUnknownToSelOp(caseOps[i], typs, resultIdx)
			if err != nil {
				__antithesis_instrumentation__.Notify(276251)
				return nil, resultIdx, typs, err
			} else {
				__antithesis_instrumentation__.Notify(276252)
			}
			__antithesis_instrumentation__.Notify(276243)

			caseOps[i], thenIdxs[i], typs, err = planProjectionOperators(
				ctx, evalCtx, when.Val.(tree.TypedExpr), typs, caseOps[i], acc, factory, releasables,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(276253)
				return nil, resultIdx, typs, err
			} else {
				__antithesis_instrumentation__.Notify(276254)
			}
			__antithesis_instrumentation__.Notify(276244)
			if !typs[thenIdxs[i]].Identical(typs[caseOutputIdx]) {
				__antithesis_instrumentation__.Notify(276255)

				fromType, toType := typs[thenIdxs[i]], typs[caseOutputIdx]
				caseOps[i], thenIdxs[i], typs, err = planCastOperator(
					ctx, acc, typs, caseOps[i], thenIdxs[i], fromType, toType, factory, evalCtx,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(276256)
					return nil, resultIdx, typs, err
				} else {
					__antithesis_instrumentation__.Notify(276257)
				}
			} else {
				__antithesis_instrumentation__.Notify(276258)
			}
		}
		__antithesis_instrumentation__.Notify(276209)
		var elseOp colexecop.Operator
		elseExpr := t.Else
		if elseExpr == nil {
			__antithesis_instrumentation__.Notify(276259)

			elseExpr = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(276260)
		}
		__antithesis_instrumentation__.Notify(276210)
		elseOp, thenIdxs[len(t.Whens)], typs, err = planProjectionOperators(
			ctx, evalCtx, elseExpr.(tree.TypedExpr), typs, buffer, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276261)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276262)
		}
		__antithesis_instrumentation__.Notify(276211)
		if !typs[thenIdxs[len(t.Whens)]].Identical(typs[caseOutputIdx]) {
			__antithesis_instrumentation__.Notify(276263)

			elseIdx := thenIdxs[len(t.Whens)]
			fromType, toType := typs[elseIdx], typs[caseOutputIdx]
			elseOp, thenIdxs[len(t.Whens)], typs, err = planCastOperator(
				ctx, acc, typs, elseOp, elseIdx, fromType, toType, factory, evalCtx,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(276264)
				return nil, resultIdx, typs, err
			} else {
				__antithesis_instrumentation__.Notify(276265)
			}
		} else {
			__antithesis_instrumentation__.Notify(276266)
		}
		__antithesis_instrumentation__.Notify(276212)

		schemaEnforcer.SetTypes(typs)
		op = colexec.NewCaseOp(allocator, buffer, caseOps, elseOp, thenIdxs, caseOutputIdx, caseOutputType)
		return op, caseOutputIdx, typs, err
	case *tree.CastExpr:
		__antithesis_instrumentation__.Notify(276213)
		expr := t.Expr.(tree.TypedExpr)
		op, resultIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, expr, columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276267)
			return nil, 0, nil, err
		} else {
			__antithesis_instrumentation__.Notify(276268)
		}
		__antithesis_instrumentation__.Notify(276214)
		op, resultIdx, typs, err = planCastOperator(ctx, acc, typs, op, resultIdx, expr.ResolvedType(), t.ResolvedType(), factory, evalCtx)
		return op, resultIdx, typs, err
	case *tree.CoalesceExpr:
		__antithesis_instrumentation__.Notify(276215)

		whens := make([]*tree.When, len(t.Exprs))
		for i := range whens {
			__antithesis_instrumentation__.Notify(276269)
			whens[i] = &tree.When{
				Cond: tree.NewTypedComparisonExpr(
					treecmp.MakeComparisonOperator(treecmp.IsDistinctFrom),
					t.Exprs[i].(tree.TypedExpr),
					tree.DNull,
				),
				Val: t.Exprs[i],
			}
		}
		__antithesis_instrumentation__.Notify(276216)
		caseExpr, err := tree.NewTypedCaseExpr(
			nil,
			whens,
			nil,
			t.ResolvedType(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276270)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276271)
		}
		__antithesis_instrumentation__.Notify(276217)
		return planProjectionOperators(ctx, evalCtx, caseExpr, columnTypes, input, acc, factory, releasables)
	case *tree.ComparisonExpr:
		__antithesis_instrumentation__.Notify(276218)
		return planProjectionExpr(
			ctx, evalCtx, t.Operator, t.ResolvedType(), t.TypedLeft(), t.TypedRight(),
			columnTypes, input, acc, factory, nil, t, releasables,
		)
	case tree.Datum:
		__antithesis_instrumentation__.Notify(276219)
		op, err = projectDatum(t)
		return op, resultIdx, typs, err
	case *tree.FuncExpr:
		__antithesis_instrumentation__.Notify(276220)
		var inputCols []int
		typs = make([]*types.T, len(columnTypes))
		copy(typs, columnTypes)
		op = input
		for _, e := range t.Exprs {
			__antithesis_instrumentation__.Notify(276272)
			var err error

			op, resultIdx, typs, err = planProjectionOperators(
				ctx, evalCtx, e.(tree.TypedExpr), typs, op, acc, factory, releasables,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(276274)
				return nil, resultIdx, nil, err
			} else {
				__antithesis_instrumentation__.Notify(276275)
			}
			__antithesis_instrumentation__.Notify(276273)
			inputCols = append(inputCols, resultIdx)
		}
		__antithesis_instrumentation__.Notify(276221)
		resultIdx = len(typs)
		op, err = colexec.NewBuiltinFunctionOperator(
			colmem.NewAllocator(ctx, acc, factory), evalCtx, t, typs, inputCols, resultIdx, op,
		)
		if r, ok := op.(execinfra.Releasable); ok {
			__antithesis_instrumentation__.Notify(276276)
			*releasables = append(*releasables, r)
		} else {
			__antithesis_instrumentation__.Notify(276277)
		}
		__antithesis_instrumentation__.Notify(276222)
		typs = appendOneType(typs, t.ResolvedType())
		return op, resultIdx, typs, err
	case *tree.IfExpr:
		__antithesis_instrumentation__.Notify(276223)

		caseExpr, err := tree.NewTypedCaseExpr(
			nil,
			[]*tree.When{{Cond: t.Cond, Val: t.True}},
			t.TypedElseExpr(),
			t.ResolvedType(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276278)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276279)
		}
		__antithesis_instrumentation__.Notify(276224)
		return planProjectionOperators(ctx, evalCtx, caseExpr, columnTypes, input, acc, factory, releasables)
	case *tree.IndexedVar:
		__antithesis_instrumentation__.Notify(276225)
		return input, t.Idx, columnTypes, nil
	case *tree.IsNotNullExpr:
		__antithesis_instrumentation__.Notify(276226)
		return planIsNullProjectionOp(ctx, evalCtx, t.ResolvedType(), t.TypedInnerExpr(), columnTypes, input, acc, true, factory, releasables)
	case *tree.IsNullExpr:
		__antithesis_instrumentation__.Notify(276227)
		return planIsNullProjectionOp(ctx, evalCtx, t.ResolvedType(), t.TypedInnerExpr(), columnTypes, input, acc, false, factory, releasables)
	case *tree.NotExpr:
		__antithesis_instrumentation__.Notify(276228)
		op, resultIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, t.TypedInnerExpr(), columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276280)
			return op, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276281)
		}
		__antithesis_instrumentation__.Notify(276229)
		outputIdx, allocator := len(typs), colmem.NewAllocator(ctx, acc, factory)
		op, err = colexec.NewNotExprProjOp(
			t.TypedInnerExpr().ResolvedType().Family(), allocator, op, resultIdx, outputIdx,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276282)
			return op, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276283)
		}
		__antithesis_instrumentation__.Notify(276230)
		typs = appendOneType(typs, t.ResolvedType())
		return op, outputIdx, typs, nil
	case *tree.NullIfExpr:
		__antithesis_instrumentation__.Notify(276231)

		caseExpr, err := tree.NewTypedCaseExpr(
			nil,
			[]*tree.When{{
				Cond: tree.NewTypedComparisonExpr(
					treecmp.MakeComparisonOperator(treecmp.EQ),
					t.Expr1.(tree.TypedExpr),
					t.Expr2.(tree.TypedExpr),
				),
				Val: tree.DNull,
			}},
			t.Expr1.(tree.TypedExpr),
			t.ResolvedType(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276284)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276285)
		}
		__antithesis_instrumentation__.Notify(276232)
		return planProjectionOperators(ctx, evalCtx, caseExpr, columnTypes, input, acc, factory, releasables)
	case *tree.OrExpr:
		__antithesis_instrumentation__.Notify(276233)
		return planLogicalProjectionOp(ctx, evalCtx, expr, columnTypes, input, acc, factory, releasables)
	case *tree.Tuple:
		__antithesis_instrumentation__.Notify(276234)
		tuple, isConstTuple := evalTupleIfConst(evalCtx, t)
		if isConstTuple {
			__antithesis_instrumentation__.Notify(276286)

			op, err = projectDatum(tuple)
			return op, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276287)
		}
		__antithesis_instrumentation__.Notify(276235)
		outputType := t.ResolvedType()
		typs = make([]*types.T, len(columnTypes))
		copy(typs, columnTypes)
		tupleContentsIdxs := make([]int, len(t.Exprs))
		for i, expr := range t.Exprs {
			__antithesis_instrumentation__.Notify(276288)
			input, tupleContentsIdxs[i], typs, err = planProjectionOperators(
				ctx, evalCtx, expr.(tree.TypedExpr), typs, input, acc, factory, releasables,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(276289)
				return nil, resultIdx, typs, err
			} else {
				__antithesis_instrumentation__.Notify(276290)
			}
		}
		__antithesis_instrumentation__.Notify(276236)
		resultIdx = len(typs)
		op = colexec.NewTupleProjOp(
			colmem.NewAllocator(ctx, acc, factory), typs, tupleContentsIdxs, outputType, input, resultIdx,
		)
		*releasables = append(*releasables, op.(execinfra.Releasable))
		typs = appendOneType(typs, outputType)
		return op, resultIdx, typs, err
	default:
		__antithesis_instrumentation__.Notify(276237)
		return nil, resultIdx, nil, errors.Errorf("unhandled projection expression type: %s", reflect.TypeOf(t))
	}
}

func checkSupportedProjectionExpr(left, right tree.TypedExpr) error {
	__antithesis_instrumentation__.Notify(276291)
	leftTyp := left.ResolvedType()
	rightTyp := right.ResolvedType()
	if leftTyp.Equivalent(rightTyp) || func() bool {
		__antithesis_instrumentation__.Notify(276294)
		return leftTyp.Family() == types.UnknownFamily == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(276295)
		return rightTyp.Family() == types.UnknownFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(276296)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(276297)
	}
	__antithesis_instrumentation__.Notify(276292)

	for _, t := range []*types.T{leftTyp, rightTyp} {
		__antithesis_instrumentation__.Notify(276298)
		switch t.Family() {
		case types.DateFamily, types.TimestampFamily, types.TimestampTZFamily:
			__antithesis_instrumentation__.Notify(276299)
			return errors.New("dates and timestamp(tz) not supported in mixed-type expressions in the vectorized engine")
		default:
			__antithesis_instrumentation__.Notify(276300)
		}
	}
	__antithesis_instrumentation__.Notify(276293)
	return nil
}

func checkSupportedBinaryExpr(left, right tree.TypedExpr, outputType *types.T) error {
	__antithesis_instrumentation__.Notify(276301)
	leftDatumBacked := typeconv.TypeFamilyToCanonicalTypeFamily(left.ResolvedType().Family()) == typeconv.DatumVecCanonicalTypeFamily
	rightDatumBacked := typeconv.TypeFamilyToCanonicalTypeFamily(right.ResolvedType().Family()) == typeconv.DatumVecCanonicalTypeFamily
	outputDatumBacked := typeconv.TypeFamilyToCanonicalTypeFamily(outputType.Family()) == typeconv.DatumVecCanonicalTypeFamily
	if (leftDatumBacked && func() bool {
		__antithesis_instrumentation__.Notify(276303)
		return rightDatumBacked == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(276304)
		return !outputDatumBacked == true
	}() == true {
		__antithesis_instrumentation__.Notify(276305)
		return errors.New("datum-backed arguments on both sides and not datum-backed " +
			"output of a binary expression is currently not supported")
	} else {
		__antithesis_instrumentation__.Notify(276306)
	}
	__antithesis_instrumentation__.Notify(276302)
	return nil
}

func planProjectionExpr(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	projOp tree.Operator,
	outputType *types.T,
	left, right tree.TypedExpr,
	columnTypes []*types.T,
	input colexecop.Operator,
	acc *mon.BoundAccount,
	factory coldata.ColumnFactory,
	binFn tree.TwoArgFn,
	cmpExpr *tree.ComparisonExpr,
	releasables *[]execinfra.Releasable,
) (op colexecop.Operator, resultIdx int, typs []*types.T, err error) {
	__antithesis_instrumentation__.Notify(276307)
	if err := checkSupportedProjectionExpr(left, right); err != nil {
		__antithesis_instrumentation__.Notify(276313)
		return nil, resultIdx, typs, err
	} else {
		__antithesis_instrumentation__.Notify(276314)
	}
	__antithesis_instrumentation__.Notify(276308)
	allocator := colmem.NewAllocator(ctx, acc, factory)
	resultIdx = -1

	cmpProjOp, isCmpProjOp := projOp.(treecmp.ComparisonOperator)
	var hasOptimizedOp bool
	if isCmpProjOp {
		__antithesis_instrumentation__.Notify(276315)
		switch cmpProjOp.Symbol {
		case treecmp.Like, treecmp.NotLike, treecmp.In, treecmp.NotIn, treecmp.IsDistinctFrom, treecmp.IsNotDistinctFrom:
			__antithesis_instrumentation__.Notify(276316)
			hasOptimizedOp = true
		default:
			__antithesis_instrumentation__.Notify(276317)
		}
	} else {
		__antithesis_instrumentation__.Notify(276318)
	}
	__antithesis_instrumentation__.Notify(276309)

	if lConstArg, lConst := left.(tree.Datum); lConst && func() bool {
		__antithesis_instrumentation__.Notify(276319)
		return !hasOptimizedOp == true
	}() == true {
		__antithesis_instrumentation__.Notify(276320)

		var rightIdx int
		input, rightIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, right, columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276322)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276323)
		}
		__antithesis_instrumentation__.Notify(276321)
		resultIdx = len(typs)

		op, err = colexecproj.GetProjectionLConstOperator(
			allocator, typs, left.ResolvedType(), outputType, projOp, input,
			rightIdx, lConstArg, resultIdx, evalCtx, binFn, cmpExpr,
		)
	} else {
		__antithesis_instrumentation__.Notify(276324)
		var leftIdx int
		input, leftIdx, typs, err = planProjectionOperators(
			ctx, evalCtx, left, columnTypes, input, acc, factory, releasables,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(276328)
			return nil, resultIdx, typs, err
		} else {
			__antithesis_instrumentation__.Notify(276329)
		}
		__antithesis_instrumentation__.Notify(276325)

		if tuple, ok := right.(*tree.Tuple); ok {
			__antithesis_instrumentation__.Notify(276330)

			tupleDatum, ok := evalTupleIfConst(evalCtx, tuple)
			if ok {
				__antithesis_instrumentation__.Notify(276331)
				right = tupleDatum
			} else {
				__antithesis_instrumentation__.Notify(276332)
			}
		} else {
			__antithesis_instrumentation__.Notify(276333)
		}
		__antithesis_instrumentation__.Notify(276326)

		if isCmpProjOp && func() bool {
			__antithesis_instrumentation__.Notify(276334)
			return (cmpProjOp.Symbol == treecmp.IsDistinctFrom || func() bool {
				__antithesis_instrumentation__.Notify(276335)
				return cmpProjOp.Symbol == treecmp.IsNotDistinctFrom == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(276336)
			if cast, ok := right.(*tree.CastExpr); ok {
				__antithesis_instrumentation__.Notify(276337)
				if cast.Expr == tree.DNull {
					__antithesis_instrumentation__.Notify(276338)
					right = tree.DNull
				} else {
					__antithesis_instrumentation__.Notify(276339)
				}
			} else {
				__antithesis_instrumentation__.Notify(276340)
			}
		} else {
			__antithesis_instrumentation__.Notify(276341)
		}
		__antithesis_instrumentation__.Notify(276327)
		if rConstArg, rConst := right.(tree.Datum); rConst {
			__antithesis_instrumentation__.Notify(276342)

			resultIdx = len(typs)
			if isCmpProjOp {
				__antithesis_instrumentation__.Notify(276344)
				switch cmpProjOp.Symbol {
				case treecmp.Like, treecmp.NotLike:
					__antithesis_instrumentation__.Notify(276345)
					negate := cmpProjOp.Symbol == treecmp.NotLike
					op, err = colexecproj.GetLikeProjectionOperator(
						allocator, evalCtx, input, leftIdx, resultIdx,
						string(tree.MustBeDString(rConstArg)), negate,
					)
				case treecmp.In, treecmp.NotIn:
					__antithesis_instrumentation__.Notify(276346)
					negate := cmpProjOp.Symbol == treecmp.NotIn
					datumTuple, ok := tree.AsDTuple(rConstArg)
					if !ok || func() bool {
						__antithesis_instrumentation__.Notify(276351)
						return useDefaultCmpOpForIn(datumTuple) == true
					}() == true {
						__antithesis_instrumentation__.Notify(276352)
						break
					} else {
						__antithesis_instrumentation__.Notify(276353)
					}
					__antithesis_instrumentation__.Notify(276347)
					op, err = colexec.GetInProjectionOperator(
						evalCtx, allocator, typs[leftIdx], input, leftIdx, resultIdx, datumTuple, negate,
					)
				case treecmp.IsDistinctFrom, treecmp.IsNotDistinctFrom:
					__antithesis_instrumentation__.Notify(276348)
					if right != tree.DNull {
						__antithesis_instrumentation__.Notify(276354)

						break
					} else {
						__antithesis_instrumentation__.Notify(276355)
					}
					__antithesis_instrumentation__.Notify(276349)

					negate := cmpProjOp.Symbol == treecmp.IsDistinctFrom
					op = colexec.NewIsNullProjOp(
						allocator, input, leftIdx, resultIdx, negate, false,
					)
				default:
					__antithesis_instrumentation__.Notify(276350)
				}
			} else {
				__antithesis_instrumentation__.Notify(276356)
			}
			__antithesis_instrumentation__.Notify(276343)
			if op == nil || func() bool {
				__antithesis_instrumentation__.Notify(276357)
				return err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(276358)

				op, err = colexecproj.GetProjectionRConstOperator(
					allocator, typs, right.ResolvedType(), outputType, projOp,
					input, leftIdx, rConstArg, resultIdx, evalCtx, binFn, cmpExpr,
				)
			} else {
				__antithesis_instrumentation__.Notify(276359)
			}
		} else {
			__antithesis_instrumentation__.Notify(276360)

			var rightIdx int
			input, rightIdx, typs, err = planProjectionOperators(
				ctx, evalCtx, right, typs, input, acc, factory, releasables,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(276362)
				return nil, resultIdx, nil, err
			} else {
				__antithesis_instrumentation__.Notify(276363)
			}
			__antithesis_instrumentation__.Notify(276361)
			resultIdx = len(typs)
			op, err = colexecproj.GetProjectionOperator(
				allocator, typs, outputType, projOp, input, leftIdx, rightIdx,
				resultIdx, evalCtx, binFn, cmpExpr,
			)
		}
	}
	__antithesis_instrumentation__.Notify(276310)
	if err != nil {
		__antithesis_instrumentation__.Notify(276364)
		return op, resultIdx, typs, err
	} else {
		__antithesis_instrumentation__.Notify(276365)
	}
	__antithesis_instrumentation__.Notify(276311)
	if r, ok := op.(execinfra.Releasable); ok {
		__antithesis_instrumentation__.Notify(276366)
		*releasables = append(*releasables, r)
	} else {
		__antithesis_instrumentation__.Notify(276367)
	}
	__antithesis_instrumentation__.Notify(276312)
	typs = appendOneType(typs, outputType)
	return op, resultIdx, typs, err
}

func planLogicalProjectionOp(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	expr tree.TypedExpr,
	columnTypes []*types.T,
	input colexecop.Operator,
	acc *mon.BoundAccount,
	factory coldata.ColumnFactory,
	releasables *[]execinfra.Releasable,
) (op colexecop.Operator, resultIdx int, typs []*types.T, err error) {
	__antithesis_instrumentation__.Notify(276368)

	resultIdx = len(columnTypes)
	typs = appendOneType(columnTypes, types.Bool)
	var (
		typedLeft, typedRight             tree.TypedExpr
		leftProjOpChain, rightProjOpChain colexecop.Operator
		leftIdx, rightIdx                 int
	)
	leftFeedOp := colexecop.NewFeedOperator()
	rightFeedOp := colexecop.NewFeedOperator()
	switch t := expr.(type) {
	case *tree.AndExpr:
		__antithesis_instrumentation__.Notify(276373)
		typedLeft = t.TypedLeft()
		typedRight = t.TypedRight()
	case *tree.OrExpr:
		__antithesis_instrumentation__.Notify(276374)
		typedLeft = t.TypedLeft()
		typedRight = t.TypedRight()
	default:
		__antithesis_instrumentation__.Notify(276375)
		colexecerror.InternalError(errors.AssertionFailedf("unexpected logical expression type %s", t.String()))
	}
	__antithesis_instrumentation__.Notify(276369)
	leftProjOpChain, leftIdx, typs, err = planProjectionOperators(
		ctx, evalCtx, typedLeft, typs, leftFeedOp, acc, factory, releasables,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(276376)
		return nil, resultIdx, typs, err
	} else {
		__antithesis_instrumentation__.Notify(276377)
	}
	__antithesis_instrumentation__.Notify(276370)
	rightProjOpChain, rightIdx, typs, err = planProjectionOperators(
		ctx, evalCtx, typedRight, typs, rightFeedOp, acc, factory, releasables,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(276378)
		return nil, resultIdx, typs, err
	} else {
		__antithesis_instrumentation__.Notify(276379)
	}
	__antithesis_instrumentation__.Notify(276371)
	allocator := colmem.NewAllocator(ctx, acc, factory)
	input = colexecutils.NewBatchSchemaSubsetEnforcer(allocator, input, typs, resultIdx, len(typs))
	switch expr.(type) {
	case *tree.AndExpr:
		__antithesis_instrumentation__.Notify(276380)
		op, err = colexec.NewAndProjOp(
			allocator,
			input, leftProjOpChain, rightProjOpChain,
			leftFeedOp, rightFeedOp,
			typs[leftIdx], typs[rightIdx],
			leftIdx, rightIdx, resultIdx,
		)
	case *tree.OrExpr:
		__antithesis_instrumentation__.Notify(276381)
		op, err = colexec.NewOrProjOp(
			allocator,
			input, leftProjOpChain, rightProjOpChain,
			leftFeedOp, rightFeedOp,
			typs[leftIdx], typs[rightIdx],
			leftIdx, rightIdx, resultIdx,
		)
	}
	__antithesis_instrumentation__.Notify(276372)
	return op, resultIdx, typs, err
}

func planIsNullProjectionOp(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	outputType *types.T,
	expr tree.TypedExpr,
	columnTypes []*types.T,
	input colexecop.Operator,
	acc *mon.BoundAccount,
	negate bool,
	factory coldata.ColumnFactory,
	releasables *[]execinfra.Releasable,
) (op colexecop.Operator, resultIdx int, typs []*types.T, err error) {
	__antithesis_instrumentation__.Notify(276382)
	op, resultIdx, typs, err = planProjectionOperators(
		ctx, evalCtx, expr, columnTypes, input, acc, factory, releasables,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(276384)
		return op, resultIdx, typs, err
	} else {
		__antithesis_instrumentation__.Notify(276385)
	}
	__antithesis_instrumentation__.Notify(276383)
	outputIdx := len(typs)
	isTupleNull := typs[resultIdx].Family() == types.TupleFamily
	op = colexec.NewIsNullProjOp(
		colmem.NewAllocator(ctx, acc, factory), op, resultIdx, outputIdx, negate, isTupleNull,
	)
	typs = appendOneType(typs, outputType)
	return op, outputIdx, typs, nil
}

func appendOneType(typs []*types.T, t *types.T) []*types.T {
	__antithesis_instrumentation__.Notify(276386)
	newTyps := make([]*types.T, len(typs)+1)
	copy(newTyps, typs)
	newTyps[len(newTyps)-1] = t
	return newTyps
}

func useDefaultCmpOpForIn(tuple *tree.DTuple) bool {
	__antithesis_instrumentation__.Notify(276387)
	tupleContents := tuple.ResolvedType().TupleContents()
	if len(tupleContents) == 0 {
		__antithesis_instrumentation__.Notify(276390)
		return true
	} else {
		__antithesis_instrumentation__.Notify(276391)
	}
	__antithesis_instrumentation__.Notify(276388)
	for _, typ := range tupleContents {
		__antithesis_instrumentation__.Notify(276392)
		if typ.Family() == types.TupleFamily {
			__antithesis_instrumentation__.Notify(276393)
			return true
		} else {
			__antithesis_instrumentation__.Notify(276394)
		}
	}
	__antithesis_instrumentation__.Notify(276389)
	return false
}

func evalTupleIfConst(evalCtx *tree.EvalContext, t *tree.Tuple) (_ tree.Datum, ok bool) {
	__antithesis_instrumentation__.Notify(276395)
	for _, expr := range t.Exprs {
		__antithesis_instrumentation__.Notify(276398)
		if _, isDatum := expr.(tree.Datum); !isDatum {
			__antithesis_instrumentation__.Notify(276399)

			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(276400)
		}
	}
	__antithesis_instrumentation__.Notify(276396)
	tuple, err := t.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(276401)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(276402)
	}
	__antithesis_instrumentation__.Notify(276397)
	return tuple, true
}
