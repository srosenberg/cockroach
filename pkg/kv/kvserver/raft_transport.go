package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sort"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"google.golang.org/grpc"
)

const (
	raftSendBufferSize = 10000

	raftIdleTimeout = time.Minute
)

var targetRaftOutgoingBatchSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"kv.raft.command.target_batch_size",
	"size of a batch of raft commands after which it will be sent without further batching",
	64<<20,
	func(size int64) error {
		__antithesis_instrumentation__.Notify(113021)
		if size < 1 {
			__antithesis_instrumentation__.Notify(113023)
			return errors.New("must be positive")
		} else {
			__antithesis_instrumentation__.Notify(113024)
		}
		__antithesis_instrumentation__.Notify(113022)
		return nil
	},
)

type RaftMessageResponseStream interface {
	Send(*kvserverpb.RaftMessageResponse) error
}

type lockedRaftMessageResponseStream struct {
	wrapped MultiRaft_RaftMessageBatchServer
	sendMu  syncutil.Mutex
}

func (s *lockedRaftMessageResponseStream) Send(resp *kvserverpb.RaftMessageResponse) error {
	__antithesis_instrumentation__.Notify(113025)
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return s.wrapped.Send(resp)
}

func (s *lockedRaftMessageResponseStream) Recv() (*kvserverpb.RaftMessageRequestBatch, error) {
	__antithesis_instrumentation__.Notify(113026)

	return s.wrapped.Recv()
}

type SnapshotResponseStream interface {
	Send(*kvserverpb.SnapshotResponse) error
	Recv() (*kvserverpb.SnapshotRequest, error)
}

type RaftMessageHandler interface {
	HandleRaftRequest(ctx context.Context, req *kvserverpb.RaftMessageRequest,
		respStream RaftMessageResponseStream) *roachpb.Error

	HandleRaftResponse(context.Context, *kvserverpb.RaftMessageResponse) error

	HandleSnapshot(ctx context.Context, header *kvserverpb.SnapshotRequest_Header, respStream SnapshotResponseStream) error
}

type raftTransportStats struct {
	nodeID        roachpb.NodeID
	queue         int
	queueMax      int32
	clientSent    int64
	clientRecv    int64
	clientDropped int64
	serverSent    int64
	serverRecv    int64
}

type raftTransportStatsSlice []*raftTransportStats

func (s raftTransportStatsSlice) Len() int {
	__antithesis_instrumentation__.Notify(113027)
	return len(s)
}
func (s raftTransportStatsSlice) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(113028)
	s[i], s[j] = s[j], s[i]
}
func (s raftTransportStatsSlice) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(113029)
	return s[i].nodeID < s[j].nodeID
}

type RaftTransport struct {
	log.AmbientContext
	st *cluster.Settings

	stopper *stop.Stopper

	queues   [rpc.NumConnectionClasses]syncutil.IntMap
	stats    [rpc.NumConnectionClasses]syncutil.IntMap
	dialer   *nodedialer.Dialer
	handlers syncutil.IntMap
}

func NewDummyRaftTransport(st *cluster.Settings, tracer *tracing.Tracer) *RaftTransport {
	__antithesis_instrumentation__.Notify(113030)
	resolver := func(roachpb.NodeID) (net.Addr, error) {
		__antithesis_instrumentation__.Notify(113032)
		return nil, errors.New("dummy resolver")
	}
	__antithesis_instrumentation__.Notify(113031)
	return NewRaftTransport(log.MakeTestingAmbientContext(tracer), st,
		nodedialer.New(nil, resolver), nil, nil)
}

