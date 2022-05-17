package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/contention"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirecancel"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tenantStatusServer struct {
	baseStatusServer
}

var _ serverpb.StatusServer = &tenantStatusServer{}

func (t *tenantStatusServer) RegisterService(g *grpc.Server) {
	__antithesis_instrumentation__.Notify(238416)
	serverpb.RegisterStatusServer(g, t)
}

func (t *tenantStatusServer) RegisterGateway(
	ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn,
) error {
	__antithesis_instrumentation__.Notify(238417)
	ctx = t.AnnotateCtx(ctx)
	return serverpb.RegisterStatusHandler(ctx, mux, conn)
}

var _ grpcGatewayServer = &tenantStatusServer{}

func newTenantStatusServer(
	ambient log.AmbientContext,
	privilegeChecker *adminPrivilegeChecker,
	sessionRegistry *sql.SessionRegistry,
	flowScheduler *flowinfra.FlowScheduler,
	st *cluster.Settings,
	sqlServer *SQLServer,
	rpcCtx *rpc.Context,
	stopper *stop.Stopper,
) *tenantStatusServer {
	__antithesis_instrumentation__.Notify(238418)
	ambient.AddLogTag("tenant-status", nil)
	return &tenantStatusServer{
		baseStatusServer: baseStatusServer{
			AmbientContext:   ambient,
			privilegeChecker: privilegeChecker,
			sessionRegistry:  sessionRegistry,
			flowScheduler:    flowScheduler,
			st:               st,
			sqlServer:        sqlServer,
			rpcCtx:           rpcCtx,
			stopper:          stopper,
		},
	}
}

func (t *tenantStatusServer) dialCallback(
	ctx context.Context, instanceID base.SQLInstanceID, addr string,
) (interface{}, error) {
	__antithesis_instrumentation__.Notify(238419)
	client, err := t.dialPod(ctx, instanceID, addr)
	return client, err
}

func (t *tenantStatusServer) ListSessions(
	ctx context.Context, req *serverpb.ListSessionsRequest,
) (*serverpb.ListSessionsResponse, error) {
	__antithesis_instrumentation__.Notify(238420)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238425)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238426)
	}
	__antithesis_instrumentation__.Notify(238421)
	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238427)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238428)
	}
	__antithesis_instrumentation__.Notify(238422)

	response := &serverpb.ListSessionsResponse{
		InternalAppNamePrefix: catconstants.InternalAppNamePrefix,
	}
	nodeStatement := func(ctx context.Context, client interface{}, instanceID base.SQLInstanceID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(238429)
		statusClient := client.(serverpb.StatusClient)
		localResponse, err := statusClient.ListLocalSessions(ctx, req)
		if localResponse == nil {
			__antithesis_instrumentation__.Notify(238432)
			log.Errorf(ctx, "listing local sessions on %d produced a nil result with error %v",
				instanceID,
				err)
		} else {
			__antithesis_instrumentation__.Notify(238433)
		}
		__antithesis_instrumentation__.Notify(238430)
		if err != nil {
			__antithesis_instrumentation__.Notify(238434)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(238435)
		}
		__antithesis_instrumentation__.Notify(238431)
		return localResponse, nil
	}
	__antithesis_instrumentation__.Notify(238423)
	if err := t.iteratePods(ctx, "sessions for nodes",
		t.dialCallback,
		nodeStatement,
		func(instanceID base.SQLInstanceID, resp interface{}) {
			__antithesis_instrumentation__.Notify(238436)
			sessionResp := resp.(*serverpb.ListSessionsResponse)
			response.Sessions = append(response.Sessions, sessionResp.Sessions...)
			response.Errors = append(response.Errors, sessionResp.Errors...)
		},
		func(instanceID base.SQLInstanceID, err error) {
			__antithesis_instrumentation__.Notify(238437)

			log.Warningf(ctx, "fan out statements request recorded error from node %d: %v", instanceID, err)
			response.Errors = append(response.Errors,
				serverpb.ListSessionsError{
					Message: err.Error(),
					NodeID:  roachpb.NodeID(instanceID),
				})
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(238438)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(238439)
	}
	__antithesis_instrumentation__.Notify(238424)
	return response, nil
}

func (t *tenantStatusServer) ListLocalSessions(
	ctx context.Context, request *serverpb.ListSessionsRequest,
) (*serverpb.ListSessionsResponse, error) {
	__antithesis_instrumentation__.Notify(238440)
	sessions, err := t.getLocalSessions(ctx, request)
	if err != nil {
		__antithesis_instrumentation__.Notify(238443)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238444)
	}
	__antithesis_instrumentation__.Notify(238441)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238445)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238446)
	}
	__antithesis_instrumentation__.Notify(238442)

	return &serverpb.ListSessionsResponse{
		Sessions:              sessions,
		InternalAppNamePrefix: catconstants.InternalAppNamePrefix,
	}, nil
}

