package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

const FollowerReadTimestampFunctionName = "follower_read_timestamp"

const FollowerReadTimestampExperimentalFunctionName = "experimental_follower_read_timestamp"

const WithMinTimestampFunctionName = "with_min_timestamp"

const WithMaxStalenessFunctionName = "with_max_staleness"

func IsFollowerReadTimestampFunction(asOf AsOfClause, searchPath sessiondata.SearchPath) bool {
	__antithesis_instrumentation__.Notify(603259)
	return resolveAsOfFuncType(asOf, searchPath) == asOfFuncTypeFollowerRead
}

type asOfFuncType int

const (
	asOfFuncTypeInvalid asOfFuncType = iota
	asOfFuncTypeFollowerRead
	asOfFuncTypeBoundedStaleness
)

func resolveAsOfFuncType(asOf AsOfClause, searchPath sessiondata.SearchPath) asOfFuncType {
	__antithesis_instrumentation__.Notify(603260)
	fe, ok := asOf.Expr.(*FuncExpr)
	if !ok {
		__antithesis_instrumentation__.Notify(603264)
		return asOfFuncTypeInvalid
	} else {
		__antithesis_instrumentation__.Notify(603265)
	}
	__antithesis_instrumentation__.Notify(603261)
	def, err := fe.Func.Resolve(searchPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(603266)
		return asOfFuncTypeInvalid
	} else {
		__antithesis_instrumentation__.Notify(603267)
	}
	__antithesis_instrumentation__.Notify(603262)
	switch def.Name {
	case FollowerReadTimestampFunctionName, FollowerReadTimestampExperimentalFunctionName:
		__antithesis_instrumentation__.Notify(603268)
		return asOfFuncTypeFollowerRead
	case WithMinTimestampFunctionName, WithMaxStalenessFunctionName:
		__antithesis_instrumentation__.Notify(603269)
		return asOfFuncTypeBoundedStaleness
	default:
		__antithesis_instrumentation__.Notify(603270)
	}
	__antithesis_instrumentation__.Notify(603263)
	return asOfFuncTypeInvalid
}

type AsOfSystemTime struct {
	Timestamp hlc.Timestamp

	BoundedStaleness bool

	NearestOnly bool

	MaxTimestampBound hlc.Timestamp
}

type evalAsOfTimestampOptions struct {
	allowBoundedStaleness bool
}

type EvalAsOfTimestampOption func(o evalAsOfTimestampOptions) evalAsOfTimestampOptions

var EvalAsOfTimestampOptionAllowBoundedStaleness EvalAsOfTimestampOption = func(
	o evalAsOfTimestampOptions,
) evalAsOfTimestampOptions {
	__antithesis_instrumentation__.Notify(603271)
	o.allowBoundedStaleness = true
	return o
}

