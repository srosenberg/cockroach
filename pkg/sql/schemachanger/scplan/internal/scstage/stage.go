package scstage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraph"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/errors"
)

type Stage struct {
	Before, After []scpb.Status

	EdgeOps, ExtraOps []scop.Op

	Phase scop.Phase

	Ordinal, StagesInPhase int
}

func (s Stage) Type() scop.Type {
	__antithesis_instrumentation__.Notify(594683)
	if len(s.EdgeOps) == 0 {
		__antithesis_instrumentation__.Notify(594685)
		return scop.MutationType
	} else {
		__antithesis_instrumentation__.Notify(594686)
	}
	__antithesis_instrumentation__.Notify(594684)
	return s.EdgeOps[0].Type()
}

func (s Stage) Ops() []scop.Op {
	__antithesis_instrumentation__.Notify(594687)
	ops := make([]scop.Op, 0, len(s.EdgeOps)+len(s.ExtraOps))
	ops = append(ops, s.EdgeOps...)
	ops = append(ops, s.ExtraOps...)
	return ops
}

func (s Stage) String() string {
	__antithesis_instrumentation__.Notify(594688)
	ops := "no ops"
	if n := len(s.Ops()); n > 1 {
		__antithesis_instrumentation__.Notify(594690)
		ops = fmt.Sprintf("%d %s ops", n, s.Type())
	} else {
		__antithesis_instrumentation__.Notify(594691)
		if n == 1 {
			__antithesis_instrumentation__.Notify(594692)
			ops = fmt.Sprintf("1 %s op", s.Type())
		} else {
			__antithesis_instrumentation__.Notify(594693)
		}
	}
	__antithesis_instrumentation__.Notify(594689)
	return fmt.Sprintf("%s stage %d of %d with %s",
		s.Phase.String(), s.Ordinal, s.StagesInPhase, ops)
}

func ValidateStages(ts scpb.TargetState, stages []Stage, g *scgraph.Graph) error {
	__antithesis_instrumentation__.Notify(594694)
	if len(stages) == 0 {
		__antithesis_instrumentation__.Notify(594701)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(594702)
	}
	__antithesis_instrumentation__.Notify(594695)

	for _, stage := range stages {
		__antithesis_instrumentation__.Notify(594703)
		if na, nb := len(stage.After), len(stage.Before); na != nb {
			__antithesis_instrumentation__.Notify(594704)
			return errors.Errorf("%s: Before state has %d nodes and After state has %d nodes",
				stage, nb, na)
		} else {
			__antithesis_instrumentation__.Notify(594705)
		}
	}
	__antithesis_instrumentation__.Notify(594696)

	for i := range stages {
		__antithesis_instrumentation__.Notify(594706)
		if i == 0 {
			__antithesis_instrumentation__.Notify(594708)
			continue
		} else {
			__antithesis_instrumentation__.Notify(594709)
		}
		__antithesis_instrumentation__.Notify(594707)
		if err := validateAdjacentStagesStates(stages[i-1], stages[i]); err != nil {
			__antithesis_instrumentation__.Notify(594710)
			return errors.Wrapf(err, "stages %d and %d of %d", i, i+1, len(stages))
		} else {
			__antithesis_instrumentation__.Notify(594711)
		}
	}
	__antithesis_instrumentation__.Notify(594697)

	final := stages[len(stages)-1].After
	for i, actual := range final {
		__antithesis_instrumentation__.Notify(594712)
		expected := ts.Targets[i].TargetStatus
		if actual != expected {
			__antithesis_instrumentation__.Notify(594713)
			return errors.Errorf("final status is %s instead of %s at index %d for adding %s",
				actual, expected, i, screl.ElementString(ts.Targets[i].Element()))
		} else {
			__antithesis_instrumentation__.Notify(594714)
		}
	}
	__antithesis_instrumentation__.Notify(594698)

	currentPhase := scop.EarliestPhase
	for _, stage := range stages {
		__antithesis_instrumentation__.Notify(594715)
		if stage.Phase < currentPhase {
			__antithesis_instrumentation__.Notify(594716)
			return errors.Errorf("%s: preceded by %s stage",
				stage.String(), currentPhase)
		} else {
			__antithesis_instrumentation__.Notify(594717)
		}
	}
	__antithesis_instrumentation__.Notify(594699)

	for _, stage := range stages {
		__antithesis_instrumentation__.Notify(594718)
		if err := validateStageSubgraph(ts, stage, g); err != nil {
			__antithesis_instrumentation__.Notify(594719)
			return errors.Wrapf(err, "%s", stage.String())
		} else {
			__antithesis_instrumentation__.Notify(594720)
		}
	}
	__antithesis_instrumentation__.Notify(594700)
	return nil
}

func validateAdjacentStagesStates(previous, next Stage) error {
	__antithesis_instrumentation__.Notify(594721)
	if na, nb := len(previous.After), len(next.Before); na != nb {
		__antithesis_instrumentation__.Notify(594724)
		return errors.Errorf("node count mismatch: %d != %d",
			na, nb)
	} else {
		__antithesis_instrumentation__.Notify(594725)
	}
	__antithesis_instrumentation__.Notify(594722)
	for j, before := range next.Before {
		__antithesis_instrumentation__.Notify(594726)
		after := previous.After[j]
		if before != after {
			__antithesis_instrumentation__.Notify(594727)
			return errors.Errorf("node status mismatch at index %d: %s != %s",
				j, after.String(), before.String())
		} else {
			__antithesis_instrumentation__.Notify(594728)
		}
	}
	__antithesis_instrumentation__.Notify(594723)
	return nil
}

