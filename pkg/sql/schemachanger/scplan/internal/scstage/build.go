package scstage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraph"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
)

func BuildStages(
	init scpb.CurrentState, phase scop.Phase, g *scgraph.Graph, scJobIDSupplier func() jobspb.JobID,
) []Stage {
	__antithesis_instrumentation__.Notify(594485)
	c := buildContext{
		rollback:               init.InRollback,
		g:                      g,
		scJobIDSupplier:        scJobIDSupplier,
		isRevertibilityIgnored: true,
		targetState:            init.TargetState,
		startingStatuses:       init.Current,
		startingPhase:          phase,
	}

	stages := buildStages(c)
	if n := len(stages); n > 0 && func() bool {
		__antithesis_instrumentation__.Notify(594489)
		return stages[n-1].Phase > scop.PreCommitPhase == true
	}() == true {
		__antithesis_instrumentation__.Notify(594490)
		c.isRevertibilityIgnored = false
		stages = buildStages(c)
	} else {
		__antithesis_instrumentation__.Notify(594491)
	}
	__antithesis_instrumentation__.Notify(594486)

	if len(stages) > 0 {
		__antithesis_instrumentation__.Notify(594492)
		phaseMap := map[scop.Phase][]int{}
		for i, s := range stages {
			__antithesis_instrumentation__.Notify(594494)
			phaseMap[s.Phase] = append(phaseMap[s.Phase], i)
		}
		__antithesis_instrumentation__.Notify(594493)
		for _, indexes := range phaseMap {
			__antithesis_instrumentation__.Notify(594495)
			for i, j := range indexes {
				__antithesis_instrumentation__.Notify(594496)
				s := &stages[j]
				s.Ordinal = i + 1
				s.StagesInPhase = len(indexes)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(594497)
	}
	__antithesis_instrumentation__.Notify(594487)

	for i := range stages {
		__antithesis_instrumentation__.Notify(594498)
		var cur, next *Stage
		if i+1 < len(stages) {
			__antithesis_instrumentation__.Notify(594500)
			next = &stages[i+1]
		} else {
			__antithesis_instrumentation__.Notify(594501)
		}
		__antithesis_instrumentation__.Notify(594499)
		cur = &stages[i]
		jobOps := c.computeExtraJobOps(cur, next)
		cur.ExtraOps = jobOps
	}
	__antithesis_instrumentation__.Notify(594488)

	return stages
}

type buildContext struct {
	rollback               bool
	g                      *scgraph.Graph
	scJobIDSupplier        func() jobspb.JobID
	isRevertibilityIgnored bool
	targetState            scpb.TargetState
	startingStatuses       []scpb.Status
	startingPhase          scop.Phase
}

func buildStages(bc buildContext) (stages []Stage) {
	__antithesis_instrumentation__.Notify(594502)

	bs := buildState{
		incumbent: make([]scpb.Status, len(bc.startingStatuses)),
		phase:     bc.startingPhase,
		fulfilled: make(map[*screl.Node]struct{}, bc.g.Order()),
	}
	for i, n := range bc.nodes(bc.startingStatuses) {
		__antithesis_instrumentation__.Notify(594505)
		bs.incumbent[i] = n.CurrentStatus
		bs.fulfilled[n] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(594503)

	for !bc.isStateTerminal(bs.incumbent) {
		__antithesis_instrumentation__.Notify(594506)

		sb := bc.makeStageBuilder(bs)
		for !sb.canMakeProgress() {
			__antithesis_instrumentation__.Notify(594509)

			if bs.phase == scop.PreCommitPhase {
				__antithesis_instrumentation__.Notify(594512)

				break
			} else {
				__antithesis_instrumentation__.Notify(594513)
			}
			__antithesis_instrumentation__.Notify(594510)
			if bs.phase == scop.LatestPhase {
				__antithesis_instrumentation__.Notify(594514)

				panic(errors.AssertionFailedf("unable to make progress"))
			} else {
				__antithesis_instrumentation__.Notify(594515)
			}
			__antithesis_instrumentation__.Notify(594511)
			bs.phase++
			sb = bc.makeStageBuilder(bs)
		}
		__antithesis_instrumentation__.Notify(594507)

		stage := sb.build()
		stages = append(stages, stage)

		for n := range sb.fulfilling {
			__antithesis_instrumentation__.Notify(594516)
			bs.fulfilled[n] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(594508)
		bs.incumbent = stage.After
		switch bs.phase {
		case scop.StatementPhase, scop.PreCommitPhase:
			__antithesis_instrumentation__.Notify(594517)

			bs.phase++
		default:
			__antithesis_instrumentation__.Notify(594518)
		}
	}
	__antithesis_instrumentation__.Notify(594504)

	return stages
}

type buildState struct {
	incumbent []scpb.Status
	phase     scop.Phase
	fulfilled map[*screl.Node]struct{}
}

func (bc buildContext) isStateTerminal(current []scpb.Status) bool {
	__antithesis_instrumentation__.Notify(594519)
	for _, n := range bc.nodes(current) {
		__antithesis_instrumentation__.Notify(594521)
		if _, found := bc.g.GetOpEdgeFrom(n); found {
			__antithesis_instrumentation__.Notify(594522)
			return false
		} else {
			__antithesis_instrumentation__.Notify(594523)
		}
	}
	__antithesis_instrumentation__.Notify(594520)
	return true
}

func (bc buildContext) makeStageBuilder(bs buildState) (sb stageBuilder) {
	__antithesis_instrumentation__.Notify(594524)
	opTypes := []scop.Type{scop.BackfillType, scop.ValidationType, scop.MutationType}
	switch bs.phase {
	case scop.StatementPhase, scop.PreCommitPhase:
		__antithesis_instrumentation__.Notify(594527)

		opTypes = []scop.Type{scop.MutationType}
	default:
		__antithesis_instrumentation__.Notify(594528)
	}
	__antithesis_instrumentation__.Notify(594525)
	for _, opType := range opTypes {
		__antithesis_instrumentation__.Notify(594529)
		sb = bc.makeStageBuilderForType(bs, opType)
		if sb.canMakeProgress() {
			__antithesis_instrumentation__.Notify(594530)
			break
		} else {
			__antithesis_instrumentation__.Notify(594531)
		}
	}
	__antithesis_instrumentation__.Notify(594526)
	return sb
}

func (bc buildContext) makeStageBuilderForType(bs buildState, opType scop.Type) stageBuilder {
	__antithesis_instrumentation__.Notify(594532)
	numTargets := len(bc.targetState.Targets)
	sb := stageBuilder{
		bc:         bc,
		bs:         bs,
		opType:     opType,
		current:    make([]currentTargetState, numTargets),
		fulfilling: map[*screl.Node]struct{}{},
		lut:        make(map[*scpb.Target]*currentTargetState, numTargets),
		visited:    make(map[*screl.Node]uint64, numTargets),
	}
	for i, n := range bc.nodes(bs.incumbent) {
		__antithesis_instrumentation__.Notify(594535)
		t := sb.makeCurrentTargetState(n)
		sb.current[i] = t
		sb.lut[t.n.Target] = &sb.current[i]
	}
	__antithesis_instrumentation__.Notify(594533)

	for isDone := false; !isDone; {
		__antithesis_instrumentation__.Notify(594536)
		isDone = true
		for i, t := range sb.current {
			__antithesis_instrumentation__.Notify(594537)
			if t.e == nil {
				__antithesis_instrumentation__.Notify(594541)
				continue
			} else {
				__antithesis_instrumentation__.Notify(594542)
			}
			__antithesis_instrumentation__.Notify(594538)
			if sb.hasUnmetInboundDeps(t.e.To()) {
				__antithesis_instrumentation__.Notify(594543)
				continue
			} else {
				__antithesis_instrumentation__.Notify(594544)
			}
			__antithesis_instrumentation__.Notify(594539)

			sb.visitEpoch++
			if sb.hasUnmeetableOutboundDeps(t.e.To()) {
				__antithesis_instrumentation__.Notify(594545)
				continue
			} else {
				__antithesis_instrumentation__.Notify(594546)
			}
			__antithesis_instrumentation__.Notify(594540)
			sb.opEdges = append(sb.opEdges, t.e)
			sb.fulfilling[t.e.To()] = struct{}{}
			sb.current[i] = sb.nextTargetState(t)
			isDone = false
		}
	}
	__antithesis_instrumentation__.Notify(594534)
	return sb
}

type stageBuilder struct {
	bc         buildContext
	bs         buildState
	opType     scop.Type
	current    []currentTargetState
	fulfilling map[*screl.Node]struct{}
	opEdges    []*scgraph.OpEdge

	lut        map[*scpb.Target]*currentTargetState
	visited    map[*screl.Node]uint64
	visitEpoch uint64
}

type currentTargetState struct {
	n *screl.Node
	e *scgraph.OpEdge

	hasOpEdgeWithOps bool
}

func (sb stageBuilder) makeCurrentTargetState(n *screl.Node) currentTargetState {
	__antithesis_instrumentation__.Notify(594547)
	e, found := sb.bc.g.GetOpEdgeFrom(n)
	if !found || func() bool {
		__antithesis_instrumentation__.Notify(594549)
		return !sb.isOutgoingOpEdgeAllowed(e) == true
	}() == true {
		__antithesis_instrumentation__.Notify(594550)
		return currentTargetState{n: n}
	} else {
		__antithesis_instrumentation__.Notify(594551)
	}
	__antithesis_instrumentation__.Notify(594548)
	return currentTargetState{
		n:                n,
		e:                e,
		hasOpEdgeWithOps: !sb.bc.g.IsNoOp(e),
	}
}

func (sb stageBuilder) isOutgoingOpEdgeAllowed(e *scgraph.OpEdge) bool {
	__antithesis_instrumentation__.Notify(594552)
	if _, isFulfilled := sb.bs.fulfilled[e.To()]; isFulfilled {
		__antithesis_instrumentation__.Notify(594558)
		panic(errors.AssertionFailedf(
			"node %s is unexpectedly already fulfilled in a previous stage",
			screl.NodeString(e.To())))
	} else {
		__antithesis_instrumentation__.Notify(594559)
	}
	__antithesis_instrumentation__.Notify(594553)
	if _, isCandidate := sb.fulfilling[e.To()]; isCandidate {
		__antithesis_instrumentation__.Notify(594560)
		panic(errors.AssertionFailedf(
			"node %s is unexpectedly already scheduled to be fulfilled in the upcoming stage",
			screl.NodeString(e.To())))
	} else {
		__antithesis_instrumentation__.Notify(594561)
	}
	__antithesis_instrumentation__.Notify(594554)
	if e.Type() != sb.opType {
		__antithesis_instrumentation__.Notify(594562)
		return false
	} else {
		__antithesis_instrumentation__.Notify(594563)
	}
	__antithesis_instrumentation__.Notify(594555)
	if !e.IsPhaseSatisfied(sb.bs.phase) {
		__antithesis_instrumentation__.Notify(594564)
		return false
	} else {
		__antithesis_instrumentation__.Notify(594565)
	}
	__antithesis_instrumentation__.Notify(594556)
	if !sb.bc.isRevertibilityIgnored && func() bool {
		__antithesis_instrumentation__.Notify(594566)
		return sb.bs.phase == scop.PostCommitPhase == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(594567)
		return !e.Revertible() == true
	}() == true {
		__antithesis_instrumentation__.Notify(594568)
		return false
	} else {
		__antithesis_instrumentation__.Notify(594569)
	}
	__antithesis_instrumentation__.Notify(594557)
	return true
}

func (sb stageBuilder) canMakeProgress() bool {
	__antithesis_instrumentation__.Notify(594570)
	return len(sb.opEdges) > 0
}

func (sb stageBuilder) nextTargetState(t currentTargetState) currentTargetState {
	__antithesis_instrumentation__.Notify(594571)
	next := sb.makeCurrentTargetState(t.e.To())
	if t.hasOpEdgeWithOps {
		__antithesis_instrumentation__.Notify(594573)
		if next.hasOpEdgeWithOps {
			__antithesis_instrumentation__.Notify(594574)

			next.e = nil
		} else {
			__antithesis_instrumentation__.Notify(594575)
			next.hasOpEdgeWithOps = true
		}
	} else {
		__antithesis_instrumentation__.Notify(594576)
	}
	__antithesis_instrumentation__.Notify(594572)
	return next
}

func (sb stageBuilder) hasUnmetInboundDeps(n *screl.Node) (ret bool) {
	__antithesis_instrumentation__.Notify(594577)
	_ = sb.bc.g.ForEachDepEdgeTo(n, func(de *scgraph.DepEdge) error {
		__antithesis_instrumentation__.Notify(594579)
		if sb.isUnmetInboundDep(de) {
			__antithesis_instrumentation__.Notify(594581)
			ret = true
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(594582)
		}
		__antithesis_instrumentation__.Notify(594580)
		return nil
	})
	__antithesis_instrumentation__.Notify(594578)
	return ret
}

func (sb *stageBuilder) isUnmetInboundDep(de *scgraph.DepEdge) bool {
	__antithesis_instrumentation__.Notify(594583)
	_, fromIsFulfilled := sb.bs.fulfilled[de.From()]
	_, fromIsCandidate := sb.fulfilling[de.From()]
	switch de.Kind() {
	case scgraph.Precedence:
		__antithesis_instrumentation__.Notify(594585)

		return !fromIsFulfilled && func() bool {
			__antithesis_instrumentation__.Notify(594589)
			return !fromIsCandidate == true
		}() == true

	case scgraph.SameStagePrecedence:
		__antithesis_instrumentation__.Notify(594586)
		if fromIsFulfilled {
			__antithesis_instrumentation__.Notify(594590)

			break
		} else {
			__antithesis_instrumentation__.Notify(594591)
		}
		__antithesis_instrumentation__.Notify(594587)

		return !fromIsCandidate

	default:
		__antithesis_instrumentation__.Notify(594588)
		panic(errors.AssertionFailedf("unknown dep edge kind %q", de.Kind()))
	}
	__antithesis_instrumentation__.Notify(594584)

	panic(errors.AssertionFailedf("failed to satisfy %s rule %q",
		de.String(), de.Name()))
}

func (sb stageBuilder) hasUnmeetableOutboundDeps(n *screl.Node) (ret bool) {
	__antithesis_instrumentation__.Notify(594592)

	if sb.visited[n] == sb.visitEpoch {
		__antithesis_instrumentation__.Notify(594600)

		return false
	} else {
		__antithesis_instrumentation__.Notify(594601)
	}
	__antithesis_instrumentation__.Notify(594593)

	sb.visited[n] = sb.visitEpoch

	if _, isFulfilled := sb.bs.fulfilled[n]; isFulfilled {
		__antithesis_instrumentation__.Notify(594602)

		panic(errors.AssertionFailedf("%s should not yet be fulfilled",
			screl.NodeString(n)))
	} else {
		__antithesis_instrumentation__.Notify(594603)
	}
	__antithesis_instrumentation__.Notify(594594)
	if _, isFulfilling := sb.bs.fulfilled[n]; isFulfilling {
		__antithesis_instrumentation__.Notify(594604)

		panic(errors.AssertionFailedf("%s should not yet be scheduled for this stage",
			screl.NodeString(n)))
	} else {
		__antithesis_instrumentation__.Notify(594605)
	}
	__antithesis_instrumentation__.Notify(594595)

	if t := sb.lut[n.Target]; t == nil {
		__antithesis_instrumentation__.Notify(594606)

		panic(errors.AssertionFailedf("%s target not found in look-up table",
			screl.NodeString(n)))
	} else {
		__antithesis_instrumentation__.Notify(594607)
		if t.e == nil || func() bool {
			__antithesis_instrumentation__.Notify(594608)
			return t.e.To() != n == true
		}() == true {
			__antithesis_instrumentation__.Notify(594609)

			return true
		} else {
			__antithesis_instrumentation__.Notify(594610)
		}
	}
	__antithesis_instrumentation__.Notify(594596)

	_ = sb.bc.g.ForEachDepEdgeTo(n, func(de *scgraph.DepEdge) error {
		__antithesis_instrumentation__.Notify(594611)
		if sb.visited[de.From()] == sb.visitEpoch {
			__antithesis_instrumentation__.Notify(594615)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(594616)
		}
		__antithesis_instrumentation__.Notify(594612)
		if !sb.isUnmetInboundDep(de) {
			__antithesis_instrumentation__.Notify(594617)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(594618)
		}
		__antithesis_instrumentation__.Notify(594613)
		if de.Kind() != scgraph.SameStagePrecedence || func() bool {
			__antithesis_instrumentation__.Notify(594619)
			return sb.hasUnmeetableOutboundDeps(de.From()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(594620)
			ret = true
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(594621)
		}
		__antithesis_instrumentation__.Notify(594614)
		return nil
	})
	__antithesis_instrumentation__.Notify(594597)
	if ret {
		__antithesis_instrumentation__.Notify(594622)
		return true
	} else {
		__antithesis_instrumentation__.Notify(594623)
	}
	__antithesis_instrumentation__.Notify(594598)
	_ = sb.bc.g.ForEachDepEdgeFrom(n, func(de *scgraph.DepEdge) error {
		__antithesis_instrumentation__.Notify(594624)
		if sb.visited[de.To()] == sb.visitEpoch {
			__antithesis_instrumentation__.Notify(594627)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(594628)
		}
		__antithesis_instrumentation__.Notify(594625)
		if de.Kind() == scgraph.SameStagePrecedence && func() bool {
			__antithesis_instrumentation__.Notify(594629)
			return sb.hasUnmeetableOutboundDeps(de.To()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(594630)
			ret = true
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(594631)
		}
		__antithesis_instrumentation__.Notify(594626)
		return nil
	})
	__antithesis_instrumentation__.Notify(594599)
	return ret
}

func (sb stageBuilder) build() Stage {
	__antithesis_instrumentation__.Notify(594632)
	after := make([]scpb.Status, len(sb.current))
	for i, t := range sb.current {
		__antithesis_instrumentation__.Notify(594635)
		after[i] = t.n.CurrentStatus
	}
	__antithesis_instrumentation__.Notify(594633)
	s := Stage{
		Before: sb.bs.incumbent,
		After:  after,
		Phase:  sb.bs.phase,
	}
	for _, e := range sb.opEdges {
		__antithesis_instrumentation__.Notify(594636)
		if sb.bc.g.IsNoOp(e) {
			__antithesis_instrumentation__.Notify(594638)
			continue
		} else {
			__antithesis_instrumentation__.Notify(594639)
		}
		__antithesis_instrumentation__.Notify(594637)
		s.EdgeOps = append(s.EdgeOps, e.Op()...)
	}
	__antithesis_instrumentation__.Notify(594634)
	return s
}

func (bc buildContext) computeExtraJobOps(s, next *Stage) []scop.Op {
	__antithesis_instrumentation__.Notify(594640)
	revertible := next != nil && func() bool {
		__antithesis_instrumentation__.Notify(594641)
		return next.Phase < scop.PostCommitNonRevertiblePhase == true
	}() == true
	switch s.Phase {
	case scop.PreCommitPhase:
		__antithesis_instrumentation__.Notify(594642)

		if next != nil {
			__antithesis_instrumentation__.Notify(594649)
			const initialize = true
			runningStatus := fmt.Sprintf("%s pending", next)
			return append(bc.setJobStateOnDescriptorOps(initialize, revertible, s.After),
				bc.createSchemaChangeJobOp(revertible, runningStatus))
		} else {
			__antithesis_instrumentation__.Notify(594650)
		}
		__antithesis_instrumentation__.Notify(594643)
		return nil
	case scop.PostCommitPhase, scop.PostCommitNonRevertiblePhase:
		__antithesis_instrumentation__.Notify(594644)
		if s.Type() != scop.MutationType {
			__antithesis_instrumentation__.Notify(594651)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(594652)
		}
		__antithesis_instrumentation__.Notify(594645)
		var ops []scop.Op
		if next == nil {
			__antithesis_instrumentation__.Notify(594653)

			ops = bc.removeJobReferenceOps()
		} else {
			__antithesis_instrumentation__.Notify(594654)
			const initialize = false
			ops = bc.setJobStateOnDescriptorOps(initialize, revertible, s.After)
		}
		__antithesis_instrumentation__.Notify(594646)

		runningStatus := "all stages completed"
		if next != nil {
			__antithesis_instrumentation__.Notify(594655)
			runningStatus = fmt.Sprintf("%s pending", next)
		} else {
			__antithesis_instrumentation__.Notify(594656)
		}
		__antithesis_instrumentation__.Notify(594647)
		ops = append(ops, bc.updateJobProgressOp(revertible, runningStatus))
		return ops
	default:
		__antithesis_instrumentation__.Notify(594648)
		return nil
	}
}

func (bc buildContext) createSchemaChangeJobOp(revertible bool, runningStatus string) scop.Op {
	__antithesis_instrumentation__.Notify(594657)
	return &scop.CreateSchemaChangerJob{
		JobID:         bc.scJobIDSupplier(),
		Statements:    bc.targetState.Statements,
		Authorization: bc.targetState.Authorization,
		DescriptorIDs: screl.AllTargetDescIDs(bc.targetState).Ordered(),
		NonCancelable: !revertible,
		RunningStatus: runningStatus,
	}
}

func (bc buildContext) updateJobProgressOp(revertible bool, runningStatus string) scop.Op {
	__antithesis_instrumentation__.Notify(594658)
	return &scop.UpdateSchemaChangerJob{
		JobID:           bc.scJobIDSupplier(),
		IsNonCancelable: !revertible,
		RunningStatus:   runningStatus,
	}
}

func (bc buildContext) removeJobReferenceOps() (ops []scop.Op) {
	__antithesis_instrumentation__.Notify(594659)
	jobID := bc.scJobIDSupplier()
	screl.AllTargetDescIDs(bc.targetState).ForEach(func(descID descpb.ID) {
		__antithesis_instrumentation__.Notify(594661)
		ops = append(ops, &scop.RemoveJobStateFromDescriptor{
			DescriptorID: descID,
			JobID:        jobID,
		})
	})
	__antithesis_instrumentation__.Notify(594660)
	return ops
}

func (bc buildContext) nodes(current []scpb.Status) []*screl.Node {
	__antithesis_instrumentation__.Notify(594662)
	nodes := make([]*screl.Node, len(bc.targetState.Targets))
	for i, status := range current {
		__antithesis_instrumentation__.Notify(594664)
		t := &bc.targetState.Targets[i]
		n, ok := bc.g.GetNode(t, status)
		if !ok {
			__antithesis_instrumentation__.Notify(594666)
			panic(errors.AssertionFailedf("could not find node for element %s, target status %s, current status %s",
				screl.ElementString(t.Element()), t.TargetStatus, status))
		} else {
			__antithesis_instrumentation__.Notify(594667)
		}
		__antithesis_instrumentation__.Notify(594665)
		nodes[i] = n
	}
	__antithesis_instrumentation__.Notify(594663)
	return nodes
}

func (bc buildContext) setJobStateOnDescriptorOps(
	initialize, revertible bool, after []scpb.Status,
) []scop.Op {
	__antithesis_instrumentation__.Notify(594668)
	descIDs, states := makeDescriptorStates(
		bc.scJobIDSupplier(), bc.rollback, revertible, bc.targetState, after,
	)
	ops := make([]scop.Op, 0, descIDs.Len())
	descIDs.ForEach(func(descID descpb.ID) {
		__antithesis_instrumentation__.Notify(594670)
		ops = append(ops, &scop.SetJobStateOnDescriptor{
			DescriptorID: descID,
			Initialize:   initialize,
			State:        *states[descID],
		})
	})
	__antithesis_instrumentation__.Notify(594669)
	return ops
}

func makeDescriptorStates(
	jobID jobspb.JobID, inRollback, revertible bool, ts scpb.TargetState, statuses []scpb.Status,
) (catalog.DescriptorIDSet, map[descpb.ID]*scpb.DescriptorState) {
	__antithesis_instrumentation__.Notify(594671)
	descIDs := screl.AllTargetDescIDs(ts)
	states := make(map[descpb.ID]*scpb.DescriptorState, descIDs.Len())
	descIDs.ForEach(func(id descpb.ID) {
		__antithesis_instrumentation__.Notify(594675)
		states[id] = &scpb.DescriptorState{
			Authorization: ts.Authorization,
			JobID:         jobID,
			InRollback:    inRollback,
			Revertible:    revertible,
		}
	})
	__antithesis_instrumentation__.Notify(594672)
	noteRelevantStatement := func(state *scpb.DescriptorState, stmtRank uint32) {
		__antithesis_instrumentation__.Notify(594676)
		for i := range state.RelevantStatements {
			__antithesis_instrumentation__.Notify(594678)
			stmt := &state.RelevantStatements[i]
			if stmt.StatementRank != stmtRank {
				__antithesis_instrumentation__.Notify(594680)
				continue
			} else {
				__antithesis_instrumentation__.Notify(594681)
			}
			__antithesis_instrumentation__.Notify(594679)
			return
		}
		__antithesis_instrumentation__.Notify(594677)
		state.RelevantStatements = append(state.RelevantStatements,
			scpb.DescriptorState_Statement{
				Statement:     ts.Statements[stmtRank],
				StatementRank: stmtRank,
			})
	}
	__antithesis_instrumentation__.Notify(594673)
	for i, t := range ts.Targets {
		__antithesis_instrumentation__.Notify(594682)
		descID := screl.GetDescID(t.Element())
		state := states[descID]
		stmtID := t.Metadata.StatementID
		noteRelevantStatement(state, stmtID)
		state.Targets = append(state.Targets, t)
		state.TargetRanks = append(state.TargetRanks, uint32(i))
		state.CurrentStatuses = append(state.CurrentStatuses, statuses[i])
	}
	__antithesis_instrumentation__.Notify(594674)
	return descIDs, states
}
