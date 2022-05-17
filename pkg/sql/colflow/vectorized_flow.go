package colflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/coldataext"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/colcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecargs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colflow/colrpc"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/marusama/semaphore"
)

type countingSemaphore struct {
	semaphore.Semaphore
	globalCount *metric.Gauge
	count       int64
}

var countingSemaphorePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(456382)
		return &countingSemaphore{}
	},
}

func newCountingSemaphore(sem semaphore.Semaphore, globalCount *metric.Gauge) *countingSemaphore {
	__antithesis_instrumentation__.Notify(456383)
	s := countingSemaphorePool.Get().(*countingSemaphore)
	s.Semaphore = sem
	s.globalCount = globalCount
	return s
}

func (s *countingSemaphore) Acquire(ctx context.Context, n int) error {
	__antithesis_instrumentation__.Notify(456384)
	if err := s.Semaphore.Acquire(ctx, n); err != nil {
		__antithesis_instrumentation__.Notify(456386)
		return err
	} else {
		__antithesis_instrumentation__.Notify(456387)
	}
	__antithesis_instrumentation__.Notify(456385)
	atomic.AddInt64(&s.count, int64(n))
	s.globalCount.Inc(int64(n))
	return nil
}

func (s *countingSemaphore) TryAcquire(n int) bool {
	__antithesis_instrumentation__.Notify(456388)
	success := s.Semaphore.TryAcquire(n)
	if !success {
		__antithesis_instrumentation__.Notify(456390)
		return false
	} else {
		__antithesis_instrumentation__.Notify(456391)
	}
	__antithesis_instrumentation__.Notify(456389)
	atomic.AddInt64(&s.count, int64(n))
	s.globalCount.Inc(int64(n))
	return success
}

func (s *countingSemaphore) Release(n int) int {
	__antithesis_instrumentation__.Notify(456392)
	atomic.AddInt64(&s.count, int64(-n))
	s.globalCount.Dec(int64(n))
	return s.Semaphore.Release(n)
}

func (s *countingSemaphore) ReleaseToPool() {
	__antithesis_instrumentation__.Notify(456393)
	if unreleased := atomic.LoadInt64(&s.count); unreleased != 0 {
		__antithesis_instrumentation__.Notify(456395)
		colexecerror.InternalError(errors.Newf("unexpectedly %d count on the semaphore when releasing it to the pool", unreleased))
	} else {
		__antithesis_instrumentation__.Notify(456396)
	}
	__antithesis_instrumentation__.Notify(456394)
	*s = countingSemaphore{}
	countingSemaphorePool.Put(s)
}

type vectorizedFlow struct {
	*flowinfra.FlowBase

	creator *vectorizedFlowCreator

	batchFlowCoordinator *BatchFlowCoordinator

	countingSemaphore *countingSemaphore

	tempStorage struct {
		syncutil.Mutex

		path string
	}

	testingInfo struct {
		numClosers int32

		numClosed *int32
	}

	testingKnobs struct {
		onSetupFlow func(*vectorizedFlowCreator)
	}
}

var _ flowinfra.Flow = &vectorizedFlow{}
var _ execinfra.Releasable = &vectorizedFlow{}

var vectorizedFlowPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(456397)
		return &vectorizedFlow{}
	},
}

func NewVectorizedFlow(base *flowinfra.FlowBase) flowinfra.Flow {
	__antithesis_instrumentation__.Notify(456398)
	vf := vectorizedFlowPool.Get().(*vectorizedFlow)
	vf.FlowBase = base
	return vf
}

func (f *vectorizedFlow) Setup(
	ctx context.Context, spec *execinfrapb.FlowSpec, opt flowinfra.FuseOpt,
) (context.Context, execinfra.OpChains, error) {
	__antithesis_instrumentation__.Notify(456399)
	var err error
	ctx, _, err = f.FlowBase.Setup(ctx, spec, opt)
	if err != nil {
		__antithesis_instrumentation__.Notify(456406)
		return ctx, nil, err
	} else {
		__antithesis_instrumentation__.Notify(456407)
	}
	__antithesis_instrumentation__.Notify(456400)
	log.VEvent(ctx, 2, "setting up vectorized flow")
	recordingStats := false
	if execinfra.ShouldCollectStats(ctx, &f.FlowCtx) {
		__antithesis_instrumentation__.Notify(456408)
		recordingStats = true
	} else {
		__antithesis_instrumentation__.Notify(456409)
	}
	__antithesis_instrumentation__.Notify(456401)
	helper := newVectorizedFlowCreatorHelper(f.FlowBase)

	diskQueueCfg := colcontainer.DiskQueueCfg{
		FS:                  f.Cfg.TempFS,
		GetPather:           f,
		SpilledBytesWritten: f.Cfg.Metrics.SpilledBytesWritten,
		SpilledBytesRead:    f.Cfg.Metrics.SpilledBytesRead,
	}
	if err := diskQueueCfg.EnsureDefaults(); err != nil {
		__antithesis_instrumentation__.Notify(456410)
		return ctx, nil, err
	} else {
		__antithesis_instrumentation__.Notify(456411)
	}
	__antithesis_instrumentation__.Notify(456402)
	f.countingSemaphore = newCountingSemaphore(f.Cfg.VecFDSemaphore, f.Cfg.Metrics.VecOpenFDs)
	flowCtx := f.GetFlowCtx()
	f.creator = newVectorizedFlowCreator(
		helper,
		vectorizedRemoteComponentCreator{},
		recordingStats,
		f.Gateway,
		f.GetWaitGroup(),
		f.GetRowSyncFlowConsumer(),
		f.GetBatchSyncFlowConsumer(),
		flowCtx.Cfg.PodNodeDialer,
		f.GetID(),
		diskQueueCfg,
		f.countingSemaphore,
		flowCtx.NewTypeResolver(flowCtx.EvalCtx.Txn),
		f.FlowBase.GetAdmissionInfo(),
	)
	if f.testingKnobs.onSetupFlow != nil {
		__antithesis_instrumentation__.Notify(456412)
		f.testingKnobs.onSetupFlow(f.creator)
	} else {
		__antithesis_instrumentation__.Notify(456413)
	}
	__antithesis_instrumentation__.Notify(456403)
	opChains, batchFlowCoordinator, err := f.creator.setupFlow(ctx, flowCtx, spec.Processors, f.GetLocalProcessors(), opt)
	if err != nil {
		__antithesis_instrumentation__.Notify(456414)

		f.creator.cleanup(ctx)
		f.creator.Release()
		log.VEventf(ctx, 1, "failed to vectorize: %v", err)
		return ctx, nil, err
	} else {
		__antithesis_instrumentation__.Notify(456415)
	}
	__antithesis_instrumentation__.Notify(456404)
	f.batchFlowCoordinator = batchFlowCoordinator
	f.testingInfo.numClosers = f.creator.numClosers
	f.testingInfo.numClosed = &f.creator.numClosed
	f.SetStartedGoroutines(f.creator.operatorConcurrency)
	log.VEventf(ctx, 2, "vectorized flow setup succeeded")
	if !f.IsLocal() {
		__antithesis_instrumentation__.Notify(456416)

		opChains = nil
	} else {
		__antithesis_instrumentation__.Notify(456417)
	}
	__antithesis_instrumentation__.Notify(456405)
	return ctx, opChains, nil
}

