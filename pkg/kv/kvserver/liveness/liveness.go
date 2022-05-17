package liveness

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var (
	ErrMissingRecord = errors.New("missing liveness record")

	ErrRecordCacheMiss = errors.New("liveness record not found in cache")

	errChangeMembershipStatusFailed = errors.New("failed to change the membership status")

	ErrEpochIncremented = errors.New("heartbeat failed on epoch increment")

	ErrEpochAlreadyIncremented = errors.New("epoch already incremented")
)

type errRetryLiveness struct {
	error
}

func (e *errRetryLiveness) Cause() error {
	__antithesis_instrumentation__.Notify(107958)
	return e.error
}

func (e *errRetryLiveness) Error() string {
	__antithesis_instrumentation__.Notify(107959)
	return fmt.Sprintf("%T: %s", *e, e.error)
}

func isErrRetryLiveness(ctx context.Context, err error) bool {
	__antithesis_instrumentation__.Notify(107960)
	if errors.HasType(err, (*roachpb.AmbiguousResultError)(nil)) {
		__antithesis_instrumentation__.Notify(107962)

		return ctx.Err() == nil
	} else {
		__antithesis_instrumentation__.Notify(107963)
		if errors.HasType(err, (*roachpb.TransactionStatusError)(nil)) {
			__antithesis_instrumentation__.Notify(107964)

			return true
		} else {
			__antithesis_instrumentation__.Notify(107965)
			if errors.Is(err, kv.OnePCNotAllowedError{}) {
				__antithesis_instrumentation__.Notify(107966)
				return true
			} else {
				__antithesis_instrumentation__.Notify(107967)
			}
		}
	}
	__antithesis_instrumentation__.Notify(107961)
	return false
}

var (
	metaLiveNodes = metric.Metadata{
		Name:        "liveness.livenodes",
		Help:        "Number of live nodes in the cluster (will be 0 if this node is not itself live)",
		Measurement: "Nodes",
		Unit:        metric.Unit_COUNT,
	}
	metaHeartbeatsInFlight = metric.Metadata{
		Name:        "liveness.heartbeatsinflight",
		Help:        "Number of in-flight liveness heartbeats from this node",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	metaHeartbeatSuccesses = metric.Metadata{
		Name:        "liveness.heartbeatsuccesses",
		Help:        "Number of successful node liveness heartbeats from this node",
		Measurement: "Messages",
		Unit:        metric.Unit_COUNT,
	}
	metaHeartbeatFailures = metric.Metadata{
		Name:        "liveness.heartbeatfailures",
		Help:        "Number of failed node liveness heartbeats from this node",
		Measurement: "Messages",
		Unit:        metric.Unit_COUNT,
	}
	metaEpochIncrements = metric.Metadata{
		Name:        "liveness.epochincrements",
		Help:        "Number of times this node has incremented its liveness epoch",
		Measurement: "Epochs",
		Unit:        metric.Unit_COUNT,
	}
	metaHeartbeatLatency = metric.Metadata{
		Name:        "liveness.heartbeatlatency",
		Help:        "Node liveness heartbeat latency",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
)

type Metrics struct {
	LiveNodes          *metric.Gauge
	HeartbeatsInFlight *metric.Gauge
	HeartbeatSuccesses *metric.Counter
	HeartbeatFailures  telemetry.CounterWithMetric
	EpochIncrements    telemetry.CounterWithMetric
	HeartbeatLatency   *metric.Histogram
}

type IsLiveCallback func(livenesspb.Liveness)

type HeartbeatCallback func(context.Context)

type NodeLiveness struct {
	ambientCtx        log.AmbientContext
	clock             *hlc.Clock
	db                *kv.DB
	gossip            *gossip.Gossip
	livenessThreshold time.Duration
	renewalDuration   time.Duration
	selfSem           chan struct{}
	st                *cluster.Settings
	otherSem          chan struct{}

	heartbeatPaused      uint32
	heartbeatToken       chan struct{}
	metrics              Metrics
	onNodeDecommissioned func(livenesspb.Liveness)

	mu struct {
		syncutil.RWMutex
		onIsLive []IsLiveCallback

		nodes      map[roachpb.NodeID]Record
		onSelfLive HeartbeatCallback

		engines []storage.Engine
	}
}

type Record struct {
	livenesspb.Liveness

	raw []byte
}

type NodeLivenessOptions struct {
	AmbientCtx              log.AmbientContext
	Settings                *cluster.Settings
	Gossip                  *gossip.Gossip
	Clock                   *hlc.Clock
	DB                      *kv.DB
	LivenessThreshold       time.Duration
	RenewalDuration         time.Duration
	HistogramWindowInterval time.Duration

	OnNodeDecommissioned func(livenesspb.Liveness)
}

func NewNodeLiveness(opts NodeLivenessOptions) *NodeLiveness {
	__antithesis_instrumentation__.Notify(107968)
	nl := &NodeLiveness{
		ambientCtx:           opts.AmbientCtx,
		clock:                opts.Clock,
		db:                   opts.DB,
		gossip:               opts.Gossip,
		livenessThreshold:    opts.LivenessThreshold,
		renewalDuration:      opts.RenewalDuration,
		selfSem:              make(chan struct{}, 1),
		st:                   opts.Settings,
		otherSem:             make(chan struct{}, 1),
		heartbeatToken:       make(chan struct{}, 1),
		onNodeDecommissioned: opts.OnNodeDecommissioned,
	}
	nl.metrics = Metrics{
		LiveNodes:          metric.NewFunctionalGauge(metaLiveNodes, nl.numLiveNodes),
		HeartbeatsInFlight: metric.NewGauge(metaHeartbeatsInFlight),
		HeartbeatSuccesses: metric.NewCounter(metaHeartbeatSuccesses),
		HeartbeatFailures:  telemetry.NewCounterWithMetric(metaHeartbeatFailures),
		EpochIncrements:    telemetry.NewCounterWithMetric(metaEpochIncrements),
		HeartbeatLatency:   metric.NewLatency(metaHeartbeatLatency, opts.HistogramWindowInterval),
	}
	nl.mu.nodes = make(map[roachpb.NodeID]Record)
	nl.heartbeatToken <- struct{}{}

	livenessRegex := gossip.MakePrefixPattern(gossip.KeyNodeLivenessPrefix)
	nl.gossip.RegisterCallback(livenessRegex, nl.livenessGossipUpdate)

	return nl
}

var errNodeDrainingSet = errors.New("node is already draining")

func (nl *NodeLiveness) sem(nodeID roachpb.NodeID) chan struct{} {
	__antithesis_instrumentation__.Notify(107969)
	if nodeID == nl.gossip.NodeID.Get() {
		__antithesis_instrumentation__.Notify(107971)
		return nl.selfSem
	} else {
		__antithesis_instrumentation__.Notify(107972)
	}
	__antithesis_instrumentation__.Notify(107970)
	return nl.otherSem
}

func (nl *NodeLiveness) SetDraining(
	ctx context.Context, drain bool, reporter func(int, redact.SafeString),
) error {
	__antithesis_instrumentation__.Notify(107973)
	ctx = nl.ambientCtx.AnnotateCtx(ctx)
	for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(107976)
		oldLivenessRec, ok := nl.SelfEx()
		if !ok {
			__antithesis_instrumentation__.Notify(107979)

			nodeID := nl.gossip.NodeID.Get()
			livenessRec, err := nl.getLivenessRecordFromKV(ctx, nodeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(107981)
				return err
			} else {
				__antithesis_instrumentation__.Notify(107982)
			}
			__antithesis_instrumentation__.Notify(107980)
			oldLivenessRec = livenessRec
		} else {
			__antithesis_instrumentation__.Notify(107983)
		}
		__antithesis_instrumentation__.Notify(107977)
		if err := nl.setDrainingInternal(ctx, oldLivenessRec, drain, reporter); err != nil {
			__antithesis_instrumentation__.Notify(107984)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(107987)
				log.Infof(ctx, "attempting to set liveness draining status to %v: %v", drain, err)
			} else {
				__antithesis_instrumentation__.Notify(107988)
			}
			__antithesis_instrumentation__.Notify(107985)
			if grpcutil.IsConnectionRejected(err) {
				__antithesis_instrumentation__.Notify(107989)
				return err
			} else {
				__antithesis_instrumentation__.Notify(107990)
			}
			__antithesis_instrumentation__.Notify(107986)
			continue
		} else {
			__antithesis_instrumentation__.Notify(107991)
		}
		__antithesis_instrumentation__.Notify(107978)
		return nil
	}
	__antithesis_instrumentation__.Notify(107974)
	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(107992)
		return err
	} else {
		__antithesis_instrumentation__.Notify(107993)
	}
	__antithesis_instrumentation__.Notify(107975)
	return errors.New("failed to drain self")
}

