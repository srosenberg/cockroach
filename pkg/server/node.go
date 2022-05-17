package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvtenant"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/server/tenantsettingswatcher"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

const (
	gossipStatusInterval = 1 * time.Minute

	graphiteIntervalKey = "external.graphite.interval"
	maxGraphiteInterval = 15 * time.Minute
)

var (
	metaExecLatency = metric.Metadata{
		Name: "exec.latency",
		Help: `Latency of batch KV requests (including errors) executed on this node.

This measures requests already addressed to a single replica, from the moment
at which they arrive at the internal gRPC endpoint to the moment at which the
response (or an error) is returned.

This latency includes in particular commit waits, conflict resolution and replication,
and end-users can easily produce high measurements via long-running transactions that
conflict with foreground traffic. This metric thus does not provide a good signal for
understanding the health of the KV layer.
`,
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaExecSuccess = metric.Metadata{
		Name: "exec.success",
		Help: `Number of batch KV requests executed successfully on this node.

A request is considered to have executed 'successfully' if it either returns a result
or a transaction restart/abort error.
`,
		Measurement: "Batch KV Requests",
		Unit:        metric.Unit_COUNT,
	}
	metaExecError = metric.Metadata{
		Name: "exec.error",
		Help: `Number of batch KV requests that failed to execute on this node.

This count excludes transaction restart/abort errors. However, it will include
other errors expected during normal operation, such as ConditionFailedError.
This metric is thus not an indicator of KV health.`,
		Measurement: "Batch KV Requests",
		Unit:        metric.Unit_COUNT,
	}

	metaDiskStalls = metric.Metadata{
		Name:        "engine.stalls",
		Help:        "Number of disk stalls detected on this node",
		Measurement: "Disk stalls detected",
		Unit:        metric.Unit_COUNT,
	}

	metaInternalBatchRPCMethodCount = metric.Metadata{
		Name:        "rpc.method.%s.recv",
		Help:        "Number of %s requests processed",
		Measurement: "RPCs",
		Unit:        metric.Unit_COUNT,
	}

	metaInternalBatchRPCCount = metric.Metadata{
		Name:        "rpc.batches.recv",
		Help:        "Number of batches processed",
		Measurement: "Batches",
		Unit:        metric.Unit_COUNT,
	}
)

var (
	graphiteEndpoint = settings.RegisterStringSetting(
		settings.TenantWritable,
		"external.graphite.endpoint",
		"if nonempty, push server metrics to the Graphite or Carbon server at the specified host:port",
		"",
	).WithPublic()

	graphiteInterval = settings.RegisterDurationSetting(
		settings.TenantWritable,
		graphiteIntervalKey,
		"the interval at which metrics are pushed to Graphite (if enabled)",
		10*time.Second,
		settings.NonNegativeDurationWithMaximum(maxGraphiteInterval),
	).WithPublic()
)

type nodeMetrics struct {
	Latency    *metric.Histogram
	Success    *metric.Counter
	Err        *metric.Counter
	DiskStalls *metric.Counter

	BatchCount   *metric.Counter
	MethodCounts [roachpb.NumMethods]*metric.Counter
}

func makeNodeMetrics(reg *metric.Registry, histogramWindow time.Duration) nodeMetrics {
	__antithesis_instrumentation__.Notify(194341)
	nm := nodeMetrics{
		Latency:    metric.NewLatency(metaExecLatency, histogramWindow),
		Success:    metric.NewCounter(metaExecSuccess),
		Err:        metric.NewCounter(metaExecError),
		DiskStalls: metric.NewCounter(metaDiskStalls),
		BatchCount: metric.NewCounter(metaInternalBatchRPCCount),
	}

	for i := range nm.MethodCounts {
		__antithesis_instrumentation__.Notify(194343)
		method := roachpb.Method(i).String()
		meta := metaInternalBatchRPCMethodCount
		meta.Name = fmt.Sprintf(meta.Name, strings.ToLower(method))
		meta.Help = fmt.Sprintf(meta.Help, method)
		nm.MethodCounts[i] = metric.NewCounter(meta)
	}
	__antithesis_instrumentation__.Notify(194342)

	reg.AddMetricStruct(nm)
	return nm
}

func (nm nodeMetrics) callComplete(d time.Duration, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(194344)
	if pErr != nil && func() bool {
		__antithesis_instrumentation__.Notify(194346)
		return pErr.TransactionRestart() == roachpb.TransactionRestart_NONE == true
	}() == true {
		__antithesis_instrumentation__.Notify(194347)
		nm.Err.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(194348)
		nm.Success.Inc(1)
	}
	__antithesis_instrumentation__.Notify(194345)
	nm.Latency.RecordValue(d.Nanoseconds())
}

type Node struct {
	stopper      *stop.Stopper
	clusterID    *base.ClusterIDContainer
	Descriptor   roachpb.NodeDescriptor
	storeCfg     kvserver.StoreConfig
	sqlExec      *sql.InternalExecutor
	stores       *kvserver.Stores
	metrics      nodeMetrics
	recorder     *status.MetricsRecorder
	startedAt    int64
	lastUp       int64
	initialStart bool
	txnMetrics   kvcoord.TxnMetrics

	additionalStoreInitCh chan struct{}

	perReplicaServer kvserver.Server

	admissionController kvserver.KVAdmissionController

	tenantUsage multitenant.TenantUsageServer

	tenantSettingsWatcher *tenantsettingswatcher.Watcher

	spanConfigAccessor spanconfig.KVAccessor

	suppressNodeStatus syncutil.AtomicBool

	testingErrorEvent func(context.Context, *roachpb.BatchRequest, error)
}

var _ roachpb.InternalServer = &Node{}

func allocateNodeID(ctx context.Context, db *kv.DB) (roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(194349)
	val, err := kv.IncrementValRetryable(ctx, db, keys.NodeIDGenerator, 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(194351)
		return 0, errors.Wrap(err, "unable to allocate node ID")
	} else {
		__antithesis_instrumentation__.Notify(194352)
	}
	__antithesis_instrumentation__.Notify(194350)
	return roachpb.NodeID(val), nil
}

func allocateStoreIDs(
	ctx context.Context, nodeID roachpb.NodeID, count int64, db *kv.DB,
) (roachpb.StoreID, error) {
	__antithesis_instrumentation__.Notify(194353)
	val, err := kv.IncrementValRetryable(ctx, db, keys.StoreIDGenerator, count)
	if err != nil {
		__antithesis_instrumentation__.Notify(194355)
		return 0, errors.Wrapf(err, "unable to allocate %d store IDs for node %d", count, nodeID)
	} else {
		__antithesis_instrumentation__.Notify(194356)
	}
	__antithesis_instrumentation__.Notify(194354)
	return roachpb.StoreID(val - count + 1), nil
}

func GetBootstrapSchema(
	defaultZoneConfig *zonepb.ZoneConfig, defaultSystemZoneConfig *zonepb.ZoneConfig,
) bootstrap.MetadataSchema {
	__antithesis_instrumentation__.Notify(194357)
	return bootstrap.MakeMetadataSchema(keys.SystemSQLCodec, defaultZoneConfig, defaultSystemZoneConfig)
}

func bootstrapCluster(
	ctx context.Context, engines []storage.Engine, initCfg initServerCfg,
) (*initState, error) {
	__antithesis_instrumentation__.Notify(194358)
	clusterID := uuid.MakeV4()

	var bootstrapVersion clusterversion.ClusterVersion
	for i, eng := range engines {
		__antithesis_instrumentation__.Notify(194360)
		cv, err := kvserver.ReadClusterVersion(ctx, eng)
		if err != nil {
			__antithesis_instrumentation__.Notify(194364)
			return nil, errors.Wrapf(err, "reading cluster version of %s", eng)
		} else {
			__antithesis_instrumentation__.Notify(194365)
			if cv.Major == 0 {
				__antithesis_instrumentation__.Notify(194366)
				return nil, errors.Errorf("missing bootstrap version")
			} else {
				__antithesis_instrumentation__.Notify(194367)
			}
		}
		__antithesis_instrumentation__.Notify(194361)

		if i == 0 {
			__antithesis_instrumentation__.Notify(194368)
			bootstrapVersion = cv
		} else {
			__antithesis_instrumentation__.Notify(194369)
			if bootstrapVersion != cv {
				__antithesis_instrumentation__.Notify(194370)
				return nil, errors.Errorf("found cluster versions %s and %s", bootstrapVersion, cv)
			} else {
				__antithesis_instrumentation__.Notify(194371)
			}
		}
		__antithesis_instrumentation__.Notify(194362)

		sIdent := roachpb.StoreIdent{
			ClusterID: clusterID,
			NodeID:    kvserver.FirstNodeID,
			StoreID:   kvserver.FirstStoreID + roachpb.StoreID(i),
		}

		if err := kvserver.InitEngine(ctx, eng, sIdent); err != nil {
			__antithesis_instrumentation__.Notify(194372)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(194373)
		}
		__antithesis_instrumentation__.Notify(194363)

		if i == 0 {
			__antithesis_instrumentation__.Notify(194374)
			schema := GetBootstrapSchema(&initCfg.defaultZoneConfig, &initCfg.defaultSystemZoneConfig)
			initialValues, tableSplits := schema.GetInitialValues()
			splits := append(config.StaticSplits(), tableSplits...)
			sort.Slice(splits, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(194377)
				return splits[i].Less(splits[j])
			})
			__antithesis_instrumentation__.Notify(194375)

			var storeKnobs kvserver.StoreTestingKnobs
			if kn, ok := initCfg.testingKnobs.Store.(*kvserver.StoreTestingKnobs); ok {
				__antithesis_instrumentation__.Notify(194378)
				storeKnobs = *kn
			} else {
				__antithesis_instrumentation__.Notify(194379)
			}
			__antithesis_instrumentation__.Notify(194376)
			if err := kvserver.WriteInitialClusterData(
				ctx, eng, initialValues,
				bootstrapVersion.Version, len(engines), splits,
				hlc.UnixNano(), storeKnobs,
			); err != nil {
				__antithesis_instrumentation__.Notify(194380)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(194381)
			}
		} else {
			__antithesis_instrumentation__.Notify(194382)
		}
	}
	__antithesis_instrumentation__.Notify(194359)

	return inspectEngines(ctx, engines, initCfg.binaryVersion, initCfg.binaryMinSupportedVersion)
}

