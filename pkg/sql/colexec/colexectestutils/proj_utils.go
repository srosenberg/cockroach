package colexectestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecargs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

type MockTypeContext struct {
	Typs []*types.T
}

var _ tree.IndexedVarContainer = &MockTypeContext{}

func (p *MockTypeContext) IndexedVarEval(idx int, ctx *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(431006)
	return tree.DNull.Eval(ctx)
}

func (p *MockTypeContext) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(431007)
	return p.Typs[idx]
}

func (p *MockTypeContext) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(431008)
	n := tree.Name(fmt.Sprintf("$%d", idx))
	return &n
}

func CreateTestProjectingOperator(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	input colexecop.Operator,
	inputTypes []*types.T,
	projectingExpr string,
	testMemAcc *mon.BoundAccount,
) (colexecop.Operator, error) {
	__antithesis_instrumentation__.Notify(431009)
	expr, err := parser.ParseExpr(projectingExpr)
	if err != nil {
		__antithesis_instrumentation__.Notify(431014)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(431015)
	}
	__antithesis_instrumentation__.Notify(431010)
	p := &MockTypeContext{Typs: inputTypes}
	semaCtx := tree.MakeSemaContext()
	semaCtx.IVarContainer = p
	typedExpr, err := tree.TypeCheck(ctx, expr, &semaCtx, types.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(431016)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(431017)
	}
	__antithesis_instrumentation__.Notify(431011)
	renderExprs := make([]execinfrapb.Expression, len(inputTypes)+1)
	for i := range inputTypes {
		__antithesis_instrumentation__.Notify(431018)
		renderExprs[i].Expr = fmt.Sprintf("@%d", i+1)
	}
	__antithesis_instrumentation__.Notify(431012)
	renderExprs[len(inputTypes)].LocalExpr = typedExpr
	spec := &execinfrapb.ProcessorSpec{
		Input: []execinfrapb.InputSyncSpec{{ColumnTypes: inputTypes}},
		Core: execinfrapb.ProcessorCoreUnion{
			Noop: &execinfrapb.NoopCoreSpec{},
		},
		Post: execinfrapb.PostProcessSpec{
			RenderExprs: renderExprs,
		},
		ResultTypes: append(inputTypes, typedExpr.ResolvedType()),
	}
	args := &colexecargs.NewColOperatorArgs{
		Spec:                spec,
		Inputs:              []colexecargs.OpWithMetaInfo{{Root: input}},
		StreamingMemAccount: testMemAcc,
	}
	result, err := colexecargs.TestNewColOperator(ctx, flowCtx, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(431019)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(431020)
	}
	__antithesis_instrumentation__.Notify(431013)
	return result.Root, nil
}
