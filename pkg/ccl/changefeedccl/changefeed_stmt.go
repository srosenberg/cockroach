package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/backupccl/backupresolver"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

var featureChangefeedEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"feature.changefeed.enabled",
	"set to true to enable changefeeds, false to disable; default is true",
	featureflag.FeatureFlagEnabledDefault,
).WithPublic()

func init() {
	sql.AddPlanHook("changefeed", changefeedPlanHook)
	jobs.RegisterConstructor(
		jobspb.TypeChangefeed,
		func(job *jobs.Job, _ *cluster.Settings) jobs.Resumer {
			return &changefeedResumer{job: job}
		},
	)
}

type annotatedChangefeedStatement struct {
	*tree.CreateChangefeed
	originalSpecs map[tree.ChangefeedTarget]jobspb.ChangefeedTargetSpecification
}

func getChangefeedStatement(stmt tree.Statement) *annotatedChangefeedStatement {
	__antithesis_instrumentation__.Notify(16085)
	switch changefeed := stmt.(type) {
	case *annotatedChangefeedStatement:
		__antithesis_instrumentation__.Notify(16086)
		return changefeed
	case *tree.CreateChangefeed:
		__antithesis_instrumentation__.Notify(16087)
		return &annotatedChangefeedStatement{CreateChangefeed: changefeed}
	default:
		__antithesis_instrumentation__.Notify(16088)
		return nil
	}
}

func changefeedPlanHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(16089)
	changefeedStmt := getChangefeedStatement(stmt)
	if changefeedStmt == nil {
		__antithesis_instrumentation__.Notify(16095)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(16096)
	}
	__antithesis_instrumentation__.Notify(16090)

	var sinkURIFn func() (string, error)
	var header colinfo.ResultColumns
	unspecifiedSink := changefeedStmt.SinkURI == nil
	avoidBuffering := false

	if unspecifiedSink {
		__antithesis_instrumentation__.Notify(16097)

		sinkURIFn = func() (string, error) { __antithesis_instrumentation__.Notify(16099); return ``, nil }
		__antithesis_instrumentation__.Notify(16098)
		header = colinfo.ResultColumns{
			{Name: "table", Typ: types.String},
			{Name: "key", Typ: types.Bytes},
			{Name: "value", Typ: types.Bytes},
		}
		avoidBuffering = true
	} else {
		__antithesis_instrumentation__.Notify(16100)
		var err error
		sinkURIFn, err = p.TypeAsString(ctx, changefeedStmt.SinkURI, `CREATE CHANGEFEED`)
		if err != nil {
			__antithesis_instrumentation__.Notify(16102)
			return nil, nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(16103)
		}
		__antithesis_instrumentation__.Notify(16101)
		header = colinfo.ResultColumns{
			{Name: "job_id", Typ: types.Int},
		}
	}
	__antithesis_instrumentation__.Notify(16091)

	optsFn, err := p.TypeAsStringOpts(ctx, changefeedStmt.Options, changefeedbase.ChangefeedOptionExpectValues)
	if err != nil {
		__antithesis_instrumentation__.Notify(16104)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(16105)
	}
	__antithesis_instrumentation__.Notify(16092)

	rowFn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(16106)
		ctx, span := tracing.ChildSpan(ctx, stmt.StatementTag())
		defer span.Finish()

		if err := validateSettings(ctx, p); err != nil {
			__antithesis_instrumentation__.Notify(16115)
			return err
		} else {
			__antithesis_instrumentation__.Notify(16116)
		}
		__antithesis_instrumentation__.Notify(16107)

		sinkURI, err := sinkURIFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(16117)
			return changefeedbase.MarkTaggedError(err, changefeedbase.UserInput)
		} else {
			__antithesis_instrumentation__.Notify(16118)
		}
		__antithesis_instrumentation__.Notify(16108)

		if !unspecifiedSink && func() bool {
			__antithesis_instrumentation__.Notify(16119)
			return sinkURI == `` == true
		}() == true {
			__antithesis_instrumentation__.Notify(16120)

			return errors.New(`omit the SINK clause for inline results`)
		} else {
			__antithesis_instrumentation__.Notify(16121)
		}
		__antithesis_instrumentation__.Notify(16109)

		opts, err := optsFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(16122)
			return err
		} else {
			__antithesis_instrumentation__.Notify(16123)
		}
		__antithesis_instrumentation__.Notify(16110)

		jr, err := createChangefeedJobRecord(
			ctx,
			p,
			changefeedStmt,
			sinkURI,
			opts,
			jobspb.InvalidJobID,
			`changefeed.create`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(16124)
			return changefeedbase.MarkTaggedError(err, changefeedbase.UserInput)
		} else {
			__antithesis_instrumentation__.Notify(16125)
		}
		__antithesis_instrumentation__.Notify(16111)

		details := jr.Details.(jobspb.ChangefeedDetails)
		progress := jobspb.Progress{
			Progress: &jobspb.Progress_HighWater{},
			Details: &jobspb.Progress_Changefeed{
				Changefeed: &jobspb.ChangefeedProgress{},
			},
		}

		if details.SinkURI == `` {
			__antithesis_instrumentation__.Notify(16126)

			p.ExtendedEvalContext().Descs.ReleaseAll(ctx)

			telemetry.Count(`changefeed.create.core`)
			logChangefeedCreateTelemetry(ctx, jr)
			err := distChangefeedFlow(ctx, p, 0, details, progress, resultsCh)
			if err != nil {
				__antithesis_instrumentation__.Notify(16128)
				telemetry.Count(`changefeed.core.error`)
			} else {
				__antithesis_instrumentation__.Notify(16129)
			}
			__antithesis_instrumentation__.Notify(16127)
			return changefeedbase.MaybeStripRetryableErrorMarker(err)
		} else {
			__antithesis_instrumentation__.Notify(16130)
		}
		__antithesis_instrumentation__.Notify(16112)

		var sj *jobs.StartableJob
		jobID := p.ExecCfg().JobRegistry.MakeJobID()
		{
			__antithesis_instrumentation__.Notify(16131)
			var ptr *ptpb.Record
			codec := p.ExecCfg().Codec

			activeTimestampProtection := changefeedbase.ActiveProtectedTimestampsEnabled.Get(&p.ExecCfg().Settings.SV)
			initialScanType, err := initialScanTypeFromOpts(details.Opts)
			if err != nil {
				__antithesis_instrumentation__.Notify(16134)
				return err
			} else {
				__antithesis_instrumentation__.Notify(16135)
			}
			__antithesis_instrumentation__.Notify(16132)
			shouldProtectTimestamp := activeTimestampProtection || func() bool {
				__antithesis_instrumentation__.Notify(16136)
				return (initialScanType != changefeedbase.NoInitialScan) == true
			}() == true
			if shouldProtectTimestamp {
				__antithesis_instrumentation__.Notify(16137)
				ptr = createProtectedTimestampRecord(ctx, codec, jobID, AllTargets(details), details.StatementTime, progress.GetChangefeed())
			} else {
				__antithesis_instrumentation__.Notify(16138)
			}
			__antithesis_instrumentation__.Notify(16133)

			jr.Progress = *progress.GetChangefeed()

			if err := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(16139)
				if err := p.ExecCfg().JobRegistry.CreateStartableJobWithTxn(ctx, &sj, jobID, txn, *jr); err != nil {
					__antithesis_instrumentation__.Notify(16142)
					return err
				} else {
					__antithesis_instrumentation__.Notify(16143)
				}
				__antithesis_instrumentation__.Notify(16140)
				if ptr != nil {
					__antithesis_instrumentation__.Notify(16144)
					return p.ExecCfg().ProtectedTimestampProvider.Protect(ctx, txn, ptr)
				} else {
					__antithesis_instrumentation__.Notify(16145)
				}
				__antithesis_instrumentation__.Notify(16141)
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(16146)
				if sj != nil {
					__antithesis_instrumentation__.Notify(16148)
					if err := sj.CleanupOnRollback(ctx); err != nil {
						__antithesis_instrumentation__.Notify(16149)
						log.Warningf(ctx, "failed to cleanup aborted job: %v", err)
					} else {
						__antithesis_instrumentation__.Notify(16150)
					}
				} else {
					__antithesis_instrumentation__.Notify(16151)
				}
				__antithesis_instrumentation__.Notify(16147)
				return err
			} else {
				__antithesis_instrumentation__.Notify(16152)
			}
		}
		__antithesis_instrumentation__.Notify(16113)

		if err := sj.Start(ctx); err != nil {
			__antithesis_instrumentation__.Notify(16153)
			return err
		} else {
			__antithesis_instrumentation__.Notify(16154)
		}
		__antithesis_instrumentation__.Notify(16114)

		logChangefeedCreateTelemetry(ctx, jr)

		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(16155)
			return ctx.Err()
		case resultsCh <- tree.Datums{
			tree.NewDInt(tree.DInt(jobID)),
		}:
			__antithesis_instrumentation__.Notify(16156)
			return nil
		}
	}
	__antithesis_instrumentation__.Notify(16093)

	rowFnLogErrors := func(ctx context.Context, pn []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(16157)
		err := rowFn(ctx, pn, resultsCh)
		if err != nil {
			__antithesis_instrumentation__.Notify(16159)
			logChangefeedFailedTelemetry(ctx, nil, failureTypeForStartupError(err))
		} else {
			__antithesis_instrumentation__.Notify(16160)
		}
		__antithesis_instrumentation__.Notify(16158)
		return err
	}
	__antithesis_instrumentation__.Notify(16094)
	return rowFnLogErrors, header, nil, avoidBuffering, nil
}

