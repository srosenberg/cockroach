package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type serverInfo struct {
	createdAt time.Time
	peerID    roachpb.NodeID
}

type server struct {
	log.AmbientContext

	clusterID *base.ClusterIDContainer
	NodeID    *base.NodeIDContainer

	stopper *stop.Stopper

	mu struct {
		syncutil.RWMutex
		is       *infoStore
		incoming nodeSet
		nodeMap  map[util.UnresolvedAddr]serverInfo

		ready chan struct{}
	}
	tighten chan struct{}

	nodeMetrics   Metrics
	serverMetrics Metrics

	simulationCycler *sync.Cond
}

func newServer(
	ambient log.AmbientContext,
	clusterID *base.ClusterIDContainer,
	nodeID *base.NodeIDContainer,
	stopper *stop.Stopper,
	registry *metric.Registry,
) *server {
	__antithesis_instrumentation__.Notify(68013)
	s := &server{
		AmbientContext: ambient,
		clusterID:      clusterID,
		NodeID:         nodeID,
		stopper:        stopper,
		tighten:        make(chan struct{}, 1),
		nodeMetrics:    makeMetrics(),
		serverMetrics:  makeMetrics(),
	}

	s.mu.is = newInfoStore(s.AmbientContext, nodeID, util.UnresolvedAddr{}, stopper)
	s.mu.incoming = makeNodeSet(minPeers, metric.NewGauge(MetaConnectionsIncomingGauge))
	s.mu.nodeMap = make(map[util.UnresolvedAddr]serverInfo)
	s.mu.ready = make(chan struct{})

	registry.AddMetric(s.mu.incoming.gauge)
	registry.AddMetricStruct(s.nodeMetrics)

	return s
}

func (s *server) GetNodeMetrics() *Metrics {
	__antithesis_instrumentation__.Notify(68014)
	return &s.nodeMetrics
}