func (nl *NodeLiveness) SetMembershipStatus(
	ctx context.Context, nodeID roachpb.NodeID, targetStatus livenesspb.MembershipStatus,
) (statusChanged bool, err error) {
	__antithesis_instrumentation__.Notify(107994)
	ctx = nl.ambientCtx.AnnotateCtx(ctx)

	attempt := func() (bool, error) {
		__antithesis_instrumentation__.Notify(107996)

		sem := nl.sem(nodeID)
		select {
		case sem <- struct{}{}:
			__antithesis_instrumentation__.Notify(108002)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(108003)
			return false, ctx.Err()
		}
		__antithesis_instrumentation__.Notify(107997)
		defer func() {
			__antithesis_instrumentation__.Notify(108004)
			<-sem
		}()
		__antithesis_instrumentation__.Notify(107998)

		var oldLiveness livenesspb.Liveness
		kv, err := nl.db.Get(ctx, keys.NodeLivenessKey(nodeID))
		if err != nil {
			__antithesis_instrumentation__.Notify(108005)
			return false, errors.Wrap(err, "unable to get liveness")
		} else {
			__antithesis_instrumentation__.Notify(108006)
		}
		__antithesis_instrumentation__.Notify(107999)
		if kv.Value == nil {
			__antithesis_instrumentation__.Notify(108007)

			return false, ErrMissingRecord
		} else {
			__antithesis_instrumentation__.Notify(108008)
		}
		__antithesis_instrumentation__.Notify(108000)
		if err := kv.Value.GetProto(&oldLiveness); err != nil {
			__antithesis_instrumentation__.Notify(108009)
			return false, errors.Wrap(err, "invalid liveness record")
		} else {
			__antithesis_instrumentation__.Notify(108010)
		}
		__antithesis_instrumentation__.Notify(108001)

		oldLivenessRec := Record{
			Liveness: oldLiveness,
			raw:      kv.Value.TagAndDataBytes(),
		}

		nl.maybeUpdate(ctx, oldLivenessRec)
		return nl.setMembershipStatusInternal(ctx, oldLivenessRec, targetStatus)
	}
	__antithesis_instrumentation__.Notify(107995)

	for {
		__antithesis_instrumentation__.Notify(108011)
		statusChanged, err := attempt()
		if errors.Is(err, errChangeMembershipStatusFailed) {
			__antithesis_instrumentation__.Notify(108013)

			continue
		} else {
			__antithesis_instrumentation__.Notify(108014)
		}
		__antithesis_instrumentation__.Notify(108012)
		return statusChanged, err
	}
}

func (nl *NodeLiveness) setDrainingInternal(
	ctx context.Context, oldLivenessRec Record, drain bool, reporter func(int, redact.SafeString),
) error {
	__antithesis_instrumentation__.Notify(108015)
	nodeID := nl.gossip.NodeID.Get()
	sem := nl.sem(nodeID)

	select {
	case sem <- struct{}{}:
		__antithesis_instrumentation__.Notify(108022)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(108023)
		return ctx.Err()
	}
	__antithesis_instrumentation__.Notify(108016)
	defer func() {
		__antithesis_instrumentation__.Notify(108024)
		<-sem
	}()
	__antithesis_instrumentation__.Notify(108017)

	if oldLivenessRec.Liveness == (livenesspb.Liveness{}) {
		__antithesis_instrumentation__.Notify(108025)
		return errors.AssertionFailedf("invalid old liveness record; found to be empty")
	} else {
		__antithesis_instrumentation__.Notify(108026)
	}
	__antithesis_instrumentation__.Notify(108018)

	newLiveness := oldLivenessRec.Liveness

	if reporter != nil && func() bool {
		__antithesis_instrumentation__.Notify(108027)
		return drain == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(108028)
		return !newLiveness.Draining == true
	}() == true {
		__antithesis_instrumentation__.Notify(108029)

		reporter(1, "liveness record")
	} else {
		__antithesis_instrumentation__.Notify(108030)
	}
	__antithesis_instrumentation__.Notify(108019)
	newLiveness.Draining = drain

	update := livenessUpdate{
		oldLiveness: oldLivenessRec.Liveness,
		newLiveness: newLiveness,
		oldRaw:      oldLivenessRec.raw,
		ignoreCache: true,
	}
	written, err := nl.updateLiveness(ctx, update, func(actual Record) error {
		__antithesis_instrumentation__.Notify(108031)
		nl.maybeUpdate(ctx, actual)

		if actual.Draining == update.newLiveness.Draining {
			__antithesis_instrumentation__.Notify(108033)
			return errNodeDrainingSet
		} else {
			__antithesis_instrumentation__.Notify(108034)
		}
		__antithesis_instrumentation__.Notify(108032)
		return errors.New("failed to update liveness record because record has changed")
	})
	__antithesis_instrumentation__.Notify(108020)
	if err != nil {
		__antithesis_instrumentation__.Notify(108035)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(108038)
			log.Infof(ctx, "updating liveness record: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(108039)
		}
		__antithesis_instrumentation__.Notify(108036)
		if errors.Is(err, errNodeDrainingSet) {
			__antithesis_instrumentation__.Notify(108040)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(108041)
		}
		__antithesis_instrumentation__.Notify(108037)
		return err
	} else {
		__antithesis_instrumentation__.Notify(108042)
	}
	__antithesis_instrumentation__.Notify(108021)

	nl.maybeUpdate(ctx, written)
	return nil
}

