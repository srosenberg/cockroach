package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/colflow"
	"github.com/cockroachdb/cockroach/pkg/sql/contention"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/sql/distsql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/ring"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

const numRunners = 16

const clientRejectedMsg string = "client rejected when attempting to run DistSQL plan"

type runnerRequest struct {
	ctx           context.Context
	nodeDialer    *nodedialer.Dialer
	flowReq       *execinfrapb.SetupFlowRequest
	sqlInstanceID base.SQLInstanceID
	resultChan    chan<- runnerResult
}

type runnerResult struct {
	nodeID base.SQLInstanceID
	err    error
}

func (req runnerRequest) run() {
	__antithesis_instrumentation__.Notify(467865)
	res := runnerResult{nodeID: req.sqlInstanceID}

	conn, err := req.nodeDialer.Dial(req.ctx, roachpb.NodeID(req.sqlInstanceID), rpc.DefaultClass)
	if err != nil {
		__antithesis_instrumentation__.Notify(467867)
		res.err = err
	} else {
		__antithesis_instrumentation__.Notify(467868)
		client := execinfrapb.NewDistSQLClient(conn)

		if sp := tracing.SpanFromContext(req.ctx); sp != nil && func() bool {
			__antithesis_instrumentation__.Notify(467870)
			return !sp.IsNoop() == true
		}() == true {
			__antithesis_instrumentation__.Notify(467871)
			req.flowReq.TraceInfo = sp.Meta().ToProto()
		} else {
			__antithesis_instrumentation__.Notify(467872)
		}
		__antithesis_instrumentation__.Notify(467869)
		resp, err := client.SetupFlow(req.ctx, req.flowReq)
		if err != nil {
			__antithesis_instrumentation__.Notify(467873)
			res.err = err
		} else {
			__antithesis_instrumentation__.Notify(467874)
			res.err = resp.Error.ErrorDetail(req.ctx)
		}
	}
	__antithesis_instrumentation__.Notify(467866)
	req.resultChan <- res
}

func (dsp *DistSQLPlanner) initRunners(ctx context.Context) {
	__antithesis_instrumentation__.Notify(467875)

	dsp.runnerChan = make(chan runnerRequest)
	for i := 0; i < numRunners; i++ {
		__antithesis_instrumentation__.Notify(467876)
		_ = dsp.stopper.RunAsyncTask(ctx, "distsql-runner", func(context.Context) {
			__antithesis_instrumentation__.Notify(467877)
			runnerChan := dsp.runnerChan
			stopChan := dsp.stopper.ShouldQuiesce()
			for {
				__antithesis_instrumentation__.Notify(467878)
				select {
				case req := <-runnerChan:
					__antithesis_instrumentation__.Notify(467879)
					req.run()

				case <-stopChan:
					__antithesis_instrumentation__.Notify(467880)
					return
				}
			}
		})
	}
}

const numCancelingWorkers = numRunners / 4

func (dsp *DistSQLPlanner) initCancelingWorkers(initCtx context.Context) {
	__antithesis_instrumentation__.Notify(467881)
	dsp.cancelFlowsCoordinator.workerWait = make(chan struct{}, numCancelingWorkers)
	const cancelRequestTimeout = 10 * time.Second
	for i := 0; i < numCancelingWorkers; i++ {
		__antithesis_instrumentation__.Notify(467882)
		workerID := i + 1
		_ = dsp.stopper.RunAsyncTask(initCtx, "distsql-canceling-worker", func(parentCtx context.Context) {
			__antithesis_instrumentation__.Notify(467883)
			stopChan := dsp.stopper.ShouldQuiesce()
			for {
				__antithesis_instrumentation__.Notify(467884)
				select {
				case <-stopChan:
					__antithesis_instrumentation__.Notify(467885)
					return

				case <-dsp.cancelFlowsCoordinator.workerWait:
					__antithesis_instrumentation__.Notify(467886)
					req, sqlInstanceID := dsp.cancelFlowsCoordinator.getFlowsToCancel()
					if req == nil {
						__antithesis_instrumentation__.Notify(467889)

						log.VEventf(parentCtx, 2, "worker %d woke up but didn't find any flows to cancel", workerID)
						continue
					} else {
						__antithesis_instrumentation__.Notify(467890)
					}
					__antithesis_instrumentation__.Notify(467887)
					log.VEventf(parentCtx, 2, "worker %d is canceling at most %d flows on node %d", workerID, len(req.FlowIDs), sqlInstanceID)

					conn, err := dsp.podNodeDialer.Dial(parentCtx, roachpb.NodeID(sqlInstanceID), rpc.DefaultClass)
					if err != nil {
						__antithesis_instrumentation__.Notify(467891)

						continue
					} else {
						__antithesis_instrumentation__.Notify(467892)
					}
					__antithesis_instrumentation__.Notify(467888)
					client := execinfrapb.NewDistSQLClient(conn)
					_ = contextutil.RunWithTimeout(
						parentCtx,
						"cancel dead flows",
						cancelRequestTimeout,
						func(ctx context.Context) error {
							__antithesis_instrumentation__.Notify(467893)
							_, _ = client.CancelDeadFlows(ctx, req)
							return nil
						})
				}
			}
		})
	}
}

type deadFlowsOnNode struct {
	ids           []execinfrapb.FlowID
	sqlInstanceID base.SQLInstanceID
}

type cancelFlowsCoordinator struct {
	mu struct {
		syncutil.Mutex

		deadFlowsByNode ring.Buffer
	}

	workerWait chan struct{}
}

func (c *cancelFlowsCoordinator) getFlowsToCancel() (
	*execinfrapb.CancelDeadFlowsRequest,
	base.SQLInstanceID,
) {
	__antithesis_instrumentation__.Notify(467894)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mu.deadFlowsByNode.Len() == 0 {
		__antithesis_instrumentation__.Notify(467896)
		return nil, base.SQLInstanceID(0)
	} else {
		__antithesis_instrumentation__.Notify(467897)
	}
	__antithesis_instrumentation__.Notify(467895)
	deadFlows := c.mu.deadFlowsByNode.GetFirst().(*deadFlowsOnNode)
	c.mu.deadFlowsByNode.RemoveFirst()
	req := &execinfrapb.CancelDeadFlowsRequest{
		FlowIDs: deadFlows.ids,
	}
	return req, deadFlows.sqlInstanceID
}

func (c *cancelFlowsCoordinator) addFlowsToCancel(
	flows map[base.SQLInstanceID]*execinfrapb.FlowSpec,
) {
	__antithesis_instrumentation__.Notify(467898)
	c.mu.Lock()
	for sqlInstanceID, f := range flows {
		__antithesis_instrumentation__.Notify(467901)
		if sqlInstanceID != f.Gateway {
			__antithesis_instrumentation__.Notify(467902)

			found := false
			for j := 0; j < c.mu.deadFlowsByNode.Len(); j++ {
				__antithesis_instrumentation__.Notify(467904)
				deadFlows := c.mu.deadFlowsByNode.Get(j).(*deadFlowsOnNode)
				if sqlInstanceID == deadFlows.sqlInstanceID {
					__antithesis_instrumentation__.Notify(467905)
					deadFlows.ids = append(deadFlows.ids, f.FlowID)
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(467906)
				}
			}
			__antithesis_instrumentation__.Notify(467903)
			if !found {
				__antithesis_instrumentation__.Notify(467907)
				c.mu.deadFlowsByNode.AddLast(&deadFlowsOnNode{
					ids:           []execinfrapb.FlowID{f.FlowID},
					sqlInstanceID: sqlInstanceID,
				})
			} else {
				__antithesis_instrumentation__.Notify(467908)
			}
		} else {
			__antithesis_instrumentation__.Notify(467909)
		}
	}
	__antithesis_instrumentation__.Notify(467899)
	queueLength := c.mu.deadFlowsByNode.Len()
	c.mu.Unlock()

	numWorkersToWakeUp := numCancelingWorkers
	if numWorkersToWakeUp > queueLength {
		__antithesis_instrumentation__.Notify(467910)
		numWorkersToWakeUp = queueLength
	} else {
		__antithesis_instrumentation__.Notify(467911)
	}
	__antithesis_instrumentation__.Notify(467900)
	for i := 0; i < numWorkersToWakeUp; i++ {
		__antithesis_instrumentation__.Notify(467912)
		select {
		case c.workerWait <- struct{}{}:
			__antithesis_instrumentation__.Notify(467913)
		default:
			__antithesis_instrumentation__.Notify(467914)

			return
		}
	}
}

