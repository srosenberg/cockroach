package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

func constructPlan(
	planner *planner,
	root exec.Node,
	subqueries []exec.Subquery,
	cascades []exec.Cascade,
	checks []exec.Node,
	rootRowCount int64,
) (exec.Plan, error) {
	__antithesis_instrumentation__.Notify(470121)
	res := &planComponents{}
	assignPlan := func(plan *planMaybePhysical, node exec.Node) {
		__antithesis_instrumentation__.Notify(470126)
		switch n := node.(type) {
		case planNode:
			__antithesis_instrumentation__.Notify(470127)
			plan.planNode = n
		case planMaybePhysical:
			__antithesis_instrumentation__.Notify(470128)
			*plan = n
		default:
			__antithesis_instrumentation__.Notify(470129)
			panic(errors.AssertionFailedf("unexpected node type %T", node))
		}
	}
	__antithesis_instrumentation__.Notify(470122)
	assignPlan(&res.main, root)
	res.mainRowCount = rootRowCount
	if len(subqueries) > 0 {
		__antithesis_instrumentation__.Notify(470130)
		res.subqueryPlans = make([]subquery, len(subqueries))
		for i := range subqueries {
			__antithesis_instrumentation__.Notify(470131)
			in := &subqueries[i]
			out := &res.subqueryPlans[i]
			out.subquery = in.ExprNode
			switch in.Mode {
			case exec.SubqueryExists:
				__antithesis_instrumentation__.Notify(470133)
				out.execMode = rowexec.SubqueryExecModeExists
			case exec.SubqueryOneRow:
				__antithesis_instrumentation__.Notify(470134)
				out.execMode = rowexec.SubqueryExecModeOneRow
			case exec.SubqueryAnyRows:
				__antithesis_instrumentation__.Notify(470135)
				out.execMode = rowexec.SubqueryExecModeAllRowsNormalized
			case exec.SubqueryAllRows:
				__antithesis_instrumentation__.Notify(470136)
				out.execMode = rowexec.SubqueryExecModeAllRows
			default:
				__antithesis_instrumentation__.Notify(470137)
				return nil, errors.Errorf("invalid SubqueryMode %d", in.Mode)
			}
			__antithesis_instrumentation__.Notify(470132)
			out.expanded = true
			out.rowCount = in.RowCount
			assignPlan(&out.plan, in.Root)
		}
	} else {
		__antithesis_instrumentation__.Notify(470138)
	}
	__antithesis_instrumentation__.Notify(470123)
	if len(cascades) > 0 {
		__antithesis_instrumentation__.Notify(470139)
		res.cascades = make([]cascadeMetadata, len(cascades))
		for i := range cascades {
			__antithesis_instrumentation__.Notify(470140)
			res.cascades[i].Cascade = cascades[i]
		}
	} else {
		__antithesis_instrumentation__.Notify(470141)
	}
	__antithesis_instrumentation__.Notify(470124)
	if len(checks) > 0 {
		__antithesis_instrumentation__.Notify(470142)
		res.checkPlans = make([]checkPlan, len(checks))
		for i := range checks {
			__antithesis_instrumentation__.Notify(470143)
			assignPlan(&res.checkPlans[i].plan, checks[i])
		}
	} else {
		__antithesis_instrumentation__.Notify(470144)
	}
	__antithesis_instrumentation__.Notify(470125)

	return res, nil
}

func makeScanColumnsConfig(table cat.Table, cols exec.TableColumnOrdinalSet) scanColumnsConfig {
	__antithesis_instrumentation__.Notify(470145)
	colCfg := scanColumnsConfig{
		wantedColumns: make([]tree.ColumnID, 0, cols.Len()),
	}
	for ord, ok := cols.Next(0); ok; ord, ok = cols.Next(ord + 1) {
		__antithesis_instrumentation__.Notify(470147)
		col := table.Column(ord)
		if col.Kind() == cat.Inverted {
			__antithesis_instrumentation__.Notify(470149)
			colCfg.invertedColumnType = col.DatumType()
			colOrd := col.InvertedSourceColumnOrdinal()
			col = table.Column(colOrd)
			colCfg.invertedColumnID = tree.ColumnID(col.ColID())
		} else {
			__antithesis_instrumentation__.Notify(470150)
		}
		__antithesis_instrumentation__.Notify(470148)
		colCfg.wantedColumns = append(colCfg.wantedColumns, tree.ColumnID(col.ColID()))
	}
	__antithesis_instrumentation__.Notify(470146)
	return colCfg
}

