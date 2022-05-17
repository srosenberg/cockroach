// cockroach-oss is an entry point for a CockroachDB binary that excludes all
// CCL-licensed code.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/cli"
	_ "github.com/cockroachdb/cockroach/pkg/ui/distoss"
)

func main() {
	__antithesis_instrumentation__.Notify(38142)
	cli.Main()
}