func NewRaftTransport(
	ambient log.AmbientContext,
	st *cluster.Settings,
	dialer *nodedialer.Dialer,
	grpcServer *grpc.Server,
	stopper *stop.Stopper,
) *RaftTransport {
	__antithesis_instrumentation__.Notify(113033)
	t := &RaftTransport{
		AmbientContext: ambient,
		st:             st,

		stopper: stopper,
		dialer:  dialer,
	}

	if grpcServer != nil {
		__antithesis_instrumentation__.Notify(113037)
		RegisterMultiRaftServer(grpcServer, t)
	} else {
		__antithesis_instrumentation__.Notify(113038)
	}
	__antithesis_instrumentation__.Notify(113034)

	statsMap := make(map[roachpb.NodeID]*raftTransportStats)
	clearStatsMap := func() {
		__antithesis_instrumentation__.Notify(113039)
		for k := range statsMap {
			__antithesis_instrumentation__.Notify(113040)
			delete(statsMap, k)
		}
	}
	__antithesis_instrumentation__.Notify(113035)
	if t.stopper != nil && func() bool {
		__antithesis_instrumentation__.Notify(113041)
		return log.V(1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(113042)
		ctx := t.AnnotateCtx(context.Background())
		_ = t.stopper.RunAsyncTask(ctx, "raft-transport", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(113043)
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			lastStats := make(map[roachpb.NodeID]raftTransportStats)
			lastTime := timeutil.Now()
			var stats raftTransportStatsSlice
			for {
				__antithesis_instrumentation__.Notify(113044)
				select {
				case <-ticker.C:
					__antithesis_instrumentation__.Notify(113045)
					stats = stats[:0]
					getStats := func(k int64, v unsafe.Pointer) bool {
						__antithesis_instrumentation__.Notify(113051)
						s := (*raftTransportStats)(v)

						s.queue = 0
						stats = append(stats, s)
						statsMap[roachpb.NodeID(k)] = s
						return true
					}
					__antithesis_instrumentation__.Notify(113046)
					setQueueLength := func(k int64, v unsafe.Pointer) bool {
						__antithesis_instrumentation__.Notify(113052)
						ch := *(*chan *kvserverpb.RaftMessageRequest)(v)
						if s, ok := statsMap[roachpb.NodeID(k)]; ok {
							__antithesis_instrumentation__.Notify(113054)
							s.queue += len(ch)
						} else {
							__antithesis_instrumentation__.Notify(113055)
						}
						__antithesis_instrumentation__.Notify(113053)
						return true
					}
					__antithesis_instrumentation__.Notify(113047)
					for c := range t.stats {
						__antithesis_instrumentation__.Notify(113056)
						clearStatsMap()
						t.stats[c].Range(getStats)
						t.queues[c].Range(setQueueLength)
					}
					__antithesis_instrumentation__.Notify(113048)
					clearStatsMap()

					now := timeutil.Now()
					elapsed := now.Sub(lastTime).Seconds()
					sort.Sort(stats)

					var buf bytes.Buffer

					fmt.Fprintf(&buf,
						"         qlen   qmax   qdropped client-sent client-recv server-sent server-recv\n")
					for _, s := range stats {
						__antithesis_instrumentation__.Notify(113057)
						last := lastStats[s.nodeID]
						cur := raftTransportStats{
							nodeID:        s.nodeID,
							queue:         s.queue,
							queueMax:      atomic.LoadInt32(&s.queueMax),
							clientDropped: atomic.LoadInt64(&s.clientDropped),
							clientSent:    atomic.LoadInt64(&s.clientSent),
							clientRecv:    atomic.LoadInt64(&s.clientRecv),
							serverSent:    atomic.LoadInt64(&s.serverSent),
							serverRecv:    atomic.LoadInt64(&s.serverRecv),
						}
						fmt.Fprintf(&buf, "  %3d: %6d %6d %10d %11.1f %11.1f %11.1f %11.1f\n",
							cur.nodeID, cur.queue, cur.queueMax, cur.clientDropped,
							float64(cur.clientSent-last.clientSent)/elapsed,
							float64(cur.clientRecv-last.clientRecv)/elapsed,
							float64(cur.serverSent-last.serverSent)/elapsed,
							float64(cur.serverRecv-last.serverRecv)/elapsed)
						lastStats[s.nodeID] = cur
					}
					__antithesis_instrumentation__.Notify(113049)
					lastTime = now
					log.Infof(ctx, "stats:\n%s", buf.String())
				case <-t.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(113050)
					return
				}
			}
		})
	} else {
		__antithesis_instrumentation__.Notify(113058)
	}
	__antithesis_instrumentation__.Notify(113036)

	return t
}