func NewNode(
	cfg kvserver.StoreConfig,
	recorder *status.MetricsRecorder,
	reg *metric.Registry,
	stopper *stop.Stopper,
	txnMetrics kvcoord.TxnMetrics,
	stores *kvserver.Stores,
	execCfg *sql.ExecutorConfig,
	clusterID *base.ClusterIDContainer,
	kvAdmissionQ *admission.WorkQueue,
	storeGrantCoords *admission.StoreGrantCoordinators,
	tenantUsage multitenant.TenantUsageServer,
	tenantSettingsWatcher *tenantsettingswatcher.Watcher,
	spanConfigAccessor spanconfig.KVAccessor,
) *Node {
	__antithesis_instrumentation__.Notify(194383)
	var sqlExec *sql.InternalExecutor
	if execCfg != nil {
		__antithesis_instrumentation__.Notify(194385)
		sqlExec = execCfg.InternalExecutor
	} else {
		__antithesis_instrumentation__.Notify(194386)
	}
	__antithesis_instrumentation__.Notify(194384)
	n := &Node{
		storeCfg:   cfg,
		stopper:    stopper,
		recorder:   recorder,
		metrics:    makeNodeMetrics(reg, cfg.HistogramWindowInterval),
		stores:     stores,
		txnMetrics: txnMetrics,
		sqlExec:    sqlExec,
		clusterID:  clusterID,
		admissionController: kvserver.MakeKVAdmissionController(
			kvAdmissionQ, storeGrantCoords, cfg.Settings),
		tenantUsage:           tenantUsage,
		tenantSettingsWatcher: tenantSettingsWatcher,
		spanConfigAccessor:    spanConfigAccessor,
		testingErrorEvent:     cfg.TestingKnobs.TestingResponseErrorEvent,
	}
	n.storeCfg.KVAdmissionController = n.admissionController
	n.perReplicaServer = kvserver.MakeServer(&n.Descriptor, n.stores)
	return n
}

func (n *Node) InitLogger(execCfg *sql.ExecutorConfig) {
	__antithesis_instrumentation__.Notify(194387)
	n.sqlExec = execCfg.InternalExecutor
}

func (n *Node) String() string {
	__antithesis_instrumentation__.Notify(194388)
	return fmt.Sprintf("node=%d", n.Descriptor.NodeID)
}

func (n *Node) AnnotateCtx(ctx context.Context) context.Context {
	__antithesis_instrumentation__.Notify(194389)
	return n.storeCfg.AmbientCtx.AnnotateCtx(ctx)
}

func (n *Node) AnnotateCtxWithSpan(
	ctx context.Context, opName string,
) (context.Context, *tracing.Span) {
	__antithesis_instrumentation__.Notify(194390)
	return n.storeCfg.AmbientCtx.AnnotateCtxWithSpan(ctx, opName)
}

func (n *Node) start(
	ctx context.Context,
	addr, sqlAddr, httpAddr net.Addr,
	state initState,
	initialStart bool,
	clusterName string,
	attrs roachpb.Attributes,
	locality roachpb.Locality,
	localityAddress []roachpb.LocalityAddress,
	nodeDescriptorCallback func(descriptor roachpb.NodeDescriptor),
) error {
	__antithesis_instrumentation__.Notify(194391)
	n.initialStart = initialStart
	n.startedAt = n.storeCfg.Clock.Now().WallTime
	n.Descriptor = roachpb.NodeDescriptor{
		NodeID:          state.nodeID,
		Address:         util.MakeUnresolvedAddr(addr.Network(), addr.String()),
		SQLAddress:      util.MakeUnresolvedAddr(sqlAddr.Network(), sqlAddr.String()),
		Attrs:           attrs,
		Locality:        locality,
		LocalityAddress: localityAddress,
		ClusterName:     clusterName,
		ServerVersion:   n.storeCfg.Settings.Version.BinaryVersion(),
		BuildTag:        build.GetInfo().Tag,
		StartedAt:       n.startedAt,
		HTTPAddress:     util.MakeUnresolvedAddr(httpAddr.Network(), httpAddr.String()),
	}

	if nodeDescriptorCallback != nil {
		__antithesis_instrumentation__.Notify(194400)
		nodeDescriptorCallback(n.Descriptor)
	} else {
		__antithesis_instrumentation__.Notify(194401)
	}
	__antithesis_instrumentation__.Notify(194392)

	n.storeCfg.Gossip.NodeID.Set(ctx, n.Descriptor.NodeID)
	if err := n.storeCfg.Gossip.SetNodeDescriptor(&n.Descriptor); err != nil {
		__antithesis_instrumentation__.Notify(194402)
		return errors.Wrapf(err, "couldn't gossip descriptor for node %d", n.Descriptor.NodeID)
	} else {
		__antithesis_instrumentation__.Notify(194403)
	}
	__antithesis_instrumentation__.Notify(194393)

	for _, e := range state.initializedEngines {
		__antithesis_instrumentation__.Notify(194404)
		s := kvserver.NewStore(ctx, n.storeCfg, e, &n.Descriptor)
		if err := s.Start(ctx, n.stopper); err != nil {
			__antithesis_instrumentation__.Notify(194406)
			return errors.Wrap(err, "failed to start store")
		} else {
			__antithesis_instrumentation__.Notify(194407)
		}
		__antithesis_instrumentation__.Notify(194405)

		n.addStore(ctx, s)
		log.Infof(ctx, "initialized store s%s", s.StoreID())
	}
	__antithesis_instrumentation__.Notify(194394)

	if err := n.validateStores(ctx); err != nil {
		__antithesis_instrumentation__.Notify(194408)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194409)
	}
	__antithesis_instrumentation__.Notify(194395)
	log.VEventf(ctx, 2, "validated stores")

	var mostRecentTimestamp hlc.Timestamp
	if err := n.stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194410)
		timestamp, err := s.ReadLastUpTimestamp(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(194413)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194414)
		}
		__antithesis_instrumentation__.Notify(194411)
		if mostRecentTimestamp.Less(timestamp) {
			__antithesis_instrumentation__.Notify(194415)
			mostRecentTimestamp = timestamp
		} else {
			__antithesis_instrumentation__.Notify(194416)
		}
		__antithesis_instrumentation__.Notify(194412)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(194417)
		return errors.Wrapf(err, "failed to read last up timestamp from stores")
	} else {
		__antithesis_instrumentation__.Notify(194418)
	}
	__antithesis_instrumentation__.Notify(194396)
	n.lastUp = mostRecentTimestamp.WallTime

	if err := n.storeCfg.Gossip.SetStorage(n.stores); err != nil {
		__antithesis_instrumentation__.Notify(194419)
		return errors.Wrap(err, "failed to initialize the gossip interface")
	} else {
		__antithesis_instrumentation__.Notify(194420)
	}
	__antithesis_instrumentation__.Notify(194397)

	if len(state.uninitializedEngines) > 0 {
		__antithesis_instrumentation__.Notify(194421)

		n.additionalStoreInitCh = make(chan struct{})
		if err := n.stopper.RunAsyncTask(ctx, "initialize-additional-stores", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(194422)
			if err := n.initializeAdditionalStores(ctx, state.uninitializedEngines, n.stopper); err != nil {
				__antithesis_instrumentation__.Notify(194424)
				log.Fatalf(ctx, "while initializing additional stores: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(194425)
			}
			__antithesis_instrumentation__.Notify(194423)
			close(n.additionalStoreInitCh)
		}); err != nil {
			__antithesis_instrumentation__.Notify(194426)
			close(n.additionalStoreInitCh)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194427)
		}
	} else {
		__antithesis_instrumentation__.Notify(194428)
	}
	__antithesis_instrumentation__.Notify(194398)

	n.startComputePeriodicMetrics(n.stopper, base.DefaultMetricsSampleInterval)

	n.admissionController.SetTenantWeightProvider(n, n.stopper)

	n.startGossiping(ctx, n.stopper)

	allEngines := append([]storage.Engine(nil), state.initializedEngines...)
	allEngines = append(allEngines, state.uninitializedEngines...)
	for _, e := range allEngines {
		__antithesis_instrumentation__.Notify(194429)
		t := e.Type()
		log.Infof(ctx, "started with engine type %v", t)
	}
	__antithesis_instrumentation__.Notify(194399)
	log.Infof(ctx, "started with attributes %v", attrs.Attrs)
	return nil
}

