package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/lib/pq/oid"
)

type SpecializedVectorizedBuiltin int

const (
	_ SpecializedVectorizedBuiltin = iota
	SubstringStringIntInt
)

type Overload struct {
	Types      TypeList
	ReturnType ReturnTyper
	Volatility Volatility

	PreferredOverload bool

	Info string

	AggregateFunc func([]*types.T, *EvalContext, Datums) AggregateFunc
	WindowFunc    func([]*types.T, *EvalContext) WindowFunc

	Fn func(*EvalContext, Datums) (Datum, error)

	FnWithExprs func(*EvalContext, Exprs) (Datum, error)

	Generator GeneratorFactory

	GeneratorWithExprs GeneratorWithExprsFactory

	SQLFn func(*EvalContext, Datums) (string, error)

	counter telemetry.Counter

	SpecializedVecBuiltin SpecializedVectorizedBuiltin

	IgnoreVolatilityCheck bool

	Oid oid.Oid

	DistsqlBlocklist bool
}

func (b Overload) params() TypeList { __antithesis_instrumentation__.Notify(611024); return b.Types }

func (b Overload) returnType() ReturnTyper {
	__antithesis_instrumentation__.Notify(611025)
	return b.ReturnType
}

func (b Overload) preferred() bool {
	__antithesis_instrumentation__.Notify(611026)
	return b.PreferredOverload
}

func (b Overload) FixedReturnType() *types.T {
	__antithesis_instrumentation__.Notify(611027)
	if b.ReturnType == nil {
		__antithesis_instrumentation__.Notify(611029)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(611030)
	}
	__antithesis_instrumentation__.Notify(611028)
	return returnTypeToFixedType(b.ReturnType, nil)
}

func (b Overload) InferReturnTypeFromInputArgTypes(inputTypes []*types.T) *types.T {
	__antithesis_instrumentation__.Notify(611031)
	retTyp := b.FixedReturnType()

	if retTyp.IsAmbiguous() {
		__antithesis_instrumentation__.Notify(611033)
		args := make([]TypedExpr, len(inputTypes))
		for i, t := range inputTypes {
			__antithesis_instrumentation__.Notify(611035)
			args[i] = &TypedDummy{Typ: t}
		}
		__antithesis_instrumentation__.Notify(611034)

		retTyp = returnTypeToFixedType(b.ReturnType, args)
	} else {
		__antithesis_instrumentation__.Notify(611036)
	}
	__antithesis_instrumentation__.Notify(611032)
	return retTyp
}

func (b Overload) IsGenerator() bool {
	__antithesis_instrumentation__.Notify(611037)
	return b.Generator != nil || func() bool {
		__antithesis_instrumentation__.Notify(611038)
		return b.GeneratorWithExprs != nil == true
	}() == true
}

func (b Overload) Signature(simplify bool) string {
	__antithesis_instrumentation__.Notify(611039)
	retType := b.FixedReturnType()
	if simplify {
		__antithesis_instrumentation__.Notify(611041)
		if retType.Family() == types.TupleFamily && func() bool {
			__antithesis_instrumentation__.Notify(611042)
			return len(retType.TupleContents()) == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(611043)
			retType = retType.TupleContents()[0]
		} else {
			__antithesis_instrumentation__.Notify(611044)
		}
	} else {
		__antithesis_instrumentation__.Notify(611045)
	}
	__antithesis_instrumentation__.Notify(611040)
	return fmt.Sprintf("(%s) -> %s", b.Types.String(), retType)
}

type overloadImpl interface {
	params() TypeList
	returnType() ReturnTyper

	preferred() bool
}

var _ overloadImpl = &Overload{}
var _ overloadImpl = &UnaryOp{}
var _ overloadImpl = &BinOp{}
var _ overloadImpl = &CmpOp{}

func GetParamsAndReturnType(impl overloadImpl) (TypeList, ReturnTyper) {
	__antithesis_instrumentation__.Notify(611046)
	return impl.params(), impl.returnType()
}

type TypeList interface {
	Match(types []*types.T) bool

	MatchAt(typ *types.T, i int) bool

	MatchLen(l int) bool

	GetAt(i int) *types.T

	Length() int

	Types() []*types.T

	String() string
}

var _ TypeList = ArgTypes{}
var _ TypeList = HomogeneousType{}
var _ TypeList = VariadicType{}

type ArgTypes []struct {
	Name string
	Typ  *types.T
}

func (a ArgTypes) Match(types []*types.T) bool {
	__antithesis_instrumentation__.Notify(611047)
	if len(types) != len(a) {
		__antithesis_instrumentation__.Notify(611050)
		return false
	} else {
		__antithesis_instrumentation__.Notify(611051)
	}
	__antithesis_instrumentation__.Notify(611048)
	for i := range types {
		__antithesis_instrumentation__.Notify(611052)
		if !a.MatchAt(types[i], i) {
			__antithesis_instrumentation__.Notify(611053)
			return false
		} else {
			__antithesis_instrumentation__.Notify(611054)
		}
	}
	__antithesis_instrumentation__.Notify(611049)
	return true
}

func (a ArgTypes) MatchAt(typ *types.T, i int) bool {
	__antithesis_instrumentation__.Notify(611055)

	if typ.Family() == types.TupleFamily {
		__antithesis_instrumentation__.Notify(611057)
		typ = types.AnyTuple
	} else {
		__antithesis_instrumentation__.Notify(611058)
	}
	__antithesis_instrumentation__.Notify(611056)
	return i < len(a) && func() bool {
		__antithesis_instrumentation__.Notify(611059)
		return (typ.Family() == types.UnknownFamily || func() bool {
			__antithesis_instrumentation__.Notify(611060)
			return a[i].Typ.Equivalent(typ) == true
		}() == true) == true
	}() == true
}

func (a ArgTypes) MatchLen(l int) bool {
	__antithesis_instrumentation__.Notify(611061)
	return len(a) == l
}

func (a ArgTypes) GetAt(i int) *types.T {
	__antithesis_instrumentation__.Notify(611062)
	return a[i].Typ
}