func (t *tenantStatusServer) CancelQuery(
	ctx context.Context, req *serverpb.CancelQueryRequest,
) (*serverpb.CancelQueryResponse, error) {
	__antithesis_instrumentation__.Notify(238447)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	queryID, err := sql.StringToClusterWideID(req.QueryID)
	if err != nil {
		__antithesis_instrumentation__.Notify(238452)
		return &serverpb.CancelQueryResponse{
			Canceled: false,
			Error:    errors.Wrapf(err, "query ID %s malformed", queryID).Error(),
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(238453)
	}
	__antithesis_instrumentation__.Notify(238448)
	instanceID := base.SQLInstanceID(queryID.GetNodeID())

	if instanceID != t.sqlServer.SQLInstanceID() {
		__antithesis_instrumentation__.Notify(238454)
		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, instanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238457)
			if errors.Is(err, sqlinstance.NonExistentInstanceError) {
				__antithesis_instrumentation__.Notify(238459)
				return &serverpb.CancelQueryResponse{
					Canceled: false,
					Error:    fmt.Sprintf("query ID %s not found", queryID),
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(238460)
			}
			__antithesis_instrumentation__.Notify(238458)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(238461)
		}
		__antithesis_instrumentation__.Notify(238455)
		statusClient, err := t.dialPod(ctx, instanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238462)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(238463)
		}
		__antithesis_instrumentation__.Notify(238456)
		return statusClient.CancelQuery(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(238464)
	}
	__antithesis_instrumentation__.Notify(238449)

	reqUsername := security.MakeSQLUsernameFromPreNormalizedString(req.Username)
	if err := t.checkCancelPrivilege(ctx, reqUsername, findSessionByQueryID(req.QueryID)); err != nil {
		__antithesis_instrumentation__.Notify(238465)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238466)
	}
	__antithesis_instrumentation__.Notify(238450)
	resp := &serverpb.CancelQueryResponse{}
	resp.Canceled, err = t.sessionRegistry.CancelQuery(req.QueryID)
	if err != nil {
		__antithesis_instrumentation__.Notify(238467)
		resp.Error = err.Error()
	} else {
		__antithesis_instrumentation__.Notify(238468)
	}
	__antithesis_instrumentation__.Notify(238451)
	return resp, nil

}

func (t *tenantStatusServer) CancelQueryByKey(
	ctx context.Context, req *serverpb.CancelQueryByKeyRequest,
) (resp *serverpb.CancelQueryByKeyResponse, retErr error) {
	__antithesis_instrumentation__.Notify(238469)

	local := req.SQLInstanceID == t.sqlServer.SQLInstanceID()
	resp = &serverpb.CancelQueryByKeyResponse{}

	alloc, err := pgwirecancel.CancelSemaphore.TryAcquire(ctx, 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(238475)
		return nil, status.Errorf(codes.ResourceExhausted, "exceeded rate limit of pgwire cancellation requests")
	} else {
		__antithesis_instrumentation__.Notify(238476)
	}
	__antithesis_instrumentation__.Notify(238470)
	defer func() {
		__antithesis_instrumentation__.Notify(238477)

		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(238479)
			return (resp != nil && func() bool {
				__antithesis_instrumentation__.Notify(238480)
				return !resp.Canceled == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(238481)
			time.Sleep(1 * time.Second)
		} else {
			__antithesis_instrumentation__.Notify(238482)
		}
		__antithesis_instrumentation__.Notify(238478)
		alloc.Release()
	}()
	__antithesis_instrumentation__.Notify(238471)

	if local {
		__antithesis_instrumentation__.Notify(238483)
		resp.Canceled, err = t.sessionRegistry.CancelQueryByKey(req.CancelQueryKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(238485)
			resp.Error = err.Error()
		} else {
			__antithesis_instrumentation__.Notify(238486)
		}
		__antithesis_instrumentation__.Notify(238484)
		return resp, nil
	} else {
		__antithesis_instrumentation__.Notify(238487)
	}
	__antithesis_instrumentation__.Notify(238472)

	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)
	instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, req.SQLInstanceID)
	if err != nil {
		__antithesis_instrumentation__.Notify(238488)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238489)
	}
	__antithesis_instrumentation__.Notify(238473)
	statusClient, err := t.dialPod(ctx, req.SQLInstanceID, instance.InstanceAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(238490)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238491)
	}
	__antithesis_instrumentation__.Notify(238474)
	return statusClient.CancelQueryByKey(ctx, req)
}

func (t *tenantStatusServer) CancelSession(
	ctx context.Context, req *serverpb.CancelSessionRequest,
) (*serverpb.CancelSessionResponse, error) {
	__antithesis_instrumentation__.Notify(238492)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	sessionID := sql.BytesToClusterWideID(req.SessionID)
	instanceID := base.SQLInstanceID(sessionID.GetNodeID())

	if instanceID != t.sqlServer.SQLInstanceID() {
		__antithesis_instrumentation__.Notify(238495)
		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, instanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238498)
			if errors.Is(err, sqlinstance.NonExistentInstanceError) {
				__antithesis_instrumentation__.Notify(238500)
				return &serverpb.CancelSessionResponse{
					Canceled: false,
					Error:    fmt.Sprintf("session ID %s not found", sessionID),
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(238501)
			}
			__antithesis_instrumentation__.Notify(238499)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(238502)
		}
		__antithesis_instrumentation__.Notify(238496)
		statusClient, err := t.dialPod(ctx, instanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238503)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(238504)
		}
		__antithesis_instrumentation__.Notify(238497)
		return statusClient.CancelSession(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(238505)
	}
	__antithesis_instrumentation__.Notify(238493)

	reqUsername := security.MakeSQLUsernameFromPreNormalizedString(req.Username)
	if err := t.checkCancelPrivilege(ctx, reqUsername, findSessionBySessionID(req.SessionID)); err != nil {
		__antithesis_instrumentation__.Notify(238506)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238507)
	}
	__antithesis_instrumentation__.Notify(238494)
	return t.sessionRegistry.CancelSession(req.SessionID)
}

