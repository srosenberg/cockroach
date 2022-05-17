package sidetransport

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"google.golang.org/grpc"
)

type Sender struct {
	stopper *stop.Stopper
	st      *cluster.Settings
	clock   *hlc.Clock
	nodeID  roachpb.NodeID

	connFactory connFactory

	trackedMu struct {
		syncutil.Mutex
		streamState

		closingFailures [MaxReason]int
	}

	leaseholdersMu struct {
		syncutil.Mutex
		leaseholders map[roachpb.RangeID]leaseholder
	}

	buf *updatesBuf

	connsMu struct {
		syncutil.Mutex
		conns map[roachpb.NodeID]conn
	}
}

type streamState struct {
	lastSeqNum ctpb.SeqNum

	lastClosed [roachpb.MAX_CLOSED_TIMESTAMP_POLICY]hlc.Timestamp

	tracked map[roachpb.RangeID]trackedRange
}

type connTestingKnobs struct {
	beforeSend func(destNodeID roachpb.NodeID, msg *ctpb.Update)
}

type trackedRange struct {
	lai    ctpb.LAI
	policy roachpb.RangeClosedTimestampPolicy
}

type leaseholder struct {
	Replica
	leaseSeq roachpb.LeaseSequence
}

type Replica interface {
	StoreID() roachpb.StoreID
	GetRangeID() roachpb.RangeID

	BumpSideTransportClosed(
		ctx context.Context,
		now hlc.ClockTimestamp,
		targetByPolicy [roachpb.MAX_CLOSED_TIMESTAMP_POLICY]hlc.Timestamp,
	) BumpSideTransportClosedResult
}

type BumpSideTransportClosedResult struct {
	OK         bool
	FailReason CantCloseReason

	Desc *roachpb.RangeDescriptor

	LAI ctpb.LAI

	Policy roachpb.RangeClosedTimestampPolicy
}

type CantCloseReason int

const (
	ReasonUnknown CantCloseReason = iota
	ReplicaDestroyed
	InvalidLease
	TargetOverLeaseExpiration
	MergeInProgress
	ProposalsInFlight
	RequestsEvaluatingBelowTarget
	MaxReason
)

func NewSender(
	stopper *stop.Stopper, st *cluster.Settings, clock *hlc.Clock, dialer *nodedialer.Dialer,
) *Sender {
	__antithesis_instrumentation__.Notify(98700)
	return newSenderWithConnFactory(stopper, st, clock, newRPCConnFactory(dialer, connTestingKnobs{}))
}

func newSenderWithConnFactory(
	stopper *stop.Stopper, st *cluster.Settings, clock *hlc.Clock, connFactory connFactory,
) *Sender {
	__antithesis_instrumentation__.Notify(98701)
	s := &Sender{
		stopper:     stopper,
		st:          st,
		clock:       clock,
		connFactory: connFactory,
		buf:         newUpdatesBuf(),
	}
	s.trackedMu.tracked = make(map[roachpb.RangeID]trackedRange)
	s.leaseholdersMu.leaseholders = make(map[roachpb.RangeID]leaseholder)
	s.connsMu.conns = make(map[roachpb.NodeID]conn)
	return s
}

func (s *Sender) Run(ctx context.Context, nodeID roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(98702)
	s.nodeID = nodeID
	confCh := make(chan struct{}, 1)
	confChanged := func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(98704)
		select {
		case confCh <- struct{}{}:
			__antithesis_instrumentation__.Notify(98705)
		default:
			__antithesis_instrumentation__.Notify(98706)
		}
	}
	__antithesis_instrumentation__.Notify(98703)
	closedts.SideTransportCloseInterval.SetOnChange(&s.st.SV, confChanged)

	_ = s.stopper.RunAsyncTask(ctx, "closedts side-transport publisher",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(98707)
			defer func() {
				__antithesis_instrumentation__.Notify(98709)

				s.buf.Close()
			}()
			__antithesis_instrumentation__.Notify(98708)

			timer := timeutil.NewTimer()
			defer timer.Stop()
			for {
				__antithesis_instrumentation__.Notify(98710)
				interval := closedts.SideTransportCloseInterval.Get(&s.st.SV)
				if interval > 0 {
					__antithesis_instrumentation__.Notify(98712)
					timer.Reset(interval)
				} else {
					__antithesis_instrumentation__.Notify(98713)

					timer.Stop()
					timer = timeutil.NewTimer()
				}
				__antithesis_instrumentation__.Notify(98711)
				select {
				case <-timer.C:
					__antithesis_instrumentation__.Notify(98714)
					timer.Read = true
					s.publish(ctx)
				case <-confCh:
					__antithesis_instrumentation__.Notify(98715)

					continue
				case <-s.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(98716)
					return
				}
			}
		})
}

