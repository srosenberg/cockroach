package execinfrapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/logtags"
	"google.golang.org/grpc"
)

func newInsecureRPCContext(ctx context.Context, stopper *stop.Stopper) *rpc.Context {
	__antithesis_instrumentation__.Notify(490806)
	nc := &base.NodeIDContainer{}
	ctx = logtags.AddTag(ctx, "n", nc)
	return rpc.NewContext(ctx,
		rpc.ContextOptions{
			TenantID: roachpb.SystemTenantID,
			NodeID:   nc,
			Config:   &base.Config{Insecure: true},
			Clock:    hlc.NewClock(hlc.UnixNano, time.Nanosecond),
			Stopper:  stopper,
			Settings: cluster.MakeTestingClusterSettings(),
		})
}

func StartMockDistSQLServer(
	ctx context.Context, clock *hlc.Clock, stopper *stop.Stopper, sqlInstanceID base.SQLInstanceID,
) (uuid.UUID, *MockDistSQLServer, net.Addr, error) {
	__antithesis_instrumentation__.Notify(490807)
	rpcContext := newInsecureRPCContext(ctx, stopper)
	rpcContext.NodeID.Set(context.TODO(), roachpb.NodeID(sqlInstanceID))
	server := rpc.NewServer(rpcContext)
	mock := newMockDistSQLServer()
	RegisterDistSQLServer(server, mock)
	ln, err := netutil.ListenAndServeGRPC(stopper, server, util.IsolatedTestAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(490809)
		return uuid.Nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(490810)
	}
	__antithesis_instrumentation__.Notify(490808)
	return rpcContext.StorageClusterID.Get(), mock, ln.Addr(), nil
}

type MockDistSQLServer struct {
	InboundStreams chan InboundStreamNotification
}

type InboundStreamNotification struct {
	Stream DistSQL_FlowStreamServer
	Donec  chan<- error
}

var _ DistSQLServer = &MockDistSQLServer{}

func newMockDistSQLServer() *MockDistSQLServer {
	__antithesis_instrumentation__.Notify(490811)
	return &MockDistSQLServer{
		InboundStreams: make(chan InboundStreamNotification),
	}
}

func (ds *MockDistSQLServer) SetupFlow(
	_ context.Context, req *SetupFlowRequest,
) (*SimpleResponse, error) {
	__antithesis_instrumentation__.Notify(490812)
	return nil, nil
}

func (ds *MockDistSQLServer) CancelDeadFlows(
	_ context.Context, req *CancelDeadFlowsRequest,
) (*SimpleResponse, error) {
	__antithesis_instrumentation__.Notify(490813)
	return nil, nil
}

func (ds *MockDistSQLServer) FlowStream(stream DistSQL_FlowStreamServer) error {
	__antithesis_instrumentation__.Notify(490814)
	donec := make(chan error)
	ds.InboundStreams <- InboundStreamNotification{Stream: stream, Donec: donec}
	return <-donec
}

type MockDialer struct {
	Addr net.Addr
	mu   struct {
		syncutil.Mutex
		conn *grpc.ClientConn
	}
}

func (d *MockDialer) DialNoBreaker(
	context.Context, roachpb.NodeID, rpc.ConnectionClass,
) (*grpc.ClientConn, error) {
	__antithesis_instrumentation__.Notify(490815)
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.mu.conn != nil {
		__antithesis_instrumentation__.Notify(490817)
		return d.mu.conn, nil
	} else {
		__antithesis_instrumentation__.Notify(490818)
	}
	__antithesis_instrumentation__.Notify(490816)
	var err error

	d.mu.conn, err = grpc.Dial(d.Addr.String(), grpc.WithInsecure(), grpc.WithBlock())
	return d.mu.conn, err
}

func (d *MockDialer) Close() {
	__antithesis_instrumentation__.Notify(490819)
	err := d.mu.conn.Close()
	if err != nil {
		__antithesis_instrumentation__.Notify(490820)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(490821)
	}
}
