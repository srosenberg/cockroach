package gcjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func gcTenant(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	tenID uint64,
	progress *jobspb.SchemaChangeGCProgress,
) error {
	__antithesis_instrumentation__.Notify(492727)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(492733)
		log.Infof(ctx, "GC is being considered for tenant: %d", tenID)
	} else {
		__antithesis_instrumentation__.Notify(492734)
	}
	__antithesis_instrumentation__.Notify(492728)

	if progress.Tenant.Status == jobspb.SchemaChangeGCProgress_WAITING_FOR_GC {
		__antithesis_instrumentation__.Notify(492735)
		return errors.AssertionFailedf(
			"Tenant id %d is expired and should not be in state %+v",
			tenID, jobspb.SchemaChangeGCProgress_WAITING_FOR_GC,
		)
	} else {
		__antithesis_instrumentation__.Notify(492736)
	}
	__antithesis_instrumentation__.Notify(492729)

	info, err := sql.GetTenantRecord(ctx, execCfg, nil, tenID)
	if err != nil {
		__antithesis_instrumentation__.Notify(492737)
		if pgerror.GetPGCode(err) == pgcode.UndefinedObject {
			__antithesis_instrumentation__.Notify(492739)

			if progress.Tenant.Status != jobspb.SchemaChangeGCProgress_DELETED {
				__antithesis_instrumentation__.Notify(492741)

				log.Errorf(ctx, "tenant id %d not found while attempting to GC", tenID)
				progress.Tenant.Status = jobspb.SchemaChangeGCProgress_DELETED
			} else {
				__antithesis_instrumentation__.Notify(492742)
			}
			__antithesis_instrumentation__.Notify(492740)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(492743)
		}
		__antithesis_instrumentation__.Notify(492738)
		return errors.Wrapf(err, "fetching tenant %d", info.ID)
	} else {
		__antithesis_instrumentation__.Notify(492744)
	}
	__antithesis_instrumentation__.Notify(492730)

	if progress.Tenant.Status == jobspb.SchemaChangeGCProgress_DELETED {
		__antithesis_instrumentation__.Notify(492745)
		return errors.AssertionFailedf("GC state for tenant %+v is DELETED yet the tenant row still exists", info)
	} else {
		__antithesis_instrumentation__.Notify(492746)
	}
	__antithesis_instrumentation__.Notify(492731)

	if err := sql.GCTenantSync(ctx, execCfg, info); err != nil {
		__antithesis_instrumentation__.Notify(492747)
		return errors.Wrapf(err, "gc tenant %d", info.ID)
	} else {
		__antithesis_instrumentation__.Notify(492748)
	}
	__antithesis_instrumentation__.Notify(492732)

	progress.Tenant.Status = jobspb.SchemaChangeGCProgress_DELETED
	return nil
}