func (f *vectorizedFlow) Run(ctx context.Context, doneFn func()) {
	__antithesis_instrumentation__.Notify(456418)
	if f.batchFlowCoordinator == nil {
		__antithesis_instrumentation__.Notify(456421)

		f.FlowBase.Run(ctx, doneFn)
		return
	} else {
		__antithesis_instrumentation__.Notify(456422)
	}
	__antithesis_instrumentation__.Notify(456419)

	defer f.Wait()

	if err := f.StartInternal(ctx, nil, doneFn); err != nil {
		__antithesis_instrumentation__.Notify(456423)
		f.GetRowSyncFlowConsumer().Push(nil, &execinfrapb.ProducerMetadata{Err: err})
		f.GetRowSyncFlowConsumer().ProducerDone()
		return
	} else {
		__antithesis_instrumentation__.Notify(456424)
	}
	__antithesis_instrumentation__.Notify(456420)

	log.VEvent(ctx, 1, "running the batch flow coordinator in the flow's goroutine")
	f.batchFlowCoordinator.Run(ctx)
}

var _ colcontainer.GetPather = &vectorizedFlow{}

func (f *vectorizedFlow) GetPath(ctx context.Context) string {
	__antithesis_instrumentation__.Notify(456425)
	f.tempStorage.Lock()
	defer f.tempStorage.Unlock()
	if f.tempStorage.path != "" {
		__antithesis_instrumentation__.Notify(456428)

		return f.tempStorage.path
	} else {
		__antithesis_instrumentation__.Notify(456429)
	}
	__antithesis_instrumentation__.Notify(456426)

	tempDirName := f.GetID().String()
	f.tempStorage.path = filepath.Join(f.Cfg.TempStoragePath, tempDirName)
	log.VEventf(ctx, 1, "flow %s spilled to disk, stack trace: %s", f.ID, util.GetSmallTrace(2))
	if err := f.Cfg.TempFS.MkdirAll(f.tempStorage.path); err != nil {
		__antithesis_instrumentation__.Notify(456430)
		colexecerror.InternalError(errors.Wrap(err, "unable to create temporary storage directory"))
	} else {
		__antithesis_instrumentation__.Notify(456431)
	}
	__antithesis_instrumentation__.Notify(456427)

	f.Cfg.Metrics.QueriesSpilled.Inc(1)
	return f.tempStorage.path
}

func (f *vectorizedFlow) IsVectorized() bool {
	__antithesis_instrumentation__.Notify(456432)
	return true
}

func (f *vectorizedFlow) ConcurrentTxnUse() bool {
	__antithesis_instrumentation__.Notify(456433)
	return f.creator.operatorConcurrency || func() bool {
		__antithesis_instrumentation__.Notify(456434)
		return f.FlowBase.ConcurrentTxnUse() == true
	}() == true
}

func (f *vectorizedFlow) Release() {
	__antithesis_instrumentation__.Notify(456435)
	f.creator.Release()
	f.countingSemaphore.ReleaseToPool()
	*f = vectorizedFlow{}
	vectorizedFlowPool.Put(f)
}

func (f *vectorizedFlow) Cleanup(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456436)

	f.creator.cleanup(ctx)

	if buildutil.CrdbTestBuild && func() bool {
		__antithesis_instrumentation__.Notify(456440)
		return f.FlowBase.Started() == true
	}() == true {
		__antithesis_instrumentation__.Notify(456441)

		if numClosed := atomic.LoadInt32(f.testingInfo.numClosed); numClosed != f.testingInfo.numClosers {
			__antithesis_instrumentation__.Notify(456442)
			colexecerror.InternalError(errors.AssertionFailedf("expected %d components to be closed, but found that only %d were", f.testingInfo.numClosers, numClosed))
		} else {
			__antithesis_instrumentation__.Notify(456443)
		}
	} else {
		__antithesis_instrumentation__.Notify(456444)
	}
	__antithesis_instrumentation__.Notify(456437)

	f.tempStorage.Lock()
	created := f.tempStorage.path != ""
	f.tempStorage.Unlock()
	if created {
		__antithesis_instrumentation__.Notify(456445)
		if err := f.Cfg.TempFS.RemoveAll(f.GetPath(ctx)); err != nil {
			__antithesis_instrumentation__.Notify(456446)

			log.Warningf(
				ctx,
				"unable to remove flow %s's temporary directory at %s, files may be left over: %v",
				f.GetID().Short(),
				f.GetPath(ctx),
				err,
			)
		} else {
			__antithesis_instrumentation__.Notify(456447)
		}
	} else {
		__antithesis_instrumentation__.Notify(456448)
	}
	__antithesis_instrumentation__.Notify(456438)

	if unreleased := atomic.LoadInt64(&f.countingSemaphore.count); unreleased > 0 {
		__antithesis_instrumentation__.Notify(456449)
		f.countingSemaphore.Release(int(unreleased))
	} else {
		__antithesis_instrumentation__.Notify(456450)
	}
	__antithesis_instrumentation__.Notify(456439)
	f.FlowBase.Cleanup(ctx)
	f.Release()
}

func (s *vectorizedFlowCreator) wrapWithVectorizedStatsCollectorBase(
	op *colexecargs.OpWithMetaInfo,
	kvReader colexecop.KVReader,
	columnarizer colexecop.VectorizedStatsCollector,
	inputs []colexecargs.OpWithMetaInfo,
	component execinfrapb.ComponentID,
	monitors []*mon.BytesMonitor,
) error {
	__antithesis_instrumentation__.Notify(456451)
	inputWatch := timeutil.NewStopWatch()
	var memMonitors, diskMonitors []*mon.BytesMonitor
	for _, m := range monitors {
		__antithesis_instrumentation__.Notify(456454)
		if m.Resource() == mon.DiskResource {
			__antithesis_instrumentation__.Notify(456455)
			diskMonitors = append(diskMonitors, m)
		} else {
			__antithesis_instrumentation__.Notify(456456)
			memMonitors = append(memMonitors, m)
		}
	}
	__antithesis_instrumentation__.Notify(456452)
	inputStatsCollectors := make([]childStatsCollector, len(inputs))
	for i, input := range inputs {
		__antithesis_instrumentation__.Notify(456457)
		sc, ok := input.Root.(childStatsCollector)
		if !ok {
			__antithesis_instrumentation__.Notify(456459)
			return errors.New("unexpectedly an input is not collecting stats")
		} else {
			__antithesis_instrumentation__.Notify(456460)
		}
		__antithesis_instrumentation__.Notify(456458)
		inputStatsCollectors[i] = sc
	}
	__antithesis_instrumentation__.Notify(456453)
	vsc := newVectorizedStatsCollector(
		op.Root, kvReader, columnarizer, component, inputWatch,
		memMonitors, diskMonitors, inputStatsCollectors,
	)
	op.Root = vsc
	op.StatsCollectors = append(op.StatsCollectors, vsc)
	maybeAddStatsInvariantChecker(op)
	return nil
}

func (s *vectorizedFlowCreator) wrapWithNetworkVectorizedStatsCollector(
	op *colexecargs.OpWithMetaInfo,
	inbox *colrpc.Inbox,
	component execinfrapb.ComponentID,
	latency time.Duration,
) {
	__antithesis_instrumentation__.Notify(456461)
	inputWatch := timeutil.NewStopWatch()
	nvsc := newNetworkVectorizedStatsCollector(op.Root, component, inputWatch, inbox, latency)
	op.Root = nvsc
	op.StatsCollectors = []colexecop.VectorizedStatsCollector{nvsc}
	maybeAddStatsInvariantChecker(op)
}

