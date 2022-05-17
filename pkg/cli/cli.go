package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	_ "github.com/cockroachdb/cockroach/pkg/cloud/impl"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"

	_ "github.com/cockroachdb/cockroach/pkg/workload/bank"
	_ "github.com/cockroachdb/cockroach/pkg/workload/bulkingest"
	workloadcli "github.com/cockroachdb/cockroach/pkg/workload/cli"
	_ "github.com/cockroachdb/cockroach/pkg/workload/examples"
	_ "github.com/cockroachdb/cockroach/pkg/workload/kv"
	_ "github.com/cockroachdb/cockroach/pkg/workload/movr"
	_ "github.com/cockroachdb/cockroach/pkg/workload/tpcc"
	_ "github.com/cockroachdb/cockroach/pkg/workload/tpch"
	_ "github.com/cockroachdb/cockroach/pkg/workload/ttllogger"
	_ "github.com/cockroachdb/cockroach/pkg/workload/ycsb"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

func Main() {
	__antithesis_instrumentation__.Notify(28060)

	rand.Seed(randutil.NewPseudoSeed())

	if len(os.Args) == 1 {
		__antithesis_instrumentation__.Notify(28063)
		os.Args = append(os.Args, "help")
	} else {
		__antithesis_instrumentation__.Notify(28064)
	}
	__antithesis_instrumentation__.Notify(28061)

	cmd, _, _ := cockroachCmd.Find(os.Args[1:])

	cmdName := commandName(cmd)

	err := doMain(cmd, cmdName)
	errCode := exit.Success()
	if err != nil {
		__antithesis_instrumentation__.Notify(28065)

		clierror.OutputError(stderr, err, true, false)

		fmt.Fprintf(stderr, "Failed running %q\n", cmdName)

		errCode = getExitCode(err)
	} else {
		__antithesis_instrumentation__.Notify(28066)
	}
	__antithesis_instrumentation__.Notify(28062)

	exit.WithCode(errCode)
}

func getExitCode(err error) (errCode exit.Code) {
	__antithesis_instrumentation__.Notify(28067)
	errCode = exit.UnspecifiedError()
	var cliErr *clierror.Error
	if errors.As(err, &cliErr) {
		__antithesis_instrumentation__.Notify(28069)
		errCode = cliErr.GetExitCode()
	} else {
		__antithesis_instrumentation__.Notify(28070)
	}
	__antithesis_instrumentation__.Notify(28068)
	return errCode
}

func doMain(cmd *cobra.Command, cmdName string) error {
	__antithesis_instrumentation__.Notify(28071)
	defer debugSignalSetup()()

	if cmd != nil {
		__antithesis_instrumentation__.Notify(28073)

		if err := processEnvVarDefaults(cmd); err != nil {
			__antithesis_instrumentation__.Notify(28075)
			return err
		} else {
			__antithesis_instrumentation__.Notify(28076)
		}
		__antithesis_instrumentation__.Notify(28074)

		if !cmdHasCustomLoggingSetup(cmd) {
			__antithesis_instrumentation__.Notify(28077)

			wrapped := cmd.PreRunE
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				__antithesis_instrumentation__.Notify(28078)

				err := setupLogging(context.Background(), cmd,
					false, true)

				if wrapped != nil {
					__antithesis_instrumentation__.Notify(28080)
					if err := wrapped(cmd, args); err != nil {
						__antithesis_instrumentation__.Notify(28081)
						return err
					} else {
						__antithesis_instrumentation__.Notify(28082)
					}
				} else {
					__antithesis_instrumentation__.Notify(28083)
				}
				__antithesis_instrumentation__.Notify(28079)

				return err
			}
		} else {
			__antithesis_instrumentation__.Notify(28084)
		}
	} else {
		__antithesis_instrumentation__.Notify(28085)
	}
	__antithesis_instrumentation__.Notify(28072)

	logcrash.SetupCrashReporter(
		context.Background(),
		cmdName,
	)

	defer logcrash.RecoverAndReportPanic(context.Background(), &serverCfg.Settings.SV)

	return Run(os.Args[1:])
}

