package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logconfig"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/cockroach/pkg/workload/workloadsql"
	"github.com/cockroachdb/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/time/rate"
)

var runFlags = pflag.NewFlagSet(`run`, pflag.ContinueOnError)
var tolerateErrors = runFlags.Bool("tolerate-errors", false, "Keep running on error")
var maxRate = runFlags.Float64(
	"max-rate", 0, "Maximum frequency of operations (reads/writes). If 0, no limit.")
var maxOps = runFlags.Uint64("max-ops", 0, "Maximum number of operations to run")
var countErrors = runFlags.Bool("count-errors", false, "If true, unsuccessful operations count towards --max-ops limit.")
var duration = runFlags.Duration("duration", 0,
	"The duration to run (in addition to --ramp). If 0, run forever.")
var doInit = runFlags.Bool("init", false, "Automatically run init. DEPRECATED: Use workload init instead.")
var ramp = runFlags.Duration("ramp", 0*time.Second, "The duration over which to ramp up load.")

var initFlags = pflag.NewFlagSet(`init`, pflag.ContinueOnError)
var drop = initFlags.Bool("drop", false, "Drop the existing database, if it exists")

var sharedFlags = pflag.NewFlagSet(`shared`, pflag.ContinueOnError)
var pprofport = sharedFlags.Int("pprofport", 33333, "Port for pprof endpoint.")
var dataLoader = sharedFlags.String("data-loader", `INSERT`,
	"How to load initial table data. Options are INSERT and IMPORT")
var initConns = sharedFlags.Int("init-conns", 16,
	"The number of connections to use during INSERT init")

var displayEvery = runFlags.Duration("display-every", time.Second, "How much time between every one-line activity reports.")

var displayFormat = runFlags.String("display-format", "simple", "Output display format (simple, incremental-json)")

var prometheusPort = sharedFlags.Int(
	"prometheus-port",
	2112,
	"Port to expose prometheus metrics if the workload has a prometheus gatherer set.",
)

var histograms = runFlags.String(
	"histograms", "",
	"File to write per-op incremental and cumulative histogram data.")
var histogramsMaxLatency = runFlags.Duration(
	"histograms-max-latency", 100*time.Second,
	"Expected maximum latency of running a query")

var securityFlags = pflag.NewFlagSet(`security`, pflag.ContinueOnError)
var secure = securityFlags.Bool("secure", false,
	"Run in secure mode (sslmode=require). "+
		"Running in secure mode expects the relevant certs to have been created for the user in the certs/ directory."+
		"For example when using root, certs/client.root.crt certs/client.root.key should exist.")
var user = securityFlags.String("user", "root", "Specify a user to run the workload as")
var password = securityFlags.String("password", "", "Optionally specify a password for the user")

