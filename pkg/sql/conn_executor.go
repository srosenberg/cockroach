package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/contention/txnidcache"
	"github.com/cockroachdb/cockroach/pkg/sql/execstats"
	"github.com/cockroachdb/cockroach/pkg/sql/idxusage"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirecancel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessionphase"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/sslocal"
	"github.com/cockroachdb/cockroach/pkg/sql/stmtdiagnostics"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/fsm"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"golang.org/x/net/trace"
)

var noteworthyMemoryUsageBytes = envutil.EnvOrDefaultInt64("COCKROACH_NOTEWORTHY_SESSION_MEMORY_USAGE", 1024*1024)

type Server struct {
	_ util.NoCopy

	cfg *ExecutorConfig

	sqlStats *persistedsqlstats.PersistedSQLStats

	sqlStatsController *persistedsqlstats.Controller

	indexUsageStatsController *idxusage.Controller

	reportedStats sqlstats.Provider

	reportedStatsController *sslocal.Controller

	reCache *tree.RegexpCache

	pool *mon.BytesMonitor

	indexUsageStats *idxusage.LocalIndexUsageStats

	txnIDCache *txnidcache.Cache

	Metrics Metrics

	InternalMetrics Metrics

	ServerMetrics ServerMetrics

	TelemetryLoggingMetrics *TelemetryLoggingMetrics

	mu struct {
		syncutil.Mutex
		connectionCount int64
	}
}

type Metrics struct {
	EngineMetrics EngineMetrics

	StartedStatementCounters StatementCounters

	ExecutedStatementCounters StatementCounters

	GuardrailMetrics GuardrailMetrics
}

type ServerMetrics struct {
	StatsMetrics StatsMetrics

	ContentionSubsystemMetrics txnidcache.Metrics
}

func NewServer(cfg *ExecutorConfig, pool *mon.BytesMonitor) *Server {
	__antithesis_instrumentation__.Notify(457140)
	metrics := makeMetrics(false)
	serverMetrics := makeServerMetrics(cfg)
	reportedSQLStats := sslocal.New(
		cfg.Settings,
		sqlstats.MaxMemReportedSQLStatsStmtFingerprints,
		sqlstats.MaxMemReportedSQLStatsTxnFingerprints,
		serverMetrics.StatsMetrics.ReportedSQLStatsMemoryCurBytesCount,
		serverMetrics.StatsMetrics.ReportedSQLStatsMemoryMaxBytesHist,
		pool,
		nil,
		cfg.SQLStatsTestingKnobs,
	)
	reportedSQLStatsController :=
		reportedSQLStats.GetController(cfg.SQLStatusServer, cfg.DB, cfg.InternalExecutor)
	memSQLStats := sslocal.New(
		cfg.Settings,
		sqlstats.MaxMemSQLStatsStmtFingerprints,
		sqlstats.MaxMemSQLStatsTxnFingerprints,
		serverMetrics.StatsMetrics.SQLStatsMemoryCurBytesCount,
		serverMetrics.StatsMetrics.SQLStatsMemoryMaxBytesHist,
		pool,
		reportedSQLStats,
		cfg.SQLStatsTestingKnobs,
	)
	s := &Server{
		cfg:                     cfg,
		Metrics:                 metrics,
		InternalMetrics:         makeMetrics(true),
		ServerMetrics:           serverMetrics,
		pool:                    pool,
		reportedStats:           reportedSQLStats,
		reportedStatsController: reportedSQLStatsController,
		reCache:                 tree.NewRegexpCache(512),
		indexUsageStats: idxusage.NewLocalIndexUsageStats(&idxusage.Config{
			ChannelSize: idxusage.DefaultChannelSize,
			Setting:     cfg.Settings,
		}),
		txnIDCache: txnidcache.NewTxnIDCache(
			cfg.Settings,
			&serverMetrics.ContentionSubsystemMetrics),
	}

	telemetryLoggingMetrics := &TelemetryLoggingMetrics{}

	telemetryLoggingMetrics.Knobs = cfg.TelemetryLoggingTestingKnobs
	s.TelemetryLoggingMetrics = telemetryLoggingMetrics

	sqlStatsInternalExecutor := MakeInternalExecutor(context.Background(), s, MemoryMetrics{}, cfg.Settings)
	persistedSQLStats := persistedsqlstats.New(&persistedsqlstats.Config{
		Settings:         s.cfg.Settings,
		InternalExecutor: &sqlStatsInternalExecutor,
		KvDB:             cfg.DB,
		SQLIDContainer:   cfg.NodeID,
		JobRegistry:      s.cfg.JobRegistry,
		Knobs:            cfg.SQLStatsTestingKnobs,
		FlushCounter:     serverMetrics.StatsMetrics.SQLStatsFlushStarted,
		FailureCounter:   serverMetrics.StatsMetrics.SQLStatsFlushFailure,
		FlushDuration:    serverMetrics.StatsMetrics.SQLStatsFlushDuration,
	}, memSQLStats)

	s.sqlStats = persistedSQLStats
	s.sqlStatsController = persistedSQLStats.GetController(cfg.SQLStatusServer)
	s.indexUsageStatsController = idxusage.NewController(cfg.SQLStatusServer)
	return s
}

func makeMetrics(internal bool) Metrics {
	__antithesis_instrumentation__.Notify(457141)
	return Metrics{
		EngineMetrics: EngineMetrics{
			DistSQLSelectCount:    metric.NewCounter(getMetricMeta(MetaDistSQLSelect, internal)),
			SQLOptFallbackCount:   metric.NewCounter(getMetricMeta(MetaSQLOptFallback, internal)),
			SQLOptPlanCacheHits:   metric.NewCounter(getMetricMeta(MetaSQLOptPlanCacheHits, internal)),
			SQLOptPlanCacheMisses: metric.NewCounter(getMetricMeta(MetaSQLOptPlanCacheMisses, internal)),

			DistSQLExecLatency: metric.NewLatency(getMetricMeta(MetaDistSQLExecLatency, internal),
				6*metricsSampleInterval),
			SQLExecLatency: metric.NewLatency(getMetricMeta(MetaSQLExecLatency, internal),
				6*metricsSampleInterval),
			DistSQLServiceLatency: metric.NewLatency(getMetricMeta(MetaDistSQLServiceLatency, internal),
				6*metricsSampleInterval),
			SQLServiceLatency: metric.NewLatency(getMetricMeta(MetaSQLServiceLatency, internal),
				6*metricsSampleInterval),
			SQLTxnLatency: metric.NewLatency(getMetricMeta(MetaSQLTxnLatency, internal),
				6*metricsSampleInterval),
			SQLTxnsOpen:         metric.NewGauge(getMetricMeta(MetaSQLTxnsOpen, internal)),
			SQLActiveStatements: metric.NewGauge(getMetricMeta(MetaSQLActiveQueries, internal)),

			TxnAbortCount:                     metric.NewCounter(getMetricMeta(MetaTxnAbort, internal)),
			FailureCount:                      metric.NewCounter(getMetricMeta(MetaFailure, internal)),
			FullTableOrIndexScanCount:         metric.NewCounter(getMetricMeta(MetaFullTableOrIndexScan, internal)),
			FullTableOrIndexScanRejectedCount: metric.NewCounter(getMetricMeta(MetaFullTableOrIndexScanRejected, internal)),
		},
		StartedStatementCounters:  makeStartedStatementCounters(internal),
		ExecutedStatementCounters: makeExecutedStatementCounters(internal),
		GuardrailMetrics: GuardrailMetrics{
			TxnRowsWrittenLogCount: metric.NewCounter(getMetricMeta(MetaTxnRowsWrittenLog, internal)),
			TxnRowsWrittenErrCount: metric.NewCounter(getMetricMeta(MetaTxnRowsWrittenErr, internal)),
			TxnRowsReadLogCount:    metric.NewCounter(getMetricMeta(MetaTxnRowsReadLog, internal)),
			TxnRowsReadErrCount:    metric.NewCounter(getMetricMeta(MetaTxnRowsReadErr, internal)),
		},
	}
}

func makeServerMetrics(cfg *ExecutorConfig) ServerMetrics {
	__antithesis_instrumentation__.Notify(457142)
	return ServerMetrics{
		StatsMetrics: StatsMetrics{
			SQLStatsMemoryMaxBytesHist: metric.NewHistogram(
				MetaSQLStatsMemMaxBytes,
				cfg.HistogramWindowInterval,
				log10int64times1000,
				3,
			),
			SQLStatsMemoryCurBytesCount: metric.NewGauge(MetaSQLStatsMemCurBytes),
			ReportedSQLStatsMemoryMaxBytesHist: metric.NewHistogram(
				MetaReportedSQLStatsMemMaxBytes,
				cfg.HistogramWindowInterval,
				log10int64times1000,
				3,
			),
			ReportedSQLStatsMemoryCurBytesCount: metric.NewGauge(MetaReportedSQLStatsMemCurBytes),
			DiscardedStatsCount:                 metric.NewCounter(MetaDiscardedSQLStats),
			SQLStatsFlushStarted:                metric.NewCounter(MetaSQLStatsFlushStarted),
			SQLStatsFlushFailure:                metric.NewCounter(MetaSQLStatsFlushFailure),
			SQLStatsFlushDuration: metric.NewLatency(
				MetaSQLStatsFlushDuration, 6*metricsSampleInterval,
			),
			SQLStatsRemovedRows: metric.NewCounter(MetaSQLStatsRemovedRows),
			SQLTxnStatsCollectionOverhead: metric.NewLatency(
				MetaSQLTxnStatsCollectionOverhead, 6*metricsSampleInterval,
			),
		},
		ContentionSubsystemMetrics: txnidcache.NewMetrics(),
	}
}

func (s *Server) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(457143)
	s.sqlStats.Start(ctx, stopper)

	s.reportedStats.Start(ctx, stopper)

	s.txnIDCache.Start(ctx, stopper)
}

func (s *Server) GetSQLStatsController() *persistedsqlstats.Controller {
	__antithesis_instrumentation__.Notify(457144)
	return s.sqlStatsController
}

func (s *Server) GetIndexUsageStatsController() *idxusage.Controller {
	__antithesis_instrumentation__.Notify(457145)
	return s.indexUsageStatsController
}

func (s *Server) GetSQLStatsProvider() sqlstats.Provider {
	__antithesis_instrumentation__.Notify(457146)
	return s.sqlStats
}

func (s *Server) GetReportedSQLStatsController() *sslocal.Controller {
	__antithesis_instrumentation__.Notify(457147)
	return s.reportedStatsController
}

func (s *Server) GetTxnIDCache() *txnidcache.Cache {
	__antithesis_instrumentation__.Notify(457148)
	return s.txnIDCache
}

func (s *Server) GetScrubbedStmtStats(
	ctx context.Context,
) ([]roachpb.CollectedStatementStatistics, error) {
	__antithesis_instrumentation__.Notify(457149)
	return s.getScrubbedStmtStats(ctx, s.sqlStats.GetLocalMemProvider())
}

var _ = (*Server).GetScrubbedStmtStats

func (s *Server) GetUnscrubbedStmtStats(
	ctx context.Context,
) ([]roachpb.CollectedStatementStatistics, error) {
	__antithesis_instrumentation__.Notify(457150)
	var stmtStats []roachpb.CollectedStatementStatistics
	stmtStatsVisitor := func(_ context.Context, stat *roachpb.CollectedStatementStatistics) error {
		__antithesis_instrumentation__.Notify(457153)
		stmtStats = append(stmtStats, *stat)
		return nil
	}
	__antithesis_instrumentation__.Notify(457151)
	err :=
		s.sqlStats.GetLocalMemProvider().IterateStatementStats(ctx, &sqlstats.IteratorOptions{}, stmtStatsVisitor)

	if err != nil {
		__antithesis_instrumentation__.Notify(457154)
		return nil, errors.Wrap(err, "failed to fetch statement stats")
	} else {
		__antithesis_instrumentation__.Notify(457155)
	}
	__antithesis_instrumentation__.Notify(457152)

	return stmtStats, nil
}

func (s *Server) GetUnscrubbedTxnStats(
	ctx context.Context,
) ([]roachpb.CollectedTransactionStatistics, error) {
	__antithesis_instrumentation__.Notify(457156)
	var txnStats []roachpb.CollectedTransactionStatistics
	txnStatsVisitor := func(_ context.Context, stat *roachpb.CollectedTransactionStatistics) error {
		__antithesis_instrumentation__.Notify(457159)
		txnStats = append(txnStats, *stat)
		return nil
	}
	__antithesis_instrumentation__.Notify(457157)
	err :=
		s.sqlStats.GetLocalMemProvider().IterateTransactionStats(ctx, &sqlstats.IteratorOptions{}, txnStatsVisitor)

	if err != nil {
		__antithesis_instrumentation__.Notify(457160)
		return nil, errors.Wrap(err, "failed to fetch statement stats")
	} else {
		__antithesis_instrumentation__.Notify(457161)
	}
	__antithesis_instrumentation__.Notify(457158)

	return txnStats, nil
}

func (s *Server) GetScrubbedReportingStats(
	ctx context.Context,
) ([]roachpb.CollectedStatementStatistics, error) {
	__antithesis_instrumentation__.Notify(457162)
	return s.getScrubbedStmtStats(ctx, s.reportedStats)
}

func (s *Server) getScrubbedStmtStats(
	ctx context.Context, statsProvider sqlstats.Provider,
) ([]roachpb.CollectedStatementStatistics, error) {
	__antithesis_instrumentation__.Notify(457163)
	salt := ClusterSecret.Get(&s.cfg.Settings.SV)

	var scrubbedStats []roachpb.CollectedStatementStatistics
	stmtStatsVisitor := func(_ context.Context, stat *roachpb.CollectedStatementStatistics) error {
		__antithesis_instrumentation__.Notify(457166)

		scrubbedQueryStr, ok := scrubStmtStatKey(s.cfg.VirtualSchemas, stat.Key.Query)

		if !ok {
			__antithesis_instrumentation__.Notify(457169)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(457170)
		}
		__antithesis_instrumentation__.Notify(457167)

		stat.Key.Query = scrubbedQueryStr

		if !strings.HasPrefix(stat.Key.App, catconstants.ReportableAppNamePrefix) {
			__antithesis_instrumentation__.Notify(457171)
			stat.Key.App = HashForReporting(salt, stat.Key.App)
		} else {
			__antithesis_instrumentation__.Notify(457172)
		}
		__antithesis_instrumentation__.Notify(457168)

		quantizeCounts(&stat.Stats)
		stat.Stats.SensitiveInfo = stat.Stats.SensitiveInfo.GetScrubbedCopy()

		scrubbedStats = append(scrubbedStats, *stat)
		return nil
	}
	__antithesis_instrumentation__.Notify(457164)

	err :=
		statsProvider.IterateStatementStats(ctx, &sqlstats.IteratorOptions{}, stmtStatsVisitor)

	if err != nil {
		__antithesis_instrumentation__.Notify(457173)
		return nil, errors.Wrap(err, "failed to fetch scrubbed statement stats")
	} else {
		__antithesis_instrumentation__.Notify(457174)
	}
	__antithesis_instrumentation__.Notify(457165)

	return scrubbedStats, nil
}

func (s *Server) GetStmtStatsLastReset() time.Time {
	__antithesis_instrumentation__.Notify(457175)
	return s.sqlStats.GetLastReset()
}

