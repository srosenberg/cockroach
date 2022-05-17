package colflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecargs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexechash"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecutils"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/marusama/semaphore"
)

type routerOutput interface {
	execinfra.OpNode

	initWithHashRouter(*HashRouter)

	addBatch(context.Context, coldata.Batch) bool

	cancel(context.Context, error)

	forwardErr(error)

	resetForTests(context.Context)
}

func getDefaultRouterOutputBlockedThreshold() int {
	__antithesis_instrumentation__.Notify(456180)
	return coldata.BatchSize() * 2
}

type routerOutputOpState int

const (
	routerOutputOpRunning routerOutputOpState = iota

	routerOutputDoneAdding

	routerOutputOpDraining
)

type drainCoordinator interface {
	encounteredError(context.Context)

	drainMeta() []execinfrapb.ProducerMetadata
}

type routerOutputOp struct {
	colexecop.InitHelper

	input execinfra.OpNode

	drainCoordinator drainCoordinator

	types []*types.T

	unblockedEventsChan chan<- struct{}

	mu struct {
		syncutil.Mutex
		state routerOutputOpState

		forwardedErr error
		cond         *sync.Cond

		data      *colexecutils.SpillingQueue
		numUnread int
		blocked   bool
	}

	testingKnobs routerOutputOpTestingKnobs
}

func (o *routerOutputOp) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(456181)
	return 1
}

func (o *routerOutputOp) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(456182)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(456184)
		return o.input
	} else {
		__antithesis_instrumentation__.Notify(456185)
	}
	__antithesis_instrumentation__.Notify(456183)
	colexecerror.InternalError(errors.AssertionFailedf("invalid index %d", nth))

	return nil
}

var _ colexecop.Operator = &routerOutputOp{}

type routerOutputOpTestingKnobs struct {
	blockedThreshold int

	addBatchTestInducedErrorCb func() error

	nextTestInducedErrorCb func() error
}

type routerOutputOpArgs struct {
	types []*types.T

	unlimitedAllocator *colmem.Allocator

	memoryLimit int64
	diskAcc     *mon.BoundAccount
	cfg         colcontainer.DiskQueueCfg
	fdSemaphore semaphore.Semaphore

	unblockedEventsChan chan<- struct{}

	testingKnobs routerOutputOpTestingKnobs
}

func newRouterOutputOp(args routerOutputOpArgs) *routerOutputOp {
	__antithesis_instrumentation__.Notify(456186)
	if args.testingKnobs.blockedThreshold == 0 {
		__antithesis_instrumentation__.Notify(456188)
		args.testingKnobs.blockedThreshold = getDefaultRouterOutputBlockedThreshold()
	} else {
		__antithesis_instrumentation__.Notify(456189)
	}
	__antithesis_instrumentation__.Notify(456187)

	o := &routerOutputOp{
		types:               args.types,
		unblockedEventsChan: args.unblockedEventsChan,
		testingKnobs:        args.testingKnobs,
	}
	o.mu.cond = sync.NewCond(&o.mu)
	args.cfg.SetCacheMode(colcontainer.DiskQueueCacheModeIntertwinedCalls)
	o.mu.data = colexecutils.NewSpillingQueue(
		&colexecutils.NewSpillingQueueArgs{
			UnlimitedAllocator: args.unlimitedAllocator,
			Types:              args.types,
			MemoryLimit:        args.memoryLimit,
			DiskQueueCfg:       args.cfg,
			FDSemaphore:        args.fdSemaphore,
			DiskAcc:            args.diskAcc,
		},
	)

	return o
}

func (o *routerOutputOp) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456190)
	o.InitHelper.Init(ctx)
}

func (o *routerOutputOp) nextErrorLocked(err error) {
	__antithesis_instrumentation__.Notify(456191)
	o.mu.state = routerOutputOpDraining
	o.maybeUnblockLocked()

	o.mu.Unlock()
	o.drainCoordinator.encounteredError(o.Ctx)
	o.mu.Lock()
	colexecerror.InternalError(err)
}

