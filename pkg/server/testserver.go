package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	circuit "github.com/cockroachdb/circuitbreaker"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/ts"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	addrutil "github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
)

func makeTestConfig(st *cluster.Settings, tr *tracing.Tracer) Config {
	__antithesis_instrumentation__.Notify(239005)
	if tr == nil {
		__antithesis_instrumentation__.Notify(239007)
		panic("nil Tracer")
	} else {
		__antithesis_instrumentation__.Notify(239008)
	}
	__antithesis_instrumentation__.Notify(239006)
	return Config{
		BaseConfig: makeTestBaseConfig(st, tr),
		KVConfig:   makeTestKVConfig(),
		SQLConfig:  makeTestSQLConfig(st, roachpb.SystemTenantID),
	}
}

func makeTestBaseConfig(st *cluster.Settings, tr *tracing.Tracer) BaseConfig {
	__antithesis_instrumentation__.Notify(239009)
	if tr == nil {
		__antithesis_instrumentation__.Notify(239011)
		panic("nil Tracer")
	} else {
		__antithesis_instrumentation__.Notify(239012)
	}
	__antithesis_instrumentation__.Notify(239010)
	baseCfg := MakeBaseConfig(st, tr)

	baseCfg.Insecure = false

	baseCfg.StorageEngine = storage.DefaultStorageEngine

	baseCfg.SSLCertsDir = security.EmbeddedCertsDir

	baseCfg.Addr = util.TestAddr.String()
	baseCfg.AdvertiseAddr = util.TestAddr.String()
	baseCfg.SQLAddr = util.TestAddr.String()
	baseCfg.SQLAdvertiseAddr = util.TestAddr.String()
	baseCfg.SplitListenSQL = true
	baseCfg.HTTPAddr = util.TestAddr.String()

	baseCfg.User = security.NodeUserName()

	baseCfg.EnableWebSessionAuthentication = true
	return baseCfg
}

func makeTestKVConfig() KVConfig {
	__antithesis_instrumentation__.Notify(239013)
	kvCfg := MakeKVConfig(base.DefaultTestStoreSpec)
	return kvCfg
}

func makeTestSQLConfig(st *cluster.Settings, tenID roachpb.TenantID) SQLConfig {
	__antithesis_instrumentation__.Notify(239014)
	return MakeSQLConfig(tenID, base.DefaultTestTempStorageConfig(st))
}

func initTraceDir(dir string) error {
	__antithesis_instrumentation__.Notify(239015)
	if dir == "" {
		__antithesis_instrumentation__.Notify(239018)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(239019)
	}
	__antithesis_instrumentation__.Notify(239016)
	if err := os.MkdirAll(dir, 0755); err != nil {
		__antithesis_instrumentation__.Notify(239020)
		return errors.Wrap(err, "cannot create trace dir; traces will not be dumped")
	} else {
		__antithesis_instrumentation__.Notify(239021)
	}
	__antithesis_instrumentation__.Notify(239017)
	return nil
}