func getResultColumnsForSimpleProject(
	cols []exec.NodeColumnOrdinal,
	colNames []string,
	resultTypes []*types.T,
	inputCols colinfo.ResultColumns,
) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(470151)
	resultCols := make(colinfo.ResultColumns, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(470153)
		if colNames == nil {
			__antithesis_instrumentation__.Notify(470154)
			resultCols[i] = inputCols[col]

			resultCols[i].Hidden = false
		} else {
			__antithesis_instrumentation__.Notify(470155)
			resultCols[i] = colinfo.ResultColumn{
				Name:           colNames[i],
				Typ:            resultTypes[i],
				TableID:        inputCols[col].TableID,
				PGAttributeNum: inputCols[col].PGAttributeNum,
			}
		}
	}
	__antithesis_instrumentation__.Notify(470152)
	return resultCols
}

func getEqualityIndicesAndMergeJoinOrdering(
	leftOrdering, rightOrdering colinfo.ColumnOrdering,
) (
	leftEqualityIndices, rightEqualityIndices []exec.NodeColumnOrdinal,
	mergeJoinOrdering colinfo.ColumnOrdering,
	err error,
) {
	__antithesis_instrumentation__.Notify(470156)
	n := len(leftOrdering)
	if n == 0 || func() bool {
		__antithesis_instrumentation__.Notify(470160)
		return len(rightOrdering) != n == true
	}() == true {
		__antithesis_instrumentation__.Notify(470161)
		return nil, nil, nil, errors.Errorf(
			"orderings from the left and right side must be the same non-zero length",
		)
	} else {
		__antithesis_instrumentation__.Notify(470162)
	}
	__antithesis_instrumentation__.Notify(470157)
	leftEqualityIndices = make([]exec.NodeColumnOrdinal, n)
	rightEqualityIndices = make([]exec.NodeColumnOrdinal, n)
	for i := 0; i < n; i++ {
		__antithesis_instrumentation__.Notify(470163)
		leftColIdx, rightColIdx := leftOrdering[i].ColIdx, rightOrdering[i].ColIdx
		leftEqualityIndices[i] = exec.NodeColumnOrdinal(leftColIdx)
		rightEqualityIndices[i] = exec.NodeColumnOrdinal(rightColIdx)
	}
	__antithesis_instrumentation__.Notify(470158)

	mergeJoinOrdering = make(colinfo.ColumnOrdering, n)
	for i := 0; i < n; i++ {
		__antithesis_instrumentation__.Notify(470164)

		mergeJoinOrdering[i].ColIdx = i
		mergeJoinOrdering[i].Direction = leftOrdering[i].Direction
	}
	__antithesis_instrumentation__.Notify(470159)
	return leftEqualityIndices, rightEqualityIndices, mergeJoinOrdering, nil
}

func getResultColumnsForGroupBy(
	inputCols colinfo.ResultColumns, groupCols []exec.NodeColumnOrdinal, aggregations []exec.AggInfo,
) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(470165)
	columns := make(colinfo.ResultColumns, 0, len(groupCols)+len(aggregations))
	for _, col := range groupCols {
		__antithesis_instrumentation__.Notify(470168)
		columns = append(columns, inputCols[col])
	}
	__antithesis_instrumentation__.Notify(470166)
	for _, agg := range aggregations {
		__antithesis_instrumentation__.Notify(470169)
		columns = append(columns, colinfo.ResultColumn{
			Name: agg.FuncName,
			Typ:  agg.ResultType,
		})
	}
	__antithesis_instrumentation__.Notify(470167)
	return columns
}

func convertNodeOrdinalsToInts(ordinals []exec.NodeColumnOrdinal) []int {
	__antithesis_instrumentation__.Notify(470170)
	ints := make([]int, len(ordinals))
	for i := range ordinals {
		__antithesis_instrumentation__.Notify(470172)
		ints[i] = int(ordinals[i])
	}
	__antithesis_instrumentation__.Notify(470171)
	return ints
}

func convertTableOrdinalsToInts(ordinals []exec.TableColumnOrdinal) []int {
	__antithesis_instrumentation__.Notify(470173)
	ints := make([]int, len(ordinals))
	for i := range ordinals {
		__antithesis_instrumentation__.Notify(470175)
		ints[i] = int(ordinals[i])
	}
	__antithesis_instrumentation__.Notify(470174)
	return ints
}

