package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type windowPlanState struct {
	infos   []*windowFuncInfo
	n       *windowNode
	planCtx *PlanningCtx
	plan    *PhysicalPlan
}

func createWindowPlanState(
	n *windowNode, planCtx *PlanningCtx, plan *PhysicalPlan,
) *windowPlanState {
	__antithesis_instrumentation__.Notify(467801)
	infos := make([]*windowFuncInfo, 0, len(n.funcs))
	for _, holder := range n.funcs {
		__antithesis_instrumentation__.Notify(467803)
		infos = append(infos, &windowFuncInfo{holder: holder})
	}
	__antithesis_instrumentation__.Notify(467802)
	return &windowPlanState{
		infos:   infos,
		n:       n,
		planCtx: planCtx,
		plan:    plan,
	}
}

type windowFuncInfo struct {
	holder *windowFuncHolder

	isProcessed bool
}

func (s *windowPlanState) findUnprocessedWindowFnsWithSamePartition() (
	samePartitionFuncs []*windowFuncHolder,
	partitionIdxs []uint32,
) {
	__antithesis_instrumentation__.Notify(467804)
	windowFnToProcessIdx := -1
	for windowFnIdx, windowFn := range s.infos {
		__antithesis_instrumentation__.Notify(467809)
		if !windowFn.isProcessed {
			__antithesis_instrumentation__.Notify(467810)
			windowFnToProcessIdx = windowFnIdx
			break
		} else {
			__antithesis_instrumentation__.Notify(467811)
		}
	}
	__antithesis_instrumentation__.Notify(467805)
	if windowFnToProcessIdx == -1 {
		__antithesis_instrumentation__.Notify(467812)
		panic("unexpected: no unprocessed window function")
	} else {
		__antithesis_instrumentation__.Notify(467813)
	}
	__antithesis_instrumentation__.Notify(467806)

	windowFnToProcess := s.infos[windowFnToProcessIdx].holder
	partitionIdxs = make([]uint32, len(windowFnToProcess.partitionIdxs))
	for i, idx := range windowFnToProcess.partitionIdxs {
		__antithesis_instrumentation__.Notify(467814)
		partitionIdxs[i] = uint32(idx)
	}
	__antithesis_instrumentation__.Notify(467807)

	samePartitionFuncs = make([]*windowFuncHolder, 0, len(s.infos)-windowFnToProcessIdx)
	samePartitionFuncs = append(samePartitionFuncs, windowFnToProcess)
	s.infos[windowFnToProcessIdx].isProcessed = true
	for _, windowFn := range s.infos[windowFnToProcessIdx+1:] {
		__antithesis_instrumentation__.Notify(467815)
		if windowFn.isProcessed {
			__antithesis_instrumentation__.Notify(467817)
			continue
		} else {
			__antithesis_instrumentation__.Notify(467818)
		}
		__antithesis_instrumentation__.Notify(467816)
		if windowFnToProcess.samePartition(windowFn.holder) {
			__antithesis_instrumentation__.Notify(467819)
			samePartitionFuncs = append(samePartitionFuncs, windowFn.holder)
			windowFn.isProcessed = true
		} else {
			__antithesis_instrumentation__.Notify(467820)
		}
	}
	__antithesis_instrumentation__.Notify(467808)

	return samePartitionFuncs, partitionIdxs
}

