package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	circuit "github.com/cockroachdb/circuitbreaker"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/growstack"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/syncmap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
)

func init() {

	grpc.EnableTracing = false
}

const (
	maximumPingDurationMult = 2
)

const (
	defaultWindowSize = 65535
)

func getWindowSize(name string, c ConnectionClass, defaultSize int) int32 {
	__antithesis_instrumentation__.Notify(184295)
	const maxWindowSize = defaultWindowSize * 32
	s := envutil.EnvOrDefaultInt(name, defaultSize)
	if s > maxWindowSize {
		__antithesis_instrumentation__.Notify(184298)
		log.Warningf(context.Background(), "%s value too large; trimmed to %d", name, maxWindowSize)
		s = maxWindowSize
	} else {
		__antithesis_instrumentation__.Notify(184299)
	}
	__antithesis_instrumentation__.Notify(184296)
	if s <= defaultWindowSize {
		__antithesis_instrumentation__.Notify(184300)
		log.Warningf(context.Background(),
			"%s RPC will use dynamic window sizes due to %s value lower than %d", c, name, defaultSize)
	} else {
		__antithesis_instrumentation__.Notify(184301)
	}
	__antithesis_instrumentation__.Notify(184297)
	return int32(s)
}

var (
	initialWindowSize = getWindowSize(
		"COCKROACH_RPC_INITIAL_WINDOW_SIZE", DefaultClass, defaultWindowSize*32)
	initialConnWindowSize = initialWindowSize * 16

	rangefeedInitialWindowSize = getWindowSize(
		"COCKROACH_RANGEFEED_RPC_INITIAL_WINDOW_SIZE", RangefeedClass, 2*defaultWindowSize)
)

const minConnectionTimeout = 20 * time.Second

var errDialRejected = grpcstatus.Error(codes.PermissionDenied, "refusing to dial; node is quiescing")

var sourceAddr = func() net.Addr {
	__antithesis_instrumentation__.Notify(184302)
	const envKey = "COCKROACH_SOURCE_IP_ADDRESS"
	if sourceAddr, ok := envutil.EnvString(envKey, 0); ok {
		__antithesis_instrumentation__.Notify(184304)
		sourceIP := net.ParseIP(sourceAddr)
		if sourceIP == nil {
			__antithesis_instrumentation__.Notify(184306)
			panic(fmt.Sprintf("unable to parse %s '%s' as IP address", envKey, sourceAddr))
		} else {
			__antithesis_instrumentation__.Notify(184307)
		}
		__antithesis_instrumentation__.Notify(184305)
		return &net.TCPAddr{
			IP: sourceIP,
		}
	} else {
		__antithesis_instrumentation__.Notify(184308)
	}
	__antithesis_instrumentation__.Notify(184303)
	return nil
}()

var enableRPCCompression = envutil.EnvOrDefaultBool("COCKROACH_ENABLE_RPC_COMPRESSION", true)

type serverOpts struct {
	interceptor func(fullMethod string) error
}

type ServerOption func(*serverOpts)

func WithInterceptor(f func(fullMethod string) error) ServerOption {
	__antithesis_instrumentation__.Notify(184309)
	return func(opts *serverOpts) {
		__antithesis_instrumentation__.Notify(184310)
		if opts.interceptor == nil {
			__antithesis_instrumentation__.Notify(184311)
			opts.interceptor = f
		} else {
			__antithesis_instrumentation__.Notify(184312)
			f := opts.interceptor
			opts.interceptor = func(fullMethod string) error {
				__antithesis_instrumentation__.Notify(184313)
				if err := f(fullMethod); err != nil {
					__antithesis_instrumentation__.Notify(184315)
					return err
				} else {
					__antithesis_instrumentation__.Notify(184316)
				}
				__antithesis_instrumentation__.Notify(184314)
				return f(fullMethod)
			}
		}
	}
}

func NewServer(rpcCtx *Context, opts ...ServerOption) *grpc.Server {
	__antithesis_instrumentation__.Notify(184317)
	var o serverOpts
	for _, f := range opts {
		__antithesis_instrumentation__.Notify(184325)
		f(&o)
	}
	__antithesis_instrumentation__.Notify(184318)
	grpcOpts := []grpc.ServerOption{

		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32),

		grpc.InitialWindowSize(initialWindowSize),
		grpc.InitialConnWindowSize(initialConnWindowSize),

		grpc.MaxConcurrentStreams(math.MaxInt32),
		grpc.KeepaliveParams(serverKeepalive),
		grpc.KeepaliveEnforcementPolicy(serverEnforcement),

		grpc.StatsHandler(&rpcCtx.stats),
	}
	if !rpcCtx.Config.Insecure {
		__antithesis_instrumentation__.Notify(184326)
		tlsConfig, err := rpcCtx.GetServerTLSConfig()
		if err != nil {
			__antithesis_instrumentation__.Notify(184328)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(184329)
		}
		__antithesis_instrumentation__.Notify(184327)
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	} else {
		__antithesis_instrumentation__.Notify(184330)
	}
	__antithesis_instrumentation__.Notify(184319)

	var unaryInterceptor []grpc.UnaryServerInterceptor
	var streamInterceptor []grpc.StreamServerInterceptor
	unaryInterceptor = append(unaryInterceptor, func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		__antithesis_instrumentation__.Notify(184331)
		var resp interface{}
		if err := rpcCtx.Stopper.RunTaskWithErr(ctx, info.FullMethod, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(184333)
			var err error
			resp, err = handler(ctx, req)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(184334)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(184335)
		}
		__antithesis_instrumentation__.Notify(184332)
		return resp, nil
	})
	__antithesis_instrumentation__.Notify(184320)
	streamInterceptor = append(streamInterceptor, func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		__antithesis_instrumentation__.Notify(184336)
		return rpcCtx.Stopper.RunTaskWithErr(ss.Context(), info.FullMethod, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(184337)
			return handler(srv, ss)
		})
	})
	__antithesis_instrumentation__.Notify(184321)

	if !rpcCtx.Config.Insecure {
		__antithesis_instrumentation__.Notify(184338)
		a := kvAuth{
			tenant: tenantAuthorizer{
				tenantID: rpcCtx.tenID,
			},
		}

		unaryInterceptor = append(unaryInterceptor, a.AuthUnary())
		streamInterceptor = append(streamInterceptor, a.AuthStream())
	} else {
		__antithesis_instrumentation__.Notify(184339)
	}
	__antithesis_instrumentation__.Notify(184322)

	if o.interceptor != nil {
		__antithesis_instrumentation__.Notify(184340)
		unaryInterceptor = append(unaryInterceptor, func(
			ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
		) (interface{}, error) {
			__antithesis_instrumentation__.Notify(184342)
			if err := o.interceptor(info.FullMethod); err != nil {
				__antithesis_instrumentation__.Notify(184344)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(184345)
			}
			__antithesis_instrumentation__.Notify(184343)
			return handler(ctx, req)
		})
		__antithesis_instrumentation__.Notify(184341)

		streamInterceptor = append(streamInterceptor, func(
			srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
		) error {
			__antithesis_instrumentation__.Notify(184346)
			if err := o.interceptor(info.FullMethod); err != nil {
				__antithesis_instrumentation__.Notify(184348)
				return err
			} else {
				__antithesis_instrumentation__.Notify(184349)
			}
			__antithesis_instrumentation__.Notify(184347)
			return handler(srv, stream)
		})
	} else {
		__antithesis_instrumentation__.Notify(184350)
	}
	__antithesis_instrumentation__.Notify(184323)

	if tracer := rpcCtx.Stopper.Tracer(); tracer != nil {
		__antithesis_instrumentation__.Notify(184351)
		unaryInterceptor = append(unaryInterceptor, tracing.ServerInterceptor(tracer))
		streamInterceptor = append(streamInterceptor, tracing.StreamServerInterceptor(tracer))
	} else {
		__antithesis_instrumentation__.Notify(184352)
	}
	__antithesis_instrumentation__.Notify(184324)

	grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(unaryInterceptor...))
	grpcOpts = append(grpcOpts, grpc.ChainStreamInterceptor(streamInterceptor...))

	s := grpc.NewServer(grpcOpts...)
	RegisterHeartbeatServer(s, rpcCtx.NewHeartbeatService())
	return s
}