func (s *Server) GetExecutorConfig() *ExecutorConfig {
	__antithesis_instrumentation__.Notify(457176)
	return s.cfg
}

func (s *Server) SetupConn(
	ctx context.Context,
	args SessionArgs,
	stmtBuf *StmtBuf,
	clientComm ClientComm,
	memMetrics MemoryMetrics,
	onDefaultIntSizeChange func(newSize int32),
) (ConnectionHandler, error) {
	__antithesis_instrumentation__.Notify(457177)
	sd := s.newSessionData(args)
	sds := sessiondata.NewStack(sd)

	sdMutIterator := s.makeSessionDataMutatorIterator(sds, args.SessionDefaults)
	sdMutIterator.onDefaultIntSizeChange = onDefaultIntSizeChange
	if err := sdMutIterator.applyOnEachMutatorError(func(m sessionDataMutator) error {
		__antithesis_instrumentation__.Notify(457179)
		return resetSessionVars(ctx, m)
	}); err != nil {
		__antithesis_instrumentation__.Notify(457180)
		log.Errorf(ctx, "error setting up client session: %s", err)
		return ConnectionHandler{}, err
	} else {
		__antithesis_instrumentation__.Notify(457181)
	}
	__antithesis_instrumentation__.Notify(457178)

	ex := s.newConnExecutor(
		ctx, sdMutIterator, stmtBuf, clientComm, memMetrics, &s.Metrics,
		s.sqlStats.GetApplicationStats(sd.ApplicationName),
	)
	return ConnectionHandler{ex}, nil
}

func (s *Server) IncrementConnectionCount() {
	__antithesis_instrumentation__.Notify(457182)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.connectionCount++
}

func (s *Server) DecrementConnectionCount() {
	__antithesis_instrumentation__.Notify(457183)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.connectionCount--
}

func (s *Server) IncrementConnectionCountIfLessThan(max int64) bool {
	__antithesis_instrumentation__.Notify(457184)
	s.mu.Lock()
	defer s.mu.Unlock()
	lt := s.mu.connectionCount < max
	if lt {
		__antithesis_instrumentation__.Notify(457186)
		s.mu.connectionCount++
	} else {
		__antithesis_instrumentation__.Notify(457187)
	}
	__antithesis_instrumentation__.Notify(457185)
	return lt
}

func (s *Server) GetConnectionCount() int64 {
	__antithesis_instrumentation__.Notify(457188)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.connectionCount
}

type ConnectionHandler struct {
	ex *connExecutor
}

func (h ConnectionHandler) GetParamStatus(ctx context.Context, varName string) string {
	__antithesis_instrumentation__.Notify(457189)
	name := strings.ToLower(varName)
	v, ok := varGen[name]
	if !ok {
		__antithesis_instrumentation__.Notify(457192)
		log.Fatalf(ctx, "programming error: status param %q must be defined session var", varName)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(457193)
	}
	__antithesis_instrumentation__.Notify(457190)
	hasDefault, defVal := getSessionVarDefaultString(name, v, h.ex.dataMutatorIterator.sessionDataMutatorBase)
	if !hasDefault {
		__antithesis_instrumentation__.Notify(457194)
		log.Fatalf(ctx, "programming error: status param %q must have a default value", varName)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(457195)
	}
	__antithesis_instrumentation__.Notify(457191)
	return defVal
}

func (h ConnectionHandler) GetQueryCancelKey() pgwirecancel.BackendKeyData {
	__antithesis_instrumentation__.Notify(457196)
	return h.ex.queryCancelKey
}

func (s *Server) ServeConn(
	ctx context.Context, h ConnectionHandler, reserved mon.BoundAccount, cancel context.CancelFunc,
) error {
	__antithesis_instrumentation__.Notify(457197)
	defer func() {
		__antithesis_instrumentation__.Notify(457199)
		r := recover()
		h.ex.closeWrapper(ctx, r)
	}()
	__antithesis_instrumentation__.Notify(457198)
	return h.ex.run(ctx, s.pool, reserved, cancel)
}

func (s *Server) GetLocalIndexStatistics() *idxusage.LocalIndexUsageStats {
	__antithesis_instrumentation__.Notify(457200)
	return s.indexUsageStats
}

func (s *Server) newSessionData(args SessionArgs) *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(457201)
	sd := &sessiondata.SessionData{
		SessionData: sessiondatapb.SessionData{
			UserProto: args.User.EncodeProto(),
		},
		LocalUnmigratableSessionData: sessiondata.LocalUnmigratableSessionData{
			RemoteAddr: args.RemoteAddr,
		},
		LocalOnlySessionData: sessiondatapb.LocalOnlySessionData{
			ResultsBufferSize: args.ConnResultsBufferSize,
			IsSuperuser:       args.IsSuperuser,
		},
	}
	if len(args.CustomOptionSessionDefaults) > 0 {
		__antithesis_instrumentation__.Notify(457203)
		sd.CustomOptions = make(map[string]string)
		for k, v := range args.CustomOptionSessionDefaults {
			__antithesis_instrumentation__.Notify(457204)
			sd.CustomOptions[k] = v
		}
	} else {
		__antithesis_instrumentation__.Notify(457205)
	}
	__antithesis_instrumentation__.Notify(457202)
	s.populateMinimalSessionData(sd)
	return sd
}

func (s *Server) makeSessionDataMutatorIterator(
	sds *sessiondata.Stack, defaults SessionDefaults,
) *sessionDataMutatorIterator {
	__antithesis_instrumentation__.Notify(457206)
	return &sessionDataMutatorIterator{
		sds: sds,
		sessionDataMutatorBase: sessionDataMutatorBase{
			defaults: defaults,
			settings: s.cfg.Settings,
		},
		sessionDataMutatorCallbacks: sessionDataMutatorCallbacks{},
	}
}

func (s *Server) populateMinimalSessionData(sd *sessiondata.SessionData) {
	__antithesis_instrumentation__.Notify(457207)
	if sd.SequenceState == nil {
		__antithesis_instrumentation__.Notify(457210)
		sd.SequenceState = sessiondata.NewSequenceState()
	} else {
		__antithesis_instrumentation__.Notify(457211)
	}
	__antithesis_instrumentation__.Notify(457208)
	if sd.Location == nil {
		__antithesis_instrumentation__.Notify(457212)
		sd.Location = time.UTC
	} else {
		__antithesis_instrumentation__.Notify(457213)
	}
	__antithesis_instrumentation__.Notify(457209)
	if len(sd.SearchPath.GetPathArray()) == 0 {
		__antithesis_instrumentation__.Notify(457214)
		sd.SearchPath = sessiondata.DefaultSearchPathForUser(sd.User())
	} else {
		__antithesis_instrumentation__.Notify(457215)
	}
}

func (s *Server) newConnExecutor(
	ctx context.Context,
	sdMutIterator *sessionDataMutatorIterator,
	stmtBuf *StmtBuf,
	clientComm ClientComm,
	memMetrics MemoryMetrics,
	srvMetrics *Metrics,
	applicationStats sqlstats.ApplicationStats,
) *connExecutor {
	__antithesis_instrumentation__.Notify(457216)

	sessionRootMon := mon.NewMonitor(
		"session root",
		mon.MemoryResource,
		memMetrics.CurBytesCount,
		memMetrics.MaxBytesHist,
		-1, math.MaxInt64, s.cfg.Settings,
	)
	sessionMon := mon.NewMonitor(
		"session",
		mon.MemoryResource,
		memMetrics.SessionCurBytesCount,
		memMetrics.SessionMaxBytesHist,
		-1, noteworthyMemoryUsageBytes, s.cfg.Settings,
	)

	txnMon := mon.NewMonitor(
		"txn",
		mon.MemoryResource,
		memMetrics.TxnCurBytesCount,
		memMetrics.TxnMaxBytesHist,
		-1, noteworthyMemoryUsageBytes, s.cfg.Settings,
	)

	nodeIDOrZero, _ := s.cfg.NodeID.OptionalNodeID()
	ex := &connExecutor{
		server:              s,
		metrics:             srvMetrics,
		stmtBuf:             stmtBuf,
		clientComm:          clientComm,
		mon:                 sessionRootMon,
		sessionMon:          sessionMon,
		sessionDataStack:    sdMutIterator.sds,
		dataMutatorIterator: sdMutIterator,
		state: txnState{
			mon:                          txnMon,
			connCtx:                      ctx,
			testingForceRealTracingSpans: s.cfg.TestingKnobs.ForceRealTracingSpans,
		},
		transitionCtx: transitionCtx{
			db:           s.cfg.DB,
			nodeIDOrZero: nodeIDOrZero,
			clock:        s.cfg.Clock,

			connMon:          sessionRootMon,
			tracer:           s.cfg.AmbientCtx.Tracer,
			settings:         s.cfg.Settings,
			execTestingKnobs: s.GetExecutorConfig().TestingKnobs,
		},
		memMetrics: memMetrics,
		planner:    planner{execCfg: s.cfg, alloc: &tree.DatumAlloc{}},

		ctxHolder:                 ctxHolder{connCtx: ctx},
		phaseTimes:                sessionphase.NewTimes(),
		rng:                       rand.New(rand.NewSource(timeutil.Now().UnixNano())),
		executorType:              executorTypeExec,
		hasCreatedTemporarySchema: false,
		stmtDiagnosticsRecorder:   s.cfg.StmtDiagnosticsRecorder,
		indexUsageStats:           s.indexUsageStats,
		txnIDCacheWriter:          s.txnIDCache,
	}

	ex.state.txnAbortCount = ex.metrics.EngineMetrics.TxnAbortCount

	ex.dataMutatorIterator.setCurTxnReadOnly = func(val bool) {
		__antithesis_instrumentation__.Notify(457220)
		ex.state.readOnly = val
	}
	__antithesis_instrumentation__.Notify(457217)
	ex.dataMutatorIterator.onTempSchemaCreation = func() {
		__antithesis_instrumentation__.Notify(457221)
		ex.hasCreatedTemporarySchema = true
	}
	__antithesis_instrumentation__.Notify(457218)

	ex.applicationName.Store(ex.sessionData().ApplicationName)
	ex.applicationStats = applicationStats
	ex.statsCollector = sslocal.NewStatsCollector(
		s.cfg.Settings,
		applicationStats,
		ex.phaseTimes,
		s.cfg.SQLStatsTestingKnobs,
	)
	ex.dataMutatorIterator.onApplicationNameChange = func(newName string) {
		__antithesis_instrumentation__.Notify(457222)
		ex.applicationName.Store(newName)
		ex.applicationStats = ex.server.sqlStats.GetApplicationStats(newName)
	}
	__antithesis_instrumentation__.Notify(457219)

	ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionInit, timeutil.Now())

	ex.extraTxnState.prepStmtsNamespace = prepStmtNamespace{
		prepStmts: make(map[string]*PreparedStatement),
		portals:   make(map[string]PreparedPortal),
	}
	ex.extraTxnState.prepStmtsNamespaceAtTxnRewindPos = prepStmtNamespace{
		prepStmts: make(map[string]*PreparedStatement),
		portals:   make(map[string]PreparedPortal),
	}
	ex.extraTxnState.prepStmtsNamespaceMemAcc = ex.sessionMon.MakeBoundAccount()
	ex.extraTxnState.descCollection = s.cfg.CollectionFactory.MakeCollection(ctx, descs.NewTemporarySchemaProvider(sdMutIterator.sds))
	ex.extraTxnState.txnRewindPos = -1
	ex.extraTxnState.schemaChangeJobRecords = make(map[descpb.ID]*jobs.Record)
	ex.queryCancelKey = pgwirecancel.MakeBackendKeyData(ex.rng, ex.server.cfg.NodeID.SQLInstanceID())
	ex.mu.ActiveQueries = make(map[ClusterWideID]*queryMeta)
	ex.machine = fsm.MakeMachine(TxnStateTransitions, stateNoTxn{}, &ex.state)

	ex.sessionTracing.ex = ex
	ex.transitionCtx.sessionTracing = &ex.sessionTracing

	ex.extraTxnState.hasAdminRoleCache = HasAdminRoleCache{}

	ex.extraTxnState.atomicAutoRetryCounter = new(int32)

	ex.extraTxnState.createdSequences = make(map[descpb.ID]struct{})

	ex.initPlanner(ctx, &ex.planner)

	return ex
}

func (s *Server) newConnExecutorWithTxn(
	ctx context.Context,
	sdMutIterator *sessionDataMutatorIterator,
	stmtBuf *StmtBuf,
	clientComm ClientComm,
	parentMon *mon.BytesMonitor,
	memMetrics MemoryMetrics,
	srvMetrics *Metrics,
	txn *kv.Txn,
	syntheticDescs []catalog.Descriptor,
	applicationStats sqlstats.ApplicationStats,
) *connExecutor {
	__antithesis_instrumentation__.Notify(457223)
	ex := s.newConnExecutor(
		ctx,
		sdMutIterator,
		stmtBuf,
		clientComm,
		memMetrics,
		srvMetrics,
		applicationStats,
	)
	if txn.Type() == kv.LeafTxn {
		__antithesis_instrumentation__.Notify(457225)

		ex.dataMutatorIterator.applyOnEachMutator(func(m sessionDataMutator) {
			__antithesis_instrumentation__.Notify(457226)
			m.SetReadOnly(true)
		})
	} else {
		__antithesis_instrumentation__.Notify(457227)
	}
	__antithesis_instrumentation__.Notify(457224)

	ex.activate(ctx, parentMon, mon.BoundAccount{})

	ex.machine = fsm.MakeMachine(
		BoundTxnStateTransitions,
		stateOpen{ImplicitTxn: fsm.False},
		&ex.state,
	)
	ex.state.resetForNewSQLTxn(
		ctx,
		explicitTxn,
		txn.ReadTimestamp().GoTime(),
		nil,
		roachpb.UnspecifiedUserPriority,
		tree.ReadWrite,
		txn,
		ex.transitionCtx,
		ex.QualityOfService())

	ex.extraTxnState.descCollection.SetSyntheticDescriptors(syntheticDescs)
	return ex
}

type closeType int

const (
	normalClose closeType = iota
	panicClose

	externalTxnClose
)

func (ex *connExecutor) closeWrapper(ctx context.Context, recovered interface{}) {
	__antithesis_instrumentation__.Notify(457228)
	if recovered != nil {
		__antithesis_instrumentation__.Notify(457230)
		panicErr := logcrash.PanicAsError(1, recovered)

		if ex.curStmtAST != nil {
			__antithesis_instrumentation__.Notify(457232)

			log.SqlExec.Shoutf(ctx, severity.ERROR,
				"a SQL panic has occurred while executing the following statement:\n%s",

				truncateStatementStringForTelemetry(ex.curStmtAST.String()))

			vt := ex.planner.extendedEvalCtx.VirtualSchemas
			panicErr = WithAnonymizedStatement(panicErr, ex.curStmtAST, vt)
		} else {
			__antithesis_instrumentation__.Notify(457233)
		}
		__antithesis_instrumentation__.Notify(457231)

		logcrash.ReportPanic(ctx, &ex.server.cfg.Settings.SV, panicErr, 1)

		ex.close(ctx, panicClose)

		panic(panicErr)
	} else {
		__antithesis_instrumentation__.Notify(457234)
	}
	__antithesis_instrumentation__.Notify(457229)

	closeCtx := ex.server.cfg.AmbientCtx.AnnotateCtx(context.Background())

	closeCtx = logtags.AddTags(closeCtx, logtags.FromContext(ctx))
	ex.close(closeCtx, normalClose)
}