func (n *Node) waitForAdditionalStoreInit() {
	__antithesis_instrumentation__.Notify(194430)
	if n.additionalStoreInitCh != nil {
		__antithesis_instrumentation__.Notify(194431)
		<-n.additionalStoreInitCh
	} else {
		__antithesis_instrumentation__.Notify(194432)
	}
}

func (n *Node) IsDraining() bool {
	__antithesis_instrumentation__.Notify(194433)
	var isDraining bool
	if err := n.stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194435)
		isDraining = isDraining || func() bool {
			__antithesis_instrumentation__.Notify(194436)
			return s.IsDraining() == true
		}() == true
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(194437)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(194438)
	}
	__antithesis_instrumentation__.Notify(194434)
	return isDraining
}

func (n *Node) SetDraining(drain bool, reporter func(int, redact.SafeString), verbose bool) error {
	__antithesis_instrumentation__.Notify(194439)
	return n.stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194440)
		s.SetDraining(drain, reporter, verbose)
		return nil
	})
}

func (n *Node) SetHLCUpperBound(ctx context.Context, hlcUpperBound int64) error {
	__antithesis_instrumentation__.Notify(194441)
	return n.stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194442)
		return s.WriteHLCUpperBound(ctx, hlcUpperBound)
	})
}

func (n *Node) addStore(ctx context.Context, store *kvserver.Store) {
	__antithesis_instrumentation__.Notify(194443)
	cv, err := kvserver.ReadClusterVersion(context.TODO(), store.Engine())
	if err != nil {
		__antithesis_instrumentation__.Notify(194446)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(194447)
	}
	__antithesis_instrumentation__.Notify(194444)
	if cv == (clusterversion.ClusterVersion{}) {
		__antithesis_instrumentation__.Notify(194448)

		log.Fatal(ctx, "attempting to add a store without a version")
	} else {
		__antithesis_instrumentation__.Notify(194449)
	}
	__antithesis_instrumentation__.Notify(194445)
	n.stores.AddStore(store)
	n.recorder.AddStore(store)
}

func (n *Node) validateStores(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(194450)
	return n.stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194451)
		if n.Descriptor.NodeID != s.Ident.NodeID {
			__antithesis_instrumentation__.Notify(194453)
			return errors.Errorf("store %s node ID doesn't match node ID: %d", s, n.Descriptor.NodeID)
		} else {
			__antithesis_instrumentation__.Notify(194454)
		}
		__antithesis_instrumentation__.Notify(194452)
		return nil
	})
}

func (n *Node) initializeAdditionalStores(
	ctx context.Context, engines []storage.Engine, stopper *stop.Stopper,
) error {
	__antithesis_instrumentation__.Notify(194455)
	if n.clusterID.Get() == uuid.Nil {
		__antithesis_instrumentation__.Notify(194458)
		return errors.New("missing cluster ID during initialization of additional store")
	} else {
		__antithesis_instrumentation__.Notify(194459)
	}

	{
		__antithesis_instrumentation__.Notify(194460)

		storeIDAlloc := int64(len(engines))
		startID, err := allocateStoreIDs(ctx, n.Descriptor.NodeID, storeIDAlloc, n.storeCfg.DB)
		if err != nil {
			__antithesis_instrumentation__.Notify(194462)
			return errors.Wrap(err, "error allocating store ids")
		} else {
			__antithesis_instrumentation__.Notify(194463)
		}
		__antithesis_instrumentation__.Notify(194461)

		sIdent := roachpb.StoreIdent{
			ClusterID: n.clusterID.Get(),
			NodeID:    n.Descriptor.NodeID,
			StoreID:   startID,
		}
		for _, eng := range engines {
			__antithesis_instrumentation__.Notify(194464)
			if err := kvserver.InitEngine(ctx, eng, sIdent); err != nil {
				__antithesis_instrumentation__.Notify(194468)
				return err
			} else {
				__antithesis_instrumentation__.Notify(194469)
			}
			__antithesis_instrumentation__.Notify(194465)

			s := kvserver.NewStore(ctx, n.storeCfg, eng, &n.Descriptor)
			if err := s.Start(ctx, stopper); err != nil {
				__antithesis_instrumentation__.Notify(194470)
				return err
			} else {
				__antithesis_instrumentation__.Notify(194471)
			}
			__antithesis_instrumentation__.Notify(194466)

			n.addStore(ctx, s)
			log.Infof(ctx, "initialized store s%s", s.StoreID())

			if err := s.GossipStore(ctx, false); err != nil {
				__antithesis_instrumentation__.Notify(194472)
				log.Warningf(ctx, "error doing initial gossiping: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(194473)
			}
			__antithesis_instrumentation__.Notify(194467)

			sIdent.StoreID++
		}
	}
	__antithesis_instrumentation__.Notify(194456)

	if err := n.writeNodeStatus(ctx, 0, false); err != nil {
		__antithesis_instrumentation__.Notify(194474)
		log.Warningf(ctx, "error writing node summary after store bootstrap: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(194475)
	}
	__antithesis_instrumentation__.Notify(194457)

	return nil
}

func (n *Node) startGossiping(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(194476)
	ctx = n.AnnotateCtx(ctx)
	_ = stopper.RunAsyncTask(ctx, "start-gossip", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(194477)

		if _, err := n.storeCfg.Gossip.GetNodeDescriptor(n.Descriptor.NodeID); err != nil {
			__antithesis_instrumentation__.Notify(194480)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(194481)
		}
		__antithesis_instrumentation__.Notify(194478)

		statusTicker := time.NewTicker(gossipStatusInterval)
		storesTicker := time.NewTicker(gossip.StoresInterval)
		nodeTicker := time.NewTicker(gossip.NodeDescriptorInterval)
		defer func() {
			__antithesis_instrumentation__.Notify(194482)
			nodeTicker.Stop()
			storesTicker.Stop()
			statusTicker.Stop()
		}()
		__antithesis_instrumentation__.Notify(194479)

		n.gossipStores(ctx)
		for {
			__antithesis_instrumentation__.Notify(194483)
			select {
			case <-statusTicker.C:
				__antithesis_instrumentation__.Notify(194484)
				n.storeCfg.Gossip.LogStatus()
			case <-storesTicker.C:
				__antithesis_instrumentation__.Notify(194485)
				n.gossipStores(ctx)
			case <-nodeTicker.C:
				__antithesis_instrumentation__.Notify(194486)
				if err := n.storeCfg.Gossip.SetNodeDescriptor(&n.Descriptor); err != nil {
					__antithesis_instrumentation__.Notify(194488)
					log.Warningf(ctx, "couldn't gossip descriptor for node %d: %s", n.Descriptor.NodeID, err)
				} else {
					__antithesis_instrumentation__.Notify(194489)
				}
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(194487)
				return
			}
		}
	})
}

func (n *Node) gossipStores(ctx context.Context) {
	__antithesis_instrumentation__.Notify(194490)
	if err := n.stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194491)
		return s.GossipStore(ctx, false)
	}); err != nil {
		__antithesis_instrumentation__.Notify(194492)
		log.Warningf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(194493)
	}
}

