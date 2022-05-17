package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"os"

	_ "github.com/cockroachdb/cockroach/pkg/ccl/workloadccl/allccl"
	_ "github.com/cockroachdb/cockroach/pkg/ccl/workloadccl/cliccl"
	workloadcli "github.com/cockroachdb/cockroach/pkg/workload/cli"
)

func main() {
	__antithesis_instrumentation__.Notify(53359)
	if err := workloadcli.WorkloadCmd(false).Execute(); err != nil {
		__antithesis_instrumentation__.Notify(53360)

		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(53361)
	}
}