type heartbeatResult struct {
	everSucceeded bool
	err           error
}

func (hr heartbeatResult) state() (s heartbeatState) {
	__antithesis_instrumentation__.Notify(184353)
	switch {
	case !hr.everSucceeded && func() bool {
		__antithesis_instrumentation__.Notify(184359)
		return hr.err != nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(184355)
		s = heartbeatInitializing
	case hr.everSucceeded && func() bool {
		__antithesis_instrumentation__.Notify(184360)
		return hr.err == nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(184356)
		s = heartbeatNominal
	case hr.everSucceeded && func() bool {
		__antithesis_instrumentation__.Notify(184361)
		return hr.err != nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(184357)
		s = heartbeatFailed
	default:
		__antithesis_instrumentation__.Notify(184358)
	}
	__antithesis_instrumentation__.Notify(184354)
	return s
}

type Connection struct {
	grpcConn             *grpc.ClientConn
	dialErr              error
	heartbeatResult      atomic.Value
	initialHeartbeatDone chan struct{}
	stopper              *stop.Stopper

	remoteNodeID roachpb.NodeID

	initOnce sync.Once
}

func newConnectionToNodeID(stopper *stop.Stopper, remoteNodeID roachpb.NodeID) *Connection {
	__antithesis_instrumentation__.Notify(184362)
	c := &Connection{
		initialHeartbeatDone: make(chan struct{}),
		stopper:              stopper,
		remoteNodeID:         remoteNodeID,
	}
	c.heartbeatResult.Store(heartbeatResult{err: ErrNotHeartbeated})
	return c
}

func (c *Connection) Connect(ctx context.Context) (*grpc.ClientConn, error) {
	__antithesis_instrumentation__.Notify(184363)
	if c.dialErr != nil {
		__antithesis_instrumentation__.Notify(184367)
		return nil, c.dialErr
	} else {
		__antithesis_instrumentation__.Notify(184368)
	}
	__antithesis_instrumentation__.Notify(184364)

	select {
	case <-c.initialHeartbeatDone:
		__antithesis_instrumentation__.Notify(184369)
	case <-c.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(184370)
		return nil, errors.Errorf("stopped")
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(184371)
		return nil, ctx.Err()
	}
	__antithesis_instrumentation__.Notify(184365)

	h := c.heartbeatResult.Load().(heartbeatResult)
	if !h.everSucceeded {
		__antithesis_instrumentation__.Notify(184372)

		return nil, netutil.NewInitialHeartBeatFailedError(h.err)
	} else {
		__antithesis_instrumentation__.Notify(184373)
	}
	__antithesis_instrumentation__.Notify(184366)
	return c.grpcConn, nil
}

func (c *Connection) Health() error {
	__antithesis_instrumentation__.Notify(184374)
	return c.heartbeatResult.Load().(heartbeatResult).err
}

type Context struct {
	ContextOptions
	SecurityContext

	breakerClock breakerClock
	RemoteClocks *RemoteClockMonitor
	MasterCtx    context.Context

	heartbeatTimeout time.Duration
	HeartbeatCB      func()

	rpcCompression bool

	localInternalClient roachpb.InternalClient

	conns syncmap.Map

	stats StatsHandler

	metrics Metrics

	BreakerFactory  func() *circuit.Breaker
	testingDialOpts []grpc.DialOption

	TestingAllowNamedRPCToAnonymousServer bool
}

type connKey struct {
	targetAddr string

	nodeID roachpb.NodeID
	class  ConnectionClass
}

var _ redact.SafeFormatter = connKey{}

func (c connKey) SafeFormat(p redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(184375)
	p.Printf("{n%d: %s (%v)}", c.nodeID, c.targetAddr, c.class)
}

type ContextOptions struct {
	TenantID roachpb.TenantID
	Config   *base.Config
	Clock    *hlc.Clock
	Stopper  *stop.Stopper
	Settings *cluster.Settings

	OnIncomingPing func(context.Context, *PingRequest) error

	OnOutgoingPing func(context.Context, *PingRequest) error
	Knobs          ContextTestingKnobs

	NodeID *base.NodeIDContainer

	StorageClusterID *base.ClusterIDContainer

	LogicalClusterID *base.ClusterIDContainer

	ClientOnly bool
}

func (c ContextOptions) validate() error {
	__antithesis_instrumentation__.Notify(184376)
	if c.TenantID == (roachpb.TenantID{}) {
		__antithesis_instrumentation__.Notify(184382)
		return errors.New("must specify TenantID")
	} else {
		__antithesis_instrumentation__.Notify(184383)
	}
	__antithesis_instrumentation__.Notify(184377)
	if c.Config == nil {
		__antithesis_instrumentation__.Notify(184384)
		return errors.New("Config must be set")
	} else {
		__antithesis_instrumentation__.Notify(184385)
	}
	__antithesis_instrumentation__.Notify(184378)
	if c.Clock == nil {
		__antithesis_instrumentation__.Notify(184386)
		return errors.New("Clock must be set")
	} else {
		__antithesis_instrumentation__.Notify(184387)
	}
	__antithesis_instrumentation__.Notify(184379)
	if c.Stopper == nil {
		__antithesis_instrumentation__.Notify(184388)
		return errors.New("Stopper must be set")
	} else {
		__antithesis_instrumentation__.Notify(184389)
	}
	__antithesis_instrumentation__.Notify(184380)
	if c.Settings == nil {
		__antithesis_instrumentation__.Notify(184390)
		return errors.New("Settings must be set")
	} else {
		__antithesis_instrumentation__.Notify(184391)
	}
	__antithesis_instrumentation__.Notify(184381)

	_, _ = c.OnOutgoingPing, c.OnIncomingPing

	return nil
}

func NewContext(ctx context.Context, opts ContextOptions) *Context {
	__antithesis_instrumentation__.Notify(184392)
	if err := opts.validate(); err != nil {
		__antithesis_instrumentation__.Notify(184400)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(184401)
	}
	__antithesis_instrumentation__.Notify(184393)

	if opts.NodeID == nil {
		__antithesis_instrumentation__.Notify(184402)

		var c base.NodeIDContainer
		opts.NodeID = &c
	} else {
		__antithesis_instrumentation__.Notify(184403)
	}
	__antithesis_instrumentation__.Notify(184394)

	if opts.StorageClusterID == nil {
		__antithesis_instrumentation__.Notify(184404)

		var c base.ClusterIDContainer
		opts.StorageClusterID = &c
	} else {
		__antithesis_instrumentation__.Notify(184405)
	}
	__antithesis_instrumentation__.Notify(184395)

	if opts.LogicalClusterID == nil {
		__antithesis_instrumentation__.Notify(184406)
		if opts.TenantID.IsSystem() {
			__antithesis_instrumentation__.Notify(184407)

			opts.LogicalClusterID = opts.StorageClusterID
		} else {
			__antithesis_instrumentation__.Notify(184408)

			logicalClusterID := &base.ClusterIDContainer{}
			hasher := fnv.New64a()
			var b [8]byte
			binary.BigEndian.PutUint64(b[:], opts.TenantID.ToUint64())
			hasher.Write(b[:])
			hashedTenantID := hasher.Sum64()

			prevOnSet := opts.StorageClusterID.OnSet
			opts.StorageClusterID.OnSet = func(id uuid.UUID) {
				__antithesis_instrumentation__.Notify(184410)
				if prevOnSet != nil {
					__antithesis_instrumentation__.Notify(184412)
					prevOnSet(id)
				} else {
					__antithesis_instrumentation__.Notify(184413)
				}
				__antithesis_instrumentation__.Notify(184411)
				hiLo := id.ToUint128()
				hiLo.Lo += hashedTenantID
				logicalClusterID.Set(ctx, uuid.FromUint128(hiLo))
			}
			__antithesis_instrumentation__.Notify(184409)
			opts.LogicalClusterID = logicalClusterID
		}
	} else {
		__antithesis_instrumentation__.Notify(184414)
	}
	__antithesis_instrumentation__.Notify(184396)

	masterCtx, cancel := context.WithCancel(ctx)

	rpcCtx := &Context{
		ContextOptions:  opts,
		SecurityContext: MakeSecurityContext(opts.Config, security.ClusterTLSSettings(opts.Settings), opts.TenantID),
		breakerClock: breakerClock{
			clock: opts.Clock,
		},
		RemoteClocks: newRemoteClockMonitor(
			opts.Clock, 10*opts.Config.RPCHeartbeatInterval, opts.Config.HistogramWindowInterval()),
		rpcCompression:   enableRPCCompression,
		MasterCtx:        masterCtx,
		metrics:          makeMetrics(),
		heartbeatTimeout: 2 * opts.Config.RPCHeartbeatInterval,
	}
	if id := opts.Knobs.StorageClusterID; id != nil {
		__antithesis_instrumentation__.Notify(184415)
		rpcCtx.StorageClusterID.Set(masterCtx, *id)
	} else {
		__antithesis_instrumentation__.Notify(184416)
	}
	__antithesis_instrumentation__.Notify(184397)

	waitQuiesce := func(context.Context) {
		__antithesis_instrumentation__.Notify(184417)
		<-rpcCtx.Stopper.ShouldQuiesce()

		cancel()
		rpcCtx.conns.Range(func(k, v interface{}) bool {
			__antithesis_instrumentation__.Notify(184418)
			conn := v.(*Connection)
			conn.initOnce.Do(func() {
				__antithesis_instrumentation__.Notify(184420)

				if conn.dialErr == nil {
					__antithesis_instrumentation__.Notify(184421)
					conn.dialErr = errDialRejected
				} else {
					__antithesis_instrumentation__.Notify(184422)
				}
			})
			__antithesis_instrumentation__.Notify(184419)
			rpcCtx.removeConn(conn, k.(connKey))
			return true
		})
	}
	__antithesis_instrumentation__.Notify(184398)
	if err := rpcCtx.Stopper.RunAsyncTask(rpcCtx.MasterCtx, "wait-rpcctx-quiesce", waitQuiesce); err != nil {
		__antithesis_instrumentation__.Notify(184423)
		waitQuiesce(rpcCtx.MasterCtx)
	} else {
		__antithesis_instrumentation__.Notify(184424)
	}
	__antithesis_instrumentation__.Notify(184399)
	return rpcCtx
}

func (rpcCtx *Context) ClusterName() string {
	__antithesis_instrumentation__.Notify(184425)
	if rpcCtx == nil {
		__antithesis_instrumentation__.Notify(184427)

		return "<MISSING RPC CONTEXT>"
	} else {
		__antithesis_instrumentation__.Notify(184428)
	}
	__antithesis_instrumentation__.Notify(184426)
	return rpcCtx.Config.ClusterName
}

func (rpcCtx *Context) GetStatsMap() *syncmap.Map {
	__antithesis_instrumentation__.Notify(184429)
	return &rpcCtx.stats.stats
}

func (rpcCtx *Context) Metrics() *Metrics {
	__antithesis_instrumentation__.Notify(184430)
	return &rpcCtx.metrics
}

func (rpcCtx *Context) GetLocalInternalClientForAddr(
	target string, nodeID roachpb.NodeID,
) roachpb.InternalClient {
	__antithesis_instrumentation__.Notify(184431)
	if target == rpcCtx.Config.AdvertiseAddr && func() bool {
		__antithesis_instrumentation__.Notify(184433)
		return nodeID == rpcCtx.NodeID.Get() == true
	}() == true {
		__antithesis_instrumentation__.Notify(184434)
		return rpcCtx.localInternalClient
	} else {
		__antithesis_instrumentation__.Notify(184435)
	}
	__antithesis_instrumentation__.Notify(184432)
	return nil
}

type internalClientAdapter struct {
	server roachpb.InternalServer
}

func (a internalClientAdapter) Batch(
	ctx context.Context, ba *roachpb.BatchRequest, _ ...grpc.CallOption,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(184436)

	ba.AdmissionHeader.SourceLocation = roachpb.AdmissionHeader_LOCAL
	return a.server.Batch(ctx, ba)
}

func (a internalClientAdapter) RangeLookup(
	ctx context.Context, rl *roachpb.RangeLookupRequest, _ ...grpc.CallOption,
) (*roachpb.RangeLookupResponse, error) {
	__antithesis_instrumentation__.Notify(184437)
	return a.server.RangeLookup(ctx, rl)
}

func (a internalClientAdapter) Join(
	ctx context.Context, req *roachpb.JoinNodeRequest, _ ...grpc.CallOption,
) (*roachpb.JoinNodeResponse, error) {
	__antithesis_instrumentation__.Notify(184438)
	return a.server.Join(ctx, req)
}

func (a internalClientAdapter) ResetQuorum(
	ctx context.Context, req *roachpb.ResetQuorumRequest, _ ...grpc.CallOption,
) (*roachpb.ResetQuorumResponse, error) {
	__antithesis_instrumentation__.Notify(184439)
	return a.server.ResetQuorum(ctx, req)
}

func (a internalClientAdapter) TokenBucket(
	ctx context.Context, in *roachpb.TokenBucketRequest, opts ...grpc.CallOption,
) (*roachpb.TokenBucketResponse, error) {
	__antithesis_instrumentation__.Notify(184440)
	return a.server.TokenBucket(ctx, in)
}

func (a internalClientAdapter) GetSpanConfigs(
	ctx context.Context, req *roachpb.GetSpanConfigsRequest, _ ...grpc.CallOption,
) (*roachpb.GetSpanConfigsResponse, error) {
	__antithesis_instrumentation__.Notify(184441)
	return a.server.GetSpanConfigs(ctx, req)
}

func (a internalClientAdapter) GetAllSystemSpanConfigsThatApply(
	ctx context.Context, req *roachpb.GetAllSystemSpanConfigsThatApplyRequest, _ ...grpc.CallOption,
) (*roachpb.GetAllSystemSpanConfigsThatApplyResponse, error) {
	__antithesis_instrumentation__.Notify(184442)
	return a.server.GetAllSystemSpanConfigsThatApply(ctx, req)
}

func (a internalClientAdapter) UpdateSpanConfigs(
	ctx context.Context, req *roachpb.UpdateSpanConfigsRequest, _ ...grpc.CallOption,
) (*roachpb.UpdateSpanConfigsResponse, error) {
	__antithesis_instrumentation__.Notify(184443)
	return a.server.UpdateSpanConfigs(ctx, req)
}

type respStreamClientAdapter struct {
	ctx   context.Context
	respC chan interface{}
	errC  chan error
}

func makeRespStreamClientAdapter(ctx context.Context) respStreamClientAdapter {
	__antithesis_instrumentation__.Notify(184444)
	return respStreamClientAdapter{
		ctx:   ctx,
		respC: make(chan interface{}, 128),
		errC:  make(chan error, 1),
	}
}

func (respStreamClientAdapter) Header() (metadata.MD, error) {
	__antithesis_instrumentation__.Notify(184445)
	panic("unimplemented")
}
func (respStreamClientAdapter) Trailer() metadata.MD {
	__antithesis_instrumentation__.Notify(184446)
	panic("unimplemented")
}
func (respStreamClientAdapter) CloseSend() error {
	__antithesis_instrumentation__.Notify(184447)
	panic("unimplemented")
}

func (respStreamClientAdapter) SetHeader(metadata.MD) error {
	__antithesis_instrumentation__.Notify(184448)
	panic("unimplemented")
}
func (respStreamClientAdapter) SendHeader(metadata.MD) error {
	__antithesis_instrumentation__.Notify(184449)
	panic("unimplemented")
}
func (respStreamClientAdapter) SetTrailer(metadata.MD) {
	__antithesis_instrumentation__.Notify(184450)
	panic("unimplemented")
}

func (a respStreamClientAdapter) Context() context.Context {
	__antithesis_instrumentation__.Notify(184451)
	return a.ctx
}
func (respStreamClientAdapter) SendMsg(m interface{}) error {
	__antithesis_instrumentation__.Notify(184452)
	panic("unimplemented")
}
func (respStreamClientAdapter) RecvMsg(m interface{}) error {
	__antithesis_instrumentation__.Notify(184453)
	panic("unimplemented")
}

func (a respStreamClientAdapter) recvInternal() (interface{}, error) {
	__antithesis_instrumentation__.Notify(184454)

	select {
	case e := <-a.respC:
		__antithesis_instrumentation__.Notify(184455)
		return e, nil
	case err := <-a.errC:
		__antithesis_instrumentation__.Notify(184456)
		select {
		case e := <-a.respC:
			__antithesis_instrumentation__.Notify(184457)
			a.errC <- err
			return e, nil
		default:
			__antithesis_instrumentation__.Notify(184458)
			return nil, err
		}
	}
}

func (a respStreamClientAdapter) sendInternal(e interface{}) error {
	__antithesis_instrumentation__.Notify(184459)
	select {
	case a.respC <- e:
		__antithesis_instrumentation__.Notify(184460)
		return nil
	case <-a.ctx.Done():
		__antithesis_instrumentation__.Notify(184461)
		return a.ctx.Err()
	}
}

type rangeFeedClientAdapter struct {
	respStreamClientAdapter
}

func (a rangeFeedClientAdapter) Recv() (*roachpb.RangeFeedEvent, error) {
	__antithesis_instrumentation__.Notify(184462)
	e, err := a.recvInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(184464)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(184465)
	}
	__antithesis_instrumentation__.Notify(184463)
	return e.(*roachpb.RangeFeedEvent), nil
}

func (a rangeFeedClientAdapter) Send(e *roachpb.RangeFeedEvent) error {
	__antithesis_instrumentation__.Notify(184466)
	return a.sendInternal(e)
}

var _ roachpb.Internal_RangeFeedClient = rangeFeedClientAdapter{}
var _ roachpb.Internal_RangeFeedServer = rangeFeedClientAdapter{}

func (a internalClientAdapter) RangeFeed(
	ctx context.Context, args *roachpb.RangeFeedRequest, _ ...grpc.CallOption,
) (roachpb.Internal_RangeFeedClient, error) {
	__antithesis_instrumentation__.Notify(184467)
	ctx, cancel := context.WithCancel(ctx)
	ctx, sp := tracing.ChildSpan(ctx, "/cockroach.roachpb.Internal/RangeFeed")
	rfAdapter := rangeFeedClientAdapter{
		respStreamClientAdapter: makeRespStreamClientAdapter(ctx),
	}

	args.AdmissionHeader.SourceLocation = roachpb.AdmissionHeader_LOCAL
	go func() {
		__antithesis_instrumentation__.Notify(184469)
		defer cancel()
		defer sp.Finish()
		err := a.server.RangeFeed(args, rfAdapter)
		if err == nil {
			__antithesis_instrumentation__.Notify(184471)
			err = io.EOF
		} else {
			__antithesis_instrumentation__.Notify(184472)
		}
		__antithesis_instrumentation__.Notify(184470)
		rfAdapter.errC <- err
	}()
	__antithesis_instrumentation__.Notify(184468)

	return rfAdapter, nil
}

type gossipSubscriptionClientAdapter struct {
	respStreamClientAdapter
}

func (a gossipSubscriptionClientAdapter) Recv() (*roachpb.GossipSubscriptionEvent, error) {
	__antithesis_instrumentation__.Notify(184473)
	e, err := a.recvInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(184475)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(184476)
	}
	__antithesis_instrumentation__.Notify(184474)
	return e.(*roachpb.GossipSubscriptionEvent), nil
}

func (a gossipSubscriptionClientAdapter) Send(e *roachpb.GossipSubscriptionEvent) error {
	__antithesis_instrumentation__.Notify(184477)
	return a.sendInternal(e)
}

var _ roachpb.Internal_GossipSubscriptionClient = gossipSubscriptionClientAdapter{}
var _ roachpb.Internal_GossipSubscriptionServer = gossipSubscriptionClientAdapter{}

func (a internalClientAdapter) GossipSubscription(
	ctx context.Context, args *roachpb.GossipSubscriptionRequest, _ ...grpc.CallOption,
) (roachpb.Internal_GossipSubscriptionClient, error) {
	__antithesis_instrumentation__.Notify(184478)
	ctx, cancel := context.WithCancel(ctx)
	ctx, sp := tracing.ChildSpan(ctx, "/cockroach.roachpb.Internal/GossipSubscription")
	gsAdapter := gossipSubscriptionClientAdapter{
		respStreamClientAdapter: makeRespStreamClientAdapter(ctx),
	}

	go func() {
		__antithesis_instrumentation__.Notify(184480)
		defer cancel()
		defer sp.Finish()
		err := a.server.GossipSubscription(args, gsAdapter)
		if err == nil {
			__antithesis_instrumentation__.Notify(184482)
			err = io.EOF
		} else {
			__antithesis_instrumentation__.Notify(184483)
		}
		__antithesis_instrumentation__.Notify(184481)
		gsAdapter.errC <- err
	}()
	__antithesis_instrumentation__.Notify(184479)

	return gsAdapter, nil
}

type tenantSettingsClientAdapter struct {
	respStreamClientAdapter
}

func (a tenantSettingsClientAdapter) Recv() (*roachpb.TenantSettingsEvent, error) {
	__antithesis_instrumentation__.Notify(184484)
	e, err := a.recvInternal()
	if err != nil {
		__antithesis_instrumentation__.Notify(184486)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(184487)
	}
	__antithesis_instrumentation__.Notify(184485)
	return e.(*roachpb.TenantSettingsEvent), nil
}

func (a tenantSettingsClientAdapter) Send(e *roachpb.TenantSettingsEvent) error {
	__antithesis_instrumentation__.Notify(184488)
	return a.sendInternal(e)
}

var _ roachpb.Internal_TenantSettingsClient = tenantSettingsClientAdapter{}
var _ roachpb.Internal_TenantSettingsServer = tenantSettingsClientAdapter{}

func (a internalClientAdapter) TenantSettings(
	ctx context.Context, args *roachpb.TenantSettingsRequest, _ ...grpc.CallOption,
) (roachpb.Internal_TenantSettingsClient, error) {
	__antithesis_instrumentation__.Notify(184489)
	ctx, cancel := context.WithCancel(ctx)
	gsAdapter := tenantSettingsClientAdapter{
		respStreamClientAdapter: makeRespStreamClientAdapter(ctx),
	}

	go func() {
		__antithesis_instrumentation__.Notify(184491)
		defer cancel()
		err := a.server.TenantSettings(args, gsAdapter)
		if err == nil {
			__antithesis_instrumentation__.Notify(184493)
			err = io.EOF
		} else {
			__antithesis_instrumentation__.Notify(184494)
		}
		__antithesis_instrumentation__.Notify(184492)
		gsAdapter.errC <- err
	}()
	__antithesis_instrumentation__.Notify(184490)

	return gsAdapter, nil
}

var _ roachpb.InternalClient = internalClientAdapter{}

func IsLocal(iface roachpb.InternalClient) bool {
	__antithesis_instrumentation__.Notify(184495)
	_, ok := iface.(internalClientAdapter)
	return ok
}

func (rpcCtx *Context) SetLocalInternalServer(internalServer roachpb.InternalServer) {
	__antithesis_instrumentation__.Notify(184496)
	rpcCtx.localInternalClient = internalClientAdapter{internalServer}
}

func (rpcCtx *Context) removeConn(conn *Connection, keys ...connKey) {
	__antithesis_instrumentation__.Notify(184497)
	for _, key := range keys {
		__antithesis_instrumentation__.Notify(184499)
		rpcCtx.conns.Delete(key)
	}
	__antithesis_instrumentation__.Notify(184498)
	log.Health.Infof(rpcCtx.MasterCtx, "closing %+v", keys)
	if grpcConn := conn.grpcConn; grpcConn != nil {
		__antithesis_instrumentation__.Notify(184500)
		err := grpcConn.Close()
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(184501)
			return !grpcutil.IsClosedConnection(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(184502)
			log.Health.Warningf(rpcCtx.MasterCtx, "failed to close client connection: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(184503)
		}
	} else {
		__antithesis_instrumentation__.Notify(184504)
	}
}

func (rpcCtx *Context) ConnHealth(
	target string, nodeID roachpb.NodeID, class ConnectionClass,
) error {
	__antithesis_instrumentation__.Notify(184505)

	if rpcCtx.GetLocalInternalClientForAddr(target, nodeID) != nil {
		__antithesis_instrumentation__.Notify(184508)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(184509)
	}
	__antithesis_instrumentation__.Notify(184506)
	if value, ok := rpcCtx.conns.Load(connKey{target, nodeID, class}); ok {
		__antithesis_instrumentation__.Notify(184510)
		return value.(*Connection).Health()
	} else {
		__antithesis_instrumentation__.Notify(184511)
	}
	__antithesis_instrumentation__.Notify(184507)
	return ErrNoConnection
}

func (rpcCtx *Context) GRPCDialOptions() ([]grpc.DialOption, error) {
	__antithesis_instrumentation__.Notify(184512)
	return rpcCtx.grpcDialOptions("", DefaultClass)
}

func (rpcCtx *Context) grpcDialOptions(
	target string, class ConnectionClass,
) ([]grpc.DialOption, error) {
	__antithesis_instrumentation__.Notify(184513)
	var dialOpts []grpc.DialOption
	if rpcCtx.Config.Insecure {
		__antithesis_instrumentation__.Notify(184522)

		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		__antithesis_instrumentation__.Notify(184523)
		var tlsConfig *tls.Config
		var err error
		if rpcCtx.tenID == roachpb.SystemTenantID {
			__antithesis_instrumentation__.Notify(184526)
			tlsConfig, err = rpcCtx.GetClientTLSConfig()
		} else {
			__antithesis_instrumentation__.Notify(184527)
			tlsConfig, err = rpcCtx.GetTenantTLSConfig()
		}
		__antithesis_instrumentation__.Notify(184524)
		if err != nil {
			__antithesis_instrumentation__.Notify(184528)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(184529)
		}
		__antithesis_instrumentation__.Notify(184525)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}
	__antithesis_instrumentation__.Notify(184514)

	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
		grpc.MaxCallSendMsgSize(math.MaxInt32),
	))

	if rpcCtx.rpcCompression {
		__antithesis_instrumentation__.Notify(184530)
		dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(grpc.UseCompressor((snappyCompressor{}).Name())))
	} else {
		__antithesis_instrumentation__.Notify(184531)
	}
	__antithesis_instrumentation__.Notify(184515)

	dialOpts = append(dialOpts, grpc.WithNoProxy())

	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor

	if tracer := rpcCtx.Stopper.Tracer(); tracer != nil {
		__antithesis_instrumentation__.Notify(184532)

		tagger := func(span *tracing.Span) {
			__antithesis_instrumentation__.Notify(184536)
			span.SetTag("node", attribute.IntValue(int(rpcCtx.NodeID.Get())))
		}
		__antithesis_instrumentation__.Notify(184533)
		compatMode := func(reqCtx context.Context) bool {
			__antithesis_instrumentation__.Notify(184537)
			return !rpcCtx.ContextOptions.Settings.Version.IsActive(reqCtx, clusterversion.SelectRPCsTakeTracingInfoInband)
		}
		__antithesis_instrumentation__.Notify(184534)

		if rpcCtx.ClientOnly {
			__antithesis_instrumentation__.Notify(184538)

			tagger = func(span *tracing.Span) { __antithesis_instrumentation__.Notify(184540) }
			__antithesis_instrumentation__.Notify(184539)
			compatMode = func(_ context.Context) bool { __antithesis_instrumentation__.Notify(184541); return false }
		} else {
			__antithesis_instrumentation__.Notify(184542)
		}
		__antithesis_instrumentation__.Notify(184535)

		unaryInterceptors = append(unaryInterceptors,
			tracing.ClientInterceptor(tracer, tagger, compatMode))
		streamInterceptors = append(streamInterceptors,
			tracing.StreamClientInterceptor(tracer, tagger))
	} else {
		__antithesis_instrumentation__.Notify(184543)
	}
	__antithesis_instrumentation__.Notify(184516)
	if rpcCtx.Knobs.UnaryClientInterceptor != nil {
		__antithesis_instrumentation__.Notify(184544)
		testingUnaryInterceptor := rpcCtx.Knobs.UnaryClientInterceptor(target, class)
		if testingUnaryInterceptor != nil {
			__antithesis_instrumentation__.Notify(184545)
			unaryInterceptors = append(unaryInterceptors, testingUnaryInterceptor)
		} else {
			__antithesis_instrumentation__.Notify(184546)
		}
	} else {
		__antithesis_instrumentation__.Notify(184547)
	}
	__antithesis_instrumentation__.Notify(184517)
	if rpcCtx.Knobs.StreamClientInterceptor != nil {
		__antithesis_instrumentation__.Notify(184548)
		testingStreamInterceptor := rpcCtx.Knobs.StreamClientInterceptor(target, class)
		if testingStreamInterceptor != nil {
			__antithesis_instrumentation__.Notify(184549)
			streamInterceptors = append(streamInterceptors, testingStreamInterceptor)
		} else {
			__antithesis_instrumentation__.Notify(184550)
		}
	} else {
		__antithesis_instrumentation__.Notify(184551)
	}
	__antithesis_instrumentation__.Notify(184518)
	if rpcCtx.Knobs.ArtificialLatencyMap != nil {
		__antithesis_instrumentation__.Notify(184552)
		dialerFunc := func(ctx context.Context, target string) (net.Conn, error) {
			__antithesis_instrumentation__.Notify(184554)
			dialer := net.Dialer{
				LocalAddr: sourceAddr,
			}
			return dialer.DialContext(ctx, "tcp", target)
		}
		__antithesis_instrumentation__.Notify(184553)
		latency := rpcCtx.Knobs.ArtificialLatencyMap[target]
		log.VEventf(rpcCtx.MasterCtx, 1, "connecting to node %s with simulated latency %dms", target, latency)
		dialer := artificialLatencyDialer{
			dialerFunc: dialerFunc,
			latencyMS:  latency,
		}
		dialerFunc = dialer.dial
		dialOpts = append(dialOpts, grpc.WithContextDialer(dialerFunc))
	} else {
		__antithesis_instrumentation__.Notify(184555)
	}
	__antithesis_instrumentation__.Notify(184519)

	if len(unaryInterceptors) > 0 {
		__antithesis_instrumentation__.Notify(184556)
		dialOpts = append(dialOpts, grpc.WithChainUnaryInterceptor(unaryInterceptors...))
	} else {
		__antithesis_instrumentation__.Notify(184557)
	}
	__antithesis_instrumentation__.Notify(184520)
	if len(streamInterceptors) > 0 {
		__antithesis_instrumentation__.Notify(184558)
		dialOpts = append(dialOpts, grpc.WithChainStreamInterceptor(streamInterceptors...))
	} else {
		__antithesis_instrumentation__.Notify(184559)
	}
	__antithesis_instrumentation__.Notify(184521)
	return dialOpts, nil
}

