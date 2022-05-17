package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var logStatementsExecuteEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.trace.log_statement_execute",
	"set to true to enable logging of executed statements",
	false,
).WithPublic()

var slowQueryLogThreshold = settings.RegisterPublicDurationSettingWithExplicitUnit(
	settings.TenantWritable,
	"sql.log.slow_query.latency_threshold",
	"when set to non-zero, log statements whose service latency exceeds "+
		"the threshold to a secondary logger on each node",
	0,
	settings.NonNegativeDuration,
)

var slowInternalQueryLogEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.log.slow_query.internal_queries.enabled",
	"when set to true, internal queries which exceed the slow query log threshold "+
		"are logged to a separate log. Must have the slow query log enabled for this "+
		"setting to have any effect.",
	false,
).WithPublic()

var slowQueryLogFullTableScans = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.log.slow_query.experimental_full_table_scans.enabled",
	"when set to true, statements that perform a full table/index scan will be logged to the "+
		"slow query log even if they do not meet the latency threshold. Must have the slow query "+
		"log enabled for this setting to have any effect.",
	false,
).WithPublic()

var unstructuredQueryLog = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.log.unstructured_entries.enabled",
	"when set, SQL execution and audit logs use the pre-v21.1 unstrucured format",
	false,
)

var adminAuditLogEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.log.admin_audit.enabled",
	"when set, log SQL queries that are executed by a user with admin privileges",
	false,
)

var telemetryLoggingEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.telemetry.query_sampling.enabled",
	"when set to true, executed queries will emit an event on the telemetry logging channel",

	envutil.EnvOrDefaultBool("COCKROACH_SQL_TELEMETRY_QUERY_SAMPLING_ENABLED", false),
).WithPublic()

type executorType int

const (
	executorTypeExec executorType = iota
	executorTypeInternal
)

func (s executorType) vLevel() log.Level {
	__antithesis_instrumentation__.Notify(470223)
	return log.Level(s) + 2
}

var logLabels = []string{"exec", "exec-internal"}

func (s executorType) logLabel() string {
	__antithesis_instrumentation__.Notify(470224)
	return logLabels[s]
}

var sqlPerfLogger log.ChannelLogger = log.SqlPerf
var sqlPerfInternalLogger log.ChannelLogger = log.SqlInternalPerf

func (p *planner) maybeLogStatement(
	ctx context.Context,
	execType executorType,
	numRetries, txnCounter, rows int,
	err error,
	queryReceived time.Time,
	hasAdminRoleCache *HasAdminRoleCache,
	telemetryLoggingMetrics *TelemetryLoggingMetrics,
) {
	__antithesis_instrumentation__.Notify(470225)
	p.maybeLogStatementInternal(ctx, execType, numRetries, txnCounter, rows, err, queryReceived, hasAdminRoleCache, telemetryLoggingMetrics)
}

