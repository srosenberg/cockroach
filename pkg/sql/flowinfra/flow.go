package flowinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type flowStatus int

const (
	flowNotStarted flowStatus = iota
	flowRunning
	flowFinished
)

type Startable interface {
	Start(ctx context.Context, wg *sync.WaitGroup, flowCtxCancel context.CancelFunc)
}

type StartableFn func(context.Context, *sync.WaitGroup, context.CancelFunc)

func (f StartableFn) Start(
	ctx context.Context, wg *sync.WaitGroup, flowCtxCancel context.CancelFunc,
) {
	__antithesis_instrumentation__.Notify(491507)
	f(ctx, wg, flowCtxCancel)
}

type FuseOpt bool

const (
	FuseNormally FuseOpt = false

	FuseAggressively = true
)

type Flow interface {
	Setup(ctx context.Context, spec *execinfrapb.FlowSpec, opt FuseOpt) (context.Context, execinfra.OpChains, error)

	SetTxn(*kv.Txn)

	Start(_ context.Context, doneFn func()) error

	Run(_ context.Context, doneFn func())

	Wait()

	IsLocal() bool

	HasInboundStreams() bool

	IsVectorized() bool

	StatementSQL() string

	GetFlowCtx() *execinfra.FlowCtx

	AddStartable(Startable)

	GetID() execinfrapb.FlowID

	Cleanup(context.Context)

	ConcurrentTxnUse() bool
}

type FlowBase struct {
	execinfra.FlowCtx

	flowRegistry *FlowRegistry

	processors []execinfra.Processor

	startables []Startable

	rowSyncFlowConsumer execinfra.RowReceiver

	batchSyncFlowConsumer execinfra.BatchReceiver

	localProcessors []execinfra.LocalProcessor

	startedGoroutines bool

	inboundStreams map[execinfrapb.StreamID]*InboundStreamInfo

	waitGroup sync.WaitGroup

	onFlowCleanup func()

	statementSQL string

	doneFn func()

	status flowStatus

	ctxCancel context.CancelFunc
	ctxDone   <-chan struct{}

	sp *tracing.Span

	spec *execinfrapb.FlowSpec

	admissionInfo admission.WorkInfo
}

func (f *FlowBase) Setup(
	ctx context.Context, spec *execinfrapb.FlowSpec, _ FuseOpt,
) (context.Context, execinfra.OpChains, error) {
	__antithesis_instrumentation__.Notify(491508)
	ctx, f.ctxCancel = contextutil.WithCancel(ctx)
	f.ctxDone = ctx.Done()
	f.spec = spec
	return ctx, nil, nil
}

func (f *FlowBase) SetTxn(txn *kv.Txn) {
	__antithesis_instrumentation__.Notify(491509)
	f.FlowCtx.Txn = txn
	f.EvalCtx.Txn = txn
}

func (f *FlowBase) ConcurrentTxnUse() bool {
	__antithesis_instrumentation__.Notify(491510)
	numProcessorsThatMightUseTxn := 0
	for _, proc := range f.processors {
		__antithesis_instrumentation__.Notify(491512)
		if txnUser, ok := proc.(execinfra.DoesNotUseTxn); !ok || func() bool {
			__antithesis_instrumentation__.Notify(491513)
			return !txnUser.DoesNotUseTxn() == true
		}() == true {
			__antithesis_instrumentation__.Notify(491514)
			numProcessorsThatMightUseTxn++
			if numProcessorsThatMightUseTxn > 1 {
				__antithesis_instrumentation__.Notify(491515)
				return true
			} else {
				__antithesis_instrumentation__.Notify(491516)
			}
		} else {
			__antithesis_instrumentation__.Notify(491517)
		}
	}
	__antithesis_instrumentation__.Notify(491511)
	return false
}

func (f *FlowBase) SetStartedGoroutines(val bool) {
	__antithesis_instrumentation__.Notify(491518)
	f.startedGoroutines = val
}

func (f *FlowBase) Started() bool {
	__antithesis_instrumentation__.Notify(491519)
	return f.status != flowNotStarted
}

var _ Flow = &FlowBase{}

