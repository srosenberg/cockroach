package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/metric"

var (
	metaHeartbeatsInitializing = metric.Metadata{
		Name:        "rpc.heartbeats.initializing",
		Help:        "Gauge of current connections in the initializing state",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	metaHeartbeatsNominal = metric.Metadata{
		Name:        "rpc.heartbeats.nominal",
		Help:        "Gauge of current connections in the nominal state",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	metaHeartbeatsFailed = metric.Metadata{
		Name:        "rpc.heartbeats.failed",
		Help:        "Gauge of current connections in the failed state",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}

	metaHeartbeatLoopsStarted = metric.Metadata{
		Name: "rpc.heartbeats.loops.started",
		Help: "Counter of the number of connection heartbeat loops which " +
			"have been started",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	metaHeartbeatLoopsExited = metric.Metadata{
		Name: "rpc.heartbeats.loops.exited",
		Help: "Counter of the number of connection heartbeat loops which " +
			"have exited with an error",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
)

type heartbeatState int

const (
	heartbeatNotRunning heartbeatState = iota
	heartbeatInitializing
	heartbeatNominal
	heartbeatFailed
)

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(185362)
	return Metrics{
		HeartbeatLoopsStarted:  metric.NewCounter(metaHeartbeatLoopsStarted),
		HeartbeatLoopsExited:   metric.NewCounter(metaHeartbeatLoopsExited),
		HeartbeatsInitializing: metric.NewGauge(metaHeartbeatsInitializing),
		HeartbeatsNominal:      metric.NewGauge(metaHeartbeatsNominal),
		HeartbeatsFailed:       metric.NewGauge(metaHeartbeatsFailed),
	}
}

type Metrics struct {
	HeartbeatLoopsStarted *metric.Counter

	HeartbeatLoopsExited *metric.Counter

	HeartbeatsInitializing *metric.Gauge

	HeartbeatsNominal *metric.Gauge

	HeartbeatsFailed *metric.Gauge
}

func updateHeartbeatState(m *Metrics, old, new heartbeatState) heartbeatState {
	__antithesis_instrumentation__.Notify(185363)
	if old == new {
		__antithesis_instrumentation__.Notify(185367)
		return new
	} else {
		__antithesis_instrumentation__.Notify(185368)
	}
	__antithesis_instrumentation__.Notify(185364)
	if g := heartbeatGauge(m, new); g != nil {
		__antithesis_instrumentation__.Notify(185369)
		g.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(185370)
	}
	__antithesis_instrumentation__.Notify(185365)
	if g := heartbeatGauge(m, old); g != nil {
		__antithesis_instrumentation__.Notify(185371)
		g.Dec(1)
	} else {
		__antithesis_instrumentation__.Notify(185372)
	}
	__antithesis_instrumentation__.Notify(185366)
	return new
}

func heartbeatGauge(m *Metrics, s heartbeatState) (g *metric.Gauge) {
	__antithesis_instrumentation__.Notify(185373)
	switch s {
	case heartbeatInitializing:
		__antithesis_instrumentation__.Notify(185375)
		g = m.HeartbeatsInitializing
	case heartbeatNominal:
		__antithesis_instrumentation__.Notify(185376)
		g = m.HeartbeatsNominal
	case heartbeatFailed:
		__antithesis_instrumentation__.Notify(185377)
		g = m.HeartbeatsFailed
	default:
		__antithesis_instrumentation__.Notify(185378)
	}
	__antithesis_instrumentation__.Notify(185374)
	return g
}
