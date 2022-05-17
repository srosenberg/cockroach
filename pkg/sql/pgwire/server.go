package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/hba"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/identmap"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirecancel"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var connResultsBufferSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"sql.defaults.results_buffer.size",
	"default size of the buffer that accumulates results for a statement or a batch "+
		"of statements before they are sent to the client. This can be overridden on "+
		"an individual connection with the 'results_buffer_size' parameter. Note that auto-retries "+
		"generally only happen while no results have been delivered to the client, so "+
		"reducing this size can increase the number of retriable errors a client "+
		"receives. On the other hand, increasing the buffer size can increase the "+
		"delay until the client receives the first result row. "+
		"Updating the setting only affects new connections. "+
		"Setting to 0 disables any buffering.",
	16<<10,
).WithPublic()

var logConnAuth = settings.RegisterBoolSetting(
	settings.TenantWritable,
	sql.ConnAuditingClusterSettingName,
	"if set, log SQL client connect and disconnect events (note: may hinder performance on loaded nodes)",
	false).WithPublic()

var logSessionAuth = settings.RegisterBoolSetting(
	settings.TenantWritable,
	sql.AuthAuditingClusterSettingName,
	"if set, log SQL session login/disconnection events (note: may hinder performance on loaded nodes)",
	false).WithPublic()

var maxNumConnections = settings.RegisterIntSetting(
	settings.TenantWritable,
	"server.max_connections_per_gateway",
	"the maximum number of non-superuser SQL connections per gateway allowed at a given time "+
		"(note: this will only limit future connection attempts and will not affect already established connections). "+
		"Negative values result in unlimited number of connections. Superusers are not affected by this limit.",
	-1,
).WithPublic()

const (
	ErrSSLRequired = "node is running secure mode, SSL connection required"

	ErrDrainingNewConn = "server is not accepting clients, try another node"

	ErrDrainingExistingConn = "server is shutting down"
)

