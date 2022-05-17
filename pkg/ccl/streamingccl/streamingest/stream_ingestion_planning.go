package streamingest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"
	"os"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

func streamIngestionJobDescription(
	p sql.PlanHookState, streamIngestion *tree.StreamIngestion,
) (string, error) {
	__antithesis_instrumentation__.Notify(25350)
	ann := p.ExtendedEvalContext().Annotations
	return tree.AsStringWithFQNames(streamIngestion, ann), nil
}

func maybeAddInlineSecurityCredentials(pgURL url.URL) (url.URL, error) {
	__antithesis_instrumentation__.Notify(25351)
	options := pgURL.Query()
	if options.Get("sslinline") == "true" {
		__antithesis_instrumentation__.Notify(25356)
		return pgURL, nil
	} else {
		__antithesis_instrumentation__.Notify(25357)
	}
	__antithesis_instrumentation__.Notify(25352)

	loadPathContentAsOption := func(optionKey string) error {
		__antithesis_instrumentation__.Notify(25358)
		if !options.Has(optionKey) {
			__antithesis_instrumentation__.Notify(25361)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(25362)
		}
		__antithesis_instrumentation__.Notify(25359)
		content, err := os.ReadFile(options.Get(optionKey))
		if err != nil {
			__antithesis_instrumentation__.Notify(25363)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25364)
		}
		__antithesis_instrumentation__.Notify(25360)
		options.Set(optionKey, string(content))
		return nil
	}
	__antithesis_instrumentation__.Notify(25353)

	if err := loadPathContentAsOption("sslrootcert"); err != nil {
		__antithesis_instrumentation__.Notify(25365)
		return url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(25366)
	}
	__antithesis_instrumentation__.Notify(25354)

	if options.Get("sslmode") == "verify-full" {
		__antithesis_instrumentation__.Notify(25367)
		if err := loadPathContentAsOption("sslcert"); err != nil {
			__antithesis_instrumentation__.Notify(25369)
			return url.URL{}, err
		} else {
			__antithesis_instrumentation__.Notify(25370)
		}
		__antithesis_instrumentation__.Notify(25368)
		if err := loadPathContentAsOption("sslkey"); err != nil {
			__antithesis_instrumentation__.Notify(25371)
			return url.URL{}, err
		} else {
			__antithesis_instrumentation__.Notify(25372)
		}
	} else {
		__antithesis_instrumentation__.Notify(25373)
	}
	__antithesis_instrumentation__.Notify(25355)
	options.Set("sslinline", "true")
	res := pgURL
	res.RawQuery = options.Encode()
	return res, nil
}

func ingestionPlanHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(25374)
	ingestionStmt, ok := stmt.(*tree.StreamIngestion)
	if !ok {
		__antithesis_instrumentation__.Notify(25379)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(25380)
	}
	__antithesis_instrumentation__.Notify(25375)

	if !p.SessionData().EnableStreamReplication {
		__antithesis_instrumentation__.Notify(25381)
		return nil, nil, nil, false, errors.WithTelemetry(
			pgerror.WithCandidateCode(
				errors.WithHint(
					errors.Newf("stream replication is only supported experimentally"),
					"You can enable stream replication by running `SET enable_experimental_stream_replication = true`.",
				),
				pgcode.FeatureNotSupported,
			),
			"replication.ingest.disabled",
		)
	} else {
		__antithesis_instrumentation__.Notify(25382)
	}
	__antithesis_instrumentation__.Notify(25376)

	fromFn, err := p.TypeAsStringArray(ctx, tree.Exprs(ingestionStmt.From), "INGESTION")
	if err != nil {
		__antithesis_instrumentation__.Notify(25383)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(25384)
	}
	__antithesis_instrumentation__.Notify(25377)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(25385)
		ctx, span := tracing.ChildSpan(ctx, stmt.StatementTag())
		defer span.Finish()

		if err := utilccl.CheckEnterpriseEnabled(
			p.ExecCfg().Settings, p.ExecCfg().LogicalClusterID(), p.ExecCfg().Organization(),
			"RESTORE FROM REPLICATION STREAM",
		); err != nil {
			__antithesis_instrumentation__.Notify(25396)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25397)
		}
		__antithesis_instrumentation__.Notify(25386)

		from, err := fromFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(25398)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25399)
		}
		__antithesis_instrumentation__.Notify(25387)

		if !ingestionStmt.Targets.TenantID.IsSet() {
			__antithesis_instrumentation__.Notify(25400)
			return errors.Newf("no tenant specified in ingestion query: %s", ingestionStmt.String())
		} else {
			__antithesis_instrumentation__.Notify(25401)
		}
		__antithesis_instrumentation__.Notify(25388)

		streamAddress := streamingccl.StreamAddress(from[0])
		url, err := streamAddress.URL()
		if err != nil {
			__antithesis_instrumentation__.Notify(25402)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25403)
		}
		__antithesis_instrumentation__.Notify(25389)
		q := url.Query()

		if hasPostgresAuthentication := (q.Get("sslmode") == "verify-full") && func() bool {
			__antithesis_instrumentation__.Notify(25404)
			return q.Has("sslrootcert") == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(25405)
			return q.Has("sslkey") == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(25406)
			return q.Has("sslcert") == true
		}() == true; (url.Scheme == "postgres") && func() bool {
			__antithesis_instrumentation__.Notify(25407)
			return !hasPostgresAuthentication == true
		}() == true {
			__antithesis_instrumentation__.Notify(25408)
			return errors.Errorf(
				"stream replication address should have cert authentication if in postgres scheme: %s", streamAddress)
		} else {
			__antithesis_instrumentation__.Notify(25409)
		}
		__antithesis_instrumentation__.Notify(25390)

		*url, err = maybeAddInlineSecurityCredentials(*url)
		if err != nil {
			__antithesis_instrumentation__.Notify(25410)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25411)
		}
		__antithesis_instrumentation__.Notify(25391)
		streamAddress = streamingccl.StreamAddress(url.String())

		if ingestionStmt.Targets.Types != nil || func() bool {
			__antithesis_instrumentation__.Notify(25412)
			return ingestionStmt.Targets.Databases != nil == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(25413)
			return ingestionStmt.Targets.Tables != nil == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(25414)
			return ingestionStmt.Targets.Schemas != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(25415)
			return errors.Newf("unsupported target in ingestion query, "+
				"only tenant ingestion is supported: %s", ingestionStmt.String())
		} else {
			__antithesis_instrumentation__.Notify(25416)
		}
		__antithesis_instrumentation__.Notify(25392)

		prefix := keys.MakeTenantPrefix(ingestionStmt.Targets.TenantID.TenantID)
		startTime := hlc.Timestamp{WallTime: timeutil.Now().UnixNano()}
		if ingestionStmt.AsOf.Expr != nil {
			__antithesis_instrumentation__.Notify(25417)
			asOf, err := p.EvalAsOfTimestamp(ctx, ingestionStmt.AsOf)
			if err != nil {
				__antithesis_instrumentation__.Notify(25419)
				return err
			} else {
				__antithesis_instrumentation__.Notify(25420)
			}
			__antithesis_instrumentation__.Notify(25418)
			startTime = asOf.Timestamp
		} else {
			__antithesis_instrumentation__.Notify(25421)
		}
		__antithesis_instrumentation__.Notify(25393)

		streamIngestionDetails := jobspb.StreamIngestionDetails{
			StreamAddress: string(streamAddress),
			TenantID:      ingestionStmt.Targets.TenantID.TenantID,
			Span:          roachpb.Span{Key: prefix, EndKey: prefix.PrefixEnd()},
			StartTime:     startTime,
		}

		jobDescription, err := streamIngestionJobDescription(p, ingestionStmt)
		if err != nil {
			__antithesis_instrumentation__.Notify(25422)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25423)
		}
		__antithesis_instrumentation__.Notify(25394)

		jr := jobs.Record{
			Description:   jobDescription,
			Username:      p.User(),
			Progress:      jobspb.StreamIngestionProgress{},
			Details:       streamIngestionDetails,
			NonCancelable: true,
		}

		jobID := p.ExecCfg().JobRegistry.MakeJobID()
		sj, err := p.ExecCfg().JobRegistry.CreateAdoptableJobWithTxn(ctx, jr,
			jobID, p.ExtendedEvalContext().Txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(25424)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25425)
		}
		__antithesis_instrumentation__.Notify(25395)
		resultsCh <- tree.Datums{tree.NewDInt(tree.DInt(sj.ID()))}
		return nil
	}
	__antithesis_instrumentation__.Notify(25378)

	return fn, jobs.DetachedJobExecutionResultHeader, nil, false, nil
}

func init() {
	sql.AddPlanHook("ingestion", ingestionPlanHook)
	jobs.RegisterConstructor(
		jobspb.TypeStreamIngestion,
		func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
			return &streamIngestionResumer{
				job: job,
			}
		},
	)
}