func createChangefeedJobRecord(
	ctx context.Context,
	p sql.PlanHookState,
	changefeedStmt *annotatedChangefeedStatement,
	sinkURI string,
	opts map[string]string,
	jobID jobspb.JobID,
	telemetryPath string,
) (*jobs.Record, error) {
	__antithesis_instrumentation__.Notify(16161)
	unspecifiedSink := changefeedStmt.SinkURI == nil

	for key, value := range opts {
		__antithesis_instrumentation__.Notify(16183)
		if clusterVersion, ok := changefeedbase.VersionGateOptions[key]; ok {
			__antithesis_instrumentation__.Notify(16185)
			if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterVersion) {
				__antithesis_instrumentation__.Notify(16186)
				return nil, errors.Newf(
					`option %s is not supported until upgrade to version %s or higher is finalized`,
					key, clusterVersion.String(),
				)
			} else {
				__antithesis_instrumentation__.Notify(16187)
			}
		} else {
			__antithesis_instrumentation__.Notify(16188)
		}
		__antithesis_instrumentation__.Notify(16184)

		if _, ok := changefeedbase.CaseInsensitiveOpts[key]; ok {
			__antithesis_instrumentation__.Notify(16189)
			opts[key] = strings.ToLower(value)
		} else {
			__antithesis_instrumentation__.Notify(16190)
		}
	}
	__antithesis_instrumentation__.Notify(16162)

	if newFormat, ok := changefeedbase.NoLongerExperimental[opts[changefeedbase.OptFormat]]; ok {
		__antithesis_instrumentation__.Notify(16191)
		p.BufferClientNotice(ctx, pgnotice.Newf(
			`%[1]s is no longer experimental, use %[2]s=%[1]s`,
			newFormat, changefeedbase.OptFormat),
		)

	} else {
		__antithesis_instrumentation__.Notify(16192)
	}
	__antithesis_instrumentation__.Notify(16163)

	jobDescription, err := changefeedJobDescription(p, changefeedStmt.CreateChangefeed, sinkURI, opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(16193)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16194)
	}
	__antithesis_instrumentation__.Notify(16164)

	statementTime := hlc.Timestamp{
		WallTime: p.ExtendedEvalContext().GetStmtTimestamp().UnixNano(),
	}
	var initialHighWater hlc.Timestamp
	if cursor, ok := opts[changefeedbase.OptCursor]; ok {
		__antithesis_instrumentation__.Notify(16195)
		asOfClause := tree.AsOfClause{Expr: tree.NewStrVal(cursor)}
		var err error
		asOf, err := p.EvalAsOfTimestamp(ctx, asOfClause)
		if err != nil {
			__antithesis_instrumentation__.Notify(16197)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16198)
		}
		__antithesis_instrumentation__.Notify(16196)
		initialHighWater = asOf.Timestamp
		statementTime = initialHighWater
	} else {
		__antithesis_instrumentation__.Notify(16199)
	}
	__antithesis_instrumentation__.Notify(16165)

	endTime := hlc.Timestamp{}
	if endTimeOpt, ok := opts[changefeedbase.OptEndTime]; ok {
		__antithesis_instrumentation__.Notify(16200)
		asOfClause := tree.AsOfClause{Expr: tree.NewStrVal(endTimeOpt)}
		asOf, err := tree.EvalAsOfTimestamp(ctx, asOfClause, p.SemaCtx(), &p.ExtendedEvalContext().EvalContext)
		if err != nil {
			__antithesis_instrumentation__.Notify(16202)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16203)
		}
		__antithesis_instrumentation__.Notify(16201)
		endTime = asOf.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(16204)
	}

	{
		__antithesis_instrumentation__.Notify(16205)
		initialScanType, err := initialScanTypeFromOpts(opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(16207)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16208)
		}
		__antithesis_instrumentation__.Notify(16206)
		if initialScanType == changefeedbase.OnlyInitialScan {
			__antithesis_instrumentation__.Notify(16209)
			endTime = statementTime
		} else {
			__antithesis_instrumentation__.Notify(16210)
		}
	}
	__antithesis_instrumentation__.Notify(16166)

	targetList := uniqueTableNames(changefeedStmt.Targets)

	targetDescs, err := getTableDescriptors(ctx, p, &targetList, statementTime, initialHighWater)
	if err != nil {
		__antithesis_instrumentation__.Notify(16211)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16212)
	}
	__antithesis_instrumentation__.Notify(16167)

	targets, tables, err := getTargetsAndTables(ctx, p, targetDescs, changefeedStmt.Targets, changefeedStmt.originalSpecs, opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(16213)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16214)
	}
	__antithesis_instrumentation__.Notify(16168)
	for _, desc := range targetDescs {
		__antithesis_instrumentation__.Notify(16215)
		if table, isTable := desc.(catalog.TableDescriptor); isTable {
			__antithesis_instrumentation__.Notify(16216)
			if err := changefeedbase.ValidateTable(targets, table, opts); err != nil {
				__antithesis_instrumentation__.Notify(16218)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(16219)
			}
			__antithesis_instrumentation__.Notify(16217)
			for _, warning := range changefeedbase.WarningsForTable(tables, table, opts) {
				__antithesis_instrumentation__.Notify(16220)
				p.BufferClientNotice(ctx, pgnotice.Newf("%s", warning))
			}
		} else {
			__antithesis_instrumentation__.Notify(16221)
		}
	}
	__antithesis_instrumentation__.Notify(16169)

	details := jobspb.ChangefeedDetails{
		Tables:               tables,
		Opts:                 opts,
		SinkURI:              sinkURI,
		StatementTime:        statementTime,
		EndTime:              endTime,
		TargetSpecifications: targets,
	}

	parsedSink, err := url.Parse(sinkURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(16222)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16223)
	}
	__antithesis_instrumentation__.Notify(16170)
	if newScheme, ok := changefeedbase.NoLongerExperimental[parsedSink.Scheme]; ok {
		__antithesis_instrumentation__.Notify(16224)
		parsedSink.Scheme = newScheme
		p.BufferClientNotice(ctx, pgnotice.Newf(`%[1]s is no longer experimental, use %[1]s://`,
			newScheme),
		)
	} else {
		__antithesis_instrumentation__.Notify(16225)
	}
	__antithesis_instrumentation__.Notify(16171)

	if details, err = validateDetails(details); err != nil {
		__antithesis_instrumentation__.Notify(16226)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16227)
	}
	__antithesis_instrumentation__.Notify(16172)

	if _, err := getEncoder(details.Opts, AllTargets(details)); err != nil {
		__antithesis_instrumentation__.Notify(16228)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16229)
	}
	__antithesis_instrumentation__.Notify(16173)

	if isCloudStorageSink(parsedSink) || func() bool {
		__antithesis_instrumentation__.Notify(16230)
		return isWebhookSink(parsedSink) == true
	}() == true {
		__antithesis_instrumentation__.Notify(16231)
		details.Opts[changefeedbase.OptKeyInValue] = ``
	} else {
		__antithesis_instrumentation__.Notify(16232)
	}
	__antithesis_instrumentation__.Notify(16174)
	if isWebhookSink(parsedSink) {
		__antithesis_instrumentation__.Notify(16233)
		details.Opts[changefeedbase.OptTopicInValue] = ``
	} else {
		__antithesis_instrumentation__.Notify(16234)
	}
	__antithesis_instrumentation__.Notify(16175)

	if !unspecifiedSink && func() bool {
		__antithesis_instrumentation__.Notify(16235)
		return p.ExecCfg().ExternalIODirConfig.DisableOutbound == true
	}() == true {
		__antithesis_instrumentation__.Notify(16236)
		return nil, errors.Errorf("Outbound IO is disabled by configuration, cannot create changefeed into %s", parsedSink.Scheme)
	} else {
		__antithesis_instrumentation__.Notify(16237)
	}
	__antithesis_instrumentation__.Notify(16176)

	if telemetryPath != `` {
		__antithesis_instrumentation__.Notify(16238)

		telemetrySink := parsedSink.Scheme
		if telemetrySink == `` {
			__antithesis_instrumentation__.Notify(16240)
			telemetrySink = `sinkless`
		} else {
			__antithesis_instrumentation__.Notify(16241)
		}
		__antithesis_instrumentation__.Notify(16239)
		telemetry.Count(telemetryPath + `.sink.` + telemetrySink)
		telemetry.Count(telemetryPath + `.format.` + details.Opts[changefeedbase.OptFormat])
		telemetry.CountBucketed(telemetryPath+`.num_tables`, int64(len(tables)))
	} else {
		__antithesis_instrumentation__.Notify(16242)
	}
	__antithesis_instrumentation__.Notify(16177)

	if scope, ok := opts[changefeedbase.OptMetricsScope]; ok {
		__antithesis_instrumentation__.Notify(16243)
		if err := utilccl.CheckEnterpriseEnabled(
			p.ExecCfg().Settings, p.ExecCfg().LogicalClusterID(), p.ExecCfg().Organization(), "CHANGEFEED",
		); err != nil {
			__antithesis_instrumentation__.Notify(16245)
			return nil, errors.Wrapf(err,
				"use of %q option requires enterprise license.", changefeedbase.OptMetricsScope)
		} else {
			__antithesis_instrumentation__.Notify(16246)
		}
		__antithesis_instrumentation__.Notify(16244)

		if scope == defaultSLIScope {
			__antithesis_instrumentation__.Notify(16247)
			return nil, pgerror.Newf(pgcode.InvalidParameterValue,
				"%[1]q=%[2]q is the default metrics scope which keeps track of statistics "+
					"across all changefeeds without explicit label.  "+
					"If this is an intended behavior, please re-run the statement "+
					"without specifying %[1]q parameter.  "+
					"Otherwise, please re-run with a different %[1]q value.",
				changefeedbase.OptMetricsScope, defaultSLIScope)
		} else {
			__antithesis_instrumentation__.Notify(16248)
		}
	} else {
		__antithesis_instrumentation__.Notify(16249)
	}
	__antithesis_instrumentation__.Notify(16178)

	if details.SinkURI == `` {
		__antithesis_instrumentation__.Notify(16250)

		sinklessRecord := &jobs.Record{
			Details: details,
		}
		return sinklessRecord, nil
	} else {
		__antithesis_instrumentation__.Notify(16251)
	}
	__antithesis_instrumentation__.Notify(16179)

	if err := utilccl.CheckEnterpriseEnabled(
		p.ExecCfg().Settings, p.ExecCfg().LogicalClusterID(), p.ExecCfg().Organization(), "CHANGEFEED",
	); err != nil {
		__antithesis_instrumentation__.Notify(16252)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16253)
	}
	__antithesis_instrumentation__.Notify(16180)

	if telemetryPath != `` {
		__antithesis_instrumentation__.Notify(16254)
		telemetry.Count(telemetryPath + `.enterprise`)
	} else {
		__antithesis_instrumentation__.Notify(16255)
	}
	__antithesis_instrumentation__.Notify(16181)

	err = validateSink(ctx, p, jobID, details, opts)

	jr := &jobs.Record{
		Description: jobDescription,
		Username:    p.User(),
		DescriptorIDs: func() (sqlDescIDs []descpb.ID) {
			__antithesis_instrumentation__.Notify(16256)
			for _, desc := range targetDescs {
				__antithesis_instrumentation__.Notify(16258)
				sqlDescIDs = append(sqlDescIDs, desc.GetID())
			}
			__antithesis_instrumentation__.Notify(16257)
			return sqlDescIDs
		}(),
		Details: details,
	}
	__antithesis_instrumentation__.Notify(16182)

	return jr, err
}

