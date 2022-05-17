package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type valuesNode struct {
	columns colinfo.ResultColumns
	tuples  [][]tree.TypedExpr

	specifiedInQuery bool

	valuesRun
}

func (p *planner) newContainerValuesNode(columns colinfo.ResultColumns, capacity int) *valuesNode {
	__antithesis_instrumentation__.Notify(631662)
	return &valuesNode{
		columns: columns,
		valuesRun: valuesRun{
			rows: rowcontainer.NewRowContainerWithCapacity(
				p.EvalContext().Mon.MakeBoundAccount(),
				colinfo.ColTypeInfoFromResCols(columns),
				capacity,
			),
		},
	}
}

type valuesRun struct {
	rows    *rowcontainer.RowContainer
	nextRow int
}

func (n *valuesNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(631663)
	if n.rows != nil {
		__antithesis_instrumentation__.Notify(631666)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(631667)
	}
	__antithesis_instrumentation__.Notify(631664)

	n.rows = rowcontainer.NewRowContainerWithCapacity(
		params.extendedEvalCtx.Mon.MakeBoundAccount(),
		colinfo.ColTypeInfoFromResCols(n.columns),
		len(n.tuples),
	)

	row := make([]tree.Datum, len(n.columns))
	for _, tupleRow := range n.tuples {
		__antithesis_instrumentation__.Notify(631668)
		for i, typedExpr := range tupleRow {
			__antithesis_instrumentation__.Notify(631670)
			var err error
			row[i], err = typedExpr.Eval(params.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(631671)
				return err
			} else {
				__antithesis_instrumentation__.Notify(631672)
			}
		}
		__antithesis_instrumentation__.Notify(631669)
		if _, err := n.rows.AddRow(params.ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(631673)
			return err
		} else {
			__antithesis_instrumentation__.Notify(631674)
		}
	}
	__antithesis_instrumentation__.Notify(631665)

	return nil
}

func (n *valuesNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631675)
	if n.nextRow >= n.rows.Len() {
		__antithesis_instrumentation__.Notify(631677)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(631678)
	}
	__antithesis_instrumentation__.Notify(631676)
	n.nextRow++
	return true, nil
}

func (n *valuesNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(631679)
	return n.rows.At(n.nextRow - 1)
}

func (n *valuesNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(631680)
	if n.rows != nil {
		__antithesis_instrumentation__.Notify(631681)
		n.rows.Close(ctx)
		n.rows = nil
	} else {
		__antithesis_instrumentation__.Notify(631682)
	}
}
