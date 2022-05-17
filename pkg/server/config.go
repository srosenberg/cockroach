package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/ts"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/cockroachdb/redact"
)

const (
	DefaultCacheSize         = 128 << 20
	defaultSQLMemoryPoolSize = 128 << 20
	defaultScanInterval      = 10 * time.Minute
	defaultScanMinIdleTime   = 10 * time.Millisecond
	defaultScanMaxIdleTime   = 1 * time.Second

	DefaultStorePath = "cockroach-data"

	DefaultSQLNodeStorePathPrefix = "cockroach-data-tenant-"

	TempDirPrefix = "cockroach-temp"

	TempDirsRecordFilename = "temp-dirs-record.txt"
	defaultEventLogEnabled = true

	maximumMaxClockOffset = 5 * time.Second

	minimumNetworkFileDescriptors     = 256
	recommendedNetworkFileDescriptors = 5000

	defaultSQLTableStatCacheSize = 256

	defaultSQLQueryCacheSize = 8 * 1024 * 1024
)

var productionSettingsWebpage = fmt.Sprintf(
	"please see %s for more details",
	docs.URL("recommended-production-settings.html"),
)

type MaxOffsetType time.Duration

func (mo *MaxOffsetType) Type() string {
	__antithesis_instrumentation__.Notify(189952)
	return "MaxOffset"
}

func (mo *MaxOffsetType) Set(v string) error {
	__antithesis_instrumentation__.Notify(189953)
	nanos, err := time.ParseDuration(v)
	if err != nil {
		__antithesis_instrumentation__.Notify(189956)
		return err
	} else {
		__antithesis_instrumentation__.Notify(189957)
	}
	__antithesis_instrumentation__.Notify(189954)
	if nanos > maximumMaxClockOffset {
		__antithesis_instrumentation__.Notify(189958)
		return errors.Errorf("%s is not a valid max offset, must be less than %v.", v, maximumMaxClockOffset)
	} else {
		__antithesis_instrumentation__.Notify(189959)
	}
	__antithesis_instrumentation__.Notify(189955)
	*mo = MaxOffsetType(nanos)
	return nil
}

func (mo *MaxOffsetType) String() string {
	__antithesis_instrumentation__.Notify(189960)
	return time.Duration(*mo).String()
}

type BaseConfig struct {
	Settings *cluster.Settings
	*base.Config

	Tracer *tracing.Tracer

	idProvider *idProvider

	IDContainer *base.NodeIDContainer

	ClusterIDContainer *base.ClusterIDContainer

	AmbientCtx log.AmbientContext

	MaxOffset MaxOffsetType

	GoroutineDumpDirName string

	HeapProfileDirName string

	CPUProfileDirName string

	InflightTraceDirName string

	DefaultZoneConfig zonepb.ZoneConfig

	Locality roachpb.Locality

	StorageEngine enginepb.EngineType

	SpanConfigsDisabled bool

	TestingKnobs base.TestingKnobs

	EnableWebSessionAuthentication bool

	EnableDemoLoginEndpoint bool
}

func MakeBaseConfig(st *cluster.Settings, tr *tracing.Tracer) BaseConfig {
	__antithesis_instrumentation__.Notify(189961)
	if tr == nil {
		__antithesis_instrumentation__.Notify(189963)
		panic("nil Tracer")
	} else {
		__antithesis_instrumentation__.Notify(189964)
	}
	__antithesis_instrumentation__.Notify(189962)
	idsProvider := &idProvider{
		clusterID: &base.ClusterIDContainer{},
		serverID:  &base.NodeIDContainer{},
	}
	disableWebLogin := envutil.EnvOrDefaultBool("COCKROACH_DISABLE_WEB_LOGIN", false)
	baseCfg := BaseConfig{
		Tracer:                         tr,
		idProvider:                     idsProvider,
		IDContainer:                    idsProvider.serverID,
		ClusterIDContainer:             idsProvider.clusterID,
		AmbientCtx:                     log.MakeServerAmbientContext(tr, idsProvider),
		Config:                         new(base.Config),
		Settings:                       st,
		MaxOffset:                      MaxOffsetType(base.DefaultMaxClockOffset),
		DefaultZoneConfig:              zonepb.DefaultZoneConfig(),
		StorageEngine:                  storage.DefaultStorageEngine,
		EnableWebSessionAuthentication: !disableWebLogin,
	}

	baseCfg.AmbientCtx.AddLogTag("n", baseCfg.IDContainer)
	baseCfg.InitDefaults()
	return baseCfg
}