func NewFlowBase(
	flowCtx execinfra.FlowCtx,
	sp *tracing.Span,
	flowReg *FlowRegistry,
	rowSyncFlowConsumer execinfra.RowReceiver,
	batchSyncFlowConsumer execinfra.BatchReceiver,
	localProcessors []execinfra.LocalProcessor,
	onFlowCleanup func(),
	statementSQL string,
) *FlowBase {
	__antithesis_instrumentation__.Notify(491520)

	admissionInfo := admission.WorkInfo{TenantID: roachpb.SystemTenantID}
	if flowCtx.Txn == nil {
		__antithesis_instrumentation__.Notify(491522)
		admissionInfo.Priority = admission.NormalPri
		admissionInfo.CreateTime = timeutil.Now().UnixNano()
	} else {
		__antithesis_instrumentation__.Notify(491523)
		h := flowCtx.Txn.AdmissionHeader()
		admissionInfo.Priority = admission.WorkPriority(h.Priority)
		admissionInfo.CreateTime = h.CreateTime
	}
	__antithesis_instrumentation__.Notify(491521)
	return &FlowBase{
		FlowCtx:               flowCtx,
		sp:                    sp,
		flowRegistry:          flowReg,
		rowSyncFlowConsumer:   rowSyncFlowConsumer,
		batchSyncFlowConsumer: batchSyncFlowConsumer,
		localProcessors:       localProcessors,
		admissionInfo:         admissionInfo,
		onFlowCleanup:         onFlowCleanup,
		status:                flowNotStarted,
		statementSQL:          statementSQL,
	}
}

func (f *FlowBase) StatementSQL() string {
	__antithesis_instrumentation__.Notify(491524)
	return f.statementSQL
}

func (f *FlowBase) GetFlowCtx() *execinfra.FlowCtx {
	__antithesis_instrumentation__.Notify(491525)
	return &f.FlowCtx
}

func (f *FlowBase) AddStartable(s Startable) {
	__antithesis_instrumentation__.Notify(491526)
	f.startables = append(f.startables, s)
}

func (f *FlowBase) GetID() execinfrapb.FlowID {
	__antithesis_instrumentation__.Notify(491527)
	return f.ID
}

func (f *FlowBase) CheckInboundStreamID(sid execinfrapb.StreamID) error {
	__antithesis_instrumentation__.Notify(491528)
	if _, found := f.inboundStreams[sid]; found {
		__antithesis_instrumentation__.Notify(491531)
		return errors.Errorf("inbound stream %d already exists in map", sid)
	} else {
		__antithesis_instrumentation__.Notify(491532)
	}
	__antithesis_instrumentation__.Notify(491529)
	if f.inboundStreams == nil {
		__antithesis_instrumentation__.Notify(491533)
		f.inboundStreams = make(map[execinfrapb.StreamID]*InboundStreamInfo)
	} else {
		__antithesis_instrumentation__.Notify(491534)
	}
	__antithesis_instrumentation__.Notify(491530)
	return nil
}

func (f *FlowBase) GetWaitGroup() *sync.WaitGroup {
	__antithesis_instrumentation__.Notify(491535)
	return &f.waitGroup
}

func (f *FlowBase) GetCtxDone() <-chan struct{} {
	__antithesis_instrumentation__.Notify(491536)
	return f.ctxDone
}

func (f *FlowBase) GetCancelFlowFn() context.CancelFunc {
	__antithesis_instrumentation__.Notify(491537)
	return f.ctxCancel
}

func (f *FlowBase) SetProcessors(processors []execinfra.Processor) {
	__antithesis_instrumentation__.Notify(491538)
	f.processors = processors
}

func (f *FlowBase) AddRemoteStream(streamID execinfrapb.StreamID, streamInfo *InboundStreamInfo) {
	__antithesis_instrumentation__.Notify(491539)
	f.inboundStreams[streamID] = streamInfo
}

func (f *FlowBase) GetRowSyncFlowConsumer() execinfra.RowReceiver {
	__antithesis_instrumentation__.Notify(491540)
	return f.rowSyncFlowConsumer
}

func (f *FlowBase) GetBatchSyncFlowConsumer() execinfra.BatchReceiver {
	__antithesis_instrumentation__.Notify(491541)
	return f.batchSyncFlowConsumer
}

func (f *FlowBase) GetLocalProcessors() []execinfra.LocalProcessor {
	__antithesis_instrumentation__.Notify(491542)
	return f.localProcessors
}

func (f *FlowBase) GetAdmissionInfo() admission.WorkInfo {
	__antithesis_instrumentation__.Notify(491543)
	return f.admissionInfo
}

