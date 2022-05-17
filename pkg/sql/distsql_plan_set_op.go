package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

func mergeResultTypesForSetOp(leftPlan, rightPlan *PhysicalPlan) ([]*types.T, error) {
	__antithesis_instrumentation__.Notify(467719)
	left, right := leftPlan.GetResultTypes(), rightPlan.GetResultTypes()
	if len(left) != len(right) {
		__antithesis_instrumentation__.Notify(467722)
		return nil, errors.Errorf("ResultTypes length mismatch: %d and %d", len(left), len(right))
	} else {
		__antithesis_instrumentation__.Notify(467723)
	}
	__antithesis_instrumentation__.Notify(467720)
	merged := make([]*types.T, len(left))
	for i := range left {
		__antithesis_instrumentation__.Notify(467724)
		leftType, rightType := left[i], right[i]
		if rightType.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(467725)
			merged[i] = leftType
		} else {
			__antithesis_instrumentation__.Notify(467726)
			if leftType.Family() == types.UnknownFamily {
				__antithesis_instrumentation__.Notify(467727)
				merged[i] = rightType
			} else {
				__antithesis_instrumentation__.Notify(467728)
				if leftType.Equivalent(rightType) {
					__antithesis_instrumentation__.Notify(467729)

					merged[i] = leftType
				} else {
					__antithesis_instrumentation__.Notify(467730)
					return nil, errors.Errorf(
						"conflicting ColumnTypes: %s and %s", leftType.DebugString(), rightType.DebugString())
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(467721)
	updateUnknownTypesForSetOp(leftPlan, merged)
	updateUnknownTypesForSetOp(rightPlan, merged)
	return merged, nil
}

func updateUnknownTypesForSetOp(plan *PhysicalPlan, merged []*types.T) {
	__antithesis_instrumentation__.Notify(467731)
	currentTypes := plan.GetResultTypes()
	for i := range merged {
		__antithesis_instrumentation__.Notify(467732)
		if merged[i].Family() == types.TupleFamily && func() bool {
			__antithesis_instrumentation__.Notify(467733)
			return currentTypes[i].Family() == types.UnknownFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(467734)
			for _, procIdx := range plan.ResultRouters {
				__antithesis_instrumentation__.Notify(467735)
				plan.Processors[procIdx].Spec.ResultTypes[i] = merged[i]
			}
		} else {
			__antithesis_instrumentation__.Notify(467736)
		}
	}
}
