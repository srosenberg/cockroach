package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

func DequalifyAndTypeCheckExpr(
	ctx context.Context,
	desc catalog.TableDescriptor,
	expr tree.Expr,
	semaCtx *tree.SemaContext,
	tn *tree.TableName,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(268126)
	nonDropColumns := desc.NonDropColumns()
	sourceInfo := colinfo.NewSourceInfoForSingleTable(
		*tn, colinfo.ResultColumnsFromColumns(desc.GetID(), nonDropColumns),
	)
	expr, err := dequalifyColumnRefs(ctx, sourceInfo, expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(268130)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268131)
	}
	__antithesis_instrumentation__.Notify(268127)

	replacedExpr, _, err := replaceColumnVars(desc, expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(268132)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268133)
	}
	__antithesis_instrumentation__.Notify(268128)

	typedExpr, err := tree.TypeCheck(ctx, replacedExpr, semaCtx, types.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(268134)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268135)
	}
	__antithesis_instrumentation__.Notify(268129)

	return typedExpr, nil
}

func DequalifyAndValidateExpr(
	ctx context.Context,
	desc catalog.TableDescriptor,
	expr tree.Expr,
	typ *types.T,
	context string,
	semaCtx *tree.SemaContext,
	maxVolatility tree.Volatility,
	tn *tree.TableName,
) (string, *types.T, catalog.TableColSet, error) {
	__antithesis_instrumentation__.Notify(268136)
	var colIDs catalog.TableColSet
	nonDropColumns := desc.NonDropColumns()
	sourceInfo := colinfo.NewSourceInfoForSingleTable(
		*tn, colinfo.ResultColumnsFromColumns(desc.GetID(), nonDropColumns),
	)
	expr, err := dequalifyColumnRefs(ctx, sourceInfo, expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(268140)
		return "", nil, colIDs, err
	} else {
		__antithesis_instrumentation__.Notify(268141)
	}
	__antithesis_instrumentation__.Notify(268137)

	replacedExpr, colIDs, err := replaceColumnVars(desc, expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(268142)
		return "", nil, colIDs, err
	} else {
		__antithesis_instrumentation__.Notify(268143)
	}
	__antithesis_instrumentation__.Notify(268138)

	typedExpr, err := SanitizeVarFreeExpr(
		ctx,
		replacedExpr,
		typ,
		context,
		semaCtx,
		maxVolatility,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(268144)
		return "", nil, colIDs, err
	} else {
		__antithesis_instrumentation__.Notify(268145)
	}
	__antithesis_instrumentation__.Notify(268139)

	return tree.Serialize(typedExpr), typedExpr.ResolvedType(), colIDs, nil
}

func ExtractColumnIDs(
	desc catalog.TableDescriptor, rootExpr tree.Expr,
) (catalog.TableColSet, error) {
	__antithesis_instrumentation__.Notify(268146)
	var colIDs catalog.TableColSet

	_, err := tree.SimpleVisit(rootExpr, func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(268148)
		vBase, ok := expr.(tree.VarName)
		if !ok {
			__antithesis_instrumentation__.Notify(268153)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(268154)
		}
		__antithesis_instrumentation__.Notify(268149)

		v, err := vBase.NormalizeVarName()
		if err != nil {
			__antithesis_instrumentation__.Notify(268155)
			return false, nil, err
		} else {
			__antithesis_instrumentation__.Notify(268156)
		}
		__antithesis_instrumentation__.Notify(268150)

		c, ok := v.(*tree.ColumnItem)
		if !ok {
			__antithesis_instrumentation__.Notify(268157)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(268158)
		}
		__antithesis_instrumentation__.Notify(268151)

		col, err := desc.FindColumnWithName(c.ColumnName)
		if err != nil {
			__antithesis_instrumentation__.Notify(268159)
			return false, nil, err
		} else {
			__antithesis_instrumentation__.Notify(268160)
		}
		__antithesis_instrumentation__.Notify(268152)

		colIDs.Add(col.GetID())
		return false, expr, nil
	})
	__antithesis_instrumentation__.Notify(268147)

	return colIDs, err
}

type returnFalse struct{}

func (returnFalse) Error() string {
	__antithesis_instrumentation__.Notify(268161)
	panic("unimplemented")
}

var returnFalsePseudoError error = returnFalse{}