type livenessUpdate struct {
	newLiveness livenesspb.Liveness
	oldLiveness livenesspb.Liveness

	ignoreCache bool

	oldRaw []byte
}

func (nl *NodeLiveness) CreateLivenessRecord(ctx context.Context, nodeID roachpb.NodeID) error {
	__antithesis_instrumentation__.Notify(108043)
	for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(108046)

		liveness := livenesspb.Liveness{NodeID: nodeID, Epoch: 0}

		v := new(roachpb.Value)
		err := nl.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(108050)
			b := txn.NewBatch()
			key := keys.NodeLivenessKey(nodeID)
			if err := v.SetProto(&liveness); err != nil {
				__antithesis_instrumentation__.Notify(108052)
				log.Fatalf(ctx, "failed to marshall proto: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(108053)
			}
			__antithesis_instrumentation__.Notify(108051)

			b.CPut(key, v, nil)

			b.AddRawRequest(&roachpb.EndTxnRequest{
				Commit:     true,
				Require1PC: true,
			})
			return txn.Run(ctx, b)
		})
		__antithesis_instrumentation__.Notify(108047)

		if err == nil {
			__antithesis_instrumentation__.Notify(108054)

			log.Infof(ctx, "created liveness record for n%d", nodeID)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(108055)
		}
		__antithesis_instrumentation__.Notify(108048)
		if !isErrRetryLiveness(ctx, err) {
			__antithesis_instrumentation__.Notify(108056)
			return err
		} else {
			__antithesis_instrumentation__.Notify(108057)
		}
		__antithesis_instrumentation__.Notify(108049)
		log.VEventf(ctx, 2, "failed to create liveness record for node %d, because of %s. retrying...", nodeID, err)
	}
	__antithesis_instrumentation__.Notify(108044)
	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(108058)
		return err
	} else {
		__antithesis_instrumentation__.Notify(108059)
	}
	__antithesis_instrumentation__.Notify(108045)
	return errors.AssertionFailedf("unexpected problem while creating liveness record for node %d", nodeID)
}

func (nl *NodeLiveness) setMembershipStatusInternal(
	ctx context.Context, oldLivenessRec Record, targetStatus livenesspb.MembershipStatus,
) (statusChanged bool, err error) {
	__antithesis_instrumentation__.Notify(108060)
	if oldLivenessRec.Liveness == (livenesspb.Liveness{}) {
		__antithesis_instrumentation__.Notify(108065)
		return false, errors.AssertionFailedf("invalid old liveness record; found to be empty")
	} else {
		__antithesis_instrumentation__.Notify(108066)
	}
	__antithesis_instrumentation__.Notify(108061)

	newLiveness := oldLivenessRec.Liveness
	newLiveness.Membership = targetStatus
	if oldLivenessRec.Membership == newLiveness.Membership {
		__antithesis_instrumentation__.Notify(108067)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(108068)
		if oldLivenessRec.Membership.Decommissioned() && func() bool {
			__antithesis_instrumentation__.Notify(108069)
			return newLiveness.Membership.Decommissioning() == true
		}() == true {
			__antithesis_instrumentation__.Notify(108070)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(108071)
		}
	}
	__antithesis_instrumentation__.Notify(108062)

	if err := livenesspb.ValidateTransition(oldLivenessRec.Liveness, newLiveness); err != nil {
		__antithesis_instrumentation__.Notify(108072)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(108073)
	}
	__antithesis_instrumentation__.Notify(108063)

	update := livenessUpdate{
		newLiveness: newLiveness,
		oldLiveness: oldLivenessRec.Liveness,
		oldRaw:      oldLivenessRec.raw,
		ignoreCache: true,
	}
	statusChanged = true
	if _, err := nl.updateLiveness(ctx, update, func(actual Record) error {
		__antithesis_instrumentation__.Notify(108074)
		if actual.Membership != update.newLiveness.Membership {
			__antithesis_instrumentation__.Notify(108076)

			return errChangeMembershipStatusFailed
		} else {
			__antithesis_instrumentation__.Notify(108077)
		}
		__antithesis_instrumentation__.Notify(108075)

		statusChanged = false
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(108078)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(108079)
	}
	__antithesis_instrumentation__.Notify(108064)

	return statusChanged, nil
}

func (nl *NodeLiveness) GetLivenessThreshold() time.Duration {
	__antithesis_instrumentation__.Notify(108080)
	return nl.livenessThreshold
}

func (nl *NodeLiveness) IsLive(nodeID roachpb.NodeID) (bool, error) {
	__antithesis_instrumentation__.Notify(108081)
	liveness, ok := nl.GetLiveness(nodeID)
	if !ok {
		__antithesis_instrumentation__.Notify(108083)

		return false, ErrRecordCacheMiss
	} else {
		__antithesis_instrumentation__.Notify(108084)
	}
	__antithesis_instrumentation__.Notify(108082)

	return liveness.IsLive(nl.clock.Now().GoTime()), nil
}

func (nl *NodeLiveness) IsAvailable(nodeID roachpb.NodeID) bool {
	__antithesis_instrumentation__.Notify(108085)
	liveness, ok := nl.GetLiveness(nodeID)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(108086)
		return liveness.IsLive(nl.clock.Now().GoTime()) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(108087)
		return !liveness.Membership.Decommissioned() == true
	}() == true
}

func (nl *NodeLiveness) IsAvailableNotDraining(nodeID roachpb.NodeID) bool {
	__antithesis_instrumentation__.Notify(108088)
	liveness, ok := nl.GetLiveness(nodeID)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(108089)
		return liveness.IsLive(nl.clock.Now().GoTime()) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(108090)
		return !liveness.Membership.Decommissioning() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(108091)
		return !liveness.Membership.Decommissioned() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(108092)
		return !liveness.Draining == true
	}() == true
}

