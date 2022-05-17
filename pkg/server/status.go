package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/diagnostics/diagnosticspb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/contention"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirecancel"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logpb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	raft "go.etcd.io/etcd/raft/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultMaxLogEntries = 1000

	statusPrefix = "/_status/"

	statusVars = statusPrefix + "vars"

	loadStatusVars = statusPrefix + "load"

	raftStateDormant = "StateDormant"

	maxConcurrentRequests = 100

	maxConcurrentPaginatedRequests = 4
)

var (
	localRE = regexp.MustCompile(`(?i)local`)

	telemetryPrometheusVars = telemetry.GetCounterOnce("monitoring.prometheus.vars")

	telemetryHealthCheck = telemetry.GetCounterOnce("monitoring.health.details")
)

type metricMarshaler interface {
	json.Marshaler
	PrintAsText(io.Writer) error
	ScrapeIntoPrometheus(pm *metric.PrometheusExporter)
}

func propagateGatewayMetadata(ctx context.Context) context.Context {
	__antithesis_instrumentation__.Notify(236789)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		__antithesis_instrumentation__.Notify(236791)
		return metadata.NewOutgoingContext(ctx, md)
	} else {
		__antithesis_instrumentation__.Notify(236792)
	}
	__antithesis_instrumentation__.Notify(236790)
	return ctx
}

type baseStatusServer struct {
	serverpb.UnimplementedStatusServer

	log.AmbientContext
	privilegeChecker *adminPrivilegeChecker
	sessionRegistry  *sql.SessionRegistry
	flowScheduler    *flowinfra.FlowScheduler
	st               *cluster.Settings
	sqlServer        *SQLServer
	rpcCtx           *rpc.Context
	stopper          *stop.Stopper
}

func (b *baseStatusServer) getLocalSessions(
	ctx context.Context, req *serverpb.ListSessionsRequest,
) ([]serverpb.Session, error) {
	__antithesis_instrumentation__.Notify(236793)
	ctx = propagateGatewayMetadata(ctx)
	ctx = b.AnnotateCtx(ctx)

	sessionUser, isAdmin, err := b.privilegeChecker.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(236800)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(236801)
	}
	__antithesis_instrumentation__.Notify(236794)

	hasViewActivityRedacted, err := b.privilegeChecker.hasRoleOption(ctx, sessionUser, roleoption.VIEWACTIVITYREDACTED)
	if err != nil {
		__antithesis_instrumentation__.Notify(236802)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(236803)
	}
	__antithesis_instrumentation__.Notify(236795)

	hasViewActivity, err := b.privilegeChecker.hasRoleOption(ctx, sessionUser, roleoption.VIEWACTIVITY)
	if err != nil {
		__antithesis_instrumentation__.Notify(236804)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(236805)
	}
	__antithesis_instrumentation__.Notify(236796)

	reqUsername, err := security.MakeSQLUsernameFromPreNormalizedStringChecked(req.Username)
	if err != nil {
		__antithesis_instrumentation__.Notify(236806)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(236807)
	}
	__antithesis_instrumentation__.Notify(236797)

	if !isAdmin && func() bool {
		__antithesis_instrumentation__.Notify(236808)
		return !hasViewActivity == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(236809)
		return !hasViewActivityRedacted == true
	}() == true {
		__antithesis_instrumentation__.Notify(236810)

		if reqUsername.Undefined() {
			__antithesis_instrumentation__.Notify(236812)
			reqUsername = sessionUser
		} else {
			__antithesis_instrumentation__.Notify(236813)
		}
		__antithesis_instrumentation__.Notify(236811)

		if sessionUser != reqUsername {
			__antithesis_instrumentation__.Notify(236814)
			return nil, status.Errorf(
				codes.PermissionDenied,
				"client user %q does not have permission to view sessions from user %q",
				sessionUser, reqUsername)
		} else {
			__antithesis_instrumentation__.Notify(236815)
		}
	} else {
		__antithesis_instrumentation__.Notify(236816)
	}
	__antithesis_instrumentation__.Notify(236798)

	showAll := reqUsername.Undefined()

	registry := b.sessionRegistry
	sessions := registry.SerializeAll()
	userSessions := make([]serverpb.Session, 0, len(sessions))

	for _, session := range sessions {
		__antithesis_instrumentation__.Notify(236817)
		if reqUsername.Normalized() != session.Username && func() bool {
			__antithesis_instrumentation__.Notify(236820)
			return !showAll == true
		}() == true {
			__antithesis_instrumentation__.Notify(236821)
			continue
		} else {
			__antithesis_instrumentation__.Notify(236822)
		}
		__antithesis_instrumentation__.Notify(236818)

		if !isAdmin && func() bool {
			__antithesis_instrumentation__.Notify(236823)
			return hasViewActivityRedacted == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(236824)
			return (sessionUser != reqUsername) == true
		}() == true {
			__antithesis_instrumentation__.Notify(236825)

			for _, query := range session.ActiveQueries {
				__antithesis_instrumentation__.Notify(236827)
				query.Sql = ""
			}
			__antithesis_instrumentation__.Notify(236826)
			session.LastActiveQuery = ""
		} else {
			__antithesis_instrumentation__.Notify(236828)
		}
		__antithesis_instrumentation__.Notify(236819)

		userSessions = append(userSessions, session)
	}
	__antithesis_instrumentation__.Notify(236799)
	return userSessions, nil
}

type sessionFinder func(sessions []serverpb.Session) (serverpb.Session, error)

func findSessionBySessionID(sessionID []byte) sessionFinder {
	__antithesis_instrumentation__.Notify(236829)
	return func(sessions []serverpb.Session) (serverpb.Session, error) {
		__antithesis_instrumentation__.Notify(236830)
		var session serverpb.Session
		for _, s := range sessions {
			__antithesis_instrumentation__.Notify(236833)
			if bytes.Equal(sessionID, s.ID) {
				__antithesis_instrumentation__.Notify(236834)
				session = s
				break
			} else {
				__antithesis_instrumentation__.Notify(236835)
			}
		}
		__antithesis_instrumentation__.Notify(236831)
		if len(session.ID) == 0 {
			__antithesis_instrumentation__.Notify(236836)
			return session, fmt.Errorf("session ID %s not found", sql.BytesToClusterWideID(sessionID))
		} else {
			__antithesis_instrumentation__.Notify(236837)
		}
		__antithesis_instrumentation__.Notify(236832)
		return session, nil
	}
}

func findSessionByQueryID(queryID string) sessionFinder {
	__antithesis_instrumentation__.Notify(236838)
	return func(sessions []serverpb.Session) (serverpb.Session, error) {
		__antithesis_instrumentation__.Notify(236839)
		var session serverpb.Session
		for _, s := range sessions {
			__antithesis_instrumentation__.Notify(236842)
			for _, q := range s.ActiveQueries {
				__antithesis_instrumentation__.Notify(236843)
				if queryID == q.ID {
					__antithesis_instrumentation__.Notify(236844)
					session = s
					break
				} else {
					__antithesis_instrumentation__.Notify(236845)
				}
			}
		}
		__antithesis_instrumentation__.Notify(236840)
		if len(session.ID) == 0 {
			__antithesis_instrumentation__.Notify(236846)
			return session, fmt.Errorf("query ID %s not found", queryID)
		} else {
			__antithesis_instrumentation__.Notify(236847)
		}
		__antithesis_instrumentation__.Notify(236841)
		return session, nil
	}
}

func (b *baseStatusServer) checkCancelPrivilege(
	ctx context.Context, username security.SQLUsername, findSession sessionFinder,
) error {
	__antithesis_instrumentation__.Notify(236848)
	ctx = propagateGatewayMetadata(ctx)
	ctx = b.AnnotateCtx(ctx)

	var reqUser security.SQLUsername
	{
		__antithesis_instrumentation__.Notify(236852)
		sessionUser, isAdmin, err := b.privilegeChecker.getUserAndRole(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(236854)
			return err
		} else {
			__antithesis_instrumentation__.Notify(236855)
		}
		__antithesis_instrumentation__.Notify(236853)
		if username.Undefined() || func() bool {
			__antithesis_instrumentation__.Notify(236856)
			return username == sessionUser == true
		}() == true {
			__antithesis_instrumentation__.Notify(236857)
			reqUser = sessionUser
		} else {
			__antithesis_instrumentation__.Notify(236858)

			if !isAdmin {
				__antithesis_instrumentation__.Notify(236860)
				return errRequiresAdmin
			} else {
				__antithesis_instrumentation__.Notify(236861)
			}
			__antithesis_instrumentation__.Notify(236859)
			reqUser = username
		}
	}
	__antithesis_instrumentation__.Notify(236849)

	hasAdmin, err := b.privilegeChecker.hasAdminRole(ctx, reqUser)
	if err != nil {
		__antithesis_instrumentation__.Notify(236862)
		return serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(236863)
	}
	__antithesis_instrumentation__.Notify(236850)

	if !hasAdmin {
		__antithesis_instrumentation__.Notify(236864)

		session, err := findSession(b.sessionRegistry.SerializeAll())
		if err != nil {
			__antithesis_instrumentation__.Notify(236866)
			return serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(236867)
		}
		__antithesis_instrumentation__.Notify(236865)

		sessionUser := security.MakeSQLUsernameFromPreNormalizedString(session.Username)
		if sessionUser != reqUser {
			__antithesis_instrumentation__.Notify(236868)

			ok, err := b.privilegeChecker.hasRoleOption(ctx, reqUser, roleoption.CANCELQUERY)
			if err != nil {
				__antithesis_instrumentation__.Notify(236872)
				return serverError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(236873)
			}
			__antithesis_instrumentation__.Notify(236869)
			if !ok {
				__antithesis_instrumentation__.Notify(236874)
				return errRequiresRoleOption(roleoption.CANCELQUERY)
			} else {
				__antithesis_instrumentation__.Notify(236875)
			}
			__antithesis_instrumentation__.Notify(236870)

			isAdminSession, err := b.privilegeChecker.hasAdminRole(ctx, sessionUser)
			if err != nil {
				__antithesis_instrumentation__.Notify(236876)
				return serverError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(236877)
			}
			__antithesis_instrumentation__.Notify(236871)
			if isAdminSession {
				__antithesis_instrumentation__.Notify(236878)
				return status.Error(
					codes.PermissionDenied, "permission denied to cancel admin session")
			} else {
				__antithesis_instrumentation__.Notify(236879)
			}
		} else {
			__antithesis_instrumentation__.Notify(236880)
		}
	} else {
		__antithesis_instrumentation__.Notify(236881)
	}
	__antithesis_instrumentation__.Notify(236851)

	return nil
}

func (b *baseStatusServer) ListLocalContentionEvents(
	ctx context.Context, _ *serverpb.ListContentionEventsRequest,
) (*serverpb.ListContentionEventsResponse, error) {
	__antithesis_instrumentation__.Notify(236882)
	ctx = propagateGatewayMetadata(ctx)
	ctx = b.AnnotateCtx(ctx)

	if err := b.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(236884)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(236885)
	}
	__antithesis_instrumentation__.Notify(236883)

	return &serverpb.ListContentionEventsResponse{
		Events: b.sqlServer.execCfg.ContentionRegistry.Serialize(),
	}, nil
}

func (b *baseStatusServer) ListLocalDistSQLFlows(
	ctx context.Context, _ *serverpb.ListDistSQLFlowsRequest,
) (*serverpb.ListDistSQLFlowsResponse, error) {
	__antithesis_instrumentation__.Notify(236886)
	ctx = propagateGatewayMetadata(ctx)
	ctx = b.AnnotateCtx(ctx)

	if err := b.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(236891)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(236892)
	}
	__antithesis_instrumentation__.Notify(236887)

	nodeIDOrZero, _ := b.sqlServer.sqlIDContainer.OptionalNodeID()

	running, queued := b.flowScheduler.Serialize()
	response := &serverpb.ListDistSQLFlowsResponse{
		Flows: make([]serverpb.DistSQLRemoteFlows, 0, len(running)+len(queued)),
	}
	for _, f := range running {
		__antithesis_instrumentation__.Notify(236893)
		response.Flows = append(response.Flows, serverpb.DistSQLRemoteFlows{
			FlowID: f.FlowID,
			Infos: []serverpb.DistSQLRemoteFlows_Info{{
				NodeID:    nodeIDOrZero,
				Timestamp: f.Timestamp,
				Status:    serverpb.DistSQLRemoteFlows_RUNNING,
				Stmt:      f.StatementSQL,
			}},
		})
	}
	__antithesis_instrumentation__.Notify(236888)
	for _, f := range queued {
		__antithesis_instrumentation__.Notify(236894)
		response.Flows = append(response.Flows, serverpb.DistSQLRemoteFlows{
			FlowID: f.FlowID,
			Infos: []serverpb.DistSQLRemoteFlows_Info{{
				NodeID:    nodeIDOrZero,
				Timestamp: f.Timestamp,
				Status:    serverpb.DistSQLRemoteFlows_QUEUED,
				Stmt:      f.StatementSQL,
			}},
		})
	}
	__antithesis_instrumentation__.Notify(236889)

	sort.Slice(response.Flows, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(236895)
		return bytes.Compare(response.Flows[i].FlowID.GetBytes(), response.Flows[j].FlowID.GetBytes()) < 0
	})
	__antithesis_instrumentation__.Notify(236890)
	return response, nil
}

func (b *baseStatusServer) localTxnIDResolution(
	req *serverpb.TxnIDResolutionRequest,
) *serverpb.TxnIDResolutionResponse {
	__antithesis_instrumentation__.Notify(236896)
	txnIDCache := b.sqlServer.pgServer.SQLServer.GetTxnIDCache()

	unresolvedTxnIDs := make(map[uuid.UUID]struct{}, len(req.TxnIDs))
	for _, txnID := range req.TxnIDs {
		__antithesis_instrumentation__.Notify(236900)
		unresolvedTxnIDs[txnID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(236897)

	resp := &serverpb.TxnIDResolutionResponse{
		ResolvedTxnIDs: make([]contentionpb.ResolvedTxnID, 0, len(req.TxnIDs)),
	}

	for i := range req.TxnIDs {
		__antithesis_instrumentation__.Notify(236901)
		if txnFingerprintID, found := txnIDCache.Lookup(req.TxnIDs[i]); found {
			__antithesis_instrumentation__.Notify(236902)
			resp.ResolvedTxnIDs = append(resp.ResolvedTxnIDs, contentionpb.ResolvedTxnID{
				TxnID:            req.TxnIDs[i],
				TxnFingerprintID: txnFingerprintID,
			})
		} else {
			__antithesis_instrumentation__.Notify(236903)
		}
	}
	__antithesis_instrumentation__.Notify(236898)

	if len(unresolvedTxnIDs) > 0 {
		__antithesis_instrumentation__.Notify(236904)
		txnIDCache.DrainWriteBuffer()
	} else {
		__antithesis_instrumentation__.Notify(236905)
	}
	__antithesis_instrumentation__.Notify(236899)

	return resp
}

func (b *baseStatusServer) localTransactionContentionEvents(
	shouldRedactContendingKey bool,
) *serverpb.TransactionContentionEventsResponse {
	__antithesis_instrumentation__.Notify(236906)
	registry := b.sqlServer.execCfg.ContentionRegistry

	resp := &serverpb.TransactionContentionEventsResponse{
		Events: make([]contentionpb.ExtendedContentionEvent, 0),
	}

	_ = registry.ForEachEvent(func(event *contentionpb.ExtendedContentionEvent) error {
		__antithesis_instrumentation__.Notify(236908)
		if shouldRedactContendingKey {
			__antithesis_instrumentation__.Notify(236910)
			event.BlockingEvent.Key = []byte{}
		} else {
			__antithesis_instrumentation__.Notify(236911)
		}
		__antithesis_instrumentation__.Notify(236909)
		resp.Events = append(resp.Events, *event)
		return nil
	})
	__antithesis_instrumentation__.Notify(236907)

	return resp
}

type statusServer struct {
	*baseStatusServer

	cfg                      *base.Config
	admin                    *adminServer
	db                       *kv.DB
	gossip                   *gossip.Gossip
	metricSource             metricMarshaler
	nodeLiveness             *liveness.NodeLiveness
	storePool                *kvserver.StorePool
	stores                   *kvserver.Stores
	si                       systemInfoOnce
	stmtDiagnosticsRequester StmtDiagnosticsRequester
	internalExecutor         *sql.InternalExecutor
}

type StmtDiagnosticsRequester interface {
	InsertRequest(
		ctx context.Context,
		stmtFingerprint string,
		minExecutionLatency time.Duration,
		expiresAfter time.Duration,
	) error

	CancelRequest(ctx context.Context, requestID int64) error
}

func newStatusServer(
	ambient log.AmbientContext,
	st *cluster.Settings,
	cfg *base.Config,
	adminAuthzCheck *adminPrivilegeChecker,
	adminServer *adminServer,
	db *kv.DB,
	gossip *gossip.Gossip,
	metricSource metricMarshaler,
	nodeLiveness *liveness.NodeLiveness,
	storePool *kvserver.StorePool,
	rpcCtx *rpc.Context,
	stores *kvserver.Stores,
	stopper *stop.Stopper,
	sessionRegistry *sql.SessionRegistry,
	flowScheduler *flowinfra.FlowScheduler,
	internalExecutor *sql.InternalExecutor,
) *statusServer {
	__antithesis_instrumentation__.Notify(236912)
	ambient.AddLogTag("status", nil)
	server := &statusServer{
		baseStatusServer: &baseStatusServer{
			AmbientContext:   ambient,
			privilegeChecker: adminAuthzCheck,
			sessionRegistry:  sessionRegistry,
			flowScheduler:    flowScheduler,
			st:               st,
			rpcCtx:           rpcCtx,
			stopper:          stopper,
		},
		cfg:              cfg,
		admin:            adminServer,
		db:               db,
		gossip:           gossip,
		metricSource:     metricSource,
		nodeLiveness:     nodeLiveness,
		storePool:        storePool,
		stores:           stores,
		internalExecutor: internalExecutor,
	}

	return server
}

func (s *statusServer) setStmtDiagnosticsRequester(sr StmtDiagnosticsRequester) {
	__antithesis_instrumentation__.Notify(236913)
	s.stmtDiagnosticsRequester = sr
}

func (s *statusServer) RegisterService(g *grpc.Server) {
	__antithesis_instrumentation__.Notify(236914)
	serverpb.RegisterStatusServer(g, s)
}

func (s *statusServer) RegisterGateway(
	ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn,
) error {
	__antithesis_instrumentation__.Notify(236915)
	ctx = s.AnnotateCtx(ctx)
	return serverpb.RegisterStatusHandler(ctx, mux, conn)
}

func (s *statusServer) parseNodeID(nodeIDParam string) (roachpb.NodeID, bool, error) {
	__antithesis_instrumentation__.Notify(236916)
	return parseNodeID(s.gossip, nodeIDParam)
}

func parseNodeID(
	gossip *gossip.Gossip, nodeIDParam string,
) (nodeID roachpb.NodeID, isLocal bool, err error) {
	__antithesis_instrumentation__.Notify(236917)

	if len(nodeIDParam) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(236920)
		return localRE.MatchString(nodeIDParam) == true
	}() == true {
		__antithesis_instrumentation__.Notify(236921)
		return gossip.NodeID.Get(), true, nil
	} else {
		__antithesis_instrumentation__.Notify(236922)
	}
	__antithesis_instrumentation__.Notify(236918)

	id, err := strconv.ParseInt(nodeIDParam, 0, 32)
	if err != nil {
		__antithesis_instrumentation__.Notify(236923)
		return 0, false, errors.Wrap(err, "node ID could not be parsed")
	} else {
		__antithesis_instrumentation__.Notify(236924)
	}
	__antithesis_instrumentation__.Notify(236919)
	nodeID = roachpb.NodeID(id)
	return nodeID, nodeID == gossip.NodeID.Get(), nil
}

func (s *statusServer) dialNode(
	ctx context.Context, nodeID roachpb.NodeID,
) (serverpb.StatusClient, error) {
	__antithesis_instrumentation__.Notify(236925)
	addr, err := s.gossip.GetNodeIDAddress(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(236928)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(236929)
	}
	__antithesis_instrumentation__.Notify(236926)
	conn, err := s.rpcCtx.GRPCDialNode(addr.String(), nodeID,
		rpc.DefaultClass).Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(236930)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(236931)
	}
	__antithesis_instrumentation__.Notify(236927)
	return serverpb.NewStatusClient(conn), nil
}