func init() {

	_ = sharedFlags.MarkHidden("pprofport")

	AddSubCmd(func(userFacing bool) *cobra.Command {
		var initCmd = SetCmdDefaults(&cobra.Command{
			Use:   `init`,
			Short: `set up tables for a workload`,
		})
		for _, meta := range workload.Registered() {
			gen := meta.New()
			var genFlags *pflag.FlagSet
			if f, ok := gen.(workload.Flagser); ok {
				genFlags = f.Flags().FlagSet
			}

			genInitCmd := SetCmdDefaults(&cobra.Command{
				Use:   meta.Name + " [pgurl...]",
				Short: meta.Description,
				Long:  meta.Description + meta.Details,
				Args:  cobra.ArbitraryArgs,
			})
			genInitCmd.Flags().AddFlagSet(initFlags)
			genInitCmd.Flags().AddFlagSet(sharedFlags)
			genInitCmd.Flags().AddFlagSet(genFlags)
			genInitCmd.Flags().AddFlagSet(securityFlags)
			genInitCmd.Run = CmdHelper(gen, runInit)
			if userFacing && !meta.PublicFacing {
				genInitCmd.Hidden = true
			}
			initCmd.AddCommand(genInitCmd)
		}
		return initCmd
	})
	AddSubCmd(func(userFacing bool) *cobra.Command {
		var runCmd = SetCmdDefaults(&cobra.Command{
			Use:   `run`,
			Short: `run a workload's operations against a cluster`,
		})
		for _, meta := range workload.Registered() {
			gen := meta.New()
			if _, ok := gen.(workload.Opser); !ok {

				continue
			}

			var genFlags *pflag.FlagSet
			if f, ok := gen.(workload.Flagser); ok {
				genFlags = f.Flags().FlagSet
			}

			genRunCmd := SetCmdDefaults(&cobra.Command{
				Use:   meta.Name + " [pgurl...]",
				Short: meta.Description,
				Long:  meta.Description + meta.Details,
				Args:  cobra.ArbitraryArgs,
			})
			genRunCmd.Flags().AddFlagSet(runFlags)
			genRunCmd.Flags().AddFlagSet(sharedFlags)
			genRunCmd.Flags().AddFlagSet(genFlags)
			genRunCmd.Flags().AddFlagSet(securityFlags)
			initFlags.VisitAll(func(initFlag *pflag.Flag) {

				f := *initFlag
				f.Usage += ` (implies --init)`
				genRunCmd.Flags().AddFlag(&f)
			})
			genRunCmd.Run = CmdHelper(gen, runRun)
			if userFacing && !meta.PublicFacing {
				genRunCmd.Hidden = true
			}
			runCmd.AddCommand(genRunCmd)
		}
		return runCmd
	})
}

func CmdHelper(
	gen workload.Generator, fn func(gen workload.Generator, urls []string, dbName string) error,
) func(*cobra.Command, []string) {
	__antithesis_instrumentation__.Notify(693690)
	return HandleErrs(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(693691)

		if active, _ := log.IsActive(); !active {
			__antithesis_instrumentation__.Notify(693697)
			cfg := logconfig.DefaultStderrConfig()
			if err := cfg.Validate(nil); err != nil {
				__antithesis_instrumentation__.Notify(693699)
				return err
			} else {
				__antithesis_instrumentation__.Notify(693700)
			}
			__antithesis_instrumentation__.Notify(693698)
			if _, err := log.ApplyConfig(cfg); err != nil {
				__antithesis_instrumentation__.Notify(693701)
				return err
			} else {
				__antithesis_instrumentation__.Notify(693702)
			}
		} else {
			__antithesis_instrumentation__.Notify(693703)
		}
		__antithesis_instrumentation__.Notify(693692)

		if h, ok := gen.(workload.Hookser); ok {
			__antithesis_instrumentation__.Notify(693704)
			if h.Hooks().Validate != nil {
				__antithesis_instrumentation__.Notify(693705)
				if err := h.Hooks().Validate(); err != nil {
					__antithesis_instrumentation__.Notify(693706)
					return errors.Wrapf(err, "could not validate")
				} else {
					__antithesis_instrumentation__.Notify(693707)
				}
			} else {
				__antithesis_instrumentation__.Notify(693708)
			}
		} else {
			__antithesis_instrumentation__.Notify(693709)
		}
		__antithesis_instrumentation__.Notify(693693)

		var dbOverride string
		if dbFlag := cmd.Flag(`db`); dbFlag != nil {
			__antithesis_instrumentation__.Notify(693710)
			dbOverride = dbFlag.Value.String()
		} else {
			__antithesis_instrumentation__.Notify(693711)
		}
		__antithesis_instrumentation__.Notify(693694)

		urls := args
		if len(urls) == 0 {
			__antithesis_instrumentation__.Notify(693712)
			crdbDefaultURL := fmt.Sprintf(`postgres://%s@localhost:26257?sslmode=disable`, *user)
			if *secure {
				__antithesis_instrumentation__.Notify(693714)
				if *password != "" {
					__antithesis_instrumentation__.Notify(693715)
					crdbDefaultURL = fmt.Sprintf(
						`postgres://%s:%s@localhost:26257?sslmode=require&sslrootcert=certs/ca.crt`,
						*user, *password)
				} else {
					__antithesis_instrumentation__.Notify(693716)
					crdbDefaultURL = fmt.Sprintf(

						`postgres://%s@localhost:26257?sslcert=certs/client.%s.crt&sslkey=certs/client.%s.key&sslrootcert=certs/ca.crt&sslmode=require`,
						*user, *user, *user)
				}
			} else {
				__antithesis_instrumentation__.Notify(693717)
			}
			__antithesis_instrumentation__.Notify(693713)

			urls = []string{crdbDefaultURL}
		} else {
			__antithesis_instrumentation__.Notify(693718)
		}
		__antithesis_instrumentation__.Notify(693695)
		dbName, err := workload.SanitizeUrls(gen, dbOverride, urls)
		if err != nil {
			__antithesis_instrumentation__.Notify(693719)
			return err
		} else {
			__antithesis_instrumentation__.Notify(693720)
		}
		__antithesis_instrumentation__.Notify(693696)
		return fn(gen, urls, dbName)
	})
}

