package nodedialer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net"
	"time"
	"unsafe"

	circuit "github.com/cockroachdb/circuitbreaker"
	"github.com/cockroachdb/cockroach/pkg/kv/kvbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
)

const logPerNodeFailInterval = time.Minute

type wrappedBreaker struct {
	*circuit.Breaker
	log.EveryN
}

type AddressResolver func(roachpb.NodeID) (net.Addr, error)

type Dialer struct {
	rpcContext   *rpc.Context
	resolver     AddressResolver
	testingKnobs DialerTestingKnobs

	breakers [rpc.NumConnectionClasses]syncutil.IntMap
}

type DialerOpt struct {
	TestingKnobs DialerTestingKnobs
}

type DialerTestingKnobs struct {
	TestingNoLocalClientOptimization bool
}

func (DialerTestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(185379) }

func New(rpcContext *rpc.Context, resolver AddressResolver) *Dialer {
	__antithesis_instrumentation__.Notify(185380)
	return &Dialer{
		rpcContext: rpcContext,
		resolver:   resolver,
	}
}

func NewWithOpt(rpcContext *rpc.Context, resolver AddressResolver, opt DialerOpt) *Dialer {
	__antithesis_instrumentation__.Notify(185381)
	d := New(rpcContext, resolver)
	d.testingKnobs = opt.TestingKnobs
	return d
}

func (n *Dialer) Stopper() *stop.Stopper {
	__antithesis_instrumentation__.Notify(185382)
	return n.rpcContext.Stopper
}

var _ = (*Dialer).Stopper

func (n *Dialer) Dial(
	ctx context.Context, nodeID roachpb.NodeID, class rpc.ConnectionClass,
) (_ *grpc.ClientConn, err error) {
	__antithesis_instrumentation__.Notify(185383)
	if n == nil || func() bool {
		__antithesis_instrumentation__.Notify(185387)
		return n.resolver == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(185388)
		return nil, errors.New("no node dialer configured")
	} else {
		__antithesis_instrumentation__.Notify(185389)
	}
	__antithesis_instrumentation__.Notify(185384)

	if ctxErr := ctx.Err(); ctxErr != nil {
		__antithesis_instrumentation__.Notify(185390)
		return nil, ctxErr
	} else {
		__antithesis_instrumentation__.Notify(185391)
	}
	__antithesis_instrumentation__.Notify(185385)
	breaker := n.getBreaker(nodeID, class)
	addr, err := n.resolver(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(185392)
		err = errors.Wrapf(err, "failed to resolve n%d", nodeID)
		breaker.Fail(err)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(185393)
	}
	__antithesis_instrumentation__.Notify(185386)
	return n.dial(ctx, nodeID, addr, breaker, true, class)
}

func (n *Dialer) DialNoBreaker(
	ctx context.Context, nodeID roachpb.NodeID, class rpc.ConnectionClass,
) (_ *grpc.ClientConn, err error) {
	__antithesis_instrumentation__.Notify(185394)
	if n == nil || func() bool {
		__antithesis_instrumentation__.Notify(185397)
		return n.resolver == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(185398)
		return nil, errors.New("no node dialer configured")
	} else {
		__antithesis_instrumentation__.Notify(185399)
	}
	__antithesis_instrumentation__.Notify(185395)
	addr, err := n.resolver(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(185400)
		if ctx.Err() == nil {
			__antithesis_instrumentation__.Notify(185402)
			n.getBreaker(nodeID, class).Fail(err)
		} else {
			__antithesis_instrumentation__.Notify(185403)
		}
		__antithesis_instrumentation__.Notify(185401)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(185404)
	}
	__antithesis_instrumentation__.Notify(185396)
	return n.dial(ctx, nodeID, addr, n.getBreaker(nodeID, class), false, class)
}