func (s *statusServer) Gossip(
	ctx context.Context, req *serverpb.GossipRequest,
) (*gossip.InfoStatus, error) {
	__antithesis_instrumentation__.Notify(236932)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(236937)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(236938)
	}
	__antithesis_instrumentation__.Notify(236933)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(236939)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(236940)
	}
	__antithesis_instrumentation__.Notify(236934)

	if local {
		__antithesis_instrumentation__.Notify(236941)
		infoStatus := s.gossip.GetInfoStatus()
		return &infoStatus, nil
	} else {
		__antithesis_instrumentation__.Notify(236942)
	}
	__antithesis_instrumentation__.Notify(236935)
	status, err := s.dialNode(ctx, nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(236943)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(236944)
	}
	__antithesis_instrumentation__.Notify(236936)
	return status.Gossip(ctx, req)
}

func (s *statusServer) EngineStats(
	ctx context.Context, req *serverpb.EngineStatsRequest,
) (*serverpb.EngineStatsResponse, error) {
	__antithesis_instrumentation__.Notify(236945)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(236951)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(236952)
	}
	__antithesis_instrumentation__.Notify(236946)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(236953)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(236954)
	}
	__antithesis_instrumentation__.Notify(236947)

	if !local {
		__antithesis_instrumentation__.Notify(236955)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(236957)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(236958)
		}
		__antithesis_instrumentation__.Notify(236956)
		return status.EngineStats(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(236959)
	}
	__antithesis_instrumentation__.Notify(236948)

	resp := new(serverpb.EngineStatsResponse)
	err = s.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(236960)
		engineStatsInfo := serverpb.EngineStatsInfo{
			StoreID:              store.Ident.StoreID,
			TickersAndHistograms: nil,
			EngineType:           store.Engine().Type(),
		}

		resp.Stats = append(resp.Stats, engineStatsInfo)
		return nil
	})
	__antithesis_instrumentation__.Notify(236949)
	if err != nil {
		__antithesis_instrumentation__.Notify(236961)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(236962)
	}
	__antithesis_instrumentation__.Notify(236950)
	return resp, nil
}

func (s *statusServer) Allocator(
	ctx context.Context, req *serverpb.AllocatorRequest,
) (*serverpb.AllocatorResponse, error) {
	__antithesis_instrumentation__.Notify(236963)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(236969)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(236970)
	}
	__antithesis_instrumentation__.Notify(236964)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(236971)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(236972)
	}
	__antithesis_instrumentation__.Notify(236965)

	if !local {
		__antithesis_instrumentation__.Notify(236973)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(236975)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(236976)
		}
		__antithesis_instrumentation__.Notify(236974)
		return status.Allocator(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(236977)
	}
	__antithesis_instrumentation__.Notify(236966)

	output := new(serverpb.AllocatorResponse)
	err = s.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(236978)

		if len(req.RangeIDs) == 0 {
			__antithesis_instrumentation__.Notify(236981)
			var err error
			store.VisitReplicas(
				func(rep *kvserver.Replica) bool {
					__antithesis_instrumentation__.Notify(236983)
					if !rep.OwnsValidLease(ctx, store.Clock().NowAsClockTimestamp()) {
						__antithesis_instrumentation__.Notify(236986)
						return true
					} else {
						__antithesis_instrumentation__.Notify(236987)
					}
					__antithesis_instrumentation__.Notify(236984)
					var allocatorSpans tracing.Recording
					allocatorSpans, err = store.AllocatorDryRun(ctx, rep)
					if err != nil {
						__antithesis_instrumentation__.Notify(236988)
						return false
					} else {
						__antithesis_instrumentation__.Notify(236989)
					}
					__antithesis_instrumentation__.Notify(236985)
					output.DryRuns = append(output.DryRuns, &serverpb.AllocatorDryRun{
						RangeID: rep.RangeID,
						Events:  recordedSpansToTraceEvents(allocatorSpans),
					})
					return true
				},
				kvserver.WithReplicasInOrder(),
			)
			__antithesis_instrumentation__.Notify(236982)
			return err
		} else {
			__antithesis_instrumentation__.Notify(236990)
		}
		__antithesis_instrumentation__.Notify(236979)

		for _, rid := range req.RangeIDs {
			__antithesis_instrumentation__.Notify(236991)
			rep, err := store.GetReplica(rid)
			if err != nil {
				__antithesis_instrumentation__.Notify(236995)

				continue
			} else {
				__antithesis_instrumentation__.Notify(236996)
			}
			__antithesis_instrumentation__.Notify(236992)
			if !rep.OwnsValidLease(ctx, store.Clock().NowAsClockTimestamp()) {
				__antithesis_instrumentation__.Notify(236997)
				continue
			} else {
				__antithesis_instrumentation__.Notify(236998)
			}
			__antithesis_instrumentation__.Notify(236993)
			allocatorSpans, err := store.AllocatorDryRun(ctx, rep)
			if err != nil {
				__antithesis_instrumentation__.Notify(236999)
				return err
			} else {
				__antithesis_instrumentation__.Notify(237000)
			}
			__antithesis_instrumentation__.Notify(236994)
			output.DryRuns = append(output.DryRuns, &serverpb.AllocatorDryRun{
				RangeID: rep.RangeID,
				Events:  recordedSpansToTraceEvents(allocatorSpans),
			})
		}
		__antithesis_instrumentation__.Notify(236980)
		return nil
	})
	__antithesis_instrumentation__.Notify(236967)
	if err != nil {
		__antithesis_instrumentation__.Notify(237001)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237002)
	}
	__antithesis_instrumentation__.Notify(236968)
	return output, nil
}

func recordedSpansToTraceEvents(spans []tracingpb.RecordedSpan) []*serverpb.TraceEvent {
	__antithesis_instrumentation__.Notify(237003)
	var output []*serverpb.TraceEvent
	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(237005)
		for _, entry := range sp.Logs {
			__antithesis_instrumentation__.Notify(237006)
			event := &serverpb.TraceEvent{
				Time:    entry.Time,
				Message: entry.Msg().StripMarkers(),
			}
			output = append(output, event)
		}
	}
	__antithesis_instrumentation__.Notify(237004)
	return output
}

func (s *statusServer) AllocatorRange(
	ctx context.Context, req *serverpb.AllocatorRangeRequest,
) (*serverpb.AllocatorRangeResponse, error) {
	__antithesis_instrumentation__.Notify(237007)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237012)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237013)
	}
	__antithesis_instrumentation__.Notify(237008)

	isLiveMap := s.nodeLiveness.GetIsLiveMap()
	type nodeResponse struct {
		nodeID roachpb.NodeID
		resp   *serverpb.AllocatorResponse
		err    error
	}
	responses := make(chan nodeResponse)

	for nodeID := range isLiveMap {
		__antithesis_instrumentation__.Notify(237014)
		nodeID := nodeID
		if err := s.stopper.RunAsyncTask(
			ctx,
			"server.statusServer: requesting remote Allocator simulation",
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(237015)
				_ = contextutil.RunWithTimeout(ctx, "allocator range", base.NetworkTimeout, func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(237016)
					status, err := s.dialNode(ctx, nodeID)
					var allocatorResponse *serverpb.AllocatorResponse
					if err == nil {
						__antithesis_instrumentation__.Notify(237019)
						allocatorRequest := &serverpb.AllocatorRequest{
							RangeIDs: []roachpb.RangeID{roachpb.RangeID(req.RangeId)},
						}
						allocatorResponse, err = status.Allocator(ctx, allocatorRequest)
					} else {
						__antithesis_instrumentation__.Notify(237020)
					}
					__antithesis_instrumentation__.Notify(237017)
					response := nodeResponse{
						nodeID: nodeID,
						resp:   allocatorResponse,
						err:    err,
					}

					select {
					case responses <- response:
						__antithesis_instrumentation__.Notify(237021)

					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(237022)

					}
					__antithesis_instrumentation__.Notify(237018)
					return nil
				})
			}); err != nil {
			__antithesis_instrumentation__.Notify(237023)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237024)
		}
	}
	__antithesis_instrumentation__.Notify(237009)

	errs := make(map[roachpb.NodeID]error)
	for remainingResponses := len(isLiveMap); remainingResponses > 0; remainingResponses-- {
		__antithesis_instrumentation__.Notify(237025)
		select {
		case resp := <-responses:
			__antithesis_instrumentation__.Notify(237026)
			if resp.err != nil {
				__antithesis_instrumentation__.Notify(237029)
				errs[resp.nodeID] = resp.err
				continue
			} else {
				__antithesis_instrumentation__.Notify(237030)
			}
			__antithesis_instrumentation__.Notify(237027)
			if len(resp.resp.DryRuns) > 0 {
				__antithesis_instrumentation__.Notify(237031)
				return &serverpb.AllocatorRangeResponse{
					NodeID: resp.nodeID,
					DryRun: resp.resp.DryRuns[0],
				}, nil
			} else {
				__antithesis_instrumentation__.Notify(237032)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(237028)
			return nil, status.Errorf(codes.DeadlineExceeded, "request timed out")
		}
	}
	__antithesis_instrumentation__.Notify(237010)

	if len(errs) > 0 {
		__antithesis_instrumentation__.Notify(237033)
		var buf bytes.Buffer
		for nodeID, err := range errs {
			__antithesis_instrumentation__.Notify(237035)
			if buf.Len() > 0 {
				__antithesis_instrumentation__.Notify(237037)
				buf.WriteByte('\n')
			} else {
				__antithesis_instrumentation__.Notify(237038)
			}
			__antithesis_instrumentation__.Notify(237036)
			fmt.Fprintf(&buf, "n%d: %s", nodeID, err)
		}
		__antithesis_instrumentation__.Notify(237034)
		return nil, serverErrorf(ctx, "%v", buf)
	} else {
		__antithesis_instrumentation__.Notify(237039)
	}
	__antithesis_instrumentation__.Notify(237011)
	return &serverpb.AllocatorRangeResponse{}, nil
}

func (s *statusServer) Certificates(
	ctx context.Context, req *serverpb.CertificatesRequest,
) (*serverpb.CertificatesResponse, error) {
	__antithesis_instrumentation__.Notify(237040)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237048)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237049)
	}
	__antithesis_instrumentation__.Notify(237041)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237050)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237051)
	}
	__antithesis_instrumentation__.Notify(237042)

	if s.cfg.Insecure {
		__antithesis_instrumentation__.Notify(237052)
		return nil, status.Errorf(codes.Unavailable, "server is in insecure mode, cannot examine certificates")
	} else {
		__antithesis_instrumentation__.Notify(237053)
	}
	__antithesis_instrumentation__.Notify(237043)

	if !local {
		__antithesis_instrumentation__.Notify(237054)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237056)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237057)
		}
		__antithesis_instrumentation__.Notify(237055)
		return status.Certificates(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237058)
	}
	__antithesis_instrumentation__.Notify(237044)

	cm, err := s.rpcCtx.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(237059)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237060)
	}
	__antithesis_instrumentation__.Notify(237045)

	certs, err := cm.ListCertificates()
	if err != nil {
		__antithesis_instrumentation__.Notify(237061)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237062)
	}
	__antithesis_instrumentation__.Notify(237046)

	cr := &serverpb.CertificatesResponse{}
	for _, cert := range certs {
		__antithesis_instrumentation__.Notify(237063)
		details := serverpb.CertificateDetails{}
		switch cert.FileUsage {
		case security.CAPem:
			__antithesis_instrumentation__.Notify(237066)
			details.Type = serverpb.CertificateDetails_CA
		case security.ClientCAPem:
			__antithesis_instrumentation__.Notify(237067)
			details.Type = serverpb.CertificateDetails_CLIENT_CA
		case security.UICAPem:
			__antithesis_instrumentation__.Notify(237068)
			details.Type = serverpb.CertificateDetails_UI_CA
		case security.NodePem:
			__antithesis_instrumentation__.Notify(237069)
			details.Type = serverpb.CertificateDetails_NODE
		case security.UIPem:
			__antithesis_instrumentation__.Notify(237070)
			details.Type = serverpb.CertificateDetails_UI
		case security.ClientPem:
			__antithesis_instrumentation__.Notify(237071)
			details.Type = serverpb.CertificateDetails_CLIENT
		default:
			__antithesis_instrumentation__.Notify(237072)
			return nil, serverErrorf(ctx, "unknown certificate type %v for file %s", cert.FileUsage, cert.Filename)
		}
		__antithesis_instrumentation__.Notify(237064)

		if cert.Error == nil {
			__antithesis_instrumentation__.Notify(237073)
			details.Data = cert.FileContents
			if err := extractCertFields(details.Data, &details); err != nil {
				__antithesis_instrumentation__.Notify(237074)
				details.ErrorMessage = err.Error()
			} else {
				__antithesis_instrumentation__.Notify(237075)
			}
		} else {
			__antithesis_instrumentation__.Notify(237076)
			details.ErrorMessage = cert.Error.Error()
		}
		__antithesis_instrumentation__.Notify(237065)
		cr.Certificates = append(cr.Certificates, details)
	}
	__antithesis_instrumentation__.Notify(237047)

	return cr, nil
}

func formatCertNames(p pkix.Name) string {
	__antithesis_instrumentation__.Notify(237077)
	return fmt.Sprintf("CommonName=%s, Organization=%s", p.CommonName, strings.Join(p.Organization, ","))
}