func makeTestConfigFromParams(params base.TestServerArgs) Config {
	__antithesis_instrumentation__.Notify(239022)
	st := params.Settings
	if params.Settings == nil {
		__antithesis_instrumentation__.Notify(239052)
		st = cluster.MakeClusterSettings()
	} else {
		__antithesis_instrumentation__.Notify(239053)
	}
	__antithesis_instrumentation__.Notify(239023)
	st.ExternalIODir = params.ExternalIODir
	tr := params.Tracer
	if params.Tracer == nil {
		__antithesis_instrumentation__.Notify(239054)
		tr = tracing.NewTracerWithOpt(context.TODO(), tracing.WithClusterSettings(&st.SV), tracing.WithTracingMode(params.TracingDefault))
	} else {
		__antithesis_instrumentation__.Notify(239055)
	}
	__antithesis_instrumentation__.Notify(239024)
	cfg := makeTestConfig(st, tr)
	cfg.TestingKnobs = params.Knobs
	cfg.RaftConfig = params.RaftConfig
	cfg.RaftConfig.SetDefaults()
	if params.JoinAddr != "" {
		__antithesis_instrumentation__.Notify(239056)
		cfg.JoinList = []string{params.JoinAddr}
	} else {
		__antithesis_instrumentation__.Notify(239057)
	}
	__antithesis_instrumentation__.Notify(239025)
	cfg.ClusterName = params.ClusterName
	cfg.ExternalIODirConfig = params.ExternalIODirConfig
	cfg.Insecure = params.Insecure
	cfg.AutoInitializeCluster = !params.NoAutoInitializeCluster
	cfg.SocketFile = params.SocketFile
	cfg.RetryOptions = params.RetryOptions
	cfg.Locality = params.Locality
	if params.TraceDir != "" {
		__antithesis_instrumentation__.Notify(239058)
		if err := initTraceDir(params.TraceDir); err == nil {
			__antithesis_instrumentation__.Notify(239059)
			cfg.InflightTraceDirName = params.TraceDir
		} else {
			__antithesis_instrumentation__.Notify(239060)
		}
	} else {
		__antithesis_instrumentation__.Notify(239061)
	}
	__antithesis_instrumentation__.Notify(239026)
	if knobs := params.Knobs.Store; knobs != nil {
		__antithesis_instrumentation__.Notify(239062)
		if mo := knobs.(*kvserver.StoreTestingKnobs).MaxOffset; mo != 0 {
			__antithesis_instrumentation__.Notify(239063)
			cfg.MaxOffset = MaxOffsetType(mo)
		} else {
			__antithesis_instrumentation__.Notify(239064)
		}
	} else {
		__antithesis_instrumentation__.Notify(239065)
	}
	__antithesis_instrumentation__.Notify(239027)
	if params.Knobs.Server != nil {
		__antithesis_instrumentation__.Notify(239066)
		if zoneConfig := params.Knobs.Server.(*TestingKnobs).DefaultZoneConfigOverride; zoneConfig != nil {
			__antithesis_instrumentation__.Notify(239068)
			cfg.DefaultZoneConfig = *zoneConfig
		} else {
			__antithesis_instrumentation__.Notify(239069)
		}
		__antithesis_instrumentation__.Notify(239067)
		if systemZoneConfig := params.Knobs.Server.(*TestingKnobs).DefaultSystemZoneConfigOverride; systemZoneConfig != nil {
			__antithesis_instrumentation__.Notify(239070)
			cfg.DefaultSystemZoneConfig = *systemZoneConfig
		} else {
			__antithesis_instrumentation__.Notify(239071)
		}
	} else {
		__antithesis_instrumentation__.Notify(239072)
	}
	__antithesis_instrumentation__.Notify(239028)
	if params.ScanInterval != 0 {
		__antithesis_instrumentation__.Notify(239073)
		cfg.ScanInterval = params.ScanInterval
	} else {
		__antithesis_instrumentation__.Notify(239074)
	}
	__antithesis_instrumentation__.Notify(239029)
	if params.ScanMinIdleTime != 0 {
		__antithesis_instrumentation__.Notify(239075)
		cfg.ScanMinIdleTime = params.ScanMinIdleTime
	} else {
		__antithesis_instrumentation__.Notify(239076)
	}
	__antithesis_instrumentation__.Notify(239030)
	if params.ScanMaxIdleTime != 0 {
		__antithesis_instrumentation__.Notify(239077)
		cfg.ScanMaxIdleTime = params.ScanMaxIdleTime
	} else {
		__antithesis_instrumentation__.Notify(239078)
	}
	__antithesis_instrumentation__.Notify(239031)
	if params.SSLCertsDir != "" {
		__antithesis_instrumentation__.Notify(239079)
		cfg.SSLCertsDir = params.SSLCertsDir
	} else {
		__antithesis_instrumentation__.Notify(239080)
	}
	__antithesis_instrumentation__.Notify(239032)
	if params.TimeSeriesQueryWorkerMax != 0 {
		__antithesis_instrumentation__.Notify(239081)
		cfg.TimeSeriesServerConfig.QueryWorkerMax = params.TimeSeriesQueryWorkerMax
	} else {
		__antithesis_instrumentation__.Notify(239082)
	}
	__antithesis_instrumentation__.Notify(239033)
	if params.TimeSeriesQueryMemoryBudget != 0 {
		__antithesis_instrumentation__.Notify(239083)
		cfg.TimeSeriesServerConfig.QueryMemoryMax = params.TimeSeriesQueryMemoryBudget
	} else {
		__antithesis_instrumentation__.Notify(239084)
	}
	__antithesis_instrumentation__.Notify(239034)
	if params.DisableEventLog {
		__antithesis_instrumentation__.Notify(239085)
		cfg.EventLogEnabled = false
	} else {
		__antithesis_instrumentation__.Notify(239086)
	}
	__antithesis_instrumentation__.Notify(239035)
	if params.SQLMemoryPoolSize != 0 {
		__antithesis_instrumentation__.Notify(239087)
		cfg.MemoryPoolSize = params.SQLMemoryPoolSize
	} else {
		__antithesis_instrumentation__.Notify(239088)
	}
	__antithesis_instrumentation__.Notify(239036)
	if params.CacheSize != 0 {
		__antithesis_instrumentation__.Notify(239089)
		cfg.CacheSize = params.CacheSize
	} else {
		__antithesis_instrumentation__.Notify(239090)
	}
	__antithesis_instrumentation__.Notify(239037)

	if params.JoinAddr != "" {
		__antithesis_instrumentation__.Notify(239091)
		cfg.JoinList = []string{params.JoinAddr}
	} else {
		__antithesis_instrumentation__.Notify(239092)
	}
	__antithesis_instrumentation__.Notify(239038)
	if cfg.Insecure {
		__antithesis_instrumentation__.Notify(239093)

		cfg.Addr = util.IsolatedTestAddr.String()
		cfg.AdvertiseAddr = util.IsolatedTestAddr.String()
		cfg.SQLAddr = util.IsolatedTestAddr.String()
		cfg.SQLAdvertiseAddr = util.IsolatedTestAddr.String()
		cfg.HTTPAddr = util.IsolatedTestAddr.String()
	} else {
		__antithesis_instrumentation__.Notify(239094)
	}
	__antithesis_instrumentation__.Notify(239039)
	if params.Addr != "" {
		__antithesis_instrumentation__.Notify(239095)
		cfg.Addr = params.Addr
		cfg.AdvertiseAddr = params.Addr
	} else {
		__antithesis_instrumentation__.Notify(239096)
	}
	__antithesis_instrumentation__.Notify(239040)
	if params.SQLAddr != "" {
		__antithesis_instrumentation__.Notify(239097)
		cfg.SQLAddr = params.SQLAddr
		cfg.SQLAdvertiseAddr = params.SQLAddr
		cfg.SplitListenSQL = true
	} else {
		__antithesis_instrumentation__.Notify(239098)
	}
	__antithesis_instrumentation__.Notify(239041)
	if params.HTTPAddr != "" {
		__antithesis_instrumentation__.Notify(239099)
		cfg.HTTPAddr = params.HTTPAddr
	} else {
		__antithesis_instrumentation__.Notify(239100)
	}
	__antithesis_instrumentation__.Notify(239042)
	cfg.DisableTLSForHTTP = params.DisableTLSForHTTP
	if params.DisableWebSessionAuthentication {
		__antithesis_instrumentation__.Notify(239101)
		cfg.EnableWebSessionAuthentication = false
	} else {
		__antithesis_instrumentation__.Notify(239102)
	}
	__antithesis_instrumentation__.Notify(239043)
	if params.EnableDemoLoginEndpoint {
		__antithesis_instrumentation__.Notify(239103)
		cfg.EnableDemoLoginEndpoint = true
	} else {
		__antithesis_instrumentation__.Notify(239104)
	}
	__antithesis_instrumentation__.Notify(239044)
	if params.DisableSpanConfigs {
		__antithesis_instrumentation__.Notify(239105)
		cfg.SpanConfigsDisabled = true
	} else {
		__antithesis_instrumentation__.Notify(239106)
	}
	__antithesis_instrumentation__.Notify(239045)

	if len(params.StoreSpecs) == 0 {
		__antithesis_instrumentation__.Notify(239107)
		params.StoreSpecs = []base.StoreSpec{base.DefaultTestStoreSpec}
	} else {
		__antithesis_instrumentation__.Notify(239108)
	}
	__antithesis_instrumentation__.Notify(239046)

	for _, storeSpec := range params.StoreSpecs {
		__antithesis_instrumentation__.Notify(239109)
		if storeSpec.InMemory {
			__antithesis_instrumentation__.Notify(239110)
			if storeSpec.Size.Percent > 0 {
				__antithesis_instrumentation__.Notify(239111)
				panic(fmt.Sprintf("test server does not yet support in memory stores based on percentage of total memory: %s", storeSpec))
			} else {
				__antithesis_instrumentation__.Notify(239112)
			}
		} else {
			__antithesis_instrumentation__.Notify(239113)

			if cfg.HeapProfileDirName == "" {
				__antithesis_instrumentation__.Notify(239116)
				cfg.HeapProfileDirName = filepath.Join(storeSpec.Path, "logs", base.HeapProfileDir)
			} else {
				__antithesis_instrumentation__.Notify(239117)
			}
			__antithesis_instrumentation__.Notify(239114)
			if cfg.GoroutineDumpDirName == "" {
				__antithesis_instrumentation__.Notify(239118)
				cfg.GoroutineDumpDirName = filepath.Join(storeSpec.Path, "logs", base.GoroutineDumpDir)
			} else {
				__antithesis_instrumentation__.Notify(239119)
			}
			__antithesis_instrumentation__.Notify(239115)
			if cfg.InflightTraceDirName == "" {
				__antithesis_instrumentation__.Notify(239120)
				cfg.InflightTraceDirName = filepath.Join(storeSpec.Path, "logs", base.InflightTraceDir)
			} else {
				__antithesis_instrumentation__.Notify(239121)
			}
		}
	}
	__antithesis_instrumentation__.Notify(239047)
	cfg.Stores = base.StoreSpecList{Specs: params.StoreSpecs}
	if params.TempStorageConfig.InMemory || func() bool {
		__antithesis_instrumentation__.Notify(239122)
		return params.TempStorageConfig.Path != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(239123)
		cfg.TempStorageConfig = params.TempStorageConfig
	} else {
		__antithesis_instrumentation__.Notify(239124)
	}
	__antithesis_instrumentation__.Notify(239048)

	if cfg.TestingKnobs.Store == nil {
		__antithesis_instrumentation__.Notify(239125)
		cfg.TestingKnobs.Store = &kvserver.StoreTestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(239126)
	}
	__antithesis_instrumentation__.Notify(239049)
	cfg.TestingKnobs.Store.(*kvserver.StoreTestingKnobs).SkipMinSizeCheck = true

	if params.Knobs.SQLExecutor == nil {
		__antithesis_instrumentation__.Notify(239127)
		cfg.TestingKnobs.SQLExecutor = &sql.ExecutorTestingKnobs{}
	} else {
		__antithesis_instrumentation__.Notify(239128)
	}
	__antithesis_instrumentation__.Notify(239050)

	if params.Knobs.AdmissionControl == nil {
		__antithesis_instrumentation__.Notify(239129)
		cfg.TestingKnobs.AdmissionControl = &admission.Options{}
	} else {
		__antithesis_instrumentation__.Notify(239130)
	}
	__antithesis_instrumentation__.Notify(239051)

	return cfg
}