func (n *Node) startComputePeriodicMetrics(stopper *stop.Stopper, interval time.Duration) {
	__antithesis_instrumentation__.Notify(194494)
	ctx := n.AnnotateCtx(context.Background())
	_ = stopper.RunAsyncTask(ctx, "compute-metrics", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(194495)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for tick := 0; ; tick++ {
			__antithesis_instrumentation__.Notify(194496)
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(194497)
				if err := n.computePeriodicMetrics(ctx, tick); err != nil {
					__antithesis_instrumentation__.Notify(194499)
					log.Errorf(ctx, "failed computing periodic metrics: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(194500)
				}
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(194498)
				return
			}
		}
	})
}

func (n *Node) computePeriodicMetrics(ctx context.Context, tick int) error {
	__antithesis_instrumentation__.Notify(194501)
	return n.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194502)
		if err := store.ComputeMetrics(ctx, tick); err != nil {
			__antithesis_instrumentation__.Notify(194504)
			log.Warningf(ctx, "%s: unable to compute metrics: %s", store, err)
		} else {
			__antithesis_instrumentation__.Notify(194505)
		}
		__antithesis_instrumentation__.Notify(194503)
		return nil
	})
}

func (n *Node) GetPebbleMetrics() []admission.StoreMetrics {
	__antithesis_instrumentation__.Notify(194506)
	var metrics []admission.StoreMetrics
	_ = n.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194508)
		m := store.Engine().GetMetrics()
		metrics = append(
			metrics, admission.StoreMetrics{StoreID: int32(store.StoreID()), Metrics: m.Metrics})
		return nil
	})
	__antithesis_instrumentation__.Notify(194507)
	return metrics
}

func (n *Node) GetTenantWeights() kvserver.TenantWeights {
	__antithesis_instrumentation__.Notify(194509)
	weights := kvserver.TenantWeights{
		Node: make(map[uint64]uint32),
	}
	_ = n.stores.VisitStores(func(store *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194511)
		sw := make(map[uint64]uint32)
		weights.Stores = append(weights.Stores, kvserver.TenantWeightsForStore{
			StoreID: store.StoreID(),
			Weights: sw,
		})
		store.VisitReplicas(func(r *kvserver.Replica) bool {
			__antithesis_instrumentation__.Notify(194513)
			tid, valid := r.TenantID()
			if valid {
				__antithesis_instrumentation__.Notify(194515)
				weights.Node[tid.ToUint64()]++
				sw[tid.ToUint64()]++
			} else {
				__antithesis_instrumentation__.Notify(194516)
			}
			__antithesis_instrumentation__.Notify(194514)
			return true
		})
		__antithesis_instrumentation__.Notify(194512)
		return nil
	})
	__antithesis_instrumentation__.Notify(194510)
	return weights
}

func (n *Node) startGraphiteStatsExporter(st *cluster.Settings) {
	__antithesis_instrumentation__.Notify(194517)
	ctx := logtags.AddTag(n.AnnotateCtx(context.Background()), "graphite stats exporter", nil)
	pm := metric.MakePrometheusExporter()

	_ = n.stopper.RunAsyncTask(ctx, "graphite-exporter", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(194518)
		var timer timeutil.Timer
		defer timer.Stop()
		for {
			__antithesis_instrumentation__.Notify(194519)
			timer.Reset(graphiteInterval.Get(&st.SV))
			select {
			case <-n.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(194520)
				return
			case <-timer.C:
				__antithesis_instrumentation__.Notify(194521)
				timer.Read = true
				endpoint := graphiteEndpoint.Get(&st.SV)
				if endpoint != "" {
					__antithesis_instrumentation__.Notify(194522)
					if err := n.recorder.ExportToGraphite(ctx, endpoint, &pm); err != nil {
						__antithesis_instrumentation__.Notify(194523)
						log.Infof(ctx, "error pushing metrics to graphite: %s\n", err)
					} else {
						__antithesis_instrumentation__.Notify(194524)
					}
				} else {
					__antithesis_instrumentation__.Notify(194525)
				}
			}
		}
	})
}

func (n *Node) startWriteNodeStatus(frequency time.Duration) error {
	__antithesis_instrumentation__.Notify(194526)
	ctx := logtags.AddTag(n.AnnotateCtx(context.Background()), "summaries", nil)

	if err := n.writeNodeStatus(ctx, 0, false); err != nil {
		__antithesis_instrumentation__.Notify(194528)
		return errors.Wrap(err, "error recording initial status summaries")
	} else {
		__antithesis_instrumentation__.Notify(194529)
	}
	__antithesis_instrumentation__.Notify(194527)
	return n.stopper.RunAsyncTask(ctx, "write-node-status", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(194530)

		ticker := time.NewTicker(frequency)
		defer ticker.Stop()
		for {
			__antithesis_instrumentation__.Notify(194531)
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(194532)

				if err := n.writeNodeStatus(ctx, 2*frequency, true); err != nil {
					__antithesis_instrumentation__.Notify(194534)
					log.Warningf(ctx, "error recording status summaries: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(194535)
				}
			case <-n.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(194533)
				return
			}
		}
	})
}

func (n *Node) writeNodeStatus(ctx context.Context, alertTTL time.Duration, mustExist bool) error {
	__antithesis_instrumentation__.Notify(194536)
	if n.suppressNodeStatus.Get() {
		__antithesis_instrumentation__.Notify(194539)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(194540)
	}
	__antithesis_instrumentation__.Notify(194537)
	var err error
	if runErr := n.stopper.RunTask(ctx, "node.Node: writing summary", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(194541)
		nodeStatus := n.recorder.GenerateNodeStatus(ctx)
		if nodeStatus == nil {
			__antithesis_instrumentation__.Notify(194544)
			return
		} else {
			__antithesis_instrumentation__.Notify(194545)
		}
		__antithesis_instrumentation__.Notify(194542)

		if result := n.recorder.CheckHealth(ctx, *nodeStatus); len(result.Alerts) != 0 {
			__antithesis_instrumentation__.Notify(194546)
			var numNodes int
			if err := n.storeCfg.Gossip.IterateInfos(gossip.KeyNodeIDPrefix, func(k string, info gossip.Info) error {
				__antithesis_instrumentation__.Notify(194549)
				numNodes++
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(194550)
				log.Warningf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(194551)
			}
			__antithesis_instrumentation__.Notify(194547)
			if numNodes > 1 {
				__antithesis_instrumentation__.Notify(194552)

				log.Warningf(ctx, "health alerts detected: %+v", result)
			} else {
				__antithesis_instrumentation__.Notify(194553)
			}
			__antithesis_instrumentation__.Notify(194548)
			if err := n.storeCfg.Gossip.AddInfoProto(
				gossip.MakeNodeHealthAlertKey(n.Descriptor.NodeID), &result, alertTTL,
			); err != nil {
				__antithesis_instrumentation__.Notify(194554)
				log.Warningf(ctx, "unable to gossip health alerts: %+v", result)
			} else {
				__antithesis_instrumentation__.Notify(194555)
			}

		} else {
			__antithesis_instrumentation__.Notify(194556)
		}
		__antithesis_instrumentation__.Notify(194543)

		err = n.recorder.WriteNodeStatus(ctx, n.storeCfg.DB, *nodeStatus, mustExist)
	}); runErr != nil {
		__antithesis_instrumentation__.Notify(194557)
		err = runErr
	} else {
		__antithesis_instrumentation__.Notify(194558)
	}
	__antithesis_instrumentation__.Notify(194538)
	return err
}

