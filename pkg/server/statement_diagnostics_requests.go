package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type stmtDiagnosticsRequest struct {
	ID                     int
	StatementFingerprint   string
	Completed              bool
	StatementDiagnosticsID int
	RequestedAt            time.Time

	MinExecutionLatency time.Duration

	ExpiresAt time.Time
}

type stmtDiagnostics struct {
	ID                   int
	StatementFingerprint string
	CollectedAt          time.Time
}

func (request *stmtDiagnosticsRequest) toProto() serverpb.StatementDiagnosticsReport {
	__antithesis_instrumentation__.Notify(235311)
	resp := serverpb.StatementDiagnosticsReport{
		Id:                     int64(request.ID),
		Completed:              request.Completed,
		StatementFingerprint:   request.StatementFingerprint,
		StatementDiagnosticsId: int64(request.StatementDiagnosticsID),
		RequestedAt:            request.RequestedAt,
		MinExecutionLatency:    request.MinExecutionLatency,
		ExpiresAt:              request.ExpiresAt,
	}
	return resp
}

func (diagnostics *stmtDiagnostics) toProto() serverpb.StatementDiagnostics {
	__antithesis_instrumentation__.Notify(235312)
	resp := serverpb.StatementDiagnostics{
		Id:                   int64(diagnostics.ID),
		StatementFingerprint: diagnostics.StatementFingerprint,
		CollectedAt:          diagnostics.CollectedAt,
	}
	return resp
}

func (s *statusServer) CreateStatementDiagnosticsReport(
	ctx context.Context, req *serverpb.CreateStatementDiagnosticsReportRequest,
) (*serverpb.CreateStatementDiagnosticsReportResponse, error) {
	__antithesis_instrumentation__.Notify(235313)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityAndNoViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(235316)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235317)
	}
	__antithesis_instrumentation__.Notify(235314)

	response := &serverpb.CreateStatementDiagnosticsReportResponse{
		Report: &serverpb.StatementDiagnosticsReport{},
	}

	err := s.stmtDiagnosticsRequester.InsertRequest(
		ctx, req.StatementFingerprint, req.MinExecutionLatency, req.ExpiresAfter,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(235318)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235319)
	}
	__antithesis_instrumentation__.Notify(235315)

	response.Report.StatementFingerprint = req.StatementFingerprint
	return response, nil
}

func (s *statusServer) CancelStatementDiagnosticsReport(
	ctx context.Context, req *serverpb.CancelStatementDiagnosticsReportRequest,
) (*serverpb.CancelStatementDiagnosticsReportResponse, error) {
	__antithesis_instrumentation__.Notify(235320)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityAndNoViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(235323)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235324)
	}
	__antithesis_instrumentation__.Notify(235321)

	var response serverpb.CancelStatementDiagnosticsReportResponse
	err := s.stmtDiagnosticsRequester.CancelRequest(ctx, req.RequestID)
	if err != nil {
		__antithesis_instrumentation__.Notify(235325)
		response.Canceled = false
		response.Error = err.Error()
	} else {
		__antithesis_instrumentation__.Notify(235326)
		response.Canceled = true
	}
	__antithesis_instrumentation__.Notify(235322)
	return &response, nil
}

