package streamingest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streamclient"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type streamIngestionResumer struct {
	job *jobs.Job
}

func ingest(
	ctx context.Context,
	execCtx sql.JobExecContext,
	streamAddress streamingccl.StreamAddress,
	tenantID roachpb.TenantID,
	startTime hlc.Timestamp,
	progress jobspb.Progress,
	jobID jobspb.JobID,
) error {
	__antithesis_instrumentation__.Notify(25297)

	client, err := streamclient.NewStreamClient(streamAddress)
	if err != nil {
		__antithesis_instrumentation__.Notify(25300)
		return err
	} else {
		__antithesis_instrumentation__.Notify(25301)
	}
	__antithesis_instrumentation__.Notify(25298)
	ingestWithClient := func() error {
		__antithesis_instrumentation__.Notify(25302)

		streamID, err := client.Create(ctx, tenantID)
		if err != nil {
			__antithesis_instrumentation__.Notify(25308)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25309)
		}
		__antithesis_instrumentation__.Notify(25303)

		topology, err := client.Plan(ctx, streamID)
		if err != nil {
			__antithesis_instrumentation__.Notify(25310)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25311)
		}
		__antithesis_instrumentation__.Notify(25304)

		initialHighWater := startTime
		if h := progress.GetHighWater(); h != nil && func() bool {
			__antithesis_instrumentation__.Notify(25312)
			return !h.IsEmpty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(25313)
			initialHighWater = *h
		} else {
			__antithesis_instrumentation__.Notify(25314)
		}
		__antithesis_instrumentation__.Notify(25305)

		evalCtx := execCtx.ExtendedEvalContext()
		dsp := execCtx.DistSQLPlanner()

		planCtx, sqlInstanceIDs, err := dsp.SetupAllNodesPlanning(ctx, evalCtx, execCtx.ExecCfg())
		if err != nil {
			__antithesis_instrumentation__.Notify(25315)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25316)
		}
		__antithesis_instrumentation__.Notify(25306)

		streamIngestionSpecs, streamIngestionFrontierSpec, err := distStreamIngestionPlanSpecs(
			streamAddress, topology, sqlInstanceIDs, initialHighWater, jobID, streamID)
		if err != nil {
			__antithesis_instrumentation__.Notify(25317)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25318)
		}
		__antithesis_instrumentation__.Notify(25307)

		return distStreamIngest(ctx, execCtx, sqlInstanceIDs, jobID, planCtx, dsp, streamIngestionSpecs,
			streamIngestionFrontierSpec)
	}
	__antithesis_instrumentation__.Notify(25299)
	return errors.CombineErrors(ingestWithClient(), client.Close())
}

func (s *streamIngestionResumer) Resume(resumeCtx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(25319)
	details := s.job.Details().(jobspb.StreamIngestionDetails)
	p := execCtx.(sql.JobExecContext)

	streamAddress := streamingccl.StreamAddress(details.StreamAddress)
	err := ingest(resumeCtx, p, streamAddress, details.TenantID, details.StartTime, s.job.Progress(), s.job.ID())
	if err != nil {
		__antithesis_instrumentation__.Notify(25321)
		return err
	} else {
		__antithesis_instrumentation__.Notify(25322)
	}
	__antithesis_instrumentation__.Notify(25320)

	return s.revertToCutoverTimestamp(resumeCtx, execCtx)
}

func (s *streamIngestionResumer) revertToCutoverTimestamp(
	ctx context.Context, execCtx interface{},
) error {
	__antithesis_instrumentation__.Notify(25323)
	p := execCtx.(sql.JobExecContext)
	db := p.ExecCfg().DB
	j, err := p.ExecCfg().JobRegistry.LoadJob(ctx, s.job.ID())
	if err != nil {
		__antithesis_instrumentation__.Notify(25329)
		return err
	} else {
		__antithesis_instrumentation__.Notify(25330)
	}
	__antithesis_instrumentation__.Notify(25324)
	details := j.Details()
	var sd jobspb.StreamIngestionDetails
	var ok bool
	if sd, ok = details.(jobspb.StreamIngestionDetails); !ok {
		__antithesis_instrumentation__.Notify(25331)
		return errors.Newf("unknown details type %T in stream ingestion job %d",
			details, s.job.ID())
	} else {
		__antithesis_instrumentation__.Notify(25332)
	}
	__antithesis_instrumentation__.Notify(25325)
	progress := j.Progress()
	var sp *jobspb.Progress_StreamIngest
	if sp, ok = progress.GetDetails().(*jobspb.Progress_StreamIngest); !ok {
		__antithesis_instrumentation__.Notify(25333)
		return errors.Newf("unknown progress type %T in stream ingestion job %d",
			j.Progress().Progress, s.job.ID())
	} else {
		__antithesis_instrumentation__.Notify(25334)
	}
	__antithesis_instrumentation__.Notify(25326)

	if sp.StreamIngest.CutoverTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(25335)
		return errors.AssertionFailedf("cutover time is unexpectedly empty, " +
			"cannot revert to a consistent state")
	} else {
		__antithesis_instrumentation__.Notify(25336)
	}
	__antithesis_instrumentation__.Notify(25327)

	spans := []roachpb.Span{sd.Span}
	for len(spans) != 0 {
		__antithesis_instrumentation__.Notify(25337)
		var b kv.Batch
		for _, span := range spans {
			__antithesis_instrumentation__.Notify(25340)
			b.AddRawRequest(&roachpb.RevertRangeRequest{
				RequestHeader: roachpb.RequestHeader{
					Key:    span.Key,
					EndKey: span.EndKey,
				},
				TargetTime:                          sp.StreamIngest.CutoverTime,
				EnableTimeBoundIteratorOptimization: true,
			})
		}
		__antithesis_instrumentation__.Notify(25338)
		b.Header.MaxSpanRequestKeys = sql.RevertTableDefaultBatchSize
		if err := db.Run(ctx, &b); err != nil {
			__antithesis_instrumentation__.Notify(25341)
			return err
		} else {
			__antithesis_instrumentation__.Notify(25342)
		}
		__antithesis_instrumentation__.Notify(25339)

		spans = spans[:0]
		for _, raw := range b.RawResponse().Responses {
			__antithesis_instrumentation__.Notify(25343)
			r := raw.GetRevertRange()
			if r.ResumeSpan != nil {
				__antithesis_instrumentation__.Notify(25344)
				if !r.ResumeSpan.Valid() {
					__antithesis_instrumentation__.Notify(25346)
					return errors.Errorf("invalid resume span: %s", r.ResumeSpan)
				} else {
					__antithesis_instrumentation__.Notify(25347)
				}
				__antithesis_instrumentation__.Notify(25345)
				spans = append(spans, *r.ResumeSpan)
			} else {
				__antithesis_instrumentation__.Notify(25348)
			}
		}
	}
	__antithesis_instrumentation__.Notify(25328)

	return nil
}

func (s *streamIngestionResumer) OnFailOrCancel(_ context.Context, _ interface{}) error {
	__antithesis_instrumentation__.Notify(25349)
	return nil
}

var _ jobs.Resumer = &streamIngestionResumer{}

func init() {
	jobs.RegisterConstructor(
		jobspb.TypeStreamIngestion,
		func(job *jobs.Job,
			settings *cluster.Settings) jobs.Resumer {
			return &streamIngestionResumer{job: job}
		})
}