func (n *Node) recordJoinEvent(ctx context.Context) {
	__antithesis_instrumentation__.Notify(194559)
	var event eventpb.EventPayload
	var nodeDetails *eventpb.CommonNodeEventDetails
	if !n.initialStart {
		__antithesis_instrumentation__.Notify(194562)
		ev := &eventpb.NodeRestart{}
		event = ev
		nodeDetails = &ev.CommonNodeEventDetails
		nodeDetails.LastUp = n.lastUp
	} else {
		__antithesis_instrumentation__.Notify(194563)
		ev := &eventpb.NodeJoin{}
		event = ev
		nodeDetails = &ev.CommonNodeEventDetails
		nodeDetails.LastUp = n.startedAt
	}
	__antithesis_instrumentation__.Notify(194560)
	event.CommonDetails().Timestamp = timeutil.Now().UnixNano()
	nodeDetails.StartedAt = n.startedAt
	nodeDetails.NodeID = int32(n.Descriptor.NodeID)

	log.StructuredEvent(ctx, event)

	if !n.storeCfg.LogRangeEvents {
		__antithesis_instrumentation__.Notify(194564)
		return
	} else {
		__antithesis_instrumentation__.Notify(194565)
	}
	__antithesis_instrumentation__.Notify(194561)

	_ = n.stopper.RunAsyncTask(ctx, "record-join", func(bgCtx context.Context) {
		__antithesis_instrumentation__.Notify(194566)
		ctx, span := n.AnnotateCtxWithSpan(bgCtx, "record-join-event")
		defer span.Finish()
		retryOpts := base.DefaultRetryOptions()
		retryOpts.Closer = n.stopper.ShouldQuiesce()
		for r := retry.Start(retryOpts); r.Next(); {
			__antithesis_instrumentation__.Notify(194567)
			if err := n.storeCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(194568)
				return sql.InsertEventRecord(ctx, n.sqlExec,
					txn,
					int32(n.Descriptor.NodeID),
					sql.LogToSystemTable|sql.LogToDevChannelIfVerbose,
					int32(n.Descriptor.NodeID),
					event,
				)
			}); err != nil {
				__antithesis_instrumentation__.Notify(194569)
				log.Warningf(ctx, "%s: unable to log event %v: %v", n, event, err)
			} else {
				__antithesis_instrumentation__.Notify(194570)
				return
			}
		}
	})
}

func checkNoUnknownRequest(reqs []roachpb.RequestUnion) *roachpb.UnsupportedRequestError {
	__antithesis_instrumentation__.Notify(194571)
	for _, req := range reqs {
		__antithesis_instrumentation__.Notify(194573)
		if req.GetValue() == nil {
			__antithesis_instrumentation__.Notify(194574)
			return &roachpb.UnsupportedRequestError{}
		} else {
			__antithesis_instrumentation__.Notify(194575)
		}
	}
	__antithesis_instrumentation__.Notify(194572)
	return nil
}

func (n *Node) batchInternal(
	ctx context.Context, tenID roachpb.TenantID, args *roachpb.BatchRequest,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(194576)
	if detail := checkNoUnknownRequest(args.Requests); detail != nil {
		__antithesis_instrumentation__.Notify(194579)
		var br roachpb.BatchResponse
		br.Error = roachpb.NewError(detail)
		return &br, nil
	} else {
		__antithesis_instrumentation__.Notify(194580)
	}
	__antithesis_instrumentation__.Notify(194577)

	var br *roachpb.BatchResponse
	if err := n.stopper.RunTaskWithErr(ctx, "node.Node: batch", func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(194581)
		var reqSp spanForRequest

		ctx, reqSp = n.setupSpanForIncomingRPC(ctx, tenID, args)

		defer func() { __antithesis_instrumentation__.Notify(194587); reqSp.finish(ctx, br) }()
		__antithesis_instrumentation__.Notify(194582)
		if log.HasSpanOrEvent(ctx) {
			__antithesis_instrumentation__.Notify(194588)
			log.Eventf(ctx, "node received request: %s", args.Summary())
		} else {
			__antithesis_instrumentation__.Notify(194589)
		}
		__antithesis_instrumentation__.Notify(194583)

		tStart := timeutil.Now()
		handle, err := n.admissionController.AdmitKVWork(ctx, tenID, args)
		defer n.admissionController.AdmittedKVWorkDone(handle)
		if err != nil {
			__antithesis_instrumentation__.Notify(194590)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194591)
		}
		__antithesis_instrumentation__.Notify(194584)
		var pErr *roachpb.Error
		br, pErr = n.stores.Send(ctx, *args)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(194592)
			br = &roachpb.BatchResponse{}
			log.VErrEventf(ctx, 3, "error from stores.Send: %s", pErr)
		} else {
			__antithesis_instrumentation__.Notify(194593)
		}
		__antithesis_instrumentation__.Notify(194585)
		if br.Error != nil {
			__antithesis_instrumentation__.Notify(194594)
			panic(roachpb.ErrorUnexpectedlySet(n.stores, br))
		} else {
			__antithesis_instrumentation__.Notify(194595)
		}
		__antithesis_instrumentation__.Notify(194586)
		n.metrics.callComplete(timeutil.Since(tStart), pErr)
		br.Error = pErr
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(194596)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194597)
	}
	__antithesis_instrumentation__.Notify(194578)
	return br, nil
}

func (n *Node) incrementBatchCounters(ba *roachpb.BatchRequest) {
	__antithesis_instrumentation__.Notify(194598)
	n.metrics.BatchCount.Inc(1)
	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(194599)
		m := ru.GetInner().Method()
		n.metrics.MethodCounts[m].Inc(1)
	}
}

func (n *Node) Batch(
	ctx context.Context, args *roachpb.BatchRequest,
) (*roachpb.BatchResponse, error) {
	__antithesis_instrumentation__.Notify(194600)
	n.incrementBatchCounters(args)

	ctx = n.storeCfg.AmbientCtx.ResetAndAnnotateCtx(ctx)

	tenantID, ok := roachpb.TenantFromContext(ctx)
	if !ok {
		__antithesis_instrumentation__.Notify(194605)
		tenantID = roachpb.SystemTenantID
	} else {
		__antithesis_instrumentation__.Notify(194606)
	}
	__antithesis_instrumentation__.Notify(194601)

	if args.GatewayNodeID == 0 && func() bool {
		__antithesis_instrumentation__.Notify(194607)
		return tenantID != roachpb.SystemTenantID == true
	}() == true {
		__antithesis_instrumentation__.Notify(194608)
		args.GatewayNodeID = n.Descriptor.NodeID
	} else {
		__antithesis_instrumentation__.Notify(194609)
	}
	__antithesis_instrumentation__.Notify(194602)

	br, err := n.batchInternal(ctx, tenantID, args)

	if err != nil {
		__antithesis_instrumentation__.Notify(194610)
		if br == nil {
			__antithesis_instrumentation__.Notify(194613)
			br = &roachpb.BatchResponse{}
		} else {
			__antithesis_instrumentation__.Notify(194614)
		}
		__antithesis_instrumentation__.Notify(194611)
		if br.Error != nil {
			__antithesis_instrumentation__.Notify(194615)
			log.Fatalf(
				ctx, "attempting to return both a plain error (%s) and roachpb.Error (%s)", err, br.Error,
			)
		} else {
			__antithesis_instrumentation__.Notify(194616)
		}
		__antithesis_instrumentation__.Notify(194612)
		br.Error = roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(194617)
	}
	__antithesis_instrumentation__.Notify(194603)
	if buildutil.CrdbTestBuild && func() bool {
		__antithesis_instrumentation__.Notify(194618)
		return br.Error != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(194619)
		return n.testingErrorEvent != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(194620)
		n.testingErrorEvent(ctx, args, errors.DecodeError(ctx, br.Error.EncodedError))
	} else {
		__antithesis_instrumentation__.Notify(194621)
	}
	__antithesis_instrumentation__.Notify(194604)
	return br, nil
}

type spanForRequest struct {
	sp            *tracing.Span
	needRecording bool
	tenID         roachpb.TenantID
}

