package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/idxusage"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats/sqlstatsutil"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/sslocal"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/collector"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/errors"
)

const CrdbInternalName = catconstants.CRDBInternalSchemaName

var crdbInternal = virtualSchema{
	name: CrdbInternalName,
	tableDefs: map[descpb.ID]virtualSchemaDef{
		catconstants.CrdbInternalBackwardDependenciesTableID:        crdbInternalBackwardDependenciesTable,
		catconstants.CrdbInternalBuildInfoTableID:                   crdbInternalBuildInfoTable,
		catconstants.CrdbInternalBuiltinFunctionsTableID:            crdbInternalBuiltinFunctionsTable,
		catconstants.CrdbInternalClusterContendedIndexesViewID:      crdbInternalClusterContendedIndexesView,
		catconstants.CrdbInternalClusterContendedKeysViewID:         crdbInternalClusterContendedKeysView,
		catconstants.CrdbInternalClusterContendedTablesViewID:       crdbInternalClusterContendedTablesView,
		catconstants.CrdbInternalClusterContentionEventsTableID:     crdbInternalClusterContentionEventsTable,
		catconstants.CrdbInternalClusterDistSQLFlowsTableID:         crdbInternalClusterDistSQLFlowsTable,
		catconstants.CrdbInternalClusterLocksTableID:                crdbInternalClusterLocksTable,
		catconstants.CrdbInternalClusterQueriesTableID:              crdbInternalClusterQueriesTable,
		catconstants.CrdbInternalClusterTransactionsTableID:         crdbInternalClusterTxnsTable,
		catconstants.CrdbInternalClusterSessionsTableID:             crdbInternalClusterSessionsTable,
		catconstants.CrdbInternalClusterSettingsTableID:             crdbInternalClusterSettingsTable,
		catconstants.CrdbInternalClusterStmtStatsTableID:            crdbInternalClusterStmtStatsTable,
		catconstants.CrdbInternalCreateSchemaStmtsTableID:           crdbInternalCreateSchemaStmtsTable,
		catconstants.CrdbInternalCreateStmtsTableID:                 crdbInternalCreateStmtsTable,
		catconstants.CrdbInternalCreateTypeStmtsTableID:             crdbInternalCreateTypeStmtsTable,
		catconstants.CrdbInternalDatabasesTableID:                   crdbInternalDatabasesTable,
		catconstants.CrdbInternalSuperRegions:                       crdbInternalSuperRegions,
		catconstants.CrdbInternalFeatureUsageID:                     crdbInternalFeatureUsage,
		catconstants.CrdbInternalForwardDependenciesTableID:         crdbInternalForwardDependenciesTable,
		catconstants.CrdbInternalGossipNodesTableID:                 crdbInternalGossipNodesTable,
		catconstants.CrdbInternalKVNodeLivenessTableID:              crdbInternalKVNodeLivenessTable,
		catconstants.CrdbInternalGossipAlertsTableID:                crdbInternalGossipAlertsTable,
		catconstants.CrdbInternalGossipLivenessTableID:              crdbInternalGossipLivenessTable,
		catconstants.CrdbInternalGossipNetworkTableID:               crdbInternalGossipNetworkTable,
		catconstants.CrdbInternalTransactionContentionEvents:        crdbInternalTransactionContentionEventsTable,
		catconstants.CrdbInternalIndexColumnsTableID:                crdbInternalIndexColumnsTable,
		catconstants.CrdbInternalIndexUsageStatisticsTableID:        crdbInternalIndexUsageStatistics,
		catconstants.CrdbInternalInflightTraceSpanTableID:           crdbInternalInflightTraceSpanTable,
		catconstants.CrdbInternalJobsTableID:                        crdbInternalJobsTable,
		catconstants.CrdbInternalKVNodeStatusTableID:                crdbInternalKVNodeStatusTable,
		catconstants.CrdbInternalKVStoreStatusTableID:               crdbInternalKVStoreStatusTable,
		catconstants.CrdbInternalLeasesTableID:                      crdbInternalLeasesTable,
		catconstants.CrdbInternalLocalContentionEventsTableID:       crdbInternalLocalContentionEventsTable,
		catconstants.CrdbInternalLocalDistSQLFlowsTableID:           crdbInternalLocalDistSQLFlowsTable,
		catconstants.CrdbInternalLocalQueriesTableID:                crdbInternalLocalQueriesTable,
		catconstants.CrdbInternalLocalTransactionsTableID:           crdbInternalLocalTxnsTable,
		catconstants.CrdbInternalLocalSessionsTableID:               crdbInternalLocalSessionsTable,
		catconstants.CrdbInternalLocalMetricsTableID:                crdbInternalLocalMetricsTable,
		catconstants.CrdbInternalNodeStmtStatsTableID:               crdbInternalNodeStmtStatsTable,
		catconstants.CrdbInternalNodeTxnStatsTableID:                crdbInternalNodeTxnStatsTable,
		catconstants.CrdbInternalPartitionsTableID:                  crdbInternalPartitionsTable,
		catconstants.CrdbInternalPredefinedCommentsTableID:          crdbInternalPredefinedCommentsTable,
		catconstants.CrdbInternalRangesNoLeasesTableID:              crdbInternalRangesNoLeasesTable,
		catconstants.CrdbInternalRangesViewID:                       crdbInternalRangesView,
		catconstants.CrdbInternalRuntimeInfoTableID:                 crdbInternalRuntimeInfoTable,
		catconstants.CrdbInternalSchemaChangesTableID:               crdbInternalSchemaChangesTable,
		catconstants.CrdbInternalSessionTraceTableID:                crdbInternalSessionTraceTable,
		catconstants.CrdbInternalSessionVariablesTableID:            crdbInternalSessionVariablesTable,
		catconstants.CrdbInternalStmtStatsTableID:                   crdbInternalStmtStatsView,
		catconstants.CrdbInternalTableColumnsTableID:                crdbInternalTableColumnsTable,
		catconstants.CrdbInternalTableIndexesTableID:                crdbInternalTableIndexesTable,
		catconstants.CrdbInternalTablesTableLastStatsID:             crdbInternalTablesTableLastStats,
		catconstants.CrdbInternalTablesTableID:                      crdbInternalTablesTable,
		catconstants.CrdbInternalClusterTxnStatsTableID:             crdbInternalClusterTxnStatsTable,
		catconstants.CrdbInternalTxnStatsTableID:                    crdbInternalTxnStatsView,
		catconstants.CrdbInternalTransactionStatsTableID:            crdbInternalTransactionStatisticsTable,
		catconstants.CrdbInternalZonesTableID:                       crdbInternalZonesTable,
		catconstants.CrdbInternalInvalidDescriptorsTableID:          crdbInternalInvalidDescriptorsTable,
		catconstants.CrdbInternalClusterDatabasePrivilegesTableID:   crdbInternalClusterDatabasePrivilegesTable,
		catconstants.CrdbInternalCrossDbRefrences:                   crdbInternalCrossDbReferences,
		catconstants.CrdbInternalLostTableDescriptors:               crdbLostTableDescriptors,
		catconstants.CrdbInternalClusterInflightTracesTable:         crdbInternalClusterInflightTracesTable,
		catconstants.CrdbInternalRegionsTable:                       crdbInternalRegionsTable,
		catconstants.CrdbInternalDefaultPrivilegesTable:             crdbInternalDefaultPrivilegesTable,
		catconstants.CrdbInternalActiveRangeFeedsTable:              crdbInternalActiveRangeFeedsTable,
		catconstants.CrdbInternalTenantUsageDetailsViewID:           crdbInternalTenantUsageDetailsView,
		catconstants.CrdbInternalPgCatalogTableIsImplementedTableID: crdbInternalPgCatalogTableIsImplementedTable,
	},
	validWithNoDatabaseContext: true,
}

var crdbInternalBuildInfoTable = virtualSchemaTable{
	comment: `detailed identification strings (RAM, local node only)`,
	schema: `
CREATE TABLE crdb_internal.node_build_info (
  node_id INT NOT NULL,
  field   STRING NOT NULL,
  value   STRING NOT NULL
)`,
	populate: func(_ context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461102)
		execCfg := p.ExecCfg()
		nodeID, _ := execCfg.NodeID.OptionalNodeID()

		info := build.GetInfo()
		for k, v := range map[string]string{
			"Name":         "CockroachDB",
			"ClusterID":    execCfg.LogicalClusterID().String(),
			"Organization": execCfg.Organization(),
			"Build":        info.Short(),
			"Version":      info.Tag,
			"Channel":      info.Channel,
		} {
			__antithesis_instrumentation__.Notify(461104)
			if err := addRow(
				tree.NewDInt(tree.DInt(nodeID)),
				tree.NewDString(k),
				tree.NewDString(v),
			); err != nil {
				__antithesis_instrumentation__.Notify(461105)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461106)
			}
		}
		__antithesis_instrumentation__.Notify(461103)
		return nil
	},
}

var crdbInternalRuntimeInfoTable = virtualSchemaTable{
	comment: `server parameters, useful to construct connection URLs (RAM, local node only)`,
	schema: `
CREATE TABLE crdb_internal.node_runtime_info (
  node_id   INT NOT NULL,
  component STRING NOT NULL,
  field     STRING NOT NULL,
  value     STRING NOT NULL
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461107)
		if err := p.RequireAdminRole(ctx, "access the node runtime information"); err != nil {
			__antithesis_instrumentation__.Notify(461111)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461112)
		}
		__antithesis_instrumentation__.Notify(461108)

		node := p.ExecCfg().NodeInfo

		nodeID, _ := node.NodeID.OptionalNodeID()
		dbURL, err := node.PGURL(url.User(security.RootUser))
		if err != nil {
			__antithesis_instrumentation__.Notify(461113)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461114)
		}
		__antithesis_instrumentation__.Notify(461109)

		for _, item := range []struct {
			component string
			url       *url.URL
		}{
			{"DB", dbURL.ToPQ()}, {"UI", node.AdminURL()},
		} {
			__antithesis_instrumentation__.Notify(461115)
			var user string
			if item.url.User != nil {
				__antithesis_instrumentation__.Notify(461118)
				user = item.url.User.String()
			} else {
				__antithesis_instrumentation__.Notify(461119)
			}
			__antithesis_instrumentation__.Notify(461116)
			host, port, err := net.SplitHostPort(item.url.Host)
			if err != nil {
				__antithesis_instrumentation__.Notify(461120)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461121)
			}
			__antithesis_instrumentation__.Notify(461117)
			for _, kv := range [][2]string{
				{"URL", item.url.String()},
				{"Scheme", item.url.Scheme},
				{"User", user},
				{"Host", host},
				{"Port", port},
				{"URI", item.url.RequestURI()},
			} {
				__antithesis_instrumentation__.Notify(461122)
				k, v := kv[0], kv[1]
				if err := addRow(
					tree.NewDInt(tree.DInt(nodeID)),
					tree.NewDString(item.component),
					tree.NewDString(k),
					tree.NewDString(v),
				); err != nil {
					__antithesis_instrumentation__.Notify(461123)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461124)
				}
			}
		}
		__antithesis_instrumentation__.Notify(461110)
		return nil
	},
}

var crdbInternalDatabasesTable = virtualSchemaTable{
	comment: `databases accessible by the current user (KV scan)`,
	schema: `
CREATE TABLE crdb_internal.databases (
	id INT NOT NULL,
	name STRING NOT NULL,
	owner NAME NOT NULL,
	primary_region STRING,
	regions STRING[],
	survival_goal STRING,
	placement_policy STRING,
	create_statement STRING NOT NULL
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461125)
		return forEachDatabaseDesc(ctx, p, nil, true,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(461126)
				var survivalGoal tree.Datum = tree.DNull
				var primaryRegion tree.Datum = tree.DNull
				var placement tree.Datum = tree.DNull
				regions := tree.NewDArray(types.String)

				createNode := tree.CreateDatabase{}
				createNode.ConnectionLimit = -1
				createNode.Name = tree.Name(db.GetName())
				if db.IsMultiRegion() {
					__antithesis_instrumentation__.Notify(461128)
					primaryRegion = tree.NewDString(string(db.GetRegionConfig().PrimaryRegion))

					createNode.PrimaryRegion = tree.Name(db.GetRegionConfig().PrimaryRegion)

					regionConfig, err := SynthesizeRegionConfig(ctx, p.txn, db.GetID(), p.Descriptors())
					if err != nil {
						__antithesis_instrumentation__.Notify(461132)
						return err
					} else {
						__antithesis_instrumentation__.Notify(461133)
					}
					__antithesis_instrumentation__.Notify(461129)

					createNode.Regions = make(tree.NameList, len(regionConfig.Regions()))
					for i, region := range regionConfig.Regions() {
						__antithesis_instrumentation__.Notify(461134)
						if err := regions.Append(tree.NewDString(string(region))); err != nil {
							__antithesis_instrumentation__.Notify(461136)
							return err
						} else {
							__antithesis_instrumentation__.Notify(461137)
						}
						__antithesis_instrumentation__.Notify(461135)
						createNode.Regions[i] = tree.Name(region)
					}
					__antithesis_instrumentation__.Notify(461130)

					if db.GetRegionConfig().Placement == descpb.DataPlacement_RESTRICTED {
						__antithesis_instrumentation__.Notify(461138)
						placement = tree.NewDString("restricted")
						createNode.Placement = tree.DataPlacementRestricted
					} else {
						__antithesis_instrumentation__.Notify(461139)
						placement = tree.NewDString("default")

						createNode.Placement = tree.DataPlacementUnspecified
					}
					__antithesis_instrumentation__.Notify(461131)

					createNode.SurvivalGoal = tree.SurvivalGoalDefault
					switch db.GetRegionConfig().SurvivalGoal {
					case descpb.SurvivalGoal_ZONE_FAILURE:
						__antithesis_instrumentation__.Notify(461140)
						survivalGoal = tree.NewDString("zone")
						createNode.SurvivalGoal = tree.SurvivalGoalZoneFailure
					case descpb.SurvivalGoal_REGION_FAILURE:
						__antithesis_instrumentation__.Notify(461141)
						survivalGoal = tree.NewDString("region")
						createNode.SurvivalGoal = tree.SurvivalGoalRegionFailure
					default:
						__antithesis_instrumentation__.Notify(461142)
						return errors.Newf("unknown survival goal: %d", db.GetRegionConfig().SurvivalGoal)
					}
				} else {
					__antithesis_instrumentation__.Notify(461143)
				}
				__antithesis_instrumentation__.Notify(461127)

				return addRow(
					tree.NewDInt(tree.DInt(db.GetID())),
					tree.NewDString(db.GetName()),
					tree.NewDName(getOwnerOfDesc(db).Normalized()),
					primaryRegion,
					regions,
					survivalGoal,
					placement,
					tree.NewDString(createNode.String()),
				)
			})
	},
}

var crdbInternalSuperRegions = virtualSchemaTable{
	comment: `list super regions of databases visible to the current user`,
	schema: `
CREATE TABLE crdb_internal.super_regions (
	id INT NOT NULL,
	database_name STRING NOT NULL,
  super_region_name STRING NOT NULL,
	regions STRING[]
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461144)
		return forEachDatabaseDesc(ctx, p, nil, true,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(461145)
				if !db.IsMultiRegion() {
					__antithesis_instrumentation__.Notify(461151)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(461152)
				}
				__antithesis_instrumentation__.Notify(461146)

				typeID, err := db.MultiRegionEnumID()
				if err != nil {
					__antithesis_instrumentation__.Notify(461153)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461154)
				}
				__antithesis_instrumentation__.Notify(461147)
				typeDesc, err := p.Descriptors().GetImmutableTypeByID(ctx, p.txn, typeID,
					tree.ObjectLookupFlags{CommonLookupFlags: tree.CommonLookupFlags{
						Required: true,
					}},
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(461155)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461156)
				}
				__antithesis_instrumentation__.Notify(461148)

				superRegions, err := typeDesc.SuperRegions()
				if err != nil {
					__antithesis_instrumentation__.Notify(461157)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461158)
				}
				__antithesis_instrumentation__.Notify(461149)
				for _, superRegion := range superRegions {
					__antithesis_instrumentation__.Notify(461159)
					regionList := tree.NewDArray(types.String)
					for _, region := range superRegion.Regions {
						__antithesis_instrumentation__.Notify(461161)
						if err := regionList.Append(tree.NewDString(region.String())); err != nil {
							__antithesis_instrumentation__.Notify(461162)
							return err
						} else {
							__antithesis_instrumentation__.Notify(461163)
						}
					}
					__antithesis_instrumentation__.Notify(461160)

					if err := addRow(
						tree.NewDInt(tree.DInt(db.GetID())),
						tree.NewDString(db.GetName()),
						tree.NewDString(superRegion.SuperRegionName),
						regionList,
					); err != nil {
						__antithesis_instrumentation__.Notify(461164)
						return err
					} else {
						__antithesis_instrumentation__.Notify(461165)
					}
				}
				__antithesis_instrumentation__.Notify(461150)
				return nil
			})
	},
}

var crdbInternalTablesTable = virtualSchemaTable{
	comment: `table descriptors accessible by current user, including non-public and virtual (KV scan; expensive!)`,
	schema: `
CREATE TABLE crdb_internal.tables (
  table_id                 INT NOT NULL,
  parent_id                INT NOT NULL,
  name                     STRING NOT NULL,
  database_name            STRING,
  version                  INT NOT NULL,
  mod_time                 TIMESTAMP NOT NULL,
  mod_time_logical         DECIMAL NOT NULL,
  format_version           STRING NOT NULL,
  state                    STRING NOT NULL,
  sc_lease_node_id         INT,
  sc_lease_expiration_time TIMESTAMP,
  drop_time                TIMESTAMP,
  audit_mode               STRING NOT NULL,
  schema_name              STRING NOT NULL,
  parent_schema_id         INT NOT NULL,
  locality                 TEXT
)`,
	generator: func(ctx context.Context, p *planner, dbDesc catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(461166)
		row := make(tree.Datums, 14)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(461168)
			all, err := p.Descriptors().GetAllDescriptors(ctx, p.txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(461174)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461175)
			}
			__antithesis_instrumentation__.Notify(461169)
			descs := all.OrderedDescriptors()
			dbNames := make(map[descpb.ID]string)
			scNames := make(map[descpb.ID]string)

			scNames[keys.PublicSchemaID] = catconstants.PublicSchemaName

			for _, desc := range descs {
				__antithesis_instrumentation__.Notify(461176)
				if dbDesc, ok := desc.(catalog.DatabaseDescriptor); ok {
					__antithesis_instrumentation__.Notify(461178)
					dbNames[dbDesc.GetID()] = dbDesc.GetName()
				} else {
					__antithesis_instrumentation__.Notify(461179)
				}
				__antithesis_instrumentation__.Notify(461177)
				if scDesc, ok := desc.(catalog.SchemaDescriptor); ok {
					__antithesis_instrumentation__.Notify(461180)
					scNames[scDesc.GetID()] = scDesc.GetName()
				} else {
					__antithesis_instrumentation__.Notify(461181)
				}
			}
			__antithesis_instrumentation__.Notify(461170)

			addDesc := func(table catalog.TableDescriptor, dbName tree.Datum, scName string) error {
				__antithesis_instrumentation__.Notify(461182)
				leaseNodeDatum := tree.DNull
				leaseExpDatum := tree.DNull
				if lease := table.GetLease(); lease != nil {
					__antithesis_instrumentation__.Notify(461186)
					leaseNodeDatum = tree.NewDInt(tree.DInt(int64(lease.NodeID)))
					leaseExpDatum, err = tree.MakeDTimestamp(
						timeutil.Unix(0, lease.ExpirationTime), time.Nanosecond,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(461187)
						return err
					} else {
						__antithesis_instrumentation__.Notify(461188)
					}
				} else {
					__antithesis_instrumentation__.Notify(461189)
				}
				__antithesis_instrumentation__.Notify(461183)
				dropTimeDatum := tree.DNull
				if dropTime := table.GetDropTime(); dropTime != 0 {
					__antithesis_instrumentation__.Notify(461190)
					dropTimeDatum, err = tree.MakeDTimestamp(
						timeutil.Unix(0, dropTime), time.Nanosecond,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(461191)
						return err
					} else {
						__antithesis_instrumentation__.Notify(461192)
					}
				} else {
					__antithesis_instrumentation__.Notify(461193)
				}
				__antithesis_instrumentation__.Notify(461184)
				locality := tree.DNull
				if c := table.GetLocalityConfig(); c != nil {
					__antithesis_instrumentation__.Notify(461194)
					f := p.EvalContext().FmtCtx(tree.FmtSimple)
					if err := multiregion.FormatTableLocalityConfig(c, f); err != nil {
						__antithesis_instrumentation__.Notify(461196)
						return err
					} else {
						__antithesis_instrumentation__.Notify(461197)
					}
					__antithesis_instrumentation__.Notify(461195)
					locality = tree.NewDString(f.String())
				} else {
					__antithesis_instrumentation__.Notify(461198)
				}
				__antithesis_instrumentation__.Notify(461185)
				row = row[:0]
				row = append(row,
					tree.NewDInt(tree.DInt(int64(table.GetID()))),
					tree.NewDInt(tree.DInt(int64(table.GetParentID()))),
					tree.NewDString(table.GetName()),
					dbName,
					tree.NewDInt(tree.DInt(int64(table.GetVersion()))),
					tree.TimestampToInexactDTimestamp(table.GetModificationTime()),
					tree.TimestampToDecimalDatum(table.GetModificationTime()),
					tree.NewDString(table.GetFormatVersion().String()),
					tree.NewDString(table.GetState().String()),
					leaseNodeDatum,
					leaseExpDatum,
					dropTimeDatum,
					tree.NewDString(table.GetAuditMode().String()),
					tree.NewDString(scName),
					tree.NewDInt(tree.DInt(int64(table.GetParentSchemaID()))),
					locality,
				)
				return pusher.pushRow(row...)
			}
			__antithesis_instrumentation__.Notify(461171)

			for _, desc := range descs {
				__antithesis_instrumentation__.Notify(461199)
				table, ok := desc.(catalog.TableDescriptor)
				if !ok || func() bool {
					__antithesis_instrumentation__.Notify(461203)
					return p.CheckAnyPrivilege(ctx, table) != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(461204)
					continue
				} else {
					__antithesis_instrumentation__.Notify(461205)
				}
				__antithesis_instrumentation__.Notify(461200)
				dbName := dbNames[table.GetParentID()]
				if dbName == "" {
					__antithesis_instrumentation__.Notify(461206)

					dbName = fmt.Sprintf("[%d]", table.GetParentID())
				} else {
					__antithesis_instrumentation__.Notify(461207)
				}
				__antithesis_instrumentation__.Notify(461201)
				schemaName := scNames[table.GetParentSchemaID()]
				if schemaName == "" {
					__antithesis_instrumentation__.Notify(461208)

					schemaName = fmt.Sprintf("[%d]", table.GetParentSchemaID())
				} else {
					__antithesis_instrumentation__.Notify(461209)
				}
				__antithesis_instrumentation__.Notify(461202)
				if err := addDesc(table, tree.NewDString(dbName), schemaName); err != nil {
					__antithesis_instrumentation__.Notify(461210)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461211)
				}
			}
			__antithesis_instrumentation__.Notify(461172)

			vt := p.getVirtualTabler()
			vSchemas := vt.getSchemas()
			for _, virtSchemaName := range vt.getSchemaNames() {
				__antithesis_instrumentation__.Notify(461212)
				e := vSchemas[virtSchemaName]
				for _, tName := range e.orderedDefNames {
					__antithesis_instrumentation__.Notify(461213)
					vTableEntry := e.defs[tName]
					if err := addDesc(vTableEntry.desc, tree.DNull, virtSchemaName); err != nil {
						__antithesis_instrumentation__.Notify(461214)
						return err
					} else {
						__antithesis_instrumentation__.Notify(461215)
					}
				}
			}
			__antithesis_instrumentation__.Notify(461173)
			return nil
		}
		__antithesis_instrumentation__.Notify(461167)
		return setupGenerator(ctx, worker, stopper)
	},
}

var crdbInternalPgCatalogTableIsImplementedTable = virtualSchemaTable{
	comment: `table descriptors accessible by current user, including non-public and virtual (KV scan; expensive!)`,
	schema: `
CREATE TABLE crdb_internal.pg_catalog_table_is_implemented (
  name                     STRING NOT NULL,
  implemented              BOOL
)`,
	generator: func(ctx context.Context, p *planner, dbDesc catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(461216)
		row := make(tree.Datums, 14)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(461218)
			addDesc := func(table *virtualDefEntry, dbName tree.Datum, scName string) error {
				__antithesis_instrumentation__.Notify(461221)
				tableDesc := table.desc
				row = row[:0]
				row = append(row,
					tree.NewDString(tableDesc.GetName()),
					tree.MakeDBool(tree.DBool(table.unimplemented)),
				)
				return pusher.pushRow(row...)
			}
			__antithesis_instrumentation__.Notify(461219)
			vt := p.getVirtualTabler()
			vSchemas := vt.getSchemas()
			e := vSchemas["pg_catalog"]
			for _, tName := range e.orderedDefNames {
				__antithesis_instrumentation__.Notify(461222)
				vTableEntry := e.defs[tName]
				if err := addDesc(vTableEntry, tree.DNull, "pg_catalog"); err != nil {
					__antithesis_instrumentation__.Notify(461223)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461224)
				}
			}
			__antithesis_instrumentation__.Notify(461220)
			return nil
		}
		__antithesis_instrumentation__.Notify(461217)
		return setupGenerator(ctx, worker, stopper)
	},
}

var statsAsOfTimeClusterMode = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.crdb_internal.table_row_statistics.as_of_time",
	"historical query time used to build the crdb_internal.table_row_statistics table",
	-10*time.Second,
)

var crdbInternalTablesTableLastStats = virtualSchemaTable{
	comment: "stats for all tables accessible by current user in current database as of 10s ago",
	schema: `
CREATE TABLE crdb_internal.table_row_statistics (
  table_id                   INT         NOT NULL,
  table_name                 STRING      NOT NULL,
  estimated_row_count        INT
)`,
	populate: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461225)

		query := fmt.Sprintf(`
           SELECT s."tableID", max(s."rowCount")
             FROM system.table_statistics AS s
             JOIN (
                    SELECT "tableID", max("createdAt") AS last_dt
                      FROM system.table_statistics
                     GROUP BY "tableID"
                  ) AS l ON l."tableID" = s."tableID" AND l.last_dt = s."createdAt"
            AS OF SYSTEM TIME '%s'
            GROUP BY s."tableID"`, statsAsOfTimeClusterMode.String(&p.ExecCfg().Settings.SV))
		statRows, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryBufferedEx(
			ctx, "crdb-internal-statistics-table", nil,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			query)
		if err != nil {
			__antithesis_instrumentation__.Notify(461228)

			if errors.Is(err, catalog.ErrDescriptorNotFound) {
				__antithesis_instrumentation__.Notify(461230)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(461231)
			}
			__antithesis_instrumentation__.Notify(461229)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461232)
		}
		__antithesis_instrumentation__.Notify(461226)

		statMap := make(map[tree.DInt]tree.Datum, len(statRows))
		for _, r := range statRows {
			__antithesis_instrumentation__.Notify(461233)
			statMap[tree.MustBeDInt(r[0])] = r[1]
		}
		__antithesis_instrumentation__.Notify(461227)

		return forEachTableDescAll(ctx, p, db, virtualMany,
			func(db catalog.DatabaseDescriptor, _ string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(461234)
				tableID := tree.DInt(table.GetID())
				rowCount := tree.DNull

				if !table.IsVirtualTable() {
					__antithesis_instrumentation__.Notify(461236)
					rowCount = tree.NewDInt(0)
					if cnt, ok := statMap[tableID]; ok {
						__antithesis_instrumentation__.Notify(461237)
						rowCount = cnt
					} else {
						__antithesis_instrumentation__.Notify(461238)
					}
				} else {
					__antithesis_instrumentation__.Notify(461239)
				}
				__antithesis_instrumentation__.Notify(461235)
				return addRow(
					tree.NewDInt(tableID),
					tree.NewDString(table.GetName()),
					rowCount,
				)
			},
		)
	},
}

var crdbInternalSchemaChangesTable = virtualSchemaTable{
	comment: `ongoing schema changes, across all descriptors accessible by current user (KV scan; expensive!)`,
	schema: `
CREATE TABLE crdb_internal.schema_changes (
  table_id      INT NOT NULL,
  parent_id     INT NOT NULL,
  name          STRING NOT NULL,
  type          STRING NOT NULL,
  target_id     INT,
  target_name   STRING,
  state         STRING NOT NULL,
  direction     STRING NOT NULL
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461240)
		all, err := p.Descriptors().GetAllDescriptors(ctx, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(461243)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461244)
		}
		__antithesis_instrumentation__.Notify(461241)
		descs := all.OrderedDescriptors()

		for _, desc := range descs {
			__antithesis_instrumentation__.Notify(461245)
			table, ok := desc.(catalog.TableDescriptor)
			if !ok || func() bool {
				__antithesis_instrumentation__.Notify(461247)
				return p.CheckAnyPrivilege(ctx, table) != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(461248)
				continue
			} else {
				__antithesis_instrumentation__.Notify(461249)
			}
			__antithesis_instrumentation__.Notify(461246)
			tableID := tree.NewDInt(tree.DInt(int64(table.GetID())))
			parentID := tree.NewDInt(tree.DInt(int64(table.GetParentID())))
			tableName := tree.NewDString(table.GetName())
			for _, mut := range table.TableDesc().Mutations {
				__antithesis_instrumentation__.Notify(461250)
				mutType := "UNKNOWN"
				targetID := tree.DNull
				targetName := tree.DNull
				switch d := mut.Descriptor_.(type) {
				case *descpb.DescriptorMutation_Column:
					__antithesis_instrumentation__.Notify(461252)
					mutType = "COLUMN"
					targetID = tree.NewDInt(tree.DInt(int64(d.Column.ID)))
					targetName = tree.NewDString(d.Column.Name)
				case *descpb.DescriptorMutation_Index:
					__antithesis_instrumentation__.Notify(461253)
					mutType = "INDEX"
					targetID = tree.NewDInt(tree.DInt(int64(d.Index.ID)))
					targetName = tree.NewDString(d.Index.Name)
				case *descpb.DescriptorMutation_Constraint:
					__antithesis_instrumentation__.Notify(461254)
					mutType = "CONSTRAINT VALIDATION"
					targetName = tree.NewDString(d.Constraint.Name)
				}
				__antithesis_instrumentation__.Notify(461251)
				if err := addRow(
					tableID,
					parentID,
					tableName,
					tree.NewDString(mutType),
					targetID,
					targetName,
					tree.NewDString(mut.State.String()),
					tree.NewDString(mut.Direction.String()),
				); err != nil {
					__antithesis_instrumentation__.Notify(461255)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461256)
				}
			}
		}
		__antithesis_instrumentation__.Notify(461242)
		return nil
	},
}

var crdbInternalLeasesTable = virtualSchemaTable{
	comment: `acquired table leases (RAM; local node only)`,
	schema: `
CREATE TABLE crdb_internal.leases (
  node_id     INT NOT NULL,
  table_id    INT NOT NULL,
  name        STRING NOT NULL,
  parent_id   INT NOT NULL,
  expiration  TIMESTAMP NOT NULL,
  deleted     BOOL NOT NULL
)`,
	populate: func(
		ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error,
	) (err error) {
		__antithesis_instrumentation__.Notify(461257)
		nodeID, _ := p.execCfg.NodeID.OptionalNodeID()
		p.LeaseMgr().VisitLeases(func(desc catalog.Descriptor, takenOffline bool, _ int, expiration tree.DTimestamp) (wantMore bool) {
			__antithesis_instrumentation__.Notify(461259)
			if p.CheckAnyPrivilege(ctx, desc) != nil {
				__antithesis_instrumentation__.Notify(461261)

				return true
			} else {
				__antithesis_instrumentation__.Notify(461262)
			}
			__antithesis_instrumentation__.Notify(461260)

			err = addRow(
				tree.NewDInt(tree.DInt(nodeID)),
				tree.NewDInt(tree.DInt(int64(desc.GetID()))),
				tree.NewDString(desc.GetName()),
				tree.NewDInt(tree.DInt(int64(desc.GetParentID()))),
				&expiration,
				tree.MakeDBool(tree.DBool(takenOffline)),
			)
			return err == nil
		})
		__antithesis_instrumentation__.Notify(461258)
		return err
	},
}

func tsOrNull(micros int64) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(461263)
	if micros == 0 {
		__antithesis_instrumentation__.Notify(461265)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(461266)
	}
	__antithesis_instrumentation__.Notify(461264)
	ts := timeutil.Unix(0, micros*time.Microsecond.Nanoseconds())
	return tree.MakeDTimestamp(ts, time.Microsecond)
}

