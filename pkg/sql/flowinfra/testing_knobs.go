package flowinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type TestingKnobs struct {
	FlowRegistryDraining func() bool
}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(492172) }