func (sp *spanForRequest) finish(ctx context.Context, br *roachpb.BatchResponse) {
	__antithesis_instrumentation__.Notify(194622)
	var rec tracing.Recording

	sp.needRecording = sp.needRecording && func() bool {
		__antithesis_instrumentation__.Notify(194624)
		return br != nil == true
	}() == true

	if !sp.needRecording {
		__antithesis_instrumentation__.Notify(194625)
		sp.sp.Finish()
		return
	} else {
		__antithesis_instrumentation__.Notify(194626)
	}
	__antithesis_instrumentation__.Notify(194623)

	rec = sp.sp.FinishAndGetConfiguredRecording()
	if rec != nil {
		__antithesis_instrumentation__.Notify(194627)

		needRedaction := sp.tenID != roachpb.SystemTenantID
		if needRedaction {
			__antithesis_instrumentation__.Notify(194629)
			if err := redactRecordingForTenant(sp.tenID, rec); err != nil {
				__antithesis_instrumentation__.Notify(194630)
				log.Errorf(ctx, "error redacting trace recording: %s", err)
				rec = nil
			} else {
				__antithesis_instrumentation__.Notify(194631)
			}
		} else {
			__antithesis_instrumentation__.Notify(194632)
		}
		__antithesis_instrumentation__.Notify(194628)
		br.CollectedSpans = append(br.CollectedSpans, rec...)
	} else {
		__antithesis_instrumentation__.Notify(194633)
	}
}

func (n *Node) setupSpanForIncomingRPC(
	ctx context.Context, tenID roachpb.TenantID, ba *roachpb.BatchRequest,
) (context.Context, spanForRequest) {
	__antithesis_instrumentation__.Notify(194634)
	return setupSpanForIncomingRPC(ctx, tenID, ba, n.storeCfg.AmbientCtx.Tracer)
}

func setupSpanForIncomingRPC(
	ctx context.Context, tenID roachpb.TenantID, ba *roachpb.BatchRequest, tr *tracing.Tracer,
) (context.Context, spanForRequest) {
	__antithesis_instrumentation__.Notify(194635)
	var newSpan *tracing.Span
	parentSpan := tracing.SpanFromContext(ctx)
	localRequest := grpcutil.IsLocalRequestContext(ctx)

	needRecordingCollection := !localRequest
	if localRequest {
		__antithesis_instrumentation__.Notify(194637)

		ctx, newSpan = tracing.EnsureChildSpan(ctx, tr, tracing.BatchMethodName, tracing.WithServerSpanKind)
	} else {
		__antithesis_instrumentation__.Notify(194638)
		if parentSpan == nil {
			__antithesis_instrumentation__.Notify(194639)

			var remoteParent tracing.SpanMeta
			if !ba.TraceInfo.Empty() {
				__antithesis_instrumentation__.Notify(194640)
				ctx, newSpan = tr.StartSpanCtx(ctx, tracing.BatchMethodName,
					tracing.WithRemoteParentFromTraceInfo(&ba.TraceInfo),
					tracing.WithServerSpanKind)
			} else {
				__antithesis_instrumentation__.Notify(194641)

				var err error
				remoteParent, err = tracing.ExtractSpanMetaFromGRPCCtx(ctx, tr)
				if err != nil {
					__antithesis_instrumentation__.Notify(194643)
					log.Warningf(ctx, "error extracting tracing info from gRPC: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(194644)
				}
				__antithesis_instrumentation__.Notify(194642)
				ctx, newSpan = tr.StartSpanCtx(ctx, tracing.BatchMethodName,
					tracing.WithRemoteParentFromSpanMeta(remoteParent),
					tracing.WithServerSpanKind)
			}
		} else {
			__antithesis_instrumentation__.Notify(194645)

			ctx, newSpan = tr.StartSpanCtx(ctx, tracing.BatchMethodName,
				tracing.WithParent(parentSpan),
				tracing.WithServerSpanKind)
		}
	}
	__antithesis_instrumentation__.Notify(194636)

	return ctx, spanForRequest{
		needRecording: needRecordingCollection,
		tenID:         tenID,
		sp:            newSpan,
	}
}

func (n *Node) RangeLookup(
	ctx context.Context, req *roachpb.RangeLookupRequest,
) (*roachpb.RangeLookupResponse, error) {
	__antithesis_instrumentation__.Notify(194646)
	ctx = n.storeCfg.AmbientCtx.AnnotateCtx(ctx)

	sender := n.storeCfg.DB.NonTransactionalSender()
	rs, preRs, err := kv.RangeLookup(
		ctx,
		sender,
		req.Key.AsRawKey(),
		req.ReadConsistency,
		req.PrefetchNum,
		req.PrefetchReverse,
	)
	resp := new(roachpb.RangeLookupResponse)
	if err != nil {
		__antithesis_instrumentation__.Notify(194648)
		resp.Error = roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(194649)
		resp.Descriptors = rs
		resp.PrefetchedDescriptors = preRs
	}
	__antithesis_instrumentation__.Notify(194647)
	return resp, nil
}

func (n *Node) RangeFeed(
	args *roachpb.RangeFeedRequest, stream roachpb.Internal_RangeFeedServer,
) error {
	__antithesis_instrumentation__.Notify(194650)
	pErr := n.stores.RangeFeed(args, stream)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(194652)
		var event roachpb.RangeFeedEvent
		event.SetValue(&roachpb.RangeFeedError{
			Error: *pErr,
		})
		return stream.Send(&event)
	} else {
		__antithesis_instrumentation__.Notify(194653)
	}
	__antithesis_instrumentation__.Notify(194651)
	return nil
}

func (n *Node) ResetQuorum(
	ctx context.Context, req *roachpb.ResetQuorumRequest,
) (_ *roachpb.ResetQuorumResponse, rErr error) {
	__antithesis_instrumentation__.Notify(194654)

	var desc roachpb.RangeDescriptor
	var expValue roachpb.Value
	txnTries := 0
	if err := n.storeCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(194665)
		txnTries++
		if txnTries > 1 {
			__antithesis_instrumentation__.Notify(194669)
			log.Infof(ctx, "failed to retrieve range descriptor for r%d, retrying...", req.RangeID)
		} else {
			__antithesis_instrumentation__.Notify(194670)
		}
		__antithesis_instrumentation__.Notify(194666)
		kvs, err := kvclient.ScanMetaKVs(ctx, txn, roachpb.Span{
			Key:    roachpb.KeyMin,
			EndKey: roachpb.KeyMax,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(194671)
			return err
		} else {
			__antithesis_instrumentation__.Notify(194672)
		}
		__antithesis_instrumentation__.Notify(194667)

		for i := range kvs {
			__antithesis_instrumentation__.Notify(194673)
			if err := kvs[i].Value.GetProto(&desc); err != nil {
				__antithesis_instrumentation__.Notify(194675)
				return err
			} else {
				__antithesis_instrumentation__.Notify(194676)
			}
			__antithesis_instrumentation__.Notify(194674)
			if desc.RangeID == roachpb.RangeID(req.RangeID) {
				__antithesis_instrumentation__.Notify(194677)
				expValue = *kvs[i].Value
				return nil
			} else {
				__antithesis_instrumentation__.Notify(194678)
			}
		}
		__antithesis_instrumentation__.Notify(194668)
		return errors.Errorf("r%d not found", req.RangeID)
	}); err != nil {
		__antithesis_instrumentation__.Notify(194679)
		log.Errorf(ctx, "range descriptor for r%d could not be read: %v", req.RangeID, err)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194680)
	}
	__antithesis_instrumentation__.Notify(194655)
	log.Infof(ctx, "retrieved original range descriptor %s", desc)

	livenessMap := n.storeCfg.NodeLiveness.GetIsLiveMap()
	available := desc.Replicas().CanMakeProgress(func(rDesc roachpb.ReplicaDescriptor) bool {
		__antithesis_instrumentation__.Notify(194681)
		return livenessMap[rDesc.NodeID].IsLive
	})
	__antithesis_instrumentation__.Notify(194656)
	if available {
		__antithesis_instrumentation__.Notify(194682)
		return nil, errors.Errorf("targeted range to recover has not lost quorum.")
	} else {
		__antithesis_instrumentation__.Notify(194683)
	}
	__antithesis_instrumentation__.Notify(194657)

	if bytes.HasPrefix(desc.StartKey, keys.Meta1Prefix) || func() bool {
		__antithesis_instrumentation__.Notify(194684)
		return bytes.HasPrefix(desc.StartKey, keys.Meta2Prefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(194685)
		return nil, errors.Errorf("targeted range to recover is a meta1 or meta2 range.")
	} else {
		__antithesis_instrumentation__.Notify(194686)
	}
	__antithesis_instrumentation__.Notify(194658)

	deadReplicas := append([]roachpb.ReplicaDescriptor(nil), desc.Replicas().Descriptors()...)
	for _, rd := range deadReplicas {
		__antithesis_instrumentation__.Notify(194687)
		desc.RemoveReplica(rd.NodeID, rd.StoreID)
	}
	__antithesis_instrumentation__.Notify(194659)

	var storeID roachpb.StoreID
	if err := n.stores.VisitStores(func(s *kvserver.Store) error {
		__antithesis_instrumentation__.Notify(194688)
		if storeID == 0 {
			__antithesis_instrumentation__.Notify(194690)
			storeID = s.StoreID()
		} else {
			__antithesis_instrumentation__.Notify(194691)
		}
		__antithesis_instrumentation__.Notify(194689)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(194692)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194693)
	}
	__antithesis_instrumentation__.Notify(194660)
	if storeID == 0 {
		__antithesis_instrumentation__.Notify(194694)
		return nil, errors.New("no store found")
	} else {
		__antithesis_instrumentation__.Notify(194695)
	}
	__antithesis_instrumentation__.Notify(194661)

	toReplicaDescriptor := desc.AddReplica(n.Descriptor.NodeID, storeID, roachpb.VOTER_FULL)

	desc.IncrementGeneration()

	log.Infof(ctx, "initiating recovery process using %s", desc)

	metaKey := keys.RangeMetaKey(desc.EndKey).AsRawKey()
	if err := n.storeCfg.DB.CPut(ctx, metaKey, &desc, expValue.TagAndDataBytes()); err != nil {
		__antithesis_instrumentation__.Notify(194696)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194697)
	}
	__antithesis_instrumentation__.Notify(194662)
	log.Infof(ctx, "updated meta2 entry for r%d", desc.RangeID)

	conn, err := n.storeCfg.NodeDialer.Dial(ctx, n.Descriptor.NodeID, rpc.SystemClass)
	if err != nil {
		__antithesis_instrumentation__.Notify(194698)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194699)
	}
	__antithesis_instrumentation__.Notify(194663)

	if err := kvserver.SendEmptySnapshot(
		ctx,
		n.storeCfg.Settings,
		conn,
		n.storeCfg.Clock.Now(),
		desc,
		toReplicaDescriptor,
	); err != nil {
		__antithesis_instrumentation__.Notify(194700)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194701)
	}
	__antithesis_instrumentation__.Notify(194664)
	log.Infof(ctx, "sent empty snapshot to %s", toReplicaDescriptor)

	return &roachpb.ResetQuorumResponse{}, nil
}

