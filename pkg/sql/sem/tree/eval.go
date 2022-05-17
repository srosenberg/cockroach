package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/arith"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

var (
	ErrIntOutOfRange = pgerror.New(pgcode.NumericValueOutOfRange, "integer out of range")

	ErrInt4OutOfRange = pgerror.New(pgcode.NumericValueOutOfRange, "integer out of range for type int4")

	ErrInt2OutOfRange = pgerror.New(pgcode.NumericValueOutOfRange, "integer out of range for type int2")

	ErrFloatOutOfRange = pgerror.New(pgcode.NumericValueOutOfRange, "float out of range")

	ErrDecOutOfRange = pgerror.New(pgcode.NumericValueOutOfRange, "decimal out of range")

	errCharOutOfRange = pgerror.New(pgcode.NumericValueOutOfRange, "\"char\" out of range")

	ErrDivByZero       = pgerror.New(pgcode.DivisionByZero, "division by zero")
	errSqrtOfNegNumber = pgerror.New(pgcode.InvalidArgumentForPowerFunction, "cannot take square root of a negative number")

	ErrShiftArgOutOfRange = pgerror.New(pgcode.InvalidParameterValue, "shift argument out of range")

	big10E6  = apd.NewBigInt(1e6)
	big10E10 = apd.NewBigInt(1e10)
)

func NewCannotMixBitArraySizesError(op string) error {
	__antithesis_instrumentation__.Notify(607688)
	return pgerror.Newf(pgcode.StringDataLengthMismatch,
		"cannot %s bit strings of different sizes", op)
}

type UnaryOp struct {
	Typ        *types.T
	ReturnType *types.T
	Fn         func(*EvalContext, Datum) (Datum, error)
	Volatility Volatility

	types   TypeList
	retType ReturnTyper

	counter telemetry.Counter
}

func (op *UnaryOp) params() TypeList {
	__antithesis_instrumentation__.Notify(607689)
	return op.types
}

func (op *UnaryOp) returnType() ReturnTyper {
	__antithesis_instrumentation__.Notify(607690)
	return op.retType
}

func (*UnaryOp) preferred() bool {
	__antithesis_instrumentation__.Notify(607691)
	return false
}

func unaryOpFixups(
	ops map[UnaryOperatorSymbol]unaryOpOverload,
) map[UnaryOperatorSymbol]unaryOpOverload {
	__antithesis_instrumentation__.Notify(607692)
	for op, overload := range ops {
		__antithesis_instrumentation__.Notify(607694)
		for i, impl := range overload {
			__antithesis_instrumentation__.Notify(607695)
			casted := impl.(*UnaryOp)
			casted.types = ArgTypes{{"arg", casted.Typ}}
			casted.retType = FixedReturnType(casted.ReturnType)
			ops[op][i] = casted
		}
	}
	__antithesis_instrumentation__.Notify(607693)
	return ops
}

type unaryOpOverload []overloadImpl

var UnaryOps = unaryOpFixups(map[UnaryOperatorSymbol]unaryOpOverload{
	UnaryPlus: {
		&UnaryOp{
			Typ:        types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607696)
				return d, nil
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607697)
				return d, nil
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607698)
				return d, nil
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Interval,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607699)
				return d, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	UnaryMinus: {
		&UnaryOp{
			Typ:        types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607700)
				i := MustBeDInt(d)
				if i == math.MinInt64 {
					__antithesis_instrumentation__.Notify(607702)
					return nil, ErrIntOutOfRange
				} else {
					__antithesis_instrumentation__.Notify(607703)
				}
				__antithesis_instrumentation__.Notify(607701)
				return NewDInt(-i), nil
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607704)
				return NewDFloat(-*d.(*DFloat)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607705)
				dec := &d.(*DDecimal).Decimal
				dd := &DDecimal{}
				dd.Decimal.Neg(dec)
				return dd, nil
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Interval,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607706)
				i := d.(*DInterval).Duration
				i.SetNanos(-i.Nanos())
				i.Days = -i.Days
				i.Months = -i.Months
				return &DInterval{Duration: i}, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	UnaryComplement: {
		&UnaryOp{
			Typ:        types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607707)
				return NewDInt(^MustBeDInt(d)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.VarBit,
			ReturnType: types.VarBit,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607708)
				p := MustBeDBitArray(d)
				return &DBitArray{BitArray: bitarray.Not(p.BitArray)}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.INet,
			ReturnType: types.INet,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607709)
				ipAddr := MustBeDIPAddr(d).IPAddr
				return NewDIPAddr(DIPAddr{ipAddr.Complement()}), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	UnarySqrt: {
		&UnaryOp{
			Typ:        types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607710)
				return Sqrt(float64(*d.(*DFloat)))
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607711)
				dec := &d.(*DDecimal).Decimal
				return DecimalSqrt(dec)
			},
			Volatility: VolatilityImmutable,
		},
	},

	UnaryCbrt: {
		&UnaryOp{
			Typ:        types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607712)
				return Cbrt(float64(*d.(*DFloat)))
			},
			Volatility: VolatilityImmutable,
		},
		&UnaryOp{
			Typ:        types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, d Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607713)
				dec := &d.(*DDecimal).Decimal
				return DecimalCbrt(dec)
			},
			Volatility: VolatilityImmutable,
		},
	},
})

type TwoArgFn func(*EvalContext, Datum, Datum) (Datum, error)

type BinOp struct {
	LeftType          *types.T
	RightType         *types.T
	ReturnType        *types.T
	NullableArgs      bool
	Fn                TwoArgFn
	Volatility        Volatility
	PreferredOverload bool

	types   TypeList
	retType ReturnTyper

	counter telemetry.Counter
}

func (op *BinOp) params() TypeList {
	__antithesis_instrumentation__.Notify(607714)
	return op.types
}

func (op *BinOp) matchParams(l, r *types.T) bool {
	__antithesis_instrumentation__.Notify(607715)
	return op.params().MatchAt(l, 0) && func() bool {
		__antithesis_instrumentation__.Notify(607716)
		return op.params().MatchAt(r, 1) == true
	}() == true
}

func (op *BinOp) returnType() ReturnTyper {
	__antithesis_instrumentation__.Notify(607717)
	return op.retType
}

func (op *BinOp) preferred() bool {
	__antithesis_instrumentation__.Notify(607718)
	return op.PreferredOverload
}

func AppendToMaybeNullArray(typ *types.T, left Datum, right Datum) (Datum, error) {
	__antithesis_instrumentation__.Notify(607719)
	result := NewDArray(typ)
	if left != DNull {
		__antithesis_instrumentation__.Notify(607722)
		for _, e := range MustBeDArray(left).Array {
			__antithesis_instrumentation__.Notify(607723)
			if err := result.Append(e); err != nil {
				__antithesis_instrumentation__.Notify(607724)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(607725)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(607726)
	}
	__antithesis_instrumentation__.Notify(607720)
	if err := result.Append(right); err != nil {
		__antithesis_instrumentation__.Notify(607727)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(607728)
	}
	__antithesis_instrumentation__.Notify(607721)
	return result, nil
}

func PrependToMaybeNullArray(typ *types.T, left Datum, right Datum) (Datum, error) {
	__antithesis_instrumentation__.Notify(607729)
	result := NewDArray(typ)
	if err := result.Append(left); err != nil {
		__antithesis_instrumentation__.Notify(607732)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(607733)
	}
	__antithesis_instrumentation__.Notify(607730)
	if right != DNull {
		__antithesis_instrumentation__.Notify(607734)
		for _, e := range MustBeDArray(right).Array {
			__antithesis_instrumentation__.Notify(607735)
			if err := result.Append(e); err != nil {
				__antithesis_instrumentation__.Notify(607736)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(607737)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(607738)
	}
	__antithesis_instrumentation__.Notify(607731)
	return result, nil
}

func initArrayElementConcatenation() {
	__antithesis_instrumentation__.Notify(607739)
	for _, t := range types.Scalar {
		__antithesis_instrumentation__.Notify(607740)
		typ := t
		BinOps[treebin.Concat] = append(BinOps[treebin.Concat], &BinOp{
			LeftType:     types.MakeArray(typ),
			RightType:    typ,
			ReturnType:   types.MakeArray(typ),
			NullableArgs: true,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607742)
				return AppendToMaybeNullArray(typ, left, right)
			},
			Volatility: VolatilityImmutable,
		})
		__antithesis_instrumentation__.Notify(607741)

		BinOps[treebin.Concat] = append(BinOps[treebin.Concat], &BinOp{
			LeftType:     typ,
			RightType:    types.MakeArray(typ),
			ReturnType:   types.MakeArray(typ),
			NullableArgs: true,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607743)
				return PrependToMaybeNullArray(typ, left, right)
			},
			Volatility: VolatilityImmutable,
		})
	}
}

func ConcatArrays(typ *types.T, left Datum, right Datum) (Datum, error) {
	__antithesis_instrumentation__.Notify(607744)
	if left == DNull && func() bool {
		__antithesis_instrumentation__.Notify(607748)
		return right == DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(607749)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(607750)
	}
	__antithesis_instrumentation__.Notify(607745)
	result := NewDArray(typ)
	if left != DNull {
		__antithesis_instrumentation__.Notify(607751)
		for _, e := range MustBeDArray(left).Array {
			__antithesis_instrumentation__.Notify(607752)
			if err := result.Append(e); err != nil {
				__antithesis_instrumentation__.Notify(607753)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(607754)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(607755)
	}
	__antithesis_instrumentation__.Notify(607746)
	if right != DNull {
		__antithesis_instrumentation__.Notify(607756)
		for _, e := range MustBeDArray(right).Array {
			__antithesis_instrumentation__.Notify(607757)
			if err := result.Append(e); err != nil {
				__antithesis_instrumentation__.Notify(607758)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(607759)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(607760)
	}
	__antithesis_instrumentation__.Notify(607747)
	return result, nil
}

func ArrayContains(ctx *EvalContext, haystack *DArray, needles *DArray) (*DBool, error) {
	__antithesis_instrumentation__.Notify(607761)
	if !haystack.ParamTyp.Equivalent(needles.ParamTyp) {
		__antithesis_instrumentation__.Notify(607764)
		return DBoolFalse, pgerror.New(pgcode.DatatypeMismatch, "cannot compare arrays with different element types")
	} else {
		__antithesis_instrumentation__.Notify(607765)
	}
	__antithesis_instrumentation__.Notify(607762)
	for _, needle := range needles.Array {
		__antithesis_instrumentation__.Notify(607766)

		if needle == DNull {
			__antithesis_instrumentation__.Notify(607769)
			return DBoolFalse, nil
		} else {
			__antithesis_instrumentation__.Notify(607770)
		}
		__antithesis_instrumentation__.Notify(607767)
		var found bool
		for _, hay := range haystack.Array {
			__antithesis_instrumentation__.Notify(607771)
			if needle.Compare(ctx, hay) == 0 {
				__antithesis_instrumentation__.Notify(607772)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(607773)
			}
		}
		__antithesis_instrumentation__.Notify(607768)
		if !found {
			__antithesis_instrumentation__.Notify(607774)
			return DBoolFalse, nil
		} else {
			__antithesis_instrumentation__.Notify(607775)
		}
	}
	__antithesis_instrumentation__.Notify(607763)
	return DBoolTrue, nil
}

func JSONExistsAny(_ *EvalContext, json DJSON, dArray *DArray) (*DBool, error) {
	__antithesis_instrumentation__.Notify(607776)

	for _, k := range dArray.Array {
		__antithesis_instrumentation__.Notify(607778)
		if k == DNull {
			__antithesis_instrumentation__.Notify(607781)
			continue
		} else {
			__antithesis_instrumentation__.Notify(607782)
		}
		__antithesis_instrumentation__.Notify(607779)
		e, err := json.JSON.Exists(string(MustBeDString(k)))
		if err != nil {
			__antithesis_instrumentation__.Notify(607783)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(607784)
		}
		__antithesis_instrumentation__.Notify(607780)
		if e {
			__antithesis_instrumentation__.Notify(607785)
			return DBoolTrue, nil
		} else {
			__antithesis_instrumentation__.Notify(607786)
		}
	}
	__antithesis_instrumentation__.Notify(607777)
	return DBoolFalse, nil
}

func initArrayToArrayConcatenation() {
	__antithesis_instrumentation__.Notify(607787)
	for _, t := range types.Scalar {
		__antithesis_instrumentation__.Notify(607788)
		typ := t
		BinOps[treebin.Concat] = append(BinOps[treebin.Concat], &BinOp{
			LeftType:     types.MakeArray(typ),
			RightType:    types.MakeArray(typ),
			ReturnType:   types.MakeArray(typ),
			NullableArgs: true,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607789)
				return ConcatArrays(typ, left, right)
			},
			Volatility: VolatilityImmutable,
		})
	}
}

func initNonArrayToNonArrayConcatenation() {
	__antithesis_instrumentation__.Notify(607790)
	addConcat := func(leftType, rightType *types.T, volatility Volatility) {
		__antithesis_instrumentation__.Notify(607793)
		BinOps[treebin.Concat] = append(BinOps[treebin.Concat], &BinOp{
			LeftType:     leftType,
			RightType:    rightType,
			ReturnType:   types.String,
			NullableArgs: false,
			Fn: func(evalCtx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607794)
				if leftType == types.String {
					__antithesis_instrumentation__.Notify(607797)
					casted, err := PerformCast(evalCtx, right, types.String)
					if err != nil {
						__antithesis_instrumentation__.Notify(607799)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(607800)
					}
					__antithesis_instrumentation__.Notify(607798)
					return NewDString(string(MustBeDString(left)) + string(MustBeDString(casted))), nil
				} else {
					__antithesis_instrumentation__.Notify(607801)
				}
				__antithesis_instrumentation__.Notify(607795)
				if rightType == types.String {
					__antithesis_instrumentation__.Notify(607802)
					casted, err := PerformCast(evalCtx, left, types.String)
					if err != nil {
						__antithesis_instrumentation__.Notify(607804)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(607805)
					}
					__antithesis_instrumentation__.Notify(607803)
					return NewDString(string(MustBeDString(casted)) + string(MustBeDString(right))), nil
				} else {
					__antithesis_instrumentation__.Notify(607806)
				}
				__antithesis_instrumentation__.Notify(607796)
				return nil, errors.New("neither LHS or RHS matched DString")
			},
			Volatility: volatility,
		})
	}
	__antithesis_instrumentation__.Notify(607791)
	fromTypeToVolatility := make(map[oid.Oid]Volatility)
	ForEachCast(func(src, tgt oid.Oid) {
		__antithesis_instrumentation__.Notify(607807)
		if tgt == oid.T_text {
			__antithesis_instrumentation__.Notify(607808)
			fromTypeToVolatility[src] = castMap[src][tgt].volatility
		} else {
			__antithesis_instrumentation__.Notify(607809)
		}
	})
	__antithesis_instrumentation__.Notify(607792)

	for _, t := range append([]*types.T{types.AnyTuple}, types.Scalar...) {
		__antithesis_instrumentation__.Notify(607810)

		if t != types.String && func() bool {
			__antithesis_instrumentation__.Notify(607811)
			return t != types.Bytes == true
		}() == true {
			__antithesis_instrumentation__.Notify(607812)
			addConcat(t, types.String, fromTypeToVolatility[t.Oid()])
			addConcat(types.String, t, fromTypeToVolatility[t.Oid()])
		} else {
			__antithesis_instrumentation__.Notify(607813)
		}
	}
}

func init() {
	initArrayElementConcatenation()
	initArrayToArrayConcatenation()
	initNonArrayToNonArrayConcatenation()
}

func init() {
	for op, overload := range BinOps {
		for i, impl := range overload {
			casted := impl.(*BinOp)
			casted.types = ArgTypes{{"left", casted.LeftType}, {"right", casted.RightType}}
			casted.retType = FixedReturnType(casted.ReturnType)
			BinOps[op][i] = casted
		}
	}
}

type binOpOverload []overloadImpl

func (o binOpOverload) lookupImpl(left, right *types.T) (*BinOp, bool) {
	__antithesis_instrumentation__.Notify(607814)
	for _, fn := range o {
		__antithesis_instrumentation__.Notify(607816)
		casted := fn.(*BinOp)
		if casted.matchParams(left, right) {
			__antithesis_instrumentation__.Notify(607817)
			return casted, true
		} else {
			__antithesis_instrumentation__.Notify(607818)
		}
	}
	__antithesis_instrumentation__.Notify(607815)
	return nil, false
}

func GetJSONPath(j json.JSON, ary DArray) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(607819)

	path := make([]string, len(ary.Array))
	for i, v := range ary.Array {
		__antithesis_instrumentation__.Notify(607821)
		if v == DNull {
			__antithesis_instrumentation__.Notify(607823)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(607824)
		}
		__antithesis_instrumentation__.Notify(607822)
		path[i] = string(MustBeDString(v))
	}
	__antithesis_instrumentation__.Notify(607820)
	return json.FetchPath(j, path)
}