func (t *RaftTransport) queuedMessageCount() int64 {
	__antithesis_instrumentation__.Notify(113059)
	var n int64
	addLength := func(k int64, v unsafe.Pointer) bool {
		__antithesis_instrumentation__.Notify(113062)
		ch := *(*chan *kvserverpb.RaftMessageRequest)(v)
		n += int64(len(ch))
		return true
	}
	__antithesis_instrumentation__.Notify(113060)
	for class := range t.queues {
		__antithesis_instrumentation__.Notify(113063)
		t.queues[class].Range(addLength)
	}
	__antithesis_instrumentation__.Notify(113061)
	return n
}

func (t *RaftTransport) getHandler(storeID roachpb.StoreID) (RaftMessageHandler, bool) {
	__antithesis_instrumentation__.Notify(113064)
	if value, ok := t.handlers.Load(int64(storeID)); ok {
		__antithesis_instrumentation__.Notify(113066)
		return *(*RaftMessageHandler)(value), true
	} else {
		__antithesis_instrumentation__.Notify(113067)
	}
	__antithesis_instrumentation__.Notify(113065)
	return nil, false
}

func (t *RaftTransport) handleRaftRequest(
	ctx context.Context, req *kvserverpb.RaftMessageRequest, respStream RaftMessageResponseStream,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(113068)
	handler, ok := t.getHandler(req.ToReplica.StoreID)
	if !ok {
		__antithesis_instrumentation__.Notify(113070)
		log.Warningf(ctx, "unable to accept Raft message from %+v: no handler registered for %+v",
			req.FromReplica, req.ToReplica)
		return roachpb.NewError(roachpb.NewStoreNotFoundError(req.ToReplica.StoreID))
	} else {
		__antithesis_instrumentation__.Notify(113071)
	}
	__antithesis_instrumentation__.Notify(113069)

	return handler.HandleRaftRequest(ctx, req, respStream)
}

func newRaftMessageResponse(
	req *kvserverpb.RaftMessageRequest, pErr *roachpb.Error,
) *kvserverpb.RaftMessageResponse {
	__antithesis_instrumentation__.Notify(113072)
	resp := &kvserverpb.RaftMessageResponse{
		RangeID: req.RangeID,

		ToReplica:   req.FromReplica,
		FromReplica: req.ToReplica,
	}
	if pErr != nil {
		__antithesis_instrumentation__.Notify(113074)
		resp.Union.SetValue(pErr)
	} else {
		__antithesis_instrumentation__.Notify(113075)
	}
	__antithesis_instrumentation__.Notify(113073)
	return resp
}

func (t *RaftTransport) getStats(
	nodeID roachpb.NodeID, class rpc.ConnectionClass,
) *raftTransportStats {
	__antithesis_instrumentation__.Notify(113076)
	statsMap := &t.stats[class]
	value, ok := statsMap.Load(int64(nodeID))
	if !ok {
		__antithesis_instrumentation__.Notify(113078)
		stats := &raftTransportStats{nodeID: nodeID}
		value, _ = statsMap.LoadOrStore(int64(nodeID), unsafe.Pointer(stats))
	} else {
		__antithesis_instrumentation__.Notify(113079)
	}
	__antithesis_instrumentation__.Notify(113077)
	return (*raftTransportStats)(value)
}

