package opgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraph"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type registry struct {
	targets []target
}

var opRegistry = &registry{}

func BuildGraph(cs scpb.CurrentState) (*scgraph.Graph, error) {
	__antithesis_instrumentation__.Notify(594034)
	return opRegistry.buildGraph(cs)
}

func (r *registry) buildGraph(cs scpb.CurrentState) (_ *scgraph.Graph, err error) {
	__antithesis_instrumentation__.Notify(594035)
	start := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(594039)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(594041)
			return !log.V(2) == true
		}() == true {
			__antithesis_instrumentation__.Notify(594042)
			return
		} else {
			__antithesis_instrumentation__.Notify(594043)
		}
		__antithesis_instrumentation__.Notify(594040)
		log.Infof(context.TODO(), "operation graph generation took %v", timeutil.Since(start))
	}()
	__antithesis_instrumentation__.Notify(594036)
	g, err := scgraph.New(cs)
	if err != nil {
		__antithesis_instrumentation__.Notify(594044)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(594045)
	}
	__antithesis_instrumentation__.Notify(594037)

	type toAdd struct {
		transition
		n *screl.Node
	}
	var edgesToAdd []toAdd
	md := makeTargetsWithElementMap(cs)
	for _, t := range r.targets {
		__antithesis_instrumentation__.Notify(594046)
		edgesToAdd = edgesToAdd[:0]
		if err := t.iterateFunc(g.Database(), func(n *screl.Node) error {
			__antithesis_instrumentation__.Notify(594048)
			status := n.CurrentStatus
			for _, op := range t.transitions {
				__antithesis_instrumentation__.Notify(594050)
				if op.from == status {
					__antithesis_instrumentation__.Notify(594051)
					edgesToAdd = append(edgesToAdd, toAdd{
						transition: op,
						n:          n,
					})
					status = op.to
				} else {
					__antithesis_instrumentation__.Notify(594052)
				}
			}
			__antithesis_instrumentation__.Notify(594049)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(594053)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(594054)
		}
		__antithesis_instrumentation__.Notify(594047)
		for _, e := range edgesToAdd {
			__antithesis_instrumentation__.Notify(594055)
			var ops []scop.Op
			if e.ops != nil {
				__antithesis_instrumentation__.Notify(594057)
				ops = e.ops(e.n.Element(), md)
			} else {
				__antithesis_instrumentation__.Notify(594058)
			}
			__antithesis_instrumentation__.Notify(594056)
			if err := g.AddOpEdges(
				e.n.Target, e.from, e.to, e.revertible, e.minPhase, ops...,
			); err != nil {
				__antithesis_instrumentation__.Notify(594059)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(594060)
			}
		}

	}
	__antithesis_instrumentation__.Notify(594038)
	return g, nil
}
