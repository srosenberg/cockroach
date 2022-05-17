package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"golang.org/x/text/language"
)

type SemaContext struct {
	Annotations Annotations

	Placeholders PlaceholderInfo

	IVarContainer IndexedVarContainer

	SearchPath sessiondata.SearchPath

	TypeResolver TypeReferenceResolver

	TableNameResolver QualifiedNameResolver

	IntervalStyleEnabled bool

	DateStyleEnabled bool

	Properties SemaProperties

	DateStyle pgdate.DateStyle

	IntervalStyle duration.IntervalStyle
}

type SemaProperties struct {
	required semaRequirements

	Derived ScalarProperties
}

type semaRequirements struct {
	context string

	rejectFlags SemaRejectFlags
}

func (s *SemaProperties) Require(context string, rejectFlags SemaRejectFlags) {
	__antithesis_instrumentation__.Notify(614737)
	s.required.context = context
	s.required.rejectFlags = rejectFlags
	s.Derived.Clear()
}

func (s *SemaProperties) IsSet(rejectFlags SemaRejectFlags) bool {
	__antithesis_instrumentation__.Notify(614738)
	return s.required.rejectFlags&rejectFlags != 0
}

func (s *SemaProperties) Restore(orig SemaProperties) {
	__antithesis_instrumentation__.Notify(614739)
	*s = orig
}

type SemaRejectFlags int

const (
	RejectAggregates SemaRejectFlags = 1 << iota

	RejectNestedAggregates

	RejectNestedWindowFunctions

	RejectWindowApplications

	RejectGenerators

	RejectNestedGenerators

	RejectStableOperators

	RejectVolatileFunctions

	RejectSubqueries

	RejectSpecial = RejectAggregates | RejectGenerators | RejectWindowApplications
)

type ScalarProperties struct {
	SeenAggregate bool

	SeenWindowApplication bool

	SeenGenerator bool

	inFuncExpr bool

	InWindowFunc bool
}

func (sp *ScalarProperties) Clear() {
	__antithesis_instrumentation__.Notify(614740)
	*sp = ScalarProperties{}
}

func MakeSemaContext() SemaContext {
	__antithesis_instrumentation__.Notify(614741)
	return SemaContext{}
}

func (sc *SemaContext) isUnresolvedPlaceholder(expr Expr) bool {
	__antithesis_instrumentation__.Notify(614742)
	if sc == nil {
		__antithesis_instrumentation__.Notify(614744)
		return false
	} else {
		__antithesis_instrumentation__.Notify(614745)
	}
	__antithesis_instrumentation__.Notify(614743)
	return sc.Placeholders.IsUnresolvedPlaceholder(expr)
}

func (sc *SemaContext) GetTypeResolver() TypeReferenceResolver {
	__antithesis_instrumentation__.Notify(614746)
	if sc == nil {
		__antithesis_instrumentation__.Notify(614748)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(614749)
	}
	__antithesis_instrumentation__.Notify(614747)
	return sc.TypeResolver
}

func placeholderTypeAmbiguityError(idx PlaceholderIdx) error {
	__antithesis_instrumentation__.Notify(614750)
	return errors.WithHint(
		pgerror.WithCandidateCode(
			&placeholderTypeAmbiguityErr{idx},
			pgcode.IndeterminateDatatype),
		"consider adding explicit type casts to the placeholder arguments",
	)
}

type placeholderTypeAmbiguityErr struct {
	idx PlaceholderIdx
}

func (err *placeholderTypeAmbiguityErr) Error() string {
	__antithesis_instrumentation__.Notify(614751)
	return fmt.Sprintf("could not determine data type of placeholder %s", err.idx)
}

func unexpectedTypeError(expr Expr, want, got *types.T) error {
	__antithesis_instrumentation__.Notify(614752)
	return pgerror.Newf(pgcode.InvalidParameterValue,
		"expected %s to be of type %s, found type %s", expr, errors.Safe(want), errors.Safe(got))
}

func decorateTypeCheckError(err error, format string, a ...interface{}) error {
	__antithesis_instrumentation__.Notify(614753)
	if !errors.HasType(err, (*placeholderTypeAmbiguityErr)(nil)) {
		__antithesis_instrumentation__.Notify(614755)
		return pgerror.Wrapf(err, pgcode.InvalidParameterValue, format, a...)
	} else {
		__antithesis_instrumentation__.Notify(614756)
	}
	__antithesis_instrumentation__.Notify(614754)
	return errors.WithStack(err)
}

func TypeCheck(
	ctx context.Context, expr Expr, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614757)
	if desired == nil {
		__antithesis_instrumentation__.Notify(614759)
		return nil, errors.AssertionFailedf(
			"the desired type for tree.TypeCheck cannot be nil, use types.Any instead: %T", expr)
	} else {
		__antithesis_instrumentation__.Notify(614760)
	}
	__antithesis_instrumentation__.Notify(614758)

	return expr.TypeCheck(ctx, semaCtx, desired)
}

func TypeCheckAndRequire(
	ctx context.Context, expr Expr, semaCtx *SemaContext, required *types.T, op string,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614761)
	typedExpr, err := TypeCheck(ctx, expr, semaCtx, required)
	if err != nil {
		__antithesis_instrumentation__.Notify(614764)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614765)
	}
	__antithesis_instrumentation__.Notify(614762)
	if typ := typedExpr.ResolvedType(); !(typ.Equivalent(required) || func() bool {
		__antithesis_instrumentation__.Notify(614766)
		return typ.Family() == types.UnknownFamily == true
	}() == true) {
		__antithesis_instrumentation__.Notify(614767)
		return typedExpr, pgerror.Newf(
			pgcode.DatatypeMismatch, "argument of %s must be type %s, not type %s", op, required, typ)
	} else {
		__antithesis_instrumentation__.Notify(614768)
	}
	__antithesis_instrumentation__.Notify(614763)
	return typedExpr, nil
}

func (expr *AndExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614769)
	leftTyped, err := typeCheckAndRequireBoolean(ctx, semaCtx, expr.Left, "AND argument")
	if err != nil {
		__antithesis_instrumentation__.Notify(614772)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614773)
	}
	__antithesis_instrumentation__.Notify(614770)
	rightTyped, err := typeCheckAndRequireBoolean(ctx, semaCtx, expr.Right, "AND argument")
	if err != nil {
		__antithesis_instrumentation__.Notify(614774)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614775)
	}
	__antithesis_instrumentation__.Notify(614771)
	expr.Left, expr.Right = leftTyped, rightTyped
	expr.typ = types.Bool
	return expr, nil
}

func (expr *BinaryExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614776)
	ops := BinOps[expr.Operator.Symbol]

	typedSubExprs, fns, err := typeCheckOverloadedExprs(ctx, semaCtx, desired, ops, true, expr.Left, expr.Right)
	if err != nil {
		__antithesis_instrumentation__.Notify(614782)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614783)
	}
	__antithesis_instrumentation__.Notify(614777)

	leftTyped, rightTyped := typedSubExprs[0], typedSubExprs[1]
	leftReturn := leftTyped.ResolvedType()
	rightReturn := rightTyped.ResolvedType()

	if leftReturn.Family() == types.UnknownFamily || func() bool {
		__antithesis_instrumentation__.Notify(614784)
		return rightReturn.Family() == types.UnknownFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(614785)
		if len(fns) > 0 {
			__antithesis_instrumentation__.Notify(614786)
			noneAcceptNull := true
			for _, e := range fns {
				__antithesis_instrumentation__.Notify(614788)
				if e.(*BinOp).NullableArgs {
					__antithesis_instrumentation__.Notify(614789)
					noneAcceptNull = false
					break
				} else {
					__antithesis_instrumentation__.Notify(614790)
				}
			}
			__antithesis_instrumentation__.Notify(614787)
			if noneAcceptNull {
				__antithesis_instrumentation__.Notify(614791)
				return DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(614792)
			}
		} else {
			__antithesis_instrumentation__.Notify(614793)
		}
	} else {
		__antithesis_instrumentation__.Notify(614794)
	}
	__antithesis_instrumentation__.Notify(614778)

	if len(fns) != 1 {
		__antithesis_instrumentation__.Notify(614795)
		var desStr string
		if desired.Family() != types.AnyFamily {
			__antithesis_instrumentation__.Notify(614798)
			desStr = fmt.Sprintf(" (desired <%s>)", desired)
		} else {
			__antithesis_instrumentation__.Notify(614799)
		}
		__antithesis_instrumentation__.Notify(614796)
		sig := fmt.Sprintf("<%s> %s <%s>%s", leftReturn, expr.Operator, rightReturn, desStr)
		if len(fns) == 0 {
			__antithesis_instrumentation__.Notify(614800)
			return nil,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedBinaryOpErrFmt, sig)
		} else {
			__antithesis_instrumentation__.Notify(614801)
		}
		__antithesis_instrumentation__.Notify(614797)
		fnsStr := formatCandidates(expr.Operator.String(), fns)
		err = pgerror.Newf(pgcode.AmbiguousFunction, ambiguousBinaryOpErrFmt, sig)
		err = errors.WithHintf(err, candidatesHintFmt, fnsStr)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614802)
	}
	__antithesis_instrumentation__.Notify(614779)

	binOp := fns[0].(*BinOp)
	if err := semaCtx.checkVolatility(binOp.Volatility); err != nil {
		__antithesis_instrumentation__.Notify(614803)
		return nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "%s", expr.Operator)
	} else {
		__antithesis_instrumentation__.Notify(614804)
	}
	__antithesis_instrumentation__.Notify(614780)

	if binOp.counter != nil {
		__antithesis_instrumentation__.Notify(614805)
		telemetry.Inc(binOp.counter)
	} else {
		__antithesis_instrumentation__.Notify(614806)
	}
	__antithesis_instrumentation__.Notify(614781)

	expr.Left, expr.Right = leftTyped, rightTyped
	expr.Fn = binOp
	expr.typ = binOp.returnType()(typedSubExprs)
	return expr, nil
}

func (expr *CaseExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614807)
	var err error
	tmpExprs := make([]Expr, 0, len(expr.Whens)+1)
	if expr.Expr != nil {
		__antithesis_instrumentation__.Notify(614814)
		tmpExprs = tmpExprs[:0]
		tmpExprs = append(tmpExprs, expr.Expr)
		for _, when := range expr.Whens {
			__antithesis_instrumentation__.Notify(614817)
			tmpExprs = append(tmpExprs, when.Cond)
		}
		__antithesis_instrumentation__.Notify(614815)

		typedSubExprs, _, err := TypeCheckSameTypedExprs(ctx, semaCtx, types.Any, tmpExprs...)
		if err != nil {
			__antithesis_instrumentation__.Notify(614818)
			return nil, decorateTypeCheckError(err, "incompatible condition type:")
		} else {
			__antithesis_instrumentation__.Notify(614819)
		}
		__antithesis_instrumentation__.Notify(614816)
		expr.Expr = typedSubExprs[0]
		for i, whenCond := range typedSubExprs[1:] {
			__antithesis_instrumentation__.Notify(614820)
			expr.Whens[i].Cond = whenCond
		}
	} else {
		__antithesis_instrumentation__.Notify(614821)

		for i, when := range expr.Whens {
			__antithesis_instrumentation__.Notify(614822)
			typedCond, err := typeCheckAndRequireBoolean(ctx, semaCtx, when.Cond, "condition")
			if err != nil {
				__antithesis_instrumentation__.Notify(614824)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(614825)
			}
			__antithesis_instrumentation__.Notify(614823)
			expr.Whens[i].Cond = typedCond
		}
	}
	__antithesis_instrumentation__.Notify(614808)

	tmpExprs = tmpExprs[:0]
	for _, when := range expr.Whens {
		__antithesis_instrumentation__.Notify(614826)
		tmpExprs = append(tmpExprs, when.Val)
	}
	__antithesis_instrumentation__.Notify(614809)
	if expr.Else != nil {
		__antithesis_instrumentation__.Notify(614827)
		tmpExprs = append(tmpExprs, expr.Else)
	} else {
		__antithesis_instrumentation__.Notify(614828)
	}
	__antithesis_instrumentation__.Notify(614810)
	typedSubExprs, retType, err := TypeCheckSameTypedExprs(ctx, semaCtx, desired, tmpExprs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(614829)
		return nil, decorateTypeCheckError(err, "incompatible value type")
	} else {
		__antithesis_instrumentation__.Notify(614830)
	}
	__antithesis_instrumentation__.Notify(614811)
	if expr.Else != nil {
		__antithesis_instrumentation__.Notify(614831)
		expr.Else = typedSubExprs[len(typedSubExprs)-1]
		typedSubExprs = typedSubExprs[:len(typedSubExprs)-1]
	} else {
		__antithesis_instrumentation__.Notify(614832)
	}
	__antithesis_instrumentation__.Notify(614812)
	for i, whenVal := range typedSubExprs {
		__antithesis_instrumentation__.Notify(614833)
		expr.Whens[i].Val = whenVal
	}
	__antithesis_instrumentation__.Notify(614813)
	expr.typ = retType
	return expr, nil
}

func invalidCastError(castFrom, castTo *types.T) error {
	__antithesis_instrumentation__.Notify(614834)
	return pgerror.Newf(pgcode.CannotCoerce, "invalid cast: %s -> %s", castFrom, castTo)
}