func (s *vectorizedFlowCreator) makeGetStatsFnForOutbox(
	flowCtx *execinfra.FlowCtx,
	statsCollectors []colexecop.VectorizedStatsCollector,
	originSQLInstanceID base.SQLInstanceID,
) func() []*execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(456462)
	if !s.recordingStats {
		__antithesis_instrumentation__.Notify(456464)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(456465)
	}
	__antithesis_instrumentation__.Notify(456463)
	return func() []*execinfrapb.ComponentStats {
		__antithesis_instrumentation__.Notify(456466)
		lastOutboxOnRemoteNode := atomic.AddInt32(&s.numOutboxesDrained, 1) == atomic.LoadInt32(&s.numOutboxes) && func() bool {
			__antithesis_instrumentation__.Notify(456470)
			return !s.isGatewayNode == true
		}() == true
		numResults := len(statsCollectors)
		if lastOutboxOnRemoteNode {
			__antithesis_instrumentation__.Notify(456471)
			numResults++
		} else {
			__antithesis_instrumentation__.Notify(456472)
		}
		__antithesis_instrumentation__.Notify(456467)
		result := make([]*execinfrapb.ComponentStats, 0, numResults)
		for _, s := range statsCollectors {
			__antithesis_instrumentation__.Notify(456473)
			result = append(result, s.GetStats())
		}
		__antithesis_instrumentation__.Notify(456468)
		if lastOutboxOnRemoteNode {
			__antithesis_instrumentation__.Notify(456474)

			result = append(result, &execinfrapb.ComponentStats{
				Component: execinfrapb.FlowComponentID(originSQLInstanceID, flowCtx.ID),
				FlowStats: execinfrapb.FlowStats{
					MaxMemUsage:  optional.MakeUint(uint64(flowCtx.EvalCtx.Mon.MaximumBytes())),
					MaxDiskUsage: optional.MakeUint(uint64(flowCtx.DiskMonitor.MaximumBytes())),
				},
			})
		} else {
			__antithesis_instrumentation__.Notify(456475)
		}
		__antithesis_instrumentation__.Notify(456469)
		return result
	}
}

type runFn func(_ context.Context, flowCtxCancel context.CancelFunc)

type flowCreatorHelper interface {
	execinfra.Releasable

	addStreamEndpoint(execinfrapb.StreamID, *colrpc.Inbox, *sync.WaitGroup)

	checkInboundStreamID(execinfrapb.StreamID) error

	accumulateAsyncComponent(runFn)

	addFlowCoordinator(coordinator *FlowCoordinator)

	getFlowCtxDone() <-chan struct{}

	getCancelFlowFn() context.CancelFunc
}

type admissionOptions struct {
	admissionQ    *admission.WorkQueue
	admissionInfo admission.WorkInfo
}

type remoteComponentCreator interface {
	newOutbox(
		allocator *colmem.Allocator,
		input colexecargs.OpWithMetaInfo,
		typs []*types.T,
		getStats func() []*execinfrapb.ComponentStats,
	) (*colrpc.Outbox, error)
	newInbox(
		allocator *colmem.Allocator,
		typs []*types.T,
		streamID execinfrapb.StreamID,
		flowCtxDone <-chan struct{},
		admissionOpts admissionOptions,
	) (*colrpc.Inbox, error)
}

type vectorizedRemoteComponentCreator struct{}

func (vectorizedRemoteComponentCreator) newOutbox(
	allocator *colmem.Allocator,
	input colexecargs.OpWithMetaInfo,
	typs []*types.T,
	getStats func() []*execinfrapb.ComponentStats,
) (*colrpc.Outbox, error) {
	__antithesis_instrumentation__.Notify(456476)
	return colrpc.NewOutbox(allocator, input, typs, getStats)
}

func (vectorizedRemoteComponentCreator) newInbox(
	allocator *colmem.Allocator,
	typs []*types.T,
	streamID execinfrapb.StreamID,
	flowCtxDone <-chan struct{},
	admissionOpts admissionOptions,
) (*colrpc.Inbox, error) {
	__antithesis_instrumentation__.Notify(456477)
	return colrpc.NewInboxWithAdmissionControl(
		allocator, typs, streamID, flowCtxDone,
		admissionOpts.admissionQ, admissionOpts.admissionInfo,
	)
}

type vectorizedFlowCreator struct {
	flowCreatorHelper
	remoteComponentCreator

	rowReceiver execinfra.RowReceiver

	batchReceiver execinfra.BatchReceiver

	batchFlowCoordinator *BatchFlowCoordinator

	streamIDToInputOp map[execinfrapb.StreamID]colexecargs.OpWithMetaInfo
	streamIDToSpecIdx map[execinfrapb.StreamID]int
	recordingStats    bool
	isGatewayNode     bool
	waitGroup         *sync.WaitGroup
	nodeDialer        *nodedialer.Dialer
	flowID            execinfrapb.FlowID
	exprHelper        *colexecargs.ExprHelper
	typeResolver      descs.DistSQLTypeResolver
	admissionInfo     admission.WorkInfo

	numOutboxes int32

	numOutboxesDrained int32

	procIdxQueue []int

	opChains execinfra.OpChains

	operatorConcurrency bool

	releasables []execinfra.Releasable

	monitorRegistry colexecargs.MonitorRegistry
	diskQueueCfg    colcontainer.DiskQueueCfg
	fdSemaphore     semaphore.Semaphore

	numClosers int32
	numClosed  int32
}

var _ execinfra.Releasable = &vectorizedFlowCreator{}

var vectorizedFlowCreatorPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(456478)
		return &vectorizedFlowCreator{
			streamIDToInputOp: make(map[execinfrapb.StreamID]colexecargs.OpWithMetaInfo),
			streamIDToSpecIdx: make(map[execinfrapb.StreamID]int),
			exprHelper:        colexecargs.NewExprHelper(),
		}
	},
}

func newVectorizedFlowCreator(
	helper flowCreatorHelper,
	componentCreator remoteComponentCreator,
	recordingStats bool,
	isGatewayNode bool,
	waitGroup *sync.WaitGroup,
	rowSyncFlowConsumer execinfra.RowReceiver,
	batchSyncFlowConsumer execinfra.BatchReceiver,
	nodeDialer *nodedialer.Dialer,
	flowID execinfrapb.FlowID,
	diskQueueCfg colcontainer.DiskQueueCfg,
	fdSemaphore semaphore.Semaphore,
	typeResolver descs.DistSQLTypeResolver,
	admissionInfo admission.WorkInfo,
) *vectorizedFlowCreator {
	__antithesis_instrumentation__.Notify(456479)
	creator := vectorizedFlowCreatorPool.Get().(*vectorizedFlowCreator)
	*creator = vectorizedFlowCreator{
		flowCreatorHelper:      helper,
		remoteComponentCreator: componentCreator,
		streamIDToInputOp:      creator.streamIDToInputOp,
		streamIDToSpecIdx:      creator.streamIDToSpecIdx,
		recordingStats:         recordingStats,
		isGatewayNode:          isGatewayNode,
		waitGroup:              waitGroup,
		rowReceiver:            rowSyncFlowConsumer,
		batchReceiver:          batchSyncFlowConsumer,
		nodeDialer:             nodeDialer,
		flowID:                 flowID,
		exprHelper:             creator.exprHelper,
		typeResolver:           typeResolver,
		admissionInfo:          admissionInfo,
		procIdxQueue:           creator.procIdxQueue,
		opChains:               creator.opChains,
		releasables:            creator.releasables,
		monitorRegistry:        creator.monitorRegistry,
		diskQueueCfg:           diskQueueCfg,
		fdSemaphore:            fdSemaphore,
	}
	return creator
}

