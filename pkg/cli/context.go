package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl"
	"github.com/cockroachdb/cockroach/pkg/cli/clicfg"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlcfg"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlshell"
	"github.com/cockroachdb/cockroach/pkg/cli/democluster"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/ts"
	"github.com/cockroachdb/cockroach/pkg/util/log/logconfig"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func initCLIDefaults() {
	__antithesis_instrumentation__.Notify(30339)
	setServerContextDefaults()

	baseCfg.InitDefaults()
	setCliContextDefaults()
	setSQLConnContextDefaults()
	setSQLExecContextDefaults()
	setSQLContextDefaults()
	setZipContextDefaults()
	setDumpContextDefaults()
	setDebugContextDefaults()
	setStartContextDefaults()
	setQuitContextDefaults()
	setNodeContextDefaults()
	setSqlfmtContextDefaults()
	setConvContextDefaults()
	setDemoContextDefaults()
	setStmtDiagContextDefaults()
	setAuthContextDefaults()
	setImportContextDefaults()
	setProxyContextDefaults()
	setTestDirectorySvrContextDefaults()
	setUserfileContextDefaults()
	setCertContextDefaults()
	setDebugRecoverContextDefaults()
	setDebugSendKVBatchContextDefaults()

	initPreFlagsDefaults()

	clearFlagChanges(cockroachCmd)
}

func clearFlagChanges(cmd *cobra.Command) {
	__antithesis_instrumentation__.Notify(30340)
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) { __antithesis_instrumentation__.Notify(30342); f.Changed = false })
	__antithesis_instrumentation__.Notify(30341)
	for _, subCmd := range cmd.Commands() {
		__antithesis_instrumentation__.Notify(30343)
		clearFlagChanges(subCmd)
	}
}

var serverCfg = func() server.Config {
	__antithesis_instrumentation__.Notify(30344)
	st := cluster.MakeClusterSettings()
	logcrash.SetGlobalSettings(&st.SV)

	return server.MakeConfig(context.Background(), st)
}()

func setServerContextDefaults() {
	__antithesis_instrumentation__.Notify(30345)
	serverCfg.BaseConfig.DefaultZoneConfig = zonepb.DefaultZoneConfig()

	serverCfg.ClockDevicePath = ""
	serverCfg.ExternalIODirConfig = base.ExternalIODirConfig{}
	serverCfg.GoroutineDumpDirName = ""
	serverCfg.HeapProfileDirName = ""
	serverCfg.CPUProfileDirName = ""
	serverCfg.InflightTraceDirName = ""

	serverCfg.AutoInitializeCluster = false
	serverCfg.KVConfig.ReadyFn = nil
	serverCfg.KVConfig.DelayedBootstrapFn = nil
	serverCfg.KVConfig.JoinList = nil
	serverCfg.KVConfig.JoinPreferSRVRecords = false
	serverCfg.KVConfig.DefaultSystemZoneConfig = zonepb.DefaultSystemZoneConfig()

	storeSpec, _ := base.NewStoreSpec(server.DefaultStorePath)
	serverCfg.Stores = base.StoreSpecList{Specs: []base.StoreSpec{storeSpec}}

	serverCfg.TenantKVAddrs = []string{"127.0.0.1:26257"}

	serverCfg.SQLConfig.SocketFile = ""

	if bytes, _ := memoryPercentResolver(25); bytes != 0 {
		__antithesis_instrumentation__.Notify(30347)
		serverCfg.SQLConfig.MemoryPoolSize = bytes
	} else {
		__antithesis_instrumentation__.Notify(30348)
	}
	__antithesis_instrumentation__.Notify(30346)

	if bytes, _ := memoryPercentResolver(1); bytes != 0 {
		__antithesis_instrumentation__.Notify(30349)
		if bytes > ts.DefaultQueryMemoryMax {
			__antithesis_instrumentation__.Notify(30350)
			serverCfg.TimeSeriesServerConfig.QueryMemoryMax = bytes
		} else {
			__antithesis_instrumentation__.Notify(30351)
			serverCfg.TimeSeriesServerConfig.QueryMemoryMax = ts.DefaultQueryMemoryMax
		}
	} else {
		__antithesis_instrumentation__.Notify(30352)
	}
}

var baseCfg = serverCfg.Config