func (ex *connExecutor) close(ctx context.Context, closeType closeType) {
	__antithesis_instrumentation__.Notify(457235)
	ex.sessionEventf(ctx, "finishing connExecutor")

	txnEvType := noEvent
	if _, noTxn := ex.machine.CurState().(stateNoTxn); !noTxn {
		__antithesis_instrumentation__.Notify(457243)
		txnEvType = txnRollback
	} else {
		__antithesis_instrumentation__.Notify(457244)
	}
	__antithesis_instrumentation__.Notify(457236)

	if closeType == normalClose {
		__antithesis_instrumentation__.Notify(457245)

		ev := eventNonRetriableErr{IsCommit: fsm.True}
		payload := eventNonRetriableErrPayload{err: pgerror.Newf(pgcode.AdminShutdown,
			"connExecutor closing")}
		if err := ex.machine.ApplyWithPayload(ctx, ev, payload); err != nil {
			__antithesis_instrumentation__.Notify(457248)
			log.Warningf(ctx, "error while cleaning up connExecutor: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(457249)
		}
		__antithesis_instrumentation__.Notify(457246)
		switch t := ex.machine.CurState().(type) {
		case stateNoTxn:
			__antithesis_instrumentation__.Notify(457250)

		case stateAborted:
			__antithesis_instrumentation__.Notify(457251)

		case stateCommitWait:
			__antithesis_instrumentation__.Notify(457252)
			ex.state.finishSQLTxn()
		default:
			__antithesis_instrumentation__.Notify(457253)
			if buildutil.CrdbTestBuild {
				__antithesis_instrumentation__.Notify(457254)
				panic(errors.AssertionFailedf("unexpected state in conn executor after ApplyWithPayload %T", t))
			} else {
				__antithesis_instrumentation__.Notify(457255)
			}
		}
		__antithesis_instrumentation__.Notify(457247)
		if buildutil.CrdbTestBuild && func() bool {
			__antithesis_instrumentation__.Notify(457256)
			return ex.state.Ctx != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(457257)
			panic(errors.AssertionFailedf("txn span not closed in state %s", ex.machine.CurState()))
		} else {
			__antithesis_instrumentation__.Notify(457258)
		}
	} else {
		__antithesis_instrumentation__.Notify(457259)
		if closeType == externalTxnClose {
			__antithesis_instrumentation__.Notify(457260)
			ex.state.finishExternalTxn()
		} else {
			__antithesis_instrumentation__.Notify(457261)
		}
	}
	__antithesis_instrumentation__.Notify(457237)

	if err := ex.resetExtraTxnState(ctx, txnEvent{eventType: txnEvType}); err != nil {
		__antithesis_instrumentation__.Notify(457262)
		log.Warningf(ctx, "error while cleaning up connExecutor: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(457263)
	}
	__antithesis_instrumentation__.Notify(457238)

	if ex.hasCreatedTemporarySchema && func() bool {
		__antithesis_instrumentation__.Notify(457264)
		return !ex.server.cfg.TestingKnobs.DisableTempObjectsCleanupOnSessionExit == true
	}() == true {
		__antithesis_instrumentation__.Notify(457265)
		ie := MakeInternalExecutor(ctx, ex.server, MemoryMetrics{}, ex.server.cfg.Settings)
		err := cleanupSessionTempObjects(
			ctx,
			ex.server.cfg.Settings,
			ex.server.cfg.CollectionFactory,
			ex.server.cfg.DB,
			ex.server.cfg.Codec,
			&ie,
			ex.sessionID,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(457266)
			log.Errorf(
				ctx,
				"error deleting temporary objects at session close, "+
					"the temp tables deletion job will retry periodically: %s",
				err,
			)
		} else {
			__antithesis_instrumentation__.Notify(457267)
		}
	} else {
		__antithesis_instrumentation__.Notify(457268)
	}
	__antithesis_instrumentation__.Notify(457239)

	if closeType != panicClose {
		__antithesis_instrumentation__.Notify(457269)

		ex.extraTxnState.prepStmtsNamespace.resetToEmpty(
			ctx, &ex.extraTxnState.prepStmtsNamespaceMemAcc,
		)
		ex.extraTxnState.prepStmtsNamespaceAtTxnRewindPos.resetToEmpty(
			ctx, &ex.extraTxnState.prepStmtsNamespaceMemAcc,
		)
		ex.extraTxnState.prepStmtsNamespaceMemAcc.Close(ctx)
		ex.extraTxnState.sqlCursors.closeAll()
	} else {
		__antithesis_instrumentation__.Notify(457270)
	}
	__antithesis_instrumentation__.Notify(457240)

	if ex.sessionTracing.Enabled() {
		__antithesis_instrumentation__.Notify(457271)
		if err := ex.sessionTracing.StopTracing(); err != nil {
			__antithesis_instrumentation__.Notify(457272)
			log.Warningf(ctx, "error stopping tracing: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(457273)
		}
	} else {
		__antithesis_instrumentation__.Notify(457274)
	}
	__antithesis_instrumentation__.Notify(457241)

	if ex.eventLog != nil {
		__antithesis_instrumentation__.Notify(457275)
		ex.eventLog.Finish()
		ex.eventLog = nil
	} else {
		__antithesis_instrumentation__.Notify(457276)
	}
	__antithesis_instrumentation__.Notify(457242)

	ex.mu.IdleInSessionTimeout.Stop()
	ex.mu.IdleInTransactionSessionTimeout.Stop()

	if closeType != panicClose {
		__antithesis_instrumentation__.Notify(457277)
		ex.state.mon.Stop(ctx)
		ex.sessionMon.Stop(ctx)
		ex.mon.Stop(ctx)
	} else {
		__antithesis_instrumentation__.Notify(457278)
		ex.state.mon.EmergencyStop(ctx)
		ex.sessionMon.EmergencyStop(ctx)
		ex.mon.EmergencyStop(ctx)
	}
}

type HasAdminRoleCache struct {
	HasAdminRole bool

	IsSet bool
}

type connExecutor struct {
	_ util.NoCopy

	server *Server

	metrics *Metrics

	mon        *mon.BytesMonitor
	sessionMon *mon.BytesMonitor

	memMetrics MemoryMetrics

	stmtBuf *StmtBuf

	clientComm ClientComm

	machine fsm.Machine

	state          txnState
	transitionCtx  transitionCtx
	sessionTracing SessionTracing

	eventLog trace.EventLog

	extraTxnState struct {
		descCollection descs.Collection

		jobs jobsCollection

		schemaChangeJobRecords map[descpb.ID]*jobs.Record

		atomicAutoRetryCounter *int32

		autoRetryReason error

		firstStmtExecuted bool

		numDDL int

		numRows int

		txnCounter int

		txnRewindPos CmdPos

		prepStmtsNamespace prepStmtNamespace

		prepStmtsNamespaceAtTxnRewindPos prepStmtNamespace

		prepStmtsNamespaceMemAcc mon.BoundAccount

		sqlCursors cursorMap

		shouldExecuteOnTxnFinish bool

		txnFinishClosure struct {
			txnStartTime time.Time

			implicit bool
		}

		shouldExecuteOnTxnRestart bool

		savepoints savepointStack

		rewindPosSnapshot struct {
			savepoints       savepointStack
			sessionDataStack *sessiondata.Stack
		}

		transactionStatementFingerprintIDs []roachpb.StmtFingerprintID

		transactionStatementsHash util.FNV64

		schemaChangerState SchemaChangerState

		shouldCollectTxnExecutionStats bool

		accumulatedStats execstats.QueryLevelStats

		rowsRead  int64
		bytesRead int64

		rowsWritten int64

		rowsWrittenLogged bool
		rowsReadLogged    bool

		hasAdminRoleCache HasAdminRoleCache

		createdSequences map[descpb.ID]struct{}
	}

	sessionDataStack *sessiondata.Stack

	dataMutatorIterator *sessionDataMutatorIterator

	applicationStats sqlstats.ApplicationStats

	statsCollector sqlstats.StatsCollector

	applicationName atomic.Value

	ctxHolder ctxHolder

	onCancelSession context.CancelFunc

	planner planner

	phaseTimes *sessionphase.Times

	rng *rand.Rand

	mu struct {
		syncutil.RWMutex

		ActiveQueries map[ClusterWideID]*queryMeta

		LastActiveQuery tree.Statement

		IdleInSessionTimeout timeout

		IdleInTransactionSessionTimeout timeout
	}

	curStmtAST tree.Statement

	queryCancelKey pgwirecancel.BackendKeyData

	sessionID ClusterWideID

	activated bool

	draining bool

	executorType executorType

	hasCreatedTemporarySchema bool

	stmtDiagnosticsRecorder *stmtdiagnostics.Registry

	indexUsageStats *idxusage.LocalIndexUsageStats

	txnIDCacheWriter txnidcache.Writer
}

type ctxHolder struct {
	connCtx           context.Context
	sessionTracingCtx context.Context
}

type timeout struct {
	timeout *time.Timer
}

func (t timeout) Stop() {
	__antithesis_instrumentation__.Notify(457279)
	if t.timeout != nil {
		__antithesis_instrumentation__.Notify(457280)
		t.timeout.Stop()
	} else {
		__antithesis_instrumentation__.Notify(457281)
	}
}

func (ch *ctxHolder) ctx() context.Context {
	__antithesis_instrumentation__.Notify(457282)
	if ch.sessionTracingCtx != nil {
		__antithesis_instrumentation__.Notify(457284)
		return ch.sessionTracingCtx
	} else {
		__antithesis_instrumentation__.Notify(457285)
	}
	__antithesis_instrumentation__.Notify(457283)
	return ch.connCtx
}

func (ch *ctxHolder) hijack(sessionTracingCtx context.Context) {
	__antithesis_instrumentation__.Notify(457286)
	if ch.sessionTracingCtx != nil {
		__antithesis_instrumentation__.Notify(457288)
		panic("hijack already in effect")
	} else {
		__antithesis_instrumentation__.Notify(457289)
	}
	__antithesis_instrumentation__.Notify(457287)
	ch.sessionTracingCtx = sessionTracingCtx
}

func (ch *ctxHolder) unhijack() {
	__antithesis_instrumentation__.Notify(457290)
	if ch.sessionTracingCtx == nil {
		__antithesis_instrumentation__.Notify(457292)
		panic("hijack not in effect")
	} else {
		__antithesis_instrumentation__.Notify(457293)
	}
	__antithesis_instrumentation__.Notify(457291)
	ch.sessionTracingCtx = nil
}

type prepStmtNamespace struct {
	prepStmts map[string]*PreparedStatement

	portals map[string]PreparedPortal
}

func (ns prepStmtNamespace) HasActivePortals() bool {
	__antithesis_instrumentation__.Notify(457294)
	return len(ns.portals) > 0
}

func (ns prepStmtNamespace) HasPortal(s string) bool {
	__antithesis_instrumentation__.Notify(457295)
	_, ok := ns.portals[s]
	return ok
}

func (ns prepStmtNamespace) MigratablePreparedStatements() []sessiondatapb.MigratableSession_PreparedStatement {
	__antithesis_instrumentation__.Notify(457296)
	ret := make([]sessiondatapb.MigratableSession_PreparedStatement, 0, len(ns.prepStmts))
	for name, stmt := range ns.prepStmts {
		__antithesis_instrumentation__.Notify(457298)
		ret = append(
			ret,
			sessiondatapb.MigratableSession_PreparedStatement{
				Name:                 name,
				PlaceholderTypeHints: stmt.InferredTypes,
				SQL:                  stmt.SQL,
			},
		)
	}
	__antithesis_instrumentation__.Notify(457297)
	return ret
}

func (ns prepStmtNamespace) String() string {
	__antithesis_instrumentation__.Notify(457299)
	var sb strings.Builder
	sb.WriteString("Prep stmts: ")
	for name := range ns.prepStmts {
		__antithesis_instrumentation__.Notify(457302)
		sb.WriteString(name + " ")
	}
	__antithesis_instrumentation__.Notify(457300)
	sb.WriteString("Portals: ")
	for name := range ns.portals {
		__antithesis_instrumentation__.Notify(457303)
		sb.WriteString(name + " ")
	}
	__antithesis_instrumentation__.Notify(457301)
	return sb.String()
}

func (ns *prepStmtNamespace) resetToEmpty(
	ctx context.Context, prepStmtsNamespaceMemAcc *mon.BoundAccount,
) {
	__antithesis_instrumentation__.Notify(457304)

	_ = ns.resetTo(ctx, prepStmtNamespace{}, prepStmtsNamespaceMemAcc)
}

func (ns *prepStmtNamespace) resetTo(
	ctx context.Context, to prepStmtNamespace, prepStmtsNamespaceMemAcc *mon.BoundAccount,
) error {
	__antithesis_instrumentation__.Notify(457305)
	for name, p := range ns.prepStmts {
		__antithesis_instrumentation__.Notify(457310)
		p.decRef(ctx)
		delete(ns.prepStmts, name)
	}
	__antithesis_instrumentation__.Notify(457306)
	for name, p := range ns.portals {
		__antithesis_instrumentation__.Notify(457311)
		p.close(ctx, prepStmtsNamespaceMemAcc, name)
		delete(ns.portals, name)
	}
	__antithesis_instrumentation__.Notify(457307)

	for name, ps := range to.prepStmts {
		__antithesis_instrumentation__.Notify(457312)
		ps.incRef(ctx)
		ns.prepStmts[name] = ps
	}
	__antithesis_instrumentation__.Notify(457308)
	for name, p := range to.portals {
		__antithesis_instrumentation__.Notify(457313)
		if err := p.accountForCopy(ctx, prepStmtsNamespaceMemAcc, name); err != nil {
			__antithesis_instrumentation__.Notify(457315)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457316)
		}
		__antithesis_instrumentation__.Notify(457314)
		ns.portals[name] = p
	}
	__antithesis_instrumentation__.Notify(457309)
	return nil
}

func (ex *connExecutor) resetExtraTxnState(ctx context.Context, ev txnEvent) error {
	__antithesis_instrumentation__.Notify(457317)
	ex.extraTxnState.jobs = nil
	ex.extraTxnState.firstStmtExecuted = false
	ex.extraTxnState.hasAdminRoleCache = HasAdminRoleCache{}
	ex.extraTxnState.schemaChangerState = SchemaChangerState{
		mode: ex.sessionData().NewSchemaChangerMode,
	}

	for k := range ex.extraTxnState.schemaChangeJobRecords {
		__antithesis_instrumentation__.Notify(457321)
		delete(ex.extraTxnState.schemaChangeJobRecords, k)
	}
	__antithesis_instrumentation__.Notify(457318)

	ex.extraTxnState.descCollection.ReleaseAll(ctx)

	for name, p := range ex.extraTxnState.prepStmtsNamespace.portals {
		__antithesis_instrumentation__.Notify(457322)
		p.close(ctx, &ex.extraTxnState.prepStmtsNamespaceMemAcc, name)
		delete(ex.extraTxnState.prepStmtsNamespace.portals, name)
	}
	__antithesis_instrumentation__.Notify(457319)

	ex.extraTxnState.sqlCursors.closeAll()

	ex.extraTxnState.createdSequences = make(map[descpb.ID]struct{})

	switch ev.eventType {
	case txnCommit, txnRollback:
		__antithesis_instrumentation__.Notify(457323)
		for name, p := range ex.extraTxnState.prepStmtsNamespaceAtTxnRewindPos.portals {
			__antithesis_instrumentation__.Notify(457327)
			p.close(ctx, &ex.extraTxnState.prepStmtsNamespaceMemAcc, name)
			delete(ex.extraTxnState.prepStmtsNamespaceAtTxnRewindPos.portals, name)
		}
		__antithesis_instrumentation__.Notify(457324)
		ex.extraTxnState.savepoints.clear()
		ex.onTxnFinish(ctx, ev)
	case txnRestart:
		__antithesis_instrumentation__.Notify(457325)
		ex.onTxnRestart(ctx)
		ex.state.mu.Lock()
		defer ex.state.mu.Unlock()
		ex.state.mu.stmtCount = 0
	default:
		__antithesis_instrumentation__.Notify(457326)
	}
	__antithesis_instrumentation__.Notify(457320)

	return nil
}

func (ex *connExecutor) Ctx() context.Context {
	__antithesis_instrumentation__.Notify(457328)
	ctx := ex.state.Ctx
	if _, ok := ex.machine.CurState().(stateNoTxn); ok {
		__antithesis_instrumentation__.Notify(457331)
		ctx = ex.ctxHolder.ctx()
	} else {
		__antithesis_instrumentation__.Notify(457332)
	}
	__antithesis_instrumentation__.Notify(457329)

	if _, ok := ex.machine.CurState().(stateInternalError); ok {
		__antithesis_instrumentation__.Notify(457333)
		ctx = ex.ctxHolder.ctx()
	} else {
		__antithesis_instrumentation__.Notify(457334)
	}
	__antithesis_instrumentation__.Notify(457330)
	return ctx
}

func (ex *connExecutor) sessionData() *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(457335)
	if ex.sessionDataStack == nil {
		__antithesis_instrumentation__.Notify(457337)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(457338)
	}
	__antithesis_instrumentation__.Notify(457336)
	return ex.sessionDataStack.Top()
}

func (ex *connExecutor) activate(
	ctx context.Context, parentMon *mon.BytesMonitor, reserved mon.BoundAccount,
) {
	__antithesis_instrumentation__.Notify(457339)

	ex.mon.Start(ctx, parentMon, reserved)
	ex.sessionMon.Start(ctx, ex.mon, mon.BoundAccount{})

	if traceSessionEventLogEnabled.Get(&ex.server.cfg.Settings.SV) {
		__antithesis_instrumentation__.Notify(457341)
		remoteStr := "<admin>"
		if ex.sessionData().RemoteAddr != nil {
			__antithesis_instrumentation__.Notify(457343)
			remoteStr = ex.sessionData().RemoteAddr.String()
		} else {
			__antithesis_instrumentation__.Notify(457344)
		}
		__antithesis_instrumentation__.Notify(457342)
		ex.eventLog = trace.NewEventLog(
			fmt.Sprintf("sql session [%s]", ex.sessionData().User()), remoteStr)
	} else {
		__antithesis_instrumentation__.Notify(457345)
	}
	__antithesis_instrumentation__.Notify(457340)

	ex.activated = true
}

func (ex *connExecutor) run(
	ctx context.Context,
	parentMon *mon.BytesMonitor,
	reserved mon.BoundAccount,
	onCancel context.CancelFunc,
) error {
	__antithesis_instrumentation__.Notify(457346)
	if !ex.activated {
		__antithesis_instrumentation__.Notify(457348)
		ex.activate(ctx, parentMon, reserved)
	} else {
		__antithesis_instrumentation__.Notify(457349)
	}
	__antithesis_instrumentation__.Notify(457347)
	ex.ctxHolder.connCtx = ctx
	ex.onCancelSession = onCancel

	ex.sessionID = ex.generateID()
	ex.server.cfg.SessionRegistry.register(ex.sessionID, ex.queryCancelKey, ex)
	ex.planner.extendedEvalCtx.setSessionID(ex.sessionID)
	defer ex.server.cfg.SessionRegistry.deregister(ex.sessionID, ex.queryCancelKey)
	for {
		__antithesis_instrumentation__.Notify(457350)
		ex.curStmtAST = nil
		if err := ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(457352)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457353)
		}
		__antithesis_instrumentation__.Notify(457351)

		var err error
		if err = ex.execCmd(); err != nil {
			__antithesis_instrumentation__.Notify(457354)
			if errors.IsAny(err, io.EOF, errDrainingComplete) {
				__antithesis_instrumentation__.Notify(457356)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(457357)
			}
			__antithesis_instrumentation__.Notify(457355)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457358)
		}
	}
}

