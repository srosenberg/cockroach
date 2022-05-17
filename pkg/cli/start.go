package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/cgroups"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/cockroach/pkg/util/sdnotify"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/cockroachdb/redact"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var debugTSImportFile = envutil.EnvOrDefaultString("COCKROACH_DEBUG_TS_IMPORT_FILE", "")
var debugTSImportMappingFile = envutil.EnvOrDefaultString("COCKROACH_DEBUG_TS_IMPORT_MAPPING_FILE", "")

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start a node in a multi-node cluster",
	Long: `
Start a CockroachDB node, which will export data from one or more
storage devices, specified via --store flags.

Specify the --join flag to point to another node or nodes that are
part of the same cluster. The other nodes do not need to be started
yet, and if the address of the other nodes to be added are not yet
known it is legal for the first node to join itself.

To initialize the cluster, use 'cockroach init'.
`,
	Example: `  cockroach start --insecure --store=attrs=ssd,path=/mnt/ssd1 --join=host:port,[host:port]`,
	Args:    cobra.NoArgs,
	RunE:    clierrorplus.MaybeShoutError(clierrorplus.MaybeDecorateError(runStartJoin)),
}

var startSingleNodeCmd = &cobra.Command{
	Use:   "start-single-node",
	Short: "start a single-node cluster",
	Long: `
Start a CockroachDB node, which will export data from one or more
storage devices, specified via --store flags.
The cluster will also be automatically initialized with
replication disabled (replication factor = 1).
`,
	Example: `  cockroach start-single-node --insecure --store=attrs=ssd,path=/mnt/ssd1`,
	Args:    cobra.NoArgs,
	RunE:    clierrorplus.MaybeShoutError(clierrorplus.MaybeDecorateError(runStartSingleNode)),
}

var StartCmds = []*cobra.Command{startCmd, startSingleNodeCmd}

var serverCmds = append(StartCmds, mtStartSQLCmd)

var customLoggingSetupCmds = append(
	serverCmds, debugCheckLogConfigCmd, demoCmd, mtStartSQLProxyCmd, mtTestDirectorySvr, statementBundleRecreateCmd,
)

func initBlockProfile() {
	__antithesis_instrumentation__.Notify(33991)

	d := envutil.EnvOrDefaultInt64("COCKROACH_BLOCK_PROFILE_RATE", 0)
	runtime.SetBlockProfileRate(int(d))
}

func initMutexProfile() {
	__antithesis_instrumentation__.Notify(33992)

	d := envutil.EnvOrDefaultInt("COCKROACH_MUTEX_PROFILE_RATE",
		1000)
	runtime.SetMutexProfileFraction(d)
}

