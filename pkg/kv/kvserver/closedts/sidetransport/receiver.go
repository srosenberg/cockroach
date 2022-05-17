package sidetransport

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Receiver struct {
	log.AmbientContext
	stop         *stop.Stopper
	stores       Stores
	testingKnobs receiverTestingKnobs

	mu struct {
		syncutil.RWMutex

		conns map[roachpb.NodeID]*incomingStream
	}

	historyMu struct {
		syncutil.Mutex
		lastClosed map[roachpb.NodeID]streamCloseInfo
	}
}

type streamCloseInfo struct {
	nodeID    roachpb.NodeID
	closeErr  error
	closeTime time.Time
}

type receiverTestingKnobs map[roachpb.NodeID]incomingStreamTestingKnobs

var _ ctpb.SideTransportServer = &Receiver{}

func NewReceiver(
	nodeID *base.NodeIDContainer,
	stop *stop.Stopper,
	stores Stores,
	testingKnobs receiverTestingKnobs,
) *Receiver {
	__antithesis_instrumentation__.Notify(98623)
	r := &Receiver{
		stop:         stop,
		stores:       stores,
		testingKnobs: testingKnobs,
	}
	r.AmbientContext.AddLogTag("n", nodeID)
	r.mu.conns = make(map[roachpb.NodeID]*incomingStream)
	r.historyMu.lastClosed = make(map[roachpb.NodeID]streamCloseInfo)
	return r
}

func (s *Receiver) PushUpdates(stream ctpb.SideTransport_PushUpdatesServer) error {
	__antithesis_instrumentation__.Notify(98624)

	ctx := s.AnnotateCtx(stream.Context())
	return newIncomingStream(s, s.stores).Run(ctx, s.stop, stream)
}

func (s *Receiver) GetClosedTimestamp(
	ctx context.Context, rangeID roachpb.RangeID, leaseholderNode roachpb.NodeID,
) (hlc.Timestamp, ctpb.LAI) {
	__antithesis_instrumentation__.Notify(98625)
	s.mu.RLock()
	conn, ok := s.mu.conns[leaseholderNode]
	s.mu.RUnlock()
	if !ok {
		__antithesis_instrumentation__.Notify(98627)
		return hlc.Timestamp{}, 0
	} else {
		__antithesis_instrumentation__.Notify(98628)
	}
	__antithesis_instrumentation__.Notify(98626)
	return conn.GetClosedTimestamp(ctx, rangeID)
}

