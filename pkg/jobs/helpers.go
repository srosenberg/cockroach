package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/jobs/jobspb"

func ResetConstructors() func() {
	__antithesis_instrumentation__.Notify(70259)
	old := make(map[jobspb.Type]Constructor)
	for k, v := range constructors {
		__antithesis_instrumentation__.Notify(70261)
		old[k] = v
	}
	__antithesis_instrumentation__.Notify(70260)
	return func() { __antithesis_instrumentation__.Notify(70262); constructors = old }
}