func (s *server) Gossip(stream Gossip_GossipServer) error {
	__antithesis_instrumentation__.Notify(68015)
	args, err := stream.Recv()
	if err != nil {
		__antithesis_instrumentation__.Notify(68021)
		return err
	} else {
		__antithesis_instrumentation__.Notify(68022)
	}
	__antithesis_instrumentation__.Notify(68016)
	if (args.ClusterID != uuid.UUID{}) && func() bool {
		__antithesis_instrumentation__.Notify(68023)
		return args.ClusterID != s.clusterID.Get() == true
	}() == true {
		__antithesis_instrumentation__.Notify(68024)
		return errors.Errorf("gossip connection refused from different cluster %s", args.ClusterID)
	} else {
		__antithesis_instrumentation__.Notify(68025)
	}
	__antithesis_instrumentation__.Notify(68017)

	ctx, cancel := context.WithCancel(s.AnnotateCtx(stream.Context()))
	defer cancel()
	syncChan := make(chan struct{}, 1)
	send := func(reply *Response) error {
		__antithesis_instrumentation__.Notify(68026)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(68027)
			return ctx.Err()
		case syncChan <- struct{}{}:
			__antithesis_instrumentation__.Notify(68028)
			defer func() { __antithesis_instrumentation__.Notify(68030); <-syncChan }()
			__antithesis_instrumentation__.Notify(68029)

			bytesSent := int64(reply.Size())
			infoCount := int64(len(reply.Delta))
			s.nodeMetrics.BytesSent.Inc(bytesSent)
			s.nodeMetrics.InfosSent.Inc(infoCount)
			s.serverMetrics.BytesSent.Inc(bytesSent)
			s.serverMetrics.InfosSent.Inc(infoCount)

			return stream.Send(reply)
		}
	}
	__antithesis_instrumentation__.Notify(68018)

	defer func() { __antithesis_instrumentation__.Notify(68031); syncChan <- struct{}{} }()
	__antithesis_instrumentation__.Notify(68019)

	errCh := make(chan error, 1)

	if err := s.stopper.RunAsyncTask(ctx, "gossip receiver", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(68032)
		errCh <- s.gossipReceiver(ctx, &args, send, stream.Recv)
	}); err != nil {
		__antithesis_instrumentation__.Notify(68033)
		return err
	} else {
		__antithesis_instrumentation__.Notify(68034)
	}
	__antithesis_instrumentation__.Notify(68020)

	reply := new(Response)

	for init := true; ; init = false {
		__antithesis_instrumentation__.Notify(68035)
		s.mu.Lock()

		ready := s.mu.ready
		delta := s.mu.is.delta(args.HighWaterStamps)
		if init {
			__antithesis_instrumentation__.Notify(68039)
			s.mu.is.populateMostDistantMarkers(delta)
		} else {
			__antithesis_instrumentation__.Notify(68040)
		}
		__antithesis_instrumentation__.Notify(68036)
		if args.HighWaterStamps == nil {
			__antithesis_instrumentation__.Notify(68041)
			args.HighWaterStamps = make(map[roachpb.NodeID]int64)
		} else {
			__antithesis_instrumentation__.Notify(68042)
		}
		__antithesis_instrumentation__.Notify(68037)

		if infoCount := len(delta); init || func() bool {
			__antithesis_instrumentation__.Notify(68043)
			return infoCount > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(68044)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(68048)
				log.Infof(ctx, "returning %d info(s) to n%d: %s",
					infoCount, args.NodeID, extractKeys(delta))
			} else {
				__antithesis_instrumentation__.Notify(68049)
			}
			__antithesis_instrumentation__.Notify(68045)

			for _, i := range delta {
				__antithesis_instrumentation__.Notify(68050)
				ratchetHighWaterStamp(args.HighWaterStamps, i.NodeID, i.OrigStamp)
			}
			__antithesis_instrumentation__.Notify(68046)

			*reply = Response{
				NodeID:          s.NodeID.Get(),
				HighWaterStamps: s.mu.is.getHighWaterStamps(),
				Delta:           delta,
			}

			s.mu.Unlock()
			if err := send(reply); err != nil {
				__antithesis_instrumentation__.Notify(68051)
				return err
			} else {
				__antithesis_instrumentation__.Notify(68052)
			}
			__antithesis_instrumentation__.Notify(68047)
			s.mu.Lock()
		} else {
			__antithesis_instrumentation__.Notify(68053)
		}
		__antithesis_instrumentation__.Notify(68038)

		s.mu.Unlock()

		select {
		case <-s.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(68054)
			return nil
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(68055)
			return err
		case <-ready:
			__antithesis_instrumentation__.Notify(68056)
		}
	}
}