func (s *vectorizedFlowCreator) cleanup(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456480)
	s.monitorRegistry.Close(ctx)
}

func (s *vectorizedFlowCreator) Release() {
	__antithesis_instrumentation__.Notify(456481)
	for k := range s.streamIDToInputOp {
		__antithesis_instrumentation__.Notify(456488)
		delete(s.streamIDToInputOp, k)
	}
	__antithesis_instrumentation__.Notify(456482)
	for k := range s.streamIDToSpecIdx {
		__antithesis_instrumentation__.Notify(456489)
		delete(s.streamIDToSpecIdx, k)
	}
	__antithesis_instrumentation__.Notify(456483)
	s.flowCreatorHelper.Release()
	for _, r := range s.releasables {
		__antithesis_instrumentation__.Notify(456490)
		r.Release()
	}
	__antithesis_instrumentation__.Notify(456484)

	for i := range s.opChains {
		__antithesis_instrumentation__.Notify(456491)
		s.opChains[i] = nil
	}
	__antithesis_instrumentation__.Notify(456485)
	for i := range s.releasables {
		__antithesis_instrumentation__.Notify(456492)
		s.releasables[i] = nil
	}
	__antithesis_instrumentation__.Notify(456486)
	if s.exprHelper != nil {
		__antithesis_instrumentation__.Notify(456493)
		s.exprHelper.SemaCtx = nil
	} else {
		__antithesis_instrumentation__.Notify(456494)
	}
	__antithesis_instrumentation__.Notify(456487)
	s.monitorRegistry.Reset()
	*s = vectorizedFlowCreator{
		streamIDToInputOp: s.streamIDToInputOp,
		streamIDToSpecIdx: s.streamIDToSpecIdx,
		exprHelper:        s.exprHelper,

		procIdxQueue:    s.procIdxQueue[:0],
		opChains:        s.opChains[:0],
		releasables:     s.releasables[:0],
		monitorRegistry: s.monitorRegistry,
	}
	vectorizedFlowCreatorPool.Put(s)
}

func (s *vectorizedFlowCreator) setupRemoteOutputStream(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	op colexecargs.OpWithMetaInfo,
	outputTyps []*types.T,
	stream *execinfrapb.StreamEndpointSpec,
	factory coldata.ColumnFactory,
	getStats func() []*execinfrapb.ComponentStats,
) (execinfra.OpNode, error) {
	__antithesis_instrumentation__.Notify(456495)
	outbox, err := s.remoteComponentCreator.newOutbox(
		colmem.NewAllocator(ctx, s.monitorRegistry.NewStreamingMemAccount(flowCtx), factory),
		op, outputTyps, getStats,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(456498)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456499)
	}
	__antithesis_instrumentation__.Notify(456496)

	atomic.AddInt32(&s.numOutboxes, 1)
	run := func(ctx context.Context, flowCtxCancel context.CancelFunc) {
		__antithesis_instrumentation__.Notify(456500)
		outbox.Run(
			ctx,
			s.nodeDialer,
			stream.TargetNodeID,
			s.flowID,
			stream.StreamID,
			flowCtxCancel,
			flowinfra.SettingFlowStreamTimeout.Get(&flowCtx.Cfg.Settings.SV),
		)
	}
	__antithesis_instrumentation__.Notify(456497)
	s.accumulateAsyncComponent(run)
	return outbox, nil
}

func (s *vectorizedFlowCreator) setupRouter(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	input colexecargs.OpWithMetaInfo,
	outputTyps []*types.T,
	output *execinfrapb.OutputRouterSpec,
	factory coldata.ColumnFactory,
) error {
	__antithesis_instrumentation__.Notify(456501)
	if output.Type != execinfrapb.OutputRouterSpec_BY_HASH {
		__antithesis_instrumentation__.Notify(456508)
		return errors.Errorf("vectorized output router type %s unsupported", output.Type)
	} else {
		__antithesis_instrumentation__.Notify(456509)
	}
	__antithesis_instrumentation__.Notify(456502)

	streamIDs := make([]string, len(output.Streams))
	for i, s := range output.Streams {
		__antithesis_instrumentation__.Notify(456510)
		streamIDs[i] = strconv.Itoa(int(s.StreamID))
	}
	__antithesis_instrumentation__.Notify(456503)
	mmName := "hash-router-[" + strings.Join(streamIDs, ",") + "]"

	hashRouterMemMonitor, accounts := s.monitorRegistry.CreateUnlimitedMemAccounts(ctx, flowCtx, mmName, len(output.Streams))
	allocators := make([]*colmem.Allocator, len(output.Streams))
	for i := range allocators {
		__antithesis_instrumentation__.Notify(456511)
		allocators[i] = colmem.NewAllocator(ctx, accounts[i], factory)
	}
	__antithesis_instrumentation__.Notify(456504)
	diskMon, diskAccounts := s.monitorRegistry.CreateDiskAccounts(ctx, flowCtx, mmName, len(output.Streams))
	router, outputs := NewHashRouter(
		allocators, input, outputTyps, output.HashColumns, execinfra.GetWorkMemLimit(flowCtx),
		s.diskQueueCfg, s.fdSemaphore, diskAccounts,
	)
	runRouter := func(ctx context.Context, _ context.CancelFunc) {
		__antithesis_instrumentation__.Notify(456512)
		router.Run(logtags.AddTag(ctx, "hashRouterID", strings.Join(streamIDs, ",")))
	}
	__antithesis_instrumentation__.Notify(456505)
	s.accumulateAsyncComponent(runRouter)

	foundLocalOutput := false
	for i, op := range outputs {
		__antithesis_instrumentation__.Notify(456513)
		if buildutil.CrdbTestBuild {
			__antithesis_instrumentation__.Notify(456515)
			op = colexec.NewInvariantsChecker(op)
		} else {
			__antithesis_instrumentation__.Notify(456516)
		}
		__antithesis_instrumentation__.Notify(456514)
		stream := &output.Streams[i]
		switch stream.Type {
		case execinfrapb.StreamEndpointSpec_SYNC_RESPONSE:
			__antithesis_instrumentation__.Notify(456517)
			return errors.Errorf("unexpected sync response output when setting up router")
		case execinfrapb.StreamEndpointSpec_REMOTE:
			__antithesis_instrumentation__.Notify(456518)

			if _, err := s.setupRemoteOutputStream(
				ctx, flowCtx, colexecargs.OpWithMetaInfo{
					Root:            op,
					MetadataSources: colexecop.MetadataSources{op},
				}, outputTyps, stream, factory, nil,
			); err != nil {
				__antithesis_instrumentation__.Notify(456522)
				return err
			} else {
				__antithesis_instrumentation__.Notify(456523)
			}
		case execinfrapb.StreamEndpointSpec_LOCAL:
			__antithesis_instrumentation__.Notify(456519)
			foundLocalOutput = true
			opWithMetaInfo := colexecargs.OpWithMetaInfo{
				Root:            op,
				MetadataSources: colexecop.MetadataSources{op},

				ToClose: nil,
			}
			if s.recordingStats {
				__antithesis_instrumentation__.Notify(456524)
				mons := []*mon.BytesMonitor{hashRouterMemMonitor, diskMon}

				if err := s.wrapWithVectorizedStatsCollectorBase(
					&opWithMetaInfo, nil, nil,
					nil, flowCtx.StreamComponentID(stream.StreamID), mons,
				); err != nil {
					__antithesis_instrumentation__.Notify(456525)
					return err
				} else {
					__antithesis_instrumentation__.Notify(456526)
				}
			} else {
				__antithesis_instrumentation__.Notify(456527)
			}
			__antithesis_instrumentation__.Notify(456520)
			s.streamIDToInputOp[stream.StreamID] = opWithMetaInfo
		default:
			__antithesis_instrumentation__.Notify(456521)
		}
	}
	__antithesis_instrumentation__.Notify(456506)
	if !foundLocalOutput {
		__antithesis_instrumentation__.Notify(456528)

		s.opChains = append(s.opChains, router)
	} else {
		__antithesis_instrumentation__.Notify(456529)
	}
	__antithesis_instrumentation__.Notify(456507)
	return nil
}

