package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/inverted"
	"github.com/cockroachdb/cockroach/pkg/sql/opt"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/constraint"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec/explain"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treewindow"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/errors"
)

type execFactory struct {
	planner *planner

	isExplain bool
}

var _ exec.Factory = &execFactory{}

func newExecFactory(p *planner) *execFactory {
	__antithesis_instrumentation__.Notify(551395)
	return &execFactory{
		planner: p,
	}
}

func (ef *execFactory) ConstructValues(
	rows [][]tree.TypedExpr, cols colinfo.ResultColumns,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551396)
	if len(cols) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(551399)
		return len(rows) == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(551400)
		return &unaryNode{}, nil
	} else {
		__antithesis_instrumentation__.Notify(551401)
	}
	__antithesis_instrumentation__.Notify(551397)
	if len(rows) == 0 {
		__antithesis_instrumentation__.Notify(551402)
		return &zeroNode{columns: cols}, nil
	} else {
		__antithesis_instrumentation__.Notify(551403)
	}
	__antithesis_instrumentation__.Notify(551398)
	return &valuesNode{
		columns:          cols,
		tuples:           rows,
		specifiedInQuery: true,
	}, nil
}

func (ef *execFactory) ConstructScan(
	table cat.Table, index cat.Index, params exec.ScanParams, reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551404)
	if table.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(551412)
		return ef.constructVirtualScan(table, index, params, reqOrdering)
	} else {
		__antithesis_instrumentation__.Notify(551413)
	}
	__antithesis_instrumentation__.Notify(551405)

	tabDesc := table.(*optTable).desc
	idx := index.(*optIndex).idx

	scan := ef.planner.Scan()
	colCfg := makeScanColumnsConfig(table, params.NeededCols)

	ctx := ef.planner.extendedEvalCtx.Ctx()
	if err := scan.initTable(ctx, ef.planner, tabDesc, colCfg); err != nil {
		__antithesis_instrumentation__.Notify(551414)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551415)
	}
	__antithesis_instrumentation__.Notify(551406)

	if params.IndexConstraint != nil && func() bool {
		__antithesis_instrumentation__.Notify(551416)
		return params.IndexConstraint.IsContradiction() == true
	}() == true {
		__antithesis_instrumentation__.Notify(551417)
		return newZeroNode(scan.resultColumns), nil
	} else {
		__antithesis_instrumentation__.Notify(551418)
	}
	__antithesis_instrumentation__.Notify(551407)

	scan.index = idx
	scan.hardLimit = params.HardLimit
	scan.softLimit = params.SoftLimit

	scan.reverse = params.Reverse
	scan.parallelize = params.Parallelize
	var err error
	scan.spans, err = generateScanSpans(ef.planner.EvalContext(), ef.planner.ExecCfg().Codec, tabDesc, idx, params)
	if err != nil {
		__antithesis_instrumentation__.Notify(551419)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551420)
	}
	__antithesis_instrumentation__.Notify(551408)

	scan.isFull = len(scan.spans) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(551421)
		return scan.spans[0].EqualValue(
			scan.desc.IndexSpan(ef.planner.ExecCfg().Codec, scan.index.GetID()),
		) == true
	}() == true
	if err = colCfg.assertValidReqOrdering(reqOrdering); err != nil {
		__antithesis_instrumentation__.Notify(551422)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551423)
	}
	__antithesis_instrumentation__.Notify(551409)
	scan.reqOrdering = ReqOrdering(reqOrdering)
	scan.estimatedRowCount = uint64(params.EstimatedRowCount)
	if params.Locking != nil {
		__antithesis_instrumentation__.Notify(551424)
		scan.lockingStrength = descpb.ToScanLockingStrength(params.Locking.Strength)
		scan.lockingWaitPolicy = descpb.ToScanLockingWaitPolicy(params.Locking.WaitPolicy)
	} else {
		__antithesis_instrumentation__.Notify(551425)
	}
	__antithesis_instrumentation__.Notify(551410)
	scan.localityOptimized = params.LocalityOptimized
	if !ef.isExplain {
		__antithesis_instrumentation__.Notify(551426)
		idxUsageKey := roachpb.IndexUsageKey{
			TableID: roachpb.TableID(tabDesc.GetID()),
			IndexID: roachpb.IndexID(idx.GetID()),
		}
		ef.planner.extendedEvalCtx.indexUsageStats.RecordRead(idxUsageKey)
	} else {
		__antithesis_instrumentation__.Notify(551427)
	}
	__antithesis_instrumentation__.Notify(551411)

	return scan, nil
}

func generateScanSpans(
	evalCtx *tree.EvalContext,
	codec keys.SQLCodec,
	tabDesc catalog.TableDescriptor,
	index catalog.Index,
	params exec.ScanParams,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(551428)
	var sb span.Builder
	sb.Init(evalCtx, codec, tabDesc, index)
	if params.InvertedConstraint != nil {
		__antithesis_instrumentation__.Notify(551430)
		return sb.SpansFromInvertedSpans(params.InvertedConstraint, params.IndexConstraint, nil)
	} else {
		__antithesis_instrumentation__.Notify(551431)
	}
	__antithesis_instrumentation__.Notify(551429)
	splitter := span.MakeSplitter(tabDesc, index, params.NeededCols)
	return sb.SpansFromConstraint(params.IndexConstraint, splitter)
}

func (ef *execFactory) constructVirtualScan(
	table cat.Table, index cat.Index, params exec.ScanParams, reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551432)
	return constructVirtualScan(
		ef, ef.planner, table, index, params, reqOrdering,
		func(d *delayedNode) (exec.Node, error) { __antithesis_instrumentation__.Notify(551433); return d, nil },
	)
}

func asDataSource(n exec.Node) planDataSource {
	__antithesis_instrumentation__.Notify(551434)
	plan := n.(planNode)
	return planDataSource{
		columns: planColumns(plan),
		plan:    plan,
	}
}

func (ef *execFactory) ConstructFilter(
	n exec.Node, filter tree.TypedExpr, reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551435)

	src := asDataSource(n)
	f := &filterNode{
		source: src,
	}
	f.ivarHelper = tree.MakeIndexedVarHelper(f, len(src.columns))
	f.filter = f.ivarHelper.Rebind(filter)
	if f.filter == nil {
		__antithesis_instrumentation__.Notify(551438)

		return n, nil
	} else {
		__antithesis_instrumentation__.Notify(551439)
	}
	__antithesis_instrumentation__.Notify(551436)
	f.reqOrdering = ReqOrdering(reqOrdering)

	if spool, ok := f.source.plan.(*spoolNode); ok {
		__antithesis_instrumentation__.Notify(551440)
		f.source.plan = spool.source
		spool.source = f
		return spool, nil
	} else {
		__antithesis_instrumentation__.Notify(551441)
	}
	__antithesis_instrumentation__.Notify(551437)
	return f, nil
}

func (ef *execFactory) ConstructInvertedFilter(
	n exec.Node,
	invFilter *inverted.SpanExpression,
	preFiltererExpr tree.TypedExpr,
	preFiltererType *types.T,
	invColumn exec.NodeColumnOrdinal,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551442)
	inputCols := planColumns(n.(planNode))
	columns := make(colinfo.ResultColumns, len(inputCols))
	copy(columns, inputCols)
	n = &invertedFilterNode{
		input:           n.(planNode),
		expression:      invFilter,
		preFiltererExpr: preFiltererExpr,
		preFiltererType: preFiltererType,
		invColumn:       int(invColumn),
		resultColumns:   columns,
	}
	return n, nil
}

func (ef *execFactory) ConstructSimpleProject(
	n exec.Node, cols []exec.NodeColumnOrdinal, reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551443)
	return constructSimpleProjectForPlanNode(n.(planNode), cols, nil, reqOrdering)
}

func constructSimpleProjectForPlanNode(
	n planNode, cols []exec.NodeColumnOrdinal, colNames []string, reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551444)

	if r, ok := n.(*renderNode); ok && func() bool {
		__antithesis_instrumentation__.Notify(551448)
		return !hasDuplicates(cols) == true
	}() == true {
		__antithesis_instrumentation__.Notify(551449)
		oldCols, oldRenders := r.columns, r.render
		r.columns = make(colinfo.ResultColumns, len(cols))
		r.render = make([]tree.TypedExpr, len(cols))
		for i, ord := range cols {
			__antithesis_instrumentation__.Notify(551451)
			r.columns[i] = oldCols[ord]
			if colNames != nil {
				__antithesis_instrumentation__.Notify(551453)
				r.columns[i].Name = colNames[i]
			} else {
				__antithesis_instrumentation__.Notify(551454)
			}
			__antithesis_instrumentation__.Notify(551452)
			r.render[i] = oldRenders[ord]
		}
		__antithesis_instrumentation__.Notify(551450)
		r.reqOrdering = ReqOrdering(reqOrdering)
		return r, nil
	} else {
		__antithesis_instrumentation__.Notify(551455)
	}
	__antithesis_instrumentation__.Notify(551445)
	inputCols := planColumns(n)
	var rb renderBuilder
	rb.init(n, reqOrdering)

	exprs := make(tree.TypedExprs, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(551456)
		exprs[i] = rb.r.ivarHelper.IndexedVar(int(col))
	}
	__antithesis_instrumentation__.Notify(551446)
	var resultTypes []*types.T
	if colNames != nil {
		__antithesis_instrumentation__.Notify(551457)

		resultTypes = make([]*types.T, len(cols))
		for i := range exprs {
			__antithesis_instrumentation__.Notify(551458)
			resultTypes[i] = exprs[i].ResolvedType()
		}
	} else {
		__antithesis_instrumentation__.Notify(551459)
	}
	__antithesis_instrumentation__.Notify(551447)
	resultCols := getResultColumnsForSimpleProject(cols, colNames, resultTypes, inputCols)
	rb.setOutput(exprs, resultCols)
	return rb.res, nil
}

func hasDuplicates(cols []exec.NodeColumnOrdinal) bool {
	__antithesis_instrumentation__.Notify(551460)
	var set util.FastIntSet
	for _, c := range cols {
		__antithesis_instrumentation__.Notify(551462)
		if set.Contains(int(c)) {
			__antithesis_instrumentation__.Notify(551464)
			return true
		} else {
			__antithesis_instrumentation__.Notify(551465)
		}
		__antithesis_instrumentation__.Notify(551463)
		set.Add(int(c))
	}
	__antithesis_instrumentation__.Notify(551461)
	return false
}