type Config struct {
	BaseConfig
	KVConfig
	SQLConfig
}

type KVConfig struct {
	base.RaftConfig

	Stores base.StoreSpecList

	Attrs string

	JoinList base.JoinListType

	JoinPreferSRVRecords bool

	RetryOptions retry.Options

	CacheSize int64

	TimeSeriesServerConfig ts.ServerConfig

	NodeAttributes roachpb.Attributes

	GossipBootstrapAddresses []util.UnresolvedAddr

	Linearizable bool

	ScanInterval time.Duration

	ScanMinIdleTime time.Duration

	ScanMaxIdleTime time.Duration

	DefaultSystemZoneConfig zonepb.ZoneConfig

	LocalityAddresses []roachpb.LocalityAddress

	EventLogEnabled bool

	ReadyFn func(waitForInit bool)

	DelayedBootstrapFn func()

	enginesCreated bool
}

func MakeKVConfig(storeSpec base.StoreSpec) KVConfig {
	__antithesis_instrumentation__.Notify(189965)
	kvCfg := KVConfig{
		DefaultSystemZoneConfig: zonepb.DefaultSystemZoneConfig(),
		CacheSize:               DefaultCacheSize,
		ScanInterval:            defaultScanInterval,
		ScanMinIdleTime:         defaultScanMinIdleTime,
		ScanMaxIdleTime:         defaultScanMaxIdleTime,
		EventLogEnabled:         defaultEventLogEnabled,
		Stores: base.StoreSpecList{
			Specs: []base.StoreSpec{storeSpec},
		},
	}
	kvCfg.RaftConfig.SetDefaults()
	return kvCfg
}

type SQLConfig struct {
	TenantID roachpb.TenantID

	SocketFile string

	TempStorageConfig base.TempStorageConfig

	ExternalIODirConfig base.ExternalIODirConfig

	MemoryPoolSize int64

	TableStatCacheSize int

	QueryCacheSize int64

	TenantKVAddrs []string
}

func MakeSQLConfig(tenID roachpb.TenantID, tempStorageCfg base.TempStorageConfig) SQLConfig {
	__antithesis_instrumentation__.Notify(189966)
	sqlCfg := SQLConfig{
		TenantID:           tenID,
		MemoryPoolSize:     defaultSQLMemoryPoolSize,
		TableStatCacheSize: defaultSQLTableStatCacheSize,
		QueryCacheSize:     defaultSQLQueryCacheSize,
		TempStorageConfig:  tempStorageCfg,
	}
	return sqlCfg
}

func setOpenFileLimit(physicalStoreCount int) (uint64, error) {
	__antithesis_instrumentation__.Notify(189967)
	return setOpenFileLimitInner(physicalStoreCount)
}

func SetOpenFileLimitForOneStore() (uint64, error) {
	__antithesis_instrumentation__.Notify(189968)
	return setOpenFileLimit(1)
}

