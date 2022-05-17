package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

type joinPlanningInfo struct {
	leftPlan, rightPlan *PhysicalPlan
	joinType            descpb.JoinType
	joinResultTypes     []*types.T
	onExpr              execinfrapb.Expression
	post                execinfrapb.PostProcessSpec
	joinToStreamColMap  []int

	leftEqCols, rightEqCols             []uint32
	leftEqColsAreKey, rightEqColsAreKey bool

	leftMergeOrd, rightMergeOrd                 execinfrapb.Ordering
	leftPlanDistribution, rightPlanDistribution physicalplan.PlanDistribution
	allowPartialDistribution                    bool
}

func (info *joinPlanningInfo) makeCoreSpec() execinfrapb.ProcessorCoreUnion {
	__antithesis_instrumentation__.Notify(467671)
	var core execinfrapb.ProcessorCoreUnion
	if len(info.leftMergeOrd.Columns) != len(info.rightMergeOrd.Columns) {
		__antithesis_instrumentation__.Notify(467674)
		panic(errors.AssertionFailedf(
			"unexpectedly different merge join ordering lengths: left %d, right %d",
			len(info.leftMergeOrd.Columns), len(info.rightMergeOrd.Columns),
		))
	} else {
		__antithesis_instrumentation__.Notify(467675)
	}
	__antithesis_instrumentation__.Notify(467672)
	if len(info.leftMergeOrd.Columns) == 0 {
		__antithesis_instrumentation__.Notify(467676)

		core.HashJoiner = &execinfrapb.HashJoinerSpec{
			LeftEqColumns:        info.leftEqCols,
			RightEqColumns:       info.rightEqCols,
			OnExpr:               info.onExpr,
			Type:                 info.joinType,
			LeftEqColumnsAreKey:  info.leftEqColsAreKey,
			RightEqColumnsAreKey: info.rightEqColsAreKey,
		}
	} else {
		__antithesis_instrumentation__.Notify(467677)
		core.MergeJoiner = &execinfrapb.MergeJoinerSpec{
			LeftOrdering:         info.leftMergeOrd,
			RightOrdering:        info.rightMergeOrd,
			OnExpr:               info.onExpr,
			Type:                 info.joinType,
			LeftEqColumnsAreKey:  info.leftEqColsAreKey,
			RightEqColumnsAreKey: info.rightEqColsAreKey,
		}
	}
	__antithesis_instrumentation__.Notify(467673)
	return core
}

type joinPlanningHelper struct {
	numLeftOutCols, numRightOutCols int

	numAllLeftCols                                  int
	leftPlanToStreamColMap, rightPlanToStreamColMap []int
}

func (h *joinPlanningHelper) joinOutColumns(
	joinType descpb.JoinType, columns colinfo.ResultColumns,
) (post execinfrapb.PostProcessSpec, joinToStreamColMap []int) {
	__antithesis_instrumentation__.Notify(467678)
	joinToStreamColMap = makePlanToStreamColMap(len(columns))
	post.Projection = true

	addOutCol := func(col uint32) int {
		__antithesis_instrumentation__.Notify(467682)
		idx := len(post.OutputColumns)
		post.OutputColumns = append(post.OutputColumns, col)
		return idx
	}
	__antithesis_instrumentation__.Notify(467679)

	var numLeftOutCols int
	var numAllLeftCols int
	if joinType.ShouldIncludeLeftColsInOutput() {
		__antithesis_instrumentation__.Notify(467683)
		numLeftOutCols = h.numLeftOutCols
		numAllLeftCols = h.numAllLeftCols
		for i := 0; i < h.numLeftOutCols; i++ {
			__antithesis_instrumentation__.Notify(467684)
			joinToStreamColMap[i] = addOutCol(uint32(h.leftPlanToStreamColMap[i]))
		}
	} else {
		__antithesis_instrumentation__.Notify(467685)
	}
	__antithesis_instrumentation__.Notify(467680)

	if joinType.ShouldIncludeRightColsInOutput() {
		__antithesis_instrumentation__.Notify(467686)
		for i := 0; i < h.numRightOutCols; i++ {
			__antithesis_instrumentation__.Notify(467687)
			joinToStreamColMap[numLeftOutCols+i] = addOutCol(
				uint32(numAllLeftCols + h.rightPlanToStreamColMap[i]),
			)
		}
	} else {
		__antithesis_instrumentation__.Notify(467688)
	}
	__antithesis_instrumentation__.Notify(467681)

	return post, joinToStreamColMap
}

