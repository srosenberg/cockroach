package flowinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/list"
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

const flowDoneChanSize = 8

var settingMaxRunningFlows = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.distsql.max_running_flows",
	"the value - when positive - used as is, or the value - when negative - "+
		"multiplied by the number of CPUs on a node, to determine the "+
		"maximum number of concurrent remote flows that can be run on the node",
	-128,
).WithPublic()

func getMaxRunningFlows(settings *cluster.Settings) int64 {
	__antithesis_instrumentation__.Notify(491765)
	maxRunningFlows := settingMaxRunningFlows.Get(&settings.SV)
	if maxRunningFlows < 0 {
		__antithesis_instrumentation__.Notify(491767)

		return -maxRunningFlows * int64(runtime.GOMAXPROCS(0))
	} else {
		__antithesis_instrumentation__.Notify(491768)
	}
	__antithesis_instrumentation__.Notify(491766)
	return maxRunningFlows
}

type FlowScheduler struct {
	log.AmbientContext
	stopper    *stop.Stopper
	flowDoneCh chan Flow
	metrics    *execinfra.DistSQLMetrics

	mu struct {
		syncutil.Mutex

		queue *list.List

		runningFlows map[execinfrapb.FlowID]execinfrapb.DistSQLRemoteFlowInfo
	}

	atomics struct {
		numRunning      int32
		maxRunningFlows int32
	}

	TestingKnobs struct {
		CancelDeadFlowsCallback func(numCanceled int)
	}
}

type flowWithCtx struct {
	ctx         context.Context
	flow        Flow
	enqueueTime time.Time
}

func (f *flowWithCtx) cleanupBeforeRun() {
	__antithesis_instrumentation__.Notify(491769)

	f.flow.Cleanup(f.ctx)
}

func NewFlowScheduler(
	ambient log.AmbientContext, stopper *stop.Stopper, settings *cluster.Settings,
) *FlowScheduler {
	__antithesis_instrumentation__.Notify(491770)
	fs := &FlowScheduler{
		AmbientContext: ambient,
		stopper:        stopper,
		flowDoneCh:     make(chan Flow, flowDoneChanSize),
	}
	fs.mu.queue = list.New()
	maxRunningFlows := getMaxRunningFlows(settings)
	fs.mu.runningFlows = make(map[execinfrapb.FlowID]execinfrapb.DistSQLRemoteFlowInfo, maxRunningFlows)
	fs.atomics.maxRunningFlows = int32(maxRunningFlows)
	settingMaxRunningFlows.SetOnChange(&settings.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(491772)
		atomic.StoreInt32(&fs.atomics.maxRunningFlows, int32(getMaxRunningFlows(settings)))
	})
	__antithesis_instrumentation__.Notify(491771)
	return fs
}

func (fs *FlowScheduler) Init(metrics *execinfra.DistSQLMetrics) {
	__antithesis_instrumentation__.Notify(491773)
	fs.metrics = metrics
}

func (fs *FlowScheduler) canRunFlow(_ Flow) bool {
	__antithesis_instrumentation__.Notify(491774)

	newNumRunning := atomic.AddInt32(&fs.atomics.numRunning, 1)
	if newNumRunning <= atomic.LoadInt32(&fs.atomics.maxRunningFlows) {
		__antithesis_instrumentation__.Notify(491776)

		return true
	} else {
		__antithesis_instrumentation__.Notify(491777)
	}
	__antithesis_instrumentation__.Notify(491775)
	atomic.AddInt32(&fs.atomics.numRunning, -1)
	return false
}

func (fs *FlowScheduler) runFlowNow(ctx context.Context, f Flow, locked bool) error {
	__antithesis_instrumentation__.Notify(491778)
	log.VEventf(
		ctx, 1, "flow scheduler running flow %s, currently running %d", f.GetID(), atomic.LoadInt32(&fs.atomics.numRunning)-1,
	)
	fs.metrics.FlowStart()
	if !locked {
		__antithesis_instrumentation__.Notify(491783)
		fs.mu.Lock()
	} else {
		__antithesis_instrumentation__.Notify(491784)
	}
	__antithesis_instrumentation__.Notify(491779)
	fs.mu.runningFlows[f.GetID()] = execinfrapb.DistSQLRemoteFlowInfo{
		FlowID:       f.GetID(),
		Timestamp:    timeutil.Now(),
		StatementSQL: f.StatementSQL(),
	}
	if !locked {
		__antithesis_instrumentation__.Notify(491785)
		fs.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(491786)
	}
	__antithesis_instrumentation__.Notify(491780)
	if err := f.Start(ctx, func() { __antithesis_instrumentation__.Notify(491787); fs.flowDoneCh <- f }); err != nil {
		__antithesis_instrumentation__.Notify(491788)
		f.Cleanup(ctx)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491789)
	}
	__antithesis_instrumentation__.Notify(491781)

	go func() {
		__antithesis_instrumentation__.Notify(491790)
		f.Wait()
		fs.mu.Lock()
		delete(fs.mu.runningFlows, f.GetID())
		fs.mu.Unlock()
		f.Cleanup(ctx)
	}()
	__antithesis_instrumentation__.Notify(491782)
	return nil
}

