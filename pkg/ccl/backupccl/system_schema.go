package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	fmt "fmt"
	"math"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	descpb "github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type clusterBackupInclusion int

const (
	invalidBackupInclusion = iota

	optInToClusterBackup

	optOutOfClusterBackup
)

type systemBackupConfiguration struct {
	shouldIncludeInClusterBackup clusterBackupInclusion

	restoreBeforeData bool

	restoreInOrder int32

	migrationFunc func(ctx context.Context, execCtx *sql.ExecutorConfig, txn *kv.Txn, tempTableName string) error

	customRestoreFunc func(ctx context.Context, execCtx *sql.ExecutorConfig, txn *kv.Txn, systemTableName, tempTableName string) error
}

func defaultSystemTableRestoreFunc(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	txn *kv.Txn,
	systemTableName, tempTableName string,
) error {
	__antithesis_instrumentation__.Notify(13229)
	executor := execCfg.InternalExecutor

	deleteQuery := fmt.Sprintf("DELETE FROM system.%s WHERE true", systemTableName)
	opName := systemTableName + "-data-deletion"
	log.Eventf(ctx, "clearing data from system table %s with query %q",
		systemTableName, deleteQuery)

	_, err := executor.Exec(ctx, opName, txn, deleteQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(13232)
		return errors.Wrapf(err, "deleting data from system.%s", systemTableName)
	} else {
		__antithesis_instrumentation__.Notify(13233)
	}
	__antithesis_instrumentation__.Notify(13230)

	restoreQuery := fmt.Sprintf("INSERT INTO system.%s (SELECT * FROM %s);",
		systemTableName, tempTableName)
	opName = systemTableName + "-data-insert"
	if _, err := executor.Exec(ctx, opName, txn, restoreQuery); err != nil {
		__antithesis_instrumentation__.Notify(13234)
		return errors.Wrapf(err, "inserting data to system.%s", systemTableName)
	} else {
		__antithesis_instrumentation__.Notify(13235)
	}
	__antithesis_instrumentation__.Notify(13231)
	return nil
}

func tenantSettingsTableRestoreFunc(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	txn *kv.Txn,
	systemTableName, tempTableName string,
) error {
	__antithesis_instrumentation__.Notify(13236)
	if execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(13239)
		return defaultSystemTableRestoreFunc(ctx, execCfg, txn, systemTableName, tempTableName)
	} else {
		__antithesis_instrumentation__.Notify(13240)
	}
	__antithesis_instrumentation__.Notify(13237)

	if count, err := queryTableRowCount(ctx, execCfg.InternalExecutor, txn, tempTableName); err == nil && func() bool {
		__antithesis_instrumentation__.Notify(13241)
		return count > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(13242)
		log.Warningf(ctx, "skipping restore of %d entries in system.tenant_settings table", count)
	} else {
		__antithesis_instrumentation__.Notify(13243)
		if err != nil {
			__antithesis_instrumentation__.Notify(13244)
			log.Warningf(ctx, "skipping restore of entries in system.tenant_settings table (count failed: %s)", err.Error())
		} else {
			__antithesis_instrumentation__.Notify(13245)
		}
	}
	__antithesis_instrumentation__.Notify(13238)
	return nil
}

func queryTableRowCount(
	ctx context.Context, ie *sql.InternalExecutor, txn *kv.Txn, tableName string,
) (int64, error) {
	__antithesis_instrumentation__.Notify(13246)
	countQuery := fmt.Sprintf("SELECT count(1) FROM %s", tableName)
	row, err := ie.QueryRow(ctx, fmt.Sprintf("count-%s", tableName), txn, countQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(13249)
		return 0, errors.Wrapf(err, "counting rows in %q", tableName)
	} else {
		__antithesis_instrumentation__.Notify(13250)
	}
	__antithesis_instrumentation__.Notify(13247)

	count, ok := row[0].(*tree.DInt)
	if !ok {
		__antithesis_instrumentation__.Notify(13251)
		return 0, errors.AssertionFailedf("failed to read count as DInt (was %T)", row[0])
	} else {
		__antithesis_instrumentation__.Notify(13252)
	}
	__antithesis_instrumentation__.Notify(13248)
	return int64(*count), nil
}

