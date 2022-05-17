package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"flag"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log/logflags"
	"github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var serverListenPort, serverSocketDir string
var serverAdvertiseAddr, serverAdvertisePort string
var serverSQLAddr, serverSQLPort string
var serverSQLAdvertiseAddr, serverSQLAdvertisePort string
var serverHTTPAddr, serverHTTPPort string
var localityAdvertiseHosts localityList
var startBackground bool
var storeSpecs base.StoreSpecList

func initPreFlagsDefaults() {
	__antithesis_instrumentation__.Notify(32680)
	serverListenPort = base.DefaultPort
	serverSocketDir = ""
	serverAdvertiseAddr = ""
	serverAdvertisePort = ""

	serverSQLAddr = ""
	serverSQLPort = ""
	serverSQLAdvertiseAddr = ""
	serverSQLAdvertisePort = ""

	serverHTTPAddr = ""
	serverHTTPPort = base.DefaultHTTPPort

	localityAdvertiseHosts = localityList{}

	startBackground = false

	storeSpecs = base.StoreSpecList{}
}

func AddPersistentPreRunE(cmd *cobra.Command, fn func(*cobra.Command, []string) error) {
	__antithesis_instrumentation__.Notify(32681)

	wrapped := cmd.PersistentPreRunE

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(32682)

		if wrapped != nil {
			__antithesis_instrumentation__.Notify(32684)
			if err := wrapped(cmd, args); err != nil {
				__antithesis_instrumentation__.Notify(32685)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32686)
			}
		} else {
			__antithesis_instrumentation__.Notify(32687)
		}
		__antithesis_instrumentation__.Notify(32683)

		return fn(cmd, args)
	}
}

