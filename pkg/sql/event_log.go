package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type eventLogEntry struct {
	targetID int32

	event eventpb.EventPayload
}

func (p *planner) logEvent(
	ctx context.Context, descID descpb.ID, event eventpb.EventPayload,
) error {
	__antithesis_instrumentation__.Notify(470039)
	return p.logEventsWithOptions(ctx,
		2,
		eventLogOptions{dst: LogEverywhere},
		eventLogEntry{targetID: int32(descID), event: event})
}

func (p *planner) logEvents(ctx context.Context, entries ...eventLogEntry) error {
	__antithesis_instrumentation__.Notify(470040)
	return p.logEventsWithOptions(ctx,
		2,
		eventLogOptions{dst: LogEverywhere},
		entries...)
}

type eventLogOptions struct {
	dst LogEventDestination

	verboseTraceLevel log.Level

	rOpts redactionOptions
}

type redactionOptions struct {
	omitSQLNameRedaction bool
}

func (ro *redactionOptions) toFlags() tree.FmtFlags {
	__antithesis_instrumentation__.Notify(470041)
	if ro.omitSQLNameRedaction {
		__antithesis_instrumentation__.Notify(470043)
		return tree.FmtOmitNameRedaction
	} else {
		__antithesis_instrumentation__.Notify(470044)
	}
	__antithesis_instrumentation__.Notify(470042)
	return tree.FmtSimple
}

var defaultRedactionOptions = redactionOptions{
	omitSQLNameRedaction: false,
}

func (p *planner) getCommonSQLEventDetails(opt redactionOptions) eventpb.CommonSQLEventDetails {
	__antithesis_instrumentation__.Notify(470045)
	redactableStmt := formatStmtKeyAsRedactableString(
		p.extendedEvalCtx.VirtualSchemas, p.stmt.AST,
		p.extendedEvalCtx.EvalContext.Annotations, opt.toFlags(),
	)
	commonSQLEventDetails := eventpb.CommonSQLEventDetails{
		Statement:       redactableStmt,
		Tag:             p.stmt.AST.StatementTag(),
		User:            p.User().Normalized(),
		ApplicationName: p.SessionData().ApplicationName,
	}
	if pls := p.extendedEvalCtx.EvalContext.Placeholders.Values; len(pls) > 0 {
		__antithesis_instrumentation__.Notify(470047)
		commonSQLEventDetails.PlaceholderValues = make([]string, len(pls))
		for idx, val := range pls {
			__antithesis_instrumentation__.Notify(470048)
			commonSQLEventDetails.PlaceholderValues[idx] = val.String()
		}
	} else {
		__antithesis_instrumentation__.Notify(470049)
	}
	__antithesis_instrumentation__.Notify(470046)
	return commonSQLEventDetails
}

func (p *planner) logEventsWithOptions(
	ctx context.Context, depth int, opts eventLogOptions, entries ...eventLogEntry,
) error {
	__antithesis_instrumentation__.Notify(470050)
	return logEventInternalForSQLStatements(ctx,
		p.extendedEvalCtx.ExecCfg, p.txn,
		1+depth,
		opts,
		p.getCommonSQLEventDetails(opts.rOpts),
		entries...)
}

func logEventInternalForSchemaChanges(
	ctx context.Context,
	execCfg *ExecutorConfig,
	txn *kv.Txn,
	sqlInstanceID base.SQLInstanceID,
	descID descpb.ID,
	mutationID descpb.MutationID,
	event eventpb.EventPayload,
) error {
	__antithesis_instrumentation__.Notify(470051)
	event.CommonDetails().Timestamp = txn.ReadTimestamp().WallTime
	scCommon, ok := event.(eventpb.EventWithCommonSchemaChangePayload)
	if !ok {
		__antithesis_instrumentation__.Notify(470053)
		return errors.AssertionFailedf("unknown event type: %T", event)
	} else {
		__antithesis_instrumentation__.Notify(470054)
	}
	__antithesis_instrumentation__.Notify(470052)
	m := scCommon.CommonSchemaChangeDetails()
	m.InstanceID = int32(sqlInstanceID)
	m.DescriptorID = uint32(descID)
	m.MutationID = uint32(mutationID)

	return insertEventRecords(
		ctx, execCfg.InternalExecutor,
		txn,
		int32(execCfg.NodeID.SQLInstanceID()),
		1,
		eventLogOptions{dst: LogEverywhere},
		eventLogEntry{
			targetID: int32(descID),
			event:    event,
		},
	)
}

