package skip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/envutil"

var stress = envutil.EnvOrDefaultBool("COCKROACH_NIGHTLY_STRESS", false)

func NightlyStress() bool {
	__antithesis_instrumentation__.Notify(646020)
	return stress
}
