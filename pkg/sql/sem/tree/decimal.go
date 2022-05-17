package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

var (
	DecimalCtx = &apd.Context{
		Precision:   20,
		Rounding:    apd.RoundHalfUp,
		MaxExponent: 2000,
		MinExponent: -2000,

		Traps: apd.DefaultTraps &^ apd.InvalidOperation,
	}

	ExactCtx = DecimalCtx.WithPrecision(0)

	HighPrecisionCtx = DecimalCtx.WithPrecision(2000)

	IntermediateCtx = DecimalCtx.WithPrecision(DecimalCtx.Precision + 5)

	RoundCtx = func() *apd.Context {
		__antithesis_instrumentation__.Notify(607583)
		ctx := *HighPrecisionCtx
		ctx.Rounding = apd.RoundHalfEven
		return &ctx
	}()

	errScaleOutOfRange = pgerror.New(pgcode.NumericValueOutOfRange, "scale out of range")
)

func LimitDecimalWidth(d *apd.Decimal, precision, scale int) error {
	__antithesis_instrumentation__.Notify(607584)
	if d.Form != apd.Finite || func() bool {
		__antithesis_instrumentation__.Notify(607589)
		return precision <= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(607590)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(607591)
	}
	__antithesis_instrumentation__.Notify(607585)

	if scale < math.MinInt32+1 || func() bool {
		__antithesis_instrumentation__.Notify(607592)
		return scale > math.MaxInt32 == true
	}() == true {
		__antithesis_instrumentation__.Notify(607593)
		return errScaleOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(607594)
	}
	__antithesis_instrumentation__.Notify(607586)
	if scale > precision {
		__antithesis_instrumentation__.Notify(607595)
		return pgerror.Newf(pgcode.InvalidParameterValue, "scale (%d) must be between 0 and precision (%d)", scale, precision)
	} else {
		__antithesis_instrumentation__.Notify(607596)
	}
	__antithesis_instrumentation__.Notify(607587)

	c := DecimalCtx.WithPrecision(uint32(precision))
	c.Traps = apd.InvalidOperation

	if _, err := c.Quantize(d, d, -int32(scale)); err != nil {
		__antithesis_instrumentation__.Notify(607597)
		var lt string
		switch v := precision - scale; v {
		case 0:
			__antithesis_instrumentation__.Notify(607599)
			lt = "1"
		default:
			__antithesis_instrumentation__.Notify(607600)
			lt = fmt.Sprintf("10^%d", v)
		}
		__antithesis_instrumentation__.Notify(607598)
		return pgerror.Newf(pgcode.NumericValueOutOfRange, "value with precision %d, scale %d must round to an absolute value less than %s", precision, scale, lt)
	} else {
		__antithesis_instrumentation__.Notify(607601)
	}
	__antithesis_instrumentation__.Notify(607588)
	return nil
}