type TestServer struct {
	Cfg    *Config
	params base.TestServerArgs

	*Server

	*httpTestServer
}

var _ serverutils.TestServerInterface = &TestServer{}

func (ts *TestServer) Node() interface{} {
	__antithesis_instrumentation__.Notify(239131)
	return ts.node
}

func (ts *TestServer) NodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(239132)
	return ts.rpcContext.NodeID.Get()
}

func (ts *TestServer) Stopper() *stop.Stopper {
	__antithesis_instrumentation__.Notify(239133)
	return ts.stopper
}

func (ts *TestServer) GossipI() interface{} {
	__antithesis_instrumentation__.Notify(239134)
	return ts.Gossip()
}

func (ts *TestServer) Gossip() *gossip.Gossip {
	__antithesis_instrumentation__.Notify(239135)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239137)
		return ts.gossip
	} else {
		__antithesis_instrumentation__.Notify(239138)
	}
	__antithesis_instrumentation__.Notify(239136)
	return nil
}

func (ts *TestServer) RangeFeedFactory() interface{} {
	__antithesis_instrumentation__.Notify(239139)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239141)
		return ts.sqlServer.execCfg.RangeFeedFactory
	} else {
		__antithesis_instrumentation__.Notify(239142)
	}
	__antithesis_instrumentation__.Notify(239140)
	return (*rangefeed.Factory)(nil)
}

func (ts *TestServer) Clock() *hlc.Clock {
	__antithesis_instrumentation__.Notify(239143)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239145)
		return ts.clock
	} else {
		__antithesis_instrumentation__.Notify(239146)
	}
	__antithesis_instrumentation__.Notify(239144)
	return nil
}

func (ts *TestServer) SQLLivenessProvider() interface{} {
	__antithesis_instrumentation__.Notify(239147)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239149)
		return ts.sqlServer.execCfg.SQLLiveness
	} else {
		__antithesis_instrumentation__.Notify(239150)
	}
	__antithesis_instrumentation__.Notify(239148)
	return nil
}

func (ts *TestServer) JobRegistry() interface{} {
	__antithesis_instrumentation__.Notify(239151)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239153)
		return ts.sqlServer.jobRegistry
	} else {
		__antithesis_instrumentation__.Notify(239154)
	}
	__antithesis_instrumentation__.Notify(239152)
	return nil
}

func (ts *TestServer) StartupMigrationsManager() interface{} {
	__antithesis_instrumentation__.Notify(239155)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239157)
		return ts.sqlServer.startupMigrationsMgr
	} else {
		__antithesis_instrumentation__.Notify(239158)
	}
	__antithesis_instrumentation__.Notify(239156)
	return nil
}

func (ts *TestServer) NodeLiveness() interface{} {
	__antithesis_instrumentation__.Notify(239159)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239161)
		return ts.nodeLiveness
	} else {
		__antithesis_instrumentation__.Notify(239162)
	}
	__antithesis_instrumentation__.Notify(239160)
	return nil
}

func (ts *TestServer) NodeDialer() interface{} {
	__antithesis_instrumentation__.Notify(239163)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239165)
		return ts.nodeDialer
	} else {
		__antithesis_instrumentation__.Notify(239166)
	}
	__antithesis_instrumentation__.Notify(239164)
	return nil
}

func (ts *TestServer) HeartbeatNodeLiveness() error {
	__antithesis_instrumentation__.Notify(239167)
	if ts == nil {
		__antithesis_instrumentation__.Notify(239171)
		return errors.New("no node liveness instance")
	} else {
		__antithesis_instrumentation__.Notify(239172)
	}
	__antithesis_instrumentation__.Notify(239168)
	nl := ts.nodeLiveness
	l, ok := nl.Self()
	if !ok {
		__antithesis_instrumentation__.Notify(239173)
		return errors.New("liveness not found")
	} else {
		__antithesis_instrumentation__.Notify(239174)
	}
	__antithesis_instrumentation__.Notify(239169)

	var err error
	ctx := context.Background()
	for r := retry.StartWithCtx(ctx, retry.Options{MaxRetries: 5}); r.Next(); {
		__antithesis_instrumentation__.Notify(239175)
		if err = nl.Heartbeat(ctx, l); !errors.Is(err, liveness.ErrEpochIncremented) {
			__antithesis_instrumentation__.Notify(239176)
			break
		} else {
			__antithesis_instrumentation__.Notify(239177)
		}
	}
	__antithesis_instrumentation__.Notify(239170)
	return err
}

func (ts *TestServer) SQLInstanceID() base.SQLInstanceID {
	__antithesis_instrumentation__.Notify(239178)
	return ts.sqlServer.sqlIDContainer.SQLInstanceID()
}

func (ts *TestServer) StatusServer() interface{} {
	__antithesis_instrumentation__.Notify(239179)
	return ts.status
}

func (ts *TestServer) RPCContext() *rpc.Context {
	__antithesis_instrumentation__.Notify(239180)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239182)
		return ts.rpcContext
	} else {
		__antithesis_instrumentation__.Notify(239183)
	}
	__antithesis_instrumentation__.Notify(239181)
	return nil
}

func (ts *TestServer) TsDB() *ts.DB {
	__antithesis_instrumentation__.Notify(239184)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239186)
		return ts.tsDB
	} else {
		__antithesis_instrumentation__.Notify(239187)
	}
	__antithesis_instrumentation__.Notify(239185)
	return nil
}

func (ts *TestServer) DB() *kv.DB {
	__antithesis_instrumentation__.Notify(239188)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239190)
		return ts.db
	} else {
		__antithesis_instrumentation__.Notify(239191)
	}
	__antithesis_instrumentation__.Notify(239189)
	return nil
}

func (ts *TestServer) PGServer() interface{} {
	__antithesis_instrumentation__.Notify(239192)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239194)
		return ts.sqlServer.pgServer
	} else {
		__antithesis_instrumentation__.Notify(239195)
	}
	__antithesis_instrumentation__.Notify(239193)
	return nil
}

func (ts *TestServer) RaftTransport() *kvserver.RaftTransport {
	__antithesis_instrumentation__.Notify(239196)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239198)
		return ts.raftTransport
	} else {
		__antithesis_instrumentation__.Notify(239199)
	}
	__antithesis_instrumentation__.Notify(239197)
	return nil
}

func (ts *TestServer) AmbientCtx() log.AmbientContext {
	__antithesis_instrumentation__.Notify(239200)
	return ts.Cfg.AmbientCtx
}

func (ts *TestServer) TestingKnobs() *base.TestingKnobs {
	__antithesis_instrumentation__.Notify(239201)
	if ts != nil {
		__antithesis_instrumentation__.Notify(239203)
		return &ts.Cfg.TestingKnobs
	} else {
		__antithesis_instrumentation__.Notify(239204)
	}
	__antithesis_instrumentation__.Notify(239202)
	return nil
}