func (o *routerOutputOp) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(456192)
	o.mu.Lock()
	defer o.mu.Unlock()
	for o.mu.forwardedErr == nil && func() bool {
		__antithesis_instrumentation__.Notify(456200)
		return o.mu.state == routerOutputOpRunning == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(456201)
		return o.mu.data.Empty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(456202)

		o.mu.cond.Wait()
	}
	__antithesis_instrumentation__.Notify(456193)
	if o.mu.forwardedErr != nil {
		__antithesis_instrumentation__.Notify(456203)
		colexecerror.ExpectedError(o.mu.forwardedErr)
	} else {
		__antithesis_instrumentation__.Notify(456204)
	}
	__antithesis_instrumentation__.Notify(456194)
	if o.mu.state == routerOutputOpDraining {
		__antithesis_instrumentation__.Notify(456205)
		return coldata.ZeroBatch
	} else {
		__antithesis_instrumentation__.Notify(456206)
	}
	__antithesis_instrumentation__.Notify(456195)
	b, err := o.mu.data.Dequeue(o.Ctx)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(456207)
		return o.testingKnobs.nextTestInducedErrorCb != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(456208)
		err = o.testingKnobs.nextTestInducedErrorCb()
	} else {
		__antithesis_instrumentation__.Notify(456209)
	}
	__antithesis_instrumentation__.Notify(456196)
	if err != nil {
		__antithesis_instrumentation__.Notify(456210)
		o.nextErrorLocked(err)
	} else {
		__antithesis_instrumentation__.Notify(456211)
	}
	__antithesis_instrumentation__.Notify(456197)
	o.mu.numUnread -= b.Length()
	if o.mu.numUnread <= o.testingKnobs.blockedThreshold {
		__antithesis_instrumentation__.Notify(456212)
		o.maybeUnblockLocked()
	} else {
		__antithesis_instrumentation__.Notify(456213)
	}
	__antithesis_instrumentation__.Notify(456198)
	if b.Length() == 0 {
		__antithesis_instrumentation__.Notify(456214)
		if o.testingKnobs.nextTestInducedErrorCb != nil {
			__antithesis_instrumentation__.Notify(456216)
			if err := o.testingKnobs.nextTestInducedErrorCb(); err != nil {
				__antithesis_instrumentation__.Notify(456217)
				o.nextErrorLocked(err)
			} else {
				__antithesis_instrumentation__.Notify(456218)
			}
		} else {
			__antithesis_instrumentation__.Notify(456219)
		}
		__antithesis_instrumentation__.Notify(456215)

		o.closeLocked(o.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(456220)
	}
	__antithesis_instrumentation__.Notify(456199)
	return b
}

func (o *routerOutputOp) DrainMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(456221)
	o.mu.Lock()
	o.mu.state = routerOutputOpDraining
	o.maybeUnblockLocked()
	o.mu.Unlock()
	return o.drainCoordinator.drainMeta()
}

func (o *routerOutputOp) initWithHashRouter(r *HashRouter) {
	__antithesis_instrumentation__.Notify(456222)
	o.input = r
	o.drainCoordinator = r
}

func (o *routerOutputOp) closeLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456223)
	o.mu.state = routerOutputOpDraining
	if err := o.mu.data.Close(ctx); err != nil {
		__antithesis_instrumentation__.Notify(456224)

		log.Infof(ctx, "error closing vectorized hash router output, files may be left over: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(456225)
	}
}

func (o *routerOutputOp) cancel(ctx context.Context, err error) {
	__antithesis_instrumentation__.Notify(456226)
	o.mu.Lock()
	defer o.mu.Unlock()
	o.closeLocked(ctx)
	o.forwardErrLocked(err)

	o.mu.cond.Signal()
}

func (o *routerOutputOp) forwardErrLocked(err error) {
	__antithesis_instrumentation__.Notify(456227)
	if err != nil {
		__antithesis_instrumentation__.Notify(456228)
		o.mu.forwardedErr = err
	} else {
		__antithesis_instrumentation__.Notify(456229)
	}
}

func (o *routerOutputOp) forwardErr(err error) {
	__antithesis_instrumentation__.Notify(456230)
	o.mu.Lock()
	defer o.mu.Unlock()
	o.forwardErrLocked(err)
	o.mu.cond.Signal()
}