var (
	MetaConns = metric.Metadata{
		Name:        "sql.conns",
		Help:        "Number of active sql connections",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	MetaNewConns = metric.Metadata{
		Name:        "sql.new_conns",
		Help:        "Counter of the number of sql connections created",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	MetaBytesIn = metric.Metadata{
		Name:        "sql.bytesin",
		Help:        "Number of sql bytes received",
		Measurement: "SQL Bytes",
		Unit:        metric.Unit_BYTES,
	}
	MetaBytesOut = metric.Metadata{
		Name:        "sql.bytesout",
		Help:        "Number of sql bytes sent",
		Measurement: "SQL Bytes",
		Unit:        metric.Unit_BYTES,
	}
	MetaConnLatency = metric.Metadata{
		Name:        "sql.conn.latency",
		Help:        "Latency to establish and authenticate a SQL connection",
		Measurement: "Nanoseconds",
		Unit:        metric.Unit_NANOSECONDS,
	}
	MetaPGWireCancelTotal = metric.Metadata{
		Name:        "sql.pgwire_cancel.total",
		Help:        "Counter of the number of pgwire query cancel requests",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	MetaPGWireCancelIgnored = metric.Metadata{
		Name:        "sql.pgwire_cancel.ignored",
		Help:        "Counter of the number of pgwire query cancel requests that were ignored due to rate limiting",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	MetaPGWireCancelSuccessful = metric.Metadata{
		Name:        "sql.pgwire_cancel.successful",
		Help:        "Counter of the number of pgwire query cancel requests that were successful",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
)

const (
	version30     = 196608
	versionCancel = 80877102
	versionSSL    = 80877103
	versionGSSENC = 80877104
)

const cancelMaxWait = 1 * time.Second

var baseSQLMemoryBudget = envutil.EnvOrDefaultInt64("COCKROACH_BASE_SQL_MEMORY_BUDGET",
	int64(2.1*float64(mon.DefaultPoolAllocationSize)))

var connReservationBatchSize = 5

var (
	sslSupported   = []byte{'S'}
	sslUnsupported = []byte{'N'}
)

type cancelChanMap map[chan struct{}]context.CancelFunc

type Server struct {
	AmbientCtx log.AmbientContext
	cfg        *base.Config
	SQLServer  *sql.Server
	execCfg    *sql.ExecutorConfig

	metrics ServerMetrics

	mu struct {
		syncutil.Mutex

		connCancelMap cancelChanMap

		draining bool

		rejectNewConnections bool
	}

	auth struct {
		syncutil.RWMutex
		conf        *hba.Conf
		identityMap *identmap.Conf
	}

	sqlMemoryPool *mon.BytesMonitor
	connMonitor   *mon.BytesMonitor

	testingConnLogEnabled int32
	testingAuthLogEnabled int32

	trustClientProvidedRemoteAddr syncutil.AtomicBool
}

type ServerMetrics struct {
	BytesInCount                *metric.Counter
	BytesOutCount               *metric.Counter
	Conns                       *metric.Gauge
	NewConns                    *metric.Counter
	ConnLatency                 *metric.Histogram
	PGWireCancelTotalCount      *metric.Counter
	PGWireCancelIgnoredCount    *metric.Counter
	PGWireCancelSuccessfulCount *metric.Counter
	ConnMemMetrics              sql.BaseMemoryMetrics
	SQLMemMetrics               sql.MemoryMetrics
}

func makeServerMetrics(
	sqlMemMetrics sql.MemoryMetrics, histogramWindow time.Duration,
) ServerMetrics {
	__antithesis_instrumentation__.Notify(561469)
	return ServerMetrics{
		BytesInCount:                metric.NewCounter(MetaBytesIn),
		BytesOutCount:               metric.NewCounter(MetaBytesOut),
		Conns:                       metric.NewGauge(MetaConns),
		NewConns:                    metric.NewCounter(MetaNewConns),
		ConnLatency:                 metric.NewLatency(MetaConnLatency, histogramWindow),
		PGWireCancelTotalCount:      metric.NewCounter(MetaPGWireCancelTotal),
		PGWireCancelIgnoredCount:    metric.NewCounter(MetaPGWireCancelIgnored),
		PGWireCancelSuccessfulCount: metric.NewCounter(MetaPGWireCancelSuccessful),
		ConnMemMetrics:              sql.MakeBaseMemMetrics("conns", histogramWindow),
		SQLMemMetrics:               sqlMemMetrics,
	}
}

var noteworthySQLMemoryUsageBytes = envutil.EnvOrDefaultInt64("COCKROACH_NOTEWORTHY_SQL_MEMORY_USAGE", 100*1024*1024)

var noteworthyConnMemoryUsageBytes = envutil.EnvOrDefaultInt64("COCKROACH_NOTEWORTHY_CONN_MEMORY_USAGE", 2*1024*1024)

func MakeServer(
	ambientCtx log.AmbientContext,
	cfg *base.Config,
	st *cluster.Settings,
	sqlMemMetrics sql.MemoryMetrics,
	parentMemoryMonitor *mon.BytesMonitor,
	histogramWindow time.Duration,
	executorConfig *sql.ExecutorConfig,
) *Server {
	__antithesis_instrumentation__.Notify(561470)
	server := &Server{
		AmbientCtx: ambientCtx,
		cfg:        cfg,
		execCfg:    executorConfig,
		metrics:    makeServerMetrics(sqlMemMetrics, histogramWindow),
	}
	server.sqlMemoryPool = mon.NewMonitor("sql",
		mon.MemoryResource,

		nil,
		nil,
		0, noteworthySQLMemoryUsageBytes, st)
	server.sqlMemoryPool.Start(context.Background(), parentMemoryMonitor, mon.BoundAccount{})
	server.SQLServer = sql.NewServer(executorConfig, server.sqlMemoryPool)

	server.trustClientProvidedRemoteAddr.Set(trustClientProvidedRemoteAddrOverride)

	server.connMonitor = mon.NewMonitor("conn",
		mon.MemoryResource,
		server.metrics.ConnMemMetrics.CurBytesCount,
		server.metrics.ConnMemMetrics.MaxBytesHist,
		int64(connReservationBatchSize)*baseSQLMemoryBudget, noteworthyConnMemoryUsageBytes, st)
	server.connMonitor.Start(context.Background(), server.sqlMemoryPool, mon.BoundAccount{})

	server.mu.Lock()
	server.mu.connCancelMap = make(cancelChanMap)
	server.mu.Unlock()

	connAuthConf.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(561473)
		loadLocalHBAConfigUponRemoteSettingChange(
			ambientCtx.AnnotateCtx(context.Background()), server, st)
	})
	__antithesis_instrumentation__.Notify(561471)
	connIdentityMapConf.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(561474)
		loadLocalIdentityMapUponRemoteSettingChange(
			ambientCtx.AnnotateCtx(context.Background()), server, st)
	})
	__antithesis_instrumentation__.Notify(561472)

	return server
}

func (s *Server) BytesOut() uint64 {
	__antithesis_instrumentation__.Notify(561475)
	return uint64(s.metrics.BytesOutCount.Count())
}

func (s *Server) AnnotateCtxForIncomingConn(ctx context.Context, conn net.Conn) context.Context {
	__antithesis_instrumentation__.Notify(561476)
	tag := "client"
	if s.trustClientProvidedRemoteAddr.Get() {
		__antithesis_instrumentation__.Notify(561478)
		tag = "peer"
	} else {
		__antithesis_instrumentation__.Notify(561479)
	}
	__antithesis_instrumentation__.Notify(561477)
	return logtags.AddTag(ctx, tag, conn.RemoteAddr().String())
}

func Match(rd io.Reader) bool {
	__antithesis_instrumentation__.Notify(561480)
	buf := pgwirebase.MakeReadBuffer()
	_, err := buf.ReadUntypedMsg(rd)
	if err != nil {
		__antithesis_instrumentation__.Notify(561483)
		return false
	} else {
		__antithesis_instrumentation__.Notify(561484)
	}
	__antithesis_instrumentation__.Notify(561481)
	version, err := buf.GetUint32()
	if err != nil {
		__antithesis_instrumentation__.Notify(561485)
		return false
	} else {
		__antithesis_instrumentation__.Notify(561486)
	}
	__antithesis_instrumentation__.Notify(561482)
	return version == version30 || func() bool {
		__antithesis_instrumentation__.Notify(561487)
		return version == versionSSL == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(561488)
		return version == versionCancel == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(561489)
		return version == versionGSSENC == true
	}() == true
}

func (s *Server) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(561490)
	s.SQLServer.Start(ctx, stopper)
}

func (s *Server) IsDraining() bool {
	__antithesis_instrumentation__.Notify(561491)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.draining
}

func (s *Server) Metrics() (res []interface{}) {
	__antithesis_instrumentation__.Notify(561492)
	return []interface{}{
		&s.metrics,
		&s.SQLServer.Metrics.StartedStatementCounters,
		&s.SQLServer.Metrics.ExecutedStatementCounters,
		&s.SQLServer.Metrics.EngineMetrics,
		&s.SQLServer.Metrics.GuardrailMetrics,
		&s.SQLServer.InternalMetrics.StartedStatementCounters,
		&s.SQLServer.InternalMetrics.ExecutedStatementCounters,
		&s.SQLServer.InternalMetrics.EngineMetrics,
		&s.SQLServer.InternalMetrics.GuardrailMetrics,
		&s.SQLServer.ServerMetrics.StatsMetrics,
		&s.SQLServer.ServerMetrics.ContentionSubsystemMetrics,
	}
}

func (s *Server) Drain(
	ctx context.Context,
	queryWait time.Duration,
	reporter func(int, redact.SafeString),
	stopper *stop.Stopper,
) error {
	__antithesis_instrumentation__.Notify(561493)
	return s.drainImpl(ctx, queryWait, cancelMaxWait, reporter, stopper)
}

func (s *Server) Undrain() {
	__antithesis_instrumentation__.Notify(561494)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setRejectNewConnectionsLocked(false)
	s.setDrainingLocked(false)
}