func (s *Sender) RegisterLeaseholder(
	ctx context.Context, r Replica, leaseSeq roachpb.LeaseSequence,
) {
	__antithesis_instrumentation__.Notify(98717)
	s.leaseholdersMu.Lock()
	defer s.leaseholdersMu.Unlock()

	if lh, ok := s.leaseholdersMu.leaseholders[r.GetRangeID()]; ok {
		__antithesis_instrumentation__.Notify(98719)

		if lh.leaseSeq >= leaseSeq {
			__antithesis_instrumentation__.Notify(98720)
			return
		} else {
			__antithesis_instrumentation__.Notify(98721)
		}

	} else {
		__antithesis_instrumentation__.Notify(98722)
	}
	__antithesis_instrumentation__.Notify(98718)
	s.leaseholdersMu.leaseholders[r.GetRangeID()] = leaseholder{
		Replica:  r,
		leaseSeq: leaseSeq,
	}
}

func (s *Sender) UnregisterLeaseholder(
	ctx context.Context, storeID roachpb.StoreID, rangeID roachpb.RangeID,
) {
	__antithesis_instrumentation__.Notify(98723)
	s.leaseholdersMu.Lock()
	defer s.leaseholdersMu.Unlock()

	if lh, ok := s.leaseholdersMu.leaseholders[rangeID]; ok && func() bool {
		__antithesis_instrumentation__.Notify(98724)
		return lh.StoreID() == storeID == true
	}() == true {
		__antithesis_instrumentation__.Notify(98725)
		delete(s.leaseholdersMu.leaseholders, rangeID)
	} else {
		__antithesis_instrumentation__.Notify(98726)
	}
}