func (dsp *DistSQLPlanner) setupFlows(
	ctx context.Context,
	evalCtx *extendedEvalContext,
	leafInputState *roachpb.LeafTxnInputState,
	flows map[base.SQLInstanceID]*execinfrapb.FlowSpec,
	recv *DistSQLReceiver,
	localState distsql.LocalState,
	collectStats bool,
	statementSQL string,
) (context.Context, flowinfra.Flow, execinfra.OpChains, error) {
	__antithesis_instrumentation__.Notify(467915)
	thisNodeID := dsp.gatewaySQLInstanceID
	_, ok := flows[thisNodeID]
	if !ok {
		__antithesis_instrumentation__.Notify(467925)
		return nil, nil, nil, errors.AssertionFailedf("missing gateway flow")
	} else {
		__antithesis_instrumentation__.Notify(467926)
	}
	__antithesis_instrumentation__.Notify(467916)
	if localState.IsLocal && func() bool {
		__antithesis_instrumentation__.Notify(467927)
		return len(flows) != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(467928)
		return nil, nil, nil, errors.AssertionFailedf("IsLocal set but there's multiple flows")
	} else {
		__antithesis_instrumentation__.Notify(467929)
	}
	__antithesis_instrumentation__.Notify(467917)

	const setupFlowRequestStmtMaxLength = 500
	if len(statementSQL) > setupFlowRequestStmtMaxLength {
		__antithesis_instrumentation__.Notify(467930)
		statementSQL = statementSQL[:setupFlowRequestStmtMaxLength]
	} else {
		__antithesis_instrumentation__.Notify(467931)
	}
	__antithesis_instrumentation__.Notify(467918)
	setupReq := execinfrapb.SetupFlowRequest{
		LeafTxnInputState: leafInputState,
		Version:           execinfra.Version,
		EvalContext:       execinfrapb.MakeEvalContext(&evalCtx.EvalContext),
		TraceKV:           evalCtx.Tracing.KVTracingEnabled(),
		CollectStats:      collectStats,
		StatementSQL:      statementSQL,
	}

	var resultChan chan runnerResult
	if len(flows) > 1 {
		__antithesis_instrumentation__.Notify(467932)
		resultChan = make(chan runnerResult, len(flows)-1)
	} else {
		__antithesis_instrumentation__.Notify(467933)
	}
	__antithesis_instrumentation__.Notify(467919)

	if vectorizeMode := evalCtx.SessionData().VectorizeMode; vectorizeMode != sessiondatapb.VectorizeOff {
		__antithesis_instrumentation__.Notify(467934)

		for _, spec := range flows {
			__antithesis_instrumentation__.Notify(467935)
			if err := colflow.IsSupported(vectorizeMode, spec); err != nil {
				__antithesis_instrumentation__.Notify(467936)
				log.VEventf(ctx, 2, "failed to vectorize: %s", err)
				if vectorizeMode == sessiondatapb.VectorizeExperimentalAlways {
					__antithesis_instrumentation__.Notify(467938)
					return nil, nil, nil, err
				} else {
					__antithesis_instrumentation__.Notify(467939)
				}
				__antithesis_instrumentation__.Notify(467937)

				setupReq.EvalContext.SessionData.VectorizeMode = sessiondatapb.VectorizeOff
				break
			} else {
				__antithesis_instrumentation__.Notify(467940)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(467941)
	}
	__antithesis_instrumentation__.Notify(467920)
	for nodeID, flowSpec := range flows {
		__antithesis_instrumentation__.Notify(467942)
		if nodeID == thisNodeID {
			__antithesis_instrumentation__.Notify(467944)

			continue
		} else {
			__antithesis_instrumentation__.Notify(467945)
		}
		__antithesis_instrumentation__.Notify(467943)
		req := setupReq
		req.Flow = *flowSpec
		runReq := runnerRequest{
			ctx:           ctx,
			nodeDialer:    dsp.podNodeDialer,
			flowReq:       &req,
			sqlInstanceID: nodeID,
			resultChan:    resultChan,
		}

		select {
		case dsp.runnerChan <- runReq:
			__antithesis_instrumentation__.Notify(467946)
		default:
			__antithesis_instrumentation__.Notify(467947)
			runReq.run()
		}
	}
	__antithesis_instrumentation__.Notify(467921)

	var firstErr error

	for i := 0; i < len(flows)-1; i++ {
		__antithesis_instrumentation__.Notify(467948)
		res := <-resultChan
		if firstErr == nil {
			__antithesis_instrumentation__.Notify(467949)
			firstErr = res.err
		} else {
			__antithesis_instrumentation__.Notify(467950)
		}

	}
	__antithesis_instrumentation__.Notify(467922)
	if firstErr != nil {
		__antithesis_instrumentation__.Notify(467951)
		return nil, nil, nil, firstErr
	} else {
		__antithesis_instrumentation__.Notify(467952)
	}
	__antithesis_instrumentation__.Notify(467923)

	setupReq.Flow = *flows[thisNodeID]
	var batchReceiver execinfra.BatchReceiver
	if recv.batchWriter != nil {
		__antithesis_instrumentation__.Notify(467953)

		batchReceiver = recv
	} else {
		__antithesis_instrumentation__.Notify(467954)
	}
	__antithesis_instrumentation__.Notify(467924)
	return dsp.distSQLSrv.SetupLocalSyncFlow(ctx, evalCtx.Mon, &setupReq, recv, batchReceiver, localState)
}

func (dsp *DistSQLPlanner) Run(
	ctx context.Context,
	planCtx *PlanningCtx,
	txn *kv.Txn,
	plan *PhysicalPlan,
	recv *DistSQLReceiver,
	evalCtx *extendedEvalContext,
	finishedSetupFn func(),
) (cleanup func()) {
	__antithesis_instrumentation__.Notify(467955)
	cleanup = func() { __antithesis_instrumentation__.Notify(467974) }
	__antithesis_instrumentation__.Notify(467956)

	flows := plan.GenerateFlowSpecs()
	defer func() {
		__antithesis_instrumentation__.Notify(467975)
		for _, flowSpec := range flows {
			__antithesis_instrumentation__.Notify(467976)
			physicalplan.ReleaseFlowSpec(flowSpec)
		}
	}()
	__antithesis_instrumentation__.Notify(467957)
	if _, ok := flows[dsp.gatewaySQLInstanceID]; !ok {
		__antithesis_instrumentation__.Notify(467977)
		recv.SetError(errors.Errorf("expected to find gateway flow"))
		return cleanup
	} else {
		__antithesis_instrumentation__.Notify(467978)
	}
	__antithesis_instrumentation__.Notify(467958)

	var (
		localState     distsql.LocalState
		leafInputState *roachpb.LeafTxnInputState
	)

	localState.EvalContext = &evalCtx.EvalContext
	localState.Txn = txn
	localState.LocalProcs = plan.LocalProcessors

	localState.PreserveFlowSpecs = planCtx.saveFlows != nil

	if planCtx.planner != nil && func() bool {
		__antithesis_instrumentation__.Notify(467979)
		return !planCtx.planner.isInternalPlanner == true
	}() == true {
		__antithesis_instrumentation__.Notify(467980)
		localState.Collection = planCtx.planner.Descriptors()
	} else {
		__antithesis_instrumentation__.Notify(467981)
	}
	__antithesis_instrumentation__.Notify(467959)

	if planCtx.isLocal {
		__antithesis_instrumentation__.Notify(467982)
		localState.IsLocal = true
		if planCtx.parallelizeScansIfLocal {
			__antithesis_instrumentation__.Notify(467983)

			for _, flow := range flows {
				__antithesis_instrumentation__.Notify(467984)
				localState.HasConcurrency = localState.HasConcurrency || func() bool {
					__antithesis_instrumentation__.Notify(467985)
					return execinfra.HasParallelProcessors(flow) == true
				}() == true
			}
		} else {
			__antithesis_instrumentation__.Notify(467986)
		}
	} else {
		__antithesis_instrumentation__.Notify(467987)
	}
	__antithesis_instrumentation__.Notify(467960)

	if row.CanUseStreamer(ctx, dsp.st) {
		__antithesis_instrumentation__.Notify(467988)
		for _, proc := range plan.Processors {
			__antithesis_instrumentation__.Notify(467989)
			if jr := proc.Spec.Core.JoinReader; jr != nil {
				__antithesis_instrumentation__.Notify(467990)
				if jr.IsIndexJoin() {
					__antithesis_instrumentation__.Notify(467991)

					localState.HasConcurrency = true
					break
				} else {
					__antithesis_instrumentation__.Notify(467992)
				}
			} else {
				__antithesis_instrumentation__.Notify(467993)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(467994)
	}
	__antithesis_instrumentation__.Notify(467961)
	if localState.MustUseLeafTxn() && func() bool {
		__antithesis_instrumentation__.Notify(467995)
		return txn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(467996)

		tis, err := txn.GetLeafTxnInputStateOrRejectClient(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(467998)
			log.Infof(ctx, "%s: %s", clientRejectedMsg, err)
			recv.SetError(err)
			return cleanup
		} else {
			__antithesis_instrumentation__.Notify(467999)
		}
		__antithesis_instrumentation__.Notify(467997)
		leafInputState = tis
	} else {
		__antithesis_instrumentation__.Notify(468000)
	}
	__antithesis_instrumentation__.Notify(467962)

	if logPlanDiagram {
		__antithesis_instrumentation__.Notify(468001)
		log.VEvent(ctx, 3, "creating plan diagram for logging")
		var stmtStr string
		if planCtx.planner != nil && func() bool {
			__antithesis_instrumentation__.Notify(468003)
			return planCtx.planner.stmt.AST != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(468004)
			stmtStr = planCtx.planner.stmt.String()
		} else {
			__antithesis_instrumentation__.Notify(468005)
		}
		__antithesis_instrumentation__.Notify(468002)
		_, url, err := execinfrapb.GeneratePlanDiagramURL(stmtStr, flows, execinfrapb.DiagramFlags{})
		if err != nil {
			__antithesis_instrumentation__.Notify(468006)
			log.Infof(ctx, "error generating diagram: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(468007)
			log.Infof(ctx, "plan diagram URL:\n%s", url.String())
		}
	} else {
		__antithesis_instrumentation__.Notify(468008)
	}
	__antithesis_instrumentation__.Notify(467963)

	log.VEvent(ctx, 2, "running DistSQL plan")

	dsp.distSQLSrv.ServerConfig.Metrics.QueryStart()
	defer dsp.distSQLSrv.ServerConfig.Metrics.QueryStop()

	recv.outputTypes = plan.GetResultTypes()
	recv.contendedQueryMetric = dsp.distSQLSrv.Metrics.ContendedQueriesCount

	if len(flows) == 1 {
		__antithesis_instrumentation__.Notify(468009)

		localState.IsLocal = true
	} else {
		__antithesis_instrumentation__.Notify(468010)
		defer func() {
			__antithesis_instrumentation__.Notify(468011)
			if recv.resultWriter.Err() != nil {
				__antithesis_instrumentation__.Notify(468012)

				dsp.cancelFlowsCoordinator.addFlowsToCancel(flows)
			} else {
				__antithesis_instrumentation__.Notify(468013)
			}
		}()
	}
	__antithesis_instrumentation__.Notify(467964)

	var statementSQL string
	if planCtx.planner != nil {
		__antithesis_instrumentation__.Notify(468014)
		statementSQL = planCtx.planner.stmt.StmtNoConstants
	} else {
		__antithesis_instrumentation__.Notify(468015)
	}
	__antithesis_instrumentation__.Notify(467965)
	ctx, flow, opChains, err := dsp.setupFlows(
		ctx, evalCtx, leafInputState, flows, recv, localState, planCtx.collectExecStats, statementSQL,
	)

	if flow != nil {
		__antithesis_instrumentation__.Notify(468016)
		cleanup = func() {
			__antithesis_instrumentation__.Notify(468017)
			flow.Cleanup(ctx)
		}
	} else {
		__antithesis_instrumentation__.Notify(468018)
	}
	__antithesis_instrumentation__.Notify(467966)
	if err != nil {
		__antithesis_instrumentation__.Notify(468019)
		recv.SetError(err)
		return cleanup
	} else {
		__antithesis_instrumentation__.Notify(468020)
	}
	__antithesis_instrumentation__.Notify(467967)

	if finishedSetupFn != nil {
		__antithesis_instrumentation__.Notify(468021)
		finishedSetupFn()
	} else {
		__antithesis_instrumentation__.Notify(468022)
	}
	__antithesis_instrumentation__.Notify(467968)

	if planCtx.planner != nil && func() bool {
		__antithesis_instrumentation__.Notify(468023)
		return flow.IsVectorized() == true
	}() == true {
		__antithesis_instrumentation__.Notify(468024)
		planCtx.planner.curPlan.flags.Set(planFlagVectorized)
	} else {
		__antithesis_instrumentation__.Notify(468025)
	}
	__antithesis_instrumentation__.Notify(467969)

	if planCtx.saveFlows != nil {
		__antithesis_instrumentation__.Notify(468026)
		if err := planCtx.saveFlows(flows, opChains); err != nil {
			__antithesis_instrumentation__.Notify(468027)
			recv.SetError(err)
			return cleanup
		} else {
			__antithesis_instrumentation__.Notify(468028)
		}
	} else {
		__antithesis_instrumentation__.Notify(468029)
	}
	__antithesis_instrumentation__.Notify(467970)

	if txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(468030)
		return !localState.MustUseLeafTxn() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(468031)
		return flow.ConcurrentTxnUse() == true
	}() == true {
		__antithesis_instrumentation__.Notify(468032)
		recv.SetError(errors.AssertionFailedf(
			"unexpected concurrency for a flow that was forced to be planned locally"))
		return cleanup
	} else {
		__antithesis_instrumentation__.Notify(468033)
	}
	__antithesis_instrumentation__.Notify(467971)

	flow.Run(ctx, func() { __antithesis_instrumentation__.Notify(468034) })
	__antithesis_instrumentation__.Notify(467972)

	if planCtx.planner != nil && func() bool {
		__antithesis_instrumentation__.Notify(468035)
		return !planCtx.ignoreClose == true
	}() == true {
		__antithesis_instrumentation__.Notify(468036)

		curPlan := &planCtx.planner.curPlan
		return func() {
			__antithesis_instrumentation__.Notify(468037)

			curPlan.close(ctx)
			flow.Cleanup(ctx)
		}
	} else {
		__antithesis_instrumentation__.Notify(468038)
	}
	__antithesis_instrumentation__.Notify(467973)

	return cleanup
}

type DistSQLReceiver struct {
	ctx context.Context

	resultWriter rowResultWriter
	batchWriter  batchResultWriter

	stmtType tree.StatementReturnType

	outputTypes []*types.T

	existsMode bool

	discardRows bool

	commErr error

	row    tree.Datums
	status execinfra.ConsumerStatus
	alloc  tree.DatumAlloc
	closed bool

	rangeCache *rangecache.RangeCache
	tracing    *SessionTracing

	cleanup func()

	txn *kv.Txn

	clockUpdater clockUpdater

	stats *topLevelQueryStats

	expectedRowsRead int64
	progressAtomic   *uint64

	contendedQueryMetric *metric.Counter

	contentionRegistry *contention.Registry

	testingKnobs struct {
		pushCallback func(rowenc.EncDatumRow, *execinfrapb.ProducerMetadata)
	}
}

type rowResultWriter interface {
	AddRow(ctx context.Context, row tree.Datums) error
	IncrementRowsAffected(ctx context.Context, n int)
	SetError(error)
	Err() error
}

type batchResultWriter interface {
	AddBatch(context.Context, coldata.Batch) error
}

type MetadataResultWriter interface {
	AddMeta(ctx context.Context, meta *execinfrapb.ProducerMetadata)
}

type MetadataCallbackWriter struct {
	rowResultWriter
	fn func(ctx context.Context, meta *execinfrapb.ProducerMetadata) error
}

func (w *MetadataCallbackWriter) AddMeta(ctx context.Context, meta *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(468039)
	if err := w.fn(ctx, meta); err != nil {
		__antithesis_instrumentation__.Notify(468040)
		w.SetError(err)
	} else {
		__antithesis_instrumentation__.Notify(468041)
	}
}

func NewMetadataCallbackWriter(
	rowResultWriter rowResultWriter,
	metaFn func(ctx context.Context, meta *execinfrapb.ProducerMetadata) error,
) *MetadataCallbackWriter {
	__antithesis_instrumentation__.Notify(468042)
	return &MetadataCallbackWriter{rowResultWriter: rowResultWriter, fn: metaFn}
}

type errOnlyResultWriter struct {
	err error
}

var _ rowResultWriter = &errOnlyResultWriter{}
var _ batchResultWriter = &errOnlyResultWriter{}

func (w *errOnlyResultWriter) SetError(err error) {
	__antithesis_instrumentation__.Notify(468043)
	w.err = err
}
func (w *errOnlyResultWriter) Err() error {
	__antithesis_instrumentation__.Notify(468044)
	return w.err
}

func (w *errOnlyResultWriter) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(468045)
	panic("AddRow not supported by errOnlyResultWriter")
}

func (w *errOnlyResultWriter) AddBatch(ctx context.Context, batch coldata.Batch) error {
	__antithesis_instrumentation__.Notify(468046)
	panic("AddBatch not supported by errOnlyResultWriter")
}

func (w *errOnlyResultWriter) IncrementRowsAffected(ctx context.Context, n int) {
	__antithesis_instrumentation__.Notify(468047)
	panic("IncrementRowsAffected not supported by errOnlyResultWriter")
}

type RowResultWriter struct {
	rowContainer *rowContainerHelper
	rowsAffected int
	err          error
}

var _ rowResultWriter = &RowResultWriter{}

func NewRowResultWriter(rowContainer *rowContainerHelper) *RowResultWriter {
	__antithesis_instrumentation__.Notify(468048)
	return &RowResultWriter{rowContainer: rowContainer}
}

func (b *RowResultWriter) IncrementRowsAffected(ctx context.Context, n int) {
	__antithesis_instrumentation__.Notify(468049)
	b.rowsAffected += n
}

func (b *RowResultWriter) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(468050)
	if b.rowContainer != nil {
		__antithesis_instrumentation__.Notify(468052)
		return b.rowContainer.AddRow(ctx, row)
	} else {
		__antithesis_instrumentation__.Notify(468053)
	}
	__antithesis_instrumentation__.Notify(468051)
	return nil
}

func (b *RowResultWriter) SetError(err error) {
	__antithesis_instrumentation__.Notify(468054)
	b.err = err
}

func (b *RowResultWriter) Err() error {
	__antithesis_instrumentation__.Notify(468055)
	return b.err
}

type CallbackResultWriter struct {
	fn           func(ctx context.Context, row tree.Datums) error
	rowsAffected int
	err          error
}

var _ rowResultWriter = &CallbackResultWriter{}

func NewCallbackResultWriter(
	fn func(ctx context.Context, row tree.Datums) error,
) *CallbackResultWriter {
	__antithesis_instrumentation__.Notify(468056)
	return &CallbackResultWriter{fn: fn}
}

func (c *CallbackResultWriter) IncrementRowsAffected(ctx context.Context, n int) {
	__antithesis_instrumentation__.Notify(468057)
	c.rowsAffected += n
}

func (c *CallbackResultWriter) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(468058)
	return c.fn(ctx, row)
}

func (c *CallbackResultWriter) SetError(err error) {
	__antithesis_instrumentation__.Notify(468059)
	c.err = err
}

func (c *CallbackResultWriter) Err() error {
	__antithesis_instrumentation__.Notify(468060)
	return c.err
}

var _ execinfra.RowReceiver = &DistSQLReceiver{}
var _ execinfra.BatchReceiver = &DistSQLReceiver{}

var receiverSyncPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(468061)
		return &DistSQLReceiver{}
	},
}

