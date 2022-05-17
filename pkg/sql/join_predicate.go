package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
)

type joinPredicate struct {
	_ util.NoCopy

	joinType descpb.JoinType

	numLeftCols, numRightCols int

	leftEqualityIndices  []exec.NodeColumnOrdinal
	rightEqualityIndices []exec.NodeColumnOrdinal

	leftColNames tree.NameList

	rightColNames tree.NameList

	iVarHelper tree.IndexedVarHelper
	curRow     tree.Datums

	onCond tree.TypedExpr

	leftCols  colinfo.ResultColumns
	rightCols colinfo.ResultColumns
	cols      colinfo.ResultColumns

	leftEqKey bool

	rightEqKey bool
}

func getJoinResultColumns(
	joinType descpb.JoinType, left, right colinfo.ResultColumns,
) colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(498813)
	columns := make(colinfo.ResultColumns, 0, len(left)+len(right))
	if joinType.ShouldIncludeLeftColsInOutput() {
		__antithesis_instrumentation__.Notify(498816)
		columns = append(columns, left...)
	} else {
		__antithesis_instrumentation__.Notify(498817)
	}
	__antithesis_instrumentation__.Notify(498814)
	if joinType.ShouldIncludeRightColsInOutput() {
		__antithesis_instrumentation__.Notify(498818)
		columns = append(columns, right...)
	} else {
		__antithesis_instrumentation__.Notify(498819)
	}
	__antithesis_instrumentation__.Notify(498815)
	return columns
}

func makePredicate(joinType descpb.JoinType, left, right colinfo.ResultColumns) *joinPredicate {
	__antithesis_instrumentation__.Notify(498820)
	pred := &joinPredicate{
		joinType:     joinType,
		numLeftCols:  len(left),
		numRightCols: len(right),
		leftCols:     left,
		rightCols:    right,
		cols:         getJoinResultColumns(joinType, left, right),
	}

	pred.curRow = make(tree.Datums, len(left)+len(right))
	pred.iVarHelper = tree.MakeIndexedVarHelper(pred, len(pred.curRow))

	return pred
}

func (p *joinPredicate) IndexedVarEval(idx int, ctx *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(498821)
	return p.curRow[idx].Eval(ctx)
}

func (p *joinPredicate) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(498822)
	if idx < p.numLeftCols {
		__antithesis_instrumentation__.Notify(498824)
		return p.leftCols[idx].Typ
	} else {
		__antithesis_instrumentation__.Notify(498825)
	}
	__antithesis_instrumentation__.Notify(498823)
	return p.rightCols[idx-p.numLeftCols].Typ
}

func (p *joinPredicate) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(498826)
	if idx < p.numLeftCols {
		__antithesis_instrumentation__.Notify(498828)
		return p.leftCols.NodeFormatter(idx)
	} else {
		__antithesis_instrumentation__.Notify(498829)
	}
	__antithesis_instrumentation__.Notify(498827)
	return p.rightCols.NodeFormatter(idx - p.numLeftCols)
}

func (p *joinPredicate) eval(ctx *tree.EvalContext, leftRow, rightRow tree.Datums) (bool, error) {
	__antithesis_instrumentation__.Notify(498830)
	if p.onCond != nil {
		__antithesis_instrumentation__.Notify(498832)
		copy(p.curRow[:len(leftRow)], leftRow)
		copy(p.curRow[len(leftRow):], rightRow)
		ctx.PushIVarContainer(p.iVarHelper.Container())
		pred, err := execinfrapb.RunFilter(p.onCond, ctx)
		ctx.PopIVarContainer()
		return pred, err
	} else {
		__antithesis_instrumentation__.Notify(498833)
	}
	__antithesis_instrumentation__.Notify(498831)
	return true, nil
}

func (p *joinPredicate) prepareRow(result, leftRow, rightRow tree.Datums) {
	__antithesis_instrumentation__.Notify(498834)
	copy(result[:len(leftRow)], leftRow)
	copy(result[len(leftRow):], rightRow)
}
