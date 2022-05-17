package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobsprotectedts"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvprober"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/sidetransport"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptprovider"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptreconcile"
	serverrangefeed "github.com/cockroachdb/cockroach/pkg/kv/kvserver/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/reports"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/server/debug"
	"github.com/cockroachdb/cockroach/pkg/server/diagnostics"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/server/systemconfigwatcher"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/server/tenantsettingswatcher"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	_ "github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigjob"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigkvaccessor"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigkvsubscriber"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigptsreader"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	_ "github.com/cockroachdb/cockroach/pkg/sql/gcjob"
	_ "github.com/cockroachdb/cockroach/pkg/sql/importer"
	"github.com/cockroachdb/cockroach/pkg/sql/optionalnodeliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire"
	_ "github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scjob"
	_ "github.com/cockroachdb/cockroach/pkg/sql/ttl/ttljob"
	_ "github.com/cockroachdb/cockroach/pkg/sql/ttl/ttlschedule"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/ts"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/goschedstats"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	sentry "github.com/getsentry/sentry-go"
	"google.golang.org/grpc/codes"
)

type Server struct {
	nodeIDContainer *base.NodeIDContainer
	cfg             Config
	st              *cluster.Settings
	clock           *hlc.Clock
	rpcContext      *rpc.Context
	engines         Engines

	grpc             *grpcServer
	gossip           *gossip.Gossip
	nodeDialer       *nodedialer.Dialer
	nodeLiveness     *liveness.NodeLiveness
	storePool        *kvserver.StorePool
	tcsFactory       *kvcoord.TxnCoordSenderFactory
	distSender       *kvcoord.DistSender
	db               *kv.DB
	node             *Node
	registry         *metric.Registry
	recorder         *status.MetricsRecorder
	runtime          *status.RuntimeStatSampler
	ruleRegistry     *metric.RuleRegistry
	promRuleExporter *metric.PrometheusRuleExporter
	updates          *diagnostics.UpdateChecker
	ctSender         *sidetransport.Sender

	http            *httpServer
	adminAuthzCheck *adminPrivilegeChecker
	admin           *adminServer
	status          *statusServer
	drain           *drainServer
	authentication  *authenticationServer
	migrationServer *migrationServer
	tsDB            *ts.DB
	tsServer        *ts.Server
	raftTransport   *kvserver.RaftTransport
	stopper         *stop.Stopper

	debug    *debug.Server
	kvProber *kvprober.Prober

	replicationReporter *reports.Reporter
	protectedtsProvider protectedts.Provider

	spanConfigSubscriber spanconfig.KVSubscriber

	sqlServer *SQLServer

	externalStorageBuilder *externalStorageBuilder

	storeGrantCoords *admission.StoreGrantCoordinators

	kvMemoryMonitor *mon.BytesMonitor

	startTime time.Time
}