func (s *vectorizedFlowCreator) setupInput(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	input execinfrapb.InputSyncSpec,
	opt flowinfra.FuseOpt,
	factory coldata.ColumnFactory,
) (colexecargs.OpWithMetaInfo, error) {
	__antithesis_instrumentation__.Notify(456530)
	inputStreamOps := make([]colexecargs.OpWithMetaInfo, 0, len(input.Streams))

	if err := s.typeResolver.HydrateTypeSlice(ctx, input.ColumnTypes); err != nil {
		__antithesis_instrumentation__.Notify(456534)
		return colexecargs.OpWithMetaInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(456535)
	}
	__antithesis_instrumentation__.Notify(456531)

	for _, inputStream := range input.Streams {
		__antithesis_instrumentation__.Notify(456536)
		switch inputStream.Type {
		case execinfrapb.StreamEndpointSpec_LOCAL:
			__antithesis_instrumentation__.Notify(456537)
			in := s.streamIDToInputOp[inputStream.StreamID]
			inputStreamOps = append(inputStreamOps, in)
		case execinfrapb.StreamEndpointSpec_REMOTE:
			__antithesis_instrumentation__.Notify(456538)

			if err := s.checkInboundStreamID(inputStream.StreamID); err != nil {
				__antithesis_instrumentation__.Notify(456545)
				return colexecargs.OpWithMetaInfo{}, err
			} else {
				__antithesis_instrumentation__.Notify(456546)
			}
			__antithesis_instrumentation__.Notify(456539)

			latency, err := s.nodeDialer.Latency(roachpb.NodeID(inputStream.OriginNodeID))
			if err != nil {
				__antithesis_instrumentation__.Notify(456547)

				latency = 0
				log.VEventf(ctx, 1, "an error occurred during vectorized planning while getting latency: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(456548)
			}
			__antithesis_instrumentation__.Notify(456540)
			inbox, err := s.remoteComponentCreator.newInbox(
				colmem.NewAllocator(ctx, s.monitorRegistry.NewStreamingMemAccount(flowCtx), factory),
				input.ColumnTypes,
				inputStream.StreamID,
				s.flowCreatorHelper.getFlowCtxDone(),
				admissionOptions{
					admissionQ:    flowCtx.Cfg.SQLSQLResponseAdmissionQ,
					admissionInfo: s.admissionInfo,
				})

			if err != nil {
				__antithesis_instrumentation__.Notify(456549)
				return colexecargs.OpWithMetaInfo{}, err
			} else {
				__antithesis_instrumentation__.Notify(456550)
			}
			__antithesis_instrumentation__.Notify(456541)
			s.addStreamEndpoint(inputStream.StreamID, inbox, s.waitGroup)
			op := colexecop.Operator(inbox)
			ms := colexecop.MetadataSource(inbox)
			if buildutil.CrdbTestBuild {
				__antithesis_instrumentation__.Notify(456551)
				op = colexec.NewInvariantsChecker(op)
				ms = op.(colexecop.MetadataSource)
			} else {
				__antithesis_instrumentation__.Notify(456552)
			}
			__antithesis_instrumentation__.Notify(456542)
			opWithMetaInfo := colexecargs.OpWithMetaInfo{
				Root:            op,
				MetadataSources: colexecop.MetadataSources{ms},
			}
			if s.recordingStats {
				__antithesis_instrumentation__.Notify(456553)

				compID := execinfrapb.StreamComponentID(
					inputStream.OriginNodeID, flowCtx.ID, inputStream.StreamID,
				)
				s.wrapWithNetworkVectorizedStatsCollector(&opWithMetaInfo, inbox, compID, latency)
			} else {
				__antithesis_instrumentation__.Notify(456554)
			}
			__antithesis_instrumentation__.Notify(456543)
			inputStreamOps = append(inputStreamOps, opWithMetaInfo)
		default:
			__antithesis_instrumentation__.Notify(456544)
			return colexecargs.OpWithMetaInfo{}, errors.Errorf("unsupported input stream type %s", inputStream.Type)
		}
	}
	__antithesis_instrumentation__.Notify(456532)
	opWithMetaInfo := inputStreamOps[0]
	if len(inputStreamOps) > 1 {
		__antithesis_instrumentation__.Notify(456555)
		statsInputs := inputStreamOps
		if input.Type == execinfrapb.InputSyncSpec_ORDERED {
			__antithesis_instrumentation__.Notify(456558)
			os := colexec.NewOrderedSynchronizer(
				colmem.NewAllocator(ctx, s.monitorRegistry.NewStreamingMemAccount(flowCtx), factory),
				execinfra.GetWorkMemLimit(flowCtx), inputStreamOps,
				input.ColumnTypes, execinfrapb.ConvertToColumnOrdering(input.Ordering),
			)
			opWithMetaInfo = colexecargs.OpWithMetaInfo{
				Root:            os,
				MetadataSources: colexecop.MetadataSources{os},
				ToClose:         colexecop.Closers{os},
			}
		} else {
			__antithesis_instrumentation__.Notify(456559)
			if input.Type == execinfrapb.InputSyncSpec_SERIAL_UNORDERED || func() bool {
				__antithesis_instrumentation__.Notify(456560)
				return opt == flowinfra.FuseAggressively == true
			}() == true {
				__antithesis_instrumentation__.Notify(456561)
				sync := colexec.NewSerialUnorderedSynchronizer(inputStreamOps)
				opWithMetaInfo = colexecargs.OpWithMetaInfo{
					Root:            sync,
					MetadataSources: colexecop.MetadataSources{sync},
					ToClose:         colexecop.Closers{sync},
				}
			} else {
				__antithesis_instrumentation__.Notify(456562)

				sync := colexec.NewParallelUnorderedSynchronizer(inputStreamOps, s.waitGroup)
				sync.LocalPlan = flowCtx.Local
				opWithMetaInfo = colexecargs.OpWithMetaInfo{
					Root:            sync,
					MetadataSources: colexecop.MetadataSources{sync},
					ToClose:         colexecop.Closers{sync},
				}
				s.operatorConcurrency = true

				statsInputs = nil
			}
		}
		__antithesis_instrumentation__.Notify(456556)
		if buildutil.CrdbTestBuild {
			__antithesis_instrumentation__.Notify(456563)
			opWithMetaInfo.Root = colexec.NewInvariantsChecker(opWithMetaInfo.Root)
			opWithMetaInfo.MetadataSources[0] = opWithMetaInfo.Root.(colexecop.MetadataSource)
		} else {
			__antithesis_instrumentation__.Notify(456564)
		}
		__antithesis_instrumentation__.Notify(456557)
		if s.recordingStats {
			__antithesis_instrumentation__.Notify(456565)
			statsInputsAsOps := make([]colexecargs.OpWithMetaInfo, len(statsInputs))
			for i := range statsInputs {
				__antithesis_instrumentation__.Notify(456567)
				statsInputsAsOps[i].Root = statsInputs[i].Root
			}
			__antithesis_instrumentation__.Notify(456566)

			if err := s.wrapWithVectorizedStatsCollectorBase(
				&opWithMetaInfo, nil, nil,
				statsInputsAsOps, execinfrapb.ComponentID{}, nil,
			); err != nil {
				__antithesis_instrumentation__.Notify(456568)
				return colexecargs.OpWithMetaInfo{}, err
			} else {
				__antithesis_instrumentation__.Notify(456569)
			}
		} else {
			__antithesis_instrumentation__.Notify(456570)
		}
	} else {
		__antithesis_instrumentation__.Notify(456571)
	}
	__antithesis_instrumentation__.Notify(456533)
	return opWithMetaInfo, nil
}

func (s *vectorizedFlowCreator) setupOutput(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	pspec *execinfrapb.ProcessorSpec,
	opWithMetaInfo colexecargs.OpWithMetaInfo,
	opOutputTypes []*types.T,
	factory coldata.ColumnFactory,
) error {
	__antithesis_instrumentation__.Notify(456572)
	output := &pspec.Output[0]
	if output.Type != execinfrapb.OutputRouterSpec_PASS_THROUGH {
		__antithesis_instrumentation__.Notify(456576)
		return s.setupRouter(
			ctx,
			flowCtx,
			opWithMetaInfo,
			opOutputTypes,
			output,
			factory,
		)
	} else {
		__antithesis_instrumentation__.Notify(456577)
	}
	__antithesis_instrumentation__.Notify(456573)

	if len(output.Streams) != 1 {
		__antithesis_instrumentation__.Notify(456578)
		return errors.Errorf("unsupported multi outputstream proc (%d streams)", len(output.Streams))
	} else {
		__antithesis_instrumentation__.Notify(456579)
	}
	__antithesis_instrumentation__.Notify(456574)
	outputStream := &output.Streams[0]
	switch outputStream.Type {
	case execinfrapb.StreamEndpointSpec_LOCAL:
		__antithesis_instrumentation__.Notify(456580)
		s.streamIDToInputOp[outputStream.StreamID] = opWithMetaInfo
	case execinfrapb.StreamEndpointSpec_REMOTE:
		__antithesis_instrumentation__.Notify(456581)

		outbox, err := s.setupRemoteOutputStream(
			ctx, flowCtx, opWithMetaInfo, opOutputTypes, outputStream, factory,
			s.makeGetStatsFnForOutbox(flowCtx, opWithMetaInfo.StatsCollectors, outputStream.OriginNodeID),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(456585)
			return err
		} else {
			__antithesis_instrumentation__.Notify(456586)
		}
		__antithesis_instrumentation__.Notify(456582)

		s.opChains = append(s.opChains, outbox)
	case execinfrapb.StreamEndpointSpec_SYNC_RESPONSE:
		__antithesis_instrumentation__.Notify(456583)

		input := colbuilder.MaybeRemoveRootColumnarizer(opWithMetaInfo)
		if input == nil && func() bool {
			__antithesis_instrumentation__.Notify(456587)
			return s.batchReceiver != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(456588)

			s.batchFlowCoordinator = NewBatchFlowCoordinator(
				flowCtx,
				pspec.ProcessorID,
				opWithMetaInfo,
				s.batchReceiver,
				s.getCancelFlowFn(),
			)

			s.opChains = append(s.opChains, s.batchFlowCoordinator)
			s.releasables = append(s.releasables, s.batchFlowCoordinator)
		} else {
			__antithesis_instrumentation__.Notify(456589)

			if input != nil {
				__antithesis_instrumentation__.Notify(456591)

				if buildutil.CrdbTestBuild {
					__antithesis_instrumentation__.Notify(456592)

					s.numClosers--
				} else {
					__antithesis_instrumentation__.Notify(456593)
				}
			} else {
				__antithesis_instrumentation__.Notify(456594)
				input = colexec.NewMaterializerNoEvalCtxCopy(
					flowCtx,
					pspec.ProcessorID,
					opWithMetaInfo,
					opOutputTypes,
				)
			}
			__antithesis_instrumentation__.Notify(456590)

			f := NewFlowCoordinator(
				flowCtx,
				pspec.ProcessorID,
				input,
				s.rowReceiver,
				s.getCancelFlowFn(),
			)

			s.opChains = append(s.opChains, f)

			s.addFlowCoordinator(f)
		}

	default:
		__antithesis_instrumentation__.Notify(456584)
		return errors.Errorf("unsupported output stream type %s", outputStream.Type)
	}
	__antithesis_instrumentation__.Notify(456575)
	return nil
}

type callbackCloser struct {
	closeCb func(context.Context) error
}

var _ colexecop.Closer = &callbackCloser{}

func (c *callbackCloser) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(456595)
	return c.closeCb(ctx)
}

