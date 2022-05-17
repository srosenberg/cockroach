package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
)

func tenantSettingsTableMigration(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128730)

	if !d.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(128732)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128733)
	}
	__antithesis_instrumentation__.Notify(128731)
	return createSystemTable(
		ctx, d.DB, d.Codec, systemschema.TenantSettingsTable,
	)
}