type clockUpdater interface {
	Update(observedTS hlc.ClockTimestamp)
}

func MakeDistSQLReceiver(
	ctx context.Context,
	resultWriter rowResultWriter,
	stmtType tree.StatementReturnType,
	rangeCache *rangecache.RangeCache,
	txn *kv.Txn,
	clockUpdater clockUpdater,
	tracing *SessionTracing,
	contentionRegistry *contention.Registry,
	testingPushCallback func(rowenc.EncDatumRow, *execinfrapb.ProducerMetadata),
) *DistSQLReceiver {
	__antithesis_instrumentation__.Notify(468062)
	consumeCtx, cleanup := tracing.TraceExecConsume(ctx)
	r := receiverSyncPool.Get().(*DistSQLReceiver)

	var batchWriter batchResultWriter
	if commandResult, ok := resultWriter.(RestrictedCommandResult); ok {
		__antithesis_instrumentation__.Notify(468064)
		if commandResult.SupportsAddBatch() {
			__antithesis_instrumentation__.Notify(468065)
			batchWriter = commandResult
		} else {
			__antithesis_instrumentation__.Notify(468066)
		}
	} else {
		__antithesis_instrumentation__.Notify(468067)
	}
	__antithesis_instrumentation__.Notify(468063)
	*r = DistSQLReceiver{
		ctx:                consumeCtx,
		cleanup:            cleanup,
		resultWriter:       resultWriter,
		batchWriter:        batchWriter,
		rangeCache:         rangeCache,
		txn:                txn,
		clockUpdater:       clockUpdater,
		stats:              &topLevelQueryStats{},
		stmtType:           stmtType,
		tracing:            tracing,
		contentionRegistry: contentionRegistry,
	}
	r.testingKnobs.pushCallback = testingPushCallback
	return r
}