func MakeConfig(ctx context.Context, st *cluster.Settings) Config {
	__antithesis_instrumentation__.Notify(189969)
	storeSpec, err := base.NewStoreSpec(DefaultStorePath)
	if err != nil {
		__antithesis_instrumentation__.Notify(189972)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(189973)
	}
	__antithesis_instrumentation__.Notify(189970)
	tempStorageCfg := base.TempStorageConfigFromEnv(
		ctx, st, storeSpec, "", base.DefaultTempStorageMaxSizeBytes)

	sqlCfg := MakeSQLConfig(roachpb.SystemTenantID, tempStorageCfg)
	tr := tracing.NewTracerWithOpt(ctx, tracing.WithClusterSettings(&st.SV))

	st.Version.SetOnChange(func(ctx context.Context, newVersion clusterversion.ClusterVersion) {
		__antithesis_instrumentation__.Notify(189974)
		tr.SetBackwardsCompatibilityWith211(!newVersion.IsActive(clusterversion.TraceIDDoesntImplyStructuredRecording))
	})
	__antithesis_instrumentation__.Notify(189971)
	baseCfg := MakeBaseConfig(st, tr)
	kvCfg := MakeKVConfig(storeSpec)

	cfg := Config{
		BaseConfig: baseCfg,
		KVConfig:   kvCfg,
		SQLConfig:  sqlCfg,
	}

	return cfg
}

func (cfg *Config) String() string {
	__antithesis_instrumentation__.Notify(189975)
	var buf bytes.Buffer

	w := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)
	fmt.Fprintln(w, "max offset\t", cfg.MaxOffset)
	fmt.Fprintln(w, "cache size\t", humanizeutil.IBytes(cfg.CacheSize))
	fmt.Fprintln(w, "SQL memory pool size\t", humanizeutil.IBytes(cfg.MemoryPoolSize))
	fmt.Fprintln(w, "scan interval\t", cfg.ScanInterval)
	fmt.Fprintln(w, "scan min idle time\t", cfg.ScanMinIdleTime)
	fmt.Fprintln(w, "scan max idle time\t", cfg.ScanMaxIdleTime)
	fmt.Fprintln(w, "event log enabled\t", cfg.EventLogEnabled)
	if cfg.Linearizable {
		__antithesis_instrumentation__.Notify(189978)
		fmt.Fprintln(w, "linearizable\t", cfg.Linearizable)
	} else {
		__antithesis_instrumentation__.Notify(189979)
	}
	__antithesis_instrumentation__.Notify(189976)
	if !cfg.SpanConfigsDisabled {
		__antithesis_instrumentation__.Notify(189980)
		fmt.Fprintln(w, "span configs enabled\t", !cfg.SpanConfigsDisabled)
	} else {
		__antithesis_instrumentation__.Notify(189981)
	}
	__antithesis_instrumentation__.Notify(189977)
	_ = w.Flush()

	return buf.String()
}