var BinOps = map[treebin.BinaryOperatorSymbol]binOpOverload{
	treebin.Bitand: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607825)
				return NewDInt(MustBeDInt(left) & MustBeDInt(right)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.VarBit,
			ReturnType: types.VarBit,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607826)
				lhs := MustBeDBitArray(left)
				rhs := MustBeDBitArray(right)
				if lhs.BitLen() != rhs.BitLen() {
					__antithesis_instrumentation__.Notify(607828)
					return nil, NewCannotMixBitArraySizesError("AND")
				} else {
					__antithesis_instrumentation__.Notify(607829)
				}
				__antithesis_instrumentation__.Notify(607827)
				return &DBitArray{
					BitArray: bitarray.And(lhs.BitArray, rhs.BitArray),
				}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.INet,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607830)
				ipAddr := MustBeDIPAddr(left).IPAddr
				other := MustBeDIPAddr(right).IPAddr
				newIPAddr, err := ipAddr.And(&other)
				return NewDIPAddr(DIPAddr{newIPAddr}), err
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Bitor: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607831)
				return NewDInt(MustBeDInt(left) | MustBeDInt(right)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.VarBit,
			ReturnType: types.VarBit,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607832)
				lhs := MustBeDBitArray(left)
				rhs := MustBeDBitArray(right)
				if lhs.BitLen() != rhs.BitLen() {
					__antithesis_instrumentation__.Notify(607834)
					return nil, NewCannotMixBitArraySizesError("OR")
				} else {
					__antithesis_instrumentation__.Notify(607835)
				}
				__antithesis_instrumentation__.Notify(607833)
				return &DBitArray{
					BitArray: bitarray.Or(lhs.BitArray, rhs.BitArray),
				}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.INet,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607836)
				ipAddr := MustBeDIPAddr(left).IPAddr
				other := MustBeDIPAddr(right).IPAddr
				newIPAddr, err := ipAddr.Or(&other)
				return NewDIPAddr(DIPAddr{newIPAddr}), err
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Bitxor: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607837)
				return NewDInt(MustBeDInt(left) ^ MustBeDInt(right)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.VarBit,
			ReturnType: types.VarBit,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607838)
				lhs := MustBeDBitArray(left)
				rhs := MustBeDBitArray(right)
				if lhs.BitLen() != rhs.BitLen() {
					__antithesis_instrumentation__.Notify(607840)
					return nil, NewCannotMixBitArraySizesError("XOR")
				} else {
					__antithesis_instrumentation__.Notify(607841)
				}
				__antithesis_instrumentation__.Notify(607839)
				return &DBitArray{
					BitArray: bitarray.Xor(lhs.BitArray, rhs.BitArray),
				}, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Plus: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607842)
				a, b := MustBeDInt(left), MustBeDInt(right)
				r, ok := arith.AddWithOverflow(int64(a), int64(b))
				if !ok {
					__antithesis_instrumentation__.Notify(607844)
					return nil, ErrIntOutOfRange
				} else {
					__antithesis_instrumentation__.Notify(607845)
				}
				__antithesis_instrumentation__.Notify(607843)
				return NewDInt(DInt(r)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607846)
				return NewDFloat(*left.(*DFloat) + *right.(*DFloat)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607847)
				l := &left.(*DDecimal).Decimal
				r := &right.(*DDecimal).Decimal
				dd := &DDecimal{}
				_, err := ExactCtx.Add(&dd.Decimal, l, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607848)
				l := &left.(*DDecimal).Decimal
				r := MustBeDInt(right)
				dd := &DDecimal{}
				dd.SetInt64(int64(r))
				_, err := ExactCtx.Add(&dd.Decimal, l, &dd.Decimal)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607849)
				l := MustBeDInt(left)
				r := &right.(*DDecimal).Decimal
				dd := &DDecimal{}
				dd.SetInt64(int64(l))
				_, err := ExactCtx.Add(&dd.Decimal, &dd.Decimal, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Int,
			ReturnType: types.Date,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607850)
				d, err := left.(*DDate).AddDays(int64(MustBeDInt(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(607852)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607853)
				}
				__antithesis_instrumentation__.Notify(607851)
				return NewDDate(d), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Date,
			ReturnType: types.Date,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607854)
				d, err := right.(*DDate).AddDays(int64(MustBeDInt(left)))
				if err != nil {
					__antithesis_instrumentation__.Notify(607856)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607857)
				}
				__antithesis_instrumentation__.Notify(607855)
				return NewDDate(d), nil

			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Time,
			ReturnType: types.Timestamp,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607858)
				leftTime, err := left.(*DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(607860)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607861)
				}
				__antithesis_instrumentation__.Notify(607859)
				t := time.Duration(*right.(*DTime)) * time.Microsecond
				return MakeDTimestamp(leftTime.Add(t), time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Time,
			RightType:  types.Date,
			ReturnType: types.Timestamp,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607862)
				rightTime, err := right.(*DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(607864)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607865)
				}
				__antithesis_instrumentation__.Notify(607863)
				t := time.Duration(*left.(*DTime)) * time.Microsecond
				return MakeDTimestamp(rightTime.Add(t), time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.TimeTZ,
			ReturnType: types.TimestampTZ,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607866)
				leftTime, err := left.(*DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(607868)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607869)
				}
				__antithesis_instrumentation__.Notify(607867)
				t := leftTime.Add(right.(*DTimeTZ).ToDuration())
				return MakeDTimestampTZ(t, time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimeTZ,
			RightType:  types.Date,
			ReturnType: types.TimestampTZ,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607870)
				rightTime, err := right.(*DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(607872)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607873)
				}
				__antithesis_instrumentation__.Notify(607871)
				t := rightTime.Add(left.(*DTimeTZ).ToDuration())
				return MakeDTimestampTZ(t, time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Time,
			RightType:  types.Interval,
			ReturnType: types.Time,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607874)
				t := timeofday.TimeOfDay(*left.(*DTime))
				return MakeDTime(t.Add(right.(*DInterval).Duration)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Time,
			ReturnType: types.Time,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607875)
				t := timeofday.TimeOfDay(*right.(*DTime))
				return MakeDTime(t.Add(left.(*DInterval).Duration)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimeTZ,
			RightType:  types.Interval,
			ReturnType: types.TimeTZ,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607876)
				t := left.(*DTimeTZ)
				duration := right.(*DInterval).Duration
				return NewDTimeTZFromOffset(t.Add(duration), t.OffsetSecs), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.TimeTZ,
			ReturnType: types.TimeTZ,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607877)
				t := right.(*DTimeTZ)
				duration := left.(*DInterval).Duration
				return NewDTimeTZFromOffset(t.Add(duration), t.OffsetSecs), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Timestamp,
			RightType:  types.Interval,
			ReturnType: types.Timestamp,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607878)
				return MakeDTimestamp(duration.Add(
					left.(*DTimestamp).Time, right.(*DInterval).Duration), time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Timestamp,
			ReturnType: types.Timestamp,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607879)
				return MakeDTimestamp(duration.Add(
					right.(*DTimestamp).Time, left.(*DInterval).Duration), time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimestampTZ,
			RightType:  types.Interval,
			ReturnType: types.TimestampTZ,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607880)

				t := duration.Add(left.(*DTimestampTZ).Time.In(ctx.GetLocation()), right.(*DInterval).Duration)
				return MakeDTimestampTZ(t, time.Microsecond)
			},
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.TimestampTZ,
			ReturnType: types.TimestampTZ,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607881)

				t := duration.Add(right.(*DTimestampTZ).Time.In(ctx.GetLocation()), left.(*DInterval).Duration)
				return MakeDTimestampTZ(t, time.Microsecond)
			},
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607882)
				return &DInterval{Duration: left.(*DInterval).Duration.Add(right.(*DInterval).Duration)}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Interval,
			ReturnType: types.Timestamp,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607883)
				leftTime, err := left.(*DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(607885)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607886)
				}
				__antithesis_instrumentation__.Notify(607884)
				t := duration.Add(leftTime, right.(*DInterval).Duration)
				return MakeDTimestamp(t, time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Date,
			ReturnType: types.Timestamp,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607887)
				rightTime, err := right.(*DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(607889)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607890)
				}
				__antithesis_instrumentation__.Notify(607888)
				t := duration.Add(rightTime, left.(*DInterval).Duration)
				return MakeDTimestamp(t, time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.Int,
			ReturnType: types.INet,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607891)
				ipAddr := MustBeDIPAddr(left).IPAddr
				i := MustBeDInt(right)
				newIPAddr, err := ipAddr.Add(int64(i))
				return NewDIPAddr(DIPAddr{newIPAddr}), err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.INet,
			ReturnType: types.INet,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607892)
				i := MustBeDInt(left)
				ipAddr := MustBeDIPAddr(right).IPAddr
				newIPAddr, err := ipAddr.Add(int64(i))
				return NewDIPAddr(DIPAddr{newIPAddr}), err
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Minus: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607893)
				a, b := MustBeDInt(left), MustBeDInt(right)
				r, ok := arith.SubWithOverflow(int64(a), int64(b))
				if !ok {
					__antithesis_instrumentation__.Notify(607895)
					return nil, ErrIntOutOfRange
				} else {
					__antithesis_instrumentation__.Notify(607896)
				}
				__antithesis_instrumentation__.Notify(607894)
				return NewDInt(DInt(r)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607897)
				return NewDFloat(*left.(*DFloat) - *right.(*DFloat)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607898)
				l := &left.(*DDecimal).Decimal
				r := &right.(*DDecimal).Decimal
				dd := &DDecimal{}
				_, err := ExactCtx.Sub(&dd.Decimal, l, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607899)
				l := &left.(*DDecimal).Decimal
				r := MustBeDInt(right)
				dd := &DDecimal{}
				dd.SetInt64(int64(r))
				_, err := ExactCtx.Sub(&dd.Decimal, l, &dd.Decimal)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607900)
				l := MustBeDInt(left)
				r := &right.(*DDecimal).Decimal
				dd := &DDecimal{}
				dd.SetInt64(int64(l))
				_, err := ExactCtx.Sub(&dd.Decimal, &dd.Decimal, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Int,
			ReturnType: types.Date,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607901)
				d, err := left.(*DDate).SubDays(int64(MustBeDInt(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(607903)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607904)
				}
				__antithesis_instrumentation__.Notify(607902)
				return NewDDate(d), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Date,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607905)
				l, r := left.(*DDate).Date, right.(*DDate).Date
				if !l.IsFinite() || func() bool {
					__antithesis_instrumentation__.Notify(607907)
					return !r.IsFinite() == true
				}() == true {
					__antithesis_instrumentation__.Notify(607908)
					return nil, pgerror.New(pgcode.DatetimeFieldOverflow, "cannot subtract infinite dates")
				} else {
					__antithesis_instrumentation__.Notify(607909)
				}
				__antithesis_instrumentation__.Notify(607906)
				a := l.PGEpochDays()
				b := r.PGEpochDays()

				return NewDInt(DInt(int64(a) - int64(b))), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Time,
			ReturnType: types.Timestamp,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607910)
				leftTime, err := left.(*DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(607912)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607913)
				}
				__antithesis_instrumentation__.Notify(607911)
				t := time.Duration(*right.(*DTime)) * time.Microsecond
				return MakeDTimestamp(leftTime.Add(-1*t), time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Time,
			RightType:  types.Time,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607914)
				t1 := timeofday.TimeOfDay(*left.(*DTime))
				t2 := timeofday.TimeOfDay(*right.(*DTime))
				diff := timeofday.Difference(t1, t2)
				return &DInterval{Duration: diff}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Timestamp,
			RightType:  types.Timestamp,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607915)
				nanos := left.(*DTimestamp).Sub(right.(*DTimestamp).Time).Nanoseconds()
				return &DInterval{Duration: duration.MakeDurationJustifyHours(nanos, 0, 0)}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimestampTZ,
			RightType:  types.TimestampTZ,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607916)
				nanos := left.(*DTimestampTZ).Sub(right.(*DTimestampTZ).Time).Nanoseconds()
				return &DInterval{Duration: duration.MakeDurationJustifyHours(nanos, 0, 0)}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Timestamp,
			RightType:  types.TimestampTZ,
			ReturnType: types.Interval,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607917)

				stripped, err := right.(*DTimestampTZ).stripTimeZone(ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(607919)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607920)
				}
				__antithesis_instrumentation__.Notify(607918)
				nanos := left.(*DTimestamp).Sub(stripped.Time).Nanoseconds()
				return &DInterval{Duration: duration.MakeDurationJustifyHours(nanos, 0, 0)}, nil
			},
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.TimestampTZ,
			RightType:  types.Timestamp,
			ReturnType: types.Interval,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607921)

				stripped, err := left.(*DTimestampTZ).stripTimeZone(ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(607923)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607924)
				}
				__antithesis_instrumentation__.Notify(607922)
				nanos := stripped.Sub(right.(*DTimestamp).Time).Nanoseconds()
				return &DInterval{Duration: duration.MakeDurationJustifyHours(nanos, 0, 0)}, nil
			},
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.Time,
			RightType:  types.Interval,
			ReturnType: types.Time,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607925)
				t := timeofday.TimeOfDay(*left.(*DTime))
				return MakeDTime(t.Add(right.(*DInterval).Duration.Mul(-1))), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimeTZ,
			RightType:  types.Interval,
			ReturnType: types.TimeTZ,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607926)
				t := left.(*DTimeTZ)
				duration := right.(*DInterval).Duration
				return NewDTimeTZFromOffset(t.Add(duration.Mul(-1)), t.OffsetSecs), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Timestamp,
			RightType:  types.Interval,
			ReturnType: types.Timestamp,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607927)
				return MakeDTimestamp(duration.Add(
					left.(*DTimestamp).Time, right.(*DInterval).Duration.Mul(-1)), time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.TimestampTZ,
			RightType:  types.Interval,
			ReturnType: types.TimestampTZ,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607928)
				t := duration.Add(
					left.(*DTimestampTZ).Time.In(ctx.GetLocation()),
					right.(*DInterval).Duration.Mul(-1),
				)
				return MakeDTimestampTZ(t, time.Microsecond)
			},
			Volatility: VolatilityStable,
		},
		&BinOp{
			LeftType:   types.Date,
			RightType:  types.Interval,
			ReturnType: types.Timestamp,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607929)
				leftTime, err := left.(*DDate).ToTime()
				if err != nil {
					__antithesis_instrumentation__.Notify(607931)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607932)
				}
				__antithesis_instrumentation__.Notify(607930)
				t := duration.Add(leftTime, right.(*DInterval).Duration.Mul(-1))
				return MakeDTimestamp(t, time.Microsecond)
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607933)
				return &DInterval{Duration: left.(*DInterval).Duration.Sub(right.(*DInterval).Duration)}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.String,
			ReturnType: types.Jsonb,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607934)
				j, _, err := left.(*DJSON).JSON.RemoveString(string(MustBeDString(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(607936)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607937)
				}
				__antithesis_instrumentation__.Notify(607935)
				return &DJSON{j}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.Int,
			ReturnType: types.Jsonb,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607938)
				j, _, err := left.(*DJSON).JSON.RemoveIndex(int(MustBeDInt(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(607940)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607941)
				}
				__antithesis_instrumentation__.Notify(607939)
				return &DJSON{j}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.MakeArray(types.String),
			ReturnType: types.Jsonb,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607942)
				j := left.(*DJSON).JSON
				arr := *MustBeDArray(right)

				for _, str := range arr.Array {
					__antithesis_instrumentation__.Notify(607944)
					if str == DNull {
						__antithesis_instrumentation__.Notify(607946)
						continue
					} else {
						__antithesis_instrumentation__.Notify(607947)
					}
					__antithesis_instrumentation__.Notify(607945)
					var err error
					j, _, err = j.RemoveString(string(MustBeDString(str)))
					if err != nil {
						__antithesis_instrumentation__.Notify(607948)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(607949)
					}
				}
				__antithesis_instrumentation__.Notify(607943)
				return &DJSON{j}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607950)
				ipAddr := MustBeDIPAddr(left).IPAddr
				other := MustBeDIPAddr(right).IPAddr
				diff, err := ipAddr.SubIPAddr(&other)
				return NewDInt(DInt(diff)), err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{

			LeftType:   types.INet,
			RightType:  types.Int,
			ReturnType: types.INet,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607951)
				ipAddr := MustBeDIPAddr(left).IPAddr
				i := MustBeDInt(right)
				newIPAddr, err := ipAddr.Sub(int64(i))
				return NewDIPAddr(DIPAddr{newIPAddr}), err
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Mult: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607952)

				a, b := MustBeDInt(left), MustBeDInt(right)
				c := a * b
				if a == 0 || func() bool {
					__antithesis_instrumentation__.Notify(607954)
					return b == 0 == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(607955)
					return a == 1 == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(607956)
					return b == 1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(607957)

				} else {
					__antithesis_instrumentation__.Notify(607958)
					if a == math.MinInt64 || func() bool {
						__antithesis_instrumentation__.Notify(607959)
						return b == math.MinInt64 == true
					}() == true {
						__antithesis_instrumentation__.Notify(607960)

						return nil, ErrIntOutOfRange
					} else {
						__antithesis_instrumentation__.Notify(607961)
						if c/b != a {
							__antithesis_instrumentation__.Notify(607962)
							return nil, ErrIntOutOfRange
						} else {
							__antithesis_instrumentation__.Notify(607963)
						}
					}
				}
				__antithesis_instrumentation__.Notify(607953)
				return NewDInt(c), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607964)
				return NewDFloat(*left.(*DFloat) * *right.(*DFloat)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607965)
				l := &left.(*DDecimal).Decimal
				r := &right.(*DDecimal).Decimal
				dd := &DDecimal{}
				_, err := ExactCtx.Mul(&dd.Decimal, l, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},

		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607966)
				l := &left.(*DDecimal).Decimal
				r := MustBeDInt(right)
				dd := &DDecimal{}
				dd.SetInt64(int64(r))
				_, err := ExactCtx.Mul(&dd.Decimal, l, &dd.Decimal)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607967)
				l := MustBeDInt(left)
				r := &right.(*DDecimal).Decimal
				dd := &DDecimal{}
				dd.SetInt64(int64(l))
				_, err := ExactCtx.Mul(&dd.Decimal, &dd.Decimal, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607968)
				return &DInterval{Duration: right.(*DInterval).Duration.Mul(int64(MustBeDInt(left)))}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Int,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607969)
				return &DInterval{Duration: left.(*DInterval).Duration.Mul(int64(MustBeDInt(right)))}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Float,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607970)
				r := float64(*right.(*DFloat))
				return &DInterval{Duration: left.(*DInterval).Duration.MulFloat(r)}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607971)
				l := float64(*left.(*DFloat))
				return &DInterval{Duration: right.(*DInterval).Duration.MulFloat(l)}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Interval,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607972)
				l := &left.(*DDecimal).Decimal
				t, err := l.Float64()
				if err != nil {
					__antithesis_instrumentation__.Notify(607974)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607975)
				}
				__antithesis_instrumentation__.Notify(607973)
				return &DInterval{Duration: right.(*DInterval).Duration.MulFloat(t)}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Decimal,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607976)
				r := &right.(*DDecimal).Decimal
				t, err := r.Float64()
				if err != nil {
					__antithesis_instrumentation__.Notify(607978)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(607979)
				}
				__antithesis_instrumentation__.Notify(607977)
				return &DInterval{Duration: left.(*DInterval).Duration.MulFloat(t)}, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Div: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607980)
				rInt := MustBeDInt(right)
				if rInt == 0 {
					__antithesis_instrumentation__.Notify(607982)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(607983)
				}
				__antithesis_instrumentation__.Notify(607981)
				var div apd.Decimal
				div.SetInt64(int64(rInt))
				dd := &DDecimal{}
				dd.SetInt64(int64(MustBeDInt(left)))
				_, err := DecimalCtx.Quo(&dd.Decimal, &dd.Decimal, &div)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607984)
				r := *right.(*DFloat)
				if r == 0.0 {
					__antithesis_instrumentation__.Notify(607986)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(607987)
				}
				__antithesis_instrumentation__.Notify(607985)
				return NewDFloat(*left.(*DFloat) / r), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607988)
				l := &left.(*DDecimal).Decimal
				r := &right.(*DDecimal).Decimal
				if r.IsZero() {
					__antithesis_instrumentation__.Notify(607990)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(607991)
				}
				__antithesis_instrumentation__.Notify(607989)
				dd := &DDecimal{}
				_, err := DecimalCtx.Quo(&dd.Decimal, l, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607992)
				l := &left.(*DDecimal).Decimal
				r := MustBeDInt(right)
				if r == 0 {
					__antithesis_instrumentation__.Notify(607994)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(607995)
				}
				__antithesis_instrumentation__.Notify(607993)
				dd := &DDecimal{}
				dd.SetInt64(int64(r))
				_, err := DecimalCtx.Quo(&dd.Decimal, l, &dd.Decimal)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(607996)
				l := MustBeDInt(left)
				r := &right.(*DDecimal).Decimal
				if r.IsZero() {
					__antithesis_instrumentation__.Notify(607998)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(607999)
				}
				__antithesis_instrumentation__.Notify(607997)
				dd := &DDecimal{}
				dd.SetInt64(int64(l))
				_, err := DecimalCtx.Quo(&dd.Decimal, &dd.Decimal, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Int,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608000)
				rInt := MustBeDInt(right)
				if rInt == 0 {
					__antithesis_instrumentation__.Notify(608002)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608003)
				}
				__antithesis_instrumentation__.Notify(608001)
				return &DInterval{Duration: left.(*DInterval).Duration.Div(int64(rInt))}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Interval,
			RightType:  types.Float,
			ReturnType: types.Interval,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608004)
				r := float64(*right.(*DFloat))
				if r == 0.0 {
					__antithesis_instrumentation__.Notify(608006)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608007)
				}
				__antithesis_instrumentation__.Notify(608005)
				return &DInterval{Duration: left.(*DInterval).Duration.DivFloat(r)}, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.FloorDiv: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608008)
				rInt := MustBeDInt(right)
				if rInt == 0 {
					__antithesis_instrumentation__.Notify(608010)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608011)
				}
				__antithesis_instrumentation__.Notify(608009)
				return NewDInt(MustBeDInt(left) / rInt), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608012)
				l := float64(*left.(*DFloat))
				r := float64(*right.(*DFloat))
				if r == 0.0 {
					__antithesis_instrumentation__.Notify(608014)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608015)
				}
				__antithesis_instrumentation__.Notify(608013)
				return NewDFloat(DFloat(math.Trunc(l / r))), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608016)
				l := &left.(*DDecimal).Decimal
				r := &right.(*DDecimal).Decimal
				if r.IsZero() {
					__antithesis_instrumentation__.Notify(608018)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608019)
				}
				__antithesis_instrumentation__.Notify(608017)
				dd := &DDecimal{}
				_, err := HighPrecisionCtx.QuoInteger(&dd.Decimal, l, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608020)
				l := &left.(*DDecimal).Decimal
				r := MustBeDInt(right)
				if r == 0 {
					__antithesis_instrumentation__.Notify(608022)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608023)
				}
				__antithesis_instrumentation__.Notify(608021)
				dd := &DDecimal{}
				dd.SetInt64(int64(r))
				_, err := HighPrecisionCtx.QuoInteger(&dd.Decimal, l, &dd.Decimal)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608024)
				l := MustBeDInt(left)
				r := &right.(*DDecimal).Decimal
				if r.IsZero() {
					__antithesis_instrumentation__.Notify(608026)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608027)
				}
				__antithesis_instrumentation__.Notify(608025)
				dd := &DDecimal{}
				dd.SetInt64(int64(l))
				_, err := HighPrecisionCtx.QuoInteger(&dd.Decimal, &dd.Decimal, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Mod: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608028)
				r := MustBeDInt(right)
				if r == 0 {
					__antithesis_instrumentation__.Notify(608030)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608031)
				}
				__antithesis_instrumentation__.Notify(608029)
				return NewDInt(MustBeDInt(left) % r), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608032)
				l := float64(*left.(*DFloat))
				r := float64(*right.(*DFloat))
				if r == 0.0 {
					__antithesis_instrumentation__.Notify(608034)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608035)
				}
				__antithesis_instrumentation__.Notify(608033)
				return NewDFloat(DFloat(math.Mod(l, r))), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608036)
				l := &left.(*DDecimal).Decimal
				r := &right.(*DDecimal).Decimal
				if r.IsZero() {
					__antithesis_instrumentation__.Notify(608038)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608039)
				}
				__antithesis_instrumentation__.Notify(608037)
				dd := &DDecimal{}
				_, err := HighPrecisionCtx.Rem(&dd.Decimal, l, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608040)
				l := &left.(*DDecimal).Decimal
				r := MustBeDInt(right)
				if r == 0 {
					__antithesis_instrumentation__.Notify(608042)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608043)
				}
				__antithesis_instrumentation__.Notify(608041)
				dd := &DDecimal{}
				dd.SetInt64(int64(r))
				_, err := HighPrecisionCtx.Rem(&dd.Decimal, l, &dd.Decimal)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608044)
				l := MustBeDInt(left)
				r := &right.(*DDecimal).Decimal
				if r.IsZero() {
					__antithesis_instrumentation__.Notify(608046)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(608047)
				}
				__antithesis_instrumentation__.Notify(608045)
				dd := &DDecimal{}
				dd.SetInt64(int64(l))
				_, err := HighPrecisionCtx.Rem(&dd.Decimal, &dd.Decimal, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Concat: {
		&BinOp{
			LeftType:   types.String,
			RightType:  types.String,
			ReturnType: types.String,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608048)
				return NewDString(string(MustBeDString(left) + MustBeDString(right))), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Bytes,
			RightType:  types.Bytes,
			ReturnType: types.Bytes,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608049)
				return NewDBytes(*left.(*DBytes) + *right.(*DBytes)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.VarBit,
			ReturnType: types.VarBit,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608050)
				lhs := MustBeDBitArray(left)
				rhs := MustBeDBitArray(right)
				return &DBitArray{
					BitArray: bitarray.Concat(lhs.BitArray, rhs.BitArray),
				}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.Jsonb,
			ReturnType: types.Jsonb,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608051)
				j, err := MustBeDJSON(left).JSON.Concat(MustBeDJSON(right).JSON)
				if err != nil {
					__antithesis_instrumentation__.Notify(608053)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608054)
				}
				__antithesis_instrumentation__.Notify(608052)
				return &DJSON{j}, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.LShift: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608055)
				rval := MustBeDInt(right)
				if rval < 0 || func() bool {
					__antithesis_instrumentation__.Notify(608057)
					return rval >= 64 == true
				}() == true {
					__antithesis_instrumentation__.Notify(608058)
					telemetry.Inc(sqltelemetry.LargeLShiftArgumentCounter)
					return nil, ErrShiftArgOutOfRange
				} else {
					__antithesis_instrumentation__.Notify(608059)
				}
				__antithesis_instrumentation__.Notify(608056)
				return NewDInt(MustBeDInt(left) << uint(rval)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.Int,
			ReturnType: types.VarBit,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608060)
				lhs := MustBeDBitArray(left)
				rhs := MustBeDInt(right)
				return &DBitArray{
					BitArray: lhs.BitArray.LeftShiftAny(int64(rhs)),
				}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.Bool,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608061)
				ipAddr := MustBeDIPAddr(left).IPAddr
				other := MustBeDIPAddr(right).IPAddr
				return MakeDBool(DBool(ipAddr.ContainedBy(&other))), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.RShift: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608062)
				rval := MustBeDInt(right)
				if rval < 0 || func() bool {
					__antithesis_instrumentation__.Notify(608064)
					return rval >= 64 == true
				}() == true {
					__antithesis_instrumentation__.Notify(608065)
					telemetry.Inc(sqltelemetry.LargeRShiftArgumentCounter)
					return nil, ErrShiftArgOutOfRange
				} else {
					__antithesis_instrumentation__.Notify(608066)
				}
				__antithesis_instrumentation__.Notify(608063)
				return NewDInt(MustBeDInt(left) >> uint(rval)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.VarBit,
			RightType:  types.Int,
			ReturnType: types.VarBit,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608067)
				lhs := MustBeDBitArray(left)
				rhs := MustBeDInt(right)
				return &DBitArray{
					BitArray: lhs.BitArray.LeftShiftAny(-int64(rhs)),
				}, nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.INet,
			RightType:  types.INet,
			ReturnType: types.Bool,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608068)
				ipAddr := MustBeDIPAddr(left).IPAddr
				other := MustBeDIPAddr(right).IPAddr
				return MakeDBool(DBool(ipAddr.Contains(&other))), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.Pow: {
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Int,
			ReturnType: types.Int,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608069)
				return IntPow(MustBeDInt(left), MustBeDInt(right))
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Float,
			RightType:  types.Float,
			ReturnType: types.Float,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608070)
				f := math.Pow(float64(*left.(*DFloat)), float64(*right.(*DFloat)))
				return NewDFloat(DFloat(f)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608071)
				l := &left.(*DDecimal).Decimal
				r := &right.(*DDecimal).Decimal
				dd := &DDecimal{}
				_, err := DecimalCtx.Pow(&dd.Decimal, l, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Decimal,
			RightType:  types.Int,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608072)
				l := &left.(*DDecimal).Decimal
				r := MustBeDInt(right)
				dd := &DDecimal{}
				dd.SetInt64(int64(r))
				_, err := DecimalCtx.Pow(&dd.Decimal, l, &dd.Decimal)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Int,
			RightType:  types.Decimal,
			ReturnType: types.Decimal,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608073)
				l := MustBeDInt(left)
				r := &right.(*DDecimal).Decimal
				dd := &DDecimal{}
				dd.SetInt64(int64(l))
				_, err := DecimalCtx.Pow(&dd.Decimal, &dd.Decimal, r)
				return dd, err
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.JSONFetchVal: {
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.String,
			ReturnType: types.Jsonb,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608074)
				j, err := left.(*DJSON).JSON.FetchValKey(string(MustBeDString(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(608077)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608078)
				}
				__antithesis_instrumentation__.Notify(608075)
				if j == nil {
					__antithesis_instrumentation__.Notify(608079)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608080)
				}
				__antithesis_instrumentation__.Notify(608076)
				return &DJSON{j}, nil
			},
			PreferredOverload: true,
			Volatility:        VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.Int,
			ReturnType: types.Jsonb,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608081)
				j, err := left.(*DJSON).JSON.FetchValIdx(int(MustBeDInt(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(608084)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608085)
				}
				__antithesis_instrumentation__.Notify(608082)
				if j == nil {
					__antithesis_instrumentation__.Notify(608086)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608087)
				}
				__antithesis_instrumentation__.Notify(608083)
				return &DJSON{j}, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.JSONFetchValPath: {
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.MakeArray(types.String),
			ReturnType: types.Jsonb,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608088)
				path, err := GetJSONPath(left.(*DJSON).JSON, *MustBeDArray(right))
				if err != nil {
					__antithesis_instrumentation__.Notify(608091)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608092)
				}
				__antithesis_instrumentation__.Notify(608089)
				if path == nil {
					__antithesis_instrumentation__.Notify(608093)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608094)
				}
				__antithesis_instrumentation__.Notify(608090)
				return &DJSON{path}, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.JSONFetchText: {
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.String,
			ReturnType: types.String,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608095)
				res, err := left.(*DJSON).JSON.FetchValKey(string(MustBeDString(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(608100)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608101)
				}
				__antithesis_instrumentation__.Notify(608096)
				if res == nil {
					__antithesis_instrumentation__.Notify(608102)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608103)
				}
				__antithesis_instrumentation__.Notify(608097)
				text, err := res.AsText()
				if err != nil {
					__antithesis_instrumentation__.Notify(608104)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608105)
				}
				__antithesis_instrumentation__.Notify(608098)
				if text == nil {
					__antithesis_instrumentation__.Notify(608106)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608107)
				}
				__antithesis_instrumentation__.Notify(608099)
				return NewDString(*text), nil
			},
			PreferredOverload: true,
			Volatility:        VolatilityImmutable,
		},
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.Int,
			ReturnType: types.String,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608108)
				res, err := left.(*DJSON).JSON.FetchValIdx(int(MustBeDInt(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(608113)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608114)
				}
				__antithesis_instrumentation__.Notify(608109)
				if res == nil {
					__antithesis_instrumentation__.Notify(608115)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608116)
				}
				__antithesis_instrumentation__.Notify(608110)
				text, err := res.AsText()
				if err != nil {
					__antithesis_instrumentation__.Notify(608117)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608118)
				}
				__antithesis_instrumentation__.Notify(608111)
				if text == nil {
					__antithesis_instrumentation__.Notify(608119)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608120)
				}
				__antithesis_instrumentation__.Notify(608112)
				return NewDString(*text), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treebin.JSONFetchTextPath: {
		&BinOp{
			LeftType:   types.Jsonb,
			RightType:  types.MakeArray(types.String),
			ReturnType: types.String,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608121)
				res, err := GetJSONPath(left.(*DJSON).JSON, *MustBeDArray(right))
				if err != nil {
					__antithesis_instrumentation__.Notify(608126)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608127)
				}
				__antithesis_instrumentation__.Notify(608122)
				if res == nil {
					__antithesis_instrumentation__.Notify(608128)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608129)
				}
				__antithesis_instrumentation__.Notify(608123)
				text, err := res.AsText()
				if err != nil {
					__antithesis_instrumentation__.Notify(608130)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608131)
				}
				__antithesis_instrumentation__.Notify(608124)
				if text == nil {
					__antithesis_instrumentation__.Notify(608132)
					return DNull, nil
				} else {
					__antithesis_instrumentation__.Notify(608133)
				}
				__antithesis_instrumentation__.Notify(608125)
				return NewDString(*text), nil
			},
			Volatility: VolatilityImmutable,
		},
	},
}