func (n *Dialer) DialInternalClient(
	ctx context.Context, nodeID roachpb.NodeID, class rpc.ConnectionClass,
) (context.Context, roachpb.InternalClient, error) {
	__antithesis_instrumentation__.Notify(185405)
	if n == nil || func() bool {
		__antithesis_instrumentation__.Notify(185409)
		return n.resolver == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(185410)
		return nil, nil, errors.New("no node dialer configured")
	} else {
		__antithesis_instrumentation__.Notify(185411)
	}
	__antithesis_instrumentation__.Notify(185406)
	addr, err := n.resolver(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(185412)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185413)
	}

	{
		__antithesis_instrumentation__.Notify(185414)

		localClient := n.rpcContext.GetLocalInternalClientForAddr(addr.String(), nodeID)
		if localClient != nil && func() bool {
			__antithesis_instrumentation__.Notify(185415)
			return !n.testingKnobs.TestingNoLocalClientOptimization == true
		}() == true {
			__antithesis_instrumentation__.Notify(185416)
			log.VEvent(ctx, 2, kvbase.RoutingRequestLocallyMsg)

			localCtx := grpcutil.NewLocalRequestContext(ctx)

			return localCtx, localClient, nil
		} else {
			__antithesis_instrumentation__.Notify(185417)
		}
	}
	__antithesis_instrumentation__.Notify(185407)
	log.VEventf(ctx, 2, "sending request to %s", addr)
	conn, err := n.dial(ctx, nodeID, addr, n.getBreaker(nodeID, class), true, class)
	if err != nil {
		__antithesis_instrumentation__.Notify(185418)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(185419)
	}
	__antithesis_instrumentation__.Notify(185408)
	return ctx, TracingInternalClient{InternalClient: roachpb.NewInternalClient(conn)}, err
}

func (n *Dialer) dial(
	ctx context.Context,
	nodeID roachpb.NodeID,
	addr net.Addr,
	breaker *wrappedBreaker,
	checkBreaker bool,
	class rpc.ConnectionClass,
) (_ *grpc.ClientConn, err error) {
	__antithesis_instrumentation__.Notify(185420)

	if ctxErr := ctx.Err(); ctxErr != nil {
		__antithesis_instrumentation__.Notify(185427)
		return nil, ctxErr
	} else {
		__antithesis_instrumentation__.Notify(185428)
	}
	__antithesis_instrumentation__.Notify(185421)
	if checkBreaker && func() bool {
		__antithesis_instrumentation__.Notify(185429)
		return !breaker.Ready() == true
	}() == true {
		__antithesis_instrumentation__.Notify(185430)
		err = errors.Wrapf(circuit.ErrBreakerOpen, "unable to dial n%d", nodeID)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(185431)
	}
	__antithesis_instrumentation__.Notify(185422)
	defer func() {
		__antithesis_instrumentation__.Notify(185432)

		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(185433)
			return ctx.Err() == nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(185434)
			return breaker != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(185435)
			return breaker.ShouldLog() == true
		}() == true {
			__antithesis_instrumentation__.Notify(185436)
			log.Health.Warningf(ctx, "unable to connect to n%d: %s", nodeID, err)
		} else {
			__antithesis_instrumentation__.Notify(185437)
		}
	}()
	__antithesis_instrumentation__.Notify(185423)
	conn, err := n.rpcContext.GRPCDialNode(addr.String(), nodeID, class).Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(185438)

		if ctxErr := ctx.Err(); ctxErr != nil {
			__antithesis_instrumentation__.Notify(185441)
			return nil, ctxErr
		} else {
			__antithesis_instrumentation__.Notify(185442)
		}
		__antithesis_instrumentation__.Notify(185439)
		err = errors.Wrapf(err, "failed to connect to n%d at %v", nodeID, addr)
		if breaker != nil {
			__antithesis_instrumentation__.Notify(185443)
			breaker.Fail(err)
		} else {
			__antithesis_instrumentation__.Notify(185444)
		}
		__antithesis_instrumentation__.Notify(185440)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(185445)
	}
	__antithesis_instrumentation__.Notify(185424)

	if err := grpcutil.ConnectionReady(conn); err != nil {
		__antithesis_instrumentation__.Notify(185446)
		err = errors.Wrapf(err, "failed to check for ready connection to n%d at %v", nodeID, addr)
		if breaker != nil {
			__antithesis_instrumentation__.Notify(185448)
			breaker.Fail(err)
		} else {
			__antithesis_instrumentation__.Notify(185449)
		}
		__antithesis_instrumentation__.Notify(185447)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(185450)
	}
	__antithesis_instrumentation__.Notify(185425)

	if breaker != nil {
		__antithesis_instrumentation__.Notify(185451)
		breaker.Success()
	} else {
		__antithesis_instrumentation__.Notify(185452)
	}
	__antithesis_instrumentation__.Notify(185426)
	return conn, nil
}

