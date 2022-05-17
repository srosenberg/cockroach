package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/blobs"
	"github.com/cockroachdb/cockroach/pkg/blobs/blobspb"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/bulk"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvtenant"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/migration/migrationcluster"
	"github.com/cockroachdb/cockroach/pkg/migration/migrationmanager"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/diagnostics"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/settingswatcher"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/server/systemconfigwatcher"
	"github.com/cockroachdb/cockroach/pkg/server/tracedumper"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfiglimiter"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigmanager"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigreconciler"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigsplitter"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigsqltranslator"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigsqlwatcher"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/hydratedtables"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec"
	"github.com/cockroachdb/cockroach/pkg/sql/contention"
	"github.com/cockroachdb/cockroach/pkg/sql/descmetadata"
	"github.com/cockroachdb/cockroach/pkg/sql/distsql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/gcjob/gcjobnotifier"
	"github.com/cockroachdb/cockroach/pkg/sql/optionalnodeliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire"
	"github.com/cockroachdb/cockroach/pkg/sql/querycache"
	"github.com/cockroachdb/cockroach/pkg/sql/scheduledlogging"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessioninit"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance/instanceprovider"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness/slprovider"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/stmtdiagnostics"
	"github.com/cockroachdb/cockroach/pkg/startupmigrations"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/collector"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/service"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingservicepb"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/marusama/semaphore"
	"google.golang.org/grpc"
)

type SQLServer struct {
	ambientCtx       log.AmbientContext
	stopper          *stop.Stopper
	sqlIDContainer   *base.SQLIDContainer
	pgServer         *pgwire.Server
	distSQLServer    *distsql.ServerImpl
	execCfg          *sql.ExecutorConfig
	cfg              *BaseConfig
	internalExecutor *sql.InternalExecutor
	leaseMgr         *lease.Manager
	blobService      *blobs.Service
	tracingService   *service.Service
	tenantConnect    kvtenant.Connector

	sessionRegistry        *sql.SessionRegistry
	jobRegistry            *jobs.Registry
	startupMigrationsMgr   *startupmigrations.Manager
	statsRefresher         *stats.Refresher
	temporaryObjectCleaner *sql.TemporaryObjectCleaner
	internalMemMetrics     sql.MemoryMetrics

	sqlMemMetrics           sql.MemoryMetrics
	stmtDiagnosticsRegistry *stmtdiagnostics.Registry

	sqlLivenessSessionID           sqlliveness.SessionID
	sqlLivenessProvider            sqlliveness.Provider
	sqlInstanceProvider            sqlinstance.Provider
	metricsRegistry                *metric.Registry
	diagnosticsReporter            *diagnostics.Reporter
	spanconfigMgr                  *spanconfigmanager.Manager
	spanconfigSQLTranslatorFactory *spanconfigsqltranslator.Factory
	spanconfigSQLWatcher           *spanconfigsqlwatcher.SQLWatcher
	settingsWatcher                *settingswatcher.SettingsWatcher

	systemConfigWatcher *systemconfigwatcher.Cache

	isMeta1Leaseholder func(context.Context, hlc.ClockTimestamp) (bool, error)

	pgL net.Listener

	connManager netutil.Server

	isReady syncutil.AtomicBool
}

type sqlServerOptionalKVArgs struct {
	nodesStatusServer serverpb.OptionalNodesStatusServer

	nodeLiveness optionalnodeliveness.Container

	gossip gossip.OptionalGossip

	grpcServer *grpc.Server

	isMeta1Leaseholder func(context.Context, hlc.ClockTimestamp) (bool, error)

	nodeIDContainer *base.SQLIDContainer

	externalStorage        cloud.ExternalStorageFactory
	externalStorageFromURI cloud.ExternalStorageFromURIFactory

	sqlSQLResponseAdmissionQ *admission.WorkQueue

	spanConfigKVAccessor spanconfig.KVAccessor

	kvStoresIterator kvserverbase.StoresIterator
}

type sqlServerOptionalTenantArgs struct {
	tenantConnect kvtenant.Connector

	advertiseAddr string
}

type sqlServerArgs struct {
	sqlServerOptionalKVArgs
	sqlServerOptionalTenantArgs

	*SQLConfig
	*BaseConfig

	stopper *stop.Stopper

	clock *hlc.Clock

	runtime *status.RuntimeStatSampler

	rpcContext *rpc.Context

	nodeDescs kvcoord.NodeDescStore

	systemConfigWatcher *systemconfigwatcher.Cache

	spanConfigAccessor spanconfig.KVAccessor

	nodeDialer *nodedialer.Dialer

	podNodeDialer *nodedialer.Dialer

	distSender *kvcoord.DistSender

	db *kv.DB

	registry *metric.Registry

	recorder *status.MetricsRecorder

	sessionRegistry *sql.SessionRegistry

	flowScheduler *flowinfra.FlowScheduler

	circularInternalExecutor *sql.InternalExecutor

	sqlLivenessProvider sqlliveness.Provider

	sqlInstanceProvider sqlinstance.Provider

	circularJobRegistry *jobs.Registry
	jobAdoptionStopFile string

	protectedtsProvider protectedts.Provider

	sqlStatusServer serverpb.SQLStatusServer

	rangeFeedFactory *rangefeed.Factory

	regionsServer serverpb.RegionsServer

	tenantStatusServer serverpb.TenantStatusServer

	tenantUsageServer multitenant.TenantUsageServer

	costController multitenant.TenantSideCostController

	monitorAndMetrics monitorAndMetrics

	settingsStorage settingswatcher.Storage

	grpc *grpcServer
}

type monitorAndMetrics struct {
	rootSQLMemoryMonitor *mon.BytesMonitor
	rootSQLMetrics       sql.BaseMemoryMetrics
}

type monitorAndMetricsOptions struct {
	memoryPoolSize          int64
	histogramWindowInterval time.Duration
	settings                *cluster.Settings
}

func newRootSQLMemoryMonitor(opts monitorAndMetricsOptions) monitorAndMetrics {
	__antithesis_instrumentation__.Notify(195878)
	rootSQLMetrics := sql.MakeBaseMemMetrics("root", opts.histogramWindowInterval)

	rootSQLMemoryMonitor := mon.NewMonitor(
		"root",
		mon.MemoryResource,
		rootSQLMetrics.CurBytesCount,
		rootSQLMetrics.MaxBytesHist,
		-1,
		math.MaxInt64,
		opts.settings,
	)

	rootSQLMemoryMonitor.Start(
		context.Background(), nil, mon.MakeStandaloneBudget(opts.memoryPoolSize))
	return monitorAndMetrics{
		rootSQLMemoryMonitor: rootSQLMemoryMonitor,
		rootSQLMetrics:       rootSQLMetrics,
	}
}