func extractCertFields(contents []byte, details *serverpb.CertificateDetails) error {
	__antithesis_instrumentation__.Notify(237078)
	certs, err := security.PEMContentsToX509(contents)
	if err != nil {
		__antithesis_instrumentation__.Notify(237081)
		return err
	} else {
		__antithesis_instrumentation__.Notify(237082)
	}
	__antithesis_instrumentation__.Notify(237079)

	for _, c := range certs {
		__antithesis_instrumentation__.Notify(237083)
		addresses := c.DNSNames
		for _, ip := range c.IPAddresses {
			__antithesis_instrumentation__.Notify(237087)
			addresses = append(addresses, ip.String())
		}
		__antithesis_instrumentation__.Notify(237084)

		extKeyUsage := make([]string, len(c.ExtKeyUsage))
		for i, eku := range c.ExtKeyUsage {
			__antithesis_instrumentation__.Notify(237088)
			extKeyUsage[i] = security.ExtKeyUsageToString(eku)
		}
		__antithesis_instrumentation__.Notify(237085)

		var pubKeyInfo string
		if rsaPub, ok := c.PublicKey.(*rsa.PublicKey); ok {
			__antithesis_instrumentation__.Notify(237089)
			pubKeyInfo = fmt.Sprintf("%d bit RSA", rsaPub.N.BitLen())
		} else {
			__antithesis_instrumentation__.Notify(237090)
			if ecdsaPub, ok := c.PublicKey.(*ecdsa.PublicKey); ok {
				__antithesis_instrumentation__.Notify(237091)
				pubKeyInfo = fmt.Sprintf("%d bit ECDSA", ecdsaPub.Params().BitSize)
			} else {
				__antithesis_instrumentation__.Notify(237092)

				pubKeyInfo = fmt.Sprintf("unknown key type %T", c.PublicKey)
			}
		}
		__antithesis_instrumentation__.Notify(237086)

		details.Fields = append(details.Fields, serverpb.CertificateDetails_Fields{
			Issuer:             formatCertNames(c.Issuer),
			Subject:            formatCertNames(c.Subject),
			ValidFrom:          c.NotBefore.UnixNano(),
			ValidUntil:         c.NotAfter.UnixNano(),
			Addresses:          addresses,
			SignatureAlgorithm: c.SignatureAlgorithm.String(),
			PublicKey:          pubKeyInfo,
			KeyUsage:           security.KeyUsageToString(c.KeyUsage),
			ExtendedKeyUsage:   extKeyUsage,
		})
	}
	__antithesis_instrumentation__.Notify(237080)
	return nil
}

func (s *statusServer) Details(
	ctx context.Context, req *serverpb.DetailsRequest,
) (*serverpb.DetailsResponse, error) {
	__antithesis_instrumentation__.Notify(237093)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237099)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237100)
	}
	__antithesis_instrumentation__.Notify(237094)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237101)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237102)
	}
	__antithesis_instrumentation__.Notify(237095)
	if !local {
		__antithesis_instrumentation__.Notify(237103)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237105)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237106)
		}
		__antithesis_instrumentation__.Notify(237104)
		return status.Details(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237107)
	}
	__antithesis_instrumentation__.Notify(237096)

	remoteNodeID := s.gossip.NodeID.Get()
	resp := &serverpb.DetailsResponse{
		NodeID:     remoteNodeID,
		BuildInfo:  build.GetInfo(),
		SystemInfo: s.si.systemInfo(ctx),
	}
	if addr, err := s.gossip.GetNodeIDAddress(remoteNodeID); err == nil {
		__antithesis_instrumentation__.Notify(237108)
		resp.Address = *addr
	} else {
		__antithesis_instrumentation__.Notify(237109)
	}
	__antithesis_instrumentation__.Notify(237097)
	if addr, err := s.gossip.GetNodeIDSQLAddress(remoteNodeID); err == nil {
		__antithesis_instrumentation__.Notify(237110)
		resp.SQLAddress = *addr
	} else {
		__antithesis_instrumentation__.Notify(237111)
	}
	__antithesis_instrumentation__.Notify(237098)

	return resp, nil
}

func (s *statusServer) GetFiles(
	ctx context.Context, req *serverpb.GetFilesRequest,
) (*serverpb.GetFilesResponse, error) {
	__antithesis_instrumentation__.Notify(237112)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237116)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237117)
	}
	__antithesis_instrumentation__.Notify(237113)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237118)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237119)
	}
	__antithesis_instrumentation__.Notify(237114)
	if !local {
		__antithesis_instrumentation__.Notify(237120)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237122)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237123)
		}
		__antithesis_instrumentation__.Notify(237121)
		return status.GetFiles(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237124)
	}
	__antithesis_instrumentation__.Notify(237115)

	return getLocalFiles(req, s.sqlServer.cfg.HeapProfileDirName, s.sqlServer.cfg.GoroutineDumpDirName)
}

func checkFilePattern(pattern string) error {
	__antithesis_instrumentation__.Notify(237125)
	if strings.Contains(pattern, string(os.PathSeparator)) {
		__antithesis_instrumentation__.Notify(237127)
		return errors.New("invalid pattern: cannot have path seperators")
	} else {
		__antithesis_instrumentation__.Notify(237128)
	}
	__antithesis_instrumentation__.Notify(237126)
	return nil
}

func (s *statusServer) LogFilesList(
	ctx context.Context, req *serverpb.LogFilesListRequest,
) (*serverpb.LogFilesListResponse, error) {
	__antithesis_instrumentation__.Notify(237129)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237134)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237135)
	}
	__antithesis_instrumentation__.Notify(237130)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237136)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237137)
	}
	__antithesis_instrumentation__.Notify(237131)
	if !local {
		__antithesis_instrumentation__.Notify(237138)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237140)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237141)
		}
		__antithesis_instrumentation__.Notify(237139)
		return status.LogFilesList(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237142)
	}
	__antithesis_instrumentation__.Notify(237132)
	log.Flush()
	logFiles, err := log.ListLogFiles()
	if err != nil {
		__antithesis_instrumentation__.Notify(237143)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237144)
	}
	__antithesis_instrumentation__.Notify(237133)
	return &serverpb.LogFilesListResponse{Files: logFiles}, nil
}

func (s *statusServer) LogFile(
	ctx context.Context, req *serverpb.LogFileRequest,
) (*serverpb.LogEntriesResponse, error) {
	__antithesis_instrumentation__.Notify(237145)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237152)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237153)
	}
	__antithesis_instrumentation__.Notify(237146)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237154)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237155)
	}
	__antithesis_instrumentation__.Notify(237147)
	if !local {
		__antithesis_instrumentation__.Notify(237156)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237158)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237159)
		}
		__antithesis_instrumentation__.Notify(237157)
		return status.LogFile(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237160)
	}
	__antithesis_instrumentation__.Notify(237148)

	inputEditMode := log.SelectEditMode(req.Redact, log.KeepRedactable)

	log.Flush()

	reader, err := log.GetLogReader(req.File)
	if err != nil {
		__antithesis_instrumentation__.Notify(237161)
		return nil, serverError(ctx, errors.Wrapf(err, "log file %q could not be opened", req.File))
	} else {
		__antithesis_instrumentation__.Notify(237162)
	}
	__antithesis_instrumentation__.Notify(237149)
	defer reader.Close()

	var resp serverpb.LogEntriesResponse
	decoder, err := log.NewEntryDecoder(reader, inputEditMode)
	if err != nil {
		__antithesis_instrumentation__.Notify(237163)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237164)
	}
	__antithesis_instrumentation__.Notify(237150)
	for {
		__antithesis_instrumentation__.Notify(237165)
		var entry logpb.Entry
		if err := decoder.Decode(&entry); err != nil {
			__antithesis_instrumentation__.Notify(237167)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(237169)
				break
			} else {
				__antithesis_instrumentation__.Notify(237170)
			}
			__antithesis_instrumentation__.Notify(237168)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237171)
		}
		__antithesis_instrumentation__.Notify(237166)
		resp.Entries = append(resp.Entries, entry)
	}
	__antithesis_instrumentation__.Notify(237151)

	return &resp, nil
}

func parseInt64WithDefault(s string, defaultValue int64) (int64, error) {
	__antithesis_instrumentation__.Notify(237172)
	if len(s) == 0 {
		__antithesis_instrumentation__.Notify(237175)
		return defaultValue, nil
	} else {
		__antithesis_instrumentation__.Notify(237176)
	}
	__antithesis_instrumentation__.Notify(237173)
	result, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(237177)
		return defaultValue, err
	} else {
		__antithesis_instrumentation__.Notify(237178)
	}
	__antithesis_instrumentation__.Notify(237174)
	return result, nil
}

func (s *statusServer) Logs(
	ctx context.Context, req *serverpb.LogsRequest,
) (*serverpb.LogEntriesResponse, error) {
	__antithesis_instrumentation__.Notify(237179)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237190)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237191)
	}
	__antithesis_instrumentation__.Notify(237180)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237192)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237193)
	}
	__antithesis_instrumentation__.Notify(237181)
	if !local {
		__antithesis_instrumentation__.Notify(237194)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237196)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237197)
		}
		__antithesis_instrumentation__.Notify(237195)
		return status.Logs(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237198)
	}
	__antithesis_instrumentation__.Notify(237182)

	inputEditMode := log.SelectEditMode(req.Redact, log.KeepRedactable)

	startTimestamp, err := parseInt64WithDefault(
		req.StartTime,
		timeutil.Now().AddDate(0, 0, -1).UnixNano())
	if err != nil {
		__antithesis_instrumentation__.Notify(237199)
		return nil, status.Errorf(codes.InvalidArgument, "StartTime could not be parsed: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(237200)
	}
	__antithesis_instrumentation__.Notify(237183)

	endTimestamp, err := parseInt64WithDefault(req.EndTime, timeutil.Now().UnixNano())
	if err != nil {
		__antithesis_instrumentation__.Notify(237201)
		return nil, status.Errorf(codes.InvalidArgument, "EndTime could not be parsed: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(237202)
	}
	__antithesis_instrumentation__.Notify(237184)

	if startTimestamp > endTimestamp {
		__antithesis_instrumentation__.Notify(237203)
		return nil, status.Errorf(codes.InvalidArgument, "StartTime: %d should not be greater than endtime: %d", startTimestamp, endTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(237204)
	}
	__antithesis_instrumentation__.Notify(237185)

	maxEntries, err := parseInt64WithDefault(req.Max, defaultMaxLogEntries)
	if err != nil {
		__antithesis_instrumentation__.Notify(237205)
		return nil, status.Errorf(codes.InvalidArgument, "Max could not be parsed: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(237206)
	}
	__antithesis_instrumentation__.Notify(237186)
	if maxEntries < 1 {
		__antithesis_instrumentation__.Notify(237207)
		return nil, status.Errorf(codes.InvalidArgument, "Max: %d should be set to a value greater than 0", maxEntries)
	} else {
		__antithesis_instrumentation__.Notify(237208)
	}
	__antithesis_instrumentation__.Notify(237187)

	var regex *regexp.Regexp
	if len(req.Pattern) > 0 {
		__antithesis_instrumentation__.Notify(237209)
		if regex, err = regexp.Compile(req.Pattern); err != nil {
			__antithesis_instrumentation__.Notify(237210)
			return nil, status.Errorf(codes.InvalidArgument, "regex pattern could not be compiled: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(237211)
		}
	} else {
		__antithesis_instrumentation__.Notify(237212)
	}
	__antithesis_instrumentation__.Notify(237188)

	log.Flush()

	entries, err := log.FetchEntriesFromFiles(
		startTimestamp, endTimestamp, int(maxEntries), regex, inputEditMode)
	if err != nil {
		__antithesis_instrumentation__.Notify(237213)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237214)
	}
	__antithesis_instrumentation__.Notify(237189)

	return &serverpb.LogEntriesResponse{Entries: entries}, nil
}

func (s *statusServer) Stacks(
	ctx context.Context, req *serverpb.StacksRequest,
) (*serverpb.JSONResponse, error) {
	__antithesis_instrumentation__.Notify(237215)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237219)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237220)
	}
	__antithesis_instrumentation__.Notify(237216)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237221)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237222)
	}
	__antithesis_instrumentation__.Notify(237217)

	if !local {
		__antithesis_instrumentation__.Notify(237223)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237225)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237226)
		}
		__antithesis_instrumentation__.Notify(237224)
		return status.Stacks(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237227)
	}
	__antithesis_instrumentation__.Notify(237218)

	return stacksLocal(req)
}

func (s *statusServer) Profile(
	ctx context.Context, req *serverpb.ProfileRequest,
) (*serverpb.JSONResponse, error) {
	__antithesis_instrumentation__.Notify(237228)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237232)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237233)
	}
	__antithesis_instrumentation__.Notify(237229)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237234)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237235)
	}
	__antithesis_instrumentation__.Notify(237230)

	if !local {
		__antithesis_instrumentation__.Notify(237236)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237238)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237239)
		}
		__antithesis_instrumentation__.Notify(237237)
		return status.Profile(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237240)
	}
	__antithesis_instrumentation__.Notify(237231)

	return profileLocal(ctx, req, s.st)
}

func (s *statusServer) Regions(
	ctx context.Context, req *serverpb.RegionsRequest,
) (*serverpb.RegionsResponse, error) {
	__antithesis_instrumentation__.Notify(237241)
	resp, _, err := s.nodesHelper(ctx, 0, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(237243)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237244)
	}
	__antithesis_instrumentation__.Notify(237242)
	return regionsResponseFromNodesResponse(resp), nil
}

func regionsResponseFromNodesResponse(nr *serverpb.NodesResponse) *serverpb.RegionsResponse {
	__antithesis_instrumentation__.Notify(237245)
	regionsToZones := make(map[string]map[string]struct{})
	for _, node := range nr.Nodes {
		__antithesis_instrumentation__.Notify(237248)
		var region string
		var zone string
		for _, tier := range node.Desc.Locality.Tiers {
			__antithesis_instrumentation__.Notify(237252)
			switch tier.Key {
			case "region":
				__antithesis_instrumentation__.Notify(237253)
				region = tier.Value
			case "zone", "availability-zone", "az":
				__antithesis_instrumentation__.Notify(237254)
				zone = tier.Value
			default:
				__antithesis_instrumentation__.Notify(237255)
			}
		}
		__antithesis_instrumentation__.Notify(237249)
		if region == "" {
			__antithesis_instrumentation__.Notify(237256)
			continue
		} else {
			__antithesis_instrumentation__.Notify(237257)
		}
		__antithesis_instrumentation__.Notify(237250)
		if _, ok := regionsToZones[region]; !ok {
			__antithesis_instrumentation__.Notify(237258)
			regionsToZones[region] = make(map[string]struct{})
		} else {
			__antithesis_instrumentation__.Notify(237259)
		}
		__antithesis_instrumentation__.Notify(237251)
		if zone != "" {
			__antithesis_instrumentation__.Notify(237260)
			regionsToZones[region][zone] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(237261)
		}
	}
	__antithesis_instrumentation__.Notify(237246)
	ret := &serverpb.RegionsResponse{
		Regions: make(map[string]*serverpb.RegionsResponse_Region, len(regionsToZones)),
	}
	for region, zones := range regionsToZones {
		__antithesis_instrumentation__.Notify(237262)
		zonesArr := make([]string, 0, len(zones))
		for z := range zones {
			__antithesis_instrumentation__.Notify(237264)
			zonesArr = append(zonesArr, z)
		}
		__antithesis_instrumentation__.Notify(237263)
		sort.Strings(zonesArr)
		ret.Regions[region] = &serverpb.RegionsResponse_Region{
			Zones: zonesArr,
		}
	}
	__antithesis_instrumentation__.Notify(237247)
	return ret
}

func (s *statusServer) NodesList(
	ctx context.Context, _ *serverpb.NodesListRequest,
) (*serverpb.NodesListResponse, error) {
	__antithesis_instrumentation__.Notify(237265)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237269)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237270)
	}
	__antithesis_instrumentation__.Notify(237266)
	statuses, _, err := s.getNodeStatuses(ctx, 0, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(237271)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237272)
	}
	__antithesis_instrumentation__.Notify(237267)
	resp := &serverpb.NodesListResponse{
		Nodes: make([]serverpb.NodeDetails, len(statuses)),
	}
	for i, status := range statuses {
		__antithesis_instrumentation__.Notify(237273)
		resp.Nodes[i].NodeID = int32(status.Desc.NodeID)
		resp.Nodes[i].Address = status.Desc.Address
		resp.Nodes[i].SQLAddress = status.Desc.SQLAddress
	}
	__antithesis_instrumentation__.Notify(237268)
	return resp, nil
}

func (s *statusServer) Nodes(
	ctx context.Context, req *serverpb.NodesRequest,
) (*serverpb.NodesResponse, error) {
	__antithesis_instrumentation__.Notify(237274)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	err := s.privilegeChecker.requireViewActivityPermission(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(237277)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237278)
	}
	__antithesis_instrumentation__.Notify(237275)

	resp, _, err := s.nodesHelper(ctx, 0, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(237279)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237280)
	}
	__antithesis_instrumentation__.Notify(237276)
	return resp, nil
}