type NodeLivenessStartOptions struct {
	Stopper *stop.Stopper
	Engines []storage.Engine

	OnSelfLive HeartbeatCallback
}

func (nl *NodeLiveness) Start(ctx context.Context, opts NodeLivenessStartOptions) {
	__antithesis_instrumentation__.Notify(108093)
	log.VEventf(ctx, 1, "starting node liveness instance")
	retryOpts := base.DefaultRetryOptions()
	retryOpts.Closer = opts.Stopper.ShouldQuiesce()

	if len(opts.Engines) == 0 {
		__antithesis_instrumentation__.Notify(108095)

		log.Fatalf(ctx, "must supply at least one engine")
	} else {
		__antithesis_instrumentation__.Notify(108096)
	}
	__antithesis_instrumentation__.Notify(108094)

	nl.mu.Lock()
	nl.mu.onSelfLive = opts.OnSelfLive
	nl.mu.engines = opts.Engines
	nl.mu.Unlock()

	_ = opts.Stopper.RunAsyncTaskEx(ctx, stop.TaskOpts{TaskName: "liveness-hb", SpanOpt: stop.SterileRootSpan}, func(context.Context) {
		__antithesis_instrumentation__.Notify(108097)
		ambient := nl.ambientCtx
		ambient.AddLogTag("liveness-hb", nil)
		ctx, cancel := opts.Stopper.WithCancelOnQuiesce(context.Background())
		defer cancel()
		ctx, sp := ambient.AnnotateCtxWithSpan(ctx, "liveness heartbeat loop")
		defer sp.Finish()

		incrementEpoch := true
		heartbeatInterval := nl.livenessThreshold - nl.renewalDuration
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			__antithesis_instrumentation__.Notify(108098)
			select {
			case <-nl.heartbeatToken:
				__antithesis_instrumentation__.Notify(108101)
			case <-opts.Stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(108102)
				return
			}
			__antithesis_instrumentation__.Notify(108099)

			if err := contextutil.RunWithTimeout(ctx, "node liveness heartbeat", nl.renewalDuration,
				func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(108103)

					for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
						__antithesis_instrumentation__.Notify(108105)
						oldLiveness, ok := nl.Self()
						if !ok {
							__antithesis_instrumentation__.Notify(108108)
							nodeID := nl.gossip.NodeID.Get()
							liveness, err := nl.getLivenessFromKV(ctx, nodeID)
							if err != nil {
								__antithesis_instrumentation__.Notify(108110)
								log.Infof(ctx, "unable to get liveness record from KV: %s", err)
								if grpcutil.IsConnectionRejected(err) {
									__antithesis_instrumentation__.Notify(108112)
									return err
								} else {
									__antithesis_instrumentation__.Notify(108113)
								}
								__antithesis_instrumentation__.Notify(108111)
								continue
							} else {
								__antithesis_instrumentation__.Notify(108114)
							}
							__antithesis_instrumentation__.Notify(108109)
							oldLiveness = liveness
						} else {
							__antithesis_instrumentation__.Notify(108115)
						}
						__antithesis_instrumentation__.Notify(108106)
						if err := nl.heartbeatInternal(ctx, oldLiveness, incrementEpoch); err != nil {
							__antithesis_instrumentation__.Notify(108116)
							if errors.Is(err, ErrEpochIncremented) {
								__antithesis_instrumentation__.Notify(108118)
								log.Infof(ctx, "%s; retrying", err)
								continue
							} else {
								__antithesis_instrumentation__.Notify(108119)
							}
							__antithesis_instrumentation__.Notify(108117)
							return err
						} else {
							__antithesis_instrumentation__.Notify(108120)
						}
						__antithesis_instrumentation__.Notify(108107)
						incrementEpoch = false
						break
					}
					__antithesis_instrumentation__.Notify(108104)
					return nil
				}); err != nil {
				__antithesis_instrumentation__.Notify(108121)
				log.Warningf(ctx, heartbeatFailureLogFormat, err)
			} else {
				__antithesis_instrumentation__.Notify(108122)
			}
			__antithesis_instrumentation__.Notify(108100)

			nl.heartbeatToken <- struct{}{}
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(108123)
			case <-opts.Stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(108124)
				return
			}
		}
	})
}

const heartbeatFailureLogFormat = `failed node liveness heartbeat: %+v

An inability to maintain liveness will prevent a node from participating in a
cluster. If this problem persists, it may be a sign of resource starvation or
of network connectivity problems. For help troubleshooting, visit:

    https://www.cockroachlabs.com/docs/stable/cluster-setup-troubleshooting.html#node-liveness-issues

`

func (nl *NodeLiveness) PauseHeartbeatLoopForTest() func() {
	__antithesis_instrumentation__.Notify(108125)
	if swapped := atomic.CompareAndSwapUint32(&nl.heartbeatPaused, 0, 1); swapped {
		__antithesis_instrumentation__.Notify(108127)
		<-nl.heartbeatToken
	} else {
		__antithesis_instrumentation__.Notify(108128)
	}
	__antithesis_instrumentation__.Notify(108126)
	return func() {
		__antithesis_instrumentation__.Notify(108129)
		if swapped := atomic.CompareAndSwapUint32(&nl.heartbeatPaused, 1, 0); swapped {
			__antithesis_instrumentation__.Notify(108130)
			nl.heartbeatToken <- struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(108131)
		}
	}
}

func (nl *NodeLiveness) PauseSynchronousHeartbeatsForTest() func() {
	__antithesis_instrumentation__.Notify(108132)
	nl.selfSem <- struct{}{}
	nl.otherSem <- struct{}{}
	return func() {
		__antithesis_instrumentation__.Notify(108133)
		<-nl.selfSem
		<-nl.otherSem
	}
}

func (nl *NodeLiveness) PauseAllHeartbeatsForTest() func() {
	__antithesis_instrumentation__.Notify(108134)
	enableLoop := nl.PauseHeartbeatLoopForTest()
	enableSync := nl.PauseSynchronousHeartbeatsForTest()
	return func() {
		__antithesis_instrumentation__.Notify(108135)
		enableLoop()
		enableSync()
	}
}

var errNodeAlreadyLive = errors.New("node already live")

func (nl *NodeLiveness) Heartbeat(ctx context.Context, liveness livenesspb.Liveness) error {
	__antithesis_instrumentation__.Notify(108136)
	return nl.heartbeatInternal(ctx, liveness, false)
}

