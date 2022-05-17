// Package migrationmanager provides an implementation of migration.Manager
// for use on kv nodes.
package migrationmanager

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/migration/migrationjob"
	"github.com/cockroachdb/cockroach/pkg/migration/migrations"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
)

type Manager struct {
	deps     migration.SystemDeps
	lm       *lease.Manager
	ie       sqlutil.InternalExecutor
	jr       *jobs.Registry
	codec    keys.SQLCodec
	settings *cluster.Settings
	knobs    migration.TestingKnobs
}

func (m *Manager) GetMigration(key clusterversion.ClusterVersion) (migration.Migration, bool) {
	__antithesis_instrumentation__.Notify(128265)
	if m.knobs.RegistryOverride != nil {
		__antithesis_instrumentation__.Notify(128267)
		if m, ok := m.knobs.RegistryOverride(key); ok {
			__antithesis_instrumentation__.Notify(128268)
			return m, ok
		} else {
			__antithesis_instrumentation__.Notify(128269)
		}
	} else {
		__antithesis_instrumentation__.Notify(128270)
	}
	__antithesis_instrumentation__.Notify(128266)
	return migrations.GetMigration(key)
}

func (m *Manager) SystemDeps() migration.SystemDeps {
	__antithesis_instrumentation__.Notify(128271)
	return m.deps
}

func NewManager(
	deps migration.SystemDeps,
	lm *lease.Manager,
	ie sqlutil.InternalExecutor,
	jr *jobs.Registry,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	testingKnobs *migration.TestingKnobs,
) *Manager {
	__antithesis_instrumentation__.Notify(128272)
	var knobs migration.TestingKnobs
	if testingKnobs != nil {
		__antithesis_instrumentation__.Notify(128274)
		knobs = *testingKnobs
	} else {
		__antithesis_instrumentation__.Notify(128275)
	}
	__antithesis_instrumentation__.Notify(128273)
	return &Manager{
		deps:     deps,
		lm:       lm,
		ie:       ie,
		jr:       jr,
		codec:    codec,
		settings: settings,
		knobs:    knobs,
	}
}

var _ migration.JobDeps = (*Manager)(nil)

func (m *Manager) Migrate(
	ctx context.Context,
	user security.SQLUsername,
	from, to clusterversion.ClusterVersion,
	updateSystemVersionSetting sql.UpdateVersionSystemSettingHook,
) error {
	__antithesis_instrumentation__.Notify(128276)

	ctx = logtags.AddTag(ctx, "migration-mgr", nil)
	if from == to {
		__antithesis_instrumentation__.Notify(128281)

		log.Infof(ctx, "no need to migrate, cluster already at newest version")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128282)
	}
	__antithesis_instrumentation__.Notify(128277)

	clusterVersions := m.listBetween(from, to)
	log.Infof(ctx, "migrating cluster from %s to %s (stepping through %s)", from, to, clusterVersions)
	if len(clusterVersions) == 0 {
		__antithesis_instrumentation__.Notify(128283)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128284)
	}

	{
		__antithesis_instrumentation__.Notify(128285)
		finalVersion := clusterVersions[len(clusterVersions)-1]
		if err := validateTargetClusterVersion(ctx, m.deps.Cluster, finalVersion); err != nil {
			__antithesis_instrumentation__.Notify(128286)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128287)
		}
	}
	__antithesis_instrumentation__.Notify(128278)

	if err := m.checkPreconditions(ctx, clusterVersions); err != nil {
		__antithesis_instrumentation__.Notify(128288)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128289)
	}
	__antithesis_instrumentation__.Notify(128279)

	for _, clusterVersion := range clusterVersions {
		__antithesis_instrumentation__.Notify(128290)
		log.Infof(ctx, "stepping through %s", clusterVersion)

		if err := m.runMigration(ctx, user, clusterVersion); err != nil {
			__antithesis_instrumentation__.Notify(128294)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128295)
		}

		{
			__antithesis_instrumentation__.Notify(128296)

			fenceVersion := migration.FenceVersionFor(ctx, clusterVersion)
			if err := bumpClusterVersion(ctx, m.deps.Cluster, fenceVersion); err != nil {
				__antithesis_instrumentation__.Notify(128297)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128298)
			}
		}
		__antithesis_instrumentation__.Notify(128291)

		if err := validateTargetClusterVersion(ctx, m.deps.Cluster, clusterVersion); err != nil {
			__antithesis_instrumentation__.Notify(128299)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128300)
		}
		__antithesis_instrumentation__.Notify(128292)

		err := bumpClusterVersion(ctx, m.deps.Cluster, clusterVersion)
		if err != nil {
			__antithesis_instrumentation__.Notify(128301)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128302)
		}
		__antithesis_instrumentation__.Notify(128293)

		err = updateSystemVersionSetting(ctx, clusterVersion)
		if err != nil {
			__antithesis_instrumentation__.Notify(128303)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128304)
		}
	}
	__antithesis_instrumentation__.Notify(128280)

	return nil
}