func (s *statusServer) NodesUI(
	ctx context.Context, req *serverpb.NodesRequest,
) (*serverpb.NodesResponseExternal, error) {
	__antithesis_instrumentation__.Notify(237281)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	hasViewActivity := false
	err := s.privilegeChecker.requireViewActivityPermission(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(237285)
		if !grpcutil.IsAuthError(err) {
			__antithesis_instrumentation__.Notify(237286)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(237287)
		}
	} else {
		__antithesis_instrumentation__.Notify(237288)
		hasViewActivity = true
	}
	__antithesis_instrumentation__.Notify(237282)

	internalResp, _, err := s.nodesHelper(ctx, 0, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(237289)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237290)
	}
	__antithesis_instrumentation__.Notify(237283)
	resp := &serverpb.NodesResponseExternal{
		Nodes:            make([]serverpb.NodeResponse, len(internalResp.Nodes)),
		LivenessByNodeID: internalResp.LivenessByNodeID,
	}
	for i, nodeStatus := range internalResp.Nodes {
		__antithesis_instrumentation__.Notify(237291)
		resp.Nodes[i] = nodeStatusToResp(&nodeStatus, hasViewActivity)
	}
	__antithesis_instrumentation__.Notify(237284)

	return resp, nil
}

func nodeStatusToResp(n *statuspb.NodeStatus, hasViewActivity bool) serverpb.NodeResponse {
	__antithesis_instrumentation__.Notify(237292)
	tiers := make([]serverpb.Tier, len(n.Desc.Locality.Tiers))
	for j, t := range n.Desc.Locality.Tiers {
		__antithesis_instrumentation__.Notify(237297)
		tiers[j] = serverpb.Tier{
			Key:   t.Key,
			Value: t.Value,
		}
	}
	__antithesis_instrumentation__.Notify(237293)

	activity := make(map[roachpb.NodeID]serverpb.NodeResponse_NetworkActivity, len(n.Activity))
	for k, v := range n.Activity {
		__antithesis_instrumentation__.Notify(237298)
		activity[k] = serverpb.NodeResponse_NetworkActivity{
			Incoming: v.Incoming,
			Outgoing: v.Outgoing,
			Latency:  v.Latency,
		}
	}
	__antithesis_instrumentation__.Notify(237294)

	nodeDescriptor := serverpb.NodeDescriptor{
		NodeID:  n.Desc.NodeID,
		Address: util.UnresolvedAddr{},
		Attrs:   roachpb.Attributes{},
		Locality: serverpb.Locality{
			Tiers: tiers,
		},
		ServerVersion: serverpb.Version{
			Major:    n.Desc.ServerVersion.Major,
			Minor:    n.Desc.ServerVersion.Minor,
			Patch:    n.Desc.ServerVersion.Patch,
			Internal: n.Desc.ServerVersion.Internal,
		},
		BuildTag:        n.Desc.BuildTag,
		StartedAt:       n.Desc.StartedAt,
		LocalityAddress: nil,
		ClusterName:     n.Desc.ClusterName,
		SQLAddress:      util.UnresolvedAddr{},
	}

	statuses := make([]serverpb.StoreStatus, len(n.StoreStatuses))
	for i, ss := range n.StoreStatuses {
		__antithesis_instrumentation__.Notify(237299)
		statuses[i] = serverpb.StoreStatus{
			Desc: serverpb.StoreDescriptor{
				StoreID:  ss.Desc.StoreID,
				Attrs:    ss.Desc.Attrs,
				Node:     nodeDescriptor,
				Capacity: ss.Desc.Capacity,

				Properties: roachpb.StoreProperties{
					ReadOnly:  ss.Desc.Properties.ReadOnly,
					Encrypted: ss.Desc.Properties.Encrypted,
				},
			},
			Metrics: ss.Metrics,
		}
		if fsprops := ss.Desc.Properties.FileStoreProperties; fsprops != nil {
			__antithesis_instrumentation__.Notify(237300)
			sfsprops := &roachpb.FileStoreProperties{
				FsType: fsprops.FsType,
			}
			if hasViewActivity {
				__antithesis_instrumentation__.Notify(237302)
				sfsprops.Path = fsprops.Path
				sfsprops.BlockDevice = fsprops.BlockDevice
				sfsprops.MountPoint = fsprops.MountPoint
				sfsprops.MountOptions = fsprops.MountOptions
			} else {
				__antithesis_instrumentation__.Notify(237303)
			}
			__antithesis_instrumentation__.Notify(237301)
			statuses[i].Desc.Properties.FileStoreProperties = sfsprops
		} else {
			__antithesis_instrumentation__.Notify(237304)
		}
	}
	__antithesis_instrumentation__.Notify(237295)

	resp := serverpb.NodeResponse{
		Desc:              nodeDescriptor,
		BuildInfo:         n.BuildInfo,
		StartedAt:         n.StartedAt,
		UpdatedAt:         n.UpdatedAt,
		Metrics:           n.Metrics,
		StoreStatuses:     statuses,
		Args:              nil,
		Env:               nil,
		Latencies:         n.Latencies,
		Activity:          activity,
		TotalSystemMemory: n.TotalSystemMemory,
		NumCpus:           n.NumCpus,
	}

	if hasViewActivity {
		__antithesis_instrumentation__.Notify(237305)
		resp.Args = n.Args
		resp.Env = n.Env
		resp.Desc.Attrs = n.Desc.Attrs
		resp.Desc.Address = n.Desc.Address
		resp.Desc.LocalityAddress = n.Desc.LocalityAddress
		resp.Desc.SQLAddress = n.Desc.SQLAddress
		for _, n := range resp.StoreStatuses {
			__antithesis_instrumentation__.Notify(237306)
			n.Desc.Node = resp.Desc
		}
	} else {
		__antithesis_instrumentation__.Notify(237307)
	}
	__antithesis_instrumentation__.Notify(237296)

	return resp
}

func (s *statusServer) ListNodesInternal(
	ctx context.Context, req *serverpb.NodesRequest,
) (*serverpb.NodesResponse, error) {
	__antithesis_instrumentation__.Notify(237308)
	resp, _, err := s.nodesHelper(ctx, 0, 0)
	return resp, err
}

func (s *statusServer) getNodeStatuses(
	ctx context.Context, limit, offset int,
) (statuses []statuspb.NodeStatus, next int, _ error) {
	__antithesis_instrumentation__.Notify(237309)
	startKey := keys.StatusNodePrefix
	endKey := startKey.PrefixEnd()

	b := &kv.Batch{}
	b.Scan(startKey, endKey)
	if err := s.db.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(237313)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(237314)
	}
	__antithesis_instrumentation__.Notify(237310)

	var rows []kv.KeyValue
	if len(b.Results[0].Rows) > 0 {
		__antithesis_instrumentation__.Notify(237315)
		var rowsInterface interface{}
		rowsInterface, next = simplePaginate(b.Results[0].Rows, limit, offset)
		rows = rowsInterface.([]kv.KeyValue)
	} else {
		__antithesis_instrumentation__.Notify(237316)
	}
	__antithesis_instrumentation__.Notify(237311)

	statuses = make([]statuspb.NodeStatus, len(rows))
	for i, row := range rows {
		__antithesis_instrumentation__.Notify(237317)
		if err := row.ValueProto(&statuses[i]); err != nil {
			__antithesis_instrumentation__.Notify(237318)
			return nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(237319)
		}
	}
	__antithesis_instrumentation__.Notify(237312)
	return statuses, next, nil
}

func (s *statusServer) nodesHelper(
	ctx context.Context, limit, offset int,
) (*serverpb.NodesResponse, int, error) {
	__antithesis_instrumentation__.Notify(237320)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	statuses, next, err := s.getNodeStatuses(ctx, limit, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(237322)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(237323)
	}
	__antithesis_instrumentation__.Notify(237321)
	resp := serverpb.NodesResponse{
		Nodes: statuses,
	}

	clock := s.admin.server.clock
	resp.LivenessByNodeID = getLivenessStatusMap(s.nodeLiveness, clock.Now().GoTime(), s.st)
	return &resp, next, nil
}

func (s *statusServer) nodesStatusWithLiveness(
	ctx context.Context,
) (map[roachpb.NodeID]nodeStatusWithLiveness, error) {
	__antithesis_instrumentation__.Notify(237324)
	nodes, err := s.ListNodesInternal(ctx, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(237327)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237328)
	}
	__antithesis_instrumentation__.Notify(237325)
	clock := s.admin.server.clock
	statusMap := getLivenessStatusMap(s.nodeLiveness, clock.Now().GoTime(), s.st)
	ret := make(map[roachpb.NodeID]nodeStatusWithLiveness)
	for _, node := range nodes.Nodes {
		__antithesis_instrumentation__.Notify(237329)
		nodeID := node.Desc.NodeID
		livenessStatus := statusMap[nodeID]
		if livenessStatus == livenesspb.NodeLivenessStatus_DECOMMISSIONED {
			__antithesis_instrumentation__.Notify(237331)

			continue
		} else {
			__antithesis_instrumentation__.Notify(237332)
		}
		__antithesis_instrumentation__.Notify(237330)
		ret[nodeID] = nodeStatusWithLiveness{
			NodeStatus:     node,
			livenessStatus: livenessStatus,
		}
	}
	__antithesis_instrumentation__.Notify(237326)
	return ret, nil
}

type nodeStatusWithLiveness struct {
	statuspb.NodeStatus
	livenessStatus livenesspb.NodeLivenessStatus
}

func (s *statusServer) Node(
	ctx context.Context, req *serverpb.NodeRequest,
) (*statuspb.NodeStatus, error) {
	__antithesis_instrumentation__.Notify(237333)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237335)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237336)
	}
	__antithesis_instrumentation__.Notify(237334)

	return s.nodeStatus(ctx, req)
}

func (s *statusServer) nodeStatus(
	ctx context.Context, req *serverpb.NodeRequest,
) (*statuspb.NodeStatus, error) {
	__antithesis_instrumentation__.Notify(237337)
	nodeID, _, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237341)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237342)
	}
	__antithesis_instrumentation__.Notify(237338)

	key := keys.NodeStatusKey(nodeID)
	b := &kv.Batch{}
	b.Get(key)
	if err := s.db.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(237343)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237344)
	}
	__antithesis_instrumentation__.Notify(237339)

	var nodeStatus statuspb.NodeStatus
	if err := b.Results[0].Rows[0].ValueProto(&nodeStatus); err != nil {
		__antithesis_instrumentation__.Notify(237345)
		err = errors.Wrapf(err, "could not unmarshal NodeStatus from %s", key)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237346)
	}
	__antithesis_instrumentation__.Notify(237340)
	return &nodeStatus, nil
}

func (s *statusServer) NodeUI(
	ctx context.Context, req *serverpb.NodeRequest,
) (*serverpb.NodeResponse, error) {
	__antithesis_instrumentation__.Notify(237347)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	_, isAdmin, err := s.privilegeChecker.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(237350)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237351)
	}
	__antithesis_instrumentation__.Notify(237348)

	nodeStatus, err := s.nodeStatus(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(237352)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237353)
	}
	__antithesis_instrumentation__.Notify(237349)
	resp := nodeStatusToResp(nodeStatus, isAdmin)
	return &resp, nil
}

func (s *statusServer) Metrics(
	ctx context.Context, req *serverpb.MetricsRequest,
) (*serverpb.JSONResponse, error) {
	__antithesis_instrumentation__.Notify(237354)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237358)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237359)
	}
	__antithesis_instrumentation__.Notify(237355)

	if !local {
		__antithesis_instrumentation__.Notify(237360)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237362)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237363)
		}
		__antithesis_instrumentation__.Notify(237361)
		return status.Metrics(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237364)
	}
	__antithesis_instrumentation__.Notify(237356)
	j, err := marshalJSONResponse(s.metricSource)
	if err != nil {
		__antithesis_instrumentation__.Notify(237365)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237366)
	}
	__antithesis_instrumentation__.Notify(237357)
	return j, nil
}

func (s *statusServer) RaftDebug(
	ctx context.Context, req *serverpb.RaftDebugRequest,
) (*serverpb.RaftDebugResponse, error) {
	__antithesis_instrumentation__.Notify(237367)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237372)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237373)
	}
	__antithesis_instrumentation__.Notify(237368)

	nodes, err := s.ListNodesInternal(ctx, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(237374)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237375)
	}
	__antithesis_instrumentation__.Notify(237369)

	mu := struct {
		syncutil.Mutex
		resp serverpb.RaftDebugResponse
	}{
		resp: serverpb.RaftDebugResponse{
			Ranges: make(map[roachpb.RangeID]serverpb.RaftRangeStatus),
		},
	}

	var wg sync.WaitGroup
	for _, node := range nodes.Nodes {
		__antithesis_instrumentation__.Notify(237376)
		wg.Add(1)
		nodeID := node.Desc.NodeID
		go func() {
			__antithesis_instrumentation__.Notify(237377)
			defer wg.Done()
			ranges, err := s.Ranges(ctx, &serverpb.RangesRequest{NodeId: nodeID.String(), RangeIDs: req.RangeIDs})

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				__antithesis_instrumentation__.Notify(237379)
				err := errors.Wrapf(err, "failed to get ranges from %d", nodeID)
				mu.resp.Errors = append(mu.resp.Errors, serverpb.RaftRangeError{Message: err.Error()})
				return
			} else {
				__antithesis_instrumentation__.Notify(237380)
			}
			__antithesis_instrumentation__.Notify(237378)

			for _, rng := range ranges.Ranges {
				__antithesis_instrumentation__.Notify(237381)
				rangeID := rng.State.Desc.RangeID
				status, ok := mu.resp.Ranges[rangeID]
				if !ok {
					__antithesis_instrumentation__.Notify(237383)
					status = serverpb.RaftRangeStatus{
						RangeID: rangeID,
					}
				} else {
					__antithesis_instrumentation__.Notify(237384)
				}
				__antithesis_instrumentation__.Notify(237382)
				status.Nodes = append(status.Nodes, serverpb.RaftRangeNode{
					NodeID: nodeID,
					Range:  rng,
				})
				mu.resp.Ranges[rangeID] = status
			}
		}()
	}
	__antithesis_instrumentation__.Notify(237370)
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()

	for i, rng := range mu.resp.Ranges {
		__antithesis_instrumentation__.Notify(237385)
		for j, node := range rng.Nodes {
			__antithesis_instrumentation__.Notify(237386)
			desc := node.Range.State.Desc

			containsNode := false
			for _, replica := range desc.Replicas().Descriptors() {
				__antithesis_instrumentation__.Notify(237390)
				if replica.NodeID == node.NodeID {
					__antithesis_instrumentation__.Notify(237391)
					containsNode = true
				} else {
					__antithesis_instrumentation__.Notify(237392)
				}
			}
			__antithesis_instrumentation__.Notify(237387)
			if !containsNode {
				__antithesis_instrumentation__.Notify(237393)
				rng.Errors = append(rng.Errors, serverpb.RaftRangeError{
					Message: fmt.Sprintf("node %d not in range descriptor and should be GCed", node.NodeID),
				})
			} else {
				__antithesis_instrumentation__.Notify(237394)
			}
			__antithesis_instrumentation__.Notify(237388)

			if j > 0 {
				__antithesis_instrumentation__.Notify(237395)
				prevDesc := rng.Nodes[j-1].Range.State.Desc
				if !desc.Equal(prevDesc) {
					__antithesis_instrumentation__.Notify(237396)
					prevNodeID := rng.Nodes[j-1].NodeID
					rng.Errors = append(rng.Errors, serverpb.RaftRangeError{
						Message: fmt.Sprintf("node %d range descriptor does not match node %d", node.NodeID, prevNodeID),
					})
				} else {
					__antithesis_instrumentation__.Notify(237397)
				}
			} else {
				__antithesis_instrumentation__.Notify(237398)
			}
			__antithesis_instrumentation__.Notify(237389)
			mu.resp.Ranges[i] = rng
		}
	}
	__antithesis_instrumentation__.Notify(237371)
	return &mu.resp, nil
}

type varsHandler struct {
	metricSource metricMarshaler
	st           *cluster.Settings
}