type growStackCodec struct {
	encoding.Codec
}

func (c growStackCodec) Unmarshal(data []byte, v interface{}) error {
	__antithesis_instrumentation__.Notify(184560)
	if _, ok := v.(*roachpb.BatchRequest); ok {
		__antithesis_instrumentation__.Notify(184562)
		growstack.Grow()
	} else {
		__antithesis_instrumentation__.Notify(184563)
	}
	__antithesis_instrumentation__.Notify(184561)
	return c.Codec.Unmarshal(data, v)
}

func init() {
	encoding.RegisterCodec(growStackCodec{Codec: codec{}})
}

type onlyOnceDialer struct {
	syncutil.Mutex
	dialed     bool
	closed     bool
	redialChan chan struct{}
}

func (ood *onlyOnceDialer) dial(ctx context.Context, addr string) (net.Conn, error) {
	__antithesis_instrumentation__.Notify(184564)
	ood.Lock()
	defer ood.Unlock()
	if !ood.dialed {
		__antithesis_instrumentation__.Notify(184566)
		ood.dialed = true
		dialer := net.Dialer{
			LocalAddr: sourceAddr,
		}
		return dialer.DialContext(ctx, "tcp", addr)
	} else {
		__antithesis_instrumentation__.Notify(184567)
		if !ood.closed {
			__antithesis_instrumentation__.Notify(184568)
			ood.closed = true
			close(ood.redialChan)
		} else {
			__antithesis_instrumentation__.Notify(184569)
		}
	}
	__antithesis_instrumentation__.Notify(184565)
	return nil, grpcutil.ErrCannotReuseClientConn
}

