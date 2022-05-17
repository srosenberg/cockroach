package catprivilege

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
)

var (
	readSystemTables = []catconstants.SystemTableName{
		catconstants.NamespaceTableName,
		catconstants.DescriptorTableName,
		catconstants.DescIDSequenceTableName,
		catconstants.TenantsTableName,
		catconstants.ProtectedTimestampsMetaTableName,
		catconstants.ProtectedTimestampsRecordsTableName,
		catconstants.StatementStatisticsTableName,
		catconstants.TransactionStatisticsTableName,

		catconstants.PreMigrationNamespaceTableName,
	}

	readWriteSystemTables = []catconstants.SystemTableName{
		catconstants.UsersTableName,
		catconstants.ZonesTableName,
		catconstants.SettingsTableName,
		catconstants.LeaseTableName,
		catconstants.EventLogTableName,
		catconstants.RangeEventTableName,
		catconstants.UITableName,
		catconstants.JobsTableName,
		catconstants.WebSessionsTableName,
		catconstants.TableStatisticsTableName,
		catconstants.LocationsTableName,
		catconstants.RoleMembersTableName,
		catconstants.CommentsTableName,
		catconstants.ReportsMetaTableName,
		catconstants.ReplicationConstraintStatsTableName,
		catconstants.ReplicationCriticalLocalitiesTableName,
		catconstants.ReplicationStatsTableName,
		catconstants.RoleOptionsTableName,
		catconstants.StatementBundleChunksTableName,
		catconstants.StatementDiagnosticsRequestsTableName,
		catconstants.StatementDiagnosticsTableName,
		catconstants.ScheduledJobsTableName,
		catconstants.SqllivenessTableName,
		catconstants.MigrationsTableName,
		catconstants.JoinTokensTableName,
		catconstants.DatabaseRoleSettingsTableName,
		catconstants.TenantUsageTableName,
		catconstants.SQLInstancesTableName,
		catconstants.SpanConfigurationsTableName,
		catconstants.TenantSettingsTableName,
		catconstants.SpanCountTableName,
	}

	systemSuperuserPrivileges = func() map[descpb.NameInfo]privilege.List {
		__antithesis_instrumentation__.Notify(250878)
		m := make(map[descpb.NameInfo]privilege.List)
		tableKey := descpb.NameInfo{
			ParentID:       keys.SystemDatabaseID,
			ParentSchemaID: keys.SystemPublicSchemaID,
		}
		for _, rw := range readWriteSystemTables {
			__antithesis_instrumentation__.Notify(250881)
			tableKey.Name = string(rw)
			m[tableKey] = privilege.ReadWriteData
		}
		__antithesis_instrumentation__.Notify(250879)
		for _, r := range readSystemTables {
			__antithesis_instrumentation__.Notify(250882)
			tableKey.Name = string(r)
			m[tableKey] = privilege.ReadData
		}
		__antithesis_instrumentation__.Notify(250880)
		m[descpb.NameInfo{Name: catconstants.SystemDatabaseName}] = privilege.List{privilege.CONNECT}
		return m
	}()
)

func SystemSuperuserPrivileges(nameKey catalog.NameKey) privilege.List {
	__antithesis_instrumentation__.Notify(250883)
	key := descpb.NameInfo{
		ParentID:       nameKey.GetParentID(),
		ParentSchemaID: nameKey.GetParentSchemaID(),
		Name:           nameKey.GetName(),
	}
	return systemSuperuserPrivileges[key]
}
