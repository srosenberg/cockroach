package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type setVarNode struct {
	name  string
	local bool
	v     sessionVar

	typedValues []tree.TypedExpr
}

type resetAllNode struct{}

func (p *planner) SetVar(ctx context.Context, n *tree.SetVar) (planNode, error) {
	__antithesis_instrumentation__.Notify(622057)
	if n.ResetAll {
		__antithesis_instrumentation__.Notify(622065)
		return &resetAllNode{}, nil
	} else {
		__antithesis_instrumentation__.Notify(622066)
	}
	__antithesis_instrumentation__.Notify(622058)
	if n.Name == "" {
		__antithesis_instrumentation__.Notify(622067)

		return nil, pgerror.Newf(pgcode.Syntax,
			"invalid variable name: %q", n.Name)
	} else {
		__antithesis_instrumentation__.Notify(622068)
	}
	__antithesis_instrumentation__.Notify(622059)

	name := strings.ToLower(n.Name)
	_, v, err := getSessionVar(name, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(622069)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622070)
	}
	__antithesis_instrumentation__.Notify(622060)
	if _, ok := settings.Lookup(name, settings.LookupForLocalAccess, p.ExecCfg().Codec.ForSystemTenant()); ok {
		__antithesis_instrumentation__.Notify(622071)
		p.BufferClientNotice(
			ctx,
			errors.WithHint(
				pgnotice.Newf("setting custom variable %q", name),
				"did you mean SET CLUSTER SETTING?",
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(622072)
	}
	__antithesis_instrumentation__.Notify(622061)

	var typedValues []tree.TypedExpr
	if len(n.Values) > 0 {
		__antithesis_instrumentation__.Notify(622073)
		isReset := false
		if len(n.Values) == 1 {
			__antithesis_instrumentation__.Notify(622075)
			if _, ok := n.Values[0].(tree.DefaultVal); ok {
				__antithesis_instrumentation__.Notify(622076)

				isReset = true
			} else {
				__antithesis_instrumentation__.Notify(622077)
			}
		} else {
			__antithesis_instrumentation__.Notify(622078)
		}
		__antithesis_instrumentation__.Notify(622074)

		if !isReset {
			__antithesis_instrumentation__.Notify(622079)
			typedValues = make([]tree.TypedExpr, len(n.Values))
			for i, expr := range n.Values {
				__antithesis_instrumentation__.Notify(622080)
				expr = paramparse.UnresolvedNameToStrVal(expr)

				var dummyHelper tree.IndexedVarHelper
				typedValue, err := p.analyzeExpr(
					ctx, expr, nil, dummyHelper, types.String, false, "SET SESSION "+name)
				if err != nil {
					__antithesis_instrumentation__.Notify(622082)
					return nil, wrapSetVarError(err, name, expr.String())
				} else {
					__antithesis_instrumentation__.Notify(622083)
				}
				__antithesis_instrumentation__.Notify(622081)
				typedValues[i] = typedValue
			}
		} else {
			__antithesis_instrumentation__.Notify(622084)
		}
	} else {
		__antithesis_instrumentation__.Notify(622085)
	}
	__antithesis_instrumentation__.Notify(622062)

	if v.Set == nil && func() bool {
		__antithesis_instrumentation__.Notify(622086)
		return v.RuntimeSet == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(622087)
		return v.SetWithPlanner == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(622088)
		return nil, newCannotChangeParameterError(name)
	} else {
		__antithesis_instrumentation__.Notify(622089)
	}
	__antithesis_instrumentation__.Notify(622063)

	if typedValues == nil {
		__antithesis_instrumentation__.Notify(622090)

		if _, ok := p.sessionDataMutatorIterator.defaults[name]; !ok && func() bool {
			__antithesis_instrumentation__.Notify(622091)
			return v.GlobalDefault == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(622092)
			return nil, newCannotChangeParameterError(name)
		} else {
			__antithesis_instrumentation__.Notify(622093)
		}
	} else {
		__antithesis_instrumentation__.Notify(622094)
	}
	__antithesis_instrumentation__.Notify(622064)

	return &setVarNode{name: name, local: n.Local, v: v, typedValues: typedValues}, nil
}

func (n *setVarNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(622095)
	var strVal string

	if _, ok := DummyVars[n.name]; ok {
		__antithesis_instrumentation__.Notify(622098)
		telemetry.Inc(sqltelemetry.DummySessionVarValueCounter(n.name))
		params.p.BufferClientNotice(
			params.ctx,
			pgnotice.NewWithSeverityf("WARNING", "setting session var %q is a no-op", n.name),
		)
	} else {
		__antithesis_instrumentation__.Notify(622099)
	}
	__antithesis_instrumentation__.Notify(622096)
	if n.typedValues != nil {
		__antithesis_instrumentation__.Notify(622100)
		for i, v := range n.typedValues {
			__antithesis_instrumentation__.Notify(622103)
			d, err := v.Eval(params.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(622105)
				return err
			} else {
				__antithesis_instrumentation__.Notify(622106)
			}
			__antithesis_instrumentation__.Notify(622104)
			n.typedValues[i] = d
		}
		__antithesis_instrumentation__.Notify(622101)
		var err error
		if n.v.GetStringVal != nil {
			__antithesis_instrumentation__.Notify(622107)
			strVal, err = n.v.GetStringVal(params.ctx, params.extendedEvalCtx, n.typedValues)
		} else {
			__antithesis_instrumentation__.Notify(622108)

			strVal, err = getStringVal(params.EvalContext(), n.name, n.typedValues)
		}
		__antithesis_instrumentation__.Notify(622102)
		if err != nil {
			__antithesis_instrumentation__.Notify(622109)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622110)
		}
	} else {
		__antithesis_instrumentation__.Notify(622111)

		_, strVal = getSessionVarDefaultString(
			n.name,
			n.v,
			params.p.sessionDataMutatorIterator.sessionDataMutatorBase,
		)
	}
	__antithesis_instrumentation__.Notify(622097)

	return params.p.SetSessionVar(params.ctx, n.name, strVal, n.local)
}