func (s *server) gossipReceiver(
	ctx context.Context,
	argsPtr **Request,
	senderFn func(*Response) error,
	receiverFn func() (*Request, error),
) error {
	__antithesis_instrumentation__.Notify(68057)
	s.mu.Lock()
	defer s.mu.Unlock()

	reply := new(Response)

	nodeIdentified := false

	for {
		__antithesis_instrumentation__.Notify(68058)
		args := *argsPtr
		if args.NodeID == 0 {
			__antithesis_instrumentation__.Notify(68065)

			log.Infof(ctx, "received initial cluster-verification connection from %s", args.Addr)
		} else {
			__antithesis_instrumentation__.Notify(68066)
			if !nodeIdentified {
				__antithesis_instrumentation__.Notify(68067)
				nodeIdentified = true

				if args.NodeID == s.NodeID.Get() {
					__antithesis_instrumentation__.Notify(68068)

					if log.V(2) {
						__antithesis_instrumentation__.Notify(68069)
						log.Infof(ctx, "ignoring gossip from n%d (loopback)", args.NodeID)
					} else {
						__antithesis_instrumentation__.Notify(68070)
					}
				} else {
					__antithesis_instrumentation__.Notify(68071)
					if _, ok := s.mu.nodeMap[args.Addr]; ok {
						__antithesis_instrumentation__.Notify(68072)

						if log.V(2) {
							__antithesis_instrumentation__.Notify(68074)
							log.Infof(ctx, "duplicate connection received from n%d at %s", args.NodeID, args.Addr)
						} else {
							__antithesis_instrumentation__.Notify(68075)
						}
						__antithesis_instrumentation__.Notify(68073)
						return errors.Errorf("duplicate connection from node at %s", args.Addr)
					} else {
						__antithesis_instrumentation__.Notify(68076)
						if s.mu.incoming.hasSpace() {
							__antithesis_instrumentation__.Notify(68077)
							log.VEventf(ctx, 2, "adding n%d to incoming set", args.NodeID)

							s.mu.incoming.addNode(args.NodeID)
							s.mu.nodeMap[args.Addr] = serverInfo{
								peerID:    args.NodeID,
								createdAt: timeutil.Now(),
							}

							defer func(nodeID roachpb.NodeID, addr util.UnresolvedAddr) {
								__antithesis_instrumentation__.Notify(68078)
								log.VEventf(ctx, 2, "removing n%d from incoming set", args.NodeID)
								s.mu.incoming.removeNode(nodeID)
								delete(s.mu.nodeMap, addr)
							}(args.NodeID, args.Addr)
						} else {
							__antithesis_instrumentation__.Notify(68079)

							var alternateAddr util.UnresolvedAddr
							var alternateNodeID roachpb.NodeID

							altIdx := rand.Intn(len(s.mu.nodeMap))
							for addr, info := range s.mu.nodeMap {
								__antithesis_instrumentation__.Notify(68081)
								if altIdx == 0 {
									__antithesis_instrumentation__.Notify(68083)
									alternateAddr = addr
									alternateNodeID = info.peerID
									break
								} else {
									__antithesis_instrumentation__.Notify(68084)
								}
								__antithesis_instrumentation__.Notify(68082)
								altIdx--
							}
							__antithesis_instrumentation__.Notify(68080)

							s.nodeMetrics.ConnectionsRefused.Inc(1)
							log.Infof(ctx, "refusing gossip from n%d (max %d conns); forwarding to n%d (%s)",
								args.NodeID, s.mu.incoming.maxSize, alternateNodeID, alternateAddr)

							*reply = Response{
								NodeID:          s.NodeID.Get(),
								AlternateAddr:   &alternateAddr,
								AlternateNodeID: alternateNodeID,
							}

							s.mu.Unlock()
							err := senderFn(reply)
							s.mu.Lock()

							if err != nil {
								__antithesis_instrumentation__.Notify(68085)
								return err
							} else {
								__antithesis_instrumentation__.Notify(68086)
							}
						}
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(68087)
			}
		}
		__antithesis_instrumentation__.Notify(68059)

		bytesReceived := int64(args.Size())
		infosReceived := int64(len(args.Delta))
		s.nodeMetrics.BytesReceived.Inc(bytesReceived)
		s.nodeMetrics.InfosReceived.Inc(infosReceived)
		s.serverMetrics.BytesReceived.Inc(bytesReceived)
		s.serverMetrics.InfosReceived.Inc(infosReceived)

		freshCount, err := s.mu.is.combine(args.Delta, args.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(68088)
			log.Warningf(ctx, "failed to fully combine gossip delta from n%d: %s", args.NodeID, err)
		} else {
			__antithesis_instrumentation__.Notify(68089)
		}
		__antithesis_instrumentation__.Notify(68060)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(68090)
			log.Infof(ctx, "received %s from n%d (%d fresh)", extractKeys(args.Delta), args.NodeID, freshCount)
		} else {
			__antithesis_instrumentation__.Notify(68091)
		}
		__antithesis_instrumentation__.Notify(68061)
		s.maybeTightenLocked()

		*reply = Response{
			NodeID:          s.NodeID.Get(),
			HighWaterStamps: s.mu.is.getHighWaterStamps(),
		}

		s.mu.Unlock()
		err = senderFn(reply)
		s.mu.Lock()
		if err != nil {
			__antithesis_instrumentation__.Notify(68092)
			return err
		} else {
			__antithesis_instrumentation__.Notify(68093)
		}
		__antithesis_instrumentation__.Notify(68062)

		if cycler := s.simulationCycler; cycler != nil {
			__antithesis_instrumentation__.Notify(68094)
			cycler.Wait()
		} else {
			__antithesis_instrumentation__.Notify(68095)
		}
		__antithesis_instrumentation__.Notify(68063)

		s.mu.Unlock()
		recvArgs, err := receiverFn()
		s.mu.Lock()
		if err != nil {
			__antithesis_instrumentation__.Notify(68096)
			return err
		} else {
			__antithesis_instrumentation__.Notify(68097)
		}
		__antithesis_instrumentation__.Notify(68064)

		mergeHighWaterStamps(&recvArgs.HighWaterStamps, (*argsPtr).HighWaterStamps)
		*argsPtr = recvArgs
	}
}

func (s *server) maybeTightenLocked() {
	__antithesis_instrumentation__.Notify(68098)
	select {
	case s.tighten <- struct{}{}:
		__antithesis_instrumentation__.Notify(68099)
	default:
		__antithesis_instrumentation__.Notify(68100)
	}
}

func (s *server) start(addr net.Addr) {
	__antithesis_instrumentation__.Notify(68101)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.is.NodeAddr = util.MakeUnresolvedAddr(addr.Network(), addr.String())

	broadcast := func() {
		__antithesis_instrumentation__.Notify(68105)

		s.mu.Lock()
		defer s.mu.Unlock()
		ready := make(chan struct{})
		close(s.mu.ready)
		s.mu.ready = ready
	}
	__antithesis_instrumentation__.Notify(68102)

	unregister := s.mu.is.registerCallback(".*", func(_ string, _ roachpb.Value) {
		__antithesis_instrumentation__.Notify(68106)
		broadcast()
	}, Redundant)
	__antithesis_instrumentation__.Notify(68103)

	waitQuiesce := func(context.Context) {
		__antithesis_instrumentation__.Notify(68107)
		<-s.stopper.ShouldQuiesce()

		s.mu.Lock()
		unregister()
		s.mu.Unlock()

		broadcast()
	}
	__antithesis_instrumentation__.Notify(68104)
	bgCtx := s.AnnotateCtx(context.Background())
	if err := s.stopper.RunAsyncTask(bgCtx, "gossip-wait-quiesce", waitQuiesce); err != nil {
		__antithesis_instrumentation__.Notify(68108)
		waitQuiesce(bgCtx)
	} else {
		__antithesis_instrumentation__.Notify(68109)
	}
}

func (s *server) status() ServerStatus {
	__antithesis_instrumentation__.Notify(68110)
	s.mu.RLock()
	defer s.mu.RUnlock()

	var status ServerStatus
	status.ConnStatus = make([]ConnStatus, 0, len(s.mu.nodeMap))
	status.MaxConns = int32(s.mu.incoming.maxSize)
	status.MetricSnap = s.serverMetrics.Snapshot()

	for addr, info := range s.mu.nodeMap {
		__antithesis_instrumentation__.Notify(68112)
		status.ConnStatus = append(status.ConnStatus, ConnStatus{
			NodeID:   info.peerID,
			Address:  addr.String(),
			AgeNanos: timeutil.Since(info.createdAt).Nanoseconds(),
		})
	}
	__antithesis_instrumentation__.Notify(68111)
	return status
}

func roundSecs(d time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(68113)
	return time.Duration(d.Seconds()+0.5) * time.Second
}

func (s *server) GetNodeAddr() *util.UnresolvedAddr {
	__antithesis_instrumentation__.Notify(68114)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &s.mu.is.NodeAddr
}
