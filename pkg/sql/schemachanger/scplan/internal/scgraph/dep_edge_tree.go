package scgraph

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/google/btree"
)

type depEdgeTree struct {
	t     *btree.BTree
	order edgeTreeOrder
	cmp   nodeCmpFn
}

type nodeCmpFn func(a, b *screl.Node) (less, eq bool)

func newDepEdgeTree(order edgeTreeOrder, cmp nodeCmpFn) *depEdgeTree {
	__antithesis_instrumentation__.Notify(594209)
	const degree = 8
	return &depEdgeTree{
		t:     btree.New(degree),
		order: order,
		cmp:   cmp,
	}
}

type edgeTreeOrder bool

func (o edgeTreeOrder) first(e Edge) *screl.Node {
	__antithesis_instrumentation__.Notify(594210)
	if o == fromTo {
		__antithesis_instrumentation__.Notify(594212)
		return e.From()
	} else {
		__antithesis_instrumentation__.Notify(594213)
	}
	__antithesis_instrumentation__.Notify(594211)
	return e.To()
}

func (o edgeTreeOrder) second(e Edge) *screl.Node {
	__antithesis_instrumentation__.Notify(594214)
	if o == toFrom {
		__antithesis_instrumentation__.Notify(594216)
		return e.From()
	} else {
		__antithesis_instrumentation__.Notify(594217)
	}
	__antithesis_instrumentation__.Notify(594215)
	return e.To()
}

const (
	fromTo edgeTreeOrder = true
	toFrom edgeTreeOrder = false
)

type edgeTreeEntry struct {
	t    *depEdgeTree
	edge *DepEdge
}

func (et *depEdgeTree) insert(e *DepEdge) {
	__antithesis_instrumentation__.Notify(594218)
	et.t.ReplaceOrInsert(&edgeTreeEntry{
		t:    et,
		edge: e,
	})
}

func (et *depEdgeTree) iterateSourceNode(n *screl.Node, it DepEdgeIterator) (err error) {
	__antithesis_instrumentation__.Notify(594219)
	e := &edgeTreeEntry{t: et, edge: &DepEdge{}}
	if et.order == fromTo {
		__antithesis_instrumentation__.Notify(594223)
		e.edge.from = n
	} else {
		__antithesis_instrumentation__.Notify(594224)
		e.edge.to = n
	}
	__antithesis_instrumentation__.Notify(594220)
	et.t.AscendGreaterOrEqual(e, func(i btree.Item) (wantMore bool) {
		__antithesis_instrumentation__.Notify(594225)
		e := i.(*edgeTreeEntry)
		if et.order.first(e.edge) != n {
			__antithesis_instrumentation__.Notify(594227)
			return false
		} else {
			__antithesis_instrumentation__.Notify(594228)
		}
		__antithesis_instrumentation__.Notify(594226)
		err = it(e.edge)
		return err == nil
	})
	__antithesis_instrumentation__.Notify(594221)
	if iterutil.Done(err) {
		__antithesis_instrumentation__.Notify(594229)
		err = nil
	} else {
		__antithesis_instrumentation__.Notify(594230)
	}
	__antithesis_instrumentation__.Notify(594222)
	return err
}

func (e *edgeTreeEntry) Less(otherItem btree.Item) bool {
	__antithesis_instrumentation__.Notify(594231)
	o := otherItem.(*edgeTreeEntry)
	if less, eq := e.t.cmp(e.t.order.first(e.edge), e.t.order.first(o.edge)); !eq {
		__antithesis_instrumentation__.Notify(594233)
		return less
	} else {
		__antithesis_instrumentation__.Notify(594234)
	}
	__antithesis_instrumentation__.Notify(594232)
	less, _ := e.t.cmp(e.t.order.second(e.edge), e.t.order.second(o.edge))
	return less
}