func NewServer(cfg Config, stopper *stop.Stopper) (*Server, error) {
	__antithesis_instrumentation__.Notify(195465)
	if err := cfg.ValidateAddrs(context.Background()); err != nil {
		__antithesis_instrumentation__.Notify(195501)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195502)
	}
	__antithesis_instrumentation__.Notify(195466)

	st := cfg.Settings

	if cfg.AmbientCtx.Tracer == nil {
		__antithesis_instrumentation__.Notify(195503)
		panic(errors.New("no tracer set in AmbientCtx"))
	} else {
		__antithesis_instrumentation__.Notify(195504)
	}
	__antithesis_instrumentation__.Notify(195467)

	var clock *hlc.Clock
	if cfg.ClockDevicePath != "" {
		__antithesis_instrumentation__.Notify(195505)
		clockSrc, err := hlc.MakeClockSource(context.Background(), cfg.ClockDevicePath)
		if err != nil {
			__antithesis_instrumentation__.Notify(195507)
			return nil, errors.Wrap(err, "instantiating clock source")
		} else {
			__antithesis_instrumentation__.Notify(195508)
		}
		__antithesis_instrumentation__.Notify(195506)
		clock = hlc.NewClock(clockSrc.UnixNano, time.Duration(cfg.MaxOffset))
	} else {
		__antithesis_instrumentation__.Notify(195509)
		if cfg.TestingKnobs.Server != nil && func() bool {
			__antithesis_instrumentation__.Notify(195510)
			return cfg.TestingKnobs.Server.(*TestingKnobs).ClockSource != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(195511)
			clock = hlc.NewClock(cfg.TestingKnobs.Server.(*TestingKnobs).ClockSource,
				time.Duration(cfg.MaxOffset))
		} else {
			__antithesis_instrumentation__.Notify(195512)
			clock = hlc.NewClock(hlc.UnixNano, time.Duration(cfg.MaxOffset))
		}
	}
	__antithesis_instrumentation__.Notify(195468)
	registry := metric.NewRegistry()
	ruleRegistry := metric.NewRuleRegistry()
	promRuleExporter := metric.NewPrometheusRuleExporter(ruleRegistry)
	stopper.SetTracer(cfg.AmbientCtx.Tracer)
	stopper.AddCloser(cfg.AmbientCtx.Tracer)

	nodeIDContainer := cfg.IDContainer
	idContainer := base.NewSQLIDContainerForNode(nodeIDContainer)

	ctx := cfg.AmbientCtx.AnnotateCtx(context.Background())

	engines, err := cfg.CreateEngines(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(195513)
		return nil, errors.Wrap(err, "failed to create engines")
	} else {
		__antithesis_instrumentation__.Notify(195514)
	}
	__antithesis_instrumentation__.Notify(195469)
	stopper.AddCloser(&engines)

	nodeTombStorage, checkPingFor := getPingCheckDecommissionFn(engines)

	rpcCtxOpts := rpc.ContextOptions{
		TenantID:         roachpb.SystemTenantID,
		NodeID:           cfg.IDContainer,
		StorageClusterID: cfg.ClusterIDContainer,
		Config:           cfg.Config,
		Clock:            clock,
		Stopper:          stopper,
		Settings:         cfg.Settings,
		OnOutgoingPing: func(ctx context.Context, req *rpc.PingRequest) error {
			__antithesis_instrumentation__.Notify(195515)

			return checkPingFor(ctx, req.TargetNodeID, codes.FailedPrecondition)
		},
		OnIncomingPing: func(ctx context.Context, req *rpc.PingRequest) error {
			__antithesis_instrumentation__.Notify(195516)

			if tenantID, isTenant := roachpb.TenantFromContext(ctx); isTenant && func() bool {
				__antithesis_instrumentation__.Notify(195518)
				return !roachpb.IsSystemTenantID(tenantID.ToUint64()) == true
			}() == true {
				__antithesis_instrumentation__.Notify(195519)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(195520)
			}
			__antithesis_instrumentation__.Notify(195517)

			return checkPingFor(ctx, req.OriginNodeID, codes.PermissionDenied)
		},
	}
	__antithesis_instrumentation__.Notify(195470)
	if knobs := cfg.TestingKnobs.Server; knobs != nil {
		__antithesis_instrumentation__.Notify(195521)
		serverKnobs := knobs.(*TestingKnobs)
		rpcCtxOpts.Knobs = serverKnobs.ContextTestingKnobs
	} else {
		__antithesis_instrumentation__.Notify(195522)
	}
	__antithesis_instrumentation__.Notify(195471)
	rpcContext := rpc.NewContext(ctx, rpcCtxOpts)

	rpcContext.HeartbeatCB = func() {
		__antithesis_instrumentation__.Notify(195523)
		if err := rpcContext.RemoteClocks.VerifyClockOffset(ctx); err != nil {
			__antithesis_instrumentation__.Notify(195524)
			log.Ops.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(195525)
		}
	}
	__antithesis_instrumentation__.Notify(195472)
	registry.AddMetricStruct(rpcContext.Metrics())

	if !cfg.Insecure {
		__antithesis_instrumentation__.Notify(195526)

		if _, err := rpcContext.GetServerTLSConfig(); err != nil {
			__antithesis_instrumentation__.Notify(195531)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(195532)
		}
		__antithesis_instrumentation__.Notify(195527)
		if _, err := rpcContext.GetUIServerTLSConfig(); err != nil {
			__antithesis_instrumentation__.Notify(195533)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(195534)
		}
		__antithesis_instrumentation__.Notify(195528)
		if _, err := rpcContext.GetClientTLSConfig(); err != nil {
			__antithesis_instrumentation__.Notify(195535)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(195536)
		}
		__antithesis_instrumentation__.Notify(195529)
		cm, err := rpcContext.GetCertificateManager()
		if err != nil {
			__antithesis_instrumentation__.Notify(195537)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(195538)
		}
		__antithesis_instrumentation__.Notify(195530)
		cm.RegisterSignalHandler(stopper)
		registry.AddMetricStruct(cm.Metrics())
	} else {
		__antithesis_instrumentation__.Notify(195539)
	}
	__antithesis_instrumentation__.Notify(195473)

	rpcContext.CheckCertificateAddrs(ctx)

	grpcServer := newGRPCServer(rpcContext)

	g := gossip.New(
		cfg.AmbientCtx,
		rpcContext.StorageClusterID,
		nodeIDContainer,
		rpcContext,
		grpcServer.Server,
		stopper,
		registry,
		cfg.Locality,
		&cfg.DefaultZoneConfig,
	)

	var dialerKnobs nodedialer.DialerTestingKnobs
	if dk := cfg.TestingKnobs.DialerKnobs; dk != nil {
		__antithesis_instrumentation__.Notify(195540)
		dialerKnobs = dk.(nodedialer.DialerTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195541)
	}
	__antithesis_instrumentation__.Notify(195474)

	nodeDialer := nodedialer.NewWithOpt(rpcContext, gossip.AddressResolver(g),
		nodedialer.DialerOpt{TestingKnobs: dialerKnobs})

	runtimeSampler := status.NewRuntimeStatSampler(ctx, clock)
	registry.AddMetricStruct(runtimeSampler)

	registry.AddMetric(base.LicenseTTL)
	err = base.UpdateMetricOnLicenseChange(ctx, cfg.Settings, base.LicenseTTL, timeutil.DefaultTimeSource{}, stopper)
	if err != nil {
		__antithesis_instrumentation__.Notify(195542)
		log.Errorf(ctx, "unable to initialize periodic license metric update: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(195543)
	}
	__antithesis_instrumentation__.Notify(195475)

	kvserver.CreateAndAddRules(ctx, ruleRegistry)

	var clientTestingKnobs kvcoord.ClientTestingKnobs
	if kvKnobs := cfg.TestingKnobs.KVClient; kvKnobs != nil {
		__antithesis_instrumentation__.Notify(195544)
		clientTestingKnobs = *kvKnobs.(*kvcoord.ClientTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195545)
	}
	__antithesis_instrumentation__.Notify(195476)
	retryOpts := cfg.RetryOptions
	if retryOpts == (retry.Options{}) {
		__antithesis_instrumentation__.Notify(195546)
		retryOpts = base.DefaultRetryOptions()
	} else {
		__antithesis_instrumentation__.Notify(195547)
	}
	__antithesis_instrumentation__.Notify(195477)
	retryOpts.Closer = stopper.ShouldQuiesce()
	distSenderCfg := kvcoord.DistSenderConfig{
		AmbientCtx:         cfg.AmbientCtx,
		Settings:           st,
		Clock:              clock,
		NodeDescs:          g,
		RPCContext:         rpcContext,
		RPCRetryOptions:    &retryOpts,
		NodeDialer:         nodeDialer,
		FirstRangeProvider: g,
		TestingKnobs:       clientTestingKnobs,
	}
	distSender := kvcoord.NewDistSender(distSenderCfg)
	registry.AddMetricStruct(distSender.Metrics())

	txnMetrics := kvcoord.MakeTxnMetrics(cfg.HistogramWindowInterval())
	registry.AddMetricStruct(txnMetrics)
	txnCoordSenderFactoryCfg := kvcoord.TxnCoordSenderFactoryConfig{
		AmbientCtx:   cfg.AmbientCtx,
		Settings:     st,
		Clock:        clock,
		Stopper:      stopper,
		Linearizable: cfg.Linearizable,
		Metrics:      txnMetrics,
		TestingKnobs: clientTestingKnobs,
	}
	tcsFactory := kvcoord.NewTxnCoordSenderFactory(txnCoordSenderFactoryCfg, distSender)

	admissionOptions := admission.DefaultOptions
	if opts, ok := cfg.TestingKnobs.AdmissionControl.(*admission.Options); ok {
		__antithesis_instrumentation__.Notify(195548)
		admissionOptions.Override(opts)
	} else {
		__antithesis_instrumentation__.Notify(195549)
	}
	__antithesis_instrumentation__.Notify(195478)
	admissionOptions.Settings = st
	gcoords, metrics := admission.NewGrantCoordinators(cfg.AmbientCtx, admissionOptions)
	for i := range metrics {
		__antithesis_instrumentation__.Notify(195550)
		registry.AddMetricStruct(metrics[i])
	}
	__antithesis_instrumentation__.Notify(195479)
	cbID := goschedstats.RegisterRunnableCountCallback(gcoords.Regular.CPULoad)
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(195551)
		goschedstats.UnregisterRunnableCountCallback(cbID)
	}))
	__antithesis_instrumentation__.Notify(195480)
	stopper.AddCloser(gcoords)

	dbCtx := kv.DefaultDBContext(stopper)
	dbCtx.NodeID = idContainer
	dbCtx.Stopper = stopper
	db := kv.NewDBWithContext(cfg.AmbientCtx, tcsFactory, clock, dbCtx)
	db.SQLKVResponseAdmissionQ = gcoords.Regular.GetWorkQueue(admission.SQLKVResponseWork)

	nlActive, nlRenewal := cfg.NodeLivenessDurations()
	if knobs := cfg.TestingKnobs.NodeLiveness; knobs != nil {
		__antithesis_instrumentation__.Notify(195552)
		nlKnobs := knobs.(kvserver.NodeLivenessTestingKnobs)
		if duration := nlKnobs.LivenessDuration; duration != 0 {
			__antithesis_instrumentation__.Notify(195554)
			nlActive = duration
		} else {
			__antithesis_instrumentation__.Notify(195555)
		}
		__antithesis_instrumentation__.Notify(195553)
		if duration := nlKnobs.RenewalDuration; duration != 0 {
			__antithesis_instrumentation__.Notify(195556)
			nlRenewal = duration
		} else {
			__antithesis_instrumentation__.Notify(195557)
		}
	} else {
		__antithesis_instrumentation__.Notify(195558)
	}
	__antithesis_instrumentation__.Notify(195481)

	rangeFeedKnobs, _ := cfg.TestingKnobs.RangeFeed.(*rangefeed.TestingKnobs)
	rangeFeedFactory, err := rangefeed.NewFactory(stopper, db, st, rangeFeedKnobs)
	if err != nil {
		__antithesis_instrumentation__.Notify(195559)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195560)
	}
	__antithesis_instrumentation__.Notify(195482)

	nodeLiveness := liveness.NewNodeLiveness(liveness.NodeLivenessOptions{
		AmbientCtx:              cfg.AmbientCtx,
		Clock:                   clock,
		DB:                      db,
		Gossip:                  g,
		LivenessThreshold:       nlActive,
		RenewalDuration:         nlRenewal,
		Settings:                st,
		HistogramWindowInterval: cfg.HistogramWindowInterval(),
		OnNodeDecommissioned: func(liveness livenesspb.Liveness) {
			__antithesis_instrumentation__.Notify(195561)
			if knobs, ok := cfg.TestingKnobs.Server.(*TestingKnobs); ok && func() bool {
				__antithesis_instrumentation__.Notify(195563)
				return knobs.OnDecommissionedCallback != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(195564)
				knobs.OnDecommissionedCallback(liveness)
			} else {
				__antithesis_instrumentation__.Notify(195565)
			}
			__antithesis_instrumentation__.Notify(195562)
			if err := nodeTombStorage.SetDecommissioned(
				ctx, liveness.NodeID, timeutil.Unix(0, liveness.Expiration.WallTime).UTC(),
			); err != nil {
				__antithesis_instrumentation__.Notify(195566)
				log.Fatalf(ctx, "unable to add tombstone for n%d: %s", liveness.NodeID, err)
			} else {
				__antithesis_instrumentation__.Notify(195567)
			}
		},
	})
	__antithesis_instrumentation__.Notify(195483)
	registry.AddMetricStruct(nodeLiveness.Metrics())

	nodeLivenessFn := kvserver.MakeStorePoolNodeLivenessFunc(nodeLiveness)
	if nodeLivenessKnobs, ok := cfg.TestingKnobs.NodeLiveness.(kvserver.NodeLivenessTestingKnobs); ok && func() bool {
		__antithesis_instrumentation__.Notify(195568)
		return nodeLivenessKnobs.StorePoolNodeLivenessFn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(195569)
		nodeLivenessFn = nodeLivenessKnobs.StorePoolNodeLivenessFn
	} else {
		__antithesis_instrumentation__.Notify(195570)
	}
	__antithesis_instrumentation__.Notify(195484)
	storePool := kvserver.NewStorePool(
		cfg.AmbientCtx,
		st,
		g,
		clock,
		nodeLiveness.GetNodeCount,
		nodeLivenessFn,
		false,
	)

	raftTransport := kvserver.NewRaftTransport(
		cfg.AmbientCtx, st, nodeDialer, grpcServer.Server, stopper,
	)

	ctSender := sidetransport.NewSender(stopper, st, clock, nodeDialer)
	stores := kvserver.NewStores(cfg.AmbientCtx, clock)
	ctReceiver := sidetransport.NewReceiver(nodeIDContainer, stopper, stores, nil)

	internalExecutor := &sql.InternalExecutor{}
	jobRegistry := &jobs.Registry{}

	externalStorageBuilder := &externalStorageBuilder{}
	externalStorage := externalStorageBuilder.makeExternalStorage
	externalStorageFromURI := externalStorageBuilder.makeExternalStorageFromURI

	protectedtsKnobs, _ := cfg.TestingKnobs.ProtectedTS.(*protectedts.TestingKnobs)
	protectedtsProvider, err := ptprovider.New(ptprovider.Config{
		DB:               db,
		InternalExecutor: internalExecutor,
		Settings:         st,
		Knobs:            protectedtsKnobs,
		ReconcileStatusFuncs: ptreconcile.StatusFuncs{
			jobsprotectedts.GetMetaType(jobsprotectedts.Jobs): jobsprotectedts.MakeStatusFunc(
				jobRegistry, internalExecutor, jobsprotectedts.Jobs),
			jobsprotectedts.GetMetaType(jobsprotectedts.Schedules): jobsprotectedts.MakeStatusFunc(jobRegistry,
				internalExecutor, jobsprotectedts.Schedules),
		},
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(195571)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195572)
	}
	__antithesis_instrumentation__.Notify(195485)
	registry.AddMetricStruct(protectedtsProvider.Metrics())

	sqlMonitorAndMetrics := newRootSQLMemoryMonitor(monitorAndMetricsOptions{
		memoryPoolSize:          cfg.MemoryPoolSize,
		histogramWindowInterval: cfg.HistogramWindowInterval(),
		settings:                cfg.Settings,
	})
	kvMemoryMonitor := mon.NewMonitorInheritWithLimit(
		"kv-mem", 0, sqlMonitorAndMetrics.rootSQLMemoryMonitor)
	kvMemoryMonitor.Start(ctx, sqlMonitorAndMetrics.rootSQLMemoryMonitor, mon.BoundAccount{})
	rangeReedBudgetFactory := serverrangefeed.NewBudgetFactory(
		ctx,
		serverrangefeed.CreateBudgetFactoryConfig(
			kvMemoryMonitor,
			cfg.MemoryPoolSize,
			cfg.HistogramWindowInterval(),
			func(limit int64) int64 {
				__antithesis_instrumentation__.Notify(195573)
				if !serverrangefeed.RangefeedBudgetsEnabled.Get(&st.SV) {
					__antithesis_instrumentation__.Notify(195576)
					return 0
				} else {
					__antithesis_instrumentation__.Notify(195577)
				}
				__antithesis_instrumentation__.Notify(195574)
				if raftCmdLimit := kvserver.MaxCommandSize.Get(&st.SV); raftCmdLimit > limit {
					__antithesis_instrumentation__.Notify(195578)
					return raftCmdLimit
				} else {
					__antithesis_instrumentation__.Notify(195579)
				}
				__antithesis_instrumentation__.Notify(195575)
				return limit
			},
			&st.SV))
	__antithesis_instrumentation__.Notify(195486)
	if rangeReedBudgetFactory != nil {
		__antithesis_instrumentation__.Notify(195580)
		registry.AddMetricStruct(rangeReedBudgetFactory.Metrics())
	} else {
		__antithesis_instrumentation__.Notify(195581)
	}
	__antithesis_instrumentation__.Notify(195487)

	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(195582)
		rangeReedBudgetFactory.Stop(ctx)
	}))
	__antithesis_instrumentation__.Notify(195488)
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(195583)
		kvMemoryMonitor.Stop(ctx)
	}))
	__antithesis_instrumentation__.Notify(195489)

	tsDB := ts.NewDB(db, cfg.Settings)
	registry.AddMetricStruct(tsDB.Metrics())
	nodeCountFn := func() int64 {
		__antithesis_instrumentation__.Notify(195584)
		return nodeLiveness.Metrics().LiveNodes.Value()
	}
	__antithesis_instrumentation__.Notify(195490)
	sTS := ts.MakeServer(
		cfg.AmbientCtx, tsDB, nodeCountFn, cfg.TimeSeriesServerConfig,
		sqlMonitorAndMetrics.rootSQLMemoryMonitor, stopper,
	)

	systemConfigWatcher := systemconfigwatcher.New(
		keys.SystemSQLCodec, clock, rangeFeedFactory, &cfg.DefaultZoneConfig,
	)

	var spanConfig struct {
		kvAccessor spanconfig.KVAccessor

		subscriber spanconfig.KVSubscriber

		kvAccessorForTenantRecords spanconfig.KVAccessor
	}
	if !cfg.SpanConfigsDisabled {
		__antithesis_instrumentation__.Notify(195585)
		spanConfigKnobs, _ := cfg.TestingKnobs.SpanConfig.(*spanconfig.TestingKnobs)
		if spanConfigKnobs != nil && func() bool {
			__antithesis_instrumentation__.Notify(195587)
			return spanConfigKnobs.StoreKVSubscriberOverride != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(195588)
			spanConfig.subscriber = spanConfigKnobs.StoreKVSubscriberOverride
		} else {
			__antithesis_instrumentation__.Notify(195589)

			fallbackConf := cfg.DefaultZoneConfig.AsSpanConfig()
			fallbackConf.RangefeedEnabled = true

			fallbackConf.GCPolicy.IgnoreStrictEnforcement = true

			spanConfig.subscriber = spanconfigkvsubscriber.New(
				clock,
				rangeFeedFactory,
				keys.SpanConfigurationsTableID,
				1<<20,
				fallbackConf,
				spanConfigKnobs,
			)
		}
		__antithesis_instrumentation__.Notify(195586)

		scKVAccessor := spanconfigkvaccessor.New(
			db, internalExecutor, cfg.Settings, clock,
			systemschema.SpanConfigurationsTableName.FQString(),
			spanConfigKnobs,
		)
		spanConfig.kvAccessor, spanConfig.kvAccessorForTenantRecords = scKVAccessor, scKVAccessor
	} else {
		__antithesis_instrumentation__.Notify(195590)

		spanConfig.kvAccessor = spanconfigkvaccessor.DisabledKVAccessor

		spanConfig.kvAccessorForTenantRecords = spanconfigkvaccessor.NoopKVAccessor

		spanConfig.subscriber = spanconfigkvsubscriber.NewNoopSubscriber(clock)
	}
	__antithesis_instrumentation__.Notify(195491)

	var protectedTSReader spanconfig.ProtectedTSReader
	if cfg.TestingKnobs.SpanConfig != nil && func() bool {
		__antithesis_instrumentation__.Notify(195591)
		return cfg.TestingKnobs.SpanConfig.(*spanconfig.TestingKnobs).ProtectedTSReaderOverrideFn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(195592)
		fn := cfg.TestingKnobs.SpanConfig.(*spanconfig.TestingKnobs).ProtectedTSReaderOverrideFn
		protectedTSReader = fn(clock)
	} else {
		__antithesis_instrumentation__.Notify(195593)
		protectedTSReader = spanconfigptsreader.NewAdapter(protectedtsProvider.(*ptprovider.Provider).Cache, spanConfig.subscriber)
	}
	__antithesis_instrumentation__.Notify(195492)

	storeCfg := kvserver.StoreConfig{
		DefaultSpanConfig:        cfg.DefaultZoneConfig.AsSpanConfig(),
		Settings:                 st,
		AmbientCtx:               cfg.AmbientCtx,
		RaftConfig:               cfg.RaftConfig,
		Clock:                    clock,
		DB:                       db,
		Gossip:                   g,
		NodeLiveness:             nodeLiveness,
		Transport:                raftTransport,
		NodeDialer:               nodeDialer,
		RPCContext:               rpcContext,
		ScanInterval:             cfg.ScanInterval,
		ScanMinIdleTime:          cfg.ScanMinIdleTime,
		ScanMaxIdleTime:          cfg.ScanMaxIdleTime,
		HistogramWindowInterval:  cfg.HistogramWindowInterval(),
		StorePool:                storePool,
		SQLExecutor:              internalExecutor,
		LogRangeEvents:           cfg.EventLogEnabled,
		RangeDescriptorCache:     distSender.RangeDescriptorCache(),
		TimeSeriesDataStore:      tsDB,
		ClosedTimestampSender:    ctSender,
		ClosedTimestampReceiver:  ctReceiver,
		ExternalStorage:          externalStorage,
		ExternalStorageFromURI:   externalStorageFromURI,
		ProtectedTimestampReader: protectedTSReader,
		KVMemoryMonitor:          kvMemoryMonitor,
		RangefeedBudgetFactory:   rangeReedBudgetFactory,
		SystemConfigProvider:     systemConfigWatcher,
		SpanConfigSubscriber:     spanConfig.subscriber,
		SpanConfigsDisabled:      cfg.SpanConfigsDisabled,
	}

	if storeTestingKnobs := cfg.TestingKnobs.Store; storeTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(195594)
		storeCfg.TestingKnobs = *storeTestingKnobs.(*kvserver.StoreTestingKnobs)
	} else {
		__antithesis_instrumentation__.Notify(195595)
	}
	__antithesis_instrumentation__.Notify(195493)

	recorder := status.NewMetricsRecorder(clock, nodeLiveness, rpcContext, g, st)
	registry.AddMetricStruct(rpcContext.RemoteClocks.Metrics())

	updates := &diagnostics.UpdateChecker{
		StartTime:        timeutil.Now(),
		AmbientCtx:       &cfg.AmbientCtx,
		Config:           cfg.BaseConfig.Config,
		Settings:         cfg.Settings,
		StorageClusterID: rpcContext.StorageClusterID.Get,
		LogicalClusterID: rpcContext.LogicalClusterID.Get,
		NodeID:           nodeIDContainer.Get,
		SQLInstanceID:    idContainer.SQLInstanceID,
	}

	if cfg.TestingKnobs.Server != nil {
		__antithesis_instrumentation__.Notify(195596)
		updates.TestingKnobs = &cfg.TestingKnobs.Server.(*TestingKnobs).DiagnosticsTestingKnobs
	} else {
		__antithesis_instrumentation__.Notify(195597)
	}
	__antithesis_instrumentation__.Notify(195494)

	tenantUsage := NewTenantUsageServer(st, db, internalExecutor)
	registry.AddMetricStruct(tenantUsage.Metrics())

	tenantSettingsWatcher := tenantsettingswatcher.New(
		clock, rangeFeedFactory, stopper, st,
	)

	node := NewNode(
		storeCfg, recorder, registry, stopper,
		txnMetrics, stores, nil, cfg.ClusterIDContainer,
		gcoords.Regular.GetWorkQueue(admission.KVWork), gcoords.Stores,
		tenantUsage, tenantSettingsWatcher, spanConfig.kvAccessor,
	)
	roachpb.RegisterInternalServer(grpcServer.Server, node)
	kvserver.RegisterPerReplicaServer(grpcServer.Server, node.perReplicaServer)
	kvserver.RegisterPerStoreServer(grpcServer.Server, node.perReplicaServer)
	ctpb.RegisterSideTransportServer(grpcServer.Server, ctReceiver)

	replicationReporter := reports.NewReporter(
		db, node.stores, storePool, st, nodeLiveness, internalExecutor, systemConfigWatcher,
	)

	lateBoundServer := &Server{}

	adminAuthzCheck := &adminPrivilegeChecker{ie: internalExecutor}
	sAdmin := newAdminServer(lateBoundServer, adminAuthzCheck, internalExecutor)

	parseNodeIDFn := func(s string) (roachpb.NodeID, bool, error) {
		__antithesis_instrumentation__.Notify(195598)
		return parseNodeID(g, s)
	}
	__antithesis_instrumentation__.Notify(195495)
	getNodeIDHTTPAddressFn := func(id roachpb.NodeID) (*util.UnresolvedAddr, error) {
		__antithesis_instrumentation__.Notify(195599)
		return g.GetNodeIDHTTPAddress(id)
	}
	__antithesis_instrumentation__.Notify(195496)
	sHTTP := newHTTPServer(cfg.BaseConfig, rpcContext, parseNodeIDFn, getNodeIDHTTPAddressFn)

	sessionRegistry := sql.NewSessionRegistry()
	flowScheduler := flowinfra.NewFlowScheduler(cfg.AmbientCtx, stopper, st)

	sStatus := newStatusServer(
		cfg.AmbientCtx,
		st,
		cfg.Config,
		adminAuthzCheck,
		sAdmin,
		db,
		g,
		recorder,
		nodeLiveness,
		storePool,
		rpcContext,
		node.stores,
		stopper,
		sessionRegistry,
		flowScheduler,
		internalExecutor,
	)

	var jobAdoptionStopFile string
	for _, spec := range cfg.Stores.Specs {
		__antithesis_instrumentation__.Notify(195600)
		if !spec.InMemory && func() bool {
			__antithesis_instrumentation__.Notify(195601)
			return spec.Path != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(195602)
			jobAdoptionStopFile = filepath.Join(spec.Path, jobs.PreventAdoptionFile)
			break
		} else {
			__antithesis_instrumentation__.Notify(195603)
		}
	}
	__antithesis_instrumentation__.Notify(195497)

	kvProber := kvprober.NewProber(kvprober.Opts{
		Tracer:                  cfg.AmbientCtx.Tracer,
		DB:                      db,
		Settings:                st,
		HistogramWindowInterval: cfg.HistogramWindowInterval(),
	})
	registry.AddMetricStruct(kvProber.Metrics())

	settingsWriter := newSettingsCacheWriter(engines[0], stopper)
	sqlServer, err := newSQLServer(ctx, sqlServerArgs{
		sqlServerOptionalKVArgs: sqlServerOptionalKVArgs{
			nodesStatusServer:        serverpb.MakeOptionalNodesStatusServer(sStatus),
			nodeLiveness:             optionalnodeliveness.MakeContainer(nodeLiveness),
			gossip:                   gossip.MakeOptionalGossip(g),
			grpcServer:               grpcServer.Server,
			nodeIDContainer:          idContainer,
			externalStorage:          externalStorage,
			externalStorageFromURI:   externalStorageFromURI,
			isMeta1Leaseholder:       node.stores.IsMeta1Leaseholder,
			sqlSQLResponseAdmissionQ: gcoords.Regular.GetWorkQueue(admission.SQLSQLResponseWork),
			spanConfigKVAccessor:     spanConfig.kvAccessorForTenantRecords,
			kvStoresIterator:         kvserver.MakeStoresIterator(node.stores),
		},
		SQLConfig:                &cfg.SQLConfig,
		BaseConfig:               &cfg.BaseConfig,
		stopper:                  stopper,
		clock:                    clock,
		runtime:                  runtimeSampler,
		rpcContext:               rpcContext,
		nodeDescs:                g,
		systemConfigWatcher:      systemConfigWatcher,
		spanConfigAccessor:       spanConfig.kvAccessor,
		nodeDialer:               nodeDialer,
		distSender:               distSender,
		db:                       db,
		registry:                 registry,
		recorder:                 recorder,
		sessionRegistry:          sessionRegistry,
		flowScheduler:            flowScheduler,
		circularInternalExecutor: internalExecutor,
		circularJobRegistry:      jobRegistry,
		jobAdoptionStopFile:      jobAdoptionStopFile,
		protectedtsProvider:      protectedtsProvider,
		rangeFeedFactory:         rangeFeedFactory,
		sqlStatusServer:          sStatus,
		regionsServer:            sStatus,
		tenantStatusServer:       sStatus,
		tenantUsageServer:        tenantUsage,
		monitorAndMetrics:        sqlMonitorAndMetrics,
		settingsStorage:          settingsWriter,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(195604)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195605)
	}
	__antithesis_instrumentation__.Notify(195498)

	sAuth := newAuthenticationServer(cfg.Config, sqlServer)
	for i, gw := range []grpcGatewayServer{sAdmin, sStatus, sAuth, &sTS} {
		__antithesis_instrumentation__.Notify(195606)
		if reflect.ValueOf(gw).IsNil() {
			__antithesis_instrumentation__.Notify(195608)
			return nil, errors.Errorf("%d: nil", i)
		} else {
			__antithesis_instrumentation__.Notify(195609)
		}
		__antithesis_instrumentation__.Notify(195607)
		gw.RegisterService(grpcServer.Server)
	}
	__antithesis_instrumentation__.Notify(195499)

	sStatus.setStmtDiagnosticsRequester(sqlServer.execCfg.StmtDiagnosticsRecorder)
	sStatus.baseStatusServer.sqlServer = sqlServer
	debugServer := debug.NewServer(cfg.BaseConfig.AmbientCtx, st, sqlServer.pgServer.HBADebugFn(), sStatus)
	node.InitLogger(sqlServer.execCfg)

	drain := newDrainServer(cfg.BaseConfig, stopper, grpcServer, sqlServer)
	drain.setNode(node, nodeLiveness)

	*lateBoundServer = Server{
		nodeIDContainer:        nodeIDContainer,
		cfg:                    cfg,
		st:                     st,
		clock:                  clock,
		rpcContext:             rpcContext,
		engines:                engines,
		grpc:                   grpcServer,
		gossip:                 g,
		nodeDialer:             nodeDialer,
		nodeLiveness:           nodeLiveness,
		storePool:              storePool,
		tcsFactory:             tcsFactory,
		distSender:             distSender,
		db:                     db,
		node:                   node,
		registry:               registry,
		recorder:               recorder,
		ruleRegistry:           ruleRegistry,
		promRuleExporter:       promRuleExporter,
		updates:                updates,
		ctSender:               ctSender,
		runtime:                runtimeSampler,
		http:                   sHTTP,
		adminAuthzCheck:        adminAuthzCheck,
		admin:                  sAdmin,
		status:                 sStatus,
		drain:                  drain,
		authentication:         sAuth,
		tsDB:                   tsDB,
		tsServer:               &sTS,
		raftTransport:          raftTransport,
		stopper:                stopper,
		debug:                  debugServer,
		kvProber:               kvProber,
		replicationReporter:    replicationReporter,
		protectedtsProvider:    protectedtsProvider,
		spanConfigSubscriber:   spanConfig.subscriber,
		sqlServer:              sqlServer,
		externalStorageBuilder: externalStorageBuilder,
		storeGrantCoords:       gcoords.Stores,
		kvMemoryMonitor:        kvMemoryMonitor,
	}

	if err = startPurgeOldSessions(ctx, sAuth); err != nil {
		__antithesis_instrumentation__.Notify(195610)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(195611)
	}
	__antithesis_instrumentation__.Notify(195500)

	return lateBoundServer, err
}