func (s *Server) setDrainingLocked(drain bool) bool {
	__antithesis_instrumentation__.Notify(561495)
	if s.mu.draining == drain {
		__antithesis_instrumentation__.Notify(561497)
		return false
	} else {
		__antithesis_instrumentation__.Notify(561498)
	}
	__antithesis_instrumentation__.Notify(561496)
	s.mu.draining = drain
	return true
}

func (s *Server) setRejectNewConnectionsLocked(rej bool) {
	__antithesis_instrumentation__.Notify(561499)
	s.mu.rejectNewConnections = rej
}

func (s *Server) GetConnCancelMapLen() int {
	__antithesis_instrumentation__.Notify(561500)
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.mu.connCancelMap)
}

func (s *Server) WaitForSQLConnsToClose(
	ctx context.Context, connectionWait time.Duration, stopper *stop.Stopper,
) error {
	__antithesis_instrumentation__.Notify(561501)

	if s.IsDraining() {
		__antithesis_instrumentation__.Notify(561505)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(561506)
	}
	__antithesis_instrumentation__.Notify(561502)

	s.mu.Lock()
	s.setRejectNewConnectionsLocked(true)
	s.mu.Unlock()

	if connectionWait == 0 {
		__antithesis_instrumentation__.Notify(561507)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(561508)
	}
	__antithesis_instrumentation__.Notify(561503)

	log.Ops.Info(ctx, "waiting for clients to close existing SQL connections")

	timer := time.NewTimer(connectionWait)
	defer timer.Stop()

	_, allConnsDone, quitWaitingForConns := s.waitConnsDone()
	defer close(quitWaitingForConns)

	select {

	case <-time.After(connectionWait):
		__antithesis_instrumentation__.Notify(561509)
		log.Ops.Warningf(ctx,
			"%d connections remain after waiting %s; proceeding to drain SQL connections",
			s.GetConnCancelMapLen(),
			connectionWait,
		)
	case <-allConnsDone:
		__antithesis_instrumentation__.Notify(561510)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(561511)
		return ctx.Err()
	case <-stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(561512)
		return context.Canceled
	}
	__antithesis_instrumentation__.Notify(561504)

	return nil
}

func (s *Server) waitConnsDone() (cancelChanMap, chan struct{}, chan struct{}) {
	__antithesis_instrumentation__.Notify(561513)
	connCancelMap := func() cancelChanMap {
		__antithesis_instrumentation__.Notify(561516)
		s.mu.Lock()
		defer s.mu.Unlock()
		connCancelMap := make(cancelChanMap)
		for done, cancel := range s.mu.connCancelMap {
			__antithesis_instrumentation__.Notify(561518)
			connCancelMap[done] = cancel
		}
		__antithesis_instrumentation__.Notify(561517)
		return connCancelMap
	}()
	__antithesis_instrumentation__.Notify(561514)

	allConnsDone := make(chan struct{}, 1)

	quitWaitingForConns := make(chan struct{}, 1)

	go func() {
		__antithesis_instrumentation__.Notify(561519)
		defer close(allConnsDone)

		for done := range connCancelMap {
			__antithesis_instrumentation__.Notify(561520)
			select {
			case <-done:
				__antithesis_instrumentation__.Notify(561521)
			case <-quitWaitingForConns:
				__antithesis_instrumentation__.Notify(561522)
				return
			}
		}
	}()
	__antithesis_instrumentation__.Notify(561515)

	return connCancelMap, allConnsDone, quitWaitingForConns
}

func (s *Server) drainImpl(
	ctx context.Context,
	queryWait time.Duration,
	cancelWait time.Duration,
	reporter func(int, redact.SafeString),
	stopper *stop.Stopper,
) error {
	__antithesis_instrumentation__.Notify(561523)

	s.mu.Lock()
	if !s.setDrainingLocked(true) {
		__antithesis_instrumentation__.Notify(561530)

		s.mu.Unlock()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(561531)
	}
	__antithesis_instrumentation__.Notify(561524)
	s.mu.Unlock()

	if s.GetConnCancelMapLen() == 0 {
		__antithesis_instrumentation__.Notify(561532)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(561533)
	}
	__antithesis_instrumentation__.Notify(561525)

	log.Ops.Info(ctx, "starting draining SQL connections")

	connCancelMap, allConnsDone, quitWaitingForConns := s.waitConnsDone()
	defer close(quitWaitingForConns)

	if reporter != nil {
		__antithesis_instrumentation__.Notify(561534)

		reporter(len(connCancelMap), "SQL clients")
	} else {
		__antithesis_instrumentation__.Notify(561535)
	}
	__antithesis_instrumentation__.Notify(561526)

	select {
	case <-time.After(queryWait):
		__antithesis_instrumentation__.Notify(561536)
		log.Ops.Warningf(ctx, "canceling all sessions after waiting %s", queryWait)
	case <-allConnsDone:
		__antithesis_instrumentation__.Notify(561537)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(561538)
		return ctx.Err()
	case <-stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(561539)
		return context.Canceled
	}
	__antithesis_instrumentation__.Notify(561527)

	if stop := func() bool {
		__antithesis_instrumentation__.Notify(561540)
		s.mu.Lock()
		defer s.mu.Unlock()
		if !s.mu.draining {
			__antithesis_instrumentation__.Notify(561543)
			return true
		} else {
			__antithesis_instrumentation__.Notify(561544)
		}
		__antithesis_instrumentation__.Notify(561541)
		for _, cancel := range connCancelMap {
			__antithesis_instrumentation__.Notify(561545)

			cancel()
		}
		__antithesis_instrumentation__.Notify(561542)
		return false
	}(); stop {
		__antithesis_instrumentation__.Notify(561546)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(561547)
	}
	__antithesis_instrumentation__.Notify(561528)

	select {
	case <-time.After(cancelWait):
		__antithesis_instrumentation__.Notify(561548)
		return errors.Errorf("some sessions did not respond to cancellation within %s", cancelWait)
	case <-allConnsDone:
		__antithesis_instrumentation__.Notify(561549)
	}
	__antithesis_instrumentation__.Notify(561529)
	return nil
}

type SocketType bool

const (
	SocketTCP SocketType = true

	SocketUnix SocketType = false
)