func (s *Sender) publish(ctx context.Context) hlc.ClockTimestamp {
	__antithesis_instrumentation__.Notify(98727)
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()
	log.VEventf(ctx, 4, "side-transport generating a new message")
	s.trackedMu.closingFailures = [MaxReason]int{}

	msg := &ctpb.Update{
		NodeID:           s.nodeID,
		ClosedTimestamps: make([]ctpb.Update_GroupUpdate, len(s.trackedMu.lastClosed)),
	}

	s.trackedMu.lastSeqNum++
	msg.SeqNum = s.trackedMu.lastSeqNum

	msg.Snapshot = msg.SeqNum == 1

	now := s.clock.NowAsClockTimestamp()
	maxClockOffset := s.clock.MaxOffset()
	lagTargetDuration := closedts.TargetDuration.Get(&s.st.SV)
	leadTargetOverride := closedts.LeadForGlobalReadsOverride.Get(&s.st.SV)
	sideTransportCloseInterval := closedts.SideTransportCloseInterval.Get(&s.st.SV)
	for i := range s.trackedMu.lastClosed {
		__antithesis_instrumentation__.Notify(98732)
		pol := roachpb.RangeClosedTimestampPolicy(i)
		target := closedts.TargetForPolicy(
			now,
			maxClockOffset,
			lagTargetDuration,
			leadTargetOverride,
			sideTransportCloseInterval,
			pol,
		)
		s.trackedMu.lastClosed[pol] = target
		msg.ClosedTimestamps[pol] = ctpb.Update_GroupUpdate{
			Policy:          pol,
			ClosedTimestamp: target,
		}
	}
	__antithesis_instrumentation__.Notify(98728)

	s.leaseholdersMu.Lock()
	leaseholders := make(map[roachpb.RangeID]leaseholder, len(s.leaseholdersMu.leaseholders))
	for k, v := range s.leaseholdersMu.leaseholders {
		__antithesis_instrumentation__.Notify(98733)
		leaseholders[k] = v
	}
	__antithesis_instrumentation__.Notify(98729)
	s.leaseholdersMu.Unlock()

	nodesWithFollowers := util.MakeFastIntSet()

	for rid := range s.trackedMu.tracked {
		__antithesis_instrumentation__.Notify(98734)
		if _, ok := leaseholders[rid]; !ok {
			__antithesis_instrumentation__.Notify(98735)
			msg.Removed = append(msg.Removed, rid)
			delete(s.trackedMu.tracked, rid)
		} else {
			__antithesis_instrumentation__.Notify(98736)
		}
	}
	__antithesis_instrumentation__.Notify(98730)

	for _, lh := range leaseholders {
		__antithesis_instrumentation__.Notify(98737)
		lhRangeID := lh.GetRangeID()
		lastMsg, tracked := s.trackedMu.tracked[lhRangeID]

		closeRes := lh.BumpSideTransportClosed(ctx, now, s.trackedMu.lastClosed)

		repls := closeRes.Desc.Replicas().Descriptors()
		for i := range repls {
			__antithesis_instrumentation__.Notify(98741)
			nodesWithFollowers.Add(int(repls[i].NodeID))
		}
		__antithesis_instrumentation__.Notify(98738)

		if !closeRes.OK {
			__antithesis_instrumentation__.Notify(98742)
			s.trackedMu.closingFailures[closeRes.FailReason]++

			if tracked {
				__antithesis_instrumentation__.Notify(98744)
				msg.Removed = append(msg.Removed, lhRangeID)
				delete(s.trackedMu.tracked, lhRangeID)
			} else {
				__antithesis_instrumentation__.Notify(98745)
			}
			__antithesis_instrumentation__.Notify(98743)
			continue
		} else {
			__antithesis_instrumentation__.Notify(98746)
		}
		__antithesis_instrumentation__.Notify(98739)

		needExplicit := false
		if !tracked {
			__antithesis_instrumentation__.Notify(98747)

			needExplicit = true
		} else {
			__antithesis_instrumentation__.Notify(98748)
			if lastMsg.lai < closeRes.LAI {
				__antithesis_instrumentation__.Notify(98749)

				needExplicit = true
			} else {
				__antithesis_instrumentation__.Notify(98750)
				if lastMsg.policy != closeRes.Policy {
					__antithesis_instrumentation__.Notify(98751)

					needExplicit = true
				} else {
					__antithesis_instrumentation__.Notify(98752)
				}
			}
		}
		__antithesis_instrumentation__.Notify(98740)
		if needExplicit {
			__antithesis_instrumentation__.Notify(98753)
			msg.AddedOrUpdated = append(msg.AddedOrUpdated, ctpb.Update_RangeUpdate{
				RangeID: lhRangeID,
				LAI:     closeRes.LAI,
				Policy:  closeRes.Policy,
			})
			s.trackedMu.tracked[lhRangeID] = trackedRange{lai: closeRes.LAI, policy: closeRes.Policy}
		} else {
			__antithesis_instrumentation__.Notify(98754)
		}
	}

	{
		__antithesis_instrumentation__.Notify(98755)
		s.connsMu.Lock()
		for nodeID, c := range s.connsMu.conns {
			__antithesis_instrumentation__.Notify(98758)
			if !nodesWithFollowers.Contains(int(nodeID)) {
				__antithesis_instrumentation__.Notify(98759)
				delete(s.connsMu.conns, nodeID)
				c.close()
			} else {
				__antithesis_instrumentation__.Notify(98760)
			}
		}
		__antithesis_instrumentation__.Notify(98756)

		nodesWithFollowers.ForEach(func(nid int) {
			__antithesis_instrumentation__.Notify(98761)
			nodeID := roachpb.NodeID(nid)

			if _, ok := s.connsMu.conns[nodeID]; !ok && func() bool {
				__antithesis_instrumentation__.Notify(98762)
				return nodeID != s.nodeID == true
			}() == true {
				__antithesis_instrumentation__.Notify(98763)
				c := s.connFactory.new(s, nodeID)
				c.run(ctx, s.stopper)
				s.connsMu.conns[nodeID] = c
			} else {
				__antithesis_instrumentation__.Notify(98764)
			}
		})
		__antithesis_instrumentation__.Notify(98757)
		s.connsMu.Unlock()
	}
	__antithesis_instrumentation__.Notify(98731)

	log.VEventf(ctx, 4, "side-transport publishing message with closed timestamps: %v (%v)", msg.ClosedTimestamps, msg)
	s.buf.Push(ctx, msg)

	return now
}

