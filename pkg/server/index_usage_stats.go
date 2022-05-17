package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/idxusage"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *statusServer) IndexUsageStatistics(
	ctx context.Context, req *serverpb.IndexUsageStatisticsRequest,
) (*serverpb.IndexUsageStatisticsResponse, error) {
	__antithesis_instrumentation__.Notify(193752)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(193760)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193761)
	}
	__antithesis_instrumentation__.Notify(193753)

	localReq := &serverpb.IndexUsageStatisticsRequest{
		NodeID: "local",
	}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(193762)
		requestedNodeID, local, err := s.parseNodeID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(193766)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(193767)
		}
		__antithesis_instrumentation__.Notify(193763)
		if local {
			__antithesis_instrumentation__.Notify(193768)
			statsReader := s.sqlServer.pgServer.SQLServer.GetLocalIndexStatistics()
			return indexUsageStatsLocal(statsReader)
		} else {
			__antithesis_instrumentation__.Notify(193769)
		}
		__antithesis_instrumentation__.Notify(193764)

		statusClient, err := s.dialNode(ctx, requestedNodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(193770)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(193771)
		}
		__antithesis_instrumentation__.Notify(193765)

		return statusClient.IndexUsageStatistics(ctx, localReq)
	} else {
		__antithesis_instrumentation__.Notify(193772)
	}
	__antithesis_instrumentation__.Notify(193754)

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(193773)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(193755)

	fetchIndexUsageStats := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(193774)
		statusClient := client.(serverpb.StatusClient)
		return statusClient.IndexUsageStatistics(ctx, localReq)
	}
	__antithesis_instrumentation__.Notify(193756)

	resp := &serverpb.IndexUsageStatisticsResponse{}
	aggFn := func(_ roachpb.NodeID, nodeResp interface{}) {
		__antithesis_instrumentation__.Notify(193775)
		stats := nodeResp.(*serverpb.IndexUsageStatisticsResponse)
		resp.Statistics = append(resp.Statistics, stats.Statistics...)
	}
	__antithesis_instrumentation__.Notify(193757)

	var combinedError error
	errFn := func(_ roachpb.NodeID, nodeFnError error) {
		__antithesis_instrumentation__.Notify(193776)
		combinedError = errors.CombineErrors(combinedError, nodeFnError)
	}
	__antithesis_instrumentation__.Notify(193758)

	if err := s.iterateNodes(ctx,
		"requesting index usage stats",
		dialFn, fetchIndexUsageStats, aggFn, errFn); err != nil {
		__antithesis_instrumentation__.Notify(193777)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193778)
	}
	__antithesis_instrumentation__.Notify(193759)

	resp.LastReset = s.sqlServer.pgServer.SQLServer.GetLocalIndexStatistics().GetLastReset()

	return resp, nil
}

func indexUsageStatsLocal(
	idxUsageStats *idxusage.LocalIndexUsageStats,
) (*serverpb.IndexUsageStatisticsResponse, error) {
	__antithesis_instrumentation__.Notify(193779)
	resp := &serverpb.IndexUsageStatisticsResponse{}
	if err := idxUsageStats.ForEach(idxusage.IteratorOptions{}, func(key *roachpb.IndexUsageKey, value *roachpb.IndexUsageStatistics) error {
		__antithesis_instrumentation__.Notify(193781)
		resp.Statistics = append(resp.Statistics, roachpb.CollectedIndexUsageStatistics{Key: *key,
			Stats: *value,
		})
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(193782)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193783)
	}
	__antithesis_instrumentation__.Notify(193780)

	resp.LastReset = idxUsageStats.GetLastReset()
	return resp, nil
}