func (t *RaftTransport) RaftMessageBatch(stream MultiRaft_RaftMessageBatchServer) error {
	__antithesis_instrumentation__.Notify(113080)
	errCh := make(chan error, 1)

	taskCtx, cancel := t.stopper.WithCancelOnQuiesce(stream.Context())
	defer cancel()
	if err := t.stopper.RunAsyncTaskEx(
		taskCtx,
		stop.TaskOpts{
			TaskName: "storage.RaftTransport: processing batch",
			SpanOpt:  stop.ChildSpan,
		}, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(113082)
			errCh <- func() error {
				__antithesis_instrumentation__.Notify(113083)
				var stats *raftTransportStats
				stream := &lockedRaftMessageResponseStream{wrapped: stream}
				for {
					__antithesis_instrumentation__.Notify(113084)
					batch, err := stream.Recv()
					if err != nil {
						__antithesis_instrumentation__.Notify(113088)
						return err
					} else {
						__antithesis_instrumentation__.Notify(113089)
					}
					__antithesis_instrumentation__.Notify(113085)
					if len(batch.Requests) == 0 {
						__antithesis_instrumentation__.Notify(113090)
						continue
					} else {
						__antithesis_instrumentation__.Notify(113091)
					}
					__antithesis_instrumentation__.Notify(113086)

					if stats == nil {
						__antithesis_instrumentation__.Notify(113092)
						stats = t.getStats(batch.Requests[0].FromReplica.NodeID, rpc.DefaultClass)
					} else {
						__antithesis_instrumentation__.Notify(113093)
					}
					__antithesis_instrumentation__.Notify(113087)

					for i := range batch.Requests {
						__antithesis_instrumentation__.Notify(113094)
						req := &batch.Requests[i]
						atomic.AddInt64(&stats.serverRecv, 1)
						if pErr := t.handleRaftRequest(ctx, req, stream); pErr != nil {
							__antithesis_instrumentation__.Notify(113095)
							atomic.AddInt64(&stats.serverSent, 1)
							if err := stream.Send(newRaftMessageResponse(req, pErr)); err != nil {
								__antithesis_instrumentation__.Notify(113096)
								return err
							} else {
								__antithesis_instrumentation__.Notify(113097)
							}
						} else {
							__antithesis_instrumentation__.Notify(113098)
						}
					}
				}
			}()
		}); err != nil {
		__antithesis_instrumentation__.Notify(113099)
		return err
	} else {
		__antithesis_instrumentation__.Notify(113100)
	}
	__antithesis_instrumentation__.Notify(113081)

	select {
	case err := <-errCh:
		__antithesis_instrumentation__.Notify(113101)
		return err
	case <-t.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(113102)
		return nil
	}
}

func (t *RaftTransport) RaftSnapshot(stream MultiRaft_RaftSnapshotServer) error {
	__antithesis_instrumentation__.Notify(113103)
	errCh := make(chan error, 1)
	taskCtx, cancel := t.stopper.WithCancelOnQuiesce(stream.Context())
	defer cancel()
	if err := t.stopper.RunAsyncTaskEx(
		taskCtx,
		stop.TaskOpts{
			TaskName: "storage.RaftTransport: processing snapshot",
			SpanOpt:  stop.ChildSpan,
		}, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(113105)
			errCh <- func() error {
				__antithesis_instrumentation__.Notify(113106)
				req, err := stream.Recv()
				if err != nil {
					__antithesis_instrumentation__.Notify(113110)
					return err
				} else {
					__antithesis_instrumentation__.Notify(113111)
				}
				__antithesis_instrumentation__.Notify(113107)
				if req.Header == nil {
					__antithesis_instrumentation__.Notify(113112)
					return stream.Send(&kvserverpb.SnapshotResponse{
						Status:  kvserverpb.SnapshotResponse_ERROR,
						Message: "client error: no header in first snapshot request message"})
				} else {
					__antithesis_instrumentation__.Notify(113113)
				}
				__antithesis_instrumentation__.Notify(113108)
				rmr := req.Header.RaftMessageRequest
				handler, ok := t.getHandler(rmr.ToReplica.StoreID)
				if !ok {
					__antithesis_instrumentation__.Notify(113114)
					log.Warningf(ctx, "unable to accept Raft message from %+v: no handler registered for %+v",
						rmr.FromReplica, rmr.ToReplica)
					return roachpb.NewStoreNotFoundError(rmr.ToReplica.StoreID)
				} else {
					__antithesis_instrumentation__.Notify(113115)
				}
				__antithesis_instrumentation__.Notify(113109)
				return handler.HandleSnapshot(ctx, req.Header, stream)
			}()
		}); err != nil {
		__antithesis_instrumentation__.Notify(113116)
		return err
	} else {
		__antithesis_instrumentation__.Notify(113117)
	}
	__antithesis_instrumentation__.Notify(113104)
	select {
	case <-t.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(113118)
		return nil
	case err := <-errCh:
		__antithesis_instrumentation__.Notify(113119)
		return err
	}
}

