package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

var CteUseCounter = telemetry.GetCounterOnce("sql.plan.cte")

var RecursiveCteUseCounter = telemetry.GetCounterOnce("sql.plan.cte.recursive")

var SubqueryUseCounter = telemetry.GetCounterOnce("sql.plan.subquery")

var CorrelatedSubqueryUseCounter = telemetry.GetCounterOnce("sql.plan.subquery.correlated")

var UniqueChecksUseCounter = telemetry.GetCounterOnce("sql.plan.unique.checks")

var ForeignKeyChecksUseCounter = telemetry.GetCounterOnce("sql.plan.fk.checks")

var ForeignKeyCascadesUseCounter = telemetry.GetCounterOnce("sql.plan.fk.cascades")

var LateralJoinUseCounter = telemetry.GetCounterOnce("sql.plan.lateral-join")

var HashJoinHintUseCounter = telemetry.GetCounterOnce("sql.plan.hints.hash-join")

var MergeJoinHintUseCounter = telemetry.GetCounterOnce("sql.plan.hints.merge-join")

var LookupJoinHintUseCounter = telemetry.GetCounterOnce("sql.plan.hints.lookup-join")

var InvertedJoinHintUseCounter = telemetry.GetCounterOnce("sql.plan.hints.inverted-join")

var IndexHintUseCounter = telemetry.GetCounterOnce("sql.plan.hints.index")

var IndexHintSelectUseCounter = telemetry.GetCounterOnce("sql.plan.hints.index.select")

var IndexHintUpdateUseCounter = telemetry.GetCounterOnce("sql.plan.hints.index.update")

var IndexHintDeleteUseCounter = telemetry.GetCounterOnce("sql.plan.hints.index.delete")

var ExplainPlanUseCounter = telemetry.GetCounterOnce("sql.plan.explain")

var ExplainDistSQLUseCounter = telemetry.GetCounterOnce("sql.plan.explain-distsql")

var ExplainAnalyzeUseCounter = telemetry.GetCounterOnce("sql.plan.explain-analyze")

var ExplainAnalyzeDistSQLUseCounter = telemetry.GetCounterOnce("sql.plan.explain-analyze-distsql")

var ExplainAnalyzeDebugUseCounter = telemetry.GetCounterOnce("sql.plan.explain-analyze-debug")

var ExplainOptUseCounter = telemetry.GetCounterOnce("sql.plan.explain-opt")

var ExplainVecUseCounter = telemetry.GetCounterOnce("sql.plan.explain-vec")

var ExplainDDL = telemetry.GetCounterOnce("sql.plan.explain-ddl")

var ExplainDDLVerbose = telemetry.GetCounterOnce("sql.plan.explain-ddl-verbose")

var ExplainDDLViz = telemetry.GetCounterOnce("sql.plan.explain-ddl-viz")

var ExplainOptVerboseUseCounter = telemetry.GetCounterOnce("sql.plan.explain-opt-verbose")

var ExplainGist = telemetry.GetCounterOnce("sql.plan.explain-gist")

var CreateStatisticsUseCounter = telemetry.GetCounterOnce("sql.plan.stats.created")

var OrderByNullsNonStandardCounter = telemetry.GetCounterOnce("sql.plan.opt.order-by-nulls-non-standard")

var TurnAutoStatsOnUseCounter = telemetry.GetCounterOnce("sql.plan.automatic-stats.enabled")

var TurnAutoStatsOffUseCounter = telemetry.GetCounterOnce("sql.plan.automatic-stats.disabled")

var StatsHistogramOOMCounter = telemetry.GetCounterOnce("sql.plan.stats.histogram-oom")

var JoinAlgoHashUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.algo.hash")

var JoinAlgoMergeUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.algo.merge")

var JoinAlgoLookupUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.algo.lookup")

var JoinAlgoCrossUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.algo.cross")

var JoinTypeInnerUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.type.inner")

var JoinTypeLeftUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.type.left-outer")

var JoinTypeFullUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.type.full-outer")

var JoinTypeSemiUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.type.semi")

var JoinTypeAntiUseCounter = telemetry.GetCounterOnce("sql.plan.opt.node.join.type.anti")

var PartialIndexScanUseCounter = telemetry.GetCounterOnce("sql.plan.opt.partial-index.scan")

var PartialIndexLookupJoinUseCounter = telemetry.GetCounterOnce("sql.plan.opt.partial-index.lookup-join")

var LocalityOptimizedSearchUseCounter = telemetry.GetCounterOnce("sql.plan.opt.locality-optimized-search")

var CancelQueriesUseCounter = telemetry.GetCounterOnce("sql.session.cancel-queries")

var CancelSessionsUseCounter = telemetry.GetCounterOnce("sql.session.cancel-sessions")

var reorderJoinLimitUseCounters []telemetry.Counter

const reorderJoinsCounters = 12

func init() {
	reorderJoinLimitUseCounters = make([]telemetry.Counter, reorderJoinsCounters)

	for i := 0; i < reorderJoinsCounters; i++ {
		reorderJoinLimitUseCounters[i] = telemetry.GetCounterOnce(
			fmt.Sprintf("sql.plan.reorder-joins.set-limit-%d", i),
		)
	}
}

var reorderJoinLimitMoreCounter = telemetry.GetCounterOnce("sql.plan.reorder-joins.set-limit-more")

func ReportJoinReorderLimit(value int) {
	__antithesis_instrumentation__.Notify(625807)
	if value < reorderJoinsCounters {
		__antithesis_instrumentation__.Notify(625808)
		telemetry.Inc(reorderJoinLimitUseCounters[value])
	} else {
		__antithesis_instrumentation__.Notify(625809)
		telemetry.Inc(reorderJoinLimitMoreCounter)
	}
}

func WindowFunctionCounter(wf string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625810)
	return telemetry.GetCounter("sql.plan.window_function." + wf)
}

func OptNodeCounter(nodeType string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625811)
	return telemetry.GetCounterOnce(fmt.Sprintf("sql.plan.opt.node.%s", nodeType))
}