func (r *DistSQLReceiver) Release() {
	__antithesis_instrumentation__.Notify(468068)
	r.cleanup()
	*r = DistSQLReceiver{}
	receiverSyncPool.Put(r)
}

func (r *DistSQLReceiver) clone() *DistSQLReceiver {
	__antithesis_instrumentation__.Notify(468069)
	ret := receiverSyncPool.Get().(*DistSQLReceiver)
	*ret = DistSQLReceiver{
		ctx:                r.ctx,
		cleanup:            func() { __antithesis_instrumentation__.Notify(468071) },
		rangeCache:         r.rangeCache,
		txn:                r.txn,
		clockUpdater:       r.clockUpdater,
		stats:              r.stats,
		stmtType:           tree.Rows,
		tracing:            r.tracing,
		contentionRegistry: r.contentionRegistry,
	}
	__antithesis_instrumentation__.Notify(468070)
	return ret
}

func (r *DistSQLReceiver) SetError(err error) {
	__antithesis_instrumentation__.Notify(468072)
	r.resultWriter.SetError(err)

	if r.ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(468073)
		log.VEventf(r.ctx, 1, "encountered error (transitioning to shutting down): %v", r.ctx.Err())
		r.status = execinfra.ConsumerClosed
	} else {
		__antithesis_instrumentation__.Notify(468074)
		log.VEventf(r.ctx, 1, "encountered error (transitioning to draining): %v", err)
		r.status = execinfra.DrainRequested
	}
}