func (s *Sender) GetSnapshot() *ctpb.Update {
	__antithesis_instrumentation__.Notify(98765)
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()

	msg := &ctpb.Update{
		NodeID: s.nodeID,

		SeqNum:           s.trackedMu.lastSeqNum,
		Snapshot:         true,
		ClosedTimestamps: make([]ctpb.Update_GroupUpdate, len(s.trackedMu.lastClosed)),
		AddedOrUpdated:   make([]ctpb.Update_RangeUpdate, 0, len(s.trackedMu.tracked)),
	}
	for pol, ts := range s.trackedMu.lastClosed {
		__antithesis_instrumentation__.Notify(98768)
		msg.ClosedTimestamps[pol] = ctpb.Update_GroupUpdate{
			Policy:          roachpb.RangeClosedTimestampPolicy(pol),
			ClosedTimestamp: ts,
		}
	}
	__antithesis_instrumentation__.Notify(98766)
	for rid, r := range s.trackedMu.tracked {
		__antithesis_instrumentation__.Notify(98769)
		msg.AddedOrUpdated = append(msg.AddedOrUpdated, ctpb.Update_RangeUpdate{
			RangeID: rid,
			LAI:     r.lai,
			Policy:  r.policy,
		})
	}
	__antithesis_instrumentation__.Notify(98767)
	return msg
}

type updatesBuf struct {
	mu struct {
		syncutil.Mutex

		updated sync.Cond

		data []*ctpb.Update

		head, tail int

		closed bool
	}
}

const updatesBufSize = 50

func newUpdatesBuf() *updatesBuf {
	__antithesis_instrumentation__.Notify(98770)
	buf := &updatesBuf{}
	buf.mu.updated.L = &buf.mu
	buf.mu.data = make([]*ctpb.Update, updatesBufSize)
	return buf
}

func (b *updatesBuf) Push(ctx context.Context, update *ctpb.Update) {
	__antithesis_instrumentation__.Notify(98771)
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.sizeLocked() != 0 {
		__antithesis_instrumentation__.Notify(98774)
		lastIdx := b.lastIdxLocked()
		if prevSeq := b.mu.data[lastIdx].SeqNum; prevSeq != update.SeqNum-1 {
			__antithesis_instrumentation__.Notify(98775)
			log.Fatalf(ctx, "bad sequence number; expected %d, got %d", prevSeq+1, update.SeqNum)
		} else {
			__antithesis_instrumentation__.Notify(98776)
		}
	} else {
		__antithesis_instrumentation__.Notify(98777)
	}
	__antithesis_instrumentation__.Notify(98772)

	overwrite := b.fullLocked()
	b.mu.data[b.mu.tail] = update
	b.mu.tail = (b.mu.tail + 1) % len(b.mu.data)

	if overwrite {
		__antithesis_instrumentation__.Notify(98778)
		b.mu.head = (b.mu.head + 1) % len(b.mu.data)
	} else {
		__antithesis_instrumentation__.Notify(98779)
	}
	__antithesis_instrumentation__.Notify(98773)

	b.mu.updated.Broadcast()
}

func (b *updatesBuf) lastIdxLocked() int {
	__antithesis_instrumentation__.Notify(98780)
	lastIdx := b.mu.tail - 1
	if lastIdx < 0 {
		__antithesis_instrumentation__.Notify(98782)
		lastIdx += len(b.mu.data)
	} else {
		__antithesis_instrumentation__.Notify(98783)
	}
	__antithesis_instrumentation__.Notify(98781)
	return lastIdx
}

