package colflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"reflect"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/treeprinter"
	"github.com/cockroachdb/errors"
)

func convertToVecTree(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	flow *execinfrapb.FlowSpec,
	localProcessors []execinfra.LocalProcessor,
	isPlanLocal bool,
) (opChains execinfra.OpChains, cleanup func(), err error) {
	__antithesis_instrumentation__.Notify(456029)
	if !isPlanLocal && func() bool {
		__antithesis_instrumentation__.Notify(456032)
		return len(localProcessors) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(456033)
		return nil, func() { __antithesis_instrumentation__.Notify(456034) }, errors.AssertionFailedf("unexpectedly non-empty LocalProcessors when plan is not local")
	} else {
		__antithesis_instrumentation__.Notify(456035)
	}
	__antithesis_instrumentation__.Notify(456030)
	fuseOpt := flowinfra.FuseNormally
	if isPlanLocal && func() bool {
		__antithesis_instrumentation__.Notify(456036)
		return !execinfra.HasParallelProcessors(flow) == true
	}() == true {
		__antithesis_instrumentation__.Notify(456037)
		fuseOpt = flowinfra.FuseAggressively
	} else {
		__antithesis_instrumentation__.Notify(456038)
	}
	__antithesis_instrumentation__.Notify(456031)

	creator := newVectorizedFlowCreator(
		newNoopFlowCreatorHelper(), vectorizedRemoteComponentCreator{}, false, false,
		nil, &execinfra.RowChannel{}, &fakeBatchReceiver{}, flowCtx.Cfg.PodNodeDialer, execinfrapb.FlowID{}, colcontainer.DiskQueueCfg{},
		flowCtx.Cfg.VecFDSemaphore, flowCtx.NewTypeResolver(flowCtx.EvalCtx.Txn),
		admission.WorkInfo{},
	)

	memoryMonitor := mon.NewMonitor(
		"convert-to-vec-tree",
		mon.MemoryResource,
		nil,
		nil,
		-1,
		math.MaxInt64,
		flowCtx.Cfg.Settings,
	)
	memoryMonitor.Start(ctx, nil, mon.MakeStandaloneBudget(math.MaxInt64))
	defer memoryMonitor.Stop(ctx)
	defer creator.cleanup(ctx)
	opChains, _, err = creator.setupFlow(ctx, flowCtx, flow.Processors, localProcessors, fuseOpt)
	return opChains, creator.Release, err
}

type fakeBatchReceiver struct{}

var _ execinfra.BatchReceiver = &fakeBatchReceiver{}

func (f fakeBatchReceiver) ProducerDone() { __antithesis_instrumentation__.Notify(456039) }

func (f fakeBatchReceiver) PushBatch(
	coldata.Batch, *execinfrapb.ProducerMetadata,
) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(456040)
	return execinfra.ConsumerClosed
}

type flowWithNode struct {
	sqlInstanceID base.SQLInstanceID
	flow          *execinfrapb.FlowSpec
}