func logEventInternalForSQLStatements(
	ctx context.Context,
	execCfg *ExecutorConfig,
	txn *kv.Txn,
	depth int,
	opts eventLogOptions,
	commonSQLEventDetails eventpb.CommonSQLEventDetails,
	entries ...eventLogEntry,
) error {
	__antithesis_instrumentation__.Notify(470055)

	injectCommonFields := func(entry eventLogEntry) error {
		__antithesis_instrumentation__.Notify(470058)
		event := entry.event
		event.CommonDetails().Timestamp = txn.ReadTimestamp().WallTime
		sqlCommon, ok := event.(eventpb.EventWithCommonSQLPayload)
		if !ok {
			__antithesis_instrumentation__.Notify(470060)
			return errors.AssertionFailedf("unknown event type: %T", event)
		} else {
			__antithesis_instrumentation__.Notify(470061)
		}
		__antithesis_instrumentation__.Notify(470059)
		m := sqlCommon.CommonSQLDetails()
		*m = commonSQLEventDetails
		m.DescriptorID = uint32(entry.targetID)
		return nil
	}
	__antithesis_instrumentation__.Notify(470056)

	for i := range entries {
		__antithesis_instrumentation__.Notify(470062)
		if err := injectCommonFields(entries[i]); err != nil {
			__antithesis_instrumentation__.Notify(470063)
			return err
		} else {
			__antithesis_instrumentation__.Notify(470064)
		}
	}
	__antithesis_instrumentation__.Notify(470057)

	return insertEventRecords(
		ctx,
		execCfg.InternalExecutor,
		txn,
		int32(execCfg.NodeID.SQLInstanceID()),
		1+depth,
		opts,
		entries...,
	)
}

type schemaChangerEventLogger struct {
	txn     *kv.Txn
	execCfg *ExecutorConfig
	depth   int
}

var _ scexec.EventLogger = (*schemaChangerEventLogger)(nil)

func NewSchemaChangerEventLogger(
	txn *kv.Txn, execCfg *ExecutorConfig, depth int,
) scexec.EventLogger {
	__antithesis_instrumentation__.Notify(470065)
	return &schemaChangerEventLogger{
		txn:     txn,
		execCfg: execCfg,
		depth:   depth,
	}
}

func (l schemaChangerEventLogger) LogEvent(
	ctx context.Context,
	descID descpb.ID,
	details eventpb.CommonSQLEventDetails,
	event eventpb.EventPayload,
) error {
	__antithesis_instrumentation__.Notify(470066)
	entry := eventLogEntry{targetID: int32(descID), event: event}
	return logEventInternalForSQLStatements(ctx,
		l.execCfg,
		l.txn,
		l.depth,
		eventLogOptions{dst: LogEverywhere},
		details,
		entry)
}

