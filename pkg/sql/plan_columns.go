package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

var noColumns = make(colinfo.ResultColumns, 0)

func planColumns(plan planNode) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(562656)
	return getPlanColumns(plan, false)
}

func planMutableColumns(plan planNode) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(562657)
	return getPlanColumns(plan, true)
}

func getPlanColumns(plan planNode, mut bool) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(562658)
	switch n := plan.(type) {

	case *delayedNode:
		__antithesis_instrumentation__.Notify(562660)
		return n.columns
	case *groupNode:
		__antithesis_instrumentation__.Notify(562661)
		return n.columns
	case *joinNode:
		__antithesis_instrumentation__.Notify(562662)
		return n.columns
	case *ordinalityNode:
		__antithesis_instrumentation__.Notify(562663)
		return n.columns
	case *renderNode:
		__antithesis_instrumentation__.Notify(562664)
		return n.columns
	case *scanNode:
		__antithesis_instrumentation__.Notify(562665)
		return n.resultColumns
	case *unionNode:
		__antithesis_instrumentation__.Notify(562666)
		return n.columns
	case *valuesNode:
		__antithesis_instrumentation__.Notify(562667)
		return n.columns
	case *virtualTableNode:
		__antithesis_instrumentation__.Notify(562668)
		return n.columns
	case *windowNode:
		__antithesis_instrumentation__.Notify(562669)
		return n.columns
	case *showTraceNode:
		__antithesis_instrumentation__.Notify(562670)
		return n.columns
	case *zeroNode:
		__antithesis_instrumentation__.Notify(562671)
		return n.columns
	case *deleteNode:
		__antithesis_instrumentation__.Notify(562672)
		return n.columns
	case *updateNode:
		__antithesis_instrumentation__.Notify(562673)
		return n.columns
	case *insertNode:
		__antithesis_instrumentation__.Notify(562674)
		return n.columns
	case *insertFastPathNode:
		__antithesis_instrumentation__.Notify(562675)
		return n.columns
	case *upsertNode:
		__antithesis_instrumentation__.Notify(562676)
		return n.columns
	case *indexJoinNode:
		__antithesis_instrumentation__.Notify(562677)
		return n.resultColumns
	case *projectSetNode:
		__antithesis_instrumentation__.Notify(562678)
		return n.columns
	case *applyJoinNode:
		__antithesis_instrumentation__.Notify(562679)
		return n.columns
	case *lookupJoinNode:
		__antithesis_instrumentation__.Notify(562680)
		return n.columns
	case *zigzagJoinNode:
		__antithesis_instrumentation__.Notify(562681)
		return n.columns
	case *vTableLookupJoinNode:
		__antithesis_instrumentation__.Notify(562682)
		return n.columns
	case *invertedFilterNode:
		__antithesis_instrumentation__.Notify(562683)
		return n.resultColumns
	case *invertedJoinNode:
		__antithesis_instrumentation__.Notify(562684)
		return n.columns

	case *scrubNode:
		__antithesis_instrumentation__.Notify(562685)
		return n.getColumns(mut, colinfo.ScrubColumns)
	case *explainDDLNode:
		__antithesis_instrumentation__.Notify(562686)
		return n.getColumns(mut, colinfo.ExplainPlanColumns)
	case *explainPlanNode:
		__antithesis_instrumentation__.Notify(562687)
		return n.getColumns(mut, colinfo.ExplainPlanColumns)
	case *explainVecNode:
		__antithesis_instrumentation__.Notify(562688)
		return n.getColumns(mut, colinfo.ExplainPlanColumns)
	case *relocateNode:
		__antithesis_instrumentation__.Notify(562689)
		return n.getColumns(mut, colinfo.AlterTableRelocateColumns)
	case *relocateRange:
		__antithesis_instrumentation__.Notify(562690)
		return n.getColumns(mut, colinfo.AlterRangeRelocateColumns)
	case *scatterNode:
		__antithesis_instrumentation__.Notify(562691)
		return n.getColumns(mut, colinfo.AlterTableScatterColumns)
	case *showFingerprintsNode:
		__antithesis_instrumentation__.Notify(562692)
		return n.getColumns(mut, colinfo.ShowFingerprintsColumns)
	case *splitNode:
		__antithesis_instrumentation__.Notify(562693)
		return n.getColumns(mut, colinfo.AlterTableSplitColumns)
	case *unsplitNode:
		__antithesis_instrumentation__.Notify(562694)
		return n.getColumns(mut, colinfo.AlterTableUnsplitColumns)
	case *unsplitAllNode:
		__antithesis_instrumentation__.Notify(562695)
		return n.getColumns(mut, colinfo.AlterTableUnsplitColumns)
	case *showTraceReplicaNode:
		__antithesis_instrumentation__.Notify(562696)
		return n.getColumns(mut, colinfo.ShowReplicaTraceColumns)
	case *sequenceSelectNode:
		__antithesis_instrumentation__.Notify(562697)
		return n.getColumns(mut, colinfo.SequenceSelectColumns)
	case *exportNode:
		__antithesis_instrumentation__.Notify(562698)
		return n.getColumns(mut, colinfo.ExportColumns)

	case *hookFnNode:
		__antithesis_instrumentation__.Notify(562699)
		return n.getColumns(mut, n.header)

	case *bufferNode:
		__antithesis_instrumentation__.Notify(562700)
		return getPlanColumns(n.plan, mut)
	case *distinctNode:
		__antithesis_instrumentation__.Notify(562701)
		return getPlanColumns(n.plan, mut)
	case *fetchNode:
		__antithesis_instrumentation__.Notify(562702)
		return n.cursor.Types()
	case *filterNode:
		__antithesis_instrumentation__.Notify(562703)
		return getPlanColumns(n.source.plan, mut)
	case *max1RowNode:
		__antithesis_instrumentation__.Notify(562704)
		return getPlanColumns(n.plan, mut)
	case *limitNode:
		__antithesis_instrumentation__.Notify(562705)
		return getPlanColumns(n.plan, mut)
	case *spoolNode:
		__antithesis_instrumentation__.Notify(562706)
		return getPlanColumns(n.source, mut)
	case *serializeNode:
		__antithesis_instrumentation__.Notify(562707)
		return getPlanColumns(n.source, mut)
	case *saveTableNode:
		__antithesis_instrumentation__.Notify(562708)
		return getPlanColumns(n.source, mut)
	case *scanBufferNode:
		__antithesis_instrumentation__.Notify(562709)
		return getPlanColumns(n.buffer, mut)
	case *sortNode:
		__antithesis_instrumentation__.Notify(562710)
		return getPlanColumns(n.plan, mut)
	case *topKNode:
		__antithesis_instrumentation__.Notify(562711)
		return getPlanColumns(n.plan, mut)
	case *recursiveCTENode:
		__antithesis_instrumentation__.Notify(562712)
		return getPlanColumns(n.initial, mut)

	case *showVarNode:
		__antithesis_instrumentation__.Notify(562713)
		return colinfo.ResultColumns{
			{Name: n.name, Typ: types.String},
		}
	case *rowSourceToPlanNode:
		__antithesis_instrumentation__.Notify(562714)
		return n.planCols
	}
	__antithesis_instrumentation__.Notify(562659)

	return noColumns
}

func planTypes(plan planNode) []*types.T {
	__antithesis_instrumentation__.Notify(562715)
	columns := planColumns(plan)
	typs := make([]*types.T, len(columns))
	for i := range typs {
		__antithesis_instrumentation__.Notify(562717)
		typs[i] = columns[i].Typ
	}
	__antithesis_instrumentation__.Notify(562716)
	return typs
}

type optColumnsSlot struct {
	columns colinfo.ResultColumns
}

func (c *optColumnsSlot) getColumns(mut bool, cols colinfo.ResultColumns) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(562718)
	if c.columns != nil {
		__antithesis_instrumentation__.Notify(562721)
		return c.columns
	} else {
		__antithesis_instrumentation__.Notify(562722)
	}
	__antithesis_instrumentation__.Notify(562719)
	if !mut {
		__antithesis_instrumentation__.Notify(562723)
		return cols
	} else {
		__antithesis_instrumentation__.Notify(562724)
	}
	__antithesis_instrumentation__.Notify(562720)
	c.columns = make(colinfo.ResultColumns, len(cols))
	copy(c.columns, cols)
	return c.columns
}