func (s *Server) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(195612)
	return s.st
}

func (s *Server) AnnotateCtx(ctx context.Context) context.Context {
	__antithesis_instrumentation__.Notify(195613)
	return s.cfg.AmbientCtx.AnnotateCtx(ctx)
}

func (s *Server) AnnotateCtxWithSpan(
	ctx context.Context, opName string,
) (context.Context, *tracing.Span) {
	__antithesis_instrumentation__.Notify(195614)
	return s.cfg.AmbientCtx.AnnotateCtxWithSpan(ctx, opName)
}

func (s *Server) StorageClusterID() uuid.UUID {
	__antithesis_instrumentation__.Notify(195615)
	return s.rpcContext.StorageClusterID.Get()
}

func (s *Server) NodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(195616)
	return s.node.Descriptor.NodeID
}

func (s *Server) InitialStart() bool {
	__antithesis_instrumentation__.Notify(195617)
	return s.node.initialStart
}

type listenerInfo struct {
	listenRPC    string
	advertiseRPC string
	listenSQL    string
	advertiseSQL string
	listenHTTP   string
}

func (li listenerInfo) Iter() map[string]string {
	__antithesis_instrumentation__.Notify(195618)
	return map[string]string{
		"cockroach.listen-addr":        li.listenRPC,
		"cockroach.advertise-addr":     li.advertiseRPC,
		"cockroach.sql-addr":           li.listenSQL,
		"cockroach.advertise-sql-addr": li.advertiseSQL,
		"cockroach.http-addr":          li.listenHTTP,
	}
}

