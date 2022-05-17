// Package migrationjob contains the jobs.Resumer implementation
// used for long-running migrations.
package migrationjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func init() {
	jobs.RegisterConstructor(jobspb.TypeMigration, func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
		return &resumer{j: job}
	})
}

func NewRecord(
	version clusterversion.ClusterVersion, user security.SQLUsername, name string,
) jobs.Record {
	__antithesis_instrumentation__.Notify(128240)
	return jobs.Record{
		Description: name,
		Details: jobspb.MigrationDetails{
			ClusterVersion: &version,
		},
		Username:      user,
		Progress:      jobspb.MigrationProgress{},
		NonCancelable: true,
	}
}

type resumer struct {
	j *jobs.Job
}

var _ jobs.Resumer = (*resumer)(nil)

func (r resumer) Resume(ctx context.Context, execCtxI interface{}) error {
	__antithesis_instrumentation__.Notify(128241)
	execCtx := execCtxI.(sql.JobExecContext)
	pl := r.j.Payload()
	cv := *pl.GetMigration().ClusterVersion
	ie := execCtx.ExecCfg().InternalExecutor

	alreadyCompleted, err := CheckIfMigrationCompleted(ctx, nil, ie, cv)
	if alreadyCompleted || func() bool {
		__antithesis_instrumentation__.Notify(128247)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(128248)
		return errors.Wrapf(err, "checking migration completion for %v", cv)
	} else {
		__antithesis_instrumentation__.Notify(128249)
	}
	__antithesis_instrumentation__.Notify(128242)
	mc := execCtx.MigrationJobDeps()
	m, ok := mc.GetMigration(cv)
	if !ok {
		__antithesis_instrumentation__.Notify(128250)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(128251)
	}
	__antithesis_instrumentation__.Notify(128243)
	switch m := m.(type) {
	case *migration.SystemMigration:
		__antithesis_instrumentation__.Notify(128252)
		err = m.Run(ctx, cv, mc.SystemDeps(), r.j)
	case *migration.TenantMigration:
		__antithesis_instrumentation__.Notify(128253)
		tenantDeps := migration.TenantDeps{
			DB:                execCtx.ExecCfg().DB,
			Codec:             execCtx.ExecCfg().Codec,
			Settings:          execCtx.ExecCfg().Settings,
			CollectionFactory: execCtx.ExecCfg().CollectionFactory,
			LeaseManager:      execCtx.ExecCfg().LeaseManager,
			InternalExecutor:  execCtx.ExecCfg().InternalExecutor,
			TestingKnobs:      execCtx.ExecCfg().MigrationTestingKnobs,
		}
		tenantDeps.SpanConfig.KVAccessor = execCtx.ExecCfg().SpanConfigKVAccessor
		tenantDeps.SpanConfig.Splitter = execCtx.ExecCfg().SpanConfigSplitter
		tenantDeps.SpanConfig.Default = execCtx.ExecCfg().DefaultZoneConfig.AsSpanConfig()

		err = m.Run(ctx, cv, tenantDeps, r.j)
	default:
		__antithesis_instrumentation__.Notify(128254)
		return errors.AssertionFailedf("unknown migration type %T", m)
	}
	__antithesis_instrumentation__.Notify(128244)
	if err != nil {
		__antithesis_instrumentation__.Notify(128255)
		return errors.Wrapf(err, "running migration for %v", cv)
	} else {
		__antithesis_instrumentation__.Notify(128256)
	}
	__antithesis_instrumentation__.Notify(128245)

	if err := markMigrationCompleted(ctx, ie, cv); err != nil {
		__antithesis_instrumentation__.Notify(128257)
		return errors.Wrapf(err, "marking migration complete for %v", cv)
	} else {
		__antithesis_instrumentation__.Notify(128258)
	}
	__antithesis_instrumentation__.Notify(128246)
	return nil
}

func CheckIfMigrationCompleted(
	ctx context.Context, txn *kv.Txn, ie sqlutil.InternalExecutor, cv clusterversion.ClusterVersion,
) (alreadyCompleted bool, _ error) {
	__antithesis_instrumentation__.Notify(128259)
	row, err := ie.QueryRow(
		ctx,
		"migration-job-find-already-completed",
		txn,
		`
SELECT EXISTS(
        SELECT *
          FROM system.migrations
         WHERE major = $1
           AND minor = $2
           AND patch = $3
           AND internal = $4
       );
`,
		cv.Major,
		cv.Minor,
		cv.Patch,
		cv.Internal)
	if err != nil {
		__antithesis_instrumentation__.Notify(128261)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(128262)
	}
	__antithesis_instrumentation__.Notify(128260)
	return bool(*row[0].(*tree.DBool)), nil
}

func markMigrationCompleted(
	ctx context.Context, ie sqlutil.InternalExecutor, cv clusterversion.ClusterVersion,
) error {
	__antithesis_instrumentation__.Notify(128263)
	_, err := ie.ExecEx(
		ctx,
		"migration-job-mark-job-succeeded",
		nil,
		sessiondata.NodeUserSessionDataOverride,
		`
INSERT
  INTO system.migrations
        (
            major,
            minor,
            patch,
            internal,
            completed_at
        )
VALUES ($1, $2, $3, $4, $5)`,
		cv.Major,
		cv.Minor,
		cv.Patch,
		cv.Internal,
		timeutil.Now())
	return err
}

func (r resumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(128264)
	return nil
}