func EvalAsOfTimestamp(
	ctx context.Context,
	asOf AsOfClause,
	semaCtx *SemaContext,
	evalCtx *EvalContext,
	opts ...EvalAsOfTimestampOption,
) (AsOfSystemTime, error) {
	__antithesis_instrumentation__.Notify(603272)
	o := evalAsOfTimestampOptions{}
	for _, f := range opts {
		__antithesis_instrumentation__.Notify(603278)
		o = f(o)
	}
	__antithesis_instrumentation__.Notify(603273)

	newInvalidExprError := func() error {
		__antithesis_instrumentation__.Notify(603279)
		var optFuncs string
		if o.allowBoundedStaleness {
			__antithesis_instrumentation__.Notify(603281)
			optFuncs = fmt.Sprintf(
				", %s, %s,",
				WithMinTimestampFunctionName,
				WithMaxStalenessFunctionName,
			)
		} else {
			__antithesis_instrumentation__.Notify(603282)
		}
		__antithesis_instrumentation__.Notify(603280)
		return errors.Errorf(
			"AS OF SYSTEM TIME: only constant expressions%s or %s are allowed",
			optFuncs,
			FollowerReadTimestampFunctionName,
		)
	}
	__antithesis_instrumentation__.Notify(603274)

	scalarProps := &semaCtx.Properties
	defer scalarProps.Restore(*scalarProps)
	scalarProps.Require("AS OF SYSTEM TIME", RejectSpecial|RejectSubqueries)

	var ret AsOfSystemTime

	var te TypedExpr
	if asOfFuncExpr, ok := asOf.Expr.(*FuncExpr); ok {
		__antithesis_instrumentation__.Notify(603283)
		switch resolveAsOfFuncType(asOf, semaCtx.SearchPath) {
		case asOfFuncTypeFollowerRead:
			__antithesis_instrumentation__.Notify(603285)
		case asOfFuncTypeBoundedStaleness:
			__antithesis_instrumentation__.Notify(603286)
			if !o.allowBoundedStaleness {
				__antithesis_instrumentation__.Notify(603289)
				return AsOfSystemTime{}, newInvalidExprError()
			} else {
				__antithesis_instrumentation__.Notify(603290)
			}
			__antithesis_instrumentation__.Notify(603287)
			ret.BoundedStaleness = true

			if len(asOfFuncExpr.Exprs) == 2 {
				__antithesis_instrumentation__.Notify(603291)
				nearestOnlyExpr, err := asOfFuncExpr.Exprs[1].TypeCheck(ctx, semaCtx, types.Bool)
				if err != nil {
					__antithesis_instrumentation__.Notify(603295)
					return AsOfSystemTime{}, err
				} else {
					__antithesis_instrumentation__.Notify(603296)
				}
				__antithesis_instrumentation__.Notify(603292)
				nearestOnlyEval, err := nearestOnlyExpr.Eval(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(603297)
					return AsOfSystemTime{}, err
				} else {
					__antithesis_instrumentation__.Notify(603298)
				}
				__antithesis_instrumentation__.Notify(603293)
				nearestOnly, ok := nearestOnlyEval.(*DBool)
				if !ok {
					__antithesis_instrumentation__.Notify(603299)
					return AsOfSystemTime{}, pgerror.Newf(
						pgcode.InvalidParameterValue,
						"%s: expected bool argument for nearest_only",
						asOfFuncExpr.Func.String(),
					)
				} else {
					__antithesis_instrumentation__.Notify(603300)
				}
				__antithesis_instrumentation__.Notify(603294)
				ret.NearestOnly = bool(*nearestOnly)
			} else {
				__antithesis_instrumentation__.Notify(603301)
			}
		default:
			__antithesis_instrumentation__.Notify(603288)
			return AsOfSystemTime{}, newInvalidExprError()
		}
		__antithesis_instrumentation__.Notify(603284)
		var err error
		te, err = asOf.Expr.TypeCheck(ctx, semaCtx, types.TimestampTZ)
		if err != nil {
			__antithesis_instrumentation__.Notify(603302)
			return AsOfSystemTime{}, err
		} else {
			__antithesis_instrumentation__.Notify(603303)
		}
	} else {
		__antithesis_instrumentation__.Notify(603304)
		var err error
		te, err = asOf.Expr.TypeCheck(ctx, semaCtx, types.String)
		if err != nil {
			__antithesis_instrumentation__.Notify(603306)
			return AsOfSystemTime{}, err
		} else {
			__antithesis_instrumentation__.Notify(603307)
		}
		__antithesis_instrumentation__.Notify(603305)
		if !IsConst(evalCtx, te) {
			__antithesis_instrumentation__.Notify(603308)
			return AsOfSystemTime{}, newInvalidExprError()
		} else {
			__antithesis_instrumentation__.Notify(603309)
		}
	}
	__antithesis_instrumentation__.Notify(603275)

	d, err := te.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(603310)
		return AsOfSystemTime{}, err
	} else {
		__antithesis_instrumentation__.Notify(603311)
	}
	__antithesis_instrumentation__.Notify(603276)

	stmtTimestamp := evalCtx.GetStmtTimestamp()
	ret.Timestamp, err = DatumToHLC(evalCtx, stmtTimestamp, d)
	if err != nil {
		__antithesis_instrumentation__.Notify(603312)
		return AsOfSystemTime{}, errors.Wrap(err, "AS OF SYSTEM TIME")
	} else {
		__antithesis_instrumentation__.Notify(603313)
	}
	__antithesis_instrumentation__.Notify(603277)
	return ret, nil
}

