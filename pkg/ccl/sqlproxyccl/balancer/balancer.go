package balancer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/list"
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/tenant"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/marusama/semaphore"
)

var ErrNoAvailablePods = errors.New("no available pods")

const (
	rebalanceInterval = 30 * time.Second

	minDrainPeriod = 1 * time.Minute

	defaultMaxConcurrentRebalances = 100

	maxTransferAttempts = 3
)

type balancerOptions struct {
	maxConcurrentRebalances int
	noRebalanceLoop         bool
	timeSource              timeutil.TimeSource
}

type Option func(opts *balancerOptions)

func MaxConcurrentRebalances(max int) Option {
	__antithesis_instrumentation__.Notify(20969)
	return func(opts *balancerOptions) {
		__antithesis_instrumentation__.Notify(20970)
		opts.maxConcurrentRebalances = max
	}
}

func NoRebalanceLoop() Option {
	__antithesis_instrumentation__.Notify(20971)
	return func(opts *balancerOptions) {
		__antithesis_instrumentation__.Notify(20972)
		opts.noRebalanceLoop = true
	}
}

func TimeSource(ts timeutil.TimeSource) Option {
	__antithesis_instrumentation__.Notify(20973)
	return func(opts *balancerOptions) {
		__antithesis_instrumentation__.Notify(20974)
		opts.timeSource = ts
	}
}

type Balancer struct {
	mu struct {
		syncutil.Mutex

		rng *rand.Rand
	}

	stopper *stop.Stopper

	directoryCache tenant.DirectoryCache

	connTracker *ConnTracker

	queue *rebalancerQueue

	processSem semaphore.Semaphore

	timeSource timeutil.TimeSource
}

func NewBalancer(
	ctx context.Context,
	stopper *stop.Stopper,
	directoryCache tenant.DirectoryCache,
	connTracker *ConnTracker,
	opts ...Option,
) (*Balancer, error) {
	__antithesis_instrumentation__.Notify(20975)

	options := &balancerOptions{}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(20982)
		opt(options)
	}
	__antithesis_instrumentation__.Notify(20976)
	if options.maxConcurrentRebalances == 0 {
		__antithesis_instrumentation__.Notify(20983)
		options.maxConcurrentRebalances = defaultMaxConcurrentRebalances
	} else {
		__antithesis_instrumentation__.Notify(20984)
	}
	__antithesis_instrumentation__.Notify(20977)
	if options.timeSource == nil {
		__antithesis_instrumentation__.Notify(20985)
		options.timeSource = timeutil.DefaultTimeSource{}
	} else {
		__antithesis_instrumentation__.Notify(20986)
	}
	__antithesis_instrumentation__.Notify(20978)

	ctx, _ = stopper.WithCancelOnQuiesce(ctx)

	q, err := newRebalancerQueue(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(20987)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20988)
	}
	__antithesis_instrumentation__.Notify(20979)

	b := &Balancer{
		stopper:        stopper,
		connTracker:    connTracker,
		directoryCache: directoryCache,
		queue:          q,
		processSem:     semaphore.New(options.maxConcurrentRebalances),
		timeSource:     options.timeSource,
	}
	b.mu.rng, _ = randutil.NewPseudoRand()

	if err := b.stopper.RunAsyncTask(ctx, "processQueue", b.processQueue); err != nil {
		__antithesis_instrumentation__.Notify(20989)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20990)
	}
	__antithesis_instrumentation__.Notify(20980)

	if !options.noRebalanceLoop {
		__antithesis_instrumentation__.Notify(20991)

		if err := b.stopper.RunAsyncTask(ctx, "rebalanceLoop", b.rebalanceLoop); err != nil {
			__antithesis_instrumentation__.Notify(20992)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(20993)
		}
	} else {
		__antithesis_instrumentation__.Notify(20994)
	}
	__antithesis_instrumentation__.Notify(20981)

	return b, nil
}

func (b *Balancer) SelectTenantPod(pods []*tenant.Pod) (*tenant.Pod, error) {
	__antithesis_instrumentation__.Notify(20995)
	pod := selectTenantPod(b.randFloat32(), pods)
	if pod == nil {
		__antithesis_instrumentation__.Notify(20997)
		return nil, ErrNoAvailablePods
	} else {
		__antithesis_instrumentation__.Notify(20998)
	}
	__antithesis_instrumentation__.Notify(20996)
	return pod, nil
}

