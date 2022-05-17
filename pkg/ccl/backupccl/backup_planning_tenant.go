package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

const tenantMetadataQuery = `
SELECT
  tenants.id,                        /* 0 */
  tenants.active,                    /* 1 */
  tenants.info,                      /* 2 */
  tenant_usage.ru_burst_limit,       /* 3 */
  tenant_usage.ru_refill_rate,       /* 4 */
  tenant_usage.ru_current,           /* 5 */
  tenant_usage.total_consumption     /* 6 */
FROM
  system.tenants
  LEFT JOIN system.tenant_usage ON
	  tenants.id = tenant_usage.tenant_id AND tenant_usage.instance_id = 0`

func tenantMetadataFromRow(row tree.Datums) (descpb.TenantInfoWithUsage, error) {
	__antithesis_instrumentation__.Notify(8629)
	if len(row) != 7 {
		__antithesis_instrumentation__.Notify(8634)
		return descpb.TenantInfoWithUsage{}, errors.AssertionFailedf(
			"unexpected row size %d from tenant metadata query", len(row),
		)
	} else {
		__antithesis_instrumentation__.Notify(8635)
	}
	__antithesis_instrumentation__.Notify(8630)

	id := uint64(tree.MustBeDInt(row[0]))
	res := descpb.TenantInfoWithUsage{
		TenantInfo: descpb.TenantInfo{
			ID: id,
		},
	}
	infoBytes := []byte(tree.MustBeDBytes(row[2]))
	if err := protoutil.Unmarshal(infoBytes, &res.TenantInfo); err != nil {
		__antithesis_instrumentation__.Notify(8636)
		return descpb.TenantInfoWithUsage{}, err
	} else {
		__antithesis_instrumentation__.Notify(8637)
	}
	__antithesis_instrumentation__.Notify(8631)

	for _, d := range row[3:5] {
		__antithesis_instrumentation__.Notify(8638)
		if d == tree.DNull {
			__antithesis_instrumentation__.Notify(8639)
			return res, nil
		} else {
			__antithesis_instrumentation__.Notify(8640)
		}
	}
	__antithesis_instrumentation__.Notify(8632)
	res.Usage = &descpb.TenantInfoWithUsage_Usage{
		RUBurstLimit: float64(tree.MustBeDFloat(row[3])),
		RURefillRate: float64(tree.MustBeDFloat(row[4])),
		RUCurrent:    float64(tree.MustBeDFloat(row[5])),
	}
	if row[6] != tree.DNull {
		__antithesis_instrumentation__.Notify(8641)
		consumptionBytes := []byte(tree.MustBeDBytes(row[6]))
		if err := protoutil.Unmarshal(consumptionBytes, &res.Usage.Consumption); err != nil {
			__antithesis_instrumentation__.Notify(8642)
			return descpb.TenantInfoWithUsage{}, err
		} else {
			__antithesis_instrumentation__.Notify(8643)
		}
	} else {
		__antithesis_instrumentation__.Notify(8644)
	}
	__antithesis_instrumentation__.Notify(8633)
	return res, nil
}

func retrieveSingleTenantMetadata(
	ctx context.Context, ie *sql.InternalExecutor, txn *kv.Txn, tenantID roachpb.TenantID,
) (descpb.TenantInfoWithUsage, error) {
	__antithesis_instrumentation__.Notify(8645)
	row, err := ie.QueryRow(
		ctx, "backup-lookup-tenant", txn,
		tenantMetadataQuery+` WHERE id = $1`, tenantID.ToUint64(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(8649)
		return descpb.TenantInfoWithUsage{}, err
	} else {
		__antithesis_instrumentation__.Notify(8650)
	}
	__antithesis_instrumentation__.Notify(8646)
	if row == nil {
		__antithesis_instrumentation__.Notify(8651)
		return descpb.TenantInfoWithUsage{}, errors.Errorf("tenant %s does not exist", tenantID)
	} else {
		__antithesis_instrumentation__.Notify(8652)
	}
	__antithesis_instrumentation__.Notify(8647)
	if !tree.MustBeDBool(row[1]) {
		__antithesis_instrumentation__.Notify(8653)
		return descpb.TenantInfoWithUsage{}, errors.Errorf("tenant %s is not active", tenantID)
	} else {
		__antithesis_instrumentation__.Notify(8654)
	}
	__antithesis_instrumentation__.Notify(8648)

	return tenantMetadataFromRow(row)
}

func retrieveAllTenantsMetadata(
	ctx context.Context, ie *sql.InternalExecutor, txn *kv.Txn,
) ([]descpb.TenantInfoWithUsage, error) {
	__antithesis_instrumentation__.Notify(8655)
	rows, err := ie.QueryBuffered(
		ctx, "backup-lookup-tenants", txn,

		tenantMetadataQuery,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(8658)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(8659)
	}
	__antithesis_instrumentation__.Notify(8656)
	res := make([]descpb.TenantInfoWithUsage, len(rows))
	for i := range rows {
		__antithesis_instrumentation__.Notify(8660)
		res[i], err = tenantMetadataFromRow(rows[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(8661)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(8662)
		}
	}
	__antithesis_instrumentation__.Notify(8657)
	return res, nil
}
