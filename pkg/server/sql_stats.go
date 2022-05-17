package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *statusServer) ResetSQLStats(
	ctx context.Context, req *serverpb.ResetSQLStatsRequest,
) (*serverpb.ResetSQLStatsResponse, error) {
	__antithesis_instrumentation__.Notify(235228)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(235235)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235236)
	}
	__antithesis_instrumentation__.Notify(235229)

	response := &serverpb.ResetSQLStatsResponse{}
	controller := s.sqlServer.pgServer.SQLServer.GetSQLStatsController()

	if req.ResetPersistedStats {
		__antithesis_instrumentation__.Notify(235237)
		if err := controller.ResetClusterSQLStats(ctx); err != nil {
			__antithesis_instrumentation__.Notify(235239)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(235240)
		}
		__antithesis_instrumentation__.Notify(235238)

		return response, nil
	} else {
		__antithesis_instrumentation__.Notify(235241)
	}
	__antithesis_instrumentation__.Notify(235230)

	localReq := &serverpb.ResetSQLStatsRequest{
		NodeID: "local",

		ResetPersistedStats: false,
	}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(235242)
		requestedNodeID, local, err := s.parseNodeID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(235246)
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(235247)
		}
		__antithesis_instrumentation__.Notify(235243)
		if local {
			__antithesis_instrumentation__.Notify(235248)
			controller.ResetLocalSQLStats(ctx)
			return response, nil
		} else {
			__antithesis_instrumentation__.Notify(235249)
		}
		__antithesis_instrumentation__.Notify(235244)
		status, err := s.dialNode(ctx, requestedNodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(235250)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(235251)
		}
		__antithesis_instrumentation__.Notify(235245)
		return status.ResetSQLStats(ctx, localReq)
	} else {
		__antithesis_instrumentation__.Notify(235252)
	}
	__antithesis_instrumentation__.Notify(235231)

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(235253)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(235232)

	resetSQLStats := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(235254)
		status := client.(serverpb.StatusClient)
		return status.ResetSQLStats(ctx, localReq)
	}
	__antithesis_instrumentation__.Notify(235233)

	var fanoutError error

	if err := s.iterateNodes(ctx, "reset SQL statistics",
		dialFn,
		resetSQLStats,
		func(nodeID roachpb.NodeID, resp interface{}) {
			__antithesis_instrumentation__.Notify(235255)

		},
		func(nodeID roachpb.NodeID, nodeFnError error) {
			__antithesis_instrumentation__.Notify(235256)
			if nodeFnError != nil {
				__antithesis_instrumentation__.Notify(235257)
				fanoutError = errors.CombineErrors(fanoutError, nodeFnError)
			} else {
				__antithesis_instrumentation__.Notify(235258)
			}
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(235259)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235260)
	}
	__antithesis_instrumentation__.Notify(235234)

	return response, fanoutError
}
