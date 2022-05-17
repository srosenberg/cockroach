package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/ts/catalog"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingui"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	gwutil "github.com/grpc-ecosystem/grpc-gateway/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	adminPrefix = "/_admin/v1/"

	adminHealth = adminPrefix + "health"

	defaultAPIEventLimit = 1000
)

func nonTableDescriptorRangeCount() int64 {
	__antithesis_instrumentation__.Notify(187535)

	return int64(len([]int{
		keys.MetaRangesID,
		keys.SystemRangesID,
		keys.TimeseriesRangesID,
		keys.LivenessRangesID,
		keys.PublicSchemaID,
		keys.TenantsRangesID,
	}))
}

var errAdminAPIError = status.Errorf(codes.Internal, "An internal server error "+
	"has occurred. Please check your CockroachDB logs for more details.")

type adminServer struct {
	*adminPrivilegeChecker
	internalExecutor *sql.InternalExecutor
	server           *Server
	memMonitor       *mon.BytesMonitor
}

var noteworthyAdminMemoryUsageBytes = envutil.EnvOrDefaultInt64("COCKROACH_NOTEWORTHY_ADMIN_MEMORY_USAGE", 100*1024)

func newAdminServer(
	s *Server, adminAuthzCheck *adminPrivilegeChecker, ie *sql.InternalExecutor,
) *adminServer {
	__antithesis_instrumentation__.Notify(187536)
	server := &adminServer{
		adminPrivilegeChecker: adminAuthzCheck,
		internalExecutor:      ie,
		server:                s,
	}

	server.memMonitor = mon.NewUnlimitedMonitor(
		context.Background(),
		"admin",
		mon.MemoryResource,
		nil,
		nil,
		noteworthyAdminMemoryUsageBytes,
		s.ClusterSettings(),
	)
	return server
}

func (s *adminServer) RegisterService(g *grpc.Server) {
	__antithesis_instrumentation__.Notify(187537)
	serverpb.RegisterAdminServer(g, s)
}

func (s *adminServer) RegisterGateway(
	ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn,
) error {
	__antithesis_instrumentation__.Notify(187538)

	stmtBundlePattern := gwruntime.MustPattern(gwruntime.NewPattern(
		1,
		[]int{
			int(gwutil.OpLitPush), 0, int(gwutil.OpLitPush), 1, int(gwutil.OpLitPush), 2,
			int(gwutil.OpPush), 0, int(gwutil.OpConcatN), 1, int(gwutil.OpCapture), 3},
		[]string{"_admin", "v1", "stmtbundle", "id"},
		"",
	))

	mux.Handle("GET", stmtBundlePattern, func(
		w http.ResponseWriter, req *http.Request, pathParams map[string]string,
	) {
		__antithesis_instrumentation__.Notify(187540)
		idStr, ok := pathParams["id"]
		if !ok {
			__antithesis_instrumentation__.Notify(187543)
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(187544)
		}
		__antithesis_instrumentation__.Notify(187541)
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(187545)
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		} else {
			__antithesis_instrumentation__.Notify(187546)
		}
		__antithesis_instrumentation__.Notify(187542)
		s.getStatementBundle(ctx, id, w)
	})
	__antithesis_instrumentation__.Notify(187539)

	return serverpb.RegisterAdminHandler(ctx, mux, conn)
}

func serverError(ctx context.Context, err error) error {
	__antithesis_instrumentation__.Notify(187547)
	log.ErrorfDepth(ctx, 1, "%+s", err)
	return errAdminAPIError
}

func serverErrorf(ctx context.Context, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(187548)
	log.ErrorfDepth(ctx, 1, format, args...)
	return errAdminAPIError
}

func isNotFoundError(err error) bool {
	__antithesis_instrumentation__.Notify(187549)

	return err != nil && func() bool {
		__antithesis_instrumentation__.Notify(187550)
		return strings.HasSuffix(err.Error(), "does not exist") == true
	}() == true
}

func (s *adminServer) AllMetricMetadata(
	ctx context.Context, req *serverpb.MetricMetadataRequest,
) (*serverpb.MetricMetadataResponse, error) {
	__antithesis_instrumentation__.Notify(187551)

	resp := &serverpb.MetricMetadataResponse{
		Metadata: s.server.recorder.GetMetricsMetadata(),
	}

	return resp, nil
}

func (s *adminServer) ChartCatalog(
	ctx context.Context, req *serverpb.ChartCatalogRequest,
) (*serverpb.ChartCatalogResponse, error) {
	__antithesis_instrumentation__.Notify(187552)
	metricsMetadata := s.server.recorder.GetMetricsMetadata()

	chartCatalog, err := catalog.GenerateCatalog(metricsMetadata)
	if err != nil {
		__antithesis_instrumentation__.Notify(187554)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187555)
	}
	__antithesis_instrumentation__.Notify(187553)

	resp := &serverpb.ChartCatalogResponse{
		Catalog: chartCatalog,
	}

	return resp, nil
}

func (s *adminServer) Databases(
	ctx context.Context, req *serverpb.DatabasesRequest,
) (_ *serverpb.DatabasesResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187556)
	ctx = s.server.AnnotateCtx(ctx)

	sessionUser, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(187558)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187559)
	}
	__antithesis_instrumentation__.Notify(187557)

	r, err := s.databasesHelper(ctx, req, sessionUser, 0, 0)
	return r, maybeHandleNotFoundError(ctx, err)
}

func (s *adminServer) databasesHelper(
	ctx context.Context,
	req *serverpb.DatabasesRequest,
	sessionUser security.SQLUsername,
	limit, offset int,
) (_ *serverpb.DatabasesResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187560)
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-show-dbs", nil,
		sessiondata.InternalExecutorOverride{User: sessionUser},
		"SHOW DATABASES",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(187565)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187566)
	}
	__antithesis_instrumentation__.Notify(187561)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(187567)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	__antithesis_instrumentation__.Notify(187562)

	var resp serverpb.DatabasesResponse
	var hasNext bool
	for hasNext, err = it.Next(ctx); hasNext; hasNext, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(187568)
		row := it.Cur()
		dbDatum, ok := tree.AsDString(row[0])
		if !ok {
			__antithesis_instrumentation__.Notify(187570)
			return nil, errors.Errorf("type assertion failed on db name: %T", row[0])
		} else {
			__antithesis_instrumentation__.Notify(187571)
		}
		__antithesis_instrumentation__.Notify(187569)
		dbName := string(dbDatum)
		resp.Databases = append(resp.Databases, dbName)
	}
	__antithesis_instrumentation__.Notify(187563)
	if err != nil {
		__antithesis_instrumentation__.Notify(187572)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187573)
	}
	__antithesis_instrumentation__.Notify(187564)

	return &resp, nil
}

func maybeHandleNotFoundError(ctx context.Context, err error) error {
	__antithesis_instrumentation__.Notify(187574)
	if err == nil {
		__antithesis_instrumentation__.Notify(187577)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(187578)
	}
	__antithesis_instrumentation__.Notify(187575)
	if isNotFoundError(err) {
		__antithesis_instrumentation__.Notify(187579)
		return status.Errorf(codes.NotFound, "%s", err)
	} else {
		__antithesis_instrumentation__.Notify(187580)
	}
	__antithesis_instrumentation__.Notify(187576)
	return serverError(ctx, err)
}

func (s *adminServer) DatabaseDetails(
	ctx context.Context, req *serverpb.DatabaseDetailsRequest,
) (_ *serverpb.DatabaseDetailsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187581)
	ctx = s.server.AnnotateCtx(ctx)
	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(187583)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187584)
	}
	__antithesis_instrumentation__.Notify(187582)

	r, err := s.databaseDetailsHelper(ctx, req, userName)
	return r, maybeHandleNotFoundError(ctx, err)
}

func (s *adminServer) getDatabaseGrants(
	ctx context.Context,
	req *serverpb.DatabaseDetailsRequest,
	userName security.SQLUsername,
	limit, offset int,
) (resp []serverpb.DatabaseDetailsResponse_Grant, retErr error) {
	__antithesis_instrumentation__.Notify(187585)
	escDBName := tree.NameStringP(&req.Database)

	query := makeSQLQuery()

	query.Append(fmt.Sprintf("SELECT * FROM [SHOW GRANTS ON DATABASE %s]", escDBName))
	if limit > 0 {
		__antithesis_instrumentation__.Notify(187589)
		query.Append(" LIMIT $", limit)
		if offset > 0 {
			__antithesis_instrumentation__.Notify(187590)
			query.Append(" OFFSET $", offset)
		} else {
			__antithesis_instrumentation__.Notify(187591)
		}
	} else {
		__antithesis_instrumentation__.Notify(187592)
	}
	__antithesis_instrumentation__.Notify(187586)
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-show-grants", nil,
		sessiondata.InternalExecutorOverride{User: userName},

		query.String(), query.QueryArguments()...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(187593)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187594)
	}
	__antithesis_instrumentation__.Notify(187587)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(187595)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	{
		__antithesis_instrumentation__.Notify(187596)
		const (
			userCol       = "grantee"
			privilegesCol = "privilege_type"
		)

		ok, err := it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(187598)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187599)
		}
		__antithesis_instrumentation__.Notify(187597)
		if ok {
			__antithesis_instrumentation__.Notify(187600)

			scanner := makeResultScanner(it.Types())
			for ; ok; ok, err = it.Next(ctx) {
				__antithesis_instrumentation__.Notify(187602)
				row := it.Cur()

				var grant serverpb.DatabaseDetailsResponse_Grant
				var privileges string
				if err := scanner.Scan(row, userCol, &grant.User); err != nil {
					__antithesis_instrumentation__.Notify(187605)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187606)
				}
				__antithesis_instrumentation__.Notify(187603)
				if err := scanner.Scan(row, privilegesCol, &privileges); err != nil {
					__antithesis_instrumentation__.Notify(187607)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187608)
				}
				__antithesis_instrumentation__.Notify(187604)
				grant.Privileges = strings.Split(privileges, ",")
				resp = append(resp, grant)
			}
			__antithesis_instrumentation__.Notify(187601)
			if err = maybeHandleNotFoundError(ctx, err); err != nil {
				__antithesis_instrumentation__.Notify(187609)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(187610)
			}
		} else {
			__antithesis_instrumentation__.Notify(187611)
		}
	}
	__antithesis_instrumentation__.Notify(187588)
	return resp, nil
}

func (s *adminServer) getDatabaseTables(
	ctx context.Context,
	req *serverpb.DatabaseDetailsRequest,
	userName security.SQLUsername,
	limit, offset int,
) (resp []string, retErr error) {
	__antithesis_instrumentation__.Notify(187612)
	query := makeSQLQuery()
	query.Append(`SELECT table_schema, table_name FROM information_schema.tables
WHERE table_catalog = $ AND table_type != 'SYSTEM VIEW'`, req.Database)
	query.Append(" ORDER BY table_name")
	if limit > 0 {
		__antithesis_instrumentation__.Notify(187616)
		query.Append(" LIMIT $", limit)
		if offset > 0 {
			__antithesis_instrumentation__.Notify(187617)
			query.Append(" OFFSET $", offset)
		} else {
			__antithesis_instrumentation__.Notify(187618)
		}
	} else {
		__antithesis_instrumentation__.Notify(187619)
	}
	__antithesis_instrumentation__.Notify(187613)

	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-show-tables", nil,
		sessiondata.InternalExecutorOverride{User: userName, Database: req.Database},
		query.String(), query.QueryArguments()...)
	if err != nil {
		__antithesis_instrumentation__.Notify(187620)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187621)
	}
	__antithesis_instrumentation__.Notify(187614)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(187622)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	{
		__antithesis_instrumentation__.Notify(187623)
		ok, err := it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(187625)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187626)
		}
		__antithesis_instrumentation__.Notify(187624)

		if ok {
			__antithesis_instrumentation__.Notify(187627)

			scanner := makeResultScanner(it.Types())
			for ; ok; ok, err = it.Next(ctx) {
				__antithesis_instrumentation__.Notify(187629)
				row := it.Cur()
				var schemaName, tableName string
				if err := scanner.Scan(row, "table_schema", &schemaName); err != nil {
					__antithesis_instrumentation__.Notify(187632)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187633)
				}
				__antithesis_instrumentation__.Notify(187630)
				if err := scanner.Scan(row, "table_name", &tableName); err != nil {
					__antithesis_instrumentation__.Notify(187634)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187635)
				}
				__antithesis_instrumentation__.Notify(187631)
				resp = append(resp, fmt.Sprintf("%s.%s",
					tree.NameStringP(&schemaName), tree.NameStringP(&tableName)))
			}
			__antithesis_instrumentation__.Notify(187628)
			if err != nil {
				__antithesis_instrumentation__.Notify(187636)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(187637)
			}
		} else {
			__antithesis_instrumentation__.Notify(187638)
		}
	}
	__antithesis_instrumentation__.Notify(187615)
	return resp, nil
}

func (s *adminServer) getMiscDatabaseDetails(
	ctx context.Context,
	req *serverpb.DatabaseDetailsRequest,
	userName security.SQLUsername,
	resp *serverpb.DatabaseDetailsResponse,
) (*serverpb.DatabaseDetailsResponse, error) {
	__antithesis_instrumentation__.Notify(187639)
	if resp == nil {
		__antithesis_instrumentation__.Notify(187645)
		resp = &serverpb.DatabaseDetailsResponse{}
	} else {
		__antithesis_instrumentation__.Notify(187646)
	}
	__antithesis_instrumentation__.Notify(187640)

	databaseID, err := s.queryDatabaseID(ctx, userName, req.Database)
	if err != nil {
		__antithesis_instrumentation__.Notify(187647)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187648)
	}
	__antithesis_instrumentation__.Notify(187641)
	resp.DescriptorID = int64(databaseID)

	id, zone, zoneExists, err := s.queryZonePath(ctx, userName, []descpb.ID{databaseID})
	if err != nil {
		__antithesis_instrumentation__.Notify(187649)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187650)
	}
	__antithesis_instrumentation__.Notify(187642)

	if !zoneExists {
		__antithesis_instrumentation__.Notify(187651)
		zone = s.server.cfg.DefaultZoneConfig
	} else {
		__antithesis_instrumentation__.Notify(187652)
	}
	__antithesis_instrumentation__.Notify(187643)
	resp.ZoneConfig = zone

	switch id {
	case databaseID:
		__antithesis_instrumentation__.Notify(187653)
		resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_DATABASE
	default:
		__antithesis_instrumentation__.Notify(187654)
		resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_CLUSTER
	}
	__antithesis_instrumentation__.Notify(187644)
	return resp, nil
}

func (s *adminServer) databaseDetailsHelper(
	ctx context.Context, req *serverpb.DatabaseDetailsRequest, userName security.SQLUsername,
) (_ *serverpb.DatabaseDetailsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187655)
	var resp serverpb.DatabaseDetailsResponse
	var err error

	resp.Grants, err = s.getDatabaseGrants(ctx, req, userName, 0, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(187660)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187661)
	}
	__antithesis_instrumentation__.Notify(187656)
	resp.TableNames, err = s.getDatabaseTables(ctx, req, userName, 0, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(187662)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187663)
	}
	__antithesis_instrumentation__.Notify(187657)

	_, err = s.getMiscDatabaseDetails(ctx, req, userName, &resp)
	if err != nil {
		__antithesis_instrumentation__.Notify(187664)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187665)
	}
	__antithesis_instrumentation__.Notify(187658)

	if req.IncludeStats {
		__antithesis_instrumentation__.Notify(187666)
		tableSpans, err := s.getDatabaseTableSpans(ctx, userName, req.Database, resp.TableNames)
		if err != nil {
			__antithesis_instrumentation__.Notify(187668)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187669)
		}
		__antithesis_instrumentation__.Notify(187667)
		resp.Stats, err = s.getDatabaseStats(ctx, tableSpans)
		if err != nil {
			__antithesis_instrumentation__.Notify(187670)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187671)
		}
	} else {
		__antithesis_instrumentation__.Notify(187672)
	}
	__antithesis_instrumentation__.Notify(187659)

	return &resp, nil
}

func (s *adminServer) getDatabaseTableSpans(
	ctx context.Context, userName security.SQLUsername, dbName string, tableNames []string,
) (map[string]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(187673)
	tableSpans := make(map[string]roachpb.Span, len(tableNames))

	for _, tableName := range tableNames {
		__antithesis_instrumentation__.Notify(187675)
		fullyQualifiedTableName, err := getFullyQualifiedTableName(dbName, tableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(187678)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187679)
		}
		__antithesis_instrumentation__.Notify(187676)
		tableID, err := s.queryTableID(ctx, userName, dbName, fullyQualifiedTableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(187680)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187681)
		}
		__antithesis_instrumentation__.Notify(187677)
		tableSpans[tableName] = generateTableSpan(tableID)
	}
	__antithesis_instrumentation__.Notify(187674)
	return tableSpans, nil
}

func (s *adminServer) getDatabaseStats(
	ctx context.Context, tableSpans map[string]roachpb.Span,
) (*serverpb.DatabaseDetailsResponse_Stats, error) {
	__antithesis_instrumentation__.Notify(187682)
	var stats serverpb.DatabaseDetailsResponse_Stats

	type tableStatsResponse struct {
		name string
		resp *serverpb.TableStatsResponse
		err  error
	}

	responses := make(chan tableStatsResponse, len(tableSpans))

	for tableName, tableSpan := range tableSpans {
		__antithesis_instrumentation__.Notify(187687)

		tableName := tableName
		tableSpan := tableSpan
		if err := s.server.stopper.RunAsyncTask(
			ctx, "server.adminServer: requesting table stats",
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(187688)
				statsResponse, err := s.statsForSpan(ctx, tableSpan)

				responses <- tableStatsResponse{
					name: tableName,
					resp: statsResponse,
					err:  err,
				}
			}); err != nil {
			__antithesis_instrumentation__.Notify(187689)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187690)
		}
	}
	__antithesis_instrumentation__.Notify(187683)

	nodeIDs := make(map[roachpb.NodeID]struct{})
	for i := 0; i < len(tableSpans); i++ {
		__antithesis_instrumentation__.Notify(187691)
		select {
		case response := <-responses:
			__antithesis_instrumentation__.Notify(187692)
			if response.err != nil {
				__antithesis_instrumentation__.Notify(187694)
				stats.MissingTables = append(
					stats.MissingTables,
					&serverpb.DatabaseDetailsResponse_Stats_MissingTable{
						Name:         response.name,
						ErrorMessage: response.err.Error(),
					})
			} else {
				__antithesis_instrumentation__.Notify(187695)
				stats.RangeCount += response.resp.RangeCount
				stats.ApproximateDiskBytes += response.resp.ApproximateDiskBytes
				for _, id := range response.resp.NodeIDs {
					__antithesis_instrumentation__.Notify(187696)
					nodeIDs[id] = struct{}{}
				}
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(187693)
			return nil, ctx.Err()
		}
	}
	__antithesis_instrumentation__.Notify(187684)

	stats.NodeIDs = make([]roachpb.NodeID, 0, len(nodeIDs))
	for id := range nodeIDs {
		__antithesis_instrumentation__.Notify(187697)
		stats.NodeIDs = append(stats.NodeIDs, id)
	}
	__antithesis_instrumentation__.Notify(187685)
	sort.Slice(stats.NodeIDs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(187698)
		return stats.NodeIDs[i] < stats.NodeIDs[j]
	})
	__antithesis_instrumentation__.Notify(187686)

	return &stats, nil
}

