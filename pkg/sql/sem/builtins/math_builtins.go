package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

func initMathBuiltins() {
	__antithesis_instrumentation__.Notify(601478)

	for k, v := range mathBuiltins {
		__antithesis_instrumentation__.Notify(601479)
		if _, exists := builtins[k]; exists {
			__antithesis_instrumentation__.Notify(601481)
			panic("duplicate builtin: " + k)
		} else {
			__antithesis_instrumentation__.Notify(601482)
		}
		__antithesis_instrumentation__.Notify(601480)
		builtins[k] = v
	}
}

var (
	errAbsOfMinInt64  = pgerror.New(pgcode.NumericValueOutOfRange, "abs of min integer value (-9223372036854775808) not defined")
	errLogOfNegNumber = pgerror.New(pgcode.InvalidArgumentForLogarithm, "cannot take logarithm of a negative number")
	errLogOfZero      = pgerror.New(pgcode.InvalidArgumentForLogarithm, "cannot take logarithm of zero")
)

const (
	degToRad = math.Pi / 180.0
	radToDeg = 180.0 / math.Pi
)

var mathBuiltins = map[string]builtinDefinition{
	"abs": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601483)
			return tree.NewDFloat(tree.DFloat(math.Abs(x))), nil
		}, "Calculates the absolute value of `val`.", tree.VolatilityImmutable),
		decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601484)
			dd := &tree.DDecimal{}
			dd.Abs(x)
			return dd, nil
		}, "Calculates the absolute value of `val`.", tree.VolatilityImmutable),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601485)
				x := tree.MustBeDInt(args[0])
				switch {
				case x == math.MinInt64:
					__antithesis_instrumentation__.Notify(601487)
					return nil, errAbsOfMinInt64
				case x < 0:
					__antithesis_instrumentation__.Notify(601488)
					return tree.NewDInt(-x), nil
				default:
					__antithesis_instrumentation__.Notify(601489)
				}
				__antithesis_instrumentation__.Notify(601486)
				return args[0], nil
			},
			Info:       "Calculates the absolute value of `val`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"acos": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601490)
			return tree.NewDFloat(tree.DFloat(math.Acos(x))), nil
		}, "Calculates the inverse cosine of `val`.", tree.VolatilityImmutable),
	),

	"acosd": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601491)
			return tree.NewDFloat(tree.DFloat(radToDeg * math.Acos(x))), nil
		}, "Calculates the inverse cosine of `val` with the result in degrees", tree.VolatilityImmutable),
	),

	"acosh": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601492)
			return tree.NewDFloat(tree.DFloat(math.Acosh(x))), nil
		}, "Calculates the inverse hyperbolic cosine of `val`.", tree.VolatilityImmutable),
	),

	"asin": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601493)
			return tree.NewDFloat(tree.DFloat(math.Asin(x))), nil
		}, "Calculates the inverse sine of `val`.", tree.VolatilityImmutable),
	),

	"asind": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601494)
			return tree.NewDFloat(tree.DFloat(radToDeg * math.Asin(x))), nil
		}, "Calculates the inverse sine of `val` with the result in degrees.", tree.VolatilityImmutable),
	),

	"asinh": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601495)
			return tree.NewDFloat(tree.DFloat(math.Asinh(x))), nil
		}, "Calculates the inverse hyperbolic sine of `val`.", tree.VolatilityImmutable),
	),

	"atan": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601496)
			return tree.NewDFloat(tree.DFloat(math.Atan(x))), nil
		}, "Calculates the inverse tangent of `val`.", tree.VolatilityImmutable),
	),

	"atand": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601497)
			return tree.NewDFloat(tree.DFloat(radToDeg * math.Atan(x))), nil
		}, "Calculates the inverse tangent of `val` with the result in degrees.", tree.VolatilityImmutable),
	),

	"atanh": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601498)
			return tree.NewDFloat(tree.DFloat(math.Atanh(x))), nil
		}, "Calculates the inverse hyperbolic tangent of `val`.", tree.VolatilityImmutable),
	),

	"atan2": makeBuiltin(defProps(),
		floatOverload2("x", "y", func(x, y float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601499)
			return tree.NewDFloat(tree.DFloat(math.Atan2(x, y))), nil
		}, "Calculates the inverse tangent of `x`/`y`.", tree.VolatilityImmutable),
	),

	"atan2d": makeBuiltin(defProps(),
		floatOverload2("x", "y", func(x, y float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601500)
			return tree.NewDFloat(tree.DFloat(radToDeg * math.Atan2(x, y))), nil
		}, "Calculates the inverse tangent of `x`/`y` with the result in degrees", tree.VolatilityImmutable),
	),

	"cbrt": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601501)
			return tree.Cbrt(x)
		}, "Calculates the cube root (∛) of `val`.", tree.VolatilityImmutable),
		decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601502)
			return tree.DecimalCbrt(x)
		}, "Calculates the cube root (∛) of `val`.", tree.VolatilityImmutable),
	),

	"ceil":    ceilImpl,
	"ceiling": ceilImpl,

	"cos": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601503)
			return tree.NewDFloat(tree.DFloat(math.Cos(x))), nil
		}, "Calculates the cosine of `val`.", tree.VolatilityImmutable),
	),

	"cosd": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601504)
			return tree.NewDFloat(tree.DFloat(math.Cos(degToRad * x))), nil
		}, "Calculates the cosine of `val` where `val` is in degrees.", tree.VolatilityImmutable),
	),

	"cosh": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601505)
			return tree.NewDFloat(tree.DFloat(math.Cosh(x))), nil
		}, "Calculates the hyperbolic cosine of `val`.", tree.VolatilityImmutable),
	),

	"cot": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601506)
			return tree.NewDFloat(tree.DFloat(1 / math.Tan(x))), nil
		}, "Calculates the cotangent of `val`.", tree.VolatilityImmutable),
	),

	"cotd": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601507)
			return tree.NewDFloat(tree.DFloat(1 / math.Tan(degToRad*x))), nil
		}, "Calculates the cotangent of `val` where `val` is in degrees.", tree.VolatilityImmutable),
	),

	"degrees": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601508)
			return tree.NewDFloat(tree.DFloat(radToDeg * x)), nil
		}, "Converts `val` as a radian value to a degree value.", tree.VolatilityImmutable),
	),

	"div": makeBuiltin(defProps(),
		floatOverload2("x", "y", func(x, y float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601509)
			return tree.NewDFloat(tree.DFloat(math.Trunc(x / y))), nil
		}, "Calculates the integer quotient of `x`/`y`.", tree.VolatilityImmutable),
		decimalOverload2("x", "y", func(x, y *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601510)
			if y.Sign() == 0 {
				__antithesis_instrumentation__.Notify(601512)
				return nil, tree.ErrDivByZero
			} else {
				__antithesis_instrumentation__.Notify(601513)
			}
			__antithesis_instrumentation__.Notify(601511)
			dd := &tree.DDecimal{}
			_, err := tree.HighPrecisionCtx.QuoInteger(&dd.Decimal, x, y)
			return dd, err
		}, "Calculates the integer quotient of `x`/`y`.", tree.VolatilityImmutable),
		tree.Overload{
			Types:      tree.ArgTypes{{"x", types.Int}, {"y", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601514)
				y := tree.MustBeDInt(args[1])
				if y == 0 {
					__antithesis_instrumentation__.Notify(601516)
					return nil, tree.ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(601517)
				}
				__antithesis_instrumentation__.Notify(601515)
				x := tree.MustBeDInt(args[0])
				return tree.NewDInt(x / y), nil
			},
			Info:       "Calculates the integer quotient of `x`/`y`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"exp": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601518)
			return tree.NewDFloat(tree.DFloat(math.Exp(x))), nil
		}, "Calculates *e* ^ `val`.", tree.VolatilityImmutable),
		decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601519)
			dd := &tree.DDecimal{}
			_, err := tree.DecimalCtx.Exp(&dd.Decimal, x)
			return dd, err
		}, "Calculates *e* ^ `val`.", tree.VolatilityImmutable),
	),

	"floor": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601520)
			return tree.NewDFloat(tree.DFloat(math.Floor(x))), nil
		}, "Calculates the largest integer not greater than `val`.", tree.VolatilityImmutable),
		decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601521)
			dd := &tree.DDecimal{}
			_, err := tree.ExactCtx.Floor(&dd.Decimal, x)
			return dd, err
		}, "Calculates the largest integer not greater than `val`.", tree.VolatilityImmutable),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Int}},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601522)
				return tree.NewDFloat(tree.DFloat(float64(*args[0].(*tree.DInt)))), nil
			},
			Info:       "Calculates the largest integer not greater than `val`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"isnan": makeBuiltin(defProps(),
		tree.Overload{

			Types:      tree.ArgTypes{{"val", types.Float}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601523)
				return tree.MakeDBool(tree.DBool(math.IsNaN(float64(*args[0].(*tree.DFloat))))), nil
			},
			Info:       "Returns true if `val` is NaN, false otherwise.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Decimal}},
			ReturnType: tree.FixedReturnType(types.Bool),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601524)
				isNaN := args[0].(*tree.DDecimal).Decimal.Form == apd.NaN
				return tree.MakeDBool(tree.DBool(isNaN)), nil
			},
			Info:       "Returns true if `val` is NaN, false otherwise.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"ln": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601525)
			return tree.NewDFloat(tree.DFloat(math.Log(x))), nil
		}, "Calculates the natural log of `val`.", tree.VolatilityImmutable),
		decimalLogFn(tree.DecimalCtx.Ln, "Calculates the natural log of `val`.", tree.VolatilityImmutable),
	),

	"log": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601526)
			return tree.NewDFloat(tree.DFloat(math.Log10(x))), nil
		}, "Calculates the base 10 log of `val`.", tree.VolatilityImmutable),
		floatOverload2("b", "x", func(b, x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601527)
			switch {
			case x < 0.0:
				__antithesis_instrumentation__.Notify(601530)
				return nil, errLogOfNegNumber
			case x == 0.0:
				__antithesis_instrumentation__.Notify(601531)
				return nil, errLogOfZero
			default:
				__antithesis_instrumentation__.Notify(601532)
			}
			__antithesis_instrumentation__.Notify(601528)
			switch {
			case b < 0.0:
				__antithesis_instrumentation__.Notify(601533)
				return nil, errLogOfNegNumber
			case b == 0.0:
				__antithesis_instrumentation__.Notify(601534)
				return nil, errLogOfZero
			default:
				__antithesis_instrumentation__.Notify(601535)
			}
			__antithesis_instrumentation__.Notify(601529)
			return tree.NewDFloat(tree.DFloat(math.Log10(x) / math.Log10(b))), nil
		}, "Calculates the base `b` log of `val`.", tree.VolatilityImmutable),
		decimalLogFn(tree.DecimalCtx.Log10, "Calculates the base 10 log of `val`.", tree.VolatilityImmutable),
		decimalOverload2("b", "x", func(b, x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601536)
			switch x.Sign() {
			case -1:
				__antithesis_instrumentation__.Notify(601541)
				return nil, errLogOfNegNumber
			case 0:
				__antithesis_instrumentation__.Notify(601542)
				return nil, errLogOfZero
			default:
				__antithesis_instrumentation__.Notify(601543)
			}
			__antithesis_instrumentation__.Notify(601537)
			switch b.Sign() {
			case -1:
				__antithesis_instrumentation__.Notify(601544)
				return nil, errLogOfNegNumber
			case 0:
				__antithesis_instrumentation__.Notify(601545)
				return nil, errLogOfZero
			default:
				__antithesis_instrumentation__.Notify(601546)
			}
			__antithesis_instrumentation__.Notify(601538)

			top := new(apd.Decimal)
			if _, err := tree.IntermediateCtx.Ln(top, x); err != nil {
				__antithesis_instrumentation__.Notify(601547)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601548)
			}
			__antithesis_instrumentation__.Notify(601539)
			bot := new(apd.Decimal)
			if _, err := tree.IntermediateCtx.Ln(bot, b); err != nil {
				__antithesis_instrumentation__.Notify(601549)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(601550)
			}
			__antithesis_instrumentation__.Notify(601540)

			dd := &tree.DDecimal{}
			_, err := tree.DecimalCtx.Quo(&dd.Decimal, top, bot)
			return dd, err
		}, "Calculates the base `b` log of `val`.", tree.VolatilityImmutable),
	),

	"mod": makeBuiltin(defProps(),
		floatOverload2("x", "y", func(x, y float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601551)
			return tree.NewDFloat(tree.DFloat(math.Mod(x, y))), nil
		}, "Calculates `x`%`y`.", tree.VolatilityImmutable),
		decimalOverload2("x", "y", func(x, y *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601552)
			if y.Sign() == 0 {
				__antithesis_instrumentation__.Notify(601554)
				return nil, tree.ErrDivByZero
			} else {
				__antithesis_instrumentation__.Notify(601555)
			}
			__antithesis_instrumentation__.Notify(601553)
			dd := &tree.DDecimal{}
			_, err := tree.HighPrecisionCtx.Rem(&dd.Decimal, x, y)
			return dd, err
		}, "Calculates `x`%`y`.", tree.VolatilityImmutable),
		tree.Overload{
			Types:      tree.ArgTypes{{"x", types.Int}, {"y", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601556)
				y := tree.MustBeDInt(args[1])
				if y == 0 {
					__antithesis_instrumentation__.Notify(601558)
					return nil, tree.ErrDivByZero
				} else {
					__antithesis_instrumentation__.Notify(601559)
				}
				__antithesis_instrumentation__.Notify(601557)
				x := tree.MustBeDInt(args[0])
				return tree.NewDInt(x % y), nil
			},
			Info:       "Calculates `x`%`y`.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"pi": makeBuiltin(defProps(),
		tree.Overload{
			Types:      tree.ArgTypes{},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601560)
				return tree.NewDFloat(math.Pi), nil
			},
			Info:       "Returns the value for pi (3.141592653589793).",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"pow":   powImpls,
	"power": powImpls,

	"radians": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601561)
			return tree.NewDFloat(tree.DFloat(x * degToRad)), nil
		}, "Converts `val` as a degree value to a radians value.", tree.VolatilityImmutable),
	),

	"round": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601562)
			return tree.NewDFloat(tree.DFloat(math.RoundToEven(x))), nil
		}, "Rounds `val` to the nearest integer using half to even (banker's) rounding.", tree.VolatilityImmutable),
		decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601563)
			return roundDecimal(x, 0)
		}, "Rounds `val` to the nearest integer, half away from zero: "+
			"round(+/-2.4) = +/-2, round(+/-2.5) = +/-3.", tree.VolatilityImmutable),
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.Float}, {"decimal_accuracy", types.Int}},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601564)
				f := float64(*args[0].(*tree.DFloat))
				if math.IsInf(f, 0) || func() bool {
					__antithesis_instrumentation__.Notify(601569)
					return math.IsNaN(f) == true
				}() == true {
					__antithesis_instrumentation__.Notify(601570)
					return args[0], nil
				} else {
					__antithesis_instrumentation__.Notify(601571)
				}
				__antithesis_instrumentation__.Notify(601565)
				var x apd.Decimal
				if _, err := x.SetFloat64(f); err != nil {
					__antithesis_instrumentation__.Notify(601572)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601573)
				}
				__antithesis_instrumentation__.Notify(601566)

				scale := int32(tree.MustBeDInt(args[1]))

				var d apd.Decimal
				if _, err := tree.RoundCtx.Quantize(&d, &x, -scale); err != nil {
					__antithesis_instrumentation__.Notify(601574)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601575)
				}
				__antithesis_instrumentation__.Notify(601567)

				f, err := d.Float64()
				if err != nil {
					__antithesis_instrumentation__.Notify(601576)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(601577)
				}
				__antithesis_instrumentation__.Notify(601568)

				return tree.NewDFloat(tree.DFloat(f)), nil
			},
			Info: "Keeps `decimal_accuracy` number of figures to the right of the zero position " +
				" in `input` using half to even (banker's) rounding.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"input", types.Decimal}, {"decimal_accuracy", types.Int}},
			ReturnType: tree.FixedReturnType(types.Decimal),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601578)

				scale := int32(tree.MustBeDInt(args[1]))
				return roundDecimal(&args[0].(*tree.DDecimal).Decimal, scale)
			},
			Info: "Keeps `decimal_accuracy` number of figures to the right of the zero position " +
				"in `input` using half away from zero rounding. If `decimal_accuracy` " +
				"is not in the range -2^31...(2^31-1), the results are undefined.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"sin": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601579)
			return tree.NewDFloat(tree.DFloat(math.Sin(x))), nil
		}, "Calculates the sine of `val`.", tree.VolatilityImmutable),
	),

	"sind": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601580)
			return tree.NewDFloat(tree.DFloat(math.Sin(degToRad * x))), nil
		}, "Calculates the sine of `val` where `val` is in degrees.", tree.VolatilityImmutable),
	),

	"sinh": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601581)
			return tree.NewDFloat(tree.DFloat(math.Sinh(x))), nil
		}, "Calculates the hyperbolic sine of `val`.", tree.VolatilityImmutable),
	),

	"sign": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601582)
			switch {
			case x < 0:
				__antithesis_instrumentation__.Notify(601584)
				return tree.NewDFloat(-1), nil
			case x == 0:
				__antithesis_instrumentation__.Notify(601585)
				return tree.NewDFloat(0), nil
			default:
				__antithesis_instrumentation__.Notify(601586)
			}
			__antithesis_instrumentation__.Notify(601583)
			return tree.NewDFloat(1), nil
		}, "Determines the sign of `val`: **1** for positive; **0** for 0 values; **-1** for "+
			"negative.", tree.VolatilityImmutable),
		decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601587)
			d := &tree.DDecimal{}
			d.Decimal.SetInt64(int64(x.Sign()))
			return d, nil
		}, "Determines the sign of `val`: **1** for positive; **0** for 0 values; **-1** for "+
			"negative.", tree.VolatilityImmutable),
		tree.Overload{
			Types:      tree.ArgTypes{{"val", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601588)
				x := tree.MustBeDInt(args[0])
				switch {
				case x < 0:
					__antithesis_instrumentation__.Notify(601590)
					return tree.NewDInt(-1), nil
				case x == 0:
					__antithesis_instrumentation__.Notify(601591)
					return tree.DZero, nil
				default:
					__antithesis_instrumentation__.Notify(601592)
				}
				__antithesis_instrumentation__.Notify(601589)
				return tree.NewDInt(1), nil
			},
			Info: "Determines the sign of `val`: **1** for positive; **0** for 0 values; **-1** " +
				"for negative.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"sqrt": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601593)
			return tree.Sqrt(x)
		}, "Calculates the square root of `val`.", tree.VolatilityImmutable),
		decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601594)
			return tree.DecimalSqrt(x)
		}, "Calculates the square root of `val`.", tree.VolatilityImmutable),
	),

	"tan": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601595)
			return tree.NewDFloat(tree.DFloat(math.Tan(x))), nil
		}, "Calculates the tangent of `val`.", tree.VolatilityImmutable),
	),

	"tand": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601596)
			return tree.NewDFloat(tree.DFloat(math.Tan(degToRad * x))), nil
		}, "Calculates the tangent of `val` where `val` is in degrees.", tree.VolatilityImmutable),
	),

	"tanh": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601597)
			return tree.NewDFloat(tree.DFloat(math.Tanh(x))), nil
		}, "Calculates the hyperbolic tangent of `val`.", tree.VolatilityImmutable),
	),

	"trunc": makeBuiltin(defProps(),
		floatOverload1(func(x float64) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601598)
			return tree.NewDFloat(tree.DFloat(math.Trunc(x))), nil
		}, "Truncates the decimal values of `val`.", tree.VolatilityImmutable),
		decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601599)
			dd := &tree.DDecimal{}
			x.Modf(&dd.Decimal, nil)
			return dd, nil
		}, "Truncates the decimal values of `val`.", tree.VolatilityImmutable),
	),

	"width_bucket": makeBuiltin(defProps(),
		tree.Overload{
			Types: tree.ArgTypes{{"operand", types.Decimal}, {"b1", types.Decimal},
				{"b2", types.Decimal}, {"count", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601600)
				operand, _ := args[0].(*tree.DDecimal).Float64()
				b1, _ := args[1].(*tree.DDecimal).Float64()
				b2, _ := args[2].(*tree.DDecimal).Float64()
				if math.IsInf(operand, 0) || func() bool {
					__antithesis_instrumentation__.Notify(601602)
					return math.IsInf(b1, 0) == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(601603)
					return math.IsInf(b2, 0) == true
				}() == true {
					__antithesis_instrumentation__.Notify(601604)
					return nil, pgerror.New(
						pgcode.InvalidParameterValue,
						"operand, lower bound, and upper bound cannot be infinity",
					)
				} else {
					__antithesis_instrumentation__.Notify(601605)
				}
				__antithesis_instrumentation__.Notify(601601)
				count := int(tree.MustBeDInt(args[3]))
				return tree.NewDInt(tree.DInt(widthBucket(operand, b1, b2, count))), nil
			},
			Info: "return the bucket number to which operand would be assigned in a histogram having count " +
				"equal-width buckets spanning the range b1 to b2.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types: tree.ArgTypes{{"operand", types.Int}, {"b1", types.Int},
				{"b2", types.Int}, {"count", types.Int}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601606)
				operand := float64(tree.MustBeDInt(args[0]))
				b1 := float64(tree.MustBeDInt(args[1]))
				b2 := float64(tree.MustBeDInt(args[2]))
				count := int(tree.MustBeDInt(args[3]))
				return tree.NewDInt(tree.DInt(widthBucket(operand, b1, b2, count))), nil
			},
			Info: "return the bucket number to which operand would be assigned in a histogram having count " +
				"equal-width buckets spanning the range b1 to b2.",
			Volatility: tree.VolatilityImmutable,
		},
		tree.Overload{
			Types:      tree.ArgTypes{{"operand", types.Any}, {"thresholds", types.AnyArray}},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(601607)
				operand := args[0]
				thresholds := tree.MustBeDArray(args[1])

				if !operand.ResolvedType().Equivalent(thresholds.ParamTyp) {
					__antithesis_instrumentation__.Notify(601610)
					return tree.NewDInt(0), errors.New("operand and thresholds must be of the same type")
				} else {
					__antithesis_instrumentation__.Notify(601611)
				}
				__antithesis_instrumentation__.Notify(601608)

				for i, v := range thresholds.Array {
					__antithesis_instrumentation__.Notify(601612)
					if operand.Compare(ctx, v) < 0 {
						__antithesis_instrumentation__.Notify(601613)
						return tree.NewDInt(tree.DInt(i)), nil
					} else {
						__antithesis_instrumentation__.Notify(601614)
					}
				}
				__antithesis_instrumentation__.Notify(601609)

				return tree.NewDInt(tree.DInt(thresholds.Len())), nil
			},
			Info: "return the bucket number to which operand would be assigned given an array listing the " +
				"lower bounds of the buckets; returns 0 for an input less than the first lower bound; the " +
				"thresholds array must be sorted, smallest first, or unexpected results will be obtained",
			Volatility: tree.VolatilityImmutable,
		},
	),
}

