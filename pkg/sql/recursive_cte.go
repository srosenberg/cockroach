package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type recursiveCTENode struct {
	initial planNode

	genIterationFn exec.RecursiveCTEIterationFn

	label string

	deduplicate bool

	recursiveCTERun
}

type recursiveCTERun struct {
	typs []*types.T

	workingRows rowContainerHelper
	iterator    *rowContainerIterator
	currentRow  tree.Datums

	allRows rowContainerHelper

	initialDone bool
	done        bool

	err error
}

func (n *recursiveCTENode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(564926)
	n.typs = planTypes(n.initial)
	n.workingRows.Init(n.typs, params.extendedEvalCtx, "cte")
	if n.deduplicate {
		__antithesis_instrumentation__.Notify(564928)
		n.allRows.InitWithDedup(n.typs, params.extendedEvalCtx, "cte-all")
	} else {
		__antithesis_instrumentation__.Notify(564929)
	}
	__antithesis_instrumentation__.Notify(564927)
	return nil
}

func (n *recursiveCTENode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(564930)
	if err := params.p.cancelChecker.Check(); err != nil {
		__antithesis_instrumentation__.Notify(564940)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(564941)
	}
	__antithesis_instrumentation__.Notify(564931)

	if !n.initialDone {
		__antithesis_instrumentation__.Notify(564942)

		for {
			__antithesis_instrumentation__.Notify(564944)
			ok, err := n.initial.Next(params)
			if err != nil {
				__antithesis_instrumentation__.Notify(564947)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(564948)
			}
			__antithesis_instrumentation__.Notify(564945)
			if !ok {
				__antithesis_instrumentation__.Notify(564949)
				break
			} else {
				__antithesis_instrumentation__.Notify(564950)
			}
			__antithesis_instrumentation__.Notify(564946)
			if err := n.AddRow(params.ctx, n.initial.Values()); err != nil {
				__antithesis_instrumentation__.Notify(564951)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(564952)
			}
		}
		__antithesis_instrumentation__.Notify(564943)
		n.iterator = newRowContainerIterator(params.ctx, n.workingRows, n.typs)
		n.initialDone = true
	} else {
		__antithesis_instrumentation__.Notify(564953)
	}
	__antithesis_instrumentation__.Notify(564932)

	if n.done {
		__antithesis_instrumentation__.Notify(564954)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(564955)
	}
	__antithesis_instrumentation__.Notify(564933)

	if n.workingRows.Len() == 0 {
		__antithesis_instrumentation__.Notify(564956)

		n.done = true
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(564957)
	}
	__antithesis_instrumentation__.Notify(564934)

	var err error
	n.currentRow, err = n.iterator.Next()
	if err != nil {
		__antithesis_instrumentation__.Notify(564958)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(564959)
	}
	__antithesis_instrumentation__.Notify(564935)
	if n.currentRow != nil {
		__antithesis_instrumentation__.Notify(564960)

		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(564961)
	}
	__antithesis_instrumentation__.Notify(564936)

	n.iterator.Close()
	n.iterator = nil
	lastWorkingRows := n.workingRows
	defer lastWorkingRows.Close(params.ctx)

	n.workingRows = rowContainerHelper{}
	n.workingRows.Init(n.typs, params.extendedEvalCtx, "cte")

	buf := &bufferNode{

		plan:  n.initial,
		typs:  n.typs,
		rows:  lastWorkingRows,
		label: n.label,
	}
	newPlan, err := n.genIterationFn(newExecFactory(params.p), buf)
	if err != nil {
		__antithesis_instrumentation__.Notify(564962)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(564963)
	}
	__antithesis_instrumentation__.Notify(564937)

	if err := runPlanInsidePlan(params, newPlan.(*planComponents), rowResultWriter(n)); err != nil {
		__antithesis_instrumentation__.Notify(564964)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(564965)
	}
	__antithesis_instrumentation__.Notify(564938)

	n.iterator = newRowContainerIterator(params.ctx, n.workingRows, n.typs)
	n.currentRow, err = n.iterator.Next()
	if err != nil {
		__antithesis_instrumentation__.Notify(564966)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(564967)
	}
	__antithesis_instrumentation__.Notify(564939)
	return n.currentRow != nil, nil
}

func (n *recursiveCTENode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(564968)
	return n.currentRow
}

func (n *recursiveCTENode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(564969)
	n.initial.Close(ctx)
	if n.deduplicate {
		__antithesis_instrumentation__.Notify(564971)
		n.allRows.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(564972)
	}
	__antithesis_instrumentation__.Notify(564970)
	n.workingRows.Close(ctx)
	if n.iterator != nil {
		__antithesis_instrumentation__.Notify(564973)
		n.iterator.Close()
		n.iterator = nil
	} else {
		__antithesis_instrumentation__.Notify(564974)
	}
}

var _ rowResultWriter = (*recursiveCTENode)(nil)

func (n *recursiveCTENode) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(564975)
	if n.deduplicate {
		__antithesis_instrumentation__.Notify(564977)
		if ok, err := n.allRows.AddRowWithDedup(ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(564978)
			return err
		} else {
			__antithesis_instrumentation__.Notify(564979)
			if !ok {
				__antithesis_instrumentation__.Notify(564980)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(564981)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(564982)
	}
	__antithesis_instrumentation__.Notify(564976)
	return n.workingRows.AddRow(ctx, row)
}

func (n *recursiveCTENode) IncrementRowsAffected(context.Context, int) {
	__antithesis_instrumentation__.Notify(564983)
}

func (n *recursiveCTENode) SetError(err error) {
	__antithesis_instrumentation__.Notify(564984)
	n.err = err
}

func (n *recursiveCTENode) Err() error {
	__antithesis_instrumentation__.Notify(564985)
	return n.err
}