var crdbInternalJobsTable = virtualSchemaTable{
	schema: `
CREATE TABLE crdb_internal.jobs (
  job_id                INT,
  job_type              STRING,
  description           STRING,
  statement             STRING,
  user_name             STRING,
  descriptor_ids        INT[],
  status                STRING,
  running_status        STRING,
  created               TIMESTAMP,
  started               TIMESTAMP,
  finished              TIMESTAMP,
  modified              TIMESTAMP,
  fraction_completed    FLOAT,
  high_water_timestamp  DECIMAL,
  error                 STRING,
  coordinator_id        INT,
  trace_id              INT,
  last_run              TIMESTAMP,
  next_run              TIMESTAMP,
  num_runs              INT,
  execution_errors      STRING[],
  execution_events      JSONB
)`,
	comment: `decoded job metadata from system.jobs (KV scan)`,
	generator: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, _ *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(461267)
		currentUser := p.SessionData().User()
		isAdmin, err := p.HasAdminRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461273)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(461274)
		}
		__antithesis_instrumentation__.Notify(461268)

		hasControlJob, err := p.HasRoleOption(ctx, roleoption.CONTROLJOB)
		if err != nil {
			__antithesis_instrumentation__.Notify(461275)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(461276)
		}
		__antithesis_instrumentation__.Notify(461269)

		const (
			qSelect          = `SELECT id, status, created, payload, progress, claim_session_id, claim_instance_id`
			qFrom            = ` FROM system.jobs`
			backoffArgs      = `(SELECT $1::FLOAT AS initial_delay, $2::FLOAT AS max_delay) args`
			queryWithBackoff = qSelect + `, last_run, COALESCE(num_runs, 0), ` + jobs.NextRunClause + ` as next_run` + qFrom + ", " + backoffArgs
		)

		it, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryIteratorEx(
			ctx, "crdb-internal-jobs-table", p.txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			queryWithBackoff, p.execCfg.JobRegistry.RetryInitialDelay(), p.execCfg.JobRegistry.RetryMaxDelay())
		if err != nil {
			__antithesis_instrumentation__.Notify(461277)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(461278)
		}
		__antithesis_instrumentation__.Notify(461270)

		cleanup := func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(461279)
			if err := it.Close(); err != nil {
				__antithesis_instrumentation__.Notify(461280)

				log.Warningf(ctx, "error closing an iterator: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(461281)
			}
		}
		__antithesis_instrumentation__.Notify(461271)

		container := make(tree.Datums, 0, 21)
		sessionJobs := make([]*jobs.Record, 0, len(p.extendedEvalCtx.SchemaChangeJobRecords))
		uniqueJobs := make(map[*jobs.Record]struct{})
		for _, job := range p.extendedEvalCtx.SchemaChangeJobRecords {
			__antithesis_instrumentation__.Notify(461282)
			if _, ok := uniqueJobs[job]; ok {
				__antithesis_instrumentation__.Notify(461284)
				continue
			} else {
				__antithesis_instrumentation__.Notify(461285)
			}
			__antithesis_instrumentation__.Notify(461283)
			sessionJobs = append(sessionJobs, job)
			uniqueJobs[job] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(461272)
		return func() (datums tree.Datums, e error) {
			__antithesis_instrumentation__.Notify(461286)

			for {
				__antithesis_instrumentation__.Notify(461287)
				ok, err := it.Next(ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(461296)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(461297)
				}
				__antithesis_instrumentation__.Notify(461288)
				var id, status, created, payloadBytes, progressBytes, sessionIDBytes,
					instanceID tree.Datum
				lastRun, nextRun, numRuns := tree.DNull, tree.DNull, tree.DNull
				if ok {
					__antithesis_instrumentation__.Notify(461298)
					r := it.Cur()
					id, status, created, payloadBytes, progressBytes, sessionIDBytes, instanceID =
						r[0], r[1], r[2], r[3], r[4], r[5], r[6]
					lastRun, numRuns, nextRun = r[7], r[8], r[9]
				} else {
					__antithesis_instrumentation__.Notify(461299)
					if !ok {
						__antithesis_instrumentation__.Notify(461300)
						if len(sessionJobs) == 0 {
							__antithesis_instrumentation__.Notify(461303)
							return nil, nil
						} else {
							__antithesis_instrumentation__.Notify(461304)
						}
						__antithesis_instrumentation__.Notify(461301)
						job := sessionJobs[len(sessionJobs)-1]
						sessionJobs = sessionJobs[:len(sessionJobs)-1]

						id = tree.NewDInt(tree.DInt(job.JobID))
						status = tree.NewDString(string(jobs.StatusPending))
						created = tree.TimestampToInexactDTimestamp(p.txn.ReadTimestamp())
						progressBytes, payloadBytes, err = getPayloadAndProgressFromJobsRecord(p, job)
						if err != nil {
							__antithesis_instrumentation__.Notify(461305)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(461306)
						}
						__antithesis_instrumentation__.Notify(461302)
						sessionIDBytes = tree.NewDBytes(tree.DBytes(p.extendedEvalCtx.SessionID.GetBytes()))
						instanceID = tree.NewDInt(tree.DInt(p.extendedEvalCtx.ExecCfg.JobRegistry.ID()))
					} else {
						__antithesis_instrumentation__.Notify(461307)
					}
				}
				__antithesis_instrumentation__.Notify(461289)

				var jobType, description, statement, username, descriptorIDs, started, runningStatus,
					finished, modified, fractionCompleted, highWaterTimestamp, errorStr, coordinatorID,
					traceID, executionErrors, executionEvents = tree.DNull, tree.DNull, tree.DNull,
					tree.DNull, tree.DNull, tree.DNull, tree.DNull, tree.DNull, tree.DNull, tree.DNull,
					tree.DNull, tree.DNull, tree.DNull, tree.DNull, tree.DNull, tree.DNull

				payload, err := jobs.UnmarshalPayload(payloadBytes)

				ownedByAdmin := false
				var sqlUsername security.SQLUsername
				if payload != nil {
					__antithesis_instrumentation__.Notify(461308)
					sqlUsername = payload.UsernameProto.Decode()
					ownedByAdmin, err = p.UserHasAdminRole(ctx, sqlUsername)
					if err != nil {
						__antithesis_instrumentation__.Notify(461309)
						errorStr = tree.NewDString(fmt.Sprintf("error decoding payload: %v", err))
					} else {
						__antithesis_instrumentation__.Notify(461310)
					}
				} else {
					__antithesis_instrumentation__.Notify(461311)
				}
				__antithesis_instrumentation__.Notify(461290)
				if sessionID, ok := sessionIDBytes.(*tree.DBytes); ok {
					__antithesis_instrumentation__.Notify(461312)
					if isAlive, err := p.EvalContext().SQLLivenessReader.IsAlive(
						ctx, sqlliveness.SessionID(*sessionID),
					); err != nil {
						__antithesis_instrumentation__.Notify(461313)

					} else {
						__antithesis_instrumentation__.Notify(461314)
						if instanceID, ok := instanceID.(*tree.DInt); ok && func() bool {
							__antithesis_instrumentation__.Notify(461315)
							return isAlive == true
						}() == true {
							__antithesis_instrumentation__.Notify(461316)
							coordinatorID = instanceID
						} else {
							__antithesis_instrumentation__.Notify(461317)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(461318)
				}
				__antithesis_instrumentation__.Notify(461291)

				sameUser := payload != nil && func() bool {
					__antithesis_instrumentation__.Notify(461319)
					return sqlUsername == currentUser == true
				}() == true

				if canAccess := isAdmin || func() bool {
					__antithesis_instrumentation__.Notify(461320)
					return (!ownedByAdmin && func() bool {
						__antithesis_instrumentation__.Notify(461321)
						return hasControlJob == true
					}() == true) == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(461322)
					return sameUser == true
				}() == true; !canAccess {
					__antithesis_instrumentation__.Notify(461323)
					continue
				} else {
					__antithesis_instrumentation__.Notify(461324)
				}
				__antithesis_instrumentation__.Notify(461292)

				if err != nil {
					__antithesis_instrumentation__.Notify(461325)
					errorStr = tree.NewDString(fmt.Sprintf("error decoding payload: %v", err))
				} else {
					__antithesis_instrumentation__.Notify(461326)
					jobType = tree.NewDString(payload.Type().String())
					description = tree.NewDString(payload.Description)
					statement = tree.NewDString(strings.Join(payload.Statement, "; "))
					username = tree.NewDString(sqlUsername.Normalized())
					descriptorIDsArr := tree.NewDArray(types.Int)
					for _, descID := range payload.DescriptorIDs {
						__antithesis_instrumentation__.Notify(461330)
						if err := descriptorIDsArr.Append(tree.NewDInt(tree.DInt(int(descID)))); err != nil {
							__antithesis_instrumentation__.Notify(461331)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(461332)
						}
					}
					__antithesis_instrumentation__.Notify(461327)
					descriptorIDs = descriptorIDsArr
					started, err = tsOrNull(payload.StartedMicros)
					if err != nil {
						__antithesis_instrumentation__.Notify(461333)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(461334)
					}
					__antithesis_instrumentation__.Notify(461328)
					finished, err = tsOrNull(payload.FinishedMicros)
					if err != nil {
						__antithesis_instrumentation__.Notify(461335)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(461336)
					}
					__antithesis_instrumentation__.Notify(461329)
					errorStr = tree.NewDString(payload.Error)
				}
				__antithesis_instrumentation__.Notify(461293)

				if progressBytes != tree.DNull {
					__antithesis_instrumentation__.Notify(461337)
					progress, err := jobs.UnmarshalProgress(progressBytes)
					if err != nil {
						__antithesis_instrumentation__.Notify(461338)
						baseErr := ""
						if s, ok := errorStr.(*tree.DString); ok {
							__antithesis_instrumentation__.Notify(461340)
							baseErr = string(*s)
							if baseErr != "" {
								__antithesis_instrumentation__.Notify(461341)
								baseErr += "\n"
							} else {
								__antithesis_instrumentation__.Notify(461342)
							}
						} else {
							__antithesis_instrumentation__.Notify(461343)
						}
						__antithesis_instrumentation__.Notify(461339)
						errorStr = tree.NewDString(fmt.Sprintf("%serror decoding progress: %v", baseErr, err))
					} else {
						__antithesis_instrumentation__.Notify(461344)

						if highwater := progress.GetHighWater(); highwater != nil {
							__antithesis_instrumentation__.Notify(461348)
							highWaterTimestamp = tree.TimestampToDecimalDatum(*highwater)
						} else {
							__antithesis_instrumentation__.Notify(461349)
							fractionCompleted = tree.NewDFloat(tree.DFloat(progress.GetFractionCompleted()))
						}
						__antithesis_instrumentation__.Notify(461345)
						modified, err = tsOrNull(progress.ModifiedMicros)
						if err != nil {
							__antithesis_instrumentation__.Notify(461350)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(461351)
						}
						__antithesis_instrumentation__.Notify(461346)

						if s, ok := status.(*tree.DString); ok {
							__antithesis_instrumentation__.Notify(461352)
							if jobs.Status(*s) == jobs.StatusRunning && func() bool {
								__antithesis_instrumentation__.Notify(461353)
								return len(progress.RunningStatus) > 0 == true
							}() == true {
								__antithesis_instrumentation__.Notify(461354)
								runningStatus = tree.NewDString(progress.RunningStatus)
							} else {
								__antithesis_instrumentation__.Notify(461355)
								if jobs.Status(*s) == jobs.StatusPaused && func() bool {
									__antithesis_instrumentation__.Notify(461356)
									return payload != nil == true
								}() == true && func() bool {
									__antithesis_instrumentation__.Notify(461357)
									return payload.PauseReason != "" == true
								}() == true {
									__antithesis_instrumentation__.Notify(461358)
									errorStr = tree.NewDString(fmt.Sprintf("%s: %s", jobs.PauseRequestExplained, payload.PauseReason))
								} else {
									__antithesis_instrumentation__.Notify(461359)
								}
							}
						} else {
							__antithesis_instrumentation__.Notify(461360)
						}
						__antithesis_instrumentation__.Notify(461347)
						traceID = tree.NewDInt(tree.DInt(progress.TraceID))
					}
				} else {
					__antithesis_instrumentation__.Notify(461361)
				}
				__antithesis_instrumentation__.Notify(461294)
				if payload != nil {
					__antithesis_instrumentation__.Notify(461362)
					executionErrors = jobs.FormatRetriableExecutionErrorLogToStringArray(
						ctx, payload.RetriableExecutionFailureLog,
					)

					var err error
					executionEvents, err = jobs.FormatRetriableExecutionErrorLogToJSON(
						ctx, payload.RetriableExecutionFailureLog,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(461363)
						if errorStr == tree.DNull {
							__antithesis_instrumentation__.Notify(461364)
							errorStr = tree.NewDString(errors.Wrap(err, "failed to marshal execution error log").Error())
						} else {
							__antithesis_instrumentation__.Notify(461365)
							executionEvents = tree.DNull
						}
					} else {
						__antithesis_instrumentation__.Notify(461366)
					}
				} else {
					__antithesis_instrumentation__.Notify(461367)
				}
				__antithesis_instrumentation__.Notify(461295)

				container = container[:0]
				container = append(container,
					id,
					jobType,
					description,
					statement,
					username,
					descriptorIDs,
					status,
					runningStatus,
					created,
					started,
					finished,
					modified,
					fractionCompleted,
					highWaterTimestamp,
					errorStr,
					coordinatorID,
					traceID,
					lastRun,
					nextRun,
					numRuns,
					executionErrors,
					executionEvents,
				)
				return container, nil
			}
		}, cleanup, nil
	},
}

func execStatAvg(count int64, n roachpb.NumericStat) tree.Datum {
	__antithesis_instrumentation__.Notify(461368)
	if count == 0 {
		__antithesis_instrumentation__.Notify(461370)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(461371)
	}
	__antithesis_instrumentation__.Notify(461369)
	return tree.NewDFloat(tree.DFloat(n.Mean))
}

func execStatVar(count int64, n roachpb.NumericStat) tree.Datum {
	__antithesis_instrumentation__.Notify(461372)
	if count == 0 {
		__antithesis_instrumentation__.Notify(461374)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(461375)
	}
	__antithesis_instrumentation__.Notify(461373)
	return tree.NewDFloat(tree.DFloat(n.GetVariance(count)))
}

func getSQLStats(
	p *planner, virtualTableName string,
) (*persistedsqlstats.PersistedSQLStats, error) {
	__antithesis_instrumentation__.Notify(461376)
	if p.extendedEvalCtx.statsProvider == nil {
		__antithesis_instrumentation__.Notify(461378)
		return nil, errors.Newf("%s cannot be used in this context", virtualTableName)
	} else {
		__antithesis_instrumentation__.Notify(461379)
	}
	__antithesis_instrumentation__.Notify(461377)
	return p.extendedEvalCtx.statsProvider, nil
}

var crdbInternalNodeStmtStatsTable = virtualSchemaTable{
	comment: `statement statistics (in-memory, not durable; local node only). ` +
		`This table is wiped periodically (by default, at least every two hours)`,
	schema: `
CREATE TABLE crdb_internal.node_statement_statistics (
  node_id             INT NOT NULL,
  application_name    STRING NOT NULL,
  flags               STRING NOT NULL,
  statement_id        STRING NOT NULL,
  key                 STRING NOT NULL,
  anonymized          STRING,
  count               INT NOT NULL,
  first_attempt_count INT NOT NULL,
  max_retries         INT NOT NULL,
  last_error          STRING,
  rows_avg            FLOAT NOT NULL,
  rows_var            FLOAT NOT NULL,
  parse_lat_avg       FLOAT NOT NULL,
  parse_lat_var       FLOAT NOT NULL,
  plan_lat_avg        FLOAT NOT NULL,
  plan_lat_var        FLOAT NOT NULL,
  run_lat_avg         FLOAT NOT NULL,
  run_lat_var         FLOAT NOT NULL,
  service_lat_avg     FLOAT NOT NULL,
  service_lat_var     FLOAT NOT NULL,
  overhead_lat_avg    FLOAT NOT NULL,
  overhead_lat_var    FLOAT NOT NULL,
  bytes_read_avg      FLOAT NOT NULL,
  bytes_read_var      FLOAT NOT NULL,
  rows_read_avg       FLOAT NOT NULL,
  rows_read_var       FLOAT NOT NULL,
  network_bytes_avg   FLOAT,
  network_bytes_var   FLOAT,
  network_msgs_avg    FLOAT,
  network_msgs_var    FLOAT,
  max_mem_usage_avg   FLOAT,
  max_mem_usage_var   FLOAT,
  max_disk_usage_avg  FLOAT,
  max_disk_usage_var  FLOAT,
  contention_time_avg FLOAT,
  contention_time_var FLOAT,
  implicit_txn        BOOL NOT NULL,
  full_scan           BOOL NOT NULL,
  sample_plan         JSONB,
  database_name       STRING NOT NULL,
  exec_node_ids       INT[] NOT NULL
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461380)
		hasViewActivityOrViewActivityRedacted, err := p.HasViewActivityOrViewActivityRedactedRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461385)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461386)
		}
		__antithesis_instrumentation__.Notify(461381)
		if !hasViewActivityOrViewActivityRedacted {
			__antithesis_instrumentation__.Notify(461387)
			return pgerror.Newf(pgcode.InsufficientPrivilege,
				"user %s does not have %s or %s privilege", p.User(), roleoption.VIEWACTIVITY, roleoption.VIEWACTIVITYREDACTED)
		} else {
			__antithesis_instrumentation__.Notify(461388)
		}
		__antithesis_instrumentation__.Notify(461382)

		sqlStats, err := getSQLStats(p, "crdb_internal.node_statement_statistics")
		if err != nil {
			__antithesis_instrumentation__.Notify(461389)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461390)
		}
		__antithesis_instrumentation__.Notify(461383)

		nodeID, _ := p.execCfg.NodeID.OptionalNodeID()

		statementVisitor := func(_ context.Context, stats *roachpb.CollectedStatementStatistics) error {
			__antithesis_instrumentation__.Notify(461391)
			anonymized := tree.DNull
			anonStr, ok := scrubStmtStatKey(p.getVirtualTabler(), stats.Key.Query)
			if ok {
				__antithesis_instrumentation__.Notify(461398)
				anonymized = tree.NewDString(anonStr)
			} else {
				__antithesis_instrumentation__.Notify(461399)
			}
			__antithesis_instrumentation__.Notify(461392)

			errString := tree.DNull
			if stats.Stats.SensitiveInfo.LastErr != "" {
				__antithesis_instrumentation__.Notify(461400)
				errString = tree.NewDString(stats.Stats.SensitiveInfo.LastErr)
			} else {
				__antithesis_instrumentation__.Notify(461401)
			}
			__antithesis_instrumentation__.Notify(461393)
			var flags string
			if stats.Key.DistSQL {
				__antithesis_instrumentation__.Notify(461402)
				flags = "+"
			} else {
				__antithesis_instrumentation__.Notify(461403)
			}
			__antithesis_instrumentation__.Notify(461394)
			if stats.Key.Failed {
				__antithesis_instrumentation__.Notify(461404)
				flags = "!" + flags
			} else {
				__antithesis_instrumentation__.Notify(461405)
			}
			__antithesis_instrumentation__.Notify(461395)

			samplePlan := sqlstatsutil.ExplainTreePlanNodeToJSON(&stats.Stats.SensitiveInfo.MostRecentPlanDescription)

			execNodeIDs := tree.NewDArray(types.Int)
			for _, nodeID := range stats.Stats.Nodes {
				__antithesis_instrumentation__.Notify(461406)
				if err := execNodeIDs.Append(tree.NewDInt(tree.DInt(nodeID))); err != nil {
					__antithesis_instrumentation__.Notify(461407)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461408)
				}
			}
			__antithesis_instrumentation__.Notify(461396)

			err := addRow(
				tree.NewDInt(tree.DInt(nodeID)),
				tree.NewDString(stats.Key.App),
				tree.NewDString(flags),
				tree.NewDString(strconv.FormatUint(uint64(stats.ID), 10)),
				tree.NewDString(stats.Key.Query),
				anonymized,
				tree.NewDInt(tree.DInt(stats.Stats.Count)),
				tree.NewDInt(tree.DInt(stats.Stats.FirstAttemptCount)),
				tree.NewDInt(tree.DInt(stats.Stats.MaxRetries)),
				errString,
				tree.NewDFloat(tree.DFloat(stats.Stats.NumRows.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.NumRows.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.ParseLat.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.ParseLat.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.PlanLat.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.PlanLat.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.RunLat.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.RunLat.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.ServiceLat.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.ServiceLat.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.OverheadLat.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.OverheadLat.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.BytesRead.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.BytesRead.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.RowsRead.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.RowsRead.GetVariance(stats.Stats.Count))),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.NetworkBytes),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.NetworkBytes),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.NetworkMessages),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.NetworkMessages),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.MaxMemUsage),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.MaxMemUsage),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.MaxDiskUsage),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.MaxDiskUsage),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.ContentionTime),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.ContentionTime),
				tree.MakeDBool(tree.DBool(stats.Key.ImplicitTxn)),
				tree.MakeDBool(tree.DBool(stats.Key.FullScan)),
				tree.NewDJSON(samplePlan),
				tree.NewDString(stats.Key.Database),
				execNodeIDs,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(461409)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461410)
			}
			__antithesis_instrumentation__.Notify(461397)

			return nil
		}
		__antithesis_instrumentation__.Notify(461384)

		return sqlStats.GetLocalMemProvider().IterateStatementStats(ctx, &sqlstats.IteratorOptions{
			SortedAppNames: true,
			SortedKey:      true,
		}, statementVisitor)
	},
}

var crdbInternalTransactionStatisticsTable = virtualSchemaTable{
	comment: `finer-grained transaction statistics (in-memory, not durable; local node only). ` +
		`This table is wiped periodically (by default, at least every two hours)`,
	schema: `
CREATE TABLE crdb_internal.node_transaction_statistics (
  node_id             INT NOT NULL,
  application_name    STRING NOT NULL,
  key                 STRING,
  statement_ids       STRING[],
  count               INT,
  max_retries         INT,
  service_lat_avg     FLOAT NOT NULL,
  service_lat_var     FLOAT NOT NULL,
  retry_lat_avg       FLOAT NOT NULL,
  retry_lat_var       FLOAT NOT NULL,
  commit_lat_avg      FLOAT NOT NULL,
  commit_lat_var      FLOAT NOT NULL,
  rows_read_avg       FLOAT NOT NULL,
  rows_read_var       FLOAT NOT NULL,
  network_bytes_avg   FLOAT,
  network_bytes_var   FLOAT,
  network_msgs_avg    FLOAT,
  network_msgs_var    FLOAT,
  max_mem_usage_avg   FLOAT,
  max_mem_usage_var   FLOAT,
  max_disk_usage_avg  FLOAT,
  max_disk_usage_var  FLOAT,
  contention_time_avg FLOAT, 
  contention_time_var FLOAT
)
`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461411)
		hasViewActivityOrhasViewActivityRedacted, err := p.HasViewActivityOrViewActivityRedactedRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461416)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461417)
		}
		__antithesis_instrumentation__.Notify(461412)
		if !hasViewActivityOrhasViewActivityRedacted {
			__antithesis_instrumentation__.Notify(461418)
			return pgerror.Newf(pgcode.InsufficientPrivilege,
				"user %s does not have %s or %s privilege", p.User(), roleoption.VIEWACTIVITY, roleoption.VIEWACTIVITYREDACTED)
		} else {
			__antithesis_instrumentation__.Notify(461419)
		}
		__antithesis_instrumentation__.Notify(461413)

		sqlStats, err := getSQLStats(p, "crdb_internal.node_transaction_statistics")
		if err != nil {
			__antithesis_instrumentation__.Notify(461420)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461421)
		}
		__antithesis_instrumentation__.Notify(461414)

		nodeID, _ := p.execCfg.NodeID.OptionalNodeID()

		transactionVisitor := func(_ context.Context, stats *roachpb.CollectedTransactionStatistics) error {
			__antithesis_instrumentation__.Notify(461422)
			stmtFingerprintIDsDatum := tree.NewDArray(types.String)
			for _, stmtFingerprintID := range stats.StatementFingerprintIDs {
				__antithesis_instrumentation__.Notify(461425)
				if err := stmtFingerprintIDsDatum.Append(tree.NewDString(strconv.FormatUint(uint64(stmtFingerprintID), 10))); err != nil {
					__antithesis_instrumentation__.Notify(461426)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461427)
				}
			}
			__antithesis_instrumentation__.Notify(461423)

			err := addRow(
				tree.NewDInt(tree.DInt(nodeID)),
				tree.NewDString(stats.App),
				tree.NewDString(strconv.FormatUint(uint64(stats.TransactionFingerprintID), 10)),
				stmtFingerprintIDsDatum,
				tree.NewDInt(tree.DInt(stats.Stats.Count)),
				tree.NewDInt(tree.DInt(stats.Stats.MaxRetries)),
				tree.NewDFloat(tree.DFloat(stats.Stats.ServiceLat.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.ServiceLat.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.RetryLat.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.RetryLat.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.CommitLat.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.CommitLat.GetVariance(stats.Stats.Count))),
				tree.NewDFloat(tree.DFloat(stats.Stats.NumRows.Mean)),
				tree.NewDFloat(tree.DFloat(stats.Stats.NumRows.GetVariance(stats.Stats.Count))),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.NetworkBytes),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.NetworkBytes),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.NetworkMessages),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.NetworkMessages),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.MaxMemUsage),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.MaxMemUsage),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.MaxDiskUsage),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.MaxDiskUsage),
				execStatAvg(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.ContentionTime),
				execStatVar(stats.Stats.ExecStats.Count, stats.Stats.ExecStats.ContentionTime),
			)

			if err != nil {
				__antithesis_instrumentation__.Notify(461428)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461429)
			}
			__antithesis_instrumentation__.Notify(461424)

			return nil
		}
		__antithesis_instrumentation__.Notify(461415)

		return sqlStats.GetLocalMemProvider().IterateTransactionStats(ctx, &sqlstats.IteratorOptions{
			SortedAppNames: true,
			SortedKey:      true,
		}, transactionVisitor)
	},
}

var crdbInternalNodeTxnStatsTable = virtualSchemaTable{
	comment: `per-application transaction statistics (in-memory, not durable; local node only). ` +
		`This table is wiped periodically (by default, at least every two hours)`,
	schema: `
CREATE TABLE crdb_internal.node_txn_stats (
  node_id            INT NOT NULL,
  application_name   STRING NOT NULL,
  txn_count          INT NOT NULL,
  txn_time_avg_sec   FLOAT NOT NULL,
  txn_time_var_sec   FLOAT NOT NULL,
  committed_count    INT NOT NULL,
  implicit_count     INT NOT NULL
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461430)
		if err := p.RequireAdminRole(ctx, "access application statistics"); err != nil {
			__antithesis_instrumentation__.Notify(461434)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461435)
		}
		__antithesis_instrumentation__.Notify(461431)

		sqlStats, err := getSQLStats(p, "crdb_internal.node_txn_stats")
		if err != nil {
			__antithesis_instrumentation__.Notify(461436)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461437)
		}
		__antithesis_instrumentation__.Notify(461432)

		nodeID, _ := p.execCfg.NodeID.OptionalNodeID()

		appTxnStatsVisitor := func(appName string, stats *roachpb.TxnStats) error {
			__antithesis_instrumentation__.Notify(461438)
			return addRow(
				tree.NewDInt(tree.DInt(nodeID)),
				tree.NewDString(appName),
				tree.NewDInt(tree.DInt(stats.TxnCount)),
				tree.NewDFloat(tree.DFloat(stats.TxnTimeSec.Mean)),
				tree.NewDFloat(tree.DFloat(stats.TxnTimeSec.GetVariance(stats.TxnCount))),
				tree.NewDInt(tree.DInt(stats.CommittedCount)),
				tree.NewDInt(tree.DInt(stats.ImplicitCount)),
			)
		}
		__antithesis_instrumentation__.Notify(461433)

		return sqlStats.IterateAggregatedTransactionStats(ctx, &sqlstats.IteratorOptions{
			SortedAppNames: true,
		}, appTxnStatsVisitor)
	},
}

var crdbInternalSessionTraceTable = virtualSchemaTable{
	comment: `session trace accumulated so far (RAM)`,
	schema: `
CREATE TABLE crdb_internal.session_trace (
  span_idx    INT NOT NULL,        -- The span's index.
  message_idx INT NOT NULL,        -- The message's index within its span.
  timestamp   TIMESTAMPTZ NOT NULL,-- The message's timestamp.
  duration    INTERVAL,            -- The span's duration. Set only on the first
                                   -- (dummy) message on a span.
                                   -- NULL if the span was not finished at the time
                                   -- the trace has been collected.
  operation   STRING NULL,         -- The span's operation.
  loc         STRING NOT NULL,     -- The file name / line number prefix, if any.
  tag         STRING NOT NULL,     -- The logging tag, if any.
  message     STRING NOT NULL,     -- The logged message.
  age         INTERVAL NOT NULL    -- The age of this message relative to the beginning of the trace.
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461439)
		rows, err := p.ExtendedEvalContext().Tracing.getSessionTrace()
		if err != nil {
			__antithesis_instrumentation__.Notify(461442)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461443)
		}
		__antithesis_instrumentation__.Notify(461440)
		for _, r := range rows {
			__antithesis_instrumentation__.Notify(461444)
			if err := addRow(r[:]...); err != nil {
				__antithesis_instrumentation__.Notify(461445)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461446)
			}
		}
		__antithesis_instrumentation__.Notify(461441)
		return nil
	},
}

var crdbInternalClusterInflightTracesTable = virtualSchemaTable{
	comment: `traces for in-flight spans across all nodes in the cluster (cluster RPC; expensive!)`,
	schema: `
CREATE TABLE crdb_internal.cluster_inflight_traces (
  trace_id     INT NOT NULL,    -- The trace's ID.
  node_id      INT NOT NULL,    -- The node's ID.
  root_op_name STRING NOT NULL, -- The operation name of the root span in the current trace.
  trace_str    STRING NULL,     -- human readable representation of the traced remote operation.
  jaeger_json  STRING NULL,     -- Jaeger JSON representation of the traced remote operation.
  INDEX(trace_id)
)`,
	indexes: []virtualIndex{{populate: func(ctx context.Context, constraint tree.Datum, p *planner,
		db catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) (matched bool, err error) {
		__antithesis_instrumentation__.Notify(461447)
		var traceID tracingpb.TraceID
		d := tree.UnwrapDatum(p.EvalContext(), constraint)
		if d == tree.DNull {
			__antithesis_instrumentation__.Notify(461454)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(461455)
		}
		__antithesis_instrumentation__.Notify(461448)
		switch t := d.(type) {
		case *tree.DInt:
			__antithesis_instrumentation__.Notify(461456)
			traceID = tracingpb.TraceID(*t)
		default:
			__antithesis_instrumentation__.Notify(461457)
			return false, errors.AssertionFailedf(
				"unexpected type %T for trace_id column in virtual table crdb_internal.cluster_inflight_traces", d)
		}
		__antithesis_instrumentation__.Notify(461449)

		if !p.ExecCfg().Codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(461458)

			return false, pgerror.New(pgcode.FeatureNotSupported,
				"table crdb_internal.cluster_inflight_traces is not implemented on tenants")
		} else {
			__antithesis_instrumentation__.Notify(461459)
		}
		__antithesis_instrumentation__.Notify(461450)
		traceCollector := p.ExecCfg().TraceCollector
		var iter *collector.Iterator
		for iter, err = traceCollector.StartIter(ctx, traceID); err == nil && func() bool {
			__antithesis_instrumentation__.Notify(461460)
			return iter.Valid() == true
		}() == true; iter.Next() {
			__antithesis_instrumentation__.Notify(461461)
			nodeID, recording := iter.Value()
			traceString := recording.String()
			traceJaegerJSON, err := recording.ToJaegerJSON("", "", fmt.Sprintf("node %d", nodeID))
			if err != nil {
				__antithesis_instrumentation__.Notify(461463)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(461464)
			}
			__antithesis_instrumentation__.Notify(461462)
			if err := addRow(tree.NewDInt(tree.DInt(traceID)),
				tree.NewDInt(tree.DInt(nodeID)),
				tree.NewDString(recording[0].Operation),
				tree.NewDString(traceString),
				tree.NewDString(traceJaegerJSON)); err != nil {
				__antithesis_instrumentation__.Notify(461465)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(461466)
			}
		}
		__antithesis_instrumentation__.Notify(461451)
		if err != nil {
			__antithesis_instrumentation__.Notify(461467)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(461468)
		}
		__antithesis_instrumentation__.Notify(461452)
		if iter.Error() != nil {
			__antithesis_instrumentation__.Notify(461469)
			return false, iter.Error()
		} else {
			__antithesis_instrumentation__.Notify(461470)
		}
		__antithesis_instrumentation__.Notify(461453)

		return true, nil
	}}},
	populate: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor,
		addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461471)

		return pgerror.New(pgcode.FeatureNotSupported, "a trace_id value needs to be specified")
	},
}

var crdbInternalInflightTraceSpanTable = virtualSchemaTable{
	comment: `in-flight spans (RAM; local node only)`,
	schema: `
CREATE TABLE crdb_internal.node_inflight_trace_spans (
  trace_id       INT NOT NULL,    -- The trace's ID.
  parent_span_id INT NOT NULL,    -- The span's parent ID.
  span_id        INT NOT NULL,    -- The span's ID.
  goroutine_id   INT NOT NULL,    -- The ID of the goroutine on which the span was created.
  finished       BOOL NOT NULL,   -- True if the span has been Finish()ed, false otherwise.
  start_time     TIMESTAMPTZ,     -- The span's start time.
  duration       INTERVAL,        -- The span's duration, measured from start to Finish().
                                  -- A span whose recording is collected before it's finished will
                                  -- have the duration set as the "time of collection - start time".
  operation      STRING NULL      -- The span's operation.
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461472)
		hasAdmin, err := p.HasAdminRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461475)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461476)
		}
		__antithesis_instrumentation__.Notify(461473)
		if !hasAdmin {
			__antithesis_instrumentation__.Notify(461477)
			return pgerror.Newf(pgcode.InsufficientPrivilege,
				"only users with the admin role are allowed to read crdb_internal.node_inflight_trace_spans")
		} else {
			__antithesis_instrumentation__.Notify(461478)
		}
		__antithesis_instrumentation__.Notify(461474)
		return p.ExecCfg().AmbientCtx.Tracer.VisitSpans(func(span tracing.RegistrySpan) error {
			__antithesis_instrumentation__.Notify(461479)
			for _, rec := range span.GetFullRecording(tracing.RecordingVerbose) {
				__antithesis_instrumentation__.Notify(461481)
				traceID := rec.TraceID
				parentSpanID := rec.ParentSpanID
				spanID := rec.SpanID
				goroutineID := rec.GoroutineID
				finished := rec.Finished

				startTime, err := tree.MakeDTimestampTZ(rec.StartTime, time.Microsecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(461483)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461484)
				}
				__antithesis_instrumentation__.Notify(461482)

				spanDuration := rec.Duration
				operation := rec.Operation

				if err := addRow(

					tree.NewDInt(tree.DInt(traceID)),
					tree.NewDInt(tree.DInt(parentSpanID)),
					tree.NewDInt(tree.DInt(spanID)),
					tree.NewDInt(tree.DInt(goroutineID)),
					tree.MakeDBool(tree.DBool(finished)),
					startTime,
					tree.NewDInterval(
						duration.MakeDuration(spanDuration.Nanoseconds(), 0, 0),
						types.DefaultIntervalTypeMetadata,
					),
					tree.NewDString(operation),
				); err != nil {
					__antithesis_instrumentation__.Notify(461485)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461486)
				}
			}
			__antithesis_instrumentation__.Notify(461480)
			return nil
		})
	},
}

