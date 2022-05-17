package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type replicaQueue interface {
	Start(*stop.Stopper)

	MaybeAddAsync(context.Context, replicaInQueue, hlc.ClockTimestamp)

	MaybeRemove(roachpb.RangeID)

	Name() string

	NeedsLease() bool

	SetDisabled(disabled bool)
}

type replicaSet interface {
	Visit(func(*Replica) bool)

	EstimatedCount() int
}

type replicaScanner struct {
	log.AmbientContext
	clock   *hlc.Clock
	stopper *stop.Stopper

	targetInterval time.Duration
	minIdleTime    time.Duration
	maxIdleTime    time.Duration
	waitTimer      timeutil.Timer
	replicas       replicaSet
	queues         []replicaQueue
	removed        chan *Replica

	mu struct {
		syncutil.Mutex
		scanCount        int64
		waitEnabledCount int64
		total            time.Duration

		disabled bool
	}

	setDisabledCh chan struct{}
}

func newReplicaScanner(
	ambient log.AmbientContext,
	clock *hlc.Clock,
	targetInterval, minIdleTime, maxIdleTime time.Duration,
	replicas replicaSet,
) *replicaScanner {
	__antithesis_instrumentation__.Notify(122144)
	if targetInterval < 0 {
		__antithesis_instrumentation__.Notify(122147)
		panic("scanner interval must be greater than or equal to zero")
	} else {
		__antithesis_instrumentation__.Notify(122148)
	}
	__antithesis_instrumentation__.Notify(122145)
	rs := &replicaScanner{
		AmbientContext: ambient,
		clock:          clock,
		targetInterval: targetInterval,
		minIdleTime:    minIdleTime,
		maxIdleTime:    maxIdleTime,
		replicas:       replicas,
		removed:        make(chan *Replica),
		setDisabledCh:  make(chan struct{}, 1),
	}
	if targetInterval == 0 {
		__antithesis_instrumentation__.Notify(122149)
		rs.SetDisabled(true)
	} else {
		__antithesis_instrumentation__.Notify(122150)
	}
	__antithesis_instrumentation__.Notify(122146)
	return rs
}

func (rs *replicaScanner) AddQueues(queues ...replicaQueue) {
	__antithesis_instrumentation__.Notify(122151)
	rs.queues = append(rs.queues, queues...)
}

func (rs *replicaScanner) Start() {
	__antithesis_instrumentation__.Notify(122152)
	for _, queue := range rs.queues {
		__antithesis_instrumentation__.Notify(122154)
		queue.Start(rs.stopper)
	}
	__antithesis_instrumentation__.Notify(122153)
	rs.scanLoop()
}

func (rs *replicaScanner) scanCount() int64 {
	__antithesis_instrumentation__.Notify(122155)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.mu.scanCount
}

func (rs *replicaScanner) waitEnabledCount() int64 {
	__antithesis_instrumentation__.Notify(122156)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.mu.waitEnabledCount
}

func (rs *replicaScanner) SetDisabled(disabled bool) {
	__antithesis_instrumentation__.Notify(122157)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.mu.disabled = disabled

	select {
	case rs.setDisabledCh <- struct{}{}:
		__antithesis_instrumentation__.Notify(122158)
	default:
		__antithesis_instrumentation__.Notify(122159)
	}
}

func (rs *replicaScanner) GetDisabled() bool {
	__antithesis_instrumentation__.Notify(122160)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	return rs.mu.disabled
}

func (rs *replicaScanner) avgScan() time.Duration {
	__antithesis_instrumentation__.Notify(122161)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.mu.scanCount == 0 {
		__antithesis_instrumentation__.Notify(122163)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(122164)
	}
	__antithesis_instrumentation__.Notify(122162)
	return time.Duration(rs.mu.total.Nanoseconds() / rs.mu.scanCount)
}

func (rs *replicaScanner) RemoveReplica(repl *Replica) {
	__antithesis_instrumentation__.Notify(122165)
	select {
	case rs.removed <- repl:
		__antithesis_instrumentation__.Notify(122166)
	case <-rs.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(122167)
	}
}

func (rs *replicaScanner) paceInterval(start, now time.Time) time.Duration {
	__antithesis_instrumentation__.Notify(122168)
	elapsed := now.Sub(start)
	remainingNanos := rs.targetInterval.Nanoseconds() - elapsed.Nanoseconds()
	if remainingNanos < 0 {
		__antithesis_instrumentation__.Notify(122173)
		remainingNanos = 0
	} else {
		__antithesis_instrumentation__.Notify(122174)
	}
	__antithesis_instrumentation__.Notify(122169)
	count := rs.replicas.EstimatedCount()
	if count < 1 {
		__antithesis_instrumentation__.Notify(122175)
		count = 1
	} else {
		__antithesis_instrumentation__.Notify(122176)
	}
	__antithesis_instrumentation__.Notify(122170)
	interval := time.Duration(remainingNanos / int64(count))
	if rs.minIdleTime > 0 && func() bool {
		__antithesis_instrumentation__.Notify(122177)
		return interval < rs.minIdleTime == true
	}() == true {
		__antithesis_instrumentation__.Notify(122178)
		interval = rs.minIdleTime
	} else {
		__antithesis_instrumentation__.Notify(122179)
	}
	__antithesis_instrumentation__.Notify(122171)
	if rs.maxIdleTime > 0 && func() bool {
		__antithesis_instrumentation__.Notify(122180)
		return interval > rs.maxIdleTime == true
	}() == true {
		__antithesis_instrumentation__.Notify(122181)
		interval = rs.maxIdleTime
	} else {
		__antithesis_instrumentation__.Notify(122182)
	}
	__antithesis_instrumentation__.Notify(122172)
	return interval
}

