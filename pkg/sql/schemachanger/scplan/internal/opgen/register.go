package opgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/errors"
)

func equiv(from scpb.Status) transitionSpec {
	__antithesis_instrumentation__.Notify(594068)
	return transitionSpec{from: from, revertible: true}
}

func notImplemented(e scpb.Element) *scop.NotImplemented {
	__antithesis_instrumentation__.Notify(594069)
	return &scop.NotImplemented{
		ElementType: reflect.ValueOf(e).Type().Elem().String(),
	}
}

func toPublic(initialStatus scpb.Status, specs ...transitionSpec) targetSpec {
	__antithesis_instrumentation__.Notify(594070)
	return asTargetSpec(scpb.Status_PUBLIC, initialStatus, specs...)
}

func toAbsent(initialStatus scpb.Status, specs ...transitionSpec) targetSpec {
	__antithesis_instrumentation__.Notify(594071)
	return asTargetSpec(scpb.Status_ABSENT, initialStatus, specs...)
}

func asTargetSpec(to, from scpb.Status, specs ...transitionSpec) targetSpec {
	__antithesis_instrumentation__.Notify(594072)
	return targetSpec{from: from, to: to, transitionSpecs: specs}
}

func (r *registry) register(e scpb.Element, targetSpecs ...targetSpec) {
	__antithesis_instrumentation__.Notify(594073)
	onErrPanic := func(err error) {
		__antithesis_instrumentation__.Notify(594076)
		if err != nil {
			__antithesis_instrumentation__.Notify(594077)
			panic(errors.NewAssertionErrorWithWrappedErrf(err, "element %T", e))
		} else {
			__antithesis_instrumentation__.Notify(594078)
		}
	}
	__antithesis_instrumentation__.Notify(594074)
	targets := make([]target, len(targetSpecs))
	for i, spec := range targetSpecs {
		__antithesis_instrumentation__.Notify(594079)
		var err error
		targets[i], err = makeTarget(e, spec)
		onErrPanic(err)
	}
	__antithesis_instrumentation__.Notify(594075)
	onErrPanic(validateTargets(targets))
	r.targets = append(r.targets, targets...)
}

func validateTargets(targets []target) error {
	__antithesis_instrumentation__.Notify(594080)
	allStatuses := map[scpb.Status]bool{}
	targetStatuses := make([]map[scpb.Status]bool, len(targets))
	for i, tgt := range targets {
		__antithesis_instrumentation__.Notify(594083)
		m := map[scpb.Status]bool{}
		for _, t := range tgt.transitions {
			__antithesis_instrumentation__.Notify(594085)
			m[t.from] = true
			allStatuses[t.from] = true
			m[t.to] = true
			allStatuses[t.to] = true
		}
		__antithesis_instrumentation__.Notify(594084)
		targetStatuses[i] = m
	}
	__antithesis_instrumentation__.Notify(594081)

	for i, tgt := range targets {
		__antithesis_instrumentation__.Notify(594086)
		m := targetStatuses[i]
		for s := range allStatuses {
			__antithesis_instrumentation__.Notify(594087)
			if !m[s] {
				__antithesis_instrumentation__.Notify(594088)
				return errors.Errorf("target %s: status %s is missing here but is featured in other targets",
					tgt.status.String(), s.String())
			} else {
				__antithesis_instrumentation__.Notify(594089)
			}
		}
	}
	__antithesis_instrumentation__.Notify(594082)
	return nil
}
