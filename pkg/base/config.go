package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
)

const (
	defaultInsecure = false
	defaultUser     = security.RootUser
	httpScheme      = "http"
	httpsScheme     = "https"

	DefaultPort = "26257"

	DefaultHTTPPort = "8080"

	defaultAddr     = ":" + DefaultPort
	defaultSQLAddr  = ":" + DefaultPort
	defaultHTTPAddr = ":" + DefaultHTTPPort

	NetworkTimeout = 3 * time.Second

	DefaultCertsDirectory = "${HOME}/.cockroach-certs"

	defaultRaftTickInterval = 200 * time.Millisecond

	defaultRangeLeaseRaftElectionTimeoutMultiplier = 3

	DefaultMetricsSampleInterval = 10 * time.Second

	defaultRaftHeartbeatIntervalTicks = 5

	defaultRPCHeartbeatInterval = 3 * time.Second

	defaultRangeLeaseRenewalFraction = 0.5

	livenessRenewalFraction = 0.5

	DefaultDescriptorLeaseDuration = 5 * time.Minute

	DefaultDescriptorLeaseJitterFraction = 0.05

	DefaultDescriptorLeaseRenewalTimeout = time.Minute
)

func DefaultHistogramWindowInterval() time.Duration {
	__antithesis_instrumentation__.Notify(1400)
	const defHWI = 6 * DefaultMetricsSampleInterval

	if defHWI < DefaultMetricsSampleInterval {
		__antithesis_instrumentation__.Notify(1402)
		return DefaultMetricsSampleInterval
	} else {
		__antithesis_instrumentation__.Notify(1403)
	}
	__antithesis_instrumentation__.Notify(1401)
	return defHWI
}

var (
	defaultRaftElectionTimeoutTicks = envutil.EnvOrDefaultInt(
		"COCKROACH_RAFT_ELECTION_TIMEOUT_TICKS", 15)

	defaultRaftLogTruncationThreshold = envutil.EnvOrDefaultInt64(
		"COCKROACH_RAFT_LOG_TRUNCATION_THRESHOLD", 16<<20)

	defaultRaftMaxSizePerMsg = envutil.EnvOrDefaultInt(
		"COCKROACH_RAFT_MAX_SIZE_PER_MSG", 32<<10)

	defaultRaftMaxCommittedSizePerReady = envutil.EnvOrDefaultInt(
		"COCKROACH_RAFT_MAX_COMMITTED_SIZE_PER_READY", 64<<20)

	defaultRaftMaxInflightMsgs = envutil.EnvOrDefaultInt(
		"COCKROACH_RAFT_MAX_INFLIGHT_MSGS", 128)
)

type Config struct {
	Insecure bool

	AcceptSQLWithoutTLS bool

	SSLCAKey string

	SSLCertsDir string

	User security.SQLUsername

	Addr string

	AdvertiseAddr string

	ClusterName string

	DisableClusterNameVerification bool

	SplitListenSQL bool

	SQLAddr string

	SQLAdvertiseAddr string

	HTTPAddr string

	DisableTLSForHTTP bool

	HTTPAdvertiseAddr string

	RPCHeartbeatInterval time.Duration

	ClockDevicePath string

	AutoInitializeCluster bool
}

func (*Config) HistogramWindowInterval() time.Duration {
	__antithesis_instrumentation__.Notify(1404)
	return DefaultHistogramWindowInterval()
}

func (cfg *Config) InitDefaults() {
	__antithesis_instrumentation__.Notify(1405)
	cfg.Insecure = defaultInsecure
	cfg.User = security.MakeSQLUsernameFromPreNormalizedString(defaultUser)
	cfg.Addr = defaultAddr
	cfg.AdvertiseAddr = cfg.Addr
	cfg.HTTPAddr = defaultHTTPAddr
	cfg.DisableTLSForHTTP = false
	cfg.HTTPAdvertiseAddr = ""
	cfg.SplitListenSQL = false
	cfg.SQLAddr = defaultSQLAddr
	cfg.SQLAdvertiseAddr = cfg.SQLAddr
	cfg.SSLCertsDir = DefaultCertsDirectory
	cfg.RPCHeartbeatInterval = defaultRPCHeartbeatInterval
	cfg.ClusterName = ""
	cfg.DisableClusterNameVerification = false
	cfg.ClockDevicePath = ""
	cfg.AcceptSQLWithoutTLS = false
}

func (cfg *Config) HTTPRequestScheme() string {
	__antithesis_instrumentation__.Notify(1406)
	if cfg.Insecure || func() bool {
		__antithesis_instrumentation__.Notify(1408)
		return cfg.DisableTLSForHTTP == true
	}() == true {
		__antithesis_instrumentation__.Notify(1409)
		return httpScheme
	} else {
		__antithesis_instrumentation__.Notify(1410)
	}
	__antithesis_instrumentation__.Notify(1407)
	return httpsScheme
}