var crdbInternalClusterSettingsTable = virtualSchemaTable{
	comment: `cluster settings (RAM)`,
	schema: `
CREATE TABLE crdb_internal.cluster_settings (
  variable      STRING NOT NULL,
  value         STRING NOT NULL,
  type          STRING NOT NULL,
  public        BOOL NOT NULL, -- whether the setting is documented, which implies the user can expect support.
  description   STRING NOT NULL
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461487)
		hasAdmin, err := p.HasAdminRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461491)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461492)
		}
		__antithesis_instrumentation__.Notify(461488)
		if !hasAdmin {
			__antithesis_instrumentation__.Notify(461493)
			hasModify, err := p.HasRoleOption(ctx, roleoption.MODIFYCLUSTERSETTING)
			if err != nil {
				__antithesis_instrumentation__.Notify(461496)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461497)
			}
			__antithesis_instrumentation__.Notify(461494)
			hasView, err := p.HasRoleOption(ctx, roleoption.VIEWCLUSTERSETTING)
			if err != nil {
				__antithesis_instrumentation__.Notify(461498)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461499)
			}
			__antithesis_instrumentation__.Notify(461495)
			if !hasModify && func() bool {
				__antithesis_instrumentation__.Notify(461500)
				return !hasView == true
			}() == true {
				__antithesis_instrumentation__.Notify(461501)
				return pgerror.Newf(pgcode.InsufficientPrivilege,
					"only users with either %s or %s privileges are allowed to read "+
						"crdb_internal.cluster_settings", roleoption.MODIFYCLUSTERSETTING, roleoption.VIEWCLUSTERSETTING)
			} else {
				__antithesis_instrumentation__.Notify(461502)
			}
		} else {
			__antithesis_instrumentation__.Notify(461503)
		}
		__antithesis_instrumentation__.Notify(461489)
		for _, k := range settings.Keys(p.ExecCfg().Codec.ForSystemTenant()) {
			__antithesis_instrumentation__.Notify(461504)
			if !hasAdmin && func() bool {
				__antithesis_instrumentation__.Notify(461506)
				return settings.AdminOnly(k) == true
			}() == true {
				__antithesis_instrumentation__.Notify(461507)
				continue
			} else {
				__antithesis_instrumentation__.Notify(461508)
			}
			__antithesis_instrumentation__.Notify(461505)
			setting, _ := settings.Lookup(
				k, settings.LookupForLocalAccess, p.ExecCfg().Codec.ForSystemTenant(),
			)
			strVal := setting.String(&p.ExecCfg().Settings.SV)
			isPublic := setting.Visibility() == settings.Public
			desc := setting.Description()
			if err := addRow(
				tree.NewDString(k),
				tree.NewDString(strVal),
				tree.NewDString(setting.Typ()),
				tree.MakeDBool(tree.DBool(isPublic)),
				tree.NewDString(desc),
			); err != nil {
				__antithesis_instrumentation__.Notify(461509)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461510)
			}
		}
		__antithesis_instrumentation__.Notify(461490)
		return nil
	},
}

var crdbInternalSessionVariablesTable = virtualSchemaTable{
	comment: `session variables (RAM)`,
	schema: `
CREATE TABLE crdb_internal.session_variables (
  variable STRING NOT NULL,
  value    STRING NOT NULL,
  hidden   BOOL   NOT NULL
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461511)
		for _, vName := range varNames {
			__antithesis_instrumentation__.Notify(461513)
			gen := varGen[vName]
			value, err := gen.Get(&p.extendedEvalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(461515)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461516)
			}
			__antithesis_instrumentation__.Notify(461514)
			if err := addRow(
				tree.NewDString(vName),
				tree.NewDString(value),
				tree.MakeDBool(tree.DBool(gen.Hidden)),
			); err != nil {
				__antithesis_instrumentation__.Notify(461517)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461518)
			}
		}
		__antithesis_instrumentation__.Notify(461512)
		return nil
	},
}

const txnsSchemaPattern = `
CREATE TABLE crdb_internal.%s (
  id UUID,                 -- the unique ID of the transaction
  node_id INT,             -- the ID of the node running the transaction
  session_id STRING,       -- the ID of the session
  start TIMESTAMP,         -- the start time of the transaction
  txn_string STRING,       -- the string representation of the transcation
  application_name STRING, -- the name of the application as per SET application_name
  num_stmts INT,           -- the number of statements executed so far
  num_retries INT,         -- the number of times the transaction was restarted
  num_auto_retries INT     -- the number of times the transaction was automatically restarted
)`

var crdbInternalLocalTxnsTable = virtualSchemaTable{
	comment: "running user transactions visible by the current user (RAM; local node only)",
	schema:  fmt.Sprintf(txnsSchemaPattern, "node_transactions"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461519)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.node_transactions"); err != nil {
			__antithesis_instrumentation__.Notify(461523)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461524)
		}
		__antithesis_instrumentation__.Notify(461520)
		req, err := p.makeSessionsRequest(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461525)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461526)
		}
		__antithesis_instrumentation__.Notify(461521)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListLocalSessions(ctx, &req)
		if err != nil {
			__antithesis_instrumentation__.Notify(461527)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461528)
		}
		__antithesis_instrumentation__.Notify(461522)
		return populateTransactionsTable(ctx, addRow, response)
	},
}

var crdbInternalClusterTxnsTable = virtualSchemaTable{
	comment: "running user transactions visible by the current user (cluster RPC; expensive!)",
	schema:  fmt.Sprintf(txnsSchemaPattern, "cluster_transactions"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461529)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.cluster_transactions"); err != nil {
			__antithesis_instrumentation__.Notify(461533)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461534)
		}
		__antithesis_instrumentation__.Notify(461530)
		req, err := p.makeSessionsRequest(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461535)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461536)
		}
		__antithesis_instrumentation__.Notify(461531)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListSessions(ctx, &req)
		if err != nil {
			__antithesis_instrumentation__.Notify(461537)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461538)
		}
		__antithesis_instrumentation__.Notify(461532)
		return populateTransactionsTable(ctx, addRow, response)
	},
}

func populateTransactionsTable(
	ctx context.Context, addRow func(...tree.Datum) error, response *serverpb.ListSessionsResponse,
) error {
	__antithesis_instrumentation__.Notify(461539)
	for _, session := range response.Sessions {
		__antithesis_instrumentation__.Notify(461542)
		sessionID := getSessionID(session)
		if txn := session.ActiveTxn; txn != nil {
			__antithesis_instrumentation__.Notify(461543)
			ts, err := tree.MakeDTimestamp(txn.Start, time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(461545)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461546)
			}
			__antithesis_instrumentation__.Notify(461544)
			if err := addRow(
				tree.NewDUuid(tree.DUuid{UUID: txn.ID}),
				tree.NewDInt(tree.DInt(session.NodeID)),
				sessionID,
				ts,
				tree.NewDString(txn.TxnDescription),
				tree.NewDString(session.ApplicationName),
				tree.NewDInt(tree.DInt(txn.NumStatementsExecuted)),
				tree.NewDInt(tree.DInt(txn.NumRetries)),
				tree.NewDInt(tree.DInt(txn.NumAutoRetries)),
			); err != nil {
				__antithesis_instrumentation__.Notify(461547)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461548)
			}
		} else {
			__antithesis_instrumentation__.Notify(461549)
		}
	}
	__antithesis_instrumentation__.Notify(461540)
	for _, rpcErr := range response.Errors {
		__antithesis_instrumentation__.Notify(461550)
		log.Warningf(ctx, "%v", rpcErr.Message)
		if rpcErr.NodeID != 0 {
			__antithesis_instrumentation__.Notify(461551)

			if err := addRow(
				tree.DNull,
				tree.NewDInt(tree.DInt(rpcErr.NodeID)),
				tree.DNull,
				tree.DNull,
				tree.NewDString("-- "+rpcErr.Message),
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
			); err != nil {
				__antithesis_instrumentation__.Notify(461552)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461553)
			}
		} else {
			__antithesis_instrumentation__.Notify(461554)
		}
	}
	__antithesis_instrumentation__.Notify(461541)
	return nil
}

const queriesSchemaPattern = `
CREATE TABLE crdb_internal.%s (
  query_id         STRING,         -- the cluster-unique ID of the query
  txn_id           UUID,           -- the unique ID of the query's transaction 
  node_id          INT NOT NULL,   -- the node on which the query is running
  session_id       STRING,         -- the ID of the session
  user_name        STRING,         -- the user running the query
  start            TIMESTAMP,      -- the start time of the query
  query            STRING,         -- the SQL code of the query
  client_address   STRING,         -- the address of the client that issued the query
  application_name STRING,         -- the name of the application as per SET application_name
  distributed      BOOL,           -- whether the query is running distributed
  phase            STRING          -- the current execution phase
)`

func (p *planner) makeSessionsRequest(ctx context.Context) (serverpb.ListSessionsRequest, error) {
	__antithesis_instrumentation__.Notify(461555)
	req := serverpb.ListSessionsRequest{Username: p.SessionData().User().Normalized()}
	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(461558)
		return serverpb.ListSessionsRequest{}, err
	} else {
		__antithesis_instrumentation__.Notify(461559)
	}
	__antithesis_instrumentation__.Notify(461556)
	if hasAdmin {
		__antithesis_instrumentation__.Notify(461560)
		req.Username = ""
	} else {
		__antithesis_instrumentation__.Notify(461561)
		hasViewActivityOrhasViewActivityRedacted, err := p.HasViewActivityOrViewActivityRedactedRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461563)
			return serverpb.ListSessionsRequest{}, err
		} else {
			__antithesis_instrumentation__.Notify(461564)
		}
		__antithesis_instrumentation__.Notify(461562)
		if hasViewActivityOrhasViewActivityRedacted {
			__antithesis_instrumentation__.Notify(461565)
			req.Username = ""
		} else {
			__antithesis_instrumentation__.Notify(461566)
		}
	}
	__antithesis_instrumentation__.Notify(461557)
	return req, nil
}

func getSessionID(session serverpb.Session) tree.Datum {
	__antithesis_instrumentation__.Notify(461567)

	var sessionID tree.Datum
	if session.ID == nil {
		__antithesis_instrumentation__.Notify(461569)

		telemetry.RecordError(
			pgerror.NewInternalTrackingError(32517, "null"))
		sessionID = tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(461570)
		if len(session.ID) != 16 {
			__antithesis_instrumentation__.Notify(461571)

			telemetry.RecordError(
				pgerror.NewInternalTrackingError(32517, fmt.Sprintf("len=%d", len(session.ID))))
			sessionID = tree.NewDString("<invalid>")
		} else {
			__antithesis_instrumentation__.Notify(461572)
			clusterSessionID := BytesToClusterWideID(session.ID)
			sessionID = tree.NewDString(clusterSessionID.String())
		}
	}
	__antithesis_instrumentation__.Notify(461568)
	return sessionID
}

var crdbInternalLocalQueriesTable = virtualSchemaTable{
	comment: "running queries visible by current user (RAM; local node only)",
	schema:  fmt.Sprintf(queriesSchemaPattern, "node_queries"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461573)
		req, err := p.makeSessionsRequest(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461576)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461577)
		}
		__antithesis_instrumentation__.Notify(461574)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListLocalSessions(ctx, &req)
		if err != nil {
			__antithesis_instrumentation__.Notify(461578)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461579)
		}
		__antithesis_instrumentation__.Notify(461575)
		return populateQueriesTable(ctx, addRow, response)
	},
}

var crdbInternalClusterQueriesTable = virtualSchemaTable{
	comment: "running queries visible by current user (cluster RPC; expensive!)",
	schema:  fmt.Sprintf(queriesSchemaPattern, "cluster_queries"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461580)
		req, err := p.makeSessionsRequest(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461583)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461584)
		}
		__antithesis_instrumentation__.Notify(461581)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListSessions(ctx, &req)
		if err != nil {
			__antithesis_instrumentation__.Notify(461585)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461586)
		}
		__antithesis_instrumentation__.Notify(461582)
		return populateQueriesTable(ctx, addRow, response)
	},
}

func populateQueriesTable(
	ctx context.Context, addRow func(...tree.Datum) error, response *serverpb.ListSessionsResponse,
) error {
	__antithesis_instrumentation__.Notify(461587)
	for _, session := range response.Sessions {
		__antithesis_instrumentation__.Notify(461590)
		sessionID := getSessionID(session)
		for _, query := range session.ActiveQueries {
			__antithesis_instrumentation__.Notify(461591)
			isDistributedDatum := tree.DNull
			phase := strings.ToLower(query.Phase.String())
			if phase == "executing" {
				__antithesis_instrumentation__.Notify(461596)
				isDistributedDatum = tree.DBoolFalse
				if query.IsDistributed {
					__antithesis_instrumentation__.Notify(461597)
					isDistributedDatum = tree.DBoolTrue
				} else {
					__antithesis_instrumentation__.Notify(461598)
				}
			} else {
				__antithesis_instrumentation__.Notify(461599)
			}
			__antithesis_instrumentation__.Notify(461592)

			if query.Progress > 0 {
				__antithesis_instrumentation__.Notify(461600)
				phase = fmt.Sprintf("%s (%.2f%%)", phase, query.Progress*100)
			} else {
				__antithesis_instrumentation__.Notify(461601)
			}
			__antithesis_instrumentation__.Notify(461593)

			var txnID tree.Datum

			if query.ID == "" {
				__antithesis_instrumentation__.Notify(461602)
				txnID = tree.DNull
			} else {
				__antithesis_instrumentation__.Notify(461603)
				txnID = tree.NewDUuid(tree.DUuid{UUID: query.TxnID})
			}
			__antithesis_instrumentation__.Notify(461594)

			ts, err := tree.MakeDTimestamp(query.Start, time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(461604)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461605)
			}
			__antithesis_instrumentation__.Notify(461595)
			if err := addRow(
				tree.NewDString(query.ID),
				txnID,
				tree.NewDInt(tree.DInt(session.NodeID)),
				sessionID,
				tree.NewDString(session.Username),
				ts,
				tree.NewDString(query.Sql),
				tree.NewDString(session.ClientAddress),
				tree.NewDString(session.ApplicationName),
				isDistributedDatum,
				tree.NewDString(phase),
			); err != nil {
				__antithesis_instrumentation__.Notify(461606)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461607)
			}
		}
	}
	__antithesis_instrumentation__.Notify(461588)

	for _, rpcErr := range response.Errors {
		__antithesis_instrumentation__.Notify(461608)
		log.Warningf(ctx, "%v", rpcErr.Message)
		if rpcErr.NodeID != 0 {
			__antithesis_instrumentation__.Notify(461609)

			if err := addRow(
				tree.DNull,
				tree.DNull,
				tree.NewDInt(tree.DInt(rpcErr.NodeID)),
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.NewDString("-- "+rpcErr.Message),
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
			); err != nil {
				__antithesis_instrumentation__.Notify(461610)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461611)
			}
		} else {
			__antithesis_instrumentation__.Notify(461612)
		}
	}
	__antithesis_instrumentation__.Notify(461589)
	return nil
}

const sessionsSchemaPattern = `
CREATE TABLE crdb_internal.%s (
  node_id            INT NOT NULL,   -- the node on which the query is running
  session_id         STRING,         -- the ID of the session
  user_name          STRING,         -- the user running the query
  client_address     STRING,         -- the address of the client that issued the query
  application_name   STRING,         -- the name of the application as per SET application_name
  active_queries     STRING,         -- the currently running queries as SQL
  last_active_query  STRING,         -- the query that finished last on this session as SQL
  session_start      TIMESTAMP,      -- the time when the session was opened
  oldest_query_start TIMESTAMP,      -- the time when the oldest query in the session was started
  kv_txn             STRING,         -- the ID of the current KV transaction
  alloc_bytes        INT,            -- the number of bytes allocated by the session
  max_alloc_bytes    INT             -- the high water mark of bytes allocated by the session
)
`

var crdbInternalLocalSessionsTable = virtualSchemaTable{
	comment: "running sessions visible by current user (RAM; local node only)",
	schema:  fmt.Sprintf(sessionsSchemaPattern, "node_sessions"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461613)
		req, err := p.makeSessionsRequest(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461616)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461617)
		}
		__antithesis_instrumentation__.Notify(461614)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListLocalSessions(ctx, &req)
		if err != nil {
			__antithesis_instrumentation__.Notify(461618)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461619)
		}
		__antithesis_instrumentation__.Notify(461615)
		return populateSessionsTable(ctx, addRow, response)
	},
}

var crdbInternalClusterSessionsTable = virtualSchemaTable{
	comment: "running sessions visible to current user (cluster RPC; expensive!)",
	schema:  fmt.Sprintf(sessionsSchemaPattern, "cluster_sessions"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461620)
		req, err := p.makeSessionsRequest(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461623)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461624)
		}
		__antithesis_instrumentation__.Notify(461621)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListSessions(ctx, &req)
		if err != nil {
			__antithesis_instrumentation__.Notify(461625)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461626)
		}
		__antithesis_instrumentation__.Notify(461622)
		return populateSessionsTable(ctx, addRow, response)
	},
}

func populateSessionsTable(
	ctx context.Context, addRow func(...tree.Datum) error, response *serverpb.ListSessionsResponse,
) error {
	__antithesis_instrumentation__.Notify(461627)
	for _, session := range response.Sessions {
		__antithesis_instrumentation__.Notify(461630)

		var activeQueries bytes.Buffer
		var oldestStart time.Time
		var oldestStartDatum tree.Datum

		for idx, query := range session.ActiveQueries {
			__antithesis_instrumentation__.Notify(461635)
			if idx > 0 {
				__antithesis_instrumentation__.Notify(461637)
				activeQueries.WriteString("; ")
			} else {
				__antithesis_instrumentation__.Notify(461638)
			}
			__antithesis_instrumentation__.Notify(461636)
			activeQueries.WriteString(query.Sql)

			if oldestStart.IsZero() || func() bool {
				__antithesis_instrumentation__.Notify(461639)
				return query.Start.Before(oldestStart) == true
			}() == true {
				__antithesis_instrumentation__.Notify(461640)
				oldestStart = query.Start
			} else {
				__antithesis_instrumentation__.Notify(461641)
			}
		}
		__antithesis_instrumentation__.Notify(461631)

		var err error
		if oldestStart.IsZero() {
			__antithesis_instrumentation__.Notify(461642)
			oldestStartDatum = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(461643)
			oldestStartDatum, err = tree.MakeDTimestamp(oldestStart, time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(461644)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461645)
			}
		}
		__antithesis_instrumentation__.Notify(461632)

		kvTxnIDDatum := tree.DNull
		if session.ActiveTxn != nil {
			__antithesis_instrumentation__.Notify(461646)
			kvTxnIDDatum = tree.NewDString(session.ActiveTxn.ID.String())
		} else {
			__antithesis_instrumentation__.Notify(461647)
		}
		__antithesis_instrumentation__.Notify(461633)

		sessionID := getSessionID(session)
		startTSDatum, err := tree.MakeDTimestamp(session.Start, time.Microsecond)
		if err != nil {
			__antithesis_instrumentation__.Notify(461648)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461649)
		}
		__antithesis_instrumentation__.Notify(461634)
		if err := addRow(
			tree.NewDInt(tree.DInt(session.NodeID)),
			sessionID,
			tree.NewDString(session.Username),
			tree.NewDString(session.ClientAddress),
			tree.NewDString(session.ApplicationName),
			tree.NewDString(activeQueries.String()),
			tree.NewDString(session.LastActiveQuery),
			startTSDatum,
			oldestStartDatum,
			kvTxnIDDatum,
			tree.NewDInt(tree.DInt(session.AllocBytes)),
			tree.NewDInt(tree.DInt(session.MaxAllocBytes)),
		); err != nil {
			__antithesis_instrumentation__.Notify(461650)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461651)
		}
	}
	__antithesis_instrumentation__.Notify(461628)

	for _, rpcErr := range response.Errors {
		__antithesis_instrumentation__.Notify(461652)
		log.Warningf(ctx, "%v", rpcErr.Message)
		if rpcErr.NodeID != 0 {
			__antithesis_instrumentation__.Notify(461653)

			if err := addRow(
				tree.NewDInt(tree.DInt(rpcErr.NodeID)),
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.NewDString("-- "+rpcErr.Message),
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
				tree.DNull,
			); err != nil {
				__antithesis_instrumentation__.Notify(461654)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461655)
			}
		} else {
			__antithesis_instrumentation__.Notify(461656)
		}
	}
	__antithesis_instrumentation__.Notify(461629)

	return nil
}

var crdbInternalClusterContendedTablesView = virtualSchemaView{
	schema: `
CREATE VIEW crdb_internal.cluster_contended_tables (
  database_name,
  schema_name,
  table_name,
  num_contention_events
) AS
  SELECT
    database_name, schema_name, name, sum(num_contention_events)
  FROM
    (
      SELECT
        DISTINCT
        database_name,
        schema_name,
        name,
        index_id,
        num_contention_events
      FROM
        crdb_internal.cluster_contention_events
        JOIN crdb_internal.tables ON
            crdb_internal.cluster_contention_events.table_id
            = crdb_internal.tables.table_id
    )
  GROUP BY
    database_name, schema_name, name
`,
	resultColumns: colinfo.ResultColumns{
		{Name: "database_name", Typ: types.String},
		{Name: "schema_name", Typ: types.String},
		{Name: "table_name", Typ: types.String},
		{Name: "num_contention_events", Typ: types.Int},
	},
}

var crdbInternalClusterContendedIndexesView = virtualSchemaView{
	schema: `
CREATE VIEW crdb_internal.cluster_contended_indexes (
  database_name,
  schema_name,
  table_name,
  index_name,
  num_contention_events
) AS
  SELECT
    DISTINCT
    database_name,
    schema_name,
    name,
    index_name,
    num_contention_events
  FROM
    crdb_internal.cluster_contention_events,
    crdb_internal.tables,
    crdb_internal.table_indexes
  WHERE
    crdb_internal.cluster_contention_events.index_id
    = crdb_internal.table_indexes.index_id
    AND crdb_internal.cluster_contention_events.table_id
      = crdb_internal.table_indexes.descriptor_id
    AND crdb_internal.cluster_contention_events.table_id
      = crdb_internal.tables.table_id
  ORDER BY
    num_contention_events DESC
`,
	resultColumns: colinfo.ResultColumns{
		{Name: "database_name", Typ: types.String},
		{Name: "schema_name", Typ: types.String},
		{Name: "table_name", Typ: types.String},
		{Name: "index_name", Typ: types.String},
		{Name: "num_contention_events", Typ: types.Int},
	},
}

var crdbInternalClusterContendedKeysView = virtualSchemaView{
	schema: `
CREATE VIEW crdb_internal.cluster_contended_keys (
  database_name,
  schema_name,
  table_name,
  index_name,
  key,
  num_contention_events
) AS
  SELECT
    database_name,
    schema_name,
    name,
    index_name,
    crdb_internal.pretty_key(key, 0),
    sum(count)
  FROM
    crdb_internal.cluster_contention_events,
    crdb_internal.tables,
    crdb_internal.table_indexes
  WHERE
    crdb_internal.cluster_contention_events.index_id
    = crdb_internal.table_indexes.index_id
    AND crdb_internal.cluster_contention_events.table_id
      = crdb_internal.tables.table_id
  GROUP BY
    database_name, schema_name, name, index_name, key
`,
	resultColumns: colinfo.ResultColumns{
		{Name: "database_name", Typ: types.String},
		{Name: "schema_name", Typ: types.String},
		{Name: "table_name", Typ: types.String},
		{Name: "index_name", Typ: types.String},
		{Name: "key", Typ: types.Bytes},
		{Name: "num_contention_events", Typ: types.Int},
	},
}

const contentionEventsSchemaPattern = `
CREATE TABLE crdb_internal.%s (
  table_id                   INT,
  index_id                   INT,
  num_contention_events      INT NOT NULL,
  cumulative_contention_time INTERVAL NOT NULL,
  key                        BYTES NOT NULL,
  txn_id                     UUID NOT NULL,
  count                      INT NOT NULL
)
`
const contentionEventsCommentPattern = `contention information %s

All of the contention information internally stored in three levels:
- on the highest, it is grouped by tableID/indexID pair
- on the middle, it is grouped by key
- on the lowest, it is grouped by txnID.
Each of the levels is maintained as an LRU cache with limited size, so
it is possible that not all of the contention information ever observed
is contained in this table.
`

var crdbInternalLocalContentionEventsTable = virtualSchemaTable{
	schema:  fmt.Sprintf(contentionEventsSchemaPattern, "node_contention_events"),
	comment: fmt.Sprintf(contentionEventsCommentPattern, "(RAM; local node only)"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461657)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListLocalContentionEvents(ctx, &serverpb.ListContentionEventsRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(461659)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461660)
		}
		__antithesis_instrumentation__.Notify(461658)
		return populateContentionEventsTable(ctx, addRow, response)
	},
}

var crdbInternalClusterContentionEventsTable = virtualSchemaTable{
	schema:  fmt.Sprintf(contentionEventsSchemaPattern, "cluster_contention_events"),
	comment: fmt.Sprintf(contentionEventsCommentPattern, "(cluster RPC; expensive!)"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461661)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListContentionEvents(ctx, &serverpb.ListContentionEventsRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(461663)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461664)
		}
		__antithesis_instrumentation__.Notify(461662)
		return populateContentionEventsTable(ctx, addRow, response)
	},
}

func populateContentionEventsTable(
	ctx context.Context,
	addRow func(...tree.Datum) error,
	response *serverpb.ListContentionEventsResponse,
) error {
	__antithesis_instrumentation__.Notify(461665)
	for _, ice := range response.Events.IndexContentionEvents {
		__antithesis_instrumentation__.Notify(461669)
		for _, skc := range ice.Events {
			__antithesis_instrumentation__.Notify(461670)
			for _, stc := range skc.Txns {
				__antithesis_instrumentation__.Notify(461671)
				cumulativeContentionTime := tree.NewDInterval(
					duration.MakeDuration(ice.CumulativeContentionTime.Nanoseconds(), 0, 0),
					types.DefaultIntervalTypeMetadata,
				)
				if err := addRow(
					tree.NewDInt(tree.DInt(ice.TableID)),
					tree.NewDInt(tree.DInt(ice.IndexID)),
					tree.NewDInt(tree.DInt(ice.NumContentionEvents)),
					cumulativeContentionTime,
					tree.NewDBytes(tree.DBytes(skc.Key)),
					tree.NewDUuid(tree.DUuid{UUID: stc.TxnID}),
					tree.NewDInt(tree.DInt(stc.Count)),
				); err != nil {
					__antithesis_instrumentation__.Notify(461672)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461673)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(461666)
	for _, nkc := range response.Events.NonSQLKeysContention {
		__antithesis_instrumentation__.Notify(461674)
		for _, stc := range nkc.Txns {
			__antithesis_instrumentation__.Notify(461675)
			cumulativeContentionTime := tree.NewDInterval(
				duration.MakeDuration(nkc.CumulativeContentionTime.Nanoseconds(), 0, 0),
				types.DefaultIntervalTypeMetadata,
			)
			if err := addRow(
				tree.DNull,
				tree.DNull,
				tree.NewDInt(tree.DInt(nkc.NumContentionEvents)),
				cumulativeContentionTime,
				tree.NewDBytes(tree.DBytes(nkc.Key)),
				tree.NewDUuid(tree.DUuid{UUID: stc.TxnID}),
				tree.NewDInt(tree.DInt(stc.Count)),
			); err != nil {
				__antithesis_instrumentation__.Notify(461676)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461677)
			}
		}
	}
	__antithesis_instrumentation__.Notify(461667)
	for _, rpcErr := range response.Errors {
		__antithesis_instrumentation__.Notify(461678)
		log.Warningf(ctx, "%v", rpcErr.Message)
	}
	__antithesis_instrumentation__.Notify(461668)
	return nil
}

const distSQLFlowsSchemaPattern = `
CREATE TABLE crdb_internal.%s (
  flow_id UUID NOT NULL,
  node_id INT NOT NULL,
  stmt    STRING NULL,
  since   TIMESTAMPTZ NOT NULL,
  status  STRING NOT NULL
)
`

const distSQLFlowsCommentPattern = `DistSQL remote flows information %s

This virtual table contains all of the remote flows of the DistSQL execution
that are currently running or queued on %s. The local
flows (those that are running on the same node as the query originated on)
are not included.
`

var crdbInternalLocalDistSQLFlowsTable = virtualSchemaTable{
	schema:  fmt.Sprintf(distSQLFlowsSchemaPattern, "node_distsql_flows"),
	comment: fmt.Sprintf(distSQLFlowsCommentPattern, "(RAM; local node only)", "this node"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461679)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListLocalDistSQLFlows(ctx, &serverpb.ListDistSQLFlowsRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(461681)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461682)
		}
		__antithesis_instrumentation__.Notify(461680)
		return populateDistSQLFlowsTable(ctx, addRow, response)
	},
}

var crdbInternalClusterDistSQLFlowsTable = virtualSchemaTable{
	schema:  fmt.Sprintf(distSQLFlowsSchemaPattern, "cluster_distsql_flows"),
	comment: fmt.Sprintf(distSQLFlowsCommentPattern, "(cluster RPC; expensive!)", "any node in the cluster"),
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461683)
		response, err := p.extendedEvalCtx.SQLStatusServer.ListDistSQLFlows(ctx, &serverpb.ListDistSQLFlowsRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(461685)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461686)
		}
		__antithesis_instrumentation__.Notify(461684)
		return populateDistSQLFlowsTable(ctx, addRow, response)
	},
}

func populateDistSQLFlowsTable(
	ctx context.Context,
	addRow func(...tree.Datum) error,
	response *serverpb.ListDistSQLFlowsResponse,
) error {
	__antithesis_instrumentation__.Notify(461687)
	for _, f := range response.Flows {
		__antithesis_instrumentation__.Notify(461690)
		flowID := tree.NewDUuid(tree.DUuid{UUID: f.FlowID.UUID})
		for _, info := range f.Infos {
			__antithesis_instrumentation__.Notify(461691)
			nodeID := tree.NewDInt(tree.DInt(info.NodeID))
			stmt := tree.NewDString(info.Stmt)
			since, err := tree.MakeDTimestampTZ(info.Timestamp, time.Millisecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(461693)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461694)
			}
			__antithesis_instrumentation__.Notify(461692)
			status := tree.NewDString(strings.ToLower(info.Status.String()))
			if err = addRow(flowID, nodeID, stmt, since, status); err != nil {
				__antithesis_instrumentation__.Notify(461695)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461696)
			}
		}
	}
	__antithesis_instrumentation__.Notify(461688)
	for _, rpcErr := range response.Errors {
		__antithesis_instrumentation__.Notify(461697)
		log.Warningf(ctx, "%v", rpcErr.Message)
	}
	__antithesis_instrumentation__.Notify(461689)
	return nil
}