func SetCmdDefaults(cmd *cobra.Command) *cobra.Command {
	__antithesis_instrumentation__.Notify(693721)
	if cmd.Run == nil && func() bool {
		__antithesis_instrumentation__.Notify(693724)
		return cmd.RunE == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(693725)
		cmd.Run = func(cmd *cobra.Command, args []string) {
			__antithesis_instrumentation__.Notify(693726)
			_ = cmd.Usage()
		}
	} else {
		__antithesis_instrumentation__.Notify(693727)
	}
	__antithesis_instrumentation__.Notify(693722)
	if cmd.Args == nil {
		__antithesis_instrumentation__.Notify(693728)
		cmd.Args = cobra.NoArgs
	} else {
		__antithesis_instrumentation__.Notify(693729)
	}
	__antithesis_instrumentation__.Notify(693723)
	return cmd
}

var numOps uint64

func workerRun(
	ctx context.Context,
	errCh chan<- error,
	wg *sync.WaitGroup,
	limiter *rate.Limiter,
	workFn func(context.Context) error,
) {
	__antithesis_instrumentation__.Notify(693730)
	if wg != nil {
		__antithesis_instrumentation__.Notify(693732)
		defer wg.Done()
	} else {
		__antithesis_instrumentation__.Notify(693733)
	}
	__antithesis_instrumentation__.Notify(693731)

	for {
		__antithesis_instrumentation__.Notify(693734)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(693738)
			return
		} else {
			__antithesis_instrumentation__.Notify(693739)
		}
		__antithesis_instrumentation__.Notify(693735)

		if limiter != nil {
			__antithesis_instrumentation__.Notify(693740)
			if err := limiter.Wait(ctx); err != nil {
				__antithesis_instrumentation__.Notify(693741)
				return
			} else {
				__antithesis_instrumentation__.Notify(693742)
			}
		} else {
			__antithesis_instrumentation__.Notify(693743)
		}
		__antithesis_instrumentation__.Notify(693736)

		if err := workFn(ctx); err != nil {
			__antithesis_instrumentation__.Notify(693744)
			if ctx.Err() != nil && func() bool {
				__antithesis_instrumentation__.Notify(693746)
				return (errors.Is(err, ctx.Err()) || func() bool {
					__antithesis_instrumentation__.Notify(693747)
					return errors.Is(err, driver.ErrBadConn) == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(693748)

				return
			} else {
				__antithesis_instrumentation__.Notify(693749)
			}
			__antithesis_instrumentation__.Notify(693745)
			errCh <- err
			if !*countErrors {
				__antithesis_instrumentation__.Notify(693750)

				continue
			} else {
				__antithesis_instrumentation__.Notify(693751)
			}
		} else {
			__antithesis_instrumentation__.Notify(693752)
		}
		__antithesis_instrumentation__.Notify(693737)

		v := atomic.AddUint64(&numOps, 1)
		if *maxOps > 0 && func() bool {
			__antithesis_instrumentation__.Notify(693753)
			return v >= *maxOps == true
		}() == true {
			__antithesis_instrumentation__.Notify(693754)
			return
		} else {
			__antithesis_instrumentation__.Notify(693755)
		}
	}
}

func runInit(gen workload.Generator, urls []string, dbName string) error {
	__antithesis_instrumentation__.Notify(693756)
	ctx := context.Background()

	initDB, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(693758)
		return err
	} else {
		__antithesis_instrumentation__.Notify(693759)
	}
	__antithesis_instrumentation__.Notify(693757)

	startPProfEndPoint(ctx)
	return runInitImpl(ctx, gen, initDB, dbName)
}