type CmpOp struct {
	types TypeList

	LeftType  *types.T
	RightType *types.T

	Fn TwoArgFn

	counter telemetry.Counter

	NullableArgs bool

	Volatility Volatility

	PreferredOverload bool
}

func (op *CmpOp) params() TypeList {
	__antithesis_instrumentation__.Notify(608134)
	return op.types
}

func (op *CmpOp) matchParams(l, r *types.T) bool {
	__antithesis_instrumentation__.Notify(608135)
	return op.params().MatchAt(l, 0) && func() bool {
		__antithesis_instrumentation__.Notify(608136)
		return op.params().MatchAt(r, 1) == true
	}() == true
}

var cmpOpReturnType = FixedReturnType(types.Bool)

func (op *CmpOp) returnType() ReturnTyper {
	__antithesis_instrumentation__.Notify(608137)
	return cmpOpReturnType
}

func (op *CmpOp) preferred() bool {
	__antithesis_instrumentation__.Notify(608138)
	return op.PreferredOverload
}

func cmpOpFixups(
	cmpOps map[treecmp.ComparisonOperatorSymbol]cmpOpOverload,
) map[treecmp.ComparisonOperatorSymbol]cmpOpOverload {
	__antithesis_instrumentation__.Notify(608139)
	findVolatility := func(op treecmp.ComparisonOperatorSymbol, t *types.T) Volatility {
		__antithesis_instrumentation__.Notify(608143)
		for _, impl := range cmpOps[treecmp.EQ] {
			__antithesis_instrumentation__.Notify(608145)
			o := impl.(*CmpOp)
			if o.LeftType.Equivalent(t) && func() bool {
				__antithesis_instrumentation__.Notify(608146)
				return o.RightType.Equivalent(t) == true
			}() == true {
				__antithesis_instrumentation__.Notify(608147)
				return o.Volatility
			} else {
				__antithesis_instrumentation__.Notify(608148)
			}
		}
		__antithesis_instrumentation__.Notify(608144)
		panic(errors.AssertionFailedf("could not find cmp op %s(%s,%s)", op, t, t))
	}
	__antithesis_instrumentation__.Notify(608140)

	for _, t := range append(types.Scalar, types.AnyEnum) {
		__antithesis_instrumentation__.Notify(608149)
		cmpOps[treecmp.EQ] = append(cmpOps[treecmp.EQ], &CmpOp{
			LeftType:   types.MakeArray(t),
			RightType:  types.MakeArray(t),
			Fn:         cmpOpScalarEQFn,
			Volatility: findVolatility(treecmp.EQ, t),
		})
		cmpOps[treecmp.LE] = append(cmpOps[treecmp.LE], &CmpOp{
			LeftType:   types.MakeArray(t),
			RightType:  types.MakeArray(t),
			Fn:         cmpOpScalarLEFn,
			Volatility: findVolatility(treecmp.LE, t),
		})
		cmpOps[treecmp.LT] = append(cmpOps[treecmp.LT], &CmpOp{
			LeftType:   types.MakeArray(t),
			RightType:  types.MakeArray(t),
			Fn:         cmpOpScalarLTFn,
			Volatility: findVolatility(treecmp.LT, t),
		})

		cmpOps[treecmp.IsNotDistinctFrom] = append(cmpOps[treecmp.IsNotDistinctFrom], &CmpOp{
			LeftType:     types.MakeArray(t),
			RightType:    types.MakeArray(t),
			Fn:           cmpOpScalarIsFn,
			NullableArgs: true,
			Volatility:   findVolatility(treecmp.IsNotDistinctFrom, t),
		})
	}
	__antithesis_instrumentation__.Notify(608141)

	for op, overload := range cmpOps {
		__antithesis_instrumentation__.Notify(608150)
		for i, impl := range overload {
			__antithesis_instrumentation__.Notify(608151)
			casted := impl.(*CmpOp)
			casted.types = ArgTypes{{"left", casted.LeftType}, {"right", casted.RightType}}
			cmpOps[op][i] = casted
		}
	}
	__antithesis_instrumentation__.Notify(608142)

	return cmpOps
}

type cmpOpOverload []overloadImpl

func (o cmpOpOverload) LookupImpl(left, right *types.T) (*CmpOp, bool) {
	__antithesis_instrumentation__.Notify(608152)
	for _, fn := range o {
		__antithesis_instrumentation__.Notify(608154)
		casted := fn.(*CmpOp)
		if casted.matchParams(left, right) {
			__antithesis_instrumentation__.Notify(608155)
			return casted, true
		} else {
			__antithesis_instrumentation__.Notify(608156)
		}
	}
	__antithesis_instrumentation__.Notify(608153)
	return nil, false
}

func makeCmpOpOverload(
	fn func(ctx *EvalContext, left, right Datum) (Datum, error),
	a, b *types.T,
	nullableArgs bool,
	v Volatility,
) *CmpOp {
	__antithesis_instrumentation__.Notify(608157)
	return &CmpOp{
		LeftType:     a,
		RightType:    b,
		Fn:           fn,
		NullableArgs: nullableArgs,
		Volatility:   v,
	}
}

func makeEqFn(a, b *types.T, v Volatility) *CmpOp {
	__antithesis_instrumentation__.Notify(608158)
	return makeCmpOpOverload(cmpOpScalarEQFn, a, b, false, v)
}
func makeLtFn(a, b *types.T, v Volatility) *CmpOp {
	__antithesis_instrumentation__.Notify(608159)
	return makeCmpOpOverload(cmpOpScalarLTFn, a, b, false, v)
}
func makeLeFn(a, b *types.T, v Volatility) *CmpOp {
	__antithesis_instrumentation__.Notify(608160)
	return makeCmpOpOverload(cmpOpScalarLEFn, a, b, false, v)
}
func makeIsFn(a, b *types.T, v Volatility) *CmpOp {
	__antithesis_instrumentation__.Notify(608161)
	return makeCmpOpOverload(cmpOpScalarIsFn, a, b, true, v)
}

var CmpOps = cmpOpFixups(map[treecmp.ComparisonOperatorSymbol]cmpOpOverload{
	treecmp.EQ: {

		makeEqFn(types.AnyEnum, types.AnyEnum, VolatilityImmutable),
		makeEqFn(types.Bool, types.Bool, VolatilityLeakProof),
		makeEqFn(types.Bytes, types.Bytes, VolatilityLeakProof),
		makeEqFn(types.Date, types.Date, VolatilityLeakProof),
		makeEqFn(types.Decimal, types.Decimal, VolatilityImmutable),

		makeEqFn(types.AnyCollatedString, types.AnyCollatedString, VolatilityLeakProof),
		makeEqFn(types.Float, types.Float, VolatilityLeakProof),
		makeEqFn(types.Box2D, types.Box2D, VolatilityLeakProof),
		makeEqFn(types.Geography, types.Geography, VolatilityLeakProof),
		makeEqFn(types.Geometry, types.Geometry, VolatilityLeakProof),
		makeEqFn(types.INet, types.INet, VolatilityLeakProof),
		makeEqFn(types.Int, types.Int, VolatilityLeakProof),
		makeEqFn(types.Interval, types.Interval, VolatilityLeakProof),
		makeEqFn(types.Jsonb, types.Jsonb, VolatilityImmutable),
		makeEqFn(types.Oid, types.Oid, VolatilityLeakProof),
		makeEqFn(types.String, types.String, VolatilityLeakProof),
		makeEqFn(types.Time, types.Time, VolatilityLeakProof),
		makeEqFn(types.TimeTZ, types.TimeTZ, VolatilityLeakProof),
		makeEqFn(types.Timestamp, types.Timestamp, VolatilityLeakProof),
		makeEqFn(types.TimestampTZ, types.TimestampTZ, VolatilityLeakProof),
		makeEqFn(types.Uuid, types.Uuid, VolatilityLeakProof),
		makeEqFn(types.VarBit, types.VarBit, VolatilityLeakProof),

		makeEqFn(types.Date, types.Timestamp, VolatilityImmutable),
		makeEqFn(types.Date, types.TimestampTZ, VolatilityStable),
		makeEqFn(types.Decimal, types.Float, VolatilityLeakProof),
		makeEqFn(types.Decimal, types.Int, VolatilityLeakProof),
		makeEqFn(types.Float, types.Decimal, VolatilityLeakProof),
		makeEqFn(types.Float, types.Int, VolatilityLeakProof),
		makeEqFn(types.Int, types.Decimal, VolatilityLeakProof),
		makeEqFn(types.Int, types.Float, VolatilityLeakProof),
		makeEqFn(types.Int, types.Oid, VolatilityLeakProof),
		makeEqFn(types.Oid, types.Int, VolatilityLeakProof),
		makeEqFn(types.Timestamp, types.Date, VolatilityImmutable),
		makeEqFn(types.Timestamp, types.TimestampTZ, VolatilityStable),
		makeEqFn(types.TimestampTZ, types.Date, VolatilityStable),
		makeEqFn(types.TimestampTZ, types.Timestamp, VolatilityStable),
		makeEqFn(types.Time, types.TimeTZ, VolatilityStable),
		makeEqFn(types.TimeTZ, types.Time, VolatilityStable),

		&CmpOp{
			LeftType:  types.AnyTuple,
			RightType: types.AnyTuple,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608162)
				return cmpOpTupleFn(ctx, *left.(*DTuple), *right.(*DTuple), treecmp.MakeComparisonOperator(treecmp.EQ)), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.LT: {

		makeLtFn(types.AnyEnum, types.AnyEnum, VolatilityImmutable),
		makeLtFn(types.Bool, types.Bool, VolatilityLeakProof),
		makeLtFn(types.Bytes, types.Bytes, VolatilityLeakProof),
		makeLtFn(types.Date, types.Date, VolatilityLeakProof),
		makeLtFn(types.Decimal, types.Decimal, VolatilityImmutable),

		makeLtFn(types.AnyCollatedString, types.AnyCollatedString, VolatilityLeakProof),
		makeLtFn(types.Float, types.Float, VolatilityLeakProof),
		makeLtFn(types.Box2D, types.Box2D, VolatilityLeakProof),
		makeLtFn(types.Geography, types.Geography, VolatilityLeakProof),
		makeLtFn(types.Geometry, types.Geometry, VolatilityLeakProof),
		makeLtFn(types.INet, types.INet, VolatilityLeakProof),
		makeLtFn(types.Int, types.Int, VolatilityLeakProof),
		makeLtFn(types.Interval, types.Interval, VolatilityLeakProof),
		makeLtFn(types.Oid, types.Oid, VolatilityLeakProof),
		makeLtFn(types.String, types.String, VolatilityLeakProof),
		makeLtFn(types.Time, types.Time, VolatilityLeakProof),
		makeLtFn(types.TimeTZ, types.TimeTZ, VolatilityLeakProof),
		makeLtFn(types.Timestamp, types.Timestamp, VolatilityLeakProof),
		makeLtFn(types.TimestampTZ, types.TimestampTZ, VolatilityLeakProof),
		makeLtFn(types.Uuid, types.Uuid, VolatilityLeakProof),
		makeLtFn(types.VarBit, types.VarBit, VolatilityLeakProof),

		makeLtFn(types.Date, types.Timestamp, VolatilityImmutable),
		makeLtFn(types.Date, types.TimestampTZ, VolatilityStable),
		makeLtFn(types.Decimal, types.Float, VolatilityLeakProof),
		makeLtFn(types.Decimal, types.Int, VolatilityLeakProof),
		makeLtFn(types.Float, types.Decimal, VolatilityLeakProof),
		makeLtFn(types.Float, types.Int, VolatilityLeakProof),
		makeLtFn(types.Int, types.Decimal, VolatilityLeakProof),
		makeLtFn(types.Int, types.Float, VolatilityLeakProof),
		makeLtFn(types.Int, types.Oid, VolatilityLeakProof),
		makeLtFn(types.Oid, types.Int, VolatilityLeakProof),
		makeLtFn(types.Timestamp, types.Date, VolatilityImmutable),
		makeLtFn(types.Timestamp, types.TimestampTZ, VolatilityStable),
		makeLtFn(types.TimestampTZ, types.Date, VolatilityStable),
		makeLtFn(types.TimestampTZ, types.Timestamp, VolatilityStable),
		makeLtFn(types.Time, types.TimeTZ, VolatilityStable),
		makeLtFn(types.TimeTZ, types.Time, VolatilityStable),

		&CmpOp{
			LeftType:  types.AnyTuple,
			RightType: types.AnyTuple,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608163)
				return cmpOpTupleFn(ctx, *left.(*DTuple), *right.(*DTuple), treecmp.MakeComparisonOperator(treecmp.LT)), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.LE: {

		makeLeFn(types.AnyEnum, types.AnyEnum, VolatilityImmutable),
		makeLeFn(types.Bool, types.Bool, VolatilityLeakProof),
		makeLeFn(types.Bytes, types.Bytes, VolatilityLeakProof),
		makeLeFn(types.Date, types.Date, VolatilityLeakProof),
		makeLeFn(types.Decimal, types.Decimal, VolatilityImmutable),

		makeLeFn(types.AnyCollatedString, types.AnyCollatedString, VolatilityLeakProof),
		makeLeFn(types.Float, types.Float, VolatilityLeakProof),
		makeLeFn(types.Box2D, types.Box2D, VolatilityLeakProof),
		makeLeFn(types.Geography, types.Geography, VolatilityLeakProof),
		makeLeFn(types.Geometry, types.Geometry, VolatilityLeakProof),
		makeLeFn(types.INet, types.INet, VolatilityLeakProof),
		makeLeFn(types.Int, types.Int, VolatilityLeakProof),
		makeLeFn(types.Interval, types.Interval, VolatilityLeakProof),
		makeLeFn(types.Oid, types.Oid, VolatilityLeakProof),
		makeLeFn(types.String, types.String, VolatilityLeakProof),
		makeLeFn(types.Time, types.Time, VolatilityLeakProof),
		makeLeFn(types.TimeTZ, types.TimeTZ, VolatilityLeakProof),
		makeLeFn(types.Timestamp, types.Timestamp, VolatilityLeakProof),
		makeLeFn(types.TimestampTZ, types.TimestampTZ, VolatilityLeakProof),
		makeLeFn(types.Uuid, types.Uuid, VolatilityLeakProof),
		makeLeFn(types.VarBit, types.VarBit, VolatilityLeakProof),

		makeLeFn(types.Date, types.Timestamp, VolatilityImmutable),
		makeLeFn(types.Date, types.TimestampTZ, VolatilityStable),
		makeLeFn(types.Decimal, types.Float, VolatilityLeakProof),
		makeLeFn(types.Decimal, types.Int, VolatilityLeakProof),
		makeLeFn(types.Float, types.Decimal, VolatilityLeakProof),
		makeLeFn(types.Float, types.Int, VolatilityLeakProof),
		makeLeFn(types.Int, types.Decimal, VolatilityLeakProof),
		makeLeFn(types.Int, types.Float, VolatilityLeakProof),
		makeLeFn(types.Int, types.Oid, VolatilityLeakProof),
		makeLeFn(types.Oid, types.Int, VolatilityLeakProof),
		makeLeFn(types.Timestamp, types.Date, VolatilityImmutable),
		makeLeFn(types.Timestamp, types.TimestampTZ, VolatilityStable),
		makeLeFn(types.TimestampTZ, types.Date, VolatilityStable),
		makeLeFn(types.TimestampTZ, types.Timestamp, VolatilityStable),
		makeLeFn(types.Time, types.TimeTZ, VolatilityStable),
		makeLeFn(types.TimeTZ, types.Time, VolatilityStable),

		&CmpOp{
			LeftType:  types.AnyTuple,
			RightType: types.AnyTuple,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608164)
				return cmpOpTupleFn(ctx, *left.(*DTuple), *right.(*DTuple), treecmp.MakeComparisonOperator(treecmp.LE)), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.IsNotDistinctFrom: {
		&CmpOp{
			LeftType:     types.Unknown,
			RightType:    types.Unknown,
			Fn:           cmpOpScalarIsFn,
			NullableArgs: true,

			PreferredOverload: true,
			Volatility:        VolatilityLeakProof,
		},
		&CmpOp{
			LeftType:     types.AnyArray,
			RightType:    types.Unknown,
			Fn:           cmpOpScalarIsFn,
			NullableArgs: true,
			Volatility:   VolatilityLeakProof,
		},

		makeIsFn(types.AnyEnum, types.AnyEnum, VolatilityImmutable),
		makeIsFn(types.Bool, types.Bool, VolatilityLeakProof),
		makeIsFn(types.Bytes, types.Bytes, VolatilityLeakProof),
		makeIsFn(types.Date, types.Date, VolatilityLeakProof),
		makeIsFn(types.Decimal, types.Decimal, VolatilityImmutable),

		makeIsFn(types.AnyCollatedString, types.AnyCollatedString, VolatilityLeakProof),
		makeIsFn(types.Float, types.Float, VolatilityLeakProof),
		makeIsFn(types.Box2D, types.Box2D, VolatilityLeakProof),
		makeIsFn(types.Geography, types.Geography, VolatilityLeakProof),
		makeIsFn(types.Geometry, types.Geometry, VolatilityLeakProof),
		makeIsFn(types.INet, types.INet, VolatilityLeakProof),
		makeIsFn(types.Int, types.Int, VolatilityLeakProof),
		makeIsFn(types.Interval, types.Interval, VolatilityLeakProof),
		makeIsFn(types.Jsonb, types.Jsonb, VolatilityImmutable),
		makeIsFn(types.Oid, types.Oid, VolatilityLeakProof),
		makeIsFn(types.String, types.String, VolatilityLeakProof),
		makeIsFn(types.Time, types.Time, VolatilityLeakProof),
		makeIsFn(types.TimeTZ, types.TimeTZ, VolatilityLeakProof),
		makeIsFn(types.Timestamp, types.Timestamp, VolatilityLeakProof),
		makeIsFn(types.TimestampTZ, types.TimestampTZ, VolatilityLeakProof),
		makeIsFn(types.Uuid, types.Uuid, VolatilityLeakProof),
		makeIsFn(types.VarBit, types.VarBit, VolatilityLeakProof),

		makeIsFn(types.Date, types.Timestamp, VolatilityImmutable),
		makeIsFn(types.Date, types.TimestampTZ, VolatilityStable),
		makeIsFn(types.Decimal, types.Float, VolatilityLeakProof),
		makeIsFn(types.Decimal, types.Int, VolatilityLeakProof),
		makeIsFn(types.Float, types.Decimal, VolatilityLeakProof),
		makeIsFn(types.Float, types.Int, VolatilityLeakProof),
		makeIsFn(types.Int, types.Decimal, VolatilityLeakProof),
		makeIsFn(types.Int, types.Float, VolatilityLeakProof),
		makeIsFn(types.Int, types.Oid, VolatilityLeakProof),
		makeIsFn(types.Oid, types.Int, VolatilityLeakProof),
		makeIsFn(types.Timestamp, types.Date, VolatilityImmutable),
		makeIsFn(types.Timestamp, types.TimestampTZ, VolatilityStable),
		makeIsFn(types.TimestampTZ, types.Date, VolatilityStable),
		makeIsFn(types.TimestampTZ, types.Timestamp, VolatilityStable),
		makeIsFn(types.Time, types.TimeTZ, VolatilityStable),
		makeIsFn(types.TimeTZ, types.Time, VolatilityStable),

		&CmpOp{
			LeftType:     types.AnyTuple,
			RightType:    types.AnyTuple,
			NullableArgs: true,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608165)
				if left == DNull || func() bool {
					__antithesis_instrumentation__.Notify(608167)
					return right == DNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(608168)
					return MakeDBool(left == DNull && func() bool {
						__antithesis_instrumentation__.Notify(608169)
						return right == DNull == true
					}() == true), nil
				} else {
					__antithesis_instrumentation__.Notify(608170)
				}
				__antithesis_instrumentation__.Notify(608166)
				return cmpOpTupleFn(ctx, *left.(*DTuple), *right.(*DTuple), treecmp.MakeComparisonOperator(treecmp.IsNotDistinctFrom)), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.In: {
		makeEvalTupleIn(types.AnyEnum, VolatilityLeakProof),
		makeEvalTupleIn(types.Bool, VolatilityLeakProof),
		makeEvalTupleIn(types.Bytes, VolatilityLeakProof),
		makeEvalTupleIn(types.Date, VolatilityLeakProof),
		makeEvalTupleIn(types.Decimal, VolatilityLeakProof),
		makeEvalTupleIn(types.AnyCollatedString, VolatilityLeakProof),
		makeEvalTupleIn(types.AnyTuple, VolatilityLeakProof),
		makeEvalTupleIn(types.Float, VolatilityLeakProof),
		makeEvalTupleIn(types.Box2D, VolatilityLeakProof),
		makeEvalTupleIn(types.Geography, VolatilityLeakProof),
		makeEvalTupleIn(types.Geometry, VolatilityLeakProof),
		makeEvalTupleIn(types.INet, VolatilityLeakProof),
		makeEvalTupleIn(types.Int, VolatilityLeakProof),
		makeEvalTupleIn(types.Interval, VolatilityLeakProof),
		makeEvalTupleIn(types.Jsonb, VolatilityLeakProof),
		makeEvalTupleIn(types.Oid, VolatilityLeakProof),
		makeEvalTupleIn(types.String, VolatilityLeakProof),
		makeEvalTupleIn(types.Time, VolatilityLeakProof),
		makeEvalTupleIn(types.TimeTZ, VolatilityLeakProof),
		makeEvalTupleIn(types.Timestamp, VolatilityLeakProof),
		makeEvalTupleIn(types.TimestampTZ, VolatilityLeakProof),
		makeEvalTupleIn(types.Uuid, VolatilityLeakProof),
		makeEvalTupleIn(types.VarBit, VolatilityLeakProof),
	},

	treecmp.Like: {
		&CmpOp{
			LeftType:  types.String,
			RightType: types.String,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608171)
				return matchLike(ctx, left, right, false)
			},
			Volatility: VolatilityLeakProof,
		},
	},

	treecmp.ILike: {
		&CmpOp{
			LeftType:  types.String,
			RightType: types.String,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608172)
				return matchLike(ctx, left, right, true)
			},
			Volatility: VolatilityLeakProof,
		},
	},

	treecmp.SimilarTo: {
		&CmpOp{
			LeftType:  types.String,
			RightType: types.String,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608173)
				key := similarToKey{s: string(MustBeDString(right)), escape: '\\'}
				return matchRegexpWithKey(ctx, left, key)
			},
			Volatility: VolatilityLeakProof,
		},
	},

	treecmp.RegMatch: append(
		cmpOpOverload{
			&CmpOp{
				LeftType:  types.String,
				RightType: types.String,
				Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
					__antithesis_instrumentation__.Notify(608174)
					key := regexpKey{s: string(MustBeDString(right)), caseInsensitive: false}
					return matchRegexpWithKey(ctx, left, key)
				},
				Volatility: VolatilityImmutable,
			},
		},
		makeBox2DComparisonOperators(
			func(lhs, rhs *geo.CartesianBoundingBox) bool {
				__antithesis_instrumentation__.Notify(608175)
				return lhs.Covers(rhs)
			},
		)...,
	),

	treecmp.RegIMatch: {
		&CmpOp{
			LeftType:  types.String,
			RightType: types.String,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608176)
				key := regexpKey{s: string(MustBeDString(right)), caseInsensitive: true}
				return matchRegexpWithKey(ctx, left, key)
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.JSONExists: {
		&CmpOp{
			LeftType:  types.Jsonb,
			RightType: types.String,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608177)
				e, err := left.(*DJSON).JSON.Exists(string(MustBeDString(right)))
				if err != nil {
					__antithesis_instrumentation__.Notify(608180)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608181)
				}
				__antithesis_instrumentation__.Notify(608178)
				if e {
					__antithesis_instrumentation__.Notify(608182)
					return DBoolTrue, nil
				} else {
					__antithesis_instrumentation__.Notify(608183)
				}
				__antithesis_instrumentation__.Notify(608179)
				return DBoolFalse, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.JSONSomeExists: {
		&CmpOp{
			LeftType:  types.Jsonb,
			RightType: types.StringArray,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608184)
				return JSONExistsAny(ctx, MustBeDJSON(left), MustBeDArray(right))
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.JSONAllExists: {
		&CmpOp{
			LeftType:  types.Jsonb,
			RightType: types.StringArray,
			Fn: func(_ *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608185)

				for _, k := range MustBeDArray(right).Array {
					__antithesis_instrumentation__.Notify(608187)
					if k == DNull {
						__antithesis_instrumentation__.Notify(608190)
						continue
					} else {
						__antithesis_instrumentation__.Notify(608191)
					}
					__antithesis_instrumentation__.Notify(608188)
					e, err := left.(*DJSON).JSON.Exists(string(MustBeDString(k)))
					if err != nil {
						__antithesis_instrumentation__.Notify(608192)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(608193)
					}
					__antithesis_instrumentation__.Notify(608189)
					if !e {
						__antithesis_instrumentation__.Notify(608194)
						return DBoolFalse, nil
					} else {
						__antithesis_instrumentation__.Notify(608195)
					}
				}
				__antithesis_instrumentation__.Notify(608186)
				return DBoolTrue, nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.Contains: {
		&CmpOp{
			LeftType:  types.AnyArray,
			RightType: types.AnyArray,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608196)
				haystack := MustBeDArray(left)
				needles := MustBeDArray(right)
				return ArrayContains(ctx, haystack, needles)
			},
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:  types.Jsonb,
			RightType: types.Jsonb,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608197)
				c, err := json.Contains(left.(*DJSON).JSON, right.(*DJSON).JSON)
				if err != nil {
					__antithesis_instrumentation__.Notify(608199)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608200)
				}
				__antithesis_instrumentation__.Notify(608198)
				return MakeDBool(DBool(c)), nil
			},
			Volatility: VolatilityImmutable,
		},
	},

	treecmp.ContainedBy: {
		&CmpOp{
			LeftType:  types.AnyArray,
			RightType: types.AnyArray,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608201)
				needles := MustBeDArray(left)
				haystack := MustBeDArray(right)
				return ArrayContains(ctx, haystack, needles)
			},
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:  types.Jsonb,
			RightType: types.Jsonb,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608202)
				c, err := json.Contains(right.(*DJSON).JSON, left.(*DJSON).JSON)
				if err != nil {
					__antithesis_instrumentation__.Notify(608204)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608205)
				}
				__antithesis_instrumentation__.Notify(608203)
				return MakeDBool(DBool(c)), nil
			},
			Volatility: VolatilityImmutable,
		},
	},
	treecmp.Overlaps: append(
		cmpOpOverload{
			&CmpOp{
				LeftType:  types.AnyArray,
				RightType: types.AnyArray,
				Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
					__antithesis_instrumentation__.Notify(608206)
					array := MustBeDArray(left)
					other := MustBeDArray(right)
					if !array.ParamTyp.Equivalent(other.ParamTyp) {
						__antithesis_instrumentation__.Notify(608209)
						return nil, pgerror.New(pgcode.DatatypeMismatch, "cannot compare arrays with different element types")
					} else {
						__antithesis_instrumentation__.Notify(608210)
					}
					__antithesis_instrumentation__.Notify(608207)
					for _, needle := range array.Array {
						__antithesis_instrumentation__.Notify(608211)

						if needle == DNull {
							__antithesis_instrumentation__.Notify(608213)
							continue
						} else {
							__antithesis_instrumentation__.Notify(608214)
						}
						__antithesis_instrumentation__.Notify(608212)
						for _, hay := range other.Array {
							__antithesis_instrumentation__.Notify(608215)
							if needle.Compare(ctx, hay) == 0 {
								__antithesis_instrumentation__.Notify(608216)
								return DBoolTrue, nil
							} else {
								__antithesis_instrumentation__.Notify(608217)
							}
						}
					}
					__antithesis_instrumentation__.Notify(608208)
					return DBoolFalse, nil
				},
				Volatility: VolatilityImmutable,
			},
			&CmpOp{
				LeftType:  types.INet,
				RightType: types.INet,
				Fn: func(_ *EvalContext, left, right Datum) (Datum, error) {
					__antithesis_instrumentation__.Notify(608218)
					ipAddr := MustBeDIPAddr(left).IPAddr
					other := MustBeDIPAddr(right).IPAddr
					return MakeDBool(DBool(ipAddr.ContainsOrContainedBy(&other))), nil
				},
				Volatility: VolatilityImmutable,
			},
		},
		makeBox2DComparisonOperators(
			func(lhs, rhs *geo.CartesianBoundingBox) bool {
				__antithesis_instrumentation__.Notify(608219)
				return lhs.Intersects(rhs)
			},
		)...,
	),
})