func (o *routerOutputOp) addBatch(ctx context.Context, batch coldata.Batch) bool {
	__antithesis_instrumentation__.Notify(456231)
	o.mu.Lock()
	defer o.mu.Unlock()
	switch o.mu.state {
	case routerOutputDoneAdding:
		__antithesis_instrumentation__.Notify(456236)
		colexecerror.InternalError(errors.AssertionFailedf("a batch was added to routerOutput in DoneAdding state"))
	case routerOutputOpDraining:
		__antithesis_instrumentation__.Notify(456237)

		return false
	default:
		__antithesis_instrumentation__.Notify(456238)
	}
	__antithesis_instrumentation__.Notify(456232)

	o.mu.numUnread += batch.Length()
	o.mu.data.Enqueue(ctx, batch)
	if o.testingKnobs.addBatchTestInducedErrorCb != nil {
		__antithesis_instrumentation__.Notify(456239)
		if err := o.testingKnobs.addBatchTestInducedErrorCb(); err != nil {
			__antithesis_instrumentation__.Notify(456240)
			colexecerror.InternalError(err)
		} else {
			__antithesis_instrumentation__.Notify(456241)
		}
	} else {
		__antithesis_instrumentation__.Notify(456242)
	}
	__antithesis_instrumentation__.Notify(456233)

	if batch.Length() == 0 {
		__antithesis_instrumentation__.Notify(456243)
		o.mu.state = routerOutputDoneAdding
		o.mu.cond.Signal()
		return false
	} else {
		__antithesis_instrumentation__.Notify(456244)
	}
	__antithesis_instrumentation__.Notify(456234)

	stateChanged := false
	if o.mu.numUnread > o.testingKnobs.blockedThreshold && func() bool {
		__antithesis_instrumentation__.Notify(456245)
		return !o.mu.blocked == true
	}() == true {
		__antithesis_instrumentation__.Notify(456246)

		o.mu.blocked = true
		stateChanged = true
	} else {
		__antithesis_instrumentation__.Notify(456247)
	}
	__antithesis_instrumentation__.Notify(456235)
	o.mu.cond.Signal()
	return stateChanged
}

func (o *routerOutputOp) maybeUnblockLocked() {
	__antithesis_instrumentation__.Notify(456248)
	if o.mu.blocked {
		__antithesis_instrumentation__.Notify(456249)
		o.mu.blocked = false
		o.unblockedEventsChan <- struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(456250)
	}
}

func (o *routerOutputOp) resetForTests(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456251)
	o.mu.Lock()
	defer o.mu.Unlock()
	o.mu.state = routerOutputOpRunning
	o.mu.forwardedErr = nil
	o.mu.data.Reset(ctx)
	o.mu.numUnread = 0
	o.mu.blocked = false
}

type hashRouterDrainState int

const (
	hashRouterDrainStateRunning = iota

	hashRouterDrainStateRequested

	hashRouterDrainStateCompleted
)

type HashRouter struct {
	colexecop.OneInputNode

	inputMetaInfo colexecargs.OpWithMetaInfo

	hashCols []uint32

	outputs []routerOutput

	unblockedEventsChan <-chan struct{}
	numBlockedOutputs   int

	bufferedMeta []execinfrapb.ProducerMetadata

	atomics struct {
		drainState        int32
		numDrainedOutputs int32
	}

	waitForMetadata chan []execinfrapb.ProducerMetadata

	tupleDistributor *colexechash.TupleHashDistributor
}

func NewHashRouter(
	unlimitedAllocators []*colmem.Allocator,
	input colexecargs.OpWithMetaInfo,
	types []*types.T,
	hashCols []uint32,
	memoryLimit int64,
	diskQueueCfg colcontainer.DiskQueueCfg,
	fdSemaphore semaphore.Semaphore,
	diskAccounts []*mon.BoundAccount,
) (*HashRouter, []colexecop.DrainableOperator) {
	__antithesis_instrumentation__.Notify(456252)
	outputs := make([]routerOutput, len(unlimitedAllocators))
	outputsAsOps := make([]colexecop.DrainableOperator, len(unlimitedAllocators))

	unblockEventsChan := make(chan struct{}, 2*len(unlimitedAllocators))
	memoryLimitPerOutput := memoryLimit / int64(len(unlimitedAllocators))
	if memoryLimit == 1 {
		__antithesis_instrumentation__.Notify(456255)

		memoryLimitPerOutput = 1
	} else {
		__antithesis_instrumentation__.Notify(456256)
	}
	__antithesis_instrumentation__.Notify(456253)
	for i := range unlimitedAllocators {
		__antithesis_instrumentation__.Notify(456257)
		op := newRouterOutputOp(
			routerOutputOpArgs{
				types:               types,
				unlimitedAllocator:  unlimitedAllocators[i],
				memoryLimit:         memoryLimitPerOutput,
				diskAcc:             diskAccounts[i],
				cfg:                 diskQueueCfg,
				fdSemaphore:         fdSemaphore,
				unblockedEventsChan: unblockEventsChan,
			},
		)
		outputs[i] = op
		outputsAsOps[i] = op
	}
	__antithesis_instrumentation__.Notify(456254)
	return newHashRouterWithOutputs(input, hashCols, unblockEventsChan, outputs), outputsAsOps
}

