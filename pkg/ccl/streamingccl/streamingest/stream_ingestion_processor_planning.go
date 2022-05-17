package streamingest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streamclient"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

func distStreamIngestionPlanSpecs(
	streamAddress streamingccl.StreamAddress,
	topology streamclient.Topology,
	sqlInstanceIDs []base.SQLInstanceID,
	initialHighWater hlc.Timestamp,
	jobID jobspb.JobID,
	streamID streaming.StreamID,
) ([]*execinfrapb.StreamIngestionDataSpec, *execinfrapb.StreamIngestionFrontierSpec, error) {
	__antithesis_instrumentation__.Notify(25584)

	streamIngestionSpecs := make([]*execinfrapb.StreamIngestionDataSpec, 0, len(sqlInstanceIDs))

	trackedSpans := make([]roachpb.Span, 0)
	for i, partition := range topology {
		__antithesis_instrumentation__.Notify(25586)

		if i < len(sqlInstanceIDs) {
			__antithesis_instrumentation__.Notify(25588)
			spec := &execinfrapb.StreamIngestionDataSpec{
				StreamID:           uint64(streamID),
				JobID:              int64(jobID),
				StartTime:          initialHighWater,
				StreamAddress:      string(streamAddress),
				PartitionAddresses: make([]string, 0),
			}
			streamIngestionSpecs = append(streamIngestionSpecs, spec)
		} else {
			__antithesis_instrumentation__.Notify(25589)
		}
		__antithesis_instrumentation__.Notify(25587)
		n := i % len(sqlInstanceIDs)

		streamIngestionSpecs[n].PartitionIds = append(streamIngestionSpecs[n].PartitionIds, partition.ID)
		streamIngestionSpecs[n].PartitionSpecs = append(streamIngestionSpecs[n].PartitionSpecs,
			string(partition.SubscriptionToken))
		streamIngestionSpecs[n].PartitionAddresses = append(streamIngestionSpecs[n].PartitionAddresses,
			string(partition.SrcAddr))

		trackedSpans = append(trackedSpans, roachpb.Span{
			Key:    roachpb.Key(partition.ID),
			EndKey: roachpb.Key(partition.ID).Next(),
		})
	}
	__antithesis_instrumentation__.Notify(25585)

	streamIngestionFrontierSpec := &execinfrapb.StreamIngestionFrontierSpec{
		HighWaterAtStart: initialHighWater,
		TrackedSpans:     trackedSpans,
		JobID:            int64(jobID),
		StreamID:         uint64(streamID),
		StreamAddress:    string(streamAddress),
	}

	return streamIngestionSpecs, streamIngestionFrontierSpec, nil
}