func HasValidColumnReferences(desc catalog.TableDescriptor, rootExpr tree.Expr) (bool, error) {
	__antithesis_instrumentation__.Notify(268162)
	_, err := tree.SimpleVisit(rootExpr, func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(268166)
		vBase, ok := expr.(tree.VarName)
		if !ok {
			__antithesis_instrumentation__.Notify(268171)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(268172)
		}
		__antithesis_instrumentation__.Notify(268167)

		v, err := vBase.NormalizeVarName()
		if err != nil {
			__antithesis_instrumentation__.Notify(268173)
			return false, nil, err
		} else {
			__antithesis_instrumentation__.Notify(268174)
		}
		__antithesis_instrumentation__.Notify(268168)

		c, ok := v.(*tree.ColumnItem)
		if !ok {
			__antithesis_instrumentation__.Notify(268175)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(268176)
		}
		__antithesis_instrumentation__.Notify(268169)

		_, err = desc.FindColumnWithName(c.ColumnName)
		if err != nil {
			__antithesis_instrumentation__.Notify(268177)
			return false, expr, returnFalsePseudoError
		} else {
			__antithesis_instrumentation__.Notify(268178)
		}
		__antithesis_instrumentation__.Notify(268170)

		return false, expr, nil
	})
	__antithesis_instrumentation__.Notify(268163)
	if errors.Is(err, returnFalsePseudoError) {
		__antithesis_instrumentation__.Notify(268179)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(268180)
	}
	__antithesis_instrumentation__.Notify(268164)
	if err != nil {
		__antithesis_instrumentation__.Notify(268181)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(268182)
	}
	__antithesis_instrumentation__.Notify(268165)
	return true, nil
}

func FormatExprForDisplay(
	ctx context.Context,
	desc catalog.TableDescriptor,
	exprStr string,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	fmtFlags tree.FmtFlags,
) (string, error) {
	__antithesis_instrumentation__.Notify(268183)
	return formatExprForDisplayImpl(
		ctx,
		desc,
		exprStr,
		semaCtx,
		sessionData,
		fmtFlags,
		false,
	)
}

func FormatExprForExpressionIndexDisplay(
	ctx context.Context,
	desc catalog.TableDescriptor,
	exprStr string,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	fmtFlags tree.FmtFlags,
) (string, error) {
	__antithesis_instrumentation__.Notify(268184)
	return formatExprForDisplayImpl(
		ctx,
		desc,
		exprStr,
		semaCtx,
		sessionData,
		fmtFlags,
		true,
	)
}

func formatExprForDisplayImpl(
	ctx context.Context,
	desc catalog.TableDescriptor,
	exprStr string,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	fmtFlags tree.FmtFlags,
	wrapNonFuncExprs bool,
) (string, error) {
	__antithesis_instrumentation__.Notify(268185)
	expr, err := deserializeExprForFormatting(ctx, desc, exprStr, semaCtx, fmtFlags)
	if err != nil {
		__antithesis_instrumentation__.Notify(268190)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(268191)
	}
	__antithesis_instrumentation__.Notify(268186)

	replacedExpr, err := ReplaceIDsWithFQNames(ctx, expr, semaCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(268192)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(268193)
	}
	__antithesis_instrumentation__.Notify(268187)
	f := tree.NewFmtCtx(fmtFlags, tree.FmtDataConversionConfig(sessionData.DataConversionConfig))
	_, isFunc := expr.(*tree.FuncExpr)
	if wrapNonFuncExprs && func() bool {
		__antithesis_instrumentation__.Notify(268194)
		return !isFunc == true
	}() == true {
		__antithesis_instrumentation__.Notify(268195)
		f.WriteByte('(')
	} else {
		__antithesis_instrumentation__.Notify(268196)
	}
	__antithesis_instrumentation__.Notify(268188)
	f.FormatNode(replacedExpr)
	if wrapNonFuncExprs && func() bool {
		__antithesis_instrumentation__.Notify(268197)
		return !isFunc == true
	}() == true {
		__antithesis_instrumentation__.Notify(268198)
		f.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(268199)
	}
	__antithesis_instrumentation__.Notify(268189)
	return f.CloseAndGetString(), nil
}

func deserializeExprForFormatting(
	ctx context.Context,
	desc catalog.TableDescriptor,
	exprStr string,
	semaCtx *tree.SemaContext,
	fmtFlags tree.FmtFlags,
) (tree.Expr, error) {
	__antithesis_instrumentation__.Notify(268200)
	expr, err := parser.ParseExpr(exprStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(268205)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268206)
	}
	__antithesis_instrumentation__.Notify(268201)

	replacedExpr, _, err := replaceColumnVars(desc, expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(268207)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268208)
	}
	__antithesis_instrumentation__.Notify(268202)

	typedExpr, err := replacedExpr.TypeCheck(ctx, semaCtx, types.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(268209)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268210)
	}
	__antithesis_instrumentation__.Notify(268203)

	if fmtFlags == tree.FmtPGCatalog {
		__antithesis_instrumentation__.Notify(268211)
		sanitizedExpr, err := SanitizeVarFreeExpr(ctx, expr, typedExpr.ResolvedType(), "FORMAT", semaCtx,
			tree.VolatilityImmutable)

		if err == nil {
			__antithesis_instrumentation__.Notify(268212)

			d, err := sanitizedExpr.Eval(&tree.EvalContext{})
			if err == nil {
				__antithesis_instrumentation__.Notify(268213)
				return d, nil
			} else {
				__antithesis_instrumentation__.Notify(268214)
			}
		} else {
			__antithesis_instrumentation__.Notify(268215)
		}
	} else {
		__antithesis_instrumentation__.Notify(268216)
	}
	__antithesis_instrumentation__.Notify(268204)

	return typedExpr, nil
}

