package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/http"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobsprotectedts"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvtenant"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptprovider"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptreconcile"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/server/debug"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/server/systemconfigwatcher"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/optionalnodeliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func StartTenant(
	ctx context.Context,
	stopper *stop.Stopper,
	kvClusterName string,
	baseCfg BaseConfig,
	sqlCfg SQLConfig,
) (*SQLServerWrapper, error) {
	__antithesis_instrumentation__.Notify(238271)
	sqlServer, authServer, drainServer, pgAddr, httpAddr, err := startTenantInternal(ctx, stopper, kvClusterName, baseCfg, sqlCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(238273)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(238274)
	}
	__antithesis_instrumentation__.Notify(238272)
	return &SQLServerWrapper{
		SQLServer:   sqlServer,
		authServer:  authServer,
		drainServer: drainServer,
		pgAddr:      pgAddr,
		httpAddr:    httpAddr,
	}, err
}

type SQLServerWrapper struct {
	*SQLServer
	authServer  *authenticationServer
	drainServer *drainServer
	pgAddr      string
	httpAddr    string
}

func (s *SQLServerWrapper) Drain(
	ctx context.Context, verbose bool,
) (remaining uint64, info redact.RedactableString, err error) {
	__antithesis_instrumentation__.Notify(238275)
	return s.drainServer.runDrain(ctx, verbose)
}