func initTraceDir(ctx context.Context, dir string) {
	__antithesis_instrumentation__.Notify(33993)
	if dir == "" {
		__antithesis_instrumentation__.Notify(33995)
		return
	} else {
		__antithesis_instrumentation__.Notify(33996)
	}
	__antithesis_instrumentation__.Notify(33994)
	if err := os.MkdirAll(dir, 0755); err != nil {
		__antithesis_instrumentation__.Notify(33997)

		log.Warningf(ctx, "cannot create trace dir; traces will not be dumped: %+v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(33998)
	}
}

var cacheSizeValue = newBytesOrPercentageValue(&serverCfg.CacheSize, memoryPercentResolver)
var sqlSizeValue = newBytesOrPercentageValue(&serverCfg.MemoryPoolSize, memoryPercentResolver)
var diskTempStorageSizeValue = newBytesOrPercentageValue(nil, nil)
var tsdbSizeValue = newBytesOrPercentageValue(&serverCfg.TimeSeriesServerConfig.QueryMemoryMax, memoryPercentResolver)

func initExternalIODir(ctx context.Context, firstStore base.StoreSpec) (string, error) {
	__antithesis_instrumentation__.Notify(33999)
	externalIODir := startCtx.externalIODir
	if externalIODir == "" && func() bool {
		__antithesis_instrumentation__.Notify(34003)
		return !firstStore.InMemory == true
	}() == true {
		__antithesis_instrumentation__.Notify(34004)
		externalIODir = filepath.Join(firstStore.Path, "extern")
	} else {
		__antithesis_instrumentation__.Notify(34005)
	}
	__antithesis_instrumentation__.Notify(34000)
	if externalIODir == "" || func() bool {
		__antithesis_instrumentation__.Notify(34006)
		return externalIODir == "disabled" == true
	}() == true {
		__antithesis_instrumentation__.Notify(34007)
		return "", nil
	} else {
		__antithesis_instrumentation__.Notify(34008)
	}
	__antithesis_instrumentation__.Notify(34001)
	if !filepath.IsAbs(externalIODir) {
		__antithesis_instrumentation__.Notify(34009)
		return "", errors.Errorf("%s path must be absolute", cliflags.ExternalIODir.Name)
	} else {
		__antithesis_instrumentation__.Notify(34010)
	}
	__antithesis_instrumentation__.Notify(34002)
	return externalIODir, nil
}

func initTempStorageConfig(
	ctx context.Context, st *cluster.Settings, stopper *stop.Stopper, stores base.StoreSpecList,
) (base.TempStorageConfig, error) {
	__antithesis_instrumentation__.Notify(34011)

	var specIdx = 0
	for i, spec := range stores.Specs {
		__antithesis_instrumentation__.Notify(34019)
		if spec.IsEncrypted() {
			__antithesis_instrumentation__.Notify(34022)

			specIdx = i
		} else {
			__antithesis_instrumentation__.Notify(34023)
		}
		__antithesis_instrumentation__.Notify(34020)
		if spec.InMemory {
			__antithesis_instrumentation__.Notify(34024)
			continue
		} else {
			__antithesis_instrumentation__.Notify(34025)
		}
		__antithesis_instrumentation__.Notify(34021)
		recordPath := filepath.Join(spec.Path, server.TempDirsRecordFilename)
		if err := fs.CleanupTempDirs(recordPath); err != nil {
			__antithesis_instrumentation__.Notify(34026)
			return base.TempStorageConfig{}, errors.Wrap(err,
				"could not cleanup temporary directories from record file")
		} else {
			__antithesis_instrumentation__.Notify(34027)
		}
	}
	__antithesis_instrumentation__.Notify(34012)

	useStore := stores.Specs[specIdx]

	var recordPath string
	if !useStore.InMemory {
		__antithesis_instrumentation__.Notify(34028)
		recordPath = filepath.Join(useStore.Path, server.TempDirsRecordFilename)
	} else {
		__antithesis_instrumentation__.Notify(34029)
	}
	__antithesis_instrumentation__.Notify(34013)

	var tempStorePercentageResolver percentResolverFunc
	if !useStore.InMemory {
		__antithesis_instrumentation__.Notify(34030)
		dir := useStore.Path

		if err := os.MkdirAll(dir, 0755); err != nil {
			__antithesis_instrumentation__.Notify(34032)
			return base.TempStorageConfig{}, errors.Wrapf(err, "failed to create dir for first store: %s", dir)
		} else {
			__antithesis_instrumentation__.Notify(34033)
		}
		__antithesis_instrumentation__.Notify(34031)
		var err error
		tempStorePercentageResolver, err = diskPercentResolverFactory(dir)
		if err != nil {
			__antithesis_instrumentation__.Notify(34034)
			return base.TempStorageConfig{}, errors.Wrapf(err, "failed to create resolver for: %s", dir)
		} else {
			__antithesis_instrumentation__.Notify(34035)
		}
	} else {
		__antithesis_instrumentation__.Notify(34036)
		tempStorePercentageResolver = memoryPercentResolver
	}
	__antithesis_instrumentation__.Notify(34014)
	var tempStorageMaxSizeBytes int64
	if err := diskTempStorageSizeValue.Resolve(
		&tempStorageMaxSizeBytes, tempStorePercentageResolver,
	); err != nil {
		__antithesis_instrumentation__.Notify(34037)
		return base.TempStorageConfig{}, err
	} else {
		__antithesis_instrumentation__.Notify(34038)
	}
	__antithesis_instrumentation__.Notify(34015)
	if !diskTempStorageSizeValue.IsSet() {
		__antithesis_instrumentation__.Notify(34039)

		if startCtx.tempDir == "" && func() bool {
			__antithesis_instrumentation__.Notify(34040)
			return useStore.InMemory == true
		}() == true {
			__antithesis_instrumentation__.Notify(34041)
			tempStorageMaxSizeBytes = base.DefaultInMemTempStorageMaxSizeBytes
		} else {
			__antithesis_instrumentation__.Notify(34042)
			tempStorageMaxSizeBytes = base.DefaultTempStorageMaxSizeBytes
		}
	} else {
		__antithesis_instrumentation__.Notify(34043)
	}
	__antithesis_instrumentation__.Notify(34016)

	tempStorageConfig := base.TempStorageConfigFromEnv(
		ctx,
		st,
		useStore,
		startCtx.tempDir,
		tempStorageMaxSizeBytes,
	)

	tempDir := startCtx.tempDir
	if tempDir == "" && func() bool {
		__antithesis_instrumentation__.Notify(34044)
		return !tempStorageConfig.InMemory == true
	}() == true {
		__antithesis_instrumentation__.Notify(34045)
		tempDir = useStore.Path
	} else {
		__antithesis_instrumentation__.Notify(34046)
	}

	{
		__antithesis_instrumentation__.Notify(34047)
		var err error
		if tempStorageConfig.Path, err = fs.CreateTempDir(tempDir, server.TempDirPrefix, stopper); err != nil {
			__antithesis_instrumentation__.Notify(34048)
			return base.TempStorageConfig{}, errors.Wrap(err, "could not create temporary directory for temp storage")
		} else {
			__antithesis_instrumentation__.Notify(34049)
		}
	}
	__antithesis_instrumentation__.Notify(34017)

	if recordPath != "" {
		__antithesis_instrumentation__.Notify(34050)
		if err := fs.RecordTempDir(recordPath, tempStorageConfig.Path); err != nil {
			__antithesis_instrumentation__.Notify(34051)
			return base.TempStorageConfig{}, errors.Wrapf(
				err,
				"could not record temporary directory path to record file: %s",
				recordPath,
			)
		} else {
			__antithesis_instrumentation__.Notify(34052)
		}
	} else {
		__antithesis_instrumentation__.Notify(34053)
	}
	__antithesis_instrumentation__.Notify(34018)

	return tempStorageConfig, nil
}

var errCannotUseJoin = errors.New("cannot use --join with 'cockroach start-single-node' -- use 'cockroach start' instead")

func runStartSingleNode(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(34054)
	joinFlag := flagSetForCmd(cmd).Lookup(cliflags.Join.Name)
	if joinFlag.Changed {
		__antithesis_instrumentation__.Notify(34057)
		return errCannotUseJoin
	} else {
		__antithesis_instrumentation__.Notify(34058)
	}
	__antithesis_instrumentation__.Notify(34055)

	joinFlag.Changed = true

	serverCfg.AutoInitializeCluster = true

	if debugTSImportFile != "" {
		__antithesis_instrumentation__.Notify(34059)
		serverCfg.TestingKnobs.Server = &server.TestingKnobs{
			ImportTimeseriesFile:        debugTSImportFile,
			ImportTimeseriesMappingFile: debugTSImportMappingFile,
		}
	} else {
		__antithesis_instrumentation__.Notify(34060)
	}
	__antithesis_instrumentation__.Notify(34056)

	return runStart(cmd, args, true)
}

func runStartJoin(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(34061)
	return runStart(cmd, args, false)
}

func runStart(cmd *cobra.Command, args []string, startSingleNode bool) (returnErr error) {
	__antithesis_instrumentation__.Notify(34062)
	tBegin := timeutil.Now()

	if ok, err := maybeRerunBackground(); ok {
		__antithesis_instrumentation__.Notify(34075)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34076)
	}
	__antithesis_instrumentation__.Notify(34063)

	disableOtherPermissionBits()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, drainSignals...)

	if err := exitIfDiskFull(vfs.Default, serverCfg.Stores.Specs); err != nil {
		__antithesis_instrumentation__.Notify(34077)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34078)
	}
	__antithesis_instrumentation__.Notify(34064)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ambientCtx := serverCfg.AmbientCtx

	var startupSpan *tracing.Span
	ctx, startupSpan = ambientCtx.AnnotateCtxWithSpan(ctx, "server start")

	stopper, err := setupAndInitializeLoggingAndProfiling(ctx, cmd, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(34079)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34080)
	}
	__antithesis_instrumentation__.Notify(34065)

	if err := serverCfg.Stores.PriorCriticalAlertError(); err != nil {
		__antithesis_instrumentation__.Notify(34081)
		return clierror.NewError(err, exit.FatalError())
	} else {
		__antithesis_instrumentation__.Notify(34082)
	}
	__antithesis_instrumentation__.Notify(34066)

	grpcutil.LowerSeverity(severity.WARNING)

	cgroups.AdjustMaxProcs(ctx)

	if !flagSetForCmd(cmd).Lookup(cliflags.Join.Name).Changed {
		__antithesis_instrumentation__.Notify(34083)
		err := errors.WithHint(
			errors.New("no --join flags provided to 'cockroach start'"),
			"Consider using 'cockroach init' or 'cockroach start-single-node' instead")
		return err
	} else {
		__antithesis_instrumentation__.Notify(34084)
	}
	__antithesis_instrumentation__.Notify(34067)

	if serverCfg.Settings.ExternalIODir, err = initExternalIODir(ctx, serverCfg.Stores.Specs[0]); err != nil {
		__antithesis_instrumentation__.Notify(34085)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34086)
	}
	__antithesis_instrumentation__.Notify(34068)

	if serverCfg.TempStorageConfig, err = initTempStorageConfig(
		ctx, serverCfg.Settings, stopper, serverCfg.Stores,
	); err != nil {
		__antithesis_instrumentation__.Notify(34087)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34088)
	}
	__antithesis_instrumentation__.Notify(34069)

	if serverCfg.StorageEngine == enginepb.EngineTypeDefault {
		__antithesis_instrumentation__.Notify(34089)
		serverCfg.StorageEngine = enginepb.EngineTypePebble
	} else {
		__antithesis_instrumentation__.Notify(34090)
	}
	__antithesis_instrumentation__.Notify(34070)

	if err := serverCfg.InitNode(ctx); err != nil {
		__antithesis_instrumentation__.Notify(34091)
		return errors.Wrap(err, "failed to initialize node")
	} else {
		__antithesis_instrumentation__.Notify(34092)
	}
	__antithesis_instrumentation__.Notify(34071)

	reportConfiguration(ctx)

	serverCfg.ReadyFn = func(waitForInit bool) {
		__antithesis_instrumentation__.Notify(34093)

		hintServerCmdFlags(ctx, cmd)

		if startCtx.pidFile != "" {
			__antithesis_instrumentation__.Notify(34097)
			log.Ops.Infof(ctx, "PID file: %s", startCtx.pidFile)
			if err := ioutil.WriteFile(startCtx.pidFile, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644); err != nil {
				__antithesis_instrumentation__.Notify(34098)
				log.Ops.Errorf(ctx, "failed writing the PID: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(34099)
			}
		} else {
			__antithesis_instrumentation__.Notify(34100)
		}
		__antithesis_instrumentation__.Notify(34094)

		if startCtx.listeningURLFile != "" {
			__antithesis_instrumentation__.Notify(34101)
			log.Ops.Infof(ctx, "listening URL file: %s", startCtx.listeningURLFile)

			sCtx := rpc.MakeSecurityContext(serverCfg.Config, security.ClusterTLSSettings(serverCfg.Settings), roachpb.SystemTenantID)
			pgURL, err := sCtx.PGURL(url.User(security.RootUser))
			if err != nil {
				__antithesis_instrumentation__.Notify(34103)
				log.Errorf(ctx, "failed computing the URL: %v", err)
				return
			} else {
				__antithesis_instrumentation__.Notify(34104)
			}
			__antithesis_instrumentation__.Notify(34102)

			if err = ioutil.WriteFile(startCtx.listeningURLFile, []byte(fmt.Sprintf("%s\n", pgURL.ToPQ())), 0644); err != nil {
				__antithesis_instrumentation__.Notify(34105)
				log.Ops.Errorf(ctx, "failed writing the URL: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(34106)
			}
		} else {
			__antithesis_instrumentation__.Notify(34107)
		}
		__antithesis_instrumentation__.Notify(34095)

		if waitForInit {
			__antithesis_instrumentation__.Notify(34108)
			log.Ops.Shout(ctx, severity.INFO,
				"initial startup completed.\n"+
					"Node will now attempt to join a running cluster, or wait for `cockroach init`.\n"+
					"Client connections will be accepted after this completes successfully.\n"+
					"Check the log file(s) for progress. ")
		} else {
			__antithesis_instrumentation__.Notify(34109)
		}
		__antithesis_instrumentation__.Notify(34096)

		log.Flush()

		if err := sdnotify.Ready(); err != nil {
			__antithesis_instrumentation__.Notify(34110)
			log.Ops.Errorf(ctx, "failed to signal readiness using systemd protocol: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(34111)
		}
	}
	__antithesis_instrumentation__.Notify(34072)

	serverCfg.DelayedBootstrapFn = func() {
		__antithesis_instrumentation__.Notify(34112)
		const msg = `The server appears to be unable to contact the other nodes in the cluster. Please try:

- starting the other nodes, if you haven't already;
- double-checking that the '--join' and '--listen'/'--advertise' flags are set up correctly;
- running the 'cockroach init' command if you are trying to initialize a new cluster.

If problems persist, please see %s.`
		docLink := docs.URL("cluster-setup-troubleshooting.html")
		if !startCtx.inBackground {
			__antithesis_instrumentation__.Notify(34113)
			log.Ops.Shoutf(ctx, severity.WARNING, msg, docLink)
		} else {
			__antithesis_instrumentation__.Notify(34114)

			log.Ops.Warningf(ctx, msg, docLink)
		}
	}
	__antithesis_instrumentation__.Notify(34073)

	initGEOS(ctx)

	log.Ops.Info(ctx, "starting cockroach node")

	var serverStatusMu serverStatus
	var s *server.Server
	errChan := make(chan error, 1)
	go func() {
		__antithesis_instrumentation__.Notify(34115)

		defer log.Flush()

		defer func() {
			__antithesis_instrumentation__.Notify(34117)
			if s != nil {
				__antithesis_instrumentation__.Notify(34118)

				logcrash.RecoverAndReportPanic(ctx, &s.ClusterSettings().SV)
			} else {
				__antithesis_instrumentation__.Notify(34119)
			}
		}()
		__antithesis_instrumentation__.Notify(34116)

		defer startupSpan.Finish()

		if err := func() error {
			__antithesis_instrumentation__.Notify(34120)

			var err error
			s, err = server.NewServer(serverCfg, stopper)
			if err != nil {
				__antithesis_instrumentation__.Notify(34127)
				return errors.Wrap(err, "failed to start server")
			} else {
				__antithesis_instrumentation__.Notify(34128)
			}
			__antithesis_instrumentation__.Notify(34121)

			serverStatusMu.Lock()
			draining := serverStatusMu.draining
			serverStatusMu.Unlock()
			if draining {
				__antithesis_instrumentation__.Notify(34129)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(34130)
			}
			__antithesis_instrumentation__.Notify(34122)

			if err := s.PreStart(ctx); err != nil {
				__antithesis_instrumentation__.Notify(34131)
				if le := (*server.ListenError)(nil); errors.As(err, &le) {
					__antithesis_instrumentation__.Notify(34133)
					const errorPrefix = "consider changing the port via --%s"
					if le.Addr == serverCfg.Addr {
						__antithesis_instrumentation__.Notify(34134)
						err = errors.Wrapf(err, errorPrefix, cliflags.ListenAddr.Name)
					} else {
						__antithesis_instrumentation__.Notify(34135)
						if le.Addr == serverCfg.HTTPAddr {
							__antithesis_instrumentation__.Notify(34136)
							err = errors.Wrapf(err, errorPrefix, cliflags.ListenHTTPAddr.Name)
						} else {
							__antithesis_instrumentation__.Notify(34137)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(34138)
				}
				__antithesis_instrumentation__.Notify(34132)

				return errors.Wrap(err, "cockroach server exited with error")
			} else {
				__antithesis_instrumentation__.Notify(34139)
			}
			__antithesis_instrumentation__.Notify(34123)

			serverStatusMu.Lock()
			serverStatusMu.started = true
			serverStatusMu.Unlock()

			if !cluster.TelemetryOptOut() {
				__antithesis_instrumentation__.Notify(34140)
				s.StartDiagnostics(ctx)
			} else {
				__antithesis_instrumentation__.Notify(34141)
			}
			__antithesis_instrumentation__.Notify(34124)
			initialStart := s.InitialStart()

			if err := runInitialSQL(ctx, s, startSingleNode, "", ""); err != nil {
				__antithesis_instrumentation__.Notify(34142)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34143)
			}
			__antithesis_instrumentation__.Notify(34125)

			if err := s.AcceptClients(ctx); err != nil {
				__antithesis_instrumentation__.Notify(34144)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34145)
			}
			__antithesis_instrumentation__.Notify(34126)

			return reportServerInfo(ctx, tBegin, &serverCfg, s.ClusterSettings(),
				true, initialStart, uuid.UUID{})
		}(); err != nil {
			__antithesis_instrumentation__.Notify(34146)
			errChan <- err
		} else {
			__antithesis_instrumentation__.Notify(34147)
		}
	}()
	__antithesis_instrumentation__.Notify(34074)

	return waitForShutdown(

		func() serverShutdownInterface { __antithesis_instrumentation__.Notify(34148); return s },
		stopper, errChan, signalCh,
		&serverStatusMu)
}

type serverStatus struct {
	syncutil.Mutex

	started, draining bool
}

type serverShutdownInterface interface {
	AnnotateCtx(context.Context) context.Context
	Drain(ctx context.Context, verbose bool) (uint64, redact.RedactableString, error)
}

func waitForShutdown(
	getS func() serverShutdownInterface,
	stopper *stop.Stopper,
	errChan chan error,
	signalCh chan os.Signal,
	serverStatusMu *serverStatus,
) (returnErr error) {
	__antithesis_instrumentation__.Notify(34149)

	shutdownCtx, shutdownSpan := serverCfg.AmbientCtx.AnnotateCtxWithSpan(context.Background(), "server shutdown")
	defer shutdownSpan.Finish()

	stopWithoutDrain := make(chan struct{})

	select {
	case err := <-errChan:
		__antithesis_instrumentation__.Notify(34153)

		log.StartAlwaysFlush()
		return err

	case <-stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(34154)

		<-stopper.IsStopped()

		log.StartAlwaysFlush()
		return nil

	case sig := <-signalCh:
		__antithesis_instrumentation__.Notify(34155)

		log.StartAlwaysFlush()

		log.Ops.Infof(shutdownCtx, "received signal '%s'", sig)
		switch sig {
		case os.Interrupt:
			__antithesis_instrumentation__.Notify(34158)

			returnErr = clierror.NewErrorWithSeverity(
				errors.New("interrupted"),
				exit.Interrupted(),

				severity.INFO,
			)
			msgDouble := "Note: a second interrupt will skip graceful shutdown and terminate forcefully"
			fmt.Fprintln(os.Stdout, msgDouble)
		default:
			__antithesis_instrumentation__.Notify(34159)
		}
		__antithesis_instrumentation__.Notify(34156)

		go func() {
			__antithesis_instrumentation__.Notify(34160)
			serverStatusMu.Lock()
			serverStatusMu.draining = true
			drainingIsSafe := serverStatusMu.started
			serverStatusMu.Unlock()

			if !drainingIsSafe {
				__antithesis_instrumentation__.Notify(34163)
				close(stopWithoutDrain)
				return
			} else {
				__antithesis_instrumentation__.Notify(34164)
			}
			__antithesis_instrumentation__.Notify(34161)

			drainCtx := logtags.AddTag(getS().AnnotateCtx(context.Background()), "server drain process", nil)

			var (
				remaining     = uint64(math.MaxUint64)
				prevRemaining = uint64(math.MaxUint64)
				verbose       = false
			)

			for ; ; prevRemaining = remaining {
				__antithesis_instrumentation__.Notify(34165)
				var err error
				remaining, _, err = getS().Drain(drainCtx, verbose)
				if err != nil {
					__antithesis_instrumentation__.Notify(34169)
					log.Ops.Errorf(drainCtx, "graceful drain failed: %v", err)
					break
				} else {
					__antithesis_instrumentation__.Notify(34170)
				}
				__antithesis_instrumentation__.Notify(34166)
				if remaining == 0 {
					__antithesis_instrumentation__.Notify(34171)

					break
				} else {
					__antithesis_instrumentation__.Notify(34172)
				}
				__antithesis_instrumentation__.Notify(34167)

				if remaining >= prevRemaining {
					__antithesis_instrumentation__.Notify(34173)
					verbose = true
				} else {
					__antithesis_instrumentation__.Notify(34174)
				}
				__antithesis_instrumentation__.Notify(34168)

				time.Sleep(200 * time.Millisecond)
			}
			__antithesis_instrumentation__.Notify(34162)

			stopper.Stop(drainCtx)
		}()

	case <-log.FatalChan():
		__antithesis_instrumentation__.Notify(34157)

		stopper.Stop(shutdownCtx)

		select {}
	}
	__antithesis_instrumentation__.Notify(34150)

	const msgDrain = "initiating graceful shutdown of server"
	log.Ops.Info(shutdownCtx, msgDrain)
	fmt.Fprintln(os.Stdout, msgDrain)

	go func() {
		__antithesis_instrumentation__.Notify(34175)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			__antithesis_instrumentation__.Notify(34176)
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(34177)
				log.Ops.Infof(shutdownCtx, "%d running tasks", stopper.NumTasks())
			case <-stopper.IsStopped():
				__antithesis_instrumentation__.Notify(34178)
				return
			case <-stopWithoutDrain:
				__antithesis_instrumentation__.Notify(34179)
				return
			}
		}
	}()
	__antithesis_instrumentation__.Notify(34151)

	const hardShutdownHint = " - node may take longer to restart & clients may need to wait for leases to expire"
	for {
		__antithesis_instrumentation__.Notify(34180)
		select {
		case sig := <-signalCh:
			__antithesis_instrumentation__.Notify(34182)
			switch sig {
			case termSignal:
				__antithesis_instrumentation__.Notify(34186)

				log.Ops.Infof(shutdownCtx, "received additional signal '%s'; continuing graceful shutdown", sig)
				continue
			default:
				__antithesis_instrumentation__.Notify(34187)
			}
			__antithesis_instrumentation__.Notify(34183)

			log.Ops.Shoutf(shutdownCtx, severity.ERROR,
				"received signal '%s' during shutdown, initiating hard shutdown%s",
				redact.Safe(sig), redact.Safe(hardShutdownHint))
			handleSignalDuringShutdown(sig)
			panic("unreachable")

		case <-stopper.IsStopped():
			__antithesis_instrumentation__.Notify(34184)
			const msgDone = "server drained and shutdown completed"
			log.Ops.Infof(shutdownCtx, msgDone)
			fmt.Fprintln(os.Stdout, msgDone)

		case <-stopWithoutDrain:
			__antithesis_instrumentation__.Notify(34185)
			const msgDone = "too early to drain; used hard shutdown instead"
			log.Ops.Infof(shutdownCtx, msgDone)
			fmt.Fprintln(os.Stdout, msgDone)
		}
		__antithesis_instrumentation__.Notify(34181)
		break
	}
	__antithesis_instrumentation__.Notify(34152)

	return returnErr
}