type dialerFunc func(context.Context, string) (net.Conn, error)

type artificialLatencyDialer struct {
	dialerFunc dialerFunc
	latencyMS  int
}

func (ald *artificialLatencyDialer) dial(ctx context.Context, addr string) (net.Conn, error) {
	__antithesis_instrumentation__.Notify(184570)
	conn, err := ald.dialerFunc(ctx, addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(184572)
		return conn, err
	} else {
		__antithesis_instrumentation__.Notify(184573)
	}
	__antithesis_instrumentation__.Notify(184571)
	return &delayingConn{
		Conn:    conn,
		latency: time.Duration(ald.latencyMS) * time.Millisecond,
		readBuf: new(bytes.Buffer),
	}, nil
}

type delayingListener struct {
	net.Listener
}

func NewDelayingListener(l net.Listener) net.Listener {
	__antithesis_instrumentation__.Notify(184574)
	return delayingListener{Listener: l}
}

func (d delayingListener) Accept() (net.Conn, error) {
	__antithesis_instrumentation__.Notify(184575)
	c, err := d.Listener.Accept()
	if err != nil {
		__antithesis_instrumentation__.Notify(184577)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(184578)
	}
	__antithesis_instrumentation__.Notify(184576)
	return &delayingConn{
		Conn: c,

		latency: time.Duration(0) * time.Millisecond,
		readBuf: new(bytes.Buffer),
	}, nil
}