func (r *DistSQLReceiver) pushMeta(meta *execinfrapb.ProducerMetadata) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(468075)
	if metaWriter, ok := r.resultWriter.(MetadataResultWriter); ok {
		__antithesis_instrumentation__.Notify(468082)
		metaWriter.AddMeta(r.ctx, meta)
	} else {
		__antithesis_instrumentation__.Notify(468083)
	}
	__antithesis_instrumentation__.Notify(468076)
	if meta.LeafTxnFinalState != nil {
		__antithesis_instrumentation__.Notify(468084)
		if r.txn != nil {
			__antithesis_instrumentation__.Notify(468085)
			if r.txn.ID() == meta.LeafTxnFinalState.Txn.ID {
				__antithesis_instrumentation__.Notify(468086)
				if err := r.txn.UpdateRootWithLeafFinalState(r.ctx, meta.LeafTxnFinalState); err != nil {
					__antithesis_instrumentation__.Notify(468087)
					r.SetError(err)
				} else {
					__antithesis_instrumentation__.Notify(468088)
				}
			} else {
				__antithesis_instrumentation__.Notify(468089)
			}
		} else {
			__antithesis_instrumentation__.Notify(468090)
			r.SetError(
				errors.Errorf("received a leaf final state (%s); but have no root", meta.LeafTxnFinalState))
		}
	} else {
		__antithesis_instrumentation__.Notify(468091)
	}
	__antithesis_instrumentation__.Notify(468077)
	if meta.Err != nil {
		__antithesis_instrumentation__.Notify(468092)

		if roachpb.ErrPriority(meta.Err) > roachpb.ErrPriority(r.resultWriter.Err()) {
			__antithesis_instrumentation__.Notify(468093)
			if r.txn != nil {
				__antithesis_instrumentation__.Notify(468095)
				if retryErr := (*roachpb.UnhandledRetryableError)(nil); errors.As(meta.Err, &retryErr) {
					__antithesis_instrumentation__.Notify(468096)

					meta.Err = r.txn.UpdateStateOnRemoteRetryableErr(r.ctx, &retryErr.PErr)

					if r.clockUpdater != nil {
						__antithesis_instrumentation__.Notify(468097)
						r.clockUpdater.Update(retryErr.PErr.Now)
					} else {
						__antithesis_instrumentation__.Notify(468098)
					}
				} else {
					__antithesis_instrumentation__.Notify(468099)
				}
			} else {
				__antithesis_instrumentation__.Notify(468100)
			}
			__antithesis_instrumentation__.Notify(468094)
			r.SetError(meta.Err)
		} else {
			__antithesis_instrumentation__.Notify(468101)
		}
	} else {
		__antithesis_instrumentation__.Notify(468102)
	}
	__antithesis_instrumentation__.Notify(468078)
	if len(meta.Ranges) > 0 {
		__antithesis_instrumentation__.Notify(468103)
		r.rangeCache.Insert(r.ctx, meta.Ranges...)
	} else {
		__antithesis_instrumentation__.Notify(468104)
	}
	__antithesis_instrumentation__.Notify(468079)
	if len(meta.TraceData) > 0 {
		__antithesis_instrumentation__.Notify(468105)
		if span := tracing.SpanFromContext(r.ctx); span != nil {
			__antithesis_instrumentation__.Notify(468107)
			span.ImportRemoteSpans(meta.TraceData)
		} else {
			__antithesis_instrumentation__.Notify(468108)
		}
		__antithesis_instrumentation__.Notify(468106)
		var ev roachpb.ContentionEvent
		for i := range meta.TraceData {
			__antithesis_instrumentation__.Notify(468109)
			meta.TraceData[i].Structured(func(any *pbtypes.Any, _ time.Time) {
				__antithesis_instrumentation__.Notify(468110)
				if !pbtypes.Is(any, &ev) {
					__antithesis_instrumentation__.Notify(468115)
					return
				} else {
					__antithesis_instrumentation__.Notify(468116)
				}
				__antithesis_instrumentation__.Notify(468111)
				if err := pbtypes.UnmarshalAny(any, &ev); err != nil {
					__antithesis_instrumentation__.Notify(468117)
					return
				} else {
					__antithesis_instrumentation__.Notify(468118)
				}
				__antithesis_instrumentation__.Notify(468112)
				if r.contendedQueryMetric != nil {
					__antithesis_instrumentation__.Notify(468119)

					r.contendedQueryMetric.Inc(1)
					r.contendedQueryMetric = nil
				} else {
					__antithesis_instrumentation__.Notify(468120)
				}
				__antithesis_instrumentation__.Notify(468113)
				contentionEvent := contentionpb.ExtendedContentionEvent{
					BlockingEvent: ev,
				}
				if r.txn != nil {
					__antithesis_instrumentation__.Notify(468121)
					contentionEvent.WaitingTxnID = r.txn.ID()
				} else {
					__antithesis_instrumentation__.Notify(468122)
				}
				__antithesis_instrumentation__.Notify(468114)
				r.contentionRegistry.AddContentionEvent(contentionEvent)
			})
		}
	} else {
		__antithesis_instrumentation__.Notify(468123)
	}
	__antithesis_instrumentation__.Notify(468080)
	if meta.Metrics != nil {
		__antithesis_instrumentation__.Notify(468124)
		r.stats.bytesRead += meta.Metrics.BytesRead
		r.stats.rowsRead += meta.Metrics.RowsRead
		r.stats.rowsWritten += meta.Metrics.RowsWritten
		if r.progressAtomic != nil && func() bool {
			__antithesis_instrumentation__.Notify(468126)
			return r.expectedRowsRead != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(468127)
			progress := float64(r.stats.rowsRead) / float64(r.expectedRowsRead)
			atomic.StoreUint64(r.progressAtomic, math.Float64bits(progress))
		} else {
			__antithesis_instrumentation__.Notify(468128)
		}
		__antithesis_instrumentation__.Notify(468125)
		meta.Metrics.Release()
	} else {
		__antithesis_instrumentation__.Notify(468129)
	}
	__antithesis_instrumentation__.Notify(468081)

	meta.Release()
	return r.status
}

func (r *DistSQLReceiver) handleCommErr(commErr error) {
	__antithesis_instrumentation__.Notify(468130)

	if errors.Is(commErr, ErrLimitedResultClosed) {
		__antithesis_instrumentation__.Notify(468131)
		log.VEvent(r.ctx, 1, "encountered ErrLimitedResultClosed (transitioning to draining)")
		r.status = execinfra.DrainRequested
	} else {
		__antithesis_instrumentation__.Notify(468132)
		if errors.Is(commErr, errIEResultChannelClosed) {
			__antithesis_instrumentation__.Notify(468133)
			log.VEvent(r.ctx, 1, "encountered errIEResultChannelClosed (transitioning to draining)")
			r.status = execinfra.DrainRequested
		} else {
			__antithesis_instrumentation__.Notify(468134)

			r.SetError(commErr)

			if !errors.Is(commErr, ErrLimitedResultNotSupported) {
				__antithesis_instrumentation__.Notify(468135)
				r.commErr = commErr
			} else {
				__antithesis_instrumentation__.Notify(468136)
			}
		}
	}
}

func (r *DistSQLReceiver) Push(
	row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata,
) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(468137)
	if r.testingKnobs.pushCallback != nil {
		__antithesis_instrumentation__.Notify(468146)
		r.testingKnobs.pushCallback(row, meta)
	} else {
		__antithesis_instrumentation__.Notify(468147)
	}
	__antithesis_instrumentation__.Notify(468138)
	if meta != nil {
		__antithesis_instrumentation__.Notify(468148)
		return r.pushMeta(meta)
	} else {
		__antithesis_instrumentation__.Notify(468149)
	}
	__antithesis_instrumentation__.Notify(468139)
	if r.resultWriter.Err() == nil && func() bool {
		__antithesis_instrumentation__.Notify(468150)
		return r.ctx.Err() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(468151)
		r.SetError(r.ctx.Err())
	} else {
		__antithesis_instrumentation__.Notify(468152)
	}
	__antithesis_instrumentation__.Notify(468140)
	if r.status != execinfra.NeedMoreRows {
		__antithesis_instrumentation__.Notify(468153)
		return r.status
	} else {
		__antithesis_instrumentation__.Notify(468154)
	}
	__antithesis_instrumentation__.Notify(468141)

	if r.stmtType != tree.Rows {
		__antithesis_instrumentation__.Notify(468155)
		n := int(tree.MustBeDInt(row[0].Datum))

		r.resultWriter.IncrementRowsAffected(r.ctx, n)
		return r.status
	} else {
		__antithesis_instrumentation__.Notify(468156)
	}
	__antithesis_instrumentation__.Notify(468142)

	if r.discardRows {
		__antithesis_instrumentation__.Notify(468157)

		return r.status
	} else {
		__antithesis_instrumentation__.Notify(468158)
	}
	__antithesis_instrumentation__.Notify(468143)

	if r.existsMode {
		__antithesis_instrumentation__.Notify(468159)

		r.row = []tree.Datum{}
		log.VEvent(r.ctx, 2, `a row is pushed in "exists" mode, so transition to draining`)
		r.status = execinfra.DrainRequested
	} else {
		__antithesis_instrumentation__.Notify(468160)
		if r.row == nil {
			__antithesis_instrumentation__.Notify(468162)
			r.row = make(tree.Datums, len(row))
		} else {
			__antithesis_instrumentation__.Notify(468163)
		}
		__antithesis_instrumentation__.Notify(468161)
		for i, encDatum := range row {
			__antithesis_instrumentation__.Notify(468164)
			err := encDatum.EnsureDecoded(r.outputTypes[i], &r.alloc)
			if err != nil {
				__antithesis_instrumentation__.Notify(468166)
				r.SetError(err)
				return r.status
			} else {
				__antithesis_instrumentation__.Notify(468167)
			}
			__antithesis_instrumentation__.Notify(468165)
			r.row[i] = encDatum.Datum
		}
	}
	__antithesis_instrumentation__.Notify(468144)
	r.tracing.TraceExecRowsResult(r.ctx, r.row)
	if commErr := r.resultWriter.AddRow(r.ctx, r.row); commErr != nil {
		__antithesis_instrumentation__.Notify(468168)
		r.handleCommErr(commErr)
	} else {
		__antithesis_instrumentation__.Notify(468169)
	}
	__antithesis_instrumentation__.Notify(468145)
	return r.status
}