func resolveCast(
	context string,
	castFrom, castTo *types.T,
	allowStable bool,
	intervalStyleEnabled bool,
	dateStyleEnabled bool,
) error {
	__antithesis_instrumentation__.Notify(614835)
	toFamily := castTo.Family()
	fromFamily := castFrom.Family()
	switch {
	case toFamily == types.ArrayFamily && func() bool {
		__antithesis_instrumentation__.Notify(614847)
		return fromFamily == types.ArrayFamily == true
	}() == true:
		__antithesis_instrumentation__.Notify(614836)
		err := resolveCast(
			context,
			castFrom.ArrayContents(),
			castTo.ArrayContents(),
			allowStable,
			intervalStyleEnabled,
			dateStyleEnabled,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(614848)
			return err
		} else {
			__antithesis_instrumentation__.Notify(614849)
		}
		__antithesis_instrumentation__.Notify(614837)
		telemetry.Inc(GetCastCounter(fromFamily, toFamily))
		return nil

	case toFamily == types.EnumFamily && func() bool {
		__antithesis_instrumentation__.Notify(614850)
		return fromFamily == types.EnumFamily == true
	}() == true:
		__antithesis_instrumentation__.Notify(614838)

		if !castFrom.Equivalent(castTo) {
			__antithesis_instrumentation__.Notify(614851)
			return invalidCastError(castFrom, castTo)
		} else {
			__antithesis_instrumentation__.Notify(614852)
		}
		__antithesis_instrumentation__.Notify(614839)
		telemetry.Inc(GetCastCounter(fromFamily, toFamily))
		return nil

	case toFamily == types.TupleFamily && func() bool {
		__antithesis_instrumentation__.Notify(614853)
		return fromFamily == types.TupleFamily == true
	}() == true:
		__antithesis_instrumentation__.Notify(614840)

		if castTo == types.AnyTuple {
			__antithesis_instrumentation__.Notify(614854)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(614855)
		}
		__antithesis_instrumentation__.Notify(614841)
		fromTuple := castFrom.TupleContents()
		toTuple := castTo.TupleContents()
		if len(fromTuple) != len(toTuple) {
			__antithesis_instrumentation__.Notify(614856)
			return invalidCastError(castFrom, castTo)
		} else {
			__antithesis_instrumentation__.Notify(614857)
		}
		__antithesis_instrumentation__.Notify(614842)
		for i, from := range fromTuple {
			__antithesis_instrumentation__.Notify(614858)
			to := toTuple[i]
			err := resolveCast(
				context,
				from,
				to,
				allowStable,
				intervalStyleEnabled,
				dateStyleEnabled,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(614859)
				return err
			} else {
				__antithesis_instrumentation__.Notify(614860)
			}
		}
		__antithesis_instrumentation__.Notify(614843)
		telemetry.Inc(GetCastCounter(fromFamily, toFamily))
		return nil

	default:
		__antithesis_instrumentation__.Notify(614844)
		cast, ok := lookupCast(castFrom, castTo, intervalStyleEnabled, dateStyleEnabled)
		if !ok {
			__antithesis_instrumentation__.Notify(614861)
			return invalidCastError(castFrom, castTo)
		} else {
			__antithesis_instrumentation__.Notify(614862)
		}
		__antithesis_instrumentation__.Notify(614845)
		if !allowStable && func() bool {
			__antithesis_instrumentation__.Notify(614863)
			return cast.volatility >= VolatilityStable == true
		}() == true {
			__antithesis_instrumentation__.Notify(614864)
			err := NewContextDependentOpsNotAllowedError(context)
			err = pgerror.Wrapf(err, pgcode.InvalidParameterValue, "%s::%s", castFrom, castTo)
			if cast.volatilityHint != "" {
				__antithesis_instrumentation__.Notify(614866)
				err = errors.WithHint(err, cast.volatilityHint)
			} else {
				__antithesis_instrumentation__.Notify(614867)
			}
			__antithesis_instrumentation__.Notify(614865)
			return err
		} else {
			__antithesis_instrumentation__.Notify(614868)
		}
		__antithesis_instrumentation__.Notify(614846)
		telemetry.Inc(GetCastCounter(fromFamily, toFamily))
		return nil
	}
}

func isArrayExpr(expr Expr) bool {
	__antithesis_instrumentation__.Notify(614869)
	_, ok := expr.(*Array)
	return ok
}

func (expr *CastExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, _ *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614870)

	desired := types.Any
	exprType, err := ResolveType(ctx, expr.Type, semaCtx.GetTypeResolver())
	if err != nil {
		__antithesis_instrumentation__.Notify(614877)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614878)
	}
	__antithesis_instrumentation__.Notify(614871)
	expr.Type = exprType
	canElideCast := true
	switch {
	case isConstant(expr.Expr):
		__antithesis_instrumentation__.Notify(614879)
		c := expr.Expr.(Constant)
		if canConstantBecome(c, exprType) {
			__antithesis_instrumentation__.Notify(614883)

			desired = exprType
		} else {
			__antithesis_instrumentation__.Notify(614884)
		}
	case semaCtx.isUnresolvedPlaceholder(expr.Expr):
		__antithesis_instrumentation__.Notify(614880)

		desired = exprType
	case isArrayExpr(expr.Expr):
		__antithesis_instrumentation__.Notify(614881)

		if exprType.Family() == types.ArrayFamily {
			__antithesis_instrumentation__.Notify(614885)

			contents := exprType.ArrayContents()
			if baseType, ok := types.OidToType[contents.Oid()]; ok && func() bool {
				__antithesis_instrumentation__.Notify(614887)
				return !baseType.Identical(contents) == true
			}() == true {
				__antithesis_instrumentation__.Notify(614888)
				canElideCast = false
			} else {
				__antithesis_instrumentation__.Notify(614889)
			}
			__antithesis_instrumentation__.Notify(614886)
			desired = exprType
		} else {
			__antithesis_instrumentation__.Notify(614890)
		}
	default:
		__antithesis_instrumentation__.Notify(614882)
	}
	__antithesis_instrumentation__.Notify(614872)

	typedSubExpr, err := expr.Expr.TypeCheck(ctx, semaCtx, desired)
	if err != nil {
		__antithesis_instrumentation__.Notify(614891)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614892)
	}
	__antithesis_instrumentation__.Notify(614873)

	if canElideCast && func() bool {
		__antithesis_instrumentation__.Notify(614893)
		return typedSubExpr.ResolvedType().Identical(exprType) == true
	}() == true {
		__antithesis_instrumentation__.Notify(614894)
		return typedSubExpr, nil
	} else {
		__antithesis_instrumentation__.Notify(614895)
	}
	__antithesis_instrumentation__.Notify(614874)

	castFrom := typedSubExpr.ResolvedType()
	allowStable := true
	context := ""
	if semaCtx != nil && func() bool {
		__antithesis_instrumentation__.Notify(614896)
		return semaCtx.Properties.required.rejectFlags&RejectStableOperators != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(614897)
		allowStable = false
		context = semaCtx.Properties.required.context
	} else {
		__antithesis_instrumentation__.Notify(614898)
	}
	__antithesis_instrumentation__.Notify(614875)
	err = resolveCast(
		context,
		castFrom,
		exprType,
		allowStable,
		semaCtx != nil && func() bool {
			__antithesis_instrumentation__.Notify(614899)
			return semaCtx.IntervalStyleEnabled == true
		}() == true,
		semaCtx != nil && func() bool {
			__antithesis_instrumentation__.Notify(614900)
			return semaCtx.DateStyleEnabled == true
		}() == true,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(614901)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614902)
	}
	__antithesis_instrumentation__.Notify(614876)
	expr.Expr = typedSubExpr
	expr.Type = exprType
	expr.typ = exprType
	return expr, nil
}

func (expr *IndirectionExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614903)
	for i, t := range expr.Indirection {
		__antithesis_instrumentation__.Notify(614907)
		if t.Slice {
			__antithesis_instrumentation__.Notify(614911)
			return nil, unimplemented.NewWithIssuef(32551, "ARRAY slicing in %s", expr)
		} else {
			__antithesis_instrumentation__.Notify(614912)
		}
		__antithesis_instrumentation__.Notify(614908)
		if i > 0 {
			__antithesis_instrumentation__.Notify(614913)
			return nil, unimplemented.NewWithIssueDetailf(32552, "ind", "multidimensional indexing: %s", expr)
		} else {
			__antithesis_instrumentation__.Notify(614914)
		}
		__antithesis_instrumentation__.Notify(614909)

		beginExpr, err := typeCheckAndRequire(ctx, semaCtx, t.Begin, types.Int, "ARRAY subscript")
		if err != nil {
			__antithesis_instrumentation__.Notify(614915)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(614916)
		}
		__antithesis_instrumentation__.Notify(614910)
		t.Begin = beginExpr
	}
	__antithesis_instrumentation__.Notify(614904)

	subExpr, err := expr.Expr.TypeCheck(ctx, semaCtx, types.MakeArray(desired))
	if err != nil {
		__antithesis_instrumentation__.Notify(614917)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614918)
	}
	__antithesis_instrumentation__.Notify(614905)
	typ := subExpr.ResolvedType()
	if typ.Family() != types.ArrayFamily {
		__antithesis_instrumentation__.Notify(614919)
		return nil, pgerror.Newf(pgcode.DatatypeMismatch, "cannot subscript type %s because it is not an array", typ)
	} else {
		__antithesis_instrumentation__.Notify(614920)
	}
	__antithesis_instrumentation__.Notify(614906)
	expr.Expr = subExpr
	expr.typ = typ.ArrayContents()

	telemetry.Inc(sqltelemetry.ArraySubscriptCounter)
	return expr, nil
}

func (expr *AnnotateTypeExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614921)
	annotateType, err := ResolveType(ctx, expr.Type, semaCtx.GetTypeResolver())
	if err != nil {
		__antithesis_instrumentation__.Notify(614924)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614925)
	}
	__antithesis_instrumentation__.Notify(614922)
	expr.Type = annotateType
	subExpr, err := typeCheckAndRequire(
		ctx,
		semaCtx,
		expr.Expr,
		annotateType,
		fmt.Sprintf(
			"type annotation for %v as %s, found",
			expr.Expr,
			annotateType,
		),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(614926)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614927)
	}
	__antithesis_instrumentation__.Notify(614923)
	return subExpr, nil
}

func (expr *CollateExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614928)
	if strings.ToLower(expr.Locale) == DefaultCollationTag {
		__antithesis_instrumentation__.Notify(614933)
		return nil, errors.WithHint(
			unimplemented.NewWithIssuef(
				57255,
				"DEFAULT collations are not supported",
			),
			`omit the 'COLLATE "default"' clause in your statement`,
		)
	} else {
		__antithesis_instrumentation__.Notify(614934)
	}
	__antithesis_instrumentation__.Notify(614929)
	_, err := language.Parse(expr.Locale)
	if err != nil {
		__antithesis_instrumentation__.Notify(614935)
		return nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue,
			"invalid locale %s", expr.Locale)
	} else {
		__antithesis_instrumentation__.Notify(614936)
	}
	__antithesis_instrumentation__.Notify(614930)
	subExpr, err := expr.Expr.TypeCheck(ctx, semaCtx, types.String)
	if err != nil {
		__antithesis_instrumentation__.Notify(614937)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614938)
	}
	__antithesis_instrumentation__.Notify(614931)
	t := subExpr.ResolvedType()
	if types.IsStringType(t) {
		__antithesis_instrumentation__.Notify(614939)
		expr.Expr = subExpr
		expr.typ = types.MakeCollatedString(t, expr.Locale)
		return expr, nil
	} else {
		__antithesis_instrumentation__.Notify(614940)
		if t.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(614941)
			expr.Expr = subExpr
			expr.typ = types.MakeCollatedString(types.String, expr.Locale)
			return expr, nil
		} else {
			__antithesis_instrumentation__.Notify(614942)
		}
	}
	__antithesis_instrumentation__.Notify(614932)
	return nil, pgerror.Newf(pgcode.DatatypeMismatch,
		"incompatible type for COLLATE: %s", t)
}

func NewTypeIsNotCompositeError(resolvedType *types.T) error {
	__antithesis_instrumentation__.Notify(614943)
	return pgerror.Newf(pgcode.WrongObjectType,
		"type %s is not composite", resolvedType,
	)
}

func (expr *TupleStar) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614944)
	subExpr, err := expr.Expr.TypeCheck(ctx, semaCtx, desired)
	if err != nil {
		__antithesis_instrumentation__.Notify(614947)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614948)
	}
	__antithesis_instrumentation__.Notify(614945)
	expr.Expr = subExpr
	resolvedType := subExpr.ResolvedType()

	if resolvedType.Family() != types.TupleFamily {
		__antithesis_instrumentation__.Notify(614949)
		return nil, NewTypeIsNotCompositeError(resolvedType)
	} else {
		__antithesis_instrumentation__.Notify(614950)
	}
	__antithesis_instrumentation__.Notify(614946)

	return subExpr, err
}

func (expr *TupleStar) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(614951)
	return expr.Expr.(TypedExpr).ResolvedType()
}

func (expr *ColumnAccessExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614952)

	subExpr, err := expr.Expr.TypeCheck(ctx, semaCtx, types.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(614958)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614959)
	}
	__antithesis_instrumentation__.Notify(614953)

	expr.Expr = subExpr
	resolvedType := subExpr.ResolvedType()

	if resolvedType.Family() != types.TupleFamily {
		__antithesis_instrumentation__.Notify(614960)
		return nil, NewTypeIsNotCompositeError(resolvedType)
	} else {
		__antithesis_instrumentation__.Notify(614961)
	}
	__antithesis_instrumentation__.Notify(614954)

	if !expr.ByIndex && func() bool {
		__antithesis_instrumentation__.Notify(614962)
		return len(resolvedType.TupleLabels()) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(614963)
		return nil, pgerror.Newf(pgcode.UndefinedColumn, "could not identify column %q in record data type",
			expr.ColName)
	} else {
		__antithesis_instrumentation__.Notify(614964)
	}
	__antithesis_instrumentation__.Notify(614955)

	if expr.ByIndex {
		__antithesis_instrumentation__.Notify(614965)

		if expr.ColIndex < 0 || func() bool {
			__antithesis_instrumentation__.Notify(614966)
			return expr.ColIndex >= len(resolvedType.TupleContents()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(614967)
			return nil, pgerror.Newf(pgcode.Syntax, "tuple column %d does not exist", expr.ColIndex+1)
		} else {
			__antithesis_instrumentation__.Notify(614968)
		}
	} else {
		__antithesis_instrumentation__.Notify(614969)

		expr.ColIndex = -1
		for i, label := range resolvedType.TupleLabels() {
			__antithesis_instrumentation__.Notify(614971)
			if label == string(expr.ColName) {
				__antithesis_instrumentation__.Notify(614972)
				if expr.ColIndex != -1 {
					__antithesis_instrumentation__.Notify(614974)

					return nil, pgerror.Newf(pgcode.AmbiguousColumn, "column reference %q is ambiguous", label)
				} else {
					__antithesis_instrumentation__.Notify(614975)
				}
				__antithesis_instrumentation__.Notify(614973)
				expr.ColIndex = i
			} else {
				__antithesis_instrumentation__.Notify(614976)
			}
		}
		__antithesis_instrumentation__.Notify(614970)
		if expr.ColIndex < 0 {
			__antithesis_instrumentation__.Notify(614977)
			return nil, pgerror.Newf(pgcode.UndefinedColumn,
				"could not identify column %q in %s",
				ErrString(&expr.ColName), resolvedType,
			)
		} else {
			__antithesis_instrumentation__.Notify(614978)
		}
	}
	__antithesis_instrumentation__.Notify(614956)

	if tExpr, ok := expr.Expr.(*Tuple); ok {
		__antithesis_instrumentation__.Notify(614979)
		return tExpr.Exprs[expr.ColIndex].(TypedExpr), nil
	} else {
		__antithesis_instrumentation__.Notify(614980)
	}
	__antithesis_instrumentation__.Notify(614957)

	expr.typ = resolvedType.TupleContents()[expr.ColIndex]
	return expr, nil
}

func (expr *CoalesceExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614981)
	typedSubExprs, retType, err := TypeCheckSameTypedExprs(ctx, semaCtx, desired, expr.Exprs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(614984)
		return nil, decorateTypeCheckError(err, "incompatible %s expressions", redact.Safe(expr.Name))
	} else {
		__antithesis_instrumentation__.Notify(614985)
	}
	__antithesis_instrumentation__.Notify(614982)

	for i, subExpr := range typedSubExprs {
		__antithesis_instrumentation__.Notify(614986)
		expr.Exprs[i] = subExpr
	}
	__antithesis_instrumentation__.Notify(614983)
	expr.typ = retType
	return expr, nil
}