const experimentalBox2DClusterSettingName = "sql.spatial.experimental_box2d_comparison_operators.enabled"

var experimentalBox2DClusterSetting = settings.RegisterBoolSetting(
	settings.TenantWritable,
	experimentalBox2DClusterSettingName,
	"enables the use of certain experimental box2d comparison operators",
	false,
).WithPublic()

func checkExperimentalBox2DComparisonOperatorEnabled(ctx *EvalContext) error {
	__antithesis_instrumentation__.Notify(608220)
	if !experimentalBox2DClusterSetting.Get(&ctx.Settings.SV) {
		__antithesis_instrumentation__.Notify(608222)
		return errors.WithHintf(
			pgerror.Newf(
				pgcode.FeatureNotSupported,
				"this box2d comparison operator is experimental",
			),
			"To enable box2d comparators, use `SET CLUSTER SETTING %s = on`.",
			experimentalBox2DClusterSettingName,
		)
	} else {
		__antithesis_instrumentation__.Notify(608223)
	}
	__antithesis_instrumentation__.Notify(608221)
	return nil
}

func makeBox2DComparisonOperators(op func(lhs, rhs *geo.CartesianBoundingBox) bool) cmpOpOverload {
	__antithesis_instrumentation__.Notify(608224)
	return cmpOpOverload{
		&CmpOp{
			LeftType:  types.Box2D,
			RightType: types.Box2D,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608225)
				if err := checkExperimentalBox2DComparisonOperatorEnabled(ctx); err != nil {
					__antithesis_instrumentation__.Notify(608227)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608228)
				}
				__antithesis_instrumentation__.Notify(608226)
				ret := op(
					&MustBeDBox2D(left).CartesianBoundingBox,
					&MustBeDBox2D(right).CartesianBoundingBox,
				)
				return MakeDBool(DBool(ret)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:  types.Box2D,
			RightType: types.Geometry,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608229)
				if err := checkExperimentalBox2DComparisonOperatorEnabled(ctx); err != nil {
					__antithesis_instrumentation__.Notify(608231)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608232)
				}
				__antithesis_instrumentation__.Notify(608230)
				ret := op(
					&MustBeDBox2D(left).CartesianBoundingBox,
					MustBeDGeometry(right).CartesianBoundingBox(),
				)
				return MakeDBool(DBool(ret)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:  types.Geometry,
			RightType: types.Box2D,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608233)
				if err := checkExperimentalBox2DComparisonOperatorEnabled(ctx); err != nil {
					__antithesis_instrumentation__.Notify(608235)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608236)
				}
				__antithesis_instrumentation__.Notify(608234)
				ret := op(
					MustBeDGeometry(left).CartesianBoundingBox(),
					&MustBeDBox2D(right).CartesianBoundingBox,
				)
				return MakeDBool(DBool(ret)), nil
			},
			Volatility: VolatilityImmutable,
		},
		&CmpOp{
			LeftType:  types.Geometry,
			RightType: types.Geometry,
			Fn: func(ctx *EvalContext, left Datum, right Datum) (Datum, error) {
				__antithesis_instrumentation__.Notify(608237)
				if err := checkExperimentalBox2DComparisonOperatorEnabled(ctx); err != nil {
					__antithesis_instrumentation__.Notify(608239)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(608240)
				}
				__antithesis_instrumentation__.Notify(608238)
				ret := op(
					MustBeDGeometry(left).CartesianBoundingBox(),
					MustBeDGeometry(right).CartesianBoundingBox(),
				)
				return MakeDBool(DBool(ret)), nil
			},
			Volatility: VolatilityImmutable,
		},
	}
}

var cmpOpsInverse map[treecmp.ComparisonOperatorSymbol]treecmp.ComparisonOperatorSymbol

func init() {
	cmpOpsInverse = make(map[treecmp.ComparisonOperatorSymbol]treecmp.ComparisonOperatorSymbol)
	for cmpOp := treecmp.ComparisonOperatorSymbol(0); cmpOp < treecmp.NumComparisonOperatorSymbols; cmpOp++ {
		newOp, _, _, _, _ := FoldComparisonExpr(treecmp.MakeComparisonOperator(cmpOp), DNull, DNull)
		if newOp.Symbol != cmpOp {
			cmpOpsInverse[newOp.Symbol] = cmpOp
			cmpOpsInverse[cmpOp] = newOp.Symbol
		}
	}
}

func CmpOpInverse(i treecmp.ComparisonOperatorSymbol) (treecmp.ComparisonOperatorSymbol, bool) {
	__antithesis_instrumentation__.Notify(608241)
	inverse, ok := cmpOpsInverse[i]
	return inverse, ok
}

func boolFromCmp(cmp int, op treecmp.ComparisonOperator) *DBool {
	__antithesis_instrumentation__.Notify(608242)
	switch op.Symbol {
	case treecmp.EQ, treecmp.IsNotDistinctFrom:
		__antithesis_instrumentation__.Notify(608243)
		return MakeDBool(cmp == 0)
	case treecmp.LT:
		__antithesis_instrumentation__.Notify(608244)
		return MakeDBool(cmp < 0)
	case treecmp.LE:
		__antithesis_instrumentation__.Notify(608245)
		return MakeDBool(cmp <= 0)
	default:
		__antithesis_instrumentation__.Notify(608246)
		panic(errors.AssertionFailedf("unexpected ComparisonOperator in boolFromCmp: %v", errors.Safe(op)))
	}
}

func cmpOpScalarFn(ctx *EvalContext, left, right Datum, op treecmp.ComparisonOperator) Datum {
	__antithesis_instrumentation__.Notify(608247)

	if left == DNull || func() bool {
		__antithesis_instrumentation__.Notify(608249)
		return right == DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(608250)
		switch op.Symbol {
		case treecmp.IsNotDistinctFrom:
			__antithesis_instrumentation__.Notify(608251)
			return MakeDBool((left == DNull) == (right == DNull))

		default:
			__antithesis_instrumentation__.Notify(608252)

			return DNull
		}
	} else {
		__antithesis_instrumentation__.Notify(608253)
	}
	__antithesis_instrumentation__.Notify(608248)
	cmp := left.Compare(ctx, right)
	return boolFromCmp(cmp, op)
}

func cmpOpScalarEQFn(ctx *EvalContext, left, right Datum) (Datum, error) {
	__antithesis_instrumentation__.Notify(608254)
	return cmpOpScalarFn(ctx, left, right, treecmp.MakeComparisonOperator(treecmp.EQ)), nil
}
func cmpOpScalarLTFn(ctx *EvalContext, left, right Datum) (Datum, error) {
	__antithesis_instrumentation__.Notify(608255)
	return cmpOpScalarFn(ctx, left, right, treecmp.MakeComparisonOperator(treecmp.LT)), nil
}
func cmpOpScalarLEFn(ctx *EvalContext, left, right Datum) (Datum, error) {
	__antithesis_instrumentation__.Notify(608256)
	return cmpOpScalarFn(ctx, left, right, treecmp.MakeComparisonOperator(treecmp.LE)), nil
}
func cmpOpScalarIsFn(ctx *EvalContext, left, right Datum) (Datum, error) {
	__antithesis_instrumentation__.Notify(608257)
	return cmpOpScalarFn(ctx, left, right, treecmp.MakeComparisonOperator(treecmp.IsNotDistinctFrom)), nil
}

func cmpOpTupleFn(ctx *EvalContext, left, right DTuple, op treecmp.ComparisonOperator) Datum {
	__antithesis_instrumentation__.Notify(608258)
	cmp := 0
	sawNull := false
	for i, leftElem := range left.D {
		__antithesis_instrumentation__.Notify(608261)
		rightElem := right.D[i]

		if leftElem == DNull || func() bool {
			__antithesis_instrumentation__.Notify(608262)
			return rightElem == DNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(608263)
			switch op.Symbol {
			case treecmp.EQ:
				__antithesis_instrumentation__.Notify(608264)

				sawNull = true

			case treecmp.IsNotDistinctFrom:
				__antithesis_instrumentation__.Notify(608265)

				if leftElem != DNull || func() bool {
					__antithesis_instrumentation__.Notify(608267)
					return rightElem != DNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(608268)
					return DBoolFalse
				} else {
					__antithesis_instrumentation__.Notify(608269)
				}

			default:
				__antithesis_instrumentation__.Notify(608266)

				return DNull
			}
		} else {
			__antithesis_instrumentation__.Notify(608270)
			cmp = leftElem.Compare(ctx, rightElem)
			if cmp != 0 {
				__antithesis_instrumentation__.Notify(608271)
				break
			} else {
				__antithesis_instrumentation__.Notify(608272)
			}
		}
	}
	__antithesis_instrumentation__.Notify(608259)
	b := boolFromCmp(cmp, op)
	if b == DBoolTrue && func() bool {
		__antithesis_instrumentation__.Notify(608273)
		return sawNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(608274)

		return DNull
	} else {
		__antithesis_instrumentation__.Notify(608275)
	}
	__antithesis_instrumentation__.Notify(608260)
	return b
}