func (s SocketType) asConnType() (hba.ConnType, error) {
	__antithesis_instrumentation__.Notify(561550)
	switch s {
	case SocketTCP:
		__antithesis_instrumentation__.Notify(561551)
		return hba.ConnHostNoSSL, nil
	case SocketUnix:
		__antithesis_instrumentation__.Notify(561552)
		return hba.ConnLocal, nil
	default:
		__antithesis_instrumentation__.Notify(561553)
		return 0, errors.AssertionFailedf("unimplemented socket type: %v", errors.Safe(s))
	}
}

func (s *Server) connLogEnabled() bool {
	__antithesis_instrumentation__.Notify(561554)
	return atomic.LoadInt32(&s.testingConnLogEnabled) != 0 || func() bool {
		__antithesis_instrumentation__.Notify(561555)
		return logConnAuth.Get(&s.execCfg.Settings.SV) == true
	}() == true
}

func (s *Server) TestingEnableConnLogging() {
	__antithesis_instrumentation__.Notify(561556)
	atomic.StoreInt32(&s.testingConnLogEnabled, 1)
}

func (s *Server) TestingEnableAuthLogging() {
	__antithesis_instrumentation__.Notify(561557)
	atomic.StoreInt32(&s.testingAuthLogEnabled, 1)
}

func (s *Server) ServeConn(ctx context.Context, conn net.Conn, socketType SocketType) error {
	__antithesis_instrumentation__.Notify(561558)
	ctx, rejectNewConnections, onCloseFn := s.registerConn(ctx)
	defer onCloseFn()

	connDetails := eventpb.CommonConnectionDetails{
		InstanceID:    int32(s.execCfg.NodeID.SQLInstanceID()),
		Network:       conn.RemoteAddr().Network(),
		RemoteAddress: conn.RemoteAddr().String(),
	}

	connStart := timeutil.Now()
	if s.connLogEnabled() {
		__antithesis_instrumentation__.Notify(561571)
		ev := &eventpb.ClientConnectionStart{
			CommonEventDetails:      eventpb.CommonEventDetails{Timestamp: connStart.UnixNano()},
			CommonConnectionDetails: connDetails,
		}
		log.StructuredEvent(ctx, ev)
	} else {
		__antithesis_instrumentation__.Notify(561572)
	}
	__antithesis_instrumentation__.Notify(561559)
	defer func() {
		__antithesis_instrumentation__.Notify(561573)

		if s.connLogEnabled() {
			__antithesis_instrumentation__.Notify(561574)
			endTime := timeutil.Now()
			ev := &eventpb.ClientConnectionEnd{
				CommonEventDetails:      eventpb.CommonEventDetails{Timestamp: endTime.UnixNano()},
				CommonConnectionDetails: connDetails,
				Duration:                endTime.Sub(connStart).Nanoseconds(),
			}
			log.StructuredEvent(ctx, ev)
		} else {
			__antithesis_instrumentation__.Notify(561575)
		}
	}()
	__antithesis_instrumentation__.Notify(561560)

	version, buf, err := s.readVersion(conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(561576)
		return err
	} else {
		__antithesis_instrumentation__.Notify(561577)
	}
	__antithesis_instrumentation__.Notify(561561)

	switch version {
	case versionCancel:
		__antithesis_instrumentation__.Notify(561578)

		s.handleCancel(ctx, conn, &buf)
		return nil

	case versionGSSENC:
		__antithesis_instrumentation__.Notify(561579)

		err := pgerror.New(pgcode.ProtocolViolation, "GSS encryption is not yet supported")

		err = errors.WithTelemetry(err, "#52184")
		return s.sendErr(ctx, conn, err)
	default:
		__antithesis_instrumentation__.Notify(561580)
	}
	__antithesis_instrumentation__.Notify(561562)

	if rejectNewConnections {
		__antithesis_instrumentation__.Notify(561581)
		log.Ops.Info(ctx, "rejecting new connection while server is draining")
		return s.sendErr(ctx, conn, newAdminShutdownErr(ErrDrainingNewConn))
	} else {
		__antithesis_instrumentation__.Notify(561582)
	}
	__antithesis_instrumentation__.Notify(561563)

	connType, err := socketType.asConnType()
	if err != nil {
		__antithesis_instrumentation__.Notify(561583)
		return err
	} else {
		__antithesis_instrumentation__.Notify(561584)
	}
	__antithesis_instrumentation__.Notify(561564)

	var clientErr error
	conn, connType, version, clientErr, err = s.maybeUpgradeToSecureConn(ctx, conn, connType, version, &buf)
	if err != nil {
		__antithesis_instrumentation__.Notify(561585)
		return err
	} else {
		__antithesis_instrumentation__.Notify(561586)
	}
	__antithesis_instrumentation__.Notify(561565)
	if clientErr != nil {
		__antithesis_instrumentation__.Notify(561587)
		return s.sendErr(ctx, conn, clientErr)
	} else {
		__antithesis_instrumentation__.Notify(561588)
	}
	__antithesis_instrumentation__.Notify(561566)
	sp := tracing.SpanFromContext(ctx)
	sp.SetTag("conn_type", attribute.StringValue(connType.String()))

	switch version {
	case version30:
		__antithesis_instrumentation__.Notify(561589)

	case versionCancel:
		__antithesis_instrumentation__.Notify(561590)

		s.handleCancel(ctx, conn, &buf)
		return nil

	default:
		__antithesis_instrumentation__.Notify(561591)

		err := pgerror.Newf(pgcode.ProtocolViolation, "unknown protocol version %d", version)
		err = errors.WithTelemetry(err, fmt.Sprintf("protocol-version-%d", version))
		return s.sendErr(ctx, conn, err)
	}
	__antithesis_instrumentation__.Notify(561567)

	reserved := s.connMonitor.MakeBoundAccount()
	if err := reserved.Grow(ctx, baseSQLMemoryBudget); err != nil {
		__antithesis_instrumentation__.Notify(561592)
		return errors.Wrapf(err, "unable to pre-allocate %d bytes for this connection",
			baseSQLMemoryBudget)
	} else {
		__antithesis_instrumentation__.Notify(561593)
	}
	__antithesis_instrumentation__.Notify(561568)

	var sArgs sql.SessionArgs
	if sArgs, err = parseClientProvidedSessionParameters(ctx, &s.execCfg.Settings.SV, &buf,
		conn.RemoteAddr(), s.trustClientProvidedRemoteAddr.Get()); err != nil {
		__antithesis_instrumentation__.Notify(561594)
		return s.sendErr(ctx, conn, err)
	} else {
		__antithesis_instrumentation__.Notify(561595)
	}
	__antithesis_instrumentation__.Notify(561569)

	connDetails.RemoteAddress = sArgs.RemoteAddr.String()
	ctx = logtags.AddTag(ctx, "client", log.SafeOperational(connDetails.RemoteAddress))
	sp.SetTag("client", attribute.StringValue(connDetails.RemoteAddress))

	var testingAuthHook func(context.Context) error
	if k := s.execCfg.PGWireTestingKnobs; k != nil {
		__antithesis_instrumentation__.Notify(561596)
		testingAuthHook = k.AuthHook
	} else {
		__antithesis_instrumentation__.Notify(561597)
	}
	__antithesis_instrumentation__.Notify(561570)

	hbaConf, identMap := s.GetAuthenticationConfiguration()

	s.serveConn(
		ctx, conn, sArgs,
		reserved,
		connStart,
		authOptions{
			connType:        connType,
			connDetails:     connDetails,
			insecure:        s.cfg.Insecure,
			ie:              s.execCfg.InternalExecutor,
			auth:            hbaConf,
			identMap:        identMap,
			testingAuthHook: testingAuthHook,
		},
	)
	return nil
}