func getFullyQualifiedTableName(dbName string, tableName string) (string, error) {
	__antithesis_instrumentation__.Notify(187699)
	name, err := parser.ParseQualifiedTableName(tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(187702)

		name, err = parser.ParseQualifiedTableName(tree.NameStringP(&tableName))
		if err != nil {
			__antithesis_instrumentation__.Notify(187703)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(187704)
		}
	} else {
		__antithesis_instrumentation__.Notify(187705)
	}
	__antithesis_instrumentation__.Notify(187700)
	if !name.ExplicitSchema {
		__antithesis_instrumentation__.Notify(187706)

		name.SchemaName = tree.Name(dbName)
		name.ExplicitSchema = true
	} else {
		__antithesis_instrumentation__.Notify(187707)

		name.CatalogName = tree.Name(dbName)
		name.ExplicitCatalog = true
	}
	__antithesis_instrumentation__.Notify(187701)
	return name.String(), nil
}

func (s *adminServer) TableDetails(
	ctx context.Context, req *serverpb.TableDetailsRequest,
) (_ *serverpb.TableDetailsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187708)
	ctx = s.server.AnnotateCtx(ctx)
	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(187710)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187711)
	}
	__antithesis_instrumentation__.Notify(187709)

	r, err := s.tableDetailsHelper(ctx, req, userName)
	return r, maybeHandleNotFoundError(ctx, err)
}

func (s *adminServer) tableDetailsHelper(
	ctx context.Context, req *serverpb.TableDetailsRequest, userName security.SQLUsername,
) (_ *serverpb.TableDetailsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187712)
	escQualTable, err := getFullyQualifiedTableName(req.Database, req.Table)
	if err != nil {
		__antithesis_instrumentation__.Notify(187725)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187726)
	}
	__antithesis_instrumentation__.Notify(187713)

	var resp serverpb.TableDetailsResponse

	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-show-columns",
		nil,
		sessiondata.InternalExecutorOverride{User: userName},
		fmt.Sprintf("SHOW COLUMNS FROM %s", escQualTable),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(187727)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187728)
	}
	__antithesis_instrumentation__.Notify(187714)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(187729)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)

	{
		__antithesis_instrumentation__.Notify(187730)
		const (
			colCol     = "column_name"
			typeCol    = "data_type"
			nullCol    = "is_nullable"
			defaultCol = "column_default"
			genCol     = "generation_expression"
			hiddenCol  = "is_hidden"
		)
		ok, err := it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(187732)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187733)
		}
		__antithesis_instrumentation__.Notify(187731)
		if ok {
			__antithesis_instrumentation__.Notify(187734)

			scanner := makeResultScanner(it.Types())
			for ; ok; ok, err = it.Next(ctx) {
				__antithesis_instrumentation__.Notify(187736)
				row := it.Cur()
				var col serverpb.TableDetailsResponse_Column
				if err := scanner.Scan(row, colCol, &col.Name); err != nil {
					__antithesis_instrumentation__.Notify(187745)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187746)
				}
				__antithesis_instrumentation__.Notify(187737)
				if err := scanner.Scan(row, typeCol, &col.Type); err != nil {
					__antithesis_instrumentation__.Notify(187747)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187748)
				}
				__antithesis_instrumentation__.Notify(187738)
				if err := scanner.Scan(row, nullCol, &col.Nullable); err != nil {
					__antithesis_instrumentation__.Notify(187749)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187750)
				}
				__antithesis_instrumentation__.Notify(187739)
				if err := scanner.Scan(row, hiddenCol, &col.Hidden); err != nil {
					__antithesis_instrumentation__.Notify(187751)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187752)
				}
				__antithesis_instrumentation__.Notify(187740)
				isDefaultNull, err := scanner.IsNull(row, defaultCol)
				if err != nil {
					__antithesis_instrumentation__.Notify(187753)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187754)
				}
				__antithesis_instrumentation__.Notify(187741)
				if !isDefaultNull {
					__antithesis_instrumentation__.Notify(187755)
					if err := scanner.Scan(row, defaultCol, &col.DefaultValue); err != nil {
						__antithesis_instrumentation__.Notify(187756)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(187757)
					}
				} else {
					__antithesis_instrumentation__.Notify(187758)
				}
				__antithesis_instrumentation__.Notify(187742)
				isGenNull, err := scanner.IsNull(row, genCol)
				if err != nil {
					__antithesis_instrumentation__.Notify(187759)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187760)
				}
				__antithesis_instrumentation__.Notify(187743)
				if !isGenNull {
					__antithesis_instrumentation__.Notify(187761)
					if err := scanner.Scan(row, genCol, &col.GenerationExpression); err != nil {
						__antithesis_instrumentation__.Notify(187762)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(187763)
					}
				} else {
					__antithesis_instrumentation__.Notify(187764)
				}
				__antithesis_instrumentation__.Notify(187744)
				resp.Columns = append(resp.Columns, col)
			}
			__antithesis_instrumentation__.Notify(187735)
			if err != nil {
				__antithesis_instrumentation__.Notify(187765)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(187766)
			}
		} else {
			__antithesis_instrumentation__.Notify(187767)
		}
	}
	__antithesis_instrumentation__.Notify(187715)

	it, err = s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-showindex", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		fmt.Sprintf("SHOW INDEX FROM %s", escQualTable),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(187768)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187769)
	}
	__antithesis_instrumentation__.Notify(187716)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(187770)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	{
		__antithesis_instrumentation__.Notify(187771)
		const (
			nameCol      = "index_name"
			nonUniqueCol = "non_unique"
			seqCol       = "seq_in_index"
			columnCol    = "column_name"
			directionCol = "direction"
			storingCol   = "storing"
			implicitCol  = "implicit"
		)
		ok, err := it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(187773)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187774)
		}
		__antithesis_instrumentation__.Notify(187772)
		if ok {
			__antithesis_instrumentation__.Notify(187775)

			scanner := makeResultScanner(it.Types())
			for ; ok; ok, err = it.Next(ctx) {
				__antithesis_instrumentation__.Notify(187777)
				row := it.Cur()

				var index serverpb.TableDetailsResponse_Index
				if err := scanner.Scan(row, nameCol, &index.Name); err != nil {
					__antithesis_instrumentation__.Notify(187785)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187786)
				}
				__antithesis_instrumentation__.Notify(187778)
				var nonUnique bool
				if err := scanner.Scan(row, nonUniqueCol, &nonUnique); err != nil {
					__antithesis_instrumentation__.Notify(187787)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187788)
				}
				__antithesis_instrumentation__.Notify(187779)
				index.Unique = !nonUnique
				if err := scanner.Scan(row, seqCol, &index.Seq); err != nil {
					__antithesis_instrumentation__.Notify(187789)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187790)
				}
				__antithesis_instrumentation__.Notify(187780)
				if err := scanner.Scan(row, columnCol, &index.Column); err != nil {
					__antithesis_instrumentation__.Notify(187791)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187792)
				}
				__antithesis_instrumentation__.Notify(187781)
				if err := scanner.Scan(row, directionCol, &index.Direction); err != nil {
					__antithesis_instrumentation__.Notify(187793)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187794)
				}
				__antithesis_instrumentation__.Notify(187782)
				if err := scanner.Scan(row, storingCol, &index.Storing); err != nil {
					__antithesis_instrumentation__.Notify(187795)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187796)
				}
				__antithesis_instrumentation__.Notify(187783)
				if err := scanner.Scan(row, implicitCol, &index.Implicit); err != nil {
					__antithesis_instrumentation__.Notify(187797)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187798)
				}
				__antithesis_instrumentation__.Notify(187784)
				resp.Indexes = append(resp.Indexes, index)
			}
			__antithesis_instrumentation__.Notify(187776)
			if err != nil {
				__antithesis_instrumentation__.Notify(187799)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(187800)
			}
		} else {
			__antithesis_instrumentation__.Notify(187801)
		}
	}
	__antithesis_instrumentation__.Notify(187717)

	it, err = s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-show-grants", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		fmt.Sprintf("SHOW GRANTS ON TABLE %s", escQualTable),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(187802)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187803)
	}
	__antithesis_instrumentation__.Notify(187718)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(187804)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	{
		__antithesis_instrumentation__.Notify(187805)
		const (
			userCol       = "grantee"
			privilegesCol = "privilege_type"
		)
		ok, err := it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(187807)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187808)
		}
		__antithesis_instrumentation__.Notify(187806)
		if ok {
			__antithesis_instrumentation__.Notify(187809)

			scanner := makeResultScanner(it.Types())
			for ; ok; ok, err = it.Next(ctx) {
				__antithesis_instrumentation__.Notify(187811)
				row := it.Cur()

				var grant serverpb.TableDetailsResponse_Grant
				var privileges string
				if err := scanner.Scan(row, userCol, &grant.User); err != nil {
					__antithesis_instrumentation__.Notify(187814)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187815)
				}
				__antithesis_instrumentation__.Notify(187812)
				if err := scanner.Scan(row, privilegesCol, &privileges); err != nil {
					__antithesis_instrumentation__.Notify(187816)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(187817)
				}
				__antithesis_instrumentation__.Notify(187813)
				grant.Privileges = strings.Split(privileges, ",")
				resp.Grants = append(resp.Grants, grant)
			}
			__antithesis_instrumentation__.Notify(187810)
			if err != nil {
				__antithesis_instrumentation__.Notify(187818)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(187819)
			}
		} else {
			__antithesis_instrumentation__.Notify(187820)
		}
	}
	__antithesis_instrumentation__.Notify(187719)

	row, cols, err := s.server.sqlServer.internalExecutor.QueryRowExWithCols(
		ctx, "admin-show-create", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		fmt.Sprintf("SHOW CREATE %s", escQualTable),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(187821)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187822)
	}
	{
		__antithesis_instrumentation__.Notify(187823)
		const createCol = "create_statement"
		if row == nil {
			__antithesis_instrumentation__.Notify(187826)
			return nil, errors.New("create response not available")
		} else {
			__antithesis_instrumentation__.Notify(187827)
		}
		__antithesis_instrumentation__.Notify(187824)

		scanner := makeResultScanner(cols)
		var createStmt string
		if err := scanner.Scan(row, createCol, &createStmt); err != nil {
			__antithesis_instrumentation__.Notify(187828)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187829)
		}
		__antithesis_instrumentation__.Notify(187825)

		resp.CreateTableStatement = createStmt
	}
	__antithesis_instrumentation__.Notify(187720)

	row, cols, err = s.server.sqlServer.internalExecutor.QueryRowExWithCols(
		ctx, "admin-show-statistics", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		fmt.Sprintf("SELECT max(created) AS created FROM [SHOW STATISTICS FOR TABLE %s]", escQualTable),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(187830)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187831)
	}
	__antithesis_instrumentation__.Notify(187721)
	if row != nil {
		__antithesis_instrumentation__.Notify(187832)
		scanner := makeResultScanner(cols)
		const createdCol = "created"
		var createdTs *time.Time
		if err := scanner.Scan(row, createdCol, &createdTs); err != nil {
			__antithesis_instrumentation__.Notify(187834)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187835)
		}
		__antithesis_instrumentation__.Notify(187833)
		resp.StatsLastCreatedAt = createdTs
	} else {
		__antithesis_instrumentation__.Notify(187836)
	}
	__antithesis_instrumentation__.Notify(187722)

	row, cols, err = s.server.sqlServer.internalExecutor.QueryRowExWithCols(
		ctx, "admin-show-zone-config", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		fmt.Sprintf("SHOW ZONE CONFIGURATION FOR TABLE %s", escQualTable))
	if err != nil {
		__antithesis_instrumentation__.Notify(187837)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187838)
	}
	{
		__antithesis_instrumentation__.Notify(187839)
		const rawConfigSQLColName = "raw_config_sql"
		if row == nil {
			__antithesis_instrumentation__.Notify(187842)
			return nil, errors.New("show zone config response not available")
		} else {
			__antithesis_instrumentation__.Notify(187843)
		}
		__antithesis_instrumentation__.Notify(187840)

		scanner := makeResultScanner(cols)
		var configureZoneStmt string
		if err := scanner.Scan(row, rawConfigSQLColName, &configureZoneStmt); err != nil {
			__antithesis_instrumentation__.Notify(187844)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187845)
		}
		__antithesis_instrumentation__.Notify(187841)
		resp.ConfigureZoneStatement = configureZoneStmt
	}
	__antithesis_instrumentation__.Notify(187723)

	var tableID descpb.ID

	{
		__antithesis_instrumentation__.Notify(187846)
		databaseID, err := s.queryDatabaseID(ctx, userName, req.Database)
		if err != nil {
			__antithesis_instrumentation__.Notify(187851)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187852)
		}
		__antithesis_instrumentation__.Notify(187847)
		tableID, err = s.queryTableID(ctx, userName, req.Database, escQualTable)
		if err != nil {
			__antithesis_instrumentation__.Notify(187853)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187854)
		}
		__antithesis_instrumentation__.Notify(187848)
		resp.DescriptorID = int64(tableID)

		id, zone, zoneExists, err := s.queryZonePath(ctx, userName, []descpb.ID{databaseID, tableID})
		if err != nil {
			__antithesis_instrumentation__.Notify(187855)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187856)
		}
		__antithesis_instrumentation__.Notify(187849)

		if !zoneExists {
			__antithesis_instrumentation__.Notify(187857)
			zone = s.server.cfg.DefaultZoneConfig
		} else {
			__antithesis_instrumentation__.Notify(187858)
		}
		__antithesis_instrumentation__.Notify(187850)
		resp.ZoneConfig = zone

		switch id {
		case databaseID:
			__antithesis_instrumentation__.Notify(187859)
			resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_DATABASE
		case tableID:
			__antithesis_instrumentation__.Notify(187860)
			resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_TABLE
		default:
			__antithesis_instrumentation__.Notify(187861)
			resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_CLUSTER
		}
	}

	{
		__antithesis_instrumentation__.Notify(187862)
		tableSpan := generateTableSpan(tableID)
		tableRSpan, err := keys.SpanAddr(tableSpan)
		if err != nil {
			__antithesis_instrumentation__.Notify(187865)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187866)
		}
		__antithesis_instrumentation__.Notify(187863)
		rangeCount, err := s.server.distSender.CountRanges(ctx, tableRSpan)
		if err != nil {
			__antithesis_instrumentation__.Notify(187867)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187868)
		}
		__antithesis_instrumentation__.Notify(187864)
		resp.RangeCount = rangeCount
	}
	__antithesis_instrumentation__.Notify(187724)

	return &resp, nil
}

func generateTableSpan(tableID descpb.ID) roachpb.Span {
	__antithesis_instrumentation__.Notify(187869)
	tableStartKey := keys.TODOSQLCodec.TablePrefix(uint32(tableID))
	tableEndKey := tableStartKey.PrefixEnd()
	return roachpb.Span{Key: tableStartKey, EndKey: tableEndKey}
}

func (s *adminServer) TableStats(
	ctx context.Context, req *serverpb.TableStatsRequest,
) (*serverpb.TableStatsResponse, error) {
	__antithesis_instrumentation__.Notify(187870)
	ctx = s.server.AnnotateCtx(ctx)
	userName, err := s.requireAdminUser(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(187875)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187876)
	}
	__antithesis_instrumentation__.Notify(187871)
	escQualTable, err := getFullyQualifiedTableName(req.Database, req.Table)
	if err != nil {
		__antithesis_instrumentation__.Notify(187877)
		return nil, maybeHandleNotFoundError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187878)
	}
	__antithesis_instrumentation__.Notify(187872)

	tableID, err := s.queryTableID(ctx, userName, req.Database, escQualTable)
	if err != nil {
		__antithesis_instrumentation__.Notify(187879)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187880)
	}
	__antithesis_instrumentation__.Notify(187873)
	tableSpan := generateTableSpan(tableID)

	r, err := s.statsForSpan(ctx, tableSpan)
	if err != nil {
		__antithesis_instrumentation__.Notify(187881)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187882)
	}
	__antithesis_instrumentation__.Notify(187874)
	return r, nil
}

func (s *adminServer) NonTableStats(
	ctx context.Context, req *serverpb.NonTableStatsRequest,
) (*serverpb.NonTableStatsResponse, error) {
	__antithesis_instrumentation__.Notify(187883)
	ctx = s.server.AnnotateCtx(ctx)
	if _, err := s.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(187887)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187888)
	}
	__antithesis_instrumentation__.Notify(187884)

	timeSeriesStats, err := s.statsForSpan(ctx, roachpb.Span{
		Key:    keys.TimeseriesPrefix,
		EndKey: keys.TimeseriesPrefix.PrefixEnd(),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(187889)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187890)
	}
	__antithesis_instrumentation__.Notify(187885)
	response := serverpb.NonTableStatsResponse{
		TimeSeriesStats: timeSeriesStats,
	}

	spansForInternalUse := []roachpb.Span{
		{
			Key:    keys.LocalMax,
			EndKey: keys.TimeseriesPrefix,
		},
		{
			Key:    keys.TimeseriesKeyMax,
			EndKey: keys.TableDataMin,
		},
	}
	for _, span := range spansForInternalUse {
		__antithesis_instrumentation__.Notify(187891)
		nonTableStats, err := s.statsForSpan(ctx, span)
		if err != nil {
			__antithesis_instrumentation__.Notify(187893)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(187894)
		}
		__antithesis_instrumentation__.Notify(187892)
		if response.InternalUseStats == nil {
			__antithesis_instrumentation__.Notify(187895)
			response.InternalUseStats = nonTableStats
		} else {
			__antithesis_instrumentation__.Notify(187896)
			response.InternalUseStats.Add(nonTableStats)
		}
	}
	__antithesis_instrumentation__.Notify(187886)

	response.InternalUseStats.RangeCount += nonTableDescriptorRangeCount()

	return &response, nil
}

