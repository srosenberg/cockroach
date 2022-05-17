package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	circuit "github.com/cockroachdb/circuitbreaker"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type client struct {
	log.AmbientContext

	createdAt             time.Time
	peerID                roachpb.NodeID
	resolvedPlaceholder   bool
	addr                  net.Addr
	forwardAddr           *util.UnresolvedAddr
	remoteHighWaterStamps map[roachpb.NodeID]int64
	closer                chan struct{}
	clientMetrics         Metrics
	nodeMetrics           Metrics
}

func extractKeys(delta map[string]*Info) string {
	__antithesis_instrumentation__.Notify(65077)
	keys := make([]string, 0, len(delta))
	for key := range delta {
		__antithesis_instrumentation__.Notify(65079)
		keys = append(keys, key)
	}
	__antithesis_instrumentation__.Notify(65078)
	return fmt.Sprintf("%s", keys)
}

func newClient(ambient log.AmbientContext, addr net.Addr, nodeMetrics Metrics) *client {
	__antithesis_instrumentation__.Notify(65080)
	return &client{
		AmbientContext:        ambient,
		createdAt:             timeutil.Now(),
		addr:                  addr,
		remoteHighWaterStamps: map[roachpb.NodeID]int64{},
		closer:                make(chan struct{}),
		clientMetrics:         makeMetrics(),
		nodeMetrics:           nodeMetrics,
	}
}

