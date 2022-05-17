// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
//
// bazci is glue code to make debugging Bazel builds and tests in Teamcity as
// painless as possible.
//
// bazci [build|test] \
//     --artifacts_dir=$ARTIFACTS_DIR targets... -- [command-line options]
//
// bazci will invoke a `bazel build` or `bazel test` of all the given targets
// and stage the resultant build/test artifacts in the given `artifacts_dir`.
// The build/test artifacts are munged slightly such that TC can easily parse
// them.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	__antithesis_instrumentation__.Notify(37658)
	log.SetFlags(0)
	log.SetPrefix("")

	if _, err := exec.LookPath("bazel"); err != nil {
		__antithesis_instrumentation__.Notify(37660)
		log.Printf("ERROR: bazel not found in $PATH")
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(37661)
	}
	__antithesis_instrumentation__.Notify(37659)

	if err := rootCmd.Execute(); err != nil {
		__antithesis_instrumentation__.Notify(37662)
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(37663)
	}
}