func (ef *execFactory) ConstructSerializingProject(
	n exec.Node, cols []exec.NodeColumnOrdinal, colNames []string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551466)
	node := n.(planNode)

	if len(cols) == len(planColumns(node)) {
		__antithesis_instrumentation__.Notify(551470)
		identity := true
		for i := range cols {
			__antithesis_instrumentation__.Notify(551472)
			if cols[i] != exec.NodeColumnOrdinal(i) {
				__antithesis_instrumentation__.Notify(551473)
				identity = false
				break
			} else {
				__antithesis_instrumentation__.Notify(551474)
			}
		}
		__antithesis_instrumentation__.Notify(551471)
		if identity {
			__antithesis_instrumentation__.Notify(551475)
			inputCols := planMutableColumns(node)
			for i := range inputCols {
				__antithesis_instrumentation__.Notify(551478)
				inputCols[i].Name = colNames[i]
			}
			__antithesis_instrumentation__.Notify(551476)

			if r, ok := n.(*renderNode); ok {
				__antithesis_instrumentation__.Notify(551479)
				r.serialize = true
			} else {
				__antithesis_instrumentation__.Notify(551480)
			}
			__antithesis_instrumentation__.Notify(551477)
			return n, nil
		} else {
			__antithesis_instrumentation__.Notify(551481)
		}
	} else {
		__antithesis_instrumentation__.Notify(551482)
	}
	__antithesis_instrumentation__.Notify(551467)
	res, err := constructSimpleProjectForPlanNode(node, cols, colNames, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(551483)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551484)
	}
	__antithesis_instrumentation__.Notify(551468)
	switch r := res.(type) {
	case *renderNode:
		__antithesis_instrumentation__.Notify(551485)
		r.serialize = true
	case *spoolNode:
		__antithesis_instrumentation__.Notify(551486)

	default:
		__antithesis_instrumentation__.Notify(551487)
		return nil, errors.AssertionFailedf("unexpected planNode type %T in ConstructSerializingProject", res)
	}
	__antithesis_instrumentation__.Notify(551469)
	return res, nil
}

func (ef *execFactory) ConstructRender(
	n exec.Node,
	columns colinfo.ResultColumns,
	exprs tree.TypedExprs,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551488)
	var rb renderBuilder
	rb.init(n, reqOrdering)
	for i, expr := range exprs {
		__antithesis_instrumentation__.Notify(551490)
		exprs[i] = rb.r.ivarHelper.Rebind(expr)
	}
	__antithesis_instrumentation__.Notify(551489)
	rb.setOutput(exprs, columns)
	return rb.res, nil
}

func (ef *execFactory) ConstructHashJoin(
	joinType descpb.JoinType,
	left, right exec.Node,
	leftEqCols, rightEqCols []exec.NodeColumnOrdinal,
	leftEqColsAreKey, rightEqColsAreKey bool,
	extraOnCond tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551491)
	p := ef.planner
	leftSrc := asDataSource(left)
	rightSrc := asDataSource(right)
	pred := makePredicate(joinType, leftSrc.columns, rightSrc.columns)

	numEqCols := len(leftEqCols)
	pred.leftEqualityIndices = leftEqCols
	pred.rightEqualityIndices = rightEqCols
	nameBuf := make(tree.NameList, 2*numEqCols)
	pred.leftColNames = nameBuf[:numEqCols:numEqCols]
	pred.rightColNames = nameBuf[numEqCols:]

	for i := range leftEqCols {
		__antithesis_instrumentation__.Notify(551493)
		pred.leftColNames[i] = tree.Name(leftSrc.columns[leftEqCols[i]].Name)
		pred.rightColNames[i] = tree.Name(rightSrc.columns[rightEqCols[i]].Name)
	}
	__antithesis_instrumentation__.Notify(551492)
	pred.leftEqKey = leftEqColsAreKey
	pred.rightEqKey = rightEqColsAreKey

	pred.onCond = pred.iVarHelper.Rebind(extraOnCond)

	return p.makeJoinNode(leftSrc, rightSrc, pred), nil
}

func (ef *execFactory) ConstructApplyJoin(
	joinType descpb.JoinType,
	left exec.Node,
	rightColumns colinfo.ResultColumns,
	onCond tree.TypedExpr,
	planRightSideFn exec.ApplyJoinPlanRightSideFn,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551494)
	leftSrc := asDataSource(left)
	pred := makePredicate(joinType, leftSrc.columns, rightColumns)
	pred.onCond = pred.iVarHelper.Rebind(onCond)
	return newApplyJoinNode(joinType, leftSrc, rightColumns, pred, planRightSideFn)
}

func (ef *execFactory) ConstructMergeJoin(
	joinType descpb.JoinType,
	left, right exec.Node,
	onCond tree.TypedExpr,
	leftOrdering, rightOrdering colinfo.ColumnOrdering,
	reqOrdering exec.OutputOrdering,
	leftEqColsAreKey, rightEqColsAreKey bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551495)
	var err error
	p := ef.planner
	leftSrc := asDataSource(left)
	rightSrc := asDataSource(right)
	pred := makePredicate(joinType, leftSrc.columns, rightSrc.columns)
	pred.onCond = pred.iVarHelper.Rebind(onCond)
	node := p.makeJoinNode(leftSrc, rightSrc, pred)
	pred.leftEqKey = leftEqColsAreKey
	pred.rightEqKey = rightEqColsAreKey

	pred.leftEqualityIndices, pred.rightEqualityIndices, node.mergeJoinOrdering, err = getEqualityIndicesAndMergeJoinOrdering(leftOrdering, rightOrdering)
	if err != nil {
		__antithesis_instrumentation__.Notify(551498)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551499)
	}
	__antithesis_instrumentation__.Notify(551496)
	n := len(leftOrdering)
	pred.leftColNames = make(tree.NameList, n)
	pred.rightColNames = make(tree.NameList, n)
	for i := 0; i < n; i++ {
		__antithesis_instrumentation__.Notify(551500)
		leftColIdx, rightColIdx := leftOrdering[i].ColIdx, rightOrdering[i].ColIdx
		pred.leftColNames[i] = tree.Name(leftSrc.columns[leftColIdx].Name)
		pred.rightColNames[i] = tree.Name(rightSrc.columns[rightColIdx].Name)
	}
	__antithesis_instrumentation__.Notify(551497)

	node.reqOrdering = ReqOrdering(reqOrdering)

	return node, nil
}

func (ef *execFactory) ConstructScalarGroupBy(
	input exec.Node, aggregations []exec.AggInfo,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551501)

	var inputCols colinfo.ResultColumns
	var groupCols []exec.NodeColumnOrdinal
	n := &groupNode{
		plan:     input.(planNode),
		funcs:    make([]*aggregateFuncHolder, 0, len(aggregations)),
		columns:  getResultColumnsForGroupBy(inputCols, groupCols, aggregations),
		isScalar: true,
	}
	if err := ef.addAggregations(n, aggregations); err != nil {
		__antithesis_instrumentation__.Notify(551503)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551504)
	}
	__antithesis_instrumentation__.Notify(551502)
	return n, nil
}

func (ef *execFactory) ConstructGroupBy(
	input exec.Node,
	groupCols []exec.NodeColumnOrdinal,
	groupColOrdering colinfo.ColumnOrdering,
	aggregations []exec.AggInfo,
	reqOrdering exec.OutputOrdering,
	groupingOrderType exec.GroupingOrderType,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551505)
	inputPlan := input.(planNode)
	inputCols := planColumns(inputPlan)

	n := &groupNode{
		plan:             inputPlan,
		funcs:            make([]*aggregateFuncHolder, 0, len(groupCols)+len(aggregations)),
		columns:          getResultColumnsForGroupBy(inputCols, groupCols, aggregations),
		groupCols:        convertNodeOrdinalsToInts(groupCols),
		groupColOrdering: groupColOrdering,
		isScalar:         false,
		reqOrdering:      ReqOrdering(reqOrdering),
	}
	for _, col := range n.groupCols {
		__antithesis_instrumentation__.Notify(551508)

		f := newAggregateFuncHolder(
			builtins.AnyNotNull,
			[]int{col},
			nil,
			false,
		)
		n.funcs = append(n.funcs, f)
	}
	__antithesis_instrumentation__.Notify(551506)
	if err := ef.addAggregations(n, aggregations); err != nil {
		__antithesis_instrumentation__.Notify(551509)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551510)
	}
	__antithesis_instrumentation__.Notify(551507)
	return n, nil
}

func (ef *execFactory) addAggregations(n *groupNode, aggregations []exec.AggInfo) error {
	__antithesis_instrumentation__.Notify(551511)
	for i := range aggregations {
		__antithesis_instrumentation__.Notify(551513)
		agg := &aggregations[i]
		renderIdxs := convertNodeOrdinalsToInts(agg.ArgCols)

		f := newAggregateFuncHolder(
			agg.FuncName,
			renderIdxs,
			agg.ConstArgs,
			agg.Distinct,
		)
		f.filterRenderIdx = int(agg.Filter)

		n.funcs = append(n.funcs, f)
	}
	__antithesis_instrumentation__.Notify(551512)
	return nil
}

func (ef *execFactory) ConstructDistinct(
	input exec.Node,
	distinctCols, orderedCols exec.NodeColumnOrdinalSet,
	reqOrdering exec.OutputOrdering,
	nullsAreDistinct bool,
	errorOnDup string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551514)
	return &distinctNode{
		plan:              input.(planNode),
		distinctOnColIdxs: distinctCols,
		columnsInOrder:    orderedCols,
		reqOrdering:       ReqOrdering(reqOrdering),
		nullsAreDistinct:  nullsAreDistinct,
		errorOnDup:        errorOnDup,
	}, nil
}

func (ef *execFactory) ConstructHashSetOp(
	typ tree.UnionType, all bool, left, right exec.Node,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551515)
	return ef.planner.newUnionNode(
		typ, all, left.(planNode), right.(planNode), nil, nil, 0,
	)
}

func (ef *execFactory) ConstructStreamingSetOp(
	typ tree.UnionType,
	all bool,
	left, right exec.Node,
	streamingOrdering colinfo.ColumnOrdering,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551516)
	return ef.planner.newUnionNode(
		typ,
		all,
		left.(planNode),
		right.(planNode),
		streamingOrdering,
		ReqOrdering(reqOrdering),
		0,
	)
}

func (ef *execFactory) ConstructUnionAll(
	left, right exec.Node, reqOrdering exec.OutputOrdering, hardLimit uint64,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551517)
	return ef.planner.newUnionNode(
		tree.UnionOp,
		true,
		left.(planNode),
		right.(planNode),
		colinfo.ColumnOrdering(reqOrdering),
		ReqOrdering(reqOrdering),
		hardLimit,
	)
}