func (fs *FlowScheduler) ScheduleFlow(ctx context.Context, f Flow) error {
	__antithesis_instrumentation__.Notify(491791)
	err := fs.stopper.RunTaskWithErr(
		ctx, "flowinfra.FlowScheduler: scheduling flow", func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(491794)
			fs.metrics.FlowsScheduled.Inc(1)
			telemetry.Inc(sqltelemetry.DistSQLFlowsScheduled)
			if fs.canRunFlow(f) {
				__antithesis_instrumentation__.Notify(491796)
				return fs.runFlowNow(ctx, f, false)
			} else {
				__antithesis_instrumentation__.Notify(491797)
			}
			__antithesis_instrumentation__.Notify(491795)
			fs.mu.Lock()
			defer fs.mu.Unlock()
			log.VEventf(ctx, 1, "flow scheduler enqueuing flow %s to be run later", f.GetID())
			fs.metrics.FlowsQueued.Inc(1)
			telemetry.Inc(sqltelemetry.DistSQLFlowsQueued)
			fs.mu.queue.PushBack(&flowWithCtx{
				ctx:         ctx,
				flow:        f,
				enqueueTime: timeutil.Now(),
			})
			return nil

		})
	__antithesis_instrumentation__.Notify(491792)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(491798)
		return errors.Is(err, stop.ErrUnavailable) == true
	}() == true {
		__antithesis_instrumentation__.Notify(491799)

		f.Cleanup(ctx)
	} else {
		__antithesis_instrumentation__.Notify(491800)
	}
	__antithesis_instrumentation__.Notify(491793)
	return err
}

func (fs *FlowScheduler) NumFlowsInQueue() int {
	__antithesis_instrumentation__.Notify(491801)
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.mu.queue.Len()
}

func (fs *FlowScheduler) NumRunningFlows() int {
	__antithesis_instrumentation__.Notify(491802)
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return len(fs.mu.runningFlows)
}

