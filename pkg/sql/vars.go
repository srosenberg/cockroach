package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/delegate"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/errors"
)

const (
	PgServerVersion = "13.0.0"

	PgServerVersionNum = "130000"

	PgCompatLocale = "C.UTF-8"
)

type getStringValFn = func(
	ctx context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr,
) (string, error)

type sessionVar struct {
	Hidden bool

	Get func(evalCtx *extendedEvalContext) (string, error)

	GetFromSessionData func(sd *sessiondata.SessionData) string

	GetStringVal getStringValFn

	Set func(ctx context.Context, m sessionDataMutator, val string) error

	RuntimeSet func(_ context.Context, evalCtx *extendedEvalContext, local bool, s string) error

	SetWithPlanner func(_ context.Context, p *planner, local bool, s string) error

	GlobalDefault func(sv *settings.Values) string

	Equal func(a, b *sessiondata.SessionData) bool
}

func formatBoolAsPostgresSetting(b bool) string {
	__antithesis_instrumentation__.Notify(631683)
	if b {
		__antithesis_instrumentation__.Notify(631685)
		return "on"
	} else {
		__antithesis_instrumentation__.Notify(631686)
	}
	__antithesis_instrumentation__.Notify(631684)
	return "off"
}

func formatFloatAsPostgresSetting(f float64) string {
	__antithesis_instrumentation__.Notify(631687)
	return strconv.FormatFloat(f, 'G', -1, 64)
}

func makeDummyBooleanSessionVar(
	name string,
	getFunc func(*extendedEvalContext) (string, error),
	setFunc func(sessionDataMutator, bool),
	sv func(_ *settings.Values) string,
) sessionVar {
	__antithesis_instrumentation__.Notify(631688)
	return sessionVar{
		GetStringVal: makePostgresBoolGetStringValFn(name),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631689)
			b, err := paramparse.ParseBoolVar(name, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631691)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631692)
			}
			__antithesis_instrumentation__.Notify(631690)
			setFunc(m, b)
			return nil
		},
		Get:           getFunc,
		GlobalDefault: sv,
	}
}