func (b *updatesBuf) GetBySeq(ctx context.Context, seqNum ctpb.SeqNum) (*ctpb.Update, bool) {
	__antithesis_instrumentation__.Notify(98784)
	b.mu.Lock()
	defer b.mu.Unlock()

	for {
		__antithesis_instrumentation__.Notify(98785)
		if b.mu.closed {
			__antithesis_instrumentation__.Notify(98791)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(98792)
		}
		__antithesis_instrumentation__.Notify(98786)

		var firstSeq, lastSeq ctpb.SeqNum
		if b.sizeLocked() == 0 {
			__antithesis_instrumentation__.Notify(98793)
			firstSeq, lastSeq = 0, 0
		} else {
			__antithesis_instrumentation__.Notify(98794)
			firstSeq, lastSeq = b.mu.data[b.mu.head].SeqNum, b.mu.data[b.lastIdxLocked()].SeqNum
		}
		__antithesis_instrumentation__.Notify(98787)
		if seqNum < firstSeq {
			__antithesis_instrumentation__.Notify(98795)

			return nil, true
		} else {
			__antithesis_instrumentation__.Notify(98796)
		}
		__antithesis_instrumentation__.Notify(98788)

		if seqNum == lastSeq+1 {
			__antithesis_instrumentation__.Notify(98797)
			b.mu.updated.Wait()
			continue
		} else {
			__antithesis_instrumentation__.Notify(98798)
		}
		__antithesis_instrumentation__.Notify(98789)
		if seqNum > lastSeq+1 {
			__antithesis_instrumentation__.Notify(98799)
			log.Fatalf(ctx, "skipping sequence numbers; requested: %d, last: %d", seqNum, lastSeq)
		} else {
			__antithesis_instrumentation__.Notify(98800)
		}
		__antithesis_instrumentation__.Notify(98790)
		idx := (b.mu.head + (int)(seqNum-firstSeq)) % len(b.mu.data)
		return b.mu.data[idx], true
	}
}

func (b *updatesBuf) sizeLocked() int {
	__antithesis_instrumentation__.Notify(98801)
	if b.mu.head < b.mu.tail {
		__antithesis_instrumentation__.Notify(98802)
		return b.mu.tail - b.mu.head
	} else {
		__antithesis_instrumentation__.Notify(98803)
		if b.mu.head == b.mu.tail {
			__antithesis_instrumentation__.Notify(98804)

			if b.mu.head == 0 && func() bool {
				__antithesis_instrumentation__.Notify(98806)
				return b.mu.data[0] == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(98807)
				return 0
			} else {
				__antithesis_instrumentation__.Notify(98808)
			}
			__antithesis_instrumentation__.Notify(98805)
			return len(b.mu.data)
		} else {
			__antithesis_instrumentation__.Notify(98809)
			return len(b.mu.data) + b.mu.tail - b.mu.head
		}
	}
}

func (b *updatesBuf) fullLocked() bool {
	__antithesis_instrumentation__.Notify(98810)
	return b.sizeLocked() == len(b.mu.data)
}

func (b *updatesBuf) Close() {
	__antithesis_instrumentation__.Notify(98811)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.mu.closed = true
	b.mu.updated.Broadcast()
}

type connFactory interface {
	new(*Sender, roachpb.NodeID) conn
}

type conn interface {
	run(context.Context, *stop.Stopper)
	close()
	getState() connState
}

type rpcConnFactory struct {
	dialer       nodeDialer
	testingKnobs connTestingKnobs
}

func newRPCConnFactory(dialer nodeDialer, testingKnobs connTestingKnobs) connFactory {
	__antithesis_instrumentation__.Notify(98812)
	return &rpcConnFactory{
		dialer:       dialer,
		testingKnobs: testingKnobs,
	}
}

func (f *rpcConnFactory) new(s *Sender, nodeID roachpb.NodeID) conn {
	__antithesis_instrumentation__.Notify(98813)
	return newRPCConn(f.dialer, s, nodeID, f.testingKnobs)
}

type nodeDialer interface {
	Dial(ctx context.Context, nodeID roachpb.NodeID, class rpc.ConnectionClass) (_ *grpc.ClientConn, err error)
}

type rpcConn struct {
	log.AmbientContext
	dialer       nodeDialer
	producer     *Sender
	nodeID       roachpb.NodeID
	testingKnobs connTestingKnobs

	stream   ctpb.SideTransport_PushUpdatesClient
	lastSent ctpb.SeqNum

	cancelStreamCtx context.CancelFunc
	closed          int32

	mu struct {
		syncutil.Mutex
		state connState
	}
}

func newRPCConn(
	dialer nodeDialer, producer *Sender, nodeID roachpb.NodeID, testingKnobs connTestingKnobs,
) conn {
	__antithesis_instrumentation__.Notify(98814)
	r := &rpcConn{
		dialer:       dialer,
		producer:     producer,
		nodeID:       nodeID,
		testingKnobs: testingKnobs,
	}
	r.mu.state.connected = false
	r.AddLogTag("ctstream", nodeID)
	return r
}

