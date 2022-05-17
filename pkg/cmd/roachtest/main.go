package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/tests"
	"github.com/cockroachdb/cockroach/pkg/roachprod"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const ExitCodeTestsFailed = 10

const ExitCodeClusterProvisioningFailed = 11

const runnerLogsDir = "_runner-logs"

func parseCreateOpts(flags *pflag.FlagSet, opts *vm.CreateOpts) {
	__antithesis_instrumentation__.Notify(44108)

	flags.DurationVar(&opts.Lifetime,
		"lifetime", opts.Lifetime, "Lifetime of the cluster")
	flags.BoolVar(&opts.SSDOpts.UseLocalSSD,
		"roachprod-local-ssd", opts.SSDOpts.UseLocalSSD, "Use local SSD")
	flags.StringVar(&opts.SSDOpts.FileSystem,
		"filesystem", opts.SSDOpts.FileSystem, "The underlying file system(ext4/zfs).")
	flags.BoolVar(&opts.SSDOpts.NoExt4Barrier,
		"local-ssd-no-ext4-barrier", opts.SSDOpts.NoExt4Barrier,
		`Mount the local SSD with the "-o nobarrier" flag. `+
			`Ignored if --local-ssd=false is specified.`)
	flags.IntVarP(&overrideNumNodes,
		"nodes", "n", -1, "Total number of nodes")
	flags.IntVarP(&opts.OsVolumeSize,
		"os-volume-size", "", opts.OsVolumeSize, "OS disk volume size in GB")
	flags.BoolVar(&opts.GeoDistributed,
		"geo", opts.GeoDistributed, "Create geo-distributed cluster")
}