var ceilImpl = makeBuiltin(defProps(),
	floatOverload1(func(x float64) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(601615)
		return tree.NewDFloat(tree.DFloat(math.Ceil(x))), nil
	}, "Calculates the smallest integer not smaller than `val`.", tree.VolatilityImmutable),
	decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(601616)
		dd := &tree.DDecimal{}
		_, err := tree.ExactCtx.Ceil(&dd.Decimal, x)
		if dd.IsZero() {
			__antithesis_instrumentation__.Notify(601618)
			dd.Negative = false
		} else {
			__antithesis_instrumentation__.Notify(601619)
		}
		__antithesis_instrumentation__.Notify(601617)
		return dd, err
	}, "Calculates the smallest integer not smaller than `val`.", tree.VolatilityImmutable),
	tree.Overload{
		Types:      tree.ArgTypes{{"val", types.Int}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601620)
			return tree.NewDFloat(tree.DFloat(float64(*args[0].(*tree.DInt)))), nil
		},
		Info:       "Calculates the smallest integer not smaller than `val`.",
		Volatility: tree.VolatilityImmutable,
	},
)

var powImpls = makeBuiltin(defProps(),
	floatOverload2("x", "y", func(x, y float64) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(601621)
		return tree.NewDFloat(tree.DFloat(math.Pow(x, y))), nil
	}, "Calculates `x`^`y`.", tree.VolatilityImmutable),
	decimalOverload2("x", "y", func(x, y *apd.Decimal) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(601622)
		dd := &tree.DDecimal{}
		_, err := tree.DecimalCtx.Pow(&dd.Decimal, x, y)
		return dd, err
	}, "Calculates `x`^`y`.", tree.VolatilityImmutable),
	tree.Overload{
		Types: tree.ArgTypes{
			{"x", types.Int},
			{"y", types.Int},
		},
		ReturnType: tree.FixedReturnType(types.Int),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601623)
			return tree.IntPow(tree.MustBeDInt(args[0]), tree.MustBeDInt(args[1]))
		},
		Info:       "Calculates `x`^`y`.",
		Volatility: tree.VolatilityImmutable,
	},
)