func (t *tenantStatusServer) ListContentionEvents(
	ctx context.Context, req *serverpb.ListContentionEventsRequest,
) (*serverpb.ListContentionEventsResponse, error) {
	__antithesis_instrumentation__.Notify(238508)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238515)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238516)
	}
	__antithesis_instrumentation__.Notify(238509)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238517)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238518)
	}
	__antithesis_instrumentation__.Notify(238510)

	var response serverpb.ListContentionEventsResponse

	podFn := func(ctx context.Context, client interface{}, _ base.SQLInstanceID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(238519)
		statusClient := client.(serverpb.StatusClient)
		resp, err := statusClient.ListLocalContentionEvents(ctx, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(238522)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238523)
		}
		__antithesis_instrumentation__.Notify(238520)
		if len(resp.Errors) > 0 {
			__antithesis_instrumentation__.Notify(238524)
			return nil, errors.Errorf("%s", resp.Errors[0].Message)
		} else {
			__antithesis_instrumentation__.Notify(238525)
		}
		__antithesis_instrumentation__.Notify(238521)
		return resp, nil
	}
	__antithesis_instrumentation__.Notify(238511)
	responseFn := func(_ base.SQLInstanceID, nodeResp interface{}) {
		__antithesis_instrumentation__.Notify(238526)
		if nodeResp == nil {
			__antithesis_instrumentation__.Notify(238528)
			return
		} else {
			__antithesis_instrumentation__.Notify(238529)
		}
		__antithesis_instrumentation__.Notify(238527)
		events := nodeResp.(*serverpb.ListContentionEventsResponse).Events
		response.Events = contention.MergeSerializedRegistries(response.Events, events)
	}
	__antithesis_instrumentation__.Notify(238512)
	errorFn := func(instanceID base.SQLInstanceID, err error) {
		__antithesis_instrumentation__.Notify(238530)
		errResponse := serverpb.ListActivityError{
			NodeID:  roachpb.NodeID(instanceID),
			Message: err.Error(),
		}
		response.Errors = append(response.Errors, errResponse)
	}
	__antithesis_instrumentation__.Notify(238513)

	if err := t.iteratePods(
		ctx,
		"contention events list",
		t.dialCallback,
		podFn,
		responseFn,
		errorFn,
	); err != nil {
		__antithesis_instrumentation__.Notify(238531)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238532)
	}
	__antithesis_instrumentation__.Notify(238514)
	return &response, nil
}

func (t *tenantStatusServer) ListLocalContentionEvents(
	ctx context.Context, req *serverpb.ListContentionEventsRequest,
) (*serverpb.ListContentionEventsResponse, error) {
	__antithesis_instrumentation__.Notify(238533)
	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238535)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238536)
	}
	__antithesis_instrumentation__.Notify(238534)
	return t.baseStatusServer.ListLocalContentionEvents(ctx, req)
}

func (t *tenantStatusServer) ResetSQLStats(
	ctx context.Context, req *serverpb.ResetSQLStatsRequest,
) (*serverpb.ResetSQLStatsResponse, error) {
	__antithesis_instrumentation__.Notify(238537)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if _, err := t.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238544)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238545)
	}
	__antithesis_instrumentation__.Notify(238538)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238546)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238547)
	}
	__antithesis_instrumentation__.Notify(238539)

	response := &serverpb.ResetSQLStatsResponse{}
	controller := t.sqlServer.pgServer.SQLServer.GetSQLStatsController()

	if req.ResetPersistedStats {
		__antithesis_instrumentation__.Notify(238548)
		if err := controller.ResetClusterSQLStats(ctx); err != nil {
			__antithesis_instrumentation__.Notify(238550)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238551)
		}
		__antithesis_instrumentation__.Notify(238549)

		return response, nil
	} else {
		__antithesis_instrumentation__.Notify(238552)
	}
	__antithesis_instrumentation__.Notify(238540)

	localReq := &serverpb.ResetSQLStatsRequest{
		NodeID: "local",

		ResetPersistedStats: false,
	}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(238553)
		parsedInstanceID, local, err := t.parseInstanceID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238558)
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(238559)
		}
		__antithesis_instrumentation__.Notify(238554)
		if local {
			__antithesis_instrumentation__.Notify(238560)
			controller.ResetLocalSQLStats(ctx)
			return response, nil
		} else {
			__antithesis_instrumentation__.Notify(238561)
		}
		__antithesis_instrumentation__.Notify(238555)

		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, parsedInstanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238562)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238563)
		}
		__antithesis_instrumentation__.Notify(238556)
		statusClient, err := t.dialPod(ctx, parsedInstanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238564)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238565)
		}
		__antithesis_instrumentation__.Notify(238557)
		return statusClient.ResetSQLStats(ctx, localReq)
	} else {
		__antithesis_instrumentation__.Notify(238566)
	}
	__antithesis_instrumentation__.Notify(238541)

	nodeResetFn := func(
		ctx context.Context,
		client interface{},
		instanceID base.SQLInstanceID,
	) (interface{}, error) {
		__antithesis_instrumentation__.Notify(238567)
		statusClient := client.(serverpb.StatusClient)
		return statusClient.ResetSQLStats(ctx, localReq)
	}
	__antithesis_instrumentation__.Notify(238542)

	var fanoutError error

	if err := t.iteratePods(ctx, "reset SQL statistics",
		t.dialCallback,
		nodeResetFn,
		func(instanceID base.SQLInstanceID, resp interface{}) {
			__antithesis_instrumentation__.Notify(238568)

		},
		func(instanceID base.SQLInstanceID, err error) {
			__antithesis_instrumentation__.Notify(238569)
			if err != nil {
				__antithesis_instrumentation__.Notify(238570)
				fanoutError = errors.CombineErrors(fanoutError, err)
			} else {
				__antithesis_instrumentation__.Notify(238571)
			}
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(238572)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238573)
	}
	__antithesis_instrumentation__.Notify(238543)

	return response, fanoutError
}

