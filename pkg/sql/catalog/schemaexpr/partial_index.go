package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/transform"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func ValidatePartialIndexPredicate(
	ctx context.Context,
	desc catalog.TableDescriptor,
	e tree.Expr,
	tn *tree.TableName,
	semaCtx *tree.SemaContext,
) (string, error) {
	__antithesis_instrumentation__.Notify(268251)
	expr, _, _, err := DequalifyAndValidateExpr(
		ctx,
		desc,
		e,
		types.Bool,
		"index predicate",
		semaCtx,
		tree.VolatilityImmutable,
		tn,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(268253)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(268254)
	}
	__antithesis_instrumentation__.Notify(268252)

	return expr, nil
}

func MakePartialIndexExprs(
	ctx context.Context,
	indexes []catalog.Index,
	cols []catalog.Column,
	tableDesc catalog.TableDescriptor,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
) (_ map[descpb.IndexID]tree.TypedExpr, refColIDs catalog.TableColSet, _ error) {
	__antithesis_instrumentation__.Notify(268255)

	partialIndexCount := 0
	for i := range indexes {
		__antithesis_instrumentation__.Notify(268259)
		if indexes[i].IsPartial() {
			__antithesis_instrumentation__.Notify(268260)
			partialIndexCount++
		} else {
			__antithesis_instrumentation__.Notify(268261)
		}
	}
	__antithesis_instrumentation__.Notify(268256)
	if partialIndexCount == 0 {
		__antithesis_instrumentation__.Notify(268262)
		return nil, refColIDs, nil
	} else {
		__antithesis_instrumentation__.Notify(268263)
	}
	__antithesis_instrumentation__.Notify(268257)

	exprs := make(map[descpb.IndexID]tree.TypedExpr, partialIndexCount)

	tn := tree.NewUnqualifiedTableName(tree.Name(tableDesc.GetName()))
	nr := newNameResolver(evalCtx, tableDesc.GetID(), tn, cols)
	nr.addIVarContainerToSemaCtx(semaCtx)

	var txCtx transform.ExprTransformContext
	for _, idx := range indexes {
		__antithesis_instrumentation__.Notify(268264)
		if idx.IsPartial() {
			__antithesis_instrumentation__.Notify(268265)
			expr, err := parser.ParseExpr(idx.GetPredicate())
			if err != nil {
				__antithesis_instrumentation__.Notify(268271)
				return nil, refColIDs, err
			} else {
				__antithesis_instrumentation__.Notify(268272)
			}
			__antithesis_instrumentation__.Notify(268266)

			colIDs, err := ExtractColumnIDs(tableDesc, expr)
			if err != nil {
				__antithesis_instrumentation__.Notify(268273)
				return nil, refColIDs, err
			} else {
				__antithesis_instrumentation__.Notify(268274)
			}
			__antithesis_instrumentation__.Notify(268267)
			refColIDs.UnionWith(colIDs)

			expr, err = nr.resolveNames(expr)
			if err != nil {
				__antithesis_instrumentation__.Notify(268275)
				return nil, refColIDs, err
			} else {
				__antithesis_instrumentation__.Notify(268276)
			}
			__antithesis_instrumentation__.Notify(268268)

			typedExpr, err := tree.TypeCheck(ctx, expr, semaCtx, types.Bool)
			if err != nil {
				__antithesis_instrumentation__.Notify(268277)
				return nil, refColIDs, err
			} else {
				__antithesis_instrumentation__.Notify(268278)
			}
			__antithesis_instrumentation__.Notify(268269)

			if typedExpr, err = txCtx.NormalizeExpr(evalCtx, typedExpr); err != nil {
				__antithesis_instrumentation__.Notify(268279)
				return nil, refColIDs, err
			} else {
				__antithesis_instrumentation__.Notify(268280)
			}
			__antithesis_instrumentation__.Notify(268270)

			exprs[idx.GetID()] = typedExpr
		} else {
			__antithesis_instrumentation__.Notify(268281)
		}
	}
	__antithesis_instrumentation__.Notify(268258)

	return exprs, refColIDs, nil
}