type cliContext struct {
	*base.Config

	clicfg.Context

	cmdTimeout time.Duration

	clientConnHost string

	clientConnPort string

	certPrincipalMap []string

	sqlConnUser, sqlConnDBName string

	sqlConnURL *pgurl.URL

	allowUnencryptedClientPassword bool

	logConfigInput settableString

	logConfig logconfig.Config

	deprecatedLogOverrides *logConfigFlags

	ambiguousLogDir bool

	showVersionUsingOnlyBuildTag bool
}

var cliCtx = cliContext{
	Config: baseCfg,

	deprecatedLogOverrides: newLogConfigOverrides(),
}

func setCliContextDefaults() {
	__antithesis_instrumentation__.Notify(30353)

	cliCtx.IsInteractive = false
	cliCtx.EmbeddedMode = false
	cliCtx.cmdTimeout = 0
	cliCtx.clientConnHost = ""
	cliCtx.clientConnPort = base.DefaultPort
	cliCtx.certPrincipalMap = nil
	cliCtx.sqlConnURL = nil
	cliCtx.sqlConnUser = security.RootUser
	cliCtx.sqlConnDBName = ""
	cliCtx.allowUnencryptedClientPassword = false
	cliCtx.logConfigInput = settableString{s: ""}
	cliCtx.logConfig = logconfig.Config{}
	cliCtx.ambiguousLogDir = false

	cliCtx.deprecatedLogOverrides.reset()
	cliCtx.showVersionUsingOnlyBuildTag = false
}

var sqlConnCtx = clisqlclient.Context{
	CliCtx: &cliCtx.Context,
}

func setSQLConnContextDefaults() {
	__antithesis_instrumentation__.Notify(30354)

	sqlConnCtx.DebugMode = false
	sqlConnCtx.Echo = false
	sqlConnCtx.EnableServerExecutionTimings = false
}

var certCtx struct {
	certsDir              string
	caKey                 string
	keySize               int
	caCertificateLifetime time.Duration
	certificateLifetime   time.Duration
	allowCAKeyReuse       bool
	overwriteFiles        bool
	generatePKCS8Key      bool

	certPrincipalMap []string
}

func setCertContextDefaults() {
	__antithesis_instrumentation__.Notify(30355)
	certCtx.certsDir = base.DefaultCertsDirectory
	certCtx.caKey = ""
	certCtx.keySize = defaultKeySize
	certCtx.caCertificateLifetime = defaultCALifetime
	certCtx.certificateLifetime = defaultCertLifetime
	certCtx.allowCAKeyReuse = false
	certCtx.overwriteFiles = false
	certCtx.generatePKCS8Key = false
	certCtx.certPrincipalMap = nil
}

var sqlExecCtx = clisqlexec.Context{
	CliCtx: &cliCtx.Context,
}

func PrintQueryOutput(w io.Writer, cols []string, allRows clisqlexec.RowStrIter) error {
	__antithesis_instrumentation__.Notify(30356)
	return sqlExecCtx.PrintQueryOutput(w, stderr, cols, allRows)
}

func setSQLExecContextDefaults() {
	__antithesis_instrumentation__.Notify(30357)

	sqlExecCtx.TerminalOutput = isatty.IsTerminal(os.Stdout.Fd())
	sqlExecCtx.TableDisplayFormat = clisqlexec.TableDisplayTSV
	sqlExecCtx.TableBorderMode = 0
	if sqlExecCtx.TerminalOutput {
		__antithesis_instrumentation__.Notify(30359)

		sqlExecCtx.TableDisplayFormat = clisqlexec.TableDisplayTable
	} else {
		__antithesis_instrumentation__.Notify(30360)
	}
	__antithesis_instrumentation__.Notify(30358)
	sqlExecCtx.ShowTimes = false
	sqlExecCtx.VerboseTimings = false
}

var sqlCtx = func() *clisqlcfg.Context {
	__antithesis_instrumentation__.Notify(30361)
	cfg := &clisqlcfg.Context{
		CliCtx:  &cliCtx.Context,
		ConnCtx: &sqlConnCtx,
		ExecCtx: &sqlExecCtx,
	}
	return cfg
}()

func setSQLContextDefaults() {
	__antithesis_instrumentation__.Notify(30362)
	sqlCtx.LoadDefaults(os.Stdout, stderr)
}

var zipCtx zipContext

