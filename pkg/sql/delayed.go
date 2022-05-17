package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/constraint"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type delayedNode struct {
	name            string
	columns         colinfo.ResultColumns
	indexConstraint *constraint.Constraint
	constructor     nodeConstructor
	plan            planNode
}

type nodeConstructor func(context.Context, *planner) (planNode, error)

func (d *delayedNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(465376)
	return d.plan.Next(params)
}
func (d *delayedNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(465377)
	return d.plan.Values()
}

func (d *delayedNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(465378)
	if d.plan != nil {
		__antithesis_instrumentation__.Notify(465379)
		d.plan.Close(ctx)
		d.plan = nil
	} else {
		__antithesis_instrumentation__.Notify(465380)
	}
}

func (d *delayedNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(465381)
	if d.plan != nil {
		__antithesis_instrumentation__.Notify(465384)
		panic("wrapped plan should not yet exist")
	} else {
		__antithesis_instrumentation__.Notify(465385)
	}
	__antithesis_instrumentation__.Notify(465382)

	plan, err := d.constructor(params.ctx, params.p)
	if err != nil {
		__antithesis_instrumentation__.Notify(465386)
		return err
	} else {
		__antithesis_instrumentation__.Notify(465387)
	}
	__antithesis_instrumentation__.Notify(465383)
	d.plan = plan

	return startExec(params, plan)
}