func (cfg *Config) AdminURL() *url.URL {
	__antithesis_instrumentation__.Notify(1411)
	return &url.URL{
		Scheme: cfg.HTTPRequestScheme(),
		Host:   cfg.HTTPAdvertiseAddr,
	}
}

type RaftConfig struct {
	RaftTickInterval time.Duration

	RaftElectionTimeoutTicks int

	RaftHeartbeatIntervalTicks int

	RangeLeaseRaftElectionTimeoutMultiplier float64

	RangeLeaseRenewalFraction float64

	RaftLogTruncationThreshold int64

	RaftProposalQuota int64

	RaftMaxUncommittedEntriesSize uint64

	RaftMaxSizePerMsg uint64

	RaftMaxCommittedSizePerReady uint64

	RaftMaxInflightMsgs int

	RaftDelaySplitToSuppressSnapshotTicks int
}

func (cfg *RaftConfig) SetDefaults() {
	__antithesis_instrumentation__.Notify(1412)
	if cfg.RaftTickInterval == 0 {
		__antithesis_instrumentation__.Notify(1425)
		cfg.RaftTickInterval = defaultRaftTickInterval
	} else {
		__antithesis_instrumentation__.Notify(1426)
	}
	__antithesis_instrumentation__.Notify(1413)
	if cfg.RaftElectionTimeoutTicks == 0 {
		__antithesis_instrumentation__.Notify(1427)
		cfg.RaftElectionTimeoutTicks = defaultRaftElectionTimeoutTicks
	} else {
		__antithesis_instrumentation__.Notify(1428)
	}
	__antithesis_instrumentation__.Notify(1414)
	if cfg.RaftHeartbeatIntervalTicks == 0 {
		__antithesis_instrumentation__.Notify(1429)
		cfg.RaftHeartbeatIntervalTicks = defaultRaftHeartbeatIntervalTicks
	} else {
		__antithesis_instrumentation__.Notify(1430)
	}
	__antithesis_instrumentation__.Notify(1415)
	if cfg.RangeLeaseRaftElectionTimeoutMultiplier == 0 {
		__antithesis_instrumentation__.Notify(1431)
		cfg.RangeLeaseRaftElectionTimeoutMultiplier = defaultRangeLeaseRaftElectionTimeoutMultiplier
	} else {
		__antithesis_instrumentation__.Notify(1432)
	}
	__antithesis_instrumentation__.Notify(1416)
	if cfg.RangeLeaseRenewalFraction == 0 {
		__antithesis_instrumentation__.Notify(1433)
		cfg.RangeLeaseRenewalFraction = defaultRangeLeaseRenewalFraction
	} else {
		__antithesis_instrumentation__.Notify(1434)
	}
	__antithesis_instrumentation__.Notify(1417)

	if cfg.RaftLogTruncationThreshold == 0 {
		__antithesis_instrumentation__.Notify(1435)
		cfg.RaftLogTruncationThreshold = defaultRaftLogTruncationThreshold
	} else {
		__antithesis_instrumentation__.Notify(1436)
	}
	__antithesis_instrumentation__.Notify(1418)
	if cfg.RaftProposalQuota == 0 {
		__antithesis_instrumentation__.Notify(1437)

		cfg.RaftProposalQuota = cfg.RaftLogTruncationThreshold / 2
	} else {
		__antithesis_instrumentation__.Notify(1438)
	}
	__antithesis_instrumentation__.Notify(1419)
	if cfg.RaftMaxUncommittedEntriesSize == 0 {
		__antithesis_instrumentation__.Notify(1439)

		cfg.RaftMaxUncommittedEntriesSize = uint64(2 * cfg.RaftProposalQuota)
	} else {
		__antithesis_instrumentation__.Notify(1440)
	}
	__antithesis_instrumentation__.Notify(1420)
	if cfg.RaftMaxSizePerMsg == 0 {
		__antithesis_instrumentation__.Notify(1441)
		cfg.RaftMaxSizePerMsg = uint64(defaultRaftMaxSizePerMsg)
	} else {
		__antithesis_instrumentation__.Notify(1442)
	}
	__antithesis_instrumentation__.Notify(1421)
	if cfg.RaftMaxCommittedSizePerReady == 0 {
		__antithesis_instrumentation__.Notify(1443)
		cfg.RaftMaxCommittedSizePerReady = uint64(defaultRaftMaxCommittedSizePerReady)
	} else {
		__antithesis_instrumentation__.Notify(1444)
	}
	__antithesis_instrumentation__.Notify(1422)
	if cfg.RaftMaxInflightMsgs == 0 {
		__antithesis_instrumentation__.Notify(1445)
		cfg.RaftMaxInflightMsgs = defaultRaftMaxInflightMsgs
	} else {
		__antithesis_instrumentation__.Notify(1446)
	}
	__antithesis_instrumentation__.Notify(1423)
	if cfg.RaftDelaySplitToSuppressSnapshotTicks == 0 {
		__antithesis_instrumentation__.Notify(1447)

		cfg.RaftDelaySplitToSuppressSnapshotTicks = 3*cfg.RaftElectionTimeoutTicks + 200
	} else {
		__antithesis_instrumentation__.Notify(1448)
	}
	__antithesis_instrumentation__.Notify(1424)

	if cfg.RaftProposalQuota > int64(cfg.RaftMaxUncommittedEntriesSize) {
		__antithesis_instrumentation__.Notify(1449)
		panic("raft proposal quota should not be above max uncommitted entries size")
	} else {
		__antithesis_instrumentation__.Notify(1450)
	}
}