func (p *planner) applyOnSessionDataMutators(
	ctx context.Context, local bool, applyFunc func(m sessionDataMutator) error,
) error {
	__antithesis_instrumentation__.Notify(622112)
	if local {
		__antithesis_instrumentation__.Notify(622114)

		if p.extendedEvalCtx.TxnImplicit {
			__antithesis_instrumentation__.Notify(622116)
			p.BufferClientNotice(
				ctx,
				pgnotice.NewWithSeverityf(
					"WARNING",
					"SET LOCAL can only be used in transaction blocks",
				),
			)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(622117)
		}
		__antithesis_instrumentation__.Notify(622115)
		return p.sessionDataMutatorIterator.applyOnTopMutator(applyFunc)
	} else {
		__antithesis_instrumentation__.Notify(622118)
	}
	__antithesis_instrumentation__.Notify(622113)
	return p.sessionDataMutatorIterator.applyOnEachMutatorError(applyFunc)
}

func getSessionVarDefaultString(
	varName string, v sessionVar, m sessionDataMutatorBase,
) (bool, string) {
	__antithesis_instrumentation__.Notify(622119)
	if defVal, ok := m.defaults[varName]; ok {
		__antithesis_instrumentation__.Notify(622122)
		return true, defVal
	} else {
		__antithesis_instrumentation__.Notify(622123)
	}
	__antithesis_instrumentation__.Notify(622120)
	if v.GlobalDefault != nil {
		__antithesis_instrumentation__.Notify(622124)
		return true, v.GlobalDefault(&m.settings.SV)
	} else {
		__antithesis_instrumentation__.Notify(622125)
	}
	__antithesis_instrumentation__.Notify(622121)
	return false, ""
}

func (n *setVarNode) Next(_ runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(622126)
	return false, nil
}
func (n *setVarNode) Values() tree.Datums     { __antithesis_instrumentation__.Notify(622127); return nil }
func (n *setVarNode) Close(_ context.Context) { __antithesis_instrumentation__.Notify(622128) }

