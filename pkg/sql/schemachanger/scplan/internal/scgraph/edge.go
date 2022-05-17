package scgraph

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
)

type Edge interface {
	From() *screl.Node
	To() *screl.Node
}

type OpEdge struct {
	from, to   *screl.Node
	op         []scop.Op
	typ        scop.Type
	revertible bool
	minPhase   scop.Phase
}

func (oe *OpEdge) From() *screl.Node { __antithesis_instrumentation__.Notify(594241); return oe.from }

func (oe *OpEdge) To() *screl.Node { __antithesis_instrumentation__.Notify(594242); return oe.to }

func (oe *OpEdge) Op() []scop.Op { __antithesis_instrumentation__.Notify(594243); return oe.op }

func (oe *OpEdge) Revertible() bool {
	__antithesis_instrumentation__.Notify(594244)
	return oe.revertible
}

func (oe *OpEdge) Type() scop.Type {
	__antithesis_instrumentation__.Notify(594245)
	return oe.typ
}

func (oe *OpEdge) IsPhaseSatisfied(phase scop.Phase) bool {
	__antithesis_instrumentation__.Notify(594246)
	return phase >= oe.minPhase
}

func (oe *OpEdge) String() string {
	__antithesis_instrumentation__.Notify(594247)
	from := screl.NodeString(oe.from)
	nonRevertible := ""
	if !oe.revertible {
		__antithesis_instrumentation__.Notify(594249)
		nonRevertible = "non-revertible"
	} else {
		__antithesis_instrumentation__.Notify(594250)
	}
	__antithesis_instrumentation__.Notify(594248)
	return fmt.Sprintf("%s -op-%s-> %s", from, nonRevertible, oe.to.CurrentStatus)
}

type DepEdgeKind int

const (
	_ DepEdgeKind = iota

	Precedence

	SameStagePrecedence
)

type DepEdge struct {
	from, to *screl.Node
	kind     DepEdgeKind

	rule string
}

func (de *DepEdge) From() *screl.Node { __antithesis_instrumentation__.Notify(594251); return de.from }

func (de *DepEdge) To() *screl.Node { __antithesis_instrumentation__.Notify(594252); return de.to }

func (de *DepEdge) Name() string { __antithesis_instrumentation__.Notify(594253); return de.rule }

func (de *DepEdge) Kind() DepEdgeKind { __antithesis_instrumentation__.Notify(594254); return de.kind }

func (de *DepEdge) String() string {
	__antithesis_instrumentation__.Notify(594255)
	from := screl.NodeString(de.from)
	to := screl.NodeString(de.to)
	return fmt.Sprintf("%s -dep-%s-> %s", from, de.kind, to)
}