func (expr *ComparisonExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(614987)
	var leftTyped, rightTyped TypedExpr
	var cmpOp *CmpOp
	var alwaysNull bool
	var err error
	if expr.Operator.Symbol.HasSubOperator() {
		__antithesis_instrumentation__.Notify(614993)
		leftTyped, rightTyped, cmpOp, alwaysNull, err = typeCheckComparisonOpWithSubOperator(
			ctx,
			semaCtx,
			expr.Operator,
			expr.SubOperator,
			expr.Left,
			expr.Right,
		)
	} else {
		__antithesis_instrumentation__.Notify(614994)
		leftTyped, rightTyped, cmpOp, alwaysNull, err = typeCheckComparisonOp(
			ctx,
			semaCtx,
			expr.Operator,
			expr.Left,
			expr.Right,
		)
	}
	__antithesis_instrumentation__.Notify(614988)
	if err != nil {
		__antithesis_instrumentation__.Notify(614995)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(614996)
	}
	__antithesis_instrumentation__.Notify(614989)

	if alwaysNull {
		__antithesis_instrumentation__.Notify(614997)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(614998)
	}
	__antithesis_instrumentation__.Notify(614990)

	if err := semaCtx.checkVolatility(cmpOp.Volatility); err != nil {
		__antithesis_instrumentation__.Notify(614999)
		return nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "%s", expr.Operator)
	} else {
		__antithesis_instrumentation__.Notify(615000)
	}
	__antithesis_instrumentation__.Notify(614991)

	if cmpOp.counter != nil {
		__antithesis_instrumentation__.Notify(615001)
		telemetry.Inc(cmpOp.counter)
	} else {
		__antithesis_instrumentation__.Notify(615002)
	}
	__antithesis_instrumentation__.Notify(614992)

	expr.Left, expr.Right = leftTyped, rightTyped
	expr.Fn = cmpOp
	expr.typ = types.Bool
	return expr, nil
}

var (
	errStarNotAllowed      = pgerror.New(pgcode.Syntax, "cannot use \"*\" in this context")
	errInvalidDefaultUsage = pgerror.New(pgcode.Syntax, "DEFAULT can only appear in a VALUES list within INSERT or on the right side of a SET")
	errInvalidMaxUsage     = pgerror.New(pgcode.Syntax, "MAXVALUE can only appear within a range partition expression")
	errInvalidMinUsage     = pgerror.New(pgcode.Syntax, "MINVALUE can only appear within a range partition expression")
	errPrivateFunction     = pgerror.New(pgcode.ReservedName, "function reserved for internal use")
)

func NewAggInAggError() error {
	__antithesis_instrumentation__.Notify(615003)
	return pgerror.Newf(pgcode.Grouping, "aggregate function calls cannot be nested")
}

func NewInvalidNestedSRFError(context string) error {
	__antithesis_instrumentation__.Notify(615004)
	return pgerror.Newf(pgcode.FeatureNotSupported,
		"set-returning functions must appear at the top level of %s", context)
}

func NewInvalidFunctionUsageError(class FunctionClass, context string) error {
	__antithesis_instrumentation__.Notify(615005)
	var cat string
	var code pgcode.Code
	switch class {
	case AggregateClass:
		__antithesis_instrumentation__.Notify(615007)
		cat = "aggregate"
		code = pgcode.Grouping
	case WindowClass:
		__antithesis_instrumentation__.Notify(615008)
		cat = "window"
		code = pgcode.Windowing
	case GeneratorClass:
		__antithesis_instrumentation__.Notify(615009)
		cat = "generator"
		code = pgcode.FeatureNotSupported
	default:
		__antithesis_instrumentation__.Notify(615010)
	}
	__antithesis_instrumentation__.Notify(615006)
	return pgerror.Newf(code, "%s functions are not allowed in %s", cat, context)
}

func (sc *SemaContext) checkFunctionUsage(expr *FuncExpr, def *FunctionDefinition) error {
	__antithesis_instrumentation__.Notify(615011)
	if def.UnsupportedWithIssue != 0 {
		__antithesis_instrumentation__.Notify(615017)

		const msg = "this function is not yet supported"
		if def.UnsupportedWithIssue < 0 {
			__antithesis_instrumentation__.Notify(615019)
			return unimplemented.New(def.Name+"()", msg)
		} else {
			__antithesis_instrumentation__.Notify(615020)
		}
		__antithesis_instrumentation__.Notify(615018)
		return unimplemented.NewWithIssueDetail(def.UnsupportedWithIssue, def.Name, msg)
	} else {
		__antithesis_instrumentation__.Notify(615021)
	}
	__antithesis_instrumentation__.Notify(615012)
	if def.Private {
		__antithesis_instrumentation__.Notify(615022)
		return pgerror.Wrapf(errPrivateFunction, pgcode.ReservedName,
			"%s()", errors.Safe(def.Name))
	} else {
		__antithesis_instrumentation__.Notify(615023)
	}
	__antithesis_instrumentation__.Notify(615013)
	if sc == nil {
		__antithesis_instrumentation__.Notify(615024)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(615025)
	}
	__antithesis_instrumentation__.Notify(615014)

	if expr.IsWindowFunctionApplication() {
		__antithesis_instrumentation__.Notify(615026)
		if sc.Properties.required.rejectFlags&RejectWindowApplications != 0 {
			__antithesis_instrumentation__.Notify(615029)
			return NewInvalidFunctionUsageError(WindowClass, sc.Properties.required.context)
		} else {
			__antithesis_instrumentation__.Notify(615030)
		}
		__antithesis_instrumentation__.Notify(615027)

		if sc.Properties.Derived.InWindowFunc && func() bool {
			__antithesis_instrumentation__.Notify(615031)
			return sc.Properties.required.rejectFlags&RejectNestedWindowFunctions != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(615032)
			return pgerror.Newf(pgcode.Windowing, "window function calls cannot be nested")
		} else {
			__antithesis_instrumentation__.Notify(615033)
		}
		__antithesis_instrumentation__.Notify(615028)
		sc.Properties.Derived.SeenWindowApplication = true
	} else {
		__antithesis_instrumentation__.Notify(615034)

		if def.Class == AggregateClass {
			__antithesis_instrumentation__.Notify(615035)
			if sc.Properties.Derived.inFuncExpr && func() bool {
				__antithesis_instrumentation__.Notify(615038)
				return sc.Properties.required.rejectFlags&RejectNestedAggregates != 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(615039)
				return NewAggInAggError()
			} else {
				__antithesis_instrumentation__.Notify(615040)
			}
			__antithesis_instrumentation__.Notify(615036)
			if sc.Properties.required.rejectFlags&RejectAggregates != 0 {
				__antithesis_instrumentation__.Notify(615041)
				return NewInvalidFunctionUsageError(AggregateClass, sc.Properties.required.context)
			} else {
				__antithesis_instrumentation__.Notify(615042)
			}
			__antithesis_instrumentation__.Notify(615037)
			sc.Properties.Derived.SeenAggregate = true
		} else {
			__antithesis_instrumentation__.Notify(615043)
		}
	}
	__antithesis_instrumentation__.Notify(615015)
	if def.Class == GeneratorClass {
		__antithesis_instrumentation__.Notify(615044)
		if sc.Properties.Derived.inFuncExpr && func() bool {
			__antithesis_instrumentation__.Notify(615047)
			return sc.Properties.required.rejectFlags&RejectNestedGenerators != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(615048)
			return NewInvalidNestedSRFError(sc.Properties.required.context)
		} else {
			__antithesis_instrumentation__.Notify(615049)
		}
		__antithesis_instrumentation__.Notify(615045)
		if sc.Properties.required.rejectFlags&RejectGenerators != 0 {
			__antithesis_instrumentation__.Notify(615050)
			return NewInvalidFunctionUsageError(GeneratorClass, sc.Properties.required.context)
		} else {
			__antithesis_instrumentation__.Notify(615051)
		}
		__antithesis_instrumentation__.Notify(615046)
		sc.Properties.Derived.SeenGenerator = true
	} else {
		__antithesis_instrumentation__.Notify(615052)
	}
	__antithesis_instrumentation__.Notify(615016)
	return nil
}

func NewContextDependentOpsNotAllowedError(context string) error {
	__antithesis_instrumentation__.Notify(615053)

	return pgerror.Newf(pgcode.FeatureNotSupported,
		"context-dependent operators are not allowed in %s", context,
	)
}

func (sc *SemaContext) checkVolatility(v Volatility) error {
	__antithesis_instrumentation__.Notify(615054)
	if sc == nil {
		__antithesis_instrumentation__.Notify(615057)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(615058)
	}
	__antithesis_instrumentation__.Notify(615055)
	switch v {
	case VolatilityVolatile:
		__antithesis_instrumentation__.Notify(615059)
		if sc.Properties.required.rejectFlags&RejectVolatileFunctions != 0 {
			__antithesis_instrumentation__.Notify(615062)

			return pgerror.Newf(pgcode.FeatureNotSupported,
				"volatile functions are not allowed in %s", sc.Properties.required.context)
		} else {
			__antithesis_instrumentation__.Notify(615063)
		}
	case VolatilityStable:
		__antithesis_instrumentation__.Notify(615060)
		if sc.Properties.required.rejectFlags&RejectStableOperators != 0 {
			__antithesis_instrumentation__.Notify(615064)
			return NewContextDependentOpsNotAllowedError(sc.Properties.required.context)
		} else {
			__antithesis_instrumentation__.Notify(615065)
		}
	default:
		__antithesis_instrumentation__.Notify(615061)
	}
	__antithesis_instrumentation__.Notify(615056)
	return nil
}

func CheckIsWindowOrAgg(def *FunctionDefinition) error {
	__antithesis_instrumentation__.Notify(615066)
	switch def.Class {
	case AggregateClass:
		__antithesis_instrumentation__.Notify(615068)
	case WindowClass:
		__antithesis_instrumentation__.Notify(615069)
	default:
		__antithesis_instrumentation__.Notify(615070)
		return pgerror.Newf(pgcode.WrongObjectType,
			"OVER specified, but %s() is neither a window function nor an aggregate function",
			def.Name)
	}
	__antithesis_instrumentation__.Notify(615067)
	return nil
}

func (expr *FuncExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615071)
	var searchPath sessiondata.SearchPath
	if semaCtx != nil {
		__antithesis_instrumentation__.Notify(615087)
		searchPath = semaCtx.SearchPath
	} else {
		__antithesis_instrumentation__.Notify(615088)
	}
	__antithesis_instrumentation__.Notify(615072)
	def, err := expr.Func.Resolve(searchPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(615089)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615090)
	}
	__antithesis_instrumentation__.Notify(615073)

	if err := semaCtx.checkFunctionUsage(expr, def); err != nil {
		__antithesis_instrumentation__.Notify(615091)
		return nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue,
			"%s()", def.Name)
	} else {
		__antithesis_instrumentation__.Notify(615092)
	}
	__antithesis_instrumentation__.Notify(615074)
	if semaCtx != nil {
		__antithesis_instrumentation__.Notify(615093)

		defer func(semaCtx *SemaContext, prevFunc bool, prevWindow bool) {
			__antithesis_instrumentation__.Notify(615095)
			semaCtx.Properties.Derived.inFuncExpr = prevFunc
			semaCtx.Properties.Derived.InWindowFunc = prevWindow
		}(
			semaCtx,
			semaCtx.Properties.Derived.inFuncExpr,
			semaCtx.Properties.Derived.InWindowFunc,
		)
		__antithesis_instrumentation__.Notify(615094)
		semaCtx.Properties.Derived.inFuncExpr = true
		if expr.WindowDef != nil {
			__antithesis_instrumentation__.Notify(615096)
			semaCtx.Properties.Derived.InWindowFunc = true
		} else {
			__antithesis_instrumentation__.Notify(615097)
		}
	} else {
		__antithesis_instrumentation__.Notify(615098)
	}
	__antithesis_instrumentation__.Notify(615075)

	typedSubExprs, fns, err := typeCheckOverloadedExprs(ctx, semaCtx, desired, def.Definition, false, expr.Exprs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(615099)
		return nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "%s()", def.Name)
	} else {
		__antithesis_instrumentation__.Notify(615100)
	}
	__antithesis_instrumentation__.Notify(615076)

	if !def.NullableArgs && func() bool {
		__antithesis_instrumentation__.Notify(615101)
		return def.FunctionProperties.Class == AggregateClass == true
	}() == true {
		__antithesis_instrumentation__.Notify(615102)
		for i := range typedSubExprs {
			__antithesis_instrumentation__.Notify(615103)
			if typedSubExprs[i].ResolvedType().Family() == types.UnknownFamily {
				__antithesis_instrumentation__.Notify(615104)
				var filtered []overloadImpl
				for j := range fns {
					__antithesis_instrumentation__.Notify(615106)
					if fns[j].params().GetAt(i).Equivalent(types.String) {
						__antithesis_instrumentation__.Notify(615107)
						if filtered == nil {
							__antithesis_instrumentation__.Notify(615109)
							filtered = make([]overloadImpl, 0, len(fns)-j)
						} else {
							__antithesis_instrumentation__.Notify(615110)
						}
						__antithesis_instrumentation__.Notify(615108)
						filtered = append(filtered, fns[j])
					} else {
						__antithesis_instrumentation__.Notify(615111)
					}
				}
				__antithesis_instrumentation__.Notify(615105)

				if filtered != nil {
					__antithesis_instrumentation__.Notify(615112)
					fns = filtered

					typedSubExprs[i] = NewTypedCastExpr(typedSubExprs[i], types.String)
				} else {
					__antithesis_instrumentation__.Notify(615113)
				}
			} else {
				__antithesis_instrumentation__.Notify(615114)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(615115)
	}
	__antithesis_instrumentation__.Notify(615077)

	if !def.NullableArgs && func() bool {
		__antithesis_instrumentation__.Notify(615116)
		return def.FunctionProperties.Class != GeneratorClass == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(615117)
		return def.FunctionProperties.Class != AggregateClass == true
	}() == true {
		__antithesis_instrumentation__.Notify(615118)
		for _, expr := range typedSubExprs {
			__antithesis_instrumentation__.Notify(615119)
			if expr.ResolvedType().Family() == types.UnknownFamily {
				__antithesis_instrumentation__.Notify(615120)
				return DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(615121)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(615122)
	}
	__antithesis_instrumentation__.Notify(615078)

	if len(fns) != 1 {
		__antithesis_instrumentation__.Notify(615123)
		typeNames := make([]string, 0, len(expr.Exprs))
		for _, expr := range typedSubExprs {
			__antithesis_instrumentation__.Notify(615127)
			typeNames = append(typeNames, expr.ResolvedType().String())
		}
		__antithesis_instrumentation__.Notify(615124)
		var desStr string
		if desired.Family() != types.AnyFamily {
			__antithesis_instrumentation__.Notify(615128)
			desStr = fmt.Sprintf(" (desired <%s>)", desired)
		} else {
			__antithesis_instrumentation__.Notify(615129)
		}
		__antithesis_instrumentation__.Notify(615125)
		sig := fmt.Sprintf("%s(%s)%s", &expr.Func, strings.Join(typeNames, ", "), desStr)
		if len(fns) == 0 {
			__antithesis_instrumentation__.Notify(615130)
			return nil, pgerror.Newf(pgcode.UndefinedFunction, "unknown signature: %s", sig)
		} else {
			__antithesis_instrumentation__.Notify(615131)
		}
		__antithesis_instrumentation__.Notify(615126)
		fnsStr := formatCandidates(expr.Func.String(), fns)
		return nil, pgerror.Newf(pgcode.AmbiguousFunction, "ambiguous call: %s, candidates are:\n%s", sig, fnsStr)
	} else {
		__antithesis_instrumentation__.Notify(615132)
	}
	__antithesis_instrumentation__.Notify(615079)
	overloadImpl := fns[0].(*Overload)

	if expr.IsWindowFunctionApplication() {
		__antithesis_instrumentation__.Notify(615133)

		if err := CheckIsWindowOrAgg(def); err != nil {
			__antithesis_instrumentation__.Notify(615135)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615136)
		}
		__antithesis_instrumentation__.Notify(615134)
		if expr.Type == DistinctFuncType {
			__antithesis_instrumentation__.Notify(615137)
			return nil, pgerror.New(pgcode.FeatureNotSupported, "DISTINCT is not implemented for window functions")
		} else {
			__antithesis_instrumentation__.Notify(615138)
		}
	} else {
		__antithesis_instrumentation__.Notify(615139)

		if def.Class == WindowClass {
			__antithesis_instrumentation__.Notify(615140)
			return nil, pgerror.Newf(pgcode.WrongObjectType,
				"window function %s() requires an OVER clause", &expr.Func)
		} else {
			__antithesis_instrumentation__.Notify(615141)
		}
	}
	__antithesis_instrumentation__.Notify(615080)

	if expr.Filter != nil {
		__antithesis_instrumentation__.Notify(615142)
		if def.Class != AggregateClass {
			__antithesis_instrumentation__.Notify(615145)

			return nil, pgerror.Newf(pgcode.WrongObjectType,
				"FILTER specified but %s() is not an aggregate function", &expr.Func)
		} else {
			__antithesis_instrumentation__.Notify(615146)
		}
		__antithesis_instrumentation__.Notify(615143)

		typedFilter, err := typeCheckAndRequireBoolean(ctx, semaCtx, expr.Filter, "FILTER expression")
		if err != nil {
			__antithesis_instrumentation__.Notify(615147)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615148)
		}
		__antithesis_instrumentation__.Notify(615144)
		expr.Filter = typedFilter
	} else {
		__antithesis_instrumentation__.Notify(615149)
	}
	__antithesis_instrumentation__.Notify(615081)

	if expr.OrderBy != nil {
		__antithesis_instrumentation__.Notify(615150)
		for i := range expr.OrderBy {
			__antithesis_instrumentation__.Notify(615151)
			typedExpr, err := expr.OrderBy[i].Expr.TypeCheck(ctx, semaCtx, types.Any)
			if err != nil {
				__antithesis_instrumentation__.Notify(615153)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(615154)
			}
			__antithesis_instrumentation__.Notify(615152)
			expr.OrderBy[i].Expr = typedExpr
		}
	} else {
		__antithesis_instrumentation__.Notify(615155)
	}
	__antithesis_instrumentation__.Notify(615082)

	for i, subExpr := range typedSubExprs {
		__antithesis_instrumentation__.Notify(615156)
		expr.Exprs[i] = subExpr
	}
	__antithesis_instrumentation__.Notify(615083)
	expr.fn = overloadImpl
	expr.fnProps = &def.FunctionProperties
	expr.typ = overloadImpl.returnType()(typedSubExprs)
	if expr.typ == UnknownReturnType {
		__antithesis_instrumentation__.Notify(615157)
		typeNames := make([]string, 0, len(expr.Exprs))
		for _, expr := range typedSubExprs {
			__antithesis_instrumentation__.Notify(615159)
			typeNames = append(typeNames, expr.ResolvedType().String())
		}
		__antithesis_instrumentation__.Notify(615158)
		return nil, pgerror.Newf(
			pgcode.DatatypeMismatch,
			"could not determine polymorphic type: %s(%s)",
			&expr.Func,
			strings.Join(typeNames, ", "),
		)
	} else {
		__antithesis_instrumentation__.Notify(615160)
	}
	__antithesis_instrumentation__.Notify(615084)
	if err := semaCtx.checkVolatility(overloadImpl.Volatility); err != nil {
		__antithesis_instrumentation__.Notify(615161)
		return nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "%s()", def.Name)
	} else {
		__antithesis_instrumentation__.Notify(615162)
	}
	__antithesis_instrumentation__.Notify(615085)
	if overloadImpl.counter != nil {
		__antithesis_instrumentation__.Notify(615163)
		telemetry.Inc(overloadImpl.counter)
	} else {
		__antithesis_instrumentation__.Notify(615164)
	}
	__antithesis_instrumentation__.Notify(615086)
	return expr, nil
}