func (s *Server) Start(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(195619)
	if err := s.PreStart(ctx); err != nil {
		__antithesis_instrumentation__.Notify(195621)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195622)
	}
	__antithesis_instrumentation__.Notify(195620)
	return s.AcceptClients(ctx)
}

func (s *Server) PreStart(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(195623)
	ctx = s.AnnotateCtx(ctx)

	s.startTime = timeutil.Now()
	if err := s.startMonitoringForwardClockJumps(ctx); err != nil {
		__antithesis_instrumentation__.Notify(195660)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195661)
	}
	__antithesis_instrumentation__.Notify(195624)

	s.rpcContext.SetLocalInternalServer(s.node)

	uiTLSConfig, err := s.rpcContext.GetUIServerTLSConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(195662)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195663)
	}
	__antithesis_instrumentation__.Notify(195625)

	connManager := netutil.MakeServer(s.stopper, uiTLSConfig, http.HandlerFunc(s.http.baseHandler))

	workersCtx := s.AnnotateCtx(context.Background())

	if err := s.http.start(ctx, workersCtx, connManager, uiTLSConfig, s.stopper); err != nil {
		__antithesis_instrumentation__.Notify(195664)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195665)
	}
	__antithesis_instrumentation__.Notify(195626)

	fileTableInternalExecutor := sql.MakeInternalExecutor(ctx, s.PGServer().SQLServer, sql.MemoryMetrics{}, s.st)
	s.externalStorageBuilder.init(
		ctx,
		s.cfg.ExternalIODirConfig,
		s.st,
		s.nodeIDContainer,
		s.nodeDialer,
		s.cfg.TestingKnobs,
		&fileTableInternalExecutor,
		s.db)

	filtered := s.cfg.FilterGossipBootstrapAddresses(ctx)

	var initServer *initServer
	{
		__antithesis_instrumentation__.Notify(195666)
		dialOpts, err := s.rpcContext.GRPCDialOptions()
		if err != nil {
			__antithesis_instrumentation__.Notify(195669)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195670)
		}
		__antithesis_instrumentation__.Notify(195667)

		initConfig := newInitServerConfig(ctx, s.cfg, dialOpts)
		inspectedDiskState, err := inspectEngines(
			ctx,
			s.engines,
			s.cfg.Settings.Version.BinaryVersion(),
			s.cfg.Settings.Version.BinaryMinSupportedVersion(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(195671)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195672)
		}
		__antithesis_instrumentation__.Notify(195668)

		initServer = newInitServer(s.cfg.AmbientCtx, inspectedDiskState, initConfig)
	}
	__antithesis_instrumentation__.Notify(195627)

	initialDiskClusterVersion := initServer.DiskClusterVersion()
	{
		__antithesis_instrumentation__.Notify(195673)

		if err := kvserver.WriteClusterVersionToEngines(
			ctx, s.engines, initialDiskClusterVersion,
		); err != nil {
			__antithesis_instrumentation__.Notify(195675)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195676)
		}
		__antithesis_instrumentation__.Notify(195674)

		if err := clusterversion.Initialize(ctx, initialDiskClusterVersion.Version, &s.cfg.Settings.SV); err != nil {
			__antithesis_instrumentation__.Notify(195677)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195678)
		}

	}
	__antithesis_instrumentation__.Notify(195628)

	serverpb.RegisterInitServer(s.grpc.Server, initServer)

	migrationServer := &migrationServer{server: s}
	serverpb.RegisterMigrationServer(s.grpc.Server, migrationServer)
	s.migrationServer = migrationServer

	pgL, startRPCServer, err := startListenRPCAndSQL(ctx, workersCtx, s.cfg.BaseConfig, s.stopper, s.grpc)
	if err != nil {
		__antithesis_instrumentation__.Notify(195679)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195680)
	}
	__antithesis_instrumentation__.Notify(195629)

	if s.cfg.TestingKnobs.Server != nil {
		__antithesis_instrumentation__.Notify(195681)
		knobs := s.cfg.TestingKnobs.Server.(*TestingKnobs)
		if knobs.SignalAfterGettingRPCAddress != nil {
			__antithesis_instrumentation__.Notify(195683)
			log.Infof(ctx, "signaling caller that RPC address is ready")
			close(knobs.SignalAfterGettingRPCAddress)
		} else {
			__antithesis_instrumentation__.Notify(195684)
		}
		__antithesis_instrumentation__.Notify(195682)
		if knobs.PauseAfterGettingRPCAddress != nil {
			__antithesis_instrumentation__.Notify(195685)
			log.Infof(ctx, "waiting for signal from caller to proceed with initialization")
			select {
			case <-knobs.PauseAfterGettingRPCAddress:
				__antithesis_instrumentation__.Notify(195687)

			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(195688)

				return errors.CombineErrors(errors.New("server stopping prematurely from context shutdown"), ctx.Err())

			case <-s.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(195689)

				return errors.New("server stopping prematurely")
			}
			__antithesis_instrumentation__.Notify(195686)
			log.Infof(ctx, "caller is letting us proceed with initialization")
		} else {
			__antithesis_instrumentation__.Notify(195690)
		}
	} else {
		__antithesis_instrumentation__.Notify(195691)
	}
	__antithesis_instrumentation__.Notify(195630)

	gwMux, gwCtx, conn, err := configureGRPCGateway(
		ctx,
		workersCtx,
		s.cfg.AmbientCtx,
		s.rpcContext,
		s.stopper,
		s.grpc,
		s.cfg.AdvertiseAddr,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(195692)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195693)
	}
	__antithesis_instrumentation__.Notify(195631)

	for _, gw := range []grpcGatewayServer{s.admin, s.status, s.authentication, s.tsServer} {
		__antithesis_instrumentation__.Notify(195694)
		if err := gw.RegisterGateway(gwCtx, gwMux, conn); err != nil {
			__antithesis_instrumentation__.Notify(195695)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195696)
		}
	}
	__antithesis_instrumentation__.Notify(195632)

	s.http.handleHealth(gwMux)

	listenerFiles := listenerInfo{
		listenRPC:    s.cfg.Addr,
		advertiseRPC: s.cfg.AdvertiseAddr,
		listenSQL:    s.cfg.SQLAddr,
		advertiseSQL: s.cfg.SQLAdvertiseAddr,
		listenHTTP:   s.cfg.HTTPAdvertiseAddr,
	}.Iter()

	encryptedStore := false
	for _, storeSpec := range s.cfg.Stores.Specs {
		__antithesis_instrumentation__.Notify(195697)
		if storeSpec.InMemory {
			__antithesis_instrumentation__.Notify(195700)
			continue
		} else {
			__antithesis_instrumentation__.Notify(195701)
		}
		__antithesis_instrumentation__.Notify(195698)
		if storeSpec.IsEncrypted() {
			__antithesis_instrumentation__.Notify(195702)
			encryptedStore = true
		} else {
			__antithesis_instrumentation__.Notify(195703)
		}
		__antithesis_instrumentation__.Notify(195699)

		for name, val := range listenerFiles {
			__antithesis_instrumentation__.Notify(195704)
			file := filepath.Join(storeSpec.Path, name)
			if err := ioutil.WriteFile(file, []byte(val), 0644); err != nil {
				__antithesis_instrumentation__.Notify(195705)
				return errors.Wrapf(err, "failed to write %s", file)
			} else {
				__antithesis_instrumentation__.Notify(195706)
			}
		}
	}
	__antithesis_instrumentation__.Notify(195633)

	if s.cfg.DelayedBootstrapFn != nil {
		__antithesis_instrumentation__.Notify(195707)
		defer time.AfterFunc(30*time.Second, s.cfg.DelayedBootstrapFn).Stop()
	} else {
		__antithesis_instrumentation__.Notify(195708)
	}
	__antithesis_instrumentation__.Notify(195634)

	selfBootstrap := s.cfg.AutoInitializeCluster && func() bool {
		__antithesis_instrumentation__.Notify(195709)
		return initServer.NeedsBootstrap() == true
	}() == true
	if selfBootstrap {
		__antithesis_instrumentation__.Notify(195710)
		if _, err := initServer.Bootstrap(ctx, &serverpb.BootstrapRequest{}); err != nil {
			__antithesis_instrumentation__.Notify(195711)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195712)
		}
	} else {
		__antithesis_instrumentation__.Notify(195713)
	}
	__antithesis_instrumentation__.Notify(195635)

	var onSuccessfulReturnFn, onInitServerReady func()
	{
		__antithesis_instrumentation__.Notify(195714)
		readyFn := func(bool) { __antithesis_instrumentation__.Notify(195717) }
		__antithesis_instrumentation__.Notify(195715)
		if s.cfg.ReadyFn != nil {
			__antithesis_instrumentation__.Notify(195718)
			readyFn = s.cfg.ReadyFn
		} else {
			__antithesis_instrumentation__.Notify(195719)
		}
		__antithesis_instrumentation__.Notify(195716)
		if !initServer.NeedsBootstrap() || func() bool {
			__antithesis_instrumentation__.Notify(195720)
			return selfBootstrap == true
		}() == true {
			__antithesis_instrumentation__.Notify(195721)
			onSuccessfulReturnFn = func() { __antithesis_instrumentation__.Notify(195723); readyFn(false) }
			__antithesis_instrumentation__.Notify(195722)
			onInitServerReady = func() { __antithesis_instrumentation__.Notify(195724) }
		} else {
			__antithesis_instrumentation__.Notify(195725)
			onSuccessfulReturnFn = func() { __antithesis_instrumentation__.Notify(195727) }
			__antithesis_instrumentation__.Notify(195726)
			onInitServerReady = func() { __antithesis_instrumentation__.Notify(195728); readyFn(true) }
		}
	}
	__antithesis_instrumentation__.Notify(195636)

	startRPCServer(workersCtx)
	onInitServerReady()
	state, initialStart, err := initServer.ServeAndWait(ctx, s.stopper, &s.cfg.Settings.SV)
	if err != nil {
		__antithesis_instrumentation__.Notify(195729)
		return errors.Wrap(err, "during init")
	} else {
		__antithesis_instrumentation__.Notify(195730)
	}
	__antithesis_instrumentation__.Notify(195637)
	if err := state.validate(); err != nil {
		__antithesis_instrumentation__.Notify(195731)
		return errors.Wrap(err, "invalid init state")
	} else {
		__antithesis_instrumentation__.Notify(195732)
	}
	__antithesis_instrumentation__.Notify(195638)

	if err := initializeCachedSettings(
		ctx, keys.SystemSQLCodec, s.st.MakeUpdater(), state.initialSettingsKVs,
	); err != nil {
		__antithesis_instrumentation__.Notify(195733)
		return errors.Wrap(err, "during initializing settings updater")
	} else {
		__antithesis_instrumentation__.Notify(195734)
	}
	__antithesis_instrumentation__.Notify(195639)

	if state.clusterVersion != initialDiskClusterVersion {
		__antithesis_instrumentation__.Notify(195735)

		if err := kvserver.WriteClusterVersionToEngines(
			ctx, s.engines, state.clusterVersion,
		); err != nil {
			__antithesis_instrumentation__.Notify(195737)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195738)
		}
		__antithesis_instrumentation__.Notify(195736)

		if err := s.ClusterSettings().Version.SetActiveVersion(ctx, state.clusterVersion); err != nil {
			__antithesis_instrumentation__.Notify(195739)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195740)
		}
	} else {
		__antithesis_instrumentation__.Notify(195741)
	}
	__antithesis_instrumentation__.Notify(195640)

	s.rpcContext.StorageClusterID.Set(ctx, state.clusterID)
	s.rpcContext.NodeID.Set(ctx, state.nodeID)

	_ = s.stopper.RunAsyncTask(ctx, "connect-gossip", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(195742)
		log.Ops.Infof(ctx, "connecting to gossip network to verify cluster ID %q", state.clusterID)
		select {
		case <-s.gossip.Connected:
			__antithesis_instrumentation__.Notify(195743)
			log.Ops.Infof(ctx, "node connected via gossip")
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(195744)
		case <-s.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(195745)
		}
	})
	__antithesis_instrumentation__.Notify(195641)

	hlcUpperBoundExists, err := s.checkHLCUpperBoundExistsAndEnsureMonotonicity(ctx, initialStart)
	if err != nil {
		__antithesis_instrumentation__.Notify(195746)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195747)
	}
	__antithesis_instrumentation__.Notify(195642)

	orphanedLeasesTimeThresholdNanos := s.clock.Now().WallTime

	onSuccessfulReturnFn()

	advAddrU := util.NewUnresolvedAddr("tcp", s.cfg.AdvertiseAddr)

	s.gossip.Start(advAddrU, filtered)
	log.Event(ctx, "started gossip")

	advSQLAddrU := util.NewUnresolvedAddr("tcp", s.cfg.SQLAdvertiseAddr)

	advHTTPAddrU := util.NewUnresolvedAddr("tcp", s.cfg.HTTPAdvertiseAddr)

	if err := s.node.start(
		ctx,
		advAddrU,
		advSQLAddrU,
		advHTTPAddrU,
		*state,
		initialStart,
		s.cfg.ClusterName,
		s.cfg.NodeAttributes,
		s.cfg.Locality,
		s.cfg.LocalityAddresses,
		s.sqlServer.execCfg.DistSQLPlanner.SetSQLInstanceInfo,
	); err != nil {
		__antithesis_instrumentation__.Notify(195748)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195749)
	}
	__antithesis_instrumentation__.Notify(195643)

	log.Event(ctx, "started node")
	if err := s.startPersistingHLCUpperBound(ctx, hlcUpperBoundExists); err != nil {
		__antithesis_instrumentation__.Notify(195750)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195751)
	}
	__antithesis_instrumentation__.Notify(195644)
	s.replicationReporter.Start(ctx, s.stopper)

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		__antithesis_instrumentation__.Notify(195752)
		scope.SetTags(map[string]string{
			"cluster":         s.StorageClusterID().String(),
			"node":            s.NodeID().String(),
			"server_id":       fmt.Sprintf("%s-%s", s.StorageClusterID().Short(), s.NodeID()),
			"engine_type":     s.cfg.StorageEngine.String(),
			"encrypted_store": strconv.FormatBool(encryptedStore),
		})
	})
	__antithesis_instrumentation__.Notify(195645)

	s.recorder.AddNode(
		s.registry,
		s.node.Descriptor,
		s.node.startedAt,
		s.cfg.AdvertiseAddr,
		s.cfg.HTTPAdvertiseAddr,
		s.cfg.SQLAdvertiseAddr,
	)

	if err := startSampleEnvironment(s.AnnotateCtx(ctx),
		s.ClusterSettings(),
		s.stopper,
		s.cfg.GoroutineDumpDirName,
		s.cfg.HeapProfileDirName,
		s.runtime,
		s.status.sessionRegistry,
	); err != nil {
		__antithesis_instrumentation__.Notify(195753)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195754)
	}
	__antithesis_instrumentation__.Notify(195646)

	var graphiteOnce sync.Once
	graphiteEndpoint.SetOnChange(&s.st.SV, func(context.Context) {
		__antithesis_instrumentation__.Notify(195755)
		if graphiteEndpoint.Get(&s.st.SV) != "" {
			__antithesis_instrumentation__.Notify(195756)
			graphiteOnce.Do(func() {
				__antithesis_instrumentation__.Notify(195757)
				s.node.startGraphiteStatsExporter(s.st)
			})
		} else {
			__antithesis_instrumentation__.Notify(195758)
		}
	})
	__antithesis_instrumentation__.Notify(195647)

	if err := s.protectedtsProvider.Start(ctx, s.stopper); err != nil {
		__antithesis_instrumentation__.Notify(195759)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195760)
	}
	__antithesis_instrumentation__.Notify(195648)

	s.grpc.setMode(modeOperational)

	s.node.waitForAdditionalStoreInit()

	s.storeGrantCoords.SetPebbleMetricsProvider(ctx, s.node)

	logPendingLossOfQuorumRecoveryEvents(ctx, s.node.stores)

	log.Ops.Infof(ctx, "starting %s server at %s (use: %s)",
		redact.Safe(s.cfg.HTTPRequestScheme()), s.cfg.HTTPAddr, s.cfg.HTTPAdvertiseAddr)
	rpcConnType := redact.SafeString("grpc/postgres")
	if s.cfg.SplitListenSQL {
		__antithesis_instrumentation__.Notify(195761)
		rpcConnType = "grpc"
		log.Ops.Infof(ctx, "starting postgres server at %s (use: %s)", s.cfg.SQLAddr, s.cfg.SQLAdvertiseAddr)
	} else {
		__antithesis_instrumentation__.Notify(195762)
	}
	__antithesis_instrumentation__.Notify(195649)
	log.Ops.Infof(ctx, "starting %s server at %s", rpcConnType, s.cfg.Addr)
	log.Ops.Infof(ctx, "advertising CockroachDB node at %s", s.cfg.AdvertiseAddr)

	log.Event(ctx, "accepting connections")

	s.nodeLiveness.Start(ctx, liveness.NodeLivenessStartOptions{
		Stopper: s.stopper,
		Engines: s.engines,
		OnSelfLive: func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(195763)
			now := s.clock.Now()
			if err := s.node.stores.VisitStores(func(s *kvserver.Store) error {
				__antithesis_instrumentation__.Notify(195764)
				return s.WriteLastUpTimestamp(ctx, now)
			}); err != nil {
				__antithesis_instrumentation__.Notify(195765)
				log.Ops.Warningf(ctx, "writing last up timestamp: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(195766)
			}
		},
	})
	__antithesis_instrumentation__.Notify(195650)

	if err := s.node.startWriteNodeStatus(base.DefaultMetricsSampleInterval); err != nil {
		__antithesis_instrumentation__.Notify(195767)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195768)
	}
	__antithesis_instrumentation__.Notify(195651)

	if !s.cfg.SpanConfigsDisabled && func() bool {
		__antithesis_instrumentation__.Notify(195769)
		return s.spanConfigSubscriber != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(195770)
		if subscriber, ok := s.spanConfigSubscriber.(*spanconfigkvsubscriber.KVSubscriber); ok {
			__antithesis_instrumentation__.Notify(195771)
			if err := subscriber.Start(ctx, s.stopper); err != nil {
				__antithesis_instrumentation__.Notify(195772)
				return err
			} else {
				__antithesis_instrumentation__.Notify(195773)
			}
		} else {
			__antithesis_instrumentation__.Notify(195774)
		}
	} else {
		__antithesis_instrumentation__.Notify(195775)
	}
	__antithesis_instrumentation__.Notify(195652)

	s.startSystemLogsGC(ctx)

	if err := s.http.setupRoutes(ctx,
		s.authentication,
		s.adminAuthzCheck,
		s.recorder,
		s.runtime,
		gwMux,
		s.debug,
		newAPIV2Server(ctx, s),
	); err != nil {
		__antithesis_instrumentation__.Notify(195776)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195777)
	}
	__antithesis_instrumentation__.Notify(195653)

	nodeStartCounter := "storage.engine."
	switch s.cfg.StorageEngine {
	case enginepb.EngineTypeDefault:
		__antithesis_instrumentation__.Notify(195778)
		fallthrough
	case enginepb.EngineTypePebble:
		__antithesis_instrumentation__.Notify(195779)
		nodeStartCounter += "pebble."
	default:
		__antithesis_instrumentation__.Notify(195780)
	}
	__antithesis_instrumentation__.Notify(195654)
	if s.InitialStart() {
		__antithesis_instrumentation__.Notify(195781)
		nodeStartCounter += "initial-boot"
	} else {
		__antithesis_instrumentation__.Notify(195782)
		nodeStartCounter += "restart"
	}
	__antithesis_instrumentation__.Notify(195655)
	telemetry.Count(nodeStartCounter)

	s.node.recordJoinEvent(ctx)

	if err := s.sqlServer.preStart(
		workersCtx,
		s.stopper,
		s.cfg.TestingKnobs,
		connManager,
		pgL,
		s.cfg.SocketFile,
		orphanedLeasesTimeThresholdNanos,
	); err != nil {
		__antithesis_instrumentation__.Notify(195783)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195784)
	}
	__antithesis_instrumentation__.Notify(195656)

	if err := s.debug.RegisterEngines(s.cfg.Stores.Specs, s.engines); err != nil {
		__antithesis_instrumentation__.Notify(195785)
		return errors.Wrapf(err, "failed to register engines with debug server")
	} else {
		__antithesis_instrumentation__.Notify(195786)
	}
	__antithesis_instrumentation__.Notify(195657)
	s.debug.RegisterClosedTimestampSideTransport(s.ctSender, s.node.storeCfg.ClosedTimestampReceiver)

	s.ctSender.Run(ctx, state.nodeID)

	s.startAttemptUpgrade(ctx)

	if err := s.node.tenantSettingsWatcher.Start(ctx, s.sqlServer.execCfg.SystemTableIDResolver); err != nil {
		__antithesis_instrumentation__.Notify(195787)
		return errors.Wrap(err, "failed to initialize the tenant settings watcher")
	} else {
		__antithesis_instrumentation__.Notify(195788)
	}
	__antithesis_instrumentation__.Notify(195658)

	if err := s.kvProber.Start(ctx, s.stopper); err != nil {
		__antithesis_instrumentation__.Notify(195789)
		return errors.Wrapf(err, "failed to start KV prober")
	} else {
		__antithesis_instrumentation__.Notify(195790)
	}
	__antithesis_instrumentation__.Notify(195659)

	publishPendingLossOfQuorumRecoveryEvents(ctx, s.node.stores, s.stopper)

	log.Event(ctx, "server initialized")

	s.tsDB.PollSource(
		s.cfg.AmbientCtx, s.recorder, base.DefaultMetricsSampleInterval, ts.Resolution10s, s.stopper,
	)

	return maybeImportTS(ctx, s)
}