func (nl *NodeLiveness) heartbeatInternal(
	ctx context.Context, oldLiveness livenesspb.Liveness, incrementEpoch bool,
) (err error) {
	__antithesis_instrumentation__.Notify(108137)
	ctx, sp := tracing.EnsureChildSpan(ctx, nl.ambientCtx.Tracer, "liveness heartbeat")
	defer sp.Finish()
	defer func(start time.Time) {
		__antithesis_instrumentation__.Notify(108147)
		dur := timeutil.Since(start)
		nl.metrics.HeartbeatLatency.RecordValue(dur.Nanoseconds())
		if dur > time.Second {
			__antithesis_instrumentation__.Notify(108148)
			log.Warningf(ctx, "slow heartbeat took %s; err=%v", dur, err)
		} else {
			__antithesis_instrumentation__.Notify(108149)
		}
	}(timeutil.Now())
	__antithesis_instrumentation__.Notify(108138)

	beforeQueueTS := nl.clock.Now()
	minExpiration := beforeQueueTS.Add(nl.livenessThreshold.Nanoseconds(), 0).ToLegacyTimestamp()

	nl.metrics.HeartbeatsInFlight.Inc(1)
	defer nl.metrics.HeartbeatsInFlight.Dec(1)

	nodeID := nl.gossip.NodeID.Get()
	sem := nl.sem(nodeID)
	select {
	case sem <- struct{}{}:
		__antithesis_instrumentation__.Notify(108150)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(108151)
		return ctx.Err()
	}
	__antithesis_instrumentation__.Notify(108139)
	defer func() {
		__antithesis_instrumentation__.Notify(108152)
		<-sem
	}()
	__antithesis_instrumentation__.Notify(108140)

	if !incrementEpoch {
		__antithesis_instrumentation__.Notify(108153)
		curLiveness, ok := nl.Self()
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(108154)
			return minExpiration.Less(curLiveness.Expiration) == true
		}() == true {
			__antithesis_instrumentation__.Notify(108155)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(108156)
		}
	} else {
		__antithesis_instrumentation__.Notify(108157)
	}
	__antithesis_instrumentation__.Notify(108141)

	if oldLiveness == (livenesspb.Liveness{}) {
		__antithesis_instrumentation__.Notify(108158)
		return errors.AssertionFailedf("invalid old liveness record; found to be empty")
	} else {
		__antithesis_instrumentation__.Notify(108159)
	}
	__antithesis_instrumentation__.Notify(108142)

	newLiveness := oldLiveness
	if incrementEpoch {
		__antithesis_instrumentation__.Notify(108160)
		newLiveness.Epoch++
		newLiveness.Draining = false
	} else {
		__antithesis_instrumentation__.Notify(108161)
	}
	__antithesis_instrumentation__.Notify(108143)

	afterQueueTS := nl.clock.Now()
	newLiveness.Expiration = afterQueueTS.Add(nl.livenessThreshold.Nanoseconds(), 0).ToLegacyTimestamp()

	if newLiveness.Expiration.Less(oldLiveness.Expiration) {
		__antithesis_instrumentation__.Notify(108162)
		return errors.Errorf("proposed liveness update expires earlier than previous record")
	} else {
		__antithesis_instrumentation__.Notify(108163)
	}
	__antithesis_instrumentation__.Notify(108144)

	update := livenessUpdate{
		oldLiveness: oldLiveness,
		newLiveness: newLiveness,
	}
	written, err := nl.updateLiveness(ctx, update, func(actual Record) error {
		__antithesis_instrumentation__.Notify(108164)

		nl.maybeUpdate(ctx, actual)

		if actual.IsLive(nl.clock.Now().GoTime()) && func() bool {
			__antithesis_instrumentation__.Notify(108166)
			return !incrementEpoch == true
		}() == true {
			__antithesis_instrumentation__.Notify(108167)
			return errNodeAlreadyLive
		} else {
			__antithesis_instrumentation__.Notify(108168)
		}
		__antithesis_instrumentation__.Notify(108165)

		return ErrEpochIncremented
	})
	__antithesis_instrumentation__.Notify(108145)
	if err != nil {
		__antithesis_instrumentation__.Notify(108169)
		if errors.Is(err, errNodeAlreadyLive) {
			__antithesis_instrumentation__.Notify(108171)
			nl.metrics.HeartbeatSuccesses.Inc(1)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(108172)
		}
		__antithesis_instrumentation__.Notify(108170)
		nl.metrics.HeartbeatFailures.Inc()
		return err
	} else {
		__antithesis_instrumentation__.Notify(108173)
	}
	__antithesis_instrumentation__.Notify(108146)

	log.VEventf(ctx, 1, "heartbeat %+v", written.Expiration)
	nl.maybeUpdate(ctx, written)
	nl.metrics.HeartbeatSuccesses.Inc(1)
	return nil
}

func (nl *NodeLiveness) Self() (_ livenesspb.Liveness, ok bool) {
	__antithesis_instrumentation__.Notify(108174)
	rec, ok := nl.SelfEx()
	if !ok {
		__antithesis_instrumentation__.Notify(108176)
		return livenesspb.Liveness{}, false
	} else {
		__antithesis_instrumentation__.Notify(108177)
	}
	__antithesis_instrumentation__.Notify(108175)
	return rec.Liveness, true
}

func (nl *NodeLiveness) SelfEx() (_ Record, ok bool) {
	__antithesis_instrumentation__.Notify(108178)
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	return nl.getLivenessLocked(nl.gossip.NodeID.Get())
}

type IsLiveMapEntry struct {
	livenesspb.Liveness
	IsLive bool
}

type IsLiveMap map[roachpb.NodeID]IsLiveMapEntry

func (nl *NodeLiveness) GetIsLiveMap() IsLiveMap {
	__antithesis_instrumentation__.Notify(108179)
	lMap := IsLiveMap{}
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	now := nl.clock.Now().GoTime()
	for nID, l := range nl.mu.nodes {
		__antithesis_instrumentation__.Notify(108181)
		isLive := l.IsLive(now)
		if !isLive && func() bool {
			__antithesis_instrumentation__.Notify(108183)
			return !l.Membership.Active() == true
		}() == true {
			__antithesis_instrumentation__.Notify(108184)

			continue
		} else {
			__antithesis_instrumentation__.Notify(108185)
		}
		__antithesis_instrumentation__.Notify(108182)
		lMap[nID] = IsLiveMapEntry{
			Liveness: l.Liveness,
			IsLive:   isLive,
		}
	}
	__antithesis_instrumentation__.Notify(108180)
	return lMap
}