func (s *adminServer) statsForSpan(
	ctx context.Context, span roachpb.Span,
) (*serverpb.TableStatsResponse, error) {
	__antithesis_instrumentation__.Notify(187897)
	startKey, err := keys.Addr(span.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(187907)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187908)
	}
	__antithesis_instrumentation__.Notify(187898)
	endKey, err := keys.Addr(span.EndKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(187909)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187910)
	}
	__antithesis_instrumentation__.Notify(187899)

	startMetaKey := keys.RangeMetaKey(startKey)
	if bytes.Equal(startMetaKey, roachpb.RKeyMin) {
		__antithesis_instrumentation__.Notify(187911)

		startMetaKey = keys.RangeMetaKey(keys.MustAddr(keys.Meta2Prefix))
	} else {
		__antithesis_instrumentation__.Notify(187912)
	}
	__antithesis_instrumentation__.Notify(187900)

	rangeDescKVs, err := s.server.db.Scan(ctx, startMetaKey, keys.RangeMetaKey(endKey), 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(187913)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187914)
	}
	__antithesis_instrumentation__.Notify(187901)

	nodeIDs := make(map[roachpb.NodeID]struct{})
	for _, kv := range rangeDescKVs {
		__antithesis_instrumentation__.Notify(187915)
		var rng roachpb.RangeDescriptor
		if err := kv.Value.GetProto(&rng); err != nil {
			__antithesis_instrumentation__.Notify(187917)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187918)
		}
		__antithesis_instrumentation__.Notify(187916)
		for _, repl := range rng.Replicas().Descriptors() {
			__antithesis_instrumentation__.Notify(187919)
			nodeIDs[repl.NodeID] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(187902)

	nodeIDList := make([]roachpb.NodeID, 0, len(nodeIDs))
	for id := range nodeIDs {
		__antithesis_instrumentation__.Notify(187920)
		nodeIDList = append(nodeIDList, id)
	}
	__antithesis_instrumentation__.Notify(187903)
	sort.Slice(nodeIDList, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(187921)
		return nodeIDList[i] < nodeIDList[j]
	})
	__antithesis_instrumentation__.Notify(187904)

	tableStatResponse := serverpb.TableStatsResponse{
		NodeCount: int64(len(nodeIDs)),
		NodeIDs:   nodeIDList,

		RangeCount: int64(len(rangeDescKVs)),
	}
	type nodeResponse struct {
		nodeID roachpb.NodeID
		resp   *serverpb.SpanStatsResponse
		err    error
	}

	responses := make(chan nodeResponse, len(nodeIDs))
	for nodeID := range nodeIDs {
		__antithesis_instrumentation__.Notify(187922)
		nodeID := nodeID
		if err := s.server.stopper.RunAsyncTask(
			ctx, "server.adminServer: requesting remote stats",
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(187923)

				var spanResponse *serverpb.SpanStatsResponse
				err := contextutil.RunWithTimeout(ctx, "request remote stats", 5*base.NetworkTimeout,
					func(ctx context.Context) error {
						__antithesis_instrumentation__.Notify(187925)
						client, err := s.server.status.dialNode(ctx, nodeID)
						if err == nil {
							__antithesis_instrumentation__.Notify(187927)
							req := serverpb.SpanStatsRequest{
								StartKey: startKey,
								EndKey:   endKey,
								NodeID:   nodeID.String(),
							}
							spanResponse, err = client.SpanStats(ctx, &req)
						} else {
							__antithesis_instrumentation__.Notify(187928)
						}
						__antithesis_instrumentation__.Notify(187926)
						return err
					})
				__antithesis_instrumentation__.Notify(187924)

				responses <- nodeResponse{
					nodeID: nodeID,
					resp:   spanResponse,
					err:    err,
				}
			}); err != nil {
			__antithesis_instrumentation__.Notify(187929)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(187930)
		}
	}
	__antithesis_instrumentation__.Notify(187905)
	for remainingResponses := len(nodeIDs); remainingResponses > 0; remainingResponses-- {
		__antithesis_instrumentation__.Notify(187931)
		select {
		case resp := <-responses:
			__antithesis_instrumentation__.Notify(187932)

			if resp.err != nil {
				__antithesis_instrumentation__.Notify(187934)
				tableStatResponse.MissingNodes = append(
					tableStatResponse.MissingNodes,
					serverpb.TableStatsResponse_MissingNode{
						NodeID:       resp.nodeID.String(),
						ErrorMessage: resp.err.Error(),
					},
				)
			} else {
				__antithesis_instrumentation__.Notify(187935)
				tableStatResponse.Stats.Add(resp.resp.TotalStats)
				tableStatResponse.ReplicaCount += int64(resp.resp.RangeCount)
				tableStatResponse.ApproximateDiskBytes += resp.resp.ApproximateDiskBytes
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(187933)

			return nil, ctx.Err()
		}
	}
	__antithesis_instrumentation__.Notify(187906)

	return &tableStatResponse, nil
}

func (s *adminServer) Users(
	ctx context.Context, req *serverpb.UsersRequest,
) (_ *serverpb.UsersResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187936)
	ctx = s.server.AnnotateCtx(ctx)
	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(187939)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187940)
	}
	__antithesis_instrumentation__.Notify(187937)
	r, err := s.usersHelper(ctx, req, userName)
	if err != nil {
		__antithesis_instrumentation__.Notify(187941)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187942)
	}
	__antithesis_instrumentation__.Notify(187938)
	return r, nil
}

func (s *adminServer) usersHelper(
	ctx context.Context, req *serverpb.UsersRequest, userName security.SQLUsername,
) (_ *serverpb.UsersResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187943)
	query := `SELECT username FROM system.users WHERE "isRole" = false`
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-users", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		query,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(187948)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187949)
	}
	__antithesis_instrumentation__.Notify(187944)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(187950)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	__antithesis_instrumentation__.Notify(187945)

	var resp serverpb.UsersResponse
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(187951)
		row := it.Cur()
		resp.Users = append(resp.Users, serverpb.UsersResponse_User{Username: string(tree.MustBeDString(row[0]))})
	}
	__antithesis_instrumentation__.Notify(187946)
	if err != nil {
		__antithesis_instrumentation__.Notify(187952)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187953)
	}
	__antithesis_instrumentation__.Notify(187947)
	return &resp, nil
}

var eventSetClusterSettingName = eventpb.GetEventTypeName(&eventpb.SetClusterSetting{})

func combineAllErrors(errs []error) error {
	__antithesis_instrumentation__.Notify(187954)
	var combinedErrors error
	for _, err := range errs {
		__antithesis_instrumentation__.Notify(187956)
		combinedErrors = errors.CombineErrors(combinedErrors, err)
	}
	__antithesis_instrumentation__.Notify(187955)
	return combinedErrors
}

func (s *adminServer) Events(
	ctx context.Context, req *serverpb.EventsRequest,
) (_ *serverpb.EventsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187957)
	ctx = s.server.AnnotateCtx(ctx)

	userName, err := s.requireAdminUser(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(187961)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187962)
	}
	__antithesis_instrumentation__.Notify(187958)
	redactEvents := !req.UnredactedEvents

	limit := req.Limit
	if limit == 0 {
		__antithesis_instrumentation__.Notify(187963)
		limit = defaultAPIEventLimit
	} else {
		__antithesis_instrumentation__.Notify(187964)
	}
	__antithesis_instrumentation__.Notify(187959)

	r, err := s.eventsHelper(ctx, req, userName, int(limit), 0, redactEvents)
	if err != nil {
		__antithesis_instrumentation__.Notify(187965)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(187966)
	}
	__antithesis_instrumentation__.Notify(187960)
	return r, nil
}

func (s *adminServer) eventsHelper(
	ctx context.Context,
	req *serverpb.EventsRequest,
	userName security.SQLUsername,
	limit, offset int,
	redactEvents bool,
) (_ *serverpb.EventsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(187967)

	q := makeSQLQuery()
	q.Append(`SELECT timestamp, "eventType", "targetID", "reportingID", info, "uniqueID" `)
	q.Append("FROM system.eventlog ")
	q.Append("WHERE true ")
	if len(req.Type) > 0 {
		__antithesis_instrumentation__.Notify(187978)
		q.Append(`AND "eventType" = $ `, req.Type)
	} else {
		__antithesis_instrumentation__.Notify(187979)
	}
	__antithesis_instrumentation__.Notify(187968)
	if req.TargetId > 0 {
		__antithesis_instrumentation__.Notify(187980)
		q.Append(`AND "targetID" = $ `, req.TargetId)
	} else {
		__antithesis_instrumentation__.Notify(187981)
	}
	__antithesis_instrumentation__.Notify(187969)
	q.Append("ORDER BY timestamp DESC ")
	if limit > 0 {
		__antithesis_instrumentation__.Notify(187982)
		q.Append("LIMIT $", limit)
		if offset > 0 {
			__antithesis_instrumentation__.Notify(187983)
			q.Append(" OFFSET $", offset)
		} else {
			__antithesis_instrumentation__.Notify(187984)
		}
	} else {
		__antithesis_instrumentation__.Notify(187985)
	}
	__antithesis_instrumentation__.Notify(187970)
	if len(q.Errors()) > 0 {
		__antithesis_instrumentation__.Notify(187986)
		return nil, combineAllErrors(q.Errors())
	} else {
		__antithesis_instrumentation__.Notify(187987)
	}
	__antithesis_instrumentation__.Notify(187971)
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-events", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		q.String(), q.QueryArguments()...)
	if err != nil {
		__antithesis_instrumentation__.Notify(187988)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187989)
	}
	__antithesis_instrumentation__.Notify(187972)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(187990)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	__antithesis_instrumentation__.Notify(187973)

	var resp serverpb.EventsResponse
	ok, err := it.Next(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(187991)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187992)
	}
	__antithesis_instrumentation__.Notify(187974)
	if !ok {
		__antithesis_instrumentation__.Notify(187993)

		return &resp, nil
	} else {
		__antithesis_instrumentation__.Notify(187994)
	}
	__antithesis_instrumentation__.Notify(187975)
	scanner := makeResultScanner(it.Types())
	for ; ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(187995)
		row := it.Cur()
		var event serverpb.EventsResponse_Event
		var ts time.Time
		if err := scanner.ScanIndex(row, 0, &ts); err != nil {
			__antithesis_instrumentation__.Notify(188004)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188005)
		}
		__antithesis_instrumentation__.Notify(187996)
		event.Timestamp = ts
		if err := scanner.ScanIndex(row, 1, &event.EventType); err != nil {
			__antithesis_instrumentation__.Notify(188006)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188007)
		}
		__antithesis_instrumentation__.Notify(187997)
		if err := scanner.ScanIndex(row, 2, &event.TargetID); err != nil {
			__antithesis_instrumentation__.Notify(188008)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188009)
		}
		__antithesis_instrumentation__.Notify(187998)
		if err := scanner.ScanIndex(row, 3, &event.ReportingID); err != nil {
			__antithesis_instrumentation__.Notify(188010)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188011)
		}
		__antithesis_instrumentation__.Notify(187999)
		if err := scanner.ScanIndex(row, 4, &event.Info); err != nil {
			__antithesis_instrumentation__.Notify(188012)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188013)
		}
		__antithesis_instrumentation__.Notify(188000)
		if event.EventType == eventSetClusterSettingName {
			__antithesis_instrumentation__.Notify(188014)
			if redactEvents {
				__antithesis_instrumentation__.Notify(188015)
				event.Info = redactSettingsChange(event.Info)
			} else {
				__antithesis_instrumentation__.Notify(188016)
			}
		} else {
			__antithesis_instrumentation__.Notify(188017)
		}
		__antithesis_instrumentation__.Notify(188001)
		if err := scanner.ScanIndex(row, 5, &event.UniqueID); err != nil {
			__antithesis_instrumentation__.Notify(188018)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188019)
		}
		__antithesis_instrumentation__.Notify(188002)
		if redactEvents {
			__antithesis_instrumentation__.Notify(188020)
			event.Info = redactStatement(event.Info)
		} else {
			__antithesis_instrumentation__.Notify(188021)
		}
		__antithesis_instrumentation__.Notify(188003)

		resp.Events = append(resp.Events, event)
	}
	__antithesis_instrumentation__.Notify(187976)
	if err != nil {
		__antithesis_instrumentation__.Notify(188022)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188023)
	}
	__antithesis_instrumentation__.Notify(187977)
	return &resp, nil
}

func redactSettingsChange(info string) string {
	__antithesis_instrumentation__.Notify(188024)
	var s eventpb.SetClusterSetting
	if err := json.Unmarshal([]byte(info), &s); err != nil {
		__antithesis_instrumentation__.Notify(188027)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(188028)
	}
	__antithesis_instrumentation__.Notify(188025)
	s.Value = "<hidden>"
	ret, err := json.Marshal(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(188029)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(188030)
	}
	__antithesis_instrumentation__.Notify(188026)
	return string(ret)
}

func redactStatement(info string) string {
	__antithesis_instrumentation__.Notify(188031)
	s := map[string]interface{}{}
	if err := json.Unmarshal([]byte(info), &s); err != nil {
		__antithesis_instrumentation__.Notify(188035)
		return info
	} else {
		__antithesis_instrumentation__.Notify(188036)
	}
	__antithesis_instrumentation__.Notify(188032)
	if _, ok := s["Statement"]; ok {
		__antithesis_instrumentation__.Notify(188037)
		s["Statement"] = "<hidden>"
	} else {
		__antithesis_instrumentation__.Notify(188038)
	}
	__antithesis_instrumentation__.Notify(188033)
	ret, err := json.Marshal(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(188039)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(188040)
	}
	__antithesis_instrumentation__.Notify(188034)
	return string(ret)
}

func (s *adminServer) RangeLog(
	ctx context.Context, req *serverpb.RangeLogRequest,
) (_ *serverpb.RangeLogResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188041)
	ctx = s.server.AnnotateCtx(ctx)

	userName, err := s.requireAdminUser(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188044)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188045)
	}
	__antithesis_instrumentation__.Notify(188042)

	r, err := s.rangeLogHelper(ctx, req, userName)
	if err != nil {
		__antithesis_instrumentation__.Notify(188046)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188047)
	}
	__antithesis_instrumentation__.Notify(188043)
	return r, nil
}