func startTenantInternal(
	ctx context.Context,
	stopper *stop.Stopper,
	kvClusterName string,
	baseCfg BaseConfig,
	sqlCfg SQLConfig,
) (
	sqlServer *SQLServer,
	authServer *authenticationServer,
	drainServer *drainServer,
	pgAddr string,
	httpAddr string,
	_ error,
) {
	__antithesis_instrumentation__.Notify(238276)
	err := ApplyTenantLicense()
	if err != nil {
		__antithesis_instrumentation__.Notify(238295)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238296)
	}
	__antithesis_instrumentation__.Notify(238277)

	baseCfg.idProvider.SetTenant(sqlCfg.TenantID)

	args, err := makeTenantSQLServerArgs(ctx, stopper, kvClusterName, baseCfg, sqlCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(238297)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238298)
	}
	__antithesis_instrumentation__.Notify(238278)
	err = args.ValidateAddrs(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(238299)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238300)
	}
	__antithesis_instrumentation__.Notify(238279)
	args.monitorAndMetrics = newRootSQLMemoryMonitor(monitorAndMetricsOptions{
		memoryPoolSize:          args.MemoryPoolSize,
		histogramWindowInterval: args.HistogramWindowInterval(),
		settings:                args.Settings,
	})

	grpcMain := newGRPCServer(args.rpcContext)
	grpcMain.setMode(modeOperational)

	args.grpcServer = grpcMain.Server

	baseCfg.SplitListenSQL = false

	ctx = args.BaseConfig.AmbientCtx.AnnotateCtx(ctx)

	background := args.BaseConfig.AmbientCtx.AnnotateCtx(context.Background())

	baseCfg.Addr = baseCfg.SQLAddr
	baseCfg.AdvertiseAddr = baseCfg.SQLAdvertiseAddr
	pgL, startRPCServer, err := startListenRPCAndSQL(ctx, background, baseCfg, stopper, grpcMain)
	if err != nil {
		__antithesis_instrumentation__.Notify(238301)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238302)
	}

	{
		__antithesis_instrumentation__.Notify(238303)
		waitQuiesce := func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(238305)
			<-args.stopper.ShouldQuiesce()

			_ = pgL.Close()
		}
		__antithesis_instrumentation__.Notify(238304)
		if err := args.stopper.RunAsyncTask(background, "wait-quiesce-pgl", waitQuiesce); err != nil {
			__antithesis_instrumentation__.Notify(238306)
			waitQuiesce(background)
			return nil, nil, nil, "", "", err
		} else {
			__antithesis_instrumentation__.Notify(238307)
		}
	}
	__antithesis_instrumentation__.Notify(238280)

	serverTLSConfig, err := args.rpcContext.GetUIServerTLSConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(238308)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238309)
	}
	__antithesis_instrumentation__.Notify(238281)

	args.advertiseAddr = baseCfg.AdvertiseAddr

	tenantStatusServer := newTenantStatusServer(
		baseCfg.AmbientCtx, &adminPrivilegeChecker{ie: args.circularInternalExecutor},
		args.sessionRegistry, args.flowScheduler, baseCfg.Settings, nil,
		args.rpcContext, args.stopper,
	)

	args.sqlStatusServer = tenantStatusServer
	s, err := newSQLServer(ctx, args)
	tenantStatusServer.sqlServer = s

	if err != nil {
		__antithesis_instrumentation__.Notify(238310)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238311)
	}
	__antithesis_instrumentation__.Notify(238282)

	drainServer = newDrainServer(baseCfg, args.stopper, args.grpc, s)

	tenantAdminServer := newTenantAdminServer(baseCfg.AmbientCtx, s, tenantStatusServer, drainServer)

	s.execCfg.DistSQLPlanner.SetSQLInstanceInfo(roachpb.NodeDescriptor{NodeID: 0})

	authServer = newAuthenticationServer(baseCfg.Config, s)

	for _, gw := range []grpcGatewayServer{tenantAdminServer, tenantStatusServer, authServer} {
		__antithesis_instrumentation__.Notify(238312)
		gw.RegisterService(grpcMain.Server)
	}
	__antithesis_instrumentation__.Notify(238283)
	startRPCServer(background)

	gwMux, gwCtx, conn, err := configureGRPCGateway(
		ctx,
		background,
		args.AmbientCtx,
		args.rpcContext,
		s.stopper,
		grpcMain,
		baseCfg.AdvertiseAddr,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(238313)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238314)
	}
	__antithesis_instrumentation__.Notify(238284)

	for _, gw := range []grpcGatewayServer{tenantAdminServer, tenantStatusServer, authServer} {
		__antithesis_instrumentation__.Notify(238315)
		if err := gw.RegisterGateway(gwCtx, gwMux, conn); err != nil {
			__antithesis_instrumentation__.Notify(238316)
			return nil, nil, nil, "", "", err
		} else {
			__antithesis_instrumentation__.Notify(238317)
		}
	}
	__antithesis_instrumentation__.Notify(238285)

	debugServer := debug.NewServer(baseCfg.AmbientCtx, args.Settings, s.pgServer.HBADebugFn(), s.execCfg.SQLStatusServer)
	adminAuthzCheck := &adminPrivilegeChecker{ie: s.execCfg.InternalExecutor}

	parseNodeIDFn := func(s string) (roachpb.NodeID, bool, error) {
		__antithesis_instrumentation__.Notify(238318)
		return roachpb.NodeID(0), false, errors.New("tenants cannot proxy to KV Nodes")
	}
	__antithesis_instrumentation__.Notify(238286)
	getNodeIDHTTPAddressFn := func(id roachpb.NodeID) (*util.UnresolvedAddr, error) {
		__antithesis_instrumentation__.Notify(238319)
		return nil, errors.New("tenants cannot proxy to KV Nodes")
	}
	__antithesis_instrumentation__.Notify(238287)
	httpServer := newHTTPServer(baseCfg, args.rpcContext, parseNodeIDFn, getNodeIDHTTPAddressFn)

	httpServer.handleHealth(gwMux)

	if err := httpServer.setupRoutes(ctx,
		authServer,
		adminAuthzCheck,
		args.recorder,
		args.runtime,
		gwMux,
		debugServer,
		nil,
	); err != nil {
		__antithesis_instrumentation__.Notify(238320)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238321)
	}
	__antithesis_instrumentation__.Notify(238288)

	connManager := netutil.MakeServer(
		args.stopper,
		serverTLSConfig,
		http.HandlerFunc(httpServer.baseHandler),
	)
	if err := httpServer.start(ctx, background, connManager, serverTLSConfig, args.stopper); err != nil {
		__antithesis_instrumentation__.Notify(238322)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238323)
	}
	__antithesis_instrumentation__.Notify(238289)

	args.recorder.AddNode(
		args.registry,
		roachpb.NodeDescriptor{},
		timeutil.Now().UnixNano(),
		baseCfg.AdvertiseAddr,
		baseCfg.HTTPAdvertiseAddr,
		baseCfg.SQLAdvertiseAddr,
	)

	const (
		socketFile = ""
	)
	orphanedLeasesTimeThresholdNanos := args.clock.Now().WallTime

	if err := startSampleEnvironment(ctx,
		args.Settings,
		args.stopper,
		args.GoroutineDumpDirName,
		args.HeapProfileDirName,
		args.runtime,
		args.sessionRegistry,
	); err != nil {
		__antithesis_instrumentation__.Notify(238324)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238325)
	}
	__antithesis_instrumentation__.Notify(238290)

	if err := s.preStart(ctx,
		args.stopper,
		args.TestingKnobs,
		connManager,
		pgL,
		socketFile,
		orphanedLeasesTimeThresholdNanos,
	); err != nil {
		__antithesis_instrumentation__.Notify(238326)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238327)
	}
	__antithesis_instrumentation__.Notify(238291)

	externalUsageFn := func(ctx context.Context) multitenant.ExternalUsage {
		__antithesis_instrumentation__.Notify(238328)
		userTimeMillis, _, err := status.GetCPUTime(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(238330)
			log.Ops.Errorf(ctx, "unable to get cpu usage: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(238331)
		}
		__antithesis_instrumentation__.Notify(238329)
		return multitenant.ExternalUsage{
			CPUSecs:           float64(userTimeMillis) * 1e-3,
			PGWireEgressBytes: s.pgServer.BytesOut(),
		}
	}
	__antithesis_instrumentation__.Notify(238292)

	nextLiveInstanceIDFn := makeNextLiveInstanceIDFn(s.sqlInstanceProvider, s.SQLInstanceID())

	if err := args.costController.Start(
		ctx, args.stopper, s.SQLInstanceID(), s.sqlLivenessSessionID,
		externalUsageFn, nextLiveInstanceIDFn,
	); err != nil {
		__antithesis_instrumentation__.Notify(238332)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238333)
	}
	__antithesis_instrumentation__.Notify(238293)

	if err := s.startServeSQL(ctx,
		args.stopper,
		s.connManager,
		s.pgL,
		socketFile); err != nil {
		__antithesis_instrumentation__.Notify(238334)
		return nil, nil, nil, "", "", err
	} else {
		__antithesis_instrumentation__.Notify(238335)
	}
	__antithesis_instrumentation__.Notify(238294)

	return s, authServer, drainServer, baseCfg.SQLAddr, baseCfg.HTTPAddr, nil
}

