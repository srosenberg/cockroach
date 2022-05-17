package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var quitCmd = &cobra.Command{
	Use:   "quit",
	Short: "drain and shut down a node\n",
	Long: `
Shut down the server. The first stage is drain, where the server stops accepting
client connections, then stops extant connections, and finally pushes range
leases onto other nodes, subject to various timeout parameters configurable via
cluster settings. After the first stage completes, the server process is shut
down.

If an argument is specified, the command affects the node
whose ID is given. If --self is specified, the command
affects the node that the command is connected to (via --host).
`,
	Args: cobra.MaximumNArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runQuit),
	Deprecated: `see 'cockroach node drain' instead to drain a 
server without terminating the server process (which can in turn be done using 
an orchestration layer or a process manager, or by sending a termination signal
directly).`,
}

func runQuit(cmd *cobra.Command, args []string) (err error) {
	__antithesis_instrumentation__.Notify(33827)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		__antithesis_instrumentation__.Notify(33833)
		if err == nil {
			__antithesis_instrumentation__.Notify(33834)
			fmt.Println("ok")
		} else {
			__antithesis_instrumentation__.Notify(33835)
		}
	}()
	__antithesis_instrumentation__.Notify(33828)

	if !quitCtx.nodeDrainSelf && func() bool {
		__antithesis_instrumentation__.Notify(33836)
		return len(args) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(33837)
		fmt.Fprintf(stderr, "warning: draining a node without node ID or passing --self explicitly is deprecated.\n")
		quitCtx.nodeDrainSelf = true
	} else {
		__antithesis_instrumentation__.Notify(33838)
	}
	__antithesis_instrumentation__.Notify(33829)
	if quitCtx.nodeDrainSelf && func() bool {
		__antithesis_instrumentation__.Notify(33839)
		return len(args) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(33840)
		return errors.Newf("cannot use --%s with an explicit node ID", cliflags.NodeDrainSelf.Name)
	} else {
		__antithesis_instrumentation__.Notify(33841)
	}
	__antithesis_instrumentation__.Notify(33830)

	targetNode := "local"
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(33842)
		targetNode = args[0]
	} else {
		__antithesis_instrumentation__.Notify(33843)
	}
	__antithesis_instrumentation__.Notify(33831)

	c, finish, err := getAdminClient(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(33844)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33845)
	}
	__antithesis_instrumentation__.Notify(33832)
	defer finish()

	return drainAndShutdown(ctx, c, targetNode)
}

func drainAndShutdown(ctx context.Context, c serverpb.AdminClient, targetNode string) (err error) {
	__antithesis_instrumentation__.Notify(33846)
	hardError, remainingWork, err := doDrain(ctx, c, targetNode)
	if hardError {
		__antithesis_instrumentation__.Notify(33851)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33852)
	}
	__antithesis_instrumentation__.Notify(33847)

	if remainingWork {
		__antithesis_instrumentation__.Notify(33853)
		log.Warningf(ctx, "graceful shutdown may not have completed successfully; check the node's logs for details.")
	} else {
		__antithesis_instrumentation__.Notify(33854)
	}
	__antithesis_instrumentation__.Notify(33848)

	if err != nil {
		__antithesis_instrumentation__.Notify(33855)
		log.Warningf(ctx, "drain did not complete successfully; hard shutdown may cause disruption")
	} else {
		__antithesis_instrumentation__.Notify(33856)
	}
	__antithesis_instrumentation__.Notify(33849)

	hardErr, err := doShutdown(ctx, c, targetNode)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(33857)
		return !hardErr == true
	}() == true {
		__antithesis_instrumentation__.Notify(33858)
		log.Warningf(ctx, "hard shutdown attempt failed, retrying: %v", err)
		_, err = doShutdown(ctx, c, targetNode)
	} else {
		__antithesis_instrumentation__.Notify(33859)
	}
	__antithesis_instrumentation__.Notify(33850)
	return errors.Wrap(err, "hard shutdown failed")
}