func constructVirtualScan(
	ef exec.Factory,
	p *planner,
	table cat.Table,
	index cat.Index,
	params exec.ScanParams,
	reqOrdering exec.OutputOrdering,

	delayedNodeCallback func(*delayedNode) (exec.Node, error),
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(470176)
	tn := &table.(*optVirtualTable).name
	virtual, err := p.getVirtualTabler().getVirtualTableEntry(tn)
	if err != nil {
		__antithesis_instrumentation__.Notify(470186)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(470187)
	}
	__antithesis_instrumentation__.Notify(470177)
	if !canQueryVirtualTable(p.EvalContext(), virtual) {
		__antithesis_instrumentation__.Notify(470188)
		return nil, newUnimplementedVirtualTableError(tn.Schema(), tn.Table())
	} else {
		__antithesis_instrumentation__.Notify(470189)
	}
	__antithesis_instrumentation__.Notify(470178)
	idx := index.(*optVirtualIndex).idx
	columns, constructor := virtual.getPlanInfo(
		table.(*optVirtualTable).desc,
		idx, params.IndexConstraint, p.execCfg.DistSQLPlanner.stopper)

	n, err := delayedNodeCallback(&delayedNode{
		name:            fmt.Sprintf("%s@%s", table.Name(), index.Name()),
		columns:         columns,
		indexConstraint: params.IndexConstraint,
		constructor: func(ctx context.Context, p *planner) (planNode, error) {
			__antithesis_instrumentation__.Notify(470190)
			return constructor(ctx, p, tn.Catalog())
		},
	})
	__antithesis_instrumentation__.Notify(470179)
	if err != nil {
		__antithesis_instrumentation__.Notify(470191)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(470192)
	}
	__antithesis_instrumentation__.Notify(470180)

	if params.NeededCols.Contains(0) {
		__antithesis_instrumentation__.Notify(470193)
		return nil, errors.Errorf("use of %s column not allowed.", table.Column(0).ColName())
	} else {
		__antithesis_instrumentation__.Notify(470194)
	}
	__antithesis_instrumentation__.Notify(470181)
	if params.Locking != nil {
		__antithesis_instrumentation__.Notify(470195)

		return nil, errors.AssertionFailedf("locking cannot be used with virtual table")
	} else {
		__antithesis_instrumentation__.Notify(470196)
	}
	__antithesis_instrumentation__.Notify(470182)
	if needed := params.NeededCols; needed.Len() != len(columns) {
		__antithesis_instrumentation__.Notify(470197)

		cols := make([]exec.NodeColumnOrdinal, 0, needed.Len())
		for ord, ok := needed.Next(0); ok; ord, ok = needed.Next(ord + 1) {
			__antithesis_instrumentation__.Notify(470199)
			cols = append(cols, exec.NodeColumnOrdinal(ord-1))
		}
		__antithesis_instrumentation__.Notify(470198)
		n, err = ef.ConstructSimpleProject(n, cols, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(470200)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(470201)
		}
	} else {
		__antithesis_instrumentation__.Notify(470202)
	}
	__antithesis_instrumentation__.Notify(470183)
	if params.HardLimit != 0 {
		__antithesis_instrumentation__.Notify(470203)
		n, err = ef.ConstructLimit(n, tree.NewDInt(tree.DInt(params.HardLimit)), nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(470204)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(470205)
		}
	} else {
		__antithesis_instrumentation__.Notify(470206)
	}
	__antithesis_instrumentation__.Notify(470184)

	if len(reqOrdering) != 0 {
		__antithesis_instrumentation__.Notify(470207)
		n, err = ef.ConstructSort(n, reqOrdering, 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(470208)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(470209)
		}
	} else {
		__antithesis_instrumentation__.Notify(470210)
	}
	__antithesis_instrumentation__.Notify(470185)
	return n, nil
}

func scanContainsSystemColumns(colCfg *scanColumnsConfig) bool {
	__antithesis_instrumentation__.Notify(470211)
	for _, id := range colCfg.wantedColumns {
		__antithesis_instrumentation__.Notify(470213)
		if colinfo.IsColIDSystemColumn(id) {
			__antithesis_instrumentation__.Notify(470214)
			return true
		} else {
			__antithesis_instrumentation__.Notify(470215)
		}
	}
	__antithesis_instrumentation__.Notify(470212)
	return false
}

func constructOpaque(metadata opt.OpaqueMetadata) (planNode, error) {
	__antithesis_instrumentation__.Notify(470216)
	o, ok := metadata.(*opaqueMetadata)
	if !ok {
		__antithesis_instrumentation__.Notify(470218)
		return nil, errors.AssertionFailedf("unexpected OpaqueMetadata object type %T", metadata)
	} else {
		__antithesis_instrumentation__.Notify(470219)
	}
	__antithesis_instrumentation__.Notify(470217)
	return o.plan, nil
}

func convertFastIntSetToUint32Slice(colIdxs util.FastIntSet) []uint32 {
	__antithesis_instrumentation__.Notify(470220)
	cols := make([]uint32, 0, colIdxs.Len())
	for i, ok := colIdxs.Next(0); ok; i, ok = colIdxs.Next(i + 1) {
		__antithesis_instrumentation__.Notify(470222)
		cols = append(cols, uint32(i))
	}
	__antithesis_instrumentation__.Notify(470221)
	return cols
}