func (expr *IfErrExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615165)
	var typedCond, typedElse TypedExpr
	var retType *types.T
	var err error
	if expr.Else == nil {
		__antithesis_instrumentation__.Notify(615168)
		typedCond, err = expr.Cond.TypeCheck(ctx, semaCtx, types.Any)
		if err != nil {
			__antithesis_instrumentation__.Notify(615170)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615171)
		}
		__antithesis_instrumentation__.Notify(615169)
		retType = types.Bool
	} else {
		__antithesis_instrumentation__.Notify(615172)
		var typedSubExprs []TypedExpr
		typedSubExprs, retType, err = TypeCheckSameTypedExprs(ctx, semaCtx, desired, expr.Cond, expr.Else)
		if err != nil {
			__antithesis_instrumentation__.Notify(615174)
			return nil, decorateTypeCheckError(err, "incompatible IFERROR expressions")
		} else {
			__antithesis_instrumentation__.Notify(615175)
		}
		__antithesis_instrumentation__.Notify(615173)
		typedCond, typedElse = typedSubExprs[0], typedSubExprs[1]
	}
	__antithesis_instrumentation__.Notify(615166)

	var typedErrCode TypedExpr
	if expr.ErrCode != nil {
		__antithesis_instrumentation__.Notify(615176)
		typedErrCode, err = typeCheckAndRequire(ctx, semaCtx, expr.ErrCode, types.String, "IFERROR")
		if err != nil {
			__antithesis_instrumentation__.Notify(615177)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615178)
		}
	} else {
		__antithesis_instrumentation__.Notify(615179)
	}
	__antithesis_instrumentation__.Notify(615167)

	expr.Cond = typedCond
	expr.Else = typedElse
	expr.ErrCode = typedErrCode
	expr.typ = retType

	telemetry.Inc(sqltelemetry.IfErrCounter)
	return expr, nil
}

func (expr *IfExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615180)
	typedCond, err := typeCheckAndRequireBoolean(ctx, semaCtx, expr.Cond, "IF condition")
	if err != nil {
		__antithesis_instrumentation__.Notify(615183)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615184)
	}
	__antithesis_instrumentation__.Notify(615181)

	typedSubExprs, retType, err := TypeCheckSameTypedExprs(ctx, semaCtx, desired, expr.True, expr.Else)
	if err != nil {
		__antithesis_instrumentation__.Notify(615185)
		return nil, decorateTypeCheckError(err, "incompatible IF expressions")
	} else {
		__antithesis_instrumentation__.Notify(615186)
	}
	__antithesis_instrumentation__.Notify(615182)

	expr.Cond = typedCond
	expr.True, expr.Else = typedSubExprs[0], typedSubExprs[1]
	expr.typ = retType
	return expr, nil
}

func (expr *IsOfTypeExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615187)
	exprTyped, err := expr.Expr.TypeCheck(ctx, semaCtx, types.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(615190)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615191)
	}
	__antithesis_instrumentation__.Notify(615188)
	expr.resolvedTypes = make([]*types.T, len(expr.Types))
	for i := range expr.Types {
		__antithesis_instrumentation__.Notify(615192)
		typ, err := ResolveType(ctx, expr.Types[i], semaCtx.GetTypeResolver())
		if err != nil {
			__antithesis_instrumentation__.Notify(615194)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615195)
		}
		__antithesis_instrumentation__.Notify(615193)
		expr.Types[i] = typ
		expr.resolvedTypes[i] = typ
	}
	__antithesis_instrumentation__.Notify(615189)
	expr.Expr = exprTyped
	expr.typ = types.Bool
	return expr, nil
}

func (expr *NotExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615196)
	exprTyped, err := typeCheckAndRequireBoolean(ctx, semaCtx, expr.Expr, "NOT argument")
	if err != nil {
		__antithesis_instrumentation__.Notify(615198)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615199)
	}
	__antithesis_instrumentation__.Notify(615197)
	expr.Expr = exprTyped
	expr.typ = types.Bool
	return expr, nil
}

func (expr *IsNullExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615200)
	exprTyped, err := expr.Expr.TypeCheck(ctx, semaCtx, types.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(615202)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615203)
	}
	__antithesis_instrumentation__.Notify(615201)
	expr.Expr = exprTyped
	expr.typ = types.Bool
	return expr, nil
}

func (expr *IsNotNullExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615204)
	exprTyped, err := expr.Expr.TypeCheck(ctx, semaCtx, types.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(615206)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615207)
	}
	__antithesis_instrumentation__.Notify(615205)
	expr.Expr = exprTyped
	expr.typ = types.Bool
	return expr, nil
}

func (expr *NullIfExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615208)
	typedSubExprs, retType, err := TypeCheckSameTypedExprs(ctx, semaCtx, desired, expr.Expr1, expr.Expr2)
	if err != nil {
		__antithesis_instrumentation__.Notify(615210)
		return nil, decorateTypeCheckError(err, "incompatible NULLIF expressions")
	} else {
		__antithesis_instrumentation__.Notify(615211)
	}
	__antithesis_instrumentation__.Notify(615209)

	expr.Expr1, expr.Expr2 = typedSubExprs[0], typedSubExprs[1]
	expr.typ = retType
	return expr, nil
}

func (expr *OrExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615212)
	leftTyped, err := typeCheckAndRequireBoolean(ctx, semaCtx, expr.Left, "OR argument")
	if err != nil {
		__antithesis_instrumentation__.Notify(615215)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615216)
	}
	__antithesis_instrumentation__.Notify(615213)
	rightTyped, err := typeCheckAndRequireBoolean(ctx, semaCtx, expr.Right, "OR argument")
	if err != nil {
		__antithesis_instrumentation__.Notify(615217)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615218)
	}
	__antithesis_instrumentation__.Notify(615214)
	expr.Left, expr.Right = leftTyped, rightTyped
	expr.typ = types.Bool
	return expr, nil
}

func (expr *ParenExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615219)
	exprTyped, err := expr.Expr.TypeCheck(ctx, semaCtx, desired)
	if err != nil {
		__antithesis_instrumentation__.Notify(615221)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615222)
	}
	__antithesis_instrumentation__.Notify(615220)

	return exprTyped, nil
}

func (expr *ColumnItem) TypeCheck(
	_ context.Context, _ *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615223)
	name := expr.String()
	if _, ok := presetTypesForTesting[name]; ok {
		__antithesis_instrumentation__.Notify(615225)
		return expr, nil
	} else {
		__antithesis_instrumentation__.Notify(615226)
	}
	__antithesis_instrumentation__.Notify(615224)
	return nil, pgerror.Newf(pgcode.UndefinedColumn,
		"column %q does not exist", ErrString(expr))
}

func (expr UnqualifiedStar) TypeCheck(
	_ context.Context, _ *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615227)
	return nil, errStarNotAllowed
}

func (expr *UnresolvedName) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615228)
	v, err := expr.NormalizeVarName()
	if err != nil {
		__antithesis_instrumentation__.Notify(615230)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615231)
	}
	__antithesis_instrumentation__.Notify(615229)
	return v.TypeCheck(ctx, semaCtx, desired)
}

func (expr *AllColumnsSelector) TypeCheck(
	_ context.Context, _ *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615232)
	return nil, pgerror.Newf(pgcode.Syntax, "cannot use %q in this context", expr)
}

func (expr *RangeCond) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615233)
	leftFromTyped, fromTyped, _, _, err := typeCheckComparisonOp(ctx, semaCtx, treecmp.MakeComparisonOperator(treecmp.GT), expr.Left, expr.From)
	if err != nil {
		__antithesis_instrumentation__.Notify(615237)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615238)
	}
	__antithesis_instrumentation__.Notify(615234)
	leftToTyped, toTyped, _, _, err := typeCheckComparisonOp(ctx, semaCtx, treecmp.MakeComparisonOperator(treecmp.LT), expr.Left, expr.To)
	if err != nil {
		__antithesis_instrumentation__.Notify(615239)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615240)
	}
	__antithesis_instrumentation__.Notify(615235)

	_, _, _, _, err = typeCheckComparisonOp(ctx, semaCtx, treecmp.MakeComparisonOperator(treecmp.LT), expr.From, expr.To)
	if err != nil {
		__antithesis_instrumentation__.Notify(615241)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615242)
	}
	__antithesis_instrumentation__.Notify(615236)
	expr.Left, expr.From = leftFromTyped, fromTyped
	expr.leftTo, expr.To = leftToTyped, toTyped
	expr.typ = types.Bool
	return expr, nil
}

func (expr *Subquery) TypeCheck(_ context.Context, sc *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615243)
	if sc != nil && func() bool {
		__antithesis_instrumentation__.Notify(615245)
		return sc.Properties.required.rejectFlags&RejectSubqueries != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(615246)
		return nil, pgerror.Newf(pgcode.FeatureNotSupported,
			"subqueries are not allowed in %s", sc.Properties.required.context)
	} else {
		__antithesis_instrumentation__.Notify(615247)
	}
	__antithesis_instrumentation__.Notify(615244)
	expr.assertTyped()
	return expr, nil
}

func (expr *UnaryExpr) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615248)
	ops := UnaryOps[expr.Operator.Symbol]

	typedSubExprs, fns, err := typeCheckOverloadedExprs(ctx, semaCtx, desired, ops, false, expr.Expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(615254)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615255)
	}
	__antithesis_instrumentation__.Notify(615249)

	exprTyped := typedSubExprs[0]
	exprReturn := exprTyped.ResolvedType()

	if len(fns) > 0 {
		__antithesis_instrumentation__.Notify(615256)
		if exprReturn.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(615257)
			return DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(615258)
		}
	} else {
		__antithesis_instrumentation__.Notify(615259)
	}
	__antithesis_instrumentation__.Notify(615250)

	if len(fns) != 1 {
		__antithesis_instrumentation__.Notify(615260)
		var desStr string
		if desired.Family() != types.AnyFamily {
			__antithesis_instrumentation__.Notify(615263)
			desStr = fmt.Sprintf(" (desired <%s>)", desired)
		} else {
			__antithesis_instrumentation__.Notify(615264)
		}
		__antithesis_instrumentation__.Notify(615261)
		sig := fmt.Sprintf("%s <%s>%s", expr.Operator, exprReturn, desStr)
		if len(fns) == 0 {
			__antithesis_instrumentation__.Notify(615265)
			return nil,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedUnaryOpErrFmt, sig)
		} else {
			__antithesis_instrumentation__.Notify(615266)
		}
		__antithesis_instrumentation__.Notify(615262)
		fnsStr := formatCandidates(expr.Operator.String(), fns)
		err = pgerror.Newf(pgcode.AmbiguousFunction, ambiguousUnaryOpErrFmt, sig)
		err = errors.WithHintf(err, candidatesHintFmt, fnsStr)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615267)
	}
	__antithesis_instrumentation__.Notify(615251)

	unaryOp := fns[0].(*UnaryOp)
	if err := semaCtx.checkVolatility(unaryOp.Volatility); err != nil {
		__antithesis_instrumentation__.Notify(615268)
		return nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "%s", expr.Operator)
	} else {
		__antithesis_instrumentation__.Notify(615269)
	}
	__antithesis_instrumentation__.Notify(615252)

	if unaryOp.counter != nil {
		__antithesis_instrumentation__.Notify(615270)
		telemetry.Inc(unaryOp.counter)
	} else {
		__antithesis_instrumentation__.Notify(615271)
	}
	__antithesis_instrumentation__.Notify(615253)

	expr.Expr = exprTyped
	expr.fn = unaryOp
	expr.typ = unaryOp.returnType()(typedSubExprs)
	return expr, nil
}