func (n *Node) GossipSubscription(
	args *roachpb.GossipSubscriptionRequest, stream roachpb.Internal_GossipSubscriptionServer,
) error {
	__antithesis_instrumentation__.Notify(194702)
	ctx := n.storeCfg.AmbientCtx.AnnotateCtx(stream.Context())
	ctxDone := ctx.Done()

	_, isSecondaryTenant := roachpb.TenantFromContext(ctx)

	entC := make(chan *roachpb.GossipSubscriptionEvent, 256)
	entCClosed := false
	var callbackMu syncutil.Mutex
	var systemConfigUpdateCh <-chan struct{}
	for i := range args.Patterns {
		__antithesis_instrumentation__.Notify(194705)
		pattern := args.Patterns[i]
		switch pattern {

		case gossip.KeyDeprecatedSystemConfig:
			__antithesis_instrumentation__.Notify(194706)
			var unregister func()
			systemConfigUpdateCh, unregister = n.storeCfg.SystemConfigProvider.RegisterSystemConfigChannel()
			defer unregister()
		default:
			__antithesis_instrumentation__.Notify(194707)
			callback := func(key string, content roachpb.Value) {
				__antithesis_instrumentation__.Notify(194709)
				callbackMu.Lock()
				defer callbackMu.Unlock()
				if entCClosed {
					__antithesis_instrumentation__.Notify(194711)
					return
				} else {
					__antithesis_instrumentation__.Notify(194712)
				}
				__antithesis_instrumentation__.Notify(194710)
				var event roachpb.GossipSubscriptionEvent
				event.Key = key
				event.Content = content
				event.PatternMatched = pattern
				const maxBlockDur = 1 * time.Millisecond
				select {
				case entC <- &event:
					__antithesis_instrumentation__.Notify(194713)
				default:
					__antithesis_instrumentation__.Notify(194714)
					select {
					case entC <- &event:
						__antithesis_instrumentation__.Notify(194715)
					case <-time.After(maxBlockDur):
						__antithesis_instrumentation__.Notify(194716)

						close(entC)
						entCClosed = true
					}
				}
			}
			__antithesis_instrumentation__.Notify(194708)
			unregister := n.storeCfg.Gossip.RegisterCallback(pattern, callback)
			defer unregister()
		}
	}
	__antithesis_instrumentation__.Notify(194703)
	handleSystemConfigUpdate := func() error {
		__antithesis_instrumentation__.Notify(194717)
		cfg := n.storeCfg.SystemConfigProvider.GetSystemConfig()
		ents := cfg.SystemConfigEntries
		if isSecondaryTenant {
			__antithesis_instrumentation__.Notify(194720)
			ents = kvtenant.GossipSubscriptionSystemConfigMask.Apply(ents)
		} else {
			__antithesis_instrumentation__.Notify(194721)
		}
		__antithesis_instrumentation__.Notify(194718)
		var event roachpb.GossipSubscriptionEvent
		var content roachpb.Value
		if err := content.SetProto(&ents); err != nil {
			__antithesis_instrumentation__.Notify(194722)
			event.Error = roachpb.NewError(errors.Wrap(err, "could not marshal system config"))
		} else {
			__antithesis_instrumentation__.Notify(194723)
			event.Key = gossip.KeyDeprecatedSystemConfig
			event.Content = content
			event.PatternMatched = gossip.KeyDeprecatedSystemConfig
		}
		__antithesis_instrumentation__.Notify(194719)
		return stream.Send(&event)
	}
	__antithesis_instrumentation__.Notify(194704)
	for {
		__antithesis_instrumentation__.Notify(194724)
		select {
		case <-systemConfigUpdateCh:
			__antithesis_instrumentation__.Notify(194725)
			if err := handleSystemConfigUpdate(); err != nil {
				__antithesis_instrumentation__.Notify(194730)
				return errors.Wrap(err, "handling system config update")
			} else {
				__antithesis_instrumentation__.Notify(194731)
			}
		case e, ok := <-entC:
			__antithesis_instrumentation__.Notify(194726)
			if !ok {
				__antithesis_instrumentation__.Notify(194732)

				err := roachpb.NewErrorf("subscription terminated due to slow consumption")
				log.Warningf(ctx, "%v", err)
				e = &roachpb.GossipSubscriptionEvent{Error: err}
			} else {
				__antithesis_instrumentation__.Notify(194733)
			}
			__antithesis_instrumentation__.Notify(194727)
			if err := stream.Send(e); err != nil {
				__antithesis_instrumentation__.Notify(194734)
				return err
			} else {
				__antithesis_instrumentation__.Notify(194735)
			}
		case <-ctxDone:
			__antithesis_instrumentation__.Notify(194728)
			return ctx.Err()
		case <-n.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(194729)
			return stop.ErrUnavailable
		}
	}
}

func (n *Node) TenantSettings(
	args *roachpb.TenantSettingsRequest, stream roachpb.Internal_TenantSettingsServer,
) error {
	__antithesis_instrumentation__.Notify(194736)
	ctx := n.storeCfg.AmbientCtx.AnnotateCtx(stream.Context())
	ctxDone := ctx.Done()

	w := n.tenantSettingsWatcher
	if err := w.WaitForStart(ctx); err != nil {
		__antithesis_instrumentation__.Notify(194741)
		return stream.Send(&roachpb.TenantSettingsEvent{
			Error: errors.EncodeError(ctx, err),
		})
	} else {
		__antithesis_instrumentation__.Notify(194742)
	}
	__antithesis_instrumentation__.Notify(194737)

	send := func(precedence roachpb.TenantSettingsPrecedence, overrides []roachpb.TenantSetting) error {
		__antithesis_instrumentation__.Notify(194743)
		log.VInfof(ctx, 1, "sending precedence %d: %v", precedence, overrides)
		return stream.Send(&roachpb.TenantSettingsEvent{
			Precedence:  precedence,
			Incremental: false,
			Overrides:   overrides,
		})
	}
	__antithesis_instrumentation__.Notify(194738)

	allOverrides, allCh := w.GetAllTenantOverrides()
	if err := send(roachpb.AllTenantsOverrides, allOverrides); err != nil {
		__antithesis_instrumentation__.Notify(194744)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194745)
	}
	__antithesis_instrumentation__.Notify(194739)

	tenantOverrides, tenantCh := w.GetTenantOverrides(args.TenantID)
	if err := send(roachpb.SpecificTenantOverrides, tenantOverrides); err != nil {
		__antithesis_instrumentation__.Notify(194746)
		return err
	} else {
		__antithesis_instrumentation__.Notify(194747)
	}
	__antithesis_instrumentation__.Notify(194740)

	for {
		__antithesis_instrumentation__.Notify(194748)
		select {
		case <-allCh:
			__antithesis_instrumentation__.Notify(194749)

			allOverrides, allCh = w.GetAllTenantOverrides()
			if err := send(roachpb.AllTenantsOverrides, allOverrides); err != nil {
				__antithesis_instrumentation__.Notify(194753)
				return err
			} else {
				__antithesis_instrumentation__.Notify(194754)
			}

		case <-tenantCh:
			__antithesis_instrumentation__.Notify(194750)

			tenantOverrides, tenantCh = w.GetTenantOverrides(args.TenantID)
			if err := send(roachpb.SpecificTenantOverrides, tenantOverrides); err != nil {
				__antithesis_instrumentation__.Notify(194755)
				return err
			} else {
				__antithesis_instrumentation__.Notify(194756)
			}

		case <-ctxDone:
			__antithesis_instrumentation__.Notify(194751)
			return ctx.Err()

		case <-n.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(194752)
			return stop.ErrUnavailable
		}
	}
}