func (h *joinPlanningHelper) remapOnExpr(
	planCtx *PlanningCtx, onCond tree.TypedExpr,
) (execinfrapb.Expression, error) {
	__antithesis_instrumentation__.Notify(467689)
	if onCond == nil {
		__antithesis_instrumentation__.Notify(467693)
		return execinfrapb.Expression{}, nil
	} else {
		__antithesis_instrumentation__.Notify(467694)
	}
	__antithesis_instrumentation__.Notify(467690)

	joinColMap := make([]int, h.numLeftOutCols+h.numRightOutCols)
	idx := 0
	leftCols := 0
	for i := 0; i < h.numLeftOutCols; i++ {
		__antithesis_instrumentation__.Notify(467695)
		joinColMap[idx] = h.leftPlanToStreamColMap[i]
		if h.leftPlanToStreamColMap[i] != -1 {
			__antithesis_instrumentation__.Notify(467697)
			leftCols++
		} else {
			__antithesis_instrumentation__.Notify(467698)
		}
		__antithesis_instrumentation__.Notify(467696)
		idx++
	}
	__antithesis_instrumentation__.Notify(467691)
	for i := 0; i < h.numRightOutCols; i++ {
		__antithesis_instrumentation__.Notify(467699)
		joinColMap[idx] = leftCols + h.rightPlanToStreamColMap[i]
		idx++
	}
	__antithesis_instrumentation__.Notify(467692)

	return physicalplan.MakeExpression(onCond, planCtx, joinColMap)
}

func eqCols(eqIndices []exec.NodeColumnOrdinal, planToColMap []int) []uint32 {
	__antithesis_instrumentation__.Notify(467700)
	eqCols := make([]uint32, len(eqIndices))
	for i, planCol := range eqIndices {
		__antithesis_instrumentation__.Notify(467702)
		eqCols[i] = uint32(planToColMap[planCol])
	}
	__antithesis_instrumentation__.Notify(467701)

	return eqCols
}

func distsqlOrdering(
	mergeJoinOrdering colinfo.ColumnOrdering, eqCols []uint32,
) execinfrapb.Ordering {
	__antithesis_instrumentation__.Notify(467703)
	var ord execinfrapb.Ordering
	ord.Columns = make([]execinfrapb.Ordering_Column, len(mergeJoinOrdering))
	for i, c := range mergeJoinOrdering {
		__antithesis_instrumentation__.Notify(467705)
		ord.Columns[i].ColIdx = eqCols[c.ColIdx]
		dir := execinfrapb.Ordering_Column_ASC
		if c.Direction == encoding.Descending {
			__antithesis_instrumentation__.Notify(467707)
			dir = execinfrapb.Ordering_Column_DESC
		} else {
			__antithesis_instrumentation__.Notify(467708)
		}
		__antithesis_instrumentation__.Notify(467706)
		ord.Columns[i].Direction = dir
	}
	__antithesis_instrumentation__.Notify(467704)

	return ord
}

func distsqlSetOpJoinType(setOpType tree.UnionType) descpb.JoinType {
	__antithesis_instrumentation__.Notify(467709)
	switch setOpType {
	case tree.ExceptOp:
		__antithesis_instrumentation__.Notify(467710)
		return descpb.ExceptAllJoin
	case tree.IntersectOp:
		__antithesis_instrumentation__.Notify(467711)
		return descpb.IntersectAllJoin
	default:
		__antithesis_instrumentation__.Notify(467712)
		panic(errors.AssertionFailedf("set op type %v unsupported by joins", setOpType))
	}
}

func getSQLInstanceIDsOfRouters(
	routers []physicalplan.ProcessorIdx, processors []physicalplan.Processor,
) (sqlInstanceIDs []base.SQLInstanceID) {
	__antithesis_instrumentation__.Notify(467713)
	seen := make(map[base.SQLInstanceID]struct{})
	for _, pIdx := range routers {
		__antithesis_instrumentation__.Notify(467715)
		n := processors[pIdx].SQLInstanceID
		if _, ok := seen[n]; !ok {
			__antithesis_instrumentation__.Notify(467716)
			seen[n] = struct{}{}
			sqlInstanceIDs = append(sqlInstanceIDs, n)
		} else {
			__antithesis_instrumentation__.Notify(467717)
		}
	}
	__antithesis_instrumentation__.Notify(467714)
	return sqlInstanceIDs
}

func findJoinProcessorNodes(
	leftRouters, rightRouters []physicalplan.ProcessorIdx, processors []physicalplan.Processor,
) (instances []base.SQLInstanceID) {
	__antithesis_instrumentation__.Notify(467718)

	return getSQLInstanceIDsOfRouters(append(leftRouters, rightRouters...), processors)
}