func (expr DefaultVal) TypeCheck(
	_ context.Context, _ *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615272)
	return nil, errInvalidDefaultUsage
}

func (expr PartitionMinVal) TypeCheck(
	_ context.Context, _ *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615273)
	return nil, errInvalidMinUsage
}

func (expr PartitionMaxVal) TypeCheck(
	_ context.Context, _ *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615274)
	return nil, errInvalidMaxUsage
}

func (expr *NumVal) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615275)
	return typeCheckConstant(ctx, semaCtx, expr, desired)
}

func (expr *StrVal) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615276)
	return typeCheckConstant(ctx, semaCtx, expr, desired)
}

func (expr *Tuple) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615277)

	if len(expr.Labels) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(615281)
		return len(expr.Labels) != len(expr.Exprs) == true
	}() == true {
		__antithesis_instrumentation__.Notify(615282)
		return nil, pgerror.Newf(pgcode.Syntax,
			"mismatch in tuple definition: %d expressions, %d labels",
			len(expr.Exprs), len(expr.Labels),
		)
	} else {
		__antithesis_instrumentation__.Notify(615283)
	}
	__antithesis_instrumentation__.Notify(615278)

	var labels []string
	contents := make([]*types.T, len(expr.Exprs))
	for i, subExpr := range expr.Exprs {
		__antithesis_instrumentation__.Notify(615284)
		desiredElem := types.Any
		if desired.Family() == types.TupleFamily && func() bool {
			__antithesis_instrumentation__.Notify(615287)
			return len(desired.TupleContents()) > i == true
		}() == true {
			__antithesis_instrumentation__.Notify(615288)
			desiredElem = desired.TupleContents()[i]
		} else {
			__antithesis_instrumentation__.Notify(615289)
		}
		__antithesis_instrumentation__.Notify(615285)
		typedExpr, err := subExpr.TypeCheck(ctx, semaCtx, desiredElem)
		if err != nil {
			__antithesis_instrumentation__.Notify(615290)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615291)
		}
		__antithesis_instrumentation__.Notify(615286)
		expr.Exprs[i] = typedExpr
		contents[i] = typedExpr.ResolvedType()
	}
	__antithesis_instrumentation__.Notify(615279)

	if len(expr.Labels) > 0 {
		__antithesis_instrumentation__.Notify(615292)
		labels = make([]string, len(expr.Labels))
		copy(labels, expr.Labels)
	} else {
		__antithesis_instrumentation__.Notify(615293)
	}
	__antithesis_instrumentation__.Notify(615280)
	expr.typ = types.MakeLabeledTuple(contents, labels)
	return expr, nil
}

var errAmbiguousArrayType = pgerror.Newf(pgcode.IndeterminateDatatype, "cannot determine type of empty array. "+
	"Consider annotating with the desired type, for example ARRAY[]:::int[]")

func (expr *Array) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615294)
	desiredParam := types.Any
	if desired.Family() == types.ArrayFamily {
		__antithesis_instrumentation__.Notify(615299)
		desiredParam = desired.ArrayContents()
	} else {
		__antithesis_instrumentation__.Notify(615300)
	}
	__antithesis_instrumentation__.Notify(615295)

	if len(expr.Exprs) == 0 {
		__antithesis_instrumentation__.Notify(615301)
		if desiredParam.Family() == types.AnyFamily {
			__antithesis_instrumentation__.Notify(615303)
			return nil, errAmbiguousArrayType
		} else {
			__antithesis_instrumentation__.Notify(615304)
		}
		__antithesis_instrumentation__.Notify(615302)
		expr.typ = types.MakeArray(desiredParam)
		return expr, nil
	} else {
		__antithesis_instrumentation__.Notify(615305)
	}
	__antithesis_instrumentation__.Notify(615296)

	typedSubExprs, typ, err := TypeCheckSameTypedExprs(ctx, semaCtx, desiredParam, expr.Exprs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(615306)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615307)
	}
	__antithesis_instrumentation__.Notify(615297)

	expr.typ = types.MakeArray(typ)
	for i := range typedSubExprs {
		__antithesis_instrumentation__.Notify(615308)
		expr.Exprs[i] = typedSubExprs[i]
	}
	__antithesis_instrumentation__.Notify(615298)

	telemetry.Inc(sqltelemetry.ArrayConstructorCounter)
	return expr, nil
}

func (expr *ArrayFlatten) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615309)
	desiredParam := types.Any
	if desired.Family() == types.ArrayFamily {
		__antithesis_instrumentation__.Notify(615312)
		desiredParam = desired.ArrayContents()
	} else {
		__antithesis_instrumentation__.Notify(615313)
	}
	__antithesis_instrumentation__.Notify(615310)

	subqueryTyped, err := expr.Subquery.TypeCheck(ctx, semaCtx, desiredParam)
	if err != nil {
		__antithesis_instrumentation__.Notify(615314)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615315)
	}
	__antithesis_instrumentation__.Notify(615311)
	expr.Subquery = subqueryTyped
	expr.typ = types.MakeArray(subqueryTyped.ResolvedType())

	telemetry.Inc(sqltelemetry.ArrayFlattenCounter)
	return expr, nil
}

func (expr *Placeholder) TypeCheck(
	ctx context.Context, semaCtx *SemaContext, desired *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615316)

	desired = desired.WithoutTypeModifiers()
	if typ, ok, err := semaCtx.Placeholders.Type(expr.Idx); err != nil {
		__antithesis_instrumentation__.Notify(615320)
		return expr, err
	} else {
		__antithesis_instrumentation__.Notify(615321)
		if ok {
			__antithesis_instrumentation__.Notify(615322)
			typ = typ.WithoutTypeModifiers()
			if !desired.Equivalent(typ) {
				__antithesis_instrumentation__.Notify(615325)

				typ = desired
			} else {
				__antithesis_instrumentation__.Notify(615326)
			}
			__antithesis_instrumentation__.Notify(615323)

			if err := semaCtx.Placeholders.SetType(expr.Idx, typ); err != nil {
				__antithesis_instrumentation__.Notify(615327)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(615328)
			}
			__antithesis_instrumentation__.Notify(615324)
			expr.typ = typ
			return expr, nil
		} else {
			__antithesis_instrumentation__.Notify(615329)
		}
	}
	__antithesis_instrumentation__.Notify(615317)
	if desired.IsAmbiguous() {
		__antithesis_instrumentation__.Notify(615330)
		return nil, placeholderTypeAmbiguityError(expr.Idx)
	} else {
		__antithesis_instrumentation__.Notify(615331)
	}
	__antithesis_instrumentation__.Notify(615318)
	if err := semaCtx.Placeholders.SetType(expr.Idx, desired); err != nil {
		__antithesis_instrumentation__.Notify(615332)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615333)
	}
	__antithesis_instrumentation__.Notify(615319)
	expr.typ = desired
	return expr, nil
}

func (d *DBitArray) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615334)
	return d, nil
}

func (d *DBool) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615335)
	return d, nil
}

func (d *DInt) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615336)
	return d, nil
}

func (d *DFloat) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615337)
	return d, nil
}

func (d *DEnum) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615338)
	return d, nil
}

func (d *DDecimal) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615339)
	return d, nil
}

func (d *DString) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615340)
	return d, nil
}

func (d *DCollatedString) TypeCheck(
	_ context.Context, _ *SemaContext, _ *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615341)
	return d, nil
}

func (d *DBytes) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615342)
	return d, nil
}

func (d *DUuid) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615343)
	return d, nil
}

func (d *DIPAddr) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615344)
	return d, nil
}

func (d *DDate) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615345)
	return d, nil
}

func (d *DTime) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615346)
	return d, nil
}

func (d *DTimeTZ) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615347)
	return d, nil
}

func (d *DTimestamp) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615348)
	return d, nil
}

func (d *DTimestampTZ) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615349)
	return d, nil
}

func (d *DInterval) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615350)
	return d, nil
}

func (d *DBox2D) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615351)
	return d, nil
}

func (d *DGeography) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615352)
	return d, nil
}

func (d *DGeometry) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615353)
	return d, nil
}

func (d *DJSON) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615354)
	return d, nil
}

func (d *DTuple) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615355)
	return d, nil
}

func (d *DVoid) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615356)
	return d, nil
}

func (d *DArray) TypeCheck(_ context.Context, _ *SemaContext, desired *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615357)

	if (d.ParamTyp == types.Unknown || func() bool {
		__antithesis_instrumentation__.Notify(615359)
		return d.ParamTyp == types.Any == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(615360)
		return (!d.HasNonNulls) == true
	}() == true {
		__antithesis_instrumentation__.Notify(615361)
		if desired.Family() != types.ArrayFamily {
			__antithesis_instrumentation__.Notify(615363)

			return d, nil
		} else {
			__antithesis_instrumentation__.Notify(615364)
		}
		__antithesis_instrumentation__.Notify(615362)
		dCopy := &DArray{}
		*dCopy = *d
		dCopy.ParamTyp = desired.ArrayContents()
		return dCopy, nil
	} else {
		__antithesis_instrumentation__.Notify(615365)
	}
	__antithesis_instrumentation__.Notify(615358)
	return d, nil
}

func (d *DOid) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615366)
	return d, nil
}

func (d *DOidWrapper) TypeCheck(_ context.Context, _ *SemaContext, _ *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615367)
	return d, nil
}

func (d dNull) TypeCheck(_ context.Context, _ *SemaContext, desired *types.T) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615368)
	return d, nil
}

func typeCheckAndRequireTupleElems(
	ctx context.Context,
	semaCtx *SemaContext,
	expr TypedExpr,
	tuple *Tuple,
	op treecmp.ComparisonOperator,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615369)
	tuple.typ = types.MakeTuple(make([]*types.T, len(tuple.Exprs)))
	for i, subExpr := range tuple.Exprs {
		__antithesis_instrumentation__.Notify(615372)

		_, rightTyped, _, _, err := typeCheckComparisonOp(ctx, semaCtx, op, expr, subExpr)
		if err != nil {
			__antithesis_instrumentation__.Notify(615374)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615375)
		}
		__antithesis_instrumentation__.Notify(615373)
		tuple.Exprs[i] = rightTyped
		tuple.typ.TupleContents()[i] = rightTyped.ResolvedType()
	}
	__antithesis_instrumentation__.Notify(615370)
	if len(tuple.typ.TupleContents()) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(615376)
		return tuple.typ.TupleContents()[0].Family() == types.CollatedStringFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(615377)

		var typWithLocale *types.T
		for _, typ := range tuple.typ.TupleContents() {
			__antithesis_instrumentation__.Notify(615379)
			if typ.Locale() != "" {
				__antithesis_instrumentation__.Notify(615380)
				typWithLocale = typ
				break
			} else {
				__antithesis_instrumentation__.Notify(615381)
			}
		}
		__antithesis_instrumentation__.Notify(615378)
		if typWithLocale != nil {
			__antithesis_instrumentation__.Notify(615382)
			for i := range tuple.typ.TupleContents() {
				__antithesis_instrumentation__.Notify(615383)
				tuple.typ.TupleContents()[i] = typWithLocale
			}
		} else {
			__antithesis_instrumentation__.Notify(615384)
		}
	} else {
		__antithesis_instrumentation__.Notify(615385)
	}
	__antithesis_instrumentation__.Notify(615371)
	return tuple, nil
}

func typeCheckAndRequireBoolean(
	ctx context.Context, semaCtx *SemaContext, expr Expr, op string,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615386)
	return typeCheckAndRequire(ctx, semaCtx, expr, types.Bool, op)
}

func typeCheckAndRequire(
	ctx context.Context, semaCtx *SemaContext, expr Expr, required *types.T, op string,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615387)
	typedExpr, err := expr.TypeCheck(ctx, semaCtx, required)
	if err != nil {
		__antithesis_instrumentation__.Notify(615390)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(615391)
	}
	__antithesis_instrumentation__.Notify(615388)
	if typ := typedExpr.ResolvedType(); !(typ.Family() == types.UnknownFamily || func() bool {
		__antithesis_instrumentation__.Notify(615392)
		return typ.Equivalent(required) == true
	}() == true) {
		__antithesis_instrumentation__.Notify(615393)
		return nil, pgerror.Newf(pgcode.DatatypeMismatch, "incompatible %s type: %s", op, typ)
	} else {
		__antithesis_instrumentation__.Notify(615394)
	}
	__antithesis_instrumentation__.Notify(615389)
	return typedExpr, nil
}

const (
	compSignatureFmt          = "<%s> %s <%s>"
	compSignatureWithSubOpFmt = "<%s> %s %s <%s>"
	compExprsFmt              = "%s %s %s: %v"
	compExprsWithSubOpFmt     = "%s %s %s %s: %v"
	invalidCompErrFmt         = "invalid comparison between different %s types: %s"
	unsupportedCompErrFmt     = "unsupported comparison operator: %s"
	unsupportedUnaryOpErrFmt  = "unsupported unary operator: %s"
	unsupportedBinaryOpErrFmt = "unsupported binary operator: %s"
	ambiguousCompErrFmt       = "ambiguous comparison operator: %s"
	ambiguousUnaryOpErrFmt    = "ambiguous unary operator: %s"
	ambiguousBinaryOpErrFmt   = "ambiguous binary operator: %s"
	candidatesHintFmt         = "candidates are:\n%s"
)