func (n *Dialer) ConnHealth(nodeID roachpb.NodeID, class rpc.ConnectionClass) error {
	__antithesis_instrumentation__.Notify(185453)
	if n == nil || func() bool {
		__antithesis_instrumentation__.Notify(185457)
		return n.resolver == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(185458)
		return errors.New("no node dialer configured")
	} else {
		__antithesis_instrumentation__.Notify(185459)
	}
	__antithesis_instrumentation__.Notify(185454)

	if n.getBreaker(nodeID, class).Tripped() {
		__antithesis_instrumentation__.Notify(185460)
		return circuit.ErrBreakerOpen
	} else {
		__antithesis_instrumentation__.Notify(185461)
	}
	__antithesis_instrumentation__.Notify(185455)
	addr, err := n.resolver(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(185462)
		return err
	} else {
		__antithesis_instrumentation__.Notify(185463)
	}
	__antithesis_instrumentation__.Notify(185456)
	return n.rpcContext.ConnHealth(addr.String(), nodeID, class)
}

func (n *Dialer) ConnHealthTryDial(nodeID roachpb.NodeID, class rpc.ConnectionClass) error {
	__antithesis_instrumentation__.Notify(185464)
	err := n.ConnHealth(nodeID, class)
	if err == nil || func() bool {
		__antithesis_instrumentation__.Notify(185467)
		return !n.getBreaker(nodeID, class).Ready() == true
	}() == true {
		__antithesis_instrumentation__.Notify(185468)
		return err
	} else {
		__antithesis_instrumentation__.Notify(185469)
	}
	__antithesis_instrumentation__.Notify(185465)
	addr, err := n.resolver(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(185470)
		return err
	} else {
		__antithesis_instrumentation__.Notify(185471)
	}
	__antithesis_instrumentation__.Notify(185466)
	return n.rpcContext.GRPCDialNode(addr.String(), nodeID, class).Health()
}

func (n *Dialer) GetCircuitBreaker(
	nodeID roachpb.NodeID, class rpc.ConnectionClass,
) *circuit.Breaker {
	__antithesis_instrumentation__.Notify(185472)
	return n.getBreaker(nodeID, class).Breaker
}

func (n *Dialer) getBreaker(nodeID roachpb.NodeID, class rpc.ConnectionClass) *wrappedBreaker {
	__antithesis_instrumentation__.Notify(185473)
	breakers := &n.breakers[class]
	value, ok := breakers.Load(int64(nodeID))
	if !ok {
		__antithesis_instrumentation__.Notify(185475)
		name := fmt.Sprintf("rpc %v [n%d]", n.rpcContext.Config.Addr, nodeID)
		breaker := &wrappedBreaker{Breaker: n.rpcContext.NewBreaker(name), EveryN: log.Every(logPerNodeFailInterval)}
		value, _ = breakers.LoadOrStore(int64(nodeID), unsafe.Pointer(breaker))
	} else {
		__antithesis_instrumentation__.Notify(185476)
	}
	__antithesis_instrumentation__.Notify(185474)
	return (*wrappedBreaker)(value)
}

func (n *Dialer) Latency(nodeID roachpb.NodeID) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(185477)
	if n == nil || func() bool {
		__antithesis_instrumentation__.Notify(185481)
		return n.resolver == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(185482)
		return 0, errors.New("no node dialer configured")
	} else {
		__antithesis_instrumentation__.Notify(185483)
	}
	__antithesis_instrumentation__.Notify(185478)
	addr, err := n.resolver(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(185484)

		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(185485)
	}
	__antithesis_instrumentation__.Notify(185479)
	latency, ok := n.rpcContext.RemoteClocks.Latency(addr.String())
	if !ok {
		__antithesis_instrumentation__.Notify(185486)
		latency = 0
	} else {
		__antithesis_instrumentation__.Notify(185487)
	}
	__antithesis_instrumentation__.Notify(185480)
	return latency, nil
}

type TracingInternalClient struct {
	roachpb.InternalClient
}

func (tic TracingInternalClient) Batch(
	ctx context.Context, req *roachpb.BatchRequest, opts ...grpc.CallOption,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(185488)
	sp := tracing.SpanFromContext(ctx)
	if sp != nil && func() bool {
		__antithesis_instrumentation__.Notify(185490)
		return !sp.IsNoop() == true
	}() == true {
		__antithesis_instrumentation__.Notify(185491)
		req.TraceInfo = sp.Meta().ToProto()
	} else {
		__antithesis_instrumentation__.Notify(185492)
	}
	__antithesis_instrumentation__.Notify(185489)
	return tic.InternalClient.Batch(ctx, req, opts...)
}