func (ts *TestServer) TenantStatusServer() interface{} {
	__antithesis_instrumentation__.Notify(239205)
	return ts.status
}

func (ts *TestServer) Start(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(239206)
	return ts.Server.Start(ctx)
}

type tenantProtectedTSProvider struct {
	protectedts.Provider
	st *cluster.Settings
}

func (d tenantProtectedTSProvider) Protect(
	ctx context.Context, txn *kv.Txn, rec *ptpb.Record,
) error {
	__antithesis_instrumentation__.Notify(239207)
	if !d.st.Version.IsActive(ctx, clusterversion.EnableProtectedTimestampsForTenant) {
		__antithesis_instrumentation__.Notify(239209)
		return errors.Newf("%s is inactive, tenant cannot write protected timestamp records",
			clusterversion.EnableProtectedTimestampsForTenant.String())
	} else {
		__antithesis_instrumentation__.Notify(239210)
	}
	__antithesis_instrumentation__.Notify(239208)
	return d.Provider.Protect(ctx, txn, rec)
}

type TestTenant struct {
	*SQLServer
	Cfg      *BaseConfig
	sqlAddr  string
	httpAddr string
	*httpTestServer
	drain *drainServer
}

var _ serverutils.TestTenantInterface = &TestTenant{}

func (t *TestTenant) SQLAddr() string {
	__antithesis_instrumentation__.Notify(239211)
	return t.sqlAddr
}

func (t *TestTenant) HTTPAddr() string {
	__antithesis_instrumentation__.Notify(239212)
	return t.httpAddr
}

func (t *TestTenant) RPCAddr() string {
	__antithesis_instrumentation__.Notify(239213)

	return t.sqlAddr
}

func (t *TestTenant) PGServer() interface{} {
	__antithesis_instrumentation__.Notify(239214)
	return t.pgServer
}

func (t *TestTenant) DiagnosticsReporter() interface{} {
	__antithesis_instrumentation__.Notify(239215)
	return t.diagnosticsReporter
}

func (t *TestTenant) StatusServer() interface{} {
	__antithesis_instrumentation__.Notify(239216)
	return t.execCfg.SQLStatusServer
}

func (t *TestTenant) TenantStatusServer() interface{} {
	__antithesis_instrumentation__.Notify(239217)
	return t.execCfg.TenantStatusServer
}

func (t *TestTenant) DistSQLServer() interface{} {
	__antithesis_instrumentation__.Notify(239218)
	return t.SQLServer.distSQLServer
}

func (t *TestTenant) RPCContext() *rpc.Context {
	__antithesis_instrumentation__.Notify(239219)
	return t.execCfg.RPCContext
}

func (t *TestTenant) JobRegistry() interface{} {
	__antithesis_instrumentation__.Notify(239220)
	return t.SQLServer.jobRegistry
}

func (t *TestTenant) ExecutorConfig() interface{} {
	__antithesis_instrumentation__.Notify(239221)
	return *t.SQLServer.execCfg
}

func (t *TestTenant) RangeFeedFactory() interface{} {
	__antithesis_instrumentation__.Notify(239222)
	return t.SQLServer.execCfg.RangeFeedFactory
}

func (t *TestTenant) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(239223)
	return t.Cfg.Settings
}

func (t *TestTenant) Stopper() *stop.Stopper {
	__antithesis_instrumentation__.Notify(239224)
	return t.stopper
}

func (t *TestTenant) Clock() *hlc.Clock {
	__antithesis_instrumentation__.Notify(239225)
	return t.SQLServer.execCfg.Clock
}

func (t *TestTenant) AmbientCtx() log.AmbientContext {
	__antithesis_instrumentation__.Notify(239226)
	return t.Cfg.AmbientCtx
}

func (t *TestTenant) TestingKnobs() *base.TestingKnobs {
	__antithesis_instrumentation__.Notify(239227)
	return &t.Cfg.TestingKnobs
}

func (t *TestTenant) SpanConfigKVAccessor() interface{} {
	__antithesis_instrumentation__.Notify(239228)
	return t.SQLServer.tenantConnect
}

func (t *TestTenant) SpanConfigReconciler() interface{} {
	__antithesis_instrumentation__.Notify(239229)
	return t.SQLServer.spanconfigMgr.Reconciler
}

func (t *TestTenant) SpanConfigSQLTranslatorFactory() interface{} {
	__antithesis_instrumentation__.Notify(239230)
	return t.SQLServer.spanconfigSQLTranslatorFactory
}

func (t *TestTenant) SpanConfigSQLWatcher() interface{} {
	__antithesis_instrumentation__.Notify(239231)
	return t.SQLServer.spanconfigSQLWatcher
}

func (t *TestTenant) SystemConfigProvider() config.SystemConfigProvider {
	__antithesis_instrumentation__.Notify(239232)
	return t.SQLServer.systemConfigWatcher
}

func (t *TestTenant) DrainClients(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(239233)
	return t.drain.drainClients(ctx, nil)
}