func (n *resetAllNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(622129)
	for varName, v := range varGen {
		__antithesis_instrumentation__.Notify(622132)
		if v.Set == nil && func() bool {
			__antithesis_instrumentation__.Notify(622136)
			return v.RuntimeSet == nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(622137)
			return v.SetWithPlanner == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(622138)
			continue
		} else {
			__antithesis_instrumentation__.Notify(622139)
		}
		__antithesis_instrumentation__.Notify(622133)

		if varName == "role" {
			__antithesis_instrumentation__.Notify(622140)
			continue
		} else {
			__antithesis_instrumentation__.Notify(622141)
		}
		__antithesis_instrumentation__.Notify(622134)
		hasDefault, defVal := getSessionVarDefaultString(
			varName,
			v,
			params.p.sessionDataMutatorIterator.sessionDataMutatorBase,
		)
		if !hasDefault {
			__antithesis_instrumentation__.Notify(622142)
			continue
		} else {
			__antithesis_instrumentation__.Notify(622143)
		}
		__antithesis_instrumentation__.Notify(622135)
		if err := params.p.SetSessionVar(params.ctx, varName, defVal, false); err != nil {
			__antithesis_instrumentation__.Notify(622144)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622145)
		}
	}
	__antithesis_instrumentation__.Notify(622130)
	for varName := range params.SessionData().CustomOptions {
		__antithesis_instrumentation__.Notify(622146)
		_, v, err := getSessionVar(varName, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(622148)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622149)
		}
		__antithesis_instrumentation__.Notify(622147)
		_, defVal := getSessionVarDefaultString(
			varName,
			v,
			params.p.sessionDataMutatorIterator.sessionDataMutatorBase,
		)
		if err := params.p.SetSessionVar(params.ctx, varName, defVal, false); err != nil {
			__antithesis_instrumentation__.Notify(622150)
			return err
		} else {
			__antithesis_instrumentation__.Notify(622151)
		}
	}
	__antithesis_instrumentation__.Notify(622131)
	return nil
}

func (n *resetAllNode) Next(_ runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(622152)
	return false, nil
}
func (n *resetAllNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(622153)
	return nil
}
func (n *resetAllNode) Close(_ context.Context) { __antithesis_instrumentation__.Notify(622154) }

func getStringVal(evalCtx *tree.EvalContext, name string, values []tree.TypedExpr) (string, error) {
	__antithesis_instrumentation__.Notify(622155)
	if len(values) != 1 {
		__antithesis_instrumentation__.Notify(622157)
		return "", newSingleArgVarError(name)
	} else {
		__antithesis_instrumentation__.Notify(622158)
	}
	__antithesis_instrumentation__.Notify(622156)
	return paramparse.DatumAsString(evalCtx, name, values[0])
}

func getIntVal(evalCtx *tree.EvalContext, name string, values []tree.TypedExpr) (int64, error) {
	__antithesis_instrumentation__.Notify(622159)
	if len(values) != 1 {
		__antithesis_instrumentation__.Notify(622161)
		return 0, newSingleArgVarError(name)
	} else {
		__antithesis_instrumentation__.Notify(622162)
	}
	__antithesis_instrumentation__.Notify(622160)
	return paramparse.DatumAsInt(evalCtx, name, values[0])
}

func getFloatVal(evalCtx *tree.EvalContext, name string, values []tree.TypedExpr) (float64, error) {
	__antithesis_instrumentation__.Notify(622163)
	if len(values) != 1 {
		__antithesis_instrumentation__.Notify(622165)
		return 0, newSingleArgVarError(name)
	} else {
		__antithesis_instrumentation__.Notify(622166)
	}
	__antithesis_instrumentation__.Notify(622164)
	return paramparse.DatumAsFloat(evalCtx, name, values[0])
}