func LogEventForJobs(
	ctx context.Context,
	execCfg *ExecutorConfig,
	txn *kv.Txn,
	event eventpb.EventPayload,
	jobID int64,
	payload jobspb.Payload,
	user security.SQLUsername,
	status jobs.Status,
) error {
	__antithesis_instrumentation__.Notify(470067)
	event.CommonDetails().Timestamp = txn.ReadTimestamp().WallTime
	jobCommon, ok := event.(eventpb.EventWithCommonJobPayload)
	if !ok {
		__antithesis_instrumentation__.Notify(470070)
		return errors.AssertionFailedf("unknown event type: %T", event)
	} else {
		__antithesis_instrumentation__.Notify(470071)
	}
	__antithesis_instrumentation__.Notify(470068)
	m := jobCommon.CommonJobDetails()
	m.JobID = jobID
	m.JobType = payload.Type().String()
	m.User = user.Normalized()
	m.Status = string(status)
	for _, id := range payload.DescriptorIDs {
		__antithesis_instrumentation__.Notify(470072)
		m.DescriptorIDs = append(m.DescriptorIDs, uint32(id))
	}
	__antithesis_instrumentation__.Notify(470069)
	m.Description = payload.Description

	return insertEventRecords(
		ctx, execCfg.InternalExecutor,
		txn,
		int32(execCfg.NodeID.SQLInstanceID()),
		1,
		eventLogOptions{dst: LogEverywhere},
		eventLogEntry{event: event},
	)
}

var eventLogSystemTableEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"server.eventlog.enabled",
	"if set, logged notable events are also stored in the table system.eventlog",
	true,
).WithPublic()

type LogEventDestination int

func (d LogEventDestination) hasFlag(f LogEventDestination) bool {
	__antithesis_instrumentation__.Notify(470073)
	return d&f != 0
}

const (
	LogToSystemTable LogEventDestination = 1 << iota

	LogExternally

	LogToDevChannelIfVerbose

	LogEverywhere LogEventDestination = LogExternally | LogToSystemTable | LogToDevChannelIfVerbose
)

func InsertEventRecord(
	ctx context.Context,
	ex *InternalExecutor,
	txn *kv.Txn,
	reportingID int32,
	dst LogEventDestination,
	targetID int32,
	info eventpb.EventPayload,
) error {
	__antithesis_instrumentation__.Notify(470074)

	return insertEventRecords(ctx, ex, txn, reportingID,
		1,
		eventLogOptions{dst: dst},
		eventLogEntry{targetID: targetID, event: info})
}

