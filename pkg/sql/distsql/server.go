package distsql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/colflow"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/faketreeeval"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowflow"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

const minFlowDrainWait = 1 * time.Second

const MultiTenancyIssueNo = 47900

var noteworthyMemoryUsageBytes = envutil.EnvOrDefaultInt64("COCKROACH_NOTEWORTHY_DISTSQL_MEMORY_USAGE", 1024*1024)

type ServerImpl struct {
	execinfra.ServerConfig
	flowRegistry  *flowinfra.FlowRegistry
	flowScheduler *flowinfra.FlowScheduler
	memMonitor    *mon.BytesMonitor
	regexpCache   *tree.RegexpCache
}

var _ execinfrapb.DistSQLServer = &ServerImpl{}

func NewServer(
	ctx context.Context, cfg execinfra.ServerConfig, flowScheduler *flowinfra.FlowScheduler,
) *ServerImpl {
	__antithesis_instrumentation__.Notify(466211)
	ds := &ServerImpl{
		ServerConfig:  cfg,
		regexpCache:   tree.NewRegexpCache(512),
		flowRegistry:  flowinfra.NewFlowRegistry(),
		flowScheduler: flowScheduler,
		memMonitor: mon.NewMonitor(
			"distsql",
			mon.MemoryResource,
			cfg.Metrics.CurBytesCount,
			cfg.Metrics.MaxBytesHist,
			-1,
			noteworthyMemoryUsageBytes,
			cfg.Settings,
		),
	}
	ds.memMonitor.Start(ctx, cfg.ParentMemoryMonitor, mon.BoundAccount{})

	ds.flowScheduler.Init(ds.Metrics)

	return ds
}

func (ds *ServerImpl) Start() {
	__antithesis_instrumentation__.Notify(466212)

	if g, ok := ds.ServerConfig.Gossip.Optional(MultiTenancyIssueNo); ok {
		__antithesis_instrumentation__.Notify(466215)
		if nodeID, ok := ds.ServerConfig.NodeID.OptionalNodeID(); ok {
			__antithesis_instrumentation__.Notify(466216)
			if err := g.AddInfoProto(
				gossip.MakeDistSQLNodeVersionKey(base.SQLInstanceID(nodeID)),
				&execinfrapb.DistSQLVersionGossipInfo{
					Version:            execinfra.Version,
					MinAcceptedVersion: execinfra.MinAcceptedVersion,
				},
				0,
			); err != nil {
				__antithesis_instrumentation__.Notify(466217)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(466218)
			}
		} else {
			__antithesis_instrumentation__.Notify(466219)
		}
	} else {
		__antithesis_instrumentation__.Notify(466220)
	}
	__antithesis_instrumentation__.Notify(466213)

	if err := ds.setDraining(false); err != nil {
		__antithesis_instrumentation__.Notify(466221)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(466222)
	}
	__antithesis_instrumentation__.Notify(466214)

	ds.flowScheduler.Start()
}

func (ds *ServerImpl) NumRemoteFlowsInQueue() int {
	__antithesis_instrumentation__.Notify(466223)
	return ds.flowScheduler.NumFlowsInQueue()
}

func (ds *ServerImpl) NumRemoteRunningFlows() int {
	__antithesis_instrumentation__.Notify(466224)
	return ds.flowScheduler.NumRunningFlows()
}

func (ds *ServerImpl) SetCancelDeadFlowsCallback(cb func(int)) {
	__antithesis_instrumentation__.Notify(466225)
	ds.flowScheduler.TestingKnobs.CancelDeadFlowsCallback = cb
}