func (a ArgTypes) Length() int {
	__antithesis_instrumentation__.Notify(611063)
	return len(a)
}

func (a ArgTypes) Types() []*types.T {
	__antithesis_instrumentation__.Notify(611064)
	n := len(a)
	ret := make([]*types.T, n)
	for i, s := range a {
		__antithesis_instrumentation__.Notify(611066)
		ret[i] = s.Typ
	}
	__antithesis_instrumentation__.Notify(611065)
	return ret
}

func (a ArgTypes) String() string {
	__antithesis_instrumentation__.Notify(611067)
	var s strings.Builder
	for i, arg := range a {
		__antithesis_instrumentation__.Notify(611069)
		if i > 0 {
			__antithesis_instrumentation__.Notify(611071)
			s.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(611072)
		}
		__antithesis_instrumentation__.Notify(611070)
		s.WriteString(arg.Name)
		s.WriteString(": ")
		s.WriteString(arg.Typ.String())
	}
	__antithesis_instrumentation__.Notify(611068)
	return s.String()
}

type HomogeneousType struct{}

func (HomogeneousType) Match(types []*types.T) bool {
	__antithesis_instrumentation__.Notify(611073)
	return true
}

func (HomogeneousType) MatchAt(typ *types.T, i int) bool {
	__antithesis_instrumentation__.Notify(611074)
	return true
}

func (HomogeneousType) MatchLen(l int) bool {
	__antithesis_instrumentation__.Notify(611075)
	return true
}

func (HomogeneousType) GetAt(i int) *types.T {
	__antithesis_instrumentation__.Notify(611076)
	return types.Any
}

func (HomogeneousType) Length() int {
	__antithesis_instrumentation__.Notify(611077)
	return 1
}

func (HomogeneousType) Types() []*types.T {
	__antithesis_instrumentation__.Notify(611078)
	return []*types.T{types.Any}
}

func (HomogeneousType) String() string {
	__antithesis_instrumentation__.Notify(611079)
	return "anyelement..."
}

type VariadicType struct {
	FixedTypes []*types.T
	VarType    *types.T
}

func (v VariadicType) Match(types []*types.T) bool {
	__antithesis_instrumentation__.Notify(611080)
	for i := range types {
		__antithesis_instrumentation__.Notify(611082)
		if !v.MatchAt(types[i], i) {
			__antithesis_instrumentation__.Notify(611083)
			return false
		} else {
			__antithesis_instrumentation__.Notify(611084)
		}
	}
	__antithesis_instrumentation__.Notify(611081)
	return true
}

