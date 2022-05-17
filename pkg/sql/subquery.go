package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type subquery struct {
	subquery tree.NodeFormatter
	execMode rowexec.SubqueryExecMode
	expanded bool
	started  bool
	plan     planMaybePhysical

	rowCount int64
	result   tree.Datum
}

func (p *planner) EvalSubquery(expr *tree.Subquery) (result tree.Datum, err error) {
	__antithesis_instrumentation__.Notify(627612)
	if expr.Idx == 0 {
		__antithesis_instrumentation__.Notify(627616)
		return nil, errors.AssertionFailedf("subquery %q was not processed", expr)
	} else {
		__antithesis_instrumentation__.Notify(627617)
	}
	__antithesis_instrumentation__.Notify(627613)
	if expr.Idx < 0 || func() bool {
		__antithesis_instrumentation__.Notify(627618)
		return expr.Idx-1 >= len(p.curPlan.subqueryPlans) == true
	}() == true {
		__antithesis_instrumentation__.Notify(627619)
		return nil, errors.AssertionFailedf("subquery eval: invalid index %d for %q", expr.Idx, expr)
	} else {
		__antithesis_instrumentation__.Notify(627620)
	}
	__antithesis_instrumentation__.Notify(627614)

	s := &p.curPlan.subqueryPlans[expr.Idx-1]
	if !s.started {
		__antithesis_instrumentation__.Notify(627621)
		return nil, errors.AssertionFailedf("subquery %d (%q) not started prior to evaluation", expr.Idx, expr)
	} else {
		__antithesis_instrumentation__.Notify(627622)
	}
	__antithesis_instrumentation__.Notify(627615)
	return s.result, nil
}
