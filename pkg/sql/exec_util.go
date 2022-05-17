package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvtenant"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/contention"
	"github.com/cockroachdb/cockroach/pkg/sql/distsql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/gcjob/gcjobnotifier"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/optionalnodeliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirecancel"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/querycache"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/scheduledlogging"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/stmtdiagnostics"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/collector"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var ClusterOrganization = settings.RegisterStringSetting(
	settings.TenantWritable,
	"cluster.organization",
	"organization name",
	"",
).WithPublic()

func ClusterIsInternal(sv *settings.Values) bool {
	__antithesis_instrumentation__.Notify(470335)
	return strings.Contains(ClusterOrganization.Get(sv), "Cockroach Labs")
}

var ClusterSecret = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(470336)
	s := settings.RegisterStringSetting(
		settings.TenantWritable,
		"cluster.secret",
		"cluster specific secret",
		"",
	)

	s.SetReportable(false)
	return s
}()

var defaultIntSize = func() *settings.IntSetting {
	__antithesis_instrumentation__.Notify(470337)
	s := settings.RegisterIntSetting(
		settings.TenantWritable,
		"sql.defaults.default_int_size",
		"the size, in bytes, of an INT type", 8, func(i int64) error {
			__antithesis_instrumentation__.Notify(470339)
			if i != 4 && func() bool {
				__antithesis_instrumentation__.Notify(470341)
				return i != 8 == true
			}() == true {
				__antithesis_instrumentation__.Notify(470342)
				return errors.New("only 4 or 8 are valid values")
			} else {
				__antithesis_instrumentation__.Notify(470343)
			}
			__antithesis_instrumentation__.Notify(470340)
			return nil
		}).WithPublic()
	__antithesis_instrumentation__.Notify(470338)
	s.SetVisibility(settings.Public)
	return s
}()

const allowCrossDatabaseFKsSetting = "sql.cross_db_fks.enabled"

var allowCrossDatabaseFKs = settings.RegisterBoolSetting(
	settings.TenantWritable,
	allowCrossDatabaseFKsSetting,
	"if true, creating foreign key references across databases is allowed",
	false,
).WithPublic()

const allowCrossDatabaseViewsSetting = "sql.cross_db_views.enabled"

var allowCrossDatabaseViews = settings.RegisterBoolSetting(
	settings.TenantWritable,
	allowCrossDatabaseViewsSetting,
	"if true, creating views that refer to other databases is allowed",
	false,
).WithPublic()

const allowCrossDatabaseSeqOwnerSetting = "sql.cross_db_sequence_owners.enabled"

var allowCrossDatabaseSeqOwner = settings.RegisterBoolSetting(
	settings.TenantWritable,
	allowCrossDatabaseSeqOwnerSetting,
	"if true, creating sequences owned by tables from other databases is allowed",
	false,
).WithPublic()

const allowCrossDatabaseSeqReferencesSetting = "sql.cross_db_sequence_references.enabled"

var allowCrossDatabaseSeqReferences = settings.RegisterBoolSetting(
	settings.TenantWritable,
	allowCrossDatabaseSeqReferencesSetting,
	"if true, sequences referenced by tables from other databases are allowed",
	false,
).WithPublic()

const SecondaryTenantsZoneConfigsEnabledSettingName = "sql.zone_configs.allow_for_secondary_tenant.enabled"

var secondaryTenantZoneConfigsEnabled = settings.RegisterBoolSetting(
	settings.TenantReadOnly,
	SecondaryTenantsZoneConfigsEnabledSettingName,
	"allow secondary tenants to set zone configurations; does not affect the system tenant",
	false,
)

var traceTxnThreshold = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.trace.txn.enable_threshold",
	"duration beyond which all transactions are traced (set to 0 to "+
		"disable). This setting is coarser grained than"+
		"sql.trace.stmt.enable_threshold because it applies to all statements "+
		"within a transaction as well as client communication (e.g. retries).", 0,
).WithPublic()

var TraceStmtThreshold = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.trace.stmt.enable_threshold",
	"duration beyond which all statements are traced (set to 0 to disable). "+
		"This applies to individual statements within a transaction and is therefore "+
		"finer-grained than sql.trace.txn.enable_threshold.",
	0,
).WithPublic()

var traceSessionEventLogEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.trace.session_eventlog.enabled",
	"set to true to enable session tracing. "+
		"Note that enabling this may have a non-trivial negative performance impact.",
	false,
).WithPublic()

const ReorderJoinsLimitClusterSettingName = "sql.defaults.reorder_joins_limit"

var ReorderJoinsLimitClusterValue = settings.RegisterIntSetting(
	settings.TenantWritable,
	ReorderJoinsLimitClusterSettingName,
	"default number of joins to reorder",
	opt.DefaultJoinOrderLimit,
	func(limit int64) error {
		__antithesis_instrumentation__.Notify(470344)
		if limit < 0 || func() bool {
			__antithesis_instrumentation__.Notify(470346)
			return limit > opt.MaxReorderJoinsLimit == true
		}() == true {
			__antithesis_instrumentation__.Notify(470347)
			return pgerror.Newf(pgcode.InvalidParameterValue,
				"cannot set %s to a value less than 0 or greater than %v",
				ReorderJoinsLimitClusterSettingName,
				opt.MaxReorderJoinsLimit,
			)
		} else {
			__antithesis_instrumentation__.Notify(470348)
		}
		__antithesis_instrumentation__.Notify(470345)
		return nil
	},
).WithPublic()

var requireExplicitPrimaryKeysClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.require_explicit_primary_keys.enabled",
	"default value for requiring explicit primary keys in CREATE TABLE statements",
	false,
).WithPublic()

var placementEnabledClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.multiregion_placement_policy.enabled",
	"default value for enable_multiregion_placement_policy;"+
		" allows for use of PLACEMENT RESTRICTED",
	false,
)

var autoRehomingEnabledClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.experimental_auto_rehoming.enabled",
	"default value for experimental_enable_auto_rehoming;"+
		" allows for rows in REGIONAL BY ROW tables to be auto-rehomed on UPDATE",
	false,
).WithPublic()

var onUpdateRehomeRowEnabledClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.on_update_rehome_row.enabled",
	"default value for on_update_rehome_row;"+
		" enables ON UPDATE rehome_row() expressions to trigger on updates",
	true,
).WithPublic()

var temporaryTablesEnabledClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.experimental_temporary_tables.enabled",
	"default value for experimental_enable_temp_tables; allows for use of temporary tables by default",
	false,
).WithPublic()

var implicitColumnPartitioningEnabledClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.experimental_implicit_column_partitioning.enabled",
	"default value for experimental_enable_temp_tables; allows for the use of implicit column partitioning",
	false,
).WithPublic()

var overrideMultiRegionZoneConfigClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.override_multi_region_zone_config.enabled",
	"default value for override_multi_region_zone_config; "+
		"allows for overriding the zone configs of a multi-region table or database",
	false,
).WithPublic()

var maxHashShardedIndexRangePreSplit = settings.RegisterIntSetting(
	settings.SystemOnly,
	"sql.hash_sharded_range_pre_split.max",
	"max pre-split ranges to have when adding hash sharded index to an existing table",
	16,
	settings.PositiveInt,
).WithPublic()

var zigzagJoinClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.zigzag_join.enabled",
	"default value for enable_zigzag_join session setting; allows use of zig-zag join by default",
	true,
).WithPublic()

var optDrivenFKCascadesClusterLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.defaults.foreign_key_cascades_limit",
	"default value for foreign_key_cascades_limit session setting; limits the number of cascading operations that run as part of a single query",
	10000,
	settings.NonNegativeInt,
).WithPublic()

var preferLookupJoinsForFKs = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.prefer_lookup_joins_for_fks.enabled",
	"default value for prefer_lookup_joins_for_fks session setting; causes foreign key operations to use lookup joins when possible",
	false,
).WithPublic()

var optUseHistogramsClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.optimizer_use_histograms.enabled",
	"default value for optimizer_use_histograms session setting; enables usage of histograms in the optimizer by default",
	true,
).WithPublic()

var optUseMultiColStatsClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.optimizer_use_multicol_stats.enabled",
	"default value for optimizer_use_multicol_stats session setting; enables usage of multi-column stats in the optimizer by default",
	true,
).WithPublic()

var localityOptimizedSearchMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.locality_optimized_partitioned_index_scan.enabled",
	"default value for locality_optimized_partitioned_index_scan session setting; "+
		"enables searching for rows in the current region before searching remote regions",
	true,
).WithPublic()

var implicitSelectForUpdateClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.implicit_select_for_update.enabled",
	"default value for enable_implicit_select_for_update session setting; enables FOR UPDATE locking during the row-fetch phase of mutation statements",
	true,
).WithPublic()

var insertFastPathClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.insert_fast_path.enabled",
	"default value for enable_insert_fast_path session setting; enables a specialized insert path",
	true,
).WithPublic()

var experimentalAlterColumnTypeGeneralMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.experimental_alter_column_type.enabled",
	"default value for experimental_alter_column_type session setting; "+
		"enables the use of ALTER COLUMN TYPE for general conversions",
	false,
).WithPublic()

var clusterStatementTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.defaults.statement_timeout",
	"default value for the statement_timeout; "+
		"default value for the statement_timeout session setting; controls the "+
		"duration a query is permitted to run before it is canceled; if set to 0, "+
		"there is no timeout",
	0,
	settings.NonNegativeDuration,
).WithPublic()

var clusterLockTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.defaults.lock_timeout",
	"default value for the lock_timeout; "+
		"default value for the lock_timeout session setting; controls the "+
		"duration a query is permitted to wait while attempting to acquire "+
		"a lock on a key or while blocking on an existing lock in order to "+
		"perform a non-locking read on a key; if set to 0, there is no timeout",
	0,
	settings.NonNegativeDuration,
).WithPublic()

var clusterIdleInSessionTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.defaults.idle_in_session_timeout",
	"default value for the idle_in_session_timeout; "+
		"default value for the idle_in_session_timeout session setting; controls the "+
		"duration a session is permitted to idle before the session is terminated; "+
		"if set to 0, there is no timeout",
	0,
	settings.NonNegativeDuration,
).WithPublic()

var clusterIdleInTransactionSessionTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.defaults.idle_in_transaction_session_timeout",
	"default value for the idle_in_transaction_session_timeout; controls the "+
		"duration a session is permitted to idle in a transaction before the "+
		"session is terminated; if set to 0, there is no timeout",
	0,
	settings.NonNegativeDuration,
).WithPublic()

var experimentalUniqueWithoutIndexConstraintsMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.experimental_enable_unique_without_index_constraints.enabled",
	"default value for experimental_enable_unique_without_index_constraints session setting;"+
		"disables unique without index constraints by default",
	false,
).WithPublic()

var experimentalUseNewSchemaChanger = settings.RegisterEnumSetting(
	settings.TenantWritable,
	"sql.defaults.use_declarative_schema_changer",
	"default value for use_declarative_schema_changer session setting;"+
		"disables new schema changer by default",
	"on",
	map[int64]string{
		int64(sessiondatapb.UseNewSchemaChangerOff):          "off",
		int64(sessiondatapb.UseNewSchemaChangerOn):           "on",
		int64(sessiondatapb.UseNewSchemaChangerUnsafe):       "unsafe",
		int64(sessiondatapb.UseNewSchemaChangerUnsafeAlways): "unsafe_always",
	},
).WithPublic()

var experimentalStreamReplicationEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.experimental_stream_replication.enabled",
	"default value for experimental_stream_replication session setting;"+
		"enables the ability to setup a replication stream",
	false,
).WithPublic()

var stubCatalogTablesEnabledClusterValue = settings.RegisterBoolSetting(
	settings.TenantWritable,
	`sql.defaults.stub_catalog_tables.enabled`,
	`default value for stub_catalog_tables session setting`,
	true,
).WithPublic()

var experimentalComputedColumnRewrites = settings.RegisterValidatedStringSetting(
	settings.TenantWritable,
	"sql.defaults.experimental_computed_column_rewrites",
	"allows rewriting computed column expressions in CREATE TABLE and ALTER TABLE statements; "+
		"the format is: '(before expression) -> (after expression) [, (before expression) -> (after expression) ...]'",
	"",
	func(_ *settings.Values, val string) error {
		__antithesis_instrumentation__.Notify(470349)
		_, err := schemaexpr.ParseComputedColumnRewrites(val)
		return err
	},
)

var propagateInputOrdering = settings.RegisterBoolSetting(
	settings.TenantWritable,
	`sql.defaults.propagate_input_ordering.enabled`,
	`default value for the experimental propagate_input_ordering session variable`,
	false,
)

var settingWorkMemBytes = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"sql.distsql.temp_storage.workmem",
	"maximum amount of memory in bytes a processor can use before falling back to temp storage",
	execinfra.DefaultMemoryLimit,
).WithPublic()

const ExperimentalDistSQLPlanningClusterSettingName = "sql.defaults.experimental_distsql_planning"

var experimentalDistSQLPlanningClusterMode = settings.RegisterEnumSetting(
	settings.TenantWritable,
	ExperimentalDistSQLPlanningClusterSettingName,
	"default experimental_distsql_planning mode; enables experimental opt-driven DistSQL planning",
	"off",
	map[int64]string{
		int64(sessiondatapb.ExperimentalDistSQLPlanningOff): "off",
		int64(sessiondatapb.ExperimentalDistSQLPlanningOn):  "on",
	},
).WithPublic()