func (v VariadicType) MatchAt(typ *types.T, i int) bool {
	__antithesis_instrumentation__.Notify(611085)
	if i < len(v.FixedTypes) {
		__antithesis_instrumentation__.Notify(611087)
		return typ.Family() == types.UnknownFamily || func() bool {
			__antithesis_instrumentation__.Notify(611088)
			return v.FixedTypes[i].Equivalent(typ) == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(611089)
	}
	__antithesis_instrumentation__.Notify(611086)
	return typ.Family() == types.UnknownFamily || func() bool {
		__antithesis_instrumentation__.Notify(611090)
		return v.VarType.Equivalent(typ) == true
	}() == true
}

func (v VariadicType) MatchLen(l int) bool {
	__antithesis_instrumentation__.Notify(611091)
	return l >= len(v.FixedTypes)
}

func (v VariadicType) GetAt(i int) *types.T {
	__antithesis_instrumentation__.Notify(611092)
	if i < len(v.FixedTypes) {
		__antithesis_instrumentation__.Notify(611094)
		return v.FixedTypes[i]
	} else {
		__antithesis_instrumentation__.Notify(611095)
	}
	__antithesis_instrumentation__.Notify(611093)
	return v.VarType
}

func (v VariadicType) Length() int {
	__antithesis_instrumentation__.Notify(611096)
	return len(v.FixedTypes) + 1
}

func (v VariadicType) Types() []*types.T {
	__antithesis_instrumentation__.Notify(611097)
	result := make([]*types.T, len(v.FixedTypes)+1)
	for i := range v.FixedTypes {
		__antithesis_instrumentation__.Notify(611099)
		result[i] = v.FixedTypes[i]
	}
	__antithesis_instrumentation__.Notify(611098)
	result[len(result)-1] = v.VarType
	return result
}

func (v VariadicType) String() string {
	__antithesis_instrumentation__.Notify(611100)
	var s bytes.Buffer
	for i, t := range v.FixedTypes {
		__antithesis_instrumentation__.Notify(611103)
		if i != 0 {
			__antithesis_instrumentation__.Notify(611105)
			s.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(611106)
		}
		__antithesis_instrumentation__.Notify(611104)
		s.WriteString(t.String())
	}
	__antithesis_instrumentation__.Notify(611101)
	if len(v.FixedTypes) > 0 {
		__antithesis_instrumentation__.Notify(611107)
		s.WriteString(", ")
	} else {
		__antithesis_instrumentation__.Notify(611108)
	}
	__antithesis_instrumentation__.Notify(611102)
	fmt.Fprintf(&s, "%s...", v.VarType)
	return s.String()
}

var UnknownReturnType *types.T

type ReturnTyper func(args []TypedExpr) *types.T

func FixedReturnType(typ *types.T) ReturnTyper {
	__antithesis_instrumentation__.Notify(611109)
	return func(args []TypedExpr) *types.T { __antithesis_instrumentation__.Notify(611110); return typ }
}

func IdentityReturnType(idx int) ReturnTyper {
	__antithesis_instrumentation__.Notify(611111)
	return func(args []TypedExpr) *types.T {
		__antithesis_instrumentation__.Notify(611112)
		if len(args) == 0 {
			__antithesis_instrumentation__.Notify(611114)
			return UnknownReturnType
		} else {
			__antithesis_instrumentation__.Notify(611115)
		}
		__antithesis_instrumentation__.Notify(611113)
		return args[idx].ResolvedType()
	}
}

func ArrayOfFirstNonNullReturnType() ReturnTyper {
	__antithesis_instrumentation__.Notify(611116)
	return func(args []TypedExpr) *types.T {
		__antithesis_instrumentation__.Notify(611117)
		if len(args) == 0 {
			__antithesis_instrumentation__.Notify(611120)
			return UnknownReturnType
		} else {
			__antithesis_instrumentation__.Notify(611121)
		}
		__antithesis_instrumentation__.Notify(611118)
		for _, arg := range args {
			__antithesis_instrumentation__.Notify(611122)
			if t := arg.ResolvedType(); t.Family() != types.UnknownFamily {
				__antithesis_instrumentation__.Notify(611123)
				return types.MakeArray(t)
			} else {
				__antithesis_instrumentation__.Notify(611124)
			}
		}
		__antithesis_instrumentation__.Notify(611119)
		return types.Unknown
	}
}

func FirstNonNullReturnType() ReturnTyper {
	__antithesis_instrumentation__.Notify(611125)
	return func(args []TypedExpr) *types.T {
		__antithesis_instrumentation__.Notify(611126)
		if len(args) == 0 {
			__antithesis_instrumentation__.Notify(611129)
			return UnknownReturnType
		} else {
			__antithesis_instrumentation__.Notify(611130)
		}
		__antithesis_instrumentation__.Notify(611127)
		for _, arg := range args {
			__antithesis_instrumentation__.Notify(611131)
			if t := arg.ResolvedType(); t.Family() != types.UnknownFamily {
				__antithesis_instrumentation__.Notify(611132)
				return t
			} else {
				__antithesis_instrumentation__.Notify(611133)
			}
		}
		__antithesis_instrumentation__.Notify(611128)
		return types.Unknown
	}
}

func returnTypeToFixedType(s ReturnTyper, inputTyps []TypedExpr) *types.T {
	__antithesis_instrumentation__.Notify(611134)
	if t := s(inputTyps); t != UnknownReturnType {
		__antithesis_instrumentation__.Notify(611136)
		return t
	} else {
		__antithesis_instrumentation__.Notify(611137)
	}
	__antithesis_instrumentation__.Notify(611135)
	return types.Any
}

type typeCheckOverloadState struct {
	overloads       []overloadImpl
	overloadIdxs    []uint8
	exprs           []Expr
	typedExprs      []TypedExpr
	resolvableIdxs  []int
	constIdxs       []int
	placeholderIdxs []int
}

func typeCheckOverloadedExprs(
	ctx context.Context,
	semaCtx *SemaContext,
	desired *types.T,
	overloads []overloadImpl,
	inBinOp bool,
	exprs ...Expr,
) ([]TypedExpr, []overloadImpl, error) {
	__antithesis_instrumentation__.Notify(611138)
	if len(overloads) > math.MaxUint8 {
		__antithesis_instrumentation__.Notify(611157)
		return nil, nil, errors.AssertionFailedf("too many overloads (%d > 255)", len(overloads))
	} else {
		__antithesis_instrumentation__.Notify(611158)
	}
	__antithesis_instrumentation__.Notify(611139)

	var s typeCheckOverloadState
	s.exprs = exprs
	s.overloads = overloads

	for i, overload := range overloads {
		__antithesis_instrumentation__.Notify(611159)

		if _, ok := overload.params().(HomogeneousType); ok {
			__antithesis_instrumentation__.Notify(611160)
			if len(overloads) > 1 {
				__antithesis_instrumentation__.Notify(611163)
				return nil, nil, errors.AssertionFailedf(
					"only one overload can have HomogeneousType parameters")
			} else {
				__antithesis_instrumentation__.Notify(611164)
			}
			__antithesis_instrumentation__.Notify(611161)
			typedExprs, _, err := TypeCheckSameTypedExprs(ctx, semaCtx, desired, exprs...)
			if err != nil {
				__antithesis_instrumentation__.Notify(611165)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(611166)
			}
			__antithesis_instrumentation__.Notify(611162)
			return typedExprs, overloads[i : i+1], nil
		} else {
			__antithesis_instrumentation__.Notify(611167)
		}
	}
	__antithesis_instrumentation__.Notify(611140)

	s.typedExprs = make([]TypedExpr, len(exprs))
	s.constIdxs, s.placeholderIdxs, s.resolvableIdxs = typeCheckSplitExprs(ctx, semaCtx, exprs)

	if len(overloads) == 0 {
		__antithesis_instrumentation__.Notify(611168)
		for _, i := range s.resolvableIdxs {
			__antithesis_instrumentation__.Notify(611171)
			typ, err := exprs[i].TypeCheck(ctx, semaCtx, types.Any)
			if err != nil {
				__antithesis_instrumentation__.Notify(611173)
				return nil, nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue,
					"error type checking resolved expression:")
			} else {
				__antithesis_instrumentation__.Notify(611174)
			}
			__antithesis_instrumentation__.Notify(611172)
			s.typedExprs[i] = typ
		}
		__antithesis_instrumentation__.Notify(611169)
		if err := defaultTypeCheck(ctx, semaCtx, &s, false); err != nil {
			__antithesis_instrumentation__.Notify(611175)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(611176)
		}
		__antithesis_instrumentation__.Notify(611170)
		return s.typedExprs, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(611177)
	}
	__antithesis_instrumentation__.Notify(611141)

	s.overloadIdxs = make([]uint8, len(overloads))
	for i := 0; i < len(overloads); i++ {
		__antithesis_instrumentation__.Notify(611178)
		s.overloadIdxs[i] = uint8(i)
	}
	__antithesis_instrumentation__.Notify(611142)

	s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
		func(o overloadImpl) bool {
			__antithesis_instrumentation__.Notify(611179)
			return o.params().MatchLen(len(exprs))
		})
	__antithesis_instrumentation__.Notify(611143)

	for _, i := range s.constIdxs {
		__antithesis_instrumentation__.Notify(611180)
		constExpr := exprs[i].(Constant)
		s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
			func(o overloadImpl) bool {
				__antithesis_instrumentation__.Notify(611181)
				return canConstantBecome(constExpr, o.params().GetAt(i))
			})
	}
	__antithesis_instrumentation__.Notify(611144)

	for _, i := range s.resolvableIdxs {
		__antithesis_instrumentation__.Notify(611182)
		paramDesired := types.Any

		var sameType *types.T
		for _, ovIdx := range s.overloadIdxs {
			__antithesis_instrumentation__.Notify(611186)
			typ := s.overloads[ovIdx].params().GetAt(i)
			if sameType == nil {
				__antithesis_instrumentation__.Notify(611187)
				sameType = typ
			} else {
				__antithesis_instrumentation__.Notify(611188)
				if !typ.Identical(sameType) {
					__antithesis_instrumentation__.Notify(611189)
					sameType = nil
					break
				} else {
					__antithesis_instrumentation__.Notify(611190)
				}
			}
		}
		__antithesis_instrumentation__.Notify(611183)
		if sameType != nil {
			__antithesis_instrumentation__.Notify(611191)
			paramDesired = sameType
		} else {
			__antithesis_instrumentation__.Notify(611192)
		}
		__antithesis_instrumentation__.Notify(611184)
		typ, err := exprs[i].TypeCheck(ctx, semaCtx, paramDesired)
		if err != nil {
			__antithesis_instrumentation__.Notify(611193)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(611194)
		}
		__antithesis_instrumentation__.Notify(611185)
		s.typedExprs[i] = typ
		s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
			func(o overloadImpl) bool {
				__antithesis_instrumentation__.Notify(611195)
				return o.params().MatchAt(typ.ResolvedType(), i)
			})
	}
	__antithesis_instrumentation__.Notify(611145)

	if ok, typedExprs, fns, err := checkReturn(ctx, semaCtx, &s); ok {
		__antithesis_instrumentation__.Notify(611196)
		return typedExprs, fns, err
	} else {
		__antithesis_instrumentation__.Notify(611197)
	}
	__antithesis_instrumentation__.Notify(611146)

	if desired.Family() != types.AnyFamily {
		__antithesis_instrumentation__.Notify(611198)
		s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
			func(o overloadImpl) bool {
				__antithesis_instrumentation__.Notify(611200)

				if t := o.returnType()(nil); t != UnknownReturnType {
					__antithesis_instrumentation__.Notify(611202)
					return t.Equivalent(desired)
				} else {
					__antithesis_instrumentation__.Notify(611203)
				}
				__antithesis_instrumentation__.Notify(611201)
				return true
			})
		__antithesis_instrumentation__.Notify(611199)
		if ok, typedExprs, fns, err := checkReturn(ctx, semaCtx, &s); ok {
			__antithesis_instrumentation__.Notify(611204)
			return typedExprs, fns, err
		} else {
			__antithesis_instrumentation__.Notify(611205)
		}
	} else {
		__antithesis_instrumentation__.Notify(611206)
	}
	__antithesis_instrumentation__.Notify(611147)

	var homogeneousTyp *types.T
	if len(s.resolvableIdxs) > 0 {
		__antithesis_instrumentation__.Notify(611207)
		homogeneousTyp = s.typedExprs[s.resolvableIdxs[0]].ResolvedType()
		for _, i := range s.resolvableIdxs[1:] {
			__antithesis_instrumentation__.Notify(611208)
			if !homogeneousTyp.Equivalent(s.typedExprs[i].ResolvedType()) {
				__antithesis_instrumentation__.Notify(611209)
				homogeneousTyp = nil
				break
			} else {
				__antithesis_instrumentation__.Notify(611210)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(611211)
	}
	__antithesis_instrumentation__.Notify(611148)

	if len(s.constIdxs) > 0 {
		__antithesis_instrumentation__.Notify(611212)
		allConstantsAreHomogenous := false
		if ok, typedExprs, fns, err := filterAttempt(ctx, semaCtx, &s, func() {
			__antithesis_instrumentation__.Notify(611218)

			if homogeneousTyp != nil {
				__antithesis_instrumentation__.Notify(611219)
				allConstantsAreHomogenous = true
				for _, i := range s.constIdxs {
					__antithesis_instrumentation__.Notify(611221)
					if !canConstantBecome(exprs[i].(Constant), homogeneousTyp) {
						__antithesis_instrumentation__.Notify(611222)
						allConstantsAreHomogenous = false
						break
					} else {
						__antithesis_instrumentation__.Notify(611223)
					}
				}
				__antithesis_instrumentation__.Notify(611220)
				if allConstantsAreHomogenous {
					__antithesis_instrumentation__.Notify(611224)
					for _, i := range s.constIdxs {
						__antithesis_instrumentation__.Notify(611225)
						s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
							func(o overloadImpl) bool {
								__antithesis_instrumentation__.Notify(611226)
								return o.params().GetAt(i).Equivalent(homogeneousTyp)
							})
					}
				} else {
					__antithesis_instrumentation__.Notify(611227)
				}
			} else {
				__antithesis_instrumentation__.Notify(611228)
			}
		}); ok {
			__antithesis_instrumentation__.Notify(611229)
			return typedExprs, fns, err
		} else {
			__antithesis_instrumentation__.Notify(611230)
		}
		__antithesis_instrumentation__.Notify(611213)

		if ok, typedExprs, fns, err := filterAttempt(ctx, semaCtx, &s, func() {
			__antithesis_instrumentation__.Notify(611231)

			for _, i := range s.constIdxs {
				__antithesis_instrumentation__.Notify(611232)
				natural := naturalConstantType(exprs[i].(Constant))
				if natural != nil {
					__antithesis_instrumentation__.Notify(611233)
					s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
						func(o overloadImpl) bool {
							__antithesis_instrumentation__.Notify(611234)
							return o.params().GetAt(i).Equivalent(natural)
						})
				} else {
					__antithesis_instrumentation__.Notify(611235)
				}
			}
		}); ok {
			__antithesis_instrumentation__.Notify(611236)
			return typedExprs, fns, err
		} else {
			__antithesis_instrumentation__.Notify(611237)
		}
		__antithesis_instrumentation__.Notify(611214)

		if len(s.overloadIdxs) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(611238)
			return allConstantsAreHomogenous == true
		}() == true {
			__antithesis_instrumentation__.Notify(611239)
			overloadParamsAreHomogenous := true
			p := s.overloads[s.overloadIdxs[0]].params()
			for _, i := range s.constIdxs {
				__antithesis_instrumentation__.Notify(611241)
				if !p.GetAt(i).Equivalent(homogeneousTyp) {
					__antithesis_instrumentation__.Notify(611242)
					overloadParamsAreHomogenous = false
					break
				} else {
					__antithesis_instrumentation__.Notify(611243)
				}
			}
			__antithesis_instrumentation__.Notify(611240)
			if overloadParamsAreHomogenous {
				__antithesis_instrumentation__.Notify(611244)

				for _, i := range s.constIdxs {
					__antithesis_instrumentation__.Notify(611246)
					typ, err := s.exprs[i].TypeCheck(ctx, semaCtx, homogeneousTyp)
					if err != nil {
						__antithesis_instrumentation__.Notify(611248)
						return nil, nil, err
					} else {
						__antithesis_instrumentation__.Notify(611249)
					}
					__antithesis_instrumentation__.Notify(611247)
					s.typedExprs[i] = typ
				}
				__antithesis_instrumentation__.Notify(611245)
				_, typedExprs, fn, err := checkReturnPlaceholdersAtIdx(ctx, semaCtx, &s, int(s.overloadIdxs[0]))
				return typedExprs, fn, err
			} else {
				__antithesis_instrumentation__.Notify(611250)
			}
		} else {
			__antithesis_instrumentation__.Notify(611251)
		}
		__antithesis_instrumentation__.Notify(611215)
		for _, i := range s.constIdxs {
			__antithesis_instrumentation__.Notify(611252)
			constExpr := exprs[i].(Constant)
			s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
				func(o overloadImpl) bool {
					__antithesis_instrumentation__.Notify(611253)
					semaCtx := MakeSemaContext()
					_, err := constExpr.ResolveAsType(ctx, &semaCtx, o.params().GetAt(i))
					return err == nil
				})
		}
		__antithesis_instrumentation__.Notify(611216)
		if ok, typedExprs, fn, err := checkReturn(ctx, semaCtx, &s); ok {
			__antithesis_instrumentation__.Notify(611254)
			return typedExprs, fn, err
		} else {
			__antithesis_instrumentation__.Notify(611255)
		}
		__antithesis_instrumentation__.Notify(611217)

		if bestConstType, ok := commonConstantType(s.exprs, s.constIdxs); ok {
			__antithesis_instrumentation__.Notify(611256)

			prevOverloadIdxs := s.overloadIdxs
			for _, i := range s.constIdxs {
				__antithesis_instrumentation__.Notify(611259)
				s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
					func(o overloadImpl) bool {
						__antithesis_instrumentation__.Notify(611260)
						return o.params().GetAt(i).Equivalent(bestConstType)
					})
			}
			__antithesis_instrumentation__.Notify(611257)
			if ok, typedExprs, fns, err := checkReturn(ctx, semaCtx, &s); ok {
				__antithesis_instrumentation__.Notify(611261)
				if len(fns) == 0 {
					__antithesis_instrumentation__.Notify(611263)
					var overloadImpls []overloadImpl
					for i := range prevOverloadIdxs {
						__antithesis_instrumentation__.Notify(611265)
						overloadImpls = append(overloadImpls, s.overloads[i])
					}
					__antithesis_instrumentation__.Notify(611264)
					return typedExprs, overloadImpls, err
				} else {
					__antithesis_instrumentation__.Notify(611266)
				}
				__antithesis_instrumentation__.Notify(611262)
				return typedExprs, fns, err
			} else {
				__antithesis_instrumentation__.Notify(611267)
			}
			__antithesis_instrumentation__.Notify(611258)
			if homogeneousTyp != nil {
				__antithesis_instrumentation__.Notify(611268)
				if !homogeneousTyp.Equivalent(bestConstType) {
					__antithesis_instrumentation__.Notify(611269)
					homogeneousTyp = nil
				} else {
					__antithesis_instrumentation__.Notify(611270)
				}
			} else {
				__antithesis_instrumentation__.Notify(611271)
				homogeneousTyp = bestConstType
			}
		} else {
			__antithesis_instrumentation__.Notify(611272)
		}
	} else {
		__antithesis_instrumentation__.Notify(611273)
	}
	__antithesis_instrumentation__.Notify(611149)

	if ok, typedExprs, fns, err := filterAttempt(ctx, semaCtx, &s, func() {
		__antithesis_instrumentation__.Notify(611274)
		s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs, func(o overloadImpl) bool {
			__antithesis_instrumentation__.Notify(611275)
			return o.preferred()
		})
	}); ok {
		__antithesis_instrumentation__.Notify(611276)
		return typedExprs, fns, err
	} else {
		__antithesis_instrumentation__.Notify(611277)
	}
	__antithesis_instrumentation__.Notify(611150)

	if homogeneousTyp != nil && func() bool {
		__antithesis_instrumentation__.Notify(611278)
		return len(s.placeholderIdxs) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(611279)

		for _, i := range s.placeholderIdxs {
			__antithesis_instrumentation__.Notify(611281)
			if _, err := exprs[i].TypeCheck(ctx, semaCtx, homogeneousTyp); err != nil {
				__antithesis_instrumentation__.Notify(611283)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(611284)
			}
			__antithesis_instrumentation__.Notify(611282)
			s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
				func(o overloadImpl) bool {
					__antithesis_instrumentation__.Notify(611285)
					return o.params().GetAt(i).Equivalent(homogeneousTyp)
				})
		}
		__antithesis_instrumentation__.Notify(611280)
		if ok, typedExprs, fns, err := checkReturn(ctx, semaCtx, &s); ok {
			__antithesis_instrumentation__.Notify(611286)
			return typedExprs, fns, err
		} else {
			__antithesis_instrumentation__.Notify(611287)
		}
	} else {
		__antithesis_instrumentation__.Notify(611288)
	}
	__antithesis_instrumentation__.Notify(611151)

	if len(s.overloadIdxs) == 1 {
		__antithesis_instrumentation__.Notify(611289)
		params := s.overloads[s.overloadIdxs[0]].params()
		var knownEnum *types.T

		attemptAnyEnumCast := func() bool {
			__antithesis_instrumentation__.Notify(611291)
			for i := 0; i < params.Length(); i++ {
				__antithesis_instrumentation__.Notify(611293)
				typ := params.GetAt(i)

				if !(typ.Identical(types.AnyEnum) || func() bool {
					__antithesis_instrumentation__.Notify(611295)
					return typ.Identical(types.MakeArray(types.AnyEnum)) == true
				}() == true) {
					__antithesis_instrumentation__.Notify(611296)
					return false
				} else {
					__antithesis_instrumentation__.Notify(611297)
				}
				__antithesis_instrumentation__.Notify(611294)
				if s.typedExprs[i] != nil {
					__antithesis_instrumentation__.Notify(611298)

					posEnum := s.typedExprs[i].ResolvedType()
					if !posEnum.UserDefined() {
						__antithesis_instrumentation__.Notify(611301)
						return false
					} else {
						__antithesis_instrumentation__.Notify(611302)
					}
					__antithesis_instrumentation__.Notify(611299)
					if posEnum.Family() == types.ArrayFamily {
						__antithesis_instrumentation__.Notify(611303)
						posEnum = posEnum.ArrayContents()
					} else {
						__antithesis_instrumentation__.Notify(611304)
					}
					__antithesis_instrumentation__.Notify(611300)
					if knownEnum == nil {
						__antithesis_instrumentation__.Notify(611305)
						knownEnum = posEnum
					} else {
						__antithesis_instrumentation__.Notify(611306)
						if !posEnum.Identical(knownEnum) {
							__antithesis_instrumentation__.Notify(611307)
							return false
						} else {
							__antithesis_instrumentation__.Notify(611308)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(611309)
				}
			}
			__antithesis_instrumentation__.Notify(611292)
			return knownEnum != nil
		}()
		__antithesis_instrumentation__.Notify(611290)

		if attemptAnyEnumCast {
			__antithesis_instrumentation__.Notify(611310)

			sCopy := s
			sCopy.exprs = make([]Expr, len(s.exprs))
			copy(sCopy.exprs, s.exprs)
			if ok, typedExprs, fns, err := filterAttempt(ctx, semaCtx, &sCopy, func() {
				__antithesis_instrumentation__.Notify(611311)
				for _, idx := range append(s.constIdxs, s.placeholderIdxs...) {
					__antithesis_instrumentation__.Notify(611312)
					p := params.GetAt(idx)
					typCast := knownEnum
					if p.Family() == types.ArrayFamily {
						__antithesis_instrumentation__.Notify(611314)
						typCast = types.MakeArray(knownEnum)
					} else {
						__antithesis_instrumentation__.Notify(611315)
					}
					__antithesis_instrumentation__.Notify(611313)
					sCopy.exprs[idx] = &CastExpr{Expr: sCopy.exprs[idx], Type: typCast, SyntaxMode: CastShort}
				}
			}); ok {
				__antithesis_instrumentation__.Notify(611316)
				return typedExprs, fns, err
			} else {
				__antithesis_instrumentation__.Notify(611317)
			}
		} else {
			__antithesis_instrumentation__.Notify(611318)
		}
	} else {
		__antithesis_instrumentation__.Notify(611319)
	}
	__antithesis_instrumentation__.Notify(611152)

	if inBinOp && func() bool {
		__antithesis_instrumentation__.Notify(611320)
		return len(s.exprs) == 2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(611321)
		if ok, typedExprs, fns, err := filterAttempt(ctx, semaCtx, &s, func() {
			__antithesis_instrumentation__.Notify(611322)
			var err error
			left := s.typedExprs[0]
			if left == nil {
				__antithesis_instrumentation__.Notify(611325)
				left, err = s.exprs[0].TypeCheck(ctx, semaCtx, types.Any)
				if err != nil {
					__antithesis_instrumentation__.Notify(611326)
					return
				} else {
					__antithesis_instrumentation__.Notify(611327)
				}
			} else {
				__antithesis_instrumentation__.Notify(611328)
			}
			__antithesis_instrumentation__.Notify(611323)
			right := s.typedExprs[1]
			if right == nil {
				__antithesis_instrumentation__.Notify(611329)
				right, err = s.exprs[1].TypeCheck(ctx, semaCtx, types.Any)
				if err != nil {
					__antithesis_instrumentation__.Notify(611330)
					return
				} else {
					__antithesis_instrumentation__.Notify(611331)
				}
			} else {
				__antithesis_instrumentation__.Notify(611332)
			}
			__antithesis_instrumentation__.Notify(611324)
			leftType := left.ResolvedType()
			rightType := right.ResolvedType()
			leftIsNull := leftType.Family() == types.UnknownFamily
			rightIsNull := rightType.Family() == types.UnknownFamily
			oneIsNull := (leftIsNull || func() bool {
				__antithesis_instrumentation__.Notify(611333)
				return rightIsNull == true
			}() == true) && func() bool {
				__antithesis_instrumentation__.Notify(611334)
				return !(leftIsNull && func() bool {
					__antithesis_instrumentation__.Notify(611335)
					return rightIsNull == true
				}() == true) == true
			}() == true
			if oneIsNull {
				__antithesis_instrumentation__.Notify(611336)
				if leftIsNull {
					__antithesis_instrumentation__.Notify(611339)
					leftType = rightType
				} else {
					__antithesis_instrumentation__.Notify(611340)
				}
				__antithesis_instrumentation__.Notify(611337)
				if rightIsNull {
					__antithesis_instrumentation__.Notify(611341)
					rightType = leftType
				} else {
					__antithesis_instrumentation__.Notify(611342)
				}
				__antithesis_instrumentation__.Notify(611338)
				s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
					func(o overloadImpl) bool {
						__antithesis_instrumentation__.Notify(611343)
						return o.params().GetAt(0).Equivalent(leftType) && func() bool {
							__antithesis_instrumentation__.Notify(611344)
							return o.params().GetAt(1).Equivalent(rightType) == true
						}() == true
					})
			} else {
				__antithesis_instrumentation__.Notify(611345)
			}
		}); ok {
			__antithesis_instrumentation__.Notify(611346)
			return typedExprs, fns, err
		} else {
			__antithesis_instrumentation__.Notify(611347)
		}
	} else {
		__antithesis_instrumentation__.Notify(611348)
	}
	__antithesis_instrumentation__.Notify(611153)

	if inBinOp && func() bool {
		__antithesis_instrumentation__.Notify(611349)
		return len(s.exprs) == 2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(611350)
		if ok, typedExprs, fns, err := filterAttempt(ctx, semaCtx, &s, func() {
			__antithesis_instrumentation__.Notify(611351)
			var err error
			left := s.typedExprs[0]
			if left == nil {
				__antithesis_instrumentation__.Notify(611354)
				left, err = s.exprs[0].TypeCheck(ctx, semaCtx, types.Any)
				if err != nil {
					__antithesis_instrumentation__.Notify(611355)
					return
				} else {
					__antithesis_instrumentation__.Notify(611356)
				}
			} else {
				__antithesis_instrumentation__.Notify(611357)
			}
			__antithesis_instrumentation__.Notify(611352)
			right := s.typedExprs[1]
			if right == nil {
				__antithesis_instrumentation__.Notify(611358)
				right, err = s.exprs[1].TypeCheck(ctx, semaCtx, types.Any)
				if err != nil {
					__antithesis_instrumentation__.Notify(611359)
					return
				} else {
					__antithesis_instrumentation__.Notify(611360)
				}
			} else {
				__antithesis_instrumentation__.Notify(611361)
			}
			__antithesis_instrumentation__.Notify(611353)
			leftType := left.ResolvedType()
			rightType := right.ResolvedType()
			leftIsNull := leftType.Family() == types.UnknownFamily
			rightIsNull := rightType.Family() == types.UnknownFamily
			oneIsNull := (leftIsNull || func() bool {
				__antithesis_instrumentation__.Notify(611362)
				return rightIsNull == true
			}() == true) && func() bool {
				__antithesis_instrumentation__.Notify(611363)
				return !(leftIsNull && func() bool {
					__antithesis_instrumentation__.Notify(611364)
					return rightIsNull == true
				}() == true) == true
			}() == true
			if oneIsNull {
				__antithesis_instrumentation__.Notify(611365)
				if leftIsNull {
					__antithesis_instrumentation__.Notify(611368)
					leftType = types.String
				} else {
					__antithesis_instrumentation__.Notify(611369)
				}
				__antithesis_instrumentation__.Notify(611366)
				if rightIsNull {
					__antithesis_instrumentation__.Notify(611370)
					rightType = types.String
				} else {
					__antithesis_instrumentation__.Notify(611371)
				}
				__antithesis_instrumentation__.Notify(611367)
				s.overloadIdxs = filterOverloads(s.overloads, s.overloadIdxs,
					func(o overloadImpl) bool {
						__antithesis_instrumentation__.Notify(611372)
						return o.params().GetAt(0).Equivalent(leftType) && func() bool {
							__antithesis_instrumentation__.Notify(611373)
							return o.params().GetAt(1).Equivalent(rightType) == true
						}() == true
					})
			} else {
				__antithesis_instrumentation__.Notify(611374)
			}
		}); ok {
			__antithesis_instrumentation__.Notify(611375)
			return typedExprs, fns, err
		} else {
			__antithesis_instrumentation__.Notify(611376)
		}
	} else {
		__antithesis_instrumentation__.Notify(611377)
	}
	__antithesis_instrumentation__.Notify(611154)

	if err := defaultTypeCheck(ctx, semaCtx, &s, len(s.overloads) > 0); err != nil {
		__antithesis_instrumentation__.Notify(611378)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(611379)
	}
	__antithesis_instrumentation__.Notify(611155)

	possibleOverloads := make([]overloadImpl, len(s.overloadIdxs))
	for i, o := range s.overloadIdxs {
		__antithesis_instrumentation__.Notify(611380)
		possibleOverloads[i] = s.overloads[o]
	}
	__antithesis_instrumentation__.Notify(611156)
	return s.typedExprs, possibleOverloads, nil
}