func (ts *TestServer) StartTenant(
	ctx context.Context, params base.TestTenantArgs,
) (serverutils.TestTenantInterface, error) {
	__antithesis_instrumentation__.Notify(239234)
	if !params.Existing {
		__antithesis_instrumentation__.Notify(239247)
		if _, err := ts.InternalExecutor().(*sql.InternalExecutor).Exec(
			ctx, "testserver-create-tenant", nil, "SELECT crdb_internal.create_tenant($1)", params.TenantID.ToUint64(),
		); err != nil {
			__antithesis_instrumentation__.Notify(239248)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(239249)
		}
	} else {
		__antithesis_instrumentation__.Notify(239250)
	}
	__antithesis_instrumentation__.Notify(239235)

	if !params.SkipTenantCheck {
		__antithesis_instrumentation__.Notify(239251)
		rowCount, err := ts.InternalExecutor().(*sql.InternalExecutor).Exec(
			ctx, "testserver-check-tenant-active", nil,
			"SELECT 1 FROM system.tenants WHERE id=$1 AND active=true",
			params.TenantID.ToUint64(),
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(239253)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(239254)
		}
		__antithesis_instrumentation__.Notify(239252)
		if rowCount == 0 {
			__antithesis_instrumentation__.Notify(239255)
			return nil, errors.New("not found")
		} else {
			__antithesis_instrumentation__.Notify(239256)
		}
	} else {
		__antithesis_instrumentation__.Notify(239257)
	}
	__antithesis_instrumentation__.Notify(239236)
	st := params.Settings
	if st == nil {
		__antithesis_instrumentation__.Notify(239258)
		st = cluster.MakeTestingClusterSettings()
	} else {
		__antithesis_instrumentation__.Notify(239259)
	}
	__antithesis_instrumentation__.Notify(239237)

	st.ExternalIODir = params.ExternalIODir
	sqlCfg := makeTestSQLConfig(st, params.TenantID)
	sqlCfg.TenantKVAddrs = []string{ts.ServingRPCAddr()}
	sqlCfg.ExternalIODirConfig = params.ExternalIODirConfig
	if params.MemoryPoolSize != 0 {
		__antithesis_instrumentation__.Notify(239260)
		sqlCfg.MemoryPoolSize = params.MemoryPoolSize
	} else {
		__antithesis_instrumentation__.Notify(239261)
	}
	__antithesis_instrumentation__.Notify(239238)
	if params.TempStorageConfig != nil {
		__antithesis_instrumentation__.Notify(239262)
		sqlCfg.TempStorageConfig = *params.TempStorageConfig
	} else {
		__antithesis_instrumentation__.Notify(239263)
	}
	__antithesis_instrumentation__.Notify(239239)

	stopper := params.Stopper
	if stopper == nil {
		__antithesis_instrumentation__.Notify(239264)

		tr := tracing.NewTracerWithOpt(ctx, tracing.WithClusterSettings(&st.SV), tracing.WithTracingMode(params.TracingDefault))
		stopper = stop.NewStopper(stop.WithTracer(tr))

		ts.Stopper().AddCloser(stop.CloserFn(func() { __antithesis_instrumentation__.Notify(239265); stopper.Stop(context.Background()) }))
	} else {
		__antithesis_instrumentation__.Notify(239266)
		if stopper.Tracer() == nil {
			__antithesis_instrumentation__.Notify(239267)
			tr := tracing.NewTracerWithOpt(ctx, tracing.WithClusterSettings(&st.SV), tracing.WithTracingMode(params.TracingDefault))
			stopper.SetTracer(tr)
		} else {
			__antithesis_instrumentation__.Notify(239268)
		}
	}
	__antithesis_instrumentation__.Notify(239240)

	baseCfg := makeTestBaseConfig(st, stopper.Tracer())
	baseCfg.TestingKnobs = params.TestingKnobs
	baseCfg.Insecure = params.ForceInsecure
	baseCfg.Locality = params.Locality
	baseCfg.HeapProfileDirName = params.HeapProfileDirName
	baseCfg.GoroutineDumpDirName = params.GoroutineDumpDirName
	if params.SSLCertsDir != "" {
		__antithesis_instrumentation__.Notify(239269)
		baseCfg.SSLCertsDir = params.SSLCertsDir
	} else {
		__antithesis_instrumentation__.Notify(239270)
	}
	__antithesis_instrumentation__.Notify(239241)
	if params.StartingSQLPort > 0 {
		__antithesis_instrumentation__.Notify(239271)
		addr, _, err := addrutil.SplitHostPort(baseCfg.SQLAddr, strconv.Itoa(params.StartingSQLPort))
		if err != nil {
			__antithesis_instrumentation__.Notify(239273)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(239274)
		}
		__antithesis_instrumentation__.Notify(239272)
		newAddr := net.JoinHostPort(addr, strconv.Itoa(params.StartingSQLPort+int(params.TenantID.ToUint64())))
		baseCfg.SQLAddr = newAddr
		baseCfg.SQLAdvertiseAddr = newAddr
	} else {
		__antithesis_instrumentation__.Notify(239275)
	}
	__antithesis_instrumentation__.Notify(239242)
	if params.StartingHTTPPort > 0 {
		__antithesis_instrumentation__.Notify(239276)
		addr, _, err := addrutil.SplitHostPort(baseCfg.SQLAddr, strconv.Itoa(params.StartingHTTPPort))
		if err != nil {
			__antithesis_instrumentation__.Notify(239278)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(239279)
		}
		__antithesis_instrumentation__.Notify(239277)
		newAddr := net.JoinHostPort(addr, strconv.Itoa(params.StartingHTTPPort+int(params.TenantID.ToUint64())))
		baseCfg.HTTPAddr = newAddr
		baseCfg.HTTPAdvertiseAddr = newAddr
	} else {
		__antithesis_instrumentation__.Notify(239280)
	}
	__antithesis_instrumentation__.Notify(239243)
	if params.AllowSettingClusterSettings {
		__antithesis_instrumentation__.Notify(239281)
		tenantKnobs, ok := baseCfg.TestingKnobs.TenantTestingKnobs.(*sql.TenantTestingKnobs)
		if !ok {
			__antithesis_instrumentation__.Notify(239283)
			tenantKnobs = &sql.TenantTestingKnobs{}
			baseCfg.TestingKnobs.TenantTestingKnobs = tenantKnobs
		} else {
			__antithesis_instrumentation__.Notify(239284)
		}
		__antithesis_instrumentation__.Notify(239282)
		if tenantKnobs.ClusterSettingsUpdater == nil {
			__antithesis_instrumentation__.Notify(239285)
			tenantKnobs.ClusterSettingsUpdater = st.MakeUpdater()
		} else {
			__antithesis_instrumentation__.Notify(239286)
		}
	} else {
		__antithesis_instrumentation__.Notify(239287)
	}
	__antithesis_instrumentation__.Notify(239244)
	if params.RPCHeartbeatInterval != 0 {
		__antithesis_instrumentation__.Notify(239288)
		baseCfg.RPCHeartbeatInterval = params.RPCHeartbeatInterval
	} else {
		__antithesis_instrumentation__.Notify(239289)
	}
	__antithesis_instrumentation__.Notify(239245)
	sqlServer, authServer, drainServer, addr, httpAddr, err := startTenantInternal(
		ctx,
		stopper,
		ts.Cfg.ClusterName,
		baseCfg,
		sqlCfg,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(239290)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(239291)
	}
	__antithesis_instrumentation__.Notify(239246)

	hts := &httpTestServer{}
	hts.t.authentication = authServer
	hts.t.sqlServer = sqlServer

	return &TestTenant{
		SQLServer:      sqlServer,
		Cfg:            &baseCfg,
		sqlAddr:        addr,
		httpAddr:       httpAddr,
		httpTestServer: hts,
		drain:          drainServer,
	}, err
}

func (ts *TestServer) ExpectedInitialRangeCount() (int, error) {
	__antithesis_instrumentation__.Notify(239292)
	return ExpectedInitialRangeCount(
		ts.sqlServer.execCfg.Codec,
		&ts.cfg.DefaultZoneConfig,
		&ts.cfg.DefaultSystemZoneConfig,
	)
}

func ExpectedInitialRangeCount(
	codec keys.SQLCodec,
	defaultZoneConfig *zonepb.ZoneConfig,
	defaultSystemZoneConfig *zonepb.ZoneConfig,
) (int, error) {
	__antithesis_instrumentation__.Notify(239293)
	_, splits := bootstrap.MakeMetadataSchema(codec, defaultZoneConfig, defaultSystemZoneConfig).GetInitialValues()

	return len(config.StaticSplits()) + len(splits) + 1, nil
}

func (ts *TestServer) Stores() *kvserver.Stores {
	__antithesis_instrumentation__.Notify(239294)
	return ts.node.stores
}

func (ts *TestServer) GetStores() interface{} {
	__antithesis_instrumentation__.Notify(239295)
	return ts.node.stores
}

func (ts *TestServer) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(239296)
	return ts.Cfg.Settings
}

func (ts *TestServer) Engines() []storage.Engine {
	__antithesis_instrumentation__.Notify(239297)
	return ts.engines
}

func (ts *TestServer) ServingRPCAddr() string {
	__antithesis_instrumentation__.Notify(239298)
	return ts.cfg.AdvertiseAddr
}

func (ts *TestServer) ServingSQLAddr() string {
	__antithesis_instrumentation__.Notify(239299)
	return ts.cfg.SQLAdvertiseAddr
}

func (ts *TestServer) HTTPAddr() string {
	__antithesis_instrumentation__.Notify(239300)
	return ts.cfg.HTTPAddr
}

func (ts *TestServer) RPCAddr() string {
	__antithesis_instrumentation__.Notify(239301)
	return ts.cfg.Addr
}

func (ts *TestServer) SQLAddr() string {
	__antithesis_instrumentation__.Notify(239302)
	return ts.cfg.SQLAddr
}