type nameResolver struct {
	evalCtx    *tree.EvalContext
	tableID    descpb.ID
	source     *colinfo.DataSourceInfo
	nrc        *nameResolverIVarContainer
	ivarHelper *tree.IndexedVarHelper
}

func newNameResolver(
	evalCtx *tree.EvalContext, tableID descpb.ID, tn *tree.TableName, cols []catalog.Column,
) *nameResolver {
	__antithesis_instrumentation__.Notify(268217)
	source := colinfo.NewSourceInfoForSingleTable(
		*tn,
		colinfo.ResultColumnsFromColumns(tableID, cols),
	)
	nrc := &nameResolverIVarContainer{append(make([]catalog.Column, 0, len(cols)), cols...)}
	ivarHelper := tree.MakeIndexedVarHelper(nrc, len(cols))

	return &nameResolver{
		evalCtx:    evalCtx,
		tableID:    tableID,
		source:     source,
		nrc:        nrc,
		ivarHelper: &ivarHelper,
	}
}

func (nr *nameResolver) resolveNames(expr tree.Expr) (tree.Expr, error) {
	__antithesis_instrumentation__.Notify(268218)
	var v NameResolutionVisitor
	return ResolveNamesUsingVisitor(&v, expr, nr.source, *nr.ivarHelper, nr.evalCtx.SessionData().SearchPath)
}

func (nr *nameResolver) addColumn(col catalog.Column) {
	__antithesis_instrumentation__.Notify(268219)
	nr.ivarHelper.AppendSlot()
	nr.nrc.cols = append(nr.nrc.cols, col)
	newCols := colinfo.ResultColumnsFromColumns(nr.tableID, []catalog.Column{col})
	nr.source.SourceColumns = append(nr.source.SourceColumns, newCols...)
}

func (nr *nameResolver) addIVarContainerToSemaCtx(semaCtx *tree.SemaContext) {
	__antithesis_instrumentation__.Notify(268220)
	semaCtx.IVarContainer = nr.nrc
}

type nameResolverIVarContainer struct {
	cols []catalog.Column
}

func (nrc *nameResolverIVarContainer) IndexedVarEval(
	idx int, ctx *tree.EvalContext,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(268221)
	panic("unsupported")
}

func (nrc *nameResolverIVarContainer) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(268222)
	return nrc.cols[idx].GetType()
}

func (nrc *nameResolverIVarContainer) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(268223)
	return nil
}

func SanitizeVarFreeExpr(
	ctx context.Context,
	expr tree.Expr,
	expectedType *types.T,
	context string,
	semaCtx *tree.SemaContext,
	maxVolatility tree.Volatility,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(268224)
	if tree.ContainsVars(expr) {
		__antithesis_instrumentation__.Notify(268229)
		return nil, pgerror.Newf(pgcode.Syntax,
			"variable sub-expressions are not allowed in %s", context)
	} else {
		__antithesis_instrumentation__.Notify(268230)
	}
	__antithesis_instrumentation__.Notify(268225)

	defer semaCtx.Properties.Restore(semaCtx.Properties)

	flags := tree.RejectSpecial

	switch maxVolatility {
	case tree.VolatilityImmutable:
		__antithesis_instrumentation__.Notify(268231)
		flags |= tree.RejectStableOperators
		fallthrough

	case tree.VolatilityStable:
		__antithesis_instrumentation__.Notify(268232)
		flags |= tree.RejectVolatileFunctions

	case tree.VolatilityVolatile:
		__antithesis_instrumentation__.Notify(268233)

	default:
		__antithesis_instrumentation__.Notify(268234)
		panic(errors.AssertionFailedf("maxVolatility %s not supported", maxVolatility))
	}
	__antithesis_instrumentation__.Notify(268226)
	semaCtx.Properties.Require(context, flags)

	typedExpr, err := tree.TypeCheck(ctx, expr, semaCtx, expectedType)
	if err != nil {
		__antithesis_instrumentation__.Notify(268235)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(268236)
	}
	__antithesis_instrumentation__.Notify(268227)

	actualType := typedExpr.ResolvedType()
	if !expectedType.Equivalent(actualType) && func() bool {
		__antithesis_instrumentation__.Notify(268237)
		return typedExpr != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(268238)

		return nil, fmt.Errorf("expected %s expression to have type %s, but '%s' has type %s",
			context, expectedType, expr, actualType)
	} else {
		__antithesis_instrumentation__.Notify(268239)
	}
	__antithesis_instrumentation__.Notify(268228)
	return typedExpr, nil
}