func validateStageSubgraph(ts scpb.TargetState, stage Stage, g *scgraph.Graph) error {
	__antithesis_instrumentation__.Notify(594729)

	var queue []*scgraph.OpEdge
	for _, op := range stage.EdgeOps {
		__antithesis_instrumentation__.Notify(594734)
		oe := g.GetOpEdgeFromOp(op)
		if oe == nil {
			__antithesis_instrumentation__.Notify(594736)

			return errors.Errorf("cannot find op edge for op %s", op)
		} else {
			__antithesis_instrumentation__.Notify(594737)
		}
		__antithesis_instrumentation__.Notify(594735)
		if len(queue) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(594738)
			return queue[len(queue)-1] != oe == true
		}() == true {
			__antithesis_instrumentation__.Notify(594739)
			queue = append(queue, oe)
		} else {
			__antithesis_instrumentation__.Notify(594740)
		}
	}
	__antithesis_instrumentation__.Notify(594730)

	fulfilled := map[*screl.Node]bool{}
	current := make([]*screl.Node, len(ts.Targets))
	for i, status := range stage.Before {
		__antithesis_instrumentation__.Notify(594741)
		t := &ts.Targets[i]
		n, ok := g.GetNode(t, status)
		if !ok {
			__antithesis_instrumentation__.Notify(594743)

			return errors.Errorf("cannot find starting node for %s", screl.ElementString(t.Element()))
		} else {
			__antithesis_instrumentation__.Notify(594744)
		}
		__antithesis_instrumentation__.Notify(594742)
		current[i] = n
	}
	{
		__antithesis_instrumentation__.Notify(594745)
		edgesTo := make(map[*screl.Node][]scgraph.Edge, g.Order())
		_ = g.ForEachEdge(func(e scgraph.Edge) error {
			__antithesis_instrumentation__.Notify(594748)
			edgesTo[e.To()] = append(edgesTo[e.To()], e)
			return nil
		})
		__antithesis_instrumentation__.Notify(594746)
		var dfs func(n *screl.Node)
		dfs = func(n *screl.Node) {
			__antithesis_instrumentation__.Notify(594749)
			if _, found := fulfilled[n]; found {
				__antithesis_instrumentation__.Notify(594751)
				return
			} else {
				__antithesis_instrumentation__.Notify(594752)
			}
			__antithesis_instrumentation__.Notify(594750)
			fulfilled[n] = true
			for _, e := range edgesTo[n] {
				__antithesis_instrumentation__.Notify(594753)
				dfs(e.From())
			}
		}
		__antithesis_instrumentation__.Notify(594747)
		for _, n := range current {
			__antithesis_instrumentation__.Notify(594754)
			dfs(n)
		}
	}
	__antithesis_instrumentation__.Notify(594731)

	for hasProgressed := true; hasProgressed; {
		__antithesis_instrumentation__.Notify(594755)
		hasProgressed = false

		for i, n := range current {
			__antithesis_instrumentation__.Notify(594756)
			if n.CurrentStatus == stage.After[i] {
				__antithesis_instrumentation__.Notify(594762)

				continue
			} else {
				__antithesis_instrumentation__.Notify(594763)
			}
			__antithesis_instrumentation__.Notify(594757)
			oe, ok := g.GetOpEdgeFrom(n)
			if !ok {
				__antithesis_instrumentation__.Notify(594764)

				return errors.Errorf("cannot find op-edge path from %s to %s for %s",
					stage.Before[i], stage.After[i], screl.ElementString(ts.Targets[i].Element()))
			} else {
				__antithesis_instrumentation__.Notify(594765)
			}
			__antithesis_instrumentation__.Notify(594758)

			var hasUnmetDeps bool
			if err := g.ForEachDepEdgeTo(oe.To(), func(de *scgraph.DepEdge) error {
				__antithesis_instrumentation__.Notify(594766)
				hasUnmetDeps = hasUnmetDeps || func() bool {
					__antithesis_instrumentation__.Notify(594767)
					return !fulfilled[de.From()] == true
				}() == true
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(594768)
				return err
			} else {
				__antithesis_instrumentation__.Notify(594769)
			}
			__antithesis_instrumentation__.Notify(594759)
			if hasUnmetDeps {
				__antithesis_instrumentation__.Notify(594770)
				continue
			} else {
				__antithesis_instrumentation__.Notify(594771)
			}
			__antithesis_instrumentation__.Notify(594760)

			if len(queue) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(594772)
				return oe == queue[0] == true
			}() == true {
				__antithesis_instrumentation__.Notify(594773)
				queue = queue[1:]
			} else {
				__antithesis_instrumentation__.Notify(594774)
				if !g.IsNoOp(oe) {
					__antithesis_instrumentation__.Notify(594775)
					continue
				} else {
					__antithesis_instrumentation__.Notify(594776)
				}
			}
			__antithesis_instrumentation__.Notify(594761)

			current[i] = oe.To()
			fulfilled[oe.To()] = true
			hasProgressed = true
		}
	}
	__antithesis_instrumentation__.Notify(594732)

	for i, n := range current {
		__antithesis_instrumentation__.Notify(594777)
		if n.CurrentStatus != stage.After[i] {
			__antithesis_instrumentation__.Notify(594778)
			return errors.Errorf("internal inconsistency, "+
				"ended in non-terminal status %s after walking the graph towards %s for %s",
				n.CurrentStatus, stage.After[i], screl.ElementString(ts.Targets[i].Element()))
		} else {
			__antithesis_instrumentation__.Notify(594779)
		}
	}
	__antithesis_instrumentation__.Notify(594733)

	return nil
}