func (s *vectorizedFlowCreator) setupFlow(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	processorSpecs []execinfrapb.ProcessorSpec,
	localProcessors []execinfra.LocalProcessor,
	opt flowinfra.FuseOpt,
) (opChains execinfra.OpChains, batchFlowCoordinator *BatchFlowCoordinator, err error) {
	__antithesis_instrumentation__.Notify(456596)
	if vecErr := colexecerror.CatchVectorizedRuntimeError(func() {
		__antithesis_instrumentation__.Notify(456598)

		factory := coldataext.NewExtendedColumnFactory(flowCtx.EvalCtx)
		for i := range processorSpecs {
			__antithesis_instrumentation__.Notify(456600)
			hasLocalInput := false
			for j := range processorSpecs[i].Input {
				__antithesis_instrumentation__.Notify(456603)
				input := &processorSpecs[i].Input[j]
				for k := range input.Streams {
					__antithesis_instrumentation__.Notify(456604)
					stream := &input.Streams[k]
					s.streamIDToSpecIdx[stream.StreamID] = i
					if stream.Type != execinfrapb.StreamEndpointSpec_REMOTE {
						__antithesis_instrumentation__.Notify(456605)
						hasLocalInput = true
					} else {
						__antithesis_instrumentation__.Notify(456606)
					}
				}
			}
			__antithesis_instrumentation__.Notify(456601)
			if hasLocalInput {
				__antithesis_instrumentation__.Notify(456607)
				continue
			} else {
				__antithesis_instrumentation__.Notify(456608)
			}
			__antithesis_instrumentation__.Notify(456602)

			s.procIdxQueue = append(s.procIdxQueue, i)
		}
		__antithesis_instrumentation__.Notify(456599)

		for procIdxQueuePos := 0; procIdxQueuePos < len(processorSpecs); procIdxQueuePos++ {
			__antithesis_instrumentation__.Notify(456609)
			pspec := &processorSpecs[s.procIdxQueue[procIdxQueuePos]]
			if len(pspec.Output) > 1 {
				__antithesis_instrumentation__.Notify(456620)
				err = errors.Errorf("unsupported multi-output proc (%d outputs)", len(pspec.Output))
				return
			} else {
				__antithesis_instrumentation__.Notify(456621)
			}
			__antithesis_instrumentation__.Notify(456610)

			inputs := make([]colexecargs.OpWithMetaInfo, len(pspec.Input))
			for i := range pspec.Input {
				__antithesis_instrumentation__.Notify(456622)
				inputs[i], err = s.setupInput(ctx, flowCtx, pspec.Input[i], opt, factory)
				if err != nil {
					__antithesis_instrumentation__.Notify(456623)
					return
				} else {
					__antithesis_instrumentation__.Notify(456624)
				}
			}
			__antithesis_instrumentation__.Notify(456611)

			if err = s.typeResolver.HydrateTypeSlice(ctx, pspec.ResultTypes); err != nil {
				__antithesis_instrumentation__.Notify(456625)
				return
			} else {
				__antithesis_instrumentation__.Notify(456626)
			}
			__antithesis_instrumentation__.Notify(456612)

			args := &colexecargs.NewColOperatorArgs{
				Spec:                 pspec,
				Inputs:               inputs,
				StreamingMemAccount:  s.monitorRegistry.NewStreamingMemAccount(flowCtx),
				ProcessorConstructor: rowexec.NewProcessor,
				LocalProcessors:      localProcessors,
				DiskQueueCfg:         s.diskQueueCfg,
				FDSemaphore:          s.fdSemaphore,
				ExprHelper:           s.exprHelper,
				Factory:              factory,
				MonitorRegistry:      &s.monitorRegistry,
			}
			numOldMonitors := len(s.monitorRegistry.GetMonitors())
			if args.ExprHelper.SemaCtx == nil {
				__antithesis_instrumentation__.Notify(456627)
				args.ExprHelper.SemaCtx = flowCtx.NewSemaContext(flowCtx.EvalCtx.Txn)
			} else {
				__antithesis_instrumentation__.Notify(456628)
			}
			__antithesis_instrumentation__.Notify(456613)
			var result *colexecargs.NewColOperatorResult
			result, err = colbuilder.NewColOperator(ctx, flowCtx, args)
			if result != nil {
				__antithesis_instrumentation__.Notify(456629)
				s.releasables = append(s.releasables, result)
			} else {
				__antithesis_instrumentation__.Notify(456630)
			}
			__antithesis_instrumentation__.Notify(456614)
			if err != nil {
				__antithesis_instrumentation__.Notify(456631)
				err = errors.Wrapf(err, "unable to vectorize execution plan")
				return
			} else {
				__antithesis_instrumentation__.Notify(456632)
			}
			__antithesis_instrumentation__.Notify(456615)
			if flowCtx.EvalCtx.SessionData().TestingVectorizeInjectPanics {
				__antithesis_instrumentation__.Notify(456633)
				result.Root = newPanicInjector(result.Root)
			} else {
				__antithesis_instrumentation__.Notify(456634)
			}
			__antithesis_instrumentation__.Notify(456616)
			if buildutil.CrdbTestBuild {
				__antithesis_instrumentation__.Notify(456635)
				toCloseCopy := append(colexecop.Closers{}, result.ToClose...)
				for i := range toCloseCopy {
					__antithesis_instrumentation__.Notify(456637)
					func(idx int) {
						__antithesis_instrumentation__.Notify(456638)
						closed := false
						result.ToClose[idx] = &callbackCloser{closeCb: func(ctx context.Context) error {
							__antithesis_instrumentation__.Notify(456639)
							if !closed {
								__antithesis_instrumentation__.Notify(456641)
								closed = true
								atomic.AddInt32(&s.numClosed, 1)
							} else {
								__antithesis_instrumentation__.Notify(456642)
							}
							__antithesis_instrumentation__.Notify(456640)
							return toCloseCopy[idx].Close(ctx)
						}}
					}(i)
				}
				__antithesis_instrumentation__.Notify(456636)
				s.numClosers += int32(len(result.ToClose))
			} else {
				__antithesis_instrumentation__.Notify(456643)
			}
			__antithesis_instrumentation__.Notify(456617)

			if s.recordingStats {
				__antithesis_instrumentation__.Notify(456644)
				newMonitors := s.monitorRegistry.GetMonitors()[numOldMonitors:]
				if err := s.wrapWithVectorizedStatsCollectorBase(
					&result.OpWithMetaInfo, result.KVReader, result.Columnarizer, inputs,
					flowCtx.ProcessorComponentID(pspec.ProcessorID), newMonitors,
				); err != nil {
					__antithesis_instrumentation__.Notify(456645)
					return
				} else {
					__antithesis_instrumentation__.Notify(456646)
				}
			} else {
				__antithesis_instrumentation__.Notify(456647)
			}
			__antithesis_instrumentation__.Notify(456618)

			if err = s.setupOutput(
				ctx, flowCtx, pspec, result.OpWithMetaInfo, result.ColumnTypes, factory,
			); err != nil {
				__antithesis_instrumentation__.Notify(456648)
				return
			} else {
				__antithesis_instrumentation__.Notify(456649)
			}
			__antithesis_instrumentation__.Notify(456619)

		NEXTOUTPUT:
			for i := range pspec.Output {
				__antithesis_instrumentation__.Notify(456650)
				for j := range pspec.Output[i].Streams {
					__antithesis_instrumentation__.Notify(456651)
					outputStream := &pspec.Output[i].Streams[j]
					if outputStream.Type != execinfrapb.StreamEndpointSpec_LOCAL {
						__antithesis_instrumentation__.Notify(456655)
						continue
					} else {
						__antithesis_instrumentation__.Notify(456656)
					}
					__antithesis_instrumentation__.Notify(456652)
					procIdx, ok := s.streamIDToSpecIdx[outputStream.StreamID]
					if !ok {
						__antithesis_instrumentation__.Notify(456657)
						err = errors.Errorf("couldn't find stream %d", outputStream.StreamID)
						return
					} else {
						__antithesis_instrumentation__.Notify(456658)
					}
					__antithesis_instrumentation__.Notify(456653)
					outputSpec := &processorSpecs[procIdx]
					for k := range outputSpec.Input {
						__antithesis_instrumentation__.Notify(456659)
						for l := range outputSpec.Input[k].Streams {
							__antithesis_instrumentation__.Notify(456660)
							inputStream := outputSpec.Input[k].Streams[l]
							if inputStream.Type == execinfrapb.StreamEndpointSpec_REMOTE {
								__antithesis_instrumentation__.Notify(456662)

								continue
							} else {
								__antithesis_instrumentation__.Notify(456663)
							}
							__antithesis_instrumentation__.Notify(456661)
							if _, ok := s.streamIDToInputOp[inputStream.StreamID]; !ok {
								__antithesis_instrumentation__.Notify(456664)
								continue NEXTOUTPUT
							} else {
								__antithesis_instrumentation__.Notify(456665)
							}
						}
					}
					__antithesis_instrumentation__.Notify(456654)

					s.procIdxQueue = append(s.procIdxQueue, procIdx)
				}
			}
		}
	}); vecErr != nil {
		__antithesis_instrumentation__.Notify(456666)
		return s.opChains, s.batchFlowCoordinator, vecErr
	} else {
		__antithesis_instrumentation__.Notify(456667)
	}
	__antithesis_instrumentation__.Notify(456597)
	return s.opChains, s.batchFlowCoordinator, err
}