var crdbInternalLocalMetricsTable = virtualSchemaTable{
	comment: "current values for metrics (RAM; local node only)",
	schema: `CREATE TABLE crdb_internal.node_metrics (
  store_id 	         INT NULL,         -- the store, if any, for this metric
  name               STRING NOT NULL,  -- name of the metric
  value							 FLOAT NOT NULL    -- value of the metric
)`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461698)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.node_metrics"); err != nil {
			__antithesis_instrumentation__.Notify(461702)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461703)
		}
		__antithesis_instrumentation__.Notify(461699)

		mr := p.ExecCfg().MetricsRecorder
		if mr == nil {
			__antithesis_instrumentation__.Notify(461704)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(461705)
		}
		__antithesis_instrumentation__.Notify(461700)
		nodeStatus := mr.GenerateNodeStatus(ctx)
		for i := 0; i <= len(nodeStatus.StoreStatuses); i++ {
			__antithesis_instrumentation__.Notify(461706)
			storeID := tree.DNull
			mtr := nodeStatus.Metrics
			if i > 0 {
				__antithesis_instrumentation__.Notify(461708)
				storeID = tree.NewDInt(tree.DInt(nodeStatus.StoreStatuses[i-1].Desc.StoreID))
				mtr = nodeStatus.StoreStatuses[i-1].Metrics
			} else {
				__antithesis_instrumentation__.Notify(461709)
			}
			__antithesis_instrumentation__.Notify(461707)
			for name, value := range mtr {
				__antithesis_instrumentation__.Notify(461710)
				if err := addRow(
					storeID,
					tree.NewDString(name),
					tree.NewDFloat(tree.DFloat(value)),
				); err != nil {
					__antithesis_instrumentation__.Notify(461711)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461712)
				}
			}
		}
		__antithesis_instrumentation__.Notify(461701)
		return nil
	},
}

var crdbInternalBuiltinFunctionsTable = virtualSchemaTable{
	comment: "built-in functions (RAM/static)",
	schema: `
CREATE TABLE crdb_internal.builtin_functions (
  function  STRING NOT NULL,
  signature STRING NOT NULL,
  category  STRING NOT NULL,
  details   STRING NOT NULL
)`,
	populate: func(ctx context.Context, _ *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461713)
		for _, name := range builtins.AllBuiltinNames {
			__antithesis_instrumentation__.Notify(461715)
			props, overloads := builtins.GetBuiltinProperties(name)
			for _, f := range overloads {
				__antithesis_instrumentation__.Notify(461716)
				if err := addRow(
					tree.NewDString(name),
					tree.NewDString(f.Signature(false)),
					tree.NewDString(props.Category),
					tree.NewDString(f.Info),
				); err != nil {
					__antithesis_instrumentation__.Notify(461717)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461718)
				}
			}
		}
		__antithesis_instrumentation__.Notify(461714)
		return nil
	},
}

func writeCreateTypeDescRow(
	db catalog.DatabaseDescriptor,
	sc string,
	typeDesc catalog.TypeDescriptor,
	addRow func(...tree.Datum) error,
) (written bool, err error) {
	__antithesis_instrumentation__.Notify(461719)
	switch typeDesc.GetKind() {
	case descpb.TypeDescriptor_ENUM:
		__antithesis_instrumentation__.Notify(461720)
		var enumLabels tree.EnumValueList
		enumLabelsDatum := tree.NewDArray(types.String)
		for i := 0; i < typeDesc.NumEnumMembers(); i++ {
			__antithesis_instrumentation__.Notify(461726)
			rep := typeDesc.GetMemberLogicalRepresentation(i)
			enumLabels = append(enumLabels, tree.EnumValue(rep))
			if err := enumLabelsDatum.Append(tree.NewDString(rep)); err != nil {
				__antithesis_instrumentation__.Notify(461727)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(461728)
			}
		}
		__antithesis_instrumentation__.Notify(461721)
		name, err := tree.NewUnresolvedObjectName(2, [3]string{typeDesc.GetName(), sc}, 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(461729)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(461730)
		}
		__antithesis_instrumentation__.Notify(461722)
		node := &tree.CreateType{
			Variety:    tree.Enum,
			TypeName:   name,
			EnumLabels: enumLabels,
		}
		return true, addRow(
			tree.NewDInt(tree.DInt(db.GetID())),
			tree.NewDString(db.GetName()),
			tree.NewDString(sc),
			tree.NewDInt(tree.DInt(typeDesc.GetID())),
			tree.NewDString(typeDesc.GetName()),
			tree.NewDString(tree.AsString(node)),
			enumLabelsDatum,
		)
	case descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(461723)

		return false, nil
	case descpb.TypeDescriptor_ALIAS:
		__antithesis_instrumentation__.Notify(461724)

		return false, nil
	default:
		__antithesis_instrumentation__.Notify(461725)
		return false, errors.AssertionFailedf("unknown type descriptor kind %s", typeDesc.GetKind().String())
	}
}

var crdbInternalCreateTypeStmtsTable = virtualSchemaTable{
	comment: "CREATE statements for all user defined types accessible by the current user in current database (KV scan)",
	schema: `
CREATE TABLE crdb_internal.create_type_statements (
	database_id        INT,
  database_name      STRING,
  schema_name        STRING,
  descriptor_id      INT,
  descriptor_name    STRING,
  create_statement   STRING,
  enum_members       STRING[], -- populated only for ENUM types
  INDEX (descriptor_id)
)
`,
	populate: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461731)
		return forEachTypeDesc(ctx, p, db, func(db catalog.DatabaseDescriptor, sc string, typeDesc catalog.TypeDescriptor) error {
			__antithesis_instrumentation__.Notify(461732)
			_, err := writeCreateTypeDescRow(db, sc, typeDesc, addRow)
			return err
		})
	},
	indexes: []virtualIndex{
		{
			populate: func(
				ctx context.Context,
				constraint tree.Datum,
				p *planner,
				db catalog.DatabaseDescriptor,
				addRow func(...tree.Datum) error,
			) (matched bool, err error) {
				__antithesis_instrumentation__.Notify(461733)
				d := tree.UnwrapDatum(p.EvalContext(), constraint)
				if d == tree.DNull {
					__antithesis_instrumentation__.Notify(461736)
					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(461737)
				}
				__antithesis_instrumentation__.Notify(461734)
				id := descpb.ID(tree.MustBeDInt(d))
				scName, typDesc, err := getSchemaAndTypeByTypeID(ctx, p, id)
				if err != nil || func() bool {
					__antithesis_instrumentation__.Notify(461738)
					return typDesc == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(461739)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(461740)
				}
				__antithesis_instrumentation__.Notify(461735)
				return writeCreateTypeDescRow(db, scName, typDesc, addRow)
			},
		},
	},
}

var crdbInternalCreateSchemaStmtsTable = virtualSchemaTable{
	comment: "CREATE statements for all user defined schemas accessible by the current user in current database (KV scan)",
	schema: `
CREATE TABLE crdb_internal.create_schema_statements (
	database_id        INT,
  database_name      STRING,
  schema_name        STRING,
  descriptor_id      INT,
  create_statement   STRING
)
`,
	populate: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461741)
		var dbDescs []catalog.DatabaseDescriptor
		if db == nil {
			__antithesis_instrumentation__.Notify(461744)
			var err error
			dbDescs, err = p.Descriptors().GetAllDatabaseDescriptors(ctx, p.Txn())
			if err != nil {
				__antithesis_instrumentation__.Notify(461745)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461746)
			}
		} else {
			__antithesis_instrumentation__.Notify(461747)
			dbDescs = append(dbDescs, db)
		}
		__antithesis_instrumentation__.Notify(461742)
		for _, db := range dbDescs {
			__antithesis_instrumentation__.Notify(461748)
			return forEachSchema(ctx, p, db, func(schemaDesc catalog.SchemaDescriptor) error {
				__antithesis_instrumentation__.Notify(461749)
				switch schemaDesc.SchemaKind() {
				case catalog.SchemaUserDefined:
					__antithesis_instrumentation__.Notify(461751)
					node := &tree.CreateSchema{
						Schema: tree.ObjectNamePrefix{
							SchemaName:     tree.Name(schemaDesc.GetName()),
							ExplicitSchema: true,
						},
					}
					if err := addRow(
						tree.NewDInt(tree.DInt(db.GetID())),
						tree.NewDString(db.GetName()),
						tree.NewDString(schemaDesc.GetName()),
						tree.NewDInt(tree.DInt(schemaDesc.GetID())),
						tree.NewDString(tree.AsString(node)),
					); err != nil {
						__antithesis_instrumentation__.Notify(461753)
						return err
					} else {
						__antithesis_instrumentation__.Notify(461754)
					}
				default:
					__antithesis_instrumentation__.Notify(461752)
				}
				__antithesis_instrumentation__.Notify(461750)
				return nil
			})
		}
		__antithesis_instrumentation__.Notify(461743)
		return nil
	},
}

var typeView = tree.NewDString("view")
var typeTable = tree.NewDString("table")
var typeSequence = tree.NewDString("sequence")

var crdbInternalCreateStmtsTable = makeAllRelationsVirtualTableWithDescriptorIDIndex(
	`CREATE and ALTER statements for all tables accessible by current user in current database (KV scan)`,
	`
CREATE TABLE crdb_internal.create_statements (
  database_id                   INT,
  database_name                 STRING,
  schema_name                   STRING NOT NULL,
  descriptor_id                 INT,
  descriptor_type               STRING NOT NULL,
  descriptor_name               STRING NOT NULL,
  create_statement              STRING NOT NULL,
  state                         STRING NOT NULL,
  create_nofks                  STRING NOT NULL,
  alter_statements              STRING[] NOT NULL,
  validate_statements           STRING[] NOT NULL,
  has_partitions                BOOL NOT NULL,
  is_multi_region               BOOL NOT NULL,
  is_virtual                    BOOL NOT NULL,
  is_temporary                  BOOL NOT NULL,
  INDEX(descriptor_id)
)
`, virtualCurrentDB, false,
	func(ctx context.Context, p *planner, h oidHasher, db catalog.DatabaseDescriptor, scName string,
		table catalog.TableDescriptor, lookup simpleSchemaResolver, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461755)
		contextName := ""
		parentNameStr := tree.DNull
		if db != nil {
			__antithesis_instrumentation__.Notify(461761)
			contextName = db.GetName()
			parentNameStr = tree.NewDString(contextName)
		} else {
			__antithesis_instrumentation__.Notify(461762)
		}
		__antithesis_instrumentation__.Notify(461756)
		scNameStr := tree.NewDString(scName)

		var descType tree.Datum
		var stmt, createNofk string
		alterStmts := tree.NewDArray(types.String)
		validateStmts := tree.NewDArray(types.String)
		namePrefix := tree.ObjectNamePrefix{SchemaName: tree.Name(scName), ExplicitSchema: true}
		name := tree.MakeTableNameFromPrefix(namePrefix, tree.Name(table.GetName()))
		var err error
		if table.IsView() {
			__antithesis_instrumentation__.Notify(461763)
			descType = typeView
			stmt, err = ShowCreateView(ctx, &p.semaCtx, p.SessionData(), &name, table)
		} else {
			__antithesis_instrumentation__.Notify(461764)
			if table.IsSequence() {
				__antithesis_instrumentation__.Notify(461765)
				descType = typeSequence
				stmt, err = ShowCreateSequence(ctx, &name, table)
			} else {
				__antithesis_instrumentation__.Notify(461766)
				descType = typeTable
				displayOptions := ShowCreateDisplayOptions{
					FKDisplayMode: OmitFKClausesFromCreate,
				}
				createNofk, err = ShowCreateTable(ctx, p, &name, contextName, table, lookup, displayOptions)
				if err != nil {
					__antithesis_instrumentation__.Notify(461769)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461770)
				}
				__antithesis_instrumentation__.Notify(461767)
				if err := showAlterStatement(ctx, &name, contextName, lookup, table, alterStmts, validateStmts); err != nil {
					__antithesis_instrumentation__.Notify(461771)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461772)
				}
				__antithesis_instrumentation__.Notify(461768)
				displayOptions.FKDisplayMode = IncludeFkClausesInCreate
				stmt, err = ShowCreateTable(ctx, p, &name, contextName, table, lookup, displayOptions)
			}
		}
		__antithesis_instrumentation__.Notify(461757)
		if err != nil {
			__antithesis_instrumentation__.Notify(461773)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461774)
		}
		__antithesis_instrumentation__.Notify(461758)

		descID := tree.NewDInt(tree.DInt(table.GetID()))
		dbDescID := tree.NewDInt(tree.DInt(table.GetParentID()))
		if createNofk == "" {
			__antithesis_instrumentation__.Notify(461775)
			createNofk = stmt
		} else {
			__antithesis_instrumentation__.Notify(461776)
		}
		__antithesis_instrumentation__.Notify(461759)
		hasPartitions := nil != catalog.FindIndex(table, catalog.IndexOpts{}, func(idx catalog.Index) bool {
			__antithesis_instrumentation__.Notify(461777)
			return idx.GetPartitioning().NumColumns() != 0
		})
		__antithesis_instrumentation__.Notify(461760)
		return addRow(
			dbDescID,
			parentNameStr,
			scNameStr,
			descID,
			descType,
			tree.NewDString(table.GetName()),
			tree.NewDString(stmt),
			tree.NewDString(table.GetState().String()),
			tree.NewDString(createNofk),
			alterStmts,
			validateStmts,
			tree.MakeDBool(tree.DBool(hasPartitions)),
			tree.MakeDBool(tree.DBool(db != nil && func() bool {
				__antithesis_instrumentation__.Notify(461778)
				return db.IsMultiRegion() == true
			}() == true)),
			tree.MakeDBool(tree.DBool(table.IsVirtualTable())),
			tree.MakeDBool(tree.DBool(table.IsTemporary())),
		)
	})

func showAlterStatement(
	ctx context.Context,
	tn *tree.TableName,
	contextName string,
	lCtx simpleSchemaResolver,
	table catalog.TableDescriptor,
	alterStmts *tree.DArray,
	validateStmts *tree.DArray,
) error {
	__antithesis_instrumentation__.Notify(461779)
	return table.ForeachOutboundFK(func(fk *descpb.ForeignKeyConstraint) error {
		__antithesis_instrumentation__.Notify(461780)
		f := tree.NewFmtCtx(tree.FmtSimple)
		f.WriteString("ALTER TABLE ")
		f.FormatNode(tn)
		f.WriteString(" ADD CONSTRAINT ")
		f.FormatNameP(&fk.Name)
		f.WriteByte(' ')

		if err := showForeignKeyConstraint(
			&f.Buffer,
			contextName,
			table,
			fk,
			lCtx,
			sessiondata.EmptySearchPath,
		); err != nil {
			__antithesis_instrumentation__.Notify(461783)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461784)
		}
		__antithesis_instrumentation__.Notify(461781)
		if err := alterStmts.Append(tree.NewDString(f.CloseAndGetString())); err != nil {
			__antithesis_instrumentation__.Notify(461785)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461786)
		}
		__antithesis_instrumentation__.Notify(461782)

		f = tree.NewFmtCtx(tree.FmtSimple)
		f.WriteString("ALTER TABLE ")
		f.FormatNode(tn)
		f.WriteString(" VALIDATE CONSTRAINT ")
		f.FormatNameP(&fk.Name)

		return validateStmts.Append(tree.NewDString(f.CloseAndGetString()))
	})
}

var crdbInternalTableColumnsTable = virtualSchemaTable{
	comment: "details for all columns accessible by current user in current database (KV scan)",
	schema: `
CREATE TABLE crdb_internal.table_columns (
  descriptor_id    INT,
  descriptor_name  STRING NOT NULL,
  column_id        INT NOT NULL,
  column_name      STRING NOT NULL,
  column_type      STRING NOT NULL,
  nullable         BOOL NOT NULL,
  default_expr     STRING,
  hidden           BOOL NOT NULL
)
`,
	generator: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(461787)
		row := make(tree.Datums, 8)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(461789)
			return forEachTableDescAll(ctx, p, dbContext, hideVirtual,
				func(db catalog.DatabaseDescriptor, _ string, table catalog.TableDescriptor) error {
					__antithesis_instrumentation__.Notify(461790)
					tableID := tree.NewDInt(tree.DInt(table.GetID()))
					tableName := tree.NewDString(table.GetName())
					columns := table.PublicColumns()
					for _, col := range columns {
						__antithesis_instrumentation__.Notify(461792)
						defStr := tree.DNull
						if col.HasDefault() {
							__antithesis_instrumentation__.Notify(461794)
							defExpr, err := schemaexpr.FormatExprForDisplay(
								ctx, table, col.GetDefaultExpr(), &p.semaCtx, p.SessionData(), tree.FmtParsable,
							)
							if err != nil {
								__antithesis_instrumentation__.Notify(461796)
								return err
							} else {
								__antithesis_instrumentation__.Notify(461797)
							}
							__antithesis_instrumentation__.Notify(461795)
							defStr = tree.NewDString(defExpr)
						} else {
							__antithesis_instrumentation__.Notify(461798)
						}
						__antithesis_instrumentation__.Notify(461793)
						row = row[:0]
						row = append(row,
							tableID,
							tableName,
							tree.NewDInt(tree.DInt(col.GetID())),
							tree.NewDString(col.GetName()),
							tree.NewDString(col.GetType().DebugString()),
							tree.MakeDBool(tree.DBool(col.IsNullable())),
							defStr,
							tree.MakeDBool(tree.DBool(col.IsHidden())),
						)
						if err := pusher.pushRow(row...); err != nil {
							__antithesis_instrumentation__.Notify(461799)
							return err
						} else {
							__antithesis_instrumentation__.Notify(461800)
						}
					}
					__antithesis_instrumentation__.Notify(461791)
					return nil
				},
			)
		}
		__antithesis_instrumentation__.Notify(461788)
		return setupGenerator(ctx, worker, stopper)
	},
}

var crdbInternalTableIndexesTable = virtualSchemaTable{
	comment: "indexes accessible by current user in current database (KV scan)",
	schema: `
CREATE TABLE crdb_internal.table_indexes (
  descriptor_id       INT,
  descriptor_name     STRING NOT NULL,
  index_id            INT NOT NULL,
  index_name          STRING NOT NULL,
  index_type          STRING NOT NULL,
  is_unique           BOOL NOT NULL,
  is_inverted         BOOL NOT NULL,
  is_sharded          BOOL NOT NULL,
  shard_bucket_count  INT,
  created_at          TIMESTAMP
)
`,
	generator: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(461801)
		primary := tree.NewDString("primary")
		secondary := tree.NewDString("secondary")
		row := make(tree.Datums, 7)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(461803)
			return forEachTableDescAll(ctx, p, dbContext, hideVirtual,
				func(db catalog.DatabaseDescriptor, _ string, table catalog.TableDescriptor) error {
					__antithesis_instrumentation__.Notify(461804)
					tableID := tree.NewDInt(tree.DInt(table.GetID()))
					tableName := tree.NewDString(table.GetName())

					return catalog.ForEachIndex(table, catalog.IndexOpts{
						NonPhysicalPrimaryIndex: true,
					}, func(idx catalog.Index) error {
						__antithesis_instrumentation__.Notify(461805)
						row = row[:0]
						idxType := secondary
						if idx.Primary() {
							__antithesis_instrumentation__.Notify(461809)
							idxType = primary
						} else {
							__antithesis_instrumentation__.Notify(461810)
						}
						__antithesis_instrumentation__.Notify(461806)
						createdAt := tree.DNull
						if ts := idx.CreatedAt(); !ts.IsZero() {
							__antithesis_instrumentation__.Notify(461811)
							tsDatum, err := tree.MakeDTimestamp(ts, time.Nanosecond)
							if err != nil {
								__antithesis_instrumentation__.Notify(461812)
								log.Warningf(ctx, "failed to construct timestamp for index: %v", err)
							} else {
								__antithesis_instrumentation__.Notify(461813)
								createdAt = tsDatum
							}
						} else {
							__antithesis_instrumentation__.Notify(461814)
						}
						__antithesis_instrumentation__.Notify(461807)
						shardBucketCnt := tree.DNull
						if idx.IsSharded() {
							__antithesis_instrumentation__.Notify(461815)
							shardBucketCnt = tree.NewDInt(tree.DInt(idx.GetSharded().ShardBuckets))
						} else {
							__antithesis_instrumentation__.Notify(461816)
						}
						__antithesis_instrumentation__.Notify(461808)
						row = append(row,
							tableID,
							tableName,
							tree.NewDInt(tree.DInt(idx.GetID())),
							tree.NewDString(idx.GetName()),
							idxType,
							tree.MakeDBool(tree.DBool(idx.IsUnique())),
							tree.MakeDBool(idx.GetType() == descpb.IndexDescriptor_INVERTED),
							tree.MakeDBool(tree.DBool(idx.IsSharded())),
							shardBucketCnt,
							createdAt,
						)
						return pusher.pushRow(row...)
					})
				},
			)
		}
		__antithesis_instrumentation__.Notify(461802)
		return setupGenerator(ctx, worker, stopper)
	},
}

var crdbInternalIndexColumnsTable = virtualSchemaTable{
	comment: "index columns for all indexes accessible by current user in current database (KV scan)",
	schema: `
CREATE TABLE crdb_internal.index_columns (
  descriptor_id    INT,
  descriptor_name  STRING NOT NULL,
  index_id         INT NOT NULL,
  index_name       STRING NOT NULL,
  column_type      STRING NOT NULL,
  column_id        INT NOT NULL,
  column_name      STRING,
  column_direction STRING,
  implicit         BOOL
)
`,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461817)
		key := tree.NewDString("key")
		storing := tree.NewDString("storing")
		extra := tree.NewDString("extra")
		composite := tree.NewDString("composite")
		idxDirMap := map[descpb.IndexDescriptor_Direction]tree.Datum{
			descpb.IndexDescriptor_ASC:  tree.NewDString(descpb.IndexDescriptor_ASC.String()),
			descpb.IndexDescriptor_DESC: tree.NewDString(descpb.IndexDescriptor_DESC.String()),
		}

		return forEachTableDescAll(ctx, p, dbContext, hideVirtual,
			func(parent catalog.DatabaseDescriptor, _ string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(461818)
				tableID := tree.NewDInt(tree.DInt(table.GetID()))
				parentName := parent.GetName()
				tableName := tree.NewDString(table.GetName())

				reportIndex := func(idx catalog.Index) error {
					__antithesis_instrumentation__.Notify(461820)
					idxID := tree.NewDInt(tree.DInt(idx.GetID()))
					idxName := tree.NewDString(idx.GetName())

					for i := 0; i < idx.NumKeyColumns(); i++ {
						__antithesis_instrumentation__.Notify(461825)
						c := idx.GetKeyColumnID(i)
						colName := tree.DNull
						colDir := tree.DNull
						if i >= len(idx.IndexDesc().KeyColumnNames) {
							__antithesis_instrumentation__.Notify(461828)

							log.Errorf(ctx, "index descriptor for [%d@%d] (%s.%s@%s) has more key column IDs (%d) than names (%d) (corrupted schema?)",
								table.GetID(), idx.GetID(), parentName, table.GetName(), idx.GetName(),
								len(idx.IndexDesc().KeyColumnIDs), len(idx.IndexDesc().KeyColumnNames))
						} else {
							__antithesis_instrumentation__.Notify(461829)
							colName = tree.NewDString(idx.GetKeyColumnName(i))
						}
						__antithesis_instrumentation__.Notify(461826)
						if i >= len(idx.IndexDesc().KeyColumnDirections) {
							__antithesis_instrumentation__.Notify(461830)

							log.Errorf(ctx, "index descriptor for [%d@%d] (%s.%s@%s) has more key column IDs (%d) than directions (%d) (corrupted schema?)",
								table.GetID(), idx.GetID(), parentName, table.GetName(), idx.GetName(),
								len(idx.IndexDesc().KeyColumnIDs), len(idx.IndexDesc().KeyColumnDirections))
						} else {
							__antithesis_instrumentation__.Notify(461831)
							colDir = idxDirMap[idx.GetKeyColumnDirection(i)]
						}
						__antithesis_instrumentation__.Notify(461827)

						if err := addRow(
							tableID, tableName, idxID, idxName,
							key, tree.NewDInt(tree.DInt(c)), colName, colDir,
							tree.MakeDBool(i < idx.ExplicitColumnStartIdx()),
						); err != nil {
							__antithesis_instrumentation__.Notify(461832)
							return err
						} else {
							__antithesis_instrumentation__.Notify(461833)
						}
					}
					__antithesis_instrumentation__.Notify(461821)

					notImplicit := tree.DBoolFalse

					for i := 0; i < idx.NumSecondaryStoredColumns(); i++ {
						__antithesis_instrumentation__.Notify(461834)
						c := idx.GetStoredColumnID(i)
						if err := addRow(
							tableID, tableName, idxID, idxName,
							storing, tree.NewDInt(tree.DInt(c)), tree.DNull, tree.DNull,
							notImplicit,
						); err != nil {
							__antithesis_instrumentation__.Notify(461835)
							return err
						} else {
							__antithesis_instrumentation__.Notify(461836)
						}
					}
					__antithesis_instrumentation__.Notify(461822)

					for i := 0; i < idx.NumKeySuffixColumns(); i++ {
						__antithesis_instrumentation__.Notify(461837)
						c := idx.GetKeySuffixColumnID(i)
						if err := addRow(
							tableID, tableName, idxID, idxName,
							extra, tree.NewDInt(tree.DInt(c)), tree.DNull, tree.DNull,
							notImplicit,
						); err != nil {
							__antithesis_instrumentation__.Notify(461838)
							return err
						} else {
							__antithesis_instrumentation__.Notify(461839)
						}
					}
					__antithesis_instrumentation__.Notify(461823)

					for i := 0; i < idx.NumCompositeColumns(); i++ {
						__antithesis_instrumentation__.Notify(461840)
						c := idx.GetCompositeColumnID(i)
						if err := addRow(
							tableID, tableName, idxID, idxName,
							composite, tree.NewDInt(tree.DInt(c)), tree.DNull, tree.DNull,
							notImplicit,
						); err != nil {
							__antithesis_instrumentation__.Notify(461841)
							return err
						} else {
							__antithesis_instrumentation__.Notify(461842)
						}
					}
					__antithesis_instrumentation__.Notify(461824)

					return nil
				}
				__antithesis_instrumentation__.Notify(461819)

				return catalog.ForEachIndex(table, catalog.IndexOpts{NonPhysicalPrimaryIndex: true}, reportIndex)
			})
	},
}

var crdbInternalBackwardDependenciesTable = virtualSchemaTable{
	comment: "backward inter-descriptor dependencies starting from tables accessible by current user in current database (KV scan)",
	schema: `
CREATE TABLE crdb_internal.backward_dependencies (
  descriptor_id      INT,
  descriptor_name    STRING NOT NULL,
  index_id           INT,
  column_id          INT,
  dependson_id       INT NOT NULL,
  dependson_type     STRING NOT NULL,
  dependson_index_id INT,
  dependson_name     STRING,
  dependson_details  STRING
)
`,
	populate: func(
		ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor,
		addRow func(...tree.Datum) error,
	) error {
		__antithesis_instrumentation__.Notify(461843)
		fkDep := tree.NewDString("fk")
		viewDep := tree.NewDString("view")
		sequenceDep := tree.NewDString("sequence")

		return forEachTableDescAllWithTableLookup(ctx, p, dbContext, hideVirtual, func(
			db catalog.DatabaseDescriptor, _ string, table catalog.TableDescriptor, tableLookup tableLookupFn,
		) error {
			__antithesis_instrumentation__.Notify(461844)
			tableID := tree.NewDInt(tree.DInt(table.GetID()))
			tableName := tree.NewDString(table.GetName())

			if err := table.ForeachOutboundFK(func(fk *descpb.ForeignKeyConstraint) error {
				__antithesis_instrumentation__.Notify(461849)
				refTbl, err := tableLookup.getTableByID(fk.ReferencedTableID)
				if err != nil {
					__antithesis_instrumentation__.Notify(461853)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461854)
				}
				__antithesis_instrumentation__.Notify(461850)
				refConstraint, err := tabledesc.FindFKReferencedUniqueConstraint(refTbl, fk.ReferencedColumnIDs)
				if err != nil {
					__antithesis_instrumentation__.Notify(461855)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461856)
				}
				__antithesis_instrumentation__.Notify(461851)
				var refIdxID descpb.IndexID
				if refIdx, ok := refConstraint.(*descpb.IndexDescriptor); ok {
					__antithesis_instrumentation__.Notify(461857)
					refIdxID = refIdx.ID
				} else {
					__antithesis_instrumentation__.Notify(461858)
				}
				__antithesis_instrumentation__.Notify(461852)
				return addRow(
					tableID, tableName,
					tree.DNull,
					tree.DNull,
					tree.NewDInt(tree.DInt(fk.ReferencedTableID)),
					fkDep,
					tree.NewDInt(tree.DInt(refIdxID)),
					tree.NewDString(fk.Name),
					tree.DNull,
				)
			}); err != nil {
				__antithesis_instrumentation__.Notify(461859)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461860)
			}
			__antithesis_instrumentation__.Notify(461845)

			for _, tIdx := range table.GetDependsOn() {
				__antithesis_instrumentation__.Notify(461861)
				if err := addRow(
					tableID, tableName,
					tree.DNull,
					tree.DNull,
					tree.NewDInt(tree.DInt(tIdx)),
					viewDep,
					tree.DNull,
					tree.DNull,
					tree.DNull,
				); err != nil {
					__antithesis_instrumentation__.Notify(461862)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461863)
				}
			}
			__antithesis_instrumentation__.Notify(461846)
			for _, tIdx := range table.GetDependsOnTypes() {
				__antithesis_instrumentation__.Notify(461864)
				if err := addRow(
					tableID, tableName,
					tree.DNull,
					tree.DNull,
					tree.NewDInt(tree.DInt(tIdx)),
					viewDep,
					tree.DNull,
					tree.DNull,
					tree.DNull,
				); err != nil {
					__antithesis_instrumentation__.Notify(461865)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461866)
				}
			}
			__antithesis_instrumentation__.Notify(461847)

			for _, col := range table.PublicColumns() {
				__antithesis_instrumentation__.Notify(461867)
				for i := 0; i < col.NumUsesSequences(); i++ {
					__antithesis_instrumentation__.Notify(461868)
					sequenceID := col.GetUsesSequenceID(i)
					if err := addRow(
						tableID, tableName,
						tree.DNull,
						tree.NewDInt(tree.DInt(col.GetID())),
						tree.NewDInt(tree.DInt(sequenceID)),
						sequenceDep,
						tree.DNull,
						tree.DNull,
						tree.DNull,
					); err != nil {
						__antithesis_instrumentation__.Notify(461869)
						return err
					} else {
						__antithesis_instrumentation__.Notify(461870)
					}
				}
			}
			__antithesis_instrumentation__.Notify(461848)
			return nil
		})
	},
}

var crdbInternalFeatureUsage = virtualSchemaTable{
	comment: "telemetry counters (RAM; local node only)",
	schema: `
CREATE TABLE crdb_internal.feature_usage (
  feature_name          STRING NOT NULL,
  usage_count           INT NOT NULL
)
`,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461871)
		for feature, count := range telemetry.GetFeatureCounts(telemetry.Raw, telemetry.ReadOnly) {
			__antithesis_instrumentation__.Notify(461873)
			if count == 0 {
				__antithesis_instrumentation__.Notify(461875)

				continue
			} else {
				__antithesis_instrumentation__.Notify(461876)
			}
			__antithesis_instrumentation__.Notify(461874)
			if err := addRow(
				tree.NewDString(feature),
				tree.NewDInt(tree.DInt(int64(count))),
			); err != nil {
				__antithesis_instrumentation__.Notify(461877)
				return err
			} else {
				__antithesis_instrumentation__.Notify(461878)
			}
		}
		__antithesis_instrumentation__.Notify(461872)
		return nil
	},
}

var crdbInternalForwardDependenciesTable = virtualSchemaTable{
	comment: "forward inter-descriptor dependencies starting from tables accessible by current user in current database (KV scan)",
	schema: `
CREATE TABLE crdb_internal.forward_dependencies (
  descriptor_id         INT,
  descriptor_name       STRING NOT NULL,
  index_id              INT,
  dependedonby_id       INT NOT NULL,
  dependedonby_type     STRING NOT NULL,
  dependedonby_index_id INT,
  dependedonby_name     STRING,
  dependedonby_details  STRING
)
`,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461879)
		fkDep := tree.NewDString("fk")
		viewDep := tree.NewDString("view")
		sequenceDep := tree.NewDString("sequence")
		return forEachTableDescAll(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, _ string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(461880)
				tableID := tree.NewDInt(tree.DInt(table.GetID()))
				tableName := tree.NewDString(table.GetName())

				if err := table.ForeachInboundFK(func(fk *descpb.ForeignKeyConstraint) error {
					__antithesis_instrumentation__.Notify(461884)
					return addRow(
						tableID, tableName,
						tree.DNull,
						tree.NewDInt(tree.DInt(fk.OriginTableID)),
						fkDep,
						tree.DNull,
						tree.DNull,
						tree.DNull,
					)
				}); err != nil {
					__antithesis_instrumentation__.Notify(461885)
					return err
				} else {
					__antithesis_instrumentation__.Notify(461886)
				}
				__antithesis_instrumentation__.Notify(461881)

				reportDependedOnBy := func(
					dep *descpb.TableDescriptor_Reference, depTypeString *tree.DString,
				) error {
					__antithesis_instrumentation__.Notify(461887)
					return addRow(
						tableID, tableName,
						tree.DNull,
						tree.NewDInt(tree.DInt(dep.ID)),
						depTypeString,
						tree.NewDInt(tree.DInt(dep.IndexID)),
						tree.DNull,
						tree.NewDString(fmt.Sprintf("Columns: %v", dep.ColumnIDs)),
					)
				}
				__antithesis_instrumentation__.Notify(461882)

				if table.IsTable() || func() bool {
					__antithesis_instrumentation__.Notify(461888)
					return table.IsView() == true
				}() == true {
					__antithesis_instrumentation__.Notify(461889)
					return table.ForeachDependedOnBy(func(dep *descpb.TableDescriptor_Reference) error {
						__antithesis_instrumentation__.Notify(461890)
						return reportDependedOnBy(dep, viewDep)
					})
				} else {
					__antithesis_instrumentation__.Notify(461891)
					if table.IsSequence() {
						__antithesis_instrumentation__.Notify(461892)
						return table.ForeachDependedOnBy(func(dep *descpb.TableDescriptor_Reference) error {
							__antithesis_instrumentation__.Notify(461893)
							return reportDependedOnBy(dep, sequenceDep)
						})
					} else {
						__antithesis_instrumentation__.Notify(461894)
					}
				}
				__antithesis_instrumentation__.Notify(461883)
				return nil
			})
	},
}

