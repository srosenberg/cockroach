package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/circuit"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3"
)

type replicaInCircuitBreaker interface {
	Clock() *hlc.Clock
	Desc() *roachpb.RangeDescriptor
	Send(context.Context, roachpb.BatchRequest) (*roachpb.BatchResponse, *roachpb.Error)
	slowReplicationThreshold(ba *roachpb.BatchRequest) (time.Duration, bool)
	replicaUnavailableError(err error) error
	poisonInflightLatches(err error)
}

var defaultReplicaCircuitBreakerSlowReplicationThreshold = envutil.EnvOrDefaultDuration(
	"COCKROACH_REPLICA_CIRCUIT_BREAKER_SLOW_REPLICATION_THRESHOLD",

	4*base.SlowRequestThreshold,
)

var replicaCircuitBreakerSlowReplicationThreshold = settings.RegisterPublicDurationSettingWithExplicitUnit(
	settings.SystemOnly,
	"kv.replica_circuit_breaker.slow_replication_threshold",
	"duration after which slow proposals trip the per-Replica circuit breaker (zero duration disables breakers)",
	defaultReplicaCircuitBreakerSlowReplicationThreshold,
	func(d time.Duration) error {
		__antithesis_instrumentation__.Notify(115698)

		const min = 500 * time.Millisecond
		if d == 0 {
			__antithesis_instrumentation__.Notify(115701)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(115702)
		}
		__antithesis_instrumentation__.Notify(115699)
		if d <= min {
			__antithesis_instrumentation__.Notify(115703)
			return errors.Errorf("must specify a minimum of %s", min)
		} else {
			__antithesis_instrumentation__.Notify(115704)
		}
		__antithesis_instrumentation__.Notify(115700)
		return nil
	},
)

var telemetryTripAsync = telemetry.GetCounterOnce("kv.replica_circuit_breaker.num_tripped_events")

type replicaCircuitBreaker struct {
	ambCtx  log.AmbientContext
	stopper *stop.Stopper
	r       replicaInCircuitBreaker
	st      *cluster.Settings
	wrapped *circuit.Breaker

	versionIsActive int32
}

func (br *replicaCircuitBreaker) HasMark(err error) bool {
	__antithesis_instrumentation__.Notify(115705)
	return br.wrapped.HasMark(err)
}

func (br *replicaCircuitBreaker) canEnable() bool {
	__antithesis_instrumentation__.Notify(115706)
	b := atomic.LoadInt32(&br.versionIsActive) == 1
	if b {
		__antithesis_instrumentation__.Notify(115709)
		return true
	} else {
		__antithesis_instrumentation__.Notify(115710)
	}
	__antithesis_instrumentation__.Notify(115707)

	if br.st.Version.IsActive(context.Background(), clusterversion.ProbeRequest) {
		__antithesis_instrumentation__.Notify(115711)
		atomic.StoreInt32(&br.versionIsActive, 1)
		return true
	} else {
		__antithesis_instrumentation__.Notify(115712)
	}
	__antithesis_instrumentation__.Notify(115708)
	return false
}

func (br *replicaCircuitBreaker) enabled() bool {
	__antithesis_instrumentation__.Notify(115713)
	return replicaCircuitBreakerSlowReplicationThreshold.Get(&br.st.SV) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(115714)
		return br.canEnable() == true
	}() == true
}

func (br *replicaCircuitBreaker) TripAsync(err error) {
	__antithesis_instrumentation__.Notify(115715)
	if !br.enabled() {
		__antithesis_instrumentation__.Notify(115717)
		return
	} else {
		__antithesis_instrumentation__.Notify(115718)
	}
	__antithesis_instrumentation__.Notify(115716)

	_ = br.stopper.RunAsyncTask(
		br.ambCtx.AnnotateCtx(context.Background()), "trip-breaker",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(115719)
			br.tripSync(err)
		},
	)
}

func (br *replicaCircuitBreaker) tripSync(err error) {
	__antithesis_instrumentation__.Notify(115720)
	br.wrapped.Report(br.r.replicaUnavailableError(err))
}

type signaller interface {
	Err() error
	C() <-chan struct{}
}

type neverTripSignaller struct{}

func (s neverTripSignaller) Err() error { __antithesis_instrumentation__.Notify(115721); return nil }
func (s neverTripSignaller) C() <-chan struct{} {
	__antithesis_instrumentation__.Notify(115722)
	return nil
}

func (br *replicaCircuitBreaker) Signal() signaller {
	__antithesis_instrumentation__.Notify(115723)
	if !br.enabled() {
		__antithesis_instrumentation__.Notify(115725)
		return neverTripSignaller{}
	} else {
		__antithesis_instrumentation__.Notify(115726)
	}
	__antithesis_instrumentation__.Notify(115724)
	return br.wrapped.Signal()
}

func newReplicaCircuitBreaker(
	cs *cluster.Settings,
	stopper *stop.Stopper,
	ambientCtx log.AmbientContext,
	r replicaInCircuitBreaker,
	onTrip func(),
	onReset func(),
) *replicaCircuitBreaker {
	__antithesis_instrumentation__.Notify(115727)
	br := &replicaCircuitBreaker{
		stopper: stopper,
		ambCtx:  ambientCtx,
		r:       r,
		st:      cs,
	}
	br.wrapped = circuit.NewBreaker(circuit.Options{
		Name:       "breaker",
		AsyncProbe: br.asyncProbe,
		EventHandler: &replicaCircuitBreakerLogger{
			EventHandler: &circuit.EventLogger{
				Log: func(buf redact.StringBuilder) {
					__antithesis_instrumentation__.Notify(115729)
					log.Infof(ambientCtx.AnnotateCtx(context.Background()), "%s", buf)
				},
			},
			onTrip:  onTrip,
			onReset: onReset,
		},
	})
	__antithesis_instrumentation__.Notify(115728)

	return br
}