var varGen = map[string]sessionVar{

	`application_name`: {
		Set: func(
			_ context.Context, m sessionDataMutator, s string,
		) error {
			__antithesis_instrumentation__.Notify(631693)
			m.SetApplicationName(s)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631694)
			return evalCtx.SessionData().ApplicationName, nil
		},
		GetFromSessionData: func(sd *sessiondata.SessionData) string {
			__antithesis_instrumentation__.Notify(631695)
			return sd.ApplicationName
		},
		GlobalDefault: func(_ *settings.Values) string {
			__antithesis_instrumentation__.Notify(631696)
			return ""
		},
		Equal: func(a, b *sessiondata.SessionData) bool {
			__antithesis_instrumentation__.Notify(631697)
			return a.ApplicationName == b.ApplicationName
		},
	},

	`avoid_buffering`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631698)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().AvoidBuffering), nil
		},
		GetStringVal: makePostgresBoolGetStringValFn("avoid_buffering"),
		Set: func(ctx context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631699)
			b, err := paramparse.ParseBoolVar(`avoid_buffering`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631701)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631702)
			}
			__antithesis_instrumentation__.Notify(631700)
			m.SetAvoidBuffering(b)
			return nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631703)
			return "false"
		},
	},

	`bytea_output`: {
		Set: func(
			_ context.Context, m sessionDataMutator, s string,
		) error {
			__antithesis_instrumentation__.Notify(631704)
			mode, ok := lex.BytesEncodeFormatFromString(s)
			if !ok {
				__antithesis_instrumentation__.Notify(631706)
				return newVarValueError(`bytea_output`, s, "hex", "escape", "base64")
			} else {
				__antithesis_instrumentation__.Notify(631707)
			}
			__antithesis_instrumentation__.Notify(631705)
			m.SetBytesEncodeFormat(mode)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631708)
			return evalCtx.SessionData().DataConversionConfig.BytesEncodeFormat.String(), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631709)
			return lex.BytesEncodeHex.String()
		},
	},

	`client_min_messages`: {
		Set: func(
			_ context.Context, m sessionDataMutator, s string,
		) error {
			__antithesis_instrumentation__.Notify(631710)
			severity, ok := pgnotice.ParseDisplaySeverity(s)
			if !ok {
				__antithesis_instrumentation__.Notify(631712)
				return errors.WithHintf(
					pgerror.Newf(
						pgcode.InvalidParameterValue,
						"%s is not supported",
						severity,
					),
					"Valid severities are: %s.",
					strings.Join(pgnotice.ValidDisplaySeverities(), ", "),
				)
			} else {
				__antithesis_instrumentation__.Notify(631713)
			}
			__antithesis_instrumentation__.Notify(631711)
			m.SetNoticeDisplaySeverity(severity)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631714)
			return pgnotice.DisplaySeverity(evalCtx.SessionData().NoticeDisplaySeverity).String(), nil
		},
		GlobalDefault: func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(631715); return "notice" },
	},

	`client_encoding`: {
		Set: func(
			_ context.Context, m sessionDataMutator, s string,
		) error {
			__antithesis_instrumentation__.Notify(631716)
			encoding := builtins.CleanEncodingName(s)
			switch encoding {
			case "utf8", "unicode", "cp65001":
				__antithesis_instrumentation__.Notify(631717)
				return nil
			default:
				__antithesis_instrumentation__.Notify(631718)
				return unimplemented.NewWithIssueDetailf(35882,
					"client_encoding "+encoding,
					"unimplemented client encoding: %q", encoding)
			}
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631719)
			return "UTF8", nil
		},
		GlobalDefault: func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(631720); return "UTF8" },
	},

	`server_encoding`: makeReadOnlyVar("UTF8"),

	`database`: {
		GetStringVal: func(
			ctx context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr,
		) (string, error) {
			__antithesis_instrumentation__.Notify(631721)
			dbName, err := getStringVal(&evalCtx.EvalContext, `database`, values)
			if err != nil {
				__antithesis_instrumentation__.Notify(631725)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(631726)
			}
			__antithesis_instrumentation__.Notify(631722)

			if len(dbName) == 0 && func() bool {
				__antithesis_instrumentation__.Notify(631727)
				return evalCtx.SessionData().SafeUpdates == true
			}() == true {
				__antithesis_instrumentation__.Notify(631728)
				return "", pgerror.DangerousStatementf("SET database to empty string")
			} else {
				__antithesis_instrumentation__.Notify(631729)
			}
			__antithesis_instrumentation__.Notify(631723)

			if len(dbName) != 0 {
				__antithesis_instrumentation__.Notify(631730)

				if _, err := evalCtx.Descs.GetImmutableDatabaseByName(
					ctx, evalCtx.Txn, dbName, tree.DatabaseLookupFlags{Required: true},
				); err != nil {
					__antithesis_instrumentation__.Notify(631731)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(631732)
				}
			} else {
				__antithesis_instrumentation__.Notify(631733)
			}
			__antithesis_instrumentation__.Notify(631724)
			return dbName, nil
		},
		Set: func(_ context.Context, m sessionDataMutator, dbName string) error {
			__antithesis_instrumentation__.Notify(631734)
			m.SetDatabase(dbName)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631735)
			return evalCtx.SessionData().Database, nil
		},
		GlobalDefault: func(_ *settings.Values) string {
			__antithesis_instrumentation__.Notify(631736)

			return ""
		},
	},

	`datestyle`: {
		Set: func(ctx context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631737)
			ds, err := pgdate.ParseDateStyle(s, m.data.GetDateStyle())
			if err != nil {
				__antithesis_instrumentation__.Notify(631742)
				return newVarValueError("DateStyle", s, pgdate.AllowedDateStyles()...)
			} else {
				__antithesis_instrumentation__.Notify(631743)
			}
			__antithesis_instrumentation__.Notify(631738)
			if ds.Style != pgdate.Style_ISO {
				__antithesis_instrumentation__.Notify(631744)
				return unimplemented.NewWithIssue(41773, "only ISO style is supported")
			} else {
				__antithesis_instrumentation__.Notify(631745)
			}
			__antithesis_instrumentation__.Notify(631739)
			allowed := m.data.DateStyleEnabled
			if m.settings.Version.IsActive(ctx, clusterversion.DateStyleIntervalStyleCastRewrite) {
				__antithesis_instrumentation__.Notify(631746)
				allowed = true
			} else {
				__antithesis_instrumentation__.Notify(631747)
			}
			__antithesis_instrumentation__.Notify(631740)
			if ds.Order != pgdate.Order_MDY && func() bool {
				__antithesis_instrumentation__.Notify(631748)
				return !allowed == true
			}() == true {
				__antithesis_instrumentation__.Notify(631749)
				return errors.WithDetailf(
					errors.WithHintf(
						pgerror.Newf(
							pgcode.FeatureNotSupported,
							"setting DateStyle is not enabled",
						),
						"You can enable DateStyle customization for all sessions with the cluster setting %s, or per session using SET datestyle_enabled = true.",
						dateStyleEnabledClusterSetting,
					),
					"Setting DateStyle changes the volatility of timestamp/timestamptz/date::string "+
						"and string::timestamp/timestamptz/date/time/timetz casts from immutable to stable. "+
						"No computed columns, partial indexes, partitions and check constraints can "+
						"use this casts. "+
						"Use to_char_with_style or parse_{timestamp,timestamptz,date,time,timetz} "+
						"instead if you need these casts to work in the aforementioned cases.",
				)
			} else {
				__antithesis_instrumentation__.Notify(631750)
			}
			__antithesis_instrumentation__.Notify(631741)
			m.SetDateStyle(ds)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631751)
			return evalCtx.GetDateStyle().SQLString(), nil
		},
		GetFromSessionData: func(sd *sessiondata.SessionData) string {
			__antithesis_instrumentation__.Notify(631752)
			return sd.GetDateStyle().SQLString()
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631753)
			return dateStyleEnumMap[dateStyle.Get(sv)]
		},
		Equal: func(a, b *sessiondata.SessionData) bool {
			__antithesis_instrumentation__.Notify(631754)
			return a.GetDateStyle() == b.GetDateStyle()
		},
	},

	`datestyle_enabled`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631755)
			if evalCtx.Settings.Version.IsActive(evalCtx.Ctx(), clusterversion.DateStyleIntervalStyleCastRewrite) {
				__antithesis_instrumentation__.Notify(631757)
				return formatBoolAsPostgresSetting(true), nil
			} else {
				__antithesis_instrumentation__.Notify(631758)
			}
			__antithesis_instrumentation__.Notify(631756)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().DateStyleEnabled), nil
		},
		GetStringVal: makePostgresBoolGetStringValFn("datestyle_enabled"),
		SetWithPlanner: func(ctx context.Context, p *planner, local bool, s string) error {
			__antithesis_instrumentation__.Notify(631759)
			b, err := paramparse.ParseBoolVar(`datestyle_enabled`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631764)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631765)
			}
			__antithesis_instrumentation__.Notify(631760)
			if p.execCfg.Settings.Version.IsActive(ctx, clusterversion.DateStyleIntervalStyleCastRewrite) {
				__antithesis_instrumentation__.Notify(631766)
				p.BufferClientNotice(ctx, pgnotice.Newf("ignoring datestyle_enabled setting; it is always true"))
				b = true
			} else {
				__antithesis_instrumentation__.Notify(631767)
			}
			__antithesis_instrumentation__.Notify(631761)
			applyFunc := func(m sessionDataMutator) error {
				__antithesis_instrumentation__.Notify(631768)
				m.SetDateStyleEnabled(b)
				return nil
			}
			__antithesis_instrumentation__.Notify(631762)
			if local {
				__antithesis_instrumentation__.Notify(631769)
				return p.sessionDataMutatorIterator.applyOnTopMutator(applyFunc)
			} else {
				__antithesis_instrumentation__.Notify(631770)
			}
			__antithesis_instrumentation__.Notify(631763)
			return p.sessionDataMutatorIterator.applyOnEachMutatorError(applyFunc)

		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631771)
			return formatBoolAsPostgresSetting(dateStyleEnabled.Get(sv))
		},
	},

	`default_int_size`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631772)
			return strconv.FormatInt(int64(evalCtx.SessionData().DefaultIntSize), 10), nil
		},
		GetStringVal: makeIntGetStringValFn("default_int_size"),
		Set: func(ctx context.Context, m sessionDataMutator, val string) error {
			__antithesis_instrumentation__.Notify(631773)
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(631777)
				return wrapSetVarError(err, "default_int_size", val)
			} else {
				__antithesis_instrumentation__.Notify(631778)
			}
			__antithesis_instrumentation__.Notify(631774)
			if i != 4 && func() bool {
				__antithesis_instrumentation__.Notify(631779)
				return i != 8 == true
			}() == true {
				__antithesis_instrumentation__.Notify(631780)
				return pgerror.New(pgcode.InvalidParameterValue,
					`only 4 or 8 are supported by default_int_size`)
			} else {
				__antithesis_instrumentation__.Notify(631781)
			}
			__antithesis_instrumentation__.Notify(631775)

			if i == 4 {
				__antithesis_instrumentation__.Notify(631782)
				telemetry.Inc(sqltelemetry.DefaultIntSize4Counter)
			} else {
				__antithesis_instrumentation__.Notify(631783)
			}
			__antithesis_instrumentation__.Notify(631776)
			m.SetDefaultIntSize(int32(i))
			return nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631784)
			return strconv.FormatInt(defaultIntSize.Get(sv), 10)
		},
	},

	`default_tablespace`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631785)
			if s != "" {
				__antithesis_instrumentation__.Notify(631787)
				return newVarValueError(`default_tablespace`, s, "")
			} else {
				__antithesis_instrumentation__.Notify(631788)
			}
			__antithesis_instrumentation__.Notify(631786)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631789)
			return "", nil
		},
		GlobalDefault: func(sv *settings.Values) string { __antithesis_instrumentation__.Notify(631790); return "" },
	},

	`default_transaction_isolation`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631791)
			switch strings.ToUpper(s) {
			case `READ UNCOMMITTED`, `READ COMMITTED`, `SNAPSHOT`, `REPEATABLE READ`, `SERIALIZABLE`, `DEFAULT`:
				__antithesis_instrumentation__.Notify(631793)

			default:
				__antithesis_instrumentation__.Notify(631794)
				return newVarValueError(`default_transaction_isolation`, s, "serializable")
			}
			__antithesis_instrumentation__.Notify(631792)

			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631795)
			return "serializable", nil
		},
		GlobalDefault: func(sv *settings.Values) string { __antithesis_instrumentation__.Notify(631796); return "default" },
	},

	`default_transaction_priority`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631797)
			pri, ok := tree.UserPriorityFromString(s)
			if !ok {
				__antithesis_instrumentation__.Notify(631799)
				return newVarValueError(`default_transaction_isolation`, s, "low", "normal", "high")
			} else {
				__antithesis_instrumentation__.Notify(631800)
			}
			__antithesis_instrumentation__.Notify(631798)
			m.SetDefaultTransactionPriority(pri)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631801)
			pri := tree.UserPriority(evalCtx.SessionData().DefaultTxnPriority)
			if pri == tree.UnspecifiedUserPriority {
				__antithesis_instrumentation__.Notify(631803)
				pri = tree.Normal
			} else {
				__antithesis_instrumentation__.Notify(631804)
			}
			__antithesis_instrumentation__.Notify(631802)
			return strings.ToLower(pri.String()), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631805)
			return strings.ToLower(tree.Normal.String())
		},
	},

	`default_transaction_read_only`: {
		GetStringVal: makePostgresBoolGetStringValFn("default_transaction_read_only"),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631806)
			b, err := paramparse.ParseBoolVar("default_transaction_read_only", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631808)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631809)
			}
			__antithesis_instrumentation__.Notify(631807)
			m.SetDefaultTransactionReadOnly(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631810)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().DefaultTxnReadOnly), nil
		},
		GlobalDefault: globalFalse,
	},

	`default_transaction_use_follower_reads`: {
		GetStringVal: makePostgresBoolGetStringValFn("default_transaction_use_follower_reads"),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631811)
			b, err := paramparse.ParseBoolVar("default_transaction_use_follower_reads", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631813)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631814)
			}
			__antithesis_instrumentation__.Notify(631812)
			m.SetDefaultTransactionUseFollowerReads(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631815)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().DefaultTxnUseFollowerReads), nil
		},
		GlobalDefault: globalFalse,
	},

	`disable_plan_gists`: {
		GetStringVal: makePostgresBoolGetStringValFn(`disable_plan_gists`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631816)
			b, err := paramparse.ParseBoolVar("disable_plan_gists", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631818)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631819)
			}
			__antithesis_instrumentation__.Notify(631817)
			m.SetDisablePlanGists(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631820)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().DisablePlanGists), nil
		},
		GlobalDefault: globalFalse,
	},

	`index_recommendations_enabled`: {
		GetStringVal: makePostgresBoolGetStringValFn(`index_recommendations_enabled`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631821)
			b, err := paramparse.ParseBoolVar("index_recommendations_enabled", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631823)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631824)
			}
			__antithesis_instrumentation__.Notify(631822)
			m.SetIndexRecommendationsEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631825)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().IndexRecommendationsEnabled), nil
		},
		GlobalDefault: globalTrue,
	},

	`distsql`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631826)
			mode, ok := sessiondatapb.DistSQLExecModeFromString(s)
			if !ok {
				__antithesis_instrumentation__.Notify(631828)
				return newVarValueError(`distsql`, s, "on", "off", "auto", "always", "2.0-auto", "2.0-off")
			} else {
				__antithesis_instrumentation__.Notify(631829)
			}
			__antithesis_instrumentation__.Notify(631827)
			m.SetDistSQLMode(mode)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631830)
			return evalCtx.SessionData().DistSQLMode.String(), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631831)
			return sessiondatapb.DistSQLExecMode(DistSQLClusterExecMode.Get(sv)).String()
		},
	},

	`distsql_workmem`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631832)
			limit, err := humanizeutil.ParseBytes(s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631835)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631836)
			}
			__antithesis_instrumentation__.Notify(631833)
			if limit <= 0 {
				__antithesis_instrumentation__.Notify(631837)
				return errors.New("distsql_workmem can only be set to a positive value")
			} else {
				__antithesis_instrumentation__.Notify(631838)
			}
			__antithesis_instrumentation__.Notify(631834)
			m.SetDistSQLWorkMem(limit)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631839)
			return string(humanizeutil.IBytes(evalCtx.SessionData().WorkMemLimit)), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631840)
			return string(humanizeutil.IBytes(settingWorkMemBytes.Get(sv)))
		},
	},

	`experimental_distsql_planning`: {
		GetStringVal: makePostgresBoolGetStringValFn(`experimental_distsql_planning`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631841)
			mode, ok := sessiondatapb.ExperimentalDistSQLPlanningModeFromString(s)
			if !ok {
				__antithesis_instrumentation__.Notify(631843)
				return newVarValueError(`experimental_distsql_planning`, s,
					"off", "on", "always")
			} else {
				__antithesis_instrumentation__.Notify(631844)
			}
			__antithesis_instrumentation__.Notify(631842)
			m.SetExperimentalDistSQLPlanning(mode)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631845)
			return evalCtx.SessionData().ExperimentalDistSQLPlanningMode.String(), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631846)
			return sessiondatapb.ExperimentalDistSQLPlanningMode(experimentalDistSQLPlanningClusterMode.Get(sv)).String()
		},
	},

	`disable_partially_distributed_plans`: {
		GetStringVal: makePostgresBoolGetStringValFn(`disable_partially_distributed_plans`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631847)
			b, err := paramparse.ParseBoolVar("disable_partially_distributed_plans", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631849)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631850)
			}
			__antithesis_instrumentation__.Notify(631848)
			m.SetPartiallyDistributedPlansDisabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631851)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().PartiallyDistributedPlansDisabled), nil
		},
		GlobalDefault: globalFalse,
	},

	`enable_zigzag_join`: {
		GetStringVal: makePostgresBoolGetStringValFn(`enable_zigzag_join`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631852)
			b, err := paramparse.ParseBoolVar("enable_zigzag_join", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631854)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631855)
			}
			__antithesis_instrumentation__.Notify(631853)
			m.SetZigzagJoinEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631856)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().ZigzagJoinEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631857)
			return formatBoolAsPostgresSetting(zigzagJoinClusterMode.Get(sv))
		},
	},

	`reorder_joins_limit`: {
		GetStringVal: makeIntGetStringValFn(`reorder_joins_limit`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631858)
			b, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(631861)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631862)
			}
			__antithesis_instrumentation__.Notify(631859)
			if b < 0 {
				__antithesis_instrumentation__.Notify(631863)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"cannot set reorder_joins_limit to a negative value: %d", b)
			} else {
				__antithesis_instrumentation__.Notify(631864)
			}
			__antithesis_instrumentation__.Notify(631860)
			m.SetReorderJoinsLimit(int(b))
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631865)
			return strconv.FormatInt(evalCtx.SessionData().ReorderJoinsLimit, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631866)
			return strconv.FormatInt(ReorderJoinsLimitClusterValue.Get(sv), 10)
		},
	},

	`require_explicit_primary_keys`: {
		GetStringVal: makePostgresBoolGetStringValFn(`require_explicit_primary_keys`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631867)
			b, err := paramparse.ParseBoolVar("require_explicit_primary_key", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631869)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631870)
			}
			__antithesis_instrumentation__.Notify(631868)
			m.SetRequireExplicitPrimaryKeys(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631871)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().RequireExplicitPrimaryKeys), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631872)
			return formatBoolAsPostgresSetting(requireExplicitPrimaryKeysClusterMode.Get(sv))
		},
	},

	`vectorize`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631873)
			mode, ok := sessiondatapb.VectorizeExecModeFromString(s)
			if !ok {
				__antithesis_instrumentation__.Notify(631875)
				return newVarValueError(`vectorize`, s,
					"off", "on", "experimental_always")
			} else {
				__antithesis_instrumentation__.Notify(631876)
			}
			__antithesis_instrumentation__.Notify(631874)
			m.SetVectorize(mode)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631877)
			return evalCtx.SessionData().VectorizeMode.String(), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631878)
			return sessiondatapb.VectorizeExecMode(
				VectorizeClusterMode.Get(sv)).String()
		},
	},

	`testing_vectorize_inject_panics`: {
		GetStringVal: makePostgresBoolGetStringValFn(`testing_vectorize_inject_panics`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631879)
			b, err := paramparse.ParseBoolVar("testing_vectorize_inject_panics", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631881)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631882)
			}
			__antithesis_instrumentation__.Notify(631880)
			m.SetTestingVectorizeInjectPanics(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631883)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().TestingVectorizeInjectPanics), nil
		},
		GlobalDefault: globalFalse,
	},

	`optimizer`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631884)
			if strings.ToUpper(s) != "ON" {
				__antithesis_instrumentation__.Notify(631886)
				return newVarValueError(`optimizer`, s, "on")
			} else {
				__antithesis_instrumentation__.Notify(631887)
			}
			__antithesis_instrumentation__.Notify(631885)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631888)
			return "on", nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631889)
			return "on"
		},
	},

	`foreign_key_cascades_limit`: {
		GetStringVal: makeIntGetStringValFn(`foreign_key_cascades_limit`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631890)
			b, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(631893)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631894)
			}
			__antithesis_instrumentation__.Notify(631891)
			if b < 0 {
				__antithesis_instrumentation__.Notify(631895)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"cannot set foreign_key_cascades_limit to a negative value: %d", b)
			} else {
				__antithesis_instrumentation__.Notify(631896)
			}
			__antithesis_instrumentation__.Notify(631892)
			m.SetOptimizerFKCascadesLimit(int(b))
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631897)
			return strconv.FormatInt(evalCtx.SessionData().OptimizerFKCascadesLimit, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631898)
			return strconv.FormatInt(optDrivenFKCascadesClusterLimit.Get(sv), 10)
		},
	},

	`optimizer_use_histograms`: {
		GetStringVal: makePostgresBoolGetStringValFn(`optimizer_use_histograms`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631899)
			b, err := paramparse.ParseBoolVar("optimizer_use_histograms", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631901)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631902)
			}
			__antithesis_instrumentation__.Notify(631900)
			m.SetOptimizerUseHistograms(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631903)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().OptimizerUseHistograms), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631904)
			return formatBoolAsPostgresSetting(optUseHistogramsClusterMode.Get(sv))
		},
	},

	`optimizer_use_multicol_stats`: {
		GetStringVal: makePostgresBoolGetStringValFn(`optimizer_use_multicol_stats`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631905)
			b, err := paramparse.ParseBoolVar("optimizer_use_multicol_stats", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631907)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631908)
			}
			__antithesis_instrumentation__.Notify(631906)
			m.SetOptimizerUseMultiColStats(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631909)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().OptimizerUseMultiColStats), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631910)
			return formatBoolAsPostgresSetting(optUseMultiColStatsClusterMode.Get(sv))
		},
	},

	`locality_optimized_partitioned_index_scan`: {
		GetStringVal: makePostgresBoolGetStringValFn(`locality_optimized_partitioned_index_scan`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631911)
			b, err := paramparse.ParseBoolVar(`locality_optimized_partitioned_index_scan`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631913)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631914)
			}
			__antithesis_instrumentation__.Notify(631912)
			m.SetLocalityOptimizedSearch(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631915)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().LocalityOptimizedSearch), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631916)
			return formatBoolAsPostgresSetting(localityOptimizedSearchMode.Get(sv))
		},
	},

	`enable_implicit_select_for_update`: {
		GetStringVal: makePostgresBoolGetStringValFn(`enable_implicit_select_for_update`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631917)
			b, err := paramparse.ParseBoolVar("enabled_implicit_select_for_update", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631919)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631920)
			}
			__antithesis_instrumentation__.Notify(631918)
			m.SetImplicitSelectForUpdate(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631921)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().ImplicitSelectForUpdate), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631922)
			return formatBoolAsPostgresSetting(implicitSelectForUpdateClusterMode.Get(sv))
		},
	},

	`enable_insert_fast_path`: {
		GetStringVal: makePostgresBoolGetStringValFn(`enable_insert_fast_path`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631923)
			b, err := paramparse.ParseBoolVar("enable_insert_fast_path", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631925)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631926)
			}
			__antithesis_instrumentation__.Notify(631924)
			m.SetInsertFastPath(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631927)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().InsertFastPath), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631928)
			return formatBoolAsPostgresSetting(insertFastPathClusterMode.Get(sv))
		},
	},

	`serial_normalization`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631929)
			mode, ok := sessiondatapb.SerialNormalizationModeFromString(s)
			if !ok {
				__antithesis_instrumentation__.Notify(631931)
				return newVarValueError(`serial_normalization`, s,
					"rowid", "virtual_sequence", "sql_sequence", "sql_sequence_cached")
			} else {
				__antithesis_instrumentation__.Notify(631932)
			}
			__antithesis_instrumentation__.Notify(631930)
			m.SetSerialNormalizationMode(mode)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631933)
			return evalCtx.SessionData().SerialNormalizationMode.String(), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631934)
			return sessiondatapb.SerialNormalizationMode(
				SerialNormalizationMode.Get(sv)).String()
		},
	},

	`stub_catalog_tables`: {
		GetStringVal: makePostgresBoolGetStringValFn(`stub_catalog_tables`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631935)
			b, err := paramparse.ParseBoolVar("stub_catalog_tables", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631937)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631938)
			}
			__antithesis_instrumentation__.Notify(631936)
			m.SetStubCatalogTablesEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631939)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().StubCatalogTablesEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631940)
			return formatBoolAsPostgresSetting(stubCatalogTablesEnabledClusterValue.Get(sv))
		},
	},

	`extra_float_digits`: {
		GetStringVal: makeIntGetStringValFn(`extra_float_digits`),
		Set: func(
			_ context.Context, m sessionDataMutator, s string,
		) error {
			__antithesis_instrumentation__.Notify(631941)
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(631944)
				return wrapSetVarError(err, "extra_float_digits", s)
			} else {
				__antithesis_instrumentation__.Notify(631945)
			}
			__antithesis_instrumentation__.Notify(631942)

			if i < -15 || func() bool {
				__antithesis_instrumentation__.Notify(631946)
				return i > 3 == true
			}() == true {
				__antithesis_instrumentation__.Notify(631947)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					`%d is outside the valid range for parameter "extra_float_digits" (-15 .. 3)`, i)
			} else {
				__antithesis_instrumentation__.Notify(631948)
			}
			__antithesis_instrumentation__.Notify(631943)
			m.SetExtraFloatDigits(int32(i))
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631949)
			return fmt.Sprintf("%d", evalCtx.SessionData().DataConversionConfig.ExtraFloatDigits), nil
		},
		GlobalDefault: func(sv *settings.Values) string { __antithesis_instrumentation__.Notify(631950); return "0" },
	},

	`force_savepoint_restart`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631951)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().ForceSavepointRestart), nil
		},
		GetStringVal: makePostgresBoolGetStringValFn("force_savepoint_restart"),
		Set: func(_ context.Context, m sessionDataMutator, val string) error {
			__antithesis_instrumentation__.Notify(631952)
			b, err := paramparse.ParseBoolVar("force_savepoint_restart", val)
			if err != nil {
				__antithesis_instrumentation__.Notify(631955)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631956)
			}
			__antithesis_instrumentation__.Notify(631953)
			if b {
				__antithesis_instrumentation__.Notify(631957)
				telemetry.Inc(sqltelemetry.ForceSavepointRestartCounter)
			} else {
				__antithesis_instrumentation__.Notify(631958)
			}
			__antithesis_instrumentation__.Notify(631954)
			m.SetForceSavepointRestart(b)
			return nil
		},
		GlobalDefault: globalFalse,
	},

	`integer_datetimes`: makeReadOnlyVar("on"),

	`intervalstyle`: {
		Set: func(ctx context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631959)
			styleVal, ok := duration.IntervalStyle_value[strings.ToUpper(s)]
			if !ok {
				__antithesis_instrumentation__.Notify(631963)
				validIntervalStyles := make([]string, 0, len(duration.IntervalStyle_value))
				for k := range duration.IntervalStyle_value {
					__antithesis_instrumentation__.Notify(631965)
					validIntervalStyles = append(validIntervalStyles, strings.ToLower(k))
				}
				__antithesis_instrumentation__.Notify(631964)
				return newVarValueError(`IntervalStyle`, s, validIntervalStyles...)
			} else {
				__antithesis_instrumentation__.Notify(631966)
			}
			__antithesis_instrumentation__.Notify(631960)
			style := duration.IntervalStyle(styleVal)
			allowed := m.data.IntervalStyleEnabled
			if m.settings.Version.IsActive(ctx, clusterversion.DateStyleIntervalStyleCastRewrite) {
				__antithesis_instrumentation__.Notify(631967)
				allowed = true
			} else {
				__antithesis_instrumentation__.Notify(631968)
			}
			__antithesis_instrumentation__.Notify(631961)
			if style != duration.IntervalStyle_POSTGRES && func() bool {
				__antithesis_instrumentation__.Notify(631969)
				return !allowed == true
			}() == true {
				__antithesis_instrumentation__.Notify(631970)
				return errors.WithDetailf(
					errors.WithHintf(
						pgerror.Newf(
							pgcode.FeatureNotSupported,
							"setting IntervalStyle is not enabled",
						),
						"You can enable IntervalStyle customization for all sessions with the cluster setting %s, or per session using SET intervalstyle_enabled = true.",
						intervalStyleEnabledClusterSetting,
					),
					"Setting IntervalStyle changes the volatility of string::interval or interval::string "+
						"casts from immutable to stable. No computed columns, partial indexes, partitions "+
						"and check constraints can use this casts. "+
						"Use to_char_with_style or parse_interval instead if you need these casts to work "+
						"in the aforementioned cases.",
				)
			} else {
				__antithesis_instrumentation__.Notify(631971)
			}
			__antithesis_instrumentation__.Notify(631962)
			m.SetIntervalStyle(style)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631972)
			return strings.ToLower(evalCtx.SessionData().GetIntervalStyle().String()), nil
		},
		GetFromSessionData: func(sd *sessiondata.SessionData) string {
			__antithesis_instrumentation__.Notify(631973)
			return strings.ToLower(sd.GetIntervalStyle().String())
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631974)
			return strings.ToLower(duration.IntervalStyle_name[int32(intervalStyle.Get(sv))])
		},
		Equal: func(a, b *sessiondata.SessionData) bool {
			__antithesis_instrumentation__.Notify(631975)
			return a.GetIntervalStyle() == b.GetIntervalStyle()
		},
	},

	`intervalstyle_enabled`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631976)
			if evalCtx.Settings.Version.IsActive(evalCtx.Ctx(), clusterversion.DateStyleIntervalStyleCastRewrite) {
				__antithesis_instrumentation__.Notify(631978)
				return formatBoolAsPostgresSetting(true), nil
			} else {
				__antithesis_instrumentation__.Notify(631979)
			}
			__antithesis_instrumentation__.Notify(631977)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().IntervalStyleEnabled), nil
		},
		GetStringVal: makePostgresBoolGetStringValFn("intervalstyle_enabled"),
		SetWithPlanner: func(ctx context.Context, p *planner, local bool, s string) error {
			__antithesis_instrumentation__.Notify(631980)
			b, err := paramparse.ParseBoolVar(`intervalstyle_enabled`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(631985)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631986)
			}
			__antithesis_instrumentation__.Notify(631981)
			if p.execCfg.Settings.Version.IsActive(ctx, clusterversion.DateStyleIntervalStyleCastRewrite) {
				__antithesis_instrumentation__.Notify(631987)
				p.BufferClientNotice(ctx, pgnotice.Newf("ignoring intervalstyle_enabled setting; it is always true"))
				b = true
			} else {
				__antithesis_instrumentation__.Notify(631988)
			}
			__antithesis_instrumentation__.Notify(631982)
			applyFunc := func(m sessionDataMutator) error {
				__antithesis_instrumentation__.Notify(631989)
				m.SetIntervalStyleEnabled(b)
				return nil
			}
			__antithesis_instrumentation__.Notify(631983)
			if local {
				__antithesis_instrumentation__.Notify(631990)
				return p.sessionDataMutatorIterator.applyOnTopMutator(applyFunc)
			} else {
				__antithesis_instrumentation__.Notify(631991)
			}
			__antithesis_instrumentation__.Notify(631984)
			return p.sessionDataMutatorIterator.applyOnEachMutatorError(applyFunc)
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(631992)
			return formatBoolAsPostgresSetting(intervalStyleEnabled.Get(sv))
		},
	},

	`is_superuser`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631993)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().IsSuperuser), nil
		},
		GetFromSessionData: func(sd *sessiondata.SessionData) string {
			__antithesis_instrumentation__.Notify(631994)
			return formatBoolAsPostgresSetting(sd.IsSuperuser)
		},
		GetStringVal:  makePostgresBoolGetStringValFn("is_superuser"),
		GlobalDefault: globalFalse,
		Equal: func(a, b *sessiondata.SessionData) bool {
			__antithesis_instrumentation__.Notify(631995)
			return a.IsSuperuser == b.IsSuperuser
		},
	},

	`large_full_scan_rows`: {
		GetStringVal: makeFloatGetStringValFn(`large_full_scan_rows`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(631996)
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(631998)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631999)
			}
			__antithesis_instrumentation__.Notify(631997)
			m.SetLargeFullScanRows(f)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632000)
			return formatFloatAsPostgresSetting(evalCtx.SessionData().LargeFullScanRows), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632001)
			return formatFloatAsPostgresSetting(largeFullScanRows.Get(sv))
		},
	},

	`locality`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632002)
			return evalCtx.Locality.String(), nil
		},
	},

	`lock_timeout`: {
		GetStringVal: makeTimeoutVarGetter(`lock_timeout`),
		Set:          lockTimeoutVarSet,
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632003)
			ms := evalCtx.SessionData().LockTimeout.Nanoseconds() / int64(time.Millisecond)
			return strconv.FormatInt(ms, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632004)
			return clusterLockTimeout.String(sv)
		},
	},

	`default_table_access_method`: makeCompatStringVar(`default_table_access_method`, `heap`),

	`backslash_quote`: makeCompatStringVar(`backslash_quote`, `safe_encoding`),

	`default_with_oids`: makeCompatBoolVar(`default_with_oids`, false, false),

	`xmloption`: makeCompatStringVar(`xmloption`, `content`),

	`max_identifier_length`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632005)
			return "128", nil
		},
	},

	`max_index_keys`: makeReadOnlyVar("32"),

	`node_id`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632006)
			nodeID, _ := evalCtx.NodeID.OptionalNodeID()
			return fmt.Sprintf("%d", nodeID), nil
		},
	},

	`results_buffer_size`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632007)
			return strconv.FormatInt(evalCtx.SessionData().ResultsBufferSize, 10), nil
		},
	},

	`sql_safe_updates`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632008)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().SafeUpdates), nil
		},
		GetStringVal: makePostgresBoolGetStringValFn("sql_safe_updates"),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632009)
			b, err := paramparse.ParseBoolVar("sql_safe_updates", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632011)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632012)
			}
			__antithesis_instrumentation__.Notify(632010)
			m.SetSafeUpdates(b)
			return nil
		},
		GlobalDefault: globalFalse,
	},

	`check_function_bodies`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632013)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().CheckFunctionBodies), nil
		},
		GetStringVal: makePostgresBoolGetStringValFn("check_function_bodies"),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632014)
			b, err := paramparse.ParseBoolVar("check_function_bodies", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632016)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632017)
			}
			__antithesis_instrumentation__.Notify(632015)
			m.SetCheckFunctionBodies(b)
			return nil
		},
		GlobalDefault: globalTrue,
	},

	`prefer_lookup_joins_for_fks`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632018)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().PreferLookupJoinsForFKs), nil
		},
		GetStringVal: makePostgresBoolGetStringValFn("prefer_lookup_joins_for_fks"),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632019)
			b, err := paramparse.ParseBoolVar("prefer_lookup_joins_for_fks", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632021)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632022)
			}
			__antithesis_instrumentation__.Notify(632020)
			m.SetPreferLookupJoinsForFKs(b)
			return nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632023)
			return formatBoolAsPostgresSetting(preferLookupJoinsForFKs.Get(sv))
		},
	},

	`role`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632024)
			if evalCtx.SessionData().SessionUserProto == "" {
				__antithesis_instrumentation__.Notify(632026)
				return security.NoneRole, nil
			} else {
				__antithesis_instrumentation__.Notify(632027)
			}
			__antithesis_instrumentation__.Notify(632025)
			return evalCtx.SessionData().User().Normalized(), nil
		},

		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632028)
			return security.NoneRole
		},
	},

	`search_path`: {
		GetStringVal: func(
			_ context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr,
		) (string, error) {
			__antithesis_instrumentation__.Notify(632029)
			comma := ""
			var buf bytes.Buffer
			for _, v := range values {
				__antithesis_instrumentation__.Notify(632031)
				s, err := paramparse.DatumAsString(&evalCtx.EvalContext, "search_path", v)
				if err != nil {
					__antithesis_instrumentation__.Notify(632034)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(632035)
				}
				__antithesis_instrumentation__.Notify(632032)
				if strings.Contains(s, ",") {
					__antithesis_instrumentation__.Notify(632036)

					return "",
						errors.WithHintf(unimplemented.NewWithIssuef(53971,
							`schema name %q has commas so is not supported in search_path.`, s),
							`Did you mean to omit quotes? SET search_path = %s`, s)
				} else {
					__antithesis_instrumentation__.Notify(632037)
				}
				__antithesis_instrumentation__.Notify(632033)
				buf.WriteString(comma)
				buf.WriteString(s)
				comma = ","
			}
			__antithesis_instrumentation__.Notify(632030)
			return buf.String(), nil
		},
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632038)
			paths := strings.Split(s, ",")
			m.UpdateSearchPath(paths)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632039)
			return evalCtx.SessionData().SearchPath.SQLIdentifiers(), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632040)
			return sessiondata.DefaultSearchPath.String()
		},
	},

	`server_version`: makeReadOnlyVar(PgServerVersion),

	`server_version_num`: makeReadOnlyVar(PgServerVersionNum),

	`lc_collate`: makeReadOnlyVar(PgCompatLocale),

	`lc_ctype`: makeReadOnlyVar(PgCompatLocale),

	`lc_messages`: makeCompatStringVar("lc_messages", PgCompatLocale),

	`lc_monetary`: makeCompatStringVar("lc_monetary", PgCompatLocale),

	`lc_numeric`: makeCompatStringVar("lc_numeric", PgCompatLocale),

	`lc_time`: makeCompatStringVar("lc_time", PgCompatLocale),

	`ssl_renegotiation_limit`: {
		Hidden:       true,
		GetStringVal: makeIntGetStringValFn(`ssl_renegotiation_limit`),
		Get: func(_ *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632041)
			return "0", nil
		},
		GlobalDefault: func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(632042); return "0" },
		Set: func(_ context.Context, _ sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632043)
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(632046)
				return wrapSetVarError(err, "ssl_renegotiation_limit", s)
			} else {
				__antithesis_instrumentation__.Notify(632047)
			}
			__antithesis_instrumentation__.Notify(632044)
			if i != 0 {
				__antithesis_instrumentation__.Notify(632048)

				return newVarValueError("ssl_renegotiation_limit", s, "0")
			} else {
				__antithesis_instrumentation__.Notify(632049)
			}
			__antithesis_instrumentation__.Notify(632045)
			return nil
		},
	},

	`crdb_version`: makeReadOnlyVarWithFn(func() string {
		__antithesis_instrumentation__.Notify(632050)
		return build.GetInfo().Short()
	}),

	`session_id`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632051)
			return evalCtx.SessionID.String(), nil
		},
	},

	`session_user`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632052)
			return evalCtx.SessionData().User().Normalized(), nil
		},
	},

	`session_authorization`: {
		Hidden: true,
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632053)
			return evalCtx.SessionData().User().Normalized(), nil
		},
	},

	`password_encryption`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632054)
			return security.GetConfiguredPasswordHashMethod(evalCtx.Ctx(), &evalCtx.Settings.SV).String(), nil
		},
		SetWithPlanner: func(ctx context.Context, p *planner, local bool, val string) error {
			__antithesis_instrumentation__.Notify(632055)
			method := security.GetConfiguredPasswordHashMethod(ctx, &p.ExecCfg().Settings.SV)
			if val != method.String() {
				__antithesis_instrumentation__.Notify(632057)
				return newCannotChangeParameterError("password_encryption")
			} else {
				__antithesis_instrumentation__.Notify(632058)
			}
			__antithesis_instrumentation__.Notify(632056)
			return nil
		},
	},

	`standard_conforming_strings`: makeCompatBoolVar(`standard_conforming_strings`, true, false),

	`escape_string_warning`: makeCompatBoolVar(`escape_string_warning`, true, true),

	`synchronize_seqscans`: makeCompatBoolVar(`synchronize_seqscans`, true, true),

	`row_security`: makeCompatBoolVar(`row_security`, false, true),

	`statement_timeout`: {
		GetStringVal: makeTimeoutVarGetter(`statement_timeout`),
		Set:          stmtTimeoutVarSet,
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632059)
			ms := evalCtx.SessionData().StmtTimeout.Nanoseconds() / int64(time.Millisecond)
			return strconv.FormatInt(ms, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632060)
			return clusterStatementTimeout.String(sv)
		},
	},

	`idle_in_session_timeout`: {
		GetStringVal: makeTimeoutVarGetter(`idle_in_session_timeout`),
		Set:          idleInSessionTimeoutVarSet,
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632061)
			ms := evalCtx.SessionData().IdleInSessionTimeout.Nanoseconds() / int64(time.Millisecond)
			return strconv.FormatInt(ms, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632062)
			return clusterIdleInSessionTimeout.String(sv)
		},
	},

	`idle_in_transaction_session_timeout`: {
		GetStringVal: makeTimeoutVarGetter(`idle_in_transaction_session_timeout`),
		Set:          idleInTransactionSessionTimeoutVarSet,
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632063)
			ms := evalCtx.SessionData().IdleInTransactionSessionTimeout.Nanoseconds() / int64(time.Millisecond)
			return strconv.FormatInt(ms, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632064)
			return clusterIdleInTransactionSessionTimeout.String(sv)
		},
	},

	`timezone`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632065)
			return sessionDataTimeZoneFormat(evalCtx.SessionData().GetLocation()), nil
		},
		GetFromSessionData: func(sd *sessiondata.SessionData) string {
			__antithesis_instrumentation__.Notify(632066)
			return sessionDataTimeZoneFormat(sd.GetLocation())
		},
		GetStringVal:  timeZoneVarGetStringVal,
		Set:           timeZoneVarSet,
		GlobalDefault: func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(632067); return "UTC" },
		Equal: func(a, b *sessiondata.SessionData) bool {
			__antithesis_instrumentation__.Notify(632068)

			return a.GetLocation().String() == b.GetLocation().String()
		},
	},

	`transaction_isolation`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632069)
			return "serializable", nil
		},
		RuntimeSet: func(_ context.Context, evalCtx *extendedEvalContext, local bool, s string) error {
			__antithesis_instrumentation__.Notify(632070)
			_, ok := tree.IsolationLevelMap[s]
			if !ok {
				__antithesis_instrumentation__.Notify(632072)
				return newVarValueError(`transaction_isolation`, s, "serializable")
			} else {
				__antithesis_instrumentation__.Notify(632073)
			}
			__antithesis_instrumentation__.Notify(632071)
			return nil
		},
		GlobalDefault: func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(632074); return "serializable" },
	},

	`transaction_priority`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632075)
			return evalCtx.Txn.UserPriority().String(), nil
		},
	},

	`transaction_status`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632076)
			return evalCtx.TxnState, nil
		},
	},

	`transaction_read_only`: {
		GetStringVal: makePostgresBoolGetStringValFn("transaction_read_only"),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632077)
			b, err := paramparse.ParseBoolVar("transaction_read_only", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632079)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632080)
			}
			__antithesis_instrumentation__.Notify(632078)
			m.SetReadOnly(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632081)
			return formatBoolAsPostgresSetting(evalCtx.TxnReadOnly), nil
		},
		GlobalDefault: globalFalse,
	},

	`tracing`: {
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632082)
			sessTracing := evalCtx.Tracing
			if sessTracing.Enabled() {
				__antithesis_instrumentation__.Notify(632084)
				val := "on"
				if sessTracing.KVTracingEnabled() {
					__antithesis_instrumentation__.Notify(632086)
					val += ", kv"
				} else {
					__antithesis_instrumentation__.Notify(632087)
				}
				__antithesis_instrumentation__.Notify(632085)
				return val, nil
			} else {
				__antithesis_instrumentation__.Notify(632088)
			}
			__antithesis_instrumentation__.Notify(632083)
			return "off", nil
		},
	},

	`allow_prepare_as_opt_plan`: {
		Hidden: true,
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632089)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().AllowPrepareAsOptPlan), nil
		},
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632090)
			b, err := paramparse.ParseBoolVar("allow_prepare_as_opt_plan", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632092)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632093)
			}
			__antithesis_instrumentation__.Notify(632091)
			m.SetAllowPrepareAsOptPlan(b)
			return nil
		},
		GlobalDefault: globalFalse,
	},

	`save_tables_prefix`: {
		Hidden: true,
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632094)
			return evalCtx.SessionData().SaveTablesPrefix, nil
		},
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632095)
			m.SetSaveTablesPrefix(s)
			return nil
		},
		GlobalDefault: func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(632096); return "" },
	},

	`experimental_enable_temp_tables`: {
		GetStringVal: makePostgresBoolGetStringValFn(`experimental_enable_temp_tables`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632097)
			b, err := paramparse.ParseBoolVar("experimental_enable_temp_tables", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632099)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632100)
			}
			__antithesis_instrumentation__.Notify(632098)
			m.SetTempTablesEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632101)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().TempTablesEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632102)
			return formatBoolAsPostgresSetting(temporaryTablesEnabledClusterMode.Get(sv))
		},
	},

	`enable_multiregion_placement_policy`: {
		GetStringVal: makePostgresBoolGetStringValFn(`enable_multiregion_placement_policy`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632103)
			b, err := paramparse.ParseBoolVar("enable_multiregion_placement_policy", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632105)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632106)
			}
			__antithesis_instrumentation__.Notify(632104)
			m.SetPlacementEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632107)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().PlacementEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632108)
			return formatBoolAsPostgresSetting(placementEnabledClusterMode.Get(sv))
		},
	},

	`experimental_enable_auto_rehoming`: {
		GetStringVal: makePostgresBoolGetStringValFn(`experimental_enable_auto_rehoming`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632109)
			b, err := paramparse.ParseBoolVar("experimental_enable_auto_rehoming", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632111)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632112)
			}
			__antithesis_instrumentation__.Notify(632110)
			m.SetAutoRehomingEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632113)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().AutoRehomingEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632114)
			return formatBoolAsPostgresSetting(autoRehomingEnabledClusterMode.Get(sv))
		},
	},

	`on_update_rehome_row_enabled`: {
		GetStringVal: makePostgresBoolGetStringValFn(`on_update_rehome_row_enabled`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632115)
			b, err := paramparse.ParseBoolVar("on_update_rehome_row_enabled", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632117)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632118)
			}
			__antithesis_instrumentation__.Notify(632116)
			m.SetOnUpdateRehomeRowEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632119)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().OnUpdateRehomeRowEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632120)
			return formatBoolAsPostgresSetting(onUpdateRehomeRowEnabledClusterMode.Get(sv))
		},
	},

	`experimental_enable_implicit_column_partitioning`: {
		GetStringVal: makePostgresBoolGetStringValFn(`experimental_enable_implicit_column_partitioning`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632121)
			b, err := paramparse.ParseBoolVar("experimental_enable_implicit_column_partitioning", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632123)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632124)
			}
			__antithesis_instrumentation__.Notify(632122)
			m.SetImplicitColumnPartitioningEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632125)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().ImplicitColumnPartitioningEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632126)
			return formatBoolAsPostgresSetting(implicitColumnPartitioningEnabledClusterMode.Get(sv))
		},
	},

	`enable_drop_enum_value`: makeBackwardsCompatBoolVar(
		"enable_drop_enum_value", true,
	),

	`override_multi_region_zone_config`: {
		GetStringVal: makePostgresBoolGetStringValFn(`override_multi_region_zone_config`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632127)
			b, err := paramparse.ParseBoolVar("override_multi_region_zone_config", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632129)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632130)
			}
			__antithesis_instrumentation__.Notify(632128)
			m.SetOverrideMultiRegionZoneConfigEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632131)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().OverrideMultiRegionZoneConfigEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632132)
			return formatBoolAsPostgresSetting(overrideMultiRegionZoneConfigClusterMode.Get(sv))
		},
	},

	`experimental_enable_hash_sharded_indexes`: makeBackwardsCompatBoolVar(
		"experimental_enable_hash_sharded_indexes", true,
	),

	`disallow_full_table_scans`: {
		GetStringVal: makePostgresBoolGetStringValFn(`disallow_full_table_scans`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632133)
			b, err := paramparse.ParseBoolVar(`disallow_full_table_scans`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632135)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632136)
			}
			__antithesis_instrumentation__.Notify(632134)
			m.SetDisallowFullTableScans(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632137)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().DisallowFullTableScans), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632138)
			return formatBoolAsPostgresSetting(disallowFullTableScans.Get(sv))
		},
	},

	`enable_experimental_alter_column_type_general`: {
		GetStringVal: makePostgresBoolGetStringValFn(`enable_experimental_alter_column_type_general`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632139)
			b, err := paramparse.ParseBoolVar("enable_experimental_alter_column_type_general", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632141)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632142)
			}
			__antithesis_instrumentation__.Notify(632140)
			m.SetAlterColumnTypeGeneral(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632143)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().AlterColumnTypeGeneralEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632144)
			return formatBoolAsPostgresSetting(experimentalAlterColumnTypeGeneralMode.Get(sv))
		},
	},

	`experimental_enable_unique_without_index_constraints`: {
		GetStringVal: makePostgresBoolGetStringValFn(`experimental_enable_unique_without_index_constraints`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632145)
			b, err := paramparse.ParseBoolVar(`experimental_enable_unique_without_index_constraints`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632147)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632148)
			}
			__antithesis_instrumentation__.Notify(632146)
			m.SetUniqueWithoutIndexConstraints(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632149)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().EnableUniqueWithoutIndexConstraints), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632150)
			return formatBoolAsPostgresSetting(experimentalUniqueWithoutIndexConstraintsMode.Get(sv))
		},
	},

	`use_declarative_schema_changer`: {
		GetStringVal: makePostgresBoolGetStringValFn(`use_declarative_schema_changer`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632151)
			mode, ok := sessiondatapb.NewSchemaChangerModeFromString(s)
			if !ok {
				__antithesis_instrumentation__.Notify(632153)
				return newVarValueError(`use_declarative_schema_changer`, s,
					"off", "on", "unsafe", "unsafe_always")
			} else {
				__antithesis_instrumentation__.Notify(632154)
			}
			__antithesis_instrumentation__.Notify(632152)
			m.SetUseNewSchemaChanger(mode)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632155)
			return evalCtx.SessionData().NewSchemaChangerMode.String(), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632156)
			return sessiondatapb.NewSchemaChangerMode(experimentalUseNewSchemaChanger.Get(sv)).String()
		},
	},

	`enable_experimental_stream_replication`: {
		GetStringVal: makePostgresBoolGetStringValFn(`enable_experimental_stream_replication`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632157)
			b, err := paramparse.ParseBoolVar(`enable_experimental_stream_replication`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632159)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632160)
			}
			__antithesis_instrumentation__.Notify(632158)
			m.SetStreamReplicationEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632161)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().EnableStreamReplication), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632162)
			return formatBoolAsPostgresSetting(experimentalStreamReplicationEnabled.Get(sv))
		},
	},

	`experimental_computed_column_rewrites`: {
		Hidden:       true,
		GetStringVal: makePostgresBoolGetStringValFn(`experimental_computed_column_rewrites`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632163)
			_, err := schemaexpr.ParseComputedColumnRewrites(s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632165)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632166)
			}
			__antithesis_instrumentation__.Notify(632164)
			m.SetExperimentalComputedColumnRewrites(s)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632167)
			return evalCtx.SessionData().ExperimentalComputedColumnRewrites, nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632168)
			return experimentalComputedColumnRewrites.Get(sv)
		},
	},

	`null_ordered_last`: {
		GetStringVal: makePostgresBoolGetStringValFn(`null_ordered_last`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632169)
			b, err := paramparse.ParseBoolVar(`null_ordered_last`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632171)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632172)
			}
			__antithesis_instrumentation__.Notify(632170)
			m.SetNullOrderedLast(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632173)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().NullOrderedLast), nil
		},
		GlobalDefault: globalFalse,
	},

	`propagate_input_ordering`: {
		GetStringVal: makePostgresBoolGetStringValFn(`propagate_input_ordering`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632174)
			b, err := paramparse.ParseBoolVar(`propagate_input_ordering`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632176)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632177)
			}
			__antithesis_instrumentation__.Notify(632175)
			m.SetPropagateInputOrdering(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632178)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().PropagateInputOrdering), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632179)
			return formatBoolAsPostgresSetting(propagateInputOrdering.Get(sv))
		},
	},

	`transaction_rows_written_log`: {
		GetStringVal: makeIntGetStringValFn(`transaction_rows_written_log`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632180)
			b, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(632183)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632184)
			}
			__antithesis_instrumentation__.Notify(632181)
			if b < 0 {
				__antithesis_instrumentation__.Notify(632185)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"cannot set transaction_rows_written_log to a negative value: %d", b)
			} else {
				__antithesis_instrumentation__.Notify(632186)
			}
			__antithesis_instrumentation__.Notify(632182)
			m.SetTxnRowsWrittenLog(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632187)
			return strconv.FormatInt(evalCtx.SessionData().TxnRowsWrittenLog, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632188)
			return strconv.FormatInt(txnRowsWrittenLog.Get(sv), 10)
		},
	},

	`transaction_rows_written_err`: {
		GetStringVal: makeIntGetStringValFn(`transaction_rows_written_err`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632189)
			b, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(632192)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632193)
			}
			__antithesis_instrumentation__.Notify(632190)
			if b < 0 {
				__antithesis_instrumentation__.Notify(632194)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"cannot set transaction_rows_written_err to a negative value: %d", b)
			} else {
				__antithesis_instrumentation__.Notify(632195)
			}
			__antithesis_instrumentation__.Notify(632191)
			m.SetTxnRowsWrittenErr(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632196)
			return strconv.FormatInt(evalCtx.SessionData().TxnRowsWrittenErr, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632197)
			return strconv.FormatInt(txnRowsWrittenErr.Get(sv), 10)
		},
	},

	`transaction_rows_read_log`: {
		GetStringVal: makeIntGetStringValFn(`transaction_rows_read_log`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632198)
			b, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(632201)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632202)
			}
			__antithesis_instrumentation__.Notify(632199)
			if b < 0 {
				__antithesis_instrumentation__.Notify(632203)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"cannot set transaction_rows_read_log to a negative value: %d", b)
			} else {
				__antithesis_instrumentation__.Notify(632204)
			}
			__antithesis_instrumentation__.Notify(632200)
			m.SetTxnRowsReadLog(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632205)
			return strconv.FormatInt(evalCtx.SessionData().TxnRowsReadLog, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632206)
			return strconv.FormatInt(txnRowsReadLog.Get(sv), 10)
		},
	},

	`transaction_rows_read_err`: {
		GetStringVal: makeIntGetStringValFn(`transaction_rows_read_err`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632207)
			b, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(632210)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632211)
			}
			__antithesis_instrumentation__.Notify(632208)
			if b < 0 {
				__antithesis_instrumentation__.Notify(632212)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"cannot set transaction_rows_read_err to a negative value: %d", b)
			} else {
				__antithesis_instrumentation__.Notify(632213)
			}
			__antithesis_instrumentation__.Notify(632209)
			m.SetTxnRowsReadErr(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632214)
			return strconv.FormatInt(evalCtx.SessionData().TxnRowsReadErr, 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632215)
			return strconv.FormatInt(txnRowsReadErr.Get(sv), 10)
		},
	},

	`inject_retry_errors_enabled`: {
		GetStringVal: makePostgresBoolGetStringValFn(`retry_errors_enabled`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632216)
			b, err := paramparse.ParseBoolVar("inject_retry_errors_enabled", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632218)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632219)
			}
			__antithesis_instrumentation__.Notify(632217)
			m.SetInjectRetryErrorsEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632220)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().InjectRetryErrorsEnabled), nil
		},
		GlobalDefault: globalFalse,
	},

	`join_reader_ordering_strategy_batch_size`: {
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632221)
			size, err := humanizeutil.ParseBytes(s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632224)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632225)
			}
			__antithesis_instrumentation__.Notify(632222)
			if size <= 0 {
				__antithesis_instrumentation__.Notify(632226)
				return errors.New("join_reader_ordering_strategy_batch_size can only be set to a positive value")
			} else {
				__antithesis_instrumentation__.Notify(632227)
			}
			__antithesis_instrumentation__.Notify(632223)
			m.SetJoinReaderOrderingStrategyBatchSize(size)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632228)
			return string(humanizeutil.IBytes(evalCtx.SessionData().JoinReaderOrderingStrategyBatchSize)), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632229)
			return string(humanizeutil.IBytes(rowexec.JoinReaderOrderingStrategyBatchSize.Get(sv)))
		},
	},

	`parallelize_multi_key_lookup_joins_enabled`: {
		GetStringVal: makePostgresBoolGetStringValFn(`parallelize_multi_key_lookup_joins_enabled`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632230)
			b, err := paramparse.ParseBoolVar("parallelize_multi_key_lookup_joins_enabled", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632232)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632233)
			}
			__antithesis_instrumentation__.Notify(632231)
			m.SetParallelizeMultiKeyLookupJoinsEnabled(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632234)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().ParallelizeMultiKeyLookupJoinsEnabled), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632235)
			return rowexec.ParallelizeMultiKeyLookupJoinsEnabled.String(sv)
		},
	},

	`cost_scans_with_default_col_size`: {
		GetStringVal: makePostgresBoolGetStringValFn(`cost_scans_with_default_col_size`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632236)
			b, err := paramparse.ParseBoolVar(`cost_scans_with_default_col_size`, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632238)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632239)
			}
			__antithesis_instrumentation__.Notify(632237)
			m.SetCostScansWithDefaultColSize(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632240)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().CostScansWithDefaultColSize), nil
		},
		GlobalDefault: globalFalse,
	},
	`default_transaction_quality_of_service`: {
		GetStringVal: makePostgresBoolGetStringValFn(`default_transaction_quality_of_service`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632241)
			qosLevel, ok := sessiondatapb.ParseQoSLevelFromString(s)
			if !ok {
				__antithesis_instrumentation__.Notify(632243)
				return newVarValueError(`default_transaction_quality_of_service`, s,
					sessiondatapb.NormalName, sessiondatapb.UserHighName, sessiondatapb.UserLowName)
			} else {
				__antithesis_instrumentation__.Notify(632244)
			}
			__antithesis_instrumentation__.Notify(632242)
			m.SetQualityOfService(qosLevel)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632245)
			return evalCtx.QualityOfService().String(), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632246)
			return sessiondatapb.Normal.String()
		},
	},
	`opt_split_scan_limit`: {
		GetStringVal: makeIntGetStringValFn(`opt_split_scan_limit`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632247)
			b, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(632251)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632252)
			}
			__antithesis_instrumentation__.Notify(632248)
			if b < 0 {
				__antithesis_instrumentation__.Notify(632253)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"cannot set opt_split_scan_limit to a negative value: %d", b)
			} else {
				__antithesis_instrumentation__.Notify(632254)
			}
			__antithesis_instrumentation__.Notify(632249)
			if b > math.MaxInt32 {
				__antithesis_instrumentation__.Notify(632255)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"cannot set opt_split_scan_limit to a value greater than %d", math.MaxInt32)
			} else {
				__antithesis_instrumentation__.Notify(632256)
			}
			__antithesis_instrumentation__.Notify(632250)

			m.SetOptSplitScanLimit(int32(b))
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632257)
			return strconv.FormatInt(int64(evalCtx.SessionData().OptSplitScanLimit), 10), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632258)
			return strconv.FormatInt(int64(tabledesc.MaxBucketAllowed), 10)
		},
	},

	`enable_super_regions`: {
		GetStringVal: makePostgresBoolGetStringValFn(`enable_super_regions`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632259)
			b, err := paramparse.ParseBoolVar("enable_super_regions", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632261)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632262)
			}
			__antithesis_instrumentation__.Notify(632260)
			m.SetEnableSuperRegions(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632263)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().EnableSuperRegions), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632264)
			return formatBoolAsPostgresSetting(enableSuperRegions.Get(sv))
		},
	},

	`alter_primary_region_super_region_override`: {
		GetStringVal: makePostgresBoolGetStringValFn(`alter_primary_region_super_region_override`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632265)
			b, err := paramparse.ParseBoolVar("alter_primary_region_super_region_override", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632267)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632268)
			}
			__antithesis_instrumentation__.Notify(632266)
			m.SetEnableOverrideAlterPrimaryRegionInSuperRegion(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632269)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().OverrideAlterPrimaryRegionInSuperRegion), nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632270)
			return formatBoolAsPostgresSetting(overrideAlterPrimaryRegionInSuperRegion.Get(sv))
		}},

	`enable_implicit_transaction_for_batch_statements`: {
		GetStringVal: makePostgresBoolGetStringValFn(`enable_implicit_transaction_for_batch_statements`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632271)
			b, err := paramparse.ParseBoolVar("enable_implicit_transaction_for_batch_statements", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632273)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632274)
			}
			__antithesis_instrumentation__.Notify(632272)
			m.SetEnableImplicitTransactionForBatchStatements(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632275)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().EnableImplicitTransactionForBatchStatements), nil
		},
		GlobalDefault: globalFalse,
	},

	`expect_and_ignore_not_visible_columns_in_copy`: {
		GetStringVal: makePostgresBoolGetStringValFn(`expect_and_ignore_not_visible_columns_in_copy`),
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632276)
			b, err := paramparse.ParseBoolVar("expect_and_ignore_not_visible_columns_in_copy", s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632278)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632279)
			}
			__antithesis_instrumentation__.Notify(632277)
			m.SetExpectAndIgnoreNotVisibleColumnsInCopy(b)
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632280)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().ExpectAndIgnoreNotVisibleColumnsInCopy), nil
		},
		GlobalDefault: globalFalse,
	},
}