var errDrainingComplete = fmt.Errorf("draining done. this is a good time to finish this session")

func (ex *connExecutor) execCmd() error {
	__antithesis_instrumentation__.Notify(457359)
	ctx := ex.Ctx()
	cmd, pos, err := ex.stmtBuf.CurCmd()
	if err != nil {
		__antithesis_instrumentation__.Notify(457369)
		return err
	} else {
		__antithesis_instrumentation__.Notify(457370)
	}
	__antithesis_instrumentation__.Notify(457360)

	if log.ExpensiveLogEnabled(ctx, 2) || func() bool {
		__antithesis_instrumentation__.Notify(457371)
		return ex.eventLog != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(457372)
		ex.sessionEventf(ctx, "[%s pos:%d] executing %s",
			ex.machine.CurState(), pos, cmd)
	} else {
		__antithesis_instrumentation__.Notify(457373)
	}
	__antithesis_instrumentation__.Notify(457361)

	var ev fsm.Event
	var payload fsm.EventPayload
	var res ResultBase
	switch tcmd := cmd.(type) {
	case ExecStmt:
		__antithesis_instrumentation__.Notify(457374)
		ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionQueryReceived, tcmd.TimeReceived)
		ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionStartParse, tcmd.ParseStart)
		ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionEndParse, tcmd.ParseEnd)

		err := func() error {
			__antithesis_instrumentation__.Notify(457390)
			if tcmd.AST == nil {
				__antithesis_instrumentation__.Notify(457392)
				res = ex.clientComm.CreateEmptyQueryResult(pos)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(457393)
			}
			__antithesis_instrumentation__.Notify(457391)
			ex.curStmtAST = tcmd.AST

			stmtRes := ex.clientComm.CreateStatementResult(
				tcmd.AST,
				NeedRowDesc,
				pos,
				nil,
				ex.sessionData().DataConversionConfig,
				ex.sessionData().GetLocation(),
				0,
				"",
				ex.implicitTxn(),
			)
			res = stmtRes

			implicitTxnForBatch := ex.sessionData().EnableImplicitTransactionForBatchStatements
			canAutoCommit := ex.implicitTxn() && func() bool {
				__antithesis_instrumentation__.Notify(457394)
				return (tcmd.LastInBatch || func() bool {
					__antithesis_instrumentation__.Notify(457395)
					return !implicitTxnForBatch == true
				}() == true) == true
			}() == true
			ev, payload, err = ex.execStmt(
				ctx, tcmd.Statement, nil, nil, stmtRes, canAutoCommit,
			)
			return err
		}()
		__antithesis_instrumentation__.Notify(457375)

		ex.statsCollector.PhaseTimes().SetSessionPhaseTime(sessionphase.SessionQueryServiced, timeutil.Now())
		if err != nil {
			__antithesis_instrumentation__.Notify(457396)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457397)
		}

	case ExecPortal:
		__antithesis_instrumentation__.Notify(457376)

		ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionQueryReceived, tcmd.TimeReceived)

		ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionStartParse, time.Time{})
		ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionEndParse, time.Time{})

		err := func() error {
			__antithesis_instrumentation__.Notify(457398)
			portalName := tcmd.Name
			portal, ok := ex.extraTxnState.prepStmtsNamespace.portals[portalName]
			if !ok {
				__antithesis_instrumentation__.Notify(457402)
				err := pgerror.Newf(
					pgcode.InvalidCursorName, "unknown portal %q", portalName)
				ev = eventNonRetriableErr{IsCommit: fsm.False}
				payload = eventNonRetriableErrPayload{err: err}
				res = ex.clientComm.CreateErrorResult(pos)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(457403)
			}
			__antithesis_instrumentation__.Notify(457399)
			if portal.Stmt.AST == nil {
				__antithesis_instrumentation__.Notify(457404)
				res = ex.clientComm.CreateEmptyQueryResult(pos)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(457405)
			}
			__antithesis_instrumentation__.Notify(457400)

			if log.ExpensiveLogEnabled(ctx, 2) {
				__antithesis_instrumentation__.Notify(457406)
				log.VEventf(ctx, 2, "portal resolved to: %s", portal.Stmt.AST.String())
			} else {
				__antithesis_instrumentation__.Notify(457407)
			}
			__antithesis_instrumentation__.Notify(457401)
			ex.curStmtAST = portal.Stmt.AST

			pinfo := &tree.PlaceholderInfo{
				PlaceholderTypesInfo: tree.PlaceholderTypesInfo{
					TypeHints: portal.Stmt.TypeHints,
					Types:     portal.Stmt.Types,
				},
				Values: portal.Qargs,
			}

			stmtRes := ex.clientComm.CreateStatementResult(
				portal.Stmt.AST,

				DontNeedRowDesc,
				pos, portal.OutFormats,
				ex.sessionData().DataConversionConfig,
				ex.sessionData().GetLocation(),
				tcmd.Limit,
				portalName,
				ex.implicitTxn(),
			)
			res = stmtRes

			canAutoCommit := ex.implicitTxn() && func() bool {
				__antithesis_instrumentation__.Notify(457408)
				return tcmd.FollowedBySync == true
			}() == true
			ev, payload, err = ex.execPortal(ctx, portal, portalName, stmtRes, pinfo, canAutoCommit)
			return err
		}()
		__antithesis_instrumentation__.Notify(457377)

		ex.statsCollector.PhaseTimes().SetSessionPhaseTime(sessionphase.SessionQueryServiced, timeutil.Now())
		if err != nil {
			__antithesis_instrumentation__.Notify(457409)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457410)
		}
		__antithesis_instrumentation__.Notify(457378)

		cmd, pos, err = ex.stmtBuf.CurCmd()
		if err != nil {
			__antithesis_instrumentation__.Notify(457411)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457412)
		}

	case PrepareStmt:
		__antithesis_instrumentation__.Notify(457379)
		ex.curStmtAST = tcmd.AST
		res = ex.clientComm.CreatePrepareResult(pos)
		stmtCtx := withStatement(ctx, ex.curStmtAST)
		ev, payload = ex.execPrepare(stmtCtx, tcmd)
	case DescribeStmt:
		__antithesis_instrumentation__.Notify(457380)
		descRes := ex.clientComm.CreateDescribeResult(pos)
		res = descRes
		ev, payload = ex.execDescribe(ctx, tcmd, descRes)
	case BindStmt:
		__antithesis_instrumentation__.Notify(457381)
		res = ex.clientComm.CreateBindResult(pos)
		ev, payload = ex.execBind(ctx, tcmd)
	case DeletePreparedStmt:
		__antithesis_instrumentation__.Notify(457382)
		res = ex.clientComm.CreateDeleteResult(pos)
		ev, payload = ex.execDelPrepStmt(ctx, tcmd)
	case SendError:
		__antithesis_instrumentation__.Notify(457383)
		res = ex.clientComm.CreateErrorResult(pos)
		ev = eventNonRetriableErr{IsCommit: fsm.False}
		payload = eventNonRetriableErrPayload{err: tcmd.Err}
	case Sync:
		__antithesis_instrumentation__.Notify(457384)

		if ex.implicitTxn() {
			__antithesis_instrumentation__.Notify(457413)

			ev, payload = ex.handleAutoCommit(ctx, &tree.CommitTransaction{})
		} else {
			__antithesis_instrumentation__.Notify(457414)
		}
		__antithesis_instrumentation__.Notify(457385)

		res = ex.clientComm.CreateSyncResult(pos)
		if ex.draining {
			__antithesis_instrumentation__.Notify(457415)

			if ex.idleConn() {
				__antithesis_instrumentation__.Notify(457416)

				res.Close(ctx, stateToTxnStatusIndicator(ex.machine.CurState()))
				return errDrainingComplete
			} else {
				__antithesis_instrumentation__.Notify(457417)
			}
		} else {
			__antithesis_instrumentation__.Notify(457418)
		}
	case CopyIn:
		__antithesis_instrumentation__.Notify(457386)
		res = ex.clientComm.CreateCopyInResult(pos)
		var err error
		ev, payload, err = ex.execCopyIn(ctx, tcmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(457419)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457420)
		}
	case DrainRequest:
		__antithesis_instrumentation__.Notify(457387)

		ex.draining = true
		res = ex.clientComm.CreateDrainResult(pos)
		if ex.idleConn() {
			__antithesis_instrumentation__.Notify(457421)
			return errDrainingComplete
		} else {
			__antithesis_instrumentation__.Notify(457422)
		}
	case Flush:
		__antithesis_instrumentation__.Notify(457388)

		res = ex.clientComm.CreateFlushResult(pos)
	default:
		__antithesis_instrumentation__.Notify(457389)
		panic(errors.AssertionFailedf("unsupported command type: %T", cmd))
	}
	__antithesis_instrumentation__.Notify(457362)

	var advInfo advanceInfo

	if ev != nil {
		__antithesis_instrumentation__.Notify(457423)
		var err error
		advInfo, err = ex.txnStateTransitionsApplyWrapper(ev, payload, res, pos)
		if err != nil {
			__antithesis_instrumentation__.Notify(457426)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457427)
		}
		__antithesis_instrumentation__.Notify(457424)

		if _, ok := cmd.(Sync); ok {
			__antithesis_instrumentation__.Notify(457428)
			switch advInfo.code {
			case skipBatch:
				__antithesis_instrumentation__.Notify(457429)

				advInfo = advanceInfo{code: advanceOne}
			case advanceOne:
				__antithesis_instrumentation__.Notify(457430)
			case rewind:
				__antithesis_instrumentation__.Notify(457431)
			case stayInPlace:
				__antithesis_instrumentation__.Notify(457432)
				return errors.AssertionFailedf("unexpected advance code stayInPlace when processing Sync")
			default:
				__antithesis_instrumentation__.Notify(457433)
			}
		} else {
			__antithesis_instrumentation__.Notify(457434)
		}
		__antithesis_instrumentation__.Notify(457425)

		ctx = ex.Ctx()
	} else {
		__antithesis_instrumentation__.Notify(457435)

		advInfo = advanceInfo{code: advanceOne}
	}
	__antithesis_instrumentation__.Notify(457363)

	if advInfo.code != stayInPlace && func() bool {
		__antithesis_instrumentation__.Notify(457436)
		return advInfo.code != rewind == true
	}() == true {
		__antithesis_instrumentation__.Notify(457437)

		resErr := res.Err()

		pe, ok := payload.(payloadWithError)
		if ok {
			__antithesis_instrumentation__.Notify(457439)
			ex.sessionEventf(ctx, "execution error: %s", pe.errorCause())
			if resErr == nil {
				__antithesis_instrumentation__.Notify(457440)
				res.SetError(pe.errorCause())
			} else {
				__antithesis_instrumentation__.Notify(457441)
			}
		} else {
			__antithesis_instrumentation__.Notify(457442)
		}
		__antithesis_instrumentation__.Notify(457438)
		res.Close(ctx, stateToTxnStatusIndicator(ex.machine.CurState()))
	} else {
		__antithesis_instrumentation__.Notify(457443)
		res.Discard()
	}
	__antithesis_instrumentation__.Notify(457364)

	switch advInfo.code {
	case advanceOne:
		__antithesis_instrumentation__.Notify(457444)
		ex.stmtBuf.AdvanceOne()
	case skipBatch:
		__antithesis_instrumentation__.Notify(457445)

		if err := ex.clientComm.Flush(pos); err != nil {
			__antithesis_instrumentation__.Notify(457451)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457452)
		}
		__antithesis_instrumentation__.Notify(457446)
		if err := ex.stmtBuf.seekToNextBatch(); err != nil {
			__antithesis_instrumentation__.Notify(457453)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457454)
		}
	case rewind:
		__antithesis_instrumentation__.Notify(457447)
		if err := ex.rewindPrepStmtNamespace(ctx); err != nil {
			__antithesis_instrumentation__.Notify(457455)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457456)
		}
		__antithesis_instrumentation__.Notify(457448)
		ex.extraTxnState.savepoints = ex.extraTxnState.rewindPosSnapshot.savepoints

		ex.sessionDataStack.Replace(ex.extraTxnState.rewindPosSnapshot.sessionDataStack)
		advInfo.rewCap.rewindAndUnlock(ctx)
	case stayInPlace:
		__antithesis_instrumentation__.Notify(457449)

	default:
		__antithesis_instrumentation__.Notify(457450)
		panic(errors.AssertionFailedf("unexpected advance code: %s", advInfo.code))
	}
	__antithesis_instrumentation__.Notify(457365)

	if err := ex.updateTxnRewindPosMaybe(ctx, cmd, pos, advInfo); err != nil {
		__antithesis_instrumentation__.Notify(457457)
		return err
	} else {
		__antithesis_instrumentation__.Notify(457458)
	}
	__antithesis_instrumentation__.Notify(457366)

	if rewindCapability, canRewind := ex.getRewindTxnCapability(); !canRewind {
		__antithesis_instrumentation__.Notify(457459)

		ex.stmtBuf.Ltrim(ctx, pos)
	} else {
		__antithesis_instrumentation__.Notify(457460)
		rewindCapability.close()
	}
	__antithesis_instrumentation__.Notify(457367)

	if ex.server.cfg.TestingKnobs.AfterExecCmd != nil {
		__antithesis_instrumentation__.Notify(457461)
		ex.server.cfg.TestingKnobs.AfterExecCmd(ctx, cmd, ex.stmtBuf)
	} else {
		__antithesis_instrumentation__.Notify(457462)
	}
	__antithesis_instrumentation__.Notify(457368)

	return nil
}

