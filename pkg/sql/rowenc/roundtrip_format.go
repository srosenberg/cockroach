package rowenc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func ParseDatumStringAs(t *types.T, s string, evalCtx *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(570960)
	switch t.Family() {

	case types.ArrayFamily, types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(570961)
		return parseAsTyp(evalCtx, t, s)
	default:
		__antithesis_instrumentation__.Notify(570962)
		res, _, err := tree.ParseAndRequireString(t, s, evalCtx)
		return res, err
	}
}

func ParseDatumStringAsWithRawBytes(
	t *types.T, s string, evalCtx *tree.EvalContext,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(570963)
	switch t.Family() {
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(570964)
		return tree.NewDBytes(tree.DBytes(s)), nil
	default:
		__antithesis_instrumentation__.Notify(570965)
		return ParseDatumStringAs(t, s, evalCtx)
	}
}

func parseAsTyp(evalCtx *tree.EvalContext, typ *types.T, s string) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(570966)
	expr, err := parser.ParseExpr(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(570969)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570970)
	}
	__antithesis_instrumentation__.Notify(570967)
	semaCtx := tree.MakeSemaContext()
	typedExpr, err := tree.TypeCheck(evalCtx.Context, expr, &semaCtx, typ)
	if err != nil {
		__antithesis_instrumentation__.Notify(570971)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570972)
	}
	__antithesis_instrumentation__.Notify(570968)
	datum, err := typedExpr.Eval(evalCtx)
	return datum, err
}