func validateSettings(ctx context.Context, p sql.PlanHookState) error {
	__antithesis_instrumentation__.Notify(16259)
	if err := featureflag.CheckEnabled(
		ctx,
		p.ExecCfg(),
		featureChangefeedEnabled,
		"CHANGEFEED",
	); err != nil {
		__antithesis_instrumentation__.Notify(16264)
		return err
	} else {
		__antithesis_instrumentation__.Notify(16265)
	}
	__antithesis_instrumentation__.Notify(16260)

	if !kvserver.RangefeedEnabled.Get(&p.ExecCfg().Settings.SV) {
		__antithesis_instrumentation__.Notify(16266)
		return errors.Errorf("rangefeeds require the kv.rangefeed.enabled setting. See %s",
			docs.URL(`change-data-capture.html#enable-rangefeeds-to-reduce-latency`))
	} else {
		__antithesis_instrumentation__.Notify(16267)
	}
	__antithesis_instrumentation__.Notify(16261)

	ok, err := p.HasRoleOption(ctx, roleoption.CONTROLCHANGEFEED)
	if err != nil {
		__antithesis_instrumentation__.Notify(16268)
		return err
	} else {
		__antithesis_instrumentation__.Notify(16269)
	}
	__antithesis_instrumentation__.Notify(16262)
	if !ok {
		__antithesis_instrumentation__.Notify(16270)
		return pgerror.New(pgcode.InsufficientPrivilege, "current user must have a role WITH CONTROLCHANGEFEED")
	} else {
		__antithesis_instrumentation__.Notify(16271)
	}
	__antithesis_instrumentation__.Notify(16263)

	return nil
}

func getTableDescriptors(
	ctx context.Context,
	p sql.PlanHookState,
	targets *tree.TargetList,
	statementTime hlc.Timestamp,
	initialHighWater hlc.Timestamp,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(16272)

	if len(targets.Databases) > 0 {
		__antithesis_instrumentation__.Notify(16276)
		return nil, errors.Errorf(`CHANGEFEED cannot target %s`,
			tree.AsString(targets))
	} else {
		__antithesis_instrumentation__.Notify(16277)
	}
	__antithesis_instrumentation__.Notify(16273)
	for _, t := range targets.Tables {
		__antithesis_instrumentation__.Notify(16278)
		p, err := t.NormalizeTablePattern()
		if err != nil {
			__antithesis_instrumentation__.Notify(16280)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16281)
		}
		__antithesis_instrumentation__.Notify(16279)
		if _, ok := p.(*tree.TableName); !ok {
			__antithesis_instrumentation__.Notify(16282)
			return nil, errors.Errorf(`CHANGEFEED cannot target %s`, tree.AsString(t))
		} else {
			__antithesis_instrumentation__.Notify(16283)
		}
	}
	__antithesis_instrumentation__.Notify(16274)

	targetDescs, _, err := backupresolver.ResolveTargetsToDescriptors(
		ctx, p, statementTime, targets)
	if err != nil {
		__antithesis_instrumentation__.Notify(16284)
		var m *backupresolver.MissingTableErr
		if errors.As(err, &m) {
			__antithesis_instrumentation__.Notify(16286)
			tableName := m.GetTableName()
			err = errors.Errorf("table %q does not exist", tableName)
		} else {
			__antithesis_instrumentation__.Notify(16287)
		}
		__antithesis_instrumentation__.Notify(16285)
		err = errors.Wrap(err, "failed to resolve targets in the CHANGEFEED stmt")
		if !initialHighWater.IsEmpty() {
			__antithesis_instrumentation__.Notify(16288)

			err = errors.WithHintf(err,
				"do the targets exist at the specified cursor time %s?", initialHighWater)
		} else {
			__antithesis_instrumentation__.Notify(16289)
		}
	} else {
		__antithesis_instrumentation__.Notify(16290)
	}
	__antithesis_instrumentation__.Notify(16275)
	return targetDescs, err
}