func (h varsHandler) handleVars(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(237399)
	ctx := r.Context()

	w.Header().Set(httputil.ContentTypeHeader, httputil.PlaintextContentType)
	err := h.metricSource.PrintAsText(w)
	if err != nil {
		__antithesis_instrumentation__.Notify(237401)
		log.Errorf(ctx, "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		__antithesis_instrumentation__.Notify(237402)
	}
	__antithesis_instrumentation__.Notify(237400)

	telemetry.Inc(telemetryPrometheusVars)
}

func (s *statusServer) Ranges(
	ctx context.Context, req *serverpb.RangesRequest,
) (*serverpb.RangesResponse, error) {
	__antithesis_instrumentation__.Notify(237403)
	resp, _, err := s.rangesHelper(ctx, req, 0, 0)
	return resp, err
}

func (s *statusServer) rangesHelper(
	ctx context.Context, req *serverpb.RangesRequest, limit, offset int,
) (*serverpb.RangesResponse, int, error) {
	__antithesis_instrumentation__.Notify(237404)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237414)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(237415)
	}
	__antithesis_instrumentation__.Notify(237405)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237416)
		return nil, 0, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237417)
	}
	__antithesis_instrumentation__.Notify(237406)

	if !local {
		__antithesis_instrumentation__.Notify(237418)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237421)
			return nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(237422)
		}
		__antithesis_instrumentation__.Notify(237419)
		resp, err := status.Ranges(ctx, req)
		if resp != nil && func() bool {
			__antithesis_instrumentation__.Notify(237423)
			return len(resp.Ranges) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(237424)
			resultInterface, next := simplePaginate(resp.Ranges, limit, offset)
			resp.Ranges = resultInterface.([]serverpb.RangeInfo)
			return resp, next, err
		} else {
			__antithesis_instrumentation__.Notify(237425)
		}
		__antithesis_instrumentation__.Notify(237420)
		return resp, 0, err
	} else {
		__antithesis_instrumentation__.Notify(237426)
	}
	__antithesis_instrumentation__.Notify(237407)

	output := serverpb.RangesResponse{
		Ranges: make([]serverpb.RangeInfo, 0, s.stores.GetStoreCount()),
	}

	convertRaftStatus := func(raftStatus *raft.Status) serverpb.RaftState {
		__antithesis_instrumentation__.Notify(237427)
		if raftStatus == nil {
			__antithesis_instrumentation__.Notify(237430)
			return serverpb.RaftState{
				State: raftStateDormant,
			}
		} else {
			__antithesis_instrumentation__.Notify(237431)
		}
		__antithesis_instrumentation__.Notify(237428)

		state := serverpb.RaftState{
			ReplicaID:      raftStatus.ID,
			HardState:      raftStatus.HardState,
			Applied:        raftStatus.Applied,
			Lead:           raftStatus.Lead,
			State:          raftStatus.RaftState.String(),
			Progress:       make(map[uint64]serverpb.RaftState_Progress),
			LeadTransferee: raftStatus.LeadTransferee,
		}

		for id, progress := range raftStatus.Progress {
			__antithesis_instrumentation__.Notify(237432)
			state.Progress[id] = serverpb.RaftState_Progress{
				Match:           progress.Match,
				Next:            progress.Next,
				Paused:          progress.IsPaused(),
				PendingSnapshot: progress.PendingSnapshot,
				State:           progress.State.String(),
			}
		}
		__antithesis_instrumentation__.Notify(237429)

		return state
	}
	__antithesis_instrumentation__.Notify(237408)

	constructRangeInfo := func(
		rep *kvserver.Replica, storeID roachpb.StoreID, metrics kvserver.ReplicaMetrics,
	) serverpb.RangeInfo {
		__antithesis_instrumentation__.Notify(237433)
		raftStatus := rep.RaftStatus()
		raftState := convertRaftStatus(raftStatus)
		leaseHistory := rep.GetLeaseHistory()
		var span serverpb.PrettySpan
		desc := rep.Desc()
		span.StartKey = desc.StartKey.String()
		span.EndKey = desc.EndKey.String()
		state := rep.State(ctx)
		var topKLocksByWaiters []serverpb.RangeInfo_LockInfo
		for _, lm := range metrics.LockTableMetrics.TopKLocksByWaiters {
			__antithesis_instrumentation__.Notify(237436)
			if lm.Key == nil {
				__antithesis_instrumentation__.Notify(237438)
				break
			} else {
				__antithesis_instrumentation__.Notify(237439)
			}
			__antithesis_instrumentation__.Notify(237437)
			topKLocksByWaiters = append(topKLocksByWaiters, serverpb.RangeInfo_LockInfo{
				PrettyKey:      lm.Key.String(),
				Key:            lm.Key,
				Held:           lm.Held,
				Waiters:        lm.Waiters,
				WaitingReaders: lm.WaitingReaders,
				WaitingWriters: lm.WaitingWriters,
			})
		}
		__antithesis_instrumentation__.Notify(237434)
		qps, _ := rep.QueriesPerSecond()
		locality := serverpb.Locality{}
		for _, tier := range rep.GetNodeLocality().Tiers {
			__antithesis_instrumentation__.Notify(237440)
			locality.Tiers = append(locality.Tiers, serverpb.Tier{
				Key:   tier.Key,
				Value: tier.Value,
			})
		}
		__antithesis_instrumentation__.Notify(237435)
		return serverpb.RangeInfo{
			Span:          span,
			RaftState:     raftState,
			State:         state,
			SourceNodeID:  nodeID,
			SourceStoreID: storeID,
			LeaseHistory:  leaseHistory,
			Stats: serverpb.RangeStatistics{
				QueriesPerSecond: qps,
				WritesPerSecond:  rep.WritesPerSecond(),
			},
			Problems: serverpb.RangeProblems{
				Unavailable: metrics.Unavailable,
				LeaderNotLeaseHolder: metrics.Leader && func() bool {
					__antithesis_instrumentation__.Notify(237441)
					return metrics.LeaseValid == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(237442)
					return !metrics.Leaseholder == true
				}() == true,
				NoRaftLeader: !kvserver.HasRaftLeader(raftStatus) && func() bool {
					__antithesis_instrumentation__.Notify(237443)
					return !metrics.Quiescent == true
				}() == true,
				Underreplicated: metrics.Underreplicated,
				Overreplicated:  metrics.Overreplicated,
				NoLease: metrics.Leader && func() bool {
					__antithesis_instrumentation__.Notify(237444)
					return !metrics.LeaseValid == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(237445)
					return !metrics.Quiescent == true
				}() == true,
				QuiescentEqualsTicking: raftStatus != nil && func() bool {
					__antithesis_instrumentation__.Notify(237446)
					return metrics.Quiescent == metrics.Ticking == true
				}() == true,
				RaftLogTooLarge:     metrics.RaftLogTooLarge,
				CircuitBreakerError: len(state.CircuitBreakerError) > 0,
			},
			LeaseStatus:                 metrics.LeaseStatus,
			Quiescent:                   metrics.Quiescent,
			Ticking:                     metrics.Ticking,
			ReadLatches:                 metrics.LatchMetrics.ReadCount,
			WriteLatches:                metrics.LatchMetrics.WriteCount,
			Locks:                       metrics.LockTableMetrics.Locks,
			LocksWithWaitQueues:         metrics.LockTableMetrics.LocksWithWaitQueues,
			LockWaitQueueWaiters:        metrics.LockTableMetrics.Waiters,
			TopKLocksByWaitQueueWaiters: topKLocksByWaiters,
			Locality:                    &locality,
			IsLeaseholder:               metrics.Leaseholder,
			LeaseValid:                  metrics.LeaseValid,
		}
	}
	__antithesis_instrumentation__.Notify(237409)

	isLiveMap := s.nodeLiveness.GetIsLiveMap()
	clusterNodes := s.storePool.ClusterNodeCount()

	if len(req.RangeIDs) > 0 {
		__antithesis_instrumentation__.Notify(237447)
		sort.Slice(req.RangeIDs, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(237448)
			return req.RangeIDs[i] < req.RangeIDs[j]
		})
	} else {
		__antithesis_instrumentation__.Notify(237449)
	}
	__antithesis_instrumentation__.Notify(237410)

	err = s.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(237450)
		now := store.Clock().NowAsClockTimestamp()
		if len(req.RangeIDs) == 0 {
			__antithesis_instrumentation__.Notify(237453)

			store.VisitReplicas(
				func(rep *kvserver.Replica) bool {
					__antithesis_instrumentation__.Notify(237455)
					output.Ranges = append(output.Ranges,
						constructRangeInfo(
							rep,
							store.Ident.StoreID,
							rep.Metrics(ctx, now, isLiveMap, clusterNodes),
						))
					return true
				},
				kvserver.WithReplicasInOrder(),
			)
			__antithesis_instrumentation__.Notify(237454)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(237456)
		}
		__antithesis_instrumentation__.Notify(237451)

		for _, rid := range req.RangeIDs {
			__antithesis_instrumentation__.Notify(237457)
			rep, err := store.GetReplica(rid)
			if err != nil {
				__antithesis_instrumentation__.Notify(237459)

				continue
			} else {
				__antithesis_instrumentation__.Notify(237460)
			}
			__antithesis_instrumentation__.Notify(237458)
			output.Ranges = append(output.Ranges,
				constructRangeInfo(
					rep,
					store.Ident.StoreID,
					rep.Metrics(ctx, now, isLiveMap, clusterNodes),
				))
		}
		__antithesis_instrumentation__.Notify(237452)
		return nil
	})
	__antithesis_instrumentation__.Notify(237411)
	if err != nil {
		__antithesis_instrumentation__.Notify(237461)
		return nil, 0, status.Errorf(codes.Internal, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237462)
	}
	__antithesis_instrumentation__.Notify(237412)
	var next int
	if len(req.RangeIDs) > 0 {
		__antithesis_instrumentation__.Notify(237463)
		var outputInterface interface{}
		outputInterface, next = simplePaginate(output.Ranges, limit, offset)
		output.Ranges = outputInterface.([]serverpb.RangeInfo)
	} else {
		__antithesis_instrumentation__.Notify(237464)
	}
	__antithesis_instrumentation__.Notify(237413)
	return &output, next, nil
}

func (s *statusServer) TenantRanges(
	ctx context.Context, _ *serverpb.TenantRangesRequest,
) (*serverpb.TenantRangesResponse, error) {
	__antithesis_instrumentation__.Notify(237465)
	propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)
	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237472)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237473)
	}
	__antithesis_instrumentation__.Notify(237466)

	tID, ok := roachpb.TenantFromContext(ctx)
	if !ok {
		__antithesis_instrumentation__.Notify(237474)
		return nil, status.Error(codes.Internal, "no tenant ID found in context")
	} else {
		__antithesis_instrumentation__.Notify(237475)
	}
	__antithesis_instrumentation__.Notify(237467)

	tenantPrefix := keys.MakeTenantPrefix(tID)
	tenantKeySpan := roachpb.Span{
		Key:    tenantPrefix,
		EndKey: tenantPrefix.PrefixEnd(),
	}

	rangeIDs := make([]roachpb.RangeID, 0)

	replicaNodeIDs := make(map[roachpb.NodeID]struct{})
	if err := s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(237476)
		rangeKVs, err := kvclient.ScanMetaKVs(ctx, txn, tenantKeySpan)
		if err != nil {
			__antithesis_instrumentation__.Notify(237479)
			return err
		} else {
			__antithesis_instrumentation__.Notify(237480)
		}
		__antithesis_instrumentation__.Notify(237477)

		for _, rangeKV := range rangeKVs {
			__antithesis_instrumentation__.Notify(237481)
			var desc roachpb.RangeDescriptor
			if err := rangeKV.ValueProto(&desc); err != nil {
				__antithesis_instrumentation__.Notify(237483)
				return err
			} else {
				__antithesis_instrumentation__.Notify(237484)
			}
			__antithesis_instrumentation__.Notify(237482)
			rangeIDs = append(rangeIDs, desc.RangeID)
			for _, rep := range desc.Replicas().Descriptors() {
				__antithesis_instrumentation__.Notify(237485)
				_, ok := replicaNodeIDs[rep.NodeID]
				if !ok {
					__antithesis_instrumentation__.Notify(237486)
					replicaNodeIDs[rep.NodeID] = struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(237487)
				}
			}
		}
		__antithesis_instrumentation__.Notify(237478)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(237488)
		return nil, status.Error(
			codes.Internal,
			errors.Wrap(err, "there was a problem with the initial fetch of range IDs").Error())
	} else {
		__antithesis_instrumentation__.Notify(237489)
	}
	__antithesis_instrumentation__.Notify(237468)

	nodeResults := make([][]serverpb.RangeInfo, 0, len(replicaNodeIDs))
	for nodeID := range replicaNodeIDs {
		__antithesis_instrumentation__.Notify(237490)
		nodeIDString := nodeID.String()
		_, local, err := s.parseNodeID(nodeIDString)
		if err != nil {
			__antithesis_instrumentation__.Notify(237493)
			return nil, status.Errorf(codes.Internal, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(237494)
		}
		__antithesis_instrumentation__.Notify(237491)

		req := &serverpb.RangesRequest{
			NodeId:   nodeIDString,
			RangeIDs: rangeIDs,
		}

		var resp *serverpb.RangesResponse
		if local {
			__antithesis_instrumentation__.Notify(237495)
			resp, _, err = s.rangesHelper(ctx, req, 0, 0)
			if err != nil {
				__antithesis_instrumentation__.Notify(237496)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(237497)
			}
		} else {
			__antithesis_instrumentation__.Notify(237498)
			statusServer, err := s.dialNode(ctx, nodeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(237500)
				return nil, serverError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(237501)
			}
			__antithesis_instrumentation__.Notify(237499)

			resp, err = statusServer.Ranges(ctx, req)
			if err != nil {
				__antithesis_instrumentation__.Notify(237502)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(237503)
			}
		}
		__antithesis_instrumentation__.Notify(237492)

		nodeResults = append(nodeResults, resp.Ranges)
	}
	__antithesis_instrumentation__.Notify(237469)

	transformTenantRange := func(
		rep serverpb.RangeInfo,
	) (string, *serverpb.TenantRangeInfo) {
		__antithesis_instrumentation__.Notify(237504)
		topKLocksByWaiters := make([]serverpb.TenantRangeInfo_LockInfo, 0, len(rep.TopKLocksByWaitQueueWaiters))
		for _, lm := range rep.TopKLocksByWaitQueueWaiters {
			__antithesis_instrumentation__.Notify(237507)
			topKLocksByWaiters = append(topKLocksByWaiters, serverpb.TenantRangeInfo_LockInfo{
				PrettyKey:      lm.Key.String(),
				Key:            lm.Key,
				Held:           lm.Held,
				Waiters:        lm.Waiters,
				WaitingReaders: lm.WaitingReaders,
				WaitingWriters: lm.WaitingWriters,
			})
		}
		__antithesis_instrumentation__.Notify(237505)
		azKey := "az"
		localityKey := "locality-unset"
		for _, tier := range rep.Locality.Tiers {
			__antithesis_instrumentation__.Notify(237508)
			if tier.Key == azKey {
				__antithesis_instrumentation__.Notify(237509)
				localityKey = tier.Value
			} else {
				__antithesis_instrumentation__.Notify(237510)
			}
		}
		__antithesis_instrumentation__.Notify(237506)
		return localityKey, &serverpb.TenantRangeInfo{
			RangeID:                     rep.State.Desc.RangeID,
			Span:                        rep.Span,
			Locality:                    rep.Locality,
			IsLeaseholder:               rep.IsLeaseholder,
			LeaseValid:                  rep.LeaseValid,
			RangeStats:                  rep.Stats,
			MvccStats:                   rep.State.Stats,
			ReadLatches:                 rep.ReadLatches,
			WriteLatches:                rep.WriteLatches,
			Locks:                       rep.Locks,
			LocksWithWaitQueues:         rep.LocksWithWaitQueues,
			LockWaitQueueWaiters:        rep.LockWaitQueueWaiters,
			TopKLocksByWaitQueueWaiters: topKLocksByWaiters,
		}
	}
	__antithesis_instrumentation__.Notify(237470)

	resp := &serverpb.TenantRangesResponse{
		RangesByLocality: make(map[string]serverpb.TenantRangesResponse_TenantRangeList),
	}

	for _, rangeMetas := range nodeResults {
		__antithesis_instrumentation__.Notify(237511)
		for _, rangeMeta := range rangeMetas {
			__antithesis_instrumentation__.Notify(237512)
			localityKey, rangeInfo := transformTenantRange(rangeMeta)
			rangeList, ok := resp.RangesByLocality[localityKey]
			if !ok {
				__antithesis_instrumentation__.Notify(237514)
				rangeList = serverpb.TenantRangesResponse_TenantRangeList{
					Ranges: make([]serverpb.TenantRangeInfo, 0),
				}
			} else {
				__antithesis_instrumentation__.Notify(237515)
			}
			__antithesis_instrumentation__.Notify(237513)
			rangeList.Ranges = append(rangeList.Ranges, *rangeInfo)
			resp.RangesByLocality[localityKey] = rangeList
		}
	}
	__antithesis_instrumentation__.Notify(237471)

	return resp, nil
}