func timeZoneVarGetStringVal(
	_ context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr,
) (string, error) {
	__antithesis_instrumentation__.Notify(622167)
	if len(values) != 1 {
		__antithesis_instrumentation__.Notify(622172)
		return "", newSingleArgVarError("timezone")
	} else {
		__antithesis_instrumentation__.Notify(622173)
	}
	__antithesis_instrumentation__.Notify(622168)
	d, err := values[0].Eval(&evalCtx.EvalContext)
	if err != nil {
		__antithesis_instrumentation__.Notify(622174)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(622175)
	}
	__antithesis_instrumentation__.Notify(622169)

	var loc *time.Location
	var offset int64
	switch v := tree.UnwrapDatum(&evalCtx.EvalContext, d).(type) {
	case *tree.DString:
		__antithesis_instrumentation__.Notify(622176)
		location := string(*v)
		loc, err = timeutil.TimeZoneStringToLocation(
			location,
			timeutil.TimeZoneStringToLocationISO8601Standard,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(622183)
			return "", wrapSetVarError(errors.Wrapf(err, "cannot find time zone %q", location), "timezone", values[0].String())
		} else {
			__antithesis_instrumentation__.Notify(622184)
		}

	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(622177)
		offset, _, _, err = v.Duration.Encode()
		if err != nil {
			__antithesis_instrumentation__.Notify(622185)
			return "", wrapSetVarError(err, "timezone", values[0].String())
		} else {
			__antithesis_instrumentation__.Notify(622186)
		}
		__antithesis_instrumentation__.Notify(622178)
		offset /= int64(time.Second)

	case *tree.DInt:
		__antithesis_instrumentation__.Notify(622179)
		offset = int64(*v) * 60 * 60

	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(622180)
		offset = int64(float64(*v) * 60.0 * 60.0)

	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(622181)
		sixty := apd.New(60, 0)
		ed := apd.MakeErrDecimal(tree.ExactCtx)
		ed.Mul(sixty, sixty, sixty)
		ed.Mul(sixty, sixty, &v.Decimal)
		offset = ed.Int64(sixty)
		if ed.Err() != nil {
			__antithesis_instrumentation__.Notify(622187)
			return "", wrapSetVarError(errors.Newf("time zone value %s would overflow an int64", sixty), "timezone", values[0].String())
		} else {
			__antithesis_instrumentation__.Notify(622188)
		}

	default:
		__antithesis_instrumentation__.Notify(622182)
		return "", newVarValueError("timezone", values[0].String())
	}
	__antithesis_instrumentation__.Notify(622170)
	if loc == nil {
		__antithesis_instrumentation__.Notify(622189)
		loc = timeutil.TimeZoneOffsetToLocation(int(offset))
	} else {
		__antithesis_instrumentation__.Notify(622190)
	}
	__antithesis_instrumentation__.Notify(622171)

	return loc.String(), nil
}

func timeZoneVarSet(_ context.Context, m sessionDataMutator, s string) error {
	__antithesis_instrumentation__.Notify(622191)
	loc, err := timeutil.TimeZoneStringToLocation(
		s,
		timeutil.TimeZoneStringToLocationISO8601Standard,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(622193)
		return wrapSetVarError(err, "TimeZone", s)
	} else {
		__antithesis_instrumentation__.Notify(622194)
	}
	__antithesis_instrumentation__.Notify(622192)

	m.SetLocation(loc)
	return nil
}

func makeTimeoutVarGetter(
	varName string,
) func(
	ctx context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr) (string, error) {
	__antithesis_instrumentation__.Notify(622195)
	return func(
		ctx context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr,
	) (string, error) {
		__antithesis_instrumentation__.Notify(622196)
		if len(values) != 1 {
			__antithesis_instrumentation__.Notify(622200)
			return "", newSingleArgVarError(varName)
		} else {
			__antithesis_instrumentation__.Notify(622201)
		}
		__antithesis_instrumentation__.Notify(622197)
		d, err := values[0].Eval(&evalCtx.EvalContext)
		if err != nil {
			__antithesis_instrumentation__.Notify(622202)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(622203)
		}
		__antithesis_instrumentation__.Notify(622198)

		var timeout time.Duration
		switch v := tree.UnwrapDatum(&evalCtx.EvalContext, d).(type) {
		case *tree.DString:
			__antithesis_instrumentation__.Notify(622204)
			return string(*v), nil
		case *tree.DInterval:
			__antithesis_instrumentation__.Notify(622205)
			timeout, err = intervalToDuration(v)
			if err != nil {
				__antithesis_instrumentation__.Notify(622207)
				return "", wrapSetVarError(err, varName, values[0].String())
			} else {
				__antithesis_instrumentation__.Notify(622208)
			}
		case *tree.DInt:
			__antithesis_instrumentation__.Notify(622206)
			timeout = time.Duration(*v) * time.Millisecond
		}
		__antithesis_instrumentation__.Notify(622199)
		return timeout.String(), nil
	}
}