func cmdHasCustomLoggingSetup(thisCmd *cobra.Command) bool {
	__antithesis_instrumentation__.Notify(28086)
	if thisCmd == nil {
		__antithesis_instrumentation__.Notify(28090)
		return false
	} else {
		__antithesis_instrumentation__.Notify(28091)
	}
	__antithesis_instrumentation__.Notify(28087)
	for _, cmd := range customLoggingSetupCmds {
		__antithesis_instrumentation__.Notify(28092)
		if cmd == thisCmd {
			__antithesis_instrumentation__.Notify(28093)
			return true
		} else {
			__antithesis_instrumentation__.Notify(28094)
		}
	}
	__antithesis_instrumentation__.Notify(28088)
	hasCustomLogging := false
	thisCmd.VisitParents(func(parent *cobra.Command) {
		__antithesis_instrumentation__.Notify(28095)
		for _, cmd := range customLoggingSetupCmds {
			__antithesis_instrumentation__.Notify(28096)
			if cmd == parent {
				__antithesis_instrumentation__.Notify(28097)
				hasCustomLogging = true
			} else {
				__antithesis_instrumentation__.Notify(28098)
			}
		}
	})
	__antithesis_instrumentation__.Notify(28089)
	return hasCustomLogging
}

func commandName(cmd *cobra.Command) string {
	__antithesis_instrumentation__.Notify(28099)
	rootName := cockroachCmd.CommandPath()
	if cmd != nil {
		__antithesis_instrumentation__.Notify(28101)
		return strings.TrimPrefix(cmd.CommandPath(), rootName+" ")
	} else {
		__antithesis_instrumentation__.Notify(28102)
	}
	__antithesis_instrumentation__.Notify(28100)
	return rootName
}

var stderr = log.OrigStderr

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "output version information",
	Long: `
Output build version information.
`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(28103)
		if cliCtx.showVersionUsingOnlyBuildTag {
			__antithesis_instrumentation__.Notify(28105)
			info := build.GetInfo()
			fmt.Println(info.Tag)
		} else {
			__antithesis_instrumentation__.Notify(28106)
			fmt.Println(fullVersionString())
		}
		__antithesis_instrumentation__.Notify(28104)
		return nil
	},
}

func fullVersionString() string {
	__antithesis_instrumentation__.Notify(28107)
	info := build.GetInfo()
	return info.Long()
}

var cockroachCmd = &cobra.Command{
	Use:   "cockroach [command] (flags)",
	Short: "CockroachDB command-line interface and server",

	Long: `CockroachDB command-line interface and server.`,

	SilenceUsage: true,

	SilenceErrors: true,

	Version: "details:\n" + fullVersionString() +
		"\n(use '" + os.Args[0] + " version --build-tag' to display only the build tag)",

	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var workloadCmd = workloadcli.WorkloadCmd(true)

func init() {
	cobra.EnableCommandSorting = false

	cockroachCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		if err := c.Usage(); err != nil {
			return err
		}
		fmt.Fprintln(c.OutOrStderr())
		return clierror.NewError(err, exit.CommandLineFlagError())
	})

	cockroachCmd.AddCommand(
		startCmd,
		startSingleNodeCmd,
		connectCmd,
		initCmd,
		certCmd,
		quitCmd,

		sqlShellCmd,
		stmtDiagCmd,
		authCmd,
		nodeCmd,
		nodeLocalCmd,
		userFileCmd,
		importCmd,

		demoCmd,
		convertURLCmd,
		genCmd,
		versionCmd,
		DebugCmd,
		sqlfmtCmd,
		workloadCmd,
	)
}

func isWorkloadCmd(cmd *cobra.Command) bool {
	__antithesis_instrumentation__.Notify(28108)
	return hasParentCmd(cmd, workloadCmd)
}

func isDemoCmd(cmd *cobra.Command) bool {
	__antithesis_instrumentation__.Notify(28109)
	return hasParentCmd(cmd, demoCmd) || func() bool {
		__antithesis_instrumentation__.Notify(28110)
		return hasParentCmd(cmd, debugStatementBundleCmd) == true
	}() == true
}

