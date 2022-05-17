package streamproducer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

func makeTenantSpan(tenantID uint64) *roachpb.Span {
	__antithesis_instrumentation__.Notify(26953)
	prefix := keys.MakeTenantPrefix(roachpb.MakeTenantID(tenantID))
	return &roachpb.Span{Key: prefix, EndKey: prefix.PrefixEnd()}
}

func makeProducerJobRecord(
	registry *jobs.Registry,
	tenantID uint64,
	timeout time.Duration,
	username security.SQLUsername,
	ptsID uuid.UUID,
) jobs.Record {
	__antithesis_instrumentation__.Notify(26954)
	return jobs.Record{
		JobID:       registry.MakeJobID(),
		Description: fmt.Sprintf("stream replication for tenant %d", tenantID),
		Username:    username,
		Details: jobspb.StreamReplicationDetails{
			ProtectedTimestampRecord: &ptsID,
			Spans:                    []*roachpb.Span{makeTenantSpan(tenantID)},
		},
		Progress: jobspb.StreamReplicationProgress{
			Expiration: timeutil.Now().Add(timeout),
		},
	}
}

type producerJobResumer struct {
	job *jobs.Job

	timeSource timeutil.TimeSource
	timer      timeutil.TimerI
}

func (p *producerJobResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(26955)
	jobExec := execCtx.(sql.JobExecContext)
	execCfg := jobExec.ExecCfg()
	isTimedOut := func(job *jobs.Job) bool {
		__antithesis_instrumentation__.Notify(26958)
		progress := p.job.Progress()
		return progress.GetStreamReplication().Expiration.Before(p.timeSource.Now())
	}
	__antithesis_instrumentation__.Notify(26956)
	trackFrequency := streamingccl.StreamReplicationStreamLivenessTrackFrequency.Get(execCfg.SV())
	if isTimedOut(p.job) {
		__antithesis_instrumentation__.Notify(26959)
		return errors.Errorf("replication stream %d timed out", p.job.ID())
	} else {
		__antithesis_instrumentation__.Notify(26960)
	}
	__antithesis_instrumentation__.Notify(26957)
	p.timer.Reset(trackFrequency)
	for {
		__antithesis_instrumentation__.Notify(26961)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(26962)
			return ctx.Err()
		case <-p.timer.Ch():
			__antithesis_instrumentation__.Notify(26963)
			p.timer.MarkRead()
			p.timer.Reset(trackFrequency)
			j, err := execCfg.JobRegistry.LoadJob(ctx, p.job.ID())
			if err != nil {
				__antithesis_instrumentation__.Notify(26965)
				return err
			} else {
				__antithesis_instrumentation__.Notify(26966)
			}
			__antithesis_instrumentation__.Notify(26964)
			if isTimedOut(j) {
				__antithesis_instrumentation__.Notify(26967)
				return errors.Errorf("replication stream %d timed out", p.job.ID())
			} else {
				__antithesis_instrumentation__.Notify(26968)
			}
		}
	}
}

func (p *producerJobResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(26969)
	jobExec := execCtx.(sql.JobExecContext)
	execCfg := jobExec.ExecCfg()

	ptr := p.job.Details().(jobspb.StreamReplicationDetails).ProtectedTimestampRecord
	return execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(26970)
		err := execCfg.ProtectedTimestampProvider.Release(ctx, txn, *ptr)

		if errors.Is(err, exec.ErrNotFound) {
			__antithesis_instrumentation__.Notify(26972)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(26973)
		}
		__antithesis_instrumentation__.Notify(26971)
		return err
	})
}

func init() {
	jobs.RegisterConstructor(
		jobspb.TypeStreamReplication,
		func(job *jobs.Job, _ *cluster.Settings) jobs.Resumer {
			ts := timeutil.DefaultTimeSource{}
			return &producerJobResumer{
				job:        job,
				timeSource: ts,
				timer:      ts.NewTimer(),
			}
		},
	)
}