func (n *Node) Join(
	ctx context.Context, req *roachpb.JoinNodeRequest,
) (*roachpb.JoinNodeResponse, error) {
	__antithesis_instrumentation__.Notify(194757)
	ctx, span := n.AnnotateCtxWithSpan(ctx, "alloc-{node,store}-id")
	defer span.Finish()

	activeVersion := n.storeCfg.Settings.Version.ActiveVersion(ctx)
	if req.BinaryVersion.Less(activeVersion.Version) {
		__antithesis_instrumentation__.Notify(194762)
		return nil, grpcstatus.Error(codes.PermissionDenied, ErrIncompatibleBinaryVersion.Error())
	} else {
		__antithesis_instrumentation__.Notify(194763)
	}
	__antithesis_instrumentation__.Notify(194758)

	nodeID, err := allocateNodeID(ctx, n.storeCfg.DB)
	if err != nil {
		__antithesis_instrumentation__.Notify(194764)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194765)
	}
	__antithesis_instrumentation__.Notify(194759)

	storeID, err := allocateStoreIDs(ctx, nodeID, 1, n.storeCfg.DB)
	if err != nil {
		__antithesis_instrumentation__.Notify(194766)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194767)
	}
	__antithesis_instrumentation__.Notify(194760)

	if err := n.storeCfg.NodeLiveness.CreateLivenessRecord(ctx, nodeID); err != nil {
		__antithesis_instrumentation__.Notify(194768)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194769)
	}
	__antithesis_instrumentation__.Notify(194761)

	log.Infof(ctx, "allocated IDs: n%d, s%d", nodeID, storeID)

	return &roachpb.JoinNodeResponse{
		ClusterID:     n.clusterID.Get().GetBytes(),
		NodeID:        int32(nodeID),
		StoreID:       int32(storeID),
		ActiveVersion: &activeVersion.Version,
	}, nil
}

func (n *Node) TokenBucket(
	ctx context.Context, in *roachpb.TokenBucketRequest,
) (*roachpb.TokenBucketResponse, error) {
	__antithesis_instrumentation__.Notify(194770)

	if in.TenantID == 0 || func() bool {
		__antithesis_instrumentation__.Notify(194772)
		return in.TenantID == roachpb.SystemTenantID.ToUint64() == true
	}() == true {
		__antithesis_instrumentation__.Notify(194773)
		return &roachpb.TokenBucketResponse{
			Error: errors.EncodeError(ctx, errors.Errorf(
				"token bucket request with invalid tenant ID %d", in.TenantID,
			)),
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(194774)
	}
	__antithesis_instrumentation__.Notify(194771)
	tenantID := roachpb.MakeTenantID(in.TenantID)
	return n.tenantUsage.TokenBucketRequest(ctx, tenantID, in), nil
}

var NewTenantUsageServer = func(
	settings *cluster.Settings,
	db *kv.DB,
	executor *sql.InternalExecutor,
) multitenant.TenantUsageServer {
	__antithesis_instrumentation__.Notify(194775)
	return dummyTenantUsageServer{}
}

type dummyTenantUsageServer struct{}

func (dummyTenantUsageServer) TokenBucketRequest(
	ctx context.Context, tenantID roachpb.TenantID, in *roachpb.TokenBucketRequest,
) *roachpb.TokenBucketResponse {
	__antithesis_instrumentation__.Notify(194776)
	return &roachpb.TokenBucketResponse{
		Error: errors.EncodeError(ctx, errors.New("tenant usage requires a CCL binary")),
	}
}

func (dummyTenantUsageServer) ReconfigureTokenBucket(
	ctx context.Context,
	txn *kv.Txn,
	tenantID roachpb.TenantID,
	availableRU float64,
	refillRate float64,
	maxBurstRU float64,
	asOf time.Time,
	asOfConsumedRequestUnits float64,
) error {
	__antithesis_instrumentation__.Notify(194777)
	return errors.Errorf("tenant resource limits require a CCL binary")
}

func (dummyTenantUsageServer) Metrics() metric.Struct {
	__antithesis_instrumentation__.Notify(194778)
	return emptyMetricStruct{}
}

type emptyMetricStruct struct{}

var _ metric.Struct = emptyMetricStruct{}

func (emptyMetricStruct) MetricStruct() { __antithesis_instrumentation__.Notify(194779) }

func (n *Node) GetSpanConfigs(
	ctx context.Context, req *roachpb.GetSpanConfigsRequest,
) (*roachpb.GetSpanConfigsResponse, error) {
	__antithesis_instrumentation__.Notify(194780)
	targets, err := spanconfig.TargetsFromProtos(req.Targets)
	if err != nil {
		__antithesis_instrumentation__.Notify(194783)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194784)
	}
	__antithesis_instrumentation__.Notify(194781)
	records, err := n.spanConfigAccessor.GetSpanConfigRecords(ctx, targets)
	if err != nil {
		__antithesis_instrumentation__.Notify(194785)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194786)
	}
	__antithesis_instrumentation__.Notify(194782)

	return &roachpb.GetSpanConfigsResponse{
		SpanConfigEntries: spanconfig.RecordsToEntries(records),
	}, nil
}

func (n *Node) GetAllSystemSpanConfigsThatApply(
	ctx context.Context, req *roachpb.GetAllSystemSpanConfigsThatApplyRequest,
) (*roachpb.GetAllSystemSpanConfigsThatApplyResponse, error) {
	__antithesis_instrumentation__.Notify(194787)
	spanConfigs, err := n.spanConfigAccessor.GetAllSystemSpanConfigsThatApply(ctx, req.TenantID)
	if err != nil {
		__antithesis_instrumentation__.Notify(194789)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194790)
	}
	__antithesis_instrumentation__.Notify(194788)

	return &roachpb.GetAllSystemSpanConfigsThatApplyResponse{
		SpanConfigs: spanConfigs,
	}, nil
}

func (n *Node) UpdateSpanConfigs(
	ctx context.Context, req *roachpb.UpdateSpanConfigsRequest,
) (*roachpb.UpdateSpanConfigsResponse, error) {
	__antithesis_instrumentation__.Notify(194791)
	toUpsert, err := spanconfig.EntriesToRecords(req.ToUpsert)
	if err != nil {
		__antithesis_instrumentation__.Notify(194795)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194796)
	}
	__antithesis_instrumentation__.Notify(194792)
	toDelete, err := spanconfig.TargetsFromProtos(req.ToDelete)
	if err != nil {
		__antithesis_instrumentation__.Notify(194797)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(194798)
	}
	__antithesis_instrumentation__.Notify(194793)
	if err := n.spanConfigAccessor.UpdateSpanConfigRecords(
		ctx, toDelete, toUpsert, req.MinCommitTimestamp, req.MaxCommitTimestamp,
	); err != nil {
		__antithesis_instrumentation__.Notify(194799)
		return &roachpb.UpdateSpanConfigsResponse{
			Error: errors.EncodeError(ctx, err),
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(194800)
	}
	__antithesis_instrumentation__.Notify(194794)
	return &roachpb.UpdateSpanConfigsResponse{}, nil
}