type replicaCircuitBreakerLogger struct {
	circuit.EventHandler
	onTrip  func()
	onReset func()
}

func (r replicaCircuitBreakerLogger) OnTrip(br *circuit.Breaker, prev, cur error) {
	__antithesis_instrumentation__.Notify(115730)
	if prev == nil {
		__antithesis_instrumentation__.Notify(115732)
		r.onTrip()
	} else {
		__antithesis_instrumentation__.Notify(115733)
	}
	__antithesis_instrumentation__.Notify(115731)
	r.EventHandler.OnTrip(br, prev, cur)
}

func (r replicaCircuitBreakerLogger) OnReset(br *circuit.Breaker) {
	__antithesis_instrumentation__.Notify(115734)
	r.onReset()
	r.EventHandler.OnReset(br)
}

func (br *replicaCircuitBreaker) asyncProbe(report func(error), done func()) {
	__antithesis_instrumentation__.Notify(115735)
	bgCtx := br.ambCtx.AnnotateCtx(context.Background())
	if err := br.stopper.RunAsyncTask(bgCtx, "replica-probe", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(115736)
		defer done()

		if !br.enabled() {
			__antithesis_instrumentation__.Notify(115739)
			report(nil)
			return
		} else {
			__antithesis_instrumentation__.Notify(115740)
		}
		__antithesis_instrumentation__.Notify(115737)

		brErr := br.Signal().Err()
		if brErr == nil {
			__antithesis_instrumentation__.Notify(115741)

			return
		} else {
			__antithesis_instrumentation__.Notify(115742)
		}
		__antithesis_instrumentation__.Notify(115738)

		br.r.poisonInflightLatches(brErr)
		err := sendProbe(ctx, br.r)
		report(err)
	}); err != nil {
		__antithesis_instrumentation__.Notify(115743)
		done()
	} else {
		__antithesis_instrumentation__.Notify(115744)
	}
}

func sendProbe(ctx context.Context, r replicaInCircuitBreaker) error {
	__antithesis_instrumentation__.Notify(115745)

	desc := r.Desc()
	if !desc.IsInitialized() {
		__antithesis_instrumentation__.Notify(115749)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(115750)
	}
	__antithesis_instrumentation__.Notify(115746)
	ba := roachpb.BatchRequest{}
	ba.Timestamp = r.Clock().Now()
	ba.RangeID = r.Desc().RangeID
	probeReq := &roachpb.ProbeRequest{}
	probeReq.Key = desc.StartKey.AsRawKey()
	ba.Add(probeReq)
	thresh, ok := r.slowReplicationThreshold(&ba)
	if !ok {
		__antithesis_instrumentation__.Notify(115751)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(115752)
	}
	__antithesis_instrumentation__.Notify(115747)
	if err := contextutil.RunWithTimeout(ctx, "probe", thresh,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(115753)
			_, pErr := r.Send(ctx, ba)
			return pErr.GoError()
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(115754)
		return r.replicaUnavailableError(err)
	} else {
		__antithesis_instrumentation__.Notify(115755)
	}
	__antithesis_instrumentation__.Notify(115748)
	return nil
}

func replicaUnavailableError(
	err error,
	desc *roachpb.RangeDescriptor,
	replDesc roachpb.ReplicaDescriptor,
	lm liveness.IsLiveMap,
	rs *raft.Status,
) error {
	__antithesis_instrumentation__.Notify(115756)
	nonLiveRepls := roachpb.MakeReplicaSet(nil)
	for _, rDesc := range desc.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(115760)
		if lm[rDesc.NodeID].IsLive {
			__antithesis_instrumentation__.Notify(115762)
			continue
		} else {
			__antithesis_instrumentation__.Notify(115763)
		}
		__antithesis_instrumentation__.Notify(115761)
		nonLiveRepls.AddReplica(rDesc)
	}
	__antithesis_instrumentation__.Notify(115757)

	canMakeProgress := desc.Replicas().CanMakeProgress(
		func(replDesc roachpb.ReplicaDescriptor) bool {
			__antithesis_instrumentation__.Notify(115764)
			return lm[replDesc.NodeID].IsLive
		},
	)
	__antithesis_instrumentation__.Notify(115758)

	var _ redact.SafeFormatter = nonLiveRepls
	var _ redact.SafeFormatter = desc
	var _ redact.SafeFormatter = replDesc

	if len(nonLiveRepls.AsProto()) > 0 {
		__antithesis_instrumentation__.Notify(115765)
		err = errors.Wrapf(err, "replicas on non-live nodes: %v (lost quorum: %t)", nonLiveRepls, !canMakeProgress)
	} else {
		__antithesis_instrumentation__.Notify(115766)
	}
	__antithesis_instrumentation__.Notify(115759)

	err = errors.Wrapf(
		err,
		"raft status: %+v", redact.Safe(rs),
	)

	return roachpb.NewReplicaUnavailableError(err, desc, replDesc)
}

func (r *Replica) replicaUnavailableError(err error) error {
	__antithesis_instrumentation__.Notify(115767)
	desc := r.Desc()
	replDesc, _ := desc.GetReplicaDescriptor(r.store.StoreID())

	isLiveMap, _ := r.store.livenessMap.Load().(liveness.IsLiveMap)
	return replicaUnavailableError(err, desc, replDesc, isLiveMap, r.RaftStatus())
}