func newSQLServer(ctx context.Context, cfg sqlServerArgs) (*SQLServer, error) {
	__antithesis_instrumentation__.Notify(195879)

	if err := cfg.Config.ValidateAddrs(ctx); err != nil {
		__antithesis_instrumentation__.Notify(195917)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195918)
	}
	__antithesis_instrumentation__.Notify(195880)
	execCfg := &sql.ExecutorConfig{}
	codec := keys.MakeSQLCodec(cfg.SQLConfig.TenantID)
	if knobs := cfg.TestingKnobs.TenantTestingKnobs; knobs != nil {
		__antithesis_instrumentation__.Notify(195919)
		override := knobs.(*sql.TenantTestingKnobs).TenantIDCodecOverride
		if override != (roachpb.TenantID{}) {
			__antithesis_instrumentation__.Notify(195920)
			codec = keys.MakeSQLCodec(override)
		} else {
			__antithesis_instrumentation__.Notify(195921)
		}
	} else {
		__antithesis_instrumentation__.Notify(195922)
	}
	__antithesis_instrumentation__.Notify(195881)

	blobService, err := blobs.NewBlobService(cfg.Settings.ExternalIODir)
	if err != nil {
		__antithesis_instrumentation__.Notify(195923)
		return nil, errors.Wrap(err, "creating blob service")
	} else {
		__antithesis_instrumentation__.Notify(195924)
	}
	__antithesis_instrumentation__.Notify(195882)
	blobspb.RegisterBlobServer(cfg.grpcServer, blobService)

	tracingService := service.New(cfg.Tracer)
	tracingservicepb.RegisterTracingServer(cfg.grpcServer, tracingService)

	sqllivenessKnobs, _ := cfg.TestingKnobs.SQLLivenessKnobs.(*sqlliveness.TestingKnobs)
	cfg.sqlLivenessProvider = slprovider.New(
		cfg.AmbientCtx,
		cfg.stopper, cfg.clock, cfg.db, codec, cfg.Settings, sqllivenessKnobs,
	)
	cfg.sqlInstanceProvider = instanceprovider.New(
		cfg.stopper, cfg.db, codec, cfg.sqlLivenessProvider, cfg.advertiseAddr, cfg.rangeFeedFactory, cfg.clock,
	)

	if !codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(195925)

		addressResolver := func(nodeID roachpb.NodeID) (net.Addr, error) {
			__antithesis_instrumentation__.Notify(195927)
			if cfg.sqlInstanceProvider == nil {
				__antithesis_instrumentation__.Notify(195930)
				return nil, errors.Errorf("no sqlInstanceProvider")
			} else {
				__antithesis_instrumentation__.Notify(195931)
			}
			__antithesis_instrumentation__.Notify(195928)
			info, err := cfg.sqlInstanceProvider.GetInstance(cfg.rpcContext.MasterCtx, base.SQLInstanceID(nodeID))
			if err != nil {
				__antithesis_instrumentation__.Notify(195932)
				return nil, errors.Errorf("unable to look up descriptor for nsql%d", nodeID)
			} else {
				__antithesis_instrumentation__.Notify(195933)
			}
			__antithesis_instrumentation__.Notify(195929)
			return &util.UnresolvedAddr{AddressField: info.InstanceAddr}, nil
		}
		__antithesis_instrumentation__.Notify(195926)
		cfg.podNodeDialer = nodedialer.New(cfg.rpcContext, addressResolver)
	} else {
		__antithesis_instrumentation__.Notify(195934)
		cfg.podNodeDialer = cfg.nodeDialer
	}
	__antithesis_instrumentation__.Notify(195883)

	jobRegistry := cfg.circularJobRegistry
	{
		__antithesis_instrumentation__.Notify(195935)
		cfg.registry.AddMetricStruct(cfg.sqlLivenessProvider.Metrics())

		var jobsKnobs *jobs.TestingKnobs
		if cfg.TestingKnobs.JobsTestingKnobs != nil {
			__antithesis_instrumentation__.Notify(195937)
			jobsKnobs = cfg.TestingKnobs.JobsTestingKnobs.(*jobs.TestingKnobs)
		} else {
			__antithesis_instrumentation__.Notify(195938)
		}
		__antithesis_instrumentation__.Notify(195936)

		td := tracedumper.NewTraceDumper(ctx, cfg.InflightTraceDirName, cfg.Settings)
		*jobRegistry = *jobs.MakeRegistry(
			ctx,
			cfg.AmbientCtx,
			cfg.stopper,
			cfg.clock,
			cfg.db,
			cfg.circularInternalExecutor,
			cfg.rpcContext.LogicalClusterID,
			cfg.nodeIDContainer,
			cfg.sqlLivenessProvider,
			cfg.Settings,
			cfg.HistogramWindowInterval(),
			func(opName string, user security.SQLUsername) (interface{}, func()) {
				__antithesis_instrumentation__.Notify(195939)

				return sql.MakeJobExecContext(opName, user, &sql.MemoryMetrics{}, execCfg)
			},
			cfg.jobAdoptionStopFile,
			td,
			jobsKnobs,
		)
	}
	__antithesis_instrumentation__.Notify(195884)
	cfg.registry.AddMetricStruct(jobRegistry.MetricsStruct())

	var lmKnobs lease.ManagerTestingKnobs
	if leaseManagerTestingKnobs := cfg.TestingKnobs.SQLLeaseManager; leaseManagerTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195940)
		lmKnobs = *leaseManagerTestingKnobs.(*lease.ManagerTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195941)
	}
	__antithesis_instrumentation__.Notify(195885)

	leaseMgr := lease.NewLeaseManager(
		cfg.AmbientCtx,
		cfg.nodeIDContainer,
		cfg.db,
		cfg.clock,
		cfg.circularInternalExecutor,
		cfg.Settings,
		codec,
		lmKnobs,
		cfg.stopper,
		cfg.rangeFeedFactory,
	)
	cfg.registry.AddMetricStruct(leaseMgr.MetricsStruct())

	rootSQLMetrics := cfg.monitorAndMetrics.rootSQLMetrics
	cfg.registry.AddMetricStruct(rootSQLMetrics)

	internalMemMetrics := sql.MakeMemMetrics("internal", cfg.HistogramWindowInterval())
	cfg.registry.AddMetricStruct(internalMemMetrics)

	rootSQLMemoryMonitor := cfg.monitorAndMetrics.rootSQLMemoryMonitor

	bulkMemoryMonitor := mon.NewMonitorInheritWithLimit("bulk-mon", 0, rootSQLMemoryMonitor)
	bulkMetrics := bulk.MakeBulkMetrics(cfg.HistogramWindowInterval())
	cfg.registry.AddMetricStruct(bulkMetrics)
	bulkMemoryMonitor.SetMetrics(bulkMetrics.CurBytesCount, bulkMetrics.MaxBytesHist)
	bulkMemoryMonitor.Start(context.Background(), rootSQLMemoryMonitor, mon.BoundAccount{})

	backfillMemoryMonitor := execinfra.NewMonitor(ctx, bulkMemoryMonitor, "backfill-mon")
	backupMemoryMonitor := execinfra.NewMonitor(ctx, bulkMemoryMonitor, "backup-mon")

	serverCacheMemoryMonitor := mon.NewMonitorInheritWithLimit(
		"server-cache-mon", 0, rootSQLMemoryMonitor,
	)
	serverCacheMemoryMonitor.Start(context.Background(), rootSQLMemoryMonitor, mon.BoundAccount{})

	useStoreSpec := cfg.TempStorageConfig.Spec
	tempEngine, tempFS, err := storage.NewTempEngine(ctx, cfg.TempStorageConfig, useStoreSpec)
	if err != nil {
		__antithesis_instrumentation__.Notify(195942)
		return nil, errors.Wrap(err, "creating temp storage")
	} else {
		__antithesis_instrumentation__.Notify(195943)
	}
	__antithesis_instrumentation__.Notify(195886)
	cfg.stopper.AddCloser(tempEngine)

	cfg.stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(195944)
		useStore := cfg.TempStorageConfig.Spec
		var err error
		if useStore.InMemory {
			__antithesis_instrumentation__.Notify(195946)

			err = os.RemoveAll(cfg.TempStorageConfig.Path)
		} else {
			__antithesis_instrumentation__.Notify(195947)

			recordPath := filepath.Join(useStore.Path, TempDirsRecordFilename)
			err = fs.CleanupTempDirs(recordPath)
		}
		__antithesis_instrumentation__.Notify(195945)
		if err != nil {
			__antithesis_instrumentation__.Notify(195948)
			log.Errorf(ctx, "could not remove temporary store directory: %v", err.Error())
		} else {
			__antithesis_instrumentation__.Notify(195949)
		}
	}))
	__antithesis_instrumentation__.Notify(195887)

	distSQLMetrics := execinfra.MakeDistSQLMetrics(cfg.HistogramWindowInterval())
	cfg.registry.AddMetricStruct(distSQLMetrics)
	rowMetrics := sql.NewRowMetrics(false)
	cfg.registry.AddMetricStruct(rowMetrics)
	internalRowMetrics := sql.NewRowMetrics(true)
	cfg.registry.AddMetricStruct(internalRowMetrics)

	virtualSchemas, err := sql.NewVirtualSchemaHolder(ctx, cfg.Settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(195950)
		return nil, errors.Wrap(err, "creating virtual schema holder")
	} else {
		__antithesis_instrumentation__.Notify(195951)
	}
	__antithesis_instrumentation__.Notify(195888)

	hydratedTablesCache := hydratedtables.NewCache(cfg.Settings)
	cfg.registry.AddMetricStruct(hydratedTablesCache.Metrics())

	gcJobNotifier := gcjobnotifier.New(cfg.Settings, cfg.systemConfigWatcher, codec, cfg.stopper)

	var compactEngineSpanFunc tree.CompactEngineSpanFunc
	if !codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(195952)
		compactEngineSpanFunc = func(
			ctx context.Context, nodeID, storeID int32, startKey, endKey []byte,
		) error {
			__antithesis_instrumentation__.Notify(195953)
			return errorutil.UnsupportedWithMultiTenancy(errorutil.FeatureNotAvailableToNonSystemTenantsIssue)
		}
	} else {
		__antithesis_instrumentation__.Notify(195954)
		cli := kvserver.NewCompactEngineSpanClient(cfg.nodeDialer)
		compactEngineSpanFunc = cli.CompactEngineSpan
	}
	__antithesis_instrumentation__.Notify(195889)

	spanConfig := struct {
		manager              *spanconfigmanager.Manager
		sqlTranslatorFactory *spanconfigsqltranslator.Factory
		sqlWatcher           *spanconfigsqlwatcher.SQLWatcher
		splitter             spanconfig.Splitter
		limiter              spanconfig.Limiter
	}{}
	if codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(195955)
		spanConfig.limiter = spanconfiglimiter.NoopLimiter{}
		spanConfig.splitter = spanconfigsplitter.NoopSplitter{}
	} else {
		__antithesis_instrumentation__.Notify(195956)
		spanConfigKnobs, _ := cfg.TestingKnobs.SpanConfig.(*spanconfig.TestingKnobs)
		spanConfig.splitter = spanconfigsplitter.New(codec, spanConfigKnobs)
		spanConfig.limiter = spanconfiglimiter.New(
			cfg.circularInternalExecutor,
			cfg.Settings,
			spanConfigKnobs,
		)
	}
	__antithesis_instrumentation__.Notify(195890)

	collectionFactory := descs.NewCollectionFactory(
		cfg.Settings,
		leaseMgr,
		virtualSchemas,
		hydratedTablesCache,
		spanConfig.splitter,
		spanConfig.limiter,
	)

	clusterIDForSQL := cfg.rpcContext.LogicalClusterID

	distSQLCfg := execinfra.ServerConfig{
		AmbientContext:   cfg.AmbientCtx,
		Settings:         cfg.Settings,
		RuntimeStats:     cfg.runtime,
		LogicalClusterID: clusterIDForSQL,
		ClusterName:      cfg.ClusterName,
		NodeID:           cfg.nodeIDContainer,
		Locality:         cfg.Locality,
		Codec:            codec,
		DB:               cfg.db,
		Executor:         cfg.circularInternalExecutor,
		RPCContext:       cfg.rpcContext,
		Stopper:          cfg.stopper,

		TempStorage:     tempEngine,
		TempStoragePath: cfg.TempStorageConfig.Path,
		TempFS:          tempFS,

		VecFDSemaphore:    semaphore.New(envutil.EnvOrDefaultInt("COCKROACH_VEC_MAX_OPEN_FDS", colexec.VecMaxOpenFDsLimit)),
		ParentDiskMonitor: cfg.TempStorageConfig.Mon,
		BackfillerMonitor: backfillMemoryMonitor,
		BackupMonitor:     backupMemoryMonitor,

		ParentMemoryMonitor: rootSQLMemoryMonitor,
		BulkAdder: func(
			ctx context.Context, db *kv.DB, ts hlc.Timestamp, opts kvserverbase.BulkAdderOptions,
		) (kvserverbase.BulkAdder, error) {
			__antithesis_instrumentation__.Notify(195957)

			bulkMon := execinfra.NewMonitor(ctx, bulkMemoryMonitor, "bulk-adder-monitor")
			return bulk.MakeBulkAdder(ctx, db, cfg.distSender.RangeDescriptorCache(), cfg.Settings, ts, opts, bulkMon)
		},

		Metrics:            &distSQLMetrics,
		RowMetrics:         &rowMetrics,
		InternalRowMetrics: &internalRowMetrics,

		SQLLivenessReader: cfg.sqlLivenessProvider,
		JobRegistry:       jobRegistry,
		Gossip:            cfg.gossip,
		NodeDialer:        cfg.nodeDialer,
		PodNodeDialer:     cfg.podNodeDialer,
		LeaseManager:      leaseMgr,

		ExternalStorage:        cfg.externalStorage,
		ExternalStorageFromURI: cfg.externalStorageFromURI,

		DistSender:               cfg.distSender,
		RangeCache:               cfg.distSender.RangeDescriptorCache(),
		SQLSQLResponseAdmissionQ: cfg.sqlSQLResponseAdmissionQ,
		CollectionFactory:        collectionFactory,
	}
	__antithesis_instrumentation__.Notify(195891)
	cfg.TempStorageConfig.Mon.SetMetrics(distSQLMetrics.CurDiskBytesCount, distSQLMetrics.MaxDiskBytesHist)
	if distSQLTestingKnobs := cfg.TestingKnobs.DistSQL; distSQLTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195958)
		distSQLCfg.TestingKnobs = *distSQLTestingKnobs.(*execinfra.TestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195959)
	}
	__antithesis_instrumentation__.Notify(195892)
	if cfg.TestingKnobs.JobsTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195960)
		distSQLCfg.TestingKnobs.JobsTestingKnobs = cfg.TestingKnobs.JobsTestingKnobs
	} else {
		__antithesis_instrumentation__.Notify(195961)
	}
	__antithesis_instrumentation__.Notify(195893)

	distSQLServer := distsql.NewServer(ctx, distSQLCfg, cfg.flowScheduler)
	execinfrapb.RegisterDistSQLServer(cfg.grpcServer, distSQLServer)

	var sqlExecutorTestingKnobs sql.ExecutorTestingKnobs
	if k := cfg.TestingKnobs.SQLExecutor; k != nil {
		__antithesis_instrumentation__.Notify(195962)
		sqlExecutorTestingKnobs = *k.(*sql.ExecutorTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195963)
		sqlExecutorTestingKnobs = sql.ExecutorTestingKnobs{}
	}
	__antithesis_instrumentation__.Notify(195894)

	nodeInfo := sql.NodeInfo{
		AdminURL:         cfg.AdminURL,
		PGURL:            cfg.rpcContext.PGURL,
		LogicalClusterID: cfg.rpcContext.LogicalClusterID.Get,
		NodeID:           cfg.nodeIDContainer,
	}

	var isAvailable func(sqlInstanceID base.SQLInstanceID) bool
	nodeLiveness, hasNodeLiveness := cfg.nodeLiveness.Optional(47900)
	if hasNodeLiveness {
		__antithesis_instrumentation__.Notify(195964)

		isAvailable = func(sqlInstanceID base.SQLInstanceID) bool {
			__antithesis_instrumentation__.Notify(195965)
			return nodeLiveness.IsAvailable(roachpb.NodeID(sqlInstanceID))
		}
	} else {
		__antithesis_instrumentation__.Notify(195966)

		isAvailable = func(sqlInstanceID base.SQLInstanceID) bool {
			__antithesis_instrumentation__.Notify(195967)
			return true
		}
	}
	__antithesis_instrumentation__.Notify(195895)

	var traceCollector *collector.TraceCollector
	if hasNodeLiveness {
		__antithesis_instrumentation__.Notify(195968)
		traceCollector = collector.New(cfg.nodeDialer, nodeLiveness, cfg.Tracer)
	} else {
		__antithesis_instrumentation__.Notify(195969)
	}
	__antithesis_instrumentation__.Notify(195896)
	contentionMetrics := contention.NewMetrics()
	cfg.registry.AddMetricStruct(contentionMetrics)

	contentionRegistry := contention.NewRegistry(
		cfg.Settings,
		cfg.sqlStatusServer.TxnIDResolution,
		&contentionMetrics,
	)
	contentionRegistry.Start(ctx, cfg.stopper)

	*execCfg = sql.ExecutorConfig{
		Settings:                cfg.Settings,
		NodeInfo:                nodeInfo,
		Codec:                   codec,
		DefaultZoneConfig:       &cfg.DefaultZoneConfig,
		Locality:                cfg.Locality,
		AmbientCtx:              cfg.AmbientCtx,
		DB:                      cfg.db,
		Gossip:                  cfg.gossip,
		NodeLiveness:            cfg.nodeLiveness,
		SystemConfig:            cfg.systemConfigWatcher,
		MetricsRecorder:         cfg.recorder,
		DistSender:              cfg.distSender,
		RPCContext:              cfg.rpcContext,
		LeaseManager:            leaseMgr,
		TenantStatusServer:      cfg.tenantStatusServer,
		Clock:                   cfg.clock,
		DistSQLSrv:              distSQLServer,
		NodesStatusServer:       cfg.nodesStatusServer,
		SQLStatusServer:         cfg.sqlStatusServer,
		RegionsServer:           cfg.regionsServer,
		SessionRegistry:         cfg.sessionRegistry,
		ContentionRegistry:      contentionRegistry,
		SQLLiveness:             cfg.sqlLivenessProvider,
		JobRegistry:             jobRegistry,
		VirtualSchemas:          virtualSchemas,
		HistogramWindowInterval: cfg.HistogramWindowInterval(),
		RangeDescriptorCache:    cfg.distSender.RangeDescriptorCache(),
		RoleMemberCache:         sql.NewMembershipCache(serverCacheMemoryMonitor.MakeBoundAccount(), cfg.stopper),
		SessionInitCache:        sessioninit.NewCache(serverCacheMemoryMonitor.MakeBoundAccount(), cfg.stopper),
		RootMemoryMonitor:       rootSQLMemoryMonitor,
		TestingKnobs:            sqlExecutorTestingKnobs,
		CompactEngineSpanFunc:   compactEngineSpanFunc,
		TraceCollector:          traceCollector,
		TenantUsageServer:       cfg.tenantUsageServer,
		KVStoresIterator:        cfg.kvStoresIterator,

		DistSQLPlanner: sql.NewDistSQLPlanner(
			ctx,
			execinfra.Version,
			cfg.Settings,
			cfg.nodeIDContainer.SQLInstanceID(),
			cfg.rpcContext,
			distSQLServer,
			cfg.distSender,
			cfg.nodeDescs,
			cfg.gossip,
			cfg.stopper,
			isAvailable,
			cfg.nodeDialer,
			cfg.podNodeDialer,
			codec,
			cfg.sqlInstanceProvider,
		),

		TableStatsCache: stats.NewTableStatisticsCache(
			ctx,
			cfg.TableStatCacheSize,
			cfg.db,
			cfg.circularInternalExecutor,
			codec,
			cfg.Settings,
			cfg.rangeFeedFactory,
			collectionFactory,
		),

		QueryCache:                 querycache.New(cfg.QueryCacheSize),
		RowMetrics:                 &rowMetrics,
		InternalRowMetrics:         &internalRowMetrics,
		ProtectedTimestampProvider: cfg.protectedtsProvider,
		ExternalIODirConfig:        cfg.ExternalIODirConfig,
		GCJobNotifier:              gcJobNotifier,
		RangeFeedFactory:           cfg.rangeFeedFactory,
		CollectionFactory:          collectionFactory,
		SystemTableIDResolver:      descs.MakeSystemTableIDResolver(collectionFactory, cfg.circularInternalExecutor, cfg.db),
	}

	if sqlSchemaChangerTestingKnobs := cfg.TestingKnobs.SQLSchemaChanger; sqlSchemaChangerTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195970)
		execCfg.SchemaChangerTestingKnobs = sqlSchemaChangerTestingKnobs.(*sql.SchemaChangerTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195971)
		execCfg.SchemaChangerTestingKnobs = new(sql.SchemaChangerTestingKnobs)
	}
	__antithesis_instrumentation__.Notify(195897)
	if declarativeSchemaChangerTestingKnobs := cfg.TestingKnobs.SQLDeclarativeSchemaChanger; declarativeSchemaChangerTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195972)
		execCfg.DeclarativeSchemaChangerTestingKnobs = declarativeSchemaChangerTestingKnobs.(*scrun.TestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195973)
		execCfg.DeclarativeSchemaChangerTestingKnobs = new(scrun.TestingKnobs)
	}
	__antithesis_instrumentation__.Notify(195898)
	if sqlTypeSchemaChangerTestingKnobs := cfg.TestingKnobs.SQLTypeSchemaChanger; sqlTypeSchemaChangerTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195974)
		execCfg.TypeSchemaChangerTestingKnobs = sqlTypeSchemaChangerTestingKnobs.(*sql.TypeSchemaChangerTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195975)
		execCfg.TypeSchemaChangerTestingKnobs = new(sql.TypeSchemaChangerTestingKnobs)
	}
	__antithesis_instrumentation__.Notify(195899)
	execCfg.SchemaChangerMetrics = sql.NewSchemaChangerMetrics()
	cfg.registry.AddMetricStruct(execCfg.SchemaChangerMetrics)

	execCfg.FeatureFlagMetrics = featureflag.NewFeatureFlagMetrics()
	cfg.registry.AddMetricStruct(execCfg.FeatureFlagMetrics)

	if gcJobTestingKnobs := cfg.TestingKnobs.GCJob; gcJobTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195976)
		execCfg.GCJobTestingKnobs = gcJobTestingKnobs.(*sql.GCJobTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195977)
		execCfg.GCJobTestingKnobs = new(sql.GCJobTestingKnobs)
	}
	__antithesis_instrumentation__.Notify(195900)
	if distSQLRunTestingKnobs := cfg.TestingKnobs.DistSQL; distSQLRunTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195978)
		execCfg.DistSQLRunTestingKnobs = distSQLRunTestingKnobs.(*execinfra.TestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195979)
		execCfg.DistSQLRunTestingKnobs = new(execinfra.TestingKnobs)
	}
	__antithesis_instrumentation__.Notify(195901)
	if sqlEvalContext := cfg.TestingKnobs.SQLEvalContext; sqlEvalContext != nil {
		__antithesis_instrumentation__.Notify(195980)
		execCfg.EvalContextTestingKnobs = *sqlEvalContext.(*tree.EvalContextTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195981)
	}
	__antithesis_instrumentation__.Notify(195902)
	if pgwireKnobs := cfg.TestingKnobs.PGWireTestingKnobs; pgwireKnobs != nil {
		__antithesis_instrumentation__.Notify(195982)
		execCfg.PGWireTestingKnobs = pgwireKnobs.(*sql.PGWireTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195983)
	}
	__antithesis_instrumentation__.Notify(195903)
	if tenantKnobs := cfg.TestingKnobs.TenantTestingKnobs; tenantKnobs != nil {
		__antithesis_instrumentation__.Notify(195984)
		execCfg.TenantTestingKnobs = tenantKnobs.(*sql.TenantTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195985)
	}
	__antithesis_instrumentation__.Notify(195904)
	if backupRestoreKnobs := cfg.TestingKnobs.BackupRestore; backupRestoreKnobs != nil {
		__antithesis_instrumentation__.Notify(195986)
		execCfg.BackupRestoreTestingKnobs = backupRestoreKnobs.(*sql.BackupRestoreTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195987)
	}
	__antithesis_instrumentation__.Notify(195905)
	if ttlKnobs := cfg.TestingKnobs.TTL; ttlKnobs != nil {
		__antithesis_instrumentation__.Notify(195988)
		execCfg.TTLTestingKnobs = ttlKnobs.(*sql.TTLTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195989)
	}
	__antithesis_instrumentation__.Notify(195906)
	if sqlStatsKnobs := cfg.TestingKnobs.SQLStatsKnobs; sqlStatsKnobs != nil {
		__antithesis_instrumentation__.Notify(195990)
		execCfg.SQLStatsTestingKnobs = sqlStatsKnobs.(*sqlstats.TestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195991)
	}
	__antithesis_instrumentation__.Notify(195907)
	if telemetryLoggingKnobs := cfg.TestingKnobs.TelemetryLoggingKnobs; telemetryLoggingKnobs != nil {
		__antithesis_instrumentation__.Notify(195992)
		execCfg.TelemetryLoggingTestingKnobs = telemetryLoggingKnobs.(*sql.TelemetryLoggingTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195993)
	}
	__antithesis_instrumentation__.Notify(195908)
	if spanConfigKnobs := cfg.TestingKnobs.SpanConfig; spanConfigKnobs != nil {
		__antithesis_instrumentation__.Notify(195994)
		execCfg.SpanConfigTestingKnobs = spanConfigKnobs.(*spanconfig.TestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195995)
	}
	__antithesis_instrumentation__.Notify(195909)
	if capturedIndexUsageStatsKnobs := cfg.TestingKnobs.CapturedIndexUsageStatsKnobs; capturedIndexUsageStatsKnobs != nil {
		__antithesis_instrumentation__.Notify(195996)
		execCfg.CaptureIndexUsageStatsKnobs = capturedIndexUsageStatsKnobs.(*scheduledlogging.CaptureIndexUsageStatsTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195997)
	}
	__antithesis_instrumentation__.Notify(195910)

	statsRefresher := stats.MakeRefresher(
		cfg.AmbientCtx,
		cfg.Settings,
		cfg.circularInternalExecutor,
		execCfg.TableStatsCache,
		stats.DefaultAsOfTime,
	)
	execCfg.StatsRefresher = statsRefresher

	sqlMemMetrics := sql.MakeMemMetrics("sql", cfg.HistogramWindowInterval())
	pgServer := pgwire.MakeServer(
		cfg.AmbientCtx,
		cfg.Config,
		cfg.Settings,
		sqlMemMetrics,
		rootSQLMemoryMonitor,
		cfg.HistogramWindowInterval(),
		execCfg,
	)

	distSQLServer.ServerConfig.SQLStatsController = pgServer.SQLServer.GetSQLStatsController()
	distSQLServer.ServerConfig.IndexUsageStatsController = pgServer.SQLServer.GetIndexUsageStatsController()

	ieFactory := func(
		ctx context.Context, sessionData *sessiondata.SessionData,
	) sqlutil.InternalExecutor {
		__antithesis_instrumentation__.Notify(195998)
		ie := sql.MakeInternalExecutor(
			ctx,
			pgServer.SQLServer,
			internalMemMetrics,
			cfg.Settings,
		)
		ie.SetSessionData(sessionData)
		return &ie
	}
	__antithesis_instrumentation__.Notify(195911)

	distSQLServer.ServerConfig.SessionBoundInternalExecutorFactory = ieFactory
	jobRegistry.SetSessionBoundInternalExecutorFactory(ieFactory)
	execCfg.IndexBackfiller = sql.NewIndexBackfiller(execCfg, ieFactory)
	execCfg.IndexValidator = scdeps.NewIndexValidator(
		execCfg.DB,
		execCfg.Codec,
		execCfg.Settings,
		ieFactory,
		sql.ValidateForwardIndexes,
		sql.ValidateInvertedIndexes,
		sql.NewFakeSessionData,
	)

	execCfg.DescMetadaUpdaterFactory = descmetadata.NewMetadataUpdaterFactory(
		ieFactory,
		collectionFactory,
		&execCfg.Settings.SV,
	)
	execCfg.InternalExecutorFactory = ieFactory

	distSQLServer.ServerConfig.ProtectedTimestampProvider = execCfg.ProtectedTimestampProvider

	for _, m := range pgServer.Metrics() {
		__antithesis_instrumentation__.Notify(195999)
		cfg.registry.AddMetricStruct(m)
	}
	__antithesis_instrumentation__.Notify(195912)
	*cfg.circularInternalExecutor = sql.MakeInternalExecutor(
		ctx, pgServer.SQLServer, internalMemMetrics, cfg.Settings,
	)
	execCfg.InternalExecutor = cfg.circularInternalExecutor
	stmtDiagnosticsRegistry := stmtdiagnostics.NewRegistry(
		cfg.circularInternalExecutor,
		cfg.db,
		cfg.gossip,
		cfg.Settings,
	)
	execCfg.StmtDiagnosticsRecorder = stmtDiagnosticsRegistry

	{
		__antithesis_instrumentation__.Notify(196000)

		var c migration.Cluster
		var systemDeps migration.SystemDeps
		if codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(196002)
			c = migrationcluster.New(migrationcluster.ClusterConfig{
				NodeLiveness: nodeLiveness,
				Dialer:       cfg.nodeDialer,
				DB:           cfg.db,
			})
			systemDeps = migration.SystemDeps{
				Cluster:    c,
				DB:         cfg.db,
				DistSender: cfg.distSender,
				Stopper:    cfg.stopper,
			}
		} else {
			__antithesis_instrumentation__.Notify(196003)
			c = migrationcluster.NewTenantCluster(cfg.db)
			systemDeps = migration.SystemDeps{
				Cluster: c,
				DB:      cfg.db,
			}
		}
		__antithesis_instrumentation__.Notify(196001)

		knobs, _ := cfg.TestingKnobs.MigrationManager.(*migration.TestingKnobs)
		migrationMgr := migrationmanager.NewManager(
			systemDeps, leaseMgr, cfg.circularInternalExecutor, jobRegistry, codec,
			cfg.Settings, knobs,
		)
		execCfg.MigrationJobDeps = migrationMgr
		execCfg.VersionUpgradeHook = migrationMgr.Migrate
		execCfg.MigrationTestingKnobs = knobs
	}
	__antithesis_instrumentation__.Notify(195913)

	if !codec.ForSystemTenant() || func() bool {
		__antithesis_instrumentation__.Notify(196004)
		return !cfg.SpanConfigsDisabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(196005)

		spanConfigKnobs, _ := cfg.TestingKnobs.SpanConfig.(*spanconfig.TestingKnobs)
		spanConfig.sqlTranslatorFactory = spanconfigsqltranslator.NewFactory(
			execCfg.ProtectedTimestampProvider, codec, spanConfigKnobs,
		)
		spanConfig.sqlWatcher = spanconfigsqlwatcher.New(
			codec,
			cfg.Settings,
			cfg.rangeFeedFactory,
			1<<20,
			cfg.stopper,

			30*time.Second,
			spanConfigKnobs,
		)
		spanConfigReconciler := spanconfigreconciler.New(
			spanConfig.sqlWatcher,
			spanConfig.sqlTranslatorFactory,
			cfg.spanConfigAccessor,
			execCfg,
			codec,
			cfg.TenantID,
			spanConfigKnobs,
		)
		spanConfig.manager = spanconfigmanager.New(
			cfg.db,
			jobRegistry,
			cfg.circularInternalExecutor,
			cfg.stopper,
			cfg.Settings,
			spanConfigReconciler,
			spanConfigKnobs,
		)

		execCfg.SpanConfigReconciler = spanConfigReconciler
	} else {
		__antithesis_instrumentation__.Notify(196006)
	}
	__antithesis_instrumentation__.Notify(195914)
	execCfg.SpanConfigKVAccessor = cfg.spanConfigAccessor
	execCfg.SpanConfigLimiter = spanConfig.limiter
	execCfg.SpanConfigSplitter = spanConfig.splitter

	temporaryObjectCleaner := sql.NewTemporaryObjectCleaner(
		cfg.Settings,
		cfg.db,
		codec,
		cfg.registry,
		distSQLServer.ServerConfig.SessionBoundInternalExecutorFactory,
		cfg.sqlStatusServer,
		cfg.isMeta1Leaseholder,
		sqlExecutorTestingKnobs,
		collectionFactory,
	)

	reporter := &diagnostics.Reporter{
		StartTime:        timeutil.Now(),
		AmbientCtx:       &cfg.AmbientCtx,
		Config:           cfg.BaseConfig.Config,
		Settings:         cfg.Settings,
		StorageClusterID: cfg.rpcContext.StorageClusterID.Get,
		LogicalClusterID: clusterIDForSQL.Get,
		TenantID:         cfg.rpcContext.TenantID,
		SQLInstanceID:    cfg.nodeIDContainer.SQLInstanceID,
		SQLServer:        pgServer.SQLServer,
		InternalExec:     cfg.circularInternalExecutor,
		DB:               cfg.db,
		Recorder:         cfg.recorder,
		Locality:         cfg.Locality,
	}

	if cfg.TestingKnobs.Server != nil {
		__antithesis_instrumentation__.Notify(196007)
		reporter.TestingKnobs = &cfg.TestingKnobs.Server.(*TestingKnobs).DiagnosticsTestingKnobs
	} else {
		__antithesis_instrumentation__.Notify(196008)
	}
	__antithesis_instrumentation__.Notify(195915)

	var settingsWatcher *settingswatcher.SettingsWatcher
	if codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(196009)
		settingsWatcher = settingswatcher.New(
			cfg.clock, codec, cfg.Settings, cfg.rangeFeedFactory, cfg.stopper, cfg.settingsStorage,
		)
	} else {
		__antithesis_instrumentation__.Notify(196010)

		settingsWatcher = settingswatcher.NewWithOverrides(
			cfg.clock, codec, cfg.Settings, cfg.rangeFeedFactory, cfg.stopper, cfg.tenantConnect, cfg.settingsStorage,
		)
	}
	__antithesis_instrumentation__.Notify(195916)

	return &SQLServer{
		ambientCtx:                     cfg.BaseConfig.AmbientCtx,
		stopper:                        cfg.stopper,
		sqlIDContainer:                 cfg.nodeIDContainer,
		pgServer:                       pgServer,
		distSQLServer:                  distSQLServer,
		execCfg:                        execCfg,
		internalExecutor:               cfg.circularInternalExecutor,
		leaseMgr:                       leaseMgr,
		blobService:                    blobService,
		tracingService:                 tracingService,
		tenantConnect:                  cfg.tenantConnect,
		sessionRegistry:                cfg.sessionRegistry,
		jobRegistry:                    jobRegistry,
		statsRefresher:                 statsRefresher,
		temporaryObjectCleaner:         temporaryObjectCleaner,
		internalMemMetrics:             internalMemMetrics,
		sqlMemMetrics:                  sqlMemMetrics,
		stmtDiagnosticsRegistry:        stmtDiagnosticsRegistry,
		sqlLivenessProvider:            cfg.sqlLivenessProvider,
		sqlInstanceProvider:            cfg.sqlInstanceProvider,
		metricsRegistry:                cfg.registry,
		diagnosticsReporter:            reporter,
		spanconfigMgr:                  spanConfig.manager,
		spanconfigSQLTranslatorFactory: spanConfig.sqlTranslatorFactory,
		spanconfigSQLWatcher:           spanConfig.sqlWatcher,
		settingsWatcher:                settingsWatcher,
		systemConfigWatcher:            cfg.systemConfigWatcher,
		isMeta1Leaseholder:             cfg.isMeta1Leaseholder,
		cfg:                            cfg.BaseConfig,
	}, nil
}