func (t *RaftTransport) Listen(storeID roachpb.StoreID, handler RaftMessageHandler) {
	__antithesis_instrumentation__.Notify(113120)
	t.handlers.Store(int64(storeID), unsafe.Pointer(&handler))
}

func (t *RaftTransport) Stop(storeID roachpb.StoreID) {
	__antithesis_instrumentation__.Notify(113121)
	t.handlers.Delete(int64(storeID))
}

func (t *RaftTransport) processQueue(
	nodeID roachpb.NodeID,
	ch chan *kvserverpb.RaftMessageRequest,
	stats *raftTransportStats,
	stream MultiRaft_RaftMessageBatchClient,
	class rpc.ConnectionClass,
) error {
	__antithesis_instrumentation__.Notify(113122)
	errCh := make(chan error, 1)

	ctx := stream.Context()

	if err := t.stopper.RunAsyncTask(
		ctx, "storage.RaftTransport: processing queue",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(113124)
			errCh <- func() error {
				__antithesis_instrumentation__.Notify(113125)
				for {
					__antithesis_instrumentation__.Notify(113126)
					resp, err := stream.Recv()
					if err != nil {
						__antithesis_instrumentation__.Notify(113129)
						return err
					} else {
						__antithesis_instrumentation__.Notify(113130)
					}
					__antithesis_instrumentation__.Notify(113127)
					atomic.AddInt64(&stats.clientRecv, 1)
					handler, ok := t.getHandler(resp.ToReplica.StoreID)
					if !ok {
						__antithesis_instrumentation__.Notify(113131)
						log.Warningf(ctx, "no handler found for store %s in response %s",
							resp.ToReplica.StoreID, resp)
						continue
					} else {
						__antithesis_instrumentation__.Notify(113132)
					}
					__antithesis_instrumentation__.Notify(113128)
					if err := handler.HandleRaftResponse(ctx, resp); err != nil {
						__antithesis_instrumentation__.Notify(113133)
						return err
					} else {
						__antithesis_instrumentation__.Notify(113134)
					}
				}
			}()
		}); err != nil {
		__antithesis_instrumentation__.Notify(113135)
		return err
	} else {
		__antithesis_instrumentation__.Notify(113136)
	}
	__antithesis_instrumentation__.Notify(113123)

	var raftIdleTimer timeutil.Timer
	defer raftIdleTimer.Stop()
	batch := &kvserverpb.RaftMessageRequestBatch{}
	for {
		__antithesis_instrumentation__.Notify(113137)
		raftIdleTimer.Reset(raftIdleTimeout)
		select {
		case <-t.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(113138)
			return nil
		case <-raftIdleTimer.C:
			__antithesis_instrumentation__.Notify(113139)
			raftIdleTimer.Read = true
			return nil
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(113140)
			return err
		case req := <-ch:
			__antithesis_instrumentation__.Notify(113141)
			budget := targetRaftOutgoingBatchSize.Get(&t.st.SV) - int64(req.Size())
			batch.Requests = append(batch.Requests, *req)
			releaseRaftMessageRequest(req)

			for budget > 0 {
				__antithesis_instrumentation__.Notify(113145)
				select {
				case req = <-ch:
					__antithesis_instrumentation__.Notify(113146)
					budget -= int64(req.Size())
					batch.Requests = append(batch.Requests, *req)
					releaseRaftMessageRequest(req)
				default:
					__antithesis_instrumentation__.Notify(113147)
					budget = -1
				}
			}
			__antithesis_instrumentation__.Notify(113142)

			err := stream.Send(batch)
			if err != nil {
				__antithesis_instrumentation__.Notify(113148)
				return err
			} else {
				__antithesis_instrumentation__.Notify(113149)
			}
			__antithesis_instrumentation__.Notify(113143)

			for i := range batch.Requests {
				__antithesis_instrumentation__.Notify(113150)
				batch.Requests[i] = kvserverpb.RaftMessageRequest{}
			}
			__antithesis_instrumentation__.Notify(113144)
			batch.Requests = batch.Requests[:0]

			atomic.AddInt64(&stats.clientSent, 1)
		}
	}
}

