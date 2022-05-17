package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize a cluster",
	Long: `
Perform one-time-only initialization of a CockroachDB cluster.

After starting one or more nodes with --join flags, run the init
command on one node (passing the same --host and certificate flags
you would use for the sql command).
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runInit),
}

func runInit(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(33193)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, finish, err := waitForClientReadinessAndGetClientGRPCConn(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(33196)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33197)
	}
	__antithesis_instrumentation__.Notify(33194)
	defer finish()

	c := serverpb.NewInitClient(conn)
	if _, err = c.Bootstrap(ctx, &serverpb.BootstrapRequest{}); err != nil {
		__antithesis_instrumentation__.Notify(33198)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33199)
	}
	__antithesis_instrumentation__.Notify(33195)

	fmt.Fprintln(os.Stdout, "Cluster successfully initialized")
	return nil
}

func waitForClientReadinessAndGetClientGRPCConn(
	ctx context.Context,
) (conn *grpc.ClientConn, finish func(), err error) {
	__antithesis_instrumentation__.Notify(33200)
	defer func() {
		__antithesis_instrumentation__.Notify(33203)

		if finish != nil && func() bool {
			__antithesis_instrumentation__.Notify(33204)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(33205)
			finish()
		} else {
			__antithesis_instrumentation__.Notify(33206)
		}
	}()
	__antithesis_instrumentation__.Notify(33201)

	retryOpts := retry.Options{InitialBackoff: time.Second, MaxBackoff: time.Second}
	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(33207)
		if err = contextutil.RunWithTimeout(ctx, "init-open-conn", 5*time.Second,
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(33209)

				conn, _, finish, err = getClientGRPCConn(ctx, serverCfg)
				if err != nil {
					__antithesis_instrumentation__.Notify(33211)
					return err
				} else {
					__antithesis_instrumentation__.Notify(33212)
				}
				__antithesis_instrumentation__.Notify(33210)

				ac := serverpb.NewAdminClient(conn)
				_, err := ac.Health(ctx, &serverpb.HealthRequest{})
				return err
			}); err != nil {
			__antithesis_instrumentation__.Notify(33213)
			err = errors.Wrapf(err, "node not ready to perform cluster initialization")
			fmt.Fprintln(stderr, "warning:", err, "(retrying)")

			if finish != nil {
				__antithesis_instrumentation__.Notify(33215)
				finish()
				finish = nil
			} else {
				__antithesis_instrumentation__.Notify(33216)
			}
			__antithesis_instrumentation__.Notify(33214)

			continue
		} else {
			__antithesis_instrumentation__.Notify(33217)
		}
		__antithesis_instrumentation__.Notify(33208)

		return conn, finish, err
	}
	__antithesis_instrumentation__.Notify(33202)
	err = errors.New("maximum number of retries exceeded")
	return
}