func doDrain(
	ctx context.Context, c serverpb.AdminClient, targetNode string,
) (hardError, remainingWork bool, err error) {
	__antithesis_instrumentation__.Notify(33860)

	if quitCtx.drainWait == 0 {
		__antithesis_instrumentation__.Notify(33864)
		return doDrainNoTimeout(ctx, c, targetNode)
	} else {
		__antithesis_instrumentation__.Notify(33865)
	}
	__antithesis_instrumentation__.Notify(33861)

	err = contextutil.RunWithTimeout(ctx, "drain", quitCtx.drainWait, func(ctx context.Context) (err error) {
		__antithesis_instrumentation__.Notify(33866)
		hardError, remainingWork, err = doDrainNoTimeout(ctx, c, targetNode)
		return err
	})
	__antithesis_instrumentation__.Notify(33862)
	if errors.HasType(err, (*contextutil.TimeoutError)(nil)) || func() bool {
		__antithesis_instrumentation__.Notify(33867)
		return grpcutil.IsTimeout(err) == true
	}() == true {
		__antithesis_instrumentation__.Notify(33868)
		log.Infof(ctx, "drain timed out: %v", err)
		err = errors.New("drain timeout, consider adjusting --drain-wait, especially under " +
			"custom server.shutdown.{drain,query,connection,lease_transfer}_wait cluster settings")
	} else {
		__antithesis_instrumentation__.Notify(33869)
	}
	__antithesis_instrumentation__.Notify(33863)
	return
}

func doDrainNoTimeout(
	ctx context.Context, c serverpb.AdminClient, targetNode string,
) (hardError, remainingWork bool, err error) {
	__antithesis_instrumentation__.Notify(33870)
	defer func() {
		__antithesis_instrumentation__.Notify(33873)
		if server.IsWaitingForInit(err) {
			__antithesis_instrumentation__.Notify(33874)
			log.Infof(ctx, "%v", err)
			err = errors.New("node cannot be drained before it has been initialized")
		} else {
			__antithesis_instrumentation__.Notify(33875)
		}
	}()
	__antithesis_instrumentation__.Notify(33871)

	var (
		remaining     = uint64(math.MaxUint64)
		prevRemaining = uint64(math.MaxUint64)
		verbose       = false
	)
	for ; ; prevRemaining = remaining {
		__antithesis_instrumentation__.Notify(33876)

		fmt.Fprintf(stderr, "node is draining... ")

		stream, err := c.Drain(ctx, &serverpb.DrainRequest{
			DoDrain:  true,
			Shutdown: false,
			NodeId:   targetNode,
			Verbose:  verbose,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(33881)
			fmt.Fprintf(stderr, "\n")
			return !grpcutil.IsTimeout(err), remaining > 0, errors.Wrap(err, "error sending drain request")
		} else {
			__antithesis_instrumentation__.Notify(33882)
		}
		__antithesis_instrumentation__.Notify(33877)
		for {
			__antithesis_instrumentation__.Notify(33883)
			resp, err := stream.Recv()
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(33887)

				break
			} else {
				__antithesis_instrumentation__.Notify(33888)
			}
			__antithesis_instrumentation__.Notify(33884)
			if err != nil {
				__antithesis_instrumentation__.Notify(33889)

				fmt.Fprintf(stderr, "\n")
				log.Infof(ctx, "graceful drain failed: %v", err)
				return false, remaining > 0, err
			} else {
				__antithesis_instrumentation__.Notify(33890)
			}
			__antithesis_instrumentation__.Notify(33885)

			if resp.IsDraining {
				__antithesis_instrumentation__.Notify(33891)

				remaining = resp.DrainRemainingIndicator
				finalString := ""
				if remaining == 0 {
					__antithesis_instrumentation__.Notify(33893)
					finalString = " (complete)"
				} else {
					__antithesis_instrumentation__.Notify(33894)
				}
				__antithesis_instrumentation__.Notify(33892)

				fmt.Fprintf(stderr, "remaining: %d%s\n", remaining, finalString)
			} else {
				__antithesis_instrumentation__.Notify(33895)

				remaining = 0
				fmt.Fprintf(stderr, "done\n")
			}
			__antithesis_instrumentation__.Notify(33886)

			if resp.DrainRemainingDescription != "" {
				__antithesis_instrumentation__.Notify(33896)

				log.Infof(ctx, "drain details: %s\n", resp.DrainRemainingDescription)
			} else {
				__antithesis_instrumentation__.Notify(33897)
			}

		}
		__antithesis_instrumentation__.Notify(33878)
		if remaining == 0 {
			__antithesis_instrumentation__.Notify(33898)

			break
		} else {
			__antithesis_instrumentation__.Notify(33899)
		}
		__antithesis_instrumentation__.Notify(33879)

		if remaining >= prevRemaining {
			__antithesis_instrumentation__.Notify(33900)
			verbose = true
		} else {
			__antithesis_instrumentation__.Notify(33901)
		}
		__antithesis_instrumentation__.Notify(33880)

		time.Sleep(200 * time.Millisecond)
	}
	__antithesis_instrumentation__.Notify(33872)
	return false, remaining > 0, nil
}