func (s *statusServer) ResetIndexUsageStats(
	ctx context.Context, req *serverpb.ResetIndexUsageStatsRequest,
) (*serverpb.ResetIndexUsageStatsResponse, error) {
	__antithesis_instrumentation__.Notify(193784)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(193792)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193793)
	}
	__antithesis_instrumentation__.Notify(193785)

	localReq := &serverpb.ResetIndexUsageStatsRequest{
		NodeID: "local",
	}
	resp := &serverpb.ResetIndexUsageStatsResponse{}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(193794)
		requestedNodeID, local, err := s.parseNodeID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(193798)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(193799)
		}
		__antithesis_instrumentation__.Notify(193795)
		if local {
			__antithesis_instrumentation__.Notify(193800)
			s.sqlServer.pgServer.SQLServer.GetLocalIndexStatistics().Reset()
			return resp, nil
		} else {
			__antithesis_instrumentation__.Notify(193801)
		}
		__antithesis_instrumentation__.Notify(193796)

		statusClient, err := s.dialNode(ctx, requestedNodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(193802)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(193803)
		}
		__antithesis_instrumentation__.Notify(193797)

		return statusClient.ResetIndexUsageStats(ctx, localReq)
	} else {
		__antithesis_instrumentation__.Notify(193804)
	}
	__antithesis_instrumentation__.Notify(193786)

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(193805)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(193787)

	resetIndexUsageStats := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(193806)
		statusClient := client.(serverpb.StatusClient)
		return statusClient.ResetIndexUsageStats(ctx, localReq)
	}
	__antithesis_instrumentation__.Notify(193788)

	aggFn := func(_ roachpb.NodeID, nodeResp interface{}) {
		__antithesis_instrumentation__.Notify(193807)

	}
	__antithesis_instrumentation__.Notify(193789)

	var combinedError error
	errFn := func(_ roachpb.NodeID, nodeFnError error) {
		__antithesis_instrumentation__.Notify(193808)
		combinedError = errors.CombineErrors(combinedError, nodeFnError)
	}
	__antithesis_instrumentation__.Notify(193790)

	if err := s.iterateNodes(ctx,
		"Resetting index usage stats",
		dialFn, resetIndexUsageStats, aggFn, errFn); err != nil {
		__antithesis_instrumentation__.Notify(193809)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193810)
	}
	__antithesis_instrumentation__.Notify(193791)

	return resp, nil
}

func (s *statusServer) TableIndexStats(
	ctx context.Context, req *serverpb.TableIndexStatsRequest,
) (*serverpb.TableIndexStatsResponse, error) {
	__antithesis_instrumentation__.Notify(193811)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(193813)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193814)
	}
	__antithesis_instrumentation__.Notify(193812)
	return getTableIndexUsageStats(ctx, req, s.sqlServer.pgServer.SQLServer.GetLocalIndexStatistics(),
		s.sqlServer.internalExecutor)
}

func getTableIndexUsageStats(
	ctx context.Context,
	req *serverpb.TableIndexStatsRequest,
	idxUsageStatsProvider *idxusage.LocalIndexUsageStats,
	ie *sql.InternalExecutor,
) (*serverpb.TableIndexStatsResponse, error) {
	__antithesis_instrumentation__.Notify(193815)
	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(193821)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193822)
	}
	__antithesis_instrumentation__.Notify(193816)

	tableID, err := getTableIDFromDatabaseAndTableName(ctx, req.Database, req.Table, ie, userName)

	if err != nil {
		__antithesis_instrumentation__.Notify(193823)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193824)
	}
	__antithesis_instrumentation__.Notify(193817)

	q := makeSQLQuery()

	q.Append(`
		SELECT
			ti.index_id,
			ti.index_name,
			ti.index_type,
			total_reads,
			last_read,
			indexdef,
			ti.created_at
		FROM crdb_internal.index_usage_statistics AS us
  	JOIN crdb_internal.table_indexes AS ti ON us.index_id = ti.index_id 
		AND us.table_id = ti.descriptor_id
  	JOIN pg_catalog.pg_index AS pgidx ON indrelid = us.table_id
  	JOIN pg_catalog.pg_indexes AS pgidxs ON pgidxs.crdb_oid = indexrelid
		AND indexname = ti.index_name
 		WHERE ti.descriptor_id = $::REGCLASS`,
		tableID,
	)

	const expectedNumDatums = 7

	it, err := ie.QueryIteratorEx(ctx, "index-usage-stats", nil,
		sessiondata.InternalExecutorOverride{
			User:     userName,
			Database: req.Database,
		}, q.String(), q.QueryArguments()...)

	if err != nil {
		__antithesis_instrumentation__.Notify(193825)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(193826)
	}
	__antithesis_instrumentation__.Notify(193818)

	var idxUsageStats []*serverpb.TableIndexStatsResponse_ExtendedCollectedIndexUsageStatistics
	var ok bool

	defer func() { __antithesis_instrumentation__.Notify(193827); err = errors.CombineErrors(err, it.Close()) }()
	__antithesis_instrumentation__.Notify(193819)

	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(193828)
		var row tree.Datums
		if row = it.Cur(); row == nil {
			__antithesis_instrumentation__.Notify(193834)
			return nil, errors.New("unexpected null row")
		} else {
			__antithesis_instrumentation__.Notify(193835)
		}
		__antithesis_instrumentation__.Notify(193829)

		if row.Len() != expectedNumDatums {
			__antithesis_instrumentation__.Notify(193836)
			return nil, errors.Newf("expected %d columns, received %d", expectedNumDatums, row.Len())
		} else {
			__antithesis_instrumentation__.Notify(193837)
		}
		__antithesis_instrumentation__.Notify(193830)

		indexID := tree.MustBeDInt(row[0])
		indexName := tree.MustBeDString(row[1])
		indexType := tree.MustBeDString(row[2])
		totalReads := uint64(tree.MustBeDInt(row[3]))
		lastRead := time.Time{}
		if row[4] != tree.DNull {
			__antithesis_instrumentation__.Notify(193838)
			lastRead = tree.MustBeDTimestampTZ(row[4]).Time
		} else {
			__antithesis_instrumentation__.Notify(193839)
		}
		__antithesis_instrumentation__.Notify(193831)
		createStmt := tree.MustBeDString(row[5])
		var createdAt *time.Time
		if row[6] != tree.DNull {
			__antithesis_instrumentation__.Notify(193840)
			ts := tree.MustBeDTimestamp(row[6])
			createdAt = &ts.Time
		} else {
			__antithesis_instrumentation__.Notify(193841)
		}
		__antithesis_instrumentation__.Notify(193832)

		if err != nil {
			__antithesis_instrumentation__.Notify(193842)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(193843)
		}
		__antithesis_instrumentation__.Notify(193833)

		idxStatsRow := &serverpb.TableIndexStatsResponse_ExtendedCollectedIndexUsageStatistics{
			Statistics: &roachpb.CollectedIndexUsageStatistics{
				Key: roachpb.IndexUsageKey{
					TableID: roachpb.TableID(tableID),
					IndexID: roachpb.IndexID(indexID),
				},
				Stats: roachpb.IndexUsageStatistics{
					TotalReadCount: totalReads,
					LastRead:       lastRead,
				},
			},
			IndexName:       string(indexName),
			IndexType:       string(indexType),
			CreateStatement: string(createStmt),
			CreatedAt:       createdAt,
		}

		idxUsageStats = append(idxUsageStats, idxStatsRow)
	}
	__antithesis_instrumentation__.Notify(193820)

	lastReset := idxUsageStatsProvider.GetLastReset()

	resp := &serverpb.TableIndexStatsResponse{
		Statistics: idxUsageStats,
		LastReset:  &lastReset,
	}

	return resp, nil
}

