package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *statusServer) Statements(
	ctx context.Context, req *serverpb.StatementsRequest,
) (*serverpb.StatementsResponse, error) {
	__antithesis_instrumentation__.Notify(235376)
	if req.Combined {
		__antithesis_instrumentation__.Notify(235384)
		combinedRequest := serverpb.CombinedStatementsStatsRequest{
			Start: req.Start,
			End:   req.End,
		}
		return s.CombinedStatementStats(ctx, &combinedRequest)
	} else {
		__antithesis_instrumentation__.Notify(235385)
	}
	__antithesis_instrumentation__.Notify(235377)

	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(235386)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235387)
	}
	__antithesis_instrumentation__.Notify(235378)

	if s.gossip.NodeID.Get() == 0 {
		__antithesis_instrumentation__.Notify(235388)
		return nil, status.Errorf(codes.Unavailable, "nodeID not set")
	} else {
		__antithesis_instrumentation__.Notify(235389)
	}
	__antithesis_instrumentation__.Notify(235379)

	response := &serverpb.StatementsResponse{
		Statements:            []serverpb.StatementsResponse_CollectedStatementStatistics{},
		LastReset:             timeutil.Now(),
		InternalAppNamePrefix: catconstants.InternalAppNamePrefix,
	}

	localReq := &serverpb.StatementsRequest{
		NodeID:    "local",
		FetchMode: req.FetchMode,
	}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(235390)
		requestedNodeID, local, err := s.parseNodeID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(235394)
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(235395)
		}
		__antithesis_instrumentation__.Notify(235391)
		if local {
			__antithesis_instrumentation__.Notify(235396)
			return statementsLocal(
				ctx,
				s.gossip.NodeID.Get(),
				s.admin.server.sqlServer,
				req.FetchMode)
		} else {
			__antithesis_instrumentation__.Notify(235397)
		}
		__antithesis_instrumentation__.Notify(235392)
		status, err := s.dialNode(ctx, requestedNodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(235398)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(235399)
		}
		__antithesis_instrumentation__.Notify(235393)
		return status.Statements(ctx, localReq)
	} else {
		__antithesis_instrumentation__.Notify(235400)
	}
	__antithesis_instrumentation__.Notify(235380)

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(235401)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(235381)
	nodeStatement := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(235402)
		status := client.(serverpb.StatusClient)
		return status.Statements(ctx, localReq)
	}
	__antithesis_instrumentation__.Notify(235382)

	if err := s.iterateNodes(ctx, "statement statistics",
		dialFn,
		nodeStatement,
		func(nodeID roachpb.NodeID, resp interface{}) {
			__antithesis_instrumentation__.Notify(235403)
			statementsResp := resp.(*serverpb.StatementsResponse)
			response.Statements = append(response.Statements, statementsResp.Statements...)
			response.Transactions = append(response.Transactions, statementsResp.Transactions...)
			if response.LastReset.After(statementsResp.LastReset) {
				__antithesis_instrumentation__.Notify(235404)
				response.LastReset = statementsResp.LastReset
			} else {
				__antithesis_instrumentation__.Notify(235405)
			}
		},
		func(nodeID roachpb.NodeID, err error) {
			__antithesis_instrumentation__.Notify(235406)

		},
	); err != nil {
		__antithesis_instrumentation__.Notify(235407)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235408)
	}
	__antithesis_instrumentation__.Notify(235383)

	return response, nil
}

func statementsLocal(
	ctx context.Context,
	nodeID roachpb.NodeID,
	sqlServer *SQLServer,
	fetchMode serverpb.StatementsRequest_FetchMode,
) (*serverpb.StatementsResponse, error) {
	__antithesis_instrumentation__.Notify(235409)
	var stmtStats []roachpb.CollectedStatementStatistics
	var txnStats []roachpb.CollectedTransactionStatistics
	var err error

	if fetchMode != serverpb.StatementsRequest_TxnStatsOnly {
		__antithesis_instrumentation__.Notify(235414)
		stmtStats, err = sqlServer.pgServer.SQLServer.GetUnscrubbedStmtStats(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(235415)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(235416)
		}
	} else {
		__antithesis_instrumentation__.Notify(235417)
	}
	__antithesis_instrumentation__.Notify(235410)

	if fetchMode != serverpb.StatementsRequest_StmtStatsOnly {
		__antithesis_instrumentation__.Notify(235418)
		txnStats, err = sqlServer.pgServer.SQLServer.GetUnscrubbedTxnStats(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(235419)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(235420)
		}
	} else {
		__antithesis_instrumentation__.Notify(235421)
	}
	__antithesis_instrumentation__.Notify(235411)

	lastReset := sqlServer.pgServer.SQLServer.GetStmtStatsLastReset()

	resp := &serverpb.StatementsResponse{
		Statements:            make([]serverpb.StatementsResponse_CollectedStatementStatistics, len(stmtStats)),
		LastReset:             lastReset,
		InternalAppNamePrefix: catconstants.InternalAppNamePrefix,
		Transactions:          make([]serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics, len(txnStats)),
	}

	for i, txn := range txnStats {
		__antithesis_instrumentation__.Notify(235422)
		resp.Transactions[i] = serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics{
			StatsData: txn,
			NodeID:    nodeID,
		}
	}
	__antithesis_instrumentation__.Notify(235412)

	for i, stmt := range stmtStats {
		__antithesis_instrumentation__.Notify(235423)
		resp.Statements[i] = serverpb.StatementsResponse_CollectedStatementStatistics{
			Key: serverpb.StatementsResponse_ExtendedStatementStatisticsKey{
				KeyData: stmt.Key,
				NodeID:  nodeID,
			},
			ID:    stmt.ID,
			Stats: stmt.Stats,
		}
	}
	__antithesis_instrumentation__.Notify(235413)

	return resp, nil
}
