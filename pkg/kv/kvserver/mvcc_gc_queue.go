package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/gc"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/intentresolver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const (
	mvccGCQueueTimerDuration = 1 * time.Second

	mvccGCQueueTimeout = 10 * time.Minute

	mvccGCQueueIntentBatchTimeout = 2 * time.Minute

	mvccGCQueueIntentCooldownDuration = 2 * time.Hour

	intentAgeNormalization = 8 * time.Hour

	mvccGCKeyScoreThreshold    = 2
	mvccGCIntentScoreThreshold = 1

	probablyLargeAbortSpanSysCountThreshold = 10000
	largeAbortSpanBytesThreshold            = 16 * (1 << 20)
)

func largeAbortSpan(ms enginepb.MVCCStats) bool {
	__antithesis_instrumentation__.Notify(110287)

	definitelyLargeAbortSpan := ms.AbortSpanBytes >= largeAbortSpanBytesThreshold
	probablyLargeAbortSpan := ms.SysBytes >= largeAbortSpanBytesThreshold && func() bool {
		__antithesis_instrumentation__.Notify(110288)
		return ms.SysCount >= probablyLargeAbortSpanSysCountThreshold == true
	}() == true
	return definitelyLargeAbortSpan || func() bool {
		__antithesis_instrumentation__.Notify(110289)
		return probablyLargeAbortSpan == true
	}() == true
}

type mvccGCQueue struct {
	*baseQueue
}

func newMVCCGCQueue(store *Store) *mvccGCQueue {
	__antithesis_instrumentation__.Notify(110290)
	mgcq := &mvccGCQueue{}
	mgcq.baseQueue = newBaseQueue(
		"mvccGC", mgcq, store,
		queueConfig{
			maxSize:              defaultQueueMaxSize,
			needsLease:           true,
			needsSystemConfig:    true,
			acceptsUnsplitRanges: false,
			processTimeoutFunc: func(st *cluster.Settings, _ replicaInQueue) time.Duration {
				__antithesis_instrumentation__.Notify(110292)
				timeout := mvccGCQueueTimeout
				if d := queueGuaranteedProcessingTimeBudget.Get(&st.SV); d > timeout {
					__antithesis_instrumentation__.Notify(110294)
					timeout = d
				} else {
					__antithesis_instrumentation__.Notify(110295)
				}
				__antithesis_instrumentation__.Notify(110293)
				return timeout
			},
			successes:       store.metrics.MVCCGCQueueSuccesses,
			failures:        store.metrics.MVCCGCQueueFailures,
			pending:         store.metrics.MVCCGCQueuePending,
			processingNanos: store.metrics.MVCCGCQueueProcessingNanos,
		},
	)
	__antithesis_instrumentation__.Notify(110291)
	return mgcq
}

type mvccGCQueueScore struct {
	TTL                 time.Duration
	LastGC              time.Duration
	DeadFraction        float64
	ValuesScalableScore float64
	IntentScore         float64
	FuzzFactor          float64
	FinalScore          float64
	ShouldQueue         bool

	GCBytes                  int64
	GCByteAge                int64
	ExpMinGCByteAgeReduction int64
}

func (r mvccGCQueueScore) String() string {
	__antithesis_instrumentation__.Notify(110296)
	if (r == mvccGCQueueScore{}) {
		__antithesis_instrumentation__.Notify(110300)
		return "(empty)"
	} else {
		__antithesis_instrumentation__.Notify(110301)
	}
	__antithesis_instrumentation__.Notify(110297)
	if r.ExpMinGCByteAgeReduction < 0 {
		__antithesis_instrumentation__.Notify(110302)
		r.ExpMinGCByteAgeReduction = 0
	} else {
		__antithesis_instrumentation__.Notify(110303)
	}
	__antithesis_instrumentation__.Notify(110298)
	lastGC := "never"
	if r.LastGC != 0 {
		__antithesis_instrumentation__.Notify(110304)
		lastGC = fmt.Sprintf("%s ago", r.LastGC)
	} else {
		__antithesis_instrumentation__.Notify(110305)
	}
	__antithesis_instrumentation__.Notify(110299)
	return fmt.Sprintf("queue=%t with %.2f/fuzz(%.2f)=%.2f=valScaleScore(%.2f)*deadFrac(%.2f)+intentScore(%.2f)\n"+
		"likely last GC: %s, %s non-live, curr. age %s*s, min exp. reduction: %s*s",
		r.ShouldQueue, r.FinalScore, r.FuzzFactor, r.FinalScore/r.FuzzFactor, r.ValuesScalableScore,
		r.DeadFraction, r.IntentScore, lastGC, humanizeutil.IBytes(r.GCBytes),
		humanizeutil.IBytes(r.GCByteAge), humanizeutil.IBytes(r.ExpMinGCByteAgeReduction))
}

