package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type unaryNode struct {
	run unaryRun
}

type unaryRun struct {
	consumed bool
}

func (*unaryNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(631215)
	return nil
}

func (*unaryNode) Values() tree.Datums { __antithesis_instrumentation__.Notify(631216); return nil }

func (u *unaryNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631217)
	r := !u.run.consumed
	u.run.consumed = true
	return r, nil
}

func (*unaryNode) Close(context.Context) { __antithesis_instrumentation__.Notify(631218) }