func runInitImpl(
	ctx context.Context, gen workload.Generator, initDB *gosql.DB, dbName string,
) error {
	__antithesis_instrumentation__.Notify(693760)
	if *drop {
		__antithesis_instrumentation__.Notify(693764)
		if _, err := initDB.ExecContext(ctx, `DROP DATABASE IF EXISTS `+dbName); err != nil {
			__antithesis_instrumentation__.Notify(693765)
			return err
		} else {
			__antithesis_instrumentation__.Notify(693766)
		}
	} else {
		__antithesis_instrumentation__.Notify(693767)
	}
	__antithesis_instrumentation__.Notify(693761)
	if _, err := initDB.ExecContext(ctx, `CREATE DATABASE IF NOT EXISTS `+dbName); err != nil {
		__antithesis_instrumentation__.Notify(693768)
		return err
	} else {
		__antithesis_instrumentation__.Notify(693769)
	}
	__antithesis_instrumentation__.Notify(693762)

	var l workload.InitialDataLoader
	switch strings.ToLower(*dataLoader) {
	case `insert`, `inserts`:
		__antithesis_instrumentation__.Notify(693770)
		l = workloadsql.InsertsDataLoader{
			Concurrency: *initConns,
		}
	case `import`, `imports`:
		__antithesis_instrumentation__.Notify(693771)
		l = workload.ImportDataLoader
	default:
		__antithesis_instrumentation__.Notify(693772)
		return errors.Errorf(`unknown data loader: %s`, *dataLoader)
	}
	__antithesis_instrumentation__.Notify(693763)

	_, err := workloadsql.Setup(ctx, initDB, gen, l)
	return err
}

func startPProfEndPoint(ctx context.Context) {
	__antithesis_instrumentation__.Notify(693773)
	b := envutil.EnvOrDefaultInt64("COCKROACH_BLOCK_PROFILE_RATE",
		10000000)

	m := envutil.EnvOrDefaultInt("COCKROACH_MUTEX_PROFILE_RATE",
		1000)
	runtime.SetBlockProfileRate(int(b))
	runtime.SetMutexProfileFraction(m)

	go func() {
		__antithesis_instrumentation__.Notify(693774)
		err := http.ListenAndServe(":"+strconv.Itoa(*pprofport), nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(693775)
			log.Errorf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(693776)
		}
	}()
}

