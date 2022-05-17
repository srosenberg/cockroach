package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	debugBase         = "debug"
	eventsName        = debugBase + "/events"
	livenessName      = debugBase + "/liveness"
	nodesPrefix       = debugBase + "/nodes"
	rangelogName      = debugBase + "/rangelog"
	reportsPrefix     = debugBase + "/reports"
	schemaPrefix      = debugBase + "/schema"
	settingsName      = debugBase + "/settings"
	problemRangesName = reportsPrefix + "/problemranges"
	tenantRangesName  = debugBase + "/tenant_ranges"
)

func makeClusterWideZipRequests(
	admin serverpb.AdminClient, status serverpb.StatusClient,
) []zipRequest {
	__antithesis_instrumentation__.Notify(35288)
	return []zipRequest{

		{
			fn: func(ctx context.Context) (interface{}, error) {
				__antithesis_instrumentation__.Notify(35289)
				return admin.Events(ctx, &serverpb.EventsRequest{})
			},
			pathName: eventsName,
		},
		{
			fn: func(ctx context.Context) (interface{}, error) {
				__antithesis_instrumentation__.Notify(35290)
				return admin.RangeLog(ctx, &serverpb.RangeLogRequest{})
			},
			pathName: rangelogName,
		},
		{
			fn: func(ctx context.Context) (interface{}, error) {
				__antithesis_instrumentation__.Notify(35291)
				return admin.Settings(ctx, &serverpb.SettingsRequest{})
			},
			pathName: settingsName,
		},
		{
			fn: func(ctx context.Context) (interface{}, error) {
				__antithesis_instrumentation__.Notify(35292)
				return status.ProblemRanges(ctx, &serverpb.ProblemRangesRequest{})
			},
			pathName: problemRangesName,
		},
	}
}

var debugZipTablesPerCluster = []string{
	"crdb_internal.cluster_contention_events",
	"crdb_internal.cluster_distsql_flows",
	"crdb_internal.cluster_database_privileges",
	"crdb_internal.cluster_locks",
	"crdb_internal.cluster_queries",
	"crdb_internal.cluster_sessions",
	"crdb_internal.cluster_settings",
	"crdb_internal.cluster_transactions",

	"crdb_internal.default_privileges",

	"crdb_internal.jobs",
	"system.jobs",
	"system.descriptor",
	"system.namespace",
	"system.scheduled_jobs",
	"system.settings",
	"system.replication_stats",
	"system.replication_critical_localities",
	"system.replication_constraint_stats",

	`"".crdb_internal.create_schema_statements`,
	`"".crdb_internal.create_statements`,

	`"".crdb_internal.create_type_statements`,

	"crdb_internal.kv_node_liveness",
	"crdb_internal.kv_node_status",
	"crdb_internal.kv_store_status",

	"crdb_internal.regions",
	"crdb_internal.schema_changes",
	"crdb_internal.super_regions",
	"crdb_internal.partitions",
	"crdb_internal.zones",
	"crdb_internal.invalid_objects",
	"crdb_internal.index_usage_statistics",
	"crdb_internal.table_indexes",
	"crdb_internal.transaction_contention_events",
}

type nodesInfo struct {
	nodesStatusResponse *serverpb.NodesResponse
	nodesListResponse   *serverpb.NodesListResponse
}