func makeEvalTupleIn(typ *types.T, v Volatility) *CmpOp {
	__antithesis_instrumentation__.Notify(608276)
	return &CmpOp{
		LeftType:  typ,
		RightType: types.AnyTuple,
		Fn: func(ctx *EvalContext, arg, values Datum) (Datum, error) {
			__antithesis_instrumentation__.Notify(608277)
			vtuple := values.(*DTuple)

			if len(vtuple.D) == 0 {
				__antithesis_instrumentation__.Notify(608283)

				return DBoolFalse, nil
			} else {
				__antithesis_instrumentation__.Notify(608284)
			}
			__antithesis_instrumentation__.Notify(608278)
			if arg == DNull {
				__antithesis_instrumentation__.Notify(608285)
				return DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(608286)
			}
			__antithesis_instrumentation__.Notify(608279)
			argTuple, argIsTuple := arg.(*DTuple)
			if vtuple.Sorted() && func() bool {
				__antithesis_instrumentation__.Notify(608287)
				return !(argIsTuple && func() bool {
					__antithesis_instrumentation__.Notify(608288)
					return argTuple.ContainsNull() == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(608289)

				_, result := vtuple.SearchSorted(ctx, arg)
				return MakeDBool(DBool(result)), nil
			} else {
				__antithesis_instrumentation__.Notify(608290)
			}
			__antithesis_instrumentation__.Notify(608280)

			sawNull := false
			if !argIsTuple {
				__antithesis_instrumentation__.Notify(608291)

				for _, val := range vtuple.D {
					__antithesis_instrumentation__.Notify(608292)
					if val == DNull {
						__antithesis_instrumentation__.Notify(608293)
						sawNull = true
					} else {
						__antithesis_instrumentation__.Notify(608294)
						if val.Compare(ctx, arg) == 0 {
							__antithesis_instrumentation__.Notify(608295)
							return DBoolTrue, nil
						} else {
							__antithesis_instrumentation__.Notify(608296)
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(608297)

				for _, val := range vtuple.D {
					__antithesis_instrumentation__.Notify(608298)
					if val == DNull {
						__antithesis_instrumentation__.Notify(608299)

						sawNull = true
					} else {
						__antithesis_instrumentation__.Notify(608300)

						if res := cmpOpTupleFn(ctx, *argTuple, *val.(*DTuple), treecmp.MakeComparisonOperator(treecmp.EQ)); res == DNull {
							__antithesis_instrumentation__.Notify(608301)
							sawNull = true
						} else {
							__antithesis_instrumentation__.Notify(608302)
							if res == DBoolTrue {
								__antithesis_instrumentation__.Notify(608303)
								return DBoolTrue, nil
							} else {
								__antithesis_instrumentation__.Notify(608304)
							}
						}
					}
				}
			}
			__antithesis_instrumentation__.Notify(608281)
			if sawNull {
				__antithesis_instrumentation__.Notify(608305)
				return DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(608306)
			}
			__antithesis_instrumentation__.Notify(608282)
			return DBoolFalse, nil
		},
		NullableArgs: true,
		Volatility:   v,
	}
}

func evalDatumsCmp(
	ctx *EvalContext, op, subOp treecmp.ComparisonOperator, fn *CmpOp, left Datum, right Datums,
) (Datum, error) {
	__antithesis_instrumentation__.Notify(608307)
	all := op.Symbol == treecmp.All
	any := !all
	sawNull := false
	for _, elem := range right {
		__antithesis_instrumentation__.Notify(608311)
		if elem == DNull {
			__antithesis_instrumentation__.Notify(608315)
			sawNull = true
			continue
		} else {
			__antithesis_instrumentation__.Notify(608316)
		}
		__antithesis_instrumentation__.Notify(608312)

		_, newLeft, newRight, _, not := FoldComparisonExpr(subOp, left, elem)
		d, err := fn.Fn(ctx, newLeft.(Datum), newRight.(Datum))
		if err != nil {
			__antithesis_instrumentation__.Notify(608317)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608318)
		}
		__antithesis_instrumentation__.Notify(608313)
		if d == DNull {
			__antithesis_instrumentation__.Notify(608319)
			sawNull = true
			continue
		} else {
			__antithesis_instrumentation__.Notify(608320)
		}
		__antithesis_instrumentation__.Notify(608314)

		b := d.(*DBool)
		res := *b != DBool(not)
		if any && func() bool {
			__antithesis_instrumentation__.Notify(608321)
			return res == true
		}() == true {
			__antithesis_instrumentation__.Notify(608322)
			return DBoolTrue, nil
		} else {
			__antithesis_instrumentation__.Notify(608323)
			if all && func() bool {
				__antithesis_instrumentation__.Notify(608324)
				return !res == true
			}() == true {
				__antithesis_instrumentation__.Notify(608325)
				return DBoolFalse, nil
			} else {
				__antithesis_instrumentation__.Notify(608326)
			}
		}
	}
	__antithesis_instrumentation__.Notify(608308)

	if sawNull {
		__antithesis_instrumentation__.Notify(608327)

		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608328)
	}
	__antithesis_instrumentation__.Notify(608309)

	if all {
		__antithesis_instrumentation__.Notify(608329)

		return DBoolTrue, nil
	} else {
		__antithesis_instrumentation__.Notify(608330)
	}
	__antithesis_instrumentation__.Notify(608310)

	return DBoolFalse, nil
}

func MatchLikeEscape(
	ctx *EvalContext, unescaped, pattern, escape string, caseInsensitive bool,
) (Datum, error) {
	__antithesis_instrumentation__.Notify(608331)
	var escapeRune rune
	if len(escape) > 0 {
		__antithesis_instrumentation__.Notify(608336)
		var width int
		escapeRune, width = utf8.DecodeRuneInString(escape)
		if len(escape) > width {
			__antithesis_instrumentation__.Notify(608337)
			return DBoolFalse, pgerror.Newf(pgcode.InvalidEscapeSequence, "invalid escape string")
		} else {
			__antithesis_instrumentation__.Notify(608338)
		}
	} else {
		__antithesis_instrumentation__.Notify(608339)
	}
	__antithesis_instrumentation__.Notify(608332)

	if len(unescaped) == 0 {
		__antithesis_instrumentation__.Notify(608340)

		for _, c := range pattern {
			__antithesis_instrumentation__.Notify(608342)
			if c != '%' || func() bool {
				__antithesis_instrumentation__.Notify(608343)
				return (c == '%' && func() bool {
					__antithesis_instrumentation__.Notify(608344)
					return escape == `%` == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(608345)
				return DBoolFalse, nil
			} else {
				__antithesis_instrumentation__.Notify(608346)
			}
		}
		__antithesis_instrumentation__.Notify(608341)
		return DBoolTrue, nil
	} else {
		__antithesis_instrumentation__.Notify(608347)
	}
	__antithesis_instrumentation__.Notify(608333)

	like, err := optimizedLikeFunc(pattern, caseInsensitive, escapeRune)
	if err != nil {
		__antithesis_instrumentation__.Notify(608348)
		return DBoolFalse, pgerror.Wrap(
			err, pgcode.InvalidRegularExpression, "LIKE regexp compilation failed")
	} else {
		__antithesis_instrumentation__.Notify(608349)
	}
	__antithesis_instrumentation__.Notify(608334)

	if like == nil {
		__antithesis_instrumentation__.Notify(608350)
		re, err := ConvertLikeToRegexp(ctx, pattern, caseInsensitive, escapeRune)
		if err != nil {
			__antithesis_instrumentation__.Notify(608352)
			return DBoolFalse, err
		} else {
			__antithesis_instrumentation__.Notify(608353)
		}
		__antithesis_instrumentation__.Notify(608351)
		like = func(s string) (bool, error) {
			__antithesis_instrumentation__.Notify(608354)
			return re.MatchString(s), nil
		}
	} else {
		__antithesis_instrumentation__.Notify(608355)
	}
	__antithesis_instrumentation__.Notify(608335)
	matches, err := like(unescaped)
	return MakeDBool(DBool(matches)), err
}

func ConvertLikeToRegexp(
	ctx *EvalContext, pattern string, caseInsensitive bool, escape rune,
) (*regexp.Regexp, error) {
	__antithesis_instrumentation__.Notify(608356)
	key := likeKey{s: pattern, caseInsensitive: caseInsensitive, escape: escape}
	re, err := ctx.ReCache.GetRegexp(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(608358)
		return nil, pgerror.Wrap(
			err, pgcode.InvalidRegularExpression, "LIKE regexp compilation failed")
	} else {
		__antithesis_instrumentation__.Notify(608359)
	}
	__antithesis_instrumentation__.Notify(608357)
	return re, nil
}

func matchLike(ctx *EvalContext, left, right Datum, caseInsensitive bool) (Datum, error) {
	__antithesis_instrumentation__.Notify(608360)
	if left == DNull || func() bool {
		__antithesis_instrumentation__.Notify(608365)
		return right == DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(608366)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608367)
	}
	__antithesis_instrumentation__.Notify(608361)
	s, pattern := string(MustBeDString(left)), string(MustBeDString(right))
	if len(s) == 0 {
		__antithesis_instrumentation__.Notify(608368)

		for _, c := range pattern {
			__antithesis_instrumentation__.Notify(608370)
			if c != '%' {
				__antithesis_instrumentation__.Notify(608371)
				return DBoolFalse, nil
			} else {
				__antithesis_instrumentation__.Notify(608372)
			}
		}
		__antithesis_instrumentation__.Notify(608369)
		return DBoolTrue, nil
	} else {
		__antithesis_instrumentation__.Notify(608373)
	}
	__antithesis_instrumentation__.Notify(608362)

	like, err := optimizedLikeFunc(pattern, caseInsensitive, '\\')
	if err != nil {
		__antithesis_instrumentation__.Notify(608374)
		return DBoolFalse, pgerror.Wrap(
			err, pgcode.InvalidRegularExpression, "LIKE regexp compilation failed")
	} else {
		__antithesis_instrumentation__.Notify(608375)
	}
	__antithesis_instrumentation__.Notify(608363)

	if like == nil {
		__antithesis_instrumentation__.Notify(608376)
		re, err := ConvertLikeToRegexp(ctx, pattern, caseInsensitive, '\\')
		if err != nil {
			__antithesis_instrumentation__.Notify(608378)
			return DBoolFalse, err
		} else {
			__antithesis_instrumentation__.Notify(608379)
		}
		__antithesis_instrumentation__.Notify(608377)
		like = func(s string) (bool, error) {
			__antithesis_instrumentation__.Notify(608380)
			return re.MatchString(s), nil
		}
	} else {
		__antithesis_instrumentation__.Notify(608381)
	}
	__antithesis_instrumentation__.Notify(608364)
	matches, err := like(s)
	return MakeDBool(DBool(matches)), err
}

func matchRegexpWithKey(ctx *EvalContext, str Datum, key RegexpCacheKey) (Datum, error) {
	__antithesis_instrumentation__.Notify(608382)
	re, err := ctx.ReCache.GetRegexp(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(608384)
		return DBoolFalse, pgerror.Wrap(err, pgcode.InvalidRegularExpression, "invalid regular expression")
	} else {
		__antithesis_instrumentation__.Notify(608385)
	}
	__antithesis_instrumentation__.Notify(608383)
	return MakeDBool(DBool(re.MatchString(string(MustBeDString(str))))), nil
}

type MultipleResultsError struct {
	SQL string
}

func (e *MultipleResultsError) Error() string {
	__antithesis_instrumentation__.Notify(608386)
	return fmt.Sprintf("%s: unexpected multiple results", e.SQL)
}

type DatabaseRegionConfig interface {
	IsValidRegionNameString(r string) bool
	PrimaryRegionString() string
}

type HasAnyPrivilegeResult = int8

const (
	HasPrivilege HasAnyPrivilegeResult = 1

	HasNoPrivilege HasAnyPrivilegeResult = 0

	ObjectNotFound HasAnyPrivilegeResult = -1
)

type EvalDatabase interface {
	ParseQualifiedTableName(sql string) (*TableName, error)

	ResolveTableName(ctx context.Context, tn *TableName) (ID, error)

	SchemaExists(ctx context.Context, dbName, scName string) (found bool, err error)

	IsTableVisible(
		ctx context.Context, curDB string, searchPath sessiondata.SearchPath, tableID oid.Oid,
	) (isVisible bool, exists bool, err error)

	IsTypeVisible(
		ctx context.Context, curDB string, searchPath sessiondata.SearchPath, typeID oid.Oid,
	) (isVisible bool, exists bool, err error)

	HasAnyPrivilege(ctx context.Context, specifier HasPrivilegeSpecifier, user security.SQLUsername, privs []privilege.Privilege) (HasAnyPrivilegeResult, error)
}

type HasPrivilegeSpecifier struct {
	DatabaseName *string
	DatabaseOID  *oid.Oid

	SchemaName *string

	SchemaDatabaseName *string

	SchemaIsRequired *bool

	TableName *string
	TableOID  *oid.Oid

	IsSequence *bool

	ColumnName   *Name
	ColumnAttNum *uint32
}

type TypeResolver interface {
	TypeReferenceResolver

	ResolveOIDFromString(
		ctx context.Context, resultType *types.T, toResolve *DString,
	) (*DOid, error)

	ResolveOIDFromOID(
		ctx context.Context, resultType *types.T, toResolve *DOid,
	) (*DOid, error)
}

type EvalPlanner interface {
	EvalDatabase
	TypeResolver

	ExecutorConfig() interface{}

	GetImmutableTableInterfaceByID(ctx context.Context, id int) (interface{}, error)

	GetTypeFromValidSQLSyntax(sql string) (*types.T, error)

	EvalSubquery(expr *Subquery) (Datum, error)

	UnsafeUpsertDescriptor(
		ctx context.Context, descID int64, encodedDescriptor []byte, force bool,
	) error

	UnsafeDeleteDescriptor(ctx context.Context, descID int64, force bool) error

	ForceDeleteTableData(ctx context.Context, descID int64) error

	UnsafeUpsertNamespaceEntry(
		ctx context.Context,
		parentID, parentSchemaID int64,
		name string,
		descID int64,
		force bool,
	) error

	UnsafeDeleteNamespaceEntry(
		ctx context.Context,
		parentID, parentSchemaID int64,
		name string,
		descID int64,
		force bool,
	) error

	UserHasAdminRole(ctx context.Context, user security.SQLUsername) (bool, error)

	MemberOfWithAdminOption(
		ctx context.Context,
		member security.SQLUsername,
	) (map[security.SQLUsername]bool, error)

	ExternalReadFile(ctx context.Context, uri string) ([]byte, error)

	ExternalWriteFile(ctx context.Context, uri string, content []byte) error

	DecodeGist(gist string) ([]string, error)

	SerializeSessionState() (*DBytes, error)

	DeserializeSessionState(state *DBytes) (*DBool, error)

	CreateSessionRevivalToken() (*DBytes, error)

	ValidateSessionRevivalToken(token *DBytes) (*DBool, error)

	RevalidateUniqueConstraintsInCurrentDB(ctx context.Context) error

	RevalidateUniqueConstraintsInTable(ctx context.Context, tableID int) error

	RevalidateUniqueConstraint(ctx context.Context, tableID int, constraintName string) error

	ValidateTTLScheduledJobsInCurrentDB(ctx context.Context) error

	RepairTTLScheduledJobForTable(ctx context.Context, tableID int64) error

	QueryRowEx(
		ctx context.Context,
		opName string,
		txn *kv.Txn,
		override sessiondata.InternalExecutorOverride,
		stmt string,
		qargs ...interface{}) (Datums, error)

	QueryIteratorEx(
		ctx context.Context,
		opName string,
		txn *kv.Txn,
		override sessiondata.InternalExecutorOverride,
		stmt string,
		qargs ...interface{},
	) (InternalRows, error)
}

type InternalRows interface {
	Next(context.Context) (bool, error)

	Cur() Datums

	Close() error
}

type CompactEngineSpanFunc func(
	ctx context.Context, nodeID, storeID int32, startKey, endKey []byte,
) error

type EvalSessionAccessor interface {
	SetSessionVar(ctx context.Context, settingName, newValue string, isLocal bool) error

	GetSessionVar(ctx context.Context, settingName string, missingOk bool) (bool, string, error)

	HasAdminRole(ctx context.Context) (bool, error)

	HasRoleOption(ctx context.Context, roleOption roleoption.Option) (bool, error)
}

type PreparedStatementState interface {
	HasActivePortals() bool

	MigratablePreparedStatements() []sessiondatapb.MigratableSession_PreparedStatement

	HasPortal(s string) bool
}

type ClientNoticeSender interface {
	BufferClientNotice(ctx context.Context, notice pgnotice.Notice)
}

type PrivilegedAccessor interface {
	LookupNamespaceID(
		ctx context.Context, parentID int64, parentSchemaID int64, name string,
	) (DInt, bool, error)

	LookupZoneConfigByNamespaceID(ctx context.Context, id int64) (DBytes, bool, error)
}

type RegionOperator interface {
	CurrentDatabaseRegionConfig(ctx context.Context) (DatabaseRegionConfig, error)

	ValidateAllMultiRegionZoneConfigsInCurrentDatabase(ctx context.Context) error

	ResetMultiRegionZoneConfigsForTable(ctx context.Context, id int64) error

	ResetMultiRegionZoneConfigsForDatabase(ctx context.Context, id int64) error
}

type SequenceOperators interface {
	GetSerialSequenceNameFromColumn(ctx context.Context, tableName *TableName, columnName Name) (*TableName, error)

	IncrementSequenceByID(ctx context.Context, seqID int64) (int64, error)

	GetLatestValueInSessionForSequenceByID(ctx context.Context, seqID int64) (int64, error)

	SetSequenceValueByID(ctx context.Context, seqID uint32, newVal int64, isCalled bool) error
}

type TenantOperator interface {
	CreateTenant(ctx context.Context, tenantID uint64) error

	DestroyTenant(ctx context.Context, tenantID uint64, synchronous bool) error

	GCTenant(ctx context.Context, tenantID uint64) error

	UpdateTenantResourceLimits(
		ctx context.Context,
		tenantID uint64,
		availableRU float64,
		refillRate float64,
		maxBurstRU float64,
		asOf time.Time,
		asOfConsumedRequestUnits float64,
	) error
}

type JoinTokenCreator interface {
	CreateJoinToken(ctx context.Context) (string, error)
}

type EvalContextTestingKnobs struct {
	AssertFuncExprReturnTypes bool

	AssertUnaryExprReturnTypes bool

	AssertBinaryExprReturnTypes bool

	DisableOptimizerRuleProbability float64

	OptimizerCostPerturbation float64

	ForceProductionBatchSizes bool

	CallbackGenerators map[string]*CallbackValueGenerator
}

var _ base.ModuleTestingKnobs = &EvalContextTestingKnobs{}

func (*EvalContextTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(608387) }

type SQLStatsController interface {
	ResetClusterSQLStats(ctx context.Context) error
	CreateSQLStatsCompactionSchedule(ctx context.Context) error
}

type IndexUsageStatsController interface {
	ResetIndexUsageStats(ctx context.Context) error
}

type EvalContext struct {
	SessionDataStack *sessiondata.Stack

	TxnState string

	TxnReadOnly bool

	TxnImplicit bool

	TxnIsSingleStmt bool

	Settings *cluster.Settings

	ClusterID uuid.UUID

	ClusterName string

	NodeID *base.SQLIDContainer
	Codec  keys.SQLCodec

	Locality roachpb.Locality

	Tracer *tracing.Tracer

	StmtTimestamp time.Time

	TxnTimestamp time.Time

	AsOfSystemTime *AsOfSystemTime

	Placeholders *PlaceholderInfo

	Annotations *Annotations

	IVarContainer IndexedVarContainer

	iVarContainerStack []IndexedVarContainer

	Context context.Context

	Planner EvalPlanner

	JobExecContext interface{}

	PrivilegedAccessor PrivilegedAccessor

	SessionAccessor EvalSessionAccessor

	ClientNoticeSender ClientNoticeSender

	Sequence SequenceOperators

	Tenant TenantOperator

	Regions RegionOperator

	JoinTokenCreator JoinTokenCreator

	PreparedStatementState PreparedStatementState

	Txn *kv.Txn

	DB *kv.DB

	ReCache *RegexpCache

	PrepareOnly bool

	SkipNormalize bool

	CollationEnv CollationEnvironment

	TestingKnobs EvalContextTestingKnobs

	Mon *mon.BytesMonitor

	SingleDatumAggMemAccount *mon.BoundAccount

	SQLLivenessReader sqlliveness.Reader

	SQLStatsController SQLStatsController

	IndexUsageStatsController IndexUsageStatsController

	CompactEngineSpan CompactEngineSpanFunc

	KVStoresIterator kvserverbase.StoresIterator
}

func MakeTestingEvalContext(st *cluster.Settings) EvalContext {
	__antithesis_instrumentation__.Notify(608388)
	monitor := mon.NewMonitor(
		"test-monitor",
		mon.MemoryResource,
		nil,
		nil,
		-1,
		math.MaxInt64,
		st,
	)
	return MakeTestingEvalContextWithMon(st, monitor)
}

func MakeTestingEvalContextWithMon(st *cluster.Settings, monitor *mon.BytesMonitor) EvalContext {
	__antithesis_instrumentation__.Notify(608389)
	ctx := EvalContext{
		Codec:            keys.SystemSQLCodec,
		Txn:              &kv.Txn{},
		SessionDataStack: sessiondata.NewStack(&sessiondata.SessionData{}),
		Settings:         st,
		NodeID:           base.TestingIDContainer,
	}
	monitor.Start(context.Background(), nil, mon.MakeStandaloneBudget(math.MaxInt64))
	ctx.Mon = monitor
	ctx.Context = context.TODO()
	now := timeutil.Now()
	ctx.SetTxnTimestamp(now)
	ctx.SetStmtTimestamp(now)
	return ctx
}

func (ctx *EvalContext) SessionData() *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(608390)
	if ctx.SessionDataStack == nil {
		__antithesis_instrumentation__.Notify(608392)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(608393)
	}
	__antithesis_instrumentation__.Notify(608391)
	return ctx.SessionDataStack.Top()
}

func (ctx *EvalContext) Copy() *EvalContext {
	__antithesis_instrumentation__.Notify(608394)
	ctxCopy := *ctx
	ctxCopy.iVarContainerStack = make([]IndexedVarContainer, len(ctx.iVarContainerStack), cap(ctx.iVarContainerStack))
	copy(ctxCopy.iVarContainerStack, ctx.iVarContainerStack)
	return &ctxCopy
}

func (ctx *EvalContext) PushIVarContainer(c IndexedVarContainer) {
	__antithesis_instrumentation__.Notify(608395)
	ctx.iVarContainerStack = append(ctx.iVarContainerStack, ctx.IVarContainer)
	ctx.IVarContainer = c
}

func (ctx *EvalContext) PopIVarContainer() {
	__antithesis_instrumentation__.Notify(608396)
	ctx.IVarContainer = ctx.iVarContainerStack[len(ctx.iVarContainerStack)-1]
	ctx.iVarContainerStack = ctx.iVarContainerStack[:len(ctx.iVarContainerStack)-1]
}

func (ctx *EvalContext) QualityOfService() sessiondatapb.QoSLevel {
	__antithesis_instrumentation__.Notify(608397)
	if ctx.SessionData() == nil {
		__antithesis_instrumentation__.Notify(608399)
		return sessiondatapb.Normal
	} else {
		__antithesis_instrumentation__.Notify(608400)
	}
	__antithesis_instrumentation__.Notify(608398)
	return ctx.SessionData().DefaultTxnQualityOfService
}

func NewTestingEvalContext(st *cluster.Settings) *EvalContext {
	__antithesis_instrumentation__.Notify(608401)
	ctx := MakeTestingEvalContext(st)
	return &ctx
}

func (ctx *EvalContext) Stop(c context.Context) {
	__antithesis_instrumentation__.Notify(608402)
	if r := recover(); r != nil {
		__antithesis_instrumentation__.Notify(608403)
		ctx.Mon.EmergencyStop(c)
		panic(r)
	} else {
		__antithesis_instrumentation__.Notify(608404)
		ctx.Mon.Stop(c)
	}
}

func (ctx *EvalContext) FmtCtx(f FmtFlags, opts ...FmtCtxOption) *FmtCtx {
	__antithesis_instrumentation__.Notify(608405)
	if ctx.SessionData() != nil {
		__antithesis_instrumentation__.Notify(608407)
		opts = append(
			[]FmtCtxOption{FmtDataConversionConfig(ctx.SessionData().DataConversionConfig)},
			opts...,
		)
	} else {
		__antithesis_instrumentation__.Notify(608408)
	}
	__antithesis_instrumentation__.Notify(608406)
	return NewFmtCtx(
		f,
		opts...,
	)
}

func (ctx *EvalContext) GetStmtTimestamp() time.Time {
	__antithesis_instrumentation__.Notify(608409)

	if !ctx.PrepareOnly && func() bool {
		__antithesis_instrumentation__.Notify(608411)
		return ctx.StmtTimestamp.IsZero() == true
	}() == true {
		__antithesis_instrumentation__.Notify(608412)
		panic(errors.AssertionFailedf("zero statement timestamp in EvalContext"))
	} else {
		__antithesis_instrumentation__.Notify(608413)
	}
	__antithesis_instrumentation__.Notify(608410)
	return ctx.StmtTimestamp
}

func (ctx *EvalContext) GetClusterTimestamp() *DDecimal {
	__antithesis_instrumentation__.Notify(608414)
	ts := ctx.Txn.CommitTimestamp()
	if ts.IsEmpty() {
		__antithesis_instrumentation__.Notify(608416)
		panic(errors.AssertionFailedf("zero cluster timestamp in txn"))
	} else {
		__antithesis_instrumentation__.Notify(608417)
	}
	__antithesis_instrumentation__.Notify(608415)
	return TimestampToDecimalDatum(ts)
}

func (ctx *EvalContext) HasPlaceholders() bool {
	__antithesis_instrumentation__.Notify(608418)
	return ctx.Placeholders != nil
}

const regionKey = "region"

func (ctx *EvalContext) GetLocalRegion() (regionName string, ok bool) {
	__antithesis_instrumentation__.Notify(608419)
	return ctx.Locality.Find(regionKey)
}

func TimestampToDecimal(ts hlc.Timestamp) apd.Decimal {
	__antithesis_instrumentation__.Notify(608420)

	var res apd.Decimal
	val := &res.Coeff
	val.SetInt64(ts.WallTime)
	val.Mul(val, big10E10)
	val.Add(val, apd.NewBigInt(int64(ts.Logical)))

	res.Negative = val.Sign() < 0
	val.Abs(val)

	res.Exponent = -10
	return res
}

func DecimalToInexactDTimestampTZ(d *DDecimal) (*DTimestampTZ, error) {
	__antithesis_instrumentation__.Notify(608421)
	ts, err := decimalToHLC(d)
	if err != nil {
		__antithesis_instrumentation__.Notify(608423)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608424)
	}
	__antithesis_instrumentation__.Notify(608422)
	return MakeDTimestampTZ(timeutil.Unix(0, ts.WallTime), time.Microsecond)
}

func decimalToHLC(d *DDecimal) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(608425)
	var coef apd.BigInt
	coef.Set(&d.Decimal.Coeff)

	coef.Div(&coef, big10E10)
	if !coef.IsInt64() {
		__antithesis_instrumentation__.Notify(608427)
		return hlc.Timestamp{}, pgerror.Newf(
			pgcode.DatetimeFieldOverflow,
			"timestamp value out of range: %s", d.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(608428)
	}
	__antithesis_instrumentation__.Notify(608426)
	return hlc.Timestamp{WallTime: coef.Int64()}, nil
}

func DecimalToInexactDTimestamp(d *DDecimal) (*DTimestamp, error) {
	__antithesis_instrumentation__.Notify(608429)
	ts, err := decimalToHLC(d)
	if err != nil {
		__antithesis_instrumentation__.Notify(608431)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608432)
	}
	__antithesis_instrumentation__.Notify(608430)
	return TimestampToInexactDTimestamp(ts), nil
}

func TimestampToDecimalDatum(ts hlc.Timestamp) *DDecimal {
	__antithesis_instrumentation__.Notify(608433)
	res := TimestampToDecimal(ts)
	return &DDecimal{
		Decimal: res,
	}
}

func TimestampToInexactDTimestamp(ts hlc.Timestamp) *DTimestamp {
	__antithesis_instrumentation__.Notify(608434)
	return MustMakeDTimestamp(timeutil.Unix(0, ts.WallTime), time.Microsecond)
}

func (ctx *EvalContext) GetRelativeParseTime() time.Time {
	__antithesis_instrumentation__.Notify(608435)
	ret := ctx.TxnTimestamp
	if ret.IsZero() {
		__antithesis_instrumentation__.Notify(608437)
		ret = timeutil.Now()
	} else {
		__antithesis_instrumentation__.Notify(608438)
	}
	__antithesis_instrumentation__.Notify(608436)
	return ret.In(ctx.GetLocation())
}

func (ctx *EvalContext) GetTxnTimestamp(precision time.Duration) *DTimestampTZ {
	__antithesis_instrumentation__.Notify(608439)

	if !ctx.PrepareOnly && func() bool {
		__antithesis_instrumentation__.Notify(608441)
		return ctx.TxnTimestamp.IsZero() == true
	}() == true {
		__antithesis_instrumentation__.Notify(608442)
		panic(errors.AssertionFailedf("zero transaction timestamp in EvalContext"))
	} else {
		__antithesis_instrumentation__.Notify(608443)
	}
	__antithesis_instrumentation__.Notify(608440)
	return MustMakeDTimestampTZ(ctx.GetRelativeParseTime(), precision)
}

func (ctx *EvalContext) GetTxnTimestampNoZone(precision time.Duration) *DTimestamp {
	__antithesis_instrumentation__.Notify(608444)

	if !ctx.PrepareOnly && func() bool {
		__antithesis_instrumentation__.Notify(608446)
		return ctx.TxnTimestamp.IsZero() == true
	}() == true {
		__antithesis_instrumentation__.Notify(608447)
		panic(errors.AssertionFailedf("zero transaction timestamp in EvalContext"))
	} else {
		__antithesis_instrumentation__.Notify(608448)
	}
	__antithesis_instrumentation__.Notify(608445)

	t := ctx.GetRelativeParseTime()
	_, offsetSecs := t.Zone()
	return MustMakeDTimestamp(t.Add(time.Second*time.Duration(offsetSecs)).In(time.UTC), precision)
}

func (ctx *EvalContext) GetTxnTime(precision time.Duration) *DTimeTZ {
	__antithesis_instrumentation__.Notify(608449)

	if !ctx.PrepareOnly && func() bool {
		__antithesis_instrumentation__.Notify(608451)
		return ctx.TxnTimestamp.IsZero() == true
	}() == true {
		__antithesis_instrumentation__.Notify(608452)
		panic(errors.AssertionFailedf("zero transaction timestamp in EvalContext"))
	} else {
		__antithesis_instrumentation__.Notify(608453)
	}
	__antithesis_instrumentation__.Notify(608450)
	return NewDTimeTZFromTime(ctx.GetRelativeParseTime().Round(precision))
}

func (ctx *EvalContext) GetTxnTimeNoZone(precision time.Duration) *DTime {
	__antithesis_instrumentation__.Notify(608454)

	if !ctx.PrepareOnly && func() bool {
		__antithesis_instrumentation__.Notify(608456)
		return ctx.TxnTimestamp.IsZero() == true
	}() == true {
		__antithesis_instrumentation__.Notify(608457)
		panic(errors.AssertionFailedf("zero transaction timestamp in EvalContext"))
	} else {
		__antithesis_instrumentation__.Notify(608458)
	}
	__antithesis_instrumentation__.Notify(608455)
	return MakeDTime(timeofday.FromTime(ctx.GetRelativeParseTime().Round(precision)))
}

func (ctx *EvalContext) SetTxnTimestamp(ts time.Time) {
	__antithesis_instrumentation__.Notify(608459)
	ctx.TxnTimestamp = ts
}

func (ctx *EvalContext) SetStmtTimestamp(ts time.Time) {
	__antithesis_instrumentation__.Notify(608460)
	ctx.StmtTimestamp = ts
}

func (ctx *EvalContext) GetLocation() *time.Location {
	__antithesis_instrumentation__.Notify(608461)
	return ctx.SessionData().GetLocation()
}

func (ctx *EvalContext) GetIntervalStyle() duration.IntervalStyle {
	__antithesis_instrumentation__.Notify(608462)
	if ctx.SessionData() == nil {
		__antithesis_instrumentation__.Notify(608464)
		return duration.IntervalStyle_POSTGRES
	} else {
		__antithesis_instrumentation__.Notify(608465)
	}
	__antithesis_instrumentation__.Notify(608463)
	return ctx.SessionData().GetIntervalStyle()
}

func (ctx *EvalContext) GetDateStyle() pgdate.DateStyle {
	__antithesis_instrumentation__.Notify(608466)
	if ctx.SessionData() == nil {
		__antithesis_instrumentation__.Notify(608468)
		return pgdate.DefaultDateStyle()
	} else {
		__antithesis_instrumentation__.Notify(608469)
	}
	__antithesis_instrumentation__.Notify(608467)
	return ctx.SessionData().GetDateStyle()
}

func (ctx *EvalContext) Ctx() context.Context {
	__antithesis_instrumentation__.Notify(608470)
	return ctx.Context
}

func (expr *AndExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608471)
	left, err := expr.Left.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608477)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608478)
	}
	__antithesis_instrumentation__.Notify(608472)
	if left != DNull {
		__antithesis_instrumentation__.Notify(608479)
		if v, err := GetBool(left); err != nil {
			__antithesis_instrumentation__.Notify(608480)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608481)
			if !v {
				__antithesis_instrumentation__.Notify(608482)
				return left, nil
			} else {
				__antithesis_instrumentation__.Notify(608483)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(608484)
	}
	__antithesis_instrumentation__.Notify(608473)
	right, err := expr.Right.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608485)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608486)
	}
	__antithesis_instrumentation__.Notify(608474)
	if right == DNull {
		__antithesis_instrumentation__.Notify(608487)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608488)
	}
	__antithesis_instrumentation__.Notify(608475)
	if v, err := GetBool(right); err != nil {
		__antithesis_instrumentation__.Notify(608489)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608490)
		if !v {
			__antithesis_instrumentation__.Notify(608491)
			return right, nil
		} else {
			__antithesis_instrumentation__.Notify(608492)
		}
	}
	__antithesis_instrumentation__.Notify(608476)
	return left, nil
}