func (rs *replicaScanner) waitAndProcess(ctx context.Context, start time.Time, repl *Replica) bool {
	__antithesis_instrumentation__.Notify(122183)
	waitInterval := rs.paceInterval(start, timeutil.Now())
	rs.waitTimer.Reset(waitInterval)
	if log.V(6) {
		__antithesis_instrumentation__.Notify(122185)
		log.Infof(ctx, "wait timer interval set to %s", waitInterval)
	} else {
		__antithesis_instrumentation__.Notify(122186)
	}
	__antithesis_instrumentation__.Notify(122184)
	for {
		__antithesis_instrumentation__.Notify(122187)
		select {
		case <-rs.waitTimer.C:
			__antithesis_instrumentation__.Notify(122188)
			if log.V(6) {
				__antithesis_instrumentation__.Notify(122195)
				log.Infof(ctx, "wait timer fired")
			} else {
				__antithesis_instrumentation__.Notify(122196)
			}
			__antithesis_instrumentation__.Notify(122189)
			rs.waitTimer.Read = true
			if repl == nil {
				__antithesis_instrumentation__.Notify(122197)
				return false
			} else {
				__antithesis_instrumentation__.Notify(122198)
			}
			__antithesis_instrumentation__.Notify(122190)

			if log.V(2) {
				__antithesis_instrumentation__.Notify(122199)
				log.Infof(ctx, "replica scanner processing %s", repl)
			} else {
				__antithesis_instrumentation__.Notify(122200)
			}
			__antithesis_instrumentation__.Notify(122191)
			for _, q := range rs.queues {
				__antithesis_instrumentation__.Notify(122201)
				q.MaybeAddAsync(ctx, repl, rs.clock.NowAsClockTimestamp())
			}
			__antithesis_instrumentation__.Notify(122192)
			return false

		case repl := <-rs.removed:
			__antithesis_instrumentation__.Notify(122193)
			rs.removeReplica(repl)

		case <-rs.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(122194)
			return true
		}
	}
}

func (rs *replicaScanner) removeReplica(repl *Replica) {
	__antithesis_instrumentation__.Notify(122202)

	rangeID := repl.RangeID
	for _, q := range rs.queues {
		__antithesis_instrumentation__.Notify(122204)
		q.MaybeRemove(rangeID)
	}
	__antithesis_instrumentation__.Notify(122203)
	if log.V(6) {
		__antithesis_instrumentation__.Notify(122205)
		ctx := rs.AnnotateCtx(context.TODO())
		log.Infof(ctx, "removed replica %s", repl)
	} else {
		__antithesis_instrumentation__.Notify(122206)
	}
}

func (rs *replicaScanner) scanLoop() {
	__antithesis_instrumentation__.Notify(122207)
	ctx := rs.AnnotateCtx(context.Background())
	_ = rs.stopper.RunAsyncTask(ctx, "scan-loop", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(122208)
		start := timeutil.Now()

		defer rs.waitTimer.Stop()

		for {
			__antithesis_instrumentation__.Notify(122209)
			if rs.GetDisabled() {
				__antithesis_instrumentation__.Notify(122216)
				if done := rs.waitEnabled(); done {
					__antithesis_instrumentation__.Notify(122218)
					return
				} else {
					__antithesis_instrumentation__.Notify(122219)
				}
				__antithesis_instrumentation__.Notify(122217)
				continue
			} else {
				__antithesis_instrumentation__.Notify(122220)
			}
			__antithesis_instrumentation__.Notify(122210)
			var shouldStop bool
			count := 0
			rs.replicas.Visit(func(repl *Replica) bool {
				__antithesis_instrumentation__.Notify(122221)
				count++
				shouldStop = rs.waitAndProcess(ctx, start, repl)
				return !shouldStop
			})
			__antithesis_instrumentation__.Notify(122211)
			if count == 0 {
				__antithesis_instrumentation__.Notify(122222)

				shouldStop = rs.waitAndProcess(ctx, start, nil)
			} else {
				__antithesis_instrumentation__.Notify(122223)
			}
			__antithesis_instrumentation__.Notify(122212)

			if shouldStop {
				__antithesis_instrumentation__.Notify(122224)
				return
			} else {
				__antithesis_instrumentation__.Notify(122225)
			}
			__antithesis_instrumentation__.Notify(122213)

			func() {
				__antithesis_instrumentation__.Notify(122226)
				rs.mu.Lock()
				defer rs.mu.Unlock()
				rs.mu.scanCount++
				rs.mu.total += timeutil.Since(start)
			}()
			__antithesis_instrumentation__.Notify(122214)
			if log.V(6) {
				__antithesis_instrumentation__.Notify(122227)
				log.Infof(ctx, "reset replica scan iteration")
			} else {
				__antithesis_instrumentation__.Notify(122228)
			}
			__antithesis_instrumentation__.Notify(122215)

			start = timeutil.Now()
		}
	})
}

func (rs *replicaScanner) waitEnabled() bool {
	__antithesis_instrumentation__.Notify(122229)
	rs.mu.Lock()
	rs.mu.waitEnabledCount++
	rs.mu.Unlock()
	for {
		__antithesis_instrumentation__.Notify(122230)
		if !rs.GetDisabled() {
			__antithesis_instrumentation__.Notify(122232)
			return false
		} else {
			__antithesis_instrumentation__.Notify(122233)
		}
		__antithesis_instrumentation__.Notify(122231)
		select {
		case <-rs.setDisabledCh:
			__antithesis_instrumentation__.Notify(122234)
			continue

		case repl := <-rs.removed:
			__antithesis_instrumentation__.Notify(122235)
			rs.removeReplica(repl)

		case <-rs.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(122236)
			return true
		}
	}
}