func newHashRouterWithOutputs(
	input colexecargs.OpWithMetaInfo,
	hashCols []uint32,
	unblockEventsChan <-chan struct{},
	outputs []routerOutput,
) *HashRouter {
	__antithesis_instrumentation__.Notify(456258)
	r := &HashRouter{
		OneInputNode:        colexecop.NewOneInputNode(input.Root),
		inputMetaInfo:       input,
		hashCols:            hashCols,
		outputs:             outputs,
		unblockedEventsChan: unblockEventsChan,

		waitForMetadata:  make(chan []execinfrapb.ProducerMetadata, 1),
		tupleDistributor: colexechash.NewTupleHashDistributor(colexechash.DefaultInitHashValue, len(outputs)),
	}
	for i := range outputs {
		__antithesis_instrumentation__.Notify(456260)
		outputs[i].initWithHashRouter(r)
	}
	__antithesis_instrumentation__.Notify(456259)
	return r
}

func (r *HashRouter) cancelOutputs(ctx context.Context, errToForward error) {
	__antithesis_instrumentation__.Notify(456261)
	for _, o := range r.outputs {
		__antithesis_instrumentation__.Notify(456262)
		if err := colexecerror.CatchVectorizedRuntimeError(func() {
			__antithesis_instrumentation__.Notify(456263)
			o.cancel(ctx, errToForward)
		}); err != nil {
			__antithesis_instrumentation__.Notify(456264)

			o.forwardErr(err)
		} else {
			__antithesis_instrumentation__.Notify(456265)
		}
	}
}

func (r *HashRouter) setDrainState(drainState hashRouterDrainState) {
	__antithesis_instrumentation__.Notify(456266)
	atomic.StoreInt32(&r.atomics.drainState, int32(drainState))
}

func (r *HashRouter) getDrainState() hashRouterDrainState {
	__antithesis_instrumentation__.Notify(456267)
	return hashRouterDrainState(atomic.LoadInt32(&r.atomics.drainState))
}

