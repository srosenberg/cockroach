package serverpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	context "context"

	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
)

type SQLStatusServer interface {
	ListSessions(context.Context, *ListSessionsRequest) (*ListSessionsResponse, error)
	ListLocalSessions(context.Context, *ListSessionsRequest) (*ListSessionsResponse, error)
	CancelQuery(context.Context, *CancelQueryRequest) (*CancelQueryResponse, error)
	CancelQueryByKey(context.Context, *CancelQueryByKeyRequest) (*CancelQueryByKeyResponse, error)
	CancelSession(context.Context, *CancelSessionRequest) (*CancelSessionResponse, error)
	ListContentionEvents(context.Context, *ListContentionEventsRequest) (*ListContentionEventsResponse, error)
	ListLocalContentionEvents(context.Context, *ListContentionEventsRequest) (*ListContentionEventsResponse, error)
	ResetSQLStats(context.Context, *ResetSQLStatsRequest) (*ResetSQLStatsResponse, error)
	CombinedStatementStats(context.Context, *CombinedStatementsStatsRequest) (*StatementsResponse, error)
	Statements(context.Context, *StatementsRequest) (*StatementsResponse, error)
	StatementDetails(context.Context, *StatementDetailsRequest) (*StatementDetailsResponse, error)
	ListDistSQLFlows(context.Context, *ListDistSQLFlowsRequest) (*ListDistSQLFlowsResponse, error)
	ListLocalDistSQLFlows(context.Context, *ListDistSQLFlowsRequest) (*ListDistSQLFlowsResponse, error)
	Profile(context.Context, *ProfileRequest) (*JSONResponse, error)
	IndexUsageStatistics(context.Context, *IndexUsageStatisticsRequest) (*IndexUsageStatisticsResponse, error)
	ResetIndexUsageStats(context.Context, *ResetIndexUsageStatsRequest) (*ResetIndexUsageStatsResponse, error)
	TableIndexStats(context.Context, *TableIndexStatsRequest) (*TableIndexStatsResponse, error)
	UserSQLRoles(context.Context, *UserSQLRolesRequest) (*UserSQLRolesResponse, error)
	TxnIDResolution(context.Context, *TxnIDResolutionRequest) (*TxnIDResolutionResponse, error)
	TransactionContentionEvents(context.Context, *TransactionContentionEventsRequest) (*TransactionContentionEventsResponse, error)
	NodesList(context.Context, *NodesListRequest) (*NodesListResponse, error)
}

type OptionalNodesStatusServer struct {
	w errorutil.TenantSQLDeprecatedWrapper
}

func MakeOptionalNodesStatusServer(s NodesStatusServer) OptionalNodesStatusServer {
	__antithesis_instrumentation__.Notify(211262)
	return OptionalNodesStatusServer{

		w: errorutil.MakeTenantSQLDeprecatedWrapper(s, s != nil),
	}
}

type NodesStatusServer interface {
	ListNodesInternal(context.Context, *NodesRequest) (*NodesResponse, error)
}

type RegionsServer interface {
	Regions(context.Context, *RegionsRequest) (*RegionsResponse, error)
}

type TenantStatusServer interface {
	TenantRanges(context.Context, *TenantRangesRequest) (*TenantRangesResponse, error)
}

func (s *OptionalNodesStatusServer) OptionalNodesStatusServer(
	issue int,
) (NodesStatusServer, error) {
	__antithesis_instrumentation__.Notify(211263)
	v, err := s.w.OptionalErr(issue)
	if err != nil {
		__antithesis_instrumentation__.Notify(211265)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(211266)
	}
	__antithesis_instrumentation__.Notify(211264)
	return v.(NodesStatusServer), nil
}