func (c *client) startLocked(
	g *Gossip,
	disconnected chan *client,
	rpcCtx *rpc.Context,
	stopper *stop.Stopper,
	breaker *circuit.Breaker,
) {
	__antithesis_instrumentation__.Notify(65081)

	g.outgoing.addPlaceholder()

	ctx, cancel := context.WithCancel(c.AnnotateCtx(context.Background()))
	if err := stopper.RunAsyncTask(ctx, "gossip-client", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(65082)
		var wg sync.WaitGroup
		defer func() {
			__antithesis_instrumentation__.Notify(65085)

			cancel()

			wg.Wait()
			disconnected <- c
		}()
		__antithesis_instrumentation__.Notify(65083)

		consecFailures := breaker.ConsecFailures()
		var stream Gossip_GossipClient
		if err := breaker.Call(func() error {
			__antithesis_instrumentation__.Notify(65086)

			conn, err := rpcCtx.GRPCUnvalidatedDial(c.addr.String()).Connect(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(65089)
				return err
			} else {
				__antithesis_instrumentation__.Notify(65090)
			}
			__antithesis_instrumentation__.Notify(65087)
			if stream, err = NewGossipClient(conn).Gossip(ctx); err != nil {
				__antithesis_instrumentation__.Notify(65091)
				return err
			} else {
				__antithesis_instrumentation__.Notify(65092)
			}
			__antithesis_instrumentation__.Notify(65088)
			return c.requestGossip(g, stream)
		}, 0); err != nil {
			__antithesis_instrumentation__.Notify(65093)
			if consecFailures == 0 {
				__antithesis_instrumentation__.Notify(65095)
				log.Warningf(ctx, "failed to start gossip client to %s: %s", c.addr, err)
			} else {
				__antithesis_instrumentation__.Notify(65096)
			}
			__antithesis_instrumentation__.Notify(65094)
			return
		} else {
			__antithesis_instrumentation__.Notify(65097)
		}
		__antithesis_instrumentation__.Notify(65084)

		log.Infof(ctx, "started gossip client to n%d (%s)", c.peerID, c.addr)
		if err := c.gossip(ctx, g, stream, stopper, &wg); err != nil {
			__antithesis_instrumentation__.Notify(65098)
			if !grpcutil.IsClosedConnection(err) {
				__antithesis_instrumentation__.Notify(65099)
				g.mu.RLock()
				if c.peerID != 0 {
					__antithesis_instrumentation__.Notify(65101)
					log.Infof(ctx, "closing client to n%d (%s): %s", c.peerID, c.addr, err)
				} else {
					__antithesis_instrumentation__.Notify(65102)
					log.Infof(ctx, "closing client to %s: %s", c.addr, err)
				}
				__antithesis_instrumentation__.Notify(65100)
				g.mu.RUnlock()
			} else {
				__antithesis_instrumentation__.Notify(65103)
			}
		} else {
			__antithesis_instrumentation__.Notify(65104)
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(65105)
		disconnected <- c
	} else {
		__antithesis_instrumentation__.Notify(65106)
	}
}

func (c *client) close() {
	__antithesis_instrumentation__.Notify(65107)
	select {
	case <-c.closer:
		__antithesis_instrumentation__.Notify(65108)
	default:
		__antithesis_instrumentation__.Notify(65109)
		close(c.closer)
	}
}

func (c *client) requestGossip(g *Gossip, stream Gossip_GossipClient) error {
	__antithesis_instrumentation__.Notify(65110)
	g.mu.RLock()
	args := &Request{
		NodeID:          g.NodeID.Get(),
		Addr:            g.mu.is.NodeAddr,
		HighWaterStamps: g.mu.is.getHighWaterStamps(),
		ClusterID:       g.clusterID.Get(),
	}
	g.mu.RUnlock()

	bytesSent := int64(args.Size())
	c.clientMetrics.BytesSent.Inc(bytesSent)
	c.nodeMetrics.BytesSent.Inc(bytesSent)

	return stream.Send(args)
}

func (c *client) sendGossip(g *Gossip, stream Gossip_GossipClient, firstReq bool) error {
	__antithesis_instrumentation__.Notify(65111)
	g.mu.Lock()
	delta := g.mu.is.delta(c.remoteHighWaterStamps)
	if firstReq {
		__antithesis_instrumentation__.Notify(65114)
		g.mu.is.populateMostDistantMarkers(delta)
	} else {
		__antithesis_instrumentation__.Notify(65115)
	}
	__antithesis_instrumentation__.Notify(65112)
	if len(delta) > 0 {
		__antithesis_instrumentation__.Notify(65116)

		for _, i := range delta {
			__antithesis_instrumentation__.Notify(65119)
			ratchetHighWaterStamp(c.remoteHighWaterStamps, i.NodeID, i.OrigStamp)
		}
		__antithesis_instrumentation__.Notify(65117)

		args := Request{
			NodeID:          g.NodeID.Get(),
			Addr:            g.mu.is.NodeAddr,
			Delta:           delta,
			HighWaterStamps: g.mu.is.getHighWaterStamps(),
			ClusterID:       g.clusterID.Get(),
		}

		bytesSent := int64(args.Size())
		infosSent := int64(len(delta))
		c.clientMetrics.BytesSent.Inc(bytesSent)
		c.clientMetrics.InfosSent.Inc(infosSent)
		c.nodeMetrics.BytesSent.Inc(bytesSent)
		c.nodeMetrics.InfosSent.Inc(infosSent)

		if log.V(1) {
			__antithesis_instrumentation__.Notify(65120)
			ctx := c.AnnotateCtx(stream.Context())
			if c.peerID != 0 {
				__antithesis_instrumentation__.Notify(65121)
				log.Infof(ctx, "sending %s to n%d (%s)", extractKeys(args.Delta), c.peerID, c.addr)
			} else {
				__antithesis_instrumentation__.Notify(65122)
				log.Infof(ctx, "sending %s to %s", extractKeys(args.Delta), c.addr)
			}
		} else {
			__antithesis_instrumentation__.Notify(65123)
		}
		__antithesis_instrumentation__.Notify(65118)

		g.mu.Unlock()
		return stream.Send(&args)
	} else {
		__antithesis_instrumentation__.Notify(65124)
	}
	__antithesis_instrumentation__.Notify(65113)
	g.mu.Unlock()
	return nil
}

func (c *client) handleResponse(ctx context.Context, g *Gossip, reply *Response) error {
	__antithesis_instrumentation__.Notify(65125)
	g.mu.Lock()
	defer g.mu.Unlock()

	bytesReceived := int64(reply.Size())
	infosReceived := int64(len(reply.Delta))
	c.clientMetrics.BytesReceived.Inc(bytesReceived)
	c.clientMetrics.InfosReceived.Inc(infosReceived)
	c.nodeMetrics.BytesReceived.Inc(bytesReceived)
	c.nodeMetrics.InfosReceived.Inc(infosReceived)

	if reply.Delta != nil {
		__antithesis_instrumentation__.Notify(65130)
		freshCount, err := g.mu.is.combine(reply.Delta, reply.NodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(65133)
			log.Warningf(ctx, "failed to fully combine delta from n%d: %s", reply.NodeID, err)
		} else {
			__antithesis_instrumentation__.Notify(65134)
		}
		__antithesis_instrumentation__.Notify(65131)
		if infoCount := len(reply.Delta); infoCount > 0 {
			__antithesis_instrumentation__.Notify(65135)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(65136)
				log.Infof(ctx, "received %s from n%d (%d fresh)", extractKeys(reply.Delta), reply.NodeID, freshCount)
			} else {
				__antithesis_instrumentation__.Notify(65137)
			}
		} else {
			__antithesis_instrumentation__.Notify(65138)
		}
		__antithesis_instrumentation__.Notify(65132)
		g.maybeTightenLocked()
	} else {
		__antithesis_instrumentation__.Notify(65139)
	}
	__antithesis_instrumentation__.Notify(65126)
	c.peerID = reply.NodeID
	mergeHighWaterStamps(&c.remoteHighWaterStamps, reply.HighWaterStamps)

	if !c.resolvedPlaceholder && func() bool {
		__antithesis_instrumentation__.Notify(65140)
		return c.peerID != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(65141)
		c.resolvedPlaceholder = true
		g.outgoing.resolvePlaceholder(c.peerID)
	} else {
		__antithesis_instrumentation__.Notify(65142)
	}
	__antithesis_instrumentation__.Notify(65127)

	if reply.AlternateAddr != nil {
		__antithesis_instrumentation__.Notify(65143)
		if g.hasIncomingLocked(reply.AlternateNodeID) || func() bool {
			__antithesis_instrumentation__.Notify(65146)
			return g.hasOutgoingLocked(reply.AlternateNodeID) == true
		}() == true {
			__antithesis_instrumentation__.Notify(65147)
			return errors.Errorf(
				"received forward from n%d to n%d (%s); already have active connection, skipping",
				reply.NodeID, reply.AlternateNodeID, reply.AlternateAddr)
		} else {
			__antithesis_instrumentation__.Notify(65148)
		}
		__antithesis_instrumentation__.Notify(65144)

		if _, err := reply.AlternateAddr.Resolve(); err != nil {
			__antithesis_instrumentation__.Notify(65149)
			return errors.Wrapf(err, "unable to resolve alternate address %s for n%d",
				reply.AlternateAddr, reply.AlternateNodeID)
		} else {
			__antithesis_instrumentation__.Notify(65150)
		}
		__antithesis_instrumentation__.Notify(65145)
		c.forwardAddr = reply.AlternateAddr
		return errors.Errorf("received forward from n%d to n%d (%s)",
			reply.NodeID, reply.AlternateNodeID, reply.AlternateAddr)
	} else {
		__antithesis_instrumentation__.Notify(65151)
	}
	__antithesis_instrumentation__.Notify(65128)

	g.signalConnectedLocked()

	if nodeID := g.NodeID.Get(); nodeID == c.peerID {
		__antithesis_instrumentation__.Notify(65152)
		return errors.Errorf("stopping outgoing client to n%d (%s); loopback connection", c.peerID, c.addr)
	} else {
		__antithesis_instrumentation__.Notify(65153)
		if g.hasIncomingLocked(c.peerID) && func() bool {
			__antithesis_instrumentation__.Notify(65154)
			return nodeID > c.peerID == true
		}() == true {
			__antithesis_instrumentation__.Notify(65155)

			return errors.Errorf("stopping outgoing client to n%d (%s); already have incoming", c.peerID, c.addr)
		} else {
			__antithesis_instrumentation__.Notify(65156)
		}
	}
	__antithesis_instrumentation__.Notify(65129)

	return nil
}