func filterAttempt(
	ctx context.Context, semaCtx *SemaContext, s *typeCheckOverloadState, attempt func(),
) (ok bool, _ []TypedExpr, _ []overloadImpl, _ error) {
	__antithesis_instrumentation__.Notify(611381)
	before := s.overloadIdxs
	attempt()
	if len(s.overloadIdxs) == 1 {
		__antithesis_instrumentation__.Notify(611383)
		ok, typedExprs, fns, err := checkReturn(ctx, semaCtx, s)
		if err != nil {
			__antithesis_instrumentation__.Notify(611385)
			return false, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(611386)
		}
		__antithesis_instrumentation__.Notify(611384)
		if ok {
			__antithesis_instrumentation__.Notify(611387)
			return true, typedExprs, fns, err
		} else {
			__antithesis_instrumentation__.Notify(611388)
		}
	} else {
		__antithesis_instrumentation__.Notify(611389)
	}
	__antithesis_instrumentation__.Notify(611382)
	s.overloadIdxs = before
	return false, nil, nil, nil
}

func filterOverloads(
	overloads []overloadImpl, overloadIdxs []uint8, fn func(overloadImpl) bool,
) []uint8 {
	__antithesis_instrumentation__.Notify(611390)
	for i := 0; i < len(overloadIdxs); {
		__antithesis_instrumentation__.Notify(611392)
		if fn(overloads[overloadIdxs[i]]) {
			__antithesis_instrumentation__.Notify(611393)
			i++
		} else {
			__antithesis_instrumentation__.Notify(611394)
			overloadIdxs[i], overloadIdxs[len(overloadIdxs)-1] = overloadIdxs[len(overloadIdxs)-1], overloadIdxs[i]
			overloadIdxs = overloadIdxs[:len(overloadIdxs)-1]
		}
	}
	__antithesis_instrumentation__.Notify(611391)
	return overloadIdxs
}

