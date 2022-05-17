package colflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecargs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colflow/colrpc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type childStatsCollector interface {
	getElapsedTime() time.Duration
}

type batchInfoCollector struct {
	colexecop.Operator
	colexecop.NonExplainable
	componentID execinfrapb.ComponentID

	mu struct {
		syncutil.Mutex
		initialized           bool
		numBatches, numTuples uint64
	}

	batch coldata.Batch

	stopwatch *timeutil.StopWatch

	childStatsCollectors []childStatsCollector
}

var _ colexecop.Operator = &batchInfoCollector{}

func makeBatchInfoCollector(
	op colexecop.Operator,
	id execinfrapb.ComponentID,
	inputWatch *timeutil.StopWatch,
	childStatsCollectors []childStatsCollector,
) batchInfoCollector {
	__antithesis_instrumentation__.Notify(456335)
	if inputWatch == nil {
		__antithesis_instrumentation__.Notify(456337)
		colexecerror.InternalError(errors.AssertionFailedf("input watch is nil"))
	} else {
		__antithesis_instrumentation__.Notify(456338)
	}
	__antithesis_instrumentation__.Notify(456336)
	return batchInfoCollector{
		Operator:             op,
		componentID:          id,
		stopwatch:            inputWatch,
		childStatsCollectors: childStatsCollectors,
	}
}

func (bic *batchInfoCollector) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456339)
	bic.Operator.Init(ctx)
	bic.mu.Lock()

	bic.mu.initialized = true
	bic.mu.Unlock()
}

func (bic *batchInfoCollector) next() {
	__antithesis_instrumentation__.Notify(456340)
	bic.batch = bic.Operator.Next()
}

func (bic *batchInfoCollector) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(456341)
	bic.stopwatch.Start()

	err := colexecerror.CatchVectorizedRuntimeError(bic.next)
	bic.stopwatch.Stop()
	if err != nil {
		__antithesis_instrumentation__.Notify(456344)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(456345)
	}
	__antithesis_instrumentation__.Notify(456342)
	if bic.batch.Length() > 0 {
		__antithesis_instrumentation__.Notify(456346)
		bic.mu.Lock()
		bic.mu.numBatches++
		bic.mu.numTuples += uint64(bic.batch.Length())
		bic.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(456347)
	}
	__antithesis_instrumentation__.Notify(456343)
	return bic.batch
}

func (bic *batchInfoCollector) finishAndGetStats() (
	numBatches, numTuples uint64,
	time time.Duration,
	ok bool,
) {
	__antithesis_instrumentation__.Notify(456348)
	tm := bic.stopwatch.Elapsed()

	for _, statsCollectors := range bic.childStatsCollectors {
		__antithesis_instrumentation__.Notify(456350)
		tm -= statsCollectors.getElapsedTime()
	}
	__antithesis_instrumentation__.Notify(456349)
	bic.mu.Lock()
	defer bic.mu.Unlock()
	return bic.mu.numBatches, bic.mu.numTuples, tm, bic.mu.initialized
}

func (bic *batchInfoCollector) getElapsedTime() time.Duration {
	__antithesis_instrumentation__.Notify(456351)
	return bic.stopwatch.Elapsed()
}

func newVectorizedStatsCollector(
	op colexecop.Operator,
	kvReader colexecop.KVReader,
	columnarizer colexecop.VectorizedStatsCollector,
	id execinfrapb.ComponentID,
	inputWatch *timeutil.StopWatch,
	memMonitors []*mon.BytesMonitor,
	diskMonitors []*mon.BytesMonitor,
	inputStatsCollectors []childStatsCollector,
) colexecop.VectorizedStatsCollector {
	__antithesis_instrumentation__.Notify(456352)

	return &vectorizedStatsCollectorImpl{
		batchInfoCollector: makeBatchInfoCollector(op, id, inputWatch, inputStatsCollectors),
		kvReader:           kvReader,
		columnarizer:       columnarizer,
		memMonitors:        memMonitors,
		diskMonitors:       diskMonitors,
	}
}

type vectorizedStatsCollectorImpl struct {
	batchInfoCollector

	kvReader     colexecop.KVReader
	columnarizer colexecop.VectorizedStatsCollector
	memMonitors  []*mon.BytesMonitor
	diskMonitors []*mon.BytesMonitor
}

