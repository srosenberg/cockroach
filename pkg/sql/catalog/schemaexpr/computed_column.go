package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/transform"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/errors"
)

func ValidateComputedColumnExpression(
	ctx context.Context,
	desc catalog.TableDescriptor,
	d *tree.ColumnTableDef,
	tn *tree.TableName,
	context string,
	semaCtx *tree.SemaContext,
) (serializedExpr string, _ *types.T, _ error) {
	__antithesis_instrumentation__.Notify(267962)
	if d.HasDefaultExpr() {
		__antithesis_instrumentation__.Notify(267969)
		return "", nil, pgerror.Newf(
			pgcode.InvalidTableDefinition,
			"%s cannot have default values",
			context,
		)
	} else {
		__antithesis_instrumentation__.Notify(267970)
	}
	__antithesis_instrumentation__.Notify(267963)

	var depColIDs catalog.TableColSet

	err := iterColDescriptors(desc, d.Computed.Expr, func(c catalog.Column) error {
		__antithesis_instrumentation__.Notify(267971)
		if c.IsInaccessible() {
			__antithesis_instrumentation__.Notify(267974)
			return pgerror.Newf(
				pgcode.UndefinedColumn,
				"column %q is inaccessible and cannot be referenced in a computed column expression",
				c.GetName(),
			)
		} else {
			__antithesis_instrumentation__.Notify(267975)
		}
		__antithesis_instrumentation__.Notify(267972)
		if c.IsComputed() {
			__antithesis_instrumentation__.Notify(267976)
			return pgerror.Newf(
				pgcode.InvalidTableDefinition,
				"%s expression cannot reference computed columns",
				context,
			)
		} else {
			__antithesis_instrumentation__.Notify(267977)
		}
		__antithesis_instrumentation__.Notify(267973)
		depColIDs.Add(c.GetID())

		return nil
	})
	__antithesis_instrumentation__.Notify(267964)
	if err != nil {
		__antithesis_instrumentation__.Notify(267978)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(267979)
	}
	__antithesis_instrumentation__.Notify(267965)

	defType, err := tree.ResolveType(ctx, d.Type, semaCtx.GetTypeResolver())
	if err != nil {
		__antithesis_instrumentation__.Notify(267980)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(267981)
	}
	__antithesis_instrumentation__.Notify(267966)

	expr, typ, _, err := DequalifyAndValidateExpr(
		ctx,
		desc,
		d.Computed.Expr,
		defType,
		context,
		semaCtx,
		tree.VolatilityImmutable,
		tn,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(267982)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(267983)
	}
	__antithesis_instrumentation__.Notify(267967)

	if d.IsVirtual() {
		__antithesis_instrumentation__.Notify(267984)
		var mutationColumnNames []string
		var err error
		depColIDs.ForEach(func(colID descpb.ColumnID) {
			__antithesis_instrumentation__.Notify(267987)
			if err != nil {
				__antithesis_instrumentation__.Notify(267990)
				return
			} else {
				__antithesis_instrumentation__.Notify(267991)
			}
			__antithesis_instrumentation__.Notify(267988)
			var col catalog.Column
			if col, err = desc.FindColumnWithID(colID); err != nil {
				__antithesis_instrumentation__.Notify(267992)
				err = errors.WithAssertionFailure(err)
				return
			} else {
				__antithesis_instrumentation__.Notify(267993)
			}
			__antithesis_instrumentation__.Notify(267989)
			if !col.Public() {
				__antithesis_instrumentation__.Notify(267994)
				mutationColumnNames = append(mutationColumnNames,
					strconv.Quote(col.GetName()))
			} else {
				__antithesis_instrumentation__.Notify(267995)
			}
		})
		__antithesis_instrumentation__.Notify(267985)
		if err != nil {
			__antithesis_instrumentation__.Notify(267996)
			return "", nil, err
		} else {
			__antithesis_instrumentation__.Notify(267997)
		}
		__antithesis_instrumentation__.Notify(267986)
		if len(mutationColumnNames) > 0 {
			__antithesis_instrumentation__.Notify(267998)
			if context == "index element" {
				__antithesis_instrumentation__.Notify(268000)
				return "", nil, unimplemented.Newf(
					"index element expression referencing mutation columns",
					"index element expression referencing columns (%s) added in the current transaction",
					strings.Join(mutationColumnNames, ", "))
			} else {
				__antithesis_instrumentation__.Notify(268001)
			}
			__antithesis_instrumentation__.Notify(267999)
			return "", nil, unimplemented.Newf(
				"virtual computed columns referencing mutation columns",
				"virtual computed column %q referencing columns (%s) added in the "+
					"current transaction", d.Name, strings.Join(mutationColumnNames, ", "))
		} else {
			__antithesis_instrumentation__.Notify(268002)
		}
	} else {
		__antithesis_instrumentation__.Notify(268003)
	}
	__antithesis_instrumentation__.Notify(267968)

	return expr, typ, nil
}