func reportServerInfo(
	ctx context.Context,
	startTime time.Time,
	serverCfg *server.Config,
	st *cluster.Settings,
	isHostNode, initialStart bool,
	tenantClusterID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(34188)
	srvS := redact.SafeString("SQL server")
	if isHostNode {
		__antithesis_instrumentation__.Notify(34203)
		srvS = "node"
	} else {
		__antithesis_instrumentation__.Notify(34204)
	}
	__antithesis_instrumentation__.Notify(34189)

	var buf redact.StringBuilder
	info := build.GetInfo()
	buf.Printf("CockroachDB %s starting at %s (took %0.1fs)\n", srvS, timeutil.Now(), timeutil.Since(startTime).Seconds())
	buf.Printf("build:\t%s %s @ %s (%s)\n",
		redact.Safe(info.Distribution), redact.Safe(info.Tag), redact.Safe(info.Time), redact.Safe(info.GoVersion))
	buf.Printf("webui:\t%s\n", serverCfg.AdminURL())

	sCtx := rpc.MakeSecurityContext(serverCfg.Config, security.ClusterTLSSettings(serverCfg.Settings), roachpb.SystemTenantID)
	pgURL, err := sCtx.PGURL(url.User(security.RootUser))
	if err != nil {
		__antithesis_instrumentation__.Notify(34205)
		log.Ops.Errorf(ctx, "failed computing the URL: %v", err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34206)
	}
	__antithesis_instrumentation__.Notify(34190)
	buf.Printf("sql:\t%s\n", pgURL.ToPQ())
	buf.Printf("sql (JDBC):\t%s\n", pgURL.ToJDBC())

	buf.Printf("RPC client flags:\t%s\n", clientFlagsRPC())
	if len(serverCfg.SocketFile) != 0 {
		__antithesis_instrumentation__.Notify(34207)
		buf.Printf("socket:\t%s\n", serverCfg.SocketFile)
	} else {
		__antithesis_instrumentation__.Notify(34208)
	}
	__antithesis_instrumentation__.Notify(34191)
	logNum := 1
	_ = cliCtx.logConfig.IterateDirectories(func(d string) error {
		__antithesis_instrumentation__.Notify(34209)
		if logNum == 1 {
			__antithesis_instrumentation__.Notify(34211)

			buf.Printf("logs:\t%s\n", d)
		} else {
			__antithesis_instrumentation__.Notify(34212)
			buf.Printf("logs[%d]:\t%s\n", logNum, d)
		}
		__antithesis_instrumentation__.Notify(34210)
		logNum++
		return nil
	})
	__antithesis_instrumentation__.Notify(34192)
	if serverCfg.Attrs != "" {
		__antithesis_instrumentation__.Notify(34213)
		buf.Printf("attrs:\t%s\n", serverCfg.Attrs)
	} else {
		__antithesis_instrumentation__.Notify(34214)
	}
	__antithesis_instrumentation__.Notify(34193)
	if len(serverCfg.Locality.Tiers) > 0 {
		__antithesis_instrumentation__.Notify(34215)
		buf.Printf("locality:\t%s\n", serverCfg.Locality)
	} else {
		__antithesis_instrumentation__.Notify(34216)
	}
	__antithesis_instrumentation__.Notify(34194)
	if tmpDir := serverCfg.SQLConfig.TempStorageConfig.Path; tmpDir != "" {
		__antithesis_instrumentation__.Notify(34217)
		buf.Printf("temp dir:\t%s\n", tmpDir)
	} else {
		__antithesis_instrumentation__.Notify(34218)
	}
	__antithesis_instrumentation__.Notify(34195)
	if ext := st.ExternalIODir; ext != "" {
		__antithesis_instrumentation__.Notify(34219)
		buf.Printf("external I/O path: \t%s\n", ext)
	} else {
		__antithesis_instrumentation__.Notify(34220)
		buf.Printf("external I/O path: \t<disabled>\n")
	}
	__antithesis_instrumentation__.Notify(34196)
	for i, spec := range serverCfg.Stores.Specs {
		__antithesis_instrumentation__.Notify(34221)
		buf.Printf("store[%d]:\t%s\n", i, spec)
	}
	__antithesis_instrumentation__.Notify(34197)
	buf.Printf("storage engine: \t%s\n", &serverCfg.StorageEngine)

	if baseCfg.ClusterName != "" {
		__antithesis_instrumentation__.Notify(34222)
		buf.Printf("cluster name:\t%s\n", baseCfg.ClusterName)
	} else {
		__antithesis_instrumentation__.Notify(34223)
	}
	__antithesis_instrumentation__.Notify(34198)
	clusterID := serverCfg.BaseConfig.ClusterIDContainer.Get()
	if tenantClusterID.Equal(uuid.Nil) {
		__antithesis_instrumentation__.Notify(34224)
		buf.Printf("clusterID:\t%s\n", clusterID)
	} else {
		__antithesis_instrumentation__.Notify(34225)
		buf.Printf("storage clusterID:\t%s\n", clusterID)
		buf.Printf("tenant clusterID:\t%s\n", tenantClusterID)
	}
	__antithesis_instrumentation__.Notify(34199)
	nodeID := serverCfg.BaseConfig.IDContainer.Get()
	if isHostNode {
		__antithesis_instrumentation__.Notify(34226)
		if initialStart {
			__antithesis_instrumentation__.Notify(34228)
			if nodeID == kvserver.FirstNodeID {
				__antithesis_instrumentation__.Notify(34229)
				buf.Printf("status:\tinitialized new cluster\n")
			} else {
				__antithesis_instrumentation__.Notify(34230)
				buf.Printf("status:\tinitialized new node, joined pre-existing cluster\n")
			}
		} else {
			__antithesis_instrumentation__.Notify(34231)
			buf.Printf("status:\trestarted pre-existing node\n")
		}
		__antithesis_instrumentation__.Notify(34227)

		buf.Printf("nodeID:\t%d\n", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(34232)

		buf.Printf("tenantID:\t%s\n", serverCfg.SQLConfig.TenantID)
		buf.Printf("instanceID:\t%d\n", nodeID)

		if kvAddrs := serverCfg.SQLConfig.TenantKVAddrs; len(kvAddrs) > 0 {
			__antithesis_instrumentation__.Notify(34233)

			buf.Printf("KV addresses:\t")
			comma := redact.SafeString("")
			for _, addr := range serverCfg.SQLConfig.TenantKVAddrs {
				__antithesis_instrumentation__.Notify(34235)
				buf.Printf("%s%s", comma, addr)
				comma = ", "
			}
			__antithesis_instrumentation__.Notify(34234)
			buf.Printf("\n")
		} else {
			__antithesis_instrumentation__.Notify(34236)
		}
	}
	__antithesis_instrumentation__.Notify(34200)

	msg, err := expandTabsInRedactableBytes(buf.RedactableBytes())
	if err != nil {
		__antithesis_instrumentation__.Notify(34237)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34238)
	}
	__antithesis_instrumentation__.Notify(34201)
	msgS := msg.ToString()
	log.Ops.Infof(ctx, "%s startup completed:\n%s", srvS, msgS)
	if !startCtx.inBackground && func() bool {
		__antithesis_instrumentation__.Notify(34239)
		return !log.LoggingToStderr(severity.INFO) == true
	}() == true {
		__antithesis_instrumentation__.Notify(34240)
		fmt.Print(msgS.StripMarkers())
	} else {
		__antithesis_instrumentation__.Notify(34241)
	}
	__antithesis_instrumentation__.Notify(34202)

	return nil
}