func defaultTypeCheck(
	ctx context.Context, semaCtx *SemaContext, s *typeCheckOverloadState, errorOnPlaceholders bool,
) error {
	__antithesis_instrumentation__.Notify(611395)
	for _, i := range s.constIdxs {
		__antithesis_instrumentation__.Notify(611398)
		typ, err := s.exprs[i].TypeCheck(ctx, semaCtx, types.Any)
		if err != nil {
			__antithesis_instrumentation__.Notify(611400)
			return pgerror.Wrapf(err, pgcode.InvalidParameterValue,
				"error type checking constant value")
		} else {
			__antithesis_instrumentation__.Notify(611401)
		}
		__antithesis_instrumentation__.Notify(611399)
		s.typedExprs[i] = typ
	}
	__antithesis_instrumentation__.Notify(611396)
	for _, i := range s.placeholderIdxs {
		__antithesis_instrumentation__.Notify(611402)
		if errorOnPlaceholders {
			__antithesis_instrumentation__.Notify(611404)
			_, err := s.exprs[i].TypeCheck(ctx, semaCtx, types.Any)
			return err
		} else {
			__antithesis_instrumentation__.Notify(611405)
		}
		__antithesis_instrumentation__.Notify(611403)

		s.typedExprs[i] = StripParens(s.exprs[i]).(*Placeholder)
	}
	__antithesis_instrumentation__.Notify(611397)
	return nil
}

