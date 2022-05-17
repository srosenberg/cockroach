package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"go/constant"
	"go/token"
	"math"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type Constant interface {
	Expr

	AvailableTypes() []*types.T

	DesirableTypes() []*types.T

	ResolveAsType(context.Context, *SemaContext, *types.T) (TypedExpr, error)
}

var _ Constant = &NumVal{}
var _ Constant = &StrVal{}

func isConstant(expr Expr) bool {
	__antithesis_instrumentation__.Notify(604450)
	_, ok := expr.(Constant)
	return ok
}

func typeCheckConstant(
	ctx context.Context, semaCtx *SemaContext, c Constant, desired *types.T,
) (ret TypedExpr, err error) {
	__antithesis_instrumentation__.Notify(604451)
	avail := c.AvailableTypes()
	if !desired.IsAmbiguous() {
		__antithesis_instrumentation__.Notify(604454)
		for _, typ := range avail {
			__antithesis_instrumentation__.Notify(604455)
			if desired.Equivalent(typ) {
				__antithesis_instrumentation__.Notify(604456)
				return c.ResolveAsType(ctx, semaCtx, desired)
			} else {
				__antithesis_instrumentation__.Notify(604457)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(604458)
	}
	__antithesis_instrumentation__.Notify(604452)

	if desired.Family() == types.IntFamily {
		__antithesis_instrumentation__.Notify(604459)
		if n, ok := c.(*NumVal); ok {
			__antithesis_instrumentation__.Notify(604460)
			_, err := n.AsInt64()
			switch {
			case errors.Is(err, errConstOutOfRange64):
				__antithesis_instrumentation__.Notify(604461)
				return nil, err
			case errors.Is(err, errConstNotInt):
				__antithesis_instrumentation__.Notify(604462)
			default:
				__antithesis_instrumentation__.Notify(604463)
				return nil, errors.NewAssertionErrorWithWrappedErrf(err, "unexpected error")
			}
		} else {
			__antithesis_instrumentation__.Notify(604464)
		}
	} else {
		__antithesis_instrumentation__.Notify(604465)
	}
	__antithesis_instrumentation__.Notify(604453)

	natural := avail[0]
	return c.ResolveAsType(ctx, semaCtx, natural)
}

func naturalConstantType(c Constant) *types.T {
	__antithesis_instrumentation__.Notify(604466)
	return c.AvailableTypes()[0]
}

func canConstantBecome(c Constant, typ *types.T) bool {
	__antithesis_instrumentation__.Notify(604467)
	avail := c.AvailableTypes()
	for _, availTyp := range avail {
		__antithesis_instrumentation__.Notify(604469)
		if availTyp.Equivalent(typ) {
			__antithesis_instrumentation__.Notify(604470)
			return true
		} else {
			__antithesis_instrumentation__.Notify(604471)
		}
	}
	__antithesis_instrumentation__.Notify(604468)
	return false
}

type NumVal struct {
	value constant.Value

	negative bool

	origString string

	resInt     DInt
	resFloat   DFloat
	resDecimal DDecimal
}

var _ Constant = &NumVal{}

func NewNumVal(value constant.Value, origString string, negative bool) *NumVal {
	__antithesis_instrumentation__.Notify(604472)
	return &NumVal{value: value, origString: origString, negative: negative}
}

func (expr *NumVal) Kind() constant.Kind {
	__antithesis_instrumentation__.Notify(604473)
	return expr.value.Kind()
}

func (expr *NumVal) ExactString() string {
	__antithesis_instrumentation__.Notify(604474)
	return expr.value.ExactString()
}

func (expr *NumVal) OrigString() string {
	__antithesis_instrumentation__.Notify(604475)
	return expr.origString
}

func (expr *NumVal) SetNegative() {
	__antithesis_instrumentation__.Notify(604476)
	expr.negative = true
}

func (expr *NumVal) Negate() {
	__antithesis_instrumentation__.Notify(604477)
	expr.negative = !expr.negative
}

func (expr *NumVal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604478)
	s := expr.origString
	if s == "" {
		__antithesis_instrumentation__.Notify(604481)
		s = expr.value.String()
	} else {
		__antithesis_instrumentation__.Notify(604482)
		if strings.EqualFold(s, "NaN") {
			__antithesis_instrumentation__.Notify(604483)
			s = "'NaN'"
		} else {
			__antithesis_instrumentation__.Notify(604484)
		}
	}
	__antithesis_instrumentation__.Notify(604479)
	if expr.negative {
		__antithesis_instrumentation__.Notify(604485)
		ctx.WriteByte('-')
	} else {
		__antithesis_instrumentation__.Notify(604486)
	}
	__antithesis_instrumentation__.Notify(604480)
	ctx.WriteString(s)
}

func (expr *NumVal) canBeInt64() bool {
	__antithesis_instrumentation__.Notify(604487)
	_, err := expr.AsInt64()
	return err == nil
}

func (expr *NumVal) ShouldBeInt64() bool {
	__antithesis_instrumentation__.Notify(604488)
	return expr.Kind() == constant.Int && func() bool {
		__antithesis_instrumentation__.Notify(604489)
		return expr.canBeInt64() == true
	}() == true
}

var errConstNotInt = pgerror.New(pgcode.NumericValueOutOfRange, "cannot represent numeric constant as an int")
var errConstOutOfRange64 = pgerror.New(pgcode.NumericValueOutOfRange, "numeric constant out of int64 range")
var errConstOutOfRange32 = pgerror.New(pgcode.NumericValueOutOfRange, "numeric constant out of int32 range")

func (expr *NumVal) AsInt64() (int64, error) {
	__antithesis_instrumentation__.Notify(604490)
	intVal, ok := expr.AsConstantInt()
	if !ok {
		__antithesis_instrumentation__.Notify(604493)
		return 0, errConstNotInt
	} else {
		__antithesis_instrumentation__.Notify(604494)
	}
	__antithesis_instrumentation__.Notify(604491)
	i, exact := constant.Int64Val(intVal)
	if !exact {
		__antithesis_instrumentation__.Notify(604495)
		return 0, errConstOutOfRange64
	} else {
		__antithesis_instrumentation__.Notify(604496)
	}
	__antithesis_instrumentation__.Notify(604492)
	expr.resInt = DInt(i)
	return i, nil
}

func (expr *NumVal) AsInt32() (int32, error) {
	__antithesis_instrumentation__.Notify(604497)
	intVal, ok := expr.AsConstantInt()
	if !ok {
		__antithesis_instrumentation__.Notify(604501)
		return 0, errConstNotInt
	} else {
		__antithesis_instrumentation__.Notify(604502)
	}
	__antithesis_instrumentation__.Notify(604498)
	i, exact := constant.Int64Val(intVal)
	if !exact {
		__antithesis_instrumentation__.Notify(604503)
		return 0, errConstOutOfRange32
	} else {
		__antithesis_instrumentation__.Notify(604504)
	}
	__antithesis_instrumentation__.Notify(604499)
	if i > math.MaxInt32 || func() bool {
		__antithesis_instrumentation__.Notify(604505)
		return i < math.MinInt32 == true
	}() == true {
		__antithesis_instrumentation__.Notify(604506)
		return 0, errConstOutOfRange32
	} else {
		__antithesis_instrumentation__.Notify(604507)
	}
	__antithesis_instrumentation__.Notify(604500)
	expr.resInt = DInt(i)
	return int32(i), nil
}

func (expr *NumVal) AsConstantValue() constant.Value {
	__antithesis_instrumentation__.Notify(604508)
	v := expr.value
	if expr.negative {
		__antithesis_instrumentation__.Notify(604510)
		v = constant.UnaryOp(token.SUB, v, 0)
	} else {
		__antithesis_instrumentation__.Notify(604511)
	}
	__antithesis_instrumentation__.Notify(604509)
	return v
}

func (expr *NumVal) AsConstantInt() (constant.Value, bool) {
	__antithesis_instrumentation__.Notify(604512)
	v := expr.AsConstantValue()
	intVal := constant.ToInt(v)
	if intVal.Kind() == constant.Int {
		__antithesis_instrumentation__.Notify(604514)
		return intVal, true
	} else {
		__antithesis_instrumentation__.Notify(604515)
	}
	__antithesis_instrumentation__.Notify(604513)
	return nil, false
}

var (
	intLikeTypes     = []*types.T{types.Int, types.Oid}
	decimalLikeTypes = []*types.T{types.Decimal, types.Float}

	NumValAvailInteger = append(intLikeTypes, decimalLikeTypes...)

	NumValAvailDecimalNoFraction = append(decimalLikeTypes, intLikeTypes...)

	NumValAvailDecimalWithFraction = decimalLikeTypes
)

func (expr *NumVal) AvailableTypes() []*types.T {
	__antithesis_instrumentation__.Notify(604516)
	switch {
	case expr.canBeInt64():
		__antithesis_instrumentation__.Notify(604517)
		if expr.Kind() == constant.Int {
			__antithesis_instrumentation__.Notify(604520)
			return NumValAvailInteger
		} else {
			__antithesis_instrumentation__.Notify(604521)
		}
		__antithesis_instrumentation__.Notify(604518)
		return NumValAvailDecimalNoFraction
	default:
		__antithesis_instrumentation__.Notify(604519)
		return NumValAvailDecimalWithFraction
	}
}

func (expr *NumVal) DesirableTypes() []*types.T {
	__antithesis_instrumentation__.Notify(604522)
	if expr.ShouldBeInt64() {
		__antithesis_instrumentation__.Notify(604524)
		return NumValAvailInteger
	} else {
		__antithesis_instrumentation__.Notify(604525)
	}
	__antithesis_instrumentation__.Notify(604523)
	return NumValAvailDecimalWithFraction
}

func (expr *NumVal) ResolveAsType(
	ctx context.Context, semaCtx *SemaContext, typ *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(604526)
	switch typ.Family() {
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(604527)

		if expr.resInt == 0 {
			__antithesis_instrumentation__.Notify(604538)
			if _, err := expr.AsInt64(); err != nil {
				__antithesis_instrumentation__.Notify(604539)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604540)
			}
		} else {
			__antithesis_instrumentation__.Notify(604541)
		}
		__antithesis_instrumentation__.Notify(604528)
		return AdjustValueToType(typ, &expr.resInt)
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(604529)
		if strings.EqualFold(expr.origString, "NaN") {
			__antithesis_instrumentation__.Notify(604542)

			expr.resFloat = DFloat(math.NaN())
		} else {
			__antithesis_instrumentation__.Notify(604543)
			f, _ := constant.Float64Val(expr.value)
			if expr.negative {
				__antithesis_instrumentation__.Notify(604545)
				f = -f
			} else {
				__antithesis_instrumentation__.Notify(604546)
			}
			__antithesis_instrumentation__.Notify(604544)
			expr.resFloat = DFloat(f)
		}
		__antithesis_instrumentation__.Notify(604530)
		return &expr.resFloat, nil
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(604531)
		dd := &expr.resDecimal
		s := expr.origString
		if s == "" {
			__antithesis_instrumentation__.Notify(604547)

			s = expr.ExactString()
		} else {
			__antithesis_instrumentation__.Notify(604548)
		}
		__antithesis_instrumentation__.Notify(604532)
		if idx := strings.IndexRune(s, '/'); idx != -1 {
			__antithesis_instrumentation__.Notify(604549)

			num, den := s[:idx], s[idx+1:]
			if err := dd.SetString(num); err != nil {
				__antithesis_instrumentation__.Notify(604552)
				return nil, pgerror.Wrapf(err, pgcode.Syntax,
					"could not evaluate numerator of %v as Datum type DDecimal from string %q",
					expr, num)
			} else {
				__antithesis_instrumentation__.Notify(604553)
			}
			__antithesis_instrumentation__.Notify(604550)

			denDec, err := ParseDDecimal(den)
			if err != nil {
				__antithesis_instrumentation__.Notify(604554)
				return nil, pgerror.Wrapf(err, pgcode.Syntax,
					"could not evaluate denominator %v as Datum type DDecimal from string %q",
					expr, den)
			} else {
				__antithesis_instrumentation__.Notify(604555)
			}
			__antithesis_instrumentation__.Notify(604551)
			if cond, err := DecimalCtx.Quo(&dd.Decimal, &dd.Decimal, &denDec.Decimal); err != nil {
				__antithesis_instrumentation__.Notify(604556)
				if cond.DivisionByZero() {
					__antithesis_instrumentation__.Notify(604558)
					return nil, ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(604559)
				}
				__antithesis_instrumentation__.Notify(604557)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(604560)
			}
		} else {
			__antithesis_instrumentation__.Notify(604561)
			if err := dd.SetString(s); err != nil {
				__antithesis_instrumentation__.Notify(604562)
				return nil, pgerror.Wrapf(err, pgcode.Syntax,
					"could not evaluate %v as Datum type DDecimal from string %q", expr, s)
			} else {
				__antithesis_instrumentation__.Notify(604563)
			}
		}
		__antithesis_instrumentation__.Notify(604533)
		if !dd.IsZero() {
			__antithesis_instrumentation__.Notify(604564)

			dd.Negative = dd.Negative != expr.negative
		} else {
			__antithesis_instrumentation__.Notify(604565)
		}
		__antithesis_instrumentation__.Notify(604534)
		return dd, nil
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(604535)
		d, err := expr.ResolveAsType(ctx, semaCtx, types.Int)
		if err != nil {
			__antithesis_instrumentation__.Notify(604566)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(604567)
		}
		__antithesis_instrumentation__.Notify(604536)
		oid := NewDOid(*d.(*DInt))
		return oid, nil
	default:
		__antithesis_instrumentation__.Notify(604537)
		return nil, errors.AssertionFailedf("could not resolve %T %v into a %T", expr, expr, typ)
	}
}

