package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/spf13/cobra"
)

var mtStartSQLProxyCmd = &cobra.Command{
	Use:   "start-proxy",
	Short: "start a sql proxy",
	Long: `Starts a SQL proxy.

This proxy accepts incoming connections and relays them to a backend server
determined by the arguments used.
`,
	RunE: clierrorplus.MaybeDecorateError(runStartSQLProxy),
	Args: cobra.NoArgs,
}

func runStartSQLProxy(cmd *cobra.Command, args []string) (returnErr error) {
	__antithesis_instrumentation__.Notify(33416)

	ctx, stopper, err := initLogging(cmd)
	if err != nil {
		__antithesis_instrumentation__.Notify(33424)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33425)
	}
	__antithesis_instrumentation__.Notify(33417)
	defer stopper.Stop(ctx)

	log.Infof(ctx, "New proxy with opts: %+v", proxyContext)

	proxyLn, err := net.Listen("tcp", proxyContext.ListenAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(33426)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33427)
	}
	__antithesis_instrumentation__.Notify(33418)

	metricsLn, err := net.Listen("tcp", proxyContext.MetricsAddress)
	if err != nil {
		__antithesis_instrumentation__.Notify(33428)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33429)
	}
	__antithesis_instrumentation__.Notify(33419)
	stopper.AddCloser(stop.CloserFn(func() { __antithesis_instrumentation__.Notify(33430); _ = metricsLn.Close() }))
	__antithesis_instrumentation__.Notify(33420)

	server, err := sqlproxyccl.NewServer(ctx, stopper, proxyContext)
	if err != nil {
		__antithesis_instrumentation__.Notify(33431)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33432)
	}
	__antithesis_instrumentation__.Notify(33421)

	errChan := make(chan error, 1)

	if err := stopper.RunAsyncTask(ctx, "serve-http", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(33433)
		log.Infof(ctx, "HTTP metrics server listening at %s", metricsLn.Addr())
		if err := server.ServeHTTP(ctx, metricsLn); err != nil {
			__antithesis_instrumentation__.Notify(33434)
			errChan <- err
		} else {
			__antithesis_instrumentation__.Notify(33435)
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(33436)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33437)
	}
	__antithesis_instrumentation__.Notify(33422)

	if err := stopper.RunAsyncTask(ctx, "serve-proxy", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(33438)
		log.Infof(ctx, "proxy server listening at %s", proxyLn.Addr())
		if err := server.Serve(ctx, proxyLn); err != nil {
			__antithesis_instrumentation__.Notify(33439)
			errChan <- err
		} else {
			__antithesis_instrumentation__.Notify(33440)
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(33441)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33442)
	}
	__antithesis_instrumentation__.Notify(33423)

	return waitForSignals(ctx, stopper, errChan)
}

func initLogging(cmd *cobra.Command) (ctx context.Context, stopper *stop.Stopper, err error) {
	__antithesis_instrumentation__.Notify(33443)

	serverCfg.Stores.Specs = nil
	serverCfg.ClusterName = ""

	ctx = context.Background()
	stopper, err = setupAndInitializeLoggingAndProfiling(ctx, cmd, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(33445)
		return
	} else {
		__antithesis_instrumentation__.Notify(33446)
	}
	__antithesis_instrumentation__.Notify(33444)
	ctx, _ = stopper.WithCancelOnQuiesce(ctx)
	return ctx, stopper, err
}

func waitForSignals(
	ctx context.Context, stopper *stop.Stopper, errChan chan error,
) (returnErr error) {
	__antithesis_instrumentation__.Notify(33447)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, drainSignals...)

	select {
	case err := <-errChan:
		__antithesis_instrumentation__.Notify(33450)
		log.StartAlwaysFlush()
		return err
	case <-stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(33451)

		<-stopper.IsStopped()

		log.StartAlwaysFlush()
	case sig := <-signalCh:
		__antithesis_instrumentation__.Notify(33452)
		log.StartAlwaysFlush()
		log.Ops.Infof(ctx, "received signal '%s'", sig)
		if sig == os.Interrupt {
			__antithesis_instrumentation__.Notify(33455)
			returnErr = errors.New("interrupted")
		} else {
			__antithesis_instrumentation__.Notify(33456)
		}
		__antithesis_instrumentation__.Notify(33453)
		go func() {
			__antithesis_instrumentation__.Notify(33457)
			log.Infof(ctx, "server stopping")
			stopper.Stop(ctx)
		}()
	case <-log.FatalChan():
		__antithesis_instrumentation__.Notify(33454)
		stopper.Stop(ctx)
		select {}
	}
	__antithesis_instrumentation__.Notify(33448)

	for {
		__antithesis_instrumentation__.Notify(33458)
		select {
		case sig := <-signalCh:
			__antithesis_instrumentation__.Notify(33460)
			switch sig {
			case os.Interrupt:
				__antithesis_instrumentation__.Notify(33463)
				log.Ops.Infof(ctx, "received additional signal '%s'; continuing graceful shutdown", sig)
				continue
			default:
				__antithesis_instrumentation__.Notify(33464)
			}
			__antithesis_instrumentation__.Notify(33461)

			log.Ops.Shoutf(ctx, severity.ERROR,
				"received signal '%s' during shutdown, initiating hard shutdown", redact.Safe(sig))
			panic("terminate")
		case <-stopper.IsStopped():
			__antithesis_instrumentation__.Notify(33462)
			const msgDone = "server shutdown completed"
			log.Ops.Infof(ctx, msgDone)
			fmt.Fprintln(os.Stdout, msgDone)
		}
		__antithesis_instrumentation__.Notify(33459)
		break
	}
	__antithesis_instrumentation__.Notify(33449)

	return returnErr
}