func validateTimeoutVar(
	style duration.IntervalStyle, timeString string, varName string,
) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(622209)
	interval, err := tree.ParseDIntervalWithTypeMetadata(
		style,
		timeString,
		types.IntervalTypeMetadata{
			DurationField: types.IntervalDurationField{
				DurationType: types.IntervalDurationType_MILLISECOND,
			},
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(622213)
		return 0, wrapSetVarError(err, varName, timeString)
	} else {
		__antithesis_instrumentation__.Notify(622214)
	}
	__antithesis_instrumentation__.Notify(622210)
	timeout, err := intervalToDuration(interval)
	if err != nil {
		__antithesis_instrumentation__.Notify(622215)
		return 0, wrapSetVarError(err, varName, timeString)
	} else {
		__antithesis_instrumentation__.Notify(622216)
	}
	__antithesis_instrumentation__.Notify(622211)

	if timeout < 0 {
		__antithesis_instrumentation__.Notify(622217)
		return 0, wrapSetVarError(errors.Newf("%v cannot have a negative duration", redact.SafeString(varName)), varName, timeString)
	} else {
		__antithesis_instrumentation__.Notify(622218)
	}
	__antithesis_instrumentation__.Notify(622212)

	return timeout, nil
}

func stmtTimeoutVarSet(ctx context.Context, m sessionDataMutator, s string) error {
	__antithesis_instrumentation__.Notify(622219)
	timeout, err := validateTimeoutVar(
		m.data.GetIntervalStyle(),
		s,
		"statement_timeout",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(622221)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622222)
	}
	__antithesis_instrumentation__.Notify(622220)

	m.SetStmtTimeout(timeout)
	return nil
}

func lockTimeoutVarSet(ctx context.Context, m sessionDataMutator, s string) error {
	__antithesis_instrumentation__.Notify(622223)
	timeout, err := validateTimeoutVar(
		m.data.GetIntervalStyle(),
		s,
		"lock_timeout",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(622225)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622226)
	}
	__antithesis_instrumentation__.Notify(622224)

	m.SetLockTimeout(timeout)
	return nil
}

func idleInSessionTimeoutVarSet(ctx context.Context, m sessionDataMutator, s string) error {
	__antithesis_instrumentation__.Notify(622227)
	timeout, err := validateTimeoutVar(
		m.data.GetIntervalStyle(),
		s,
		"idle_in_session_timeout",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(622229)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622230)
	}
	__antithesis_instrumentation__.Notify(622228)

	m.SetIdleInSessionTimeout(timeout)
	return nil
}

func idleInTransactionSessionTimeoutVarSet(
	ctx context.Context, m sessionDataMutator, s string,
) error {
	__antithesis_instrumentation__.Notify(622231)
	timeout, err := validateTimeoutVar(
		m.data.GetIntervalStyle(),
		s,
		"idle_in_transaction_session_timeout",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(622233)
		return err
	} else {
		__antithesis_instrumentation__.Notify(622234)
	}
	__antithesis_instrumentation__.Notify(622232)

	m.SetIdleInTransactionSessionTimeout(timeout)
	return nil
}

func intervalToDuration(interval *tree.DInterval) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(622235)
	nanos, _, _, err := interval.Encode()
	if err != nil {
		__antithesis_instrumentation__.Notify(622237)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(622238)
	}
	__antithesis_instrumentation__.Notify(622236)
	return time.Duration(nanos), nil
}

func newSingleArgVarError(varName string) error {
	__antithesis_instrumentation__.Notify(622239)
	return pgerror.Newf(pgcode.InvalidParameterValue,
		"SET %s takes only one argument", varName)
}

func wrapSetVarError(cause error, varName, actualValue string) error {
	__antithesis_instrumentation__.Notify(622240)
	return pgerror.Wrapf(
		cause,
		pgcode.InvalidParameterValue,
		"invalid value for parameter %q: %q",
		redact.SafeString(varName),
		actualValue,
	)
}

func newVarValueError(varName, actualVal string, allowedVals ...string) (err error) {
	__antithesis_instrumentation__.Notify(622241)
	err = pgerror.Newf(pgcode.InvalidParameterValue,
		"invalid value for parameter %q: %q", varName, actualVal)
	if len(allowedVals) > 0 {
		__antithesis_instrumentation__.Notify(622243)
		err = errors.WithHintf(err, "Available values: %s", strings.Join(allowedVals, ","))
	} else {
		__antithesis_instrumentation__.Notify(622244)
	}
	__antithesis_instrumentation__.Notify(622242)
	return err
}

func newCannotChangeParameterError(varName string) error {
	__antithesis_instrumentation__.Notify(622245)
	return pgerror.Newf(pgcode.CantChangeRuntimeParam,
		"parameter %q cannot be changed", varName)
}