func maybeCheckTenantExists(ctx context.Context, codec keys.SQLCodec, db *kv.DB) error {
	__antithesis_instrumentation__.Notify(196011)
	if codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(196015)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(196016)
	}
	__antithesis_instrumentation__.Notify(196012)
	key := catalogkeys.MakeDatabaseNameKey(codec, systemschema.SystemDatabaseName)
	result, err := db.Get(ctx, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(196017)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196018)
	}
	__antithesis_instrumentation__.Notify(196013)
	if result.Value == nil || func() bool {
		__antithesis_instrumentation__.Notify(196019)
		return result.ValueInt() != keys.SystemDatabaseID == true
	}() == true {
		__antithesis_instrumentation__.Notify(196020)
		return errors.New("system DB uninitialized, check if tenant is non existent")
	} else {
		__antithesis_instrumentation__.Notify(196021)
	}
	__antithesis_instrumentation__.Notify(196014)

	return nil
}

func (s *SQLServer) startSQLLivenessAndInstanceProviders(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(196022)

	if s.tenantConnect != nil {
		__antithesis_instrumentation__.Notify(196025)
		if err := s.tenantConnect.Start(ctx); err != nil {
			__antithesis_instrumentation__.Notify(196026)
			return err
		} else {
			__antithesis_instrumentation__.Notify(196027)
		}
	} else {
		__antithesis_instrumentation__.Notify(196028)
	}
	__antithesis_instrumentation__.Notify(196023)
	s.sqlLivenessProvider.Start(ctx)

	if err := s.sqlInstanceProvider.Start(ctx); err != nil {
		__antithesis_instrumentation__.Notify(196029)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196030)
	}
	__antithesis_instrumentation__.Notify(196024)
	return nil
}