func (r *DistSQLReceiver) PushBatch(
	batch coldata.Batch, meta *execinfrapb.ProducerMetadata,
) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(468170)
	if meta != nil {
		__antithesis_instrumentation__.Notify(468179)
		return r.pushMeta(meta)
	} else {
		__antithesis_instrumentation__.Notify(468180)
	}
	__antithesis_instrumentation__.Notify(468171)
	if r.resultWriter.Err() == nil && func() bool {
		__antithesis_instrumentation__.Notify(468181)
		return r.ctx.Err() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(468182)
		r.SetError(r.ctx.Err())
	} else {
		__antithesis_instrumentation__.Notify(468183)
	}
	__antithesis_instrumentation__.Notify(468172)
	if r.status != execinfra.NeedMoreRows {
		__antithesis_instrumentation__.Notify(468184)
		return r.status
	} else {
		__antithesis_instrumentation__.Notify(468185)
	}
	__antithesis_instrumentation__.Notify(468173)

	if batch.Length() == 0 {
		__antithesis_instrumentation__.Notify(468186)

		return r.status
	} else {
		__antithesis_instrumentation__.Notify(468187)
	}
	__antithesis_instrumentation__.Notify(468174)

	if r.stmtType != tree.Rows {
		__antithesis_instrumentation__.Notify(468188)

		r.resultWriter.IncrementRowsAffected(r.ctx, int(batch.ColVec(0).Int64()[0]))
		return r.status
	} else {
		__antithesis_instrumentation__.Notify(468189)
	}
	__antithesis_instrumentation__.Notify(468175)

	if r.discardRows {
		__antithesis_instrumentation__.Notify(468190)

		return r.status
	} else {
		__antithesis_instrumentation__.Notify(468191)
	}
	__antithesis_instrumentation__.Notify(468176)

	if r.existsMode {
		__antithesis_instrumentation__.Notify(468192)

		panic("unsupported exists mode for PushBatch")
	} else {
		__antithesis_instrumentation__.Notify(468193)
	}
	__antithesis_instrumentation__.Notify(468177)
	r.tracing.TraceExecBatchResult(r.ctx, batch)
	if commErr := r.batchWriter.AddBatch(r.ctx, batch); commErr != nil {
		__antithesis_instrumentation__.Notify(468194)
		r.handleCommErr(commErr)
	} else {
		__antithesis_instrumentation__.Notify(468195)
	}
	__antithesis_instrumentation__.Notify(468178)
	return r.status
}

var (
	ErrLimitedResultNotSupported = unimplemented.NewWithIssue(40195, "multiple active portals not supported")

	ErrLimitedResultClosed = errors.New("row count limit closed")
)

func (r *DistSQLReceiver) ProducerDone() {
	__antithesis_instrumentation__.Notify(468196)
	if r.closed {
		__antithesis_instrumentation__.Notify(468198)
		panic("double close")
	} else {
		__antithesis_instrumentation__.Notify(468199)
	}
	__antithesis_instrumentation__.Notify(468197)
	r.closed = true
}

func (dsp *DistSQLPlanner) PlanAndRunSubqueries(
	ctx context.Context,
	planner *planner,
	evalCtxFactory func() *extendedEvalContext,
	subqueryPlans []subquery,
	recv *DistSQLReceiver,
	subqueryResultMemAcc *mon.BoundAccount,
) bool {
	__antithesis_instrumentation__.Notify(468200)
	for planIdx, subqueryPlan := range subqueryPlans {
		__antithesis_instrumentation__.Notify(468202)
		if err := dsp.planAndRunSubquery(
			ctx,
			planIdx,
			subqueryPlan,
			planner,
			evalCtxFactory(),
			subqueryPlans,
			recv,
			subqueryResultMemAcc,
		); err != nil {
			__antithesis_instrumentation__.Notify(468203)
			recv.SetError(err)

			for i := range subqueryPlans {
				__antithesis_instrumentation__.Notify(468205)
				subqueryPlans[i].plan.Close(ctx)
			}
			__antithesis_instrumentation__.Notify(468204)
			return false
		} else {
			__antithesis_instrumentation__.Notify(468206)
		}
	}
	__antithesis_instrumentation__.Notify(468201)

	return true
}