func intersectTypeSlices(xs, ys []*types.T) (out []*types.T) {
	__antithesis_instrumentation__.Notify(604568)
	seen := make(map[oid.Oid]struct{})
	for _, x := range xs {
		__antithesis_instrumentation__.Notify(604570)
		for _, y := range ys {
			__antithesis_instrumentation__.Notify(604572)
			_, ok := seen[x.Oid()]
			if x.Oid() == y.Oid() && func() bool {
				__antithesis_instrumentation__.Notify(604573)
				return !ok == true
			}() == true {
				__antithesis_instrumentation__.Notify(604574)
				out = append(out, x)
			} else {
				__antithesis_instrumentation__.Notify(604575)
			}
		}
		__antithesis_instrumentation__.Notify(604571)
		seen[x.Oid()] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(604569)
	return out
}

func commonConstantType(vals []Expr, idxs []int) (*types.T, bool) {
	__antithesis_instrumentation__.Notify(604576)
	var candidates []*types.T

	for _, i := range idxs {
		__antithesis_instrumentation__.Notify(604579)
		availableTypes := vals[i].(Constant).DesirableTypes()
		if candidates == nil {
			__antithesis_instrumentation__.Notify(604580)
			candidates = availableTypes
		} else {
			__antithesis_instrumentation__.Notify(604581)
			candidates = intersectTypeSlices(candidates, availableTypes)
		}
	}
	__antithesis_instrumentation__.Notify(604577)

	if len(candidates) > 0 {
		__antithesis_instrumentation__.Notify(604582)
		return candidates[0], true
	} else {
		__antithesis_instrumentation__.Notify(604583)
	}
	__antithesis_instrumentation__.Notify(604578)
	return nil, false
}

type StrVal struct {
	s string

	scannedAsBytes bool

	resString DString
	resBytes  DBytes
}

func NewStrVal(s string) *StrVal {
	__antithesis_instrumentation__.Notify(604584)
	return &StrVal{s: s}
}

func NewBytesStrVal(s string) *StrVal {
	__antithesis_instrumentation__.Notify(604585)
	return &StrVal{s: s, scannedAsBytes: true}
}

func (expr *StrVal) RawString() string {
	__antithesis_instrumentation__.Notify(604586)
	return expr.s
}

func (expr *StrVal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604587)
	buf, f := &ctx.Buffer, ctx.flags
	if expr.scannedAsBytes {
		__antithesis_instrumentation__.Notify(604588)
		lexbase.EncodeSQLBytes(buf, expr.s)
	} else {
		__antithesis_instrumentation__.Notify(604589)
		lexbase.EncodeSQLStringWithFlags(buf, expr.s, f.EncodeFlags())
	}
}