type vectorizedInboundStreamHandler struct {
	*colrpc.Inbox
}

var _ flowinfra.InboundStreamHandler = vectorizedInboundStreamHandler{}

func (s vectorizedInboundStreamHandler) Run(
	ctx context.Context,
	stream execinfrapb.DistSQL_FlowStreamServer,
	_ *execinfrapb.ProducerMessage,
	_ *flowinfra.FlowBase,
) error {
	__antithesis_instrumentation__.Notify(456668)
	return s.RunWithStream(ctx, stream)
}

func (s vectorizedInboundStreamHandler) Timeout(err error) {
	__antithesis_instrumentation__.Notify(456669)
	s.Inbox.Timeout(err)
}

type vectorizedFlowCreatorHelper struct {
	f          *flowinfra.FlowBase
	processors []execinfra.Processor
}

var _ flowCreatorHelper = &vectorizedFlowCreatorHelper{}

var vectorizedFlowCreatorHelperPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(456670)
		return &vectorizedFlowCreatorHelper{
			processors: make([]execinfra.Processor, 0, 1),
		}
	},
}

func newVectorizedFlowCreatorHelper(f *flowinfra.FlowBase) *vectorizedFlowCreatorHelper {
	__antithesis_instrumentation__.Notify(456671)
	helper := vectorizedFlowCreatorHelperPool.Get().(*vectorizedFlowCreatorHelper)
	helper.f = f
	return helper
}