func (t *tenantStatusServer) CombinedStatementStats(
	ctx context.Context, req *serverpb.CombinedStatementsStatsRequest,
) (*serverpb.StatementsResponse, error) {
	__antithesis_instrumentation__.Notify(238574)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238577)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238578)
	}
	__antithesis_instrumentation__.Notify(238575)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238579)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238580)
	}
	__antithesis_instrumentation__.Notify(238576)

	return getCombinedStatementStats(ctx, req, t.sqlServer.pgServer.SQLServer.GetSQLStatsProvider(),
		t.sqlServer.internalExecutor, t.st, t.sqlServer.execCfg.SQLStatsTestingKnobs)
}

func (t *tenantStatusServer) StatementDetails(
	ctx context.Context, req *serverpb.StatementDetailsRequest,
) (*serverpb.StatementDetailsResponse, error) {
	__antithesis_instrumentation__.Notify(238581)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238584)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238585)
	}
	__antithesis_instrumentation__.Notify(238582)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238586)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238587)
	}
	__antithesis_instrumentation__.Notify(238583)

	return getStatementDetails(ctx, req, t.sqlServer.internalExecutor, t.st, t.sqlServer.execCfg.SQLStatsTestingKnobs)
}

func (t *tenantStatusServer) Statements(
	ctx context.Context, req *serverpb.StatementsRequest,
) (*serverpb.StatementsResponse, error) {
	__antithesis_instrumentation__.Notify(238588)
	if req.Combined {
		__antithesis_instrumentation__.Notify(238595)
		combinedRequest := serverpb.CombinedStatementsStatsRequest{
			Start: req.Start,
			End:   req.End,
		}
		return t.CombinedStatementStats(ctx, &combinedRequest)
	} else {
		__antithesis_instrumentation__.Notify(238596)
	}
	__antithesis_instrumentation__.Notify(238589)

	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238597)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238598)
	}
	__antithesis_instrumentation__.Notify(238590)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238599)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238600)
	}
	__antithesis_instrumentation__.Notify(238591)

	response := &serverpb.StatementsResponse{
		Statements:            []serverpb.StatementsResponse_CollectedStatementStatistics{},
		LastReset:             timeutil.Now(),
		InternalAppNamePrefix: catconstants.InternalAppNamePrefix,
	}

	localReq := &serverpb.StatementsRequest{
		NodeID: "local",
	}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(238601)

		parsedInstanceID, local, err := t.parseInstanceID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238606)
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(238607)
		}
		__antithesis_instrumentation__.Notify(238602)
		if local {
			__antithesis_instrumentation__.Notify(238608)
			return statementsLocal(
				ctx,
				roachpb.NodeID(t.sqlServer.SQLInstanceID()),
				t.sqlServer,
				req.FetchMode)
		} else {
			__antithesis_instrumentation__.Notify(238609)
		}
		__antithesis_instrumentation__.Notify(238603)
		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, parsedInstanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238610)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238611)
		}
		__antithesis_instrumentation__.Notify(238604)
		statusClient, err := t.dialPod(ctx, parsedInstanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238612)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238613)
		}
		__antithesis_instrumentation__.Notify(238605)
		return statusClient.Statements(ctx, localReq)
	} else {
		__antithesis_instrumentation__.Notify(238614)
	}
	__antithesis_instrumentation__.Notify(238592)

	nodeStatement := func(ctx context.Context, client interface{}, instanceID base.SQLInstanceID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(238615)
		statusClient := client.(serverpb.StatusClient)
		localResponse, err := statusClient.Statements(ctx, localReq)
		if localResponse == nil {
			__antithesis_instrumentation__.Notify(238617)
			log.Errorf(ctx, "listing statements on %d produced a nil result with err: %v",
				instanceID,
				err)
		} else {
			__antithesis_instrumentation__.Notify(238618)
		}
		__antithesis_instrumentation__.Notify(238616)
		return localResponse, err
	}
	__antithesis_instrumentation__.Notify(238593)

	if err := t.iteratePods(ctx, "statement statistics",
		t.dialCallback,
		nodeStatement,
		func(instanceID base.SQLInstanceID, resp interface{}) {
			__antithesis_instrumentation__.Notify(238619)
			statementsResp := resp.(*serverpb.StatementsResponse)
			response.Statements = append(response.Statements, statementsResp.Statements...)
			response.Transactions = append(response.Transactions, statementsResp.Transactions...)
			if response.LastReset.After(statementsResp.LastReset) {
				__antithesis_instrumentation__.Notify(238620)
				response.LastReset = statementsResp.LastReset
			} else {
				__antithesis_instrumentation__.Notify(238621)
			}
		},
		func(instanceID base.SQLInstanceID, err error) {
			__antithesis_instrumentation__.Notify(238622)

			log.Warningf(ctx, "fan out statements request recorded error from node %d: %v", instanceID, err)
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(238623)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238624)
	}
	__antithesis_instrumentation__.Notify(238594)

	return response, nil
}

func (t *tenantStatusServer) parseInstanceID(
	instanceIDParam string,
) (instanceID base.SQLInstanceID, isLocal bool, err error) {
	__antithesis_instrumentation__.Notify(238625)

	if len(instanceIDParam) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(238628)
		return localRE.MatchString(instanceIDParam) == true
	}() == true {
		__antithesis_instrumentation__.Notify(238629)
		return t.sqlServer.SQLInstanceID(), true, nil
	} else {
		__antithesis_instrumentation__.Notify(238630)
	}
	__antithesis_instrumentation__.Notify(238626)

	id, err := strconv.ParseInt(instanceIDParam, 0, 32)
	if err != nil {
		__antithesis_instrumentation__.Notify(238631)
		return 0, false, errors.Wrap(err, "instance ID could not be parsed")
	} else {
		__antithesis_instrumentation__.Notify(238632)
	}
	__antithesis_instrumentation__.Notify(238627)
	instanceID = base.SQLInstanceID(id)
	return instanceID, instanceID == t.sqlServer.SQLInstanceID(), nil
}