func getTableIDFromDatabaseAndTableName(
	ctx context.Context,
	database string,
	table string,
	ie *sql.InternalExecutor,
	userName security.SQLUsername,
) (int, error) {
	__antithesis_instrumentation__.Notify(193844)

	fqtName, err := getFullyQualifiedTableName(database, table)
	if err != nil {
		__antithesis_instrumentation__.Notify(193850)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(193851)
	}
	__antithesis_instrumentation__.Notify(193845)
	names := strings.Split(fqtName, ".")

	q := makeSQLQuery()
	q.Append(`SELECT table_id FROM crdb_internal.tables WHERE database_name = $ `, names[0])

	if len(names) == 2 {
		__antithesis_instrumentation__.Notify(193852)
		q.Append(`AND name = $`, names[1])
	} else {
		__antithesis_instrumentation__.Notify(193853)
		if len(names) == 3 {
			__antithesis_instrumentation__.Notify(193854)
			q.Append(`AND schema_name = $ AND name = $`, names[1], names[2])
		} else {
			__antithesis_instrumentation__.Notify(193855)
			return 0, errors.Newf("expected array length 2 or 3, received %d", len(names))
		}
	}
	__antithesis_instrumentation__.Notify(193846)
	if len(q.Errors()) > 0 {
		__antithesis_instrumentation__.Notify(193856)
		return 0, combineAllErrors(q.Errors())
	} else {
		__antithesis_instrumentation__.Notify(193857)
	}
	__antithesis_instrumentation__.Notify(193847)

	datums, err := ie.QueryRowEx(ctx, "get-table-id", nil,
		sessiondata.InternalExecutorOverride{
			User:     userName,
			Database: database,
		}, q.String(), q.QueryArguments()...)

	if err != nil {
		__antithesis_instrumentation__.Notify(193858)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(193859)
	}
	__antithesis_instrumentation__.Notify(193848)
	if datums == nil {
		__antithesis_instrumentation__.Notify(193860)
		return 0, errors.Newf("expected to find table ID for table %s, but found nothing", fqtName)
	} else {
		__antithesis_instrumentation__.Notify(193861)
	}
	__antithesis_instrumentation__.Notify(193849)

	tableID := int(tree.MustBeDInt(datums[0]))
	return tableID, nil
}
