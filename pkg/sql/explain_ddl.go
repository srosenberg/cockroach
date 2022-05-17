package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type explainDDLNode struct {
	optColumnsSlot
	options *tree.ExplainOptions
	plan    planComponents
	next    int
	values  []tree.Datums
}

func (n *explainDDLNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(491213)
	if n.next >= len(n.values) {
		__antithesis_instrumentation__.Notify(491215)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(491216)
	}
	__antithesis_instrumentation__.Notify(491214)
	n.next++
	return true, nil
}

func (n *explainDDLNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(491217)
	return n.values[n.next-1]
}

func (n *explainDDLNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(491218)
	n.next = len(n.values)
}

var _ planNode = (*explainDDLNode)(nil)

var explainNotPossibleError = pgerror.New(pgcode.FeatureNotSupported,
	"cannot explain a statement which is not supported by the declarative schema changer")

func (n *explainDDLNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(491219)

	scNode, ok := n.plan.main.planNode.(*schemaChangePlanNode)
	if !ok {
		__antithesis_instrumentation__.Notify(491221)
		if n.plan.main.physPlan == nil {
			__antithesis_instrumentation__.Notify(491222)
			return explainNotPossibleError
		} else {
			__antithesis_instrumentation__.Notify(491223)
			if len(n.plan.main.physPlan.planNodesToClose) > 0 {
				__antithesis_instrumentation__.Notify(491224)
				scNode, ok = n.plan.main.physPlan.planNodesToClose[0].(*schemaChangePlanNode)
				if !ok {
					__antithesis_instrumentation__.Notify(491225)
					return explainNotPossibleError
				} else {
					__antithesis_instrumentation__.Notify(491226)
				}
			} else {
				__antithesis_instrumentation__.Notify(491227)
				return explainNotPossibleError
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(491228)
	}
	__antithesis_instrumentation__.Notify(491220)
	return n.setExplainValues(scNode.plannedState)
}

func (n *explainDDLNode) setExplainValues(scState scpb.CurrentState) (err error) {
	__antithesis_instrumentation__.Notify(491229)
	defer func() {
		__antithesis_instrumentation__.Notify(491235)
		err = errors.WithAssertionFailure(err)
	}()
	__antithesis_instrumentation__.Notify(491230)
	var p scplan.Plan
	p, err = scplan.MakePlan(scState, scplan.Params{
		ExecutionPhase:             scop.StatementPhase,
		SchemaChangerJobIDSupplier: func() jobspb.JobID { __antithesis_instrumentation__.Notify(491236); return 1 },
	})
	__antithesis_instrumentation__.Notify(491231)
	if err != nil {
		__antithesis_instrumentation__.Notify(491237)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491238)
	}
	__antithesis_instrumentation__.Notify(491232)
	if n.options.Flags[tree.ExplainFlagViz] {
		__antithesis_instrumentation__.Notify(491239)
		stagesURL, depsURL, err := p.ExplainViz()
		n.values = []tree.Datums{
			{tree.NewDString(stagesURL)},
			{tree.NewDString(depsURL)},
		}
		return err
	} else {
		__antithesis_instrumentation__.Notify(491240)
	}
	__antithesis_instrumentation__.Notify(491233)

	var info string
	if n.options.Flags[tree.ExplainFlagVerbose] {
		__antithesis_instrumentation__.Notify(491241)
		info, err = p.ExplainVerbose()
	} else {
		__antithesis_instrumentation__.Notify(491242)
		info, err = p.ExplainCompact()
	}
	__antithesis_instrumentation__.Notify(491234)
	n.values = []tree.Datums{
		{tree.NewDString(info)},
	}
	return err
}