const compatErrMsg = "this parameter is currently recognized only for compatibility and has no effect in CockroachDB."

func init() {

	for _, p := range []struct {
		name string
		fn   func(ctx context.Context, p *planner, local bool, s string) error
	}{
		{
			name: `role`,
			fn: func(ctx context.Context, p *planner, local bool, s string) error {
				u, err := security.MakeSQLUsernameFromUserInput(s, security.UsernameValidation)
				if err != nil {
					return err
				}
				return p.setRole(ctx, local, u)
			},
		},
	} {
		v := varGen[p.name]
		v.SetWithPlanner = p.fn
		varGen[p.name] = v
	}
	for k, v := range DummyVars {
		varGen[k] = v
	}

	varGen[`idle_session_timeout`] = varGen[`idle_in_session_timeout`]

	for v := range varGen {
		delegate.ValidVars[v] = struct{}{}
	}

	varNames = func() []string {
		res := make([]string, 0, len(varGen))
		for vName := range varGen {
			res = append(res, vName)
			if strings.Contains(vName, ".") {
				panic(fmt.Sprintf(`no session variables with "." can be created as they are reserved for custom options, found %s`, vName))
			}
		}
		sort.Strings(res)
		return res
	}()
}

func makePostgresBoolGetStringValFn(varName string) getStringValFn {
	__antithesis_instrumentation__.Notify(632281)
	return func(
		ctx context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr,
	) (string, error) {
		__antithesis_instrumentation__.Notify(632282)
		if len(values) != 1 {
			__antithesis_instrumentation__.Notify(632287)
			return "", newSingleArgVarError(varName)
		} else {
			__antithesis_instrumentation__.Notify(632288)
		}
		__antithesis_instrumentation__.Notify(632283)
		val, err := values[0].Eval(&evalCtx.EvalContext)
		if err != nil {
			__antithesis_instrumentation__.Notify(632289)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(632290)
		}
		__antithesis_instrumentation__.Notify(632284)
		if s, ok := val.(*tree.DString); ok {
			__antithesis_instrumentation__.Notify(632291)
			return string(*s), nil
		} else {
			__antithesis_instrumentation__.Notify(632292)
		}
		__antithesis_instrumentation__.Notify(632285)
		s, err := paramparse.GetSingleBool(varName, val)
		if err != nil {
			__antithesis_instrumentation__.Notify(632293)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(632294)
		}
		__antithesis_instrumentation__.Notify(632286)
		return strconv.FormatBool(bool(*s)), nil
	}
}

