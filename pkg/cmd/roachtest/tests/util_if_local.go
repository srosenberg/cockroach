package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"

func ifLocal(c cluster.Cluster, trueVal, falseVal string) string {
	__antithesis_instrumentation__.Notify(52249)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(52251)
		return trueVal
	} else {
		__antithesis_instrumentation__.Notify(52252)
	}
	__antithesis_instrumentation__.Notify(52250)
	return falseVal
}