func (cfg *Config) Report(ctx context.Context) {
	__antithesis_instrumentation__.Notify(189982)
	if memSize, err := status.GetTotalMemory(ctx); err != nil {
		__antithesis_instrumentation__.Notify(189984)
		log.Infof(ctx, "unable to retrieve system total memory: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(189985)
		log.Infof(ctx, "system total memory: %s", humanizeutil.IBytes(memSize))
	}
	__antithesis_instrumentation__.Notify(189983)
	log.Infof(ctx, "server configuration:\n%s", cfg)
}

type Engines []storage.Engine

func (e *Engines) Close() {
	__antithesis_instrumentation__.Notify(189986)
	for _, eng := range *e {
		__antithesis_instrumentation__.Notify(189988)
		eng.Close()
	}
	__antithesis_instrumentation__.Notify(189987)
	*e = nil
}

func (cfg *Config) CreateEngines(ctx context.Context) (Engines, error) {
	__antithesis_instrumentation__.Notify(189989)
	engines := Engines(nil)
	defer engines.Close()

	if cfg.enginesCreated {
		__antithesis_instrumentation__.Notify(189997)
		return Engines{}, errors.Errorf("engines already created")
	} else {
		__antithesis_instrumentation__.Notify(189998)
	}
	__antithesis_instrumentation__.Notify(189990)
	cfg.enginesCreated = true
	details := []redact.RedactableString{redact.Sprintf("Pebble cache size: %s", humanizeutil.IBytes(cfg.CacheSize))}
	pebbleCache := pebble.NewCache(cfg.CacheSize)
	defer pebbleCache.Unref()

	var physicalStores int
	for _, spec := range cfg.Stores.Specs {
		__antithesis_instrumentation__.Notify(189999)
		if !spec.InMemory {
			__antithesis_instrumentation__.Notify(190000)
			physicalStores++
		} else {
			__antithesis_instrumentation__.Notify(190001)
		}
	}
	__antithesis_instrumentation__.Notify(189991)
	openFileLimitPerStore, err := setOpenFileLimit(physicalStores)
	if err != nil {
		__antithesis_instrumentation__.Notify(190002)
		return Engines{}, err
	} else {
		__antithesis_instrumentation__.Notify(190003)
	}
	__antithesis_instrumentation__.Notify(189992)

	log.Event(ctx, "initializing engines")

	var tableCache *pebble.TableCache
	if physicalStores > 0 {
		__antithesis_instrumentation__.Notify(190004)
		perStoreLimit := pebble.TableCacheSize(int(openFileLimitPerStore))
		totalFileLimit := perStoreLimit * physicalStores
		tableCache = pebble.NewTableCache(pebbleCache, runtime.GOMAXPROCS(0), totalFileLimit)
	} else {
		__antithesis_instrumentation__.Notify(190005)
	}
	__antithesis_instrumentation__.Notify(189993)

	skipSizeCheck := cfg.TestingKnobs.Store != nil && func() bool {
		__antithesis_instrumentation__.Notify(190006)
		return cfg.TestingKnobs.Store.(*kvserver.StoreTestingKnobs).SkipMinSizeCheck == true
	}() == true
	for i, spec := range cfg.Stores.Specs {
		__antithesis_instrumentation__.Notify(190007)
		log.Eventf(ctx, "initializing %+v", spec)
		var sizeInBytes = spec.Size.InBytes
		if spec.InMemory {
			__antithesis_instrumentation__.Notify(190008)
			if spec.Size.Percent > 0 {
				__antithesis_instrumentation__.Notify(190011)
				sysMem, err := status.GetTotalMemory(ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(190013)
					return Engines{}, errors.Errorf("could not retrieve system memory")
				} else {
					__antithesis_instrumentation__.Notify(190014)
				}
				__antithesis_instrumentation__.Notify(190012)
				sizeInBytes = int64(float64(sysMem) * spec.Size.Percent / 100)
			} else {
				__antithesis_instrumentation__.Notify(190015)
			}
			__antithesis_instrumentation__.Notify(190009)
			if sizeInBytes != 0 && func() bool {
				__antithesis_instrumentation__.Notify(190016)
				return !skipSizeCheck == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(190017)
				return sizeInBytes < base.MinimumStoreSize == true
			}() == true {
				__antithesis_instrumentation__.Notify(190018)
				return Engines{}, errors.Errorf("%f%% of memory is only %s bytes, which is below the minimum requirement of %s",
					spec.Size.Percent, humanizeutil.IBytes(sizeInBytes), humanizeutil.IBytes(base.MinimumStoreSize))
			} else {
				__antithesis_instrumentation__.Notify(190019)
			}
			__antithesis_instrumentation__.Notify(190010)
			details = append(details, redact.Sprintf("store %d: in-memory, size %s",
				i, humanizeutil.IBytes(sizeInBytes)))
			if spec.StickyInMemoryEngineID != "" {
				__antithesis_instrumentation__.Notify(190020)
				if cfg.TestingKnobs.Server == nil {
					__antithesis_instrumentation__.Notify(190024)
					return Engines{}, errors.AssertionFailedf("Could not create a sticky " +
						"engine no server knobs available to get a registry. " +
						"Please use Knobs.Server.StickyEngineRegistry to provide one.")
				} else {
					__antithesis_instrumentation__.Notify(190025)
				}
				__antithesis_instrumentation__.Notify(190021)
				knobs := cfg.TestingKnobs.Server.(*TestingKnobs)
				if knobs.StickyEngineRegistry == nil {
					__antithesis_instrumentation__.Notify(190026)
					return Engines{}, errors.Errorf("Could not create a sticky " +
						"engine no registry available. Please use " +
						"Knobs.Server.StickyEngineRegistry to provide one.")
				} else {
					__antithesis_instrumentation__.Notify(190027)
				}
				__antithesis_instrumentation__.Notify(190022)
				e, err := knobs.StickyEngineRegistry.GetOrCreateStickyInMemEngine(ctx, cfg, spec)
				if err != nil {
					__antithesis_instrumentation__.Notify(190028)
					return Engines{}, err
				} else {
					__antithesis_instrumentation__.Notify(190029)
				}
				__antithesis_instrumentation__.Notify(190023)
				details = append(details, redact.Sprintf("store %d: %+v", i, e.Properties()))
				engines = append(engines, e)
			} else {
				__antithesis_instrumentation__.Notify(190030)
				e, err := storage.Open(ctx,
					storage.InMemory(),
					storage.Attributes(spec.Attributes),
					storage.CacheSize(cfg.CacheSize),
					storage.MaxSize(sizeInBytes),
					storage.EncryptionAtRest(spec.EncryptionOptions),
					storage.Settings(cfg.Settings))
				if err != nil {
					__antithesis_instrumentation__.Notify(190032)
					return Engines{}, err
				} else {
					__antithesis_instrumentation__.Notify(190033)
				}
				__antithesis_instrumentation__.Notify(190031)
				engines = append(engines, e)
			}
		} else {
			__antithesis_instrumentation__.Notify(190034)
			if err := vfs.Default.MkdirAll(spec.Path, 0755); err != nil {
				__antithesis_instrumentation__.Notify(190042)
				return Engines{}, errors.Wrap(err, "creating store directory")
			} else {
				__antithesis_instrumentation__.Notify(190043)
			}
			__antithesis_instrumentation__.Notify(190035)
			du, err := vfs.Default.GetDiskUsage(spec.Path)
			if err != nil {
				__antithesis_instrumentation__.Notify(190044)
				return Engines{}, errors.Wrap(err, "retrieving disk usage")
			} else {
				__antithesis_instrumentation__.Notify(190045)
			}
			__antithesis_instrumentation__.Notify(190036)
			if spec.Size.Percent > 0 {
				__antithesis_instrumentation__.Notify(190046)
				sizeInBytes = int64(float64(du.TotalBytes) * spec.Size.Percent / 100)
			} else {
				__antithesis_instrumentation__.Notify(190047)
			}
			__antithesis_instrumentation__.Notify(190037)
			if sizeInBytes != 0 && func() bool {
				__antithesis_instrumentation__.Notify(190048)
				return !skipSizeCheck == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(190049)
				return sizeInBytes < base.MinimumStoreSize == true
			}() == true {
				__antithesis_instrumentation__.Notify(190050)
				return Engines{}, errors.Errorf("%f%% of %s's total free space is only %s bytes, which is below the minimum requirement of %s",
					spec.Size.Percent, spec.Path, humanizeutil.IBytes(sizeInBytes), humanizeutil.IBytes(base.MinimumStoreSize))
			} else {
				__antithesis_instrumentation__.Notify(190051)
			}
			__antithesis_instrumentation__.Notify(190038)

			details = append(details, redact.Sprintf("store %d: max size %s, max open file limit %d",
				i, humanizeutil.IBytes(sizeInBytes), openFileLimitPerStore))

			storageConfig := base.StorageConfig{
				Attrs:             spec.Attributes,
				Dir:               spec.Path,
				MaxSize:           sizeInBytes,
				BallastSize:       storage.BallastSizeBytes(spec, du),
				Settings:          cfg.Settings,
				UseFileRegistry:   spec.UseFileRegistry,
				EncryptionOptions: spec.EncryptionOptions,
			}
			pebbleConfig := storage.PebbleConfig{
				StorageConfig: storageConfig,
				Opts:          storage.DefaultPebbleOptions(),
			}
			pebbleConfig.Opts.Cache = pebbleCache
			pebbleConfig.Opts.TableCache = tableCache
			pebbleConfig.Opts.MaxOpenFiles = int(openFileLimitPerStore)

			if len(spec.PebbleOptions) > 0 {
				__antithesis_instrumentation__.Notify(190052)
				err := pebbleConfig.Opts.Parse(spec.PebbleOptions, &pebble.ParseHooks{})
				if err != nil {
					__antithesis_instrumentation__.Notify(190053)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(190054)
				}
			} else {
				__antithesis_instrumentation__.Notify(190055)
			}
			__antithesis_instrumentation__.Notify(190039)
			if len(spec.RocksDBOptions) > 0 {
				__antithesis_instrumentation__.Notify(190056)
				return nil, errors.Errorf("store %d: using Pebble storage engine but StoreSpec provides RocksDB options", i)
			} else {
				__antithesis_instrumentation__.Notify(190057)
			}
			__antithesis_instrumentation__.Notify(190040)
			eng, err := storage.NewPebble(ctx, pebbleConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(190058)
				return Engines{}, err
			} else {
				__antithesis_instrumentation__.Notify(190059)
			}
			__antithesis_instrumentation__.Notify(190041)
			details = append(details, redact.Sprintf("store %d: %+v", i, eng.Properties()))
			engines = append(engines, eng)
		}
	}
	__antithesis_instrumentation__.Notify(189994)

	if tableCache != nil {
		__antithesis_instrumentation__.Notify(190060)

		if err := tableCache.Unref(); err != nil {
			__antithesis_instrumentation__.Notify(190061)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(190062)
		}
	} else {
		__antithesis_instrumentation__.Notify(190063)
	}
	__antithesis_instrumentation__.Notify(189995)

	log.Infof(ctx, "%d storage engine%s initialized",
		len(engines), util.Pluralize(int64(len(engines))))
	for _, s := range details {
		__antithesis_instrumentation__.Notify(190064)
		log.Infof(ctx, "%v", s)
	}
	__antithesis_instrumentation__.Notify(189996)
	enginesCopy := engines
	engines = nil
	return enginesCopy, nil
}

func (cfg *Config) InitNode(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(190065)
	cfg.readEnvironmentVariables()

	cfg.NodeAttributes = parseAttributes(cfg.Attrs)

	addresses, err := cfg.parseGossipBootstrapAddresses(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(190068)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190069)
	}
	__antithesis_instrumentation__.Notify(190066)
	if len(addresses) > 0 {
		__antithesis_instrumentation__.Notify(190070)
		cfg.GossipBootstrapAddresses = addresses
	} else {
		__antithesis_instrumentation__.Notify(190071)
	}
	__antithesis_instrumentation__.Notify(190067)

	return nil
}