type zipContext struct {
	nodes nodeSelection

	redactLogs bool

	cpuProfDuration time.Duration

	concurrency int

	files fileSelection
}

func setZipContextDefaults() {
	__antithesis_instrumentation__.Notify(30363)
	zipCtx.nodes = nodeSelection{}
	zipCtx.files = fileSelection{}
	zipCtx.redactLogs = false
	zipCtx.cpuProfDuration = 5 * time.Second
	zipCtx.concurrency = 15

	now := timeutil.Now()
	zipCtx.files.startTimestamp = timestampValue(now.Add(-48 * time.Hour))
	zipCtx.files.endTimestamp = timestampValue(now.Add(24 * time.Hour))
}

var dumpCtx struct {
	dumpMode dumpMode

	asOf string

	dumpAll bool
}

func setDumpContextDefaults() {
	__antithesis_instrumentation__.Notify(30364)
	dumpCtx.dumpMode = dumpBoth
	dumpCtx.asOf = ""
	dumpCtx.dumpAll = false
}

var authCtx struct {
	onlyCookie     bool
	validityPeriod time.Duration
}

func setAuthContextDefaults() {
	__antithesis_instrumentation__.Notify(30365)
	authCtx.onlyCookie = false
	authCtx.validityPeriod = 1 * time.Hour
}

var debugCtx struct {
	startKey, endKey  storage.MVCCKey
	values            bool
	sizes             bool
	replicated        bool
	inputFile         string
	ballastSize       base.SizeSpec
	printSystemConfig bool
	maxResults        int
	decodeAsTableDesc string
	verbose           bool
	keyTypes          keyTypeFilter
}

func setDebugContextDefaults() {
	__antithesis_instrumentation__.Notify(30366)
	debugCtx.startKey = storage.NilKey
	debugCtx.endKey = storage.NilKey
	debugCtx.values = false
	debugCtx.sizes = false
	debugCtx.replicated = false
	debugCtx.inputFile = ""
	debugCtx.ballastSize = base.SizeSpec{InBytes: 1000000000}
	debugCtx.maxResults = 0
	debugCtx.printSystemConfig = false
	debugCtx.decodeAsTableDesc = ""
	debugCtx.verbose = false
	debugCtx.keyTypes = showAll
}

var startCtx struct {
	serverInsecure         bool
	serverSSLCertsDir      string
	serverCertPrincipalMap []string
	serverListenAddr       string

	initToken             string
	numExpectedNodes      int
	genCertsForSingleNode bool

	unencryptedLocalhostHTTP bool

	tempDir string

	externalIODir string

	inBackground bool

	listeningURLFile string

	pidFile string

	geoLibsDir string
}

func setStartContextDefaults() {
	__antithesis_instrumentation__.Notify(30367)
	startCtx.serverInsecure = baseCfg.Insecure
	startCtx.serverSSLCertsDir = base.DefaultCertsDirectory
	startCtx.serverCertPrincipalMap = nil
	startCtx.serverListenAddr = ""
	startCtx.initToken = ""
	startCtx.numExpectedNodes = 0
	startCtx.genCertsForSingleNode = false
	startCtx.unencryptedLocalhostHTTP = false
	startCtx.tempDir = ""
	startCtx.externalIODir = ""
	startCtx.listeningURLFile = ""
	startCtx.pidFile = ""
	startCtx.inBackground = false
	startCtx.geoLibsDir = "/usr/local/lib/cockroach"
}

var quitCtx struct {
	drainWait time.Duration

	nodeDrainSelf bool
}

func setQuitContextDefaults() {
	__antithesis_instrumentation__.Notify(30368)
	quitCtx.drainWait = 10 * time.Minute
	quitCtx.nodeDrainSelf = false
}

var nodeCtx struct {
	nodeDecommissionWait   nodeDecommissionWaitType
	nodeDecommissionSelf   bool
	statusShowRanges       bool
	statusShowStats        bool
	statusShowDecommission bool
	statusShowAll          bool
}

func setNodeContextDefaults() {
	__antithesis_instrumentation__.Notify(30369)
	nodeCtx.nodeDecommissionWait = nodeDecommissionWaitAll
	nodeCtx.nodeDecommissionSelf = false
	nodeCtx.statusShowRanges = false
	nodeCtx.statusShowStats = false
	nodeCtx.statusShowAll = false
	nodeCtx.statusShowDecommission = false
}