func (f *FlowBase) StartInternal(
	ctx context.Context, processors []execinfra.Processor, doneFn func(),
) error {
	__antithesis_instrumentation__.Notify(491544)
	f.doneFn = doneFn
	log.VEventf(
		ctx, 1, "starting (%d processors, %d startables) asynchronously", len(processors), len(f.startables),
	)

	if f.HasInboundStreams() {
		__antithesis_instrumentation__.Notify(491549)

		f.waitGroup.Add(len(f.inboundStreams))

		if err := f.flowRegistry.RegisterFlow(
			ctx, f.ID, f, f.inboundStreams, SettingFlowStreamTimeout.Get(&f.FlowCtx.Cfg.Settings.SV),
		); err != nil {
			__antithesis_instrumentation__.Notify(491550)
			return err
		} else {
			__antithesis_instrumentation__.Notify(491551)
		}
	} else {
		__antithesis_instrumentation__.Notify(491552)
	}
	__antithesis_instrumentation__.Notify(491545)

	f.status = flowRunning

	if log.V(1) {
		__antithesis_instrumentation__.Notify(491553)
		log.Infof(ctx, "registered flow %s", f.ID.Short())
	} else {
		__antithesis_instrumentation__.Notify(491554)
	}
	__antithesis_instrumentation__.Notify(491546)
	for _, s := range f.startables {
		__antithesis_instrumentation__.Notify(491555)
		s.Start(ctx, &f.waitGroup, f.ctxCancel)
	}
	__antithesis_instrumentation__.Notify(491547)
	for i := 0; i < len(processors); i++ {
		__antithesis_instrumentation__.Notify(491556)
		f.waitGroup.Add(1)
		go func(i int) {
			__antithesis_instrumentation__.Notify(491557)
			processors[i].Run(ctx)
			f.waitGroup.Done()
		}(i)
	}
	__antithesis_instrumentation__.Notify(491548)

	f.startedGoroutines = f.startedGoroutines || func() bool {
		__antithesis_instrumentation__.Notify(491558)
		return len(f.startables) > 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(491559)
		return len(processors) > 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(491560)
		return f.HasInboundStreams() == true
	}() == true
	return nil
}

func (f *FlowBase) IsLocal() bool {
	__antithesis_instrumentation__.Notify(491561)
	return f.Local
}

func (f *FlowBase) HasInboundStreams() bool {
	__antithesis_instrumentation__.Notify(491562)
	return len(f.inboundStreams) != 0
}

func (f *FlowBase) IsVectorized() bool {
	__antithesis_instrumentation__.Notify(491563)
	panic("IsVectorized should not be called on FlowBase")
}

func (f *FlowBase) Start(ctx context.Context, doneFn func()) error {
	__antithesis_instrumentation__.Notify(491564)
	return f.StartInternal(ctx, f.processors, doneFn)
}

func (f *FlowBase) Run(ctx context.Context, doneFn func()) {
	__antithesis_instrumentation__.Notify(491565)
	defer f.Wait()

	var headProc execinfra.Processor
	if len(f.processors) == 0 {
		__antithesis_instrumentation__.Notify(491568)
		f.rowSyncFlowConsumer.Push(nil, &execinfrapb.ProducerMetadata{Err: errors.AssertionFailedf("no processors in flow")})
		f.rowSyncFlowConsumer.ProducerDone()
		return
	} else {
		__antithesis_instrumentation__.Notify(491569)
	}
	__antithesis_instrumentation__.Notify(491566)
	headProc = f.processors[len(f.processors)-1]
	otherProcs := f.processors[:len(f.processors)-1]

	var err error
	if err = f.StartInternal(ctx, otherProcs, doneFn); err != nil {
		__antithesis_instrumentation__.Notify(491570)
		f.rowSyncFlowConsumer.Push(nil, &execinfrapb.ProducerMetadata{Err: err})
		f.rowSyncFlowConsumer.ProducerDone()
		return
	} else {
		__antithesis_instrumentation__.Notify(491571)
	}
	__antithesis_instrumentation__.Notify(491567)
	log.VEventf(ctx, 1, "running %T in the flow's goroutine", headProc)
	headProc.Run(ctx)
}