func (s *adminServer) rangeLogHelper(
	ctx context.Context, req *serverpb.RangeLogRequest, userName security.SQLUsername,
) (_ *serverpb.RangeLogResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188048)
	limit := req.Limit
	if limit == 0 {
		__antithesis_instrumentation__.Notify(188060)
		limit = defaultAPIEventLimit
	} else {
		__antithesis_instrumentation__.Notify(188061)
	}
	__antithesis_instrumentation__.Notify(188049)

	q := makeSQLQuery()
	q.Append(`SELECT timestamp, "rangeID", "storeID", "eventType", "otherRangeID", info `)
	q.Append("FROM system.rangelog ")
	if req.RangeId > 0 {
		__antithesis_instrumentation__.Notify(188062)
		rangeID := tree.NewDInt(tree.DInt(req.RangeId))
		q.Append(`WHERE "rangeID" = $ OR "otherRangeID" = $`, rangeID, rangeID)
	} else {
		__antithesis_instrumentation__.Notify(188063)
	}
	__antithesis_instrumentation__.Notify(188050)
	if limit > 0 {
		__antithesis_instrumentation__.Notify(188064)
		q.Append("ORDER BY timestamp desc ")
		q.Append("LIMIT $", tree.NewDInt(tree.DInt(limit)))
	} else {
		__antithesis_instrumentation__.Notify(188065)
	}
	__antithesis_instrumentation__.Notify(188051)
	if len(q.Errors()) > 0 {
		__antithesis_instrumentation__.Notify(188066)
		return nil, combineAllErrors(q.Errors())
	} else {
		__antithesis_instrumentation__.Notify(188067)
	}
	__antithesis_instrumentation__.Notify(188052)
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-range-log", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		q.String(), q.QueryArguments()...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188068)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188069)
	}
	__antithesis_instrumentation__.Notify(188053)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(188070)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	__antithesis_instrumentation__.Notify(188054)

	var resp serverpb.RangeLogResponse
	ok, err := it.Next(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188071)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188072)
	}
	__antithesis_instrumentation__.Notify(188055)
	if !ok {
		__antithesis_instrumentation__.Notify(188073)

		return &resp, nil
	} else {
		__antithesis_instrumentation__.Notify(188074)
	}
	__antithesis_instrumentation__.Notify(188056)
	cols := it.Types()
	if len(cols) != 6 {
		__antithesis_instrumentation__.Notify(188075)
		return nil, errors.Errorf("incorrect number of columns in response, expected 6, got %d", len(cols))
	} else {
		__antithesis_instrumentation__.Notify(188076)
	}
	__antithesis_instrumentation__.Notify(188057)
	scanner := makeResultScanner(cols)
	for ; ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(188077)
		row := it.Cur()
		var event kvserverpb.RangeLogEvent
		var ts time.Time
		if err := scanner.ScanIndex(row, 0, &ts); err != nil {
			__antithesis_instrumentation__.Notify(188085)
			return nil, errors.Wrapf(err, "timestamp didn't parse correctly: %s", row[0].String())
		} else {
			__antithesis_instrumentation__.Notify(188086)
		}
		__antithesis_instrumentation__.Notify(188078)
		event.Timestamp = ts
		var rangeID int64
		if err := scanner.ScanIndex(row, 1, &rangeID); err != nil {
			__antithesis_instrumentation__.Notify(188087)
			return nil, errors.Wrapf(err, "RangeID didn't parse correctly: %s", row[1].String())
		} else {
			__antithesis_instrumentation__.Notify(188088)
		}
		__antithesis_instrumentation__.Notify(188079)
		event.RangeID = roachpb.RangeID(rangeID)
		var storeID int64
		if err := scanner.ScanIndex(row, 2, &storeID); err != nil {
			__antithesis_instrumentation__.Notify(188089)
			return nil, errors.Wrapf(err, "StoreID didn't parse correctly: %s", row[2].String())
		} else {
			__antithesis_instrumentation__.Notify(188090)
		}
		__antithesis_instrumentation__.Notify(188080)
		event.StoreID = roachpb.StoreID(int32(storeID))
		var eventTypeString string
		if err := scanner.ScanIndex(row, 3, &eventTypeString); err != nil {
			__antithesis_instrumentation__.Notify(188091)
			return nil, errors.Wrapf(err, "EventType didn't parse correctly: %s", row[3].String())
		} else {
			__antithesis_instrumentation__.Notify(188092)
		}
		__antithesis_instrumentation__.Notify(188081)
		if eventType, ok := kvserverpb.RangeLogEventType_value[eventTypeString]; ok {
			__antithesis_instrumentation__.Notify(188093)
			event.EventType = kvserverpb.RangeLogEventType(eventType)
		} else {
			__antithesis_instrumentation__.Notify(188094)
			return nil, errors.Errorf("EventType didn't parse correctly: %s", eventTypeString)
		}
		__antithesis_instrumentation__.Notify(188082)

		var otherRangeID int64
		if row[4].String() != "NULL" {
			__antithesis_instrumentation__.Notify(188095)
			if err := scanner.ScanIndex(row, 4, &otherRangeID); err != nil {
				__antithesis_instrumentation__.Notify(188097)
				return nil, errors.Wrapf(err, "OtherRangeID didn't parse correctly: %s", row[4].String())
			} else {
				__antithesis_instrumentation__.Notify(188098)
			}
			__antithesis_instrumentation__.Notify(188096)
			event.OtherRangeID = roachpb.RangeID(otherRangeID)
		} else {
			__antithesis_instrumentation__.Notify(188099)
		}
		__antithesis_instrumentation__.Notify(188083)

		var prettyInfo serverpb.RangeLogResponse_PrettyInfo
		if row[5].String() != "NULL" {
			__antithesis_instrumentation__.Notify(188100)
			var info string
			if err := scanner.ScanIndex(row, 5, &info); err != nil {
				__antithesis_instrumentation__.Notify(188107)
				return nil, errors.Wrapf(err, "info didn't parse correctly: %s", row[5].String())
			} else {
				__antithesis_instrumentation__.Notify(188108)
			}
			__antithesis_instrumentation__.Notify(188101)
			if err := json.Unmarshal([]byte(info), &event.Info); err != nil {
				__antithesis_instrumentation__.Notify(188109)
				return nil, errors.Wrapf(err, "info didn't parse correctly: %s", info)
			} else {
				__antithesis_instrumentation__.Notify(188110)
			}
			__antithesis_instrumentation__.Notify(188102)
			if event.Info.NewDesc != nil {
				__antithesis_instrumentation__.Notify(188111)
				prettyInfo.NewDesc = event.Info.NewDesc.String()
			} else {
				__antithesis_instrumentation__.Notify(188112)
			}
			__antithesis_instrumentation__.Notify(188103)
			if event.Info.UpdatedDesc != nil {
				__antithesis_instrumentation__.Notify(188113)
				prettyInfo.UpdatedDesc = event.Info.UpdatedDesc.String()
			} else {
				__antithesis_instrumentation__.Notify(188114)
			}
			__antithesis_instrumentation__.Notify(188104)
			if event.Info.AddedReplica != nil {
				__antithesis_instrumentation__.Notify(188115)
				prettyInfo.AddedReplica = event.Info.AddedReplica.String()
			} else {
				__antithesis_instrumentation__.Notify(188116)
			}
			__antithesis_instrumentation__.Notify(188105)
			if event.Info.RemovedReplica != nil {
				__antithesis_instrumentation__.Notify(188117)
				prettyInfo.RemovedReplica = event.Info.RemovedReplica.String()
			} else {
				__antithesis_instrumentation__.Notify(188118)
			}
			__antithesis_instrumentation__.Notify(188106)
			prettyInfo.Reason = string(event.Info.Reason)
			prettyInfo.Details = event.Info.Details
		} else {
			__antithesis_instrumentation__.Notify(188119)
		}
		__antithesis_instrumentation__.Notify(188084)

		resp.Events = append(resp.Events, serverpb.RangeLogResponse_Event{
			Event:      event,
			PrettyInfo: prettyInfo,
		})
	}
	__antithesis_instrumentation__.Notify(188058)
	if err != nil {
		__antithesis_instrumentation__.Notify(188120)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188121)
	}
	__antithesis_instrumentation__.Notify(188059)
	return &resp, nil
}

func (s *adminServer) getUIData(
	ctx context.Context, userName security.SQLUsername, keys []string,
) (_ *serverpb.GetUIDataResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188122)
	if len(keys) == 0 {
		__antithesis_instrumentation__.Notify(188130)
		return &serverpb.GetUIDataResponse{}, nil
	} else {
		__antithesis_instrumentation__.Notify(188131)
	}
	__antithesis_instrumentation__.Notify(188123)

	query := makeSQLQuery()
	query.Append(`SELECT key, value, "lastUpdated" FROM system.ui WHERE key IN (`)
	for i, key := range keys {
		__antithesis_instrumentation__.Notify(188132)
		if i != 0 {
			__antithesis_instrumentation__.Notify(188134)
			query.Append(",")
		} else {
			__antithesis_instrumentation__.Notify(188135)
		}
		__antithesis_instrumentation__.Notify(188133)
		query.Append("$", tree.NewDString(makeUIKey(userName, key)))
	}
	__antithesis_instrumentation__.Notify(188124)
	query.Append(");")
	if errs := query.Errors(); len(errs) > 0 {
		__antithesis_instrumentation__.Notify(188136)
		var err error
		for _, e := range errs {
			__antithesis_instrumentation__.Notify(188138)
			err = errors.CombineErrors(err, e)
		}
		__antithesis_instrumentation__.Notify(188137)
		return nil, errors.Wrap(err, "error constructing query")
	} else {
		__antithesis_instrumentation__.Notify(188139)
	}
	__antithesis_instrumentation__.Notify(188125)
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-getUIData", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		query.String(), query.QueryArguments()...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188140)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188141)
	}
	__antithesis_instrumentation__.Notify(188126)

	defer func() {
		__antithesis_instrumentation__.Notify(188142)
		retErr = errors.CombineErrors(retErr, it.Close())
	}()
	__antithesis_instrumentation__.Notify(188127)

	resp := serverpb.GetUIDataResponse{KeyValues: make(map[string]serverpb.GetUIDataResponse_Value)}
	var hasNext bool
	for hasNext, err = it.Next(ctx); hasNext; hasNext, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(188143)
		row := it.Cur()
		dKey, ok := tree.AsDString(row[0])
		if !ok {
			__antithesis_instrumentation__.Notify(188147)
			return nil, errors.Errorf("unexpected type for UI key: %T", row[0])
		} else {
			__antithesis_instrumentation__.Notify(188148)
		}
		__antithesis_instrumentation__.Notify(188144)
		_, key := splitUIKey(string(dKey))
		dKey = tree.DString(key)

		dValue, ok := row[1].(*tree.DBytes)
		if !ok {
			__antithesis_instrumentation__.Notify(188149)
			return nil, errors.Errorf("unexpected type for UI value: %T", row[1])
		} else {
			__antithesis_instrumentation__.Notify(188150)
		}
		__antithesis_instrumentation__.Notify(188145)
		dLastUpdated, ok := row[2].(*tree.DTimestamp)
		if !ok {
			__antithesis_instrumentation__.Notify(188151)
			return nil, errors.Errorf("unexpected type for UI lastUpdated: %T", row[2])
		} else {
			__antithesis_instrumentation__.Notify(188152)
		}
		__antithesis_instrumentation__.Notify(188146)

		resp.KeyValues[string(dKey)] = serverpb.GetUIDataResponse_Value{
			Value:       []byte(*dValue),
			LastUpdated: dLastUpdated.Time,
		}
	}
	__antithesis_instrumentation__.Notify(188128)
	if err != nil {
		__antithesis_instrumentation__.Notify(188153)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188154)
	}
	__antithesis_instrumentation__.Notify(188129)
	return &resp, nil
}

func makeUIKey(username security.SQLUsername, key string) string {
	__antithesis_instrumentation__.Notify(188155)
	return username.Normalized() + "$" + key
}

func splitUIKey(combined string) (string, string) {
	__antithesis_instrumentation__.Notify(188156)
	pair := strings.SplitN(combined, "$", 2)
	return pair[0], pair[1]
}

func (s *adminServer) SetUIData(
	ctx context.Context, req *serverpb.SetUIDataRequest,
) (*serverpb.SetUIDataResponse, error) {
	__antithesis_instrumentation__.Notify(188157)
	ctx = s.server.AnnotateCtx(ctx)

	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188161)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188162)
	}
	__antithesis_instrumentation__.Notify(188158)

	if len(req.KeyValues) == 0 {
		__antithesis_instrumentation__.Notify(188163)
		return nil, status.Errorf(codes.InvalidArgument, "KeyValues cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(188164)
	}
	__antithesis_instrumentation__.Notify(188159)

	for key, val := range req.KeyValues {
		__antithesis_instrumentation__.Notify(188165)

		query := `UPSERT INTO system.ui (key, value, "lastUpdated") VALUES ($1, $2, now())`
		rowsAffected, err := s.server.sqlServer.internalExecutor.ExecEx(
			ctx, "admin-set-ui-data", nil,
			sessiondata.InternalExecutorOverride{
				User: security.RootUserName(),
			},
			query, makeUIKey(userName, key), val)
		if err != nil {
			__antithesis_instrumentation__.Notify(188167)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(188168)
		}
		__antithesis_instrumentation__.Notify(188166)
		if rowsAffected != 1 {
			__antithesis_instrumentation__.Notify(188169)
			return nil, serverErrorf(ctx, "rows affected %d != expected %d", rowsAffected, 1)
		} else {
			__antithesis_instrumentation__.Notify(188170)
		}
	}
	__antithesis_instrumentation__.Notify(188160)

	return &serverpb.SetUIDataResponse{}, nil
}

func (s *adminServer) GetUIData(
	ctx context.Context, req *serverpb.GetUIDataRequest,
) (*serverpb.GetUIDataResponse, error) {
	__antithesis_instrumentation__.Notify(188171)
	ctx = s.server.AnnotateCtx(ctx)

	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188175)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188176)
	}
	__antithesis_instrumentation__.Notify(188172)

	if len(req.Keys) == 0 {
		__antithesis_instrumentation__.Notify(188177)
		return nil, status.Errorf(codes.InvalidArgument, "keys cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(188178)
	}
	__antithesis_instrumentation__.Notify(188173)

	resp, err := s.getUIData(ctx, userName, req.Keys)
	if err != nil {
		__antithesis_instrumentation__.Notify(188179)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188180)
	}
	__antithesis_instrumentation__.Notify(188174)

	return resp, nil
}

func (s *adminServer) Settings(
	ctx context.Context, req *serverpb.SettingsRequest,
) (*serverpb.SettingsResponse, error) {
	__antithesis_instrumentation__.Notify(188181)
	ctx = s.server.AnnotateCtx(ctx)

	keys := req.Keys
	if len(keys) == 0 {
		__antithesis_instrumentation__.Notify(188186)
		keys = settings.Keys(settings.ForSystemTenant)
	} else {
		__antithesis_instrumentation__.Notify(188187)
	}
	__antithesis_instrumentation__.Notify(188182)

	user, isAdmin, err := s.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188188)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188189)
	}
	__antithesis_instrumentation__.Notify(188183)

	var lookupPurpose settings.LookupPurpose
	if isAdmin {
		__antithesis_instrumentation__.Notify(188190)

		lookupPurpose = settings.LookupForReporting
		if req.UnredactedValues {
			__antithesis_instrumentation__.Notify(188191)
			lookupPurpose = settings.LookupForLocalAccess
		} else {
			__antithesis_instrumentation__.Notify(188192)
		}
	} else {
		__antithesis_instrumentation__.Notify(188193)

		lookupPurpose = settings.LookupForReporting

		hasView, err := s.hasRoleOption(ctx, user, roleoption.VIEWCLUSTERSETTING)
		if err != nil {
			__antithesis_instrumentation__.Notify(188196)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188197)
		}
		__antithesis_instrumentation__.Notify(188194)

		hasModify, err := s.hasRoleOption(ctx, user, roleoption.MODIFYCLUSTERSETTING)
		if err != nil {
			__antithesis_instrumentation__.Notify(188198)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188199)
		}
		__antithesis_instrumentation__.Notify(188195)

		if !hasModify && func() bool {
			__antithesis_instrumentation__.Notify(188200)
			return !hasView == true
		}() == true {
			__antithesis_instrumentation__.Notify(188201)
			return nil, status.Errorf(
				codes.PermissionDenied, "this operation requires either %s or %s role options",
				roleoption.VIEWCLUSTERSETTING, roleoption.MODIFYCLUSTERSETTING)
		} else {
			__antithesis_instrumentation__.Notify(188202)
		}
	}
	__antithesis_instrumentation__.Notify(188184)

	resp := serverpb.SettingsResponse{KeyValues: make(map[string]serverpb.SettingsResponse_Value)}
	for _, k := range keys {
		__antithesis_instrumentation__.Notify(188203)
		v, ok := settings.Lookup(k, lookupPurpose, settings.ForSystemTenant)
		if !ok {
			__antithesis_instrumentation__.Notify(188205)
			continue
		} else {
			__antithesis_instrumentation__.Notify(188206)
		}
		__antithesis_instrumentation__.Notify(188204)
		resp.KeyValues[k] = serverpb.SettingsResponse_Value{
			Type: v.Typ(),

			Value:       v.String(&s.server.st.SV),
			Description: v.Description(),
			Public:      v.Visibility() == settings.Public,
		}
	}
	__antithesis_instrumentation__.Notify(188185)

	return &resp, nil
}

func (s *adminServer) Cluster(
	_ context.Context, req *serverpb.ClusterRequest,
) (*serverpb.ClusterResponse, error) {
	__antithesis_instrumentation__.Notify(188207)
	storageClusterID := s.server.StorageClusterID()
	if storageClusterID == (uuid.UUID{}) {
		__antithesis_instrumentation__.Notify(188209)
		return nil, status.Errorf(codes.Unavailable, "cluster ID not yet available")
	} else {
		__antithesis_instrumentation__.Notify(188210)
	}
	__antithesis_instrumentation__.Notify(188208)

	organization := sql.ClusterOrganization.Get(&s.server.st.SV)
	enterpriseEnabled := base.CheckEnterpriseEnabled(
		s.server.st,
		s.server.rpcContext.LogicalClusterID.Get(),
		organization,
		"BACKUP") == nil

	return &serverpb.ClusterResponse{

		ClusterID:         storageClusterID.String(),
		ReportingEnabled:  logcrash.DiagnosticsReportingEnabled.Get(&s.server.st.SV),
		EnterpriseEnabled: enterpriseEnabled,
	}, nil
}

func (s *adminServer) Health(
	ctx context.Context, req *serverpb.HealthRequest,
) (*serverpb.HealthResponse, error) {
	__antithesis_instrumentation__.Notify(188211)
	telemetry.Inc(telemetryHealthCheck)

	resp := &serverpb.HealthResponse{}

	if !req.Ready {
		__antithesis_instrumentation__.Notify(188214)
		return resp, nil
	} else {
		__antithesis_instrumentation__.Notify(188215)
	}
	__antithesis_instrumentation__.Notify(188212)

	if err := s.checkReadinessForHealthCheck(ctx); err != nil {
		__antithesis_instrumentation__.Notify(188216)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188217)
	}
	__antithesis_instrumentation__.Notify(188213)
	return resp, nil
}

func (s *adminServer) checkReadinessForHealthCheck(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(188218)
	serveMode := s.server.grpc.mode.get()
	switch serveMode {
	case modeInitializing:
		__antithesis_instrumentation__.Notify(188224)
		return status.Error(codes.Unavailable, "node is waiting for cluster initialization")
	case modeDraining:
		__antithesis_instrumentation__.Notify(188225)

		return status.Errorf(codes.Unavailable, "node is shutting down")
	case modeOperational:
		__antithesis_instrumentation__.Notify(188226)
		break
	default:
		__antithesis_instrumentation__.Notify(188227)
		return serverError(ctx, errors.Newf("unknown mode: %v", serveMode))
	}
	__antithesis_instrumentation__.Notify(188219)

	l, ok := s.server.nodeLiveness.GetLiveness(s.server.NodeID())
	if !ok {
		__antithesis_instrumentation__.Notify(188228)
		return status.Error(codes.Unavailable, "liveness record not found")
	} else {
		__antithesis_instrumentation__.Notify(188229)
	}
	__antithesis_instrumentation__.Notify(188220)
	if !l.IsLive(s.server.clock.Now().GoTime()) {
		__antithesis_instrumentation__.Notify(188230)
		return status.Errorf(codes.Unavailable, "node is not healthy")
	} else {
		__antithesis_instrumentation__.Notify(188231)
	}
	__antithesis_instrumentation__.Notify(188221)
	if l.Draining {
		__antithesis_instrumentation__.Notify(188232)

		return status.Errorf(codes.Unavailable, "node is shutting down")
	} else {
		__antithesis_instrumentation__.Notify(188233)
	}
	__antithesis_instrumentation__.Notify(188222)

	if !s.server.sqlServer.isReady.Get() {
		__antithesis_instrumentation__.Notify(188234)
		return status.Errorf(codes.Unavailable, "node is not accepting SQL clients")
	} else {
		__antithesis_instrumentation__.Notify(188235)
	}
	__antithesis_instrumentation__.Notify(188223)

	return nil
}

func getLivenessStatusMap(
	nl *liveness.NodeLiveness, now time.Time, st *cluster.Settings,
) map[roachpb.NodeID]livenesspb.NodeLivenessStatus {
	__antithesis_instrumentation__.Notify(188236)
	livenesses := nl.GetLivenesses()
	threshold := kvserver.TimeUntilStoreDead.Get(&st.SV)

	statusMap := make(map[roachpb.NodeID]livenesspb.NodeLivenessStatus, len(livenesses))
	for _, liveness := range livenesses {
		__antithesis_instrumentation__.Notify(188238)
		status := kvserver.LivenessStatus(liveness, now, threshold)
		statusMap[liveness.NodeID] = status
	}
	__antithesis_instrumentation__.Notify(188237)
	return statusMap
}