type delayingConn struct {
	net.Conn
	latency     time.Duration
	lastSendEnd time.Time
	readBuf     *bytes.Buffer
}

func (d delayingConn) Write(b []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(184579)
	tNow := timeutil.Now()
	if d.lastSendEnd.Before(tNow) {
		__antithesis_instrumentation__.Notify(184582)
		d.lastSendEnd = tNow
	} else {
		__antithesis_instrumentation__.Notify(184583)
	}
	__antithesis_instrumentation__.Notify(184580)
	hdr := delayingHeader{
		Magic:    magic,
		ReadTime: d.lastSendEnd.Add(d.latency).UnixNano(),
		Sz:       int32(len(b)),
		DelayMS:  int32(d.latency / time.Millisecond),
	}
	if err := binary.Write(d.Conn, binary.BigEndian, hdr); err != nil {
		__antithesis_instrumentation__.Notify(184584)
		return n, err
	} else {
		__antithesis_instrumentation__.Notify(184585)
	}
	__antithesis_instrumentation__.Notify(184581)
	x, err := d.Conn.Write(b)
	n += x
	return n, err
}

var errMagicNotFound = errors.New("didn't get expected magic bytes header")

func (d *delayingConn) Read(b []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(184586)
	if d.readBuf.Len() == 0 {
		__antithesis_instrumentation__.Notify(184588)
		var hdr delayingHeader
		if err := binary.Read(d.Conn, binary.BigEndian, &hdr); err != nil {
			__antithesis_instrumentation__.Notify(184593)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(184594)
		}
		__antithesis_instrumentation__.Notify(184589)

		if hdr.Magic != magic {
			__antithesis_instrumentation__.Notify(184595)
			return 0, errors.WithStack(errMagicNotFound)
		} else {
			__antithesis_instrumentation__.Notify(184596)
		}
		__antithesis_instrumentation__.Notify(184590)

		if d.latency == 0 && func() bool {
			__antithesis_instrumentation__.Notify(184597)
			return hdr.DelayMS != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(184598)
			d.latency = time.Duration(hdr.DelayMS) * time.Millisecond
		} else {
			__antithesis_instrumentation__.Notify(184599)
		}
		__antithesis_instrumentation__.Notify(184591)
		defer func() {
			__antithesis_instrumentation__.Notify(184600)
			time.Sleep(timeutil.Until(timeutil.Unix(0, hdr.ReadTime)))
		}()
		__antithesis_instrumentation__.Notify(184592)
		if _, err := io.CopyN(d.readBuf, d.Conn, int64(hdr.Sz)); err != nil {
			__antithesis_instrumentation__.Notify(184601)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(184602)
		}
	} else {
		__antithesis_instrumentation__.Notify(184603)
	}
	__antithesis_instrumentation__.Notify(184587)
	return d.readBuf.Read(b)
}