func typeCheckComparisonOpWithSubOperator(
	ctx context.Context, semaCtx *SemaContext, op, subOp treecmp.ComparisonOperator, left, right Expr,
) (_ TypedExpr, _ TypedExpr, _ *CmpOp, alwaysNull bool, _ error) {
	__antithesis_instrumentation__.Notify(615395)

	left = StripParens(left)
	right = StripParens(right)

	foldedOp, _, _, _, _ := FoldComparisonExpr(subOp, nil, nil)
	ops := CmpOps[foldedOp.Symbol]

	var cmpTypeLeft, cmpTypeRight *types.T
	var leftTyped, rightTyped TypedExpr
	if array, isConstructor := right.(*Array); isConstructor {
		__antithesis_instrumentation__.Notify(615398)

		sameTypeExprs := make([]Expr, len(array.Exprs)+1)
		sameTypeExprs[0] = left
		copy(sameTypeExprs[1:], array.Exprs)

		typedSubExprs, retType, err := TypeCheckSameTypedExprs(ctx, semaCtx, types.Any, sameTypeExprs...)
		if err != nil {
			__antithesis_instrumentation__.Notify(615401)
			sigWithErr := fmt.Sprintf(compExprsWithSubOpFmt, left, subOp, op, right, err)
			return nil, nil, nil, false,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sigWithErr)
		} else {
			__antithesis_instrumentation__.Notify(615402)
		}
		__antithesis_instrumentation__.Notify(615399)

		leftTyped = typedSubExprs[0]
		cmpTypeLeft = retType

		for i, typedExpr := range typedSubExprs[1:] {
			__antithesis_instrumentation__.Notify(615403)
			array.Exprs[i] = typedExpr
		}
		__antithesis_instrumentation__.Notify(615400)
		array.typ = types.MakeArray(retType)

		rightTyped = array
		cmpTypeRight = retType

		if leftTyped.ResolvedType().Family() == types.UnknownFamily || func() bool {
			__antithesis_instrumentation__.Notify(615404)
			return retType.Family() == types.UnknownFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(615405)
			return leftTyped, rightTyped, nil, true, nil
		} else {
			__antithesis_instrumentation__.Notify(615406)
		}
	} else {
		__antithesis_instrumentation__.Notify(615407)

		var err error
		leftTyped, err = left.TypeCheck(ctx, semaCtx, types.Any)
		if err != nil {
			__antithesis_instrumentation__.Notify(615411)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(615412)
		}
		__antithesis_instrumentation__.Notify(615408)
		cmpTypeLeft = leftTyped.ResolvedType()

		if tuple, ok := right.(*Tuple); ok {
			__antithesis_instrumentation__.Notify(615413)

			rightTyped, err = typeCheckAndRequireTupleElems(ctx, semaCtx, leftTyped, tuple, subOp)
			if err != nil {
				__antithesis_instrumentation__.Notify(615414)
				return nil, nil, nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(615415)
			}
		} else {
			__antithesis_instrumentation__.Notify(615416)

			rightTyped, err = right.TypeCheck(ctx, semaCtx, types.MakeArray(cmpTypeLeft))
			if err != nil {
				__antithesis_instrumentation__.Notify(615417)
				return nil, nil, nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(615418)
			}
		}
		__antithesis_instrumentation__.Notify(615409)

		rightReturn := rightTyped.ResolvedType()
		if cmpTypeLeft.Family() == types.UnknownFamily || func() bool {
			__antithesis_instrumentation__.Notify(615419)
			return rightReturn.Family() == types.UnknownFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(615420)
			return leftTyped, rightTyped, nil, true, nil
		} else {
			__antithesis_instrumentation__.Notify(615421)
		}
		__antithesis_instrumentation__.Notify(615410)

		switch rightReturn.Family() {
		case types.ArrayFamily:
			__antithesis_instrumentation__.Notify(615422)
			cmpTypeRight = rightReturn.ArrayContents()
		case types.TupleFamily:
			__antithesis_instrumentation__.Notify(615423)
			if len(rightReturn.TupleContents()) == 0 {
				__antithesis_instrumentation__.Notify(615425)

				cmpTypeRight = cmpTypeLeft
			} else {
				__antithesis_instrumentation__.Notify(615426)

				cmpTypeRight = rightReturn.TupleContents()[0]
			}
		default:
			__antithesis_instrumentation__.Notify(615424)
			sigWithErr := fmt.Sprintf(compExprsWithSubOpFmt, left, subOp, op, right,
				fmt.Sprintf("op %s <right> requires array, tuple or subquery on right side", op))
			return nil, nil, nil, false, pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sigWithErr)
		}
	}
	__antithesis_instrumentation__.Notify(615396)
	fn, ok := ops.LookupImpl(cmpTypeLeft, cmpTypeRight)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(615427)
		return !deepCheckValidCmpOp(ops, cmpTypeLeft, cmpTypeRight) == true
	}() == true {
		__antithesis_instrumentation__.Notify(615428)
		return nil, nil, nil, false, subOpCompError(cmpTypeLeft, rightTyped.ResolvedType(), subOp, op)
	} else {
		__antithesis_instrumentation__.Notify(615429)
	}
	__antithesis_instrumentation__.Notify(615397)
	return leftTyped, rightTyped, fn, false, nil
}

func deepCheckValidCmpOp(ops cmpOpOverload, leftType, rightType *types.T) bool {
	__antithesis_instrumentation__.Notify(615430)
	if leftType.Family() == types.TupleFamily && func() bool {
		__antithesis_instrumentation__.Notify(615432)
		return rightType.Family() == types.TupleFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(615433)
		l := leftType.TupleContents()
		r := rightType.TupleContents()
		if len(l) != len(r) {
			__antithesis_instrumentation__.Notify(615435)
			return false
		} else {
			__antithesis_instrumentation__.Notify(615436)
		}
		__antithesis_instrumentation__.Notify(615434)
		for i := range l {
			__antithesis_instrumentation__.Notify(615437)
			if _, ok := ops.LookupImpl(l[i], r[i]); !ok {
				__antithesis_instrumentation__.Notify(615439)
				return false
			} else {
				__antithesis_instrumentation__.Notify(615440)
			}
			__antithesis_instrumentation__.Notify(615438)
			if !deepCheckValidCmpOp(ops, l[i], r[i]) {
				__antithesis_instrumentation__.Notify(615441)
				return false
			} else {
				__antithesis_instrumentation__.Notify(615442)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(615443)
	}
	__antithesis_instrumentation__.Notify(615431)
	return true
}

func subOpCompError(leftType, rightType *types.T, subOp, op treecmp.ComparisonOperator) error {
	__antithesis_instrumentation__.Notify(615444)
	sig := fmt.Sprintf(compSignatureWithSubOpFmt, leftType, subOp, op, rightType)
	return pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sig)
}

func typeCheckSubqueryWithIn(left, right *types.T) error {
	__antithesis_instrumentation__.Notify(615445)
	if right.Family() == types.TupleFamily {
		__antithesis_instrumentation__.Notify(615447)

		if len(right.TupleContents()) != 1 {
			__antithesis_instrumentation__.Notify(615449)
			return pgerror.Newf(pgcode.InvalidParameterValue,
				unsupportedCompErrFmt, fmt.Sprintf(compSignatureFmt, left, treecmp.In, right))
		} else {
			__antithesis_instrumentation__.Notify(615450)
		}
		__antithesis_instrumentation__.Notify(615448)
		if !left.EquivalentOrNull(right.TupleContents()[0], false) {
			__antithesis_instrumentation__.Notify(615451)
			return pgerror.Newf(pgcode.InvalidParameterValue,
				unsupportedCompErrFmt, fmt.Sprintf(compSignatureFmt, left, treecmp.In, right))
		} else {
			__antithesis_instrumentation__.Notify(615452)
		}
	} else {
		__antithesis_instrumentation__.Notify(615453)
	}
	__antithesis_instrumentation__.Notify(615446)
	return nil
}

func typeCheckComparisonOp(
	ctx context.Context, semaCtx *SemaContext, op treecmp.ComparisonOperator, left, right Expr,
) (_ TypedExpr, _ TypedExpr, _ *CmpOp, alwaysNull bool, _ error) {
	__antithesis_instrumentation__.Notify(615454)
	foldedOp, foldedLeft, foldedRight, switched, _ := FoldComparisonExpr(op, left, right)
	ops := CmpOps[foldedOp.Symbol]

	_, leftIsTuple := foldedLeft.(*Tuple)
	rightTuple, rightIsTuple := foldedRight.(*Tuple)

	_, rightIsSubquery := foldedRight.(SubqueryExpr)
	switch {
	case foldedOp.Symbol == treecmp.In && func() bool {
		__antithesis_instrumentation__.Notify(615475)
		return rightIsTuple == true
	}() == true:
		__antithesis_instrumentation__.Notify(615461)
		sameTypeExprs := make([]Expr, len(rightTuple.Exprs)+1)
		sameTypeExprs[0] = foldedLeft
		copy(sameTypeExprs[1:], rightTuple.Exprs)

		typedSubExprs, retType, err := TypeCheckSameTypedExprs(ctx, semaCtx, types.Any, sameTypeExprs...)
		if err != nil {
			__antithesis_instrumentation__.Notify(615476)
			sigWithErr := fmt.Sprintf(compExprsFmt, left, op, right, err)
			return nil, nil, nil, false,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sigWithErr)
		} else {
			__antithesis_instrumentation__.Notify(615477)
		}
		__antithesis_instrumentation__.Notify(615462)

		fn, ok := ops.LookupImpl(retType, types.AnyTuple)
		if !ok {
			__antithesis_instrumentation__.Notify(615478)
			sig := fmt.Sprintf(compSignatureFmt, retType, op, types.AnyTuple)
			return nil, nil, nil, false,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sig)
		} else {
			__antithesis_instrumentation__.Notify(615479)
		}
		__antithesis_instrumentation__.Notify(615463)

		typedLeft := typedSubExprs[0]
		typedSubExprs = typedSubExprs[1:]

		rightTuple.typ = types.MakeTuple(make([]*types.T, len(typedSubExprs)))
		for i, typedExpr := range typedSubExprs {
			__antithesis_instrumentation__.Notify(615480)
			rightTuple.Exprs[i] = typedExpr
			rightTuple.typ.TupleContents()[i] = retType
		}
		__antithesis_instrumentation__.Notify(615464)
		if switched {
			__antithesis_instrumentation__.Notify(615481)
			return rightTuple, typedLeft, fn, false, nil
		} else {
			__antithesis_instrumentation__.Notify(615482)
		}
		__antithesis_instrumentation__.Notify(615465)
		return typedLeft, rightTuple, fn, false, nil

	case foldedOp.Symbol == treecmp.In && func() bool {
		__antithesis_instrumentation__.Notify(615483)
		return rightIsSubquery == true
	}() == true:
		__antithesis_instrumentation__.Notify(615466)
		typedLeft, err := foldedLeft.TypeCheck(ctx, semaCtx, types.Any)
		if err != nil {
			__antithesis_instrumentation__.Notify(615484)
			sigWithErr := fmt.Sprintf(compExprsFmt, left, op, right, err)
			return nil, nil, nil, false,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sigWithErr)
		} else {
			__antithesis_instrumentation__.Notify(615485)
		}
		__antithesis_instrumentation__.Notify(615467)

		typ := typedLeft.ResolvedType()
		fn, ok := ops.LookupImpl(typ, types.AnyTuple)
		if !ok {
			__antithesis_instrumentation__.Notify(615486)
			sig := fmt.Sprintf(compSignatureFmt, typ, op, types.AnyTuple)
			return nil, nil, nil, false,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sig)
		} else {
			__antithesis_instrumentation__.Notify(615487)
		}
		__antithesis_instrumentation__.Notify(615468)

		desired := types.MakeTuple([]*types.T{typ})
		typedRight, err := foldedRight.TypeCheck(ctx, semaCtx, desired)
		if err != nil {
			__antithesis_instrumentation__.Notify(615488)
			sigWithErr := fmt.Sprintf(compExprsFmt, left, op, right, err)
			return nil, nil, nil, false,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sigWithErr)
		} else {
			__antithesis_instrumentation__.Notify(615489)
		}
		__antithesis_instrumentation__.Notify(615469)

		if err := typeCheckSubqueryWithIn(
			typedLeft.ResolvedType(), typedRight.ResolvedType(),
		); err != nil {
			__antithesis_instrumentation__.Notify(615490)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(615491)
		}
		__antithesis_instrumentation__.Notify(615470)
		return typedLeft, typedRight, fn, false, nil

	case leftIsTuple && func() bool {
		__antithesis_instrumentation__.Notify(615492)
		return rightIsTuple == true
	}() == true:
		__antithesis_instrumentation__.Notify(615471)
		fn, ok := ops.LookupImpl(types.AnyTuple, types.AnyTuple)
		if !ok {
			__antithesis_instrumentation__.Notify(615493)
			sig := fmt.Sprintf(compSignatureFmt, types.AnyTuple, op, types.AnyTuple)
			return nil, nil, nil, false,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sig)
		} else {
			__antithesis_instrumentation__.Notify(615494)
		}
		__antithesis_instrumentation__.Notify(615472)

		typedLeft, typedRight, err := typeCheckTupleComparison(ctx, semaCtx, op, left.(*Tuple), right.(*Tuple))
		if err != nil {
			__antithesis_instrumentation__.Notify(615495)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(615496)
		}
		__antithesis_instrumentation__.Notify(615473)
		return typedLeft, typedRight, fn, false, nil
	default:
		__antithesis_instrumentation__.Notify(615474)
	}
	__antithesis_instrumentation__.Notify(615455)

	typedSubExprs, fns, err := typeCheckOverloadedExprs(
		ctx, semaCtx, types.Any, ops, true, foldedLeft, foldedRight,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(615497)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(615498)
	}
	__antithesis_instrumentation__.Notify(615456)

	leftExpr, rightExpr := typedSubExprs[0], typedSubExprs[1]
	if switched {
		__antithesis_instrumentation__.Notify(615499)
		leftExpr, rightExpr = rightExpr, leftExpr
	} else {
		__antithesis_instrumentation__.Notify(615500)
	}
	__antithesis_instrumentation__.Notify(615457)
	leftReturn := leftExpr.ResolvedType()
	rightReturn := rightExpr.ResolvedType()
	leftFamily := leftReturn.Family()
	rightFamily := rightReturn.Family()

	nullComparison := false
	if leftFamily == types.UnknownFamily || func() bool {
		__antithesis_instrumentation__.Notify(615501)
		return rightFamily == types.UnknownFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(615502)
		nullComparison = true
		if len(fns) > 0 {
			__antithesis_instrumentation__.Notify(615503)
			noneAcceptNull := true
			for _, e := range fns {
				__antithesis_instrumentation__.Notify(615505)
				if e.(*CmpOp).NullableArgs {
					__antithesis_instrumentation__.Notify(615506)
					noneAcceptNull = false
					break
				} else {
					__antithesis_instrumentation__.Notify(615507)
				}
			}
			__antithesis_instrumentation__.Notify(615504)
			if noneAcceptNull {
				__antithesis_instrumentation__.Notify(615508)
				return leftExpr, rightExpr, nil, true, err
			} else {
				__antithesis_instrumentation__.Notify(615509)
			}
		} else {
			__antithesis_instrumentation__.Notify(615510)
		}
	} else {
		__antithesis_instrumentation__.Notify(615511)
	}
	__antithesis_instrumentation__.Notify(615458)

	leftIsGeneric := leftFamily == types.CollatedStringFamily || func() bool {
		__antithesis_instrumentation__.Notify(615512)
		return leftFamily == types.ArrayFamily == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(615513)
		return leftFamily == types.EnumFamily == true
	}() == true
	rightIsGeneric := rightFamily == types.CollatedStringFamily || func() bool {
		__antithesis_instrumentation__.Notify(615514)
		return rightFamily == types.ArrayFamily == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(615515)
		return rightFamily == types.EnumFamily == true
	}() == true
	genericComparison := leftIsGeneric && func() bool {
		__antithesis_instrumentation__.Notify(615516)
		return rightIsGeneric == true
	}() == true

	typeMismatch := false
	if genericComparison && func() bool {
		__antithesis_instrumentation__.Notify(615517)
		return !nullComparison == true
	}() == true {
		__antithesis_instrumentation__.Notify(615518)

		typeMismatch = !leftReturn.Equivalent(rightReturn)
	} else {
		__antithesis_instrumentation__.Notify(615519)
	}
	__antithesis_instrumentation__.Notify(615459)

	if len(fns) != 1 || func() bool {
		__antithesis_instrumentation__.Notify(615520)
		return typeMismatch == true
	}() == true {
		__antithesis_instrumentation__.Notify(615521)
		sig := fmt.Sprintf(compSignatureFmt, leftReturn, op, rightReturn)
		if len(fns) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(615523)
			return typeMismatch == true
		}() == true {
			__antithesis_instrumentation__.Notify(615524)

			if typeMismatch && func() bool {
				__antithesis_instrumentation__.Notify(615526)
				return leftFamily == types.EnumFamily == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(615527)
				return rightFamily == types.EnumFamily == true
			}() == true {
				__antithesis_instrumentation__.Notify(615528)
				return nil, nil, nil, false,
					pgerror.Newf(pgcode.InvalidParameterValue, invalidCompErrFmt, "enum", sig)
			} else {
				__antithesis_instrumentation__.Notify(615529)
			}
			__antithesis_instrumentation__.Notify(615525)
			return nil, nil, nil, false,
				pgerror.Newf(pgcode.InvalidParameterValue, unsupportedCompErrFmt, sig)
		} else {
			__antithesis_instrumentation__.Notify(615530)
		}
		__antithesis_instrumentation__.Notify(615522)
		fnsStr := formatCandidates(op.String(), fns)
		err = pgerror.Newf(pgcode.AmbiguousFunction, ambiguousCompErrFmt, sig)
		err = errors.WithHintf(err, candidatesHintFmt, fnsStr)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(615531)
	}
	__antithesis_instrumentation__.Notify(615460)

	return leftExpr, rightExpr, fns[0].(*CmpOp), false, nil
}

