package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/tenantdirsvr"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/spf13/cobra"
)

var mtTestDirectorySvr = &cobra.Command{
	Use:   "test-directory",
	Short: "run a test directory service",
	Long: `
Run a test directory service that starts and manages tenant SQL instances as
processes on the local machine.

Use two dashes (--) to separate the test directory command's arguments from
the remaining arguments that specify the executable (and the arguments) that 
will be ran when starting each tenant.

For example:
cockroach mt test-directory --port 1234 -- cockroach mt start-sql --kv-addrs=:2222 --certs-dir=./certs --base-dir=./base
or 
cockroach mt test-directory --port 1234 -- bash -c ./tenant_start.sh 

test-directory command will always add the following arguments (in that order):
--sql-addr <addr/host>[:<port>] 
--http-addr <addr/host>[:<port>]
--tenant-id number
`,
	Args: nil,
	RunE: clierrorplus.MaybeDecorateError(runDirectorySvr),
}

func runDirectorySvr(cmd *cobra.Command, args []string) (returnErr error) {
	__antithesis_instrumentation__.Notify(33496)
	ctx := context.Background()
	serverCfg.Stores.Specs = nil

	stopper, err := setupAndInitializeLoggingAndProfiling(ctx, cmd, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(33501)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33502)
	}
	__antithesis_instrumentation__.Notify(33497)
	defer stopper.Stop(ctx)

	tds, err := tenantdirsvr.New(stopper, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(33503)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33504)
	}
	__antithesis_instrumentation__.Notify(33498)

	listenPort, err := net.Listen(
		"tcp", fmt.Sprintf(":%d", testDirectorySvrContext.port),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(33505)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33506)
	}
	__antithesis_instrumentation__.Notify(33499)
	stopper.AddCloser(stop.CloserFn(func() { __antithesis_instrumentation__.Notify(33507); _ = listenPort.Close() }))
	__antithesis_instrumentation__.Notify(33500)
	return tds.Serve(listenPort)
}