func (s *Receiver) onFirstMsg(ctx context.Context, r *incomingStream, nodeID roachpb.NodeID) error {
	__antithesis_instrumentation__.Notify(98629)
	s.mu.Lock()
	defer s.mu.Unlock()

	log.VEventf(ctx, 2, "n%d opened a closed timestamps side-transport connection", nodeID)

	if _, ok := s.mu.conns[nodeID]; ok {
		__antithesis_instrumentation__.Notify(98631)
		return errors.Errorf("connection from n%d already exists", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(98632)
	}
	__antithesis_instrumentation__.Notify(98630)
	s.mu.conns[nodeID] = r
	r.testingKnobs = s.testingKnobs[nodeID]
	return nil
}

func (s *Receiver) onRecvErr(ctx context.Context, nodeID roachpb.NodeID, err error) {
	__antithesis_instrumentation__.Notify(98633)
	s.mu.Lock()
	defer s.mu.Unlock()

	if err != io.EOF {
		__antithesis_instrumentation__.Notify(98635)
		log.Warningf(ctx, "closed timestamps side-transport connection dropped from node: %d", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(98636)
		log.VEventf(ctx, 2, "closed timestamps side-transport connection dropped from node: %d (%s)", nodeID, err)
	}
	__antithesis_instrumentation__.Notify(98634)
	if nodeID != 0 {
		__antithesis_instrumentation__.Notify(98637)

		delete(s.mu.conns, nodeID)
		s.historyMu.Lock()
		s.historyMu.lastClosed[nodeID] = streamCloseInfo{
			nodeID:    nodeID,
			closeErr:  err,
			closeTime: timeutil.Now(),
		}
		s.historyMu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(98638)
	}
}

type incomingStream struct {
	server       *Receiver
	stores       Stores
	testingKnobs incomingStreamTestingKnobs
	connectedAt  time.Time

	nodeID roachpb.NodeID

	mu struct {
		syncutil.RWMutex
		streamState
		lastReceived time.Time
	}
}

type incomingStreamTestingKnobs struct {
	onFirstMsg chan struct{}
	onRecvErr  func(sender roachpb.NodeID, err error)
	onMsg      chan *ctpb.Update
}

type Stores interface {
	ForwardSideTransportClosedTimestampForRange(
		ctx context.Context, rangeID roachpb.RangeID, closedTS hlc.Timestamp, lai ctpb.LAI)
}

func newIncomingStream(s *Receiver, stores Stores) *incomingStream {
	__antithesis_instrumentation__.Notify(98639)
	r := &incomingStream{
		server:      s,
		stores:      stores,
		connectedAt: timeutil.Now(),
	}
	return r
}

func (r *incomingStream) GetClosedTimestamp(
	ctx context.Context, rangeID roachpb.RangeID,
) (hlc.Timestamp, ctpb.LAI) {
	__antithesis_instrumentation__.Notify(98640)
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.mu.tracked[rangeID]
	if !ok {
		__antithesis_instrumentation__.Notify(98642)
		return hlc.Timestamp{}, 0
	} else {
		__antithesis_instrumentation__.Notify(98643)
	}
	__antithesis_instrumentation__.Notify(98641)
	return r.mu.lastClosed[info.policy], info.lai
}

func (r *incomingStream) processUpdate(ctx context.Context, msg *ctpb.Update) {
	__antithesis_instrumentation__.Notify(98644)
	log.VEventf(ctx, 4, "received side-transport update: %v", msg)

	if msg.NodeID == 0 {
		__antithesis_instrumentation__.Notify(98651)
		log.Fatalf(ctx, "missing NodeID in message: %s", msg)
	} else {
		__antithesis_instrumentation__.Notify(98652)
	}
	__antithesis_instrumentation__.Notify(98645)

	if msg.NodeID != r.nodeID {
		__antithesis_instrumentation__.Notify(98653)
		log.Fatalf(ctx, "wrong NodeID; expected %d, got %d", r.nodeID, msg.NodeID)
	} else {
		__antithesis_instrumentation__.Notify(98654)
	}
	__antithesis_instrumentation__.Notify(98646)

	if len(msg.Removed) != 0 {
		__antithesis_instrumentation__.Notify(98655)

		r.mu.RLock()
		for _, rangeID := range msg.Removed {
			__antithesis_instrumentation__.Notify(98657)
			info, ok := r.mu.tracked[rangeID]
			if !ok {
				__antithesis_instrumentation__.Notify(98659)
				log.Fatalf(ctx, "attempting to unregister a missing range: r%d", rangeID)
			} else {
				__antithesis_instrumentation__.Notify(98660)
			}
			__antithesis_instrumentation__.Notify(98658)
			r.stores.ForwardSideTransportClosedTimestampForRange(
				ctx, rangeID, r.mu.lastClosed[info.policy], info.lai)
		}
		__antithesis_instrumentation__.Notify(98656)
		r.mu.RUnlock()
	} else {
		__antithesis_instrumentation__.Notify(98661)
	}
	__antithesis_instrumentation__.Notify(98647)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.mu.lastReceived = timeutil.Now()

	if msg.Snapshot {
		__antithesis_instrumentation__.Notify(98662)
		for i := range r.mu.lastClosed {
			__antithesis_instrumentation__.Notify(98664)
			r.mu.lastClosed[i] = hlc.Timestamp{}
		}
		__antithesis_instrumentation__.Notify(98663)
		r.mu.tracked = make(map[roachpb.RangeID]trackedRange, len(r.mu.tracked))
	} else {
		__antithesis_instrumentation__.Notify(98665)
		if msg.SeqNum != r.mu.lastSeqNum+1 {
			__antithesis_instrumentation__.Notify(98666)
			log.Fatalf(ctx, "expected closed timestamp side-transport message with sequence number "+
				"%d, got %d", r.mu.lastSeqNum+1, msg.SeqNum)
		} else {
			__antithesis_instrumentation__.Notify(98667)
		}
	}
	__antithesis_instrumentation__.Notify(98648)
	r.mu.lastSeqNum = msg.SeqNum

	for _, rng := range msg.AddedOrUpdated {
		__antithesis_instrumentation__.Notify(98668)
		r.mu.tracked[rng.RangeID] = trackedRange{
			lai:    rng.LAI,
			policy: rng.Policy,
		}
	}
	__antithesis_instrumentation__.Notify(98649)
	for _, rangeID := range msg.Removed {
		__antithesis_instrumentation__.Notify(98669)
		delete(r.mu.tracked, rangeID)
	}
	__antithesis_instrumentation__.Notify(98650)
	for _, update := range msg.ClosedTimestamps {
		__antithesis_instrumentation__.Notify(98670)
		r.mu.lastClosed[update.Policy] = update.ClosedTimestamp
	}
}

func (r *incomingStream) Run(
	ctx context.Context,
	stopper *stop.Stopper,

	stream ctpb.SideTransport_PushUpdatesServer,
) error {
	__antithesis_instrumentation__.Notify(98671)

	streamDone := make(chan struct{})
	if err := stopper.RunAsyncTask(ctx, "closedts side-transport server conn", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(98674)

		defer close(streamDone)
		for {
			__antithesis_instrumentation__.Notify(98675)
			msg, err := stream.Recv()
			if err != nil {
				__antithesis_instrumentation__.Notify(98678)
				if fn := r.testingKnobs.onRecvErr; fn != nil {
					__antithesis_instrumentation__.Notify(98680)
					fn(r.nodeID, err)
				} else {
					__antithesis_instrumentation__.Notify(98681)
				}
				__antithesis_instrumentation__.Notify(98679)

				r.server.onRecvErr(ctx, r.nodeID, err)
				return
			} else {
				__antithesis_instrumentation__.Notify(98682)
			}
			__antithesis_instrumentation__.Notify(98676)

			if r.nodeID == 0 {
				__antithesis_instrumentation__.Notify(98683)
				r.nodeID = msg.NodeID

				if err := r.server.onFirstMsg(ctx, r, r.nodeID); err != nil {
					__antithesis_instrumentation__.Notify(98685)
					log.Warningf(ctx, "%s", err.Error())
					return
				} else {
					__antithesis_instrumentation__.Notify(98686)
					if ch := r.testingKnobs.onFirstMsg; ch != nil {
						__antithesis_instrumentation__.Notify(98687)
						ch <- struct{}{}
					} else {
						__antithesis_instrumentation__.Notify(98688)
					}
				}
				__antithesis_instrumentation__.Notify(98684)
				if !msg.Snapshot {
					__antithesis_instrumentation__.Notify(98689)
					log.Fatal(ctx, "expected the first message to be a snapshot")
				} else {
					__antithesis_instrumentation__.Notify(98690)
				}
			} else {
				__antithesis_instrumentation__.Notify(98691)
			}
			__antithesis_instrumentation__.Notify(98677)

			r.processUpdate(ctx, msg)
			if ch := r.testingKnobs.onMsg; ch != nil {
				__antithesis_instrumentation__.Notify(98692)
				select {
				case ch <- msg:
					__antithesis_instrumentation__.Notify(98693)
				default:
					__antithesis_instrumentation__.Notify(98694)
				}
			} else {
				__antithesis_instrumentation__.Notify(98695)
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(98696)
		return err
	} else {
		__antithesis_instrumentation__.Notify(98697)
	}
	__antithesis_instrumentation__.Notify(98672)

	select {
	case <-streamDone:
		__antithesis_instrumentation__.Notify(98698)
	case <-stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(98699)
	}
	__antithesis_instrumentation__.Notify(98673)

	return nil
}