func (s *Server) AcceptClients(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(195791)
	workersCtx := s.AnnotateCtx(context.Background())

	if err := s.sqlServer.startServeSQL(
		workersCtx,
		s.stopper,
		s.sqlServer.connManager,
		s.sqlServer.pgL,
		s.cfg.SocketFile,
	); err != nil {
		__antithesis_instrumentation__.Notify(195793)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195794)
	}
	__antithesis_instrumentation__.Notify(195792)

	log.Event(ctx, "server ready")
	return nil
}

func (s *Server) Stop() {
	__antithesis_instrumentation__.Notify(195795)
	s.stopper.Stop(context.Background())
}

func (s *Server) TempDir() string {
	__antithesis_instrumentation__.Notify(195796)
	return s.cfg.TempStorageConfig.Path
}

func (s *Server) PGServer() *pgwire.Server {
	__antithesis_instrumentation__.Notify(195797)
	return s.sqlServer.pgServer
}

func (s *Server) StartDiagnostics(ctx context.Context) {
	__antithesis_instrumentation__.Notify(195798)
	s.updates.PeriodicallyCheckForUpdates(ctx, s.stopper)
	s.sqlServer.StartDiagnostics(ctx)
}

func init() {
	tracing.RegisterTagRemapping("n", "node")
}

func (s *Server) RunLocalSQL(
	ctx context.Context, fn func(ctx context.Context, sqlExec *sql.InternalExecutor) error,
) error {
	__antithesis_instrumentation__.Notify(195799)
	return fn(ctx, s.sqlServer.internalExecutor)
}

func (s *Server) Insecure() bool {
	__antithesis_instrumentation__.Notify(195800)
	return s.cfg.Insecure
}

func (s *Server) Drain(
	ctx context.Context, verbose bool,
) (remaining uint64, info redact.RedactableString, err error) {
	__antithesis_instrumentation__.Notify(195801)
	return s.drain.runDrain(ctx, verbose)
}