func (b *Balancer) randFloat32() float32 {
	__antithesis_instrumentation__.Notify(20999)
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.mu.rng.Float32()
}

func (b *Balancer) processQueue(ctx context.Context) {
	__antithesis_instrumentation__.Notify(21000)

	processOneReq := func() (canContinue bool) {
		__antithesis_instrumentation__.Notify(21002)
		if err := b.processSem.Acquire(ctx, 1); err != nil {
			__antithesis_instrumentation__.Notify(21006)
			log.Errorf(ctx, "could not acquire processSem: %v", err.Error())
			return false
		} else {
			__antithesis_instrumentation__.Notify(21007)
		}
		__antithesis_instrumentation__.Notify(21003)

		req, err := b.queue.dequeue(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(21008)

			log.Errorf(ctx, "could not dequeue from rebalancer queue: %v", err.Error())
			return false
		} else {
			__antithesis_instrumentation__.Notify(21009)
		}
		__antithesis_instrumentation__.Notify(21004)

		if err := b.stopper.RunAsyncTask(ctx, "processQueue-item", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(21010)
			defer b.processSem.Release(1)

			for i := 0; i < maxTransferAttempts && func() bool {
				__antithesis_instrumentation__.Notify(21011)
				return ctx.Err() == nil == true
			}() == true; i++ {
				__antithesis_instrumentation__.Notify(21012)

				err := req.conn.TransferConnection()
				if err == nil || func() bool {
					__antithesis_instrumentation__.Notify(21014)
					return errors.Is(err, context.Canceled) == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(21015)
					return req.dst == req.conn.ServerRemoteAddr() == true
				}() == true {
					__antithesis_instrumentation__.Notify(21016)
					break
				} else {
					__antithesis_instrumentation__.Notify(21017)
				}
				__antithesis_instrumentation__.Notify(21013)

				time.Sleep(250 * time.Millisecond)
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(21018)

			log.Errorf(ctx, "could not run async task for processQueue-item: %v", err.Error())
		} else {
			__antithesis_instrumentation__.Notify(21019)
		}
		__antithesis_instrumentation__.Notify(21005)
		return true
	}
	__antithesis_instrumentation__.Notify(21001)
	for ctx.Err() == nil && func() bool {
		__antithesis_instrumentation__.Notify(21020)
		return processOneReq() == true
	}() == true {
		__antithesis_instrumentation__.Notify(21021)
	}
}

func (b *Balancer) rebalanceLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(21022)
	timer := b.timeSource.NewTimer()
	defer timer.Stop()
	for {
		__antithesis_instrumentation__.Notify(21023)
		timer.Reset(rebalanceInterval)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(21024)
			return
		case <-timer.Ch():
			__antithesis_instrumentation__.Notify(21025)
			timer.MarkRead()
			b.rebalance(ctx)
		}
	}
}

