package rules

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraph"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/errors"
)

func idInIDs(objects []descpb.ID, id descpb.ID) bool {
	__antithesis_instrumentation__.Notify(594162)
	for _, other := range objects {
		__antithesis_instrumentation__.Notify(594164)
		if other == id {
			__antithesis_instrumentation__.Notify(594165)
			return true
		} else {
			__antithesis_instrumentation__.Notify(594166)
		}
	}
	__antithesis_instrumentation__.Notify(594163)
	return false
}

func targetNodeVars(el rel.Var) (element, target, node rel.Var) {
	__antithesis_instrumentation__.Notify(594167)
	return el, el + "-target", el + "-node"
}

type depRuleSpec struct {
	ruleName              string
	edgeKind              scgraph.DepEdgeKind
	targetStatus          scpb.Status
	from, to              elementSpec
	joinAttrs             []screl.Attr
	joinReferencingDescID bool
	filterLabel           string
	filter                interface{}
}

func depRule(
	ruleName string,
	edgeKind scgraph.DepEdgeKind,
	targetStatus scpb.TargetStatus,
	from, to elementSpec,
	joinAttrs ...screl.Attr,
) depRuleSpec {
	__antithesis_instrumentation__.Notify(594168)
	return depRuleSpec{
		ruleName:     ruleName,
		edgeKind:     edgeKind,
		targetStatus: targetStatus.Status(),
		from:         from,
		to:           to,
		joinAttrs:    joinAttrs,
	}
}

type elementSpec struct {
	status scpb.Status
	types  []interface{}
}

func element(status scpb.Status, types ...interface{}) elementSpec {
	__antithesis_instrumentation__.Notify(594169)
	return elementSpec{
		status: status,
		types:  types,
	}
}

func (d depRuleSpec) withFilter(label string, predicate interface{}) depRuleSpec {
	__antithesis_instrumentation__.Notify(594170)
	d.filterLabel = label
	d.filter = predicate
	return d
}

func (d depRuleSpec) withJoinFromReferencedDescIDWithToDescID() depRuleSpec {
	__antithesis_instrumentation__.Notify(594171)
	d.joinReferencingDescID = true
	return d
}

func (d depRuleSpec) register() {
	__antithesis_instrumentation__.Notify(594172)
	var (
		from, fromTarget, fromNode = targetNodeVars("from")
		to, toTarget, toNode       = targetNodeVars("to")
	)
	if from == to {
		__antithesis_instrumentation__.Notify(594177)
		panic(errors.AssertionFailedf("elements cannot share same label %q", from))
	} else {
		__antithesis_instrumentation__.Notify(594178)
	}
	__antithesis_instrumentation__.Notify(594173)
	c := rel.Clauses{
		from.Type(d.from.types[0], d.from.types[1:]...),
		fromTarget.AttrEq(screl.TargetStatus, d.targetStatus),
		to.Type(d.to.types[0], d.to.types[1:]...),
		toTarget.AttrEq(screl.TargetStatus, d.targetStatus),

		fromNode.AttrEq(screl.CurrentStatus, d.from.status),
		toNode.AttrEq(screl.CurrentStatus, d.to.status),

		screl.JoinTargetNode(from, fromTarget, fromNode),
		screl.JoinTargetNode(to, toTarget, toNode),
	}
	for _, attr := range d.joinAttrs {
		__antithesis_instrumentation__.Notify(594179)
		v := rel.Var(attr.String() + "-join-var")
		c = append(c, from.AttrEqVar(attr, v), to.AttrEqVar(attr, v))
	}
	__antithesis_instrumentation__.Notify(594174)
	if d.joinReferencingDescID {
		__antithesis_instrumentation__.Notify(594180)
		v := rel.Var("joined-from-ref-desc-id-with-to-desc-id-var")
		c = append(c, from.AttrEqVar(screl.ReferencedDescID, v), to.AttrEqVar(screl.DescID, v))
	} else {
		__antithesis_instrumentation__.Notify(594181)
	}
	__antithesis_instrumentation__.Notify(594175)
	if d.filter != nil {
		__antithesis_instrumentation__.Notify(594182)
		c = append(c, rel.Filter(d.filterLabel, from, to)(d.filter))
	} else {
		__antithesis_instrumentation__.Notify(594183)
	}
	__antithesis_instrumentation__.Notify(594176)
	registerDepRule(d.ruleName, d.edgeKind, fromNode, toNode, screl.MustQuery(c...))
}
