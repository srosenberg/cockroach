package tenantrate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/metric/aggmetric"
)

type Metrics struct {
	Tenants               *metric.Gauge
	CurrentBlocked        *aggmetric.AggGauge
	ReadRequestsAdmitted  *aggmetric.AggCounter
	WriteRequestsAdmitted *aggmetric.AggCounter
	ReadBytesAdmitted     *aggmetric.AggCounter
	WriteBytesAdmitted    *aggmetric.AggCounter
}

var _ metric.Struct = (*Metrics)(nil)

var (
	metaTenants = metric.Metadata{
		Name:        "kv.tenant_rate_limit.num_tenants",
		Help:        "Number of tenants currently being tracked",
		Measurement: "Tenants",
		Unit:        metric.Unit_COUNT,
	}
	metaCurrentBlocked = metric.Metadata{
		Name:        "kv.tenant_rate_limit.current_blocked",
		Help:        "Number of requests currently blocked by the rate limiter",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	metaReadRequestsAdmitted = metric.Metadata{
		Name:        "kv.tenant_rate_limit.read_requests_admitted",
		Help:        "Number of read requests admitted by the rate limiter",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	metaWriteRequestsAdmitted = metric.Metadata{
		Name:        "kv.tenant_rate_limit.write_requests_admitted",
		Help:        "Number of write requests admitted by the rate limiter",
		Measurement: "Requests",
		Unit:        metric.Unit_COUNT,
	}
	metaReadBytesAdmitted = metric.Metadata{
		Name:        "kv.tenant_rate_limit.read_bytes_admitted",
		Help:        "Number of read bytes admitted by the rate limiter",
		Measurement: "Bytes",
		Unit:        metric.Unit_BYTES,
	}
	metaWriteBytesAdmitted = metric.Metadata{
		Name:        "kv.tenant_rate_limit.write_bytes_admitted",
		Help:        "Number of write bytes admitted by the rate limiter",
		Measurement: "Bytes",
		Unit:        metric.Unit_BYTES,
	}
)

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(126699)
	b := aggmetric.MakeBuilder(multitenant.TenantIDLabel)
	return Metrics{
		Tenants:               metric.NewGauge(metaTenants),
		CurrentBlocked:        b.Gauge(metaCurrentBlocked),
		ReadRequestsAdmitted:  b.Counter(metaReadRequestsAdmitted),
		WriteRequestsAdmitted: b.Counter(metaWriteRequestsAdmitted),
		ReadBytesAdmitted:     b.Counter(metaReadBytesAdmitted),
		WriteBytesAdmitted:    b.Counter(metaWriteBytesAdmitted),
	}
}

func (m *Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(126700) }

type tenantMetrics struct {
	currentBlocked        *aggmetric.Gauge
	readRequestsAdmitted  *aggmetric.Counter
	writeRequestsAdmitted *aggmetric.Counter
	readBytesAdmitted     *aggmetric.Counter
	writeBytesAdmitted    *aggmetric.Counter
}

func (m *Metrics) tenantMetrics(tenantID roachpb.TenantID) tenantMetrics {
	__antithesis_instrumentation__.Notify(126701)
	tid := tenantID.String()
	return tenantMetrics{
		currentBlocked:        m.CurrentBlocked.AddChild(tid),
		readRequestsAdmitted:  m.ReadRequestsAdmitted.AddChild(tid),
		writeRequestsAdmitted: m.WriteRequestsAdmitted.AddChild(tid),
		readBytesAdmitted:     m.ReadBytesAdmitted.AddChild(tid),
		writeBytesAdmitted:    m.WriteBytesAdmitted.AddChild(tid),
	}
}

func (tm *tenantMetrics) destroy() {
	__antithesis_instrumentation__.Notify(126702)
	tm.currentBlocked.Destroy()
	tm.readRequestsAdmitted.Destroy()
	tm.writeRequestsAdmitted.Destroy()
	tm.readBytesAdmitted.Destroy()
	tm.writeBytesAdmitted.Destroy()
}