func expandTabsInRedactableBytes(s redact.RedactableBytes) (redact.RedactableBytes, error) {
	__antithesis_instrumentation__.Notify(34242)
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)
	if _, err := tw.Write([]byte(s)); err != nil {
		__antithesis_instrumentation__.Notify(34245)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(34246)
	}
	__antithesis_instrumentation__.Notify(34243)
	if err := tw.Flush(); err != nil {
		__antithesis_instrumentation__.Notify(34247)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(34248)
	}
	__antithesis_instrumentation__.Notify(34244)
	return redact.RedactableBytes(buf.Bytes()), nil
}

func hintServerCmdFlags(ctx context.Context, cmd *cobra.Command) {
	__antithesis_instrumentation__.Notify(34249)
	pf := flagSetForCmd(cmd)

	listenAddrSpecified := pf.Lookup(cliflags.ListenAddr.Name).Changed || func() bool {
		__antithesis_instrumentation__.Notify(34250)
		return pf.Lookup(cliflags.ServerHost.Name).Changed == true
	}() == true
	advAddrSpecified := pf.Lookup(cliflags.AdvertiseAddr.Name).Changed || func() bool {
		__antithesis_instrumentation__.Notify(34251)
		return pf.Lookup(cliflags.AdvertiseHost.Name).Changed == true
	}() == true

	if !listenAddrSpecified && func() bool {
		__antithesis_instrumentation__.Notify(34252)
		return !advAddrSpecified == true
	}() == true {
		__antithesis_instrumentation__.Notify(34253)
		host, _, _ := net.SplitHostPort(serverCfg.AdvertiseAddr)
		log.Ops.Shoutf(ctx, severity.WARNING,
			"neither --listen-addr nor --advertise-addr was specified.\n"+
				"The server will advertise %q to other nodes, is this routable?\n\n"+
				"Consider using:\n"+
				"- for local-only servers:  --listen-addr=localhost\n"+
				"- for multi-node clusters: --advertise-addr=<host/IP addr>\n", host)
	} else {
		__antithesis_instrumentation__.Notify(34254)
	}
}

