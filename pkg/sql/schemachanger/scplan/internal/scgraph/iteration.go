package scgraph

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
)

type NodeIterator func(n *screl.Node) error

func (g *Graph) ForEachNode(it NodeIterator) error {
	__antithesis_instrumentation__.Notify(594351)
	for _, m := range g.targetNodes {
		__antithesis_instrumentation__.Notify(594353)
		for i := 0; i < scpb.NumStatus; i++ {
			__antithesis_instrumentation__.Notify(594354)
			if ts, ok := m[scpb.Status(i)]; ok {
				__antithesis_instrumentation__.Notify(594355)
				if err := it(ts); err != nil {
					__antithesis_instrumentation__.Notify(594356)
					if iterutil.Done(err) {
						__antithesis_instrumentation__.Notify(594358)
						err = nil
					} else {
						__antithesis_instrumentation__.Notify(594359)
					}
					__antithesis_instrumentation__.Notify(594357)
					return err
				} else {
					__antithesis_instrumentation__.Notify(594360)
				}
			} else {
				__antithesis_instrumentation__.Notify(594361)
			}
		}
	}
	__antithesis_instrumentation__.Notify(594352)
	return nil
}

type EdgeIterator func(e Edge) error

func (g *Graph) ForEachEdge(it EdgeIterator) error {
	__antithesis_instrumentation__.Notify(594362)
	for _, e := range g.edges {
		__antithesis_instrumentation__.Notify(594364)
		if err := it(e); err != nil {
			__antithesis_instrumentation__.Notify(594365)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(594367)
				err = nil
			} else {
				__antithesis_instrumentation__.Notify(594368)
			}
			__antithesis_instrumentation__.Notify(594366)
			return err
		} else {
			__antithesis_instrumentation__.Notify(594369)
		}
	}
	__antithesis_instrumentation__.Notify(594363)
	return nil
}

type DepEdgeIterator func(de *DepEdge) error

func (g *Graph) ForEachDepEdgeFrom(n *screl.Node, it DepEdgeIterator) (err error) {
	__antithesis_instrumentation__.Notify(594370)
	return g.depEdgesFrom.iterateSourceNode(n, it)
}

func (g *Graph) ForEachDepEdgeTo(n *screl.Node, it DepEdgeIterator) (err error) {
	__antithesis_instrumentation__.Notify(594371)
	return g.depEdgesTo.iterateSourceNode(n, it)
}