func (mgcq *mvccGCQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, repl *Replica, _ spanconfig.StoreReader,
) (bool, float64) {
	__antithesis_instrumentation__.Notify(110306)

	_, conf := repl.DescAndSpanConfig()
	canGC, _, gcTimestamp, oldThreshold, newThreshold, err := repl.checkProtectedTimestampsForGC(ctx, conf.TTL())
	if err != nil {
		__antithesis_instrumentation__.Notify(110310)
		log.VErrEventf(ctx, 2, "failed to check protected timestamp for gc: %v", err)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(110311)
	}
	__antithesis_instrumentation__.Notify(110307)
	if !canGC {
		__antithesis_instrumentation__.Notify(110312)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(110313)
	}
	__antithesis_instrumentation__.Notify(110308)
	canAdvanceGCThreshold := !newThreshold.Equal(oldThreshold)
	lastGC, err := repl.getQueueLastProcessed(ctx, mgcq.name)
	if err != nil {
		__antithesis_instrumentation__.Notify(110314)
		log.VErrEventf(ctx, 2, "failed to fetch last processed time: %v", err)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(110315)
	}
	__antithesis_instrumentation__.Notify(110309)
	r := makeMVCCGCQueueScore(ctx, repl, gcTimestamp, lastGC, conf.TTL(), canAdvanceGCThreshold)
	return r.ShouldQueue, r.FinalScore
}

func makeMVCCGCQueueScore(
	ctx context.Context,
	repl *Replica,
	now hlc.Timestamp,
	lastGC hlc.Timestamp,
	gcTTL time.Duration,
	canAdvanceGCThreshold bool,
) mvccGCQueueScore {
	__antithesis_instrumentation__.Notify(110316)
	repl.mu.Lock()
	ms := *repl.mu.state.Stats
	repl.mu.Unlock()

	if repl.store.cfg.TestingKnobs.DisableLastProcessedCheck {
		__antithesis_instrumentation__.Notify(110318)
		lastGC = hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(110319)
	}
	__antithesis_instrumentation__.Notify(110317)

	r := makeMVCCGCQueueScoreImpl(
		ctx, int64(repl.RangeID), now, ms, gcTTL, lastGC, canAdvanceGCThreshold,
	)
	return r
}