func clientFlagsRPC() string {
	__antithesis_instrumentation__.Notify(34255)
	flags := []string{os.Args[0], "<client cmd>"}
	if serverCfg.AdvertiseAddr != "" {
		__antithesis_instrumentation__.Notify(34258)
		flags = append(flags, "--host="+serverCfg.AdvertiseAddr)
	} else {
		__antithesis_instrumentation__.Notify(34259)
	}
	__antithesis_instrumentation__.Notify(34256)
	if startCtx.serverInsecure {
		__antithesis_instrumentation__.Notify(34260)
		flags = append(flags, "--insecure")
	} else {
		__antithesis_instrumentation__.Notify(34261)
		flags = append(flags, "--certs-dir="+startCtx.serverSSLCertsDir)
	}
	__antithesis_instrumentation__.Notify(34257)
	return strings.Join(flags, " ")
}

func reportConfiguration(ctx context.Context) {
	__antithesis_instrumentation__.Notify(34262)
	serverCfg.Report(ctx)
	if envVarsUsed := envutil.GetEnvVarsUsed(); len(envVarsUsed) > 0 {
		__antithesis_instrumentation__.Notify(34264)
		log.Ops.Infof(ctx, "using local environment variables:\n%s", redact.Join("\n", envVarsUsed))
	} else {
		__antithesis_instrumentation__.Notify(34265)
	}
	__antithesis_instrumentation__.Notify(34263)

	log.Ops.Infof(ctx, "process identity: %s", sysutil.ProcessIdentity())
}

