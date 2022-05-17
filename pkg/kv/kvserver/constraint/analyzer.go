package constraint

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type AnalyzedConstraints struct {
	Constraints []roachpb.ConstraintsConjunction

	UnconstrainedReplicas bool

	SatisfiedBy [][]roachpb.StoreID

	Satisfies map[roachpb.StoreID][]int
}

var EmptyAnalyzedConstraints = AnalyzedConstraints{}

func AnalyzeConstraints(
	ctx context.Context,
	getStoreDescFn func(roachpb.StoreID) (roachpb.StoreDescriptor, bool),
	existing []roachpb.ReplicaDescriptor,
	numReplicas int32,
	constraints []roachpb.ConstraintsConjunction,
) AnalyzedConstraints {
	__antithesis_instrumentation__.Notify(101053)
	result := AnalyzedConstraints{
		Constraints: constraints,
	}

	if len(constraints) > 0 {
		__antithesis_instrumentation__.Notify(101057)
		result.SatisfiedBy = make([][]roachpb.StoreID, len(constraints))
		result.Satisfies = make(map[roachpb.StoreID][]int)
	} else {
		__antithesis_instrumentation__.Notify(101058)
	}
	__antithesis_instrumentation__.Notify(101054)

	var constrainedReplicas int32
	for i, subConstraints := range constraints {
		__antithesis_instrumentation__.Notify(101059)
		constrainedReplicas += subConstraints.NumReplicas
		for _, repl := range existing {
			__antithesis_instrumentation__.Notify(101060)

			store, ok := getStoreDescFn(repl.StoreID)
			if !ok || func() bool {
				__antithesis_instrumentation__.Notify(101061)
				return ConjunctionsCheck(store, subConstraints.Constraints) == true
			}() == true {
				__antithesis_instrumentation__.Notify(101062)
				result.SatisfiedBy[i] = append(result.SatisfiedBy[i], store.StoreID)
				result.Satisfies[store.StoreID] = append(result.Satisfies[store.StoreID], i)
			} else {
				__antithesis_instrumentation__.Notify(101063)
			}
		}
	}
	__antithesis_instrumentation__.Notify(101055)
	if constrainedReplicas > 0 && func() bool {
		__antithesis_instrumentation__.Notify(101064)
		return constrainedReplicas < numReplicas == true
	}() == true {
		__antithesis_instrumentation__.Notify(101065)
		result.UnconstrainedReplicas = true
	} else {
		__antithesis_instrumentation__.Notify(101066)
	}
	__antithesis_instrumentation__.Notify(101056)
	return result
}

func ConjunctionsCheck(store roachpb.StoreDescriptor, constraints []roachpb.Constraint) bool {
	__antithesis_instrumentation__.Notify(101067)
	for _, constraint := range constraints {
		__antithesis_instrumentation__.Notify(101069)

		hasConstraint := roachpb.StoreMatchesConstraint(store, constraint)
		if (constraint.Type == roachpb.Constraint_REQUIRED && func() bool {
			__antithesis_instrumentation__.Notify(101070)
			return !hasConstraint == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(101071)
			return (constraint.Type == roachpb.Constraint_PROHIBITED && func() bool {
				__antithesis_instrumentation__.Notify(101072)
				return hasConstraint == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(101073)
			return false
		} else {
			__antithesis_instrumentation__.Notify(101074)
		}
	}
	__antithesis_instrumentation__.Notify(101068)
	return true
}