func stringFlag(f *pflag.FlagSet, valPtr *string, flagInfo cliflags.FlagInfo) {
	__antithesis_instrumentation__.Notify(32688)
	f.StringVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

func intFlag(f *pflag.FlagSet, valPtr *int, flagInfo cliflags.FlagInfo) {
	__antithesis_instrumentation__.Notify(32689)
	f.IntVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

func boolFlag(f *pflag.FlagSet, valPtr *bool, flagInfo cliflags.FlagInfo) {
	__antithesis_instrumentation__.Notify(32690)
	f.BoolVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

func durationFlag(f *pflag.FlagSet, valPtr *time.Duration, flagInfo cliflags.FlagInfo) {
	__antithesis_instrumentation__.Notify(32691)
	f.DurationVarP(valPtr, flagInfo.Name, flagInfo.Shorthand, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

func varFlag(f *pflag.FlagSet, value pflag.Value, flagInfo cliflags.FlagInfo) {
	__antithesis_instrumentation__.Notify(32692)
	f.VarP(value, flagInfo.Name, flagInfo.Shorthand, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

func stringSliceFlag(f *pflag.FlagSet, valPtr *[]string, flagInfo cliflags.FlagInfo) {
	__antithesis_instrumentation__.Notify(32693)
	f.StringSliceVar(valPtr, flagInfo.Name, *valPtr, flagInfo.Usage())
	registerEnvVarDefault(f, flagInfo)
}

type aliasStrVar struct{ p *string }

func (a aliasStrVar) String() string { __antithesis_instrumentation__.Notify(32694); return "" }

func (a aliasStrVar) Set(v string) error {
	__antithesis_instrumentation__.Notify(32695)
	if v != "" {
		__antithesis_instrumentation__.Notify(32697)
		*a.p = v
	} else {
		__antithesis_instrumentation__.Notify(32698)
	}
	__antithesis_instrumentation__.Notify(32696)
	return nil
}

func (a aliasStrVar) Type() string { __antithesis_instrumentation__.Notify(32699); return "string" }

type addrSetter struct {
	addr *string
	port *string
}

func (a addrSetter) String() string {
	__antithesis_instrumentation__.Notify(32700)
	return net.JoinHostPort(*a.addr, *a.port)
}

func (a addrSetter) Type() string {
	__antithesis_instrumentation__.Notify(32701)
	return "<addr/host>[:<port>]"
}

func (a addrSetter) Set(v string) error {
	__antithesis_instrumentation__.Notify(32702)
	addr, port, err := addr.SplitHostPort(v, *a.port)
	if err != nil {
		__antithesis_instrumentation__.Notify(32704)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32705)
	}
	__antithesis_instrumentation__.Notify(32703)
	*a.addr = addr
	*a.port = port
	return nil
}

type clusterNameSetter struct {
	clusterName *string
}

func (a clusterNameSetter) String() string {
	__antithesis_instrumentation__.Notify(32706)
	return *a.clusterName
}

func (a clusterNameSetter) Type() string {
	__antithesis_instrumentation__.Notify(32707)
	return "<identifier>"
}

func (a clusterNameSetter) Set(v string) error {
	__antithesis_instrumentation__.Notify(32708)
	if v == "" {
		__antithesis_instrumentation__.Notify(32712)
		return errors.New("cluster name cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(32713)
	}
	__antithesis_instrumentation__.Notify(32709)
	if len(v) > maxClusterNameLength {
		__antithesis_instrumentation__.Notify(32714)
		return errors.Newf(`cluster name can contain at most %d characters`, maxClusterNameLength)
	} else {
		__antithesis_instrumentation__.Notify(32715)
	}
	__antithesis_instrumentation__.Notify(32710)
	if !clusterNameRe.MatchString(v) {
		__antithesis_instrumentation__.Notify(32716)
		return errClusterNameInvalidFormat
	} else {
		__antithesis_instrumentation__.Notify(32717)
	}
	__antithesis_instrumentation__.Notify(32711)
	*a.clusterName = v
	return nil
}

var errClusterNameInvalidFormat = errors.New(`cluster name must contain only letters, numbers or the "-" and "." characters`)

var clusterNameRe = regexp.MustCompile(`^[a-zA-Z](?:[-a-zA-Z0-9]*[a-zA-Z0-9]|)$`)

const maxClusterNameLength = 256

type keyTypeFilter int8

const (
	showAll keyTypeFilter = iota
	showValues
	showIntents
	showTxns
)

func (f *keyTypeFilter) String() string {
	__antithesis_instrumentation__.Notify(32718)
	switch *f {
	case showValues:
		__antithesis_instrumentation__.Notify(32720)
		return "values"
	case showIntents:
		__antithesis_instrumentation__.Notify(32721)
		return "intents"
	case showTxns:
		__antithesis_instrumentation__.Notify(32722)
		return "txns"
	default:
		__antithesis_instrumentation__.Notify(32723)
	}
	__antithesis_instrumentation__.Notify(32719)
	return "all"
}

func (f *keyTypeFilter) Type() string {
	__antithesis_instrumentation__.Notify(32724)
	return "<key type>"
}

func (f *keyTypeFilter) Set(v string) error {
	__antithesis_instrumentation__.Notify(32725)
	switch v {
	case "values":
		__antithesis_instrumentation__.Notify(32727)
		*f = showValues
	case "intents":
		__antithesis_instrumentation__.Notify(32728)
		*f = showIntents
	case "txns":
		__antithesis_instrumentation__.Notify(32729)
		*f = showTxns
	default:
		__antithesis_instrumentation__.Notify(32730)
		return errors.Newf("invalid key filter type '%s'", v)
	}
	__antithesis_instrumentation__.Notify(32726)
	return nil
}

const backgroundEnvVar = "COCKROACH_BACKGROUND_RESTART"

func flagSetForCmd(cmd *cobra.Command) *pflag.FlagSet {
	__antithesis_instrumentation__.Notify(32731)
	_ = cmd.LocalFlags()
	return cmd.Flags()
}

func init() {
	initCLIDefaults()

	AddPersistentPreRunE(cockroachCmd, func(cmd *cobra.Command, _ []string) error {
		return extraClientFlagInit()
	})

	for _, cmd := range append(serverCmds, connectInitCmd, connectJoinCmd) {
		AddPersistentPreRunE(cmd, func(cmd *cobra.Command, _ []string) error {

			return extraServerFlagInit(cmd)
		})
		AddPersistentPreRunE(cmd, func(cmd *cobra.Command, _ []string) error {
			return extraStoreFlagInit(cmd)
		})
	}

	AddPersistentPreRunE(DebugPebbleCmd, func(cmd *cobra.Command, _ []string) error {
		return extraStoreFlagInit(cmd)
	})

	AddPersistentPreRunE(mtStartSQLCmd, func(cmd *cobra.Command, _ []string) error {
		return mtStartSQLFlagsInit(cmd)
	})

	pf := cockroachCmd.PersistentFlags()
	flag.VisitAll(func(f *flag.Flag) {
		flag := pflag.PFlagFromGoFlag(f)

		if strings.HasPrefix(flag.Name, "lightstep_") {
			flag.Hidden = true
		}
		if strings.HasPrefix(flag.Name, "httptest.") {

			flag.Hidden = true
		}
		if strings.HasPrefix(flag.Name, "datadriven-") {

			flag.Hidden = true
		}
		if strings.EqualFold(flag.Name, "log_err_stacks") {

			flag.Hidden = true
		}
		if flag.Name == logflags.ShowLogsName || flag.Name == logflags.TestLogConfigName {

			flag.Hidden = true
		}
		pf.AddFlag(flag)
	})

	{

		var unused bool
		pf.BoolVarP(&unused, "verbose", "v", false, "")
		_ = pf.MarkHidden("verbose")
	}

	{

		varFlag(pf, &stringValue{settableString: &cliCtx.logConfigInput}, cliflags.Log)
		varFlag(pf, &fileContentsValue{settableString: &cliCtx.logConfigInput, fileName: "<unset>"}, cliflags.LogConfigFile)

		varFlag(pf, &cliCtx.deprecatedLogOverrides.stderrThreshold, cliflags.DeprecatedStderrThreshold)
		_ = pf.MarkDeprecated(cliflags.DeprecatedStderrThreshold.Name,
			"use --"+cliflags.Log.Name+" instead to specify 'sinks: {stderr: {filter: ...}}'.")

		pf.Lookup(cliflags.DeprecatedStderrThreshold.Name).NoOptDefVal = "DEFAULT"

		varFlag(pf, &cliCtx.deprecatedLogOverrides.stderrNoColor, cliflags.DeprecatedStderrNoColor)
		_ = pf.MarkDeprecated(cliflags.DeprecatedStderrNoColor.Name,
			"use --"+cliflags.Log.Name+" instead to specify 'sinks: {stderr: {no-color: true}}'")

		varFlag(pf, &stringValue{&cliCtx.deprecatedLogOverrides.logDir}, cliflags.DeprecatedLogDir)
		_ = pf.MarkDeprecated(cliflags.DeprecatedLogDir.Name,
			"use --"+cliflags.Log.Name+" instead to specify 'file-defaults: {dir: ...}'")

		varFlag(pf, cliCtx.deprecatedLogOverrides.fileMaxSizeVal, cliflags.DeprecatedLogFileMaxSize)
		_ = pf.MarkDeprecated(cliflags.DeprecatedLogFileMaxSize.Name,
			"use --"+cliflags.Log.Name+" instead to specify 'file-defaults: {max-file-size: ...}'")

		varFlag(pf, cliCtx.deprecatedLogOverrides.maxGroupSizeVal, cliflags.DeprecatedLogGroupMaxSize)
		_ = pf.MarkDeprecated(cliflags.DeprecatedLogGroupMaxSize.Name,
			"use --"+cliflags.Log.Name+" instead to specify 'file-defaults: {max-group-size: ...}'")

		varFlag(pf, &cliCtx.deprecatedLogOverrides.fileThreshold, cliflags.DeprecatedFileThreshold)
		_ = pf.MarkDeprecated(cliflags.DeprecatedFileThreshold.Name,
			"use --"+cliflags.Log.Name+" instead to specify 'file-defaults: {filter: ...}'")

		varFlag(pf, &cliCtx.deprecatedLogOverrides.redactableLogs, cliflags.DeprecatedRedactableLogs)
		_ = pf.MarkDeprecated(cliflags.DeprecatedRedactableLogs.Name,
			"use --"+cliflags.Log.Name+" instead to specify 'file-defaults: {redactable: ...}")

		varFlag(pf, &stringValue{&cliCtx.deprecatedLogOverrides.sqlAuditLogDir}, cliflags.DeprecatedSQLAuditLogDir)
		_ = pf.MarkDeprecated(cliflags.DeprecatedSQLAuditLogDir.Name,
			"use --"+cliflags.Log.Name+" instead to specify 'sinks: {file-groups: {sql-audit: {channels: SENSITIVE_ACCESS, dir: ...}}}")
	}

	_, startCtx.inBackground = envutil.EnvString(backgroundEnvVar, 1)

	for _, cmd := range append(StartCmds, connectInitCmd, connectJoinCmd) {
		f := cmd.Flags()

		varFlag(f, addrSetter{&startCtx.serverListenAddr, &serverListenPort}, cliflags.ListenAddr)
		varFlag(f, addrSetter{&serverAdvertiseAddr, &serverAdvertisePort}, cliflags.AdvertiseAddr)
		varFlag(f, addrSetter{&serverSQLAddr, &serverSQLPort}, cliflags.ListenSQLAddr)
		varFlag(f, addrSetter{&serverSQLAdvertiseAddr, &serverSQLAdvertisePort}, cliflags.SQLAdvertiseAddr)
		varFlag(f, addrSetter{&serverHTTPAddr, &serverHTTPPort}, cliflags.ListenHTTPAddr)

		stringFlag(f, &startCtx.serverSSLCertsDir, cliflags.ServerCertsDir)

		varFlag(f, &serverCfg.JoinList, cliflags.Join)
		boolFlag(f, &serverCfg.JoinPreferSRVRecords, cliflags.JoinPreferSRVRecords)
	}

	for _, cmd := range append(StartCmds, connectInitCmd) {
		f := cmd.Flags()

		stringFlag(f, &startCtx.initToken, cliflags.InitToken)
		intFlag(f, &startCtx.numExpectedNodes, cliflags.NumExpectedInitialNodes)
		boolFlag(f, &startCtx.genCertsForSingleNode, cliflags.SingleNode)

		if cmd == startSingleNodeCmd {

			_ = f.MarkHidden(cliflags.Join.Name)
			_ = f.MarkHidden(cliflags.JoinPreferSRVRecords.Name)
			_ = f.MarkHidden(cliflags.AdvertiseAddr.Name)
			_ = f.MarkHidden(cliflags.SQLAdvertiseAddr.Name)
			_ = f.MarkHidden(cliflags.InitToken.Name)
		}

		varFlag(f, aliasStrVar{&startCtx.serverListenAddr}, cliflags.ServerHost)
		_ = f.MarkHidden(cliflags.ServerHost.Name)
		varFlag(f, aliasStrVar{&serverListenPort}, cliflags.ServerPort)
		_ = f.MarkHidden(cliflags.ServerPort.Name)

		varFlag(f, aliasStrVar{&serverAdvertiseAddr}, cliflags.AdvertiseHost)
		_ = f.MarkHidden(cliflags.AdvertiseHost.Name)
		varFlag(f, aliasStrVar{&serverAdvertisePort}, cliflags.AdvertisePort)
		_ = f.MarkHidden(cliflags.AdvertisePort.Name)

		varFlag(f, aliasStrVar{&serverHTTPAddr}, cliflags.ListenHTTPAddrAlias)
		_ = f.MarkHidden(cliflags.ListenHTTPAddrAlias.Name)
		varFlag(f, aliasStrVar{&serverHTTPPort}, cliflags.ListenHTTPPort)
		_ = f.MarkHidden(cliflags.ListenHTTPPort.Name)

	}

	for _, cmd := range StartCmds {
		f := cmd.Flags()

		stringFlag(f, &serverSocketDir, cliflags.SocketDir)
		boolFlag(f, &startCtx.unencryptedLocalhostHTTP, cliflags.UnencryptedLocalhostHTTP)

		boolFlag(f, &serverCfg.AcceptSQLWithoutTLS, cliflags.AcceptSQLWithoutTLS)
		_ = f.MarkHidden(cliflags.AcceptSQLWithoutTLS.Name)

		varFlag(f, &localityAdvertiseHosts, cliflags.LocalityAdvertiseAddr)

		stringFlag(f, &serverCfg.Attrs, cliflags.Attrs)
		varFlag(f, &serverCfg.Locality, cliflags.Locality)

		varFlag(f, &storeSpecs, cliflags.Store)
		varFlag(f, &serverCfg.StorageEngine, cliflags.StorageEngine)
		varFlag(f, &serverCfg.MaxOffset, cliflags.MaxOffset)
		stringFlag(f, &serverCfg.ClockDevicePath, cliflags.ClockDevice)

		stringFlag(f, &startCtx.listeningURLFile, cliflags.ListeningURLFile)

		stringFlag(f, &startCtx.pidFile, cliflags.PIDFile)
		stringFlag(f, &startCtx.geoLibsDir, cliflags.GeoLibsDir)

		boolFlag(f, &startCtx.serverInsecure, cliflags.ServerInsecure)

		boolFlag(f, &serverCfg.ExternalIODirConfig.DisableHTTP, cliflags.ExternalIODisableHTTP)
		boolFlag(f, &serverCfg.ExternalIODirConfig.DisableOutbound, cliflags.ExternalIODisabled)
		boolFlag(f, &serverCfg.ExternalIODirConfig.DisableImplicitCredentials, cliflags.ExternalIODisableImplicitCredentials)
		boolFlag(f, &serverCfg.ExternalIODirConfig.EnableNonAdminImplicitAndArbitraryOutbound, cliflags.ExternalIOEnableNonAdminImplicitAndArbitraryOutbound)

		stringSliceFlag(f, &startCtx.serverCertPrincipalMap, cliflags.CertPrincipalMap)

		varFlag(f, clusterNameSetter{&baseCfg.ClusterName}, cliflags.ClusterName)
		boolFlag(f, &baseCfg.DisableClusterNameVerification, cliflags.DisableClusterNameVerification)
		if cmd == startSingleNodeCmd {

			_ = f.MarkHidden(cliflags.ClusterName.Name)
			_ = f.MarkHidden(cliflags.DisableClusterNameVerification.Name)
			_ = f.MarkHidden(cliflags.MaxOffset.Name)
			_ = f.MarkHidden(cliflags.LocalityAdvertiseAddr.Name)
		}

		varFlag(f, cacheSizeValue, cliflags.Cache)
		varFlag(f, sqlSizeValue, cliflags.SQLMem)
		varFlag(f, tsdbSizeValue, cliflags.TSDBMem)

		varFlag(f, diskTempStorageSizeValue, cliflags.SQLTempStorage)
		stringFlag(f, &startCtx.tempDir, cliflags.TempDir)
		stringFlag(f, &startCtx.externalIODir, cliflags.ExternalIODir)

		if backgroundFlagDefined {
			boolFlag(f, &startBackground, cliflags.Background)
		}
	}

	telemetryEnabledCmds := append(serverCmds, demoCmd)
	telemetryEnabledCmds = append(telemetryEnabledCmds, demoCmd.Commands()...)
	for _, cmd := range telemetryEnabledCmds {

		AddPersistentPreRunE(cmd, func(cmd *cobra.Command, _ []string) error {
			prefix := "cli." + cmd.Name()

			cmd.Flags().Visit(func(fl *pflag.Flag) {
				telemetry.Count(prefix + ".explicitflags." + fl.Name)
			})

			telemetry.Count(prefix + ".runs")
			return nil
		})
	}

	for _, cmd := range certCmds {
		f := cmd.Flags()

		stringFlag(f, &certCtx.certsDir, cliflags.CertsDir)

		stringSliceFlag(f, &certCtx.certPrincipalMap, cliflags.CertPrincipalMap)

		if cmd == listCertsCmd {

			continue
		}

		stringFlag(f, &certCtx.caKey, cliflags.CAKey)
		intFlag(f, &certCtx.keySize, cliflags.KeySize)
		boolFlag(f, &certCtx.overwriteFiles, cliflags.OverwriteFiles)

		if strings.HasSuffix(cmd.Name(), "-ca") {

			durationFlag(f, &certCtx.caCertificateLifetime, cliflags.CertificateLifetime)

			boolFlag(f, &certCtx.allowCAKeyReuse, cliflags.AllowCAKeyReuse)
		} else {

			durationFlag(f, &certCtx.certificateLifetime, cliflags.CertificateLifetime)
		}

		if cmd == createClientCertCmd {
			boolFlag(f, &certCtx.generatePKCS8Key, cliflags.GeneratePKCS8Key)
		}
	}

	clientCmds := []*cobra.Command{
		debugJobTraceFromClusterCmd,
		debugGossipValuesCmd,
		debugTimeSeriesDumpCmd,
		debugZipCmd,
		debugListFilesCmd,
		debugSendKVBatchCmd,
		doctorExamineClusterCmd,
		doctorExamineFallbackClusterCmd,
		doctorRecreateClusterCmd,
		genHAProxyCmd,
		initCmd,
		quitCmd,
		sqlShellCmd,
	}
	clientCmds = append(clientCmds, authCmds...)
	clientCmds = append(clientCmds, nodeCmds...)
	clientCmds = append(clientCmds, nodeLocalCmds...)
	clientCmds = append(clientCmds, importCmds...)
	clientCmds = append(clientCmds, userFileCmds...)
	clientCmds = append(clientCmds, stmtDiagCmds...)
	clientCmds = append(clientCmds, debugResetQuorumCmd)
	for _, cmd := range clientCmds {
		f := cmd.PersistentFlags()
		varFlag(f, addrSetter{&cliCtx.clientConnHost, &cliCtx.clientConnPort}, cliflags.ClientHost)
		stringFlag(f, &cliCtx.clientConnPort, cliflags.ClientPort)
		_ = f.MarkHidden(cliflags.ClientPort.Name)

		boolFlag(f, &baseCfg.Insecure, cliflags.ClientInsecure)

		stringFlag(f, &baseCfg.SSLCertsDir, cliflags.CertsDir)

		stringSliceFlag(f, &cliCtx.certPrincipalMap, cliflags.CertPrincipalMap)
	}

	{
		f := convertURLCmd.PersistentFlags()
		stringFlag(f, &convertCtx.url, cliflags.URL)
	}

	{
		f := loginCmd.Flags()
		durationFlag(f, &authCtx.validityPeriod, cliflags.AuthTokenValidityPeriod)
		boolFlag(f, &authCtx.onlyCookie, cliflags.OnlyCookie)
	}

	timeoutCmds := []*cobra.Command{
		statusNodeCmd,
		lsNodesCmd,
		debugJobTraceFromClusterCmd,
		debugZipCmd,
		doctorExamineClusterCmd,
		doctorExamineFallbackClusterCmd,
		doctorRecreateClusterCmd,
	}

	for _, cmd := range timeoutCmds {
		durationFlag(cmd.Flags(), &cliCtx.cmdTimeout, cliflags.Timeout)
	}

	{
		f := statusNodeCmd.Flags()
		boolFlag(f, &nodeCtx.statusShowRanges, cliflags.NodeRanges)
		boolFlag(f, &nodeCtx.statusShowStats, cliflags.NodeStats)
		boolFlag(f, &nodeCtx.statusShowAll, cliflags.NodeAll)
		boolFlag(f, &nodeCtx.statusShowDecommission, cliflags.NodeDecommission)
	}

	{
		f := debugZipCmd.Flags()
		boolFlag(f, &zipCtx.redactLogs, cliflags.ZipRedactLogs)
		durationFlag(f, &zipCtx.cpuProfDuration, cliflags.ZipCPUProfileDuration)
		intFlag(f, &zipCtx.concurrency, cliflags.ZipConcurrency)
	}

	for _, cmd := range []*cobra.Command{debugZipCmd, debugListFilesCmd} {
		f := cmd.Flags()
		varFlag(f, &zipCtx.nodes.inclusive, cliflags.ZipNodes)
		varFlag(f, &zipCtx.nodes.exclusive, cliflags.ZipExcludeNodes)
		stringSliceFlag(f, &zipCtx.files.includePatterns, cliflags.ZipIncludedFiles)
		stringSliceFlag(f, &zipCtx.files.excludePatterns, cliflags.ZipExcludedFiles)
		varFlag(f, &zipCtx.files.startTimestamp, cliflags.ZipFilesFrom)
		varFlag(f, &zipCtx.files.endTimestamp, cliflags.ZipFilesUntil)
	}

	varFlag(decommissionNodeCmd.Flags(), &nodeCtx.nodeDecommissionWait, cliflags.Wait)

	for _, cmd := range []*cobra.Command{decommissionNodeCmd, recommissionNodeCmd} {
		f := cmd.Flags()
		boolFlag(f, &nodeCtx.nodeDecommissionSelf, cliflags.NodeDecommissionSelf)
	}

	for _, cmd := range []*cobra.Command{quitCmd, drainNodeCmd} {
		f := cmd.Flags()
		durationFlag(f, &quitCtx.drainWait, cliflags.DrainWait)
		boolFlag(f, &quitCtx.nodeDrainSelf, cliflags.NodeDrainSelf)
	}

	for _, cmd := range append([]*cobra.Command{sqlShellCmd, demoCmd}, demoCmd.Commands()...) {
		f := cmd.Flags()
		varFlag(f, &sqlCtx.ShellCtx.SetStmts, cliflags.Set)
		varFlag(f, &sqlCtx.ShellCtx.ExecStmts, cliflags.Execute)
		stringFlag(f, &sqlCtx.InputFile, cliflags.File)
		durationFlag(f, &sqlCtx.ShellCtx.RepeatDelay, cliflags.Watch)
		varFlag(f, &sqlCtx.SafeUpdates, cliflags.SafeUpdates)
		boolFlag(f, &sqlCtx.ReadOnly, cliflags.ReadOnly)

		f.Lookup(cliflags.SafeUpdates.Name).NoOptDefVal = "true"
		boolFlag(f, &sqlConnCtx.DebugMode, cliflags.CliDebugMode)
		boolFlag(f, &cliCtx.EmbeddedMode, cliflags.EmbeddedMode)
	}

	sqlCmds := []*cobra.Command{
		sqlShellCmd,
		demoCmd,
		debugJobTraceFromClusterCmd,
		doctorExamineClusterCmd,
		doctorExamineFallbackClusterCmd,
		doctorRecreateClusterCmd,
		statementBundleRecreateCmd,
		lsNodesCmd,
		statusNodeCmd,
	}
	sqlCmds = append(sqlCmds, authCmds...)
	sqlCmds = append(sqlCmds, demoCmd.Commands()...)
	sqlCmds = append(sqlCmds, stmtDiagCmds...)
	sqlCmds = append(sqlCmds, nodeLocalCmds...)
	sqlCmds = append(sqlCmds, importCmds...)
	sqlCmds = append(sqlCmds, userFileCmds...)
	for _, cmd := range sqlCmds {
		f := cmd.Flags()

		boolFlag(f, &sqlConnCtx.Echo, cliflags.EchoSQL)

		varFlag(f, urlParser{cmd, &cliCtx, false}, cliflags.URL)
		stringFlag(f, &cliCtx.sqlConnUser, cliflags.User)
		if cmd == demoCmd {

			_ = f.MarkHidden(cliflags.URL.Name)
			_ = f.MarkHidden(cliflags.User.Name)
		}

		varFlag(f, addrSetter{&cliCtx.clientConnHost, &cliCtx.clientConnPort}, cliflags.ClientHost)
		_ = f.MarkHidden(cliflags.ClientHost.Name)
		stringFlag(f, &cliCtx.clientConnPort, cliflags.ClientPort)
		_ = f.MarkHidden(cliflags.ClientPort.Name)

		if cmd == sqlShellCmd || cmd == demoCmd {
			stringFlag(f, &cliCtx.sqlConnDBName, cliflags.Database)
			if cmd == demoCmd {

				_ = f.MarkHidden(cliflags.Database.Name)
			}
		}
	}

	for _, cmd := range clientCmds {
		if fl := flagSetForCmd(cmd).Lookup(cliflags.URL.Name); fl != nil {

			continue
		}

		f := cmd.PersistentFlags()
		varFlag(f, urlParser{cmd, &cliCtx, true}, cliflags.URL)

		varFlag(f, clusterNameSetter{&baseCfg.ClusterName}, cliflags.ClusterName)
		boolFlag(f, &baseCfg.DisableClusterNameVerification, cliflags.DisableClusterNameVerification)
	}

	tableOutputCommands := append(
		[]*cobra.Command{
			sqlShellCmd,
			genSettingsListCmd,
			demoCmd,
			statementBundleRecreateCmd,
			debugListFilesCmd,
			debugJobTraceFromClusterCmd,
		},
		demoCmd.Commands()...)
	tableOutputCommands = append(tableOutputCommands, nodeCmds...)
	tableOutputCommands = append(tableOutputCommands, authCmds...)

	for _, cmd := range tableOutputCommands {
		f := cmd.PersistentFlags()
		varFlag(f, &sqlExecCtx.TableDisplayFormat, cliflags.TableDisplayFormat)
	}

	{

		f := demoCmd.PersistentFlags()

		intFlag(f, &demoCtx.NumNodes, cliflags.DemoNodes)
		boolFlag(f, &demoCtx.RunWorkload, cliflags.RunDemoWorkload)
		intFlag(f, &demoCtx.WorkloadMaxQPS, cliflags.DemoWorkloadMaxQPS)
		varFlag(f, &demoCtx.Localities, cliflags.DemoNodeLocality)
		boolFlag(f, &demoCtx.GeoPartitionedReplicas, cliflags.DemoGeoPartitionedReplicas)
		varFlag(f, demoNodeSQLMemSizeValue, cliflags.DemoNodeSQLMemSize)
		varFlag(f, demoNodeCacheSizeValue, cliflags.DemoNodeCacheSize)
		boolFlag(f, &demoCtx.Insecure, cliflags.ClientInsecure)

		_ = f.MarkDeprecated(cliflags.ServerInsecure.Name,
			"to start a test server without any security, run start-single-node --insecure\n"+
				"For details, see: "+build.MakeIssueURL(53404))

		boolFlag(f, &demoCtx.DisableLicenseAcquisition, cliflags.DemoNoLicense)

		boolFlag(f, &demoCtx.Multitenant, cliflags.DemoMultitenant)

		_ = f.MarkHidden(cliflags.DemoMultitenant.Name)

		boolFlag(f, &demoCtx.SimulateLatency, cliflags.Global)

		boolFlag(demoCmd.Flags(), &demoCtx.NoExampleDatabase, cliflags.UseEmptyDatabase)
		_ = f.MarkDeprecated(cliflags.UseEmptyDatabase.Name, "use --no-workload-database")
		boolFlag(demoCmd.Flags(), &demoCtx.NoExampleDatabase, cliflags.NoExampleDatabase)

		stringFlag(f, &startCtx.geoLibsDir, cliflags.GeoLibsDir)

		intFlag(f, &demoCtx.SQLPort, cliflags.DemoSQLPort)
		intFlag(f, &demoCtx.HTTPPort, cliflags.DemoHTTPPort)
		stringFlag(f, &demoCtx.ListeningURLFile, cliflags.ListeningURLFile)
	}

	{
		boolFlag(stmtDiagDeleteCmd.Flags(), &stmtDiagCtx.all, cliflags.StmtDiagDeleteAll)
		boolFlag(stmtDiagCancelCmd.Flags(), &stmtDiagCtx.all, cliflags.StmtDiagCancelAll)
	}

	{
		d := importDumpFileCmd.Flags()
		boolFlag(d, &importCtx.skipForeignKeys, cliflags.ImportSkipForeignKeys)
		intFlag(d, &importCtx.maxRowSize, cliflags.ImportMaxRowSize)
		intFlag(d, &importCtx.rowLimit, cliflags.ImportRowLimit)
		boolFlag(d, &importCtx.ignoreUnsupported, cliflags.ImportIgnoreUnsupportedStatements)
		stringFlag(d, &importCtx.ignoreUnsupportedLog, cliflags.ImportLogIgnoredStatements)
		stringFlag(d, &cliCtx.sqlConnDBName, cliflags.Database)

		t := importDumpTableCmd.Flags()
		boolFlag(t, &importCtx.skipForeignKeys, cliflags.ImportSkipForeignKeys)
		intFlag(t, &importCtx.maxRowSize, cliflags.ImportMaxRowSize)
		intFlag(t, &importCtx.rowLimit, cliflags.ImportRowLimit)
		boolFlag(t, &importCtx.ignoreUnsupported, cliflags.ImportIgnoreUnsupportedStatements)
		stringFlag(t, &importCtx.ignoreUnsupportedLog, cliflags.ImportLogIgnoredStatements)
		stringFlag(t, &cliCtx.sqlConnDBName, cliflags.Database)
	}

	{
		f := sqlfmtCmd.Flags()
		varFlag(f, &sqlfmtCtx.execStmts, cliflags.Execute)
		intFlag(f, &sqlfmtCtx.len, cliflags.SQLFmtLen)
		boolFlag(f, &sqlfmtCtx.useSpaces, cliflags.SQLFmtSpaces)
		intFlag(f, &sqlfmtCtx.tabWidth, cliflags.SQLFmtTabWidth)
		boolFlag(f, &sqlfmtCtx.noSimplify, cliflags.SQLFmtNoSimplify)
		boolFlag(f, &sqlfmtCtx.align, cliflags.SQLFmtAlign)
	}

	{
		f := versionCmd.Flags()
		boolFlag(f, &cliCtx.showVersionUsingOnlyBuildTag, cliflags.BuildTag)
	}

	{
		f := debugKeysCmd.Flags()
		varFlag(f, (*mvccKey)(&debugCtx.startKey), cliflags.From)
		varFlag(f, (*mvccKey)(&debugCtx.endKey), cliflags.To)
		intFlag(f, &debugCtx.maxResults, cliflags.Limit)
		boolFlag(f, &debugCtx.values, cliflags.Values)
		boolFlag(f, &debugCtx.sizes, cliflags.Sizes)
		stringFlag(f, &debugCtx.decodeAsTableDesc, cliflags.DecodeAsTable)
		varFlag(f, &debugCtx.keyTypes, cliflags.FilterKeys)
	}
	{
		f := debugCheckLogConfigCmd.Flags()
		varFlag(f, &storeSpecs, cliflags.Store)
	}
	{
		f := debugRangeDataCmd.Flags()
		boolFlag(f, &debugCtx.replicated, cliflags.Replicated)
		intFlag(f, &debugCtx.maxResults, cliflags.Limit)
	}
	{
		f := debugGossipValuesCmd.Flags()
		stringFlag(f, &debugCtx.inputFile, cliflags.GossipInputFile)
		boolFlag(f, &debugCtx.printSystemConfig, cliflags.PrintSystemConfig)
	}
	{
		f := debugBallastCmd.Flags()
		varFlag(f, &debugCtx.ballastSize, cliflags.Size)
	}
	{

		f := DebugPebbleCmd.PersistentFlags()
		varFlag(f, &storeSpecs, cliflags.Store)
	}
	{
		for _, c := range []*cobra.Command{
			debugJobTraceFromClusterCmd,
			doctorExamineClusterCmd,
			doctorExamineZipDirCmd,
			doctorExamineFallbackClusterCmd,
			doctorExamineFallbackZipDirCmd,
			doctorRecreateClusterCmd,
			doctorRecreateZipDirCmd,
		} {
			f := c.Flags()
			if f.Lookup(cliflags.Verbose.Name) == nil {
				boolFlag(f, &debugCtx.verbose, cliflags.Verbose)
			}
		}
	}

	{
		f := mtStartSQLCmd.Flags()
		varFlag(f, &tenantIDWrapper{&serverCfg.SQLConfig.TenantID}, cliflags.TenantID)

		_ = extraServerFlagInit

		boolFlag(f, &startCtx.serverInsecure, cliflags.ServerInsecure)

		stringFlag(f, &startCtx.serverSSLCertsDir, cliflags.ServerCertsDir)

		varFlag(f, addrSetter{&serverSQLAddr, &serverSQLPort}, cliflags.ListenSQLAddr)
		varFlag(f, addrSetter{&serverHTTPAddr, &serverHTTPPort}, cliflags.ListenHTTPAddr)
		varFlag(f, addrSetter{&serverAdvertiseAddr, &serverAdvertisePort}, cliflags.AdvertiseAddr)

		varFlag(f, &serverCfg.Locality, cliflags.Locality)
		varFlag(f, &serverCfg.MaxOffset, cliflags.MaxOffset)
		varFlag(f, &storeSpecs, cliflags.Store)
		stringFlag(f, &startCtx.pidFile, cliflags.PIDFile)
		stringFlag(f, &startCtx.geoLibsDir, cliflags.GeoLibsDir)

		stringSliceFlag(f, &serverCfg.SQLConfig.TenantKVAddrs, cliflags.KVAddrs)

		boolFlag(f, &serverCfg.ExternalIODirConfig.DisableHTTP, cliflags.ExternalIODisableHTTP)
		boolFlag(f, &serverCfg.ExternalIODirConfig.DisableOutbound, cliflags.ExternalIODisabled)
		boolFlag(f, &serverCfg.ExternalIODirConfig.DisableImplicitCredentials, cliflags.ExternalIODisableImplicitCredentials)

		varFlag(f, sqlSizeValue, cliflags.SQLMem)
		varFlag(f, tsdbSizeValue, cliflags.TSDBMem)

		varFlag(f, diskTempStorageSizeValue, cliflags.SQLTempStorage)
		stringFlag(f, &startCtx.tempDir, cliflags.TempDir)

		if backgroundFlagDefined {
			boolFlag(f, &startBackground, cliflags.Background)
		}
	}

	{
		f := mtStartSQLProxyCmd.Flags()
		stringFlag(f, &proxyContext.Denylist, cliflags.DenyList)
		stringFlag(f, &proxyContext.ListenAddr, cliflags.ProxyListenAddr)
		stringFlag(f, &proxyContext.ListenCert, cliflags.ListenCert)
		stringFlag(f, &proxyContext.ListenKey, cliflags.ListenKey)
		stringFlag(f, &proxyContext.MetricsAddress, cliflags.ListenMetrics)
		stringFlag(f, &proxyContext.RoutingRule, cliflags.RoutingRule)
		stringFlag(f, &proxyContext.DirectoryAddr, cliflags.DirectoryAddr)
		boolFlag(f, &proxyContext.SkipVerify, cliflags.SkipVerify)
		boolFlag(f, &proxyContext.Insecure, cliflags.InsecureBackend)
		durationFlag(f, &proxyContext.ValidateAccessInterval, cliflags.ValidateAccessInterval)
		durationFlag(f, &proxyContext.PollConfigInterval, cliflags.PollConfigInterval)
		durationFlag(f, &proxyContext.DrainTimeout, cliflags.DrainTimeout)
		durationFlag(f, &proxyContext.ThrottleBaseDelay, cliflags.ThrottleBaseDelay)
	}

	{
		f := mtTestDirectorySvr.Flags()
		intFlag(f, &testDirectorySvrContext.port, cliflags.TestDirectoryListenPort)
		stringFlag(f, &testDirectorySvrContext.certsDir, cliflags.TestDirectoryTenantCertsDir)
		stringFlag(f, &testDirectorySvrContext.tenantBaseDir, cliflags.TestDirectoryTenantBaseDir)
		stringFlag(f, &testDirectorySvrContext.kvAddrs, cliflags.KVAddrs)
	}

	{
		boolFlag(userFileUploadCmd.Flags(), &userfileCtx.recursive, cliflags.Recursive)
	}
}

type tenantIDWrapper struct {
	tenID *roachpb.TenantID
}

func (w *tenantIDWrapper) String() string {
	__antithesis_instrumentation__.Notify(32732)
	return w.tenID.String()
}
func (w *tenantIDWrapper) Set(s string) error {
	__antithesis_instrumentation__.Notify(32733)
	tenID, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(32736)
		return errors.Wrap(err, "invalid tenant ID")
	} else {
		__antithesis_instrumentation__.Notify(32737)
	}
	__antithesis_instrumentation__.Notify(32734)
	if tenID == 0 {
		__antithesis_instrumentation__.Notify(32738)
		return errors.New("invalid tenant ID")
	} else {
		__antithesis_instrumentation__.Notify(32739)
	}
	__antithesis_instrumentation__.Notify(32735)
	*w.tenID = roachpb.MakeTenantID(tenID)
	return nil
}

func (w *tenantIDWrapper) Type() string {
	__antithesis_instrumentation__.Notify(32740)
	return "number"
}

func processEnvVarDefaults(cmd *cobra.Command) error {
	__antithesis_instrumentation__.Notify(32741)
	fl := flagSetForCmd(cmd)

	var retErr error
	fl.VisitAll(func(f *pflag.Flag) {
		__antithesis_instrumentation__.Notify(32743)
		envv, ok := f.Annotations[envValueAnnotationKey]
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(32745)
			return len(envv) < 2 == true
		}() == true {
			__antithesis_instrumentation__.Notify(32746)

			return
		} else {
			__antithesis_instrumentation__.Notify(32747)
		}
		__antithesis_instrumentation__.Notify(32744)
		varName, value := envv[0], envv[1]
		if err := fl.Set(f.Name, value); err != nil {
			__antithesis_instrumentation__.Notify(32748)
			retErr = errors.CombineErrors(retErr,
				errors.Wrapf(err, "setting --%s from %s", f.Name, varName))
		} else {
			__antithesis_instrumentation__.Notify(32749)
		}
	})
	__antithesis_instrumentation__.Notify(32742)
	return retErr
}

const (
	envValueAnnotationKey = "envvalue"
)

func registerEnvVarDefault(f *pflag.FlagSet, flagInfo cliflags.FlagInfo) {
	__antithesis_instrumentation__.Notify(32750)
	if flagInfo.EnvVar == "" {
		__antithesis_instrumentation__.Notify(32753)
		return
	} else {
		__antithesis_instrumentation__.Notify(32754)
	}
	__antithesis_instrumentation__.Notify(32751)

	value, set := envutil.EnvString(flagInfo.EnvVar, 2)
	if !set {
		__antithesis_instrumentation__.Notify(32755)

		return
	} else {
		__antithesis_instrumentation__.Notify(32756)
	}
	__antithesis_instrumentation__.Notify(32752)

	if err := f.SetAnnotation(flagInfo.Name, envValueAnnotationKey, []string{flagInfo.EnvVar, value}); err != nil {
		__antithesis_instrumentation__.Notify(32757)

		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(32758)
	}
}

func extraServerFlagInit(cmd *cobra.Command) error {
	__antithesis_instrumentation__.Notify(32759)
	if err := security.SetCertPrincipalMap(startCtx.serverCertPrincipalMap); err != nil {
		__antithesis_instrumentation__.Notify(32772)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32773)
	}
	__antithesis_instrumentation__.Notify(32760)
	serverCfg.User = security.NodeUserName()
	serverCfg.Insecure = startCtx.serverInsecure
	serverCfg.SSLCertsDir = startCtx.serverSSLCertsDir

	serverCfg.Addr = net.JoinHostPort(startCtx.serverListenAddr, serverListenPort)

	fs := flagSetForCmd(cmd)

	changed := func(set *pflag.FlagSet, name string) bool {
		__antithesis_instrumentation__.Notify(32774)
		f := set.Lookup(name)
		return f != nil && func() bool {
			__antithesis_instrumentation__.Notify(32775)
			return f.Changed == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(32761)

	if changed(fs, cliflags.SocketDir.Name) {
		__antithesis_instrumentation__.Notify(32776)
		if serverSocketDir == "" {
			__antithesis_instrumentation__.Notify(32777)
			serverCfg.SocketFile = ""
		} else {
			__antithesis_instrumentation__.Notify(32778)
			serverCfg.SocketFile = filepath.Join(serverSocketDir, ".s.PGSQL."+serverListenPort)
		}
	} else {
		__antithesis_instrumentation__.Notify(32779)
	}
	__antithesis_instrumentation__.Notify(32762)

	if serverAdvertiseAddr == "" {
		__antithesis_instrumentation__.Notify(32780)
		serverAdvertiseAddr = startCtx.serverListenAddr
	} else {
		__antithesis_instrumentation__.Notify(32781)
	}
	__antithesis_instrumentation__.Notify(32763)
	if serverAdvertisePort == "" {
		__antithesis_instrumentation__.Notify(32782)
		serverAdvertisePort = serverListenPort
	} else {
		__antithesis_instrumentation__.Notify(32783)
	}
	__antithesis_instrumentation__.Notify(32764)
	serverCfg.AdvertiseAddr = net.JoinHostPort(serverAdvertiseAddr, serverAdvertisePort)

	if serverSQLAddr == "" {
		__antithesis_instrumentation__.Notify(32784)
		serverSQLAddr = startCtx.serverListenAddr
	} else {
		__antithesis_instrumentation__.Notify(32785)
	}
	__antithesis_instrumentation__.Notify(32765)
	if serverSQLPort == "" {
		__antithesis_instrumentation__.Notify(32786)
		serverSQLPort = serverListenPort
	} else {
		__antithesis_instrumentation__.Notify(32787)
	}
	__antithesis_instrumentation__.Notify(32766)
	serverCfg.SQLAddr = net.JoinHostPort(serverSQLAddr, serverSQLPort)
	serverCfg.SplitListenSQL = fs.Lookup(cliflags.ListenSQLAddr.Name).Changed

	advSpecified := changed(fs, cliflags.AdvertiseAddr.Name) || func() bool {
		__antithesis_instrumentation__.Notify(32788)
		return changed(fs, cliflags.AdvertiseHost.Name) == true
	}() == true
	if serverSQLAdvertiseAddr == "" {
		__antithesis_instrumentation__.Notify(32789)
		if advSpecified {
			__antithesis_instrumentation__.Notify(32790)
			serverSQLAdvertiseAddr = serverAdvertiseAddr
		} else {
			__antithesis_instrumentation__.Notify(32791)
			serverSQLAdvertiseAddr = serverSQLAddr
		}
	} else {
		__antithesis_instrumentation__.Notify(32792)
	}
	__antithesis_instrumentation__.Notify(32767)
	if serverSQLAdvertisePort == "" {
		__antithesis_instrumentation__.Notify(32793)
		if advSpecified && func() bool {
			__antithesis_instrumentation__.Notify(32794)
			return !serverCfg.SplitListenSQL == true
		}() == true {
			__antithesis_instrumentation__.Notify(32795)
			serverSQLAdvertisePort = serverAdvertisePort
		} else {
			__antithesis_instrumentation__.Notify(32796)
			serverSQLAdvertisePort = serverSQLPort
		}
	} else {
		__antithesis_instrumentation__.Notify(32797)
	}
	__antithesis_instrumentation__.Notify(32768)
	serverCfg.SQLAdvertiseAddr = net.JoinHostPort(serverSQLAdvertiseAddr, serverSQLAdvertisePort)

	if serverHTTPAddr == "" {
		__antithesis_instrumentation__.Notify(32798)
		serverHTTPAddr = startCtx.serverListenAddr
	} else {
		__antithesis_instrumentation__.Notify(32799)
	}
	__antithesis_instrumentation__.Notify(32769)
	if startCtx.unencryptedLocalhostHTTP {
		__antithesis_instrumentation__.Notify(32800)

		if (changed(fs, cliflags.ListenHTTPAddr.Name) || func() bool {
			__antithesis_instrumentation__.Notify(32802)
			return changed(fs, cliflags.ListenHTTPAddrAlias.Name) == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(32803)
			return (serverHTTPAddr != "" && func() bool {
				__antithesis_instrumentation__.Notify(32804)
				return serverHTTPAddr != "localhost" == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(32805)
			return errors.WithHintf(
				errors.Newf("--unencrypted-localhost-http is incompatible with --http-addr=%s:%s",
					serverHTTPAddr, serverHTTPPort),
				`When --unencrypted-localhost-http is specified, use --http-addr=:%s or omit --http-addr entirely.`, serverHTTPPort)
		} else {
			__antithesis_instrumentation__.Notify(32806)
		}
		__antithesis_instrumentation__.Notify(32801)

		serverHTTPAddr = "localhost"

		serverCfg.DisableTLSForHTTP = true
	} else {
		__antithesis_instrumentation__.Notify(32807)
	}
	__antithesis_instrumentation__.Notify(32770)
	serverCfg.HTTPAddr = net.JoinHostPort(serverHTTPAddr, serverHTTPPort)

	for i, a := range localityAdvertiseHosts {
		__antithesis_instrumentation__.Notify(32808)
		host, port, err := addr.SplitHostPort(a.Address.AddressField, serverAdvertisePort)
		if err != nil {
			__antithesis_instrumentation__.Notify(32810)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32811)
		}
		__antithesis_instrumentation__.Notify(32809)
		localityAdvertiseHosts[i].Address.AddressField = net.JoinHostPort(host, port)
	}
	__antithesis_instrumentation__.Notify(32771)
	serverCfg.LocalityAddresses = localityAdvertiseHosts

	return nil
}

func extraStoreFlagInit(cmd *cobra.Command) error {
	__antithesis_instrumentation__.Notify(32812)
	fs := flagSetForCmd(cmd)
	if fs.Changed(cliflags.Store.Name) {
		__antithesis_instrumentation__.Notify(32814)
		serverCfg.Stores = storeSpecs
	} else {
		__antithesis_instrumentation__.Notify(32815)
	}
	__antithesis_instrumentation__.Notify(32813)
	return nil
}

func extraClientFlagInit() error {
	__antithesis_instrumentation__.Notify(32816)

	principalMap := certCtx.certPrincipalMap
	if principalMap == nil {
		__antithesis_instrumentation__.Notify(32821)
		principalMap = cliCtx.certPrincipalMap
	} else {
		__antithesis_instrumentation__.Notify(32822)
	}
	__antithesis_instrumentation__.Notify(32817)
	if err := security.SetCertPrincipalMap(principalMap); err != nil {
		__antithesis_instrumentation__.Notify(32823)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32824)
	}
	__antithesis_instrumentation__.Notify(32818)
	serverCfg.Addr = net.JoinHostPort(cliCtx.clientConnHost, cliCtx.clientConnPort)
	serverCfg.AdvertiseAddr = serverCfg.Addr
	serverCfg.SQLAddr = net.JoinHostPort(cliCtx.clientConnHost, cliCtx.clientConnPort)
	serverCfg.SQLAdvertiseAddr = serverCfg.SQLAddr
	if serverHTTPAddr == "" {
		__antithesis_instrumentation__.Notify(32825)
		serverHTTPAddr = startCtx.serverListenAddr
	} else {
		__antithesis_instrumentation__.Notify(32826)
	}
	__antithesis_instrumentation__.Notify(32819)
	serverCfg.HTTPAddr = net.JoinHostPort(serverHTTPAddr, serverHTTPPort)

	if sqlConnCtx.DebugMode {
		__antithesis_instrumentation__.Notify(32827)
		sqlConnCtx.Echo = true
	} else {
		__antithesis_instrumentation__.Notify(32828)
	}
	__antithesis_instrumentation__.Notify(32820)
	return nil
}

func mtStartSQLFlagsInit(cmd *cobra.Command) error {
	__antithesis_instrumentation__.Notify(32829)

	fs := flagSetForCmd(cmd)
	if !fs.Changed(cliflags.Store.Name) {
		__antithesis_instrumentation__.Notify(32831)

		tenantID := fs.Lookup(cliflags.TenantID.Name).Value.String()
		serverCfg.Stores.Specs[0].Path = server.DefaultSQLNodeStorePathPrefix + tenantID
	} else {
		__antithesis_instrumentation__.Notify(32832)
	}
	__antithesis_instrumentation__.Notify(32830)
	return nil
}

var VarFlag = varFlag
