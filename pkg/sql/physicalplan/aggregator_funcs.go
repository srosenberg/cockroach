package physicalplan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type FinalStageInfo struct {
	Fn execinfrapb.AggregatorSpec_Func

	LocalIdxs []uint32
}

type DistAggregationInfo struct {
	LocalStage []execinfrapb.AggregatorSpec_Func

	FinalStage []FinalStageInfo

	FinalRendering func(h *tree.IndexedVarHelper, varIdxs []int) (tree.TypedExpr, error)
}

var passThroughLocalIdxs = []uint32{0}

var DistAggregationTable = map[execinfrapb.AggregatorSpec_Func]DistAggregationInfo{
	execinfrapb.AnyNotNull: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.AnyNotNull},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.AnyNotNull,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.Avg: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{
			execinfrapb.Sum,
			execinfrapb.Count,
		},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.Sum,
				LocalIdxs: []uint32{0},
			},
			{
				Fn:        execinfrapb.SumInt,
				LocalIdxs: []uint32{1},
			},
		},
		FinalRendering: func(h *tree.IndexedVarHelper, varIdxs []int) (tree.TypedExpr, error) {
			__antithesis_instrumentation__.Notify(562075)
			if len(varIdxs) < 2 {
				__antithesis_instrumentation__.Notify(562078)
				panic("fewer than two final aggregation values passed into final render")
			} else {
				__antithesis_instrumentation__.Notify(562079)
			}
			__antithesis_instrumentation__.Notify(562076)
			sum := h.IndexedVar(varIdxs[0])
			count := h.IndexedVar(varIdxs[1])

			expr := &tree.BinaryExpr{
				Operator: treebin.MakeBinaryOperator(treebin.Div),
				Left:     sum,
				Right:    count,
			}

			if sum.ResolvedType().Family() == types.FloatFamily {
				__antithesis_instrumentation__.Notify(562080)
				expr.Right = &tree.CastExpr{
					Expr: count,
					Type: types.Float,
				}
			} else {
				__antithesis_instrumentation__.Notify(562081)
			}
			__antithesis_instrumentation__.Notify(562077)
			semaCtx := tree.MakeSemaContext()
			semaCtx.IVarContainer = h.Container()
			return expr.TypeCheck(context.TODO(), &semaCtx, types.Any)
		},
	},

	execinfrapb.BoolAnd: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.BoolAnd},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.BoolAnd,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.BoolOr: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.BoolOr},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.BoolOr,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.Count: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.Count},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.SumInt,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.Max: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.Max},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.Max,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.Min: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.Min},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.Min,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.Stddev: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{
			execinfrapb.Sqrdiff,
			execinfrapb.Sum,
			execinfrapb.Count,
		},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalStddev,
				LocalIdxs: []uint32{0, 1, 2},
			},
		},
	},

	execinfrapb.StddevPop: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{
			execinfrapb.Sqrdiff,
			execinfrapb.Sum,
			execinfrapb.Count,
		},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalStddevPop,
				LocalIdxs: []uint32{0, 1, 2},
			},
		},
	},

	execinfrapb.Sum: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.Sum},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.Sum,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.SumInt: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.SumInt},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.SumInt,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.Variance: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{
			execinfrapb.Sqrdiff,
			execinfrapb.Sum,
			execinfrapb.Count,
		},

		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalVariance,
				LocalIdxs: []uint32{0, 1, 2},
			},
		},
	},

	execinfrapb.VarPop: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{
			execinfrapb.Sqrdiff,
			execinfrapb.Sum,
			execinfrapb.Count,
		},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalVarPop,
				LocalIdxs: []uint32{0, 1, 2},
			},
		},
	},

	execinfrapb.XorAgg: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.XorAgg},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.XorAgg,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.CountRows: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.CountRows},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.SumInt,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.BitAnd: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.BitAnd},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.BitAnd,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.BitOr: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.BitOr},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.BitOr,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.CovarPop: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalCovarPop,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrSxx: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalRegrSxx,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrSxy: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalRegrSxy,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrSyy: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalRegrSyy,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrAvgx: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalRegrAvgx,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrAvgy: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalRegrAvgy,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrIntercept: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalRegrIntercept,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrR2: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalRegrR2,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrSlope: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalRegrSlope,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.CovarSamp: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalCovarSamp,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.Corr: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.TransitionRegrAggregate},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalCorr,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.RegrCount: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{execinfrapb.RegrCount},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.SumInt,
				LocalIdxs: passThroughLocalIdxs,
			},
		},
	},

	execinfrapb.Sqrdiff: {
		LocalStage: []execinfrapb.AggregatorSpec_Func{
			execinfrapb.Sqrdiff,
			execinfrapb.Sum,
			execinfrapb.Count,
		},
		FinalStage: []FinalStageInfo{
			{
				Fn:        execinfrapb.FinalSqrdiff,
				LocalIdxs: []uint32{0, 1, 2},
			},
		},
	},
}