func (s *adminServer) Liveness(
	context.Context, *serverpb.LivenessRequest,
) (*serverpb.LivenessResponse, error) {
	__antithesis_instrumentation__.Notify(188239)
	clock := s.server.clock
	statusMap := getLivenessStatusMap(
		s.server.nodeLiveness, clock.Now().GoTime(), s.server.st)
	livenesses := s.server.nodeLiveness.GetLivenesses()
	return &serverpb.LivenessResponse{
		Livenesses: livenesses,
		Statuses:   statusMap,
	}, nil
}

func (s *adminServer) Jobs(
	ctx context.Context, req *serverpb.JobsRequest,
) (_ *serverpb.JobsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188240)
	ctx = s.server.AnnotateCtx(ctx)

	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188243)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188244)
	}
	__antithesis_instrumentation__.Notify(188241)

	j, err := s.jobsHelper(ctx, req, userName)
	if err != nil {
		__antithesis_instrumentation__.Notify(188245)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188246)
	}
	__antithesis_instrumentation__.Notify(188242)
	return j, nil
}

func (s *adminServer) jobsHelper(
	ctx context.Context, req *serverpb.JobsRequest, userName security.SQLUsername,
) (_ *serverpb.JobsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188247)
	retryRunningCondition := "status='running' AND next_run > now() AND num_runs > 1"
	retryRevertingCondition := "status='reverting' AND next_run > now() AND num_runs > 1"

	q := makeSQLQuery()
	q.Append(`
      SELECT job_id, job_type, description, statement, user_name, descriptor_ids,
            case
              when ` + retryRunningCondition + ` then 'retry-running' 
              when ` + retryRevertingCondition + ` then 'retry-reverting' 
              else status
            end as status, running_status, created, started, finished, modified, fraction_completed,
            high_water_timestamp, error, last_run, next_run, num_runs, execution_events::string::bytes
        FROM crdb_internal.jobs
       WHERE true
	`)
	if req.Status == "retrying" {
		__antithesis_instrumentation__.Notify(188257)
		q.Append(" AND ( ( " + retryRunningCondition + " ) OR ( " + retryRevertingCondition + " ) )")
	} else {
		__antithesis_instrumentation__.Notify(188258)
		if req.Status != "" {
			__antithesis_instrumentation__.Notify(188259)
			q.Append(" AND status = $", req.Status)
		} else {
			__antithesis_instrumentation__.Notify(188260)
		}
	}
	__antithesis_instrumentation__.Notify(188248)
	if req.Type != jobspb.TypeUnspecified {
		__antithesis_instrumentation__.Notify(188261)
		q.Append(" AND job_type = $", req.Type.String())
	} else {
		__antithesis_instrumentation__.Notify(188262)

		q.Append(" AND (")
		for idx, jobType := range jobspb.AutomaticJobTypes {
			__antithesis_instrumentation__.Notify(188264)
			q.Append("job_type != $", jobType.String())
			if idx < len(jobspb.AutomaticJobTypes)-1 {
				__antithesis_instrumentation__.Notify(188265)
				q.Append(" AND ")
			} else {
				__antithesis_instrumentation__.Notify(188266)
			}
		}
		__antithesis_instrumentation__.Notify(188263)
		q.Append(" OR job_type IS NULL)")
	}
	__antithesis_instrumentation__.Notify(188249)
	q.Append("ORDER BY created DESC")
	if req.Limit > 0 {
		__antithesis_instrumentation__.Notify(188267)
		q.Append(" LIMIT $", tree.DInt(req.Limit))
	} else {
		__antithesis_instrumentation__.Notify(188268)
	}
	__antithesis_instrumentation__.Notify(188250)
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-jobs", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		q.String(), q.QueryArguments()...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188269)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188270)
	}
	__antithesis_instrumentation__.Notify(188251)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(188271)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	__antithesis_instrumentation__.Notify(188252)

	ok, err := it.Next(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188272)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188273)
	}
	__antithesis_instrumentation__.Notify(188253)

	var resp serverpb.JobsResponse
	if !ok {
		__antithesis_instrumentation__.Notify(188274)

		return &resp, nil
	} else {
		__antithesis_instrumentation__.Notify(188275)
	}
	__antithesis_instrumentation__.Notify(188254)
	scanner := makeResultScanner(it.Types())
	for ; ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(188276)
		row := it.Cur()
		var job serverpb.JobResponse
		err := scanRowIntoJob(scanner, row, &job)
		if err != nil {
			__antithesis_instrumentation__.Notify(188278)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188279)
		}
		__antithesis_instrumentation__.Notify(188277)

		resp.Jobs = append(resp.Jobs, job)
	}
	__antithesis_instrumentation__.Notify(188255)

	if err != nil {
		__antithesis_instrumentation__.Notify(188280)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188281)
	}
	__antithesis_instrumentation__.Notify(188256)
	return &resp, nil
}

func scanRowIntoJob(scanner resultScanner, row tree.Datums, job *serverpb.JobResponse) error {
	__antithesis_instrumentation__.Notify(188282)
	var fractionCompletedOrNil *float32
	var highwaterOrNil *apd.Decimal
	var runningStatusOrNil *string
	var executionFailures []byte
	if err := scanner.ScanAll(
		row,
		&job.ID,
		&job.Type,
		&job.Description,
		&job.Statement,
		&job.Username,
		&job.DescriptorIDs,
		&job.Status,
		&runningStatusOrNil,
		&job.Created,
		&job.Started,
		&job.Finished,
		&job.Modified,
		&fractionCompletedOrNil,
		&highwaterOrNil,
		&job.Error,
		&job.LastRun,
		&job.NextRun,
		&job.NumRuns,
		&executionFailures,
	); err != nil {
		__antithesis_instrumentation__.Notify(188287)
		return errors.Wrap(err, "scan")
	} else {
		__antithesis_instrumentation__.Notify(188288)
	}
	__antithesis_instrumentation__.Notify(188283)
	if highwaterOrNil != nil {
		__antithesis_instrumentation__.Notify(188289)
		highwaterTimestamp, err := tree.DecimalToHLC(highwaterOrNil)
		if err != nil {
			__antithesis_instrumentation__.Notify(188291)
			return errors.Wrap(err, "highwater timestamp had unexpected format")
		} else {
			__antithesis_instrumentation__.Notify(188292)
		}
		__antithesis_instrumentation__.Notify(188290)
		goTime := highwaterTimestamp.GoTime()
		job.HighwaterTimestamp = &goTime
		job.HighwaterDecimal = highwaterOrNil.String()
	} else {
		__antithesis_instrumentation__.Notify(188293)
	}
	__antithesis_instrumentation__.Notify(188284)
	if fractionCompletedOrNil != nil {
		__antithesis_instrumentation__.Notify(188294)
		job.FractionCompleted = *fractionCompletedOrNil
	} else {
		__antithesis_instrumentation__.Notify(188295)
	}
	__antithesis_instrumentation__.Notify(188285)
	if runningStatusOrNil != nil {
		__antithesis_instrumentation__.Notify(188296)
		job.RunningStatus = *runningStatusOrNil
	} else {
		__antithesis_instrumentation__.Notify(188297)
	}
	{
		__antithesis_instrumentation__.Notify(188298)
		failures, err := jobs.ParseRetriableExecutionErrorLogFromJSON(executionFailures)
		if err != nil {
			__antithesis_instrumentation__.Notify(188300)
			return errors.Wrap(err, "parse")
		} else {
			__antithesis_instrumentation__.Notify(188301)
		}
		__antithesis_instrumentation__.Notify(188299)
		job.ExecutionFailures = make([]*serverpb.JobResponse_ExecutionFailure, len(failures))
		for i, f := range failures {
			__antithesis_instrumentation__.Notify(188302)
			start := time.UnixMicro(f.ExecutionStartMicros)
			end := time.UnixMicro(f.ExecutionEndMicros)
			job.ExecutionFailures[i] = &serverpb.JobResponse_ExecutionFailure{
				Status: f.Status,
				Start:  &start,
				End:    &end,
				Error:  f.TruncatedError,
			}
		}
	}
	__antithesis_instrumentation__.Notify(188286)
	return nil
}

func (s *adminServer) Job(
	ctx context.Context, request *serverpb.JobRequest,
) (_ *serverpb.JobResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188303)
	ctx = s.server.AnnotateCtx(ctx)

	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188306)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188307)
	}
	__antithesis_instrumentation__.Notify(188304)
	r, err := s.jobHelper(ctx, request, userName)
	if err != nil {
		__antithesis_instrumentation__.Notify(188308)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188309)
	}
	__antithesis_instrumentation__.Notify(188305)
	return r, nil
}

func (s *adminServer) jobHelper(
	ctx context.Context, request *serverpb.JobRequest, userName security.SQLUsername,
) (_ *serverpb.JobResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188310)
	const query = `
	        SELECT job_id, job_type, description, statement, user_name, descriptor_ids, status,
	  						 running_status, created, started, finished, modified,
	  						 fraction_completed, high_water_timestamp, error, last_run,
	  						 next_run, num_runs, execution_events::string::bytes
	          FROM crdb_internal.jobs
	         WHERE job_id = $1`
	row, cols, err := s.server.sqlServer.internalExecutor.QueryRowExWithCols(
		ctx, "admin-job", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		query,
		request.JobId,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(188314)
		return nil, errors.Wrapf(err, "expected to find 1 job with job_id=%d", request.JobId)
	} else {
		__antithesis_instrumentation__.Notify(188315)
	}
	__antithesis_instrumentation__.Notify(188311)

	if row == nil {
		__antithesis_instrumentation__.Notify(188316)
		return nil, errors.Errorf(
			"could not get job for job_id %d; 0 rows returned", request.JobId,
		)
	} else {
		__antithesis_instrumentation__.Notify(188317)
	}
	__antithesis_instrumentation__.Notify(188312)

	scanner := makeResultScanner(cols)

	var job serverpb.JobResponse
	err = scanRowIntoJob(scanner, row, &job)
	if err != nil {
		__antithesis_instrumentation__.Notify(188318)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188319)
	}
	__antithesis_instrumentation__.Notify(188313)

	return &job, nil
}

func (s *adminServer) Locations(
	ctx context.Context, req *serverpb.LocationsRequest,
) (_ *serverpb.LocationsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188320)
	ctx = s.server.AnnotateCtx(ctx)

	_, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188323)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188324)
	}
	__antithesis_instrumentation__.Notify(188321)

	r, err := s.locationsHelper(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(188325)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188326)
	}
	__antithesis_instrumentation__.Notify(188322)
	return r, nil
}

func (s *adminServer) locationsHelper(
	ctx context.Context, req *serverpb.LocationsRequest,
) (_ *serverpb.LocationsResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188327)
	q := makeSQLQuery()
	q.Append(`SELECT "localityKey", "localityValue", latitude, longitude FROM system.locations`)
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-locations", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		q.String(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188334)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188335)
	}
	__antithesis_instrumentation__.Notify(188328)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(188336)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	__antithesis_instrumentation__.Notify(188329)

	ok, err := it.Next(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188337)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188338)
	}
	__antithesis_instrumentation__.Notify(188330)

	var resp serverpb.LocationsResponse
	if !ok {
		__antithesis_instrumentation__.Notify(188339)

		return &resp, nil
	} else {
		__antithesis_instrumentation__.Notify(188340)
	}
	__antithesis_instrumentation__.Notify(188331)
	scanner := makeResultScanner(it.Types())
	for ; ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(188341)
		row := it.Cur()
		var loc serverpb.LocationsResponse_Location
		lat, lon := new(apd.Decimal), new(apd.Decimal)
		if err := scanner.ScanAll(
			row, &loc.LocalityKey, &loc.LocalityValue, lat, lon); err != nil {
			__antithesis_instrumentation__.Notify(188345)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188346)
		}
		__antithesis_instrumentation__.Notify(188342)
		if loc.Latitude, err = lat.Float64(); err != nil {
			__antithesis_instrumentation__.Notify(188347)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188348)
		}
		__antithesis_instrumentation__.Notify(188343)
		if loc.Longitude, err = lon.Float64(); err != nil {
			__antithesis_instrumentation__.Notify(188349)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188350)
		}
		__antithesis_instrumentation__.Notify(188344)
		resp.Locations = append(resp.Locations, loc)
	}
	__antithesis_instrumentation__.Notify(188332)

	if err != nil {
		__antithesis_instrumentation__.Notify(188351)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188352)
	}
	__antithesis_instrumentation__.Notify(188333)
	return &resp, nil
}

func (s *adminServer) QueryPlan(
	ctx context.Context, req *serverpb.QueryPlanRequest,
) (*serverpb.QueryPlanResponse, error) {
	__antithesis_instrumentation__.Notify(188353)
	ctx = s.server.AnnotateCtx(ctx)

	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188360)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188361)
	}
	__antithesis_instrumentation__.Notify(188354)

	stmts, err := parser.Parse(req.Query)
	if err != nil {
		__antithesis_instrumentation__.Notify(188362)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188363)
	}
	__antithesis_instrumentation__.Notify(188355)
	if len(stmts) > 1 {
		__antithesis_instrumentation__.Notify(188364)
		return nil, serverErrorf(ctx, "more than one query provided")
	} else {
		__antithesis_instrumentation__.Notify(188365)
	}
	__antithesis_instrumentation__.Notify(188356)

	explain := fmt.Sprintf(
		"EXPLAIN (DISTSQL, JSON) %s",
		strings.Trim(req.Query, ";"))
	row, err := s.server.sqlServer.internalExecutor.QueryRowEx(
		ctx, "admin-query-plan", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		explain,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188366)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188367)
	}
	__antithesis_instrumentation__.Notify(188357)
	if row == nil {
		__antithesis_instrumentation__.Notify(188368)
		return nil, serverErrorf(ctx, "failed to query the physical plan")
	} else {
		__antithesis_instrumentation__.Notify(188369)
	}
	__antithesis_instrumentation__.Notify(188358)

	dbDatum, ok := tree.AsDString(row[0])
	if !ok {
		__antithesis_instrumentation__.Notify(188370)
		return nil, serverErrorf(ctx, "type assertion failed on json: %T", row)
	} else {
		__antithesis_instrumentation__.Notify(188371)
	}
	__antithesis_instrumentation__.Notify(188359)

	return &serverpb.QueryPlanResponse{
		DistSQLPhysicalQueryPlan: string(dbDatum),
	}, nil
}

func (s *adminServer) getStatementBundle(ctx context.Context, id int64, w http.ResponseWriter) {
	__antithesis_instrumentation__.Notify(188372)
	sessionUser, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188377)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(188378)
	}
	__antithesis_instrumentation__.Notify(188373)
	row, err := s.server.sqlServer.internalExecutor.QueryRowEx(
		ctx, "admin-stmt-bundle", nil,
		sessiondata.InternalExecutorOverride{User: sessionUser},
		"SELECT bundle_chunks FROM system.statement_diagnostics WHERE id=$1 AND bundle_chunks IS NOT NULL",
		id,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188379)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		__antithesis_instrumentation__.Notify(188380)
	}
	__antithesis_instrumentation__.Notify(188374)
	if row == nil {
		__antithesis_instrumentation__.Notify(188381)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else {
		__antithesis_instrumentation__.Notify(188382)
	}
	__antithesis_instrumentation__.Notify(188375)

	var bundle bytes.Buffer
	chunkIDs := row[0].(*tree.DArray).Array
	for _, chunkID := range chunkIDs {
		__antithesis_instrumentation__.Notify(188383)
		chunkRow, err := s.server.sqlServer.internalExecutor.QueryRowEx(
			ctx, "admin-stmt-bundle", nil,
			sessiondata.InternalExecutorOverride{User: sessionUser},
			"SELECT data FROM system.statement_bundle_chunks WHERE id=$1",
			chunkID,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(188386)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			__antithesis_instrumentation__.Notify(188387)
		}
		__antithesis_instrumentation__.Notify(188384)
		if chunkRow == nil {
			__antithesis_instrumentation__.Notify(188388)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		} else {
			__antithesis_instrumentation__.Notify(188389)
		}
		__antithesis_instrumentation__.Notify(188385)
		data := chunkRow[0].(*tree.DBytes)
		bundle.WriteString(string(*data))
	}
	__antithesis_instrumentation__.Notify(188376)

	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=stmt-bundle-%d.zip", id),
	)

	_, _ = io.Copy(w, &bundle)
}

func (s *adminServer) DecommissionStatus(
	ctx context.Context, req *serverpb.DecommissionStatusRequest,
) (*serverpb.DecommissionStatusResponse, error) {
	__antithesis_instrumentation__.Notify(188390)
	r, err := s.decommissionStatusHelper(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(188392)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188393)
	}
	__antithesis_instrumentation__.Notify(188391)
	return r, nil
}

