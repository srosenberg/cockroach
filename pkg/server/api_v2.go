// CockroachDB v2 API
//
// API for querying information about CockroachDB health, nodes, ranges,
// sessions, and other meta entities.
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /api/v2/
//     Version: 2.0.0
//     License: Business Source License
//
//     Produces:
//     - application/json
//
//     SecurityDefinitions:
//       api_session:
//          type: apiKey
//          name: X-Cockroach-API-Session
//          description: Handle to logged-in REST session. Use `/login/` to
//            log in and get a session.
//          in: header
//
// swagger:meta
package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/gorilla/mux"
)

const (
	apiV2Path       = "/api/v2/"
	apiV2AuthHeader = "X-Cockroach-API-Session"
)

func writeJSONResponse(ctx context.Context, w http.ResponseWriter, code int, payload interface{}) {
	__antithesis_instrumentation__.Notify(188887)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	res, err := json.Marshal(payload)
	if err != nil {
		__antithesis_instrumentation__.Notify(188889)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188890)
	}
	__antithesis_instrumentation__.Notify(188888)
	_, _ = w.Write(res)
}

func getSQLUsername(ctx context.Context) security.SQLUsername {
	__antithesis_instrumentation__.Notify(188891)
	return security.MakeSQLUsernameFromPreNormalizedString(ctx.Value(webSessionUserKey{}).(string))
}

type apiV2Server struct {
	admin            *adminServer
	authServer       *authenticationV2Server
	status           *statusServer
	promRuleExporter *metric.PrometheusRuleExporter
	mux              *mux.Router
}

func newAPIV2Server(ctx context.Context, s *Server) *apiV2Server {
	__antithesis_instrumentation__.Notify(188892)
	authServer := newAuthenticationV2Server(ctx, s, apiV2Path)
	innerMux := mux.NewRouter()

	authMux := newAuthenticationV2Mux(authServer, innerMux)
	outerMux := mux.NewRouter()
	a := &apiV2Server{
		admin:            s.admin,
		authServer:       authServer,
		status:           s.status,
		mux:              outerMux,
		promRuleExporter: s.promRuleExporter,
	}
	a.registerRoutes(innerMux, authMux)
	return a
}

func (a *apiV2Server) registerRoutes(innerMux *mux.Router, authMux http.Handler) {
	__antithesis_instrumentation__.Notify(188893)
	var noOption roleoption.Option

	routeDefinitions := []struct {
		url          string
		handler      http.HandlerFunc
		requiresAuth bool
		role         apiRole
		option       roleoption.Option
	}{

		{"login/", a.authServer.ServeHTTP, false, regularRole, noOption},
		{"logout/", a.authServer.ServeHTTP, false, regularRole, noOption},

		{"sessions/", a.listSessions, true, adminRole, noOption},
		{"nodes/", a.listNodes, true, adminRole, noOption},

		{"nodes/{node_id}/ranges/", a.listNodeRanges, true, adminRole, noOption},
		{"ranges/hot/", a.listHotRanges, true, adminRole, noOption},
		{"ranges/{range_id:[0-9]+}/", a.listRange, true, adminRole, noOption},
		{"health/", a.health, false, regularRole, noOption},
		{"users/", a.listUsers, true, regularRole, noOption},
		{"events/", a.listEvents, true, adminRole, noOption},
		{"databases/", a.listDatabases, true, regularRole, noOption},
		{"databases/{database_name:[\\w.]+}/", a.databaseDetails, true, regularRole, noOption},
		{"databases/{database_name:[\\w.]+}/grants/", a.databaseGrants, true, regularRole, noOption},
		{"databases/{database_name:[\\w.]+}/tables/", a.databaseTables, true, regularRole, noOption},
		{"databases/{database_name:[\\w.]+}/tables/{table_name:[\\w.]+}/", a.tableDetails, true, regularRole, noOption},
		{"rules/", a.listRules, false, regularRole, noOption},
	}

	for _, route := range routeDefinitions {
		__antithesis_instrumentation__.Notify(188894)
		var handler http.Handler
		handler = &callCountDecorator{
			counter: telemetry.GetCounter(fmt.Sprintf("api.v2.%s", route.url)),
			inner:   route.handler,
		}
		if route.requiresAuth {
			__antithesis_instrumentation__.Notify(188895)
			a.mux.Handle(apiV2Path+route.url, authMux)
			if route.role != regularRole {
				__antithesis_instrumentation__.Notify(188897)
				handler = &roleAuthorizationMux{
					ie:     a.admin.ie,
					role:   route.role,
					option: route.option,
					inner:  handler,
				}
			} else {
				__antithesis_instrumentation__.Notify(188898)
			}
			__antithesis_instrumentation__.Notify(188896)
			innerMux.Handle(apiV2Path+route.url, handler)
		} else {
			__antithesis_instrumentation__.Notify(188899)
			a.mux.Handle(apiV2Path+route.url, handler)
		}
	}
}