func (t *tenantStatusServer) dialPod(
	ctx context.Context, instanceID base.SQLInstanceID, addr string,
) (serverpb.StatusClient, error) {
	__antithesis_instrumentation__.Notify(238633)
	conn, err := t.rpcCtx.GRPCDialPod(addr, instanceID, rpc.DefaultClass).Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(238635)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238636)
	}
	__antithesis_instrumentation__.Notify(238634)

	return serverpb.NewStatusClient(conn), nil
}

func (t *tenantStatusServer) iteratePods(
	ctx context.Context,
	errorCtx string,
	dialFn func(ctx context.Context, instanceID base.SQLInstanceID, addr string) (interface{}, error),
	instanceFn func(ctx context.Context, client interface{}, instanceID base.SQLInstanceID) (interface{}, error),
	responseFn func(instanceID base.SQLInstanceID, resp interface{}),
	errorFn func(instanceID base.SQLInstanceID, nodeFnError error),
) error {
	__antithesis_instrumentation__.Notify(238637)
	liveTenantInstances, err := t.sqlServer.sqlInstanceProvider.GetAllInstances(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(238642)
		return err
	} else {
		__antithesis_instrumentation__.Notify(238643)
	}
	__antithesis_instrumentation__.Notify(238638)

	type instanceResponse struct {
		instanceID base.SQLInstanceID
		response   interface{}
		err        error
	}

	numInstances := len(liveTenantInstances)
	responseChan := make(chan instanceResponse, numInstances)

	instanceQuery := func(ctx context.Context, instance sqlinstance.InstanceInfo) {
		__antithesis_instrumentation__.Notify(238644)
		var client interface{}
		err := contextutil.RunWithTimeout(ctx, "dial instance", base.NetworkTimeout, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(238648)
			var err error
			client, err = dialFn(ctx, instance.InstanceID, instance.InstanceAddr)
			return err
		})
		__antithesis_instrumentation__.Notify(238645)

		instanceID := instance.InstanceID
		if err != nil {
			__antithesis_instrumentation__.Notify(238649)
			err = errors.Wrapf(err, "failed to dial into node %d",
				instanceID)
			responseChan <- instanceResponse{instanceID: instanceID, err: err}
			return
		} else {
			__antithesis_instrumentation__.Notify(238650)
		}
		__antithesis_instrumentation__.Notify(238646)

		res, err := instanceFn(ctx, client, instanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238651)
			err = errors.Wrapf(err, "error requesting %s from instance %d",
				errorCtx, instanceID)
			responseChan <- instanceResponse{instanceID: instanceID, err: err}
			return
		} else {
			__antithesis_instrumentation__.Notify(238652)
		}
		__antithesis_instrumentation__.Notify(238647)
		responseChan <- instanceResponse{instanceID: instanceID, response: res}
	}
	__antithesis_instrumentation__.Notify(238639)

	sem := quotapool.NewIntPool("instance status", maxConcurrentRequests)
	ctx, cancel := t.stopper.WithCancelOnQuiesce(ctx)
	defer cancel()

	for _, instance := range liveTenantInstances {
		__antithesis_instrumentation__.Notify(238653)
		instance := instance
		if err := t.stopper.RunAsyncTaskEx(
			ctx, stop.TaskOpts{
				TaskName:   fmt.Sprintf("server.tenantStatusServer: requesting %s", errorCtx),
				Sem:        sem,
				WaitForSem: true,
			},
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(238654)
				instanceQuery(ctx, instance)
			}); err != nil {
			__antithesis_instrumentation__.Notify(238655)
			return err
		} else {
			__antithesis_instrumentation__.Notify(238656)
		}
	}
	__antithesis_instrumentation__.Notify(238640)

	var resultErr error
	for numInstances > 0 {
		__antithesis_instrumentation__.Notify(238657)
		select {
		case res := <-responseChan:
			__antithesis_instrumentation__.Notify(238659)
			if res.err != nil {
				__antithesis_instrumentation__.Notify(238661)
				errorFn(res.instanceID, res.err)
			} else {
				__antithesis_instrumentation__.Notify(238662)
				responseFn(res.instanceID, res.response)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(238660)
			resultErr = errors.Errorf("request of %s canceled before completion", errorCtx)
		}
		__antithesis_instrumentation__.Notify(238658)
		numInstances--
	}
	__antithesis_instrumentation__.Notify(238641)
	return resultErr
}

func (t *tenantStatusServer) ListDistSQLFlows(
	ctx context.Context, request *serverpb.ListDistSQLFlowsRequest,
) (*serverpb.ListDistSQLFlowsResponse, error) {
	__antithesis_instrumentation__.Notify(238663)
	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238665)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238666)
	}
	__antithesis_instrumentation__.Notify(238664)

	return t.ListLocalDistSQLFlows(ctx, request)
}

func (t *tenantStatusServer) ListLocalDistSQLFlows(
	ctx context.Context, request *serverpb.ListDistSQLFlowsRequest,
) (*serverpb.ListDistSQLFlowsResponse, error) {
	__antithesis_instrumentation__.Notify(238667)
	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238669)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238670)
	}
	__antithesis_instrumentation__.Notify(238668)

	return t.baseStatusServer.ListLocalDistSQLFlows(ctx, request)
}

