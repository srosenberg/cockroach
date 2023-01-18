// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// cockroach-short is an entry point for a CockroachDB binary that excludes
// certain components that are slow to build or have heavyweight dependencies.
// At present, the only excluded component is the web UI.
//
// The ccl hook import below means building this will produce CCL'ed binaries.
// This file itself remains Apache2 to preserve the organization of ccl code
// under the /pkg/ccl subtree, but is unused for pure FLOSS builds.
package main

import (
	"os"
	"fmt"
	"flag"
	"github.com/bazelbuild/rules_go/go/tools/bzltestutil"
	"github.com/bazelbuild/rules_go/go/tools/coverdata"
	_ "github.com/cockroachdb/cockroach/pkg/ccl" // ccl init hooks
	"github.com/cockroachdb/cockroach/pkg/cli"
	"sync/atomic"
)

// coverReport reports the coverage percentage and writes a coverage profile if requested.
func coverReport() {
	var f *os.File
	var err error
	/*if *coverProfile != "" {
	                f, err = os.Create(toOutputDir(*coverProfile))
			                mustBeNil(err)
					                fmt.Fprintf(f, "mode: %s\n", cover.Mode)
							                defer func() { mustBeNil(f.Close()) }()
									        }*/
	f, err = os.Create("/tmp/roachtest.cover")
	mustBeNil(err)
	defer func() { mustBeNil(f.Close()) }()

	var active, total int64
	var count uint32
	for name, counts := range coverdata.Counters {
		blocks := coverdata.Blocks[name]
		for i := range counts {
			stmts := int64(blocks[i].Stmts)
			total += stmts
			count = atomic.LoadUint32(&counts[i]) // For -mode=atomic.
			if count > 0 {
				active += stmts
			}
			if f != nil {
				_, err := fmt.Fprintf(f, "%s:%d.%d,%d.%d %d %d\n", name,
					blocks[i].Line0, blocks[i].Col0,
					blocks[i].Line1, blocks[i].Col1,
					stmts,
					count)
				mustBeNil(err)
			}
		}
	}
	if total == 0 {
		fmt.Println("coverage: [no statements]")
		return
	}
	//fmt.Printf("coverage: %.1f%% of statements%s\n", 100*float64(active)/float64(total), coverdata.CoveredPackages)

	_ = flag.String("test.coverprofile", "", "write a coverage profile to `file`")
	flag.Lookup("test.coverprofile").Value.Set("/tmp/roachtest" + ".cover")
	bzltestutil.ConvertCoverToLcov()
}

// mustBeNil checks the error and, if present, reports it and exits.
func mustBeNil(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "testing: %s\n", err)
		os.Exit(2)
	}
}

func main() {
	defer coverReport()

	cli.Main()
}