func (f *FlowBase) Wait() {
	__antithesis_instrumentation__.Notify(491572)
	if !f.startedGoroutines {
		__antithesis_instrumentation__.Notify(491577)
		return
	} else {
		__antithesis_instrumentation__.Notify(491578)
	}
	__antithesis_instrumentation__.Notify(491573)

	var panicVal interface{}
	if panicVal = recover(); panicVal != nil {
		__antithesis_instrumentation__.Notify(491579)

		f.ctxCancel()
	} else {
		__antithesis_instrumentation__.Notify(491580)
	}
	__antithesis_instrumentation__.Notify(491574)
	waitChan := make(chan struct{})

	go func() {
		__antithesis_instrumentation__.Notify(491581)
		f.waitGroup.Wait()
		close(waitChan)
	}()
	__antithesis_instrumentation__.Notify(491575)

	select {
	case <-f.ctxDone:
		__antithesis_instrumentation__.Notify(491582)
		f.cancel()
		<-waitChan
	case <-waitChan:
		__antithesis_instrumentation__.Notify(491583)

	}
	__antithesis_instrumentation__.Notify(491576)
	if panicVal != nil {
		__antithesis_instrumentation__.Notify(491584)
		panic(panicVal)
	} else {
		__antithesis_instrumentation__.Notify(491585)
	}
}

func (f *FlowBase) Cleanup(ctx context.Context) {
	__antithesis_instrumentation__.Notify(491586)
	if f.status == flowFinished {
		__antithesis_instrumentation__.Notify(491594)
		panic("flow cleanup called twice")
	} else {
		__antithesis_instrumentation__.Notify(491595)
	}
	__antithesis_instrumentation__.Notify(491587)

	if f.Descriptors != nil && func() bool {
		__antithesis_instrumentation__.Notify(491596)
		return f.IsDescriptorsCleanupRequired == true
	}() == true {
		__antithesis_instrumentation__.Notify(491597)
		f.Descriptors.ReleaseAll(ctx)
	} else {
		__antithesis_instrumentation__.Notify(491598)
	}
	__antithesis_instrumentation__.Notify(491588)

	if f.sp != nil {
		__antithesis_instrumentation__.Notify(491599)
		defer f.sp.Finish()
		if f.Gateway && func() bool {
			__antithesis_instrumentation__.Notify(491600)
			return f.CollectStats == true
		}() == true {
			__antithesis_instrumentation__.Notify(491601)

			f.sp.RecordStructured(&execinfrapb.ComponentStats{
				Component: execinfrapb.FlowComponentID(f.NodeID.SQLInstanceID(), f.FlowCtx.ID),
				FlowStats: execinfrapb.FlowStats{
					MaxMemUsage:  optional.MakeUint(uint64(f.FlowCtx.EvalCtx.Mon.MaximumBytes())),
					MaxDiskUsage: optional.MakeUint(uint64(f.FlowCtx.DiskMonitor.MaximumBytes())),
				},
			})
		} else {
			__antithesis_instrumentation__.Notify(491602)
		}
	} else {
		__antithesis_instrumentation__.Notify(491603)
	}
	__antithesis_instrumentation__.Notify(491589)

	f.DiskMonitor.Stop(ctx)

	f.EvalCtx.Stop(ctx)
	for _, p := range f.processors {
		__antithesis_instrumentation__.Notify(491604)
		if d, ok := p.(execinfra.Releasable); ok {
			__antithesis_instrumentation__.Notify(491605)
			d.Release()
		} else {
			__antithesis_instrumentation__.Notify(491606)
		}
	}
	__antithesis_instrumentation__.Notify(491590)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(491607)
		log.Infof(ctx, "cleaning up")
	} else {
		__antithesis_instrumentation__.Notify(491608)
	}
	__antithesis_instrumentation__.Notify(491591)
	if f.HasInboundStreams() && func() bool {
		__antithesis_instrumentation__.Notify(491609)
		return f.Started() == true
	}() == true {
		__antithesis_instrumentation__.Notify(491610)
		f.flowRegistry.UnregisterFlow(f.ID)
	} else {
		__antithesis_instrumentation__.Notify(491611)
	}
	__antithesis_instrumentation__.Notify(491592)
	f.status = flowFinished
	f.ctxCancel()
	if f.onFlowCleanup != nil {
		__antithesis_instrumentation__.Notify(491612)
		f.onFlowCleanup()
	} else {
		__antithesis_instrumentation__.Notify(491613)
	}
	__antithesis_instrumentation__.Notify(491593)
	if f.doneFn != nil {
		__antithesis_instrumentation__.Notify(491614)
		f.doneFn()
	} else {
		__antithesis_instrumentation__.Notify(491615)
	}
}

func (f *FlowBase) cancel() {
	__antithesis_instrumentation__.Notify(491616)
	if !f.HasInboundStreams() {
		__antithesis_instrumentation__.Notify(491618)
		return
	} else {
		__antithesis_instrumentation__.Notify(491619)
	}
	__antithesis_instrumentation__.Notify(491617)

	f.flowRegistry.cancelPendingStreams(f.ID, cancelchecker.QueryCanceledError)
}