func decimalLogFn(
	logFn func(*apd.Decimal, *apd.Decimal) (apd.Condition, error),
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601624)
	return decimalOverload1(func(x *apd.Decimal) (tree.Datum, error) {
		__antithesis_instrumentation__.Notify(601625)
		switch x.Sign() {
		case -1:
			__antithesis_instrumentation__.Notify(601627)
			return nil, errLogOfNegNumber
		case 0:
			__antithesis_instrumentation__.Notify(601628)
			return nil, errLogOfZero
		default:
			__antithesis_instrumentation__.Notify(601629)
		}
		__antithesis_instrumentation__.Notify(601626)
		dd := &tree.DDecimal{}
		_, err := logFn(&dd.Decimal, x)
		return dd, err
	}, info, tree.VolatilityImmutable)
}

func floatOverload1(
	f func(float64) (tree.Datum, error), info string, volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601630)
	return tree.Overload{
		Types:      tree.ArgTypes{{"val", types.Float}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601631)
			return f(float64(*args[0].(*tree.DFloat)))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func floatOverload2(
	a, b string,
	f func(float64, float64) (tree.Datum, error),
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601632)
	return tree.Overload{
		Types:      tree.ArgTypes{{a, types.Float}, {b, types.Float}},
		ReturnType: tree.FixedReturnType(types.Float),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601633)
			return f(float64(*args[0].(*tree.DFloat)),
				float64(*args[1].(*tree.DFloat)))
		},
		Info:       info,
		Volatility: volatility,
	}
}