func (t *tenantStatusServer) Profile(
	ctx context.Context, request *serverpb.ProfileRequest,
) (*serverpb.JSONResponse, error) {
	__antithesis_instrumentation__.Notify(238671)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if _, err := t.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238676)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238677)
	}
	__antithesis_instrumentation__.Notify(238672)
	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238678)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238679)
	}
	__antithesis_instrumentation__.Notify(238673)

	instanceID, local, err := t.parseInstanceID(request.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(238680)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(238681)
	}
	__antithesis_instrumentation__.Notify(238674)
	if !local {
		__antithesis_instrumentation__.Notify(238682)
		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, instanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238685)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238686)
		}
		__antithesis_instrumentation__.Notify(238683)
		status, err := t.dialPod(ctx, instanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238687)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238688)
		}
		__antithesis_instrumentation__.Notify(238684)
		return status.Profile(ctx, request)
	} else {
		__antithesis_instrumentation__.Notify(238689)
	}
	__antithesis_instrumentation__.Notify(238675)
	return profileLocal(ctx, request, t.st)
}

func (t *tenantStatusServer) Stacks(
	ctx context.Context, request *serverpb.StacksRequest,
) (*serverpb.JSONResponse, error) {
	__antithesis_instrumentation__.Notify(238690)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if _, err := t.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238695)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238696)
	}
	__antithesis_instrumentation__.Notify(238691)
	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238697)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238698)
	}
	__antithesis_instrumentation__.Notify(238692)

	instanceID, local, err := t.parseInstanceID(request.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(238699)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(238700)
	}
	__antithesis_instrumentation__.Notify(238693)
	if !local {
		__antithesis_instrumentation__.Notify(238701)
		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, instanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238704)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238705)
		}
		__antithesis_instrumentation__.Notify(238702)
		status, err := t.dialPod(ctx, instanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238706)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238707)
		}
		__antithesis_instrumentation__.Notify(238703)
		return status.Stacks(ctx, request)
	} else {
		__antithesis_instrumentation__.Notify(238708)
	}
	__antithesis_instrumentation__.Notify(238694)
	return stacksLocal(request)
}

func (t *tenantStatusServer) IndexUsageStatistics(
	ctx context.Context, req *serverpb.IndexUsageStatisticsRequest,
) (*serverpb.IndexUsageStatisticsResponse, error) {
	__antithesis_instrumentation__.Notify(238709)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238717)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238718)
	}
	__antithesis_instrumentation__.Notify(238710)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238719)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238720)
	}
	__antithesis_instrumentation__.Notify(238711)

	localReq := &serverpb.IndexUsageStatisticsRequest{
		NodeID: "local",
	}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(238721)
		parsedInstanceID, local, err := t.parseInstanceID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238726)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(238727)
		}
		__antithesis_instrumentation__.Notify(238722)
		if local {
			__antithesis_instrumentation__.Notify(238728)
			statsReader := t.sqlServer.pgServer.SQLServer.GetLocalIndexStatistics()
			return indexUsageStatsLocal(statsReader)
		} else {
			__antithesis_instrumentation__.Notify(238729)
		}
		__antithesis_instrumentation__.Notify(238723)

		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, parsedInstanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238730)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238731)
		}
		__antithesis_instrumentation__.Notify(238724)
		statusClient, err := t.dialPod(ctx, parsedInstanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238732)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238733)
		}
		__antithesis_instrumentation__.Notify(238725)

		return statusClient.IndexUsageStatistics(ctx, localReq)
	} else {
		__antithesis_instrumentation__.Notify(238734)
	}
	__antithesis_instrumentation__.Notify(238712)

	fetchIndexUsageStats := func(ctx context.Context, client interface{}, _ base.SQLInstanceID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(238735)
		statusClient := client.(serverpb.StatusClient)
		return statusClient.IndexUsageStatistics(ctx, localReq)
	}
	__antithesis_instrumentation__.Notify(238713)

	resp := &serverpb.IndexUsageStatisticsResponse{}
	aggFn := func(_ base.SQLInstanceID, nodeResp interface{}) {
		__antithesis_instrumentation__.Notify(238736)
		stats := nodeResp.(*serverpb.IndexUsageStatisticsResponse)
		resp.Statistics = append(resp.Statistics, stats.Statistics...)
	}
	__antithesis_instrumentation__.Notify(238714)

	var combinedError error
	errFn := func(_ base.SQLInstanceID, nodeFnError error) {
		__antithesis_instrumentation__.Notify(238737)
		combinedError = errors.CombineErrors(combinedError, nodeFnError)
	}
	__antithesis_instrumentation__.Notify(238715)

	if err := t.iteratePods(ctx, "requesting index usage stats",
		t.dialCallback,
		fetchIndexUsageStats,
		aggFn,
		errFn,
	); err != nil {
		__antithesis_instrumentation__.Notify(238738)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238739)
	}
	__antithesis_instrumentation__.Notify(238716)

	resp.LastReset = t.sqlServer.pgServer.SQLServer.GetLocalIndexStatistics().GetLastReset()

	return resp, nil
}

