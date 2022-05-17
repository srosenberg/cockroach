package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

var consistencyCheckInterval = settings.RegisterDurationSetting(
	settings.SystemOnly,
	"server.consistency_check.interval",
	"the time between range consistency checks; set to 0 to disable consistency checking."+
		" Note that intervals that are too short can negatively impact performance.",
	24*time.Hour,
	settings.NonNegativeDuration,
)

var consistencyCheckRate = settings.RegisterByteSizeSetting(
	settings.SystemOnly,
	"server.consistency_check.max_rate",
	"the rate limit (bytes/sec) to use for consistency checks; used in "+
		"conjunction with server.consistency_check.interval to control the "+
		"frequency of consistency checks. Note that setting this too high can "+
		"negatively impact performance.",
	8<<20,
	settings.PositiveInt,
).WithPublic()

const consistencyCheckRateBurstFactor = 8

const consistencyCheckRateMinWait = 100 * time.Millisecond

const consistencyCheckAsyncConcurrency = 7

const consistencyCheckAsyncTimeout = time.Hour

var testingAggressiveConsistencyChecks = envutil.EnvOrDefaultBool("COCKROACH_CONSISTENCY_AGGRESSIVE", false)

type consistencyQueue struct {
	*baseQueue
	interval       func() time.Duration
	replicaCountFn func() int
}

type consistencyShouldQueueData struct {
	desc                      *roachpb.RangeDescriptor
	getQueueLastProcessed     func(ctx context.Context) (hlc.Timestamp, error)
	isNodeAvailable           func(nodeID roachpb.NodeID) bool
	disableLastProcessedCheck bool
	interval                  time.Duration
}

func newConsistencyQueue(store *Store) *consistencyQueue {
	__antithesis_instrumentation__.Notify(101000)
	q := &consistencyQueue{
		interval: func() time.Duration {
			__antithesis_instrumentation__.Notify(101002)
			return consistencyCheckInterval.Get(&store.ClusterSettings().SV)
		},
		replicaCountFn: store.ReplicaCount,
	}
	__antithesis_instrumentation__.Notify(101001)
	q.baseQueue = newBaseQueue(
		"consistencyChecker", q, store,
		queueConfig{
			maxSize:              defaultQueueMaxSize,
			needsLease:           true,
			needsSystemConfig:    false,
			acceptsUnsplitRanges: true,
			successes:            store.metrics.ConsistencyQueueSuccesses,
			failures:             store.metrics.ConsistencyQueueFailures,
			pending:              store.metrics.ConsistencyQueuePending,
			processingNanos:      store.metrics.ConsistencyQueueProcessingNanos,
			processTimeoutFunc:   makeRateLimitedTimeoutFunc(consistencyCheckRate),
		},
	)
	return q
}

func (q *consistencyQueue) shouldQueue(
	ctx context.Context, now hlc.ClockTimestamp, repl *Replica, _ spanconfig.StoreReader,
) (bool, float64) {
	__antithesis_instrumentation__.Notify(101003)
	return consistencyQueueShouldQueueImpl(ctx, now,
		consistencyShouldQueueData{
			desc: repl.Desc(),
			getQueueLastProcessed: func(ctx context.Context) (hlc.Timestamp, error) {
				__antithesis_instrumentation__.Notify(101004)
				return repl.getQueueLastProcessed(ctx, q.name)
			},
			isNodeAvailable: func(nodeID roachpb.NodeID) bool {
				__antithesis_instrumentation__.Notify(101005)
				if repl.store.cfg.NodeLiveness != nil {
					__antithesis_instrumentation__.Notify(101007)
					return repl.store.cfg.NodeLiveness.IsAvailableNotDraining(nodeID)
				} else {
					__antithesis_instrumentation__.Notify(101008)
				}
				__antithesis_instrumentation__.Notify(101006)

				return true
			},
			disableLastProcessedCheck: repl.store.cfg.TestingKnobs.DisableLastProcessedCheck,
			interval:                  q.interval(),
		})
}