var crdbInternalRangesView = virtualSchemaView{
	schema: `
CREATE VIEW crdb_internal.ranges AS SELECT
	range_id,
	start_key,
	start_pretty,
	end_key,
	end_pretty,
  table_id,
	database_name,
  schema_name,
	table_name,
	index_name,
	replicas,
	replica_localities,
	voting_replicas,
	non_voting_replicas,
	learner_replicas,
	split_enforced_until,
	crdb_internal.lease_holder(start_key) AS lease_holder,
	(crdb_internal.range_stats(start_key)->>'key_bytes')::INT +
	(crdb_internal.range_stats(start_key)->>'val_bytes')::INT AS range_size
FROM crdb_internal.ranges_no_leases
`,
	resultColumns: colinfo.ResultColumns{
		{Name: "range_id", Typ: types.Int},
		{Name: "start_key", Typ: types.Bytes},
		{Name: "start_pretty", Typ: types.String},
		{Name: "end_key", Typ: types.Bytes},
		{Name: "end_pretty", Typ: types.String},
		{Name: "table_id", Typ: types.Int},
		{Name: "database_name", Typ: types.String},
		{Name: "schema_name", Typ: types.String},
		{Name: "table_name", Typ: types.String},
		{Name: "index_name", Typ: types.String},
		{Name: "replicas", Typ: types.Int2Vector},
		{Name: "replica_localities", Typ: types.StringArray},
		{Name: "voting_replicas", Typ: types.Int2Vector},
		{Name: "non_voting_replicas", Typ: types.Int2Vector},
		{Name: "learner_replicas", Typ: types.Int2Vector},
		{Name: "split_enforced_until", Typ: types.Timestamp},
		{Name: "lease_holder", Typ: types.Int},
		{Name: "range_size", Typ: types.Int},
	},
}

func descriptorsByType(
	descs []catalog.Descriptor, privCheckerFunc func(desc catalog.Descriptor) bool,
) (
	hasPermission bool,
	dbNames map[uint32]string,
	tableNames map[uint32]string,
	schemaNames map[uint32]string,
	indexNames map[uint32]map[uint32]string,
	schemaParents map[uint32]uint32,
	parents map[uint32]uint32,
) {
	__antithesis_instrumentation__.Notify(461895)

	dbNames = make(map[uint32]string)
	tableNames = make(map[uint32]string)
	schemaNames = make(map[uint32]string)
	indexNames = make(map[uint32]map[uint32]string)
	schemaParents = make(map[uint32]uint32)
	parents = make(map[uint32]uint32)
	hasPermission = false
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(461897)
		id := uint32(desc.GetID())
		if !privCheckerFunc(desc) {
			__antithesis_instrumentation__.Notify(461899)
			continue
		} else {
			__antithesis_instrumentation__.Notify(461900)
		}
		__antithesis_instrumentation__.Notify(461898)
		hasPermission = true
		switch desc := desc.(type) {
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(461901)
			parents[id] = uint32(desc.GetParentID())
			schemaParents[id] = uint32(desc.GetParentSchemaID())
			tableNames[id] = desc.GetName()
			indexNames[id] = make(map[uint32]string)
			for _, idx := range desc.PublicNonPrimaryIndexes() {
				__antithesis_instrumentation__.Notify(461904)
				indexNames[id][uint32(idx.GetID())] = idx.GetName()
			}
		case catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(461902)
			dbNames[id] = desc.GetName()
		case catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(461903)
			schemaNames[id] = desc.GetName()
		}
	}
	__antithesis_instrumentation__.Notify(461896)

	return hasPermission, dbNames, tableNames, schemaNames, indexNames, schemaParents, parents
}

func lookupNamesByKey(
	p *planner,
	key roachpb.Key,
	dbNames, tableNames, schemaNames map[uint32]string,
	indexNames map[uint32]map[uint32]string,
	schemaParents, parents map[uint32]uint32,
) (tableID uint32, dbName string, schemaName string, tableName string, indexName string) {
	__antithesis_instrumentation__.Notify(461905)
	var err error
	if _, tableID, err = p.ExecCfg().Codec.DecodeTablePrefix(key); err == nil {
		__antithesis_instrumentation__.Notify(461907)
		schemaParent := schemaParents[tableID]
		if schemaParent != 0 {
			__antithesis_instrumentation__.Notify(461909)
			schemaName = schemaNames[schemaParent]
		} else {
			__antithesis_instrumentation__.Notify(461910)

			schemaName = string(tree.PublicSchemaName)
		}
		__antithesis_instrumentation__.Notify(461908)
		parent := parents[tableID]
		if parent != 0 {
			__antithesis_instrumentation__.Notify(461911)
			tableName = tableNames[tableID]
			dbName = dbNames[parent]
			if _, _, idxID, err := p.ExecCfg().Codec.DecodeIndexPrefix(key); err == nil {
				__antithesis_instrumentation__.Notify(461912)
				indexName = indexNames[tableID][idxID]
			} else {
				__antithesis_instrumentation__.Notify(461913)
			}
		} else {
			__antithesis_instrumentation__.Notify(461914)
			dbName = dbNames[tableID]
		}
	} else {
		__antithesis_instrumentation__.Notify(461915)
	}
	__antithesis_instrumentation__.Notify(461906)

	return tableID, dbName, schemaName, tableName, indexName
}

var crdbInternalRangesNoLeasesTable = virtualSchemaTable{
	comment: `range metadata without leaseholder details (KV join; expensive!)`,

	schema: `
CREATE TABLE crdb_internal.ranges_no_leases (
  range_id             INT NOT NULL,
  start_key            BYTES NOT NULL,
  start_pretty         STRING NOT NULL,
  end_key              BYTES NOT NULL,
  end_pretty           STRING NOT NULL,
  table_id             INT NOT NULL,
  database_name        STRING NOT NULL,
  schema_name          STRING NOT NULL,
  table_name           STRING NOT NULL,
  index_name           STRING NOT NULL,
  replicas             INT[] NOT NULL,
  replica_localities   STRING[] NOT NULL,
  voting_replicas      INT[] NOT NULL,
  non_voting_replicas  INT[] NOT NULL,
  learner_replicas     INT[] NOT NULL,
  split_enforced_until TIMESTAMP
)
`,
	generator: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, _ *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(461916)
		hasAdmin, err := p.HasAdminRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461924)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(461925)
		}
		__antithesis_instrumentation__.Notify(461917)
		all, err := p.Descriptors().GetAllDescriptors(ctx, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(461926)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(461927)
		}
		__antithesis_instrumentation__.Notify(461918)
		descs := all.OrderedDescriptors()

		privCheckerFunc := func(desc catalog.Descriptor) bool {
			__antithesis_instrumentation__.Notify(461928)
			if hasAdmin {
				__antithesis_instrumentation__.Notify(461930)
				return true
			} else {
				__antithesis_instrumentation__.Notify(461931)
			}
			__antithesis_instrumentation__.Notify(461929)

			return p.CheckPrivilege(ctx, desc, privilege.ZONECONFIG) == nil
		}
		__antithesis_instrumentation__.Notify(461919)

		hasPermission, dbNames, tableNames, schemaNames, indexNames, schemaParents, parents :=
			descriptorsByType(descs, privCheckerFunc)

		if !hasPermission {
			__antithesis_instrumentation__.Notify(461932)
			return nil, nil, pgerror.Newf(pgcode.InsufficientPrivilege, "only users with the ZONECONFIG privilege or the admin role can read crdb_internal.ranges_no_leases")
		} else {
			__antithesis_instrumentation__.Notify(461933)
		}
		__antithesis_instrumentation__.Notify(461920)
		ranges, err := kvclient.ScanMetaKVs(ctx, p.txn, roachpb.Span{
			Key:    keys.MinKey,
			EndKey: keys.MaxKey,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(461934)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(461935)
		}
		__antithesis_instrumentation__.Notify(461921)

		descriptors, err := getAllNodeDescriptors(p)
		if err != nil {
			__antithesis_instrumentation__.Notify(461936)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(461937)
		}
		__antithesis_instrumentation__.Notify(461922)
		nodeIDToLocality := make(map[roachpb.NodeID]roachpb.Locality)
		for _, desc := range descriptors {
			__antithesis_instrumentation__.Notify(461938)
			nodeIDToLocality[desc.NodeID] = desc.Locality
		}
		__antithesis_instrumentation__.Notify(461923)

		var desc roachpb.RangeDescriptor

		i := 0

		return func() (tree.Datums, error) {
			__antithesis_instrumentation__.Notify(461939)
			if i >= len(ranges) {
				__antithesis_instrumentation__.Notify(461950)
				return nil, nil
			} else {
				__antithesis_instrumentation__.Notify(461951)
			}
			__antithesis_instrumentation__.Notify(461940)

			r := ranges[i]
			i++

			if err := r.ValueProto(&desc); err != nil {
				__antithesis_instrumentation__.Notify(461952)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(461953)
			}
			__antithesis_instrumentation__.Notify(461941)

			votersAndNonVoters := append([]roachpb.ReplicaDescriptor(nil),
				desc.Replicas().VoterAndNonVoterDescriptors()...)
			var learnerReplicaStoreIDs []int
			for _, rd := range desc.Replicas().LearnerDescriptors() {
				__antithesis_instrumentation__.Notify(461954)
				learnerReplicaStoreIDs = append(learnerReplicaStoreIDs, int(rd.StoreID))
			}
			__antithesis_instrumentation__.Notify(461942)
			sort.Slice(votersAndNonVoters, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(461955)
				return votersAndNonVoters[i].StoreID < votersAndNonVoters[j].StoreID
			})
			__antithesis_instrumentation__.Notify(461943)
			sort.Ints(learnerReplicaStoreIDs)
			votersAndNonVotersArr := tree.NewDArray(types.Int)
			for _, replica := range votersAndNonVoters {
				__antithesis_instrumentation__.Notify(461956)
				if err := votersAndNonVotersArr.Append(tree.NewDInt(tree.DInt(replica.StoreID))); err != nil {
					__antithesis_instrumentation__.Notify(461957)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(461958)
				}
			}
			__antithesis_instrumentation__.Notify(461944)
			votersArr := tree.NewDArray(types.Int)
			for _, replica := range desc.Replicas().VoterDescriptors() {
				__antithesis_instrumentation__.Notify(461959)
				if err := votersArr.Append(tree.NewDInt(tree.DInt(replica.StoreID))); err != nil {
					__antithesis_instrumentation__.Notify(461960)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(461961)
				}
			}
			__antithesis_instrumentation__.Notify(461945)
			nonVotersArr := tree.NewDArray(types.Int)
			for _, replica := range desc.Replicas().NonVoterDescriptors() {
				__antithesis_instrumentation__.Notify(461962)
				if err := nonVotersArr.Append(tree.NewDInt(tree.DInt(replica.StoreID))); err != nil {
					__antithesis_instrumentation__.Notify(461963)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(461964)
				}
			}
			__antithesis_instrumentation__.Notify(461946)
			learnersArr := tree.NewDArray(types.Int)
			for _, replica := range learnerReplicaStoreIDs {
				__antithesis_instrumentation__.Notify(461965)
				if err := learnersArr.Append(tree.NewDInt(tree.DInt(replica))); err != nil {
					__antithesis_instrumentation__.Notify(461966)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(461967)
				}
			}
			__antithesis_instrumentation__.Notify(461947)

			replicaLocalityArr := tree.NewDArray(types.String)
			for _, replica := range votersAndNonVoters {
				__antithesis_instrumentation__.Notify(461968)
				replicaLocality := nodeIDToLocality[replica.NodeID].String()
				if err := replicaLocalityArr.Append(tree.NewDString(replicaLocality)); err != nil {
					__antithesis_instrumentation__.Notify(461969)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(461970)
				}
			}
			__antithesis_instrumentation__.Notify(461948)

			tableID, dbName, schemaName, tableName, indexName := lookupNamesByKey(
				p, desc.StartKey.AsRawKey(), dbNames, tableNames, schemaNames,
				indexNames, schemaParents, parents,
			)

			splitEnforcedUntil := tree.DNull
			if !desc.GetStickyBit().IsEmpty() {
				__antithesis_instrumentation__.Notify(461971)
				splitEnforcedUntil = tree.TimestampToInexactDTimestamp(*desc.StickyBit)
			} else {
				__antithesis_instrumentation__.Notify(461972)
			}
			__antithesis_instrumentation__.Notify(461949)

			return tree.Datums{
				tree.NewDInt(tree.DInt(desc.RangeID)),
				tree.NewDBytes(tree.DBytes(desc.StartKey)),
				tree.NewDString(keys.PrettyPrint(nil, desc.StartKey.AsRawKey())),
				tree.NewDBytes(tree.DBytes(desc.EndKey)),
				tree.NewDString(keys.PrettyPrint(nil, desc.EndKey.AsRawKey())),
				tree.NewDInt(tree.DInt(tableID)),
				tree.NewDString(dbName),
				tree.NewDString(schemaName),
				tree.NewDString(tableName),
				tree.NewDString(indexName),
				votersAndNonVotersArr,
				replicaLocalityArr,
				votersArr,
				nonVotersArr,
				learnersArr,
				splitEnforcedUntil,
			}, nil
		}, nil, nil
	},
}

func (p *planner) getAllNames(ctx context.Context) (map[descpb.ID]catalog.NameKey, error) {
	__antithesis_instrumentation__.Notify(461973)
	return getAllNames(ctx, p.txn, p.ExtendedEvalContext().ExecCfg.InternalExecutor)
}

func TestingGetAllNames(
	ctx context.Context, txn *kv.Txn, executor *InternalExecutor,
) (map[descpb.ID]catalog.NameKey, error) {
	__antithesis_instrumentation__.Notify(461974)
	return getAllNames(ctx, txn, executor)
}

func getAllNames(
	ctx context.Context, txn *kv.Txn, executor *InternalExecutor,
) (map[descpb.ID]catalog.NameKey, error) {
	__antithesis_instrumentation__.Notify(461975)
	namespace := map[descpb.ID]catalog.NameKey{}
	it, err := executor.QueryIterator(
		ctx, "get-all-names", txn,
		`SELECT id, "parentID", "parentSchemaID", name FROM system.namespace`,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(461979)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(461980)
	}
	__antithesis_instrumentation__.Notify(461976)
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(461981)
		r := it.Cur()
		id, parentID, parentSchemaID, name := tree.MustBeDInt(r[0]), tree.MustBeDInt(r[1]), tree.MustBeDInt(r[2]), tree.MustBeDString(r[3])
		namespace[descpb.ID(id)] = descpb.NameInfo{
			ParentID:       descpb.ID(parentID),
			ParentSchemaID: descpb.ID(parentSchemaID),
			Name:           string(name),
		}
	}
	__antithesis_instrumentation__.Notify(461977)
	if err != nil {
		__antithesis_instrumentation__.Notify(461982)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(461983)
	}
	__antithesis_instrumentation__.Notify(461978)

	return namespace, nil
}

var crdbInternalZonesTable = virtualSchemaTable{
	comment: "decoded zone configurations from system.zones (KV scan)",
	schema: `
CREATE TABLE crdb_internal.zones (
  zone_id              INT NOT NULL,
  subzone_id           INT NOT NULL,
  target               STRING,
  range_name           STRING,
  database_name        STRING,
  schema_name          STRING,
  table_name           STRING,
  index_name           STRING,
  partition_name       STRING,
  raw_config_yaml      STRING NOT NULL,
  raw_config_sql       STRING, -- this column can be NULL if there is no specifier syntax
                               -- possible (e.g. the object was deleted).
	raw_config_protobuf  BYTES NOT NULL,
	full_config_yaml     STRING NOT NULL,
	full_config_sql      STRING
)
`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(461984)
		namespace, err := p.getAllNames(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(461990)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461991)
		}
		__antithesis_instrumentation__.Notify(461985)
		resolveID := func(id uint32) (parentID, parentSchemaID uint32, name string, err error) {
			__antithesis_instrumentation__.Notify(461992)

			if id == keys.PublicSchemaID {
				__antithesis_instrumentation__.Notify(461995)
				return 0, 0, string(tree.PublicSchemaName), nil
			} else {
				__antithesis_instrumentation__.Notify(461996)
			}
			__antithesis_instrumentation__.Notify(461993)
			if entry, ok := namespace[descpb.ID(id)]; ok {
				__antithesis_instrumentation__.Notify(461997)
				return uint32(entry.GetParentID()), uint32(entry.GetParentSchemaID()), entry.GetName(), nil
			} else {
				__antithesis_instrumentation__.Notify(461998)
			}
			__antithesis_instrumentation__.Notify(461994)
			return 0, 0, "", errors.AssertionFailedf(
				"object with ID %d does not exist", errors.Safe(id))
		}
		__antithesis_instrumentation__.Notify(461986)

		getKey := func(key roachpb.Key) (*roachpb.Value, error) {
			__antithesis_instrumentation__.Notify(461999)
			kv, err := p.txn.Get(ctx, key)
			if err != nil {
				__antithesis_instrumentation__.Notify(462001)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(462002)
			}
			__antithesis_instrumentation__.Notify(462000)
			return kv.Value, nil
		}
		__antithesis_instrumentation__.Notify(461987)

		rows, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryBuffered(
			ctx, "crdb-internal-zones-table", p.txn, `SELECT id, config FROM system.zones`)
		if err != nil {
			__antithesis_instrumentation__.Notify(462003)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462004)
		}
		__antithesis_instrumentation__.Notify(461988)
		values := make(tree.Datums, len(showZoneConfigColumns))
		for _, r := range rows {
			__antithesis_instrumentation__.Notify(462005)
			id := uint32(tree.MustBeDInt(r[0]))

			var zoneSpecifier *tree.ZoneSpecifier
			zs, err := zonepb.ZoneSpecifierFromID(id, resolveID)
			if err != nil {
				__antithesis_instrumentation__.Notify(462011)

				continue
			} else {
				__antithesis_instrumentation__.Notify(462012)
				zoneSpecifier = &zs
			}
			__antithesis_instrumentation__.Notify(462006)

			configBytes := []byte(*r[1].(*tree.DBytes))
			var configProto zonepb.ZoneConfig
			if err := protoutil.Unmarshal(configBytes, &configProto); err != nil {
				__antithesis_instrumentation__.Notify(462013)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462014)
			}
			__antithesis_instrumentation__.Notify(462007)
			subzones := configProto.Subzones

			fullZone := configProto
			if err := completeZoneConfig(
				&fullZone, p.ExecCfg().Codec, descpb.ID(tree.MustBeDInt(r[0])), getKey,
			); err != nil {
				__antithesis_instrumentation__.Notify(462015)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462016)
			}
			__antithesis_instrumentation__.Notify(462008)

			var table catalog.TableDescriptor
			if zs.Database != "" {
				__antithesis_instrumentation__.Notify(462017)
				_, database, err := p.Descriptors().GetImmutableDatabaseByID(ctx, p.txn, descpb.ID(id), tree.DatabaseLookupFlags{
					Required:    true,
					AvoidLeased: true,
				})
				if err != nil {
					__antithesis_instrumentation__.Notify(462019)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462020)
				}
				__antithesis_instrumentation__.Notify(462018)
				if p.CheckAnyPrivilege(ctx, database) != nil {
					__antithesis_instrumentation__.Notify(462021)
					continue
				} else {
					__antithesis_instrumentation__.Notify(462022)
				}
			} else {
				__antithesis_instrumentation__.Notify(462023)
				if zoneSpecifier.TableOrIndex.Table.ObjectName != "" {
					__antithesis_instrumentation__.Notify(462024)
					tableEntry, err := p.LookupTableByID(ctx, descpb.ID(id))
					if err != nil {
						__antithesis_instrumentation__.Notify(462027)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462028)
					}
					__antithesis_instrumentation__.Notify(462025)
					if p.CheckAnyPrivilege(ctx, tableEntry) != nil {
						__antithesis_instrumentation__.Notify(462029)
						continue
					} else {
						__antithesis_instrumentation__.Notify(462030)
					}
					__antithesis_instrumentation__.Notify(462026)
					table = tableEntry
				} else {
					__antithesis_instrumentation__.Notify(462031)
				}
			}
			__antithesis_instrumentation__.Notify(462009)

			if !configProto.IsSubzonePlaceholder() {
				__antithesis_instrumentation__.Notify(462032)

				configProto.Subzones = nil
				configProto.SubzoneSpans = nil

				if err := generateZoneConfigIntrospectionValues(
					values,
					r[0],
					tree.NewDInt(tree.DInt(0)),
					zoneSpecifier,
					&configProto,
					&fullZone,
				); err != nil {
					__antithesis_instrumentation__.Notify(462034)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462035)
				}
				__antithesis_instrumentation__.Notify(462033)

				if err := addRow(values...); err != nil {
					__antithesis_instrumentation__.Notify(462036)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462037)
				}
			} else {
				__antithesis_instrumentation__.Notify(462038)
			}
			__antithesis_instrumentation__.Notify(462010)

			if len(subzones) > 0 {
				__antithesis_instrumentation__.Notify(462039)
				if table == nil {
					__antithesis_instrumentation__.Notify(462041)
					return errors.AssertionFailedf(
						"object id %d with #subzones %d is not a table",
						id,
						len(subzones),
					)
				} else {
					__antithesis_instrumentation__.Notify(462042)
				}
				__antithesis_instrumentation__.Notify(462040)

				for i, s := range subzones {
					__antithesis_instrumentation__.Notify(462043)
					index := catalog.FindActiveIndex(table, func(idx catalog.Index) bool {
						__antithesis_instrumentation__.Notify(462049)
						return idx.GetID() == descpb.IndexID(s.IndexID)
					})
					__antithesis_instrumentation__.Notify(462044)
					if index == nil {
						__antithesis_instrumentation__.Notify(462050)

						continue
					} else {
						__antithesis_instrumentation__.Notify(462051)
					}
					__antithesis_instrumentation__.Notify(462045)
					if zoneSpecifier != nil {
						__antithesis_instrumentation__.Notify(462052)
						zs := zs
						zs.TableOrIndex.Index = tree.UnrestrictedName(index.GetName())
						zs.Partition = tree.Name(s.PartitionName)
						zoneSpecifier = &zs
					} else {
						__antithesis_instrumentation__.Notify(462053)
					}
					__antithesis_instrumentation__.Notify(462046)

					subZoneConfig := s.Config

					if s.PartitionName == "" {
						__antithesis_instrumentation__.Notify(462054)
						subZoneConfig.InheritFromParent(&fullZone)
					} else {
						__antithesis_instrumentation__.Notify(462055)

						if indexSubzone := fullZone.GetSubzone(uint32(index.GetID()), ""); indexSubzone != nil {
							__antithesis_instrumentation__.Notify(462057)
							subZoneConfig.InheritFromParent(&indexSubzone.Config)
						} else {
							__antithesis_instrumentation__.Notify(462058)
						}
						__antithesis_instrumentation__.Notify(462056)

						subZoneConfig.InheritFromParent(&fullZone)
					}
					__antithesis_instrumentation__.Notify(462047)

					if err := generateZoneConfigIntrospectionValues(
						values,
						r[0],
						tree.NewDInt(tree.DInt(i+1)),
						zoneSpecifier,
						&s.Config,
						&subZoneConfig,
					); err != nil {
						__antithesis_instrumentation__.Notify(462059)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462060)
					}
					__antithesis_instrumentation__.Notify(462048)

					if err := addRow(values...); err != nil {
						__antithesis_instrumentation__.Notify(462061)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462062)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(462063)
			}
		}
		__antithesis_instrumentation__.Notify(461989)
		return nil
	},
}

func getAllNodeDescriptors(p *planner) ([]roachpb.NodeDescriptor, error) {
	__antithesis_instrumentation__.Notify(462064)
	g, err := p.ExecCfg().Gossip.OptionalErr(47899)
	if err != nil {
		__antithesis_instrumentation__.Notify(462067)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462068)
	}
	__antithesis_instrumentation__.Notify(462065)
	var descriptors []roachpb.NodeDescriptor
	if err := g.IterateInfos(gossip.KeyNodeIDPrefix, func(key string, i gossip.Info) error {
		__antithesis_instrumentation__.Notify(462069)
		bytes, err := i.Value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(462073)
			return errors.NewAssertionErrorWithWrappedErrf(err,
				"failed to extract bytes for key %q", key)
		} else {
			__antithesis_instrumentation__.Notify(462074)
		}
		__antithesis_instrumentation__.Notify(462070)

		var d roachpb.NodeDescriptor
		if err := protoutil.Unmarshal(bytes, &d); err != nil {
			__antithesis_instrumentation__.Notify(462075)
			return errors.NewAssertionErrorWithWrappedErrf(err,
				"failed to parse value for key %q", key)
		} else {
			__antithesis_instrumentation__.Notify(462076)
		}
		__antithesis_instrumentation__.Notify(462071)

		if d.NodeID != 0 {
			__antithesis_instrumentation__.Notify(462077)
			descriptors = append(descriptors, d)
		} else {
			__antithesis_instrumentation__.Notify(462078)
		}
		__antithesis_instrumentation__.Notify(462072)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(462079)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462080)
	}
	__antithesis_instrumentation__.Notify(462066)
	return descriptors, nil
}

var crdbInternalGossipNodesTable = virtualSchemaTable{
	comment: "locally known gossiped node details (RAM; local node only)",
	schema: `
CREATE TABLE crdb_internal.gossip_nodes (
  node_id               INT NOT NULL,
  network               STRING NOT NULL,
  address               STRING NOT NULL,
  advertise_address     STRING NOT NULL,
  sql_network           STRING NOT NULL,
  sql_address           STRING NOT NULL,
  advertise_sql_address STRING NOT NULL,
  attrs                 JSON NOT NULL,
  locality              STRING NOT NULL,
  cluster_name          STRING NOT NULL,
  server_version        STRING NOT NULL,
  build_tag             STRING NOT NULL,
  started_at            TIMESTAMP NOT NULL,
  is_live               BOOL NOT NULL,
  ranges                INT NOT NULL,
  leases                INT NOT NULL
)
	`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462081)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.gossip_nodes"); err != nil {
			__antithesis_instrumentation__.Notify(462089)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462090)
		}
		__antithesis_instrumentation__.Notify(462082)

		g, err := p.ExecCfg().Gossip.OptionalErr(47899)
		if err != nil {
			__antithesis_instrumentation__.Notify(462091)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462092)
		}
		__antithesis_instrumentation__.Notify(462083)

		descriptors, err := getAllNodeDescriptors(p)
		if err != nil {
			__antithesis_instrumentation__.Notify(462093)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462094)
		}
		__antithesis_instrumentation__.Notify(462084)

		alive := make(map[roachpb.NodeID]tree.DBool)
		for _, d := range descriptors {
			__antithesis_instrumentation__.Notify(462095)
			if _, err := g.GetInfo(gossip.MakeGossipClientsKey(d.NodeID)); err == nil {
				__antithesis_instrumentation__.Notify(462096)
				alive[d.NodeID] = true
			} else {
				__antithesis_instrumentation__.Notify(462097)
			}
		}
		__antithesis_instrumentation__.Notify(462085)

		sort.Slice(descriptors, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(462098)
			return descriptors[i].NodeID < descriptors[j].NodeID
		})
		__antithesis_instrumentation__.Notify(462086)

		type nodeStats struct {
			ranges int32
			leases int32
		}

		stats := make(map[roachpb.NodeID]nodeStats)
		if err := g.IterateInfos(gossip.KeyStorePrefix, func(key string, i gossip.Info) error {
			__antithesis_instrumentation__.Notify(462099)
			bytes, err := i.Value.GetBytes()
			if err != nil {
				__antithesis_instrumentation__.Notify(462102)
				return errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to extract bytes for key %q", key)
			} else {
				__antithesis_instrumentation__.Notify(462103)
			}
			__antithesis_instrumentation__.Notify(462100)

			var desc roachpb.StoreDescriptor
			if err := protoutil.Unmarshal(bytes, &desc); err != nil {
				__antithesis_instrumentation__.Notify(462104)
				return errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to parse value for key %q", key)
			} else {
				__antithesis_instrumentation__.Notify(462105)
			}
			__antithesis_instrumentation__.Notify(462101)

			s := stats[desc.Node.NodeID]
			s.ranges += desc.Capacity.RangeCount
			s.leases += desc.Capacity.LeaseCount
			stats[desc.Node.NodeID] = s
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(462106)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462107)
		}
		__antithesis_instrumentation__.Notify(462087)

		for _, d := range descriptors {
			__antithesis_instrumentation__.Notify(462108)
			attrs := json.NewArrayBuilder(len(d.Attrs.Attrs))
			for _, a := range d.Attrs.Attrs {
				__antithesis_instrumentation__.Notify(462113)
				attrs.Add(json.FromString(a))
			}
			__antithesis_instrumentation__.Notify(462109)

			listenAddrRPC := d.Address
			listenAddrSQL := d.CheckedSQLAddress()

			advAddrRPC, err := g.GetNodeIDAddress(d.NodeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(462114)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462115)
			}
			__antithesis_instrumentation__.Notify(462110)
			advAddrSQL, err := g.GetNodeIDSQLAddress(d.NodeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(462116)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462117)
			}
			__antithesis_instrumentation__.Notify(462111)

			startTSDatum, err := tree.MakeDTimestamp(timeutil.Unix(0, d.StartedAt), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(462118)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462119)
			}
			__antithesis_instrumentation__.Notify(462112)
			if err := addRow(
				tree.NewDInt(tree.DInt(d.NodeID)),
				tree.NewDString(listenAddrRPC.NetworkField),
				tree.NewDString(listenAddrRPC.AddressField),
				tree.NewDString(advAddrRPC.String()),
				tree.NewDString(listenAddrSQL.NetworkField),
				tree.NewDString(listenAddrSQL.AddressField),
				tree.NewDString(advAddrSQL.String()),
				tree.NewDJSON(attrs.Build()),
				tree.NewDString(d.Locality.String()),
				tree.NewDString(d.ClusterName),
				tree.NewDString(d.ServerVersion.String()),
				tree.NewDString(d.BuildTag),
				startTSDatum,
				tree.MakeDBool(alive[d.NodeID]),
				tree.NewDInt(tree.DInt(stats[d.NodeID].ranges)),
				tree.NewDInt(tree.DInt(stats[d.NodeID].leases)),
			); err != nil {
				__antithesis_instrumentation__.Notify(462120)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462121)
			}
		}
		__antithesis_instrumentation__.Notify(462088)
		return nil
	},
}

var crdbInternalKVNodeLivenessTable = virtualSchemaTable{
	comment: "node liveness status, as seen by kv",
	schema: `
CREATE TABLE crdb_internal.kv_node_liveness (
  node_id          INT NOT NULL,
  epoch            INT NOT NULL,
  expiration       STRING NOT NULL,
  draining         BOOL NOT NULL,
  membership       STRING NOT NULL
)
	`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462122)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.node_liveness"); err != nil {
			__antithesis_instrumentation__.Notify(462128)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462129)
		}
		__antithesis_instrumentation__.Notify(462123)

		nl, err := p.ExecCfg().NodeLiveness.OptionalErr(47900)
		if err != nil {
			__antithesis_instrumentation__.Notify(462130)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462131)
		}
		__antithesis_instrumentation__.Notify(462124)

		livenesses, err := nl.GetLivenessesFromKV(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(462132)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462133)
		}
		__antithesis_instrumentation__.Notify(462125)

		sort.Slice(livenesses, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(462134)
			return livenesses[i].NodeID < livenesses[j].NodeID
		})
		__antithesis_instrumentation__.Notify(462126)

		for i := range livenesses {
			__antithesis_instrumentation__.Notify(462135)
			l := &livenesses[i]
			if err := addRow(
				tree.NewDInt(tree.DInt(l.NodeID)),
				tree.NewDInt(tree.DInt(l.Epoch)),
				tree.NewDString(l.Expiration.String()),
				tree.MakeDBool(tree.DBool(l.Draining)),
				tree.NewDString(l.Membership.String()),
			); err != nil {
				__antithesis_instrumentation__.Notify(462136)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462137)
			}
		}
		__antithesis_instrumentation__.Notify(462127)
		return nil
	},
}