func (nl *NodeLiveness) GetLivenesses() []livenesspb.Liveness {
	__antithesis_instrumentation__.Notify(108186)
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	livenesses := make([]livenesspb.Liveness, 0, len(nl.mu.nodes))
	for _, l := range nl.mu.nodes {
		__antithesis_instrumentation__.Notify(108188)
		livenesses = append(livenesses, l.Liveness)
	}
	__antithesis_instrumentation__.Notify(108187)
	return livenesses
}

func (nl *NodeLiveness) GetLivenessesFromKV(ctx context.Context) ([]livenesspb.Liveness, error) {
	__antithesis_instrumentation__.Notify(108189)
	kvs, err := nl.db.Scan(ctx, keys.NodeLivenessPrefix, keys.NodeLivenessKeyMax, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(108192)
		return nil, errors.Wrap(err, "unable to get liveness")
	} else {
		__antithesis_instrumentation__.Notify(108193)
	}
	__antithesis_instrumentation__.Notify(108190)

	var results []livenesspb.Liveness
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(108194)
		if kv.Value == nil {
			__antithesis_instrumentation__.Notify(108197)
			return nil, errors.AssertionFailedf("missing liveness record")
		} else {
			__antithesis_instrumentation__.Notify(108198)
		}
		__antithesis_instrumentation__.Notify(108195)
		var liveness livenesspb.Liveness
		if err := kv.Value.GetProto(&liveness); err != nil {
			__antithesis_instrumentation__.Notify(108199)
			return nil, errors.Wrap(err, "invalid liveness record")
		} else {
			__antithesis_instrumentation__.Notify(108200)
		}
		__antithesis_instrumentation__.Notify(108196)

		livenessRec := Record{
			Liveness: liveness,
			raw:      kv.Value.TagAndDataBytes(),
		}

		nl.maybeUpdate(ctx, livenessRec)

		results = append(results, liveness)
	}
	__antithesis_instrumentation__.Notify(108191)

	return results, nil
}

func (nl *NodeLiveness) GetLiveness(nodeID roachpb.NodeID) (_ Record, ok bool) {
	__antithesis_instrumentation__.Notify(108201)
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	return nl.getLivenessLocked(nodeID)
}

func (nl *NodeLiveness) getLivenessLocked(nodeID roachpb.NodeID) (_ Record, ok bool) {
	__antithesis_instrumentation__.Notify(108202)
	if l, ok := nl.mu.nodes[nodeID]; ok {
		__antithesis_instrumentation__.Notify(108204)
		return l, true
	} else {
		__antithesis_instrumentation__.Notify(108205)
	}
	__antithesis_instrumentation__.Notify(108203)
	return Record{}, false
}

func (nl *NodeLiveness) getLivenessFromKV(
	ctx context.Context, nodeID roachpb.NodeID,
) (livenesspb.Liveness, error) {
	__antithesis_instrumentation__.Notify(108206)
	livenessRec, err := nl.getLivenessRecordFromKV(ctx, nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(108208)
		return livenesspb.Liveness{}, err
	} else {
		__antithesis_instrumentation__.Notify(108209)
	}
	__antithesis_instrumentation__.Notify(108207)
	return livenessRec.Liveness, nil
}

func (nl *NodeLiveness) getLivenessRecordFromKV(
	ctx context.Context, nodeID roachpb.NodeID,
) (Record, error) {
	__antithesis_instrumentation__.Notify(108210)
	kv, err := nl.db.Get(ctx, keys.NodeLivenessKey(nodeID))
	if err != nil {
		__antithesis_instrumentation__.Notify(108214)
		return Record{}, errors.Wrap(err, "unable to get liveness")
	} else {
		__antithesis_instrumentation__.Notify(108215)
	}
	__antithesis_instrumentation__.Notify(108211)
	if kv.Value == nil {
		__antithesis_instrumentation__.Notify(108216)
		return Record{}, errors.AssertionFailedf("missing liveness record")
	} else {
		__antithesis_instrumentation__.Notify(108217)
	}
	__antithesis_instrumentation__.Notify(108212)
	var liveness livenesspb.Liveness
	if err := kv.Value.GetProto(&liveness); err != nil {
		__antithesis_instrumentation__.Notify(108218)
		return Record{}, errors.Wrap(err, "invalid liveness record")
	} else {
		__antithesis_instrumentation__.Notify(108219)
	}
	__antithesis_instrumentation__.Notify(108213)

	livenessRec := Record{
		Liveness: liveness,
		raw:      kv.Value.TagAndDataBytes(),
	}

	nl.maybeUpdate(ctx, livenessRec)
	return livenessRec, nil
}

func (nl *NodeLiveness) IncrementEpoch(ctx context.Context, liveness livenesspb.Liveness) error {
	__antithesis_instrumentation__.Notify(108220)

	sem := nl.sem(liveness.NodeID)
	select {
	case sem <- struct{}{}:
		__antithesis_instrumentation__.Notify(108226)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(108227)
		return ctx.Err()
	}
	__antithesis_instrumentation__.Notify(108221)
	defer func() {
		__antithesis_instrumentation__.Notify(108228)
		<-sem
	}()
	__antithesis_instrumentation__.Notify(108222)

	if liveness.IsLive(nl.clock.Now().GoTime()) {
		__antithesis_instrumentation__.Notify(108229)
		return errors.Errorf("cannot increment epoch on live node: %+v", liveness)
	} else {
		__antithesis_instrumentation__.Notify(108230)
	}
	__antithesis_instrumentation__.Notify(108223)

	update := livenessUpdate{
		newLiveness: liveness,
		oldLiveness: liveness,
	}
	update.newLiveness.Epoch++

	written, err := nl.updateLiveness(ctx, update, func(actual Record) error {
		__antithesis_instrumentation__.Notify(108231)
		nl.maybeUpdate(ctx, actual)

		if actual.Epoch > liveness.Epoch {
			__antithesis_instrumentation__.Notify(108233)
			return ErrEpochAlreadyIncremented
		} else {
			__antithesis_instrumentation__.Notify(108234)
			if actual.Epoch < liveness.Epoch {
				__antithesis_instrumentation__.Notify(108235)
				return errors.Errorf("unexpected liveness epoch %d; expected >= %d", actual.Epoch, liveness.Epoch)
			} else {
				__antithesis_instrumentation__.Notify(108236)
			}
		}
		__antithesis_instrumentation__.Notify(108232)
		return errors.Errorf("mismatch incrementing epoch for %+v; actual is %+v", liveness, actual)
	})
	__antithesis_instrumentation__.Notify(108224)
	if err != nil {
		__antithesis_instrumentation__.Notify(108237)
		return err
	} else {
		__antithesis_instrumentation__.Notify(108238)
	}
	__antithesis_instrumentation__.Notify(108225)

	log.Infof(ctx, "incremented n%d liveness epoch to %d", written.NodeID, written.Epoch)
	nl.maybeUpdate(ctx, written)
	nl.metrics.EpochIncrements.Inc()
	return nil
}

