package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/cdctest"
)

func main() {
	__antithesis_instrumentation__.Notify(19517)
	reg := cdctest.StartTestSchemaRegistry()
	defer reg.Close()

	fmt.Printf("Stub Schema Registry listening at %s\n", reg.URL())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	sig := <-c
	fmt.Printf("Shutting down on signal: %s\n", sig)
}