func (ef *execFactory) ConstructSort(
	input exec.Node, ordering exec.OutputOrdering, alreadyOrderedPrefix int,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551518)
	return &sortNode{
		plan:                 input.(planNode),
		ordering:             colinfo.ColumnOrdering(ordering),
		alreadyOrderedPrefix: alreadyOrderedPrefix,
	}, nil
}

func (ef *execFactory) ConstructOrdinality(input exec.Node, colName string) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551519)
	plan := input.(planNode)
	inputColumns := planColumns(plan)
	cols := make(colinfo.ResultColumns, len(inputColumns)+1)
	copy(cols, inputColumns)
	cols[len(cols)-1] = colinfo.ResultColumn{
		Name: colName,
		Typ:  types.Int,
	}
	return &ordinalityNode{
		source:  plan,
		columns: cols,
	}, nil
}

func (ef *execFactory) ConstructIndexJoin(
	input exec.Node,
	table cat.Table,
	keyCols []exec.NodeColumnOrdinal,
	tableCols exec.TableColumnOrdinalSet,
	reqOrdering exec.OutputOrdering,
	limitHint int64,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551520)
	tabDesc := table.(*optTable).desc
	colCfg := makeScanColumnsConfig(table, tableCols)
	cols := makeColList(table, tableCols)

	tableScan := ef.planner.Scan()

	ctx := ef.planner.extendedEvalCtx.Ctx()
	if err := tableScan.initTable(ctx, ef.planner, tabDesc, colCfg); err != nil {
		__antithesis_instrumentation__.Notify(551524)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551525)
	}
	__antithesis_instrumentation__.Notify(551521)

	idx := tabDesc.GetPrimaryIndex()
	tableScan.index = idx
	tableScan.disableBatchLimit()

	if !ef.isExplain {
		__antithesis_instrumentation__.Notify(551526)
		idxUsageKey := roachpb.IndexUsageKey{
			TableID: roachpb.TableID(tabDesc.GetID()),
			IndexID: roachpb.IndexID(idx.GetID()),
		}
		ef.planner.extendedEvalCtx.indexUsageStats.RecordRead(idxUsageKey)
	} else {
		__antithesis_instrumentation__.Notify(551527)
	}
	__antithesis_instrumentation__.Notify(551522)

	n := &indexJoinNode{
		input:         input.(planNode),
		table:         tableScan,
		cols:          cols,
		resultColumns: colinfo.ResultColumnsFromColumns(tabDesc.GetID(), cols),
		reqOrdering:   ReqOrdering(reqOrdering),
		limitHint:     limitHint,
	}

	n.keyCols = make([]int, len(keyCols))
	for i, c := range keyCols {
		__antithesis_instrumentation__.Notify(551528)
		n.keyCols[i] = int(c)
	}
	__antithesis_instrumentation__.Notify(551523)

	return n, nil
}

func (ef *execFactory) ConstructLookupJoin(
	joinType descpb.JoinType,
	input exec.Node,
	table cat.Table,
	index cat.Index,
	eqCols []exec.NodeColumnOrdinal,
	eqColsAreKey bool,
	lookupExpr tree.TypedExpr,
	remoteLookupExpr tree.TypedExpr,
	lookupCols exec.TableColumnOrdinalSet,
	onCond tree.TypedExpr,
	isFirstJoinInPairedJoiner bool,
	isSecondJoinInPairedJoiner bool,
	reqOrdering exec.OutputOrdering,
	locking *tree.LockingItem,
	limitHint int64,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551529)
	if table.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(551539)
		return ef.constructVirtualTableLookupJoin(joinType, input, table, index, eqCols, lookupCols, onCond)
	} else {
		__antithesis_instrumentation__.Notify(551540)
	}
	__antithesis_instrumentation__.Notify(551530)
	tabDesc := table.(*optTable).desc
	idx := index.(*optIndex).idx
	colCfg := makeScanColumnsConfig(table, lookupCols)
	tableScan := ef.planner.Scan()

	ctx := ef.planner.extendedEvalCtx.Ctx()
	if err := tableScan.initTable(ctx, ef.planner, tabDesc, colCfg); err != nil {
		__antithesis_instrumentation__.Notify(551541)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551542)
	}
	__antithesis_instrumentation__.Notify(551531)

	tableScan.index = idx
	if locking != nil {
		__antithesis_instrumentation__.Notify(551543)
		tableScan.lockingStrength = descpb.ToScanLockingStrength(locking.Strength)
		tableScan.lockingWaitPolicy = descpb.ToScanLockingWaitPolicy(locking.WaitPolicy)
	} else {
		__antithesis_instrumentation__.Notify(551544)
	}
	__antithesis_instrumentation__.Notify(551532)

	if !ef.isExplain {
		__antithesis_instrumentation__.Notify(551545)
		idxUsageKey := roachpb.IndexUsageKey{
			TableID: roachpb.TableID(tabDesc.GetID()),
			IndexID: roachpb.IndexID(idx.GetID()),
		}
		ef.planner.extendedEvalCtx.indexUsageStats.RecordRead(idxUsageKey)
	} else {
		__antithesis_instrumentation__.Notify(551546)
	}
	__antithesis_instrumentation__.Notify(551533)

	n := &lookupJoinNode{
		input:                      input.(planNode),
		table:                      tableScan,
		joinType:                   joinType,
		eqColsAreKey:               eqColsAreKey,
		isFirstJoinInPairedJoiner:  isFirstJoinInPairedJoiner,
		isSecondJoinInPairedJoiner: isSecondJoinInPairedJoiner,
		reqOrdering:                ReqOrdering(reqOrdering),
		limitHint:                  limitHint,
	}
	n.eqCols = make([]int, len(eqCols))
	for i, c := range eqCols {
		__antithesis_instrumentation__.Notify(551547)
		n.eqCols[i] = int(c)
	}
	__antithesis_instrumentation__.Notify(551534)
	pred := makePredicate(joinType, planColumns(input.(planNode)), planColumns(tableScan))
	if lookupExpr != nil {
		__antithesis_instrumentation__.Notify(551548)
		n.lookupExpr = pred.iVarHelper.Rebind(lookupExpr)
	} else {
		__antithesis_instrumentation__.Notify(551549)
	}
	__antithesis_instrumentation__.Notify(551535)
	if remoteLookupExpr != nil {
		__antithesis_instrumentation__.Notify(551550)
		n.remoteLookupExpr = pred.iVarHelper.Rebind(remoteLookupExpr)
	} else {
		__antithesis_instrumentation__.Notify(551551)
	}
	__antithesis_instrumentation__.Notify(551536)
	if onCond != nil && func() bool {
		__antithesis_instrumentation__.Notify(551552)
		return onCond != tree.DBoolTrue == true
	}() == true {
		__antithesis_instrumentation__.Notify(551553)
		n.onCond = pred.iVarHelper.Rebind(onCond)
	} else {
		__antithesis_instrumentation__.Notify(551554)
	}
	__antithesis_instrumentation__.Notify(551537)
	n.columns = pred.cols
	if isFirstJoinInPairedJoiner {
		__antithesis_instrumentation__.Notify(551555)
		n.columns = append(n.columns, colinfo.ResultColumn{Name: "cont", Typ: types.Bool})
	} else {
		__antithesis_instrumentation__.Notify(551556)
	}
	__antithesis_instrumentation__.Notify(551538)

	return n, nil
}

func (ef *execFactory) constructVirtualTableLookupJoin(
	joinType descpb.JoinType,
	input exec.Node,
	table cat.Table,
	index cat.Index,
	eqCols []exec.NodeColumnOrdinal,
	lookupCols exec.TableColumnOrdinalSet,
	onCond tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551557)
	tn := &table.(*optVirtualTable).name
	virtual, err := ef.planner.getVirtualTabler().getVirtualTableEntry(tn)
	if err != nil {
		__antithesis_instrumentation__.Notify(551564)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551565)
	}
	__antithesis_instrumentation__.Notify(551558)
	if !canQueryVirtualTable(ef.planner.EvalContext(), virtual) {
		__antithesis_instrumentation__.Notify(551566)
		return nil, newUnimplementedVirtualTableError(tn.Schema(), tn.Table())
	} else {
		__antithesis_instrumentation__.Notify(551567)
	}
	__antithesis_instrumentation__.Notify(551559)
	if len(eqCols) > 1 {
		__antithesis_instrumentation__.Notify(551568)
		return nil, errors.AssertionFailedf("vtable indexes with more than one column aren't supported yet")
	} else {
		__antithesis_instrumentation__.Notify(551569)
	}
	__antithesis_instrumentation__.Notify(551560)

	if lookupCols.Contains(0) {
		__antithesis_instrumentation__.Notify(551570)
		return nil, errors.Errorf("use of %s column not allowed.", table.Column(0).ColName())
	} else {
		__antithesis_instrumentation__.Notify(551571)
	}
	__antithesis_instrumentation__.Notify(551561)
	idx := index.(*optVirtualIndex).idx
	tableDesc := table.(*optVirtualTable).desc

	inputCols := planColumns(input.(planNode))

	if onCond == tree.DBoolTrue {
		__antithesis_instrumentation__.Notify(551572)
		onCond = nil
	} else {
		__antithesis_instrumentation__.Notify(551573)
	}
	__antithesis_instrumentation__.Notify(551562)

	var tableScan scanNode

	colCfg := makeScanColumnsConfig(table, lookupCols)
	ctx := ef.planner.extendedEvalCtx.Ctx()
	if err := tableScan.initTable(ctx, ef.planner, tableDesc, colCfg); err != nil {
		__antithesis_instrumentation__.Notify(551574)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551575)
	}
	__antithesis_instrumentation__.Notify(551563)
	tableScan.index = idx
	vtableCols := colinfo.ResultColumnsFromColumns(tableDesc.GetID(), tableDesc.PublicColumns())
	projectedVtableCols := planColumns(&tableScan)
	outputCols := make(colinfo.ResultColumns, 0, len(inputCols)+len(projectedVtableCols))
	outputCols = append(outputCols, inputCols...)
	outputCols = append(outputCols, projectedVtableCols...)

	pred := makePredicate(joinType, inputCols, projectedVtableCols)
	pred.onCond = pred.iVarHelper.Rebind(onCond)
	n := &vTableLookupJoinNode{
		input:             input.(planNode),
		joinType:          joinType,
		virtualTableEntry: virtual,
		dbName:            tn.Catalog(),
		table:             tableDesc,
		index:             idx,
		eqCol:             int(eqCols[0]),
		inputCols:         inputCols,
		vtableCols:        vtableCols,
		lookupCols:        lookupCols,
		columns:           outputCols,
		pred:              pred,
	}
	return n, nil
}