func decimalOverload1(
	f func(*apd.Decimal) (tree.Datum, error), info string, volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601634)
	return tree.Overload{
		Types:      tree.ArgTypes{{"val", types.Decimal}},
		ReturnType: tree.FixedReturnType(types.Decimal),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601635)
			dec := &args[0].(*tree.DDecimal).Decimal
			return f(dec)
		},
		Info:       info,
		Volatility: volatility,
	}
}

func decimalOverload2(
	a, b string,
	f func(*apd.Decimal, *apd.Decimal) (tree.Datum, error),
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(601636)
	return tree.Overload{
		Types:      tree.ArgTypes{{a, types.Decimal}, {b, types.Decimal}},
		ReturnType: tree.FixedReturnType(types.Decimal),
		Fn: func(_ *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(601637)
			dec1 := &args[0].(*tree.DDecimal).Decimal
			dec2 := &args[1].(*tree.DDecimal).Decimal
			return f(dec1, dec2)
		},
		Info:       info,
		Volatility: volatility,
	}
}

func roundDDecimal(d *tree.DDecimal, scale int32) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(601638)

	if -d.Exponent <= scale {
		__antithesis_instrumentation__.Notify(601640)
		return d, nil
	} else {
		__antithesis_instrumentation__.Notify(601641)
	}
	__antithesis_instrumentation__.Notify(601639)
	return roundDecimal(&d.Decimal, scale)
}