func consistencyQueueShouldQueueImpl(
	ctx context.Context, now hlc.ClockTimestamp, data consistencyShouldQueueData,
) (bool, float64) {
	__antithesis_instrumentation__.Notify(101009)
	if data.interval <= 0 {
		__antithesis_instrumentation__.Notify(101013)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(101014)
	}
	__antithesis_instrumentation__.Notify(101010)

	shouldQ, priority := true, float64(0)
	if !data.disableLastProcessedCheck {
		__antithesis_instrumentation__.Notify(101015)
		lpTS, err := data.getQueueLastProcessed(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(101017)
			return false, 0
		} else {
			__antithesis_instrumentation__.Notify(101018)
		}
		__antithesis_instrumentation__.Notify(101016)
		if shouldQ, priority = shouldQueueAgain(now.ToTimestamp(), lpTS, data.interval); !shouldQ {
			__antithesis_instrumentation__.Notify(101019)
			return false, 0
		} else {
			__antithesis_instrumentation__.Notify(101020)
		}
	} else {
		__antithesis_instrumentation__.Notify(101021)
	}
	__antithesis_instrumentation__.Notify(101011)

	for _, rep := range data.desc.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(101022)
		if !data.isNodeAvailable(rep.NodeID) {
			__antithesis_instrumentation__.Notify(101023)
			return false, 0
		} else {
			__antithesis_instrumentation__.Notify(101024)
		}
	}
	__antithesis_instrumentation__.Notify(101012)
	return true, priority
}

func (q *consistencyQueue) process(
	ctx context.Context, repl *Replica, _ spanconfig.StoreReader,
) (bool, error) {
	__antithesis_instrumentation__.Notify(101025)
	if q.interval() <= 0 {
		__antithesis_instrumentation__.Notify(101030)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(101031)
	}
	__antithesis_instrumentation__.Notify(101026)

	if err := repl.setQueueLastProcessed(ctx, q.name, repl.store.Clock().Now()); err != nil {
		__antithesis_instrumentation__.Notify(101032)
		log.VErrEventf(ctx, 2, "failed to update last processed time: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(101033)
	}
	__antithesis_instrumentation__.Notify(101027)

	req := roachpb.CheckConsistencyRequest{

		Mode: roachpb.ChecksumMode_CHECK_VIA_QUEUE,
	}
	resp, pErr := repl.CheckConsistency(ctx, req)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(101034)
		var shouldQuiesce bool
		select {
		case <-repl.store.Stopper().ShouldQuiesce():
			__antithesis_instrumentation__.Notify(101037)
			shouldQuiesce = true
		default:
			__antithesis_instrumentation__.Notify(101038)
		}
		__antithesis_instrumentation__.Notify(101035)

		if shouldQuiesce && func() bool {
			__antithesis_instrumentation__.Notify(101039)
			return grpcutil.IsClosedConnection(pErr.GoError()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(101040)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(101041)
		}
		__antithesis_instrumentation__.Notify(101036)
		err := pErr.GoError()
		log.Errorf(ctx, "%v", err)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(101042)
	}
	__antithesis_instrumentation__.Notify(101028)
	if fn := repl.store.cfg.TestingKnobs.ConsistencyTestingKnobs.ConsistencyQueueResultHook; fn != nil {
		__antithesis_instrumentation__.Notify(101043)
		fn(resp)
	} else {
		__antithesis_instrumentation__.Notify(101044)
	}
	__antithesis_instrumentation__.Notify(101029)
	return true, nil
}

func (q *consistencyQueue) timer(duration time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(101045)

	replicaCount := q.replicaCountFn()
	if replicaCount == 0 {
		__antithesis_instrumentation__.Notify(101048)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(101049)
	}
	__antithesis_instrumentation__.Notify(101046)
	replInterval := q.interval() / time.Duration(replicaCount)
	if replInterval < duration {
		__antithesis_instrumentation__.Notify(101050)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(101051)
	}
	__antithesis_instrumentation__.Notify(101047)
	return replInterval - duration
}

func (*consistencyQueue) purgatoryChan() <-chan time.Time {
	__antithesis_instrumentation__.Notify(101052)
	return nil
}