func makeReadOnlyVar(value string) sessionVar {
	__antithesis_instrumentation__.Notify(632295)
	return sessionVar{
		Get: func(_ *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632296)
			return value, nil
		},
		GlobalDefault: func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(632297); return value },
	}
}

func makeReadOnlyVarWithFn(fn func() string) sessionVar {
	__antithesis_instrumentation__.Notify(632298)
	return sessionVar{
		Get: func(_ *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632299)
			return fn(), nil
		},
		GlobalDefault: func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(632300); return fn() },
	}
}

func displayPgBool(val bool) func(_ *settings.Values) string {
	__antithesis_instrumentation__.Notify(632301)
	strVal := formatBoolAsPostgresSetting(val)
	return func(_ *settings.Values) string { __antithesis_instrumentation__.Notify(632302); return strVal }
}

var globalFalse = displayPgBool(false)
var globalTrue = displayPgBool(true)

func sessionDataTimeZoneFormat(loc *time.Location) string {
	__antithesis_instrumentation__.Notify(632303)
	locStr := loc.String()
	_, origRepr, parsed := timeutil.ParseTimeZoneOffset(locStr, timeutil.TimeZoneStringToLocationISO8601Standard)
	if parsed {
		__antithesis_instrumentation__.Notify(632305)
		return origRepr
	} else {
		__antithesis_instrumentation__.Notify(632306)
	}
	__antithesis_instrumentation__.Notify(632304)
	return locStr
}