func bumpClusterVersion(
	ctx context.Context, c migration.Cluster, clusterVersion clusterversion.ClusterVersion,
) error {
	__antithesis_instrumentation__.Notify(128305)
	req := &serverpb.BumpClusterVersionRequest{ClusterVersion: &clusterVersion}
	op := fmt.Sprintf("bump-cluster-version=%s", req.ClusterVersion.PrettyPrint())
	return forEveryNodeUntilClusterStable(ctx, op, c, func(
		ctx context.Context, client serverpb.MigrationClient,
	) error {
		__antithesis_instrumentation__.Notify(128306)
		_, err := client.BumpClusterVersion(ctx, req)
		return err
	})
}

func validateTargetClusterVersion(
	ctx context.Context, c migration.Cluster, clusterVersion clusterversion.ClusterVersion,
) error {
	__antithesis_instrumentation__.Notify(128307)
	req := &serverpb.ValidateTargetClusterVersionRequest{ClusterVersion: &clusterVersion}
	op := fmt.Sprintf("validate-cluster-version=%s", req.ClusterVersion.PrettyPrint())
	return forEveryNodeUntilClusterStable(ctx, op, c, func(
		tx context.Context, client serverpb.MigrationClient,
	) error {
		__antithesis_instrumentation__.Notify(128308)
		_, err := client.ValidateTargetClusterVersion(ctx, req)
		return err
	})
}

func forEveryNodeUntilClusterStable(
	ctx context.Context,
	op string,
	c migration.Cluster,
	f func(ctx context.Context, client serverpb.MigrationClient) error,
) error {
	__antithesis_instrumentation__.Notify(128309)
	return c.UntilClusterStable(ctx, func() error {
		__antithesis_instrumentation__.Notify(128310)
		return c.ForEveryNode(ctx, op, f)
	})
}

func (m *Manager) runMigration(
	ctx context.Context, user security.SQLUsername, version clusterversion.ClusterVersion,
) error {
	__antithesis_instrumentation__.Notify(128311)
	mig, exists := m.GetMigration(version)
	if !exists {
		__antithesis_instrumentation__.Notify(128315)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128316)
	}
	__antithesis_instrumentation__.Notify(128312)
	_, isSystemMigration := mig.(*migration.SystemMigration)
	if isSystemMigration && func() bool {
		__antithesis_instrumentation__.Notify(128317)
		return !m.codec.ForSystemTenant() == true
	}() == true {
		__antithesis_instrumentation__.Notify(128318)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128319)
	}
	__antithesis_instrumentation__.Notify(128313)
	alreadyCompleted, id, err := m.getOrCreateMigrationJob(ctx, user, version, mig.Name())
	if alreadyCompleted || func() bool {
		__antithesis_instrumentation__.Notify(128320)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(128321)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128322)
	}
	__antithesis_instrumentation__.Notify(128314)
	return m.jr.Run(ctx, m.ie, []jobspb.JobID{id})
}

func (m *Manager) getOrCreateMigrationJob(
	ctx context.Context,
	user security.SQLUsername,
	version clusterversion.ClusterVersion,
	name string,
) (alreadyCompleted bool, jobID jobspb.JobID, _ error) {
	__antithesis_instrumentation__.Notify(128323)
	newJobID := m.jr.MakeJobID()
	if err := m.deps.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(128325)
		alreadyCompleted, err = migrationjob.CheckIfMigrationCompleted(ctx, txn, m.ie, version)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(128330)
			return ctx.Err() == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(128331)
			log.Warningf(ctx, "failed to check if migration already completed: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(128332)
		}
		__antithesis_instrumentation__.Notify(128326)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(128333)
			return alreadyCompleted == true
		}() == true {
			__antithesis_instrumentation__.Notify(128334)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128335)
		}
		__antithesis_instrumentation__.Notify(128327)
		var found bool
		found, jobID, err = m.getRunningMigrationJob(ctx, txn, version)
		if err != nil {
			__antithesis_instrumentation__.Notify(128336)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128337)
		}
		__antithesis_instrumentation__.Notify(128328)
		if found {
			__antithesis_instrumentation__.Notify(128338)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(128339)
		}
		__antithesis_instrumentation__.Notify(128329)
		jobID = newJobID
		_, err = m.jr.CreateJobWithTxn(ctx, migrationjob.NewRecord(version, user, name), jobID, txn)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(128340)
		return false, 0, err
	} else {
		__antithesis_instrumentation__.Notify(128341)
	}
	__antithesis_instrumentation__.Notify(128324)
	return alreadyCompleted, jobID, nil
}

