package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type SendOptions struct {
	class   rpc.ConnectionClass
	metrics *DistSenderMetrics

	dontConsiderConnHealth bool
}

type TransportFactory func(
	SendOptions, *nodedialer.Dialer, ReplicaSlice,
) (Transport, error)

type Transport interface {
	IsExhausted() bool

	SendNext(context.Context, roachpb.BatchRequest) (*roachpb.BatchResponse, error)

	NextInternalClient(context.Context) (context.Context, roachpb.InternalClient, error)

	NextReplica() roachpb.ReplicaDescriptor

	SkipReplica()

	MoveToFront(roachpb.ReplicaDescriptor)

	Release()
}

const (
	healthUnhealthy = iota
	healthHealthy
)

func grpcTransportFactoryImpl(
	opts SendOptions, nodeDialer *nodedialer.Dialer, rs ReplicaSlice,
) (Transport, error) {
	__antithesis_instrumentation__.Notify(87976)
	transport := grpcTransportPool.Get().(*grpcTransport)

	replicas := transport.replicas

	if cap(replicas) < len(rs) {
		__antithesis_instrumentation__.Notify(87980)
		replicas = make([]roachpb.ReplicaDescriptor, len(rs))
	} else {
		__antithesis_instrumentation__.Notify(87981)
		replicas = replicas[:len(rs)]
	}
	__antithesis_instrumentation__.Notify(87977)

	var health util.FastIntMap
	for i := range rs {
		__antithesis_instrumentation__.Notify(87982)
		r := &rs[i]
		replicas[i] = r.ReplicaDescriptor
		healthy := nodeDialer.ConnHealth(r.NodeID, opts.class) == nil
		if healthy {
			__antithesis_instrumentation__.Notify(87983)
			health.Set(i, healthHealthy)
		} else {
			__antithesis_instrumentation__.Notify(87984)
			health.Set(i, healthUnhealthy)
		}
	}
	__antithesis_instrumentation__.Notify(87978)

	*transport = grpcTransport{
		opts:          opts,
		nodeDialer:    nodeDialer,
		class:         opts.class,
		replicas:      replicas,
		replicaHealth: health,
	}

	if !opts.dontConsiderConnHealth {
		__antithesis_instrumentation__.Notify(87985)

		transport.splitHealthy()
	} else {
		__antithesis_instrumentation__.Notify(87986)
	}
	__antithesis_instrumentation__.Notify(87979)

	return transport, nil
}

type grpcTransport struct {
	opts       SendOptions
	nodeDialer *nodedialer.Dialer
	class      rpc.ConnectionClass

	replicas []roachpb.ReplicaDescriptor

	replicaHealth util.FastIntMap

	nextReplicaIdx int
}

var grpcTransportPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(87987); return &grpcTransport{} },
}

func (gt *grpcTransport) Release() {
	__antithesis_instrumentation__.Notify(87988)
	*gt = grpcTransport{
		replicas: gt.replicas[0:],
	}
	grpcTransportPool.Put(gt)
}

func (gt *grpcTransport) IsExhausted() bool {
	__antithesis_instrumentation__.Notify(87989)
	return gt.nextReplicaIdx >= len(gt.replicas)
}

func (gt *grpcTransport) SendNext(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(87990)
	r := gt.replicas[gt.nextReplicaIdx]
	ctx, iface, err := gt.NextInternalClient(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(87992)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(87993)
	}
	__antithesis_instrumentation__.Notify(87991)

	ba.Replica = r
	return gt.sendBatch(ctx, r.NodeID, iface, ba)
}