func hasParentCmd(cmd, refParent *cobra.Command) bool {
	__antithesis_instrumentation__.Notify(28111)
	if cmd == refParent {
		__antithesis_instrumentation__.Notify(28114)
		return true
	} else {
		__antithesis_instrumentation__.Notify(28115)
	}
	__antithesis_instrumentation__.Notify(28112)
	hasParent := false
	cmd.VisitParents(func(thisParent *cobra.Command) {
		__antithesis_instrumentation__.Notify(28116)
		if thisParent == refParent {
			__antithesis_instrumentation__.Notify(28117)
			hasParent = true
		} else {
			__antithesis_instrumentation__.Notify(28118)
		}
	})
	__antithesis_instrumentation__.Notify(28113)
	return hasParent
}

func Run(args []string) error {
	__antithesis_instrumentation__.Notify(28119)
	cockroachCmd.SetArgs(args)
	return cockroachCmd.Execute()
}

func UsageAndErr(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(28120)
	if err := cmd.Usage(); err != nil {
		__antithesis_instrumentation__.Notify(28122)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28123)
	}
	__antithesis_instrumentation__.Notify(28121)
	return fmt.Errorf("unknown sub-command: %q", strings.Join(args, " "))
}

func debugSignalSetup() func() {
	__antithesis_instrumentation__.Notify(28124)
	exit := make(chan struct{})
	ctx := context.Background()

	if quitSignal != nil {
		__antithesis_instrumentation__.Notify(28127)
		quitSignalCh := make(chan os.Signal, 1)
		signal.Notify(quitSignalCh, quitSignal)
		go func() {
			__antithesis_instrumentation__.Notify(28128)
			for {
				__antithesis_instrumentation__.Notify(28129)
				select {
				case <-exit:
					__antithesis_instrumentation__.Notify(28130)
					return
				case <-quitSignalCh:
					__antithesis_instrumentation__.Notify(28131)
					log.DumpStacks(ctx, "SIGQUIT received")
				}
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(28132)
	}
	__antithesis_instrumentation__.Notify(28125)

	if debugSignal != nil {
		__antithesis_instrumentation__.Notify(28133)
		debugSignalCh := make(chan os.Signal, 1)
		signal.Notify(debugSignalCh, debugSignal)
		go func() {
			__antithesis_instrumentation__.Notify(28134)
			for {
				__antithesis_instrumentation__.Notify(28135)
				select {
				case <-exit:
					__antithesis_instrumentation__.Notify(28136)
					return
				case <-debugSignalCh:
					__antithesis_instrumentation__.Notify(28137)
					log.Shout(ctx, severity.INFO, "setting up localhost debugging endpoint...")
					mux := http.NewServeMux()
					mux.HandleFunc("/debug/pprof/", pprof.Index)
					mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
					mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
					mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
					mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

					listenAddr := "localhost:0"
					listener, err := net.Listen("tcp", listenAddr)
					if err != nil {
						__antithesis_instrumentation__.Notify(28140)
						log.Shoutf(ctx, severity.WARNING, "debug server could not start listening on %s: %v", listenAddr, err)
						continue
					} else {
						__antithesis_instrumentation__.Notify(28141)
					}
					__antithesis_instrumentation__.Notify(28138)

					server := http.Server{Handler: mux}
					go func() {
						__antithesis_instrumentation__.Notify(28142)
						if err := server.Serve(listener); err != nil {
							__antithesis_instrumentation__.Notify(28143)
							log.Warningf(ctx, "debug server: %v", err)
						} else {
							__antithesis_instrumentation__.Notify(28144)
						}
					}()
					__antithesis_instrumentation__.Notify(28139)
					log.Shoutf(ctx, severity.INFO, "debug server listening on %s", listener.Addr())
					<-exit
					if err := server.Shutdown(ctx); err != nil {
						__antithesis_instrumentation__.Notify(28145)
						log.Warningf(ctx, "error shutting down debug server: %s", err)
					} else {
						__antithesis_instrumentation__.Notify(28146)
					}
				}
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(28147)
	}
	__antithesis_instrumentation__.Notify(28126)
	return func() {
		__antithesis_instrumentation__.Notify(28148)
		close(exit)
	}
}