func (cfg *Config) FilterGossipBootstrapAddresses(ctx context.Context) []util.UnresolvedAddr {
	__antithesis_instrumentation__.Notify(190072)
	var listen, advert net.Addr
	listen = util.NewUnresolvedAddr("tcp", cfg.Addr)
	advert = util.NewUnresolvedAddr("tcp", cfg.AdvertiseAddr)
	filtered := make([]util.UnresolvedAddr, 0, len(cfg.GossipBootstrapAddresses))
	addrs := make([]string, 0, len(cfg.GossipBootstrapAddresses))

	for _, addr := range cfg.GossipBootstrapAddresses {
		__antithesis_instrumentation__.Notify(190075)
		if addr.String() == advert.String() || func() bool {
			__antithesis_instrumentation__.Notify(190076)
			return addr.String() == listen.String() == true
		}() == true {
			__antithesis_instrumentation__.Notify(190077)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(190078)
				log.Infof(ctx, "skipping -join address %q, because a node cannot join itself", addr)
			} else {
				__antithesis_instrumentation__.Notify(190079)
			}
		} else {
			__antithesis_instrumentation__.Notify(190080)
			filtered = append(filtered, addr)
			addrs = append(addrs, addr.String())
		}
	}
	__antithesis_instrumentation__.Notify(190073)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(190081)
		log.Infof(ctx, "initial addresses: %v", addrs)
	} else {
		__antithesis_instrumentation__.Notify(190082)
	}
	__antithesis_instrumentation__.Notify(190074)
	return filtered
}