func makeCompatBoolVar(varName string, displayValue, anyValAllowed bool) sessionVar {
	__antithesis_instrumentation__.Notify(632307)
	displayValStr := formatBoolAsPostgresSetting(displayValue)
	return sessionVar{
		Get: func(_ *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632308)
			return displayValStr, nil
		},
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632309)
			b, err := paramparse.ParseBoolVar(varName, s)
			if err != nil {
				__antithesis_instrumentation__.Notify(632313)
				return err
			} else {
				__antithesis_instrumentation__.Notify(632314)
			}
			__antithesis_instrumentation__.Notify(632310)
			if anyValAllowed || func() bool {
				__antithesis_instrumentation__.Notify(632315)
				return b == displayValue == true
			}() == true {
				__antithesis_instrumentation__.Notify(632316)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(632317)
			}
			__antithesis_instrumentation__.Notify(632311)
			telemetry.Inc(sqltelemetry.UnimplementedSessionVarValueCounter(varName, s))
			allowedVals := []string{displayValStr}
			if anyValAllowed {
				__antithesis_instrumentation__.Notify(632318)
				allowedVals = append(allowedVals, formatBoolAsPostgresSetting(!displayValue))
			} else {
				__antithesis_instrumentation__.Notify(632319)
			}
			__antithesis_instrumentation__.Notify(632312)
			err = newVarValueError(varName, s, allowedVals...)
			err = errors.WithDetail(err, compatErrMsg)
			return err
		},
		GlobalDefault: func(sv *settings.Values) string { __antithesis_instrumentation__.Notify(632320); return displayValStr },
		GetStringVal:  makePostgresBoolGetStringValFn(varName),
	}
}