func distStreamIngest(
	ctx context.Context,
	execCtx sql.JobExecContext,
	sqlInstanceIDs []base.SQLInstanceID,
	jobID jobspb.JobID,
	planCtx *sql.PlanningCtx,
	dsp *sql.DistSQLPlanner,
	streamIngestionSpecs []*execinfrapb.StreamIngestionDataSpec,
	streamIngestionFrontierSpec *execinfrapb.StreamIngestionFrontierSpec,
) error {
	__antithesis_instrumentation__.Notify(25590)
	ctx = logtags.AddTag(ctx, "stream-ingest-distsql", nil)
	evalCtx := execCtx.ExtendedEvalContext()
	var noTxn *kv.Txn

	if len(streamIngestionSpecs) == 0 {
		__antithesis_instrumentation__.Notify(25594)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(25595)
	}
	__antithesis_instrumentation__.Notify(25591)

	corePlacement := make([]physicalplan.ProcessorCorePlacement, len(streamIngestionSpecs))
	for i := range streamIngestionSpecs {
		__antithesis_instrumentation__.Notify(25596)
		corePlacement[i].SQLInstanceID = sqlInstanceIDs[i]
		corePlacement[i].Core.StreamIngestionData = streamIngestionSpecs[i]
	}
	__antithesis_instrumentation__.Notify(25592)

	p := planCtx.NewPhysicalPlan()
	p.AddNoInputStage(
		corePlacement,
		execinfrapb.PostProcessSpec{},
		streamIngestionResultTypes,
		execinfrapb.Ordering{},
	)

	execCfg := execCtx.ExecCfg()
	gatewayNodeID, err := execCfg.NodeID.OptionalNodeIDErr(48274)
	if err != nil {
		__antithesis_instrumentation__.Notify(25597)
		return err
	} else {
		__antithesis_instrumentation__.Notify(25598)
	}
	__antithesis_instrumentation__.Notify(25593)

	p.AddSingleGroupStage(base.SQLInstanceID(gatewayNodeID),
		execinfrapb.ProcessorCoreUnion{StreamIngestionFrontier: streamIngestionFrontierSpec},
		execinfrapb.PostProcessSpec{}, streamIngestionResultTypes)

	p.PlanToStreamColMap = []int{0}
	dsp.FinalizePlan(planCtx, p)

	rw := makeStreamIngestionResultWriter(ctx, jobID, execCfg.JobRegistry)

	recv := sql.MakeDistSQLReceiver(
		ctx,
		rw,
		tree.Rows,
		nil,
		noTxn,
		nil,
		evalCtx.Tracing,
		execCfg.ContentionRegistry,
		nil,
	)
	defer recv.Release()

	evalCtxCopy := *evalCtx
	dsp.Run(ctx, planCtx, noTxn, p, recv, &evalCtxCopy, nil)()
	return rw.Err()
}

type streamIngestionResultWriter struct {
	ctx          context.Context
	registry     *jobs.Registry
	jobID        jobspb.JobID
	rowsAffected int
	err          error
}

func makeStreamIngestionResultWriter(
	ctx context.Context, jobID jobspb.JobID, registry *jobs.Registry,
) *streamIngestionResultWriter {
	__antithesis_instrumentation__.Notify(25599)
	return &streamIngestionResultWriter{
		ctx:      ctx,
		registry: registry,
		jobID:    jobID,
	}
}

func (s *streamIngestionResultWriter) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(25600)
	if len(row) == 0 {
		__antithesis_instrumentation__.Notify(25604)
		return errors.New("streamIngestionResultWriter received an empty row")
	} else {
		__antithesis_instrumentation__.Notify(25605)
	}
	__antithesis_instrumentation__.Notify(25601)
	if row[0] == nil {
		__antithesis_instrumentation__.Notify(25606)
		return errors.New("streamIngestionResultWriter expects non-nil row entry")
	} else {
		__antithesis_instrumentation__.Notify(25607)
	}
	__antithesis_instrumentation__.Notify(25602)

	var ingestedHighWatermark hlc.Timestamp
	if err := protoutil.Unmarshal([]byte(*row[0].(*tree.DBytes)),
		&ingestedHighWatermark); err != nil {
		__antithesis_instrumentation__.Notify(25608)
		return errors.NewAssertionErrorWithWrappedErrf(err, `unmarshalling resolved timestamp`)
	} else {
		__antithesis_instrumentation__.Notify(25609)
	}
	__antithesis_instrumentation__.Notify(25603)
	return s.registry.UpdateJobWithTxn(ctx, s.jobID, nil, false,
		func(txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater) error {
			__antithesis_instrumentation__.Notify(25610)
			return jobs.UpdateHighwaterProgressed(ingestedHighWatermark, md, ju)
		})
}

func (s *streamIngestionResultWriter) IncrementRowsAffected(ctx context.Context, n int) {
	__antithesis_instrumentation__.Notify(25611)
	s.rowsAffected += n
}

func (s *streamIngestionResultWriter) SetError(err error) {
	__antithesis_instrumentation__.Notify(25612)
	s.err = err
}

func (s *streamIngestionResultWriter) Err() error {
	__antithesis_instrumentation__.Notify(25613)
	return s.err
}