func (ds *ServerImpl) Drain(
	ctx context.Context, flowDrainWait time.Duration, reporter func(int, redact.SafeString),
) {
	__antithesis_instrumentation__.Notify(466226)
	if err := ds.setDraining(true); err != nil {
		__antithesis_instrumentation__.Notify(466229)
		log.Warningf(ctx, "unable to gossip distsql draining state: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(466230)
	}
	__antithesis_instrumentation__.Notify(466227)

	flowWait := flowDrainWait
	minWait := minFlowDrainWait
	if ds.ServerConfig.TestingKnobs.DrainFast {
		__antithesis_instrumentation__.Notify(466231)
		flowWait = 0
		minWait = 0
	} else {
		__antithesis_instrumentation__.Notify(466232)
		if g, ok := ds.Gossip.Optional(MultiTenancyIssueNo); !ok || func() bool {
			__antithesis_instrumentation__.Notify(466233)
			return len(g.Outgoing()) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(466234)

			minWait = 0
		} else {
			__antithesis_instrumentation__.Notify(466235)
		}
	}
	__antithesis_instrumentation__.Notify(466228)
	ds.flowRegistry.Drain(flowWait, minWait, reporter)
}

func (ds *ServerImpl) setDraining(drain bool) error {
	__antithesis_instrumentation__.Notify(466236)
	nodeID, ok := ds.ServerConfig.NodeID.OptionalNodeID()
	if !ok {
		__antithesis_instrumentation__.Notify(466239)

		_ = MultiTenancyIssueNo
		return nil
	} else {
		__antithesis_instrumentation__.Notify(466240)
	}
	__antithesis_instrumentation__.Notify(466237)
	if g, ok := ds.ServerConfig.Gossip.Optional(MultiTenancyIssueNo); ok {
		__antithesis_instrumentation__.Notify(466241)
		return g.AddInfoProto(
			gossip.MakeDistSQLDrainingKey(base.SQLInstanceID(nodeID)),
			&execinfrapb.DistSQLDrainingInfo{
				Draining: drain,
			},
			0,
		)
	} else {
		__antithesis_instrumentation__.Notify(466242)
	}
	__antithesis_instrumentation__.Notify(466238)
	return nil
}

func FlowVerIsCompatible(
	flowVer, minAcceptedVersion, serverVersion execinfrapb.DistSQLVersion,
) bool {
	__antithesis_instrumentation__.Notify(466243)
	return flowVer >= minAcceptedVersion && func() bool {
		__antithesis_instrumentation__.Notify(466244)
		return flowVer <= serverVersion == true
	}() == true
}

func (ds *ServerImpl) setupFlow(
	ctx context.Context,
	parentSpan *tracing.Span,
	parentMonitor *mon.BytesMonitor,
	req *execinfrapb.SetupFlowRequest,
	rowSyncFlowConsumer execinfra.RowReceiver,
	batchSyncFlowConsumer execinfra.BatchReceiver,
	localState LocalState,
) (retCtx context.Context, _ flowinfra.Flow, _ execinfra.OpChains, retErr error) {
	__antithesis_instrumentation__.Notify(466245)
	if !FlowVerIsCompatible(req.Version, execinfra.MinAcceptedVersion, execinfra.Version) {
		__antithesis_instrumentation__.Notify(466257)
		err := errors.Errorf(
			"version mismatch in flow request: %d; this node accepts %d through %d",
			req.Version, execinfra.MinAcceptedVersion, execinfra.Version,
		)
		log.Warningf(ctx, "%v", err)
		return ctx, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(466258)
	}
	__antithesis_instrumentation__.Notify(466246)

	var sp *tracing.Span
	var monitor *mon.BytesMonitor
	var onFlowCleanup func()

	defer func() {
		__antithesis_instrumentation__.Notify(466259)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(466260)
			if sp != nil {
				__antithesis_instrumentation__.Notify(466264)
				sp.Finish()
			} else {
				__antithesis_instrumentation__.Notify(466265)
			}
			__antithesis_instrumentation__.Notify(466261)
			if monitor != nil {
				__antithesis_instrumentation__.Notify(466266)
				monitor.Stop(ctx)
			} else {
				__antithesis_instrumentation__.Notify(466267)
			}
			__antithesis_instrumentation__.Notify(466262)
			if onFlowCleanup != nil {
				__antithesis_instrumentation__.Notify(466268)
				onFlowCleanup()
			} else {
				__antithesis_instrumentation__.Notify(466269)
			}
			__antithesis_instrumentation__.Notify(466263)
			retCtx = tracing.ContextWithSpan(ctx, nil)
		} else {
			__antithesis_instrumentation__.Notify(466270)
		}
	}()
	__antithesis_instrumentation__.Notify(466247)

	const opName = "flow"
	if parentSpan == nil {
		__antithesis_instrumentation__.Notify(466271)
		ctx, sp = ds.Tracer.StartSpanCtx(ctx, opName)
	} else {
		__antithesis_instrumentation__.Notify(466272)
		if localState.IsLocal {
			__antithesis_instrumentation__.Notify(466273)

			ctx, sp = ds.Tracer.StartSpanCtx(ctx, opName, tracing.WithParent(parentSpan))
		} else {
			__antithesis_instrumentation__.Notify(466274)

			ctx, sp = ds.Tracer.StartSpanCtx(
				ctx,
				opName,
				tracing.WithParent(parentSpan),
				tracing.WithFollowsFrom(),
			)
		}
	}
	__antithesis_instrumentation__.Notify(466248)

	monitor = mon.NewMonitor(
		"flow",
		mon.MemoryResource,
		ds.Metrics.CurBytesCount,
		ds.Metrics.MaxBytesHist,
		-1,
		noteworthyMemoryUsageBytes,
		ds.Settings,
	)
	monitor.Start(ctx, parentMonitor, mon.BoundAccount{})

	makeLeaf := func(req *execinfrapb.SetupFlowRequest) (*kv.Txn, error) {
		__antithesis_instrumentation__.Notify(466275)
		tis := req.LeafTxnInputState
		if tis == nil {
			__antithesis_instrumentation__.Notify(466278)

			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(466279)
		}
		__antithesis_instrumentation__.Notify(466276)
		if tis.Txn.Status != roachpb.PENDING {
			__antithesis_instrumentation__.Notify(466280)
			return nil, errors.AssertionFailedf("cannot create flow in non-PENDING txn: %s",
				tis.Txn)
		} else {
			__antithesis_instrumentation__.Notify(466281)
		}
		__antithesis_instrumentation__.Notify(466277)

		return kv.NewLeafTxn(ctx, ds.DB, roachpb.NodeID(req.Flow.Gateway), tis), nil
	}
	__antithesis_instrumentation__.Notify(466249)

	var evalCtx *tree.EvalContext
	var leafTxn *kv.Txn
	if localState.EvalContext != nil {
		__antithesis_instrumentation__.Notify(466282)
		evalCtx = localState.EvalContext

		origMon := evalCtx.Mon
		origTxn := evalCtx.Txn
		onFlowCleanup = func() {
			__antithesis_instrumentation__.Notify(466284)
			evalCtx.Mon = origMon
			evalCtx.Txn = origTxn
		}
		__antithesis_instrumentation__.Notify(466283)
		evalCtx.Mon = monitor
		if localState.HasConcurrency {
			__antithesis_instrumentation__.Notify(466285)
			var err error
			leafTxn, err = makeLeaf(req)
			if err != nil {
				__antithesis_instrumentation__.Notify(466287)
				return nil, nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(466288)
			}
			__antithesis_instrumentation__.Notify(466286)

			evalCtx.Txn = leafTxn
		} else {
			__antithesis_instrumentation__.Notify(466289)
		}
	} else {
		__antithesis_instrumentation__.Notify(466290)
		if localState.IsLocal {
			__antithesis_instrumentation__.Notify(466294)
			return nil, nil, nil, errors.AssertionFailedf(
				"EvalContext expected to be populated when IsLocal is set")
		} else {
			__antithesis_instrumentation__.Notify(466295)
		}
		__antithesis_instrumentation__.Notify(466291)

		sd, err := sessiondata.UnmarshalNonLocal(req.EvalContext.SessionData)
		if err != nil {
			__antithesis_instrumentation__.Notify(466296)
			return ctx, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(466297)
		}
		__antithesis_instrumentation__.Notify(466292)

		leafTxn, err = makeLeaf(req)
		if err != nil {
			__antithesis_instrumentation__.Notify(466298)
			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(466299)
		}
		__antithesis_instrumentation__.Notify(466293)
		evalCtx = &tree.EvalContext{
			Settings:         ds.ServerConfig.Settings,
			SessionDataStack: sessiondata.NewStack(sd),
			ClusterID:        ds.ServerConfig.LogicalClusterID.Get(),
			ClusterName:      ds.ServerConfig.ClusterName,
			NodeID:           ds.ServerConfig.NodeID,
			Codec:            ds.ServerConfig.Codec,
			ReCache:          ds.regexpCache,
			Mon:              monitor,
			Locality:         ds.ServerConfig.Locality,
			Tracer:           ds.ServerConfig.Tracer,

			Context:                   ctx,
			Planner:                   &faketreeeval.DummyEvalPlanner{},
			PrivilegedAccessor:        &faketreeeval.DummyPrivilegedAccessor{},
			SessionAccessor:           &faketreeeval.DummySessionAccessor{},
			ClientNoticeSender:        &faketreeeval.DummyClientNoticeSender{},
			Sequence:                  &faketreeeval.DummySequenceOperators{},
			Tenant:                    &faketreeeval.DummyTenantOperator{},
			Regions:                   &faketreeeval.DummyRegionOperator{},
			Txn:                       leafTxn,
			SQLLivenessReader:         ds.ServerConfig.SQLLivenessReader,
			SQLStatsController:        ds.ServerConfig.SQLStatsController,
			IndexUsageStatsController: ds.ServerConfig.IndexUsageStatsController,
		}
		evalCtx.SetStmtTimestamp(timeutil.Unix(0, req.EvalContext.StmtTimestampNanos))
		evalCtx.SetTxnTimestamp(timeutil.Unix(0, req.EvalContext.TxnTimestampNanos))
	}
	__antithesis_instrumentation__.Notify(466250)

	flowCtx := ds.newFlowContext(
		ctx, req.Flow.FlowID, evalCtx, req.TraceKV, req.CollectStats, localState, req.Flow.Gateway == ds.NodeID.SQLInstanceID(),
	)

	isVectorized := req.EvalContext.SessionData.VectorizeMode != sessiondatapb.VectorizeOff
	f := newFlow(
		flowCtx, sp, ds.flowRegistry, rowSyncFlowConsumer, batchSyncFlowConsumer,
		localState.LocalProcs, isVectorized, onFlowCleanup, req.StatementSQL,
	)
	opt := flowinfra.FuseNormally
	if !localState.MustUseLeafTxn() {
		__antithesis_instrumentation__.Notify(466300)

		opt = flowinfra.FuseAggressively
	} else {
		__antithesis_instrumentation__.Notify(466301)
	}
	__antithesis_instrumentation__.Notify(466251)

	var opChains execinfra.OpChains
	var err error
	ctx, opChains, err = f.Setup(ctx, &req.Flow, opt)
	if err != nil {
		__antithesis_instrumentation__.Notify(466302)
		log.Errorf(ctx, "error setting up flow: %s", err)
		return ctx, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(466303)
	}
	__antithesis_instrumentation__.Notify(466252)
	if !f.IsLocal() {
		__antithesis_instrumentation__.Notify(466304)
		flowCtx.AmbientContext.AddLogTag("f", f.GetFlowCtx().ID.Short())
		ctx = flowCtx.AmbientContext.AnnotateCtx(ctx)
		telemetry.Inc(sqltelemetry.DistSQLExecCounter)
	} else {
		__antithesis_instrumentation__.Notify(466305)
	}
	__antithesis_instrumentation__.Notify(466253)
	if f.IsVectorized() {
		__antithesis_instrumentation__.Notify(466306)
		telemetry.Inc(sqltelemetry.VecExecCounter)
	} else {
		__antithesis_instrumentation__.Notify(466307)
	}
	__antithesis_instrumentation__.Notify(466254)

	useLeaf := false
	if req.LeafTxnInputState != nil && func() bool {
		__antithesis_instrumentation__.Notify(466308)
		return row.CanUseStreamer(ctx, ds.Settings) == true
	}() == true {
		__antithesis_instrumentation__.Notify(466309)
		for _, proc := range req.Flow.Processors {
			__antithesis_instrumentation__.Notify(466310)
			if jr := proc.Core.JoinReader; jr != nil {
				__antithesis_instrumentation__.Notify(466311)
				if jr.IsIndexJoin() {
					__antithesis_instrumentation__.Notify(466312)

					useLeaf = true
					break
				} else {
					__antithesis_instrumentation__.Notify(466313)
				}
			} else {
				__antithesis_instrumentation__.Notify(466314)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(466315)
	}
	__antithesis_instrumentation__.Notify(466255)
	var txn *kv.Txn
	if localState.IsLocal && func() bool {
		__antithesis_instrumentation__.Notify(466316)
		return !f.ConcurrentTxnUse() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(466317)
		return !useLeaf == true
	}() == true {
		__antithesis_instrumentation__.Notify(466318)
		txn = localState.Txn
	} else {
		__antithesis_instrumentation__.Notify(466319)

		if leafTxn == nil {
			__antithesis_instrumentation__.Notify(466321)
			leafTxn, err = makeLeaf(req)
			if err != nil {
				__antithesis_instrumentation__.Notify(466322)
				return nil, nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(466323)
			}
		} else {
			__antithesis_instrumentation__.Notify(466324)
		}
		__antithesis_instrumentation__.Notify(466320)
		txn = leafTxn
	}
	__antithesis_instrumentation__.Notify(466256)

	f.SetTxn(txn)

	return ctx, f, opChains, nil
}

func (ds *ServerImpl) newFlowContext(
	ctx context.Context,
	id execinfrapb.FlowID,
	evalCtx *tree.EvalContext,
	traceKV bool,
	collectStats bool,
	localState LocalState,
	isGatewayNode bool,
) execinfra.FlowCtx {
	__antithesis_instrumentation__.Notify(466325)

	flowCtx := execinfra.FlowCtx{
		AmbientContext: ds.AmbientContext,
		Cfg:            &ds.ServerConfig,
		ID:             id,
		EvalCtx:        evalCtx,
		Txn:            evalCtx.Txn,
		NodeID:         ds.ServerConfig.NodeID,
		TraceKV:        traceKV,
		CollectStats:   collectStats,
		Local:          localState.IsLocal,
		Gateway:        isGatewayNode,

		DiskMonitor: execinfra.NewMonitor(
			ctx, ds.ParentDiskMonitor, "flow-disk-monitor",
		),
		PreserveFlowSpecs: localState.PreserveFlowSpecs,
	}

	if localState.IsLocal && func() bool {
		__antithesis_instrumentation__.Notify(466327)
		return localState.Collection != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(466328)

		flowCtx.Descriptors = localState.Collection
	} else {
		__antithesis_instrumentation__.Notify(466329)

		flowCtx.Descriptors = ds.CollectionFactory.NewCollection(ctx, descs.NewTemporarySchemaProvider(evalCtx.SessionDataStack))
		flowCtx.IsDescriptorsCleanupRequired = true
	}
	__antithesis_instrumentation__.Notify(466326)
	return flowCtx
}

func newFlow(
	flowCtx execinfra.FlowCtx,
	sp *tracing.Span,
	flowReg *flowinfra.FlowRegistry,
	rowSyncFlowConsumer execinfra.RowReceiver,
	batchSyncFlowConsumer execinfra.BatchReceiver,
	localProcessors []execinfra.LocalProcessor,
	isVectorized bool,
	onFlowCleanup func(),
	statementSQL string,
) flowinfra.Flow {
	__antithesis_instrumentation__.Notify(466330)
	base := flowinfra.NewFlowBase(flowCtx, sp, flowReg, rowSyncFlowConsumer, batchSyncFlowConsumer, localProcessors, onFlowCleanup, statementSQL)
	if isVectorized {
		__antithesis_instrumentation__.Notify(466332)
		return colflow.NewVectorizedFlow(base)
	} else {
		__antithesis_instrumentation__.Notify(466333)
	}
	__antithesis_instrumentation__.Notify(466331)
	return rowflow.NewRowBasedFlow(base)
}

type LocalState struct {
	EvalContext *tree.EvalContext

	Collection *descs.Collection

	IsLocal bool

	HasConcurrency bool

	Txn *kv.Txn

	LocalProcs []execinfra.LocalProcessor

	PreserveFlowSpecs bool
}

func (l LocalState) MustUseLeafTxn() bool {
	__antithesis_instrumentation__.Notify(466334)
	return !l.IsLocal || func() bool {
		__antithesis_instrumentation__.Notify(466335)
		return l.HasConcurrency == true
	}() == true
}

func (ds *ServerImpl) SetupLocalSyncFlow(
	ctx context.Context,
	parentMonitor *mon.BytesMonitor,
	req *execinfrapb.SetupFlowRequest,
	output execinfra.RowReceiver,
	batchOutput execinfra.BatchReceiver,
	localState LocalState,
) (context.Context, flowinfra.Flow, execinfra.OpChains, error) {
	__antithesis_instrumentation__.Notify(466336)
	ctx, f, opChains, err := ds.setupFlow(
		ctx, tracing.SpanFromContext(ctx), parentMonitor, req, output, batchOutput, localState,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(466338)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(466339)
	}
	__antithesis_instrumentation__.Notify(466337)
	return ctx, f, opChains, err
}

func (ds *ServerImpl) setupSpanForIncomingRPC(
	ctx context.Context, req *execinfrapb.SetupFlowRequest,
) (context.Context, *tracing.Span) {
	__antithesis_instrumentation__.Notify(466340)
	tr := ds.ServerConfig.AmbientContext.Tracer
	parentSpan := tracing.SpanFromContext(ctx)
	if parentSpan != nil {
		__antithesis_instrumentation__.Notify(466344)

		return tr.StartSpanCtx(ctx, tracing.SetupFlowMethodName,
			tracing.WithParent(parentSpan),
			tracing.WithServerSpanKind)
	} else {
		__antithesis_instrumentation__.Notify(466345)
	}
	__antithesis_instrumentation__.Notify(466341)

	if !req.TraceInfo.Empty() {
		__antithesis_instrumentation__.Notify(466346)
		return tr.StartSpanCtx(ctx, tracing.SetupFlowMethodName,
			tracing.WithRemoteParentFromTraceInfo(&req.TraceInfo),
			tracing.WithServerSpanKind)
	} else {
		__antithesis_instrumentation__.Notify(466347)
	}
	__antithesis_instrumentation__.Notify(466342)

	remoteParent, err := tracing.ExtractSpanMetaFromGRPCCtx(ctx, tr)
	if err != nil {
		__antithesis_instrumentation__.Notify(466348)
		log.Warningf(ctx, "error extracting tracing info from gRPC: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(466349)
	}
	__antithesis_instrumentation__.Notify(466343)
	return tr.StartSpanCtx(ctx, tracing.SetupFlowMethodName,
		tracing.WithRemoteParentFromSpanMeta(remoteParent),
		tracing.WithServerSpanKind)
}

func (ds *ServerImpl) SetupFlow(
	ctx context.Context, req *execinfrapb.SetupFlowRequest,
) (*execinfrapb.SimpleResponse, error) {
	__antithesis_instrumentation__.Notify(466350)
	log.VEventf(ctx, 1, "received SetupFlow request from n%v for flow %v", req.Flow.Gateway, req.Flow.FlowID)
	_, rpcSpan := ds.setupSpanForIncomingRPC(ctx, req)
	defer rpcSpan.Finish()

	ctx = ds.AnnotateCtx(context.Background())
	ctx, f, _, err := ds.setupFlow(
		ctx, rpcSpan, ds.memMonitor, req, nil,
		nil, LocalState{},
	)
	if err == nil {
		__antithesis_instrumentation__.Notify(466353)
		err = ds.flowScheduler.ScheduleFlow(ctx, f)
	} else {
		__antithesis_instrumentation__.Notify(466354)
	}
	__antithesis_instrumentation__.Notify(466351)
	if err != nil {
		__antithesis_instrumentation__.Notify(466355)

		return &execinfrapb.SimpleResponse{Error: execinfrapb.NewError(ctx, err)}, nil
	} else {
		__antithesis_instrumentation__.Notify(466356)
	}
	__antithesis_instrumentation__.Notify(466352)
	return &execinfrapb.SimpleResponse{}, nil
}

func (ds *ServerImpl) CancelDeadFlows(
	_ context.Context, req *execinfrapb.CancelDeadFlowsRequest,
) (*execinfrapb.SimpleResponse, error) {
	__antithesis_instrumentation__.Notify(466357)
	ds.flowScheduler.CancelDeadFlows(req)
	return &execinfrapb.SimpleResponse{}, nil
}

func (ds *ServerImpl) flowStreamInt(
	ctx context.Context, stream execinfrapb.DistSQL_FlowStreamServer,
) error {
	__antithesis_instrumentation__.Notify(466358)

	msg, err := stream.Recv()
	if err != nil {
		__antithesis_instrumentation__.Notify(466363)
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(466365)
			return errors.AssertionFailedf("missing header message")
		} else {
			__antithesis_instrumentation__.Notify(466366)
		}
		__antithesis_instrumentation__.Notify(466364)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466367)
	}
	__antithesis_instrumentation__.Notify(466359)
	if msg.Header == nil {
		__antithesis_instrumentation__.Notify(466368)
		return errors.AssertionFailedf("no header in first message")
	} else {
		__antithesis_instrumentation__.Notify(466369)
	}
	__antithesis_instrumentation__.Notify(466360)
	flowID := msg.Header.FlowID
	streamID := msg.Header.StreamID
	if log.V(1) {
		__antithesis_instrumentation__.Notify(466370)
		log.Infof(ctx, "connecting inbound stream %s/%d", flowID.Short(), streamID)
	} else {
		__antithesis_instrumentation__.Notify(466371)
	}
	__antithesis_instrumentation__.Notify(466361)
	f, streamStrategy, cleanup, err := ds.flowRegistry.ConnectInboundStream(
		ctx, flowID, streamID, stream, flowinfra.SettingFlowStreamTimeout.Get(&ds.Settings.SV),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(466372)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466373)
	}
	__antithesis_instrumentation__.Notify(466362)
	defer cleanup()
	log.VEventf(ctx, 1, "connected inbound stream %s/%d", flowID.Short(), streamID)
	return streamStrategy.Run(f.AmbientContext.AnnotateCtx(ctx), stream, msg, f)
}