func (ef *execFactory) ConstructInvertedJoin(
	joinType descpb.JoinType,
	invertedExpr tree.TypedExpr,
	input exec.Node,
	table cat.Table,
	index cat.Index,
	prefixEqCols []exec.NodeColumnOrdinal,
	lookupCols exec.TableColumnOrdinalSet,
	onCond tree.TypedExpr,
	isFirstJoinInPairedJoiner bool,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551576)
	tabDesc := table.(*optTable).desc
	idx := index.(*optIndex).idx

	colCfg := makeScanColumnsConfig(table, lookupCols)
	tableScan := ef.planner.Scan()

	ctx := ef.planner.extendedEvalCtx.Ctx()
	if err := tableScan.initTable(ctx, ef.planner, tabDesc, colCfg); err != nil {
		__antithesis_instrumentation__.Notify(551584)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551585)
	}
	__antithesis_instrumentation__.Notify(551577)
	tableScan.index = idx

	if !ef.isExplain {
		__antithesis_instrumentation__.Notify(551586)
		idxUsageKey := roachpb.IndexUsageKey{
			TableID: roachpb.TableID(tabDesc.GetID()),
			IndexID: roachpb.IndexID(idx.GetID()),
		}
		ef.planner.extendedEvalCtx.indexUsageStats.RecordRead(idxUsageKey)
	} else {
		__antithesis_instrumentation__.Notify(551587)
	}
	__antithesis_instrumentation__.Notify(551578)

	n := &invertedJoinNode{
		input:                     input.(planNode),
		table:                     tableScan,
		joinType:                  joinType,
		invertedExpr:              invertedExpr,
		isFirstJoinInPairedJoiner: isFirstJoinInPairedJoiner,
		reqOrdering:               ReqOrdering(reqOrdering),
	}
	if len(prefixEqCols) > 0 {
		__antithesis_instrumentation__.Notify(551588)
		n.prefixEqCols = make([]int, len(prefixEqCols))
		for i, c := range prefixEqCols {
			__antithesis_instrumentation__.Notify(551589)
			n.prefixEqCols[i] = int(c)
		}
	} else {
		__antithesis_instrumentation__.Notify(551590)
	}
	__antithesis_instrumentation__.Notify(551579)
	if onCond != nil && func() bool {
		__antithesis_instrumentation__.Notify(551591)
		return onCond != tree.DBoolTrue == true
	}() == true {
		__antithesis_instrumentation__.Notify(551592)
		n.onExpr = onCond
	} else {
		__antithesis_instrumentation__.Notify(551593)
	}
	__antithesis_instrumentation__.Notify(551580)

	inputCols := planColumns(input.(planNode))
	var scanCols colinfo.ResultColumns
	if joinType.ShouldIncludeRightColsInOutput() {
		__antithesis_instrumentation__.Notify(551594)
		scanCols = planColumns(tableScan)
	} else {
		__antithesis_instrumentation__.Notify(551595)
	}
	__antithesis_instrumentation__.Notify(551581)
	numCols := len(inputCols) + len(scanCols)
	if isFirstJoinInPairedJoiner {
		__antithesis_instrumentation__.Notify(551596)
		numCols++
	} else {
		__antithesis_instrumentation__.Notify(551597)
	}
	__antithesis_instrumentation__.Notify(551582)
	n.columns = make(colinfo.ResultColumns, 0, numCols)
	n.columns = append(n.columns, inputCols...)
	n.columns = append(n.columns, scanCols...)
	if isFirstJoinInPairedJoiner {
		__antithesis_instrumentation__.Notify(551598)
		n.columns = append(n.columns, colinfo.ResultColumn{Name: "cont", Typ: types.Bool})
	} else {
		__antithesis_instrumentation__.Notify(551599)
	}
	__antithesis_instrumentation__.Notify(551583)
	return n, nil
}

func (ef *execFactory) constructScanForZigzag(
	index catalog.Index, tableDesc catalog.TableDescriptor, cols exec.TableColumnOrdinalSet,
) (*scanNode, error) {
	__antithesis_instrumentation__.Notify(551600)

	colCfg := scanColumnsConfig{
		wantedColumns: make([]tree.ColumnID, 0, cols.Len()),
	}

	for c, ok := cols.Next(0); ok; c, ok = cols.Next(c + 1) {
		__antithesis_instrumentation__.Notify(551604)
		colCfg.wantedColumns = append(colCfg.wantedColumns, tableDesc.PublicColumns()[c].GetID())
	}
	__antithesis_instrumentation__.Notify(551601)

	scan := ef.planner.Scan()
	ctx := ef.planner.extendedEvalCtx.Ctx()
	if err := scan.initTable(ctx, ef.planner, tableDesc, colCfg); err != nil {
		__antithesis_instrumentation__.Notify(551605)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551606)
	}
	__antithesis_instrumentation__.Notify(551602)

	if !ef.isExplain {
		__antithesis_instrumentation__.Notify(551607)
		idxUsageKey := roachpb.IndexUsageKey{
			TableID: roachpb.TableID(tableDesc.GetID()),
			IndexID: roachpb.IndexID(index.GetID()),
		}
		ef.planner.extendedEvalCtx.indexUsageStats.RecordRead(idxUsageKey)
	} else {
		__antithesis_instrumentation__.Notify(551608)
	}
	__antithesis_instrumentation__.Notify(551603)

	scan.index = index

	return scan, nil
}

func (ef *execFactory) ConstructZigzagJoin(
	leftTable cat.Table,
	leftIndex cat.Index,
	leftCols exec.TableColumnOrdinalSet,
	leftFixedVals []tree.TypedExpr,
	leftEqCols []exec.TableColumnOrdinal,
	rightTable cat.Table,
	rightIndex cat.Index,
	rightCols exec.TableColumnOrdinalSet,
	rightFixedVals []tree.TypedExpr,
	rightEqCols []exec.TableColumnOrdinal,
	onCond tree.TypedExpr,
	reqOrdering exec.OutputOrdering,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551609)
	leftIdx := leftIndex.(*optIndex).idx
	leftTabDesc := leftTable.(*optTable).desc
	rightIdx := rightIndex.(*optIndex).idx
	rightTabDesc := rightTable.(*optTable).desc

	leftScan, err := ef.constructScanForZigzag(leftIdx, leftTabDesc, leftCols)
	if err != nil {
		__antithesis_instrumentation__.Notify(551616)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551617)
	}
	__antithesis_instrumentation__.Notify(551610)
	rightScan, err := ef.constructScanForZigzag(rightIdx, rightTabDesc, rightCols)
	if err != nil {
		__antithesis_instrumentation__.Notify(551618)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551619)
	}
	__antithesis_instrumentation__.Notify(551611)

	n := &zigzagJoinNode{
		reqOrdering: ReqOrdering(reqOrdering),
	}
	if onCond != nil && func() bool {
		__antithesis_instrumentation__.Notify(551620)
		return onCond != tree.DBoolTrue == true
	}() == true {
		__antithesis_instrumentation__.Notify(551621)
		n.onCond = onCond
	} else {
		__antithesis_instrumentation__.Notify(551622)
	}
	__antithesis_instrumentation__.Notify(551612)
	n.sides = make([]zigzagJoinSide, 2)
	n.sides[0].scan = leftScan
	n.sides[1].scan = rightScan
	n.sides[0].eqCols = make([]int, len(leftEqCols))
	n.sides[1].eqCols = make([]int, len(rightEqCols))

	if len(leftEqCols) != len(rightEqCols) {
		__antithesis_instrumentation__.Notify(551623)
		panic("creating zigzag join with unequal number of equated cols")
	} else {
		__antithesis_instrumentation__.Notify(551624)
	}
	__antithesis_instrumentation__.Notify(551613)

	for i, c := range leftEqCols {
		__antithesis_instrumentation__.Notify(551625)
		n.sides[0].eqCols[i] = int(c)
		n.sides[1].eqCols[i] = int(rightEqCols[i])
	}
	__antithesis_instrumentation__.Notify(551614)

	n.columns = make(
		colinfo.ResultColumns,
		0,
		len(leftScan.resultColumns)+len(rightScan.resultColumns),
	)
	n.columns = append(n.columns, leftScan.resultColumns...)
	n.columns = append(n.columns, rightScan.resultColumns...)

	mkFixedVals := func(fixedVals []tree.TypedExpr, index cat.Index) *valuesNode {
		__antithesis_instrumentation__.Notify(551626)
		cols := make(colinfo.ResultColumns, len(fixedVals))
		for i := range cols {
			__antithesis_instrumentation__.Notify(551628)
			col := index.Column(i)
			cols[i].Name = string(col.ColName())
			cols[i].Typ = col.DatumType()
		}
		__antithesis_instrumentation__.Notify(551627)
		return &valuesNode{
			columns:          cols,
			tuples:           [][]tree.TypedExpr{fixedVals},
			specifiedInQuery: true,
		}
	}
	__antithesis_instrumentation__.Notify(551615)
	n.sides[0].fixedVals = mkFixedVals(leftFixedVals, leftIndex)
	n.sides[1].fixedVals = mkFixedVals(rightFixedVals, rightIndex)
	return n, nil
}

func (ef *execFactory) ConstructLimit(
	input exec.Node, limit, offset tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551629)
	plan := input.(planNode)

	if l, ok := plan.(*limitNode); ok && func() bool {
		__antithesis_instrumentation__.Notify(551632)
		return l.countExpr == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(551633)
		return offset == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(551634)
		l.countExpr = limit
		return l, nil
	} else {
		__antithesis_instrumentation__.Notify(551635)
	}
	__antithesis_instrumentation__.Notify(551630)

	if spool, ok := plan.(*spoolNode); ok {
		__antithesis_instrumentation__.Notify(551636)
		if val, ok := limit.(*tree.DInt); ok {
			__antithesis_instrumentation__.Notify(551637)
			spool.hardLimit = int64(*val)
		} else {
			__antithesis_instrumentation__.Notify(551638)
		}
	} else {
		__antithesis_instrumentation__.Notify(551639)
	}
	__antithesis_instrumentation__.Notify(551631)
	return &limitNode{
		plan:       plan,
		countExpr:  limit,
		offsetExpr: offset,
	}, nil
}

func (ef *execFactory) ConstructTopK(
	input exec.Node, k int64, ordering exec.OutputOrdering, alreadyOrderedPrefix int,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551640)
	return &topKNode{
		plan:                 input.(planNode),
		k:                    k,
		ordering:             colinfo.ColumnOrdering(ordering),
		alreadyOrderedPrefix: alreadyOrderedPrefix,
	}, nil
}

