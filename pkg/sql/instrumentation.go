package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/execstats"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec/explain"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessionphase"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/stmtdiagnostics"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/errors"
)

var collectTxnStatsSampleRate = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"sql.txn_stats.sample_rate",
	"the probability that a given transaction will collect execution statistics (displayed in the DB Console)",
	0.01,
	func(f float64) error {
		__antithesis_instrumentation__.Notify(497576)
		if f < 0 || func() bool {
			__antithesis_instrumentation__.Notify(497578)
			return f > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(497579)
			return errors.New("value must be between 0 and 1 inclusive")
		} else {
			__antithesis_instrumentation__.Notify(497580)
		}
		__antithesis_instrumentation__.Notify(497577)
		return nil
	},
)

type instrumentationHelper struct {
	outputMode outputMode

	explainFlags explain.Flags

	fingerprint string
	implicitTxn bool
	codec       keys.SQLCodec

	collectBundle bool

	collectExecStats bool

	discardRows bool

	diagRequestID           stmtdiagnostics.RequestID
	diagRequest             stmtdiagnostics.Request
	stmtDiagnosticsRecorder *stmtdiagnostics.Registry
	withStatementTrace      func(trace tracing.Recording, stmt string)

	sp *tracing.Span

	shouldFinishSpan bool
	origCtx          context.Context
	evalCtx          *tree.EvalContext

	savePlanForStats bool

	explainPlan  *explain.Plan
	distribution physicalplan.PlanDistribution
	vectorized   bool

	traceMetadata execNodeTraceMetadata

	regions []string

	planGist explain.PlanGist

	costEstimate float64

	indexRecommendations []string
}

type outputMode int8

const (
	unmodifiedOutput outputMode = iota
	explainAnalyzeDebugOutput
	explainAnalyzePlanOutput
	explainAnalyzeDistSQLOutput
)

func (ih *instrumentationHelper) SetOutputMode(outputMode outputMode, explainFlags explain.Flags) {
	__antithesis_instrumentation__.Notify(497581)
	ih.outputMode = outputMode
	ih.explainFlags = explainFlags
}