func (ds *ServerImpl) FlowStream(stream execinfrapb.DistSQL_FlowStreamServer) error {
	__antithesis_instrumentation__.Notify(466374)
	ctx := ds.AnnotateCtx(stream.Context())
	err := ds.flowStreamInt(ctx, stream)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(466376)
		return log.V(2) == true
	}() == true {
		__antithesis_instrumentation__.Notify(466377)

		log.Infof(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(466378)
	}
	__antithesis_instrumentation__.Notify(466375)
	return err
}

type lazyInternalExecutor struct {
	sqlutil.InternalExecutor

	once sync.Once

	newInternalExecutor func() sqlutil.InternalExecutor
}

var _ sqlutil.InternalExecutor = &lazyInternalExecutor{}

func (ie *lazyInternalExecutor) QueryRowEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	opts sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(466379)
	ie.once.Do(func() {
		__antithesis_instrumentation__.Notify(466381)
		ie.InternalExecutor = ie.newInternalExecutor()
	})
	__antithesis_instrumentation__.Notify(466380)
	return ie.InternalExecutor.QueryRowEx(ctx, opName, txn, opts, stmt, qargs...)
}

func (ie *lazyInternalExecutor) QueryRow(
	ctx context.Context, opName string, txn *kv.Txn, stmt string, qargs ...interface{},
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(466382)
	ie.once.Do(func() {
		__antithesis_instrumentation__.Notify(466384)
		ie.InternalExecutor = ie.newInternalExecutor()
	})
	__antithesis_instrumentation__.Notify(466383)
	return ie.InternalExecutor.QueryRow(ctx, opName, txn, stmt, qargs...)
}
