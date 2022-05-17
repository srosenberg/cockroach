package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

const (
	InnerJoin        = JoinType_INNER
	LeftOuterJoin    = JoinType_LEFT_OUTER
	RightOuterJoin   = JoinType_RIGHT_OUTER
	FullOuterJoin    = JoinType_FULL_OUTER
	LeftSemiJoin     = JoinType_LEFT_SEMI
	LeftAntiJoin     = JoinType_LEFT_ANTI
	IntersectAllJoin = JoinType_INTERSECT_ALL
	ExceptAllJoin    = JoinType_EXCEPT_ALL
	RightSemiJoin    = JoinType_RIGHT_SEMI
	RightAntiJoin    = JoinType_RIGHT_ANTI
)

func JoinTypeFromAstString(joinStr string) JoinType {
	__antithesis_instrumentation__.Notify(252449)
	switch joinStr {
	case "", tree.AstInner, tree.AstCross:
		__antithesis_instrumentation__.Notify(252450)
		return InnerJoin

	case tree.AstLeft:
		__antithesis_instrumentation__.Notify(252451)
		return LeftOuterJoin

	case tree.AstRight:
		__antithesis_instrumentation__.Notify(252452)
		return RightOuterJoin

	case tree.AstFull:
		__antithesis_instrumentation__.Notify(252453)
		return FullOuterJoin

	default:
		__antithesis_instrumentation__.Notify(252454)
		panic(errors.AssertionFailedf("unknown join string %s", joinStr))
	}
}

func (j JoinType) IsSetOpJoin() bool {
	__antithesis_instrumentation__.Notify(252455)
	return j == IntersectAllJoin || func() bool {
		__antithesis_instrumentation__.Notify(252456)
		return j == ExceptAllJoin == true
	}() == true
}

func (j JoinType) ShouldIncludeLeftColsInOutput() bool {
	__antithesis_instrumentation__.Notify(252457)
	switch j {
	case RightSemiJoin, RightAntiJoin:
		__antithesis_instrumentation__.Notify(252458)
		return false
	default:
		__antithesis_instrumentation__.Notify(252459)
		return true
	}
}

func (j JoinType) ShouldIncludeRightColsInOutput() bool {
	__antithesis_instrumentation__.Notify(252460)
	switch j {
	case LeftSemiJoin, LeftAntiJoin, IntersectAllJoin, ExceptAllJoin:
		__antithesis_instrumentation__.Notify(252461)
		return false
	default:
		__antithesis_instrumentation__.Notify(252462)
		return true
	}
}

func (j JoinType) IsEmptyOutputWhenRightIsEmpty() bool {
	__antithesis_instrumentation__.Notify(252463)
	switch j {
	case InnerJoin, RightOuterJoin, LeftSemiJoin,
		RightSemiJoin, IntersectAllJoin, RightAntiJoin:
		__antithesis_instrumentation__.Notify(252464)
		return true
	default:
		__antithesis_instrumentation__.Notify(252465)
		return false
	}
}

func (j JoinType) IsLeftOuterOrFullOuter() bool {
	__antithesis_instrumentation__.Notify(252466)
	return j == LeftOuterJoin || func() bool {
		__antithesis_instrumentation__.Notify(252467)
		return j == FullOuterJoin == true
	}() == true
}

func (j JoinType) IsLeftAntiOrExceptAll() bool {
	__antithesis_instrumentation__.Notify(252468)
	return j == LeftAntiJoin || func() bool {
		__antithesis_instrumentation__.Notify(252469)
		return j == ExceptAllJoin == true
	}() == true
}

func (j JoinType) IsRightSemiOrRightAnti() bool {
	__antithesis_instrumentation__.Notify(252470)
	return j == RightSemiJoin || func() bool {
		__antithesis_instrumentation__.Notify(252471)
		return j == RightAntiJoin == true
	}() == true
}

func (j JoinType) MakeOutputTypes(left, right []*types.T) []*types.T {
	__antithesis_instrumentation__.Notify(252472)
	numOutputTypes := 0
	if j.ShouldIncludeLeftColsInOutput() {
		__antithesis_instrumentation__.Notify(252477)
		numOutputTypes += len(left)
	} else {
		__antithesis_instrumentation__.Notify(252478)
	}
	__antithesis_instrumentation__.Notify(252473)
	if j.ShouldIncludeRightColsInOutput() {
		__antithesis_instrumentation__.Notify(252479)
		numOutputTypes += len(right)
	} else {
		__antithesis_instrumentation__.Notify(252480)
	}
	__antithesis_instrumentation__.Notify(252474)
	outputTypes := make([]*types.T, 0, numOutputTypes)
	if j.ShouldIncludeLeftColsInOutput() {
		__antithesis_instrumentation__.Notify(252481)
		outputTypes = append(outputTypes, left...)
	} else {
		__antithesis_instrumentation__.Notify(252482)
	}
	__antithesis_instrumentation__.Notify(252475)
	if j.ShouldIncludeRightColsInOutput() {
		__antithesis_instrumentation__.Notify(252483)
		outputTypes = append(outputTypes, right...)
	} else {
		__antithesis_instrumentation__.Notify(252484)
	}
	__antithesis_instrumentation__.Notify(252476)
	return outputTypes
}