func (s *adminServer) decommissionStatusHelper(
	ctx context.Context, req *serverpb.DecommissionStatusRequest,
) (*serverpb.DecommissionStatusResponse, error) {
	__antithesis_instrumentation__.Notify(188394)

	nodeIDs := req.NodeIDs

	if len(nodeIDs) == 0 {
		__antithesis_instrumentation__.Notify(188400)
		ns, err := s.server.status.ListNodesInternal(ctx, &serverpb.NodesRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(188402)
			return nil, errors.Wrap(err, "loading node statuses")
		} else {
			__antithesis_instrumentation__.Notify(188403)
		}
		__antithesis_instrumentation__.Notify(188401)
		for _, status := range ns.Nodes {
			__antithesis_instrumentation__.Notify(188404)
			nodeIDs = append(nodeIDs, status.Desc.NodeID)
		}
	} else {
		__antithesis_instrumentation__.Notify(188405)
	}
	__antithesis_instrumentation__.Notify(188395)

	var replicaCounts map[roachpb.NodeID]int64
	if err := s.server.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(188406)
		const pageSize = 10000
		replicaCounts = make(map[roachpb.NodeID]int64)
		for _, nodeID := range nodeIDs {
			__antithesis_instrumentation__.Notify(188408)
			replicaCounts[nodeID] = 0
		}
		__antithesis_instrumentation__.Notify(188407)
		return txn.Iterate(ctx, keys.Meta2Prefix, keys.MetaMax, pageSize,
			func(rows []kv.KeyValue) error {
				__antithesis_instrumentation__.Notify(188409)
				rangeDesc := roachpb.RangeDescriptor{}
				for _, row := range rows {
					__antithesis_instrumentation__.Notify(188411)
					if err := row.ValueProto(&rangeDesc); err != nil {
						__antithesis_instrumentation__.Notify(188413)
						return errors.Wrapf(err, "%s: unable to unmarshal range descriptor", row.Key)
					} else {
						__antithesis_instrumentation__.Notify(188414)
					}
					__antithesis_instrumentation__.Notify(188412)
					for _, r := range rangeDesc.Replicas().Descriptors() {
						__antithesis_instrumentation__.Notify(188415)
						if _, ok := replicaCounts[r.NodeID]; ok {
							__antithesis_instrumentation__.Notify(188416)
							replicaCounts[r.NodeID]++
						} else {
							__antithesis_instrumentation__.Notify(188417)
						}
					}
				}
				__antithesis_instrumentation__.Notify(188410)
				return nil
			})
	}); err != nil {
		__antithesis_instrumentation__.Notify(188418)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188419)
	}
	__antithesis_instrumentation__.Notify(188396)

	var res serverpb.DecommissionStatusResponse
	livenessMap := map[roachpb.NodeID]livenesspb.Liveness{}
	{
		__antithesis_instrumentation__.Notify(188420)

		ls, err := s.server.nodeLiveness.GetLivenessesFromKV(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(188422)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188423)
		}
		__antithesis_instrumentation__.Notify(188421)
		for _, rec := range ls {
			__antithesis_instrumentation__.Notify(188424)
			livenessMap[rec.NodeID] = rec
		}
	}
	__antithesis_instrumentation__.Notify(188397)

	for nodeID := range replicaCounts {
		__antithesis_instrumentation__.Notify(188425)
		l, ok := livenessMap[nodeID]
		if !ok {
			__antithesis_instrumentation__.Notify(188428)
			return nil, errors.Newf("unable to get liveness for %d", nodeID)
		} else {
			__antithesis_instrumentation__.Notify(188429)
		}
		__antithesis_instrumentation__.Notify(188426)
		nodeResp := serverpb.DecommissionStatusResponse_Status{
			NodeID:       l.NodeID,
			ReplicaCount: replicaCounts[l.NodeID],
			Membership:   l.Membership,
			Draining:     l.Draining,
		}
		if l.IsLive(s.server.clock.Now().GoTime()) {
			__antithesis_instrumentation__.Notify(188430)
			nodeResp.IsLive = true
		} else {
			__antithesis_instrumentation__.Notify(188431)
		}
		__antithesis_instrumentation__.Notify(188427)

		res.Status = append(res.Status, nodeResp)
	}
	__antithesis_instrumentation__.Notify(188398)

	sort.Slice(res.Status, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(188432)
		return res.Status[i].NodeID < res.Status[j].NodeID
	})
	__antithesis_instrumentation__.Notify(188399)

	return &res, nil
}

func (s *adminServer) Decommission(
	ctx context.Context, req *serverpb.DecommissionRequest,
) (*serverpb.DecommissionStatusResponse, error) {
	__antithesis_instrumentation__.Notify(188433)
	nodeIDs := req.NodeIDs

	if len(nodeIDs) == 0 {
		__antithesis_instrumentation__.Notify(188437)
		return nil, status.Errorf(codes.InvalidArgument, "no node ID specified")
	} else {
		__antithesis_instrumentation__.Notify(188438)
	}
	__antithesis_instrumentation__.Notify(188434)

	if err := s.server.Decommission(ctx, req.TargetMembership, nodeIDs); err != nil {
		__antithesis_instrumentation__.Notify(188439)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188440)
	}
	__antithesis_instrumentation__.Notify(188435)

	if req.TargetMembership == livenesspb.MembershipStatus_DECOMMISSIONED {
		__antithesis_instrumentation__.Notify(188441)
		return &serverpb.DecommissionStatusResponse{}, nil
	} else {
		__antithesis_instrumentation__.Notify(188442)
	}
	__antithesis_instrumentation__.Notify(188436)

	return s.DecommissionStatus(ctx, &serverpb.DecommissionStatusRequest{NodeIDs: nodeIDs})
}

func (s *adminServer) DataDistribution(
	ctx context.Context, req *serverpb.DataDistributionRequest,
) (_ *serverpb.DataDistributionResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188443)
	if _, err := s.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(188447)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188448)
	}
	__antithesis_instrumentation__.Notify(188444)

	userName, err := userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188449)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188450)
	}
	__antithesis_instrumentation__.Notify(188445)

	r, err := s.dataDistributionHelper(ctx, req, userName)
	if err != nil {
		__antithesis_instrumentation__.Notify(188451)
		return nil, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188452)
	}
	__antithesis_instrumentation__.Notify(188446)
	return r, nil
}

func (s *adminServer) dataDistributionHelper(
	ctx context.Context, req *serverpb.DataDistributionRequest, userName security.SQLUsername,
) (resp *serverpb.DataDistributionResponse, retErr error) {
	__antithesis_instrumentation__.Notify(188453)
	resp = &serverpb.DataDistributionResponse{
		DatabaseInfo: make(map[string]serverpb.DataDistributionResponse_DatabaseInfo),
		ZoneConfigs:  make(map[string]serverpb.DataDistributionResponse_ZoneConfig),
	}

	tablesQuery := `SELECT name, schema_name, table_id, database_name, drop_time FROM
									"".crdb_internal.tables WHERE database_name IS NOT NULL`
	it, err := s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-replica-matrix", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		tablesQuery,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188463)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188464)
	}
	__antithesis_instrumentation__.Notify(188454)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(188465)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	__antithesis_instrumentation__.Notify(188455)

	tableInfosByTableID := map[uint32]serverpb.DataDistributionResponse_TableInfo{}

	var hasNext bool
	for hasNext, err = it.Next(ctx); hasNext; hasNext, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(188466)
		row := it.Cur()
		tableName := (*string)(row[0].(*tree.DString))
		schemaName := (*string)(row[1].(*tree.DString))
		fqTableName := fmt.Sprintf("%s.%s",
			tree.NameStringP(schemaName), tree.NameStringP(tableName))
		tableID := uint32(tree.MustBeDInt(row[2]))
		dbName := (*string)(row[3].(*tree.DString))

		var droppedAtTime *time.Time
		droppedAtDatum, ok := row[4].(*tree.DTimestamp)
		if ok {
			__antithesis_instrumentation__.Notify(188470)
			droppedAtTime = &droppedAtDatum.Time
		} else {
			__antithesis_instrumentation__.Notify(188471)
		}
		__antithesis_instrumentation__.Notify(188467)

		dbInfo, ok := resp.DatabaseInfo[*dbName]
		if !ok {
			__antithesis_instrumentation__.Notify(188472)
			dbInfo = serverpb.DataDistributionResponse_DatabaseInfo{
				TableInfo: make(map[string]serverpb.DataDistributionResponse_TableInfo),
			}
			resp.DatabaseInfo[*dbName] = dbInfo
		} else {
			__antithesis_instrumentation__.Notify(188473)
		}
		__antithesis_instrumentation__.Notify(188468)

		zcID := int64(0)

		if droppedAtTime == nil {
			__antithesis_instrumentation__.Notify(188474)

			zoneConfigQuery := fmt.Sprintf(
				`SELECT zone_id FROM [SHOW ZONE CONFIGURATION FOR TABLE %s.%s.%s]`,
				(*tree.Name)(dbName), (*tree.Name)(schemaName), (*tree.Name)(tableName),
			)
			row, err := s.server.sqlServer.internalExecutor.QueryRowEx(
				ctx, "admin-replica-matrix", nil,
				sessiondata.InternalExecutorOverride{User: userName},
				zoneConfigQuery,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(188477)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(188478)
			}
			__antithesis_instrumentation__.Notify(188475)
			if row == nil {
				__antithesis_instrumentation__.Notify(188479)
				return nil, errors.Errorf(
					"could not get zone config for table %s; 0 rows returned", *tableName,
				)
			} else {
				__antithesis_instrumentation__.Notify(188480)
			}
			__antithesis_instrumentation__.Notify(188476)

			zcID = int64(tree.MustBeDInt(row[0]))
		} else {
			__antithesis_instrumentation__.Notify(188481)
		}
		__antithesis_instrumentation__.Notify(188469)

		tableInfo := serverpb.DataDistributionResponse_TableInfo{
			ReplicaCountByNodeId: make(map[roachpb.NodeID]int64),
			ZoneConfigId:         zcID,
			DroppedAt:            droppedAtTime,
		}
		dbInfo.TableInfo[fqTableName] = tableInfo
		tableInfosByTableID[tableID] = tableInfo
	}
	__antithesis_instrumentation__.Notify(188456)
	if err != nil {
		__antithesis_instrumentation__.Notify(188482)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188483)
	}
	__antithesis_instrumentation__.Notify(188457)

	if err := s.server.db.Txn(ctx, func(txnCtx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(188484)
		acct := s.memMonitor.MakeBoundAccount()
		defer acct.Close(txnCtx)

		kvs, err := kvclient.ScanMetaKVs(ctx, txn, roachpb.Span{
			Key:    keys.SystemSQLCodec.TablePrefix(keys.MaxReservedDescID + 1),
			EndKey: keys.MaxKey,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(188487)
			return err
		} else {
			__antithesis_instrumentation__.Notify(188488)
		}
		__antithesis_instrumentation__.Notify(188485)

		var rangeDesc roachpb.RangeDescriptor
		for _, kv := range kvs {
			__antithesis_instrumentation__.Notify(188489)
			if err := acct.Grow(txnCtx, int64(len(kv.Key)+len(kv.Value.RawBytes))); err != nil {
				__antithesis_instrumentation__.Notify(188494)
				return err
			} else {
				__antithesis_instrumentation__.Notify(188495)
			}
			__antithesis_instrumentation__.Notify(188490)
			if err := kv.ValueProto(&rangeDesc); err != nil {
				__antithesis_instrumentation__.Notify(188496)
				return err
			} else {
				__antithesis_instrumentation__.Notify(188497)
			}
			__antithesis_instrumentation__.Notify(188491)

			_, tenID, err := keys.DecodeTenantPrefix(rangeDesc.StartKey.AsRawKey())
			if err != nil {
				__antithesis_instrumentation__.Notify(188498)
				return err
			} else {
				__antithesis_instrumentation__.Notify(188499)
			}
			__antithesis_instrumentation__.Notify(188492)
			_, tableID, err := keys.MakeSQLCodec(tenID).DecodeTablePrefix(rangeDesc.StartKey.AsRawKey())
			if err != nil {
				__antithesis_instrumentation__.Notify(188500)
				return err
			} else {
				__antithesis_instrumentation__.Notify(188501)
			}
			__antithesis_instrumentation__.Notify(188493)
			for _, replicaDesc := range rangeDesc.Replicas().Descriptors() {
				__antithesis_instrumentation__.Notify(188502)
				tableInfo, found := tableInfosByTableID[tableID]
				if !found {
					__antithesis_instrumentation__.Notify(188504)

					continue
				} else {
					__antithesis_instrumentation__.Notify(188505)
				}
				__antithesis_instrumentation__.Notify(188503)
				tableInfo.ReplicaCountByNodeId[replicaDesc.NodeID]++
			}
		}
		__antithesis_instrumentation__.Notify(188486)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(188506)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188507)
	}
	__antithesis_instrumentation__.Notify(188458)

	zoneConfigsQuery := `
		SELECT target, raw_config_sql, raw_config_protobuf
		FROM crdb_internal.zones
		WHERE target IS NOT NULL
	`
	it, err = s.server.sqlServer.internalExecutor.QueryIteratorEx(
		ctx, "admin-replica-matrix", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		zoneConfigsQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(188508)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188509)
	}
	__antithesis_instrumentation__.Notify(188459)

	defer func(it sqlutil.InternalRows) {
		__antithesis_instrumentation__.Notify(188510)
		retErr = errors.CombineErrors(retErr, it.Close())
	}(it)
	__antithesis_instrumentation__.Notify(188460)

	for hasNext, err = it.Next(ctx); hasNext; hasNext, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(188511)
		row := it.Cur()
		target := string(tree.MustBeDString(row[0]))
		zcSQL := tree.MustBeDString(row[1])
		zcBytes := tree.MustBeDBytes(row[2])
		var zcProto zonepb.ZoneConfig
		if err := protoutil.Unmarshal([]byte(zcBytes), &zcProto); err != nil {
			__antithesis_instrumentation__.Notify(188513)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188514)
		}
		__antithesis_instrumentation__.Notify(188512)

		resp.ZoneConfigs[target] = serverpb.DataDistributionResponse_ZoneConfig{
			Target:    target,
			Config:    zcProto,
			ConfigSQL: string(zcSQL),
		}
	}
	__antithesis_instrumentation__.Notify(188461)
	if err != nil {
		__antithesis_instrumentation__.Notify(188515)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188516)
	}
	__antithesis_instrumentation__.Notify(188462)

	return resp, nil
}

func (s *adminServer) EnqueueRange(
	ctx context.Context, req *serverpb.EnqueueRangeRequest,
) (*serverpb.EnqueueRangeResponse, error) {
	__antithesis_instrumentation__.Notify(188517)
	ctx = propagateGatewayMetadata(ctx)
	ctx = s.server.AnnotateCtx(ctx)

	if _, err := s.requireAdminUser(ctx); err != nil {
		__antithesis_instrumentation__.Notify(188528)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188529)
	}
	__antithesis_instrumentation__.Notify(188518)

	if req.NodeID < 0 {
		__antithesis_instrumentation__.Notify(188530)
		return nil, status.Errorf(codes.InvalidArgument, "node_id must be non-negative; got %d", req.NodeID)
	} else {
		__antithesis_instrumentation__.Notify(188531)
	}
	__antithesis_instrumentation__.Notify(188519)
	if req.Queue == "" {
		__antithesis_instrumentation__.Notify(188532)
		return nil, status.Errorf(codes.InvalidArgument, "queue name must be non-empty")
	} else {
		__antithesis_instrumentation__.Notify(188533)
	}
	__antithesis_instrumentation__.Notify(188520)
	if req.RangeID <= 0 {
		__antithesis_instrumentation__.Notify(188534)
		return nil, status.Errorf(codes.InvalidArgument, "range_id must be positive; got %d", req.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(188535)
	}
	__antithesis_instrumentation__.Notify(188521)

	if req.NodeID == s.server.NodeID() {
		__antithesis_instrumentation__.Notify(188536)
		return s.enqueueRangeLocal(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(188537)
		if req.NodeID != 0 {
			__antithesis_instrumentation__.Notify(188538)
			admin, err := s.dialNode(ctx, req.NodeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(188540)
				return nil, serverError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(188541)
			}
			__antithesis_instrumentation__.Notify(188539)
			return admin.EnqueueRange(ctx, req)
		} else {
			__antithesis_instrumentation__.Notify(188542)
		}
	}
	__antithesis_instrumentation__.Notify(188522)

	response := &serverpb.EnqueueRangeResponse{}

	dialFn := func(ctx context.Context, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(188543)
		client, err := s.dialNode(ctx, nodeID)
		return client, err
	}
	__antithesis_instrumentation__.Notify(188523)
	nodeFn := func(ctx context.Context, client interface{}, nodeID roachpb.NodeID) (interface{}, error) {
		__antithesis_instrumentation__.Notify(188544)
		admin := client.(serverpb.AdminClient)
		req := *req
		req.NodeID = nodeID
		return admin.EnqueueRange(ctx, &req)
	}
	__antithesis_instrumentation__.Notify(188524)
	responseFn := func(_ roachpb.NodeID, nodeResp interface{}) {
		__antithesis_instrumentation__.Notify(188545)
		nodeDetails := nodeResp.(*serverpb.EnqueueRangeResponse)
		response.Details = append(response.Details, nodeDetails.Details...)
	}
	__antithesis_instrumentation__.Notify(188525)
	errorFn := func(nodeID roachpb.NodeID, err error) {
		__antithesis_instrumentation__.Notify(188546)
		errDetail := &serverpb.EnqueueRangeResponse_Details{
			NodeID: nodeID,
			Error:  err.Error(),
		}
		response.Details = append(response.Details, errDetail)
	}
	__antithesis_instrumentation__.Notify(188526)

	if err := contextutil.RunWithTimeout(ctx, "enqueue range", time.Minute, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(188547)
		return s.server.status.iterateNodes(
			ctx, fmt.Sprintf("enqueue r%d in queue %s", req.RangeID, req.Queue),
			dialFn, nodeFn, responseFn, errorFn,
		)
	}); err != nil {
		__antithesis_instrumentation__.Notify(188548)
		if len(response.Details) == 0 {
			__antithesis_instrumentation__.Notify(188550)
			return nil, serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(188551)
		}
		__antithesis_instrumentation__.Notify(188549)
		response.Details = append(response.Details, &serverpb.EnqueueRangeResponse_Details{
			Error: err.Error(),
		})
	} else {
		__antithesis_instrumentation__.Notify(188552)
	}
	__antithesis_instrumentation__.Notify(188527)

	return response, nil
}