func (r *vectorizedFlowCreatorHelper) addStreamEndpoint(
	streamID execinfrapb.StreamID, inbox *colrpc.Inbox, wg *sync.WaitGroup,
) {
	__antithesis_instrumentation__.Notify(456672)
	r.f.AddRemoteStream(streamID, flowinfra.NewInboundStreamInfo(
		vectorizedInboundStreamHandler{inbox},
		wg,
	))
}

func (r *vectorizedFlowCreatorHelper) checkInboundStreamID(sid execinfrapb.StreamID) error {
	__antithesis_instrumentation__.Notify(456673)
	return r.f.CheckInboundStreamID(sid)
}

func (r *vectorizedFlowCreatorHelper) accumulateAsyncComponent(run runFn) {
	__antithesis_instrumentation__.Notify(456674)
	r.f.AddStartable(
		flowinfra.StartableFn(func(ctx context.Context, wg *sync.WaitGroup, flowCtxCancel context.CancelFunc) {
			__antithesis_instrumentation__.Notify(456675)
			if wg != nil {
				__antithesis_instrumentation__.Notify(456677)
				wg.Add(1)
			} else {
				__antithesis_instrumentation__.Notify(456678)
			}
			__antithesis_instrumentation__.Notify(456676)
			go func() {
				__antithesis_instrumentation__.Notify(456679)
				run(ctx, flowCtxCancel)
				if wg != nil {
					__antithesis_instrumentation__.Notify(456680)
					wg.Done()
				} else {
					__antithesis_instrumentation__.Notify(456681)
				}
			}()
		}))
}

func (r *vectorizedFlowCreatorHelper) addFlowCoordinator(f *FlowCoordinator) {
	__antithesis_instrumentation__.Notify(456682)
	r.processors = append(r.processors, f)
	r.f.SetProcessors(r.processors)
}

func (r *vectorizedFlowCreatorHelper) getFlowCtxDone() <-chan struct{} {
	__antithesis_instrumentation__.Notify(456683)
	return r.f.GetCtxDone()
}

func (r *vectorizedFlowCreatorHelper) getCancelFlowFn() context.CancelFunc {
	__antithesis_instrumentation__.Notify(456684)
	return r.f.GetCancelFlowFn()
}

func (r *vectorizedFlowCreatorHelper) Release() {
	__antithesis_instrumentation__.Notify(456685)

	if len(r.processors) == 1 {
		__antithesis_instrumentation__.Notify(456687)
		r.processors[0] = nil
	} else {
		__antithesis_instrumentation__.Notify(456688)
	}
	__antithesis_instrumentation__.Notify(456686)
	*r = vectorizedFlowCreatorHelper{
		processors: r.processors[:0],
	}
	vectorizedFlowCreatorHelperPool.Put(r)
}

type noopFlowCreatorHelper struct {
	inboundStreams map[execinfrapb.StreamID]struct{}
}

var _ flowCreatorHelper = &noopFlowCreatorHelper{}

var noopFlowCreatorHelperPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(456689)
		return &noopFlowCreatorHelper{
			inboundStreams: make(map[execinfrapb.StreamID]struct{}),
		}
	},
}

func newNoopFlowCreatorHelper() *noopFlowCreatorHelper {
	__antithesis_instrumentation__.Notify(456690)
	return noopFlowCreatorHelperPool.Get().(*noopFlowCreatorHelper)
}

func (r *noopFlowCreatorHelper) addStreamEndpoint(
	streamID execinfrapb.StreamID, _ *colrpc.Inbox, _ *sync.WaitGroup,
) {
	__antithesis_instrumentation__.Notify(456691)
	r.inboundStreams[streamID] = struct{}{}
}

func (r *noopFlowCreatorHelper) checkInboundStreamID(sid execinfrapb.StreamID) error {
	__antithesis_instrumentation__.Notify(456692)
	if _, found := r.inboundStreams[sid]; found {
		__antithesis_instrumentation__.Notify(456694)
		return errors.Errorf("inbound stream %d already exists in map", sid)
	} else {
		__antithesis_instrumentation__.Notify(456695)
	}
	__antithesis_instrumentation__.Notify(456693)
	return nil
}

func (r *noopFlowCreatorHelper) accumulateAsyncComponent(runFn) {
	__antithesis_instrumentation__.Notify(456696)
}

func (r *noopFlowCreatorHelper) addFlowCoordinator(coordinator *FlowCoordinator) {
	__antithesis_instrumentation__.Notify(456697)
}

func (r *noopFlowCreatorHelper) getFlowCtxDone() <-chan struct{} {
	__antithesis_instrumentation__.Notify(456698)
	return nil
}

func (r *noopFlowCreatorHelper) getCancelFlowFn() context.CancelFunc {
	__antithesis_instrumentation__.Notify(456699)
	return nil
}

func (r *noopFlowCreatorHelper) Release() {
	__antithesis_instrumentation__.Notify(456700)
	for k := range r.inboundStreams {
		__antithesis_instrumentation__.Notify(456702)
		delete(r.inboundStreams, k)
	}
	__antithesis_instrumentation__.Notify(456701)
	noopFlowCreatorHelperPool.Put(r)
}

func IsSupported(mode sessiondatapb.VectorizeExecMode, spec *execinfrapb.FlowSpec) error {
	__antithesis_instrumentation__.Notify(456703)
	for pIdx := range spec.Processors {
		__antithesis_instrumentation__.Notify(456705)
		if err := colbuilder.IsSupported(mode, &spec.Processors[pIdx]); err != nil {
			__antithesis_instrumentation__.Notify(456707)
			return err
		} else {
			__antithesis_instrumentation__.Notify(456708)
		}
		__antithesis_instrumentation__.Notify(456706)
		for _, procOutput := range spec.Processors[pIdx].Output {
			__antithesis_instrumentation__.Notify(456709)
			switch procOutput.Type {
			case execinfrapb.OutputRouterSpec_PASS_THROUGH,
				execinfrapb.OutputRouterSpec_BY_HASH:
				__antithesis_instrumentation__.Notify(456710)
			default:
				__antithesis_instrumentation__.Notify(456711)
				return errors.New("only pass-through and hash routers are supported")
			}
		}
	}
	__antithesis_instrumentation__.Notify(456704)
	return nil
}