func (c *client) gossip(
	ctx context.Context,
	g *Gossip,
	stream Gossip_GossipClient,
	stopper *stop.Stopper,
	wg *sync.WaitGroup,
) error {
	__antithesis_instrumentation__.Notify(65157)
	sendGossipChan := make(chan struct{}, 1)

	updateCallback := func(_ string, _ roachpb.Value) {
		__antithesis_instrumentation__.Notify(65162)
		select {
		case sendGossipChan <- struct{}{}:
			__antithesis_instrumentation__.Notify(65163)
		default:
			__antithesis_instrumentation__.Notify(65164)
		}
	}
	__antithesis_instrumentation__.Notify(65158)

	errCh := make(chan error, 1)
	initCh := make(chan struct{}, 1)

	wg.Add(1)
	if err := stopper.RunAsyncTask(ctx, "client-gossip", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(65165)
		defer wg.Done()

		errCh <- func() error {
			__antithesis_instrumentation__.Notify(65166)
			var peerID roachpb.NodeID

			initCh := initCh
			for init := true; ; init = false {
				__antithesis_instrumentation__.Notify(65167)
				reply, err := stream.Recv()
				if err != nil {
					__antithesis_instrumentation__.Notify(65171)
					return err
				} else {
					__antithesis_instrumentation__.Notify(65172)
				}
				__antithesis_instrumentation__.Notify(65168)
				if err := c.handleResponse(ctx, g, reply); err != nil {
					__antithesis_instrumentation__.Notify(65173)
					return err
				} else {
					__antithesis_instrumentation__.Notify(65174)
				}
				__antithesis_instrumentation__.Notify(65169)
				if init {
					__antithesis_instrumentation__.Notify(65175)
					initCh <- struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(65176)
				}
				__antithesis_instrumentation__.Notify(65170)
				if peerID == 0 && func() bool {
					__antithesis_instrumentation__.Notify(65177)
					return c.peerID != 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(65178)
					peerID = c.peerID
					g.updateClients()
				} else {
					__antithesis_instrumentation__.Notify(65179)
				}
			}
		}()
	}); err != nil {
		__antithesis_instrumentation__.Notify(65180)
		wg.Done()
		return err
	} else {
		__antithesis_instrumentation__.Notify(65181)
	}
	__antithesis_instrumentation__.Notify(65159)

	var unregister func()
	defer func() {
		__antithesis_instrumentation__.Notify(65182)
		if unregister != nil {
			__antithesis_instrumentation__.Notify(65183)
			unregister()
		} else {
			__antithesis_instrumentation__.Notify(65184)
		}
	}()
	__antithesis_instrumentation__.Notify(65160)
	maybeRegister := func() {
		__antithesis_instrumentation__.Notify(65185)
		if unregister == nil {
			__antithesis_instrumentation__.Notify(65186)

			unregister = g.RegisterCallback(".*", updateCallback, Redundant)
		} else {
			__antithesis_instrumentation__.Notify(65187)
		}
	}
	__antithesis_instrumentation__.Notify(65161)
	initTimer := time.NewTimer(time.Second)
	defer initTimer.Stop()

	for count := 0; ; {
		__antithesis_instrumentation__.Notify(65188)
		select {
		case <-c.closer:
			__antithesis_instrumentation__.Notify(65189)
			return nil
		case <-stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(65190)
			return nil
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(65191)
			return err
		case <-initCh:
			__antithesis_instrumentation__.Notify(65192)
			maybeRegister()
		case <-initTimer.C:
			__antithesis_instrumentation__.Notify(65193)
			maybeRegister()
		case <-sendGossipChan:
			__antithesis_instrumentation__.Notify(65194)
			if err := c.sendGossip(g, stream, count == 0); err != nil {
				__antithesis_instrumentation__.Notify(65196)
				return err
			} else {
				__antithesis_instrumentation__.Notify(65197)
			}
			__antithesis_instrumentation__.Notify(65195)
			count++
		}
	}
}