func (s *statusServer) StatementDiagnosticsRequests(
	ctx context.Context, _ *serverpb.StatementDiagnosticsReportsRequest,
) (*serverpb.StatementDiagnosticsReportsResponse, error) {
	__antithesis_instrumentation__.Notify(235327)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityAndNoViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(235334)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235335)
	}
	__antithesis_instrumentation__.Notify(235328)

	var err error

	var extraColumns string
	if s.admin.server.st.Version.IsActive(ctx, clusterversion.AlterSystemStmtDiagReqs) {
		__antithesis_instrumentation__.Notify(235336)
		extraColumns = `,
			min_execution_latency,
			expires_at`
	} else {
		__antithesis_instrumentation__.Notify(235337)
	}
	__antithesis_instrumentation__.Notify(235329)

	it, err := s.internalExecutor.QueryIteratorEx(ctx, "stmt-diag-get-all", nil,
		sessiondata.InternalExecutorOverride{
			User: security.RootUserName(),
		},
		fmt.Sprintf(`SELECT
			id,
			statement_fingerprint,
			completed,
			statement_diagnostics_id,
			requested_at%s
		FROM
			system.statement_diagnostics_requests`, extraColumns))
	if err != nil {
		__antithesis_instrumentation__.Notify(235338)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235339)
	}
	__antithesis_instrumentation__.Notify(235330)

	var requests []stmtDiagnosticsRequest
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(235340)
		row := it.Cur()
		id := int(*row[0].(*tree.DInt))
		statementFingerprint := string(*row[1].(*tree.DString))
		completed := bool(*row[2].(*tree.DBool))
		req := stmtDiagnosticsRequest{
			ID:                   id,
			StatementFingerprint: statementFingerprint,
			Completed:            completed,
		}
		if row[3] != tree.DNull {
			__antithesis_instrumentation__.Notify(235344)
			sdi := int(*row[3].(*tree.DInt))
			req.StatementDiagnosticsID = sdi
		} else {
			__antithesis_instrumentation__.Notify(235345)
		}
		__antithesis_instrumentation__.Notify(235341)
		if requestedAt, ok := row[4].(*tree.DTimestampTZ); ok {
			__antithesis_instrumentation__.Notify(235346)
			req.RequestedAt = requestedAt.Time
		} else {
			__antithesis_instrumentation__.Notify(235347)
		}
		__antithesis_instrumentation__.Notify(235342)
		if extraColumns != "" {
			__antithesis_instrumentation__.Notify(235348)
			if minExecutionLatency, ok := row[5].(*tree.DInterval); ok {
				__antithesis_instrumentation__.Notify(235350)
				req.MinExecutionLatency = time.Duration(minExecutionLatency.Duration.Nanos())
			} else {
				__antithesis_instrumentation__.Notify(235351)
			}
			__antithesis_instrumentation__.Notify(235349)
			if expiresAt, ok := row[6].(*tree.DTimestampTZ); ok {
				__antithesis_instrumentation__.Notify(235352)
				req.ExpiresAt = expiresAt.Time

				if req.ExpiresAt.Before(timeutil.Now()) {
					__antithesis_instrumentation__.Notify(235353)
					continue
				} else {
					__antithesis_instrumentation__.Notify(235354)
				}
			} else {
				__antithesis_instrumentation__.Notify(235355)
			}
		} else {
			__antithesis_instrumentation__.Notify(235356)
		}
		__antithesis_instrumentation__.Notify(235343)

		requests = append(requests, req)
	}
	__antithesis_instrumentation__.Notify(235331)
	if err != nil {
		__antithesis_instrumentation__.Notify(235357)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235358)
	}
	__antithesis_instrumentation__.Notify(235332)

	response := &serverpb.StatementDiagnosticsReportsResponse{
		Reports: make([]serverpb.StatementDiagnosticsReport, len(requests)),
	}

	for i, request := range requests {
		__antithesis_instrumentation__.Notify(235359)
		response.Reports[i] = request.toProto()
	}
	__antithesis_instrumentation__.Notify(235333)
	return response, nil
}

func (s *statusServer) StatementDiagnostics(
	ctx context.Context, req *serverpb.StatementDiagnosticsRequest,
) (*serverpb.StatementDiagnosticsResponse, error) {
	__antithesis_instrumentation__.Notify(235360)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.AnnotateCtx(ctx)

	if err := s.privilegeChecker.requireViewActivityAndNoViewActivityRedactedPermission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(235366)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235367)
	}
	__antithesis_instrumentation__.Notify(235361)

	var err error
	row, err := s.internalExecutor.QueryRowEx(ctx, "stmt-diag-get-one", nil,
		sessiondata.InternalExecutorOverride{
			User: security.RootUserName(),
		},
		`SELECT
			id,
			statement_fingerprint,
			collected_at
		FROM
			system.statement_diagnostics
		WHERE
			id = $1`, req.StatementDiagnosticsId)
	if err != nil {
		__antithesis_instrumentation__.Notify(235368)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(235369)
	}
	__antithesis_instrumentation__.Notify(235362)
	if row == nil {
		__antithesis_instrumentation__.Notify(235370)
		return nil, errors.Newf(
			"requested a statement diagnostic (%d) that does not exist",
			req.StatementDiagnosticsId,
		)
	} else {
		__antithesis_instrumentation__.Notify(235371)
	}
	__antithesis_instrumentation__.Notify(235363)

	diagnostics := stmtDiagnostics{
		ID: int(req.StatementDiagnosticsId),
	}

	if statementFingerprint, ok := row[1].(*tree.DString); ok {
		__antithesis_instrumentation__.Notify(235372)
		diagnostics.StatementFingerprint = statementFingerprint.String()
	} else {
		__antithesis_instrumentation__.Notify(235373)
	}
	__antithesis_instrumentation__.Notify(235364)

	if collectedAt, ok := row[2].(*tree.DTimestampTZ); ok {
		__antithesis_instrumentation__.Notify(235374)
		diagnostics.CollectedAt = collectedAt.Time
	} else {
		__antithesis_instrumentation__.Notify(235375)
	}
	__antithesis_instrumentation__.Notify(235365)

	diagnosticsProto := diagnostics.toProto()
	response := &serverpb.StatementDiagnosticsResponse{
		Diagnostics: &diagnosticsProto,
	}

	return response, nil
}