var crdbInternalGossipLivenessTable = virtualSchemaTable{
	comment: "locally known gossiped node liveness (RAM; local node only)",
	schema: `
CREATE TABLE crdb_internal.gossip_liveness (
  node_id          INT NOT NULL,
  epoch            INT NOT NULL,
  expiration       STRING NOT NULL,
  draining         BOOL NOT NULL,
  decommissioning  BOOL NOT NULL,
  membership       STRING NOT NULL,
  updated_at       TIMESTAMP
)
	`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462138)

		if err := p.RequireAdminRole(ctx, "read crdb_internal.gossip_liveness"); err != nil {
			__antithesis_instrumentation__.Notify(462144)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462145)
		}
		__antithesis_instrumentation__.Notify(462139)

		g, err := p.ExecCfg().Gossip.OptionalErr(47899)
		if err != nil {
			__antithesis_instrumentation__.Notify(462146)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462147)
		}
		__antithesis_instrumentation__.Notify(462140)

		type nodeInfo struct {
			liveness  livenesspb.Liveness
			updatedAt int64
		}

		var nodes []nodeInfo
		if err := g.IterateInfos(gossip.KeyNodeLivenessPrefix, func(key string, i gossip.Info) error {
			__antithesis_instrumentation__.Notify(462148)
			bytes, err := i.Value.GetBytes()
			if err != nil {
				__antithesis_instrumentation__.Notify(462151)
				return errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to extract bytes for key %q", key)
			} else {
				__antithesis_instrumentation__.Notify(462152)
			}
			__antithesis_instrumentation__.Notify(462149)

			var l livenesspb.Liveness
			if err := protoutil.Unmarshal(bytes, &l); err != nil {
				__antithesis_instrumentation__.Notify(462153)
				return errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to parse value for key %q", key)
			} else {
				__antithesis_instrumentation__.Notify(462154)
			}
			__antithesis_instrumentation__.Notify(462150)
			nodes = append(nodes, nodeInfo{
				liveness:  l,
				updatedAt: i.OrigStamp,
			})
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(462155)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462156)
		}
		__antithesis_instrumentation__.Notify(462141)

		sort.Slice(nodes, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(462157)
			return nodes[i].liveness.NodeID < nodes[j].liveness.NodeID
		})
		__antithesis_instrumentation__.Notify(462142)

		for i := range nodes {
			__antithesis_instrumentation__.Notify(462158)
			n := &nodes[i]
			l := &n.liveness
			updatedTSDatum, err := tree.MakeDTimestamp(timeutil.Unix(0, n.updatedAt), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(462160)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462161)
			}
			__antithesis_instrumentation__.Notify(462159)
			if err := addRow(
				tree.NewDInt(tree.DInt(l.NodeID)),
				tree.NewDInt(tree.DInt(l.Epoch)),
				tree.NewDString(l.Expiration.String()),
				tree.MakeDBool(tree.DBool(l.Draining)),
				tree.MakeDBool(tree.DBool(!l.Membership.Active())),
				tree.NewDString(l.Membership.String()),
				updatedTSDatum,
			); err != nil {
				__antithesis_instrumentation__.Notify(462162)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462163)
			}
		}
		__antithesis_instrumentation__.Notify(462143)
		return nil
	},
}

var crdbInternalGossipAlertsTable = virtualSchemaTable{
	comment: "locally known gossiped health alerts (RAM; local node only)",
	schema: `
CREATE TABLE crdb_internal.gossip_alerts (
  node_id         INT NOT NULL,
  store_id        INT NULL,        -- null for alerts not associated to a store
  category        STRING NOT NULL, -- type of alert, usually by subsystem
  description     STRING NOT NULL, -- name of the alert (depends on subsystem)
  value           FLOAT NOT NULL   -- value of the alert (depends on subsystem, can be NaN)
)
	`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462164)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.gossip_alerts"); err != nil {
			__antithesis_instrumentation__.Notify(462169)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462170)
		}
		__antithesis_instrumentation__.Notify(462165)

		g, err := p.ExecCfg().Gossip.OptionalErr(47899)
		if err != nil {
			__antithesis_instrumentation__.Notify(462171)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462172)
		}
		__antithesis_instrumentation__.Notify(462166)

		type resultWithNodeID struct {
			roachpb.NodeID
			statuspb.HealthCheckResult
		}
		var results []resultWithNodeID
		if err := g.IterateInfos(gossip.KeyNodeHealthAlertPrefix, func(key string, i gossip.Info) error {
			__antithesis_instrumentation__.Notify(462173)
			bytes, err := i.Value.GetBytes()
			if err != nil {
				__antithesis_instrumentation__.Notify(462177)
				return errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to extract bytes for key %q", key)
			} else {
				__antithesis_instrumentation__.Notify(462178)
			}
			__antithesis_instrumentation__.Notify(462174)

			var d statuspb.HealthCheckResult
			if err := protoutil.Unmarshal(bytes, &d); err != nil {
				__antithesis_instrumentation__.Notify(462179)
				return errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to parse value for key %q", key)
			} else {
				__antithesis_instrumentation__.Notify(462180)
			}
			__antithesis_instrumentation__.Notify(462175)
			nodeID, err := gossip.NodeIDFromKey(key, gossip.KeyNodeHealthAlertPrefix)
			if err != nil {
				__antithesis_instrumentation__.Notify(462181)
				return errors.NewAssertionErrorWithWrappedErrf(err,
					"failed to parse node ID from key %q", key)
			} else {
				__antithesis_instrumentation__.Notify(462182)
			}
			__antithesis_instrumentation__.Notify(462176)
			results = append(results, resultWithNodeID{nodeID, d})
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(462183)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462184)
		}
		__antithesis_instrumentation__.Notify(462167)

		for _, result := range results {
			__antithesis_instrumentation__.Notify(462185)
			for _, alert := range result.Alerts {
				__antithesis_instrumentation__.Notify(462186)
				storeID := tree.DNull
				if alert.StoreID != 0 {
					__antithesis_instrumentation__.Notify(462188)
					storeID = tree.NewDInt(tree.DInt(alert.StoreID))
				} else {
					__antithesis_instrumentation__.Notify(462189)
				}
				__antithesis_instrumentation__.Notify(462187)
				if err := addRow(
					tree.NewDInt(tree.DInt(result.NodeID)),
					storeID,
					tree.NewDString(strings.ToLower(alert.Category.String())),
					tree.NewDString(alert.Description),
					tree.NewDFloat(tree.DFloat(alert.Value)),
				); err != nil {
					__antithesis_instrumentation__.Notify(462190)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462191)
				}
			}
		}
		__antithesis_instrumentation__.Notify(462168)
		return nil
	},
}

var crdbInternalGossipNetworkTable = virtualSchemaTable{
	comment: "locally known edges in the gossip network (RAM; local node only)",
	schema: `
CREATE TABLE crdb_internal.gossip_network (
  source_id       INT NOT NULL,    -- source node of a gossip connection
  target_id       INT NOT NULL     -- target node of a gossip connection
)
	`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462192)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.gossip_network"); err != nil {
			__antithesis_instrumentation__.Notify(462196)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462197)
		}
		__antithesis_instrumentation__.Notify(462193)

		g, err := p.ExecCfg().Gossip.OptionalErr(47899)
		if err != nil {
			__antithesis_instrumentation__.Notify(462198)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462199)
		}
		__antithesis_instrumentation__.Notify(462194)

		c := g.Connectivity()
		for _, conn := range c.ClientConns {
			__antithesis_instrumentation__.Notify(462200)
			if err := addRow(
				tree.NewDInt(tree.DInt(conn.SourceID)),
				tree.NewDInt(tree.DInt(conn.TargetID)),
			); err != nil {
				__antithesis_instrumentation__.Notify(462201)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462202)
			}
		}
		__antithesis_instrumentation__.Notify(462195)
		return nil
	},
}

func addPartitioningRows(
	ctx context.Context,
	p *planner,
	database string,
	table catalog.TableDescriptor,
	index catalog.Index,
	partitioning catalog.Partitioning,
	parentName tree.Datum,
	colOffset int,
	addRow func(...tree.Datum) error,
) error {
	__antithesis_instrumentation__.Notify(462203)
	tableID := tree.NewDInt(tree.DInt(table.GetID()))
	indexID := tree.NewDInt(tree.DInt(index.GetID()))
	numColumns := tree.NewDInt(tree.DInt(partitioning.NumColumns()))

	var buf bytes.Buffer
	for i := uint32(colOffset); i < uint32(colOffset+partitioning.NumColumns()); i++ {
		__antithesis_instrumentation__.Notify(462209)
		if i != uint32(colOffset) {
			__antithesis_instrumentation__.Notify(462211)
			buf.WriteString(`, `)
		} else {
			__antithesis_instrumentation__.Notify(462212)
		}
		__antithesis_instrumentation__.Notify(462210)
		buf.WriteString(index.GetKeyColumnName(int(i)))
	}
	__antithesis_instrumentation__.Notify(462204)
	colNames := tree.NewDString(buf.String())

	var datumAlloc tree.DatumAlloc

	fakePrefixDatums := make([]tree.Datum, colOffset)
	for i := range fakePrefixDatums {
		__antithesis_instrumentation__.Notify(462213)
		fakePrefixDatums[i] = tree.DNull
	}
	__antithesis_instrumentation__.Notify(462205)

	err := partitioning.ForEachList(func(name string, values [][]byte, subPartitioning catalog.Partitioning) error {
		__antithesis_instrumentation__.Notify(462214)
		var buf bytes.Buffer
		for j, values := range values {
			__antithesis_instrumentation__.Notify(462219)
			if j != 0 {
				__antithesis_instrumentation__.Notify(462222)
				buf.WriteString(`, `)
			} else {
				__antithesis_instrumentation__.Notify(462223)
			}
			__antithesis_instrumentation__.Notify(462220)
			tuple, _, err := rowenc.DecodePartitionTuple(
				&datumAlloc, p.ExecCfg().Codec, table, index, partitioning, values, fakePrefixDatums,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(462224)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462225)
			}
			__antithesis_instrumentation__.Notify(462221)
			buf.WriteString(tuple.String())
		}
		__antithesis_instrumentation__.Notify(462215)

		partitionValue := tree.NewDString(buf.String())
		nameDString := tree.NewDString(name)

		zoneID, zone, subzone, err := GetZoneConfigInTxn(
			ctx, p.txn, p.ExecCfg().Codec, table.GetID(), index, name, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(462226)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462227)
		}
		__antithesis_instrumentation__.Notify(462216)
		subzoneID := base.SubzoneID(0)
		if subzone != nil {
			__antithesis_instrumentation__.Notify(462228)
			for i, s := range zone.Subzones {
				__antithesis_instrumentation__.Notify(462229)
				if s.IndexID == subzone.IndexID && func() bool {
					__antithesis_instrumentation__.Notify(462230)
					return s.PartitionName == subzone.PartitionName == true
				}() == true {
					__antithesis_instrumentation__.Notify(462231)
					subzoneID = base.SubzoneIDFromIndex(i)
				} else {
					__antithesis_instrumentation__.Notify(462232)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(462233)
		}
		__antithesis_instrumentation__.Notify(462217)

		if err := addRow(
			tableID,
			indexID,
			parentName,
			nameDString,
			numColumns,
			colNames,
			partitionValue,
			tree.DNull,
			tree.NewDInt(tree.DInt(zoneID)),
			tree.NewDInt(tree.DInt(subzoneID)),
		); err != nil {
			__antithesis_instrumentation__.Notify(462234)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462235)
		}
		__antithesis_instrumentation__.Notify(462218)
		return addPartitioningRows(ctx, p, database, table, index, subPartitioning, nameDString,
			colOffset+partitioning.NumColumns(), addRow)
	})
	__antithesis_instrumentation__.Notify(462206)
	if err != nil {
		__antithesis_instrumentation__.Notify(462236)
		return err
	} else {
		__antithesis_instrumentation__.Notify(462237)
	}
	__antithesis_instrumentation__.Notify(462207)

	err = partitioning.ForEachRange(func(name string, from, to []byte) error {
		__antithesis_instrumentation__.Notify(462238)
		var buf bytes.Buffer
		fromTuple, _, err := rowenc.DecodePartitionTuple(
			&datumAlloc, p.ExecCfg().Codec, table, index, partitioning, from, fakePrefixDatums,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(462243)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462244)
		}
		__antithesis_instrumentation__.Notify(462239)
		buf.WriteString(fromTuple.String())
		buf.WriteString(" TO ")
		toTuple, _, err := rowenc.DecodePartitionTuple(
			&datumAlloc, p.ExecCfg().Codec, table, index, partitioning, to, fakePrefixDatums,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(462245)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462246)
		}
		__antithesis_instrumentation__.Notify(462240)
		buf.WriteString(toTuple.String())
		partitionRange := tree.NewDString(buf.String())

		zoneID, zone, subzone, err := GetZoneConfigInTxn(
			ctx, p.txn, p.ExecCfg().Codec, table.GetID(), index, name, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(462247)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462248)
		}
		__antithesis_instrumentation__.Notify(462241)
		subzoneID := base.SubzoneID(0)
		if subzone != nil {
			__antithesis_instrumentation__.Notify(462249)
			for i, s := range zone.Subzones {
				__antithesis_instrumentation__.Notify(462250)
				if s.IndexID == subzone.IndexID && func() bool {
					__antithesis_instrumentation__.Notify(462251)
					return s.PartitionName == subzone.PartitionName == true
				}() == true {
					__antithesis_instrumentation__.Notify(462252)
					subzoneID = base.SubzoneIDFromIndex(i)
				} else {
					__antithesis_instrumentation__.Notify(462253)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(462254)
		}
		__antithesis_instrumentation__.Notify(462242)

		return addRow(
			tableID,
			indexID,
			parentName,
			tree.NewDString(name),
			numColumns,
			colNames,
			tree.DNull,
			partitionRange,
			tree.NewDInt(tree.DInt(zoneID)),
			tree.NewDInt(tree.DInt(subzoneID)),
		)
	})
	__antithesis_instrumentation__.Notify(462208)

	return err
}

var crdbInternalPartitionsTable = virtualSchemaTable{
	comment: "defined partitions for all tables/indexes accessible by the current user in the current database (KV scan)",
	schema: `
CREATE TABLE crdb_internal.partitions (
	table_id    INT NOT NULL,
	index_id    INT NOT NULL,
	parent_name STRING,
	name        STRING NOT NULL,
	columns     INT NOT NULL,
	column_names STRING,
	list_value  STRING,
	range_value STRING,
	zone_id INT, -- references a zone id in the crdb_internal.zones table
	subzone_id INT -- references a subzone id in the crdb_internal.zones table
)
	`,
	generator: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(462255)
		dbName := ""
		if dbContext != nil {
			__antithesis_instrumentation__.Notify(462258)
			dbName = dbContext.GetName()
		} else {
			__antithesis_instrumentation__.Notify(462259)
		}
		__antithesis_instrumentation__.Notify(462256)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(462260)
			return forEachTableDescAll(ctx, p, dbContext, hideVirtual,
				func(db catalog.DatabaseDescriptor, _ string, table catalog.TableDescriptor) error {
					__antithesis_instrumentation__.Notify(462261)
					return catalog.ForEachIndex(table, catalog.IndexOpts{
						AddMutations: true,
					}, func(index catalog.Index) error {
						__antithesis_instrumentation__.Notify(462262)
						return addPartitioningRows(ctx, p, dbName, table, index, index.GetPartitioning(),
							tree.DNull, 0, pusher.pushRow)
					})
				})
		}
		__antithesis_instrumentation__.Notify(462257)
		return setupGenerator(ctx, worker, stopper)
	},
}

var crdbInternalRegionsTable = virtualSchemaTable{
	comment: "available regions for the cluster",
	schema: `
CREATE TABLE crdb_internal.regions (
	region STRING NOT NULL,
	zones STRING[] NOT NULL
)
	`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462263)
		resp, err := p.extendedEvalCtx.RegionsServer.Regions(ctx, &serverpb.RegionsRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(462266)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462267)
		}
		__antithesis_instrumentation__.Notify(462264)
		for regionName, regionMeta := range resp.Regions {
			__antithesis_instrumentation__.Notify(462268)
			zones := tree.NewDArray(types.String)
			for _, zone := range regionMeta.Zones {
				__antithesis_instrumentation__.Notify(462270)
				if err := zones.Append(tree.NewDString(zone)); err != nil {
					__antithesis_instrumentation__.Notify(462271)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462272)
				}
			}
			__antithesis_instrumentation__.Notify(462269)
			if err := addRow(
				tree.NewDString(regionName),
				zones,
			); err != nil {
				__antithesis_instrumentation__.Notify(462273)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462274)
			}
		}
		__antithesis_instrumentation__.Notify(462265)
		return nil
	},
}

var crdbInternalKVNodeStatusTable = virtualSchemaTable{
	comment: "node details across the entire cluster (cluster RPC; expensive!)",
	schema: `
CREATE TABLE crdb_internal.kv_node_status (
  node_id        INT NOT NULL,
  network        STRING NOT NULL,
  address        STRING NOT NULL,
  attrs          JSON NOT NULL,
  locality       STRING NOT NULL,
  server_version STRING NOT NULL,
  go_version     STRING NOT NULL,
  tag            STRING NOT NULL,
  time           STRING NOT NULL,
  revision       STRING NOT NULL,
  cgo_compiler   STRING NOT NULL,
  platform       STRING NOT NULL,
  distribution   STRING NOT NULL,
  type           STRING NOT NULL,
  dependencies   STRING NOT NULL,
  started_at     TIMESTAMP NOT NULL,
  updated_at     TIMESTAMP NOT NULL,
  metrics        JSON NOT NULL,
  args           JSON NOT NULL,
  env            JSON NOT NULL,
  activity       JSON NOT NULL
)
	`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462275)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.kv_node_status"); err != nil {
			__antithesis_instrumentation__.Notify(462280)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462281)
		}
		__antithesis_instrumentation__.Notify(462276)
		ss, err := p.extendedEvalCtx.NodesStatusServer.OptionalNodesStatusServer(
			errorutil.FeatureNotAvailableToNonSystemTenantsIssue)
		if err != nil {
			__antithesis_instrumentation__.Notify(462282)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462283)
		}
		__antithesis_instrumentation__.Notify(462277)
		response, err := ss.ListNodesInternal(ctx, &serverpb.NodesRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(462284)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462285)
		}
		__antithesis_instrumentation__.Notify(462278)

		for _, n := range response.Nodes {
			__antithesis_instrumentation__.Notify(462286)
			attrs := json.NewArrayBuilder(len(n.Desc.Attrs.Attrs))
			for _, a := range n.Desc.Attrs.Attrs {
				__antithesis_instrumentation__.Notify(462295)
				attrs.Add(json.FromString(a))
			}
			__antithesis_instrumentation__.Notify(462287)

			var dependencies string
			if n.BuildInfo.Dependencies == nil {
				__antithesis_instrumentation__.Notify(462296)
				dependencies = ""
			} else {
				__antithesis_instrumentation__.Notify(462297)
				dependencies = *(n.BuildInfo.Dependencies)
			}
			__antithesis_instrumentation__.Notify(462288)

			metrics := json.NewObjectBuilder(len(n.Metrics))
			for k, v := range n.Metrics {
				__antithesis_instrumentation__.Notify(462298)
				metric, err := json.FromFloat64(v)
				if err != nil {
					__antithesis_instrumentation__.Notify(462300)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462301)
				}
				__antithesis_instrumentation__.Notify(462299)
				metrics.Add(k, metric)
			}
			__antithesis_instrumentation__.Notify(462289)

			args := json.NewArrayBuilder(len(n.Args))
			for _, a := range n.Args {
				__antithesis_instrumentation__.Notify(462302)
				args.Add(json.FromString(a))
			}
			__antithesis_instrumentation__.Notify(462290)

			env := json.NewArrayBuilder(len(n.Env))
			for _, v := range n.Env {
				__antithesis_instrumentation__.Notify(462303)
				env.Add(json.FromString(v))
			}
			__antithesis_instrumentation__.Notify(462291)

			activity := json.NewObjectBuilder(len(n.Activity))
			for nodeID, values := range n.Activity {
				__antithesis_instrumentation__.Notify(462304)
				b := json.NewObjectBuilder(3)
				b.Add("incoming", json.FromInt64(values.Incoming))
				b.Add("outgoing", json.FromInt64(values.Outgoing))
				b.Add("latency", json.FromInt64(values.Latency))
				activity.Add(nodeID.String(), b.Build())
			}
			__antithesis_instrumentation__.Notify(462292)

			startTSDatum, err := tree.MakeDTimestamp(timeutil.Unix(0, n.StartedAt), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(462305)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462306)
			}
			__antithesis_instrumentation__.Notify(462293)
			endTSDatum, err := tree.MakeDTimestamp(timeutil.Unix(0, n.UpdatedAt), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(462307)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462308)
			}
			__antithesis_instrumentation__.Notify(462294)
			if err := addRow(
				tree.NewDInt(tree.DInt(n.Desc.NodeID)),
				tree.NewDString(n.Desc.Address.NetworkField),
				tree.NewDString(n.Desc.Address.AddressField),
				tree.NewDJSON(attrs.Build()),
				tree.NewDString(n.Desc.Locality.String()),
				tree.NewDString(n.Desc.ServerVersion.String()),
				tree.NewDString(n.BuildInfo.GoVersion),
				tree.NewDString(n.BuildInfo.Tag),
				tree.NewDString(n.BuildInfo.Time),
				tree.NewDString(n.BuildInfo.Revision),
				tree.NewDString(n.BuildInfo.CgoCompiler),
				tree.NewDString(n.BuildInfo.Platform),
				tree.NewDString(n.BuildInfo.Distribution),
				tree.NewDString(n.BuildInfo.Type),
				tree.NewDString(dependencies),
				startTSDatum,
				endTSDatum,
				tree.NewDJSON(metrics.Build()),
				tree.NewDJSON(args.Build()),
				tree.NewDJSON(env.Build()),
				tree.NewDJSON(activity.Build()),
			); err != nil {
				__antithesis_instrumentation__.Notify(462309)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462310)
			}
		}
		__antithesis_instrumentation__.Notify(462279)
		return nil
	},
}

var crdbInternalKVStoreStatusTable = virtualSchemaTable{
	comment: "store details and status (cluster RPC; expensive!)",
	schema: `
CREATE TABLE crdb_internal.kv_store_status (
  node_id            INT NOT NULL,
  store_id           INT NOT NULL,
  attrs              JSON NOT NULL,
  capacity           INT NOT NULL,
  available          INT NOT NULL,
  used               INT NOT NULL,
  logical_bytes      INT NOT NULL,
  range_count        INT NOT NULL,
  lease_count        INT NOT NULL,
  writes_per_second  FLOAT NOT NULL,
  bytes_per_replica  JSON NOT NULL,
  writes_per_replica JSON NOT NULL,
  metrics            JSON NOT NULL,
  properties         JSON NOT NULL
)
	`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462311)
		if err := p.RequireAdminRole(ctx, "read crdb_internal.kv_store_status"); err != nil {
			__antithesis_instrumentation__.Notify(462316)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462317)
		}
		__antithesis_instrumentation__.Notify(462312)
		ss, err := p.ExecCfg().NodesStatusServer.OptionalNodesStatusServer(
			errorutil.FeatureNotAvailableToNonSystemTenantsIssue)
		if err != nil {
			__antithesis_instrumentation__.Notify(462318)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462319)
		}
		__antithesis_instrumentation__.Notify(462313)
		response, err := ss.ListNodesInternal(ctx, &serverpb.NodesRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(462320)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462321)
		}
		__antithesis_instrumentation__.Notify(462314)

		for _, n := range response.Nodes {
			__antithesis_instrumentation__.Notify(462322)
			for _, s := range n.StoreStatuses {
				__antithesis_instrumentation__.Notify(462323)
				attrs := json.NewArrayBuilder(len(s.Desc.Attrs.Attrs))
				for _, a := range s.Desc.Attrs.Attrs {
					__antithesis_instrumentation__.Notify(462330)
					attrs.Add(json.FromString(a))
				}
				__antithesis_instrumentation__.Notify(462324)

				metrics := json.NewObjectBuilder(len(s.Metrics))
				for k, v := range s.Metrics {
					__antithesis_instrumentation__.Notify(462331)
					metric, err := json.FromFloat64(v)
					if err != nil {
						__antithesis_instrumentation__.Notify(462333)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462334)
					}
					__antithesis_instrumentation__.Notify(462332)
					metrics.Add(k, metric)
				}
				__antithesis_instrumentation__.Notify(462325)

				properties := json.NewObjectBuilder(3)
				properties.Add("read_only", json.FromBool(s.Desc.Properties.ReadOnly))
				properties.Add("encrypted", json.FromBool(s.Desc.Properties.Encrypted))
				if fsprops := s.Desc.Properties.FileStoreProperties; fsprops != nil {
					__antithesis_instrumentation__.Notify(462335)
					jprops := json.NewObjectBuilder(5)
					jprops.Add("path", json.FromString(fsprops.Path))
					jprops.Add("fs_type", json.FromString(fsprops.FsType))
					jprops.Add("mount_point", json.FromString(fsprops.MountPoint))
					jprops.Add("mount_options", json.FromString(fsprops.MountOptions))
					jprops.Add("block_device", json.FromString(fsprops.BlockDevice))
					properties.Add("file_store_properties", jprops.Build())
				} else {
					__antithesis_instrumentation__.Notify(462336)
				}
				__antithesis_instrumentation__.Notify(462326)

				percentilesToJSON := func(ps roachpb.Percentiles) (json.JSON, error) {
					__antithesis_instrumentation__.Notify(462337)
					b := json.NewObjectBuilder(5)
					v, err := json.FromFloat64(ps.P10)
					if err != nil {
						__antithesis_instrumentation__.Notify(462344)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(462345)
					}
					__antithesis_instrumentation__.Notify(462338)
					b.Add("P10", v)
					v, err = json.FromFloat64(ps.P25)
					if err != nil {
						__antithesis_instrumentation__.Notify(462346)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(462347)
					}
					__antithesis_instrumentation__.Notify(462339)
					b.Add("P25", v)
					v, err = json.FromFloat64(ps.P50)
					if err != nil {
						__antithesis_instrumentation__.Notify(462348)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(462349)
					}
					__antithesis_instrumentation__.Notify(462340)
					b.Add("P50", v)
					v, err = json.FromFloat64(ps.P75)
					if err != nil {
						__antithesis_instrumentation__.Notify(462350)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(462351)
					}
					__antithesis_instrumentation__.Notify(462341)
					b.Add("P75", v)
					v, err = json.FromFloat64(ps.P90)
					if err != nil {
						__antithesis_instrumentation__.Notify(462352)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(462353)
					}
					__antithesis_instrumentation__.Notify(462342)
					b.Add("P90", v)
					v, err = json.FromFloat64(ps.PMax)
					if err != nil {
						__antithesis_instrumentation__.Notify(462354)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(462355)
					}
					__antithesis_instrumentation__.Notify(462343)
					b.Add("PMax", v)
					return b.Build(), nil
				}
				__antithesis_instrumentation__.Notify(462327)

				bytesPerReplica, err := percentilesToJSON(s.Desc.Capacity.BytesPerReplica)
				if err != nil {
					__antithesis_instrumentation__.Notify(462356)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462357)
				}
				__antithesis_instrumentation__.Notify(462328)
				writesPerReplica, err := percentilesToJSON(s.Desc.Capacity.WritesPerReplica)
				if err != nil {
					__antithesis_instrumentation__.Notify(462358)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462359)
				}
				__antithesis_instrumentation__.Notify(462329)

				if err := addRow(
					tree.NewDInt(tree.DInt(s.Desc.Node.NodeID)),
					tree.NewDInt(tree.DInt(s.Desc.StoreID)),
					tree.NewDJSON(attrs.Build()),
					tree.NewDInt(tree.DInt(s.Desc.Capacity.Capacity)),
					tree.NewDInt(tree.DInt(s.Desc.Capacity.Available)),
					tree.NewDInt(tree.DInt(s.Desc.Capacity.Used)),
					tree.NewDInt(tree.DInt(s.Desc.Capacity.LogicalBytes)),
					tree.NewDInt(tree.DInt(s.Desc.Capacity.RangeCount)),
					tree.NewDInt(tree.DInt(s.Desc.Capacity.LeaseCount)),
					tree.NewDFloat(tree.DFloat(s.Desc.Capacity.WritesPerSecond)),
					tree.NewDJSON(bytesPerReplica),
					tree.NewDJSON(writesPerReplica),
					tree.NewDJSON(metrics.Build()),
					tree.NewDJSON(properties.Build()),
				); err != nil {
					__antithesis_instrumentation__.Notify(462360)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462361)
				}
			}
		}
		__antithesis_instrumentation__.Notify(462315)
		return nil
	},
}

var crdbInternalPredefinedCommentsTable = virtualSchemaTable{
	comment: `comments for predefined virtual tables (RAM/static)`,
	schema: `
CREATE TABLE crdb_internal.predefined_comments (
	TYPE      INT,
	OBJECT_ID INT,
	SUB_ID    INT,
	COMMENT   STRING
)`,
	populate: func(
		ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error,
	) error {
		__antithesis_instrumentation__.Notify(462362)
		tableCommentKey := tree.NewDInt(tree.DInt(keys.TableCommentType))
		vt := p.getVirtualTabler()
		vEntries := vt.getSchemas()
		vSchemaNames := vt.getSchemaNames()

		for _, virtSchemaName := range vSchemaNames {
			__antithesis_instrumentation__.Notify(462364)
			e := vEntries[virtSchemaName]

			for _, tName := range e.orderedDefNames {
				__antithesis_instrumentation__.Notify(462365)
				vTableEntry := e.defs[tName]
				table := vTableEntry.desc

				if vTableEntry.comment != "" {
					__antithesis_instrumentation__.Notify(462366)
					if err := addRow(
						tableCommentKey,
						tree.NewDInt(tree.DInt(table.GetID())),
						zeroVal,
						tree.NewDString(vTableEntry.comment)); err != nil {
						__antithesis_instrumentation__.Notify(462367)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462368)
					}
				} else {
					__antithesis_instrumentation__.Notify(462369)
				}
			}
		}
		__antithesis_instrumentation__.Notify(462363)

		return nil
	},
}

type marshaledJobMetadata struct {
	status                      *tree.DString
	payloadBytes, progressBytes *tree.DBytes
}

func (mj marshaledJobMetadata) size() (n int64) {
	__antithesis_instrumentation__.Notify(462370)
	return int64(8 + mj.status.Size() + mj.progressBytes.Size() + mj.payloadBytes.Size())
}

type marshaledJobMetadataMap map[jobspb.JobID]marshaledJobMetadata

func (m marshaledJobMetadataMap) GetJobMetadata(
	jobID jobspb.JobID,
) (md *jobs.JobMetadata, err error) {
	__antithesis_instrumentation__.Notify(462371)
	ujm, found := m[jobID]
	if !found {
		__antithesis_instrumentation__.Notify(462376)
		return nil, errors.New("job not found")
	} else {
		__antithesis_instrumentation__.Notify(462377)
	}
	__antithesis_instrumentation__.Notify(462372)
	md = &jobs.JobMetadata{ID: jobID}
	if ujm.status == nil {
		__antithesis_instrumentation__.Notify(462378)
		return nil, errors.New("missing status")
	} else {
		__antithesis_instrumentation__.Notify(462379)
	}
	__antithesis_instrumentation__.Notify(462373)
	md.Status = jobs.Status(*ujm.status)
	md.Payload, err = jobs.UnmarshalPayload(ujm.payloadBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(462380)
		return nil, errors.Wrap(err, "corrupt payload bytes")
	} else {
		__antithesis_instrumentation__.Notify(462381)
	}
	__antithesis_instrumentation__.Notify(462374)
	md.Progress, err = jobs.UnmarshalProgress(ujm.progressBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(462382)
		return nil, errors.Wrap(err, "corrupt progress bytes")
	} else {
		__antithesis_instrumentation__.Notify(462383)
	}
	__antithesis_instrumentation__.Notify(462375)
	return md, nil
}