func (r *rpcConn) cleanupStream(err error) {
	__antithesis_instrumentation__.Notify(98815)
	if r.stream == nil {
		__antithesis_instrumentation__.Notify(98817)
		return
	} else {
		__antithesis_instrumentation__.Notify(98818)
	}
	__antithesis_instrumentation__.Notify(98816)
	_ = r.stream.CloseSend()
	r.stream = nil
	r.cancelStreamCtx()
	r.cancelStreamCtx = nil

	r.lastSent = 0

	r.mu.Lock()
	r.mu.state.connected = false
	r.mu.state.lastDisconnect = err
	r.mu.state.lastDisconnectTime = timeutil.Now()
	r.mu.Unlock()
}

func (r *rpcConn) close() {
	__antithesis_instrumentation__.Notify(98819)
	atomic.StoreInt32(&r.closed, 1)
}

func (r *rpcConn) maybeConnect(ctx context.Context, stopper *stop.Stopper) error {
	__antithesis_instrumentation__.Notify(98820)
	if r.stream != nil {
		__antithesis_instrumentation__.Notify(98824)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(98825)
	}
	__antithesis_instrumentation__.Notify(98821)

	conn, err := r.dialer.Dial(ctx, r.nodeID, rpc.SystemClass)
	if err != nil {
		__antithesis_instrumentation__.Notify(98826)
		return err
	} else {
		__antithesis_instrumentation__.Notify(98827)
	}
	__antithesis_instrumentation__.Notify(98822)
	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := ctpb.NewSideTransportClient(conn).PushUpdates(streamCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(98828)
		cancel()
		return err
	} else {
		__antithesis_instrumentation__.Notify(98829)
	}
	__antithesis_instrumentation__.Notify(98823)
	r.recordConnect()
	r.stream = stream

	r.cancelStreamCtx = cancel
	return nil
}

func (r *rpcConn) run(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(98830)
	_ = stopper.RunAsyncTask(ctx, fmt.Sprintf("closedts publisher for n%d", r.nodeID),
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(98831)

			ctx, cancel := stopper.WithCancelOnQuiesce(r.AnnotateCtx(ctx))
			defer cancel()

			defer r.cleanupStream(nil)
			everyN := log.Every(10 * time.Second)

			const sleepOnErr = time.Second
			for {
				__antithesis_instrumentation__.Notify(98832)
				if ctx.Err() != nil {
					__antithesis_instrumentation__.Notify(98839)
					return
				} else {
					__antithesis_instrumentation__.Notify(98840)
				}
				__antithesis_instrumentation__.Notify(98833)
				if err := r.maybeConnect(ctx, stopper); err != nil {
					__antithesis_instrumentation__.Notify(98841)
					if everyN.ShouldLog() {
						__antithesis_instrumentation__.Notify(98843)
						log.Infof(ctx, "side-transport failed to connect to n%d: %s", r.nodeID, err)
					} else {
						__antithesis_instrumentation__.Notify(98844)
					}
					__antithesis_instrumentation__.Notify(98842)
					time.Sleep(sleepOnErr)
					continue
				} else {
					__antithesis_instrumentation__.Notify(98845)
				}
				__antithesis_instrumentation__.Notify(98834)

				var msg *ctpb.Update
				var ok bool
				msg, ok = r.producer.buf.GetBySeq(ctx, r.lastSent+1)

				if !ok {
					__antithesis_instrumentation__.Notify(98846)
					return
				} else {
					__antithesis_instrumentation__.Notify(98847)
				}
				__antithesis_instrumentation__.Notify(98835)
				closed := atomic.LoadInt32(&r.closed) > 0
				if closed {
					__antithesis_instrumentation__.Notify(98848)
					return
				} else {
					__antithesis_instrumentation__.Notify(98849)
				}
				__antithesis_instrumentation__.Notify(98836)

				if msg == nil {
					__antithesis_instrumentation__.Notify(98850)

					msg = r.producer.GetSnapshot()
				} else {
					__antithesis_instrumentation__.Notify(98851)
				}
				__antithesis_instrumentation__.Notify(98837)
				r.lastSent = msg.SeqNum

				if fn := r.testingKnobs.beforeSend; fn != nil {
					__antithesis_instrumentation__.Notify(98852)
					fn(r.nodeID, msg)
				} else {
					__antithesis_instrumentation__.Notify(98853)
				}
				__antithesis_instrumentation__.Notify(98838)
				if err := r.stream.Send(msg); err != nil {
					__antithesis_instrumentation__.Notify(98854)
					if err != io.EOF && func() bool {
						__antithesis_instrumentation__.Notify(98856)
						return everyN.ShouldLog() == true
					}() == true {
						__antithesis_instrumentation__.Notify(98857)
						log.Warningf(ctx, "failed to send closed timestamp message %d to n%d: %s",
							r.lastSent, r.nodeID, err)
					} else {
						__antithesis_instrumentation__.Notify(98858)
					}
					__antithesis_instrumentation__.Notify(98855)

					r.cleanupStream(err)
					time.Sleep(sleepOnErr)
				} else {
					__antithesis_instrumentation__.Notify(98859)
				}
			}
		})
}

