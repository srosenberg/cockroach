// This is the default entry point for a CockroachDB binary.
//
// The ccl hook import below means building this will produce CCL'ed binaries.
// This file itself remains Apache2 to preserve the organization of ccl code
// under the /pkg/ccl subtree, but is unused for pure FLOSS builds.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	_ "github.com/cockroachdb/cockroach/pkg/ccl"
	_ "github.com/cockroachdb/cockroach/pkg/ccl/cliccl"
	"github.com/cockroachdb/cockroach/pkg/cli"
	_ "github.com/cockroachdb/cockroach/pkg/ui/distccl"
)

func main() {
	__antithesis_instrumentation__.Notify(38141)
	cli.Main()
}