func (s *windowPlanState) createWindowFnSpec(
	funcInProgress *windowFuncHolder,
) (execinfrapb.WindowerSpec_WindowFn, *types.T, error) {
	__antithesis_instrumentation__.Notify(467821)
	for _, argIdx := range funcInProgress.argsIdxs {
		__antithesis_instrumentation__.Notify(467828)
		if argIdx >= uint32(len(s.plan.GetResultTypes())) {
			__antithesis_instrumentation__.Notify(467829)
			return execinfrapb.WindowerSpec_WindowFn{}, nil, errors.Errorf("ColIdx out of range (%d)", argIdx)
		} else {
			__antithesis_instrumentation__.Notify(467830)
		}
	}
	__antithesis_instrumentation__.Notify(467822)

	funcSpec, err := rowexec.CreateWindowerSpecFunc(funcInProgress.expr.Func.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(467831)
		return execinfrapb.WindowerSpec_WindowFn{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(467832)
	}
	__antithesis_instrumentation__.Notify(467823)
	argTypes := make([]*types.T, len(funcInProgress.argsIdxs))
	for i, argIdx := range funcInProgress.argsIdxs {
		__antithesis_instrumentation__.Notify(467833)
		argTypes[i] = s.plan.GetResultTypes()[argIdx]
	}
	__antithesis_instrumentation__.Notify(467824)
	_, outputType, err := execinfra.GetWindowFunctionInfo(funcSpec, argTypes...)
	if err != nil {
		__antithesis_instrumentation__.Notify(467834)
		return execinfrapb.WindowerSpec_WindowFn{}, outputType, err
	} else {
		__antithesis_instrumentation__.Notify(467835)
	}
	__antithesis_instrumentation__.Notify(467825)

	ordCols := make([]execinfrapb.Ordering_Column, 0, len(funcInProgress.columnOrdering))
	for _, column := range funcInProgress.columnOrdering {
		__antithesis_instrumentation__.Notify(467836)
		ordCols = append(ordCols, execinfrapb.Ordering_Column{
			ColIdx: uint32(column.ColIdx),

			Direction: execinfrapb.Ordering_Column_Direction(column.Direction - 1),
		})
	}
	__antithesis_instrumentation__.Notify(467826)
	funcInProgressSpec := execinfrapb.WindowerSpec_WindowFn{
		Func:         funcSpec,
		ArgsIdxs:     funcInProgress.argsIdxs,
		Ordering:     execinfrapb.Ordering{Columns: ordCols},
		FilterColIdx: int32(funcInProgress.filterColIdx),
		OutputColIdx: uint32(funcInProgress.outputColIdx),
	}
	if funcInProgress.frame != nil {
		__antithesis_instrumentation__.Notify(467837)

		frameSpec := execinfrapb.WindowerSpec_Frame{}
		if err := frameSpec.InitFromAST(funcInProgress.frame, s.planCtx.EvalContext()); err != nil {
			__antithesis_instrumentation__.Notify(467839)
			return execinfrapb.WindowerSpec_WindowFn{}, outputType, err
		} else {
			__antithesis_instrumentation__.Notify(467840)
		}
		__antithesis_instrumentation__.Notify(467838)
		funcInProgressSpec.Frame = &frameSpec
	} else {
		__antithesis_instrumentation__.Notify(467841)
	}
	__antithesis_instrumentation__.Notify(467827)

	return funcInProgressSpec, outputType, nil
}

var windowerMergeOrdering = execinfrapb.Ordering{}

func (s *windowPlanState) addRenderingOrProjection() error {
	__antithesis_instrumentation__.Notify(467842)

	numWindowFuncsAsIs := 0
	for _, render := range s.n.windowRender {
		__antithesis_instrumentation__.Notify(467846)
		if _, ok := render.(*windowFuncHolder); ok {
			__antithesis_instrumentation__.Notify(467847)
			numWindowFuncsAsIs++
		} else {
			__antithesis_instrumentation__.Notify(467848)
		}
	}
	__antithesis_instrumentation__.Notify(467843)
	if numWindowFuncsAsIs == len(s.infos) {
		__antithesis_instrumentation__.Notify(467849)

		columns := make([]uint32, len(s.n.windowRender))
		passedThruColIdx := uint32(0)
		for i, render := range s.n.windowRender {
			__antithesis_instrumentation__.Notify(467851)
			if render == nil {
				__antithesis_instrumentation__.Notify(467852)
				columns[i] = passedThruColIdx
				passedThruColIdx++
			} else {
				__antithesis_instrumentation__.Notify(467853)

				holder := render.(*windowFuncHolder)
				columns[i] = uint32(holder.outputColIdx)
			}
		}
		__antithesis_instrumentation__.Notify(467850)
		s.plan.AddProjection(columns, windowerMergeOrdering)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(467854)
	}
	__antithesis_instrumentation__.Notify(467844)

	renderExprs := make([]tree.TypedExpr, len(s.n.windowRender))
	visitor := replaceWindowFuncsVisitor{
		columnsMap: s.n.colAndAggContainer.idxMap,
	}

	passedThruColIdx := 0
	renderTypes := make([]*types.T, 0, len(s.n.windowRender))
	for i, render := range s.n.windowRender {
		__antithesis_instrumentation__.Notify(467855)
		if render != nil {
			__antithesis_instrumentation__.Notify(467857)

			renderExprs[i] = visitor.replace(render)
		} else {
			__antithesis_instrumentation__.Notify(467858)

			renderExprs[i] = tree.NewTypedOrdinalReference(passedThruColIdx, s.plan.GetResultTypes()[passedThruColIdx])
			passedThruColIdx++
		}
		__antithesis_instrumentation__.Notify(467856)
		outputType := renderExprs[i].ResolvedType()
		renderTypes = append(renderTypes, outputType)
	}
	__antithesis_instrumentation__.Notify(467845)
	return s.plan.AddRendering(renderExprs, s.planCtx, s.plan.PlanToStreamColMap, renderTypes, windowerMergeOrdering)
}

type replaceWindowFuncsVisitor struct {
	columnsMap map[int]int
}

var _ tree.Visitor = &replaceWindowFuncsVisitor{}

func (v *replaceWindowFuncsVisitor) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(467859)
	switch t := expr.(type) {
	case *windowFuncHolder:
		__antithesis_instrumentation__.Notify(467861)
		return false, tree.NewTypedOrdinalReference(t.outputColIdx, t.ResolvedType())
	case *tree.IndexedVar:
		__antithesis_instrumentation__.Notify(467862)
		return false, tree.NewTypedOrdinalReference(v.columnsMap[t.Idx], t.ResolvedType())
	}
	__antithesis_instrumentation__.Notify(467860)
	return true, expr
}

func (v *replaceWindowFuncsVisitor) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(467863)
	return expr
}

func (v *replaceWindowFuncsVisitor) replace(typedExpr tree.TypedExpr) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(467864)
	expr, _ := tree.WalkExpr(v, typedExpr)
	return expr.(tree.TypedExpr)
}