func ExplainVec(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	flows map[base.SQLInstanceID]*execinfrapb.FlowSpec,
	localProcessors []execinfra.LocalProcessor,
	opChains execinfra.OpChains,
	gatewaySQLInstanceID base.SQLInstanceID,
	verbose bool,
	distributed bool,
) (_ []string, cleanup func(), _ error) {
	__antithesis_instrumentation__.Notify(456041)
	tp := treeprinter.NewWithStyle(treeprinter.CompactStyle)
	root := tp.Child("│")
	var (
		cleanups      []func()
		err           error
		conversionErr error
	)
	defer func() {
		__antithesis_instrumentation__.Notify(456045)
		cleanup = func() {
			__antithesis_instrumentation__.Notify(456046)
			for _, c := range cleanups {
				__antithesis_instrumentation__.Notify(456047)
				c()
			}
		}
	}()
	__antithesis_instrumentation__.Notify(456042)

	if err = colexecerror.CatchVectorizedRuntimeError(func() {
		__antithesis_instrumentation__.Notify(456048)
		if opChains != nil {
			__antithesis_instrumentation__.Notify(456049)
			formatChains(root, gatewaySQLInstanceID, opChains, verbose)
		} else {
			__antithesis_instrumentation__.Notify(456050)
			sortedFlows := make([]flowWithNode, 0, len(flows))
			for nodeID, flow := range flows {
				__antithesis_instrumentation__.Notify(456053)
				sortedFlows = append(sortedFlows, flowWithNode{sqlInstanceID: nodeID, flow: flow})
			}
			__antithesis_instrumentation__.Notify(456051)

			sort.Slice(sortedFlows, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(456054)
				return sortedFlows[i].sqlInstanceID < sortedFlows[j].sqlInstanceID
			})
			__antithesis_instrumentation__.Notify(456052)
			for _, flow := range sortedFlows {
				__antithesis_instrumentation__.Notify(456055)
				opChains, cleanup, err = convertToVecTree(ctx, flowCtx, flow.flow, localProcessors, !distributed)
				cleanups = append(cleanups, cleanup)
				if err != nil {
					__antithesis_instrumentation__.Notify(456057)
					conversionErr = err
					return
				} else {
					__antithesis_instrumentation__.Notify(456058)
				}
				__antithesis_instrumentation__.Notify(456056)
				formatChains(root, flow.sqlInstanceID, opChains, verbose)
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(456059)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(456060)
	}
	__antithesis_instrumentation__.Notify(456043)
	if conversionErr != nil {
		__antithesis_instrumentation__.Notify(456061)
		return nil, nil, conversionErr
	} else {
		__antithesis_instrumentation__.Notify(456062)
	}
	__antithesis_instrumentation__.Notify(456044)
	return tp.FormattedRows(), nil, nil
}

func formatChains(
	root treeprinter.Node,
	sqlInstanceID base.SQLInstanceID,
	opChains execinfra.OpChains,
	verbose bool,
) {
	__antithesis_instrumentation__.Notify(456063)
	node := root.Childf("Node %d", sqlInstanceID)
	for _, op := range opChains {
		__antithesis_instrumentation__.Notify(456064)
		formatOpChain(op, node, verbose)
	}
}

func shouldOutput(operator execinfra.OpNode, verbose bool) bool {
	__antithesis_instrumentation__.Notify(456065)
	_, nonExplainable := operator.(colexecop.NonExplainable)
	return !nonExplainable || func() bool {
		__antithesis_instrumentation__.Notify(456066)
		return verbose == true
	}() == true
}

func formatOpChain(operator execinfra.OpNode, node treeprinter.Node, verbose bool) {
	__antithesis_instrumentation__.Notify(456067)
	seenOps := make(map[reflect.Value]struct{})
	if shouldOutput(operator, verbose) {
		__antithesis_instrumentation__.Notify(456068)
		doFormatOpChain(operator, node.Child(reflect.TypeOf(operator).String()), verbose, seenOps)
	} else {
		__antithesis_instrumentation__.Notify(456069)
		doFormatOpChain(operator, node, verbose, seenOps)
	}
}
func doFormatOpChain(
	operator execinfra.OpNode,
	node treeprinter.Node,
	verbose bool,
	seenOps map[reflect.Value]struct{},
) {
	__antithesis_instrumentation__.Notify(456070)
	for i := 0; i < operator.ChildCount(verbose); i++ {
		__antithesis_instrumentation__.Notify(456071)
		child := operator.Child(i, verbose)
		childOpValue := reflect.ValueOf(child)
		childOpName := reflect.TypeOf(child).String()
		if _, seenOp := seenOps[childOpValue]; seenOp {
			__antithesis_instrumentation__.Notify(456073)

			node.Child(childOpName)
			continue
		} else {
			__antithesis_instrumentation__.Notify(456074)
		}
		__antithesis_instrumentation__.Notify(456072)
		seenOps[childOpValue] = struct{}{}
		if shouldOutput(child, verbose) {
			__antithesis_instrumentation__.Notify(456075)
			doFormatOpChain(child, node.Child(childOpName), verbose, seenOps)
		} else {
			__antithesis_instrumentation__.Notify(456076)
			doFormatOpChain(child, node, verbose, seenOps)
		}
	}
}
