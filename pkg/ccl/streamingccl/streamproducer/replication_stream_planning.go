package streamproducer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeeddist"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type replicationStreamEval struct {
	*tree.ReplicationStream
	sinkURI func() (string, error)
}

const createStreamOp = "CREATE REPLICATION STREAM"

func makeReplicationStreamEval(
	ctx context.Context, p sql.PlanHookState, stream *tree.ReplicationStream,
) (*replicationStreamEval, error) {
	__antithesis_instrumentation__.Notify(26989)
	if err := utilccl.CheckEnterpriseEnabled(
		p.ExecCfg().Settings, p.ExecCfg().LogicalClusterID(),
		p.ExecCfg().Organization(), createStreamOp); err != nil {
		__antithesis_instrumentation__.Notify(26992)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(26993)
	}
	__antithesis_instrumentation__.Notify(26990)

	eval := &replicationStreamEval{ReplicationStream: stream}
	if eval.SinkURI == nil {
		__antithesis_instrumentation__.Notify(26994)
		eval.sinkURI = func() (string, error) { __antithesis_instrumentation__.Notify(26995); return "", nil }
	} else {
		__antithesis_instrumentation__.Notify(26996)
		var err error
		eval.sinkURI, err = p.TypeAsString(ctx, stream.SinkURI, createStreamOp)
		if err != nil {
			__antithesis_instrumentation__.Notify(26997)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(26998)
		}
	}
	__antithesis_instrumentation__.Notify(26991)

	return eval, nil
}

func telemetrySinkName(sink string) string {
	__antithesis_instrumentation__.Notify(26999)

	return "sinkless"
}

func streamKVs(
	ctx context.Context,
	p sql.PlanHookState,
	startTS hlc.Timestamp,
	spans []roachpb.Span,
	resultsCh chan<- tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(27000)

	statementTime := startTS
	if statementTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(27003)
		statementTime = hlc.Timestamp{
			WallTime: p.ExtendedEvalContext().GetStmtTimestamp().UnixNano(),
		}
	} else {
		__antithesis_instrumentation__.Notify(27004)
	}
	__antithesis_instrumentation__.Notify(27001)

	cfOpts := map[string]string{
		changefeedbase.OptSchemaChangePolicy: string(changefeedbase.OptSchemaChangePolicyIgnore),
		changefeedbase.OptFormat:             string(changefeedbase.OptFormatNative),
		changefeedbase.OptResolvedTimestamps: changefeedbase.OptEmitAllResolvedTimestamps,
	}

	details := jobspb.ChangefeedDetails{
		Tables:        nil,
		Opts:          cfOpts,
		SinkURI:       "",
		StatementTime: statementTime,
	}

	telemetry.Count(`replication.create.sink.` + telemetrySinkName(details.SinkURI))
	telemetry.Count(`replication.create.ok`)
	var checkpoint jobspb.ChangefeedProgress_Checkpoint
	if err := changefeeddist.StartDistChangefeed(
		ctx, p, 0, details, spans, startTS, checkpoint, resultsCh, changefeeddist.TestingKnobs{},
	); err != nil {
		__antithesis_instrumentation__.Notify(27005)
		telemetry.Count("replication.done.fail")
		return err
	} else {
		__antithesis_instrumentation__.Notify(27006)
	}
	__antithesis_instrumentation__.Notify(27002)
	telemetry.Count(`replication.done.ok`)
	return nil
}

func doCreateReplicationStream(
	ctx context.Context,
	p sql.PlanHookState,
	eval *replicationStreamEval,
	resultsCh chan<- tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(27007)
	if err := p.RequireAdminRole(ctx, createStreamOp); err != nil {
		__antithesis_instrumentation__.Notify(27014)
		return pgerror.Newf(pgcode.InsufficientPrivilege, "only the admin can backup other tenants")
	} else {
		__antithesis_instrumentation__.Notify(27015)
	}
	__antithesis_instrumentation__.Notify(27008)

	if !p.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(27016)
		return pgerror.Newf(pgcode.InsufficientPrivilege, "only the system tenant can backup other tenants")
	} else {
		__antithesis_instrumentation__.Notify(27017)
	}
	__antithesis_instrumentation__.Notify(27009)

	sinkURI, err := eval.sinkURI()
	if err != nil {
		__antithesis_instrumentation__.Notify(27018)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27019)
	}
	__antithesis_instrumentation__.Notify(27010)

	if sinkURI != "" {
		__antithesis_instrumentation__.Notify(27020)

		return pgerror.New(pgcode.FeatureNotSupported, "replication streaming into sink not supported")
	} else {
		__antithesis_instrumentation__.Notify(27021)
	}
	__antithesis_instrumentation__.Notify(27011)

	var scanStart hlc.Timestamp
	if eval.Options.Cursor != nil {
		__antithesis_instrumentation__.Notify(27022)
		asOf, err := p.EvalAsOfTimestamp(ctx, tree.AsOfClause{Expr: eval.Options.Cursor})
		if err != nil {
			__antithesis_instrumentation__.Notify(27024)
			return err
		} else {
			__antithesis_instrumentation__.Notify(27025)
		}
		__antithesis_instrumentation__.Notify(27023)
		scanStart = asOf.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(27026)
	}
	__antithesis_instrumentation__.Notify(27012)

	var spans []roachpb.Span
	if !eval.Targets.TenantID.IsSet() {
		__antithesis_instrumentation__.Notify(27027)

		return pgerror.New(pgcode.FeatureNotSupported, "granular replication streaming not supported")
	} else {
		__antithesis_instrumentation__.Notify(27028)
	}
	__antithesis_instrumentation__.Notify(27013)

	telemetry.Count(`replication.create.tenant`)
	prefix := keys.MakeTenantPrefix(roachpb.MakeTenantID(eval.Targets.TenantID.ToUint64()))
	spans = append(spans, roachpb.Span{
		Key:    prefix,
		EndKey: prefix.PrefixEnd(),
	})

	return streamKVs(ctx, p, scanStart, spans, resultsCh)
}