const magic = 0xfeedfeed

type delayingHeader struct {
	Magic    int64
	ReadTime int64
	Sz       int32
	DelayMS  int32
}

func (rpcCtx *Context) makeDialCtx(
	target string, remoteNodeID roachpb.NodeID, class ConnectionClass,
) context.Context {
	__antithesis_instrumentation__.Notify(184604)
	dialCtx := rpcCtx.MasterCtx
	var rnodeID interface{} = remoteNodeID
	if remoteNodeID == 0 {
		__antithesis_instrumentation__.Notify(184606)
		rnodeID = '?'
	} else {
		__antithesis_instrumentation__.Notify(184607)
	}
	__antithesis_instrumentation__.Notify(184605)
	dialCtx = logtags.AddTag(dialCtx, "rnode", rnodeID)
	dialCtx = logtags.AddTag(dialCtx, "raddr", target)
	dialCtx = logtags.AddTag(dialCtx, "class", class)
	return dialCtx
}

func (rpcCtx *Context) GRPCDialRaw(target string) (*grpc.ClientConn, <-chan struct{}, error) {
	__antithesis_instrumentation__.Notify(184608)
	ctx := rpcCtx.makeDialCtx(target, 0, DefaultClass)
	return rpcCtx.grpcDialRaw(ctx, target, 0, DefaultClass)
}