func getTargetsAndTables(
	ctx context.Context,
	p sql.PlanHookState,
	targetDescs []catalog.Descriptor,
	rawTargets tree.ChangefeedTargets,
	originalSpecs map[tree.ChangefeedTarget]jobspb.ChangefeedTargetSpecification,
	opts map[string]string,
) ([]jobspb.ChangefeedTargetSpecification, jobspb.ChangefeedTargets, error) {
	__antithesis_instrumentation__.Notify(16291)
	tables := make(jobspb.ChangefeedTargets, len(targetDescs))
	targets := make([]jobspb.ChangefeedTargetSpecification, len(rawTargets))
	seen := make(map[jobspb.ChangefeedTargetSpecification]tree.ChangefeedTarget)

	for i, ct := range rawTargets {
		__antithesis_instrumentation__.Notify(16293)
		td, err := matchDescriptorToTablePattern(ctx, p, targetDescs, ct.TableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(16298)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(16299)
		}
		__antithesis_instrumentation__.Notify(16294)

		if err := p.CheckPrivilege(ctx, td, privilege.SELECT); err != nil {
			__antithesis_instrumentation__.Notify(16300)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(16301)
		}
		__antithesis_instrumentation__.Notify(16295)

		if spec, ok := originalSpecs[ct]; ok {
			__antithesis_instrumentation__.Notify(16302)
			targets[i] = spec
			if table, ok := tables[td.GetID()]; ok {
				__antithesis_instrumentation__.Notify(16303)
				if table.StatementTimeName != spec.StatementTimeName {
					__antithesis_instrumentation__.Notify(16304)
					return nil, nil, errors.Errorf(
						`table with id %d is referenced with multiple statement time names: %q and %q`, td.GetID(),
						table.StatementTimeName, spec.StatementTimeName)
				} else {
					__antithesis_instrumentation__.Notify(16305)
				}
			} else {
				__antithesis_instrumentation__.Notify(16306)
				tables[td.GetID()] = jobspb.ChangefeedTargetTable{
					StatementTimeName: spec.StatementTimeName,
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(16307)
			_, qualified := opts[changefeedbase.OptFullTableName]
			name, err := getChangefeedTargetName(ctx, td, p.ExecCfg(), p.ExtendedEvalContext().Txn, qualified)
			if err != nil {
				__antithesis_instrumentation__.Notify(16310)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(16311)
			}
			__antithesis_instrumentation__.Notify(16308)

			tables[td.GetID()] = jobspb.ChangefeedTargetTable{
				StatementTimeName: name,
			}
			typ := jobspb.ChangefeedTargetSpecification_PRIMARY_FAMILY_ONLY
			if ct.FamilyName != "" {
				__antithesis_instrumentation__.Notify(16312)
				typ = jobspb.ChangefeedTargetSpecification_COLUMN_FAMILY
			} else {
				__antithesis_instrumentation__.Notify(16313)
				if td.NumFamilies() > 1 {
					__antithesis_instrumentation__.Notify(16314)
					typ = jobspb.ChangefeedTargetSpecification_EACH_FAMILY
				} else {
					__antithesis_instrumentation__.Notify(16315)
				}
			}
			__antithesis_instrumentation__.Notify(16309)
			targets[i] = jobspb.ChangefeedTargetSpecification{
				Type:              typ,
				TableID:           td.GetID(),
				FamilyName:        string(ct.FamilyName),
				StatementTimeName: tables[td.GetID()].StatementTimeName,
			}
		}
		__antithesis_instrumentation__.Notify(16296)

		if dup, isDup := seen[targets[i]]; isDup {
			__antithesis_instrumentation__.Notify(16316)
			return nil, nil, errors.Errorf(
				"CHANGEFEED targets %s and %s are duplicates",
				tree.AsString(&dup), tree.AsString(&ct),
			)
		} else {
			__antithesis_instrumentation__.Notify(16317)
		}
		__antithesis_instrumentation__.Notify(16297)
		seen[targets[i]] = ct
	}
	__antithesis_instrumentation__.Notify(16292)
	return targets, tables, nil
}

func matchDescriptorToTablePattern(
	ctx context.Context, p sql.PlanHookState, descs []catalog.Descriptor, t tree.TablePattern,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(16318)
	pattern, err := t.NormalizeTablePattern()
	if err != nil {
		__antithesis_instrumentation__.Notify(16322)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(16323)
	}
	__antithesis_instrumentation__.Notify(16319)
	name, ok := pattern.(*tree.TableName)
	if !ok {
		__antithesis_instrumentation__.Notify(16324)
		return nil, errors.Newf("%v is not a TableName", pattern)
	} else {
		__antithesis_instrumentation__.Notify(16325)
	}
	__antithesis_instrumentation__.Notify(16320)
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(16326)
		tbl, ok := desc.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(16330)
			continue
		} else {
			__antithesis_instrumentation__.Notify(16331)
		}
		__antithesis_instrumentation__.Notify(16327)
		if tbl.GetName() != string(name.ObjectName) {
			__antithesis_instrumentation__.Notify(16332)
			continue
		} else {
			__antithesis_instrumentation__.Notify(16333)
		}
		__antithesis_instrumentation__.Notify(16328)
		qtn, err := getQualifiedTableNameObj(ctx, p.ExecCfg(), p.ExtendedEvalContext().Txn, tbl)
		if err != nil {
			__antithesis_instrumentation__.Notify(16334)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(16335)
		}
		__antithesis_instrumentation__.Notify(16329)
		switch name.ToUnresolvedObjectName().NumParts {
		case 1:
			__antithesis_instrumentation__.Notify(16336)
			if qtn.CatalogName == tree.Name(p.CurrentDatabase()) {
				__antithesis_instrumentation__.Notify(16340)
				return tbl, nil
			} else {
				__antithesis_instrumentation__.Notify(16341)
			}
		case 2:
			__antithesis_instrumentation__.Notify(16337)
			if qtn.CatalogName == name.SchemaName {
				__antithesis_instrumentation__.Notify(16342)
				return tbl, nil
			} else {
				__antithesis_instrumentation__.Notify(16343)
			}
		case 3:
			__antithesis_instrumentation__.Notify(16338)
			if qtn.CatalogName == name.CatalogName && func() bool {
				__antithesis_instrumentation__.Notify(16344)
				return qtn.SchemaName == name.SchemaName == true
			}() == true {
				__antithesis_instrumentation__.Notify(16345)
				return tbl, nil
			} else {
				__antithesis_instrumentation__.Notify(16346)
			}
		default:
			__antithesis_instrumentation__.Notify(16339)
		}
	}
	__antithesis_instrumentation__.Notify(16321)
	return nil, errors.Newf("could not match %v to a fetched descriptor", t)
}

func validateSink(
	ctx context.Context,
	p sql.PlanHookState,
	jobID jobspb.JobID,
	details jobspb.ChangefeedDetails,
	opts map[string]string,
) error {
	__antithesis_instrumentation__.Notify(16347)
	metrics := p.ExecCfg().JobRegistry.MetricsStruct().Changefeed.(*Metrics)
	sli, err := metrics.getSLIMetrics(opts[changefeedbase.OptMetricsScope])
	if err != nil {
		__antithesis_instrumentation__.Notify(16352)
		return err
	} else {
		__antithesis_instrumentation__.Notify(16353)
	}
	__antithesis_instrumentation__.Notify(16348)
	var nilOracle timestampLowerBoundOracle
	canarySink, err := getSink(ctx, &p.ExecCfg().DistSQLSrv.ServerConfig, details,
		nilOracle, p.User(), jobID, sli)
	if err != nil {
		__antithesis_instrumentation__.Notify(16354)
		return changefeedbase.MaybeStripRetryableErrorMarker(err)
	} else {
		__antithesis_instrumentation__.Notify(16355)
	}
	__antithesis_instrumentation__.Notify(16349)
	if err := canarySink.Close(); err != nil {
		__antithesis_instrumentation__.Notify(16356)
		return err
	} else {
		__antithesis_instrumentation__.Notify(16357)
	}
	__antithesis_instrumentation__.Notify(16350)
	if sink, ok := canarySink.(SinkWithTopics); ok {
		__antithesis_instrumentation__.Notify(16358)
		_, resolved := opts[changefeedbase.OptResolvedTimestamps]
		_, split := opts[changefeedbase.OptSplitColumnFamilies]
		if resolved && func() bool {
			__antithesis_instrumentation__.Notify(16361)
			return split == true
		}() == true {
			__antithesis_instrumentation__.Notify(16362)
			return errors.Newf("Resolved timestamps are not currently supported with %s for this sink"+
				" as the set of topics to fan them out to may change. Instead, use TABLE tablename FAMILY familyname"+
				" to specify individual families to watch.", changefeedbase.OptSplitColumnFamilies)
		} else {
			__antithesis_instrumentation__.Notify(16363)
		}
		__antithesis_instrumentation__.Notify(16359)

		topics := sink.Topics()
		for _, topic := range topics {
			__antithesis_instrumentation__.Notify(16364)
			p.BufferClientNotice(ctx, pgnotice.Newf(`changefeed will emit to topic %s`, topic))
		}
		__antithesis_instrumentation__.Notify(16360)
		details.Opts[changefeedbase.Topics] = strings.Join(topics, ",")
	} else {
		__antithesis_instrumentation__.Notify(16365)
	}
	__antithesis_instrumentation__.Notify(16351)
	return nil
}

