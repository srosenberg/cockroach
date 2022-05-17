package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
)

type metrics struct {
	BackendDisconnectCount *metric.Counter
	IdleDisconnectCount    *metric.Counter
	BackendDownCount       *metric.Counter
	ClientDisconnectCount  *metric.Counter
	CurConnCount           *metric.Gauge
	RoutingErrCount        *metric.Counter
	RefusedConnCount       *metric.Counter
	SuccessfulConnCount    *metric.Counter
	AuthFailedCount        *metric.Counter
	ExpiredClientConnCount *metric.Counter

	ConnMigrationSuccessCount                *metric.Counter
	ConnMigrationErrorFatalCount             *metric.Counter
	ConnMigrationErrorRecoverableCount       *metric.Counter
	ConnMigrationAttemptedCount              *metric.Counter
	ConnMigrationAttemptedLatency            *metric.Histogram
	ConnMigrationTransferResponseMessageSize *metric.Histogram
}

func (metrics) MetricStruct() { __antithesis_instrumentation__.Notify(21772) }

var _ metric.Struct = metrics{}

const (
	maxExpectedTransferResponseMessageSize = 1 << 24
)

var (
	metaCurConnCount = metric.Metadata{
		Name:        "proxy.sql.conns",
		Help:        "Number of connections being proxied",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	metaRoutingErrCount = metric.Metadata{
		Name:        "proxy.err.routing",
		Help:        "Number of errors encountered when attempting to route clients",
		Measurement: "Errors",
		Unit:        metric.Unit_COUNT,
	}
	metaBackendDownCount = metric.Metadata{
		Name:        "proxy.err.backend_down",
		Help:        "Number of errors encountered when connecting to backend servers",
		Measurement: "Errors",
		Unit:        metric.Unit_COUNT,
	}
	metaBackendDisconnectCount = metric.Metadata{
		Name:        "proxy.err.backend_disconnect",
		Help:        "Number of disconnects initiated by proxied backends",
		Measurement: "Disconnects",
		Unit:        metric.Unit_COUNT,
	}
	metaIdleDisconnectCount = metric.Metadata{
		Name:        "proxy.err.idle_disconnect",
		Help:        "Number of disconnects due to idle timeout",
		Measurement: "Idle Disconnects",
		Unit:        metric.Unit_COUNT,
	}
	metaClientDisconnectCount = metric.Metadata{
		Name:        "proxy.err.client_disconnect",
		Help:        "Number of disconnects initiated by clients",
		Measurement: "Client Disconnects",
		Unit:        metric.Unit_COUNT,
	}
	metaRefusedConnCount = metric.Metadata{
		Name:        "proxy.err.refused_conn",
		Help:        "Number of refused connections initiated by a given IP",
		Measurement: "Refused",
		Unit:        metric.Unit_COUNT,
	}
	metaSuccessfulConnCount = metric.Metadata{
		Name:        "proxy.sql.successful_conns",
		Help:        "Number of successful connections that were/are being proxied",
		Measurement: "Successful Connections",
		Unit:        metric.Unit_COUNT,
	}
	metaAuthFailedCount = metric.Metadata{
		Name:        "proxy.sql.authentication_failures",
		Help:        "Number of authentication failures",
		Measurement: "Authentication Failures",
		Unit:        metric.Unit_COUNT,
	}
	metaExpiredClientConnCount = metric.Metadata{
		Name:        "proxy.sql.expired_client_conns",
		Help:        "Number of expired client connections",
		Measurement: "Expired Client Connections",
		Unit:        metric.Unit_COUNT,
	}

	metaConnMigrationSuccessCount = metric.Metadata{
		Name:        "proxy.conn_migration.success",
		Help:        "Number of successful connection migrations",
		Measurement: "Connection Migrations",
		Unit:        metric.Unit_COUNT,
	}
	metaConnMigrationErrorFatalCount = metric.Metadata{

		Name:        "proxy.conn_migration.error_fatal",
		Help:        "Number of failed connection migrations which resulted in terminations",
		Measurement: "Connection Migrations",
		Unit:        metric.Unit_COUNT,
	}
	metaConnMigrationErrorRecoverableCount = metric.Metadata{

		Name:        "proxy.conn_migration.error_recoverable",
		Help:        "Number of failed connection migrations that were recoverable",
		Measurement: "Connection Migrations",
		Unit:        metric.Unit_COUNT,
	}
	metaConnMigrationAttemptedCount = metric.Metadata{
		Name:        "proxy.conn_migration.attempted",
		Help:        "Number of attempted connection migrations",
		Measurement: "Connection Migrations",
		Unit:        metric.Unit_COUNT,
	}
	metaConnMigrationAttemptedLatency = metric.Metadata{
		Name:        "proxy.conn_migration.attempted.latency",
		Help:        "Latency histogram for attempted connection migrations",
		Measurement: "Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaConnMigrationTransferResponseMessageSize = metric.Metadata{
		Name:        "proxy.conn_migration.transfer_response.message_size",
		Help:        "Message size for the SHOW TRANSFER STATE response",
		Measurement: "Bytes",
		Unit:        metric.Unit_BYTES,
	}
)

func makeProxyMetrics() metrics {
	__antithesis_instrumentation__.Notify(21773)
	return metrics{
		BackendDisconnectCount: metric.NewCounter(metaBackendDisconnectCount),
		IdleDisconnectCount:    metric.NewCounter(metaIdleDisconnectCount),
		BackendDownCount:       metric.NewCounter(metaBackendDownCount),
		ClientDisconnectCount:  metric.NewCounter(metaClientDisconnectCount),
		CurConnCount:           metric.NewGauge(metaCurConnCount),
		RoutingErrCount:        metric.NewCounter(metaRoutingErrCount),
		RefusedConnCount:       metric.NewCounter(metaRefusedConnCount),
		SuccessfulConnCount:    metric.NewCounter(metaSuccessfulConnCount),
		AuthFailedCount:        metric.NewCounter(metaAuthFailedCount),
		ExpiredClientConnCount: metric.NewCounter(metaExpiredClientConnCount),

		ConnMigrationSuccessCount:          metric.NewCounter(metaConnMigrationSuccessCount),
		ConnMigrationErrorFatalCount:       metric.NewCounter(metaConnMigrationErrorFatalCount),
		ConnMigrationErrorRecoverableCount: metric.NewCounter(metaConnMigrationErrorRecoverableCount),
		ConnMigrationAttemptedCount:        metric.NewCounter(metaConnMigrationAttemptedCount),
		ConnMigrationAttemptedLatency: metric.NewLatency(
			metaConnMigrationAttemptedLatency,
			base.DefaultHistogramWindowInterval(),
		),
		ConnMigrationTransferResponseMessageSize: metric.NewHistogram(
			metaConnMigrationTransferResponseMessageSize,
			base.DefaultHistogramWindowInterval(),
			maxExpectedTransferResponseMessageSize,
			1,
		),
	}
}

func (metrics *metrics) updateForError(err error) {
	__antithesis_instrumentation__.Notify(21774)
	if err == nil {
		__antithesis_instrumentation__.Notify(21776)
		return
	} else {
		__antithesis_instrumentation__.Notify(21777)
	}
	__antithesis_instrumentation__.Notify(21775)
	codeErr := (*codeError)(nil)
	if errors.As(err, &codeErr) {
		__antithesis_instrumentation__.Notify(21778)
		switch codeErr.code {
		case codeExpiredClientConnection:
			__antithesis_instrumentation__.Notify(21779)
			metrics.ExpiredClientConnCount.Inc(1)
		case codeBackendDisconnected:
			__antithesis_instrumentation__.Notify(21780)
			metrics.BackendDisconnectCount.Inc(1)
		case codeClientDisconnected:
			__antithesis_instrumentation__.Notify(21781)
			metrics.ClientDisconnectCount.Inc(1)
		case codeIdleDisconnect:
			__antithesis_instrumentation__.Notify(21782)
			metrics.IdleDisconnectCount.Inc(1)
		case codeProxyRefusedConnection:
			__antithesis_instrumentation__.Notify(21783)
			metrics.RefusedConnCount.Inc(1)
			metrics.BackendDownCount.Inc(1)
		case codeParamsRoutingFailed, codeUnavailable:
			__antithesis_instrumentation__.Notify(21784)
			metrics.RoutingErrCount.Inc(1)
			metrics.BackendDownCount.Inc(1)
		case codeBackendDown:
			__antithesis_instrumentation__.Notify(21785)
			metrics.BackendDownCount.Inc(1)
		case codeAuthFailed:
			__antithesis_instrumentation__.Notify(21786)
			metrics.AuthFailedCount.Inc(1)
		default:
			__antithesis_instrumentation__.Notify(21787)
		}
	} else {
		__antithesis_instrumentation__.Notify(21788)
	}
}