func makeTenantSQLServerArgs(
	startupCtx context.Context,
	stopper *stop.Stopper,
	kvClusterName string,
	baseCfg BaseConfig,
	sqlCfg SQLConfig,
) (sqlServerArgs, error) {
	__antithesis_instrumentation__.Notify(238336)
	st := baseCfg.Settings

	instanceIDContainer := baseCfg.IDContainer.SwitchToSQLIDContainer()
	startupCtx = baseCfg.AmbientCtx.AnnotateCtx(startupCtx)

	baseCfg.ClusterName = kvClusterName

	clock := hlc.NewClock(hlc.UnixNano, time.Duration(baseCfg.MaxOffset))

	registry := metric.NewRegistry()

	var rpcTestingKnobs rpc.ContextTestingKnobs
	if p, ok := baseCfg.TestingKnobs.Server.(*TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(238345)
		rpcTestingKnobs = p.ContextTestingKnobs
	} else {
		__antithesis_instrumentation__.Notify(238346)
	}
	__antithesis_instrumentation__.Notify(238337)
	rpcContext := rpc.NewContext(startupCtx, rpc.ContextOptions{
		TenantID:         sqlCfg.TenantID,
		NodeID:           baseCfg.IDContainer,
		StorageClusterID: baseCfg.ClusterIDContainer,
		Config:           baseCfg.Config,
		Clock:            clock,
		Stopper:          stopper,
		Settings:         st,
		Knobs:            rpcTestingKnobs,
	})

	var dsKnobs kvcoord.ClientTestingKnobs
	if dsKnobsP, ok := baseCfg.TestingKnobs.DistSQL.(*kvcoord.ClientTestingKnobs); ok {
		__antithesis_instrumentation__.Notify(238347)
		dsKnobs = *dsKnobsP
	} else {
		__antithesis_instrumentation__.Notify(238348)
	}
	__antithesis_instrumentation__.Notify(238338)
	rpcRetryOptions := base.DefaultRetryOptions()

	tcCfg := kvtenant.ConnectorConfig{
		TenantID:          sqlCfg.TenantID,
		AmbientCtx:        baseCfg.AmbientCtx,
		RPCContext:        rpcContext,
		RPCRetryOptions:   rpcRetryOptions,
		DefaultZoneConfig: &baseCfg.DefaultZoneConfig,
	}
	tenantConnect, err := kvtenant.Factory.NewConnector(tcCfg, sqlCfg.TenantKVAddrs)
	if err != nil {
		__antithesis_instrumentation__.Notify(238349)
		return sqlServerArgs{}, err
	} else {
		__antithesis_instrumentation__.Notify(238350)
	}
	__antithesis_instrumentation__.Notify(238339)
	resolver := kvtenant.AddressResolver(tenantConnect)
	nodeDialer := nodedialer.New(rpcContext, resolver)

	provider := kvtenant.TokenBucketProvider(tenantConnect)
	if tenantKnobs, ok := baseCfg.TestingKnobs.TenantTestingKnobs.(*sql.TenantTestingKnobs); ok && func() bool {
		__antithesis_instrumentation__.Notify(238351)
		return tenantKnobs.OverrideTokenBucketProvider != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(238352)
		provider = tenantKnobs.OverrideTokenBucketProvider(provider)
	} else {
		__antithesis_instrumentation__.Notify(238353)
	}
	__antithesis_instrumentation__.Notify(238340)
	costController, err := NewTenantSideCostController(st, sqlCfg.TenantID, provider)
	if err != nil {
		__antithesis_instrumentation__.Notify(238354)
		return sqlServerArgs{}, err
	} else {
		__antithesis_instrumentation__.Notify(238355)
	}
	__antithesis_instrumentation__.Notify(238341)

	dsCfg := kvcoord.DistSenderConfig{
		AmbientCtx:        baseCfg.AmbientCtx,
		Settings:          st,
		Clock:             clock,
		NodeDescs:         tenantConnect,
		RPCRetryOptions:   &rpcRetryOptions,
		RPCContext:        rpcContext,
		NodeDialer:        nodeDialer,
		RangeDescriptorDB: tenantConnect,
		KVInterceptor:     costController,
		TestingKnobs:      dsKnobs,
	}
	ds := kvcoord.NewDistSender(dsCfg)

	var clientKnobs kvcoord.ClientTestingKnobs
	if p, ok := baseCfg.TestingKnobs.KVClient.(*kvcoord.ClientTestingKnobs); ok {
		__antithesis_instrumentation__.Notify(238356)
		clientKnobs = *p
	} else {
		__antithesis_instrumentation__.Notify(238357)
	}
	__antithesis_instrumentation__.Notify(238342)

	txnMetrics := kvcoord.MakeTxnMetrics(baseCfg.HistogramWindowInterval())
	registry.AddMetricStruct(txnMetrics)
	tcsFactory := kvcoord.NewTxnCoordSenderFactory(
		kvcoord.TxnCoordSenderFactoryConfig{
			AmbientCtx:        baseCfg.AmbientCtx,
			Settings:          st,
			Clock:             clock,
			Stopper:           stopper,
			HeartbeatInterval: base.DefaultTxnHeartbeatInterval,
			Linearizable:      false,
			Metrics:           txnMetrics,
			TestingKnobs:      clientKnobs,
		},
		ds,
	)
	db := kv.NewDB(baseCfg.AmbientCtx, tcsFactory, clock, stopper)
	rangeFeedKnobs, _ := baseCfg.TestingKnobs.RangeFeed.(*rangefeed.TestingKnobs)
	rangeFeedFactory, err := rangefeed.NewFactory(stopper, db, st, rangeFeedKnobs)
	if err != nil {
		__antithesis_instrumentation__.Notify(238358)
		return sqlServerArgs{}, err
	} else {
		__antithesis_instrumentation__.Notify(238359)
	}
	__antithesis_instrumentation__.Notify(238343)

	systemConfigWatcher := systemconfigwatcher.NewWithAdditionalProvider(
		keys.MakeSQLCodec(sqlCfg.TenantID), clock, rangeFeedFactory, &baseCfg.DefaultZoneConfig,
		tenantConnect,
	)

	circularInternalExecutor := &sql.InternalExecutor{}
	circularJobRegistry := &jobs.Registry{}

	var protectedTSProvider protectedts.Provider
	protectedtsKnobs, _ := baseCfg.TestingKnobs.ProtectedTS.(*protectedts.TestingKnobs)
	pp, err := ptprovider.New(ptprovider.Config{
		DB:               db,
		InternalExecutor: circularInternalExecutor,
		Settings:         st,
		Knobs:            protectedtsKnobs,
		ReconcileStatusFuncs: ptreconcile.StatusFuncs{
			jobsprotectedts.GetMetaType(jobsprotectedts.Jobs): jobsprotectedts.MakeStatusFunc(
				circularJobRegistry, circularInternalExecutor, jobsprotectedts.Jobs),
			jobsprotectedts.GetMetaType(jobsprotectedts.Schedules): jobsprotectedts.MakeStatusFunc(
				circularJobRegistry, circularInternalExecutor, jobsprotectedts.Schedules),
		},
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(238360)
		return sqlServerArgs{}, err
	} else {
		__antithesis_instrumentation__.Notify(238361)
	}
	__antithesis_instrumentation__.Notify(238344)
	registry.AddMetricStruct(pp.Metrics())
	protectedTSProvider = tenantProtectedTSProvider{Provider: pp, st: st}

	recorder := status.NewMetricsRecorder(clock, nil, rpcContext, nil, st)

	runtime := status.NewRuntimeStatSampler(startupCtx, clock)
	registry.AddMetricStruct(runtime)

	esb := &externalStorageBuilder{}
	externalStorage := esb.makeExternalStorage
	externalStorageFromURI := esb.makeExternalStorageFromURI

	esb.init(
		startupCtx,
		sqlCfg.ExternalIODirConfig,
		baseCfg.Settings,
		baseCfg.IDContainer,
		nodeDialer,
		baseCfg.TestingKnobs,
		circularInternalExecutor,
		db,
	)

	grpcServer := newGRPCServer(rpcContext)

	grpcServer.setMode(modeOperational)

	sessionRegistry := sql.NewSessionRegistry()
	flowScheduler := flowinfra.NewFlowScheduler(baseCfg.AmbientCtx, stopper, st)
	return sqlServerArgs{
		sqlServerOptionalKVArgs: sqlServerOptionalKVArgs{
			nodesStatusServer: serverpb.MakeOptionalNodesStatusServer(nil),
			nodeLiveness:      optionalnodeliveness.MakeContainer(nil),
			gossip:            gossip.MakeOptionalGossip(nil),
			grpcServer:        grpcServer.Server,
			isMeta1Leaseholder: func(_ context.Context, _ hlc.ClockTimestamp) (bool, error) {
				__antithesis_instrumentation__.Notify(238362)
				return false, errors.New("isMeta1Leaseholder is not available to secondary tenants")
			},
			externalStorage:        externalStorage,
			externalStorageFromURI: externalStorageFromURI,

			nodeIDContainer:      instanceIDContainer,
			spanConfigKVAccessor: tenantConnect,
			kvStoresIterator:     kvserverbase.UnsupportedStoresIterator{},
		},
		sqlServerOptionalTenantArgs: sqlServerOptionalTenantArgs{
			tenantConnect: tenantConnect,
		},
		SQLConfig:                &sqlCfg,
		BaseConfig:               &baseCfg,
		stopper:                  stopper,
		clock:                    clock,
		runtime:                  runtime,
		rpcContext:               rpcContext,
		nodeDescs:                tenantConnect,
		systemConfigWatcher:      systemConfigWatcher,
		spanConfigAccessor:       tenantConnect,
		nodeDialer:               nodeDialer,
		distSender:               ds,
		db:                       db,
		registry:                 registry,
		recorder:                 recorder,
		sessionRegistry:          sessionRegistry,
		flowScheduler:            flowScheduler,
		circularInternalExecutor: circularInternalExecutor,
		circularJobRegistry:      circularJobRegistry,
		protectedtsProvider:      protectedTSProvider,
		rangeFeedFactory:         rangeFeedFactory,
		regionsServer:            tenantConnect,
		tenantStatusServer:       tenantConnect,
		costController:           costController,
		grpc:                     grpcServer,
	}, nil
}

func makeNextLiveInstanceIDFn(
	sqlInstanceProvider sqlinstance.Provider, instanceID base.SQLInstanceID,
) multitenant.NextLiveInstanceIDFn {
	__antithesis_instrumentation__.Notify(238363)
	return func(ctx context.Context) base.SQLInstanceID {
		__antithesis_instrumentation__.Notify(238364)
		instances, err := sqlInstanceProvider.GetAllInstances(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(238369)
			log.Infof(ctx, "GetAllInstances failed: %v", err)
			return 0
		} else {
			__antithesis_instrumentation__.Notify(238370)
		}
		__antithesis_instrumentation__.Notify(238365)
		if len(instances) == 0 {
			__antithesis_instrumentation__.Notify(238371)
			return 0
		} else {
			__antithesis_instrumentation__.Notify(238372)
		}
		__antithesis_instrumentation__.Notify(238366)

		var minID, nextID base.SQLInstanceID
		for i := range instances {
			__antithesis_instrumentation__.Notify(238373)
			id := instances[i].InstanceID
			if minID == 0 || func() bool {
				__antithesis_instrumentation__.Notify(238375)
				return minID > id == true
			}() == true {
				__antithesis_instrumentation__.Notify(238376)
				minID = id
			} else {
				__antithesis_instrumentation__.Notify(238377)
			}
			__antithesis_instrumentation__.Notify(238374)
			if id > instanceID && func() bool {
				__antithesis_instrumentation__.Notify(238378)
				return (nextID == 0 || func() bool {
					__antithesis_instrumentation__.Notify(238379)
					return nextID > id == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(238380)
				nextID = id
			} else {
				__antithesis_instrumentation__.Notify(238381)
			}
		}
		__antithesis_instrumentation__.Notify(238367)
		if nextID == 0 {
			__antithesis_instrumentation__.Notify(238382)
			return minID
		} else {
			__antithesis_instrumentation__.Notify(238383)
		}
		__antithesis_instrumentation__.Notify(238368)
		return nextID
	}
}

var NewTenantSideCostController = func(
	st *cluster.Settings, tenantID roachpb.TenantID, provider kvtenant.TokenBucketProvider,
) (multitenant.TenantSideCostController, error) {
	__antithesis_instrumentation__.Notify(238384)

	return noopTenantSideCostController{}, nil
}

var ApplyTenantLicense = func() error { __antithesis_instrumentation__.Notify(238385); return nil }

type noopTenantSideCostController struct{}

var _ multitenant.TenantSideCostController = noopTenantSideCostController{}

func (noopTenantSideCostController) Start(
	ctx context.Context,
	stopper *stop.Stopper,
	instanceID base.SQLInstanceID,
	sessionID sqlliveness.SessionID,
	externalUsageFn multitenant.ExternalUsageFn,
	nextLiveInstanceIDFn multitenant.NextLiveInstanceIDFn,
) error {
	__antithesis_instrumentation__.Notify(238386)
	return nil
}

func (noopTenantSideCostController) OnRequestWait(
	ctx context.Context, info tenantcostmodel.RequestInfo,
) error {
	__antithesis_instrumentation__.Notify(238387)
	return nil
}

func (noopTenantSideCostController) OnResponse(
	ctx context.Context, req tenantcostmodel.RequestInfo, resp tenantcostmodel.ResponseInfo,
) {
	__antithesis_instrumentation__.Notify(238388)
}