func makeMVCCGCQueueScoreImpl(
	ctx context.Context,
	fuzzSeed int64,
	now hlc.Timestamp,
	ms enginepb.MVCCStats,
	gcTTL time.Duration,
	lastGC hlc.Timestamp,
	canAdvanceGCThreshold bool,
) mvccGCQueueScore {
	__antithesis_instrumentation__.Notify(110320)
	ms.Forward(now.WallTime)
	var r mvccGCQueueScore

	if !lastGC.IsEmpty() {
		__antithesis_instrumentation__.Notify(110326)
		r.LastGC = time.Duration(now.WallTime - lastGC.WallTime)
	} else {
		__antithesis_instrumentation__.Notify(110327)
	}
	__antithesis_instrumentation__.Notify(110321)

	r.TTL = gcTTL

	if r.TTL <= time.Second {
		__antithesis_instrumentation__.Notify(110328)
		r.TTL = time.Second
	} else {
		__antithesis_instrumentation__.Notify(110329)
	}
	__antithesis_instrumentation__.Notify(110322)

	r.GCByteAge = ms.GCByteAge(now.WallTime)
	r.GCBytes = ms.GCBytes()

	r.ExpMinGCByteAgeReduction = r.GCByteAge - r.GCBytes*int64(r.TTL.Seconds())

	clamp := func(n int64) float64 {
		__antithesis_instrumentation__.Notify(110330)
		if n < 0 {
			__antithesis_instrumentation__.Notify(110332)
			return 0.0
		} else {
			__antithesis_instrumentation__.Notify(110333)
		}
		__antithesis_instrumentation__.Notify(110331)
		return float64(n)
	}
	__antithesis_instrumentation__.Notify(110323)
	r.DeadFraction = math.Max(1-clamp(ms.LiveBytes)/(1+clamp(ms.ValBytes)+clamp(ms.KeyBytes)), 0)

	denominator := r.TTL.Seconds() * (1.0 + clamp(r.GCBytes))
	r.ValuesScalableScore = clamp(r.GCByteAge) / denominator

	r.IntentScore = ms.AvgIntentAge(now.WallTime) / float64(intentAgeNormalization.Nanoseconds()/1e9)

	r.FuzzFactor = 0.95 + 0.05*rand.New(rand.NewSource(fuzzSeed)).Float64()

	valScore := r.DeadFraction * r.ValuesScalableScore
	r.FinalScore = r.FuzzFactor * (valScore + r.IntentScore)

	r.ShouldQueue = canAdvanceGCThreshold && func() bool {
		__antithesis_instrumentation__.Notify(110334)
		return r.FuzzFactor*valScore > mvccGCKeyScoreThreshold == true
	}() == true

	if !r.ShouldQueue && func() bool {
		__antithesis_instrumentation__.Notify(110335)
		return r.FuzzFactor*r.IntentScore > mvccGCIntentScoreThreshold == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(110336)
		return (r.LastGC == 0 || func() bool {
			__antithesis_instrumentation__.Notify(110337)
			return r.LastGC >= mvccGCQueueIntentCooldownDuration == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(110338)
		r.ShouldQueue = true
	} else {
		__antithesis_instrumentation__.Notify(110339)
	}
	__antithesis_instrumentation__.Notify(110324)

	if largeAbortSpan(ms) && func() bool {
		__antithesis_instrumentation__.Notify(110340)
		return !r.ShouldQueue == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(110341)
		return (r.LastGC == 0 || func() bool {
			__antithesis_instrumentation__.Notify(110342)
			return r.LastGC > kvserverbase.TxnCleanupThreshold == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(110343)
		r.ShouldQueue = true
		r.FinalScore++
	} else {
		__antithesis_instrumentation__.Notify(110344)
	}
	__antithesis_instrumentation__.Notify(110325)

	return r
}

type replicaGCer struct {
	repl                *Replica
	count               int32
	admissionController KVAdmissionController
	storeID             roachpb.StoreID
}

var _ gc.GCer = &replicaGCer{}

func (r *replicaGCer) template() roachpb.GCRequest {
	__antithesis_instrumentation__.Notify(110345)
	desc := r.repl.Desc()
	var template roachpb.GCRequest
	template.Key = desc.StartKey.AsRawKey()
	template.EndKey = desc.EndKey.AsRawKey()

	return template
}

func (r *replicaGCer) send(ctx context.Context, req roachpb.GCRequest) error {
	__antithesis_instrumentation__.Notify(110346)
	n := atomic.AddInt32(&r.count, 1)
	log.Eventf(ctx, "sending batch %d (%d keys)", n, len(req.Keys))

	var ba roachpb.BatchRequest

	ba.RangeID = r.repl.Desc().RangeID
	ba.Timestamp = r.repl.Clock().Now()
	ba.Add(&req)

	var admissionHandle interface{}
	if r.admissionController != nil {
		__antithesis_instrumentation__.Notify(110350)
		ba.AdmissionHeader = roachpb.AdmissionHeader{

			Priority:                 int32(admission.NormalPri),
			CreateTime:               timeutil.Now().UnixNano(),
			Source:                   roachpb.AdmissionHeader_ROOT_KV,
			NoMemoryReservedAtSource: true,
		}
		ba.Replica.StoreID = r.storeID
		var err error
		admissionHandle, err = r.admissionController.AdmitKVWork(ctx, roachpb.SystemTenantID, &ba)
		if err != nil {
			__antithesis_instrumentation__.Notify(110351)
			return err
		} else {
			__antithesis_instrumentation__.Notify(110352)
		}
	} else {
		__antithesis_instrumentation__.Notify(110353)
	}
	__antithesis_instrumentation__.Notify(110347)
	_, pErr := r.repl.Send(ctx, ba)
	if r.admissionController != nil {
		__antithesis_instrumentation__.Notify(110354)
		r.admissionController.AdmittedKVWorkDone(admissionHandle)
	} else {
		__antithesis_instrumentation__.Notify(110355)
	}
	__antithesis_instrumentation__.Notify(110348)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(110356)
		log.VErrEventf(ctx, 2, "%v", pErr.String())
		return pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(110357)
	}
	__antithesis_instrumentation__.Notify(110349)
	return nil
}

func (r *replicaGCer) SetGCThreshold(ctx context.Context, thresh gc.Threshold) error {
	__antithesis_instrumentation__.Notify(110358)
	req := r.template()
	req.Threshold = thresh.Key
	return r.send(ctx, req)
}

func (r *replicaGCer) GC(ctx context.Context, keys []roachpb.GCRequest_GCKey) error {
	__antithesis_instrumentation__.Notify(110359)
	if len(keys) == 0 {
		__antithesis_instrumentation__.Notify(110361)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(110362)
	}
	__antithesis_instrumentation__.Notify(110360)
	req := r.template()
	req.Keys = keys
	return r.send(ctx, req)
}

func (mgcq *mvccGCQueue) process(
	ctx context.Context, repl *Replica, _ spanconfig.StoreReader,
) (processed bool, err error) {
	__antithesis_instrumentation__.Notify(110363)

	desc, conf := repl.DescAndSpanConfig()

	canGC, cacheTimestamp, gcTimestamp, oldThreshold, newThreshold, err := repl.checkProtectedTimestampsForGC(ctx, conf.TTL())
	if err != nil {
		__antithesis_instrumentation__.Notify(110371)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(110372)
	}
	__antithesis_instrumentation__.Notify(110364)
	if !canGC {
		__antithesis_instrumentation__.Notify(110373)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(110374)
	}
	__antithesis_instrumentation__.Notify(110365)
	canAdvanceGCThreshold := !newThreshold.Equal(oldThreshold)

	lastGC, err := repl.getQueueLastProcessed(ctx, mgcq.name)
	if err != nil {
		__antithesis_instrumentation__.Notify(110375)
		lastGC = hlc.Timestamp{}
		log.VErrEventf(ctx, 2, "failed to fetch last processed time: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(110376)
	}
	__antithesis_instrumentation__.Notify(110366)
	r := makeMVCCGCQueueScore(ctx, repl, gcTimestamp, lastGC, conf.TTL(), canAdvanceGCThreshold)
	log.VEventf(ctx, 2, "processing replica %s with score %s", repl.String(), r)

	if err := repl.markPendingGC(cacheTimestamp, newThreshold); err != nil {
		__antithesis_instrumentation__.Notify(110377)
		log.VEventf(ctx, 1, "not gc'ing replica %v due to pending protection: %v", repl, err)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(110378)
	}
	__antithesis_instrumentation__.Notify(110367)

	if err := repl.setQueueLastProcessed(ctx, mgcq.name, repl.store.Clock().Now()); err != nil {
		__antithesis_instrumentation__.Notify(110379)
		log.VErrEventf(ctx, 2, "failed to update last processed time: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(110380)
	}
	__antithesis_instrumentation__.Notify(110368)

	snap := repl.store.Engine().NewSnapshot()
	defer snap.Close()

	intentAgeThreshold := gc.IntentAgeThreshold.Get(&repl.store.ClusterSettings().SV)
	maxIntentsPerCleanupBatch := gc.MaxIntentsPerCleanupBatch.Get(&repl.store.ClusterSettings().SV)
	maxIntentKeyBytesPerCleanupBatch := gc.MaxIntentKeyBytesPerCleanupBatch.Get(&repl.store.ClusterSettings().SV)

	info, err := gc.Run(ctx, desc, snap, gcTimestamp, newThreshold,
		gc.RunOptions{
			IntentAgeThreshold:                     intentAgeThreshold,
			MaxIntentsPerIntentCleanupBatch:        maxIntentsPerCleanupBatch,
			MaxIntentKeyBytesPerIntentCleanupBatch: maxIntentKeyBytesPerCleanupBatch,
			MaxTxnsPerIntentCleanupBatch:           intentresolver.MaxTxnsPerIntentCleanupBatch,
			IntentCleanupBatchTimeout:              mvccGCQueueIntentBatchTimeout,
		},
		conf.TTL(),
		&replicaGCer{
			repl:                repl,
			admissionController: mgcq.store.cfg.KVAdmissionController,
			storeID:             mgcq.store.StoreID(),
		},
		func(ctx context.Context, intents []roachpb.Intent) error {
			__antithesis_instrumentation__.Notify(110381)
			intentCount, err := repl.store.intentResolver.
				CleanupIntents(ctx, intents, gcTimestamp, roachpb.PUSH_TOUCH)
			if err == nil {
				__antithesis_instrumentation__.Notify(110383)
				mgcq.store.metrics.GCResolveSuccess.Inc(int64(intentCount))
			} else {
				__antithesis_instrumentation__.Notify(110384)
				mgcq.store.metrics.GCResolveFailed.Inc(int64(intentCount))
			}
			__antithesis_instrumentation__.Notify(110382)
			return err
		},
		func(ctx context.Context, txn *roachpb.Transaction) error {
			__antithesis_instrumentation__.Notify(110385)
			err := repl.store.intentResolver.
				CleanupTxnIntentsOnGCAsync(ctx, repl.RangeID, txn, gcTimestamp,
					func(pushed, succeeded bool) {
						__antithesis_instrumentation__.Notify(110388)
						if pushed {
							__antithesis_instrumentation__.Notify(110390)
							mgcq.store.metrics.GCPushTxn.Inc(1)
						} else {
							__antithesis_instrumentation__.Notify(110391)
						}
						__antithesis_instrumentation__.Notify(110389)
						if succeeded {
							__antithesis_instrumentation__.Notify(110392)
							mgcq.store.metrics.GCResolveSuccess.Inc(int64(len(txn.LockSpans)))
						} else {
							__antithesis_instrumentation__.Notify(110393)
							mgcq.store.metrics.GCTxnIntentsResolveFailed.Inc(int64(len(txn.LockSpans)))
						}
					})
			__antithesis_instrumentation__.Notify(110386)
			if errors.Is(err, stop.ErrThrottled) {
				__antithesis_instrumentation__.Notify(110394)
				log.Eventf(ctx, "processing txn %s: %s; skipping for future GC", txn.ID.Short(), err)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(110395)
			}
			__antithesis_instrumentation__.Notify(110387)
			return err
		})
	__antithesis_instrumentation__.Notify(110369)
	if err != nil {
		__antithesis_instrumentation__.Notify(110396)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(110397)
	}
	__antithesis_instrumentation__.Notify(110370)

	log.Eventf(ctx, "MVCC stats after GC: %+v", repl.GetMVCCStats())
	log.Eventf(ctx, "GC score after GC: %s", makeMVCCGCQueueScore(
		ctx, repl, repl.store.Clock().Now(), lastGC, conf.TTL(), canAdvanceGCThreshold))
	updateStoreMetricsWithGCInfo(mgcq.store.metrics, info)
	return true, nil
}

func updateStoreMetricsWithGCInfo(metrics *StoreMetrics, info gc.Info) {
	__antithesis_instrumentation__.Notify(110398)
	metrics.GCNumKeysAffected.Inc(int64(info.NumKeysAffected))
	metrics.GCIntentsConsidered.Inc(int64(info.IntentsConsidered))
	metrics.GCIntentTxns.Inc(int64(info.IntentTxns))
	metrics.GCTransactionSpanScanned.Inc(int64(info.TransactionSpanTotal))
	metrics.GCTransactionSpanGCAborted.Inc(int64(info.TransactionSpanGCAborted))
	metrics.GCTransactionSpanGCCommitted.Inc(int64(info.TransactionSpanGCCommitted))
	metrics.GCTransactionSpanGCStaging.Inc(int64(info.TransactionSpanGCStaging))
	metrics.GCTransactionSpanGCPending.Inc(int64(info.TransactionSpanGCPending))
	metrics.GCAbortSpanScanned.Inc(int64(info.AbortSpanTotal))
	metrics.GCAbortSpanConsidered.Inc(int64(info.AbortSpanConsidered))
	metrics.GCAbortSpanGCNum.Inc(int64(info.AbortSpanGCNum))
	metrics.GCPushTxn.Inc(int64(info.PushTxn))
	metrics.GCResolveTotal.Inc(int64(info.ResolveTotal))
}

func (*mvccGCQueue) timer(_ time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(110399)
	return mvccGCQueueTimerDuration
}

func (*mvccGCQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(110400)
	return nil
}