func (ex *connExecutor) idleConn() bool {
	__antithesis_instrumentation__.Notify(457463)
	switch ex.machine.CurState().(type) {
	case stateNoTxn:
		__antithesis_instrumentation__.Notify(457464)
		return true
	case stateInternalError:
		__antithesis_instrumentation__.Notify(457465)
		return true
	default:
		__antithesis_instrumentation__.Notify(457466)
		return false
	}
}

func (ex *connExecutor) updateTxnRewindPosMaybe(
	ctx context.Context, cmd Command, pos CmdPos, advInfo advanceInfo,
) error {
	__antithesis_instrumentation__.Notify(457467)

	if _, ok := ex.machine.CurState().(stateOpen); !ok {
		__antithesis_instrumentation__.Notify(457470)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(457471)
	}
	__antithesis_instrumentation__.Notify(457468)
	if advInfo.txnEvent.eventType == txnStart || func() bool {
		__antithesis_instrumentation__.Notify(457472)
		return advInfo.txnEvent.eventType == txnRestart == true
	}() == true {
		__antithesis_instrumentation__.Notify(457473)
		var nextPos CmdPos
		switch advInfo.code {
		case stayInPlace:
			__antithesis_instrumentation__.Notify(457475)
			nextPos = pos
		case advanceOne:
			__antithesis_instrumentation__.Notify(457476)

			nextPos = pos + 1
		case rewind:
			__antithesis_instrumentation__.Notify(457477)
			if advInfo.rewCap.rewindPos != ex.extraTxnState.txnRewindPos {
				__antithesis_instrumentation__.Notify(457480)
				return errors.AssertionFailedf(
					"unexpected rewind position: %d when txn start is: %d",
					errors.Safe(advInfo.rewCap.rewindPos),
					errors.Safe(ex.extraTxnState.txnRewindPos))
			} else {
				__antithesis_instrumentation__.Notify(457481)
			}
			__antithesis_instrumentation__.Notify(457478)

			return nil
		default:
			__antithesis_instrumentation__.Notify(457479)
			return errors.AssertionFailedf(
				"unexpected advance code when starting a txn: %s",
				errors.Safe(advInfo.code))
		}
		__antithesis_instrumentation__.Notify(457474)
		if err := ex.setTxnRewindPos(ctx, nextPos); err != nil {
			__antithesis_instrumentation__.Notify(457482)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457483)
		}
	} else {
		__antithesis_instrumentation__.Notify(457484)

		if advInfo.code != advanceOne {
			__antithesis_instrumentation__.Notify(457486)
			panic(errors.AssertionFailedf("unexpected advanceCode: %s", advInfo.code))
		} else {
			__antithesis_instrumentation__.Notify(457487)
		}
		__antithesis_instrumentation__.Notify(457485)

		var canAdvance bool
		_, inOpen := ex.machine.CurState().(stateOpen)
		if inOpen && func() bool {
			__antithesis_instrumentation__.Notify(457488)
			return (ex.extraTxnState.txnRewindPos == pos) == true
		}() == true {
			__antithesis_instrumentation__.Notify(457489)
			switch tcmd := cmd.(type) {
			case ExecStmt:
				__antithesis_instrumentation__.Notify(457491)
				canAdvance = ex.stmtDoesntNeedRetry(tcmd.AST)
			case ExecPortal:
				__antithesis_instrumentation__.Notify(457492)
				canAdvance = true

				portal, ok := ex.extraTxnState.prepStmtsNamespace.portals[tcmd.Name]
				if ok {
					__antithesis_instrumentation__.Notify(457503)
					canAdvance = ex.stmtDoesntNeedRetry(portal.Stmt.AST)
				} else {
					__antithesis_instrumentation__.Notify(457504)
				}
			case PrepareStmt:
				__antithesis_instrumentation__.Notify(457493)
				canAdvance = true
			case DescribeStmt:
				__antithesis_instrumentation__.Notify(457494)
				canAdvance = true
			case BindStmt:
				__antithesis_instrumentation__.Notify(457495)
				canAdvance = true
			case DeletePreparedStmt:
				__antithesis_instrumentation__.Notify(457496)
				canAdvance = true
			case SendError:
				__antithesis_instrumentation__.Notify(457497)
				canAdvance = true
			case Sync:
				__antithesis_instrumentation__.Notify(457498)
				canAdvance = true
			case CopyIn:
				__antithesis_instrumentation__.Notify(457499)

			case DrainRequest:
				__antithesis_instrumentation__.Notify(457500)
				canAdvance = true
			case Flush:
				__antithesis_instrumentation__.Notify(457501)
				canAdvance = true
			default:
				__antithesis_instrumentation__.Notify(457502)
				panic(errors.AssertionFailedf("unsupported cmd: %T", cmd))
			}
			__antithesis_instrumentation__.Notify(457490)
			if canAdvance {
				__antithesis_instrumentation__.Notify(457505)
				if err := ex.setTxnRewindPos(ctx, pos+1); err != nil {
					__antithesis_instrumentation__.Notify(457506)
					return err
				} else {
					__antithesis_instrumentation__.Notify(457507)
				}
			} else {
				__antithesis_instrumentation__.Notify(457508)
			}
		} else {
			__antithesis_instrumentation__.Notify(457509)
		}
	}
	__antithesis_instrumentation__.Notify(457469)
	return nil
}

func (ex *connExecutor) setTxnRewindPos(ctx context.Context, pos CmdPos) error {
	__antithesis_instrumentation__.Notify(457510)
	if pos <= ex.extraTxnState.txnRewindPos {
		__antithesis_instrumentation__.Notify(457512)
		panic(errors.AssertionFailedf("can only move the  txnRewindPos forward. "+
			"Was: %d; new value: %d", ex.extraTxnState.txnRewindPos, pos))
	} else {
		__antithesis_instrumentation__.Notify(457513)
	}
	__antithesis_instrumentation__.Notify(457511)
	ex.extraTxnState.txnRewindPos = pos
	ex.stmtBuf.Ltrim(ctx, pos)
	ex.extraTxnState.rewindPosSnapshot.savepoints = ex.extraTxnState.savepoints.clone()
	ex.extraTxnState.rewindPosSnapshot.sessionDataStack = ex.sessionDataStack.Clone()
	return ex.commitPrepStmtNamespace(ctx)
}

func (ex *connExecutor) stmtDoesntNeedRetry(ast tree.Statement) bool {
	__antithesis_instrumentation__.Notify(457514)
	return isSavepoint(ast) || func() bool {
		__antithesis_instrumentation__.Notify(457515)
		return isSetTransaction(ast) == true
	}() == true
}

func stateToTxnStatusIndicator(s fsm.State) TransactionStatusIndicator {
	__antithesis_instrumentation__.Notify(457516)
	switch s.(type) {
	case stateOpen:
		__antithesis_instrumentation__.Notify(457517)
		return InTxnBlock
	case stateAborted:
		__antithesis_instrumentation__.Notify(457518)
		return InFailedTxnBlock
	case stateNoTxn:
		__antithesis_instrumentation__.Notify(457519)
		return IdleTxnBlock
	case stateCommitWait:
		__antithesis_instrumentation__.Notify(457520)
		return InTxnBlock
	case stateInternalError:
		__antithesis_instrumentation__.Notify(457521)
		return InTxnBlock
	default:
		__antithesis_instrumentation__.Notify(457522)
		panic(errors.AssertionFailedf("unknown state: %T", s))
	}
}

func isCopyToExternalStorage(cmd CopyIn) bool {
	__antithesis_instrumentation__.Notify(457523)
	stmt := cmd.Stmt
	return (stmt.Table.Table() == NodelocalFileUploadTable || func() bool {
		__antithesis_instrumentation__.Notify(457524)
		return stmt.Table.Table() == UserFileUploadTable == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(457525)
		return stmt.Table.SchemaName == CrdbInternalName == true
	}() == true
}

func (ex *connExecutor) execCopyIn(
	ctx context.Context, cmd CopyIn,
) (_ fsm.Event, retPayload fsm.EventPayload, retErr error) {
	__antithesis_instrumentation__.Notify(457526)
	logStatements := logStatementsExecuteEnabled.Get(ex.planner.execCfg.SV())

	ex.incrementStartedStmtCounter(cmd.Stmt)
	defer func() {
		__antithesis_instrumentation__.Notify(457537)
		if retErr == nil && func() bool {
			__antithesis_instrumentation__.Notify(457539)
			return !payloadHasError(retPayload) == true
		}() == true {
			__antithesis_instrumentation__.Notify(457540)
			ex.incrementExecutedStmtCounter(cmd.Stmt)
		} else {
			__antithesis_instrumentation__.Notify(457541)
		}
		__antithesis_instrumentation__.Notify(457538)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(457542)
			log.SqlExec.Errorf(ctx, "error executing %s: %+v", cmd, retErr)
		} else {
			__antithesis_instrumentation__.Notify(457543)
		}
	}()
	__antithesis_instrumentation__.Notify(457527)

	if logStatements {
		__antithesis_instrumentation__.Notify(457544)
		log.SqlExec.Infof(ctx, "executing %s", cmd)
	} else {
		__antithesis_instrumentation__.Notify(457545)
	}
	__antithesis_instrumentation__.Notify(457528)

	defer cmd.CopyDone.Done()

	state := ex.machine.CurState()
	_, isNoTxn := state.(stateNoTxn)
	_, isOpen := state.(stateOpen)
	if !isNoTxn && func() bool {
		__antithesis_instrumentation__.Notify(457546)
		return !isOpen == true
	}() == true {
		__antithesis_instrumentation__.Notify(457547)
		ev := eventNonRetriableErr{IsCommit: fsm.False}
		payload := eventNonRetriableErrPayload{
			err: sqlerrors.NewTransactionAbortedError("")}
		return ev, payload, nil
	} else {
		__antithesis_instrumentation__.Notify(457548)
	}
	__antithesis_instrumentation__.Notify(457529)

	var txnOpt copyTxnOpt
	if isOpen {
		__antithesis_instrumentation__.Notify(457549)
		txnOpt = copyTxnOpt{
			txn:           ex.state.mu.txn,
			txnTimestamp:  ex.state.sqlTimestamp,
			stmtTimestamp: ex.server.cfg.Clock.PhysicalTime(),
		}
	} else {
		__antithesis_instrumentation__.Notify(457550)
		txnOpt = copyTxnOpt{
			resetExtraTxnState: func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(457551)
				return ex.resetExtraTxnState(ctx, txnEvent{eventType: noEvent})
			},
		}
	}
	__antithesis_instrumentation__.Notify(457530)

	var monToStop *mon.BytesMonitor
	defer func() {
		__antithesis_instrumentation__.Notify(457552)
		if monToStop != nil {
			__antithesis_instrumentation__.Notify(457553)
			monToStop.Stop(ctx)
		} else {
			__antithesis_instrumentation__.Notify(457554)
		}
	}()
	__antithesis_instrumentation__.Notify(457531)
	if isNoTxn {
		__antithesis_instrumentation__.Notify(457555)

		ex.state.mon.Start(ctx, ex.sessionMon, mon.BoundAccount{})
		monToStop = ex.state.mon
	} else {
		__antithesis_instrumentation__.Notify(457556)
	}
	__antithesis_instrumentation__.Notify(457532)
	txnOpt.resetPlanner = func(ctx context.Context, p *planner, txn *kv.Txn, txnTS time.Time, stmtTS time.Time) {
		__antithesis_instrumentation__.Notify(457557)

		ex.state.sqlTimestamp = txnTS
		ex.statsCollector.Reset(ex.applicationStats, ex.phaseTimes)
		ex.initPlanner(ctx, p)
		ex.resetPlanner(ctx, p, txn, stmtTS)
	}
	__antithesis_instrumentation__.Notify(457533)
	var cm copyMachineInterface
	var err error
	if isCopyToExternalStorage(cmd) {
		__antithesis_instrumentation__.Notify(457558)
		cm, err = newFileUploadMachine(ctx, cmd.Conn, cmd.Stmt, txnOpt, ex.server.cfg)
	} else {
		__antithesis_instrumentation__.Notify(457559)
		cm, err = newCopyMachine(
			ctx, cmd.Conn, cmd.Stmt, txnOpt, ex.server.cfg,

			func(ctx context.Context, p *planner, res RestrictedCommandResult) error {
				__antithesis_instrumentation__.Notify(457560)
				_, err := ex.execWithDistSQLEngine(ctx, p, tree.RowsAffected, res, DistributionTypeNone, nil)
				return err
			},
		)
	}
	__antithesis_instrumentation__.Notify(457534)
	if err != nil {
		__antithesis_instrumentation__.Notify(457561)
		ev := eventNonRetriableErr{IsCommit: fsm.False}
		payload := eventNonRetriableErrPayload{err: err}
		return ev, payload, nil
	} else {
		__antithesis_instrumentation__.Notify(457562)
	}
	__antithesis_instrumentation__.Notify(457535)
	if err := cm.run(ctx); err != nil {
		__antithesis_instrumentation__.Notify(457563)

		ev := eventNonRetriableErr{IsCommit: fsm.False}
		payload := eventNonRetriableErrPayload{err: err}
		return ev, payload, nil
	} else {
		__antithesis_instrumentation__.Notify(457564)
	}
	__antithesis_instrumentation__.Notify(457536)
	return nil, nil, nil
}

