package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/transform"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func MakeDefaultExprs(
	ctx context.Context,
	cols []catalog.Column,
	txCtx *transform.ExprTransformContext,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
) ([]tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(268091)

	haveDefaults := false
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(268097)
		if col.HasDefault() {
			__antithesis_instrumentation__.Notify(268098)
			haveDefaults = true
			break
		} else {
			__antithesis_instrumentation__.Notify(268099)
		}
	}
	__antithesis_instrumentation__.Notify(268092)
	if !haveDefaults {
		__antithesis_instrumentation__.Notify(268100)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(268101)
	}
	__antithesis_instrumentation__.Notify(268093)

	defaultExprs := make([]tree.TypedExpr, 0, len(cols))
	exprStrings := make([]string, 0, len(cols))
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(268102)
		if col.HasDefault() {
			__antithesis_instrumentation__.Notify(268103)
			exprStrings = append(exprStrings, col.GetDefaultExpr())
		} else {
			__antithesis_instrumentation__.Notify(268104)
		}
	}
	__antithesis_instrumentation__.Notify(268094)
	exprs, err := parser.ParseExprs(exprStrings)
	if err != nil {
		__antithesis_instrumentation__.Notify(268105)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268106)
	}
	__antithesis_instrumentation__.Notify(268095)

	defExprIdx := 0
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(268107)
		if !col.HasDefault() {
			__antithesis_instrumentation__.Notify(268111)
			defaultExprs = append(defaultExprs, tree.DNull)
			continue
		} else {
			__antithesis_instrumentation__.Notify(268112)
		}
		__antithesis_instrumentation__.Notify(268108)
		expr := exprs[defExprIdx]
		typedExpr, err := tree.TypeCheck(ctx, expr, semaCtx, col.GetType())
		if err != nil {
			__antithesis_instrumentation__.Notify(268113)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(268114)
		}
		__antithesis_instrumentation__.Notify(268109)
		if typedExpr, err = txCtx.NormalizeExpr(evalCtx, typedExpr); err != nil {
			__antithesis_instrumentation__.Notify(268115)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(268116)
		}
		__antithesis_instrumentation__.Notify(268110)
		defaultExprs = append(defaultExprs, typedExpr)
		defExprIdx++
	}
	__antithesis_instrumentation__.Notify(268096)
	return defaultExprs, nil
}

func ProcessColumnSet(
	cols []catalog.Column, tableDesc catalog.TableDescriptor, inSet func(column catalog.Column) bool,
) []catalog.Column {
	__antithesis_instrumentation__.Notify(268117)
	var colIDSet catalog.TableColSet
	for i := range cols {
		__antithesis_instrumentation__.Notify(268120)
		colIDSet.Add(cols[i].GetID())
	}
	__antithesis_instrumentation__.Notify(268118)

	ret := make([]catalog.Column, 0, len(tableDesc.AllColumns()))
	ret = append(ret, cols...)
	for _, col := range tableDesc.WritableColumns() {
		__antithesis_instrumentation__.Notify(268121)
		if inSet(col) {
			__antithesis_instrumentation__.Notify(268122)
			if !colIDSet.Contains(col.GetID()) {
				__antithesis_instrumentation__.Notify(268123)
				colIDSet.Add(col.GetID())
				ret = append(ret, col)
			} else {
				__antithesis_instrumentation__.Notify(268124)
			}
		} else {
			__antithesis_instrumentation__.Notify(268125)
		}
	}
	__antithesis_instrumentation__.Notify(268119)
	return ret
}