func checkReturn(
	ctx context.Context, semaCtx *SemaContext, s *typeCheckOverloadState,
) (ok bool, _ []TypedExpr, _ []overloadImpl, _ error) {
	__antithesis_instrumentation__.Notify(611406)
	switch len(s.overloadIdxs) {
	case 0:
		__antithesis_instrumentation__.Notify(611407)
		if err := defaultTypeCheck(ctx, semaCtx, s, false); err != nil {
			__antithesis_instrumentation__.Notify(611412)
			return false, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(611413)
		}
		__antithesis_instrumentation__.Notify(611408)
		return true, s.typedExprs, nil, nil

	case 1:
		__antithesis_instrumentation__.Notify(611409)
		idx := s.overloadIdxs[0]
		o := s.overloads[idx]
		p := o.params()
		for _, i := range s.constIdxs {
			__antithesis_instrumentation__.Notify(611414)
			des := p.GetAt(i)
			typ, err := s.exprs[i].TypeCheck(ctx, semaCtx, des)
			if err != nil {
				__antithesis_instrumentation__.Notify(611417)
				return false, s.typedExprs, nil, pgerror.Wrapf(
					err, pgcode.InvalidParameterValue,
					"error type checking constant value",
				)
			} else {
				__antithesis_instrumentation__.Notify(611418)
			}
			__antithesis_instrumentation__.Notify(611415)
			if des != nil && func() bool {
				__antithesis_instrumentation__.Notify(611419)
				return !typ.ResolvedType().Equivalent(des) == true
			}() == true {
				__antithesis_instrumentation__.Notify(611420)
				return false, nil, nil, errors.AssertionFailedf(
					"desired constant value type %s but set type %s",
					redact.Safe(des), redact.Safe(typ.ResolvedType()),
				)
			} else {
				__antithesis_instrumentation__.Notify(611421)
			}
			__antithesis_instrumentation__.Notify(611416)
			s.typedExprs[i] = typ
		}
		__antithesis_instrumentation__.Notify(611410)

		return checkReturnPlaceholdersAtIdx(ctx, semaCtx, s, int(idx))

	default:
		__antithesis_instrumentation__.Notify(611411)
		return false, nil, nil, nil
	}
}