func (s *SQLServer) initInstanceID(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(196031)
	if _, ok := s.sqlIDContainer.OptionalNodeID(); ok {
		__antithesis_instrumentation__.Notify(196035)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(196036)
	}
	__antithesis_instrumentation__.Notify(196032)
	instanceID, sessionID, err := s.sqlInstanceProvider.Instance(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(196037)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196038)
	}
	__antithesis_instrumentation__.Notify(196033)
	err = s.sqlIDContainer.SetSQLInstanceID(ctx, instanceID)
	if err != nil {
		__antithesis_instrumentation__.Notify(196039)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196040)
	}
	__antithesis_instrumentation__.Notify(196034)
	s.sqlLivenessSessionID = sessionID
	s.execCfg.DistSQLPlanner.SetSQLInstanceInfo(roachpb.NodeDescriptor{NodeID: roachpb.NodeID(instanceID)})
	return nil
}

func (s *SQLServer) preStart(
	ctx context.Context,
	stopper *stop.Stopper,
	knobs base.TestingKnobs,
	connManager netutil.Server,
	pgL net.Listener,
	socketFile string,
	orphanedLeasesTimeThresholdNanos int64,
) error {
	__antithesis_instrumentation__.Notify(196041)

	if err := s.startSQLLivenessAndInstanceProviders(ctx); err != nil {
		__antithesis_instrumentation__.Notify(196054)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196055)
	}
	__antithesis_instrumentation__.Notify(196042)

	if err := maybeCheckTenantExists(ctx, s.execCfg.Codec, s.execCfg.DB); err != nil {
		__antithesis_instrumentation__.Notify(196056)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196057)
	}
	__antithesis_instrumentation__.Notify(196043)
	if err := s.initInstanceID(ctx); err != nil {
		__antithesis_instrumentation__.Notify(196058)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196059)
	}
	__antithesis_instrumentation__.Notify(196044)
	s.connManager = connManager
	s.pgL = pgL
	s.execCfg.GCJobNotifier.Start(ctx)
	s.temporaryObjectCleaner.Start(ctx, stopper)
	s.distSQLServer.Start()
	s.pgServer.Start(ctx, stopper)
	if err := s.statsRefresher.Start(ctx, stopper, stats.DefaultRefreshInterval); err != nil {
		__antithesis_instrumentation__.Notify(196060)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196061)
	}
	__antithesis_instrumentation__.Notify(196045)
	s.stmtDiagnosticsRegistry.Start(ctx, stopper)

	var mmKnobs startupmigrations.MigrationManagerTestingKnobs
	if migrationManagerTestingKnobs := knobs.StartupMigrationManager; migrationManagerTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(196062)
		mmKnobs = *migrationManagerTestingKnobs.(*startupmigrations.MigrationManagerTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(196063)
	}
	__antithesis_instrumentation__.Notify(196046)

	s.leaseMgr.RefreshLeases(ctx, stopper, s.execCfg.DB)
	s.leaseMgr.PeriodicallyRefreshSomeLeases(ctx)

	migrationsExecutor := sql.MakeInternalExecutor(
		ctx, s.pgServer.SQLServer, s.internalMemMetrics, s.execCfg.Settings)
	migrationsExecutor.SetSessionData(
		&sessiondata.SessionData{
			LocalOnlySessionData: sessiondatapb.LocalOnlySessionData{

				DistSQLMode: sessiondatapb.DistSQLOff,
			},
		})
	startupMigrationsMgr := startupmigrations.NewManager(
		stopper,
		s.execCfg.DB,
		s.execCfg.Codec,
		&migrationsExecutor,
		s.execCfg.Clock,
		mmKnobs,
		s.execCfg.NodeID.SQLInstanceID().String(),
		s.execCfg.Settings,
		s.jobRegistry,
	)
	s.startupMigrationsMgr = startupMigrationsMgr

	if err := s.jobRegistry.Start(ctx, stopper); err != nil {
		__antithesis_instrumentation__.Notify(196064)
		return err
	} else {
		__antithesis_instrumentation__.Notify(196065)
	}
	__antithesis_instrumentation__.Notify(196047)

	if s.spanconfigMgr != nil {
		__antithesis_instrumentation__.Notify(196066)
		if err := s.spanconfigMgr.Start(ctx); err != nil {
			__antithesis_instrumentation__.Notify(196067)
			return err
		} else {
			__antithesis_instrumentation__.Notify(196068)
		}
	} else {
		__antithesis_instrumentation__.Notify(196069)
	}
	__antithesis_instrumentation__.Notify(196048)

	var bootstrapVersion roachpb.Version
	if s.execCfg.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(196070)
		if err := s.execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(196071)
			return txn.GetProto(ctx, keys.BootstrapVersionKey, &bootstrapVersion)
		}); err != nil {
			__antithesis_instrumentation__.Notify(196072)
			return err
		} else {
			__antithesis_instrumentation__.Notify(196073)
		}
	} else {
		__antithesis_instrumentation__.Notify(196074)

		bootstrapVersion = roachpb.Version{Major: 20, Minor: 1, Internal: 1}
	}
	__antithesis_instrumentation__.Notify(196049)

	if err := s.settingsWatcher.Start(ctx); err != nil {
		__antithesis_instrumentation__.Notify(196075)
		return errors.Wrap(err, "initializing settings")
	} else {
		__antithesis_instrumentation__.Notify(196076)
	}
	__antithesis_instrumentation__.Notify(196050)
	if err := s.systemConfigWatcher.Start(ctx, s.stopper); err != nil {
		__antithesis_instrumentation__.Notify(196077)
		return errors.Wrap(err, "initializing settings")
	} else {
		__antithesis_instrumentation__.Notify(196078)
	}
	__antithesis_instrumentation__.Notify(196051)

	if err := startupMigrationsMgr.EnsureMigrations(ctx, bootstrapVersion); err != nil {
		__antithesis_instrumentation__.Notify(196079)
		return errors.Wrap(err, "ensuring SQL migrations")
	} else {
		__antithesis_instrumentation__.Notify(196080)
	}
	__antithesis_instrumentation__.Notify(196052)

	log.Infof(ctx, "done ensuring all necessary startup migrations have run")

	s.leaseMgr.DeleteOrphanedLeases(ctx, orphanedLeasesTimeThresholdNanos)

	jobs.StartJobSchedulerDaemon(
		ctx,
		stopper,
		s.metricsRegistry,
		&scheduledjobs.JobExecutionConfig{
			Settings:         s.execCfg.Settings,
			InternalExecutor: s.internalExecutor,
			DB:               s.execCfg.DB,
			TestingKnobs:     knobs.JobsTestingKnobs,
			PlanHookMaker: func(opName string, txn *kv.Txn, user security.SQLUsername) (interface{}, func()) {
				__antithesis_instrumentation__.Notify(196081)

				return sql.NewInternalPlanner(
					opName,
					txn,
					user,
					&sql.MemoryMetrics{},
					s.execCfg,
					sessiondatapb.SessionData{},
				)
			},
			ShouldRunScheduler: func(ctx context.Context, ts hlc.ClockTimestamp) (bool, error) {
				__antithesis_instrumentation__.Notify(196082)
				if s.execCfg.Codec.ForSystemTenant() {
					__antithesis_instrumentation__.Notify(196084)
					return s.isMeta1Leaseholder(ctx, ts)
				} else {
					__antithesis_instrumentation__.Notify(196085)
				}
				__antithesis_instrumentation__.Notify(196083)
				return true, nil
			},
		},
		scheduledjobs.ProdJobSchedulerEnv,
	)
	__antithesis_instrumentation__.Notify(196053)

	scheduledlogging.Start(ctx, stopper, s.execCfg.DB, s.execCfg.Settings, s.internalExecutor, s.execCfg.CaptureIndexUsageStatsKnobs)
	return nil
}