var replicationStreamHeader = colinfo.ResultColumns{
	{Name: "_", Typ: types.String},
	{Name: "key", Typ: types.Bytes},
	{Name: "value", Typ: types.Bytes},
}

func createReplicationStreamHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(27029)
	stream, ok := stmt.(*tree.ReplicationStream)
	if !ok {
		__antithesis_instrumentation__.Notify(27034)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(27035)
	}
	__antithesis_instrumentation__.Notify(27030)
	if !p.SessionData().EnableStreamReplication {
		__antithesis_instrumentation__.Notify(27036)
		return nil, nil, nil, false, errors.WithTelemetry(
			pgerror.WithCandidateCode(
				errors.WithHint(
					errors.Newf("stream replication is only supported experimentally"),
					"You can enable stream replication by running `SET enable_experimental_stream_replication = true`.",
				),
				pgcode.FeatureNotSupported,
			),
			"replication.create.disabled",
		)
	} else {
		__antithesis_instrumentation__.Notify(27037)
	}
	__antithesis_instrumentation__.Notify(27031)

	eval, err := makeReplicationStreamEval(ctx, p, stream)
	if err != nil {
		__antithesis_instrumentation__.Notify(27038)
		return nil, nil, nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(27039)
	}
	__antithesis_instrumentation__.Notify(27032)

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(27040)
		err := doCreateReplicationStream(ctx, p, eval, resultsCh)
		if err != nil {
			__antithesis_instrumentation__.Notify(27042)
			telemetry.Count("replication.create.failed")
			return err
		} else {
			__antithesis_instrumentation__.Notify(27043)
		}
		__antithesis_instrumentation__.Notify(27041)

		return nil
	}
	__antithesis_instrumentation__.Notify(27033)
	avoidBuffering := stream.SinkURI == nil
	return fn, replicationStreamHeader, nil, avoidBuffering, nil
}

func getReplicationStreamSpec(
	evalCtx *tree.EvalContext, txn *kv.Txn, streamID streaming.StreamID,
) (*streampb.ReplicationStreamSpec, error) {
	__antithesis_instrumentation__.Notify(27044)
	jobExecCtx := evalCtx.JobExecContext.(sql.JobExecContext)

	j, err := jobExecCtx.ExecCfg().JobRegistry.LoadJob(evalCtx.Ctx(), jobspb.JobID(streamID))
	if err != nil {
		__antithesis_instrumentation__.Notify(27050)
		return nil, errors.Wrapf(err, "Replication stream %d has error", streamID)
	} else {
		__antithesis_instrumentation__.Notify(27051)
	}
	__antithesis_instrumentation__.Notify(27045)
	if j.Status() != jobs.StatusRunning {
		__antithesis_instrumentation__.Notify(27052)
		return nil, errors.Errorf("Replication stream %d is not running", streamID)
	} else {
		__antithesis_instrumentation__.Notify(27053)
	}
	__antithesis_instrumentation__.Notify(27046)

	var noTxn *kv.Txn
	dsp := jobExecCtx.DistSQLPlanner()
	planCtx := dsp.NewPlanningCtx(evalCtx.Ctx(), jobExecCtx.ExtendedEvalContext(),
		nil, noTxn, sql.DistributionTypeSystemTenantOnly)

	replicatedSpans := j.Details().(jobspb.StreamReplicationDetails).Spans
	spans := make([]roachpb.Span, 0, len(replicatedSpans))
	for _, span := range replicatedSpans {
		__antithesis_instrumentation__.Notify(27054)
		spans = append(spans, *span)
	}
	__antithesis_instrumentation__.Notify(27047)
	spanPartitions, err := dsp.PartitionSpans(evalCtx.Ctx(), planCtx, spans)
	if err != nil {
		__antithesis_instrumentation__.Notify(27055)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27056)
	}
	__antithesis_instrumentation__.Notify(27048)

	res := &streampb.ReplicationStreamSpec{
		Partitions: make([]streampb.ReplicationStreamSpec_Partition, 0, len(spanPartitions)),
	}
	for _, sp := range spanPartitions {
		__antithesis_instrumentation__.Notify(27057)
		nodeInfo, err := dsp.GetSQLInstanceInfo(sp.SQLInstanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(27059)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(27060)
		}
		__antithesis_instrumentation__.Notify(27058)
		res.Partitions = append(res.Partitions, streampb.ReplicationStreamSpec_Partition{
			NodeID:     roachpb.NodeID(sp.SQLInstanceID),
			SQLAddress: nodeInfo.SQLAddress,
			Locality:   nodeInfo.Locality,
			PartitionSpec: &streampb.StreamPartitionSpec{
				Spans: sp.Spans,
				Config: streampb.StreamPartitionSpec_ExecutionConfig{
					MinCheckpointFrequency: streamingccl.StreamReplicationMinCheckpointFrequency.Get(&evalCtx.Settings.SV),
				},
			},
		})
	}
	__antithesis_instrumentation__.Notify(27049)
	return res, nil
}

func init() {
	sql.AddPlanHook("replication stream", createReplicationStreamHook)
}