func makeCompatIntVar(varName string, displayValue int, extraAllowed ...int) sessionVar {
	__antithesis_instrumentation__.Notify(632321)
	displayValueStr := strconv.Itoa(displayValue)
	extraAllowedStr := make([]string, len(extraAllowed))
	for i, v := range extraAllowed {
		__antithesis_instrumentation__.Notify(632323)
		extraAllowedStr[i] = strconv.Itoa(v)
	}
	__antithesis_instrumentation__.Notify(632322)
	varObj := makeCompatStringVar(varName, displayValueStr, extraAllowedStr...)
	varObj.GetStringVal = makeIntGetStringValFn(varName)
	return varObj
}

var _ = makeCompatIntVar

func makeCompatStringVar(varName, displayValue string, extraAllowed ...string) sessionVar {
	__antithesis_instrumentation__.Notify(632324)
	allowedVals := append(extraAllowed, strings.ToLower(displayValue))
	return sessionVar{
		Get: func(_ *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632325)
			return displayValue, nil
		},
		Set: func(_ context.Context, m sessionDataMutator, s string) error {
			__antithesis_instrumentation__.Notify(632326)
			enc := strings.ToLower(s)
			for _, a := range allowedVals {
				__antithesis_instrumentation__.Notify(632328)
				if enc == a {
					__antithesis_instrumentation__.Notify(632329)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(632330)
				}
			}
			__antithesis_instrumentation__.Notify(632327)
			telemetry.Inc(sqltelemetry.UnimplementedSessionVarValueCounter(varName, s))
			err := newVarValueError(varName, s, allowedVals...)
			err = errors.WithDetail(err, compatErrMsg)
			return err
		},
		GlobalDefault: func(sv *settings.Values) string { __antithesis_instrumentation__.Notify(632331); return displayValue },
	}
}

