package tenantcostserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/metric/aggmetric"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type Metrics struct {
	TotalRU                *aggmetric.AggGaugeFloat64
	TotalReadRequests      *aggmetric.AggGauge
	TotalReadBytes         *aggmetric.AggGauge
	TotalWriteRequests     *aggmetric.AggGauge
	TotalWriteBytes        *aggmetric.AggGauge
	TotalSQLPodsCPUSeconds *aggmetric.AggGaugeFloat64
	TotalPGWireEgressBytes *aggmetric.AggGauge

	mu struct {
		syncutil.Mutex

		tenantMetrics map[roachpb.TenantID]tenantMetrics
	}
}

var _ metric.Struct = (*Metrics)(nil)

func (m *Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(20219) }

var (
	metaTotalRU = metric.Metadata{
		Name:        "tenant.consumption.request_units",
		Help:        "Total RU consumption",
		Measurement: "Request Units",
		Unit:        metric.Unit_COUNT,
	}
	metaTotalReadRequests = metric.Metadata{
		Name:        "tenant.consumption.read_requests",
		Help:        "Total number of KV read requests",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	metaTotalReadBytes = metric.Metadata{
		Name:        "tenant.consumption.read_bytes",
		Help:        "Total number of bytes read from KV",
		Measurement: "Bytes",
		Unit:        metric.Unit_COUNT,
	}
	metaTotalWriteRequests = metric.Metadata{
		Name:        "tenant.consumption.write_requests",
		Help:        "Total number of KV write requests",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	metaTotalWriteBytes = metric.Metadata{
		Name:        "tenant.consumption.write_bytes",
		Help:        "Total number of bytes written to KV",
		Measurement: "Bytes",
		Unit:        metric.Unit_COUNT,
	}
	metaTotalSQLPodsCPUSeconds = metric.Metadata{
		Name:        "tenant.consumption.sql_pods_cpu_seconds",
		Help:        "Total number of bytes written to KV",
		Measurement: "CPU Seconds",
		Unit:        metric.Unit_SECONDS,
	}
	metaTotalPGWireEgressBytes = metric.Metadata{
		Name:        "tenant.consumption.pgwire_egress_bytes",
		Help:        "Total number of bytes transferred from a SQL pod to the client",
		Measurement: "Bytes",
		Unit:        metric.Unit_COUNT,
	}
)

func (m *Metrics) init() {
	b := aggmetric.MakeBuilder(multitenant.TenantIDLabel)
	*m = Metrics{
		TotalRU:                b.GaugeFloat64(metaTotalRU),
		TotalReadRequests:      b.Gauge(metaTotalReadRequests),
		TotalReadBytes:         b.Gauge(metaTotalReadBytes),
		TotalWriteRequests:     b.Gauge(metaTotalWriteRequests),
		TotalWriteBytes:        b.Gauge(metaTotalWriteBytes),
		TotalSQLPodsCPUSeconds: b.GaugeFloat64(metaTotalSQLPodsCPUSeconds),
		TotalPGWireEgressBytes: b.Gauge(metaTotalPGWireEgressBytes),
	}
	m.mu.tenantMetrics = make(map[roachpb.TenantID]tenantMetrics)
}

type tenantMetrics struct {
	totalRU                *aggmetric.GaugeFloat64
	totalReadRequests      *aggmetric.Gauge
	totalReadBytes         *aggmetric.Gauge
	totalWriteRequests     *aggmetric.Gauge
	totalWriteBytes        *aggmetric.Gauge
	totalSQLPodsCPUSeconds *aggmetric.GaugeFloat64
	totalPGWireEgressBytes *aggmetric.Gauge

	mutex *syncutil.Mutex
}

func (m *Metrics) getTenantMetrics(tenantID roachpb.TenantID) tenantMetrics {
	__antithesis_instrumentation__.Notify(20220)
	m.mu.Lock()
	defer m.mu.Unlock()
	tm, ok := m.mu.tenantMetrics[tenantID]
	if !ok {
		__antithesis_instrumentation__.Notify(20222)
		tid := tenantID.String()
		tm = tenantMetrics{
			totalRU:                m.TotalRU.AddChild(tid),
			totalReadRequests:      m.TotalReadRequests.AddChild(tid),
			totalReadBytes:         m.TotalReadBytes.AddChild(tid),
			totalWriteRequests:     m.TotalWriteRequests.AddChild(tid),
			totalWriteBytes:        m.TotalWriteBytes.AddChild(tid),
			totalSQLPodsCPUSeconds: m.TotalSQLPodsCPUSeconds.AddChild(tid),
			totalPGWireEgressBytes: m.TotalPGWireEgressBytes.AddChild(tid),
			mutex:                  &syncutil.Mutex{},
		}
		m.mu.tenantMetrics[tenantID] = tm
	} else {
		__antithesis_instrumentation__.Notify(20223)
	}
	__antithesis_instrumentation__.Notify(20221)
	return tm
}