func (t *tenantStatusServer) ResetIndexUsageStats(
	ctx context.Context, req *serverpb.ResetIndexUsageStatsRequest,
) (*serverpb.ResetIndexUsageStatsResponse, error) {
	__antithesis_instrumentation__.Notify(238740)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if _, err := t.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238746)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238747)
	}
	__antithesis_instrumentation__.Notify(238741)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238748)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238749)
	}
	__antithesis_instrumentation__.Notify(238742)

	localReq := &serverpb.ResetIndexUsageStatsRequest{
		NodeID: "local",
	}
	resp := &serverpb.ResetIndexUsageStatsResponse{}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(238750)
		parsedInstanceID, local, err := t.parseInstanceID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238755)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(238756)
		}
		__antithesis_instrumentation__.Notify(238751)
		if local {
			__antithesis_instrumentation__.Notify(238757)
			t.sqlServer.pgServer.SQLServer.GetLocalIndexStatistics().Reset()
			return resp, nil
		} else {
			__antithesis_instrumentation__.Notify(238758)
		}
		__antithesis_instrumentation__.Notify(238752)

		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, parsedInstanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238759)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238760)
		}
		__antithesis_instrumentation__.Notify(238753)
		statusClient, err := t.dialPod(ctx, parsedInstanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238761)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238762)
		}
		__antithesis_instrumentation__.Notify(238754)

		return statusClient.ResetIndexUsageStats(ctx, localReq)
	} else {
		__antithesis_instrumentation__.Notify(238763)
	}
	__antithesis_instrumentation__.Notify(238743)

	resetIndexUsageStats := func(ctx context.Context, client interface{}, _ base.SQLInstanceID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(238764)
		statusClient := client.(serverpb.StatusClient)
		return statusClient.ResetIndexUsageStats(ctx, localReq)
	}
	__antithesis_instrumentation__.Notify(238744)

	var combinedError error

	if err := t.iteratePods(ctx, "Resetting index usage stats for instance",
		t.dialCallback,
		resetIndexUsageStats,
		func(instanceID base.SQLInstanceID, resp interface{}) {
			__antithesis_instrumentation__.Notify(238765)

		},
		func(_ base.SQLInstanceID, err error) {
			__antithesis_instrumentation__.Notify(238766)
			combinedError = errors.CombineErrors(combinedError, err)
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(238767)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238768)
	}
	__antithesis_instrumentation__.Notify(238745)

	return resp, nil
}

func (t *tenantStatusServer) TableIndexStats(
	ctx context.Context, req *serverpb.TableIndexStatsRequest,
) (*serverpb.TableIndexStatsResponse, error) {
	__antithesis_instrumentation__.Notify(238769)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238772)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238773)
	}
	__antithesis_instrumentation__.Notify(238770)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238774)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238775)
	}
	__antithesis_instrumentation__.Notify(238771)

	return getTableIndexUsageStats(ctx, req, t.sqlServer.pgServer.SQLServer.GetLocalIndexStatistics(),
		t.sqlServer.internalExecutor)
}

func (t *tenantStatusServer) Details(
	ctx context.Context, req *serverpb.DetailsRequest,
) (*serverpb.DetailsResponse, error) {
	__antithesis_instrumentation__.Notify(238776)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238782)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238783)
	}
	__antithesis_instrumentation__.Notify(238777)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238784)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238785)
	}
	__antithesis_instrumentation__.Notify(238778)

	instanceID, local, err := t.parseInstanceID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(238786)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(238787)
	}
	__antithesis_instrumentation__.Notify(238779)
	if !local {
		__antithesis_instrumentation__.Notify(238788)
		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, instanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238791)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238792)
		}
		__antithesis_instrumentation__.Notify(238789)
		status, err := t.dialPod(ctx, instanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238793)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238794)
		}
		__antithesis_instrumentation__.Notify(238790)
		return status.Details(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(238795)
	}
	__antithesis_instrumentation__.Notify(238780)
	localInstance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, t.sqlServer.SQLInstanceID())
	if err != nil {
		__antithesis_instrumentation__.Notify(238796)
		return nil, status.Errorf(codes.Unavailable, "local instance unavailable")
	} else {
		__antithesis_instrumentation__.Notify(238797)
	}
	__antithesis_instrumentation__.Notify(238781)
	resp := &serverpb.DetailsResponse{
		NodeID:     roachpb.NodeID(instanceID),
		BuildInfo:  build.GetInfo(),
		SQLAddress: util.MakeUnresolvedAddr("tcp", localInstance.InstanceAddr),
	}

	return resp, nil
}

func (t *tenantStatusServer) NodesList(
	ctx context.Context, req *serverpb.NodesListRequest,
) (*serverpb.NodesListResponse, error) {
	__antithesis_instrumentation__.Notify(238798)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if _, err := t.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238802)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238803)
	}
	__antithesis_instrumentation__.Notify(238799)
	instances, err := t.sqlServer.sqlInstanceProvider.GetAllInstances(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(238804)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238805)
	}
	__antithesis_instrumentation__.Notify(238800)
	var resp serverpb.NodesListResponse
	for _, instance := range instances {
		__antithesis_instrumentation__.Notify(238806)

		nodeDetails := serverpb.NodeDetails{
			NodeID:     int32(instance.InstanceID),
			Address:    util.MakeUnresolvedAddr("tcp", instance.InstanceAddr),
			SQLAddress: util.MakeUnresolvedAddr("tcp", instance.InstanceAddr),
		}
		resp.Nodes = append(resp.Nodes, nodeDetails)
	}
	__antithesis_instrumentation__.Notify(238801)
	return &resp, err
}

func (t *tenantStatusServer) TxnIDResolution(
	ctx context.Context, req *serverpb.TxnIDResolutionRequest,
) (*serverpb.TxnIDResolutionResponse, error) {
	__antithesis_instrumentation__.Notify(238807)
	ctx = t.AnnotateCtx(propagateGatewayMetadata(ctx))
	if _, err := t.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238813)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238814)
	}
	__antithesis_instrumentation__.Notify(238808)

	instanceID, local, err := t.parseInstanceID(req.CoordinatorID)
	if err != nil {
		__antithesis_instrumentation__.Notify(238815)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(238816)
	}
	__antithesis_instrumentation__.Notify(238809)
	if local {
		__antithesis_instrumentation__.Notify(238817)
		return t.localTxnIDResolution(req), nil
	} else {
		__antithesis_instrumentation__.Notify(238818)
	}
	__antithesis_instrumentation__.Notify(238810)

	instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, instanceID)
	if err != nil {
		__antithesis_instrumentation__.Notify(238819)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238820)
	}
	__antithesis_instrumentation__.Notify(238811)
	statusClient, err := t.dialPod(ctx, instanceID, instance.InstanceAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(238821)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238822)
	}
	__antithesis_instrumentation__.Notify(238812)

	return statusClient.TxnIDResolution(ctx, req)
}

