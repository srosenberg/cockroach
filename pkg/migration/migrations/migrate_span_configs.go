package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

func ensureSpanConfigReconciliation(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, j *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128468)
	if !d.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(128471)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128472)
	}
	__antithesis_instrumentation__.Notify(128469)

	retryOpts := retry.Options{
		InitialBackoff: time.Second,
		MaxBackoff:     2 * time.Second,
		Multiplier:     2,
		MaxRetries:     50,
	}
	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(128473)
		row, err := d.InternalExecutor.QueryRowEx(ctx, "get-spanconfig-progress", nil,
			sessiondata.NodeUserSessionDataOverride, `
SELECT progress
  FROM system.jobs
 WHERE id = (SELECT job_id FROM [SHOW AUTOMATIC JOBS] WHERE job_type = 'AUTO SPAN CONFIG RECONCILIATION')
`)
		if err != nil {
			__antithesis_instrumentation__.Notify(128479)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128480)
		}
		__antithesis_instrumentation__.Notify(128474)
		if row == nil {
			__antithesis_instrumentation__.Notify(128481)
			log.Info(ctx, "reconciliation job not found, retrying...")
			continue
		} else {
			__antithesis_instrumentation__.Notify(128482)
		}
		__antithesis_instrumentation__.Notify(128475)
		progress, err := jobs.UnmarshalProgress(row[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(128483)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128484)
		}
		__antithesis_instrumentation__.Notify(128476)
		sp, ok := progress.GetDetails().(*jobspb.Progress_AutoSpanConfigReconciliation)
		if !ok {
			__antithesis_instrumentation__.Notify(128485)
			log.Fatal(ctx, "unexpected job progress type")
		} else {
			__antithesis_instrumentation__.Notify(128486)
		}
		__antithesis_instrumentation__.Notify(128477)
		if sp.AutoSpanConfigReconciliation.Checkpoint.IsEmpty() {
			__antithesis_instrumentation__.Notify(128487)
			log.Info(ctx, "waiting for span config reconciliation...")
			continue
		} else {
			__antithesis_instrumentation__.Notify(128488)
		}
		__antithesis_instrumentation__.Notify(128478)

		return nil
	}
	__antithesis_instrumentation__.Notify(128470)

	return errors.Newf("unable to reconcile span configs")
}

func ensureSpanConfigSubscription(
	ctx context.Context, _ clusterversion.ClusterVersion, deps migration.SystemDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128489)
	return deps.Cluster.UntilClusterStable(ctx, func() error {
		__antithesis_instrumentation__.Notify(128490)
		return deps.Cluster.ForEveryNode(ctx, "ensure-span-config-subscription",
			func(ctx context.Context, client serverpb.MigrationClient) error {
				__antithesis_instrumentation__.Notify(128491)
				req := &serverpb.WaitForSpanConfigSubscriptionRequest{}
				_, err := client.WaitForSpanConfigSubscription(ctx, req)
				return err
			})
	})
}