func (s *adminServer) enqueueRangeLocal(
	ctx context.Context, req *serverpb.EnqueueRangeRequest,
) (*serverpb.EnqueueRangeResponse, error) {
	__antithesis_instrumentation__.Notify(188553)
	response := &serverpb.EnqueueRangeResponse{
		Details: []*serverpb.EnqueueRangeResponse_Details{
			{
				NodeID: s.server.NodeID(),
			},
		},
	}

	var store *kvserver.Store
	var repl *kvserver.Replica
	if err := s.server.node.stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(188559)
		r := s.GetReplicaIfExists(req.RangeID)
		if r == nil {
			__antithesis_instrumentation__.Notify(188561)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(188562)
		}
		__antithesis_instrumentation__.Notify(188560)
		repl = r
		store = s
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(188563)
		response.Details[0].Error = err.Error()
		return response, nil
	} else {
		__antithesis_instrumentation__.Notify(188564)
	}
	__antithesis_instrumentation__.Notify(188554)

	if store == nil || func() bool {
		__antithesis_instrumentation__.Notify(188565)
		return repl == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(188566)
		response.Details[0].Error = fmt.Sprintf("n%d has no replica for r%d", s.server.NodeID(), req.RangeID)
		return response, nil
	} else {
		__antithesis_instrumentation__.Notify(188567)
	}
	__antithesis_instrumentation__.Notify(188555)

	queueName := req.Queue
	if strings.ToLower(queueName) == "gc" {
		__antithesis_instrumentation__.Notify(188568)
		queueName = "mvccGC"
	} else {
		__antithesis_instrumentation__.Notify(188569)
	}
	__antithesis_instrumentation__.Notify(188556)

	traceSpans, processErr, err := store.ManuallyEnqueue(ctx, queueName, repl, req.SkipShouldQueue)
	if err != nil {
		__antithesis_instrumentation__.Notify(188570)
		response.Details[0].Error = err.Error()
		return response, nil
	} else {
		__antithesis_instrumentation__.Notify(188571)
	}
	__antithesis_instrumentation__.Notify(188557)
	response.Details[0].Events = recordedSpansToTraceEvents(traceSpans)
	if processErr != nil {
		__antithesis_instrumentation__.Notify(188572)
		response.Details[0].Error = processErr.Error()
	} else {
		__antithesis_instrumentation__.Notify(188573)
	}
	__antithesis_instrumentation__.Notify(188558)
	return response, nil
}

func (s *adminServer) SendKVBatch(
	ctx context.Context, ba *roachpb.BatchRequest,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(188574)
	ctx = s.server.AnnotateCtx(ctx)

	user, err := s.requireAdminUser(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188579)

		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188580)
	}
	__antithesis_instrumentation__.Notify(188575)
	if ba == nil {
		__antithesis_instrumentation__.Notify(188581)
		return nil, status.Errorf(codes.InvalidArgument, "BatchRequest cannot be nil")
	} else {
		__antithesis_instrumentation__.Notify(188582)
	}
	__antithesis_instrumentation__.Notify(188576)

	jsonpb := protoutil.JSONPb{}
	baJSON, err := jsonpb.Marshal(ba)
	if err != nil {
		__antithesis_instrumentation__.Notify(188583)
		return nil, serverError(ctx, errors.Wrap(err, "failed to encode BatchRequest as JSON"))
	} else {
		__antithesis_instrumentation__.Notify(188584)
	}
	__antithesis_instrumentation__.Notify(188577)
	event := &eventpb.DebugSendKvBatch{
		CommonEventDetails: eventpb.CommonEventDetails{
			Timestamp: timeutil.Now().UnixNano(),
		},
		CommonDebugEventDetails: eventpb.CommonDebugEventDetails{
			NodeID: int32(s.server.NodeID()),
			User:   user.Normalized(),
		},
		BatchRequest: string(baJSON),
	}
	log.StructuredEvent(ctx, event)

	ctx, sp := s.server.node.setupSpanForIncomingRPC(ctx, roachpb.SystemTenantID, ba)
	var br *roachpb.BatchResponse
	defer func() { __antithesis_instrumentation__.Notify(188585); sp.finish(ctx, br) }()
	__antithesis_instrumentation__.Notify(188578)
	br, pErr := s.server.db.NonTransactionalSender().Send(ctx, *ba)
	br.Error = pErr
	return br, nil
}

type sqlQuery struct {
	buf   bytes.Buffer
	pidx  int
	qargs []interface{}
	errs  []error
}

func makeSQLQuery() *sqlQuery {
	__antithesis_instrumentation__.Notify(188586)
	res := &sqlQuery{}
	return res
}

func (q *sqlQuery) String() string {
	__antithesis_instrumentation__.Notify(188587)
	if len(q.errs) > 0 {
		__antithesis_instrumentation__.Notify(188589)
		return "couldn't generate query: please check Errors()"
	} else {
		__antithesis_instrumentation__.Notify(188590)
	}
	__antithesis_instrumentation__.Notify(188588)
	return q.buf.String()
}

func (q *sqlQuery) Errors() []error {
	__antithesis_instrumentation__.Notify(188591)
	return q.errs
}

func (q *sqlQuery) QueryArguments() []interface{} {
	__antithesis_instrumentation__.Notify(188592)
	return q.qargs
}

func (q *sqlQuery) Append(s string, params ...interface{}) {
	__antithesis_instrumentation__.Notify(188593)
	var placeholders int
	for _, r := range s {
		__antithesis_instrumentation__.Notify(188596)
		q.buf.WriteRune(r)
		if r == '$' {
			__antithesis_instrumentation__.Notify(188597)
			q.pidx++
			placeholders++
			q.buf.WriteString(strconv.Itoa(q.pidx))
		} else {
			__antithesis_instrumentation__.Notify(188598)
		}
	}
	__antithesis_instrumentation__.Notify(188594)

	if placeholders != len(params) {
		__antithesis_instrumentation__.Notify(188599)
		q.errs = append(q.errs,
			errors.Errorf("# of placeholders %d != # of params %d", placeholders, len(params)))
	} else {
		__antithesis_instrumentation__.Notify(188600)
	}
	__antithesis_instrumentation__.Notify(188595)
	q.qargs = append(q.qargs, params...)
}

type resultScanner struct {
	colNameToIdx map[string]int
}

func makeResultScanner(cols []colinfo.ResultColumn) resultScanner {
	__antithesis_instrumentation__.Notify(188601)
	rs := resultScanner{
		colNameToIdx: make(map[string]int),
	}
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(188603)
		rs.colNameToIdx[col.Name] = i
	}
	__antithesis_instrumentation__.Notify(188602)
	return rs
}

func (rs resultScanner) IsNull(row tree.Datums, col string) (bool, error) {
	__antithesis_instrumentation__.Notify(188604)
	idx, ok := rs.colNameToIdx[col]
	if !ok {
		__antithesis_instrumentation__.Notify(188606)
		return false, errors.Errorf("result is missing column %s", col)
	} else {
		__antithesis_instrumentation__.Notify(188607)
	}
	__antithesis_instrumentation__.Notify(188605)
	return row[idx] == tree.DNull, nil
}

func (rs resultScanner) ScanIndex(row tree.Datums, index int, dst interface{}) error {
	__antithesis_instrumentation__.Notify(188608)
	src := row[index]

	if dst == nil {
		__antithesis_instrumentation__.Notify(188611)
		return errors.Errorf("nil destination pointer passed in")
	} else {
		__antithesis_instrumentation__.Notify(188612)
	}
	__antithesis_instrumentation__.Notify(188609)

	switch d := dst.(type) {
	case *string:
		__antithesis_instrumentation__.Notify(188613)
		s, ok := tree.AsDString(src)
		if !ok {
			__antithesis_instrumentation__.Notify(188638)
			return errors.Errorf("source type assertion failed")
		} else {
			__antithesis_instrumentation__.Notify(188639)
		}
		__antithesis_instrumentation__.Notify(188614)
		*d = string(s)

	case **string:
		__antithesis_instrumentation__.Notify(188615)
		s, ok := tree.AsDString(src)
		if !ok {
			__antithesis_instrumentation__.Notify(188640)
			if src != tree.DNull {
				__antithesis_instrumentation__.Notify(188642)
				return errors.Errorf("source type assertion failed")
			} else {
				__antithesis_instrumentation__.Notify(188643)
			}
			__antithesis_instrumentation__.Notify(188641)
			*d = nil
			break
		} else {
			__antithesis_instrumentation__.Notify(188644)
		}
		__antithesis_instrumentation__.Notify(188616)
		val := string(s)
		*d = &val

	case *bool:
		__antithesis_instrumentation__.Notify(188617)
		s, ok := src.(*tree.DBool)
		if !ok {
			__antithesis_instrumentation__.Notify(188645)
			return errors.Errorf("source type assertion failed")
		} else {
			__antithesis_instrumentation__.Notify(188646)
		}
		__antithesis_instrumentation__.Notify(188618)
		*d = bool(*s)

	case *float32:
		__antithesis_instrumentation__.Notify(188619)
		s, ok := src.(*tree.DFloat)
		if !ok {
			__antithesis_instrumentation__.Notify(188647)
			return errors.Errorf("source type assertion failed")
		} else {
			__antithesis_instrumentation__.Notify(188648)
		}
		__antithesis_instrumentation__.Notify(188620)
		*d = float32(*s)

	case **float32:
		__antithesis_instrumentation__.Notify(188621)
		s, ok := src.(*tree.DFloat)
		if !ok {
			__antithesis_instrumentation__.Notify(188649)
			if src != tree.DNull {
				__antithesis_instrumentation__.Notify(188651)
				return errors.Errorf("source type assertion failed")
			} else {
				__antithesis_instrumentation__.Notify(188652)
			}
			__antithesis_instrumentation__.Notify(188650)
			*d = nil
			break
		} else {
			__antithesis_instrumentation__.Notify(188653)
		}
		__antithesis_instrumentation__.Notify(188622)
		val := float32(*s)
		*d = &val

	case *int64:
		__antithesis_instrumentation__.Notify(188623)
		s, ok := tree.AsDInt(src)
		if !ok {
			__antithesis_instrumentation__.Notify(188654)
			return errors.Errorf("source type assertion failed")
		} else {
			__antithesis_instrumentation__.Notify(188655)
		}
		__antithesis_instrumentation__.Notify(188624)
		*d = int64(s)

	case *[]descpb.ID:
		__antithesis_instrumentation__.Notify(188625)
		s, ok := tree.AsDArray(src)
		if !ok {
			__antithesis_instrumentation__.Notify(188656)
			return errors.Errorf("source type assertion failed")
		} else {
			__antithesis_instrumentation__.Notify(188657)
		}
		__antithesis_instrumentation__.Notify(188626)
		for i := 0; i < s.Len(); i++ {
			__antithesis_instrumentation__.Notify(188658)
			id, ok := tree.AsDInt(s.Array[i])
			if !ok {
				__antithesis_instrumentation__.Notify(188660)
				return errors.Errorf("source type assertion failed on index %d", i)
			} else {
				__antithesis_instrumentation__.Notify(188661)
			}
			__antithesis_instrumentation__.Notify(188659)
			*d = append(*d, descpb.ID(id))
		}

	case *time.Time:
		__antithesis_instrumentation__.Notify(188627)
		s, ok := src.(*tree.DTimestamp)
		if !ok {
			__antithesis_instrumentation__.Notify(188662)
			return errors.Errorf("source type assertion failed")
		} else {
			__antithesis_instrumentation__.Notify(188663)
		}
		__antithesis_instrumentation__.Notify(188628)
		*d = s.Time

	case **time.Time:
		__antithesis_instrumentation__.Notify(188629)
		s, ok := src.(*tree.DTimestamp)
		if !ok {
			__antithesis_instrumentation__.Notify(188664)
			if src != tree.DNull {
				__antithesis_instrumentation__.Notify(188666)
				return errors.Errorf("source type assertion failed")
			} else {
				__antithesis_instrumentation__.Notify(188667)
			}
			__antithesis_instrumentation__.Notify(188665)
			*d = nil
			break
		} else {
			__antithesis_instrumentation__.Notify(188668)
		}
		__antithesis_instrumentation__.Notify(188630)
		*d = &s.Time

	case *[]byte:
		__antithesis_instrumentation__.Notify(188631)
		s, ok := src.(*tree.DBytes)
		if !ok {
			__antithesis_instrumentation__.Notify(188669)
			return errors.Errorf("source type assertion failed")
		} else {
			__antithesis_instrumentation__.Notify(188670)
		}
		__antithesis_instrumentation__.Notify(188632)

		*d = []byte(*s)

	case *apd.Decimal:
		__antithesis_instrumentation__.Notify(188633)
		s, ok := src.(*tree.DDecimal)
		if !ok {
			__antithesis_instrumentation__.Notify(188671)
			return errors.Errorf("source type assertion failed")
		} else {
			__antithesis_instrumentation__.Notify(188672)
		}
		__antithesis_instrumentation__.Notify(188634)
		*d = s.Decimal

	case **apd.Decimal:
		__antithesis_instrumentation__.Notify(188635)
		s, ok := src.(*tree.DDecimal)
		if !ok {
			__antithesis_instrumentation__.Notify(188673)
			if src != tree.DNull {
				__antithesis_instrumentation__.Notify(188675)
				return errors.Errorf("source type assertion failed")
			} else {
				__antithesis_instrumentation__.Notify(188676)
			}
			__antithesis_instrumentation__.Notify(188674)
			*d = nil
			break
		} else {
			__antithesis_instrumentation__.Notify(188677)
		}
		__antithesis_instrumentation__.Notify(188636)
		*d = &s.Decimal

	default:
		__antithesis_instrumentation__.Notify(188637)
		return errors.Errorf("unimplemented type for scanCol: %T", dst)
	}
	__antithesis_instrumentation__.Notify(188610)

	return nil
}

func (rs resultScanner) ScanAll(row tree.Datums, dsts ...interface{}) error {
	__antithesis_instrumentation__.Notify(188678)
	if len(row) != len(dsts) {
		__antithesis_instrumentation__.Notify(188681)
		return fmt.Errorf(
			"ScanAll: row has %d columns but %d dests provided", len(row), len(dsts))
	} else {
		__antithesis_instrumentation__.Notify(188682)
	}
	__antithesis_instrumentation__.Notify(188679)
	for i := 0; i < len(row); i++ {
		__antithesis_instrumentation__.Notify(188683)
		if err := rs.ScanIndex(row, i, dsts[i]); err != nil {
			__antithesis_instrumentation__.Notify(188684)
			return err
		} else {
			__antithesis_instrumentation__.Notify(188685)
		}
	}
	__antithesis_instrumentation__.Notify(188680)
	return nil
}

func (rs resultScanner) Scan(row tree.Datums, colName string, dst interface{}) error {
	__antithesis_instrumentation__.Notify(188686)
	idx, ok := rs.colNameToIdx[colName]
	if !ok {
		__antithesis_instrumentation__.Notify(188688)
		return errors.Errorf("result is missing column %s", colName)
	} else {
		__antithesis_instrumentation__.Notify(188689)
	}
	__antithesis_instrumentation__.Notify(188687)
	return rs.ScanIndex(row, idx, dst)
}

func (s *adminServer) queryZone(
	ctx context.Context, userName security.SQLUsername, id descpb.ID,
) (zonepb.ZoneConfig, bool, error) {
	__antithesis_instrumentation__.Notify(188690)
	const query = `SELECT crdb_internal.get_zone_config($1)`
	row, cols, err := s.server.sqlServer.internalExecutor.QueryRowExWithCols(
		ctx,
		"admin-query-zone",
		nil,
		sessiondata.InternalExecutorOverride{User: userName},
		query,
		id,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188696)
		return *zonepb.NewZoneConfig(), false, err
	} else {
		__antithesis_instrumentation__.Notify(188697)
	}
	__antithesis_instrumentation__.Notify(188691)

	if row == nil {
		__antithesis_instrumentation__.Notify(188698)
		return *zonepb.NewZoneConfig(), false, errors.Errorf("invalid number of rows (0) returned: %s (%d)", query, id)
	} else {
		__antithesis_instrumentation__.Notify(188699)
	}
	__antithesis_instrumentation__.Notify(188692)

	var zoneBytes []byte
	scanner := makeResultScanner(cols)
	if isNull, err := scanner.IsNull(row, cols[0].Name); err != nil {
		__antithesis_instrumentation__.Notify(188700)
		return *zonepb.NewZoneConfig(), false, err
	} else {
		__antithesis_instrumentation__.Notify(188701)
		if isNull {
			__antithesis_instrumentation__.Notify(188702)
			return *zonepb.NewZoneConfig(), false, nil
		} else {
			__antithesis_instrumentation__.Notify(188703)
		}
	}
	__antithesis_instrumentation__.Notify(188693)

	err = scanner.ScanIndex(row, 0, &zoneBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(188704)
		return *zonepb.NewZoneConfig(), false, err
	} else {
		__antithesis_instrumentation__.Notify(188705)
	}
	__antithesis_instrumentation__.Notify(188694)

	var zone zonepb.ZoneConfig
	if err := protoutil.Unmarshal(zoneBytes, &zone); err != nil {
		__antithesis_instrumentation__.Notify(188706)
		return *zonepb.NewZoneConfig(), false, err
	} else {
		__antithesis_instrumentation__.Notify(188707)
	}
	__antithesis_instrumentation__.Notify(188695)
	return zone, true, nil
}

func (s *adminServer) queryZonePath(
	ctx context.Context, userName security.SQLUsername, path []descpb.ID,
) (descpb.ID, zonepb.ZoneConfig, bool, error) {
	__antithesis_instrumentation__.Notify(188708)
	for i := len(path) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(188710)
		zone, zoneExists, err := s.queryZone(ctx, userName, path[i])
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(188711)
			return zoneExists == true
		}() == true {
			__antithesis_instrumentation__.Notify(188712)
			return path[i], zone, true, err
		} else {
			__antithesis_instrumentation__.Notify(188713)
		}
	}
	__antithesis_instrumentation__.Notify(188709)
	return 0, *zonepb.NewZoneConfig(), false, nil
}

func (s *adminServer) queryDatabaseID(
	ctx context.Context, userName security.SQLUsername, name string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(188714)
	const query = `SELECT crdb_internal.get_database_id($1)`
	row, cols, err := s.server.sqlServer.internalExecutor.QueryRowExWithCols(
		ctx, "admin-query-namespace-ID", nil,
		sessiondata.InternalExecutorOverride{User: userName},
		query, name,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188719)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(188720)
	}
	__antithesis_instrumentation__.Notify(188715)

	if row == nil {
		__antithesis_instrumentation__.Notify(188721)
		return 0, errors.Errorf("invalid number of rows (0) returned: %s (%s)", query, name)
	} else {
		__antithesis_instrumentation__.Notify(188722)
	}
	__antithesis_instrumentation__.Notify(188716)

	var id int64
	scanner := makeResultScanner(cols)
	if isNull, err := scanner.IsNull(row, cols[0].Name); err != nil {
		__antithesis_instrumentation__.Notify(188723)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(188724)
		if isNull {
			__antithesis_instrumentation__.Notify(188725)
			return 0, errors.Errorf("database %s not found", name)
		} else {
			__antithesis_instrumentation__.Notify(188726)
		}
	}
	__antithesis_instrumentation__.Notify(188717)

	err = scanner.ScanIndex(row, 0, &id)
	if err != nil {
		__antithesis_instrumentation__.Notify(188727)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(188728)
	}
	__antithesis_instrumentation__.Notify(188718)

	return descpb.ID(id), nil
}