func (r *HashRouter) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456268)
	var span *tracing.Span
	ctx, span = execinfra.ProcessorSpan(ctx, "hash router")
	if span != nil {
		__antithesis_instrumentation__.Notify(456272)
		defer span.Finish()
	} else {
		__antithesis_instrumentation__.Notify(456273)
	}
	__antithesis_instrumentation__.Notify(456269)
	var inputInitialized bool

	if err := colexecerror.CatchVectorizedRuntimeError(func() {
		__antithesis_instrumentation__.Notify(456274)
		r.Input.Init(ctx)
		inputInitialized = true
		var done bool
		processNextBatch := func() {
			__antithesis_instrumentation__.Notify(456276)
			done = r.processNextBatch(ctx)
		}
		__antithesis_instrumentation__.Notify(456275)
		for {
			__antithesis_instrumentation__.Notify(456277)
			if r.getDrainState() != hashRouterDrainStateRunning {
				__antithesis_instrumentation__.Notify(456283)
				break
			} else {
				__antithesis_instrumentation__.Notify(456284)
			}
			__antithesis_instrumentation__.Notify(456278)

			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(456285)
				r.cancelOutputs(ctx, ctx.Err())
				return
			default:
				__antithesis_instrumentation__.Notify(456286)
			}
			__antithesis_instrumentation__.Notify(456279)

			for moreToRead := true; moreToRead; {
				__antithesis_instrumentation__.Notify(456287)
				select {
				case <-r.unblockedEventsChan:
					__antithesis_instrumentation__.Notify(456288)
					r.numBlockedOutputs--
				default:
					__antithesis_instrumentation__.Notify(456289)

					moreToRead = false
				}
			}
			__antithesis_instrumentation__.Notify(456280)

			if r.numBlockedOutputs == len(r.outputs) {
				__antithesis_instrumentation__.Notify(456290)

				select {
				case <-r.unblockedEventsChan:
					__antithesis_instrumentation__.Notify(456291)
					r.numBlockedOutputs--
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(456292)
					r.cancelOutputs(ctx, ctx.Err())
					return
				}
			} else {
				__antithesis_instrumentation__.Notify(456293)
			}
			__antithesis_instrumentation__.Notify(456281)

			if err := colexecerror.CatchVectorizedRuntimeError(processNextBatch); err != nil {
				__antithesis_instrumentation__.Notify(456294)
				r.cancelOutputs(ctx, err)
				return
			} else {
				__antithesis_instrumentation__.Notify(456295)
			}
			__antithesis_instrumentation__.Notify(456282)
			if done {
				__antithesis_instrumentation__.Notify(456296)

				return
			} else {
				__antithesis_instrumentation__.Notify(456297)
			}
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(456298)
		r.cancelOutputs(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(456299)
	}
	__antithesis_instrumentation__.Notify(456270)
	if inputInitialized {
		__antithesis_instrumentation__.Notify(456300)

		if span != nil {
			__antithesis_instrumentation__.Notify(456302)
			for _, s := range r.inputMetaInfo.StatsCollectors {
				__antithesis_instrumentation__.Notify(456304)
				span.RecordStructured(s.GetStats())
			}
			__antithesis_instrumentation__.Notify(456303)
			if meta := execinfra.GetTraceDataAsMetadata(span); meta != nil {
				__antithesis_instrumentation__.Notify(456305)
				r.bufferedMeta = append(r.bufferedMeta, *meta)
			} else {
				__antithesis_instrumentation__.Notify(456306)
			}
		} else {
			__antithesis_instrumentation__.Notify(456307)
		}
		__antithesis_instrumentation__.Notify(456301)
		r.bufferedMeta = append(r.bufferedMeta, r.inputMetaInfo.MetadataSources.DrainMeta()...)
	} else {
		__antithesis_instrumentation__.Notify(456308)
	}
	__antithesis_instrumentation__.Notify(456271)

	r.waitForMetadata <- r.bufferedMeta
	close(r.waitForMetadata)

	r.inputMetaInfo.ToClose.CloseAndLogOnErr(ctx, "hash router")
}

func (r *HashRouter) processNextBatch(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(456309)
	b := r.Input.Next()
	n := b.Length()
	if n == 0 {
		__antithesis_instrumentation__.Notify(456312)

		for _, o := range r.outputs {
			__antithesis_instrumentation__.Notify(456314)
			o.addBatch(ctx, b)
		}
		__antithesis_instrumentation__.Notify(456313)
		return true
	} else {
		__antithesis_instrumentation__.Notify(456315)
	}
	__antithesis_instrumentation__.Notify(456310)

	r.tupleDistributor.Init(ctx)
	selections := r.tupleDistributor.Distribute(b, r.hashCols)
	for i, o := range r.outputs {
		__antithesis_instrumentation__.Notify(456316)
		if len(selections[i]) > 0 {
			__antithesis_instrumentation__.Notify(456317)
			colexecutils.UpdateBatchState(b, len(selections[i]), true, selections[i])
			if o.addBatch(ctx, b) {
				__antithesis_instrumentation__.Notify(456318)

				r.numBlockedOutputs++
			} else {
				__antithesis_instrumentation__.Notify(456319)
			}
		} else {
			__antithesis_instrumentation__.Notify(456320)
		}
	}
	__antithesis_instrumentation__.Notify(456311)
	return false
}

func (r *HashRouter) resetForTests(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456321)
	if i, ok := r.Input.(colexecop.Resetter); ok {
		__antithesis_instrumentation__.Notify(456324)
		i.Reset(ctx)
	} else {
		__antithesis_instrumentation__.Notify(456325)
	}
	__antithesis_instrumentation__.Notify(456322)
	r.setDrainState(hashRouterDrainStateRunning)
	r.waitForMetadata = make(chan []execinfrapb.ProducerMetadata, 1)
	r.atomics.numDrainedOutputs = 0
	r.bufferedMeta = nil
	r.numBlockedOutputs = 0
	for moreToRead := true; moreToRead; {
		__antithesis_instrumentation__.Notify(456326)
		select {
		case <-r.unblockedEventsChan:
			__antithesis_instrumentation__.Notify(456327)
		default:
			__antithesis_instrumentation__.Notify(456328)
			moreToRead = false
		}
	}
	__antithesis_instrumentation__.Notify(456323)
	for _, o := range r.outputs {
		__antithesis_instrumentation__.Notify(456329)
		o.resetForTests(ctx)
	}
}

func (r *HashRouter) encounteredError(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456330)

	r.setDrainState(hashRouterDrainStateRequested)

	r.cancelOutputs(ctx, nil)
}

func (r *HashRouter) drainMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(456331)
	if int(atomic.AddInt32(&r.atomics.numDrainedOutputs, 1)) != len(r.outputs) {
		__antithesis_instrumentation__.Notify(456333)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(456334)
	}
	__antithesis_instrumentation__.Notify(456332)

	r.setDrainState(hashRouterDrainStateRequested)
	meta := <-r.waitForMetadata
	r.setDrainState(hashRouterDrainStateCompleted)
	return meta
}