func changefeedJobDescription(
	p sql.PlanHookState, changefeed *tree.CreateChangefeed, sinkURI string, opts map[string]string,
) (string, error) {
	__antithesis_instrumentation__.Notify(16366)
	cleanedSinkURI, err := cloud.SanitizeExternalStorageURI(sinkURI, []string{
		changefeedbase.SinkParamSASLPassword,
		changefeedbase.SinkParamCACert,
		changefeedbase.SinkParamClientCert,
	})

	if err != nil {
		__antithesis_instrumentation__.Notify(16370)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(16371)
	}
	__antithesis_instrumentation__.Notify(16367)

	cleanedSinkURI = redactUser(cleanedSinkURI)

	c := &tree.CreateChangefeed{
		Targets: changefeed.Targets,
		SinkURI: tree.NewDString(cleanedSinkURI),
	}
	for k, v := range opts {
		__antithesis_instrumentation__.Notify(16372)
		if k == changefeedbase.OptWebhookAuthHeader {
			__antithesis_instrumentation__.Notify(16375)
			v = redactWebhookAuthHeader(v)
		} else {
			__antithesis_instrumentation__.Notify(16376)
		}
		__antithesis_instrumentation__.Notify(16373)
		opt := tree.KVOption{Key: tree.Name(k)}
		if len(v) > 0 {
			__antithesis_instrumentation__.Notify(16377)
			opt.Value = tree.NewDString(v)
		} else {
			__antithesis_instrumentation__.Notify(16378)
		}
		__antithesis_instrumentation__.Notify(16374)
		c.Options = append(c.Options, opt)
	}
	__antithesis_instrumentation__.Notify(16368)
	sort.Slice(c.Options, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(16379)
		return c.Options[i].Key < c.Options[j].Key
	})
	__antithesis_instrumentation__.Notify(16369)
	ann := p.ExtendedEvalContext().Annotations
	return tree.AsStringWithFQNames(c, ann), nil
}

func redactUser(uri string) string {
	__antithesis_instrumentation__.Notify(16380)
	u, _ := url.Parse(uri)
	if u.User != nil {
		__antithesis_instrumentation__.Notify(16382)
		u.User = url.User(`redacted`)
	} else {
		__antithesis_instrumentation__.Notify(16383)
	}
	__antithesis_instrumentation__.Notify(16381)
	return u.String()
}

func validateNonNegativeDuration(optName string, optValue string) error {
	__antithesis_instrumentation__.Notify(16384)
	if d, err := time.ParseDuration(optValue); err != nil {
		__antithesis_instrumentation__.Notify(16386)
		return err
	} else {
		__antithesis_instrumentation__.Notify(16387)
		if d < 0 {
			__antithesis_instrumentation__.Notify(16388)
			return errors.Errorf("negative durations are not accepted: %s='%s'", optName, optValue)
		} else {
			__antithesis_instrumentation__.Notify(16389)
		}
	}
	__antithesis_instrumentation__.Notify(16385)
	return nil
}

func validateDetails(details jobspb.ChangefeedDetails) (jobspb.ChangefeedDetails, error) {
	__antithesis_instrumentation__.Notify(16390)
	if details.Opts == nil {
		__antithesis_instrumentation__.Notify(16392)

		details.Opts = map[string]string{}
	} else {
		__antithesis_instrumentation__.Notify(16393)
	}
	{
		__antithesis_instrumentation__.Notify(16394)
		const opt = changefeedbase.OptResolvedTimestamps
		if o, ok := details.Opts[opt]; ok && func() bool {
			__antithesis_instrumentation__.Notify(16395)
			return o != `` == true
		}() == true {
			__antithesis_instrumentation__.Notify(16396)
			if err := validateNonNegativeDuration(opt, o); err != nil {
				__antithesis_instrumentation__.Notify(16397)
				return jobspb.ChangefeedDetails{}, err
			} else {
				__antithesis_instrumentation__.Notify(16398)
			}
		} else {
			__antithesis_instrumentation__.Notify(16399)
		}
	}
	{
		__antithesis_instrumentation__.Notify(16400)
		const opt = changefeedbase.OptMinCheckpointFrequency
		if o, ok := details.Opts[opt]; ok && func() bool {
			__antithesis_instrumentation__.Notify(16401)
			return o != `` == true
		}() == true {
			__antithesis_instrumentation__.Notify(16402)
			if err := validateNonNegativeDuration(opt, o); err != nil {
				__antithesis_instrumentation__.Notify(16403)
				return jobspb.ChangefeedDetails{}, err
			} else {
				__antithesis_instrumentation__.Notify(16404)
			}
		} else {
			__antithesis_instrumentation__.Notify(16405)
		}
	}
	{
		__antithesis_instrumentation__.Notify(16406)
		const opt = changefeedbase.OptSchemaChangeEvents
		switch v := changefeedbase.SchemaChangeEventClass(details.Opts[opt]); v {
		case ``, changefeedbase.OptSchemaChangeEventClassDefault:
			__antithesis_instrumentation__.Notify(16407)
			details.Opts[opt] = string(changefeedbase.OptSchemaChangeEventClassDefault)
		case changefeedbase.OptSchemaChangeEventClassColumnChange:
			__antithesis_instrumentation__.Notify(16408)

		default:
			__antithesis_instrumentation__.Notify(16409)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`unknown %s: %s`, opt, v)
		}
	}
	{
		__antithesis_instrumentation__.Notify(16410)
		const opt = changefeedbase.OptSchemaChangePolicy
		switch v := changefeedbase.SchemaChangePolicy(details.Opts[opt]); v {
		case ``, changefeedbase.OptSchemaChangePolicyBackfill:
			__antithesis_instrumentation__.Notify(16411)
			details.Opts[opt] = string(changefeedbase.OptSchemaChangePolicyBackfill)
		case changefeedbase.OptSchemaChangePolicyNoBackfill:
			__antithesis_instrumentation__.Notify(16412)

		case changefeedbase.OptSchemaChangePolicyStop:
			__antithesis_instrumentation__.Notify(16413)

		default:
			__antithesis_instrumentation__.Notify(16414)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`unknown %s: %s`, opt, v)
		}
	}
	{
		__antithesis_instrumentation__.Notify(16415)
		const opt = changefeedbase.OptEnvelope
		switch v := changefeedbase.EnvelopeType(details.Opts[opt]); v {
		case changefeedbase.OptEnvelopeRow, changefeedbase.OptEnvelopeDeprecatedRow:
			__antithesis_instrumentation__.Notify(16416)
			details.Opts[opt] = string(changefeedbase.OptEnvelopeRow)
		case changefeedbase.OptEnvelopeKeyOnly:
			__antithesis_instrumentation__.Notify(16417)
			details.Opts[opt] = string(changefeedbase.OptEnvelopeKeyOnly)
		case ``, changefeedbase.OptEnvelopeWrapped:
			__antithesis_instrumentation__.Notify(16418)
			details.Opts[opt] = string(changefeedbase.OptEnvelopeWrapped)
		default:
			__antithesis_instrumentation__.Notify(16419)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`unknown %s: %s`, opt, v)
		}
	}
	{
		__antithesis_instrumentation__.Notify(16420)
		const opt = changefeedbase.OptFormat
		switch v := changefeedbase.FormatType(details.Opts[opt]); v {
		case ``, changefeedbase.OptFormatJSON:
			__antithesis_instrumentation__.Notify(16421)
			details.Opts[opt] = string(changefeedbase.OptFormatJSON)
		case changefeedbase.OptFormatAvro, changefeedbase.DeprecatedOptFormatAvro:
			__antithesis_instrumentation__.Notify(16422)

		default:
			__antithesis_instrumentation__.Notify(16423)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`unknown %s: %s`, opt, v)
		}
	}
	{
		__antithesis_instrumentation__.Notify(16424)
		const opt = changefeedbase.OptOnError
		switch v := changefeedbase.OnErrorType(details.Opts[opt]); v {
		case ``, changefeedbase.OptOnErrorFail:
			__antithesis_instrumentation__.Notify(16425)
			details.Opts[opt] = string(changefeedbase.OptOnErrorFail)
		case changefeedbase.OptOnErrorPause:
			__antithesis_instrumentation__.Notify(16426)

		default:
			__antithesis_instrumentation__.Notify(16427)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`unknown %s: %s, valid values are '%s' and '%s'`, opt, v,
				changefeedbase.OptOnErrorPause,
				changefeedbase.OptOnErrorFail)
		}
	}
	{
		__antithesis_instrumentation__.Notify(16428)
		const opt = changefeedbase.OptVirtualColumns
		switch v := changefeedbase.VirtualColumnVisibility(details.Opts[opt]); v {
		case ``, changefeedbase.OptVirtualColumnsOmitted:
			__antithesis_instrumentation__.Notify(16429)
			details.Opts[opt] = string(changefeedbase.OptVirtualColumnsOmitted)
		case changefeedbase.OptVirtualColumnsNull:
			__antithesis_instrumentation__.Notify(16430)
			details.Opts[opt] = string(changefeedbase.OptVirtualColumnsNull)
		default:
			__antithesis_instrumentation__.Notify(16431)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`unknown %s: %s`, opt, v)
		}
	}
	{
		__antithesis_instrumentation__.Notify(16432)
		initialScanType := details.Opts[changefeedbase.OptInitialScan]
		_, onlyInitialScan := details.Opts[changefeedbase.OptInitialScanOnly]
		_, endTime := details.Opts[changefeedbase.OptEndTime]
		if endTime && func() bool {
			__antithesis_instrumentation__.Notify(16435)
			return onlyInitialScan == true
		}() == true {
			__antithesis_instrumentation__.Notify(16436)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`cannot specify both %s and %s`, changefeedbase.OptInitialScanOnly,
				changefeedbase.OptEndTime)
		} else {
			__antithesis_instrumentation__.Notify(16437)
		}
		__antithesis_instrumentation__.Notify(16433)

		if strings.ToLower(initialScanType) == `only` && func() bool {
			__antithesis_instrumentation__.Notify(16438)
			return endTime == true
		}() == true {
			__antithesis_instrumentation__.Notify(16439)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`cannot specify both %s='only' and %s`, changefeedbase.OptInitialScan, changefeedbase.OptEndTime)
		} else {
			__antithesis_instrumentation__.Notify(16440)
		}
		__antithesis_instrumentation__.Notify(16434)

		if !details.EndTime.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(16441)
			return details.EndTime.Less(details.StatementTime) == true
		}() == true {
			__antithesis_instrumentation__.Notify(16442)
			return jobspb.ChangefeedDetails{}, errors.Errorf(
				`specified end time %s cannot be less than statement time %s`,
				details.EndTime.AsOfSystemTime(),
				details.StatementTime.AsOfSystemTime(),
			)
		} else {
			__antithesis_instrumentation__.Notify(16443)
		}
	}
	__antithesis_instrumentation__.Notify(16391)
	return details, nil
}

