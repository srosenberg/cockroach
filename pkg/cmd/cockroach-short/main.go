// cockroach-short is an entry point for a CockroachDB binary that excludes
// certain components that are slow to build or have heavyweight dependencies.
// At present, the only excluded component is the web UI.
//
// The ccl hook import below means building this will produce CCL'ed binaries.
// This file itself remains Apache2 to preserve the organization of ccl code
// under the /pkg/ccl subtree, but is unused for pure FLOSS builds.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	_ "github.com/cockroachdb/cockroach/pkg/ccl"
	"github.com/cockroachdb/cockroach/pkg/cli"
)

func main() {
	__antithesis_instrumentation__.Notify(38143)
	cli.Main()
}