func makeBackwardsCompatBoolVar(varName string, displayValue bool) sessionVar {
	__antithesis_instrumentation__.Notify(632332)
	displayValStr := formatBoolAsPostgresSetting(displayValue)
	return sessionVar{
		Hidden:       true,
		GetStringVal: makePostgresBoolGetStringValFn(varName),
		SetWithPlanner: func(ctx context.Context, p *planner, local bool, s string) error {
			__antithesis_instrumentation__.Notify(632333)
			p.BufferClientNotice(ctx, pgnotice.Newf("%s no longer has any effect", varName))
			return nil
		},
		Get: func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(632334)
			return displayValStr, nil
		},
		GlobalDefault: func(sv *settings.Values) string {
			__antithesis_instrumentation__.Notify(632335)
			return displayValStr
		},
	}
}

func makeIntGetStringValFn(name string) getStringValFn {
	__antithesis_instrumentation__.Notify(632336)
	return func(ctx context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr) (string, error) {
		__antithesis_instrumentation__.Notify(632337)
		s, err := getIntVal(&evalCtx.EvalContext, name, values)
		if err != nil {
			__antithesis_instrumentation__.Notify(632339)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(632340)
		}
		__antithesis_instrumentation__.Notify(632338)
		return strconv.FormatInt(s, 10), nil
	}
}

