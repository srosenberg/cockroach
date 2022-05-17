package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/marusama/semaphore"
)

type ServerConfig struct {
	log.AmbientContext

	Settings     *cluster.Settings
	RuntimeStats RuntimeStats

	LogicalClusterID *base.ClusterIDContainer

	ClusterName string

	NodeID *base.SQLIDContainer

	Locality roachpb.Locality

	Codec keys.SQLCodec

	DB *kv.DB

	Executor sqlutil.InternalExecutor

	RPCContext   *rpc.Context
	Stopper      *stop.Stopper
	TestingKnobs TestingKnobs

	ParentMemoryMonitor *mon.BytesMonitor

	TempStorage diskmap.Factory

	TempStoragePath string

	TempFS fs.FS

	VecFDSemaphore semaphore.Semaphore

	BulkAdder kvserverbase.BulkAdderFactory

	BackfillerMonitor *mon.BytesMonitor

	BackupMonitor *mon.BytesMonitor

	ParentDiskMonitor *mon.BytesMonitor

	Metrics            *DistSQLMetrics
	RowMetrics         *row.Metrics
	InternalRowMetrics *row.Metrics

	SQLLivenessReader sqlliveness.Reader

	JobRegistry *jobs.Registry

	LeaseManager interface{}

	Gossip gossip.OptionalGossip

	NodeDialer *nodedialer.Dialer

	PodNodeDialer *nodedialer.Dialer

	SessionBoundInternalExecutorFactory sqlutil.SessionBoundInternalExecutorFactory

	ExternalStorage        cloud.ExternalStorageFactory
	ExternalStorageFromURI cloud.ExternalStorageFromURIFactory

	ProtectedTimestampProvider protectedts.Provider

	DistSender *kvcoord.DistSender

	RangeCache *rangecache.RangeCache

	SQLStatsController tree.SQLStatsController

	IndexUsageStatsController tree.IndexUsageStatsController

	SQLSQLResponseAdmissionQ *admission.WorkQueue

	CollectionFactory *descs.CollectionFactory
}

type RuntimeStats interface {
	GetCPUCombinedPercentNorm() float64
}

type TestingKnobs struct {
	RunBeforeBackfillChunk func(sp roachpb.Span) error

	RunAfterBackfillChunk func()

	SerializeIndexBackfillCreationAndIngestion chan struct{}

	IndexBackfillProgressReportInterval time.Duration

	ForceDiskSpill bool

	MemoryLimitBytes int64

	TableReaderBatchBytesLimit int64

	JoinReaderBatchBytesLimit int64

	DrainFast bool

	MetadataTestLevel MetadataTestLevel

	Changefeed base.ModuleTestingKnobs

	Flowinfra base.ModuleTestingKnobs

	BulkAdderFlushesEveryBatch bool

	JobsTestingKnobs base.ModuleTestingKnobs

	BackupRestoreTestingKnobs base.ModuleTestingKnobs

	StreamingTestingKnobs base.ModuleTestingKnobs

	IndexBackfillMergerTestingKnobs base.ModuleTestingKnobs
}

type MetadataTestLevel int

const (
	Off MetadataTestLevel = iota

	NoExplain

	On
)

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(471440) }

const DefaultMemoryLimit = 64 << 20

func GetWorkMemLimit(flowCtx *FlowCtx) int64 {
	__antithesis_instrumentation__.Notify(471441)
	if flowCtx.Cfg.TestingKnobs.ForceDiskSpill && func() bool {
		__antithesis_instrumentation__.Notify(471446)
		return flowCtx.Cfg.TestingKnobs.MemoryLimitBytes != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(471447)
		panic(errors.AssertionFailedf("both ForceDiskSpill and MemoryLimitBytes set"))
	} else {
		__antithesis_instrumentation__.Notify(471448)
	}
	__antithesis_instrumentation__.Notify(471442)
	if flowCtx.Cfg.TestingKnobs.ForceDiskSpill {
		__antithesis_instrumentation__.Notify(471449)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(471450)
	}
	__antithesis_instrumentation__.Notify(471443)
	if flowCtx.Cfg.TestingKnobs.MemoryLimitBytes != 0 {
		__antithesis_instrumentation__.Notify(471451)
		return flowCtx.Cfg.TestingKnobs.MemoryLimitBytes
	} else {
		__antithesis_instrumentation__.Notify(471452)
	}
	__antithesis_instrumentation__.Notify(471444)
	if flowCtx.EvalCtx.SessionData().WorkMemLimit <= 0 {
		__antithesis_instrumentation__.Notify(471453)

		return DefaultMemoryLimit
	} else {
		__antithesis_instrumentation__.Notify(471454)
	}
	__antithesis_instrumentation__.Notify(471445)
	return flowCtx.EvalCtx.SessionData().WorkMemLimit
}

func (flowCtx *FlowCtx) GetRowMetrics() *row.Metrics {
	__antithesis_instrumentation__.Notify(471455)
	if flowCtx.EvalCtx.SessionData().Internal {
		__antithesis_instrumentation__.Notify(471457)
		return flowCtx.Cfg.InternalRowMetrics
	} else {
		__antithesis_instrumentation__.Notify(471458)
	}
	__antithesis_instrumentation__.Notify(471456)
	return flowCtx.Cfg.RowMetrics
}