func (gt *grpcTransport) sendBatch(
	ctx context.Context, nodeID roachpb.NodeID, iface roachpb.InternalClient, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(87994)

	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(87998)
		return nil, errors.Wrap(ctx.Err(), "aborted before batch send")
	} else {
		__antithesis_instrumentation__.Notify(87999)
	}
	__antithesis_instrumentation__.Notify(87995)

	gt.opts.metrics.SentCount.Inc(1)
	if rpc.IsLocal(iface) {
		__antithesis_instrumentation__.Notify(88000)
		gt.opts.metrics.LocalSentCount.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(88001)
	}
	__antithesis_instrumentation__.Notify(87996)
	reply, err := iface.Batch(ctx, &ba)

	if reply != nil && func() bool {
		__antithesis_instrumentation__.Notify(88002)
		return !rpc.IsLocal(iface) == true
	}() == true {
		__antithesis_instrumentation__.Notify(88003)
		for i := range reply.Responses {
			__antithesis_instrumentation__.Notify(88005)
			if err := reply.Responses[i].GetInner().Verify(ba.Requests[i].GetInner()); err != nil {
				__antithesis_instrumentation__.Notify(88006)
				log.Errorf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(88007)
			}
		}
		__antithesis_instrumentation__.Notify(88004)

		if len(reply.CollectedSpans) != 0 {
			__antithesis_instrumentation__.Notify(88008)
			span := tracing.SpanFromContext(ctx)
			if span == nil {
				__antithesis_instrumentation__.Notify(88010)
				return nil, errors.Errorf(
					"trying to ingest remote spans but there is no recording span set up")
			} else {
				__antithesis_instrumentation__.Notify(88011)
			}
			__antithesis_instrumentation__.Notify(88009)
			span.ImportRemoteSpans(reply.CollectedSpans)
		} else {
			__antithesis_instrumentation__.Notify(88012)
		}
	} else {
		__antithesis_instrumentation__.Notify(88013)
	}
	__antithesis_instrumentation__.Notify(87997)
	return reply, err
}

func (gt *grpcTransport) NextInternalClient(
	ctx context.Context,
) (context.Context, roachpb.InternalClient, error) {
	__antithesis_instrumentation__.Notify(88014)
	r := gt.replicas[gt.nextReplicaIdx]
	gt.nextReplicaIdx++
	return gt.nodeDialer.DialInternalClient(ctx, r.NodeID, gt.class)
}

func (gt *grpcTransport) NextReplica() roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(88015)
	if gt.IsExhausted() {
		__antithesis_instrumentation__.Notify(88017)
		return roachpb.ReplicaDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(88018)
	}
	__antithesis_instrumentation__.Notify(88016)
	return gt.replicas[gt.nextReplicaIdx]
}

func (gt *grpcTransport) SkipReplica() {
	__antithesis_instrumentation__.Notify(88019)
	if gt.IsExhausted() {
		__antithesis_instrumentation__.Notify(88021)
		return
	} else {
		__antithesis_instrumentation__.Notify(88022)
	}
	__antithesis_instrumentation__.Notify(88020)
	gt.nextReplicaIdx++
}

func (gt *grpcTransport) MoveToFront(replica roachpb.ReplicaDescriptor) {
	__antithesis_instrumentation__.Notify(88023)
	for i := range gt.replicas {
		__antithesis_instrumentation__.Notify(88024)
		if gt.replicas[i] == replica {
			__antithesis_instrumentation__.Notify(88025)

			if i < gt.nextReplicaIdx {
				__antithesis_instrumentation__.Notify(88027)
				gt.nextReplicaIdx--
			} else {
				__antithesis_instrumentation__.Notify(88028)
			}
			__antithesis_instrumentation__.Notify(88026)

			gt.replicas[i], gt.replicas[gt.nextReplicaIdx] = gt.replicas[gt.nextReplicaIdx], gt.replicas[i]
			return
		} else {
			__antithesis_instrumentation__.Notify(88029)
		}
	}
}

func (gt *grpcTransport) splitHealthy() {
	__antithesis_instrumentation__.Notify(88030)
	sort.Stable((*byHealth)(gt))
}

type byHealth grpcTransport

func (h *byHealth) Len() int { __antithesis_instrumentation__.Notify(88031); return len(h.replicas) }
func (h *byHealth) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(88032)
	h.replicas[i], h.replicas[j] = h.replicas[j], h.replicas[i]
	oldI := h.replicaHealth.GetDefault(i)
	h.replicaHealth.Set(i, h.replicaHealth.GetDefault(j))
	h.replicaHealth.Set(j, oldI)
}
func (h *byHealth) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(88033)
	ih, ok := h.replicaHealth.Get(i)
	if !ok {
		__antithesis_instrumentation__.Notify(88036)
		panic(fmt.Sprintf("missing health info for %s", h.replicas[i]))
	} else {
		__antithesis_instrumentation__.Notify(88037)
	}
	__antithesis_instrumentation__.Notify(88034)
	jh, ok := h.replicaHealth.Get(j)
	if !ok {
		__antithesis_instrumentation__.Notify(88038)
		panic(fmt.Sprintf("missing health info for %s", h.replicas[j]))
	} else {
		__antithesis_instrumentation__.Notify(88039)
	}
	__antithesis_instrumentation__.Notify(88035)
	return ih == healthHealthy && func() bool {
		__antithesis_instrumentation__.Notify(88040)
		return jh != healthHealthy == true
	}() == true
}