func maybeWarnMemorySizes(ctx context.Context) {
	__antithesis_instrumentation__.Notify(34266)

	if !cacheSizeValue.IsSet() {
		__antithesis_instrumentation__.Notify(34268)
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Using the default setting for --cache (%s).\n", cacheSizeValue)
		fmt.Fprintf(&buf, "  A significantly larger value is usually needed for good performance.\n")
		if size, err := status.GetTotalMemory(ctx); err == nil {
			__antithesis_instrumentation__.Notify(34270)
			fmt.Fprintf(&buf, "  If you have a dedicated server a reasonable setting is --cache=.25 (%s).",
				humanizeutil.IBytes(size/4))
		} else {
			__antithesis_instrumentation__.Notify(34271)
			fmt.Fprintf(&buf, "  If you have a dedicated server a reasonable setting is 25%% of physical memory.")
		}
		__antithesis_instrumentation__.Notify(34269)
		log.Ops.Warningf(ctx, "%s", buf.String())
	} else {
		__antithesis_instrumentation__.Notify(34272)
	}
	__antithesis_instrumentation__.Notify(34267)

	if maxMemory, err := status.GetTotalMemory(ctx); err == nil {
		__antithesis_instrumentation__.Notify(34273)
		requestedMem := serverCfg.CacheSize + serverCfg.MemoryPoolSize + serverCfg.TimeSeriesServerConfig.QueryMemoryMax
		maxRecommendedMem := int64(.75 * float64(maxMemory))
		if requestedMem > maxRecommendedMem {
			__antithesis_instrumentation__.Notify(34274)
			log.Ops.Shoutf(ctx, severity.WARNING,
				"the sum of --max-sql-memory (%s), --cache (%s), and --max-tsdb-memory (%s) is larger than 75%% of total RAM (%s).\nThis server is running at increased risk of memory-related failures.",
				sqlSizeValue, cacheSizeValue, tsdbSizeValue, humanizeutil.IBytes(maxRecommendedMem))
		} else {
			__antithesis_instrumentation__.Notify(34275)
		}
	} else {
		__antithesis_instrumentation__.Notify(34276)
	}
}

