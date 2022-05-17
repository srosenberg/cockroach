package scjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
)

func init() {
	jobs.RegisterConstructor(jobspb.TypeNewSchemaChange, func(
		job *jobs.Job, settings *cluster.Settings,
	) jobs.Resumer {
		return &newSchemaChangeResumer{
			job: job,
		}
	})
}

type newSchemaChangeResumer struct {
	job      *jobs.Job
	rollback bool
}

func (n *newSchemaChangeResumer) Resume(ctx context.Context, execCtxI interface{}) (err error) {
	__antithesis_instrumentation__.Notify(582367)
	return n.run(ctx, execCtxI)
}

func (n *newSchemaChangeResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(582368)
	n.rollback = true
	return n.run(ctx, execCtx)
}

func (n *newSchemaChangeResumer) run(ctx context.Context, execCtxI interface{}) error {
	__antithesis_instrumentation__.Notify(582369)
	execCtx := execCtxI.(sql.JobExecContext)
	execCfg := execCtx.ExecCfg()
	if err := n.job.Update(ctx, nil, func(txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater) error {
		__antithesis_instrumentation__.Notify(582373)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(582374)

		return err
	} else {
		__antithesis_instrumentation__.Notify(582375)
	}
	__antithesis_instrumentation__.Notify(582370)

	if err := execCfg.JobRegistry.CheckPausepoint("newschemachanger.before.exec"); err != nil {
		__antithesis_instrumentation__.Notify(582376)
		return err
	} else {
		__antithesis_instrumentation__.Notify(582377)
	}
	__antithesis_instrumentation__.Notify(582371)
	payload := n.job.Payload()
	deps := scdeps.NewJobRunDependencies(
		execCfg.CollectionFactory,
		execCfg.DB,
		execCfg.InternalExecutor,
		execCfg.IndexBackfiller,
		NewRangeCounter(execCfg.DB, execCfg.DistSQLPlanner),
		func(txn *kv.Txn) scexec.EventLogger {
			__antithesis_instrumentation__.Notify(582378)
			return sql.NewSchemaChangerEventLogger(txn, execCfg, 0)
		},
		execCfg.JobRegistry,
		n.job,
		execCfg.Codec,
		execCfg.Settings,
		execCfg.IndexValidator,
		execCfg.DescMetadaUpdaterFactory,
		execCfg.DeclarativeSchemaChangerTestingKnobs,
		payload.Statement,
		execCtx.SessionData(),
		execCtx.ExtendedEvalContext().Tracing.KVTracingEnabled(),
	)
	__antithesis_instrumentation__.Notify(582372)

	return scrun.RunSchemaChangesInJob(
		ctx,
		execCfg.DeclarativeSchemaChangerTestingKnobs,
		execCfg.Settings,
		deps,
		n.job.ID(),
		payload.DescriptorIDs,
		n.rollback,
	)
}