func checkReturnPlaceholdersAtIdx(
	ctx context.Context, semaCtx *SemaContext, s *typeCheckOverloadState, idx int,
) (bool, []TypedExpr, []overloadImpl, error) {
	__antithesis_instrumentation__.Notify(611422)
	o := s.overloads[idx]
	p := o.params()
	for _, i := range s.placeholderIdxs {
		__antithesis_instrumentation__.Notify(611424)
		des := p.GetAt(i)
		typ, err := s.exprs[i].TypeCheck(ctx, semaCtx, des)
		if err != nil {
			__antithesis_instrumentation__.Notify(611426)
			if des.IsAmbiguous() {
				__antithesis_instrumentation__.Notify(611428)
				return false, nil, nil, nil
			} else {
				__antithesis_instrumentation__.Notify(611429)
			}
			__antithesis_instrumentation__.Notify(611427)
			return false, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(611430)
		}
		__antithesis_instrumentation__.Notify(611425)
		s.typedExprs[i] = typ
	}
	__antithesis_instrumentation__.Notify(611423)
	return true, s.typedExprs, s.overloads[idx : idx+1], nil
}

func formatCandidates(prefix string, candidates []overloadImpl) string {
	__antithesis_instrumentation__.Notify(611431)
	var buf bytes.Buffer
	for _, candidate := range candidates {
		__antithesis_instrumentation__.Notify(611433)
		buf.WriteString(prefix)
		buf.WriteByte('(')
		params := candidate.params()
		tLen := params.Length()
		inputTyps := make([]TypedExpr, tLen)
		for i := 0; i < tLen; i++ {
			__antithesis_instrumentation__.Notify(611436)
			t := params.GetAt(i)
			inputTyps[i] = &TypedDummy{Typ: t}
			if i > 0 {
				__antithesis_instrumentation__.Notify(611438)
				buf.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(611439)
			}
			__antithesis_instrumentation__.Notify(611437)
			buf.WriteString(t.String())
		}
		__antithesis_instrumentation__.Notify(611434)
		buf.WriteString(") -> ")
		buf.WriteString(returnTypeToFixedType(candidate.returnType(), inputTyps).String())
		if candidate.preferred() {
			__antithesis_instrumentation__.Notify(611440)
			buf.WriteString(" [preferred]")
		} else {
			__antithesis_instrumentation__.Notify(611441)
		}
		__antithesis_instrumentation__.Notify(611435)
		buf.WriteByte('\n')
	}
	__antithesis_instrumentation__.Notify(611432)
	return buf.String()
}