func main() {
	__antithesis_instrumentation__.Notify(44109)
	rand.Seed(timeutil.Now().UnixNano())
	username := os.Getenv("ROACHPROD_USER")
	parallelism := 10
	var cpuQuota int

	var artifacts string

	var literalArtifacts string
	var httpPort int
	var debugEnabled bool
	var clusterID string
	var count = 1
	var versionsBinaryOverride map[string]string

	cobra.EnableCommandSorting = false

	var rootCmd = &cobra.Command{
		Use:   "roachtest [command] (flags)",
		Short: "roachtest tool for testing cockroach clusters",
		Long: `roachtest is a tool for testing cockroach clusters.
`,
		Version: "details:\n" + build.GetInfo().Long(),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			__antithesis_instrumentation__.Notify(44118)

			if cmd.Name() == "help" {
				__antithesis_instrumentation__.Notify(44122)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(44123)
			}
			__antithesis_instrumentation__.Notify(44119)

			if clusterName != "" && func() bool {
				__antithesis_instrumentation__.Notify(44124)
				return local == true
			}() == true {
				__antithesis_instrumentation__.Notify(44125)
				return fmt.Errorf(
					"cannot specify both an existing cluster (%s) and --local. However, if a local cluster "+
						"already exists, --clusters=local will use it",
					clusterName)
			} else {
				__antithesis_instrumentation__.Notify(44126)
			}
			__antithesis_instrumentation__.Notify(44120)
			switch cmd.Name() {
			case "run", "bench", "store-gen":
				__antithesis_instrumentation__.Notify(44127)
				initBinariesAndLibraries()
			default:
				__antithesis_instrumentation__.Notify(44128)
			}
			__antithesis_instrumentation__.Notify(44121)
			return nil
		},
	}
	__antithesis_instrumentation__.Notify(44110)

	rootCmd.PersistentFlags().StringVarP(
		&clusterName, "cluster", "c", "",
		"Comma-separated list of names existing cluster to use for running tests. "+
			"If fewer than --parallelism names are specified, then the parallelism "+
			"is capped to the number of clusters specified.")
	rootCmd.PersistentFlags().BoolVarP(
		&local, "local", "l", local, "run tests locally")
	rootCmd.PersistentFlags().StringVarP(
		&username, "user", "u", username,
		"Username to use as a cluster name prefix. "+
			"If blank, the current OS user is detected and specified.")
	rootCmd.PersistentFlags().StringVar(
		&cockroach, "cockroach", "", "path to cockroach binary to use")
	rootCmd.PersistentFlags().StringVar(
		&workload, "workload", "", "path to workload binary to use")
	f := rootCmd.PersistentFlags().VarPF(
		&encrypt, "encrypt", "", "start cluster with encryption at rest turned on")
	f.NoOptDefVal = "true"

	rootCmd.AddCommand(&cobra.Command{
		Use:   `version`,
		Short: `print version information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(44129)
			info := build.GetInfo()
			fmt.Println(info.Long())
			return nil
		}},
	)
	__antithesis_instrumentation__.Notify(44111)

	var listBench bool

	var listCmd = &cobra.Command{
		Use:   "list [tests]",
		Short: "list tests matching the patterns",
		Long: `List tests that match the given name patterns.

If no pattern is passed, all tests are matched.
Use --bench to list benchmarks instead of tests.

Each test has a set of tags. The tags are used to skip tests which don't match
the tag filter. The tag filter is specified by specifying a pattern with the
"tag:" prefix. The default tag filter is "tag:default" which matches any test
that has the "default" tag. Note that tests are selected based on their name,
and skipped based on their tag.

Examples:

   roachtest list acceptance copy/bank/.*false
   roachtest list tag:acceptance
   roachtest list tag:weekly
`,
		RunE: func(_ *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(44130)
			r, err := makeTestRegistry(cloud, instanceType, zonesF, localSSDArg)
			if err != nil {
				__antithesis_instrumentation__.Notify(44134)
				return err
			} else {
				__antithesis_instrumentation__.Notify(44135)
			}
			__antithesis_instrumentation__.Notify(44131)
			if !listBench {
				__antithesis_instrumentation__.Notify(44136)
				tests.RegisterTests(&r)
			} else {
				__antithesis_instrumentation__.Notify(44137)
				tests.RegisterBenchmarks(&r)
			}
			__antithesis_instrumentation__.Notify(44132)

			matchedTests := r.List(context.Background(), args)
			for _, test := range matchedTests {
				__antithesis_instrumentation__.Notify(44138)
				var skip string
				if test.Skip != "" {
					__antithesis_instrumentation__.Notify(44140)
					skip = " (skipped: " + test.Skip + ")"
				} else {
					__antithesis_instrumentation__.Notify(44141)
				}
				__antithesis_instrumentation__.Notify(44139)
				fmt.Printf("%s [%s]%s\n", test.Name, test.Owner, skip)
			}
			__antithesis_instrumentation__.Notify(44133)
			return nil
		},
	}
	__antithesis_instrumentation__.Notify(44112)
	listCmd.Flags().BoolVar(
		&listBench, "bench", false, "list benchmarks instead of tests")

	var runCmd = &cobra.Command{

		SilenceUsage: true,
		Use:          "run [tests]",
		Short:        "run automated tests on cockroach cluster",
		Long: `Run automated tests on existing or ephemeral cockroach clusters.

roachtest run takes a list of regex patterns and runs all the matching tests.
If no pattern is given, all tests are run. See "help list" for more details on
the test tags.

If all invoked tests passed, the exit status is zero. If at least one test
failed, it is 10. Any other exit status reports a problem with the test
runner itself.
`,
		RunE: func(_ *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(44142)
			if literalArtifacts == "" {
				__antithesis_instrumentation__.Notify(44144)
				literalArtifacts = artifacts
			} else {
				__antithesis_instrumentation__.Notify(44145)
			}
			__antithesis_instrumentation__.Notify(44143)
			return runTests(tests.RegisterTests, cliCfg{
				args:                   args,
				count:                  count,
				cpuQuota:               cpuQuota,
				debugEnabled:           debugEnabled,
				httpPort:               httpPort,
				parallelism:            parallelism,
				artifactsDir:           artifacts,
				literalArtifactsDir:    literalArtifacts,
				user:                   username,
				clusterID:              clusterID,
				versionsBinaryOverride: versionsBinaryOverride,
			})
		},
	}
	__antithesis_instrumentation__.Notify(44113)

	runCmd.Flags().StringVar(
		&buildTag, "build-tag", "", "build tag (auto-detect if empty)")
	runCmd.Flags().StringVar(
		&slackToken, "slack-token", "", "Slack bot token")
	runCmd.Flags().BoolVar(
		&teamCity, "teamcity", false, "include teamcity-specific markers in output")
	runCmd.Flags().BoolVar(
		&disableIssue, "disable-issue", false, "disable posting GitHub issue for failures")

	var benchCmd = &cobra.Command{

		SilenceUsage: true,
		Use:          "bench [benchmarks]",
		Short:        "run automated benchmarks on cockroach cluster",
		Long:         `Run automated benchmarks on existing or ephemeral cockroach clusters.`,
		RunE: func(_ *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(44146)
			if literalArtifacts == "" {
				__antithesis_instrumentation__.Notify(44148)
				literalArtifacts = artifacts
			} else {
				__antithesis_instrumentation__.Notify(44149)
			}
			__antithesis_instrumentation__.Notify(44147)
			return runTests(tests.RegisterBenchmarks, cliCfg{
				args:                   args,
				count:                  count,
				cpuQuota:               cpuQuota,
				debugEnabled:           debugEnabled,
				httpPort:               httpPort,
				parallelism:            parallelism,
				artifactsDir:           artifacts,
				user:                   username,
				clusterID:              clusterID,
				versionsBinaryOverride: versionsBinaryOverride,
			})
		},
	}
	__antithesis_instrumentation__.Notify(44114)

	for _, cmd := range []*cobra.Command{runCmd, benchCmd} {
		__antithesis_instrumentation__.Notify(44150)
		cmd.Flags().StringVar(
			&artifacts, "artifacts", "artifacts", "path to artifacts directory")
		cmd.Flags().StringVar(
			&literalArtifacts, "artifacts-literal", "", "literal path to on-agent artifacts directory. Used for messages to ##teamcity[publishArtifacts] in --teamcity mode. May be different from --artifacts; defaults to the value of --artifacts if not provided")
		cmd.Flags().StringVar(
			&cloud, "cloud", cloud, "cloud provider to use (aws, azure, or gce)")
		cmd.Flags().StringVar(
			&clusterID, "cluster-id", "", "an identifier to use in the test cluster's name")
		cmd.Flags().IntVar(
			&count, "count", 1, "the number of times to run each test")
		cmd.Flags().BoolVarP(
			&debugEnabled, "debug", "d", debugEnabled, "don't wipe and destroy cluster if test fails")
		cmd.Flags().IntVarP(
			&parallelism, "parallelism", "p", parallelism, "number of tests to run in parallel")
		cmd.Flags().StringVar(
			&deprecatedRoachprodBinary, "roachprod", "", "DEPRECATED")
		_ = cmd.Flags().MarkDeprecated("roachprod", "roachtest now uses roachprod as a library")
		cmd.Flags().BoolVar(
			&clusterWipe, "wipe", true,
			"wipe existing cluster before starting test (for use with --cluster)")
		cmd.Flags().StringVar(
			&zonesF, "zones", "",
			"Zones for the cluster. (non-geo tests use the first zone, geo tests use all zones) "+
				"(uses roachprod defaults if empty)")
		cmd.Flags().StringVar(
			&instanceType, "instance-type", instanceType,
			"the instance type to use (see https://aws.amazon.com/ec2/instance-types/, https://cloud.google.com/compute/docs/machine-types or https://docs.microsoft.com/en-us/azure/virtual-machines/windows/sizes)")
		cmd.Flags().IntVar(
			&cpuQuota, "cpu-quota", 300,
			"The number of cloud CPUs roachtest is allowed to use at any one time.")
		cmd.Flags().IntVar(
			&httpPort, "port", 8080, "the port on which to serve the HTTP interface")
		cmd.Flags().BoolVar(
			&localSSDArg, "local-ssd", true, "Use a local SSD instead of an EBS volume (only for use with AWS) (defaults to true if instance type supports local SSDs)")
		cmd.Flags().StringToStringVar(
			&versionsBinaryOverride, "versions-binary-override", nil,
			"List of <version>=<path to cockroach binary>. If a certain version <ver> "+
				"is present in the list,"+"the respective binary will be used when a "+
				"multi-version test asks for the respective binary, instead of "+
				"`roachprod stage <ver>`. Example: 20.1.4=cockroach-20.1,20.2.0=cockroach-20.2.")
	}
	__antithesis_instrumentation__.Notify(44115)

	parseCreateOpts(runCmd.Flags(), &overrideOpts)
	overrideFlagset = runCmd.Flags()

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(benchCmd)

	var err error
	config.OSUser, err = user.Current()
	if err != nil {
		__antithesis_instrumentation__.Notify(44151)
		fmt.Fprintf(os.Stderr, "unable to lookup current user: %s\n", err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(44152)
	}
	__antithesis_instrumentation__.Notify(44116)

	if err := roachprod.InitDirs(); err != nil {
		__antithesis_instrumentation__.Notify(44153)
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(44154)
	}
	__antithesis_instrumentation__.Notify(44117)

	if err := rootCmd.Execute(); err != nil {
		__antithesis_instrumentation__.Notify(44155)
		code := 1
		if errors.Is(err, errTestsFailed) {
			__antithesis_instrumentation__.Notify(44158)
			code = ExitCodeTestsFailed
		} else {
			__antithesis_instrumentation__.Notify(44159)
		}
		__antithesis_instrumentation__.Notify(44156)
		if errors.Is(err, errClusterProvisioningFailed) {
			__antithesis_instrumentation__.Notify(44160)
			code = ExitCodeClusterProvisioningFailed
		} else {
			__antithesis_instrumentation__.Notify(44161)
		}
		__antithesis_instrumentation__.Notify(44157)

		os.Exit(code)
	} else {
		__antithesis_instrumentation__.Notify(44162)
	}
}

type cliCfg struct {
	args                   []string
	count                  int
	cpuQuota               int
	debugEnabled           bool
	httpPort               int
	parallelism            int
	artifactsDir           string
	literalArtifactsDir    string
	user                   string
	clusterID              string
	versionsBinaryOverride map[string]string
}

func runTests(register func(registry.Registry), cfg cliCfg) error {
	__antithesis_instrumentation__.Notify(44163)
	if cfg.count <= 0 {
		__antithesis_instrumentation__.Notify(44170)
		return fmt.Errorf("--count (%d) must by greater than 0", cfg.count)
	} else {
		__antithesis_instrumentation__.Notify(44171)
	}
	__antithesis_instrumentation__.Notify(44164)
	r, err := makeTestRegistry(cloud, instanceType, zonesF, localSSDArg)
	if err != nil {
		__antithesis_instrumentation__.Notify(44172)
		return err
	} else {
		__antithesis_instrumentation__.Notify(44173)
	}
	__antithesis_instrumentation__.Notify(44165)
	register(&r)
	cr := newClusterRegistry()
	stopper := stop.NewStopper()
	defer stopper.Stop(context.Background())
	runner := newTestRunner(cr, stopper, r.buildVersion)

	filter := registry.NewTestFilter(cfg.args)
	clusterType := roachprodCluster
	if local {
		__antithesis_instrumentation__.Notify(44174)
		clusterType = localCluster
		if cfg.parallelism != 1 {
			__antithesis_instrumentation__.Notify(44175)
			fmt.Printf("--local specified. Overriding --parallelism to 1.\n")
			cfg.parallelism = 1
		} else {
			__antithesis_instrumentation__.Notify(44176)
		}
	} else {
		__antithesis_instrumentation__.Notify(44177)
	}
	__antithesis_instrumentation__.Notify(44166)

	opt := clustersOpt{
		typ:                       clusterType,
		clusterName:               clusterName,
		user:                      getUser(cfg.user),
		cpuQuota:                  cfg.cpuQuota,
		keepClustersOnTestFailure: cfg.debugEnabled,
		clusterID:                 cfg.clusterID,
	}
	if err := runner.runHTTPServer(cfg.httpPort, os.Stdout); err != nil {
		__antithesis_instrumentation__.Notify(44178)
		return err
	} else {
		__antithesis_instrumentation__.Notify(44179)
	}
	__antithesis_instrumentation__.Notify(44167)

	tests := testsToRun(context.Background(), r, filter)
	n := len(tests)
	if n*cfg.count < cfg.parallelism {
		__antithesis_instrumentation__.Notify(44180)

		cfg.parallelism = n * cfg.count
	} else {
		__antithesis_instrumentation__.Notify(44181)
	}
	__antithesis_instrumentation__.Notify(44168)
	runnerDir := filepath.Join(cfg.artifactsDir, runnerLogsDir)
	runnerLogPath := filepath.Join(
		runnerDir, fmt.Sprintf("test_runner-%d.log", timeutil.Now().Unix()))
	l, tee := testRunnerLogger(context.Background(), cfg.parallelism, runnerLogPath)
	lopt := loggingOpt{
		l:                   l,
		tee:                 tee,
		stdout:              os.Stdout,
		stderr:              os.Stderr,
		artifactsDir:        cfg.artifactsDir,
		literalArtifactsDir: cfg.literalArtifactsDir,
		runnerLogPath:       runnerLogPath,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	CtrlC(ctx, l, cancel, cr)
	err = runner.Run(
		ctx, tests, cfg.count, cfg.parallelism, opt,
		testOpts{versionsBinaryOverride: cfg.versionsBinaryOverride},
		lopt, nil)

	l.PrintfCtx(ctx, "runTests destroying all clusters")
	cr.destroyAllClusters(context.Background(), l)

	if teamCity {
		__antithesis_instrumentation__.Notify(44182)

		fmt.Printf("##teamcity[publishArtifacts '%s']\n", filepath.Join(cfg.literalArtifactsDir, runnerLogsDir))
	} else {
		__antithesis_instrumentation__.Notify(44183)
	}
	__antithesis_instrumentation__.Notify(44169)
	return err
}

func getUser(userFlag string) string {
	__antithesis_instrumentation__.Notify(44184)
	if userFlag != "" {
		__antithesis_instrumentation__.Notify(44187)
		return userFlag
	} else {
		__antithesis_instrumentation__.Notify(44188)
	}
	__antithesis_instrumentation__.Notify(44185)
	usr, err := user.Current()
	if err != nil {
		__antithesis_instrumentation__.Notify(44189)
		panic(fmt.Sprintf("user.Current: %s", err))
	} else {
		__antithesis_instrumentation__.Notify(44190)
	}
	__antithesis_instrumentation__.Notify(44186)
	return usr.Username
}

func CtrlC(ctx context.Context, l *logger.Logger, cancel func(), cr *clusterRegistry) {
	__antithesis_instrumentation__.Notify(44191)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		__antithesis_instrumentation__.Notify(44192)
		<-sig
		shout(ctx, l, os.Stderr,
			"Signaled received. Canceling workers and waiting up to 5s for them.")

		cancel()
		<-time.After(5 * time.Second)
		shout(ctx, l, os.Stderr, "5s elapsed. Will brutally destroy all clusters.")

		destroyCh := make(chan struct{})
		go func() {
			__antithesis_instrumentation__.Notify(44194)

			destroyCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			l.PrintfCtx(ctx, "CtrlC handler destroying all clusters")
			cr.destroyAllClusters(destroyCtx, l)
			cancel()
			close(destroyCh)
		}()
		__antithesis_instrumentation__.Notify(44193)

		select {
		case <-sig:
			__antithesis_instrumentation__.Notify(44195)
			shout(ctx, l, os.Stderr, "Second SIGINT received. Quitting. Cluster might be left behind.")
			os.Exit(2)
		case <-destroyCh:
			__antithesis_instrumentation__.Notify(44196)
			shout(ctx, l, os.Stderr, "Done destroying all clusters.")
			os.Exit(2)
		}
	}()
}

func testRunnerLogger(
	ctx context.Context, parallelism int, runnerLogPath string,
) (*logger.Logger, logger.TeeOptType) {
	__antithesis_instrumentation__.Notify(44197)
	teeOpt := logger.NoTee
	if parallelism == 1 {
		__antithesis_instrumentation__.Notify(44200)
		teeOpt = logger.TeeToStdout
	} else {
		__antithesis_instrumentation__.Notify(44201)
	}
	__antithesis_instrumentation__.Notify(44198)

	var l *logger.Logger
	if teeOpt == logger.TeeToStdout {
		__antithesis_instrumentation__.Notify(44202)
		verboseCfg := logger.Config{Stdout: os.Stdout, Stderr: os.Stderr}
		var err error
		l, err = verboseCfg.NewLogger(runnerLogPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(44203)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(44204)
		}
	} else {
		__antithesis_instrumentation__.Notify(44205)
		verboseCfg := logger.Config{}
		var err error
		l, err = verboseCfg.NewLogger(runnerLogPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(44206)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(44207)
		}
	}
	__antithesis_instrumentation__.Notify(44199)
	shout(ctx, l, os.Stdout, "test runner logs in: %s", runnerLogPath)
	return l, teeOpt
}

func testsToRun(
	ctx context.Context, r testRegistryImpl, filter *registry.TestFilter,
) []registry.TestSpec {
	__antithesis_instrumentation__.Notify(44208)
	tests := r.GetTests(ctx, filter)

	var notSkipped []registry.TestSpec
	for _, s := range tests {
		__antithesis_instrumentation__.Notify(44210)
		if s.Skip == "" {
			__antithesis_instrumentation__.Notify(44211)
			notSkipped = append(notSkipped, s)
		} else {
			__antithesis_instrumentation__.Notify(44212)
			if teamCity {
				__antithesis_instrumentation__.Notify(44214)
				fmt.Fprintf(os.Stdout, "##teamcity[testIgnored name='%s' message='%s']\n",
					s.Name, teamCityEscape(s.Skip))
			} else {
				__antithesis_instrumentation__.Notify(44215)
			}
			__antithesis_instrumentation__.Notify(44213)
			fmt.Fprintf(os.Stdout, "--- SKIP: %s (%s)\n\t%s\n", s.Name, "0.00s", s.Skip)
		}
	}
	__antithesis_instrumentation__.Notify(44209)
	return notSkipped
}