func (s *statusServer) HotRanges(
	ctx context.Context, req *serverpb.HotRangesRequest,
) (*serverpb.HotRangesResponse, error) {
	__antithesis_instrumentation__.Notify(237516)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237524)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237525)
	}
	__antithesis_instrumentation__.Notify(237517)

	response := &serverpb.HotRangesResponse{
		NodeID:            s.gossip.NodeID.Get(),
		HotRangesByNodeID: make(map[roachpb.NodeID]serverpb.HotRangesResponse_NodeResponse),
	}

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(237526)
		requestedNodeID, local, err := s.parseNodeID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237530)
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(237531)
		}
		__antithesis_instrumentation__.Notify(237527)

		if local {
			__antithesis_instrumentation__.Notify(237532)
			response.HotRangesByNodeID[requestedNodeID] = s.localHotRanges(ctx)
			return response, nil
		} else {
			__antithesis_instrumentation__.Notify(237533)
		}
		__antithesis_instrumentation__.Notify(237528)

		status, err := s.dialNode(ctx, requestedNodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237534)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237535)
		}
		__antithesis_instrumentation__.Notify(237529)
		return status.HotRanges(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237536)
	}
	__antithesis_instrumentation__.Notify(237518)

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237537)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(237519)
	remoteRequest := serverpb.HotRangesRequest{NodeID: "local"}
	nodeFn := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237538)
		status := client.(serverpb.StatusClient)
		return status.HotRanges(ctx, &remoteRequest)
	}
	__antithesis_instrumentation__.Notify(237520)
	responseFn := func(nodeID roachpb.NodeID, resp interface{}) {
		__antithesis_instrumentation__.Notify(237539)
		hotRangesResp := resp.(*serverpb.HotRangesResponse)
		response.HotRangesByNodeID[nodeID] = hotRangesResp.HotRangesByNodeID[nodeID]
	}
	__antithesis_instrumentation__.Notify(237521)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(237540)
		response.HotRangesByNodeID[nodeID] = serverpb.HotRangesResponse_NodeResponse{
			ErrorMessage: err.Error(),
		}
	}
	__antithesis_instrumentation__.Notify(237522)

	if err := s.iterateNodes(ctx, "hot ranges", dialFn, nodeFn, responseFn, errorFn); err != nil {
		__antithesis_instrumentation__.Notify(237541)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237542)
	}
	__antithesis_instrumentation__.Notify(237523)

	return response, nil
}

type hotRangeReportMeta struct {
	dbName         string
	tableName      string
	schemaName     string
	indexNames     map[uint32]string
	parentID       uint32
	schemaParentID uint32
}

func (s *statusServer) HotRangesV2(
	ctx context.Context, req *serverpb.HotRangesRequest,
) (*serverpb.HotRangesResponseV2, error) {
	__antithesis_instrumentation__.Notify(237543)
	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237555)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237556)
	}
	__antithesis_instrumentation__.Notify(237544)

	size := int(req.PageSize)
	start := paginationState{}

	if len(req.PageToken) > 0 {
		__antithesis_instrumentation__.Notify(237557)
		if err := start.UnmarshalText([]byte(req.PageToken)); err != nil {
			__antithesis_instrumentation__.Notify(237558)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(237559)
		}
	} else {
		__antithesis_instrumentation__.Notify(237560)
	}
	__antithesis_instrumentation__.Notify(237545)

	rangeReportMetas := make(map[uint32]hotRangeReportMeta)
	var descrs []catalog.Descriptor
	var err error
	if err := s.sqlServer.distSQLServer.CollectionFactory.Txn(
		ctx, s.sqlServer.internalExecutor, s.db,
		func(ctx context.Context, txn *kv.Txn, descriptors *descs.Collection) error {
			__antithesis_instrumentation__.Notify(237561)
			all, err := descriptors.GetAllDescriptors(ctx, txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(237563)
				return err
			} else {
				__antithesis_instrumentation__.Notify(237564)
			}
			__antithesis_instrumentation__.Notify(237562)
			descrs = all.OrderedDescriptors()
			return nil
		}); err != nil {
		__antithesis_instrumentation__.Notify(237565)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237566)
	}
	__antithesis_instrumentation__.Notify(237546)

	for _, desc := range descrs {
		__antithesis_instrumentation__.Notify(237567)
		id := uint32(desc.GetID())
		meta := hotRangeReportMeta{
			indexNames: map[uint32]string{},
		}
		switch desc := desc.(type) {
		case catalog.TableDescriptor:
			__antithesis_instrumentation__.Notify(237569)
			meta.tableName = desc.GetName()
			meta.parentID = uint32(desc.GetParentID())
			meta.schemaParentID = uint32(desc.GetParentSchemaID())
			for _, idx := range desc.AllIndexes() {
				__antithesis_instrumentation__.Notify(237572)
				meta.indexNames[uint32(idx.GetID())] = idx.GetName()
			}
		case catalog.SchemaDescriptor:
			__antithesis_instrumentation__.Notify(237570)
			meta.schemaName = desc.GetName()
		case catalog.DatabaseDescriptor:
			__antithesis_instrumentation__.Notify(237571)
			meta.dbName = desc.GetName()
		}
		__antithesis_instrumentation__.Notify(237568)
		rangeReportMetas[id] = meta
	}
	__antithesis_instrumentation__.Notify(237547)

	response := &serverpb.HotRangesResponseV2{
		ErrorsByNodeID: make(map[roachpb.NodeID]string),
	}

	var requestedNodes []roachpb.NodeID
	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(237573)
		requestedNodeID, _, err := s.parseNodeID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237575)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(237576)
		}
		__antithesis_instrumentation__.Notify(237574)
		requestedNodes = []roachpb.NodeID{requestedNodeID}
	} else {
		__antithesis_instrumentation__.Notify(237577)
	}
	__antithesis_instrumentation__.Notify(237548)

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237578)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(237549)
	remoteRequest := serverpb.HotRangesRequest{NodeID: "local"}
	nodeFn := func(ctx context.Context, client interface{}, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237579)
		status := client.(serverpb.StatusClient)
		resp, err := status.HotRanges(ctx, &remoteRequest)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(237582)
			return resp == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(237583)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(237584)
		}
		__antithesis_instrumentation__.Notify(237580)
		var ranges []*serverpb.HotRangesResponseV2_HotRange
		for nodeID, hr := range resp.HotRangesByNodeID {
			__antithesis_instrumentation__.Notify(237585)
			for _, store := range hr.Stores {
				__antithesis_instrumentation__.Notify(237586)
				for _, r := range store.HotRanges {
					__antithesis_instrumentation__.Notify(237587)
					var (
						dbName, tableName, indexName, schemaName string
						replicaNodeIDs                           []roachpb.NodeID
					)
					_, tableID, err := s.sqlServer.execCfg.Codec.DecodeTablePrefix(r.Desc.StartKey.AsRawKey())
					if err != nil {
						__antithesis_instrumentation__.Notify(237592)
						log.Warningf(ctx, "cannot decode tableID for range descriptor: %s. %s", r.Desc.String(), err.Error())
						continue
					} else {
						__antithesis_instrumentation__.Notify(237593)
					}
					__antithesis_instrumentation__.Notify(237588)
					parent := rangeReportMetas[tableID].parentID
					if parent != 0 {
						__antithesis_instrumentation__.Notify(237594)
						tableName = rangeReportMetas[tableID].tableName
						dbName = rangeReportMetas[parent].dbName
					} else {
						__antithesis_instrumentation__.Notify(237595)
						dbName = rangeReportMetas[tableID].dbName
					}
					__antithesis_instrumentation__.Notify(237589)
					schemaParent := rangeReportMetas[tableID].schemaParentID
					schemaName = rangeReportMetas[schemaParent].schemaName
					_, _, idxID, err := s.sqlServer.execCfg.Codec.DecodeIndexPrefix(r.Desc.StartKey.AsRawKey())
					if err == nil {
						__antithesis_instrumentation__.Notify(237596)
						indexName = rangeReportMetas[tableID].indexNames[idxID]
					} else {
						__antithesis_instrumentation__.Notify(237597)
					}
					__antithesis_instrumentation__.Notify(237590)
					for _, repl := range r.Desc.Replicas().Descriptors() {
						__antithesis_instrumentation__.Notify(237598)
						replicaNodeIDs = append(replicaNodeIDs, repl.NodeID)
					}
					__antithesis_instrumentation__.Notify(237591)
					ranges = append(ranges, &serverpb.HotRangesResponseV2_HotRange{
						RangeID:           r.Desc.RangeID,
						NodeID:            nodeID,
						QPS:               r.QueriesPerSecond,
						TableName:         tableName,
						SchemaName:        schemaName,
						DatabaseName:      dbName,
						IndexName:         indexName,
						ReplicaNodeIds:    replicaNodeIDs,
						LeaseholderNodeID: r.LeaseholderNodeID,
						StoreID:           store.StoreID,
					})
				}
			}
		}
		__antithesis_instrumentation__.Notify(237581)
		return ranges, nil
	}
	__antithesis_instrumentation__.Notify(237550)
	responseFn := func(nodeID roachpb.NodeID, resp interface{}) {
		__antithesis_instrumentation__.Notify(237599)
		if resp == nil {
			__antithesis_instrumentation__.Notify(237601)
			return
		} else {
			__antithesis_instrumentation__.Notify(237602)
		}
		__antithesis_instrumentation__.Notify(237600)
		hotRanges := resp.([]*serverpb.HotRangesResponseV2_HotRange)
		response.Ranges = append(response.Ranges, hotRanges...)
	}
	__antithesis_instrumentation__.Notify(237551)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(237603)
		response.ErrorsByNodeID[nodeID] = err.Error()
	}
	__antithesis_instrumentation__.Notify(237552)

	next, err := s.paginatedIterateNodes(
		ctx, "hotRanges", size, start, requestedNodes, dialFn,
		nodeFn, responseFn, errorFn)

	if err != nil {
		__antithesis_instrumentation__.Notify(237604)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237605)
	}
	__antithesis_instrumentation__.Notify(237553)
	var nextBytes []byte
	if nextBytes, err = next.MarshalText(); err != nil {
		__antithesis_instrumentation__.Notify(237606)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237607)
	}
	__antithesis_instrumentation__.Notify(237554)
	response.NextPageToken = string(nextBytes)
	return response, nil
}

func (s *statusServer) localHotRanges(ctx context.Context) serverpb.HotRangesResponse_NodeResponse {
	__antithesis_instrumentation__.Notify(237608)
	var resp serverpb.HotRangesResponse_NodeResponse
	err := s.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(237611)
		ranges := store.HottestReplicas()
		storeResp := &serverpb.HotRangesResponse_StoreResponse{
			StoreID:   store.StoreID(),
			HotRanges: make([]serverpb.HotRangesResponse_HotRange, len(ranges)),
		}
		for i, r := range ranges {
			__antithesis_instrumentation__.Notify(237613)
			replica, err := store.GetReplica(r.Desc.GetRangeID())
			if err == nil {
				__antithesis_instrumentation__.Notify(237615)
				storeResp.HotRanges[i].LeaseholderNodeID = replica.State(ctx).Lease.Replica.NodeID
			} else {
				__antithesis_instrumentation__.Notify(237616)
			}
			__antithesis_instrumentation__.Notify(237614)
			storeResp.HotRanges[i].Desc = *r.Desc
			storeResp.HotRanges[i].QueriesPerSecond = r.QPS
		}
		__antithesis_instrumentation__.Notify(237612)
		resp.Stores = append(resp.Stores, storeResp)
		return nil
	})
	__antithesis_instrumentation__.Notify(237609)
	if err != nil {
		__antithesis_instrumentation__.Notify(237617)
		return serverpb.HotRangesResponse_NodeResponse{ErrorMessage: err.Error()}
	} else {
		__antithesis_instrumentation__.Notify(237618)
	}
	__antithesis_instrumentation__.Notify(237610)
	return resp
}

func (s *statusServer) Range(
	ctx context.Context, req *serverpb.RangeRequest,
) (*serverpb.RangeResponse, error) {
	__antithesis_instrumentation__.Notify(237619)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237626)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237627)
	}
	__antithesis_instrumentation__.Notify(237620)

	response := &serverpb.RangeResponse{
		RangeID:           roachpb.RangeID(req.RangeId),
		NodeID:            s.gossip.NodeID.Get(),
		ResponsesByNodeID: make(map[roachpb.NodeID]serverpb.RangeResponse_NodeResponse),
	}

	rangesRequest := &serverpb.RangesRequest{
		RangeIDs: []roachpb.RangeID{roachpb.RangeID(req.RangeId)},
	}

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237628)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(237621)
	nodeFn := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237629)
		status := client.(serverpb.StatusClient)
		return status.Ranges(ctx, rangesRequest)
	}
	__antithesis_instrumentation__.Notify(237622)
	nowNanos := timeutil.Now().UnixNano()
	responseFn := func(nodeID roachpb.NodeID, resp interface{}) {
		__antithesis_instrumentation__.Notify(237630)
		rangesResp := resp.(*serverpb.RangesResponse)

		for i := range rangesResp.Ranges {
			__antithesis_instrumentation__.Notify(237632)
			rangesResp.Ranges[i].State.Stats.AgeTo(nowNanos)
		}
		__antithesis_instrumentation__.Notify(237631)
		response.ResponsesByNodeID[nodeID] = serverpb.RangeResponse_NodeResponse{
			Response: true,
			Infos:    rangesResp.Ranges,
		}
	}
	__antithesis_instrumentation__.Notify(237623)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(237633)
		response.ResponsesByNodeID[nodeID] = serverpb.RangeResponse_NodeResponse{
			ErrorMessage: err.Error(),
		}
	}
	__antithesis_instrumentation__.Notify(237624)

	if err := s.iterateNodes(
		ctx, fmt.Sprintf("details about range %d", req.RangeId), dialFn, nodeFn, responseFn, errorFn,
	); err != nil {
		__antithesis_instrumentation__.Notify(237634)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237635)
	}
	__antithesis_instrumentation__.Notify(237625)
	return response, nil
}

func (s *statusServer) ListLocalSessions(
	ctx context.Context, req *serverpb.ListSessionsRequest,
) (*serverpb.ListSessionsResponse, error) {
	__antithesis_instrumentation__.Notify(237636)
	sessions, err := s.getLocalSessions(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(237639)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237640)
	}
	__antithesis_instrumentation__.Notify(237637)
	for i := range sessions {
		__antithesis_instrumentation__.Notify(237641)
		sessions[i].NodeID = s.gossip.NodeID.Get()
	}
	__antithesis_instrumentation__.Notify(237638)
	return &serverpb.ListSessionsResponse{Sessions: sessions}, nil
}

func (s *statusServer) iterateNodes(
	ctx context.Context,
	errorCtx string,
	dialFn func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error),
	nodeFn func(ctx context.Context, client interface{}, nodeID roachpb.NodeID) (interface{}, error),
	responseFn func(nodeID roachpb.NodeID, resp interface{}),
	errorFn func(nodeID roachpb.NodeID, nodeFnError error),
) error {
	__antithesis_instrumentation__.Notify(237642)
	nodeStatuses, err := s.nodesStatusWithLiveness(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(237647)
		return err
	} else {
		__antithesis_instrumentation__.Notify(237648)
	}
	__antithesis_instrumentation__.Notify(237643)

	type nodeResponse struct {
		nodeID   roachpb.NodeID
		response interface{}
		err      error
	}

	numNodes := len(nodeStatuses)
	responseChan := make(chan nodeResponse, numNodes)

	nodeQuery := func(ctx context.Context, nodeID roachpb.NodeID) {
		__antithesis_instrumentation__.Notify(237649)
		var client interface{}
		err := contextutil.RunWithTimeout(ctx, "dial node", base.NetworkTimeout, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(237653)
			var err error
			client, err = dialFn(ctx, nodeID)
			return err
		})
		__antithesis_instrumentation__.Notify(237650)
		if err != nil {
			__antithesis_instrumentation__.Notify(237654)
			err = errors.Wrapf(err, "failed to dial into node %d (%s)",
				nodeID, nodeStatuses[nodeID].livenessStatus)
			responseChan <- nodeResponse{nodeID: nodeID, err: err}
			return
		} else {
			__antithesis_instrumentation__.Notify(237655)
		}
		__antithesis_instrumentation__.Notify(237651)

		res, err := nodeFn(ctx, client, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237656)
			err = errors.Wrapf(err, "error requesting %s from node %d (%s)",
				errorCtx, nodeID, nodeStatuses[nodeID].livenessStatus)
		} else {
			__antithesis_instrumentation__.Notify(237657)
		}
		__antithesis_instrumentation__.Notify(237652)
		responseChan <- nodeResponse{nodeID: nodeID, response: res, err: err}
	}
	__antithesis_instrumentation__.Notify(237644)

	sem := quotapool.NewIntPool("node status", maxConcurrentRequests)
	ctx, cancel := s.stopper.WithCancelOnQuiesce(ctx)
	defer cancel()
	for nodeID := range nodeStatuses {
		__antithesis_instrumentation__.Notify(237658)
		nodeID := nodeID
		if err := s.stopper.RunAsyncTaskEx(
			ctx,
			stop.TaskOpts{
				TaskName:   fmt.Sprintf("server.statusServer: requesting %s", errorCtx),
				Sem:        sem,
				WaitForSem: true,
			},
			func(ctx context.Context) { __antithesis_instrumentation__.Notify(237659); nodeQuery(ctx, nodeID) },
		); err != nil {
			__antithesis_instrumentation__.Notify(237660)
			return err
		} else {
			__antithesis_instrumentation__.Notify(237661)
		}
	}
	__antithesis_instrumentation__.Notify(237645)

	var resultErr error
	for numNodes > 0 {
		__antithesis_instrumentation__.Notify(237662)
		select {
		case res := <-responseChan:
			__antithesis_instrumentation__.Notify(237664)
			if res.err != nil {
				__antithesis_instrumentation__.Notify(237666)
				errorFn(res.nodeID, res.err)
			} else {
				__antithesis_instrumentation__.Notify(237667)
				responseFn(res.nodeID, res.response)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(237665)
			resultErr = errors.Errorf("request of %s canceled before completion", errorCtx)
		}
		__antithesis_instrumentation__.Notify(237663)
		numNodes--
	}
	__antithesis_instrumentation__.Notify(237646)
	return resultErr
}