func runRun(gen workload.Generator, urls []string, dbName string) error {
	__antithesis_instrumentation__.Notify(693777)
	ctx := context.Background()

	var formatter outputFormat
	switch *displayFormat {
	case "simple":
		__antithesis_instrumentation__.Notify(693790)
		formatter = &textFormatter{}
	case "incremental-json":
		__antithesis_instrumentation__.Notify(693791)
		formatter = &jsonFormatter{w: os.Stdout}
	default:
		__antithesis_instrumentation__.Notify(693792)
		return errors.Errorf("unknown display format: %s", *displayFormat)
	}
	__antithesis_instrumentation__.Notify(693778)

	startPProfEndPoint(ctx)
	initDB, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(693793)
		return err
	} else {
		__antithesis_instrumentation__.Notify(693794)
	}
	__antithesis_instrumentation__.Notify(693779)
	if *doInit || func() bool {
		__antithesis_instrumentation__.Notify(693795)
		return *drop == true
	}() == true {
		__antithesis_instrumentation__.Notify(693796)
		log.Info(ctx, `DEPRECATION: `+
			`the --init flag on "workload run" will no longer be supported after 19.2`)
		for {
			__antithesis_instrumentation__.Notify(693797)
			err = runInitImpl(ctx, gen, initDB, dbName)
			if err == nil {
				__antithesis_instrumentation__.Notify(693800)
				break
			} else {
				__antithesis_instrumentation__.Notify(693801)
			}
			__antithesis_instrumentation__.Notify(693798)
			if !*tolerateErrors {
				__antithesis_instrumentation__.Notify(693802)
				return err
			} else {
				__antithesis_instrumentation__.Notify(693803)
			}
			__antithesis_instrumentation__.Notify(693799)
			log.Infof(ctx, "retrying after error during init: %v", err)
		}
	} else {
		__antithesis_instrumentation__.Notify(693804)
	}
	__antithesis_instrumentation__.Notify(693780)

	var limiter *rate.Limiter
	if *maxRate > 0 {
		__antithesis_instrumentation__.Notify(693805)

		limiter = rate.NewLimiter(rate.Limit(*maxRate), 1)
	} else {
		__antithesis_instrumentation__.Notify(693806)
	}
	__antithesis_instrumentation__.Notify(693781)

	o, ok := gen.(workload.Opser)
	if !ok {
		__antithesis_instrumentation__.Notify(693807)
		return errors.Errorf(`no operations defined for %s`, gen.Meta().Name)
	} else {
		__antithesis_instrumentation__.Notify(693808)
	}
	__antithesis_instrumentation__.Notify(693782)
	reg := histogram.NewRegistry(
		*histogramsMaxLatency,
		gen.Meta().Name,
	)

	go func() {
		__antithesis_instrumentation__.Notify(693809)
		if err := http.ListenAndServe(
			fmt.Sprintf(":%d", *prometheusPort),
			promhttp.HandlerFor(reg.Gatherer(), promhttp.HandlerOpts{}),
		); err != nil {
			__antithesis_instrumentation__.Notify(693810)
			log.Errorf(context.Background(), "error serving prometheus: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(693811)
		}
	}()
	__antithesis_instrumentation__.Notify(693783)

	var ops workload.QueryLoad
	prepareStart := timeutil.Now()
	log.Infof(ctx, "creating load generator...")
	const prepareTimeout = 60 * time.Minute
	prepareCtx, cancel := context.WithTimeout(ctx, prepareTimeout)
	defer cancel()
	if prepareErr := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(693812)
		retry := retry.StartWithCtx(ctx, retry.Options{})
		var err error
		for retry.Next() {
			__antithesis_instrumentation__.Notify(693815)
			if err != nil {
				__antithesis_instrumentation__.Notify(693818)
				log.Warningf(ctx, "retrying after error while creating load: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(693819)
			}
			__antithesis_instrumentation__.Notify(693816)
			ops, err = o.Ops(ctx, urls, reg)
			if err == nil {
				__antithesis_instrumentation__.Notify(693820)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(693821)
			}
			__antithesis_instrumentation__.Notify(693817)
			err = errors.Wrapf(err, "failed to initialize the load generator")
			if !*tolerateErrors {
				__antithesis_instrumentation__.Notify(693822)
				return err
			} else {
				__antithesis_instrumentation__.Notify(693823)
			}
		}
		__antithesis_instrumentation__.Notify(693813)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(693824)

			log.Errorf(ctx, "Attempt to create load generator failed. "+
				"It's been more than %s since we started trying to create the load generator "+
				"so we're giving up. Last failure: %s", prepareTimeout, err)
		} else {
			__antithesis_instrumentation__.Notify(693825)
		}
		__antithesis_instrumentation__.Notify(693814)
		return err
	}(prepareCtx); prepareErr != nil {
		__antithesis_instrumentation__.Notify(693826)
		return prepareErr
	} else {
		__antithesis_instrumentation__.Notify(693827)
	}
	__antithesis_instrumentation__.Notify(693784)
	log.Infof(ctx, "creating load generator... done (took %s)", timeutil.Since(prepareStart))

	start := timeutil.Now()
	errCh := make(chan error)
	var rampDone chan struct{}
	if *ramp > 0 {
		__antithesis_instrumentation__.Notify(693828)

		rampDone = make(chan struct{})
	} else {
		__antithesis_instrumentation__.Notify(693829)
	}
	__antithesis_instrumentation__.Notify(693785)

	workersCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()
	var wg sync.WaitGroup
	wg.Add(len(ops.WorkerFns))
	go func() {
		__antithesis_instrumentation__.Notify(693830)

		var rampCtx context.Context
		if rampDone != nil {
			__antithesis_instrumentation__.Notify(693833)
			var cancel func()
			rampCtx, cancel = context.WithTimeout(workersCtx, *ramp)
			defer cancel()
		} else {
			__antithesis_instrumentation__.Notify(693834)
		}
		__antithesis_instrumentation__.Notify(693831)

		for i, workFn := range ops.WorkerFns {
			__antithesis_instrumentation__.Notify(693835)
			go func(i int, workFn func(context.Context) error) {
				__antithesis_instrumentation__.Notify(693836)

				if rampCtx != nil {
					__antithesis_instrumentation__.Notify(693838)
					rampPerWorker := *ramp / time.Duration(len(ops.WorkerFns))
					time.Sleep(time.Duration(i) * rampPerWorker)
				} else {
					__antithesis_instrumentation__.Notify(693839)
				}
				__antithesis_instrumentation__.Notify(693837)
				workerRun(workersCtx, errCh, &wg, limiter, workFn)
			}(i, workFn)
		}
		__antithesis_instrumentation__.Notify(693832)

		if rampCtx != nil {
			__antithesis_instrumentation__.Notify(693840)

			<-rampCtx.Done()
			close(rampDone)
		} else {
			__antithesis_instrumentation__.Notify(693841)
		}
	}()
	__antithesis_instrumentation__.Notify(693786)

	ticker := time.NewTicker(*displayEvery)
	defer ticker.Stop()
	done := make(chan os.Signal, 3)
	signal.Notify(done, exitSignals...)

	go func() {
		__antithesis_instrumentation__.Notify(693842)
		wg.Wait()
		done <- os.Interrupt
	}()
	__antithesis_instrumentation__.Notify(693787)

	if *duration > 0 {
		__antithesis_instrumentation__.Notify(693843)
		go func() {
			__antithesis_instrumentation__.Notify(693844)
			time.Sleep(*duration + *ramp)
			done <- os.Interrupt
		}()
	} else {
		__antithesis_instrumentation__.Notify(693845)
	}
	__antithesis_instrumentation__.Notify(693788)

	var jsonEnc *json.Encoder
	if *histograms != "" {
		__antithesis_instrumentation__.Notify(693846)
		_ = os.MkdirAll(filepath.Dir(*histograms), 0755)
		jsonF, err := os.Create(*histograms)
		if err != nil {
			__antithesis_instrumentation__.Notify(693848)
			return err
		} else {
			__antithesis_instrumentation__.Notify(693849)
		}
		__antithesis_instrumentation__.Notify(693847)
		jsonEnc = json.NewEncoder(jsonF)
		defer func() {
			__antithesis_instrumentation__.Notify(693850)
			if err := jsonF.Sync(); err != nil {
				__antithesis_instrumentation__.Notify(693852)
				log.Warningf(ctx, "histogram: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(693853)
			}
			__antithesis_instrumentation__.Notify(693851)

			if err := jsonF.Close(); err != nil {
				__antithesis_instrumentation__.Notify(693854)
				log.Warningf(ctx, "histogram: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(693855)
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(693856)
	}
	__antithesis_instrumentation__.Notify(693789)

	everySecond := log.Every(*displayEvery)
	for {
		__antithesis_instrumentation__.Notify(693857)
		select {
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(693858)
			formatter.outputError(err)
			if *tolerateErrors {
				__antithesis_instrumentation__.Notify(693866)
				if everySecond.ShouldLog() {
					__antithesis_instrumentation__.Notify(693868)
					log.Errorf(ctx, "%v", err)
				} else {
					__antithesis_instrumentation__.Notify(693869)
				}
				__antithesis_instrumentation__.Notify(693867)
				continue
			} else {
				__antithesis_instrumentation__.Notify(693870)
			}
			__antithesis_instrumentation__.Notify(693859)
			return err

		case <-ticker.C:
			__antithesis_instrumentation__.Notify(693860)
			startElapsed := timeutil.Since(start)
			reg.Tick(func(t histogram.Tick) {
				__antithesis_instrumentation__.Notify(693871)
				formatter.outputTick(startElapsed, t)
				if jsonEnc != nil && func() bool {
					__antithesis_instrumentation__.Notify(693872)
					return rampDone == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(693873)
					if err := jsonEnc.Encode(t.Snapshot()); err != nil {
						__antithesis_instrumentation__.Notify(693874)
						log.Warningf(ctx, "histogram: %v", err)
					} else {
						__antithesis_instrumentation__.Notify(693875)
					}
				} else {
					__antithesis_instrumentation__.Notify(693876)
				}
			})

		case <-rampDone:
			__antithesis_instrumentation__.Notify(693861)
			rampDone = nil
			start = timeutil.Now()
			formatter.rampDone()
			reg.Tick(func(t histogram.Tick) {
				__antithesis_instrumentation__.Notify(693877)
				t.Cumulative.Reset()
				t.Hist.Reset()
			})

		case <-done:
			__antithesis_instrumentation__.Notify(693862)
			cancelWorkers()
			if ops.Close != nil {
				__antithesis_instrumentation__.Notify(693878)
				ops.Close(ctx)
			} else {
				__antithesis_instrumentation__.Notify(693879)
			}
			__antithesis_instrumentation__.Notify(693863)

			startElapsed := timeutil.Since(start)
			resultTick := histogram.Tick{Name: ops.ResultHist}
			reg.Tick(func(t histogram.Tick) {
				__antithesis_instrumentation__.Notify(693880)
				formatter.outputTotal(startElapsed, t)
				if jsonEnc != nil {
					__antithesis_instrumentation__.Notify(693882)

					if err := jsonEnc.Encode(t.Snapshot()); err != nil {
						__antithesis_instrumentation__.Notify(693883)
						log.Warningf(ctx, "histogram: %v", err)
					} else {
						__antithesis_instrumentation__.Notify(693884)
					}
				} else {
					__antithesis_instrumentation__.Notify(693885)
				}
				__antithesis_instrumentation__.Notify(693881)
				if ops.ResultHist == `` || func() bool {
					__antithesis_instrumentation__.Notify(693886)
					return ops.ResultHist == t.Name == true
				}() == true {
					__antithesis_instrumentation__.Notify(693887)
					if resultTick.Cumulative == nil {
						__antithesis_instrumentation__.Notify(693888)
						resultTick.Now = t.Now
						resultTick.Cumulative = t.Cumulative
					} else {
						__antithesis_instrumentation__.Notify(693889)
						resultTick.Cumulative.Merge(t.Cumulative)
					}
				} else {
					__antithesis_instrumentation__.Notify(693890)
				}
			})
			__antithesis_instrumentation__.Notify(693864)
			formatter.outputResult(startElapsed, resultTick)

			if h, ok := gen.(workload.Hookser); ok {
				__antithesis_instrumentation__.Notify(693891)
				if h.Hooks().PostRun != nil {
					__antithesis_instrumentation__.Notify(693892)
					if err := h.Hooks().PostRun(startElapsed); err != nil {
						__antithesis_instrumentation__.Notify(693893)
						fmt.Printf("failed post-run hook: %v\n", err)
					} else {
						__antithesis_instrumentation__.Notify(693894)
					}
				} else {
					__antithesis_instrumentation__.Notify(693895)
				}
			} else {
				__antithesis_instrumentation__.Notify(693896)
			}
			__antithesis_instrumentation__.Notify(693865)

			return nil
		}
	}
}