func (ef *execFactory) ConstructMax1Row(input exec.Node, errorText string) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551641)
	plan := input.(planNode)
	return &max1RowNode{
		plan:      plan,
		errorText: errorText,
	}, nil
}

func (ef *execFactory) ConstructBuffer(input exec.Node, label string) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551642)
	return &bufferNode{
		plan:  input.(planNode),
		label: label,
	}, nil
}

func (ef *execFactory) ConstructScanBuffer(ref exec.Node, label string) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551643)
	if n, ok := ref.(*explain.Node); ok {
		__antithesis_instrumentation__.Notify(551645)

		ref = n.WrappedNode()
	} else {
		__antithesis_instrumentation__.Notify(551646)
	}
	__antithesis_instrumentation__.Notify(551644)
	return &scanBufferNode{
		buffer: ref.(*bufferNode),
		label:  label,
	}, nil
}

func (ef *execFactory) ConstructRecursiveCTE(
	initial exec.Node, fn exec.RecursiveCTEIterationFn, label string, deduplicate bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551647)
	return &recursiveCTENode{
		initial:        initial.(planNode),
		genIterationFn: fn,
		label:          label,
		deduplicate:    deduplicate,
	}, nil
}

func (ef *execFactory) ConstructProjectSet(
	n exec.Node, exprs tree.TypedExprs, zipCols colinfo.ResultColumns, numColsPerGen []int,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551648)
	src := asDataSource(n)
	cols := append(src.columns, zipCols...)
	return &projectSetNode{
		source: src.plan,
		projectSetPlanningInfo: projectSetPlanningInfo{
			columns:         cols,
			numColsInSource: len(src.columns),
			exprs:           exprs,
			numColsPerGen:   numColsPerGen,
		},
	}, nil
}

func (ef *execFactory) ConstructWindow(root exec.Node, wi exec.WindowInfo) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551649)
	p := &windowNode{
		plan:         root.(planNode),
		columns:      wi.Cols,
		windowRender: make([]tree.TypedExpr, len(wi.Cols)),
	}

	partitionIdxs := make([]int, len(wi.Partition))
	for i, idx := range wi.Partition {
		__antithesis_instrumentation__.Notify(551652)
		partitionIdxs[i] = int(idx)
	}
	__antithesis_instrumentation__.Notify(551650)

	p.funcs = make([]*windowFuncHolder, len(wi.Exprs))
	for i := range wi.Exprs {
		__antithesis_instrumentation__.Notify(551653)
		argsIdxs := make([]uint32, len(wi.ArgIdxs[i]))
		for j := range argsIdxs {
			__antithesis_instrumentation__.Notify(551656)
			argsIdxs[j] = uint32(wi.ArgIdxs[i][j])
		}
		__antithesis_instrumentation__.Notify(551654)

		p.funcs[i] = &windowFuncHolder{
			expr:           wi.Exprs[i],
			args:           wi.Exprs[i].Exprs,
			argsIdxs:       argsIdxs,
			window:         p,
			filterColIdx:   wi.FilterIdxs[i],
			outputColIdx:   wi.OutputIdxs[i],
			partitionIdxs:  partitionIdxs,
			columnOrdering: wi.Ordering,
			frame:          wi.Exprs[i].WindowDef.Frame,
		}
		if len(wi.Ordering) == 0 {
			__antithesis_instrumentation__.Notify(551657)
			frame := p.funcs[i].frame
			if frame.Mode == treewindow.RANGE && func() bool {
				__antithesis_instrumentation__.Notify(551658)
				return frame.Bounds.HasOffset() == true
			}() == true {
				__antithesis_instrumentation__.Notify(551659)

				return nil, errors.AssertionFailedf("a RANGE mode frame with an offset bound must have an ORDER BY column")
			} else {
				__antithesis_instrumentation__.Notify(551660)
			}
		} else {
			__antithesis_instrumentation__.Notify(551661)
		}
		__antithesis_instrumentation__.Notify(551655)

		p.windowRender[wi.OutputIdxs[i]] = p.funcs[i]
	}
	__antithesis_instrumentation__.Notify(551651)

	return p, nil
}

func (ef *execFactory) ConstructPlan(
	root exec.Node,
	subqueries []exec.Subquery,
	cascades []exec.Cascade,
	checks []exec.Node,
	rootRowCount int64,
) (exec.Plan, error) {
	__antithesis_instrumentation__.Notify(551662)

	if spool, ok := root.(*spoolNode); ok {
		__antithesis_instrumentation__.Notify(551664)
		root = spool.source
	} else {
		__antithesis_instrumentation__.Notify(551665)
	}
	__antithesis_instrumentation__.Notify(551663)
	return constructPlan(ef.planner, root, subqueries, cascades, checks, rootRowCount)
}

type urlOutputter struct {
	buf bytes.Buffer
}

func (e *urlOutputter) writef(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(551666)
	if e.buf.Len() > 0 {
		__antithesis_instrumentation__.Notify(551668)
		e.buf.WriteString("\n")
	} else {
		__antithesis_instrumentation__.Notify(551669)
	}
	__antithesis_instrumentation__.Notify(551667)
	fmt.Fprintf(&e.buf, format, args...)
}

func (e *urlOutputter) finish() (url.URL, error) {
	__antithesis_instrumentation__.Notify(551670)

	var compressed bytes.Buffer
	encoder := base64.NewEncoder(base64.URLEncoding, &compressed)
	compressor := zlib.NewWriter(encoder)
	if _, err := e.buf.WriteTo(compressor); err != nil {
		__antithesis_instrumentation__.Notify(551674)
		return url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(551675)
	}
	__antithesis_instrumentation__.Notify(551671)
	if err := compressor.Close(); err != nil {
		__antithesis_instrumentation__.Notify(551676)
		return url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(551677)
	}
	__antithesis_instrumentation__.Notify(551672)
	if err := encoder.Close(); err != nil {
		__antithesis_instrumentation__.Notify(551678)
		return url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(551679)
	}
	__antithesis_instrumentation__.Notify(551673)
	return url.URL{
		Scheme:   "https",
		Host:     "cockroachdb.github.io",
		Path:     "text/decode.html",
		Fragment: compressed.String(),
	}, nil
}

func (ef *execFactory) showEnv(plan string, envOpts exec.ExplainEnvData) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551680)
	var out urlOutputter

	ie := ef.planner.extendedEvalCtx.ExecCfg.InternalExecutorFactory(
		ef.planner.EvalContext().Context,
		ef.planner.SessionData(),
	)
	c := makeStmtEnvCollector(ef.planner.EvalContext().Context, ie.(*InternalExecutor))

	if err := c.PrintVersion(&out.buf); err != nil {
		__antithesis_instrumentation__.Notify(551687)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551688)
	}
	__antithesis_instrumentation__.Notify(551681)
	out.writef("")

	if err := c.PrintSessionSettings(&out.buf); err != nil {
		__antithesis_instrumentation__.Notify(551689)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551690)
	}
	__antithesis_instrumentation__.Notify(551682)

	for i := range envOpts.Sequences {
		__antithesis_instrumentation__.Notify(551691)
		out.writef("")
		if err := c.PrintCreateSequence(&out.buf, &envOpts.Sequences[i]); err != nil {
			__antithesis_instrumentation__.Notify(551692)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551693)
		}
	}
	__antithesis_instrumentation__.Notify(551683)

	for i := range envOpts.Tables {
		__antithesis_instrumentation__.Notify(551694)
		out.writef("")
		if err := c.PrintCreateTable(&out.buf, &envOpts.Tables[i]); err != nil {
			__antithesis_instrumentation__.Notify(551696)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551697)
		}
		__antithesis_instrumentation__.Notify(551695)
		out.writef("")

		err := c.PrintTableStats(&out.buf, &envOpts.Tables[i], true)
		if err != nil {
			__antithesis_instrumentation__.Notify(551698)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551699)
		}
	}
	__antithesis_instrumentation__.Notify(551684)

	for i := range envOpts.Views {
		__antithesis_instrumentation__.Notify(551700)
		out.writef("")
		if err := c.PrintCreateView(&out.buf, &envOpts.Views[i]); err != nil {
			__antithesis_instrumentation__.Notify(551701)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551702)
		}
	}
	__antithesis_instrumentation__.Notify(551685)

	out.writef("%s;\n----\n%s", ef.planner.stmt.AST.String(), plan)

	url, err := out.finish()
	if err != nil {
		__antithesis_instrumentation__.Notify(551703)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551704)
	}
	__antithesis_instrumentation__.Notify(551686)
	return &valuesNode{
		columns:          append(colinfo.ResultColumns(nil), colinfo.ExplainPlanColumns...),
		tuples:           [][]tree.TypedExpr{{tree.NewDString(url.String())}},
		specifiedInQuery: true,
	}, nil
}

func (ef *execFactory) ConstructExplainOpt(
	planText string, envOpts exec.ExplainEnvData,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551705)

	if envOpts.ShowEnv {
		__antithesis_instrumentation__.Notify(551708)
		return ef.showEnv(planText, envOpts)
	} else {
		__antithesis_instrumentation__.Notify(551709)
	}
	__antithesis_instrumentation__.Notify(551706)

	var rows [][]tree.TypedExpr
	ss := strings.Split(strings.Trim(planText, "\n"), "\n")
	for _, line := range ss {
		__antithesis_instrumentation__.Notify(551710)
		rows = append(rows, []tree.TypedExpr{tree.NewDString(line)})
	}
	__antithesis_instrumentation__.Notify(551707)

	return &valuesNode{
		columns:          append(colinfo.ResultColumns(nil), colinfo.ExplainPlanColumns...),
		tuples:           rows,
		specifiedInQuery: true,
	}, nil
}

func (ef *execFactory) ConstructShowTrace(typ tree.ShowTraceType, compact bool) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551711)
	var node planNode = ef.planner.makeShowTraceNode(compact, typ == tree.ShowTraceKV)

	ageColIdx := colinfo.GetTraceAgeColumnIdx(compact)
	node = &sortNode{
		plan: node,
		ordering: colinfo.ColumnOrdering{
			colinfo.ColumnOrderInfo{ColIdx: ageColIdx, Direction: encoding.Ascending},
		},
	}

	if typ == tree.ShowTraceReplica {
		__antithesis_instrumentation__.Notify(551713)
		node = &showTraceReplicaNode{plan: node}
	} else {
		__antithesis_instrumentation__.Notify(551714)
	}
	__antithesis_instrumentation__.Notify(551712)
	return node, nil
}