type changefeedResumer struct {
	job *jobs.Job
}

func (b *changefeedResumer) setJobRunningStatus(
	ctx context.Context, lastUpdate time.Time, fmtOrMsg string, args ...interface{},
) time.Time {
	__antithesis_instrumentation__.Notify(16444)
	if timeutil.Since(lastUpdate) < runStatusUpdateFrequency {
		__antithesis_instrumentation__.Notify(16447)
		return lastUpdate
	} else {
		__antithesis_instrumentation__.Notify(16448)
	}
	__antithesis_instrumentation__.Notify(16445)

	status := jobs.RunningStatus(fmt.Sprintf(fmtOrMsg, args...))
	if err := b.job.RunningStatus(ctx, nil,
		func(_ context.Context, _ jobspb.Details) (jobs.RunningStatus, error) {
			__antithesis_instrumentation__.Notify(16449)
			return status, nil
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(16450)
		log.Warningf(ctx, "failed to set running status: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(16451)
	}
	__antithesis_instrumentation__.Notify(16446)

	return timeutil.Now()
}

func (b *changefeedResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(16452)
	jobExec := execCtx.(sql.JobExecContext)
	execCfg := jobExec.ExecCfg()
	jobID := b.job.ID()
	details := b.job.Details().(jobspb.ChangefeedDetails)
	progress := b.job.Progress()

	err := b.resumeWithRetries(ctx, jobExec, jobID, details, progress, execCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(16454)
		return b.handleChangefeedError(ctx, err, details, jobExec)
	} else {
		__antithesis_instrumentation__.Notify(16455)
	}
	__antithesis_instrumentation__.Notify(16453)
	return nil
}

func (b *changefeedResumer) handleChangefeedError(
	ctx context.Context,
	changefeedErr error,
	details jobspb.ChangefeedDetails,
	jobExec sql.JobExecContext,
) error {
	__antithesis_instrumentation__.Notify(16456)
	switch onError := changefeedbase.OnErrorType(details.Opts[changefeedbase.OptOnError]); onError {

	case changefeedbase.OptOnErrorFail:
		__antithesis_instrumentation__.Notify(16457)
		return changefeedErr

	case changefeedbase.OptOnErrorPause:
		__antithesis_instrumentation__.Notify(16458)

		const errorFmt = "job failed (%v) but is being paused because of %s=%s"
		errorMessage := fmt.Sprintf(errorFmt, changefeedErr,
			changefeedbase.OptOnError, changefeedbase.OptOnErrorPause)
		return b.job.PauseRequested(ctx, jobExec.ExtendedEvalContext().Txn, func(ctx context.Context,
			planHookState interface{}, txn *kv.Txn, progress *jobspb.Progress) error {
			__antithesis_instrumentation__.Notify(16460)
			err := b.OnPauseRequest(ctx, jobExec, txn, progress)
			if err != nil {
				__antithesis_instrumentation__.Notify(16462)
				return err
			} else {
				__antithesis_instrumentation__.Notify(16463)
			}
			__antithesis_instrumentation__.Notify(16461)

			progress.RunningStatus = errorMessage
			log.Warningf(ctx, errorFmt, changefeedErr, changefeedbase.OptOnError, changefeedbase.OptOnErrorPause)
			return nil
		}, errorMessage)
	default:
		__antithesis_instrumentation__.Notify(16459)
		return errors.Errorf("unrecognized option value: %s=%s",
			changefeedbase.OptOnError, details.Opts[changefeedbase.OptOnError])
	}
}

func (b *changefeedResumer) resumeWithRetries(
	ctx context.Context,
	jobExec sql.JobExecContext,
	jobID jobspb.JobID,
	details jobspb.ChangefeedDetails,
	progress jobspb.Progress,
	execCfg *sql.ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(16464)

	opts := retry.Options{
		InitialBackoff: 5 * time.Millisecond,
		Multiplier:     2,
		MaxBackoff:     10 * time.Second,
	}
	var err error
	var lastRunStatusUpdate time.Time

	for r := retry.StartWithCtx(ctx, opts); r.Next(); {
		__antithesis_instrumentation__.Notify(16466)

		startedCh := make(chan tree.Datums, 1)

		if err = distChangefeedFlow(ctx, jobExec, jobID, details, progress, startedCh); err == nil {
			__antithesis_instrumentation__.Notify(16471)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(16472)
		}
		__antithesis_instrumentation__.Notify(16467)

		if knobs, ok := execCfg.DistSQLSrv.TestingKnobs.Changefeed.(*TestingKnobs); ok {
			__antithesis_instrumentation__.Notify(16473)
			if knobs != nil && func() bool {
				__antithesis_instrumentation__.Notify(16474)
				return knobs.HandleDistChangefeedError != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(16475)
				err = knobs.HandleDistChangefeedError(err)
			} else {
				__antithesis_instrumentation__.Notify(16476)
			}
		} else {
			__antithesis_instrumentation__.Notify(16477)
		}
		__antithesis_instrumentation__.Notify(16468)

		if !changefeedbase.IsRetryableError(err) {
			__antithesis_instrumentation__.Notify(16478)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(16481)
				return ctx.Err()
			} else {
				__antithesis_instrumentation__.Notify(16482)
			}
			__antithesis_instrumentation__.Notify(16479)

			if flowinfra.IsFlowRetryableError(err) {
				__antithesis_instrumentation__.Notify(16483)

				err = jobs.MarkAsRetryJobError(err)
				lastRunStatusUpdate = b.setJobRunningStatus(ctx, lastRunStatusUpdate, "retryable flow error: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(16484)
			}
			__antithesis_instrumentation__.Notify(16480)

			log.Warningf(ctx, `CHANGEFEED job %d returning with error: %+v`, jobID, err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(16485)
		}
		__antithesis_instrumentation__.Notify(16469)

		log.Warningf(ctx, `WARNING: CHANGEFEED job %d encountered retryable error: %v`, jobID, err)
		lastRunStatusUpdate = b.setJobRunningStatus(ctx, lastRunStatusUpdate, "retryable error: %s", err)
		if metrics, ok := execCfg.JobRegistry.MetricsStruct().Changefeed.(*Metrics); ok {
			__antithesis_instrumentation__.Notify(16486)
			sli, err := metrics.getSLIMetrics(details.Opts[changefeedbase.OptMetricsScope])
			if err != nil {
				__antithesis_instrumentation__.Notify(16488)
				return err
			} else {
				__antithesis_instrumentation__.Notify(16489)
			}
			__antithesis_instrumentation__.Notify(16487)
			sli.ErrorRetries.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(16490)
		}
		__antithesis_instrumentation__.Notify(16470)

		reloadedJob, reloadErr := execCfg.JobRegistry.LoadClaimedJob(ctx, jobID)
		if reloadErr != nil {
			__antithesis_instrumentation__.Notify(16491)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(16493)
				return ctx.Err()
			} else {
				__antithesis_instrumentation__.Notify(16494)
			}
			__antithesis_instrumentation__.Notify(16492)
			log.Warningf(ctx, `CHANGEFEED job %d could not reload job progress; `+
				`continuing from last known high-water of %s: %v`,
				jobID, progress.GetHighWater(), reloadErr)
		} else {
			__antithesis_instrumentation__.Notify(16495)
			progress = reloadedJob.Progress()
		}
	}
	__antithesis_instrumentation__.Notify(16465)
	return errors.Wrap(err, `ran out of retries`)
}

func (b *changefeedResumer) OnFailOrCancel(ctx context.Context, jobExec interface{}) error {
	__antithesis_instrumentation__.Notify(16496)
	exec := jobExec.(sql.JobExecContext)
	execCfg := exec.ExecCfg()
	progress := b.job.Progress()
	b.maybeCleanUpProtectedTimestamp(ctx, execCfg.DB, execCfg.ProtectedTimestampProvider,
		progress.GetChangefeed().ProtectedTimestampRecord)

	if jobs.HasErrJobCanceled(
		errors.DecodeError(ctx, *b.job.Payload().FinalResumeError),
	) {
		__antithesis_instrumentation__.Notify(16498)
		telemetry.Count(`changefeed.enterprise.cancel`)
	} else {
		__antithesis_instrumentation__.Notify(16499)
		telemetry.Count(`changefeed.enterprise.fail`)
		exec.ExecCfg().JobRegistry.MetricsStruct().Changefeed.(*Metrics).Failures.Inc(1)
		logChangefeedFailedTelemetry(ctx, b.job, changefeedbase.UnknownError)
	}
	__antithesis_instrumentation__.Notify(16497)
	return nil
}

func (b *changefeedResumer) maybeCleanUpProtectedTimestamp(
	ctx context.Context, db *kv.DB, pts protectedts.Storage, ptsID uuid.UUID,
) {
	__antithesis_instrumentation__.Notify(16500)
	if ptsID == uuid.Nil {
		__antithesis_instrumentation__.Notify(16502)
		return
	} else {
		__antithesis_instrumentation__.Notify(16503)
	}
	__antithesis_instrumentation__.Notify(16501)
	if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(16504)
		return pts.Release(ctx, txn, ptsID)
	}); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(16505)
		return !errors.Is(err, protectedts.ErrNotExists) == true
	}() == true {
		__antithesis_instrumentation__.Notify(16506)

		log.Warningf(ctx, "failed to remove protected timestamp record %v: %v", ptsID, err)
	} else {
		__antithesis_instrumentation__.Notify(16507)
	}
}