func (expr *BinaryExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608493)
	left, err := expr.Left.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608500)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608501)
	}
	__antithesis_instrumentation__.Notify(608494)
	if left == DNull && func() bool {
		__antithesis_instrumentation__.Notify(608502)
		return !expr.Fn.NullableArgs == true
	}() == true {
		__antithesis_instrumentation__.Notify(608503)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608504)
	}
	__antithesis_instrumentation__.Notify(608495)
	right, err := expr.Right.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608505)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608506)
	}
	__antithesis_instrumentation__.Notify(608496)
	if right == DNull && func() bool {
		__antithesis_instrumentation__.Notify(608507)
		return !expr.Fn.NullableArgs == true
	}() == true {
		__antithesis_instrumentation__.Notify(608508)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608509)
	}
	__antithesis_instrumentation__.Notify(608497)
	res, err := expr.Fn.Fn(ctx, left, right)
	if err != nil {
		__antithesis_instrumentation__.Notify(608510)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608511)
	}
	__antithesis_instrumentation__.Notify(608498)
	if ctx.TestingKnobs.AssertBinaryExprReturnTypes {
		__antithesis_instrumentation__.Notify(608512)
		if err := ensureExpectedType(expr.Fn.ReturnType, res); err != nil {
			__antithesis_instrumentation__.Notify(608513)
			return nil, errors.NewAssertionErrorWithWrappedErrf(err,
				"binary op %q", expr)
		} else {
			__antithesis_instrumentation__.Notify(608514)
		}
	} else {
		__antithesis_instrumentation__.Notify(608515)
	}
	__antithesis_instrumentation__.Notify(608499)
	return res, err
}

func (expr *CaseExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608516)
	if expr.Expr != nil {
		__antithesis_instrumentation__.Notify(608519)

		val, err := expr.Expr.(TypedExpr).Eval(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(608521)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608522)
		}
		__antithesis_instrumentation__.Notify(608520)

		for _, when := range expr.Whens {
			__antithesis_instrumentation__.Notify(608523)
			arg, err := when.Cond.(TypedExpr).Eval(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(608526)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(608527)
			}
			__antithesis_instrumentation__.Notify(608524)
			d, err := evalComparison(ctx, treecmp.MakeComparisonOperator(treecmp.EQ), val, arg)
			if err != nil {
				__antithesis_instrumentation__.Notify(608528)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(608529)
			}
			__antithesis_instrumentation__.Notify(608525)
			if v, err := GetBool(d); err != nil {
				__antithesis_instrumentation__.Notify(608530)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(608531)
				if v {
					__antithesis_instrumentation__.Notify(608532)
					return when.Val.(TypedExpr).Eval(ctx)
				} else {
					__antithesis_instrumentation__.Notify(608533)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(608534)

		for _, when := range expr.Whens {
			__antithesis_instrumentation__.Notify(608535)
			d, err := when.Cond.(TypedExpr).Eval(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(608537)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(608538)
			}
			__antithesis_instrumentation__.Notify(608536)
			if v, err := GetBool(d); err != nil {
				__antithesis_instrumentation__.Notify(608539)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(608540)
				if v {
					__antithesis_instrumentation__.Notify(608541)
					return when.Val.(TypedExpr).Eval(ctx)
				} else {
					__antithesis_instrumentation__.Notify(608542)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(608517)

	if expr.Else != nil {
		__antithesis_instrumentation__.Notify(608543)
		return expr.Else.(TypedExpr).Eval(ctx)
	} else {
		__antithesis_instrumentation__.Notify(608544)
	}
	__antithesis_instrumentation__.Notify(608518)
	return DNull, nil
}

var pgSignatureRegexp = regexp.MustCompile(`^\s*([\w\."]+)\s*\((?:(?:\s*[\w"]+\s*,)*\s*[\w"]+)?\s*\)\s*$`)

func (expr *CastExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608545)
	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608548)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608549)
	}
	__antithesis_instrumentation__.Notify(608546)

	if d == DNull {
		__antithesis_instrumentation__.Notify(608550)
		return d, nil
	} else {
		__antithesis_instrumentation__.Notify(608551)
	}
	__antithesis_instrumentation__.Notify(608547)
	d = UnwrapDatum(ctx, d)
	return PerformCast(ctx, d, expr.ResolvedType())
}

func (expr *IndirectionExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608552)
	var subscriptIdx int
	for i, t := range expr.Indirection {
		__antithesis_instrumentation__.Notify(608558)
		if t.Slice || func() bool {
			__antithesis_instrumentation__.Notify(608562)
			return i > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(608563)
			return nil, errors.AssertionFailedf("unsupported feature should have been rejected during planning")
		} else {
			__antithesis_instrumentation__.Notify(608564)
		}
		__antithesis_instrumentation__.Notify(608559)

		d, err := t.Begin.(TypedExpr).Eval(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(608565)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608566)
		}
		__antithesis_instrumentation__.Notify(608560)
		if d == DNull {
			__antithesis_instrumentation__.Notify(608567)
			return d, nil
		} else {
			__antithesis_instrumentation__.Notify(608568)
		}
		__antithesis_instrumentation__.Notify(608561)
		subscriptIdx = int(MustBeDInt(d))
	}
	__antithesis_instrumentation__.Notify(608553)

	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608569)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608570)
	}
	__antithesis_instrumentation__.Notify(608554)
	if d == DNull {
		__antithesis_instrumentation__.Notify(608571)
		return d, nil
	} else {
		__antithesis_instrumentation__.Notify(608572)
	}
	__antithesis_instrumentation__.Notify(608555)

	arr := MustBeDArray(d)

	switch arr.customOid {
	case oid.T_oidvector, oid.T_int2vector:
		__antithesis_instrumentation__.Notify(608573)
		subscriptIdx++
	default:
		__antithesis_instrumentation__.Notify(608574)
	}
	__antithesis_instrumentation__.Notify(608556)
	if subscriptIdx < 1 || func() bool {
		__antithesis_instrumentation__.Notify(608575)
		return subscriptIdx > arr.Len() == true
	}() == true {
		__antithesis_instrumentation__.Notify(608576)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608577)
	}
	__antithesis_instrumentation__.Notify(608557)
	return arr.Array[subscriptIdx-1], nil
}

func (expr *CollateExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608578)
	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608581)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608582)
	}
	__antithesis_instrumentation__.Notify(608579)
	unwrapped := UnwrapDatum(ctx, d)
	if unwrapped == DNull {
		__antithesis_instrumentation__.Notify(608583)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608584)
	}
	__antithesis_instrumentation__.Notify(608580)
	switch d := unwrapped.(type) {
	case *DString:
		__antithesis_instrumentation__.Notify(608585)
		return NewDCollatedString(string(*d), expr.Locale, &ctx.CollationEnv)
	case *DCollatedString:
		__antithesis_instrumentation__.Notify(608586)
		return NewDCollatedString(d.Contents, expr.Locale, &ctx.CollationEnv)
	default:
		__antithesis_instrumentation__.Notify(608587)
		return nil, pgerror.Newf(pgcode.DatatypeMismatch, "incompatible type for COLLATE: %s", d)
	}
}

func (expr *ColumnAccessExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608588)
	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608591)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608592)
	}
	__antithesis_instrumentation__.Notify(608589)
	if d == DNull {
		__antithesis_instrumentation__.Notify(608593)
		return d, nil
	} else {
		__antithesis_instrumentation__.Notify(608594)
	}
	__antithesis_instrumentation__.Notify(608590)
	return d.(*DTuple).D[expr.ColIndex], nil
}

func (expr *CoalesceExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608595)
	for _, e := range expr.Exprs {
		__antithesis_instrumentation__.Notify(608597)
		d, err := e.(TypedExpr).Eval(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(608599)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608600)
		}
		__antithesis_instrumentation__.Notify(608598)
		if d != DNull {
			__antithesis_instrumentation__.Notify(608601)
			return d, nil
		} else {
			__antithesis_instrumentation__.Notify(608602)
		}
	}
	__antithesis_instrumentation__.Notify(608596)
	return DNull, nil
}

func (expr *ComparisonExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608603)
	left, err := expr.Left.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608610)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608611)
	}
	__antithesis_instrumentation__.Notify(608604)
	right, err := expr.Right.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608612)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608613)
	}
	__antithesis_instrumentation__.Notify(608605)

	op := expr.Operator
	if op.Symbol.HasSubOperator() {
		__antithesis_instrumentation__.Notify(608614)
		return EvalComparisonExprWithSubOperator(ctx, expr, left, right)
	} else {
		__antithesis_instrumentation__.Notify(608615)
	}
	__antithesis_instrumentation__.Notify(608606)

	_, newLeft, newRight, _, not := FoldComparisonExpr(op, left, right)
	if !expr.Fn.NullableArgs && func() bool {
		__antithesis_instrumentation__.Notify(608616)
		return (newLeft == DNull || func() bool {
			__antithesis_instrumentation__.Notify(608617)
			return newRight == DNull == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(608618)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608619)
	}
	__antithesis_instrumentation__.Notify(608607)
	d, err := expr.Fn.Fn(ctx, newLeft.(Datum), newRight.(Datum))
	if d == DNull || func() bool {
		__antithesis_instrumentation__.Notify(608620)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(608621)
		return d, err
	} else {
		__antithesis_instrumentation__.Notify(608622)
	}
	__antithesis_instrumentation__.Notify(608608)
	b, ok := d.(*DBool)
	if !ok {
		__antithesis_instrumentation__.Notify(608623)
		return nil, errors.AssertionFailedf("%v is %T and not *DBool", d, d)
	} else {
		__antithesis_instrumentation__.Notify(608624)
	}
	__antithesis_instrumentation__.Notify(608609)
	return MakeDBool(*b != DBool(not)), nil
}

func EvalComparisonExprWithSubOperator(
	ctx *EvalContext, expr *ComparisonExpr, left, right Datum,
) (Datum, error) {
	__antithesis_instrumentation__.Notify(608625)
	var datums Datums

	if !expr.Fn.NullableArgs && func() bool {
		__antithesis_instrumentation__.Notify(608627)
		return right == DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(608628)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608629)
		if tuple, ok := AsDTuple(right); ok {
			__antithesis_instrumentation__.Notify(608630)
			datums = tuple.D
		} else {
			__antithesis_instrumentation__.Notify(608631)
			if array, ok := AsDArray(right); ok {
				__antithesis_instrumentation__.Notify(608632)
				datums = array.Array
			} else {
				__antithesis_instrumentation__.Notify(608633)
				return nil, errors.AssertionFailedf("unhandled right expression %s", right)
			}
		}
	}
	__antithesis_instrumentation__.Notify(608626)
	return evalDatumsCmp(ctx, expr.Operator, expr.SubOperator, expr.Fn, left, datums)
}

func (expr *FuncExpr) EvalArgsAndGetGenerator(ctx *EvalContext) (ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(608634)
	if expr.fn == nil || func() bool {
		__antithesis_instrumentation__.Notify(608638)
		return expr.fnProps.Class != GeneratorClass == true
	}() == true {
		__antithesis_instrumentation__.Notify(608639)
		return nil, errors.AssertionFailedf("cannot call EvalArgsAndGetGenerator() on non-aggregate function: %q", ErrString(expr))
	} else {
		__antithesis_instrumentation__.Notify(608640)
	}
	__antithesis_instrumentation__.Notify(608635)
	if expr.fn.GeneratorWithExprs != nil {
		__antithesis_instrumentation__.Notify(608641)
		return expr.fn.GeneratorWithExprs(ctx, expr.Exprs)
	} else {
		__antithesis_instrumentation__.Notify(608642)
	}
	__antithesis_instrumentation__.Notify(608636)
	nullArg, args, err := expr.evalArgs(ctx)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(608643)
		return nullArg == true
	}() == true {
		__antithesis_instrumentation__.Notify(608644)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608645)
	}
	__antithesis_instrumentation__.Notify(608637)
	return expr.fn.Generator(ctx, args)
}

func (expr *FuncExpr) evalArgs(ctx *EvalContext) (bool, Datums, error) {
	__antithesis_instrumentation__.Notify(608646)
	args := make(Datums, len(expr.Exprs))
	for i, e := range expr.Exprs {
		__antithesis_instrumentation__.Notify(608648)
		arg, err := e.(TypedExpr).Eval(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(608651)
			return false, nil, err
		} else {
			__antithesis_instrumentation__.Notify(608652)
		}
		__antithesis_instrumentation__.Notify(608649)
		if arg == DNull && func() bool {
			__antithesis_instrumentation__.Notify(608653)
			return !expr.fnProps.NullableArgs == true
		}() == true {
			__antithesis_instrumentation__.Notify(608654)
			return true, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(608655)
		}
		__antithesis_instrumentation__.Notify(608650)
		args[i] = arg
	}
	__antithesis_instrumentation__.Notify(608647)
	return false, args, nil
}

func (expr *FuncExpr) MaybeWrapError(err error) error {
	__antithesis_instrumentation__.Notify(608656)

	fName := expr.Func.String()
	if fName == `crdb_internal.force_error` {
		__antithesis_instrumentation__.Notify(608658)
		return err
	} else {
		__antithesis_instrumentation__.Notify(608659)
	}
	__antithesis_instrumentation__.Notify(608657)

	newErr := errors.Wrapf(err, "%s()", errors.Safe(fName))

	newErr = errors.WithTelemetry(newErr, fName+"()")
	return newErr
}

func (expr *FuncExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608660)
	if expr.fn.FnWithExprs != nil {
		__antithesis_instrumentation__.Notify(608666)
		return expr.fn.FnWithExprs(ctx, expr.Exprs)
	} else {
		__antithesis_instrumentation__.Notify(608667)
	}
	__antithesis_instrumentation__.Notify(608661)

	nullResult, args, err := expr.evalArgs(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608668)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608669)
	}
	__antithesis_instrumentation__.Notify(608662)
	if nullResult {
		__antithesis_instrumentation__.Notify(608670)
		return DNull, err
	} else {
		__antithesis_instrumentation__.Notify(608671)
	}
	__antithesis_instrumentation__.Notify(608663)

	res, err := expr.fn.Fn(ctx, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(608672)
		return nil, expr.MaybeWrapError(err)
	} else {
		__antithesis_instrumentation__.Notify(608673)
	}
	__antithesis_instrumentation__.Notify(608664)
	if ctx.TestingKnobs.AssertFuncExprReturnTypes {
		__antithesis_instrumentation__.Notify(608674)
		if err := ensureExpectedType(expr.fn.FixedReturnType(), res); err != nil {
			__antithesis_instrumentation__.Notify(608675)
			return nil, errors.NewAssertionErrorWithWrappedErrf(err, "function %q", expr)
		} else {
			__antithesis_instrumentation__.Notify(608676)
		}
	} else {
		__antithesis_instrumentation__.Notify(608677)
	}
	__antithesis_instrumentation__.Notify(608665)
	return res, nil
}

func ensureExpectedType(exp *types.T, d Datum) error {
	__antithesis_instrumentation__.Notify(608678)
	if !(exp.Family() == types.AnyFamily || func() bool {
		__antithesis_instrumentation__.Notify(608680)
		return d.ResolvedType().Family() == types.UnknownFamily == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(608681)
		return d.ResolvedType().Equivalent(exp) == true
	}() == true) {
		__antithesis_instrumentation__.Notify(608682)
		return errors.AssertionFailedf(
			"expected return type %q, got: %q", errors.Safe(exp), errors.Safe(d.ResolvedType()))
	} else {
		__antithesis_instrumentation__.Notify(608683)
	}
	__antithesis_instrumentation__.Notify(608679)
	return nil
}

func (expr *IfErrExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608684)
	cond, evalErr := expr.Cond.(TypedExpr).Eval(ctx)
	if evalErr == nil {
		__antithesis_instrumentation__.Notify(608688)
		if expr.Else == nil {
			__antithesis_instrumentation__.Notify(608690)
			return DBoolFalse, nil
		} else {
			__antithesis_instrumentation__.Notify(608691)
		}
		__antithesis_instrumentation__.Notify(608689)
		return cond, nil
	} else {
		__antithesis_instrumentation__.Notify(608692)
	}
	__antithesis_instrumentation__.Notify(608685)
	if expr.ErrCode != nil {
		__antithesis_instrumentation__.Notify(608693)
		errpat, err := expr.ErrCode.(TypedExpr).Eval(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(608696)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608697)
		}
		__antithesis_instrumentation__.Notify(608694)
		if errpat == DNull {
			__antithesis_instrumentation__.Notify(608698)
			return nil, evalErr
		} else {
			__antithesis_instrumentation__.Notify(608699)
		}
		__antithesis_instrumentation__.Notify(608695)
		errpatStr := string(MustBeDString(errpat))
		if code := pgerror.GetPGCode(evalErr); code != pgcode.MakeCode(errpatStr) {
			__antithesis_instrumentation__.Notify(608700)
			return nil, evalErr
		} else {
			__antithesis_instrumentation__.Notify(608701)
		}
	} else {
		__antithesis_instrumentation__.Notify(608702)
	}
	__antithesis_instrumentation__.Notify(608686)
	if expr.Else == nil {
		__antithesis_instrumentation__.Notify(608703)
		return DBoolTrue, nil
	} else {
		__antithesis_instrumentation__.Notify(608704)
	}
	__antithesis_instrumentation__.Notify(608687)
	return expr.Else.(TypedExpr).Eval(ctx)
}

func (expr *IfExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608705)
	cond, err := expr.Cond.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608708)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608709)
	}
	__antithesis_instrumentation__.Notify(608706)
	if cond == DBoolTrue {
		__antithesis_instrumentation__.Notify(608710)
		return expr.True.(TypedExpr).Eval(ctx)
	} else {
		__antithesis_instrumentation__.Notify(608711)
	}
	__antithesis_instrumentation__.Notify(608707)
	return expr.Else.(TypedExpr).Eval(ctx)
}

func (expr *IsOfTypeExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608712)
	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608715)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608716)
	}
	__antithesis_instrumentation__.Notify(608713)
	datumTyp := d.ResolvedType()

	for _, t := range expr.ResolvedTypes() {
		__antithesis_instrumentation__.Notify(608717)
		if datumTyp.Equivalent(t) {
			__antithesis_instrumentation__.Notify(608718)
			return MakeDBool(DBool(!expr.Not)), nil
		} else {
			__antithesis_instrumentation__.Notify(608719)
		}
	}
	__antithesis_instrumentation__.Notify(608714)
	return MakeDBool(DBool(expr.Not)), nil
}

func (expr *NotExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608720)
	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608724)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608725)
	}
	__antithesis_instrumentation__.Notify(608721)
	if d == DNull {
		__antithesis_instrumentation__.Notify(608726)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608727)
	}
	__antithesis_instrumentation__.Notify(608722)
	v, err := GetBool(d)
	if err != nil {
		__antithesis_instrumentation__.Notify(608728)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608729)
	}
	__antithesis_instrumentation__.Notify(608723)
	return MakeDBool(!v), nil
}

func (expr *IsNullExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608730)
	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608734)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608735)
	}
	__antithesis_instrumentation__.Notify(608731)
	if d == DNull {
		__antithesis_instrumentation__.Notify(608736)
		return MakeDBool(true), nil
	} else {
		__antithesis_instrumentation__.Notify(608737)
	}
	__antithesis_instrumentation__.Notify(608732)
	if t, ok := d.(*DTuple); ok {
		__antithesis_instrumentation__.Notify(608738)

		for _, tupleDatum := range t.D {
			__antithesis_instrumentation__.Notify(608740)
			if tupleDatum != DNull {
				__antithesis_instrumentation__.Notify(608741)
				return MakeDBool(false), nil
			} else {
				__antithesis_instrumentation__.Notify(608742)
			}
		}
		__antithesis_instrumentation__.Notify(608739)
		return MakeDBool(true), nil
	} else {
		__antithesis_instrumentation__.Notify(608743)
	}
	__antithesis_instrumentation__.Notify(608733)
	return MakeDBool(false), nil
}

func (expr *IsNotNullExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608744)
	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608748)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608749)
	}
	__antithesis_instrumentation__.Notify(608745)
	if d == DNull {
		__antithesis_instrumentation__.Notify(608750)
		return MakeDBool(false), nil
	} else {
		__antithesis_instrumentation__.Notify(608751)
	}
	__antithesis_instrumentation__.Notify(608746)
	if t, ok := d.(*DTuple); ok {
		__antithesis_instrumentation__.Notify(608752)

		for _, tupleDatum := range t.D {
			__antithesis_instrumentation__.Notify(608754)
			if tupleDatum == DNull {
				__antithesis_instrumentation__.Notify(608755)
				return MakeDBool(false), nil
			} else {
				__antithesis_instrumentation__.Notify(608756)
			}
		}
		__antithesis_instrumentation__.Notify(608753)
		return MakeDBool(true), nil
	} else {
		__antithesis_instrumentation__.Notify(608757)
	}
	__antithesis_instrumentation__.Notify(608747)
	return MakeDBool(true), nil
}

func (expr *NullIfExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608758)
	expr1, err := expr.Expr1.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608763)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608764)
	}
	__antithesis_instrumentation__.Notify(608759)
	expr2, err := expr.Expr2.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608765)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608766)
	}
	__antithesis_instrumentation__.Notify(608760)
	cond, err := evalComparison(ctx, treecmp.MakeComparisonOperator(treecmp.EQ), expr1, expr2)
	if err != nil {
		__antithesis_instrumentation__.Notify(608767)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608768)
	}
	__antithesis_instrumentation__.Notify(608761)
	if cond == DBoolTrue {
		__antithesis_instrumentation__.Notify(608769)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608770)
	}
	__antithesis_instrumentation__.Notify(608762)
	return expr1, nil
}

func (expr *OrExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608771)
	left, err := expr.Left.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608778)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608779)
	}
	__antithesis_instrumentation__.Notify(608772)
	if left != DNull {
		__antithesis_instrumentation__.Notify(608780)
		if v, err := GetBool(left); err != nil {
			__antithesis_instrumentation__.Notify(608781)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608782)
			if v {
				__antithesis_instrumentation__.Notify(608783)
				return left, nil
			} else {
				__antithesis_instrumentation__.Notify(608784)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(608785)
	}
	__antithesis_instrumentation__.Notify(608773)
	right, err := expr.Right.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608786)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608787)
	}
	__antithesis_instrumentation__.Notify(608774)
	if right == DNull {
		__antithesis_instrumentation__.Notify(608788)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608789)
	}
	__antithesis_instrumentation__.Notify(608775)
	if v, err := GetBool(right); err != nil {
		__antithesis_instrumentation__.Notify(608790)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608791)
		if v {
			__antithesis_instrumentation__.Notify(608792)
			return right, nil
		} else {
			__antithesis_instrumentation__.Notify(608793)
		}
	}
	__antithesis_instrumentation__.Notify(608776)
	if left == DNull {
		__antithesis_instrumentation__.Notify(608794)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608795)
	}
	__antithesis_instrumentation__.Notify(608777)
	return DBoolFalse, nil
}