func stmtHasNoData(stmt tree.Statement) bool {
	__antithesis_instrumentation__.Notify(457565)
	return stmt == nil || func() bool {
		__antithesis_instrumentation__.Notify(457566)
		return stmt.StatementReturnType() != tree.Rows == true
	}() == true
}

func (ex *connExecutor) generateID() ClusterWideID {
	__antithesis_instrumentation__.Notify(457567)
	return GenerateClusterWideID(ex.server.cfg.Clock.Now(), ex.server.cfg.NodeID.SQLInstanceID())
}

func (ex *connExecutor) commitPrepStmtNamespace(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(457568)
	return ex.extraTxnState.prepStmtsNamespaceAtTxnRewindPos.resetTo(
		ctx, ex.extraTxnState.prepStmtsNamespace, &ex.extraTxnState.prepStmtsNamespaceMemAcc,
	)
}

func (ex *connExecutor) rewindPrepStmtNamespace(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(457569)
	return ex.extraTxnState.prepStmtsNamespace.resetTo(
		ctx, ex.extraTxnState.prepStmtsNamespaceAtTxnRewindPos, &ex.extraTxnState.prepStmtsNamespaceMemAcc,
	)
}

func (ex *connExecutor) getRewindTxnCapability() (rewindCapability, bool) {
	__antithesis_instrumentation__.Notify(457570)
	cl := ex.clientComm.LockCommunication()

	if cl.ClientPos() >= ex.extraTxnState.txnRewindPos {
		__antithesis_instrumentation__.Notify(457572)
		cl.Close()
		return rewindCapability{}, false
	} else {
		__antithesis_instrumentation__.Notify(457573)
	}
	__antithesis_instrumentation__.Notify(457571)
	return rewindCapability{
		cl:        cl,
		buf:       ex.stmtBuf,
		rewindPos: ex.extraTxnState.txnRewindPos,
	}, true
}

func isCommit(stmt tree.Statement) bool {
	__antithesis_instrumentation__.Notify(457574)
	_, ok := stmt.(*tree.CommitTransaction)
	return ok
}

var retriableMinTimestampBoundUnsatisfiableError = errors.Newf(
	"retriable MinTimestampBoundUnsatisfiableError",
)

func errIsRetriable(err error) bool {
	__antithesis_instrumentation__.Notify(457575)
	return errors.HasType(err, (*roachpb.TransactionRetryWithProtoRefreshError)(nil)) || func() bool {
		__antithesis_instrumentation__.Notify(457576)
		return scerrors.ConcurrentSchemaChangeDescID(err) != descpb.InvalidID == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(457577)
		return errors.Is(err, retriableMinTimestampBoundUnsatisfiableError) == true
	}() == true
}

func (ex *connExecutor) makeErrEvent(err error, stmt tree.Statement) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(457578)

	if minTSErr := (*roachpb.MinTimestampBoundUnsatisfiableError)(nil); errors.As(err, &minTSErr) {
		__antithesis_instrumentation__.Notify(457581)
		aost := ex.planner.EvalContext().AsOfSystemTime
		if aost != nil && func() bool {
			__antithesis_instrumentation__.Notify(457582)
			return aost.BoundedStaleness == true
		}() == true {
			__antithesis_instrumentation__.Notify(457583)
			if !aost.MaxTimestampBound.IsEmpty() && func() bool {
				__antithesis_instrumentation__.Notify(457585)
				return aost.MaxTimestampBound.LessEq(minTSErr.MinTimestampBound) == true
			}() == true {
				__antithesis_instrumentation__.Notify(457586)

				err = errors.CombineErrors(
					errors.AssertionFailedf(
						"unexpected MaxTimestampBound >= txn MinTimestampBound: %s >= %s",
						aost.MaxTimestampBound,
						minTSErr.MinTimestampBound,
					),
					err,
				)
			} else {
				__antithesis_instrumentation__.Notify(457587)
			}
			__antithesis_instrumentation__.Notify(457584)
			if aost.Timestamp.Less(minTSErr.MinTimestampBound) {
				__antithesis_instrumentation__.Notify(457588)
				err = errors.Mark(err, retriableMinTimestampBoundUnsatisfiableError)
			} else {
				__antithesis_instrumentation__.Notify(457589)
			}
		} else {
			__antithesis_instrumentation__.Notify(457590)
		}
	} else {
		__antithesis_instrumentation__.Notify(457591)
	}
	__antithesis_instrumentation__.Notify(457579)

	retriable := errIsRetriable(err)
	if retriable {
		__antithesis_instrumentation__.Notify(457592)
		var rc rewindCapability
		var canAutoRetry bool
		if ex.implicitTxn() || func() bool {
			__antithesis_instrumentation__.Notify(457595)
			return !ex.sessionData().InjectRetryErrorsEnabled == true
		}() == true {
			__antithesis_instrumentation__.Notify(457596)
			rc, canAutoRetry = ex.getRewindTxnCapability()
		} else {
			__antithesis_instrumentation__.Notify(457597)
		}
		__antithesis_instrumentation__.Notify(457593)
		if canAutoRetry {
			__antithesis_instrumentation__.Notify(457598)
			ex.extraTxnState.autoRetryReason = err
		} else {
			__antithesis_instrumentation__.Notify(457599)
		}
		__antithesis_instrumentation__.Notify(457594)

		ev := eventRetriableErr{
			IsCommit:     fsm.FromBool(isCommit(stmt)),
			CanAutoRetry: fsm.FromBool(canAutoRetry),
		}
		payload := eventRetriableErrPayload{
			err:    err,
			rewCap: rc,
		}
		return ev, payload
	} else {
		__antithesis_instrumentation__.Notify(457600)
	}
	__antithesis_instrumentation__.Notify(457580)
	ev := eventNonRetriableErr{
		IsCommit: fsm.FromBool(isCommit(stmt)),
	}
	payload := eventNonRetriableErrPayload{err: err}
	return ev, payload
}

func (ex *connExecutor) setTransactionModes(
	ctx context.Context, modes tree.TransactionModes, asOfTs hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(457601)

	if modes.UserPriority != tree.UnspecifiedUserPriority {
		__antithesis_instrumentation__.Notify(457606)
		pri := txnPriorityToProto(modes.UserPriority)
		if err := ex.state.setPriority(pri); err != nil {
			__antithesis_instrumentation__.Notify(457607)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457608)
		}
	} else {
		__antithesis_instrumentation__.Notify(457609)
	}
	__antithesis_instrumentation__.Notify(457602)
	if modes.Isolation != tree.UnspecifiedIsolation && func() bool {
		__antithesis_instrumentation__.Notify(457610)
		return modes.Isolation != tree.SerializableIsolation == true
	}() == true {
		__antithesis_instrumentation__.Notify(457611)
		return errors.AssertionFailedf(
			"unknown isolation level: %s", errors.Safe(modes.Isolation))
	} else {
		__antithesis_instrumentation__.Notify(457612)
	}
	__antithesis_instrumentation__.Notify(457603)
	rwMode := modes.ReadWriteMode
	if modes.AsOf.Expr != nil && func() bool {
		__antithesis_instrumentation__.Notify(457613)
		return asOfTs.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(457614)
		return errors.AssertionFailedf("expected an evaluated AS OF timestamp")
	} else {
		__antithesis_instrumentation__.Notify(457615)
	}
	__antithesis_instrumentation__.Notify(457604)
	if !asOfTs.IsEmpty() {
		__antithesis_instrumentation__.Notify(457616)
		if err := ex.state.setHistoricalTimestamp(ctx, asOfTs); err != nil {
			__antithesis_instrumentation__.Notify(457618)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457619)
		}
		__antithesis_instrumentation__.Notify(457617)
		if rwMode == tree.UnspecifiedReadWriteMode {
			__antithesis_instrumentation__.Notify(457620)
			rwMode = tree.ReadOnly
		} else {
			__antithesis_instrumentation__.Notify(457621)
		}
	} else {
		__antithesis_instrumentation__.Notify(457622)
	}
	__antithesis_instrumentation__.Notify(457605)
	return ex.state.setReadOnlyMode(rwMode)
}

func txnPriorityToProto(mode tree.UserPriority) roachpb.UserPriority {
	__antithesis_instrumentation__.Notify(457623)
	var pri roachpb.UserPriority
	switch mode {
	case tree.UnspecifiedUserPriority:
		__antithesis_instrumentation__.Notify(457625)
		pri = roachpb.NormalUserPriority
	case tree.Low:
		__antithesis_instrumentation__.Notify(457626)
		pri = roachpb.MinUserPriority
	case tree.Normal:
		__antithesis_instrumentation__.Notify(457627)
		pri = roachpb.NormalUserPriority
	case tree.High:
		__antithesis_instrumentation__.Notify(457628)
		pri = roachpb.MaxUserPriority
	default:
		__antithesis_instrumentation__.Notify(457629)
		log.Fatalf(context.Background(), "unknown user priority: %s", mode)
	}
	__antithesis_instrumentation__.Notify(457624)
	return pri
}

func (ex *connExecutor) txnPriorityWithSessionDefault(mode tree.UserPriority) roachpb.UserPriority {
	__antithesis_instrumentation__.Notify(457630)
	if mode == tree.UnspecifiedUserPriority {
		__antithesis_instrumentation__.Notify(457632)
		mode = tree.UserPriority(ex.sessionData().DefaultTxnPriority)
	} else {
		__antithesis_instrumentation__.Notify(457633)
	}
	__antithesis_instrumentation__.Notify(457631)
	return txnPriorityToProto(mode)
}

func (ex *connExecutor) QualityOfService() sessiondatapb.QoSLevel {
	__antithesis_instrumentation__.Notify(457634)
	if ex.sessionData() == nil {
		__antithesis_instrumentation__.Notify(457636)
		return sessiondatapb.Normal
	} else {
		__antithesis_instrumentation__.Notify(457637)
	}
	__antithesis_instrumentation__.Notify(457635)
	return ex.sessionData().DefaultTxnQualityOfService
}

func (ex *connExecutor) readWriteModeWithSessionDefault(
	mode tree.ReadWriteMode,
) tree.ReadWriteMode {
	__antithesis_instrumentation__.Notify(457638)
	if mode == tree.UnspecifiedReadWriteMode {
		__antithesis_instrumentation__.Notify(457640)
		if ex.sessionData().DefaultTxnReadOnly {
			__antithesis_instrumentation__.Notify(457642)
			return tree.ReadOnly
		} else {
			__antithesis_instrumentation__.Notify(457643)
		}
		__antithesis_instrumentation__.Notify(457641)
		return tree.ReadWrite
	} else {
		__antithesis_instrumentation__.Notify(457644)
	}
	__antithesis_instrumentation__.Notify(457639)
	return mode
}

var followerReadTimestampExpr = &tree.FuncExpr{
	Func: tree.WrapFunction(tree.FollowerReadTimestampFunctionName),
}

func (ex *connExecutor) asOfClauseWithSessionDefault(expr tree.AsOfClause) tree.AsOfClause {
	__antithesis_instrumentation__.Notify(457645)
	if expr.Expr == nil {
		__antithesis_instrumentation__.Notify(457647)
		if ex.sessionData().DefaultTxnUseFollowerReads {
			__antithesis_instrumentation__.Notify(457649)
			return tree.AsOfClause{Expr: followerReadTimestampExpr}
		} else {
			__antithesis_instrumentation__.Notify(457650)
		}
		__antithesis_instrumentation__.Notify(457648)
		return tree.AsOfClause{}
	} else {
		__antithesis_instrumentation__.Notify(457651)
	}
	__antithesis_instrumentation__.Notify(457646)
	return expr
}

func (ex *connExecutor) initEvalCtx(ctx context.Context, evalCtx *extendedEvalContext, p *planner) {
	__antithesis_instrumentation__.Notify(457652)
	*evalCtx = extendedEvalContext{
		EvalContext: tree.EvalContext{
			Planner:                   p,
			PrivilegedAccessor:        p,
			SessionAccessor:           p,
			JobExecContext:            p,
			ClientNoticeSender:        p,
			Sequence:                  p,
			Tenant:                    p,
			Regions:                   p,
			JoinTokenCreator:          p,
			PreparedStatementState:    &ex.extraTxnState.prepStmtsNamespace,
			SessionDataStack:          ex.sessionDataStack,
			ReCache:                   ex.server.reCache,
			SQLStatsController:        ex.server.sqlStatsController,
			IndexUsageStatsController: ex.server.indexUsageStatsController,
		},
		Tracing:                &ex.sessionTracing,
		MemMetrics:             &ex.memMetrics,
		Descs:                  &ex.extraTxnState.descCollection,
		TxnModesSetter:         ex,
		Jobs:                   &ex.extraTxnState.jobs,
		SchemaChangeJobRecords: ex.extraTxnState.schemaChangeJobRecords,
		statsProvider:          ex.server.sqlStats,
		indexUsageStats:        ex.indexUsageStats,
		statementPreparer:      ex,
	}
	evalCtx.copyFromExecCfg(ex.server.cfg)
}

func (ex *connExecutor) resetEvalCtx(evalCtx *extendedEvalContext, txn *kv.Txn, stmtTS time.Time) {
	__antithesis_instrumentation__.Notify(457653)
	newTxn := txn == nil || func() bool {
		__antithesis_instrumentation__.Notify(457655)
		return evalCtx.Txn != txn == true
	}() == true
	evalCtx.TxnState = ex.getTransactionState()
	evalCtx.TxnReadOnly = ex.state.readOnly
	evalCtx.TxnImplicit = ex.implicitTxn()
	evalCtx.TxnIsSingleStmt = false
	if newTxn || func() bool {
		__antithesis_instrumentation__.Notify(457656)
		return !ex.implicitTxn() == true
	}() == true {
		__antithesis_instrumentation__.Notify(457657)

		evalCtx.StmtTimestamp = stmtTS
	} else {
		__antithesis_instrumentation__.Notify(457658)
	}
	__antithesis_instrumentation__.Notify(457654)
	evalCtx.TxnTimestamp = ex.state.sqlTimestamp
	evalCtx.Placeholders = nil
	evalCtx.Annotations = nil
	evalCtx.IVarContainer = nil
	evalCtx.Context = ex.Ctx()
	evalCtx.Txn = txn
	evalCtx.Mon = ex.state.mon
	evalCtx.PrepareOnly = false
	evalCtx.SkipNormalize = false
	evalCtx.SchemaChangerState = &ex.extraTxnState.schemaChangerState

	var minTSErr *roachpb.MinTimestampBoundUnsatisfiableError
	if err := ex.extraTxnState.autoRetryReason; err != nil && func() bool {
		__antithesis_instrumentation__.Notify(457659)
		return errors.As(err, &minTSErr) == true
	}() == true {
		__antithesis_instrumentation__.Notify(457660)
		nextMax := minTSErr.MinTimestampBound
		ex.extraTxnState.descCollection.SetMaxTimestampBound(nextMax)
		evalCtx.AsOfSystemTime.MaxTimestampBound = nextMax
	} else {
		__antithesis_instrumentation__.Notify(457661)
		if newTxn {
			__antithesis_instrumentation__.Notify(457662)

			ex.extraTxnState.descCollection.ResetMaxTimestampBound()
			evalCtx.AsOfSystemTime = nil
		} else {
			__antithesis_instrumentation__.Notify(457663)
		}
	}
}