func exitIfDiskFull(fs vfs.FS, specs []base.StoreSpec) error {
	__antithesis_instrumentation__.Notify(34277)
	var cause error
	var ballastPaths []string
	var ballastMissing bool
	for _, spec := range specs {
		__antithesis_instrumentation__.Notify(34281)
		isDiskFull, err := storage.IsDiskFull(fs, spec)
		if err != nil {
			__antithesis_instrumentation__.Notify(34285)
			return err
		} else {
			__antithesis_instrumentation__.Notify(34286)
		}
		__antithesis_instrumentation__.Notify(34282)
		if !isDiskFull {
			__antithesis_instrumentation__.Notify(34287)
			continue
		} else {
			__antithesis_instrumentation__.Notify(34288)
		}
		__antithesis_instrumentation__.Notify(34283)
		path := base.EmergencyBallastFile(fs.PathJoin, spec.Path)
		ballastPaths = append(ballastPaths, path)
		if _, err := fs.Stat(path); oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(34289)
			ballastMissing = true
		} else {
			__antithesis_instrumentation__.Notify(34290)
		}
		__antithesis_instrumentation__.Notify(34284)
		cause = errors.CombineErrors(cause, errors.Newf(`store %s: out of disk space`, spec.Path))
	}
	__antithesis_instrumentation__.Notify(34278)
	if cause == nil {
		__antithesis_instrumentation__.Notify(34291)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(34292)
	}
	__antithesis_instrumentation__.Notify(34279)

	err := clierror.NewError(cause, exit.DiskFull())
	if ballastMissing {
		__antithesis_instrumentation__.Notify(34293)
		return errors.WithHint(err, `At least one ballast file is missing.
You may need to replace this node because there is
insufficient disk space to start.`)
	} else {
		__antithesis_instrumentation__.Notify(34294)
	}
	__antithesis_instrumentation__.Notify(34280)

	ballastPathsStr := strings.Join(ballastPaths, "\n")
	err = errors.WithHintf(err, `Deleting or truncating the ballast file(s) at
%s
may reclaim enough space to start. Proceed with caution. Complete
disk space exhaustion may result in node loss.`, ballastPathsStr)
	return err
}

