package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type bufferNode struct {
	plan planNode

	typs       []*types.T
	rows       rowContainerHelper
	currentRow tree.Datums

	label string
}

func (n *bufferNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(247116)
	n.typs = planTypes(n.plan)
	n.rows.Init(n.typs, params.extendedEvalCtx, n.label)
	return nil
}

func (n *bufferNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(247117)
	if err := params.p.cancelChecker.Check(); err != nil {
		__antithesis_instrumentation__.Notify(247122)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(247123)
	}
	__antithesis_instrumentation__.Notify(247118)
	ok, err := n.plan.Next(params)
	if err != nil {
		__antithesis_instrumentation__.Notify(247124)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(247125)
	}
	__antithesis_instrumentation__.Notify(247119)
	if !ok {
		__antithesis_instrumentation__.Notify(247126)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(247127)
	}
	__antithesis_instrumentation__.Notify(247120)
	n.currentRow = n.plan.Values()
	if err = n.rows.AddRow(params.ctx, n.currentRow); err != nil {
		__antithesis_instrumentation__.Notify(247128)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(247129)
	}
	__antithesis_instrumentation__.Notify(247121)
	return true, nil
}

func (n *bufferNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(247130)
	return n.currentRow
}

func (n *bufferNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(247131)
	n.plan.Close(ctx)
	n.rows.Close(ctx)
}

type scanBufferNode struct {
	buffer *bufferNode

	iterator   *rowContainerIterator
	currentRow tree.Datums

	label string
}

func (n *scanBufferNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(247132)
	n.iterator = newRowContainerIterator(params.ctx, n.buffer.rows, n.buffer.typs)
	return nil
}

func (n *scanBufferNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(247133)
	var err error
	n.currentRow, err = n.iterator.Next()
	if n.currentRow == nil || func() bool {
		__antithesis_instrumentation__.Notify(247135)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(247136)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(247137)
	}
	__antithesis_instrumentation__.Notify(247134)
	return true, nil
}

func (n *scanBufferNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(247138)
	return n.currentRow
}

func (n *scanBufferNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(247139)
	if n.iterator != nil {
		__antithesis_instrumentation__.Notify(247140)
		n.iterator.Close()
		n.iterator = nil
	} else {
		__antithesis_instrumentation__.Notify(247141)
	}
}