func ValidateColumnHasNoDependents(desc catalog.TableDescriptor, col catalog.Column) error {
	__antithesis_instrumentation__.Notify(268004)
	for _, c := range desc.NonDropColumns() {
		__antithesis_instrumentation__.Notify(268006)
		if !c.IsComputed() {
			__antithesis_instrumentation__.Notify(268010)
			continue
		} else {
			__antithesis_instrumentation__.Notify(268011)
		}
		__antithesis_instrumentation__.Notify(268007)

		expr, err := parser.ParseExpr(c.GetComputeExpr())
		if err != nil {
			__antithesis_instrumentation__.Notify(268012)

			return errors.WithAssertionFailure(err)
		} else {
			__antithesis_instrumentation__.Notify(268013)
		}
		__antithesis_instrumentation__.Notify(268008)

		err = iterColDescriptors(desc, expr, func(colVar catalog.Column) error {
			__antithesis_instrumentation__.Notify(268014)
			if colVar.GetID() == col.GetID() {
				__antithesis_instrumentation__.Notify(268016)
				return pgerror.Newf(
					pgcode.InvalidColumnReference,
					"column %q is referenced by computed column %q",
					col.GetName(),
					c.GetName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(268017)
			}
			__antithesis_instrumentation__.Notify(268015)
			return nil
		})
		__antithesis_instrumentation__.Notify(268009)
		if err != nil {
			__antithesis_instrumentation__.Notify(268018)
			return err
		} else {
			__antithesis_instrumentation__.Notify(268019)
		}
	}
	__antithesis_instrumentation__.Notify(268005)
	return nil
}

func MakeComputedExprs(
	ctx context.Context,
	input, sourceColumns []catalog.Column,
	tableDesc catalog.TableDescriptor,
	tn *tree.TableName,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
) (_ []tree.TypedExpr, refColIDs catalog.TableColSet, _ error) {
	__antithesis_instrumentation__.Notify(268020)

	haveComputed := false
	for i := range input {
		__antithesis_instrumentation__.Notify(268026)
		if input[i].IsComputed() {
			__antithesis_instrumentation__.Notify(268027)
			haveComputed = true
			break
		} else {
			__antithesis_instrumentation__.Notify(268028)
		}
	}
	__antithesis_instrumentation__.Notify(268021)
	if !haveComputed {
		__antithesis_instrumentation__.Notify(268029)
		return nil, catalog.TableColSet{}, nil
	} else {
		__antithesis_instrumentation__.Notify(268030)
	}
	__antithesis_instrumentation__.Notify(268022)

	computedExprs := make([]tree.TypedExpr, 0, len(input))
	exprStrings := make([]string, 0, len(input))
	for _, col := range input {
		__antithesis_instrumentation__.Notify(268031)
		if col.IsComputed() {
			__antithesis_instrumentation__.Notify(268032)
			exprStrings = append(exprStrings, col.GetComputeExpr())
		} else {
			__antithesis_instrumentation__.Notify(268033)
		}
	}
	__antithesis_instrumentation__.Notify(268023)

	exprs, err := parser.ParseExprs(exprStrings)
	if err != nil {
		__antithesis_instrumentation__.Notify(268034)
		return nil, catalog.TableColSet{}, err
	} else {
		__antithesis_instrumentation__.Notify(268035)
	}
	__antithesis_instrumentation__.Notify(268024)

	nr := newNameResolver(evalCtx, tableDesc.GetID(), tn, sourceColumns)
	nr.addIVarContainerToSemaCtx(semaCtx)

	var txCtx transform.ExprTransformContext
	compExprIdx := 0
	for _, col := range input {
		__antithesis_instrumentation__.Notify(268036)
		if !col.IsComputed() {
			__antithesis_instrumentation__.Notify(268042)
			computedExprs = append(computedExprs, tree.DNull)
			nr.addColumn(col)
			continue
		} else {
			__antithesis_instrumentation__.Notify(268043)
		}
		__antithesis_instrumentation__.Notify(268037)

		colIDs, err := ExtractColumnIDs(tableDesc, exprs[compExprIdx])
		if err != nil {
			__antithesis_instrumentation__.Notify(268044)
			return nil, refColIDs, err
		} else {
			__antithesis_instrumentation__.Notify(268045)
		}
		__antithesis_instrumentation__.Notify(268038)
		refColIDs.UnionWith(colIDs)

		expr, err := nr.resolveNames(exprs[compExprIdx])
		if err != nil {
			__antithesis_instrumentation__.Notify(268046)
			return nil, catalog.TableColSet{}, err
		} else {
			__antithesis_instrumentation__.Notify(268047)
		}
		__antithesis_instrumentation__.Notify(268039)

		typedExpr, err := tree.TypeCheck(ctx, expr, semaCtx, col.GetType())
		if err != nil {
			__antithesis_instrumentation__.Notify(268048)
			return nil, catalog.TableColSet{}, err
		} else {
			__antithesis_instrumentation__.Notify(268049)
		}
		__antithesis_instrumentation__.Notify(268040)
		if typedExpr, err = txCtx.NormalizeExpr(evalCtx, typedExpr); err != nil {
			__antithesis_instrumentation__.Notify(268050)
			return nil, catalog.TableColSet{}, err
		} else {
			__antithesis_instrumentation__.Notify(268051)
		}
		__antithesis_instrumentation__.Notify(268041)
		computedExprs = append(computedExprs, typedExpr)
		compExprIdx++
		nr.addColumn(col)
	}
	__antithesis_instrumentation__.Notify(268025)
	return computedExprs, refColIDs, nil
}