func (cfg *BaseConfig) RequireWebSession() bool {
	__antithesis_instrumentation__.Notify(190083)
	return !cfg.Insecure && func() bool {
		__antithesis_instrumentation__.Notify(190084)
		return cfg.EnableWebSessionAuthentication == true
	}() == true
}

func (cfg *Config) readEnvironmentVariables() {
	__antithesis_instrumentation__.Notify(190085)
	cfg.SpanConfigsDisabled = envutil.EnvOrDefaultBool("COCKROACH_DISABLE_SPAN_CONFIGS", cfg.SpanConfigsDisabled)
	cfg.Linearizable = envutil.EnvOrDefaultBool("COCKROACH_EXPERIMENTAL_LINEARIZABLE", cfg.Linearizable)
	cfg.ScanInterval = envutil.EnvOrDefaultDuration("COCKROACH_SCAN_INTERVAL", cfg.ScanInterval)
	cfg.ScanMinIdleTime = envutil.EnvOrDefaultDuration("COCKROACH_SCAN_MIN_IDLE_TIME", cfg.ScanMinIdleTime)
	cfg.ScanMaxIdleTime = envutil.EnvOrDefaultDuration("COCKROACH_SCAN_MAX_IDLE_TIME", cfg.ScanMaxIdleTime)
}