func (ex *connExecutor) getTransactionState() string {
	__antithesis_instrumentation__.Notify(457664)
	state := ex.machine.CurState()
	if ex.implicitTxn() {
		__antithesis_instrumentation__.Notify(457666)

		state = stateNoTxn{}
	} else {
		__antithesis_instrumentation__.Notify(457667)
	}
	__antithesis_instrumentation__.Notify(457665)
	return state.(fmt.Stringer).String()
}

func (ex *connExecutor) implicitTxn() bool {
	__antithesis_instrumentation__.Notify(457668)
	state := ex.machine.CurState()
	os, ok := state.(stateOpen)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(457669)
		return os.ImplicitTxn.Get() == true
	}() == true
}

func (ex *connExecutor) initPlanner(ctx context.Context, p *planner) {
	__antithesis_instrumentation__.Notify(457670)
	p.cancelChecker.Reset(ctx)

	ex.initEvalCtx(ctx, &p.extendedEvalCtx, p)

	p.sessionDataMutatorIterator = ex.dataMutatorIterator
	p.noticeSender = nil
	p.preparedStatements = ex.getPrepStmtsAccessor()
	p.sqlCursors = ex.getCursorAccessor()
	p.createdSequences = ex.getCreatedSequencesAccessor()

	p.queryCacheSession.Init()
	p.optPlanningCtx.init(p)
}

func (ex *connExecutor) resetPlanner(
	ctx context.Context, p *planner, txn *kv.Txn, stmtTS time.Time,
) {
	__antithesis_instrumentation__.Notify(457671)
	p.txn = txn
	p.stmt = Statement{}
	p.instrumentation = instrumentationHelper{}

	p.cancelChecker.Reset(ctx)

	p.semaCtx = tree.MakeSemaContext()
	if p.execCfg.Settings.Version.IsActive(ctx, clusterversion.DateStyleIntervalStyleCastRewrite) {
		__antithesis_instrumentation__.Notify(457673)
		p.semaCtx.IntervalStyleEnabled = true
		p.semaCtx.DateStyleEnabled = true
	} else {
		__antithesis_instrumentation__.Notify(457674)
		p.semaCtx.IntervalStyleEnabled = ex.sessionData().IntervalStyleEnabled
		p.semaCtx.DateStyleEnabled = ex.sessionData().DateStyleEnabled
	}
	__antithesis_instrumentation__.Notify(457672)
	p.semaCtx.SearchPath = ex.sessionData().SearchPath
	p.semaCtx.Annotations = nil
	p.semaCtx.TypeResolver = p
	p.semaCtx.TableNameResolver = p
	p.semaCtx.DateStyle = ex.sessionData().GetDateStyle()
	p.semaCtx.IntervalStyle = ex.sessionData().GetIntervalStyle()

	ex.resetEvalCtx(&p.extendedEvalCtx, txn, stmtTS)

	p.autoCommit = false
	p.isPreparing = false
	p.avoidLeasedDescriptors = false
}

func (ex *connExecutor) txnStateTransitionsApplyWrapper(
	ev fsm.Event, payload fsm.EventPayload, res ResultBase, pos CmdPos,
) (advanceInfo, error) {
	__antithesis_instrumentation__.Notify(457675)

	var implicitTxn bool
	txnIsOpen := false
	if os, ok := ex.machine.CurState().(stateOpen); ok {
		__antithesis_instrumentation__.Notify(457681)
		implicitTxn = os.ImplicitTxn.Get()
		txnIsOpen = true
	} else {
		__antithesis_instrumentation__.Notify(457682)
	}
	__antithesis_instrumentation__.Notify(457676)

	ex.mu.Lock()
	err := ex.machine.ApplyWithPayload(withStatement(ex.Ctx(), ex.curStmtAST), ev, payload)
	ex.mu.Unlock()
	if err != nil {
		__antithesis_instrumentation__.Notify(457683)
		if errors.HasType(err, (*fsm.TransitionNotFoundError)(nil)) {
			__antithesis_instrumentation__.Notify(457685)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(457686)
		}
		__antithesis_instrumentation__.Notify(457684)
		return advanceInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(457687)
	}
	__antithesis_instrumentation__.Notify(457677)

	advInfo := ex.state.consumeAdvanceInfo()
	if advInfo.code == rewind {
		__antithesis_instrumentation__.Notify(457688)
		atomic.AddInt32(ex.extraTxnState.atomicAutoRetryCounter, 1)
	} else {
		__antithesis_instrumentation__.Notify(457689)
	}
	__antithesis_instrumentation__.Notify(457678)

	if p, ok := payload.(payloadWithError); ok {
		__antithesis_instrumentation__.Notify(457690)
		if descID := scerrors.ConcurrentSchemaChangeDescID(p.errorCause()); descID != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(457691)
			if err := ex.handleWaitingForConcurrentSchemaChanges(ex.Ctx(), descID); err != nil {
				__antithesis_instrumentation__.Notify(457692)
				return advanceInfo{}, err
			} else {
				__antithesis_instrumentation__.Notify(457693)
			}
		} else {
			__antithesis_instrumentation__.Notify(457694)
		}
	} else {
		__antithesis_instrumentation__.Notify(457695)
	}
	__antithesis_instrumentation__.Notify(457679)

	switch advInfo.txnEvent.eventType {
	case noEvent:
		__antithesis_instrumentation__.Notify(457696)
		_, nextStateIsAborted := ex.machine.CurState().(stateAborted)

		if txnIsOpen && func() bool {
			__antithesis_instrumentation__.Notify(457704)
			return !nextStateIsAborted == true
		}() == true {
			__antithesis_instrumentation__.Notify(457705)
			err := ex.extraTxnState.descCollection.MaybeUpdateDeadline(ex.Ctx(), ex.state.mu.txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(457706)
				return advanceInfo{}, err
			} else {
				__antithesis_instrumentation__.Notify(457707)
			}
		} else {
			__antithesis_instrumentation__.Notify(457708)
		}
	case txnStart:
		__antithesis_instrumentation__.Notify(457697)
		atomic.StoreInt32(ex.extraTxnState.atomicAutoRetryCounter, 0)
		ex.extraTxnState.autoRetryReason = nil
		ex.extraTxnState.firstStmtExecuted = false
		ex.recordTransactionStart(advInfo.txnEvent.txnID)

		ex.extraTxnState.txnCounter++
		if !ex.server.cfg.Codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(457709)

			session, err := ex.server.cfg.SQLLiveness.Session(ex.Ctx())
			if err != nil {
				__antithesis_instrumentation__.Notify(457711)
				return advanceInfo{}, err
			} else {
				__antithesis_instrumentation__.Notify(457712)
			}
			__antithesis_instrumentation__.Notify(457710)
			ex.extraTxnState.descCollection.SetSession(session)
		} else {
			__antithesis_instrumentation__.Notify(457713)
		}
	case txnCommit:
		__antithesis_instrumentation__.Notify(457698)
		if res.Err() != nil {
			__antithesis_instrumentation__.Notify(457714)

			err := errorutil.UnexpectedWithIssueErrorf(
				26687,
				"programming error: non-error event %s generated even though res.Err() has been set to: %s",
				errors.Safe(advInfo.txnEvent.eventType.String()),
				res.Err())
			log.Errorf(ex.Ctx(), "%v", err)
			errorutil.SendReport(ex.Ctx(), &ex.server.cfg.Settings.SV, err)
			return advanceInfo{}, err
		} else {
			__antithesis_instrumentation__.Notify(457715)
		}
		__antithesis_instrumentation__.Notify(457699)

		handleErr := func(err error) {
			__antithesis_instrumentation__.Notify(457716)
			if implicitTxn {
				__antithesis_instrumentation__.Notify(457717)

				res.SetError(err)
			} else {
				__antithesis_instrumentation__.Notify(457718)

				newErr := pgerror.Wrapf(err,
					pgcode.TransactionCommittedWithSchemaChangeFailure,
					"transaction committed but schema change aborted with error: (%s)",
					pgerror.GetPGCode(err))
				newErr = errors.WithHint(newErr,
					"Some of the non-DDL statements may have committed successfully, "+
						"but some of the DDL statement(s) failed.\nManual inspection may be "+
						"required to determine the actual state of the database.")
				newErr = errors.WithIssueLink(newErr,
					errors.IssueLink{IssueURL: "https://github.com/cockroachdb/cockroach/issues/42061"})
				res.SetError(newErr)
			}
		}
		__antithesis_instrumentation__.Notify(457700)
		ex.notifyStatsRefresherOfNewTables(ex.Ctx())

		ex.statsCollector.PhaseTimes().SetSessionPhaseTime(sessionphase.SessionStartPostCommitJob, timeutil.Now())
		if err := ex.server.cfg.JobRegistry.Run(
			ex.ctxHolder.connCtx,
			ex.server.cfg.InternalExecutor,
			ex.extraTxnState.jobs,
		); err != nil {
			__antithesis_instrumentation__.Notify(457719)
			handleErr(err)
		} else {
			__antithesis_instrumentation__.Notify(457720)
		}
		__antithesis_instrumentation__.Notify(457701)
		ex.statsCollector.PhaseTimes().SetSessionPhaseTime(sessionphase.SessionEndPostCommitJob, timeutil.Now())

		fallthrough
	case txnRestart, txnRollback:
		__antithesis_instrumentation__.Notify(457702)
		if err := ex.resetExtraTxnState(ex.Ctx(), advInfo.txnEvent); err != nil {
			__antithesis_instrumentation__.Notify(457721)
			return advanceInfo{}, err
		} else {
			__antithesis_instrumentation__.Notify(457722)
		}
	default:
		__antithesis_instrumentation__.Notify(457703)
		return advanceInfo{}, errors.AssertionFailedf(
			"unexpected event: %v", errors.Safe(advInfo.txnEvent))
	}
	__antithesis_instrumentation__.Notify(457680)
	return advInfo, nil
}

func (ex *connExecutor) handleWaitingForConcurrentSchemaChanges(
	ctx context.Context, descID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(457723)
	if err := ex.planner.waitForDescriptorSchemaChanges(
		ctx, descID, ex.extraTxnState.schemaChangerState,
	); err != nil {
		__antithesis_instrumentation__.Notify(457725)
		return err
	} else {
		__antithesis_instrumentation__.Notify(457726)
	}
	__antithesis_instrumentation__.Notify(457724)
	return ex.resetTransactionOnSchemaChangeRetry(ctx)
}

func (ex *connExecutor) initStatementResult(
	ctx context.Context, res RestrictedCommandResult, ast tree.Statement, cols colinfo.ResultColumns,
) error {
	__antithesis_instrumentation__.Notify(457727)
	for _, c := range cols {
		__antithesis_instrumentation__.Notify(457730)
		if err := checkResultType(c.Typ); err != nil {
			__antithesis_instrumentation__.Notify(457731)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457732)
		}
	}
	__antithesis_instrumentation__.Notify(457728)
	if ast.StatementReturnType() == tree.Rows {
		__antithesis_instrumentation__.Notify(457733)

		res.SetColumns(ctx, cols)
	} else {
		__antithesis_instrumentation__.Notify(457734)
	}
	__antithesis_instrumentation__.Notify(457729)
	return nil
}

func (ex *connExecutor) cancelQuery(queryID ClusterWideID) bool {
	__antithesis_instrumentation__.Notify(457735)
	ex.mu.Lock()
	defer ex.mu.Unlock()
	if queryMeta, exists := ex.mu.ActiveQueries[queryID]; exists {
		__antithesis_instrumentation__.Notify(457737)
		queryMeta.cancel()
		return true
	} else {
		__antithesis_instrumentation__.Notify(457738)
	}
	__antithesis_instrumentation__.Notify(457736)
	return false
}

func (ex *connExecutor) cancelCurrentQueries() bool {
	__antithesis_instrumentation__.Notify(457739)
	ex.mu.Lock()
	defer ex.mu.Unlock()
	canceled := false
	for _, queryMeta := range ex.mu.ActiveQueries {
		__antithesis_instrumentation__.Notify(457741)
		queryMeta.cancel()
		canceled = true
	}
	__antithesis_instrumentation__.Notify(457740)
	return canceled
}

func (ex *connExecutor) cancelSession() {
	__antithesis_instrumentation__.Notify(457742)
	if ex.onCancelSession == nil {
		__antithesis_instrumentation__.Notify(457744)
		return
	} else {
		__antithesis_instrumentation__.Notify(457745)
	}
	__antithesis_instrumentation__.Notify(457743)

	ex.onCancelSession()
}

func (ex *connExecutor) user() security.SQLUsername {
	__antithesis_instrumentation__.Notify(457746)
	return ex.sessionData().User()
}

