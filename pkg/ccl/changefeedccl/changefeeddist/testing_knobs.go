package changefeeddist

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"

type TestingKnobs struct {
	OnDistflowSpec func(aggregatorSpecs []*execinfrapb.ChangeAggregatorSpec, frontierSpec *execinfrapb.ChangeFrontierSpec)
}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(16712) }