func getPayloadAndProgressFromJobsRecord(
	p *planner, job *jobs.Record,
) (progressBytes *tree.DBytes, payloadBytes *tree.DBytes, err error) {
	__antithesis_instrumentation__.Notify(462384)
	progressMarshalled, err := protoutil.Marshal(&jobspb.Progress{
		ModifiedMicros: p.txn.ReadTimestamp().GoTime().UnixMicro(),
		Details:        jobspb.WrapProgressDetails(job.Progress),
		RunningStatus:  string(job.RunningStatus),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(462387)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(462388)
	}
	__antithesis_instrumentation__.Notify(462385)
	progressBytes = tree.NewDBytes(tree.DBytes(progressMarshalled))
	payloadMarshalled, err := protoutil.Marshal(&jobspb.Payload{
		Description:   job.Description,
		Statement:     job.Statements,
		UsernameProto: job.Username.EncodeProto(),
		Details:       jobspb.WrapPayloadDetails(job.Details),
		Noncancelable: job.NonCancelable,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(462389)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(462390)
	}
	__antithesis_instrumentation__.Notify(462386)
	payloadBytes = tree.NewDBytes(tree.DBytes(payloadMarshalled))
	return progressBytes, payloadBytes, nil
}

func collectMarshaledJobMetadataMap(
	ctx context.Context, p *planner, acct *mon.BoundAccount, descs []catalog.Descriptor,
) (marshaledJobMetadataMap, error) {
	__antithesis_instrumentation__.Notify(462391)

	referencedJobIDs := map[jobspb.JobID]struct{}{}
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(462398)
		tbl, ok := desc.(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(462400)
			continue
		} else {
			__antithesis_instrumentation__.Notify(462401)
		}
		__antithesis_instrumentation__.Notify(462399)
		for _, j := range tbl.GetMutationJobs() {
			__antithesis_instrumentation__.Notify(462402)
			referencedJobIDs[j.JobID] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(462392)
	if len(referencedJobIDs) == 0 {
		__antithesis_instrumentation__.Notify(462403)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(462404)
	}
	__antithesis_instrumentation__.Notify(462393)

	m := make(marshaledJobMetadataMap)
	query := `SELECT id, status, payload, progress FROM system.jobs`
	it, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryIteratorEx(
		ctx, "crdb-internal-jobs-table", p.Txn(),
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		query)
	if err != nil {
		__antithesis_instrumentation__.Notify(462405)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462406)
	}
	__antithesis_instrumentation__.Notify(462394)
	for {
		__antithesis_instrumentation__.Notify(462407)
		ok, err := it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(462411)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(462412)
		}
		__antithesis_instrumentation__.Notify(462408)
		if !ok {
			__antithesis_instrumentation__.Notify(462413)
			break
		} else {
			__antithesis_instrumentation__.Notify(462414)
		}
		__antithesis_instrumentation__.Notify(462409)
		r := it.Cur()
		id, status, payloadBytes, progressBytes := r[0], r[1], r[2], r[3]
		jobID := jobspb.JobID(*id.(*tree.DInt))
		if _, isReferencedByDesc := referencedJobIDs[jobID]; !isReferencedByDesc {
			__antithesis_instrumentation__.Notify(462415)
			continue
		} else {
			__antithesis_instrumentation__.Notify(462416)
		}
		__antithesis_instrumentation__.Notify(462410)
		mj := marshaledJobMetadata{
			status:        status.(*tree.DString),
			payloadBytes:  payloadBytes.(*tree.DBytes),
			progressBytes: progressBytes.(*tree.DBytes),
		}
		m[jobID] = mj
		if err := acct.Grow(ctx, mj.size()); err != nil {
			__antithesis_instrumentation__.Notify(462417)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(462418)
		}
	}
	__antithesis_instrumentation__.Notify(462395)
	if err := it.Close(); err != nil {
		__antithesis_instrumentation__.Notify(462419)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462420)
	}
	__antithesis_instrumentation__.Notify(462396)
	for _, record := range p.ExtendedEvalContext().SchemaChangeJobRecords {
		__antithesis_instrumentation__.Notify(462421)
		progressBytes, payloadBytes, err := getPayloadAndProgressFromJobsRecord(p, record)
		if err != nil {
			__antithesis_instrumentation__.Notify(462423)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(462424)
		}
		__antithesis_instrumentation__.Notify(462422)
		mj := marshaledJobMetadata{
			status:        tree.NewDString(string(record.RunningStatus)),
			payloadBytes:  payloadBytes,
			progressBytes: progressBytes,
		}
		m[record.JobID] = mj
	}
	__antithesis_instrumentation__.Notify(462397)
	return m, nil
}

var crdbInternalInvalidDescriptorsTable = virtualSchemaTable{
	comment: `virtual table to validate descriptors`,
	schema: `
CREATE TABLE crdb_internal.invalid_objects (
  id            INT,
  database_name STRING,
  schema_name   STRING,
  obj_name      STRING,
  error         STRING
)`,
	populate: func(
		ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error,
	) error {
		__antithesis_instrumentation__.Notify(462425)

		c, err := p.Descriptors().Direct().GetCatalogUnvalidated(ctx, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(462431)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462432)
		}
		__antithesis_instrumentation__.Notify(462426)
		descs := c.OrderedDescriptors()

		acct := p.EvalContext().Mon.MakeBoundAccount()
		defer acct.Close(ctx)
		jmg, err := collectMarshaledJobMetadataMap(ctx, p, &acct, descs)
		if err != nil {
			__antithesis_instrumentation__.Notify(462433)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462434)
		}
		__antithesis_instrumentation__.Notify(462427)
		version := p.ExecCfg().Settings.Version.ActiveVersion(ctx)

		addValidationErrorRow := func(scName string, ne catalog.NameEntry, validationError error, lCtx tableLookupFn) error {
			__antithesis_instrumentation__.Notify(462435)
			if validationError == nil {
				__antithesis_instrumentation__.Notify(462441)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(462442)
			}
			__antithesis_instrumentation__.Notify(462436)
			dbName := fmt.Sprintf("[%d]", ne.GetParentID())
			if scName == "" {
				__antithesis_instrumentation__.Notify(462443)
				scName = fmt.Sprintf("[%d]", ne.GetParentSchemaID())
			} else {
				__antithesis_instrumentation__.Notify(462444)
			}
			__antithesis_instrumentation__.Notify(462437)
			if n, ok := lCtx.dbNames[ne.GetParentID()]; ok {
				__antithesis_instrumentation__.Notify(462445)
				dbName = n
			} else {
				__antithesis_instrumentation__.Notify(462446)
			}
			__antithesis_instrumentation__.Notify(462438)
			if n, err := lCtx.getSchemaNameByID(ne.GetParentSchemaID()); err == nil {
				__antithesis_instrumentation__.Notify(462447)
				scName = n
			} else {
				__antithesis_instrumentation__.Notify(462448)
			}
			__antithesis_instrumentation__.Notify(462439)
			objName := ne.GetName()
			if ne.GetParentSchemaID() == descpb.InvalidID {
				__antithesis_instrumentation__.Notify(462449)
				scName = objName
				objName = ""
				if ne.GetParentID() == descpb.InvalidID {
					__antithesis_instrumentation__.Notify(462450)
					dbName = scName
					scName = ""
				} else {
					__antithesis_instrumentation__.Notify(462451)
				}
			} else {
				__antithesis_instrumentation__.Notify(462452)
			}
			__antithesis_instrumentation__.Notify(462440)
			return addRow(
				tree.NewDInt(tree.DInt(ne.GetID())),
				tree.NewDString(dbName),
				tree.NewDString(scName),
				tree.NewDString(objName),
				tree.NewDString(validationError.Error()),
			)
		}
		__antithesis_instrumentation__.Notify(462428)

		doDescriptorValidationErrors := func(schema string, descriptor catalog.Descriptor, lCtx tableLookupFn) (err error) {
			__antithesis_instrumentation__.Notify(462453)
			if descriptor == nil {
				__antithesis_instrumentation__.Notify(462457)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(462458)
			}
			__antithesis_instrumentation__.Notify(462454)
			doError := func(validationError error) {
				__antithesis_instrumentation__.Notify(462459)
				if err != nil {
					__antithesis_instrumentation__.Notify(462461)
					return
				} else {
					__antithesis_instrumentation__.Notify(462462)
				}
				__antithesis_instrumentation__.Notify(462460)
				err = addValidationErrorRow(schema, descriptor, validationError, lCtx)
			}
			__antithesis_instrumentation__.Notify(462455)
			ve := c.ValidateWithRecover(ctx, version, descriptor)
			for _, validationError := range ve {
				__antithesis_instrumentation__.Notify(462463)
				doError(validationError)
			}
			__antithesis_instrumentation__.Notify(462456)
			jobs.ValidateJobReferencesInDescriptor(descriptor, jmg, doError)
			return err
		}
		__antithesis_instrumentation__.Notify(462429)

		const allowAdding = true
		if err := forEachTableDescWithTableLookupInternalFromDescriptors(
			ctx, p, dbContext, hideVirtual, allowAdding, c, func(
				_ catalog.DatabaseDescriptor, schema string, descriptor catalog.TableDescriptor, lCtx tableLookupFn,
			) error {
				__antithesis_instrumentation__.Notify(462464)
				return doDescriptorValidationErrors(schema, descriptor, lCtx)
			}); err != nil {
			__antithesis_instrumentation__.Notify(462465)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462466)
		}
		__antithesis_instrumentation__.Notify(462430)

		if err := forEachTypeDescWithTableLookupInternalFromDescriptors(
			ctx, p, dbContext, allowAdding, c, func(
				_ catalog.DatabaseDescriptor, schema string, descriptor catalog.TypeDescriptor, lCtx tableLookupFn,
			) error {
				__antithesis_instrumentation__.Notify(462467)
				return doDescriptorValidationErrors(schema, descriptor, lCtx)
			}); err != nil {
			__antithesis_instrumentation__.Notify(462468)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462469)
		}

		{
			__antithesis_instrumentation__.Notify(462470)
			lCtx := newInternalLookupCtx(c.OrderedDescriptors(), dbContext)

			if err := c.ForEachDescriptorEntry(func(desc catalog.Descriptor) error {
				__antithesis_instrumentation__.Notify(462472)
				switch d := desc.(type) {
				case catalog.DatabaseDescriptor:
					__antithesis_instrumentation__.Notify(462474)
					if dbContext != nil && func() bool {
						__antithesis_instrumentation__.Notify(462477)
						return d.GetID() != dbContext.GetID() == true
					}() == true {
						__antithesis_instrumentation__.Notify(462478)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(462479)
					}
				case catalog.SchemaDescriptor:
					__antithesis_instrumentation__.Notify(462475)
					if dbContext != nil && func() bool {
						__antithesis_instrumentation__.Notify(462480)
						return d.GetParentID() != dbContext.GetID() == true
					}() == true {
						__antithesis_instrumentation__.Notify(462481)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(462482)
					}
				default:
					__antithesis_instrumentation__.Notify(462476)
					return nil
				}
				__antithesis_instrumentation__.Notify(462473)
				return doDescriptorValidationErrors("", desc, lCtx)
			}); err != nil {
				__antithesis_instrumentation__.Notify(462483)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462484)
			}
			__antithesis_instrumentation__.Notify(462471)

			return c.ForEachNamespaceEntry(func(ne catalog.NameEntry) error {
				__antithesis_instrumentation__.Notify(462485)
				if dbContext != nil {
					__antithesis_instrumentation__.Notify(462487)
					if ne.GetParentID() == descpb.InvalidID {
						__antithesis_instrumentation__.Notify(462488)
						if ne.GetID() != dbContext.GetID() {
							__antithesis_instrumentation__.Notify(462489)
							return nil
						} else {
							__antithesis_instrumentation__.Notify(462490)
						}
					} else {
						__antithesis_instrumentation__.Notify(462491)
						if ne.GetParentID() != dbContext.GetID() {
							__antithesis_instrumentation__.Notify(462492)
							return nil
						} else {
							__antithesis_instrumentation__.Notify(462493)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(462494)
				}
				__antithesis_instrumentation__.Notify(462486)
				return addValidationErrorRow("", ne, c.ValidateNamespaceEntry(ne), lCtx)
			})
		}
	},
}

var crdbInternalClusterDatabasePrivilegesTable = virtualSchemaTable{
	comment: `virtual table with database privileges`,
	schema: `
CREATE TABLE crdb_internal.cluster_database_privileges (
	database_name   STRING NOT NULL,
	grantee         STRING NOT NULL,
	privilege_type  STRING NOT NULL,
	is_grantable 		STRING
)`,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462495)
		return forEachDatabaseDesc(ctx, p, dbContext, true,
			func(db catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(462496)
				privs := db.GetPrivileges().Show(privilege.Database)
				dbNameStr := tree.NewDString(db.GetName())

				populateGrantOption := p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ValidateGrantOption)
				for _, u := range privs {
					__antithesis_instrumentation__.Notify(462498)
					userNameStr := tree.NewDString(u.User.Normalized())
					for _, priv := range u.Privileges {
						__antithesis_instrumentation__.Notify(462499)
						var isGrantable tree.Datum
						if populateGrantOption {
							__antithesis_instrumentation__.Notify(462501)
							isGrantable = yesOrNoDatum(priv.GrantOption)
						} else {
							__antithesis_instrumentation__.Notify(462502)
							isGrantable = tree.DNull
						}
						__antithesis_instrumentation__.Notify(462500)
						if err := addRow(
							dbNameStr,
							userNameStr,
							tree.NewDString(priv.Kind.String()),
							isGrantable,
						); err != nil {
							__antithesis_instrumentation__.Notify(462503)
							return err
						} else {
							__antithesis_instrumentation__.Notify(462504)
						}
					}
				}
				__antithesis_instrumentation__.Notify(462497)
				return nil
			})
	},
}

var crdbInternalCrossDbReferences = virtualSchemaTable{
	comment: `virtual table with cross db references`,
	schema: `
CREATE TABLE crdb_internal.cross_db_references (
	object_database
		STRING NOT NULL,
	object_schema
		STRING NOT NULL,
	object_name
		STRING NOT NULL,
	referenced_object_database
		STRING NOT NULL,
	referenced_object_schema
		STRING NOT NULL,
	referenced_object_name
		STRING NOT NULL,
	cross_database_reference_description
		STRING NOT NULL
);`,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462505)
		return forEachTableDescAllWithTableLookup(ctx, p, dbContext, hideVirtual,
			func(db catalog.DatabaseDescriptor, schemaName string, table catalog.TableDescriptor, lookupFn tableLookupFn) error {
				__antithesis_instrumentation__.Notify(462506)

				if table.IsTable() {
					__antithesis_instrumentation__.Notify(462508)
					objectDatabaseName := lookupFn.getDatabaseName(table)
					err := table.ForeachOutboundFK(
						func(fk *descpb.ForeignKeyConstraint) error {
							__antithesis_instrumentation__.Notify(462511)
							referencedTable, err := lookupFn.getTableByID(fk.ReferencedTableID)
							if err != nil {
								__antithesis_instrumentation__.Notify(462514)
								return err
							} else {
								__antithesis_instrumentation__.Notify(462515)
							}
							__antithesis_instrumentation__.Notify(462512)
							if referencedTable.GetParentID() != table.GetParentID() {
								__antithesis_instrumentation__.Notify(462516)
								refSchemaName, err := lookupFn.getSchemaNameByID(referencedTable.GetParentSchemaID())
								if err != nil {
									__antithesis_instrumentation__.Notify(462518)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462519)
								}
								__antithesis_instrumentation__.Notify(462517)
								refDatabaseName := lookupFn.getDatabaseName(referencedTable)

								if err := addRow(tree.NewDString(objectDatabaseName),
									tree.NewDString(schemaName),
									tree.NewDString(table.GetName()),
									tree.NewDString(refDatabaseName),
									tree.NewDString(refSchemaName),
									tree.NewDString(referencedTable.GetName()),
									tree.NewDString("table foreign key reference")); err != nil {
									__antithesis_instrumentation__.Notify(462520)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462521)
								}
							} else {
								__antithesis_instrumentation__.Notify(462522)
							}
							__antithesis_instrumentation__.Notify(462513)
							return nil
						})
					__antithesis_instrumentation__.Notify(462509)
					if err != nil {
						__antithesis_instrumentation__.Notify(462523)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462524)
					}
					__antithesis_instrumentation__.Notify(462510)

					for _, col := range table.PublicColumns() {
						__antithesis_instrumentation__.Notify(462525)
						for i := 0; i < col.NumUsesSequences(); i++ {
							__antithesis_instrumentation__.Notify(462526)
							sequenceID := col.GetUsesSequenceID(i)
							seqDesc, err := lookupFn.getTableByID(sequenceID)
							if err != nil {
								__antithesis_instrumentation__.Notify(462528)
								return err
							} else {
								__antithesis_instrumentation__.Notify(462529)
							}
							__antithesis_instrumentation__.Notify(462527)
							if seqDesc.GetParentID() != table.GetParentID() {
								__antithesis_instrumentation__.Notify(462530)
								seqSchemaName, err := lookupFn.getSchemaNameByID(seqDesc.GetParentSchemaID())
								if err != nil {
									__antithesis_instrumentation__.Notify(462532)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462533)
								}
								__antithesis_instrumentation__.Notify(462531)
								refDatabaseName := lookupFn.getDatabaseName(seqDesc)
								if err := addRow(tree.NewDString(objectDatabaseName),
									tree.NewDString(schemaName),
									tree.NewDString(table.GetName()),
									tree.NewDString(refDatabaseName),
									tree.NewDString(seqSchemaName),
									tree.NewDString(seqDesc.GetName()),
									tree.NewDString("table column refers to sequence")); err != nil {
									__antithesis_instrumentation__.Notify(462534)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462535)
								}
							} else {
								__antithesis_instrumentation__.Notify(462536)
							}
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(462537)
					if table.IsView() {
						__antithesis_instrumentation__.Notify(462538)

						dependsOn := table.GetDependsOn()
						for _, dependency := range dependsOn {
							__antithesis_instrumentation__.Notify(462540)
							dependentTable, err := lookupFn.getTableByID(dependency)
							if err != nil {
								__antithesis_instrumentation__.Notify(462542)
								return err
							} else {
								__antithesis_instrumentation__.Notify(462543)
							}
							__antithesis_instrumentation__.Notify(462541)
							if dependentTable.GetParentID() != table.GetParentID() {
								__antithesis_instrumentation__.Notify(462544)
								objectDatabaseName := lookupFn.getDatabaseName(table)
								refSchemaName, err := lookupFn.getSchemaNameByID(dependentTable.GetParentSchemaID())
								if err != nil {
									__antithesis_instrumentation__.Notify(462546)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462547)
								}
								__antithesis_instrumentation__.Notify(462545)
								refDatabaseName := lookupFn.getDatabaseName(dependentTable)

								if err := addRow(tree.NewDString(objectDatabaseName),
									tree.NewDString(schemaName),
									tree.NewDString(table.GetName()),
									tree.NewDString(refDatabaseName),
									tree.NewDString(refSchemaName),
									tree.NewDString(dependentTable.GetName()),
									tree.NewDString("view references table")); err != nil {
									__antithesis_instrumentation__.Notify(462548)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462549)
								}
							} else {
								__antithesis_instrumentation__.Notify(462550)
							}
						}
						__antithesis_instrumentation__.Notify(462539)

						dependsOnTypes := table.GetDependsOnTypes()
						for _, dependency := range dependsOnTypes {
							__antithesis_instrumentation__.Notify(462551)
							dependentType, err := lookupFn.getTypeByID(dependency)
							if err != nil {
								__antithesis_instrumentation__.Notify(462553)
								return err
							} else {
								__antithesis_instrumentation__.Notify(462554)
							}
							__antithesis_instrumentation__.Notify(462552)
							if dependentType.GetParentID() != table.GetParentID() {
								__antithesis_instrumentation__.Notify(462555)
								objectDatabaseName := lookupFn.getDatabaseName(table)
								refSchemaName, err := lookupFn.getSchemaNameByID(dependentType.GetParentSchemaID())
								if err != nil {
									__antithesis_instrumentation__.Notify(462557)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462558)
								}
								__antithesis_instrumentation__.Notify(462556)
								refDatabaseName := lookupFn.getDatabaseName(dependentType)

								if err := addRow(tree.NewDString(objectDatabaseName),
									tree.NewDString(schemaName),
									tree.NewDString(table.GetName()),
									tree.NewDString(refDatabaseName),
									tree.NewDString(refSchemaName),
									tree.NewDString(dependentType.GetName()),
									tree.NewDString("view references type")); err != nil {
									__antithesis_instrumentation__.Notify(462559)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462560)
								}
							} else {
								__antithesis_instrumentation__.Notify(462561)
							}
						}

					} else {
						__antithesis_instrumentation__.Notify(462562)
						if table.IsSequence() {
							__antithesis_instrumentation__.Notify(462563)

							sequenceOpts := table.GetSequenceOpts()
							if sequenceOpts.SequenceOwner.OwnerTableID != descpb.InvalidID {
								__antithesis_instrumentation__.Notify(462564)
								ownerTable, err := lookupFn.getTableByID(sequenceOpts.SequenceOwner.OwnerTableID)
								if err != nil {
									__antithesis_instrumentation__.Notify(462566)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462567)
								}
								__antithesis_instrumentation__.Notify(462565)
								if ownerTable.GetParentID() != table.GetParentID() {
									__antithesis_instrumentation__.Notify(462568)
									objectDatabaseName := lookupFn.getDatabaseName(table)
									refSchemaName, err := lookupFn.getSchemaNameByID(ownerTable.GetParentSchemaID())
									if err != nil {
										__antithesis_instrumentation__.Notify(462570)
										return err
									} else {
										__antithesis_instrumentation__.Notify(462571)
									}
									__antithesis_instrumentation__.Notify(462569)
									refDatabaseName := lookupFn.getDatabaseName(ownerTable)

									if err := addRow(tree.NewDString(objectDatabaseName),
										tree.NewDString(schemaName),
										tree.NewDString(table.GetName()),
										tree.NewDString(refDatabaseName),
										tree.NewDString(refSchemaName),
										tree.NewDString(ownerTable.GetName()),
										tree.NewDString("sequences owning table")); err != nil {
										__antithesis_instrumentation__.Notify(462572)
										return err
									} else {
										__antithesis_instrumentation__.Notify(462573)
									}
								} else {
									__antithesis_instrumentation__.Notify(462574)
								}
							} else {
								__antithesis_instrumentation__.Notify(462575)
							}
						} else {
							__antithesis_instrumentation__.Notify(462576)
						}
					}
				}
				__antithesis_instrumentation__.Notify(462507)
				return nil
			})
	},
}

var crdbLostTableDescriptors = virtualSchemaTable{
	comment: `virtual table with table descriptors that still have data`,
	schema: `
CREATE TABLE crdb_internal.lost_descriptors_with_data (
	descID
		INTEGER NOT NULL
);`,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462577)
		maxDescIDKeyVal, err := p.extendedEvalCtx.DB.Get(context.Background(), p.extendedEvalCtx.Codec.DescIDSequenceKey())
		if err != nil {
			__antithesis_instrumentation__.Notify(462584)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462585)
		}
		__antithesis_instrumentation__.Notify(462578)
		maxDescID, err := maxDescIDKeyVal.Value.GetInt()
		if err != nil {
			__antithesis_instrumentation__.Notify(462586)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462587)
		}
		__antithesis_instrumentation__.Notify(462579)

		c, err := p.Descriptors().Direct().GetCatalogUnvalidated(ctx, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(462588)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462589)
		}
		__antithesis_instrumentation__.Notify(462580)

		unusedDescSpan := roachpb.Span{}
		descStart := 0
		descEnd := 0
		scanAndGenerateRows := func() error {
			__antithesis_instrumentation__.Notify(462590)
			if unusedDescSpan.Key == nil {
				__antithesis_instrumentation__.Notify(462594)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(462595)
			}
			__antithesis_instrumentation__.Notify(462591)
			b := kv.Batch{}
			b.Header.MaxSpanRequestKeys = 1
			scanRequest := roachpb.NewScan(unusedDescSpan.Key, unusedDescSpan.EndKey, false).(*roachpb.ScanRequest)
			scanRequest.ScanFormat = roachpb.BATCH_RESPONSE
			b.AddRawRequest(scanRequest)
			err = p.extendedEvalCtx.DB.Run(ctx, &b)
			if err != nil {
				__antithesis_instrumentation__.Notify(462596)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462597)
			}
			__antithesis_instrumentation__.Notify(462592)

			res := b.RawResponse().Responses[0].GetScan()
			if res.NumKeys > 0 {
				__antithesis_instrumentation__.Notify(462598)
				b = kv.Batch{}
				b.Header.MaxSpanRequestKeys = 1
				for descID := descStart; descID <= descEnd; descID++ {
					__antithesis_instrumentation__.Notify(462601)
					prefix := p.extendedEvalCtx.Codec.TablePrefix(uint32(descID))
					scanRequest := roachpb.NewScan(prefix, prefix.PrefixEnd(), false).(*roachpb.ScanRequest)
					scanRequest.ScanFormat = roachpb.BATCH_RESPONSE
					b.AddRawRequest(scanRequest)
				}
				__antithesis_instrumentation__.Notify(462599)
				err = p.extendedEvalCtx.DB.Run(ctx, &b)
				if err != nil {
					__antithesis_instrumentation__.Notify(462602)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462603)
				}
				__antithesis_instrumentation__.Notify(462600)
				for idx := range b.RawResponse().Responses {
					__antithesis_instrumentation__.Notify(462604)
					res := b.RawResponse().Responses[idx].GetScan()
					if res.NumKeys == 0 {
						__antithesis_instrumentation__.Notify(462606)
						continue
					} else {
						__antithesis_instrumentation__.Notify(462607)
					}
					__antithesis_instrumentation__.Notify(462605)

					if err := addRow(tree.NewDInt(tree.DInt(idx + descStart))); err != nil {
						__antithesis_instrumentation__.Notify(462608)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462609)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(462610)
			}
			__antithesis_instrumentation__.Notify(462593)
			unusedDescSpan = roachpb.Span{}
			descStart = 0
			descEnd = 0
			return nil
		}
		__antithesis_instrumentation__.Notify(462581)

		for id := uint32(keys.MaxReservedDescID + 1); id < uint32(maxDescID); id++ {
			__antithesis_instrumentation__.Notify(462611)

			if c.LookupDescriptorEntry(descpb.ID(id)) != nil {
				__antithesis_instrumentation__.Notify(462614)
				err := scanAndGenerateRows()
				if err != nil {
					__antithesis_instrumentation__.Notify(462616)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462617)
				}
				__antithesis_instrumentation__.Notify(462615)
				continue
			} else {
				__antithesis_instrumentation__.Notify(462618)
			}
			__antithesis_instrumentation__.Notify(462612)

			prefix := p.extendedEvalCtx.Codec.TablePrefix(id)
			if unusedDescSpan.Key == nil {
				__antithesis_instrumentation__.Notify(462619)
				descStart = int(id)
				unusedDescSpan.Key = prefix
			} else {
				__antithesis_instrumentation__.Notify(462620)
			}
			__antithesis_instrumentation__.Notify(462613)
			descEnd = int(id)
			unusedDescSpan.EndKey = prefix.PrefixEnd()

		}
		__antithesis_instrumentation__.Notify(462582)
		err = scanAndGenerateRows()
		if err != nil {
			__antithesis_instrumentation__.Notify(462621)
			return err
		} else {
			__antithesis_instrumentation__.Notify(462622)
		}
		__antithesis_instrumentation__.Notify(462583)
		return nil
	},
}

var crdbInternalDefaultPrivilegesTable = virtualSchemaTable{
	comment: `virtual table with default privileges`,
	schema: `
CREATE TABLE crdb_internal.default_privileges (
	database_name   STRING NOT NULL,
	schema_name     STRING,
	role            STRING,
	for_all_roles   BOOL,
	object_type     STRING NOT NULL,
	grantee         STRING NOT NULL,
	privilege_type  STRING NOT NULL
);`,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462623)
		return forEachDatabaseDesc(ctx, p, nil, true,
			func(descriptor catalog.DatabaseDescriptor) error {
				__antithesis_instrumentation__.Notify(462624)
				addRowsHelper := func(defaultPrivilegesForRole catpb.DefaultPrivilegesForRole, schema tree.Datum) error {
					__antithesis_instrumentation__.Notify(462629)
					role := tree.DNull
					forAllRoles := tree.DBoolTrue
					if defaultPrivilegesForRole.IsExplicitRole() {
						__antithesis_instrumentation__.Notify(462633)
						role = tree.NewDString(defaultPrivilegesForRole.GetExplicitRole().UserProto.Decode().Normalized())
						forAllRoles = tree.DBoolFalse
					} else {
						__antithesis_instrumentation__.Notify(462634)
					}
					__antithesis_instrumentation__.Notify(462630)

					for objectType, privs := range defaultPrivilegesForRole.DefaultPrivilegesPerObject {
						__antithesis_instrumentation__.Notify(462635)
						privilegeObjectType := targetObjectToPrivilegeObject[objectType]
						for _, userPrivs := range privs.Users {
							__antithesis_instrumentation__.Notify(462636)
							privList := privilege.ListFromBitField(userPrivs.Privileges, privilegeObjectType)
							for _, priv := range privList {
								__antithesis_instrumentation__.Notify(462637)
								if err := addRow(
									tree.NewDString(descriptor.GetName()),

									schema,
									role,
									forAllRoles,
									tree.NewDString(objectType.String()),
									tree.NewDString(userPrivs.User().Normalized()),
									tree.NewDString(priv.String()),
								); err != nil {
									__antithesis_instrumentation__.Notify(462638)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462639)
								}
							}
						}
					}
					__antithesis_instrumentation__.Notify(462631)

					if schema == tree.DNull {
						__antithesis_instrumentation__.Notify(462640)
						for _, objectType := range tree.GetAlterDefaultPrivilegesTargetObjects() {
							__antithesis_instrumentation__.Notify(462642)
							if catprivilege.GetRoleHasAllPrivilegesOnTargetObject(&defaultPrivilegesForRole, objectType) {
								__antithesis_instrumentation__.Notify(462643)
								if err := addRow(
									tree.NewDString(descriptor.GetName()),
									tree.DNull,
									role,
									forAllRoles,
									tree.NewDString(objectType.String()),
									tree.NewDString(defaultPrivilegesForRole.GetExplicitRole().UserProto.Decode().Normalized()),
									tree.NewDString(privilege.ALL.String()),
								); err != nil {
									__antithesis_instrumentation__.Notify(462644)
									return err
								} else {
									__antithesis_instrumentation__.Notify(462645)
								}
							} else {
								__antithesis_instrumentation__.Notify(462646)
							}
						}
						__antithesis_instrumentation__.Notify(462641)
						if catprivilege.GetPublicHasUsageOnTypes(&defaultPrivilegesForRole) {
							__antithesis_instrumentation__.Notify(462647)
							if err := addRow(
								tree.NewDString(descriptor.GetName()),
								tree.DNull,
								role,
								forAllRoles,
								tree.NewDString(tree.Types.String()),
								tree.NewDString(security.PublicRoleName().Normalized()),
								tree.NewDString(privilege.USAGE.String()),
							); err != nil {
								__antithesis_instrumentation__.Notify(462648)
								return err
							} else {
								__antithesis_instrumentation__.Notify(462649)
							}
						} else {
							__antithesis_instrumentation__.Notify(462650)
						}
					} else {
						__antithesis_instrumentation__.Notify(462651)
					}
					__antithesis_instrumentation__.Notify(462632)
					return nil
				}
				__antithesis_instrumentation__.Notify(462625)

				addRowsForRole := func(role catpb.DefaultPrivilegesRole, defaultPrivilegeDescriptor catalog.DefaultPrivilegeDescriptor, schema tree.Datum) error {
					__antithesis_instrumentation__.Notify(462652)
					defaultPrivilegesForRole, found := defaultPrivilegeDescriptor.GetDefaultPrivilegesForRole(role)
					if !found {
						__antithesis_instrumentation__.Notify(462655)

						newDefaultPrivilegesForRole := catpb.InitDefaultPrivilegesForRole(role, defaultPrivilegeDescriptor.GetDefaultPrivilegeDescriptorType())
						defaultPrivilegesForRole = &newDefaultPrivilegesForRole
					} else {
						__antithesis_instrumentation__.Notify(462656)
					}
					__antithesis_instrumentation__.Notify(462653)
					if err := addRowsHelper(*defaultPrivilegesForRole, schema); err != nil {
						__antithesis_instrumentation__.Notify(462657)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462658)
					}
					__antithesis_instrumentation__.Notify(462654)
					return nil
				}
				__antithesis_instrumentation__.Notify(462626)

				addRowsForSchema := func(defaultPrivilegeDescriptor catalog.DefaultPrivilegeDescriptor, schema tree.Datum) error {
					__antithesis_instrumentation__.Notify(462659)
					if err := forEachRole(ctx, p, func(username security.SQLUsername, isRole bool, options roleOptions, settings tree.Datum) error {
						__antithesis_instrumentation__.Notify(462661)
						role := catpb.DefaultPrivilegesRole{
							Role: username,
						}
						return addRowsForRole(role, defaultPrivilegeDescriptor, schema)
					}); err != nil {
						__antithesis_instrumentation__.Notify(462662)
						return err
					} else {
						__antithesis_instrumentation__.Notify(462663)
					}
					__antithesis_instrumentation__.Notify(462660)

					role := catpb.DefaultPrivilegesRole{
						ForAllRoles: true,
					}
					return addRowsForRole(role, defaultPrivilegeDescriptor, schema)
				}
				__antithesis_instrumentation__.Notify(462627)

				if err := addRowsForSchema(descriptor.GetDefaultPrivilegeDescriptor(), tree.DNull); err != nil {
					__antithesis_instrumentation__.Notify(462664)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462665)
				}
				__antithesis_instrumentation__.Notify(462628)

				return forEachSchema(ctx, p, descriptor, func(schema catalog.SchemaDescriptor) error {
					__antithesis_instrumentation__.Notify(462666)
					return addRowsForSchema(schema.GetDefaultPrivilegeDescriptor(), tree.NewDString(schema.GetName()))
				})
			})
	},
}