var (
	StrValAvailAllParsable = []*types.T{

		types.String,
		types.Bytes,
		types.Bool,
		types.Int,
		types.Float,
		types.Decimal,
		types.Date,
		types.StringArray,
		types.BytesArray,
		types.IntArray,
		types.FloatArray,
		types.DecimalArray,
		types.BoolArray,
		types.Box2D,
		types.Geography,
		types.Geometry,
		types.Time,
		types.TimeTZ,
		types.Timestamp,
		types.TimestampTZ,
		types.Interval,
		types.Uuid,
		types.DateArray,
		types.TimeArray,
		types.TimeTZArray,
		types.TimestampArray,
		types.TimestampTZArray,
		types.IntervalArray,
		types.UUIDArray,
		types.INet,
		types.Jsonb,
		types.VarBit,
		types.AnyEnum,
		types.AnyEnumArray,
		types.INetArray,
		types.VarBitArray,
		types.AnyTuple,
		types.AnyTupleArray,
	}

	StrValAvailBytes = []*types.T{types.Bytes, types.Uuid, types.String, types.AnyEnum}
)

func (expr *StrVal) AvailableTypes() []*types.T {
	__antithesis_instrumentation__.Notify(604590)
	if expr.scannedAsBytes {
		__antithesis_instrumentation__.Notify(604592)
		return StrValAvailBytes
	} else {
		__antithesis_instrumentation__.Notify(604593)
	}
	__antithesis_instrumentation__.Notify(604591)
	return StrValAvailAllParsable
}

