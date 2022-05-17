package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

type nodeSet struct {
	nodes        map[roachpb.NodeID]struct{}
	placeholders int
	maxSize      int
	gauge        *metric.Gauge
}

func makeNodeSet(maxSize int, gauge *metric.Gauge) nodeSet {
	__antithesis_instrumentation__.Notify(67986)
	return nodeSet{
		nodes:   make(map[roachpb.NodeID]struct{}),
		maxSize: maxSize,
		gauge:   gauge,
	}
}

func (as nodeSet) hasSpace() bool {
	__antithesis_instrumentation__.Notify(67987)
	return as.len() < as.maxSize
}

func (as nodeSet) len() int {
	__antithesis_instrumentation__.Notify(67988)
	return len(as.nodes) + as.placeholders
}

func (as nodeSet) asSlice() []roachpb.NodeID {
	__antithesis_instrumentation__.Notify(67989)
	slice := make([]roachpb.NodeID, 0, len(as.nodes))
	for node := range as.nodes {
		__antithesis_instrumentation__.Notify(67991)
		slice = append(slice, node)
	}
	__antithesis_instrumentation__.Notify(67990)
	return slice
}

func (as nodeSet) filter(filterFn func(node roachpb.NodeID) bool) nodeSet {
	__antithesis_instrumentation__.Notify(67992)
	avail := makeNodeSet(as.maxSize,
		metric.NewGauge(metric.Metadata{Name: "TODO(marc)", Help: "TODO(marc)"}))
	for node := range as.nodes {
		__antithesis_instrumentation__.Notify(67994)
		if filterFn(node) {
			__antithesis_instrumentation__.Notify(67995)
			avail.addNode(node)
		} else {
			__antithesis_instrumentation__.Notify(67996)
		}
	}
	__antithesis_instrumentation__.Notify(67993)
	return avail
}

func (as nodeSet) hasNode(node roachpb.NodeID) bool {
	__antithesis_instrumentation__.Notify(67997)
	_, ok := as.nodes[node]
	return ok
}

func (as *nodeSet) setMaxSize(maxSize int) {
	__antithesis_instrumentation__.Notify(67998)
	as.maxSize = maxSize
}

func (as *nodeSet) addNode(node roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(67999)

	if !as.hasNode(node) {
		__antithesis_instrumentation__.Notify(68001)
		as.nodes[node] = struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(68002)
		as.placeholders++
	}
	__antithesis_instrumentation__.Notify(68000)
	as.updateGauge()
}

func (as *nodeSet) removeNode(node roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(68003)

	if as.hasNode(node) {
		__antithesis_instrumentation__.Notify(68005)
		delete(as.nodes, node)
	} else {
		__antithesis_instrumentation__.Notify(68006)
		as.placeholders--
	}
	__antithesis_instrumentation__.Notify(68004)
	as.updateGauge()
}

func (as *nodeSet) addPlaceholder() {
	__antithesis_instrumentation__.Notify(68007)
	as.placeholders++
	as.updateGauge()
}

func (as *nodeSet) resolvePlaceholder(node roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(68008)
	as.placeholders--
	as.addNode(node)
}

func (as *nodeSet) updateGauge() {
	__antithesis_instrumentation__.Notify(68009)
	if as.placeholders < 0 {
		__antithesis_instrumentation__.Notify(68011)
		log.Fatalf(context.TODO(),
			"nodeSet.placeholders should never be less than 0; gossip logic is broken %+v", as)
	} else {
		__antithesis_instrumentation__.Notify(68012)
	}
	__antithesis_instrumentation__.Notify(68010)
	as.gauge.Update(int64(as.len()))
}