func (nl *NodeLiveness) Metrics() Metrics {
	__antithesis_instrumentation__.Notify(108239)
	return nl.metrics
}

func (nl *NodeLiveness) RegisterCallback(cb IsLiveCallback) {
	__antithesis_instrumentation__.Notify(108240)
	nl.mu.Lock()
	defer nl.mu.Unlock()
	nl.mu.onIsLive = append(nl.mu.onIsLive, cb)
}

func (nl *NodeLiveness) updateLiveness(
	ctx context.Context, update livenessUpdate, handleCondFailed func(actual Record) error,
) (Record, error) {
	__antithesis_instrumentation__.Notify(108241)
	for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(108244)
		nl.mu.RLock()
		engines := nl.mu.engines
		nl.mu.RUnlock()
		for _, eng := range engines {
			__antithesis_instrumentation__.Notify(108247)

			if err := storage.WriteSyncNoop(ctx, eng); err != nil {
				__antithesis_instrumentation__.Notify(108248)
				return Record{}, errors.Wrapf(err, "couldn't update node liveness because disk write failed")
			} else {
				__antithesis_instrumentation__.Notify(108249)
			}
		}
		__antithesis_instrumentation__.Notify(108245)
		written, err := nl.updateLivenessAttempt(ctx, update, handleCondFailed)
		if err != nil {
			__antithesis_instrumentation__.Notify(108250)
			if errors.HasType(err, (*errRetryLiveness)(nil)) {
				__antithesis_instrumentation__.Notify(108252)
				log.Infof(ctx, "retrying liveness update after %s", err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(108253)
			}
			__antithesis_instrumentation__.Notify(108251)
			return Record{}, err
		} else {
			__antithesis_instrumentation__.Notify(108254)
		}
		__antithesis_instrumentation__.Notify(108246)
		return written, nil
	}
	__antithesis_instrumentation__.Notify(108242)
	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(108255)
		return Record{}, err
	} else {
		__antithesis_instrumentation__.Notify(108256)
	}
	__antithesis_instrumentation__.Notify(108243)
	panic("unreachable; should retry until ctx canceled")
}

func (nl *NodeLiveness) updateLivenessAttempt(
	ctx context.Context, update livenessUpdate, handleCondFailed func(actual Record) error,
) (Record, error) {
	__antithesis_instrumentation__.Notify(108257)
	var oldRaw []byte
	if update.ignoreCache {
		__antithesis_instrumentation__.Notify(108261)

		oldRaw = update.oldRaw
	} else {
		__antithesis_instrumentation__.Notify(108262)

		if update.oldRaw != nil {
			__antithesis_instrumentation__.Notify(108266)
			log.Fatalf(ctx, "unexpected oldRaw when ignoreCache not specified")
		} else {
			__antithesis_instrumentation__.Notify(108267)
		}
		__antithesis_instrumentation__.Notify(108263)

		l, ok := nl.GetLiveness(update.newLiveness.NodeID)
		if !ok {
			__antithesis_instrumentation__.Notify(108268)

			return Record{}, ErrRecordCacheMiss
		} else {
			__antithesis_instrumentation__.Notify(108269)
		}
		__antithesis_instrumentation__.Notify(108264)
		if l.Liveness != update.oldLiveness {
			__antithesis_instrumentation__.Notify(108270)
			return Record{}, handleCondFailed(l)
		} else {
			__antithesis_instrumentation__.Notify(108271)
		}
		__antithesis_instrumentation__.Notify(108265)
		oldRaw = l.raw
	}
	__antithesis_instrumentation__.Notify(108258)

	var v *roachpb.Value
	if err := nl.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(108272)

		v = new(roachpb.Value)

		b := txn.NewBatch()
		key := keys.NodeLivenessKey(update.newLiveness.NodeID)
		if err := v.SetProto(&update.newLiveness); err != nil {
			__antithesis_instrumentation__.Notify(108274)
			log.Fatalf(ctx, "failed to marshall proto: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(108275)
		}
		__antithesis_instrumentation__.Notify(108273)
		b.CPut(key, v, oldRaw)

		b.AddRawRequest(&roachpb.EndTxnRequest{
			Commit:     true,
			Require1PC: true,
			InternalCommitTrigger: &roachpb.InternalCommitTrigger{
				ModifiedSpanTrigger: &roachpb.ModifiedSpanTrigger{
					NodeLivenessSpan: &roachpb.Span{
						Key:    key,
						EndKey: key.Next(),
					},
				},
			},
		})
		return txn.Run(ctx, b)
	}); err != nil {
		__antithesis_instrumentation__.Notify(108276)
		if tErr := (*roachpb.ConditionFailedError)(nil); errors.As(err, &tErr) {
			__antithesis_instrumentation__.Notify(108278)
			if tErr.ActualValue == nil {
				__antithesis_instrumentation__.Notify(108281)
				return Record{}, handleCondFailed(Record{})
			} else {
				__antithesis_instrumentation__.Notify(108282)
			}
			__antithesis_instrumentation__.Notify(108279)
			var actualLiveness livenesspb.Liveness
			if err := tErr.ActualValue.GetProto(&actualLiveness); err != nil {
				__antithesis_instrumentation__.Notify(108283)
				return Record{}, errors.Wrapf(err, "couldn't update node liveness from CPut actual value")
			} else {
				__antithesis_instrumentation__.Notify(108284)
			}
			__antithesis_instrumentation__.Notify(108280)
			return Record{}, handleCondFailed(Record{Liveness: actualLiveness, raw: tErr.ActualValue.TagAndDataBytes()})
		} else {
			__antithesis_instrumentation__.Notify(108285)
			if isErrRetryLiveness(ctx, err) {
				__antithesis_instrumentation__.Notify(108286)
				return Record{}, &errRetryLiveness{err}
			} else {
				__antithesis_instrumentation__.Notify(108287)
			}
		}
		__antithesis_instrumentation__.Notify(108277)
		return Record{}, err
	} else {
		__antithesis_instrumentation__.Notify(108288)
	}
	__antithesis_instrumentation__.Notify(108259)

	nl.mu.RLock()
	cb := nl.mu.onSelfLive
	nl.mu.RUnlock()
	if cb != nil {
		__antithesis_instrumentation__.Notify(108289)
		cb(ctx)
	} else {
		__antithesis_instrumentation__.Notify(108290)
	}
	__antithesis_instrumentation__.Notify(108260)
	return Record{Liveness: update.newLiveness, raw: v.TagAndDataBytes()}, nil
}