func roundDecimal(x *apd.Decimal, scale int32) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(601642)
	dd := &tree.DDecimal{}
	_, err := tree.HighPrecisionCtx.Quantize(&dd.Decimal, x, -scale)
	return dd, err
}

var uniqueIntState struct {
	syncutil.Mutex
	timestamp uint64
}

var uniqueIntEpoch = time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC).UnixNano()

func widthBucket(operand float64, b1 float64, b2 float64, count int) int {
	__antithesis_instrumentation__.Notify(601643)
	bucket := 0
	if (b1 < b2 && func() bool {
		__antithesis_instrumentation__.Notify(601646)
		return operand > b2 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(601647)
		return (b1 > b2 && func() bool {
			__antithesis_instrumentation__.Notify(601648)
			return operand < b2 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(601649)
		return count + 1
	} else {
		__antithesis_instrumentation__.Notify(601650)
	}
	__antithesis_instrumentation__.Notify(601644)

	if (b1 < b2 && func() bool {
		__antithesis_instrumentation__.Notify(601651)
		return operand < b1 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(601652)
		return (b1 > b2 && func() bool {
			__antithesis_instrumentation__.Notify(601653)
			return operand > b1 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(601654)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(601655)
	}
	__antithesis_instrumentation__.Notify(601645)

	width := (b2 - b1) / float64(count)
	difference := operand - b1
	bucket = int(math.Floor(difference/width) + 1)

	return bucket
}