func insertEventRecords(
	ctx context.Context,
	ex *InternalExecutor,
	txn *kv.Txn,
	reportingID int32,
	depth int,
	opts eventLogOptions,
	entries ...eventLogEntry,
) error {
	__antithesis_instrumentation__.Notify(470075)

	for i := range entries {
		__antithesis_instrumentation__.Notify(470085)

		event := entries[i].event
		eventType := eventpb.GetEventTypeName(event)
		event.CommonDetails().EventType = eventType

		if event.CommonDetails().Timestamp == 0 {
			__antithesis_instrumentation__.Notify(470086)
			return errors.AssertionFailedf("programming error: timestamp field in event %d not populated: %T", i, event)
		} else {
			__antithesis_instrumentation__.Notify(470087)
		}
	}
	__antithesis_instrumentation__.Notify(470076)

	if opts.dst.hasFlag(LogToDevChannelIfVerbose) {
		__antithesis_instrumentation__.Notify(470088)

		level := log.Level(2)
		if opts.verboseTraceLevel != 0 {
			__antithesis_instrumentation__.Notify(470090)

			level = opts.verboseTraceLevel
		} else {
			__antithesis_instrumentation__.Notify(470091)
		}
		__antithesis_instrumentation__.Notify(470089)
		if log.VDepth(level, depth) {
			__antithesis_instrumentation__.Notify(470092)

			for i := range entries {
				__antithesis_instrumentation__.Notify(470093)
				log.InfofDepth(ctx, depth, "SQL event: target %d, payload %+v", entries[i].targetID, entries[i].event)
			}
		} else {
			__antithesis_instrumentation__.Notify(470094)
		}
	} else {
		__antithesis_instrumentation__.Notify(470095)
	}
	__antithesis_instrumentation__.Notify(470077)

	loggingToSystemTable := opts.dst.hasFlag(LogToSystemTable) && func() bool {
		__antithesis_instrumentation__.Notify(470096)
		return eventLogSystemTableEnabled.Get(&ex.s.cfg.Settings.SV) == true
	}() == true
	if !loggingToSystemTable {
		__antithesis_instrumentation__.Notify(470097)

		if opts.dst.hasFlag(LogExternally) {
			__antithesis_instrumentation__.Notify(470099)
			for i := range entries {
				__antithesis_instrumentation__.Notify(470100)
				log.StructuredEvent(ctx, entries[i].event)
			}
		} else {
			__antithesis_instrumentation__.Notify(470101)
		}
		__antithesis_instrumentation__.Notify(470098)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(470102)
	}
	__antithesis_instrumentation__.Notify(470078)

	if opts.dst.hasFlag(LogExternally) {
		__antithesis_instrumentation__.Notify(470103)
		txn.AddCommitTrigger(func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(470104)
			for i := range entries {
				__antithesis_instrumentation__.Notify(470105)
				log.StructuredEvent(ctx, entries[i].event)
			}
		})
	} else {
		__antithesis_instrumentation__.Notify(470106)
	}
	__antithesis_instrumentation__.Notify(470079)

	const colsPerEvent = 5
	const baseQuery = `
INSERT INTO system.eventlog (
  timestamp, "eventType", "targetID", "reportingID", info
)
VALUES($1, $2, $3, $4, $5)`
	args := make([]interface{}, 0, len(entries)*colsPerEvent)
	constructArgs := func(reportingID int32, entry eventLogEntry) error {
		__antithesis_instrumentation__.Notify(470107)
		event := entry.event
		infoBytes := redact.RedactableBytes("{")
		_, infoBytes = event.AppendJSONFields(false, infoBytes)
		infoBytes = append(infoBytes, '}')

		infoBytes = infoBytes.StripMarkers()
		eventType := eventpb.GetEventTypeName(event)
		args = append(
			args,
			timeutil.Unix(0, event.CommonDetails().Timestamp),
			eventType,
			entry.targetID,
			reportingID,
			string(infoBytes),
		)
		return nil
	}
	__antithesis_instrumentation__.Notify(470080)

	query := baseQuery
	if err := constructArgs(reportingID, entries[0]); err != nil {
		__antithesis_instrumentation__.Notify(470108)
		return err
	} else {
		__antithesis_instrumentation__.Notify(470109)
	}
	__antithesis_instrumentation__.Notify(470081)
	if len(entries) > 1 {
		__antithesis_instrumentation__.Notify(470110)

		var completeQuery strings.Builder
		completeQuery.WriteString(baseQuery)

		for _, extraEntry := range entries[1:] {
			__antithesis_instrumentation__.Notify(470112)
			placeholderNum := 1 + len(args)
			if err := constructArgs(reportingID, extraEntry); err != nil {
				__antithesis_instrumentation__.Notify(470114)
				return err
			} else {
				__antithesis_instrumentation__.Notify(470115)
			}
			__antithesis_instrumentation__.Notify(470113)
			fmt.Fprintf(&completeQuery, ", ($%d, $%d, $%d, $%d, $%d)",
				placeholderNum, placeholderNum+1, placeholderNum+2, placeholderNum+3, placeholderNum+4)
		}
		__antithesis_instrumentation__.Notify(470111)
		query = completeQuery.String()
	} else {
		__antithesis_instrumentation__.Notify(470116)
	}
	__antithesis_instrumentation__.Notify(470082)

	rows, err := ex.Exec(ctx, "log-event", txn, query, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(470117)
		return err
	} else {
		__antithesis_instrumentation__.Notify(470118)
	}
	__antithesis_instrumentation__.Notify(470083)
	if rows != len(entries) {
		__antithesis_instrumentation__.Notify(470119)
		return errors.Errorf("%d rows affected by log insertion; expected %d rows affected.", rows, len(entries))
	} else {
		__antithesis_instrumentation__.Notify(470120)
	}
	__antithesis_instrumentation__.Notify(470084)
	return nil
}