func (expr *ParenExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608796)
	return expr.Expr.(TypedExpr).Eval(ctx)
}

func (expr *RangeCond) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608797)
	return nil, errors.AssertionFailedf("unhandled type %T", expr)
}

func (expr *UnaryExpr) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608798)
	d, err := expr.Expr.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608803)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608804)
	}
	__antithesis_instrumentation__.Notify(608799)
	if d == DNull {
		__antithesis_instrumentation__.Notify(608805)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608806)
	}
	__antithesis_instrumentation__.Notify(608800)
	res, err := expr.fn.Fn(ctx, d)
	if err != nil {
		__antithesis_instrumentation__.Notify(608807)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608808)
	}
	__antithesis_instrumentation__.Notify(608801)
	if ctx.TestingKnobs.AssertUnaryExprReturnTypes {
		__antithesis_instrumentation__.Notify(608809)
		if err := ensureExpectedType(expr.fn.ReturnType, res); err != nil {
			__antithesis_instrumentation__.Notify(608810)
			return nil, errors.NewAssertionErrorWithWrappedErrf(err, "unary op %q", expr)
		} else {
			__antithesis_instrumentation__.Notify(608811)
		}
	} else {
		__antithesis_instrumentation__.Notify(608812)
	}
	__antithesis_instrumentation__.Notify(608802)
	return res, err
}

func (expr DefaultVal) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608813)
	return nil, errors.AssertionFailedf("unhandled type %T", expr)
}

func (expr UnqualifiedStar) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608814)
	return nil, errors.AssertionFailedf("unhandled type %T", expr)
}

func (expr *UnresolvedName) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608815)
	return nil, errors.AssertionFailedf("unhandled type %T", expr)
}

func (expr *AllColumnsSelector) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608816)
	return nil, errors.AssertionFailedf("unhandled type %T", expr)
}

func (expr *TupleStar) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608817)
	return nil, errors.AssertionFailedf("unhandled type %T", expr)
}

func (expr *ColumnItem) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608818)
	return nil, errors.AssertionFailedf("unhandled type %T", expr)
}

func (t *Tuple) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608819)
	tuple := NewDTupleWithLen(t.typ, len(t.Exprs))
	for i, v := range t.Exprs {
		__antithesis_instrumentation__.Notify(608821)
		d, err := v.(TypedExpr).Eval(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(608823)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608824)
		}
		__antithesis_instrumentation__.Notify(608822)
		tuple.D[i] = d
	}
	__antithesis_instrumentation__.Notify(608820)
	return tuple, nil
}

func arrayOfType(typ *types.T) (*DArray, error) {
	__antithesis_instrumentation__.Notify(608825)
	if typ.Family() != types.ArrayFamily {
		__antithesis_instrumentation__.Notify(608828)
		return nil, errors.AssertionFailedf("array node type (%v) is not types.TArray", typ)
	} else {
		__antithesis_instrumentation__.Notify(608829)
	}
	__antithesis_instrumentation__.Notify(608826)
	if err := types.CheckArrayElementType(typ.ArrayContents()); err != nil {
		__antithesis_instrumentation__.Notify(608830)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608831)
	}
	__antithesis_instrumentation__.Notify(608827)
	return NewDArray(typ.ArrayContents()), nil
}

func (t *Array) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608832)
	array, err := arrayOfType(t.ResolvedType())
	if err != nil {
		__antithesis_instrumentation__.Notify(608835)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608836)
	}
	__antithesis_instrumentation__.Notify(608833)

	for _, v := range t.Exprs {
		__antithesis_instrumentation__.Notify(608837)
		d, err := v.(TypedExpr).Eval(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(608839)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608840)
		}
		__antithesis_instrumentation__.Notify(608838)
		if err := array.Append(d); err != nil {
			__antithesis_instrumentation__.Notify(608841)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(608842)
		}
	}
	__antithesis_instrumentation__.Notify(608834)
	return array, nil
}

func (expr *Subquery) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608843)
	return ctx.Planner.EvalSubquery(expr)
}

func (t *ArrayFlatten) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608844)
	array, err := arrayOfType(t.ResolvedType())
	if err != nil {
		__antithesis_instrumentation__.Notify(608848)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608849)
	}
	__antithesis_instrumentation__.Notify(608845)

	d, err := t.Subquery.(TypedExpr).Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(608850)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(608851)
	}
	__antithesis_instrumentation__.Notify(608846)

	tuple, ok := d.(*DTuple)
	if !ok {
		__antithesis_instrumentation__.Notify(608852)
		return nil, errors.AssertionFailedf("array subquery result (%v) is not DTuple", d)
	} else {
		__antithesis_instrumentation__.Notify(608853)
	}
	__antithesis_instrumentation__.Notify(608847)
	array.Array = tuple.D
	return array, nil
}

func (t *DBitArray) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608854)
	return t, nil
}

func (t *DBool) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608855)
	return t, nil
}

func (t *DBytes) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608856)
	return t, nil
}

func (t *DUuid) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608857)
	return t, nil
}

func (t *DIPAddr) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608858)
	return t, nil
}

func (t *DDate) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608859)
	return t, nil
}

func (t *DTime) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608860)
	return t, nil
}

func (t *DTimeTZ) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608861)
	return t, nil
}

func (t *DFloat) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608862)
	return t, nil
}

func (t *DDecimal) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608863)
	return t, nil
}

func (t *DInt) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608864)
	return t, nil
}

func (t *DInterval) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608865)
	return t, nil
}

func (t *DBox2D) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608866)
	return t, nil
}

func (t *DGeography) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608867)
	return t, nil
}

func (t *DGeometry) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608868)
	return t, nil
}

func (t *DEnum) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608869)
	return t, nil
}

func (t *DJSON) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608870)
	return t, nil
}

func (t dNull) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608871)
	return t, nil
}

func (t *DString) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608872)
	return t, nil
}

func (t *DCollatedString) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608873)
	return t, nil
}

func (t *DTimestamp) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608874)
	return t, nil
}

func (t *DTimestampTZ) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608875)
	return t, nil
}

func (t *DTuple) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608876)
	return t, nil
}

func (t *DArray) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608877)
	return t, nil
}

func (t *DVoid) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608878)
	return t, nil
}

func (t *DOid) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608879)
	return t, nil
}

func (t *DOidWrapper) Eval(_ *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608880)
	return t, nil
}

func makeNoValueProvidedForPlaceholderErr(pIdx PlaceholderIdx) error {
	__antithesis_instrumentation__.Notify(608881)
	return pgerror.Newf(pgcode.UndefinedParameter,
		"no value provided for placeholder: $%d", pIdx+1,
	)
}

func (t *Placeholder) Eval(ctx *EvalContext) (Datum, error) {
	__antithesis_instrumentation__.Notify(608882)
	if !ctx.HasPlaceholders() {
		__antithesis_instrumentation__.Notify(608887)

		return t, nil
	} else {
		__antithesis_instrumentation__.Notify(608888)
	}
	__antithesis_instrumentation__.Notify(608883)
	e, ok := ctx.Placeholders.Value(t.Idx)
	if !ok {
		__antithesis_instrumentation__.Notify(608889)
		return nil, makeNoValueProvidedForPlaceholderErr(t.Idx)
	} else {
		__antithesis_instrumentation__.Notify(608890)
	}
	__antithesis_instrumentation__.Notify(608884)

	typ := ctx.Placeholders.Types[t.Idx]
	if typ == nil {
		__antithesis_instrumentation__.Notify(608891)

		return nil, errors.AssertionFailedf("missing type for placeholder %s", t)
	} else {
		__antithesis_instrumentation__.Notify(608892)
	}
	__antithesis_instrumentation__.Notify(608885)
	if !e.ResolvedType().Equivalent(typ) {
		__antithesis_instrumentation__.Notify(608893)

		cast := NewTypedCastExpr(e, typ)
		return cast.Eval(ctx)
	} else {
		__antithesis_instrumentation__.Notify(608894)
	}
	__antithesis_instrumentation__.Notify(608886)
	return e.Eval(ctx)
}

func evalComparison(
	ctx *EvalContext, op treecmp.ComparisonOperator, left, right Datum,
) (Datum, error) {
	__antithesis_instrumentation__.Notify(608895)
	if left == DNull || func() bool {
		__antithesis_instrumentation__.Notify(608898)
		return right == DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(608899)
		return DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(608900)
	}
	__antithesis_instrumentation__.Notify(608896)
	ltype := left.ResolvedType()
	rtype := right.ResolvedType()
	if fn, ok := CmpOps[op.Symbol].LookupImpl(ltype, rtype); ok {
		__antithesis_instrumentation__.Notify(608901)
		return fn.Fn(ctx, left, right)
	} else {
		__antithesis_instrumentation__.Notify(608902)
	}
	__antithesis_instrumentation__.Notify(608897)
	return nil, pgerror.Newf(
		pgcode.UndefinedFunction, "unsupported comparison operator: <%s> %s <%s>", ltype, op, rtype)
}

func FoldComparisonExpr(
	op treecmp.ComparisonOperator, left, right Expr,
) (newOp treecmp.ComparisonOperator, newLeft Expr, newRight Expr, flipped bool, not bool) {
	__antithesis_instrumentation__.Notify(608903)
	switch op.Symbol {
	case treecmp.NE:
		__antithesis_instrumentation__.Notify(608905)

		return treecmp.MakeComparisonOperator(treecmp.EQ), left, right, false, true
	case treecmp.GT:
		__antithesis_instrumentation__.Notify(608906)

		return treecmp.MakeComparisonOperator(treecmp.LT), right, left, true, false
	case treecmp.GE:
		__antithesis_instrumentation__.Notify(608907)

		return treecmp.MakeComparisonOperator(treecmp.LE), right, left, true, false
	case treecmp.NotIn:
		__antithesis_instrumentation__.Notify(608908)

		return treecmp.MakeComparisonOperator(treecmp.In), left, right, false, true
	case treecmp.NotLike:
		__antithesis_instrumentation__.Notify(608909)

		return treecmp.MakeComparisonOperator(treecmp.Like), left, right, false, true
	case treecmp.NotILike:
		__antithesis_instrumentation__.Notify(608910)

		return treecmp.MakeComparisonOperator(treecmp.ILike), left, right, false, true
	case treecmp.NotSimilarTo:
		__antithesis_instrumentation__.Notify(608911)

		return treecmp.MakeComparisonOperator(treecmp.SimilarTo), left, right, false, true
	case treecmp.NotRegMatch:
		__antithesis_instrumentation__.Notify(608912)

		return treecmp.MakeComparisonOperator(treecmp.RegMatch), left, right, false, true
	case treecmp.NotRegIMatch:
		__antithesis_instrumentation__.Notify(608913)

		return treecmp.MakeComparisonOperator(treecmp.RegIMatch), left, right, false, true
	case treecmp.IsDistinctFrom:
		__antithesis_instrumentation__.Notify(608914)

		return treecmp.MakeComparisonOperator(treecmp.IsNotDistinctFrom), left, right, false, true
	default:
		__antithesis_instrumentation__.Notify(608915)
	}
	__antithesis_instrumentation__.Notify(608904)
	return op, left, right, false, false
}

func hasUnescapedSuffix(s string, suffix byte, escapeToken string) bool {
	__antithesis_instrumentation__.Notify(608916)
	if s[len(s)-1] == suffix {
		__antithesis_instrumentation__.Notify(608918)
		var count int
		idx := len(s) - len(escapeToken) - 1
		for idx >= 0 && func() bool {
			__antithesis_instrumentation__.Notify(608920)
			return s[idx:idx+len(escapeToken)] == escapeToken == true
		}() == true {
			__antithesis_instrumentation__.Notify(608921)
			count++
			idx -= len(escapeToken)
		}
		__antithesis_instrumentation__.Notify(608919)
		return count%2 == 0
	} else {
		__antithesis_instrumentation__.Notify(608922)
	}
	__antithesis_instrumentation__.Notify(608917)
	return false
}

func optimizedLikeFunc(
	pattern string, caseInsensitive bool, escape rune,
) (func(string) (bool, error), error) {
	__antithesis_instrumentation__.Notify(608923)
	switch len(pattern) {
	case 0:
		__antithesis_instrumentation__.Notify(608925)
		return func(s string) (bool, error) {
			__antithesis_instrumentation__.Notify(608928)
			return s == "", nil
		}, nil
	case 1:
		__antithesis_instrumentation__.Notify(608926)
		switch pattern[0] {
		case '%':
			__antithesis_instrumentation__.Notify(608929)
			if escape == '%' {
				__antithesis_instrumentation__.Notify(608934)
				return nil, pgerror.Newf(pgcode.InvalidEscapeSequence, "LIKE pattern must not end with escape character")
			} else {
				__antithesis_instrumentation__.Notify(608935)
			}
			__antithesis_instrumentation__.Notify(608930)
			return func(s string) (bool, error) {
				__antithesis_instrumentation__.Notify(608936)
				return true, nil
			}, nil
		case '_':
			__antithesis_instrumentation__.Notify(608931)
			if escape == '_' {
				__antithesis_instrumentation__.Notify(608937)
				return nil, pgerror.Newf(pgcode.InvalidEscapeSequence, "LIKE pattern must not end with escape character")
			} else {
				__antithesis_instrumentation__.Notify(608938)
			}
			__antithesis_instrumentation__.Notify(608932)
			return func(s string) (bool, error) {
				__antithesis_instrumentation__.Notify(608939)
				if len(s) == 0 {
					__antithesis_instrumentation__.Notify(608942)
					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(608943)
				}
				__antithesis_instrumentation__.Notify(608940)
				firstChar, _ := utf8.DecodeRuneInString(s)
				if firstChar == utf8.RuneError {
					__antithesis_instrumentation__.Notify(608944)
					return false, errors.Errorf("invalid encoding of the first character in string %s", s)
				} else {
					__antithesis_instrumentation__.Notify(608945)
				}
				__antithesis_instrumentation__.Notify(608941)
				return len(s) == len(string(firstChar)), nil
			}, nil
		default:
			__antithesis_instrumentation__.Notify(608933)
		}
	default:
		__antithesis_instrumentation__.Notify(608927)
		if !strings.ContainsAny(pattern[1:len(pattern)-1], "_%") {
			__antithesis_instrumentation__.Notify(608946)

			anyEnd := hasUnescapedSuffix(pattern, '%', string(escape)) && func() bool {
				__antithesis_instrumentation__.Notify(608948)
				return escape != '%' == true
			}() == true

			anyStart := pattern[0] == '%' && func() bool {
				__antithesis_instrumentation__.Notify(608949)
				return escape != '%' == true
			}() == true

			singleAnyEnd := hasUnescapedSuffix(pattern, '_', string(escape)) && func() bool {
				__antithesis_instrumentation__.Notify(608950)
				return escape != '_' == true
			}() == true

			singleAnyStart := pattern[0] == '_' && func() bool {
				__antithesis_instrumentation__.Notify(608951)
				return escape != '_' == true
			}() == true

			var err error
			if pattern, err = unescapePattern(pattern, string(escape), true); err != nil {
				__antithesis_instrumentation__.Notify(608952)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(608953)
			}
			__antithesis_instrumentation__.Notify(608947)
			switch {
			case anyEnd && func() bool {
				__antithesis_instrumentation__.Notify(608959)
				return anyStart == true
			}() == true:
				__antithesis_instrumentation__.Notify(608954)
				return func(s string) (bool, error) {
					__antithesis_instrumentation__.Notify(608960)
					substr := pattern[1 : len(pattern)-1]
					if caseInsensitive {
						__antithesis_instrumentation__.Notify(608962)
						s, substr = strings.ToUpper(s), strings.ToUpper(substr)
					} else {
						__antithesis_instrumentation__.Notify(608963)
					}
					__antithesis_instrumentation__.Notify(608961)
					return strings.Contains(s, substr), nil
				}, nil

			case anyEnd:
				__antithesis_instrumentation__.Notify(608955)
				return func(s string) (bool, error) {
					__antithesis_instrumentation__.Notify(608964)
					prefix := pattern[:len(pattern)-1]
					if singleAnyStart {
						__antithesis_instrumentation__.Notify(608967)
						if len(s) == 0 {
							__antithesis_instrumentation__.Notify(608970)
							return false, nil
						} else {
							__antithesis_instrumentation__.Notify(608971)
						}
						__antithesis_instrumentation__.Notify(608968)
						prefix = prefix[1:]
						firstChar, _ := utf8.DecodeRuneInString(s)
						if firstChar == utf8.RuneError {
							__antithesis_instrumentation__.Notify(608972)
							return false, errors.Errorf("invalid encoding of the first character in string %s", s)
						} else {
							__antithesis_instrumentation__.Notify(608973)
						}
						__antithesis_instrumentation__.Notify(608969)
						s = s[len(string(firstChar)):]
					} else {
						__antithesis_instrumentation__.Notify(608974)
					}
					__antithesis_instrumentation__.Notify(608965)
					if caseInsensitive {
						__antithesis_instrumentation__.Notify(608975)
						s, prefix = strings.ToUpper(s), strings.ToUpper(prefix)
					} else {
						__antithesis_instrumentation__.Notify(608976)
					}
					__antithesis_instrumentation__.Notify(608966)
					return strings.HasPrefix(s, prefix), nil
				}, nil

			case anyStart:
				__antithesis_instrumentation__.Notify(608956)
				return func(s string) (bool, error) {
					__antithesis_instrumentation__.Notify(608977)
					suffix := pattern[1:]
					if singleAnyEnd {
						__antithesis_instrumentation__.Notify(608980)
						if len(s) == 0 {
							__antithesis_instrumentation__.Notify(608983)
							return false, nil
						} else {
							__antithesis_instrumentation__.Notify(608984)
						}
						__antithesis_instrumentation__.Notify(608981)

						suffix = suffix[:len(suffix)-1]
						lastChar, _ := utf8.DecodeLastRuneInString(s)
						if lastChar == utf8.RuneError {
							__antithesis_instrumentation__.Notify(608985)
							return false, errors.Errorf("invalid encoding of the last character in string %s", s)
						} else {
							__antithesis_instrumentation__.Notify(608986)
						}
						__antithesis_instrumentation__.Notify(608982)
						s = s[:len(s)-len(string(lastChar))]
					} else {
						__antithesis_instrumentation__.Notify(608987)
					}
					__antithesis_instrumentation__.Notify(608978)
					if caseInsensitive {
						__antithesis_instrumentation__.Notify(608988)
						s, suffix = strings.ToUpper(s), strings.ToUpper(suffix)
					} else {
						__antithesis_instrumentation__.Notify(608989)
					}
					__antithesis_instrumentation__.Notify(608979)
					return strings.HasSuffix(s, suffix), nil
				}, nil

			case singleAnyStart || func() bool {
				__antithesis_instrumentation__.Notify(608990)
				return singleAnyEnd == true
			}() == true:
				__antithesis_instrumentation__.Notify(608957)
				return func(s string) (bool, error) {
					__antithesis_instrumentation__.Notify(608991)
					if len(s) < 1 {
						__antithesis_instrumentation__.Notify(608999)
						return false, nil
					} else {
						__antithesis_instrumentation__.Notify(609000)
					}
					__antithesis_instrumentation__.Notify(608992)
					firstChar, _ := utf8.DecodeRuneInString(s)
					if firstChar == utf8.RuneError {
						__antithesis_instrumentation__.Notify(609001)
						return false, errors.Errorf("invalid encoding of the first character in string %s", s)
					} else {
						__antithesis_instrumentation__.Notify(609002)
					}
					__antithesis_instrumentation__.Notify(608993)
					lastChar, _ := utf8.DecodeLastRuneInString(s)
					if lastChar == utf8.RuneError {
						__antithesis_instrumentation__.Notify(609003)
						return false, errors.Errorf("invalid encoding of the last character in string %s", s)
					} else {
						__antithesis_instrumentation__.Notify(609004)
					}
					__antithesis_instrumentation__.Notify(608994)
					if singleAnyStart && func() bool {
						__antithesis_instrumentation__.Notify(609005)
						return singleAnyEnd == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(609006)
						return len(string(firstChar))+len(string(lastChar)) > len(s) == true
					}() == true {
						__antithesis_instrumentation__.Notify(609007)
						return false, nil
					} else {
						__antithesis_instrumentation__.Notify(609008)
					}
					__antithesis_instrumentation__.Notify(608995)

					if singleAnyStart {
						__antithesis_instrumentation__.Notify(609009)
						pattern = pattern[1:]
						s = s[len(string(firstChar)):]
					} else {
						__antithesis_instrumentation__.Notify(609010)
					}
					__antithesis_instrumentation__.Notify(608996)

					if singleAnyEnd {
						__antithesis_instrumentation__.Notify(609011)
						pattern = pattern[:len(pattern)-1]
						s = s[:len(s)-len(string(lastChar))]
					} else {
						__antithesis_instrumentation__.Notify(609012)
					}
					__antithesis_instrumentation__.Notify(608997)

					if caseInsensitive {
						__antithesis_instrumentation__.Notify(609013)
						s, pattern = strings.ToUpper(s), strings.ToUpper(pattern)
					} else {
						__antithesis_instrumentation__.Notify(609014)
					}
					__antithesis_instrumentation__.Notify(608998)

					return s == pattern, nil
				}, nil
			default:
				__antithesis_instrumentation__.Notify(608958)
			}
		} else {
			__antithesis_instrumentation__.Notify(609015)
		}
	}
	__antithesis_instrumentation__.Notify(608924)
	return nil, nil
}

type likeKey struct {
	s               string
	caseInsensitive bool
	escape          rune
}

func LikeEscape(pattern string) (string, error) {
	__antithesis_instrumentation__.Notify(609016)
	key := likeKey{s: pattern, caseInsensitive: false, escape: '\\'}
	re, err := key.patternNoAnchor()
	return re, err
}

func unescapePattern(
	pattern, escapeToken string, emitEscapeCharacterLastError bool,
) (string, error) {
	__antithesis_instrumentation__.Notify(609017)
	escapedEscapeToken := escapeToken + escapeToken

	nEscapes := strings.Count(pattern, escapeToken) - strings.Count(pattern, escapedEscapeToken)
	if nEscapes == 0 {
		__antithesis_instrumentation__.Notify(609020)
		return pattern, nil
	} else {
		__antithesis_instrumentation__.Notify(609021)
	}
	__antithesis_instrumentation__.Notify(609018)

	ret := make([]byte, len(pattern)-nEscapes*len(escapeToken))
	retWidth := 0
	for i := 0; i < nEscapes; i++ {
		__antithesis_instrumentation__.Notify(609022)
		nextIdx := strings.Index(pattern, escapeToken)
		if nextIdx == len(pattern)-len(escapeToken) && func() bool {
			__antithesis_instrumentation__.Notify(609025)
			return emitEscapeCharacterLastError == true
		}() == true {
			__antithesis_instrumentation__.Notify(609026)
			return "", pgerror.Newf(pgcode.InvalidEscapeSequence, `LIKE pattern must not end with escape character`)
		} else {
			__antithesis_instrumentation__.Notify(609027)
		}
		__antithesis_instrumentation__.Notify(609023)

		retWidth += copy(ret[retWidth:], pattern[:nextIdx])

		if nextIdx < len(pattern)-len(escapedEscapeToken) && func() bool {
			__antithesis_instrumentation__.Notify(609028)
			return pattern[nextIdx:nextIdx+len(escapedEscapeToken)] == escapedEscapeToken == true
		}() == true {
			__antithesis_instrumentation__.Notify(609029)

			retWidth += copy(ret[retWidth:], escapeToken)
			pattern = pattern[nextIdx+len(escapedEscapeToken):]
			continue
		} else {
			__antithesis_instrumentation__.Notify(609030)
		}
		__antithesis_instrumentation__.Notify(609024)

		pattern = pattern[nextIdx+len(escapeToken):]
	}
	__antithesis_instrumentation__.Notify(609019)

	retWidth += copy(ret[retWidth:], pattern)
	return string(ret[0:retWidth]), nil
}