func (b *Balancer) rebalance(ctx context.Context) {
	__antithesis_instrumentation__.Notify(21026)

	tenantConns := b.connTracker.GetAllConns()

	for tenantID, allConns := range tenantConns {
		__antithesis_instrumentation__.Notify(21027)
		tenantPods, err := b.directoryCache.TryLookupTenantPods(ctx, tenantID)
		if err != nil {
			__antithesis_instrumentation__.Notify(21032)

			log.Errorf(ctx, "could not lookup pods for tenant %s: %v", tenantID, err.Error())
			continue
		} else {
			__antithesis_instrumentation__.Notify(21033)
		}
		__antithesis_instrumentation__.Notify(21028)

		podMap := make(map[string]*tenant.Pod)
		hasRunningPod := false
		for _, pod := range tenantPods {
			__antithesis_instrumentation__.Notify(21034)
			podMap[pod.Addr] = pod

			if pod.State == tenant.RUNNING {
				__antithesis_instrumentation__.Notify(21035)
				hasRunningPod = true
			} else {
				__antithesis_instrumentation__.Notify(21036)
			}
		}
		__antithesis_instrumentation__.Notify(21029)

		if !hasRunningPod {
			__antithesis_instrumentation__.Notify(21037)
			continue
		} else {
			__antithesis_instrumentation__.Notify(21038)
		}
		__antithesis_instrumentation__.Notify(21030)

		connMap := make(map[string][]ConnectionHandle)
		for _, conn := range allConns {
			__antithesis_instrumentation__.Notify(21039)

			if conn.Context().Err() != nil {
				__antithesis_instrumentation__.Notify(21041)
				continue
			} else {
				__antithesis_instrumentation__.Notify(21042)
			}
			__antithesis_instrumentation__.Notify(21040)
			addr := conn.ServerRemoteAddr()
			connMap[addr] = append(connMap[addr], conn)
		}
		__antithesis_instrumentation__.Notify(21031)

		for addr, podConns := range connMap {
			__antithesis_instrumentation__.Notify(21043)
			pod, ok := podMap[addr]
			if !ok {
				__antithesis_instrumentation__.Notify(21047)

				continue
			} else {
				__antithesis_instrumentation__.Notify(21048)
			}
			__antithesis_instrumentation__.Notify(21044)

			if pod.State != tenant.DRAINING {
				__antithesis_instrumentation__.Notify(21049)
				continue
			} else {
				__antithesis_instrumentation__.Notify(21050)
			}
			__antithesis_instrumentation__.Notify(21045)

			drainingFor := b.timeSource.Now().Sub(pod.StateTimestamp)
			if drainingFor < minDrainPeriod {
				__antithesis_instrumentation__.Notify(21051)
				continue
			} else {
				__antithesis_instrumentation__.Notify(21052)
			}
			__antithesis_instrumentation__.Notify(21046)

			for _, c := range podConns {
				__antithesis_instrumentation__.Notify(21053)

				b.queue.enqueue(&rebalanceRequest{
					createdAt: b.timeSource.Now(),
					conn:      c,
				})
			}
		}
	}
}

type rebalanceRequest struct {
	createdAt time.Time
	conn      ConnectionHandle
	dst       string
}

type rebalancerQueue struct {
	mu       syncutil.Mutex
	sem      semaphore.Semaphore
	queue    *list.List
	elements map[ConnectionHandle]*list.Element
}

func newRebalancerQueue(ctx context.Context) (*rebalancerQueue, error) {
	__antithesis_instrumentation__.Notify(21054)
	q := &rebalancerQueue{
		sem:      semaphore.New(math.MaxInt32),
		queue:    list.New(),
		elements: make(map[ConnectionHandle]*list.Element),
	}

	if err := q.sem.Acquire(ctx, math.MaxInt32); err != nil {
		__antithesis_instrumentation__.Notify(21056)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21057)
	}
	__antithesis_instrumentation__.Notify(21055)
	return q, nil
}

func (q *rebalancerQueue) enqueue(req *rebalanceRequest) {
	__antithesis_instrumentation__.Notify(21058)
	q.mu.Lock()
	defer q.mu.Unlock()

	e, ok := q.elements[req.conn]
	if ok {
		__antithesis_instrumentation__.Notify(21059)

		if e.Value.(*rebalanceRequest).createdAt.Before(req.createdAt) {
			__antithesis_instrumentation__.Notify(21060)
			e.Value = req
		} else {
			__antithesis_instrumentation__.Notify(21061)
		}
	} else {
		__antithesis_instrumentation__.Notify(21062)
		e = q.queue.PushBack(req)
		q.elements[req.conn] = e
		q.sem.Release(1)
	}
}

func (q *rebalancerQueue) dequeue(ctx context.Context) (*rebalanceRequest, error) {
	__antithesis_instrumentation__.Notify(21063)

	if err := q.sem.Acquire(ctx, 1); err != nil {
		__antithesis_instrumentation__.Notify(21066)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21067)
	}
	__antithesis_instrumentation__.Notify(21064)

	q.mu.Lock()
	defer q.mu.Unlock()

	e := q.queue.Front()
	if e == nil {
		__antithesis_instrumentation__.Notify(21068)

		return nil, errors.AssertionFailedf("unexpected empty queue")
	} else {
		__antithesis_instrumentation__.Notify(21069)
	}
	__antithesis_instrumentation__.Notify(21065)

	req := q.queue.Remove(e).(*rebalanceRequest)
	delete(q.elements, req.conn)
	return req, nil
}