func DatumToHLC(evalCtx *EvalContext, stmtTimestamp time.Time, d Datum) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(603314)
	ts := hlc.Timestamp{}
	var convErr error
	switch d := d.(type) {
	case *DString:
		__antithesis_instrumentation__.Notify(603318)
		s := string(*d)

		syn := false
		if strings.HasSuffix(s, "?") {
			__antithesis_instrumentation__.Notify(603329)
			s = s[:len(s)-1]
			syn = true
		} else {
			__antithesis_instrumentation__.Notify(603330)
		}
		__antithesis_instrumentation__.Notify(603319)

		if dt, _, err := ParseDTimestamp(evalCtx, s, time.Nanosecond); err == nil {
			__antithesis_instrumentation__.Notify(603331)
			ts.WallTime = dt.Time.UnixNano()
			ts.Synthetic = syn
			break
		} else {
			__antithesis_instrumentation__.Notify(603332)
		}
		__antithesis_instrumentation__.Notify(603320)

		if dec, _, err := apd.NewFromString(s); err == nil {
			__antithesis_instrumentation__.Notify(603333)
			ts, convErr = DecimalToHLC(dec)
			ts.Synthetic = syn
			break
		} else {
			__antithesis_instrumentation__.Notify(603334)
		}
		__antithesis_instrumentation__.Notify(603321)

		if iv, err := ParseDInterval(evalCtx.GetIntervalStyle(), s); err == nil {
			__antithesis_instrumentation__.Notify(603335)
			if (iv.Duration == duration.Duration{}) {
				__antithesis_instrumentation__.Notify(603337)
				convErr = errors.Errorf("interval value %v too small, absolute value must be >= %v", d, time.Microsecond)
			} else {
				__antithesis_instrumentation__.Notify(603338)
			}
			__antithesis_instrumentation__.Notify(603336)
			ts.WallTime = duration.Add(stmtTimestamp, iv.Duration).UnixNano()
			ts.Synthetic = syn
			break
		} else {
			__antithesis_instrumentation__.Notify(603339)
		}
		__antithesis_instrumentation__.Notify(603322)
		convErr = errors.Errorf("value is neither timestamp, decimal, nor interval")
	case *DTimestamp:
		__antithesis_instrumentation__.Notify(603323)
		ts.WallTime = d.UnixNano()
	case *DTimestampTZ:
		__antithesis_instrumentation__.Notify(603324)
		ts.WallTime = d.UnixNano()
	case *DInt:
		__antithesis_instrumentation__.Notify(603325)
		ts.WallTime = int64(*d)
	case *DDecimal:
		__antithesis_instrumentation__.Notify(603326)
		ts, convErr = DecimalToHLC(&d.Decimal)
	case *DInterval:
		__antithesis_instrumentation__.Notify(603327)
		ts.WallTime = duration.Add(stmtTimestamp, d.Duration).UnixNano()
	default:
		__antithesis_instrumentation__.Notify(603328)
		convErr = errors.WithSafeDetails(
			errors.Errorf("expected timestamp, decimal, or interval, got %s", d.ResolvedType()),
			"go type: %T", d)
	}
	__antithesis_instrumentation__.Notify(603315)
	if convErr != nil {
		__antithesis_instrumentation__.Notify(603340)
		return ts, convErr
	} else {
		__antithesis_instrumentation__.Notify(603341)
	}
	__antithesis_instrumentation__.Notify(603316)
	zero := hlc.Timestamp{}
	if ts.EqOrdering(zero) {
		__antithesis_instrumentation__.Notify(603342)
		return ts, errors.Errorf("zero timestamp is invalid")
	} else {
		__antithesis_instrumentation__.Notify(603343)
		if ts.Less(zero) {
			__antithesis_instrumentation__.Notify(603344)
			return ts, errors.Errorf("timestamp before 1970-01-01T00:00:00Z is invalid")
		} else {
			__antithesis_instrumentation__.Notify(603345)
		}
	}
	__antithesis_instrumentation__.Notify(603317)
	return ts, nil
}

func DecimalToHLC(d *apd.Decimal) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(603346)
	if d.Negative {
		__antithesis_instrumentation__.Notify(603354)
		return hlc.Timestamp{}, pgerror.Newf(pgcode.Syntax, "cannot be negative")
	} else {
		__antithesis_instrumentation__.Notify(603355)
	}
	__antithesis_instrumentation__.Notify(603347)
	var integral, fractional apd.Decimal
	d.Modf(&integral, &fractional)
	timestamp, err := integral.Int64()
	if err != nil {
		__antithesis_instrumentation__.Notify(603356)
		return hlc.Timestamp{}, pgerror.Wrapf(err, pgcode.Syntax, "converting timestamp to integer")
	} else {
		__antithesis_instrumentation__.Notify(603357)
	}
	__antithesis_instrumentation__.Notify(603348)
	if fractional.IsZero() {
		__antithesis_instrumentation__.Notify(603358)

		return hlc.Timestamp{WallTime: timestamp}, nil
	} else {
		__antithesis_instrumentation__.Notify(603359)
	}
	__antithesis_instrumentation__.Notify(603349)

	var logical apd.Decimal
	multiplier := apd.New(1, 10)
	condition, err := apd.BaseContext.Mul(&logical, &fractional, multiplier)
	if err != nil {
		__antithesis_instrumentation__.Notify(603360)
		return hlc.Timestamp{}, pgerror.Wrapf(err, pgcode.Syntax, "determining value of logical clock")
	} else {
		__antithesis_instrumentation__.Notify(603361)
	}
	__antithesis_instrumentation__.Notify(603350)
	if _, err := condition.GoError(apd.DefaultTraps); err != nil {
		__antithesis_instrumentation__.Notify(603362)
		return hlc.Timestamp{}, pgerror.Wrapf(err, pgcode.Syntax, "determining value of logical clock")
	} else {
		__antithesis_instrumentation__.Notify(603363)
	}
	__antithesis_instrumentation__.Notify(603351)

	counter, err := logical.Int64()
	if err != nil {
		__antithesis_instrumentation__.Notify(603364)
		return hlc.Timestamp{}, pgerror.Newf(pgcode.Syntax, "logical part has too many digits")
	} else {
		__antithesis_instrumentation__.Notify(603365)
	}
	__antithesis_instrumentation__.Notify(603352)
	if counter > 1<<31 {
		__antithesis_instrumentation__.Notify(603366)
		return hlc.Timestamp{}, pgerror.Newf(pgcode.Syntax, "logical clock too large: %d", counter)
	} else {
		__antithesis_instrumentation__.Notify(603367)
	}
	__antithesis_instrumentation__.Notify(603353)
	return hlc.Timestamp{
		WallTime: timestamp,
		Logical:  int32(counter),
	}, nil
}

func ParseHLC(s string) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(603368)
	dec, _, err := apd.NewFromString(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(603370)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(603371)
	}
	__antithesis_instrumentation__.Notify(603369)
	return DecimalToHLC(dec)
}