func replaceUnescaped(s, oldStr, newStr string, escapeToken string) string {
	__antithesis_instrumentation__.Notify(609031)

	nOld := strings.Count(s, oldStr)
	if nOld == 0 {
		__antithesis_instrumentation__.Notify(609035)
		return s
	} else {
		__antithesis_instrumentation__.Notify(609036)
	}
	__antithesis_instrumentation__.Notify(609032)

	retLen := len(s)

	if addnBytes := nOld * (len(newStr) - len(oldStr)); addnBytes > 0 {
		__antithesis_instrumentation__.Notify(609037)
		retLen += addnBytes
	} else {
		__antithesis_instrumentation__.Notify(609038)
	}
	__antithesis_instrumentation__.Notify(609033)
	ret := make([]byte, retLen)
	retWidth := 0
	start := 0
OldLoop:
	for i := 0; i < nOld; i++ {
		__antithesis_instrumentation__.Notify(609039)
		nextIdx := start + strings.Index(s[start:], oldStr)

		escaped := false
		for {
			__antithesis_instrumentation__.Notify(609041)

			curIdx := nextIdx
			lookbehindIdx := curIdx - len(escapeToken)
			for lookbehindIdx >= 0 && func() bool {
				__antithesis_instrumentation__.Notify(609044)
				return s[lookbehindIdx:curIdx] == escapeToken == true
			}() == true {
				__antithesis_instrumentation__.Notify(609045)
				escaped = !escaped
				curIdx = lookbehindIdx
				lookbehindIdx = curIdx - len(escapeToken)
			}
			__antithesis_instrumentation__.Notify(609042)

			if !escaped {
				__antithesis_instrumentation__.Notify(609046)
				break
			} else {
				__antithesis_instrumentation__.Notify(609047)
			}
			__antithesis_instrumentation__.Notify(609043)

			retWidth += copy(ret[retWidth:], s[start:nextIdx+len(oldStr)])
			start = nextIdx + len(oldStr)

			continue OldLoop
		}
		__antithesis_instrumentation__.Notify(609040)

		retWidth += copy(ret[retWidth:], s[start:nextIdx])
		retWidth += copy(ret[retWidth:], newStr)
		start = nextIdx + len(oldStr)
	}
	__antithesis_instrumentation__.Notify(609034)

	retWidth += copy(ret[retWidth:], s[start:])
	return string(ret[0:retWidth])
}

func replaceCustomEscape(s string, escape rune) (string, error) {
	__antithesis_instrumentation__.Notify(609048)
	changed, retLen, err := calculateLengthAfterReplacingCustomEscape(s, escape)
	if err != nil {
		__antithesis_instrumentation__.Notify(609052)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(609053)
	}
	__antithesis_instrumentation__.Notify(609049)
	if !changed {
		__antithesis_instrumentation__.Notify(609054)
		return s, nil
	} else {
		__antithesis_instrumentation__.Notify(609055)
	}
	__antithesis_instrumentation__.Notify(609050)

	sLen := len(s)
	ret := make([]byte, retLen)
	retIndex, sIndex := 0, 0
	for retIndex < retLen {
		__antithesis_instrumentation__.Notify(609056)
		sRune, w := utf8.DecodeRuneInString(s[sIndex:])
		if sRune == escape {
			__antithesis_instrumentation__.Notify(609057)

			if sIndex+w < sLen {
				__antithesis_instrumentation__.Notify(609058)

				tRune, _ := utf8.DecodeRuneInString(s[(sIndex + w):])
				if tRune == escape {
					__antithesis_instrumentation__.Notify(609059)

					utf8.EncodeRune(ret[retIndex:], escape)
					retIndex += w
					sIndex += 2 * w
				} else {
					__antithesis_instrumentation__.Notify(609060)

					ret[retIndex] = '\\'
					ret[retIndex+1] = '\\'
					retIndex += 2
					sIndex += w
				}
			} else {
				__antithesis_instrumentation__.Notify(609061)

				return "", errors.AssertionFailedf(
					"unexpected: escape character is the last one in replaceCustomEscape.")
			}
		} else {
			__antithesis_instrumentation__.Notify(609062)
			if s[sIndex] == '\\' {
				__antithesis_instrumentation__.Notify(609063)

				if sIndex+1 == sLen {
					__antithesis_instrumentation__.Notify(609064)

					return "", errors.AssertionFailedf(
						"unexpected: a single backslash encountered in replaceCustomEscape.")
				} else {
					__antithesis_instrumentation__.Notify(609065)
					if s[sIndex+1] == '\\' {
						__antithesis_instrumentation__.Notify(609066)

						ret[retIndex] = '\\'
						ret[retIndex+1] = '\\'
						ret[retIndex+2] = '\\'
						ret[retIndex+3] = '\\'
						retIndex += 4
						sIndex += 2
					} else {
						__antithesis_instrumentation__.Notify(609067)

						if string(s[sIndex+1]) == string(escape) {
							__antithesis_instrumentation__.Notify(609068)

							if sIndex+2 == sLen {
								__antithesis_instrumentation__.Notify(609071)

								return "", errors.AssertionFailedf(
									"unexpected: escape character is the last one in replaceCustomEscape.")
							} else {
								__antithesis_instrumentation__.Notify(609072)
							}
							__antithesis_instrumentation__.Notify(609069)
							if sIndex+4 <= sLen {
								__antithesis_instrumentation__.Notify(609073)
								if s[sIndex+2] == '\\' && func() bool {
									__antithesis_instrumentation__.Notify(609074)
									return string(s[sIndex+3]) == string(escape) == true
								}() == true {
									__antithesis_instrumentation__.Notify(609075)

									ret[retIndex] = '\\'

									ret[retIndex+1] = string(escape)[0]
									retIndex += 2
									sIndex += 4
									continue
								} else {
									__antithesis_instrumentation__.Notify(609076)
								}
							} else {
								__antithesis_instrumentation__.Notify(609077)
							}
							__antithesis_instrumentation__.Notify(609070)

							ret[retIndex] = '\\'
							retIndex++
							sIndex += 2
						} else {
							__antithesis_instrumentation__.Notify(609078)

							ret[retIndex] = '\\'
							ret[retIndex+1] = s[sIndex+1]
							retIndex += 2
							sIndex += 2
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(609079)

				copy(ret[retIndex:], s[sIndex:sIndex+w])
				retIndex += w
				sIndex += w
			}
		}
	}
	__antithesis_instrumentation__.Notify(609051)
	return string(ret), nil
}

func calculateLengthAfterReplacingCustomEscape(s string, escape rune) (bool, int, error) {
	__antithesis_instrumentation__.Notify(609080)
	changed := false
	retLen, sLen := 0, len(s)
	for i := 0; i < sLen; {
		__antithesis_instrumentation__.Notify(609082)
		sRune, w := utf8.DecodeRuneInString(s[i:])
		if sRune == escape {
			__antithesis_instrumentation__.Notify(609083)

			if i+w < sLen {
				__antithesis_instrumentation__.Notify(609084)

				tRune, _ := utf8.DecodeRuneInString(s[(i + w):])
				if tRune == escape {
					__antithesis_instrumentation__.Notify(609085)

					changed = true
					retLen += w
					i += 2 * w
				} else {
					__antithesis_instrumentation__.Notify(609086)

					changed = true
					retLen += 2
					i += w
				}
			} else {
				__antithesis_instrumentation__.Notify(609087)

				return false, 0, pgerror.Newf(pgcode.InvalidEscapeSequence, "LIKE pattern must not end with escape character")
			}
		} else {
			__antithesis_instrumentation__.Notify(609088)
			if s[i] == '\\' {
				__antithesis_instrumentation__.Notify(609089)

				if i+1 == sLen {
					__antithesis_instrumentation__.Notify(609090)

					return false, 0, pgerror.Newf(pgcode.InvalidEscapeSequence, "Unexpected behavior during processing custom escape character.")
				} else {
					__antithesis_instrumentation__.Notify(609091)
					if s[i+1] == '\\' {
						__antithesis_instrumentation__.Notify(609092)

						changed = true
						retLen += 4
						i += 2
					} else {
						__antithesis_instrumentation__.Notify(609093)

						if string(s[i+1]) == string(escape) {
							__antithesis_instrumentation__.Notify(609094)

							if i+2 == sLen {
								__antithesis_instrumentation__.Notify(609097)

								return false, 0, pgerror.Newf(pgcode.InvalidEscapeSequence, "LIKE pattern must not end with escape character")
							} else {
								__antithesis_instrumentation__.Notify(609098)
							}
							__antithesis_instrumentation__.Notify(609095)
							if i+4 <= sLen {
								__antithesis_instrumentation__.Notify(609099)
								if s[i+2] == '\\' && func() bool {
									__antithesis_instrumentation__.Notify(609100)
									return string(s[i+3]) == string(escape) == true
								}() == true {
									__antithesis_instrumentation__.Notify(609101)

									changed = true
									retLen += 2
									i += 4
									continue
								} else {
									__antithesis_instrumentation__.Notify(609102)
								}
							} else {
								__antithesis_instrumentation__.Notify(609103)
							}
							__antithesis_instrumentation__.Notify(609096)

							changed = true
							retLen++
							i += 2
						} else {
							__antithesis_instrumentation__.Notify(609104)

							retLen += 2
							i += 2
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(609105)

				retLen += w
				i += w
			}
		}
	}
	__antithesis_instrumentation__.Notify(609081)
	return changed, retLen, nil
}

func (k likeKey) patternNoAnchor() (string, error) {
	__antithesis_instrumentation__.Notify(609106)

	pattern := regexp.QuoteMeta(k.s)
	var err error
	if k.escape == 0 {
		__antithesis_instrumentation__.Notify(609109)

		pattern = strings.Replace(pattern, `%`, `.*`, -1)
		pattern = strings.Replace(pattern, `_`, `.`, -1)
	} else {
		__antithesis_instrumentation__.Notify(609110)
		if k.escape == '\\' {
			__antithesis_instrumentation__.Notify(609111)

			pattern = replaceUnescaped(pattern, `%`, `.*`, `\\`)
			pattern = replaceUnescaped(pattern, `_`, `.`, `\\`)
		} else {
			__antithesis_instrumentation__.Notify(609112)

			if k.escape != '%' {
				__antithesis_instrumentation__.Notify(609115)

				if k.escape == '.' {
					__antithesis_instrumentation__.Notify(609116)

					pattern = replaceUnescaped(pattern, `%`, `..*`, regexp.QuoteMeta(string(k.escape)))
				} else {
					__antithesis_instrumentation__.Notify(609117)
					if k.escape == '*' {
						__antithesis_instrumentation__.Notify(609118)

						pattern = replaceUnescaped(pattern, `%`, `.**`, regexp.QuoteMeta(string(k.escape)))
					} else {
						__antithesis_instrumentation__.Notify(609119)
						pattern = replaceUnescaped(pattern, `%`, `.*`, regexp.QuoteMeta(string(k.escape)))
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(609120)
			}
			__antithesis_instrumentation__.Notify(609113)

			if k.escape != '_' {
				__antithesis_instrumentation__.Notify(609121)

				if k.escape == '.' {
					__antithesis_instrumentation__.Notify(609122)

					pattern = replaceUnescaped(pattern, `_`, `..`, regexp.QuoteMeta(string(k.escape)))
				} else {
					__antithesis_instrumentation__.Notify(609123)
					pattern = replaceUnescaped(pattern, `_`, `.`, regexp.QuoteMeta(string(k.escape)))
				}
			} else {
				__antithesis_instrumentation__.Notify(609124)
			}
			__antithesis_instrumentation__.Notify(609114)

			pattern = replaceUnescaped(pattern, string(k.escape)+`\\`, `\\`, regexp.QuoteMeta(string(k.escape)))

			if pattern, err = replaceCustomEscape(pattern, k.escape); err != nil {
				__antithesis_instrumentation__.Notify(609125)
				return pattern, err
			} else {
				__antithesis_instrumentation__.Notify(609126)
			}
		}
	}
	__antithesis_instrumentation__.Notify(609107)

	if k.escape != 0 {
		__antithesis_instrumentation__.Notify(609127)

		if pattern, err = unescapePattern(
			pattern,
			`\\`,
			k.escape == '\\',
		); err != nil {
			__antithesis_instrumentation__.Notify(609128)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(609129)
		}
	} else {
		__antithesis_instrumentation__.Notify(609130)
	}
	__antithesis_instrumentation__.Notify(609108)

	return pattern, nil
}

func (k likeKey) Pattern() (string, error) {
	__antithesis_instrumentation__.Notify(609131)
	pattern, err := k.patternNoAnchor()
	if err != nil {
		__antithesis_instrumentation__.Notify(609133)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(609134)
	}
	__antithesis_instrumentation__.Notify(609132)
	return anchorPattern(pattern, k.caseInsensitive), nil
}

type similarToKey struct {
	s      string
	escape rune
}

func (k similarToKey) Pattern() (string, error) {
	__antithesis_instrumentation__.Notify(609135)
	pattern := similarEscapeCustomChar(k.s, k.escape, k.escape != 0)
	return anchorPattern(pattern, false), nil
}

func SimilarToEscape(ctx *EvalContext, unescaped, pattern, escape string) (Datum, error) {
	__antithesis_instrumentation__.Notify(609136)
	key, err := makeSimilarToKey(pattern, escape)
	if err != nil {
		__antithesis_instrumentation__.Notify(609138)
		return DBoolFalse, err
	} else {
		__antithesis_instrumentation__.Notify(609139)
	}
	__antithesis_instrumentation__.Notify(609137)
	return matchRegexpWithKey(ctx, NewDString(unescaped), key)
}

func SimilarPattern(pattern, escape string) (Datum, error) {
	__antithesis_instrumentation__.Notify(609140)
	key, err := makeSimilarToKey(pattern, escape)
	if err != nil {
		__antithesis_instrumentation__.Notify(609143)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(609144)
	}
	__antithesis_instrumentation__.Notify(609141)
	pattern, err = key.Pattern()
	if err != nil {
		__antithesis_instrumentation__.Notify(609145)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(609146)
	}
	__antithesis_instrumentation__.Notify(609142)
	return NewDString(pattern), nil
}

func makeSimilarToKey(pattern, escape string) (similarToKey, error) {
	__antithesis_instrumentation__.Notify(609147)
	var escapeRune rune
	var width int
	escapeRune, width = utf8.DecodeRuneInString(escape)
	if len(escape) > width {
		__antithesis_instrumentation__.Notify(609149)
		return similarToKey{}, pgerror.Newf(pgcode.InvalidEscapeSequence, "invalid escape string")
	} else {
		__antithesis_instrumentation__.Notify(609150)
	}
	__antithesis_instrumentation__.Notify(609148)
	key := similarToKey{s: pattern, escape: escapeRune}
	return key, nil
}

type regexpKey struct {
	s               string
	caseInsensitive bool
}

func (k regexpKey) Pattern() (string, error) {
	__antithesis_instrumentation__.Notify(609151)
	if k.caseInsensitive {
		__antithesis_instrumentation__.Notify(609153)
		return caseInsensitive(k.s), nil
	} else {
		__antithesis_instrumentation__.Notify(609154)
	}
	__antithesis_instrumentation__.Notify(609152)
	return k.s, nil
}

func SimilarEscape(pattern string) string {
	__antithesis_instrumentation__.Notify(609155)
	return similarEscapeCustomChar(pattern, '\\', true)
}

func similarEscapeCustomChar(pattern string, escapeChar rune, isEscapeNonEmpty bool) string {
	__antithesis_instrumentation__.Notify(609156)
	patternBuilder := make([]rune, 0, utf8.RuneCountInString(pattern))

	inCharClass := false
	afterEscape := false
	numQuotes := 0
	for _, c := range pattern {
		__antithesis_instrumentation__.Notify(609158)
		switch {
		case afterEscape:
			__antithesis_instrumentation__.Notify(609159)

			if c == '"' && func() bool {
				__antithesis_instrumentation__.Notify(609170)
				return !inCharClass == true
			}() == true {
				__antithesis_instrumentation__.Notify(609171)
				if numQuotes%2 == 0 {
					__antithesis_instrumentation__.Notify(609173)
					patternBuilder = append(patternBuilder, '(')
				} else {
					__antithesis_instrumentation__.Notify(609174)
					patternBuilder = append(patternBuilder, ')')
				}
				__antithesis_instrumentation__.Notify(609172)
				numQuotes++
			} else {
				__antithesis_instrumentation__.Notify(609175)
				if c == escapeChar && func() bool {
					__antithesis_instrumentation__.Notify(609176)
					return len(string(escapeChar)) > 1 == true
				}() == true {
					__antithesis_instrumentation__.Notify(609177)

					patternBuilder = append(patternBuilder, c)
				} else {
					__antithesis_instrumentation__.Notify(609178)
					patternBuilder = append(patternBuilder, '\\', c)
				}
			}
			__antithesis_instrumentation__.Notify(609160)
			afterEscape = false
		case utf8.ValidRune(escapeChar) && func() bool {
			__antithesis_instrumentation__.Notify(609179)
			return c == escapeChar == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(609180)
			return isEscapeNonEmpty == true
		}() == true:
			__antithesis_instrumentation__.Notify(609161)

			afterEscape = true
		case inCharClass:
			__antithesis_instrumentation__.Notify(609162)
			if c == '\\' {
				__antithesis_instrumentation__.Notify(609181)
				patternBuilder = append(patternBuilder, '\\')
			} else {
				__antithesis_instrumentation__.Notify(609182)
			}
			__antithesis_instrumentation__.Notify(609163)
			patternBuilder = append(patternBuilder, c)
			if c == ']' {
				__antithesis_instrumentation__.Notify(609183)
				inCharClass = false
			} else {
				__antithesis_instrumentation__.Notify(609184)
			}
		case c == '[':
			__antithesis_instrumentation__.Notify(609164)
			patternBuilder = append(patternBuilder, c)
			inCharClass = true
		case c == '%':
			__antithesis_instrumentation__.Notify(609165)
			patternBuilder = append(patternBuilder, '.', '*')
		case c == '_':
			__antithesis_instrumentation__.Notify(609166)
			patternBuilder = append(patternBuilder, '.')
		case c == '(':
			__antithesis_instrumentation__.Notify(609167)

			patternBuilder = append(patternBuilder, '(', '?', ':')
		case c == '\\', c == '.', c == '^', c == '$':
			__antithesis_instrumentation__.Notify(609168)

			patternBuilder = append(patternBuilder, '\\', c)
		default:
			__antithesis_instrumentation__.Notify(609169)
			patternBuilder = append(patternBuilder, c)
		}
	}
	__antithesis_instrumentation__.Notify(609157)

	return string(patternBuilder)
}

func caseInsensitive(pattern string) string {
	__antithesis_instrumentation__.Notify(609185)
	return fmt.Sprintf("(?i:%s)", pattern)
}

func anchorPattern(pattern string, caseInsensitive bool) string {
	__antithesis_instrumentation__.Notify(609186)
	if caseInsensitive {
		__antithesis_instrumentation__.Notify(609188)
		return fmt.Sprintf("^(?si:%s)$", pattern)
	} else {
		__antithesis_instrumentation__.Notify(609189)
	}
	__antithesis_instrumentation__.Notify(609187)
	return fmt.Sprintf("^(?s:%s)$", pattern)
}

func FindEqualComparisonFunction(leftType, rightType *types.T) (TwoArgFn, bool) {
	__antithesis_instrumentation__.Notify(609190)
	fn, found := CmpOps[treecmp.EQ].LookupImpl(leftType, rightType)
	if found {
		__antithesis_instrumentation__.Notify(609192)
		return fn.Fn, true
	} else {
		__antithesis_instrumentation__.Notify(609193)
	}
	__antithesis_instrumentation__.Notify(609191)
	return nil, false
}

func IntPow(x, y DInt) (*DInt, error) {
	__antithesis_instrumentation__.Notify(609194)
	xd := apd.New(int64(x), 0)
	yd := apd.New(int64(y), 0)
	_, err := DecimalCtx.Pow(xd, xd, yd)
	if err != nil {
		__antithesis_instrumentation__.Notify(609197)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(609198)
	}
	__antithesis_instrumentation__.Notify(609195)
	i, err := xd.Int64()
	if err != nil {
		__antithesis_instrumentation__.Notify(609199)
		return nil, ErrIntOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(609200)
	}
	__antithesis_instrumentation__.Notify(609196)
	return NewDInt(DInt(i)), nil
}

func PickFromTuple(ctx *EvalContext, greatest bool, args Datums) (Datum, error) {
	__antithesis_instrumentation__.Notify(609201)
	g := args[0]

	for _, d := range args[1:] {
		__antithesis_instrumentation__.Notify(609203)
		var eval Datum
		var err error
		if greatest {
			__antithesis_instrumentation__.Notify(609206)
			eval, err = evalComparison(ctx, treecmp.MakeComparisonOperator(treecmp.LT), g, d)
		} else {
			__antithesis_instrumentation__.Notify(609207)
			eval, err = evalComparison(ctx, treecmp.MakeComparisonOperator(treecmp.LT), d, g)
		}
		__antithesis_instrumentation__.Notify(609204)
		if err != nil {
			__antithesis_instrumentation__.Notify(609208)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(609209)
		}
		__antithesis_instrumentation__.Notify(609205)
		if eval == DBoolTrue || func() bool {
			__antithesis_instrumentation__.Notify(609210)
			return (eval == DNull && func() bool {
				__antithesis_instrumentation__.Notify(609211)
				return g == DNull == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(609212)
			g = d
		} else {
			__antithesis_instrumentation__.Notify(609213)
		}
	}
	__antithesis_instrumentation__.Notify(609202)
	return g, nil
}

type CallbackValueGenerator struct {
	cb  func(ctx context.Context, prev int, txn *kv.Txn) (int, error)
	val int
	txn *kv.Txn
}

var _ ValueGenerator = &CallbackValueGenerator{}

func NewCallbackValueGenerator(
	cb func(ctx context.Context, prev int, txn *kv.Txn) (int, error),
) *CallbackValueGenerator {
	__antithesis_instrumentation__.Notify(609214)
	return &CallbackValueGenerator{
		cb: cb,
	}
}

func (c *CallbackValueGenerator) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(609215)
	return types.Int
}

func (c *CallbackValueGenerator) Start(_ context.Context, txn *kv.Txn) error {
	__antithesis_instrumentation__.Notify(609216)
	c.txn = txn
	return nil
}

func (c *CallbackValueGenerator) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(609217)
	var err error
	c.val, err = c.cb(ctx, c.val, c.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(609220)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(609221)
	}
	__antithesis_instrumentation__.Notify(609218)
	if c.val == -1 {
		__antithesis_instrumentation__.Notify(609222)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(609223)
	}
	__antithesis_instrumentation__.Notify(609219)
	return true, nil
}

func (c *CallbackValueGenerator) Values() (Datums, error) {
	__antithesis_instrumentation__.Notify(609224)
	return Datums{NewDInt(DInt(c.val))}, nil
}

func (c *CallbackValueGenerator) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(609225)
}

func Sqrt(x float64) (*DFloat, error) {
	__antithesis_instrumentation__.Notify(609226)
	if x < 0.0 {
		__antithesis_instrumentation__.Notify(609228)
		return nil, errSqrtOfNegNumber
	} else {
		__antithesis_instrumentation__.Notify(609229)
	}
	__antithesis_instrumentation__.Notify(609227)
	return NewDFloat(DFloat(math.Sqrt(x))), nil
}

func DecimalSqrt(x *apd.Decimal) (*DDecimal, error) {
	__antithesis_instrumentation__.Notify(609230)
	if x.Sign() < 0 {
		__antithesis_instrumentation__.Notify(609232)
		return nil, errSqrtOfNegNumber
	} else {
		__antithesis_instrumentation__.Notify(609233)
	}
	__antithesis_instrumentation__.Notify(609231)
	dd := &DDecimal{}
	_, err := DecimalCtx.Sqrt(&dd.Decimal, x)
	return dd, err
}

func Cbrt(x float64) (*DFloat, error) {
	__antithesis_instrumentation__.Notify(609234)
	return NewDFloat(DFloat(math.Cbrt(x))), nil
}

func DecimalCbrt(x *apd.Decimal) (*DDecimal, error) {
	__antithesis_instrumentation__.Notify(609235)
	dd := &DDecimal{}
	_, err := DecimalCtx.Cbrt(&dd.Decimal, x)
	return dd, err
}