func (s *adminServer) queryTableID(
	ctx context.Context, username security.SQLUsername, database string, tableName string,
) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(188729)
	row, err := s.server.sqlServer.internalExecutor.QueryRowEx(
		ctx, "admin-resolve-name", nil,
		sessiondata.InternalExecutorOverride{User: username, Database: database},
		"SELECT $1::regclass::oid", tableName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(188732)
		return descpb.InvalidID, err
	} else {
		__antithesis_instrumentation__.Notify(188733)
	}
	__antithesis_instrumentation__.Notify(188730)
	if row == nil {
		__antithesis_instrumentation__.Notify(188734)
		return descpb.InvalidID, errors.Newf("failed to resolve %q as a table name", tableName)
	} else {
		__antithesis_instrumentation__.Notify(188735)
	}
	__antithesis_instrumentation__.Notify(188731)
	return descpb.ID(tree.MustBeDOid(row[0]).DInt), nil
}

func (s *adminServer) dialNode(
	ctx context.Context, nodeID roachpb.NodeID,
) (serverpb.AdminClient, error) {
	__antithesis_instrumentation__.Notify(188736)
	addr, err := s.server.gossip.GetNodeIDAddress(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(188739)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188740)
	}
	__antithesis_instrumentation__.Notify(188737)
	conn, err := s.server.rpcContext.GRPCDialNode(
		addr.String(), nodeID, rpc.DefaultClass).Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188741)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188742)
	}
	__antithesis_instrumentation__.Notify(188738)
	return serverpb.NewAdminClient(conn), nil
}

type adminPrivilegeChecker struct {
	ie *sql.InternalExecutor
}

func (c *adminPrivilegeChecker) requireAdminUser(
	ctx context.Context,
) (userName security.SQLUsername, err error) {
	__antithesis_instrumentation__.Notify(188743)
	userName, isAdmin, err := c.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188746)
		return userName, serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188747)
	}
	__antithesis_instrumentation__.Notify(188744)
	if !isAdmin {
		__antithesis_instrumentation__.Notify(188748)
		return userName, errRequiresAdmin
	} else {
		__antithesis_instrumentation__.Notify(188749)
	}
	__antithesis_instrumentation__.Notify(188745)
	return userName, nil
}

func (c *adminPrivilegeChecker) requireViewActivityPermission(ctx context.Context) (err error) {
	__antithesis_instrumentation__.Notify(188750)
	userName, isAdmin, err := c.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188753)
		return serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188754)
	}
	__antithesis_instrumentation__.Notify(188751)
	if !isAdmin {
		__antithesis_instrumentation__.Notify(188755)
		hasViewActivity, err := c.hasRoleOption(ctx, userName, roleoption.VIEWACTIVITY)
		if err != nil {
			__antithesis_instrumentation__.Notify(188757)
			return serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(188758)
		}
		__antithesis_instrumentation__.Notify(188756)

		if !hasViewActivity {
			__antithesis_instrumentation__.Notify(188759)
			return status.Errorf(
				codes.PermissionDenied, "this operation requires the %s role option",
				roleoption.VIEWACTIVITY)
		} else {
			__antithesis_instrumentation__.Notify(188760)
		}
	} else {
		__antithesis_instrumentation__.Notify(188761)
	}
	__antithesis_instrumentation__.Notify(188752)
	return nil
}

func (c *adminPrivilegeChecker) requireViewActivityOrViewActivityRedactedPermission(
	ctx context.Context,
) (err error) {
	__antithesis_instrumentation__.Notify(188762)
	userName, isAdmin, err := c.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188765)
		return serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188766)
	}
	__antithesis_instrumentation__.Notify(188763)
	if !isAdmin {
		__antithesis_instrumentation__.Notify(188767)
		hasViewActivity, err := c.hasRoleOption(ctx, userName, roleoption.VIEWACTIVITY)
		if err != nil {
			__antithesis_instrumentation__.Notify(188769)
			return serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(188770)
		}
		__antithesis_instrumentation__.Notify(188768)

		if !hasViewActivity {
			__antithesis_instrumentation__.Notify(188771)
			hasViewActivityRedacted, err := c.hasRoleOption(ctx, userName, roleoption.VIEWACTIVITYREDACTED)
			if err != nil {
				__antithesis_instrumentation__.Notify(188773)
				return serverError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(188774)
			}
			__antithesis_instrumentation__.Notify(188772)

			if !hasViewActivityRedacted {
				__antithesis_instrumentation__.Notify(188775)
				return status.Errorf(
					codes.PermissionDenied, "this operation requires the %s or %s role options",
					roleoption.VIEWACTIVITY, roleoption.VIEWACTIVITYREDACTED)
			} else {
				__antithesis_instrumentation__.Notify(188776)
			}
		} else {
			__antithesis_instrumentation__.Notify(188777)
		}
	} else {
		__antithesis_instrumentation__.Notify(188778)
	}
	__antithesis_instrumentation__.Notify(188764)
	return nil
}

func (c *adminPrivilegeChecker) requireViewActivityAndNoViewActivityRedactedPermission(
	ctx context.Context,
) (err error) {
	__antithesis_instrumentation__.Notify(188779)
	userName, isAdmin, err := c.getUserAndRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188782)
		return serverError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(188783)
	}
	__antithesis_instrumentation__.Notify(188780)

	if !isAdmin {
		__antithesis_instrumentation__.Notify(188784)
		hasViewActivityRedacted, err := c.hasRoleOption(ctx, userName, roleoption.VIEWACTIVITYREDACTED)
		if err != nil {
			__antithesis_instrumentation__.Notify(188787)
			return serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(188788)
		}
		__antithesis_instrumentation__.Notify(188785)
		if hasViewActivityRedacted {
			__antithesis_instrumentation__.Notify(188789)
			return status.Errorf(
				codes.PermissionDenied, "this operation requires %s role option and is not allowed for %s role option",
				roleoption.VIEWACTIVITY, roleoption.VIEWACTIVITYREDACTED)
		} else {
			__antithesis_instrumentation__.Notify(188790)
		}
		__antithesis_instrumentation__.Notify(188786)
		return c.requireViewActivityPermission(ctx)
	} else {
		__antithesis_instrumentation__.Notify(188791)
	}
	__antithesis_instrumentation__.Notify(188781)
	return nil
}

func (c *adminPrivilegeChecker) getUserAndRole(
	ctx context.Context,
) (userName security.SQLUsername, isAdmin bool, err error) {
	__antithesis_instrumentation__.Notify(188792)
	userName, err = userFromContext(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188794)
		return userName, false, err
	} else {
		__antithesis_instrumentation__.Notify(188795)
	}
	__antithesis_instrumentation__.Notify(188793)
	isAdmin, err = c.hasAdminRole(ctx, userName)
	return userName, isAdmin, err
}

func (c *adminPrivilegeChecker) hasAdminRole(
	ctx context.Context, user security.SQLUsername,
) (bool, error) {
	__antithesis_instrumentation__.Notify(188796)
	if user.IsRootUser() {
		__antithesis_instrumentation__.Notify(188802)

		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(188803)
	}
	__antithesis_instrumentation__.Notify(188797)
	row, err := c.ie.QueryRowEx(
		ctx, "check-is-admin", nil,
		sessiondata.InternalExecutorOverride{User: user},
		"SELECT crdb_internal.is_admin()")
	if err != nil {
		__antithesis_instrumentation__.Notify(188804)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(188805)
	}
	__antithesis_instrumentation__.Notify(188798)
	if row == nil {
		__antithesis_instrumentation__.Notify(188806)
		return false, errors.AssertionFailedf("hasAdminRole: expected 1 row, got 0")
	} else {
		__antithesis_instrumentation__.Notify(188807)
	}
	__antithesis_instrumentation__.Notify(188799)
	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(188808)
		return false, errors.AssertionFailedf("hasAdminRole: expected 1 column, got %d", len(row))
	} else {
		__antithesis_instrumentation__.Notify(188809)
	}
	__antithesis_instrumentation__.Notify(188800)
	dbDatum, ok := tree.AsDBool(row[0])
	if !ok {
		__antithesis_instrumentation__.Notify(188810)
		return false, errors.AssertionFailedf("hasAdminRole: expected bool, got %T", row[0])
	} else {
		__antithesis_instrumentation__.Notify(188811)
	}
	__antithesis_instrumentation__.Notify(188801)
	return bool(dbDatum), nil
}

func (c *adminPrivilegeChecker) hasRoleOption(
	ctx context.Context, user security.SQLUsername, roleOption roleoption.Option,
) (bool, error) {
	__antithesis_instrumentation__.Notify(188812)
	if user.IsRootUser() {
		__antithesis_instrumentation__.Notify(188818)

		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(188819)
	}
	__antithesis_instrumentation__.Notify(188813)
	row, err := c.ie.QueryRowEx(
		ctx, "check-role-option", nil,
		sessiondata.InternalExecutorOverride{User: user},
		"SELECT crdb_internal.has_role_option($1)", roleOption.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(188820)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(188821)
	}
	__antithesis_instrumentation__.Notify(188814)
	if row == nil {
		__antithesis_instrumentation__.Notify(188822)
		return false, errors.AssertionFailedf("hasRoleOption: expected 1 row, got 0")
	} else {
		__antithesis_instrumentation__.Notify(188823)
	}
	__antithesis_instrumentation__.Notify(188815)
	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(188824)
		return false, errors.AssertionFailedf("hasRoleOption: expected 1 column, got %d", len(row))
	} else {
		__antithesis_instrumentation__.Notify(188825)
	}
	__antithesis_instrumentation__.Notify(188816)
	dbDatum, ok := tree.AsDBool(row[0])
	if !ok {
		__antithesis_instrumentation__.Notify(188826)
		return false, errors.AssertionFailedf("hasRoleOption: expected bool, got %T", row[0])
	} else {
		__antithesis_instrumentation__.Notify(188827)
	}
	__antithesis_instrumentation__.Notify(188817)
	return bool(dbDatum), nil
}

var errRequiresAdmin = status.Error(codes.PermissionDenied, "this operation requires admin privilege")

func errRequiresRoleOption(option roleoption.Option) error {
	__antithesis_instrumentation__.Notify(188828)
	return status.Errorf(
		codes.PermissionDenied, "this operation requires %s privilege", option)
}

func (s *adminServer) ListTracingSnapshots(
	ctx context.Context, req *serverpb.ListTracingSnapshotsRequest,
) (*serverpb.ListTracingSnapshotsResponse, error) {
	__antithesis_instrumentation__.Notify(188829)
	ctx = s.server.AnnotateCtx(ctx)
	_, err := s.requireAdminUser(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188832)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188833)
	}
	__antithesis_instrumentation__.Notify(188830)

	snapshotInfo := s.server.cfg.Tracer.GetSnapshots()
	snapshots := make([]*serverpb.SnapshotInfo, len(snapshotInfo))
	for i := range snapshotInfo {
		__antithesis_instrumentation__.Notify(188834)
		si := snapshotInfo[i]
		snapshots[i] = &serverpb.SnapshotInfo{
			SnapshotID: int64(si.ID),
			CapturedAt: &si.CapturedAt,
		}
	}
	__antithesis_instrumentation__.Notify(188831)
	resp := &serverpb.ListTracingSnapshotsResponse{
		Snapshots: snapshots,
	}
	return resp, nil
}

func (s *adminServer) TakeTracingSnapshot(
	ctx context.Context, req *serverpb.TakeTracingSnapshotRequest,
) (*serverpb.TakeTracingSnapshotResponse, error) {
	__antithesis_instrumentation__.Notify(188835)
	ctx = s.server.AnnotateCtx(ctx)
	_, err := s.requireAdminUser(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188837)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188838)
	}
	__antithesis_instrumentation__.Notify(188836)

	snapshot := s.server.cfg.Tracer.SaveSnapshot()
	resp := &serverpb.TakeTracingSnapshotResponse{
		Snapshot: &serverpb.SnapshotInfo{
			SnapshotID: int64(snapshot.ID),
			CapturedAt: &snapshot.CapturedAt,
		},
	}
	return resp, nil
}

func (s *adminServer) GetTracingSnapshot(
	ctx context.Context, req *serverpb.GetTracingSnapshotRequest,
) (*serverpb.GetTracingSnapshotResponse, error) {
	__antithesis_instrumentation__.Notify(188839)
	ctx = s.server.AnnotateCtx(ctx)
	_, err := s.requireAdminUser(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188844)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188845)
	}
	__antithesis_instrumentation__.Notify(188840)

	id := tracing.SnapshotID(req.SnapshotId)
	tr := s.server.cfg.Tracer
	snapshot, err := tr.GetSnapshot(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(188846)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188847)
	}
	__antithesis_instrumentation__.Notify(188841)

	spansList := tracingui.ProcessSnapshot(snapshot, tr.GetActiveSpansRegistry())

	spans := make([]*serverpb.TracingSpan, len(spansList.Spans))
	stacks := make(map[string]string)

	for i, s := range spansList.Spans {
		__antithesis_instrumentation__.Notify(188848)
		tags := make([]*serverpb.SpanTag, len(s.Tags))
		for j, t := range s.Tags {
			__antithesis_instrumentation__.Notify(188850)
			tags[j] = &serverpb.SpanTag{
				Key:             t.Key,
				Val:             t.Val,
				Caption:         t.Caption,
				Link:            t.Link,
				Hidden:          t.Hidden,
				Highlight:       t.Highlight,
				Inherit:         t.Inherit,
				Inherited:       t.Inherited,
				PropagateUp:     t.PropagateUp,
				CopiedFromChild: t.CopiedFromChild,
			}
		}
		__antithesis_instrumentation__.Notify(188849)

		spans[i] = &serverpb.TracingSpan{
			Operation:            s.Operation,
			TraceID:              s.TraceID,
			SpanID:               s.SpanID,
			ParentSpanID:         s.ParentSpanID,
			Start:                s.Start,
			GoroutineID:          s.GoroutineID,
			ProcessedTags:        tags,
			Current:              s.Current,
			CurrentRecordingMode: s.CurrentRecordingMode.ToProto(),
		}
	}
	__antithesis_instrumentation__.Notify(188842)

	for k, v := range spansList.Stacks {
		__antithesis_instrumentation__.Notify(188851)
		stacks[strconv.FormatInt(int64(k), 10)] = v
	}
	__antithesis_instrumentation__.Notify(188843)

	snapshotResp := &serverpb.TracingSnapshot{
		SnapshotID: int64(id),
		CapturedAt: &snapshot.CapturedAt,
		Spans:      spans,
		Stacks:     stacks,
	}
	resp := &serverpb.GetTracingSnapshotResponse{
		Snapshot: snapshotResp,
	}
	return resp, nil
}

func (s *adminServer) GetTrace(
	ctx context.Context, req *serverpb.GetTraceRequest,
) (*serverpb.GetTraceResponse, error) {
	__antithesis_instrumentation__.Notify(188852)
	ctx = s.server.AnnotateCtx(ctx)
	_, err := s.requireAdminUser(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(188857)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188858)
	}
	__antithesis_instrumentation__.Notify(188853)
	var recording tracing.Recording
	var snapshotID tracing.SnapshotID

	traceID := req.TraceID
	snapID := req.SnapshotID
	if snapID != 0 {
		__antithesis_instrumentation__.Notify(188859)
		snapshotID = tracing.SnapshotID(snapID)
		snapshot, err := s.server.cfg.Tracer.GetSnapshot(snapshotID)
		if err != nil {
			__antithesis_instrumentation__.Notify(188861)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(188862)
		}
		__antithesis_instrumentation__.Notify(188860)

		for _, r := range snapshot.Traces {
			__antithesis_instrumentation__.Notify(188863)
			if r[0].TraceID == traceID {
				__antithesis_instrumentation__.Notify(188864)
				recording = r
				break
			} else {
				__antithesis_instrumentation__.Notify(188865)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(188866)
	}
	__antithesis_instrumentation__.Notify(188854)

	traceStillExists := false
	if err := s.server.cfg.Tracer.SpanRegistry().VisitRoots(func(sp tracing.RegistrySpan) error {
		__antithesis_instrumentation__.Notify(188867)
		if sp.TraceID() != traceID {
			__antithesis_instrumentation__.Notify(188870)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(188871)
		}
		__antithesis_instrumentation__.Notify(188868)
		traceStillExists = true
		if recording == nil {
			__antithesis_instrumentation__.Notify(188872)
			recording = sp.GetFullRecording(tracing.RecordingVerbose)
		} else {
			__antithesis_instrumentation__.Notify(188873)
		}
		__antithesis_instrumentation__.Notify(188869)
		return iterutil.StopIteration()
	}); err != nil {
		__antithesis_instrumentation__.Notify(188874)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(188875)
	}
	__antithesis_instrumentation__.Notify(188855)

	if recording == nil {
		__antithesis_instrumentation__.Notify(188876)
		return nil, errors.Errorf("Trace %d not found.", traceID)
	} else {
		__antithesis_instrumentation__.Notify(188877)
	}
	__antithesis_instrumentation__.Notify(188856)

	return &serverpb.GetTraceResponse{
		SnapshotID:          snapID,
		TraceID:             traceID,
		StillExists:         traceStillExists,
		SerializedRecording: recording.String(),
	}, nil
}

func (s *adminServer) SetTraceRecordingType(
	ctx context.Context, req *serverpb.SetTraceRecordingTypeRequest,
) (*serverpb.SetTraceRecordingTypeResponse, error) {
	__antithesis_instrumentation__.Notify(188878)
	if req.TraceID == 0 {
		__antithesis_instrumentation__.Notify(188881)
		return nil, errors.Errorf("missing trace id")
	} else {
		__antithesis_instrumentation__.Notify(188882)
	}
	__antithesis_instrumentation__.Notify(188879)
	_ = s.server.cfg.Tracer.VisitSpans(func(sp tracing.RegistrySpan) error {
		__antithesis_instrumentation__.Notify(188883)
		if sp.TraceID() != req.TraceID {
			__antithesis_instrumentation__.Notify(188885)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(188886)
		}
		__antithesis_instrumentation__.Notify(188884)

		sp.SetRecordingType(tracing.RecordingTypeFromProto(req.RecordingMode))
		return nil
	})
	__antithesis_instrumentation__.Notify(188880)
	return &serverpb.SetTraceRecordingTypeResponse{}, nil
}