var sqlfmtCtx struct {
	len        int
	useSpaces  bool
	tabWidth   int
	noSimplify bool
	align      bool
	execStmts  clisqlshell.StatementsValue
}

func setSqlfmtContextDefaults() {
	__antithesis_instrumentation__.Notify(30370)
	cfg := tree.DefaultPrettyCfg()
	sqlfmtCtx.len = cfg.LineWidth
	sqlfmtCtx.useSpaces = !cfg.UseTabs
	sqlfmtCtx.tabWidth = cfg.TabWidth
	sqlfmtCtx.noSimplify = !cfg.Simplify
	sqlfmtCtx.align = (cfg.Align != tree.PrettyNoAlign)
	sqlfmtCtx.execStmts = nil
}

var convertCtx struct {
	url string
}

func setConvContextDefaults() {
	__antithesis_instrumentation__.Notify(30371)
	convertCtx.url = ""
}

var demoCtx = democluster.Context{
	CliCtx: &cliCtx.Context,
}

func setDemoContextDefaults() {
	__antithesis_instrumentation__.Notify(30372)
	demoCtx.NumNodes = 1
	demoCtx.SQLPoolMemorySize = 128 << 20
	demoCtx.CacheSize = 64 << 20
	demoCtx.NoExampleDatabase = false
	demoCtx.SimulateLatency = false
	demoCtx.RunWorkload = false
	demoCtx.Localities = nil
	demoCtx.GeoPartitionedReplicas = false
	demoCtx.DisableTelemetry = false
	demoCtx.DisableLicenseAcquisition = false
	demoCtx.DefaultKeySize = defaultKeySize
	demoCtx.DefaultCALifetime = defaultCALifetime
	demoCtx.DefaultCertLifetime = defaultCertLifetime
	demoCtx.Insecure = false
	demoCtx.SQLPort, _ = strconv.Atoi(base.DefaultPort)
	demoCtx.HTTPPort, _ = strconv.Atoi(base.DefaultHTTPPort)
	demoCtx.WorkloadMaxQPS = 25
}

var stmtDiagCtx struct {
	all bool
}

func setStmtDiagContextDefaults() {
	__antithesis_instrumentation__.Notify(30373)
	stmtDiagCtx.all = false
}

var importCtx struct {
	maxRowSize           int
	skipForeignKeys      bool
	ignoreUnsupported    bool
	ignoreUnsupportedLog string
	rowLimit             int
}

func setImportContextDefaults() {
	__antithesis_instrumentation__.Notify(30374)
	importCtx.maxRowSize = 512 * (1 << 10)
	importCtx.skipForeignKeys = false
	importCtx.ignoreUnsupported = false
	importCtx.ignoreUnsupportedLog = ""
	importCtx.rowLimit = 0
}

var proxyContext sqlproxyccl.ProxyOptions

func setProxyContextDefaults() {
	__antithesis_instrumentation__.Notify(30375)
	proxyContext.Denylist = ""
	proxyContext.ListenAddr = "127.0.0.1:46257"
	proxyContext.ListenCert = ""
	proxyContext.ListenKey = ""
	proxyContext.MetricsAddress = "0.0.0.0:8080"
	proxyContext.RoutingRule = ""
	proxyContext.DirectoryAddr = ""
	proxyContext.SkipVerify = false
	proxyContext.Insecure = false
	proxyContext.RatelimitBaseDelay = 50 * time.Millisecond
	proxyContext.ValidateAccessInterval = 30 * time.Second
	proxyContext.PollConfigInterval = 30 * time.Second
	proxyContext.DrainTimeout = 0
	proxyContext.ThrottleBaseDelay = time.Second
}

var testDirectorySvrContext struct {
	port          int
	certsDir      string
	kvAddrs       string
	tenantBaseDir string
}

func setTestDirectorySvrContextDefaults() {
	__antithesis_instrumentation__.Notify(30376)
	testDirectorySvrContext.port = 36257
}

var userfileCtx struct {
	recursive bool
}

func setUserfileContextDefaults() {
	__antithesis_instrumentation__.Notify(30377)
	userfileCtx.recursive = false
}

func GetServerCfgStores() base.StoreSpecList {
	__antithesis_instrumentation__.Notify(30378)
	return serverCfg.Stores
}
