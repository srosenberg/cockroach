package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net/http"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type usersResponse struct {
	serverpb.UsersResponse

	Next int `json:"next,omitempty"`
}

func (a *apiV2Server) listUsers(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189125)
	limit, offset := getSimplePaginationValues(r)
	ctx := r.Context()
	username := getSQLUsername(ctx)
	ctx = a.admin.server.AnnotateCtx(ctx)

	query := `SELECT username FROM system.users WHERE "isRole" = false ORDER BY username`
	qargs := []interface{}{}
	if limit > 0 {
		__antithesis_instrumentation__.Notify(189131)
		query += " LIMIT $"
		qargs = append(qargs, limit)
		if offset > 0 {
			__antithesis_instrumentation__.Notify(189132)
			query += " OFFSET $"
			qargs = append(qargs, offset)
		} else {
			__antithesis_instrumentation__.Notify(189133)
		}
	} else {
		__antithesis_instrumentation__.Notify(189134)
	}
	__antithesis_instrumentation__.Notify(189126)
	it, err := a.admin.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-users", nil,
		sessiondata.InternalExecutorOverride{User: username},
		query, qargs...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(189135)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(189136)
	}
	__antithesis_instrumentation__.Notify(189127)

	var resp usersResponse
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(189137)
		row := it.Cur()
		resp.Users = append(resp.Users, serverpb.UsersResponse_User{Username: string(tree.MustBeDString(row[0]))})
	}
	__antithesis_instrumentation__.Notify(189128)
	if err != nil {
		__antithesis_instrumentation__.Notify(189138)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(189139)
	}
	__antithesis_instrumentation__.Notify(189129)
	if limit > 0 && func() bool {
		__antithesis_instrumentation__.Notify(189140)
		return len(resp.Users) >= limit == true
	}() == true {
		__antithesis_instrumentation__.Notify(189141)
		resp.Next = offset + len(resp.Users)
	} else {
		__antithesis_instrumentation__.Notify(189142)
	}
	__antithesis_instrumentation__.Notify(189130)
	writeJSONResponse(ctx, w, 200, resp)
}

type eventsResponse struct {
	serverpb.EventsResponse

	Next int `json:"next,omitempty"`
}

func (a *apiV2Server) listEvents(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189143)
	limit, offset := getSimplePaginationValues(r)
	ctx := r.Context()
	username := getSQLUsername(ctx)
	ctx = a.admin.server.AnnotateCtx(ctx)
	queryValues := r.URL.Query()

	req := &serverpb.EventsRequest{}
	if typ := queryValues.Get("type"); len(typ) > 0 {
		__antithesis_instrumentation__.Notify(189148)
		req.Type = typ
	} else {
		__antithesis_instrumentation__.Notify(189149)
	}
	__antithesis_instrumentation__.Notify(189144)
	if targetID := queryValues.Get("targetID"); len(targetID) > 0 {
		__antithesis_instrumentation__.Notify(189150)
		if targetIDInt, err := strconv.ParseInt(targetID, 10, 64); err == nil {
			__antithesis_instrumentation__.Notify(189151)
			req.TargetId = targetIDInt
		} else {
			__antithesis_instrumentation__.Notify(189152)
		}
	} else {
		__antithesis_instrumentation__.Notify(189153)
	}
	__antithesis_instrumentation__.Notify(189145)

	var resp eventsResponse
	eventsResp, err := a.admin.eventsHelper(
		ctx, req, username, limit, offset, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(189154)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(189155)
	}
	__antithesis_instrumentation__.Notify(189146)
	resp.EventsResponse = *eventsResp
	if limit > 0 && func() bool {
		__antithesis_instrumentation__.Notify(189156)
		return len(resp.Events) >= limit == true
	}() == true {
		__antithesis_instrumentation__.Notify(189157)
		resp.Next = offset + len(resp.Events)
	} else {
		__antithesis_instrumentation__.Notify(189158)
	}
	__antithesis_instrumentation__.Notify(189147)
	writeJSONResponse(ctx, w, 200, resp)
}

type databasesResponse struct {
	serverpb.DatabasesResponse

	Next int `json:"next,omitempty"`
}

func (a *apiV2Server) listDatabases(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189159)
	limit, offset := getSimplePaginationValues(r)
	ctx := r.Context()
	username := getSQLUsername(ctx)
	ctx = a.admin.server.AnnotateCtx(ctx)

	var resp databasesResponse
	req := &serverpb.DatabasesRequest{}
	dbsResp, err := a.admin.databasesHelper(ctx, req, username, limit, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(189161)
		apiV2InternalError(ctx, err, w)
		return
	} else {
		__antithesis_instrumentation__.Notify(189162)
	}
	__antithesis_instrumentation__.Notify(189160)
	var databases interface{}
	databases, resp.Next = simplePaginate(dbsResp.Databases, limit, offset)
	resp.Databases = databases.([]string)
	writeJSONResponse(ctx, w, 200, resp)
}

type databaseDetailsResponse struct {
	DescriptorID int64 `json:"descriptor_id,omitempty"`
}

