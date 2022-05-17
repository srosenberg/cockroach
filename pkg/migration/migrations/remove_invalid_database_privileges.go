package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/migration"
)

func runRemoveInvalidDatabasePrivileges(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128580)
	return runPostDeserializationChangesOnAllDescriptors(ctx, d)
}