func (cfg *Config) parseGossipBootstrapAddresses(
	ctx context.Context,
) ([]util.UnresolvedAddr, error) {
	__antithesis_instrumentation__.Notify(190086)
	var bootstrapAddresses []util.UnresolvedAddr
	for _, address := range cfg.JoinList {
		__antithesis_instrumentation__.Notify(190088)
		if address == "" {
			__antithesis_instrumentation__.Notify(190091)
			continue
		} else {
			__antithesis_instrumentation__.Notify(190092)
		}
		__antithesis_instrumentation__.Notify(190089)

		if cfg.JoinPreferSRVRecords {
			__antithesis_instrumentation__.Notify(190093)

			srvAddrs, err := netutil.SRV(ctx, address)
			if err != nil {
				__antithesis_instrumentation__.Notify(190095)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(190096)
			}
			__antithesis_instrumentation__.Notify(190094)

			if len(srvAddrs) > 0 {
				__antithesis_instrumentation__.Notify(190097)
				for _, sa := range srvAddrs {
					__antithesis_instrumentation__.Notify(190099)
					bootstrapAddresses = append(bootstrapAddresses,
						util.MakeUnresolvedAddrWithDefaults("tcp", sa, base.DefaultPort))
				}
				__antithesis_instrumentation__.Notify(190098)
				continue
			} else {
				__antithesis_instrumentation__.Notify(190100)
			}
		} else {
			__antithesis_instrumentation__.Notify(190101)
		}
		__antithesis_instrumentation__.Notify(190090)

		bootstrapAddresses = append(bootstrapAddresses,
			util.MakeUnresolvedAddrWithDefaults("tcp", address, base.DefaultPort))
	}
	__antithesis_instrumentation__.Notify(190087)

	return bootstrapAddresses, nil
}

func parseAttributes(attrsStr string) roachpb.Attributes {
	__antithesis_instrumentation__.Notify(190102)
	var filtered []string
	for _, attr := range strings.Split(attrsStr, ":") {
		__antithesis_instrumentation__.Notify(190104)
		if len(attr) != 0 {
			__antithesis_instrumentation__.Notify(190105)
			filtered = append(filtered, attr)
		} else {
			__antithesis_instrumentation__.Notify(190106)
		}
	}
	__antithesis_instrumentation__.Notify(190103)
	return roachpb.Attributes{Attrs: filtered}
}

type idProvider struct {
	clusterID *base.ClusterIDContainer

	clusterStr atomic.Value

	tenantID roachpb.TenantID

	tenantStr atomic.Value

	serverID *base.NodeIDContainer

	serverStr atomic.Value
}