func (ef *execFactory) ConstructInsert(
	input exec.Node,
	table cat.Table,
	arbiterIndexes cat.IndexOrdinals,
	arbiterConstraints cat.UniqueOrdinals,
	insertColOrdSet exec.TableColumnOrdinalSet,
	returnColOrdSet exec.TableColumnOrdinalSet,
	checkOrdSet exec.CheckOrdinalSet,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551715)
	ctx := ef.planner.extendedEvalCtx.Context

	rowsNeeded := !returnColOrdSet.Empty()
	tabDesc := table.(*optTable).desc
	cols := makeColList(table, insertColOrdSet)

	if err := ef.planner.maybeSetSystemConfig(tabDesc.GetID()); err != nil {
		__antithesis_instrumentation__.Notify(551721)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551722)
	}
	__antithesis_instrumentation__.Notify(551716)

	internal := ef.planner.SessionData().Internal
	ri, err := row.MakeInserter(
		ctx,
		ef.planner.txn,
		ef.planner.ExecCfg().Codec,
		tabDesc,
		cols,
		ef.planner.alloc,
		&ef.planner.ExecCfg().Settings.SV,
		internal,
		ef.planner.ExecCfg().GetRowMetrics(internal),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(551723)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551724)
	}
	__antithesis_instrumentation__.Notify(551717)

	ins := insertNodePool.Get().(*insertNode)
	*ins = insertNode{
		source: input.(planNode),
		run: insertRun{
			ti:         tableInserter{ri: ri},
			checkOrds:  checkOrdSet,
			insertCols: ri.InsertCols,
		},
	}

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551725)
		returnCols := makeColList(table, returnColOrdSet)
		ins.columns = colinfo.ResultColumnsFromColumns(tabDesc.GetID(), returnCols)

		ins.run.tabColIdxToRetIdx = makePublicToReturnColumnIndexMapping(tabDesc, returnCols)
		ins.run.rowsNeeded = true
	} else {
		__antithesis_instrumentation__.Notify(551726)
	}
	__antithesis_instrumentation__.Notify(551718)

	if autoCommit {
		__antithesis_instrumentation__.Notify(551727)
		ins.enableAutoCommit()
	} else {
		__antithesis_instrumentation__.Notify(551728)
	}
	__antithesis_instrumentation__.Notify(551719)

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551729)
		return &spoolNode{source: &serializeNode{source: ins}}, nil
	} else {
		__antithesis_instrumentation__.Notify(551730)
	}
	__antithesis_instrumentation__.Notify(551720)

	return &rowCountNode{source: ins}, nil
}

func (ef *execFactory) ConstructInsertFastPath(
	rows [][]tree.TypedExpr,
	table cat.Table,
	insertColOrdSet exec.TableColumnOrdinalSet,
	returnColOrdSet exec.TableColumnOrdinalSet,
	checkOrdSet exec.CheckOrdinalSet,
	fkChecks []exec.InsertFastPathFKCheck,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551731)
	ctx := ef.planner.extendedEvalCtx.Context

	rowsNeeded := !returnColOrdSet.Empty()
	tabDesc := table.(*optTable).desc
	cols := makeColList(table, insertColOrdSet)

	if err := ef.planner.maybeSetSystemConfig(tabDesc.GetID()); err != nil {
		__antithesis_instrumentation__.Notify(551739)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551740)
	}
	__antithesis_instrumentation__.Notify(551732)

	internal := ef.planner.SessionData().Internal
	ri, err := row.MakeInserter(
		ctx,
		ef.planner.txn,
		ef.planner.ExecCfg().Codec,
		tabDesc,
		cols,
		ef.planner.alloc,
		&ef.planner.ExecCfg().Settings.SV,
		internal,
		ef.planner.ExecCfg().GetRowMetrics(internal),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(551741)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551742)
	}
	__antithesis_instrumentation__.Notify(551733)

	ins := insertFastPathNodePool.Get().(*insertFastPathNode)
	*ins = insertFastPathNode{
		input: rows,
		run: insertFastPathRun{
			insertRun: insertRun{
				ti:         tableInserter{ri: ri},
				checkOrds:  checkOrdSet,
				insertCols: ri.InsertCols,
			},
		},
	}

	if len(fkChecks) > 0 {
		__antithesis_instrumentation__.Notify(551743)
		ins.run.fkChecks = make([]insertFastPathFKCheck, len(fkChecks))
		for i := range fkChecks {
			__antithesis_instrumentation__.Notify(551744)
			ins.run.fkChecks[i].InsertFastPathFKCheck = fkChecks[i]
		}
	} else {
		__antithesis_instrumentation__.Notify(551745)
	}
	__antithesis_instrumentation__.Notify(551734)

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551746)
		returnCols := makeColList(table, returnColOrdSet)
		ins.columns = colinfo.ResultColumnsFromColumns(tabDesc.GetID(), returnCols)

		ins.run.tabColIdxToRetIdx = makePublicToReturnColumnIndexMapping(tabDesc, returnCols)
		ins.run.rowsNeeded = true
	} else {
		__antithesis_instrumentation__.Notify(551747)
	}
	__antithesis_instrumentation__.Notify(551735)

	if len(rows) == 0 {
		__antithesis_instrumentation__.Notify(551748)
		return &zeroNode{columns: ins.columns}, nil
	} else {
		__antithesis_instrumentation__.Notify(551749)
	}
	__antithesis_instrumentation__.Notify(551736)

	if autoCommit {
		__antithesis_instrumentation__.Notify(551750)
		ins.enableAutoCommit()
	} else {
		__antithesis_instrumentation__.Notify(551751)
	}
	__antithesis_instrumentation__.Notify(551737)

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551752)
		return &spoolNode{source: &serializeNode{source: ins}}, nil
	} else {
		__antithesis_instrumentation__.Notify(551753)
	}
	__antithesis_instrumentation__.Notify(551738)

	return &rowCountNode{source: ins}, nil
}

func (ef *execFactory) ConstructUpdate(
	input exec.Node,
	table cat.Table,
	fetchColOrdSet exec.TableColumnOrdinalSet,
	updateColOrdSet exec.TableColumnOrdinalSet,
	returnColOrdSet exec.TableColumnOrdinalSet,
	checks exec.CheckOrdinalSet,
	passthrough colinfo.ResultColumns,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551754)
	ctx := ef.planner.extendedEvalCtx.Context

	if !updateColOrdSet.SubsetOf(fetchColOrdSet) {
		__antithesis_instrumentation__.Notify(551763)
		return nil, errors.AssertionFailedf("execution requires all update columns have a fetch column")
	} else {
		__antithesis_instrumentation__.Notify(551764)
	}
	__antithesis_instrumentation__.Notify(551755)

	rowsNeeded := !returnColOrdSet.Empty()
	tabDesc := table.(*optTable).desc
	fetchCols := makeColList(table, fetchColOrdSet)

	if err := ef.planner.maybeSetSystemConfig(tabDesc.GetID()); err != nil {
		__antithesis_instrumentation__.Notify(551765)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551766)
	}
	__antithesis_instrumentation__.Notify(551756)

	updateCols := makeColList(table, updateColOrdSet)
	sourceSlots := make([]sourceSlot, len(updateCols))
	for i := range sourceSlots {
		__antithesis_instrumentation__.Notify(551767)
		sourceSlots[i] = scalarSlot{column: updateCols[i], sourceIndex: len(fetchCols) + i}
	}
	__antithesis_instrumentation__.Notify(551757)

	internal := ef.planner.SessionData().Internal
	ru, err := row.MakeUpdater(
		ctx,
		ef.planner.txn,
		ef.planner.ExecCfg().Codec,
		tabDesc,
		updateCols,
		fetchCols,
		row.UpdaterDefault,
		ef.planner.alloc,
		&ef.planner.ExecCfg().Settings.SV,
		internal,
		ef.planner.ExecCfg().GetRowMetrics(internal),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(551768)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551769)
	}
	__antithesis_instrumentation__.Notify(551758)

	var updateColsIdx catalog.TableColMap
	for i := range ru.UpdateCols {
		__antithesis_instrumentation__.Notify(551770)
		id := ru.UpdateCols[i].GetID()
		updateColsIdx.Set(id, i)
	}
	__antithesis_instrumentation__.Notify(551759)

	upd := updateNodePool.Get().(*updateNode)
	*upd = updateNode{
		source: input.(planNode),
		run: updateRun{
			tu:        tableUpdater{ru: ru},
			checkOrds: checks,
			iVarContainerForComputedCols: schemaexpr.RowIndexedVarContainer{
				CurSourceRow: make(tree.Datums, len(ru.FetchCols)),
				Cols:         ru.FetchCols,
				Mapping:      ru.FetchColIDtoRowIndex,
			},
			sourceSlots:    sourceSlots,
			updateValues:   make(tree.Datums, len(ru.UpdateCols)),
			updateColsIdx:  updateColsIdx,
			numPassthrough: len(passthrough),
		},
	}

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551771)
		returnCols := makeColList(table, returnColOrdSet)

		upd.columns = colinfo.ResultColumnsFromColumns(tabDesc.GetID(), returnCols)

		upd.columns = append(upd.columns, passthrough...)

		upd.run.rowIdxToRetIdx = row.ColMapping(ru.FetchCols, returnCols)
		upd.run.rowsNeeded = true
	} else {
		__antithesis_instrumentation__.Notify(551772)
	}
	__antithesis_instrumentation__.Notify(551760)

	if autoCommit {
		__antithesis_instrumentation__.Notify(551773)
		upd.enableAutoCommit()
	} else {
		__antithesis_instrumentation__.Notify(551774)
	}
	__antithesis_instrumentation__.Notify(551761)

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551775)
		return &spoolNode{source: &serializeNode{source: upd}}, nil
	} else {
		__antithesis_instrumentation__.Notify(551776)
	}
	__antithesis_instrumentation__.Notify(551762)

	return &rowCountNode{source: upd}, nil
}

