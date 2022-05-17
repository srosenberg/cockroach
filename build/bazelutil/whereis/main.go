// Copyright 2022 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package main

import __antithesis_instrumentation__ "antithesis.com/go/instrumentation/instrumented_module_4d5c339ebeff"

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	__antithesis_instrumentation__.Notify(1)
	if len(os.Args) != 2 {
		__antithesis_instrumentation__.Notify(4)
		panic("expected a single argument")
	} else {
		__antithesis_instrumentation__.Notify(5)
	}
	__antithesis_instrumentation__.Notify(2)
	abs, err := filepath.Abs(os.Args[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(6)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(7)
	}
	__antithesis_instrumentation__.Notify(3)
	fmt.Printf("%s\n", abs)
}
