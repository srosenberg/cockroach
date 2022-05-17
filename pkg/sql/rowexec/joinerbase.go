package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

type joinerBase struct {
	execinfra.ProcessorBase

	joinType    descpb.JoinType
	onCond      execinfrapb.ExprHelper
	emptyLeft   rowenc.EncDatumRow
	emptyRight  rowenc.EncDatumRow
	combinedRow rowenc.EncDatumRow
}

func (jb *joinerBase) init(
	self execinfra.RowSource,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	leftTypes []*types.T,
	rightTypes []*types.T,
	jType descpb.JoinType,
	onExpr execinfrapb.Expression,
	outputContinuationColumn bool,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
	opts execinfra.ProcStateOpts,
) error {
	jb.joinType = jType

	if jb.joinType.IsSetOpJoin() {
		if !onExpr.Empty() {
			return errors.Errorf("expected empty onExpr, got %v", onExpr)
		}
	}

	jb.emptyLeft = make(rowenc.EncDatumRow, len(leftTypes))
	for i := range jb.emptyLeft {
		jb.emptyLeft[i] = rowenc.DatumToEncDatum(leftTypes[i], tree.DNull)
	}
	jb.emptyRight = make(rowenc.EncDatumRow, len(rightTypes))
	for i := range jb.emptyRight {
		jb.emptyRight[i] = rowenc.DatumToEncDatum(rightTypes[i], tree.DNull)
	}

	rowSize := len(leftTypes) + len(rightTypes)
	if outputContinuationColumn {

		rowSize++
	}
	jb.combinedRow = make(rowenc.EncDatumRow, rowSize)

	onCondTypes := make([]*types.T, 0, len(leftTypes)+len(rightTypes))
	onCondTypes = append(onCondTypes, leftTypes...)
	onCondTypes = append(onCondTypes, rightTypes...)

	outputTypes := jType.MakeOutputTypes(leftTypes, rightTypes)
	if outputContinuationColumn {
		outputTypes = append(outputTypes, types.Bool)
	}

	if err := jb.ProcessorBase.Init(
		self, post, outputTypes, flowCtx, processorID, output, nil, opts,
	); err != nil {
		return err
	}
	semaCtx := flowCtx.NewSemaContext(flowCtx.EvalCtx.Txn)
	return jb.onCond.Init(onExpr, onCondTypes, semaCtx, jb.EvalCtx)
}

type joinSide uint8

const (
	leftSide joinSide = 0

	rightSide joinSide = 1
)

func (j joinSide) String() string {
	__antithesis_instrumentation__.Notify(573109)
	if j == leftSide {
		__antithesis_instrumentation__.Notify(573111)
		return "left"
	} else {
		__antithesis_instrumentation__.Notify(573112)
	}
	__antithesis_instrumentation__.Notify(573110)
	return "right"
}

func (jb *joinerBase) renderUnmatchedRow(row rowenc.EncDatumRow, side joinSide) rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(573113)
	lrow, rrow := jb.emptyLeft, jb.emptyRight
	if side == leftSide {
		__antithesis_instrumentation__.Notify(573115)
		lrow = row
	} else {
		__antithesis_instrumentation__.Notify(573116)
		rrow = row
	}
	__antithesis_instrumentation__.Notify(573114)

	return jb.renderForOutput(lrow, rrow)
}

func shouldEmitUnmatchedRow(side joinSide, joinType descpb.JoinType) bool {
	__antithesis_instrumentation__.Notify(573117)
	switch joinType {
	case descpb.InnerJoin, descpb.LeftSemiJoin, descpb.RightSemiJoin, descpb.IntersectAllJoin:
		__antithesis_instrumentation__.Notify(573118)
		return false
	case descpb.RightOuterJoin:
		__antithesis_instrumentation__.Notify(573119)
		return side == rightSide
	case descpb.LeftOuterJoin:
		__antithesis_instrumentation__.Notify(573120)
		return side == leftSide
	case descpb.LeftAntiJoin:
		__antithesis_instrumentation__.Notify(573121)
		return side == leftSide
	case descpb.RightAntiJoin:
		__antithesis_instrumentation__.Notify(573122)
		return side == rightSide
	case descpb.ExceptAllJoin:
		__antithesis_instrumentation__.Notify(573123)
		return side == leftSide
	case descpb.FullOuterJoin:
		__antithesis_instrumentation__.Notify(573124)
		return true
	default:
		__antithesis_instrumentation__.Notify(573125)
		panic(errors.AssertionFailedf("unexpected join type %s", joinType))
	}
}

