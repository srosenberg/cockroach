package scplan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/opgen"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/rules"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraph"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scstage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Params struct {
	InRollback bool

	ExecutionPhase scop.Phase

	SchemaChangerJobIDSupplier func() jobspb.JobID
}

type (
	// Graph is an exported alias of scgraph.Graph.
	Graph = scgraph.Graph

	// Stage is an exported alias of scstage.Stage.
	Stage = scstage.Stage
)

type Plan struct {
	scpb.CurrentState
	Params Params
	Graph  *scgraph.Graph
	JobID  jobspb.JobID
	Stages []Stage
}

func (p Plan) StagesForCurrentPhase() []scstage.Stage {
	__antithesis_instrumentation__.Notify(594780)
	for i, s := range p.Stages {
		__antithesis_instrumentation__.Notify(594782)
		if s.Phase > p.Params.ExecutionPhase {
			__antithesis_instrumentation__.Notify(594783)
			return p.Stages[:i]
		} else {
			__antithesis_instrumentation__.Notify(594784)
		}
	}
	__antithesis_instrumentation__.Notify(594781)
	return p.Stages
}

func MakePlan(initial scpb.CurrentState, params Params) (p Plan, err error) {
	__antithesis_instrumentation__.Notify(594785)
	p = Plan{
		CurrentState: initial,
		Params:       params,
	}
	defer func() {
		__antithesis_instrumentation__.Notify(594789)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(594790)
			rAsErr, ok := r.(error)
			if !ok {
				__antithesis_instrumentation__.Notify(594792)
				rAsErr = errors.Errorf("panic during MakePlan: %v", r)
			} else {
				__antithesis_instrumentation__.Notify(594793)
			}
			__antithesis_instrumentation__.Notify(594791)
			err = p.DecorateErrorWithPlanDetails(rAsErr)
		} else {
			__antithesis_instrumentation__.Notify(594794)
		}
	}()

	{
		__antithesis_instrumentation__.Notify(594795)
		start := timeutil.Now()
		p.Graph = buildGraph(p.CurrentState)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(594796)
			log.Infof(context.TODO(), "graph generation took %v", timeutil.Since(start))
		} else {
			__antithesis_instrumentation__.Notify(594797)
		}
	}
	{
		__antithesis_instrumentation__.Notify(594798)
		start := timeutil.Now()
		p.Stages = scstage.BuildStages(
			initial, params.ExecutionPhase, p.Graph, params.SchemaChangerJobIDSupplier,
		)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(594799)
			log.Infof(context.TODO(), "stage generation took %v", timeutil.Since(start))
		} else {
			__antithesis_instrumentation__.Notify(594800)
		}
	}
	__antithesis_instrumentation__.Notify(594786)
	if n := len(p.Stages); n > 0 && func() bool {
		__antithesis_instrumentation__.Notify(594801)
		return p.Stages[n-1].Phase > scop.PreCommitPhase == true
	}() == true {
		__antithesis_instrumentation__.Notify(594802)

		p.JobID = params.SchemaChangerJobIDSupplier()
	} else {
		__antithesis_instrumentation__.Notify(594803)
	}
	__antithesis_instrumentation__.Notify(594787)
	if err := scstage.ValidateStages(p.TargetState, p.Stages, p.Graph); err != nil {
		__antithesis_instrumentation__.Notify(594804)
		panic(errors.Wrapf(err, "invalid execution plan"))
	} else {
		__antithesis_instrumentation__.Notify(594805)
	}
	__antithesis_instrumentation__.Notify(594788)
	return p, nil
}

func buildGraph(cs scpb.CurrentState) *scgraph.Graph {
	__antithesis_instrumentation__.Notify(594806)
	g, err := opgen.BuildGraph(cs)
	if err != nil {
		__antithesis_instrumentation__.Notify(594811)
		panic(errors.Wrapf(err, "build graph op edges"))
	} else {
		__antithesis_instrumentation__.Notify(594812)
	}
	__antithesis_instrumentation__.Notify(594807)
	err = rules.ApplyDepRules(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(594813)
		panic(errors.Wrapf(err, "build graph dep edges"))
	} else {
		__antithesis_instrumentation__.Notify(594814)
	}
	__antithesis_instrumentation__.Notify(594808)
	err = g.Validate()
	if err != nil {
		__antithesis_instrumentation__.Notify(594815)
		panic(errors.Wrapf(err, "validate graph"))
	} else {
		__antithesis_instrumentation__.Notify(594816)
	}
	__antithesis_instrumentation__.Notify(594809)
	g, err = rules.ApplyOpRules(g)
	if err != nil {
		__antithesis_instrumentation__.Notify(594817)
		panic(errors.Wrapf(err, "mark op edges as no-op"))
	} else {
		__antithesis_instrumentation__.Notify(594818)
	}
	__antithesis_instrumentation__.Notify(594810)
	return g
}