func (dsp *DistSQLPlanner) planAndRunSubquery(
	ctx context.Context,
	planIdx int,
	subqueryPlan subquery,
	planner *planner,
	evalCtx *extendedEvalContext,
	subqueryPlans []subquery,
	recv *DistSQLReceiver,
	subqueryResultMemAcc *mon.BoundAccount,
) error {
	__antithesis_instrumentation__.Notify(468207)
	subqueryMonitor := mon.NewMonitor(
		"subquery",
		mon.MemoryResource,
		dsp.distSQLSrv.Metrics.CurBytesCount,
		dsp.distSQLSrv.Metrics.MaxBytesHist,
		-1,
		noteworthyMemoryUsageBytes,
		dsp.distSQLSrv.Settings,
	)
	subqueryMonitor.Start(ctx, evalCtx.Mon, mon.BoundAccount{})
	defer subqueryMonitor.Stop(ctx)

	subqueryMemAccount := subqueryMonitor.MakeBoundAccount()
	defer subqueryMemAccount.Close(ctx)

	distributeSubquery := getPlanDistribution(
		ctx, planner, planner.execCfg.NodeID, planner.SessionData().DistSQLMode, subqueryPlan.plan,
	).WillDistribute()
	distribute := DistributionType(DistributionTypeNone)
	if distributeSubquery {
		__antithesis_instrumentation__.Notify(468215)
		distribute = DistributionTypeSystemTenantOnly
	} else {
		__antithesis_instrumentation__.Notify(468216)
	}
	__antithesis_instrumentation__.Notify(468208)
	subqueryPlanCtx := dsp.NewPlanningCtx(ctx, evalCtx, planner, planner.txn,
		distribute)
	subqueryPlanCtx.stmtType = tree.Rows
	if planner.instrumentation.ShouldSaveFlows() {
		__antithesis_instrumentation__.Notify(468217)
		subqueryPlanCtx.saveFlows = subqueryPlanCtx.getDefaultSaveFlowsFunc(ctx, planner, planComponentTypeSubquery)
	} else {
		__antithesis_instrumentation__.Notify(468218)
	}
	__antithesis_instrumentation__.Notify(468209)
	subqueryPlanCtx.traceMetadata = planner.instrumentation.traceMetadata
	subqueryPlanCtx.collectExecStats = planner.instrumentation.ShouldCollectExecStats()

	subqueryPlanCtx.ignoreClose = true
	subqueryPhysPlan, physPlanCleanup, err := dsp.createPhysPlan(ctx, subqueryPlanCtx, subqueryPlan.plan)
	defer physPlanCleanup()
	if err != nil {
		__antithesis_instrumentation__.Notify(468219)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468220)
	}
	__antithesis_instrumentation__.Notify(468210)
	dsp.finalizePlanWithRowCount(subqueryPlanCtx, subqueryPhysPlan, subqueryPlan.rowCount)

	subqueryRecv := recv.clone()
	defer subqueryRecv.Release()
	var typs []*types.T
	if subqueryPlan.execMode == rowexec.SubqueryExecModeExists {
		__antithesis_instrumentation__.Notify(468221)
		subqueryRecv.existsMode = true
		typs = []*types.T{}
	} else {
		__antithesis_instrumentation__.Notify(468222)
		typs = subqueryPhysPlan.GetResultTypes()
	}
	__antithesis_instrumentation__.Notify(468211)
	var rows rowContainerHelper
	rows.Init(typs, evalCtx, "subquery")
	defer rows.Close(ctx)

	subqueryRowReceiver := NewRowResultWriter(&rows)
	subqueryRecv.resultWriter = subqueryRowReceiver
	subqueryPlans[planIdx].started = true
	dsp.Run(ctx, subqueryPlanCtx, planner.txn, subqueryPhysPlan, subqueryRecv, evalCtx, nil)()
	if err := subqueryRowReceiver.Err(); err != nil {
		__antithesis_instrumentation__.Notify(468223)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468224)
	}
	__antithesis_instrumentation__.Notify(468212)
	var alreadyAccountedFor int64
	switch subqueryPlan.execMode {
	case rowexec.SubqueryExecModeExists:
		__antithesis_instrumentation__.Notify(468225)

		hasRows := rows.Len() != 0
		subqueryPlans[planIdx].result = tree.MakeDBool(tree.DBool(hasRows))
	case rowexec.SubqueryExecModeAllRows, rowexec.SubqueryExecModeAllRowsNormalized:
		__antithesis_instrumentation__.Notify(468226)

		var result tree.DTuple
		iterator := newRowContainerIterator(ctx, rows, typs)
		defer iterator.Close()
		for {
			__antithesis_instrumentation__.Notify(468231)
			row, err := iterator.Next()
			if err != nil {
				__antithesis_instrumentation__.Notify(468236)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468237)
			}
			__antithesis_instrumentation__.Notify(468232)
			if row == nil {
				__antithesis_instrumentation__.Notify(468238)
				break
			} else {
				__antithesis_instrumentation__.Notify(468239)
			}
			__antithesis_instrumentation__.Notify(468233)
			var toAppend tree.Datum
			if row.Len() == 1 {
				__antithesis_instrumentation__.Notify(468240)

				toAppend = row[0]
			} else {
				__antithesis_instrumentation__.Notify(468241)
				toAppend = &tree.DTuple{D: row}
			}
			__antithesis_instrumentation__.Notify(468234)

			size := int64(toAppend.Size())
			alreadyAccountedFor += size
			if err = subqueryResultMemAcc.Grow(ctx, size); err != nil {
				__antithesis_instrumentation__.Notify(468242)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468243)
			}
			__antithesis_instrumentation__.Notify(468235)
			result.D = append(result.D, toAppend)
		}
		__antithesis_instrumentation__.Notify(468227)

		if subqueryPlan.execMode == rowexec.SubqueryExecModeAllRowsNormalized {
			__antithesis_instrumentation__.Notify(468244)

			result.Normalize(&evalCtx.EvalContext)
		} else {
			__antithesis_instrumentation__.Notify(468245)
		}
		__antithesis_instrumentation__.Notify(468228)
		subqueryPlans[planIdx].result = &result
	case rowexec.SubqueryExecModeOneRow:
		__antithesis_instrumentation__.Notify(468229)
		switch rows.Len() {
		case 0:
			__antithesis_instrumentation__.Notify(468246)
			subqueryPlans[planIdx].result = tree.DNull
		case 1:
			__antithesis_instrumentation__.Notify(468247)
			iterator := newRowContainerIterator(ctx, rows, typs)
			defer iterator.Close()
			row, err := iterator.Next()
			if err != nil {
				__antithesis_instrumentation__.Notify(468251)
				return err
			} else {
				__antithesis_instrumentation__.Notify(468252)
			}
			__antithesis_instrumentation__.Notify(468248)
			if row == nil {
				__antithesis_instrumentation__.Notify(468253)
				return errors.AssertionFailedf("iterator didn't return a row although container len is 1")
			} else {
				__antithesis_instrumentation__.Notify(468254)
			}
			__antithesis_instrumentation__.Notify(468249)
			switch row.Len() {
			case 1:
				__antithesis_instrumentation__.Notify(468255)
				subqueryPlans[planIdx].result = row[0]
			default:
				__antithesis_instrumentation__.Notify(468256)
				subqueryPlans[planIdx].result = &tree.DTuple{D: row}
			}
		default:
			__antithesis_instrumentation__.Notify(468250)
			return pgerror.Newf(pgcode.CardinalityViolation,
				"more than one row returned by a subquery used as an expression")
		}
	default:
		__antithesis_instrumentation__.Notify(468230)
		return fmt.Errorf("unexpected subqueryExecMode: %d", subqueryPlan.execMode)
	}
	__antithesis_instrumentation__.Notify(468213)

	actualSize := int64(subqueryPlans[planIdx].result.Size())
	if actualSize >= alreadyAccountedFor {
		__antithesis_instrumentation__.Notify(468257)
		if err := subqueryResultMemAcc.Grow(ctx, actualSize-alreadyAccountedFor); err != nil {
			__antithesis_instrumentation__.Notify(468258)
			return err
		} else {
			__antithesis_instrumentation__.Notify(468259)
		}
	} else {
		__antithesis_instrumentation__.Notify(468260)

		subqueryResultMemAcc.Shrink(ctx, alreadyAccountedFor-actualSize)
	}
	__antithesis_instrumentation__.Notify(468214)
	return nil
}

func (dsp *DistSQLPlanner) PlanAndRun(
	ctx context.Context,
	evalCtx *extendedEvalContext,
	planCtx *PlanningCtx,
	txn *kv.Txn,
	plan planMaybePhysical,
	recv *DistSQLReceiver,
) (cleanup func()) {
	__antithesis_instrumentation__.Notify(468261)
	log.VEventf(ctx, 2, "creating DistSQL plan with isLocal=%v", planCtx.isLocal)

	physPlan, physPlanCleanup, err := dsp.createPhysPlan(ctx, planCtx, plan)
	if err != nil {
		__antithesis_instrumentation__.Notify(468263)
		recv.SetError(err)
		return physPlanCleanup
	} else {
		__antithesis_instrumentation__.Notify(468264)
	}
	__antithesis_instrumentation__.Notify(468262)
	dsp.finalizePlanWithRowCount(planCtx, physPlan, planCtx.planner.curPlan.mainRowCount)
	recv.expectedRowsRead = int64(physPlan.TotalEstimatedScannedRows)
	runCleanup := dsp.Run(ctx, planCtx, txn, physPlan, recv, evalCtx, nil)
	return func() {
		__antithesis_instrumentation__.Notify(468265)
		runCleanup()
		physPlanCleanup()
	}
}

