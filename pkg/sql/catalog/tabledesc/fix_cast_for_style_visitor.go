package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type fixCastForStyleVisitor struct {
	ctx     context.Context
	semaCtx *tree.SemaContext
	tDesc   catalog.TableDescriptor
	err     error
}

var _ tree.Visitor = &fixCastForStyleVisitor{}

func (v *fixCastForStyleVisitor) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(268716)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(268719)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(268720)
	}
	__antithesis_instrumentation__.Notify(268717)

	_, _, _, err := schemaexpr.DequalifyAndValidateExpr(
		v.ctx,
		v.tDesc,
		expr,
		types.Any,
		"fixCastForStyleVisitor",
		v.semaCtx,
		tree.VolatilityImmutable,
		tree.NewUnqualifiedTableName(tree.Name(v.tDesc.GetName())),
	)
	if err == nil {
		__antithesis_instrumentation__.Notify(268721)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(268722)
	}
	__antithesis_instrumentation__.Notify(268718)

	return true, expr
}

func (v *fixCastForStyleVisitor) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(268723)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(268726)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(268727)
	}
	__antithesis_instrumentation__.Notify(268724)

	if expr, ok := expr.(*tree.CastExpr); ok {
		__antithesis_instrumentation__.Notify(268728)

		typedExpr, err := schemaexpr.DequalifyAndTypeCheckExpr(
			v.ctx,
			v.tDesc,
			expr,
			v.semaCtx,
			tree.NewUnqualifiedTableName(tree.Name(v.tDesc.GetName())),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(268732)

			return expr
		} else {
			__antithesis_instrumentation__.Notify(268733)
		}
		__antithesis_instrumentation__.Notify(268729)
		expr = typedExpr.(*tree.CastExpr)

		sd := sessiondata.SessionData{
			SessionData: sessiondatapb.SessionData{
				IntervalStyleEnabled: v.semaCtx.IntervalStyleEnabled,
				DateStyleEnabled:     v.semaCtx.DateStyleEnabled,
			},
		}
		innerExpr := expr.Expr.(tree.TypedExpr)
		outerTyp := expr.ResolvedType()
		innerTyp := innerExpr.ResolvedType()
		volatility, ok := tree.LookupCastVolatility(innerTyp, outerTyp, &sd)
		if !ok {
			__antithesis_instrumentation__.Notify(268734)
			v.err = errors.AssertionFailedf("Not a valid cast %s -> %s", innerTyp.SQLString(), outerTyp.SQLString())
		} else {
			__antithesis_instrumentation__.Notify(268735)
		}
		__antithesis_instrumentation__.Notify(268730)
		if volatility <= tree.VolatilityImmutable {
			__antithesis_instrumentation__.Notify(268736)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(268737)
		}
		__antithesis_instrumentation__.Notify(268731)

		var newExpr tree.Expr
		switch innerTyp.Family() {
		case types.StringFamily:
			__antithesis_instrumentation__.Notify(268738)
			switch outerTyp.Family() {
			case types.IntervalFamily:
				__antithesis_instrumentation__.Notify(268741)
				newExpr = &tree.CastExpr{
					Expr:       &tree.FuncExpr{Func: tree.WrapFunction("parse_interval"), Exprs: tree.Exprs{expr.Expr}},
					Type:       expr.Type,
					SyntaxMode: tree.CastShort,
				}
				return newExpr
			case types.DateFamily:
				__antithesis_instrumentation__.Notify(268742)
				newExpr = &tree.CastExpr{
					Expr:       &tree.FuncExpr{Func: tree.WrapFunction("parse_date"), Exprs: tree.Exprs{expr.Expr}},
					Type:       expr.Type,
					SyntaxMode: tree.CastShort,
				}
				return newExpr
			case types.TimeFamily:
				__antithesis_instrumentation__.Notify(268743)
				newExpr = &tree.CastExpr{
					Expr:       &tree.FuncExpr{Func: tree.WrapFunction("parse_time"), Exprs: tree.Exprs{expr.Expr}},
					Type:       expr.Type,
					SyntaxMode: tree.CastShort,
				}
				return newExpr
			case types.TimeTZFamily:
				__antithesis_instrumentation__.Notify(268744)
				newExpr = &tree.CastExpr{
					Expr:       &tree.FuncExpr{Func: tree.WrapFunction("parse_timetz"), Exprs: tree.Exprs{expr.Expr}},
					Type:       expr.Type,
					SyntaxMode: tree.CastShort,
				}
				return newExpr
			case types.TimestampFamily:
				__antithesis_instrumentation__.Notify(268745)
				newExpr = &tree.CastExpr{
					Expr:       &tree.FuncExpr{Func: tree.WrapFunction("parse_timestamp"), Exprs: tree.Exprs{expr.Expr}},
					Type:       expr.Type,
					SyntaxMode: tree.CastShort,
				}
				return newExpr
			default:
				__antithesis_instrumentation__.Notify(268746)
			}
		case types.IntervalFamily, types.DateFamily, types.TimestampFamily, types.TimeFamily, types.TimeTZFamily, types.TimestampTZFamily:
			__antithesis_instrumentation__.Notify(268739)
			if outerTyp.Family() == types.StringFamily {
				__antithesis_instrumentation__.Notify(268747)
				newExpr = &tree.CastExpr{
					Expr:       &tree.FuncExpr{Func: tree.WrapFunction("to_char"), Exprs: tree.Exprs{expr.Expr}},
					Type:       expr.Type,
					SyntaxMode: tree.CastShort,
				}
				return newExpr
			} else {
				__antithesis_instrumentation__.Notify(268748)
			}
		default:
			__antithesis_instrumentation__.Notify(268740)
		}
	} else {
		__antithesis_instrumentation__.Notify(268749)
	}
	__antithesis_instrumentation__.Notify(268725)
	return expr
}

func ResolveCastForStyleUsingVisitor(
	ctx context.Context, semaCtx *tree.SemaContext, desc *descpb.TableDescriptor, expr tree.Expr,
) (tree.Expr, bool, error) {
	__antithesis_instrumentation__.Notify(268750)
	v := &fixCastForStyleVisitor{ctx: ctx, semaCtx: semaCtx}
	descBuilder := NewBuilder(desc)
	tDesc := descBuilder.BuildImmutableTable()
	v.tDesc = tDesc

	expr, changed := tree.WalkExpr(v, expr)
	return expr, changed, v.err
}