func (zc *debugZipContext) collectClusterData(
	ctx context.Context, firstNodeDetails *serverpb.DetailsResponse,
) (ni nodesInfo, livenessByNodeID nodeLivenesses, err error) {
	__antithesis_instrumentation__.Notify(35293)
	clusterWideZipRequests := makeClusterWideZipRequests(zc.admin, zc.status)

	for _, r := range clusterWideZipRequests {
		__antithesis_instrumentation__.Notify(35296)
		if err := zc.runZipRequest(ctx, zc.clusterPrinter, r); err != nil {
			__antithesis_instrumentation__.Notify(35297)
			return nodesInfo{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(35298)
		}
	}
	__antithesis_instrumentation__.Notify(35294)

	for _, table := range debugZipTablesPerCluster {
		__antithesis_instrumentation__.Notify(35299)
		query := fmt.Sprintf(`SELECT * FROM %s`, table)
		if override, ok := customQuery[table]; ok {
			__antithesis_instrumentation__.Notify(35301)
			query = override
		} else {
			__antithesis_instrumentation__.Notify(35302)
		}
		__antithesis_instrumentation__.Notify(35300)
		if err := zc.dumpTableDataForZip(zc.clusterPrinter, zc.firstNodeSQLConn, debugBase, table, query); err != nil {
			__antithesis_instrumentation__.Notify(35303)
			return nodesInfo{}, nil, errors.Wrapf(err, "fetching %s", table)
		} else {
			__antithesis_instrumentation__.Notify(35304)
		}
	}

	{
		__antithesis_instrumentation__.Notify(35305)
		s := zc.clusterPrinter.start("requesting nodes")
		err := zc.runZipFn(ctx, s, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35311)
			ni, err = zc.nodesInfo(ctx)
			return err
		})
		__antithesis_instrumentation__.Notify(35306)
		if ni.nodesStatusResponse != nil {
			__antithesis_instrumentation__.Notify(35312)
			if cErr := zc.z.createJSONOrError(s, debugBase+"/nodes.json", ni.nodesStatusResponse, err); cErr != nil {
				__antithesis_instrumentation__.Notify(35313)
				return nodesInfo{}, nil, cErr
			} else {
				__antithesis_instrumentation__.Notify(35314)
			}
		} else {
			__antithesis_instrumentation__.Notify(35315)
			if cErr := zc.z.createJSONOrError(s, debugBase+"/nodes.json", ni.nodesListResponse, err); cErr != nil {
				__antithesis_instrumentation__.Notify(35316)
				return nodesInfo{}, nil, cErr
			} else {
				__antithesis_instrumentation__.Notify(35317)
			}
		}
		__antithesis_instrumentation__.Notify(35307)

		if ni.nodesListResponse == nil {
			__antithesis_instrumentation__.Notify(35318)

			ni.nodesListResponse = &serverpb.NodesListResponse{
				Nodes: []serverpb.NodeDetails{{
					NodeID:     int32(firstNodeDetails.NodeID),
					Address:    firstNodeDetails.Address,
					SQLAddress: firstNodeDetails.SQLAddress,
				}},
			}
		} else {
			__antithesis_instrumentation__.Notify(35319)
		}
		__antithesis_instrumentation__.Notify(35308)

		var lresponse *serverpb.LivenessResponse
		s = zc.clusterPrinter.start("requesting liveness")
		err = zc.runZipFn(ctx, s, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35320)
			lresponse, err = zc.admin.Liveness(ctx, &serverpb.LivenessRequest{})
			return err
		})
		__antithesis_instrumentation__.Notify(35309)
		if cErr := zc.z.createJSONOrError(s, livenessName+".json", nodes, err); cErr != nil {
			__antithesis_instrumentation__.Notify(35321)
			return nodesInfo{}, nil, cErr
		} else {
			__antithesis_instrumentation__.Notify(35322)
		}
		__antithesis_instrumentation__.Notify(35310)
		livenessByNodeID = map[roachpb.NodeID]livenesspb.NodeLivenessStatus{}
		if lresponse != nil {
			__antithesis_instrumentation__.Notify(35323)
			livenessByNodeID = lresponse.Statuses
		} else {
			__antithesis_instrumentation__.Notify(35324)
		}
	}

	{
		__antithesis_instrumentation__.Notify(35325)
		var tenantRanges *serverpb.TenantRangesResponse
		s := zc.clusterPrinter.start("requesting tenant ranges")
		if requestErr := zc.runZipFn(ctx, s, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35326)
			var err error
			tenantRanges, err = zc.status.TenantRanges(ctx, &serverpb.TenantRangesRequest{})
			return err
		}); requestErr != nil {
			__antithesis_instrumentation__.Notify(35327)
			if err := zc.z.createError(s, tenantRangesName, requestErr); err != nil {
				__antithesis_instrumentation__.Notify(35328)
				return nodesInfo{}, nil, errors.Wrap(err, "fetching tenant ranges")
			} else {
				__antithesis_instrumentation__.Notify(35329)
			}
		} else {
			__antithesis_instrumentation__.Notify(35330)
			s.done()
			rangesFound := 0
			for locality, rangeList := range tenantRanges.RangesByLocality {
				__antithesis_instrumentation__.Notify(35332)
				rangesFound += len(rangeList.Ranges)
				sort.Slice(rangeList.Ranges, func(i, j int) bool {
					__antithesis_instrumentation__.Notify(35335)
					return rangeList.Ranges[i].RangeID > rangeList.Ranges[j].RangeID
				})
				__antithesis_instrumentation__.Notify(35333)
				sLocality := zc.clusterPrinter.start("writing tenant ranges for locality: %s", locality)
				prefix := fmt.Sprintf("%s/%s", tenantRangesName, locality)
				for _, r := range rangeList.Ranges {
					__antithesis_instrumentation__.Notify(35336)
					sRange := zc.clusterPrinter.start("writing tenant range %d", r.RangeID)
					name := fmt.Sprintf("%s/%d", prefix, r.RangeID)
					if err := zc.z.createJSON(sRange, name+".json", r); err != nil {
						__antithesis_instrumentation__.Notify(35337)
						return nodesInfo{}, nil, errors.Wrapf(err, "writing tenant range %d for locality %s", r.RangeID, locality)
					} else {
						__antithesis_instrumentation__.Notify(35338)
					}
				}
				__antithesis_instrumentation__.Notify(35334)
				sLocality.done()
			}
			__antithesis_instrumentation__.Notify(35331)
			zc.clusterPrinter.info("%d tenant ranges found", rangesFound)
		}
	}
	__antithesis_instrumentation__.Notify(35295)

	return ni, livenessByNodeID, nil
}

func (zc *debugZipContext) nodesInfo(ctx context.Context) (ni nodesInfo, _ error) {
	__antithesis_instrumentation__.Notify(35339)
	nodesResponse, err := zc.status.Nodes(ctx, &serverpb.NodesRequest{})
	nodesList := &serverpb.NodesListResponse{}
	if code := status.Code(errors.Cause(err)); code == codes.Unimplemented {
		__antithesis_instrumentation__.Notify(35343)

		nodesList, err = zc.status.NodesList(ctx, &serverpb.NodesListRequest{})
	} else {
		__antithesis_instrumentation__.Notify(35344)
	}
	__antithesis_instrumentation__.Notify(35340)
	if err != nil {
		__antithesis_instrumentation__.Notify(35345)
		return nodesInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(35346)
	}
	__antithesis_instrumentation__.Notify(35341)
	if nodesResponse != nil {
		__antithesis_instrumentation__.Notify(35347)

		for _, node := range nodesResponse.Nodes {
			__antithesis_instrumentation__.Notify(35348)
			nodeDetails := serverpb.NodeDetails{
				NodeID:     int32(node.Desc.NodeID),
				Address:    node.Desc.Address,
				SQLAddress: node.Desc.SQLAddress,
			}
			nodesList.Nodes = append(nodesList.Nodes, nodeDetails)
		}
	} else {
		__antithesis_instrumentation__.Notify(35349)
	}
	__antithesis_instrumentation__.Notify(35342)
	ni.nodesListResponse = nodesList
	ni.nodesStatusResponse = nodesResponse

	return ni, nil
}
