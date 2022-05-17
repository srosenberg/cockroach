package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/redact"
)

type Metrics struct {
	ConnectionsRefused *metric.Counter
	BytesReceived      *metric.Counter
	BytesSent          *metric.Counter
	InfosReceived      *metric.Counter
	InfosSent          *metric.Counter
}

func makeMetrics() Metrics {
	__antithesis_instrumentation__.Notify(68188)
	return Metrics{
		ConnectionsRefused: metric.NewCounter(MetaConnectionsRefused),
		BytesReceived:      metric.NewCounter(MetaBytesReceived),
		BytesSent:          metric.NewCounter(MetaBytesSent),
		InfosReceived:      metric.NewCounter(MetaInfosReceived),
		InfosSent:          metric.NewCounter(MetaInfosSent),
	}
}

func (m Metrics) String() string {
	__antithesis_instrumentation__.Notify(68189)
	return redact.StringWithoutMarkers(m.Snapshot())
}

func (m Metrics) Snapshot() MetricSnap {
	__antithesis_instrumentation__.Notify(68190)
	return MetricSnap{
		ConnsRefused:  m.ConnectionsRefused.Count(),
		BytesReceived: m.BytesReceived.Count(),
		BytesSent:     m.BytesSent.Count(),
		InfosReceived: m.InfosReceived.Count(),
		InfosSent:     m.InfosSent.Count(),
	}
}

func (m MetricSnap) String() string {
	__antithesis_instrumentation__.Notify(68191)
	return redact.StringWithoutMarkers(m)
}

func (m MetricSnap) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(68192)
	w.Printf("infos %d/%d sent/received, bytes %dB/%dB sent/received",
		m.InfosSent, m.InfosReceived,
		m.BytesSent, m.BytesReceived)
	if m.ConnsRefused > 0 {
		__antithesis_instrumentation__.Notify(68193)
		w.Printf(", refused %d conns", m.ConnsRefused)
	} else {
		__antithesis_instrumentation__.Notify(68194)
	}
}

func (c OutgoingConnStatus) String() string {
	__antithesis_instrumentation__.Notify(68195)
	return redact.StringWithoutMarkers(c)
}

func (c OutgoingConnStatus) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(68196)
	w.Printf("%d: %s (%s: %s)",
		c.NodeID, c.Address,
		roundSecs(time.Duration(c.AgeNanos)), c.MetricSnap)
}

func (c ClientStatus) String() string {
	__antithesis_instrumentation__.Notify(68197)
	return redact.StringWithoutMarkers(c)
}

func (c ClientStatus) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(68198)
	w.Printf("gossip client (%d/%d cur/max conns)\n",
		len(c.ConnStatus), c.MaxConns)
	for _, conn := range c.ConnStatus {
		__antithesis_instrumentation__.Notify(68199)
		w.Printf("  %s\n", conn)
	}
}

func (c ConnStatus) String() string {
	__antithesis_instrumentation__.Notify(68200)
	return redact.StringWithoutMarkers(c)
}

func (c ConnStatus) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(68201)
	w.Printf("%d: %s (%s)", c.NodeID, c.Address,
		roundSecs(time.Duration(c.AgeNanos)))
}

func (s ServerStatus) String() string {
	__antithesis_instrumentation__.Notify(68202)
	return redact.StringWithoutMarkers(s)
}

func (s ServerStatus) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(68203)
	w.Printf("gossip server (%d/%d cur/max conns, %s)\n",
		len(s.ConnStatus), s.MaxConns, s.MetricSnap)
	for _, conn := range s.ConnStatus {
		__antithesis_instrumentation__.Notify(68204)
		w.Printf("  %s\n", conn)
	}
}

func (c Connectivity) String() string {
	__antithesis_instrumentation__.Notify(68205)
	return redact.StringWithoutMarkers(c)
}

func (c Connectivity) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(68206)
	w.Printf("gossip connectivity\n")
	if c.SentinelNodeID != 0 {
		__antithesis_instrumentation__.Notify(68208)
		w.Printf("  n%d [sentinel];\n", c.SentinelNodeID)
	} else {
		__antithesis_instrumentation__.Notify(68209)
	}
	__antithesis_instrumentation__.Notify(68207)
	if len(c.ClientConns) > 0 {
		__antithesis_instrumentation__.Notify(68210)
		w.SafeRune(' ')
		for _, conn := range c.ClientConns {
			__antithesis_instrumentation__.Notify(68212)
			w.Printf(" n%d -> n%d;", conn.SourceID, conn.TargetID)
		}
		__antithesis_instrumentation__.Notify(68211)
		w.SafeRune('\n')
	} else {
		__antithesis_instrumentation__.Notify(68213)
	}
}