func (ts *TestServer) DrainClients(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(239303)
	return ts.drain.drainClients(ctx, nil)
}

func (ts *TestServer) Readiness(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(239304)
	return ts.admin.checkReadinessForHealthCheck(ctx)
}

func (ts *TestServer) WriteSummaries() error {
	__antithesis_instrumentation__.Notify(239305)
	return ts.node.writeNodeStatus(context.TODO(), time.Hour, false)
}

func (ts *TestServer) UpdateChecker() interface{} {
	__antithesis_instrumentation__.Notify(239306)
	return ts.Server.updates
}

func (ts *TestServer) DiagnosticsReporter() interface{} {
	__antithesis_instrumentation__.Notify(239307)
	return ts.Server.sqlServer.diagnosticsReporter
}

const authenticatedUser = "authentic_user"

func authenticatedUserName() security.SQLUsername {
	__antithesis_instrumentation__.Notify(239308)
	return security.MakeSQLUsernameFromPreNormalizedString(authenticatedUser)
}

const authenticatedUserNoAdmin = "authentic_user_noadmin"

func authenticatedUserNameNoAdmin() security.SQLUsername {
	__antithesis_instrumentation__.Notify(239309)
	return security.MakeSQLUsernameFromPreNormalizedString(authenticatedUserNoAdmin)
}

type v2AuthDecorator struct {
	http.RoundTripper

	session string
}

func (v *v2AuthDecorator) RoundTrip(r *http.Request) (*http.Response, error) {
	__antithesis_instrumentation__.Notify(239310)
	r.Header.Add(apiV2AuthHeader, v.session)
	return v.RoundTripper.RoundTrip(r)
}

func (ts *TestServer) MustGetSQLCounter(name string) int64 {
	__antithesis_instrumentation__.Notify(239311)
	var c int64
	var found bool

	type (
		int64Valuer  interface{ Value() int64 }
		int64Counter interface{ Count() int64 }
	)

	ts.registry.Each(func(n string, v interface{}) {
		__antithesis_instrumentation__.Notify(239314)
		if name == n {
			__antithesis_instrumentation__.Notify(239315)
			switch t := v.(type) {
			case *metric.Counter:
				__antithesis_instrumentation__.Notify(239316)
				c = t.Count()
				found = true
			case *metric.Gauge:
				__antithesis_instrumentation__.Notify(239317)
				c = t.Value()
				found = true
			case int64Valuer:
				__antithesis_instrumentation__.Notify(239318)
				c = t.Value()
				found = true
			case int64Counter:
				__antithesis_instrumentation__.Notify(239319)
				c = t.Count()
				found = true
			}
		} else {
			__antithesis_instrumentation__.Notify(239320)
		}
	})
	__antithesis_instrumentation__.Notify(239312)
	if !found {
		__antithesis_instrumentation__.Notify(239321)
		panic(fmt.Sprintf("couldn't find metric %s", name))
	} else {
		__antithesis_instrumentation__.Notify(239322)
	}
	__antithesis_instrumentation__.Notify(239313)
	return c
}

func (ts *TestServer) MustGetSQLNetworkCounter(name string) int64 {
	__antithesis_instrumentation__.Notify(239323)
	var c int64
	var found bool

	reg := metric.NewRegistry()
	for _, m := range ts.sqlServer.pgServer.Metrics() {
		__antithesis_instrumentation__.Notify(239327)
		reg.AddMetricStruct(m)
	}
	__antithesis_instrumentation__.Notify(239324)
	reg.Each(func(n string, v interface{}) {
		__antithesis_instrumentation__.Notify(239328)
		if name == n {
			__antithesis_instrumentation__.Notify(239329)
			switch t := v.(type) {
			case *metric.Counter:
				__antithesis_instrumentation__.Notify(239330)
				c = t.Count()
				found = true
			case *metric.Gauge:
				__antithesis_instrumentation__.Notify(239331)
				c = t.Value()
				found = true
			}
		} else {
			__antithesis_instrumentation__.Notify(239332)
		}
	})
	__antithesis_instrumentation__.Notify(239325)
	if !found {
		__antithesis_instrumentation__.Notify(239333)
		panic(fmt.Sprintf("couldn't find metric %s", name))
	} else {
		__antithesis_instrumentation__.Notify(239334)
	}
	__antithesis_instrumentation__.Notify(239326)
	return c
}

func (ts *TestServer) Locality() *roachpb.Locality {
	__antithesis_instrumentation__.Notify(239335)
	return &ts.cfg.Locality
}

func (ts *TestServer) LeaseManager() interface{} {
	__antithesis_instrumentation__.Notify(239336)
	return ts.sqlServer.leaseMgr
}

func (ts *TestServer) InternalExecutor() interface{} {
	__antithesis_instrumentation__.Notify(239337)
	return ts.sqlServer.internalExecutor
}

func (ts *TestServer) GetNode() *Node {
	__antithesis_instrumentation__.Notify(239338)
	return ts.node
}

func (ts *TestServer) DistSenderI() interface{} {
	__antithesis_instrumentation__.Notify(239339)
	return ts.distSender
}

func (ts *TestServer) DistSender() *kvcoord.DistSender {
	__antithesis_instrumentation__.Notify(239340)
	return ts.DistSenderI().(*kvcoord.DistSender)
}

func (ts *TestServer) MigrationServer() interface{} {
	__antithesis_instrumentation__.Notify(239341)
	return ts.migrationServer
}

func (ts *TestServer) SpanConfigKVAccessor() interface{} {
	__antithesis_instrumentation__.Notify(239342)
	return ts.Server.node.spanConfigAccessor
}

func (ts *TestServer) SpanConfigReconciler() interface{} {
	__antithesis_instrumentation__.Notify(239343)
	if ts.sqlServer.spanconfigMgr == nil {
		__antithesis_instrumentation__.Notify(239345)
		panic("uninitialized; see EnableSpanConfigs testing knob to use span configs")
	} else {
		__antithesis_instrumentation__.Notify(239346)
	}
	__antithesis_instrumentation__.Notify(239344)
	return ts.sqlServer.spanconfigMgr.Reconciler
}

func (ts *TestServer) SpanConfigSQLTranslatorFactory() interface{} {
	__antithesis_instrumentation__.Notify(239347)
	if ts.sqlServer.spanconfigSQLTranslatorFactory == nil {
		__antithesis_instrumentation__.Notify(239349)
		panic("uninitialized; see EnableSpanConfigs testing knob to use span configs")
	} else {
		__antithesis_instrumentation__.Notify(239350)
	}
	__antithesis_instrumentation__.Notify(239348)
	return ts.sqlServer.spanconfigSQLTranslatorFactory
}

func (ts *TestServer) SpanConfigSQLWatcher() interface{} {
	__antithesis_instrumentation__.Notify(239351)
	if ts.sqlServer.spanconfigSQLWatcher == nil {
		__antithesis_instrumentation__.Notify(239353)
		panic("uninitialized; see EnableSpanConfigs testing knob to use span configs")
	} else {
		__antithesis_instrumentation__.Notify(239354)
	}
	__antithesis_instrumentation__.Notify(239352)
	return ts.sqlServer.spanconfigSQLWatcher
}

func (ts *TestServer) SQLServer() interface{} {
	__antithesis_instrumentation__.Notify(239355)
	return ts.sqlServer.pgServer.SQLServer
}

func (ts *TestServer) DistSQLServer() interface{} {
	__antithesis_instrumentation__.Notify(239356)
	return ts.sqlServer.distSQLServer
}