var _ log.ServerIdentificationPayload = (*idProvider)(nil)

func (s *idProvider) ServerIdentityString(key log.ServerIdentificationKey) string {
	__antithesis_instrumentation__.Notify(190107)
	switch key {
	case log.IdentifyClusterID:
		__antithesis_instrumentation__.Notify(190109)
		c := s.clusterStr.Load()
		cs, ok := c.(string)
		if !ok {
			__antithesis_instrumentation__.Notify(190118)
			cid := s.clusterID.Get()
			if cid != uuid.Nil {
				__antithesis_instrumentation__.Notify(190119)
				cs = cid.String()
				s.clusterStr.Store(cs)
			} else {
				__antithesis_instrumentation__.Notify(190120)
			}
		} else {
			__antithesis_instrumentation__.Notify(190121)
		}
		__antithesis_instrumentation__.Notify(190110)
		return cs

	case log.IdentifyTenantID:
		__antithesis_instrumentation__.Notify(190111)
		t := s.tenantStr.Load()
		ts, ok := t.(string)
		if !ok {
			__antithesis_instrumentation__.Notify(190122)
			tid := s.tenantID
			if tid.IsSet() {
				__antithesis_instrumentation__.Notify(190123)
				ts = strconv.FormatUint(tid.ToUint64(), 10)
				s.tenantStr.Store(ts)
			} else {
				__antithesis_instrumentation__.Notify(190124)
			}
		} else {
			__antithesis_instrumentation__.Notify(190125)
		}
		__antithesis_instrumentation__.Notify(190112)
		return ts

	case log.IdentifyInstanceID:
		__antithesis_instrumentation__.Notify(190113)

		if !s.tenantID.IsSet() {
			__antithesis_instrumentation__.Notify(190126)
			return ""
		} else {
			__antithesis_instrumentation__.Notify(190127)
		}
		__antithesis_instrumentation__.Notify(190114)
		return s.maybeMemoizeServerID()

	case log.IdentifyKVNodeID:
		__antithesis_instrumentation__.Notify(190115)

		if s.tenantID.IsSet() {
			__antithesis_instrumentation__.Notify(190128)
			return ""
		} else {
			__antithesis_instrumentation__.Notify(190129)
		}
		__antithesis_instrumentation__.Notify(190116)
		return s.maybeMemoizeServerID()
	default:
		__antithesis_instrumentation__.Notify(190117)
	}
	__antithesis_instrumentation__.Notify(190108)

	return ""
}

func (s *idProvider) SetTenant(tenantID roachpb.TenantID) {
	__antithesis_instrumentation__.Notify(190130)
	if !tenantID.IsSet() {
		__antithesis_instrumentation__.Notify(190133)
		panic("programming error: invalid tenant ID")
	} else {
		__antithesis_instrumentation__.Notify(190134)
	}
	__antithesis_instrumentation__.Notify(190131)
	if s.tenantID.IsSet() {
		__antithesis_instrumentation__.Notify(190135)
		panic("programming error: provider already set for tenant server")
	} else {
		__antithesis_instrumentation__.Notify(190136)
	}
	__antithesis_instrumentation__.Notify(190132)
	s.tenantID = tenantID
}

func (s *idProvider) maybeMemoizeServerID() string {
	__antithesis_instrumentation__.Notify(190137)
	si := s.serverStr.Load()
	sis, ok := si.(string)
	if !ok {
		__antithesis_instrumentation__.Notify(190139)
		sid := s.serverID.Get()
		if sid != 0 {
			__antithesis_instrumentation__.Notify(190140)
			sis = strconv.FormatUint(uint64(sid), 10)
			s.serverStr.Store(sis)
		} else {
			__antithesis_instrumentation__.Notify(190141)
		}
	} else {
		__antithesis_instrumentation__.Notify(190142)
	}
	__antithesis_instrumentation__.Notify(190138)
	return sis
}