func (expr *StrVal) DesirableTypes() []*types.T {
	__antithesis_instrumentation__.Notify(604594)
	return expr.AvailableTypes()
}

func (expr *StrVal) ResolveAsType(
	ctx context.Context, semaCtx *SemaContext, typ *types.T,
) (TypedExpr, error) {
	__antithesis_instrumentation__.Notify(604595)
	if expr.scannedAsBytes {
		__antithesis_instrumentation__.Notify(604597)

		switch typ.Family() {
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(604599)
			expr.resBytes = DBytes(expr.s)
			return &expr.resBytes, nil
		case types.EnumFamily:
			__antithesis_instrumentation__.Notify(604600)
			return MakeDEnumFromPhysicalRepresentation(typ, []byte(expr.s))
		case types.UuidFamily:
			__antithesis_instrumentation__.Notify(604601)
			return ParseDUuidFromBytes([]byte(expr.s))
		case types.StringFamily:
			__antithesis_instrumentation__.Notify(604602)
			expr.resString = DString(expr.s)
			return &expr.resString, nil
		default:
			__antithesis_instrumentation__.Notify(604603)
		}
		__antithesis_instrumentation__.Notify(604598)
		return nil, errors.AssertionFailedf("attempt to type byte array literal to %T", typ)
	} else {
		__antithesis_instrumentation__.Notify(604604)
	}
	__antithesis_instrumentation__.Notify(604596)

	switch typ.Family() {
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(604605)
		if typ.Oid() == oid.T_name {
			__antithesis_instrumentation__.Notify(604612)
			expr.resString = DString(expr.s)
			return NewDNameFromDString(&expr.resString), nil
		} else {
			__antithesis_instrumentation__.Notify(604613)
		}
		__antithesis_instrumentation__.Notify(604606)
		expr.resString = DString(expr.s)
		return &expr.resString, nil

	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(604607)
		return ParseDByte(expr.s)

	default:
		__antithesis_instrumentation__.Notify(604608)
		ptCtx := simpleParseTimeContext{

			RelativeParseTime: time.Date(2000, time.January, 2, 3, 4, 5, 0, time.UTC),
		}
		if semaCtx != nil {
			__antithesis_instrumentation__.Notify(604614)
			ptCtx.DateStyle = semaCtx.DateStyle
			ptCtx.IntervalStyle = semaCtx.IntervalStyle
		} else {
			__antithesis_instrumentation__.Notify(604615)
		}
		__antithesis_instrumentation__.Notify(604609)
		val, dependsOnContext, err := ParseAndRequireString(typ, expr.s, ptCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(604616)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(604617)
		}
		__antithesis_instrumentation__.Notify(604610)
		if !dependsOnContext {
			__antithesis_instrumentation__.Notify(604618)
			return val, nil
		} else {
			__antithesis_instrumentation__.Notify(604619)
		}
		__antithesis_instrumentation__.Notify(604611)

		expr.resString = DString(expr.s)
		c := NewTypedCastExpr(&expr.resString, typ)
		return c.TypeCheck(ctx, semaCtx, typ)
	}
}