func doShutdown(
	ctx context.Context, c serverpb.AdminClient, targetNode string,
) (hardError bool, err error) {
	__antithesis_instrumentation__.Notify(33902)
	defer func() {
		__antithesis_instrumentation__.Notify(33906)
		if err != nil {
			__antithesis_instrumentation__.Notify(33907)
			if server.IsWaitingForInit(err) {
				__antithesis_instrumentation__.Notify(33909)
				log.Infof(ctx, "encountered error: %v", err)
				err = errors.New("node cannot be shut down before it has been initialized")
				err = errors.WithHint(err, "You can still stop the process using a service manager or a signal.")
				hardError = true
			} else {
				__antithesis_instrumentation__.Notify(33910)
			}
			__antithesis_instrumentation__.Notify(33908)
			if grpcutil.IsClosedConnection(err) {
				__antithesis_instrumentation__.Notify(33911)

				err = nil
			} else {
				__antithesis_instrumentation__.Notify(33912)
			}
		} else {
			__antithesis_instrumentation__.Notify(33913)
		}
	}()
	__antithesis_instrumentation__.Notify(33903)

	err = contextutil.RunWithTimeout(ctx, "hard shutdown", 10*time.Second, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(33914)

		stream, err := c.Drain(ctx, &serverpb.DrainRequest{NodeId: targetNode, Shutdown: true})
		if err != nil {
			__antithesis_instrumentation__.Notify(33916)
			return errors.Wrap(err, "error sending shutdown request")
		} else {
			__antithesis_instrumentation__.Notify(33917)
		}
		__antithesis_instrumentation__.Notify(33915)
		for {
			__antithesis_instrumentation__.Notify(33918)
			_, err := stream.Recv()
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(33920)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(33921)
			}
			__antithesis_instrumentation__.Notify(33919)
			if err != nil {
				__antithesis_instrumentation__.Notify(33922)
				return err
			} else {
				__antithesis_instrumentation__.Notify(33923)
			}
		}
	})
	__antithesis_instrumentation__.Notify(33904)
	if !errors.HasType(err, (*contextutil.TimeoutError)(nil)) {
		__antithesis_instrumentation__.Notify(33924)
		hardError = true
	} else {
		__antithesis_instrumentation__.Notify(33925)
	}
	__antithesis_instrumentation__.Notify(33905)
	return hardError, err
}

func getAdminClient(ctx context.Context, cfg server.Config) (serverpb.AdminClient, func(), error) {
	__antithesis_instrumentation__.Notify(33926)
	conn, _, finish, err := getClientGRPCConn(ctx, cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(33928)
		return nil, nil, errors.Wrap(err, "failed to connect to the node")
	} else {
		__antithesis_instrumentation__.Notify(33929)
	}
	__antithesis_instrumentation__.Notify(33927)
	return serverpb.NewAdminClient(conn), finish, nil
}