func (s *Server) handleCancel(ctx context.Context, conn net.Conn, buf *pgwirebase.ReadBuffer) {
	__antithesis_instrumentation__.Notify(561598)
	telemetry.Inc(sqltelemetry.CancelRequestCounter)
	s.metrics.PGWireCancelTotalCount.Inc(1)

	resp, err := func() (*serverpb.CancelQueryByKeyResponse, error) {
		__antithesis_instrumentation__.Notify(561600)
		backendKeyDataBits, err := buf.GetUint64()

		_ = conn.Close()
		if err != nil {
			__antithesis_instrumentation__.Notify(561603)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561604)
		}
		__antithesis_instrumentation__.Notify(561601)
		cancelKey := pgwirecancel.BackendKeyData(backendKeyDataBits)

		req := &serverpb.CancelQueryByKeyRequest{
			SQLInstanceID:  cancelKey.GetSQLInstanceID(),
			CancelQueryKey: cancelKey,
		}
		resp, err := s.execCfg.SQLStatusServer.CancelQueryByKey(ctx, req)
		if resp != nil && func() bool {
			__antithesis_instrumentation__.Notify(561605)
			return len(resp.Error) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(561606)
			err = errors.CombineErrors(err, errors.Newf("error from CancelQueryByKeyResponse: %s", resp.Error))
		} else {
			__antithesis_instrumentation__.Notify(561607)
		}
		__antithesis_instrumentation__.Notify(561602)
		return resp, err
	}()
	__antithesis_instrumentation__.Notify(561599)

	if resp != nil && func() bool {
		__antithesis_instrumentation__.Notify(561608)
		return resp.Canceled == true
	}() == true {
		__antithesis_instrumentation__.Notify(561609)
		s.metrics.PGWireCancelSuccessfulCount.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(561610)
		if err != nil {
			__antithesis_instrumentation__.Notify(561611)
			if respStatus := status.Convert(err); respStatus.Code() == codes.ResourceExhausted {
				__antithesis_instrumentation__.Notify(561613)
				s.metrics.PGWireCancelIgnoredCount.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(561614)
			}
			__antithesis_instrumentation__.Notify(561612)
			log.Sessions.Warningf(ctx, "unexpected while handling pgwire cancellation request: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(561615)
		}
	}
}