func makeFloatGetStringValFn(name string) getStringValFn {
	__antithesis_instrumentation__.Notify(632341)
	return func(ctx context.Context, evalCtx *extendedEvalContext, values []tree.TypedExpr) (string, error) {
		__antithesis_instrumentation__.Notify(632342)
		f, err := getFloatVal(&evalCtx.EvalContext, name, values)
		if err != nil {
			__antithesis_instrumentation__.Notify(632344)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(632345)
		}
		__antithesis_instrumentation__.Notify(632343)
		return formatFloatAsPostgresSetting(f), nil
	}
}

func IsSessionVariableConfigurable(varName string) (exists, configurable bool) {
	__antithesis_instrumentation__.Notify(632346)
	v, exists := varGen[varName]
	return exists, v.Set != nil
}

func IsCustomOptionSessionVariable(varName string) bool {
	__antithesis_instrumentation__.Notify(632347)
	_, isCustom := getCustomOptionSessionVar(varName)
	return isCustom
}

func CheckSessionVariableValueValid(
	ctx context.Context, settings *cluster.Settings, varName, varValue string,
) error {
	__antithesis_instrumentation__.Notify(632348)
	_, sVar, err := getSessionVar(varName, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(632351)
		return err
	} else {
		__antithesis_instrumentation__.Notify(632352)
	}
	__antithesis_instrumentation__.Notify(632349)
	if sVar.Set == nil {
		__antithesis_instrumentation__.Notify(632353)
		return pgerror.Newf(pgcode.CantChangeRuntimeParam,
			"parameter %q cannot be changed", varName)
	} else {
		__antithesis_instrumentation__.Notify(632354)
	}
	__antithesis_instrumentation__.Notify(632350)
	fakeSessionMutator := sessionDataMutator{
		data: &sessiondata.SessionData{},
		sessionDataMutatorBase: sessionDataMutatorBase{
			defaults: SessionDefaults(map[string]string{}),
			settings: settings,
		},
		sessionDataMutatorCallbacks: sessionDataMutatorCallbacks{},
	}
	return sVar.Set(ctx, fakeSessionMutator, varValue)
}

var varNames []string

func getSessionVar(name string, missingOk bool) (bool, sessionVar, error) {
	__antithesis_instrumentation__.Notify(632355)
	if _, ok := UnsupportedVars[name]; ok {
		__antithesis_instrumentation__.Notify(632358)
		return false, sessionVar{}, unimplemented.Newf("set."+name,
			"the configuration setting %q is not supported", name)
	} else {
		__antithesis_instrumentation__.Notify(632359)
	}
	__antithesis_instrumentation__.Notify(632356)

	v, ok := varGen[name]
	if !ok {
		__antithesis_instrumentation__.Notify(632360)
		if vCustom, isCustom := getCustomOptionSessionVar(name); isCustom {
			__antithesis_instrumentation__.Notify(632363)
			return true, vCustom, nil
		} else {
			__antithesis_instrumentation__.Notify(632364)
		}
		__antithesis_instrumentation__.Notify(632361)
		if missingOk {
			__antithesis_instrumentation__.Notify(632365)
			return false, sessionVar{}, nil
		} else {
			__antithesis_instrumentation__.Notify(632366)
		}
		__antithesis_instrumentation__.Notify(632362)
		return false, sessionVar{}, pgerror.Newf(pgcode.UndefinedObject,
			"unrecognized configuration parameter %q", name)
	} else {
		__antithesis_instrumentation__.Notify(632367)
	}
	__antithesis_instrumentation__.Notify(632357)
	return true, v, nil
}

func getCustomOptionSessionVar(varName string) (sv sessionVar, isCustom bool) {
	__antithesis_instrumentation__.Notify(632368)
	if strings.Contains(varName, ".") {
		__antithesis_instrumentation__.Notify(632370)
		return sessionVar{
			Get: func(evalCtx *extendedEvalContext) (string, error) {
				__antithesis_instrumentation__.Notify(632371)
				v, ok := evalCtx.SessionData().CustomOptions[varName]
				if !ok {
					__antithesis_instrumentation__.Notify(632373)
					return "", pgerror.Newf(pgcode.UndefinedObject,
						"unrecognized configuration parameter %q", varName)
				} else {
					__antithesis_instrumentation__.Notify(632374)
				}
				__antithesis_instrumentation__.Notify(632372)
				return v, nil
			},
			Set: func(ctx context.Context, m sessionDataMutator, val string) error {
				__antithesis_instrumentation__.Notify(632375)

				m.SetCustomOption(varName, val)
				return nil
			},
			GlobalDefault: func(sv *settings.Values) string {
				__antithesis_instrumentation__.Notify(632376)
				return ""
			},
		}, true
	} else {
		__antithesis_instrumentation__.Notify(632377)
	}
	__antithesis_instrumentation__.Notify(632369)
	return sessionVar{}, false
}

func (p *planner) GetSessionVar(
	_ context.Context, varName string, missingOk bool,
) (bool, string, error) {
	__antithesis_instrumentation__.Notify(632378)
	name := strings.ToLower(varName)
	ok, v, err := getSessionVar(name, missingOk)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(632380)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(632381)
		return ok, "", err
	} else {
		__antithesis_instrumentation__.Notify(632382)
	}
	__antithesis_instrumentation__.Notify(632379)
	val, err := v.Get(&p.extendedEvalCtx)
	return true, val, err
}

func (p *planner) SetSessionVar(ctx context.Context, varName, newVal string, isLocal bool) error {
	__antithesis_instrumentation__.Notify(632383)
	name := strings.ToLower(varName)
	_, v, err := getSessionVar(name, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(632388)
		return err
	} else {
		__antithesis_instrumentation__.Notify(632389)
	}
	__antithesis_instrumentation__.Notify(632384)

	if v.Set == nil && func() bool {
		__antithesis_instrumentation__.Notify(632390)
		return v.RuntimeSet == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(632391)
		return v.SetWithPlanner == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(632392)
		return newCannotChangeParameterError(name)
	} else {
		__antithesis_instrumentation__.Notify(632393)
	}
	__antithesis_instrumentation__.Notify(632385)

	if v.RuntimeSet != nil {
		__antithesis_instrumentation__.Notify(632394)
		return v.RuntimeSet(ctx, p.ExtendedEvalContext(), isLocal, newVal)
	} else {
		__antithesis_instrumentation__.Notify(632395)
	}
	__antithesis_instrumentation__.Notify(632386)
	if v.SetWithPlanner != nil {
		__antithesis_instrumentation__.Notify(632396)
		return v.SetWithPlanner(ctx, p, isLocal, newVal)
	} else {
		__antithesis_instrumentation__.Notify(632397)
	}
	__antithesis_instrumentation__.Notify(632387)
	return p.applyOnSessionDataMutators(
		ctx,
		isLocal,
		func(m sessionDataMutator) error {
			__antithesis_instrumentation__.Notify(632398)
			return v.Set(ctx, m, newVal)
		},
	)
}
