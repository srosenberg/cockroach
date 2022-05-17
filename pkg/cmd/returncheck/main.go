package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"os"

	"github.com/cockroachdb/returncheck"
	"github.com/kisielk/gotool"
)

func main() {
	__antithesis_instrumentation__.Notify(42918)
	targetPkg := "github.com/cockroachdb/cockroach/pkg/roachpb"
	targetTypeName := "Error"
	if err := returncheck.Run(gotool.ImportPaths(os.Args[1:]), targetPkg, targetTypeName); err != nil {
		__antithesis_instrumentation__.Notify(42919)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(42920)
	}
}