func (p *planner) maybeLogStatementInternal(
	ctx context.Context,
	execType executorType,
	numRetries, txnCounter, rows int,
	err error,
	startTime time.Time,
	hasAdminRoleCache *HasAdminRoleCache,
	telemetryMetrics *TelemetryLoggingMetrics,
) {
	__antithesis_instrumentation__.Notify(470226)

	logV := log.V(2)
	logExecuteEnabled := logStatementsExecuteEnabled.Get(&p.execCfg.Settings.SV)
	slowLogThreshold := slowQueryLogThreshold.Get(&p.execCfg.Settings.SV)
	slowLogFullTableScans := slowQueryLogFullTableScans.Get(&p.execCfg.Settings.SV)
	slowQueryLogEnabled := slowLogThreshold != 0
	slowInternalQueryLogEnabled := slowInternalQueryLogEnabled.Get(&p.execCfg.Settings.SV)
	auditEventsDetected := len(p.curPlan.auditEvents) != 0
	maxEventFrequency := telemetryMaxEventFrequency.Get(&p.execCfg.Settings.SV)

	telemetryLoggingEnabled := telemetryLoggingEnabled.Get(&p.execCfg.Settings.SV) && func() bool {
		__antithesis_instrumentation__.Notify(470235)
		return execType != executorTypeInternal == true
	}() == true

	shouldLogToAdminAuditLog := hasAdminRoleCache.IsSet && func() bool {
		__antithesis_instrumentation__.Notify(470236)
		return hasAdminRoleCache.HasAdminRole == true
	}() == true

	if !logV && func() bool {
		__antithesis_instrumentation__.Notify(470237)
		return !logExecuteEnabled == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(470238)
		return !auditEventsDetected == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(470239)
		return !slowQueryLogEnabled == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(470240)
		return !shouldLogToAdminAuditLog == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(470241)
		return !telemetryLoggingEnabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(470242)

		return
	} else {
		__antithesis_instrumentation__.Notify(470243)
	}
	__antithesis_instrumentation__.Notify(470227)

	appName := p.EvalContext().SessionData().ApplicationName

	queryDuration := timeutil.Since(startTime)
	age := float32(queryDuration.Nanoseconds()) / 1e6

	execErrStr := ""
	if err != nil {
		__antithesis_instrumentation__.Notify(470244)
		execErrStr = err.Error()
	} else {
		__antithesis_instrumentation__.Notify(470245)
	}
	__antithesis_instrumentation__.Notify(470228)

	lbl := execType.logLabel()

	if unstructuredQueryLog.Get(&p.execCfg.Settings.SV) {
		__antithesis_instrumentation__.Notify(470246)

		stmtStr := p.curPlan.stmt.AST.String()
		plStr := p.extendedEvalCtx.Placeholders.Values.String()

		if logV {
			__antithesis_instrumentation__.Notify(470251)

			log.VEventf(ctx, execType.vLevel(), "%s %q %q %s %.3f %d %q %d",
				lbl, appName, stmtStr, plStr, age, rows, execErrStr, numRetries)
		} else {
			__antithesis_instrumentation__.Notify(470252)
		}
		__antithesis_instrumentation__.Notify(470247)

		if auditEventsDetected {
			__antithesis_instrumentation__.Notify(470253)
			auditErrStr := "OK"
			if err != nil {
				__antithesis_instrumentation__.Notify(470256)
				auditErrStr = "ERROR"
			} else {
				__antithesis_instrumentation__.Notify(470257)
			}
			__antithesis_instrumentation__.Notify(470254)

			var buf bytes.Buffer
			buf.WriteByte('{')
			sep := ""
			for _, ev := range p.curPlan.auditEvents {
				__antithesis_instrumentation__.Notify(470258)
				mode := "READ"
				if ev.writing {
					__antithesis_instrumentation__.Notify(470260)
					mode = "READWRITE"
				} else {
					__antithesis_instrumentation__.Notify(470261)
				}
				__antithesis_instrumentation__.Notify(470259)
				fmt.Fprintf(&buf, "%s%q[%d]:%s", sep, ev.desc.GetName(), ev.desc.GetID(), mode)
				sep = ", "
			}
			__antithesis_instrumentation__.Notify(470255)
			buf.WriteByte('}')
			logTrigger := buf.String()

			log.SensitiveAccess.Infof(ctx, "%s %q %s %q %s %.3f %d %s %d",
				lbl, appName, logTrigger, stmtStr, plStr, age, rows, auditErrStr, numRetries)
		} else {
			__antithesis_instrumentation__.Notify(470262)
		}
		__antithesis_instrumentation__.Notify(470248)
		if slowQueryLogEnabled && func() bool {
			__antithesis_instrumentation__.Notify(470263)
			return (queryDuration > slowLogThreshold || func() bool {
				__antithesis_instrumentation__.Notify(470264)
				return slowLogFullTableScans == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(470265)
			logReason, shouldLog := p.slowQueryLogReason(queryDuration, slowLogThreshold)

			var logger log.ChannelLogger

			if execType == executorTypeExec {
				__antithesis_instrumentation__.Notify(470268)
				logger = sqlPerfLogger
			} else {
				__antithesis_instrumentation__.Notify(470269)
			}
			__antithesis_instrumentation__.Notify(470266)

			if execType == executorTypeInternal && func() bool {
				__antithesis_instrumentation__.Notify(470270)
				return slowInternalQueryLogEnabled == true
			}() == true {
				__antithesis_instrumentation__.Notify(470271)
				logger = sqlPerfInternalLogger
			} else {
				__antithesis_instrumentation__.Notify(470272)
			}
			__antithesis_instrumentation__.Notify(470267)

			if logger != nil && func() bool {
				__antithesis_instrumentation__.Notify(470273)
				return shouldLog == true
			}() == true {
				__antithesis_instrumentation__.Notify(470274)
				logger.Infof(ctx, "%.3fms %s %q {} %q %s %d %q %d %s",
					age, lbl, appName, stmtStr, plStr, rows, execErrStr, numRetries, logReason)
			} else {
				__antithesis_instrumentation__.Notify(470275)
			}
		} else {
			__antithesis_instrumentation__.Notify(470276)
		}
		__antithesis_instrumentation__.Notify(470249)
		if logExecuteEnabled {
			__antithesis_instrumentation__.Notify(470277)
			log.SqlExec.Infof(ctx, "%s %q {} %q %s %.3f %d %q %d",
				lbl, appName, stmtStr, plStr, age, rows, execErrStr, numRetries)
		} else {
			__antithesis_instrumentation__.Notify(470278)
		}
		__antithesis_instrumentation__.Notify(470250)
		return
	} else {
		__antithesis_instrumentation__.Notify(470279)
	}
	__antithesis_instrumentation__.Notify(470229)

	sqlErrState := ""
	if err != nil {
		__antithesis_instrumentation__.Notify(470280)
		sqlErrState = pgerror.GetPGCode(err).String()
	} else {
		__antithesis_instrumentation__.Notify(470281)
	}
	__antithesis_instrumentation__.Notify(470230)

	execDetails := eventpb.CommonSQLExecDetails{

		ExecMode:      lbl,
		NumRows:       uint64(rows),
		SQLSTATE:      sqlErrState,
		ErrorText:     execErrStr,
		Age:           age,
		NumRetries:    uint32(numRetries),
		FullTableScan: p.curPlan.flags.IsSet(planFlagContainsFullTableScan),
		FullIndexScan: p.curPlan.flags.IsSet(planFlagContainsFullIndexScan),
		TxnCounter:    uint32(txnCounter),
	}

	if auditEventsDetected {
		__antithesis_instrumentation__.Notify(470282)

		entries := make([]eventLogEntry, len(p.curPlan.auditEvents))
		for i, ev := range p.curPlan.auditEvents {
			__antithesis_instrumentation__.Notify(470284)
			mode := "r"
			if ev.writing {
				__antithesis_instrumentation__.Notify(470287)
				mode = "rw"
			} else {
				__antithesis_instrumentation__.Notify(470288)
			}
			__antithesis_instrumentation__.Notify(470285)
			tableName := ""
			if t, ok := ev.desc.(catalog.TableDescriptor); ok {
				__antithesis_instrumentation__.Notify(470289)

				tn, err := p.getQualifiedTableName(ctx, t)
				if err != nil {
					__antithesis_instrumentation__.Notify(470290)
					log.Warningf(ctx, "name for audited table ID %d not found: %v", ev.desc.GetID(), err)
				} else {
					__antithesis_instrumentation__.Notify(470291)
					tableName = tn.FQString()
				}
			} else {
				__antithesis_instrumentation__.Notify(470292)
			}
			__antithesis_instrumentation__.Notify(470286)
			entries[i] = eventLogEntry{
				targetID: int32(ev.desc.GetID()),
				event: &eventpb.SensitiveTableAccess{
					CommonSQLExecDetails: execDetails,
					TableName:            tableName,
					AccessMode:           mode,
				},
			}
		}
		__antithesis_instrumentation__.Notify(470283)
		p.logEventsOnlyExternally(ctx, entries...)
	} else {
		__antithesis_instrumentation__.Notify(470293)
	}
	__antithesis_instrumentation__.Notify(470231)

	if slowQueryLogEnabled && func() bool {
		__antithesis_instrumentation__.Notify(470294)
		return ((slowLogFullTableScans && func() bool {
			__antithesis_instrumentation__.Notify(470295)
			return (execDetails.FullTableScan || func() bool {
				__antithesis_instrumentation__.Notify(470296)
				return execDetails.FullIndexScan == true
			}() == true) == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(470297)
			return queryDuration > slowLogThreshold == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(470298)
		switch {
		case execType == executorTypeExec:
			__antithesis_instrumentation__.Notify(470299)

			p.logEventsOnlyExternally(ctx, eventLogEntry{event: &eventpb.SlowQuery{CommonSQLExecDetails: execDetails}})

		case execType == executorTypeInternal && func() bool {
			__antithesis_instrumentation__.Notify(470302)
			return slowInternalQueryLogEnabled == true
		}() == true:
			__antithesis_instrumentation__.Notify(470300)

			p.logEventsOnlyExternally(ctx, eventLogEntry{event: &eventpb.SlowQueryInternal{CommonSQLExecDetails: execDetails}})
		default:
			__antithesis_instrumentation__.Notify(470301)
		}
	} else {
		__antithesis_instrumentation__.Notify(470303)
	}
	__antithesis_instrumentation__.Notify(470232)

	if logExecuteEnabled || func() bool {
		__antithesis_instrumentation__.Notify(470304)
		return logV == true
	}() == true {
		__antithesis_instrumentation__.Notify(470305)

		_ = p.logEventsWithOptions(ctx,
			1,
			eventLogOptions{

				dst:               LogExternally | LogToDevChannelIfVerbose,
				verboseTraceLevel: execType.vLevel(),
			},
			eventLogEntry{event: &eventpb.QueryExecute{CommonSQLExecDetails: execDetails}})
	} else {
		__antithesis_instrumentation__.Notify(470306)
	}
	__antithesis_instrumentation__.Notify(470233)

	if shouldLogToAdminAuditLog {
		__antithesis_instrumentation__.Notify(470307)
		p.logEventsOnlyExternally(ctx, eventLogEntry{event: &eventpb.AdminQuery{CommonSQLExecDetails: execDetails}})
	} else {
		__antithesis_instrumentation__.Notify(470308)
	}
	__antithesis_instrumentation__.Notify(470234)

	if telemetryLoggingEnabled {
		__antithesis_instrumentation__.Notify(470309)

		requiredTimeElapsed := 1.0 / float64(maxEventFrequency)
		if p.stmt.AST.StatementType() != tree.TypeDML {
			__antithesis_instrumentation__.Notify(470311)
			requiredTimeElapsed = 0
		} else {
			__antithesis_instrumentation__.Notify(470312)
		}
		__antithesis_instrumentation__.Notify(470310)
		if telemetryMetrics.maybeUpdateLastEmittedTime(telemetryMetrics.timeNow(), requiredTimeElapsed) {
			__antithesis_instrumentation__.Notify(470313)
			skippedQueries := telemetryMetrics.resetSkippedQueryCount()
			p.logOperationalEventsOnlyExternally(ctx, eventLogEntry{event: &eventpb.SampledQuery{
				CommonSQLExecDetails: execDetails,
				SkippedQueries:       skippedQueries,
				CostEstimate:         p.curPlan.instrumentation.costEstimate,
				Distribution:         p.curPlan.instrumentation.distribution.String(),
			}})
		} else {
			__antithesis_instrumentation__.Notify(470314)
			telemetryMetrics.incSkippedQueryCount()
		}
	} else {
		__antithesis_instrumentation__.Notify(470315)
	}
}

func (p *planner) logEventsOnlyExternally(ctx context.Context, entries ...eventLogEntry) {
	__antithesis_instrumentation__.Notify(470316)

	_ = p.logEventsWithOptions(ctx,
		2,
		eventLogOptions{dst: LogExternally},
		entries...)
}

func (p *planner) logOperationalEventsOnlyExternally(
	ctx context.Context, entries ...eventLogEntry,
) {
	__antithesis_instrumentation__.Notify(470317)

	_ = p.logEventsWithOptions(ctx,
		2,
		eventLogOptions{dst: LogExternally, rOpts: redactionOptions{omitSQLNameRedaction: true}},
		entries...)
}

func (p *planner) maybeAudit(desc catalog.Descriptor, priv privilege.Kind) {
	__antithesis_instrumentation__.Notify(470318)
	wantedMode := desc.GetAuditMode()
	if wantedMode == descpb.TableDescriptor_DISABLED {
		__antithesis_instrumentation__.Notify(470320)
		return
	} else {
		__antithesis_instrumentation__.Notify(470321)
	}
	__antithesis_instrumentation__.Notify(470319)

	switch priv {
	case privilege.INSERT, privilege.DELETE, privilege.UPDATE:
		__antithesis_instrumentation__.Notify(470322)
		p.curPlan.auditEvents = append(p.curPlan.auditEvents, auditEvent{desc: desc, writing: true})
	default:
		__antithesis_instrumentation__.Notify(470323)
		p.curPlan.auditEvents = append(p.curPlan.auditEvents, auditEvent{desc: desc, writing: false})
	}
}

func (p *planner) slowQueryLogReason(
	queryDuration time.Duration, slowLogThreshold time.Duration,
) (reason string, shouldLog bool) {
	__antithesis_instrumentation__.Notify(470324)
	var buf bytes.Buffer
	buf.WriteByte('{')
	sep := " "
	if slowLogThreshold != 0 && func() bool {
		__antithesis_instrumentation__.Notify(470328)
		return queryDuration > slowLogThreshold == true
	}() == true {
		__antithesis_instrumentation__.Notify(470329)
		fmt.Fprintf(&buf, "%sLATENCY_THRESHOLD", sep)
	} else {
		__antithesis_instrumentation__.Notify(470330)
	}
	__antithesis_instrumentation__.Notify(470325)
	if p.curPlan.flags.IsSet(planFlagContainsFullTableScan) {
		__antithesis_instrumentation__.Notify(470331)
		fmt.Fprintf(&buf, "%sFULL_TABLE_SCAN", sep)
	} else {
		__antithesis_instrumentation__.Notify(470332)
	}
	__antithesis_instrumentation__.Notify(470326)
	if p.curPlan.flags.IsSet(planFlagContainsFullIndexScan) {
		__antithesis_instrumentation__.Notify(470333)
		fmt.Fprintf(&buf, "%sFULL_SECONDARY_INDEX_SCAN", sep)
	} else {
		__antithesis_instrumentation__.Notify(470334)
	}
	__antithesis_instrumentation__.Notify(470327)
	buf.WriteByte(' ')
	buf.WriteByte('}')
	reason = buf.String()
	return reason, reason != "{ }"
}

type auditEvent struct {
	desc catalog.Descriptor

	writing bool
}
