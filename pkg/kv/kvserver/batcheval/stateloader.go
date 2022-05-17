package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"

func MakeStateLoader(rec EvalContext) stateloader.StateLoader {
	__antithesis_instrumentation__.Notify(97875)
	return stateloader.Make(rec.GetRangeID())
}