func parseClientProvidedSessionParameters(
	ctx context.Context,
	sv *settings.Values,
	buf *pgwirebase.ReadBuffer,
	origRemoteAddr net.Addr,
	trustClientProvidedRemoteAddr bool,
) (sql.SessionArgs, error) {
	__antithesis_instrumentation__.Notify(561616)
	args := sql.SessionArgs{
		SessionDefaults:             make(map[string]string),
		CustomOptionSessionDefaults: make(map[string]string),
		RemoteAddr:                  origRemoteAddr,
	}
	foundBufferSize := false

	for {
		__antithesis_instrumentation__.Notify(561621)

		key, err := buf.GetString()
		if err != nil {
			__antithesis_instrumentation__.Notify(561625)
			return sql.SessionArgs{}, pgerror.Wrap(
				err, pgcode.ProtocolViolation,
				"error reading option key",
			)
		} else {
			__antithesis_instrumentation__.Notify(561626)
		}
		__antithesis_instrumentation__.Notify(561622)
		if len(key) == 0 {
			__antithesis_instrumentation__.Notify(561627)

			break
		} else {
			__antithesis_instrumentation__.Notify(561628)
		}
		__antithesis_instrumentation__.Notify(561623)
		value, err := buf.GetString()
		if err != nil {
			__antithesis_instrumentation__.Notify(561629)
			return sql.SessionArgs{}, pgerror.Wrapf(
				err, pgcode.ProtocolViolation,
				"error reading option value for key %q", key,
			)
		} else {
			__antithesis_instrumentation__.Notify(561630)
		}
		__antithesis_instrumentation__.Notify(561624)

		key = strings.ToLower(key)

		switch key {
		case "user":
			__antithesis_instrumentation__.Notify(561631)

			args.User, _ = security.MakeSQLUsernameFromUserInput(value, security.UsernameValidation)

			args.IsSuperuser = args.User.IsRootUser()

		case "crdb:session_revival_token_base64":
			__antithesis_instrumentation__.Notify(561632)
			token, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				__antithesis_instrumentation__.Notify(561645)
				return sql.SessionArgs{}, pgerror.Wrapf(
					err, pgcode.ProtocolViolation,
					"%s", key,
				)
			} else {
				__antithesis_instrumentation__.Notify(561646)
			}
			__antithesis_instrumentation__.Notify(561633)
			args.SessionRevivalToken = token

		case "results_buffer_size":
			__antithesis_instrumentation__.Notify(561634)
			if args.ConnResultsBufferSize, err = humanizeutil.ParseBytes(value); err != nil {
				__antithesis_instrumentation__.Notify(561647)
				return sql.SessionArgs{}, errors.WithSecondaryError(
					pgerror.Newf(pgcode.ProtocolViolation,
						"error parsing results_buffer_size option value '%s' as bytes", value), err)
			} else {
				__antithesis_instrumentation__.Notify(561648)
			}
			__antithesis_instrumentation__.Notify(561635)
			if args.ConnResultsBufferSize < 0 {
				__antithesis_instrumentation__.Notify(561649)
				return sql.SessionArgs{}, pgerror.Newf(pgcode.ProtocolViolation,
					"results_buffer_size option value '%s' cannot be negative", value)
			} else {
				__antithesis_instrumentation__.Notify(561650)
			}
			__antithesis_instrumentation__.Notify(561636)
			foundBufferSize = true

		case "crdb:remote_addr":
			__antithesis_instrumentation__.Notify(561637)
			if !trustClientProvidedRemoteAddr {
				__antithesis_instrumentation__.Notify(561651)
				return sql.SessionArgs{}, pgerror.Newf(pgcode.ProtocolViolation,
					"server not configured to accept remote address override (requested: %q)", value)
			} else {
				__antithesis_instrumentation__.Notify(561652)
			}
			__antithesis_instrumentation__.Notify(561638)

			hostS, portS, err := net.SplitHostPort(value)
			if err != nil {
				__antithesis_instrumentation__.Notify(561653)
				return sql.SessionArgs{}, pgerror.Wrap(
					err, pgcode.ProtocolViolation,
					"invalid address format",
				)
			} else {
				__antithesis_instrumentation__.Notify(561654)
			}
			__antithesis_instrumentation__.Notify(561639)
			port, err := strconv.Atoi(portS)
			if err != nil {
				__antithesis_instrumentation__.Notify(561655)
				return sql.SessionArgs{}, pgerror.Wrap(
					err, pgcode.ProtocolViolation,
					"remote port is not numeric",
				)
			} else {
				__antithesis_instrumentation__.Notify(561656)
			}
			__antithesis_instrumentation__.Notify(561640)
			ip := net.ParseIP(hostS)
			if ip == nil {
				__antithesis_instrumentation__.Notify(561657)
				return sql.SessionArgs{}, pgerror.New(pgcode.ProtocolViolation,
					"remote address is not numeric")
			} else {
				__antithesis_instrumentation__.Notify(561658)
			}
			__antithesis_instrumentation__.Notify(561641)
			args.RemoteAddr = &net.TCPAddr{IP: ip, Port: port}

		case "options":
			__antithesis_instrumentation__.Notify(561642)
			opts, err := parseOptions(value)
			if err != nil {
				__antithesis_instrumentation__.Notify(561659)
				return sql.SessionArgs{}, err
			} else {
				__antithesis_instrumentation__.Notify(561660)
			}
			__antithesis_instrumentation__.Notify(561643)
			for _, opt := range opts {
				__antithesis_instrumentation__.Notify(561661)
				err = loadParameter(ctx, opt.key, opt.value, &args)
				if err != nil {
					__antithesis_instrumentation__.Notify(561662)
					return sql.SessionArgs{}, pgerror.Wrapf(err, pgerror.GetPGCode(err), "options")
				} else {
					__antithesis_instrumentation__.Notify(561663)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(561644)
			err = loadParameter(ctx, key, value, &args)
			if err != nil {
				__antithesis_instrumentation__.Notify(561664)
				return sql.SessionArgs{}, err
			} else {
				__antithesis_instrumentation__.Notify(561665)
			}
		}
	}
	__antithesis_instrumentation__.Notify(561617)

	if !foundBufferSize && func() bool {
		__antithesis_instrumentation__.Notify(561666)
		return sv != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(561667)

		args.ConnResultsBufferSize = connResultsBufferSize.Get(sv)
	} else {
		__antithesis_instrumentation__.Notify(561668)
	}
	__antithesis_instrumentation__.Notify(561618)

	if _, ok := args.SessionDefaults["database"]; !ok {
		__antithesis_instrumentation__.Notify(561669)

		args.SessionDefaults["database"] = catalogkeys.DefaultDatabaseName
	} else {
		__antithesis_instrumentation__.Notify(561670)
	}
	__antithesis_instrumentation__.Notify(561619)

	if appName, ok := args.SessionDefaults["application_name"]; ok {
		__antithesis_instrumentation__.Notify(561671)
		if appName == catconstants.ReportableAppNamePrefix+catconstants.InternalSQLAppName {
			__antithesis_instrumentation__.Notify(561672)
			telemetry.Inc(sqltelemetry.CockroachShellCounter)
		} else {
			__antithesis_instrumentation__.Notify(561673)
		}
	} else {
		__antithesis_instrumentation__.Notify(561674)
	}
	__antithesis_instrumentation__.Notify(561620)

	return args, nil
}

func loadParameter(ctx context.Context, key, value string, args *sql.SessionArgs) error {
	__antithesis_instrumentation__.Notify(561675)
	key = strings.ToLower(key)
	exists, configurable := sql.IsSessionVariableConfigurable(key)

	switch {
	case exists && func() bool {
		__antithesis_instrumentation__.Notify(561683)
		return configurable == true
	}() == true:
		__antithesis_instrumentation__.Notify(561677)
		args.SessionDefaults[key] = value
	case sql.IsCustomOptionSessionVariable(key):
		__antithesis_instrumentation__.Notify(561678)
		args.CustomOptionSessionDefaults[key] = value
	case !exists:
		__antithesis_instrumentation__.Notify(561679)
		if _, ok := sql.UnsupportedVars[key]; ok {
			__antithesis_instrumentation__.Notify(561684)
			counter := sqltelemetry.UnimplementedClientStatusParameterCounter(key)
			telemetry.Inc(counter)
		} else {
			__antithesis_instrumentation__.Notify(561685)
		}
		__antithesis_instrumentation__.Notify(561680)
		log.Warningf(ctx, "unknown configuration parameter: %q", key)

	case !configurable:
		__antithesis_instrumentation__.Notify(561681)
		return pgerror.Newf(pgcode.CantChangeRuntimeParam,
			"parameter %q cannot be changed", key)
	default:
		__antithesis_instrumentation__.Notify(561682)
	}
	__antithesis_instrumentation__.Notify(561676)
	return nil
}

type option struct {
	key   string
	value string
}

func parseOptions(optionsString string) ([]option, error) {
	__antithesis_instrumentation__.Notify(561686)
	var res []option
	optionsRaw, err := url.QueryUnescape(optionsString)
	if err != nil {
		__antithesis_instrumentation__.Notify(561689)
		return nil, pgerror.Newf(pgcode.ProtocolViolation, "failed to unescape options %q", optionsString)
	} else {
		__antithesis_instrumentation__.Notify(561690)
	}
	__antithesis_instrumentation__.Notify(561687)

	lastWasDashC := false
	opts := splitOptions(optionsRaw)

	for i := 0; i < len(opts); i++ {
		__antithesis_instrumentation__.Notify(561691)
		prefix := ""
		if len(opts[i]) > 1 {
			__antithesis_instrumentation__.Notify(561695)
			prefix = opts[i][:2]
		} else {
			__antithesis_instrumentation__.Notify(561696)
		}
		__antithesis_instrumentation__.Notify(561692)

		switch {
		case opts[i] == "-c":
			__antithesis_instrumentation__.Notify(561697)
			lastWasDashC = true
			continue
		case lastWasDashC:
			__antithesis_instrumentation__.Notify(561698)
			lastWasDashC = false

			prefix = ""
		case prefix == "--" || func() bool {
			__antithesis_instrumentation__.Notify(561701)
			return prefix == "-c" == true
		}() == true:
			__antithesis_instrumentation__.Notify(561699)
			lastWasDashC = false
		default:
			__antithesis_instrumentation__.Notify(561700)
			return nil, pgerror.Newf(pgcode.ProtocolViolation,
				"option %q is invalid, must have prefix '-c' or '--'", opts[i])
		}
		__antithesis_instrumentation__.Notify(561693)

		opt, err := splitOption(opts[i], prefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(561702)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(561703)
		}
		__antithesis_instrumentation__.Notify(561694)
		res = append(res, opt)
	}
	__antithesis_instrumentation__.Notify(561688)
	return res, nil
}

func splitOptions(options string) []string {
	__antithesis_instrumentation__.Notify(561704)
	var res []string
	var sb strings.Builder
	i := 0
	for i < len(options) {
		__antithesis_instrumentation__.Notify(561706)
		sb.Reset()

		for i < len(options) && func() bool {
			__antithesis_instrumentation__.Notify(561710)
			return unicode.IsSpace(rune(options[i])) == true
		}() == true {
			__antithesis_instrumentation__.Notify(561711)
			i++
		}
		__antithesis_instrumentation__.Notify(561707)
		if i == len(options) {
			__antithesis_instrumentation__.Notify(561712)
			break
		} else {
			__antithesis_instrumentation__.Notify(561713)
		}
		__antithesis_instrumentation__.Notify(561708)

		lastWasEscape := false

		for i < len(options) {
			__antithesis_instrumentation__.Notify(561714)
			if unicode.IsSpace(rune(options[i])) && func() bool {
				__antithesis_instrumentation__.Notify(561717)
				return !lastWasEscape == true
			}() == true {
				__antithesis_instrumentation__.Notify(561718)
				break
			} else {
				__antithesis_instrumentation__.Notify(561719)
			}
			__antithesis_instrumentation__.Notify(561715)
			if !lastWasEscape && func() bool {
				__antithesis_instrumentation__.Notify(561720)
				return options[i] == '\\' == true
			}() == true {
				__antithesis_instrumentation__.Notify(561721)
				lastWasEscape = true
			} else {
				__antithesis_instrumentation__.Notify(561722)
				lastWasEscape = false
				sb.WriteByte(options[i])
			}
			__antithesis_instrumentation__.Notify(561716)
			i++
		}
		__antithesis_instrumentation__.Notify(561709)

		res = append(res, sb.String())
	}
	__antithesis_instrumentation__.Notify(561705)

	return res
}

func splitOption(opt, prefix string) (option, error) {
	__antithesis_instrumentation__.Notify(561723)
	kv := strings.Split(opt, "=")

	if len(kv) != 2 {
		__antithesis_instrumentation__.Notify(561725)
		return option{}, pgerror.Newf(pgcode.ProtocolViolation,
			"option %q is invalid, check '='", opt)
	} else {
		__antithesis_instrumentation__.Notify(561726)
	}
	__antithesis_instrumentation__.Notify(561724)

	kv[0] = strings.TrimPrefix(kv[0], prefix)

	return option{key: strings.ReplaceAll(kv[0], "-", "_"), value: kv[1]}, nil
}

var trustClientProvidedRemoteAddrOverride = envutil.EnvOrDefaultBool("COCKROACH_TRUST_CLIENT_PROVIDED_SQL_REMOTE_ADDR", false)

func (s *Server) TestingSetTrustClientProvidedRemoteAddr(b bool) func() {
	__antithesis_instrumentation__.Notify(561727)
	prev := s.trustClientProvidedRemoteAddr.Get()
	s.trustClientProvidedRemoteAddr.Set(b)
	return func() { __antithesis_instrumentation__.Notify(561728); s.trustClientProvidedRemoteAddr.Set(prev) }
}

func (s *Server) maybeUpgradeToSecureConn(
	ctx context.Context,
	conn net.Conn,
	connType hba.ConnType,
	version uint32,
	buf *pgwirebase.ReadBuffer,
) (newConn net.Conn, newConnType hba.ConnType, newVersion uint32, clientErr, serverErr error) {
	__antithesis_instrumentation__.Notify(561729)

	newConn = conn
	newConnType = connType
	newVersion = version
	var n int

	if version != versionSSL {
		__antithesis_instrumentation__.Notify(561735)

		if s.cfg.Insecure {
			__antithesis_instrumentation__.Notify(561738)
			return
		} else {
			__antithesis_instrumentation__.Notify(561739)
		}
		__antithesis_instrumentation__.Notify(561736)

		if !s.cfg.AcceptSQLWithoutTLS && func() bool {
			__antithesis_instrumentation__.Notify(561740)
			return connType != hba.ConnLocal == true
		}() == true {
			__antithesis_instrumentation__.Notify(561741)
			clientErr = pgerror.New(pgcode.ProtocolViolation, ErrSSLRequired)
		} else {
			__antithesis_instrumentation__.Notify(561742)
		}
		__antithesis_instrumentation__.Notify(561737)
		return
	} else {
		__antithesis_instrumentation__.Notify(561743)
	}
	__antithesis_instrumentation__.Notify(561730)

	if connType == hba.ConnLocal {
		__antithesis_instrumentation__.Notify(561744)

		clientErr = pgerror.New(pgcode.ProtocolViolation,
			"cannot use SSL/TLS over local connections")
		return
	} else {
		__antithesis_instrumentation__.Notify(561745)
	}
	__antithesis_instrumentation__.Notify(561731)

	if len(buf.Msg) > 0 {
		__antithesis_instrumentation__.Notify(561746)
		serverErr = errors.Errorf("unexpected data after SSLRequest: %q", buf.Msg)
		return
	} else {
		__antithesis_instrumentation__.Notify(561747)
	}
	__antithesis_instrumentation__.Notify(561732)

	tlsConfig, serverErr := s.execCfg.RPCContext.GetServerTLSConfig()
	if serverErr != nil {
		__antithesis_instrumentation__.Notify(561748)
		return
	} else {
		__antithesis_instrumentation__.Notify(561749)
	}
	__antithesis_instrumentation__.Notify(561733)

	if tlsConfig == nil {
		__antithesis_instrumentation__.Notify(561750)

		n, serverErr = conn.Write(sslUnsupported)
		if serverErr != nil {
			__antithesis_instrumentation__.Notify(561751)
			return
		} else {
			__antithesis_instrumentation__.Notify(561752)
		}
	} else {
		__antithesis_instrumentation__.Notify(561753)

		n, serverErr = conn.Write(sslSupported)
		if serverErr != nil {
			__antithesis_instrumentation__.Notify(561755)
			return
		} else {
			__antithesis_instrumentation__.Notify(561756)
		}
		__antithesis_instrumentation__.Notify(561754)
		newConn = tls.Server(conn, tlsConfig)
		newConnType = hba.ConnHostSSL
	}
	__antithesis_instrumentation__.Notify(561734)
	s.metrics.BytesOutCount.Inc(int64(n))

	newVersion, *buf, serverErr = s.readVersion(newConn)
	return
}

func (s *Server) registerConn(
	ctx context.Context,
) (newCtx context.Context, rejectNewConnections bool, onCloseFn func()) {
	__antithesis_instrumentation__.Notify(561757)
	onCloseFn = func() { __antithesis_instrumentation__.Notify(561761) }
	__antithesis_instrumentation__.Notify(561758)
	newCtx = ctx
	s.mu.Lock()
	rejectNewConnections = s.mu.rejectNewConnections
	if !rejectNewConnections {
		__antithesis_instrumentation__.Notify(561762)
		var cancel context.CancelFunc
		newCtx, cancel = contextutil.WithCancel(ctx)
		done := make(chan struct{})
		s.mu.connCancelMap[done] = cancel
		onCloseFn = func() {
			__antithesis_instrumentation__.Notify(561763)
			cancel()
			close(done)
			s.mu.Lock()
			delete(s.mu.connCancelMap, done)
			s.mu.Unlock()
		}
	} else {
		__antithesis_instrumentation__.Notify(561764)
	}
	__antithesis_instrumentation__.Notify(561759)
	s.mu.Unlock()

	if !rejectNewConnections {
		__antithesis_instrumentation__.Notify(561765)
		s.metrics.NewConns.Inc(1)
		s.metrics.Conns.Inc(1)
		prevOnCloseFn := onCloseFn
		onCloseFn = func() { __antithesis_instrumentation__.Notify(561766); prevOnCloseFn(); s.metrics.Conns.Dec(1) }
	} else {
		__antithesis_instrumentation__.Notify(561767)
	}
	__antithesis_instrumentation__.Notify(561760)
	return
}

func (s *Server) readVersion(
	conn io.Reader,
) (version uint32, buf pgwirebase.ReadBuffer, err error) {
	__antithesis_instrumentation__.Notify(561768)
	var n int
	buf = pgwirebase.MakeReadBuffer(
		pgwirebase.ReadBufferOptionWithClusterSettings(&s.execCfg.Settings.SV),
	)
	n, err = buf.ReadUntypedMsg(conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(561771)
		return
	} else {
		__antithesis_instrumentation__.Notify(561772)
	}
	__antithesis_instrumentation__.Notify(561769)
	version, err = buf.GetUint32()
	if err != nil {
		__antithesis_instrumentation__.Notify(561773)
		return
	} else {
		__antithesis_instrumentation__.Notify(561774)
	}
	__antithesis_instrumentation__.Notify(561770)
	s.metrics.BytesInCount.Inc(int64(n))
	return
}

func (s *Server) sendErr(ctx context.Context, conn net.Conn, err error) error {
	__antithesis_instrumentation__.Notify(561775)
	msgBuilder := newWriteBuffer(s.metrics.BytesOutCount)

	_ = writeErr(ctx, &s.execCfg.Settings.SV, err, msgBuilder, conn)
	_ = conn.Close()
	return err
}

func newAdminShutdownErr(msg string) error {
	__antithesis_instrumentation__.Notify(561776)
	return pgerror.New(pgcode.AdminShutdown, msg)
}