func (ef *execFactory) ConstructUpsert(
	input exec.Node,
	table cat.Table,
	arbiterIndexes cat.IndexOrdinals,
	arbiterConstraints cat.UniqueOrdinals,
	canaryCol exec.NodeColumnOrdinal,
	insertColOrdSet exec.TableColumnOrdinalSet,
	fetchColOrdSet exec.TableColumnOrdinalSet,
	updateColOrdSet exec.TableColumnOrdinalSet,
	returnColOrdSet exec.TableColumnOrdinalSet,
	checks exec.CheckOrdinalSet,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551777)
	ctx := ef.planner.extendedEvalCtx.Context

	rowsNeeded := !returnColOrdSet.Empty()
	tabDesc := table.(*optTable).desc
	insertCols := makeColList(table, insertColOrdSet)
	fetchCols := makeColList(table, fetchColOrdSet)
	updateCols := makeColList(table, updateColOrdSet)

	if err := ef.planner.maybeSetSystemConfig(tabDesc.GetID()); err != nil {
		__antithesis_instrumentation__.Notify(551784)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551785)
	}
	__antithesis_instrumentation__.Notify(551778)

	internal := ef.planner.SessionData().Internal
	ri, err := row.MakeInserter(
		ctx,
		ef.planner.txn,
		ef.planner.ExecCfg().Codec,
		tabDesc,
		insertCols,
		ef.planner.alloc,
		&ef.planner.ExecCfg().Settings.SV,
		internal,
		ef.planner.ExecCfg().GetRowMetrics(internal),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(551786)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551787)
	}
	__antithesis_instrumentation__.Notify(551779)

	ru, err := row.MakeUpdater(
		ctx,
		ef.planner.txn,
		ef.planner.ExecCfg().Codec,
		tabDesc,
		updateCols,
		fetchCols,
		row.UpdaterDefault,
		ef.planner.alloc,
		&ef.planner.ExecCfg().Settings.SV,
		internal,
		ef.planner.ExecCfg().GetRowMetrics(internal),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(551788)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551789)
	}
	__antithesis_instrumentation__.Notify(551780)

	ups := upsertNodePool.Get().(*upsertNode)
	*ups = upsertNode{
		source: input.(planNode),
		run: upsertRun{
			checkOrds:  checks,
			insertCols: ri.InsertCols,
			tw: optTableUpserter{
				ri:            ri,
				canaryOrdinal: int(canaryCol),
				fetchCols:     fetchCols,
				updateCols:    updateCols,
				ru:            ru,
			},
		},
	}

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551790)
		returnCols := makeColList(table, returnColOrdSet)
		ups.columns = colinfo.ResultColumnsFromColumns(tabDesc.GetID(), returnCols)

		ups.run.tw.tabColIdxToRetIdx = makePublicToReturnColumnIndexMapping(tabDesc, returnCols)
		ups.run.tw.returnCols = returnCols
		ups.run.tw.rowsNeeded = true
	} else {
		__antithesis_instrumentation__.Notify(551791)
	}
	__antithesis_instrumentation__.Notify(551781)

	if autoCommit {
		__antithesis_instrumentation__.Notify(551792)
		ups.enableAutoCommit()
	} else {
		__antithesis_instrumentation__.Notify(551793)
	}
	__antithesis_instrumentation__.Notify(551782)

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551794)
		return &spoolNode{source: &serializeNode{source: ups}}, nil
	} else {
		__antithesis_instrumentation__.Notify(551795)
	}
	__antithesis_instrumentation__.Notify(551783)

	return &rowCountNode{source: ups}, nil
}

func (ef *execFactory) ConstructDelete(
	input exec.Node,
	table cat.Table,
	fetchColOrdSet exec.TableColumnOrdinalSet,
	returnColOrdSet exec.TableColumnOrdinalSet,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551796)

	rowsNeeded := !returnColOrdSet.Empty()
	tabDesc := table.(*optTable).desc
	fetchCols := makeColList(table, fetchColOrdSet)

	if err := ef.planner.maybeSetSystemConfig(tabDesc.GetID()); err != nil {
		__antithesis_instrumentation__.Notify(551801)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551802)
	}
	__antithesis_instrumentation__.Notify(551797)

	internal := ef.planner.SessionData().Internal
	rd := row.MakeDeleter(
		ef.planner.ExecCfg().Codec,
		tabDesc,
		fetchCols,
		&ef.planner.ExecCfg().Settings.SV,
		internal,
		ef.planner.ExecCfg().GetRowMetrics(internal),
	)

	del := deleteNodePool.Get().(*deleteNode)
	*del = deleteNode{
		source: input.(planNode),
		run: deleteRun{
			td:                        tableDeleter{rd: rd, alloc: ef.planner.alloc},
			partialIndexDelValsOffset: len(rd.FetchCols),
		},
	}

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551803)
		returnCols := makeColList(table, returnColOrdSet)

		del.columns = colinfo.ResultColumnsFromColumns(tabDesc.GetID(), returnCols)

		del.run.rowIdxToRetIdx = row.ColMapping(rd.FetchCols, returnCols)
		del.run.rowsNeeded = true
	} else {
		__antithesis_instrumentation__.Notify(551804)
	}
	__antithesis_instrumentation__.Notify(551798)

	if autoCommit {
		__antithesis_instrumentation__.Notify(551805)
		del.enableAutoCommit()
	} else {
		__antithesis_instrumentation__.Notify(551806)
	}
	__antithesis_instrumentation__.Notify(551799)

	if rowsNeeded {
		__antithesis_instrumentation__.Notify(551807)
		return &spoolNode{source: &serializeNode{source: del}}, nil
	} else {
		__antithesis_instrumentation__.Notify(551808)
	}
	__antithesis_instrumentation__.Notify(551800)

	return &rowCountNode{source: del}, nil
}

func (ef *execFactory) ConstructDeleteRange(
	table cat.Table,
	needed exec.TableColumnOrdinalSet,
	indexConstraint *constraint.Constraint,
	autoCommit bool,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551809)
	tabDesc := table.(*optTable).desc
	var sb span.Builder
	sb.Init(ef.planner.EvalContext(), ef.planner.ExecCfg().Codec, tabDesc, tabDesc.GetPrimaryIndex())

	if err := ef.planner.maybeSetSystemConfig(tabDesc.GetID()); err != nil {
		__antithesis_instrumentation__.Notify(551812)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551813)
	}
	__antithesis_instrumentation__.Notify(551810)

	spans, err := sb.SpansFromConstraint(indexConstraint, span.NoopSplitter())
	if err != nil {
		__antithesis_instrumentation__.Notify(551814)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551815)
	}
	__antithesis_instrumentation__.Notify(551811)

	dr := &deleteRangeNode{
		spans:             spans,
		desc:              tabDesc,
		autoCommitEnabled: autoCommit,
	}

	return dr, nil
}

func (ef *execFactory) ConstructCreateTable(
	schema cat.Schema, ct *tree.CreateTable,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551816)
	if err := checkSchemaChangeEnabled(
		ef.planner.EvalContext().Context,
		ef.planner.ExecCfg(),
		"CREATE TABLE",
	); err != nil {
		__antithesis_instrumentation__.Notify(551818)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551819)
	}
	__antithesis_instrumentation__.Notify(551817)
	return &createTableNode{
		n:      ct,
		dbDesc: schema.(*optSchema).database,
	}, nil
}

func (ef *execFactory) ConstructCreateTableAs(
	input exec.Node, schema cat.Schema, ct *tree.CreateTable,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551820)
	if err := checkSchemaChangeEnabled(
		ef.planner.EvalContext().Context,
		ef.planner.ExecCfg(),
		"CREATE TABLE",
	); err != nil {
		__antithesis_instrumentation__.Notify(551822)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551823)
	}
	__antithesis_instrumentation__.Notify(551821)

	return &createTableNode{
		n:          ct,
		dbDesc:     schema.(*optSchema).database,
		sourcePlan: input.(planNode),
	}, nil
}

func (ef *execFactory) ConstructCreateView(
	schema cat.Schema,
	viewName *cat.DataSourceName,
	ifNotExists bool,
	replace bool,
	persistence tree.Persistence,
	materialized bool,
	viewQuery string,
	columns colinfo.ResultColumns,
	deps opt.ViewDeps,
	typeDeps opt.ViewTypeDeps,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551824)

	if err := checkSchemaChangeEnabled(
		ef.planner.EvalContext().Context,
		ef.planner.ExecCfg(),
		"CREATE VIEW",
	); err != nil {
		__antithesis_instrumentation__.Notify(551828)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551829)
	}
	__antithesis_instrumentation__.Notify(551825)

	planDeps := make(planDependencies, len(deps))
	for _, d := range deps {
		__antithesis_instrumentation__.Notify(551830)
		desc, err := getDescForDataSource(d.DataSource)
		if err != nil {
			__antithesis_instrumentation__.Notify(551834)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(551835)
		}
		__antithesis_instrumentation__.Notify(551831)
		var ref descpb.TableDescriptor_Reference
		if d.SpecificIndex {
			__antithesis_instrumentation__.Notify(551836)
			idx := d.DataSource.(cat.Table).Index(d.Index)
			ref.IndexID = idx.(*optIndex).idx.GetID()
		} else {
			__antithesis_instrumentation__.Notify(551837)
		}
		__antithesis_instrumentation__.Notify(551832)
		if !d.ColumnOrdinals.Empty() {
			__antithesis_instrumentation__.Notify(551838)
			ref.ColumnIDs = make([]descpb.ColumnID, 0, d.ColumnOrdinals.Len())
			d.ColumnOrdinals.ForEach(func(ord int) {
				__antithesis_instrumentation__.Notify(551839)
				ref.ColumnIDs = append(ref.ColumnIDs, desc.AllColumns()[ord].GetID())
			})
		} else {
			__antithesis_instrumentation__.Notify(551840)
		}
		__antithesis_instrumentation__.Notify(551833)
		entry := planDeps[desc.GetID()]
		entry.desc = desc
		entry.deps = append(entry.deps, ref)
		planDeps[desc.GetID()] = entry
	}
	__antithesis_instrumentation__.Notify(551826)

	typeDepSet := make(typeDependencies, typeDeps.Len())
	typeDeps.ForEach(func(id int) {
		__antithesis_instrumentation__.Notify(551841)
		typeDepSet[descpb.ID(id)] = struct{}{}
	})
	__antithesis_instrumentation__.Notify(551827)

	return &createViewNode{
		viewName:     viewName,
		ifNotExists:  ifNotExists,
		replace:      replace,
		materialized: materialized,
		persistence:  persistence,
		viewQuery:    viewQuery,
		dbDesc:       schema.(*optSchema).database,
		columns:      columns,
		planDeps:     planDeps,
		typeDeps:     typeDepSet,
	}, nil
}

func (ef *execFactory) ConstructSequenceSelect(sequence cat.Sequence) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551842)
	return ef.planner.SequenceSelectNode(sequence.(*optSequence).desc)
}

func (ef *execFactory) ConstructSaveTable(
	input exec.Node, table *cat.DataSourceName, colNames []string,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551843)
	return ef.planner.makeSaveTable(input.(planNode), table, colNames), nil
}

func (ef *execFactory) ConstructErrorIfRows(
	input exec.Node, mkErr exec.MkErrFn,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551844)
	return &errorIfRowsNode{
		plan:  input.(planNode),
		mkErr: mkErr,
	}, nil
}

