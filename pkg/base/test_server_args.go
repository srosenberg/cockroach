package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

type TestServerArgs struct {
	Knobs TestingKnobs

	*cluster.Settings
	RaftConfig

	PartOfCluster bool

	Listener net.Listener

	Addr string

	SQLAddr string

	TenantAddr *string

	HTTPAddr string

	DisableTLSForHTTP bool

	JoinAddr string

	StoreSpecs []StoreSpec

	Locality roachpb.Locality

	TempStorageConfig TempStorageConfig

	ExternalIODir string

	ExternalIODirConfig ExternalIODirConfig

	Insecure                    bool
	RetryOptions                retry.Options
	SocketFile                  string
	ScanInterval                time.Duration
	ScanMinIdleTime             time.Duration
	ScanMaxIdleTime             time.Duration
	SSLCertsDir                 string
	TimeSeriesQueryWorkerMax    int
	TimeSeriesQueryMemoryBudget int64
	SQLMemoryPoolSize           int64
	CacheSize                   int64

	NoAutoInitializeCluster bool

	UseDatabase string

	ClusterName string

	Stopper *stop.Stopper

	DisableEventLog bool

	DisableWebSessionAuthentication bool

	EnableDemoLoginEndpoint bool

	Tracer *tracing.Tracer

	TracingDefault tracing.TracingMode

	TraceDir string

	DisableSpanConfigs bool
}

type TestClusterArgs struct {
	ServerArgs TestServerArgs

	ReplicationMode TestClusterReplicationMode

	ParallelStart bool

	ServerArgsPerNode map[int]TestServerArgs
}

var (
	DefaultTestStoreSpec = StoreSpec{
		InMemory: true,
		Size: SizeSpec{
			InBytes: 512 << 20,
		},
	}
)

func DefaultTestTempStorageConfig(st *cluster.Settings) TempStorageConfig {
	__antithesis_instrumentation__.Notify(1761)
	return DefaultTestTempStorageConfigWithSize(st, DefaultInMemTempStorageMaxSizeBytes)
}

func DefaultTestTempStorageConfigWithSize(
	st *cluster.Settings, maxSizeBytes int64,
) TempStorageConfig {
	__antithesis_instrumentation__.Notify(1762)
	monitor := mon.NewMonitor(
		"in-mem temp storage",
		mon.DiskResource,
		nil,
		nil,
		1024*1024,
		maxSizeBytes/10,
		st,
	)
	monitor.Start(context.Background(), nil, mon.MakeStandaloneBudget(maxSizeBytes))
	return TempStorageConfig{
		InMemory: true,
		Mon:      monitor,
		Settings: st,
	}
}

type TestClusterReplicationMode int

const (
	ReplicationAuto TestClusterReplicationMode = iota

	ReplicationManual
)

type TestTenantArgs struct {
	TenantID roachpb.TenantID

	Existing bool

	Settings *cluster.Settings

	AllowSettingClusterSettings bool

	Stopper *stop.Stopper

	TestingKnobs TestingKnobs

	ForceInsecure bool

	MemoryPoolSize int64

	TempStorageConfig *TempStorageConfig

	ExternalIODirConfig ExternalIODirConfig

	ExternalIODir string

	UseDatabase string

	SkipTenantCheck bool

	Locality roachpb.Locality

	SSLCertsDir string

	StartingSQLPort int

	StartingHTTPPort int

	TracingDefault tracing.TracingMode

	RPCHeartbeatInterval time.Duration

	GoroutineDumpDirName string

	HeapProfileDirName string
}