func (vsc *vectorizedStatsCollectorImpl) GetStats() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(456353)
	numBatches, numTuples, time, ok := vsc.batchInfoCollector.finishAndGetStats()
	if !ok {
		__antithesis_instrumentation__.Notify(456359)

		return &execinfrapb.ComponentStats{}
	} else {
		__antithesis_instrumentation__.Notify(456360)
	}
	__antithesis_instrumentation__.Notify(456354)

	var s *execinfrapb.ComponentStats
	if vsc.columnarizer != nil {
		__antithesis_instrumentation__.Notify(456361)
		s = vsc.columnarizer.GetStats()
	} else {
		__antithesis_instrumentation__.Notify(456362)

		s = &execinfrapb.ComponentStats{Component: vsc.componentID}
	}
	__antithesis_instrumentation__.Notify(456355)

	for _, memMon := range vsc.memMonitors {
		__antithesis_instrumentation__.Notify(456363)
		s.Exec.MaxAllocatedMem.Add(memMon.MaximumBytes())
	}
	__antithesis_instrumentation__.Notify(456356)
	for _, diskMon := range vsc.diskMonitors {
		__antithesis_instrumentation__.Notify(456364)
		s.Exec.MaxAllocatedDisk.Add(diskMon.MaximumBytes())
	}
	__antithesis_instrumentation__.Notify(456357)

	if vsc.kvReader != nil {
		__antithesis_instrumentation__.Notify(456365)

		s.KV.KVTime.Set(time)
		s.KV.TuplesRead.Set(uint64(vsc.kvReader.GetRowsRead()))
		s.KV.BytesRead.Set(uint64(vsc.kvReader.GetBytesRead()))
		s.KV.ContentionTime.Set(vsc.kvReader.GetCumulativeContentionTime())
		scanStats := vsc.kvReader.GetScanStats()
		execinfra.PopulateKVMVCCStats(&s.KV, &scanStats)
	} else {
		__antithesis_instrumentation__.Notify(456366)
		s.Exec.ExecTime.Set(time)
	}
	__antithesis_instrumentation__.Notify(456358)

	s.Output.NumBatches.Set(numBatches)
	s.Output.NumTuples.Set(numTuples)
	return s
}

func newNetworkVectorizedStatsCollector(
	op colexecop.Operator,
	id execinfrapb.ComponentID,
	inputWatch *timeutil.StopWatch,
	inbox *colrpc.Inbox,
	latency time.Duration,
) colexecop.VectorizedStatsCollector {
	__antithesis_instrumentation__.Notify(456367)
	return &networkVectorizedStatsCollectorImpl{
		batchInfoCollector: makeBatchInfoCollector(op, id, inputWatch, nil),
		inbox:              inbox,
		latency:            latency,
	}
}

type networkVectorizedStatsCollectorImpl struct {
	batchInfoCollector

	inbox   *colrpc.Inbox
	latency time.Duration
}

func (nvsc *networkVectorizedStatsCollectorImpl) GetStats() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(456368)
	numBatches, numTuples, time, ok := nvsc.batchInfoCollector.finishAndGetStats()
	if !ok {
		__antithesis_instrumentation__.Notify(456370)

		return &execinfrapb.ComponentStats{}
	} else {
		__antithesis_instrumentation__.Notify(456371)
	}
	__antithesis_instrumentation__.Notify(456369)

	s := &execinfrapb.ComponentStats{Component: nvsc.componentID}

	s.NetRx.Latency.Set(nvsc.latency)
	s.NetRx.WaitTime.Set(time)
	s.NetRx.DeserializationTime.Set(nvsc.inbox.GetDeserializationTime())
	s.NetRx.TuplesReceived.Set(uint64(nvsc.inbox.GetRowsRead()))
	s.NetRx.BytesReceived.Set(uint64(nvsc.inbox.GetBytesRead()))
	s.NetRx.MessagesReceived.Set(uint64(nvsc.inbox.GetNumMessages()))

	s.Output.NumBatches.Set(numBatches)
	s.Output.NumTuples.Set(numTuples)

	return s
}

func maybeAddStatsInvariantChecker(op *colexecargs.OpWithMetaInfo) {
	__antithesis_instrumentation__.Notify(456372)
	if buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(456373)
		c := &statsInvariantChecker{}
		op.StatsCollectors = append(op.StatsCollectors, c)
		op.MetadataSources = append(op.MetadataSources, c)
	} else {
		__antithesis_instrumentation__.Notify(456374)
	}
}

type statsInvariantChecker struct {
	colexecop.ZeroInputNode

	statsRetrieved bool
}

var _ colexecop.VectorizedStatsCollector = &statsInvariantChecker{}
var _ colexecop.MetadataSource = &statsInvariantChecker{}

func (i *statsInvariantChecker) Init(context.Context) { __antithesis_instrumentation__.Notify(456375) }

func (i *statsInvariantChecker) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(456376)
	return coldata.ZeroBatch
}

func (i *statsInvariantChecker) GetStats() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(456377)
	i.statsRetrieved = true
	return &execinfrapb.ComponentStats{}
}

func (i *statsInvariantChecker) DrainMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(456378)
	if !i.statsRetrieved {
		__antithesis_instrumentation__.Notify(456380)
		return []execinfrapb.ProducerMetadata{{Err: errors.New("GetStats wasn't called before DrainMeta")}}
	} else {
		__antithesis_instrumentation__.Notify(456381)
	}
	__antithesis_instrumentation__.Notify(456379)
	return nil
}