func (ef *execFactory) ConstructOpaque(metadata opt.OpaqueMetadata) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551845)
	return constructOpaque(metadata)
}

func (ef *execFactory) ConstructAlterTableSplit(
	index cat.Index, input exec.Node, expiration tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551846)
	if err := checkSchemaChangeEnabled(
		ef.planner.EvalContext().Context,
		ef.planner.ExecCfg(),
		"ALTER TABLE/INDEX SPLIT AT",
	); err != nil {
		__antithesis_instrumentation__.Notify(551850)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551851)
	}
	__antithesis_instrumentation__.Notify(551847)

	if !ef.planner.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(551852)
		return nil, errorutil.UnsupportedWithMultiTenancy(54254)
	} else {
		__antithesis_instrumentation__.Notify(551853)
	}
	__antithesis_instrumentation__.Notify(551848)

	expirationTime, err := parseExpirationTime(ef.planner.EvalContext(), expiration)
	if err != nil {
		__antithesis_instrumentation__.Notify(551854)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551855)
	}
	__antithesis_instrumentation__.Notify(551849)

	return &splitNode{
		tableDesc:      index.Table().(*optTable).desc,
		index:          index.(*optIndex).idx,
		rows:           input.(planNode),
		expirationTime: expirationTime,
	}, nil
}

func (ef *execFactory) ConstructAlterTableUnsplit(
	index cat.Index, input exec.Node,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551856)
	if err := checkSchemaChangeEnabled(
		ef.planner.EvalContext().Context,
		ef.planner.ExecCfg(),
		"ALTER TABLE/INDEX UNSPLIT AT",
	); err != nil {
		__antithesis_instrumentation__.Notify(551859)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551860)
	}
	__antithesis_instrumentation__.Notify(551857)

	if !ef.planner.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(551861)
		return nil, errorutil.UnsupportedWithMultiTenancy(54254)
	} else {
		__antithesis_instrumentation__.Notify(551862)
	}
	__antithesis_instrumentation__.Notify(551858)

	return &unsplitNode{
		tableDesc: index.Table().(*optTable).desc,
		index:     index.(*optIndex).idx,
		rows:      input.(planNode),
	}, nil
}

func (ef *execFactory) ConstructAlterTableUnsplitAll(index cat.Index) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551863)
	if err := checkSchemaChangeEnabled(
		ef.planner.EvalContext().Context,
		ef.planner.ExecCfg(),
		"ALTER TABLE/INDEX UNSPLIT ALL",
	); err != nil {
		__antithesis_instrumentation__.Notify(551866)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551867)
	}
	__antithesis_instrumentation__.Notify(551864)

	if !ef.planner.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(551868)
		return nil, errorutil.UnsupportedWithMultiTenancy(54254)
	} else {
		__antithesis_instrumentation__.Notify(551869)
	}
	__antithesis_instrumentation__.Notify(551865)

	return &unsplitAllNode{
		tableDesc: index.Table().(*optTable).desc,
		index:     index.(*optIndex).idx,
	}, nil
}

func (ef *execFactory) ConstructAlterTableRelocate(
	index cat.Index, input exec.Node, relocateSubject tree.RelocateSubject,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551870)
	if !ef.planner.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(551872)
		return nil, errorutil.UnsupportedWithMultiTenancy(54250)
	} else {
		__antithesis_instrumentation__.Notify(551873)
	}
	__antithesis_instrumentation__.Notify(551871)

	return &relocateNode{
		subjectReplicas: relocateSubject,
		tableDesc:       index.Table().(*optTable).desc,
		index:           index.(*optIndex).idx,
		rows:            input.(planNode),
	}, nil
}

func (ef *execFactory) ConstructAlterRangeRelocate(
	input exec.Node,
	relocateSubject tree.RelocateSubject,
	toStoreID tree.TypedExpr,
	fromStoreID tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551874)
	if !ef.planner.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(551876)
		return nil, errorutil.UnsupportedWithMultiTenancy(54250)
	} else {
		__antithesis_instrumentation__.Notify(551877)
	}
	__antithesis_instrumentation__.Notify(551875)

	return &relocateRange{
		rows:            input.(planNode),
		subjectReplicas: relocateSubject,
		toStoreID:       toStoreID,
		fromStoreID:     fromStoreID,
	}, nil
}

func (ef *execFactory) ConstructControlJobs(
	command tree.JobCommand, input exec.Node, reason tree.TypedExpr,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551878)
	reasonDatum, err := reason.Eval(ef.planner.EvalContext())
	if err != nil {
		__antithesis_instrumentation__.Notify(551881)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551882)
	}
	__antithesis_instrumentation__.Notify(551879)

	var reasonStr string
	if reasonDatum != tree.DNull {
		__antithesis_instrumentation__.Notify(551883)
		reasonStrDatum, ok := reasonDatum.(*tree.DString)
		if !ok {
			__antithesis_instrumentation__.Notify(551885)
			return nil, errors.Errorf("expected string value for the reason")
		} else {
			__antithesis_instrumentation__.Notify(551886)
		}
		__antithesis_instrumentation__.Notify(551884)
		reasonStr = string(*reasonStrDatum)
	} else {
		__antithesis_instrumentation__.Notify(551887)
	}
	__antithesis_instrumentation__.Notify(551880)

	return &controlJobsNode{
		rows:          input.(planNode),
		desiredStatus: jobCommandToDesiredStatus[command],
		reason:        reasonStr,
	}, nil
}

func (ef *execFactory) ConstructControlSchedules(
	command tree.ScheduleCommand, input exec.Node,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551888)
	return &controlSchedulesNode{
		rows:    input.(planNode),
		command: command,
	}, nil
}

func (ef *execFactory) ConstructCancelQueries(input exec.Node, ifExists bool) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551889)
	return &cancelQueriesNode{
		rows:     input.(planNode),
		ifExists: ifExists,
	}, nil
}

func (ef *execFactory) ConstructCancelSessions(input exec.Node, ifExists bool) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551890)
	return &cancelSessionsNode{
		rows:     input.(planNode),
		ifExists: ifExists,
	}, nil
}

func (ef *execFactory) ConstructCreateStatistics(cs *tree.CreateStats) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551891)
	ctx := ef.planner.extendedEvalCtx.Context
	if err := featureflag.CheckEnabled(
		ctx,
		ef.planner.ExecCfg(),
		featureStatsEnabled,
		"ANALYZE/CREATE STATISTICS",
	); err != nil {
		__antithesis_instrumentation__.Notify(551893)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551894)
	}
	__antithesis_instrumentation__.Notify(551892)

	runAsJob := !ef.isExplain && func() bool {
		__antithesis_instrumentation__.Notify(551895)
		return ef.planner.instrumentation.ShouldUseJobForCreateStats() == true
	}() == true

	return &createStatsNode{
		CreateStats: *cs,
		p:           ef.planner,
		runAsJob:    runAsJob,
	}, nil
}

func (ef *execFactory) ConstructExplain(
	options *tree.ExplainOptions,
	stmtType tree.StatementReturnType,
	buildFn exec.BuildPlanForExplainFn,
) (exec.Node, error) {
	__antithesis_instrumentation__.Notify(551896)
	if options.Flags[tree.ExplainFlagEnv] {
		__antithesis_instrumentation__.Notify(551902)
		return nil, errors.New("ENV only supported with (OPT) option")
	} else {
		__antithesis_instrumentation__.Notify(551903)
	}
	__antithesis_instrumentation__.Notify(551897)

	plan, err := buildFn(&execFactory{
		planner:   ef.planner,
		isExplain: true,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(551904)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(551905)
	}
	__antithesis_instrumentation__.Notify(551898)
	if options.Mode == tree.ExplainVec {
		__antithesis_instrumentation__.Notify(551906)
		wrappedPlan := plan.(*explain.Plan).WrappedPlan.(*planComponents)
		return &explainVecNode{
			options: options,
			plan:    *wrappedPlan,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(551907)
	}
	__antithesis_instrumentation__.Notify(551899)
	if options.Mode == tree.ExplainDDL {
		__antithesis_instrumentation__.Notify(551908)
		wrappedPlan := plan.(*explain.Plan).WrappedPlan.(*planComponents)
		return &explainDDLNode{
			options: options,
			plan:    *wrappedPlan,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(551909)
	}
	__antithesis_instrumentation__.Notify(551900)
	flags := explain.MakeFlags(options)
	if ef.planner.execCfg.TestingKnobs.DeterministicExplain {
		__antithesis_instrumentation__.Notify(551910)
		flags.Redact = explain.RedactVolatile
	} else {
		__antithesis_instrumentation__.Notify(551911)
	}
	__antithesis_instrumentation__.Notify(551901)
	n := &explainPlanNode{
		options: options,
		flags:   flags,
		plan:    plan.(*explain.Plan),
	}
	return n, nil
}

type renderBuilder struct {
	r   *renderNode
	res planNode
}

func (rb *renderBuilder) init(n exec.Node, reqOrdering exec.OutputOrdering) {
	src := asDataSource(n)
	rb.r = &renderNode{
		source: src,
	}
	rb.r.ivarHelper = tree.MakeIndexedVarHelper(rb.r, len(src.columns))
	rb.r.reqOrdering = ReqOrdering(reqOrdering)

	if spool, ok := rb.r.source.plan.(*spoolNode); ok {
		rb.r.source.plan = spool.source
		spool.source = rb.r
		rb.res = spool
	} else {
		rb.res = rb.r
	}
}

func (rb *renderBuilder) setOutput(exprs tree.TypedExprs, columns colinfo.ResultColumns) {
	__antithesis_instrumentation__.Notify(551912)
	rb.r.render = exprs
	rb.r.columns = columns
}

func makeColList(table cat.Table, cols exec.TableColumnOrdinalSet) []catalog.Column {
	__antithesis_instrumentation__.Notify(551913)
	tab := table.(optCatalogTableInterface)
	ret := make([]catalog.Column, 0, cols.Len())
	for i, n := 0, table.ColumnCount(); i < n; i++ {
		__antithesis_instrumentation__.Notify(551915)
		if !cols.Contains(i) {
			__antithesis_instrumentation__.Notify(551917)
			continue
		} else {
			__antithesis_instrumentation__.Notify(551918)
		}
		__antithesis_instrumentation__.Notify(551916)
		ret = append(ret, tab.getCol(i))
	}
	__antithesis_instrumentation__.Notify(551914)
	return ret
}

func makePublicToReturnColumnIndexMapping(
	tableDesc catalog.TableDescriptor, returnCols []catalog.Column,
) []int {
	__antithesis_instrumentation__.Notify(551919)
	return row.ColMapping(tableDesc.PublicColumns(), returnCols)
}