func jobsMigrationFunc(
	ctx context.Context, execCfg *sql.ExecutorConfig, txn *kv.Txn, tempTableName string,
) (err error) {
	__antithesis_instrumentation__.Notify(13253)
	executor := execCfg.InternalExecutor

	const statesToRevert = `('` + string(jobs.StatusRunning) + `', ` +
		`'` + string(jobs.StatusPauseRequested) + `', ` +
		`'` + string(jobs.StatusPaused) + `')`

	jobsToRevert := make([]int64, 0)
	query := `SELECT id, payload FROM ` + tempTableName + ` WHERE status IN ` + statesToRevert
	it, err := executor.QueryIteratorEx(
		ctx, "restore-fetching-job-payloads", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		query)
	if err != nil {
		__antithesis_instrumentation__.Notify(13259)
		return errors.Wrap(err, "fetching job payloads")
	} else {
		__antithesis_instrumentation__.Notify(13260)
	}
	__antithesis_instrumentation__.Notify(13254)
	defer func() {
		__antithesis_instrumentation__.Notify(13261)
		closeErr := it.Close()
		if err == nil {
			__antithesis_instrumentation__.Notify(13262)
			err = closeErr
		} else {
			__antithesis_instrumentation__.Notify(13263)
		}
	}()
	__antithesis_instrumentation__.Notify(13255)
	for {
		__antithesis_instrumentation__.Notify(13264)
		ok, err := it.Next(ctx)
		if !ok {
			__antithesis_instrumentation__.Notify(13268)
			if err != nil {
				__antithesis_instrumentation__.Notify(13270)
				return err
			} else {
				__antithesis_instrumentation__.Notify(13271)
			}
			__antithesis_instrumentation__.Notify(13269)
			break
		} else {
			__antithesis_instrumentation__.Notify(13272)
		}
		__antithesis_instrumentation__.Notify(13265)

		r := it.Cur()
		id, payloadBytes := r[0], r[1]
		rawJobID, ok := id.(*tree.DInt)
		if !ok {
			__antithesis_instrumentation__.Notify(13273)
			return errors.Errorf("job: failed to read job id as DInt (was %T)", id)
		} else {
			__antithesis_instrumentation__.Notify(13274)
		}
		__antithesis_instrumentation__.Notify(13266)
		jobID := int64(*rawJobID)

		payload, err := jobs.UnmarshalPayload(payloadBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(13275)
			return errors.Wrap(err, "failed to unmarshal job to restore")
		} else {
			__antithesis_instrumentation__.Notify(13276)
		}
		__antithesis_instrumentation__.Notify(13267)
		if payload.Type() == jobspb.TypeImport || func() bool {
			__antithesis_instrumentation__.Notify(13277)
			return payload.Type() == jobspb.TypeRestore == true
		}() == true {
			__antithesis_instrumentation__.Notify(13278)
			jobsToRevert = append(jobsToRevert, jobID)
		} else {
			__antithesis_instrumentation__.Notify(13279)
		}
	}
	__antithesis_instrumentation__.Notify(13256)

	var updateStatusQuery strings.Builder
	fmt.Fprintf(&updateStatusQuery, "UPDATE %s SET status = $1 WHERE id IN ", tempTableName)
	fmt.Fprint(&updateStatusQuery, "(")
	for i, job := range jobsToRevert {
		__antithesis_instrumentation__.Notify(13280)
		if i > 0 {
			__antithesis_instrumentation__.Notify(13282)
			fmt.Fprint(&updateStatusQuery, ", ")
		} else {
			__antithesis_instrumentation__.Notify(13283)
		}
		__antithesis_instrumentation__.Notify(13281)
		fmt.Fprintf(&updateStatusQuery, "'%d'", job)
	}
	__antithesis_instrumentation__.Notify(13257)
	fmt.Fprint(&updateStatusQuery, ")")

	if _, err := executor.Exec(ctx, "updating-job-status", txn, updateStatusQuery.String(), jobs.StatusCancelRequested); err != nil {
		__antithesis_instrumentation__.Notify(13284)
		return errors.Wrap(err, "updating restored jobs as reverting")
	} else {
		__antithesis_instrumentation__.Notify(13285)
	}
	__antithesis_instrumentation__.Notify(13258)

	return nil
}