type connState struct {
	connected          bool
	connectedTime      time.Time
	lastDisconnect     error
	lastDisconnectTime time.Time
}

func (r *rpcConn) getState() connState {
	__antithesis_instrumentation__.Notify(98860)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.mu.state
}

func (r *rpcConn) recordConnect() {
	__antithesis_instrumentation__.Notify(98861)
	r.mu.Lock()
	r.mu.state.connected = true
	r.mu.state.connectedTime = timeutil.Now()
	r.mu.Unlock()
}

func (s streamState) String() string {
	__antithesis_instrumentation__.Notify(98862)
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "ranges tracked: %d\n", len(s.tracked))

	sb.WriteString("closed timestamps: ")
	now := timeutil.Now()
	for policy, closedTS := range s.lastClosed {
		__antithesis_instrumentation__.Notify(98866)
		if policy != 0 {
			__antithesis_instrumentation__.Notify(98869)
			sb.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(98870)
		}
		__antithesis_instrumentation__.Notify(98867)
		ago := now.Sub(closedTS.GoTime()).Truncate(time.Millisecond)
		var agoMsg string
		if ago >= 0 {
			__antithesis_instrumentation__.Notify(98871)
			agoMsg = fmt.Sprintf("%s ago", ago)
		} else {
			__antithesis_instrumentation__.Notify(98872)
			agoMsg = fmt.Sprintf("%s in the future", -ago)
		}
		__antithesis_instrumentation__.Notify(98868)
		fmt.Fprintf(sb, "%s:%s (%s)", roachpb.RangeClosedTimestampPolicy(policy), closedTS, agoMsg)
	}
	__antithesis_instrumentation__.Notify(98863)

	sb.WriteString("\nTracked ranges by policy: (<range>:<LAI>)\n")
	type rangeInfo struct {
		id roachpb.RangeID
		trackedRange
	}
	rangesByPolicy := make(map[roachpb.RangeClosedTimestampPolicy][]rangeInfo)
	for rid, info := range s.tracked {
		__antithesis_instrumentation__.Notify(98873)
		rangesByPolicy[info.policy] = append(rangesByPolicy[info.policy], rangeInfo{id: rid, trackedRange: info})
	}
	__antithesis_instrumentation__.Notify(98864)
	for policy, ranges := range rangesByPolicy {
		__antithesis_instrumentation__.Notify(98874)
		fmt.Fprintf(sb, "%s: ", policy)
		sort.Slice(ranges, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(98877)
			return ranges[i].id < ranges[j].id
		})
		__antithesis_instrumentation__.Notify(98875)
		for i, rng := range ranges {
			__antithesis_instrumentation__.Notify(98878)
			if i > 0 {
				__antithesis_instrumentation__.Notify(98880)
				sb.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(98881)
			}
			__antithesis_instrumentation__.Notify(98879)
			fmt.Fprintf(sb, "r%d:%d", rng.id, rng.lai)
		}
		__antithesis_instrumentation__.Notify(98876)
		if len(ranges) != 0 {
			__antithesis_instrumentation__.Notify(98882)
			sb.WriteRune('\n')
		} else {
			__antithesis_instrumentation__.Notify(98883)
		}
	}
	__antithesis_instrumentation__.Notify(98865)
	return sb.String()
}
