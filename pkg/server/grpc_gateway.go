package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net"

	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/ts"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

type grpcGatewayServer interface {
	RegisterService(g *grpc.Server)
	RegisterGateway(
		ctx context.Context,
		mux *gwruntime.ServeMux,
		conn *grpc.ClientConn,
	) error
}

var _ grpcGatewayServer = (*adminServer)(nil)
var _ grpcGatewayServer = (*statusServer)(nil)
var _ grpcGatewayServer = (*authenticationServer)(nil)
var _ grpcGatewayServer = (*ts.Server)(nil)

func configureGRPCGateway(
	ctx, workersCtx context.Context,
	ambientCtx log.AmbientContext,
	rpcContext *rpc.Context,
	stopper *stop.Stopper,
	grpcSrv *grpcServer,
	GRPCAddr string,
) (*gwruntime.ServeMux, context.Context, *grpc.ClientConn, error) {
	__antithesis_instrumentation__.Notify(193443)
	jsonpb := &protoutil.JSONPb{
		EnumsAsInts:  true,
		EmitDefaults: true,
		Indent:       "  ",
	}
	protopb := new(protoutil.ProtoPb)
	gwMux := gwruntime.NewServeMux(
		gwruntime.WithMarshalerOption(gwruntime.MIMEWildcard, jsonpb),
		gwruntime.WithMarshalerOption(httputil.JSONContentType, jsonpb),
		gwruntime.WithMarshalerOption(httputil.AltJSONContentType, jsonpb),
		gwruntime.WithMarshalerOption(httputil.ProtoContentType, protopb),
		gwruntime.WithMarshalerOption(httputil.AltProtoContentType, protopb),
		gwruntime.WithOutgoingHeaderMatcher(authenticationHeaderMatcher),
		gwruntime.WithMetadata(forwardAuthenticationMetadata),
	)
	gwCtx, gwCancel := context.WithCancel(ambientCtx.AnnotateCtx(context.Background()))
	stopper.AddCloser(stop.CloserFn(gwCancel))

	loopback := newLoopbackListener(workersCtx, stopper)

	waitQuiesce := func(context.Context) {
		__antithesis_instrumentation__.Notify(193451)
		<-stopper.ShouldQuiesce()
		_ = loopback.Close()
	}
	__antithesis_instrumentation__.Notify(193444)
	if err := stopper.RunAsyncTask(workersCtx, "gw-quiesce", waitQuiesce); err != nil {
		__antithesis_instrumentation__.Notify(193452)
		waitQuiesce(workersCtx)
	} else {
		__antithesis_instrumentation__.Notify(193453)
	}
	__antithesis_instrumentation__.Notify(193445)

	_ = stopper.RunAsyncTask(workersCtx, "serve-loopback", func(context.Context) {
		__antithesis_instrumentation__.Notify(193454)
		netutil.FatalIfUnexpected(grpcSrv.Serve(loopback))
	})
	__antithesis_instrumentation__.Notify(193446)

	dialOpts, err := rpcContext.GRPCDialOptions()
	if err != nil {
		__antithesis_instrumentation__.Notify(193455)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(193456)
	}
	__antithesis_instrumentation__.Notify(193447)

	callCountInterceptor := func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		__antithesis_instrumentation__.Notify(193457)
		telemetry.Inc(getServerEndpointCounter(method))
		return invoker(ctx, method, req, reply, cc, opts...)
	}
	__antithesis_instrumentation__.Notify(193448)
	conn, err := grpc.DialContext(ctx, GRPCAddr, append(append(
		dialOpts,
		grpc.WithUnaryInterceptor(callCountInterceptor)),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			__antithesis_instrumentation__.Notify(193458)
			return loopback.Connect(ctx)
		}),
	)...)
	__antithesis_instrumentation__.Notify(193449)
	if err != nil {
		__antithesis_instrumentation__.Notify(193459)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(193460)
	}
	{
		__antithesis_instrumentation__.Notify(193461)
		waitQuiesce := func(workersCtx context.Context) {
			__antithesis_instrumentation__.Notify(193463)
			<-stopper.ShouldQuiesce()

			err := conn.Close()
			if err != nil {
				__antithesis_instrumentation__.Notify(193464)
				log.Ops.Fatalf(workersCtx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(193465)
			}
		}
		__antithesis_instrumentation__.Notify(193462)
		if err := stopper.RunAsyncTask(workersCtx, "wait-quiesce", waitQuiesce); err != nil {
			__antithesis_instrumentation__.Notify(193466)
			waitQuiesce(workersCtx)
		} else {
			__antithesis_instrumentation__.Notify(193467)
		}
	}
	__antithesis_instrumentation__.Notify(193450)
	return gwMux, gwCtx, conn, nil
}

func getServerEndpointCounter(method string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(193468)
	const counterPrefix = "http.grpc-gateway"
	return telemetry.GetCounter(fmt.Sprintf("%s.%s", counterPrefix, method))
}