func jobsRestoreFunc(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	txn *kv.Txn,
	systemTableName, tempTableName string,
) error {
	__antithesis_instrumentation__.Notify(13286)
	executor := execCfg.InternalExecutor

	restoreQuery := fmt.Sprintf("INSERT INTO system.%s (SELECT * FROM %s) ON CONFLICT DO NOTHING;",
		systemTableName, tempTableName)
	opName := systemTableName + "-data-insert"
	if _, err := executor.Exec(ctx, opName, txn, restoreQuery); err != nil {
		__antithesis_instrumentation__.Notify(13288)
		return errors.Wrapf(err, "inserting data to system.%s", systemTableName)
	} else {
		__antithesis_instrumentation__.Notify(13289)
	}
	__antithesis_instrumentation__.Notify(13287)
	return nil
}

func settingsRestoreFunc(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	txn *kv.Txn,
	systemTableName, tempTableName string,
) error {
	__antithesis_instrumentation__.Notify(13290)
	executor := execCfg.InternalExecutor

	deleteQuery := fmt.Sprintf("DELETE FROM system.%s WHERE name <> 'version'", systemTableName)
	opName := systemTableName + "-data-deletion"
	log.Eventf(ctx, "clearing data from system table %s with query %q",
		systemTableName, deleteQuery)

	_, err := executor.Exec(ctx, opName, txn, deleteQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(13293)
		return errors.Wrapf(err, "deleting data from system.%s", systemTableName)
	} else {
		__antithesis_instrumentation__.Notify(13294)
	}
	__antithesis_instrumentation__.Notify(13291)

	restoreQuery := fmt.Sprintf("INSERT INTO system.%s (SELECT * FROM %s WHERE name <> 'version');",
		systemTableName, tempTableName)
	opName = systemTableName + "-data-insert"
	if _, err := executor.Exec(ctx, opName, txn, restoreQuery); err != nil {
		__antithesis_instrumentation__.Notify(13295)
		return errors.Wrapf(err, "inserting data to system.%s", systemTableName)
	} else {
		__antithesis_instrumentation__.Notify(13296)
	}
	__antithesis_instrumentation__.Notify(13292)
	return nil
}

var systemTableBackupConfiguration = map[string]systemBackupConfiguration{
	systemschema.UsersTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
	},
	systemschema.ZonesTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,

		restoreBeforeData: true,
	},
	systemschema.SettingsTable.GetName(): {

		restoreInOrder:               math.MaxInt32,
		shouldIncludeInClusterBackup: optInToClusterBackup,
		customRestoreFunc:            settingsRestoreFunc,
	},
	systemschema.LocationsTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
	},
	systemschema.RoleMembersTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
	},
	systemschema.RoleOptionsTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
	},
	systemschema.UITable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
	},
	systemschema.CommentsTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
	},
	systemschema.JobsTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
		migrationFunc:                jobsMigrationFunc,
		customRestoreFunc:            jobsRestoreFunc,
	},
	systemschema.ScheduledJobsTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
	},
	systemschema.TableStatisticsTable.GetName(): {

		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.DescriptorTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.EventLogTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.LeaseTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.NamespaceTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.ProtectedTimestampsMetaTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.ProtectedTimestampsRecordsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.RangeEventTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.ReplicationConstraintStatsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.ReplicationCriticalLocalitiesTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.ReportsMetaTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.ReplicationStatsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.SqllivenessTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.StatementBundleChunksTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.StatementDiagnosticsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.StatementDiagnosticsRequestsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.TenantsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.WebSessionsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.MigrationsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.JoinTokensTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.StatementStatisticsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.TransactionStatisticsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.DatabaseRoleSettingsTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
	},
	systemschema.TenantUsageTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.SQLInstancesTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.SpanConfigurationsTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
	systemschema.TenantSettingsTable.GetName(): {
		shouldIncludeInClusterBackup: optInToClusterBackup,
		customRestoreFunc:            tenantSettingsTableRestoreFunc,
	},
	systemschema.SpanCountTable.GetName(): {
		shouldIncludeInClusterBackup: optOutOfClusterBackup,
	},
}