func (t *tenantStatusServer) TenantRanges(
	ctx context.Context, req *serverpb.TenantRangesRequest,
) (*serverpb.TenantRangesResponse, error) {
	__antithesis_instrumentation__.Notify(238823)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if _, err := t.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238825)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238826)
	}
	__antithesis_instrumentation__.Notify(238824)

	return t.sqlServer.tenantConnect.TenantRanges(ctx, req)
}

func (t *tenantStatusServer) GetFiles(
	ctx context.Context, req *serverpb.GetFilesRequest,
) (*serverpb.GetFilesResponse, error) {
	__antithesis_instrumentation__.Notify(238827)
	ctx = propagateGatewayMetadata(ctx)
	ctx = t.AnnotateCtx(ctx)

	if _, err := t.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238831)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238832)
	}
	__antithesis_instrumentation__.Notify(238828)

	instanceID, local, err := t.parseInstanceID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(238833)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(238834)
	}
	__antithesis_instrumentation__.Notify(238829)
	if !local {
		__antithesis_instrumentation__.Notify(238835)
		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, instanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238838)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238839)
		}
		__antithesis_instrumentation__.Notify(238836)
		status, err := t.dialPod(ctx, instanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238840)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238841)
		}
		__antithesis_instrumentation__.Notify(238837)
		return status.GetFiles(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(238842)
	}
	__antithesis_instrumentation__.Notify(238830)

	return getLocalFiles(req, t.sqlServer.cfg.HeapProfileDirName, t.sqlServer.cfg.GoroutineDumpDirName)
}

func (t *tenantStatusServer) TransactionContentionEvents(
	ctx context.Context, req *serverpb.TransactionContentionEventsRequest,
) (*serverpb.TransactionContentionEventsResponse, error) {
	__antithesis_instrumentation__.Notify(238843)
	ctx = t.AnnotateCtx(propagateGatewayMetadata(ctx))

	if err := t.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(238852)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238853)
	}
	__antithesis_instrumentation__.Notify(238844)

	user, isAdmin, err := t.privilegeChecker.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(238854)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(238855)
	}
	__antithesis_instrumentation__.Notify(238845)

	shouldRedactContendingKey := false
	if !isAdmin {
		__antithesis_instrumentation__.Notify(238856)
		shouldRedactContendingKey, err =
			t.privilegeChecker.hasRoleOption(ctx, user, roleoption.VIEWACTIVITYREDACTED)
		if err != nil {
			__antithesis_instrumentation__.Notify(238857)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(238858)
		}
	} else {
		__antithesis_instrumentation__.Notify(238859)
	}
	__antithesis_instrumentation__.Notify(238846)

	if t.sqlServer.SQLInstanceID() == 0 {
		__antithesis_instrumentation__.Notify(238860)
		return nil, status.Errorf(codes.Unavailable, "instanceID not set")
	} else {
		__antithesis_instrumentation__.Notify(238861)
	}
	__antithesis_instrumentation__.Notify(238847)

	resp := &serverpb.TransactionContentionEventsResponse{}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(238862)
		parsedInstanceID, local, err := t.parseInstanceID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238867)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(238868)
		}
		__antithesis_instrumentation__.Notify(238863)
		if local {
			__antithesis_instrumentation__.Notify(238869)
			return t.localTransactionContentionEvents(shouldRedactContendingKey), nil
		} else {
			__antithesis_instrumentation__.Notify(238870)
		}
		__antithesis_instrumentation__.Notify(238864)

		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, parsedInstanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238871)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238872)
		}
		__antithesis_instrumentation__.Notify(238865)
		statusClient, err := t.dialPod(ctx, parsedInstanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238873)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(238874)
		}
		__antithesis_instrumentation__.Notify(238866)

		return statusClient.TransactionContentionEvents(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(238875)
	}
	__antithesis_instrumentation__.Notify(238848)

	rpcCallFn := func(ctx context.Context, client interface{}, _ base.SQLInstanceID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(238876)
		statusClient := client.(serverpb.StatusClient)
		return statusClient.TransactionContentionEvents(ctx, &serverpb.TransactionContentionEventsRequest{
			NodeID: "local",
		})
	}
	__antithesis_instrumentation__.Notify(238849)

	if err := t.iteratePods(ctx, "txn contention events for instance",
		t.dialCallback,
		rpcCallFn,
		func(instanceID base.SQLInstanceID, nodeResp interface{}) {
			__antithesis_instrumentation__.Notify(238877)
			txnContentionEvents := nodeResp.(*serverpb.TransactionContentionEventsResponse)
			resp.Events = append(resp.Events, txnContentionEvents.Events...)
		},
		func(_ base.SQLInstanceID, err error) {
			__antithesis_instrumentation__.Notify(238878)
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(238879)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238880)
	}
	__antithesis_instrumentation__.Notify(238850)

	sort.Slice(resp.Events, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(238881)
		return resp.Events[i].CollectionTs.Before(resp.Events[j].CollectionTs)
	})
	__antithesis_instrumentation__.Notify(238851)

	return resp, nil
}
