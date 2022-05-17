package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"golang.org/x/sync/syncmap"
	"google.golang.org/grpc/stats"
)

type remoteAddrKey struct{}

type Stats struct {
	count    int64
	incoming int64
	outgoing int64
}

func (s *Stats) Count() int64 {
	__antithesis_instrumentation__.Notify(185551)
	return atomic.LoadInt64(&s.count)
}

func (s *Stats) Incoming() int64 {
	__antithesis_instrumentation__.Notify(185552)
	return atomic.LoadInt64(&s.incoming)
}

func (s *Stats) Outgoing() int64 {
	__antithesis_instrumentation__.Notify(185553)
	return atomic.LoadInt64(&s.outgoing)
}

func (s *Stats) record(rpcStats stats.RPCStats) {
	__antithesis_instrumentation__.Notify(185554)
	switch v := rpcStats.(type) {
	case *stats.InHeader:
		__antithesis_instrumentation__.Notify(185555)
		atomic.AddInt64(&s.incoming, int64(v.WireLength))
	case *stats.InPayload:
		__antithesis_instrumentation__.Notify(185556)

		atomic.AddInt64(&s.incoming, int64(v.WireLength+5))
	case *stats.InTrailer:
		__antithesis_instrumentation__.Notify(185557)
		atomic.AddInt64(&s.incoming, int64(v.WireLength))
	case *stats.OutHeader:
		__antithesis_instrumentation__.Notify(185558)

	case *stats.OutPayload:
		__antithesis_instrumentation__.Notify(185559)
		atomic.AddInt64(&s.outgoing, int64(v.WireLength))
	case *stats.OutTrailer:
		__antithesis_instrumentation__.Notify(185560)
		atomic.AddInt64(&s.outgoing, int64(v.WireLength))
	case *stats.End:
		__antithesis_instrumentation__.Notify(185561)
		atomic.AddInt64(&s.count, 1)
	}
}

type clientStatsHandler struct {
	stats *Stats
}

var _ stats.Handler = &clientStatsHandler{}

func (cs *clientStatsHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	__antithesis_instrumentation__.Notify(185562)
	return ctx
}

func (cs *clientStatsHandler) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	__antithesis_instrumentation__.Notify(185563)
	cs.stats.record(rpcStats)
}

func (cs *clientStatsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	__antithesis_instrumentation__.Notify(185564)
	return ctx
}

func (cs *clientStatsHandler) HandleConn(context.Context, stats.ConnStats) {
	__antithesis_instrumentation__.Notify(185565)
}

type StatsHandler struct {
	stats syncmap.Map
}

var _ stats.Handler = &StatsHandler{}

func (sh *StatsHandler) newClient(target string) stats.Handler {
	__antithesis_instrumentation__.Notify(185566)
	value, _ := sh.stats.LoadOrStore(target, &Stats{})
	return &clientStatsHandler{
		stats: value.(*Stats),
	}
}

func (sh *StatsHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	__antithesis_instrumentation__.Notify(185567)
	return ctx
}

func (sh *StatsHandler) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	__antithesis_instrumentation__.Notify(185568)
	remoteAddr, ok := ctx.Value(remoteAddrKey{}).(string)
	if !ok {
		__antithesis_instrumentation__.Notify(185571)
		log.Health.Warningf(ctx, "unable to record stats (%+v); remote addr not found in context", rpcStats)
		return
	} else {
		__antithesis_instrumentation__.Notify(185572)
	}
	__antithesis_instrumentation__.Notify(185569)

	value, ok := sh.stats.Load(remoteAddr)
	if !ok {
		__antithesis_instrumentation__.Notify(185573)
		value, _ = sh.stats.LoadOrStore(remoteAddr, &Stats{})
	} else {
		__antithesis_instrumentation__.Notify(185574)
	}
	__antithesis_instrumentation__.Notify(185570)
	value.(*Stats).record(rpcStats)
}

func (sh *StatsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	__antithesis_instrumentation__.Notify(185575)
	return context.WithValue(ctx, remoteAddrKey{}, cti.RemoteAddr.String())
}

func (sh *StatsHandler) HandleConn(context.Context, stats.ConnStats) {
	__antithesis_instrumentation__.Notify(185576)
}