type typeCheckExprsState struct {
	ctx     context.Context
	semaCtx *SemaContext

	exprs           []Expr
	typedExprs      []TypedExpr
	constIdxs       []int
	placeholderIdxs []int
	resolvableIdxs  []int
}

func TypeCheckSameTypedExprs(
	ctx context.Context, semaCtx *SemaContext, desired *types.T, exprs ...Expr,
) ([]TypedExpr, *types.T, error) {
	__antithesis_instrumentation__.Notify(615532)
	switch len(exprs) {
	case 0:
		__antithesis_instrumentation__.Notify(615535)
		return nil, nil, nil
	case 1:
		__antithesis_instrumentation__.Notify(615536)
		typedExpr, err := exprs[0].TypeCheck(ctx, semaCtx, desired)
		if err != nil {
			__antithesis_instrumentation__.Notify(615540)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(615541)
		}
		__antithesis_instrumentation__.Notify(615537)
		typ := typedExpr.ResolvedType()
		if typ == types.Unknown && func() bool {
			__antithesis_instrumentation__.Notify(615542)
			return desired != types.Any == true
		}() == true {
			__antithesis_instrumentation__.Notify(615543)

			typ = desired
		} else {
			__antithesis_instrumentation__.Notify(615544)
		}
		__antithesis_instrumentation__.Notify(615538)
		return []TypedExpr{typedExpr}, typ, nil
	default:
		__antithesis_instrumentation__.Notify(615539)
	}
	__antithesis_instrumentation__.Notify(615533)

	if _, ok := exprs[0].(*Tuple); ok {
		__antithesis_instrumentation__.Notify(615545)
		return typeCheckSameTypedTupleExprs(ctx, semaCtx, desired, exprs...)
	} else {
		__antithesis_instrumentation__.Notify(615546)
	}
	__antithesis_instrumentation__.Notify(615534)

	typedExprs := make([]TypedExpr, len(exprs))

	constIdxs, placeholderIdxs, resolvableIdxs := typeCheckSplitExprs(ctx, semaCtx, exprs)

	s := typeCheckExprsState{
		ctx:             ctx,
		semaCtx:         semaCtx,
		exprs:           exprs,
		typedExprs:      typedExprs,
		constIdxs:       constIdxs,
		placeholderIdxs: placeholderIdxs,
		resolvableIdxs:  resolvableIdxs,
	}

	switch {
	case len(resolvableIdxs) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(615556)
		return len(constIdxs) == 0 == true
	}() == true:
		__antithesis_instrumentation__.Notify(615547)
		if err := typeCheckSameTypedPlaceholders(s, desired); err != nil {
			__antithesis_instrumentation__.Notify(615557)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(615558)
		}
		__antithesis_instrumentation__.Notify(615548)
		return typedExprs, desired, nil
	case len(resolvableIdxs) == 0:
		__antithesis_instrumentation__.Notify(615549)
		return typeCheckConstsAndPlaceholdersWithDesired(s, desired)
	default:
		__antithesis_instrumentation__.Notify(615550)
		firstValidIdx := -1
		firstValidType := types.Unknown
		for i, j := range resolvableIdxs {
			__antithesis_instrumentation__.Notify(615559)
			typedExpr, err := exprs[j].TypeCheck(ctx, semaCtx, desired)
			if err != nil {
				__antithesis_instrumentation__.Notify(615561)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(615562)
			}
			__antithesis_instrumentation__.Notify(615560)
			typedExprs[j] = typedExpr
			if returnType := typedExpr.ResolvedType(); returnType.Family() != types.UnknownFamily {
				__antithesis_instrumentation__.Notify(615563)
				firstValidType = returnType
				firstValidIdx = i
				break
			} else {
				__antithesis_instrumentation__.Notify(615564)
			}
		}
		__antithesis_instrumentation__.Notify(615551)

		if firstValidType.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(615565)

			switch {
			case len(constIdxs) > 0:
				__antithesis_instrumentation__.Notify(615566)
				return typeCheckConstsAndPlaceholdersWithDesired(s, desired)
			case len(placeholderIdxs) > 0:
				__antithesis_instrumentation__.Notify(615567)
				p := s.exprs[placeholderIdxs[0]].(*Placeholder)
				return nil, nil, placeholderTypeAmbiguityError(p.Idx)
			default:
				__antithesis_instrumentation__.Notify(615568)
				if desired != types.Any {
					__antithesis_instrumentation__.Notify(615570)
					return typedExprs, desired, nil
				} else {
					__antithesis_instrumentation__.Notify(615571)
				}
				__antithesis_instrumentation__.Notify(615569)
				return typedExprs, types.Unknown, nil
			}
		} else {
			__antithesis_instrumentation__.Notify(615572)
		}
		__antithesis_instrumentation__.Notify(615552)

		for _, i := range resolvableIdxs[firstValidIdx+1:] {
			__antithesis_instrumentation__.Notify(615573)
			typedExpr, err := exprs[i].TypeCheck(ctx, semaCtx, firstValidType)
			if err != nil {
				__antithesis_instrumentation__.Notify(615576)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(615577)
			}
			__antithesis_instrumentation__.Notify(615574)

			if typ := typedExpr.ResolvedType(); !(typ.Equivalent(firstValidType) || func() bool {
				__antithesis_instrumentation__.Notify(615578)
				return typ.Family() == types.UnknownFamily == true
			}() == true) {
				__antithesis_instrumentation__.Notify(615579)
				return nil, nil, unexpectedTypeError(exprs[i], firstValidType, typ)
			} else {
				__antithesis_instrumentation__.Notify(615580)
			}
			__antithesis_instrumentation__.Notify(615575)
			typedExprs[i] = typedExpr
		}
		__antithesis_instrumentation__.Notify(615553)
		if len(constIdxs) > 0 {
			__antithesis_instrumentation__.Notify(615581)
			if _, err := typeCheckSameTypedConsts(s, firstValidType, true); err != nil {
				__antithesis_instrumentation__.Notify(615582)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(615583)
			}
		} else {
			__antithesis_instrumentation__.Notify(615584)
		}
		__antithesis_instrumentation__.Notify(615554)
		if len(placeholderIdxs) > 0 {
			__antithesis_instrumentation__.Notify(615585)
			if err := typeCheckSameTypedPlaceholders(s, firstValidType); err != nil {
				__antithesis_instrumentation__.Notify(615586)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(615587)
			}
		} else {
			__antithesis_instrumentation__.Notify(615588)
		}
		__antithesis_instrumentation__.Notify(615555)
		return typedExprs, firstValidType, nil
	}
}

func typeCheckSameTypedPlaceholders(s typeCheckExprsState, typ *types.T) error {
	__antithesis_instrumentation__.Notify(615589)
	for _, i := range s.placeholderIdxs {
		__antithesis_instrumentation__.Notify(615591)
		typedExpr, err := typeCheckAndRequire(s.ctx, s.semaCtx, s.exprs[i], typ, "placeholder")
		if err != nil {
			__antithesis_instrumentation__.Notify(615593)
			return err
		} else {
			__antithesis_instrumentation__.Notify(615594)
		}
		__antithesis_instrumentation__.Notify(615592)
		s.typedExprs[i] = typedExpr
	}
	__antithesis_instrumentation__.Notify(615590)
	return nil
}

func typeCheckSameTypedConsts(
	s typeCheckExprsState, typ *types.T, required bool,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(615595)
	setTypeForConsts := func(typ *types.T) (*types.T, error) {
		__antithesis_instrumentation__.Notify(615600)
		for _, i := range s.constIdxs {
			__antithesis_instrumentation__.Notify(615602)
			typedExpr, err := typeCheckAndRequire(s.ctx, s.semaCtx, s.exprs[i], typ, "constant")
			if err != nil {
				__antithesis_instrumentation__.Notify(615604)

				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(615605)
			}
			__antithesis_instrumentation__.Notify(615603)
			s.typedExprs[i] = typedExpr
		}
		__antithesis_instrumentation__.Notify(615601)
		return typ, nil
	}
	__antithesis_instrumentation__.Notify(615596)

	if typ.Family() != types.AnyFamily {
		__antithesis_instrumentation__.Notify(615606)
		all := true
		for _, i := range s.constIdxs {
			__antithesis_instrumentation__.Notify(615608)
			if !canConstantBecome(s.exprs[i].(Constant), typ) {
				__antithesis_instrumentation__.Notify(615609)
				if required {
					__antithesis_instrumentation__.Notify(615611)
					typedExpr, err := s.exprs[i].TypeCheck(s.ctx, s.semaCtx, types.Any)
					if err != nil {
						__antithesis_instrumentation__.Notify(615613)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(615614)
					}
					__antithesis_instrumentation__.Notify(615612)
					return nil, unexpectedTypeError(s.exprs[i], typ, typedExpr.ResolvedType())
				} else {
					__antithesis_instrumentation__.Notify(615615)
				}
				__antithesis_instrumentation__.Notify(615610)
				all = false
				break
			} else {
				__antithesis_instrumentation__.Notify(615616)
			}
		}
		__antithesis_instrumentation__.Notify(615607)
		if all {
			__antithesis_instrumentation__.Notify(615617)

			return setTypeForConsts(typ.WithoutTypeModifiers())
		} else {
			__antithesis_instrumentation__.Notify(615618)
		}
	} else {
		__antithesis_instrumentation__.Notify(615619)
	}
	__antithesis_instrumentation__.Notify(615597)

	if bestType, ok := commonConstantType(s.exprs, s.constIdxs); ok {
		__antithesis_instrumentation__.Notify(615620)
		return setTypeForConsts(bestType)
	} else {
		__antithesis_instrumentation__.Notify(615621)
	}
	__antithesis_instrumentation__.Notify(615598)

	reqTyp := typ
	for _, i := range s.constIdxs {
		__antithesis_instrumentation__.Notify(615622)
		typedExpr, err := s.exprs[i].TypeCheck(s.ctx, s.semaCtx, reqTyp)
		if err != nil {
			__antithesis_instrumentation__.Notify(615625)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(615626)
		}
		__antithesis_instrumentation__.Notify(615623)
		if typ := typedExpr.ResolvedType(); !typ.Equivalent(reqTyp) {
			__antithesis_instrumentation__.Notify(615627)
			return nil, unexpectedTypeError(s.exprs[i], reqTyp, typ)
		} else {
			__antithesis_instrumentation__.Notify(615628)
		}
		__antithesis_instrumentation__.Notify(615624)
		if reqTyp.Family() == types.AnyFamily {
			__antithesis_instrumentation__.Notify(615629)
			reqTyp = typedExpr.ResolvedType()
		} else {
			__antithesis_instrumentation__.Notify(615630)
		}
	}
	__antithesis_instrumentation__.Notify(615599)
	return nil, errors.AssertionFailedf("should throw error above")
}

func typeCheckConstsAndPlaceholdersWithDesired(
	s typeCheckExprsState, desired *types.T,
) ([]TypedExpr, *types.T, error) {
	__antithesis_instrumentation__.Notify(615631)
	typ, err := typeCheckSameTypedConsts(s, desired, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(615634)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(615635)
	}
	__antithesis_instrumentation__.Notify(615632)
	if len(s.placeholderIdxs) > 0 {
		__antithesis_instrumentation__.Notify(615636)
		if err := typeCheckSameTypedPlaceholders(s, typ); err != nil {
			__antithesis_instrumentation__.Notify(615637)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(615638)
		}
	} else {
		__antithesis_instrumentation__.Notify(615639)
	}
	__antithesis_instrumentation__.Notify(615633)
	return s.typedExprs, typ, nil
}

func typeCheckSplitExprs(
	ctx context.Context, semaCtx *SemaContext, exprs []Expr,
) (constIdxs []int, placeholderIdxs []int, resolvableIdxs []int) {
	__antithesis_instrumentation__.Notify(615640)
	for i, expr := range exprs {
		__antithesis_instrumentation__.Notify(615642)
		switch {
		case isConstant(expr):
			__antithesis_instrumentation__.Notify(615643)
			constIdxs = append(constIdxs, i)
		case semaCtx.isUnresolvedPlaceholder(expr):
			__antithesis_instrumentation__.Notify(615644)
			placeholderIdxs = append(placeholderIdxs, i)
		default:
			__antithesis_instrumentation__.Notify(615645)
			resolvableIdxs = append(resolvableIdxs, i)
		}
	}
	__antithesis_instrumentation__.Notify(615641)
	return constIdxs, placeholderIdxs, resolvableIdxs
}

func typeCheckTupleComparison(
	ctx context.Context,
	semaCtx *SemaContext,
	op treecmp.ComparisonOperator,
	left *Tuple,
	right *Tuple,
) (TypedExpr, TypedExpr, error) {
	__antithesis_instrumentation__.Notify(615646)

	tupLen := len(left.Exprs)
	if err := checkTupleHasLength(right, tupLen); err != nil {
		__antithesis_instrumentation__.Notify(615649)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(615650)
	}
	__antithesis_instrumentation__.Notify(615647)
	left.typ = types.MakeTuple(make([]*types.T, tupLen))
	right.typ = types.MakeTuple(make([]*types.T, tupLen))
	for elemIdx := range left.Exprs {
		__antithesis_instrumentation__.Notify(615651)
		leftSubExpr := left.Exprs[elemIdx]
		rightSubExpr := right.Exprs[elemIdx]
		leftSubExprTyped, rightSubExprTyped, _, _, err := typeCheckComparisonOp(ctx, semaCtx, op, leftSubExpr, rightSubExpr)
		if err != nil {
			__antithesis_instrumentation__.Notify(615653)
			exps := Exprs([]Expr{left, right})
			return nil, nil, pgerror.Wrapf(err, pgcode.DatatypeMismatch, "tuples %s are not comparable at index %d",
				&exps, elemIdx+1)
		} else {
			__antithesis_instrumentation__.Notify(615654)
		}
		__antithesis_instrumentation__.Notify(615652)
		left.Exprs[elemIdx] = leftSubExprTyped
		left.typ.TupleContents()[elemIdx] = leftSubExprTyped.ResolvedType()
		right.Exprs[elemIdx] = rightSubExprTyped
		right.typ.TupleContents()[elemIdx] = rightSubExprTyped.ResolvedType()
	}
	__antithesis_instrumentation__.Notify(615648)
	return left, right, nil
}

