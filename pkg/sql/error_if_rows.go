package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type errorIfRowsNode struct {
	plan planNode

	mkErr exec.MkErrFn

	nexted bool
}

func (n *errorIfRowsNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(470026)
	return nil
}

func (n *errorIfRowsNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(470027)
	if n.nexted {
		__antithesis_instrumentation__.Notify(470031)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(470032)
	}
	__antithesis_instrumentation__.Notify(470028)
	n.nexted = true

	ok, err := n.plan.Next(params)
	if err != nil {
		__antithesis_instrumentation__.Notify(470033)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(470034)
	}
	__antithesis_instrumentation__.Notify(470029)
	if ok {
		__antithesis_instrumentation__.Notify(470035)
		return false, n.mkErr(n.plan.Values())
	} else {
		__antithesis_instrumentation__.Notify(470036)
	}
	__antithesis_instrumentation__.Notify(470030)
	return false, nil
}

func (n *errorIfRowsNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(470037)
	return nil
}

func (n *errorIfRowsNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(470038)
	n.plan.Close(ctx)
}