const VectorizeClusterSettingName = "sql.defaults.vectorize"

var VectorizeClusterMode = settings.RegisterEnumSetting(
	settings.TenantWritable,
	VectorizeClusterSettingName,
	"default vectorize mode",
	"on",
	map[int64]string{
		int64(sessiondatapb.VectorizeUnset):              "on",
		int64(sessiondatapb.VectorizeOn):                 "on",
		int64(sessiondatapb.VectorizeExperimentalAlways): "experimental_always",
		int64(sessiondatapb.VectorizeOff):                "off",
	},
).WithPublic()

var DistSQLClusterExecMode = settings.RegisterEnumSetting(
	settings.TenantWritable,
	"sql.defaults.distsql",
	"default distributed SQL execution mode",
	"auto",
	map[int64]string{
		int64(sessiondatapb.DistSQLOff):    "off",
		int64(sessiondatapb.DistSQLAuto):   "auto",
		int64(sessiondatapb.DistSQLOn):     "on",
		int64(sessiondatapb.DistSQLAlways): "always",
	},
).WithPublic()

var SerialNormalizationMode = settings.RegisterEnumSetting(
	settings.TenantWritable,
	"sql.defaults.serial_normalization",
	"default handling of SERIAL in table definitions",
	"rowid",
	map[int64]string{
		int64(sessiondatapb.SerialUsesRowID):              "rowid",
		int64(sessiondatapb.SerialUsesUnorderedRowID):     "unordered_rowid",
		int64(sessiondatapb.SerialUsesVirtualSequences):   "virtual_sequence",
		int64(sessiondatapb.SerialUsesSQLSequences):       "sql_sequence",
		int64(sessiondatapb.SerialUsesCachedSQLSequences): "sql_sequence_cached",
	},
).WithPublic()

var disallowFullTableScans = settings.RegisterBoolSetting(
	settings.TenantWritable,
	`sql.defaults.disallow_full_table_scans.enabled`,
	"setting to true rejects queries that have planned a full table scan",
	false,
).WithPublic()

var intervalStyle = settings.RegisterEnumSetting(
	settings.TenantWritable,
	"sql.defaults.intervalstyle",
	"default value for IntervalStyle session setting",
	strings.ToLower(duration.IntervalStyle_POSTGRES.String()),
	func() map[int64]string {
		__antithesis_instrumentation__.Notify(470350)
		ret := make(map[int64]string, len(duration.IntervalStyle_name))
		for k, v := range duration.IntervalStyle_name {
			__antithesis_instrumentation__.Notify(470352)
			ret[int64(k)] = strings.ToLower(v)
		}
		__antithesis_instrumentation__.Notify(470351)
		return ret
	}(),
).WithPublic()

var dateStyleEnumMap = map[int64]string{
	0: "ISO, MDY",
	1: "ISO, DMY",
	2: "ISO, YMD",
}

var dateStyle = settings.RegisterEnumSetting(
	settings.TenantWritable,
	"sql.defaults.datestyle",
	"default value for DateStyle session setting",
	pgdate.DefaultDateStyle().SQLString(),
	dateStyleEnumMap,
).WithPublic()

const intervalStyleEnabledClusterSetting = "sql.defaults.intervalstyle.enabled"

var intervalStyleEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	intervalStyleEnabledClusterSetting,
	"default value for intervalstyle_enabled session setting",
	false,
)

const dateStyleEnabledClusterSetting = "sql.defaults.datestyle.enabled"

var dateStyleEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	dateStyleEnabledClusterSetting,
	"default value for datestyle_enabled session setting",
	false,
)

var txnRowsWrittenLog = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.defaults.transaction_rows_written_log",
	"the threshold for the number of rows written by a SQL transaction "+
		"which - once exceeded - will trigger a logging event to SQL_PERF (or "+
		"SQL_INTERNAL_PERF for internal transactions); use 0 to disable",
	0,
	settings.NonNegativeInt,
).WithPublic()

var txnRowsWrittenErr = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.defaults.transaction_rows_written_err",
	"the limit for the number of rows written by a SQL transaction which - "+
		"once exceeded - will fail the transaction (or will trigger a logging "+
		"event to SQL_INTERNAL_PERF for internal transactions); use 0 to disable",
	0,
	settings.NonNegativeInt,
).WithPublic()

var txnRowsReadLog = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.defaults.transaction_rows_read_log",
	"the threshold for the number of rows read by a SQL transaction "+
		"which - once exceeded - will trigger a logging event to SQL_PERF (or "+
		"SQL_INTERNAL_PERF for internal transactions); use 0 to disable",
	0,
	settings.NonNegativeInt,
).WithPublic()

var txnRowsReadErr = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.defaults.transaction_rows_read_err",
	"the limit for the number of rows read by a SQL transaction which - "+
		"once exceeded - will fail the transaction (or will trigger a logging "+
		"event to SQL_INTERNAL_PERF for internal transactions); use 0 to disable",
	0,
	settings.NonNegativeInt,
).WithPublic()

var largeFullScanRows = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"sql.defaults.large_full_scan_rows",
	"default value for large_full_scan_rows session setting which determines "+
		"the maximum table size allowed for a full scan when disallow_full_table_scans "+
		"is set to true",
	1000.0,
).WithPublic()

var costScansWithDefaultColSize = settings.RegisterBoolSetting(
	settings.TenantWritable,
	`sql.defaults.cost_scans_with_default_col_size.enabled`,
	"setting to true uses the same size for all columns to compute scan cost",
	false,
).WithPublic()

var enableSuperRegions = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.super_regions.enabled",
	"default value for enable_super_regions; "+
		"allows for the usage of super regions",
	false,
).WithPublic()

var overrideAlterPrimaryRegionInSuperRegion = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.defaults.override_alter_primary_region_in_super_region.enabled",
	"default value for override_alter_primary_region_in_super_region; "+
		"allows for altering the primary region even if the primary region is a "+
		"member of a super region",
	false,
).WithPublic()

var errNoTransactionInProgress = errors.New("there is no transaction in progress")
var errTransactionInProgress = errors.New("there is already a transaction in progress")

const sqlTxnName string = "sql txn"
const metricsSampleInterval = 10 * time.Second