func (s *Server) SetDistSQLSpanResolver(spanResolver interface{}) {
	__antithesis_instrumentation__.Notify(239357)
	s.sqlServer.execCfg.DistSQLPlanner.SetSpanResolver(spanResolver.(physicalplan.SpanResolver))
}

func (ts *TestServer) GetFirstStoreID() roachpb.StoreID {
	__antithesis_instrumentation__.Notify(239358)
	firstStoreID := roachpb.StoreID(-1)
	err := ts.Stores().VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(239361)
		if firstStoreID == -1 {
			__antithesis_instrumentation__.Notify(239363)
			firstStoreID = s.Ident.StoreID
		} else {
			__antithesis_instrumentation__.Notify(239364)
		}
		__antithesis_instrumentation__.Notify(239362)
		return nil
	})
	__antithesis_instrumentation__.Notify(239359)
	if err != nil {
		__antithesis_instrumentation__.Notify(239365)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(239366)
	}
	__antithesis_instrumentation__.Notify(239360)
	return firstStoreID
}

func (ts *TestServer) LookupRange(key roachpb.Key) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(239367)
	rs, _, err := kv.RangeLookup(context.Background(), ts.DB().NonTransactionalSender(),
		key, roachpb.CONSISTENT, 0, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(239369)
		return roachpb.RangeDescriptor{}, errors.Wrapf(
			err, "%q: lookup range unexpected error", key)
	} else {
		__antithesis_instrumentation__.Notify(239370)
	}
	__antithesis_instrumentation__.Notify(239368)
	return rs[0], nil
}

func (ts *TestServer) MergeRanges(leftKey roachpb.Key) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(239371)

	ctx := context.Background()
	mergeReq := roachpb.AdminMergeRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: leftKey,
		},
	}
	_, pErr := kv.SendWrapped(ctx, ts.DB().NonTransactionalSender(), &mergeReq)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(239373)
		return roachpb.RangeDescriptor{},
			errors.Errorf(
				"%q: merge unexpected error: %s", leftKey, pErr)
	} else {
		__antithesis_instrumentation__.Notify(239374)
	}
	__antithesis_instrumentation__.Notify(239372)
	return ts.LookupRange(leftKey)
}

func (ts *TestServer) SplitRangeWithExpiration(
	splitKey roachpb.Key, expirationTime hlc.Timestamp,
) (roachpb.RangeDescriptor, roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(239375)
	ctx := context.Background()
	splitReq := roachpb.AdminSplitRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: splitKey,
		},
		SplitKey:       splitKey,
		ExpirationTime: expirationTime,
	}
	_, pErr := kv.SendWrapped(ctx, ts.DB().NonTransactionalSender(), &splitReq)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(239378)
		return roachpb.RangeDescriptor{}, roachpb.RangeDescriptor{},
			errors.Errorf(
				"%q: split unexpected error: %s", splitReq.SplitKey, pErr)
	} else {
		__antithesis_instrumentation__.Notify(239379)
	}
	__antithesis_instrumentation__.Notify(239376)

	var leftRangeDesc, rightRangeDesc roachpb.RangeDescriptor

	var wrappedMsg string
	if err := ts.DB().Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(239380)
		leftRangeDesc, rightRangeDesc = roachpb.RangeDescriptor{}, roachpb.RangeDescriptor{}

		rs, more, err := kv.RangeLookup(ctx, txn, splitKey.Next(), roachpb.CONSISTENT, 1, true)
		if err != nil {
			__antithesis_instrumentation__.Notify(239385)
			return err
		} else {
			__antithesis_instrumentation__.Notify(239386)
		}
		__antithesis_instrumentation__.Notify(239381)
		if len(rs) == 0 {
			__antithesis_instrumentation__.Notify(239387)

			return errors.AssertionFailedf("no descriptor found for key %s", splitKey)
		} else {
			__antithesis_instrumentation__.Notify(239388)
		}
		__antithesis_instrumentation__.Notify(239382)
		if len(more) == 0 {
			__antithesis_instrumentation__.Notify(239389)
			return errors.Errorf("looking up post-split descriptor returned first range: %+v", rs[0])
		} else {
			__antithesis_instrumentation__.Notify(239390)
		}
		__antithesis_instrumentation__.Notify(239383)
		leftRangeDesc = more[0]
		rightRangeDesc = rs[0]

		if !leftRangeDesc.EndKey.Equal(rightRangeDesc.StartKey) {
			__antithesis_instrumentation__.Notify(239391)
			return errors.Errorf(
				"inconsistent left (%v) and right (%v) descriptors", leftRangeDesc, rightRangeDesc,
			)
		} else {
			__antithesis_instrumentation__.Notify(239392)
		}
		__antithesis_instrumentation__.Notify(239384)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(239393)
		if len(wrappedMsg) > 0 {
			__antithesis_instrumentation__.Notify(239395)
			err = errors.Wrapf(err, "%s", wrappedMsg)
		} else {
			__antithesis_instrumentation__.Notify(239396)
		}
		__antithesis_instrumentation__.Notify(239394)
		return roachpb.RangeDescriptor{}, roachpb.RangeDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(239397)
	}
	__antithesis_instrumentation__.Notify(239377)

	return leftRangeDesc, rightRangeDesc, nil
}

func (ts *TestServer) SplitRange(
	splitKey roachpb.Key,
) (roachpb.RangeDescriptor, roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(239398)
	return ts.SplitRangeWithExpiration(splitKey, hlc.MaxTimestamp)
}

type LeaseInfo struct {
	cur, next roachpb.Lease
}

func (l LeaseInfo) Current() roachpb.Lease {
	__antithesis_instrumentation__.Notify(239399)
	return l.cur
}

func (l LeaseInfo) CurrentOrProspective() roachpb.Lease {
	__antithesis_instrumentation__.Notify(239400)
	if !l.next.Empty() {
		__antithesis_instrumentation__.Notify(239402)
		return l.next
	} else {
		__antithesis_instrumentation__.Notify(239403)
	}
	__antithesis_instrumentation__.Notify(239401)
	return l.cur
}

type LeaseInfoOpt int

const (
	AllowQueryToBeForwardedToDifferentNode LeaseInfoOpt = iota

	QueryLocalNodeOnly
)

func (ts *TestServer) GetRangeLease(
	ctx context.Context, key roachpb.Key, queryPolicy LeaseInfoOpt,
) (_ LeaseInfo, now hlc.ClockTimestamp, _ error) {
	__antithesis_instrumentation__.Notify(239404)
	leaseReq := roachpb.LeaseInfoRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: key,
		},
	}
	leaseResp, pErr := kv.SendWrappedWith(
		ctx,
		ts.DB().NonTransactionalSender(),
		roachpb.Header{

			ReadConsistency: roachpb.INCONSISTENT,
			RoutingPolicy:   roachpb.RoutingPolicy_NEAREST,
		},
		&leaseReq,
	)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(239408)
		return LeaseInfo{}, hlc.ClockTimestamp{}, pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(239409)
	}
	__antithesis_instrumentation__.Notify(239405)

	resp := leaseResp.(*roachpb.LeaseInfoResponse)
	if queryPolicy == QueryLocalNodeOnly && func() bool {
		__antithesis_instrumentation__.Notify(239410)
		return resp.EvaluatedBy != ts.GetFirstStoreID() == true
	}() == true {
		__antithesis_instrumentation__.Notify(239411)

		return LeaseInfo{}, hlc.ClockTimestamp{}, errors.Errorf(
			"request not evaluated locally; evaluated by s%d instead of local s%d",
			resp.EvaluatedBy, ts.GetFirstStoreID())
	} else {
		__antithesis_instrumentation__.Notify(239412)
	}
	__antithesis_instrumentation__.Notify(239406)
	var l LeaseInfo
	if resp.CurrentLease != nil {
		__antithesis_instrumentation__.Notify(239413)
		l.cur = *resp.CurrentLease
		l.next = resp.Lease
	} else {
		__antithesis_instrumentation__.Notify(239414)
		l.cur = resp.Lease
	}
	__antithesis_instrumentation__.Notify(239407)
	return l, ts.Clock().NowAsClockTimestamp(), nil
}

