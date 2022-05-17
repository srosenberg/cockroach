package opgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
)

type targetSpec struct {
	from, to        scpb.Status
	transitionSpecs []transitionSpec
}

type transitionSpec struct {
	from       scpb.Status
	to         scpb.Status
	revertible bool
	minPhase   scop.Phase
	emitFns    []interface{}
}

type transitionProperty interface {
	apply(spec *transitionSpec)
}

func to(to scpb.Status, properties ...transitionProperty) transitionSpec {
	__antithesis_instrumentation__.Notify(594090)
	ts := transitionSpec{
		to:         to,
		revertible: true,
	}
	for _, p := range properties {
		__antithesis_instrumentation__.Notify(594092)
		p.apply(&ts)
	}
	__antithesis_instrumentation__.Notify(594091)
	return ts
}

func revertible(b bool) transitionProperty {
	__antithesis_instrumentation__.Notify(594093)
	return revertibleProperty(b)
}

func minPhase(p scop.Phase) transitionProperty {
	__antithesis_instrumentation__.Notify(594094)
	return phaseProperty(p)
}

func emit(fn interface{}) transitionProperty {
	__antithesis_instrumentation__.Notify(594095)
	return emitFnSpec{fn}
}

type phaseProperty scop.Phase

func (p phaseProperty) apply(spec *transitionSpec) {
	__antithesis_instrumentation__.Notify(594096)
	spec.minPhase = scop.Phase(p)
}

type revertibleProperty bool

func (r revertibleProperty) apply(spec *transitionSpec) {
	__antithesis_instrumentation__.Notify(594097)
	spec.revertible = bool(r)
}

var _ transitionProperty = revertibleProperty(true)

type emitFnSpec struct {
	fn interface{}
}

func (e emitFnSpec) apply(spec *transitionSpec) {
	__antithesis_instrumentation__.Notify(594098)
	spec.emitFns = append(spec.emitFns, e.fn)
}
