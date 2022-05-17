package colexectestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecargs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
)

func MakeInputs(sources []colexecop.Operator) []colexecargs.OpWithMetaInfo {
	__antithesis_instrumentation__.Notify(431003)
	inputs := make([]colexecargs.OpWithMetaInfo, len(sources))
	for i := range sources {
		__antithesis_instrumentation__.Notify(431005)
		inputs[i].Root = sources[i]
	}
	__antithesis_instrumentation__.Notify(431004)
	return inputs
}