func (m *Manager) getRunningMigrationJob(
	ctx context.Context, txn *kv.Txn, version clusterversion.ClusterVersion,
) (found bool, jobID jobspb.JobID, _ error) {
	__antithesis_instrumentation__.Notify(128342)
	const query = `
SELECT id, status
	FROM (
		SELECT id,
		status,
		crdb_internal.pb_to_json(
			'cockroach.sql.jobs.jobspb.Payload',
			payload,
      false -- emit_defaults
		) AS pl
	FROM system.jobs
  WHERE status IN ` + jobs.NonTerminalStatusTupleString + `
	)
	WHERE pl->'migration'->'clusterVersion' = $1::JSON;`
	jsonMsg, err := protoreflect.MessageToJSON(&version, protoreflect.FmtFlags{EmitDefaults: false})
	if err != nil {
		__antithesis_instrumentation__.Notify(128346)
		return false, 0, errors.Wrap(err, "failed to marshal version to JSON")
	} else {
		__antithesis_instrumentation__.Notify(128347)
	}
	__antithesis_instrumentation__.Notify(128343)
	rows, err := m.ie.QueryBuffered(ctx, "migration-manager-find-jobs", txn, query, jsonMsg.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(128348)
		return false, 0, err
	} else {
		__antithesis_instrumentation__.Notify(128349)
	}
	__antithesis_instrumentation__.Notify(128344)
	parseRow := func(row tree.Datums) (id jobspb.JobID, status jobs.Status) {
		__antithesis_instrumentation__.Notify(128350)
		return jobspb.JobID(*row[0].(*tree.DInt)), jobs.Status(*row[1].(*tree.DString))
	}
	__antithesis_instrumentation__.Notify(128345)
	switch len(rows) {
	case 0:
		__antithesis_instrumentation__.Notify(128351)
		return false, 0, nil
	case 1:
		__antithesis_instrumentation__.Notify(128352)
		id, status := parseRow(rows[0])
		log.Infof(ctx, "found existing migration job %d for version %v in status %s, waiting",
			id, &version, status)
		return true, id, nil
	default:
		__antithesis_instrumentation__.Notify(128353)
		var buf redact.StringBuilder
		buf.Printf("found multiple non-terminal jobs for version %s: [", redact.Safe(&version))
		for i, row := range rows {
			__antithesis_instrumentation__.Notify(128355)
			if i > 0 {
				__antithesis_instrumentation__.Notify(128357)
				buf.SafeString(", ")
			} else {
				__antithesis_instrumentation__.Notify(128358)
			}
			__antithesis_instrumentation__.Notify(128356)
			id, status := parseRow(row)
			buf.Printf("(%d, %s)", id, redact.Safe(status))
		}
		__antithesis_instrumentation__.Notify(128354)
		log.Errorf(ctx, "%s", buf)
		return false, 0, errors.AssertionFailedf("%s", buf)
	}
}

func (m *Manager) listBetween(
	from clusterversion.ClusterVersion, to clusterversion.ClusterVersion,
) []clusterversion.ClusterVersion {
	__antithesis_instrumentation__.Notify(128359)
	if m.knobs.ListBetweenOverride != nil {
		__antithesis_instrumentation__.Notify(128361)
		return m.knobs.ListBetweenOverride(from, to)
	} else {
		__antithesis_instrumentation__.Notify(128362)
	}
	__antithesis_instrumentation__.Notify(128360)
	return clusterversion.ListBetween(from, to)
}

func (m *Manager) checkPreconditions(
	ctx context.Context, versions []clusterversion.ClusterVersion,
) error {
	__antithesis_instrumentation__.Notify(128363)
	for _, v := range versions {
		__antithesis_instrumentation__.Notify(128365)
		mig, ok := m.GetMigration(v)
		if !ok {
			__antithesis_instrumentation__.Notify(128368)
			continue
		} else {
			__antithesis_instrumentation__.Notify(128369)
		}
		__antithesis_instrumentation__.Notify(128366)
		tm, ok := mig.(*migration.TenantMigration)
		if !ok {
			__antithesis_instrumentation__.Notify(128370)
			continue
		} else {
			__antithesis_instrumentation__.Notify(128371)
		}
		__antithesis_instrumentation__.Notify(128367)
		if err := tm.Precondition(ctx, v, migration.TenantDeps{
			DB:               m.deps.DB,
			Codec:            m.codec,
			Settings:         m.settings,
			LeaseManager:     m.lm,
			InternalExecutor: m.ie,
		}); err != nil {
			__antithesis_instrumentation__.Notify(128372)
			return errors.Wrapf(
				err,
				"verifying precondition for version %s",
				redact.SafeString(v.PrettyPrint()),
			)
		} else {
			__antithesis_instrumentation__.Notify(128373)
		}
	}
	__antithesis_instrumentation__.Notify(128364)
	return nil
}