func setupAndInitializeLoggingAndProfiling(
	ctx context.Context, cmd *cobra.Command, isServerCmd bool,
) (stopper *stop.Stopper, err error) {
	__antithesis_instrumentation__.Notify(34295)
	if err := setupLogging(ctx, cmd, isServerCmd, true); err != nil {
		__antithesis_instrumentation__.Notify(34299)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(34300)
	}
	__antithesis_instrumentation__.Notify(34296)

	if startCtx.serverInsecure {
		__antithesis_instrumentation__.Notify(34301)

		addr := startCtx.serverListenAddr
		if addr == "" {
			__antithesis_instrumentation__.Notify(34303)
			addr = "any of your IP addresses"
		} else {
			__antithesis_instrumentation__.Notify(34304)
		}
		__antithesis_instrumentation__.Notify(34302)
		log.Ops.Shoutf(ctx, severity.WARNING,
			"ALL SECURITY CONTROLS HAVE BEEN DISABLED!\n\n"+
				"This mode is intended for non-production testing only.\n"+
				"\n"+
				"In this mode:\n"+
				"- Your cluster is open to any client that can access %s.\n"+
				"- Intruders with access to your machine or network can observe client-server traffic.\n"+
				"- Intruders can log in without password and read or write any data in the cluster.\n"+
				"- Intruders can consume all your server's resources and cause unavailability.",
			addr)
		log.Ops.Shoutf(ctx, severity.INFO,
			"To start a secure server without mandating TLS for clients,\n"+
				"consider --accept-sql-without-tls instead. For other options, see:\n\n"+
				"- %s\n"+
				"- %s",
			build.MakeIssueURL(53404),
			redact.Safe(docs.URL("secure-a-cluster.html")),
		)
	} else {
		__antithesis_instrumentation__.Notify(34305)
	}
	__antithesis_instrumentation__.Notify(34297)

	if len(serverCfg.Locality.Tiers) > 0 {
		__antithesis_instrumentation__.Notify(34306)
		if _, containsRegion := serverCfg.Locality.Find("region"); !containsRegion {
			__antithesis_instrumentation__.Notify(34307)
			const warningString string = "The --locality flag does not contain a \"region\" tier. To add regions\n" +
				"to databases, the --locality flag must contain a \"region\" tier.\n" +
				"For more information, see:\n\n" +
				"- %s"
			log.Shoutf(ctx, severity.WARNING, warningString,
				redact.Safe(docs.URL("cockroach-start.html#locality")))
		} else {
			__antithesis_instrumentation__.Notify(34308)
		}
	} else {
		__antithesis_instrumentation__.Notify(34309)
	}
	__antithesis_instrumentation__.Notify(34298)

	maybeWarnMemorySizes(ctx)

	info := build.GetInfo()
	log.Ops.Infof(ctx, "%s", info.Short())

	initTraceDir(ctx, serverCfg.InflightTraceDirName)
	initCPUProfile(ctx, serverCfg.CPUProfileDirName, serverCfg.Settings)
	initBlockProfile()
	initMutexProfile()

	stopper = stop.NewStopper()
	log.Event(ctx, "initialized profiles")

	return stopper, nil
}

func addrWithDefaultHost(addr string) (string, error) {
	__antithesis_instrumentation__.Notify(34310)
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(34313)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(34314)
	}
	__antithesis_instrumentation__.Notify(34311)
	if host == "" {
		__antithesis_instrumentation__.Notify(34315)
		host = "localhost"
	} else {
		__antithesis_instrumentation__.Notify(34316)
	}
	__antithesis_instrumentation__.Notify(34312)
	return net.JoinHostPort(host, port), nil
}

func getClientGRPCConn(
	ctx context.Context, cfg server.Config,
) (*grpc.ClientConn, *hlc.Clock, func(), error) {
	__antithesis_instrumentation__.Notify(34317)
	if ctx.Done() == nil {
		__antithesis_instrumentation__.Notify(34325)
		return nil, nil, nil, errors.New("context must be cancellable")
	} else {
		__antithesis_instrumentation__.Notify(34326)
	}
	__antithesis_instrumentation__.Notify(34318)

	clock := hlc.NewClock(hlc.UnixNano, 0)
	tracer := cfg.Tracer
	if tracer == nil {
		__antithesis_instrumentation__.Notify(34327)
		tracer = tracing.NewTracer()
	} else {
		__antithesis_instrumentation__.Notify(34328)
	}
	__antithesis_instrumentation__.Notify(34319)
	stopper := stop.NewStopper(stop.WithTracer(tracer))
	rpcContext := rpc.NewContext(ctx,
		rpc.ContextOptions{
			TenantID: roachpb.SystemTenantID,
			Config:   cfg.Config,
			Clock:    clock,
			Stopper:  stopper,
			Settings: cfg.Settings,

			ClientOnly: true,
		})
	if cfg.TestingKnobs.Server != nil {
		__antithesis_instrumentation__.Notify(34329)
		rpcContext.Knobs = cfg.TestingKnobs.Server.(*server.TestingKnobs).ContextTestingKnobs
	} else {
		__antithesis_instrumentation__.Notify(34330)
	}
	__antithesis_instrumentation__.Notify(34320)
	addr, err := addrWithDefaultHost(cfg.AdvertiseAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(34331)
		stopper.Stop(ctx)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(34332)
	}
	__antithesis_instrumentation__.Notify(34321)

	conn, err := rpcContext.GRPCUnvalidatedDial(addr).Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(34333)
		stopper.Stop(ctx)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(34334)
	}
	__antithesis_instrumentation__.Notify(34322)
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(34335)
		_ = conn.Close()
	}))
	__antithesis_instrumentation__.Notify(34323)

	closer := func() {
		__antithesis_instrumentation__.Notify(34336)
		stopper.Stop(ctx)
	}
	__antithesis_instrumentation__.Notify(34324)
	return conn, clock, closer, nil
}

func initGEOS(ctx context.Context) {
	__antithesis_instrumentation__.Notify(34337)
	loc, err := geos.EnsureInit(geos.EnsureInitErrorDisplayPrivate, startCtx.geoLibsDir)
	if err != nil {
		__antithesis_instrumentation__.Notify(34338)
		log.Ops.Warningf(ctx, "could not initialize GEOS - spatial functions may not be available: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(34339)
		log.Ops.Infof(ctx, "GEOS loaded from directory %s", loc)
	}
}