func (s *statusServer) paginatedIterateNodes(
	ctx context.Context,
	errorCtx string,
	limit int,
	pagState paginationState,
	requestedNodes []roachpb.NodeID,
	dialFn func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error),
	nodeFn func(ctx context.Context, client interface{}, nodeID roachpb.NodeID) (interface{}, error),
	responseFn func(nodeID roachpb.NodeID, resp interface{}),
	errorFn func(nodeID roachpb.NodeID, nodeFnError error),
) (next paginationState, err error) {
	__antithesis_instrumentation__.Notify(237668)
	if limit == 0 {
		__antithesis_instrumentation__.Notify(237675)
		return paginationState{}, s.iterateNodes(ctx, errorCtx, dialFn, nodeFn, responseFn, errorFn)
	} else {
		__antithesis_instrumentation__.Notify(237676)
	}
	__antithesis_instrumentation__.Notify(237669)
	nodeStatuses, err := s.nodesStatusWithLiveness(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(237677)
		return paginationState{}, err
	} else {
		__antithesis_instrumentation__.Notify(237678)
	}
	__antithesis_instrumentation__.Notify(237670)

	numNodes := len(nodeStatuses)
	nodeIDs := make([]roachpb.NodeID, 0, numNodes)
	if len(requestedNodes) > 0 {
		__antithesis_instrumentation__.Notify(237679)
		nodeIDs = append(nodeIDs, requestedNodes...)
	} else {
		__antithesis_instrumentation__.Notify(237680)
		for nodeID := range nodeStatuses {
			__antithesis_instrumentation__.Notify(237681)
			nodeIDs = append(nodeIDs, nodeID)
		}
	}
	__antithesis_instrumentation__.Notify(237671)

	sort.Slice(nodeIDs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(237682)
		return nodeIDs[i] < nodeIDs[j]
	})
	__antithesis_instrumentation__.Notify(237672)
	pagState.mergeNodeIDs(nodeIDs)

	nodeIDs = nodeIDs[:0]
	if pagState.inProgress != 0 {
		__antithesis_instrumentation__.Notify(237683)
		nodeIDs = append(nodeIDs, pagState.inProgress)
	} else {
		__antithesis_instrumentation__.Notify(237684)
	}
	__antithesis_instrumentation__.Notify(237673)
	nodeIDs = append(nodeIDs, pagState.nodesToQuery...)

	paginator := &rpcNodePaginator{
		limit:        limit,
		numNodes:     len(nodeIDs),
		errorCtx:     errorCtx,
		pagState:     pagState,
		nodeStatuses: nodeStatuses,
		dialFn:       dialFn,
		nodeFn:       nodeFn,
		responseFn:   responseFn,
		errorFn:      errorFn,
	}

	paginator.init()

	sem := quotapool.NewIntPool("node status", maxConcurrentPaginatedRequests)
	ctx, cancel := s.stopper.WithCancelOnQuiesce(ctx)
	defer cancel()
	for idx, nodeID := range nodeIDs {
		__antithesis_instrumentation__.Notify(237685)
		nodeID := nodeID
		idx := idx
		if err := s.stopper.RunAsyncTaskEx(
			ctx,
			stop.TaskOpts{
				TaskName:   fmt.Sprintf("server.statusServer: requesting %s", errorCtx),
				Sem:        sem,
				WaitForSem: true,
			},
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(237686)
				paginator.queryNode(ctx, nodeID, idx)
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(237687)
			return pagState, err
		} else {
			__antithesis_instrumentation__.Notify(237688)
		}
	}
	__antithesis_instrumentation__.Notify(237674)

	return paginator.processResponses(ctx)
}

func (s *statusServer) listSessionsHelper(
	ctx context.Context, req *serverpb.ListSessionsRequest, limit int, start paginationState,
) (*serverpb.ListSessionsResponse, paginationState, error) {
	__antithesis_instrumentation__.Notify(237689)
	response := &serverpb.ListSessionsResponse{
		Sessions:              make([]serverpb.Session, 0),
		Errors:                make([]serverpb.ListSessionsError, 0),
		InternalAppNamePrefix: catconstants.InternalAppNamePrefix,
	}

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237695)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(237690)
	nodeFn := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237696)
		statusClient := client.(serverpb.StatusClient)
		resp, err := statusClient.ListLocalSessions(ctx, req)
		if resp != nil && func() bool {
			__antithesis_instrumentation__.Notify(237698)
			return err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(237699)
			if len(resp.Errors) > 0 {
				__antithesis_instrumentation__.Notify(237702)
				return nil, errors.Errorf("%s", resp.Errors[0].Message)
			} else {
				__antithesis_instrumentation__.Notify(237703)
			}
			__antithesis_instrumentation__.Notify(237700)
			sort.Slice(resp.Sessions, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(237704)
				return resp.Sessions[i].Start.Before(resp.Sessions[j].Start)
			})
			__antithesis_instrumentation__.Notify(237701)
			return resp.Sessions, nil
		} else {
			__antithesis_instrumentation__.Notify(237705)
		}
		__antithesis_instrumentation__.Notify(237697)
		return nil, err
	}
	__antithesis_instrumentation__.Notify(237691)
	responseFn := func(_ roachpb.NodeID, nodeResp interface{}) {
		__antithesis_instrumentation__.Notify(237706)
		if nodeResp == nil {
			__antithesis_instrumentation__.Notify(237708)
			return
		} else {
			__antithesis_instrumentation__.Notify(237709)
		}
		__antithesis_instrumentation__.Notify(237707)
		sessions := nodeResp.([]serverpb.Session)
		response.Sessions = append(response.Sessions, sessions...)
	}
	__antithesis_instrumentation__.Notify(237692)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(237710)
		errResponse := serverpb.ListSessionsError{NodeID: nodeID, Message: err.Error()}
		response.Errors = append(response.Errors, errResponse)
	}
	__antithesis_instrumentation__.Notify(237693)

	var err error
	var pagState paginationState
	if pagState, err = s.paginatedIterateNodes(
		ctx, "session list", limit, start, nil, dialFn, nodeFn, responseFn, errorFn); err != nil {
		__antithesis_instrumentation__.Notify(237711)
		err := serverpb.ListSessionsError{Message: err.Error()}
		response.Errors = append(response.Errors, err)
	} else {
		__antithesis_instrumentation__.Notify(237712)
	}
	__antithesis_instrumentation__.Notify(237694)
	return response, pagState, nil
}

func (s *statusServer) ListSessions(
	ctx context.Context, req *serverpb.ListSessionsRequest,
) (*serverpb.ListSessionsResponse, error) {
	__antithesis_instrumentation__.Notify(237713)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, _, err := s.privilegeChecker.getUserAndRole(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237716)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237717)
	}
	__antithesis_instrumentation__.Notify(237714)

	resp, _, err := s.listSessionsHelper(ctx, req, 0, paginationState{})
	if err != nil {
		__antithesis_instrumentation__.Notify(237718)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237719)
	}
	__antithesis_instrumentation__.Notify(237715)
	return resp, nil
}

func (s *statusServer) CancelSession(
	ctx context.Context, req *serverpb.CancelSessionRequest,
) (*serverpb.CancelSessionResponse, error) {
	__antithesis_instrumentation__.Notify(237720)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237726)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237727)
	}
	__antithesis_instrumentation__.Notify(237721)

	if !local {
		__antithesis_instrumentation__.Notify(237728)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237730)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237731)
		}
		__antithesis_instrumentation__.Notify(237729)
		return status.CancelSession(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237732)
	}
	__antithesis_instrumentation__.Notify(237722)

	reqUsername, err := security.MakeSQLUsernameFromPreNormalizedStringChecked(req.Username)
	if err != nil {
		__antithesis_instrumentation__.Notify(237733)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237734)
	}
	__antithesis_instrumentation__.Notify(237723)

	if err := s.checkCancelPrivilege(ctx, reqUsername, findSessionBySessionID(req.SessionID)); err != nil {
		__antithesis_instrumentation__.Notify(237735)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237736)
	}
	__antithesis_instrumentation__.Notify(237724)

	r, err := s.sessionRegistry.CancelSession(req.SessionID)
	if err != nil {
		__antithesis_instrumentation__.Notify(237737)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237738)
	}
	__antithesis_instrumentation__.Notify(237725)
	return r, nil
}

func (s *statusServer) CancelQuery(
	ctx context.Context, req *serverpb.CancelQueryRequest,
) (*serverpb.CancelQueryResponse, error) {
	__antithesis_instrumentation__.Notify(237739)
	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237745)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237746)
	}
	__antithesis_instrumentation__.Notify(237740)
	if !local {
		__antithesis_instrumentation__.Notify(237747)

		ctx = propagateGatewayMetadata(ctx)
		ctx = s.AnnotateCtx(ctx)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237749)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237750)
		}
		__antithesis_instrumentation__.Notify(237748)
		return status.CancelQuery(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237751)
	}
	__antithesis_instrumentation__.Notify(237741)

	reqUsername, err := security.MakeSQLUsernameFromPreNormalizedStringChecked(req.Username)
	if err != nil {
		__antithesis_instrumentation__.Notify(237752)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237753)
	}
	__antithesis_instrumentation__.Notify(237742)

	if err := s.checkCancelPrivilege(ctx, reqUsername, findSessionByQueryID(req.QueryID)); err != nil {
		__antithesis_instrumentation__.Notify(237754)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237755)
	}
	__antithesis_instrumentation__.Notify(237743)

	output := &serverpb.CancelQueryResponse{}
	output.Canceled, err = s.sessionRegistry.CancelQuery(req.QueryID)
	if err != nil {
		__antithesis_instrumentation__.Notify(237756)
		output.Error = err.Error()
	} else {
		__antithesis_instrumentation__.Notify(237757)
	}
	__antithesis_instrumentation__.Notify(237744)
	return output, nil
}

func (s *statusServer) CancelQueryByKey(
	ctx context.Context, req *serverpb.CancelQueryByKeyRequest,
) (resp *serverpb.CancelQueryByKeyResponse, retErr error) {
	__antithesis_instrumentation__.Notify(237758)
	local := req.SQLInstanceID == s.sqlServer.SQLInstanceID()

	alloc, err := pgwirecancel.CancelSemaphore.TryAcquire(ctx, 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(237763)
		return nil, status.Errorf(codes.ResourceExhausted, "exceeded rate limit of pgwire cancellation requests")
	} else {
		__antithesis_instrumentation__.Notify(237764)
	}
	__antithesis_instrumentation__.Notify(237759)
	defer func() {
		__antithesis_instrumentation__.Notify(237765)

		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(237767)
			return (resp != nil && func() bool {
				__antithesis_instrumentation__.Notify(237768)
				return !resp.Canceled == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(237769)
			time.Sleep(1 * time.Second)
		} else {
			__antithesis_instrumentation__.Notify(237770)
		}
		__antithesis_instrumentation__.Notify(237766)
		alloc.Release()
	}()
	__antithesis_instrumentation__.Notify(237760)

	if local {
		__antithesis_instrumentation__.Notify(237771)
		resp = &serverpb.CancelQueryByKeyResponse{}
		resp.Canceled, err = s.sessionRegistry.CancelQueryByKey(req.CancelQueryKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(237773)
			resp.Error = err.Error()
		} else {
			__antithesis_instrumentation__.Notify(237774)
		}
		__antithesis_instrumentation__.Notify(237772)
		return resp, nil
	} else {
		__antithesis_instrumentation__.Notify(237775)
	}
	__antithesis_instrumentation__.Notify(237761)

	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)
	client, err := s.dialNode(ctx, roachpb.NodeID(req.SQLInstanceID))
	if err != nil {
		__antithesis_instrumentation__.Notify(237776)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237777)
	}
	__antithesis_instrumentation__.Notify(237762)
	return client.CancelQueryByKey(ctx, req)
}

func (s *statusServer) ListContentionEvents(
	ctx context.Context, req *serverpb.ListContentionEventsRequest,
) (*serverpb.ListContentionEventsResponse, error) {
	__antithesis_instrumentation__.Notify(237778)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237785)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237786)
	}
	__antithesis_instrumentation__.Notify(237779)

	var response serverpb.ListContentionEventsResponse
	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237787)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(237780)
	nodeFn := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237788)
		statusClient := client.(serverpb.StatusClient)
		resp, err := statusClient.ListLocalContentionEvents(ctx, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(237791)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(237792)
		}
		__antithesis_instrumentation__.Notify(237789)
		if len(resp.Errors) > 0 {
			__antithesis_instrumentation__.Notify(237793)
			return nil, errors.Errorf("%s", resp.Errors[0].Message)
		} else {
			__antithesis_instrumentation__.Notify(237794)
		}
		__antithesis_instrumentation__.Notify(237790)
		return resp, nil
	}
	__antithesis_instrumentation__.Notify(237781)
	responseFn := func(_ roachpb.NodeID, nodeResp interface{}) {
		__antithesis_instrumentation__.Notify(237795)
		if nodeResp == nil {
			__antithesis_instrumentation__.Notify(237797)
			return
		} else {
			__antithesis_instrumentation__.Notify(237798)
		}
		__antithesis_instrumentation__.Notify(237796)
		events := nodeResp.(*serverpb.ListContentionEventsResponse).Events
		response.Events = contention.MergeSerializedRegistries(response.Events, events)
	}
	__antithesis_instrumentation__.Notify(237782)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(237799)
		errResponse := serverpb.ListActivityError{NodeID: nodeID, Message: err.Error()}
		response.Errors = append(response.Errors, errResponse)
	}
	__antithesis_instrumentation__.Notify(237783)

	if err := s.iterateNodes(ctx, "contention events list", dialFn, nodeFn, responseFn, errorFn); err != nil {
		__antithesis_instrumentation__.Notify(237800)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237801)
	}
	__antithesis_instrumentation__.Notify(237784)
	return &response, nil
}

func (s *statusServer) ListDistSQLFlows(
	ctx context.Context, request *serverpb.ListDistSQLFlowsRequest,
) (*serverpb.ListDistSQLFlowsResponse, error) {
	__antithesis_instrumentation__.Notify(237802)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237809)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237810)
	}
	__antithesis_instrumentation__.Notify(237803)

	var response serverpb.ListDistSQLFlowsResponse
	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237811)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(237804)
	nodeFn := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237812)
		statusClient := client.(serverpb.StatusClient)
		resp, err := statusClient.ListLocalDistSQLFlows(ctx, request)
		if err != nil {
			__antithesis_instrumentation__.Notify(237815)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(237816)
		}
		__antithesis_instrumentation__.Notify(237813)
		if len(resp.Errors) > 0 {
			__antithesis_instrumentation__.Notify(237817)
			return nil, errors.Errorf("%s", resp.Errors[0].Message)
		} else {
			__antithesis_instrumentation__.Notify(237818)
		}
		__antithesis_instrumentation__.Notify(237814)
		return resp, nil
	}
	__antithesis_instrumentation__.Notify(237805)
	responseFn := func(_ roachpb.NodeID, nodeResp interface{}) {
		__antithesis_instrumentation__.Notify(237819)
		if nodeResp == nil {
			__antithesis_instrumentation__.Notify(237821)
			return
		} else {
			__antithesis_instrumentation__.Notify(237822)
		}
		__antithesis_instrumentation__.Notify(237820)
		flows := nodeResp.(*serverpb.ListDistSQLFlowsResponse).Flows
		response.Flows = mergeDistSQLRemoteFlows(response.Flows, flows)
	}
	__antithesis_instrumentation__.Notify(237806)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(237823)
		errResponse := serverpb.ListActivityError{NodeID: nodeID, Message: err.Error()}
		response.Errors = append(response.Errors, errResponse)
	}
	__antithesis_instrumentation__.Notify(237807)

	if err := s.iterateNodes(ctx, "distsql flows list", dialFn, nodeFn, responseFn, errorFn); err != nil {
		__antithesis_instrumentation__.Notify(237824)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237825)
	}
	__antithesis_instrumentation__.Notify(237808)
	return &response, nil
}

