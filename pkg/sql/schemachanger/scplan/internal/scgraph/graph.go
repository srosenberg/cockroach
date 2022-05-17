package scgraph

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/errors"
)

type Graph struct {
	targets []*scpb.Target

	targetNodes []map[scpb.Status]*screl.Node

	targetIdxMap map[*scpb.Target]int

	opEdgesFrom map[*screl.Node]*OpEdge

	depEdgesFrom, depEdgesTo *depEdgeTree

	opToOpEdge map[scop.Op]*OpEdge

	noOpOpEdges map[*OpEdge]map[string]struct{}

	edges []Edge

	entities *rel.Database
}

func (g *Graph) Database() *rel.Database {
	__antithesis_instrumentation__.Notify(594256)
	return g.entities
}

func New(cs scpb.CurrentState) (*Graph, error) {
	__antithesis_instrumentation__.Notify(594257)
	db, err := rel.NewDatabase(screl.Schema, [][]rel.Attr{
		{screl.DescID, rel.Type, screl.ColumnID},
		{screl.ReferencedDescID, rel.Type},
		{rel.Type, screl.Element, screl.CurrentStatus},
		{rel.Type, screl.Target, screl.TargetStatus},
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(594260)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(594261)
	}
	__antithesis_instrumentation__.Notify(594258)
	g := Graph{
		targetIdxMap: map[*scpb.Target]int{},
		opEdgesFrom:  map[*screl.Node]*OpEdge{},
		noOpOpEdges:  map[*OpEdge]map[string]struct{}{},
		opToOpEdge:   map[scop.Op]*OpEdge{},
		entities:     db,
	}
	g.depEdgesFrom = newDepEdgeTree(fromTo, g.compareNodes)
	g.depEdgesTo = newDepEdgeTree(toFrom, g.compareNodes)
	for i, status := range cs.Current {
		__antithesis_instrumentation__.Notify(594262)
		t := &cs.Targets[i]
		if existing, ok := g.targetIdxMap[t]; ok {
			__antithesis_instrumentation__.Notify(594264)
			return nil, errors.Errorf("invalid initial state contains duplicate target: %v and %v", *t, cs.Targets[existing])
		} else {
			__antithesis_instrumentation__.Notify(594265)
		}
		__antithesis_instrumentation__.Notify(594263)
		idx := len(g.targets)
		g.targetIdxMap[t] = idx
		g.targets = append(g.targets, t)
		n := &screl.Node{Target: t, CurrentStatus: status}
		g.targetNodes = append(g.targetNodes, map[scpb.Status]*screl.Node{status: n})
		if err := g.entities.Insert(n); err != nil {
			__antithesis_instrumentation__.Notify(594266)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(594267)
		}
	}
	__antithesis_instrumentation__.Notify(594259)
	return &g, g.Validate()
}

func (g *Graph) ShallowClone() *Graph {
	__antithesis_instrumentation__.Notify(594268)

	clone := &Graph{
		targets:      g.targets,
		targetNodes:  g.targetNodes,
		targetIdxMap: g.targetIdxMap,
		opEdgesFrom:  g.opEdgesFrom,
		depEdgesFrom: g.depEdgesFrom,
		depEdgesTo:   g.depEdgesTo,
		opToOpEdge:   g.opToOpEdge,
		edges:        g.edges,
		entities:     g.entities,
		noOpOpEdges:  make(map[*OpEdge]map[string]struct{}),
	}

	for edge, noop := range g.noOpOpEdges {
		__antithesis_instrumentation__.Notify(594270)
		clone.noOpOpEdges[edge] = noop
	}
	__antithesis_instrumentation__.Notify(594269)
	return clone
}

func (g *Graph) GetNode(t *scpb.Target, s scpb.Status) (*screl.Node, bool) {
	__antithesis_instrumentation__.Notify(594271)
	targetStatuses := g.getTargetStatusMap(t)
	ts, ok := targetStatuses[s]
	return ts, ok
}

var _ = (*Graph)(nil).GetNode

func (g *Graph) getOrCreateNode(t *scpb.Target, s scpb.Status) (*screl.Node, error) {
	__antithesis_instrumentation__.Notify(594272)
	targetStatuses := g.getTargetStatusMap(t)
	if ts, ok := targetStatuses[s]; ok {
		__antithesis_instrumentation__.Notify(594275)
		return ts, nil
	} else {
		__antithesis_instrumentation__.Notify(594276)
	}
	__antithesis_instrumentation__.Notify(594273)
	ts := &screl.Node{
		Target:        t,
		CurrentStatus: s,
	}
	targetStatuses[s] = ts
	if err := g.entities.Insert(ts); err != nil {
		__antithesis_instrumentation__.Notify(594277)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(594278)
	}
	__antithesis_instrumentation__.Notify(594274)
	return ts, nil
}

func (g *Graph) getTargetStatusMap(target *scpb.Target) map[scpb.Status]*screl.Node {
	__antithesis_instrumentation__.Notify(594279)
	idx, ok := g.targetIdxMap[target]
	if !ok {
		__antithesis_instrumentation__.Notify(594281)
		panic(errors.Errorf("target %v does not exist", target))
	} else {
		__antithesis_instrumentation__.Notify(594282)
	}
	__antithesis_instrumentation__.Notify(594280)
	return g.targetNodes[idx]
}

func (g *Graph) containsTarget(target *scpb.Target) bool {
	__antithesis_instrumentation__.Notify(594283)
	_, exists := g.targetIdxMap[target]
	return exists
}

var _ = (*Graph)(nil).containsTarget

func (g *Graph) GetOpEdgeFrom(n *screl.Node) (*OpEdge, bool) {
	__antithesis_instrumentation__.Notify(594284)
	oe, ok := g.opEdgesFrom[n]
	return oe, ok
}

func (g *Graph) AddOpEdges(
	t *scpb.Target, from, to scpb.Status, revertible bool, minPhase scop.Phase, ops ...scop.Op,
) (err error) {
	__antithesis_instrumentation__.Notify(594285)

	oe := &OpEdge{
		op:         ops,
		revertible: revertible,
		minPhase:   minPhase,
	}
	if oe.from, err = g.getOrCreateNode(t, from); err != nil {
		__antithesis_instrumentation__.Notify(594291)
		return err
	} else {
		__antithesis_instrumentation__.Notify(594292)
	}
	__antithesis_instrumentation__.Notify(594286)
	if oe.to, err = g.getOrCreateNode(t, to); err != nil {
		__antithesis_instrumentation__.Notify(594293)
		return err
	} else {
		__antithesis_instrumentation__.Notify(594294)
	}
	__antithesis_instrumentation__.Notify(594287)
	if existing, exists := g.opEdgesFrom[oe.from]; exists {
		__antithesis_instrumentation__.Notify(594295)
		return errors.Errorf("duplicate outbound op edge %v and %v",
			oe, existing)
	} else {
		__antithesis_instrumentation__.Notify(594296)
	}
	__antithesis_instrumentation__.Notify(594288)
	g.edges = append(g.edges, oe)
	typ := scop.MutationType
	for i, op := range ops {
		__antithesis_instrumentation__.Notify(594297)
		if i == 0 {
			__antithesis_instrumentation__.Notify(594298)
			typ = op.Type()
		} else {
			__antithesis_instrumentation__.Notify(594299)
			if typ != op.Type() {
				__antithesis_instrumentation__.Notify(594300)
				return errors.Errorf("mixed type for opEdge %s->%s, %s != %s",
					screl.NodeString(oe.from), screl.NodeString(oe.to), typ, op.Type())
			} else {
				__antithesis_instrumentation__.Notify(594301)
			}
		}
	}
	__antithesis_instrumentation__.Notify(594289)
	oe.typ = typ
	g.opEdgesFrom[oe.from] = oe

	for _, op := range ops {
		__antithesis_instrumentation__.Notify(594302)
		g.opToOpEdge[op] = oe
	}
	__antithesis_instrumentation__.Notify(594290)
	return nil
}

func (g *Graph) GetOpEdgeFromOp(op scop.Op) *OpEdge {
	__antithesis_instrumentation__.Notify(594303)
	return g.opToOpEdge[op]
}

func (g *Graph) AddDepEdge(
	rule string,
	kind DepEdgeKind,
	fromTarget *scpb.Target,
	fromStatus scpb.Status,
	toTarget *scpb.Target,
	toStatus scpb.Status,
) (err error) {
	__antithesis_instrumentation__.Notify(594304)
	de := &DepEdge{rule: rule, kind: kind}
	if de.from, err = g.getOrCreateNode(fromTarget, fromStatus); err != nil {
		__antithesis_instrumentation__.Notify(594307)
		return err
	} else {
		__antithesis_instrumentation__.Notify(594308)
	}
	__antithesis_instrumentation__.Notify(594305)
	if de.to, err = g.getOrCreateNode(toTarget, toStatus); err != nil {
		__antithesis_instrumentation__.Notify(594309)
		return err
	} else {
		__antithesis_instrumentation__.Notify(594310)
	}
	__antithesis_instrumentation__.Notify(594306)
	g.edges = append(g.edges, de)
	g.depEdgesFrom.insert(de)
	g.depEdgesTo.insert(de)
	return nil
}

func (g *Graph) MarkAsNoOp(edge *OpEdge, rule ...string) {
	__antithesis_instrumentation__.Notify(594311)
	m := make(map[string]struct{})
	for _, r := range rule {
		__antithesis_instrumentation__.Notify(594313)
		m[r] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(594312)
	g.noOpOpEdges[edge] = m
}

func (g *Graph) IsNoOp(edge *OpEdge) bool {
	__antithesis_instrumentation__.Notify(594314)
	if len(edge.op) == 0 {
		__antithesis_instrumentation__.Notify(594316)
		return true
	} else {
		__antithesis_instrumentation__.Notify(594317)
	}
	__antithesis_instrumentation__.Notify(594315)
	_, isNoOp := g.noOpOpEdges[edge]
	return isNoOp
}

func (g *Graph) NoOpRules(edge *OpEdge) (rules []string) {
	__antithesis_instrumentation__.Notify(594318)
	if !g.IsNoOp(edge) {
		__antithesis_instrumentation__.Notify(594321)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(594322)
	}
	__antithesis_instrumentation__.Notify(594319)
	m := g.noOpOpEdges[edge]
	for rule := range m {
		__antithesis_instrumentation__.Notify(594323)
		rules = append(rules, rule)
	}
	__antithesis_instrumentation__.Notify(594320)
	sort.Strings(rules)
	return rules
}

func (g *Graph) Order() int {
	__antithesis_instrumentation__.Notify(594324)
	n := 0
	for _, m := range g.targetNodes {
		__antithesis_instrumentation__.Notify(594326)
		n = n + len(m)
	}
	__antithesis_instrumentation__.Notify(594325)
	return n
}

func (g *Graph) Validate() (err error) {
	__antithesis_instrumentation__.Notify(594327)
	marks := make(map[*screl.Node]bool, g.Order())
	var visit func(n *screl.Node)
	visit = func(n *screl.Node) {
		__antithesis_instrumentation__.Notify(594330)
		if err != nil {
			__antithesis_instrumentation__.Notify(594335)
			return
		} else {
			__antithesis_instrumentation__.Notify(594336)
		}
		__antithesis_instrumentation__.Notify(594331)
		permanent, marked := marks[n]
		if marked && func() bool {
			__antithesis_instrumentation__.Notify(594337)
			return !permanent == true
		}() == true {
			__antithesis_instrumentation__.Notify(594338)
			err = errors.AssertionFailedf("graph is not acyclical")
			return
		} else {
			__antithesis_instrumentation__.Notify(594339)
		}
		__antithesis_instrumentation__.Notify(594332)
		if marked && func() bool {
			__antithesis_instrumentation__.Notify(594340)
			return permanent == true
		}() == true {
			__antithesis_instrumentation__.Notify(594341)
			return
		} else {
			__antithesis_instrumentation__.Notify(594342)
		}
		__antithesis_instrumentation__.Notify(594333)
		marks[n] = false
		_ = g.ForEachDepEdgeTo(n, func(de *DepEdge) error {
			__antithesis_instrumentation__.Notify(594343)
			visit(de.From())
			return nil
		})
		__antithesis_instrumentation__.Notify(594334)
		marks[n] = true
	}
	__antithesis_instrumentation__.Notify(594328)
	_ = g.ForEachNode(func(n *screl.Node) error {
		__antithesis_instrumentation__.Notify(594344)
		visit(n)
		return nil
	})
	__antithesis_instrumentation__.Notify(594329)
	return err
}

func (g *Graph) compareNodes(a, b *screl.Node) (less, eq bool) {
	__antithesis_instrumentation__.Notify(594345)
	switch {
	case a == b:
		__antithesis_instrumentation__.Notify(594346)
		return false, true
	case a == nil:
		__antithesis_instrumentation__.Notify(594347)
		return true, false
	case b == nil:
		__antithesis_instrumentation__.Notify(594348)
		return false, false
	case a.Target == b.Target:
		__antithesis_instrumentation__.Notify(594349)
		return a.CurrentStatus < b.CurrentStatus, a.CurrentStatus == b.CurrentStatus
	default:
		__antithesis_instrumentation__.Notify(594350)
		aIdx, bIdx := g.targetIdxMap[a.Target], g.targetIdxMap[b.Target]
		return aIdx < bIdx, aIdx == bIdx
	}
}