var (
	MetaSQLExecLatency = metric.Metadata{
		Name:        "sql.exec.latency",
		Help:        "Latency of SQL statement execution",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	MetaSQLServiceLatency = metric.Metadata{
		Name:        "sql.service.latency",
		Help:        "Latency of SQL request execution",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	MetaSQLOptFallback = metric.Metadata{
		Name:        "sql.optimizer.fallback.count",
		Help:        "Number of statements which the cost-based optimizer was unable to plan",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLOptPlanCacheHits = metric.Metadata{
		Name:        "sql.optimizer.plan_cache.hits",
		Help:        "Number of non-prepared statements for which a cached plan was used",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLOptPlanCacheMisses = metric.Metadata{
		Name:        "sql.optimizer.plan_cache.misses",
		Help:        "Number of non-prepared statements for which a cached plan was not used",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaDistSQLSelect = metric.Metadata{
		Name:        "sql.distsql.select.count",
		Help:        "Number of DistSQL SELECT statements",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaDistSQLExecLatency = metric.Metadata{
		Name:        "sql.distsql.exec.latency",
		Help:        "Latency of DistSQL statement execution",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	MetaDistSQLServiceLatency = metric.Metadata{
		Name:        "sql.distsql.service.latency",
		Help:        "Latency of DistSQL request execution",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	MetaTxnAbort = metric.Metadata{
		Name:        "sql.txn.abort.count",
		Help:        "Number of SQL transaction abort errors",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaFailure = metric.Metadata{
		Name:        "sql.failure.count",
		Help:        "Number of statements resulting in a planning or runtime error",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLTxnLatency = metric.Metadata{
		Name:        "sql.txn.latency",
		Help:        "Latency of SQL transactions",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	MetaSQLTxnsOpen = metric.Metadata{
		Name:        "sql.txns.open",
		Help:        "Number of currently open user SQL transactions",
		Measurement: "Open SQL Transactions",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLActiveQueries = metric.Metadata{
		Name:        "sql.statements.active",
		Help:        "Number of currently active user SQL statements",
		Measurement: "Active Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaFullTableOrIndexScan = metric.Metadata{
		Name:        "sql.full.scan.count",
		Help:        "Number of full table or index scans",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}

	MetaQueryStarted = metric.Metadata{
		Name:        "sql.query.started.count",
		Help:        "Number of SQL queries started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnBeginStarted = metric.Metadata{
		Name:        "sql.txn.begin.started.count",
		Help:        "Number of SQL transaction BEGIN statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnCommitStarted = metric.Metadata{
		Name:        "sql.txn.commit.started.count",
		Help:        "Number of SQL transaction COMMIT statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnRollbackStarted = metric.Metadata{
		Name:        "sql.txn.rollback.started.count",
		Help:        "Number of SQL transaction ROLLBACK statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaSelectStarted = metric.Metadata{
		Name:        "sql.select.started.count",
		Help:        "Number of SQL SELECT statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaUpdateStarted = metric.Metadata{
		Name:        "sql.update.started.count",
		Help:        "Number of SQL UPDATE statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaInsertStarted = metric.Metadata{
		Name:        "sql.insert.started.count",
		Help:        "Number of SQL INSERT statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaDeleteStarted = metric.Metadata{
		Name:        "sql.delete.started.count",
		Help:        "Number of SQL DELETE statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaSavepointStarted = metric.Metadata{
		Name:        "sql.savepoint.started.count",
		Help:        "Number of SQL SAVEPOINT statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaReleaseSavepointStarted = metric.Metadata{
		Name:        "sql.savepoint.release.started.count",
		Help:        "Number of `RELEASE SAVEPOINT` statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaRollbackToSavepointStarted = metric.Metadata{
		Name:        "sql.savepoint.rollback.started.count",
		Help:        "Number of `ROLLBACK TO SAVEPOINT` statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaRestartSavepointStarted = metric.Metadata{
		Name:        "sql.restart_savepoint.started.count",
		Help:        "Number of `SAVEPOINT cockroach_restart` statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaReleaseRestartSavepointStarted = metric.Metadata{
		Name:        "sql.restart_savepoint.release.started.count",
		Help:        "Number of `RELEASE SAVEPOINT cockroach_restart` statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaRollbackToRestartSavepointStarted = metric.Metadata{
		Name:        "sql.restart_savepoint.rollback.started.count",
		Help:        "Number of `ROLLBACK TO SAVEPOINT cockroach_restart` statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaDdlStarted = metric.Metadata{
		Name:        "sql.ddl.started.count",
		Help:        "Number of SQL DDL statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaCopyStarted = metric.Metadata{
		Name:        "sql.copy.started.count",
		Help:        "Number of COPY SQL statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaMiscStarted = metric.Metadata{
		Name:        "sql.misc.started.count",
		Help:        "Number of other SQL statements started",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}

	MetaQueryExecuted = metric.Metadata{
		Name:        "sql.query.count",
		Help:        "Number of SQL queries executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnBeginExecuted = metric.Metadata{
		Name:        "sql.txn.begin.count",
		Help:        "Number of SQL transaction BEGIN statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnCommitExecuted = metric.Metadata{
		Name:        "sql.txn.commit.count",
		Help:        "Number of SQL transaction COMMIT statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnRollbackExecuted = metric.Metadata{
		Name:        "sql.txn.rollback.count",
		Help:        "Number of SQL transaction ROLLBACK statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaSelectExecuted = metric.Metadata{
		Name:        "sql.select.count",
		Help:        "Number of SQL SELECT statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaUpdateExecuted = metric.Metadata{
		Name:        "sql.update.count",
		Help:        "Number of SQL UPDATE statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaInsertExecuted = metric.Metadata{
		Name:        "sql.insert.count",
		Help:        "Number of SQL INSERT statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaDeleteExecuted = metric.Metadata{
		Name:        "sql.delete.count",
		Help:        "Number of SQL DELETE statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaSavepointExecuted = metric.Metadata{
		Name:        "sql.savepoint.count",
		Help:        "Number of SQL SAVEPOINT statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaReleaseSavepointExecuted = metric.Metadata{
		Name:        "sql.savepoint.release.count",
		Help:        "Number of `RELEASE SAVEPOINT` statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaRollbackToSavepointExecuted = metric.Metadata{
		Name:        "sql.savepoint.rollback.count",
		Help:        "Number of `ROLLBACK TO SAVEPOINT` statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaRestartSavepointExecuted = metric.Metadata{
		Name:        "sql.restart_savepoint.count",
		Help:        "Number of `SAVEPOINT cockroach_restart` statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaReleaseRestartSavepointExecuted = metric.Metadata{
		Name:        "sql.restart_savepoint.release.count",
		Help:        "Number of `RELEASE SAVEPOINT cockroach_restart` statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaRollbackToRestartSavepointExecuted = metric.Metadata{
		Name:        "sql.restart_savepoint.rollback.count",
		Help:        "Number of `ROLLBACK TO SAVEPOINT cockroach_restart` statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaDdlExecuted = metric.Metadata{
		Name:        "sql.ddl.count",
		Help:        "Number of SQL DDL statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaCopyExecuted = metric.Metadata{
		Name:        "sql.copy.count",
		Help:        "Number of COPY SQL statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaMiscExecuted = metric.Metadata{
		Name:        "sql.misc.count",
		Help:        "Number of other SQL statements successfully executed",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLStatsMemMaxBytes = metric.Metadata{
		Name:        "sql.stats.mem.max",
		Help:        "Memory usage for fingerprint storage",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	MetaSQLStatsMemCurBytes = metric.Metadata{
		Name:        "sql.stats.mem.current",
		Help:        "Current memory usage for fingerprint storage",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	MetaReportedSQLStatsMemMaxBytes = metric.Metadata{
		Name:        "sql.stats.reported.mem.max",
		Help:        "Memory usage for reported fingerprint storage",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	MetaReportedSQLStatsMemCurBytes = metric.Metadata{
		Name:        "sql.stats.reported.mem.current",
		Help:        "Current memory usage for reported fingerprint storage",
		Measurement: "Memory",
		Unit:        metric.Unit_BYTES,
	}
	MetaDiscardedSQLStats = metric.Metadata{
		Name:        "sql.stats.discarded.current",
		Help:        "Number of fingerprint statistics being discarded",
		Measurement: "Discarded SQL Stats",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLStatsFlushStarted = metric.Metadata{
		Name:        "sql.stats.flush.count",
		Help:        "Number of times SQL Stats are flushed to persistent storage",
		Measurement: "SQL Stats Flush",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLStatsFlushFailure = metric.Metadata{
		Name:        "sql.stats.flush.error",
		Help:        "Number of errors encountered when flushing SQL Stats",
		Measurement: "SQL Stats Flush",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLStatsFlushDuration = metric.Metadata{
		Name:        "sql.stats.flush.duration",
		Help:        "Time took to in nanoseconds to complete SQL Stats flush",
		Measurement: "SQL Stats Flush",
		Unit:        metric.Unit_NANOSECONDS,
	}
	MetaSQLStatsRemovedRows = metric.Metadata{
		Name:        "sql.stats.cleanup.rows_removed",
		Help:        "Number of stale statistics rows that are removed",
		Measurement: "SQL Stats Cleanup",
		Unit:        metric.Unit_COUNT,
	}
	MetaSQLTxnStatsCollectionOverhead = metric.Metadata{
		Name:        "sql.stats.txn_stats_collection.duration",
		Help:        "Time took in nanoseconds to collect transaction stats",
		Measurement: "SQL Transaction Stats Collection Overhead",
		Unit:        metric.Unit_NANOSECONDS,
	}
	MetaTxnRowsWrittenLog = metric.Metadata{
		Name:        "sql.guardrails.transaction_rows_written_log.count",
		Help:        "Number of transactions logged because of transaction_rows_written_log guardrail",
		Measurement: "Logged transactions",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnRowsWrittenErr = metric.Metadata{
		Name:        "sql.guardrails.transaction_rows_written_err.count",
		Help:        "Number of transactions errored because of transaction_rows_written_err guardrail",
		Measurement: "Errored transactions",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnRowsReadLog = metric.Metadata{
		Name:        "sql.guardrails.transaction_rows_read_log.count",
		Help:        "Number of transactions logged because of transaction_rows_read_log guardrail",
		Measurement: "Logged transactions",
		Unit:        metric.Unit_COUNT,
	}
	MetaTxnRowsReadErr = metric.Metadata{
		Name:        "sql.guardrails.transaction_rows_read_err.count",
		Help:        "Number of transactions errored because of transaction_rows_read_err guardrail",
		Measurement: "Errored transactions",
		Unit:        metric.Unit_COUNT,
	}
	MetaFullTableOrIndexScanRejected = metric.Metadata{
		Name:        "sql.guardrails.full_scan_rejected.count",
		Help:        "Number of full table or index scans that have been rejected because of `disallow_full_table_scans` guardrail",
		Measurement: "SQL Statements",
		Unit:        metric.Unit_COUNT,
	}
)

func getMetricMeta(meta metric.Metadata, internal bool) metric.Metadata {
	__antithesis_instrumentation__.Notify(470353)
	if internal {
		__antithesis_instrumentation__.Notify(470355)
		meta.Name += ".internal"
		meta.Help += " (internal queries)"
		meta.Measurement = "SQL Internal Statements"
	} else {
		__antithesis_instrumentation__.Notify(470356)
	}
	__antithesis_instrumentation__.Notify(470354)
	return meta
}

type NodeInfo struct {
	LogicalClusterID func() uuid.UUID

	NodeID *base.SQLIDContainer

	AdminURL func() *url.URL

	PGURL func(*url.Userinfo) (*pgurl.URL, error)
}

type nodeStatusGenerator interface {
	GenerateNodeStatus(ctx context.Context) *statuspb.NodeStatus
}

type ExecutorConfig struct {
	Settings *cluster.Settings
	NodeInfo
	Codec             keys.SQLCodec
	DefaultZoneConfig *zonepb.ZoneConfig
	Locality          roachpb.Locality
	AmbientCtx        log.AmbientContext
	DB                *kv.DB
	Gossip            gossip.OptionalGossip
	NodeLiveness      optionalnodeliveness.Container
	SystemConfig      config.SystemConfigProvider
	DistSender        *kvcoord.DistSender
	RPCContext        *rpc.Context
	LeaseManager      *lease.Manager
	Clock             *hlc.Clock
	DistSQLSrv        *distsql.ServerImpl

	NodesStatusServer serverpb.OptionalNodesStatusServer

	SQLStatusServer    serverpb.SQLStatusServer
	TenantStatusServer serverpb.TenantStatusServer
	RegionsServer      serverpb.RegionsServer
	MetricsRecorder    nodeStatusGenerator
	SessionRegistry    *SessionRegistry
	SQLLiveness        sqlliveness.Liveness
	JobRegistry        *jobs.Registry
	VirtualSchemas     *VirtualSchemaHolder
	DistSQLPlanner     *DistSQLPlanner
	TableStatsCache    *stats.TableStatisticsCache
	StatsRefresher     *stats.Refresher
	InternalExecutor   *InternalExecutor
	QueryCache         *querycache.C

	SchemaChangerMetrics *SchemaChangerMetrics
	FeatureFlagMetrics   *featureflag.DenialMetrics
	RowMetrics           *row.Metrics
	InternalRowMetrics   *row.Metrics

	TestingKnobs                         ExecutorTestingKnobs
	MigrationTestingKnobs                *migration.TestingKnobs
	PGWireTestingKnobs                   *PGWireTestingKnobs
	SchemaChangerTestingKnobs            *SchemaChangerTestingKnobs
	DeclarativeSchemaChangerTestingKnobs *scrun.TestingKnobs
	TypeSchemaChangerTestingKnobs        *TypeSchemaChangerTestingKnobs
	GCJobTestingKnobs                    *GCJobTestingKnobs
	DistSQLRunTestingKnobs               *execinfra.TestingKnobs
	EvalContextTestingKnobs              tree.EvalContextTestingKnobs
	TenantTestingKnobs                   *TenantTestingKnobs
	TTLTestingKnobs                      *TTLTestingKnobs
	BackupRestoreTestingKnobs            *BackupRestoreTestingKnobs
	StreamingTestingKnobs                *StreamingTestingKnobs
	SQLStatsTestingKnobs                 *sqlstats.TestingKnobs
	TelemetryLoggingTestingKnobs         *TelemetryLoggingTestingKnobs
	SpanConfigTestingKnobs               *spanconfig.TestingKnobs
	CaptureIndexUsageStatsKnobs          *scheduledlogging.CaptureIndexUsageStatsTestingKnobs

	HistogramWindowInterval time.Duration

	RangeDescriptorCache *rangecache.RangeCache

	RoleMemberCache *MembershipCache

	SessionInitCache *sessioninit.Cache

	ProtectedTimestampProvider protectedts.Provider

	StmtDiagnosticsRecorder *stmtdiagnostics.Registry

	ExternalIODirConfig base.ExternalIODirConfig

	GCJobNotifier *gcjobnotifier.Notifier

	RangeFeedFactory *rangefeed.Factory

	VersionUpgradeHook VersionUpgradeHook

	MigrationJobDeps migration.JobDeps

	IndexBackfiller *IndexBackfillPlanner

	IndexValidator scexec.IndexValidator

	DescMetadaUpdaterFactory scexec.DescriptorMetadataUpdaterFactory

	ContentionRegistry *contention.Registry

	RootMemoryMonitor *mon.BytesMonitor

	CompactEngineSpanFunc tree.CompactEngineSpanFunc

	TraceCollector *collector.TraceCollector

	TenantUsageServer multitenant.TenantUsageServer

	KVStoresIterator kvserverbase.StoresIterator

	CollectionFactory *descs.CollectionFactory

	SystemTableIDResolver catalog.SystemTableIDResolver

	SpanConfigReconciler spanconfig.Reconciler

	SpanConfigSplitter spanconfig.Splitter

	SpanConfigLimiter spanconfig.Limiter

	SpanConfigKVAccessor spanconfig.KVAccessor

	InternalExecutorFactory sqlutil.SessionBoundInternalExecutorFactory
}

type UpdateVersionSystemSettingHook func(
	ctx context.Context,
	version clusterversion.ClusterVersion,
) error

type VersionUpgradeHook func(
	ctx context.Context,
	user security.SQLUsername,
	from, to clusterversion.ClusterVersion,
	updateSystemVersionSetting UpdateVersionSystemSettingHook,
) error

func (cfg *ExecutorConfig) Organization() string {
	__antithesis_instrumentation__.Notify(470357)
	return ClusterOrganization.Get(&cfg.Settings.SV)
}

func (cfg *ExecutorConfig) GetFeatureFlagMetrics() *featureflag.DenialMetrics {
	__antithesis_instrumentation__.Notify(470358)
	return cfg.FeatureFlagMetrics
}

func (cfg *ExecutorConfig) SV() *settings.Values {
	__antithesis_instrumentation__.Notify(470359)
	return &cfg.Settings.SV
}

var _ base.ModuleTestingKnobs = &ExecutorTestingKnobs{}

func (*ExecutorTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(470360) }

type StatementFilter func(context.Context, *sessiondata.SessionData, string, error)

type ExecutorTestingKnobs struct {
	StatementFilter StatementFilter

	BeforePrepare func(ctx context.Context, stmt string, txn *kv.Txn) error

	BeforeExecute func(ctx context.Context, stmt string)

	AfterExecute func(ctx context.Context, stmt string, err error)

	AfterExecCmd func(ctx context.Context, cmd Command, buf *StmtBuf)

	BeforeRestart func(ctx context.Context, reason error)

	DisableAutoCommitDuringExec bool

	BeforeAutoCommit func(ctx context.Context, stmt string) error

	DisableTempObjectsCleanupOnSessionExit bool

	TempObjectsCleanupCh chan time.Time

	OnTempObjectsCleanupDone func()

	WithStatementTrace func(trace tracing.Recording, stmt string)

	RunAfterSCJobsCacheLookup func(record *jobs.Record)

	TestingSaveFlows func(stmt string) func(map[base.SQLInstanceID]*execinfrapb.FlowSpec, execinfra.OpChains) error

	DeterministicExplain bool

	ForceRealTracingSpans bool

	DistSQLReceiverPushCallbackFactory func(query string) func(rowenc.EncDatumRow, *execinfrapb.ProducerMetadata)

	OnTxnRetry func(autoRetryReason error, evalCtx *tree.EvalContext)

	BeforeTxnStatsRecorded func(
		sessionData *sessiondata.SessionData,
		txnID uuid.UUID,
		txnFingerprintID roachpb.TransactionFingerprintID,
	)

	AfterBackupCheckpoint func()
}

type PGWireTestingKnobs struct {
	CatchPanics bool

	AuthHook func(context.Context) error

	AfterReadMsgTestingKnob func(context.Context) error
}

var _ base.ModuleTestingKnobs = &PGWireTestingKnobs{}

func (*PGWireTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(470361) }

type TenantTestingKnobs struct {
	ClusterSettingsUpdater settings.Updater

	TenantIDCodecOverride roachpb.TenantID

	OverrideTokenBucketProvider func(origProvider kvtenant.TokenBucketProvider) kvtenant.TokenBucketProvider
}

var _ base.ModuleTestingKnobs = &TenantTestingKnobs{}

func (*TenantTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(470362) }

type TTLTestingKnobs struct {
	AOSTDuration *time.Duration

	OnStatisticsError func(err error)

	MockDescriptorVersionDuringDelete *descpb.DescriptorVersion

	OnDeleteLoopStart func() error
}

func (*TTLTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(470363) }

type BackupRestoreTestingKnobs struct {
	CaptureResolvedTableDescSpans func([]roachpb.Span)

	RunAfterProcessingRestoreSpanEntry func(ctx context.Context)

	RunAfterExportingSpanEntry func(ctx context.Context, response *roachpb.ExportResponse)

	BackupMemMonitor *mon.BytesMonitor
}

var _ base.ModuleTestingKnobs = &BackupRestoreTestingKnobs{}

func (*BackupRestoreTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(470364) }

type StreamingTestingKnobs struct {
	RunAfterReceivingEvent func(ctx context.Context)
}

var _ base.ModuleTestingKnobs = &StreamingTestingKnobs{}

func (*StreamingTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(470365) }

func shouldDistributeGivenRecAndMode(
	rec distRecommendation, mode sessiondatapb.DistSQLExecMode,
) bool {
	__antithesis_instrumentation__.Notify(470366)
	switch mode {
	case sessiondatapb.DistSQLOff:
		__antithesis_instrumentation__.Notify(470368)
		return false
	case sessiondatapb.DistSQLAuto:
		__antithesis_instrumentation__.Notify(470369)
		return rec == shouldDistribute
	case sessiondatapb.DistSQLOn, sessiondatapb.DistSQLAlways:
		__antithesis_instrumentation__.Notify(470370)
		return rec != cannotDistribute
	default:
		__antithesis_instrumentation__.Notify(470371)
	}
	__antithesis_instrumentation__.Notify(470367)
	panic(errors.AssertionFailedf("unhandled distsql mode %v", mode))
}

func getPlanDistribution(
	ctx context.Context,
	p *planner,
	nodeID *base.SQLIDContainer,
	distSQLMode sessiondatapb.DistSQLExecMode,
	plan planMaybePhysical,
) physicalplan.PlanDistribution {
	__antithesis_instrumentation__.Notify(470372)
	if plan.isPhysicalPlan() {
		__antithesis_instrumentation__.Notify(470380)
		return plan.physPlan.Distribution
	} else {
		__antithesis_instrumentation__.Notify(470381)
	}
	__antithesis_instrumentation__.Notify(470373)

	if p.Descriptors().HasUncommittedTypes() {
		__antithesis_instrumentation__.Notify(470382)
		return physicalplan.LocalPlan
	} else {
		__antithesis_instrumentation__.Notify(470383)
	}
	__antithesis_instrumentation__.Notify(470374)

	if _, singleTenant := nodeID.OptionalNodeID(); !singleTenant {
		__antithesis_instrumentation__.Notify(470384)
		return physicalplan.LocalPlan
	} else {
		__antithesis_instrumentation__.Notify(470385)
	}
	__antithesis_instrumentation__.Notify(470375)
	if distSQLMode == sessiondatapb.DistSQLOff {
		__antithesis_instrumentation__.Notify(470386)
		return physicalplan.LocalPlan
	} else {
		__antithesis_instrumentation__.Notify(470387)
	}
	__antithesis_instrumentation__.Notify(470376)

	if _, ok := plan.planNode.(*zeroNode); ok {
		__antithesis_instrumentation__.Notify(470388)
		return physicalplan.LocalPlan
	} else {
		__antithesis_instrumentation__.Notify(470389)
	}
	__antithesis_instrumentation__.Notify(470377)

	rec, err := checkSupportForPlanNode(plan.planNode)
	if err != nil {
		__antithesis_instrumentation__.Notify(470390)

		log.VEventf(ctx, 1, "query not supported for distSQL: %s", err)
		return physicalplan.LocalPlan
	} else {
		__antithesis_instrumentation__.Notify(470391)
	}
	__antithesis_instrumentation__.Notify(470378)

	if shouldDistributeGivenRecAndMode(rec, distSQLMode) {
		__antithesis_instrumentation__.Notify(470392)
		return physicalplan.FullyDistributedPlan
	} else {
		__antithesis_instrumentation__.Notify(470393)
	}
	__antithesis_instrumentation__.Notify(470379)
	return physicalplan.LocalPlan
}

func golangFillQueryArguments(args ...interface{}) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(470394)
	res := make(tree.Datums, len(args))
	for i, arg := range args {
		__antithesis_instrumentation__.Notify(470396)
		if arg == nil {
			__antithesis_instrumentation__.Notify(470400)
			res[i] = tree.DNull
			continue
		} else {
			__antithesis_instrumentation__.Notify(470401)
		}
		__antithesis_instrumentation__.Notify(470397)

		var d tree.Datum
		switch t := arg.(type) {
		case tree.Datum:
			__antithesis_instrumentation__.Notify(470402)
			d = t
		case time.Time:
			__antithesis_instrumentation__.Notify(470403)
			var err error
			d, err = tree.MakeDTimestamp(t, time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(470408)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(470409)
			}
		case time.Duration:
			__antithesis_instrumentation__.Notify(470404)
			d = &tree.DInterval{Duration: duration.MakeDuration(t.Nanoseconds(), 0, 0)}
		case bitarray.BitArray:
			__antithesis_instrumentation__.Notify(470405)
			d = &tree.DBitArray{BitArray: t}
		case *apd.Decimal:
			__antithesis_instrumentation__.Notify(470406)
			dd := &tree.DDecimal{}
			dd.Set(t)
			d = dd
		case security.SQLUsername:
			__antithesis_instrumentation__.Notify(470407)
			d = tree.NewDString(t.Normalized())
		}
		__antithesis_instrumentation__.Notify(470398)
		if d == nil {
			__antithesis_instrumentation__.Notify(470410)

			val := reflect.ValueOf(arg)
			switch val.Kind() {
			case reflect.Bool:
				__antithesis_instrumentation__.Notify(470412)
				d = tree.MakeDBool(tree.DBool(val.Bool()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				__antithesis_instrumentation__.Notify(470413)
				d = tree.NewDInt(tree.DInt(val.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				__antithesis_instrumentation__.Notify(470414)
				d = tree.NewDInt(tree.DInt(val.Uint()))
			case reflect.Float32, reflect.Float64:
				__antithesis_instrumentation__.Notify(470415)
				d = tree.NewDFloat(tree.DFloat(val.Float()))
			case reflect.String:
				__antithesis_instrumentation__.Notify(470416)
				d = tree.NewDString(val.String())
			case reflect.Slice:
				__antithesis_instrumentation__.Notify(470417)
				switch {
				case val.IsNil():
					__antithesis_instrumentation__.Notify(470419)
					d = tree.DNull
				case val.Type().Elem().Kind() == reflect.String:
					__antithesis_instrumentation__.Notify(470420)
					a := tree.NewDArray(types.String)
					for v := 0; v < val.Len(); v++ {
						__antithesis_instrumentation__.Notify(470424)
						if err := a.Append(tree.NewDString(val.Index(v).String())); err != nil {
							__antithesis_instrumentation__.Notify(470425)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(470426)
						}
					}
					__antithesis_instrumentation__.Notify(470421)
					d = a
				case val.Type().Elem().Kind() == reflect.Uint8:
					__antithesis_instrumentation__.Notify(470422)
					d = tree.NewDBytes(tree.DBytes(val.Bytes()))
				default:
					__antithesis_instrumentation__.Notify(470423)
				}
			default:
				__antithesis_instrumentation__.Notify(470418)
			}
			__antithesis_instrumentation__.Notify(470411)
			if d == nil {
				__antithesis_instrumentation__.Notify(470427)
				panic(errors.AssertionFailedf("unexpected type %T", arg))
			} else {
				__antithesis_instrumentation__.Notify(470428)
			}
		} else {
			__antithesis_instrumentation__.Notify(470429)
		}
		__antithesis_instrumentation__.Notify(470399)
		res[i] = d
	}
	__antithesis_instrumentation__.Notify(470395)
	return res, nil
}

func checkResultType(typ *types.T) error {
	__antithesis_instrumentation__.Notify(470430)

	switch typ.Family() {
	case types.UnknownFamily:
		__antithesis_instrumentation__.Notify(470432)
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(470433)
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(470434)
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(470435)
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(470436)
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(470437)
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(470438)
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(470439)
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(470440)
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(470441)
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(470442)
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(470443)
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(470444)
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(470445)
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(470446)
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(470447)
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(470448)
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(470449)
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(470450)
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(470451)
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(470452)
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(470453)
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(470454)
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(470455)
	case types.VoidFamily:
		__antithesis_instrumentation__.Notify(470456)
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(470457)
		if typ.ArrayContents().Family() == types.ArrayFamily {
			__antithesis_instrumentation__.Notify(470460)

			return unimplemented.NewWithIssueDetail(32552,
				"result", "arrays cannot have arrays as element type")
		} else {
			__antithesis_instrumentation__.Notify(470461)
		}
	case types.AnyFamily:
		__antithesis_instrumentation__.Notify(470458)

		return errors.Errorf("could not determine data type of %s", typ)
	default:
		__antithesis_instrumentation__.Notify(470459)
		return errors.Errorf("unsupported result type: %s", typ)
	}
	__antithesis_instrumentation__.Notify(470431)
	return nil
}

func (p *planner) EvalAsOfTimestamp(
	ctx context.Context, asOfClause tree.AsOfClause, opts ...tree.EvalAsOfTimestampOption,
) (tree.AsOfSystemTime, error) {
	__antithesis_instrumentation__.Notify(470462)
	asOf, err := tree.EvalAsOfTimestamp(ctx, asOfClause, &p.semaCtx, p.EvalContext(), opts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(470465)
		return tree.AsOfSystemTime{}, err
	} else {
		__antithesis_instrumentation__.Notify(470466)
	}
	__antithesis_instrumentation__.Notify(470463)
	ts := asOf.Timestamp
	if now := p.execCfg.Clock.Now(); now.Less(ts) && func() bool {
		__antithesis_instrumentation__.Notify(470467)
		return !ts.Synthetic == true
	}() == true {
		__antithesis_instrumentation__.Notify(470468)
		return tree.AsOfSystemTime{}, errors.Errorf(
			"AS OF SYSTEM TIME: cannot specify timestamp in the future (%s > %s)", ts, now)
	} else {
		__antithesis_instrumentation__.Notify(470469)
	}
	__antithesis_instrumentation__.Notify(470464)
	return asOf, nil
}

func (p *planner) isAsOf(ctx context.Context, stmt tree.Statement) (*tree.AsOfSystemTime, error) {
	__antithesis_instrumentation__.Notify(470470)
	var asOf tree.AsOfClause
	switch s := stmt.(type) {
	case *tree.Select:
		__antithesis_instrumentation__.Notify(470473)
		selStmt := s.Select
		var parenSel *tree.ParenSelect
		var ok bool
		for parenSel, ok = selStmt.(*tree.ParenSelect); ok; parenSel, ok = selStmt.(*tree.ParenSelect) {
			__antithesis_instrumentation__.Notify(470484)
			selStmt = parenSel.Select.Select
		}
		__antithesis_instrumentation__.Notify(470474)

		sc, ok := selStmt.(*tree.SelectClause)
		if !ok {
			__antithesis_instrumentation__.Notify(470485)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(470486)
		}
		__antithesis_instrumentation__.Notify(470475)
		if sc.From.AsOf.Expr == nil {
			__antithesis_instrumentation__.Notify(470487)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(470488)
		}
		__antithesis_instrumentation__.Notify(470476)

		asOf = sc.From.AsOf
	case *tree.Scrub:
		__antithesis_instrumentation__.Notify(470477)
		if s.AsOf.Expr == nil {
			__antithesis_instrumentation__.Notify(470489)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(470490)
		}
		__antithesis_instrumentation__.Notify(470478)
		asOf = s.AsOf
	case *tree.Export:
		__antithesis_instrumentation__.Notify(470479)
		return p.isAsOf(ctx, s.Query)
	case *tree.CreateStats:
		__antithesis_instrumentation__.Notify(470480)
		if s.Options.AsOf.Expr == nil {
			__antithesis_instrumentation__.Notify(470491)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(470492)
		}
		__antithesis_instrumentation__.Notify(470481)
		asOf = s.Options.AsOf
	case *tree.Explain:
		__antithesis_instrumentation__.Notify(470482)
		return p.isAsOf(ctx, s.Statement)
	default:
		__antithesis_instrumentation__.Notify(470483)
		return nil, nil
	}
	__antithesis_instrumentation__.Notify(470471)
	asOfRet, err := p.EvalAsOfTimestamp(ctx, asOf, tree.EvalAsOfTimestampOptionAllowBoundedStaleness)
	if err != nil {
		__antithesis_instrumentation__.Notify(470493)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(470494)
	}
	__antithesis_instrumentation__.Notify(470472)
	return &asOfRet, err
}

func isSavepoint(ast tree.Statement) bool {
	__antithesis_instrumentation__.Notify(470495)
	_, isSavepoint := ast.(*tree.Savepoint)
	return isSavepoint
}

func isSetTransaction(ast tree.Statement) bool {
	__antithesis_instrumentation__.Notify(470496)
	_, isSet := ast.(*tree.SetTransaction)
	return isSet
}

type queryPhase int

const (
	preparing queryPhase = 0

	executing queryPhase = 1
)

type queryMeta struct {
	txnID uuid.UUID

	start time.Time

	rawStmt string

	isDistributed bool

	phase queryPhase

	ctxCancel context.CancelFunc

	hidden bool

	progressAtomic uint64
}

func (q *queryMeta) cancel() {
	__antithesis_instrumentation__.Notify(470497)
	q.ctxCancel()
}

func (q *queryMeta) getStatement() (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(470498)
	parsed, err := parser.ParseOne(q.rawStmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(470500)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(470501)
	}
	__antithesis_instrumentation__.Notify(470499)
	return parsed.AST, nil
}

type SessionDefaults map[string]string

type SessionArgs struct {
	User                        security.SQLUsername
	IsSuperuser                 bool
	SessionDefaults             SessionDefaults
	CustomOptionSessionDefaults SessionDefaults

	RemoteAddr            net.Addr
	ConnResultsBufferSize int64

	SessionRevivalToken []byte
}

type SessionRegistry struct {
	syncutil.Mutex
	sessions            map[ClusterWideID]registrySession
	sessionsByCancelKey map[pgwirecancel.BackendKeyData]registrySession
}

func NewSessionRegistry() *SessionRegistry {
	__antithesis_instrumentation__.Notify(470502)
	return &SessionRegistry{
		sessions:            make(map[ClusterWideID]registrySession),
		sessionsByCancelKey: make(map[pgwirecancel.BackendKeyData]registrySession),
	}
}

func (r *SessionRegistry) register(
	id ClusterWideID, queryCancelKey pgwirecancel.BackendKeyData, s registrySession,
) {
	__antithesis_instrumentation__.Notify(470503)
	r.Lock()
	defer r.Unlock()
	r.sessions[id] = s
	r.sessionsByCancelKey[queryCancelKey] = s
}

func (r *SessionRegistry) deregister(id ClusterWideID, queryCancelKey pgwirecancel.BackendKeyData) {
	__antithesis_instrumentation__.Notify(470504)
	r.Lock()
	defer r.Unlock()
	delete(r.sessions, id)
	delete(r.sessionsByCancelKey, queryCancelKey)
}

type registrySession interface {
	user() security.SQLUsername
	cancelQuery(queryID ClusterWideID) bool
	cancelCurrentQueries() bool
	cancelSession()

	serialize() serverpb.Session
}

func (r *SessionRegistry) CancelQuery(queryIDStr string) (bool, error) {
	__antithesis_instrumentation__.Notify(470505)
	queryID, err := StringToClusterWideID(queryIDStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(470508)
		return false, errors.Wrapf(err, "query ID %s malformed", queryID)
	} else {
		__antithesis_instrumentation__.Notify(470509)
	}
	__antithesis_instrumentation__.Notify(470506)

	r.Lock()
	defer r.Unlock()

	for _, session := range r.sessions {
		__antithesis_instrumentation__.Notify(470510)
		if session.cancelQuery(queryID) {
			__antithesis_instrumentation__.Notify(470511)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(470512)
		}
	}
	__antithesis_instrumentation__.Notify(470507)

	return false, fmt.Errorf("query ID %s not found", queryID)
}

func (r *SessionRegistry) CancelQueryByKey(
	queryCancelKey pgwirecancel.BackendKeyData,
) (canceled bool, err error) {
	__antithesis_instrumentation__.Notify(470513)
	r.Lock()
	defer r.Unlock()
	if session, ok := r.sessionsByCancelKey[queryCancelKey]; ok {
		__antithesis_instrumentation__.Notify(470515)
		if session.cancelCurrentQueries() {
			__antithesis_instrumentation__.Notify(470517)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(470518)
		}
		__antithesis_instrumentation__.Notify(470516)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(470519)
	}
	__antithesis_instrumentation__.Notify(470514)
	return false, fmt.Errorf("session for cancel key %d not found", queryCancelKey)
}

func (r *SessionRegistry) CancelSession(
	sessionIDBytes []byte,
) (*serverpb.CancelSessionResponse, error) {
	__antithesis_instrumentation__.Notify(470520)
	if len(sessionIDBytes) != 16 {
		__antithesis_instrumentation__.Notify(470523)
		return nil, errors.Errorf("invalid non-16-byte UUID %v", sessionIDBytes)
	} else {
		__antithesis_instrumentation__.Notify(470524)
	}
	__antithesis_instrumentation__.Notify(470521)
	sessionID := BytesToClusterWideID(sessionIDBytes)

	r.Lock()
	defer r.Unlock()

	for id, session := range r.sessions {
		__antithesis_instrumentation__.Notify(470525)
		if id == sessionID {
			__antithesis_instrumentation__.Notify(470526)
			session.cancelSession()
			return &serverpb.CancelSessionResponse{Canceled: true}, nil
		} else {
			__antithesis_instrumentation__.Notify(470527)
		}
	}
	__antithesis_instrumentation__.Notify(470522)

	return &serverpb.CancelSessionResponse{
		Error: fmt.Sprintf("session ID %s not found", sessionID),
	}, nil
}

func (r *SessionRegistry) SerializeAll() []serverpb.Session {
	__antithesis_instrumentation__.Notify(470528)
	r.Lock()
	defer r.Unlock()

	response := make([]serverpb.Session, 0, len(r.sessions))

	for _, s := range r.sessions {
		__antithesis_instrumentation__.Notify(470530)
		response = append(response, s.serialize())
	}
	__antithesis_instrumentation__.Notify(470529)

	return response
}

const MaxSQLBytes = 1000

type jobsCollection []jobspb.JobID

func (jc *jobsCollection) add(ids ...jobspb.JobID) {
	__antithesis_instrumentation__.Notify(470531)
	*jc = append(*jc, ids...)
}

func truncateStatementStringForTelemetry(stmt string) string {
	__antithesis_instrumentation__.Notify(470532)

	const panicLogOutputCutoffChars = 10000
	if len(stmt) > panicLogOutputCutoffChars {
		__antithesis_instrumentation__.Notify(470534)
		stmt = stmt[:len(stmt)-6] + " [...]"
	} else {
		__antithesis_instrumentation__.Notify(470535)
	}
	__antithesis_instrumentation__.Notify(470533)
	return stmt
}

func hideNonVirtualTableNameFunc(vt VirtualTabler) func(ctx *tree.FmtCtx, name *tree.TableName) {
	__antithesis_instrumentation__.Notify(470536)
	reformatFn := func(ctx *tree.FmtCtx, tn *tree.TableName) {
		__antithesis_instrumentation__.Notify(470538)
		virtual, err := vt.getVirtualTableEntry(tn)

		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(470540)
			return virtual == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(470541)

			if ctx.HasFlags(tree.FmtMarkRedactionNode) {
				__antithesis_instrumentation__.Notify(470543)

				ctx.FormatNode(&tn.CatalogName)
				ctx.WriteByte('.')

				if tn.ObjectNamePrefix.SchemaName == "public" {
					__antithesis_instrumentation__.Notify(470545)
					ctx.WithFlags(tree.FmtParsable, func() {
						__antithesis_instrumentation__.Notify(470546)
						ctx.FormatNode(&tn.ObjectNamePrefix.SchemaName)
					})
				} else {
					__antithesis_instrumentation__.Notify(470547)

					ctx.FormatNode(&tn.ObjectNamePrefix.SchemaName)
				}
				__antithesis_instrumentation__.Notify(470544)

				ctx.WriteByte('.')
				ctx.FormatNode(&tn.ObjectName)
			} else {
				__antithesis_instrumentation__.Notify(470548)

				ctx.WriteByte('_')
			}
			__antithesis_instrumentation__.Notify(470542)
			return
		} else {
			__antithesis_instrumentation__.Notify(470549)
		}
		__antithesis_instrumentation__.Notify(470539)

		newTn := *tn
		newTn.CatalogName = "_"

		ctx.WithFlags(tree.FmtParsable, func() {
			__antithesis_instrumentation__.Notify(470550)
			ctx.WithReformatTableNames(nil, func() {
				__antithesis_instrumentation__.Notify(470551)
				ctx.FormatNode(&newTn)
			})
		})
	}
	__antithesis_instrumentation__.Notify(470537)
	return reformatFn
}

func anonymizeStmtAndConstants(stmt tree.Statement, vt VirtualTabler) string {
	__antithesis_instrumentation__.Notify(470552)

	fmtFlags := tree.FmtAnonymize | tree.FmtHideConstants
	var f *tree.FmtCtx
	if vt != nil {
		__antithesis_instrumentation__.Notify(470554)
		f = tree.NewFmtCtx(
			fmtFlags,
			tree.FmtReformatTableNames(hideNonVirtualTableNameFunc(vt)),
		)
	} else {
		__antithesis_instrumentation__.Notify(470555)
		f = tree.NewFmtCtx(fmtFlags)
	}
	__antithesis_instrumentation__.Notify(470553)
	f.FormatNode(stmt)
	return f.CloseAndGetString()
}

func WithAnonymizedStatement(err error, stmt tree.Statement, vt VirtualTabler) error {
	__antithesis_instrumentation__.Notify(470556)
	anonStmtStr := anonymizeStmtAndConstants(stmt, vt)
	anonStmtStr = truncateStatementStringForTelemetry(anonStmtStr)
	return errors.WithSafeDetails(err,
		"while executing: %s", errors.Safe(anonStmtStr))
}

type SessionTracing struct {
	enabled bool

	kvTracingEnabled bool

	showResults bool

	recordingType tracing.RecordingType

	ex *connExecutor

	connSpan *tracing.Span

	lastRecording []traceRow
}

func (st *SessionTracing) getSessionTrace() ([]traceRow, error) {
	__antithesis_instrumentation__.Notify(470557)
	if !st.enabled {
		__antithesis_instrumentation__.Notify(470559)
		return st.lastRecording, nil
	} else {
		__antithesis_instrumentation__.Notify(470560)
	}
	__antithesis_instrumentation__.Notify(470558)

	return generateSessionTraceVTable(st.connSpan.GetRecording(tracing.RecordingVerbose))
}

func (st *SessionTracing) StartTracing(
	recType tracing.RecordingType, kvTracingEnabled, showResults bool,
) error {
	__antithesis_instrumentation__.Notify(470561)
	if st.enabled {
		__antithesis_instrumentation__.Notify(470565)

		if kvTracingEnabled != st.kvTracingEnabled || func() bool {
			__antithesis_instrumentation__.Notify(470567)
			return showResults != st.showResults == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(470568)
			return recType != st.recordingType == true
		}() == true {
			__antithesis_instrumentation__.Notify(470569)
			var desiredOptions bytes.Buffer
			comma := ""
			if kvTracingEnabled {
				__antithesis_instrumentation__.Notify(470572)
				desiredOptions.WriteString("kv")
				comma = ", "
			} else {
				__antithesis_instrumentation__.Notify(470573)
			}
			__antithesis_instrumentation__.Notify(470570)
			if showResults {
				__antithesis_instrumentation__.Notify(470574)
				fmt.Fprintf(&desiredOptions, "%sresults", comma)
				comma = ", "
			} else {
				__antithesis_instrumentation__.Notify(470575)
			}
			__antithesis_instrumentation__.Notify(470571)
			recOption := "cluster"
			fmt.Fprintf(&desiredOptions, "%s%s", comma, recOption)

			err := pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"tracing is already started with different options")
			return errors.WithHintf(err,
				"reset with SET tracing = off; SET tracing = %s", desiredOptions.String())
		} else {
			__antithesis_instrumentation__.Notify(470576)
		}
		__antithesis_instrumentation__.Notify(470566)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(470577)
	}
	__antithesis_instrumentation__.Notify(470562)

	var newConnCtx context.Context
	{
		__antithesis_instrumentation__.Notify(470578)
		connCtx := st.ex.ctxHolder.connCtx
		opName := "session recording"
		newConnCtx, st.connSpan = tracing.EnsureChildSpan(
			connCtx,
			st.ex.server.cfg.AmbientCtx.Tracer,
			opName,
			tracing.WithForceRealSpan(),
		)
		st.connSpan.SetRecordingType(tracing.RecordingVerbose)
		st.ex.ctxHolder.hijack(newConnCtx)
	}
	__antithesis_instrumentation__.Notify(470563)

	if _, ok := st.ex.machine.CurState().(stateNoTxn); !ok {
		__antithesis_instrumentation__.Notify(470579)
		txnCtx := st.ex.state.Ctx
		sp := tracing.SpanFromContext(txnCtx)
		if sp == nil {
			__antithesis_instrumentation__.Notify(470581)
			return errors.Errorf("no txn span for SessionTracing")
		} else {
			__antithesis_instrumentation__.Notify(470582)
		}
		__antithesis_instrumentation__.Notify(470580)

		sp.Finish()

		st.ex.state.Ctx, _ = tracing.EnsureChildSpan(
			newConnCtx, st.ex.server.cfg.AmbientCtx.Tracer, "session tracing")
	} else {
		__antithesis_instrumentation__.Notify(470583)
	}
	__antithesis_instrumentation__.Notify(470564)

	st.enabled = true
	st.kvTracingEnabled = kvTracingEnabled
	st.showResults = showResults
	st.recordingType = recType

	return nil
}

func (st *SessionTracing) StopTracing() error {
	__antithesis_instrumentation__.Notify(470584)
	if !st.enabled {
		__antithesis_instrumentation__.Notify(470586)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(470587)
	}
	__antithesis_instrumentation__.Notify(470585)
	st.enabled = false
	st.kvTracingEnabled = false
	st.showResults = false
	st.recordingType = tracing.RecordingOff

	rec := st.connSpan.GetRecording(tracing.RecordingVerbose)

	st.connSpan.SetRecordingType(tracing.RecordingOff)
	st.connSpan.Finish()
	st.connSpan = nil
	st.ex.ctxHolder.unhijack()

	var err error
	st.lastRecording, err = generateSessionTraceVTable(rec)
	return err
}

func (st *SessionTracing) KVTracingEnabled() bool {
	__antithesis_instrumentation__.Notify(470588)
	return st.kvTracingEnabled
}

func (st *SessionTracing) Enabled() bool {
	__antithesis_instrumentation__.Notify(470589)
	return st.enabled
}

func (st *SessionTracing) TracePlanStart(ctx context.Context, stmtTag string) {
	__antithesis_instrumentation__.Notify(470590)
	if st.enabled {
		__antithesis_instrumentation__.Notify(470591)
		log.VEventf(ctx, 2, "planning starts: %s", stmtTag)
	} else {
		__antithesis_instrumentation__.Notify(470592)
	}
}

func (st *SessionTracing) TracePlanEnd(ctx context.Context, err error) {
	__antithesis_instrumentation__.Notify(470593)
	log.VEventfDepth(ctx, 2, 1, "planning ends")
	if err != nil {
		__antithesis_instrumentation__.Notify(470594)
		log.VEventfDepth(ctx, 2, 1, "planning error: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(470595)
	}
}

func (st *SessionTracing) TracePlanCheckStart(ctx context.Context) {
	__antithesis_instrumentation__.Notify(470596)
	log.VEventfDepth(ctx, 2, 1, "checking distributability")
}

func (st *SessionTracing) TracePlanCheckEnd(ctx context.Context, err error, dist bool) {
	__antithesis_instrumentation__.Notify(470597)
	if err != nil {
		__antithesis_instrumentation__.Notify(470598)
		log.VEventfDepth(ctx, 2, 1, "distributability check error: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(470599)
		log.VEventfDepth(ctx, 2, 1, "will distribute plan: %v", dist)
	}
}

func (st *SessionTracing) TraceRetryInformation(ctx context.Context, retries int, err error) {
	__antithesis_instrumentation__.Notify(470600)
	log.VEventfDepth(ctx, 2, 1, "executing after %d retries, last retry reason: %v", retries, err)
}

func (st *SessionTracing) TraceExecStart(ctx context.Context, engine string) {
	__antithesis_instrumentation__.Notify(470601)
	log.VEventfDepth(ctx, 2, 1, "execution starts: %s engine", engine)
}

func (st *SessionTracing) TraceExecConsume(ctx context.Context) (context.Context, func()) {
	__antithesis_instrumentation__.Notify(470602)
	if st.enabled {
		__antithesis_instrumentation__.Notify(470604)
		consumeCtx, sp := tracing.ChildSpan(ctx, "consuming rows")
		return consumeCtx, sp.Finish
	} else {
		__antithesis_instrumentation__.Notify(470605)
	}
	__antithesis_instrumentation__.Notify(470603)
	return ctx, func() { __antithesis_instrumentation__.Notify(470606) }
}

func (st *SessionTracing) TraceExecRowsResult(ctx context.Context, values tree.Datums) {
	__antithesis_instrumentation__.Notify(470607)
	if st.showResults {
		__antithesis_instrumentation__.Notify(470608)
		log.VEventfDepth(ctx, 2, 1, "output row: %s", values)
	} else {
		__antithesis_instrumentation__.Notify(470609)
	}
}

func (st *SessionTracing) TraceExecBatchResult(ctx context.Context, batch coldata.Batch) {
	__antithesis_instrumentation__.Notify(470610)
	if st.showResults {
		__antithesis_instrumentation__.Notify(470611)
		outputRows := coldata.VecsToStringWithRowPrefix(batch.ColVecs(), batch.Length(), batch.Selection(), "output row: ")
		for _, row := range outputRows {
			__antithesis_instrumentation__.Notify(470612)
			log.VEventfDepth(ctx, 2, 1, "%s", row)
		}
	} else {
		__antithesis_instrumentation__.Notify(470613)
	}
}

func (st *SessionTracing) TraceExecEnd(ctx context.Context, err error, count int) {
	__antithesis_instrumentation__.Notify(470614)
	log.VEventfDepth(ctx, 2, 1, "execution ends")
	if err != nil {
		__antithesis_instrumentation__.Notify(470615)
		log.VEventfDepth(ctx, 2, 1, "execution failed after %d rows: %v", count, err)
	} else {
		__antithesis_instrumentation__.Notify(470616)
		log.VEventfDepth(ctx, 2, 1, "rows affected: %d", count)
	}
}

const (
	traceSpanIdxCol = iota

	_

	traceTimestampCol

	traceDurationCol

	traceOpCol

	traceLocCol

	traceTagCol

	traceMsgCol

	traceAgeCol

	traceNumCols
)

type traceRow [traceNumCols]tree.Datum

var logMessageRE = regexp.MustCompile(
	`(?s:^((?:[^][ :]+/[^][ :]+\.[^][ :]+:[0-9]+)?) *((?:\[(?:[^][]|\[[^]]*\])*\])?) *(.*))`)

func generateSessionTraceVTable(spans []tracingpb.RecordedSpan) ([]traceRow, error) {
	__antithesis_instrumentation__.Notify(470617)

	var allLogs []logRecordRow

	seenSpans := make(map[tracingpb.SpanID]struct{})
	for spanIdx, span := range spans {
		__antithesis_instrumentation__.Notify(470622)
		if _, ok := seenSpans[span.SpanID]; ok {
			__antithesis_instrumentation__.Notify(470625)
			continue
		} else {
			__antithesis_instrumentation__.Notify(470626)
		}
		__antithesis_instrumentation__.Notify(470623)
		spanWithIndex := spanWithIndex{
			RecordedSpan: &spans[spanIdx],
			index:        spanIdx,
		}
		msgs, err := getMessagesForSubtrace(spanWithIndex, spans, seenSpans)
		if err != nil {
			__antithesis_instrumentation__.Notify(470627)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(470628)
		}
		__antithesis_instrumentation__.Notify(470624)
		allLogs = append(allLogs, msgs...)
	}
	__antithesis_instrumentation__.Notify(470618)

	opMap := make(map[tree.DInt]*tree.DString)
	durMap := make(map[tree.DInt]*tree.DInterval)
	var res []traceRow
	var minTimestamp, zeroTime time.Time
	for _, lrr := range allLogs {
		__antithesis_instrumentation__.Notify(470629)

		if lrr.index == 0 {
			__antithesis_instrumentation__.Notify(470634)
			spanIdx := tree.DInt(lrr.span.index)
			opMap[spanIdx] = tree.NewDString(lrr.span.Operation)
			if lrr.span.Duration != 0 {
				__antithesis_instrumentation__.Notify(470635)
				durMap[spanIdx] = &tree.DInterval{
					Duration: duration.MakeDuration(lrr.span.Duration.Nanoseconds(), 0, 0),
				}
			} else {
				__antithesis_instrumentation__.Notify(470636)
			}
		} else {
			__antithesis_instrumentation__.Notify(470637)
		}
		__antithesis_instrumentation__.Notify(470630)

		if minTimestamp == zeroTime || func() bool {
			__antithesis_instrumentation__.Notify(470638)
			return lrr.timestamp.Before(minTimestamp) == true
		}() == true {
			__antithesis_instrumentation__.Notify(470639)
			minTimestamp = lrr.timestamp
		} else {
			__antithesis_instrumentation__.Notify(470640)
		}
		__antithesis_instrumentation__.Notify(470631)

		loc := logMessageRE.FindStringSubmatchIndex(lrr.msg)
		if loc == nil {
			__antithesis_instrumentation__.Notify(470641)
			return nil, fmt.Errorf("unable to split trace message: %q", lrr.msg)
		} else {
			__antithesis_instrumentation__.Notify(470642)
		}
		__antithesis_instrumentation__.Notify(470632)

		tsDatum, err := tree.MakeDTimestampTZ(lrr.timestamp, time.Nanosecond)
		if err != nil {
			__antithesis_instrumentation__.Notify(470643)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(470644)
		}
		__antithesis_instrumentation__.Notify(470633)

		row := traceRow{
			tree.NewDInt(tree.DInt(lrr.span.index)),
			tree.NewDInt(tree.DInt(lrr.index)),
			tsDatum,
			tree.DNull,
			tree.DNull,
			tree.NewDString(lrr.msg[loc[2]:loc[3]]),
			tree.NewDString(lrr.msg[loc[4]:loc[5]]),
			tree.NewDString(lrr.msg[loc[6]:loc[7]]),
			tree.DNull,
		}
		res = append(res, row)
	}
	__antithesis_instrumentation__.Notify(470619)

	if len(res) == 0 {
		__antithesis_instrumentation__.Notify(470645)

		return res, nil
	} else {
		__antithesis_instrumentation__.Notify(470646)
	}
	__antithesis_instrumentation__.Notify(470620)

	for i := range res {
		__antithesis_instrumentation__.Notify(470647)
		spanIdx := res[i][traceSpanIdxCol]

		if opStr, ok := opMap[*(spanIdx.(*tree.DInt))]; ok {
			__antithesis_instrumentation__.Notify(470650)
			res[i][traceOpCol] = opStr
		} else {
			__antithesis_instrumentation__.Notify(470651)
		}
		__antithesis_instrumentation__.Notify(470648)

		if dur, ok := durMap[*(spanIdx.(*tree.DInt))]; ok {
			__antithesis_instrumentation__.Notify(470652)
			res[i][traceDurationCol] = dur
		} else {
			__antithesis_instrumentation__.Notify(470653)
		}
		__antithesis_instrumentation__.Notify(470649)

		ts := res[i][traceTimestampCol].(*tree.DTimestampTZ)
		res[i][traceAgeCol] = &tree.DInterval{
			Duration: duration.MakeDuration(ts.Sub(minTimestamp).Nanoseconds(), 0, 0),
		}
	}
	__antithesis_instrumentation__.Notify(470621)

	return res, nil
}

func getOrderedChildSpans(
	spanID tracingpb.SpanID, allSpans []tracingpb.RecordedSpan,
) []spanWithIndex {
	__antithesis_instrumentation__.Notify(470654)
	children := make([]spanWithIndex, 0)
	for i := range allSpans {
		__antithesis_instrumentation__.Notify(470656)
		if allSpans[i].ParentSpanID == spanID {
			__antithesis_instrumentation__.Notify(470657)
			children = append(
				children,
				spanWithIndex{
					RecordedSpan: &allSpans[i],
					index:        i,
				})
		} else {
			__antithesis_instrumentation__.Notify(470658)
		}
	}
	__antithesis_instrumentation__.Notify(470655)
	return children
}

func getMessagesForSubtrace(
	span spanWithIndex, allSpans []tracingpb.RecordedSpan, seenSpans map[tracingpb.SpanID]struct{},
) ([]logRecordRow, error) {
	__antithesis_instrumentation__.Notify(470659)
	if _, ok := seenSpans[span.SpanID]; ok {
		__antithesis_instrumentation__.Notify(470663)
		return nil, errors.Errorf("duplicate span %d", span.SpanID)
	} else {
		__antithesis_instrumentation__.Notify(470664)
	}
	__antithesis_instrumentation__.Notify(470660)
	var allLogs []logRecordRow
	const spanStartMsgTemplate = "=== SPAN START: %s ==="

	spanStartMsgs := make([]string, 0, len(span.Tags)+1)

	spanStartMsgs = append(spanStartMsgs, fmt.Sprintf(spanStartMsgTemplate, span.Operation))

	for name, value := range span.Tags {
		__antithesis_instrumentation__.Notify(470665)
		if !strings.HasPrefix(name, tracing.TagPrefix) {
			__antithesis_instrumentation__.Notify(470667)

			continue
		} else {
			__antithesis_instrumentation__.Notify(470668)
		}
		__antithesis_instrumentation__.Notify(470666)
		spanStartMsgs = append(spanStartMsgs, fmt.Sprintf("%s: %s", name, value))
	}
	__antithesis_instrumentation__.Notify(470661)
	sort.Strings(spanStartMsgs[1:])

	allLogs = append(
		allLogs,
		logRecordRow{
			timestamp: span.StartTime,
			msg:       strings.Join(spanStartMsgs, "\n"),
			span:      span,
			index:     0,
		},
	)

	seenSpans[span.SpanID] = struct{}{}
	childSpans := getOrderedChildSpans(span.SpanID, allSpans)
	var i, j int

	maxTime := time.Date(6000, 0, 0, 0, 0, 0, 0, time.UTC)

	for i < len(span.Logs) || func() bool {
		__antithesis_instrumentation__.Notify(470669)
		return j < len(childSpans) == true
	}() == true {
		__antithesis_instrumentation__.Notify(470670)
		logTime := maxTime
		childTime := maxTime
		if i < len(span.Logs) {
			__antithesis_instrumentation__.Notify(470673)
			logTime = span.Logs[i].Time
		} else {
			__antithesis_instrumentation__.Notify(470674)
		}
		__antithesis_instrumentation__.Notify(470671)
		if j < len(childSpans) {
			__antithesis_instrumentation__.Notify(470675)
			childTime = childSpans[j].StartTime
		} else {
			__antithesis_instrumentation__.Notify(470676)
		}
		__antithesis_instrumentation__.Notify(470672)

		if logTime.Before(childTime) {
			__antithesis_instrumentation__.Notify(470677)
			allLogs = append(allLogs,
				logRecordRow{
					timestamp: logTime,
					msg:       span.Logs[i].Msg().StripMarkers(),
					span:      span,

					index: i + 1,
				})
			i++
		} else {
			__antithesis_instrumentation__.Notify(470678)

			childMsgs, err := getMessagesForSubtrace(childSpans[j], allSpans, seenSpans)
			if err != nil {
				__antithesis_instrumentation__.Notify(470680)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(470681)
			}
			__antithesis_instrumentation__.Notify(470679)
			allLogs = append(allLogs, childMsgs...)
			j++
		}
	}
	__antithesis_instrumentation__.Notify(470662)
	return allLogs, nil
}

type logRecordRow struct {
	timestamp time.Time
	msg       string
	span      spanWithIndex

	index int
}

type spanWithIndex struct {
	*tracingpb.RecordedSpan
	index int
}

type paramStatusUpdater interface {
	BufferParamStatusUpdate(string, string)
}

type bufferableParamStatusUpdate struct {
	name      string
	lowerName string
}

var bufferableParamStatusUpdates = func() []bufferableParamStatusUpdate {
	__antithesis_instrumentation__.Notify(470682)
	params := []string{
		"application_name",
		"DateStyle",
		"IntervalStyle",
		"is_superuser",
		"TimeZone",
	}
	ret := make([]bufferableParamStatusUpdate, len(params))
	for i, param := range params {
		__antithesis_instrumentation__.Notify(470684)
		ret[i] = bufferableParamStatusUpdate{
			name:      param,
			lowerName: strings.ToLower(param),
		}
	}
	__antithesis_instrumentation__.Notify(470683)
	return ret
}()

type sessionDataMutatorBase struct {
	defaults SessionDefaults
	settings *cluster.Settings
}

type sessionDataMutatorCallbacks struct {
	paramStatusUpdater paramStatusUpdater

	setCurTxnReadOnly func(val bool)

	onTempSchemaCreation func()

	onDefaultIntSizeChange func(int32)

	onApplicationNameChange func(string)
}

type sessionDataMutatorIterator struct {
	sessionDataMutatorBase
	sds *sessiondata.Stack
	sessionDataMutatorCallbacks
}

func (it *sessionDataMutatorIterator) mutator(
	applyCallbacks bool, sd *sessiondata.SessionData,
) sessionDataMutator {
	__antithesis_instrumentation__.Notify(470685)
	ret := sessionDataMutator{
		data:                   sd,
		sessionDataMutatorBase: it.sessionDataMutatorBase,
	}

	if applyCallbacks {
		__antithesis_instrumentation__.Notify(470687)
		ret.sessionDataMutatorCallbacks = it.sessionDataMutatorCallbacks
	} else {
		__antithesis_instrumentation__.Notify(470688)
	}
	__antithesis_instrumentation__.Notify(470686)
	return ret
}

func (it *sessionDataMutatorIterator) SetSessionDefaultIntSize(size int32) {
	__antithesis_instrumentation__.Notify(470689)
	it.applyOnEachMutator(func(m sessionDataMutator) {
		__antithesis_instrumentation__.Notify(470690)
		m.SetDefaultIntSize(size)
	})
}

func (it *sessionDataMutatorIterator) applyOnTopMutator(
	applyFunc func(m sessionDataMutator) error,
) error {
	__antithesis_instrumentation__.Notify(470691)
	return applyFunc(it.mutator(true, it.sds.Top()))
}

func (it *sessionDataMutatorIterator) applyOnEachMutator(applyFunc func(m sessionDataMutator)) {
	__antithesis_instrumentation__.Notify(470692)
	elems := it.sds.Elems()
	for i, sd := range elems {
		__antithesis_instrumentation__.Notify(470693)
		applyFunc(it.mutator(i == 0, sd))
	}
}

func (it *sessionDataMutatorIterator) applyOnEachMutatorError(
	applyFunc func(m sessionDataMutator) error,
) error {
	__antithesis_instrumentation__.Notify(470694)
	elems := it.sds.Elems()
	for i, sd := range elems {
		__antithesis_instrumentation__.Notify(470696)
		if err := applyFunc(it.mutator(i == 0, sd)); err != nil {
			__antithesis_instrumentation__.Notify(470697)
			return err
		} else {
			__antithesis_instrumentation__.Notify(470698)
		}
	}
	__antithesis_instrumentation__.Notify(470695)
	return nil
}

type sessionDataMutator struct {
	data *sessiondata.SessionData
	sessionDataMutatorBase
	sessionDataMutatorCallbacks
}

func (m *sessionDataMutator) bufferParamStatusUpdate(param string, status string) {
	__antithesis_instrumentation__.Notify(470699)
	if m.paramStatusUpdater != nil {
		__antithesis_instrumentation__.Notify(470700)
		m.paramStatusUpdater.BufferParamStatusUpdate(param, status)
	} else {
		__antithesis_instrumentation__.Notify(470701)
	}
}

func (m *sessionDataMutator) SetApplicationName(appName string) {
	__antithesis_instrumentation__.Notify(470702)
	m.data.ApplicationName = appName
	if m.onApplicationNameChange != nil {
		__antithesis_instrumentation__.Notify(470704)
		m.onApplicationNameChange(appName)
	} else {
		__antithesis_instrumentation__.Notify(470705)
	}
	__antithesis_instrumentation__.Notify(470703)
	m.bufferParamStatusUpdate("application_name", appName)
}

func (m *sessionDataMutator) SetAvoidBuffering(b bool) {
	__antithesis_instrumentation__.Notify(470706)
	m.data.AvoidBuffering = b
}

func (m *sessionDataMutator) SetBytesEncodeFormat(val lex.BytesEncodeFormat) {
	__antithesis_instrumentation__.Notify(470707)
	m.data.DataConversionConfig.BytesEncodeFormat = val
}

func (m *sessionDataMutator) SetExtraFloatDigits(val int32) {
	__antithesis_instrumentation__.Notify(470708)
	m.data.DataConversionConfig.ExtraFloatDigits = val
}

func (m *sessionDataMutator) SetDatabase(dbName string) {
	__antithesis_instrumentation__.Notify(470709)
	m.data.Database = dbName
}

func (m *sessionDataMutator) SetTemporarySchemaName(scName string) {
	__antithesis_instrumentation__.Notify(470710)
	if m.onTempSchemaCreation != nil {
		__antithesis_instrumentation__.Notify(470712)
		m.onTempSchemaCreation()
	} else {
		__antithesis_instrumentation__.Notify(470713)
	}
	__antithesis_instrumentation__.Notify(470711)
	m.data.SearchPath = m.data.SearchPath.WithTemporarySchemaName(scName)
}

func (m *sessionDataMutator) SetTemporarySchemaIDForDatabase(dbID uint32, tempSchemaID uint32) {
	__antithesis_instrumentation__.Notify(470714)
	if m.data.DatabaseIDToTempSchemaID == nil {
		__antithesis_instrumentation__.Notify(470716)
		m.data.DatabaseIDToTempSchemaID = make(map[uint32]uint32)
	} else {
		__antithesis_instrumentation__.Notify(470717)
	}
	__antithesis_instrumentation__.Notify(470715)
	m.data.DatabaseIDToTempSchemaID[dbID] = tempSchemaID
}

func (m *sessionDataMutator) SetDefaultIntSize(size int32) {
	__antithesis_instrumentation__.Notify(470718)
	m.data.DefaultIntSize = size
	if m.onDefaultIntSizeChange != nil {
		__antithesis_instrumentation__.Notify(470719)
		m.onDefaultIntSizeChange(size)
	} else {
		__antithesis_instrumentation__.Notify(470720)
	}
}

func (m *sessionDataMutator) SetDefaultTransactionPriority(val tree.UserPriority) {
	__antithesis_instrumentation__.Notify(470721)
	m.data.DefaultTxnPriority = int64(val)
}

func (m *sessionDataMutator) SetDefaultTransactionReadOnly(val bool) {
	__antithesis_instrumentation__.Notify(470722)
	m.data.DefaultTxnReadOnly = val
}

func (m *sessionDataMutator) SetDefaultTransactionUseFollowerReads(val bool) {
	__antithesis_instrumentation__.Notify(470723)
	m.data.DefaultTxnUseFollowerReads = val
}

func (m *sessionDataMutator) SetEnableSeqScan(val bool) {
	__antithesis_instrumentation__.Notify(470724)
	m.data.EnableSeqScan = val
}

func (m *sessionDataMutator) SetSynchronousCommit(val bool) {
	__antithesis_instrumentation__.Notify(470725)
	m.data.SynchronousCommit = val
}

func (m *sessionDataMutator) SetDisablePlanGists(val bool) {
	__antithesis_instrumentation__.Notify(470726)
	m.data.DisablePlanGists = val
}

func (m *sessionDataMutator) SetDistSQLMode(val sessiondatapb.DistSQLExecMode) {
	__antithesis_instrumentation__.Notify(470727)
	m.data.DistSQLMode = val
}

func (m *sessionDataMutator) SetDistSQLWorkMem(val int64) {
	__antithesis_instrumentation__.Notify(470728)
	m.data.WorkMemLimit = val
}

func (m *sessionDataMutator) SetForceSavepointRestart(val bool) {
	__antithesis_instrumentation__.Notify(470729)
	m.data.ForceSavepointRestart = val
}

func (m *sessionDataMutator) SetZigzagJoinEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470730)
	m.data.ZigzagJoinEnabled = val
}

func (m *sessionDataMutator) SetIndexRecommendationsEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470731)
	m.data.IndexRecommendationsEnabled = val
}

func (m *sessionDataMutator) SetExperimentalDistSQLPlanning(
	val sessiondatapb.ExperimentalDistSQLPlanningMode,
) {
	__antithesis_instrumentation__.Notify(470732)
	m.data.ExperimentalDistSQLPlanningMode = val
}

func (m *sessionDataMutator) SetPartiallyDistributedPlansDisabled(val bool) {
	__antithesis_instrumentation__.Notify(470733)
	m.data.PartiallyDistributedPlansDisabled = val
}

func (m *sessionDataMutator) SetRequireExplicitPrimaryKeys(val bool) {
	__antithesis_instrumentation__.Notify(470734)
	m.data.RequireExplicitPrimaryKeys = val
}

func (m *sessionDataMutator) SetReorderJoinsLimit(val int) {
	__antithesis_instrumentation__.Notify(470735)
	m.data.ReorderJoinsLimit = int64(val)
}

func (m *sessionDataMutator) SetVectorize(val sessiondatapb.VectorizeExecMode) {
	__antithesis_instrumentation__.Notify(470736)
	m.data.VectorizeMode = val
}

func (m *sessionDataMutator) SetTestingVectorizeInjectPanics(val bool) {
	__antithesis_instrumentation__.Notify(470737)
	m.data.TestingVectorizeInjectPanics = val
}

func (m *sessionDataMutator) SetOptimizerFKCascadesLimit(val int) {
	__antithesis_instrumentation__.Notify(470738)
	m.data.OptimizerFKCascadesLimit = int64(val)
}

func (m *sessionDataMutator) SetOptimizerUseHistograms(val bool) {
	__antithesis_instrumentation__.Notify(470739)
	m.data.OptimizerUseHistograms = val
}

func (m *sessionDataMutator) SetOptimizerUseMultiColStats(val bool) {
	__antithesis_instrumentation__.Notify(470740)
	m.data.OptimizerUseMultiColStats = val
}

func (m *sessionDataMutator) SetLocalityOptimizedSearch(val bool) {
	__antithesis_instrumentation__.Notify(470741)
	m.data.LocalityOptimizedSearch = val
}

func (m *sessionDataMutator) SetImplicitSelectForUpdate(val bool) {
	__antithesis_instrumentation__.Notify(470742)
	m.data.ImplicitSelectForUpdate = val
}

func (m *sessionDataMutator) SetInsertFastPath(val bool) {
	__antithesis_instrumentation__.Notify(470743)
	m.data.InsertFastPath = val
}

func (m *sessionDataMutator) SetSerialNormalizationMode(val sessiondatapb.SerialNormalizationMode) {
	__antithesis_instrumentation__.Notify(470744)
	m.data.SerialNormalizationMode = val
}

func (m *sessionDataMutator) SetSafeUpdates(val bool) {
	__antithesis_instrumentation__.Notify(470745)
	m.data.SafeUpdates = val
}

func (m *sessionDataMutator) SetCheckFunctionBodies(val bool) {
	__antithesis_instrumentation__.Notify(470746)
	m.data.CheckFunctionBodies = val
}

func (m *sessionDataMutator) SetPreferLookupJoinsForFKs(val bool) {
	__antithesis_instrumentation__.Notify(470747)
	m.data.PreferLookupJoinsForFKs = val
}

func (m *sessionDataMutator) UpdateSearchPath(paths []string) {
	__antithesis_instrumentation__.Notify(470748)
	m.data.SearchPath = m.data.SearchPath.UpdatePaths(paths)
}

func (m *sessionDataMutator) SetLocation(loc *time.Location) {
	__antithesis_instrumentation__.Notify(470749)
	m.data.Location = loc
	m.bufferParamStatusUpdate("TimeZone", sessionDataTimeZoneFormat(loc))
}

func (m *sessionDataMutator) SetCustomOption(name, val string) {
	__antithesis_instrumentation__.Notify(470750)
	if m.data.CustomOptions == nil {
		__antithesis_instrumentation__.Notify(470752)
		m.data.CustomOptions = make(map[string]string)
	} else {
		__antithesis_instrumentation__.Notify(470753)
	}
	__antithesis_instrumentation__.Notify(470751)
	m.data.CustomOptions[name] = val
}

func (m *sessionDataMutator) SetReadOnly(val bool) {
	__antithesis_instrumentation__.Notify(470754)

	if m.setCurTxnReadOnly != nil {
		__antithesis_instrumentation__.Notify(470755)
		m.setCurTxnReadOnly(val)
	} else {
		__antithesis_instrumentation__.Notify(470756)
	}
}

func (m *sessionDataMutator) SetStmtTimeout(timeout time.Duration) {
	__antithesis_instrumentation__.Notify(470757)
	m.data.StmtTimeout = timeout
}

func (m *sessionDataMutator) SetLockTimeout(timeout time.Duration) {
	__antithesis_instrumentation__.Notify(470758)
	m.data.LockTimeout = timeout
}

func (m *sessionDataMutator) SetIdleInSessionTimeout(timeout time.Duration) {
	__antithesis_instrumentation__.Notify(470759)
	m.data.IdleInSessionTimeout = timeout
}

func (m *sessionDataMutator) SetIdleInTransactionSessionTimeout(timeout time.Duration) {
	__antithesis_instrumentation__.Notify(470760)
	m.data.IdleInTransactionSessionTimeout = timeout
}

func (m *sessionDataMutator) SetAllowPrepareAsOptPlan(val bool) {
	__antithesis_instrumentation__.Notify(470761)
	m.data.AllowPrepareAsOptPlan = val
}

func (m *sessionDataMutator) SetSaveTablesPrefix(prefix string) {
	__antithesis_instrumentation__.Notify(470762)
	m.data.SaveTablesPrefix = prefix
}

func (m *sessionDataMutator) SetPlacementEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470763)
	m.data.PlacementEnabled = val
}

func (m *sessionDataMutator) SetAutoRehomingEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470764)
	m.data.AutoRehomingEnabled = val
}

func (m *sessionDataMutator) SetOnUpdateRehomeRowEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470765)
	m.data.OnUpdateRehomeRowEnabled = val
}

func (m *sessionDataMutator) SetTempTablesEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470766)
	m.data.TempTablesEnabled = val
}

func (m *sessionDataMutator) SetImplicitColumnPartitioningEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470767)
	m.data.ImplicitColumnPartitioningEnabled = val
}

func (m *sessionDataMutator) SetOverrideMultiRegionZoneConfigEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470768)
	m.data.OverrideMultiRegionZoneConfigEnabled = val
}

func (m *sessionDataMutator) SetDisallowFullTableScans(val bool) {
	__antithesis_instrumentation__.Notify(470769)
	m.data.DisallowFullTableScans = val
}

func (m *sessionDataMutator) SetAlterColumnTypeGeneral(val bool) {
	__antithesis_instrumentation__.Notify(470770)
	m.data.AlterColumnTypeGeneralEnabled = val
}

func (m *sessionDataMutator) SetEnableSuperRegions(val bool) {
	__antithesis_instrumentation__.Notify(470771)
	m.data.EnableSuperRegions = val
}

func (m *sessionDataMutator) SetEnableOverrideAlterPrimaryRegionInSuperRegion(val bool) {
	__antithesis_instrumentation__.Notify(470772)
	m.data.OverrideAlterPrimaryRegionInSuperRegion = val
}

func (m *sessionDataMutator) SetUniqueWithoutIndexConstraints(val bool) {
	__antithesis_instrumentation__.Notify(470773)
	m.data.EnableUniqueWithoutIndexConstraints = val
}

func (m *sessionDataMutator) SetUseNewSchemaChanger(val sessiondatapb.NewSchemaChangerMode) {
	__antithesis_instrumentation__.Notify(470774)
	m.data.NewSchemaChangerMode = val
}

func (m *sessionDataMutator) SetQualityOfService(val sessiondatapb.QoSLevel) {
	__antithesis_instrumentation__.Notify(470775)
	m.data.DefaultTxnQualityOfService = val.Validate()
}

func (m *sessionDataMutator) SetOptSplitScanLimit(val int32) {
	__antithesis_instrumentation__.Notify(470776)
	m.data.OptSplitScanLimit = val
}

func (m *sessionDataMutator) SetStreamReplicationEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470777)
	m.data.EnableStreamReplication = val
}

func (m *sessionDataMutator) RecordLatestSequenceVal(seqID uint32, val int64) {
	__antithesis_instrumentation__.Notify(470778)
	m.data.SequenceState.RecordValue(seqID, val)
}

func (m *sessionDataMutator) SetNoticeDisplaySeverity(severity pgnotice.DisplaySeverity) {
	__antithesis_instrumentation__.Notify(470779)
	m.data.NoticeDisplaySeverity = uint32(severity)
}

func (m *sessionDataMutator) initSequenceCache() {
	__antithesis_instrumentation__.Notify(470780)
	m.data.SequenceCache = sessiondatapb.SequenceCache{}
}

func (m *sessionDataMutator) SetIntervalStyle(style duration.IntervalStyle) {
	__antithesis_instrumentation__.Notify(470781)
	m.data.DataConversionConfig.IntervalStyle = style
	m.bufferParamStatusUpdate("IntervalStyle", strings.ToLower(style.String()))
}

func (m *sessionDataMutator) SetDateStyle(style pgdate.DateStyle) {
	__antithesis_instrumentation__.Notify(470782)
	m.data.DataConversionConfig.DateStyle = style
	m.bufferParamStatusUpdate("DateStyle", style.SQLString())
}

func (m *sessionDataMutator) SetIntervalStyleEnabled(enabled bool) {
	__antithesis_instrumentation__.Notify(470783)
	m.data.IntervalStyleEnabled = enabled
}

func (m *sessionDataMutator) SetDateStyleEnabled(enabled bool) {
	__antithesis_instrumentation__.Notify(470784)
	m.data.DateStyleEnabled = enabled
}

func (m *sessionDataMutator) SetStubCatalogTablesEnabled(enabled bool) {
	__antithesis_instrumentation__.Notify(470785)
	m.data.StubCatalogTablesEnabled = enabled
}

func (m *sessionDataMutator) SetExperimentalComputedColumnRewrites(val string) {
	__antithesis_instrumentation__.Notify(470786)
	m.data.ExperimentalComputedColumnRewrites = val
}

func (m *sessionDataMutator) SetNullOrderedLast(b bool) {
	__antithesis_instrumentation__.Notify(470787)
	m.data.NullOrderedLast = b
}

func (m *sessionDataMutator) SetPropagateInputOrdering(b bool) {
	__antithesis_instrumentation__.Notify(470788)
	m.data.PropagateInputOrdering = b
}

func (m *sessionDataMutator) SetTxnRowsWrittenLog(val int64) {
	__antithesis_instrumentation__.Notify(470789)
	m.data.TxnRowsWrittenLog = val
}

func (m *sessionDataMutator) SetTxnRowsWrittenErr(val int64) {
	__antithesis_instrumentation__.Notify(470790)
	m.data.TxnRowsWrittenErr = val
}

func (m *sessionDataMutator) SetTxnRowsReadLog(val int64) {
	__antithesis_instrumentation__.Notify(470791)
	m.data.TxnRowsReadLog = val
}

func (m *sessionDataMutator) SetTxnRowsReadErr(val int64) {
	__antithesis_instrumentation__.Notify(470792)
	m.data.TxnRowsReadErr = val
}

func (m *sessionDataMutator) SetLargeFullScanRows(val float64) {
	__antithesis_instrumentation__.Notify(470793)
	m.data.LargeFullScanRows = val
}

func (m *sessionDataMutator) SetInjectRetryErrorsEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470794)
	m.data.InjectRetryErrorsEnabled = val
}

func (m *sessionDataMutator) SetJoinReaderOrderingStrategyBatchSize(val int64) {
	__antithesis_instrumentation__.Notify(470795)
	m.data.JoinReaderOrderingStrategyBatchSize = val
}

func (m *sessionDataMutator) SetParallelizeMultiKeyLookupJoinsEnabled(val bool) {
	__antithesis_instrumentation__.Notify(470796)
	m.data.ParallelizeMultiKeyLookupJoinsEnabled = val
}

func (m *sessionDataMutator) SetCostScansWithDefaultColSize(val bool) {
	__antithesis_instrumentation__.Notify(470797)
	m.data.CostScansWithDefaultColSize = val
}

func (m *sessionDataMutator) SetEnableImplicitTransactionForBatchStatements(val bool) {
	__antithesis_instrumentation__.Notify(470798)
	m.data.EnableImplicitTransactionForBatchStatements = val
}

func (m *sessionDataMutator) SetExpectAndIgnoreNotVisibleColumnsInCopy(val bool) {
	__antithesis_instrumentation__.Notify(470799)
	m.data.ExpectAndIgnoreNotVisibleColumnsInCopy = val
}

func quantizeCounts(d *roachpb.StatementStatistics) {
	__antithesis_instrumentation__.Notify(470800)
	oldCount := d.Count
	newCount := telemetry.Bucket10(oldCount)
	d.Count = newCount

	oldCountMinusOne := float64(oldCount - 1)
	newCountMinusOne := float64(newCount - 1)
	d.NumRows.SquaredDiffs = (d.NumRows.SquaredDiffs / oldCountMinusOne) * newCountMinusOne
	d.ParseLat.SquaredDiffs = (d.ParseLat.SquaredDiffs / oldCountMinusOne) * newCountMinusOne
	d.PlanLat.SquaredDiffs = (d.PlanLat.SquaredDiffs / oldCountMinusOne) * newCountMinusOne
	d.RunLat.SquaredDiffs = (d.RunLat.SquaredDiffs / oldCountMinusOne) * newCountMinusOne
	d.ServiceLat.SquaredDiffs = (d.ServiceLat.SquaredDiffs / oldCountMinusOne) * newCountMinusOne
	d.OverheadLat.SquaredDiffs = (d.OverheadLat.SquaredDiffs / oldCountMinusOne) * newCountMinusOne

	d.MaxRetries = telemetry.Bucket10(d.MaxRetries)

	d.FirstAttemptCount = int64((float64(d.FirstAttemptCount) / float64(oldCount)) * float64(newCount))
}

func scrubStmtStatKey(vt VirtualTabler, key string) (string, bool) {
	__antithesis_instrumentation__.Notify(470801)

	stmt, err := parser.ParseOne(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(470803)
		return "", false
	} else {
		__antithesis_instrumentation__.Notify(470804)
	}
	__antithesis_instrumentation__.Notify(470802)

	f := tree.NewFmtCtx(
		tree.FmtAnonymize,
		tree.FmtReformatTableNames(hideNonVirtualTableNameFunc(vt)),
	)
	f.FormatNode(stmt.AST)
	return f.CloseAndGetString(), true
}

func formatStmtKeyAsRedactableString(
	vt VirtualTabler, rootAST tree.Statement, ann *tree.Annotations, fs tree.FmtFlags,
) redact.RedactableString {
	__antithesis_instrumentation__.Notify(470805)
	f := tree.NewFmtCtx(
		tree.FmtAlwaysQualifyTableNames|tree.FmtMarkRedactionNode|fs,
		tree.FmtAnnotations(ann),
		tree.FmtReformatTableNames(hideNonVirtualTableNameFunc(vt)))
	f.FormatNode(rootAST)
	formattedRedactableStatementString := f.CloseAndGetString()
	return redact.RedactableString(formattedRedactableStatementString)
}

const FailedHashedValue = "unknown"

func HashForReporting(secret, appName string) string {
	__antithesis_instrumentation__.Notify(470806)

	if len(secret) == 0 {
		__antithesis_instrumentation__.Notify(470809)
		return FailedHashedValue
	} else {
		__antithesis_instrumentation__.Notify(470810)
	}
	__antithesis_instrumentation__.Notify(470807)
	hash := hmac.New(sha256.New, []byte(secret))
	if _, err := hash.Write([]byte(appName)); err != nil {
		__antithesis_instrumentation__.Notify(470811)
		panic(errors.NewAssertionErrorWithWrappedErrf(err,
			`"It never returns an error." -- https://golang.org/pkg/hash`))
	} else {
		__antithesis_instrumentation__.Notify(470812)
	}
	__antithesis_instrumentation__.Notify(470808)
	return hex.EncodeToString(hash.Sum(nil)[:4])
}

func formatStatementHideConstants(ast tree.Statement) string {
	__antithesis_instrumentation__.Notify(470813)
	if ast == nil {
		__antithesis_instrumentation__.Notify(470815)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(470816)
	}
	__antithesis_instrumentation__.Notify(470814)
	return tree.AsStringWithFlags(ast, tree.FmtHideConstants)
}

func formatStatementSummary(ast tree.Statement) string {
	__antithesis_instrumentation__.Notify(470817)
	if ast == nil {
		__antithesis_instrumentation__.Notify(470819)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(470820)
	}
	__antithesis_instrumentation__.Notify(470818)
	fmtFlags := tree.FmtSummary | tree.FmtHideConstants
	return tree.AsStringWithFlags(ast, fmtFlags)
}

func DescsTxn(
	ctx context.Context,
	execCfg *ExecutorConfig,
	f func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error,
) error {
	__antithesis_instrumentation__.Notify(470821)
	return execCfg.CollectionFactory.Txn(ctx, execCfg.InternalExecutor, execCfg.DB, f)
}

func TestingDescsTxn(
	ctx context.Context,
	s serverutils.TestServerInterface,
	f func(ctx context.Context, txn *kv.Txn, col *descs.Collection) error,
) error {
	__antithesis_instrumentation__.Notify(470822)
	execCfg := s.ExecutorConfig().(ExecutorConfig)
	return DescsTxn(ctx, &execCfg, f)
}

func NewRowMetrics(internal bool) row.Metrics {
	__antithesis_instrumentation__.Notify(470823)
	return row.Metrics{
		MaxRowSizeLogCount: metric.NewCounter(getMetricMeta(row.MetaMaxRowSizeLog, internal)),
		MaxRowSizeErrCount: metric.NewCounter(getMetricMeta(row.MetaMaxRowSizeErr, internal)),
	}
}

func (cfg *ExecutorConfig) GetRowMetrics(internal bool) *row.Metrics {
	__antithesis_instrumentation__.Notify(470824)
	if internal {
		__antithesis_instrumentation__.Notify(470826)
		return cfg.InternalRowMetrics
	} else {
		__antithesis_instrumentation__.Notify(470827)
	}
	__antithesis_instrumentation__.Notify(470825)
	return cfg.RowMetrics
}
