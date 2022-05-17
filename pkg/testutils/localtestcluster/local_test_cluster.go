package localtestcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/sidetransport"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/systemconfigwatcher"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig/spanconfigkvsubscriber"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

type LocalTestCluster struct {
	AmbientCtx        log.AmbientContext
	Cfg               kvserver.StoreConfig
	Manual            *hlc.ManualClock
	Clock             *hlc.Clock
	Gossip            *gossip.Gossip
	Eng               storage.Engine
	Store             *kvserver.Store
	StoreTestingKnobs *kvserver.StoreTestingKnobs
	dbContext         *kv.DBContext
	DB                *kv.DB
	Stores            *kvserver.Stores
	stopper           *stop.Stopper
	Latency           time.Duration
	tester            testing.TB

	DisableLivenessHeartbeat bool

	DontCreateSystemRanges bool
}

type InitFactoryFn func(
	ctx context.Context,
	st *cluster.Settings,
	nodeDesc *roachpb.NodeDescriptor,
	tracer *tracing.Tracer,
	clock *hlc.Clock,
	latency time.Duration,
	stores kv.Sender,
	stopper *stop.Stopper,
	gossip *gossip.Gossip,
) kv.TxnSenderFactory

func (ltc *LocalTestCluster) Stopper() *stop.Stopper {
	__antithesis_instrumentation__.Notify(645517)
	return ltc.stopper
}