func GetSystemTablesToIncludeInClusterBackup() map[string]struct{} {
	__antithesis_instrumentation__.Notify(13297)
	systemTablesToInclude := make(map[string]struct{})
	for systemTableName, backupConfig := range systemTableBackupConfiguration {
		__antithesis_instrumentation__.Notify(13299)
		if backupConfig.shouldIncludeInClusterBackup == optInToClusterBackup {
			__antithesis_instrumentation__.Notify(13300)
			systemTablesToInclude[systemTableName] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(13301)
		}
	}
	__antithesis_instrumentation__.Notify(13298)

	return systemTablesToInclude
}

func GetSystemTableIDsToExcludeFromClusterBackup(
	ctx context.Context, execCfg *sql.ExecutorConfig,
) (map[descpb.ID]struct{}, error) {
	__antithesis_instrumentation__.Notify(13302)
	systemTableIDsToExclude := make(map[descpb.ID]struct{})
	for systemTableName, backupConfig := range systemTableBackupConfiguration {
		__antithesis_instrumentation__.Notify(13304)
		if backupConfig.shouldIncludeInClusterBackup == optOutOfClusterBackup {
			__antithesis_instrumentation__.Notify(13305)
			err := sql.DescsTxn(ctx, execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error {
				__antithesis_instrumentation__.Notify(13307)
				tn := tree.MakeTableNameWithSchema("system", tree.PublicSchemaName, tree.Name(systemTableName))
				found, desc, err := col.GetImmutableTableByName(ctx, txn, &tn, tree.ObjectLookupFlags{})
				isNotFoundErr := errors.Is(err, catalog.ErrDescriptorNotFound)
				if err != nil && func() bool {
					__antithesis_instrumentation__.Notify(13310)
					return !isNotFoundErr == true
				}() == true {
					__antithesis_instrumentation__.Notify(13311)
					return err
				} else {
					__antithesis_instrumentation__.Notify(13312)
				}
				__antithesis_instrumentation__.Notify(13308)

				if !found || func() bool {
					__antithesis_instrumentation__.Notify(13313)
					return isNotFoundErr == true
				}() == true {
					__antithesis_instrumentation__.Notify(13314)
					log.Warningf(ctx, "could not find system table descriptor %q", systemTableName)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(13315)
				}
				__antithesis_instrumentation__.Notify(13309)
				systemTableIDsToExclude[desc.GetID()] = struct{}{}
				return nil
			})
			__antithesis_instrumentation__.Notify(13306)
			if err != nil {
				__antithesis_instrumentation__.Notify(13316)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(13317)
			}
		} else {
			__antithesis_instrumentation__.Notify(13318)
		}
	}
	__antithesis_instrumentation__.Notify(13303)

	return systemTableIDsToExclude, nil
}

func getSystemTablesToRestoreBeforeData() map[string]struct{} {
	__antithesis_instrumentation__.Notify(13319)
	systemTablesToRestoreBeforeData := make(map[string]struct{})
	for systemTableName, backupConfig := range systemTableBackupConfiguration {
		__antithesis_instrumentation__.Notify(13321)
		if backupConfig.shouldIncludeInClusterBackup == optInToClusterBackup && func() bool {
			__antithesis_instrumentation__.Notify(13322)
			return backupConfig.restoreBeforeData == true
		}() == true {
			__antithesis_instrumentation__.Notify(13323)
			systemTablesToRestoreBeforeData[systemTableName] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(13324)
		}
	}
	__antithesis_instrumentation__.Notify(13320)

	return systemTablesToRestoreBeforeData
}