func (a *apiV2Server) databaseDetails(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189163)
	ctx := r.Context()
	username := getSQLUsername(ctx)
	ctx = a.admin.server.AnnotateCtx(ctx)
	pathVars := mux.Vars(r)
	req := &serverpb.DatabaseDetailsRequest{
		Database: pathVars["database_name"],
	}

	dbDetailsResp, err := a.admin.getMiscDatabaseDetails(ctx, req, username, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(189165)
		if status.Code(err) == codes.NotFound || func() bool {
			__antithesis_instrumentation__.Notify(189167)
			return isNotFoundError(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(189168)
			http.Error(w, "database not found", http.StatusNotFound)
		} else {
			__antithesis_instrumentation__.Notify(189169)
			apiV2InternalError(ctx, err, w)
		}
		__antithesis_instrumentation__.Notify(189166)
		return
	} else {
		__antithesis_instrumentation__.Notify(189170)
	}
	__antithesis_instrumentation__.Notify(189164)
	resp := databaseDetailsResponse{
		DescriptorID: dbDetailsResp.DescriptorID,
	}
	writeJSONResponse(ctx, w, 200, resp)
}

type databaseGrantsResponse struct {
	Grants []serverpb.DatabaseDetailsResponse_Grant `json:"grants"`

	Next int `json:"next,omitempty"`
}

func (a *apiV2Server) databaseGrants(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189171)
	ctx := r.Context()
	limit, offset := getSimplePaginationValues(r)
	username := getSQLUsername(ctx)
	ctx = a.admin.server.AnnotateCtx(ctx)
	pathVars := mux.Vars(r)
	req := &serverpb.DatabaseDetailsRequest{
		Database: pathVars["database_name"],
	}
	grants, err := a.admin.getDatabaseGrants(ctx, req, username, limit, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(189174)
		if status.Code(err) == codes.NotFound || func() bool {
			__antithesis_instrumentation__.Notify(189176)
			return isNotFoundError(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(189177)
			http.Error(w, "database not found", http.StatusNotFound)
		} else {
			__antithesis_instrumentation__.Notify(189178)
			apiV2InternalError(ctx, err, w)
		}
		__antithesis_instrumentation__.Notify(189175)
		return
	} else {
		__antithesis_instrumentation__.Notify(189179)
	}
	__antithesis_instrumentation__.Notify(189172)
	resp := databaseGrantsResponse{Grants: grants}
	if limit > 0 && func() bool {
		__antithesis_instrumentation__.Notify(189180)
		return len(grants) >= limit == true
	}() == true {
		__antithesis_instrumentation__.Notify(189181)
		resp.Next = offset + len(grants)
	} else {
		__antithesis_instrumentation__.Notify(189182)
	}
	__antithesis_instrumentation__.Notify(189173)
	writeJSONResponse(ctx, w, 200, resp)
}

type databaseTablesResponse struct {
	TableNames []string `json:"table_names,omitempty"`

	Next int `json:"next,omitempty"`
}

func (a *apiV2Server) databaseTables(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189183)
	ctx := r.Context()
	limit, offset := getSimplePaginationValues(r)
	username := getSQLUsername(ctx)
	ctx = a.admin.server.AnnotateCtx(ctx)
	pathVars := mux.Vars(r)
	req := &serverpb.DatabaseDetailsRequest{
		Database: pathVars["database_name"],
	}
	tables, err := a.admin.getDatabaseTables(ctx, req, username, limit, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(189186)
		if status.Code(err) == codes.NotFound || func() bool {
			__antithesis_instrumentation__.Notify(189188)
			return isNotFoundError(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(189189)
			http.Error(w, "database not found", http.StatusNotFound)
		} else {
			__antithesis_instrumentation__.Notify(189190)
			apiV2InternalError(ctx, err, w)
		}
		__antithesis_instrumentation__.Notify(189187)
		return
	} else {
		__antithesis_instrumentation__.Notify(189191)
	}
	__antithesis_instrumentation__.Notify(189184)
	resp := databaseTablesResponse{TableNames: tables}
	if limit > 0 && func() bool {
		__antithesis_instrumentation__.Notify(189192)
		return len(tables) >= limit == true
	}() == true {
		__antithesis_instrumentation__.Notify(189193)
		resp.Next = offset + len(tables)
	} else {
		__antithesis_instrumentation__.Notify(189194)
	}
	__antithesis_instrumentation__.Notify(189185)
	writeJSONResponse(ctx, w, 200, resp)
}

type tableDetailsResponse serverpb.TableDetailsResponse

func (a *apiV2Server) tableDetails(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(189195)
	ctx := r.Context()
	username := getSQLUsername(ctx)
	ctx = a.admin.server.AnnotateCtx(ctx)
	pathVars := mux.Vars(r)
	req := &serverpb.TableDetailsRequest{
		Database: pathVars["database_name"],
		Table:    pathVars["table_name"],
	}

	resp, err := a.admin.tableDetailsHelper(ctx, req, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(189197)
		if status.Code(err) == codes.NotFound || func() bool {
			__antithesis_instrumentation__.Notify(189199)
			return isNotFoundError(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(189200)
			http.Error(w, "database or table not found", http.StatusNotFound)
		} else {
			__antithesis_instrumentation__.Notify(189201)
			apiV2InternalError(ctx, err, w)
		}
		__antithesis_instrumentation__.Notify(189198)
		return
	} else {
		__antithesis_instrumentation__.Notify(189202)
	}
	__antithesis_instrumentation__.Notify(189196)
	writeJSONResponse(ctx, w, 200, tableDetailsResponse(*resp))
}