var crdbInternalIndexUsageStatistics = virtualSchemaTable{
	comment: `cluster-wide index usage statistics (in-memory, not durable).` +
		`Querying this table is an expensive operation since it creates a` +
		`cluster-wide RPC fanout.`,
	schema: `
CREATE TABLE crdb_internal.index_usage_statistics (
  table_id        INT NOT NULL,
  index_id        INT NOT NULL,
  total_reads     INT NOT NULL,
  last_read       TIMESTAMPTZ
);`,
	generator: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(462667)

		stats, err :=
			p.extendedEvalCtx.SQLStatusServer.IndexUsageStatistics(ctx, &serverpb.IndexUsageStatisticsRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(462670)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462671)
		}
		__antithesis_instrumentation__.Notify(462668)
		indexStats := idxusage.NewLocalIndexUsageStatsFromExistingStats(&idxusage.Config{}, stats.Statistics)

		row := make(tree.Datums, 4)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(462672)
			return forEachTableDescAll(ctx, p, dbContext, hideVirtual,
				func(db catalog.DatabaseDescriptor, _ string, table catalog.TableDescriptor) error {
					__antithesis_instrumentation__.Notify(462673)
					tableID := table.GetID()
					return catalog.ForEachIndex(table, catalog.IndexOpts{}, func(idx catalog.Index) error {
						__antithesis_instrumentation__.Notify(462674)
						indexID := idx.GetID()
						stats := indexStats.Get(roachpb.TableID(tableID), roachpb.IndexID(indexID))

						lastScanTs := tree.DNull
						if !stats.LastRead.IsZero() {
							__antithesis_instrumentation__.Notify(462676)
							lastScanTs, err = tree.MakeDTimestampTZ(stats.LastRead, time.Nanosecond)
							if err != nil {
								__antithesis_instrumentation__.Notify(462677)
								return err
							} else {
								__antithesis_instrumentation__.Notify(462678)
							}
						} else {
							__antithesis_instrumentation__.Notify(462679)
						}
						__antithesis_instrumentation__.Notify(462675)

						row = row[:0]

						row = append(row,
							tree.NewDInt(tree.DInt(tableID)),
							tree.NewDInt(tree.DInt(indexID)),
							tree.NewDInt(tree.DInt(stats.TotalReadCount)),
							lastScanTs,
						)

						return pusher.pushRow(row...)
					})
				})
		}
		__antithesis_instrumentation__.Notify(462669)
		return setupGenerator(ctx, worker, stopper)
	},
}

var crdbInternalClusterStmtStatsTable = virtualSchemaTable{
	comment: `cluster-wide statement statistics that have not yet been flushed ` +
		`to system tables. Querying this table is a somewhat expensive operation ` +
		`since it creates a cluster-wide RPC-fanout.`,
	schema: `
CREATE TABLE crdb_internal.cluster_statement_statistics (
    aggregated_ts              TIMESTAMPTZ NOT NULL,
    fingerprint_id             BYTES NOT NULL,
    transaction_fingerprint_id BYTES NOT NULL,
    plan_hash                  BYTES NOT NULL,
    app_name                   STRING NOT NULL,
    metadata                   JSONB NOT NULL,
    statistics                 JSONB NOT NULL,
    sampled_plan               JSONB NOT NULL,
    aggregation_interval       INTERVAL NOT NULL
);`,
	generator: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(462680)

		acc := p.extendedEvalCtx.Mon.MakeBoundAccount()
		defer acc.Close(ctx)

		stats, err :=
			p.extendedEvalCtx.SQLStatusServer.Statements(ctx, &serverpb.StatementsRequest{
				FetchMode: serverpb.StatementsRequest_StmtStatsOnly,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(462685)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462686)
		}
		__antithesis_instrumentation__.Notify(462681)

		statsMemSize := stats.Size()
		if err = acc.Grow(ctx, int64(statsMemSize)); err != nil {
			__antithesis_instrumentation__.Notify(462687)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462688)
		}
		__antithesis_instrumentation__.Notify(462682)

		memSQLStats, err := sslocal.NewTempSQLStatsFromExistingStmtStats(stats.Statements)
		if err != nil {
			__antithesis_instrumentation__.Notify(462689)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462690)
		}
		__antithesis_instrumentation__.Notify(462683)

		s := p.extendedEvalCtx.statsProvider
		curAggTs := s.ComputeAggregatedTs()
		aggInterval := s.GetAggregationInterval()

		row := make(tree.Datums, 8)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(462691)
			return memSQLStats.IterateStatementStats(ctx, &sqlstats.IteratorOptions{
				SortedAppNames: true,
				SortedKey:      true,
			}, func(ctx context.Context, statistics *roachpb.CollectedStatementStatistics) error {
				__antithesis_instrumentation__.Notify(462692)

				aggregatedTs, err := tree.MakeDTimestampTZ(curAggTs, time.Microsecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(462696)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462697)
				}
				__antithesis_instrumentation__.Notify(462693)

				fingerprintID := tree.NewDBytes(
					tree.DBytes(sqlstatsutil.EncodeUint64ToBytes(uint64(statistics.ID))))

				transactionFingerprintID := tree.NewDBytes(
					tree.DBytes(sqlstatsutil.EncodeUint64ToBytes(uint64(statistics.Key.TransactionFingerprintID))))

				planHash := tree.NewDBytes(
					tree.DBytes(sqlstatsutil.EncodeUint64ToBytes(statistics.Key.PlanHash)))

				metadataJSON, err := sqlstatsutil.BuildStmtMetadataJSON(statistics)
				if err != nil {
					__antithesis_instrumentation__.Notify(462698)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462699)
				}
				__antithesis_instrumentation__.Notify(462694)
				statisticsJSON, err := sqlstatsutil.BuildStmtStatisticsJSON(&statistics.Stats)
				if err != nil {
					__antithesis_instrumentation__.Notify(462700)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462701)
				}
				__antithesis_instrumentation__.Notify(462695)
				plan := sqlstatsutil.ExplainTreePlanNodeToJSON(&statistics.Stats.SensitiveInfo.MostRecentPlanDescription)

				aggInterval := tree.NewDInterval(
					duration.MakeDuration(aggInterval.Nanoseconds(), 0, 0),
					types.DefaultIntervalTypeMetadata)

				row = row[:0]
				row = append(row,
					aggregatedTs,
					fingerprintID,
					transactionFingerprintID,
					planHash,
					tree.NewDString(statistics.Key.App),
					tree.NewDJSON(metadataJSON),
					tree.NewDJSON(statisticsJSON),
					tree.NewDJSON(plan),
					aggInterval,
				)

				return pusher.pushRow(row...)
			})
		}
		__antithesis_instrumentation__.Notify(462684)
		return setupGenerator(ctx, worker, stopper)
	},
}

var crdbInternalStmtStatsView = virtualSchemaView{
	schema: `
CREATE VIEW crdb_internal.statement_statistics AS
SELECT
  aggregated_ts,
  fingerprint_id,
  transaction_fingerprint_id,
  plan_hash,
  app_name,
  max(metadata) as metadata,
  crdb_internal.merge_statement_stats(array_agg(statistics)),
  max(sampled_plan),
  aggregation_interval
FROM (
  SELECT
      aggregated_ts,
      fingerprint_id,
      transaction_fingerprint_id,
      plan_hash,
      app_name,
      metadata,
      statistics,
      sampled_plan,
      aggregation_interval
  FROM
      crdb_internal.cluster_statement_statistics
  UNION ALL
      SELECT
          aggregated_ts,
          fingerprint_id,
          transaction_fingerprint_id,
          plan_hash,
          app_name,
          metadata,
          statistics,
          plan,
          agg_interval
      FROM
          system.statement_statistics
)
GROUP BY
  aggregated_ts,
  fingerprint_id,
  transaction_fingerprint_id,
  plan_hash,
  app_name,
  aggregation_interval`,
	resultColumns: colinfo.ResultColumns{
		{Name: "aggregated_ts", Typ: types.TimestampTZ},
		{Name: "fingerprint_id", Typ: types.Bytes},
		{Name: "transaction_fingerprint_id", Typ: types.Bytes},
		{Name: "plan_hash", Typ: types.Bytes},
		{Name: "app_name", Typ: types.String},
		{Name: "metadata", Typ: types.Jsonb},
		{Name: "statistics", Typ: types.Jsonb},
		{Name: "sampled_plan", Typ: types.Jsonb},
		{Name: "aggregation_interval", Typ: types.Interval},
	},
}

var crdbInternalActiveRangeFeedsTable = virtualSchemaTable{
	comment: `node-level table listing all currently running range feeds`,
	schema: `
CREATE TABLE crdb_internal.active_range_feeds (
  id INT,
  tags STRING,
  startTS STRING,
  diff BOOL,
  node_id INT,
  range_id INT,
  created INT,
  range_start STRING,
  range_end STRING,
  resolved STRING,
  last_event_utc INT
);`,
	populate: func(ctx context.Context, p *planner, _ catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(462702)
		return p.execCfg.DistSender.ForEachActiveRangeFeed(
			func(rfCtx kvcoord.RangeFeedContext, rf kvcoord.PartialRangeFeed) error {
				__antithesis_instrumentation__.Notify(462703)
				var lastEvent tree.Datum
				if rf.LastValueReceived.IsZero() {
					__antithesis_instrumentation__.Notify(462705)
					lastEvent = tree.DNull
				} else {
					__antithesis_instrumentation__.Notify(462706)
					lastEvent = tree.NewDInt(tree.DInt(rf.LastValueReceived.UTC().UnixNano()))
				}
				__antithesis_instrumentation__.Notify(462704)

				return addRow(
					tree.NewDInt(tree.DInt(rfCtx.ID)),
					tree.NewDString(rfCtx.CtxTags),
					tree.NewDString(rfCtx.StartFrom.AsOfSystemTime()),
					tree.MakeDBool(tree.DBool(rfCtx.WithDiff)),
					tree.NewDInt(tree.DInt(rf.NodeID)),
					tree.NewDInt(tree.DInt(rf.RangeID)),
					tree.NewDInt(tree.DInt(rf.CreatedTime.UTC().UnixNano())),
					tree.NewDString(keys.PrettyPrint(nil, rf.Span.Key)),
					tree.NewDString(keys.PrettyPrint(nil, rf.Span.EndKey)),
					tree.NewDString(rf.Resolved.AsOfSystemTime()),
					lastEvent,
				)
			},
		)
	},
}

var crdbInternalClusterTxnStatsTable = virtualSchemaTable{
	comment: `cluster-wide transaction statistics that have not yet been flushed ` +
		`to system tables. Querying this table is a somewhat expensive operation ` +
		`since it creates a cluster-wide RPC-fanout.`,
	schema: `
CREATE TABLE crdb_internal.cluster_transaction_statistics (
    aggregated_ts         TIMESTAMPTZ NOT NULL,
    fingerprint_id        BYTES NOT NULL,
    app_name              STRING NOT NULL,
    metadata              JSONB NOT NULL,
    statistics            JSONB NOT NULL,
    aggregation_interval  INTERVAL NOT NULL
);`,
	generator: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(462707)

		acc := p.extendedEvalCtx.Mon.MakeBoundAccount()
		defer acc.Close(ctx)

		stats, err :=
			p.extendedEvalCtx.SQLStatusServer.Statements(ctx, &serverpb.StatementsRequest{
				FetchMode: serverpb.StatementsRequest_TxnStatsOnly,
			})

		if err != nil {
			__antithesis_instrumentation__.Notify(462712)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462713)
		}
		__antithesis_instrumentation__.Notify(462708)

		statsMemSize := stats.Size()
		if err = acc.Grow(ctx, int64(statsMemSize)); err != nil {
			__antithesis_instrumentation__.Notify(462714)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462715)
		}
		__antithesis_instrumentation__.Notify(462709)

		memSQLStats, err :=
			sslocal.NewTempSQLStatsFromExistingTxnStats(stats.Transactions)
		if err != nil {
			__antithesis_instrumentation__.Notify(462716)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462717)
		}
		__antithesis_instrumentation__.Notify(462710)

		s := p.extendedEvalCtx.statsProvider
		curAggTs := s.ComputeAggregatedTs()
		aggInterval := s.GetAggregationInterval()

		row := make(tree.Datums, 5)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(462718)
			return memSQLStats.IterateTransactionStats(ctx, &sqlstats.IteratorOptions{
				SortedAppNames: true,
				SortedKey:      true,
			}, func(
				ctx context.Context,
				statistics *roachpb.CollectedTransactionStatistics) error {
				__antithesis_instrumentation__.Notify(462719)

				aggregatedTs, err := tree.MakeDTimestampTZ(curAggTs, time.Microsecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(462723)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462724)
				}
				__antithesis_instrumentation__.Notify(462720)

				fingerprintID := tree.NewDBytes(
					tree.DBytes(sqlstatsutil.EncodeUint64ToBytes(uint64(statistics.TransactionFingerprintID))))

				metadataJSON, err := sqlstatsutil.BuildTxnMetadataJSON(statistics)
				if err != nil {
					__antithesis_instrumentation__.Notify(462725)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462726)
				}
				__antithesis_instrumentation__.Notify(462721)
				statisticsJSON, err := sqlstatsutil.BuildTxnStatisticsJSON(statistics)
				if err != nil {
					__antithesis_instrumentation__.Notify(462727)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462728)
				}
				__antithesis_instrumentation__.Notify(462722)

				aggInterval := tree.NewDInterval(
					duration.MakeDuration(aggInterval.Nanoseconds(), 0, 0),
					types.DefaultIntervalTypeMetadata)

				row = row[:0]
				row = append(row,
					aggregatedTs,
					fingerprintID,
					tree.NewDString(statistics.App),
					tree.NewDJSON(metadataJSON),
					tree.NewDJSON(statisticsJSON),
					aggInterval,
				)

				return pusher.pushRow(row...)
			})
		}
		__antithesis_instrumentation__.Notify(462711)
		return setupGenerator(ctx, worker, stopper)
	},
}

var crdbInternalTxnStatsView = virtualSchemaView{
	schema: `
CREATE VIEW crdb_internal.transaction_statistics AS
SELECT
  aggregated_ts,
  fingerprint_id,
  app_name,
  max(metadata),
  crdb_internal.merge_transaction_stats(array_agg(statistics)),
  aggregation_interval
FROM (
  SELECT
    aggregated_ts,
    fingerprint_id,
    app_name,
    metadata,
    statistics,
    aggregation_interval
  FROM
    crdb_internal.cluster_transaction_statistics
  UNION ALL
      SELECT
        aggregated_ts,
        fingerprint_id,
        app_name,
        metadata,
        statistics,
        agg_interval
      FROM
        system.transaction_statistics
)
GROUP BY
  aggregated_ts,
  fingerprint_id,
  app_name,
  aggregation_interval`,
	resultColumns: colinfo.ResultColumns{
		{Name: "aggregated_ts", Typ: types.TimestampTZ},
		{Name: "fingerprint_id", Typ: types.Bytes},
		{Name: "app_name", Typ: types.String},
		{Name: "metadata", Typ: types.Jsonb},
		{Name: "statistics", Typ: types.Jsonb},
		{Name: "aggregation_interval", Typ: types.Interval},
	},
}

var crdbInternalTenantUsageDetailsView = virtualSchemaView{
	schema: `
CREATE VIEW crdb_internal.tenant_usage_details AS
  SELECT
    tenant_id,
    (j->>'rU')::FLOAT8 AS total_ru,
    (j->>'readBytes')::INT8 AS total_read_bytes,
    (j->>'readRequests')::INT8 AS total_read_requests,
    (j->>'writeBytes')::INT8 AS total_write_bytes,
    (j->>'writeRequests')::INT8 AS total_write_requests,
    (j->>'sqlPodsCpuSeconds')::FLOAT8 AS total_sql_pod_seconds,
    (j->>'pgwireEgressBytes')::INT8 AS total_pgwire_egress_bytes
  FROM
    (
      SELECT
        tenant_id,
        crdb_internal.pb_to_json('cockroach.roachpb.TenantConsumption', total_consumption) AS j
      FROM
        system.tenant_usage
      WHERE
        instance_id = 0
    )
`,
	resultColumns: colinfo.ResultColumns{
		{Name: "tenant_id", Typ: types.Int},
		{Name: "total_ru", Typ: types.Float},
		{Name: "total_read_bytes", Typ: types.Int},
		{Name: "total_read_requests", Typ: types.Int},
		{Name: "total_write_bytes", Typ: types.Int},
		{Name: "total_write_requests", Typ: types.Int},
		{Name: "total_sql_pod_seconds", Typ: types.Float},
		{Name: "total_pgwire_egress_bytes", Typ: types.Int},
	},
}

var crdbInternalTransactionContentionEventsTable = virtualSchemaTable{
	comment: `cluster-wide transaction contention events. Querying this table is an
		expensive operation since it creates a cluster-wide RPC-fanout.`,
	schema: `
CREATE TABLE crdb_internal.transaction_contention_events (
    collection_ts                TIMESTAMPTZ NOT NULL,

    blocking_txn_id              UUID NOT NULL,
    blocking_txn_fingerprint_id  BYTES NOT NULL,

    waiting_txn_id               UUID NOT NULL,
    waiting_txn_fingerprint_id   BYTES NOT NULL,

    contention_duration          INTERVAL NOT NULL,
    contending_key               BYTES NOT NULL
);`,
	generator: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(462729)

		hasPermission, err := p.HasViewActivityOrViewActivityRedactedRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(462737)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462738)
		}
		__antithesis_instrumentation__.Notify(462730)
		if !hasPermission {
			__antithesis_instrumentation__.Notify(462739)
			return nil, nil, errors.New("crdb_internal.transaction_contention_events " +
				"requires VIEWACTIVITY or VIEWACTIVITYREDACTED role option")
		} else {
			__antithesis_instrumentation__.Notify(462740)
		}
		__antithesis_instrumentation__.Notify(462731)

		isAdmin, err := p.HasAdminRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(462741)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462742)
		}
		__antithesis_instrumentation__.Notify(462732)

		shouldRedactContendingKey := false
		if !isAdmin {
			__antithesis_instrumentation__.Notify(462743)
			shouldRedactContendingKey, err = p.HasRoleOption(ctx, roleoption.VIEWACTIVITYREDACTED)
			if err != nil {
				__antithesis_instrumentation__.Notify(462744)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(462745)
			}
		} else {
			__antithesis_instrumentation__.Notify(462746)
		}
		__antithesis_instrumentation__.Notify(462733)

		acc := p.extendedEvalCtx.Mon.MakeBoundAccount()
		defer acc.Close(ctx)

		resp, err := p.extendedEvalCtx.SQLStatusServer.TransactionContentionEvents(
			ctx, &serverpb.TransactionContentionEventsRequest{})

		if err != nil {
			__antithesis_instrumentation__.Notify(462747)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462748)
		}
		__antithesis_instrumentation__.Notify(462734)

		respSize := resp.Size()
		if err = acc.Grow(ctx, int64(respSize)); err != nil {
			__antithesis_instrumentation__.Notify(462749)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462750)
		}
		__antithesis_instrumentation__.Notify(462735)

		row := make(tree.Datums, 6)
		worker := func(ctx context.Context, pusher rowPusher) error {
			__antithesis_instrumentation__.Notify(462751)
			for i := range resp.Events {
				__antithesis_instrumentation__.Notify(462753)
				collectionTs, err := tree.MakeDTimestampTZ(resp.Events[i].CollectionTs, time.Microsecond)
				if err != nil {
					__antithesis_instrumentation__.Notify(462756)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462757)
				}
				__antithesis_instrumentation__.Notify(462754)
				blockingFingerprintID := tree.NewDBytes(
					tree.DBytes(sqlstatsutil.EncodeUint64ToBytes(uint64(resp.Events[i].BlockingTxnFingerprintID))))

				waitingFingerprintID := tree.NewDBytes(
					tree.DBytes(sqlstatsutil.EncodeUint64ToBytes(uint64(resp.Events[i].WaitingTxnFingerprintID))))

				contentionDuration := tree.NewDInterval(
					duration.MakeDuration(
						resp.Events[i].BlockingEvent.Duration.Nanoseconds(),
						0,
						0,
					),
					types.DefaultIntervalTypeMetadata,
				)

				contendingKey := tree.NewDBytes("")
				if !shouldRedactContendingKey {
					__antithesis_instrumentation__.Notify(462758)
					contendingKey = tree.NewDBytes(
						tree.DBytes(resp.Events[i].BlockingEvent.Key))
				} else {
					__antithesis_instrumentation__.Notify(462759)
				}
				__antithesis_instrumentation__.Notify(462755)

				row = row[:0]
				row = append(row,
					collectionTs,
					tree.NewDUuid(tree.DUuid{UUID: resp.Events[i].BlockingEvent.TxnMeta.ID}),
					blockingFingerprintID,
					tree.NewDUuid(tree.DUuid{UUID: resp.Events[i].WaitingTxnID}),
					waitingFingerprintID,
					contentionDuration,
					contendingKey,
				)

				if err = pusher.pushRow(row...); err != nil {
					__antithesis_instrumentation__.Notify(462760)
					return err
				} else {
					__antithesis_instrumentation__.Notify(462761)
				}
			}
			__antithesis_instrumentation__.Notify(462752)
			return nil
		}
		__antithesis_instrumentation__.Notify(462736)
		return setupGenerator(ctx, worker, stopper)
	},
}

var crdbInternalClusterLocksTable = virtualSchemaTable{
	comment: `cluster-wide locks held in lock tables. Querying this table is an
		expensive operation since it creates a cluster-wide RPC-fanout.`,
	schema: `
CREATE TABLE crdb_internal.cluster_locks (
    range_id            INT NOT NULL,
    table_id            INT NOT NULL,
    database_name       STRING,
    schema_name         STRING,
    table_name          STRING,
    index_name          STRING,
    lock_key            BYTES NOT NULL,
    lock_key_pretty     STRING NOT NULL,
    txn_id              UUID,
    ts                  TIMESTAMP,
    lock_strength       STRING,
    durability          STRING,
    granted             BOOL,
    contended           BOOL,
    duration            INTERVAL
);`,
	indexes: nil,
	generator: func(ctx context.Context, p *planner, db catalog.DatabaseDescriptor, stopper *stop.Stopper) (virtualTableGenerator, cleanupFunc, error) {
		__antithesis_instrumentation__.Notify(462762)

		if !p.ExecCfg().Settings.Version.IsActive(ctx, clusterversion.ClusterLocksVirtualTable) {
			__antithesis_instrumentation__.Notify(462775)
			return nil, nil, pgerror.New(pgcode.FeatureNotSupported,
				"table crdb_internal.cluster_locks is not supported on this version")
		} else {
			__antithesis_instrumentation__.Notify(462776)
		}
		__antithesis_instrumentation__.Notify(462763)

		hasAdmin, err := p.HasAdminRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(462777)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462778)
		}
		__antithesis_instrumentation__.Notify(462764)
		hasViewActivityOrViewActivityRedacted, err := p.HasViewActivityOrViewActivityRedactedRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(462779)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462780)
		}
		__antithesis_instrumentation__.Notify(462765)
		if !hasViewActivityOrViewActivityRedacted {
			__antithesis_instrumentation__.Notify(462781)
			return nil, nil, pgerror.Newf(pgcode.InsufficientPrivilege,
				"user %s does not have %s or %s privilege", p.User(), roleoption.VIEWACTIVITY, roleoption.VIEWACTIVITYREDACTED)
		} else {
			__antithesis_instrumentation__.Notify(462782)
		}
		__antithesis_instrumentation__.Notify(462766)
		shouldRedactKeys := false
		if !hasAdmin {
			__antithesis_instrumentation__.Notify(462783)
			shouldRedactKeys, err = p.HasRoleOption(ctx, roleoption.VIEWACTIVITYREDACTED)
			if err != nil {
				__antithesis_instrumentation__.Notify(462784)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(462785)
			}
		} else {
			__antithesis_instrumentation__.Notify(462786)
		}
		__antithesis_instrumentation__.Notify(462767)

		all, err := p.Descriptors().GetAllDescriptors(ctx, p.txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(462787)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(462788)
		}
		__antithesis_instrumentation__.Notify(462768)
		descs := all.OrderedDescriptors()

		privCheckerFunc := func(desc catalog.Descriptor) bool {
			__antithesis_instrumentation__.Notify(462789)
			if hasAdmin {
				__antithesis_instrumentation__.Notify(462791)
				return true
			} else {
				__antithesis_instrumentation__.Notify(462792)
			}
			__antithesis_instrumentation__.Notify(462790)
			return p.CheckAnyPrivilege(ctx, desc) == nil
		}
		__antithesis_instrumentation__.Notify(462769)

		_, dbNames, tableNames, schemaNames, indexNames, schemaParents, parents :=
			descriptorsByType(descs, privCheckerFunc)

		var spansToQuery roachpb.Spans
		for _, desc := range descs {
			__antithesis_instrumentation__.Notify(462793)
			if !privCheckerFunc(desc) {
				__antithesis_instrumentation__.Notify(462795)
				continue
			} else {
				__antithesis_instrumentation__.Notify(462796)
			}
			__antithesis_instrumentation__.Notify(462794)
			switch desc := desc.(type) {
			case catalog.TableDescriptor:
				__antithesis_instrumentation__.Notify(462797)
				spansToQuery = append(spansToQuery, desc.TableSpan(p.execCfg.Codec))
			}
		}
		__antithesis_instrumentation__.Notify(462770)

		spanIdx := 0
		spansRemain := func() bool {
			__antithesis_instrumentation__.Notify(462798)
			return spanIdx < len(spansToQuery)
		}
		__antithesis_instrumentation__.Notify(462771)
		getNextSpan := func() *roachpb.Span {
			__antithesis_instrumentation__.Notify(462799)
			if !spansRemain() {
				__antithesis_instrumentation__.Notify(462801)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(462802)
			}
			__antithesis_instrumentation__.Notify(462800)

			nextSpan := spansToQuery[spanIdx]
			spanIdx++
			return &nextSpan
		}
		__antithesis_instrumentation__.Notify(462772)

		var resp *roachpb.QueryLocksResponse
		var locks []roachpb.LockStateInfo
		var resumeSpan *roachpb.Span

		fetchLocks := func(key, endKey roachpb.Key) error {
			__antithesis_instrumentation__.Notify(462803)
			b := kv.Batch{}
			queryLocksRequest := &roachpb.QueryLocksRequest{
				RequestHeader: roachpb.RequestHeader{
					Key:    key,
					EndKey: endKey,
				},
				IncludeUncontended: true,
			}
			b.AddRawRequest(queryLocksRequest)

			b.Header.MaxSpanRequestKeys = int64(rowinfra.ProductionKVBatchSize)
			b.Header.TargetBytes = int64(rowinfra.DefaultBatchBytesLimit)

			err := p.txn.Run(ctx, &b)
			if err != nil {
				__antithesis_instrumentation__.Notify(462806)
				return err
			} else {
				__antithesis_instrumentation__.Notify(462807)
			}
			__antithesis_instrumentation__.Notify(462804)

			if len(b.RawResponse().Responses) != 1 {
				__antithesis_instrumentation__.Notify(462808)
				return errors.AssertionFailedf(
					"unexpected response length of %d for QueryLocksRequest", len(b.RawResponse().Responses),
				)
			} else {
				__antithesis_instrumentation__.Notify(462809)
			}
			__antithesis_instrumentation__.Notify(462805)

			resp = b.RawResponse().Responses[0].GetQueryLocks()
			locks = resp.Locks
			resumeSpan = resp.ResumeSpan
			return nil
		}
		__antithesis_instrumentation__.Notify(462773)

		lockIdx := 0
		getNextLock := func() (*roachpb.LockStateInfo, error) {
			__antithesis_instrumentation__.Notify(462810)

			for lockIdx >= len(locks) && func() bool {
				__antithesis_instrumentation__.Notify(462813)
				return (spansRemain() || func() bool {
					__antithesis_instrumentation__.Notify(462814)
					return resumeSpan != nil == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(462815)
				var spanToQuery *roachpb.Span
				if resumeSpan != nil {
					__antithesis_instrumentation__.Notify(462817)
					spanToQuery = resumeSpan
				} else {
					__antithesis_instrumentation__.Notify(462818)
					spanToQuery = getNextSpan()
				}
				__antithesis_instrumentation__.Notify(462816)

				if spanToQuery != nil {
					__antithesis_instrumentation__.Notify(462819)
					err := fetchLocks(spanToQuery.Key, spanToQuery.EndKey)
					if err != nil {
						__antithesis_instrumentation__.Notify(462821)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(462822)
					}
					__antithesis_instrumentation__.Notify(462820)
					lockIdx = 0
				} else {
					__antithesis_instrumentation__.Notify(462823)
				}
			}
			__antithesis_instrumentation__.Notify(462811)

			if lockIdx < len(locks) {
				__antithesis_instrumentation__.Notify(462824)
				nextLock := locks[lockIdx]
				lockIdx++
				return &nextLock, nil
			} else {
				__antithesis_instrumentation__.Notify(462825)
			}
			__antithesis_instrumentation__.Notify(462812)

			return nil, nil
		}
		__antithesis_instrumentation__.Notify(462774)

		var curLock *roachpb.LockStateInfo
		var fErr error
		waiterIdx := -1

		return func() (tree.Datums, error) {
			__antithesis_instrumentation__.Notify(462826)
			if curLock == nil || func() bool {
				__antithesis_instrumentation__.Notify(462831)
				return waiterIdx >= len(curLock.Waiters) == true
			}() == true {
				__antithesis_instrumentation__.Notify(462832)
				curLock, fErr = getNextLock()
				waiterIdx = -1
			} else {
				__antithesis_instrumentation__.Notify(462833)
			}
			__antithesis_instrumentation__.Notify(462827)

			if curLock == nil || func() bool {
				__antithesis_instrumentation__.Notify(462834)
				return fErr != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(462835)
				return nil, fErr
			} else {
				__antithesis_instrumentation__.Notify(462836)
			}
			__antithesis_instrumentation__.Notify(462828)

			strengthDatum := tree.DNull
			txnIDDatum := tree.DNull
			tsDatum := tree.DNull
			durationDatum := tree.DNull
			granted := false

			if waiterIdx < 0 {
				__antithesis_instrumentation__.Notify(462837)
				if curLock.LockHolder != nil {
					__antithesis_instrumentation__.Notify(462838)
					txnIDDatum = tree.NewDUuid(tree.DUuid{UUID: curLock.LockHolder.ID})
					tsDatum = tree.TimestampToInexactDTimestamp(curLock.LockHolder.WriteTimestamp)
					strengthDatum = tree.NewDString(lock.Exclusive.String())
					durationDatum = tree.NewDInterval(
						duration.MakeDuration(curLock.HoldDuration.Nanoseconds(), 0, 0),
						types.DefaultIntervalTypeMetadata,
					)
					granted = true
				} else {
					__antithesis_instrumentation__.Notify(462839)
				}
			} else {
				__antithesis_instrumentation__.Notify(462840)
				waiter := curLock.Waiters[waiterIdx]
				if waiter.WaitingTxn != nil {
					__antithesis_instrumentation__.Notify(462842)
					txnIDDatum = tree.NewDUuid(tree.DUuid{UUID: waiter.WaitingTxn.ID})
					tsDatum = tree.TimestampToInexactDTimestamp(waiter.WaitingTxn.WriteTimestamp)
				} else {
					__antithesis_instrumentation__.Notify(462843)
				}
				__antithesis_instrumentation__.Notify(462841)
				strengthDatum = tree.NewDString(waiter.Strength.String())
				durationDatum = tree.NewDInterval(
					duration.MakeDuration(waiter.WaitDuration.Nanoseconds(), 0, 0),
					types.DefaultIntervalTypeMetadata,
				)
			}
			__antithesis_instrumentation__.Notify(462829)

			waiterIdx++

			tableID, dbName, schemaName, tableName, indexName := lookupNamesByKey(
				p, curLock.Key, dbNames, tableNames, schemaNames,
				indexNames, schemaParents, parents,
			)

			var keyOrRedacted roachpb.Key
			var prettyKeyOrRedacted string
			if !shouldRedactKeys {
				__antithesis_instrumentation__.Notify(462844)
				keyOrRedacted, _, _ = keys.DecodeTenantPrefix(curLock.Key)
				prettyKeyOrRedacted = keys.PrettyPrint(nil, keyOrRedacted)
			} else {
				__antithesis_instrumentation__.Notify(462845)
			}
			__antithesis_instrumentation__.Notify(462830)

			return tree.Datums{
				tree.NewDInt(tree.DInt(curLock.RangeID)),
				tree.NewDInt(tree.DInt(tableID)),
				tree.NewDString(dbName),
				tree.NewDString(schemaName),
				tree.NewDString(tableName),
				tree.NewDString(indexName),
				tree.NewDBytes(tree.DBytes(keyOrRedacted)),
				tree.NewDString(prettyKeyOrRedacted),
				txnIDDatum,
				tsDatum,
				strengthDatum,
				tree.NewDString(curLock.Durability.String()),
				tree.MakeDBool(tree.DBool(granted)),
				tree.MakeDBool(len(curLock.Waiters) > 0),
				durationDatum,
			}, nil

		}, nil, nil
	},
}