func (a *apiV2Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	__antithesis_instrumentation__.Notify(188900)
	a.mux.ServeHTTP(w, req)
}

type callCountDecorator struct {
	counter telemetry.Counter
	inner   http.Handler
}

func (c *callCountDecorator) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	__antithesis_instrumentation__.Notify(188901)
	telemetry.Inc(c.counter)
	c.inner.ServeHTTP(w, req)
}

type listSessionsResponse struct {
	serverpb.ListSessionsResponse

	Next string `json:"next,omitempty"`
}

func (a *apiV2Server) listSessions(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(188902)
	ctx := r.Context()
	limit, start := getRPCPaginationValues(r)
	reqUsername := r.URL.Query().Get("username")
	req := &serverpb.ListSessionsRequest{Username: reqUsername}
	response := &listSessionsResponse{}
	outgoingCtx := apiToOutgoingGatewayCtx(ctx, r)

	responseProto, pagState, err := a.status.listSessionsHelper(outgoingCtx, req, limit, start)
	if err != nil {
		__antithesis_instrumentation__.Notify(188905)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188906)
	}
	__antithesis_instrumentation__.Notify(188903)
	var nextBytes []byte
	if nextBytes, err = pagState.MarshalText(); err != nil {
		__antithesis_instrumentation__.Notify(188907)
		err := serverpb.ListSessionsError{Message: err.Error()}
		response.Errors = append(response.Errors, err)
	} else {
		__antithesis_instrumentation__.Notify(188908)
		response.Next = string(nextBytes)
	}
	__antithesis_instrumentation__.Notify(188904)
	response.ListSessionsResponse = *responseProto
	writeJSONResponse(ctx, w, http.StatusOK, response)
}

func (a *apiV2Server) health(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(188909)
	ready := false
	readyStr := r.URL.Query().Get("ready")
	if len(readyStr) > 0 {
		__antithesis_instrumentation__.Notify(188913)
		var err error
		ready, err = strconv.ParseBool(readyStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(188914)
			http.Error(w, "invalid ready value", http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(188915)
		}
	} else {
		__antithesis_instrumentation__.Notify(188916)
	}
	__antithesis_instrumentation__.Notify(188910)
	ctx := r.Context()
	resp := &serverpb.HealthResponse{}

	if !ready {
		__antithesis_instrumentation__.Notify(188917)
		writeJSONResponse(ctx, w, 200, resp)
		return
	} else {
		__antithesis_instrumentation__.Notify(188918)
	}
	__antithesis_instrumentation__.Notify(188911)

	if err := a.admin.checkReadinessForHealthCheck(ctx); err != nil {
		__antithesis_instrumentation__.Notify(188919)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188920)
	}
	__antithesis_instrumentation__.Notify(188912)
	writeJSONResponse(ctx, w, 200, resp)
}

func (a *apiV2Server) listRules(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(188921)
	a.promRuleExporter.ScrapeRegistry(r.Context())
	response, err := a.promRuleExporter.PrintAsYAML()
	if err != nil {
		__antithesis_instrumentation__.Notify(188923)
		apiV2InternalError(r.Context(), err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(188924)
	}
	__antithesis_instrumentation__.Notify(188922)
	w.Header().Set(httputil.ContentTypeHeader, httputil.PlaintextContentType)
	_, _ = w.Write(response)
}