func (nl *NodeLiveness) maybeUpdate(ctx context.Context, newLivenessRec Record) {
	__antithesis_instrumentation__.Notify(108291)
	if newLivenessRec.Liveness == (livenesspb.Liveness{}) {
		__antithesis_instrumentation__.Notify(108297)
		log.Fatal(ctx, "invalid new liveness record; found to be empty")
	} else {
		__antithesis_instrumentation__.Notify(108298)
	}
	__antithesis_instrumentation__.Notify(108292)

	var shouldReplace bool
	nl.mu.Lock()
	oldLivenessRec, ok := nl.getLivenessLocked(newLivenessRec.NodeID)
	if !ok {
		__antithesis_instrumentation__.Notify(108299)
		shouldReplace = true
	} else {
		__antithesis_instrumentation__.Notify(108300)
		shouldReplace = shouldReplaceLiveness(ctx, oldLivenessRec, newLivenessRec)
	}
	__antithesis_instrumentation__.Notify(108293)

	var onIsLive []IsLiveCallback
	if shouldReplace {
		__antithesis_instrumentation__.Notify(108301)
		nl.mu.nodes[newLivenessRec.NodeID] = newLivenessRec
		onIsLive = append(onIsLive, nl.mu.onIsLive...)
	} else {
		__antithesis_instrumentation__.Notify(108302)
	}
	__antithesis_instrumentation__.Notify(108294)
	nl.mu.Unlock()

	if !shouldReplace {
		__antithesis_instrumentation__.Notify(108303)
		return
	} else {
		__antithesis_instrumentation__.Notify(108304)
	}
	__antithesis_instrumentation__.Notify(108295)

	now := nl.clock.Now().GoTime()
	if !oldLivenessRec.IsLive(now) && func() bool {
		__antithesis_instrumentation__.Notify(108305)
		return newLivenessRec.IsLive(now) == true
	}() == true {
		__antithesis_instrumentation__.Notify(108306)
		for _, fn := range onIsLive {
			__antithesis_instrumentation__.Notify(108307)
			fn(newLivenessRec.Liveness)
		}
	} else {
		__antithesis_instrumentation__.Notify(108308)
	}
	__antithesis_instrumentation__.Notify(108296)
	if newLivenessRec.Membership.Decommissioned() && func() bool {
		__antithesis_instrumentation__.Notify(108309)
		return nl.onNodeDecommissioned != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(108310)
		nl.onNodeDecommissioned(newLivenessRec.Liveness)
	} else {
		__antithesis_instrumentation__.Notify(108311)
	}
}

func shouldReplaceLiveness(ctx context.Context, old, new Record) bool {
	__antithesis_instrumentation__.Notify(108312)
	oldL, newL := old.Liveness, new.Liveness
	if (oldL == livenesspb.Liveness{}) {
		__antithesis_instrumentation__.Notify(108315)
		log.Fatal(ctx, "invalid old liveness record; found to be empty")
	} else {
		__antithesis_instrumentation__.Notify(108316)
	}
	__antithesis_instrumentation__.Notify(108313)

	if cmp := oldL.Compare(newL); cmp != 0 {
		__antithesis_instrumentation__.Notify(108317)
		return cmp < 0
	} else {
		__antithesis_instrumentation__.Notify(108318)
	}
	__antithesis_instrumentation__.Notify(108314)

	return oldL.Draining != newL.Draining || func() bool {
		__antithesis_instrumentation__.Notify(108319)
		return oldL.Membership != newL.Membership == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(108320)
		return (oldL.Equal(newL) && func() bool {
			__antithesis_instrumentation__.Notify(108321)
			return !bytes.Equal(old.raw, new.raw) == true
		}() == true) == true
	}() == true
}

func (nl *NodeLiveness) livenessGossipUpdate(_ string, content roachpb.Value) {
	__antithesis_instrumentation__.Notify(108322)
	var liveness livenesspb.Liveness
	ctx := context.TODO()
	if err := content.GetProto(&liveness); err != nil {
		__antithesis_instrumentation__.Notify(108324)
		log.Errorf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(108325)
	}
	__antithesis_instrumentation__.Notify(108323)

	nl.maybeUpdate(ctx, Record{Liveness: liveness, raw: content.TagAndDataBytes()})
}

func (nl *NodeLiveness) numLiveNodes() int64 {
	__antithesis_instrumentation__.Notify(108326)
	selfID := nl.gossip.NodeID.Get()
	if selfID == 0 {
		__antithesis_instrumentation__.Notify(108331)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(108332)
	}
	__antithesis_instrumentation__.Notify(108327)

	nl.mu.RLock()
	defer nl.mu.RUnlock()

	self, ok := nl.getLivenessLocked(selfID)
	if !ok {
		__antithesis_instrumentation__.Notify(108333)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(108334)
	}
	__antithesis_instrumentation__.Notify(108328)
	now := nl.clock.Now().GoTime()

	if !self.IsLive(now) {
		__antithesis_instrumentation__.Notify(108335)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(108336)
	}
	__antithesis_instrumentation__.Notify(108329)
	var liveNodes int64
	for _, l := range nl.mu.nodes {
		__antithesis_instrumentation__.Notify(108337)
		if l.IsLive(now) {
			__antithesis_instrumentation__.Notify(108338)
			liveNodes++
		} else {
			__antithesis_instrumentation__.Notify(108339)
		}
	}
	__antithesis_instrumentation__.Notify(108330)
	return liveNodes
}

func (nl *NodeLiveness) GetNodeCount() int {
	__antithesis_instrumentation__.Notify(108340)
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	var count int
	for _, l := range nl.mu.nodes {
		__antithesis_instrumentation__.Notify(108342)
		if l.Membership.Active() {
			__antithesis_instrumentation__.Notify(108343)
			count++
		} else {
			__antithesis_instrumentation__.Notify(108344)
		}
	}
	__antithesis_instrumentation__.Notify(108341)
	return count
}

func (nl *NodeLiveness) TestingSetDrainingInternal(
	ctx context.Context, liveness Record, drain bool,
) error {
	__antithesis_instrumentation__.Notify(108345)
	return nl.setDrainingInternal(ctx, liveness, drain, nil)
}

func (nl *NodeLiveness) TestingSetDecommissioningInternal(
	ctx context.Context, oldLivenessRec Record, targetStatus livenesspb.MembershipStatus,
) (changeCommitted bool, err error) {
	__antithesis_instrumentation__.Notify(108346)
	return nl.setMembershipStatusInternal(ctx, oldLivenessRec, targetStatus)
}