func (dsp *DistSQLPlanner) PlanAndRunCascadesAndChecks(
	ctx context.Context,
	planner *planner,
	evalCtxFactory func() *extendedEvalContext,
	plan *planComponents,
	recv *DistSQLReceiver,
) bool {
	__antithesis_instrumentation__.Notify(468266)
	if len(plan.cascades) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(468273)
		return len(plan.checkPlans) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(468274)
		return false
	} else {
		__antithesis_instrumentation__.Notify(468275)
	}
	__antithesis_instrumentation__.Notify(468267)

	prevSteppingMode := planner.Txn().ConfigureStepping(ctx, kv.SteppingEnabled)
	defer func() {
		__antithesis_instrumentation__.Notify(468276)
		_ = planner.Txn().ConfigureStepping(ctx, prevSteppingMode)
	}()
	__antithesis_instrumentation__.Notify(468268)

	for i := 0; i < len(plan.cascades); i++ {
		__antithesis_instrumentation__.Notify(468277)

		buf := plan.cascades[i].Buffer
		var numBufferedRows int
		if buf != nil {
			__antithesis_instrumentation__.Notify(468286)
			numBufferedRows = buf.(*bufferNode).rows.rows.Len()
			if numBufferedRows == 0 {
				__antithesis_instrumentation__.Notify(468287)

				continue
			} else {
				__antithesis_instrumentation__.Notify(468288)
			}
		} else {
			__antithesis_instrumentation__.Notify(468289)
		}
		__antithesis_instrumentation__.Notify(468278)

		log.VEventf(ctx, 2, "executing cascade for constraint %s", plan.cascades[i].FKName)

		_ = planner.Txn().ConfigureStepping(ctx, kv.SteppingEnabled)
		if err := planner.Txn().Step(ctx); err != nil {
			__antithesis_instrumentation__.Notify(468290)
			recv.SetError(err)
			return false
		} else {
			__antithesis_instrumentation__.Notify(468291)
		}
		__antithesis_instrumentation__.Notify(468279)

		evalCtx := evalCtxFactory()
		execFactory := newExecFactory(planner)

		allowAutoCommit := planner.autoCommit
		if len(plan.checkPlans) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(468292)
			return i < len(plan.cascades)-1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(468293)
			allowAutoCommit = false
		} else {
			__antithesis_instrumentation__.Notify(468294)
		}
		__antithesis_instrumentation__.Notify(468280)
		cascadePlan, err := plan.cascades[i].PlanFn(
			ctx, &planner.semaCtx, &evalCtx.EvalContext, execFactory,
			buf, numBufferedRows, allowAutoCommit,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(468295)
			recv.SetError(err)
			return false
		} else {
			__antithesis_instrumentation__.Notify(468296)
		}
		__antithesis_instrumentation__.Notify(468281)
		cp := cascadePlan.(*planComponents)
		plan.cascades[i].plan = cp.main
		if len(cp.subqueryPlans) > 0 {
			__antithesis_instrumentation__.Notify(468297)
			recv.SetError(errors.AssertionFailedf("cascades should not have subqueries"))
			return false
		} else {
			__antithesis_instrumentation__.Notify(468298)
		}
		__antithesis_instrumentation__.Notify(468282)

		if len(cp.cascades) > 0 {
			__antithesis_instrumentation__.Notify(468299)
			plan.cascades = append(plan.cascades, cp.cascades...)
		} else {
			__antithesis_instrumentation__.Notify(468300)
		}
		__antithesis_instrumentation__.Notify(468283)

		if len(cp.checkPlans) > 0 {
			__antithesis_instrumentation__.Notify(468301)
			plan.checkPlans = append(plan.checkPlans, cp.checkPlans...)
		} else {
			__antithesis_instrumentation__.Notify(468302)
		}
		__antithesis_instrumentation__.Notify(468284)

		if limit := int(evalCtx.SessionData().OptimizerFKCascadesLimit); len(plan.cascades) > limit {
			__antithesis_instrumentation__.Notify(468303)
			telemetry.Inc(sqltelemetry.CascadesLimitReached)
			err := pgerror.Newf(pgcode.TriggeredActionException, "cascades limit (%d) reached", limit)
			recv.SetError(err)
			return false
		} else {
			__antithesis_instrumentation__.Notify(468304)
		}
		__antithesis_instrumentation__.Notify(468285)

		if err := dsp.planAndRunPostquery(
			ctx,
			cp.main,
			planner,
			evalCtx,
			recv,
		); err != nil {
			__antithesis_instrumentation__.Notify(468305)
			recv.SetError(err)
			return false
		} else {
			__antithesis_instrumentation__.Notify(468306)
		}
	}
	__antithesis_instrumentation__.Notify(468269)

	if len(plan.checkPlans) == 0 {
		__antithesis_instrumentation__.Notify(468307)
		return true
	} else {
		__antithesis_instrumentation__.Notify(468308)
	}
	__antithesis_instrumentation__.Notify(468270)

	_ = planner.Txn().ConfigureStepping(ctx, kv.SteppingEnabled)
	if err := planner.Txn().Step(ctx); err != nil {
		__antithesis_instrumentation__.Notify(468309)
		recv.SetError(err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(468310)
	}
	__antithesis_instrumentation__.Notify(468271)

	for i := range plan.checkPlans {
		__antithesis_instrumentation__.Notify(468311)
		log.VEventf(ctx, 2, "executing check query %d out of %d", i+1, len(plan.checkPlans))
		if err := dsp.planAndRunPostquery(
			ctx,
			plan.checkPlans[i].plan,
			planner,
			evalCtxFactory(),
			recv,
		); err != nil {
			__antithesis_instrumentation__.Notify(468312)
			recv.SetError(err)
			return false
		} else {
			__antithesis_instrumentation__.Notify(468313)
		}
	}
	__antithesis_instrumentation__.Notify(468272)

	return true
}

func (dsp *DistSQLPlanner) planAndRunPostquery(
	ctx context.Context,
	postqueryPlan planMaybePhysical,
	planner *planner,
	evalCtx *extendedEvalContext,
	recv *DistSQLReceiver,
) error {
	__antithesis_instrumentation__.Notify(468314)
	postqueryMonitor := mon.NewMonitor(
		"postquery",
		mon.MemoryResource,
		dsp.distSQLSrv.Metrics.CurBytesCount,
		dsp.distSQLSrv.Metrics.MaxBytesHist,
		-1,
		noteworthyMemoryUsageBytes,
		dsp.distSQLSrv.Settings,
	)
	postqueryMonitor.Start(ctx, evalCtx.Mon, mon.BoundAccount{})
	defer postqueryMonitor.Stop(ctx)

	postqueryMemAccount := postqueryMonitor.MakeBoundAccount()
	defer postqueryMemAccount.Close(ctx)

	distributePostquery := getPlanDistribution(
		ctx, planner, planner.execCfg.NodeID, planner.SessionData().DistSQLMode, postqueryPlan,
	).WillDistribute()
	distribute := DistributionType(DistributionTypeNone)
	if distributePostquery {
		__antithesis_instrumentation__.Notify(468318)
		distribute = DistributionTypeSystemTenantOnly
	} else {
		__antithesis_instrumentation__.Notify(468319)
	}
	__antithesis_instrumentation__.Notify(468315)
	postqueryPlanCtx := dsp.NewPlanningCtx(ctx, evalCtx, planner, planner.txn,
		distribute)
	postqueryPlanCtx.stmtType = tree.Rows
	postqueryPlanCtx.ignoreClose = true
	if planner.instrumentation.ShouldSaveFlows() {
		__antithesis_instrumentation__.Notify(468320)
		postqueryPlanCtx.saveFlows = postqueryPlanCtx.getDefaultSaveFlowsFunc(ctx, planner, planComponentTypePostquery)
	} else {
		__antithesis_instrumentation__.Notify(468321)
	}
	__antithesis_instrumentation__.Notify(468316)
	postqueryPlanCtx.traceMetadata = planner.instrumentation.traceMetadata
	postqueryPlanCtx.collectExecStats = planner.instrumentation.ShouldCollectExecStats()

	postqueryPhysPlan, physPlanCleanup, err := dsp.createPhysPlan(ctx, postqueryPlanCtx, postqueryPlan)
	defer physPlanCleanup()
	if err != nil {
		__antithesis_instrumentation__.Notify(468322)
		return err
	} else {
		__antithesis_instrumentation__.Notify(468323)
	}
	__antithesis_instrumentation__.Notify(468317)
	dsp.FinalizePlan(postqueryPlanCtx, postqueryPhysPlan)

	postqueryRecv := recv.clone()
	defer postqueryRecv.Release()

	postqueryResultWriter := &errOnlyResultWriter{}
	postqueryRecv.resultWriter = postqueryResultWriter
	postqueryRecv.batchWriter = postqueryResultWriter
	dsp.Run(ctx, postqueryPlanCtx, planner.txn, postqueryPhysPlan, postqueryRecv, evalCtx, nil)()
	return postqueryRecv.resultWriter.Err()
}