func (fs *FlowScheduler) CancelDeadFlows(req *execinfrapb.CancelDeadFlowsRequest) {
	__antithesis_instrumentation__.Notify(491803)

	fs.mu.Lock()
	isEmpty := fs.mu.queue.Len() == 0
	fs.mu.Unlock()
	if isEmpty {
		__antithesis_instrumentation__.Notify(491807)
		return
	} else {
		__antithesis_instrumentation__.Notify(491808)
	}
	__antithesis_instrumentation__.Notify(491804)

	ctx := fs.AnnotateCtx(context.Background())
	log.VEventf(ctx, 1, "flow scheduler will attempt to cancel %d dead flows", len(req.FlowIDs))

	toCancel := make(map[uuid.UUID]struct{}, len(req.FlowIDs))
	for _, f := range req.FlowIDs {
		__antithesis_instrumentation__.Notify(491809)
		toCancel[f.UUID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(491805)
	numCanceled := 0
	defer func() {
		__antithesis_instrumentation__.Notify(491810)
		log.VEventf(ctx, 1, "flow scheduler canceled %d dead flows", numCanceled)
		if fs.TestingKnobs.CancelDeadFlowsCallback != nil {
			__antithesis_instrumentation__.Notify(491811)
			fs.TestingKnobs.CancelDeadFlowsCallback(numCanceled)
		} else {
			__antithesis_instrumentation__.Notify(491812)
		}
	}()
	__antithesis_instrumentation__.Notify(491806)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	var next *list.Element
	for e := fs.mu.queue.Front(); e != nil; e = next {
		__antithesis_instrumentation__.Notify(491813)

		next = e.Next()
		f := e.Value.(*flowWithCtx)
		if _, shouldCancel := toCancel[f.flow.GetID().UUID]; shouldCancel {
			__antithesis_instrumentation__.Notify(491814)
			fs.mu.queue.Remove(e)
			fs.metrics.FlowsQueued.Dec(1)
			numCanceled++
			f.cleanupBeforeRun()
		} else {
			__antithesis_instrumentation__.Notify(491815)
		}
	}
}

func (fs *FlowScheduler) Start() {
	__antithesis_instrumentation__.Notify(491816)
	ctx := fs.AnnotateCtx(context.Background())
	_ = fs.stopper.RunAsyncTask(ctx, "flow-scheduler", func(context.Context) {
		__antithesis_instrumentation__.Notify(491817)
		stopped := false
		fs.mu.Lock()
		defer fs.mu.Unlock()

		quiesceCh := fs.stopper.ShouldQuiesce()

		for {
			__antithesis_instrumentation__.Notify(491818)
			if stopped {
				__antithesis_instrumentation__.Notify(491820)

				if l := fs.mu.queue.Len(); l > 0 {
					__antithesis_instrumentation__.Notify(491823)
					log.Infof(ctx, "abandoning %d flows that will never run", l)
				} else {
					__antithesis_instrumentation__.Notify(491824)
				}
				__antithesis_instrumentation__.Notify(491821)
				for {
					__antithesis_instrumentation__.Notify(491825)
					e := fs.mu.queue.Front()
					if e == nil {
						__antithesis_instrumentation__.Notify(491827)
						break
					} else {
						__antithesis_instrumentation__.Notify(491828)
					}
					__antithesis_instrumentation__.Notify(491826)
					fs.mu.queue.Remove(e)
					n := e.Value.(*flowWithCtx)

					n.cleanupBeforeRun()
				}
				__antithesis_instrumentation__.Notify(491822)

				if atomic.LoadInt32(&fs.atomics.numRunning) == 0 {
					__antithesis_instrumentation__.Notify(491829)
					return
				} else {
					__antithesis_instrumentation__.Notify(491830)
				}
			} else {
				__antithesis_instrumentation__.Notify(491831)
			}
			__antithesis_instrumentation__.Notify(491819)
			fs.mu.Unlock()

			select {
			case <-fs.flowDoneCh:
				__antithesis_instrumentation__.Notify(491832)
				fs.mu.Lock()

				decrementNumRunning := stopped
				fs.metrics.FlowStop()
				if !stopped {
					__antithesis_instrumentation__.Notify(491836)
					if frElem := fs.mu.queue.Front(); frElem != nil {
						__antithesis_instrumentation__.Notify(491837)
						n := frElem.Value.(*flowWithCtx)
						fs.mu.queue.Remove(frElem)
						wait := timeutil.Since(n.enqueueTime)
						log.VEventf(
							n.ctx, 1, "flow scheduler dequeued flow %s, spent %s in queue", n.flow.GetID(), wait,
						)
						fs.metrics.FlowsQueued.Dec(1)
						fs.metrics.QueueWaitHist.RecordValue(int64(wait))

						if err := fs.runFlowNow(n.ctx, n.flow, true); err != nil {
							__antithesis_instrumentation__.Notify(491838)
							log.Errorf(n.ctx, "error starting queued flow: %s", err)
						} else {
							__antithesis_instrumentation__.Notify(491839)
						}
					} else {
						__antithesis_instrumentation__.Notify(491840)
						decrementNumRunning = true
					}
				} else {
					__antithesis_instrumentation__.Notify(491841)
				}
				__antithesis_instrumentation__.Notify(491833)
				if decrementNumRunning {
					__antithesis_instrumentation__.Notify(491842)
					atomic.AddInt32(&fs.atomics.numRunning, -1)
				} else {
					__antithesis_instrumentation__.Notify(491843)
				}

			case <-quiesceCh:
				__antithesis_instrumentation__.Notify(491834)
				fs.mu.Lock()
				stopped = true
				if l := atomic.LoadInt32(&fs.atomics.numRunning); l != 0 {
					__antithesis_instrumentation__.Notify(491844)
					log.Infof(ctx, "waiting for %d running flows", l)
				} else {
					__antithesis_instrumentation__.Notify(491845)
				}
				__antithesis_instrumentation__.Notify(491835)

				quiesceCh = nil
			}
		}
	})
}

func (fs *FlowScheduler) Serialize() (
	running []execinfrapb.DistSQLRemoteFlowInfo,
	queued []execinfrapb.DistSQLRemoteFlowInfo,
) {
	__antithesis_instrumentation__.Notify(491846)
	fs.mu.Lock()
	defer fs.mu.Unlock()
	running = make([]execinfrapb.DistSQLRemoteFlowInfo, 0, len(fs.mu.runningFlows))
	for _, info := range fs.mu.runningFlows {
		__antithesis_instrumentation__.Notify(491849)
		running = append(running, info)
	}
	__antithesis_instrumentation__.Notify(491847)
	if fs.mu.queue.Len() > 0 {
		__antithesis_instrumentation__.Notify(491850)
		queued = make([]execinfrapb.DistSQLRemoteFlowInfo, 0, fs.mu.queue.Len())
		for e := fs.mu.queue.Front(); e != nil; e = e.Next() {
			__antithesis_instrumentation__.Notify(491851)
			f := e.Value.(*flowWithCtx)
			queued = append(queued, execinfrapb.DistSQLRemoteFlowInfo{
				FlowID:       f.flow.GetID(),
				Timestamp:    f.enqueueTime,
				StatementSQL: f.flow.StatementSQL(),
			})
		}
	} else {
		__antithesis_instrumentation__.Notify(491852)
	}
	__antithesis_instrumentation__.Notify(491848)
	return running, queued
}