func SenderTransportFactory(tracer *tracing.Tracer, sender kv.Sender) TransportFactory {
	__antithesis_instrumentation__.Notify(88041)
	return func(
		_ SendOptions, _ *nodedialer.Dialer, replicas ReplicaSlice,
	) (Transport, error) {
		__antithesis_instrumentation__.Notify(88042)

		replica := replicas[0].ReplicaDescriptor
		return &senderTransport{tracer, sender, replica, false}, nil
	}
}

type senderTransport struct {
	tracer  *tracing.Tracer
	sender  kv.Sender
	replica roachpb.ReplicaDescriptor

	called bool
}

func (s *senderTransport) IsExhausted() bool {
	__antithesis_instrumentation__.Notify(88043)
	return s.called
}

func (s *senderTransport) SendNext(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(88044)
	if s.called {
		__antithesis_instrumentation__.Notify(88050)
		panic("called an exhausted transport")
	} else {
		__antithesis_instrumentation__.Notify(88051)
	}
	__antithesis_instrumentation__.Notify(88045)
	s.called = true

	ba.Replica = s.replica
	log.Eventf(ctx, "%v", ba.String())
	br, pErr := s.sender.Send(ctx, ba)
	if br == nil {
		__antithesis_instrumentation__.Notify(88052)
		br = &roachpb.BatchResponse{}
	} else {
		__antithesis_instrumentation__.Notify(88053)
	}
	__antithesis_instrumentation__.Notify(88046)
	if br.Error != nil {
		__antithesis_instrumentation__.Notify(88054)
		panic(roachpb.ErrorUnexpectedlySet(s.sender, br))
	} else {
		__antithesis_instrumentation__.Notify(88055)
	}
	__antithesis_instrumentation__.Notify(88047)
	br.Error = pErr
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88056)
		log.Eventf(ctx, "error: %v", pErr.String())
	} else {
		__antithesis_instrumentation__.Notify(88057)
	}
	__antithesis_instrumentation__.Notify(88048)

	if len(br.CollectedSpans) != 0 {
		__antithesis_instrumentation__.Notify(88058)
		span := tracing.SpanFromContext(ctx)
		if span == nil {
			__antithesis_instrumentation__.Notify(88060)
			panic("trying to ingest remote spans but there is no recording span set up")
		} else {
			__antithesis_instrumentation__.Notify(88061)
		}
		__antithesis_instrumentation__.Notify(88059)
		span.ImportRemoteSpans(br.CollectedSpans)
	} else {
		__antithesis_instrumentation__.Notify(88062)
	}
	__antithesis_instrumentation__.Notify(88049)

	return br, nil
}

func (s *senderTransport) NextInternalClient(
	ctx context.Context,
) (context.Context, roachpb.InternalClient, error) {
	__antithesis_instrumentation__.Notify(88063)
	panic("unimplemented")
}

func (s *senderTransport) NextReplica() roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(88064)
	if s.IsExhausted() {
		__antithesis_instrumentation__.Notify(88066)
		return roachpb.ReplicaDescriptor{}
	} else {
		__antithesis_instrumentation__.Notify(88067)
	}
	__antithesis_instrumentation__.Notify(88065)
	return s.replica
}

func (s *senderTransport) SkipReplica() {
	__antithesis_instrumentation__.Notify(88068)

	s.called = true
}

func (s *senderTransport) MoveToFront(replica roachpb.ReplicaDescriptor) {
	__antithesis_instrumentation__.Notify(88069)
}

func (s *senderTransport) Release() { __antithesis_instrumentation__.Notify(88070) }