func (s *SQLServer) SQLInstanceID() base.SQLInstanceID {
	__antithesis_instrumentation__.Notify(196086)
	return s.sqlIDContainer.SQLInstanceID()
}

func (s *SQLServer) StartDiagnostics(ctx context.Context) {
	__antithesis_instrumentation__.Notify(196087)
	s.diagnosticsReporter.PeriodicallyReportDiagnostics(ctx, s.stopper)
}

func (s *SQLServer) AnnotateCtx(ctx context.Context) context.Context {
	__antithesis_instrumentation__.Notify(196088)
	return s.ambientCtx.AnnotateCtx(ctx)
}

func (s *SQLServer) startServeSQL(
	ctx context.Context,
	stopper *stop.Stopper,
	connManager netutil.Server,
	pgL net.Listener,
	socketFile string,
) error {
	__antithesis_instrumentation__.Notify(196089)
	log.Ops.Info(ctx, "serving sql connections")

	pgCtx := s.pgServer.AmbientCtx.AnnotateCtx(context.Background())
	tcpKeepAlive := makeTCPKeepAliveManager()

	_ = stopper.RunAsyncTaskEx(pgCtx,
		stop.TaskOpts{TaskName: "pgwire-listener", SpanOpt: stop.SterileRootSpan},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(196092)
			err := connManager.ServeWith(ctx, stopper, pgL, func(ctx context.Context, conn net.Conn) {
				__antithesis_instrumentation__.Notify(196094)
				connCtx := s.pgServer.AnnotateCtxForIncomingConn(ctx, conn)
				tcpKeepAlive.configure(connCtx, conn)

				if err := s.pgServer.ServeConn(connCtx, conn, pgwire.SocketTCP); err != nil {
					__antithesis_instrumentation__.Notify(196095)
					log.Ops.Errorf(connCtx, "serving SQL client conn: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(196096)
				}
			})
			__antithesis_instrumentation__.Notify(196093)
			netutil.FatalIfUnexpected(err)
		})
	__antithesis_instrumentation__.Notify(196090)

	if len(socketFile) != 0 {
		__antithesis_instrumentation__.Notify(196097)
		log.Ops.Infof(ctx, "starting postgres server at unix:%s", socketFile)

		unixLn, err := net.Listen("unix", socketFile)
		if err != nil {
			__antithesis_instrumentation__.Notify(196101)
			return err
		} else {
			__antithesis_instrumentation__.Notify(196102)
		}
		__antithesis_instrumentation__.Notify(196098)

		waitQuiesce := func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(196103)
			<-stopper.ShouldQuiesce()

			if err := unixLn.Close(); err != nil {
				__antithesis_instrumentation__.Notify(196104)
				log.Ops.Fatalf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(196105)
			}
		}
		__antithesis_instrumentation__.Notify(196099)
		if err := stopper.RunAsyncTaskEx(ctx,
			stop.TaskOpts{TaskName: "unix-ln-close", SpanOpt: stop.SterileRootSpan},
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(196106)
				waitQuiesce(ctx)
			}); err != nil {
			__antithesis_instrumentation__.Notify(196107)
			waitQuiesce(ctx)
			return err
		} else {
			__antithesis_instrumentation__.Notify(196108)
		}
		__antithesis_instrumentation__.Notify(196100)

		if err := stopper.RunAsyncTaskEx(pgCtx,
			stop.TaskOpts{TaskName: "unix-listener", SpanOpt: stop.SterileRootSpan},
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(196109)
				err := connManager.ServeWith(ctx, stopper, unixLn, func(ctx context.Context, conn net.Conn) {
					__antithesis_instrumentation__.Notify(196111)
					connCtx := s.pgServer.AnnotateCtxForIncomingConn(ctx, conn)
					if err := s.pgServer.ServeConn(connCtx, conn, pgwire.SocketUnix); err != nil {
						__antithesis_instrumentation__.Notify(196112)
						log.Ops.Errorf(connCtx, "%v", err)
					} else {
						__antithesis_instrumentation__.Notify(196113)
					}
				})
				__antithesis_instrumentation__.Notify(196110)
				netutil.FatalIfUnexpected(err)
			}); err != nil {
			__antithesis_instrumentation__.Notify(196114)
			return err
		} else {
			__antithesis_instrumentation__.Notify(196115)
		}
	} else {
		__antithesis_instrumentation__.Notify(196116)
	}
	__antithesis_instrumentation__.Notify(196091)

	s.isReady.Set(true)

	return nil
}

func (s *SQLServer) LogicalClusterID() uuid.UUID {
	__antithesis_instrumentation__.Notify(196117)
	return s.execCfg.LogicalClusterID()
}
