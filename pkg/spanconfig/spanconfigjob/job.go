package spanconfigjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type resumer struct {
	job *jobs.Job
}

var _ jobs.Resumer = (*resumer)(nil)

var reconciliationJobCheckpointInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"spanconfig.reconciliation_job.checkpoint_interval",
	"the frequency at which the span config reconciliation job checkpoints itself",
	5*time.Second,
	settings.NonNegativeDuration,
)

func (r *resumer) Resume(ctx context.Context, execCtxI interface{}) (jobErr error) {
	__antithesis_instrumentation__.Notify(240330)
	defer func() {
		__antithesis_instrumentation__.Notify(240337)

		jobErr = jobs.MarkAsPermanentJobError(jobErr)
	}()
	__antithesis_instrumentation__.Notify(240331)

	execCtx := execCtxI.(sql.JobExecContext)
	if !execCtx.ExecCfg().LogicalClusterID().Equal(r.job.Payload().CreationClusterID) {
		__antithesis_instrumentation__.Notify(240338)

		log.Infof(ctx, "duplicate restored job (source-cluster-id=%s, dest-cluster-id=%s); exiting",
			r.job.Payload().CreationClusterID, execCtx.ExecCfg().LogicalClusterID())
		return nil
	} else {
		__antithesis_instrumentation__.Notify(240339)
	}
	__antithesis_instrumentation__.Notify(240332)

	rc := execCtx.SpanConfigReconciler()
	stopper := execCtx.ExecCfg().DistSQLSrv.Stopper

	r.job.MarkIdle(true)

	ptsRCContext, cancel := stopper.WithCancelOnQuiesce(ctx)
	defer cancel()
	if err := execCtx.ExecCfg().ProtectedTimestampProvider.StartReconciler(
		ptsRCContext, execCtx.ExecCfg().DistSQLSrv.Stopper); err != nil {
		__antithesis_instrumentation__.Notify(240340)
		return errors.Wrap(err, "could not start protected timestamp reconciliation")
	} else {
		__antithesis_instrumentation__.Notify(240341)
	}
	__antithesis_instrumentation__.Notify(240333)

	settingValues := &execCtx.ExecCfg().Settings.SV
	persistCheckpointsMu := struct {
		syncutil.Mutex
		util.EveryN
	}{}
	persistCheckpointsMu.EveryN = util.Every(reconciliationJobCheckpointInterval.Get(settingValues))

	reconciliationJobCheckpointInterval.SetOnChange(settingValues, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(240342)
		persistCheckpointsMu.Lock()
		defer persistCheckpointsMu.Unlock()
		persistCheckpointsMu.EveryN = util.Every(reconciliationJobCheckpointInterval.Get(settingValues))
	})
	__antithesis_instrumentation__.Notify(240334)

	checkpointingDisabled := false
	shouldSkipRetry := false
	var onCheckpointInterceptor func() error

	retryOpts := retry.Options{
		InitialBackoff: 5 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     1.3,
		MaxRetries:     40,
	}

	if knobs := execCtx.ExecCfg().SpanConfigTestingKnobs; knobs != nil {
		__antithesis_instrumentation__.Notify(240343)
		if knobs.JobDisablePersistingCheckpoints {
			__antithesis_instrumentation__.Notify(240345)
			checkpointingDisabled = true
		} else {
			__antithesis_instrumentation__.Notify(240346)
		}
		__antithesis_instrumentation__.Notify(240344)
		shouldSkipRetry = knobs.JobDisableInternalRetry
		onCheckpointInterceptor = knobs.JobOnCheckpointInterceptor
		if knobs.JobOverrideRetryOptions != nil {
			__antithesis_instrumentation__.Notify(240347)
			retryOpts = *knobs.JobOverrideRetryOptions
		} else {
			__antithesis_instrumentation__.Notify(240348)
		}
	} else {
		__antithesis_instrumentation__.Notify(240349)
	}
	__antithesis_instrumentation__.Notify(240335)

	var lastCheckpoint = hlc.Timestamp{}
	const aWhile = 5 * time.Minute
	for retrier := retry.StartWithCtx(ctx, retryOpts); retrier.Next(); {
		__antithesis_instrumentation__.Notify(240350)
		started := timeutil.Now()
		if err := rc.Reconcile(ctx, lastCheckpoint, r.job.Session(), func() error {
			__antithesis_instrumentation__.Notify(240352)
			if onCheckpointInterceptor != nil {
				__antithesis_instrumentation__.Notify(240357)
				if err := onCheckpointInterceptor(); err != nil {
					__antithesis_instrumentation__.Notify(240358)
					return err
				} else {
					__antithesis_instrumentation__.Notify(240359)
				}
			} else {
				__antithesis_instrumentation__.Notify(240360)
			}
			__antithesis_instrumentation__.Notify(240353)

			if checkpointingDisabled {
				__antithesis_instrumentation__.Notify(240361)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(240362)
			}
			__antithesis_instrumentation__.Notify(240354)

			persistCheckpointsMu.Lock()
			shouldPersistCheckpoint := persistCheckpointsMu.ShouldProcess(timeutil.Now())
			persistCheckpointsMu.Unlock()
			if !shouldPersistCheckpoint {
				__antithesis_instrumentation__.Notify(240363)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(240364)
			}
			__antithesis_instrumentation__.Notify(240355)

			if timeutil.Since(started) > aWhile {
				__antithesis_instrumentation__.Notify(240365)
				retrier.Reset()
			} else {
				__antithesis_instrumentation__.Notify(240366)
			}
			__antithesis_instrumentation__.Notify(240356)

			lastCheckpoint = rc.Checkpoint()
			return r.job.SetProgress(ctx, nil, jobspb.AutoSpanConfigReconciliationProgress{
				Checkpoint: rc.Checkpoint(),
			})
		}); err != nil {
			__antithesis_instrumentation__.Notify(240367)
			if shouldSkipRetry {
				__antithesis_instrumentation__.Notify(240371)
				break
			} else {
				__antithesis_instrumentation__.Notify(240372)
			}
			__antithesis_instrumentation__.Notify(240368)
			if errors.Is(err, context.Canceled) {
				__antithesis_instrumentation__.Notify(240373)
				return err
			} else {
				__antithesis_instrumentation__.Notify(240374)
			}
			__antithesis_instrumentation__.Notify(240369)
			if spanconfig.IsMismatchedDescriptorTypesError(err) {
				__antithesis_instrumentation__.Notify(240375)

				log.Warningf(ctx, "observed %v, kicking off full reconciliation...", err)
				lastCheckpoint = hlc.Timestamp{}
				continue
			} else {
				__antithesis_instrumentation__.Notify(240376)
			}
			__antithesis_instrumentation__.Notify(240370)

			log.Errorf(ctx, "reconciler failed with %v, retrying...", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(240377)
		}
		__antithesis_instrumentation__.Notify(240351)
		return nil
	}
	__antithesis_instrumentation__.Notify(240336)

	return errors.Newf("reconciliation unsuccessful, failing job")
}

func (r *resumer) OnFailOrCancel(ctx context.Context, _ interface{}) error {
	__antithesis_instrumentation__.Notify(240378)
	if jobs.HasErrJobCanceled(errors.DecodeError(ctx, *r.job.Payload().FinalResumeError)) {
		__antithesis_instrumentation__.Notify(240380)
		return errors.AssertionFailedf("span config reconciliation job cannot be canceled")
	} else {
		__antithesis_instrumentation__.Notify(240381)
	}
	__antithesis_instrumentation__.Notify(240379)
	return nil
}

func init() {
	jobs.RegisterConstructor(jobspb.TypeAutoSpanConfigReconciliation,
		func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
			return &resumer{job: job}
		})
}