var _ jobs.PauseRequester = (*changefeedResumer)(nil)

func (b *changefeedResumer) OnPauseRequest(
	ctx context.Context, jobExec interface{}, txn *kv.Txn, progress *jobspb.Progress,
) error {
	__antithesis_instrumentation__.Notify(16508)
	details := b.job.Details().(jobspb.ChangefeedDetails)

	cp := progress.GetChangefeed()
	execCfg := jobExec.(sql.JobExecContext).ExecCfg()

	if _, shouldProtect := details.Opts[changefeedbase.OptProtectDataFromGCOnPause]; !shouldProtect {
		__antithesis_instrumentation__.Notify(16511)

		if cp.ProtectedTimestampRecord != uuid.Nil {
			__antithesis_instrumentation__.Notify(16513)
			if err := execCfg.ProtectedTimestampProvider.Release(ctx, txn, cp.ProtectedTimestampRecord); err != nil {
				__antithesis_instrumentation__.Notify(16514)
				log.Warningf(ctx, "failed to release protected timestamp %v: %v", cp.ProtectedTimestampRecord, err)
			} else {
				__antithesis_instrumentation__.Notify(16515)
				cp.ProtectedTimestampRecord = uuid.Nil
			}
		} else {
			__antithesis_instrumentation__.Notify(16516)
		}
		__antithesis_instrumentation__.Notify(16512)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(16517)
	}
	__antithesis_instrumentation__.Notify(16509)

	if cp.ProtectedTimestampRecord == uuid.Nil {
		__antithesis_instrumentation__.Notify(16518)
		resolved := progress.GetHighWater()
		if resolved == nil {
			__antithesis_instrumentation__.Notify(16520)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(16521)
		}
		__antithesis_instrumentation__.Notify(16519)
		pts := execCfg.ProtectedTimestampProvider
		ptr := createProtectedTimestampRecord(ctx, execCfg.Codec, b.job.ID(), AllTargets(details), *resolved, cp)
		return pts.Protect(ctx, txn, ptr)
	} else {
		__antithesis_instrumentation__.Notify(16522)
	}
	__antithesis_instrumentation__.Notify(16510)

	return nil
}

func getQualifiedTableName(
	ctx context.Context, execCfg *sql.ExecutorConfig, txn *kv.Txn, desc catalog.TableDescriptor,
) (string, error) {
	__antithesis_instrumentation__.Notify(16523)
	tbName, err := getQualifiedTableNameObj(ctx, execCfg, txn, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(16525)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(16526)
	}
	__antithesis_instrumentation__.Notify(16524)
	return tbName.String(), nil
}

func getQualifiedTableNameObj(
	ctx context.Context, execCfg *sql.ExecutorConfig, txn *kv.Txn, desc catalog.TableDescriptor,
) (tree.TableName, error) {
	__antithesis_instrumentation__.Notify(16527)
	col := execCfg.CollectionFactory.MakeCollection(ctx, nil)
	dbDesc, err := col.Direct().MustGetDatabaseDescByID(ctx, txn, desc.GetParentID())
	if err != nil {
		__antithesis_instrumentation__.Notify(16530)
		return tree.TableName{}, err
	} else {
		__antithesis_instrumentation__.Notify(16531)
	}
	__antithesis_instrumentation__.Notify(16528)
	schemaID := desc.GetParentSchemaID()
	schemaName, err := resolver.ResolveSchemaNameByID(ctx, txn, execCfg.Codec, dbDesc, schemaID)
	if err != nil {
		__antithesis_instrumentation__.Notify(16532)
		return tree.TableName{}, err
	} else {
		__antithesis_instrumentation__.Notify(16533)
	}
	__antithesis_instrumentation__.Notify(16529)
	tbName := tree.MakeTableNameWithSchema(
		tree.Name(dbDesc.GetName()),
		tree.Name(schemaName),
		tree.Name(desc.GetName()),
	)
	return tbName, nil
}

