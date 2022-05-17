package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

type localTestClusterTransport struct {
	Transport
	latency time.Duration
}

func (l *localTestClusterTransport) SendNext(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(87801)
	if l.latency > 0 {
		__antithesis_instrumentation__.Notify(87803)
		time.Sleep(l.latency)
	} else {
		__antithesis_instrumentation__.Notify(87804)
	}
	__antithesis_instrumentation__.Notify(87802)
	return l.Transport.SendNext(ctx, ba)
}

func InitFactoryForLocalTestCluster(
	ctx context.Context,
	st *cluster.Settings,
	nodeDesc *roachpb.NodeDescriptor,
	tracer *tracing.Tracer,
	clock *hlc.Clock,
	latency time.Duration,
	stores kv.Sender,
	stopper *stop.Stopper,
	gossip *gossip.Gossip,
) kv.TxnSenderFactory {
	__antithesis_instrumentation__.Notify(87805)
	return NewTxnCoordSenderFactory(
		TxnCoordSenderFactoryConfig{
			AmbientCtx: log.MakeTestingAmbientContext(tracer),
			Settings:   st,
			Clock:      clock,
			Stopper:    stopper,
		},
		NewDistSenderForLocalTestCluster(ctx, st, nodeDesc, tracer, clock, latency, stores, stopper, gossip),
	)
}

func NewDistSenderForLocalTestCluster(
	ctx context.Context,
	st *cluster.Settings,
	nodeDesc *roachpb.NodeDescriptor,
	tracer *tracing.Tracer,
	clock *hlc.Clock,
	latency time.Duration,
	stores kv.Sender,
	stopper *stop.Stopper,
	g *gossip.Gossip,
) *DistSender {
	__antithesis_instrumentation__.Notify(87806)
	retryOpts := base.DefaultRetryOptions()
	retryOpts.Closer = stopper.ShouldQuiesce()
	rpcContext := rpc.NewInsecureTestingContext(ctx, clock, stopper)
	senderTransportFactory := SenderTransportFactory(tracer, stores)
	return NewDistSender(DistSenderConfig{
		AmbientCtx:         log.MakeTestingAmbientContext(tracer),
		Settings:           st,
		Clock:              clock,
		NodeDescs:          g,
		RPCContext:         rpcContext,
		RPCRetryOptions:    &retryOpts,
		nodeDescriptor:     nodeDesc,
		NodeDialer:         nodedialer.New(rpcContext, gossip.AddressResolver(g)),
		FirstRangeProvider: g,
		TestingKnobs: ClientTestingKnobs{
			TransportFactory: func(
				opts SendOptions,
				nodeDialer *nodedialer.Dialer,
				replicas ReplicaSlice,
			) (Transport, error) {
				__antithesis_instrumentation__.Notify(87807)
				transport, err := senderTransportFactory(opts, nodeDialer, replicas)
				if err != nil {
					__antithesis_instrumentation__.Notify(87809)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(87810)
				}
				__antithesis_instrumentation__.Notify(87808)
				return &localTestClusterTransport{transport, latency}, nil
			},
		},
	})
}