func (ex *connExecutor) serialize() serverpb.Session {
	__antithesis_instrumentation__.Notify(457747)
	ex.mu.RLock()
	defer ex.mu.RUnlock()
	ex.state.mu.RLock()
	defer ex.state.mu.RUnlock()

	var activeTxnInfo *serverpb.TxnInfo
	txn := ex.state.mu.txn
	if txn != nil {
		__antithesis_instrumentation__.Notify(457753)
		id := txn.ID()
		activeTxnInfo = &serverpb.TxnInfo{
			ID:                    id,
			Start:                 ex.state.mu.txnStart,
			NumStatementsExecuted: int32(ex.state.mu.stmtCount),
			NumRetries:            int32(txn.Epoch()),
			NumAutoRetries:        atomic.LoadInt32(ex.extraTxnState.atomicAutoRetryCounter),
			TxnDescription:        txn.String(),
			Implicit:              ex.implicitTxn(),
			AllocBytes:            ex.state.mon.AllocBytes(),
			MaxAllocBytes:         ex.state.mon.MaximumBytes(),
			IsHistorical:          ex.state.isHistorical,
			ReadOnly:              ex.state.readOnly,
			Priority:              ex.state.priority.String(),
			QualityOfService:      sessiondatapb.ToQoSLevelString(txn.AdmissionHeader().Priority),
		}
	} else {
		__antithesis_instrumentation__.Notify(457754)
	}
	__antithesis_instrumentation__.Notify(457748)

	activeQueries := make([]serverpb.ActiveQuery, 0, len(ex.mu.ActiveQueries))
	truncateSQL := func(sql string) string {
		__antithesis_instrumentation__.Notify(457755)
		if len(sql) > MaxSQLBytes {
			__antithesis_instrumentation__.Notify(457757)
			sql = sql[:MaxSQLBytes-utf8.RuneLen('…')]

			for {
				__antithesis_instrumentation__.Notify(457759)
				if r, _ := utf8.DecodeLastRuneInString(sql); r != utf8.RuneError {
					__antithesis_instrumentation__.Notify(457761)
					break
				} else {
					__antithesis_instrumentation__.Notify(457762)
				}
				__antithesis_instrumentation__.Notify(457760)
				sql = sql[:len(sql)-1]
			}
			__antithesis_instrumentation__.Notify(457758)
			sql += "…"
		} else {
			__antithesis_instrumentation__.Notify(457763)
		}
		__antithesis_instrumentation__.Notify(457756)
		return sql
	}
	__antithesis_instrumentation__.Notify(457749)

	for id, query := range ex.mu.ActiveQueries {
		__antithesis_instrumentation__.Notify(457764)
		if query.hidden {
			__antithesis_instrumentation__.Notify(457767)
			continue
		} else {
			__antithesis_instrumentation__.Notify(457768)
		}
		__antithesis_instrumentation__.Notify(457765)
		ast, err := query.getStatement()
		if err != nil {
			__antithesis_instrumentation__.Notify(457769)
			continue
		} else {
			__antithesis_instrumentation__.Notify(457770)
		}
		__antithesis_instrumentation__.Notify(457766)
		sqlNoConstants := truncateSQL(formatStatementHideConstants(ast))
		sql := truncateSQL(ast.String())
		progress := math.Float64frombits(atomic.LoadUint64(&query.progressAtomic))
		activeQueries = append(activeQueries, serverpb.ActiveQuery{
			TxnID:          query.txnID,
			ID:             id.String(),
			Start:          query.start.UTC(),
			Sql:            sql,
			SqlNoConstants: sqlNoConstants,
			SqlSummary:     formatStatementSummary(ast),
			IsDistributed:  query.isDistributed,
			Phase:          (serverpb.ActiveQuery_Phase)(query.phase),
			Progress:       float32(progress),
		})
	}
	__antithesis_instrumentation__.Notify(457750)
	lastActiveQuery := ""
	lastActiveQueryNoConstants := ""
	if ex.mu.LastActiveQuery != nil {
		__antithesis_instrumentation__.Notify(457771)
		lastActiveQuery = truncateSQL(ex.mu.LastActiveQuery.String())
		lastActiveQueryNoConstants = truncateSQL(formatStatementHideConstants(ex.mu.LastActiveQuery))
	} else {
		__antithesis_instrumentation__.Notify(457772)
	}
	__antithesis_instrumentation__.Notify(457751)

	sd := ex.sessionDataStack.Base()

	remoteStr := "<admin>"
	if sd.RemoteAddr != nil {
		__antithesis_instrumentation__.Notify(457773)
		remoteStr = sd.RemoteAddr.String()
	} else {
		__antithesis_instrumentation__.Notify(457774)
	}
	__antithesis_instrumentation__.Notify(457752)

	return serverpb.Session{
		Username:                   sd.SessionUser().Normalized(),
		ClientAddress:              remoteStr,
		ApplicationName:            ex.applicationName.Load().(string),
		Start:                      ex.phaseTimes.GetSessionPhaseTime(sessionphase.SessionInit).UTC(),
		ActiveQueries:              activeQueries,
		ActiveTxn:                  activeTxnInfo,
		LastActiveQuery:            lastActiveQuery,
		ID:                         ex.sessionID.GetBytes(),
		AllocBytes:                 ex.mon.AllocBytes(),
		MaxAllocBytes:              ex.mon.MaximumBytes(),
		LastActiveQueryNoConstants: lastActiveQueryNoConstants,
	}
}

func (ex *connExecutor) getPrepStmtsAccessor() preparedStatementsAccessor {
	__antithesis_instrumentation__.Notify(457775)
	return connExPrepStmtsAccessor{
		ex: ex,
	}
}

func (ex *connExecutor) getCursorAccessor() sqlCursors {
	__antithesis_instrumentation__.Notify(457776)
	return connExCursorAccessor{
		ex: ex,
	}
}

func (ex *connExecutor) getCreatedSequencesAccessor() createdSequences {
	__antithesis_instrumentation__.Notify(457777)
	return connExCreatedSequencesAccessor{
		ex: ex,
	}
}

func (ex *connExecutor) sessionEventf(ctx context.Context, format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(457778)
	if log.ExpensiveLogEnabled(ctx, 2) {
		__antithesis_instrumentation__.Notify(457780)
		log.VEventfDepth(ctx, 1, 2, format, args...)
	} else {
		__antithesis_instrumentation__.Notify(457781)
	}
	__antithesis_instrumentation__.Notify(457779)
	if ex.eventLog != nil {
		__antithesis_instrumentation__.Notify(457782)
		ex.eventLog.Printf(format, args...)
	} else {
		__antithesis_instrumentation__.Notify(457783)
	}
}

func (ex *connExecutor) notifyStatsRefresherOfNewTables(ctx context.Context) {
	__antithesis_instrumentation__.Notify(457784)
	for _, desc := range ex.extraTxnState.descCollection.GetUncommittedTables() {
		__antithesis_instrumentation__.Notify(457785)

		if desc.IsTable() && func() bool {
			__antithesis_instrumentation__.Notify(457786)
			return !desc.IsAs() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(457787)
			return desc.GetVersion() == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(457788)

			ex.planner.execCfg.StatsRefresher.NotifyMutation(desc, math.MaxInt32)
		} else {
			__antithesis_instrumentation__.Notify(457789)
		}
	}
}

func (ex *connExecutor) runPreCommitStages(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(457790)
	scs := &ex.extraTxnState.schemaChangerState
	deps := newSchemaChangerTxnRunDependencies(
		ex.planner.SessionData(),
		ex.planner.User(),
		ex.server.cfg,
		ex.planner.txn,
		&ex.extraTxnState.descCollection,
		ex.planner.EvalContext(),
		ex.planner.ExtendedEvalContext().Tracing.KVTracingEnabled(),
		scs.jobID,
		scs.stmts,
	)

	after, jobID, err := scrun.RunPreCommitPhase(
		ctx, ex.server.cfg.DeclarativeSchemaChangerTestingKnobs, deps, scs.state,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(457793)
		return err
	} else {
		__antithesis_instrumentation__.Notify(457794)
	}
	__antithesis_instrumentation__.Notify(457791)
	scs.state = after
	scs.jobID = jobID
	if jobID != jobspb.InvalidJobID {
		__antithesis_instrumentation__.Notify(457795)
		ex.extraTxnState.jobs.add(jobID)
		log.Infof(ctx, "queued new schema change job %d using the new schema changer", jobID)
	} else {
		__antithesis_instrumentation__.Notify(457796)
	}
	__antithesis_instrumentation__.Notify(457792)
	return nil
}

type StatementCounters struct {
	QueryCount telemetry.CounterWithMetric

	SelectCount telemetry.CounterWithMetric
	UpdateCount telemetry.CounterWithMetric
	InsertCount telemetry.CounterWithMetric
	DeleteCount telemetry.CounterWithMetric

	TxnBeginCount    telemetry.CounterWithMetric
	TxnCommitCount   telemetry.CounterWithMetric
	TxnRollbackCount telemetry.CounterWithMetric

	SavepointCount                  telemetry.CounterWithMetric
	ReleaseSavepointCount           telemetry.CounterWithMetric
	RollbackToSavepointCount        telemetry.CounterWithMetric
	RestartSavepointCount           telemetry.CounterWithMetric
	ReleaseRestartSavepointCount    telemetry.CounterWithMetric
	RollbackToRestartSavepointCount telemetry.CounterWithMetric

	CopyCount telemetry.CounterWithMetric

	DdlCount telemetry.CounterWithMetric

	MiscCount telemetry.CounterWithMetric
}

func makeStartedStatementCounters(internal bool) StatementCounters {
	__antithesis_instrumentation__.Notify(457797)
	return StatementCounters{
		TxnBeginCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaTxnBeginStarted, internal)),
		TxnCommitCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaTxnCommitStarted, internal)),
		TxnRollbackCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaTxnRollbackStarted, internal)),
		RestartSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaRestartSavepointStarted, internal)),
		ReleaseRestartSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaReleaseRestartSavepointStarted, internal)),
		RollbackToRestartSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaRollbackToRestartSavepointStarted, internal)),
		SavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaSavepointStarted, internal)),
		ReleaseSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaReleaseSavepointStarted, internal)),
		RollbackToSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaRollbackToSavepointStarted, internal)),
		SelectCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaSelectStarted, internal)),
		UpdateCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaUpdateStarted, internal)),
		InsertCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaInsertStarted, internal)),
		DeleteCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaDeleteStarted, internal)),
		DdlCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaDdlStarted, internal)),
		CopyCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaCopyStarted, internal)),
		MiscCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaMiscStarted, internal)),
		QueryCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaQueryStarted, internal)),
	}
}

func makeExecutedStatementCounters(internal bool) StatementCounters {
	__antithesis_instrumentation__.Notify(457798)
	return StatementCounters{
		TxnBeginCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaTxnBeginExecuted, internal)),
		TxnCommitCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaTxnCommitExecuted, internal)),
		TxnRollbackCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaTxnRollbackExecuted, internal)),
		RestartSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaRestartSavepointExecuted, internal)),
		ReleaseRestartSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaReleaseRestartSavepointExecuted, internal)),
		RollbackToRestartSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaRollbackToRestartSavepointExecuted, internal)),
		SavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaSavepointExecuted, internal)),
		ReleaseSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaReleaseSavepointExecuted, internal)),
		RollbackToSavepointCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaRollbackToSavepointExecuted, internal)),
		SelectCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaSelectExecuted, internal)),
		UpdateCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaUpdateExecuted, internal)),
		InsertCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaInsertExecuted, internal)),
		DeleteCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaDeleteExecuted, internal)),
		DdlCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaDdlExecuted, internal)),
		CopyCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaCopyExecuted, internal)),
		MiscCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaMiscExecuted, internal)),
		QueryCount: telemetry.NewCounterWithMetric(
			getMetricMeta(MetaQueryExecuted, internal)),
	}
}

func (sc *StatementCounters) incrementCount(ex *connExecutor, stmt tree.Statement) {
	__antithesis_instrumentation__.Notify(457799)
	sc.QueryCount.Inc()
	switch t := stmt.(type) {
	case *tree.BeginTransaction:
		__antithesis_instrumentation__.Notify(457800)
		sc.TxnBeginCount.Inc()
	case *tree.Select:
		__antithesis_instrumentation__.Notify(457801)
		sc.SelectCount.Inc()
	case *tree.Update:
		__antithesis_instrumentation__.Notify(457802)
		sc.UpdateCount.Inc()
	case *tree.Insert:
		__antithesis_instrumentation__.Notify(457803)
		sc.InsertCount.Inc()
	case *tree.Delete:
		__antithesis_instrumentation__.Notify(457804)
		sc.DeleteCount.Inc()
	case *tree.CommitTransaction:
		__antithesis_instrumentation__.Notify(457805)
		sc.TxnCommitCount.Inc()
	case *tree.RollbackTransaction:
		__antithesis_instrumentation__.Notify(457806)

		if ex.getTransactionState() == CommitWaitStateStr {
			__antithesis_instrumentation__.Notify(457812)
			sc.TxnCommitCount.Inc()
		} else {
			__antithesis_instrumentation__.Notify(457813)
			sc.TxnRollbackCount.Inc()
		}
	case *tree.Savepoint:
		__antithesis_instrumentation__.Notify(457807)
		if ex.isCommitOnReleaseSavepoint(t.Name) {
			__antithesis_instrumentation__.Notify(457814)
			sc.RestartSavepointCount.Inc()
		} else {
			__antithesis_instrumentation__.Notify(457815)
			sc.SavepointCount.Inc()
		}
	case *tree.ReleaseSavepoint:
		__antithesis_instrumentation__.Notify(457808)
		if ex.isCommitOnReleaseSavepoint(t.Savepoint) {
			__antithesis_instrumentation__.Notify(457816)
			sc.ReleaseRestartSavepointCount.Inc()
		} else {
			__antithesis_instrumentation__.Notify(457817)
			sc.ReleaseSavepointCount.Inc()
		}
	case *tree.RollbackToSavepoint:
		__antithesis_instrumentation__.Notify(457809)
		if ex.isCommitOnReleaseSavepoint(t.Savepoint) {
			__antithesis_instrumentation__.Notify(457818)
			sc.RollbackToRestartSavepointCount.Inc()
		} else {
			__antithesis_instrumentation__.Notify(457819)
			sc.RollbackToSavepointCount.Inc()
		}
	case *tree.CopyFrom:
		__antithesis_instrumentation__.Notify(457810)
		sc.CopyCount.Inc()
	default:
		__antithesis_instrumentation__.Notify(457811)
		if tree.CanModifySchema(stmt) {
			__antithesis_instrumentation__.Notify(457820)
			sc.DdlCount.Inc()
		} else {
			__antithesis_instrumentation__.Notify(457821)
			sc.MiscCount.Inc()
		}
	}
}

type connExPrepStmtsAccessor struct {
	ex *connExecutor
}

var _ preparedStatementsAccessor = connExPrepStmtsAccessor{}

func (ps connExPrepStmtsAccessor) List() map[string]*PreparedStatement {
	__antithesis_instrumentation__.Notify(457822)

	stmts := ps.ex.extraTxnState.prepStmtsNamespace.prepStmts
	ret := make(map[string]*PreparedStatement, len(stmts))
	for key, stmt := range stmts {
		__antithesis_instrumentation__.Notify(457824)
		ret[key] = stmt
	}
	__antithesis_instrumentation__.Notify(457823)
	return ret
}

func (ps connExPrepStmtsAccessor) Get(name string) (*PreparedStatement, bool) {
	__antithesis_instrumentation__.Notify(457825)
	s, ok := ps.ex.extraTxnState.prepStmtsNamespace.prepStmts[name]
	return s, ok
}

func (ps connExPrepStmtsAccessor) Delete(ctx context.Context, name string) bool {
	__antithesis_instrumentation__.Notify(457826)
	_, ok := ps.Get(name)
	if !ok {
		__antithesis_instrumentation__.Notify(457828)
		return false
	} else {
		__antithesis_instrumentation__.Notify(457829)
	}
	__antithesis_instrumentation__.Notify(457827)
	ps.ex.deletePreparedStmt(ctx, name)
	return true
}

func (ps connExPrepStmtsAccessor) DeleteAll(ctx context.Context) {
	__antithesis_instrumentation__.Notify(457830)
	ps.ex.extraTxnState.prepStmtsNamespace.resetToEmpty(
		ctx, &ps.ex.extraTxnState.prepStmtsNamespaceMemAcc,
	)
}

type contextStatementKey struct{}

func withStatement(ctx context.Context, stmt tree.Statement) context.Context {
	__antithesis_instrumentation__.Notify(457831)
	return context.WithValue(ctx, contextStatementKey{}, stmt)
}

func statementFromCtx(ctx context.Context) tree.Statement {
	__antithesis_instrumentation__.Notify(457832)
	stmt := ctx.Value(contextStatementKey{})
	if stmt == nil {
		__antithesis_instrumentation__.Notify(457834)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(457835)
	}
	__antithesis_instrumentation__.Notify(457833)
	return stmt.(tree.Statement)
}

func init() {

	logcrash.RegisterTagFn("statement", func(ctx context.Context) string {
		stmt := statementFromCtx(ctx)
		if stmt == nil {
			return ""
		}

		return anonymizeStmtAndConstants(stmt, nil)
	})
}