func getChangefeedTargetName(
	ctx context.Context,
	desc catalog.TableDescriptor,
	execCfg *sql.ExecutorConfig,
	txn *kv.Txn,
	qualified bool,
) (string, error) {
	__antithesis_instrumentation__.Notify(16534)
	if qualified {
		__antithesis_instrumentation__.Notify(16536)
		return getQualifiedTableName(ctx, execCfg, txn, desc)
	} else {
		__antithesis_instrumentation__.Notify(16537)
	}
	__antithesis_instrumentation__.Notify(16535)
	return desc.GetName(), nil
}

func AllTargets(cd jobspb.ChangefeedDetails) (targets []jobspb.ChangefeedTargetSpecification) {
	__antithesis_instrumentation__.Notify(16538)

	if len(cd.TargetSpecifications) > 0 {
		__antithesis_instrumentation__.Notify(16540)
		for _, ts := range cd.TargetSpecifications {
			__antithesis_instrumentation__.Notify(16541)
			if ts.TableID > 0 {
				__antithesis_instrumentation__.Notify(16542)
				ts.StatementTimeName = cd.Tables[ts.TableID].StatementTimeName
				targets = append(targets, ts)
			} else {
				__antithesis_instrumentation__.Notify(16543)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(16544)
		for id, t := range cd.Tables {
			__antithesis_instrumentation__.Notify(16545)
			ct := jobspb.ChangefeedTargetSpecification{
				Type:              jobspb.ChangefeedTargetSpecification_PRIMARY_FAMILY_ONLY,
				TableID:           id,
				StatementTimeName: t.StatementTimeName,
			}
			targets = append(targets, ct)
		}
	}
	__antithesis_instrumentation__.Notify(16539)
	return
}

func uniqueTableNames(cts tree.ChangefeedTargets) tree.TargetList {
	__antithesis_instrumentation__.Notify(16546)
	uniqueTablePatterns := make(map[string]tree.TablePattern)
	for _, t := range cts {
		__antithesis_instrumentation__.Notify(16549)
		uniqueTablePatterns[t.TableName.String()] = t.TableName
	}
	__antithesis_instrumentation__.Notify(16547)

	targetList := tree.TargetList{}
	for _, t := range uniqueTablePatterns {
		__antithesis_instrumentation__.Notify(16550)
		targetList.Tables = append(targetList.Tables, t)
	}
	__antithesis_instrumentation__.Notify(16548)

	return targetList
}

func logChangefeedCreateTelemetry(ctx context.Context, jr *jobs.Record) {
	__antithesis_instrumentation__.Notify(16551)
	var changefeedEventDetails eventpb.CommonChangefeedEventDetails
	if jr != nil {
		__antithesis_instrumentation__.Notify(16553)
		changefeedDetails := jr.Details.(jobspb.ChangefeedDetails)
		changefeedEventDetails = getCommonChangefeedEventDetails(ctx, changefeedDetails, jr.Description)
	} else {
		__antithesis_instrumentation__.Notify(16554)
	}
	__antithesis_instrumentation__.Notify(16552)

	createChangefeedEvent := &eventpb.CreateChangefeed{
		CommonChangefeedEventDetails: changefeedEventDetails,
	}

	log.StructuredEvent(ctx, createChangefeedEvent)
}

func logChangefeedFailedTelemetry(
	ctx context.Context, job *jobs.Job, failureType changefeedbase.FailureType,
) {
	__antithesis_instrumentation__.Notify(16555)
	var changefeedEventDetails eventpb.CommonChangefeedEventDetails
	if job != nil {
		__antithesis_instrumentation__.Notify(16557)
		changefeedDetails := job.Details().(jobspb.ChangefeedDetails)
		changefeedEventDetails = getCommonChangefeedEventDetails(ctx, changefeedDetails, job.Payload().Description)
	} else {
		__antithesis_instrumentation__.Notify(16558)
	}
	__antithesis_instrumentation__.Notify(16556)

	changefeedFailedEvent := &eventpb.ChangefeedFailed{
		CommonChangefeedEventDetails: changefeedEventDetails,
		FailureType:                  failureType,
	}

	log.StructuredEvent(ctx, changefeedFailedEvent)
}

func getCommonChangefeedEventDetails(
	ctx context.Context, details jobspb.ChangefeedDetails, description string,
) eventpb.CommonChangefeedEventDetails {
	__antithesis_instrumentation__.Notify(16559)
	opts := details.Opts

	sinkType := "core"
	if details.SinkURI != `` {
		__antithesis_instrumentation__.Notify(16563)
		parsedSink, err := url.Parse(details.SinkURI)
		if err != nil {
			__antithesis_instrumentation__.Notify(16565)
			log.Warningf(ctx, "failed to parse sink for telemetry logging: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(16566)
		}
		__antithesis_instrumentation__.Notify(16564)
		sinkType = parsedSink.Scheme
	} else {
		__antithesis_instrumentation__.Notify(16567)
	}
	__antithesis_instrumentation__.Notify(16560)

	var initialScan string
	initialScanType, initialScanSet := opts[changefeedbase.OptInitialScan]
	_, initialScanOnlySet := opts[changefeedbase.OptInitialScanOnly]
	_, noInitialScanSet := opts[changefeedbase.OptNoInitialScan]
	if initialScanSet && func() bool {
		__antithesis_instrumentation__.Notify(16568)
		return initialScanType == `` == true
	}() == true {
		__antithesis_instrumentation__.Notify(16569)
		initialScan = `yes`
	} else {
		__antithesis_instrumentation__.Notify(16570)
		if initialScanSet && func() bool {
			__antithesis_instrumentation__.Notify(16571)
			return initialScanType != `` == true
		}() == true {
			__antithesis_instrumentation__.Notify(16572)
			initialScan = initialScanType
		} else {
			__antithesis_instrumentation__.Notify(16573)
			if initialScanOnlySet {
				__antithesis_instrumentation__.Notify(16574)
				initialScan = `only`
			} else {
				__antithesis_instrumentation__.Notify(16575)
				if noInitialScanSet {
					__antithesis_instrumentation__.Notify(16576)
					initialScan = `no`
				} else {
					__antithesis_instrumentation__.Notify(16577)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(16561)

	var resolved string
	resolvedValue, resolvedSet := opts[changefeedbase.OptResolvedTimestamps]
	if !resolvedSet {
		__antithesis_instrumentation__.Notify(16578)
		resolved = "no"
	} else {
		__antithesis_instrumentation__.Notify(16579)
		if resolved == `` {
			__antithesis_instrumentation__.Notify(16580)
			resolved = "yes"
		} else {
			__antithesis_instrumentation__.Notify(16581)
			resolved = resolvedValue
		}
	}
	__antithesis_instrumentation__.Notify(16562)

	changefeedEventDetails := eventpb.CommonChangefeedEventDetails{
		Description: description,
		SinkType:    sinkType,
		NumTables:   int32(len(AllTargets(details))),
		Resolved:    resolved,
		Format:      opts[changefeedbase.OptFormat],
		InitialScan: initialScan,
	}

	return changefeedEventDetails
}

func failureTypeForStartupError(err error) changefeedbase.FailureType {
	__antithesis_instrumentation__.Notify(16582)
	if errors.Is(err, context.Canceled) {
		__antithesis_instrumentation__.Notify(16584)
		return changefeedbase.ConnectionClosed
	} else {
		__antithesis_instrumentation__.Notify(16585)
		if isTagged, tag := changefeedbase.IsTaggedError(err); isTagged {
			__antithesis_instrumentation__.Notify(16586)
			return tag
		} else {
			__antithesis_instrumentation__.Notify(16587)
		}
	}
	__antithesis_instrumentation__.Notify(16583)
	return changefeedbase.OnStartup
}
