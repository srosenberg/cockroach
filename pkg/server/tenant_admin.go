package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tenantAdminServer struct {
	log.AmbientContext
	serverpb.UnimplementedAdminServer
	sqlServer *SQLServer
	drain     *drainServer

	status *tenantStatusServer
}

var _ serverpb.AdminServer = &tenantAdminServer{}

func (t *tenantAdminServer) RegisterService(g *grpc.Server) {
	__antithesis_instrumentation__.Notify(238389)
	serverpb.RegisterAdminServer(g, t)
}

func (t *tenantAdminServer) RegisterGateway(
	ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn,
) error {
	__antithesis_instrumentation__.Notify(238390)
	ctx = t.AnnotateCtx(ctx)
	return serverpb.RegisterAdminHandler(ctx, mux, conn)
}

var _ grpcGatewayServer = &tenantAdminServer{}

func newTenantAdminServer(
	ambientCtx log.AmbientContext,
	sqlServer *SQLServer,
	status *tenantStatusServer,
	drain *drainServer,
) *tenantAdminServer {
	__antithesis_instrumentation__.Notify(238391)
	return &tenantAdminServer{
		AmbientContext: ambientCtx,
		sqlServer:      sqlServer,
		drain:          drain,
		status:         status,
	}
}

func (t *tenantAdminServer) Health(
	ctx context.Context, req *serverpb.HealthRequest,
) (*serverpb.HealthResponse, error) {
	__antithesis_instrumentation__.Notify(238392)
	telemetry.Inc(telemetryHealthCheck)

	resp := &serverpb.HealthResponse{}

	if !req.Ready {
		__antithesis_instrumentation__.Notify(238395)
		return resp, nil
	} else {
		__antithesis_instrumentation__.Notify(238396)
	}
	__antithesis_instrumentation__.Notify(238393)

	if !t.sqlServer.isReady.Get() {
		__antithesis_instrumentation__.Notify(238397)
		return nil, status.Errorf(codes.Unavailable, "node is not accepting SQL clients")
	} else {
		__antithesis_instrumentation__.Notify(238398)
	}
	__antithesis_instrumentation__.Notify(238394)

	return resp, nil
}

func (t *tenantAdminServer) dialPod(
	ctx context.Context, instanceID base.SQLInstanceID, addr string,
) (serverpb.AdminClient, error) {
	__antithesis_instrumentation__.Notify(238399)
	conn, err := t.sqlServer.execCfg.RPCContext.GRPCDialPod(addr, instanceID, rpc.DefaultClass).Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(238401)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238402)
	}
	__antithesis_instrumentation__.Notify(238400)

	return serverpb.NewAdminClient(conn), nil
}

func (t *tenantAdminServer) Drain(
	req *serverpb.DrainRequest, stream serverpb.Admin_DrainServer,
) error {
	__antithesis_instrumentation__.Notify(238403)
	ctx := stream.Context()
	ctx = t.AnnotateCtx(ctx)

	parsedInstanceID, local, err := t.status.parseInstanceID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(238406)
		return status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(238407)
	}
	__antithesis_instrumentation__.Notify(238404)
	if !local {
		__antithesis_instrumentation__.Notify(238408)
		instance, err := t.sqlServer.sqlInstanceProvider.GetInstance(ctx, parsedInstanceID)
		if err != nil {
			__antithesis_instrumentation__.Notify(238411)
			return err
		} else {
			__antithesis_instrumentation__.Notify(238412)
		}
		__antithesis_instrumentation__.Notify(238409)

		client, err := t.dialPod(ctx, parsedInstanceID, instance.InstanceAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(238413)
			return serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(238414)
		}
		__antithesis_instrumentation__.Notify(238410)
		return delegateDrain(ctx, req, client, stream)
	} else {
		__antithesis_instrumentation__.Notify(238415)
	}
	__antithesis_instrumentation__.Notify(238405)

	return t.drain.handleDrain(ctx, req, stream)
}
