// Package paramparse parses parameters that are set in param lists
// or session vars.
package paramparse

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/errors"
)

func UnresolvedNameToStrVal(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(552229)
	if s, ok := expr.(*tree.UnresolvedName); ok {
		__antithesis_instrumentation__.Notify(552231)
		return tree.NewStrVal(tree.AsStringWithFlags(s, tree.FmtBareIdentifiers))
	} else {
		__antithesis_instrumentation__.Notify(552232)
	}
	__antithesis_instrumentation__.Notify(552230)
	return expr
}

func DatumAsFloat(evalCtx *tree.EvalContext, name string, value tree.TypedExpr) (float64, error) {
	__antithesis_instrumentation__.Notify(552233)
	val, err := value.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(552236)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(552237)
	}
	__antithesis_instrumentation__.Notify(552234)
	switch v := tree.UnwrapDatum(evalCtx, val).(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(552238)
		return strconv.ParseFloat(string(*v), 64)
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(552239)
		return float64(*v), nil
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(552240)
		return float64(*v), nil
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(552241)
		return v.Decimal.Float64()
	}
	__antithesis_instrumentation__.Notify(552235)
	err = pgerror.Newf(pgcode.InvalidParameterValue,
		"parameter %q requires a float value", name)
	err = errors.WithDetailf(err,
		"%s is a %s", value, errors.Safe(val.ResolvedType()))
	return 0, err
}

func DatumAsDuration(
	evalCtx *tree.EvalContext, name string, value tree.TypedExpr,
) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(552242)
	val, err := value.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(552246)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(552247)
	}
	__antithesis_instrumentation__.Notify(552243)
	var d duration.Duration
	switch v := tree.UnwrapDatum(evalCtx, val).(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(552248)
		datum, err := tree.ParseDInterval(evalCtx.SessionData().GetIntervalStyle(), string(*v))
		if err != nil {
			__antithesis_instrumentation__.Notify(552252)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(552253)
		}
		__antithesis_instrumentation__.Notify(552249)
		d = datum.Duration
	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(552250)
		d = v.Duration
	default:
		__antithesis_instrumentation__.Notify(552251)
		err = pgerror.Newf(pgcode.InvalidParameterValue,
			"parameter %q requires a duration value", name)
		err = errors.WithDetailf(err,
			"%s is a %s", value, errors.Safe(val.ResolvedType()))
		return 0, err
	}
	__antithesis_instrumentation__.Notify(552244)

	secs, ok := d.AsInt64()
	if !ok {
		__antithesis_instrumentation__.Notify(552254)
		return 0, pgerror.Newf(
			pgcode.InvalidParameterValue,
			"invalid duration",
		)
	} else {
		__antithesis_instrumentation__.Notify(552255)
	}
	__antithesis_instrumentation__.Notify(552245)
	return time.Duration(secs) * time.Second, nil
}

func DatumAsInt(evalCtx *tree.EvalContext, name string, value tree.TypedExpr) (int64, error) {
	__antithesis_instrumentation__.Notify(552256)
	val, err := value.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(552259)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(552260)
	}
	__antithesis_instrumentation__.Notify(552257)
	iv, ok := tree.AsDInt(val)
	if !ok {
		__antithesis_instrumentation__.Notify(552261)
		err = pgerror.Newf(pgcode.InvalidParameterValue,
			"parameter %q requires an integer value", name)
		err = errors.WithDetailf(err,
			"%s is a %s", value, errors.Safe(val.ResolvedType()))
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(552262)
	}
	__antithesis_instrumentation__.Notify(552258)
	return int64(iv), nil
}

func DatumAsString(evalCtx *tree.EvalContext, name string, value tree.TypedExpr) (string, error) {
	__antithesis_instrumentation__.Notify(552263)
	val, err := value.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(552266)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(552267)
	}
	__antithesis_instrumentation__.Notify(552264)
	s, ok := tree.AsDString(val)
	if !ok {
		__antithesis_instrumentation__.Notify(552268)
		err = pgerror.Newf(pgcode.InvalidParameterValue,
			"parameter %q requires a string value", name)
		err = errors.WithDetailf(err,
			"%s is a %s", value, errors.Safe(val.ResolvedType()))
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(552269)
	}
	__antithesis_instrumentation__.Notify(552265)
	return string(s), nil
}

func GetSingleBool(name string, val tree.Datum) (*tree.DBool, error) {
	__antithesis_instrumentation__.Notify(552270)
	b, ok := val.(*tree.DBool)
	if !ok {
		__antithesis_instrumentation__.Notify(552272)
		err := pgerror.Newf(pgcode.InvalidParameterValue,
			"parameter %q requires a Boolean value", name)
		err = errors.WithDetailf(err,
			"%s is a %s", val, errors.Safe(val.ResolvedType()))
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(552273)
	}
	__antithesis_instrumentation__.Notify(552271)
	return b, nil
}

func ParseBoolVar(varName, val string) (bool, error) {
	__antithesis_instrumentation__.Notify(552274)
	val = strings.ToLower(val)
	switch val {
	case "on":
		__antithesis_instrumentation__.Notify(552277)
		return true, nil
	case "off":
		__antithesis_instrumentation__.Notify(552278)
		return false, nil
	case "yes":
		__antithesis_instrumentation__.Notify(552279)
		return true, nil
	case "no":
		__antithesis_instrumentation__.Notify(552280)
		return false, nil
	default:
		__antithesis_instrumentation__.Notify(552281)
	}
	__antithesis_instrumentation__.Notify(552275)
	b, err := strconv.ParseBool(val)
	if err != nil {
		__antithesis_instrumentation__.Notify(552282)
		return false, pgerror.Newf(pgcode.InvalidParameterValue,
			"parameter \"%s\" requires a Boolean value", varName)
	} else {
		__antithesis_instrumentation__.Notify(552283)
	}
	__antithesis_instrumentation__.Notify(552276)
	return b, nil
}