func (ts *TestServer) ExecutorConfig() interface{} {
	__antithesis_instrumentation__.Notify(239415)
	return *ts.sqlServer.execCfg
}

func (ts *TestServer) TracerI() interface{} {
	__antithesis_instrumentation__.Notify(239416)
	return ts.Tracer()
}

func (ts *TestServer) Tracer() *tracing.Tracer {
	__antithesis_instrumentation__.Notify(239417)
	return ts.node.storeCfg.AmbientCtx.Tracer
}

func (ts *TestServer) GCSystemLog(
	ctx context.Context, table string, timestampLowerBound, timestampUpperBound time.Time,
) (time.Time, int64, error) {
	__antithesis_instrumentation__.Notify(239418)
	return ts.gcSystemLog(ctx, table, timestampLowerBound, timestampUpperBound)
}

func (ts *TestServer) ForceTableGC(
	ctx context.Context, database, table string, timestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(239419)
	tableIDQuery := `
 SELECT tables.id FROM system.namespace tables
   JOIN system.namespace dbs ON dbs.id = tables."parentID"
   WHERE dbs.name = $1 AND tables.name = $2
 `
	row, err := ts.sqlServer.internalExecutor.QueryRowEx(
		ctx, "resolve-table-id", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		tableIDQuery, database, table)
	if err != nil {
		__antithesis_instrumentation__.Notify(239423)
		return err
	} else {
		__antithesis_instrumentation__.Notify(239424)
	}
	__antithesis_instrumentation__.Notify(239420)
	if row == nil {
		__antithesis_instrumentation__.Notify(239425)
		return errors.Errorf("table not found")
	} else {
		__antithesis_instrumentation__.Notify(239426)
	}
	__antithesis_instrumentation__.Notify(239421)
	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(239427)
		return errors.AssertionFailedf("expected 1 column from internal query")
	} else {
		__antithesis_instrumentation__.Notify(239428)
	}
	__antithesis_instrumentation__.Notify(239422)
	tableID := uint32(*row[0].(*tree.DInt))
	tblKey := keys.SystemSQLCodec.TablePrefix(tableID)
	gcr := roachpb.GCRequest{
		RequestHeader: roachpb.RequestHeader{
			Key:    tblKey,
			EndKey: tblKey.PrefixEnd(),
		},
		Threshold: timestamp,
	}
	_, pErr := kv.SendWrapped(ctx, ts.distSender, &gcr)
	return pErr.GoError()
}

func (ts *TestServer) ScratchRange() (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(239429)
	_, desc, err := ts.ScratchRangeEx()
	if err != nil {
		__antithesis_instrumentation__.Notify(239431)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(239432)
	}
	__antithesis_instrumentation__.Notify(239430)
	return desc.StartKey.AsRawKey(), nil
}

func (ts *TestServer) ScratchRangeEx() (roachpb.RangeDescriptor, roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(239433)
	scratchKey := keys.ScratchRangeMin
	return ts.SplitRange(scratchKey)
}

func (ts *TestServer) ScratchRangeWithExpirationLease() (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(239434)
	_, desc, err := ts.ScratchRangeWithExpirationLeaseEx()
	if err != nil {
		__antithesis_instrumentation__.Notify(239436)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(239437)
	}
	__antithesis_instrumentation__.Notify(239435)
	return desc.StartKey.AsRawKey(), nil
}

func (ts *TestServer) ScratchRangeWithExpirationLeaseEx() (
	roachpb.RangeDescriptor,
	roachpb.RangeDescriptor,
	error,
) {
	__antithesis_instrumentation__.Notify(239438)
	scratchKey := roachpb.Key(bytes.Join([][]byte{keys.SystemPrefix,
		roachpb.RKey("\x00aaa-testing")}, nil))
	return ts.SplitRange(scratchKey)
}

func (ts *TestServer) MetricsRecorder() *status.MetricsRecorder {
	__antithesis_instrumentation__.Notify(239439)
	return ts.node.recorder
}

func (ts *TestServer) CollectionFactory() interface{} {
	__antithesis_instrumentation__.Notify(239440)
	return ts.sqlServer.execCfg.CollectionFactory
}

func (ts *TestServer) SystemTableIDResolver() interface{} {
	__antithesis_instrumentation__.Notify(239441)
	return ts.sqlServer.execCfg.SystemTableIDResolver
}

func (ts *TestServer) SpanConfigKVSubscriber() interface{} {
	__antithesis_instrumentation__.Notify(239442)
	return ts.node.storeCfg.SpanConfigSubscriber
}

func (ts *TestServer) SystemConfigProvider() config.SystemConfigProvider {
	__antithesis_instrumentation__.Notify(239443)
	return ts.node.storeCfg.SystemConfigProvider
}

type testServerFactoryImpl struct{}

var TestServerFactory = testServerFactoryImpl{}

func (testServerFactoryImpl) New(params base.TestServerArgs) (interface{}, error) {
	__antithesis_instrumentation__.Notify(239444)
	cfg := makeTestConfigFromParams(params)
	ts := &TestServer{Cfg: &cfg, params: params}

	if params.Stopper == nil {
		__antithesis_instrumentation__.Notify(239450)
		params.Stopper = stop.NewStopper()
	} else {
		__antithesis_instrumentation__.Notify(239451)
	}
	__antithesis_instrumentation__.Notify(239445)

	if !params.PartOfCluster {
		__antithesis_instrumentation__.Notify(239452)
		ts.Cfg.DefaultZoneConfig.NumReplicas = proto.Int32(1)
	} else {
		__antithesis_instrumentation__.Notify(239453)
	}
	__antithesis_instrumentation__.Notify(239446)

	ctx := context.Background()
	if err := ts.Cfg.InitNode(ctx); err != nil {
		__antithesis_instrumentation__.Notify(239454)
		params.Stopper.Stop(ctx)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(239455)
	}
	__antithesis_instrumentation__.Notify(239447)

	var err error
	ts.Server, err = NewServer(*ts.Cfg, params.Stopper)
	if err != nil {
		__antithesis_instrumentation__.Notify(239456)
		params.Stopper.Stop(ctx)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(239457)
	}
	__antithesis_instrumentation__.Notify(239448)

	ts.rpcContext.BreakerFactory = func() *circuit.Breaker {
		__antithesis_instrumentation__.Notify(239458)
		return circuit.NewBreakerWithOptions(&circuit.Options{
			BackOff: &backoff.ZeroBackOff{},
		})
	}
	__antithesis_instrumentation__.Notify(239449)

	ts.Cfg = &ts.Server.cfg

	ts.httpTestServer = &httpTestServer{}
	ts.httpTestServer.t.authentication = ts.Server.authentication
	ts.httpTestServer.t.sqlServer = ts.Server.sqlServer

	return ts, nil
}