func (ltc *LocalTestCluster) Start(t testing.TB, baseCtx *base.Config, initFactory InitFactoryFn) {
	__antithesis_instrumentation__.Notify(645518)
	manualClock := hlc.NewManualClock(123)
	clock := hlc.NewClock(manualClock.UnixNano, 50*time.Millisecond)
	cfg := kvserver.TestStoreConfig(clock)
	tr := cfg.AmbientCtx.Tracer
	ltc.stopper = stop.NewStopper(stop.WithTracer(tr))
	ltc.Manual = manualClock
	ltc.Clock = clock
	ambient := cfg.AmbientCtx

	nc := &base.NodeIDContainer{}
	ambient.AddLogTag("n", nc)
	ltc.AmbientCtx = ambient
	ctx := ambient.AnnotateCtx(context.Background())

	nodeID := roachpb.NodeID(1)
	nodeDesc := &roachpb.NodeDescriptor{
		NodeID:  nodeID,
		Address: util.MakeUnresolvedAddr("tcp", "invalid.invalid:26257"),
	}

	ltc.tester = t
	cfg.RPCContext = rpc.NewContext(ctx, rpc.ContextOptions{
		TenantID: roachpb.SystemTenantID,
		Config:   baseCtx,
		Clock:    ltc.Clock,
		Stopper:  ltc.stopper,
		Settings: cfg.Settings,
		NodeID:   nc,
	})
	cfg.RPCContext.NodeID.Set(ctx, nodeID)
	clusterID := cfg.RPCContext.StorageClusterID
	server := rpc.NewServer(cfg.RPCContext)
	ltc.Gossip = gossip.New(ambient, clusterID, nc, cfg.RPCContext, server, ltc.stopper, metric.NewRegistry(), roachpb.Locality{}, zonepb.DefaultZoneConfigRef())
	var err error
	ltc.Eng, err = storage.Open(
		ctx,
		storage.InMemory(),
		storage.CacheSize(0),
		storage.MaxSize(50<<20),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(645529)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(645530)
	}
	__antithesis_instrumentation__.Notify(645519)
	ltc.stopper.AddCloser(ltc.Eng)

	ltc.Stores = kvserver.NewStores(ambient, ltc.Clock)

	factory := initFactory(ctx, cfg.Settings, nodeDesc, ltc.stopper.Tracer(), ltc.Clock, ltc.Latency, ltc.Stores, ltc.stopper, ltc.Gossip)

	var nodeIDContainer base.NodeIDContainer
	nodeIDContainer.Set(context.Background(), nodeID)

	ltc.dbContext = &kv.DBContext{
		UserPriority: roachpb.NormalUserPriority,
		Stopper:      ltc.stopper,
		NodeID:       base.NewSQLIDContainerForNode(&nodeIDContainer),
	}
	ltc.DB = kv.NewDBWithContext(cfg.AmbientCtx, factory, ltc.Clock, *ltc.dbContext)
	transport := kvserver.NewDummyRaftTransport(cfg.Settings, cfg.AmbientCtx.Tracer)

	if ltc.StoreTestingKnobs == nil {
		__antithesis_instrumentation__.Notify(645531)
		cfg.TestingKnobs.DisableScanner = true
		cfg.TestingKnobs.DisableSplitQueue = true
	} else {
		__antithesis_instrumentation__.Notify(645532)
		cfg.TestingKnobs = *ltc.StoreTestingKnobs
	}
	__antithesis_instrumentation__.Notify(645520)
	cfg.DB = ltc.DB
	cfg.Gossip = ltc.Gossip
	cfg.HistogramWindowInterval = metric.TestSampleInterval
	active, renewal := cfg.NodeLivenessDurations()
	cfg.NodeLiveness = liveness.NewNodeLiveness(liveness.NodeLivenessOptions{
		AmbientCtx:              cfg.AmbientCtx,
		Clock:                   cfg.Clock,
		DB:                      cfg.DB,
		Gossip:                  cfg.Gossip,
		LivenessThreshold:       active,
		RenewalDuration:         renewal,
		Settings:                cfg.Settings,
		HistogramWindowInterval: cfg.HistogramWindowInterval,
	})
	kvserver.TimeUntilStoreDead.Override(ctx, &cfg.Settings.SV, kvserver.TestTimeUntilStoreDead)
	cfg.StorePool = kvserver.NewStorePool(
		cfg.AmbientCtx,
		cfg.Settings,
		cfg.Gossip,
		cfg.Clock,
		cfg.NodeLiveness.GetNodeCount,
		kvserver.MakeStorePoolNodeLivenessFunc(cfg.NodeLiveness),
		false,
	)
	cfg.Transport = transport
	cfg.ClosedTimestampReceiver = sidetransport.NewReceiver(nc, ltc.stopper, ltc.Stores, nil)

	if err := kvserver.WriteClusterVersion(ctx, ltc.Eng, clusterversion.TestingClusterVersion); err != nil {
		__antithesis_instrumentation__.Notify(645533)
		t.Fatalf("unable to write cluster version: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(645534)
	}
	__antithesis_instrumentation__.Notify(645521)
	if err := kvserver.InitEngine(
		ctx, ltc.Eng, roachpb.StoreIdent{NodeID: nodeID, StoreID: 1},
	); err != nil {
		__antithesis_instrumentation__.Notify(645535)
		t.Fatalf("unable to start local test cluster: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(645536)
	}
	__antithesis_instrumentation__.Notify(645522)

	rangeFeedFactory, err := rangefeed.NewFactory(ltc.stopper, ltc.DB, cfg.Settings, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(645537)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(645538)
	}
	__antithesis_instrumentation__.Notify(645523)
	cfg.SpanConfigSubscriber = spanconfigkvsubscriber.New(
		clock,
		rangeFeedFactory,
		keys.SpanConfigurationsTableID,
		1<<20,
		cfg.DefaultSpanConfig,
		nil,
	)
	cfg.SystemConfigProvider = systemconfigwatcher.New(
		keys.SystemSQLCodec,
		cfg.Clock,
		rangeFeedFactory,
		zonepb.DefaultZoneConfigRef(),
	)

	ltc.Store = kvserver.NewStore(ctx, cfg, ltc.Eng, nodeDesc)

	var initialValues []roachpb.KeyValue
	var splits []roachpb.RKey
	if !ltc.DontCreateSystemRanges {
		__antithesis_instrumentation__.Notify(645539)
		schema := bootstrap.MakeMetadataSchema(
			keys.SystemSQLCodec, zonepb.DefaultZoneConfigRef(), zonepb.DefaultSystemZoneConfigRef(),
		)
		var tableSplits []roachpb.RKey
		initialValues, tableSplits = schema.GetInitialValues()
		splits = append(config.StaticSplits(), tableSplits...)
		sort.Slice(splits, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(645540)
			return splits[i].Less(splits[j])
		})
	} else {
		__antithesis_instrumentation__.Notify(645541)
	}
	__antithesis_instrumentation__.Notify(645524)

	if err := kvserver.WriteInitialClusterData(
		ctx,
		ltc.Eng,
		initialValues,
		clusterversion.TestingBinaryVersion,
		1,
		splits,
		ltc.Clock.PhysicalNow(),
		cfg.TestingKnobs,
	); err != nil {
		__antithesis_instrumentation__.Notify(645542)
		t.Fatalf("unable to start local test cluster: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(645543)
	}
	__antithesis_instrumentation__.Notify(645525)

	nc.Set(ctx, nodeDesc.NodeID)
	if err := ltc.Gossip.SetNodeDescriptor(nodeDesc); err != nil {
		__antithesis_instrumentation__.Notify(645544)
		t.Fatalf("unable to set node descriptor: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(645545)
	}
	__antithesis_instrumentation__.Notify(645526)

	if !ltc.DisableLivenessHeartbeat {
		__antithesis_instrumentation__.Notify(645546)
		cfg.NodeLiveness.Start(ctx,
			liveness.NodeLivenessStartOptions{Stopper: ltc.stopper, Engines: []storage.Engine{ltc.Eng}})
	} else {
		__antithesis_instrumentation__.Notify(645547)
	}
	__antithesis_instrumentation__.Notify(645527)

	if err := ltc.Store.Start(ctx, ltc.stopper); err != nil {
		__antithesis_instrumentation__.Notify(645548)
		t.Fatalf("unable to start local test cluster: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(645549)
	}
	__antithesis_instrumentation__.Notify(645528)

	ltc.Stores.AddStore(ltc.Store)
	ltc.Cfg = cfg
}

func (ltc *LocalTestCluster) Stop() {
	__antithesis_instrumentation__.Notify(645550)

	if ltc.tester.Failed() {
		__antithesis_instrumentation__.Notify(645553)
		return
	} else {
		__antithesis_instrumentation__.Notify(645554)
	}
	__antithesis_instrumentation__.Notify(645551)
	if r := recover(); r != nil {
		__antithesis_instrumentation__.Notify(645555)
		panic(r)
	} else {
		__antithesis_instrumentation__.Notify(645556)
	}
	__antithesis_instrumentation__.Notify(645552)
	ltc.stopper.Stop(context.TODO())
}
