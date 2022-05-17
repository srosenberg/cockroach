package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/cockroachdb/cockroach/pkg/cmd/urlcheck/lib/urlcheck"
)

func main() {
	__antithesis_instrumentation__.Notify(53327)
	cmd := exec.Command("git", "grep", "-nE", urlcheck.URLRE)
	if err := urlcheck.CheckURLsFromGrepOutput(cmd); err != nil {
		__antithesis_instrumentation__.Notify(53329)
		log.Fatalf("%+v\nFAIL", err)
	} else {
		__antithesis_instrumentation__.Notify(53330)
	}
	__antithesis_instrumentation__.Notify(53328)
	fmt.Println("PASS")
}