func typeCheckSameTypedTupleExprs(
	ctx context.Context, semaCtx *SemaContext, desired *types.T, exprs ...Expr,
) ([]TypedExpr, *types.T, error) {
	__antithesis_instrumentation__.Notify(615655)

	first := exprs[0].(*Tuple)
	if err := checkAllExprsAreTuplesOrNulls(ctx, semaCtx, exprs[1:]); err != nil {
		__antithesis_instrumentation__.Notify(615660)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(615661)
	}
	__antithesis_instrumentation__.Notify(615656)

	firstLen := len(first.Exprs)
	if err := checkAllTuplesHaveLength(exprs[1:], firstLen); err != nil {
		__antithesis_instrumentation__.Notify(615662)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(615663)
	}
	__antithesis_instrumentation__.Notify(615657)

	resTypes := types.MakeLabeledTuple(make([]*types.T, firstLen), first.Labels)
	sameTypeExprs := make([]Expr, 0, len(exprs))

	sameTypeExprsIndices := make([]int, 0, len(exprs))
	for elemIdx := range first.Exprs {
		__antithesis_instrumentation__.Notify(615664)
		sameTypeExprs = sameTypeExprs[:0]
		sameTypeExprsIndices = sameTypeExprsIndices[:0]
		for exprIdx, expr := range exprs {
			__antithesis_instrumentation__.Notify(615669)

			if _, isTuple := expr.(*Tuple); !isTuple {
				__antithesis_instrumentation__.Notify(615671)
				continue
			} else {
				__antithesis_instrumentation__.Notify(615672)
			}
			__antithesis_instrumentation__.Notify(615670)
			sameTypeExprs = append(sameTypeExprs, expr.(*Tuple).Exprs[elemIdx])
			sameTypeExprsIndices = append(sameTypeExprsIndices, exprIdx)
		}
		__antithesis_instrumentation__.Notify(615665)
		desiredElem := types.Any
		if len(desired.TupleContents()) > elemIdx {
			__antithesis_instrumentation__.Notify(615673)
			desiredElem = desired.TupleContents()[elemIdx]
		} else {
			__antithesis_instrumentation__.Notify(615674)
		}
		__antithesis_instrumentation__.Notify(615666)
		typedSubExprs, resType, err := TypeCheckSameTypedExprs(ctx, semaCtx, desiredElem, sameTypeExprs...)
		if err != nil {
			__antithesis_instrumentation__.Notify(615675)
			return nil, nil, pgerror.Wrapf(err, pgcode.DatatypeMismatch, "tuples %s are not the same type", Exprs(exprs))
		} else {
			__antithesis_instrumentation__.Notify(615676)
		}
		__antithesis_instrumentation__.Notify(615667)
		for j, typedExpr := range typedSubExprs {
			__antithesis_instrumentation__.Notify(615677)
			tupleIdx := sameTypeExprsIndices[j]
			exprs[tupleIdx].(*Tuple).Exprs[elemIdx] = typedExpr
		}
		__antithesis_instrumentation__.Notify(615668)
		resTypes.TupleContents()[elemIdx] = resType
	}
	__antithesis_instrumentation__.Notify(615658)

	typedExprs := make([]TypedExpr, len(exprs))
	for tupleIdx, expr := range exprs {
		__antithesis_instrumentation__.Notify(615678)
		if t, isTuple := expr.(*Tuple); isTuple {
			__antithesis_instrumentation__.Notify(615679)

			t.typ = resTypes
			typedExprs[tupleIdx] = t
		} else {
			__antithesis_instrumentation__.Notify(615680)
			typedExpr, err := expr.TypeCheck(ctx, semaCtx, resTypes)
			if err != nil {
				__antithesis_instrumentation__.Notify(615683)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(615684)
			}
			__antithesis_instrumentation__.Notify(615681)
			if !typedExpr.ResolvedType().EquivalentOrNull(resTypes, true) {
				__antithesis_instrumentation__.Notify(615685)
				return nil, nil, unexpectedTypeError(expr, resTypes, typedExpr.ResolvedType())
			} else {
				__antithesis_instrumentation__.Notify(615686)
			}
			__antithesis_instrumentation__.Notify(615682)
			typedExprs[tupleIdx] = typedExpr
		}
	}
	__antithesis_instrumentation__.Notify(615659)
	return typedExprs, resTypes, nil
}

func checkAllExprsAreTuplesOrNulls(ctx context.Context, semaCtx *SemaContext, exprs []Expr) error {
	__antithesis_instrumentation__.Notify(615687)
	for _, expr := range exprs {
		__antithesis_instrumentation__.Notify(615689)
		_, isTuple := expr.(*Tuple)
		isNull, err := isNullOrAnnotatedNullTuple(ctx, semaCtx, expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(615691)
			return err
		} else {
			__antithesis_instrumentation__.Notify(615692)
		}
		__antithesis_instrumentation__.Notify(615690)
		if !(isTuple || func() bool {
			__antithesis_instrumentation__.Notify(615693)
			return isNull == true
		}() == true) {
			__antithesis_instrumentation__.Notify(615694)

			typedExpr, err := expr.TypeCheck(ctx, semaCtx, types.Any)
			if err != nil {
				__antithesis_instrumentation__.Notify(615696)
				return err
			} else {
				__antithesis_instrumentation__.Notify(615697)
			}
			__antithesis_instrumentation__.Notify(615695)
			if typedExpr.ResolvedType().Family() != types.TupleFamily {
				__antithesis_instrumentation__.Notify(615698)
				return unexpectedTypeError(expr, types.AnyTuple, typedExpr.ResolvedType())
			} else {
				__antithesis_instrumentation__.Notify(615699)
			}
		} else {
			__antithesis_instrumentation__.Notify(615700)
		}
	}
	__antithesis_instrumentation__.Notify(615688)
	return nil
}

func checkAllTuplesHaveLength(exprs []Expr, expectedLen int) error {
	__antithesis_instrumentation__.Notify(615701)
	for _, expr := range exprs {
		__antithesis_instrumentation__.Notify(615703)
		if t, isTuple := expr.(*Tuple); isTuple {
			__antithesis_instrumentation__.Notify(615704)
			if err := checkTupleHasLength(t, expectedLen); err != nil {
				__antithesis_instrumentation__.Notify(615705)
				return err
			} else {
				__antithesis_instrumentation__.Notify(615706)
			}
		} else {
			__antithesis_instrumentation__.Notify(615707)
		}
	}
	__antithesis_instrumentation__.Notify(615702)
	return nil
}

func checkTupleHasLength(t *Tuple, expectedLen int) error {
	__antithesis_instrumentation__.Notify(615708)
	if len(t.Exprs) != expectedLen {
		__antithesis_instrumentation__.Notify(615710)
		return pgerror.Newf(pgcode.DatatypeMismatch, "expected tuple %v to have a length of %d", t, expectedLen)
	} else {
		__antithesis_instrumentation__.Notify(615711)
	}
	__antithesis_instrumentation__.Notify(615709)
	return nil
}

func isNullOrAnnotatedNullTuple(
	ctx context.Context, semaCtx *SemaContext, expr Expr,
) (bool, error) {
	__antithesis_instrumentation__.Notify(615712)
	if expr == DNull {
		__antithesis_instrumentation__.Notify(615715)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(615716)
	}
	__antithesis_instrumentation__.Notify(615713)
	if annotate, ok := expr.(*AnnotateTypeExpr); ok && func() bool {
		__antithesis_instrumentation__.Notify(615717)
		return annotate.Expr == DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(615718)
		annotateType, err := ResolveType(ctx, annotate.Type, semaCtx.GetTypeResolver())
		if err != nil {
			__antithesis_instrumentation__.Notify(615720)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(615721)
		}
		__antithesis_instrumentation__.Notify(615719)
		return annotateType.Identical(types.AnyTuple), nil
	} else {
		__antithesis_instrumentation__.Notify(615722)
	}
	__antithesis_instrumentation__.Notify(615714)
	return false, nil
}

type placeholderAnnotationVisitor struct {
	types PlaceholderTypes
	state []annotationState
	err   error
	ctx   *SemaContext

	errIdx PlaceholderIdx
}

type annotationState uint8

const (
	noType annotationState = iota

	typeFromHint

	typeFromAnnotation

	typeFromCast

	conflictingCasts
)

func (v *placeholderAnnotationVisitor) setErr(idx PlaceholderIdx, err error) {
	__antithesis_instrumentation__.Notify(615723)
	if v.err == nil || func() bool {
		__antithesis_instrumentation__.Notify(615724)
		return v.errIdx >= idx == true
	}() == true {
		__antithesis_instrumentation__.Notify(615725)
		v.err = err
		v.errIdx = idx
	} else {
		__antithesis_instrumentation__.Notify(615726)
	}
}

func (v *placeholderAnnotationVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(615727)
	switch t := expr.(type) {
	case *AnnotateTypeExpr:
		__antithesis_instrumentation__.Notify(615729)
		if arg, ok := t.Expr.(*Placeholder); ok {
			__antithesis_instrumentation__.Notify(615732)
			tType, err := ResolveType(context.TODO(), t.Type, v.ctx.GetTypeResolver())
			if err != nil {
				__antithesis_instrumentation__.Notify(615735)
				v.setErr(arg.Idx, err)
				return false, expr
			} else {
				__antithesis_instrumentation__.Notify(615736)
			}
			__antithesis_instrumentation__.Notify(615733)
			switch v.state[arg.Idx] {
			case noType, typeFromCast, conflictingCasts:
				__antithesis_instrumentation__.Notify(615737)

				v.types[arg.Idx] = tType
				v.state[arg.Idx] = typeFromAnnotation

			case typeFromAnnotation:
				__antithesis_instrumentation__.Notify(615738)

				if !tType.Equivalent(v.types[arg.Idx]) {
					__antithesis_instrumentation__.Notify(615741)
					v.setErr(arg.Idx, pgerror.Newf(
						pgcode.DatatypeMismatch,
						"multiple conflicting type annotations around %s",
						arg.Idx,
					))
				} else {
					__antithesis_instrumentation__.Notify(615742)
				}

			case typeFromHint:
				__antithesis_instrumentation__.Notify(615739)

				if prevType := v.types[arg.Idx]; !tType.Equivalent(prevType) {
					__antithesis_instrumentation__.Notify(615743)
					v.setErr(arg.Idx, pgerror.Newf(
						pgcode.DatatypeMismatch,
						"type annotation around %s conflicts with specified type %s",
						arg.Idx, v.types[arg.Idx],
					))
				} else {
					__antithesis_instrumentation__.Notify(615744)
				}

			default:
				__antithesis_instrumentation__.Notify(615740)
				panic(errors.AssertionFailedf("unhandled state: %v", errors.Safe(v.state[arg.Idx])))
			}
			__antithesis_instrumentation__.Notify(615734)
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(615745)
		}

	case *CastExpr:
		__antithesis_instrumentation__.Notify(615730)
		if arg, ok := t.Expr.(*Placeholder); ok {
			__antithesis_instrumentation__.Notify(615746)
			tType, err := ResolveType(context.TODO(), t.Type, v.ctx.GetTypeResolver())
			if err != nil {
				__antithesis_instrumentation__.Notify(615749)
				v.setErr(arg.Idx, err)
				return false, expr
			} else {
				__antithesis_instrumentation__.Notify(615750)
			}
			__antithesis_instrumentation__.Notify(615747)
			switch v.state[arg.Idx] {
			case noType:
				__antithesis_instrumentation__.Notify(615751)
				v.types[arg.Idx] = tType
				v.state[arg.Idx] = typeFromCast

			case typeFromCast:
				__antithesis_instrumentation__.Notify(615752)

				if !tType.Equivalent(v.types[arg.Idx]) {
					__antithesis_instrumentation__.Notify(615756)
					v.state[arg.Idx] = conflictingCasts
					v.types[arg.Idx] = nil
				} else {
					__antithesis_instrumentation__.Notify(615757)
				}

			case typeFromHint, typeFromAnnotation:
				__antithesis_instrumentation__.Notify(615753)

			case conflictingCasts:
				__antithesis_instrumentation__.Notify(615754)

			default:
				__antithesis_instrumentation__.Notify(615755)
				panic(errors.AssertionFailedf("unhandled state: %v", v.state[arg.Idx]))
			}
			__antithesis_instrumentation__.Notify(615748)
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(615758)
		}

	case *Placeholder:
		__antithesis_instrumentation__.Notify(615731)
		switch v.state[t.Idx] {
		case noType, typeFromCast:
			__antithesis_instrumentation__.Notify(615759)

			v.state[t.Idx] = conflictingCasts
			v.types[t.Idx] = nil

		case typeFromHint, typeFromAnnotation:
			__antithesis_instrumentation__.Notify(615760)

		case conflictingCasts:
			__antithesis_instrumentation__.Notify(615761)

		default:
			__antithesis_instrumentation__.Notify(615762)
			panic(errors.AssertionFailedf("unhandled state: %v", v.state[t.Idx]))
		}
	}
	__antithesis_instrumentation__.Notify(615728)
	return true, expr
}

func (*placeholderAnnotationVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(615763)
	return expr
}

func ProcessPlaceholderAnnotations(
	semaCtx *SemaContext, stmt Statement, typeHints PlaceholderTypes,
) error {
	__antithesis_instrumentation__.Notify(615764)
	v := placeholderAnnotationVisitor{
		types: typeHints,
		state: make([]annotationState, len(typeHints)),
		ctx:   semaCtx,
	}

	for placeholder := range typeHints {
		__antithesis_instrumentation__.Notify(615766)
		if typeHints[placeholder] != nil {
			__antithesis_instrumentation__.Notify(615767)
			v.state[placeholder] = typeFromHint
		} else {
			__antithesis_instrumentation__.Notify(615768)
		}
	}
	__antithesis_instrumentation__.Notify(615765)

	walkStmt(&v, stmt)
	return v.err
}

func StripMemoizedFuncs(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(615769)
	expr, _ = WalkExpr(stripFuncsVisitor{}, expr)
	return expr
}

type stripFuncsVisitor struct{}

func (v stripFuncsVisitor) VisitPre(expr Expr) (recurse bool, newExpr Expr) {
	__antithesis_instrumentation__.Notify(615770)
	switch t := expr.(type) {
	case *UnaryExpr:
		__antithesis_instrumentation__.Notify(615772)
		t.fn = nil
	case *BinaryExpr:
		__antithesis_instrumentation__.Notify(615773)
		t.Fn = nil
	case *ComparisonExpr:
		__antithesis_instrumentation__.Notify(615774)
		t.Fn = nil
	case *FuncExpr:
		__antithesis_instrumentation__.Notify(615775)
		t.fn = nil
		t.fnProps = nil
	}
	__antithesis_instrumentation__.Notify(615771)
	return true, expr
}

func (stripFuncsVisitor) VisitPost(expr Expr) Expr {
	__antithesis_instrumentation__.Notify(615776)
	return expr
}