func mergeDistSQLRemoteFlows(a, b []serverpb.DistSQLRemoteFlows) []serverpb.DistSQLRemoteFlows {
	__antithesis_instrumentation__.Notify(237826)
	maxLength := len(a)
	if len(b) > len(a) {
		__antithesis_instrumentation__.Notify(237831)
		maxLength = len(b)
	} else {
		__antithesis_instrumentation__.Notify(237832)
	}
	__antithesis_instrumentation__.Notify(237827)
	result := make([]serverpb.DistSQLRemoteFlows, 0, maxLength)
	aIter, bIter := 0, 0
	for aIter < len(a) && func() bool {
		__antithesis_instrumentation__.Notify(237833)
		return bIter < len(b) == true
	}() == true {
		__antithesis_instrumentation__.Notify(237834)
		cmp := bytes.Compare(a[aIter].FlowID.GetBytes(), b[bIter].FlowID.GetBytes())
		if cmp < 0 {
			__antithesis_instrumentation__.Notify(237835)
			result = append(result, a[aIter])
			aIter++
		} else {
			__antithesis_instrumentation__.Notify(237836)
			if cmp > 0 {
				__antithesis_instrumentation__.Notify(237837)
				result = append(result, b[bIter])
				bIter++
			} else {
				__antithesis_instrumentation__.Notify(237838)
				r := a[aIter]

				r.Infos = append(r.Infos, b[bIter].Infos...)
				sort.Slice(r.Infos, func(i, j int) bool {
					__antithesis_instrumentation__.Notify(237840)
					return r.Infos[i].NodeID < r.Infos[j].NodeID
				})
				__antithesis_instrumentation__.Notify(237839)
				result = append(result, r)
				aIter++
				bIter++
			}
		}
	}
	__antithesis_instrumentation__.Notify(237828)
	if aIter < len(a) {
		__antithesis_instrumentation__.Notify(237841)
		result = append(result, a[aIter:]...)
	} else {
		__antithesis_instrumentation__.Notify(237842)
	}
	__antithesis_instrumentation__.Notify(237829)
	if bIter < len(b) {
		__antithesis_instrumentation__.Notify(237843)
		result = append(result, b[bIter:]...)
	} else {
		__antithesis_instrumentation__.Notify(237844)
	}
	__antithesis_instrumentation__.Notify(237830)
	return result
}

func (s *statusServer) SpanStats(
	ctx context.Context, req *serverpb.SpanStatsRequest,
) (*serverpb.SpanStatsResponse, error) {
	__antithesis_instrumentation__.Notify(237845)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237851)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237852)
	}
	__antithesis_instrumentation__.Notify(237846)

	nodeID, local, err := s.parseNodeID(req.NodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(237853)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237854)
	}
	__antithesis_instrumentation__.Notify(237847)

	if !local {
		__antithesis_instrumentation__.Notify(237855)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237857)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237858)
		}
		__antithesis_instrumentation__.Notify(237856)
		return status.SpanStats(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237859)
	}
	__antithesis_instrumentation__.Notify(237848)

	output := &serverpb.SpanStatsResponse{}
	err = s.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(237860)
		result, err := store.ComputeStatsForKeySpan(req.StartKey.Next(), req.EndKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(237862)
			return err
		} else {
			__antithesis_instrumentation__.Notify(237863)
		}
		__antithesis_instrumentation__.Notify(237861)
		output.TotalStats.Add(result.MVCC)
		output.RangeCount += int32(result.ReplicaCount)
		output.ApproximateDiskBytes += result.ApproximateDiskBytes
		return nil
	})
	__antithesis_instrumentation__.Notify(237849)
	if err != nil {
		__antithesis_instrumentation__.Notify(237864)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237865)
	}
	__antithesis_instrumentation__.Notify(237850)

	return output, nil
}

func (s *statusServer) Diagnostics(
	ctx context.Context, req *serverpb.DiagnosticsRequest,
) (*diagnosticspb.DiagnosticReport, error) {
	__antithesis_instrumentation__.Notify(237866)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)
	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237869)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237870)
	}
	__antithesis_instrumentation__.Notify(237867)

	if !local {
		__antithesis_instrumentation__.Notify(237871)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237873)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237874)
		}
		__antithesis_instrumentation__.Notify(237872)
		return status.Diagnostics(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237875)
	}
	__antithesis_instrumentation__.Notify(237868)

	return s.admin.server.sqlServer.diagnosticsReporter.CreateReport(ctx, telemetry.ReadOnly), nil
}

func (s *statusServer) Stores(
	ctx context.Context, req *serverpb.StoresRequest,
) (*serverpb.StoresResponse, error) {
	__antithesis_instrumentation__.Notify(237876)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237882)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237883)
	}
	__antithesis_instrumentation__.Notify(237877)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237884)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237885)
	}
	__antithesis_instrumentation__.Notify(237878)

	if !local {
		__antithesis_instrumentation__.Notify(237886)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237888)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237889)
		}
		__antithesis_instrumentation__.Notify(237887)
		return status.Stores(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237890)
	}
	__antithesis_instrumentation__.Notify(237879)

	resp := &serverpb.StoresResponse{}
	err = s.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(237891)
		storeDetails := serverpb.StoreDetails{
			StoreID: store.Ident.StoreID,
		}

		envStats, err := store.Engine().GetEnvStats()
		if err != nil {
			__antithesis_instrumentation__.Notify(237894)
			return err
		} else {
			__antithesis_instrumentation__.Notify(237895)
		}
		__antithesis_instrumentation__.Notify(237892)

		if len(envStats.EncryptionStatus) > 0 {
			__antithesis_instrumentation__.Notify(237896)
			storeDetails.EncryptionStatus = envStats.EncryptionStatus
		} else {
			__antithesis_instrumentation__.Notify(237897)
		}
		__antithesis_instrumentation__.Notify(237893)
		storeDetails.TotalFiles = envStats.TotalFiles
		storeDetails.TotalBytes = envStats.TotalBytes
		storeDetails.ActiveKeyFiles = envStats.ActiveKeyFiles
		storeDetails.ActiveKeyBytes = envStats.ActiveKeyBytes

		resp.Stores = append(resp.Stores, storeDetails)

		return nil
	})
	__antithesis_instrumentation__.Notify(237880)
	if err != nil {
		__antithesis_instrumentation__.Notify(237898)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237899)
	}
	__antithesis_instrumentation__.Notify(237881)
	return resp, nil
}

type jsonWrapper struct {
	Data interface{} `json:"d"`
}

func marshalToJSON(value interface{}) ([]byte, error) {
	__antithesis_instrumentation__.Notify(237900)
	switch reflect.ValueOf(value).Kind() {
	case reflect.Array, reflect.Slice:
		__antithesis_instrumentation__.Notify(237903)
		value = jsonWrapper{Data: value}
	default:
		__antithesis_instrumentation__.Notify(237904)
	}
	__antithesis_instrumentation__.Notify(237901)
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		__antithesis_instrumentation__.Notify(237905)
		return nil, errors.Wrapf(err, "unable to marshal %+v to json", value)
	} else {
		__antithesis_instrumentation__.Notify(237906)
	}
	__antithesis_instrumentation__.Notify(237902)
	return body, nil
}

func marshalJSONResponse(value interface{}) (*serverpb.JSONResponse, error) {
	__antithesis_instrumentation__.Notify(237907)
	data, err := marshalToJSON(value)
	if err != nil {
		__antithesis_instrumentation__.Notify(237909)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237910)
	}
	__antithesis_instrumentation__.Notify(237908)
	return &serverpb.JSONResponse{Data: data}, nil
}

func userFromContext(ctx context.Context) (res security.SQLUsername, err error) {
	__antithesis_instrumentation__.Notify(237911)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		__antithesis_instrumentation__.Notify(237915)
		return security.RootUserName(), nil
	} else {
		__antithesis_instrumentation__.Notify(237916)
	}
	__antithesis_instrumentation__.Notify(237912)
	usernames, ok := md[webSessionUserKeyStr]
	if !ok {
		__antithesis_instrumentation__.Notify(237917)

		return security.RootUserName(), nil
	} else {
		__antithesis_instrumentation__.Notify(237918)
	}
	__antithesis_instrumentation__.Notify(237913)
	if len(usernames) != 1 {
		__antithesis_instrumentation__.Notify(237919)
		log.Warningf(ctx, "context's incoming metadata contains unexpected number of usernames: %+v ", md)
		return res, fmt.Errorf(
			"context's incoming metadata contains unexpected number of usernames: %+v ", md)
	} else {
		__antithesis_instrumentation__.Notify(237920)
	}
	__antithesis_instrumentation__.Notify(237914)

	username := security.MakeSQLUsernameFromPreNormalizedString(usernames[0])
	return username, nil
}

type systemInfoOnce struct {
	once sync.Once
	info serverpb.SystemInfo
}

func (si *systemInfoOnce) systemInfo(ctx context.Context) serverpb.SystemInfo {
	__antithesis_instrumentation__.Notify(237921)

	si.once.Do(func() {
		__antithesis_instrumentation__.Notify(237923)

		cmd := exec.Command("uname", "-a")
		var errBuf bytes.Buffer
		cmd.Stderr = &errBuf
		output, err := cmd.Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(237926)
			log.Warningf(ctx, "failed to get system information: %v\nstderr: %v",
				err, errBuf.String())
			return
		} else {
			__antithesis_instrumentation__.Notify(237927)
		}
		__antithesis_instrumentation__.Notify(237924)
		si.info.SystemInfo = string(bytes.TrimSpace(output))
		cmd = exec.Command("uname", "-r")
		errBuf.Reset()
		cmd.Stderr = &errBuf
		output, err = cmd.Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(237928)
			log.Warningf(ctx, "failed to get kernel information: %v\nstderr: %v",
				err, errBuf.String())
			return
		} else {
			__antithesis_instrumentation__.Notify(237929)
		}
		__antithesis_instrumentation__.Notify(237925)
		si.info.KernelInfo = string(bytes.TrimSpace(output))
	})
	__antithesis_instrumentation__.Notify(237922)
	return si.info
}

func (s *statusServer) JobRegistryStatus(
	ctx context.Context, req *serverpb.JobRegistryStatusRequest,
) (*serverpb.JobRegistryStatusResponse, error) {
	__antithesis_instrumentation__.Notify(237930)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237935)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237936)
	}
	__antithesis_instrumentation__.Notify(237931)

	nodeID, local, err := s.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(237937)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237938)
	}
	__antithesis_instrumentation__.Notify(237932)
	if !local {
		__antithesis_instrumentation__.Notify(237939)
		status, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237941)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237942)
		}
		__antithesis_instrumentation__.Notify(237940)
		return status.JobRegistryStatus(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237943)
	}
	__antithesis_instrumentation__.Notify(237933)

	remoteNodeID := s.gossip.NodeID.Get()
	resp := &serverpb.JobRegistryStatusResponse{
		NodeID: remoteNodeID,
	}
	for _, jID := range s.admin.server.sqlServer.jobRegistry.CurrentlyRunningJobs() {
		__antithesis_instrumentation__.Notify(237944)
		job := serverpb.JobRegistryStatusResponse_Job{
			Id: int64(jID),
		}
		resp.RunningJobs = append(resp.RunningJobs, &job)
	}
	__antithesis_instrumentation__.Notify(237934)
	return resp, nil
}

func (s *statusServer) JobStatus(
	ctx context.Context, req *serverpb.JobStatusRequest,
) (*serverpb.JobStatusResponse, error) {
	__antithesis_instrumentation__.Notify(237945)
	ctx = s.AnnotateCtx(propagateGatewayMetadata(ctx))

	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237948)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237949)
	}
	__antithesis_instrumentation__.Notify(237946)

	j, err := s.admin.server.sqlServer.jobRegistry.LoadJob(ctx, jobspb.JobID(req.JobId))
	if err != nil {
		__antithesis_instrumentation__.Notify(237950)
		if je := (*jobs.JobNotFoundError)(nil); errors.As(err, &je) {
			__antithesis_instrumentation__.Notify(237952)
			return nil, status.Errorf(codes.NotFound, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(237953)
		}
		__antithesis_instrumentation__.Notify(237951)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237954)
	}
	__antithesis_instrumentation__.Notify(237947)
	res := &jobspb.Job{
		Payload:  &jobspb.Payload{},
		Progress: &jobspb.Progress{},
	}
	res.Id = j.ID()

	*res.Payload = j.Payload()
	*res.Progress = j.Progress()

	return &serverpb.JobStatusResponse{Job: res}, nil
}

func (s *statusServer) TxnIDResolution(
	ctx context.Context, req *serverpb.TxnIDResolutionRequest,
) (*serverpb.TxnIDResolutionResponse, error) {
	__antithesis_instrumentation__.Notify(237955)
	ctx = s.AnnotateCtx(propagateGatewayMetadata(ctx))
	if _, err := s.privilegeChecker.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237960)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237961)
	}
	__antithesis_instrumentation__.Notify(237956)

	requestedNodeID, local, err := s.parseNodeID(req.CoordinatorID)
	if err != nil {
		__antithesis_instrumentation__.Notify(237962)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(237963)
	}
	__antithesis_instrumentation__.Notify(237957)
	if local {
		__antithesis_instrumentation__.Notify(237964)
		return s.localTxnIDResolution(req), nil
	} else {
		__antithesis_instrumentation__.Notify(237965)
	}
	__antithesis_instrumentation__.Notify(237958)

	statusClient, err := s.dialNode(ctx, requestedNodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(237966)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237967)
	}
	__antithesis_instrumentation__.Notify(237959)

	return statusClient.TxnIDResolution(ctx, req)
}

func (s *statusServer) TransactionContentionEvents(
	ctx context.Context, req *serverpb.TransactionContentionEventsRequest,
) (*serverpb.TransactionContentionEventsResponse, error) {
	__antithesis_instrumentation__.Notify(237968)
	ctx = s.AnnotateCtx(propagateGatewayMetadata(ctx))

	if err := s.privilegeChecker.requireViewActivityOrViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(237978)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(237979)
	}
	__antithesis_instrumentation__.Notify(237969)

	user, isAdmin, err := s.privilegeChecker.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(237980)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(237981)
	}
	__antithesis_instrumentation__.Notify(237970)

	shouldRedactContendingKey := false
	if !isAdmin {
		__antithesis_instrumentation__.Notify(237982)
		shouldRedactContendingKey, err =
			s.privilegeChecker.hasRoleOption(ctx, user, roleoption.VIEWACTIVITYREDACTED)
		if err != nil {
			__antithesis_instrumentation__.Notify(237983)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(237984)
		}
	} else {
		__antithesis_instrumentation__.Notify(237985)
	}
	__antithesis_instrumentation__.Notify(237971)

	if s.gossip.NodeID.Get() == 0 {
		__antithesis_instrumentation__.Notify(237986)
		return nil, status.Errorf(codes.Unavailable, "nodeID not set")
	} else {
		__antithesis_instrumentation__.Notify(237987)
	}
	__antithesis_instrumentation__.Notify(237972)

	if len(req.NodeID) > 0 {
		__antithesis_instrumentation__.Notify(237988)
		requestedNodeID, local, err := s.parseNodeID(req.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237992)
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		} else {
			__antithesis_instrumentation__.Notify(237993)
		}
		__antithesis_instrumentation__.Notify(237989)
		if local {
			__antithesis_instrumentation__.Notify(237994)
			return s.localTransactionContentionEvents(shouldRedactContendingKey), nil
		} else {
			__antithesis_instrumentation__.Notify(237995)
		}
		__antithesis_instrumentation__.Notify(237990)

		statusClient, err := s.dialNode(ctx, requestedNodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(237996)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(237997)
		}
		__antithesis_instrumentation__.Notify(237991)
		return statusClient.TransactionContentionEvents(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(237998)
	}
	__antithesis_instrumentation__.Notify(237973)

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(237999)
		statusClient, err := s.dialNode(ctx, nodeID)
		return statusClient, err
	}
	__antithesis_instrumentation__.Notify(237974)

	rpcCallFn := func(ctx context.Context, client interface{}, _ roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(238000)
		statusClient := client.(serverpb.StatusClient)
		return statusClient.TransactionContentionEvents(ctx, &serverpb.TransactionContentionEventsRequest{
			NodeID: "local",
		})
	}
	__antithesis_instrumentation__.Notify(237975)

	resp := &serverpb.TransactionContentionEventsResponse{
		Events: make([]contentionpb.ExtendedContentionEvent, 0),
	}

	if err := s.iterateNodes(ctx, "txn contention events for node",
		dialFn,
		rpcCallFn,
		func(nodeID roachpb.NodeID, nodeResp interface{}) {
			__antithesis_instrumentation__.Notify(238001)
			txnContentionEvents := nodeResp.(*serverpb.TransactionContentionEventsResponse)
			resp.Events = append(resp.Events, txnContentionEvents.Events...)
		},
		func(nodeID roachpb.NodeID, nodeFnError error) {
			__antithesis_instrumentation__.Notify(238002)
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(238003)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238004)
	}
	__antithesis_instrumentation__.Notify(237976)

	sort.Slice(resp.Events, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(238005)
		return resp.Events[i].CollectionTs.Before(resp.Events[j].CollectionTs)
	})
	__antithesis_instrumentation__.Notify(237977)

	return resp, nil
}