func (cfg RaftConfig) RaftElectionTimeout() time.Duration {
	__antithesis_instrumentation__.Notify(1451)
	return time.Duration(cfg.RaftElectionTimeoutTicks) * cfg.RaftTickInterval
}

func (cfg RaftConfig) RangeLeaseDurations() (rangeLeaseActive, rangeLeaseRenewal time.Duration) {
	__antithesis_instrumentation__.Notify(1452)
	rangeLeaseActive = time.Duration(cfg.RangeLeaseRaftElectionTimeoutMultiplier *
		float64(cfg.RaftElectionTimeout()))
	renewalFraction := cfg.RangeLeaseRenewalFraction
	if renewalFraction == -1 {
		__antithesis_instrumentation__.Notify(1454)
		renewalFraction = 0
	} else {
		__antithesis_instrumentation__.Notify(1455)
	}
	__antithesis_instrumentation__.Notify(1453)
	rangeLeaseRenewal = time.Duration(float64(rangeLeaseActive) * renewalFraction)
	return
}

func (cfg RaftConfig) RangeLeaseActiveDuration() time.Duration {
	__antithesis_instrumentation__.Notify(1456)
	rangeLeaseActive, _ := cfg.RangeLeaseDurations()
	return rangeLeaseActive
}

func (cfg RaftConfig) RangeLeaseRenewalDuration() time.Duration {
	__antithesis_instrumentation__.Notify(1457)
	_, rangeLeaseRenewal := cfg.RangeLeaseDurations()
	return rangeLeaseRenewal
}

func (cfg RaftConfig) NodeLivenessDurations() (livenessActive, livenessRenewal time.Duration) {
	__antithesis_instrumentation__.Notify(1458)
	livenessActive = cfg.RangeLeaseActiveDuration()
	livenessRenewal = time.Duration(float64(livenessActive) * livenessRenewalFraction)
	return
}

func (cfg RaftConfig) SentinelGossipTTL() time.Duration {
	__antithesis_instrumentation__.Notify(1459)
	return cfg.RangeLeaseActiveDuration() / 2
}

func DefaultRetryOptions() retry.Options {
	__antithesis_instrumentation__.Notify(1460)

	return retry.Options{
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2,
	}
}

type StorageConfig struct {
	Attrs roachpb.Attributes

	Dir string

	MustExist bool

	MaxSize int64

	BallastSize int64

	Settings *cluster.Settings

	UseFileRegistry bool

	EncryptionOptions []byte
}

func (sc StorageConfig) IsEncrypted() bool {
	__antithesis_instrumentation__.Notify(1461)
	return len(sc.EncryptionOptions) > 0
}

const (
	DefaultTempStorageMaxSizeBytes = 32 * 1024 * 1024 * 1024

	DefaultInMemTempStorageMaxSizeBytes = 100 * 1024 * 1024
)

type TempStorageConfig struct {
	InMemory bool

	Path string

	Mon *mon.BytesMonitor

	Spec StoreSpec

	Settings *cluster.Settings
}

type ExternalIODirConfig struct {
	DisableHTTP bool

	DisableImplicitCredentials bool

	DisableOutbound bool

	EnableNonAdminImplicitAndArbitraryOutbound bool
}

func TempStorageConfigFromEnv(
	ctx context.Context,
	st *cluster.Settings,
	useStore StoreSpec,
	parentDir string,
	maxSizeBytes int64,
) TempStorageConfig {
	__antithesis_instrumentation__.Notify(1462)
	inMem := parentDir == "" && func() bool {
		__antithesis_instrumentation__.Notify(1464)
		return useStore.InMemory == true
	}() == true
	var monitorName string
	if inMem {
		__antithesis_instrumentation__.Notify(1465)
		monitorName = "in-mem temp storage"
	} else {
		__antithesis_instrumentation__.Notify(1466)
		monitorName = "temp disk storage"
	}
	__antithesis_instrumentation__.Notify(1463)
	monitor := mon.NewMonitor(
		monitorName,
		mon.DiskResource,
		nil,
		nil,
		1024*1024,
		maxSizeBytes/10,
		st,
	)
	monitor.Start(ctx, nil, mon.MakeStandaloneBudget(maxSizeBytes))
	return TempStorageConfig{
		InMemory: inMem,
		Mon:      monitor,
		Spec:     useStore,
		Settings: st,
	}
}