func (ih *instrumentationHelper) Setup(
	ctx context.Context,
	cfg *ExecutorConfig,
	statsCollector sqlstats.StatsCollector,
	p *planner,
	stmtDiagnosticsRecorder *stmtdiagnostics.Registry,
	fingerprint string,
	implicitTxn bool,
	collectTxnExecStats bool,
) (newCtx context.Context, needFinish bool) {
	__antithesis_instrumentation__.Notify(497582)
	ih.fingerprint = fingerprint
	ih.implicitTxn = implicitTxn
	ih.codec = cfg.Codec
	ih.origCtx = ctx

	switch ih.outputMode {
	case explainAnalyzeDebugOutput:
		__antithesis_instrumentation__.Notify(497587)
		ih.collectBundle = true

		ih.discardRows = true

	case explainAnalyzePlanOutput, explainAnalyzeDistSQLOutput:
		__antithesis_instrumentation__.Notify(497588)
		ih.discardRows = true

	default:
		__antithesis_instrumentation__.Notify(497589)
		ih.collectBundle, ih.diagRequestID, ih.diagRequest =
			stmtDiagnosticsRecorder.ShouldCollectDiagnostics(ctx, fingerprint)
	}
	__antithesis_instrumentation__.Notify(497583)

	ih.stmtDiagnosticsRecorder = stmtDiagnosticsRecorder
	ih.withStatementTrace = cfg.TestingKnobs.WithStatementTrace

	ih.savePlanForStats =
		statsCollector.ShouldSaveLogicalPlanDesc(fingerprint, implicitTxn, p.SessionData().Database)

	if sp := tracing.SpanFromContext(ctx); sp != nil {
		__antithesis_instrumentation__.Notify(497590)
		if sp.IsVerbose() {
			__antithesis_instrumentation__.Notify(497591)

			ih.collectExecStats = true

			ih.sp = sp
			return ctx, ih.collectBundle
		} else {
			__antithesis_instrumentation__.Notify(497592)
		}
	} else {
		__antithesis_instrumentation__.Notify(497593)
		if buildutil.CrdbTestBuild {
			__antithesis_instrumentation__.Notify(497594)
			panic(errors.AssertionFailedf("the context doesn't have a tracing span"))
		} else {
			__antithesis_instrumentation__.Notify(497595)
		}
	}
	__antithesis_instrumentation__.Notify(497584)

	ih.collectExecStats = collectTxnExecStats

	if !collectTxnExecStats && func() bool {
		__antithesis_instrumentation__.Notify(497596)
		return ih.savePlanForStats == true
	}() == true {
		__antithesis_instrumentation__.Notify(497597)

		statsCollectionDisabled := collectTxnStatsSampleRate.Get(&cfg.Settings.SV) == 0
		ih.collectExecStats = !statsCollectionDisabled
	} else {
		__antithesis_instrumentation__.Notify(497598)
	}
	__antithesis_instrumentation__.Notify(497585)

	if !ih.collectBundle && func() bool {
		__antithesis_instrumentation__.Notify(497599)
		return ih.withStatementTrace == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(497600)
		return ih.outputMode == unmodifiedOutput == true
	}() == true {
		__antithesis_instrumentation__.Notify(497601)
		if ih.collectExecStats {
			__antithesis_instrumentation__.Notify(497603)

			newCtx, ih.sp = tracing.EnsureChildSpan(ctx, cfg.AmbientCtx.Tracer, "traced statement",
				tracing.WithRecording(tracing.RecordingStructured))
			ih.shouldFinishSpan = true
			return newCtx, true
		} else {
			__antithesis_instrumentation__.Notify(497604)
		}
		__antithesis_instrumentation__.Notify(497602)
		return ctx, false
	} else {
		__antithesis_instrumentation__.Notify(497605)
	}
	__antithesis_instrumentation__.Notify(497586)

	ih.collectExecStats = true
	ih.traceMetadata = make(execNodeTraceMetadata)
	ih.evalCtx = p.EvalContext()
	newCtx, ih.sp = tracing.EnsureChildSpan(ctx, cfg.AmbientCtx.Tracer, "traced statement", tracing.WithRecording(tracing.RecordingVerbose))
	ih.shouldFinishSpan = true
	return newCtx, true
}