func (t *RaftTransport) getQueue(
	nodeID roachpb.NodeID, class rpc.ConnectionClass,
) (chan *kvserverpb.RaftMessageRequest, bool) {
	__antithesis_instrumentation__.Notify(113151)
	queuesMap := &t.queues[class]
	value, ok := queuesMap.Load(int64(nodeID))
	if !ok {
		__antithesis_instrumentation__.Notify(113153)
		ch := make(chan *kvserverpb.RaftMessageRequest, raftSendBufferSize)
		value, ok = queuesMap.LoadOrStore(int64(nodeID), unsafe.Pointer(&ch))
	} else {
		__antithesis_instrumentation__.Notify(113154)
	}
	__antithesis_instrumentation__.Notify(113152)
	return *(*chan *kvserverpb.RaftMessageRequest)(value), ok
}

func (t *RaftTransport) SendAsync(
	req *kvserverpb.RaftMessageRequest, class rpc.ConnectionClass,
) (sent bool) {
	__antithesis_instrumentation__.Notify(113155)
	toNodeID := req.ToReplica.NodeID
	stats := t.getStats(toNodeID, class)
	defer func() {
		__antithesis_instrumentation__.Notify(113161)
		if !sent {
			__antithesis_instrumentation__.Notify(113162)
			atomic.AddInt64(&stats.clientDropped, 1)
		} else {
			__antithesis_instrumentation__.Notify(113163)
		}
	}()
	__antithesis_instrumentation__.Notify(113156)

	if req.RangeID == 0 && func() bool {
		__antithesis_instrumentation__.Notify(113164)
		return len(req.Heartbeats) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(113165)
		return len(req.HeartbeatResps) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(113166)

		panic("only messages with coalesced heartbeats or heartbeat responses may be sent to range ID 0")
	} else {
		__antithesis_instrumentation__.Notify(113167)
	}
	__antithesis_instrumentation__.Notify(113157)
	if req.Message.Type == raftpb.MsgSnap {
		__antithesis_instrumentation__.Notify(113168)
		panic("snapshots must be sent using SendSnapshot")
	} else {
		__antithesis_instrumentation__.Notify(113169)
	}
	__antithesis_instrumentation__.Notify(113158)

	if !t.dialer.GetCircuitBreaker(toNodeID, class).Ready() {
		__antithesis_instrumentation__.Notify(113170)
		return false
	} else {
		__antithesis_instrumentation__.Notify(113171)
	}
	__antithesis_instrumentation__.Notify(113159)

	ch, existingQueue := t.getQueue(toNodeID, class)
	if !existingQueue {
		__antithesis_instrumentation__.Notify(113172)

		ctx := t.AnnotateCtx(context.Background())
		if !t.startProcessNewQueue(ctx, toNodeID, class, stats) {
			__antithesis_instrumentation__.Notify(113173)
			return false
		} else {
			__antithesis_instrumentation__.Notify(113174)
		}
	} else {
		__antithesis_instrumentation__.Notify(113175)
	}
	__antithesis_instrumentation__.Notify(113160)

	select {
	case ch <- req:
		__antithesis_instrumentation__.Notify(113176)
		l := int32(len(ch))
		if v := atomic.LoadInt32(&stats.queueMax); v < l {
			__antithesis_instrumentation__.Notify(113179)
			atomic.CompareAndSwapInt32(&stats.queueMax, v, l)
		} else {
			__antithesis_instrumentation__.Notify(113180)
		}
		__antithesis_instrumentation__.Notify(113177)
		return true
	default:
		__antithesis_instrumentation__.Notify(113178)
		releaseRaftMessageRequest(req)
		return false
	}
}