func (rpcCtx *Context) grpcDialRaw(
	ctx context.Context, target string, remoteNodeID roachpb.NodeID, class ConnectionClass,
) (*grpc.ClientConn, <-chan struct{}, error) {
	__antithesis_instrumentation__.Notify(184609)
	dialOpts, err := rpcCtx.grpcDialOptions(target, class)
	if err != nil {
		__antithesis_instrumentation__.Notify(184614)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(184615)
	}
	__antithesis_instrumentation__.Notify(184610)

	dialOpts = append(dialOpts, grpc.WithStatsHandler(rpcCtx.stats.newClient(target)))

	backoffConfig := backoff.DefaultConfig
	backoffConfig.MaxDelay = maxBackoff
	dialOpts = append(dialOpts, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoffConfig,
		MinConnectTimeout: minConnectionTimeout}))
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(clientKeepalive))
	dialOpts = append(dialOpts, grpc.WithInitialConnWindowSize(initialConnWindowSize))
	if class == RangefeedClass {
		__antithesis_instrumentation__.Notify(184616)
		dialOpts = append(dialOpts, grpc.WithInitialWindowSize(rangefeedInitialWindowSize))
	} else {
		__antithesis_instrumentation__.Notify(184617)
		dialOpts = append(dialOpts, grpc.WithInitialWindowSize(initialWindowSize))
	}
	__antithesis_instrumentation__.Notify(184611)

	dialer := onlyOnceDialer{
		redialChan: make(chan struct{}),
	}
	dialerFunc := dialer.dial
	if rpcCtx.Knobs.ArtificialLatencyMap != nil {
		__antithesis_instrumentation__.Notify(184618)
		latency := rpcCtx.Knobs.ArtificialLatencyMap[target]
		log.VEventf(ctx, 1, "connecting with simulated latency %dms",
			latency)
		dialer := artificialLatencyDialer{
			dialerFunc: dialerFunc,
			latencyMS:  latency,
		}
		dialerFunc = dialer.dial
	} else {
		__antithesis_instrumentation__.Notify(184619)
	}
	__antithesis_instrumentation__.Notify(184612)
	dialOpts = append(dialOpts, grpc.WithContextDialer(dialerFunc))

	dialOpts = append(dialOpts, rpcCtx.testingDialOpts...)

	log.Health.Infof(ctx, "dialing")
	conn, err := grpc.DialContext(ctx, target, dialOpts...)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(184620)
		return rpcCtx.MasterCtx.Err() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(184621)

		err = errDialRejected
	} else {
		__antithesis_instrumentation__.Notify(184622)
	}
	__antithesis_instrumentation__.Notify(184613)
	return conn, dialer.redialChan, err
}

func (rpcCtx *Context) GRPCUnvalidatedDial(target string) *Connection {
	__antithesis_instrumentation__.Notify(184623)
	ctx := rpcCtx.makeDialCtx(target, 0, SystemClass)
	return rpcCtx.grpcDialNodeInternal(ctx, target, 0, SystemClass)
}

func (rpcCtx *Context) GRPCDialNode(
	target string, remoteNodeID roachpb.NodeID, class ConnectionClass,
) *Connection {
	__antithesis_instrumentation__.Notify(184624)
	ctx := rpcCtx.makeDialCtx(target, remoteNodeID, class)
	if remoteNodeID == 0 && func() bool {
		__antithesis_instrumentation__.Notify(184626)
		return !rpcCtx.TestingAllowNamedRPCToAnonymousServer == true
	}() == true {
		__antithesis_instrumentation__.Notify(184627)
		log.Fatalf(ctx, "%v", errors.AssertionFailedf("invalid node ID 0 in GRPCDialNode()"))
	} else {
		__antithesis_instrumentation__.Notify(184628)
	}
	__antithesis_instrumentation__.Notify(184625)
	return rpcCtx.grpcDialNodeInternal(ctx, target, remoteNodeID, class)
}

func (rpcCtx *Context) GRPCDialPod(
	target string, remoteInstanceID base.SQLInstanceID, class ConnectionClass,
) *Connection {
	__antithesis_instrumentation__.Notify(184629)
	return rpcCtx.GRPCDialNode(target, roachpb.NodeID(remoteInstanceID), class)
}

func (rpcCtx *Context) grpcDialNodeInternal(
	ctx context.Context, target string, remoteNodeID roachpb.NodeID, class ConnectionClass,
) *Connection {
	__antithesis_instrumentation__.Notify(184630)
	thisConnKeys := []connKey{{target, remoteNodeID, class}}
	value, ok := rpcCtx.conns.Load(thisConnKeys[0])
	if !ok {
		__antithesis_instrumentation__.Notify(184633)
		value, _ = rpcCtx.conns.LoadOrStore(thisConnKeys[0], newConnectionToNodeID(rpcCtx.Stopper, remoteNodeID))
		if remoteNodeID != 0 {
			__antithesis_instrumentation__.Notify(184634)

			otherKey := connKey{target, 0, class}
			if _, loaded := rpcCtx.conns.LoadOrStore(otherKey, value); !loaded {
				__antithesis_instrumentation__.Notify(184635)
				thisConnKeys = append(thisConnKeys, otherKey)
			} else {
				__antithesis_instrumentation__.Notify(184636)
			}
		} else {
			__antithesis_instrumentation__.Notify(184637)
		}
	} else {
		__antithesis_instrumentation__.Notify(184638)
	}
	__antithesis_instrumentation__.Notify(184631)

	conn := value.(*Connection)
	conn.initOnce.Do(func() {
		__antithesis_instrumentation__.Notify(184639)

		var redialChan <-chan struct{}
		conn.grpcConn, redialChan, conn.dialErr = rpcCtx.grpcDialRaw(ctx, target, remoteNodeID, class)
		if conn.dialErr == nil {
			__antithesis_instrumentation__.Notify(184641)
			if err := rpcCtx.Stopper.RunAsyncTask(
				logtags.AddTag(ctx, "heartbeat", nil),
				"rpc.Context: grpc heartbeat", func(ctx context.Context) {
					__antithesis_instrumentation__.Notify(184642)
					err := rpcCtx.runHeartbeat(ctx, conn, target, redialChan)
					if err != nil && func() bool {
						__antithesis_instrumentation__.Notify(184644)
						return !grpcutil.IsClosedConnection(err) == true
					}() == true && func() bool {
						__antithesis_instrumentation__.Notify(184645)
						return !grpcutil.IsConnectionRejected(err) == true
					}() == true {
						__antithesis_instrumentation__.Notify(184646)
						log.Health.Errorf(ctx, "removing connection to %s due to error: %v", target, err)
					} else {
						__antithesis_instrumentation__.Notify(184647)
					}
					__antithesis_instrumentation__.Notify(184643)
					rpcCtx.removeConn(conn, thisConnKeys...)
				}); err != nil {
				__antithesis_instrumentation__.Notify(184648)

				_ = err
				conn.dialErr = errDialRejected
			} else {
				__antithesis_instrumentation__.Notify(184649)
			}
		} else {
			__antithesis_instrumentation__.Notify(184650)
		}
		__antithesis_instrumentation__.Notify(184640)
		if conn.dialErr != nil {
			__antithesis_instrumentation__.Notify(184651)
			rpcCtx.removeConn(conn, thisConnKeys...)
		} else {
			__antithesis_instrumentation__.Notify(184652)
		}
	})
	__antithesis_instrumentation__.Notify(184632)

	return conn
}

func (rpcCtx *Context) NewBreaker(name string) *circuit.Breaker {
	__antithesis_instrumentation__.Notify(184653)
	if rpcCtx.BreakerFactory != nil {
		__antithesis_instrumentation__.Notify(184655)
		return rpcCtx.BreakerFactory()
	} else {
		__antithesis_instrumentation__.Notify(184656)
	}
	__antithesis_instrumentation__.Notify(184654)
	return newBreaker(rpcCtx.MasterCtx, name, &rpcCtx.breakerClock)
}

var ErrNotHeartbeated = errors.New("not yet heartbeated")

var ErrNoConnection = errors.New("no connection found")