func (ih *instrumentationHelper) Finish(
	cfg *ExecutorConfig,
	statsCollector sqlstats.StatsCollector,
	txnStats *execstats.QueryLevelStats,
	collectTxnExecStats bool,
	p *planner,
	ast tree.Statement,
	stmtRawSQL string,
	res RestrictedCommandResult,
	retErr error,
) error {
	__antithesis_instrumentation__.Notify(497606)
	ctx := ih.origCtx
	if ih.sp == nil {
		__antithesis_instrumentation__.Notify(497615)
		return retErr
	} else {
		__antithesis_instrumentation__.Notify(497616)
	}
	__antithesis_instrumentation__.Notify(497607)

	var trace tracing.Recording
	if ih.shouldFinishSpan {
		__antithesis_instrumentation__.Notify(497617)
		trace = ih.sp.FinishAndGetConfiguredRecording()
	} else {
		__antithesis_instrumentation__.Notify(497618)
		trace = ih.sp.GetConfiguredRecording()
	}
	__antithesis_instrumentation__.Notify(497608)

	if ih.withStatementTrace != nil {
		__antithesis_instrumentation__.Notify(497619)
		ih.withStatementTrace(trace, stmtRawSQL)
	} else {
		__antithesis_instrumentation__.Notify(497620)
	}
	__antithesis_instrumentation__.Notify(497609)

	if ih.traceMetadata != nil && func() bool {
		__antithesis_instrumentation__.Notify(497621)
		return ih.explainPlan != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(497622)
		ih.regions = ih.traceMetadata.annotateExplain(
			ih.explainPlan,
			trace,
			cfg.TestingKnobs.DeterministicExplain,
			p,
		)
	} else {
		__antithesis_instrumentation__.Notify(497623)
	}
	__antithesis_instrumentation__.Notify(497610)

	var flowsMetadata []*execstats.FlowsMetadata
	for _, flowInfo := range p.curPlan.distSQLFlowInfos {
		__antithesis_instrumentation__.Notify(497624)
		flowsMetadata = append(flowsMetadata, flowInfo.flowsMetadata)
	}
	__antithesis_instrumentation__.Notify(497611)
	queryLevelStats, err := execstats.GetQueryLevelStats(trace, cfg.TestingKnobs.DeterministicExplain, flowsMetadata)
	if err != nil {
		__antithesis_instrumentation__.Notify(497625)
		const msg = "error getting query level stats for statement: %s: %+v"
		if buildutil.CrdbTestBuild {
			__antithesis_instrumentation__.Notify(497627)
			panic(fmt.Sprintf(msg, ih.fingerprint, err))
		} else {
			__antithesis_instrumentation__.Notify(497628)
		}
		__antithesis_instrumentation__.Notify(497626)
		log.VInfof(ctx, 1, msg, ih.fingerprint, err)
	} else {
		__antithesis_instrumentation__.Notify(497629)
		stmtStatsKey := roachpb.StatementStatisticsKey{
			Query:       ih.fingerprint,
			ImplicitTxn: ih.implicitTxn,
			Database:    p.SessionData().Database,
			Failed:      retErr != nil,
			PlanHash:    ih.planGist.Hash(),
		}

		if ih.implicitTxn {
			__antithesis_instrumentation__.Notify(497632)
			stmtFingerprintID := stmtStatsKey.FingerprintID()
			txnFingerprintHash := util.MakeFNV64()
			txnFingerprintHash.Add(uint64(stmtFingerprintID))
			stmtStatsKey.TransactionFingerprintID =
				roachpb.TransactionFingerprintID(txnFingerprintHash.Sum())
		} else {
			__antithesis_instrumentation__.Notify(497633)
		}
		__antithesis_instrumentation__.Notify(497630)
		err = statsCollector.RecordStatementExecStats(stmtStatsKey, queryLevelStats)
		if err != nil {
			__antithesis_instrumentation__.Notify(497634)
			if log.V(2) {
				__antithesis_instrumentation__.Notify(497635)
				log.Warningf(ctx, "unable to record statement exec stats: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(497636)
			}
		} else {
			__antithesis_instrumentation__.Notify(497637)
		}
		__antithesis_instrumentation__.Notify(497631)
		if collectTxnExecStats || func() bool {
			__antithesis_instrumentation__.Notify(497638)
			return ih.implicitTxn == true
		}() == true {
			__antithesis_instrumentation__.Notify(497639)
			txnStats.Accumulate(queryLevelStats)
		} else {
			__antithesis_instrumentation__.Notify(497640)
		}
	}
	__antithesis_instrumentation__.Notify(497612)

	var bundle diagnosticsBundle
	if ih.collectBundle {
		__antithesis_instrumentation__.Notify(497641)
		ie := p.extendedEvalCtx.ExecCfg.InternalExecutorFactory(
			p.EvalContext().Context,
			p.SessionData(),
		)
		phaseTimes := statsCollector.PhaseTimes()
		if ih.stmtDiagnosticsRecorder.IsExecLatencyConditionMet(
			ih.diagRequestID, ih.diagRequest, phaseTimes.GetServiceLatencyNoOverhead(),
		) {
			__antithesis_instrumentation__.Notify(497642)
			placeholders := p.extendedEvalCtx.Placeholders
			ob := ih.emitExplainAnalyzePlanToOutputBuilder(
				explain.Flags{Verbose: true, ShowTypes: true},
				phaseTimes,
				&queryLevelStats,
			)
			bundle = buildStatementBundle(
				ih.origCtx, cfg.DB, ie.(*InternalExecutor), &p.curPlan, ob.BuildString(), trace, placeholders,
			)
			bundle.insert(ctx, ih.fingerprint, ast, cfg.StmtDiagnosticsRecorder, ih.diagRequestID)
			ih.stmtDiagnosticsRecorder.RemoveOngoing(ih.diagRequestID, ih.diagRequest)
			telemetry.Inc(sqltelemetry.StatementDiagnosticsCollectedCounter)
		} else {
			__antithesis_instrumentation__.Notify(497643)
		}
	} else {
		__antithesis_instrumentation__.Notify(497644)
	}
	__antithesis_instrumentation__.Notify(497613)

	if retErr != nil {
		__antithesis_instrumentation__.Notify(497645)
		return retErr
	} else {
		__antithesis_instrumentation__.Notify(497646)
	}
	__antithesis_instrumentation__.Notify(497614)

	switch ih.outputMode {
	case explainAnalyzeDebugOutput:
		__antithesis_instrumentation__.Notify(497647)
		return setExplainBundleResult(ctx, res, bundle, cfg)

	case explainAnalyzePlanOutput, explainAnalyzeDistSQLOutput:
		__antithesis_instrumentation__.Notify(497648)
		var flows []flowInfo
		if ih.outputMode == explainAnalyzeDistSQLOutput {
			__antithesis_instrumentation__.Notify(497651)
			flows = p.curPlan.distSQLFlowInfos
		} else {
			__antithesis_instrumentation__.Notify(497652)
		}
		__antithesis_instrumentation__.Notify(497649)
		return ih.setExplainAnalyzeResult(ctx, res, statsCollector.PhaseTimes(), &queryLevelStats, flows, trace)

	default:
		__antithesis_instrumentation__.Notify(497650)
		return nil
	}
}

func (ih *instrumentationHelper) SetDiscardRows() {
	__antithesis_instrumentation__.Notify(497653)
	ih.discardRows = true
}

func (ih *instrumentationHelper) ShouldDiscardRows() bool {
	__antithesis_instrumentation__.Notify(497654)
	return ih.discardRows
}

func (ih *instrumentationHelper) ShouldSaveFlows() bool {
	__antithesis_instrumentation__.Notify(497655)
	return ih.collectBundle || func() bool {
		__antithesis_instrumentation__.Notify(497656)
		return ih.outputMode == explainAnalyzeDistSQLOutput == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(497657)
		return ih.collectExecStats == true
	}() == true
}

func (ih *instrumentationHelper) shouldSaveDiagrams() bool {
	__antithesis_instrumentation__.Notify(497658)
	return ih.collectBundle || func() bool {
		__antithesis_instrumentation__.Notify(497659)
		return ih.outputMode != unmodifiedOutput == true
	}() == true
}

func (ih *instrumentationHelper) ShouldUseJobForCreateStats() bool {
	__antithesis_instrumentation__.Notify(497660)
	return ih.outputMode == unmodifiedOutput
}

func (ih *instrumentationHelper) ShouldBuildExplainPlan() bool {
	__antithesis_instrumentation__.Notify(497661)
	return ih.collectBundle || func() bool {
		__antithesis_instrumentation__.Notify(497662)
		return ih.savePlanForStats == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(497663)
		return ih.outputMode == explainAnalyzePlanOutput == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(497664)
		return ih.outputMode == explainAnalyzeDistSQLOutput == true
	}() == true
}

func (ih *instrumentationHelper) ShouldCollectExecStats() bool {
	__antithesis_instrumentation__.Notify(497665)
	return ih.collectExecStats
}

func (ih *instrumentationHelper) ShouldSaveMemo() bool {
	__antithesis_instrumentation__.Notify(497666)
	return ih.ShouldBuildExplainPlan()
}

func (ih *instrumentationHelper) RecordExplainPlan(explainPlan *explain.Plan) {
	__antithesis_instrumentation__.Notify(497667)
	ih.explainPlan = explainPlan
}

func (ih *instrumentationHelper) RecordPlanInfo(
	distribution physicalplan.PlanDistribution, vectorized bool,
) {
	__antithesis_instrumentation__.Notify(497668)
	ih.distribution = distribution
	ih.vectorized = vectorized
}

func (ih *instrumentationHelper) PlanForStats(ctx context.Context) *roachpb.ExplainTreePlanNode {
	__antithesis_instrumentation__.Notify(497669)
	if ih.explainPlan == nil {
		__antithesis_instrumentation__.Notify(497672)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(497673)
	}
	__antithesis_instrumentation__.Notify(497670)

	ob := explain.NewOutputBuilder(explain.Flags{
		HideValues: true,
	})
	ob.AddDistribution(ih.distribution.String())
	ob.AddVectorized(ih.vectorized)
	if err := emitExplain(ob, ih.evalCtx, ih.codec, ih.explainPlan); err != nil {
		__antithesis_instrumentation__.Notify(497674)
		log.Warningf(ctx, "unable to emit explain plan tree: %v", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(497675)
	}
	__antithesis_instrumentation__.Notify(497671)
	return ob.BuildProtoTree()
}

func (ih *instrumentationHelper) emitExplainAnalyzePlanToOutputBuilder(
	flags explain.Flags, phaseTimes *sessionphase.Times, queryStats *execstats.QueryLevelStats,
) *explain.OutputBuilder {
	__antithesis_instrumentation__.Notify(497676)
	ob := explain.NewOutputBuilder(flags)
	if ih.explainPlan == nil {
		__antithesis_instrumentation__.Notify(497683)

		return ob
	} else {
		__antithesis_instrumentation__.Notify(497684)
	}
	__antithesis_instrumentation__.Notify(497677)
	ob.AddPlanningTime(phaseTimes.GetPlanningLatency())
	ob.AddExecutionTime(phaseTimes.GetRunLatency())
	ob.AddDistribution(ih.distribution.String())
	ob.AddVectorized(ih.vectorized)

	if queryStats.KVRowsRead != 0 {
		__antithesis_instrumentation__.Notify(497685)
		ob.AddKVReadStats(queryStats.KVRowsRead, queryStats.KVBytesRead)
	} else {
		__antithesis_instrumentation__.Notify(497686)
	}
	__antithesis_instrumentation__.Notify(497678)
	if queryStats.KVTime != 0 {
		__antithesis_instrumentation__.Notify(497687)
		ob.AddKVTime(queryStats.KVTime)
	} else {
		__antithesis_instrumentation__.Notify(497688)
	}
	__antithesis_instrumentation__.Notify(497679)
	if queryStats.ContentionTime != 0 {
		__antithesis_instrumentation__.Notify(497689)
		ob.AddContentionTime(queryStats.ContentionTime)
	} else {
		__antithesis_instrumentation__.Notify(497690)
	}
	__antithesis_instrumentation__.Notify(497680)

	ob.AddMaxMemUsage(queryStats.MaxMemUsage)
	ob.AddNetworkStats(queryStats.NetworkMessages, queryStats.NetworkBytesSent)
	ob.AddMaxDiskUsage(queryStats.MaxDiskUsage)

	if len(ih.regions) > 0 {
		__antithesis_instrumentation__.Notify(497691)
		ob.AddRegionsStats(ih.regions)
	} else {
		__antithesis_instrumentation__.Notify(497692)
	}
	__antithesis_instrumentation__.Notify(497681)

	if err := emitExplain(ob, ih.evalCtx, ih.codec, ih.explainPlan); err != nil {
		__antithesis_instrumentation__.Notify(497693)
		ob.AddTopLevelField("error emitting plan", fmt.Sprint(err))
	} else {
		__antithesis_instrumentation__.Notify(497694)
	}
	__antithesis_instrumentation__.Notify(497682)
	return ob
}

func (ih *instrumentationHelper) setExplainAnalyzeResult(
	ctx context.Context,
	res RestrictedCommandResult,
	phaseTimes *sessionphase.Times,
	queryLevelStats *execstats.QueryLevelStats,
	distSQLFlowInfos []flowInfo,
	trace tracing.Recording,
) (commErr error) {
	__antithesis_instrumentation__.Notify(497695)
	res.ResetStmtType(&tree.ExplainAnalyze{})
	res.SetColumns(ctx, colinfo.ExplainPlanColumns)

	if res.Err() != nil {
		__antithesis_instrumentation__.Notify(497699)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(497700)
	}
	__antithesis_instrumentation__.Notify(497696)

	ob := ih.emitExplainAnalyzePlanToOutputBuilder(ih.explainFlags, phaseTimes, queryLevelStats)
	rows := ob.BuildStringRows()
	if distSQLFlowInfos != nil {
		__antithesis_instrumentation__.Notify(497701)
		rows = append(rows, "")
		for i, d := range distSQLFlowInfos {
			__antithesis_instrumentation__.Notify(497702)
			var buf bytes.Buffer
			if len(distSQLFlowInfos) > 1 {
				__antithesis_instrumentation__.Notify(497705)
				fmt.Fprintf(&buf, "Diagram %d (%s): ", i+1, d.typ)
			} else {
				__antithesis_instrumentation__.Notify(497706)
				buf.WriteString("Diagram: ")
			}
			__antithesis_instrumentation__.Notify(497703)
			d.diagram.AddSpans(trace)
			_, url, err := d.diagram.ToURL()
			if err != nil {
				__antithesis_instrumentation__.Notify(497707)
				buf.WriteString(err.Error())
			} else {
				__antithesis_instrumentation__.Notify(497708)
				buf.WriteString(url.String())
			}
			__antithesis_instrumentation__.Notify(497704)
			rows = append(rows, buf.String())
		}
	} else {
		__antithesis_instrumentation__.Notify(497709)
	}
	__antithesis_instrumentation__.Notify(497697)
	for _, row := range rows {
		__antithesis_instrumentation__.Notify(497710)
		if err := res.AddRow(ctx, tree.Datums{tree.NewDString(row)}); err != nil {
			__antithesis_instrumentation__.Notify(497711)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497712)
		}
	}
	__antithesis_instrumentation__.Notify(497698)
	return nil
}

type execNodeTraceMetadata map[exec.Node]execComponents

type execComponents []execinfrapb.ComponentID

func (m execNodeTraceMetadata) associateNodeWithComponents(
	node exec.Node, components execComponents,
) {
	__antithesis_instrumentation__.Notify(497713)
	m[node] = components
}

func (m execNodeTraceMetadata) annotateExplain(
	plan *explain.Plan, spans []tracingpb.RecordedSpan, makeDeterministic bool, p *planner,
) []string {
	__antithesis_instrumentation__.Notify(497714)
	statsMap := execinfrapb.ExtractStatsFromSpans(spans, makeDeterministic)
	var allRegions []string

	regionsInfo := make(map[int64]string)
	descriptors, _ := getAllNodeDescriptors(p)
	for _, descriptor := range descriptors {
		__antithesis_instrumentation__.Notify(497719)
		for _, tier := range descriptor.Locality.Tiers {
			__antithesis_instrumentation__.Notify(497720)
			if tier.Key == "region" {
				__antithesis_instrumentation__.Notify(497721)
				regionsInfo[int64(descriptor.NodeID)] = tier.Value
			} else {
				__antithesis_instrumentation__.Notify(497722)
			}
		}
	}
	__antithesis_instrumentation__.Notify(497715)

	var walk func(n *explain.Node)
	walk = func(n *explain.Node) {
		__antithesis_instrumentation__.Notify(497723)
		wrapped := n.WrappedNode()
		if components, ok := m[wrapped]; ok {
			__antithesis_instrumentation__.Notify(497725)
			var nodeStats exec.ExecutionStats

			incomplete := false
			var nodes util.FastIntSet
			regionsMap := make(map[string]struct{})
			for _, c := range components {
				__antithesis_instrumentation__.Notify(497727)
				if c.Type == execinfrapb.ComponentID_PROCESSOR {
					__antithesis_instrumentation__.Notify(497730)
					nodes.Add(int(c.SQLInstanceID))
					regionsMap[regionsInfo[int64(c.SQLInstanceID)]] = struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(497731)
				}
				__antithesis_instrumentation__.Notify(497728)
				stats := statsMap[c]
				if stats == nil {
					__antithesis_instrumentation__.Notify(497732)
					incomplete = true
					break
				} else {
					__antithesis_instrumentation__.Notify(497733)
				}
				__antithesis_instrumentation__.Notify(497729)
				nodeStats.RowCount.MaybeAdd(stats.Output.NumTuples)
				nodeStats.KVTime.MaybeAdd(stats.KV.KVTime)
				nodeStats.KVContentionTime.MaybeAdd(stats.KV.ContentionTime)
				nodeStats.KVBytesRead.MaybeAdd(stats.KV.BytesRead)
				nodeStats.KVRowsRead.MaybeAdd(stats.KV.TuplesRead)
				nodeStats.StepCount.MaybeAdd(stats.KV.NumInterfaceSteps)
				nodeStats.InternalStepCount.MaybeAdd(stats.KV.NumInternalSteps)
				nodeStats.SeekCount.MaybeAdd(stats.KV.NumInterfaceSeeks)
				nodeStats.InternalSeekCount.MaybeAdd(stats.KV.NumInternalSeeks)
				nodeStats.VectorizedBatchCount.MaybeAdd(stats.Output.NumBatches)
				nodeStats.MaxAllocatedMem.MaybeAdd(stats.Exec.MaxAllocatedMem)
				nodeStats.MaxAllocatedDisk.MaybeAdd(stats.Exec.MaxAllocatedDisk)
			}
			__antithesis_instrumentation__.Notify(497726)

			if !incomplete {
				__antithesis_instrumentation__.Notify(497734)
				for i, ok := nodes.Next(0); ok; i, ok = nodes.Next(i + 1) {
					__antithesis_instrumentation__.Notify(497737)
					nodeStats.Nodes = append(nodeStats.Nodes, fmt.Sprintf("n%d", i))
				}
				__antithesis_instrumentation__.Notify(497735)
				regions := make([]string, 0, len(regionsMap))
				for r := range regionsMap {
					__antithesis_instrumentation__.Notify(497738)

					if r != "" {
						__antithesis_instrumentation__.Notify(497739)
						regions = append(regions, r)
					} else {
						__antithesis_instrumentation__.Notify(497740)
					}
				}
				__antithesis_instrumentation__.Notify(497736)
				sort.Strings(regions)
				nodeStats.Regions = regions
				allRegions = util.CombineUniqueString(allRegions, regions)
				n.Annotate(exec.ExecutionStatsID, &nodeStats)
			} else {
				__antithesis_instrumentation__.Notify(497741)
			}
		} else {
			__antithesis_instrumentation__.Notify(497742)
		}
		__antithesis_instrumentation__.Notify(497724)

		for i := 0; i < n.ChildCount(); i++ {
			__antithesis_instrumentation__.Notify(497743)
			walk(n.Child(i))
		}
	}
	__antithesis_instrumentation__.Notify(497716)

	walk(plan.Root)
	for i := range plan.Subqueries {
		__antithesis_instrumentation__.Notify(497744)
		walk(plan.Subqueries[i].Root.(*explain.Node))
	}
	__antithesis_instrumentation__.Notify(497717)
	for i := range plan.Checks {
		__antithesis_instrumentation__.Notify(497745)
		walk(plan.Checks[i])
	}
	__antithesis_instrumentation__.Notify(497718)

	return allRegions
}