func (jb *joinerBase) render(lrow, rrow rowenc.EncDatumRow) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(573126)
	outputRow := jb.renderForOutput(lrow, rrow)

	if jb.onCond.Expr != nil {
		__antithesis_instrumentation__.Notify(573128)

		var combinedRow rowenc.EncDatumRow
		if jb.joinType.ShouldIncludeLeftColsInOutput() && func() bool {
			__antithesis_instrumentation__.Notify(573130)
			return jb.joinType.ShouldIncludeRightColsInOutput() == true
		}() == true {
			__antithesis_instrumentation__.Notify(573131)

			combinedRow = outputRow
		} else {
			__antithesis_instrumentation__.Notify(573132)
			combinedRow = jb.combine(lrow, rrow)
		}
		__antithesis_instrumentation__.Notify(573129)
		res, err := jb.onCond.EvalFilter(combinedRow)
		if !res || func() bool {
			__antithesis_instrumentation__.Notify(573133)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(573134)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(573135)
		}
	} else {
		__antithesis_instrumentation__.Notify(573136)
	}
	__antithesis_instrumentation__.Notify(573127)

	return outputRow, nil
}

func (jb *joinerBase) combine(lrow, rrow rowenc.EncDatumRow) rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(573137)
	jb.combinedRow = jb.combinedRow[:len(lrow)+len(rrow)]

	if len(lrow) == 1 {
		__antithesis_instrumentation__.Notify(573140)
		jb.combinedRow[0] = lrow[0]
	} else {
		__antithesis_instrumentation__.Notify(573141)
		copy(jb.combinedRow, lrow)
	}
	__antithesis_instrumentation__.Notify(573138)
	if len(rrow) == 1 {
		__antithesis_instrumentation__.Notify(573142)
		jb.combinedRow[len(lrow)] = rrow[0]
	} else {
		__antithesis_instrumentation__.Notify(573143)
		copy(jb.combinedRow[len(lrow):], rrow)
	}
	__antithesis_instrumentation__.Notify(573139)
	return jb.combinedRow
}

func (jb *joinerBase) renderForOutput(lrow, rrow rowenc.EncDatumRow) rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(573144)
	if !jb.joinType.ShouldIncludeLeftColsInOutput() {
		__antithesis_instrumentation__.Notify(573147)
		return rrow
	} else {
		__antithesis_instrumentation__.Notify(573148)
	}
	__antithesis_instrumentation__.Notify(573145)
	if !jb.joinType.ShouldIncludeRightColsInOutput() {
		__antithesis_instrumentation__.Notify(573149)
		return lrow
	} else {
		__antithesis_instrumentation__.Notify(573150)
	}
	__antithesis_instrumentation__.Notify(573146)
	return jb.combine(lrow, rrow)
}

func (jb *joinerBase) addColumnsNeededByOnExpr(neededCols *util.FastIntSet, startIdx, endIdx int) {
	__antithesis_instrumentation__.Notify(573151)
	for _, v := range jb.onCond.Vars.GetIndexedVars() {
		__antithesis_instrumentation__.Notify(573152)
		if v.Used && func() bool {
			__antithesis_instrumentation__.Notify(573153)
			return v.Idx >= startIdx == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(573154)
			return v.Idx < endIdx == true
		}() == true {
			__antithesis_instrumentation__.Notify(573155)
			neededCols.Add(v.Idx - startIdx)
		} else {
			__antithesis_instrumentation__.Notify(573156)
		}
	}
}