func (rpcCtx *Context) runHeartbeat(
	ctx context.Context, conn *Connection, target string, redialChan <-chan struct{},
) (retErr error) {
	__antithesis_instrumentation__.Notify(184657)
	rpcCtx.metrics.HeartbeatLoopsStarted.Inc(1)

	state := updateHeartbeatState(&rpcCtx.metrics, heartbeatNotRunning, heartbeatInitializing)
	initialHeartbeatDone := false
	setInitialHeartbeatDone := func() {
		__antithesis_instrumentation__.Notify(184660)
		if !initialHeartbeatDone {
			__antithesis_instrumentation__.Notify(184661)
			close(conn.initialHeartbeatDone)
			initialHeartbeatDone = true
		} else {
			__antithesis_instrumentation__.Notify(184662)
		}
	}
	__antithesis_instrumentation__.Notify(184658)
	defer func() {
		__antithesis_instrumentation__.Notify(184663)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(184665)
			rpcCtx.metrics.HeartbeatLoopsExited.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(184666)
		}
		__antithesis_instrumentation__.Notify(184664)
		updateHeartbeatState(&rpcCtx.metrics, state, heartbeatNotRunning)
		setInitialHeartbeatDone()
	}()
	__antithesis_instrumentation__.Notify(184659)
	maxOffset := rpcCtx.Clock.MaxOffset()
	maxOffsetNanos := maxOffset.Nanoseconds()

	heartbeatClient := NewHeartbeatClient(conn.grpcConn)

	var heartbeatTimer timeutil.Timer
	defer heartbeatTimer.Stop()

	heartbeatTimer.Reset(0)
	everSucceeded := false

	returnErr := false
	for {
		__antithesis_instrumentation__.Notify(184667)
		select {
		case <-redialChan:
			__antithesis_instrumentation__.Notify(184670)
			return grpcutil.ErrCannotReuseClientConn
		case <-rpcCtx.Stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(184671)
			return nil
		case <-heartbeatTimer.C:
			__antithesis_instrumentation__.Notify(184672)
			heartbeatTimer.Read = true
		}
		__antithesis_instrumentation__.Notify(184668)

		if err := rpcCtx.Stopper.RunTaskWithErr(ctx, "rpc heartbeat", func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(184673)

			clusterID := rpcCtx.StorageClusterID.Get()
			request := &PingRequest{
				OriginNodeID:         rpcCtx.NodeID.Get(),
				OriginAddr:           rpcCtx.Config.Addr,
				OriginMaxOffsetNanos: maxOffsetNanos,
				ClusterID:            &clusterID,
				TargetNodeID:         conn.remoteNodeID,
				ServerVersion:        rpcCtx.Settings.Version.BinaryVersion(),
			}

			interceptor := func(context.Context, *PingRequest) error { __antithesis_instrumentation__.Notify(184683); return nil }
			__antithesis_instrumentation__.Notify(184674)
			if fn := rpcCtx.OnOutgoingPing; fn != nil {
				__antithesis_instrumentation__.Notify(184684)
				interceptor = fn
			} else {
				__antithesis_instrumentation__.Notify(184685)
			}
			__antithesis_instrumentation__.Notify(184675)

			var response *PingResponse
			sendTime := rpcCtx.Clock.PhysicalTime()
			ping := func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(184686)

				if err := interceptor(ctx, request); err != nil {
					__antithesis_instrumentation__.Notify(184688)
					returnErr = true
					return err
				} else {
					__antithesis_instrumentation__.Notify(184689)
				}
				__antithesis_instrumentation__.Notify(184687)
				var err error
				response, err = heartbeatClient.Ping(ctx, request)
				return err
			}
			__antithesis_instrumentation__.Notify(184676)
			var err error
			if rpcCtx.heartbeatTimeout > 0 {
				__antithesis_instrumentation__.Notify(184690)
				err = contextutil.RunWithTimeout(ctx, "rpc heartbeat", rpcCtx.heartbeatTimeout, ping)
			} else {
				__antithesis_instrumentation__.Notify(184691)
				err = ping(ctx)
			}
			__antithesis_instrumentation__.Notify(184677)

			if grpcutil.IsConnectionRejected(err) {
				__antithesis_instrumentation__.Notify(184692)
				returnErr = true
			} else {
				__antithesis_instrumentation__.Notify(184693)
			}
			__antithesis_instrumentation__.Notify(184678)

			if err == nil {
				__antithesis_instrumentation__.Notify(184694)

				if !rpcCtx.Config.DisableClusterNameVerification && func() bool {
					__antithesis_instrumentation__.Notify(184695)
					return !response.DisableClusterNameVerification == true
				}() == true {
					__antithesis_instrumentation__.Notify(184696)
					err = errors.Wrap(
						checkClusterName(rpcCtx.Config.ClusterName, response.ClusterName),
						"cluster name check failed on ping response")
					if err != nil {
						__antithesis_instrumentation__.Notify(184697)
						returnErr = true
					} else {
						__antithesis_instrumentation__.Notify(184698)
					}
				} else {
					__antithesis_instrumentation__.Notify(184699)
				}
			} else {
				__antithesis_instrumentation__.Notify(184700)
			}
			__antithesis_instrumentation__.Notify(184679)

			if err == nil {
				__antithesis_instrumentation__.Notify(184701)
				err = errors.Wrap(
					checkVersion(ctx, rpcCtx.Settings, response.ServerVersion),
					"version compatibility check failed on ping response")
				if err != nil {
					__antithesis_instrumentation__.Notify(184702)
					returnErr = true
				} else {
					__antithesis_instrumentation__.Notify(184703)
				}
			} else {
				__antithesis_instrumentation__.Notify(184704)
			}
			__antithesis_instrumentation__.Notify(184680)

			if err == nil {
				__antithesis_instrumentation__.Notify(184705)
				everSucceeded = true
				receiveTime := rpcCtx.Clock.PhysicalTime()

				pingDuration := receiveTime.Sub(sendTime)
				maxOffset := rpcCtx.Clock.MaxOffset()
				if pingDuration > maximumPingDurationMult*maxOffset {
					__antithesis_instrumentation__.Notify(184707)
					request.Offset.Reset()
				} else {
					__antithesis_instrumentation__.Notify(184708)

					request.Offset.MeasuredAt = receiveTime.UnixNano()
					request.Offset.Uncertainty = (pingDuration / 2).Nanoseconds()
					remoteTimeNow := timeutil.Unix(0, response.ServerTime).Add(pingDuration / 2)
					request.Offset.Offset = remoteTimeNow.Sub(receiveTime).Nanoseconds()
				}
				__antithesis_instrumentation__.Notify(184706)
				rpcCtx.RemoteClocks.UpdateOffset(ctx, target, request.Offset, pingDuration)

				if cb := rpcCtx.HeartbeatCB; cb != nil {
					__antithesis_instrumentation__.Notify(184709)
					cb()
				} else {
					__antithesis_instrumentation__.Notify(184710)
				}
			} else {
				__antithesis_instrumentation__.Notify(184711)
			}
			__antithesis_instrumentation__.Notify(184681)

			hr := heartbeatResult{
				everSucceeded: everSucceeded,
				err:           err,
			}
			state = updateHeartbeatState(&rpcCtx.metrics, state, hr.state())
			conn.heartbeatResult.Store(hr)
			setInitialHeartbeatDone()
			if returnErr {
				__antithesis_instrumentation__.Notify(184712)
				return err
			} else {
				__antithesis_instrumentation__.Notify(184713)
			}
			__antithesis_instrumentation__.Notify(184682)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(184714)
			return err
		} else {
			__antithesis_instrumentation__.Notify(184715)
		}
		__antithesis_instrumentation__.Notify(184669)

		heartbeatTimer.Reset(rpcCtx.Config.RPCHeartbeatInterval)
	}
}

func (rpcCtx *Context) NewHeartbeatService() *HeartbeatService {
	__antithesis_instrumentation__.Notify(184716)
	return &HeartbeatService{
		clock:                                 rpcCtx.Clock,
		remoteClockMonitor:                    rpcCtx.RemoteClocks,
		clusterName:                           rpcCtx.ClusterName(),
		disableClusterNameVerification:        rpcCtx.Config.DisableClusterNameVerification,
		clusterID:                             rpcCtx.StorageClusterID,
		nodeID:                                rpcCtx.NodeID,
		settings:                              rpcCtx.Settings,
		onHandlePing:                          rpcCtx.OnIncomingPing,
		testingAllowNamedRPCToAnonymousServer: rpcCtx.TestingAllowNamedRPCToAnonymousServer,
	}
}