func (t *RaftTransport) startProcessNewQueue(
	ctx context.Context,
	toNodeID roachpb.NodeID,
	class rpc.ConnectionClass,
	stats *raftTransportStats,
) (started bool) {
	__antithesis_instrumentation__.Notify(113181)
	cleanup := func(ch chan *kvserverpb.RaftMessageRequest) {
		__antithesis_instrumentation__.Notify(113185)

		for {
			__antithesis_instrumentation__.Notify(113186)
			select {
			case <-ch:
				__antithesis_instrumentation__.Notify(113187)
				atomic.AddInt64(&stats.clientDropped, 1)
			default:
				__antithesis_instrumentation__.Notify(113188)
				return
			}
		}
	}
	__antithesis_instrumentation__.Notify(113182)
	worker := func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(113189)
		ch, existingQueue := t.getQueue(toNodeID, class)
		if !existingQueue {
			__antithesis_instrumentation__.Notify(113193)
			log.Fatalf(ctx, "queue for n%d does not exist", toNodeID)
		} else {
			__antithesis_instrumentation__.Notify(113194)
		}
		__antithesis_instrumentation__.Notify(113190)
		defer cleanup(ch)
		defer t.queues[class].Delete(int64(toNodeID))

		conn, err := t.dialer.DialNoBreaker(ctx, toNodeID, class)
		if err != nil {
			__antithesis_instrumentation__.Notify(113195)

			return
		} else {
			__antithesis_instrumentation__.Notify(113196)
		}
		__antithesis_instrumentation__.Notify(113191)

		client := NewMultiRaftClient(conn)
		batchCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		stream, err := client.RaftMessageBatch(batchCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(113197)
			log.Warningf(ctx, "creating batch client for node %d failed: %+v", toNodeID, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(113198)
		}
		__antithesis_instrumentation__.Notify(113192)

		if err := t.processQueue(toNodeID, ch, stats, stream, class); err != nil {
			__antithesis_instrumentation__.Notify(113199)
			log.Warningf(ctx, "while processing outgoing Raft queue to node %d: %s:", toNodeID, err)
		} else {
			__antithesis_instrumentation__.Notify(113200)
		}
	}
	__antithesis_instrumentation__.Notify(113183)
	err := t.stopper.RunAsyncTask(ctx, "storage.RaftTransport: sending messages", worker)
	if err != nil {
		__antithesis_instrumentation__.Notify(113201)
		t.queues[class].Delete(int64(toNodeID))
		return false
	} else {
		__antithesis_instrumentation__.Notify(113202)
	}
	__antithesis_instrumentation__.Notify(113184)
	return true
}

func (t *RaftTransport) SendSnapshot(
	ctx context.Context,
	storePool *StorePool,
	header kvserverpb.SnapshotRequest_Header,
	snap *OutgoingSnapshot,
	newBatch func() storage.Batch,
	sent func(),
	bytesSentCounter *metric.Counter,
) error {
	__antithesis_instrumentation__.Notify(113203)
	nodeID := header.RaftMessageRequest.ToReplica.NodeID

	conn, err := t.dialer.Dial(ctx, nodeID, rpc.DefaultClass)
	if err != nil {
		__antithesis_instrumentation__.Notify(113207)
		return err
	} else {
		__antithesis_instrumentation__.Notify(113208)
	}
	__antithesis_instrumentation__.Notify(113204)
	client := NewMultiRaftClient(conn)
	stream, err := client.RaftSnapshot(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(113209)
		return err
	} else {
		__antithesis_instrumentation__.Notify(113210)
	}
	__antithesis_instrumentation__.Notify(113205)

	defer func() {
		__antithesis_instrumentation__.Notify(113211)
		if err := stream.CloseSend(); err != nil {
			__antithesis_instrumentation__.Notify(113212)
			log.Warningf(ctx, "failed to close snapshot stream: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(113213)
		}
	}()
	__antithesis_instrumentation__.Notify(113206)
	return sendSnapshot(ctx, t.st, stream, storePool, header, snap, newBatch, sent, bytesSentCounter)
}
